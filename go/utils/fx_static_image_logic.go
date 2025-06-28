package utils

import "cutlass/fcp"

// createInnerCollapseRotationKeyframes creates full breakdown rotation with acceleration phases
// 🧠 ROTATION BREAKDOWN PATTERN:
// Phase 1 (0-2.5s): STABILITY DECAY - Micro-tilts (-0.5° to +0.5°)
// Phase 2 (2.5-5s): REALITY FRACTURE - Aggressive rotation (-45° to +45°)
// Phase 3 (5-7.5s): RECURSIVE COLLAPSE - Full 720° rotation with acceleration
// Phase 4 (7.5-10s): DIGITAL DISSOLUTION - Fragment spin with data decay
//
// 🎯 MATHEMATICAL BASIS: Angular momentum conservation with chaos feedback
// Simulates gyroscopic failure and rotational instability in digital systems
func createInnerCollapseRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: STABILITY DECAY (0-2.5s) - Micro-tilts
		{Time: videoStartTime, Value: "0", Curve: "linear"},                                // Perfect stability
		{Time: calculateAbsoluteTime(videoStartTime, 0.4), Value: "-0.2", Curve: "linear"}, // First micro-tilt
		{Time: calculateAbsoluteTime(videoStartTime, 0.8), Value: "0.3", Curve: "linear"},  // Neural noise
		{Time: calculateAbsoluteTime(videoStartTime, 1.2), Value: "-0.5", Curve: "linear"}, // Increasing instability
		{Time: calculateAbsoluteTime(videoStartTime, 1.6), Value: "0.7", Curve: "linear"},  // System stress
		{Time: calculateAbsoluteTime(videoStartTime, 2.0), Value: "-1.0", Curve: "linear"}, // Breakdown warning
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "2.0", Curve: "linear"},  // Stability lost

		// PHASE 2: REALITY FRACTURE (2.5-5s) - Aggressive rotation
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "-8", Curve: "linear"},  // Reality crack
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "15", Curve: "linear"},  // Dimensional tear
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "-25", Curve: "linear"}, // Space fracture
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "35", Curve: "linear"},  // Fabric rip
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "-45", Curve: "linear"}, // Reality collapse
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "40", Curve: "linear"},  // Dimensional implosion
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "-30", Curve: "linear"}, // Chaotic rebound
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "20", Curve: "linear"},  // Fragment scatter
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "-10", Curve: "linear"}, // Reality echo
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "5", Curve: "linear"},   // Stabilization attempt
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "-2", Curve: "linear"},  // False recovery
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "0", Curve: "linear"},   // Momentary stillness

		// PHASE 3: RECURSIVE COLLAPSE (5-7.5s) - Full rotation acceleration
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "45", Curve: "linear"},   // Vortex start
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "135", Curve: "linear"},  // Acceleration phase 1
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "270", Curve: "linear"},  // Acceleration phase 2
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "450", Curve: "linear"},  // Acceleration phase 3
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "630", Curve: "linear"},  // Max velocity
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "720", Curve: "linear"},  // Vortex peak
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "765", Curve: "linear"},  // Spiral contraction
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "810", Curve: "linear"},  // Inward spiral
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "900", Curve: "linear"},  // Collapse acceleration
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "1080", Curve: "linear"}, // Vortex center approach
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "1260", Curve: "linear"}, // Near singularity
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "1440", Curve: "linear"}, // Vortex singularity

		// PHASE 4: DIGITAL DISSOLUTION (7.5-10s) - Fragment spin
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "1480", Curve: "linear"},      // Data fragment 1
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "1500", Curve: "linear"},      // Data fragment 2
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "1520", Curve: "linear"},      // Data fragment 3
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "1535", Curve: "linear"},      // Data fragment 4
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "1545", Curve: "linear"},      // Data fragment 5
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "1552", Curve: "linear"},      // Data fragment 6
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "1556", Curve: "linear"},      // Data fragment 7
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "1558", Curve: "linear"},      // Data fragment 8
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "1559", Curve: "linear"},      // Data fragment 9
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "1559.5", Curve: "linear"},    // Data fragment 10
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "1559.8", Curve: "linear"},    // Final scatter
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "1559.9", Curve: "linear"},    // Data corruption
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1560", Curve: "linear"}, // Complete dissolution
	}
}

// createInnerCollapseAnchorKeyframes creates dynamic pivot points for recursive transformation
// 🧠 ANCHOR BREAKDOWN PATTERN:
// Phase 1 (0-2.5s): STABILITY DECAY - Micro-shifts in pivot points
// Phase 2 (2.5-5s): REALITY FRACTURE - Chaotic anchor displacement
// Phase 3 (5-7.5s): RECURSIVE COLLAPSE - Spiral anchor pattern
// Phase 4 (7.5-10s): DIGITAL DISSOLUTION - Fragment anchor points
//
// 🎯 MATHEMATICAL BASIS: Recursive transformation matrices with chaos feedback
// Simulates anchor point instability during digital breakdown
func createInnerCollapseAnchorKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: STABILITY DECAY (0-2.5s) - Micro-shifts
		{Time: videoStartTime, Value: "0 0", Curve: "linear"},                                    // Perfect center
		{Time: calculateAbsoluteTime(videoStartTime, 0.5), Value: "-0.01 0.01", Curve: "linear"}, // First micro-shift
		{Time: calculateAbsoluteTime(videoStartTime, 1.0), Value: "0.02 -0.02", Curve: "linear"}, // Neural noise
		{Time: calculateAbsoluteTime(videoStartTime, 1.5), Value: "-0.03 0.03", Curve: "linear"}, // Increasing instability
		{Time: calculateAbsoluteTime(videoStartTime, 2.0), Value: "0.05 -0.05", Curve: "linear"}, // System stress
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "-0.08 0.08", Curve: "linear"}, // Stability lost

		// PHASE 2: REALITY FRACTURE (2.5-5s) - Chaotic displacement
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "-0.15 0.20", Curve: "linear"}, // Reality crack
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "0.25 -0.30", Curve: "linear"}, // Dimensional tear
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "-0.35 0.40", Curve: "linear"}, // Space fracture
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "0.45 -0.50", Curve: "linear"}, // Fabric rip
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "-0.55 0.60", Curve: "linear"}, // Reality collapse
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "0.50 -0.45", Curve: "linear"}, // Dimensional implosion
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "-0.40 0.35", Curve: "linear"}, // Chaotic rebound
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "0.30 -0.25", Curve: "linear"}, // Fragment scatter
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "-0.20 0.15", Curve: "linear"}, // Reality echo
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "0.10 -0.08", Curve: "linear"}, // Stabilization attempt
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "-0.05 0.03", Curve: "linear"}, // False recovery
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "0 0", Curve: "linear"},        // Momentary stillness

		// PHASE 3: RECURSIVE COLLAPSE (5-7.5s) - Spiral anchor pattern
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "0.3 0", Curve: "linear"},       // Spiral start
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "0.21 0.21", Curve: "linear"},   // Spiral arm 1
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "0 0.3", Curve: "linear"},       // Spiral arm 2
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "-0.21 0.21", Curve: "linear"},  // Spiral arm 3
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "-0.3 0", Curve: "linear"},      // Spiral arm 4
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "-0.21 -0.21", Curve: "linear"}, // Spiral arm 5
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "0 -0.3", Curve: "linear"},      // Spiral arm 6
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "0.21 -0.21", Curve: "linear"},  // Spiral arm 7
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "0.15 0", Curve: "linear"},      // Spiral contraction
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "0.08 0.08", Curve: "linear"},   // Inward spiral
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "0.03 0.03", Curve: "linear"},   // Collapse acceleration
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0 0", Curve: "linear"},         // Vortex center

		// PHASE 4: DIGITAL DISSOLUTION (7.5-10s) - Fragment anchor points
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "-0.4 0.3", Curve: "linear"},     // Data fragment 1
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "0.35 -0.25", Curve: "linear"},   // Data fragment 2
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "-0.3 0.2", Curve: "linear"},     // Data fragment 3
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "0.25 -0.15", Curve: "linear"},   // Data fragment 4
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "-0.2 0.1", Curve: "linear"},     // Data fragment 5
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "0.15 -0.08", Curve: "linear"},   // Data fragment 6
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "-0.1 0.05", Curve: "linear"},    // Data fragment 7
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "0.08 -0.03", Curve: "linear"},   // Data fragment 8
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "-0.05 0.02", Curve: "linear"},   // Data fragment 9
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "0.03 -0.01", Curve: "linear"},   // Data fragment 10
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "-0.01 0.005", Curve: "linear"},  // Final scatter
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "0.005 -0.002", Curve: "linear"}, // Data corruption
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0", Curve: "linear"},     // Complete dissolution
	}
}

// ============================================================================
// SHATTER ARCHIVE EFFECT - "Shatter Archive: The Dream Album"
// ============================================================================

// createShatterArchiveAnimation creates nostalgic stop-motion with analog photography decay
// 📸 SHATTER ARCHIVE PATTERN: Nostalgic stop-motion with emotional memory decay
//
// 🎬 CONCEPT: "Shatter Archive: The Dream Album" - 10 seconds of dreamlike photography
// Duration: 10.0 seconds, Stop-motion aesthetic (12fps simulation), Nostalgic mood
//
// 🔧 TECHNICAL BREAKDOWN:
// Phase 1 (0-2.5s): MEMORY AWAKENING - Gentle paper drift and sepia fade-in
// Phase 2 (2.5-5s): PHOTO REVELATION - Torn edge reveals with light leaks
// Phase 3 (5-7.5s): GLASS DISTORTION - Cracked viewing with magnification
// Phase 4 (7.5-10s): ANALOG DECAY - Film burn and fragile dissolution
//
// 🎯 ANIMATION LAYERS:
// Position: Gentle paper drift simulating aged documents moving from breeze
// Scale: Subtle breathing effect like living memories (0.95x to 1.08x)
// Rotation: Slight pendulum sway as if hanging from invisible threads
// Anchor: Shifted pivot points simulating torn photo corners
//
// 📷 AESTHETIC SIMULATION: Hand-cranked camera, torn paper masks, light leaks
// 🎭 EMOTIONAL TIMING: 12fps stutter for stop-motion authenticity
// 🕯️ NOSTALGIC DECAY: Sepia tones, film grain, and fragile imperfections
func createShatterArchiveAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			// Position Animation: Gentle paper drift like aged documents in breeze
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createShatterArchivePositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			// Scale Animation: Subtle breathing effect like living memories
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createShatterArchiveScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			// Rotation Animation: Slight pendulum sway as if hanging from threads
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createShatterArchiveRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			// Anchor Animation: Shifted pivot points simulating torn photo corners
			{
				Name: "anchor",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createShatterArchiveAnchorKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createShatterArchivePositionKeyframes creates gentle paper drift like aged documents
// 📸 POSITION DRIFT PATTERN (Stop-motion 12fps aesthetic):
// Phase 1 (0-2.5s): MEMORY AWAKENING - Slow horizontal drift from left (like slide projector)
// Phase 2 (2.5-5s): PHOTO REVELATION - Stuttered reveals with torn edge simulation
// Phase 3 (5-7.5s): GLASS DISTORTION - Subtle shifting as if viewed through cracked glass
// Phase 4 (7.5-10s): ANALOG DECAY - Final drift before memory fades
//
// 🎯 MATHEMATICAL BASIS: Organic drift patterns with 12fps stop-motion stuttering
// Simulates hand-cranked film projector and aged photo album pages turning
func createShatterArchivePositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: MEMORY AWAKENING (0-2.5s) - Slow horizontal drift
		{Time: videoStartTime, Value: "-15 5"},                             // Start off-center (like old photo placement)
		{Time: calculateAbsoluteTime(videoStartTime, 0.3), Value: "-12 4"}, // Gentle drift begins
		{Time: calculateAbsoluteTime(videoStartTime, 0.6), Value: "-8 3"},  // Continued drift
		{Time: calculateAbsoluteTime(videoStartTime, 0.9), Value: "-5 2"},  // Moving toward center
		{Time: calculateAbsoluteTime(videoStartTime, 1.2), Value: "-2 1"},  // Almost centered
		{Time: calculateAbsoluteTime(videoStartTime, 1.5), Value: "1 0"},   // Center crossing
		{Time: calculateAbsoluteTime(videoStartTime, 1.8), Value: "3 -1"},  // Past center
		{Time: calculateAbsoluteTime(videoStartTime, 2.1), Value: "5 -2"},  // Continuing right
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "8 -3"},  // End of phase 1

		// PHASE 2: PHOTO REVELATION (2.5-5s) - Stuttered reveals (12fps simulation)
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "6 -2"},  // Slight recoil (stop-motion stutter)
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "10 -4"}, // Photo edge reveal
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "8 -3"},  // Stutter back
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "12 -5"}, // Torn edge movement
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "9 -3"},  // Stop-motion adjustment
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "14 -6"}, // Light leak reveal
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "11 -4"}, // Stutter correction
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "15 -7"}, // Maximum reveal
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "12 -5"}, // Settling back
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "8 -3"},  // Return motion
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "5 -2"},  // Almost settled
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "2 0"},   // Phase 2 end

		// PHASE 3: GLASS DISTORTION (5-7.5s) - Cracked glass viewing
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "0 2"},   // Vertical shift (glass crack)
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "-3 1"},  // Diagonal distortion
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "1 -1"},  // Glass refraction
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "-2 3"},  // Crack line shift
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "4 0"},   // Glass fragment view
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "-1 -2"}, // Distortion continue
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "2 1"},   // Fragment alignment
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "-3 -1"}, // Glass stress
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "1 2"},   // Refraction shift
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "-1 0"},  // Glass settling
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "0 -1"},  // Final crack view
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0 0"},   // Return to center

		// PHASE 4: ANALOG DECAY (7.5-10s) - Film burn and memory fade
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "-5 3"},        // Film edge curl
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "-8 5"},        // Burn progression
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "-12 7"},       // Film melting
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "-15 9"},       // Analog decay
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "-18 11"},      // Memory fragmentation
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "-20 12"},      // Film dissolution
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "-22 13"},      // Final burn edge
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "-23 14"},      // Near complete fade
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "-24 14"},      // Memory echo
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "-24 15"},      // Last flicker
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "-25 15"},      // Final drift
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "-25 15"},      // Memory held
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "-25 15"}, // Dissolved away
	}
}
