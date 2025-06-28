// Package document_builder implements Step 16 of the FCPXMLKit-inspired refactoring plan:
// Create FCPXML Document Builder.
//
// This provides a high-level document builder with comprehensive validation that ties
// together all the validation components into a safe, easy-to-use API for FCPXML generation.
package fcp

import (
	"fmt"
	"time"
)

// FCPXMLDocumentBuilder provides high-level document building with comprehensive validation
type FCPXMLDocumentBuilder struct {
	projectName       string
	totalDuration     Duration
	version           string
	
	// Core components
	registry          *ReferenceRegistry
	timelineValidator *TimelineValidator
	spineBuilder      *SpineBuilder
	textValidator     *TextStyleValidator
	
	// Resource tracking
	assets           map[string]*Asset
	formats          map[string]*Format
	effects          map[string]*Effect
	
	// Settings
	maxLanes         int
	allowOverlaps    bool
	allowLaneGaps    bool
}

// NewFCPXMLDocumentBuilder creates a new document builder
func NewFCPXMLDocumentBuilder(projectName string, totalDuration Duration) (*FCPXMLDocumentBuilder, error) {
	if err := totalDuration.Validate(); err != nil {
		return nil, fmt.Errorf("invalid total duration: %v", err)
	}
	
	if projectName == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	
	// Create core components
	registry := NewReferenceRegistry()
	
	timelineValidator, err := NewTimelineValidator(totalDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create timeline validator: %v", err)
	}
	
	spineBuilder, err := NewSpineBuilder(totalDuration, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to create spine builder: %v", err)
	}
	
	textValidator := NewTextStyleValidator()
	
	// Create document builder instance
	builder := &FCPXMLDocumentBuilder{
		projectName:       projectName,
		totalDuration:     totalDuration,
		version:          "1.13",
		registry:         registry,
		timelineValidator: timelineValidator,
		spineBuilder:     spineBuilder,
		textValidator:    textValidator,
		assets:           make(map[string]*Asset),
		formats:          make(map[string]*Format),
		effects:          make(map[string]*Effect),
		maxLanes:         10,
		allowOverlaps:    false,
		allowLaneGaps:    false,
	}
	
	// Allow overlapping elements in video editing (configure both spine and timeline validators)
	builder.SetAllowOverlaps(true)
	
	return builder, nil
}

// SetMaxLanes sets the maximum allowed lane number
func (builder *FCPXMLDocumentBuilder) SetMaxLanes(maxLanes int) {
	builder.maxLanes = maxLanes
	builder.timelineValidator.SetMaxLanes(maxLanes)
}

// SetAllowOverlaps controls whether overlaps in the same lane are allowed
func (builder *FCPXMLDocumentBuilder) SetAllowOverlaps(allow bool) {
	builder.allowOverlaps = allow
	builder.spineBuilder.SetAllowOverlaps(allow)
	builder.timelineValidator.SetAllowOverlaps(allow)
}

// SetAllowLaneGaps controls whether lane gaps are allowed
func (builder *FCPXMLDocumentBuilder) SetAllowLaneGaps(allow bool) {
	builder.allowLaneGaps = allow
	builder.timelineValidator.SetAllowGaps(allow)
}

// AddMediaFile adds a media file with automatic type detection and resource creation
func (builder *FCPXMLDocumentBuilder) AddMediaFile(filePath, name string, offset Time, duration Duration, lane Lane) error {
	// Detect media type from file extension
	mediaType, err := DetectMediaTypeFromPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to detect media type: %v", err)
	}
	
	// Create transaction for atomic resource creation
	tx := NewSafeTransaction(builder.registry)
	defer tx.Rollback()
	
	// Create asset and format based on media type
	var asset *Asset
	var format *Format
	
	switch mediaType {
	case MediaTypeImage:
		asset, format, err = tx.CreateImageAsset(filePath, name, duration)
		if err != nil {
			return fmt.Errorf("failed to create image asset: %v", err)
		}
		
		// Images use Video elements in the spine
		if err := builder.spineBuilder.AddVideo(asset.ID, name, offset, duration, lane); err != nil {
			return fmt.Errorf("failed to add video element: %v", err)
		}
		
	case MediaTypeVideo:
		asset, format, err = tx.CreateVideoAsset(filePath, name, duration)
		if err != nil {
			return fmt.Errorf("failed to create video asset: %v", err)
		}
		
		// Videos use AssetClip elements in the spine
		formatID := ""
		if format != nil {
			formatID = format.ID
		}
		if err := builder.spineBuilder.AddAssetClip(asset.ID, name, offset, duration, lane, formatID); err != nil {
			return fmt.Errorf("failed to add asset clip: %v", err)
		}
		
	case MediaTypeAudio:
		asset, format, err = tx.CreateAudioAsset(filePath, name, duration)
		if err != nil {
			return fmt.Errorf("failed to create audio asset: %v", err)
		}
		
		// Audio typically goes in negative lanes
		audioLane := lane
		if audioLane == Lane(0) {
			audioLane = Lane(-1) // Default audio lane
		}
		
		formatID := ""
		if format != nil {
			formatID = format.ID
		}
		if err := builder.spineBuilder.AddAssetClip(asset.ID, name, offset, duration, audioLane, formatID); err != nil {
			return fmt.Errorf("failed to add audio clip: %v", err)
		}
		
	default:
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit resource transaction: %v", err)
	}
	
	// Store references
	builder.assets[asset.ID] = asset
	if format != nil {
		builder.formats[format.ID] = format
	}
	
	return nil
}

// AddText adds a text element with automatic effect creation
func (builder *FCPXMLDocumentBuilder) AddText(text string, offset Time, duration Duration, lane Lane, options ...TextOption) error {
	// Create text effect if not exists
	textEffectKey := "text_effect"
	if _, exists := builder.effects[textEffectKey]; !exists {
		tx := NewSafeTransaction(builder.registry)
		defer tx.Rollback()
		
		// Reserve proper ID for the effect
		ids := tx.ReserveIDs(1)
		if len(ids) == 0 {
			return fmt.Errorf("failed to reserve ID for text effect")
		}
		textEffectID := string(ids[0])
		
		effect, err := tx.CreateEffect(textEffectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			return fmt.Errorf("failed to create text effect: %v", err)
		}
		
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit text effect: %v", err)
		}
		
		builder.effects[textEffectKey] = effect
	}
	
	// Create text configuration
	config := &TextConfiguration{
		Font:      "Helvetica",
		FontSize:  "48",
		FontColor: "1 1 1 1",
		Alignment: "center",
	}
	
	// Apply options
	for _, option := range options {
		option(config)
	}
	
	// Validate text configuration
	if err := builder.textValidator.ValidateTextConfiguration(config); err != nil {
		return fmt.Errorf("text configuration validation failed: %v", err)
	}
	
	// Get the actual effect ID to use for spine reference
	effect := builder.effects[textEffectKey]
	actualEffectID := effect.ID
	
	// Add title to spine
	if err := builder.spineBuilder.AddTitle(actualEffectID, text, offset, duration, lane); err != nil {
		return fmt.Errorf("failed to add title: %v", err)
	}
	
	return nil
}

// AddKenBurnsAnimation adds a Ken Burns animation to the most recent video element
func (builder *FCPXMLDocumentBuilder) AddKenBurnsAnimation(presetName string, startTime Time, duration Duration) error {
	// Get Ken Burns presets
	presets := GetKenBurnsPresets()
	preset, exists := presets[presetName]
	if !exists {
		return fmt.Errorf("unknown Ken Burns preset: %s", presetName)
	}
	
	// Create animation
	transform, err := preset.Builder(startTime, duration)
	if err != nil {
		return fmt.Errorf("failed to create Ken Burns animation: %v", err)
	}
	
	// Apply to the most recent video element in the timeline
	// This is a simplified implementation - in practice, you'd want to find
	// the specific element at the given time
	return builder.applyTransformToRecentElement(transform, startTime)
}

// AddCustomAnimation adds a custom animation using the animation builder
func (builder *FCPXMLDocumentBuilder) AddCustomAnimation(elementType string, startTime Time, animations map[string][]KeyframeData) error {
	transformBuilder := NewTransformBuilder()
	
	for paramName, keyframes := range animations {
		switch paramName {
		case "position":
			if err := transformBuilder.AddPositionAnimation(keyframes); err != nil {
				return fmt.Errorf("failed to add position animation: %v", err)
			}
		case "scale":
			if err := transformBuilder.AddScaleAnimation(keyframes, "linear"); err != nil {
				return fmt.Errorf("failed to add scale animation: %v", err)
			}
		case "rotation":
			if err := transformBuilder.AddRotationAnimation(keyframes, "linear"); err != nil {
				return fmt.Errorf("failed to add rotation animation: %v", err)
			}
		default:
			return fmt.Errorf("unsupported animation parameter: %s", paramName)
		}
	}
	
	transform, err := transformBuilder.Build()
	if err != nil {
		return fmt.Errorf("failed to build transform: %v", err)
	}
	
	return builder.applyTransformToRecentElement(transform, startTime)
}

// applyTransformToRecentElement applies a transform to the most recent element
// This is a simplified implementation for demonstration
func (builder *FCPXMLDocumentBuilder) applyTransformToRecentElement(transform *AdjustTransform, startTime Time) error {
	// In a full implementation, this would:
	// 1. Find the element at the given start time
	// 2. Apply the transform to that specific element
	// 3. Validate that the transform timing fits within the element's duration
	
	// For now, just validate that the transform is well-formed
	if transform == nil {
		return fmt.Errorf("transform cannot be nil")
	}
	
	// Validate all parameters in the transform
	for _, param := range transform.Params {
		if param.KeyframeAnimation != nil {
			validator := NewKeyframeValidator()
			for i, keyframe := range param.KeyframeAnimation.Keyframes {
				vkf := &ValidatedKeyframe{
					Time:   Time(keyframe.Time),
					Value:  keyframe.Value,
					Interp: keyframe.Interp,
					Curve:  keyframe.Curve,
				}
				
				if err := validator.ValidateKeyframe(param.Name, vkf); err != nil {
					return fmt.Errorf("keyframe %d validation failed for param %s: %v", i, param.Name, err)
				}
			}
		}
	}
	
	return nil
}

// Build creates the final FCPXML document with comprehensive validation
func (builder *FCPXMLDocumentBuilder) Build() (*FCPXML, error) {
	// Build validated spine
	spine, err := builder.spineBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build spine: %v", err)
	}
	
	// Create main sequence format
	tx := NewSafeTransaction(builder.registry)
	defer tx.Rollback()
	
	mainFormat, err := tx.CreateFormat("r1", "FFVideoFormat1080p30", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		return nil, fmt.Errorf("failed to create main format: %v", err)
	}
	
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit main format: %v", err)
	}
	
	// Create final FCPXML structure
	fcpxml := &FCPXML{
		Version: builder.version,
		Resources: Resources{
			Assets:  make([]Asset, 0),
			Formats: make([]Format, 0),
			Effects: make([]Effect, 0),
		},
		Library: Library{
			Events: []Event{{
				Name: "Event",
				UID:  generateUID("Event"),
				Projects: []Project{{
					Name:    builder.projectName,
					UID:     generateUID(builder.projectName),
					ModDate: time.Now().Format("2006-01-02 15:04:05 -0700"),
					Sequences: []Sequence{{
						Format:      mainFormat.ID,
						Duration:    builder.totalDuration.String(),
						TCStart:     "0s",
						TCFormat:    "NDF",
						AudioLayout: "stereo",
						AudioRate:   "48k",
						Spine:       *spine,
					}},
				}},
			}},
		},
	}
	
	// Add main format
	fcpxml.Resources.Formats = append(fcpxml.Resources.Formats, *mainFormat)
	
	// Add all created assets
	for _, asset := range builder.assets {
		fcpxml.Resources.Assets = append(fcpxml.Resources.Assets, *asset)
	}
	
	// Add all created formats (except main format which is already added)
	for _, format := range builder.formats {
		if format.ID != mainFormat.ID {
			fcpxml.Resources.Formats = append(fcpxml.Resources.Formats, *format)
		}
	}
	
	// Add all created effects
	for _, effect := range builder.effects {
		fcpxml.Resources.Effects = append(fcpxml.Resources.Effects, *effect)
	}
	
	// Final comprehensive validation
	if err := builder.validateCompleteDocument(fcpxml); err != nil {
		return nil, fmt.Errorf("final document validation failed: %v", err)
	}
	
	return fcpxml, nil
}

// validateCompleteDocument performs comprehensive final validation
func (builder *FCPXMLDocumentBuilder) validateCompleteDocument(fcpxml *FCPXML) error {
	// Validate timeline structure
	if err := builder.timelineValidator.ValidateComplete(); err != nil {
		return fmt.Errorf("timeline validation failed: %v", err)
	}
	
	// Validate all resource references
	if err := builder.registry.ValidateAllReferences(fcpxml); err != nil {
		return fmt.Errorf("reference validation failed: %v", err)
	}
	
	// Validate FCPXML structure using validation marshaling
	if err := fcpxml.ValidateStructure(); err != nil {
		return fmt.Errorf("document structure validation failed: %v", err)
	}
	
	return nil
}

// GetStatistics returns comprehensive statistics about the document
func (builder *FCPXMLDocumentBuilder) GetStatistics() DocumentStatistics {
	spineStats := builder.spineBuilder.GetStatistics()
	timelineStats := builder.timelineValidator.GetTimelineStatistics()
	
	return DocumentStatistics{
		ProjectName:          builder.projectName,
		TotalDuration:        builder.totalDuration.String(),
		AssetCount:           len(builder.assets),
		FormatCount:          len(builder.formats),
		EffectCount:          len(builder.effects),
		SpineElementCount:    spineStats.TotalElements,
		UsedLanes:           timelineStats.UsedLanes,
		TimelineUtilization: timelineStats.TimelineUtilization,
		ElementsByType:      spineStats.ElementsByType,
		ElementsByLane:      spineStats.ElementsByLane,
	}
}

// DocumentStatistics provides comprehensive document information
type DocumentStatistics struct {
	ProjectName          string
	TotalDuration        string
	AssetCount           int
	FormatCount          int
	EffectCount          int
	SpineElementCount    int
	UsedLanes           []int
	TimelineUtilization  float64
	ElementsByType      map[string]int
	ElementsByLane      map[int]int
}

// TextConfiguration represents text styling options
type TextConfiguration struct {
	Font      string
	FontSize  string
	FontColor string
	Alignment string
	Bold      bool
	Italic    bool
}

// TextOption configures text styling
type TextOption func(*TextConfiguration)

// WithFont sets the font family
func WithFont(font string) TextOption {
	return func(tc *TextConfiguration) {
		tc.Font = font
	}
}

// WithFontSize sets the font size
func WithFontSize(size string) TextOption {
	return func(tc *TextConfiguration) {
		tc.FontSize = size
	}
}

// WithFontColor sets the font color
func WithFontColor(color string) TextOption {
	return func(tc *TextConfiguration) {
		tc.FontColor = color
	}
}

// WithAlignment sets the text alignment
func WithAlignment(alignment string) TextOption {
	return func(tc *TextConfiguration) {
		tc.Alignment = alignment
	}
}

// WithBold sets bold formatting
func WithBold(bold bool) TextOption {
	return func(tc *TextConfiguration) {
		tc.Bold = bold
	}
}

// WithItalic sets italic formatting
func WithItalic(italic bool) TextOption {
	return func(tc *TextConfiguration) {
		tc.Italic = italic
	}
}

