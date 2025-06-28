package fcp

import (
	"fmt"

	"strconv"
	"strings"
)

// AddImessageText creates a complete imessage structure exactly like samples/imessage001.fcpxml.
// This creates the EXACT structure with matching format, durations, and timing.
func AddImessageText(fcpxml *FCPXML, text string, offsetSeconds float64, durationSeconds float64) error {

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	ids := tx.ReserveIDs(6)
	formatID := ids[0]
	phoneAssetID := ids[1]
	phoneFormatID := ids[2]
	bubbleAssetID := ids[3]
	bubbleFormatID := ids[4]
	effectID := ids[5]

	_, err := tx.CreateFormatWithFrameDuration(formatID, "100/6000s", "1080", "1920", "1-1-1 (Rec. 709)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create main format: %v", err)
	}

	_, err = tx.CreateFormat(phoneFormatID, "FFVideoFormatRateUndefined", "452", "910", "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create phone format: %v", err)
	}

	_, err = tx.CreateAsset(phoneAssetID, "/Users/aa/Movies/Untitled.fcpbundle/6-13-25/Original Media/phone_blank001 (fcp1).png", "phone_blank001", "0s", phoneFormatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create phone asset: %v", err)
	}

	_, err = tx.CreateFormat(bubbleFormatID, "FFVideoFormatRateUndefined", "392", "206", "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create bubble format: %v", err)
	}

	_, err = tx.CreateAsset(bubbleAssetID, "/Users/aa/Movies/Untitled.fcpbundle/6-13-25/Original Media/blue_speech001 (fcp1).png", "blue_speech001", "0s", bubbleFormatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create bubble asset: %v", err)
	}

	_, err = tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create text effect: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Format = formatID
		sequence.Duration = "3300/6000s"

		phoneVideo := Video{
			Ref:      phoneAssetID,
			Offset:   "0s",
			Name:     "phone_blank001",
			Start:    "21610300/6000s",
			Duration: "3300/6000s",
			NestedVideos: []Video{{
				Ref:      bubbleAssetID,
				Lane:     "1",
				Offset:   "21610300/6000s",
				Name:     "blue_speech001",
				Start:    "21610300/6000s",
				Duration: "3300/6000s",
				AdjustTransform: &AdjustTransform{
					Position: "1.26755 -21.1954",
					Scale:    "0.617236 0.617236",
				},
			}},
			NestedTitles: []Title{{
				Ref:      effectID,
				Lane:     "2",
				Offset:   "43220600/12000s",
				Name:     text + " - Text",
				Start:    "21632100/6000s",
				Duration: "3300/6000s",
				Params: []Param{
					{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
					{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
					{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: "0 -3071"},
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
					{
						Name: "Custom Speed",
						Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{Time: "-469658744/1000000000s", Value: "0"},
								{Time: "12328542033/1000000000s", Value: "1"},
							},
						},
					},
					{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
				},
				Text: &TitleText{
					TextStyles: []TextStyleRef{{
						Ref:  "ts1",
						Text: text,
					}},
				},
				TextStyleDefs: []TextStyleDef{{
					ID: "ts1",
					TextStyle: TextStyle{
						Font:        "Arial",
						FontSize:    "2040",
						FontFace:    "Regular",
						FontColor:   "0.999995 1 1 1",
						Alignment:   "center",
						LineSpacing: "-19",
					},
				}},
			}},
		}

		sequence.Spine.Videos = []Video{phoneVideo}
		sequence.Spine.AssetClips = nil
		sequence.Spine.Titles = nil
	}

	return nil
}

// AddImessageReply adds a reply message like samples/imessage002.fcpxml.
// This appends a second video segment with white speech bubble and black text.
func AddImessageReply(fcpxml *FCPXML, originalText, replyText string, offsetSeconds float64, durationSeconds float64) error {

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	ids := tx.ReserveIDs(2)
	whiteAssetID := ids[0]
	whiteFormatID := ids[1]

	_, err := tx.CreateFormat(whiteFormatID, "FFVideoFormatRateUndefined", "391", "207", "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create white bubble format: %v", err)
	}

	_, err = tx.CreateAsset(whiteAssetID, "/Users/aa/Movies/Untitled.fcpbundle/6-13-25/Original Media/white_speech001 (fcp1).png", "white_speech001", "0s", whiteFormatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create white bubble asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Get existing assets and effects
	var phoneAssetID, blueAssetID, effectID string
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Name == "phone_blank001" {
			phoneAssetID = asset.ID
		} else if asset.Name == "blue_speech001" {
			blueAssetID = asset.ID
		}
	}
	for _, effect := range fcpxml.Resources.Effects {
		if effect.Name == "Text" {
			effectID = effect.ID
			break
		}
	}

	if phoneAssetID == "" || blueAssetID == "" || effectID == "" {
		return fmt.Errorf("required assets not found in existing FCPXML")
	}

	existingIDs := getAllExistingTextStyleIDs(fcpxml)
	replyTextStyleID := getNextUniqueTextStyleID(existingIDs)
	originalTextStyleID := getNextUniqueTextStyleID(existingIDs)

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		currentDuration := sequence.Duration
		fmt.Printf("DEBUG AddImessageReply: current sequence duration = '%s'\n", currentDuration)
		if currentDuration == "" {
			currentDuration = "0s"
		}

		// Convert to /6000s format if needed
		var nextOffset string
		if strings.HasSuffix(currentDuration, "/6000s") {
			nextOffset = currentDuration
		} else {

			nextOffset = "0/6000s"
		}

		// Parse current duration to calculate new sequence duration
		var currentSixthousandths int
		if strings.HasSuffix(currentDuration, "/6000s") {
			numeratorStr := strings.TrimSuffix(currentDuration, "/6000s")
			if numerator, err := strconv.Atoi(numeratorStr); err == nil {
				currentSixthousandths = numerator
			}
		}

		newTotalSixthousandths := currentSixthousandths + 3900
		sequence.Duration = fmt.Sprintf("%d/6000s", newTotalSixthousandths)

		secondVideo := Video{
			Ref:      phoneAssetID,
			Offset:   nextOffset,
			Name:     "phone_blank001",
			Start:    "21632800/6000s",
			Duration: "3900/6000s",
			NestedVideos: []Video{
				{
					Ref:      blueAssetID,
					Lane:     "1",
					Offset:   "21632800/6000s",
					Name:     "blue_speech001",
					Start:    "21632800/6000s",
					Duration: "3900/6000s",
					AdjustTransform: &AdjustTransform{
						Position: "1.26755 -21.1954",
						Scale:    "0.617236 0.617236",
					},
				},
				{
					Ref:      whiteAssetID,
					Lane:     "2",
					Offset:   "21632800/6000s",
					Name:     "white_speech001",
					Start:    "21632800/6000s",
					Duration: "3900/6000s",
					AdjustTransform: &AdjustTransform{
						Position: "0.635834 4.00864",
						Scale:    "0.653172 0.653172",
					},
				},
			},
			NestedTitles: []Title{
				{

					Ref:      effectID,
					Lane:     "3",
					Offset:   "86531200/24000s",
					Name:     replyText + " - Text",
					Start:    "21654300/6000s",
					Duration: "3900/6000s",
					Params: []Param{
						{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
						{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
						{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: "0 -1807"},
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
						{
							Name: "Custom Speed",
							Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
							KeyframeAnimation: &KeyframeAnimation{
								Keyframes: []Keyframe{
									{Time: "-469658744/1000000000s", Value: "0"},
									{Time: "12328542033/1000000000s", Value: "1"},
								},
							},
						},
						{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
					},
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  replyTextStyleID,
							Text: replyText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: replyTextStyleID,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "2040",
							FontFace:    "Regular",
							FontColor:   "0 0 0 1",
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
				{

					Ref:      effectID,
					Lane:     "4",
					Offset:   "21632800/6000s",
					Name:     originalText + " - Text",
					Start:    "21654600/6000s",
					Duration: "3900/6000s",
					Params: []Param{
						{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
						{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
						{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: "0 -3071"},
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
						{
							Name: "Custom Speed",
							Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
							KeyframeAnimation: &KeyframeAnimation{
								Keyframes: []Keyframe{
									{Time: "-469658744/1000000000s", Value: "0"},
									{Time: "12328542033/1000000000s", Value: "1"},
								},
							},
						},
						{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
					},
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  originalTextStyleID,
							Text: originalText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: originalTextStyleID,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "2040",
							FontFace:    "Regular",
							FontColor:   "0.999995 1 1 1",
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
			},
		}

		sequence.Spine.Videos = append(sequence.Spine.Videos, secondVideo)
	}

	return nil
}

// AddImessageContinuation automatically continues an existing imessage conversation.
// Analyzes the current conversation pattern and adds the appropriate bubble type.
func AddImessageContinuation(fcpxml *FCPXML, newText string, offsetSeconds float64, durationSeconds float64) error {

	pattern := analyzeConversationPattern(fcpxml)

	fmt.Printf("DEBUG AddImessageContinuation: next bubble type = '%s', video count = %d\n", pattern.NextBubbleType, pattern.VideoCount)

	switch pattern.NextBubbleType {
	case "blue":

		return addBlueBubbleContinuation(fcpxml, newText, pattern, offsetSeconds, durationSeconds)
	case "white":

		return addWhiteBubbleContinuation(fcpxml, newText, pattern, offsetSeconds, durationSeconds)
	default:
		return fmt.Errorf("could not determine conversation pattern")
	}
}

// analyzeConversationPattern examines existing FCPXML to determine conversation state
func analyzeConversationPattern(fcpxml *FCPXML) ConversationPattern {
	pattern := ConversationPattern{
		NextBubbleType: "blue",
		VideoCount:     0,
	}

	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 ||
		len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return pattern
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	pattern.VideoCount = len(sequence.Spine.Videos)

	if pattern.VideoCount == 0 {
		return pattern
	}

	lastVideo := sequence.Spine.Videos[pattern.VideoCount-1]

	hasWhiteBubble := false
	lastBubbleText := ""

	for _, nestedVideo := range lastVideo.NestedVideos {
		if nestedVideo.Name == "white_speech001" {
			hasWhiteBubble = true
			break
		}
	}

	for _, title := range lastVideo.NestedTitles {
		if len(title.Text.TextStyles) > 0 {
			lastBubbleText = title.Text.TextStyles[0].Text
		}
	}

	pattern.LastText = lastBubbleText

	if hasWhiteBubble {

		pattern.NextBubbleType = "blue"
	} else {

		pattern.NextBubbleType = "white"
	}

	if pattern.VideoCount == 1 {

		pattern.NextOffset = "3300/6000s"
		pattern.NextDuration = "3900/6000s"
	} else {

		totalFrameUnits := 0
		for _, video := range sequence.Spine.Videos {
			durationStr := video.Duration
			if durationStr != "" {
				durationUnits := parseFCPDuration(durationStr)
				totalFrameUnits += durationUnits
			}
		}
		pattern.NextOffset = fmt.Sprintf("%d/24000s", totalFrameUnits)
		pattern.NextDuration = "3900/6000s"
	}

	return pattern
}

// addBlueBubbleContinuation adds a blue bubble message (sender)
func addBlueBubbleContinuation(fcpxml *FCPXML, newText string, pattern ConversationPattern, offsetSeconds float64, durationSeconds float64) error {
	// Get existing assets
	var phoneAssetID, blueAssetID, effectID string
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Name == "phone_blank001" {
			phoneAssetID = asset.ID
		} else if asset.Name == "blue_speech001" {
			blueAssetID = asset.ID
		}
	}
	for _, effect := range fcpxml.Resources.Effects {
		if effect.Name == "Text" {
			effectID = effect.ID
			break
		}
	}

	if phoneAssetID == "" || blueAssetID == "" || effectID == "" {
		return fmt.Errorf("required assets not found in existing FCPXML")
	}

	existingIDs := getAllExistingTextStyleIDs(fcpxml)
	uniqueTextStyleID1 := getNextUniqueTextStyleID(existingIDs)

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		totalSixthousandths := 0
		for _, video := range sequence.Spine.Videos {
			durationStr := video.Duration
			if durationStr != "" && strings.HasSuffix(durationStr, "/6000s") {

				numeratorStr := strings.TrimSuffix(durationStr, "/6000s")
				if numerator, err := strconv.Atoi(numeratorStr); err == nil {
					totalSixthousandths += numerator
				}
			}
		}
		nextOffset := fmt.Sprintf("%d/6000s", totalSixthousandths)

		newTotalSixthousandths := totalSixthousandths + 3900
		sequence.Duration = fmt.Sprintf("%d/6000s", newTotalSixthousandths)

		nextVideo := Video{
			Ref:      phoneAssetID,
			Offset:   nextOffset,
			Name:     "phone_blank001",
			Start:    "21632800/6000s",
			Duration: "3900/6000s",
			NestedVideos: []Video{{
				Ref:      blueAssetID,
				Lane:     "1",
				Offset:   "21632800/6000s",
				Name:     "blue_speech001",
				Start:    "21632800/6000s",
				Duration: "3900/6000s",
				AdjustTransform: &AdjustTransform{
					Position: "1.26755 -21.1954",
					Scale:    "0.617236 0.617236",
				},
			}},
			NestedTitles: []Title{
				{

					Ref:      effectID,
					Lane:     "1",
					Offset:   "21610300/6000s",
					Name:     newText + " - Text",
					Start:    "21632100/6000s",
					Duration: "3900/6000s",
					Params:   createStandardTextParams("0 -3071"),
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  uniqueTextStyleID1,
							Text: newText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: uniqueTextStyleID1,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "2040",
							FontFace:    "Regular",
							FontColor:   "0.999995 1 1 1",
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
			},
		}

		sequence.Spine.Videos = append(sequence.Spine.Videos, nextVideo)
	}

	return nil
}

// addWhiteBubbleContinuation adds a white bubble message (reply) without duplicating previous messages
func addWhiteBubbleContinuation(fcpxml *FCPXML, newText string, pattern ConversationPattern, offsetSeconds float64, durationSeconds float64) error {
	// Get existing assets
	var phoneAssetID, blueAssetID, whiteAssetID, effectID string
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Name == "phone_blank001" {
			phoneAssetID = asset.ID
		} else if asset.Name == "blue_speech001" {
			blueAssetID = asset.ID
		} else if asset.Name == "white_speech001" {
			whiteAssetID = asset.ID
		}
	}
	for _, effect := range fcpxml.Resources.Effects {
		if effect.Name == "Text" {
			effectID = effect.ID
			break
		}
	}

	if phoneAssetID == "" || blueAssetID == "" || whiteAssetID == "" || effectID == "" {
		return fmt.Errorf("required assets not found in existing FCPXML")
	}

	existingIDs := getAllExistingTextStyleIDs(fcpxml)
	uniqueTextStyleID := getNextUniqueTextStyleID(existingIDs)

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		totalSixthousandths := 0
		for _, video := range sequence.Spine.Videos {
			durationStr := video.Duration
			if durationStr != "" && strings.HasSuffix(durationStr, "/6000s") {

				numeratorStr := strings.TrimSuffix(durationStr, "/6000s")
				if numerator, err := strconv.Atoi(numeratorStr); err == nil {
					totalSixthousandths += numerator
				}
			}
		}
		nextOffset := fmt.Sprintf("%d/6000s", totalSixthousandths)

		newTotalSixthousandths := totalSixthousandths + 3900
		sequence.Duration = fmt.Sprintf("%d/6000s", newTotalSixthousandths)

		nextVideo := Video{
			Ref:      phoneAssetID,
			Offset:   nextOffset,
			Name:     "phone_blank001",
			Start:    "21632800/6000s",
			Duration: "3900/6000s",
			NestedVideos: []Video{
				{
					Ref:      blueAssetID,
					Lane:     "1",
					Offset:   "21632800/6000s",
					Name:     "blue_speech001",
					Start:    "21632800/6000s",
					Duration: "3900/6000s",
					AdjustTransform: &AdjustTransform{
						Position: "1.26755 -21.1954",
						Scale:    "0.617236 0.617236",
					},
				},
				{
					Ref:      whiteAssetID,
					Lane:     "2",
					Offset:   "21632800/6000s",
					Name:     "white_speech001",
					Start:    "21632800/6000s",
					Duration: "3900/6000s",
					AdjustTransform: &AdjustTransform{
						Position: "0.635834 4.00864",
						Scale:    "0.653172 0.653172",
					},
				},
			},
			NestedTitles: []Title{
				{
					Ref:      effectID,
					Lane:     "3",
					Offset:   "86531200/24000s",
					Name:     newText + " - Text",
					Start:    "21654300/6000s",
					Duration: "3900/6000s",
					Params:   createStandardTextParams("0 -1807"),
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  uniqueTextStyleID,
							Text: newText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: uniqueTextStyleID,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "2040",
							FontFace:    "Regular",
							FontColor:   "0 0 0 1",
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
			},
		}

		sequence.Spine.Videos = append(sequence.Spine.Videos, nextVideo)
	}

	return nil
}

// createStandardTextParams creates the standard text parameters with given position
func createStandardTextParams(position string) []Param {
	return []Param{
		{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
		{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
		{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: position},
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
		{
			Name: "Custom Speed",
			Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
			KeyframeAnimation: &KeyframeAnimation{
				Keyframes: []Keyframe{
					{Time: "-469658744/1000000000s", Value: "0"},
					{Time: "12328542033/1000000000s", Value: "1"},
				},
			},
		},
		{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
	}
}
