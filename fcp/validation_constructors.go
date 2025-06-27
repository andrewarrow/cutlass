// Package validation_constructors implements Step 5 of the FCPXMLKit-inspired refactoring plan:
// Validation-aware constructors that replace direct struct creation with safe, validated constructors.
//
// This ensures all FCPXML structures are valid at construction time, preventing runtime errors.
package fcp

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatedAsset represents an asset with validation-aware construction
type ValidatedAsset struct {
	Asset
	mediaType MediaType
	validator *StructValidator
}

// NewValidatedAsset creates a new asset with full validation
func NewValidatedAsset(id ID, name, uid string, duration Duration, mediaType MediaType) (*ValidatedAsset, error) {
	// Validate inputs
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("invalid asset ID: %v", err)
	}
	
	if err := duration.Validate(); err != nil {
		return nil, fmt.Errorf("invalid duration: %v", err)
	}
	
	if err := mediaType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid media type: %v", err)
	}
	
	if name == "" {
		return nil, fmt.Errorf("asset name cannot be empty")
	}
	
	if uid == "" {
		return nil, fmt.Errorf("asset UID cannot be empty")
	}
	
	// Create basic asset structure
	asset := Asset{
		ID:       string(id),
		Name:     name,
		UID:      uid,
		Start:    "0s",
		Duration: string(duration),
		MediaRep: MediaRep{
			Kind: "original-media",
			Sig:  uid, // Use UID as signature
		},
	}
	
	// Apply media type constraints
	if err := applyMediaTypeConstraints(&asset, mediaType); err != nil {
		return nil, fmt.Errorf("media type constraint validation failed: %v", err)
	}
	
	// Create validated asset
	validatedAsset := &ValidatedAsset{
		Asset:     asset,
		mediaType: mediaType,
		validator: NewStructValidator(),
	}
	
	// Final validation
	if err := validatedAsset.Validate(); err != nil {
		return nil, fmt.Errorf("asset validation failed: %v", err)
	}
	
	return validatedAsset, nil
}

// NewValidatedAssetFromPath creates a validated asset from a file path with auto-detection
func NewValidatedAssetFromPath(id ID, filePath string, duration Duration) (*ValidatedAsset, error) {
	// Detect media type from file extension
	mediaType, err := DetectMediaTypeFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect media type: %v", err)
	}
	
	// Generate name from filename
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	
	// Generate UID from file path
	uid := generateUID(filePath)
	
	// Get absolute path for media rep
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}
	
	// Create validated asset
	validatedAsset, err := NewValidatedAsset(id, name, uid, duration, mediaType)
	if err != nil {
		return nil, err
	}
	
	// Set the source path
	validatedAsset.Asset.MediaRep.Src = "file://" + absPath
	
	return validatedAsset, nil
}

// Validate validates the asset using comprehensive validation
func (va *ValidatedAsset) Validate() error {
	// Validate using struct validator
	if err := va.validator.ValidateStruct(&va.Asset); err != nil {
		return err
	}
	
	// Validate media type constraints
	constraintValidator := NewMediaConstraintValidator()
	return constraintValidator.ValidateAsset(&va.Asset, va.mediaType)
}

// GetMediaType returns the media type
func (va *ValidatedAsset) GetMediaType() MediaType {
	return va.mediaType
}

// applyMediaTypeConstraints applies media type specific constraints to an asset
func applyMediaTypeConstraints(asset *Asset, mediaType MediaType) error {
	switch mediaType {
	case MediaTypeImage:
		// Images are timeless
		asset.Duration = "0s"
		asset.HasVideo = "1"
		asset.VideoSources = "1"
		// NO audio properties for images
		
	case MediaTypeVideo:
		// Videos have duration and video properties
		asset.HasVideo = "1"
		asset.VideoSources = "1"
		// Audio properties would be detected separately
		
	case MediaTypeAudio:
		// Audio only
		asset.HasAudio = "1"
		asset.AudioSources = "1"
		asset.AudioChannels = "2"  // Default stereo
		asset.AudioRate = "48000"  // Default 48kHz
		// NO video properties for audio
		
	default:
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}
	
	return nil
}

// ValidatedFormat represents a format with validation-aware construction
type ValidatedFormat struct {
	Format
	mediaType MediaType
	validator *StructValidator
}

// NewValidatedFormat creates a new format with full validation
func NewValidatedFormat(id ID, name string, width, height int, colorSpace ColorSpace, mediaType MediaType) (*ValidatedFormat, error) {
	// Validate inputs
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("invalid format ID: %v", err)
	}
	
	if err := colorSpace.Validate(); err != nil {
		return nil, fmt.Errorf("invalid color space: %v", err)
	}
	
	if err := mediaType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid media type: %v", err)
	}
	
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", width, height)
	}
	
	// Create basic format structure
	format := Format{
		ID:         string(id),
		Name:       name,
		Width:      fmt.Sprintf("%d", width),
		Height:     fmt.Sprintf("%d", height),
		ColorSpace: string(colorSpace),
	}
	
	// Apply media type constraints
	if err := applyFormatMediaTypeConstraints(&format, mediaType); err != nil {
		return nil, fmt.Errorf("media type constraint validation failed: %v", err)
	}
	
	// Create validated format
	validatedFormat := &ValidatedFormat{
		Format:    format,
		mediaType: mediaType,
		validator: NewStructValidator(),
	}
	
	// Final validation
	if err := validatedFormat.Validate(); err != nil {
		return nil, fmt.Errorf("format validation failed: %v", err)
	}
	
	return validatedFormat, nil
}

// NewValidatedVideoFormat creates a validated format specifically for video
func NewValidatedVideoFormat(id ID, width, height int, frameDuration Duration) (*ValidatedFormat, error) {
	// Validate frame duration
	if err := frameDuration.Validate(); err != nil {
		return nil, fmt.Errorf("invalid frame duration: %v", err)
	}
	
	if frameDuration == Duration("0s") {
		return nil, fmt.Errorf("video format cannot have 0s frame duration")
	}
	
	// Create video format with standard settings
	validatedFormat, err := NewValidatedFormat(
		id,
		generateVideoFormatName(width, height, frameDuration),
		width,
		height,
		ColorSpace("1-1-1 (Rec. 709)"),
		MediaTypeVideo,
	)
	if err != nil {
		return nil, err
	}
	
	// Set frame duration for video
	validatedFormat.Format.FrameDuration = string(frameDuration)
	
	return validatedFormat, nil
}

// NewValidatedImageFormat creates a validated format specifically for images
func NewValidatedImageFormat(id ID, width, height int) (*ValidatedFormat, error) {
	// Create image format (no frame duration)
	return NewValidatedFormat(
		id,
		"FFVideoFormatRateUndefined",
		width,
		height,
		ColorSpace("1-13-1"),
		MediaTypeImage,
	)
}

// Validate validates the format using comprehensive validation
func (vf *ValidatedFormat) Validate() error {
	// Validate using struct validator
	if err := vf.validator.ValidateStruct(&vf.Format); err != nil {
		return err
	}
	
	// Validate media type constraints
	constraintValidator := NewMediaConstraintValidator()
	return constraintValidator.ValidateFormat(&vf.Format, vf.mediaType)
}

// GetMediaType returns the media type
func (vf *ValidatedFormat) GetMediaType() MediaType {
	return vf.mediaType
}

// applyFormatMediaTypeConstraints applies media type constraints to a format
func applyFormatMediaTypeConstraints(format *Format, mediaType MediaType) error {
	switch mediaType {
	case MediaTypeImage:
		// Images CANNOT have frame duration
		format.FrameDuration = ""
		
	case MediaTypeVideo:
		// Videos MUST have frame duration (will be set by caller)
		// This is validated elsewhere
		
	case MediaTypeAudio:
		// Audio typically doesn't have format
		return fmt.Errorf("audio media type should not have format")
		
	default:
		return fmt.Errorf("unsupported media type for format: %s", mediaType)
	}
	
	return nil
}

// generateVideoFormatName generates a standard video format name
func generateVideoFormatName(width, height int, frameDuration Duration) string {
	// Convert frame duration to fps for naming
	seconds, err := frameDuration.ToSeconds()
	if err != nil {
		return "FFVideoFormatCustom"
	}
	
	fps := 1.0 / seconds
	
	// Generate name based on common resolutions and frame rates
	if width == 1920 && height == 1080 {
		if abs(fps-23.976) < 0.1 {
			return "FFVideoFormat1080p2398"
		} else if abs(fps-29.97) < 0.1 {
			return "FFVideoFormat1080p2997"
		} else if abs(fps-30.0) < 0.1 {
			return "FFVideoFormat1080p30"
		}
		return "FFVideoFormat1080pCustom"
	} else if width == 1280 && height == 720 {
		if abs(fps-23.976) < 0.1 {
			return "FFVideoFormat720p2398"
		} else if abs(fps-29.97) < 0.1 {
			return "FFVideoFormat720p2997"
		} else if abs(fps-30.0) < 0.1 {
			return "FFVideoFormat720p30"
		}
		return "FFVideoFormat720pCustom"
	} else if width == 1080 && height == 1920 {
		// Vertical format
		if abs(fps-23.976) < 0.1 {
			return "FFVideoFormat1080p2398_Vertical"
		}
		return "FFVideoFormat1080pVertical"
	}
	
	return fmt.Sprintf("FFVideoFormat%dx%d", width, height)
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ValidatedEffect represents an effect with validation-aware construction
type ValidatedEffect struct {
	Effect
	validator *StructValidator
}

// NewValidatedEffect creates a new effect with full validation
func NewValidatedEffect(id ID, name, uid string) (*ValidatedEffect, error) {
	// Validate inputs
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("invalid effect ID: %v", err)
	}
	
	if name == "" {
		return nil, fmt.Errorf("effect name cannot be empty")
	}
	
	if uid == "" {
		return nil, fmt.Errorf("effect UID cannot be empty")
	}
	
	// Validate effect UID against known effects
	if err := validateEffectUID(uid); err != nil {
		return nil, fmt.Errorf("invalid effect UID: %v", err)
	}
	
	// Create effect
	effect := Effect{
		ID:   string(id),
		Name: name,
		UID:  uid,
	}
	
	// Create validated effect
	validatedEffect := &ValidatedEffect{
		Effect:    effect,
		validator: NewStructValidator(),
	}
	
	// Final validation
	if err := validatedEffect.Validate(); err != nil {
		return nil, fmt.Errorf("effect validation failed: %v", err)
	}
	
	return validatedEffect, nil
}

// NewValidatedTextEffect creates a validated text effect
func NewValidatedTextEffect(id ID) (*ValidatedEffect, error) {
	return NewValidatedEffect(
		id,
		"Text",
		".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
	)
}

// Validate validates the effect using comprehensive validation
func (ve *ValidatedEffect) Validate() error {
	// Validate using struct validator
	return ve.validator.ValidateStruct(&ve.Effect)
}

// validateEffectUID validates known effect UIDs
func validateEffectUID(uid string) error {
	// Known working effect UIDs
	knownEffects := map[string]bool{
		".../Titles.localized/Basic Text.localized/Text.localized/Text.moti": true,
		"FFGaussianBlur":      true,
		"FFMotionBlur":        true,
		"FFColorCorrection":   true,
		"FFSuperEllipseMask":  true,
		// Add more known effects as needed
	}
	
	if !knownEffects[uid] {
		return fmt.Errorf("unknown effect UID: %s (use only verified effect UIDs)", uid)
	}
	
	return nil
}

// Note: generateUID function is already defined in ids.go

// AssetFormatPair represents a validated asset and format pair
type AssetFormatPair struct {
	Asset       *ValidatedAsset
	Format      *ValidatedFormat
	MediaType   MediaType
	validator   *MediaTypeSpecificValidator
}

// NewAssetFormatPair creates a validated asset and format pair
func NewAssetFormatPair(assetID, formatID ID, filePath string, duration Duration, width, height int) (*AssetFormatPair, error) {
	// Detect media type
	mediaType, err := DetectMediaTypeFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect media type: %v", err)
	}
	
	// Create validated asset
	asset, err := NewValidatedAssetFromPath(assetID, filePath, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %v", err)
	}
	
	// Create validated format based on media type
	var format *ValidatedFormat
	switch mediaType {
	case MediaTypeImage:
		format, err = NewValidatedImageFormat(formatID, width, height)
	case MediaTypeVideo:
		frameDuration := NewDurationFromSeconds(1.0 / 23.976) // Default to 23.976 fps
		format, err = NewValidatedVideoFormat(formatID, width, height, frameDuration)
	case MediaTypeAudio:
		format = nil // Audio doesn't use format
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create format: %v", err)
	}
	
	// Link asset to format
	if format != nil {
		asset.Asset.Format = string(formatID)
	}
	
	// Create pair
	pair := &AssetFormatPair{
		Asset:     asset,
		Format:    format,
		MediaType: mediaType,
		validator: NewMediaTypeSpecificValidator(),
	}
	
	// Validate the pair
	if err := pair.Validate(); err != nil {
		return nil, fmt.Errorf("asset-format pair validation failed: %v", err)
	}
	
	return pair, nil
}

// Validate validates the asset-format pair
func (afp *AssetFormatPair) Validate() error {
	var format *Format
	if afp.Format != nil {
		format = &afp.Format.Format
	}
	
	return afp.validator.ValidateAssetFormatPair(&afp.Asset.Asset, format, afp.MediaType)
}