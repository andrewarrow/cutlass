package fcp

import (
	"strings"
	"testing"
)

func TestNewVersionHandler(t *testing.T) {
	// Test supported versions
	supportedVersions := []string{"1.11", "1.12", "1.13"}
	
	for _, version := range supportedVersions {
		vh, err := NewVersionHandler(version)
		if err != nil {
			t.Errorf("Failed to create version handler for supported version %s: %v", version, err)
			continue
		}
		
		if vh == nil {
			t.Errorf("Version handler for %s is nil", version)
			continue
		}
		
		if vh.version.String != version {
			t.Errorf("Expected version %s, got %s", version, vh.version.String)
		}
	}
	
	// Test unsupported version
	_, err := NewVersionHandler("2.0")
	if err == nil {
		t.Error("Expected error for unsupported version 2.0")
	}
	
	_, err = NewVersionHandler("")
	if err == nil {
		t.Error("Expected error for empty version")
	}
}

func TestDetectVersion(t *testing.T) {
	tests := []struct {
		name        string
		xmlData     string
		expectError bool
		expectVersion string
	}{
		{
			name: "Valid FCPXML 1.13",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.13">
	<resources/>
	<library/>
</fcpxml>`,
			expectError: false,
			expectVersion: "1.13",
		},
		{
			name: "Valid FCPXML 1.11",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.11">
	<resources/>
	<library/>
</fcpxml>`,
			expectError: false,
			expectVersion: "1.11",
		},
		{
			name: "Missing version attribute",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml>
	<resources/>
	<library/>
</fcpxml>`,
			expectError: true,
		},
		{
			name: "Unsupported version",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="2.0">
	<resources/>
	<library/>
</fcpxml>`,
			expectError: true,
		},
		{
			name: "Not FCPXML document",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<other version="1.13">
	<data/>
</other>`,
			expectError: true,
		},
		{
			name: "Invalid XML",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.13"
	<unclosed>
</fcpxml>`,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version, err := DetectVersion([]byte(test.xmlData))
			
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !test.expectError && version != test.expectVersion {
				t.Errorf("Expected version %s, got %s", test.expectVersion, version)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	// Test valid versions
	validVersions := []string{"1.11", "1.12", "1.13"}
	for _, version := range validVersions {
		if err := ValidateVersion(version); err != nil {
			t.Errorf("Valid version %s should not produce error: %v", version, err)
		}
	}
	
	// Test invalid versions
	invalidVersions := []string{"", "1.10", "2.0", "1.14", "invalid"}
	for _, version := range invalidVersions {
		if err := ValidateVersion(version); err == nil {
			t.Errorf("Invalid version %s should produce error", version)
		}
	}
}

func TestIsVersionSupported(t *testing.T) {
	// Test supported versions
	if !IsVersionSupported("1.11") {
		t.Error("Version 1.11 should be supported")
	}
	
	if !IsVersionSupported("1.12") {
		t.Error("Version 1.12 should be supported")
	}
	
	if !IsVersionSupported("1.13") {
		t.Error("Version 1.13 should be supported")
	}
	
	// Test unsupported versions
	if IsVersionSupported("2.0") {
		t.Error("Version 2.0 should not be supported")
	}
	
	if IsVersionSupported("") {
		t.Error("Empty version should not be supported")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
		expectError bool
	}{
		{"1.11", "1.12", -1, false}, // v1 < v2
		{"1.12", "1.11", 1, false},  // v1 > v2
		{"1.13", "1.13", 0, false},  // v1 == v2
		{"1.11", "1.13", -1, false}, // v1 < v2
		{"1.13", "1.11", 1, false},  // v1 > v2
		{"2.0", "1.13", 0, true},    // Invalid v1
		{"1.13", "2.0", 0, true},    // Invalid v2
	}

	for _, test := range tests {
		t.Run(test.v1+"_vs_"+test.v2, func(t *testing.T) {
			result, err := CompareVersions(test.v1, test.v2)
			
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !test.expectError && result != test.expected {
				t.Errorf("Expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestCanUpgrade(t *testing.T) {
	tests := []struct {
		source   string
		target   string
		expected bool
		expectError bool
	}{
		{"1.11", "1.12", true, false},  // Can upgrade
		{"1.11", "1.13", true, false},  // Can upgrade
		{"1.12", "1.13", true, false},  // Can upgrade
		{"1.13", "1.13", true, false},  // Same version (allowed)
		{"1.13", "1.12", false, false}, // Cannot upgrade (downgrade)
		{"1.12", "1.11", false, false}, // Cannot upgrade (downgrade)
		{"2.0", "1.13", false, true},   // Invalid source
		{"1.13", "2.0", false, true},   // Invalid target
	}

	for _, test := range tests {
		t.Run(test.source+"_to_"+test.target, func(t *testing.T) {
			canUpgrade, err := CanUpgrade(test.source, test.target)
			
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !test.expectError && canUpgrade != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, canUpgrade)
			}
		})
	}
}

func TestCanDowngrade(t *testing.T) {
	tests := []struct {
		source   string
		target   string
		expected bool
		expectError bool
	}{
		{"1.13", "1.12", true, false},  // Can downgrade
		{"1.13", "1.11", true, false},  // Can downgrade
		{"1.12", "1.11", true, false},  // Can downgrade
		{"1.11", "1.11", true, false},  // Same version (allowed)
		{"1.11", "1.12", false, false}, // Cannot downgrade (upgrade)
		{"1.12", "1.13", false, false}, // Cannot downgrade (upgrade)
		{"2.0", "1.13", false, true},   // Invalid source
		{"1.13", "2.0", false, true},   // Invalid target
	}

	for _, test := range tests {
		t.Run(test.source+"_to_"+test.target, func(t *testing.T) {
			canDowngrade, err := CanDowngrade(test.source, test.target)
			
			if test.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !test.expectError && canDowngrade != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, canDowngrade)
			}
		})
	}
}

func TestVersionFeatures(t *testing.T) {
	// Test 1.11 features (limited)
	vh11, _ := NewVersionHandler("1.11")
	features11 := vh11.GetVersionFeatures()
	
	if features11.SupportsColorCorrection {
		t.Error("Version 1.11 should not support color correction")
	}
	
	if features11.SupportsAdvancedAudio {
		t.Error("Version 1.11 should not support advanced audio")
	}
	
	if features11.SupportsMulticam {
		t.Error("Version 1.11 should not support multicam")
	}
	
	if features11.SupportsRoles {
		t.Error("Version 1.11 should not support roles")
	}
	
	// Test 1.13 features (full)
	vh13, _ := NewVersionHandler("1.13")
	features13 := vh13.GetVersionFeatures()
	
	if !features13.SupportsColorCorrection {
		t.Error("Version 1.13 should support color correction")
	}
	
	if !features13.SupportsAdvancedAudio {
		t.Error("Version 1.13 should support advanced audio")
	}
	
	if !features13.SupportsMulticam {
		t.Error("Version 1.13 should support multicam")
	}
	
	if !features13.SupportsRoles {
		t.Error("Version 1.13 should support roles")
	}
	
	if !features13.SupportsTextStyles {
		t.Error("Version 1.13 should support text styles")
	}
}

func TestValidateFeatureSupport(t *testing.T) {
	vh11, _ := NewVersionHandler("1.11")
	vh13, _ := NewVersionHandler("1.13")
	
	// Test features not supported in 1.11
	if err := vh11.ValidateFeatureSupport("color-correction"); err == nil {
		t.Error("Color correction should not be supported in 1.11")
	}
	
	if err := vh11.ValidateFeatureSupport("advanced-audio"); err == nil {
		t.Error("Advanced audio should not be supported in 1.11")
	}
	
	if err := vh11.ValidateFeatureSupport("multicam"); err == nil {
		t.Error("Multicam should not be supported in 1.11")
	}
	
	if err := vh11.ValidateFeatureSupport("roles"); err == nil {
		t.Error("Roles should not be supported in 1.11")
	}
	
	// Test features supported in 1.13
	if err := vh13.ValidateFeatureSupport("color-correction"); err != nil {
		t.Errorf("Color correction should be supported in 1.13: %v", err)
	}
	
	if err := vh13.ValidateFeatureSupport("advanced-audio"); err != nil {
		t.Errorf("Advanced audio should be supported in 1.13: %v", err)
	}
	
	if err := vh13.ValidateFeatureSupport("multicam"); err != nil {
		t.Errorf("Multicam should be supported in 1.13: %v", err)
	}
	
	if err := vh13.ValidateFeatureSupport("roles"); err != nil {
		t.Errorf("Roles should be supported in 1.13: %v", err)
	}
	
	// Test keyframes (supported in all versions)
	if err := vh11.ValidateFeatureSupport("keyframes"); err != nil {
		t.Errorf("Keyframes should be supported in 1.11: %v", err)
	}
	
	if err := vh13.ValidateFeatureSupport("keyframes"); err != nil {
		t.Errorf("Keyframes should be supported in 1.13: %v", err)
	}
	
	// Test unknown feature
	if err := vh13.ValidateFeatureSupport("unknown-feature"); err == nil {
		t.Error("Unknown feature should not be supported")
	}
}

func TestValidateAudioChannels(t *testing.T) {
	vh11, _ := NewVersionHandler("1.11")
	vh13, _ := NewVersionHandler("1.13")
	
	// Test valid channel counts
	if err := vh11.ValidateAudioChannels(16); err != nil {
		t.Errorf("16 channels should be valid for 1.11: %v", err)
	}
	
	if err := vh13.ValidateAudioChannels(32); err != nil {
		t.Errorf("32 channels should be valid for 1.13: %v", err)
	}
	
	// Test invalid channel counts
	if err := vh11.ValidateAudioChannels(32); err == nil {
		t.Error("32 channels should be invalid for 1.11 (max 24)")
	}
	
	if err := vh13.ValidateAudioChannels(64); err == nil {
		t.Error("64 channels should be invalid for 1.13 (max 32)")
	}
	
	if err := vh13.ValidateAudioChannels(0); err == nil {
		t.Error("0 channels should be invalid")
	}
	
	if err := vh13.ValidateAudioChannels(-1); err == nil {
		t.Error("Negative channels should be invalid")
	}
}

func TestValidateVideoLayers(t *testing.T) {
	vh11, _ := NewVersionHandler("1.11")
	vh13, _ := NewVersionHandler("1.13")
	
	// Test valid layer counts
	if err := vh11.ValidateVideoLayers(25); err != nil {
		t.Errorf("25 layers should be valid for 1.11: %v", err)
	}
	
	if err := vh13.ValidateVideoLayers(100); err != nil {
		t.Errorf("100 layers should be valid for 1.13: %v", err)
	}
	
	// Test invalid layer counts
	if err := vh11.ValidateVideoLayers(75); err == nil {
		t.Error("75 layers should be invalid for 1.11 (max 50)")
	}
	
	if err := vh13.ValidateVideoLayers(150); err == nil {
		t.Error("150 layers should be invalid for 1.13 (max 100)")
	}
	
	if err := vh13.ValidateVideoLayers(0); err == nil {
		t.Error("0 layers should be invalid")
	}
}

func TestValidateColorSpace(t *testing.T) {
	vh13, _ := NewVersionHandler("1.13")
	
	// Test valid color spaces
	validColorSpaces := []string{"1-1-1", "1-13-1", "5-1-6"}
	for _, colorSpace := range validColorSpaces {
		if err := vh13.ValidateColorSpace(colorSpace); err != nil {
			t.Errorf("Color space %s should be valid: %v", colorSpace, err)
		}
	}
	
	// Test invalid color space
	if err := vh13.ValidateColorSpace("invalid-colorspace"); err == nil {
		t.Error("Invalid color space should produce error")
	}
}

func TestValidateAudioRate(t *testing.T) {
	vh13, _ := NewVersionHandler("1.13")
	
	// Test valid audio rates
	validRates := []int{44100, 48000, 96000}
	for _, rate := range validRates {
		if err := vh13.ValidateAudioRate(rate); err != nil {
			t.Errorf("Audio rate %d should be valid: %v", rate, err)
		}
	}
	
	// Test invalid audio rate
	if err := vh13.ValidateAudioRate(22050); err == nil {
		t.Error("Invalid audio rate should produce error")
	}
}

func TestGetUpgradeRequirements(t *testing.T) {
	// Test 1.11 to 1.13 upgrade
	requirements, err := GetUpgradeRequirements("1.11", "1.13")
	if err != nil {
		t.Errorf("Failed to get upgrade requirements: %v", err)
	}
	
	if len(requirements) == 0 {
		t.Error("Expected upgrade requirements but got none")
	}
	
	// Should mention new features available
	foundColorCorrection := false
	foundAdvancedAudio := false
	foundRoles := false
	
	for _, req := range requirements {
		if strings.Contains(req, "Color correction") {
			foundColorCorrection = true
		}
		if strings.Contains(req, "Advanced audio") {
			foundAdvancedAudio = true
		}
		if strings.Contains(req, "roles") {
			foundRoles = true
		}
	}
	
	if !foundColorCorrection {
		t.Error("Should mention color correction availability")
	}
	if !foundAdvancedAudio {
		t.Error("Should mention advanced audio availability")
	}
	if !foundRoles {
		t.Error("Should mention roles availability")
	}
	
	// Test same version
	requirements, err = GetUpgradeRequirements("1.13", "1.13")
	if err != nil {
		t.Errorf("Same version upgrade should not fail: %v", err)
	}
	
	if len(requirements) != 1 || !strings.Contains(requirements[0], "No significant changes") {
		t.Error("Same version should indicate no changes required")
	}
	
	// Test invalid upgrade (downgrade)
	_, err = GetUpgradeRequirements("1.13", "1.11")
	if err == nil {
		t.Error("Downgrade should not be allowed as upgrade")
	}
}

func TestGetDowngradeWarnings(t *testing.T) {
	// Test 1.13 to 1.11 downgrade
	warnings, err := GetDowngradeWarnings("1.13", "1.11")
	if err != nil {
		t.Errorf("Failed to get downgrade warnings: %v", err)
	}
	
	if len(warnings) == 0 {
		t.Error("Expected downgrade warnings but got none")
	}
	
	// Should mention features that will be lost
	foundColorCorrection := false
	foundAdvancedAudio := false
	foundRoles := false
	
	for _, warning := range warnings {
		if strings.Contains(warning, "Color correction") {
			foundColorCorrection = true
		}
		if strings.Contains(warning, "Advanced audio") {
			foundAdvancedAudio = true
		}
		if strings.Contains(warning, "roles") {
			foundRoles = true
		}
	}
	
	if !foundColorCorrection {
		t.Error("Should warn about losing color correction")
	}
	if !foundAdvancedAudio {
		t.Error("Should warn about losing advanced audio")
	}
	if !foundRoles {
		t.Error("Should warn about losing roles")
	}
	
	// Test same version
	warnings, err = GetDowngradeWarnings("1.13", "1.13")
	if err != nil {
		t.Errorf("Same version downgrade should not fail: %v", err)
	}
	
	if len(warnings) != 1 || !strings.Contains(warnings[0], "No significant features") {
		t.Error("Same version should indicate no features lost")
	}
	
	// Test invalid downgrade (upgrade)
	_, err = GetDowngradeWarnings("1.11", "1.13")
	if err == nil {
		t.Error("Upgrade should not be allowed as downgrade")
	}
}

func TestGenerateCompatibilityReport(t *testing.T) {
	// Test upgrade scenario
	report, err := GenerateCompatibilityReport("1.11", "1.13")
	if err != nil {
		t.Errorf("Failed to generate compatibility report: %v", err)
	}
	
	if report.SourceVersion != "1.11" {
		t.Errorf("Expected source version 1.11, got %s", report.SourceVersion)
	}
	
	if report.TargetVersion != "1.13" {
		t.Errorf("Expected target version 1.13, got %s", report.TargetVersion)
	}
	
	if report.CompatibilityDirection != "upgrade" {
		t.Errorf("Expected upgrade direction, got %s", report.CompatibilityDirection)
	}
	
	if !report.Supported {
		t.Error("Upgrade should be supported")
	}
	
	if len(report.Requirements) == 0 {
		t.Error("Expected upgrade requirements")
	}
	
	// Test downgrade scenario
	report, err = GenerateCompatibilityReport("1.13", "1.11")
	if err != nil {
		t.Errorf("Failed to generate downgrade report: %v", err)
	}
	
	if report.CompatibilityDirection != "downgrade" {
		t.Errorf("Expected downgrade direction, got %s", report.CompatibilityDirection)
	}
	
	if len(report.Warnings) == 0 {
		t.Error("Expected downgrade warnings")
	}
	
	// Test same version scenario
	report, err = GenerateCompatibilityReport("1.13", "1.13")
	if err != nil {
		t.Errorf("Failed to generate same version report: %v", err)
	}
	
	if report.CompatibilityDirection != "same" {
		t.Errorf("Expected same direction, got %s", report.CompatibilityDirection)
	}
	
	// Test invalid versions
	_, err = GenerateCompatibilityReport("2.0", "1.13")
	if err == nil {
		t.Error("Should fail with invalid source version")
	}
	
	_, err = GenerateCompatibilityReport("1.13", "2.0")
	if err == nil {
		t.Error("Should fail with invalid target version")
	}
}

func TestVersionCompatibilityReport_String(t *testing.T) {
	report := VersionCompatibilityReport{
		SourceVersion:          "1.11",
		TargetVersion:          "1.13",
		CompatibilityDirection: "upgrade",
		Requirements:           []string{"Color correction will become available"},
		Warnings:               []string{},
		Supported:              true,
	}
	
	reportStr := report.String()
	
	if !strings.Contains(reportStr, "Version Compatibility Report") {
		t.Error("Report should contain title")
	}
	
	if !strings.Contains(reportStr, "Source Version: 1.11") {
		t.Error("Report should contain source version")
	}
	
	if !strings.Contains(reportStr, "Target Version: 1.13") {
		t.Error("Report should contain target version")
	}
	
	if !strings.Contains(reportStr, "Direction: upgrade") {
		t.Error("Report should contain direction")
	}
	
	if !strings.Contains(reportStr, "Supported: PASSED") {
		t.Error("Report should show supported status")
	}
	
	if !strings.Contains(reportStr, "Requirements:") {
		t.Error("Report should contain requirements section")
	}
	
	if !strings.Contains(reportStr, "Color correction") {
		t.Error("Report should contain specific requirements")
	}
}