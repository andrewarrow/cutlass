package utils

import "cutlass/fcp"

// createKaleidoScaleKeyframes creates subtle scaling variations to add depth to kaleidoscope patterns
func createKaleidoScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// Start at normal scale
		{Time: videoStartTime, Value: "1 1", Curve: "linear"},

		// Breathing scale effect with variations to create dynamic kaleidoscope reflections
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.05), Value: "1.02 1.01", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.10), Value: "1.05 1.03", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "1.03 1.06", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.20), Value: "0.98 1.05", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "0.95 1.02", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.30), Value: "0.97 0.98", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.35), Value: "1.01 0.96", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.40), Value: "1.06 0.97", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "1.08 1.01", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.50), Value: "1.06 1.06", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.55), Value: "1.02 1.08", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.60), Value: "0.98 1.06", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.65), Value: "0.96 1.02", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.70), Value: "0.97 0.98", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "1.01 0.96", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.80), Value: "1.05 0.98", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.85), Value: "1.07 1.02", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.90), Value: "1.04 1.05", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.95), Value: "1.01 1.03", Curve: "linear"},

		// Return to normal
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "1 1", Curve: "linear"},
	}
}

// Using direct animation on the image Video element for visible movement

// addKaleidoscopeFilter adds a kaleidoscope effect with animated parameters to create dynamic patterns
// Based on the .fcpxmld analysis, this animates both Segment Angle and Offset Angle with many keyframes
func addKaleidoscopeFilter(fcpxml *fcp.FCPXML, imageVideo *fcp.Video, durationSeconds float64, videoStartTime string) error {
	// Use ResourceRegistry to get the next available effect ID
	registry := fcp.NewResourceRegistry(fcpxml)
	tx := fcp.NewTransaction(registry)
	defer tx.Rollback()

	// Reserve an ID for the kaleidoscope effect
	ids := tx.ReserveIDs(1)
	kaleidoscopeEffectID := ids[0]

	// Add kaleidoscope effect to resources with verified UID from samples
	kaleidoscopeEffect := fcp.Effect{
		ID:   kaleidoscopeEffectID,
		Name: "Kaleidoscope",
		UID:  ".../Effects.localized/Tiling.localized/Kaleidoscope.localized/Kaleidoscope.moef",
	}

	// Add the effect to the resources
	fcpxml.Resources.Effects = append(fcpxml.Resources.Effects, kaleidoscopeEffect)

	// Create the kaleidoscope filter with animated parameters
	kaleidoscopeFilter := fcp.FilterVideo{
		Ref:  kaleidoscopeEffectID,
		Name: "Kaleidoscope",
		Params: []fcp.Param{
			{
				Name:  "Center",
				Key:   "9999/986883875/986883879/3/986883884/1",
				Value: "0.5 0.5", // Center of the image
			},
			{
				Name:  "Mix",
				Key:   "9999/986883875/986883879/3/986883884/10001",
				Value: "1", // Full effect
			},
			{
				Name: "Segment Angle",
				Key:  "9999/986883875/986883879/3/986883884/2",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createKaleidoSegmentAngleKeyframes(durationSeconds, videoStartTime),
				},
			},
			{
				Name: "Offset Angle",
				Key:  "9999/986883875/986883879/3/986883884/3",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: createKaleidoOffsetAngleKeyframes(durationSeconds, videoStartTime),
				},
			},
		},
	}

	// Add the filter to the video
	imageVideo.FilterVideos = append(imageVideo.FilterVideos, kaleidoscopeFilter)

	// Commit the transaction
	return tx.Commit()
}

// createKaleidoSegmentAngleKeyframes creates many keyframes for the Segment Angle parameter
// This controls the size of each kaleidoscope segment - animating from small to large segments
func createKaleidoSegmentAngleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// Start with small segments for intricate patterns
		{Time: videoStartTime, Value: "30", Curve: "linear"},

		// Gradually increase segment size with creative variations
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.05), Value: "45", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.10), Value: "60", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "40", Curve: "linear"}, // Variation
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.20), Value: "80", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "120", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.30), Value: "90", Curve: "linear"}, // Variation
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.35), Value: "150", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.40), Value: "180", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "135", Curve: "linear"}, // Variation
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.50), Value: "210", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.55), Value: "270", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.60), Value: "240", Curve: "linear"}, // Variation
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.65), Value: "300", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.70), Value: "330", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "315", Curve: "linear"}, // Variation
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.80), Value: "345", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.85), Value: "355", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.90), Value: "350", Curve: "linear"}, // Variation
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.95), Value: "358", Curve: "linear"},

		// End with full circle (360 degrees) for maximum symmetry
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "360", Curve: "linear"},
	}
}

// createKaleidoOffsetAngleKeyframes creates many keyframes for the Offset Angle parameter
// This controls the rotation of the kaleidoscope pattern - creates dynamic shifting patterns
func createKaleidoOffsetAngleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// Start at 0 degrees
		{Time: videoStartTime, Value: "0", Curve: "linear"},

		// Complex rotation pattern with speed variations and direction changes
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.05), Value: "15", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.10), Value: "35", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.15), Value: "25", Curve: "linear"}, // Reverse
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.20), Value: "60", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.25), Value: "95", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.30), Value: "80", Curve: "linear"}, // Slow down
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.35), Value: "125", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.40), Value: "175", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.45), Value: "155", Curve: "linear"}, // Reverse
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.50), Value: "210", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.55), Value: "270", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.60), Value: "245", Curve: "linear"}, // Slow down
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.65), Value: "305", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.70), Value: "365", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.75), Value: "340", Curve: "linear"}, // Reverse
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.80), Value: "400", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.85), Value: "465", Curve: "linear"},
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.90), Value: "445", Curve: "linear"}, // Slow down
		{Time: calculateAbsoluteTime(videoStartTime, duration*0.95), Value: "500", Curve: "linear"},

		// End with 540 degrees (1.5 full rotations) for dramatic finish
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "540", Curve: "linear"},
	}
}
