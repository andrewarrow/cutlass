package fcp

import (
	"fmt"

	"os"

	"strconv"
	"strings"
)

// calculateTimelineDuration calculates the total duration of content in a sequence
// by examining all clips in the spine and finding the maximum offset + duration
func calculateTimelineDuration(sequence *Sequence) string {
	maxEndTime := 0

	for _, clip := range sequence.Spine.AssetClips {
		clipEndTime := parseOffsetAndDuration(clip.Offset, clip.Duration)
		if clipEndTime > maxEndTime {
			maxEndTime = clipEndTime
		}
	}

	for _, video := range sequence.Spine.Videos {
		videoEndTime := parseOffsetAndDuration(video.Offset, video.Duration)
		if videoEndTime > maxEndTime {
			maxEndTime = videoEndTime
		}
	}

	for _, title := range sequence.Spine.Titles {
		titleEndTime := parseOffsetAndDuration(title.Offset, title.Duration)
		if titleEndTime > maxEndTime {
			maxEndTime = titleEndTime
		}
	}

	for _, gap := range sequence.Spine.Gaps {
		gapEndTime := parseOffsetAndDuration(gap.Offset, gap.Duration)
		if gapEndTime > maxEndTime {
			maxEndTime = gapEndTime
		}
	}

	if maxEndTime == 0 {
		return "0s"
	}
	return fmt.Sprintf("%d/24000s", maxEndTime)
}

// parseOffsetAndDuration parses FCP time format and returns end time in 1001/24000s units
func parseOffsetAndDuration(offset, duration string) int {
	offsetFrames := parseFCPDuration(offset)
	durationFrames := parseFCPDuration(duration)
	return offsetFrames + durationFrames
}

// parseFCPDuration parses FCP duration format and returns frame-aligned values in 1001/24000s units
func parseFCPDuration(duration string) int {
	if duration == "0s" {
		return 0
	}

	if strings.HasSuffix(duration, "s") && strings.Contains(duration, "/") {

		durationNoS := strings.TrimSuffix(duration, "s")

		parts := strings.Split(durationNoS, "/")
		if len(parts) == 2 {
			numerator, err1 := strconv.Atoi(parts[0])
			denominator, err2 := strconv.Atoi(parts[1])

			if err1 == nil && err2 == nil && denominator != 0 {

				framesFloat := float64(numerator*24000) / float64(denominator*1001)
				frames := int(framesFloat + 0.5)

				return frames * 1001
			}
		}
	}

	return 0
}

// addDurations adds two FCP duration strings and returns the result
func addDurations(duration1, duration2 string) string {
	frames1 := parseFCPDuration(duration1)
	frames2 := parseFCPDuration(duration2)
	totalFrames := frames1 + frames2
	return fmt.Sprintf("%d/24000s", totalFrames)
}

// createKenBurnsAnimation creates Ken Burns effect animation (slow zoom + pan)
// Ken Burns effect combines gradual zoom-in with subtle panning motion
// 🚨 CRITICAL: Keyframe attributes follow CLAUDE.md rules:
// - Position keyframes: NO attributes (no interp/curve)
// - Scale/Rotation/Anchor keyframes: Only curve attribute (no interp)
func createKenBurnsAnimation(offsetDuration string, totalDurationSeconds float64) *AdjustTransform {
	return createKenBurnsAnimationWithFormat(offsetDuration, totalDurationSeconds, "horizontal")
}

// createKenBurnsAnimationWithFormat creates Ken Burns effect animation with format-aware scaling
func createKenBurnsAnimationWithFormat(offsetDuration string, totalDurationSeconds float64, format string) *AdjustTransform {

	videoStartFrames := 86399313

	// Ken Burns effect duration should be longer than slide (3 seconds for subtle effect)
	kenBurnsDuration := ConvertSecondsToFCPDuration(3.0)
	kenBurnsFrames := parseFCPDuration(kenBurnsDuration)

	startTime := fmt.Sprintf("%d/24000s", videoStartFrames)
	endTime := fmt.Sprintf("%d/24000s", videoStartFrames+kenBurnsFrames)

	// Adjust scale values based on format
	var startScale, endScale string
	switch format {
	case "vertical":
		// Higher zoom for vertical format to fill frame with no empty space
		startScale = "2.0 2.0"  // Start zoomed in more for vertical
		endScale = "2.4 2.4"    // End even more zoomed for Ken Burns effect
	case "horizontal":
		fallthrough
	default:
		// Original scaling for horizontal
		startScale = "1.2 1.2"
		endScale = "1.5 1.5"
	}

	return &AdjustTransform{
		Params: []Param{
			{
				Name: "anchor",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  startTime,
							Value: "0 0",
							Curve: "linear",
						},
						{
							Time:  endTime,
							Value: "0 0",
							Curve: "linear",
						},
					},
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  startTime,
							Value: "0 0",
						},
						{
							Time:  endTime,
							Value: "-20 -15",
						},
					},
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  startTime,
							Value: "0",
							Curve: "linear",
						},
						{
							Time:  endTime,
							Value: "0",
							Curve: "linear",
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  startTime,
							Value: startScale,
							Curve: "linear",
						},
						{
							Time:  endTime,
							Value: endScale,
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// AddTextFromFile reads a text file and adds staggered text elements to the FCPXML structure.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Effects, sequence.Spine.Titles
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function
// - Unique text-style-def IDs → generateUID() function for deterministic UIDs
// - Each text element appears 1 second later with 300px Y offset progression
//
// ❌ NEVER: fmt.Sprintf("<title ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddTextFromFile(fcpxml *FCPXML, textFilePath string, offsetSeconds float64, durationSeconds float64) error {

	data, err := os.ReadFile(textFilePath)
	if err != nil {
		return fmt.Errorf("failed to read text file: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	var textLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			textLines = append(textLines, line)
		}
	}

	if len(textLines) == 0 {
		return fmt.Errorf("no text lines found in file: %s", textFilePath)
	}

	registry := NewResourceRegistry(fcpxml)

	tx := NewTransaction(registry)

	textEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if strings.Contains(effect.UID, "Text.moti") {
			textEffectID = effect.ID
			break
		}
	}

	if textEffectID == "" {

		ids := tx.ReserveIDs(1)
		textEffectID = ids[0]

		_, err = tx.CreateEffect(textEffectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create text effect: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit text effect: %v", err)
	}

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		// Find the clip element that covers the text offset time
		var targetAssetClip *AssetClip = nil
		var targetVideo *Video = nil
		offsetFrames := parseFCPDuration(ConvertSecondsToFCPDuration(offsetSeconds))

		for i := range sequence.Spine.AssetClips {
			clip := &sequence.Spine.AssetClips[i]
			clipOffsetFrames := parseFCPDuration(clip.Offset)
			clipDurationFrames := parseFCPDuration(clip.Duration)
			clipEndFrames := clipOffsetFrames + clipDurationFrames

			if offsetFrames >= clipOffsetFrames && offsetFrames < clipEndFrames {
				targetAssetClip = clip
				break
			}
		}

		if targetAssetClip == nil {
			for i := range sequence.Spine.Videos {
				video := &sequence.Spine.Videos[i]
				videoOffsetFrames := parseFCPDuration(video.Offset)
				videoDurationFrames := parseFCPDuration(video.Duration)
				videoEndFrames := videoOffsetFrames + videoDurationFrames

				if offsetFrames >= videoOffsetFrames && offsetFrames < videoEndFrames {
					targetVideo = video
					break
				}
			}
		}

		if targetAssetClip == nil && targetVideo == nil {
			if len(sequence.Spine.AssetClips) > 0 {

				targetAssetClip = &sequence.Spine.AssetClips[len(sequence.Spine.AssetClips)-1]
			} else if len(sequence.Spine.Videos) > 0 {

				targetVideo = &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]
			}
		}

		if targetAssetClip == nil && targetVideo == nil {
			return fmt.Errorf("no video or asset-clip element found in spine to add text overlays to")
		}

		textDuration := ConvertSecondsToFCPDuration(durationSeconds)

		for i, textLine := range textLines {

			textTx := NewTransaction(registry)

			textStyleID := GenerateTextStyleID(textLine, fmt.Sprintf("line_%d_offset_%.1f", i, offsetSeconds))

			// Calculate staggered timing: first element at offsetSeconds in sequence timeline, each subsequent +6 seconds
			// Text timing should use the clip's start time as base for proper FCP timing
			var clipStartFrames int
			if targetAssetClip != nil {
				clipStartFrames = parseFCPDuration(targetAssetClip.Start)
			} else if targetVideo != nil {
				clipStartFrames = parseFCPDuration(targetVideo.Start)
			}

			staggerSeconds := durationSeconds * 0.5
			staggerDuration := ConvertSecondsToFCPDuration(staggerSeconds)
			staggerFramesPer := parseFCPDuration(staggerDuration)
			staggerFrames := i * staggerFramesPer
			elementOffsetFrames := clipStartFrames + staggerFrames
			elementOffset := fmt.Sprintf("%d/24000s", elementOffsetFrames)

			yOffset := i * -300
			positionValue := fmt.Sprintf("0 %d", yOffset)

			laneNumber := len(textLines) - i

			title := Title{
				Ref:      textEffectID,
				Lane:     fmt.Sprintf("%d", laneNumber),
				Offset:   elementOffset,
				Name:     fmt.Sprintf("%s - Text", textLine),
				Start:    "86486400/24000s",
				Duration: textDuration,
				Params: []Param{
					{
						Name:  "Layout Method",
						Key:   "9999/10003/13260/3296672360/2/314",
						Value: "1 (Paragraph)",
					},
					{
						Name:  "Left Margin",
						Key:   "9999/10003/13260/3296672360/2/323",
						Value: "-1730",
					},
					{
						Name:  "Right Margin",
						Key:   "9999/10003/13260/3296672360/2/324",
						Value: "1730",
					},
					{
						Name:  "Top Margin",
						Key:   "9999/10003/13260/3296672360/2/325",
						Value: "960",
					},
					{
						Name:  "Bottom Margin",
						Key:   "9999/10003/13260/3296672360/2/326",
						Value: "-960",
					},
					{
						Name:  "Alignment",
						Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
						Value: "0 (Left)",
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
							Text: textLine,
						},
					},
				},
				TextStyleDefs: []TextStyleDef{
					{
						ID: textStyleID,
						TextStyle: TextStyle{
							Font:        "Helvetica Neue",
							FontSize:    "1340",
							FontColor:   "1 1 1 1",
							Bold:        "1",
							LineSpacing: "-19",
						},
					},
				},
			}

			if i > 0 {
				positionParam := Param{
					Name:  "Position",
					Key:   "9999/10003/13260/3296672360/1/100/101",
					Value: positionValue,
				}

				title.Params = append([]Param{positionParam}, title.Params...)
			}

			err = textTx.Commit()
			if err != nil {
				return fmt.Errorf("failed to commit text transaction for element %d: %v", i, err)
			}

			if targetAssetClip != nil {
				targetAssetClip.Titles = append(targetAssetClip.Titles, title)
			} else if targetVideo != nil {
				targetVideo.NestedTitles = append(targetVideo.NestedTitles, title)
			}
		}

	}

	return nil
}

// AddSingleText adds a single text element like in samples/imessage001.fcpxml to an FCPXML file.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Effects, sequence.Spine.Titles
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function
// - Uses verified Text effect UID from samples/imessage001.fcpxml → ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti"
//
// ❌ NEVER: fmt.Sprintf("<title ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddSingleText(fcpxml *FCPXML, text string, offsetSeconds float64, durationSeconds float64) error {

	registry := NewResourceRegistry(fcpxml)

	tx := NewTransaction(registry)

	ids := tx.ReserveIDs(1)
	effectID := ids[0]

	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create text effect: %v", err)
	}

	textStyleID := GenerateTextStyleID(text, fmt.Sprintf("single_text_offset_%.1f", offsetSeconds))

	offsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)
	titleDuration := ConvertSecondsToFCPDuration(durationSeconds)

	title := Title{
		Ref:      effectID,
		Lane:     "2",
		Offset:   offsetDuration,
		Name:     text + " - Text",
		Start:    "21632100/6000s",
		Duration: titleDuration,
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
				Value: "0 -3071",
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
					Text: text,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: textStyleID,
				TextStyle: TextStyle{
					Font:        "Arial",
					FontSize:    "2040",
					FontFace:    "Regular",
					FontColor:   "0.999995 1 1 1",
					Alignment:   "center",
					LineSpacing: "-19",
				},
			},
		},
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		if len(sequence.Spine.Videos) > 0 {

			sequence.Spine.Videos[0].NestedTitles = append(sequence.Spine.Videos[0].NestedTitles, title)
		} else if len(sequence.Spine.AssetClips) > 0 {

			sequence.Spine.AssetClips[0].Titles = append(sequence.Spine.AssetClips[0].Titles, title)
		} else {

			sequence.Spine.Titles = append(sequence.Spine.Titles, title)
		}
	}

	return nil
}
