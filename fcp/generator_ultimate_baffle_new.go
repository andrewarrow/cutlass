package fcp

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
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

// createComplexVideoAssets creates multiple video assets using real files from ./assets
func createComplexVideoAssets(tx *ResourceTransaction, count int) ([]AssetInfo, error) {
	// Scan for real video files
	realVideoFiles, err := scanForRealAssets("./assets", []string{".mp4", ".mov", ".avi", ".mkv"})
	if err != nil {
		return nil, fmt.Errorf("failed to scan for video assets: %v", err)
	}
	
	if len(realVideoFiles) == 0 {
		return nil, fmt.Errorf("no video files found in ./assets directory")
	}
	
	assets := make([]AssetInfo, count)
	
	// Create format for videos (using consistent format)
	formatConfig := struct {
		frameDuration string
		width, height string
		colorSpace    string
		name          string
	}{"1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)", "ComplexVideo"}
	
	for i := 0; i < count; i++ {
		ids := tx.ReserveIDs(2)
		assetID := ids[0]
		formatID := ids[1]
		
		// Create format
		_, err := tx.CreateFormatWithFrameDuration(formatID, formatConfig.frameDuration, formatConfig.width, formatConfig.height, formatConfig.colorSpace)
		if err != nil {
			return nil, fmt.Errorf("failed to create video format %d: %v", i, err)
		}
		
		// Use real video file (cycle through available files)
		realVideoPath := realVideoFiles[i%len(realVideoFiles)]
		assetName := fmt.Sprintf("ComplexVideo_%03d", i)
		duration := ConvertSecondsToFCPDuration(30.0 + float64(i%60)) // 30-90 second duration
		
		_, err = tx.CreateAsset(assetID, realVideoPath, assetName, duration, formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create video asset %d: %v", i, err)
		}
		
		assets[i] = AssetInfo{
			ID:       assetID,
			FormatID: formatID,
			Name:     assetName,
			Path:     realVideoPath,
			Duration: duration,
			Type:     "video",
		}
	}
	
	return assets, nil
}

// createComplexImageAssets creates multiple image assets using real files from ./assets
func createComplexImageAssets(tx *ResourceTransaction, count int) ([]AssetInfo, error) {
	// Scan for real image files
	realImageFiles, err := scanForRealAssets("./assets", []string{".png", ".jpg", ".jpeg", ".gif"})
	if err != nil {
		return nil, fmt.Errorf("failed to scan for image assets: %v", err)
	}
	
	if len(realImageFiles) == 0 {
		return nil, fmt.Errorf("no image files found in ./assets directory")
	}
	
	assets := make([]AssetInfo, count)
	
	// Standard image format
	formatConfig := struct {
		width, height string
		colorSpace    string
		name          string
	}{"1920", "1080", "1-13-1", "ComplexImage"}
	
	for i := 0; i < count; i++ {
		ids := tx.ReserveIDs(2)
		assetID := ids[0]
		formatID := ids[1]
		
		// Create format (images don't have frameDuration)
		_, err := tx.CreateFormat(formatID, formatConfig.name, formatConfig.width, formatConfig.height, formatConfig.colorSpace)
		if err != nil {
			return nil, fmt.Errorf("failed to create image format %d: %v", i, err)
		}
		
		// Use real image file (cycle through available files)
		realImagePath := realImageFiles[i%len(realImageFiles)]
		assetName := fmt.Sprintf("ComplexImage_%03d", i)
		
		_, err = tx.CreateAsset(assetID, realImagePath, assetName, "0s", formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create image asset %d: %v", i, err)
		}
		
		assets[i] = AssetInfo{
			ID:       assetID,
			FormatID: formatID,
			Name:     assetName,
			Path:     realImagePath,
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
	
	// Create video elements across MANY layers - go crazy!
	elementIndex := 0
	for _, asset := range videoAssets {
		for reuse := 0; reuse < config.AssetReuseCount; reuse++ {
			// 🔥 CRAZY TIMING: Random start times with lots of overlap
			startTime := (float64(elementIndex) / float64(totalVideoElements)) * timelineDuration * 1.2 // Extends beyond timeline!
			
			// 🔥 CRAZY DURATIONS: Very short to very long clips
			baseDuration := []float64{2.0, 5.0, 8.0, 15.0, 30.0, 45.0, 60.0, 90.0}[elementIndex%8]
			duration := baseDuration + float64(elementIndex%30) // 2-120 second range
			
			// Create asset-clip
			assetClip := AssetClip{
				Ref:      asset.ID,
				Offset:   ConvertSecondsToFCPDuration(startTime),
				Duration: ConvertSecondsToFCPDuration(duration),
				Name:     fmt.Sprintf("%s_Use_%d", asset.Name, reuse),
				Start:    "0s",
			}
			
			// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML architecture)
			// Lanes are only for connected clips nested within spine elements
			// All spine elements go on the main timeline without lane attributes
			
			// Add complex animations
			assetClip.AdjustTransform = createComplexAnimation(startTime, duration, config.KeyframesPerAnimation, elementIndex)
			
			spine.AssetClips = append(spine.AssetClips, assetClip)
			elementIndex++
		}
	}
	
	// Create image elements across MANY MORE layers - even crazier!
	elementIndex = 0
	for _, asset := range imageAssets {
		for reuse := 0; reuse < config.AssetReuseCount; reuse++ {
			// 🔥 CRAZY IMAGE TIMING: Staggered and overlapping
			baseTime := (float64(elementIndex) / float64(totalImageElements)) * timelineDuration
			startTime := baseTime + float64(elementIndex%20)*2.0 // Stagger by up to 40 seconds
			
			// 🔥 CRAZY IMAGE DURATIONS: Mix of very short and long
			durationPattern := []float64{1.5, 3.0, 6.0, 12.0, 25.0, 40.0, 75.0}[elementIndex%7]
			duration := durationPattern + float64(elementIndex%15) // 1.5-90 second range
			
			// Create video element for image
			video := Video{
				Ref:      asset.ID,
				Offset:   ConvertSecondsToFCPDuration(startTime),
				Duration: ConvertSecondsToFCPDuration(duration),
				Name:     fmt.Sprintf("%s_Use_%d", asset.Name, reuse),
			}
			
			// 🚨 FIXED: Spine video elements cannot have lanes (per FCPXML architecture)
			// Lanes are only for connected clips nested within spine elements
			// All spine elements go on the main timeline without lane attributes
			
			// Add simpler animations for images (per CLAUDE.md guidance)
			video.AdjustTransform = createImageAnimation(startTime, duration, elementIndex)
			
			spine.Videos = append(spine.Videos, video)
			elementIndex++
		}
	}
	
	// Create INSANE title elements across ALL possible layers!
	for i := 0; i < config.TitleElementCount; i++ {
		effectID := titleEffects[i%len(titleEffects)]
		
		// 🔥 CRAZY TITLE TIMING: Random scattered placement
		baseTime := (float64(i) / float64(config.TitleElementCount)) * timelineDuration
		startTime := baseTime + float64(i%25)*3.0 // Scatter by up to 75 seconds
		
		// 🔥 CRAZY TITLE DURATIONS: Very quick flashes to long holds
		durationOptions := []float64{0.5, 1.0, 2.0, 4.0, 8.0, 15.0, 30.0, 60.0}[i%8]
		duration := durationOptions + float64(i%10)*0.5 // 0.5-65 second range
		
		title := createComplexTitle(effectID, startTime, duration, i)
		
		// 🚨 FIXED: Spine title elements cannot have lanes (per FCPXML architecture)
		// Lanes are only for connected clips nested within spine elements
		// All spine elements go on the main timeline without lane attributes
		
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
func createComplexTitle(effectID string, startTime, duration float64, index int) Title {
	// Generate complex text content
	textContent := generateValidTextContent(index)
	
	// Generate unique style ID
	styleID := fmt.Sprintf("complex_style_%d_%d", index, int(time.Now().UnixNano()%10000))
	
	title := Title{
		Ref:      effectID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("ComplexTitle_%d", index),
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
	
	// Lane assignment will be done by caller - don't set here
	
	return title
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

// scanForRealAssets scans a directory for real media files with specified extensions
func scanForRealAssets(dir string, extensions []string) ([]string, error) {
	var assets []string
	
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}
	
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return assets, nil // Return empty slice if directory doesn't exist
	}
	
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		
		// Check if extension matches
		for _, validExt := range extensions {
			if ext == validExt {
				absolutePath := filepath.Join(absDir, filename)
				assets = append(assets, absolutePath)
				break
			}
		}
	}
	
	return assets, nil
}