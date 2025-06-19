package fcp

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestComplexTimeline tests sophisticated timeline structures with multiple
// tracks, gaps, overlaps, and synchronized elements to validate the FCP
// package can handle professional editing workflows.
func TestComplexTimeline(t *testing.T) {
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	// Reserve IDs for multiple assets, formats, and effects
	ids := tx.ReserveIDs(15)
	video1ID := ids[0]
	video2ID := ids[1]
	video3ID := ids[2]
	audio1ID := ids[3]
	audio2ID := ids[4]
	image1ID := ids[5]
	image2ID := ids[6]
	format1ID := ids[7]
	format2ID := ids[8]
	format3ID := ids[9]
	audioFormatID := ids[10]
	imageFormatID := ids[11]
	transitionID := ids[12]
	titleEffectID := ids[13]
	audioEffectID := ids[14]

	// Create different formats for various media types
	formats := []struct {
		id           string
		frameDuration string
		width        string
		height       string
		name         string
	}{
		{format1ID, "1001/24000s", "1920", "1080", "FFVideoFormat1080p2398"},
		{format2ID, "1001/30000s", "1280", "720", "FFVideoFormat720p30"},
		{format3ID, "1001/25000s", "3840", "2160", "FFVideoFormat4K25"},
		{audioFormatID, "", "", "", ""}, // Audio format has no video dimensions
		{imageFormatID, "", "1920", "1080", "FFVideoFormatRateUndefined"}, // Image format
	}

	for _, format := range formats {
		if format.frameDuration != "" {
			_, err = tx.CreateFormatWithFrameDuration(format.id, format.frameDuration, format.width, format.height, "1-1-1 (Rec. 709)")
		} else if format.width != "" {
			_, err = tx.CreateFormat(format.id, format.name, format.width, format.height, "1-13-1")
		} else {
			_, err = tx.CreateFormat(format.id, "FFAudioFormat48k", "", "", "")
		}
		if err != nil {
			t.Fatalf("Failed to create format %s: %v", format.id, err)
		}
	}

	// Create transition and effects
	_, err = tx.CreateEffect(transitionID, "Cross Dissolve", "FFCrossDissolve")
	if err != nil {
		t.Fatalf("Failed to create transition effect: %v", err)
	}

	_, err = tx.CreateEffect(titleEffectID, "Title", ".../Titles.localized/Basic Text.localized/Title.localized/Title.moti")
	if err != nil {
		t.Fatalf("Failed to create title effect: %v", err)
	}

	_, err = tx.CreateEffect(audioEffectID, "DeEsser", "FFDeEsser")
	if err != nil {
		t.Fatalf("Failed to create audio effect: %v", err)
	}

	// Create video assets with different durations
	video1Duration := ConvertSecondsToFCPDuration(12.0)
	video2Duration := ConvertSecondsToFCPDuration(8.0)
	video3Duration := ConvertSecondsToFCPDuration(15.0)
	audio1Duration := ConvertSecondsToFCPDuration(20.0)
	audio2Duration := ConvertSecondsToFCPDuration(10.0)
	image1Duration := "0s" // Images are timeless
	image2Duration := "0s"

	video1Asset, err := tx.CreateAsset(video1ID, "/Users/aa/cs/cutlass/assets/long.mov", "MainVideo", video1Duration, format1ID)
	if err != nil {
		t.Fatalf("Failed to create video1 asset: %v", err)
	}

	video2Asset, err := tx.CreateAsset(video2ID, "/Users/aa/cs/cutlass/assets/speech1.mov", "CutawayVideo", video2Duration, format2ID)
	if err != nil {
		t.Fatalf("Failed to create video2 asset: %v", err)
	}

	video3Asset, err := tx.CreateAsset(video3ID, "/Users/aa/cs/cutlass/assets/speech2.mov", "BRollVideo", video3Duration, format3ID)
	if err != nil {
		t.Fatalf("Failed to create video3 asset: %v", err)
	}

	// Create audio assets (will set audio-specific properties)
	audio1Asset, err := tx.CreateAsset(audio1ID, "/Users/aa/cs/cutlass/assets/Ethereal Accents.caf", "Music", audio1Duration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create audio1 asset: %v", err)
	}

	audio2Asset, err := tx.CreateAsset(audio2ID, "/Users/aa/cs/cutlass/assets/Synth Zap Accent 06.caf", "SoundFX", audio2Duration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create audio2 asset: %v", err)
	}

	// Create image assets
	image1Asset, err := tx.CreateAsset(image1ID, "/Users/aa/cs/cutlass/assets/cutlass_logo_t.png", "Logo", image1Duration, imageFormatID)
	if err != nil {
		t.Fatalf("Failed to create image1 asset: %v", err)
	}

	image2Asset, err := tx.CreateAsset(image2ID, "/Users/aa/cs/cutlass/assets/waymo.png", "Overlay", image2Duration, imageFormatID)
	if err != nil {
		t.Fatalf("Failed to create image2 asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build complex timeline structure
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Calculate timing offsets for complex timeline
	gap1Start := "0s"
	gap1Duration := ConvertSecondsToFCPDuration(2.0)
	
	video1Start := ConvertSecondsToFCPDuration(2.0)
	
	gap2Start := ConvertSecondsToFCPDuration(14.0) // After video1 (2+12)
	gap2Duration := ConvertSecondsToFCPDuration(1.0)
	
	video2Start := ConvertSecondsToFCPDuration(15.0) // After gap2
	
	video3Start := ConvertSecondsToFCPDuration(20.0) // Overlap with video2 (15+8=23, but start at 20)
	
	// Audio tracks start earlier and run longer
	musicStart := "0s"
	sfxStart := ConvertSecondsToFCPDuration(5.0)

	// Image overlays at specific times
	logoStart := ConvertSecondsToFCPDuration(3.0)
	logoDisplayDuration := ConvertSecondsToFCPDuration(4.0)
	
	overlayStart := ConvertSecondsToFCPDuration(18.0)
	overlayDisplayDuration := ConvertSecondsToFCPDuration(6.0)

	// Primary timeline structure with gaps and overlaps
	sequence.Spine = Spine{
		// Start with gap (black/silence)
		Gaps: []Gap{
			{
				Name:     "Gap",
				Offset:   gap1Start,
				Duration: gap1Duration,
			},
			{
				Name:     "Transition Gap",
				Offset:   gap2Start,
				Duration: gap2Duration,
			},
		},
		// Main video timeline
		AssetClips: []AssetClip{
			// Main video with nested audio and overlays
			{
				Ref:       video1Asset.ID,
				Offset:    video1Start,
				Name:      video1Asset.Name,
				Duration:  video1Duration,
				Format:    video1Asset.Format,
				TCFormat:  "NDF",
				AudioRole: "dialogue",
				ConformRate: &ConformRate{
					ScaleEnabled: "0",
				},
				// Nested audio tracks
				NestedAssetClips: []AssetClip{
					// Background music (lane -1)
					{
						Ref:       audio1Asset.ID,
						Lane:      "-1",
						Offset:    musicStart,
						Name:      audio1Asset.Name,
						Duration:  audio1Duration,
						Format:    audio1Asset.Format,
						AudioRole: "music",
						FilterVideos: []FilterVideo{
							{
								Ref:  audioEffectID,
								Name: "DeEsser",
								Params: []Param{
									{
										Name:  "Threshold",
										Key:   "1",
										Value: "-20",
									},
									{
										Name:  "Frequency",
										Key:   "2",
										Value: "6000",
									},
								},
							},
						},
					},
					// Sound effects (lane -2)
					{
						Ref:       audio2Asset.ID,
						Lane:      "-2",
						Offset:    sfxStart,
						Name:      audio2Asset.Name,
						Duration:  audio2Duration,
						Format:    audio2Asset.Format,
						AudioRole: "effects",
					},
				},
				// Titles overlaid on main video
				Titles: []Title{
					{
						Ref:      titleEffectID,
						Lane:     "1",
						Offset:   ConvertSecondsToFCPDuration(4.0),
						Name:     "Main Title",
						Duration: ConvertSecondsToFCPDuration(3.0),
						Start:    "86486400/24000s",
						Params: []Param{
							{
								Name:  "Text",
								Key:   "9999/10003/13260/3296672360/2/354",
								Value: "Professional Timeline Test",
							},
							{
								Name:  "Font Size",
								Key:   "9999/10003/13260/3296672360/2/354/3296667315/402",
								Value: "72",
							},
						},
						Text: &TitleText{
							TextStyle: TextStyleRef{
								Ref:  GenerateTextStyleID("Professional Timeline Test", "main_title"),
								Text: "Professional Timeline Test",
							},
						},
						TextStyleDef: &TextStyleDef{
							ID: GenerateTextStyleID("Professional Timeline Test", "main_title"),
							TextStyle: TextStyle{
								Font:      "Helvetica Neue",
								FontSize:  "72",
								FontColor: "1 1 1 1",
								Bold:      "1",
								Alignment: "center",
							},
						},
					},
				},
			},
			// Cutaway video (overlapping with main video end)
			{
				Ref:      video2Asset.ID,
				Offset:   video2Start,
				Name:     video2Asset.Name,
				Duration: video2Duration,
				Format:   video2Asset.Format,
				TCFormat: "NDF",
				ConformRate: &ConformRate{
					ScaleEnabled: "0",
				},
				AdjustTransform: &AdjustTransform{
					Position: "300 200", // Position as PIP overlay
					Scale:    "0.4 0.4", // Scale down for overlay
				},
			},
			// B-roll video (overlapping with cutaway)
			{
				Ref:      video3Asset.ID,
				Lane:     "1", // Upper video layer
				Offset:   video3Start,
				Name:     video3Asset.Name,
				Duration: video3Duration,
				Format:   video3Asset.Format,
				TCFormat: "NDF",
				ConformRate: &ConformRate{
					ScaleEnabled: "0",
				},
				AdjustTransform: &AdjustTransform{
					Position: "-400 -300", // Different position
					Scale:    "0.3 0.3",   // Smaller scale
				},
			},
		},
		// Image overlays as Video elements
		Videos: []Video{
			// Logo overlay
			{
				Ref:      image1Asset.ID,
				Offset:   logoStart,
				Name:     image1Asset.Name,
				Duration: logoDisplayDuration,
				Start:    "86399313/24000s",
				AdjustTransform: &AdjustTransform{
					Position: "-600 -400", // Top left corner
					Scale:    "0.2 0.2",   // Small logo
				},
			},
			// Waymo overlay
			{
				Ref:      image2Asset.ID,
				Lane:     "2", // Higher layer
				Offset:   overlayStart,
				Name:     image2Asset.Name,
				Duration: overlayDisplayDuration,
				Start:    "86399313/24000s",
				AdjustTransform: &AdjustTransform{
					Position: "500 350", // Bottom right
					Scale:    "0.15 0.15", // Very small overlay
				},
			},
		},
	}

	// Calculate total timeline duration (longest element)
	totalTimelineFrames := 0
	
	// Check all elements to find the longest
	elements := []struct {
		offset   string
		duration string
	}{
		{gap1Start, gap1Duration},
		{gap2Start, gap2Duration},
		{video1Start, video1Duration},
		{video2Start, video2Duration},
		{video3Start, video3Duration},
		{musicStart, audio1Duration},
		{sfxStart, audio2Duration},
		{logoStart, logoDisplayDuration},
		{overlayStart, overlayDisplayDuration},
	}

	for _, elem := range elements {
		endFrames := parseOffsetAndDuration(elem.offset, elem.duration)
		if endFrames > totalTimelineFrames {
			totalTimelineFrames = endFrames
		}
	}

	sequence.Duration = fmt.Sprintf("%d/24000s", totalTimelineFrames)

	// Validate FCPXML structure
	violations := ValidateClaudeCompliance(fcpxml)
	if len(violations) > 0 {
		t.Errorf("CLAUDE.md compliance violations found:")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}

	// Test XML marshaling
	output, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	xmlContent := string(output)

	// Verify complex timeline structure in generated XML
	testCases := []struct {
		name     string
		expected string
	}{
		{"Multiple gaps", `<gap name="Gap"`},
		{"Multiple asset-clips", `<asset-clip ref="`},
		{"Multiple video elements", `<video ref="`},
		{"Nested audio tracks", `lane="-1"`},
		{"Layered video tracks", `lane="1"`},
		{"Transform positioning", `position="300 200"`},
		{"Scale transformations", `scale="0.4 0.4"`},
		{"Audio effects", `name="DeEsser"`},
		{"Title overlays", `<title ref="`},
		{"Conform rate elements", `<conform-rate scaleEnabled="0"`},
		{"Multiple format references", fmt.Sprintf(`format="%s"`, format2ID)},
		{"Gap durations", `duration="48048/24000s"`}, // 2 seconds in frame-aligned format
		{"Complex timing offsets", `offset="360360/24000s"`}, // 15 seconds in frame-aligned format
		{"Audio roles", `audioRole="music"`},
		{"Video lanes", `lane="2"`},
	}

	for _, tc := range testCases {
		if !strings.Contains(xmlContent, tc.expected) {
			t.Errorf("Test '%s' failed: expected '%s' not found in XML", tc.name, tc.expected)
		}
	}

	// Write test file for manual FCP validation
	testFileName := "/tmp/test_complex_timeline.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFileName); os.IsNotExist(err) {
		t.Errorf("Test file was not created: %s", testFileName)
	}

	fmt.Printf("✅ Complex timeline test file created: %s\n", testFileName)
	fmt.Printf("   - 3 video tracks with overlaps and gaps\n")
	fmt.Printf("   - 2 audio tracks (music and SFX) with effects\n")
	fmt.Printf("   - 2 image overlays with positioning\n")
	fmt.Printf("   - Titles and text overlays\n")
	fmt.Printf("   - Multiple formats and conform-rate handling\n")
	fmt.Printf("   - Professional multi-track editing workflow\n")
}

// TestSynchronizedElements tests elements that must stay in sync
func TestSynchronizedElements(t *testing.T) {
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	// Reserve IDs for synchronized content
	ids := tx.ReserveIDs(8)
	mainVideoID := ids[0]
	syncVideoID := ids[1]
	audioID := ids[2]
	formatID := ids[3]
	audioFormatID := ids[4]
	markerEffectID := ids[5]
	syncEffectID := ids[6]
	timecodeEffectID := ids[7]

	// Create formats
	_, err = tx.CreateFormatWithFrameDuration(formatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create video format: %v", err)
	}

	_, err = tx.CreateFormat(audioFormatID, "FFAudioFormat48k", "", "", "")
	if err != nil {
		t.Fatalf("Failed to create audio format: %v", err)
	}

	// Create effects for synchronization
	_, err = tx.CreateEffect(markerEffectID, "Marker", "FFMarker")
	if err != nil {
		t.Fatalf("Failed to create marker effect: %v", err)
	}

	_, err = tx.CreateEffect(syncEffectID, "Timecode", "FFTimecode")
	if err != nil {
		t.Fatalf("Failed to create sync effect: %v", err)
	}

	_, err = tx.CreateEffect(timecodeEffectID, "Slate", "FFSlate")
	if err != nil {
		t.Fatalf("Failed to create timecode effect: %v", err)
	}

	// Create synchronized assets
	mainDuration := ConvertSecondsToFCPDuration(30.0)
	syncDuration := ConvertSecondsToFCPDuration(30.0) // Same duration for sync
	audioDuration := ConvertSecondsToFCPDuration(30.0) // Same duration for sync

	mainAsset, err := tx.CreateAsset(mainVideoID, "/Users/aa/cs/cutlass/assets/long.mov", "MainCamera", mainDuration, formatID)
	if err != nil {
		t.Fatalf("Failed to create main asset: %v", err)
	}

	syncAsset, err := tx.CreateAsset(syncVideoID, "/Users/aa/cs/cutlass/assets/speech1.mov", "SyncCamera", syncDuration, formatID)
	if err != nil {
		t.Fatalf("Failed to create sync asset: %v", err)
	}

	audioAsset, err := tx.CreateAsset(audioID, "/Users/aa/cs/cutlass/assets/Ethereal Accents.caf", "SyncAudio", audioDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create audio asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build synchronized timeline
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// All synchronized elements start at the same time
	syncStartTime := "0s"

	// Create multicam-style synchronized clip
	mainClip := AssetClip{
		Ref:       mainAsset.ID,
		Offset:    syncStartTime,
		Name:      mainAsset.Name,
		Duration:  mainDuration,
		Format:    mainAsset.Format,
		TCFormat:  "NDF",
		AudioRole: "dialogue",
		// Synchronized audio
		NestedAssetClips: []AssetClip{
			{
				Ref:       audioAsset.ID,
				Lane:      "-1",
				Offset:    syncStartTime, // Exact sync
				Name:      audioAsset.Name,
				Duration:  audioDuration,
				Format:    audioAsset.Format,
				AudioRole: "dialogue",
			},
			// Synchronized secondary camera
			{
				Ref:      syncAsset.ID,
				Lane:     "1", // Upper video layer
				Offset:   syncStartTime, // Perfect sync
				Name:     syncAsset.Name,
				Duration: syncDuration,
				Format:   syncAsset.Format,
				AdjustTransform: &AdjustTransform{
					Position: "400 300", // PIP position
					Scale:    "0.25 0.25", // Quarter size
				},
			},
		},
		// Sync markers and timecode
		FilterVideos: []FilterVideo{
			{
				Ref:  timecodeEffectID,
				Name: "Slate",
				Params: []Param{
					{
						Name:  "Show Timecode",
						Key:   "1",
						Value: "1",
					},
					{
						Name:  "Position",
						Key:   "2",
						Value: "0 -400", // Top center
					},
					{
						Name:  "Size",
						Key:   "3",
						Value: "24", // Font size
					},
				},
			},
		},
	}

	// Add sync markers at specific intervals (every 5 seconds)
	syncMarkerTimes := []float64{0.0, 5.0, 10.0, 15.0, 20.0, 25.0, 30.0}
	
	for i, markerTime := range syncMarkerTimes {
		markerOffset := ConvertSecondsToFCPDuration(markerTime)
		markerDuration := ConvertSecondsToFCPDuration(0.1) // Brief marker
		
		marker := Title{
			Ref:      markerEffectID,
			Lane:     fmt.Sprintf("%d", i+2), // High lanes for markers
			Offset:   markerOffset,
			Name:     fmt.Sprintf("Sync Marker %d", i+1),
			Duration: markerDuration,
			Start:    "86486400/24000s",
			Params: []Param{
				{
					Name:  "Marker Type",
					Key:   "1",
					Value: "sync",
				},
				{
					Name:  "Marker Text",
					Key:   "2",
					Value: fmt.Sprintf("SYNC %02d", i+1),
				},
				{
					Name:  "Frame Flash",
					Key:   "3",
					Value: "1", // Visual sync flash
				},
			},
		}
		
		mainClip.Titles = append(mainClip.Titles, marker)
	}

	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, mainClip)
	sequence.Duration = mainDuration

	// Validate and test
	violations := ValidateClaudeCompliance(fcpxml)
	if len(violations) > 0 {
		t.Errorf("CLAUDE.md compliance violations found:")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}

	// Write test file
	testFileName := "/tmp/test_synchronized_elements.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fmt.Printf("✅ Synchronized elements test file created: %s\n", testFileName)
	fmt.Printf("   - Multicam synchronized video/audio at identical timing\n")
	fmt.Printf("   - Sync markers every 5 seconds for reference\n")
	fmt.Printf("   - Timecode overlay for visual sync verification\n")
	fmt.Printf("   - Perfect frame-aligned synchronization\n")
	fmt.Printf("   - Professional multicam workflow\n")
}