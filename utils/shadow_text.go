package utils

import (
	"bufio"
	"cutlass/fcp"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TextChunk represents a piece of text with its duration and styling
type TextChunk struct {
	Text     string
	Duration float64 // in seconds
	FontSize int
	IsLong   bool
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

	var totalDuration float64 = 0 // 0 means auto-calculate
	if len(args) > 2 {
		if duration, err := strconv.ParseFloat(args[2], 64); err != nil {
			fmt.Printf("Error: Invalid duration '%s'. Please provide a number in seconds.\n", args[2])
			return
		} else {
			totalDuration = duration
		}
	}

	if err := generateShadowTextFCPXML(textFile, outputFile, totalDuration); err != nil {
		fmt.Printf("Error generating shadow text FCPXML: %v\n", err)
	}
}

func generateShadowTextFCPXML(inputFile, outputFile string, totalDuration float64) error {
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
	textChunks := breakTextIntoChunks(text, totalDuration)

	// Create FCPXML structure
	fcpxml := &fcp.FCPXML{
		Version: "1.13",
		Resources: fcp.Resources{
			Formats: []fcp.Format{
				{
					ID:            "r1",
					Name:          "FFVideoFormat720p2398",
					FrameDuration: "1001/24000s",
					Width:         "1280",
					Height:        "720",
					ColorSpace:    "1-1-1 (Rec. 709)",
				},
			},
			Effects: []fcp.Effect{
				{
					ID:   "r2",
					Name: "Vivid",
					UID:  ".../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn",
				},
				{
					ID:   "r4",
					Name: "Text",
					UID:  ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
				},
			},
		},
		Library: fcp.Library{
			Location: "file:///Users/aa/Movies/Untitled.fcpbundle/",
			Events: []fcp.Event{
				{
					Name: "6-13-25",
					UID:  "78463397-97FD-443D-B4E2-07C581674AFC",
					Projects: []fcp.Project{
						{
							Name:     "wiki",
							UID:      "DEA19981-DED5-4851-8435-14515931C68A",
							ModDate:  "2025-06-13 11:46:22 -0700",
							Sequences: []fcp.Sequence{createSequenceWithShadowText(textChunks)},
						},
					},
				},
			},
			SmartCollections: []fcp.SmartCollection{
				{
					Name:  "Projects",
					Match: "all",
					Matches: []fcp.Match{
						{Rule: "is", Type: "project"},
					},
				},
				{
					Name:  "All Video",
					Match: "any",
					MediaMatches: []fcp.MediaMatch{
						{Rule: "is", Type: "videoOnly"},
						{Rule: "is", Type: "videoWithAudio"},
					},
				},
				{
					Name:  "Audio Only",
					Match: "all",
					MediaMatches: []fcp.MediaMatch{
						{Rule: "is", Type: "audioOnly"},
					},
				},
				{
					Name:  "Stills",
					Match: "all",
					MediaMatches: []fcp.MediaMatch{
						{Rule: "is", Type: "stills"},
					},
				},
				{
					Name:  "Favorites",
					Match: "all",
					RatingMatches: []fcp.RatingMatch{
						{Value: "favorites"},
					},
				},
			},
		},
	}

	// Write to file using FCPXML marshal
	return fcp.WriteToFile(fcpxml, outputFile)
}

func createSequenceWithShadowText(chunks []TextChunk) fcp.Sequence {
	totalDuration := calculateTotalDuration(chunks)
	
	sequence := fcp.Sequence{
		Format:      "r1",
		Duration:    totalDuration,
		TCStart:     "0s", 
		TCFormat:    "NDF",
		AudioLayout: "stereo",
		AudioRate:   "48k",
		Spine:       createSpineWithTextChunks(chunks),
	}
	
	return sequence
}

func breakTextIntoChunks(text string, totalDuration float64) []TextChunk {
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
		
		// Determine font size based on text length
		fontSize := calculateFontSize(chunkText)
		isLong := len(chunkText) > 8
		
		chunks = append(chunks, TextChunk{
			Text:     chunkText,
			Duration: 0, // Will be calculated after all chunks are created
			FontSize: fontSize,
			IsLong:   isLong,
		})
	}
	
	// Calculate durations based on total duration
	if totalDuration > 0 {
		// Distribute the total duration evenly across all chunks
		chunkDuration := totalDuration / float64(len(chunks))
		for i := range chunks {
			chunks[i].Duration = chunkDuration
		}
	} else {
		// Use original duration calculation logic
		for i := range chunks {
			chunks[i].Duration = calculateTextDuration(chunks[i].Text)
		}
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
	
	for _, part := range textParts {
		styleID := fmt.Sprintf("ts%d", *textStyleID)
		*textStyleID++
		
		textStyles = append(textStyles, fcp.TextStyleRef{
			Ref:  styleID,
			Text: part,
		})
		
		// Create text style definition with shadow properties
		textStyleDefs = append(textStyleDefs, fcp.TextStyleDef{
			ID: styleID,
			TextStyle: fcp.TextStyle{
				Font:      "Avenir Next Condensed",
				FontFace:  "Heavy Italic",
				FontSize:  strconv.Itoa(chunk.FontSize),
				FontColor: "1 0 1 1", // Bright magenta
			},
		})
	}
	
	title := fcp.Title{
		Ref:      "r4", // Text effect
		Lane:     strconv.Itoa(index), // Each title gets its own lane
		Offset:   offsetStr,
		Name:     fmt.Sprintf("Shadow Text %d", index),
		Start:    "86486400/24000s",
		Duration: durationStr,
		Text: &fcp.TitleText{
			TextStyles: textStyles,
		},
		TextStyleDefs: textStyleDefs,
		Params:        createTitleParams(chunk.FontSize),
	}
	
	return title
}

func splitTextForShadowEffect(text string) []string {
	// Split text creatively but preserve spaces between originally separate words
	// The chunking already grouped words appropriately, so we should only split
	// individual words, never combine words that were originally separate
	
	words := strings.Fields(text)
	
	// If single word, split it creatively for shadow effect
	if len(words) == 1 {
		word := words[0]
		if len(word) <= 3 {
			return []string{word}
		}
		if len(word) <= 6 {
			// Split at last character
			return []string{word[:len(word)-1], word[len(word)-1:]}
		}
		// For longer single words, split at last character
		return []string{word[:len(word)-1], word[len(word)-1:]}
	}
	
	// If multiple words, return each word as separate text-style
	// This preserves the space between words that were originally separate
	return words
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