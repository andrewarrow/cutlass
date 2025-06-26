package fcp

import (
	"fmt"

	"os"

	"path/filepath"

	"strings"
)

// getAllExistingTextStyleIDs collects all existing text style IDs from the entire FCPXML
func getAllExistingTextStyleIDs(fcpxml *FCPXML) map[string]bool {
	existingIDs := make(map[string]bool)

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		for _, title := range sequence.Spine.Titles {
			for _, styleDef := range title.TextStyleDefs {
				existingIDs[styleDef.ID] = true
			}
		}

		for _, assetClip := range sequence.Spine.AssetClips {
			for _, title := range assetClip.Titles {
				for _, styleDef := range title.TextStyleDefs {
					existingIDs[styleDef.ID] = true
				}
			}
		}

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

	counter := 1
	for {
		candidateID := fmt.Sprintf("ts%d", counter)
		if !existingIDs[candidateID] {

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

	if len(fcpxml.Library.Events) == 0 || len(fcpxml.Library.Events[0].Projects) == 0 || len(fcpxml.Library.Events[0].Projects[0].Sequences) == 0 {
		return fmt.Errorf("no sequence found in FCPXML")
	}

	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	offsetFrames := parseFCPDuration(ConvertSecondsToFCPDuration(offsetSeconds))

	// Find the video at the specified offset
	var targetVideo *Video = nil

	for i := range sequence.Spine.Videos {
		video := &sequence.Spine.Videos[i]
		videoOffsetFrames := parseFCPDuration(video.Offset)
		videoDurationFrames := parseFCPDuration(video.Duration)
		videoEndFrames := videoOffsetFrames + videoDurationFrames

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

			if offsetFrames >= clipOffsetFrames && offsetFrames < clipEndFrames {

				targetClip = &sequence.Spine.AssetClips[i]
				break
			}
		}
	}

	if targetVideo == nil && targetClip == nil {
		return fmt.Errorf("no video found at offset %.1f seconds", offsetSeconds)
	}

	if targetVideo != nil {

		if targetVideo.AdjustTransform != nil {

			for _, param := range targetVideo.AdjustTransform.Params {
				if param.Name == "position" && param.KeyframeAnimation != nil {
					return fmt.Errorf("video '%s' at offset %.1f seconds already has slide animation", targetVideo.Name, offsetSeconds)
				}
			}
		}

		videoStartFrames := parseFCPDuration(targetVideo.Start)
		if videoStartFrames == 0 {

			videoStartFrames = 86399313
			targetVideo.Start = "86399313/24000s"
		}

		targetVideo.AdjustTransform = createSlideAnimation(targetVideo.Offset, 1.0)
	}

	if targetClip != nil {

		if targetClip.AdjustTransform != nil {

			for _, param := range targetClip.AdjustTransform.Params {
				if param.Name == "position" && param.KeyframeAnimation != nil {
					return fmt.Errorf("video '%s' at offset %.1f seconds already has slide animation", targetClip.Name, offsetSeconds)
				}
			}
		}

		targetClip.AdjustTransform = createAssetClipSlideAnimation(targetClip.Offset, 1.0)
	}

	return nil
}

// createAssetClipSlideAnimation creates timeline-based slide animation for AssetClip elements (videos)
// 🚨 CRITICAL: Keyframe attributes follow CLAUDE.md rules:
// - Position keyframes: NO attributes (no interp/curve)
// - Scale/Rotation/Anchor keyframes: Only curve attribute (no interp)
func createAssetClipSlideAnimation(clipOffset string, totalDurationSeconds float64) *AdjustTransform {

	offsetFrames := parseFCPDuration(clipOffset)

	oneSecondDuration := ConvertSecondsToFCPDuration(1.0)
	oneSecondFrames := parseFCPDuration(oneSecondDuration)

	startTime := fmt.Sprintf("%d/24000s", offsetFrames)
	endTime := fmt.Sprintf("%d/24000s", offsetFrames+oneSecondFrames)

	return &AdjustTransform{
		Params: []Param{
			{
				Name: "anchor",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  endTime,
							Value: "0 0",
							Curve: "linear",
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
							Curve: "linear",
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
							Curve: "linear",
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

	if !isAudioFile(audioPath) {
		return fmt.Errorf("file is not a supported audio format (WAV, MP3, M4A, AAC, FLAC): %s", audioPath)
	}

	registry := NewResourceRegistry(fcpxml)

	if asset, exists := registry.GetOrCreateAsset(audioPath); exists {

		return addAudioAssetClipToSpine(fcpxml, asset)
	}

	tx := NewTransaction(registry)

	absPath, err := filepath.Abs(audioPath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		tx.Rollback()
		return fmt.Errorf("audio file does not exist: %s", absPath)
	}

	ids := tx.ReserveIDs(1)
	assetID := ids[0]

	audioName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))

	defaultDurationSeconds := 60.0
	frameDuration := ConvertSecondsToFCPDuration(defaultDurationSeconds)

	asset, err := tx.CreateAsset(assetID, absPath, audioName, frameDuration, "r1")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create audio asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return addAudioAssetClipToSpine(fcpxml, asset)
}

// addAudioAssetClipToSpine adds an audio asset-clip nested inside the first video element
// 🚨 CRITICAL FIX: Audio must be nested inside video elements, not as separate spine elements
// Analysis of Info.fcpxml shows audio is nested: <video><asset-clip lane="-1"/></video>
func addAudioAssetClipToSpine(fcpxml *FCPXML, asset *Asset) error {

	if len(fcpxml.Library.Events) > 0 && len(fcpxml.Library.Events[0].Projects) > 0 && len(fcpxml.Library.Events[0].Projects[0].Sequences) > 0 {
		sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

		// Find the first video element in the spine to nest audio inside
		var targetVideo *Video = nil
		for i := range sequence.Spine.Videos {
			targetVideo = &sequence.Spine.Videos[i]
			break
		}

		if targetVideo == nil && len(sequence.Spine.AssetClips) > 0 {
			clip := &sequence.Spine.AssetClips[0]
			video := Video{
				Ref:      clip.Ref,
				Offset:   clip.Offset,
				Name:     clip.Name,
				Duration: clip.Duration,
				Start:    clip.Start,
			}

			sequence.Spine.AssetClips = sequence.Spine.AssetClips[1:]
			sequence.Spine.Videos = append(sequence.Spine.Videos, video)
			targetVideo = &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]
		}

		if targetVideo == nil {
			return fmt.Errorf("no video element found to nest audio inside - audio must be nested within a video element")
		}

		audioOffset := "28799771/8000s"

		audioDuration := asset.Duration

		assetClip := AssetClip{
			Ref:       asset.ID,
			Lane:      "-1",
			Offset:    audioOffset,
			Name:      asset.Name,
			Duration:  audioDuration,
			Format:    asset.Format,
			TCFormat:  "NDF",
			AudioRole: "dialogue",
		}

		targetVideo.NestedAssetClips = append(targetVideo.NestedAssetClips, assetClip)

		currentSequenceDurationFrames := parseFCPDuration(sequence.Duration)
		audioDurationFrames := parseFCPDuration(audioDuration)

		if audioDurationFrames > currentSequenceDurationFrames {
			sequence.Duration = audioDuration
		}
	}

	return nil
}
