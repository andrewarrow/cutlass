// Package fcp provides tests for FCPXML generation.
//
// 🚨 CRITICAL: Tests MUST validate CLAUDE.md compliance:
// - AFTER changes → RUN: xmllint --dtdvalid FCPXMLv1_13.dtd test_file.fcpxml  
// - BEFORE commits → RUN: ValidateClaudeCompliance() function
// - FOR durations → USE: ConvertSecondsToFCPDuration() function  
// - VERIFY: No fmt.Sprintf() with XML content in any test
package fcp

import (
	"fmt"
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

	generatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Generate expected XML with correct UIDs
	pngUID := GenerateUID("/Users/aa/cs/cutlass/assets/cs.pitt.edu.png")
	expectedXML := fmt.Sprintf(pngxmlTemplate, pngUID, pngUID)

	if string(generatedContent) != expectedXML {
		t.Errorf("Generated XML does not match expected output.\nExpected:\n%s\n\nGenerated:\n%s", expectedXML, string(generatedContent))
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

	generatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Generate expected XML with correct UIDs
	movUID := GenerateUID("/Users/aa/cs/cutlass/assets/speech1.mov")
	expectedXML := fmt.Sprintf(movxmlTemplate, movUID, movUID)

	if string(generatedContent) != expectedXML {
		t.Errorf("Generated XML does not match expected output.\nExpected:\n%s\n\nGenerated:\n%s", expectedXML, string(generatedContent))
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

	err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/cutlass_logo_t.png", 9.0)
	if err != nil {
		t.Fatalf("Second AddImage failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	generatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Generate expected XML with correct UIDs
	pngUID := GenerateUID("/Users/aa/cs/cutlass/assets/cs.pitt.edu.png")
	logoUID := GenerateUID("/Users/aa/cs/cutlass/assets/cutlass_logo_t.png")
	expectedXML := fmt.Sprintf(appendpngxmlTemplate, pngUID, pngUID, logoUID, logoUID)

	if string(generatedContent) != expectedXML {
		t.Errorf("Generated XML does not match expected output.\nExpected:\n%s\n\nGenerated:\n%s", expectedXML, string(generatedContent))
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

	generatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Generate expected XML with correct UIDs
	pngUID := GenerateUID("/Users/aa/cs/cutlass/assets/cs.pitt.edu.png")
	movUID := GenerateUID("/Users/aa/cs/cutlass/assets/speech1.mov")
	expectedXML := fmt.Sprintf(appendmovtopngxmlTemplate, pngUID, pngUID, movUID, movUID)

	if string(generatedContent) != expectedXML {
		t.Errorf("Generated XML does not match expected output.\nExpected:\n%s\n\nGenerated:\n%s", expectedXML, string(generatedContent))
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
	err = AddImage(fcpxml, "/Users/aa/cs/cutlass/assets/cutlass_logo_t.png", 9.0)
	if err != nil {
		t.Fatalf("AddImage to existing project failed: %v", err)
	}

	err = WriteToFile(fcpxml, testFile)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	generatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Generate expected XML with correct UIDs
	pngUID := GenerateUID("/Users/aa/cs/cutlass/assets/cs.pitt.edu.png")
	logoUID := GenerateUID("/Users/aa/cs/cutlass/assets/cutlass_logo_t.png")
	expectedXML := fmt.Sprintf(appendPngToExistingTemplate, pngUID, pngUID, logoUID, logoUID)

	if string(generatedContent) != expectedXML {
		t.Errorf("Generated XML does not match expected output.\nExpected:\n%s\n\nGenerated:\n%s", expectedXML, string(generatedContent))
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

	generatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Generate expected XML with correct UIDs
	pngUID := GenerateUID("/Users/aa/cs/cutlass/assets/cs.pitt.edu.png")
	movUID := GenerateUID("/Users/aa/cs/cutlass/assets/speech1.mov")
	expectedXML := fmt.Sprintf(appendMovToExistingTemplate, pngUID, pngUID, movUID, movUID)

	if string(generatedContent) != expectedXML {
		t.Errorf("Generated XML does not match expected output.\nExpected:\n%s\n\nGenerated:\n%s", expectedXML, string(generatedContent))
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

// TestAddPipVideo tests the picture-in-picture video functionality
func TestAddPipVideo(t *testing.T) {
	// Create a temporary FCPXML file with a base video
	baseFile := "test_base_video.fcpxml"
	pipFile := "test_pip_video.fcpxml"
	
	// Cleanup function
	defer func() {
		for _, file := range []string{baseFile, pipFile} {
			if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
				t.Errorf("Failed to clean up test file %s: %v", file, err)
			}
		}
	}()

	// Step 1: Generate base FCPXML with one video
	t.Run("CreateBaseVideo", func(t *testing.T) {
		// Create a temporary video file for testing
		testMainVideo := "test_main.mov"
		defer os.Remove(testMainVideo)
		
		// Create empty file to simulate video
		if err := os.WriteFile(testMainVideo, []byte("test video content"), 0644); err != nil {
			t.Fatalf("Failed to create test video file: %v", err)
		}

		// Generate empty FCPXML first
		fcpxml, err := GenerateEmpty(baseFile)
		if err != nil {
			t.Fatalf("Failed to generate empty FCPXML: %v", err)
		}

		// Add a video to create base
		err = AddVideo(fcpxml, testMainVideo)
		if err != nil {
			t.Fatalf("Failed to add base video: %v", err)
		}

		// Save base FCPXML
		err = WriteToFile(fcpxml, baseFile)
		if err != nil {
			t.Fatalf("Failed to save base FCPXML: %v", err)
		}
	})

	// Step 2: Load base FCPXML and add PIP video
	t.Run("AddPipVideo", func(t *testing.T) {
		// Create a temporary PIP video file for testing
		testPipVideo := "test_pip.mov"
		defer os.Remove(testPipVideo)
		
		// Create empty file to simulate PIP video
		if err := os.WriteFile(testPipVideo, []byte("test pip video content"), 0644); err != nil {
			t.Fatalf("Failed to create test PIP video file: %v", err)
		}

		// Load existing FCPXML
		fcpxml, err := ReadFromFile(baseFile)
		if err != nil {
			t.Fatalf("Failed to load base FCPXML: %v", err)
		}

		// Add PIP video
		err = AddPipVideo(fcpxml, testPipVideo, 0.0)
		if err != nil {
			t.Fatalf("AddPipVideo failed: %v", err)
		}

		// Save modified FCPXML
		err = WriteToFile(fcpxml, pipFile)
		if err != nil {
			t.Fatalf("Failed to save PIP FCPXML: %v", err)
		}
	})

	// Step 3: Validate structure and requirements
	t.Run("ValidatePipStructure", func(t *testing.T) {
		// Load the PIP FCPXML for validation
		fcpxml, err := ReadFromFile(pipFile)
		if err != nil {
			t.Fatalf("Failed to load PIP FCPXML for validation: %v", err)
		}

		// Validate format requirements
		if len(fcpxml.Resources.Formats) < 3 {
			t.Errorf("Expected at least 3 formats (sequence, main, pip), got %d", len(fcpxml.Resources.Formats))
		}

		// Validate asset requirements
		if len(fcpxml.Resources.Assets) < 2 {
			t.Errorf("Expected at least 2 assets (main, pip), got %d", len(fcpxml.Resources.Assets))
		}

		// Validate Shape Mask effect
		shapeMaskFound := false
		for _, effect := range fcpxml.Resources.Effects {
			if effect.UID == "FFSuperEllipseMask" {
				shapeMaskFound = true
				break
			}
		}
		if !shapeMaskFound {
			t.Error("Shape Mask effect not found in resources")
		}

		// Validate spine structure
		if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
			t.Fatal("No sequence found in generated FCPXML")
		}

		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		if len(sequence.Spine.AssetClips) == 0 {
			t.Fatal("No asset clips found in spine")
		}

		mainClip := sequence.Spine.AssetClips[0]

		// Validate main clip has required PIP elements
		if mainClip.ConformRate == nil {
			t.Error("Main clip missing conform-rate")
		} else if mainClip.ConformRate.ScaleEnabled != "0" {
			t.Errorf("Main clip conform-rate scaleEnabled should be '0', got '%s'", mainClip.ConformRate.ScaleEnabled)
		}

		if mainClip.AdjustCrop == nil {
			t.Error("Main clip missing adjust-crop")
		}

		if mainClip.AdjustTransform == nil {
			t.Error("Main clip missing adjust-transform")
		} else {
			// Verify transform values match PIP pattern
			expectedPos := "60.3234 -35.9353"
			expectedScale := "0.28572 0.28572"
			if mainClip.AdjustTransform.Position != expectedPos {
				t.Errorf("Main clip position should be '%s', got '%s'", expectedPos, mainClip.AdjustTransform.Position)
			}
			if mainClip.AdjustTransform.Scale != expectedScale {
				t.Errorf("Main clip scale should be '%s', got '%s'", expectedScale, mainClip.AdjustTransform.Scale)
			}
		}

		// Validate nested PIP clip
		if len(mainClip.NestedAssetClips) == 0 {
			t.Error("Main clip missing nested PIP asset-clip")
		} else {
			pipClip := mainClip.NestedAssetClips[0]
			if pipClip.Lane != "-1" {
				t.Errorf("PIP clip lane should be '-1', got '%s'", pipClip.Lane)
			}
			if pipClip.ConformRate == nil {
				t.Error("PIP clip missing conform-rate")
			} else if pipClip.ConformRate.SrcFrameRate != "60" {
				t.Errorf("PIP clip srcFrameRate should be '60', got '%s'", pipClip.ConformRate.SrcFrameRate)
			}
		}

		// Validate Shape Mask filter on main clip
		if len(mainClip.FilterVideos) == 0 {
			t.Error("Main clip missing Shape Mask filter")
		} else {
			filter := mainClip.FilterVideos[0]
			if filter.Name != "Shape Mask" {
				t.Errorf("Filter name should be 'Shape Mask', got '%s'", filter.Name)
			}

			// Validate key Shape Mask parameters
			expectedParams := map[string]string{
				"Radius":     "305 190.625",
				"Curvature":  "0.3695",
				"Feather":    "100",
				"Falloff":    "-100",
				"Input Size": "1920 1080",
			}

			paramMap := make(map[string]string)
			for _, param := range filter.Params {
				paramMap[param.Name] = param.Value
			}

			for name, expectedValue := range expectedParams {
				if actualValue, exists := paramMap[name]; !exists {
					t.Errorf("Shape Mask missing parameter '%s'", name)
				} else if actualValue != expectedValue {
					t.Errorf("Shape Mask parameter '%s' should be '%s', got '%s'", name, expectedValue, actualValue)
				}
			}
		}

		// Validate format compatibility (main and pip formats differ from sequence)
		sequenceFormat := sequence.Format
		mainFormat := mainClip.Format
		if len(mainClip.NestedAssetClips) > 0 {
			pipFormat := mainClip.NestedAssetClips[0].Format
			
			if mainFormat == sequenceFormat {
				t.Error("Main clip format should differ from sequence format to enable conform-rate")
			}
			if pipFormat == sequenceFormat {
				t.Error("PIP clip format should differ from sequence format to enable conform-rate")
			}
			if mainFormat == pipFormat {
				t.Error("Main and PIP formats should be different")
			}
		}
	})

	// Step 4: Validate CLAUDE.md compliance
	t.Run("ValidateClaudeCompliance", func(t *testing.T) {
		// Read generated file content
		content, err := os.ReadFile(pipFile)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		contentStr := string(content)

		// Verify proper XML structure exists
		requiredElements := []string{
			"<conform-rate scaleEnabled=\"0\"",
			"<adjust-crop mode=\"trim\"",
			"<adjust-transform position=",
			"lane=\"-1\"",
			"<filter-video",
			"name=\"Shape Mask\"",
		}

		for _, element := range requiredElements {
			found := false
			for i := 0; i <= len(contentStr)-len(element); i++ {
				if contentStr[i:i+len(element)] == element {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Generated XML missing required element: %s", element)
			}
		}
	})
}
