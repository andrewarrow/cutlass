package utils

import "cutlass/fcp"

// createGlowAnimation simulates glow effects through scaling and movement
// 🎬 GLOW PATTERN: Breathing effect with soft pulsing motion
// Scale: Gentle pulsing (0.95 to 1.15) to simulate glow breathing
// Position: Minimal floating movement
// All effects are subtle to maintain image clarity while adding glow feel
func createGlowAnimation(durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createGlowScaleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createGlowPositionKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}
}

// SHAKE EFFECT KEYFRAMES
func createShakePositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.1), Value: "-2 1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "3 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.3), Value: "-1 3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "4 -1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "-3 2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "2 -3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.7), Value: "-4 1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.8), Value: "1 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.9), Value: "-2 4"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createShakeRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "-0.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.3), Value: "0.4", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "-0.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "0.5", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-0.4", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.9), Value: "0.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}

func createShakeScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "1.01 0.99", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "0.99 1.02", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "1.02 0.98", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.8), Value: "0.98 1.01", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

// PERSPECTIVE 3D EFFECT KEYFRAMES
func createPerspectivePositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "-15 8"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "20 -12"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-10 15"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createPerspectiveScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0.8 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "1.2 0.8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "0.9 1.1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createPerspectiveRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.33), Value: "-2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.66), Value: "3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}

// FLIP 3D EFFECT KEYFRAMES
func createFlipRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "90", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "180", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "270", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "360", Curve: "linear"},
	}
}

func createFlipScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0.1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "0.1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createFlipPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0 -20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

// 360° TILT EFFECT KEYFRAMES
func create360TiltRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "360", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "720", Curve: "linear"},
	}
}

func create360TiltScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "1.3 1.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0.8 0.8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "1.4 1.4", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func create360TiltPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "30 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0 30"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-30 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

// 360° PAN EFFECT KEYFRAMES
func create360PanPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.125), Value: "70 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "50 50"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.375), Value: "0 70"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "-50 50"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.625), Value: "-70 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-50 -50"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.875), Value: "0 -70"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func create360PanScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "1.3 1.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0.8 0.8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "1.2 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func create360PanRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "-360", Curve: "linear"},
	}
}
