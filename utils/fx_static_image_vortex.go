package utils

import "cutlass/fcp"

// SPIRAL VORTEX KEYFRAMES
func createSpiralRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "180", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "540", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "900", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1080", Curve: "linear"},
	}
}

func createSpiralScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "2 2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0.8 0.8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0.3 0.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "0.8 0.8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "2 2", Curve: "linear"},
	}
}

func createSpiralPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.125), Value: "40 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "20 20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.375), Value: "0 15"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "-10 5"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.625), Value: "-3 -3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "2 -5"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.875), Value: "8 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

// FIGURE-8 KEYFRAMES
func createFigure8PositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.125), Value: "30 20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "40 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.375), Value: "30 -20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.625), Value: "-30 20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-40 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.875), Value: "-30 -20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createFigure8RotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}

func createFigure8ScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "1.2 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0.9 0.9", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "1.2 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

// HEARTBEAT KEYFRAMES
func createHeartbeatScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	beatInterval := duration / 6 // 6 heartbeats in duration
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*0.1), Value: "1.15 1.15", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*0.2), Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*0.35), Value: "1.2 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*0.5), Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*1.1), Value: "1.15 1.15", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*1.2), Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*1.35), Value: "1.2 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, beatInterval*1.5), Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

func createHeartbeatPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.1), Value: "0 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.35), Value: "0 -3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createHeartbeatRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.1), Value: "0.5", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.35), Value: "-0.5", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.5), Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}

// WIND SWAY KEYFRAMES
func createWindPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.1), Value: "-8 2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "-15 -3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "-25 1"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.55), Value: "-12 4"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.7), Value: "-18 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.85), Value: "-8 3"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createWindRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "0", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "-2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.3), Value: "-4", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "-6", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "-3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "-5", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.9), Value: "-2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0", Curve: "linear"},
	}
}

func createWindScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.2), Value: "1.02 0.98", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.4), Value: "0.98 1.03", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.6), Value: "1.01 0.99", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.8), Value: "0.99 1.02", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

// POTPOURRI KEYFRAMES - Cycling through all effects rapidly
// 🎬 EFFECT SEQUENCE (10 seconds, ~1 second each):
// 0-1s: Shake (micro movements)
// 1-2s: Perspective (3D illusion)
// 2-3s: Flip (rotation + scale)
// 3-4s: 360-tilt (full rotation)
// 4-5s: Light-rays (pulsing)
// 5-6s: Parallax (depth movement)
// 6-7s: Breathe (organic pulsing)
// 7-8s: Pendulum (physics swing)
// 8-9s: Elastic (stretchy bounce)
// 9-10s: Spiral (vortex motion)

func createPotpourriPositionKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// 0-1s: Shake (micro movements)
		{Time: videoStartTime, Value: "0 0"},
		{Time: calculateAbsoluteTime(videoStartTime, 0.5), Value: "-2 1"},
		{Time: calculateAbsoluteTime(videoStartTime, 1), Value: "3 -2"},

		// 1-2s: Perspective (3D positioning)
		{Time: calculateAbsoluteTime(videoStartTime, 1.5), Value: "-15 8"},
		{Time: calculateAbsoluteTime(videoStartTime, 2), Value: "20 -12"},

		// 2-3s: Flip (minimal movement during flip)
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "0 -10"},
		{Time: calculateAbsoluteTime(videoStartTime, 3), Value: "0 0"},

		// 3-4s: 360-tilt (orbital movement)
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "30 0"},
		{Time: calculateAbsoluteTime(videoStartTime, 4), Value: "0 30"},

		// 4-5s: Light-rays (radiating movement)
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "5 -8"},
		{Time: calculateAbsoluteTime(videoStartTime, 5), Value: "-8 12"},

		// 5-6s: Parallax (slow depth movement)
		{Time: calculateAbsoluteTime(videoStartTime, 5.5), Value: "-25 10"},
		{Time: calculateAbsoluteTime(videoStartTime, 6), Value: "-40 25"},

		// 6-7s: Breathe (subtle floating)
		{Time: calculateAbsoluteTime(videoStartTime, 6.5), Value: "0 -2"},
		{Time: calculateAbsoluteTime(videoStartTime, 7), Value: "1 1"},

		// 7-8s: Pendulum (wide swing)
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0 -20"},
		{Time: calculateAbsoluteTime(videoStartTime, 8), Value: "50 0"},

		// 8-9s: Elastic (bouncy movement)
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "15 -8"},
		{Time: calculateAbsoluteTime(videoStartTime, 9), Value: "-20 12"},

		// 9-10s: Spiral (vortex positioning)
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "20 20"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0 0"},
	}
}

func createPotpourriScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// 0-1s: Shake (micro scale changes)
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 0.5), Value: "1.01 0.99", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 1), Value: "0.99 1.02", Curve: "linear"},

		// 1-2s: Perspective (asymmetric scaling)
		{Time: calculateAbsoluteTime(videoStartTime, 1.5), Value: "0.8 1.2", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 2), Value: "1.2 0.8", Curve: "linear"},

		// 2-3s: Flip (dramatic perspective changes)
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "0.1 1", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 3), Value: "1 1", Curve: "linear"},

		// 3-4s: 360-tilt (rhythmic zoom)
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "1.3 1.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 4), Value: "0.8 0.8", Curve: "linear"},

		// 4-5s: Light-rays (pulsing intensity)
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "1.4 1.4", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 5), Value: "1.2 1.2", Curve: "linear"},

		// 5-6s: Parallax (depth scaling)
		{Time: calculateAbsoluteTime(videoStartTime, 5.5), Value: "0.9 0.9", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 6), Value: "1.1 1.1", Curve: "linear"},

		// 6-7s: Breathe (organic pulsing)
		{Time: calculateAbsoluteTime(videoStartTime, 6.5), Value: "1.06 1.06", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 7), Value: "0.96 0.96", Curve: "linear"},

		// 7-8s: Pendulum (perspective swing)
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0.95 1.05", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 8), Value: "1 1", Curve: "linear"},

		// 8-9s: Elastic (dramatic stretching)
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "0.6 1.8", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, 9), Value: "1.4 0.7", Curve: "linear"},

		// 9-10s: Spiral (vortex scaling)
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "0.3 0.3", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}
