package fcp

import (
	"fmt"

	"math/rand"

	"path/filepath"
)

// addBaffleVideoElement adds a video element with proper lane assignment for BAFFLE system
func addBaffleVideoElement(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, startTime, duration float64, elementIndex, targetLane int, verbose bool, createdAssets, createdFormats map[string]string) error {
	if verbose {
		fmt.Printf("  Adding video: %s (%.1fs @ %.1fs) lane %d\n", filepath.Base(videoPath), duration, startTime, targetLane)
	}

	// Reuse existing asset if already created, otherwise create new one
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[videoPath]; exists {
		assetID = existingAssetID
		if existingFormatID, formatExists := createdFormats[videoPath]; formatExists {
			formatID = existingFormatID
		} else {
			return fmt.Errorf("asset exists but format missing for %s", videoPath)
		}
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), ConvertSecondsToFCPDuration(duration), formatID)
		if err != nil {
			return fmt.Errorf("failed to create video asset: %v", err)
		}

		createdAssets[videoPath] = assetID
		createdFormats[videoPath] = formatID
	}

	assetClip := AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Video_%d", elementIndex),
	}

	if targetLane > 0 {
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
	}

	if rand.Float32() < 0.4 {
		assetClip.AdjustTransform = createMinimalAnimation(startTime, duration)
	}

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.AssetClips = append(spine.AssetClips, assetClip)

	return nil
}

// addBaffleTextElement adds a text element with proper lane assignment for BAFFLE system
func addBaffleTextElement(fcpxml *FCPXML, tx *ResourceTransaction, startTime, duration float64, elementIndex, targetLane int, verbose bool) error {
	textContent := generateRandomText()

	if verbose {
		fmt.Printf("  Adding text: \"%s\" (%.1fs @ %.1fs) lane %d\n", textContent, duration, startTime, targetLane)
	}

	ids := tx.ReserveIDs(1)
	effectID := ids[0]

	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		return fmt.Errorf("failed to create text effect: %v", err)
	}

	styleID := fmt.Sprintf("ts_%d", rand.Intn(999999)+100000)

	title := Title{
		Ref:      effectID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Text_%d", elementIndex),
		Text: &TitleText{
			TextStyles: []TextStyleRef{{
				Ref:  styleID,
				Text: textContent,
			}},
		},
		TextStyleDefs: []TextStyleDef{{
			ID: styleID,
			TextStyle: TextStyle{
				Font:        randomFont(),
				FontSize:    fmt.Sprintf("%.0f", 320+rand.Float64()*320),
				FontColor:   randomColor(),
				Alignment:   randomAlignment(),
				LineSpacing: fmt.Sprintf("%.1f", 1.0+rand.Float64()*0.5),
			},
		}},
	}

	if targetLane > 0 {
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
	}

	opacity := 0.7 + rand.Float64()*0.3
	title.Params = append(title.Params, Param{
		Name:  "Opacity",
		Value: fmt.Sprintf("%.2f", opacity),
	})

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Titles = append(spine.Titles, title)

	return nil
}

// addRandomImageElement adds an image with random effects and animations (legacy function)
func addRandomImageElement(fcpxml *FCPXML, tx *ResourceTransaction, imagePath string, startTime, duration float64, elementIndex, lane int, verbose bool, createdAssets, createdFormats map[string]string) error {
	if verbose {
		fmt.Printf("  Adding image: %s (%.1fs @ %.1fs)\n", filepath.Base(imagePath), duration, startTime)
	}

	// Reuse existing asset if already created, otherwise create new one
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[imagePath]; exists {

		assetID = existingAssetID
		formatID = createdFormats[imagePath]
	} else {

		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		_, err = tx.CreateAsset(assetID, imagePath, filepath.Base(imagePath), ConvertSecondsToFCPDuration(duration), formatID)
		if err != nil {
			return fmt.Errorf("failed to create image asset: %v", err)
		}

		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return fmt.Errorf("failed to create image format: %v", err)
		}

		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}

	video := Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Image_%d", elementIndex),
	}

	if lane > 0 {
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
	}

	if rand.Float32() < 0.6 {
		video.AdjustTransform = createRandomAnimation(startTime, duration)
	}

	if rand.Float32() < 0.4 {

		if video.AdjustTransform == nil {
			video.AdjustTransform = &AdjustTransform{}
		}

		video.AdjustTransform.Params = append(video.AdjustTransform.Params, Param{
			Name:  "rotation",
			Value: fmt.Sprintf("%.1f", -15.0+rand.Float64()*30.0),
		})
	}

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Videos = append(spine.Videos, video)

	return nil
}

// addRandomVideoElement adds a video with random effects
func addRandomVideoElement(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, startTime, duration float64, elementIndex, lane int, verbose bool, createdAssets, createdFormats map[string]string) error {
	if verbose {
		fmt.Printf("  Adding video: %s (%.1fs @ %.1fs)\n", filepath.Base(videoPath), duration, startTime)
	}

	// Reuse existing asset if already created, otherwise create new one
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[videoPath]; exists {

		assetID = existingAssetID
		formatID = createdFormats[videoPath]
	} else {

		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), ConvertSecondsToFCPDuration(duration+30), formatID)
		if err != nil {
			return fmt.Errorf("failed to create video asset with detection: %v", err)
		}

		createdAssets[videoPath] = assetID
		createdFormats[videoPath] = formatID
	}

	assetClip := AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Video_%d", elementIndex),
		Start:    "0s",
	}

	if lane > 0 {
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
	}

	if rand.Float32() < 0.7 {
		assetClip.AdjustTransform = createRandomAnimation(startTime, duration)
	}

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.AssetClips = append(spine.AssetClips, assetClip)

	return nil
}

// addRandomTextElement adds a text title with random styling
func addRandomTextElement(fcpxml *FCPXML, tx *ResourceTransaction, startTime, duration float64, elementIndex, lane int, verbose bool) error {
	textContent := generateRandomText()

	if verbose {
		fmt.Printf("  Adding text: \"%s\" (%.1fs @ %.1fs)\n", textContent, duration, startTime)
	}

	effectID := tx.ReserveIDs(1)[0]

	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		return fmt.Errorf("failed to create text effect: %v", err)
	}

	styleID := fmt.Sprintf("ts_baffle_%d", elementIndex)

	title := Title{
		Ref:      effectID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Text_%d", elementIndex),
		Text: &TitleText{
			TextStyles: []TextStyleRef{{
				Ref:  styleID,
				Text: textContent,
			}},
		},
		TextStyleDefs: []TextStyleDef{{
			ID: styleID,
			TextStyle: TextStyle{
				Font:        randomFont(),
				FontSize:    fmt.Sprintf("%.0f", 360+rand.Float64()*480),
				FontColor:   randomColor(),
				Alignment:   randomAlignment(),
				LineSpacing: fmt.Sprintf("%.1f", 1.0+rand.Float64()*0.5),
			},
		}},
	}

	if lane > 0 {
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
		// textLane calculation removed since lanes are not used
	}

	opacity := 0.6 + rand.Float64()*0.3
	title.Params = append(title.Params, Param{
		Name:  "Opacity",
		Value: fmt.Sprintf("%.2f", opacity),
	})

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Titles = append(spine.Titles, title)

	return nil
}

// createRandomAnimation creates random keyframe animation
// createMinimalAnimation creates subtle animations that keep content on-screen
func createMinimalAnimation(startTime, duration float64) *AdjustTransform {
	endTime := startTime + duration

	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.0f %.0f", -10+rand.Float64()*20, -5+rand.Float64()*10),
						},
						{
							Time:  ConvertSecondsToFCPDuration(endTime),
							Value: fmt.Sprintf("%.0f %.0f", -10+rand.Float64()*20, -5+rand.Float64()*10),
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.2f %.2f", 0.95+rand.Float64()*0.1, 0.95+rand.Float64()*0.1),
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(endTime),
							Value: fmt.Sprintf("%.2f %.2f", 0.95+rand.Float64()*0.1, 0.95+rand.Float64()*0.1),
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

func createRandomAnimation(startTime, duration float64) *AdjustTransform {
	// 🚨 EXTREME: Create 20-100 keyframes with chaotic timing
	numKeyframes := 20 + rand.Intn(80)
	
	positionKeyframes := make([]Keyframe, numKeyframes)
	scaleKeyframes := make([]Keyframe, numKeyframes)
	rotationKeyframes := make([]Keyframe, numKeyframes)
	
	for i := 0; i < numKeyframes; i++ {
		// 🚨 EXTREME: Random keyframe times that can be negative or way beyond duration
		keyTime := startTime + (rand.Float64()-0.5)*duration*3.0
		
		positionKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.0f %.0f", -50000+rand.Float64()*100000, -50000+rand.Float64()*100000), // 🚨 EXTREME: Massive positions
			// Position keyframes CANNOT have curve attribute per validation rules
		}
		
		scaleKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.2f %.2f", 0.01+rand.Float64()*50, 0.01+rand.Float64()*50), // 🚨 EXTREME: Tiny to huge scaling (no negatives)
			Curve: "linear", // Only "linear" is valid per DTD validation
		}
		
		rotationKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.1f", -3600+rand.Float64()*7200), // Valid range: -3600 to +3600 degrees
			Curve: "linear",
		}
	}

	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: positionKeyframes,
				},
			},
			{
				Name: "scale", 
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: scaleKeyframes,
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: rotationKeyframes,
				},
			},
			{
				Name: "anchor",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.2f %.2f", -5.0+rand.Float64()*10.0, -5.0+rand.Float64()*10.0), // Valid range: -5.0 to +5.0
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}
