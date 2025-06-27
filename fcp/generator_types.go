package fcp

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"math"

	"os"
	"os/exec"
	"path/filepath"

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

// ConversationPattern holds information about the current conversation state
type ConversationPattern struct {
	NextBubbleType string // "blue" or "white"
	LastText       string // Text from the last bubble
	VideoCount     int    // Number of video segments
	NextOffset     string // Where to place the next segment
	NextDuration   string // Duration for next segment
}

// AssetCollection holds categorized assets
type AssetCollection struct {
	Images []string
	Videos []string
}

func oldGgenerateUID(videoID string) string {

	hasher := md5.New()
	hasher.Write([]byte("cutlass_video_" + videoID))
	hash := hasher.Sum(nil)

	return strings.ToUpper(hex.EncodeToString(hash))
}

// generateBookmark creates a macOS security bookmark for a file path using Swift
func generateBookmark(filePath string) (string, error) {

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", absPath)
	}

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

	cmd := exec.Command("swift", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {

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

	framesPerSecond := 24000.0 / 1001.0
	exactFrames := seconds * framesPerSecond

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

	return fmt.Sprintf("%d/24000s", frames*1001)
}

// GenerateEmpty creates an empty FCPXML file structure and returns a pointer to it
func GenerateEmpty(filename string) (*FCPXML, error) {
	return GenerateEmptyWithFormat(filename, "horizontal")
}

// GenerateEmptyWithFormat creates an empty FCPXML file structure with specified format
func GenerateEmptyWithFormat(filename string, format string) (*FCPXML, error) {
	var formatConfig Format
	
	switch format {
	case "vertical":
		formatConfig = Format{
			ID:            "r1",
			Name:          "FFVideoFormat1080p2398_Vertical",
			FrameDuration: "1001/24000s",
			Width:         "1080",
			Height:        "1920",
			ColorSpace:    "1-1-1 (Rec. 709)",
		}
	case "horizontal":
		fallthrough
	default:
		formatConfig = Format{
			ID:            "r1",
			Name:          "FFVideoFormat720p2398",
			FrameDuration: "1001/24000s",
			Width:         "1280",
			Height:        "720",
			ColorSpace:    "1-1-1 (Rec. 709)",
		}
	}

	fcpxml := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Formats: []Format{formatConfig},
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
// WriteToFile writes FCPXML to file using the new validation-first architecture
func WriteToFile(fcpxml *FCPXML, filename string) error {
	// Use the validation-first marshaling from Step 17
	output, err := fcpxml.ValidateAndMarshal()
	if err != nil {
		return fmt.Errorf("validation and marshaling failed: %v", err)
	}

	xmlHeader := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE fcpxml>

`

	fullXML := xmlHeader + string(output)

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

	registry := NewResourceRegistry(fcpxml)

	if asset, exists := registry.GetOrCreateAsset(videoPath); exists {

		return addAssetClipToSpine(fcpxml, asset, 10.0)
	}

	tx := NewTransaction(registry)

	absPath, err := filepath.Abs(videoPath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		tx.Rollback()
		return fmt.Errorf("video file does not exist: %s", absPath)
	}

	ids := tx.ReserveIDs(1)
	assetID := ids[0]

	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	defaultDurationSeconds := 10.0
	frameDuration := ConvertSecondsToFCPDuration(defaultDurationSeconds)

	err = tx.CreateVideoAssetWithDetection(assetID, absPath, videoName, frameDuration, "r1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create video asset with detection: %v", err)
	}

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

	return addAssetClipToSpine(fcpxml, asset, defaultDurationSeconds)
}

// addAssetClipToSpine adds an asset-clip to the sequence spine
func addAssetClipToSpine(fcpxml *FCPXML, asset *Asset, durationSeconds float64) error {

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		currentTimelineDuration := calculateTimelineDuration(sequence)

		clipDuration := ConvertSecondsToFCPDuration(durationSeconds)

		assetClip := AssetClip{
			Ref:       asset.ID,
			Offset:    currentTimelineDuration,
			Name:      asset.Name,
			Duration:  clipDuration,
			Format:    asset.Format,
			TCFormat:  "NDF",
			AudioRole: "dialogue",
		}

		sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, assetClip)

		newTimelineDuration := addDurations(currentTimelineDuration, clipDuration)
		sequence.Duration = newTimelineDuration
	}

	return nil
}
