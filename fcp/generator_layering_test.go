package fcp

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestAdvancedLayering tests complex layering scenarios with multiple videos
// and different blend modes to validate the FCP package can handle sophisticated
// video compositing setups that would be common in professional editing.
func TestAdvancedLayering(t *testing.T) {
	// Create empty FCPXML
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	// Initialize registry for proper ID management
	registry := NewResourceRegistry(fcpxml)

	// Create transaction for multiple video assets
	tx := NewTransaction(registry)

	// Reserve IDs for 3 videos + 3 formats + 1 blend effect
	ids := tx.ReserveIDs(7)
	mainVideoID := ids[0]
	overlayVideo1ID := ids[1] 
	overlayVideo2ID := ids[2]
	mainFormatID := ids[3]
	overlay1FormatID := ids[4]
	overlay2FormatID := ids[5]
	blendEffectID := ids[6]

	// Create different formats for each video to enable different properties
	_, err = tx.CreateFormatWithFrameDuration(mainFormatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create main format: %v", err)
	}

	_, err = tx.CreateFormatWithFrameDuration(overlay1FormatID, "1001/25000s", "1280", "720", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create overlay1 format: %v", err)
	}

	_, err = tx.CreateFormatWithFrameDuration(overlay2FormatID, "1001/30000s", "3840", "2160", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create overlay2 format: %v", err)
	}

	// Create blend effect
	_, err = tx.CreateEffect(blendEffectID, "Blend", "FFScreenBlend")
	if err != nil {
		t.Fatalf("Failed to create blend effect: %v", err)
	}

	// Create video assets with different durations
	mainDuration := ConvertSecondsToFCPDuration(30.0)
	overlay1Duration := ConvertSecondsToFCPDuration(15.0)
	overlay2Duration := ConvertSecondsToFCPDuration(20.0)

	mainAsset, err := tx.CreateAsset(mainVideoID, "/Users/aa/cs/cutlass/assets/long.mov", "MainVideo", mainDuration, mainFormatID)
	if err != nil {
		t.Fatalf("Failed to create main asset: %v", err)
	}

	overlay1Asset, err := tx.CreateAsset(overlayVideo1ID, "/Users/aa/cs/cutlass/assets/speech1.mov", "OverlayVideo1", overlay1Duration, overlay1FormatID)
	if err != nil {
		t.Fatalf("Failed to create overlay1 asset: %v", err)
	}

	overlay2Asset, err := tx.CreateAsset(overlayVideo2ID, "/Users/aa/cs/cutlass/assets/speech2.mov", "OverlayVideo2", overlay2Duration, overlay2FormatID)
	if err != nil {
		t.Fatalf("Failed to create overlay2 asset: %v", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build complex layering structure in the spine
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Main video layer (base layer)
	mainClip := AssetClip{
		Ref:      mainAsset.ID,
		Offset:   "0s",
		Name:     mainAsset.Name,
		Duration: mainDuration,
		Format:   mainAsset.Format,
		TCFormat: "NDF",
		AudioRole: "dialogue",
		ConformRate: &ConformRate{
			ScaleEnabled: "0",
		},
		// Nested clips for layering
		NestedAssetClips: []AssetClip{
			// First overlay - blend mode with position offset
			{
				Ref:      overlay1Asset.ID,
				Lane:     "1", // Upper layer
				Offset:   ConvertSecondsToFCPDuration(5.0), // Start 5 seconds in
				Name:     overlay1Asset.Name,
				Duration: overlay1Duration,
				Format:   overlay1Asset.Format,
				TCFormat: "NDF",
				ConformRate: &ConformRate{
					ScaleEnabled: "0",
				},
				AdjustTransform: &AdjustTransform{
					Position: "200 100", // Offset position
					Scale:    "0.5 0.5", // Scale down
				},
				FilterVideos: []FilterVideo{
					{
						Ref:  blendEffectID,
						Name: "Blend",
						Params: []Param{
							{
								Name:  "Blend Mode",
								Key:   "1",
								Value: "31 (Screen)",
							},
							{
								Name:  "Opacity",
								Key:   "2",
								Value: "75",
							},
						},
					},
				},
			},
			// Second overlay - different position and timing
			{
				Ref:      overlay2Asset.ID,
				Lane:     "2", // Even higher layer
				Offset:   ConvertSecondsToFCPDuration(10.0), // Start 10 seconds in
				Name:     overlay2Asset.Name,
				Duration: overlay2Duration,
				Format:   overlay2Asset.Format,
				TCFormat: "NDF",
				ConformRate: &ConformRate{
					ScaleEnabled: "0",
				},
				AdjustTransform: &AdjustTransform{
					Position: "-300 -200", // Different position
					Scale:    "0.3 0.3",   // Smaller scale
				},
				AdjustCrop: &AdjustCrop{
					Mode: "trim",
					TrimRect: &TrimRect{
						Left:   "10",
						Right:  "10",
						Top:    "5",
						Bottom: "5",
					},
				},
			},
		},
	}

	// Add main clip to spine
	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, mainClip)

	// Update sequence duration to match longest content
	sequence.Duration = mainDuration

	// Validate FCPXML structure
	violations := ValidateClaudeCompliance(fcpxml)
	if len(violations) > 0 {
		t.Errorf("CLAUDE.md compliance violations found:")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}

	// Test XML marshaling
	output, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	xmlContent := string(output)

	// Verify layering structure in generated XML
	testCases := []struct {
		name     string
		expected string
	}{
		{"Main clip exists", fmt.Sprintf(`ref="%s"`, mainAsset.ID)},
		{"Nested overlay 1 with lane", `lane="1"`},
		{"Nested overlay 2 with lane", `lane="2"`},
		{"Blend effect applied", `name="Blend"`},
		{"Screen blend mode", `31 (Screen)`},
		{"Opacity parameter", `<param name="Opacity"`},
		{"Transform positioning", `position="200 100"`},
		{"Scale transformation", `scale="0.5 0.5"`},
		{"Crop adjustments", `<adjust-crop mode="trim"`},
		{"Conform rate elements", `<conform-rate scaleEnabled="0"`},
		{"Multiple format IDs", fmt.Sprintf(`format="%s"`, overlay1FormatID)},
		{"Timeline offsets", `offset="120120/24000s"`}, // 5 seconds in frame-aligned format
	}

	for _, tc := range testCases {
		if !strings.Contains(xmlContent, tc.expected) {
			t.Errorf("Test '%s' failed: expected '%s' not found in XML", tc.name, tc.expected)
		}
	}

	// Write test file for manual FCP validation
	testFileName := "test_advanced_layering.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFileName); os.IsNotExist(err) {
		t.Errorf("Test file was not created: %s", testFileName)
	}

	fmt.Printf("✅ Advanced layering test file created: %s\n", testFileName)
	fmt.Printf("   - Main video (30s) with 2 overlay videos\n")
	fmt.Printf("   - Blend modes, transforms, and cropping applied\n")
	fmt.Printf("   - Different formats and conform-rate handling\n")
	fmt.Printf("   - Ready for FCP import validation\n")
}

// TestComplexCompositing tests even more sophisticated compositing with keyframes
func TestComplexCompositing(t *testing.T) {
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	// Reserve IDs for assets, formats, and effects
	ids := tx.ReserveIDs(8)
	backgroundID := ids[0]
	foregroundID := ids[1]
	bgFormatID := ids[2]
	fgFormatID := ids[3]
	blendEffectID := ids[4]
	colorCorrectionID := ids[5]
	maskEffectID := ids[6]
	vignetteEffectID := ids[7]

	// Create formats
	_, err = tx.CreateFormatWithFrameDuration(bgFormatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create background format: %v", err)
	}

	_, err = tx.CreateFormatWithFrameDuration(fgFormatID, "1001/25000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create foreground format: %v", err)
	}

	// Create effects
	_, err = tx.CreateEffect(blendEffectID, "Blend", "FFMultiplyBlend")
	if err != nil {
		t.Fatalf("Failed to create blend effect: %v", err)
	}

	_, err = tx.CreateEffect(colorCorrectionID, "Color Board", "FFColorBoard")
	if err != nil {
		t.Fatalf("Failed to create color correction effect: %v", err)
	}

	_, err = tx.CreateEffect(maskEffectID, "Shape Mask", "FFSuperEllipseMask")
	if err != nil {
		t.Fatalf("Failed to create mask effect: %v", err)
	}

	_, err = tx.CreateEffect(vignetteEffectID, "Vignette", "FFVignette")
	if err != nil {
		t.Fatalf("Failed to create vignette effect: %v", err)
	}

	// Create assets
	bgDuration := ConvertSecondsToFCPDuration(25.0)
	fgDuration := ConvertSecondsToFCPDuration(15.0)

	bgAsset, err := tx.CreateAsset(backgroundID, "/Users/aa/cs/cutlass/assets/long.mov", "Background", bgDuration, bgFormatID)
	if err != nil {
		t.Fatalf("Failed to create background asset: %v", err)
	}

	fgAsset, err := tx.CreateAsset(foregroundID, "/Users/aa/cs/cutlass/assets/speech1.mov", "Foreground", fgDuration, fgFormatID)
	if err != nil {
		t.Fatalf("Failed to create foreground asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build complex compositing setup
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Complex keyframe animation for position
	positionKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "-500 0", // Start off-screen left
			},
			{
				Time:  ConvertSecondsToFCPDuration(5.0),
				Value: "0 0", // Move to center
			},
			{
				Time:  ConvertSecondsToFCPDuration(10.0),
				Value: "500 -200", // Move to upper right
			},
			{
				Time:  ConvertSecondsToFCPDuration(15.0),
				Value: "0 400", // Move to bottom center
			},
		},
	}

	// Scale animation keyframes
	scaleKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "0.1 0.1", // Start very small
			},
			{
				Time:  ConvertSecondsToFCPDuration(3.0),
				Value: "1.2 1.2", // Scale up
			},
			{
				Time:  ConvertSecondsToFCPDuration(12.0),
				Value: "0.8 0.8", // Scale down slightly
			},
			{
				Time:  ConvertSecondsToFCPDuration(15.0),
				Value: "1.5 1.5", // End larger
			},
		},
	}

	// Opacity animation for blend effect
	opacityKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "0", // Start invisible
			},
			{
				Time:  ConvertSecondsToFCPDuration(2.0),
				Value: "100", // Fade in
			},
			{
				Time:  ConvertSecondsToFCPDuration(13.0),
				Value: "100", // Stay visible
			},
			{
				Time:  ConvertSecondsToFCPDuration(15.0),
				Value: "0", // Fade out
			},
		},
	}

	mainClip := AssetClip{
		Ref:      bgAsset.ID,
		Offset:   "0s",
		Name:     bgAsset.Name,
		Duration: bgDuration,
		Format:   bgAsset.Format,
		TCFormat: "NDF",
		AudioRole: "dialogue",
		FilterVideos: []FilterVideo{
			// Color correction on background
			{
				Ref:  colorCorrectionID,
				Name: "Color Board",
				Params: []Param{
					{
						Name:  "Color",
						Key:   "1",
						Value: "1 0.8 0.6 1", // Warm tint
					},
					{
						Name:  "Saturation",
						Key:   "2",
						Value: "1.2", // Increase saturation
					},
					{
						Name:  "Exposure",
						Key:   "3",
						Value: "-0.2", // Slight underexposure
					},
				},
			},
			// Vignette effect
			{
				Ref:  vignetteEffectID,
				Name: "Vignette",
				Params: []Param{
					{
						Name:  "Amount",
						Key:   "1",
						Value: "0.3",
					},
					{
						Name:  "Radius",
						Key:   "2",
						Value: "0.8",
					},
				},
			},
		},
		NestedAssetClips: []AssetClip{
			{
				Ref:      fgAsset.ID,
				Lane:     "1",
				Offset:   ConvertSecondsToFCPDuration(2.0), // Start 2 seconds in
				Name:     fgAsset.Name,
				Duration: fgDuration,
				Format:   fgAsset.Format,
				TCFormat: "NDF",
				ConformRate: &ConformRate{
					ScaleEnabled: "0",
				},
				AdjustTransform: &AdjustTransform{
					Params: []Param{
						{
							Name:              "position",
							KeyframeAnimation: positionKeyframes,
						},
						{
							Name:              "scale",
							KeyframeAnimation: scaleKeyframes,
						},
						{
							Name: "rotation",
							KeyframeAnimation: &KeyframeAnimation{
								Keyframes: []Keyframe{
									{
										Time:  "0s",
										Value: "0",
									},
									{
										Time:  ConvertSecondsToFCPDuration(15.0),
										Value: "360", // Full rotation
									},
								},
							},
						},
					},
				},
				FilterVideos: []FilterVideo{
					// Blend mode with animated opacity
					{
						Ref:  blendEffectID,
						Name: "Blend",
						Params: []Param{
							{
								Name:  "Blend Mode",
								Key:   "1",
								Value: "17 (Multiply)",
							},
							{
								Name:              "Opacity",
								Key:               "2",
								KeyframeAnimation: opacityKeyframes,
							},
						},
					},
					// Animated mask
					{
						Ref:  maskEffectID,
						Name: "Shape Mask",
						Params: []Param{
							{
								Name: "Radius",
								Key:  "160",
								KeyframeAnimation: &KeyframeAnimation{
									Keyframes: []Keyframe{
										{
											Time:  "0s",
											Value: "50 50", // Small mask
										},
										{
											Time:  ConvertSecondsToFCPDuration(8.0),
											Value: "300 200", // Large mask
										},
										{
											Time:  ConvertSecondsToFCPDuration(15.0),
											Value: "100 100", // Medium mask
										},
									},
								},
							},
							{
								Name:  "Curvature",
								Key:   "159",
								Value: "0.5",
							},
						},
					},
				},
			},
		},
	}

	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, mainClip)
	sequence.Duration = bgDuration

	// Validate and test
	violations := ValidateClaudeCompliance(fcpxml)
	if len(violations) > 0 {
		t.Errorf("CLAUDE.md compliance violations found:")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}

	// Write test file
	testFileName := "test_complex_compositing.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fmt.Printf("✅ Complex compositing test file created: %s\n", testFileName)
	fmt.Printf("   - Keyframe animations for position, scale, rotation\n")
	fmt.Printf("   - Multiple effects: blend, color correction, mask, vignette\n")
	fmt.Printf("   - Animated parameters with opacity and mask radius changes\n")
	fmt.Printf("   - Professional-grade compositing setup\n")
}