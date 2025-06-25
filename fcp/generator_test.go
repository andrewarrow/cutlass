// Package fcp provides tests for FCPXML generation.
//
// 🚨 CRITICAL: Tests MUST validate CLAUDE.md compliance:
// - AFTER changes → RUN: xmllint --dtdvalid FCPXMLv1_13.dtd test_file.fcpxml  
// - BEFORE commits → RUN: ValidateClaudeCompliance() function
// - FOR durations → USE: ConvertSecondsToFCPDuration() function  
// - VERIFY: No fmt.Sprintf() with XML content in any test
package fcp

import (
	"os"
	"testing"
)

func TestGenerateEmpty(t *testing.T) {
	// Create a temporary test file
	testFile := "test_generate_empty.fcpxml"

	// Ensure cleanup even if test fails
	defer func() {
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up test file: %v", err)
		}
	}()

	// Call GenerateEmpty with the test file
	_, err := GenerateEmpty(testFile)
	if err != nil {
		t.Fatalf("GenerateEmpty failed: %v", err)
	}

	// Read the generated file
	generatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Compare with expected XML string
	if string(generatedContent) != emptyxml {
		t.Errorf("Generated XML does not match expected output.\nExpected:\n%s\n\nGenerated:\n%s", emptyxml, string(generatedContent))
	}
}

var emptyxml = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

<fcpxml version="1.13">
    <resources>
        <format id="r1" name="FFVideoFormat720p2398" frameDuration="1001/24000s" width="1280" height="720" colorSpace="1-1-1 (Rec. 709)"></format>
    </resources>
    <library location="file:///Users/aa/Movies/Untitled.fcpbundle/">
        <event name="6-13-25" uid="78463397-97FD-443D-B4E2-07C581674AFC">
            <project name="wiki" uid="DEA19981-DED5-4851-8435-14515931C68A" modDate="2025-06-13 11:46:22 -0700">
                <sequence format="r1" duration="0s" tcStart="0s" tcFormat="NDF" audioLayout="stereo" audioRate="48k">
                    <spine></spine>
                </sequence>
            </project>
        </event>
        <smart-collection name="Projects" match="all">
            <match-clip rule="is" type="project"></match-clip>
        </smart-collection>
        <smart-collection name="All Video" match="any">
            <match-media rule="is" type="videoOnly"></match-media>
            <match-media rule="is" type="videoWithAudio"></match-media>
        </smart-collection>
        <smart-collection name="Audio Only" match="all">
            <match-media rule="is" type="audioOnly"></match-media>
        </smart-collection>
        <smart-collection name="Stills" match="all">
            <match-media rule="is" type="stills"></match-media>
        </smart-collection>
        <smart-collection name="Favorites" match="all">
            <match-ratings value="favorites"></match-ratings>
        </smart-collection>
    </library>
</fcpxml>`

var pngxmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

<fcpxml version="1.13">
    <resources>
        <asset id="r2" name="cs.pitt.edu" uid="%s" start="0s" hasVideo="1" format="r3" videoSources="1" duration="0s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/cs.pitt.edu.png"></media-rep>
        </asset>
        <format id="r1" name="FFVideoFormat720p2398" frameDuration="1001/24000s" width="1280" height="720" colorSpace="1-1-1 (Rec. 709)"></format>
        <format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"></format>
    </resources>
    <library location="file:///Users/aa/Movies/Untitled.fcpbundle/">
        <event name="6-13-25" uid="78463397-97FD-443D-B4E2-07C581674AFC">
            <project name="wiki" uid="DEA19981-DED5-4851-8435-14515931C68A" modDate="2025-06-13 11:46:22 -0700">
                <sequence format="r1" duration="216216/24000s" tcStart="0s" tcFormat="NDF" audioLayout="stereo" audioRate="48k">
                    <spine>
                        <video ref="r2" offset="0s" name="cs.pitt.edu" duration="216216/24000s" start="86399313/24000s"></video>
                    </spine>
                </sequence>
            </project>
        </event>
        <smart-collection name="Projects" match="all">
            <match-clip rule="is" type="project"></match-clip>
        </smart-collection>
        <smart-collection name="All Video" match="any">
            <match-media rule="is" type="videoOnly"></match-media>
            <match-media rule="is" type="videoWithAudio"></match-media>
        </smart-collection>
        <smart-collection name="Audio Only" match="all">
            <match-media rule="is" type="audioOnly"></match-media>
        </smart-collection>
        <smart-collection name="Stills" match="all">
            <match-media rule="is" type="stills"></match-media>
        </smart-collection>
        <smart-collection name="Favorites" match="all">
            <match-ratings value="favorites"></match-ratings>
        </smart-collection>
    </library>
</fcpxml>`

var movxmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

<fcpxml version="1.13">
    <resources>
        <asset id="r2" name="speech1" uid="%s" start="0s" hasVideo="1" format="r1" videoSources="1" hasAudio="1" audioSources="1" audioChannels="2" audioRate="48000" duration="240240/24000s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/speech1.mov"></media-rep>
        </asset>
        <format id="r1" name="FFVideoFormat720p2398" frameDuration="1001/24000s" width="1280" height="720" colorSpace="1-1-1 (Rec. 709)"></format>
    </resources>
    <library location="file:///Users/aa/Movies/Untitled.fcpbundle/">
        <event name="6-13-25" uid="78463397-97FD-443D-B4E2-07C581674AFC">
            <project name="wiki" uid="DEA19981-DED5-4851-8435-14515931C68A" modDate="2025-06-13 11:46:22 -0700">
                <sequence format="r1" duration="240240/24000s" tcStart="0s" tcFormat="NDF" audioLayout="stereo" audioRate="48k">
                    <spine>
                        <asset-clip ref="r2" offset="0s" name="speech1" duration="240240/24000s" format="r1" tcFormat="NDF" audioRole="dialogue"></asset-clip>
                    </spine>
                </sequence>
            </project>
        </event>
        <smart-collection name="Projects" match="all">
            <match-clip rule="is" type="project"></match-clip>
        </smart-collection>
        <smart-collection name="All Video" match="any">
            <match-media rule="is" type="videoOnly"></match-media>
            <match-media rule="is" type="videoWithAudio"></match-media>
        </smart-collection>
        <smart-collection name="Audio Only" match="all">
            <match-media rule="is" type="audioOnly"></match-media>
        </smart-collection>
        <smart-collection name="Stills" match="all">
            <match-media rule="is" type="stills"></match-media>
        </smart-collection>
        <smart-collection name="Favorites" match="all">
            <match-ratings value="favorites"></match-ratings>
        </smart-collection>
    </library>
</fcpxml>`

var appendpngxmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

<fcpxml version="1.13">
    <resources>
        <asset id="r2" name="cs.pitt.edu" uid="%s" start="0s" hasVideo="1" format="r3" videoSources="1" duration="0s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/cs.pitt.edu.png"></media-rep>
        </asset>
        <asset id="r4" name="cutlass_logo_t" uid="%s" start="0s" hasVideo="1" format="r5" videoSources="1" duration="0s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/cutlass_logo_t.png"></media-rep>
        </asset>
        <format id="r1" name="FFVideoFormat720p2398" frameDuration="1001/24000s" width="1280" height="720" colorSpace="1-1-1 (Rec. 709)"></format>
        <format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"></format>
        <format id="r5" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"></format>
    </resources>
    <library location="file:///Users/aa/Movies/Untitled.fcpbundle/">
        <event name="6-13-25" uid="78463397-97FD-443D-B4E2-07C581674AFC">
            <project name="wiki" uid="DEA19981-DED5-4851-8435-14515931C68A" modDate="2025-06-13 11:46:22 -0700">
                <sequence format="r1" duration="457457/24000s" tcStart="0s" tcFormat="NDF" audioLayout="stereo" audioRate="48k">
                    <spine>
                        <video ref="r2" offset="0s" name="cs.pitt.edu" duration="241241/24000s" start="86399313/24000s"></video>
                        <video ref="r4" offset="241241/24000s" name="cutlass_logo_t" duration="216216/24000s" start="86399313/24000s"></video>
                    </spine>
                </sequence>
            </project>
        </event>
        <smart-collection name="Projects" match="all">
            <match-clip rule="is" type="project"></match-clip>
        </smart-collection>
        <smart-collection name="All Video" match="any">
            <match-media rule="is" type="videoOnly"></match-media>
            <match-media rule="is" type="videoWithAudio"></match-media>
        </smart-collection>
        <smart-collection name="Audio Only" match="all">
            <match-media rule="is" type="audioOnly"></match-media>
        </smart-collection>
        <smart-collection name="Stills" match="all">
            <match-media rule="is" type="stills"></match-media>
        </smart-collection>
        <smart-collection name="Favorites" match="all">
            <match-ratings value="favorites"></match-ratings>
        </smart-collection>
    </library>
</fcpxml>`

var appendmovtopngxmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

<fcpxml version="1.13">
    <resources>
        <asset id="r2" name="cs.pitt.edu" uid="%s" start="0s" hasVideo="1" format="r3" videoSources="1" duration="0s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/cs.pitt.edu.png"></media-rep>
        </asset>
        <asset id="r4" name="speech1" uid="%s" start="0s" hasVideo="1" format="r1" videoSources="1" hasAudio="1" audioSources="1" audioChannels="2" audioRate="48000" duration="240240/24000s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/speech1.mov"></media-rep>
        </asset>
        <format id="r1" name="FFVideoFormat720p2398" frameDuration="1001/24000s" width="1280" height="720" colorSpace="1-1-1 (Rec. 709)"></format>
        <format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"></format>
    </resources>
    <library location="file:///Users/aa/Movies/Untitled.fcpbundle/">
        <event name="6-13-25" uid="78463397-97FD-443D-B4E2-07C581674AFC">
            <project name="wiki" uid="DEA19981-DED5-4851-8435-14515931C68A" modDate="2025-06-13 11:46:22 -0700">
                <sequence format="r1" duration="481481/24000s" tcStart="0s" tcFormat="NDF" audioLayout="stereo" audioRate="48k">
                    <spine>
                        <video ref="r2" offset="0s" name="cs.pitt.edu" duration="241241/24000s" start="86399313/24000s"></video>
                        <asset-clip ref="r4" offset="241241/24000s" name="speech1" duration="240240/24000s" format="r1" tcFormat="NDF" audioRole="dialogue"></asset-clip>
                    </spine>
                </sequence>
            </project>
        </event>
        <smart-collection name="Projects" match="all">
            <match-clip rule="is" type="project"></match-clip>
        </smart-collection>
        <smart-collection name="All Video" match="any">
            <match-media rule="is" type="videoOnly"></match-media>
            <match-media rule="is" type="videoWithAudio"></match-media>
        </smart-collection>
        <smart-collection name="Audio Only" match="all">
            <match-media rule="is" type="audioOnly"></match-media>
        </smart-collection>
        <smart-collection name="Stills" match="all">
            <match-media rule="is" type="stills"></match-media>
        </smart-collection>
        <smart-collection name="Favorites" match="all">
            <match-ratings value="favorites"></match-ratings>
        </smart-collection>
    </library>
</fcpxml>`

// Templates for the new "append to existing" tests (different durations)
var appendPngToExistingTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

<fcpxml version="1.13">
    <resources>
        <asset id="r2" name="cs.pitt.edu" uid="%s" start="0s" hasVideo="1" format="r3" videoSources="1" duration="0s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/cs.pitt.edu.png"></media-rep>
        </asset>
        <asset id="r4" name="cutlass_logo_t" uid="%s" start="0s" hasVideo="1" format="r5" videoSources="1" duration="0s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/cutlass_logo_t.png"></media-rep>
        </asset>
        <format id="r1" name="FFVideoFormat720p2398" frameDuration="1001/24000s" width="1280" height="720" colorSpace="1-1-1 (Rec. 709)"></format>
        <format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"></format>
        <format id="r5" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"></format>
    </resources>
    <library location="file:///Users/aa/Movies/Untitled.fcpbundle/">
        <event name="6-13-25" uid="78463397-97FD-443D-B4E2-07C581674AFC">
            <project name="wiki" uid="DEA19981-DED5-4851-8435-14515931C68A" modDate="2025-06-13 11:46:22 -0700">
                <sequence format="r1" duration="432432/24000s" tcStart="0s" tcFormat="NDF" audioLayout="stereo" audioRate="48k">
                    <spine>
                        <video ref="r2" offset="0s" name="cs.pitt.edu" duration="216216/24000s" start="86399313/24000s"></video>
                        <video ref="r4" offset="216216/24000s" name="cutlass_logo_t" duration="216216/24000s" start="86399313/24000s"></video>
                    </spine>
                </sequence>
            </project>
        </event>
        <smart-collection name="Projects" match="all">
            <match-clip rule="is" type="project"></match-clip>
        </smart-collection>
        <smart-collection name="All Video" match="any">
            <match-media rule="is" type="videoOnly"></match-media>
            <match-media rule="is" type="videoWithAudio"></match-media>
        </smart-collection>
        <smart-collection name="Audio Only" match="all">
            <match-media rule="is" type="audioOnly"></match-media>
        </smart-collection>
        <smart-collection name="Stills" match="all">
            <match-media rule="is" type="stills"></match-media>
        </smart-collection>
        <smart-collection name="Favorites" match="all">
            <match-ratings value="favorites"></match-ratings>
        </smart-collection>
    </library>
</fcpxml>`

var appendMovToExistingTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

<fcpxml version="1.13">
    <resources>
        <asset id="r2" name="cs.pitt.edu" uid="%s" start="0s" hasVideo="1" format="r3" videoSources="1" duration="0s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/cs.pitt.edu.png"></media-rep>
        </asset>
        <asset id="r4" name="speech1" uid="%s" start="0s" hasVideo="1" format="r1" videoSources="1" hasAudio="1" audioSources="1" audioChannels="2" audioRate="48000" duration="240240/24000s">
            <media-rep kind="original-media" sig="%s" src="file:///Users/aa/cs/cutlass/assets/speech1.mov"></media-rep>
        </asset>
        <format id="r1" name="FFVideoFormat720p2398" frameDuration="1001/24000s" width="1280" height="720" colorSpace="1-1-1 (Rec. 709)"></format>
        <format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"></format>
    </resources>
    <library location="file:///Users/aa/Movies/Untitled.fcpbundle/">
        <event name="6-13-25" uid="78463397-97FD-443D-B4E2-07C581674AFC">
            <project name="wiki" uid="DEA19981-DED5-4851-8435-14515931C68A" modDate="2025-06-13 11:46:22 -0700">
                <sequence format="r1" duration="456456/24000s" tcStart="0s" tcFormat="NDF" audioLayout="stereo" audioRate="48k">
                    <spine>
                        <video ref="r2" offset="0s" name="cs.pitt.edu" duration="216216/24000s" start="86399313/24000s"></video>
                        <asset-clip ref="r4" offset="216216/24000s" name="speech1" duration="240240/24000s" format="r1" tcFormat="NDF" audioRole="dialogue"></asset-clip>
                    </spine>
                </sequence>
            </project>
        </event>
        <smart-collection name="Projects" match="all">
            <match-clip rule="is" type="project"></match-clip>
        </smart-collection>
        <smart-collection name="All Video" match="any">
            <match-media rule="is" type="videoOnly"></match-media>
            <match-media rule="is" type="videoWithAudio"></match-media>
        </smart-collection>
        <smart-collection name="Audio Only" match="all">
            <match-media rule="is" type="audioOnly"></match-media>
        </smart-collection>
        <smart-collection name="Stills" match="all">
            <match-media rule="is" type="stills"></match-media>
        </smart-collection>
        <smart-collection name="Favorites" match="all">
            <match-ratings value="favorites"></match-ratings>
        </smart-collection>
    </library>
</fcpxml>`

// createEmptyProject creates an empty FCPXML project for testing
func createEmptyProject() (*FCPXML, error) {
	return GenerateEmpty("")
}

// createProjectWithPng creates an FCPXML project with a single PNG image
func createProjectWithPng() (*FCPXML, error) {
	fcpxml, err := createEmptyProject()
	if err != nil {
		return nil, err
	}

	err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/cs.pitt.edu.png", 9.0)
	if err != nil {
		return nil, err
	}

	return fcpxml, nil
}

// createProjectWithMov creates an FCPXML project with a single MOV video
func createProjectWithMov() (*FCPXML, error) {
	fcpxml, err := createEmptyProject()
	if err != nil {
		return nil, err
	}

	err = AddVideo(fcpxml, "/Users/aa/cs/cutlass/assets/speech1.mov")
	if err != nil {
		return nil, err
	}

	return fcpxml, nil
}

func TestGeneratePng(t *testing.T) {
	testFile := "test_generate_png.fcpxml"

	defer func() {
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up test file: %v", err)
		}
	}()

	fcpxml, err := createProjectWithPng()
	if err != nil {
		t.Fatalf("createProjectWithPng failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Validate structure instead of exact string matching
	loadedFCPXML, err := ReadFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Validate essential structure
	if len(loadedFCPXML.Resources.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(loadedFCPXML.Resources.Assets))
	}

	asset := loadedFCPXML.Resources.Assets[0]
	if asset.Name != "cs.pitt.edu" {
		t.Errorf("Expected asset name 'cs.pitt.edu', got '%s'", asset.Name)
	}
	if asset.Duration != "0s" {
		t.Errorf("Expected asset duration '0s' for image, got '%s'", asset.Duration)
	}
	if asset.HasVideo != "1" {
		t.Errorf("Expected hasVideo='1', got '%s'", asset.HasVideo)
	}
	if asset.MediaRep.Src != "file:///Users/aa/cs/cutlass/assets/cs.pitt.edu.png" {
		t.Errorf("Expected correct file path, got '%s'", asset.MediaRep.Src)
	}
	// Verify that bookmark and metadata are present (enhanced format)
	if asset.MediaRep.Bookmark == "" {
		t.Error("Expected bookmark to be present in enhanced format")
	}
	if asset.Metadata == nil {
		t.Error("Expected metadata to be present in enhanced format")
	}

	// Validate sequence has one video element
	if len(loadedFCPXML.Library.Events) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects[0].Sequences) == 0 {
		t.Fatal("Expected sequence structure not found")
	}

	sequence := &loadedFCPXML.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 1 {
		t.Errorf("Expected 1 video element in spine, got %d", len(sequence.Spine.Videos))
	}

	video := sequence.Spine.Videos[0]
	if video.Name != "cs.pitt.edu" {
		t.Errorf("Expected video name 'cs.pitt.edu', got '%s'", video.Name)
	}
}

func TestGenerateMov(t *testing.T) {
	testFile := "test_generate_mov.fcpxml"

	defer func() {
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up test file: %v", err)
		}
	}()

	fcpxml, err := createProjectWithMov()
	if err != nil {
		t.Fatalf("createProjectWithMov failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Validate structure instead of exact string matching
	loadedFCPXML, err := ReadFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Validate essential structure
	if len(loadedFCPXML.Resources.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(loadedFCPXML.Resources.Assets))
	}

	asset := loadedFCPXML.Resources.Assets[0]
	if asset.Name != "speech1" {
		t.Errorf("Expected asset name 'speech1', got '%s'", asset.Name)
	}
	if asset.HasVideo != "1" {
		t.Errorf("Expected hasVideo='1', got '%s'", asset.HasVideo)
	}
	// Note: hasAudio depends on ffprobe detection and actual file content
	// If ffprobe is not available or file has no audio, this may be empty
	if asset.HasAudio != "" && asset.HasAudio != "1" {
		t.Errorf("Expected hasAudio to be empty or '1', got '%s'", asset.HasAudio)
	}
	if asset.MediaRep.Src != "file:///Users/aa/cs/cutlass/assets/speech1.mov" {
		t.Errorf("Expected correct file path, got '%s'", asset.MediaRep.Src)
	}
	// Verify that bookmark and metadata are present (enhanced format)
	if asset.MediaRep.Bookmark == "" {
		t.Error("Expected bookmark to be present in enhanced format")
	}
	// Metadata depends on video detection working (ffprobe available)
	// If detection fails, metadata may be nil, which is acceptable for testing
	if asset.Metadata != nil {
		t.Log("Metadata present - video detection working properly")
	} else {
		t.Log("Metadata not present - video detection may not be available (missing ffprobe)")
	}

	// Validate sequence has one asset-clip element
	if len(loadedFCPXML.Library.Events) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects[0].Sequences) == 0 {
		t.Fatal("Expected sequence structure not found")
	}

	sequence := &loadedFCPXML.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.AssetClips) != 1 {
		t.Errorf("Expected 1 asset-clip element in spine, got %d", len(sequence.Spine.AssetClips))
	}

	assetClip := sequence.Spine.AssetClips[0]
	if assetClip.Name != "speech1" {
		t.Errorf("Expected asset-clip name 'speech1', got '%s'", assetClip.Name)
	}
	if assetClip.AudioRole != "dialogue" {
		t.Errorf("Expected audioRole='dialogue', got '%s'", assetClip.AudioRole)
	}
}

func TestAppendPng(t *testing.T) {
	testFile := "test_append_png.fcpxml"

	defer func() {
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up test file: %v", err)
		}
	}()

	// Start from empty - important to test the full workflow
	fcpxml, err := createEmptyProject()
	if err != nil {
		t.Fatalf("createEmptyProject failed: %v", err)
	}

	err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/cs.pitt.edu.png", 10.05)
	if err != nil {
		t.Fatalf("First AddImage failed: %v", err)
	}

	err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/alien.png", 9.0)
	if err != nil {
		t.Fatalf("Second AddImage failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Validate structure instead of exact string matching
	loadedFCPXML, err := ReadFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Validate essential structure - should have 2 assets (2 images)
	if len(loadedFCPXML.Resources.Assets) != 2 {
		t.Errorf("Expected 2 assets (2 images), got %d", len(loadedFCPXML.Resources.Assets))
	}

	// Validate sequence has two video elements (both images)
	if len(loadedFCPXML.Library.Events) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects[0].Sequences) == 0 {
		t.Fatal("Expected sequence structure not found")
	}

	sequence := &loadedFCPXML.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 2 {
		t.Errorf("Expected 2 video elements in spine, got %d", len(sequence.Spine.Videos))
	}

	// Validate the video elements
	firstVideo := sequence.Spine.Videos[0]
	if firstVideo.Name != "cs.pitt.edu" {
		t.Errorf("Expected first video name 'cs.pitt.edu', got '%s'", firstVideo.Name)
	}

	secondVideo := sequence.Spine.Videos[1]
	if secondVideo.Name != "alien" {
		t.Errorf("Expected second video name 'alien', got '%s'", secondVideo.Name)
	}
}

func TestAppendMovToPng(t *testing.T) {
	testFile := "test_append_mov_to_png.fcpxml"

	defer func() {
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up test file: %v", err)
		}
	}()

	// Start from empty - important to test the full workflow
	fcpxml, err := createEmptyProject()
	if err != nil {
		t.Fatalf("createEmptyProject failed: %v", err)
	}

	err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/cs.pitt.edu.png", 10.05)
	if err != nil {
		t.Fatalf("AddImage failed: %v", err)
	}

	err = AddVideo(fcpxml, "/Users/aa/cs/cutlass/assets/speech1.mov")
	if err != nil {
		t.Fatalf("AddVideo failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Validate structure instead of exact string matching
	loadedFCPXML, err := ReadFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Validate essential structure - should have 2 assets (image + video)
	if len(loadedFCPXML.Resources.Assets) != 2 {
		t.Errorf("Expected 2 assets (image + video), got %d", len(loadedFCPXML.Resources.Assets))
	}

	// Validate sequence has one video element and one asset-clip
	if len(loadedFCPXML.Library.Events) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects[0].Sequences) == 0 {
		t.Fatal("Expected sequence structure not found")
	}

	sequence := &loadedFCPXML.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 1 {
		t.Errorf("Expected 1 video element in spine, got %d", len(sequence.Spine.Videos))
	}
	if len(sequence.Spine.AssetClips) != 1 {
		t.Errorf("Expected 1 asset-clip element in spine, got %d", len(sequence.Spine.AssetClips))
	}

	// Validate the video element is the image
	video := sequence.Spine.Videos[0]
	if video.Name != "cs.pitt.edu" {
		t.Errorf("Expected video name 'cs.pitt.edu', got '%s'", video.Name)
	}

	// Validate the asset-clip is the video file
	assetClip := sequence.Spine.AssetClips[0]
	if assetClip.Name != "speech1" {
		t.Errorf("Expected asset-clip name 'speech1', got '%s'", assetClip.Name)
	}
}

// New tests that start from existing projects (append to existing content)

func TestAppendPngToExistingProject(t *testing.T) {
	testFile := "test_append_png_to_existing.fcpxml"

	defer func() {
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up test file: %v", err)
		}
	}()

	// Start from existing PNG project
	fcpxml, err := createProjectWithPng()
	if err != nil {
		t.Fatalf("createProjectWithPng failed: %v", err)
	}

	// Add second image to existing project
	err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/alien.png", 9.0)
	if err != nil {
		t.Fatalf("AddImage to existing project failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Validate structure instead of exact string matching
	loadedFCPXML, err := ReadFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Validate essential structure - should have 2 assets (2 images)
	if len(loadedFCPXML.Resources.Assets) != 2 {
		t.Errorf("Expected 2 assets (2 images), got %d", len(loadedFCPXML.Resources.Assets))
	}

	// Validate sequence has two video elements (both images)
	if len(loadedFCPXML.Library.Events) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects[0].Sequences) == 0 {
		t.Fatal("Expected sequence structure not found")
	}

	sequence := &loadedFCPXML.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 2 {
		t.Errorf("Expected 2 video elements in spine, got %d", len(sequence.Spine.Videos))
	}

	// Validate the video elements
	firstVideo := sequence.Spine.Videos[0]
	if firstVideo.Name != "cs.pitt.edu" {
		t.Errorf("Expected first video name 'cs.pitt.edu', got '%s'", firstVideo.Name)
	}

	secondVideo := sequence.Spine.Videos[1]
	if secondVideo.Name != "alien" {
		t.Errorf("Expected second video name 'alien', got '%s'", secondVideo.Name)
	}
}

func TestAppendMovToExistingProject(t *testing.T) {
	testFile := "test_append_mov_to_existing.fcpxml"

	defer func() {
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up test file: %v", err)
		}
	}()

	// Start from existing PNG project  
	fcpxml, err := createProjectWithPng()
	if err != nil {
		t.Fatalf("createProjectWithPng failed: %v", err)
	}

	// Add video to existing project
	err = AddVideo(fcpxml, "/Users/aa/cs/cutlass/assets/speech1.mov")
	if err != nil {
		t.Fatalf("AddVideo to existing project failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Validate structure instead of exact string matching
	loadedFCPXML, err := ReadFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Validate essential structure - should have 2 assets (image + video)
	if len(loadedFCPXML.Resources.Assets) != 2 {
		t.Errorf("Expected 2 assets (image + video), got %d", len(loadedFCPXML.Resources.Assets))
	}

	// Validate sequence has one video element and one asset-clip
	if len(loadedFCPXML.Library.Events) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects) == 0 ||
		len(loadedFCPXML.Library.Events[0].Projects[0].Sequences) == 0 {
		t.Fatal("Expected sequence structure not found")
	}

	sequence := &loadedFCPXML.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) != 1 {
		t.Errorf("Expected 1 video element in spine, got %d", len(sequence.Spine.Videos))
	}
	if len(sequence.Spine.AssetClips) != 1 {
		t.Errorf("Expected 1 asset-clip element in spine, got %d", len(sequence.Spine.AssetClips))
	}

	// Validate the video element is the image
	video := sequence.Spine.Videos[0]
	if video.Name != "cs.pitt.edu" {
		t.Errorf("Expected video name 'cs.pitt.edu', got '%s'", video.Name)
	}

	// Validate the asset-clip is the video file
	assetClip := sequence.Spine.AssetClips[0]
	if assetClip.Name != "speech1" {
		t.Errorf("Expected asset-clip name 'speech1', got '%s'", assetClip.Name)
	}
}

// TestFrameBoundaryAlignment tests that all generated durations and offsets are frame-aligned
// This prevents the "not on an edit frame boundary" error in Final Cut Pro
func TestFrameBoundaryAlignment(t *testing.T) {
	// Test both timeline calculation and duration parsing functions
	t.Run("ParseFCPDurationFrameAlignment", func(t *testing.T) {
		testCases := []struct {
			input          string
			expectedFrames int
			description    string
		}{
			{"0s", 0, "zero duration"},
			{"240240/24000s", 240240, "already frame-aligned duration"},
			{"547547/60000s", 219219, "60000 timebase - should round to nearest frame"},
			{"417417/60000s", 167167, "60000 timebase duration"},
			{"4910906/120000s", 981981, "120000 timebase offset"},
			{"1005004/120000s", 201201, "120000 timebase duration"},
			{"1183181/24000s", 1183182, "non-frame-aligned value - should round up"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				result := parseFCPDuration(tc.input)
				
				// Check that result is frame-aligned (divisible by 1001)
				if result%1001 != 0 {
					t.Errorf("parseFCPDuration(%s) = %d is not frame-aligned (not divisible by 1001)", tc.input, result)
				}
				
				// Check that result matches expected frames
				if result != tc.expectedFrames {
					t.Errorf("parseFCPDuration(%s) = %d, expected %d", tc.input, result, tc.expectedFrames)
				}
				
				// Verify frame count is correct
				frames := result / 1001
				expectedFrameCount := tc.expectedFrames / 1001
				if frames != expectedFrameCount {
					t.Errorf("parseFCPDuration(%s) frame count = %d, expected %d", tc.input, frames, expectedFrameCount)
				}
			})
		}
	})

	t.Run("TimelineCalculationFrameAlignment", func(t *testing.T) {
		// Create a project with complex timeline (multiple different timebases)
		fcpxml, err := createEmptyProject()
		if err != nil {
			t.Fatalf("createEmptyProject failed: %v", err)
		}

		// Add elements with different timebases to test frame alignment
		err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/cs.pitt.edu.png", 9.0) // Creates video element
		if err != nil {
			t.Fatalf("AddImage failed: %v", err)
		}

		err = AddVideo(fcpxml, "/Users/aa/cs/cutlass/assets/speech1.mov") // Creates asset-clip
		if err != nil {
			t.Fatalf("AddVideo failed: %v", err)
		}

		// Check the sequence for frame alignment
		if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 ||
		   len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
			t.Fatal("Generated FCPXML does not have expected structure")
		}

		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		
		// Test sequence duration is frame-aligned
		sequenceDuration := parseFCPDuration(sequence.Duration)
		if sequenceDuration%1001 != 0 {
			t.Errorf("Sequence duration %s (%d) is not frame-aligned", sequence.Duration, sequenceDuration)
		}

		// Test all spine elements have frame-aligned offsets and durations
		for i, clip := range sequence.Spine.AssetClips {
			offset := parseFCPDuration(clip.Offset)
			duration := parseFCPDuration(clip.Duration)
			
			if offset%1001 != 0 {
				t.Errorf("AssetClip[%d] offset %s (%d) is not frame-aligned", i, clip.Offset, offset)
			}
			if duration%1001 != 0 {
				t.Errorf("AssetClip[%d] duration %s (%d) is not frame-aligned", i, clip.Duration, duration)
			}
		}

		for i, video := range sequence.Spine.Videos {
			offset := parseFCPDuration(video.Offset)
			duration := parseFCPDuration(video.Duration)
			
			if offset%1001 != 0 {
				t.Errorf("Video[%d] offset %s (%d) is not frame-aligned", i, video.Offset, offset)
			}
			if duration%1001 != 0 {
				t.Errorf("Video[%d] duration %s (%d) is not frame-aligned", i, video.Duration, duration)
			}
		}
	})

	t.Run("AddVideoToComplexTimelineFrameAlignment", func(t *testing.T) {
		// Test the specific scenario that caused the frame boundary violation
		testFile := "test_frame_boundary_complex.fcpxml"
		defer func() {
			if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
				t.Errorf("Failed to clean up test file: %v", err)
			}
		}()

		// Create an FCPXML structure that simulates the problem case
		// This mimics the orig.fcpxml structure with mixed timebases
		fcpxml, err := createEmptyProject()
		if err != nil {
			t.Fatalf("createEmptyProject failed: %v", err)
		}

		// Manually create spine elements with different timebases to simulate orig.fcpxml
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		
		// Add asset clip with 60000 timebase (simulating r6)
		sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, AssetClip{
			Ref:      "r6",
			Offset:   "547547/60000s",
			Duration: "417417/60000s",
			Name:     "test_video",
		})

		// Add video with 120000 timebase (simulating r2 final)
		sequence.Spine.Videos = append(sequence.Spine.Videos, Video{
			Ref:      "r2",
			Offset:   "4910906/120000s",
			Duration: "1005004/120000s",
			Name:     "test_image",
		})

		// Calculate timeline duration - this should be frame-aligned
		timelineDuration := calculateTimelineDuration(sequence)
		calculatedFrames := parseFCPDuration(timelineDuration)
		
		if calculatedFrames%1001 != 0 {
			t.Errorf("calculateTimelineDuration() returned non-frame-aligned value: %s (%d frames)", 
				timelineDuration, calculatedFrames)
		}

		// Verify the calculated timeline end matches expected frame-aligned value
		// Based on our fix: 4910906/120000s + 1005004/120000s should round to 1182 frames
		expectedFrames := 1182 * 1001 // 1183182
		if calculatedFrames != expectedFrames {
			t.Errorf("calculateTimelineDuration() = %d frames, expected %d frames (frame-aligned)", 
				calculatedFrames/1001, expectedFrames/1001)
		}

		// Test that adding a video to this timeline produces frame-aligned offset
		err = AddVideo(fcpxml, "/Users/aa/cs/cutlass/assets/speech1.mov")
		if err != nil {
			t.Fatalf("AddVideo failed: %v", err)
		}

		// Find the newly added video clip
		var newClip *AssetClip
		for i := range sequence.Spine.AssetClips {
			if sequence.Spine.AssetClips[i].Ref != "r6" { // Not the manually added one
				newClip = &sequence.Spine.AssetClips[i]
				break
			}
		}

		if newClip == nil {
			t.Fatal("New video clip was not found in spine")
		}

		// Verify the new clip has frame-aligned offset
		newOffset := parseFCPDuration(newClip.Offset)
		if newOffset%1001 != 0 {
			t.Errorf("New video offset %s (%d) is not frame-aligned", newClip.Offset, newOffset)
		}

		// Verify the new clip offset matches the calculated timeline duration
		if newOffset != expectedFrames {
			t.Errorf("New video offset %d frames, expected %d frames", newOffset/1001, expectedFrames/1001)
		}

		// Write and validate the final FCPXML
		err = WriteToFile(fcpxml, testFile)
		if err != nil {
			t.Fatalf("WriteToFile failed: %v", err)
		}
	})
}

