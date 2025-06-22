package utils

import "cutlass/fcp"

// createShatterArchiveAnchorKeyframes creates shifted pivot points simulating torn photo corners
// 📸 ANCHOR SHIFT PATTERN (Torn photo corners):
// Phase 1 (0-2.5s): MEMORY AWAKENING - Natural photo corner placement
// Phase 2 (2.5-5s): PHOTO REVELATION - Torn edge anchor shifts
// Phase 3 (5-7.5s): GLASS DISTORTION - Refraction pivot changes
// Phase 4 (7.5-10s): ANALOG DECAY - Corner burn and anchor dissolution
//
// 🎯 MATHEMATICAL BASIS: Torn photo corner simulation with organic anchor shifts
// Simulates photo corners ripping and changing the natural pivot point
func createShatterArchiveAnchorKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: MEMORY AWAKENING (0-2.5s) - Natural photo placement
		{Time: videoStartTime, Value: "-0.02 0.03", Curve: "linear"},                             // Slightly off-center (aged photo)
		{Time: calculateAbsoluteTime(videoStartTime, 0.6), Value: "-0.01 0.02", Curve: "linear"}, // Natural settle
		{Time: calculateAbsoluteTime(videoStartTime, 1.2), Value: "0.01 0.01", Curve: "linear"},  // Center approach
		{Time: calculateAbsoluteTime(videoStartTime, 1.8), Value: "0.02 -0.01", Curve: "linear"}, // Past center
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "0.03 -0.02", Curve: "linear"}, // Ready for reveal

		// PHASE 2: PHOTO REVELATION (2.5-5s) - Torn edge anchor shifts
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "0.05 -0.03", Curve: "linear"}, // First tear
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "0.08 -0.05", Curve: "linear"}, // Torn corner
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "0.06 -0.04", Curve: "linear"}, // Stutter back
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "0.10 -0.07", Curve: "linear"}, // Edge rip
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "0.07 -0.05", Curve: "linear"}, // Adjust anchor
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "0.12 -0.08", Curve: "linear"}, // Major tear
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "0.09 -0.06", Curve: "linear"}, // Settle tear
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "0.14 -0.10", Curve: "linear"}, // Maximum tear
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "0.11 -0.07", Curve: "linear"}, // Return motion
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "0.08 -0.05", Curve: "linear"}, // Stabilize
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "0.06 -0.04", Curve: "linear"}, // Final position
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "0.05 -0.03", Curve: "linear"}, // Phase end

		// PHASE 3: GLASS DISTORTION (5-7.5s) - Refraction pivot changes
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "0.08 -0.06", Curve: "linear"}, // Glass crack shift
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "0.03 -0.02", Curve: "linear"}, // Refraction pivot
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "0.11 -0.08", Curve: "linear"}, // Fragment view
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "0.01 -0.01", Curve: "linear"}, // Glass distortion
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "0.13 -0.09", Curve: "linear"}, // Maximum refraction
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "0.04 -0.03", Curve: "linear"}, // Crack align
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "0.09 -0.06", Curve: "linear"}, // Glass settle
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "0.06 -0.04", Curve: "linear"}, // Distortion ease
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "0.07 -0.05", Curve: "linear"}, // Final refraction
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "0.05 -0.03", Curve: "linear"}, // Glass clear
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "0.04 -0.03", Curve: "linear"}, // Last distortion
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0.03 -0.02", Curve: "linear"}, // Return anchor

		// PHASE 4: ANALOG DECAY (7.5-10s) - Corner burn and dissolution
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "0.06 -0.04", Curve: "linear"},      // Film edge burn
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "0.09 -0.06", Curve: "linear"},      // Corner burning
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "0.12 -0.08", Curve: "linear"},      // Film melting
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "0.15 -0.10", Curve: "linear"},      // Analog decay
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "0.18 -0.12", Curve: "linear"},      // Memory fragment
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "0.20 -0.14", Curve: "linear"},      // Film dissolution
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "0.22 -0.15", Curve: "linear"},      // Final burn
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "0.23 -0.16", Curve: "linear"},      // Near dissolution
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "0.24 -0.16", Curve: "linear"},      // Memory echo
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "0.24 -0.17", Curve: "linear"},      // Last anchor
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "0.25 -0.17", Curve: "linear"},      // Final moments
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "0.25 -0.17", Curve: "linear"},      // Almost gone
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0.25 -0.17", Curve: "linear"}, // Dissolved away
	}
}

// Note: Removed background layer functions as they created invisible effects

// createKaleidoAnimation creates a multi-layered animation with subtle movements to complement the kaleidoscope filter
// This combines gentle rotation, scaling, and position adjustments to create dynamic kaleidoscope patterns
func createKaleidoAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createKaleidoPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createKaleidoRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createKaleidoScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createKaleidoPositionKeyframes creates smooth position movements to enhance kaleidoscope effects
func createKaleidoPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// Start centered
		{Time: videoStartTime, Value: "0 0"},

		// Gentle orbital movement to create dynamic kaleidoscope patterns
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.05), Value: "2 1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.10), Value: "3 4"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "1 6"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.20), Value: "-2 5"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "-4 3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.30), Value: "-5 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.35), Value: "-4 -3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.40), Value: "-2 -5"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "1 -6"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.50), Value: "3 -4"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.55), Value: "5 -1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.60), Value: "4 2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.65), Value: "2 5"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.70), Value: "-1 6"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-4 4"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.80), Value: "-6 1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.85), Value: "-5 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.90), Value: "-2 -4"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.95), Value: "1 -3"},

		// Return to center
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

// createKaleidoRotationKeyframes creates continuous rotation with varying speeds to enhance kaleidoscope symmetry
func createKaleidoRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// Start at 0 degrees
		{Time: videoStartTime, Value: "0", Curve: "linear"},

		// Progressive rotation with speed variations to create interesting kaleidoscope patterns
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.05), Value: "8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.10), Value: "22", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "41", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.20), Value: "65", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "94", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.30), Value: "128", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.35), Value: "167", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.40), Value: "211", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "260", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.50), Value: "314", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.55), Value: "373", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.60), Value: "437", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.65), Value: "506", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.70), Value: "580", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "659", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.80), Value: "743", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.85), Value: "832", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.90), Value: "926", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.95), Value: "1025", Curve: "linear"},

		// End with approximately 3 full rotations
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1080", Curve: "linear"},
	}
}
