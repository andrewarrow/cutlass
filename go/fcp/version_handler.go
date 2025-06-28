// Package version_handler implements FCPXML version management.
// This file is part of Step 20: Create Migration and Compatibility Layer.
//
// Provides:
// - Version detection and validation
// - Version-specific feature support
// - Cross-version compatibility handling
// - Version upgrade/downgrade logic
package fcp

import (
	"fmt"
	"strings"
	"encoding/xml"
)

// FCPXMLVersion represents a Final Cut Pro XML version
type FCPXMLVersion struct {
	Major int
	Minor int
	String string
}

// SupportedVersions lists all supported FCPXML versions
var SupportedVersions = map[string]FCPXMLVersion{
	"1.11": {Major: 1, Minor: 11, String: "1.11"},
	"1.12": {Major: 1, Minor: 12, String: "1.12"},
	"1.13": {Major: 1, Minor: 13, String: "1.13"},
}

// CurrentVersion is the default version for new FCPXML documents
const CurrentVersion = "1.13"

// MinimumSupportedVersion is the oldest version we support
const MinimumSupportedVersion = "1.11"

// VersionHandler manages FCPXML version compatibility
type VersionHandler struct {
	version FCPXMLVersion
	features VersionFeatures
}

// VersionFeatures defines what features are available in each version
type VersionFeatures struct {
	SupportsColorCorrection  bool
	SupportsAdvancedAudio    bool
	SupportsMulticam         bool
	SupportsCompoundClips    bool
	SupportsRoles            bool
	SupportsKeyframes        bool
	SupportsEffects          bool
	SupportsTransitions      bool
	SupportsGenerators       bool
	SupportsTextStyles       bool
	MaxAudioChannels         int
	MaxVideoLayers           int
	SupportedColorSpaces     []string
	SupportedAudioRates      []int
}

// NewVersionHandler creates a version handler for the specified version
func NewVersionHandler(versionString string) (*VersionHandler, error) {
	version, exists := SupportedVersions[versionString]
	if !exists {
		return nil, fmt.Errorf("unsupported FCPXML version: %s", versionString)
	}
	
	features := getVersionFeatures(version)
	
	return &VersionHandler{
		version:  version,
		features: features,
	}, nil
}

// DetectVersion detects the FCPXML version from XML data
func DetectVersion(xmlData []byte) (string, error) {
	var doc struct {
		XMLName xml.Name `xml:"fcpxml"`
		Version string   `xml:"version,attr"`
	}
	
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return "", fmt.Errorf("failed to parse XML for version detection: %v", err)
	}
	
	if doc.XMLName.Local != "fcpxml" {
		return "", fmt.Errorf("not an FCPXML document")
	}
	
	if doc.Version == "" {
		return "", fmt.Errorf("FCPXML document missing version attribute")
	}
	
	// Validate the detected version
	if _, exists := SupportedVersions[doc.Version]; !exists {
		return "", fmt.Errorf("unsupported FCPXML version: %s", doc.Version)
	}
	
	return doc.Version, nil
}

// ValidateVersion checks if a version string is valid and supported
func ValidateVersion(versionString string) error {
	if versionString == "" {
		return fmt.Errorf("version cannot be empty")
	}
	
	if _, exists := SupportedVersions[versionString]; !exists {
		return fmt.Errorf("unsupported version: %s (supported: %s)", 
			versionString, strings.Join(getSupportedVersionList(), ", "))
	}
	
	return nil
}

// IsVersionSupported checks if a version is supported
func IsVersionSupported(versionString string) bool {
	_, exists := SupportedVersions[versionString]
	return exists
}

// CompareVersions compares two version strings (-1: v1 < v2, 0: v1 == v2, 1: v1 > v2)
func CompareVersions(v1, v2 string) (int, error) {
	version1, exists := SupportedVersions[v1]
	if !exists {
		return 0, fmt.Errorf("unsupported version: %s", v1)
	}
	
	version2, exists := SupportedVersions[v2]
	if !exists {
		return 0, fmt.Errorf("unsupported version: %s", v2)
	}
	
	if version1.Major < version2.Major {
		return -1, nil
	} else if version1.Major > version2.Major {
		return 1, nil
	}
	
	if version1.Minor < version2.Minor {
		return -1, nil
	} else if version1.Minor > version2.Minor {
		return 1, nil
	}
	
	return 0, nil
}

// CanUpgrade checks if an upgrade path exists from source to target version
func CanUpgrade(sourceVersion, targetVersion string) (bool, error) {
	comparison, err := CompareVersions(sourceVersion, targetVersion)
	if err != nil {
		return false, err
	}
	
	// Can upgrade if target version is newer or same
	return comparison <= 0, nil
}

// CanDowngrade checks if a downgrade path exists from source to target version
func CanDowngrade(sourceVersion, targetVersion string) (bool, error) {
	comparison, err := CompareVersions(sourceVersion, targetVersion)
	if err != nil {
		return false, err
	}
	
	// Can downgrade if target version is older or same
	return comparison >= 0, nil
}

// GetVersionFeatures returns the features available for a version
func (vh *VersionHandler) GetVersionFeatures() VersionFeatures {
	return vh.features
}

// GetVersion returns the version information
func (vh *VersionHandler) GetVersion() FCPXMLVersion {
	return vh.version
}

// ValidateFeatureSupport validates if a feature is supported in this version
func (vh *VersionHandler) ValidateFeatureSupport(feature string) error {
	switch feature {
	case "color-correction":
		if !vh.features.SupportsColorCorrection {
			return fmt.Errorf("color correction not supported in version %s", vh.version.String)
		}
	case "advanced-audio":
		if !vh.features.SupportsAdvancedAudio {
			return fmt.Errorf("advanced audio not supported in version %s", vh.version.String)
		}
	case "multicam":
		if !vh.features.SupportsMulticam {
			return fmt.Errorf("multicam not supported in version %s", vh.version.String)
		}
	case "compound-clips":
		if !vh.features.SupportsCompoundClips {
			return fmt.Errorf("compound clips not supported in version %s", vh.version.String)
		}
	case "roles":
		if !vh.features.SupportsRoles {
			return fmt.Errorf("roles not supported in version %s", vh.version.String)
		}
	case "keyframes":
		if !vh.features.SupportsKeyframes {
			return fmt.Errorf("keyframes not supported in version %s", vh.version.String)
		}
	default:
		return fmt.Errorf("unknown feature: %s", feature)
	}
	
	return nil
}

// ValidateAudioChannels validates audio channel count for this version
func (vh *VersionHandler) ValidateAudioChannels(channels int) error {
	if channels > vh.features.MaxAudioChannels {
		return fmt.Errorf("version %s supports maximum %d audio channels, got %d", 
			vh.version.String, vh.features.MaxAudioChannels, channels)
	}
	
	if channels < 1 {
		return fmt.Errorf("audio channels must be at least 1, got %d", channels)
	}
	
	return nil
}

// ValidateVideoLayers validates video layer count for this version
func (vh *VersionHandler) ValidateVideoLayers(layers int) error {
	if layers > vh.features.MaxVideoLayers {
		return fmt.Errorf("version %s supports maximum %d video layers, got %d", 
			vh.version.String, vh.features.MaxVideoLayers, layers)
	}
	
	if layers < 1 {
		return fmt.Errorf("video layers must be at least 1, got %d", layers)
	}
	
	return nil
}

// ValidateColorSpace validates if a color space is supported in this version
func (vh *VersionHandler) ValidateColorSpace(colorSpace string) error {
	for _, supported := range vh.features.SupportedColorSpaces {
		if supported == colorSpace {
			return nil
		}
	}
	
	return fmt.Errorf("color space %s not supported in version %s (supported: %s)", 
		colorSpace, vh.version.String, strings.Join(vh.features.SupportedColorSpaces, ", "))
}

// ValidateAudioRate validates if an audio sample rate is supported
func (vh *VersionHandler) ValidateAudioRate(rate int) error {
	for _, supported := range vh.features.SupportedAudioRates {
		if supported == rate {
			return nil
		}
	}
	
	return fmt.Errorf("audio rate %d not supported in version %s (supported: %v)", 
		rate, vh.version.String, vh.features.SupportedAudioRates)
}

// GetUpgradeRequirements returns what's needed to upgrade to a target version
func GetUpgradeRequirements(sourceVersion, targetVersion string) ([]string, error) {
	canUpgrade, err := CanUpgrade(sourceVersion, targetVersion)
	if err != nil {
		return nil, err
	}
	
	if !canUpgrade {
		return nil, fmt.Errorf("cannot upgrade from %s to %s", sourceVersion, targetVersion)
	}
	
	var requirements []string
	
	sourceHandler, _ := NewVersionHandler(sourceVersion)
	targetHandler, _ := NewVersionHandler(targetVersion)
	
	// Check what features will be available after upgrade
	if !sourceHandler.features.SupportsColorCorrection && targetHandler.features.SupportsColorCorrection {
		requirements = append(requirements, "Color correction features will become available")
	}
	
	if !sourceHandler.features.SupportsAdvancedAudio && targetHandler.features.SupportsAdvancedAudio {
		requirements = append(requirements, "Advanced audio features will become available")
	}
	
	if !sourceHandler.features.SupportsMulticam && targetHandler.features.SupportsMulticam {
		requirements = append(requirements, "Multicam features will become available")
	}
	
	if !sourceHandler.features.SupportsRoles && targetHandler.features.SupportsRoles {
		requirements = append(requirements, "Audio/video roles will become available")
	}
	
	// Check for potential compatibility issues
	if targetHandler.features.MaxAudioChannels > sourceHandler.features.MaxAudioChannels {
		requirements = append(requirements, fmt.Sprintf("Audio channel limit increased from %d to %d", 
			sourceHandler.features.MaxAudioChannels, targetHandler.features.MaxAudioChannels))
	}
	
	if len(requirements) == 0 {
		requirements = append(requirements, "No significant changes required for this upgrade")
	}
	
	return requirements, nil
}

// GetDowngradeWarnings returns warnings about downgrading to a target version
func GetDowngradeWarnings(sourceVersion, targetVersion string) ([]string, error) {
	canDowngrade, err := CanDowngrade(sourceVersion, targetVersion)
	if err != nil {
		return nil, err
	}
	
	if !canDowngrade {
		return nil, fmt.Errorf("cannot downgrade from %s to %s", sourceVersion, targetVersion)
	}
	
	var warnings []string
	
	sourceHandler, _ := NewVersionHandler(sourceVersion)
	targetHandler, _ := NewVersionHandler(targetVersion)
	
	// Check what features will be lost in downgrade
	if sourceHandler.features.SupportsColorCorrection && !targetHandler.features.SupportsColorCorrection {
		warnings = append(warnings, "Color correction features will be lost")
	}
	
	if sourceHandler.features.SupportsAdvancedAudio && !targetHandler.features.SupportsAdvancedAudio {
		warnings = append(warnings, "Advanced audio features will be lost")
	}
	
	if sourceHandler.features.SupportsMulticam && !targetHandler.features.SupportsMulticam {
		warnings = append(warnings, "Multicam features will be lost")
	}
	
	if sourceHandler.features.SupportsRoles && !targetHandler.features.SupportsRoles {
		warnings = append(warnings, "Audio/video roles will be lost")
	}
	
	// Check for capability reductions
	if sourceHandler.features.MaxAudioChannels > targetHandler.features.MaxAudioChannels {
		warnings = append(warnings, fmt.Sprintf("Audio channel limit reduced from %d to %d", 
			sourceHandler.features.MaxAudioChannels, targetHandler.features.MaxAudioChannels))
	}
	
	if len(warnings) == 0 {
		warnings = append(warnings, "No significant features will be lost in this downgrade")
	}
	
	return warnings, nil
}

// Helper functions

func getVersionFeatures(version FCPXMLVersion) VersionFeatures {
	baseFeatures := VersionFeatures{
		SupportsKeyframes:     true,
		SupportsEffects:       true,
		SupportsTransitions:   true,
		MaxAudioChannels:      32,
		MaxVideoLayers:        100,
		SupportedColorSpaces:  []string{"1-1-1", "1-13-1", "5-1-6"},
		SupportedAudioRates:   []int{44100, 48000, 96000},
	}
	
	switch version.String {
	case "1.11":
		baseFeatures.SupportsColorCorrection = false
		baseFeatures.SupportsAdvancedAudio = false
		baseFeatures.SupportsMulticam = false
		baseFeatures.SupportsCompoundClips = false
		baseFeatures.SupportsRoles = false
		baseFeatures.SupportsGenerators = false
		baseFeatures.SupportsTextStyles = false
		baseFeatures.MaxAudioChannels = 24
		baseFeatures.MaxVideoLayers = 50
		
	case "1.12":
		baseFeatures.SupportsColorCorrection = true
		baseFeatures.SupportsAdvancedAudio = false
		baseFeatures.SupportsMulticam = true
		baseFeatures.SupportsCompoundClips = true
		baseFeatures.SupportsRoles = false
		baseFeatures.SupportsGenerators = true
		baseFeatures.SupportsTextStyles = false
		baseFeatures.MaxAudioChannels = 32
		baseFeatures.MaxVideoLayers = 75
		
	case "1.13":
		baseFeatures.SupportsColorCorrection = true
		baseFeatures.SupportsAdvancedAudio = true
		baseFeatures.SupportsMulticam = true
		baseFeatures.SupportsCompoundClips = true
		baseFeatures.SupportsRoles = true
		baseFeatures.SupportsGenerators = true
		baseFeatures.SupportsTextStyles = true
		baseFeatures.MaxAudioChannels = 32
		baseFeatures.MaxVideoLayers = 100
	}
	
	return baseFeatures
}

func getSupportedVersionList() []string {
	var versions []string
	for version := range SupportedVersions {
		versions = append(versions, version)
	}
	return versions
}

// VersionCompatibilityReport provides comprehensive version compatibility information
type VersionCompatibilityReport struct {
	SourceVersion    string
	TargetVersion    string
	CompatibilityDirection string // "upgrade", "downgrade", "same"
	FeatureChanges   []string
	Warnings         []string
	Requirements     []string
	Supported        bool
}

// GenerateCompatibilityReport generates a comprehensive compatibility report
func GenerateCompatibilityReport(sourceVersion, targetVersion string) (VersionCompatibilityReport, error) {
	report := VersionCompatibilityReport{
		SourceVersion: sourceVersion,
		TargetVersion: targetVersion,
		FeatureChanges: make([]string, 0),
		Warnings: make([]string, 0),
		Requirements: make([]string, 0),
	}
	
	// Validate both versions
	if err := ValidateVersion(sourceVersion); err != nil {
		return report, fmt.Errorf("invalid source version: %v", err)
	}
	
	if err := ValidateVersion(targetVersion); err != nil {
		return report, fmt.Errorf("invalid target version: %v", err)
	}
	
	// Determine compatibility direction
	comparison, err := CompareVersions(sourceVersion, targetVersion)
	if err != nil {
		return report, err
	}
	
	switch {
	case comparison < 0:
		report.CompatibilityDirection = "upgrade"
		report.Supported = true
		if requirements, err := GetUpgradeRequirements(sourceVersion, targetVersion); err == nil {
			report.Requirements = requirements
		}
		
	case comparison > 0:
		report.CompatibilityDirection = "downgrade"
		report.Supported = true
		if warnings, err := GetDowngradeWarnings(sourceVersion, targetVersion); err == nil {
			report.Warnings = warnings
		}
		
	default:
		report.CompatibilityDirection = "same"
		report.Supported = true
		report.Requirements = []string{"No migration needed - versions are identical"}
	}
	
	return report, nil
}

// String returns a human-readable compatibility report
func (vcr VersionCompatibilityReport) String() string {
	var lines []string
	
	lines = append(lines, "=== Version Compatibility Report ===")
	lines = append(lines, fmt.Sprintf("Source Version: %s", vcr.SourceVersion))
	lines = append(lines, fmt.Sprintf("Target Version: %s", vcr.TargetVersion))
	lines = append(lines, fmt.Sprintf("Direction: %s", vcr.CompatibilityDirection))
	lines = append(lines, fmt.Sprintf("Supported: %s", boolToStatus(vcr.Supported)))
	
	if len(vcr.Requirements) > 0 {
		lines = append(lines, "\nRequirements:")
		for _, req := range vcr.Requirements {
			lines = append(lines, fmt.Sprintf("  - %s", req))
		}
	}
	
	if len(vcr.Warnings) > 0 {
		lines = append(lines, "\nWarnings:")
		for _, warning := range vcr.Warnings {
			lines = append(lines, fmt.Sprintf("  - %s", warning))
		}
	}
	
	if len(vcr.FeatureChanges) > 0 {
		lines = append(lines, "\nFeature Changes:")
		for _, change := range vcr.FeatureChanges {
			lines = append(lines, fmt.Sprintf("  - %s", change))
		}
	}
	
	return strings.Join(lines, "\n")
}