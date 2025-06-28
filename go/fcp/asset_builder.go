package fcp

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// AssetBuilder provides validation-aware asset creation
// This implements Step 12 of the FCPXMLKit-inspired refactoring plan:
// Replace direct struct assignment with validation-aware constructors
type AssetBuilder struct {
	registry  *ReferenceRegistry
	mediaType string
	validator *StructValidator
}

// NewAssetBuilder creates a new asset builder for the specified media type
func NewAssetBuilder(registry *ReferenceRegistry, mediaType string) *AssetBuilder {
	return &AssetBuilder{
		registry:  registry,
		mediaType: mediaType,
		validator: NewStructValidator(),
	}
}

// CreateAsset creates a validated asset with appropriate properties based on media type
func (ab *AssetBuilder) CreateAsset(id ID, filePath, name string, duration Duration) (*Asset, *Format, error) {
	// Validate inputs
	if err := id.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid asset ID: %v", err)
	}

	if err := duration.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid duration: %v", err)
	}

	// Apply media type constraints
	switch ab.mediaType {
	case "image":
		if duration != "0s" {
			return nil, nil, fmt.Errorf("image assets must have duration='0s', got: %s", duration)
		}
	case "video", "audio":
		if duration == "0s" {
			return nil, nil, fmt.Errorf("%s assets cannot have duration='0s'", ab.mediaType)
		}
	default:
		return nil, nil, fmt.Errorf("unsupported media type: %s", ab.mediaType)
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Create asset with appropriate properties
	asset := &Asset{
		ID:       string(id),
		Name:     name,
		UID:      generateUID(absPath),
		Start:    "0s",
		Duration: string(duration),
		MediaRep: MediaRep{
			Kind: "original-media",
			Sig:  generateUID(absPath),
			Src:  "file://" + absPath,
		},
	}

	// Set media-type specific properties
	switch ab.mediaType {
	case "image":
		asset.HasVideo = "1"
		asset.VideoSources = "1"
		// Note: NO HasAudio, AudioSources for images

	case "video":
		asset.HasVideo = "1"
		asset.VideoSources = "1"
		// Note: Audio properties are NOT set by default for basic video creation
		// Use CreateVideoAssetWithDetection for audio detection

	case "audio":
		asset.HasAudio = "1"
		asset.AudioSources = "1"
		asset.AudioChannels = "2"
		asset.AudioRate = "48000"
		// Note: NO HasVideo, VideoSources for audio
	}

	// Validate final asset
	if err := ab.validateAssetMediaType(asset, ab.mediaType); err != nil {
		return nil, nil, fmt.Errorf("asset media type validation failed: %v", err)
	}

	if err := ab.validator.validateAssetStructure(asset); err != nil {
		return nil, nil, fmt.Errorf("asset struct validation failed: %v", err)
	}

	// Create appropriate format
	format, err := ab.createFormat(id, ab.mediaType)
	if err != nil {
		return nil, nil, fmt.Errorf("format creation failed: %v", err)
	}

	// Set format reference in asset
	if format != nil {
		asset.Format = format.ID
	}

	return asset, format, nil
}

// createFormat creates an appropriate format for the media type
func (ab *AssetBuilder) createFormat(assetID ID, mediaType string) (*Format, error) {
	formatID := ID(string(assetID) + "_format")

	format := &Format{
		ID:         string(formatID),
		Width:      "1920",
		Height:     "1080",
		ColorSpace: "1-1-1 (Rec. 709)",
	}

	switch mediaType {
	case "image":
		format.Name = "FFVideoFormatRateUndefined"
		// CRITICAL: NO frameDuration for images

	case "video":
		format.Name = "FFVideoFormat1080p30"
		format.FrameDuration = "1001/24000s" // Standard 23.976fps in FCP timebase

	case "audio":
		// Audio typically doesn't need a format, or has different format structure
		return nil, nil // Return nil format for audio
	}

	// Validate format
	if err := ab.validateFormatMediaType(format, mediaType); err != nil {
		return nil, fmt.Errorf("format validation failed: %v", err)
	}

	return format, nil
}

// validateAssetMediaType validates asset properties against media type constraints
func (ab *AssetBuilder) validateAssetMediaType(asset *Asset, mediaType string) error {
	switch mediaType {
	case "image":
		// Images must have duration="0s" and video properties only
		if asset.Duration != "0s" {
			return fmt.Errorf("image assets must have duration='0s', got: %s", asset.Duration)
		}
		if asset.HasVideo != "1" {
			return fmt.Errorf("image assets must have hasVideo='1'")
		}
		if asset.HasAudio != "" {
			return fmt.Errorf("image assets cannot have hasAudio attribute")
		}
		if asset.AudioSources != "" {
			return fmt.Errorf("image assets cannot have audioSources attribute")
		}

	case "video":
		// Videos must have duration and video properties
		if asset.Duration == "0s" {
			return fmt.Errorf("video assets cannot have duration='0s'")
		}
		if asset.HasVideo != "1" {
			return fmt.Errorf("video assets must have hasVideo='1'")
		}
		if asset.VideoSources != "1" {
			return fmt.Errorf("video assets must have videoSources='1'")
		}

	case "audio":
		// Audio must have duration and audio properties only
		if asset.Duration == "0s" {
			return fmt.Errorf("audio assets cannot have duration='0s'")
		}
		if asset.HasAudio != "1" {
			return fmt.Errorf("audio assets must have hasAudio='1'")
		}
		if asset.HasVideo != "" {
			return fmt.Errorf("audio assets cannot have hasVideo attribute")
		}
		if asset.VideoSources != "" {
			return fmt.Errorf("audio assets cannot have videoSources attribute")
		}
	}

	return nil
}

// validateFormatMediaType validates format properties against media type constraints
func (ab *AssetBuilder) validateFormatMediaType(format *Format, mediaType string) error {
	switch mediaType {
	case "image":
		// Images cannot have frameDuration
		if format.FrameDuration != "" {
			return fmt.Errorf("image formats cannot have frameDuration")
		}
		if format.Width == "" || format.Height == "" {
			return fmt.Errorf("image formats must have width and height")
		}

	case "video":
		// Videos must have frameDuration
		if format.FrameDuration == "" {
			return fmt.Errorf("video formats must have frameDuration")
		}
		if format.Width == "" || format.Height == "" {
			return fmt.Errorf("video formats must have width and height")
		}
	}

	return nil
}

// CreateImageAsset creates a validated image asset
func (ab *AssetBuilder) CreateImageAsset(id ID, imagePath, name string, displayDuration Duration) (*Asset, *Format, error) {
	// Images are timeless (asset duration = "0s") but have display duration for timeline
	asset, format, err := ab.CreateAsset(id, imagePath, name, Duration("0s"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create image asset: %v", err)
	}

	// Add image-specific metadata if needed
	asset.Metadata = &Metadata{
		// Could add image-specific metadata here
	}

	return asset, format, nil
}

// CreateVideoAsset creates a validated video asset
func (ab *AssetBuilder) CreateVideoAsset(id ID, videoPath, name string, duration Duration) (*Asset, *Format, error) {
	asset, format, err := ab.CreateAsset(id, videoPath, name, duration)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create video asset: %v", err)
	}

	// Video-specific enhancements could be added here
	// e.g., audio track detection, frame rate detection

	return asset, format, nil
}

// CreateAudioAsset creates a validated audio asset
func (ab *AssetBuilder) CreateAudioAsset(id ID, audioPath, name string, duration Duration) (*Asset, error) {
	// Audio assets typically don't need formats
	ab.mediaType = "audio"
	asset, _, err := ab.CreateAsset(id, audioPath, name, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio asset: %v", err)
	}

	return asset, nil
}

// CreateVideoAssetWithDetection creates a video asset with automatic property detection
func (ab *AssetBuilder) CreateVideoAssetWithDetection(id ID, videoPath, name string, duration Duration) (*Asset, *Format, error) {
	asset, format, err := ab.CreateVideoAsset(id, videoPath, name, duration)
	if err != nil {
		return nil, nil, err
	}

	// Enhanced video property detection could be added here
	// For now, use basic video properties

	// Detect if video has audio (simplified version)
	// For testing, use extension-based detection since ffprobe may not be available
	ext := filepath.Ext(videoPath)
	if ext == ".mp4" || ext == ".mov" || ext == ".avi" {
		asset.HasAudio = "1"
		asset.AudioSources = "1"
		asset.AudioChannels = "2"
		asset.AudioRate = "48000"
	}

	// Re-validate after property detection
	if err := ab.validateAssetMediaType(asset, "video"); err != nil {
		return nil, nil, fmt.Errorf("asset validation failed after detection: %v", err)
	}

	return asset, format, nil
}

// Note: hasAudioTrack function is defined in transaction.go

// FormatBuilder provides validation-aware format creation
type FormatBuilder struct {
	validator *StructValidator
}

// NewFormatBuilder creates a new format builder
func NewFormatBuilder() *FormatBuilder {
	return &FormatBuilder{
		validator: NewStructValidator(),
	}
}

// CreateFormat creates a validated format
func (fb *FormatBuilder) CreateFormat(id ID, name string, width, height int, colorSpace string, mediaType string) (*Format, error) {
	format := &Format{
		ID:         string(id),
		Name:       name,
		Width:      strconv.Itoa(width),
		Height:     strconv.Itoa(height),
		ColorSpace: colorSpace,
	}

	// Apply media type constraints
	switch mediaType {
	case "image":
		// Images don't have frame duration
		break
	case "video":
		// Videos must have frame duration
		format.FrameDuration = "1001/24000s" // Standard FCP 23.976fps
	default:
		return nil, fmt.Errorf("unsupported format media type: %s", mediaType)
	}

	// Validate format
	if err := fb.validator.validateFormatStructure(format); err != nil {
		return nil, fmt.Errorf("format validation failed: %v", err)
	}

	return format, nil
}

// EffectBuilder provides validation-aware effect creation
type EffectBuilder struct {
	validator *StructValidator
}

// NewEffectBuilder creates a new effect builder
func NewEffectBuilder() *EffectBuilder {
	return &EffectBuilder{
		validator: NewStructValidator(),
	}
}

// CreateEffect creates a validated effect
func (eb *EffectBuilder) CreateEffect(id ID, name, uid string) (*Effect, error) {
	// Validate effect UID against known effects
	if err := eb.validateEffectUID(uid); err != nil {
		return nil, fmt.Errorf("invalid effect UID: %v", err)
	}

	effect := &Effect{
		ID:   string(id),
		Name: name,
		UID:  uid,
	}

	// Validate effect structure
	if err := eb.validator.validateEffectStructure(effect); err != nil {
		return nil, fmt.Errorf("effect validation failed: %v", err)
	}

	return effect, nil
}

// validateEffectUID validates that the effect UID is a known/safe effect
func (eb *EffectBuilder) validateEffectUID(uid string) error {
	// Known safe effect UIDs
	knownEffects := map[string]bool{
		".../Titles.localized/Basic Text.localized/Text.localized/Text.moti": true,
		"FFGaussianBlur":      true,
		"FFMotionBlur":        true,
		"FFColorCorrection":   true,
		"FFSuperEllipseMask":  true,
		// Add more known effects as needed
	}

	if !knownEffects[uid] {
		return fmt.Errorf("unknown effect UID: %s", uid)
	}

	return nil
}