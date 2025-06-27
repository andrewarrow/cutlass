package fcp

import (
	"fmt"
	"math"
	"time"
)

// ComplexBaffleConfig defines parameters for generating complex but valid FCPXML
type ComplexBaffleConfig struct {
	TimelineDurationMinutes  int     // Total timeline duration
	VideoAssetCount         int     // Number of video assets to create
	ImageAssetCount         int     // Number of image assets to create
	TitleElementCount       int     // Number of title elements
	MaxLayers              int     // Maximum number of layers
	KeyframesPerAnimation   int     // Keyframes per parameter animation
	AssetReuseCount        int     // How many times to reuse each asset
	ComplexityFactor       float64 // 0.0-1.0 scale of complexity
}

// DefaultComplexBaffleConfig returns a config for very complex but valid FCPXML
func DefaultComplexBaffleConfig() ComplexBaffleConfig {
	return ComplexBaffleConfig{
		TimelineDurationMinutes: 8,  // 8 minute timeline
		VideoAssetCount:        15, // 15 unique video assets
		ImageAssetCount:        25, // 25 unique image assets  
		TitleElementCount:      50, // 50 title elements
		MaxLayers:             20, // Up to 20 layers deep
		KeyframesPerAnimation:  25, // 25 keyframes per animation
		AssetReuseCount:       8,  // Each asset used 8 times
		ComplexityFactor:      0.9, // Very high complexity
	}
}

// GenerateComplexBaffle creates genuinely complex but completely valid FCPXML
// using proper fcp package patterns and transaction management
func GenerateComplexBaffle(outputPath string, config ComplexBaffleConfig) error {
	fmt.Printf("🎬 GENERATING COMPLEX VALID BAFFLE 🎬\n")
	fmt.Printf("Timeline: %d minutes, %d video assets, %d images, %d titles\n", 
		config.TimelineDurationMinutes, config.VideoAssetCount, config.ImageAssetCount, config.TitleElementCount)
	
	// Create base FCPXML using proper pattern
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		return fmt.Errorf("failed to create base FCPXML: %v", err)
	}
	
	// Set up proper resource management
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()
	
	// Calculate timeline duration
	timelineDuration := float64(config.TimelineDurationMinutes * 60)
	
	// Update sequence duration
	fcpxml.Library.Events[0].Projects[0].Sequences[0].Duration = ConvertSecondsToFCPDuration(timelineDuration)
	
	fmt.Printf("Creating %d video assets...\n", config.VideoAssetCount)
	videoAssets, err := createComplexVideoAssets(tx, config.VideoAssetCount)
	if err != nil {
		return fmt.Errorf("failed to create video assets: %v", err)
	}
	
	fmt.Printf("Creating %d image assets...\n", config.ImageAssetCount)
	imageAssets, err := createComplexImageAssets(tx, config.ImageAssetCount)
	if err != nil {
		return fmt.Errorf("failed to create image assets: %v", err)
	}
	
	fmt.Printf("Creating %d title effects...\n", config.TitleElementCount)
	titleEffects, err := createComplexTitleEffects(tx, config.TitleElementCount)
	if err != nil {
		return fmt.Errorf("failed to create title effects: %v", err)
	}
	
	// Commit all resources
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit resources: %v", err)
	}
	
	fmt.Printf("Building complex timeline structure...\n")
	if err := buildComplexTimeline(fcpxml, videoAssets, imageAssets, titleEffects, config, timelineDuration); err != nil {
		return fmt.Errorf("failed to build timeline: %v", err)
	}
	
	fmt.Printf("Validating complex structure...\n")
	if err := fcpxml.ValidateStructure(); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}
	
	fmt.Printf("Writing FCPXML...\n")
	if err := WriteToFile(fcpxml, outputPath); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	
	fmt.Printf("✅ SUCCESS: Complex valid FCPXML generated!\n")
	fmt.Printf("File: %s\n", outputPath)
	fmt.Printf("Ready for Final Cut Pro import!\n")
	
	return nil
}

// createComplexVideoAssets creates multiple video assets with various formats
func createComplexVideoAssets(tx *ResourceTransaction, count int) ([]AssetInfo, error) {
	assets := make([]AssetInfo, count)
	
	// Create variety of video formats (using valid frame durations)
	formatConfigs := []struct {
		frameDuration string
		width, height string
		colorSpace    string
		name          string
	}{
		{"1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)", "1080p24"},
		{"2002/24000s", "1920", "1080", "1-1-1 (Rec. 709)", "1080p12"},
		{"1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)", "1080p24_alt"},
		{"1001/24000s", "3840", "2160", "1-1-1 (Rec. 709)", "4K24"},
		{"1001/24000s", "1280", "720", "1-1-1 (Rec. 709)", "720p24"},
		{"1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)", "1080p24_v2"},
	}
	
	for i := 0; i < count; i++ {
		ids := tx.ReserveIDs(2)
		assetID := ids[0]
		formatID := ids[1]
		
		// Use different format configs cyclically
		config := formatConfigs[i%len(formatConfigs)]
		
		// Create format
		_, err := tx.CreateFormatWithFrameDuration(formatID, config.frameDuration, config.width, config.height, config.colorSpace)
		if err != nil {
			return nil, fmt.Errorf("failed to create video format %d: %v", i, err)
		}
		
		// Create asset with realistic path and duration
		assetName := fmt.Sprintf("ComplexVideo_%03d", i)
		assetPath := fmt.Sprintf("/Users/aa/cs/cutlass/assets/complex_%03d.mov", i)
		duration := ConvertSecondsToFCPDuration(30.0 + float64(i%60)) // 30-90 second videos
		
		_, err = tx.CreateAsset(assetID, assetPath, assetName, duration, formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create video asset %d: %v", i, err)
		}
		
		assets[i] = AssetInfo{
			ID:       assetID,
			FormatID: formatID,
			Name:     assetName,
			Path:     assetPath,
			Duration: duration,
			Type:     "video",
		}
	}
	
	return assets, nil
}

// createComplexImageAssets creates multiple image assets with various formats
func createComplexImageAssets(tx *ResourceTransaction, count int) ([]AssetInfo, error) {
	assets := make([]AssetInfo, count)
	
	// Various image format configs
	formatConfigs := []struct {
		width, height string
		colorSpace    string
		name          string
	}{
		{"1920", "1080", "1-13-1", "HD_Image"},
		{"3840", "2160", "1-13-1", "4K_Image"},
		{"1280", "720", "1-13-1", "HD_Image_720"},
		{"2560", "1440", "1-13-1", "QHD_Image"},
		{"1080", "1920", "1-13-1", "Portrait_HD"},
		{"4096", "2160", "1-13-1", "Cinema_4K"},
	}
	
	for i := 0; i < count; i++ {
		ids := tx.ReserveIDs(2)
		assetID := ids[0]
		formatID := ids[1]
		
		// Use different format configs cyclically
		config := formatConfigs[i%len(formatConfigs)]
		
		// Create format (images don't have frameDuration)
		_, err := tx.CreateFormat(formatID, config.name, config.width, config.height, config.colorSpace)
		if err != nil {
			return nil, fmt.Errorf("failed to create image format %d: %v", i, err)
		}
		
		// Create asset with realistic path
		assetName := fmt.Sprintf("ComplexImage_%03d", i)
		assetPath := fmt.Sprintf("/Users/aa/cs/cutlass/assets/complex_%03d.png", i)
		
		_, err = tx.CreateAsset(assetID, assetPath, assetName, "0s", formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create image asset %d: %v", i, err)
		}
		
		assets[i] = AssetInfo{
			ID:       assetID,
			FormatID: formatID,
			Name:     assetName,
			Path:     assetPath,
			Duration: "0s",
			Type:     "image",
		}
	}
	
	return assets, nil
}

// createComplexTitleEffects creates multiple title effects for text
func createComplexTitleEffects(tx *ResourceTransaction, count int) ([]string, error) {
	effects := make([]string, count)
	
	for i := 0; i < count; i++ {
		effectID := tx.ReserveIDs(1)[0]
		
		// Use verified title effect UID from samples
		_, err := tx.CreateEffect(effectID, fmt.Sprintf("ComplexTitle_%03d", i), 
			".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			return nil, fmt.Errorf("failed to create title effect %d: %v", i, err)
		}
		
		effects[i] = effectID
	}
	
	return effects, nil
}

// AssetInfo holds information about created assets
type AssetInfo struct {
	ID       string
	FormatID string
	Name     string
	Path     string
	Duration string
	Type     string // "video", "image"
}

// buildComplexTimeline creates a very complex timeline structure
func buildComplexTimeline(fcpxml *FCPXML, videoAssets, imageAssets []AssetInfo, titleEffects []string, config ComplexBaffleConfig, timelineDuration float64) error {
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	
	// Calculate total elements to create
	totalVideoElements := len(videoAssets) * config.AssetReuseCount
	totalImageElements := len(imageAssets) * config.AssetReuseCount
	
	fmt.Printf("Creating %d video elements, %d image elements, %d titles...\n", 
		totalVideoElements, totalImageElements, config.TitleElementCount)
	
	// Create video elements across multiple layers
	elementIndex := 0
	for _, asset := range videoAssets {
		for reuse := 0; reuse < config.AssetReuseCount; reuse++ {
			// Calculate timing - spread across timeline
			startTime := (float64(elementIndex) / float64(totalVideoElements)) * timelineDuration * 0.8
			duration := 15.0 + float64(elementIndex%20) // 15-35 second clips
			
			// Calculate lane (distribute across layers)
			lane := (elementIndex % config.MaxLayers) + 1
			
			// Create complex asset-clip with animations
			assetClip := AssetClip{
				Ref:      asset.ID,
				Offset:   ConvertSecondsToFCPDuration(startTime),
				Duration: ConvertSecondsToFCPDuration(duration),
				Name:     fmt.Sprintf("%s_Use_%d", asset.Name, reuse),
				Lane:     fmt.Sprintf("%d", lane),
				Start:    "0s",
			}
			
			// Add complex animations
			assetClip.AdjustTransform = createComplexAnimation(startTime, duration, config.KeyframesPerAnimation, elementIndex)
			
			spine.AssetClips = append(spine.AssetClips, assetClip)
			elementIndex++
		}
	}
	
	// Create image elements (using Video elements for images per architecture)
	elementIndex = 0
	for _, asset := range imageAssets {
		for reuse := 0; reuse < config.AssetReuseCount; reuse++ {
			// Calculate timing
			startTime := (float64(elementIndex) / float64(totalImageElements)) * timelineDuration * 0.9
			duration := 8.0 + float64(elementIndex%12) // 8-20 second displays
			
			// Calculate lane
			lane := (elementIndex % config.MaxLayers) + 1
			
			// Create video element for image
			video := Video{
				Ref:      asset.ID,
				Offset:   ConvertSecondsToFCPDuration(startTime),
				Duration: ConvertSecondsToFCPDuration(duration),
				Name:     fmt.Sprintf("%s_Use_%d", asset.Name, reuse),
				Lane:     fmt.Sprintf("%d", lane),
			}
			
			// Add simpler animations for images (per CLAUDE.md guidance)
			video.AdjustTransform = createImageAnimation(startTime, duration, elementIndex)
			
			spine.Videos = append(spine.Videos, video)
			elementIndex++
		}
	}
	
	// Create complex title elements
	for i := 0; i < config.TitleElementCount; i++ {
		effectID := titleEffects[i%len(titleEffects)]
		
		// Calculate timing - spread evenly
		startTime := (float64(i) / float64(config.TitleElementCount)) * timelineDuration
		duration := 5.0 + float64(i%8) // 5-13 second titles
		
		// Calculate lane - use higher lanes for titles
		lane := (i % (config.MaxLayers/2)) + config.MaxLayers/2 + 1
		
		title := createComplexTitle(effectID, startTime, duration, i, lane)
		spine.Titles = append(spine.Titles, title)
	}
	
	return nil
}

// createComplexAnimation creates sophisticated multi-parameter keyframe animations
func createComplexAnimation(startTime, duration float64, keyframeCount, seed int) *AdjustTransform {
	// Create multiple keyframes with complex but valid patterns
	keyframes := make([]Keyframe, keyframeCount)
	
	for i := 0; i < keyframeCount; i++ {
		// Distribute keyframes across duration
		keyTime := startTime + (float64(i)/float64(keyframeCount-1))*duration
		
		// Generate complex but reasonable positions
		phase := float64(seed+i) * 0.2
		x := math.Sin(phase) * 200 + math.Cos(phase*1.7) * 100
		y := math.Cos(phase) * 150 + math.Sin(phase*1.3) * 80
		
		keyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.1f %.1f", x, y),
			// Position keyframes have NO curve/interp attributes per validation rules
		}
	}
	
	// Scale keyframes with valid curve
	scaleKeyframes := make([]Keyframe, keyframeCount/2)
	for i := 0; i < len(scaleKeyframes); i++ {
		keyTime := startTime + (float64(i)/float64(len(scaleKeyframes)-1))*duration
		scale := 0.8 + 0.4*math.Sin(float64(seed+i)*0.1)
		
		scaleKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.2f %.2f", scale, scale),
			Curve: "linear", // Only valid curve value
		}
	}
	
	// Rotation keyframes
	rotationKeyframes := make([]Keyframe, keyframeCount/3)
	for i := 0; i < len(rotationKeyframes); i++ {
		keyTime := startTime + (float64(i)/float64(len(rotationKeyframes)-1))*duration
		rotation := math.Sin(float64(seed+i)*0.15) * 45 // -45 to +45 degrees
		
		rotationKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.1f", rotation),
			Curve: "linear",
		}
	}
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: keyframes,
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: scaleKeyframes,
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: rotationKeyframes,
				},
			},
		},
	}
}

// createImageAnimation creates simpler animations suitable for images
func createImageAnimation(startTime, duration float64, seed int) *AdjustTransform {
	// Images get simpler animations per CLAUDE.md
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.1f %.1f", float64(seed%100-50), float64(seed%80-40)),
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f %.1f", float64((seed*2)%100-50), float64((seed*3)%80-40)),
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "0.9 0.9",
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "1.1 1.1",
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// createComplexTitle creates sophisticated title with valid styling
func createComplexTitle(effectID string, startTime, duration float64, index, lane int) Title {
	// Generate complex text content
	textContent := generateValidTextContent(index)
	
	// Generate unique style ID
	styleID := fmt.Sprintf("complex_style_%d_%d", index, int(time.Now().UnixNano()%10000))
	
	return Title{
		Ref:      effectID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("ComplexTitle_%d", index),
		Lane:     fmt.Sprintf("%d", lane),
		Text: &TitleText{
			TextStyles: []TextStyleRef{
				{
					Ref:  styleID,
					Text: textContent,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: styleID,
				TextStyle: TextStyle{
					Font:        getValidFont(index),
					FontSize:    fmt.Sprintf("%.0f", 48+float64(index%200)), // 48-248pt
					FontColor:   getValidColor(index),
					Alignment:   "center", // Only valid alignment per samples
					LineSpacing: fmt.Sprintf("%.1f", 1.0+float64(index%10)*0.1), // 1.0-2.0
					Bold:        getValidBool(index),
					Italic:      getValidBool(index + 1),
				},
			},
		},
	}
}

// generateValidTextContent creates interesting but valid text content
func generateValidTextContent(index int) string {
	texts := []string{
		"COMPLEX BAFFLE TEST",
		"Advanced FCPXML Generation",
		"Professional Motion Graphics",
		"Sophisticated Timeline Structure",
		"Multi-Layer Composition",
		"Keyframe Animation System",
		"Resource Management Test",
		"Transaction Validation",
		"Complex but Valid Content",
		"Final Cut Pro Compatibility",
	}
	
	baseText := texts[index%len(texts)]
	return fmt.Sprintf("%s #%d", baseText, index)
}

// getValidFont returns fonts that are commonly available
func getValidFont(index int) string {
	fonts := []string{
		"Helvetica Neue",
		"Arial",
		"Times New Roman",
		"Courier New",
		"Verdana",
		"Georgia",
		"Impact",
		"Trebuchet MS",
	}
	return fonts[index%len(fonts)]
}

// getValidColor returns valid RGBA color values
func getValidColor(index int) string {
	colors := []string{
		"1 1 1 1",     // White
		"0 0 0 1",     // Black
		"1 0 0 1",     // Red
		"0 1 0 1",     // Green
		"0 0 1 1",     // Blue
		"1 1 0 1",     // Yellow
		"1 0 1 1",     // Magenta
		"0 1 1 1",     // Cyan
		"0.5 0.5 0.5 1", // Gray
	}
	return colors[index%len(colors)]
}

// getValidBool returns valid boolean strings
func getValidBool(index int) string {
	if index%2 == 0 {
		return "1"
	}
	return "0"
}