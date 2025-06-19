package fcp

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestAudioMixing tests professional audio mixing scenarios with multiple
// tracks, levels, effects, and automation to validate the FCP package
// can handle sophisticated audio post-production workflows.
func TestAudioMixing(t *testing.T) {
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	// Reserve IDs for multiple audio assets and effects
	ids := tx.ReserveIDs(20)
	videoID := ids[0]
	dialogueID := ids[1]
	musicID := ids[2]
	sfxID := ids[3]
	ambienceID := ids[4]
	voiceoverID := ids[5]
	videoFormatID := ids[6]
	audioFormatID := ids[7]
	eqID := ids[8]
	compressorID := ids[9]
	deEsserID := ids[10]
	noiseGateID := ids[11]
	reverbID := ids[12]
	delayID := ids[13]
	limiterID := ids[14]
	chorusID := ids[15]
	pitchID := ids[16]
	autopanID := ids[17]
	duckerID := ids[18]
	matchEQID := ids[19]

	// Create formats
	_, err = tx.CreateFormatWithFrameDuration(videoFormatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
	if err != nil {
		t.Fatalf("Failed to create video format: %v", err)
	}

	_, err = tx.CreateFormat(audioFormatID, "FFAudioFormat48k", "", "", "")
	if err != nil {
		t.Fatalf("Failed to create audio format: %v", err)
	}

	// Create comprehensive audio effects suite
	audioEffects := []struct {
		id   string
		name string
		uid  string
	}{
		{eqID, "Channel EQ", "FFChannelEQ"},
		{compressorID, "Compressor", "FFCompressor"},
		{deEsserID, "DeEsser", "FFDeEsser"},
		{noiseGateID, "Noise Gate", "FFNoiseGate"},
		{reverbID, "ChromaVerb", "FFChromaVerb"},
		{delayID, "Echo", "FFEcho"},
		{limiterID, "Adaptive Limiter", "FFAdaptiveLimiter"},
		{chorusID, "Chorus", "FFChorus"},
		{pitchID, "Pitch", "FFPitch"},
		{autopanID, "AutoPan", "FFAutoPan"},
		{duckerID, "AutoDucker", "FFAutoDucker"},
		{matchEQID, "Match EQ", "FFMatchEQ"},
	}

	for _, effect := range audioEffects {
		_, err = tx.CreateEffect(effect.id, effect.name, effect.uid)
		if err != nil {
			t.Fatalf("Failed to create %s effect: %v", effect.name, err)
		}
	}

	// Create video and audio assets
	videoDuration := ConvertSecondsToFCPDuration(60.0)
	dialogueDuration := ConvertSecondsToFCPDuration(45.0)
	musicDuration := ConvertSecondsToFCPDuration(60.0)
	sfxDuration := ConvertSecondsToFCPDuration(30.0)
	ambienceDuration := ConvertSecondsToFCPDuration(60.0)
	voiceoverDuration := ConvertSecondsToFCPDuration(20.0)

	videoAsset, err := tx.CreateAsset(videoID, "/Users/aa/cs/cutlass/assets/long.mov", "MainVideo", videoDuration, videoFormatID)
	if err != nil {
		t.Fatalf("Failed to create video asset: %v", err)
	}

	dialogueAsset, err := tx.CreateAsset(dialogueID, "/Users/aa/cs/cutlass/assets/speech1.mov", "Dialogue", dialogueDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create dialogue asset: %v", err)
	}

	musicAsset, err := tx.CreateAsset(musicID, "/Users/aa/cs/cutlass/assets/Ethereal Accents.caf", "BackgroundMusic", musicDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create music asset: %v", err)
	}

	sfxAsset, err := tx.CreateAsset(sfxID, "/Users/aa/cs/cutlass/assets/Synth Zap Accent 06.caf", "SoundEffects", sfxDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create SFX asset: %v", err)
	}

	ambienceAsset, err := tx.CreateAsset(ambienceID, "/Users/aa/cs/cutlass/assets/Ethereal Accents.caf", "Ambience", ambienceDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create ambience asset: %v", err)
	}

	voiceoverAsset, err := tx.CreateAsset(voiceoverID, "/Users/aa/cs/cutlass/assets/speech2.mov", "Voiceover", voiceoverDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create voiceover asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build professional audio mix
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Create keyframe automation for audio levels and effects
	dialogueLevelKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "-12", // Start at -12dB
			},
			{
				Time:  ConvertSecondsToFCPDuration(5.0),
				Value: "-6", // Fade up to -6dB
			},
			{
				Time:  ConvertSecondsToFCPDuration(40.0),
				Value: "-6", // Hold level
			},
			{
				Time:  ConvertSecondsToFCPDuration(45.0),
				Value: "-20", // Fade out
			},
		},
	}

	musicLevelKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "-18", // Start quiet
			},
			{
				Time:  ConvertSecondsToFCPDuration(10.0),
				Value: "-12", // Fade up during intro
			},
			{
				Time:  ConvertSecondsToFCPDuration(15.0),
				Value: "-18", // Duck under dialogue
			},
			{
				Time:  ConvertSecondsToFCPDuration(50.0),
				Value: "-18", // Stay ducked
			},
			{
				Time:  ConvertSecondsToFCPDuration(60.0),
				Value: "-6", // End on high level
			},
		},
	}

	// EQ automation for dialogue
	eqFrequencyKeyframes := &KeyframeAnimation{
		Keyframes: []Keyframe{
			{
				Time:  "0s",
				Value: "2000", // Start frequency
			},
			{
				Time:  ConvertSecondsToFCPDuration(20.0),
				Value: "3000", // Sweep up for clarity
			},
			{
				Time:  ConvertSecondsToFCPDuration(45.0),
				Value: "1500", // Sweep down for warmth
			},
		},
	}

	mainClip := AssetClip{
		Ref:       videoAsset.ID,
		Offset:    "0s",
		Name:      videoAsset.Name,
		Duration:  videoDuration,
		Format:    videoAsset.Format,
		TCFormat:  "NDF",
		AudioRole: "dialogue",
		NestedAssetClips: []AssetClip{
			// Dialogue track with comprehensive processing
			{
				Ref:       dialogueAsset.ID,
				Lane:      "-1", // Primary dialogue lane
				Offset:    "0s",
				Name:      dialogueAsset.Name,
				Duration:  dialogueDuration,
				Format:    dialogueAsset.Format,
				AudioRole: "dialogue",
				FilterVideos: []FilterVideo{
					// EQ for dialogue clarity
					{
						Ref:  eqID,
						Name: "Channel EQ",
						Params: []Param{
							{
								Name:  "High Pass",
								Key:   "9999/999166631/999166650/1",
								Value: "80", // Remove low rumble
							},
							{
								Name:              "Mid Frequency",
								Key:               "9999/999166631/999166650/2",
								KeyframeAnimation: eqFrequencyKeyframes,
							},
							{
								Name:  "Mid Gain",
								Key:   "9999/999166631/999166650/3",
								Value: "3", // Boost presence
							},
							{
								Name:  "Mid Q",
								Key:   "9999/999166631/999166650/4",
								Value: "2.5", // Moderate Q
							},
							{
								Name:  "High Shelf",
								Key:   "9999/999166631/999166650/5",
								Value: "8000", // Air frequency
							},
							{
								Name:  "High Gain",
								Key:   "9999/999166631/999166650/6",
								Value: "2", // Add air
							},
						},
					},
					// DeEsser for sibilant control
					{
						Ref:  deEsserID,
						Name: "DeEsser",
						Params: []Param{
							{
								Name:  "Threshold",
								Key:   "9999/999166631/999166651/1",
								Value: "-15", // Sensitive threshold
							},
							{
								Name:  "Frequency",
								Key:   "9999/999166631/999166651/2",
								Value: "7000", // Sibilant frequency
							},
							{
								Name:  "Ratio",
								Key:   "9999/999166631/999166651/3",
								Value: "4:1", // Moderate reduction
							},
						},
					},
					// Compressor for dialogue consistency
					{
						Ref:  compressorID,
						Name: "Compressor",
						Params: []Param{
							{
								Name:  "Threshold",
								Key:   "9999/999166631/999166652/1",
								Value: "-12", // Compressor threshold
							},
							{
								Name:  "Ratio",
								Key:   "9999/999166631/999166652/2",
								Value: "3:1", // Gentle compression
							},
							{
								Name:  "Attack",
								Key:   "9999/999166631/999166652/3",
								Value: "5", // Fast attack
							},
							{
								Name:  "Release",
								Key:   "9999/999166631/999166652/4",
								Value: "100", // Medium release
							},
							{
								Name:  "Knee",
								Key:   "9999/999166631/999166652/5",
								Value: "2", // Soft knee
							},
							{
								Name:              "Makeup Gain",
								Key:               "9999/999166631/999166652/6",
								KeyframeAnimation: dialogueLevelKeyframes,
							},
						},
					},
					// Noise Gate for clean dialogue
					{
						Ref:  noiseGateID,
						Name: "Noise Gate",
						Params: []Param{
							{
								Name:  "Threshold",
								Key:   "9999/999166631/999166653/1",
								Value: "-45", // Gate threshold
							},
							{
								Name:  "Attack",
								Key:   "9999/999166631/999166653/2",
								Value: "2", // Fast gate open
							},
							{
								Name:  "Hold",
								Key:   "9999/999166631/999166653/3",
								Value: "50", // Hold time
							},
							{
								Name:  "Release",
								Key:   "9999/999166631/999166653/4",
								Value: "200", // Smooth gate close
							},
						},
					},
				},
			},
			// Background music with ducking automation
			{
				Ref:       musicAsset.ID,
				Lane:      "-2", // Music lane
				Offset:    "0s",
				Name:      musicAsset.Name,
				Duration:  musicDuration,
				Format:    musicAsset.Format,
				AudioRole: "music",
				FilterVideos: []FilterVideo{
					// EQ for music balance
					{
						Ref:  eqID,
						Name: "Channel EQ",
						Params: []Param{
							{
								Name:  "Low Shelf",
								Key:   "9999/999166631/999166654/1",
								Value: "100", // Low end control
							},
							{
								Name:  "Low Gain",
								Key:   "9999/999166631/999166654/2",
								Value: "-2", // Reduce low end
							},
							{
								Name:  "High Pass",
								Key:   "9999/999166631/999166654/3",
								Value: "40", // Remove sub-bass
							},
							{
								Name:  "Presence Cut",
								Key:   "9999/999166631/999166654/4",
								Value: "2500", // Cut dialogue range
							},
							{
								Name:  "Presence Gain",
								Key:   "9999/999166631/999166654/5",
								Value: "-3", // Make room for dialogue
							},
						},
					},
					// Auto-Ducker for dialogue interaction
					{
						Ref:  duckerID,
						Name: "AutoDucker",
						Params: []Param{
							{
								Name:  "Source",
								Key:   "9999/999166631/999166655/1",
								Value: "dialogue", // Duck when dialogue present
							},
							{
								Name:  "Threshold",
								Key:   "9999/999166631/999166655/2",
								Value: "-25", // Ducker sensitivity
							},
							{
								Name:  "Ratio",
								Key:   "9999/999166631/999166655/3",
								Value: "6:1", // Strong ducking
							},
							{
								Name:  "Attack",
								Key:   "9999/999166631/999166655/4",
								Value: "10", // Quick duck
							},
							{
								Name:  "Release",
								Key:   "9999/999166631/999166655/5",
								Value: "500", // Slow return
							},
						},
					},
					// Chorus for musical width
					{
						Ref:  chorusID,
						Name: "Chorus",
						Params: []Param{
							{
								Name:  "Rate",
								Key:   "9999/999166631/999166656/1",
								Value: "0.5", // Slow chorus rate
							},
							{
								Name:  "Depth",
								Key:   "9999/999166631/999166656/2",
								Value: "15", // Subtle depth
							},
							{
								Name:  "Mix",
								Key:   "9999/999166631/999166656/3",
								Value: "25", // Light chorus mix
							},
						},
					},
					// Compressor for music consistency
					{
						Ref:  compressorID,
						Name: "Compressor",
						Params: []Param{
							{
								Name:  "Threshold",
								Key:   "9999/999166631/999166657/1",
								Value: "-8", // Music compression threshold
							},
							{
								Name:  "Ratio",
								Key:   "9999/999166631/999166657/2",
								Value: "2.5:1", // Light compression
							},
							{
								Name:  "Attack",
								Key:   "9999/999166631/999166657/3",
								Value: "20", // Slower attack for music
							},
							{
								Name:  "Release",
								Key:   "9999/999166631/999166657/4",
								Value: "300", // Longer release
							},
							{
								Name:              "Output Level",
								Key:               "9999/999166631/999166657/5",
								KeyframeAnimation: musicLevelKeyframes,
							},
						},
					},
				},
			},
			// Sound effects with spatial processing
			{
				Ref:       sfxAsset.ID,
				Lane:      "-3", // SFX lane
				Offset:    ConvertSecondsToFCPDuration(10.0), // Start 10 seconds in
				Name:      sfxAsset.Name,
				Duration:  sfxDuration,
				Format:    sfxAsset.Format,
				AudioRole: "effects",
				FilterVideos: []FilterVideo{
					// Pitch shifting for creative effects
					{
						Ref:  pitchID,
						Name: "Pitch",
						Params: []Param{
							{
								Name: "Pitch Shift",
								Key:  "9999/999166631/999166658/1",
								KeyframeAnimation: &KeyframeAnimation{
									Keyframes: []Keyframe{
										{
											Time:  ConvertSecondsToFCPDuration(10.0),
											Value: "0", // No shift at start
										},
										{
											Time:  ConvertSecondsToFCPDuration(25.0),
											Value: "500", // Pitch up 5 semitones
										},
										{
											Time:  ConvertSecondsToFCPDuration(40.0),
											Value: "-300", // Pitch down 3 semitones
										},
									},
								},
							},
							{
								Name:  "Formant Correction",
								Key:   "9999/999166631/999166658/2",
								Value: "1", // Maintain formants
							},
						},
					},
					// AutoPan for movement
					{
						Ref:  autopanID,
						Name: "AutoPan",
						Params: []Param{
							{
								Name:  "Rate",
								Key:   "9999/999166631/999166659/1",
								Value: "0.2", // Slow pan rate
							},
							{
								Name:  "Depth",
								Key:   "9999/999166631/999166659/2",
								Value: "50", // Moderate pan depth
							},
							{
								Name:  "Waveform",
								Key:   "9999/999166631/999166659/3",
								Value: "sine", // Smooth panning
							},
						},
					},
					// Delay for space
					{
						Ref:  delayID,
						Name: "Echo",
						Params: []Param{
							{
								Name:  "Delay Time",
								Key:   "9999/999166631/999166660/1",
								Value: "250", // 250ms delay
							},
							{
								Name:  "Feedback",
								Key:   "9999/999166631/999166660/2",
								Value: "35", // Moderate feedback
							},
							{
								Name:  "Mix",
								Key:   "9999/999166631/999166660/3",
								Value: "20", // Light delay mix
							},
							{
								Name:  "High Cut",
								Key:   "9999/999166631/999166660/4",
								Value: "8000", // Roll off delay highs
							},
						},
					},
				},
			},
			// Ambience track for background
			{
				Ref:       ambienceAsset.ID,
				Lane:      "-4", // Ambience lane
				Offset:    "0s",
				Name:      ambienceAsset.Name,
				Duration:  ambienceDuration,
				Format:    ambienceAsset.Format,
				AudioRole: "effects.ambience",
				FilterVideos: []FilterVideo{
					// Reverb for space
					{
						Ref:  reverbID,
						Name: "ChromaVerb",
						Params: []Param{
							{
								Name:  "Room Type",
								Key:   "9999/999166631/999166661/1",
								Value: "Hall", // Large hall reverb
							},
							{
								Name:  "Size",
								Key:   "9999/999166631/999166661/2",
								Value: "85", // Large room size
							},
							{
								Name:  "Decay",
								Key:   "9999/999166631/999166661/3",
								Value: "3.5", // Long decay time
							},
							{
								Name:  "Pre-delay",
								Key:   "9999/999166631/999166661/4",
								Value: "50", // 50ms pre-delay
							},
							{
								Name:  "Mix",
								Key:   "9999/999166631/999166661/5",
								Value: "40", // Wet mix for ambience
							},
							{
								Name:  "High Cut",
								Key:   "9999/999166631/999166661/6",
								Value: "6000", // Darken reverb
							},
						},
					},
					// EQ for ambience shaping
					{
						Ref:  eqID,
						Name: "Channel EQ",
						Params: []Param{
							{
								Name:  "High Pass",
								Key:   "9999/999166631/999166662/1",
								Value: "60", // Remove low rumble
							},
							{
								Name:  "Low Mid Cut",
								Key:   "9999/999166631/999166662/2",
								Value: "200", // Reduce muddiness
							},
							{
								Name:  "Low Mid Gain",
								Key:   "9999/999166631/999166662/3",
								Value: "-4", // Cut low mids
							},
							{
								Name:  "High Shelf",
								Key:   "9999/999166631/999166662/4",
								Value: "10000", // High frequency shelf
							},
							{
								Name:  "High Gain",
								Key:   "9999/999166631/999166662/5",
								Value: "-6", // Reduce high end
							},
						},
					},
				},
			},
			// Voiceover with match EQ
			{
				Ref:       voiceoverAsset.ID,
				Lane:      "-5", // Voiceover lane
				Offset:    ConvertSecondsToFCPDuration(30.0), // Start at 30 seconds
				Name:      voiceoverAsset.Name,
				Duration:  voiceoverDuration,
				Format:    voiceoverAsset.Format,
				AudioRole: "dialogue.voiceover",
				FilterVideos: []FilterVideo{
					// Match EQ to dialogue
					{
						Ref:  matchEQID,
						Name: "Match EQ",
						Params: []Param{
							{
								Name:  "Reference Track",
								Key:   "9999/999166631/999166663/1",
								Value: "dialogue", // Match to dialogue track
							},
							{
								Name:  "Learn Mode",
								Key:   "9999/999166631/999166663/2",
								Value: "1", // Enable learning
							},
							{
								Name:  "Match Amount",
								Key:   "9999/999166631/999166663/3",
								Value: "75", // 75% match strength
							},
							{
								Name:  "Frequency Range",
								Key:   "9999/999166631/999166663/4",
								Value: "200-8000", // Focus on vocal range
							},
						},
					},
					// Adaptive Limiter for broadcast consistency
					{
						Ref:  limiterID,
						Name: "Adaptive Limiter",
						Params: []Param{
							{
								Name:  "Gain",
								Key:   "9999/999166631/999166664/1",
								Value: "6", // Input gain
							},
							{
								Name:  "Output Ceiling",
								Key:   "9999/999166631/999166664/2",
								Value: "-1", // -1dB ceiling
							},
							{
								Name:  "Release",
								Key:   "9999/999166631/999166664/3",
								Value: "50", // Fast release
							},
							{
								Name:  "Lookahead",
								Key:   "9999/999166631/999166664/4",
								Value: "5", // 5ms lookahead
							},
						},
					},
				},
			},
		},
	}

	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, mainClip)
	sequence.Duration = videoDuration

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

	// Verify audio mixing structure in generated XML
	testCases := []struct {
		name     string
		expected string
	}{
		{"Multiple audio lanes", `lane="-1"`},
		{"Dialogue audio role", `audioRole="dialogue"`},
		{"Music audio role", `audioRole="music"`},
		{"Effects audio role", `audioRole="effects"`},
		{"Voiceover audio role", `audioRole="dialogue.voiceover"`},
		{"Ambience audio role", `audioRole="effects.ambience"`},
		{"EQ effects", `name="Channel EQ"`},
		{"Compressor effects", `name="Compressor"`},
		{"DeEsser effects", `name="DeEsser"`},
		{"Noise Gate effects", `name="Noise Gate"`},
		{"Reverb effects", `name="ChromaVerb"`},
		{"Delay effects", `name="Echo"`},
		{"Limiter effects", `name="Adaptive Limiter"`},
		{"Chorus effects", `name="Chorus"`},
		{"Pitch effects", `name="Pitch"`},
		{"AutoPan effects", `name="AutoPan"`},
		{"AutoDucker effects", `name="AutoDucker"`},
		{"Match EQ effects", `name="Match EQ"`},
		{"Keyframe automation", `<keyframeAnimation>`},
		{"Audio parameters", `<param name="Threshold"`},
		{"Audio lanes", `lane="-5"`},
		{"Multiple filter-video elements", `</filter-video>`},
	}

	for _, tc := range testCases {
		if !strings.Contains(xmlContent, tc.expected) {
			t.Errorf("Test '%s' failed: expected '%s' not found in XML", tc.name, tc.expected)
		}
	}

	// Write test file for manual FCP validation
	testFileName := "/tmp/test_audio_mixing.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFileName); os.IsNotExist(err) {
		t.Errorf("Test file was not created: %s", testFileName)
	}

	fmt.Printf("✅ Audio mixing test file created: %s\n", testFileName)
	fmt.Printf("   - 5 audio tracks with full processing chains\n")
	fmt.Printf("   - Professional audio effects: EQ, compression, limiting\n")
	fmt.Printf("   - Creative effects: reverb, delay, chorus, pitch shifting\n")
	fmt.Printf("   - Audio automation with keyframe animation\n")
	fmt.Printf("   - Advanced tools: auto-ducking, match EQ, noise gate\n")
	fmt.Printf("   - Proper audio roles and lane management\n")
	fmt.Printf("   - Broadcast-ready audio post-production workflow\n")
}

// TestSurroundSoundMixing tests 5.1 surround sound mixing capabilities
func TestSurroundSoundMixing(t *testing.T) {
	fcpxml, err := GenerateEmpty("")
	if err != nil {
		t.Fatalf("Failed to create empty FCPXML: %v", err)
	}

	// Update sequence for 5.1 surround
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	sequence.AudioLayout = "5.1 Surround"
	sequence.AudioRate = "48k"

	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)

	// Reserve IDs for surround assets and effects
	ids := tx.ReserveIDs(12)
	centerID := ids[0]
	leftID := ids[1]
	rightID := ids[2]
	leftSurroundID := ids[3]
	rightSurroundID := ids[4]
	lfeID := ids[5]
	audioFormatID := ids[6]
	spatializerID := ids[7]
	surroundCompressorID := ids[8]
	surroundReverbID := ids[9]
	bassManagementID := ids[10]
	surroundEQID := ids[11]

	// Create audio format
	_, err = tx.CreateFormat(audioFormatID, "FFAudioFormat48k", "", "", "")
	if err != nil {
		t.Fatalf("Failed to create audio format: %v", err)
	}

	// Create surround sound effects
	surroundEffects := []struct {
		id   string
		name string
		uid  string
	}{
		{spatializerID, "Surround Spatializer", "FFSpatializer"},
		{surroundCompressorID, "Surround Compressor", "FFSurroundCompressor"},
		{surroundReverbID, "Surround Reverb", "FFSurroundReverb"},
		{bassManagementID, "Bass Management", "FFBassManagement"},
		{surroundEQID, "Surround EQ", "FFSurroundEQ"},
	}

	for _, effect := range surroundEffects {
		_, err = tx.CreateEffect(effect.id, effect.name, effect.uid)
		if err != nil {
			t.Fatalf("Failed to create %s effect: %v", effect.name, err)
		}
	}

	// Create surround channel assets
	channelDuration := ConvertSecondsToFCPDuration(30.0)

	centerAsset, err := tx.CreateAsset(centerID, "/Users/aa/cs/cutlass/assets/speech1.mov", "Center_Channel", channelDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create center asset: %v", err)
	}

	leftAsset, err := tx.CreateAsset(leftID, "/Users/aa/cs/cutlass/assets/Ethereal Accents.caf", "Left_Channel", channelDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create left asset: %v", err)
	}

	rightAsset, err := tx.CreateAsset(rightID, "/Users/aa/cs/cutlass/assets/Ethereal Accents.caf", "Right_Channel", channelDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create right asset: %v", err)
	}

	leftSurroundAsset, err := tx.CreateAsset(leftSurroundID, "/Users/aa/cs/cutlass/assets/Synth Zap Accent 06.caf", "Left_Surround", channelDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create left surround asset: %v", err)
	}

	rightSurroundAsset, err := tx.CreateAsset(rightSurroundID, "/Users/aa/cs/cutlass/assets/Synth Zap Accent 06.caf", "Right_Surround", channelDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create right surround asset: %v", err)
	}

	lfeAsset, err := tx.CreateAsset(lfeID, "/Users/aa/cs/cutlass/assets/Ethereal Accents.caf", "LFE_Channel", channelDuration, audioFormatID)
	if err != nil {
		t.Fatalf("Failed to create LFE asset: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Build 5.1 surround mix
	mainClip := AssetClip{
		Ref:       centerAsset.ID, // Center channel as main
		Offset:    "0s",
		Name:      "5.1 Surround Mix",
		Duration:  channelDuration,
		Format:    centerAsset.Format,
		AudioRole: "dialogue",
		NestedAssetClips: []AssetClip{
			// Left channel
			{
				Ref:       leftAsset.ID,
				Lane:      "-1",
				Offset:    "0s",
				Name:      leftAsset.Name,
				Duration:  channelDuration,
				Format:    leftAsset.Format,
				AudioRole: "music.left",
				FilterVideos: []FilterVideo{
					{
						Ref:  spatializerID,
						Name: "Surround Spatializer",
						Params: []Param{
							{
								Name:  "Channel Assignment",
								Key:   "1",
								Value: "Left",
							},
							{
								Name:  "Azimuth",
								Key:   "2",
								Value: "-30", // 30 degrees left
							},
							{
								Name:  "Distance",
								Key:   "3",
								Value: "1.0", // Unity distance
							},
						},
					},
				},
			},
			// Right channel
			{
				Ref:       rightAsset.ID,
				Lane:      "-2",
				Offset:    "0s",
				Name:      rightAsset.Name,
				Duration:  channelDuration,
				Format:    rightAsset.Format,
				AudioRole: "music.right",
				FilterVideos: []FilterVideo{
					{
						Ref:  spatializerID,
						Name: "Surround Spatializer",
						Params: []Param{
							{
								Name:  "Channel Assignment",
								Key:   "1",
								Value: "Right",
							},
							{
								Name:  "Azimuth",
								Key:   "2",
								Value: "30", // 30 degrees right
							},
							{
								Name:  "Distance",
								Key:   "3",
								Value: "1.0",
							},
						},
					},
				},
			},
			// Left surround
			{
				Ref:       leftSurroundAsset.ID,
				Lane:      "-3",
				Offset:    "0s",
				Name:      leftSurroundAsset.Name,
				Duration:  channelDuration,
				Format:    leftSurroundAsset.Format,
				AudioRole: "effects.surround.left",
				FilterVideos: []FilterVideo{
					{
						Ref:  spatializerID,
						Name: "Surround Spatializer",
						Params: []Param{
							{
								Name:  "Channel Assignment",
								Key:   "1",
								Value: "Left Surround",
							},
							{
								Name:  "Azimuth",
								Key:   "2",
								Value: "-110", // 110 degrees left rear
							},
							{
								Name:  "Distance",
								Key:   "3",
								Value: "1.2", // Slightly further
							},
						},
					},
					{
						Ref:  surroundReverbID,
						Name: "Surround Reverb",
						Params: []Param{
							{
								Name:  "Room Size",
								Key:   "1",
								Value: "large",
							},
							{
								Name:  "Rear Mix",
								Key:   "2",
								Value: "60", // More reverb in surrounds
							},
						},
					},
				},
			},
			// Right surround
			{
				Ref:       rightSurroundAsset.ID,
				Lane:      "-4",
				Offset:    "0s",
				Name:      rightSurroundAsset.Name,
				Duration:  channelDuration,
				Format:    rightSurroundAsset.Format,
				AudioRole: "effects.surround.right",
				FilterVideos: []FilterVideo{
					{
						Ref:  spatializerID,
						Name: "Surround Spatializer",
						Params: []Param{
							{
								Name:  "Channel Assignment",
								Key:   "1",
								Value: "Right Surround",
							},
							{
								Name:  "Azimuth",
								Key:   "2",
								Value: "110", // 110 degrees right rear
							},
							{
								Name:  "Distance",
								Key:   "3",
								Value: "1.2",
							},
						},
					},
					{
						Ref:  surroundReverbID,
						Name: "Surround Reverb",
						Params: []Param{
							{
								Name:  "Room Size",
								Key:   "1",
								Value: "large",
							},
							{
								Name:  "Rear Mix",
								Key:   "2",
								Value: "60",
							},
						},
					},
				},
			},
			// LFE channel
			{
				Ref:       lfeAsset.ID,
				Lane:      "-5",
				Offset:    "0s",
				Name:      lfeAsset.Name,
				Duration:  channelDuration,
				Format:    lfeAsset.Format,
				AudioRole: "effects.lfe",
				FilterVideos: []FilterVideo{
					{
						Ref:  bassManagementID,
						Name: "Bass Management",
						Params: []Param{
							{
								Name:  "Crossover Frequency",
								Key:   "1",
								Value: "80", // 80Hz crossover
							},
							{
								Name:  "LFE Level",
								Key:   "2",
								Value: "+10", // +10dB LFE boost
							},
							{
								Name:  "Phase",
								Key:   "3",
								Value: "0", // In phase
							},
						},
					},
				},
			},
		},
		// Master surround processing
		FilterVideos: []FilterVideo{
			{
				Ref:  surroundEQID,
				Name: "Surround EQ",
				Params: []Param{
					{
						Name:  "Center EQ",
						Key:   "1",
						Value: "flat", // Flat center response
					},
					{
						Name:  "LR EQ",
						Key:   "2",
						Value: "slight_boost", // Slight L/R boost
					},
					{
						Name:  "Surround EQ",
						Key:   "3",
						Value: "warm", // Warm surround tone
					},
					{
						Name:  "LFE EQ",
						Key:   "4",
						Value: "sub_optimized", // LFE optimization
					},
				},
			},
			{
				Ref:  surroundCompressorID,
				Name: "Surround Compressor",
				Params: []Param{
					{
						Name:  "Link Channels",
						Key:   "1",
						Value: "1", // Link all channels
					},
					{
						Name:  "Threshold",
						Key:   "2",
						Value: "-6", // Master threshold
					},
					{
						Name:  "Ratio",
						Key:   "3",
						Value: "2:1", // Gentle ratio
					},
				},
			},
		},
	}

	sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, mainClip)
	sequence.Duration = channelDuration

	// Validate and test
	violations := ValidateClaudeCompliance(fcpxml)
	if len(violations) > 0 {
		t.Errorf("CLAUDE.md compliance violations found:")
		for _, violation := range violations {
			t.Errorf("  - %s", violation)
		}
	}

	// Write test file
	testFileName := "/tmp/test_surround_sound_mixing.fcpxml"
	err = WriteToFile(fcpxml, testFileName)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fmt.Printf("✅ Surround sound mixing test file created: %s\n", testFileName)
	fmt.Printf("   - 5.1 surround sound channel configuration\n")
	fmt.Printf("   - Spatial audio positioning and distance\n")
	fmt.Printf("   - Surround-specific effects and processing\n")
	fmt.Printf("   - LFE channel with bass management\n")
	fmt.Printf("   - Master surround compression and EQ\n")
	fmt.Printf("   - Professional surround sound post-production\n")
}