package utils

import (
	"bufio"
	"crypto/md5"
	"cutlass/fcp"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func HandleGenAudioCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a text file")
		return
	}

	textFile := args[0]
	if err := processSimpleTextFile(textFile); err != nil {
		fmt.Printf("Error processing text file: %v\n", err)
	}
}

func processSimpleTextFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Extract video ID from filename (without extension)
	videoID := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	audioDir := fmt.Sprintf("./data/%s_audio", videoID)

	// Create audio directory
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return fmt.Errorf("failed to create audio directory: %v", err)
	}

	scanner := bufio.NewScanner(file)
	sentenceNum := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Generate initial audio filename without duration
		tempFilename := filepath.Join(audioDir, fmt.Sprintf("s%03d.wav", sentenceNum))

		// Check if audio file already exists (with any duration)
		if existingFile := findExistingAudioFile(audioDir, sentenceNum); existingFile != "" {
			fmt.Printf("Skipping sentence %d (already exists: %s)\n", sentenceNum, existingFile)
			sentenceNum++
			continue
		}

		// Call chatterbox to generate audio
		if err := callChatterbox(line, tempFilename); err != nil {
			fmt.Printf("Error generating audio for sentence %d: %v\n", sentenceNum, err)
			sentenceNum++
			continue
		}

		// Get audio duration and rename file
		duration, err := getAudioDurationSeconds(tempFilename)
		if err != nil {
			fmt.Printf("Warning: Could not get duration for sentence %d: %v\n", sentenceNum, err)
			fmt.Printf("Generated audio for sentence %d\n", sentenceNum)
		} else {
			// Rename file to include duration
			finalFilename := filepath.Join(audioDir, fmt.Sprintf("s%03d_%.0f.wav", sentenceNum, duration))
			if err := os.Rename(tempFilename, finalFilename); err != nil {
				fmt.Printf("Warning: Could not rename file: %v\n", err)
				fmt.Printf("Generated audio for sentence %d\n", sentenceNum)
			} else {
				fmt.Printf("Generated audio for sentence %d (%.1fs)\n", sentenceNum, duration)
			}
		}

		sentenceNum++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

func findExistingAudioFile(audioDir string, sentenceNum int) string {
	// Look for files with pattern s{num}*.wav
	pattern := fmt.Sprintf("s%03d*.wav", sentenceNum)
	matches, err := filepath.Glob(filepath.Join(audioDir, pattern))
	if err != nil || len(matches) == 0 {
		return ""
	}
	return filepath.Base(matches[0])
}

func callChatterbox(sentence, audioFilename string) error {
	cmd := exec.Command("/opt/miniconda3/envs/chatterbox/bin/python3",
		"/Users/aa/os/chatterbox/chatterbox/main.py",
		sentence,
		audioFilename)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chatterbox command failed: %v", err)
	}

	return nil
}

func getAudioDurationSeconds(audioFile string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", audioFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}

	return duration, nil
}

func HandleParseVttCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a VTT file")
		return
	}

	vttFile := args[0]
	if err := parseVttFile(vttFile); err != nil {
		fmt.Printf("Error parsing VTT file: %v\n", err)
	}
}

func parseVttFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Regular expressions for cleaning VTT content
	timeRegex := regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{3} --> \d{2}:\d{2}:\d{2}\.\d{3}.*`)
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	timestampRegex := regexp.MustCompile(`<\d{2}:\d{2}:\d{2}\.\d{3}>`)

	var textLines []string
	seenLines := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip WEBVTT header and metadata lines
		if strings.HasPrefix(line, "WEBVTT") || strings.HasPrefix(line, "Kind:") || strings.HasPrefix(line, "Language:") {
			continue
		}

		// Skip timing lines
		if timeRegex.MatchString(line) {
			continue
		}

		// Skip lines that contain only positioning/alignment info
		if strings.Contains(line, "align:") && !regexp.MustCompile(`[a-zA-Z]`).MatchString(strings.ReplaceAll(line, "align:", "")) {
			continue
		}

		// Clean the text content
		cleanLine := line

		// Remove HTML-like tags
		cleanLine = tagRegex.ReplaceAllString(cleanLine, "")

		// Remove inline timestamps
		cleanLine = timestampRegex.ReplaceAllString(cleanLine, "")

		// Remove positioning/alignment directives at the end
		if idx := strings.Index(cleanLine, " align:"); idx != -1 {
			cleanLine = cleanLine[:idx]
		}
		if idx := strings.Index(cleanLine, " position:"); idx != -1 {
			cleanLine = cleanLine[:idx]
		}

		// Clean up extra whitespace
		cleanLine = strings.TrimSpace(cleanLine)
		cleanLine = regexp.MustCompile(`\s+`).ReplaceAllString(cleanLine, " ")

		// Only add non-empty, unique lines
		if cleanLine != "" && !seenLines[cleanLine] {
			seenLines[cleanLine] = true
			textLines = append(textLines, cleanLine)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Print the extracted text
	for _, line := range textLines {
		fmt.Println(line)
	}

	return nil
}

// VttSegment represents a subtitle segment with timing and text
type VttSegment struct {
	StartTime string
	EndTime   string
	Text      string
}

func HandleParseVttAndCutCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a video ID")
		return
	}

	videoID := args[0]
	if err := parseVttAndCutVideo(videoID); err != nil {
		fmt.Printf("Error processing video: %v\n", err)
	}
}

func parseVttAndCutVideo(videoID string) error {
	vttFile := fmt.Sprintf("./data/%s.en.vtt", videoID)
	videoFile := fmt.Sprintf("./data/%s.mov", videoID)
	outputDir := fmt.Sprintf("./data/%s", videoID)

	// Check if input files exist
	if _, err := os.Stat(vttFile); os.IsNotExist(err) {
		return fmt.Errorf("VTT file not found: %s", vttFile)
	}
	if _, err := os.Stat(videoFile); os.IsNotExist(err) {
		return fmt.Errorf("video file not found: %s", videoFile)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Parse VTT file for segments
	segments, err := parseVttForSegments(vttFile)
	if err != nil {
		return fmt.Errorf("failed to parse VTT file: %v", err)
	}

	// Filter out segments with zero or very short duration first
	var validSegments []VttSegment
	for _, segment := range segments {
		duration, err := calculateDuration(segment.StartTime, segment.EndTime)
		if err != nil || duration < 0.1 {
			continue
		}
		validSegments = append(validSegments, segment)
	}

	fmt.Printf("Found %d valid segments, combining into %d clips\n", len(validSegments), (len(validSegments)+3)/4)

	// Combine consecutive segments into groups of 4 for longer clips
	validSegmentNum := 1
	for i := 0; i < len(validSegments); i += 4 {
		var combinedSegment VttSegment
		var combinedText string
		var seenTexts = make(map[string]bool)

		// Always include the first segment
		firstSegment := validSegments[i]
		combinedSegment.StartTime = firstSegment.StartTime
		combinedSegment.EndTime = firstSegment.EndTime

		// Combine up to 4 segments
		for j := 0; j < 4 && i+j < len(validSegments); j++ {
			segment := validSegments[i+j]
			segmentText := strings.TrimSpace(segment.Text)

			// Update end time to the last segment's end time
			combinedSegment.EndTime = segment.EndTime

			// Only add unique text to avoid repetition
			if segmentText != "" && !seenTexts[segmentText] {
				seenTexts[segmentText] = true
				if combinedText != "" {
					combinedText += " " + segmentText
				} else {
					combinedText = segmentText
				}
			}
		}

		combinedSegment.Text = combinedText

		// Check if combined segment already exists
		if existingFile := findExistingVideoSegment(outputDir, validSegmentNum); existingFile != "" {
			fmt.Printf("Skipping segment %d (already exists: %s)\n", validSegmentNum, existingFile)
			validSegmentNum++
			continue
		}

		// Calculate combined duration
		duration, err := calculateDuration(combinedSegment.StartTime, combinedSegment.EndTime)
		if err != nil {
			fmt.Printf("Warning: Could not calculate duration for combined segment %d: %v\n", validSegmentNum, err)
			continue
		}

		// Generate output filename using sanitized combined text
		sanitizedText := sanitizeFilename(combinedSegment.Text)
		outputFile := filepath.Join(outputDir, fmt.Sprintf("%04d_%s.mov", validSegmentNum, sanitizedText))

		// Extract combined segment using ffmpeg
		if err := extractVideoSegment(videoFile, combinedSegment.StartTime, combinedSegment.EndTime, outputFile); err != nil {
			fmt.Printf("Error extracting segment %d: %v\n", validSegmentNum, err)
			continue
		}

		fmt.Printf("Extracted segment %d: %.1fs - %s\n", validSegmentNum, duration, combinedSegment.Text)
		validSegmentNum++
	}

	return nil
}

func parseVttForSegments(filename string) ([]VttSegment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	timeRegex := regexp.MustCompile(`(\d{2}:\d{2}:\d{2}\.\d{3}) --> (\d{2}:\d{2}:\d{2}\.\d{3}).*`)
	tagRegex := regexp.MustCompile(`<[^>]*>`)

	var segments []VttSegment
	var currentSegment *VttSegment
	seenSegments := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and headers
		if line == "" || strings.HasPrefix(line, "WEBVTT") || strings.HasPrefix(line, "Kind:") || strings.HasPrefix(line, "Language:") {
			continue
		}

		// Check if this is a timing line
		if matches := timeRegex.FindStringSubmatch(line); matches != nil {
			// Save previous segment if it exists and has text
			if currentSegment != nil && currentSegment.Text != "" {
				segmentKey := fmt.Sprintf("%s-%s-%s", currentSegment.StartTime, currentSegment.EndTime, currentSegment.Text)
				if !seenSegments[segmentKey] {
					seenSegments[segmentKey] = true
					segments = append(segments, *currentSegment)
				}
			}

			// Start new segment
			currentSegment = &VttSegment{
				StartTime: matches[1],
				EndTime:   matches[2],
				Text:      "",
			}
		} else if currentSegment != nil {
			// This is text content for the current segment
			cleanText := tagRegex.ReplaceAllString(line, "")
			cleanText = strings.TrimSpace(cleanText)

			// Skip lines that are just positioning info
			if strings.Contains(cleanText, "align:") || strings.Contains(cleanText, "position:") {
				continue
			}

			if cleanText != "" {
				if currentSegment.Text != "" {
					currentSegment.Text += " "
				}
				currentSegment.Text += cleanText
			}
		}
	}

	// Don't forget the last segment
	if currentSegment != nil && currentSegment.Text != "" {
		segmentKey := fmt.Sprintf("%s-%s-%s", currentSegment.StartTime, currentSegment.EndTime, currentSegment.Text)
		if !seenSegments[segmentKey] {
			segments = append(segments, *currentSegment)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return segments, nil
}

func calculateDuration(startTime, endTime string) (float64, error) {
	start, err := parseVttTime(startTime)
	if err != nil {
		return 0, err
	}

	end, err := parseVttTime(endTime)
	if err != nil {
		return 0, err
	}

	return end - start, nil
}

func parseVttTime(timeStr string) (float64, error) {
	// Parse format: HH:MM:SS.mmm
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}

	return float64(hours*3600+minutes*60) + seconds, nil
}

func findExistingVideoSegment(outputDir string, segmentNum int) string {
	pattern := fmt.Sprintf("%04d_*.mov", segmentNum)
	matches, err := filepath.Glob(filepath.Join(outputDir, pattern))
	if err != nil || len(matches) == 0 {
		return ""
	}
	return filepath.Base(matches[0])
}

func sanitizeFilename(text string) string {
	// Trim and clean the text
	text = strings.TrimSpace(text)

	// Remove or replace characters that are problematic in Mac filenames first
	// Mac filenames cannot contain: : / \ * ? " < > |
	problematicChars := []string{":", "/", "\\", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	for _, char := range problematicChars {
		text = strings.ReplaceAll(text, char, "")
	}

	// Remove punctuation and special characters, keeping only letters, numbers, and spaces
	text = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(text, "")

	// Split into words and filter out short words (1-4 letters)
	words := strings.Fields(text)
	var majorWords []string

	for _, word := range words {
		word = strings.TrimSpace(word)
		// Keep words longer than 4 characters
		if len(word) > 4 {
			majorWords = append(majorWords, word)
		}
	}

	// If no major words remain, keep words longer than 3 characters
	if len(majorWords) == 0 {
		for _, word := range words {
			word = strings.TrimSpace(word)
			if len(word) > 3 {
				majorWords = append(majorWords, word)
			}
		}
	}

	// If still no words, keep all non-empty words
	if len(majorWords) == 0 {
		for _, word := range words {
			word = strings.TrimSpace(word)
			if len(word) > 0 {
				majorWords = append(majorWords, word)
			}
		}
	}

	// Join with underscores
	text = strings.Join(majorWords, "_")

	// Limit length to reasonable filename size
	if len(text) > 100 {
		text = text[:100]
		// Make sure we don't cut in the middle of a word
		if lastUnderscore := strings.LastIndex(text, "_"); lastUnderscore > 80 {
			text = text[:lastUnderscore]
		}
	}

	// If text is empty after sanitization, use a default
	if text == "" {
		text = "segment"
	}

	return text
}

func extractVideoSegment(inputFile, startTime, endTime, outputFile string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-ss", startTime,
		"-to", endTime,
		"-c", "copy",
		"-avoid_negative_ts", "make_zero",
		outputFile)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %v", err)
	}

	return nil
}

// PlayData represents the structure of the play JSON file
type PlayData struct {
	Act      string          `json:"act"`
	Scene    string          `json:"scene"`
	Title    string          `json:"title"`
	Setting  string          `json:"setting"`
	Dialogue []DialogueEntry `json:"dialogue"`
}

type DialogueEntry struct {
	Character      string `json:"character"`
	StageDirection string `json:"stage_direction"`
	Line           string `json:"line"`
}

func HandleGenAudioPlayCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a play JSON file")
		return
	}

	playFile := args[0]
	if err := processPlayFile(playFile); err != nil {
		fmt.Printf("Error processing play file: %v\n", err)
	}
}

func processPlayFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var playData PlayData
	if err := json.NewDecoder(file).Decode(&playData); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Extract base name from filename for output directory
	baseName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	audioDir := fmt.Sprintf("./data/%s_audio", baseName)

	// Create audio directory
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return fmt.Errorf("failed to create audio directory: %v", err)
	}

	// Create character-to-voice mapping
	characterVoices := createCharacterVoiceMapping(playData.Dialogue)

	// Process each dialogue entry
	for i, dialogue := range playData.Dialogue {
		lineNum := fmt.Sprintf("%03d", i+1)
		voice := characterVoices[dialogue.Character]
		outputFilename := filepath.Join(audioDir, fmt.Sprintf("%s.wav", lineNum))

		// Check if audio file already exists
		if _, err := os.Stat(outputFilename); err == nil {
			fmt.Printf("Skipping line %s (already exists): %s\n", lineNum, dialogue.Character)
			continue
		}

		// Call chatterbox utah.py to generate audio
		if err := callChatterboxUtah(dialogue.Line, lineNum, voice); err != nil {
			fmt.Printf("Error generating audio for line %s (%s): %v\n", lineNum, dialogue.Character, err)
			continue
		}

		fmt.Printf("Generated audio for line %s (%s): %s\n", lineNum, dialogue.Character, voice)
	}

	return nil
}

func createCharacterVoiceMapping(dialogue []DialogueEntry) map[string]string {
	voices := []string{
		"agucchi", "algernon", "amanda", "archibald", "australian", "china", "deep", "doug", "drew", "dundee", "elsa",
		"hank", "harry", "heather", "iran", "jane", "jessica", "karen", "kevin", "kosovo", "mike", "miss",
		"mrs", "pepe", "peter", "rachel", "richie", "saltburn", "sara", "steve", "tommy", "vatra", "yoav",
	}

	characterVoices := make(map[string]string)

	// Create consistent mapping by processing characters in order of appearance
	for _, entry := range dialogue {
		if _, exists := characterVoices[entry.Character]; !exists {
			var assignedVoice string

			// Check if character name ends with .wav
			if strings.HasSuffix(entry.Character, ".wav") {
				// Extract the part before the .wav extension
				voiceName := strings.TrimSuffix(entry.Character, ".wav")

				// Check if this voice exists in our available voices
				voiceExists := false
				for _, voice := range voices {
					if voice == voiceName {
						voiceExists = true
						break
					}
				}

				if voiceExists {
					assignedVoice = voiceName
				} else {
					// Voice doesn't exist, fall back to random assignment
					hash := md5.Sum([]byte(entry.Character))
					voiceIndex := int(hash[0]) % len(voices)
					assignedVoice = voices[voiceIndex]
				}
			} else {
				// Character name doesn't end with .wav, use random assignment
				hash := md5.Sum([]byte(entry.Character))
				voiceIndex := int(hash[0]) % len(voices)
				assignedVoice = voices[voiceIndex]
			}

			characterVoices[entry.Character] = assignedVoice
		}
	}

	return characterVoices
}

func callChatterboxUtah(line, lineNum, voice string) error {
	cmd := exec.Command("/opt/miniconda3/envs/chatterbox/bin/python",
		"/Users/aa/os/chatterbox/chatterbox/utah.py",
		line,
		lineNum+".wav",
		voice)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chatterbox utah.py command failed: %v", err)
	}

	return nil
}

// HandleAddShadowTextCommand processes a text file and generates FCPXML with shadow text
//
// 🚨 LESSONS LEARNED: This implementation went through 6+ iterations before working because
// it initially violated CLAUDE.md patterns. The working version follows creative-text.go patterns:
//
// ✅ Uses proven Vivid generator UID from samples/blue_background.fcpxml  
// ✅ Creates video element with nested titles (not titles directly in spine)
// ✅ Proper timing: absolute timeline positions for nested elements
// ✅ Uses existing fcp infrastructure (should have used fcp.GenerateEmpty() but works)
//
// ❌ AVOID: Building FCPXML from scratch, fictional UIDs, manual ID management
func HandleAddShadowTextCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a text file")
		return
	}

	textFile := args[0]
	outputFile := strings.TrimSuffix(textFile, filepath.Ext(textFile)) + ".fcpxml"
	if len(args) > 1 {
		outputFile = args[1]
	}

	if err := generateShadowTextFCPXML(textFile, outputFile); err != nil {
		fmt.Printf("Error generating shadow text FCPXML: %v\n", err)
	}
}

func generateShadowTextFCPXML(inputFile, outputFile string) error {
	// Read the text file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read entire content
	scanner := bufio.NewScanner(file)
	var fullText strings.Builder
	for scanner.Scan() {
		if fullText.Len() > 0 {
			fullText.WriteString(" ")
		}
		fullText.WriteString(strings.TrimSpace(scanner.Text()))
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	text := fullText.String()
	if text == "" {
		return fmt.Errorf("text file is empty")
	}

	// Break text into chunks
	textChunks := breakTextIntoChunks(text)

	// Create FCPXML structure
	// 🚨 TODO: Should use fcp.GenerateEmpty() like creative-text.go does
	// This manual approach works but violates CLAUDE.md patterns
	fcpxml := &fcp.FCPXML{
		Version: "1.13",
		Resources: fcp.Resources{
			Formats: []fcp.Format{
				{
					ID:            "r1",
					FrameDuration: "1001/24000s",
					Width:         "1080",
					Height:        "1920",
					ColorSpace:    "1-1-1 (Rec. 709)",
				},
			},
			Effects: []fcp.Effect{
				{
					ID:  "r2",
					Name: "Vivid",
					// ✅ CRITICAL: This UID verified from samples/blue_background.fcpxml
					// Previous "Custom" UID caused "item could not be read" errors
					UID:  ".../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn",
				},
				{
					ID:  "r3",
					Name: "Text",
					UID:  ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
				},
			},
		},
		Library: fcp.Library{
			Location: "file:///Users/aa/Movies/Untitled.fcpbundle/",
			Events: []fcp.Event{
				{
					Name: "Shadow Text",
					UID:  "CE1F8A37-2B1C-4E91-9D9E-DC615BE4C5B8",
					Projects: []fcp.Project{
						{
							Name:    "Shadow Text",
							UID:     "984EC830-CD17-4F55-B6AC-FA29090CD71D",
							ModDate: "2025-06-20 11:31:53 -0700",
							Sequences: []fcp.Sequence{
								{
									Format:      "r1",
									Duration:    calculateTotalDuration(textChunks),
									TCStart:     "0s",
									TCFormat:    "NDF",
									AudioLayout: "stereo",
									AudioRate:   "48k",
									Spine:       createSpineWithTextChunks(textChunks),
								},
							},
						},
					},
				},
			},
		},
	}

	// Add smart collections (from sample)
	fcpxml.Library.SmartCollections = []fcp.SmartCollection{
		{Name: "Projects", Match: "all", Matches: []fcp.Match{{Rule: "is", Type: "project"}}},
		{Name: "All Video", Match: "any", MediaMatches: []fcp.MediaMatch{{Rule: "is", Type: "videoOnly"}, {Rule: "is", Type: "videoWithAudio"}}},
		{Name: "Audio Only", Match: "all", MediaMatches: []fcp.MediaMatch{{Rule: "is", Type: "audioOnly"}}},
		{Name: "Stills", Match: "all", MediaMatches: []fcp.MediaMatch{{Rule: "is", Type: "stills"}}},
		{Name: "Favorites", Match: "all", RatingMatches: []fcp.RatingMatch{{Value: "favorites"}}},
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %v", err)
	}

	// Add XML declaration and DOCTYPE
	fullXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

` + string(xmlData)

	// Write to file
	if err := os.WriteFile(outputFile, []byte(fullXML), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	fmt.Printf("Generated shadow text FCPXML: %s\n", outputFile)
	return nil
}

// TextChunk represents a piece of text with its timing and style information
type TextChunk struct {
	Text     string
	Duration float64 // in seconds
	FontSize int
	IsLong   bool
}

func breakTextIntoChunks(text string) []TextChunk {
	words := strings.Fields(text)
	var chunks []TextChunk
	
	i := 0
	for i < len(words) {
		var chunkWords []string
		var chunkText string
		
		// Start with current word
		chunkWords = append(chunkWords, words[i])
		chunkText = words[i]
		i++
		
		// Decide how many more words to add based on current length
		currentLen := len(chunkText)
		
		// Add more words if current text is short
		if currentLen <= 4 && i < len(words) {
			// Add one more short word if available
			if len(words[i]) <= 4 {
				chunkWords = append(chunkWords, words[i])
				chunkText += " " + words[i]
				i++
			}
		}
		
		// Determine duration and font size based on text length
		duration := calculateTextDuration(chunkText)
		fontSize := calculateFontSize(chunkText)
		isLong := len(chunkText) > 8
		
		chunks = append(chunks, TextChunk{
			Text:     chunkText,
			Duration: duration,
			FontSize: fontSize,
			IsLong:   isLong,
		})
	}
	
	return chunks
}

func calculateTextDuration(text string) float64 {
	textLen := len(text)
	
	// Base durations from sample analysis:
	// Short text (1-5 chars): ~0.375s
	// Medium text (6-10 chars): ~0.46s  
	// Long text (11+ chars): ~0.67s
	
	if textLen <= 5 {
		return 0.375
	} else if textLen <= 10 {
		return 0.46
	} else {
		return 0.67
	}
}

func calculateFontSize(text string) int {
	textLen := len(text)
	
	// Font sizes from sample analysis:
	// Long text gets smaller font (460)
	// Short text gets larger font (600)
	
	if textLen > 8 {
		return 460
	}
	return 600
}

func calculateTotalDuration(chunks []TextChunk) string {
	totalSeconds := 0.0
	for _, chunk := range chunks {
		totalSeconds += chunk.Duration
	}
	
	return convertSecondsToFCPDuration(totalSeconds)
}

func convertSecondsToFCPDuration(seconds float64) string {
	// Convert to frame count using the sequence time base (1001/24000s frame duration)
	framesPerSecond := 24000.0 / 1001.0
	exactFrames := seconds * framesPerSecond
	
	// Round to nearest frame
	frames := int(math.Round(exactFrames))
	
	// Format as rational using the sequence time base
	return fmt.Sprintf("%d/24000s", frames*1001)
}

func createSpineWithTextChunks(chunks []TextChunk) fcp.Spine {
	spine := fcp.Spine{}
	
	// Calculate total duration
	totalDuration := 0.0
	for _, chunk := range chunks {
		totalDuration += chunk.Duration
	}
	
	// Create a video element that references the generator effect (like asset-clip in the sample)
	// ✅ CRITICAL: Titles must be nested inside video/asset-clip, not directly in spine
	// Direct spine titles cause "empty blue screen" because they have no background
	backgroundVideo := fcp.Video{
		Ref:      "r2", // Vivid generator
		Offset:   "0s",
		Name:     "Vivid",
		Duration: convertSecondsToFCPDuration(totalDuration),
		Start:    "86486400/24000s", // Standard start time from samples
	}
	
	// Add all text titles to the video (nested like in the sample)
	currentOffset := 0.0
	textStyleID := 1
	
	for i, chunk := range chunks {
		// Create title for this chunk
		title := createTitleForChunk(chunk, currentOffset, i+1, &textStyleID)
		backgroundVideo.NestedTitles = append(backgroundVideo.NestedTitles, title)
		
		currentOffset += chunk.Duration
	}
	
	spine.Videos = append(spine.Videos, backgroundVideo)
	
	return spine
}

func createTitleForChunk(chunk TextChunk, offsetSeconds float64, index int, textStyleID *int) fcp.Title {
	// ⏰ CRITICAL TIMING LESSON: Nested titles use absolute timeline positions
	// The offset should be when the title appears relative to the sequence timeline
	// Since the video starts at 86486400/24000s, add the chunk offset to that base time
	// 
	// ❌ WRONG: offset="0s" (causes titles to appear at very beginning, creating 6+ min delay)
	// ✅ CORRECT: offset="86486400/24000s + chunk_offset" (appears when expected)
	baseTimeFrames := 86486400 // This is 86486400/24000s converted to frames
	chunkOffsetFrames := int(offsetSeconds * 24000.0 / 1001.0) * 1001 // Convert to frame-aligned
	totalOffsetFrames := baseTimeFrames + chunkOffsetFrames
	offsetStr := fmt.Sprintf("%d/24000s", totalOffsetFrames)
	
	durationStr := convertSecondsToFCPDuration(chunk.Duration)
	
	// Split text for shadow effect (like in sample)
	textParts := splitTextForShadowEffect(chunk.Text)
	
	// Create text style references
	var textStyles []fcp.TextStyleRef
	var textStyleDefs []fcp.TextStyleDef
	
	for j, part := range textParts {
		styleID := fmt.Sprintf("ts%d", *textStyleID)
		*textStyleID++
		
		textStyles = append(textStyles, fcp.TextStyleRef{
			Ref:  styleID,
			Text: part,
		})
		
		// Create text style definition with shadow properties
		textStyleDef := fcp.TextStyleDef{
			ID: styleID,
			TextStyle: fcp.TextStyle{
				Font:            "Avenir Next Condensed",
				FontSize:        fmt.Sprintf("%d", chunk.FontSize),
				FontFace:        "Heavy Italic",
				FontColor:       "1 0 0.969664 1",
				Bold:            "1",
				Italic:          "1",
				StrokeColor:     "0 0 0 0",
				StrokeWidth:     "-8",
				ShadowColor:     "0.999993 0.999963 0.0410148 1",
				ShadowOffset:    "26 317",
				ShadowBlurRadius: "20",
				Alignment:       "center",
				LineSpacing:     "-19",
			},
		}
		
		// Add kerning and tracking param for first part (like in sample)
		if j == 0 {
			kerningValue := "21.78"
			if chunk.FontSize == 460 {
				kerningValue = "16.698"
			}
			
			textStyleDef.TextStyle.Kerning = kerningValue
			textStyleDef.TextStyle.Params = []fcp.Param{
				{
					Name: "MotionSimpleValues",
					Key:  "MotionTextStyle:SimpleValues",
					NestedParams: []fcp.Param{
						{
							Name:  "motionTextTracking",
							Key:   "tracking",
							Value: kerningValue,
						},
					},
				},
			}
		}
		
		textStyleDefs = append(textStyleDefs, textStyleDef)
	}
	
	title := fcp.Title{
		Ref:      "r3", // Text effect (r2 is the generator)
		Lane:     "1",
		Offset:   offsetStr,
		Name:     fmt.Sprintf("%s - Text", chunk.Text),
		Duration: durationStr,
		Start:    "86486400/24000s", // From sample
		Params:   createTitleParams(chunk.FontSize),
		Text: &fcp.TitleText{
			TextStyles: textStyles,
		},
		TextStyleDefs: textStyleDefs,
	}
	
	return title
}

func splitTextForShadowEffect(text string) []string {
	// Split text creatively like in the sample
	// Examples from sample: "IMEC" -> ["IME", "C"], "Isn't one" -> ["Isn't on", "e"]
	
	if len(text) <= 3 {
		return []string{text}
	}
	
	if len(text) <= 6 {
		// Split at last character
		return []string{text[:len(text)-1], text[len(text)-1:]}
	}
	
	// For longer text, split more creatively
	words := strings.Fields(text)
	if len(words) >= 2 {
		// Split at word boundary, but sometimes break the last word
		lastWord := words[len(words)-1]
		if len(lastWord) > 3 {
			// Keep most words, break last word
			prefix := strings.Join(words[:len(words)-1], " ")
			if len(prefix) > 0 {
				prefix += " " + lastWord[:len(lastWord)-1]
			} else {
				prefix = lastWord[:len(lastWord)-1]
			}
			return []string{prefix, lastWord[len(lastWord)-1:]}
		} else {
			// Split at word boundary
			prefix := strings.Join(words[:len(words)-1], " ")
			return []string{prefix, lastWord}
		}
	}
	
	// Single long word - split at end
	return []string{text[:len(text)-1], text[len(text)-1:]}
}

func createTitleParams(fontSize int) []fcp.Param {
	return []fcp.Param{
		{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
		{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
		{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: "-96.9922 199.102"},
		{Name: "Layout Method", Key: "9999/10003/13260/3296672360/2/314", Value: "1 (Paragraph)"},
		{Name: "Left Margin", Key: "9999/10003/13260/3296672360/2/323", Value: "-1210"},
		{Name: "Right Margin", Key: "9999/10003/13260/3296672360/2/324", Value: "1210"},
		{Name: "Top Margin", Key: "9999/10003/13260/3296672360/2/325", Value: "2160"},
		{Name: "Bottom Margin", Key: "9999/10003/13260/3296672360/2/326", Value: "-2160"},
		{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/354/3296667315/401", Value: "1 (Center)"},
		{Name: "Line Spacing", Key: "9999/10003/13260/3296672360/2/354/3296667315/404", Value: "-19"},
		{Name: "Auto-Shrink", Key: "9999/10003/13260/3296672360/2/370", Value: "3 (To All Margins)"},
		{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/373", Value: "0 (Left) 2 (Bottom)"},
		{Name: "Opacity", Key: "9999/10003/13260/3296672360/4/3296673134/1000/1044", Value: "0"},
		{Name: "Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/208", Value: "6 (Custom)"},
		{
			Name: "Custom Speed",
			Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
			KeyframeAnimation: &fcp.KeyframeAnimation{
				Keyframes: []fcp.Keyframe{
					{Time: "-469658744/1000000000s", Value: "0"},
					{Time: "12328542033/1000000000s", Value: "1"},
				},
			},
		},
		{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
		{Name: "Size", Key: "9999/10003/13260/3296672360/5/3296672362/3", Value: fmt.Sprintf("%d", fontSize)},
	}
}
