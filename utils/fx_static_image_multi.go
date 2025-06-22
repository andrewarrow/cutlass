package utils

import "cutlass/fcp"

// createMultiPhaseAnchorKeyframes generates dynamic pivot points for more interesting rotation centers
// 🚨 CRITICAL FIX: Anchor keyframes only support curve attribute, NOT interp (based on working samples)
// 🎬 ANCHOR POINT PATTERN FOR DYNAMIC ROTATION CENTERS:
// Phase 1 (0-25%): SLOW anchor drift (0,0) → (-0.1,0.05)
// Phase 2 (25-50%): FAST anchor change (-0.1,0.05) → (0.15,-0.1)
// Phase 3 (50-75%): SUPER FAST anchor movement (0.15,-0.1) → (-0.2,0.15)
// Phase 4 (75-100%): SLOW anchor settle (-0.2,0.15) → (0.05,-0.03)
func createMultiPhaseAnchorKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{
			Time:  videoStartTime, // Phase 1 Start: SLOW
			Value: "0 0",          // Start at center anchor
			Curve: "linear",       // Only curve attribute for anchor (like working samples)
		},
		{
			Time:  calculateAbsoluteTime(videoStartTime, duration*0.25), // 25% mark
			Value: "-0.1 0.05",                                          // Slight anchor offset
			Curve: "linear",                                             // Only curve attribute for anchor
		},
		{
			Time:  calculateAbsoluteTime(videoStartTime, duration*0.50), // 50% mark: FAST
			Value: "0.15 -0.1",                                          // Dramatic pivot shift
			Curve: "linear",                                             // Only curve attribute for anchor
		},
		{
			Time:  calculateAbsoluteTime(videoStartTime, duration*0.75), // 75% mark: SUPER FAST
			Value: "-0.2 0.15",                                          // Maximum dramatic pivot
			Curve: "linear",                                             // Only curve attribute for anchor
		},
		{
			Time:  calculateAbsoluteTime(videoStartTime, duration), // End: SLOW settle
			Value: "0.05 -0.03",                                    // Elegant final anchor point
			Curve: "linear",                                        // Only curve attribute for anchor
		},
	}
}

// createCameraShakeAnimation generates subtle handheld camera shake effects
// 🎬 SHAKE PATTERN: High-frequency micro-movements with random variations
// Position: Small random movements (-5 to +5 pixels)
// Rotation: Subtle tilt variations (-0.5° to +0.5°)
// Scale: Minor zoom fluctuations (98% to 102%)
func createCameraShakeAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createShakePositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createShakeRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createShakeScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createPerspective3DAnimation creates illusion of rotating 2D plane in 3D space
// 🎬 3D PERSPECTIVE PATTERN: Simulates depth and viewing angle changes
// Scale X/Y: Different ratios to simulate perspective (0.8-1.2 range)
// Position: Compensating movement to maintain visual center
// Rotation: Subtle tilt to enhance 3D illusion
func createPerspective3DAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPerspectivePositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPerspectiveScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPerspectiveRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createFlip3DAnimation creates dramatic 3D flip effects
// 🎬 FLIP PATTERN: Complete 180° rotations with perspective scaling
// Rotation: Full flip movements (0° → 180° → 360°)
// Scale: Dramatic perspective changes (1.0 → 0.1 → 1.0) to simulate depth
// Position: Slight movement to enhance 3D effect
func createFlip3DAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createFlipRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createFlipScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createFlipPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// create360TiltAnimation applies 360° tilt effects even on normal images
// 🎬 360° TILT PATTERN: Full rotation cycles with dynamic scaling
// Rotation: Complete 360° rotations (0° → 360° → 720°)
// Scale: Rhythmic zoom cycles synchronized with rotation
// Position: Orbital movement to enhance rotation effect
func create360TiltAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: create360TiltRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: create360TiltScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: create360TiltPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// create360PanAnimation applies 360° pan effects with orbital motion
// 🎬 360° PAN PATTERN: Circular orbital movement around center
// Position: Large circular motion (-100 to +100 pixel radius)
// Scale: Perspective changes as image "orbits" (0.8 to 1.3 range)
// Rotation: Counter-rotation to maintain orientation or enhance spin
func create360PanAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: create360PanPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: create360PanScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: create360PanRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createLightRaysAnimation simulates light rays/flares effects through transform
// 🎬 LIGHT RAYS PATTERN: Radiating movement with brightness simulation
// Scale: Pulsing effect to simulate light intensity (0.9 to 1.4)
// Position: Subtle radiating movement from center
// Rotation: Slow rotation to simulate moving light source
func createLightRaysAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createLightRaysScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createLightRaysPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createLightRaysRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}
