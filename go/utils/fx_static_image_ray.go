package utils

import "cutlass/fcp"

// LIGHT RAYS EFFECT KEYFRAMES
func createLightRaysScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "1.1 1.1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "1.4 1.4", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "1.2 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.8), Value: "1.3 1.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createLightRaysPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.33), Value: "5 -8"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.66), Value: "-8 12"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createLightRaysRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "45", Curve: "linear"},
	}
}

// GLOW EFFECT KEYFRAMES
func createGlowScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "1.05 1.05", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "1.15 1.15", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "1.08 1.08", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createGlowPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0 -3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

// ============================================================================
// CREATIVE EFFECTS - Unique movement illusions for static images
// ============================================================================

// createParallaxDepthAnimation simulates depth by layering movement at different speeds
// 🎬 PARALLAX PATTERN: Multi-layer depth illusion with foreground/background movement
// Position: Large slow movement simulating distant background
// Scale: Subtle perspective changes to enhance depth
// Rotation: Minimal tilt to add dimensionality
func createParallaxDepthAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createParallaxPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createParallaxScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createParallaxRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createBreathingAnimation makes the image seem alive with organic breathing motion
// 🎬 BREATHING PATTERN: Rhythmic expansion/contraction like living tissue
// Scale: Gentle pulsing (0.95 to 1.08) with organic timing
// Position: Subtle floating movement synchronized with breathing
// Rotation: Minimal organic tilt variations
func createBreathingAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createBreathingScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createBreathingPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createBreathingRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createPendulumAnimation simulates realistic pendulum physics with gravity
// 🎬 PENDULUM PATTERN: Physics-based swinging with realistic acceleration
// Position: Arc motion with gravity-like deceleration at peaks
// Rotation: Synchronized tilt following the swing direction
// Scale: Subtle perspective changes during swing
func createPendulumAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPendulumPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPendulumRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPendulumScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createElasticBounceAnimation creates rubber-like stretching and snapping motion
// 🎬 ELASTIC PATTERN: Stretchy deformation with snapback physics
// Scale: Dramatic stretching (0.6 to 1.8) with elastic recovery
// Position: Compensating movement to maintain visual center
// Rotation: Wobble effect during elastic deformation
func createElasticBounceAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createElasticScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createElasticPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createElasticRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createSpiralVortexAnimation creates hypnotic inward/outward spiral motion
// 🎬 SPIRAL PATTERN: Vortex-like motion with rotation and scaling
// Rotation: Continuous spinning with acceleration phases
// Scale: Dramatic zoom cycles (0.3 to 2.0) synchronized with rotation
// Position: Spiral path with increasing/decreasing radius
func createSpiralVortexAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createSpiralRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createSpiralScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createSpiralPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}
