package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
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
	
	return float64(hours*3600 + minutes*60) + seconds, nil
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
	Act      string `json:"act"`
	Scene    string `json:"scene"`
	Title    string `json:"title"`
	Setting  string `json:"setting"`
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
		"agucchi", "algernon", "amanda", "archibald", "australian", "deep", "doug", "drew", "dundee", "elsa",
		"hank", "harry", "heather", "jane", "jessica", "karen", "kevin", "kosovo", "mike", "miss",
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