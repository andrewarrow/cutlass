// Package safe_operations implements Step 17 of the FCPXMLKit-inspired refactoring plan:
// Safe High-Level Operations with built-in validation.
//
// This provides safe wrapper functions that offer high-level FCPXML operations
// with comprehensive validation, error prevention, and automatic resource management.
// These operations build on the validation infrastructure from Steps 14-16.
package fcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SafeFCPXMLOperations provides high-level operations with validation
type SafeFCPXMLOperations struct {
	builder *FCPXMLDocumentBuilder
}

// NewSafeFCPXMLOperations creates a new safe operations instance
func NewSafeFCPXMLOperations(projectName string, durationSeconds float64) (*SafeFCPXMLOperations, error) {
	if projectName == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	
	if durationSeconds <= 0 {
		return nil, fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	// Convert to frame-accurate duration
	duration := NewDurationFromSeconds(durationSeconds)
	
	builder, err := NewFCPXMLDocumentBuilder(projectName, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create document builder: %v", err)
	}
	
	return &SafeFCPXMLOperations{
		builder: builder,
	}, nil
}

// AddBackgroundVideo adds a background video to the main timeline (lane 0)
func (ops *SafeFCPXMLOperations) AddBackgroundVideo(videoPath string, durationSeconds float64) error {
	if err := ops.validateMediaPath(videoPath); err != nil {
		return fmt.Errorf("invalid video path: %v", err)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	duration := NewDurationFromSeconds(durationSeconds)
	offset := NewTimeFromSeconds(0) // Start at beginning
	lane := Lane(0)                 // Main timeline
	
	name := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	if name == "" {
		name = "Background_Video"
	}
	
	return ops.builder.AddMediaFile(videoPath, name, offset, duration, lane)
}

// AddImageOverlay adds an image overlay at the specified time and lane
func (ops *SafeFCPXMLOperations) AddImageOverlay(imagePath string, startSeconds, durationSeconds float64, lane int) error {
	if err := ops.validateMediaPath(imagePath); err != nil {
		return fmt.Errorf("invalid image path: %v", err)
	}
	
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	if err := ops.validateLane(lane); err != nil {
		return fmt.Errorf("invalid lane: %v", err)
	}
	
	offset := NewTimeFromSeconds(startSeconds)
	duration := NewDurationFromSeconds(durationSeconds)
	laneObj := Lane(lane)
	
	name := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))
	if name == "" {
		name = fmt.Sprintf("Image_Lane_%d", lane)
	}
	
	return ops.builder.AddMediaFile(imagePath, name, offset, duration, laneObj)
}

// AddVideoOverlay adds a video overlay at the specified time and lane
func (ops *SafeFCPXMLOperations) AddVideoOverlay(videoPath string, startSeconds, durationSeconds float64, lane int) error {
	if err := ops.validateMediaPath(videoPath); err != nil {
		return fmt.Errorf("invalid video path: %v", err)
	}
	
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	if err := ops.validateLane(lane); err != nil {
		return fmt.Errorf("invalid lane: %v", err)
	}
	
	offset := NewTimeFromSeconds(startSeconds)
	duration := NewDurationFromSeconds(durationSeconds)
	laneObj := Lane(lane)
	
	name := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	if name == "" {
		name = fmt.Sprintf("Video_Lane_%d", lane)
	}
	
	return ops.builder.AddMediaFile(videoPath, name, offset, duration, laneObj)
}

// AddAudioTrack adds an audio track (typically in negative lanes)
func (ops *SafeFCPXMLOperations) AddAudioTrack(audioPath string, startSeconds, durationSeconds float64, audioLane int) error {
	if err := ops.validateMediaPath(audioPath); err != nil {
		return fmt.Errorf("invalid audio path: %v", err)
	}
	
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	// Audio lanes are typically negative (-1, -2, etc.)
	if audioLane >= 0 {
		audioLane = -1 // Default audio lane
	}
	
	if err := ops.validateLane(audioLane); err != nil {
		return fmt.Errorf("invalid audio lane: %v", err)
	}
	
	offset := NewTimeFromSeconds(startSeconds)
	duration := NewDurationFromSeconds(durationSeconds)
	laneObj := Lane(audioLane)
	
	name := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	if name == "" {
		name = fmt.Sprintf("Audio_Lane_%d", audioLane)
	}
	
	return ops.builder.AddMediaFile(audioPath, name, offset, duration, laneObj)
}

// AddTitleCard adds a title card with specified styling
func (ops *SafeFCPXMLOperations) AddTitleCard(text string, startSeconds, durationSeconds float64, lane int, options ...TitleCardOption) error {
	if text == "" {
		return fmt.Errorf("title text cannot be empty")
	}
	
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	if err := ops.validateLane(lane); err != nil {
		return fmt.Errorf("invalid lane: %v", err)
	}
	
	offset := NewTimeFromSeconds(startSeconds)
	duration := NewDurationFromSeconds(durationSeconds)
	laneObj := Lane(lane)
	
	// Create default title card configuration
	config := &TitleCardConfiguration{
		Font:      "Helvetica",
		FontSize:  "48",
		FontColor: "1 1 1 1", // White
		Alignment: "center",
		Bold:      false,
		Italic:    false,
	}
	
	// Apply options
	for _, option := range options {
		option(config)
	}
	
	// Validate configuration
	if err := ops.validateTitleCardConfiguration(config); err != nil {
		return fmt.Errorf("title card configuration invalid: %v", err)
	}
	
	// Convert configuration to text options
	textOptions := []TextOption{
		WithFont(config.Font),
		WithFontSize(config.FontSize),
		WithFontColor(config.FontColor),
		WithAlignment(config.Alignment),
		WithBold(config.Bold),
		WithItalic(config.Italic),
	}
	
	return ops.builder.AddText(text, offset, duration, laneObj, textOptions...)
}

// AddKenBurnsEffect adds a Ken Burns animation to the most recent image/video
func (ops *SafeFCPXMLOperations) AddKenBurnsEffect(startSeconds float64, presetName ...string) error {
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	// Default to "zoom_in" preset if none specified
	preset := "zoom_in"
	if len(presetName) > 0 && presetName[0] != "" {
		preset = presetName[0]
	}
	
	// Validate preset exists
	presets := GetKenBurnsPresets()
	if _, exists := presets[preset]; !exists {
		return fmt.Errorf("unknown Ken Burns preset: %s", preset)
	}
	
	startTime := NewTimeFromSeconds(startSeconds)
	duration := NewDurationFromSeconds(3.0) // Default 3-second animation
	
	return ops.builder.AddKenBurnsAnimation(preset, startTime, duration)
}

// AddPanAnimation adds a custom pan animation
func (ops *SafeFCPXMLOperations) AddPanAnimation(startSeconds, durationSeconds float64, fromX, fromY, toX, toY float64) error {
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	// Validate position values (reasonable range for FCPXML)
	if err := ops.validatePosition(fromX, fromY); err != nil {
		return fmt.Errorf("invalid start position: %v", err)
	}
	
	if err := ops.validatePosition(toX, toY); err != nil {
		return fmt.Errorf("invalid end position: %v", err)
	}
	
	startTime := NewTimeFromSeconds(startSeconds)
	endTime := NewTimeFromSeconds(startSeconds + durationSeconds)
	
	// Create keyframe data
	keyframes := []KeyframeData{
		{
			Time:  startTime,
			Value: fmt.Sprintf("%.1f %.1f", fromX, fromY),
		},
		{
			Time:  endTime,
			Value: fmt.Sprintf("%.1f %.1f", toX, toY),
		},
	}
	
	animations := map[string][]KeyframeData{
		"position": keyframes,
	}
	
	return ops.builder.AddCustomAnimation("video", startTime, animations)
}

// AddZoomAnimation adds a custom zoom animation
func (ops *SafeFCPXMLOperations) AddZoomAnimation(startSeconds, durationSeconds float64, fromScale, toScale float64) error {
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	if err := ops.validateScale(fromScale); err != nil {
		return fmt.Errorf("invalid start scale: %v", err)
	}
	
	if err := ops.validateScale(toScale); err != nil {
		return fmt.Errorf("invalid end scale: %v", err)
	}
	
	startTime := NewTimeFromSeconds(startSeconds)
	endTime := NewTimeFromSeconds(startSeconds + durationSeconds)
	
	// Create keyframe data
	keyframes := []KeyframeData{
		{
			Time:  startTime,
			Value: fmt.Sprintf("%.3f %.3f", fromScale, fromScale),
		},
		{
			Time:  endTime,
			Value: fmt.Sprintf("%.3f %.3f", toScale, toScale),
		},
	}
	
	animations := map[string][]KeyframeData{
		"scale": keyframes,
	}
	
	return ops.builder.AddCustomAnimation("video", startTime, animations)
}

// AddRotationAnimation adds a rotation animation
func (ops *SafeFCPXMLOperations) AddRotationAnimation(startSeconds, durationSeconds float64, fromAngle, toAngle float64) error {
	if startSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", startSeconds)
	}
	
	if durationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", durationSeconds)
	}
	
	if err := ops.validateRotation(fromAngle); err != nil {
		return fmt.Errorf("invalid start angle: %v", err)
	}
	
	if err := ops.validateRotation(toAngle); err != nil {
		return fmt.Errorf("invalid end angle: %v", err)
	}
	
	startTime := NewTimeFromSeconds(startSeconds)
	endTime := NewTimeFromSeconds(startSeconds + durationSeconds)
	
	// Create keyframe data
	keyframes := []KeyframeData{
		{
			Time:  startTime,
			Value: fmt.Sprintf("%.1f", fromAngle),
		},
		{
			Time:  endTime,
			Value: fmt.Sprintf("%.1f", toAngle),
		},
	}
	
	animations := map[string][]KeyframeData{
		"rotation": keyframes,
	}
	
	return ops.builder.AddCustomAnimation("video", startTime, animations)
}

// AddMultiLayerComposite adds multiple media files as a composite
func (ops *SafeFCPXMLOperations) AddMultiLayerComposite(mediaFiles []MediaLayerSpec) error {
	if len(mediaFiles) == 0 {
		return fmt.Errorf("at least one media file is required")
	}
	
	if len(mediaFiles) > 10 {
		return fmt.Errorf("too many layers (max 10): %d", len(mediaFiles))
	}
	
	// Validate all specs first
	for i, spec := range mediaFiles {
		if err := ops.validateMediaLayerSpec(spec); err != nil {
			return fmt.Errorf("media layer %d invalid: %v", i, err)
		}
	}
	
	// Add each layer
	for i, spec := range mediaFiles {
		var err error
		
		switch spec.Type {
		case "video":
			err = ops.AddVideoOverlay(spec.Path, spec.StartSeconds, spec.DurationSeconds, spec.Lane)
		case "image":
			err = ops.AddImageOverlay(spec.Path, spec.StartSeconds, spec.DurationSeconds, spec.Lane)
		case "audio":
			err = ops.AddAudioTrack(spec.Path, spec.StartSeconds, spec.DurationSeconds, spec.Lane)
		default:
			err = fmt.Errorf("unsupported media type: %s", spec.Type)
		}
		
		if err != nil {
			return fmt.Errorf("failed to add layer %d: %v", i, err)
		}
		
		// Add animations if specified
		if len(spec.Animations) > 0 {
			startTime := NewTimeFromSeconds(spec.StartSeconds)
			if err := ops.builder.AddCustomAnimation("video", startTime, spec.Animations); err != nil {
				return fmt.Errorf("failed to add animations for layer %d: %v", i, err)
			}
		}
	}
	
	return nil
}

// SetDocumentSettings configures document-level settings
func (ops *SafeFCPXMLOperations) SetDocumentSettings(settings DocumentSettings) error {
	if err := ops.validateDocumentSettings(settings); err != nil {
		return fmt.Errorf("invalid document settings: %v", err)
	}
	
	ops.builder.SetMaxLanes(settings.MaxLanes)
	ops.builder.SetAllowOverlaps(settings.AllowOverlaps)
	ops.builder.SetAllowLaneGaps(settings.AllowLaneGaps)
	
	return nil
}

// GenerateAndValidate creates the final FCPXML with comprehensive validation
func (ops *SafeFCPXMLOperations) GenerateAndValidate() ([]byte, error) {
	// Build FCPXML document
	fcpxml, err := ops.builder.Build()
	if err != nil {
		return nil, fmt.Errorf("document build failed: %v", err)
	}
	
	// Generate XML with validation
	data, err := fcpxml.ValidateAndMarshal()
	if err != nil {
		return nil, fmt.Errorf("XML generation failed: %v", err)
	}
	
	return data, nil
}

// SaveToFile saves the FCPXML to a file with validation
func (ops *SafeFCPXMLOperations) SaveToFile(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	if !strings.HasSuffix(strings.ToLower(filename), ".fcpxml") {
		filename += ".fcpxml"
	}
	
	data, err := ops.GenerateAndValidate()
	if err != nil {
		return fmt.Errorf("failed to generate FCPXML: %v", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	
	return nil
}

// GetDocumentStatistics returns comprehensive document statistics
func (ops *SafeFCPXMLOperations) GetDocumentStatistics() DocumentStatistics {
	return ops.builder.GetStatistics()
}

// ===============================
// Validation Helper Functions
// ===============================

// validateMediaPath validates a media file path
func (ops *SafeFCPXMLOperations) validateMediaPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	
	// Check file extension is supported
	supportedExts := map[string]bool{
		".mp4":  true, ".mov": true, ".avi": true, ".mkv": true, ".m4v": true,
		".png":  true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".tiff": true,
		".mp3":  true, ".wav": true, ".aac": true, ".m4a": true, ".flac": true,
	}
	
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedExts[ext] {
		return fmt.Errorf("unsupported file format: %s", ext)
	}
	
	return nil
}

// validateLane validates a lane number
func (ops *SafeFCPXMLOperations) validateLane(lane int) error {
	if lane < -10 || lane > 10 {
		return fmt.Errorf("lane out of range [-10, 10]: %d", lane)
	}
	return nil
}

// validatePosition validates X,Y position values
func (ops *SafeFCPXMLOperations) validatePosition(x, y float64) error {
	// Reasonable bounds for FCPXML position values
	if x < -5000 || x > 5000 {
		return fmt.Errorf("X position out of range [-5000, 5000]: %f", x)
	}
	if y < -5000 || y > 5000 {
		return fmt.Errorf("Y position out of range [-5000, 5000]: %f", y)
	}
	return nil
}

// validateScale validates scale values
func (ops *SafeFCPXMLOperations) validateScale(scale float64) error {
	if scale <= 0 {
		return fmt.Errorf("scale must be positive: %f", scale)
	}
	if scale > 100 {
		return fmt.Errorf("scale too large (max 100): %f", scale)
	}
	return nil
}

// validateRotation validates rotation angle values
func (ops *SafeFCPXMLOperations) validateRotation(angle float64) error {
	// No strict limits on rotation, but warn about extreme values
	if angle < -3600 || angle > 3600 {
		return fmt.Errorf("rotation angle extreme (consider normalizing): %f", angle)
	}
	return nil
}

// validateTitleCardConfiguration validates title card settings
func (ops *SafeFCPXMLOperations) validateTitleCardConfiguration(config *TitleCardConfiguration) error {
	if config.Font == "" {
		return fmt.Errorf("font cannot be empty")
	}
	
	if config.FontSize == "" {
		return fmt.Errorf("font size cannot be empty")
	}
	
	// Validate font size is numeric
	if _, err := strconv.ParseFloat(config.FontSize, 64); err != nil {
		return fmt.Errorf("invalid font size: %s", config.FontSize)
	}
	
	// Validate color format (should be "R G B A" with values 0-1)
	if err := ops.validateColorString(config.FontColor); err != nil {
		return fmt.Errorf("invalid font color: %v", err)
	}
	
	// Validate alignment
	validAlignments := map[string]bool{
		"left": true, "center": true, "right": true, "justified": true,
	}
	if !validAlignments[config.Alignment] {
		return fmt.Errorf("invalid alignment: %s", config.Alignment)
	}
	
	return nil
}

// validateColorString validates an RGBA color string
func (ops *SafeFCPXMLOperations) validateColorString(colorStr string) error {
	parts := strings.Fields(colorStr)
	if len(parts) != 4 {
		return fmt.Errorf("color must have 4 components (RGBA): %s", colorStr)
	}
	
	for i, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return fmt.Errorf("invalid color component %d: %s", i, part)
		}
		if value < 0.0 || value > 1.0 {
			return fmt.Errorf("color component %d out of range [0,1]: %f", i, value)
		}
	}
	
	return nil
}

// validateMediaLayerSpec validates a media layer specification
func (ops *SafeFCPXMLOperations) validateMediaLayerSpec(spec MediaLayerSpec) error {
	if err := ops.validateMediaPath(spec.Path); err != nil {
		return fmt.Errorf("invalid path: %v", err)
	}
	
	if spec.StartSeconds < 0 {
		return fmt.Errorf("start time cannot be negative: %f", spec.StartSeconds)
	}
	
	if spec.DurationSeconds <= 0 {
		return fmt.Errorf("duration must be positive: %f", spec.DurationSeconds)
	}
	
	if err := ops.validateLane(spec.Lane); err != nil {
		return fmt.Errorf("invalid lane: %v", err)
	}
	
	// Validate type matches file extension
	detectedType, err := DetectMediaTypeFromPath(spec.Path)
	if err != nil {
		return fmt.Errorf("failed to detect media type: %v", err)
	}
	
	if spec.Type != "" && spec.Type != detectedType.String() {
		return fmt.Errorf("specified type '%s' doesn't match detected type '%s'", spec.Type, detectedType.String())
	}
	
	return nil
}

// validateDocumentSettings validates document settings
func (ops *SafeFCPXMLOperations) validateDocumentSettings(settings DocumentSettings) error {
	if settings.MaxLanes < 1 || settings.MaxLanes > 20 {
		return fmt.Errorf("max lanes out of range [1, 20]: %d", settings.MaxLanes)
	}
	
	return nil
}

// ===============================
// Configuration Types
// ===============================

// TitleCardConfiguration represents title card styling options
type TitleCardConfiguration struct {
	Font      string
	FontSize  string
	FontColor string
	Alignment string
	Bold      bool
	Italic    bool
}

// TitleCardOption configures title card styling
type TitleCardOption func(*TitleCardConfiguration)

// WithTitleFont sets the title font
func WithTitleFont(font string) TitleCardOption {
	return func(tc *TitleCardConfiguration) {
		tc.Font = font
	}
}

// WithTitleSize sets the title font size
func WithTitleSize(size string) TitleCardOption {
	return func(tc *TitleCardConfiguration) {
		tc.FontSize = size
	}
}

// WithTitleColor sets the title color
func WithTitleColor(color string) TitleCardOption {
	return func(tc *TitleCardConfiguration) {
		tc.FontColor = color
	}
}

// WithTitleAlignment sets the title alignment
func WithTitleAlignment(alignment string) TitleCardOption {
	return func(tc *TitleCardConfiguration) {
		tc.Alignment = alignment
	}
}

// WithTitleBold sets title bold formatting
func WithTitleBold(bold bool) TitleCardOption {
	return func(tc *TitleCardConfiguration) {
		tc.Bold = bold
	}
}

// WithTitleItalic sets title italic formatting
func WithTitleItalic(italic bool) TitleCardOption {
	return func(tc *TitleCardConfiguration) {
		tc.Italic = italic
	}
}

// MediaLayerSpec specifies a media layer in a composite
type MediaLayerSpec struct {
	Path            string
	Type            string  // "video", "image", "audio" (auto-detected if empty)
	StartSeconds    float64
	DurationSeconds float64
	Lane            int
	Animations      map[string][]KeyframeData // Optional animations
}

// DocumentSettings configures document-level behavior
type DocumentSettings struct {
	MaxLanes      int
	AllowOverlaps bool
	AllowLaneGaps bool
}

// DefaultDocumentSettings returns sensible defaults
func DefaultDocumentSettings() DocumentSettings {
	return DocumentSettings{
		MaxLanes:      10,
		AllowOverlaps: false,
		AllowLaneGaps: false,
	}
}

// ===============================
// Convenience Functions  
// ===============================

// CreateSimpleVideo creates a simple single-video timeline
func CreateSimpleVideo(projectName, videoPath string, durationSeconds float64) (*SafeFCPXMLOperations, error) {
	ops, err := NewSafeFCPXMLOperations(projectName, durationSeconds)
	if err != nil {
		return nil, err
	}
	
	if err := ops.AddBackgroundVideo(videoPath, durationSeconds); err != nil {
		return nil, fmt.Errorf("failed to add background video: %v", err)
	}
	
	return ops, nil
}

// CreateImageSlideshow creates a slideshow from image files
func CreateImageSlideshow(projectName string, imagePaths []string, secondsPerImage float64) (*SafeFCPXMLOperations, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("at least one image is required")
	}
	
	totalDuration := float64(len(imagePaths)) * secondsPerImage
	ops, err := NewSafeFCPXMLOperations(projectName, totalDuration)
	if err != nil {
		return nil, err
	}
	
	currentTime := 0.0
	for i, imagePath := range imagePaths {
		if err := ops.AddImageOverlay(imagePath, currentTime, secondsPerImage, 0); err != nil {
			return nil, fmt.Errorf("failed to add image %d: %v", i, err)
		}
		currentTime += secondsPerImage
	}
	
	return ops, nil
}

// CreateTitleSequence creates a sequence of title cards
func CreateTitleSequence(projectName string, titles []string, secondsPerTitle float64) (*SafeFCPXMLOperations, error) {
	if len(titles) == 0 {
		return nil, fmt.Errorf("at least one title is required")
	}
	
	totalDuration := float64(len(titles)) * secondsPerTitle
	ops, err := NewSafeFCPXMLOperations(projectName, totalDuration)
	if err != nil {
		return nil, err
	}
	
	currentTime := 0.0
	for i, title := range titles {
		if err := ops.AddTitleCard(title, currentTime, secondsPerTitle, 1); err != nil {
			return nil, fmt.Errorf("failed to add title %d: %v", i, err)
		}
		currentTime += secondsPerTitle
	}
	
	return ops, nil
}