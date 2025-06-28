package fcp

import (
	"strings"
	"testing"
	"os"
	"io"
)

func TestCompatibilityMode(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	// Test setting and getting compatibility mode
	SetCompatibilityMode(CompatibilityStrict)
	if GetCompatibilityMode() != CompatibilityStrict {
		t.Error("Failed to set strict compatibility mode")
	}

	SetCompatibilityMode(CompatibilityWarn)
	if GetCompatibilityMode() != CompatibilityWarn {
		t.Error("Failed to set warning compatibility mode")
	}

	SetCompatibilityMode(CompatibilityLenient)
	if GetCompatibilityMode() != CompatibilityLenient {
		t.Error("Failed to set lenient compatibility mode")
	}
}

func TestCreateLegacyFCPXML(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	// Test in warning mode (should work with warnings)
	SetCompatibilityMode(CompatibilityWarn)
	fcpxml, err := CreateLegacyFCPXML("1.13")
	if err != nil {
		t.Errorf("CreateLegacyFCPXML failed in warning mode: %v", err)
	}
	if fcpxml == nil {
		t.Error("Expected FCPXML but got nil")
	}
	if fcpxml.Version != "1.13" {
		t.Errorf("Expected version 1.13, got %s", fcpxml.Version)
	}

	// Test in strict mode (should fail)
	SetCompatibilityMode(CompatibilityStrict)
	_, err = CreateLegacyFCPXML("1.13")
	if err == nil {
		t.Error("Expected error in strict mode but got none")
	}

	// Test in lenient mode (should work silently)
	SetCompatibilityMode(CompatibilityLenient)
	fcpxml, err = CreateLegacyFCPXML("1.13")
	if err != nil {
		t.Errorf("CreateLegacyFCPXML failed in lenient mode: %v", err)
	}
	if fcpxml == nil {
		t.Error("Expected FCPXML but got nil")
	}
}

func TestAddLegacyAsset(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	fcpxml, _ := GenerateEmpty("1.13")

	// Test in warning mode
	SetCompatibilityMode(CompatibilityWarn)
	err := AddLegacyAsset(fcpxml, "r1", "test-asset", "test-uid", "240240/24000s", "file:///test.mp4")
	if err != nil {
		t.Errorf("AddLegacyAsset failed in warning mode: %v", err)
	}

	// Check that asset was added
	if len(fcpxml.Resources.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(fcpxml.Resources.Assets))
	}

	asset := fcpxml.Resources.Assets[0]
	if string(asset.ID) != "r1" {
		t.Errorf("Expected asset ID 'r1', got '%s'", asset.ID)
	}
	if asset.Name != "test-asset" {
		t.Errorf("Expected asset name 'test-asset', got '%s'", asset.Name)
	}

	// Test in strict mode
	SetCompatibilityMode(CompatibilityStrict)
	err = AddLegacyAsset(fcpxml, "r2", "test-asset-2", "test-uid-2", "240240/24000s", "file:///test2.mp4")
	if err == nil {
		t.Error("Expected error in strict mode but got none")
	}

	// Test with invalid duration (should fail validation)
	SetCompatibilityMode(CompatibilityWarn)
	err = AddLegacyAsset(fcpxml, "r3", "test-asset-3", "test-uid-3", "invalid-duration", "file:///test3.mp4")
	if err == nil {
		t.Error("Expected validation error for invalid duration")
	}
}

func TestAddLegacyFormat(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	fcpxml, _ := GenerateEmpty("1.13")

	// Test in warning mode
	SetCompatibilityMode(CompatibilityWarn)
	err := AddLegacyFormat(fcpxml, "r2", "FFVideoFormat1080p30", "1920", "1080")
	if err != nil {
		t.Errorf("AddLegacyFormat failed in warning mode: %v", err)
	}

	// Check that format was added (GenerateEmpty creates 1 default format, AddLegacyFormat adds another)
	if len(fcpxml.Resources.Formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(fcpxml.Resources.Formats))
	}

	// Find the newly added format (should be the second one with ID "r2")
	var addedFormat *Format
	for i := range fcpxml.Resources.Formats {
		if string(fcpxml.Resources.Formats[i].ID) == "r2" {
			addedFormat = &fcpxml.Resources.Formats[i]
			break
		}
	}
	if addedFormat == nil {
		t.Error("Could not find added format with ID 'r2'")
		return
	}
	
	if addedFormat.Width != "1920" {
		t.Errorf("Expected format width '1920', got '%s'", addedFormat.Width)
	}
	if addedFormat.Height != "1080" {
		t.Errorf("Expected format height '1080', got '%s'", addedFormat.Height)
	}

	// Test in strict mode
	SetCompatibilityMode(CompatibilityStrict)
	err = AddLegacyFormat(fcpxml, "r3", "FFVideoFormat720p30", "1280", "720")
	if err == nil {
		t.Error("Expected error in strict mode but got none")
	}

	// Test with invalid width (should fail validation)
	SetCompatibilityMode(CompatibilityWarn)
	err = AddLegacyFormat(fcpxml, "r4", "InvalidFormat", "not-a-number", "720")
	if err == nil {
		t.Error("Expected validation error for invalid width")
	}
}

func TestCreateLegacyAssetClip(t *testing.T) {
	clip := CreateLegacyAssetClip("r1", "0s", "240240/24000s", "Test Clip")

	if clip.Ref != "r1" {
		t.Errorf("Expected ref 'r1', got '%s'", clip.Ref)
	}

	if clip.Offset != "0s" {
		t.Errorf("Expected offset '0s', got '%s'", clip.Offset)
	}

	if string(clip.Duration) != "240240/24000s" {
		t.Errorf("Expected duration '240240/24000s', got '%s'", clip.Duration)
	}

	if clip.Name != "Test Clip" {
		t.Errorf("Expected name 'Test Clip', got '%s'", clip.Name)
	}
}

func TestAddLegacyAssetClipToSpine(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	spine := &Spine{}
	clip := CreateLegacyAssetClip("r1", "0s", "240240/24000s", "Test Clip")

	// Test in warning mode
	SetCompatibilityMode(CompatibilityWarn)
	err := AddLegacyAssetClipToSpine(spine, clip)
	if err != nil {
		t.Errorf("AddLegacyAssetClipToSpine failed: %v", err)
	}

	if len(spine.AssetClips) != 1 {
		t.Errorf("Expected 1 asset clip, got %d", len(spine.AssetClips))
	}

	// Test in strict mode
	SetCompatibilityMode(CompatibilityStrict)
	clip2 := CreateLegacyAssetClip("r2", "120120/24000s", "240240/24000s", "Test Clip 2")
	err = AddLegacyAssetClipToSpine(spine, clip2)
	if err == nil {
		t.Error("Expected error in strict mode but got none")
	}

	// Test with invalid duration
	SetCompatibilityMode(CompatibilityWarn)
	invalidClip := AssetClip{
		Ref:      "r3",
		Offset:   "0s",
		Duration: "invalid-duration",
		Name:     "Invalid Clip",
	}
	err = AddLegacyAssetClipToSpine(spine, invalidClip)
	if err == nil {
		t.Error("Expected validation error for invalid duration")
	}
}

func TestMarshalLegacyXML(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	fcpxml, _ := GenerateEmpty("1.13")

	// Test in warning mode
	SetCompatibilityMode(CompatibilityWarn)
	data, err := MarshalLegacyXML(fcpxml)
	if err != nil {
		t.Errorf("MarshalLegacyXML failed in warning mode: %v", err)
	}
	if data == nil {
		t.Error("Expected XML data but got nil")
	}

	// Verify it's valid XML
	if !strings.Contains(string(data), "<fcpxml") {
		t.Error("Generated XML should contain fcpxml element")
	}

	// Test in strict mode
	SetCompatibilityMode(CompatibilityStrict)
	_, err = MarshalLegacyXML(fcpxml)
	if err == nil {
		t.Error("Expected error in strict mode but got none")
	}
}

func TestLegacyResourceManager(t *testing.T) {
	fcpxml, _ := GenerateEmpty("1.13")
	lrm := NewLegacyResourceManager(fcpxml)

	// Test ID generation
	id1 := lrm.GenerateID()
	if id1 != "r1" {
		t.Errorf("Expected first ID to be 'r1', got '%s'", id1)
	}

	id2 := lrm.GenerateID()
	if id2 != "r2" {
		t.Errorf("Expected second ID to be 'r2', got '%s'", id2)
	}

	// Test adding asset
	assetID := lrm.AddAsset("test-asset", "test-uid", "240240/24000s", "file:///test.mp4")
	if assetID != "r3" {
		t.Errorf("Expected asset ID 'r3', got '%s'", assetID)
	}

	// Check that asset was added to FCPXML
	if len(fcpxml.Resources.Assets) != 1 {
		t.Errorf("Expected 1 asset in FCPXML, got %d", len(fcpxml.Resources.Assets))
	}
}

func TestValidateLegacyStructure(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	fcpxml, _ := GenerateEmpty("1.13")

	// Test in warning mode
	SetCompatibilityMode(CompatibilityWarn)
	err := ValidateLegacyStructure(fcpxml)
	if err != nil {
		t.Errorf("ValidateLegacyStructure failed in warning mode: %v", err)
	}

	// Test in strict mode
	SetCompatibilityMode(CompatibilityStrict)
	err = ValidateLegacyStructure(fcpxml)
	if err == nil {
		t.Error("Expected error in strict mode but got none")
	}
}

func TestFixLegacyIssues(t *testing.T) {
	// Create FCPXML with legacy issues
	fcpxml := &FCPXML{
		// Missing version
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r1",
					Name:     "test-asset",
					UID:      "test-uid",
					Duration: "240240/24000", // Missing 's' suffix
					Start:    "0s",
				},
			},
		},
		Library: Library{
			Events: []Event{
				{
					Name: "Test Event",
					Projects: []Project{
						{
							Name: "Test Project",
							Sequences: []Sequence{
								{
									Duration: "240240/24000s",
									TCStart:  "0s",
									Spine: Spine{
										AssetClips: []AssetClip{
											{
												Ref:      "r1",
												Offset:   "0s",
												Duration: "240240/24000", // Missing 's' suffix
												Name:     "Test Clip",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	warnings := FixLegacyIssues(fcpxml)

	// Check that version was fixed
	if fcpxml.Version != "1.13" {
		t.Errorf("Expected version to be fixed to '1.13', got '%s'", fcpxml.Version)
	}

	// Check that warnings were generated
	if len(warnings) == 0 {
		t.Error("Expected warnings but got none")
	}

	// Verify specific warnings
	foundVersionWarning := false
	foundDurationWarning := false
	foundClipWarning := false

	for _, warning := range warnings {
		if strings.Contains(warning, "version") {
			foundVersionWarning = true
		}
		if strings.Contains(warning, "duration format for asset") {
			foundDurationWarning = true
		}
		if strings.Contains(warning, "duration format for asset clip") {
			foundClipWarning = true
		}
	}

	if !foundVersionWarning {
		t.Error("Expected warning about missing version")
	}
	if !foundDurationWarning {
		t.Error("Expected warning about asset duration format")
	}
	if !foundClipWarning {
		t.Error("Expected warning about asset clip duration format")
	}
}

func TestIsLegacyPattern(t *testing.T) {
	// Test FCPXML with legacy patterns
	legacyFCPXML := &FCPXML{
		// Missing version (legacy pattern)
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r1",
					Name:     "test-asset",
					Duration: "240240/24000", // Missing 's' suffix
				},
			},
		},
	}

	if !IsLegacyPattern(legacyFCPXML) {
		t.Error("Expected legacy pattern to be detected")
	}

	// Test modern FCPXML
	modernFCPXML := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r1",
					Name:     "test-asset",
					Duration: "240240/24000s", // Correct format
				},
			},
		},
	}

	if IsLegacyPattern(modernFCPXML) {
		t.Error("Modern pattern incorrectly detected as legacy")
	}
}

func TestSuggestMigration(t *testing.T) {
	// Test FCPXML that needs migration
	legacyFCPXML := &FCPXML{
		// Missing version triggers legacy pattern detection
	}

	suggestions := SuggestMigration(legacyFCPXML)

	if len(suggestions) == 0 {
		t.Error("Expected migration suggestions but got none")
	}

	// Check for expected suggestion content
	foundLegacyMention := false
	foundMigrationManagerMention := false
	foundTransactionMention := false

	for _, suggestion := range suggestions {
		if strings.Contains(suggestion, "legacy patterns") {
			foundLegacyMention = true
		}
		if strings.Contains(suggestion, "MigrationManager") {
			foundMigrationManagerMention = true
		}
		if strings.Contains(suggestion, "Transaction") {
			foundTransactionMention = true
		}
	}

	if !foundLegacyMention {
		t.Error("Expected mention of legacy patterns")
	}
	if !foundMigrationManagerMention {
		t.Error("Expected mention of MigrationManager")
	}
	if !foundTransactionMention {
		t.Error("Expected mention of Transaction")
	}
}

func TestGetCompatibilityInfo(t *testing.T) {
	// Save original mode
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	SetCompatibilityMode(CompatibilityWarn)

	// Test with legacy FCPXML
	legacyFCPXML := &FCPXML{} // Missing version = legacy

	info := GetCompatibilityInfo(legacyFCPXML)

	if info.Mode != CompatibilityWarn {
		t.Errorf("Expected warning mode, got %v", info.Mode)
	}

	if !info.MigrationNeeded {
		t.Error("Expected migration to be needed for legacy FCPXML")
	}

	if len(info.Suggestions) == 0 {
		t.Error("Expected suggestions for legacy FCPXML")
	}

	// Test with modern FCPXML
	modernFCPXML := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r1",
					Duration: "240240/24000s",
				},
			},
		},
	}

	info = GetCompatibilityInfo(modernFCPXML)

	if info.MigrationNeeded {
		t.Error("Migration should not be needed for modern FCPXML")
	}
}

func TestCompatibilityInfo_String(t *testing.T) {
	info := CompatibilityInfo{
		Mode:            CompatibilityWarn,
		MigrationNeeded: true,
		Suggestions:     []string{"Use MigrationManager", "Use Transaction"},
	}

	infoStr := info.String()

	if !strings.Contains(infoStr, "Compatibility Status") {
		t.Error("Info string should contain status header")
	}

	if !strings.Contains(infoStr, "Warning") {
		t.Error("Info string should mention warning mode")
	}

	if !strings.Contains(infoStr, "RECOMMENDED") {
		t.Error("Info string should recommend migration")
	}

	if !strings.Contains(infoStr, "Use MigrationManager") {
		t.Error("Info string should contain suggestions")
	}
}

func TestDeprecationWarnings(t *testing.T) {
	// Save original mode and stderr
	originalMode := GetCompatibilityMode()
	defer SetCompatibilityMode(originalMode)

	// Redirect stderr to capture warnings
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test warning mode (should output warnings)
	SetCompatibilityMode(CompatibilityWarn)
	_, _ = CreateLegacyFCPXML("1.13")

	// Close writer and restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	if !strings.Contains(outputStr, "DEPRECATION WARNING") {
		t.Error("Expected deprecation warning in warning mode")
	}

	if !strings.Contains(outputStr, "CreateLegacyFCPXML") {
		t.Error("Expected specific function name in warning")
	}

	// Test lenient mode (should not output warnings)
	SetCompatibilityMode(CompatibilityLenient)

	// Redirect stderr again
	r2, w2, _ := os.Pipe()
	os.Stderr = w2

	_, _ = CreateLegacyFCPXML("1.13")

	w2.Close()
	os.Stderr = oldStderr

	output2, _ := io.ReadAll(r2)
	outputStr2 := string(output2)

	if strings.Contains(outputStr2, "DEPRECATION WARNING") {
		t.Error("Should not output warnings in lenient mode")
	}
}