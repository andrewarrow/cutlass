package fcp

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

// TestAddImageWithSlide tests the slide animation functionality
func TestAddImageWithSlide(t *testing.T) {
	// Create test image file
	testImagePath := "test_slide_image.png"
	err := os.WriteFile(testImagePath, []byte("fake png data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer os.Remove(testImagePath)

	// Generate empty FCPXML
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to generate empty FCPXML: %v", err)
	}

	// Add image with slide animation
	err = AddImageWithSlide(fcpxml, testImagePath, 9.0, true)
	if err != nil {
		t.Fatalf("Failed to add image with slide: %v", err)
	}

	// Verify the structure
	if len(fcpxml.Resources.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(fcpxml.Resources.Assets))
	}

	if len(fcpxml.Resources.Formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(fcpxml.Resources.Formats))
	}

	// Check if video element has adjust-transform with slide animation
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 1 {
		t.Fatalf("Expected 1 video element, got %d", len(sequence.Spine.Videos))
	}

	video := sequence.Spine.Videos[0]
	if video.AdjustTransform == nil {
		t.Fatalf("Expected video to have adjust-transform")
	}

	// Verify keyframe animation parameters
	params := video.AdjustTransform.Params
	if len(params) != 4 {
		t.Errorf("Expected 4 animation params (anchor, position, rotation, scale), got %d", len(params))
	}

	// Check parameter names
	expectedParams := []string{"anchor", "position", "rotation", "scale"}
	for i, expectedParam := range expectedParams {
		if i >= len(params) {
			t.Errorf("Missing param: %s", expectedParam)
			continue
		}
		if params[i].Name != expectedParam {
			t.Errorf("Expected param %s, got %s", expectedParam, params[i].Name)
		}
	}

	// Verify position keyframes specifically
	positionParam := params[1] // position is second param
	if positionParam.KeyframeAnimation == nil {
		t.Fatalf("Position param should have keyframe animation")
	}

	keyframes := positionParam.KeyframeAnimation.Keyframes
	if len(keyframes) != 2 {
		t.Errorf("Expected 2 position keyframes, got %d", len(keyframes))
	}

	// Check keyframe values for Ken Burns effect
	if keyframes[0].Value != "0 0" {
		t.Errorf("Expected first keyframe value '0 0', got '%s'", keyframes[0].Value)
	}
	if keyframes[1].Value != "-20 -15" {
		t.Errorf("Expected second keyframe value '-20 -15', got '%s'", keyframes[1].Value)
	}

	// Check keyframe timing matches Ken Burns (3 second duration)
	if keyframes[0].Time != "86399313/24000s" {
		t.Errorf("Expected first keyframe time '86399313/24000s', got '%s'", keyframes[0].Time)
	}
	if keyframes[1].Time != "86471385/24000s" {
		t.Errorf("Expected second keyframe time '86471385/24000s', got '%s'", keyframes[1].Time)
	}

	// Verify curve attributes on static keyframes
	anchorParam := params[0]
	if len(anchorParam.KeyframeAnimation.Keyframes) > 0 {
		anchorKeyframe := anchorParam.KeyframeAnimation.Keyframes[0]
		if anchorKeyframe.Curve != "linear" {
			t.Errorf("Expected anchor keyframe curve 'linear', got '%s'", anchorKeyframe.Curve)
		}
	}
}

// TestAddImageWithoutSlide tests that images without slide don't have animations
func TestAddImageWithoutSlide(t *testing.T) {
	// Create test image file
	testImagePath := "test_no_slide_image.png"
	err := os.WriteFile(testImagePath, []byte("fake png data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer os.Remove(testImagePath)

	// Generate empty FCPXML
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to generate empty FCPXML: %v", err)
	}

	// Add image without slide animation
	err = AddImageWithSlide(fcpxml, testImagePath, 9.0, false)
	if err != nil {
		t.Fatalf("Failed to add image without slide: %v", err)
	}

	// Check if video element has no adjust-transform
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 1 {
		t.Fatalf("Expected 1 video element, got %d", len(sequence.Spine.Videos))
	}

	video := sequence.Spine.Videos[0]
	if video.AdjustTransform != nil {
		t.Errorf("Expected video to have no adjust-transform, but it has one")
	}
}

// TestSlideAnimationXMLOutput tests the actual XML output structure
func TestSlideAnimationXMLOutput(t *testing.T) {
	// Create test image file
	testImagePath := "test_xml_slide_image.png"
	err := os.WriteFile(testImagePath, []byte("fake png data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer os.Remove(testImagePath)

	// Generate empty FCPXML
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to generate empty FCPXML: %v", err)
	}

	// Add image with slide animation
	err = AddImageWithSlide(fcpxml, testImagePath, 9.0, true)
	if err != nil {
		t.Fatalf("Failed to add image with slide: %v", err)
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	xmlString := string(output)

	// Check for key XML elements that should be present
	expectedElements := []string{
		"<adjust-transform>",
		`<param name="anchor">`,
		`<param name="position">`,
		`<param name="rotation">`,
		`<param name="scale">`,
		"<keyframeAnimation>",
		`time="86399313/24000s"`,
		`time="86471385/24000s"`,
		`value="0 0"`,
		`value="-20 -15"`,
		`curve="linear"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(xmlString, expected) {
			t.Errorf("Expected XML to contain '%s', but it doesn't", expected)
		}
	}

	// Check that empty key/value attributes are not present
	if strings.Contains(xmlString, `key=""`) {
		t.Errorf("XML should not contain empty key attributes")
	}
	if strings.Contains(xmlString, `value=""`) {
		t.Errorf("XML should not contain empty value attributes on params")
	}
}

// TestCreateKenBurnsAnimation tests the Ken Burns animation creation function directly
func TestCreateKenBurnsAnimation(t *testing.T) {
	// Test the createKenBurnsAnimation function
	adjustTransform := createKenBurnsAnimation("0s", 9.0)

	if adjustTransform == nil {
		t.Fatalf("createKenBurnsAnimation returned nil")
	}

	// Check that we have 4 parameters
	if len(adjustTransform.Params) != 4 {
		t.Errorf("Expected 4 params, got %d", len(adjustTransform.Params))
	}

	// Verify position parameter keyframes
	var positionParam *Param
	for _, param := range adjustTransform.Params {
		if param.Name == "position" {
			positionParam = &param
			break
		}
	}

	if positionParam == nil {
		t.Fatalf("Could not find position parameter")
	}

	if positionParam.KeyframeAnimation == nil {
		t.Fatalf("Position parameter should have keyframe animation")
	}

	keyframes := positionParam.KeyframeAnimation.Keyframes
	if len(keyframes) != 2 {
		t.Errorf("Expected 2 position keyframes, got %d", len(keyframes))
	}

	// Test timing calculation (should be exactly 3 seconds apart for Ken Burns)
	// 86471385 - 86399313 = 72072 frames = exactly 3 seconds in 1001/24000s timebase
	if keyframes[1].Time != "86471385/24000s" {
		t.Errorf("Expected end time 86471385/24000s, got %s", keyframes[1].Time)
	}
	if keyframes[0].Time != "86399313/24000s" {
		t.Errorf("Expected start time 86399313/24000s, got %s", keyframes[0].Time)
	}

	// Verify the Ken Burns position values (starts at center, pans slightly)
	if keyframes[0].Value != "0 0" {
		t.Errorf("Expected start position '0 0', got '%s'", keyframes[0].Value)
	}
	if keyframes[1].Value != "-20 -15" {
		t.Errorf("Expected end position '-20 -15', got '%s'", keyframes[1].Value)
	}
}

// TestSlideAnimationBackwardsCompatibility tests that AddImage still works without slide
func TestSlideAnimationBackwardsCompatibility(t *testing.T) {
	// Create test image file
	testImagePath := "test_compat_image.png"
	err := os.WriteFile(testImagePath, []byte("fake png data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer os.Remove(testImagePath)

	// Generate empty FCPXML
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to generate empty FCPXML: %v", err)
	}

	// Test that the original AddImage function still works (should call AddImageWithSlide with false)
	err = AddImage(fcpxml, testImagePath, 9.0)
	if err != nil {
		t.Fatalf("Failed to add image using original AddImage function: %v", err)
	}

	// Verify no animation was added
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 1 {
		t.Fatalf("Expected 1 video element, got %d", len(sequence.Spine.Videos))
	}

	video := sequence.Spine.Videos[0]
	if video.AdjustTransform != nil {
		t.Errorf("AddImage should not add animation, but adjust-transform was found")
	}
}

// TestAddSlideToVideoAtOffset tests that video files maintain AssetClip structure with timeline-based animation
func TestAddSlideToVideoAtOffset(t *testing.T) {
	// Create test video file (mock with .mov extension)
	testVideoPath := "test_slide_video.mov"
	err := os.WriteFile(testVideoPath, []byte("fake mov data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test video file: %v", err)
	}
	defer os.Remove(testVideoPath)

	// Generate empty FCPXML
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to generate empty FCPXML: %v", err)
	}

	// Add video (should create AssetClip)
	err = AddVideo(fcpxml, testVideoPath)
	if err != nil {
		t.Fatalf("Failed to add video: %v", err)
	}

	// Verify video was added as AssetClip
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.AssetClips) != 1 {
		t.Fatalf("Expected 1 asset-clip, got %d", len(sequence.Spine.AssetClips))
	}
	if len(sequence.Spine.Videos) != 0 {
		t.Fatalf("Expected 0 video elements for video file, got %d", len(sequence.Spine.Videos))
	}

	// Add slide animation at offset 0
	err = AddSlideToVideoAtOffset(fcpxml, 0.0)
	if err != nil {
		t.Fatalf("Failed to add slide animation to video: %v", err)
	}

	// Verify AssetClip still exists (not converted to Video)
	if len(sequence.Spine.AssetClips) != 1 {
		t.Fatalf("Expected 1 asset-clip after animation, got %d", len(sequence.Spine.AssetClips))
	}
	if len(sequence.Spine.Videos) != 0 {
		t.Fatalf("Expected 0 video elements after animation (AssetClip should be preserved), got %d", len(sequence.Spine.Videos))
	}

	// Verify AssetClip has adjust-transform
	assetClip := sequence.Spine.AssetClips[0]
	if assetClip.AdjustTransform == nil {
		t.Fatalf("Expected AssetClip to have adjust-transform after adding slide animation")
	}

	// Verify animation parameters
	params := assetClip.AdjustTransform.Params
	if len(params) != 4 {
		t.Errorf("Expected 4 animation params (anchor, position, rotation, scale), got %d", len(params))
	}

	// Check parameter names
	expectedParams := []string{"anchor", "position", "rotation", "scale"}
	for i, expectedParam := range expectedParams {
		if i >= len(params) {
			t.Errorf("Missing param: %s", expectedParam)
			continue
		}
		if params[i].Name != expectedParam {
			t.Errorf("Expected param %s, got %s", expectedParam, params[i].Name)
		}
	}

	// Verify position keyframes use timeline-based timing (not image start time)
	positionParam := params[1] // position is second param
	if positionParam.KeyframeAnimation == nil {
		t.Fatalf("Position param should have keyframe animation")
	}

	keyframes := positionParam.KeyframeAnimation.Keyframes
	if len(keyframes) != 2 {
		t.Errorf("Expected 2 position keyframes, got %d", len(keyframes))
	}

	// Check keyframe values (same as image animation)
	if keyframes[0].Value != "0 0" {
		t.Errorf("Expected first keyframe value '0 0', got '%s'", keyframes[0].Value)
	}
	if keyframes[1].Value != "-20 -15" {
		t.Errorf("Expected second keyframe value '-20 -15', got '%s'", keyframes[1].Value)
	}

	// Check timeline-based timing (NOT image-based timing)
	// Should start at clip offset (0s) and go for 3 frame-aligned seconds (Ken Burns)
	if keyframes[0].Time != "0/24000s" {
		t.Errorf("Expected first keyframe time '0/24000s' (timeline-based), got '%s'", keyframes[0].Time)
	}
	if keyframes[1].Time != "72072/24000s" {
		t.Errorf("Expected second keyframe time '72072/24000s' (frame-aligned 3 seconds), got '%s'", keyframes[1].Time)
	}

	// Verify this is NOT using image start time
	if keyframes[0].Time == "86399313/24000s" || keyframes[1].Time == "86423337/24000s" {
		t.Errorf("AssetClip animation should use timeline-based timing, not image start time")
	}
}

// TestCreateAssetClipKenBurnsAnimation tests the AssetClip Ken Burns animation creation function
func TestCreateAssetClipKenBurnsAnimation(t *testing.T) {
	// Test the createAssetClipKenBurnsAnimation function with offset 0
	adjustTransform := createAssetClipKenBurnsAnimation("0s", 1.0)

	if adjustTransform == nil {
		t.Fatalf("createAssetClipKenBurnsAnimation returned nil")
	}

	// Check that we have 4 parameters
	if len(adjustTransform.Params) != 4 {
		t.Errorf("Expected 4 params, got %d", len(adjustTransform.Params))
	}

	// Verify position parameter keyframes
	var positionParam *Param
	for _, param := range adjustTransform.Params {
		if param.Name == "position" {
			positionParam = &param
			break
		}
	}

	if positionParam == nil {
		t.Fatalf("Could not find position parameter")
	}

	if positionParam.KeyframeAnimation == nil {
		t.Fatalf("Position parameter should have keyframe animation")
	}

	keyframes := positionParam.KeyframeAnimation.Keyframes
	if len(keyframes) != 2 {
		t.Errorf("Expected 2 position keyframes, got %d", len(keyframes))
	}

	// Test timeline-based timing (should start at offset and go for 3 seconds for Ken Burns)
	if keyframes[0].Time != "0/24000s" {
		t.Errorf("Expected start time 0/24000s, got %s", keyframes[0].Time)
	}
	if keyframes[1].Time != "72072/24000s" {
		t.Errorf("Expected end time 72072/24000s (frame-aligned 3 seconds), got %s", keyframes[1].Time)
	}

	// Verify the Ken Burns position values (starts at center, pans slightly)
	if keyframes[0].Value != "0 0" {
		t.Errorf("Expected start position '0 0', got '%s'", keyframes[0].Value)
	}
	if keyframes[1].Value != "-20 -15" {
		t.Errorf("Expected end position '-20 -15', got '%s'", keyframes[1].Value)
	}
}

// TestVideoSlideAnimationXMLOutput tests that video slide animation produces correct XML structure
func TestVideoSlideAnimationXMLOutput(t *testing.T) {
	// Create test video file
	testVideoPath := "test_xml_video_slide.mov"
	err := os.WriteFile(testVideoPath, []byte("fake mov data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test video file: %v", err)
	}
	defer os.Remove(testVideoPath)

	// Generate empty FCPXML and add video with slide
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to generate empty FCPXML: %v", err)
	}

	err = AddVideo(fcpxml, testVideoPath)
	if err != nil {
		t.Fatalf("Failed to add video: %v", err)
	}

	err = AddSlideToVideoAtOffset(fcpxml, 0.0)
	if err != nil {
		t.Fatalf("Failed to add slide animation: %v", err)
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	xmlString := string(output)

	// Check for AssetClip with adjust-transform (not Video element)
	expectedElements := []string{
		"<asset-clip",
		"<adjust-transform>",
		`<param name="anchor">`,
		`<param name="position">`,
		`<param name="rotation">`,
		`<param name="scale">`,
		"<keyframeAnimation>",
		`time="0/24000s"`,        // Timeline-based start
		`time="72072/24000s"`,    // Frame-aligned 3 seconds
		`value="0 0"`,
		`value="-20 -15"`,
		`curve="linear"`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(xmlString, expected) {
			t.Errorf("Expected XML to contain '%s', but it doesn't", expected)
		}
	}

	// Check that Video element is NOT present for video files
	if strings.Contains(xmlString, "<video ") {
		t.Errorf("XML should not contain <video> elements for video files - should use <asset-clip>")
	}

	// Check that image-based timing is NOT present
	if strings.Contains(xmlString, "86399313/24000s") || strings.Contains(xmlString, "86423337/24000s") {
		t.Errorf("XML should not contain image-based timing for video files")
	}
}
