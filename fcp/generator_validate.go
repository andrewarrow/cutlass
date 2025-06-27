package fcp

import (
	"encoding/xml"
	"fmt"

	"os"

	"path/filepath"
	"strconv"
	"strings"
)

// ValidateClaudeCompliance performs automated checks for CLAUDE.md rule compliance.
//
// 🚨 CLAUDE.md Validation - Run this before any commit!
// This function helps catch violations of critical rules in CLAUDE.md
func ValidateClaudeCompliance(fcpxml *FCPXML) []string {
	var violations []string

	idMap := make(map[string]bool)

	for _, asset := range fcpxml.Resources.Assets {
		if idMap[asset.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Asset)", asset.ID))
		}
		idMap[asset.ID] = true
	}

	for _, format := range fcpxml.Resources.Formats {
		if idMap[format.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Format)", format.ID))
		}
		idMap[format.ID] = true
	}

	for _, effect := range fcpxml.Resources.Effects {
		if idMap[effect.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Effect)", effect.ID))
		}
		idMap[effect.ID] = true
	}

	for _, media := range fcpxml.Resources.Media {
		if idMap[media.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Media)", media.ID))
		}
		idMap[media.ID] = true
	}

	checkDuration := func(duration, location string) {
		if strings.Contains(duration, "/600s") && !strings.Contains(duration, "1001") {
			violations = append(violations, fmt.Sprintf("Potentially non-frame-aligned duration '%s' at %s - use ConvertSecondsToFCPDuration()", duration, location))
		}
		if strings.Contains(duration, "/24000s") && duration != "0s" {

			durationNoS := strings.TrimSuffix(duration, "s")
			parts := strings.Split(durationNoS, "/")
			if len(parts) == 2 {
				if numerator, err := strconv.Atoi(parts[0]); err == nil {

					if numerator%1001 != 0 {
						violations = append(violations, fmt.Sprintf("Non-frame-aligned duration '%s' at %s - must be (frames*1001)/24000s", duration, location))
					}
				}
			}
		}
	}

	for _, asset := range fcpxml.Resources.Assets {
		checkDuration(asset.Duration, fmt.Sprintf("Asset %s", asset.ID))
	}

	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {
				checkDuration(sequence.Duration, fmt.Sprintf("Sequence in Project %s", project.Name))

				for _, clip := range sequence.Spine.AssetClips {
					checkDuration(clip.Duration, fmt.Sprintf("AssetClip %s in Spine", clip.Name))
				}
			}
		}
	}

	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {

				for _, clip := range sequence.Spine.AssetClips {
					// Find the referenced asset
					var referencedAsset *Asset
					for i := range fcpxml.Resources.Assets {
						if fcpxml.Resources.Assets[i].ID == clip.Ref {
							referencedAsset = &fcpxml.Resources.Assets[i]
							break
						}
					}

					if referencedAsset != nil && clip.Format != referencedAsset.Format {
						violations = append(violations, fmt.Sprintf("Format mismatch: AssetClip '%s' has format '%s' but its referenced asset has format '%s' - asset-clips must use asset format", clip.Name, clip.Format, referencedAsset.Format))
					}
				}
			}
		}
	}

	fictionalEffectUIDs := map[string]bool{
		"FFParticleSystem": true,
		"FFReplicator":     true,
		"FFGravity":        true,
		"FFWind":           true,
		"FFEmitter":        true,
		"FFAttractor":      true,
		"FFMotion":         true,
		"FFTransform":      true,
		"FFColorize":       true,
		"FFTurbulence":     true,
		"FFWave":           true,
		"FFSpiral":         true,
		"FFAnimatedText":   true,
		"FFDistortion":     true,
	}

	for _, effect := range fcpxml.Resources.Effects {
		if fictionalEffectUIDs[effect.UID] {
			violations = append(violations, fmt.Sprintf("Fictional effect UID '%s' detected in effect '%s' - use built-in adjust-* elements instead", effect.UID, effect.Name))
		}
	}

	validateKeyframes := func(keyframes []Keyframe, location string) {
		for i, keyframe := range keyframes {

			if keyframe.Interp != "" {
				validInterps := map[string]bool{"linear": true, "ease": true, "easeIn": true, "easeOut": true}
				if !validInterps[keyframe.Interp] {
					violations = append(violations, fmt.Sprintf("Invalid keyframe interp '%s' at %s[%d] - must be: linear, ease, easeIn, easeOut", keyframe.Interp, location, i))
				}
			}

			if keyframe.Curve != "" {
				validCurves := map[string]bool{"linear": true, "smooth": true}
				if !validCurves[keyframe.Curve] {
					violations = append(violations, fmt.Sprintf("Invalid keyframe curve '%s' at %s[%d] - must be: linear, smooth", keyframe.Curve, location, i))
				}
			}
		}
	}

	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {

				for _, clip := range sequence.Spine.AssetClips {

					if clip.AdjustTransform != nil {
						for _, param := range clip.AdjustTransform.Params {
							if param.KeyframeAnimation != nil {
								validateKeyframes(param.KeyframeAnimation.Keyframes, fmt.Sprintf("AssetClip '%s' AdjustTransform param '%s'", clip.Name, param.Name))
							}
						}
					}

					for _, filter := range clip.FilterVideos {
						for _, param := range filter.Params {
							if param.KeyframeAnimation != nil {
								validateKeyframes(param.KeyframeAnimation.Keyframes, fmt.Sprintf("AssetClip '%s' FilterVideo '%s' param '%s'", clip.Name, filter.Name, param.Name))
							}
						}
					}
				}

				for _, video := range sequence.Spine.Videos {
					if video.AdjustTransform != nil {
						for _, param := range video.AdjustTransform.Params {
							if param.KeyframeAnimation != nil {
								validateKeyframes(param.KeyframeAnimation.Keyframes, fmt.Sprintf("Video '%s' AdjustTransform param '%s'", video.Name, param.Name))
							}
						}
					}
				}

				for _, title := range sequence.Spine.Titles {
					for _, param := range title.Params {
						if param.KeyframeAnimation != nil {
							validateKeyframes(param.KeyframeAnimation.Keyframes, fmt.Sprintf("Title '%s' param '%s'", title.Name, param.Name))
						}
					}
				}
			}
		}
	}

	resourceIDs := make(map[string]bool)
	for _, asset := range fcpxml.Resources.Assets {
		resourceIDs[asset.ID] = true
	}
	for _, format := range fcpxml.Resources.Formats {
		resourceIDs[format.ID] = true
	}
	for _, effect := range fcpxml.Resources.Effects {
		resourceIDs[effect.ID] = true
	}
	for _, media := range fcpxml.Resources.Media {
		resourceIDs[media.ID] = true
	}

	checkRef := func(ref, elementType string) {
		if ref != "" && !resourceIDs[ref] {
			violations = append(violations, fmt.Sprintf("Undefined reference '%s' in %s - missing resource definition", ref, elementType))
		}
	}

	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {

				for _, clip := range sequence.Spine.AssetClips {
					checkRef(clip.Ref, fmt.Sprintf("AssetClip '%s'", clip.Name))

					for _, filter := range clip.FilterVideos {
						checkRef(filter.Ref, fmt.Sprintf("FilterVideo '%s' in AssetClip '%s'", filter.Name, clip.Name))
					}
				}

				for _, video := range sequence.Spine.Videos {
					checkRef(video.Ref, fmt.Sprintf("Video '%s'", video.Name))
				}

				for _, title := range sequence.Spine.Titles {
					checkRef(title.Ref, fmt.Sprintf("Title '%s'", title.Name))
				}
			}
		}
	}

	return violations
}

// isImageFile checks if the given file is an image (PNG, JPG, JPEG).
//
// 🚨 CLAUDE.md Rule: Image vs Video Asset Properties
// - Image files should NOT have audio properties (HasAudio, AudioSources, AudioChannels)
// - Image files MUST have VideoSources = "1"
// - Duration is set by caller, not hardcoded to "0s"
func isImageFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

// AddImage adds an image asset and asset-clip to the FCPXML structure.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Assets, sequence.Spine.AssetClips
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function
// - Maintains UID consistency → generateUID() function for deterministic UIDs
// - Image-specific properties → VideoSources="1", NO audio properties (HasAudio, AudioSources, AudioChannels)
//
// ❌ NEVER: fmt.Sprintf("<asset-clip ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddImage(fcpxml *FCPXML, imagePath string, durationSeconds float64) error {
	return AddImageWithSlide(fcpxml, imagePath, durationSeconds, false)
}

// AddImageWithSlide adds an image asset with optional slide animation to the FCPXML structure.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Assets, sequence.Spine.Videos
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function
// - Maintains UID consistency → generateUID() function for deterministic UIDs
// - Image-specific properties → VideoSources="1", NO audio properties (HasAudio, AudioSources, AudioChannels)
// - Keyframe animations → Uses AdjustTransform with KeyframeAnimation structs
//
// ❌ NEVER: fmt.Sprintf("<asset-clip ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddImageWithSlide(fcpxml *FCPXML, imagePath string, durationSeconds float64, withSlide bool) error {
	return AddImageWithSlideAndFormat(fcpxml, imagePath, durationSeconds, withSlide, "horizontal")
}

func AddImageWithSlideAndFormat(fcpxml *FCPXML, imagePath string, durationSeconds float64, withSlide bool, format string) error {

	if !isImageFile(imagePath) {
		return fmt.Errorf("file is not a supported image format (PNG, JPG, JPEG): %s", imagePath)
	}

	registry := NewResourceRegistry(fcpxml)

	if asset, exists := registry.GetOrCreateAsset(imagePath); exists {

		return addImageAssetClipToSpineWithFormat(fcpxml, asset, durationSeconds, withSlide, format)
	}

	tx := NewTransaction(registry)

	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		tx.Rollback()
		return fmt.Errorf("image file does not exist: %s", absPath)
	}

	ids := tx.ReserveIDs(2)
	assetID := ids[0]
	formatID := ids[1]

	imageName := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))

	frameDuration := ConvertSecondsToFCPDuration(durationSeconds)

	// Set format dimensions based on format type
	var width, height string
	switch format {
	case "vertical":
		width, height = "1080", "1920"
	case "horizontal":
		fallthrough
	default:
		width, height = "1280", "720"
	}

	_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", width, height, "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create image format: %v", err)
	}

	asset, err := tx.CreateAsset(assetID, absPath, imageName, frameDuration, formatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return addImageAssetClipToSpineWithFormat(fcpxml, asset, durationSeconds, withSlide, format)
}

// addImageAssetClipToSpine adds an image Video element to the sequence spine
// 🚨 CRITICAL FIX: Images use Video elements, NOT AssetClip elements
// Analysis of working samples/png.fcpxml shows images use <video> in spine
// This prevents addAssetClip:toObject:parentFormatID crashes in FCP
func addImageAssetClipToSpine(fcpxml *FCPXML, asset *Asset, durationSeconds float64) error {
	return addImageAssetClipToSpineWithFormat(fcpxml, asset, durationSeconds, false, "horizontal")
}

// addImageAssetClipToSpineWithSlide adds an image Video element to the sequence spine with optional slide animation
// 🚨 CRITICAL FIX: Images use Video elements, NOT AssetClip elements
// Analysis of working samples/png.fcpxml shows images use <video> in spine
// This prevents addAssetClip:toObject:parentFormatID crashes in FCP
// Keyframe animations match samples/slide.fcpxml pattern for sliding motion
func addImageAssetClipToSpineWithSlide(fcpxml *FCPXML, asset *Asset, durationSeconds float64, withSlide bool) error {
	return addImageAssetClipToSpineWithFormat(fcpxml, asset, durationSeconds, withSlide, "horizontal")
}

// addImageAssetClipToSpineWithFormat adds an image Video element to the sequence spine with format-aware scaling
func addImageAssetClipToSpineWithFormat(fcpxml *FCPXML, asset *Asset, durationSeconds float64, withSlide bool, format string) error {

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		currentTimelineDuration := calculateTimelineDuration(sequence)

		clipDuration := ConvertSecondsToFCPDuration(durationSeconds)

		video := Video{
			Ref:      asset.ID,
			Offset:   currentTimelineDuration,
			Name:     asset.Name,
			Start:    "86399313/24000s",
			Duration: clipDuration,
		}

		if withSlide {
			// Use enhanced Ken Burns with both crop and transform for vertical format
			if format == "vertical" {
				adjustCrop, adjustTransform := createEnhancedKenBurnsWithFormat(currentTimelineDuration, durationSeconds, format)
				video.AdjustCrop = adjustCrop
				video.AdjustTransform = adjustTransform
			} else {
				video.AdjustTransform = createKenBurnsAnimationWithFormat(currentTimelineDuration, durationSeconds, format)
			}
		} else {
			// Add zoom scaling for vertical format to fill frame with no empty space
			if format == "vertical" {
				video.AdjustTransform = &AdjustTransform{
					Scale: "1.8 1.8", // Zoom in to fill vertical frame
				}
			}
		}

		sequence.Spine.Videos = append(sequence.Spine.Videos, video)

		newTimelineDuration := addDurations(currentTimelineDuration, clipDuration)
		sequence.Duration = newTimelineDuration
	}

	return nil
}

// ReadFromFile parses an existing FCPXML file into structs.
//
// 🚨 CLAUDE.md Rule: ALWAYS use structs for XML parsing
// - Reads FCPXML file and unmarshals into struct representation
// - Maintains all existing resources and timeline structure
// - Use this before AddVideo/AddImage to preserve existing content
func ReadFromFile(filename string) (*FCPXML, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	// Parse XML into struct
	var fcpxml FCPXML
	err = xml.Unmarshal(data, &fcpxml)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML from %s: %v", filename, err)
	}

	return &fcpxml, nil
}
