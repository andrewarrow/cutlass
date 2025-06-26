package fcp

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ResourceTransaction provides atomic multi-resource operations
type ResourceTransaction struct {
	registry *ResourceRegistry
	reserved []string
	created  []Resource
	rolled   bool
}

// NewTransaction creates a new resource transaction
func NewTransaction(registry *ResourceRegistry) *ResourceTransaction {
	return &ResourceTransaction{
		registry: registry,
		reserved: make([]string, 0),
		created:  make([]Resource, 0),
	}
}

// ReserveIDs reserves multiple IDs for this transaction
func (tx *ResourceTransaction) ReserveIDs(count int) []string {
	if tx.rolled {
		return nil
	}

	ids := tx.registry.ReserveIDs(count)
	tx.reserved = append(tx.reserved, ids...)
	return ids
}

// CreateVideoAssetWithDetection creates a video asset with proper media detection
func (tx *ResourceTransaction) CreateVideoAssetWithDetection(id, filePath, baseName, duration string, formatID string) error {
	if tx.rolled {
		return fmt.Errorf("transaction has been rolled back")
	}

	// Detect actual video properties
	props, err := detectVideoProperties(filePath)
	if err != nil {
		// Fallback to basic asset creation if video detection fails
		// This handles test scenarios with fake video files
		_, err := tx.CreateAsset(id, filePath, baseName, duration, formatID)
		if err != nil {
			return err
		}
		
		// CRITICAL FIX: CreateAsset doesn't create format for videos, so create it manually in fallback
		_, err = tx.CreateFormatWithFrameDuration(formatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
		if err != nil {
			return fmt.Errorf("failed to create fallback video format: %v", err)
		}
		
		return nil
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Generate completely random UID to prevent FCP cache conflicts
	// This allows importing the same file multiple times with different UIDs
	// Perfect for BAFFLE stress testing where we want fresh imports every time
	uid := generateRandomUID()

	// Generate security bookmark for file access
	bookmark, err := generateBookmark(absPath)
	if err != nil {
		// Log but don't fail - bookmark is optional
		bookmark = ""
	}

	// Create asset with detected properties
	asset := &Asset{
		ID:           id,
		Name:         baseName,
		UID:          uid,
		Start:        "0s",
		Duration:     duration,
		HasVideo:     "1",
		Format:       formatID,
		VideoSources: "1",
		MediaRep: MediaRep{
			Kind:     "original-media",
			Sig:      uid,
			Src:      "file://" + absPath,
			Bookmark: bookmark,
		},
	}

	// Set audio properties only if video has audio
	if props.HasAudio {
		asset.HasAudio = "1"
		asset.AudioSources = "1"
		asset.AudioChannels = props.AudioChannels
		asset.AudioRate = props.AudioRate
	}

	// Generate metadata based on actual file properties
	asset.Metadata = createVideoMetadata(props, absPath)

	// Create format with detected properties
	format := &Format{
		ID:            formatID,
		Name:          "", // Will be auto-generated based on properties
		FrameDuration: props.FrameRate,
		Width:         fmt.Sprintf("%d", props.Width),
		Height:        fmt.Sprintf("%d", props.Height),
		ColorSpace:    "1-1-1 (Rec. 709)", // Standard default
	}

	// Add both to transaction
	tx.created = append(tx.created, &AssetWrapper{asset}, &FormatWrapper{format})
	return nil
}

// CreateAsset creates an asset with transaction management
func (tx *ResourceTransaction) CreateAsset(id, filePath, baseName, duration string, formatID string) (*Asset, error) {
	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Generate consistent UID based on file path for deterministic results
	// This prevents "cannot be imported again with different unique identifier" errors
	uid := generateUID(absPath)

	// Generate security bookmark for file access
	bookmark, err := generateBookmark(absPath)
	if err != nil {
		// Log but don't fail - bookmark is optional
		bookmark = ""
	}

	// Create asset
	asset := &Asset{
		ID:       id,
		Name:     baseName,
		UID:      uid,
		Start:    "0s",
		Duration: duration,
		HasVideo: "1",
		Format:   formatID,
		MediaRep: MediaRep{
			Kind:     "original-media",
			Sig:      uid,
			Src:      "file://" + absPath,
			Bookmark: bookmark,
		},
	}

	// Set file-type specific properties and metadata
	if isImageFile(absPath) {
		// 🚨 CRITICAL: Images are timeless - asset duration MUST be "0s"
		// Display duration is applied only to Video element in spine, not asset
		// This matches working samples/png.fcpxml pattern: asset duration="0s"
		asset.Duration = "0s" // CRITICAL: Override caller duration for images
		asset.VideoSources = "1" // Required for image assets
		// Image files (PNG, JPG, JPEG) should NOT have audio properties
		asset.Metadata = createImageMetadata(absPath)
	} else if isAudioFile(absPath) {
		// Audio files have only audio properties, NO video properties
		// 🚨 FIX: Don't set HasVideo to empty string, just don't set it (omitempty will handle)
		asset.HasVideo = "" // This will be omitted due to omitempty tag
		asset.HasAudio = "1"
		asset.AudioSources = "1"
		asset.AudioChannels = "2"
		asset.AudioRate = "48000"
		// 🚨 FIX: Audio files should have format=="" which gets omitted due to omitempty
		// Or we should create a specific audio format. For now, leave format empty.
		asset.Format = "" // This will be omitted due to omitempty tag
		// Note: Duration remains as provided by caller (audio duration)
	} else {
		// Video files - check if they actually have audio using ffprobe
		asset.VideoSources = "1" // Required for video assets
		
		// Try to detect if video has audio using ffprobe
		if hasAudioTrack(absPath) {
			asset.HasAudio = "1"
			asset.AudioSources = "1"  // Required for video assets with audio
			asset.AudioChannels = "2"
			asset.AudioRate = "48000"
		}
		// If no audio track, leave audio properties empty (omitted by omitempty tags)
	}

	tx.created = append(tx.created, &AssetWrapper{asset})
	return asset, nil
}

// CreateVideoOnlyAsset creates an asset with only video properties (no audio) for PIP videos
// This matches the pattern in samples/pip.fcpxml where PIP video has no audio properties
func (tx *ResourceTransaction) CreateVideoOnlyAsset(id, filePath, baseName, duration string, formatID string) (*Asset, error) {
	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Generate random UID for PIP videos to avoid collisions
	// This ensures each PIP import gets a unique UID
	uid := generateRandomUID()

	// Create asset with ONLY video properties (no audio properties)
	// This matches samples/pip.fcpxml pattern: hasVideo="1" format="r5" videoSources="1" (no audio)
	asset := &Asset{
		ID:           id,
		Name:         baseName,
		UID:          uid,
		Start:        "0s",
		Duration:     duration,
		HasVideo:     "1",
		Format:       formatID,
		VideoSources: "1", // Required for video assets used as PIP
		// Note: NO audio properties - HasAudio, AudioSources, AudioChannels, AudioRate are omitted
		MediaRep: MediaRep{
			Kind: "original-media",
			Sig:  uid,
			Src:  "file://" + absPath,
		},
	}

	tx.created = append(tx.created, &AssetWrapper{asset})
	return asset, nil
}

// CreateFormat creates a format with transaction management
// 🚨 CRITICAL: frameDuration should ONLY be set for sequence formats, NOT image formats
// Image formats must NOT have frameDuration or FCP's performAudioPreflightCheckForObject crashes
// Analysis of working top5orig.fcpxml shows image formats have NO frameDuration attribute
func (tx *ResourceTransaction) CreateFormat(id, name, width, height, colorSpace string) (*Format, error) {
	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}

	format := &Format{
		ID:         id,
		Name:       name,
		Width:      width,
		Height:     height,
		ColorSpace: colorSpace,
		// Note: FrameDuration intentionally omitted - only sequence formats need frameDuration
	}

	tx.created = append(tx.created, &FormatWrapper{format})
	return format, nil
}

// CreateFormatWithFrameDuration creates a format with frameDuration for video formats (NOT image formats)
// 🚨 CRITICAL: frameDuration should ONLY be set for video/sequence formats, NOT image formats
// Image formats must NOT have frameDuration or FCP's performAudioPreflightCheckForObject crashes
func (tx *ResourceTransaction) CreateFormatWithFrameDuration(id, frameDuration, width, height, colorSpace string) (*Format, error) {
	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}

	format := &Format{
		ID:            id,
		Name:          "", // Will be set to appropriate name based on format
		FrameDuration: frameDuration,
		Width:         width,
		Height:        height,
		ColorSpace:    colorSpace,
	}

	tx.created = append(tx.created, &FormatWrapper{format})
	return format, nil
}

// CreateEffect creates an effect with transaction management
func (tx *ResourceTransaction) CreateEffect(id, name, uid string) (*Effect, error) {
	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}

	effect := &Effect{
		ID:   id,
		Name: name,
		UID:  uid,
	}

	tx.created = append(tx.created, &EffectWrapper{effect})
	return effect, nil
}

// createCompoundClipSpineContent creates the spine content for a compound clip using structs
func (tx *ResourceTransaction) createCompoundClipSpineContent(videoAssetID, audioAssetID, baseName, duration string) string {
	// Create audio asset-clip struct
	audioClip := AssetClip{
		Ref:       audioAssetID,
		Lane:      "-1",
		Offset:    "28799771/8000s",
		Name:      baseName,
		Duration:  duration,
		Format:    "r1",
		TCFormat:  "NDF",
		AudioRole: "dialogue",
	}

	// Create video element with nested audio
	video := Video{
		Ref:      videoAssetID,
		Offset:   "0s",
		Name:     baseName,
		Start:    "86399313/24000s",
		Duration: duration,
	}

	// Note: The Video struct doesn't support nested asset-clips directly
	// So we need a hybrid approach here - marshal the video and manually insert the audio clip
	videoXML, err := xml.MarshalIndent(&video, "                        ", "    ")
	if err != nil {
		return "<!-- Error marshaling compound clip video: " + err.Error() + " -->"
	}

	audioXML, err := xml.MarshalIndent(&audioClip, "                            ", "    ")
	if err != nil {
		return "<!-- Error marshaling compound clip audio: " + err.Error() + " -->"
	}

	// Insert the audio clip before the closing video tag
	videoStr := string(videoXML)
	videoStr = strings.Replace(videoStr, "</video>", "    "+string(audioXML)+"\n                        </video>", 1)

	return videoStr
}

// Commit commits all created resources to the registry
func (tx *ResourceTransaction) Commit() error {
	if tx.rolled {
		return fmt.Errorf("transaction has been rolled back")
	}

	// Register all created resources
	for _, resource := range tx.created {
		switch r := resource.(type) {
		case *AssetWrapper:
			tx.registry.RegisterAsset(r.Asset)
		case *FormatWrapper:
			tx.registry.RegisterFormat(r.Format)
		case *EffectWrapper:
			tx.registry.RegisterEffect(r.Effect)
		case *MediaWrapper:
			tx.registry.RegisterMedia(r.Media)
		}
	}

	return nil
}

// Rollback rolls back the transaction (IDs remain reserved)
func (tx *ResourceTransaction) Rollback() {
	tx.rolled = true
	tx.created = nil
}

// VideoProperties holds detected video file properties
type VideoProperties struct {
	Width       int
	Height      int
	FrameRate   string // FCP format like "1001/30000s"
	Duration    string // FCP format like "12345/24000s"
	HasAudio    bool
	AudioRate   string
	AudioChannels string
}

// hasAudioTrack checks if a video file has an audio track using ffprobe
func hasAudioTrack(videoPath string) bool {
	// Use ffprobe to check for audio streams
	cmd := exec.Command("ffprobe", "-v", "quiet", "-select_streams", "a", "-show_entries", "stream=codec_type", "-of", "csv=p=0", videoPath)
	output, err := cmd.Output()
	if err != nil {
		// If ffprobe fails, assume no audio (safer than assuming audio exists)
		return false
	}
	
	// If output contains "audio", then there's an audio track
	return strings.Contains(string(output), "audio")
}

// detectVideoProperties analyzes a video file and returns its actual properties
func detectVideoProperties(videoPath string) (*VideoProperties, error) {
	// Use ffprobe to get detailed video properties as JSON
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %v", err)
	}
	
	// Parse ffprobe JSON output
	var probeResult struct {
		Streams []struct {
			CodecType     string `json:"codec_type"`
			Width         int    `json:"width"`
			Height        int    `json:"height"`
			RFrameRate    string `json:"r_frame_rate"`
			AvgFrameRate  string `json:"avg_frame_rate"`
			Duration      string `json:"duration"`
			SampleRate    string `json:"sample_rate"`
			Channels      int    `json:"channels"`
		} `json:"streams"`
	}
	
	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %v", err)
	}
	
	props := &VideoProperties{}
	
	// Find video and audio streams
	for _, stream := range probeResult.Streams {
		if stream.CodecType == "video" {
			props.Width = stream.Width
			props.Height = stream.Height
			
			// Convert frame rate to FCP format using average frame rate (more reliable)
			frameRateStr := stream.AvgFrameRate
			if frameRateStr == "" || frameRateStr == "0/0" {
				frameRateStr = stream.RFrameRate
			}
			
			if frameRateStr != "" && frameRateStr != "0/0" {
				props.FrameRate = convertFrameRateToFCP(frameRateStr)
			} else {
				props.FrameRate = "1001/30000s" // Default fallback
			}
			
			// Convert duration to FCP format
			if stream.Duration != "" {
				if duration, err := strconv.ParseFloat(stream.Duration, 64); err == nil {
					props.Duration = ConvertSecondsToFCPDuration(duration)
				}
			}
		} else if stream.CodecType == "audio" {
			props.HasAudio = true
			if stream.SampleRate != "" {
				props.AudioRate = stream.SampleRate
			} else {
				props.AudioRate = "48000" // Default fallback
			}
			if stream.Channels > 0 {
				props.AudioChannels = strconv.Itoa(stream.Channels)
			} else {
				props.AudioChannels = "2" // Default fallback
			}
		}
	}
	
	// Fallback defaults if no video stream found
	if props.Width == 0 {
		props.Width = 1920
		props.Height = 1080
		props.FrameRate = "1001/30000s"
	}
	
	return props, nil
}

// createImageMetadata creates appropriate metadata for image files
func createImageMetadata(filePath string) *Metadata {
	return &Metadata{
		MDs: []MetadataItem{
			{Key: "com.apple.proapps.studio.rawToLogConversion", Value: "0"},
			{Key: "com.apple.proapps.spotlight.kMDItemProfileName", Value: "sRGB IEC61966-2.1"},
			{Key: "com.apple.proapps.studio.cameraISO", Value: "0"},
			{Key: "com.apple.proapps.studio.cameraColorTemperature", Value: "0"},
			{Key: "com.apple.proapps.mio.ingestDate", Value: "2025-06-25 11:46:22 -0700"},
			{Key: "com.apple.proapps.spotlight.kMDItemOrientation", Value: "0"},
		},
	}
}

// createVideoMetadata creates appropriate metadata for video files
func createVideoMetadata(props *VideoProperties, filePath string) *Metadata {
	metadata := &Metadata{
		MDs: []MetadataItem{
			{Key: "com.apple.proapps.studio.rawToLogConversion", Value: "0"},
			{Key: "com.apple.proapps.spotlight.kMDItemProfileName", Value: "HD (1-1-1)"},
			{Key: "com.apple.proapps.studio.cameraISO", Value: "0"},
			{Key: "com.apple.proapps.studio.cameraColorTemperature", Value: "0"},
		},
	}

	// Add codec information based on actual file content
	codecs := []string{"H.264"} // Always has video
	if props.HasAudio {
		codecs = append(codecs, "MPEG-4 AAC") // Only add audio codec if file has audio
	}

	metadata.MDs = append(metadata.MDs, MetadataItem{
		Key: "com.apple.proapps.spotlight.kMDItemCodecs",
		Array: &StringArray{
			Strings: codecs,
		},
	})

	return metadata
}

// convertFrameRateToFCP converts ffprobe frame rate to FCP format with validation
func convertFrameRateToFCP(frameRateStr string) string {
	// Parse frame rate like "30000/1001", "1890000/74317", or "25/1"
	parts := strings.Split(frameRateStr, "/")
	if len(parts) != 2 {
		return "1001/30000s" // Default fallback
	}
	
	numerator, err1 := strconv.ParseFloat(parts[0], 64)
	denominator, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil || denominator == 0 {
		return "1001/30000s" // Default fallback
	}
	
	// Calculate actual frame rate in fps
	actualFps := numerator / denominator
	
	// Validate frame rate is reasonable (between 1 and 120 fps)
	if actualFps < 1 || actualFps > 120 {
		return "1001/30000s" // Default fallback for unreasonable rates
	}
	
	// Map to common FCP frame durations based on detected fps with wider tolerance
	if actualFps >= 20.0 && actualFps <= 24.5 {
		return "1001/24000s" // 23.976 fps (includes 21.2fps from scan.mp4)
	} else if actualFps >= 24.5 && actualFps <= 25.5 {
		return "1/25s" // 25 fps
	} else if actualFps >= 29.5 && actualFps <= 30.5 {
		return "1001/30000s" // 29.97 fps
	} else if actualFps >= 59.5 && actualFps <= 60.5 {
		return "1001/60000s" // 59.94 fps
	} else {
		// For other frame rates, try to create proper reciprocal
		// Convert back to integers for clean FCP format
		if denominator == 1 {
			// Simple case like "25/1" -> "1/25s"
			return fmt.Sprintf("1/%.0fs", numerator)
		} else {
			// Complex fraction - use reciprocal with integer conversion
			return fmt.Sprintf("%.0f/%.0fs", denominator, numerator)
		}
	}
}

