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

	// Calculate conversation duration: exact FCP timing like Info.fcpxml
	// Info.fcpxml total duration is 30030/24000s for 2 pairs
	totalDurationFCP := "30030/24000s"

	// Update sequence duration
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Duration = totalDurationFCP
	}

	// Process messages in pairs (each pair gets its own phone background segment)
	for i := 0; i < len(messages); i += 2 {
		// Calculate offset for this phone background segment (matching Info.fcpxml pattern)
		// Info.fcpxml uses 0s, 15015/24000s (exactly 0.625625s intervals)
		segmentDurationFCP := "15015/24000s"  // Use exact FCP duration from Info.fcpxml
		segmentOffsetFCP := ""
		pairIndex := i / 2  // 0 for first pair, 1 for second pair, etc.
		if pairIndex == 0 {
			segmentOffsetFCP = "0s"
		} else if pairIndex == 1 {
			segmentOffsetFCP = "15015/24000s"  // Second segment offset matches first segment duration
		} else {
			// For more pairs, multiply by the segment duration
			segmentOffsetFCP = fmt.Sprintf("%d/24000s", 15015*pairIndex)
		}

		// Add phone background for this segment
		err = addPhoneBackgroundSegmentExact(fcpxml, phoneBackgroundPath, segmentOffsetFCP, segmentDurationFCP)
		if err != nil {
			return fmt.Errorf("failed to add phone background segment %d: %v", i/2, err)
		}

		// Add messages for this segment (white and blue pair, like Info.fcpxml)
		if i+1 < len(messages) {
			// Add white speech bubble with text
			err = addMessageToLastSegmentWithIndex(fcpxml, whiteSpeechPath, messages[i+1], false, i+1)
			if err != nil {
				return fmt.Errorf("failed to add white message %d: %v", i+1, err)
			}
		}
		
		if i < len(messages) {
			// Add blue speech bubble with text
			err = addMessageToLastSegmentWithIndex(fcpxml, blueSpeechPath, messages[i], true, i)
			if err != nil {
				return fmt.Errorf("failed to add blue message %d: %v", i, err)
			}
		}
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

	// Add text as connected clip
	err = addTextToSegmentWithIndex(fcpxml, phoneVideo, message, isBlue, index)
	if err != nil {
		return fmt.Errorf("failed to add text: %v", err)
	}

	return nil
}

// addSpeechBubbleToSegment adds a speech bubble as a connected clip to a phone background video
func addSpeechBubbleToSegment(fcpxml *FCPXML, phoneVideo *Video, speechBubblePath string, isBlue bool) error {
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
		Duration: "15015/24000s",    // Standard FCP duration (~0.6s)
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

	// Calculate text positioning and color based on speech bubble type  
	var textPosition string
	var lane string
	var textColor string
	if isBlue {
		textPosition = "0 -3071"
		lane = "4"
		textColor = "0.999995 1 1 1" // White text for blue bubble
	} else {
		textPosition = "0 -1807"
		lane = "3"
		textColor = "0 0 0 1" // Black text for white bubble
	}

	// Generate unique text-style-def ID with index to avoid duplicates
	baseName := fmt.Sprintf("conversation_%t_%d", isBlue, index)
	textStyleID := GenerateTextStyleID(message, baseName)

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
		Duration: "15015/24000s",    // Standard FCP duration (~0.6s)
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
				Value: "-19",
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
					Text: message,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: textStyleID,
				TextStyle: TextStyle{
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

	// Add as connected clip to phone background
	phoneVideo.NestedTitles = append(phoneVideo.NestedTitles, title)

	return nil
}
