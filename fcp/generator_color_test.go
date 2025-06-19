package fcp

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestColorCorrection tests various color correction and grading effects
// to validate the FCP package can handle professional color workflows
// including color wheels, curves, masks, and secondary color correction.
func TestColorCorrection(t *testing.T) {
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	// Reserve IDs for video + multiple color effects
	ids := tx.ReserveIDs(10)
	videoID := ids[0]
	formatID := ids[1]
	colorBoardID := ids[2]
	colorWheelsID := ids[3]
	colorCurvesID := ids[4]
	exposureID := ids[5]
	saturationID := ids[6]
	colorMaskID := ids[7]
	lutID := ids[8]
	noiseReductionID := ids[9]

	// Create format for video
	_, err = tx.CreateFormatWithFrameDuration(formatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create format: %v", err)
	}

	// Create various color effects
	colorEffects := []struct {
		id   string
		name string
		uid  string
	}{
		{colorBoardID, "Color Board", "FFColorBoard"},
		{colorWheelsID, "Color Wheels", "FFColorWheels"},
		{colorCurvesID, "Color Curves", "FFColorCurves"},
		{exposureID, "Exposure", "FFExposure"},
		{saturationID, "Channel EQ", "FFChannelEQ"},
		{colorMaskID, "Color Mask", "FFColorMask"},
		{lutID, "Custom LUT", "FFCustomLUT"},
		{noiseReductionID, "Noise Reduction", "FFNoiseReduction"},
	}

	for _, effect := range colorEffects {
		_, err = tx.CreateEffect(effect.id, effect.name, effect.uid)
		if err != nil {
			t.Fatalf("Failed to create %s effect: %v", effect.name, err)
		}
	}

	// Create video asset
	videoDuration := ConvertSecondsToFCPDuration(20.0)
	videoAsset, err := tx.CreateAsset(videoID, "/Users/aa/cs/cutlass/assets/long.mov", "ColorGradingTest", videoDuration, formatID)
	if err != nil {
		t.Fatalf("Failed to create video asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build complex color grading stack
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	mainClip := AssetClip{
		Ref:       videoAsset.ID,
		Offset:    "0s",
		Name:      videoAsset.Name,
		Duration:  videoDuration,
		Format:    videoAsset.Format,
		TCFormat:  "NDF",
		AudioRole: "dialogue",
		FilterVideos: []FilterVideo{
			// Primary color correction with Color Board
			{
				Ref:  colorBoardID,
				Name: "Color Board",
				Params: []Param{
					{
						Name:  "Color",
						Key:   "9999/999166631/999166633/1/100/101",
						Value: "0.02 -0.05 0.1 1", // Cool shadows
					},
					{
						Name:  "Color",
						Key:   "9999/999166631/999166633/2/100/101",
						Value: "0.0 0.0 0.0 1", // Neutral midtones
					},
					{
						Name:  "Color",
						Key:   "9999/999166631/999166633/3/100/101",
						Value: "-0.03 0.02 -0.08 1", // Warm highlights
					},
					{
						Name:  "Saturation",
						Key:   "9999/999166631/999166633/1/100/103",
						Value: "1.2", // Increase shadow saturation
					},
					{
						Name:  "Saturation",
						Key:   "9999/999166631/999166633/2/100/103",
						Value: "1.1", // Slight midtone saturation boost
					},
					{
						Name:  "Saturation",
						Key:   "9999/999166631/999166633/3/100/103",
						Value: "0.9", // Reduce highlight saturation
					},
					{
						Name:  "Exposure",
						Key:   "9999/999166631/999166633/1/100/102",
						Value: "0.3", // Lift shadows
					},
					{
						Name:  "Exposure",
						Key:   "9999/999166631/999166633/2/100/102",
						Value: "0.1", // Slight midtone lift
					},
					{
						Name:  "Exposure",
						Key:   "9999/999166631/999166633/3/100/102",
						Value: "-0.2", // Pull down highlights
					},
				},
			},
			// Color Wheels for precise control
			{
				Ref:  colorWheelsID,
				Name: "Color Wheels",
				Params: []Param{
					{
						Name:  "Shadows",
						Key:   "9999/999166631/999166634/1/200/201",
						Value: "0.52 0.48 0.6 1", // Blue-tinted shadows
					},
					{
						Name:  "Midtones",
						Key:   "9999/999166631/999166634/2/200/201",
						Value: "0.5 0.5 0.5 1", // Neutral midtones
					},
					{
						Name:  "Highlights",
						Key:   "9999/999166631/999166634/3/200/201",
						Value: "0.48 0.52 0.45 1", // Magenta-tinted highlights
					},
					{
						Name:  "Master Saturation",
						Key:   "9999/999166631/999166634/4/200/203",
						Value: "1.15", // Overall saturation boost
					},
					{
						Name:  "Master Luminance",
						Key:   "9999/999166631/999166634/4/200/202",
						Value: "1.05", // Slight brightness increase
					},
				},
			},
			// Color Curves for advanced tonal control
			{
				Ref:  colorCurvesID,
				Name: "Color Curves",
				Params: []Param{
					{
						Name:  "Luma Curve",
						Key:   "9999/999166631/999166635/1/300/301",
						Value: "0.0,0.0 0.25,0.3 0.75,0.8 1.0,1.0", // S-curve for contrast
					},
					{
						Name:  "Red Curve",
						Key:   "9999/999166631/999166635/2/300/301",
						Value: "0.0,0.0 0.5,0.45 1.0,1.0", // Slight red reduction in midtones
					},
					{
						Name:  "Green Curve",
						Key:   "9999/999166631/999166635/3/300/301",
						Value: "0.0,0.0 0.5,0.5 1.0,1.0", // Neutral green
					},
					{
						Name:  "Blue Curve",
						Key:   "9999/999166631/999166635/4/300/301",
						Value: "0.0,0.0 0.5,0.55 1.0,1.0", // Slight blue boost in midtones
					},
				},
			},
			// Exposure adjustment
			{
				Ref:  exposureID,
				Name: "Exposure",
				Params: []Param{
					{
						Name:  "Exposure",
						Key:   "9999/999166631/999166636/1",
						Value: "0.2", // Slight overexposure for bright look
					},
					{
						Name:  "Black Point",
						Key:   "9999/999166631/999166636/2",
						Value: "0.05", // Raise blacks for film look
					},
					{
						Name:  "White Point",
						Key:   "9999/999166631/999166636/3",
						Value: "0.95", // Pull down whites slightly
					},
				},
			},
			// Channel EQ for fine saturation control
			{
				Ref:  saturationID,
				Name: "Channel EQ",
				Params: []Param{
					{
						Name:  "Red Saturation",
						Key:   "9999/999166631/999166637/1",
						Value: "0.9", // Reduce red saturation
					},
					{
						Name:  "Green Saturation",
						Key:   "9999/999166631/999166637/2",
						Value: "1.1", // Boost green saturation
					},
					{
						Name:  "Blue Saturation",
						Key:   "9999/999166631/999166637/3",
						Value: "1.2", // Boost blue saturation
					},
					{
						Name:  "Cyan Saturation",
						Key:   "9999/999166631/999166637/4",
						Value: "1.15", // Boost cyan
					},
					{
						Name:  "Magenta Saturation",
						Key:   "9999/999166631/999166637/5",
						Value: "0.85", // Reduce magenta
					},
					{
						Name:  "Yellow Saturation",
						Key:   "9999/999166631/999166637/6",
						Value: "0.95", // Slightly reduce yellow
					},
				},
			},
			// Color Mask for secondary color correction
			{
				Ref:  colorMaskID,
				Name: "Color Mask",
				Params: []Param{
					{
						Name:  "Sample Color",
						Key:   "9999/999166631/999166638/1",
						Value: "0.2 0.6 0.9 1", // Target sky blue
					},
					{
						Name:  "Color Range",
						Key:   "9999/999166631/999166638/2",
						Value: "0.15", // Moderate range
					},
					{
						Name:  "Softness",
						Key:   "9999/999166631/999166638/3",
						Value: "0.3", // Soft edge
					},
					{
						Name:  "Color Correction",
						Key:   "9999/999166631/999166638/4",
						Value: "0.8 0.9 1.1 1", // Cool the selected color
					},
					{
						Name:  "Saturation Adjustment",
						Key:   "9999/999166631/999166638/5",
						Value: "1.3", // Increase saturation of sky
					},
				},
			},
			// Custom LUT application
			{
				Ref:  lutID,
				Name: "Custom LUT",
				Params: []Param{
					{
						Name:  "LUT File",
						Key:   "9999/999166631/999166639/1",
						Value: "Cinematic_Look_01.cube", // Standard cinematic LUT
					},
					{
						Name:  "Mix",
						Key:   "9999/999166631/999166639/2",
						Value: "0.7", // 70% LUT strength
					},
					{
						Name:  "Preserve Luminance",
						Key:   "9999/999166631/999166639/3",
						Value: "1", // Enable luminance preservation
					},
				},
			},
			// Noise Reduction for clean image
			{
				Ref:  noiseReductionID,
				Name: "Noise Reduction",
				Params: []Param{
					{
						Name:  "Amount",
						Key:   "9999/999166631/999166640/1",
						Value: "25", // Moderate noise reduction
					},
					{
						Name:  "Chroma",
						Key:   "9999/999166631/999166640/2",
						Value: "15", // Reduce color noise
					},
					{
						Name:  "Luma",
						Key:   "9999/999166631/999166640/3",
						Value: "10", // Light luminance noise reduction
					},
					{
						Name:  "Preserve Details",
						Key:   "9999/999166631/999166640/4",
						Value: "1", // Maintain sharpness
					},
				},
			},
		},
	}

	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, mainClip)
	sequence.Duration = videoDuration

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

	// Verify color correction structure in generated XML
	testCases := []struct {
		name     string
		expected string
	}{
		{"Color Board effect", `name="Color Board"`},
		{"Color Wheels effect", `name="Color Wheels"`},
		{"Color Curves effect", `name="Color Curves"`},
		{"Exposure effect", `name="Exposure"`},
		{"Channel EQ effect", `name="Channel EQ"`},
		{"Color Mask effect", `name="Color Mask"`},
		{"Custom LUT effect", `name="Custom LUT"`},
		{"Noise Reduction effect", `name="Noise Reduction"`},
		{"Shadow color parameter", `<param name="Color"`},
		{"Saturation parameter", `<param name="Saturation"`},
		{"Exposure parameter", `<param name="Exposure"`},
		{"Color curve data", `0.0,0.0 0.25,0.3`},
		{"LUT file reference", `Cinematic_Look_01.cube`},
		{"Noise reduction amount", `<param name="Amount"`},
		{"Multiple filter-video elements", `</filter-video>`},
	}

	for _, tc := range testCases {
		if !strings.Contains(xmlContent, tc.expected) {
			t.Errorf("Test '%s' failed: expected '%s' not found in XML", tc.name, tc.expected)
		}
	}

	// Write test file for manual FCP validation
	testFileName := "/tmp/test_color_correction.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFileName); os.IsNotExist(err) {
		t.Errorf("Test file was not created: %s", testFileName)
	}

	fmt.Printf("✅ Color correction test file created: %s\n", testFileName)
	fmt.Printf("   - Complete color grading pipeline with 8 effects\n")
	fmt.Printf("   - Primary correction (Color Board, Color Wheels)\n")
	fmt.Printf("   - Advanced tools (Color Curves, Exposure, Channel EQ)\n")
	fmt.Printf("   - Secondary correction (Color Mask)\n")
	fmt.Printf("   - Creative tools (Custom LUT, Noise Reduction)\n")
	fmt.Printf("   - Professional color workflow ready for FCP validation\n")
}

// TestHSLColorGrading tests HSL-based color grading with multiple corrections
func TestHSLColorGrading(t *testing.T) {
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	// Reserve IDs for assets and effects
	ids := tx.ReserveIDs(8)
	videoID := ids[0]
	formatID := ids[1]
	hueSatID := ids[2]
	colorBalanceID := ids[3]
	brightnessContrastID := ids[4]
	vibanceID := ids[5]
	shadowHighlightID := ids[6]
	keymixID := ids[7]

	// Create format and asset
	_, err = tx.CreateFormatWithFrameDuration(formatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create format: %v", err)
	}

	// Create HSL and color grading effects
	effects := []struct {
		id   string
		name string
		uid  string
	}{
		{hueSatID, "Hue/Saturation Curves", "FFHueSatCurves"},
		{colorBalanceID, "Color Balance", "FFColorBalance"},
		{brightnessContrastID, "Brightness & Contrast", "FFBrightnessContrast"},
		{vibanceID, "Vibrance", "FFVibrance"},
		{shadowHighlightID, "Shadow/Highlight", "FFShadowHighlight"},
		{keymixID, "Keyer", "FFKeyer"},
	}

	for _, effect := range effects {
		_, err = tx.CreateEffect(effect.id, effect.name, effect.uid)
		if err != nil {
			t.Fatalf("Failed to create %s effect: %v", effect.name, err)
		}
	}

	videoDuration := ConvertSecondsToFCPDuration(15.0)
	videoAsset, err := tx.CreateAsset(videoID, "/Users/aa/cs/cutlass/assets/speech1.mov", "HSLGradingTest", videoDuration, formatID)
	if err != nil {
		t.Fatalf("Failed to create video asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build HSL-focused color grading
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Create keyframe animation for dynamic color grading
	hueShiftKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "0", // No hue shift at start
			},
			{
				Time:  ConvertSecondsToFCPDuration(5.0),
				Value: "15", // Shift towards orange
			},
			{
				Time:  ConvertSecondsToFCPDuration(10.0),
				Value: "-10", // Shift towards cyan
			},
			{
				Time:  ConvertSecondsToFCPDuration(15.0),
				Value: "0", // Return to neutral
			},
		},
	}

	saturationKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "100", // Normal saturation
			},
			{
				Time:  ConvertSecondsToFCPDuration(7.5),
				Value: "150", // Boost saturation
			},
			{
				Time:  ConvertSecondsToFCPDuration(15.0),
				Value: "80", // Desaturate slightly
			},
		},
	}

	mainClip := AssetClip{
		Ref:       videoAsset.ID,
		Offset:    "0s",
		Name:      videoAsset.Name,
		Duration:  videoDuration,
		Format:    videoAsset.Format,
		TCFormat:  "NDF",
		AudioRole: "dialogue",
		FilterVideos: []FilterVideo{
			// Hue/Saturation Curves for selective color control
			{
				Ref:  hueSatID,
				Name: "Hue/Saturation Curves",
				Params: []Param{
					{
						Name:              "Red Hue",
						Key:               "9999/999166631/999166641/1",
						KeyframeAnimation: hueShiftKeyframes,
					},
					{
						Name:              "Red Saturation",
						Key:               "9999/999166631/999166641/2",
						KeyframeAnimation: saturationKeyframes,
					},
					{
						Name:  "Orange Hue",
						Key:   "9999/999166631/999166641/3",
						Value: "5", // Slight orange shift
					},
					{
						Name:  "Orange Saturation",
						Key:   "9999/999166631/999166641/4",
						Value: "120", // Boost orange
					},
					{
						Name:  "Yellow Hue",
						Key:   "9999/999166631/999166641/5",
						Value: "-8", // Shift yellow towards orange
					},
					{
						Name:  "Yellow Saturation",
						Key:   "9999/999166631/999166641/6",
						Value: "90", // Reduce yellow saturation
					},
					{
						Name:  "Green Hue",
						Key:   "9999/999166631/999166641/7",
						Value: "0", // Keep green neutral
					},
					{
						Name:  "Green Saturation",
						Key:   "9999/999166631/999166641/8",
						Value: "110", // Slight green boost
					},
					{
						Name:  "Cyan Hue",
						Key:   "9999/999166631/999166641/9",
						Value: "10", // Shift cyan towards blue
					},
					{
						Name:  "Cyan Saturation",
						Key:   "9999/999166631/999166641/10",
						Value: "130", // Boost cyan saturation
					},
					{
						Name:  "Blue Hue",
						Key:   "9999/999166631/999166641/11",
						Value: "-5", // Shift blue slightly towards cyan
					},
					{
						Name:  "Blue Saturation",
						Key:   "9999/999166631/999166641/12",
						Value: "125", // Boost blue saturation
					},
					{
						Name:  "Magenta Hue",
						Key:   "9999/999166631/999166641/13",
						Value: "0", // Keep magenta neutral
					},
					{
						Name:  "Magenta Saturation",
						Key:   "9999/999166631/999166641/14",
						Value: "85", // Reduce magenta saturation
					},
				},
			},
			// Color Balance for temperature and tint
			{
				Ref:  colorBalanceID,
				Name: "Color Balance",
				Params: []Param{
					{
						Name:  "Temperature",
						Key:   "9999/999166631/999166642/1",
						Value: "200", // Warmer temperature (200K shift)
					},
					{
						Name:  "Tint",
						Key:   "9999/999166631/999166642/2",
						Value: "-0.05", // Slight green tint
					},
					{
						Name: "Shadow Temperature",
						Key:  "9999/999166631/999166642/3",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{
									Time:  "0s",
									Value: "100", // Cool shadows
								},
								{
									Time:  ConvertSecondsToFCPDuration(15.0),
									Value: "300", // Warm shadows
								},
							},
						},
					},
					{
						Name:  "Highlight Temperature",
						Key:   "9999/999166631/999166642/4",
						Value: "-100", // Cool highlights
					},
				},
			},
			// Brightness & Contrast for tonal control
			{
				Ref:  brightnessContrastID,
				Name: "Brightness & Contrast",
				Params: []Param{
					{
						Name:  "Brightness",
						Key:   "9999/999166631/999166643/1",
						Value: "10", // Slight brightness boost
					},
					{
						Name: "Contrast",
						Key:  "9999/999166631/999166643/2",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{
									Time:  "0s",
									Value: "100", // Normal contrast
								},
								{
									Time:  ConvertSecondsToFCPDuration(7.5),
									Value: "130", // High contrast
								},
								{
									Time:  ConvertSecondsToFCPDuration(15.0),
									Value: "90", // Low contrast
								},
							},
						},
					},
				},
			},
			// Vibrance for intelligent saturation
			{
				Ref:  vibanceID,
				Name: "Vibrance",
				Params: []Param{
					{
						Name:  "Vibrance",
						Key:   "9999/999166631/999166644/1",
						Value: "25", // Boost vibrance
					},
					{
						Name:  "Saturation",
						Key:   "9999/999166631/999166644/2",
						Value: "15", // Additional saturation
					},
					{
						Name:  "Skin Tone Protection",
						Key:   "9999/999166631/999166644/3",
						Value: "1", // Protect skin tones
					},
				},
			},
			// Shadow/Highlight for exposure recovery
			{
				Ref:  shadowHighlightID,
				Name: "Shadow/Highlight",
				Params: []Param{
					{
						Name:  "Shadow Amount",
						Key:   "9999/999166631/999166645/1",
						Value: "30", // Lift shadows
					},
					{
						Name:  "Highlight Amount",
						Key:   "9999/999166631/999166645/2",
						Value: "20", // Recover highlights
					},
					{
						Name:  "Shadow Radius",
						Key:   "9999/999166631/999166645/3",
						Value: "50", // Shadow recovery radius
					},
					{
						Name:  "Highlight Radius",
						Key:   "9999/999166631/999166645/4",
						Value: "30", // Highlight recovery radius
					},
					{
						Name:  "Color Correction",
						Key:   "9999/999166631/999166645/5",
						Value: "20", // Color correction amount
					},
				},
			},
			// Keyer for advanced masking
			{
				Ref:  keymixID,
				Name: "Keyer",
				Params: []Param{
					{
						Name:  "Key Method",
						Key:   "9999/999166631/999166646/1",
						Value: "0", // Color key
					},
					{
						Name:  "Key Color",
						Key:   "9999/999166631/999166646/2",
						Value: "0.1 0.4 0.8 1", // Blue sky key
					},
					{
						Name:  "Tolerance",
						Key:   "9999/999166631/999166646/3",
						Value: "0.2", // Key tolerance
					},
					{
						Name:  "Softness",
						Key:   "9999/999166631/999166646/4",
						Value: "0.1", // Edge softness
					},
					{
						Name:  "Fill Color",
						Key:   "9999/999166631/999166646/5",
						Value: "0.9 0.7 0.4 1", // Warm replacement color
					},
				},
			},
		},
	}

	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, mainClip)
	sequence.Duration = videoDuration

	// Validate and test
	violations := ValidateClaudeCompliance(fcpxml)
	if len(violations) > 0 {
		t.Errorf("CLAUDE.md compliance violations found:")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}

	// Write test file
	testFileName := "/tmp/test_hsl_color_grading.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fmt.Printf("✅ HSL color grading test file created: %s\n", testFileName)
	fmt.Printf("   - HSL-based color correction workflow\n")
	fmt.Printf("   - Animated hue shifts and saturation changes\n")
	fmt.Printf("   - Color balance with temperature/tint controls\n")
	fmt.Printf("   - Brightness/contrast with keyframe animation\n")
	fmt.Printf("   - Vibrance with skin tone protection\n")
	fmt.Printf("   - Shadow/highlight recovery tools\n")
	fmt.Printf("   - Advanced keying for color replacement\n")
}