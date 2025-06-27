// Package media_constraints implements Step 3 of the FCPXMLKit-inspired refactoring plan:
// FCPXML-aware constraint system for different media types.
//
// This provides comprehensive validation rules for images, videos, and audio files
// to prevent the most common FCPXML generation errors.
package fcp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MediaTypeConstraints defines validation rules for different media types
type MediaTypeConstraints struct {
	RequiredAttributes   []string
	ForbiddenAttributes  []string
	RequiredChildren     []string
	ValidDurationPattern string
	ValidFormatRules     *FormatConstraints
	ValidSpineElement    string // "video", "asset-clip", or "both"
}

// FormatConstraints defines validation rules for format elements
type FormatConstraints struct {
	RequiredAttributes  []string
	ForbiddenAttributes []string
	ValidFrameDuration  *string // nil means forbidden, non-nil means required pattern
}

// MediaConstraintValidator validates media type constraints
type MediaConstraintValidator struct {
	rules map[MediaType]MediaTypeConstraints
}

// NewMediaConstraintValidator creates a new media constraint validator
func NewMediaConstraintValidator() *MediaConstraintValidator {
	validator := &MediaConstraintValidator{
		rules: make(map[MediaType]MediaTypeConstraints),
	}
	
	validator.initializeDefaultRules()
	return validator
}

// initializeDefaultRules sets up the default FCPXML media type rules
func (mcv *MediaConstraintValidator) initializeDefaultRules() {
	// Image constraints - the most restrictive
	mcv.rules[MediaTypeImage] = MediaTypeConstraints{
		RequiredAttributes: []string{"hasVideo"},
		ForbiddenAttributes: []string{
			"hasAudio", "audioSources", "audioChannels", "audioRate",
		},
		ValidDurationPattern: "^0s$", // Images must have 0s duration
		ValidFormatRules: &FormatConstraints{
			RequiredAttributes:  []string{"width", "height", "colorSpace"},
			ForbiddenAttributes: []string{"frameDuration"}, // CRITICAL: Images can't have frameDuration
		},
		ValidSpineElement: "video", // Images use Video elements, NOT AssetClip
	}
	
	// Video constraints - flexible but with required video properties
	mcv.rules[MediaTypeVideo] = MediaTypeConstraints{
		RequiredAttributes: []string{"hasVideo", "videoSources"},
		ForbiddenAttributes: []string{}, // Videos can have audio
		ValidDurationPattern: `^\d+/\d+s$`, // Videos have real duration in rational format
		ValidFormatRules: &FormatConstraints{
			RequiredAttributes:  []string{"frameDuration", "width", "height", "colorSpace"},
			ForbiddenAttributes: []string{},
		},
		ValidSpineElement: "asset-clip", // Videos use AssetClip elements
	}
	
	// Audio constraints - no video properties allowed
	mcv.rules[MediaTypeAudio] = MediaTypeConstraints{
		RequiredAttributes: []string{"hasAudio", "audioSources"},
		ForbiddenAttributes: []string{
			"hasVideo", "videoSources", "format", // Audio typically doesn't use format
		},
		ValidDurationPattern: `^\d+/\d+s$`, // Audio has real duration
		ValidFormatRules: nil, // Audio typically doesn't have format
		ValidSpineElement: "asset-clip", // Audio uses AssetClip elements
	}
}

// ValidateAsset validates an asset against media type constraints
func (mcv *MediaConstraintValidator) ValidateAsset(asset *Asset, mediaType MediaType) error {
	constraints, exists := mcv.rules[mediaType]
	if !exists {
		return fmt.Errorf("no constraints defined for media type: %s", mediaType)
	}
	
	// Validate required attributes
	for _, required := range constraints.RequiredAttributes {
		if !mcv.assetHasAttribute(asset, required) {
			return fmt.Errorf("missing required attribute for %s: %s", mediaType, required)
		}
	}
	
	// Validate forbidden attributes
	for _, forbidden := range constraints.ForbiddenAttributes {
		if mcv.assetHasAttribute(asset, forbidden) {
			return fmt.Errorf("forbidden attribute for %s: %s", mediaType, forbidden)
		}
	}
	
	// Validate duration pattern
	if constraints.ValidDurationPattern != "" {
		matched, err := regexp.MatchString(constraints.ValidDurationPattern, asset.Duration)
		if err != nil {
			return fmt.Errorf("invalid duration pattern regex: %v", err)
		}
		if !matched {
			return fmt.Errorf("invalid duration for %s: %s (must match %s)", 
				mediaType, asset.Duration, constraints.ValidDurationPattern)
		}
	}
	
	return nil
}

// ValidateFormat validates a format against media type constraints
func (mcv *MediaConstraintValidator) ValidateFormat(format *Format, mediaType MediaType) error {
	constraints, exists := mcv.rules[mediaType]
	if !exists {
		return fmt.Errorf("no constraints defined for media type: %s", mediaType)
	}
	
	// Some media types don't use formats
	if constraints.ValidFormatRules == nil {
		if format != nil {
			return fmt.Errorf("media type %s should not have format", mediaType)
		}
		return nil
	}
	
	if format == nil {
		return fmt.Errorf("media type %s requires format", mediaType)
	}
	
	formatRules := constraints.ValidFormatRules
	
	// Validate required attributes
	for _, required := range formatRules.RequiredAttributes {
		if !mcv.formatHasAttribute(format, required) {
			return fmt.Errorf("missing required format attribute for %s: %s", mediaType, required)
		}
	}
	
	// Validate forbidden attributes
	for _, forbidden := range formatRules.ForbiddenAttributes {
		if mcv.formatHasAttribute(format, forbidden) {
			return fmt.Errorf("forbidden format attribute for %s: %s", mediaType, forbidden)
		}
	}
	
	return nil
}

// ValidateSpineElement validates that the correct spine element type is used
func (mcv *MediaConstraintValidator) ValidateSpineElement(elementType string, mediaType MediaType) error {
	constraints, exists := mcv.rules[mediaType]
	if !exists {
		return fmt.Errorf("no constraints defined for media type: %s", mediaType)
	}
	
	validElement := constraints.ValidSpineElement
	if validElement == "both" {
		if elementType != "video" && elementType != "asset-clip" {
			return fmt.Errorf("media type %s requires 'video' or 'asset-clip' element, got: %s", 
				mediaType, elementType)
		}
	} else if elementType != validElement {
		return fmt.Errorf("media type %s requires '%s' element, got: %s", 
			mediaType, validElement, elementType)
	}
	
	return nil
}

// assetHasAttribute checks if an asset has a specific attribute set
func (mcv *MediaConstraintValidator) assetHasAttribute(asset *Asset, attrName string) bool {
	switch attrName {
	case "hasVideo":
		return asset.HasVideo != ""
	case "hasAudio":
		return asset.HasAudio != ""
	case "videoSources":
		return asset.VideoSources != ""
	case "audioSources":
		return asset.AudioSources != ""
	case "audioChannels":
		return asset.AudioChannels != ""
	case "audioRate":
		return asset.AudioRate != ""
	case "format":
		return asset.Format != ""
	default:
		return false
	}
}

// formatHasAttribute checks if a format has a specific attribute set
func (mcv *MediaConstraintValidator) formatHasAttribute(format *Format, attrName string) bool {
	switch attrName {
	case "frameDuration":
		return format.FrameDuration != ""
	case "width":
		return format.Width != ""
	case "height":
		return format.Height != ""
	case "colorSpace":
		return format.ColorSpace != ""
	case "name":
		return format.Name != ""
	default:
		return false
	}
}

// MediaTypeSpecificValidator provides validation for specific media type combinations
type MediaTypeSpecificValidator struct {
	constraintValidator *MediaConstraintValidator
}

// NewMediaTypeSpecificValidator creates a new media type specific validator
func NewMediaTypeSpecificValidator() *MediaTypeSpecificValidator {
	return &MediaTypeSpecificValidator{
		constraintValidator: NewMediaConstraintValidator(),
	}
}

// ValidateAssetFormatPair validates an asset and format pair for consistency
func (mtsv *MediaTypeSpecificValidator) ValidateAssetFormatPair(asset *Asset, format *Format, mediaType MediaType) error {
	// Validate asset constraints
	if err := mtsv.constraintValidator.ValidateAsset(asset, mediaType); err != nil {
		return fmt.Errorf("asset validation failed: %v", err)
	}
	
	// Validate format constraints
	if err := mtsv.constraintValidator.ValidateFormat(format, mediaType); err != nil {
		return fmt.Errorf("format validation failed: %v", err)
	}
	
	// Validate asset-format consistency
	return mtsv.validateAssetFormatConsistency(asset, format, mediaType)
}

// validateAssetFormatConsistency ensures asset and format are consistent
func (mtsv *MediaTypeSpecificValidator) validateAssetFormatConsistency(asset *Asset, format *Format, mediaType MediaType) error {
	if format == nil {
		return nil // Some media types don't have formats
	}
	
	// Check that asset references the format
	if asset.Format != "" && asset.Format != format.ID {
		return fmt.Errorf("asset format reference %s doesn't match format ID %s", 
			asset.Format, format.ID)
	}
	
	// Media type specific consistency checks
	switch mediaType {
	case MediaTypeImage:
		// Images must have duration="0s" and format without frameDuration
		if asset.Duration != "0s" {
			return fmt.Errorf("image asset must have duration='0s', got: %s", asset.Duration)
		}
		if format.FrameDuration != "" {
			return fmt.Errorf("image format cannot have frameDuration, got: %s", format.FrameDuration)
		}
	
	case MediaTypeVideo:
		// Videos must have real duration and format with frameDuration
		if asset.Duration == "0s" {
			return fmt.Errorf("video asset cannot have duration='0s'")
		}
		if format.FrameDuration == "" {
			return fmt.Errorf("video format must have frameDuration")
		}
		
		// Validate frameDuration format
		if !mtsv.isValidFrameDuration(format.FrameDuration) {
			return fmt.Errorf("invalid frameDuration format: %s", format.FrameDuration)
		}
	
	case MediaTypeAudio:
		// Audio constraints are handled by media type rules
		if format != nil {
			return fmt.Errorf("audio assets should not have format")
		}
	}
	
	return nil
}

// isValidFrameDuration validates frameDuration format
func (mtsv *MediaTypeSpecificValidator) isValidFrameDuration(frameDuration string) bool {
	// FrameDuration should be in rational format like "1001/30000s"
	matched, err := regexp.MatchString(`^\d+/\d+s$`, frameDuration)
	if err != nil || !matched {
		return false
	}
	
	// Additional validation could check for common frame rates
	commonFrameRates := []string{
		"1001/24000s",  // 23.976 fps
		"1/24s",        // 24 fps
		"1/25s",        // 25 fps
		"1001/30000s",  // 29.97 fps
		"1/30s",        // 30 fps
		"1/50s",        // 50 fps
		"1001/60000s",  // 59.94 fps
		"1/60s",        // 60 fps
	}
	
	for _, validRate := range commonFrameRates {
		if frameDuration == validRate {
			return true
		}
	}
	
	// Allow other rates but warn about unusual values
	return true
}

// KeyframeConstraintValidator validates keyframe constraints based on parameter type
type KeyframeConstraintValidator struct{}

// NewKeyframeConstraintValidator creates a new keyframe constraint validator
func NewKeyframeConstraintValidator() *KeyframeConstraintValidator {
	return &KeyframeConstraintValidator{}
}

// ValidateKeyframseForParameter validates keyframes for a specific parameter type
func (kcv *KeyframeConstraintValidator) ValidateKeyframseForParameter(paramName string, keyframes []Keyframe) error {
	for i, keyframe := range keyframes {
		if err := kcv.validateKeyframeForParameter(paramName, keyframe); err != nil {
			return fmt.Errorf("keyframe %d validation failed: %v", i, err)
		}
	}
	
	return nil
}

// validateKeyframeForParameter validates a single keyframe for a parameter type
func (kcv *KeyframeConstraintValidator) validateKeyframeForParameter(paramName string, keyframe Keyframe) error {
	switch paramName {
	case "position":
		// Position keyframes: NO attributes allowed
		if keyframe.Interp != "" {
			return fmt.Errorf("position keyframes cannot have interp attribute")
		}
		if keyframe.Curve != "" {
			return fmt.Errorf("position keyframes cannot have curve attribute")
		}
		
		// Validate position value format (should be "x y")
		return kcv.validatePositionValue(keyframe.Value)
		
	case "scale", "rotation", "anchor":
		// Scale/Rotation/Anchor keyframes: Only curve attribute allowed
		if keyframe.Interp != "" {
			return fmt.Errorf("%s keyframes cannot have interp attribute", paramName)
		}
		
		if keyframe.Curve != "" {
			validCurves := []string{"linear", "smooth", "hold"}
			if !contains(validCurves, keyframe.Curve) {
				return fmt.Errorf("invalid curve value for %s: %s", paramName, keyframe.Curve)
			}
		}
		
	case "opacity", "volume":
		// Opacity/Volume keyframes: Both interp and curve allowed
		if keyframe.Interp != "" {
			validInterps := []string{"linear", "easeIn", "easeOut", "easeInOut"}
			if !contains(validInterps, keyframe.Interp) {
				return fmt.Errorf("invalid interp value for %s: %s", paramName, keyframe.Interp)
			}
		}
		
		if keyframe.Curve != "" {
			validCurves := []string{"linear", "smooth", "hold"}
			if !contains(validCurves, keyframe.Curve) {
				return fmt.Errorf("invalid curve value for %s: %s", paramName, keyframe.Curve)
			}
		}
		
	default:
		// Unknown parameter - be permissive but validate basic formats
		break
	}
	
	return nil
}

// validatePositionValue validates position keyframe value format
func (kcv *KeyframeConstraintValidator) validatePositionValue(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return fmt.Errorf("position value must have 2 components: %s", value)
	}
	
	for i, part := range parts {
		// Try to parse as float to ensure it's a valid number
		if _, err := parseFloatOrInt(part); err != nil {
			return fmt.Errorf("invalid position component %d: %s", i, part)
		}
	}
	
	return nil
}

// parseFloatOrInt parses a string as either float or int
func parseFloatOrInt(s string) (float64, error) {
	// Try int first for exact representation
	if intVal, err := parseInt(s); err == nil {
		return float64(intVal), nil
	}
	
	// Try float
	return parseFloat(s)
}

// parseInt parses string as int
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseFloat parses string as float64
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}