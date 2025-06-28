package utils

import "cutlass/fcp"

func createPotpourriRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// 0-1s: Shake (micro tilts)
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 0.5), Value: "-0.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 1), Value: "0.4", Curve: "linear"},

		// 1-2s: Perspective (3D tilt)
		{Time: calculateAbsoluteTime(videoStartTime, 1.5), Value: "-2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 2), Value: "3", Curve: "linear"},

		// 2-3s: Flip (full rotation)
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "90", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 3), Value: "180", Curve: "linear"},

		// 3-4s: 360-tilt (continuous spin)
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "270", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 4), Value: "360", Curve: "linear"},

		// 4-5s: Light-rays (slow rotation)
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "380", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 5), Value: "405", Curve: "linear"},

		// 5-6s: Parallax (minimal tilt)
		{Time: calculateAbsoluteTime(videoStartTime, 5.5), Value: "404", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 6), Value: "405", Curve: "linear"},

		// 6-7s: Breathe (organic tilt)
		{Time: calculateAbsoluteTime(videoStartTime, 6.5), Value: "405.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 7), Value: "405", Curve: "linear"},

		// 7-8s: Pendulum (swing tilt)
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "397", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 8), Value: "413", Curve: "linear"},

		// 8-9s: Elastic (wobble rotation)
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "419", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 9), Value: "409", Curve: "linear"},

		// 9-10s: Spiral (rapid spin finish)
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "629", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "720", Curve: "linear"},
	}
}

// ============================================================================
// INNER COLLAPSE EFFECT - "The Inner Collapse of a Digital Mind"
// ============================================================================

// createInnerCollapseAnimation creates a complex digital mind breakdown effect
// 🧠 INNER COLLAPSE PATTERN: Multi-phase psychological breakdown with recursive decay
//
// 🎬 CONCEPT: "The Inner Collapse of a Digital Mind" - 10 seconds of progressive disintegration
// Duration: 10.0 seconds, Cinematic Scope (2048x858), 24fps
//
// 🔧 TECHNICAL BREAKDOWN:
// Phase 1 (0-2.5s): STABILITY DECAY - Subtle glitches and micro-movements
// Phase 2 (2.5-5s): REALITY FRACTURE - Aggressive fragmentation and displacement
// Phase 3 (5-7.5s): RECURSIVE COLLAPSE - Self-consuming vortex motion
// Phase 4 (7.5-10s): DIGITAL DISSOLUTION - Final breakdown into data fragments
//
// 🎯 ANIMATION LAYERS:
// Position: Chaotic displacement with recursive feedback loops
// Scale: Dramatic compression/expansion cycles (0.1x to 3.0x)
// Rotation: Full 1080° rotation with acceleration/deceleration phases
// Anchor: Dynamic pivot points creating recursive transformation centers
//
// 🌀 MATHEMATICAL PATTERN: Fibonacci-based spiral decay with exponential acceleration
// 📊 KEYFRAME DENSITY: 50+ keyframes per parameter (300+ total) for microscopic control
// 🎭 PSYCHOLOGICAL TIMING: Matches human anxiety/panic attack progression curves
func createInnerCollapseAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			// Position Animation: Chaotic displacement with recursive feedback
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createInnerCollapsePositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			// Scale Animation: Dramatic compression/expansion cycles
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createInnerCollapseScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			// Rotation Animation: Full breakdown rotation with acceleration phases
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createInnerCollapseRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			// Anchor Animation: Dynamic pivot points for recursive transformation
			{
				Name: "anchor",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createInnerCollapseAnchorKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createInnerCollapsePositionKeyframes creates chaotic displacement with recursive feedback loops
// 🧠 POSITION BREAKDOWN PATTERN:
// Phase 1 (0-2.5s): STABILITY DECAY - Micro-glitches and neural noise (-5 to +5 pixels)
// Phase 2 (2.5-5s): REALITY FRACTURE - Aggressive displacement (-80 to +120 pixels)
// Phase 3 (5-7.5s): RECURSIVE COLLAPSE - Self-consuming spiral motion (150+ pixel radius)
// Phase 4 (7.5-10s): DIGITAL DISSOLUTION - Fragment scatter and data corruption
//
// 🎯 MATHEMATICAL BASIS: Fibonacci spiral with exponential decay + chaos theory
// Each position builds on previous with recursive feedback creating breakdown effect
func createInnerCollapsePositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: STABILITY DECAY (0-2.5s) - Micro-glitches
		{Time: videoStartTime, Value: "0 0"},                                // Perfect stability
		{Time: calculateAbsoluteTime(videoStartTime, 0.2), Value: "-1 0"},   // First glitch
		{Time: calculateAbsoluteTime(videoStartTime, 0.4), Value: "2 -1"},   // Neural noise
		{Time: calculateAbsoluteTime(videoStartTime, 0.6), Value: "-3 2"},   // Increasing instability
		{Time: calculateAbsoluteTime(videoStartTime, 0.8), Value: "1 -3"},   // Random drift
		{Time: calculateAbsoluteTime(videoStartTime, 1.0), Value: "-2 1"},   // Micro-tremor
		{Time: calculateAbsoluteTime(videoStartTime, 1.2), Value: "4 -2"},   // Glitch amplification
		{Time: calculateAbsoluteTime(videoStartTime, 1.4), Value: "-5 4"},   // System stress
		{Time: calculateAbsoluteTime(videoStartTime, 1.6), Value: "3 -5"},   // Breakdown warning
		{Time: calculateAbsoluteTime(videoStartTime, 1.8), Value: "-4 3"},   // Critical instability
		{Time: calculateAbsoluteTime(videoStartTime, 2.0), Value: "6 -4"},   // Cascade failure
		{Time: calculateAbsoluteTime(videoStartTime, 2.2), Value: "-8 6"},   // System panic
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "12 -10"}, // Stability lost

		// PHASE 2: REALITY FRACTURE (2.5-5s) - Aggressive fragmentation
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "-25 20"},   // Reality crack
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "40 -35"},   // Dimensional tear
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "-60 55"},   // Space fracture
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "80 -70"},   // Fabric rip
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "-100 90"},  // Reality collapse
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "120 -110"}, // Dimensional implosion
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "-90 100"},  // Chaotic rebound
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "70 -80"},   // Fragment scatter
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "-50 60"},   // Reality echo
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "30 -40"},   // Stabilization attempt
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "-20 25"},   // False recovery
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "0 0"},      // Momentary stillness

		// PHASE 3: RECURSIVE COLLAPSE (5-7.5s) - Self-consuming vortex
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "150 0"},     // Vortex edge
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "106 106"},   // Spiral arm 1
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "0 150"},     // Spiral arm 2
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "-106 106"},  // Spiral arm 3
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "-150 0"},    // Spiral arm 4
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "-106 -106"}, // Spiral arm 5
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "0 -150"},    // Spiral arm 6
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "106 -106"},  // Spiral arm 7
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "75 0"},      // Spiral contraction
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "53 53"},     // Inward spiral
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "0 75"},      // Collapse acceleration
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0 0"},       // Vortex center

		// PHASE 4: DIGITAL DISSOLUTION (7.5-10s) - Final breakdown
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "-200 150"}, // Data fragment 1
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "180 -120"}, // Data fragment 2
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "-160 100"}, // Data fragment 3
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "140 -80"},  // Data fragment 4
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "-120 60"},  // Data fragment 5
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "100 -40"},  // Data fragment 6
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "-80 20"},   // Data fragment 7
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "60 0"},     // Data fragment 8
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "-40 -20"},  // Data fragment 9
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "20 -40"},   // Data fragment 10
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "-10 -20"},  // Final scatter
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "5 -10"},    // Data corruption
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"}, // Complete dissolution
	}
}

// createInnerCollapseScaleKeyframes creates dramatic compression/expansion cycles
// 🧠 SCALE BREAKDOWN PATTERN:
// Phase 1 (0-2.5s): STABILITY DECAY - Micro-fluctuations (0.98x to 1.02x)
// Phase 2 (2.5-5s): REALITY FRACTURE - Extreme scaling (0.3x to 2.5x)
// Phase 3 (5-7.5s): RECURSIVE COLLAPSE - Vortex compression (0.1x to 3.0x)
// Phase 4 (7.5-10s): DIGITAL DISSOLUTION - Fragment scaling with data decay
//
// 🎯 MATHEMATICAL BASIS: Exponential decay curves with harmonic oscillation
// Simulates digital compression artifacts and memory allocation failures
func createInnerCollapseScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: STABILITY DECAY (0-2.5s) - Micro-fluctuations
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},                                   // Perfect stability
		{Time: calculateAbsoluteTime(videoStartTime, 0.3), Value: "1.01 0.99", Curve: "linear"}, // First micro-glitch
		{Time: calculateAbsoluteTime(videoStartTime, 0.6), Value: "0.98 1.02", Curve: "linear"}, // Neural noise
		{Time: calculateAbsoluteTime(videoStartTime, 0.9), Value: "1.03 0.97", Curve: "linear"}, // Increasing instability
		{Time: calculateAbsoluteTime(videoStartTime, 1.2), Value: "0.96 1.04", Curve: "linear"}, // System stress
		{Time: calculateAbsoluteTime(videoStartTime, 1.5), Value: "1.05 0.95", Curve: "linear"}, // Breakdown warning
		{Time: calculateAbsoluteTime(videoStartTime, 1.8), Value: "0.94 1.06", Curve: "linear"}, // Critical instability
		{Time: calculateAbsoluteTime(videoStartTime, 2.1), Value: "1.08 0.92", Curve: "linear"}, // Cascade failure
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "0.85 1.15", Curve: "linear"}, // Stability lost

		// PHASE 2: REALITY FRACTURE (2.5-5s) - Extreme scaling
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "0.6 1.8", Curve: "linear"}, // Reality crack
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "2.2 0.4", Curve: "linear"}, // Dimensional tear
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "0.3 2.5", Curve: "linear"}, // Space fracture
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "2.8 0.2", Curve: "linear"}, // Fabric rip
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "0.1 2.0", Curve: "linear"}, // Reality collapse
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "1.9 0.3", Curve: "linear"}, // Dimensional implosion
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "0.4 1.6", Curve: "linear"}, // Chaotic rebound
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "1.7 0.5", Curve: "linear"}, // Fragment scatter
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "0.7 1.4", Curve: "linear"}, // Reality echo
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "1.3 0.8", Curve: "linear"}, // Stabilization attempt
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "0.9 1.1", Curve: "linear"}, // False recovery
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "1 1", Curve: "linear"},     // Momentary stillness

		// PHASE 3: RECURSIVE COLLAPSE (5-7.5s) - Vortex compression
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "3.0 3.0", Curve: "linear"}, // Vortex expansion
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "0.2 0.2", Curve: "linear"}, // Compression snap
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "2.5 2.5", Curve: "linear"}, // Elastic rebound
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "0.3 0.3", Curve: "linear"}, // Vortex pull
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "2.0 2.0", Curve: "linear"}, // Spiral expansion
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "0.4 0.4", Curve: "linear"}, // Compression wave
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "1.5 1.5", Curve: "linear"}, // Spiral contraction
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "0.6 0.6", Curve: "linear"}, // Inward spiral
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "1.2 1.2", Curve: "linear"}, // Collapse acceleration
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "0.8 0.8", Curve: "linear"}, // Vortex center approach
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "0.5 0.5", Curve: "linear"}, // Near singularity
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0.1 0.1", Curve: "linear"}, // Vortex singularity

		// PHASE 4: DIGITAL DISSOLUTION (7.5-10s) - Fragment scaling
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "0.8 1.6", Curve: "linear"},   // Data fragment 1
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "1.4 0.7", Curve: "linear"},   // Data fragment 2
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "0.6 1.3", Curve: "linear"},   // Data fragment 3
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "1.2 0.8", Curve: "linear"},   // Data fragment 4
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "0.9 1.1", Curve: "linear"},   // Data fragment 5
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "1.1 0.9", Curve: "linear"},   // Data fragment 6
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "0.95 1.05", Curve: "linear"}, // Data fragment 7
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "1.05 0.95", Curve: "linear"}, // Data fragment 8
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "0.98 1.02", Curve: "linear"}, // Data fragment 9
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "1.02 0.98", Curve: "linear"}, // Data fragment 10
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "0.99 1.01", Curve: "linear"}, // Final scatter
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "1.01 0.99", Curve: "linear"}, // Data corruption
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},  // Complete dissolution
	}
}
