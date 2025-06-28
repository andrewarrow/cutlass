package fcp

import (
	"fmt"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// createLaneAssetClipElement creates an asset-clip element with proper lane assignment for spine
func createLaneAssetClipElement(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, startTime, duration float64, lane, index int, verbose bool, createdAssets, createdFormats map[string]string) (*AssetClip, error) {
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[videoPath]; exists {
		assetID = existingAssetID
		formatID = createdFormats[videoPath]
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), ConvertSecondsToFCPDuration(duration), formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create video asset: %v", err)
		}

		createdAssets[videoPath] = assetID
		createdFormats[videoPath] = formatID
	}

	assetClip := &AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Lane%dVideo_%d", lane, index),
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
		AdjustTransform: &AdjustTransform{
			Position: generateRandomPosition(),
			Scale:    fmt.Sprintf("%.2f %.2f", 0.7+rand.Float64()*0.4, 0.7+rand.Float64()*0.4),
		},
	}

	return assetClip, nil
}

// createLaneImageElement creates a video element (for image) with proper lane assignment for spine
func createLaneImageElement(fcpxml *FCPXML, tx *ResourceTransaction, imagePath string, startTime, duration float64, lane, index int, verbose bool, createdAssets, createdFormats map[string]string) (*Video, error) {
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[imagePath]; exists {
		assetID = existingAssetID
		formatID = createdFormats[imagePath]
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		_, err = tx.CreateAsset(assetID, imagePath, filepath.Base(imagePath), "0s", formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create image asset: %v", err)
		}

		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return nil, fmt.Errorf("failed to create image format: %v", err)
		}

		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}

	video := &Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Lane%dImage_%d", lane, index),
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
		AdjustTransform: &AdjustTransform{
			Position: generateRandomPosition(),
			Scale:    fmt.Sprintf("%.2f %.2f", 0.6+rand.Float64()*0.5, 0.6+rand.Float64()*0.5),
		},
	}

	return video, nil
}

// generateRandomPosition generates a random but reasonable position for elements
func generateRandomPosition() string {

	x := int(-200 + rand.Float64()*400)
	y := int(-150 + rand.Float64()*300)
	return fmt.Sprintf("%d %d", x, y)
}

// createNestedVideoElement creates a main video element with nested overlays (proper multi-lane structure)
func createNestedVideoElement(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, duration float64, verbose bool, assets *AssetCollection, createdAssets, createdFormats map[string]string) (*Video, error) {
	// Create main video asset
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[videoPath]; exists {
		assetID = existingAssetID
		formatID = createdFormats[videoPath]
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), ConvertSecondsToFCPDuration(duration), formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create video asset: %v", err)
		}

		createdAssets[videoPath] = assetID
		createdFormats[videoPath] = formatID
	}

	mainVideo := &Video{
		Ref:      assetID,
		Offset:   "0s",
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     "MainBackground",
	}

	// 🚨 EXTREME NESTED CHAOS: 50-200 overlays per main video!
	numOverlays := 50 + rand.Intn(150)

	for i := 1; i <= numOverlays; i++ {
		// 🚨 EXTREME: Overlays can start/end anywhere, even negative times
		overlayStartTime := -duration + rand.Float64()*(duration*3.0)
		overlayDuration := 0.01 + rand.Float64()*(duration*2.0) // Tiny to huge durations
		
		// 🚨 EXTREME: Massive lane numbers, negatives, zero
		lane := -10 + rand.Intn(21) // Valid range: -10 to +10

		overlayType := rand.Intn(3)

		switch overlayType {
		case 0:
			if len(assets.Images) > 0 {
				imagePath := assets.Images[rand.Intn(len(assets.Images))]
				uniqueImage, err := createUniqueMediaCopy(imagePath, fmt.Sprintf("overlay_img_%d", i))
				if err != nil && verbose {
					fmt.Printf("Warning: Failed to create unique image copy: %v\n", err)
					uniqueImage = imagePath
				}

				overlay, err := createImageOverlay(fcpxml, tx, uniqueImage, overlayStartTime, overlayDuration, lane, i, verbose, createdAssets, createdFormats)
				if err != nil && verbose {
					fmt.Printf("Warning: Failed to create image overlay: %v\n", err)
				} else {
					mainVideo.NestedVideos = append(mainVideo.NestedVideos, *overlay)
				}
			}

		case 1:
			if len(assets.Videos) > 0 {
				videoPath := assets.Videos[rand.Intn(len(assets.Videos))]
				uniqueVideo, err := createUniqueMediaCopy(videoPath, fmt.Sprintf("overlay_vid_%d", i))
				if err != nil && verbose {
					fmt.Printf("Warning: Failed to create unique video copy: %v\n", err)
					uniqueVideo = videoPath
				}

				overlay, err := createVideoOverlay(fcpxml, tx, uniqueVideo, overlayStartTime, overlayDuration, lane, i, verbose, createdAssets, createdFormats)
				if err != nil && verbose {
					fmt.Printf("Warning: Failed to create video overlay: %v\n", err)
				} else {
					mainVideo.NestedAssetClips = append(mainVideo.NestedAssetClips, *overlay)
				}
			}

		case 2:
			overlay, err := createTextOverlay(fcpxml, tx, overlayStartTime, overlayDuration, lane, i, verbose)
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create text overlay: %v\n", err)
			} else {
				mainVideo.NestedTitles = append(mainVideo.NestedTitles, *overlay)
			}
		}
	}

	if verbose {
		fmt.Printf("  Created main video with %d nested overlays\n", len(mainVideo.NestedVideos)+len(mainVideo.NestedAssetClips)+len(mainVideo.NestedTitles))
		fmt.Printf("    - NestedVideos: %d\n", len(mainVideo.NestedVideos))
		fmt.Printf("    - NestedAssetClips: %d\n", len(mainVideo.NestedAssetClips))
		fmt.Printf("    - NestedTitles: %d\n", len(mainVideo.NestedTitles))
	}

	return mainVideo, nil
}

// createNestedAssetClipElement creates an asset-clip with nested overlays
func createNestedAssetClipElement(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, startTime, duration float64, index int, verbose bool, assets *AssetCollection, createdAssets, createdFormats map[string]string) (*AssetClip, error) {
	// Create video asset
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[videoPath]; exists {
		assetID = existingAssetID
		formatID = createdFormats[videoPath]
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), ConvertSecondsToFCPDuration(duration), formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create video asset: %v", err)
		}

		createdAssets[videoPath] = assetID
		createdFormats[videoPath] = formatID
	}

	assetClip := &AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("MainClip_%d", index),
	}

	numOverlays := 2 + rand.Intn(4)

	for i := 1; i <= numOverlays; i++ {
		overlayStartTime := rand.Float64() * (duration * 0.7)
		overlayDuration := 2.0 + rand.Float64()*4.0

		if overlayStartTime+overlayDuration > duration {
			overlayDuration = duration - overlayStartTime
		}

		if rand.Float32() < 0.6 && len(assets.Images) > 0 {
			imagePath := assets.Images[rand.Intn(len(assets.Images))]
			uniqueImage, err := createUniqueMediaCopy(imagePath, fmt.Sprintf("nested_img_%d_%d", index, i))
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create unique image copy: %v\n", err)
				uniqueImage = imagePath
			}

			overlay, err := createImageOverlay(fcpxml, tx, uniqueImage, overlayStartTime, overlayDuration, i, i, verbose, createdAssets, createdFormats)
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create nested image overlay: %v\n", err)
			} else {
				assetClip.Videos = append(assetClip.Videos, *overlay)
			}
		} else {
			overlay, err := createTextOverlay(fcpxml, tx, overlayStartTime, overlayDuration, i, i, verbose)
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create nested text overlay: %v\n", err)
			} else {
				assetClip.Titles = append(assetClip.Titles, *overlay)
			}
		}
	}

	return assetClip, nil
}

// createNestedImageElement creates an image element with nested overlays
func createNestedImageElement(fcpxml *FCPXML, tx *ResourceTransaction, imagePath string, startTime, duration float64, index int, verbose bool, assets *AssetCollection, createdAssets, createdFormats map[string]string) (*Video, error) {
	// Create image asset
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[imagePath]; exists {
		assetID = existingAssetID
		formatID = createdFormats[imagePath]
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		_, err = tx.CreateAsset(assetID, imagePath, filepath.Base(imagePath), "0s", formatID)
		if err != nil {
			return nil, fmt.Errorf("failed to create image asset: %v", err)
		}

		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return nil, fmt.Errorf("failed to create image format: %v", err)
		}

		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}

	video := &Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("MainImage_%d", index),
	}

	numOverlays := 1 + rand.Intn(3)

	for i := 1; i <= numOverlays; i++ {
		overlayStartTime := rand.Float64() * (duration * 0.5)
		overlayDuration := 2.0 + rand.Float64()*4.0

		if overlayStartTime+overlayDuration > duration {
			overlayDuration = duration - overlayStartTime
		}

		overlay, err := createTextOverlay(fcpxml, tx, overlayStartTime, overlayDuration, i+1, i, verbose)
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create image text overlay: %v\n", err)
		} else {
			video.NestedTitles = append(video.NestedTitles, *overlay)
		}
	}

	return video, nil
}

// addBaffleImageElement adds an image element with proper lane assignment for BAFFLE system
func addBaffleImageElement(fcpxml *FCPXML, tx *ResourceTransaction, imagePath string, startTime, duration float64, elementIndex, targetLane int, verbose bool, createdAssets, createdFormats map[string]string) error {
	if verbose {
		fmt.Printf("  Adding image: %s (%.1fs @ %.1fs) lane %d\n", filepath.Base(imagePath), duration, startTime, targetLane)
	}

	// Reuse existing asset if already created, otherwise create new one
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[imagePath]; exists {
		assetID = existingAssetID
		if existingFormatID, formatExists := createdFormats[imagePath]; formatExists {
			formatID = existingFormatID
		} else {
			return fmt.Errorf("asset exists but format missing for %s", imagePath)
		}
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		_, err = tx.CreateAsset(assetID, imagePath, filepath.Base(imagePath), "0s", formatID)
		if err != nil {
			return fmt.Errorf("failed to create image asset: %v", err)
		}

		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return fmt.Errorf("failed to create image format: %v", err)
		}

		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}

	video := Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Image_%d", elementIndex),
	}

	if targetLane > 0 {
		// 🚨 FIXED: Spine elements cannot have lanes (per FCPXML validation rules)
	}

	if rand.Float32() < 0.3 {
		video.AdjustTransform = createMinimalAnimation(startTime, duration)
	}

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Videos = append(spine.Videos, video)

	return nil
}

// PngPileConfig holds configuration for PNG pile generation  
type PngPileConfig struct {
	Duration      float64 // Total duration in seconds
	TotalImages   int     // Number of images to download/use
	OutputDir     string  // Directory to store downloaded images
	PixabayAPIKey string  // Pixabay API key (optional)
	UseExisting   bool    // Use existing images in OutputDir instead of downloading
}

// GeneratePngPile creates a PNG pile effect similar to Info.fcpxml with base video and sliding PNGs
func GeneratePngPile(duration float64, totalImages int, inputDir string, verbose bool) (*FCPXML, error) {
	config := &PngPileConfig{
		Duration:    duration,
		TotalImages: totalImages,
		OutputDir:   inputDir,
		UseExisting: true, // Use existing files for backward compatibility
	}
	return GeneratePngPileWithConfig(config, verbose)
}

// GeneratePngPileWithConfig creates a PNG pile effect with full configuration options
func GeneratePngPileWithConfig(config *PngPileConfig, verbose bool) (*FCPXML, error) {
	if verbose {
		fmt.Printf("Generating PNG pile with %.1fs duration, %d images\n", config.Duration, config.TotalImages)
	}

	// Create base FCPXML structure
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		return nil, fmt.Errorf("failed to create base FCPXML: %v", err)
	}

	// Initialize resource management
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Track created assets to avoid duplicates
	createdAssets := make(map[string]string)
	createdFormats := make(map[string]string)

	// Add base video track (164240-830460859.mp4)
	videoPath := "164240-830460859.mp4"
	if verbose {
		fmt.Printf("Adding base video track: %s\n", videoPath)
	}

	ids := tx.ReserveIDs(2)
	videoAssetID := ids[0]
	videoFormatID := ids[1]

	// 🚨 CRITICAL FIX: 164240-830460859.mp4 is only 6 seconds, need multiple clips for full duration
	const videoClipDuration = 5.87 // Actual video duration in seconds (3523/600s from Info.fcpxml)
	
	// Use actual asset duration from Info.fcpxml (3523/600s ≈ 5.87 seconds)
	_, err = tx.CreateAsset(videoAssetID, videoPath, "164240-830460859", ConvertSecondsToFCPDuration(videoClipDuration), videoFormatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create base video asset: %v", err)
	}
	
	// Create video format with 24000 timebase to match project format (avoid validation error)
	_, err = tx.CreateFormatWithFrameDuration(videoFormatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		return nil, fmt.Errorf("failed to create video format: %v", err)
	}
	
	// Set format name to match Info.fcpxml
	if len(fcpxml.Resources.Formats) > 0 {
		for i := range fcpxml.Resources.Formats {
			if fcpxml.Resources.Formats[i].ID == videoFormatID {
				fcpxml.Resources.Formats[i].Name = "FFVideoFormat1080p2997"
				break
			}
		}
	}

	// Calculate how many video clips needed to cover full duration
	numClips := int(math.Ceil(config.Duration / videoClipDuration))
	if verbose {
		fmt.Printf("Creating %d video clips of %.2fs each to cover %.1fs total\n", numClips, videoClipDuration, config.Duration)
	}
	
	// Create multiple AssetClips back-to-back to repeat the 6-second video
	var videoClips []AssetClip
	currentOffset := 0.0
	
	for i := 0; i < numClips; i++ {
		// Calculate duration for this clip (last clip might be shorter)
		clipDuration := videoClipDuration
		if currentOffset + clipDuration > config.Duration {
			clipDuration = config.Duration - currentOffset
		}
		
		clip := AssetClip{
			Ref:       videoAssetID,
			Offset:    ConvertSecondsToFCPDuration(currentOffset),
			Name:      "164240-830460859",
			Duration:  ConvertSecondsToFCPDuration(clipDuration),
			Format:    videoFormatID,
			TCFormat:  "NDF",
			ConformRate: &ConformRate{
				ScaleEnabled: "0",
				SrcFrameRate: "29.97",
			},
			AdjustTransform: &AdjustTransform{
				Scale: "3.27127 3.27127", // Match Info.fcpxml scaling
			},
		}
		
		videoClips = append(videoClips, clip)
		currentOffset += clipDuration
		
		if verbose {
			fmt.Printf("  Clip %d: offset=%.2fs, duration=%.2fs\n", i+1, currentOffset-clipDuration, clipDuration)
		}
	}

	// Get or download PNG files
	var pngFiles []string
	if config.UseExisting {
		// Use existing files from directory
		pngFiles, err = getPngFiles(config.OutputDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get PNG files: %v", err)
		}
		if verbose {
			fmt.Printf("Found %d existing PNG files in %s\n", len(pngFiles), config.OutputDir)
		}
	} else {
		// Download themed images from Pixabay
		pngFiles, err = downloadThemedImagesForPile(config, verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to download themed images: %v", err)
		}
	}

	if len(pngFiles) == 0 {
		return nil, fmt.Errorf("no PNG files available")
	}

	// Limit to requested number of images
	if len(pngFiles) > config.TotalImages {
		pngFiles = pngFiles[:config.TotalImages]
	}

	if verbose {
		fmt.Printf("Using %d images for PNG pile\n", len(pngFiles))
	}

	// Create border effect like Info.fcpxml
	effectIDs := tx.ReserveIDs(1)
	borderEffectID := effectIDs[0]
	_, err = tx.CreateEffect(borderEffectID, "Simple Border", ".../Effects.localized/Stylize.localized/Simple Border.localized/Simple Border.moef")
	if err != nil {
		return nil, fmt.Errorf("failed to create border effect: %v", err)
	}

	// Calculate timing progression (starts slow, speeds up)
	imageTimings := calculateProgessiveTiming(len(pngFiles), config.Duration)

	// Add PNG images to the FIRST video clip only (like Info.fcpxml - only first clip has images)
	if len(videoClips) > 0 {
		firstClip := &videoClips[0] // Get reference to first clip
		
		for i, pngFile := range pngFiles {
			timing := imageTimings[i]
			
			if verbose && (i < 5 || i%10 == 0) {
				fmt.Printf("Adding PNG %d/%d: %s at %.2fs, lane %d\n", i+1, len(pngFiles), filepath.Base(pngFile), timing.startTime, i+1)
			}

			err = addSlidingPngImageToAssetClip(firstClip, tx, pngFile, timing, i, borderEffectID, verbose, createdAssets, createdFormats)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Failed to add PNG %s: %v\n", pngFile, err)
				}
				continue
			}
		}
	}

	// Get spine reference and add ALL video clips back-to-back (repeat 6s video for full duration)
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	for i, clip := range videoClips {
		spine.AssetClips = append(spine.AssetClips, clip)
		if verbose {
			fmt.Printf("Added video clip %d to spine\n", i+1)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// 🚨 CRITICAL: Set sequence duration to prevent "Invalid edit with no respective media" error
	fcpxml.Library.Events[0].Projects[0].Sequences[0].Duration = ConvertSecondsToFCPDuration(config.Duration)

	// 🚨 CRITICAL: VALIDATE COMPLIANCE (per CLAUDE.md)
	violations := ValidateClaudeCompliance(fcpxml)
	if len(violations) > 0 {
		return nil, fmt.Errorf("ERROR: validation failed with %d violations:\n%s", len(violations), strings.Join(violations, "\n"))
	}

	if verbose {
		fmt.Printf("Successfully generated PNG pile with %d images\n", len(pngFiles))
	}

	return fcpxml, nil
}

// ImageTiming represents when and how long an image appears
type ImageTiming struct {
	startTime float64
	duration  float64
}

// calculateProgessiveTiming calculates start times with increasing pace
func calculateProgessiveTiming(numImages int, totalDuration float64) []ImageTiming {
	timings := make([]ImageTiming, numImages)
	
	// Progressive acceleration: start slow, speed up exponentially
	totalWeight := 0.0
	for i := 0; i < numImages; i++ {
		// Weight decreases exponentially for faster pace later
		weight := math.Pow(0.7, float64(i)/10.0)
		totalWeight += weight
	}
	
	currentTime := 0.0
	// Each image persists from its start time until the end of the video (pile up effect)
	
	for i := 0; i < numImages; i++ {
		// Each image lasts from its start time until the end of the video (pile up effect)
		remainingDuration := totalDuration - currentTime
		timings[i] = ImageTiming{
			startTime: currentTime,
			duration:  remainingDuration,
		}
		
		// Calculate time until next image (gets shorter over time)
		weight := math.Pow(0.7, float64(i)/10.0)
		timeStep := (totalDuration * 0.8) * (weight / totalWeight)
		currentTime += timeStep
		
		// Don't go past the end
		if currentTime > totalDuration-1.0 { // Leave at least 1 second for the last image
			currentTime = totalDuration - 1.0
		}
	}
	
	return timings
}

// addSlidingPngImageToAssetClip adds a PNG as nested Video within AssetClip with lane assignment (like Info.fcpxml)
func addSlidingPngImageToAssetClip(baseClip *AssetClip, tx *ResourceTransaction, pngPath string, timing ImageTiming, index int, borderEffectID string, verbose bool, createdAssets, createdFormats map[string]string) error {
	// Create image asset if not exists
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[pngPath]; exists {
		assetID = existingAssetID
		formatID = createdFormats[pngPath]
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		_, err = tx.CreateAsset(assetID, pngPath, filepath.Base(pngPath), "0s", formatID)
		if err != nil {
			return fmt.Errorf("failed to create PNG asset: %v", err)
		}

		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "400", "300", "1-13-1")
		if err != nil {
			return fmt.Errorf("failed to create PNG format: %v", err)
		}

		createdAssets[pngPath] = assetID
		createdFormats[pngPath] = formatID
	}

	// Create sliding animation from random direction with rotation like Info.fcpxml
	slideAnimation := createSlidingAnimationWithRotation(timing.startTime, timing.duration, index)
	
	// Create video element for PNG nested within AssetClip (like Info.fcpxml pattern)
	video := Video{
		Ref:      assetID,
		Lane:     fmt.Sprintf("%d", index+1), // Lane assignment like Info.fcpxml: lane="1", lane="2", etc.
		Offset:   ConvertSecondsToFCPDuration(timing.startTime), // Offset relative to AssetClip start
		Duration: ConvertSecondsToFCPDuration(timing.duration),
		Name:     fmt.Sprintf("PNG_%d_%s", index+1, strings.TrimSuffix(filepath.Base(pngPath), filepath.Ext(pngPath))),
		Start:    "3600s", // Match Info.fcpxml start time
		AdjustTransform: slideAnimation,
		FilterVideos: []FilterVideo{
			{
				Ref:  borderEffectID,
				Name: "Simple Border",
				Params: []Param{
					{
						Name:  "Color",
						Key:   "9999/987171795/987171799/3/987171806/2",
						Value: "0 0 0 1", // Black border like Info.fcpxml
					},
				},
			},
		},
	}

	// Add PNG Video as nested element within the main AssetClip (like Info.fcpxml)
	baseClip.Videos = append(baseClip.Videos, video)

	return nil
}

// addSlidingPngImage adds a PNG with sliding animation and black border (legacy function, keeping for compatibility)
func addSlidingPngImage(spine *Spine, tx *ResourceTransaction, pngPath string, timing ImageTiming, index int, borderEffectID string, verbose bool, createdAssets, createdFormats map[string]string) error {
	// Create image asset if not exists
	var assetID, formatID string
	var err error

	if existingAssetID, exists := createdAssets[pngPath]; exists {
		assetID = existingAssetID
		formatID = createdFormats[pngPath]
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]

		_, err = tx.CreateAsset(assetID, pngPath, filepath.Base(pngPath), "0s", formatID)
		if err != nil {
			return fmt.Errorf("failed to create PNG asset: %v", err)
		}

		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "400", "300", "1-13-1")
		if err != nil {
			return fmt.Errorf("failed to create PNG format: %v", err)
		}

		createdAssets[pngPath] = assetID
		createdFormats[pngPath] = formatID
	}

	// Create sliding animation from random direction
	slideAnimation := createSlidingAnimation(timing.startTime, timing.duration, index)
	
	// Create video element for PNG (images use Video elements, not AssetClip)
	video := Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(timing.startTime),
		Duration: ConvertSecondsToFCPDuration(timing.duration),
		Name:     fmt.Sprintf("PNG_%d_%s", index+1, strings.TrimSuffix(filepath.Base(pngPath), filepath.Ext(pngPath))),
		// No Lane for spine elements - only connected clips can have lanes
		AdjustTransform: slideAnimation,
		FilterVideos: []FilterVideo{
			{
				Ref:  borderEffectID,
				Name: "Simple Border",
				Params: []Param{
					{
						Name:  "Color",
						Key:   "9999/987171795/987171799/3/987171806/2",
						Value: "0 0 0 1", // Black border like Info.fcpxml
					},
				},
			},
		},
	}

	// Add PNG Video directly to spine (images must be at spine level, not nested in AssetClips)
	spine.Videos = append(spine.Videos, video)

	return nil
}

// createSlidingAnimationWithRotation creates position animation with rotation from various directions (like Info.fcpxml)
func createSlidingAnimationWithRotation(startTime, duration float64, index int) *AdjustTransform {
	// Determine slide direction and rotation based on index
	directions := []struct{ startX, endX, startY, endY, rotation string }{
		{"62.5", "0", "0", "0", "16.02"},     // Right to center with rotation (like Info.fcpxml)
		{"-62.5", "0", "0", "0", "-26.6193"}, // Left to center with counter-rotation (like Info.fcpxml) 
		{"0", "0", "45", "0", "12.5"},        // Top to center
		{"0", "0", "-45", "0", "-15.3"},      // Bottom to center
		{"44.2", "0", "31.2", "0", "22.8"},   // Top-right diagonal
		{"-44.2", "0", "31.2", "0", "-18.4"}, // Top-left diagonal
		{"44.2", "0", "-31.2", "0", "14.7"},  // Bottom-right diagonal
		{"-44.2", "0", "-31.2", "0", "-21.1"}, // Bottom-left diagonal
	}
	
	direction := directions[index%len(directions)]
	
	return &AdjustTransform{
		Rotation: direction.rotation, // Add rotation like Info.fcpxml
		Params: []Param{
			{
				Name: "position",
				NestedParams: []Param{
					{
						Name: "X",
						Key:  "1",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{
									Time:  "3600s", // Match Info.fcpxml start time exactly
									Value: direction.startX,
								},
								{
									Time:  "2594882880/720000s", // Match Info.fcpxml end time exactly
									Value: direction.endX,
								},
							},
						},
					},
					{
						Name: "Y",
						Key:  "2",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{
									Time:  "3600s", // Match Info.fcpxml exactly - only one Y keyframe
									Value: direction.startY,
									Curve: "linear",
								},
							},
						},
					},
				},
			},
		},
	}
}

// createSlidingAnimation creates position animation from various directions (legacy function)
func createSlidingAnimation(startTime, duration float64, index int) *AdjustTransform {
	// Determine slide direction based on index
	directions := []struct{ startX, endX, startY, endY string }{
		{"62.5", "0", "0", "0"},     // Right to center (like Info.fcpxml)
		{"-62.5", "0", "0", "0"},    // Left to center (like Info.fcpxml) 
		{"0", "0", "45", "0"},       // Top to center
		{"0", "0", "-45", "0"},      // Bottom to center
		{"44.2", "0", "31.2", "0"},  // Top-right diagonal
		{"-44.2", "0", "31.2", "0"}, // Top-left diagonal
		{"44.2", "0", "-31.2", "0"}, // Bottom-right diagonal
		{"-44.2", "0", "-31.2", "0"}, // Bottom-left diagonal
	}
	
	direction := directions[index%len(directions)]
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				NestedParams: []Param{
					{
						Name: "X",
						Key:  "1",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{
									Time:  ConvertSecondsToFCPDuration(startTime),
									Value: direction.startX,
								},
								{
									Time:  ConvertSecondsToFCPDuration(startTime + 1.0), // 1 second slide
									Value: direction.endX,
								},
							},
						},
					},
					{
						Name: "Y",
						Key:  "2",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{
									Time:  ConvertSecondsToFCPDuration(startTime),
									Value: direction.startY,
									Curve: "linear",
								},
								{
									Time:  ConvertSecondsToFCPDuration(startTime + 1.0),
									Value: direction.endY,
									Curve: "linear",
								},
							},
						},
					},
				},
			},
		},
	}
}

// getPngFiles finds all PNG and JPG image files in the given directory
func getPngFiles(dir string) ([]string, error) {
	var pngFiles []string
	
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		if !d.IsDir() && (ext == ".png" || ext == ".jpg" || ext == ".jpeg") {
			pngFiles = append(pngFiles, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort for consistent ordering
	sort.Strings(pngFiles)
	
	return pngFiles, nil
}

// downloadThemedImagesForPile downloads themed images for PNG pile effect
func downloadThemedImagesForPile(config *PngPileConfig, verbose bool) ([]string, error) {
	if verbose {
		fmt.Printf("Downloading %d themed images to %s\n", config.TotalImages, config.OutputDir)
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Story-based theme progression like the original concept
	themes := []string{
		// Nature/peaceful start (15 images)
		"forest", "mountain", "lake", "sunrise", "flowers", 
		"meadow", "stream", "peaceful", "calm", "serenity",
		"butterfly", "bird", "deer", "waterfall", "garden",
		
		// Journey begins (15 images) 
		"path", "road", "journey", "adventure", "exploration",
		"compass", "map", "backpack", "hiking", "travel",
		"bridge", "stairs", "door", "gate", "horizon",
		
		// Action/movement (30 images)
		"running", "flying", "soaring", "eagle", "freedom",
		"wind", "motion", "speed", "racing", "energy",
		"waves", "ocean", "storm", "lightning", "fire",
		"explosion", "burst", "jump", "dance", "celebration",
		"festival", "party", "music", "concert", "lights",
		"fireworks", "rainbow", "color", "vibrant", "bright",
		
		// Discovery/wonder (15 images)
		"magic", "mystical", "galaxy", "stars", "universe",
		"crystal", "gem", "treasure", "ancient", "castle",
		"portal", "mystery", "wonder", "dream", "fantasy",
		
		// Resolution/peace (15 images)  
		"sunset", "tranquil", "harmony", "balance", "zen",
		"meditation", "reflection", "wisdom", "peace", "home",
		"family", "love", "happiness", "smile", "heart",
	}

	// Ensure we don't exceed available themes
	if config.TotalImages > len(themes) {
		// Repeat themes if needed
		originalLen := len(themes)
		for len(themes) < config.TotalImages {
			themes = append(themes, themes[len(themes)%originalLen])
		}
	}

	// Download images for each theme
	var allFiles []string
	imagesPerTheme := 1 // One image per theme word
	
	for i, theme := range themes[:config.TotalImages] {
		if verbose && (i < 5 || i%10 == 0) {
			fmt.Printf("Downloading theme %d/%d: %s\n", i+1, config.TotalImages, theme)
		}
		
		// Use existing Pixabay download function
		attributions, err := DownloadImagesFromPixabay(theme, imagesPerTheme, config.OutputDir, config.PixabayAPIKey)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Failed to download images for theme '%s': %v\n", theme, err)
			}
			continue
		}
		
		// Extract file paths
		for _, attr := range attributions {
			allFiles = append(allFiles, attr.FilePath)
		}
		
		// Stop if we have enough images
		if len(allFiles) >= config.TotalImages {
			break
		}
	}

	if verbose {
		fmt.Printf("Successfully downloaded %d themed images\n", len(allFiles))
	}

	return allFiles, nil
}
