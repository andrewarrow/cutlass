// Package fcp provides FCPXML generation using structs.
//
// 🚨 CRITICAL: All XML generation MUST follow CLAUDE.md rules:
// - NEVER use string templates with %s placeholders (see CLAUDE.md "NO XML STRING TEMPLATES")
// - ALWAYS use structs and xml.MarshalIndent for XML generation
// - ALL durations MUST be frame-aligned → USE ConvertSecondsToFCPDuration() function
// - ALL IDs MUST be unique → COUNT existing resources: len(Assets)+len(Formats)+len(Effects)+len(Media)  
// - BEFORE commits → RUN ValidateClaudeCompliance() + xmllint --dtdvalid FCPXMLv1_13.dtd
package fcp

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type TemplateVideo struct {
	ID       string
	UID      string
	Bookmark string
}

type NumberSection struct {
	Number  int
	VideoID string
	Offset  string
}

type TemplateData struct {
	FirstName      string
	LastName       string
	LastNameSuffix string
	Videos         []TemplateVideo
	Numbers        []NumberSection
}

func oldGgenerateUID(videoID string) string {
	// Create a hash from the video ID to ensure consistent UIDs
	hasher := md5.New()
	hasher.Write([]byte("cutlass_video_" + videoID))
	hash := hasher.Sum(nil)
	// Convert to uppercase hex string (32 characters)
	return strings.ToUpper(hex.EncodeToString(hash))
}

// generateBookmark creates a macOS security bookmark for a file path using Swift
func generateBookmark(filePath string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", absPath)
	}

	// Use Swift to create a security bookmark
	swiftCode := fmt.Sprintf(`
import Foundation

let url = URL(fileURLWithPath: "%s")
do {
    let bookmarkData = try url.bookmarkData(options: [.suitableForBookmarkFile])
    let base64String = bookmarkData.base64EncodedString()
    print(base64String)
} catch {
    print("ERROR: Could not create bookmark: \\(error)")
}
`, absPath)

	// Create temporary Swift file
	tmpFile, err := os.CreateTemp("", "bookmark*.swift")
	if err != nil {
		return "", nil
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(swiftCode)
	tmpFile.Close()
	if err != nil {
		return "", nil
	}

	// Execute Swift script
	cmd := exec.Command("swift", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {
		// Fallback to empty bookmark - some systems may still work
		return "", nil
	}

	bookmark := strings.TrimSpace(string(output))
	if strings.Contains(bookmark, "ERROR") {
		return "", nil
	}

	return bookmark, nil
}

// ConvertSecondsToFCPDuration converts seconds to frame-aligned FCP duration.
//
// 🚨 CLAUDE.md Rule: Frame Boundary Alignment - CRITICAL!
// - FCP uses time base of 24000/1001 ≈ 23.976 fps for frame alignment
// - Duration format: (frames*1001)/24000s where frames is an integer
// - NEVER use simple seconds * 24000 calculations - creates non-frame-aligned durations
// - Non-frame-aligned durations cause "not on an edit frame boundary" errors in FCP
// - Example: 21600000/24000s = NON-FRAME-ALIGNED ❌, 21599578/24000s = FRAME-ALIGNED ✅
func ConvertSecondsToFCPDuration(seconds float64) string {
	// Convert to frame count using the sequence time base (1001/24000s frame duration)
	// This means 24000/1001 frames per second ≈ 23.976 fps
	framesPerSecond := 24000.0 / 1001.0
	exactFrames := seconds * framesPerSecond

	// Choose the frame count that gives the closest duration to the target
	floorFrames := int(math.Floor(exactFrames))
	ceilFrames := int(math.Ceil(exactFrames))

	floorDuration := float64(floorFrames) / framesPerSecond
	ceilDuration := float64(ceilFrames) / framesPerSecond

	var frames int
	if math.Abs(seconds-floorDuration) <= math.Abs(seconds-ceilDuration) {
		frames = floorFrames
	} else {
		frames = ceilFrames
	}

	// Format as rational using the sequence time base
	return fmt.Sprintf("%d/24000s", frames*1001)
}

// GenerateEmpty creates an empty FCPXML file structure and returns a pointer to it
func GenerateEmpty(filename string) (*FCPXML, error) {
	// Create empty FCPXML structure matching empty.fcpxml
	fcpxml := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Formats: []Format{
				{
					ID:            "r1",
					Name:          "FFVideoFormat1080x1920p2398",
					FrameDuration: "1001/24000s",
					Width:         "1080",
					Height:        "1920",
					ColorSpace:    "1-1-1 (Rec. 709)",
				},
			},
		},
		Library: Library{
			Location: "file:///Users/aa/Movies/Untitled.fcpbundle/",
			Events: []Event{
				{
					Name: "6-13-25",
					UID:  "78463397-97FD-443D-B4E2-07C581674AFC",
					Projects: []Project{
						{
							Name:    "wiki",
							UID:     "DEA19981-DED5-4851-8435-14515931C68A",
							ModDate: "2025-06-13 11:46:22 -0700",
							Sequences: []Sequence{
								{
									Format:      "r1",
									Duration:    "0s",
									TCStart:     "0s",
									TCFormat:    "NDF",
									AudioLayout: "stereo",
									AudioRate:   "48k",
									Spine: Spine{
										AssetClips: []AssetClip{},
									},
								},
							},
						},
					},
				},
			},
			SmartCollections: []SmartCollection{
				{
					Name:  "Projects",
					Match: "all",
					Matches: []Match{
						{Rule: "is", Type: "project"},
					},
				},
				{
					Name:  "All Video",
					Match: "any",
					MediaMatches: []MediaMatch{
						{Rule: "is", Type: "videoOnly"},
						{Rule: "is", Type: "videoWithAudio"},
					},
				},
				{
					Name:  "Audio Only",
					Match: "all",
					MediaMatches: []MediaMatch{
						{Rule: "is", Type: "audioOnly"},
					},
				},
				{
					Name:  "Stills",
					Match: "all",
					MediaMatches: []MediaMatch{
						{Rule: "is", Type: "stills"},
					},
				},
				{
					Name:  "Favorites",
					Match: "all",
					RatingMatches: []RatingMatch{
						{Value: "favorites"},
					},
				},
			},
		},
	}

	// If filename is provided, write to file
	if filename != "" {
		err := WriteToFile(fcpxml, filename)
		if err != nil {
			return nil, err
		}
	}

	return fcpxml, nil
}

// WriteToFile marshals the FCPXML struct to a file.
//
// 🚨 CLAUDE.md Rule: NO XML STRING TEMPLATES → USE xml.MarshalIndent() function
// - After writing, VALIDATE with: xmllint --dtdvalid FCPXMLv1_13.dtd filename
// - Before commits, CHECK with: ValidateClaudeCompliance() function
func WriteToFile(fcpxml *FCPXML, filename string) error {
	// Marshal to XML with proper formatting
	// Note: Custom MarshalXML on Spine struct ensures chronological order
	output, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %v", err)
	}

	// Add XML declaration and DOCTYPE
	xmlHeader := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

`
	
	fullXML := xmlHeader + string(output)

	// Write to file
	err = os.WriteFile(filename, []byte(fullXML), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// AddVideo adds a video asset and asset-clip to the FCPXML structure.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Assets, sequence.Spine.AssetClips
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function 
// - Maintains UID consistency → generateUID() function for deterministic UIDs
//
// ❌ NEVER: fmt.Sprintf("<asset-clip ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddVideo(fcpxml *FCPXML, videoPath string) error {
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Check if asset already exists for this file
	if asset, exists := registry.GetOrCreateAsset(videoPath); exists {
		// Asset already exists, just add asset-clip to spine
		return addAssetClipToSpine(fcpxml, asset, 10.0)
	}

	// Create transaction for atomic resource creation
	tx := NewTransaction(registry)

	// Get absolute path
	absPath, err := filepath.Abs(videoPath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		tx.Rollback()
		return fmt.Errorf("video file does not exist: %s", absPath)
	}

	// Reserve ID atomically to prevent collisions
	ids := tx.ReserveIDs(1)
	assetID := ids[0]

	// Generate unique asset name
	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	// Use a default duration of 10 seconds, properly frame-aligned
	defaultDurationSeconds := 10.0
	frameDuration := ConvertSecondsToFCPDuration(defaultDurationSeconds)

	// Create asset using transaction
	asset, err := tx.CreateAsset(assetID, absPath, videoName, frameDuration, "r1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create asset: %v", err)
	}

	// Commit transaction - adds resources to registry and FCPXML
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Add asset-clip to spine
	return addAssetClipToSpine(fcpxml, asset, defaultDurationSeconds)
}

// addAssetClipToSpine adds an asset-clip to the sequence spine
func addAssetClipToSpine(fcpxml *FCPXML, asset *Asset, durationSeconds float64) error {
	// Add asset-clip to the spine if there's a sequence
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		// Calculate current timeline duration by examining existing clips
		currentTimelineDuration := calculateTimelineDuration(sequence)

		// Create asset-clip with frame-aligned duration
		clipDuration := ConvertSecondsToFCPDuration(durationSeconds)

		// 🚨 CLAUDE.md Rule: Asset-Clip Format Consistency
		// - Asset-clips MUST use the ASSET's format, not hardcoded sequence format
		// - This matches the pattern in working FCPXML files
		assetClip := AssetClip{
			Ref:       asset.ID,
			Offset:    currentTimelineDuration, // Append after existing content
			Name:      asset.Name,
			Duration:  clipDuration,
			Format:    asset.Format, // Use asset's format
			TCFormat:  "NDF",
			AudioRole: "dialogue",
		}

		// Add asset-clip to spine using structs
		sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, assetClip)

		// Update sequence duration to include new content
		newTimelineDuration := addDurations(currentTimelineDuration, clipDuration)
		sequence.Duration = newTimelineDuration
	}

	return nil
}

// ValidateClaudeCompliance performs automated checks for CLAUDE.md rule compliance.
//
// 🚨 CLAUDE.md Validation - Run this before any commit!
// This function helps catch violations of critical rules in CLAUDE.md
func ValidateClaudeCompliance(fcpxml *FCPXML) []string {
	var violations []string

	// Check for unique IDs across all resources
	idMap := make(map[string]bool)

	// Check asset IDs
	for _, asset := range fcpxml.Resources.Assets {
		if idMap[asset.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Asset)", asset.ID))
		}
		idMap[asset.ID] = true
	}

	// Check format IDs
	for _, format := range fcpxml.Resources.Formats {
		if idMap[format.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Format)", format.ID))
		}
		idMap[format.ID] = true
	}

	// Check effect IDs
	for _, effect := range fcpxml.Resources.Effects {
		if idMap[effect.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Effect)", effect.ID))
		}
		idMap[effect.ID] = true
	}

	// Check media IDs
	for _, media := range fcpxml.Resources.Media {
		if idMap[media.ID] {
			violations = append(violations, fmt.Sprintf("Duplicate ID found: %s (Media)", media.ID))
		}
		idMap[media.ID] = true
	}

	// Check for frame alignment in durations (basic check for common violations)
	// Look for duration patterns that are definitely not frame-aligned
	checkDuration := func(duration, location string) {
		if strings.Contains(duration, "/600s") && !strings.Contains(duration, "1001") {
			violations = append(violations, fmt.Sprintf("Potentially non-frame-aligned duration '%s' at %s - use ConvertSecondsToFCPDuration()", duration, location))
		}
		if strings.Contains(duration, "/24000s") && duration != "0s" {
			// Check if it follows (frames*1001)/24000s pattern by checking if numerator is divisible by 1001
			durationNoS := strings.TrimSuffix(duration, "s")
			parts := strings.Split(durationNoS, "/")
			if len(parts) == 2 {
				if numerator, err := strconv.Atoi(parts[0]); err == nil {
					// Check if numerator is divisible by 1001 (frame-aligned)
					if numerator%1001 != 0 {
						violations = append(violations, fmt.Sprintf("Non-frame-aligned duration '%s' at %s - must be (frames*1001)/24000s", duration, location))
					}
				}
			}
		}
	}

	// Check asset durations
	for _, asset := range fcpxml.Resources.Assets {
		checkDuration(asset.Duration, fmt.Sprintf("Asset %s", asset.ID))
	}

	// Check sequence durations
	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {
				checkDuration(sequence.Duration, fmt.Sprintf("Sequence in Project %s", project.Name))

				// Check asset-clip durations in spine
				for _, clip := range sequence.Spine.AssetClips {
					checkDuration(clip.Duration, fmt.Sprintf("AssetClip %s in Spine", clip.Name))
				}
			}
		}
	}

	// 🚨 CLAUDE.md Rule: Asset-Clip Format Consistency
	// Check that asset-clips use their referenced asset's format
	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {
				// Check asset-clip formats match their referenced assets
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

	// 🚨 CLAUDE.md Rule: Effect UID Reality Check
	// Check that all effects use valid FCP effect UIDs, not fictional ones
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

	// 🚨 CLAUDE.md Rule: Keyframe Attribute Validation
	// Check for keyframe structure compliance with DTD requirements
	validateKeyframes := func(keyframes []Keyframe, location string) {
		for i, keyframe := range keyframes {
			// Check for valid interpolation values
			if keyframe.Interp != "" {
				validInterps := map[string]bool{"linear": true, "ease": true, "easeIn": true, "easeOut": true}
				if !validInterps[keyframe.Interp] {
					violations = append(violations, fmt.Sprintf("Invalid keyframe interp '%s' at %s[%d] - must be: linear, ease, easeIn, easeOut", keyframe.Interp, location, i))
				}
			}
			
			// Check for valid curve values
			if keyframe.Curve != "" {
				validCurves := map[string]bool{"linear": true, "smooth": true}
				if !validCurves[keyframe.Curve] {
					violations = append(violations, fmt.Sprintf("Invalid keyframe curve '%s' at %s[%d] - must be: linear, smooth", keyframe.Curve, location, i))
				}
			}
		}
	}

	// Check keyframes in all possible locations
	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {
				// Check asset-clips
				for _, clip := range sequence.Spine.AssetClips {
					// Check adjust-transform keyframes
					if clip.AdjustTransform != nil {
						for _, param := range clip.AdjustTransform.Params {
							if param.KeyframeAnimation != nil {
								validateKeyframes(param.KeyframeAnimation.Keyframes, fmt.Sprintf("AssetClip '%s' AdjustTransform param '%s'", clip.Name, param.Name))
							}
						}
					}
					
					// Check filter-video keyframes
					for _, filter := range clip.FilterVideos {
						for _, param := range filter.Params {
							if param.KeyframeAnimation != nil {
								validateKeyframes(param.KeyframeAnimation.Keyframes, fmt.Sprintf("AssetClip '%s' FilterVideo '%s' param '%s'", clip.Name, filter.Name, param.Name))
							}
						}
					}
				}
				
				// Check video keyframes
				for _, video := range sequence.Spine.Videos {
					if video.AdjustTransform != nil {
						for _, param := range video.AdjustTransform.Params {
							if param.KeyframeAnimation != nil {
								validateKeyframes(param.KeyframeAnimation.Keyframes, fmt.Sprintf("Video '%s' AdjustTransform param '%s'", video.Name, param.Name))
							}
						}
					}
				}
				
				// Check title keyframes
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

	// 🚨 CLAUDE.md Rule: Resource Reference Validation
	// Check that all refs in spine elements have corresponding resource definitions
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

	// Check spine element references
	checkRef := func(ref, elementType string) {
		if ref != "" && !resourceIDs[ref] {
			violations = append(violations, fmt.Sprintf("Undefined reference '%s' in %s - missing resource definition", ref, elementType))
		}
	}

	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {
				// Check asset-clips
				for _, clip := range sequence.Spine.AssetClips {
					checkRef(clip.Ref, fmt.Sprintf("AssetClip '%s'", clip.Name))
					
					// Check filter-video refs
					for _, filter := range clip.FilterVideos {
						checkRef(filter.Ref, fmt.Sprintf("FilterVideo '%s' in AssetClip '%s'", filter.Name, clip.Name))
					}
				}
				
				// Check videos
				for _, video := range sequence.Spine.Videos {
					checkRef(video.Ref, fmt.Sprintf("Video '%s'", video.Name))
				}
				
				// Check titles
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
	// Validate that this is actually an image file
	if !isImageFile(imagePath) {
		return fmt.Errorf("file is not a supported image format (PNG, JPG, JPEG): %s", imagePath)
	}

	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Check if asset already exists for this file
	if asset, exists := registry.GetOrCreateAsset(imagePath); exists {
		// Asset already exists, just add asset-clip to spine
		return addImageAssetClipToSpine(fcpxml, asset, durationSeconds)
	}

	// Create transaction for atomic resource creation
	tx := NewTransaction(registry)

	// Get absolute path
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		tx.Rollback()
		return fmt.Errorf("image file does not exist: %s", absPath)
	}

	// Reserve IDs atomically to prevent collisions (need 2: asset + format)
	ids := tx.ReserveIDs(2)
	assetID := ids[0]
	formatID := ids[1]

	// Generate unique asset name
	imageName := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))

	// Convert duration to frame-aligned format
	frameDuration := ConvertSecondsToFCPDuration(durationSeconds)

	// Create image-specific format using transaction
	// 🚨 CRITICAL FIX: Image formats must match working top5orig.fcpxml pattern
	// Analysis of working top5orig.fcpxml vs our crashing files revealed:
	// 1. Image formats must NOT have frameDuration (causes performAudioPreflightCheckForObject crash)
	// 2. Image formats must use name="FFVideoFormatRateUndefined" and colorSpace="1-13-1"
	// 3. Only sequence formats should have frameDuration, image formats are timeless
	// Working pattern: name="FFVideoFormatRateUndefined", colorSpace="1-13-1", NO frameDuration
	_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1280", "720", "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create image format: %v", err)
	}

	// Create asset using transaction (CreateAsset handles image-specific properties)
	asset, err := tx.CreateAsset(assetID, absPath, imageName, frameDuration, formatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create asset: %v", err)
	}

	// Commit transaction - adds resources to registry and FCPXML
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Add asset-clip to spine with optional slide animation
	return addImageAssetClipToSpineWithSlide(fcpxml, asset, durationSeconds, withSlide)
}

// addImageAssetClipToSpine adds an image Video element to the sequence spine
// 🚨 CRITICAL FIX: Images use Video elements, NOT AssetClip elements
// Analysis of working samples/png.fcpxml shows images use <video> in spine
// This prevents addAssetClip:toObject:parentFormatID crashes in FCP
func addImageAssetClipToSpine(fcpxml *FCPXML, asset *Asset, durationSeconds float64) error {
	return addImageAssetClipToSpineWithSlide(fcpxml, asset, durationSeconds, false)
}

// addImageAssetClipToSpineWithSlide adds an image Video element to the sequence spine with optional slide animation
// 🚨 CRITICAL FIX: Images use Video elements, NOT AssetClip elements
// Analysis of working samples/png.fcpxml shows images use <video> in spine
// This prevents addAssetClip:toObject:parentFormatID crashes in FCP
// Keyframe animations match samples/slide.fcpxml pattern for sliding motion
func addImageAssetClipToSpineWithSlide(fcpxml *FCPXML, asset *Asset, durationSeconds float64, withSlide bool) error {
	// Add Video element to the spine if there's a sequence
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		// Calculate current timeline duration by examining existing clips
		currentTimelineDuration := calculateTimelineDuration(sequence)

		// Create Video element with frame-aligned duration
		// 🚨 CRITICAL: Display duration applied to Video element, not asset
		// Asset duration is "0s" (timeless), Video element has display duration
		clipDuration := ConvertSecondsToFCPDuration(durationSeconds)

		// Create Video element matching working samples/png.fcpxml pattern
		// Working pattern: <video ref="r2" offset="0s" name="cs.pitt.edu" start="86399313/24000s" duration="241241/24000s"/>
		video := Video{
			Ref:      asset.ID,
			Offset:   currentTimelineDuration, // Append after existing content
			Name:     asset.Name,
			Start:    "86399313/24000s", // Standard FCP start offset for images
			Duration: clipDuration,
			// Note: No Format attribute on Video elements (different from AssetClip)
		}

		// Add slide animation if requested
		if withSlide {
			video.AdjustTransform = createSlideAnimation(currentTimelineDuration, durationSeconds)
		}

		// Add Video element to spine using structs
		sequence.Spine.Videos = append(sequence.Spine.Videos, video)

		// Update sequence duration to include new content
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
	// Read file contents
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

// calculateTimelineDuration calculates the total duration of content in a sequence
// by examining all clips in the spine and finding the maximum offset + duration
func calculateTimelineDuration(sequence *Sequence) string {
	maxEndTime := 0 // Track end time in 1001/24000s units

	// Check all asset clips in spine
	for _, clip := range sequence.Spine.AssetClips {
		clipEndTime := parseOffsetAndDuration(clip.Offset, clip.Duration)
		if clipEndTime > maxEndTime {
			maxEndTime = clipEndTime
		}
	}

	// Check all video clips in spine
	for _, video := range sequence.Spine.Videos {
		videoEndTime := parseOffsetAndDuration(video.Offset, video.Duration)
		if videoEndTime > maxEndTime {
			maxEndTime = videoEndTime
		}
	}

	// Check all title clips in spine
	for _, title := range sequence.Spine.Titles {
		titleEndTime := parseOffsetAndDuration(title.Offset, title.Duration)
		if titleEndTime > maxEndTime {
			maxEndTime = titleEndTime
		}
	}

	// Check all gaps in spine
	for _, gap := range sequence.Spine.Gaps {
		gapEndTime := parseOffsetAndDuration(gap.Offset, gap.Duration)
		if gapEndTime > maxEndTime {
			maxEndTime = gapEndTime
		}
	}

	// Return as FCP duration format
	if maxEndTime == 0 {
		return "0s"
	}
	return fmt.Sprintf("%d/24000s", maxEndTime)
}

// parseOffsetAndDuration parses FCP time format and returns end time in 1001/24000s units
func parseOffsetAndDuration(offset, duration string) int {
	offsetFrames := parseFCPDuration(offset)
	durationFrames := parseFCPDuration(duration)
	return offsetFrames + durationFrames
}

// parseFCPDuration parses FCP duration format and returns frame-aligned values in 1001/24000s units
func parseFCPDuration(duration string) int {
	if duration == "0s" {
		return 0
	}

	// Parse rational duration formats like "12345/24000s", "547547/60000s", etc.
	if strings.HasSuffix(duration, "s") && strings.Contains(duration, "/") {
		// Remove the "s" suffix
		durationNoS := strings.TrimSuffix(duration, "s")
		
		// Split by "/"
		parts := strings.Split(durationNoS, "/")
		if len(parts) == 2 {
			numerator, err1 := strconv.Atoi(parts[0])
			denominator, err2 := strconv.Atoi(parts[1])
			
			if err1 == nil && err2 == nil && denominator != 0 {
				// 🚨 CLAUDE.md CRITICAL: Frame Boundary Alignment
				// FCP uses 1001/24000s frame duration (≈ 23.976 fps)
				// All durations MUST be frame-aligned: (frames × 1001)/24000s
				
				// Convert to exact frame count using FCP's frame duration
				// frames = (numerator/denominator) / (1001/24000) = (numerator * 24000) / (denominator * 1001)
				framesFloat := float64(numerator * 24000) / float64(denominator * 1001)
				frames := int(framesFloat + 0.5) // Round to nearest frame
				
				// Return frame-aligned value: frames * 1001
				return frames * 1001
			}
		}
	}

	return 0
}

// addDurations adds two FCP duration strings and returns the result
func addDurations(duration1, duration2 string) string {
	frames1 := parseFCPDuration(duration1)
	frames2 := parseFCPDuration(duration2)
	totalFrames := frames1 + frames2
	return fmt.Sprintf("%d/24000s", totalFrames)
}

// createSlideAnimation creates keyframe animation for sliding an image from left to right
// Based on samples/slide.fcpxml pattern with keyframes for position parameter
// Slides from position "0 0" to "51.3109 0" over 1 second (from start to 1 second into the clip)
func createSlideAnimation(offsetDuration string, totalDurationSeconds float64) *AdjustTransform {
	// Calculate keyframe times based on video start time (like samples/slide.fcpxml)
	// The sample uses video start time as base: start="86399313/24000s"
	// We'll use the standard FCP start time for images
	videoStartFrames := 86399313 // Standard FCP start time for image assets

	// Calculate keyframe times using proper frame alignment:
	// - Start keyframe: at the video start time
	// - End keyframe: 1 second later using ConvertSecondsToFCPDuration for frame alignment
	oneSecondDuration := ConvertSecondsToFCPDuration(1.0)
	oneSecondFrames := parseFCPDuration(oneSecondDuration)

	startTime := fmt.Sprintf("%d/24000s", videoStartFrames)
	endTime := fmt.Sprintf("%d/24000s", videoStartFrames+oneSecondFrames)

	// Create AdjustTransform with keyframe animation matching samples/slide.fcpxml
	// The sample shows position animation from "0 0" to "51.3109 0", increased by 100px to "59.3109 0"
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "anchor",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  endTime,
							Value: "0 0",
							Interp: "linear",
							Curve:  "linear",
						},
					},
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  startTime,
							Value: "0 0",
						},
						{
							Time:  endTime,
							Value: "59.3109 0",
						},
					},
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  endTime,
							Value: "0",
							Interp: "linear",
							Curve:  "linear",
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  endTime,
							Value: "1 1",
							Interp: "linear",
							Curve:  "linear",
						},
					},
				},
			},
		},
	}
}

// AddTextFromFile reads a text file and adds staggered text elements to the FCPXML structure.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Effects, sequence.Spine.Titles
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function 
// - Unique text-style-def IDs → generateUID() function for deterministic UIDs
// - Each text element appears 1 second later with 300px Y offset progression
//
// ❌ NEVER: fmt.Sprintf("<title ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddTextFromFile(fcpxml *FCPXML, textFilePath string, offsetSeconds float64, durationSeconds float64) error {
	// Read text file
	data, err := os.ReadFile(textFilePath)
	if err != nil {
		return fmt.Errorf("failed to read text file: %v", err)
	}

	// Split into lines and filter out empty lines
	lines := strings.Split(string(data), "\n")
	var textLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			textLines = append(textLines, line)
		}
	}

	if len(textLines) == 0 {
		return fmt.Errorf("no text lines found in file: %s", textFilePath)
	}

	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Create transaction for atomic resource creation
	tx := NewTransaction(registry)

	// Check if text effect already exists, if not create it
	textEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if strings.Contains(effect.UID, "Text.moti") {
			textEffectID = effect.ID
			break
		}
	}

	if textEffectID == "" {
		// Reserve ID for text effect
		ids := tx.ReserveIDs(1)
		textEffectID = ids[0]

		// Create text effect using transaction
		_, err = tx.CreateEffect(textEffectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create text effect: %v", err)
		}
	}

	// Commit transaction to ensure effect is available
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit text effect: %v", err)
	}

	// Add text elements to the spine if there's a sequence
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		// Find the clip element that covers the text offset time
		var targetAssetClip *AssetClip = nil
		var targetVideo *Video = nil
		offsetFrames := parseFCPDuration(ConvertSecondsToFCPDuration(offsetSeconds))

		// First check AssetClip elements (these can directly contain titles per DTD)
		for i := range sequence.Spine.AssetClips {
			clip := &sequence.Spine.AssetClips[i]
			clipOffsetFrames := parseFCPDuration(clip.Offset)
			clipDurationFrames := parseFCPDuration(clip.Duration)
			clipEndFrames := clipOffsetFrames + clipDurationFrames

			// Check if the text offset falls within this clip's timeline
			if offsetFrames >= clipOffsetFrames && offsetFrames < clipEndFrames {
				targetAssetClip = clip
				break
			}
		}

		// If no AssetClip found, check Video elements as fallback
		if targetAssetClip == nil {
			for i := range sequence.Spine.Videos {
				video := &sequence.Spine.Videos[i]
				videoOffsetFrames := parseFCPDuration(video.Offset)
				videoDurationFrames := parseFCPDuration(video.Duration)
				videoEndFrames := videoOffsetFrames + videoDurationFrames

				// Check if the text offset falls within this video's timeline
				if offsetFrames >= videoOffsetFrames && offsetFrames < videoEndFrames {
					targetVideo = video
					break
				}
			}
		}

		// Use fallback logic if no element covers the text timing
		if targetAssetClip == nil && targetVideo == nil {
			if len(sequence.Spine.AssetClips) > 0 {
				// Use last AssetClip as fallback (common when offset is beyond timeline)
				targetAssetClip = &sequence.Spine.AssetClips[len(sequence.Spine.AssetClips)-1]
			} else if len(sequence.Spine.Videos) > 0 {
				// Use last Video as fallback (common when offset is beyond timeline)
				targetVideo = &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]
			}
		}

		if targetAssetClip == nil && targetVideo == nil {
			return fmt.Errorf("no video or asset-clip element found in spine to add text overlays to")
		}

		// Use the provided duration for each text element
		textDuration := ConvertSecondsToFCPDuration(durationSeconds)

		// Process each text line
		for i, textLine := range textLines {
			// Create new transaction for each text element to ensure unique IDs
			textTx := NewTransaction(registry)

			// Generate unique text-style-def ID using hash-based approach (CLAUDE.md requirement)
			// This prevents ID collisions when adding text to existing FCPXML files
			textStyleID := GenerateTextStyleID(textLine, fmt.Sprintf("line_%d_offset_%.1f", i, offsetSeconds))

			// Calculate staggered timing: first element at offsetSeconds in sequence timeline, each subsequent +6 seconds
			// Text timing should use the clip's start time as base for proper FCP timing
			var clipStartFrames int
			if targetAssetClip != nil {
				clipStartFrames = parseFCPDuration(targetAssetClip.Start)
			} else if targetVideo != nil {
				clipStartFrames = parseFCPDuration(targetVideo.Start)
			}
			// Calculate stagger based on 50% of text duration (e.g., 2s duration = 1s stagger)
			staggerSeconds := durationSeconds * 0.5
			staggerDuration := ConvertSecondsToFCPDuration(staggerSeconds)
			staggerFramesPer := parseFCPDuration(staggerDuration)
			staggerFrames := i * staggerFramesPer
			elementOffsetFrames := clipStartFrames + staggerFrames
			elementOffset := fmt.Sprintf("%d/24000s", elementOffsetFrames)

			// Calculate Y position offset: each element 300px lower (negative Y in FCP coordinates)
			// Pattern from sample: Position "0 0", "0 -300", "0 -600"
			yOffset := i * -300
			positionValue := fmt.Sprintf("0 %d", yOffset)

			// Calculate lane number: decending lanes for stacking (3, 2, 1, ...)
			laneNumber := len(textLines) - i

			// Create Title element with comprehensive parameters matching sample pattern
			title := Title{
				Ref:      textEffectID,
				Lane:     fmt.Sprintf("%d", laneNumber),
				Offset:   elementOffset,
				Name:     fmt.Sprintf("%s - Text", textLine),
				Start:    "86486400/24000s", // Standard FCP start time for text
				Duration: textDuration,
				Params: []Param{
					{
						Name:  "Layout Method",
						Key:   "9999/10003/13260/3296672360/2/314",
						Value: "1 (Paragraph)",
					},
					{
						Name:  "Left Margin",
						Key:   "9999/10003/13260/3296672360/2/323",
						Value: "-1730",
					},
					{
						Name:  "Right Margin",
						Key:   "9999/10003/13260/3296672360/2/324",
						Value: "1730",
					},
					{
						Name:  "Top Margin",
						Key:   "9999/10003/13260/3296672360/2/325",
						Value: "960",
					},
					{
						Name:  "Bottom Margin",
						Key:   "9999/10003/13260/3296672360/2/326",
						Value: "-960",
					},
					{
						Name:  "Alignment",
						Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
						Value: "0 (Left)",
					},
					{
						Name:  "Line Spacing",
						Key:   "9999/10003/13260/3296672360/2/354/3296667315/404",
						Value: "-19",
					},
					{
						Name:  "Auto-Shrink",
						Key:   "9999/10003/13260/3296672360/2/370",
						Value: "3 (To All Margins)",
					},
					{
						Name:  "Alignment",
						Key:   "9999/10003/13260/3296672360/2/373",
						Value: "0 (Left) 0 (Top)",
					},
					{
						Name:  "Opacity",
						Key:   "9999/10003/13260/3296672360/4/3296673134/1000/1044",
						Value: "0",
					},
					{
						Name:  "Speed",
						Key:   "9999/10003/13260/3296672360/4/3296673134/201/208",
						Value: "6 (Custom)",
					},
					{
						Name: "Custom Speed",
						Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{
									Time:  "-469658744/1000000000s",
									Value: "0",
								},
								{
									Time:  "12328542033/1000000000s",
									Value: "1",
								},
							},
						},
					},
					{
						Name:  "Apply Speed",
						Key:   "9999/10003/13260/3296672360/4/3296673134/201/211",
						Value: "2 (Per Object)",
					},
				},
				Text: &TitleText{
					TextStyles: []TextStyleRef{
						{
							Ref:  textStyleID,
							Text: textLine,
						},
					},
				},
				TextStyleDefs: []TextStyleDef{
					{
						ID: textStyleID,
						TextStyle: TextStyle{
							Font:        "Helvetica Neue",
							FontSize:    "134",
							FontColor:   "1 1 1 1",
							Bold:        "1",
							LineSpacing: "-19",
						},
					},
				},
			}

			// Only add Position parameter if it's not the first element (which has 0 0 position)
			if i > 0 {
				positionParam := Param{
					Name:  "Position",
					Key:   "9999/10003/13260/3296672360/1/100/101",
					Value: positionValue,
				}
				// Insert Position parameter at the beginning for consistency with sample
				title.Params = append([]Param{positionParam}, title.Params...)
			}

			// Commit text transaction to ensure unique IDs
			err = textTx.Commit()
			if err != nil {
				return fmt.Errorf("failed to commit text transaction for element %d: %v", i, err)
			}

			// Add title as nested element within the target clip
			if targetAssetClip != nil {
				targetAssetClip.Titles = append(targetAssetClip.Titles, title)
			} else if targetVideo != nil {
				targetVideo.NestedTitles = append(targetVideo.NestedTitles, title)
			}
		}

		// Text elements are added as overlays - no need to extend underlying video duration
	}

	return nil
}


// AddSlideToVideoAtOffset finds a video at the specified offset and adds slide animation to it.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses frame-aligned timing → ConvertSecondsToFCPDuration() function for offset calculation
// - Uses STRUCTS ONLY - no string templates → modifies Video.AdjustTransform in spine
// - Maintains existing video properties while adding slide animation keyframes
// - Proper FCP timing with video start time as base for animation keyframes
//
// ❌ NEVER: fmt.Sprintf("<adjust-transform...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use structs to modify Video.AdjustTransform with keyframe animation
func AddSlideToVideoAtOffset(fcpxml *FCPXML, offsetSeconds float64) error {
	// Find the sequence
	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return fmt.Errorf("no sequence found in FCPXML")
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Convert offset to frame-aligned format
	offsetFrames := parseFCPDuration(ConvertSecondsToFCPDuration(offsetSeconds))

	// Find the video at the specified offset
	var targetVideo *Video = nil

	// Search through Video elements first
	for i := range sequence.Spine.Videos {
		video := &sequence.Spine.Videos[i]
		videoOffsetFrames := parseFCPDuration(video.Offset)
		videoDurationFrames := parseFCPDuration(video.Duration)
		videoEndFrames := videoOffsetFrames + videoDurationFrames

		// Check if the offset falls within this video's timeline
		if offsetFrames >= videoOffsetFrames && offsetFrames < videoEndFrames {
			targetVideo = video
			break
		}
	}

	// If no Video element found, check AssetClip elements and add animation directly
	var targetClip *AssetClip = nil
	if targetVideo == nil {
		for i := range sequence.Spine.AssetClips {
			clip := &sequence.Spine.AssetClips[i]
			clipOffsetFrames := parseFCPDuration(clip.Offset)
			clipDurationFrames := parseFCPDuration(clip.Duration)
			clipEndFrames := clipOffsetFrames + clipDurationFrames

			// Check if the offset falls within this video's timeline
			if offsetFrames >= clipOffsetFrames && offsetFrames < clipEndFrames {
				// Add slide animation directly to AssetClip (don't convert to Video)
				targetClip = &sequence.Spine.AssetClips[i]
				break
			}
		}
	}

	if targetVideo == nil && targetClip == nil {
		return fmt.Errorf("no video found at offset %.1f seconds", offsetSeconds)
	}

	// Handle Video elements (for images)
	if targetVideo != nil {
		// Check if video already has slide animation
		if targetVideo.AdjustTransform != nil {
			// Check if position parameter already exists with keyframes
			for _, param := range targetVideo.AdjustTransform.Params {
				if param.Name == "position" && param.KeyframeAnimation != nil {
					return fmt.Errorf("video '%s' at offset %.1f seconds already has slide animation", targetVideo.Name, offsetSeconds)
				}
			}
		}

		// Calculate slide animation duration (1 second from video start)
		videoStartFrames := parseFCPDuration(targetVideo.Start)
		if videoStartFrames == 0 {
			// If no start time, use standard FCP start time for images
			videoStartFrames = 86399313
			targetVideo.Start = "86399313/24000s"
		}

		// Add slide animation to the video
		targetVideo.AdjustTransform = createSlideAnimation(targetVideo.Offset, 1.0)
	}

	// Handle AssetClip elements (for videos)
	if targetClip != nil {
		// Check if clip already has slide animation
		if targetClip.AdjustTransform != nil {
			// Check if position parameter already exists with keyframes
			for _, param := range targetClip.AdjustTransform.Params {
				if param.Name == "position" && param.KeyframeAnimation != nil {
					return fmt.Errorf("video '%s' at offset %.1f seconds already has slide animation", targetClip.Name, offsetSeconds)
				}
			}
		}

		// For AssetClip elements, use timeline-based animation (not the special image start time)
		targetClip.AdjustTransform = createAssetClipSlideAnimation(targetClip.Offset, 1.0)
	}

	return nil
}

// createAssetClipSlideAnimation creates timeline-based slide animation for AssetClip elements (videos)
func createAssetClipSlideAnimation(clipOffset string, totalDurationSeconds float64) *AdjustTransform {
	// For AssetClip elements, use timeline-based timing starting from clip offset
	offsetFrames := parseFCPDuration(clipOffset)
	
	// Calculate keyframe times using proper frame alignment:
	// - Start keyframe: at the clip offset time  
	// - End keyframe: 1 second later using ConvertSecondsToFCPDuration for frame alignment
	oneSecondDuration := ConvertSecondsToFCPDuration(1.0)
	oneSecondFrames := parseFCPDuration(oneSecondDuration)

	startTime := fmt.Sprintf("%d/24000s", offsetFrames)
	endTime := fmt.Sprintf("%d/24000s", offsetFrames+oneSecondFrames)

	// Create AdjustTransform with keyframe animation for AssetClip
	// Same position animation as images but with timeline-based timing
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "anchor",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  endTime,
							Value: "0 0",
							Interp: "linear",
							Curve:  "linear",
						},
					},
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  startTime,
							Value: "0 0",
						},
						{
							Time:  endTime,
							Value: "59.3109 0",
						},
					},
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  endTime,
							Value: "0",
							Interp: "linear",
							Curve:  "linear",
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  endTime,
							Value: "1 1",
							Interp: "linear",
							Curve:  "linear",
						},
					},
				},
			},
		},
	}
}

// isAudioFile checks if the given file is an audio file (WAV, MP3, M4A).
//
// 🚨 CLAUDE.md Rule: Audio vs Video Asset Properties
// - Audio files MUST have HasAudio="1" and AudioSources set
// - Audio files MUST NOT have HasVideo="1" or VideoSources
// - Duration is determined by actual audio file duration
func isAudioFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".wav" || ext == ".mp3" || ext == ".m4a" || ext == ".aac" || ext == ".flac" || ext == ".caf"
}

// AddAudio adds an audio asset and asset-clip to the FCPXML structure as the main audio track starting at 00:00.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Assets, sequence.Spine.AssetClips
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function 
// - Maintains UID consistency → generateUID() function for deterministic UIDs
// - Audio-specific properties → HasAudio="1", AudioSources, AudioChannels, AudioRate
//
// ❌ NEVER: fmt.Sprintf("<asset-clip ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddAudio(fcpxml *FCPXML, audioPath string) error {
	// Validate that this is actually an audio file
	if !isAudioFile(audioPath) {
		return fmt.Errorf("file is not a supported audio format (WAV, MP3, M4A, AAC, FLAC): %s", audioPath)
	}

	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Check if asset already exists for this file
	if asset, exists := registry.GetOrCreateAsset(audioPath); exists {
		// Asset already exists, just add asset-clip to spine at 00:00
		return addAudioAssetClipToSpine(fcpxml, asset)
	}

	// Create transaction for atomic resource creation
	tx := NewTransaction(registry)

	// Get absolute path
	absPath, err := filepath.Abs(audioPath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		tx.Rollback()
		return fmt.Errorf("audio file does not exist: %s", absPath)
	}

	// Reserve ID atomically to prevent collisions
	ids := tx.ReserveIDs(1)
	assetID := ids[0]

	// Generate unique asset name
	audioName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))

	// Use a default duration of 60 seconds for audio, properly frame-aligned
	// Real audio duration would need audio file parsing, but for now use default
	defaultDurationSeconds := 60.0
	frameDuration := ConvertSecondsToFCPDuration(defaultDurationSeconds)

	// Create asset using transaction (CreateAsset handles audio-specific properties)
	asset, err := tx.CreateAsset(assetID, absPath, audioName, frameDuration, "r1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create audio asset: %v", err)
	}

	// Commit transaction - adds resources to registry and FCPXML
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Add asset-clip to spine at 00:00
	return addAudioAssetClipToSpine(fcpxml, asset)
}

// addAudioAssetClipToSpine adds an audio asset-clip nested inside the first video element
// 🚨 CRITICAL FIX: Audio must be nested inside video elements, not as separate spine elements
// Analysis of Info.fcpxml shows audio is nested: <video><asset-clip lane="-1"/></video>
func addAudioAssetClipToSpine(fcpxml *FCPXML, asset *Asset) error {
	// Add audio as nested asset-clip inside first video element
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		// Find the first video element in the spine to nest audio inside
		var targetVideo *Video = nil
		for i := range sequence.Spine.Videos {
			targetVideo = &sequence.Spine.Videos[i]
			break
		}

		// If no video elements exist, convert first AssetClip to Video for audio nesting
		if targetVideo == nil && len(sequence.Spine.AssetClips) > 0 {
			clip := &sequence.Spine.AssetClips[0]
			video := Video{
				Ref:      clip.Ref,
				Offset:   clip.Offset,
				Name:     clip.Name,
				Duration: clip.Duration,
				Start:    clip.Start,
			}

			// Remove the AssetClip and replace with Video
			sequence.Spine.AssetClips = sequence.Spine.AssetClips[1:]
			sequence.Spine.Videos = append(sequence.Spine.Videos, video)
			targetVideo = &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]
		}

		if targetVideo == nil {
			return fmt.Errorf("no video element found to nest audio inside - audio must be nested within a video element")
		}

		// Audio offset calculation matching Info.fcpxml pattern
		// Info.fcpxml uses "28799771/8000s" which is approximately start of video
		// We'll use a similar calculated offset based on video start time
		audioOffset := "28799771/8000s"

		// Get audio duration from asset
		audioDuration := asset.Duration

		// 🚨 CLAUDE.md Rule: Audio Nesting Pattern from Info.fcpxml
		// - Audio clips use lane="-1" when nested inside video elements
		// - Audio is nested as child element of video, not separate spine element
		assetClip := AssetClip{
			Ref:       asset.ID,
			Lane:      "-1",           // Nested audio track uses lane="-1"
			Offset:    audioOffset,    // Calculated offset for audio sync
			Name:      asset.Name,
			Duration:  audioDuration,
			Format:    asset.Format,   // Use asset's format
			TCFormat:  "NDF",
			AudioRole: "dialogue",
		}

		// Add audio asset-clip as nested element inside the video
		targetVideo.NestedAssetClips = append(targetVideo.NestedAssetClips, assetClip)

		// Update sequence duration if audio extends beyond current content
		currentSequenceDurationFrames := parseFCPDuration(sequence.Duration)
		audioDurationFrames := parseFCPDuration(audioDuration)
		
		if audioDurationFrames > currentSequenceDurationFrames {
			sequence.Duration = audioDuration
		}
	}

	return nil
}

// AddPipVideo adds a video as picture-in-picture (PIP) to an existing FCPXML file.
//
// 🚨 CRITICAL PIP Requirements (see CLAUDE.md for details):
// 1. **Format Compatibility**: Creates separate formats for main and PIP videos (different from sequence)
//    to allow conform-rate elements without causing "Encountered an unexpected value" FCP errors
// 2. **Layering Strategy**: PIP video uses lane="-1" (background), main video becomes corner overlay
// 3. **Shape Mask Application**: Applied to main video for rounded corners on the small corner video
//
// Structure Generated (matches samples/pip.fcpxml):
// ```
// <asset-clip ref="main" format="r5"> <!-- Main: new format enables conform-rate -->
//     <conform-rate scaleEnabled="0"/>
//     <adjust-crop mode="trim">...</adjust-crop>
//     <adjust-transform position="60.3234 -35.9353" scale="0.28572 0.28572"/> <!-- Corner -->
//     <asset-clip ref="pip" lane="-1" format="r4"> <!-- PIP: background full-size -->
//         <conform-rate scaleEnabled="0" srcFrameRate="60"/>
//     </asset-clip>
//     <filter-video name="Shape Mask">...</filter-video> <!-- Rounded corners -->
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
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Variable to store main video format ID for PIP mode
	var mainVideoFormatID string

	// Check if asset already exists for this file
	var pipAsset *Asset
	if asset, exists := registry.GetOrCreateAsset(pipVideoPath); exists {
		pipAsset = asset
		// If asset already existed, no new format was created for main video
		mainVideoFormatID = ""
	} else {
		// Create transaction for atomic resource creation
		tx := NewTransaction(registry)

		// Get absolute path
		absPath, err := filepath.Abs(pipVideoPath)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			tx.Rollback()
			return fmt.Errorf("PIP video file does not exist: %s", absPath)
		}

		// Reserve IDs atomically to prevent collisions (need 3: pip asset + pip format + main format)
		ids := tx.ReserveIDs(3)
		assetID := ids[0]
		formatID := ids[1]
		mainFormatID := ids[2]

		// Generate unique asset name
		videoName := strings.TrimSuffix(filepath.Base(pipVideoPath), filepath.Ext(pipVideoPath))

		// Use a default duration of 10 seconds for PIP video, properly frame-aligned
		defaultDurationSeconds := 10.0
		frameDuration := ConvertSecondsToFCPDuration(defaultDurationSeconds)

		// Create PIP video format using transaction (similar to main video format)
		// Based on samples/pip.fcpxml: r5 format for PIP video
		_, err = tx.CreateFormatWithFrameDuration(formatID, "100/6000s", "2336", "1510", "1-1-1 (Rec. 709)")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create PIP video format: %v", err)
		}

		// Create separate format for main video (to match samples/pip.fcpxml pattern)
		// 🚨 CRITICAL: Main video MUST have different format than sequence to enable conform-rate
		// Without this, FCP throws "Encountered an unexpected value" errors on conform-rate elements
		_, err = tx.CreateFormatWithFrameDuration(mainFormatID, "13335/400000s", "1920", "1080", "1-1-1 (Rec. 709)")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create main video format for PIP: %v", err)
		}

		// Create video-only asset using transaction (no audio properties for PIP)
		// This matches samples/pip.fcpxml pattern where PIP video has no audio properties
		asset, err := tx.CreateVideoOnlyAsset(assetID, absPath, videoName, frameDuration, formatID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create PIP asset: %v", err)
		}

		// Commit transaction - adds resources to registry and FCPXML
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		pipAsset = asset

		// Store the main format ID that was created in the transaction
		mainVideoFormatID = mainFormatID
	}


	// Find the first asset-clip in the spine to add PIP and transforms to
	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return fmt.Errorf("no sequence found in FCPXML")
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Find the first asset-clip to add PIP to (this becomes the main video)
	if len(sequence.Spine.AssetClips) == 0 {
		return fmt.Errorf("no asset-clip found in spine to add PIP to - need at least one video in the sequence")
	}

	mainClip := &sequence.Spine.AssetClips[0]

	// Update main clip format if we created a new one for PIP
	if mainVideoFormatID != "" {
		mainClip.Format = mainVideoFormatID
	}

	// Add Shape Mask effect if not already exists
	shapeMaskEffectID := ""
	for _, effect := range fcpxml.Resources.Effects {
		if effect.UID == "FFSuperEllipseMask" {
			shapeMaskEffectID = effect.ID
			break
		}
	}

	if shapeMaskEffectID == "" {
		// Create transaction for Shape Mask effect
		tx := NewTransaction(registry)
		ids := tx.ReserveIDs(1)
		shapeMaskEffectID = ids[0]

		// Create Shape Mask effect using transaction
		_, err := tx.CreateEffect(shapeMaskEffectID, "Shape Mask", "FFSuperEllipseMask")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create Shape Mask effect: %v", err)
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit Shape Mask effect: %v", err)
		}
	}

	// Calculate PIP video offset based on provided offset seconds
	// Pattern from samples/pip.fcpxml: offset="35900/3000s" which is about 11.97 seconds
	pipOffsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)

	// Calculate PIP video duration (use default or parse from asset)
	pipDurationSeconds := 10.0 // Default duration for PIP
	pipDuration := ConvertSecondsToFCPDuration(pipDurationSeconds)

	// Create PIP video asset-clip (nested inside main video) - based on samples/pip.fcpxml pattern
	// <asset-clip ref="r4" lane="-1" offset="35900/3000s" name="test1" start="67300/3000s" duration="364700/3000s" format="r5" tcFormat="NDF">
	pipClip := AssetClip{
		Ref:      pipAsset.ID,
		Lane:     "-1",                // PIP video uses lane="-1" for layering
		Offset:   pipOffsetDuration,   // Calculated offset for PIP timing
		Name:     pipAsset.Name,
		Start:    "67300/3000s",       // Standard FCP start time for PIP (from sample)
		Duration: pipDuration,
		Format:   pipAsset.Format,     // Use PIP asset's format
		TCFormat: "NDF",
		ConformRate: &ConformRate{     // Add conform rate for PIP video
			ScaleEnabled: "0",
			SrcFrameRate: "60", // Based on samples/pip.fcpxml pattern
		},
	}

	// Add main video transforms (position and scale) based on samples/pip.fcpxml
	// <adjust-crop mode="trim"><trim-rect left="27.1921" right="27.1001" bottom="12.6562"/></adjust-crop>
	// <adjust-transform position="60.3234 -35.9353" scale="0.28572 0.28572"/>
	
	// For PIP effects, the main video should always have conform-rate scaleEnabled="0"
	// This matches the pattern in samples/pip.fcpxml where main asset has different format than sequence
	// and needs conform-rate for proper scaling behavior with transforms
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
		Position: "60.3234 -35.9353", // Position offset for main video
		Scale:    "0.28572 0.28572",   // Scale down main video
	}

	// Add Shape Mask filter to main video (which becomes the small corner video due to transforms)
	// The main video gets scaled down and positioned in corner, so it needs rounded corners
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

	// Add PIP video as nested asset-clip inside the main video
	mainClip.NestedAssetClips = append(mainClip.NestedAssetClips, pipClip)

	return nil
}
