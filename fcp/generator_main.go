package fcp

import (
	"fmt"

	"math/rand"

	"path/filepath"
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
		Lane:     fmt.Sprintf("%d", lane),
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
		Lane:     fmt.Sprintf("%d", lane),
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

	numOverlays := 6 + rand.Intn(8)

	for i := 1; i <= numOverlays; i++ {
		overlayStartTime := rand.Float64() * (duration * 0.8)
		overlayDuration := 3.0 + rand.Float64()*8.0

		if overlayStartTime+overlayDuration > duration {
			overlayDuration = duration - overlayStartTime
		}

		lane := 1 + rand.Intn(4)

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
		video.Lane = fmt.Sprintf("%d", targetLane)
	}

	if rand.Float32() < 0.3 {
		video.AdjustTransform = createMinimalAnimation(startTime, duration)
	}

	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Videos = append(spine.Videos, video)

	return nil
}
