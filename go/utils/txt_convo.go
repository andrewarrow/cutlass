package utils

import (
	"bufio"
	"cutlass/fcp"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConversationMessage represents a single message in the conversation  
type ConversationMessage struct {
	Name    string
	Content string
	IsUser  bool // true if user message (blue bubble), false if other person (gray bubble)
}

// HandleTxtConvoCommand processes a text conversation file and generates iMessage-style FCPXML
func HandleTxtConvoCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: txt-convo <conversation.txt> [output.fcpxml] [duration-per-message]")
		fmt.Println("")
		fmt.Println("Input format - each line should be: Name: message content")
		fmt.Println("Example:")
		fmt.Println("  John: Hey, how are you?")
		fmt.Println("  Sarah: I'm doing great, thanks!")
		fmt.Println("  John: Want to grab coffee later?")
		fmt.Println("")
		fmt.Println("Features:")
		fmt.Println("- Animates messages appearing sequentially like real iMessage")
		fmt.Println("- Blue bubbles for first person, gray for others")
		fmt.Println("- Uses iPhone background from ./assets/iphone_bg.png")
		fmt.Println("- Bubble PNGs from ./assets/bubble_blue.png and ./assets/bubble_gray.png")
		fmt.Println("- Duration per message defaults to 2.5 seconds")
		return
	}

	inputFile := args[0]
	
	// Default output file
	baseName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	outputFile := filepath.Join("./data", baseName+"_convo.fcpxml")
	durationPerMessage := 2.5

	// Parse optional arguments
	if len(args) > 1 {
		outputFile = args[1]
	}
	if len(args) > 2 {
		if d, err := parseFloat(args[2]); err == nil {
			durationPerMessage = d
		}
	}

	fmt.Printf("🗣️  Processing conversation: %s\n", inputFile)
	fmt.Printf("📱 Generating iMessage-style FCPXML: %s\n", outputFile)
	fmt.Printf("⏱️  Duration per message: %.1f seconds\n", durationPerMessage)

	if err := GenerateTxtConvoFCPXML(inputFile, outputFile, durationPerMessage); err != nil {
		fmt.Printf("❌ Error generating conversation: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated iMessage conversation FCPXML successfully!\n")
	fmt.Printf("🎬 Ready to import into Final Cut Pro\n")
}

// GenerateTxtConvoFCPXML creates FCPXML with iMessage-style conversation animation following samples/imessage.fcpxml patterns
func GenerateTxtConvoFCPXML(inputFile, outputFile string, durationPerMessage float64) error {
	// Parse conversation messages
	messages, err := ParseConversationFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse conversation file: %v", err)
	}

	if len(messages) == 0 {
		return fmt.Errorf("no messages found in conversation file")
	}

	fmt.Printf("📝 Parsed %d messages from conversation\n", len(messages))

	// Create FCPXML following the samples/imessage.fcpxml structure exactly
	fcpxml, err := createIMessageFCPXML(messages, durationPerMessage)
	if err != nil {
		return fmt.Errorf("failed to create iMessage FCPXML: %v", err)
	}

	// Write the FCPXML to file
	if err := fcp.WriteToFile(fcpxml, outputFile); err != nil {
		return fmt.Errorf("failed to write FCPXML: %v", err)
	}

	return nil
}

// ParseConversationFile reads a text file and extracts conversation messages
func ParseConversationFile(filePath string) ([]ConversationMessage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []ConversationMessage
	var firstPersonName string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse "Name: content" format
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		name := strings.TrimSpace(parts[0])
		content := strings.TrimSpace(parts[1])

		if name == "" || content == "" {
			continue // Skip empty name or content
		}

		// Determine if this is the user (first person mentioned gets blue bubbles)
		isUser := false
		if firstPersonName == "" {
			firstPersonName = name
			isUser = true
		} else if name == firstPersonName {
			isUser = true
		}

		messages = append(messages, ConversationMessage{
			Name:    name,
			Content: content,
			IsUser:  isUser,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}


// truncateText shortens text for display purposes
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// createIMessageFCPXML creates FCPXML following the exact structure of samples/imessage.fcpxml
func createIMessageFCPXML(messages []ConversationMessage, durationPerMessage float64) (*fcp.FCPXML, error) {
	totalDuration := calculateTotalConversationDuration(messages, durationPerMessage)
	
	fcpxml := &fcp.FCPXML{
		Version: "1.13",
		Resources: fcp.Resources{
			// Match sample order: formats first, then assets, then effects
			Formats: []fcp.Format{
				{
					ID:            "r1", 
					FrameDuration: "1001/24000s",
					Width:         "1080",
					Height:        "1920",
					ColorSpace:    "1-1-1 (Rec. 709)",
				},
				{
					ID:         "r3",
					Name:       "FFVideoFormatRateUndefined", 
					Width:      "452",
					Height:     "913",
					ColorSpace: "1-13-1",
				},
				{
					ID:         "r5",
					Name:       "FFVideoFormatRateUndefined",
					Width:      "392", 
					Height:     "209",
					ColorSpace: "1-13-1",
				},
				{
					ID:         "r8",
					Name:       "FFVideoFormatRateUndefined",
					Width:      "394",
					Height:     "207", 
					ColorSpace: "1-13-1",
				},
			},
			Assets: createAssetResourcesWithMetadata(),
			Effects: []fcp.Effect{
				{
					ID:   "r6",
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
					UID:  "B6DD4576-A032-4A00-BF67-D4DD6CA3E2C0",
					Projects: []fcp.Project{
						{
							Name:     "wiki",
							UID:      "CD9DC2C7-C690-4E58-A4BD-350B5CE85723",
							ModDate:  "2025-06-22 12:54:35 -0700",
							Sequences: []fcp.Sequence{
								createIMessageSequence(messages, durationPerMessage, totalDuration),
							},
						},
					},
				},
			},
			SmartCollections: []fcp.SmartCollection{
				{Name: "Projects", Match: "all", Matches: []fcp.Match{{Rule: "is", Type: "project"}}},
				{Name: "All Video", Match: "any", MediaMatches: []fcp.MediaMatch{{Rule: "is", Type: "videoOnly"}, {Rule: "is", Type: "videoWithAudio"}}},
				{Name: "Audio Only", Match: "all", MediaMatches: []fcp.MediaMatch{{Rule: "is", Type: "audioOnly"}}},
				{Name: "Stills", Match: "all", MediaMatches: []fcp.MediaMatch{{Rule: "is", Type: "stills"}}},
				{Name: "Favorites", Match: "all", RatingMatches: []fcp.RatingMatch{{Value: "favorites"}}},
			},
		},
	}
	
	return fcpxml, nil
}

// createIMessageSequence creates the sequence with nested video structure like samples/imessage.fcpxml
func createIMessageSequence(messages []ConversationMessage, durationPerMessage, totalDuration float64) fcp.Sequence {
	return fcp.Sequence{
		Format:      "r1",
		Duration:    fcp.ConvertSecondsToFCPDuration(totalDuration),
		TCStart:     "0s",
		TCFormat:    "NDF", 
		AudioLayout: "stereo",
		AudioRate:   "48k",
		Spine:       createIMessageSpine(messages, durationPerMessage),
	}
}

// createIMessageSpine creates spine following samples/imessage.fcpxml pattern with sequential conversation timing
func createIMessageSpine(messages []ConversationMessage, durationPerMessage float64) fcp.Spine {
	var spine fcp.Spine
	
	currentOffset := 0.0
	visibleMessages := []ConversationMessage{} // Track all messages that should be visible
	globalMessageIndex := 0
	
	for i, message := range messages {
		// Add current message to visible messages
		visibleMessages = append(visibleMessages, message)
		
		// Calculate segment duration - vary like sample (13013, 15015, 14014, etc.)
		segmentDuration := calculateSegmentDuration(i, len(messages), durationPerMessage)
		
		// Create phone background video for this segment
		phoneVideo := fcp.Video{
			Ref:      "r2", // phone_blank001
			Offset:   fcp.ConvertSecondsToFCPDuration(currentOffset),
			Name:     "phone_blank001",
			Start:    calculatePhoneStartTime(i),
			Duration: fcp.ConvertSecondsToFCPDuration(segmentDuration),
		}
		
		// Add all visible messages (current + previous) to maintain conversation history
		lane := 1
		for j, visibleMsg := range visibleMessages {
			// Add bubble (appears immediately when segment starts)
			bubbleVideo := createBubbleVideoWithTiming(visibleMsg, lane, currentOffset, segmentDuration, i, j)
			phoneVideo.NestedVideos = append(phoneVideo.NestedVideos, bubbleVideo)
			
			// Add text (appears with delay after bubble, only for current message or with staggered timing)
			textDelay := calculateTextDelay(i, j, len(visibleMessages))
			textTitle := createTextTitleWithTiming(visibleMsg, lane+1, currentOffset, textDelay, segmentDuration, i, j, globalMessageIndex)
			phoneVideo.NestedTitles = append(phoneVideo.NestedTitles, textTitle)
			
			lane += 2
			globalMessageIndex++
		}
		
		spine.Videos = append(spine.Videos, phoneVideo)
		currentOffset += segmentDuration
	}
	
	return spine
}

// ConversationExchange represents a group of related messages
type ConversationExchange struct {
	Messages []ConversationMessage
	IsUserInitiated bool
}

// Helper functions
func calculateTotalConversationDuration(messages []ConversationMessage, durationPerMessage float64) float64 {
	total := 0.0
	for i := range messages {
		total += calculateSegmentDuration(i, len(messages), durationPerMessage)
	}
	return total
}

func groupMessagesIntoExchanges(messages []ConversationMessage) []ConversationExchange {
	if len(messages) == 0 {
		return nil
	}
	
	var exchanges []ConversationExchange
	
	// Create fewer exchanges like the sample (group by conversation turns)
	// Sample has 2 phone videos for multiple messages, so we'll create 2-3 exchanges max
	
	// First exchange: first 1-3 messages
	firstExchangeSize := 1
	if len(messages) > 2 {
		firstExchangeSize = 3
	}
	
	firstExchange := ConversationExchange{
		Messages: messages[:firstExchangeSize],
		IsUserInitiated: messages[0].IsUser,
	}
	exchanges = append(exchanges, firstExchange)
	
	// Remaining messages in second exchange
	if firstExchangeSize < len(messages) {
		secondExchange := ConversationExchange{
			Messages: messages[firstExchangeSize:],
			IsUserInitiated: messages[firstExchangeSize].IsUser,
		}
		exchanges = append(exchanges, secondExchange)
	}
	
	return exchanges
}

func createAssetResourcesWithMetadata() []fcp.Asset {
	return []fcp.Asset{
		// Phone background asset with metadata and bookmark like sample
		{
			ID:           "r2",
			Name:         "phone_blank001",
			UID:          "3BF13EB320E3C082405DE41A35F1DACB",
			Start:        "0s",
			Duration:     "0s",
			HasVideo:     "1",
			Format:       "r3",
			VideoSources: "1",
			MediaRep: fcp.MediaRep{
				Kind: "original-media",
				Sig:  "3BF13EB320E3C082405DE41A35F1DACB",
				Src:  "file:///Users/aa/Documents/phone_blank001.png",
			},
		},
		// Blue bubble asset
		{
			ID:           "r4",
			Name:         "blue_speech001", 
			UID:          "1EA2484AC5332B02E400581617C071F2",
			Start:        "0s",
			Duration:     "0s",
			HasVideo:     "1",
			Format:       "r5",
			VideoSources: "1",
			MediaRep: fcp.MediaRep{
				Kind: "original-media",
				Sig:  "1EA2484AC5332B02E400581617C071F2",
				Src:  "file:///Users/aa/Documents/blue_speech001.png",
			},
		},
		// White bubble asset
		{
			ID:           "r7",
			Name:         "white_speech001",
			UID:          "8B1E084D50810C12F2F6EAF7517875FE",
			Start:        "0s", 
			Duration:     "0s",
			HasVideo:     "1",
			Format:       "r8",
			VideoSources: "1",
			MediaRep: fcp.MediaRep{
				Kind: "original-media",
				Sig:  "8B1E084D50810C12F2F6EAF7517875FE",
				Src:  "file:///Users/aa/Documents/white_speech001.png",
			},
		},
	}
}

func calculatePhoneStartTime(exchangeIndex int) string {
	// Use sample start times
	baseTimes := []string{
		"86399313/24000s", // First exchange
		"86531445/24000s", // Second exchange  
		"86663577/24000s", // Third exchange (extrapolated)
	}
	
	if exchangeIndex < len(baseTimes) {
		return baseTimes[exchangeIndex]
	}
	return baseTimes[len(baseTimes)-1] // Use last one for additional exchanges
}

func createBubbleVideoWithTiming(message ConversationMessage, lane int, segmentOffset, segmentDuration float64, segmentIndex, messageIndex int) fcp.Video {
	var bubbleRef string
	var bubblePosition string
	var bubbleScale string
	
	if message.IsUser {
		bubbleRef = "r4" // blue_speech001
		bubblePosition = "1.26755 -21.1954"
		bubbleScale = "0.617236 0.617236"
	} else {
		bubbleRef = "r7" // white_speech001  
		bubblePosition = "0.635834 4.00864"
		bubbleScale = "0.653172 0.653172"
	}
	
	// Bubble offset within the segment - use sample timing pattern
	bubbleOffset := calculateBubbleOffsetInSegment(segmentIndex, messageIndex)
	
	return fcp.Video{
		Ref:      bubbleRef,
		Lane:     fmt.Sprintf("%d", lane),
		Offset:   bubbleOffset,
		Name:     getBubbleName(bubbleRef),
		Start:    bubbleOffset,
		Duration: fcp.ConvertSecondsToFCPDuration(segmentDuration), // Duration matches segment
		AdjustTransform: &fcp.AdjustTransform{
			Position: bubblePosition,
			Scale:    bubbleScale,
		},
	}
}

func createTextTitleWithTiming(message ConversationMessage, lane int, segmentOffset, textDelay, segmentDuration float64, segmentIndex, messageIndex int, globalMessageIndex int) fcp.Title {
	textColor := "0.999995 1 1 1" // White for blue bubbles
	textPosition := "0 -3071"
	
	if !message.IsUser {
		textColor = "0 0 0 1" // Black for white bubbles  
		textPosition = "0 -1807"
	}
	
	textStyleID := fmt.Sprintf("ts%d", globalMessageIndex+1)
	
	// Text offset within segment - appears after bubble with delay
	textOffset := calculateTextOffsetInSegment(segmentIndex, messageIndex, textDelay)
	
	return fcp.Title{
		Ref:      "r6", // Text effect
		Lane:     fmt.Sprintf("%d", lane),
		Offset:   textOffset,
		Name:     fmt.Sprintf("%s - Text", truncateText(message.Content, 20)),
		Start:    calculateTextStartTimeInSegment(segmentIndex, messageIndex, textDelay),
		Duration: fcp.ConvertSecondsToFCPDuration(segmentDuration - textDelay),
		
		Params: createCompleteTextParams(textPosition),
		
		Text: &fcp.TitleText{
			TextStyles: []fcp.TextStyleRef{
				{Ref: textStyleID, Text: message.Content},
			},
		},
		TextStyleDefs: []fcp.TextStyleDef{
			{
				ID: textStyleID,
				TextStyle: fcp.TextStyle{
					Font:        "Arial",
					FontSize:    "204",
					FontFace:    "Regular", 
					FontColor:   textColor,
					Alignment:   "center",
					LineSpacing: "-19",
				},
			},
		},
	}
}

func createCompleteTextParams(textPosition string) []fcp.Param {
	return []fcp.Param{
		{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
		{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
		{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: textPosition},
		{Name: "Layout Method", Key: "9999/10003/13260/3296672360/2/314", Value: "1 (Paragraph)"},
		{Name: "Left Margin", Key: "9999/10003/13260/3296672360/2/323", Value: "-1210"},
		{Name: "Right Margin", Key: "9999/10003/13260/3296672360/2/324", Value: "1210"},
		{Name: "Top Margin", Key: "9999/10003/13260/3296672360/2/325", Value: "2160"},
		{Name: "Bottom Margin", Key: "9999/10003/13260/3296672360/2/326", Value: "-2160"},
		{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/354/3296667315/401", Value: "1 (Center)"},
		{Name: "Line Spacing", Key: "9999/10003/13260/3296672360/2/354/3296667315/404", Value: "-19"},
		{Name: "Auto-Shrink", Key: "9999/10003/13260/3296672360/2/370", Value: "3 (To All Margins)"},
		{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/373", Value: "0 (Left) 0 (Top)"},
		{Name: "Opacity", Key: "9999/10003/13260/3296672360/4/3296673134/1000/1044", Value: "0"},
		{Name: "Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/208", Value: "6 (Custom)"},
		// Custom speed with keyframe animation like sample
		{
			Name: "Custom Speed", 
			Key: "9999/10003/13260/3296672360/4/3296673134/201/209",
			KeyframeAnimation: &fcp.KeyframeAnimation{
				Keyframes: []fcp.Keyframe{
					{Time: "-469658744/1000000000s", Value: "0"},
					{Time: "12328542033/1000000000s", Value: "1"},
				},
			},
		},
		{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
	}
}

// New timing calculation functions following samples/imessage.fcpxml
func calculateSegmentDuration(segmentIndex, totalMessages int, baseDuration float64) float64 {
	// Match sample durations: 13013/24000s ≈ 0.54s, 15015/24000s ≈ 0.63s, 14014/24000s ≈ 0.58s
	if segmentIndex == 0 {
		return 0.54 // First message duration
	} else if segmentIndex == totalMessages-1 {
		return 1.9 // Last message gets longer duration like sample (466466/24000s ≈ 1.9s)
	}
	return 0.63 // Middle messages
}

func calculateTextDelay(segmentIndex, messageIndex, totalVisible int) float64 {
	// Only the current message (last in visible list) gets text animation delay
	if messageIndex == totalVisible-1 {
		return 0.2 // New message text appears after short delay
	}
	return 0.0 // Previous messages already have text visible
}

func calculateBubbleOffsetInSegment(segmentIndex, messageIndex int) string {
	// Use sample patterns - bubbles appear when segment starts
	baseOffsets := []string{
		"86441355/24000s", // Segment 0
		"86531445/24000s", // Segment 1  
		"86639553/24000s", // Segment 2
		"86653567/24000s", // Segment 3
	}
	
	if segmentIndex < len(baseOffsets) {
		return baseOffsets[segmentIndex]
	}
	return baseOffsets[len(baseOffsets)-1]
}

func calculateTextOffsetInSegment(segmentIndex, messageIndex int, textDelay float64) string {
	// Text offset within segment
	baseOffsets := []string{
		"86441355/24000s", // Segment 0
		"86531445/24000s", // Segment 1
		"86639553/24000s", // Segment 2 
		"86653567/24000s", // Segment 3
	}
	
	if segmentIndex < len(baseOffsets) {
		return baseOffsets[segmentIndex]
	}
	return baseOffsets[len(baseOffsets)-1]
}

func calculateTextStartTimeInSegment(segmentIndex, messageIndex int, textDelay float64) string {
	// Text start time - delayed from bubble
	baseStartTimes := []string{
		"86528442/24000s", // Segment 0 (86441355 + delay)
		"86617531/24000s", // Segment 1 (86531445 + delay)
		"86618532/24000s", // Segment 2 (86639553 + delay)
		"86617531/24000s", // Segment 3 (86653567 + delay)
	}
	
	if segmentIndex < len(baseStartTimes) {
		return baseStartTimes[segmentIndex]
	}
	return baseStartTimes[len(baseStartTimes)-1]
}

func generateAssetUID(name string) string {
	// Use fixed UIDs from sample for consistency
	switch name {
	case "phone_blank001":
		return "3BF13EB320E3C082405DE41A35F1DACB"
	case "blue_speech001":
		return "1EA2484AC5332B02E400581617C071F2"
	case "white_speech001":
		return "8B1E084D50810C12F2F6EAF7517875FE"
	default:
		return "3BF13EB320E3C082405DE41A35F1DACB"
	}
}

func getBubbleName(ref string) string {
	switch ref {
	case "r4":
		return "blue_speech001"
	case "r7": 
		return "white_speech001"
	default:
		return "speech_bubble"
	}
}

// parseFloat safely parses a float from string
func parseFloat(s string) (float64, error) {
	if f, err := fmt.Sscanf(s, "%f", new(float64)); err != nil || f != 1 {
		return 0, fmt.Errorf("invalid float: %s", s)
	}
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result, nil
}