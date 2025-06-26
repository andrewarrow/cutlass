package fcp

import (
	"fmt"

	"math/rand"
	"os"

	"path/filepath"

	"strings"
	"time"
)

// AddPipVideo adds a video as picture-in-picture (PIP) to an existing FCPXML file.
//
// 🚨 CRITICAL PIP Requirements (see CLAUDE.md for details):
//  1. **Format Compatibility**: Creates separate formats for main and PIP videos (different from sequence)
//     to allow conform-rate elements without causing "Encountered an unexpected value" FCP errors
//  2. **Layering Strategy**: PIP video uses lane="-1" (background), main video becomes corner overlay
//  3. **Shape Mask Application**: Applied to main video for rounded corners on the small corner video
//
// Structure Generated (matches samples/pip.fcpxml):
// ```
// <asset-clip ref="main" format="r5"> <!-- Main: new format enables conform-rate -->
//
//	<conform-rate scaleEnabled="0"/>
//	<adjust-crop mode="trim">...</adjust-crop>
//	<adjust-transform position="60.3234 -35.9353" scale="0.28572 0.28572"/> <!-- Corner -->
//	<asset-clip ref="pip" lane="-1" format="r4"> <!-- PIP: background full-size -->
//	    <conform-rate scaleEnabled="0" srcFrameRate="60"/>
//	</asset-clip>
//	<filter-video name="Shape Mask">...</filter-video> <!-- Rounded corners -->
//
// </asset-clip>
// ```
//
// Visual Result: Small rounded corner video (main) overlaid on full-size background (PIP).
//
// 🚨 CLAUDE.md Rules Applied:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → proper XML marshaling via struct fields
// - Atomic ID reservation prevents race conditions and ID collisions
// - Frame-aligned durations → ConvertSecondsToFCPDuration() function
// - UID consistency → GenerateUID() for deterministic unique identifiers
//
// ❌ NEVER: fmt.Sprintf("<asset-clip ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddPipVideo(fcpxml *FCPXML, pipVideoPath string, offsetSeconds float64) error {

	registry := NewResourceRegistry(fcpxml)

	// Variable to store main video format ID for PIP mode
	var mainVideoFormatID string

	// Check if asset already exists for this file
	var pipAsset *Asset
	if asset, exists := registry.GetOrCreateAsset(pipVideoPath); exists {
		pipAsset = asset

		mainVideoFormatID = ""
	} else {

		tx := NewTransaction(registry)

		absPath, err := filepath.Abs(pipVideoPath)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			tx.Rollback()
			return fmt.Errorf("PIP video file does not exist: %s", absPath)
		}

		ids := tx.ReserveIDs(3)
		assetID := ids[0]
		formatID := ids[1]
		mainFormatID := ids[2]

		videoName := strings.TrimSuffix(filepath.Base(pipVideoPath), filepath.Ext(pipVideoPath))

		defaultDurationSeconds := 10.0
		frameDuration := ConvertSecondsToFCPDuration(defaultDurationSeconds)

		_, err = tx.CreateFormatWithFrameDuration(formatID, "100/6000s", "2336", "1510", "1-1-1 (Rec. 709)")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create PIP video format: %v", err)
		}

		_, err = tx.CreateFormatWithFrameDuration(mainFormatID, "13335/400000s", "1920", "1080", "1-1-1 (Rec. 709)")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create main video format for PIP: %v", err)
		}

		asset, err := tx.CreateVideoOnlyAsset(assetID, absPath, videoName, frameDuration, formatID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create PIP asset: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		pipAsset = asset

		mainVideoFormatID = mainFormatID
	}

	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return fmt.Errorf("no sequence found in FCPXML")
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	if len(sequence.Spine.AssetClips) == 0 {
		return fmt.Errorf("no asset-clip found in spine to add PIP to - need at least one video in the sequence")
	}

	mainClip := &sequence.Spine.AssetClips[0]

	if mainVideoFormatID != "" {
		mainClip.Format = mainVideoFormatID
	}

	shapeMaskEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if effect.UID == "FFSuperEllipseMask" {
			shapeMaskEffectID = effect.ID
			break
		}
	}

	if shapeMaskEffectID == "" {

		tx := NewTransaction(registry)
		ids := tx.ReserveIDs(1)
		shapeMaskEffectID = ids[0]

		_, err := tx.CreateEffect(shapeMaskEffectID, "Shape Mask", "FFSuperEllipseMask")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create Shape Mask effect: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit Shape Mask effect: %v", err)
		}
	}

	pipOffsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)

	pipDurationSeconds := 10.0
	pipDuration := ConvertSecondsToFCPDuration(pipDurationSeconds)

	pipClip := AssetClip{
		Ref:      pipAsset.ID,
		Lane:     "-1",
		Offset:   pipOffsetDuration,
		Name:     pipAsset.Name,
		Start:    "67300/3000s",
		Duration: pipDuration,
		Format:   pipAsset.Format,
		TCFormat: "NDF",
		ConformRate: &ConformRate{
			ScaleEnabled: "0",
			SrcFrameRate: "60",
		},
	}

	mainClip.ConformRate = &ConformRate{
		ScaleEnabled: "0",
	}

	mainClip.AdjustCrop = &AdjustCrop{
		Mode: "trim",
		TrimRect: &TrimRect{
			Left:   "27.1921",
			Right:  "27.1001",
			Bottom: "12.6562",
		},
	}

	mainClip.AdjustTransform = &AdjustTransform{
		Position: "60.3234 -35.9353",
		Scale:    "0.28572 0.28572",
	}

	mainClip.FilterVideos = []FilterVideo{
		{
			Ref:  shapeMaskEffectID,
			Name: "Shape Mask",
			Params: []Param{
				{
					Name:  "Radius",
					Key:   "160",
					Value: "305 190.625",
				},
				{
					Name:  "Curvature",
					Key:   "159",
					Value: "0.3695",
				},
				{
					Name:  "Feather",
					Key:   "102",
					Value: "100",
				},
				{
					Name:  "Falloff",
					Key:   "158",
					Value: "-100",
				},
				{
					Name:  "Input Size",
					Key:   "205",
					Value: "1920 1080",
				},
				{
					Name: "Transforms",
					Key:  "200",
					NestedParams: []Param{
						{
							Name:  "Scale",
							Key:   "203",
							Value: "1.3449 1.9525",
						},
					},
				},
			},
		},
	}

	mainClip.NestedAssetClips = append(mainClip.NestedAssetClips, pipClip)

	return nil
}

// GenerateBaffleTimeline creates a random complex timeline for stress testing
func GenerateBaffleTimeline(minDuration, maxDuration float64, verbose bool) (*FCPXML, error) {

	fcpxml, err := GenerateEmpty("")
	if err != nil {
		return nil, fmt.Errorf("failed to create base FCPXML: %v", err)
	}

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	if verbose {
		fmt.Printf("Starting baffle timeline generation...\n")
	}

	assetsDir := "./assets"
	assets, err := scanAssetsDirectory(assetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan assets directory: %v", err)
	}

	if len(assets.Images) == 0 && len(assets.Videos) == 0 {
		return nil, fmt.Errorf("no assets found in %s directory", assetsDir)
	}

	if verbose {
		fmt.Printf("Found %d images and %d videos in assets directory\n", len(assets.Images), len(assets.Videos))
	}

	rand.Seed(time.Now().UnixNano())
	totalDuration := minDuration + rand.Float64()*(maxDuration-minDuration)

	if verbose {
		fmt.Printf("Target timeline duration: %.1f seconds (%.1f minutes)\n", totalDuration, totalDuration/60)
	}

	err = generateRandomTimelineElements(fcpxml, tx, assets, totalDuration, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to generate timeline elements: %v", err)
	}

	updateSequenceDuration(fcpxml, totalDuration)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	if verbose {
		fmt.Printf("Baffle timeline generation completed successfully\n")
	}

	return fcpxml, nil
}

// scanAssetsDirectory scans ./assets for available media files
func scanAssetsDirectory(assetsDir string) (*AssetCollection, error) {
	assets := &AssetCollection{
		Images: []string{},
		Videos: []string{},
	}

	absAssetsDir, err := filepath.Abs(assetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for assets directory: %v", err)
	}

	if _, err := os.Stat(absAssetsDir); os.IsNotExist(err) {
		return assets, nil
	}

	entries, err := os.ReadDir(absAssetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read assets directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		absolutePath := filepath.Join(absAssetsDir, filename)

		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif":
			assets.Images = append(assets.Images, absolutePath)
		case ".mp4", ".mov", ".avi", ".mkv":
			assets.Videos = append(assets.Videos, absolutePath)
		}
	}

	return assets, nil
}

// generateRandomTimelineElements fills the timeline with random elements
func generateRandomTimelineElements(fcpxml *FCPXML, tx *ResourceTransaction, assets *AssetCollection, totalDuration float64, verbose bool) error {

	createdAssets := make(map[string]string)
	createdFormats := make(map[string]string)

	if len(assets.Videos) > 0 {
		backgroundVideo := assets.Videos[rand.Intn(len(assets.Videos))]

		uniqueVideo, err := createUniqueMediaCopy(backgroundVideo, "background")
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create unique background copy: %v\n", err)
			uniqueVideo = backgroundVideo
		}

		err = addRandomVideoElement(fcpxml, tx, uniqueVideo, 0.0, totalDuration, 0, 0, verbose, createdAssets, createdFormats)
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to add background video: %v\n", err)
		} else if verbose {
			fmt.Printf("  Added background video: %s (%.1fs @ 0s)\n", filepath.Base(uniqueVideo), totalDuration)
		}
	} else if len(assets.Images) > 0 {
		backgroundImage := assets.Images[rand.Intn(len(assets.Images))]

		uniqueImage, err := createUniqueMediaCopy(backgroundImage, "background")
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create unique background copy: %v\n", err)
			uniqueImage = backgroundImage
		}

		err = addRandomImageElement(fcpxml, tx, uniqueImage, 0.0, totalDuration, 0, 0, verbose, createdAssets, createdFormats)
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to add background image: %v\n", err)
		} else if verbose {
			fmt.Printf("  Added background image: %s (%.1fs @ 0s)\n", filepath.Base(uniqueImage), totalDuration)
		}
	}

	if verbose {
		fmt.Printf("Creating nested multi-lane structure with overlays...\n")
	}

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine

	if len(assets.Videos) > 0 {

		mainVideoPath := assets.Videos[rand.Intn(len(assets.Videos))]
		uniqueMainVideo, err := createUniqueMediaCopy(mainVideoPath, "main_bg")
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create unique main video copy: %v\n", err)
			uniqueMainVideo = mainVideoPath
		}

		mainVideo, err := createNestedVideoElement(fcpxml, tx, uniqueMainVideo, totalDuration, verbose, assets, createdAssets, createdFormats)
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create main video element: %v\n", err)
		} else {
			spine.Videos = append(spine.Videos, *mainVideo)
		}
	}

	if verbose {
		fmt.Printf("DEBUG: About to create additional main spine elements...\n")
	}

	numMainElements := 3 + rand.Intn(5)
	maxLanes := 8
	currentOffset := totalDuration * 0.2

	if verbose {
		fmt.Printf("Creating %d additional main spine elements with lane distribution...\n", numMainElements)
	}

	for i := 1; i <= numMainElements; i++ {
		duration := 6.0 + rand.Float64()*15.0
		startTime := currentOffset
		currentOffset += duration * 0.4

		if startTime >= totalDuration {
			break
		}

		if startTime+duration > totalDuration {
			duration = totalDuration - startTime
		}

		lane := (i % maxLanes) + 1

		if i%2 == 0 && len(assets.Videos) > 0 {
			videoPath := assets.Videos[rand.Intn(len(assets.Videos))]
			uniqueVideo, err := createUniqueMediaCopy(videoPath, fmt.Sprintf("main_%d", i))
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create unique video copy: %v\n", err)
				uniqueVideo = videoPath
			}

			mainElement, err := createLaneAssetClipElement(fcpxml, tx, uniqueVideo, startTime, duration, lane, i, verbose, createdAssets, createdFormats)
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create lane asset-clip: %v\n", err)
			} else {
				spine.AssetClips = append(spine.AssetClips, *mainElement)
				if verbose {
					fmt.Printf("  Added lane %d video: %s (%.1fs @ %.1fs)\n", lane, filepath.Base(uniqueVideo), duration, startTime)
				}
			}
		} else if len(assets.Images) > 0 {
			imagePath := assets.Images[rand.Intn(len(assets.Images))]
			uniqueImage, err := createUniqueMediaCopy(imagePath, fmt.Sprintf("main_img_%d", i))
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create unique image copy: %v\n", err)
				uniqueImage = imagePath
			}

			mainElement, err := createLaneImageElement(fcpxml, tx, uniqueImage, startTime, duration, lane, i, verbose, createdAssets, createdFormats)
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create lane image: %v\n", err)
			} else {
				spine.Videos = append(spine.Videos, *mainElement)
				if verbose {
					fmt.Printf("  Added lane %d image: %s (%.1fs @ %.1fs)\n", lane, filepath.Base(uniqueImage), duration, startTime)
				}
			}
		}
	}

	return nil
}

// createImageOverlay creates an image overlay element with proper positioning
func createImageOverlay(fcpxml *FCPXML, tx *ResourceTransaction, imagePath string, startTime, duration float64, lane, index int, verbose bool, createdAssets, createdFormats map[string]string) (*Video, error) {
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
		Name:     fmt.Sprintf("ImageOverlay_%d", index),
		Lane:     fmt.Sprintf("%d", lane),
		AdjustTransform: &AdjustTransform{
			Position: "0 0",
			Scale:    fmt.Sprintf("%.2f %.2f", 0.5+rand.Float64()*0.3, 0.5+rand.Float64()*0.3),
		},
	}

	return video, nil
}

// createVideoOverlay creates a video overlay element with proper positioning
func createVideoOverlay(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, startTime, duration float64, lane, index int, verbose bool, createdAssets, createdFormats map[string]string) (*AssetClip, error) {
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
		Name:     fmt.Sprintf("VideoOverlay_%d", index),
		Lane:     fmt.Sprintf("%d", lane),
		AdjustTransform: &AdjustTransform{
			Position: "0 0",
			Scale:    fmt.Sprintf("%.2f %.2f", 0.6+rand.Float64()*0.3, 0.6+rand.Float64()*0.3),
		},
	}

	return assetClip, nil
}

// createTextOverlay creates a text overlay element
func createTextOverlay(fcpxml *FCPXML, tx *ResourceTransaction, startTime, duration float64, lane, index int, verbose bool) (*Title, error) {

	ids := tx.ReserveIDs(1)
	effectID := ids[0]

	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		return nil, fmt.Errorf("failed to create text effect: %v", err)
	}

	textContent := generateRandomText()
	styleID := fmt.Sprintf("ts_%d", rand.Intn(999999)+100000)

	title := &Title{
		Ref:      effectID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("TextOverlay_%d", index),
		Lane:     fmt.Sprintf("%d", lane),
		Text: &TitleText{
			TextStyles: []TextStyleRef{{
				Ref:  styleID,
				Text: textContent,
			}},
		},
		TextStyleDefs: []TextStyleDef{{
			ID: styleID,
			TextStyle: TextStyle{
				Font:        randomFont(),
				FontSize:    fmt.Sprintf("%.0f", 240+rand.Float64()*240),
				FontColor:   randomColor(),
				Alignment:   randomAlignment(),
				LineSpacing: "1.2",
			},
		}},
		Params: []Param{{
			Name:  "Opacity",
			Value: fmt.Sprintf("%.2f", 0.8+rand.Float64()*0.2),
		}},
	}

	return title, nil
}
