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

// VttSegment represents a segment from a VTT subtitle file
type VttSegment struct {
	StartTime string
	EndTime   string
	Text      string
}

// PlayData represents the structure of a play JSON file
type PlayData struct {
	Act      string          `json:"act"`
	Scene    string          `json:"scene"`
	Title    string          `json:"title"`
	Setting  string          `json:"setting"`
	Dialogue []DialogueEntry `json:"dialogue"`
}

// DialogueEntry represents a single dialogue entry in a play
type DialogueEntry struct {
	Character      string `json:"character"`
	StageDirection string `json:"stage_direction"`
	Line           string `json:"line"`
}

// HandleGenAudioCommand processes a simple text file and generates audio files
func HandleGenAudioCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a text file")
		return
	}

	textFile := args[0]
	var voice string
	if len(args) > 1 {
		voice = args[1]
	}

	if err := processSimpleTextFile(textFile, voice); err != nil {
		fmt.Printf("Error processing text file: %v\n", err)
	}
}

func processSimpleTextFile(filename, voice string) error {
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
		if line == "" {
			continue // Skip empty lines
		}

		// Check if audio file already exists
		existingFile := findExistingAudioFile(audioDir, sentenceNum)
		if existingFile != "" {
			fmt.Printf("Skipping sentence %d: %s already exists\n", sentenceNum, existingFile)
		} else {
			fmt.Printf("Processing sentence %d: %s\n", sentenceNum, line)

			// Generate audio filename with duration placeholder
			audioFilename := fmt.Sprintf("s%d_duration.wav", sentenceNum)
			audioPath := filepath.Join(audioDir, audioFilename)

			// Call chatterbox to generate audio
			if err := callChatterboxWithVoice(line, audioPath, voice); err != nil {
				fmt.Printf("Error generating audio for sentence %d: %v\n", sentenceNum, err)
				continue
			}

			// Get actual duration and rename file
			duration, err := getAudioDurationSeconds(audioPath)
			if err != nil {
				fmt.Printf("Warning: Could not get duration for %s: %v\n", audioPath, err)
			} else {
				// Rename file with actual duration
				newFilename := fmt.Sprintf("s%d_%.3fs.wav", sentenceNum, duration)
				newPath := filepath.Join(audioDir, newFilename)
				if err := os.Rename(audioPath, newPath); err != nil {
					fmt.Printf("Warning: Could not rename %s to %s: %v\n", audioPath, newPath, err)
				} else {
					fmt.Printf("Generated: %s (%.3fs)\n", newFilename, duration)
				}
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
	pattern := fmt.Sprintf("s%d*.wav", sentenceNum)
	matches, err := filepath.Glob(filepath.Join(audioDir, pattern))
	if err != nil || len(matches) == 0 {
		return ""
	}
	return filepath.Base(matches[0]) // Return just the filename
}

func callChatterbox(sentence, audioFilename string) error {
	cmd := exec.Command("/opt/miniconda3/envs/chatterbox/bin/python3",
		"/opt/miniconda3/envs/chatterbox/local.py",
		sentence,
		audioFilename)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chatterbox failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func callChatterboxWithVoice(sentence, audioFilename, voice string) error {
	if voice == "" {
		// No voice specified, use the original callChatterbox function
		return callChatterbox(sentence, audioFilename)
	}

	// Voice specified, use utah.py like genaudio-play does
	cmd := exec.Command("/opt/miniconda3/envs/chatterbox/bin/python",
		"/opt/miniconda3/envs/chatterbox/utah.py",
		sentence,
		audioFilename,
		voice)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chatterbox utah failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

func getAudioDurationSeconds(audioFile string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", audioFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

// HandleParseVttCommand processes a VTT file and extracts plain text
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

	// Skip WebVTT header
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "WEBVTT" || line == "" {
			continue
		}
		break
	}

	// Process subtitle entries
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip timing lines (contain -->)
		if strings.Contains(line, "-->") {
			continue
		}

		// Skip empty lines and sequence numbers
		if line == "" || (len(line) < 10 && !strings.Contains(line, " ")) {
			continue
		}

		// Remove HTML-like tags and positioning data
		line = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(line, "")
		line = regexp.MustCompile(`\{[^}]*\}`).ReplaceAllString(line, "")

		if strings.TrimSpace(line) != "" {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

// HandleParseVttAndCutCommand processes VTT and cuts video into segments
func HandleParseVttAndCutCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a video ID")
		return
	}

	videoID := args[0]
	if err := parseVttAndCutVideo(videoID); err != nil {
		fmt.Printf("Error processing VTT and cutting video: %v\n", err)
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

	// Parse VTT file to get segments
	segments, err := parseVttForSegments(vttFile)
	if err != nil {
		return fmt.Errorf("failed to parse VTT file: %v", err)
	}

	// Process each segment
	for i, segment := range segments {
		segmentNum := i + 1

		// Check if segment already exists
		existingFile := findExistingVideoSegment(outputDir, segmentNum)
		if existingFile != "" {
			fmt.Printf("Skipping segment %d: %s already exists\n", segmentNum, existingFile)
			continue
		}

		// Calculate duration
		duration, err := calculateDuration(segment.StartTime, segment.EndTime)
		if err != nil {
			fmt.Printf("Error calculating duration for segment %d: %v\n", segmentNum, err)
			continue
		}

		// Clean the text for filename
		cleanText := sanitizeFilename(segment.Text)
		if len(cleanText) > 50 {
			cleanText = cleanText[:50]
		}

		// Create output filename
		outputFile := fmt.Sprintf("%04d_%.3fs_%s.mov", segmentNum, duration, cleanText)
		outputPath := filepath.Join(outputDir, outputFile)

		fmt.Printf("Extracting segment %d: %s (%.3fs)\n", segmentNum, segment.Text, duration)

		// Extract video segment
		if err := extractVideoSegment(videoFile, segment.StartTime, segment.EndTime, outputPath); err != nil {
			fmt.Printf("Error extracting segment %d: %v\n", segmentNum, err)
			continue
		}

		fmt.Printf("Generated: %s\n", outputFile)
	}

	return nil
}

func parseVttForSegments(filename string) ([]VttSegment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var segments []VttSegment
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for timing lines (contain -->)
		if strings.Contains(line, "-->") {
			parts := strings.Split(line, "-->")
			if len(parts) != 2 {
				continue
			}

			startTime := strings.TrimSpace(parts[0])
			endTime := strings.TrimSpace(parts[1])

			// Read the next lines for subtitle text
			var textLines []string
			for scanner.Scan() {
				textLine := strings.TrimSpace(scanner.Text())
				if textLine == "" {
					break // End of this subtitle entry
				}
				// Remove HTML-like tags and positioning data
				textLine = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(textLine, "")
				textLine = regexp.MustCompile(`\{[^}]*\}`).ReplaceAllString(textLine, "")
				if textLine != "" {
					textLines = append(textLines, textLine)
				}
			}

			if len(textLines) > 0 {
				segments = append(segments, VttSegment{
					StartTime: startTime,
					EndTime:   endTime,
					Text:      strings.Join(textLines, " "),
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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

	// Remove or replace problematic characters
	text = regexp.MustCompile(`[^\w\s-]`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, "_")
	text = strings.Trim(text, "_")

	return text
}

func extractVideoSegment(inputFile, startTime, endTime, outputFile string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-ss", startTime,
		"-to", endTime,
		"-c", "copy",
		"-avoid_negative_ts", "make_zero",
		"-y", // Overwrite output file
		outputFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// HandleGenAudioPlayCommand processes a play JSON file and generates audio with consistent voices
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
	// Read and parse JSON file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var playData PlayData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&playData); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Create output directory
	baseName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	outputDir := fmt.Sprintf("./data/%s_audio", baseName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create character voice mapping
	voiceMapping := createCharacterVoiceMapping(playData.Dialogue)

	fmt.Printf("Voice mapping:\n")
	for character, voice := range voiceMapping {
		fmt.Printf("  %s: %s\n", character, voice)
	}

	// Process each dialogue entry
	for i, entry := range playData.Dialogue {
		lineNum := fmt.Sprintf("%03d", i+1)
		voice := voiceMapping[entry.Character]

		// Check if audio file already exists
		audioFile := filepath.Join(outputDir, lineNum+".wav")
		if _, err := os.Stat(audioFile); err == nil {
			fmt.Printf("Skipping %s.wav (already exists)\n", lineNum)
			continue
		}

		fmt.Printf("Generating %s.wav: %s (%s)\n", lineNum, entry.Line, voice)

		if err := callChatterboxUtah(entry.Line, lineNum, voice); err != nil {
			fmt.Printf("Error generating audio for line %s: %v\n", lineNum, err)
			continue
		}

		// Move generated file to output directory
		generatedFile := lineNum + ".wav"
		if _, err := os.Stat(generatedFile); err == nil {
			finalPath := filepath.Join(outputDir, generatedFile)
			if err := os.Rename(generatedFile, finalPath); err != nil {
				fmt.Printf("Warning: Could not move %s to %s: %v\n", generatedFile, finalPath, err)
			}
		}
	}

	return nil
}

func createCharacterVoiceMapping(dialogue []DialogueEntry) map[string]string {
	voices := []string{
		"David", "Aria", "Guy", "Jenny", "Ryan", "Nova", "Lewis", "Amy",
		"Brian", "Andrew", "Emma", "Maya", "Brandon", "Christopher", "Cooper",
		"Samuel", "Evan", "Greg", "Jacob", "Luna", "Davis", "Will", "Nolan",
	}

	characters := make(map[string]bool)
	for _, entry := range dialogue {
		if entry.Character != "" {
			characters[entry.Character] = true
		}
	}

	// Convert map keys to slice for consistent ordering
	var characterList []string
	for character := range characters {
		characterList = append(characterList, character)
	}

	// Create deterministic mapping using MD5 hash
	mapping := make(map[string]string)
	for _, character := range characterList {
		hasher := md5.New()
		hasher.Write([]byte(character))
		hash := fmt.Sprintf("%x", hasher.Sum(nil))

		// Use hash to deterministically select voice
		hashValue := 0
		for _, b := range hash[:4] {
			hashValue = hashValue*256 + int(b)
		}
		voiceIndex := hashValue % len(voices)
		mapping[character] = voices[voiceIndex]
	}

	return mapping
}

func callChatterboxUtah(line, lineNum, voice string) error {
	cmd := exec.Command("/opt/miniconda3/envs/chatterbox/bin/python",
		"/opt/miniconda3/envs/chatterbox/utah.py",
		line,
		lineNum,
		voice)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chatterbox utah failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}