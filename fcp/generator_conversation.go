package fcp

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GenerateConversation creates an FCPXML file with an iMessage-style conversation using alternating speech bubbles.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management  
// - Uses STRUCTS ONLY - no string templates → proper XML marshaling via struct fields
// - Atomic ID reservation prevents race conditions and ID collisions
// - Frame-aligned durations → ConvertSecondsToFCPDuration() function
// - Images use Video elements, not AssetClip elements (CRITICAL to prevent crashes)
//
// Conversation Pattern (based on Info.fcpxml analysis):
// 1. Multiple phone background segments on spine (each segment = 2 seconds)
// 2. Each phone background has connected speech bubbles and text (all appear simultaneously)
// 3. Sequential conversation created by sequential spine segments, not connected clip timing
// 4. Each segment contains one blue+white message pair with proper positioning
//
// Required Files:
// - phoneBackgroundPath: PNG of phone background (e.g., phone_blank001.png)
// - blueSpeechPath: PNG of blue speech bubble (e.g., blue_speech001.png)
// - whiteSpeechPath: PNG of white speech bubble (e.g., white_speech001.png)
// - messagesPath: Text file with alternating lines (line 1: blue, line 2: white, line 3: blue, etc.)
//
// ❌ NEVER: fmt.Sprintf("<video ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func GenerateConversation(phoneBackgroundPath, blueSpeechPath, whiteSpeechPath, messagesPath, outputPath string) error {
	// Read messages from file
	messages, err := readMessagesFromFile(messagesPath)
	if err != nil {
		return fmt.Errorf("failed to read messages: %v", err)
	}

	if len(messages) == 0 {
		return fmt.Errorf("no messages found in file: %s", messagesPath)
	}

	// Create empty FCPXML structure
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		return fmt.Errorf("failed to create base FCPXML: %v", err)
	}

	// Calculate dynamic conversation duration based on message lengths
	totalDurationFCP := calculateTotalDuration(messages)

	// Update sequence duration
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Duration = totalDurationFCP
	}

	// Process messages in pairs (each pair gets its own phone background segment)
	var cumulativeOffsetFrames int64 = 0
	
	for i := 0; i < len(messages); i += 2 {
		// Get messages for this pair
		var message1, message2 string
		if i < len(messages) {
			message1 = messages[i]
		}
		if i+1 < len(messages) {
			message2 = messages[i+1]
		}
		
		// Calculate dynamic duration for this segment based on message lengths
		segmentDurationFCP := calculateSegmentDuration(message1, message2)
		
		// Calculate offset for this segment
		var segmentOffsetFCP string
		if cumulativeOffsetFrames == 0 {
			segmentOffsetFCP = "0s"
		} else {
			segmentOffsetFCP = fmt.Sprintf("%d/24000s", cumulativeOffsetFrames)
		}
		
		// Parse segment duration to add to cumulative offset for next iteration
		var segmentFrames int64
		if n, err := fmt.Sscanf(segmentDurationFCP, "%d/24000s", &segmentFrames); n == 1 && err == nil {
			// We'll add this to cumulative offset at the end of the loop
		}

		// Add phone background for this segment
		err = addPhoneBackgroundSegmentExact(fcpxml, phoneBackgroundPath, segmentOffsetFCP, segmentDurationFCP)
		if err != nil {
			return fmt.Errorf("failed to add phone background segment %d: %v", i/2, err)
		}

		// Add messages for this segment (white and blue pair, like Info.fcpxml)
		// All elements in this segment will use the same segmentDurationFCP for sync
		if i+1 < len(messages) {
			// Add white speech bubble with text using segment duration
			err = addMessageToLastSegmentWithIndexAndDuration(fcpxml, whiteSpeechPath, messages[i+1], false, i+1, segmentDurationFCP)
			if err != nil {
				return fmt.Errorf("failed to add white message %d: %v", i+1, err)
			}
		}
		
		if i < len(messages) {
			// Add blue speech bubble with text using segment duration
			err = addMessageToLastSegmentWithIndexAndDuration(fcpxml, blueSpeechPath, messages[i], true, i, segmentDurationFCP)
			if err != nil {
				return fmt.Errorf("failed to add blue message %d: %v", i, err)
			}
		}
		
		// Update cumulative offset for next segment
		cumulativeOffsetFrames += segmentFrames
	}

	// Write FCPXML to output file
	err = WriteToFile(fcpxml, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write FCPXML: %v", err)
	}

	return nil
}

// readMessagesFromFile reads a text file and returns a slice of message strings
func readMessagesFromFile(messagesPath string) ([]string, error) {
	file, err := os.Open(messagesPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			messages = append(messages, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// calculateFontSize determines appropriate font size based on message length
// Short messages (1-10 chars): 204 (original size)
// Medium messages (11-30 chars): 180-150
// Long messages (31+ chars): 120-90
func calculateFontSize(message string) string {
	length := len(message)
	
	if length <= 10 {
		// Very short messages - keep original size
		return "204"
	} else if length <= 20 {
		// Short messages - slightly smaller
		return "180"
	} else if length <= 35 {
		// Medium messages - smaller
		return "150"
	} else if length <= 50 {
		// Long messages - much smaller
		return "120"
	} else {
		// Very long messages - smallest readable size
		return "90"
	}
}

// calculateLineSpacing adjusts line spacing based on font size for better readability
func calculateLineSpacing(fontSize string) string {
	switch fontSize {
	case "204":
		return "-19"  // Original spacing
	case "180":
		return "-15"
	case "150":
		return "-12"
	case "120":
		return "-8"
	case "90":
		return "-5"
	default:
		return "-19"
	}
}

// formatMessageWithLineBreaks adds line breaks after the 2nd word for messages with 3+ words
func formatMessageWithLineBreaks(message string) string {
	words := strings.Fields(message)
	
	// If 3 or more words, insert line break after 2nd word
	if len(words) >= 3 {
		// Join first 2 words, add newline, then join remaining words
		firstLine := strings.Join(words[:2], " ")
		secondLine := strings.Join(words[2:], " ")
		return firstLine + "\n" + secondLine
	}
	
	// For 1-2 words, return as-is
	return message
}

// calculateYPosition adjusts Y position based on whether text is multi-line
func calculateYPosition(message string, isBlue bool) string {
	words := strings.Fields(message)
	isMultiLine := len(words) >= 3
	
	if isBlue {
		// Blue text (bottom bubble)
		if isMultiLine {
			return "0 -3000"  // Move UP for 2-line text to center in bubble (less negative = up)
		}
		return "0 -3071"  // Original position for single line
	} else {
		// White text (top bubble)  
		if isMultiLine {
			return "0 -1800"  // Move UP for 2-line text to center in bubble (less negative = up)
		}
		return "0 -1807"  // Original position for single line
	}
}

// calculateSegmentDuration determines how long to display each message pair based on the longest message
// All elements in the segment (phone, bubbles, text) need to use this same duration
func calculateSegmentDuration(message1, message2 string) string {
	// Calculate duration needed for each message individually
	duration1 := calculateTextDuration(message1)
	duration2 := calculateTextDuration(message2)
	
	// Use the longer of the two durations for the entire segment
	frames1 := parseDurationToFrames(duration1)
	frames2 := parseDurationToFrames(duration2)
	
	maxFrames := frames1
	if frames2 > maxFrames {
		maxFrames = frames2
	}
	
	return fmt.Sprintf("%d/24000s", maxFrames)
}

// parseDurationToFrames converts duration string to frame count
func parseDurationToFrames(duration string) int64 {
	var frames int64
	if n, err := fmt.Sscanf(duration, "%d/24000s", &frames); n == 1 && err == nil {
		return frames
	}
	return 15015 // fallback to default
}

// calculateTotalDuration calculates the total sequence duration based on all message pairs
func calculateTotalDuration(messages []string) string {
	var totalDurationFrames int64 = 0
	
	// Process messages in pairs
	for i := 0; i < len(messages); i += 2 {
		var message1, message2 string
		
		if i < len(messages) {
			message1 = messages[i]
		}
		if i+1 < len(messages) {
			message2 = messages[i+1]
		}
		
		// Get duration for this pair and convert to frame count
		durationStr := calculateSegmentDuration(message1, message2)
		
		// Parse the fraction (e.g., "24024/24000s" -> 24024 frames)
		var frames int64
		if n, err := fmt.Sscanf(durationStr, "%d/24000s", &frames); n == 1 && err == nil {
			totalDurationFrames += frames
		}
	}
	
	return fmt.Sprintf("%d/24000s", totalDurationFrames)
}

// calculateTextDuration determines how long individual text should stay on screen based on message length
func calculateTextDuration(message string) string {
	length := len(message)
	
	if length <= 5 {
		// Very short text - quick display
		return "15015/24000s"  // ~0.625s
	} else if length <= 15 {
		// Short text - normal display
		return "24024/24000s"  // ~1.0s
	} else if length <= 30 {
		// Medium text - longer display
		return "36036/24000s"  // ~1.5s
	} else if length <= 50 {
		// Long text - much longer display
		return "48048/24000s"  // ~2.0s
	} else {
		// Very long text - longest display for readability
		return "60060/24000s"  // ~2.5s
	}
}

// addPhoneBackgroundSegment adds a phone background video segment directly to the spine
func addPhoneBackgroundSegment(fcpxml *FCPXML, phoneBackgroundPath string, offsetSeconds, durationSeconds float64) error {
	// Use existing AddImage logic but with specific offset and duration
	err := AddImage(fcpxml, phoneBackgroundPath, durationSeconds)
	if err != nil {
		return fmt.Errorf("failed to add phone background: %v", err)
	}

	// Update the last added phone background with correct offset
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		if len(sequence.Spine.Videos) > 0 {
			lastVideo := &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]
			lastVideo.Offset = ConvertSecondsToFCPDuration(offsetSeconds)
			lastVideo.Duration = ConvertSecondsToFCPDuration(durationSeconds)
		}
	}

	return nil
}

// addPhoneBackgroundSegmentExact adds a phone background video segment directly to the spine with exact FCP timing
func addPhoneBackgroundSegmentExact(fcpxml *FCPXML, phoneBackgroundPath string, offsetFCP, durationFCP string) error {
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Check if asset already exists for this file
	var phoneAsset *Asset
	if asset, exists := registry.GetOrCreateAsset(phoneBackgroundPath); exists {
		phoneAsset = asset
	} else {
		// Create transaction for atomic resource creation
		tx := NewTransaction(registry)

		// Reserve IDs atomically to prevent collisions (need 2: asset + format)
		ids := tx.ReserveIDs(2)
		assetID := ids[0]
		formatID := ids[1]

		// Generate unique asset name
		phoneName := strings.TrimSuffix(strings.TrimSuffix(phoneBackgroundPath, ".png"), ".jpg")

		// Create image-specific format using transaction
		// 🚨 CRITICAL: Image formats must NOT have frameDuration (causes crashes)
		_, err := tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1280", "720", "1-13-1")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create phone background format: %v", err)
		}

		// Create asset using transaction
		asset, err := tx.CreateAsset(assetID, phoneBackgroundPath, phoneName, "0s", formatID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create phone background asset: %v", err)
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit phone background transaction: %v", err)
		}

		phoneAsset = asset
	}

	// Create Video element with exact FCP timing (not using AddImage's timeline calculation)
	video := Video{
		Ref:      phoneAsset.ID,
		Offset:   offsetFCP,           // Use exact offset instead of timeline calculation
		Name:     phoneAsset.Name,
		Start:    "86531445/24000s",   // Match Info.fcpxml start time exactly
		Duration: durationFCP,         // Use exact duration
	}

	// Add Video element directly to spine
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Spine.Videos = append(sequence.Spine.Videos, video)
	}

	return nil
}

// addSpeechBubbleToLastSegment adds just a speech bubble to the most recently added phone background segment
func addSpeechBubbleToLastSegment(fcpxml *FCPXML, speechBubblePath string, isBlue bool) error {
	// Get the last phone background segment
	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return fmt.Errorf("no sequence found")
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) == 0 {
		return fmt.Errorf("no phone background found")
	}

	phoneVideo := &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]

	// Add speech bubble as connected clip
	err := addSpeechBubbleToSegment(fcpxml, phoneVideo, speechBubblePath, isBlue)
	if err != nil {
		return fmt.Errorf("failed to add speech bubble: %v", err)
	}

	return nil
}

// addMessageToLastSegment adds a speech bubble and text to the most recently added phone background segment
func addMessageToLastSegment(fcpxml *FCPXML, speechBubblePath, message string, isBlue bool) error {
	return addMessageToLastSegmentWithIndex(fcpxml, speechBubblePath, message, isBlue, 0)
}

// addMessageToLastSegmentWithIndex adds a speech bubble and text to the most recently added phone background segment with a unique index
func addMessageToLastSegmentWithIndex(fcpxml *FCPXML, speechBubblePath, message string, isBlue bool, index int) error {
	return addMessageToLastSegmentWithIndexAndDuration(fcpxml, speechBubblePath, message, isBlue, index, "15015/24000s")
}

// addMessageToLastSegmentWithIndexAndDuration adds a speech bubble and text with segment duration for sync
func addMessageToLastSegmentWithIndexAndDuration(fcpxml *FCPXML, speechBubblePath, message string, isBlue bool, index int, segmentDuration string) error {
	// Get the last phone background segment
	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return fmt.Errorf("no sequence found")
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) == 0 {
		return fmt.Errorf("no phone background found")
	}

	phoneVideo := &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]

	// Use the segment duration for both bubble and text to keep all elements in sync
	// segmentDuration is calculated based on the longest message in this segment pair
	
	// Add speech bubble as connected clip with segment duration
	err := addSpeechBubbleToSegmentWithDuration(fcpxml, phoneVideo, speechBubblePath, isBlue, segmentDuration)
	if err != nil {
		return fmt.Errorf("failed to add speech bubble: %v", err)
	}

	// Add text as connected clip with same segment duration
	err = addTextToSegmentWithIndexAndDuration(fcpxml, phoneVideo, message, isBlue, index, segmentDuration)
	if err != nil {
		return fmt.Errorf("failed to add text: %v", err)
	}

	return nil
}

// addSpeechBubbleToSegment adds a speech bubble as a connected clip to a phone background video
func addSpeechBubbleToSegment(fcpxml *FCPXML, phoneVideo *Video, speechBubblePath string, isBlue bool) error {
	return addSpeechBubbleToSegmentWithDuration(fcpxml, phoneVideo, speechBubblePath, isBlue, "15015/24000s")
}

// addSpeechBubbleToSegmentWithDuration adds a speech bubble with custom duration
func addSpeechBubbleToSegmentWithDuration(fcpxml *FCPXML, phoneVideo *Video, speechBubblePath string, isBlue bool, duration string) error {
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Check if asset already exists for this file
	var bubbleAsset *Asset
	if asset, exists := registry.GetOrCreateAsset(speechBubblePath); exists {
		bubbleAsset = asset
	} else {
		// Create transaction for atomic resource creation
		tx := NewTransaction(registry)

		// Reserve IDs atomically to prevent collisions (need 2: asset + format)
		ids := tx.ReserveIDs(2)
		assetID := ids[0]
		formatID := ids[1]

		// Generate unique asset name
		bubbleName := strings.TrimSuffix(strings.TrimSuffix(speechBubblePath, ".png"), ".jpg")

		// Create image-specific format using transaction
		// 🚨 CRITICAL: Image formats must NOT have frameDuration (causes crashes)
		_, err := tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "392", "206", "1-13-1")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create speech bubble format: %v", err)
		}

		// Create asset using transaction
		asset, err := tx.CreateAsset(assetID, speechBubblePath, bubbleName, "0s", formatID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create speech bubble asset: %v", err)
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit speech bubble transaction: %v", err)
		}

		bubbleAsset = asset
	}

	// Create speech bubble video element with proper positioning
	var position, scale string
	var lane string
	if isBlue {
		position = "1.26755 -21.1954"
		scale = "0.617236 0.617236"
		lane = "1"
	} else {
		position = "0.635834 4.00864"
		scale = "0.653172 0.653172"
		lane = "2"
	}

	// Create connected video element (based on Info.fcpxml pattern)
	bubbleVideo := Video{
		Ref:      bubbleAsset.ID,
		Lane:     lane,
		Offset:   "86531445/24000s", // Standard FCP offset for connected clips (matches Info.fcpxml)
		Name:     bubbleAsset.Name,
		Start:    "86531445/24000s", // Standard FCP start time
		Duration: duration,          // Dynamic duration based on message length
		AdjustTransform: &AdjustTransform{
			Position: position,
			Scale:    scale,
		},
	}

	// Add as connected clip to phone background
	phoneVideo.NestedVideos = append(phoneVideo.NestedVideos, bubbleVideo)

	return nil
}

// addTextToSegment adds text as a connected clip to a phone background video
func addTextToSegment(fcpxml *FCPXML, phoneVideo *Video, message string, isBlue bool) error {
	return addTextToSegmentWithIndex(fcpxml, phoneVideo, message, isBlue, 0)
}

// addTextToSegmentWithIndex adds text as a connected clip to a phone background video with a unique index
func addTextToSegmentWithIndex(fcpxml *FCPXML, phoneVideo *Video, message string, isBlue bool, index int) error {
	return addTextToSegmentWithIndexAndDuration(fcpxml, phoneVideo, message, isBlue, index, "15015/24000s")
}

// addTextToSegmentWithIndexAndDuration adds text as a connected clip with specified segment duration
func addTextToSegmentWithIndexAndDuration(fcpxml *FCPXML, phoneVideo *Video, message string, isBlue bool, index int, segmentDuration string) error {
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Check if text effect already exists, if not create it
	textEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if strings.Contains(effect.UID, "Text.moti") {
			textEffectID = effect.ID
			break
		}
	}

	if textEffectID == "" {
		// Create transaction for text effect
		tx := NewTransaction(registry)
		ids := tx.ReserveIDs(1)
		textEffectID = ids[0]

		// Create text effect using transaction
		_, err := tx.CreateEffect(textEffectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create text effect: %v", err)
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit text effect: %v", err)
		}
	}

	// Format message with line breaks if needed
	formattedMessage := formatMessageWithLineBreaks(message)
	
	// Calculate text positioning and color based on speech bubble type  
	var lane string
	var textColor string
	if isBlue {
		lane = "4"
		textColor = "0.999995 1 1 1" // White text for blue bubble
	} else {
		lane = "3"
		textColor = "0 0 0 1" // Black text for white bubble
	}
	
	// Calculate Y position based on whether text is multi-line
	textPosition := calculateYPosition(message, isBlue)

	// Generate unique text-style-def ID with index to avoid duplicates
	baseName := fmt.Sprintf("conversation_%t_%d", isBlue, index)
	textStyleID := GenerateTextStyleID(message, baseName)
	
	// Calculate dynamic font size and spacing based on message length
	fontSize := calculateFontSize(message)
	lineSpacing := calculateLineSpacing(fontSize)

	// Calculate start time based on text color (from Info.fcpxml pattern)
	var startTime string
	if isBlue {
		startTime = "86618532/24000s" // Blue text starts slightly later
	} else {
		startTime = "86617531/24000s" // White text starts first
	}

	// Create Title element matching Info.fcpxml pattern
	title := Title{
		Ref:      textEffectID,
		Lane:     lane,
		Offset:   "86531445/24000s", // Standard FCP offset for connected clips (matches Info.fcpxml)
		Name:     fmt.Sprintf("%s - Text", message),
		Start:    startTime,         // Different start times for blue vs white
		Duration: segmentDuration,   // Segment duration for sync with all elements
		Params: []Param{
			{
				Name:  "Build In",
				Key:   "9999/10000/2/101",
				Value: "0",
			},
			{
				Name:  "Build Out",
				Key:   "9999/10000/2/102",
				Value: "0",
			},
			{
				Name:  "Position",
				Key:   "9999/10003/13260/3296672360/1/100/101",
				Value: textPosition,
			},
			{
				Name:  "Layout Method",
				Key:   "9999/10003/13260/3296672360/2/314",
				Value: "1 (Paragraph)",
			},
			{
				Name:  "Left Margin",
				Key:   "9999/10003/13260/3296672360/2/323",
				Value: "-1210",
			},
			{
				Name:  "Right Margin",
				Key:   "9999/10003/13260/3296672360/2/324",
				Value: "1210",
			},
			{
				Name:  "Top Margin",
				Key:   "9999/10003/13260/3296672360/2/325",
				Value: "2160",
			},
			{
				Name:  "Bottom Margin",
				Key:   "9999/10003/13260/3296672360/2/326",
				Value: "-2160",
			},
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
				Value: "1 (Center)",
			},
			{
				Name:  "Line Spacing",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/404",
				Value: lineSpacing,
			},
			{
				Name:  "Auto-Shrink",
				Key:   "9999/10003/13260/3296672360/2/370",
				Value: "3 (To All Margins)",
			},
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/373",
				Value: "0 (Left) 0 (Top)",
			},
			{
				Name:  "Opacity",
				Key:   "9999/10003/13260/3296672360/4/3296673134/1000/1044",
				Value: "0",
			},
			{
				Name:  "Speed",
				Key:   "9999/10003/13260/3296672360/4/3296673134/201/208",
				Value: "6 (Custom)",
			},
			{
				Name: "Custom Speed",
				Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  "-469658744/1000000000s",
							Value: "0",
						},
						{
							Time:  "12328542033/1000000000s",
							Value: "1",
						},
					},
				},
			},
			{
				Name:  "Apply Speed",
				Key:   "9999/10003/13260/3296672360/4/3296673134/201/211",
				Value: "2 (Per Object)",
			},
		},
		Text: &TitleText{
			TextStyles: []TextStyleRef{
				{
					Ref:  textStyleID,
					Text: formattedMessage,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: textStyleID,
				TextStyle: TextStyle{
					Font:        "Arial",
					FontSize:    fontSize,
					FontFace:    "Regular",
					FontColor:   textColor,
					Alignment:   "center",
					LineSpacing: lineSpacing,
				},
			},
		},
	}

	// Add as connected clip to phone background
	phoneVideo.NestedTitles = append(phoneVideo.NestedTitles, title)

	return nil
}
