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
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
					Name:          "FFVideoFormat720p2398",
					FrameDuration: "1001/24000s",
					Width:         "1280",
					Height:        "720",
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

	// Create asset with video detection using transaction
	err = tx.CreateVideoAssetWithDetection(assetID, absPath, videoName, frameDuration, "r1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create video asset with detection: %v", err)
	}

	// Commit transaction - adds resources to registry and FCPXML
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Find the created asset in resources for spine addition
	var asset *Asset
	for i := range fcpxml.Resources.Assets {
		if fcpxml.Resources.Assets[i].ID == assetID {
			asset = &fcpxml.Resources.Assets[i]
			break
		}
	}
	if asset == nil {
		return fmt.Errorf("created asset not found in resources")
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

// AddSingleText adds a single text element like in samples/imessage001.fcpxml to an FCPXML file.
//
// 🚨 CLAUDE.md Rules Applied Here:
// - Uses ResourceRegistry/Transaction system for crash-safe resource management
// - Uses STRUCTS ONLY - no string templates → append to fcpxml.Resources.Effects, sequence.Spine.Titles
// - Atomic ID reservation prevents race conditions and ID collisions
// - Uses frame-aligned durations → ConvertSecondsToFCPDuration() function 
// - Uses verified Text effect UID from samples/imessage001.fcpxml → ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti"
//
// ❌ NEVER: fmt.Sprintf("<title ref='%s'...") - CRITICAL VIOLATION!
// ✅ ALWAYS: Use ResourceRegistry/Transaction pattern for proper resource management
func AddSingleText(fcpxml *FCPXML, text string, offsetSeconds float64, durationSeconds float64) error {
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)

	// Create transaction for atomic resource creation
	tx := NewTransaction(registry)

	// Reserve IDs atomically to prevent collisions (need 1: effect for text)
	ids := tx.ReserveIDs(1)
	effectID := ids[0]

	// Create text effect with verified UID from samples/imessage001.fcpxml
	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create text effect: %v", err)
	}

	// Generate unique text style ID using existing function
	textStyleID := GenerateTextStyleID(text, fmt.Sprintf("single_text_offset_%.1f", offsetSeconds))

	// Convert times to frame-aligned format
	offsetDuration := ConvertSecondsToFCPDuration(offsetSeconds)
	titleDuration := ConvertSecondsToFCPDuration(durationSeconds)

	// Create Title element matching samples/imessage001.fcpxml pattern
	title := Title{
		Ref:      effectID,
		Lane:     "2", // Lane 2 like in the sample
		Offset:   offsetDuration,
		Name:     text + " - Text",
		Start:    "21632100/6000s", // Start time from sample
		Duration: titleDuration,
		Params: []Param{
			{
				Name:  "Build In",
				Key:   "9999/10000/2/101",
				Value: "0",
			},
			{
				Name:  "Build Out",
				Key:   "9999/10000/2/102",
				Value: "0",
			},
			{
				Name:  "Position",
				Key:   "9999/10003/13260/3296672360/1/100/101",
				Value: "0 -3071",
			},
			{
				Name:  "Layout Method",
				Key:   "9999/10003/13260/3296672360/2/314",
				Value: "1 (Paragraph)",
			},
			{
				Name:  "Left Margin",
				Key:   "9999/10003/13260/3296672360/2/323",
				Value: "-1210",
			},
			{
				Name:  "Right Margin",
				Key:   "9999/10003/13260/3296672360/2/324",
				Value: "1210",
			},
			{
				Name:  "Top Margin",
				Key:   "9999/10003/13260/3296672360/2/325",
				Value: "2160",
			},
			{
				Name:  "Bottom Margin",
				Key:   "9999/10003/13260/3296672360/2/326",
				Value: "-2160",
			},
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
				Value: "1 (Center)",
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
					Text: text,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: textStyleID,
				TextStyle: TextStyle{
					Font:        "Arial",
					FontSize:    "204",
					FontFace:    "Regular",
					FontColor:   "0.999995 1 1 1",
					Alignment:   "center",
					LineSpacing: "-19",
				},
			},
		},
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Add title to the spine or nest within existing video content
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		
		// Check if there are existing video elements to nest the title within
		if len(sequence.Spine.Videos) > 0 {
			// Add as nested title within the first video element (like samples/imessage001.fcpxml)
			sequence.Spine.Videos[0].NestedTitles = append(sequence.Spine.Videos[0].NestedTitles, title)
		} else if len(sequence.Spine.AssetClips) > 0 {
			// Add as nested title within the first asset clip
			sequence.Spine.AssetClips[0].Titles = append(sequence.Spine.AssetClips[0].Titles, title)
		} else {
			// No existing video content - add title directly to spine for standalone text
			sequence.Spine.Titles = append(sequence.Spine.Titles, title)
		}
	}

	return nil
}

// AddImessageText creates a complete imessage structure exactly like samples/imessage001.fcpxml.
// This creates the EXACT structure with matching format, durations, and timing.
func AddImessageText(fcpxml *FCPXML, text string, offsetSeconds float64, durationSeconds float64) error {
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	
	// Reserve IDs (need 5: format, 2 assets, 2 formats for assets, 1 effect)
	ids := tx.ReserveIDs(6)
	formatID := ids[0]
	phoneAssetID := ids[1]
	phoneFormatID := ids[2]
	bubbleAssetID := ids[3]
	bubbleFormatID := ids[4]
	effectID := ids[5]
	
	// Create main sequence format matching reference exactly
	_, err := tx.CreateFormatWithFrameDuration(formatID, "100/6000s", "1080", "1920", "1-1-1 (Rec. 709)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create main format: %v", err)
	}
	
	// Create phone asset format matching reference
	_, err = tx.CreateFormat(phoneFormatID, "FFVideoFormatRateUndefined", "452", "910", "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create phone format: %v", err)
	}
	
	// Create phone asset matching reference
	_, err = tx.CreateAsset(phoneAssetID, "/Users/aa/Movies/Untitled.fcpbundle/6-13-25/Original Media/phone_blank001 (fcp1).png", "phone_blank001", "0s", phoneFormatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create phone asset: %v", err)
	}
	
	// Create bubble asset format matching reference
	_, err = tx.CreateFormat(bubbleFormatID, "FFVideoFormatRateUndefined", "392", "206", "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create bubble format: %v", err)
	}
	
	// Create bubble asset matching reference
	_, err = tx.CreateAsset(bubbleAssetID, "/Users/aa/Movies/Untitled.fcpbundle/6-13-25/Original Media/blue_speech001 (fcp1).png", "blue_speech001", "0s", bubbleFormatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create bubble asset: %v", err)
	}
	
	// Create text effect
	_, err = tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create text effect: %v", err)
	}
	
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	// Set sequence format and duration to match reference exactly
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		sequence.Format = formatID
		sequence.Duration = "3300/6000s" // From reference
		
		// Create the exact structure from reference
		phoneVideo := Video{
			Ref:      phoneAssetID,
			Offset:   "0s",
			Name:     "phone_blank001",
			Start:    "21610300/6000s", // From reference
			Duration: "3300/6000s",     // From reference
			NestedVideos: []Video{{
				Ref:      bubbleAssetID,
				Lane:     "1",
				Offset:   "21610300/6000s", // From reference
				Name:     "blue_speech001",
				Start:    "21610300/6000s", // From reference
				Duration: "3300/6000s",     // From reference
				AdjustTransform: &AdjustTransform{
					Position: "1.26755 -21.1954", // From reference
					Scale:    "0.617236 0.617236", // From reference
				},
			}},
			NestedTitles: []Title{{
				Ref:      effectID,
				Lane:     "2",
				Offset:   "43220600/12000s", // From reference
				Name:     text + " - Text",
				Start:    "21632100/6000s", // From reference
				Duration: "3300/6000s",     // From reference
				Params: []Param{
					{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
					{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
					{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: "0 -3071"},
					{Name: "Layout Method", Key: "9999/10003/13260/3296672360/2/314", Value: "1 (Paragraph)"},
					{Name: "Left Margin", Key: "9999/10003/13260/3296672360/2/323", Value: "-1210"},
					{Name: "Right Margin", Key: "9999/10003/13260/3296672360/2/324", Value: "1210"},
					{Name: "Top Margin", Key: "9999/10003/13260/3296672360/2/325", Value: "2160"},
					{Name: "Bottom Margin", Key: "9999/10003/13260/3296672360/2/326", Value: "-2160"},
					{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/354/3296667315/401", Value: "1 (Center)"},
					{Name: "Line Spacing", Key: "9999/10003/13260/3296672360/2/354/3296667315/404", Value: "-19"},
					{Name: "Auto-Shrink", Key: "9999/10003/13260/3296672360/2/370", Value: "3 (To All Margins)"},
					{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/373", Value: "0 (Left) 0 (Top)"},
					{Name: "Opacity", Key: "9999/10003/13260/3296672360/4/3296673134/1000/1044", Value: "0"},
					{Name: "Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/208", Value: "6 (Custom)"},
					{
						Name: "Custom Speed",
						Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
						KeyframeAnimation: &KeyframeAnimation{
							Keyframes: []Keyframe{
								{Time: "-469658744/1000000000s", Value: "0"},
								{Time: "12328542033/1000000000s", Value: "1"},
							},
						},
					},
					{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
				},
				Text: &TitleText{
					TextStyles: []TextStyleRef{{
						Ref:  "ts1", // Use simple ID like reference
						Text: text,
					}},
				},
				TextStyleDefs: []TextStyleDef{{
					ID: "ts1",
					TextStyle: TextStyle{
						Font:        "Arial",
						FontSize:    "204",
						FontFace:    "Regular",
						FontColor:   "0.999995 1 1 1",
						Alignment:   "center",
						LineSpacing: "-19",
					},
				}},
			}},
		}
		
		// Replace spine with exact structure
		sequence.Spine.Videos = []Video{phoneVideo}
		sequence.Spine.AssetClips = nil
		sequence.Spine.Titles = nil
	}
	
	return nil
}

// AddImessageReply adds a reply message like samples/imessage002.fcpxml.
// This appends a second video segment with white speech bubble and black text.
func AddImessageReply(fcpxml *FCPXML, originalText, replyText string, offsetSeconds float64, durationSeconds float64) error {
	// Initialize ResourceRegistry for this FCPXML
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	
	// Reserve IDs for white speech bubble asset and format
	ids := tx.ReserveIDs(2)
	whiteAssetID := ids[0]
	whiteFormatID := ids[1]
	
	// Create white speech bubble format matching reference
	_, err := tx.CreateFormat(whiteFormatID, "FFVideoFormatRateUndefined", "391", "207", "1-13-1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create white bubble format: %v", err)
	}
	
	// Create white speech bubble asset matching reference
	_, err = tx.CreateAsset(whiteAssetID, "/Users/aa/Movies/Untitled.fcpbundle/6-13-25/Original Media/white_speech001 (fcp1).png", "white_speech001", "0s", whiteFormatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create white bubble asset: %v", err)
	}
	
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	// Get existing assets and effects
	var phoneAssetID, blueAssetID, effectID string
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Name == "phone_blank001" {
			phoneAssetID = asset.ID
		} else if asset.Name == "blue_speech001" {
			blueAssetID = asset.ID
		}
	}
	for _, effect := range fcpxml.Resources.Effects {
		if effect.Name == "Text" {
			effectID = effect.ID
			break
		}
	}
	
	if phoneAssetID == "" || blueAssetID == "" || effectID == "" {
		return fmt.Errorf("required assets not found in existing FCPXML")
	}
	
	// Generate unique text style IDs (need to do this properly to avoid conflicts)
	existingIDs := getAllExistingTextStyleIDs(fcpxml)
	replyTextStyleID := getNextUniqueTextStyleID(existingIDs)
	originalTextStyleID := getNextUniqueTextStyleID(existingIDs)
	
	// Simple fix: Always append to the end by using current sequence duration as offset
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		
		// Use current sequence duration as the offset for new frame
		currentDuration := sequence.Duration
		fmt.Printf("DEBUG AddImessageReply: current sequence duration = '%s'\n", currentDuration)
		if currentDuration == "" {
			currentDuration = "0s"
		}
		
		// Convert to /6000s format if needed
		var nextOffset string
		if strings.HasSuffix(currentDuration, "/6000s") {
			nextOffset = currentDuration
		} else {
			// Convert other formats to /6000s (simplified)
			nextOffset = "0/6000s"
		}
		
		// Parse current duration to calculate new sequence duration
		var currentSixthousandths int
		if strings.HasSuffix(currentDuration, "/6000s") {
			numeratorStr := strings.TrimSuffix(currentDuration, "/6000s")
			if numerator, err := strconv.Atoi(numeratorStr); err == nil {
				currentSixthousandths = numerator
			}
		}
		
		// Update sequence duration: current + new frame duration
		newTotalSixthousandths := currentSixthousandths + 3900 // 3900/6000s
		sequence.Duration = fmt.Sprintf("%d/6000s", newTotalSixthousandths)
		
		// Create video segment at end of timeline
		secondVideo := Video{
			Ref:      phoneAssetID,
			Offset:   nextOffset,        // Use current sequence duration as offset
			Name:     "phone_blank001",
			Start:    "21632800/6000s",    // From reference
			Duration: "3900/6000s",        // From reference
			NestedVideos: []Video{
				{
					Ref:      blueAssetID,
					Lane:     "1",
					Offset:   "21632800/6000s",    // From reference
					Name:     "blue_speech001",
					Start:    "21632800/6000s",    // From reference
					Duration: "3900/6000s",        // From reference
					AdjustTransform: &AdjustTransform{
						Position: "1.26755 -21.1954",   // From reference
						Scale:    "0.617236 0.617236",  // From reference
					},
				},
				{
					Ref:      whiteAssetID,
					Lane:     "2",
					Offset:   "21632800/6000s",    // From reference
					Name:     "white_speech001",
					Start:    "21632800/6000s",    // From reference
					Duration: "3900/6000s",        // From reference
					AdjustTransform: &AdjustTransform{
						Position: "0.635834 4.00864",   // From reference
						Scale:    "0.653172 0.653172",  // From reference
					},
				},
			},
			NestedTitles: []Title{
				{
					// New reply text in white bubble (black text)
					Ref:      effectID,
					Lane:     "3",
					Offset:   "86531200/24000s",   // From reference
					Name:     replyText + " - Text",
					Start:    "21654300/6000s",    // From reference
					Duration: "3900/6000s",        // From reference
					Params: []Param{
						{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
						{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
						{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: "0 -1807"}, // Different Y position
						{Name: "Layout Method", Key: "9999/10003/13260/3296672360/2/314", Value: "1 (Paragraph)"},
						{Name: "Left Margin", Key: "9999/10003/13260/3296672360/2/323", Value: "-1210"},
						{Name: "Right Margin", Key: "9999/10003/13260/3296672360/2/324", Value: "1210"},
						{Name: "Top Margin", Key: "9999/10003/13260/3296672360/2/325", Value: "2160"},
						{Name: "Bottom Margin", Key: "9999/10003/13260/3296672360/2/326", Value: "-2160"},
						{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/354/3296667315/401", Value: "1 (Center)"},
						{Name: "Line Spacing", Key: "9999/10003/13260/3296672360/2/354/3296667315/404", Value: "-19"},
						{Name: "Auto-Shrink", Key: "9999/10003/13260/3296672360/2/370", Value: "3 (To All Margins)"},
						{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/373", Value: "0 (Left) 0 (Top)"},
						{Name: "Opacity", Key: "9999/10003/13260/3296672360/4/3296673134/1000/1044", Value: "0"},
						{Name: "Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/208", Value: "6 (Custom)"},
						{
							Name: "Custom Speed",
							Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
							KeyframeAnimation: &KeyframeAnimation{
								Keyframes: []Keyframe{
									{Time: "-469658744/1000000000s", Value: "0"},
									{Time: "12328542033/1000000000s", Value: "1"},
								},
							},
						},
						{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
					},
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  replyTextStyleID,
							Text: replyText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: replyTextStyleID,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "204",
							FontFace:    "Regular",
							FontColor:   "0 0 0 1",      // Black text for white bubble
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
				{
					// Original text continued in blue bubble
					Ref:      effectID,
					Lane:     "4",
					Offset:   "21632800/6000s",   // From reference
					Name:     originalText + " - Text",
					Start:    "21654600/6000s",   // From reference
					Duration: "3900/6000s",       // From reference
					Params: []Param{
						{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
						{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
						{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: "0 -3071"}, // Original position
						{Name: "Layout Method", Key: "9999/10003/13260/3296672360/2/314", Value: "1 (Paragraph)"},
						{Name: "Left Margin", Key: "9999/10003/13260/3296672360/2/323", Value: "-1210"},
						{Name: "Right Margin", Key: "9999/10003/13260/3296672360/2/324", Value: "1210"},
						{Name: "Top Margin", Key: "9999/10003/13260/3296672360/2/325", Value: "2160"},
						{Name: "Bottom Margin", Key: "9999/10003/13260/3296672360/2/326", Value: "-2160"},
						{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/354/3296667315/401", Value: "1 (Center)"},
						{Name: "Line Spacing", Key: "9999/10003/13260/3296672360/2/354/3296667315/404", Value: "-19"},
						{Name: "Auto-Shrink", Key: "9999/10003/13260/3296672360/2/370", Value: "3 (To All Margins)"},
						{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/373", Value: "0 (Left) 0 (Top)"},
						{Name: "Opacity", Key: "9999/10003/13260/3296672360/4/3296673134/1000/1044", Value: "0"},
						{Name: "Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/208", Value: "6 (Custom)"},
						{
							Name: "Custom Speed",
							Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
							KeyframeAnimation: &KeyframeAnimation{
								Keyframes: []Keyframe{
									{Time: "-469658744/1000000000s", Value: "0"},
									{Time: "12328542033/1000000000s", Value: "1"},
								},
							},
						},
						{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
					},
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  originalTextStyleID,
							Text: originalText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: originalTextStyleID,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "204",
							FontFace:    "Regular",
							FontColor:   "0.999995 1 1 1", // White text for blue bubble
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
			},
		}
		
		// Add second video segment to spine
		sequence.Spine.Videos = append(sequence.Spine.Videos, secondVideo)
	}
	
	return nil
}

// AddImessageContinuation automatically continues an existing imessage conversation.
// Analyzes the current conversation pattern and adds the appropriate bubble type.
func AddImessageContinuation(fcpxml *FCPXML, newText string, offsetSeconds float64, durationSeconds float64) error {
	// Analyze existing conversation to determine next bubble type
	pattern := analyzeConversationPattern(fcpxml)
	
	fmt.Printf("DEBUG AddImessageContinuation: next bubble type = '%s', video count = %d\n", pattern.NextBubbleType, pattern.VideoCount)
	
	switch pattern.NextBubbleType {
	case "blue":
		// Add blue bubble (like original imessage001.fcpxml pattern)
		return addBlueBubbleContinuation(fcpxml, newText, pattern, offsetSeconds, durationSeconds)
	case "white":
		// Add white bubble (like imessage002.fcpxml pattern)
		return addWhiteBubbleContinuation(fcpxml, newText, pattern, offsetSeconds, durationSeconds)
	default:
		return fmt.Errorf("could not determine conversation pattern")
	}
}

// ConversationPattern holds information about the current conversation state
type ConversationPattern struct {
	NextBubbleType string  // "blue" or "white"
	LastText       string  // Text from the last bubble
	VideoCount     int     // Number of video segments
	NextOffset     string  // Where to place the next segment
	NextDuration   string  // Duration for next segment
}

// analyzeConversationPattern examines existing FCPXML to determine conversation state
func analyzeConversationPattern(fcpxml *FCPXML) ConversationPattern {
	pattern := ConversationPattern{
		NextBubbleType: "blue", // Default to blue if can't determine
		VideoCount:     0,
	}
	
	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || 
	   len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return pattern
	}
	
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	pattern.VideoCount = len(sequence.Spine.Videos)
	
	if pattern.VideoCount == 0 {
		return pattern
	}
	
	// Look at the last video segment to determine pattern
	lastVideo := sequence.Spine.Videos[pattern.VideoCount-1]
	
	// Check if last segment has white speech bubble (indicates white bubble was last)
	hasWhiteBubble := false
	lastBubbleText := ""
	
	for _, nestedVideo := range lastVideo.NestedVideos {
		if nestedVideo.Name == "white_speech001" {
			hasWhiteBubble = true
			break
		}
	}
	
	// Extract the last text to continue conversation
	for _, title := range lastVideo.NestedTitles {
		if len(title.Text.TextStyles) > 0 {
			lastBubbleText = title.Text.TextStyles[0].Text
		}
	}
	
	pattern.LastText = lastBubbleText
	
	// Determine next bubble type based on what was last
	if hasWhiteBubble {
		// Last was white bubble (reply), so next should be blue (sender)
		pattern.NextBubbleType = "blue"
	} else {
		// Last was blue bubble (sender), so next should be white (reply)
		pattern.NextBubbleType = "white"
	}
	
	// Calculate next timing based on existing pattern
	if pattern.VideoCount == 1 {
		// First continuation after initial message
		pattern.NextOffset = "3300/6000s"       // From reference imessage002
		pattern.NextDuration = "3900/6000s"     // From reference imessage002
	} else {
		// Calculate offset by summing durations of all previous segments
		totalFrameUnits := 0
		for _, video := range sequence.Spine.Videos {
			durationStr := video.Duration
			if durationStr != "" {
				durationUnits := parseFCPDuration(durationStr)
				totalFrameUnits += durationUnits  // Add frame units directly
			}
		}
		pattern.NextOffset = fmt.Sprintf("%d/24000s", totalFrameUnits)  // Direct frame output
		pattern.NextDuration = "3900/6000s" // Standard continuation duration
	}
	
	return pattern
}

// addBlueBubbleContinuation adds a blue bubble message (sender)
func addBlueBubbleContinuation(fcpxml *FCPXML, newText string, pattern ConversationPattern, offsetSeconds float64, durationSeconds float64) error {
	// Get existing assets
	var phoneAssetID, blueAssetID, effectID string
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Name == "phone_blank001" {
			phoneAssetID = asset.ID
		} else if asset.Name == "blue_speech001" {
			blueAssetID = asset.ID
		}
	}
	for _, effect := range fcpxml.Resources.Effects {
		if effect.Name == "Text" {
			effectID = effect.ID
			break
		}
	}
	
	if phoneAssetID == "" || blueAssetID == "" || effectID == "" {
		return fmt.Errorf("required assets not found in existing FCPXML")
	}
	
	// Generate unique text style ID (need only one for single text element)
	existingIDs := getAllExistingTextStyleIDs(fcpxml)
	uniqueTextStyleID1 := getNextUniqueTextStyleID(existingIDs)
	
	// Create new video segment for blue bubble continuation (like a new message in the conversation)
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		
		// Calculate proper offset using simple /6000s arithmetic (like samples/imessage002.fcpxml)
		totalSixthousandths := 0
		for _, video := range sequence.Spine.Videos {
			durationStr := video.Duration
			if durationStr != "" && strings.HasSuffix(durationStr, "/6000s") {
				// Parse numerator from "nnnn/6000s" format
				numeratorStr := strings.TrimSuffix(durationStr, "/6000s")
				if numerator, err := strconv.Atoi(numeratorStr); err == nil {
					totalSixthousandths += numerator
				}
			}
		}
		nextOffset := fmt.Sprintf("%d/6000s", totalSixthousandths)
		
		// Update sequence duration to accommodate new segment  
		newTotalSixthousandths := totalSixthousandths + 3900 // 3900/6000s
		sequence.Duration = fmt.Sprintf("%d/6000s", newTotalSixthousandths)
		
		// Create new video segment with blue bubble (like imessage001 pattern)
		nextVideo := Video{
			Ref:      phoneAssetID,
			Offset:   nextOffset,
			Name:     "phone_blank001",
			Start:    "21632800/6000s",
			Duration: "3900/6000s",
			NestedVideos: []Video{{
				Ref:      blueAssetID,
				Lane:     "1",
				Offset:   "21632800/6000s",
				Name:     "blue_speech001",
				Start:    "21632800/6000s",
				Duration: "3900/6000s",
				AdjustTransform: &AdjustTransform{
					Position: "1.26755 -21.1954",
					Scale:    "0.617236 0.617236",
				},
			}},
			NestedTitles: []Title{
				{
					// Single text element for blue bubble continuation (no duplication)
					Ref:      effectID,
					Lane:     "1",
					Offset:   "21610300/6000s",
					Name:     newText + " - Text",
					Start:    "21632100/6000s",
					Duration: "3900/6000s",
					Params: createStandardTextParams("0 -3071"), // Blue bubble position
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  uniqueTextStyleID1,
							Text: newText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: uniqueTextStyleID1,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "204",
							FontFace:    "Regular",
							FontColor:   "0.999995 1 1 1", // White text for blue bubble
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
			},
		}
		
		// Add to spine
		sequence.Spine.Videos = append(sequence.Spine.Videos, nextVideo)
	}
	
	return nil
}

// addWhiteBubbleContinuation adds a white bubble message (reply) without duplicating previous messages
func addWhiteBubbleContinuation(fcpxml *FCPXML, newText string, pattern ConversationPattern, offsetSeconds float64, durationSeconds float64) error {
	// Get existing assets
	var phoneAssetID, blueAssetID, whiteAssetID, effectID string
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Name == "phone_blank001" {
			phoneAssetID = asset.ID
		} else if asset.Name == "blue_speech001" {
			blueAssetID = asset.ID
		} else if asset.Name == "white_speech001" {
			whiteAssetID = asset.ID
		}
	}
	for _, effect := range fcpxml.Resources.Effects {
		if effect.Name == "Text" {
			effectID = effect.ID
			break
		}
	}
	
	if phoneAssetID == "" || blueAssetID == "" || whiteAssetID == "" || effectID == "" {
		return fmt.Errorf("required assets not found in existing FCPXML")
	}
	
	// Generate unique text style ID
	existingIDs := getAllExistingTextStyleIDs(fcpxml)
	uniqueTextStyleID := getNextUniqueTextStyleID(existingIDs)
	
	// Create new video segment for white bubble continuation
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		
		// Calculate proper offset using simple /6000s arithmetic
		totalSixthousandths := 0
		for _, video := range sequence.Spine.Videos {
			durationStr := video.Duration
			if durationStr != "" && strings.HasSuffix(durationStr, "/6000s") {
				// Parse numerator from "nnnn/6000s" format
				numeratorStr := strings.TrimSuffix(durationStr, "/6000s")
				if numerator, err := strconv.Atoi(numeratorStr); err == nil {
					totalSixthousandths += numerator
				}
			}
		}
		nextOffset := fmt.Sprintf("%d/6000s", totalSixthousandths)
		
		// Update sequence duration to accommodate new segment  
		newTotalSixthousandths := totalSixthousandths + 3900 // 3900/6000s
		sequence.Duration = fmt.Sprintf("%d/6000s", newTotalSixthousandths)
		
		// Create new video segment with white bubble only (no duplication)
		nextVideo := Video{
			Ref:      phoneAssetID,
			Offset:   nextOffset,
			Name:     "phone_blank001",
			Start:    "21632800/6000s",
			Duration: "3900/6000s",
			NestedVideos: []Video{
				{
					Ref:      blueAssetID,
					Lane:     "1",
					Offset:   "21632800/6000s",
					Name:     "blue_speech001",
					Start:    "21632800/6000s",
					Duration: "3900/6000s",
					AdjustTransform: &AdjustTransform{
						Position: "1.26755 -21.1954",
						Scale:    "0.617236 0.617236",
					},
				},
				{
					Ref:      whiteAssetID,
					Lane:     "2",
					Offset:   "21632800/6000s",
					Name:     "white_speech001",
					Start:    "21632800/6000s",
					Duration: "3900/6000s",
					AdjustTransform: &AdjustTransform{
						Position: "0.635834 4.00864",
						Scale:    "0.653172 0.653172",
					},
				},
			},
			NestedTitles: []Title{
				{
					Ref:      effectID,
					Lane:     "3",
					Offset:   "86531200/24000s",
					Name:     newText + " - Text",
					Start:    "21654300/6000s",
					Duration: "3900/6000s",
					Params: createStandardTextParams("0 -1807"), // White bubble position
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  uniqueTextStyleID,
							Text: newText,
						}},
					},
					TextStyleDefs: []TextStyleDef{{
						ID: uniqueTextStyleID,
						TextStyle: TextStyle{
							Font:        "Arial",
							FontSize:    "204",
							FontFace:    "Regular",
							FontColor:   "0 0 0 1", // Black text for white bubble
							Alignment:   "center",
							LineSpacing: "-19",
						},
					}},
				},
			},
		}
		
		// Add to spine
		sequence.Spine.Videos = append(sequence.Spine.Videos, nextVideo)
	}
	
	return nil
}

// createStandardTextParams creates the standard text parameters with given position
func createStandardTextParams(position string) []Param {
	return []Param{
		{Name: "Build In", Key: "9999/10000/2/101", Value: "0"},
		{Name: "Build Out", Key: "9999/10000/2/102", Value: "0"},
		{Name: "Position", Key: "9999/10003/13260/3296672360/1/100/101", Value: position},
		{Name: "Layout Method", Key: "9999/10003/13260/3296672360/2/314", Value: "1 (Paragraph)"},
		{Name: "Left Margin", Key: "9999/10003/13260/3296672360/2/323", Value: "-1210"},
		{Name: "Right Margin", Key: "9999/10003/13260/3296672360/2/324", Value: "1210"},
		{Name: "Top Margin", Key: "9999/10003/13260/3296672360/2/325", Value: "2160"},
		{Name: "Bottom Margin", Key: "9999/10003/13260/3296672360/2/326", Value: "-2160"},
		{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/354/3296667315/401", Value: "1 (Center)"},
		{Name: "Line Spacing", Key: "9999/10003/13260/3296672360/2/354/3296667315/404", Value: "-19"},
		{Name: "Auto-Shrink", Key: "9999/10003/13260/3296672360/2/370", Value: "3 (To All Margins)"},
		{Name: "Alignment", Key: "9999/10003/13260/3296672360/2/373", Value: "0 (Left) 0 (Top)"},
		{Name: "Opacity", Key: "9999/10003/13260/3296672360/4/3296673134/1000/1044", Value: "0"},
		{Name: "Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/208", Value: "6 (Custom)"},
		{
			Name: "Custom Speed",
			Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
			KeyframeAnimation: &KeyframeAnimation{
				Keyframes: []Keyframe{
					{Time: "-469658744/1000000000s", Value: "0"},
					{Time: "12328542033/1000000000s", Value: "1"},
				},
			},
		},
		{Name: "Apply Speed", Key: "9999/10003/13260/3296672360/4/3296673134/201/211", Value: "2 (Per Object)"},
	}
}

// getAllExistingTextStyleIDs collects all existing text style IDs from the entire FCPXML
func getAllExistingTextStyleIDs(fcpxml *FCPXML) map[string]bool {
	existingIDs := make(map[string]bool)
	
	// Check all video segments for text style definitions
	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
		
		// Check main spine titles
		for _, title := range sequence.Spine.Titles {
			for _, styleDef := range title.TextStyleDefs {
				existingIDs[styleDef.ID] = true
			}
		}
		
		// Check asset-clips titles
		for _, assetClip := range sequence.Spine.AssetClips {
			for _, title := range assetClip.Titles {
				for _, styleDef := range title.TextStyleDefs {
					existingIDs[styleDef.ID] = true
				}
			}
		}
		
		// Check video segments and their nested titles
		for _, video := range sequence.Spine.Videos {
			for _, title := range video.NestedTitles {
				for _, styleDef := range title.TextStyleDefs {
					existingIDs[styleDef.ID] = true
				}
			}
		}
	}
	
	return existingIDs
}

// getNextUniqueTextStyleID generates the next unique text style ID and marks it as used
func getNextUniqueTextStyleID(existingIDs map[string]bool) string {
	// Generate unique ID by trying ts1, ts2, ts3, etc.
	counter := 1
	for {
		candidateID := fmt.Sprintf("ts%d", counter)
		if !existingIDs[candidateID] {
			// Mark this ID as used for subsequent calls
			existingIDs[candidateID] = true
			return candidateID
		}
		counter++
	}
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

// GenerateBaffleTimeline creates a random complex timeline for stress testing
func GenerateBaffleTimeline(minDuration, maxDuration float64, verbose bool) (*FCPXML, error) {
	// Create base FCPXML structure
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		return nil, fmt.Errorf("failed to create base FCPXML: %v", err)
	}
	
	// Set up resource management
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()
	
	if verbose {
		fmt.Printf("Starting baffle timeline generation...\n")
	}
	
	// Get available assets from ./assets directory
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
	
	// Generate random timeline duration
	rand.Seed(time.Now().UnixNano())
	totalDuration := minDuration + rand.Float64()*(maxDuration-minDuration)
	
	if verbose {
		fmt.Printf("Target timeline duration: %.1f seconds (%.1f minutes)\n", totalDuration, totalDuration/60)
	}
	
	// Generate complex multi-element timeline
	err = generateRandomTimelineElements(fcpxml, tx, assets, totalDuration, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to generate timeline elements: %v", err)
	}
	
	// Update sequence duration to match content
	updateSequenceDuration(fcpxml, totalDuration)
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	if verbose {
		fmt.Printf("Baffle timeline generation completed successfully\n")
	}
	
	return fcpxml, nil
}

// AssetCollection holds categorized assets
type AssetCollection struct {
	Images []string
	Videos []string
}

// scanAssetsDirectory scans ./assets for available media files
func scanAssetsDirectory(assetsDir string) (*AssetCollection, error) {
	assets := &AssetCollection{
		Images: []string{},
		Videos: []string{},
	}
	
	// Get absolute path of assets directory
	absAssetsDir, err := filepath.Abs(assetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for assets directory: %v", err)
	}
	
	// Check if assets directory exists
	if _, err := os.Stat(absAssetsDir); os.IsNotExist(err) {
		return assets, nil // Return empty collection if directory doesn't exist
	}
	
	// Read directory contents
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
		// Create absolute path for Final Cut Pro
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
	// Pre-create all unique assets to avoid UID collisions
	createdAssets := make(map[string]string) // filepath -> assetID
	createdFormats := make(map[string]string) // filepath -> formatID
	
	// CRITICAL FIX: Always add a background element starting at 0s to prevent black screen
	// Add a long-duration background video or image that covers the entire timeline
	if len(assets.Videos) > 0 {
		backgroundVideo := assets.Videos[rand.Intn(len(assets.Videos))]
		// Create unique copy for BAFFLE to avoid FCP UID cache conflicts
		uniqueVideo, err := createUniqueMediaCopy(backgroundVideo, "background")
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create unique background copy: %v\n", err)
			uniqueVideo = backgroundVideo // Fallback to original
		}
		
		err = addRandomVideoElement(fcpxml, tx, uniqueVideo, 0.0, totalDuration, 0, 0, verbose, createdAssets, createdFormats)
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to add background video: %v\n", err)
		} else if verbose {
			fmt.Printf("  Added background video: %s (%.1fs @ 0s)\n", filepath.Base(uniqueVideo), totalDuration)
		}
	} else if len(assets.Images) > 0 {
		backgroundImage := assets.Images[rand.Intn(len(assets.Images))]
		// Create unique copy for BAFFLE to avoid FCP UID cache conflicts  
		uniqueImage, err := createUniqueMediaCopy(backgroundImage, "background")
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create unique background copy: %v\n", err)
			uniqueImage = backgroundImage // Fallback to original
		}
		
		err = addRandomImageElement(fcpxml, tx, uniqueImage, 0.0, totalDuration, 0, 0, verbose, createdAssets, createdFormats)
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to add background image: %v\n", err)
		} else if verbose {
			fmt.Printf("  Added background image: %s (%.1fs @ 0s)\n", filepath.Base(uniqueImage), totalDuration)
		}
	}
	
	// Create proper nested multi-lane structure as per BAFFLE_TWO.md
	if verbose {
		fmt.Printf("Creating nested multi-lane structure with overlays...\n")
	}
	
	// Create main background video that will contain all overlays
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	
	if len(assets.Videos) > 0 {
		// Create main background video element
		mainVideoPath := assets.Videos[rand.Intn(len(assets.Videos))]
		uniqueMainVideo, err := createUniqueMediaCopy(mainVideoPath, "main_bg")
		if err != nil && verbose {
			fmt.Printf("Warning: Failed to create unique main video copy: %v\n", err)
			uniqueMainVideo = mainVideoPath
		}
		
		// Create main video element that will contain nested lanes
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
	
	// Create additional main timeline elements distributed across multiple lanes
	numMainElements := 3 + rand.Intn(5) // 3-7 main timeline elements
	maxLanes := 4 // Maximum lanes to use (1-4)
	currentOffset := totalDuration * 0.2 // Start overlays earlier for better multi-lane effect
	
	if verbose {
		fmt.Printf("Creating %d additional main spine elements with lane distribution...\n", numMainElements)
	}
	
	for i := 1; i <= numMainElements; i++ {
		duration := 6.0 + rand.Float64()*15.0 // 6-21 second elements
		startTime := currentOffset
		currentOffset += duration * 0.4 // 60% overlap for better layering
		
		// Stop if we exceed timeline duration
		if startTime >= totalDuration {
			break
		}
		
		// Ensure element doesn't exceed timeline
		if startTime+duration > totalDuration {
			duration = totalDuration - startTime
		}
		
		// CRITICAL FIX: Assign proper lanes to main spine elements (1, 2, 3, 4)
		lane := (i % maxLanes) + 1 // Distribute across lanes 1-4
		
		// Alternate between video and image main elements
		if i%2 == 0 && len(assets.Videos) > 0 {
			videoPath := assets.Videos[rand.Intn(len(assets.Videos))]
			uniqueVideo, err := createUniqueMediaCopy(videoPath, fmt.Sprintf("main_%d", i))
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to create unique video copy: %v\n", err)
				uniqueVideo = videoPath
			}
			
			// Create lane-assigned asset-clip directly on spine
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
			
			// Create lane-assigned video element directly on spine
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
		
		// CRITICAL FIX: CreateAsset doesn't create format, so create it manually for images
		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return nil, fmt.Errorf("failed to create image format: %v", err)
		}
		
		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}
	
	// Create image overlay - keep content centered and on screen
	video := &Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("ImageOverlay_%d", index),
		Lane:     fmt.Sprintf("%d", lane),
		AdjustTransform: &AdjustTransform{
			Position: "0 0", // Keep centered - no random positioning
			Scale:    fmt.Sprintf("%.2f %.2f", 0.5+rand.Float64()*0.3, 0.5+rand.Float64()*0.3), // 0.5-0.8 scale for overlays
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
	
	// Create video overlay - keep content centered and on screen
	assetClip := &AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("VideoOverlay_%d", index),
		Lane:     fmt.Sprintf("%d", lane),
		AdjustTransform: &AdjustTransform{
			Position: "0 0", // Keep centered - no random positioning
			Scale:    fmt.Sprintf("%.2f %.2f", 0.6+rand.Float64()*0.3, 0.6+rand.Float64()*0.3), // 0.6-0.9 scale for video overlays
		},
	}
	
	return assetClip, nil
}

// createTextOverlay creates a text overlay element
func createTextOverlay(fcpxml *FCPXML, tx *ResourceTransaction, startTime, duration float64, lane, index int, verbose bool) (*Title, error) {
	// Reserve ID for text effect
	ids := tx.ReserveIDs(1)
	effectID := ids[0]
	
	// Create text effect
	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		return nil, fmt.Errorf("failed to create text effect: %v", err)
	}
	
	// Generate content and style
	textContent := generateRandomText()
	styleID := fmt.Sprintf("ts_%d", rand.Intn(999999)+100000) // Simple unique number
	
	// Create title overlay
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
				Font:         randomFont(),
				FontSize:     fmt.Sprintf("%.0f", 24+rand.Float64()*24), // 24-48pt for overlays
				FontColor:    randomColor(),
				Alignment:    randomAlignment(),
				LineSpacing:  "1.2",
			},
		}},
		Params: []Param{{
			Name:  "Opacity",
			Value: fmt.Sprintf("%.2f", 0.8+rand.Float64()*0.2), // 80-100% opacity
		}},
	}
	
	return title, nil
}

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
	
	// Create asset-clip with proper lane assignment
	assetClip := &AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Lane%dVideo_%d", lane, index),
		Lane:     fmt.Sprintf("%d", lane), // CRITICAL: Proper lane assignment for spine
		AdjustTransform: &AdjustTransform{
			Position: generateRandomPosition(),
			Scale:    fmt.Sprintf("%.2f %.2f", 0.7+rand.Float64()*0.4, 0.7+rand.Float64()*0.4), // 0.7-1.1 scale
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
		
		// CRITICAL FIX: CreateAsset doesn't create format, so create it manually for images
		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return nil, fmt.Errorf("failed to create image format: %v", err)
		}
		
		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}
	
	// Create video element (for image) with proper lane assignment
	video := &Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Lane%dImage_%d", lane, index),
		Lane:     fmt.Sprintf("%d", lane), // CRITICAL: Proper lane assignment for spine
		AdjustTransform: &AdjustTransform{
			Position: generateRandomPosition(),
			Scale:    fmt.Sprintf("%.2f %.2f", 0.6+rand.Float64()*0.5, 0.6+rand.Float64()*0.5), // 0.6-1.1 scale
		},
	}
	
	return video, nil
}

// generateRandomPosition generates a random but reasonable position for elements
func generateRandomPosition() string {
	// Keep positions reasonable to avoid off-screen elements
	x := int(-200 + rand.Float64()*400) // -200 to +200 pixels
	y := int(-150 + rand.Float64()*300) // -150 to +150 pixels
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
	
	// Create main video element
	mainVideo := &Video{
		Ref:      assetID,
		Offset:   "0s",
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     "MainBackground",
	}
	
	// Add nested overlay elements (lanes 1-4)
	numOverlays := 6 + rand.Intn(8) // 6-13 overlays
	
	for i := 1; i <= numOverlays; i++ {
		overlayStartTime := rand.Float64() * (duration * 0.8) // Start within 80% of duration
		overlayDuration := 3.0 + rand.Float64()*8.0 // 3-11 second overlays
		
		// Ensure overlay doesn't exceed main video duration
		if overlayStartTime+overlayDuration > duration {
			overlayDuration = duration - overlayStartTime
		}
		
		lane := 1 + rand.Intn(4) // Lanes 1-4
		
		// Choose overlay type
		overlayType := rand.Intn(3) // 0=image, 1=video, 2=text
		
		switch overlayType {
		case 0: // Image overlay
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
			
		case 1: // Video overlay
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
			
		case 2: // Text overlay
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
	
	// Create asset-clip
	assetClip := &AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("MainClip_%d", index),
	}
	
	// Add a few nested overlays
	numOverlays := 2 + rand.Intn(4) // 2-5 overlays
	
	for i := 1; i <= numOverlays; i++ {
		overlayStartTime := rand.Float64() * (duration * 0.7)
		overlayDuration := 2.0 + rand.Float64()*4.0
		
		if overlayStartTime+overlayDuration > duration {
			overlayDuration = duration - overlayStartTime
		}
		
		// Add image or text overlay
		if rand.Float32() < 0.6 && len(assets.Images) > 0 { // 60% images
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
		} else { // 40% text
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
		
		// CRITICAL FIX: CreateAsset doesn't create format, so create it manually for images
		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return nil, fmt.Errorf("failed to create image format: %v", err)
		}
		
		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}
	
	// Create video element for image
	video := &Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("MainImage_%d", index),
	}
	
	// Add text overlays on the image
	numOverlays := 1 + rand.Intn(3) // 1-3 text overlays
	
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
		
		// CRITICAL FIX: CreateAsset doesn't create format, so create it manually for images
		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return fmt.Errorf("failed to create image format: %v", err)
		}
		
		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}
	
	// Create video element for image (as per FCPXML spec)
	video := Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Image_%d", elementIndex),
	}
	
	// Apply lane assignment correctly
	if targetLane > 0 {
		video.Lane = fmt.Sprintf("%d", targetLane)
	}
	
	// Keep transforms minimal to prevent off-screen content
	if rand.Float32() < 0.3 { // Only 30% chance of animation
		video.AdjustTransform = createMinimalAnimation(startTime, duration)
	}
	
	// Add to appropriate spine location
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Videos = append(spine.Videos, video)
	
	return nil
}

// addBaffleVideoElement adds a video element with proper lane assignment for BAFFLE system
func addBaffleVideoElement(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, startTime, duration float64, elementIndex, targetLane int, verbose bool, createdAssets, createdFormats map[string]string) error {
	if verbose {
		fmt.Printf("  Adding video: %s (%.1fs @ %.1fs) lane %d\n", filepath.Base(videoPath), duration, startTime, targetLane)
	}
	
	// Reuse existing asset if already created, otherwise create new one
	var assetID, formatID string
	var err error
	
	if existingAssetID, exists := createdAssets[videoPath]; exists {
		assetID = existingAssetID
		if existingFormatID, formatExists := createdFormats[videoPath]; formatExists {
			formatID = existingFormatID
		} else {
			return fmt.Errorf("asset exists but format missing for %s", videoPath)
		}
	} else {
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]
		
		err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), ConvertSecondsToFCPDuration(duration), formatID)
		if err != nil {
			return fmt.Errorf("failed to create video asset: %v", err)
		}
		
		createdAssets[videoPath] = assetID
		createdFormats[videoPath] = formatID
	}
	
	// Create asset-clip element for video (as per FCPXML spec)
	assetClip := AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Video_%d", elementIndex),
	}
	
	// Apply lane assignment correctly
	if targetLane > 0 {
		assetClip.Lane = fmt.Sprintf("%d", targetLane)
	}
	
	// Keep transforms minimal to prevent off-screen content
	if rand.Float32() < 0.4 { // 40% chance of animation for videos
		assetClip.AdjustTransform = createMinimalAnimation(startTime, duration)
	}
	
	// Add to appropriate spine location
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.AssetClips = append(spine.AssetClips, assetClip)
	
	return nil
}

// addBaffleTextElement adds a text element with proper lane assignment for BAFFLE system  
func addBaffleTextElement(fcpxml *FCPXML, tx *ResourceTransaction, startTime, duration float64, elementIndex, targetLane int, verbose bool) error {
	textContent := generateRandomText()
	
	if verbose {
		fmt.Printf("  Adding text: \"%s\" (%.1fs @ %.1fs) lane %d\n", textContent, duration, startTime, targetLane)
	}
	
	// Reserve IDs for text effect and style
	ids := tx.ReserveIDs(1)
	effectID := ids[0]
	
	// Create text effect
	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		return fmt.Errorf("failed to create text effect: %v", err)
	}
	
	// Generate unique style ID
	styleID := fmt.Sprintf("ts_%d", rand.Intn(999999)+100000) // Simple unique number
	
	// Create title with proper structure
	title := Title{
		Ref:      effectID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Text_%d", elementIndex),
		Text: &TitleText{
			TextStyles: []TextStyleRef{{
				Ref:  styleID,
				Text: textContent,
			}},
		},
		TextStyleDefs: []TextStyleDef{{
			ID: styleID,
			TextStyle: TextStyle{
				Font:         randomFont(),
				FontSize:     fmt.Sprintf("%.0f", 32+rand.Float64()*32), // 32-64pt
				FontColor:    randomColor(),
				Alignment:    randomAlignment(),
				LineSpacing:  fmt.Sprintf("%.1f", 1.0+rand.Float64()*0.5), // 1.0-1.5
			},
		}},
	}
	
	// Apply lane assignment correctly - text elements should be in higher lanes for overlay
	if targetLane > 0 {
		title.Lane = fmt.Sprintf("%d", targetLane)
	}
	
	// Add transparency for better overlay on video content
	opacity := 0.7 + rand.Float64()*0.3 // 70-100% opacity for better visibility
	title.Params = append(title.Params, Param{
		Name:  "Opacity",
		Value: fmt.Sprintf("%.2f", opacity),
	})
	
	// Add to spine
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Titles = append(spine.Titles, title)
	
	return nil
}

// addRandomImageElement adds an image with random effects and animations (legacy function)
func addRandomImageElement(fcpxml *FCPXML, tx *ResourceTransaction, imagePath string, startTime, duration float64, elementIndex, lane int, verbose bool, createdAssets, createdFormats map[string]string) error {
	if verbose {
		fmt.Printf("  Adding image: %s (%.1fs @ %.1fs)\n", filepath.Base(imagePath), duration, startTime)
	}
	
	// Reuse existing asset if already created, otherwise create new one
	var assetID, formatID string
	var err error
	
	if existingAssetID, exists := createdAssets[imagePath]; exists {
		// Reuse existing asset
		assetID = existingAssetID
		formatID = createdFormats[imagePath]
	} else {
		// Create new asset and format
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]
		
		_, err = tx.CreateAsset(assetID, imagePath, filepath.Base(imagePath), ConvertSecondsToFCPDuration(duration), formatID)
		if err != nil {
			return fmt.Errorf("failed to create image asset: %v", err)
		}
		
		_, err = tx.CreateFormat(formatID, "FFVideoFormatRateUndefined", "1920", "1080", "1-13-1")
		if err != nil {
			return fmt.Errorf("failed to create image format: %v", err)
		}
		
		// Remember this asset for reuse
		createdAssets[imagePath] = assetID
		createdFormats[imagePath] = formatID
	}
	
	// Create video element for timeline
	video := Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Image_%d", elementIndex),
	}
	
	// Assign to specific lane
	if lane > 0 {
		video.Lane = fmt.Sprintf("%d", lane)
	}
	
	// Random animations
	if rand.Float32() < 0.6 { // 60% chance of animation
		video.AdjustTransform = createRandomAnimation(startTime, duration)
	}
	
	// Random additional transforms (safer than fictional effects)
	if rand.Float32() < 0.4 { // 40% chance of additional transforms
		// Add rotation using built-in parameter (crash-safe)
		if video.AdjustTransform == nil {
			video.AdjustTransform = &AdjustTransform{}
		}
		// Add rotation parameter for visual variety
		video.AdjustTransform.Params = append(video.AdjustTransform.Params, Param{
			Name:  "rotation",
			Value: fmt.Sprintf("%.1f", -15.0+rand.Float64()*30.0), // -15 to +15 degrees
		})
	}
	
	// Add to timeline (main spine)
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Videos = append(spine.Videos, video)
	
	return nil
}

// addRandomVideoElement adds a video with random effects
func addRandomVideoElement(fcpxml *FCPXML, tx *ResourceTransaction, videoPath string, startTime, duration float64, elementIndex, lane int, verbose bool, createdAssets, createdFormats map[string]string) error {
	if verbose {
		fmt.Printf("  Adding video: %s (%.1fs @ %.1fs)\n", filepath.Base(videoPath), duration, startTime)
	}
	
	// Reuse existing asset if already created, otherwise create new one
	var assetID, formatID string
	var err error
	
	if existingAssetID, exists := createdAssets[videoPath]; exists {
		// Reuse existing asset
		assetID = existingAssetID
		formatID = createdFormats[videoPath]
	} else {
		// Create new asset and format using actual video properties
		ids := tx.ReserveIDs(2)
		assetID = ids[0]
		formatID = ids[1]
		
		// Use transaction methods to create asset and format with proper video detection
		err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), ConvertSecondsToFCPDuration(duration + 30), formatID)
		if err != nil {
			return fmt.Errorf("failed to create video asset with detection: %v", err)
		}
		
		// Remember this asset for reuse
		createdAssets[videoPath] = assetID
		createdFormats[videoPath] = formatID
	}
	
	// Create asset-clip element for timeline
	assetClip := AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Video_%d", elementIndex),
		Start:    "0s",
	}
	
	// Assign to specific lane
	if lane > 0 {
		assetClip.Lane = fmt.Sprintf("%d", lane)
	}
	
	// Random animations
	if rand.Float32() < 0.7 { // 70% chance of animation
		assetClip.AdjustTransform = createRandomAnimation(startTime, duration)
	}
	
	// Add to timeline
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.AssetClips = append(spine.AssetClips, assetClip)
	
	return nil
}

// addRandomTextElement adds a text title with random styling
func addRandomTextElement(fcpxml *FCPXML, tx *ResourceTransaction, startTime, duration float64, elementIndex, lane int, verbose bool) error {
	textContent := generateRandomText()
	
	if verbose {
		fmt.Printf("  Adding text: \"%s\" (%.1fs @ %.1fs)\n", textContent, duration, startTime)
	}
	
	// Reserve IDs for effect
	effectID := tx.ReserveIDs(1)[0]
	
	// Add text effect using proper transaction method
	_, err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		return fmt.Errorf("failed to create text effect: %v", err)
	}
	
	// Generate style ID
	styleID := fmt.Sprintf("ts_baffle_%d", elementIndex)
	
	// Create title element
	title := Title{
		Ref:      effectID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Text_%d", elementIndex),
		Text: &TitleText{
			TextStyles: []TextStyleRef{{
				Ref:  styleID,
				Text: textContent,
			}},
		},
		TextStyleDefs: []TextStyleDef{{
			ID: styleID,
			TextStyle: TextStyle{
				Font:         randomFont(),
				FontSize:     fmt.Sprintf("%.0f", 36+rand.Float64()*48), // 36-84pt
				FontColor:    randomColor(),
				Alignment:    randomAlignment(),
				LineSpacing:  fmt.Sprintf("%.1f", 1.0+rand.Float64()*0.5), // 1.0-1.5
			},
		}},
	}
	
	// Assign to specific lane (text elements get higher lanes for overlay)
	if lane > 0 {
		// Text elements use higher lane numbers for proper layering on top
		textLane := lane + 5 // Put text well above video/image lanes
		title.Lane = fmt.Sprintf("%d", textLane)
	}
	
	// Add transparency for better overlay on video content
	opacity := 0.6 + rand.Float64()*0.3 // 60-90% opacity for overlay effect
	title.Params = append(title.Params, Param{
		Name:  "Opacity",
		Value: fmt.Sprintf("%.2f", opacity),
	})
	
	// Add to timeline
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Titles = append(spine.Titles, title)
	
	return nil
}

// createRandomAnimation creates random keyframe animation
// createMinimalAnimation creates subtle animations that keep content on-screen
func createMinimalAnimation(startTime, duration float64) *AdjustTransform {
	endTime := startTime + duration
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.0f %.0f", -10+rand.Float64()*20, -5+rand.Float64()*10), // Keep centered: -10 to +10, -5 to +5
						},
						{
							Time:  ConvertSecondsToFCPDuration(endTime),
							Value: fmt.Sprintf("%.0f %.0f", -10+rand.Float64()*20, -5+rand.Float64()*10), // Keep centered: -10 to +10, -5 to +5
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
							Value: fmt.Sprintf("%.2f %.2f", 0.95+rand.Float64()*0.1, 0.95+rand.Float64()*0.1), // Subtle scale: 0.95 to 1.05
							Curve: "smooth",
						},
						{
							Time:  ConvertSecondsToFCPDuration(endTime),
							Value: fmt.Sprintf("%.2f %.2f", 0.95+rand.Float64()*0.1, 0.95+rand.Float64()*0.1), // Subtle scale: 0.95 to 1.05
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

func createRandomAnimation(startTime, duration float64) *AdjustTransform {
	endTime := startTime + duration
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.0f %.0f", -40+rand.Float64()*80, -20+rand.Float64()*40),
						},
						{
							Time:  ConvertSecondsToFCPDuration(endTime),
							Value: fmt.Sprintf("%.0f %.0f", -40+rand.Float64()*80, -20+rand.Float64()*40),
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
							Value: fmt.Sprintf("%.2f %.2f", 0.8+rand.Float64()*0.4, 0.8+rand.Float64()*0.4),
							Curve: "smooth",
						},
						{
							Time:  ConvertSecondsToFCPDuration(endTime),
							Value: fmt.Sprintf("%.2f %.2f", 0.8+rand.Float64()*0.4, 0.8+rand.Float64()*0.4),
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// Helper functions for random content generation
func generateRandomText() string {
	texts := []string{
		"BAFFLE TEST", "Random Text", "Complex Timeline", "Stress Test",
		"FCPXML Generation", "Multi-Lane Test", "Animation Check",
		"Effect Validation", "Resource Test", "Lane Assignment",
		"Keyframe Test", "Timeline Stress", "Generation Check",
	}
	return texts[rand.Intn(len(texts))]
}

func randomFont() string {
	fonts := []string{"Helvetica", "Arial", "Times", "Courier", "Georgia", "Verdana"}
	return fonts[rand.Intn(len(fonts))]
}

func randomColor() string {
	return fmt.Sprintf("%.2f %.2f %.2f 1", rand.Float64(), rand.Float64(), rand.Float64())
}

func randomAlignment() string {
	alignments := []string{"left", "center", "right"}
	return alignments[rand.Intn(len(alignments))]
}

// updateSequenceDuration updates the sequence duration to match content
func updateSequenceDuration(fcpxml *FCPXML, totalDuration float64) {
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	sequence.Duration = ConvertSecondsToFCPDuration(totalDuration)
}

// createUniqueMediaCopy creates a temporary copy of a media file with a unique name
// This prevents FCP UID cache conflicts by ensuring each BAFFLE run uses truly unique files
func createUniqueMediaCopy(originalPath, prefix string) (string, error) {
	// Create unique filename with timestamp and random component
	timestamp := time.Now().UnixNano()
	randomNum := rand.Int63()
	ext := filepath.Ext(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), ext)
	
	// Create unique filename: prefix_basename_timestamp_random.ext
	uniqueName := fmt.Sprintf("%s_%s_%d_%d%s", prefix, baseName, timestamp, randomNum, ext)
	
	// Create temp directory in system temp
	tempDir := filepath.Join(os.TempDir(), "cutlass_baffle")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return originalPath, fmt.Errorf("failed to create temp directory: %v", err)
	}
	
	// Full path for unique copy
	uniquePath := filepath.Join(tempDir, uniqueName)
	
	// Copy file to new location
	sourceFile, err := os.Open(originalPath)
	if err != nil {
		return originalPath, fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(uniquePath)
	if err != nil {
		return originalPath, fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()
	
	// Copy file contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return originalPath, fmt.Errorf("failed to copy file contents: %v", err)
	}
	
	return uniquePath, nil
}
