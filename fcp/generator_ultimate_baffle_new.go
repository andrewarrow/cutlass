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

// buildComplexTimeline creates a very complex timeline structure with proper multi-lane nesting
func buildComplexTimeline(fcpxml *FCPXML, videoAssets, imageAssets []AssetInfo, titleEffects []string, config ComplexBaffleConfig, timelineDuration float64) error {
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	
	// Calculate total elements to create
	totalVideoElements := len(videoAssets) * config.AssetReuseCount
	totalImageElements := len(imageAssets) * config.AssetReuseCount
	
	fmt.Printf("Creating %d video elements, %d image elements, %d titles with 8+ lanes...\n", 
		totalVideoElements, totalImageElements, config.TitleElementCount)
	
	// 🎬 NEW APPROACH: Create primary spine elements with nested connected clips for multi-lane structure
	numPrimaryElements := 5 + (len(videoAssets) / 3) // 5-9 primary elements
	
	for primaryIndex := 0; primaryIndex < numPrimaryElements && primaryIndex < len(videoAssets); primaryIndex++ {
		asset := videoAssets[primaryIndex]
		
		// 🎬 PRIMARY SPINE ELEMENT: Create main asset-clip (NO lane)
		primaryStartTime := (float64(primaryIndex) / float64(numPrimaryElements)) * timelineDuration
		primaryDuration := timelineDuration / float64(numPrimaryElements) * 1.5 // Some overlap
		
		primaryClip := AssetClip{
			Ref:      asset.ID,
			Offset:   ConvertSecondsToFCPDuration(primaryStartTime),
			Duration: ConvertSecondsToFCPDuration(primaryDuration),
			Name:     fmt.Sprintf("Primary_%s", asset.Name),
			Start:    "0s",
		}
		
		// Add primary element animations
		primaryClip.AdjustTransform = createComplexAnimation(primaryStartTime, primaryDuration, config.KeyframesPerAnimation, primaryIndex)
		
		// 🎬 NESTED CONNECTED CLIPS: Create 6-12 connected clips with lanes inside this primary element
		numConnectedClips := 6 + (primaryIndex % 7) // 6-12 connected clips per primary
		
		for connectedIndex := 0; connectedIndex < numConnectedClips; connectedIndex++ {
			// Choose assets for connected clips
			var connectedAsset AssetInfo
			if connectedIndex%2 == 0 && len(videoAssets) > (connectedIndex/2+1) {
				// Use different video asset
				connectedAsset = videoAssets[(primaryIndex+connectedIndex/2+1)%len(videoAssets)]
			} else if len(imageAssets) > (connectedIndex%len(imageAssets)) {
				// Use image asset (create nested Video element for images)
				imageAsset := imageAssets[connectedIndex%len(imageAssets)]
				
				// Create nested video element (for image) with lane
				nestedVideo := Video{
					Ref:      imageAsset.ID,
					Offset:   ConvertSecondsToFCPDuration(float64(connectedIndex) * 2.0), // Stagger timing
					Duration: ConvertSecondsToFCPDuration(5.0 + float64(connectedIndex%10)),
					Name:     fmt.Sprintf("Connected_Image_%d_%d", primaryIndex, connectedIndex),
					Lane:     fmt.Sprintf("%d", ((connectedIndex%8)-4)), // Lanes: -4 to +3
				}
				nestedVideo.AdjustTransform = createImageAnimation(primaryStartTime+float64(connectedIndex)*2.0, 5.0, connectedIndex)
				primaryClip.Videos = append(primaryClip.Videos, nestedVideo)
				continue
			} else {
				continue // Skip if no asset available
			}
			
			// Create nested asset-clip with lane
			nestedClip := AssetClip{
				Ref:      connectedAsset.ID,
				Offset:   ConvertSecondsToFCPDuration(float64(connectedIndex) * 3.0), // Stagger timing
				Duration: ConvertSecondsToFCPDuration(8.0 + float64(connectedIndex%15)),
				Name:     fmt.Sprintf("Connected_%s_%d_%d", connectedAsset.Name, primaryIndex, connectedIndex),
				Start:    "0s",
				Lane:     fmt.Sprintf("%d", ((connectedIndex%10)-5)), // Lanes: -5 to +4
			}
			
			// Add animations to connected clips
			nestedClip.AdjustTransform = createComplexAnimation(
				primaryStartTime+float64(connectedIndex)*3.0, 
				8.0+float64(connectedIndex%15), 
				config.KeyframesPerAnimation/2, 
				primaryIndex*100+connectedIndex,
			)
			
			primaryClip.NestedAssetClips = append(primaryClip.NestedAssetClips, nestedClip)
		}
		
		// Add nested titles with lanes to primary element
		for titleIndex := 0; titleIndex < 3 && (primaryIndex*3+titleIndex) < len(titleEffects); titleIndex++ {
			effectID := titleEffects[primaryIndex*3+titleIndex]
			
			nestedTitle := createComplexTitle(
				effectID, 
				primaryStartTime+float64(titleIndex)*4.0, 
				6.0+float64(titleIndex%8), 
				primaryIndex*10+titleIndex,
			)
			
			// Set lane for nested title (THIS IS CORRECT - nested elements can have lanes)
			nestedTitle.Lane = fmt.Sprintf("%d", ((titleIndex%6)+5)) // Lanes: +5 to +10
			
			primaryClip.Titles = append(primaryClip.Titles, nestedTitle)
		}
		
		spine.AssetClips = append(spine.AssetClips, primaryClip)
	}
	
	// 🎬 ADDITIONAL SPINE ELEMENTS: Add some standalone image and title elements on main spine
	// These are primary spine elements (NO lanes) to add more complexity
	
	// Add a few standalone image elements on main spine
	numStandaloneImages := 3
	for i := 0; i < numStandaloneImages && i < len(imageAssets); i++ {
		asset := imageAssets[i]
		startTime := timelineDuration * 0.8 + float64(i)*5.0 // Near end of timeline
		duration := 10.0 + float64(i)*2.0
		
		video := Video{
			Ref:      asset.ID,
			Offset:   ConvertSecondsToFCPDuration(startTime),
			Duration: ConvertSecondsToFCPDuration(duration),
			Name:     fmt.Sprintf("Standalone_Image_%d", i),
		}
		
		video.AdjustTransform = createImageAnimation(startTime, duration, i+1000)
		spine.Videos = append(spine.Videos, video)
	}
	
	// Add a few standalone title elements on main spine  
	numStandaloneTitles := 2
	for i := 0; i < numStandaloneTitles && i < len(titleEffects); i++ {
		effectID := titleEffects[len(titleEffects)-1-i] // Use different effects
		startTime := timelineDuration * 0.1 + float64(i)*8.0 // Near beginning
		duration := 12.0 + float64(i)*3.0
		
		title := createComplexTitle(effectID, startTime, duration, i+2000)
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