package utils

import "cutlass/fcp"

// createFigure8Animation creates infinity symbol motion with variable speeds
// 🎬 FIGURE-8 PATTERN: Infinity symbol path with smooth transitions
// Position: Complex figure-8 trajectory with varying speeds
// Rotation: Following the curve direction with banking
// Scale: Perspective changes during the loop
func createFigure8Animation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createFigure8PositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createFigure8RotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createFigure8ScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createHeartbeatAnimation creates sharp, rhythmic cardiac-like pulses
// 🎬 HEARTBEAT PATTERN: Medical heartbeat rhythm with sharp peaks
// Scale: Sharp pulses (1.0 → 1.2 → 1.0) with realistic cardiac timing
// Position: Slight bump movement synchronized with beats
// Rotation: Minimal tilt during pulse peaks
func createHeartbeatAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createHeartbeatScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createHeartbeatPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createHeartbeatRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createPotpourriAnimation cycles through all effect types rapidly for maximum visual variety
// 🎬 POTPOURRI PATTERN: Fast-switching showcase of all effects in 1-second intervals
// Each second features a different effect's signature movement pattern
// Position, Scale, Rotation: Rapid style changes every second for dynamic presentation
func createPotpourriAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPotpourriPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPotpourriScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createPotpourriRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// createWindSwayAnimation simulates organic wind effects with irregular motion
// 🎬 WIND PATTERN: Organic, irregular swaying like a tree in wind
// Position: Irregular swaying with gusts and calm periods
// Rotation: Natural tilt variations following wind direction
// Scale: Subtle breathing effect from wind pressure
func createWindSwayAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createWindPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createWindRotationKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createWindScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// ============================================================================
// CREATIVE EFFECTS KEYFRAMES - Mathematical patterns for organic movement
// ============================================================================

// PARALLAX DEPTH KEYFRAMES
func createParallaxPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "-25 10"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "-40 25"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "-30 40"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.8), Value: "-10 30"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createParallaxScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.3), Value: "0.9 0.9", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.7), Value: "1.1 1.1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createParallaxRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "-1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}

// BREATHING KEYFRAMES
func createBreathingScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	breathCycle := duration / 4 // 4 breath cycles
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, breathCycle*0.4), Value: "1.06 1.06", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, breathCycle), Value: "0.96 0.96", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, breathCycle*1.4), Value: "1.08 1.08", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, breathCycle*2), Value: "0.95 0.95", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, breathCycle*2.4), Value: "1.07 1.07", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, breathCycle*3), Value: "0.97 0.97", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, breathCycle*3.4), Value: "1.05 1.05", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createBreathingPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "1 1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-1 -1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createBreathingRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}

// PENDULUM KEYFRAMES
func createPendulumPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "-50 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0 -20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "50 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "0 -20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "-50 0"},
	}
}

func createPendulumRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "-8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "-8", Curve: "linear"},
	}
}

func createPendulumScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0.95 1.05", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "0.95 1.05", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

// ELASTIC BOUNCE KEYFRAMES
func createElasticScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "0.6 1.8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.3), Value: "1.4 0.7", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "0.8 1.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "1.2 0.9", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "0.9 1.1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.9), Value: "1.05 0.95", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createElasticPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "15 -8"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "-20 12"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "8 -5"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.8), Value: "-3 2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createElasticRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "6", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "-4", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.8), Value: "-1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}
