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
// Conversation Pattern (from samples/imessage.fcpxml):
// 1. Phone background as base image (Video element, duration spans entire conversation)
// 2. Alternating speech bubbles: blue (user question), white (response)
// 3. Text overlays positioned within speech bubbles with proper FCP timing
// 4. Each message appears 2 seconds apart for readability
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

	// Calculate total conversation duration: 2 seconds per message + 2 seconds padding
	totalDurationSeconds := float64(len(messages)*2 + 2)

	// Add phone background as base layer (spans entire conversation)
	err = AddImage(fcpxml, phoneBackgroundPath, totalDurationSeconds)
	if err != nil {
		return fmt.Errorf("failed to add phone background: %v", err)
	}

	// Process each message with alternating speech bubbles
	for i, message := range messages {
		// Determine which speech bubble to use (odd = blue, even = white)
		var speechBubblePath string
		var textColor string
		if i%2 == 0 {
			// First message (index 0) and every other = blue (user)
			speechBubblePath = blueSpeechPath
			textColor = "0 0 0 1" // Black text for blue bubble
		} else {
			// Second message (index 1) and every other = white (response)
			speechBubblePath = whiteSpeechPath
			textColor = "0.999995 1 1 1" // White text for white bubble
		}

		// Calculate message timing: each message appears 2 seconds after the previous
		messageOffsetSeconds := float64(i * 2)

		// Add speech bubble image for this message (2 second duration)
		err = addConversationBubble(fcpxml, speechBubblePath, messageOffsetSeconds, 2.0, i%2 == 0)
		if err != nil {
			return fmt.Errorf("failed to add speech bubble %d: %v", i, err)
		}

		// Add text overlay for this message
		err = addConversationText(fcpxml, message, messageOffsetSeconds, 2.0, textColor, i%2 == 0)
		if err != nil {
			return fmt.Errorf("failed to add text for message %d: %v", i, err)
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

// addConversationBubble adds a speech bubble image as a Video element to the conversation
// Based on samples/imessage.fcpxml positioning and scaling patterns
func addConversationBubble(fcpxml *FCPXML, speechBubblePath string, offsetSeconds, durationSeconds float64, isBlue bool) error {
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

	// Add speech bubble as Video element to spine (images use Video, not AssetClip)
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		// Calculate offset in FCP format
		offsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)
		bubbleDuration := ConvertSecondsToFCPDuration(durationSeconds)

		// Create Video element for speech bubble
		// Positioning based on samples/imessage.fcpxml:
		// - Blue bubble: position="1.26755 -21.1954" scale="0.617236 0.617236"
		// - White bubble: position="0.635834 4.00864" scale="0.653172 0.653172"
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

		// Find the phone background video to nest the speech bubble inside
		if len(sequence.Spine.Videos) > 0 {
			phoneVideo := &sequence.Spine.Videos[0]

			// Create nested video for speech bubble (following samples/imessage.fcpxml pattern)
			bubbleVideo := Video{
				Ref:      bubbleAsset.ID,
				Lane:     lane,
				Offset:   offsetDuration,
				Name:     bubbleAsset.Name,
				Start:    offsetDuration, // Start time matches offset for nested elements
				Duration: bubbleDuration,
				AdjustTransform: &AdjustTransform{
					Position: position,
					Scale:    scale,
				},
			}

			// Add as nested video inside phone background
			phoneVideo.NestedVideos = append(phoneVideo.NestedVideos, bubbleVideo)
		}
	}

	return nil
}

// addConversationText adds text overlay positioned within a speech bubble
// Based on samples/imessage.fcpxml text positioning patterns
func addConversationText(fcpxml *FCPXML, message string, offsetSeconds, durationSeconds float64, textColor string, isBlue bool) error {
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

	// Add text as Title element nested within phone background video
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		if len(sequence.Spine.Videos) > 0 {
			phoneVideo := &sequence.Spine.Videos[0]

			// Calculate text positioning based on speech bubble type
			// From samples/imessage.fcpxml:
			// - Blue text: Position "0 -3071" (higher up)
			// - White text: Position "0 -1807" (lower down)
			var textPosition string
			var lane string
			if isBlue {
				textPosition = "0 -3071"
				lane = "4"
			} else {
				textPosition = "0 -1807"
				lane = "3"
			}

			// Calculate timing
			offsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)
			textDuration := ConvertSecondsToFCPDuration(durationSeconds)

			// Generate unique text-style-def ID
			textStyleID := GenerateTextStyleID(message, fmt.Sprintf("conversation_%.1f_%t", offsetSeconds, isBlue))

			// Create Title element matching samples/imessage.fcpxml pattern
			title := Title{
				Ref:      textEffectID,
				Lane:     lane,
				Offset:   offsetDuration,
				Name:     fmt.Sprintf("%s - Text", message),
				Start:    "86617531/24000s", // Standard FCP start time for conversation text
				Duration: textDuration,
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

			// Add title as nested element within phone background video
			phoneVideo.NestedTitles = append(phoneVideo.NestedTitles, title)
		}
	}

	return nil
}