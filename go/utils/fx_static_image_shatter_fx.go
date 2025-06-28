package utils

import "cutlass/fcp"

// createShatterArchiveScaleKeyframes creates subtle breathing effect like living memories
// 📸 SCALE BREATHING PATTERN (Organic memory pulse):
// Phase 1 (0-2.5s): MEMORY AWAKENING - Gentle expansion like photo developing
// Phase 2 (2.5-5s): PHOTO REVELATION - Subtle pulsing during torn reveals
// Phase 3 (5-7.5s): GLASS DISTORTION - Magnification effects through cracked glass
// Phase 4 (7.5-10s): ANALOG DECAY - Shrinking as memory dissolves
//
// 🎯 MATHEMATICAL BASIS: Organic breathing patterns with film grain simulation
// Simulates photo paper expanding/contracting and analog magnification
func createShatterArchiveScaleKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: MEMORY AWAKENING (0-2.5s) - Photo developing expansion
		{Time: videoStartTime, Value: "0.95 0.95", Curve: "linear"},                             // Start slightly small (undeveloped)
		{Time: calculateAbsoluteTime(videoStartTime, 0.4), Value: "0.98 0.98", Curve: "linear"}, // Gentle expansion
		{Time: calculateAbsoluteTime(videoStartTime, 0.8), Value: "1.01 1.01", Curve: "linear"}, // Photo developing
		{Time: calculateAbsoluteTime(videoStartTime, 1.2), Value: "0.99 0.99", Curve: "linear"}, // Slight contraction
		{Time: calculateAbsoluteTime(videoStartTime, 1.6), Value: "1.02 1.02", Curve: "linear"}, // Expansion continue
		{Time: calculateAbsoluteTime(videoStartTime, 2.0), Value: "1.00 1.00", Curve: "linear"}, // Return to normal
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "1.03 1.03", Curve: "linear"}, // Ready for reveal

		// PHASE 2: PHOTO REVELATION (2.5-5s) - Pulsing during reveals
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "1.05 1.05", Curve: "linear"}, // Revelation pulse
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "1.02 1.02", Curve: "linear"}, // Settle back
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "1.06 1.06", Curve: "linear"}, // Torn edge reveal
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "1.03 1.03", Curve: "linear"}, // Stop-motion adjust
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "1.07 1.07", Curve: "linear"}, // Light leak pulse
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "1.04 1.04", Curve: "linear"}, // Breathing rhythm
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "1.08 1.08", Curve: "linear"}, // Maximum reveal
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "1.05 1.05", Curve: "linear"}, // Photo stability
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "1.03 1.03", Curve: "linear"}, // Gentle return
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "1.04 1.04", Curve: "linear"}, // Breathing maintain
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "1.02 1.02", Curve: "linear"}, // Settle rhythm
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "1.00 1.00", Curve: "linear"}, // Phase transition

		// PHASE 3: GLASS DISTORTION (5-7.5s) - Magnification through cracked glass
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "1.12 1.12", Curve: "linear"}, // Glass magnification
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "0.92 0.92", Curve: "linear"}, // Crack distortion
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "1.15 1.15", Curve: "linear"}, // Fragment zoom
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "0.88 0.88", Curve: "linear"}, // Glass refraction
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "1.18 1.18", Curve: "linear"}, // Maximum magnify
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "0.85 0.85", Curve: "linear"}, // Crack minimize
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "1.10 1.10", Curve: "linear"}, // Glass focus
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "0.95 0.95", Curve: "linear"}, // Distortion ease
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "1.05 1.05", Curve: "linear"}, // Final magnify
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "1.00 1.00", Curve: "linear"}, // Glass clear
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "1.02 1.02", Curve: "linear"}, // Last distortion
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "1.00 1.00", Curve: "linear"}, // Return to normal

		// PHASE 4: ANALOG DECAY (7.5-10s) - Memory shrinking as it dissolves
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "0.98 0.98", Curve: "linear"},      // Film edge curl
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "0.95 0.95", Curve: "linear"},      // Burn shrinkage
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "0.92 0.92", Curve: "linear"},      // Film melting
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "0.88 0.88", Curve: "linear"},      // Analog decay
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "0.85 0.85", Curve: "linear"},      // Memory fragment
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "0.82 0.82", Curve: "linear"},      // Film dissolution
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "0.78 0.78", Curve: "linear"},      // Final burn
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "0.75 0.75", Curve: "linear"},      // Near fade
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "0.72 0.72", Curve: "linear"},      // Memory echo
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "0.70 0.70", Curve: "linear"},      // Last flicker
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "0.68 0.68", Curve: "linear"},      // Final moments
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "0.65 0.65", Curve: "linear"},      // Almost gone
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "0.60 0.60", Curve: "linear"}, // Dissolved away
	}
}

// createShatterArchiveRotationKeyframes creates pendulum sway as if hanging from threads
// 📸 ROTATION PENDULUM PATTERN (Hanging photo aesthetic):
// Phase 1 (0-2.5s): MEMORY AWAKENING - Gentle sway like photos on string
// Phase 2 (2.5-5s): PHOTO REVELATION - Stop-motion stutter with torn edge tilts
// Phase 3 (5-7.5s): GLASS DISTORTION - Refraction angles through cracked glass
// Phase 4 (7.5-10s): ANALOG DECAY - Final tilt as memory falls away
//
// 🎯 MATHEMATICAL BASIS: Pendulum physics with stop-motion stuttering
// Simulates photos hanging from invisible threads with organic movement
func createShatterArchiveRotationKeyframes(duration float64, videoStartTime string) []fcp.Keyframe {
	return []fcp.Keyframe{
		// PHASE 1: MEMORY AWAKENING (0-2.5s) - Gentle pendulum sway
		{Time: videoStartTime, Value: "-2.5", Curve: "linear"},                             // Start tilted left (hanging)
		{Time: calculateAbsoluteTime(videoStartTime, 0.5), Value: "-1.8", Curve: "linear"}, // Swing toward center
		{Time: calculateAbsoluteTime(videoStartTime, 1.0), Value: "-0.8", Curve: "linear"}, // Past center
		{Time: calculateAbsoluteTime(videoStartTime, 1.5), Value: "0.5", Curve: "linear"},  // Swing right
		{Time: calculateAbsoluteTime(videoStartTime, 2.0), Value: "1.2", Curve: "linear"},  // Maximum right
		{Time: calculateAbsoluteTime(videoStartTime, 2.5), Value: "0.8", Curve: "linear"},  // Return swing

		// PHASE 2: PHOTO REVELATION (2.5-5s) - Stop-motion stutter tilts
		{Time: calculateAbsoluteTime(videoStartTime, 2.7), Value: "1.2", Curve: "linear"},  // Stutter back
		{Time: calculateAbsoluteTime(videoStartTime, 2.9), Value: "0.3", Curve: "linear"},  // Torn edge tilt
		{Time: calculateAbsoluteTime(videoStartTime, 3.1), Value: "0.8", Curve: "linear"},  // Stutter adjust
		{Time: calculateAbsoluteTime(videoStartTime, 3.3), Value: "-0.2", Curve: "linear"}, // Photo reveal angle
		{Time: calculateAbsoluteTime(videoStartTime, 3.5), Value: "0.5", Curve: "linear"},  // Stop-motion correct
		{Time: calculateAbsoluteTime(videoStartTime, 3.7), Value: "-0.7", Curve: "linear"}, // Light leak angle
		{Time: calculateAbsoluteTime(videoStartTime, 3.9), Value: "0.1", Curve: "linear"},  // Stutter settle
		{Time: calculateAbsoluteTime(videoStartTime, 4.1), Value: "-0.4", Curve: "linear"}, // Reveal position
		{Time: calculateAbsoluteTime(videoStartTime, 4.3), Value: "0.3", Curve: "linear"},  // Return motion
		{Time: calculateAbsoluteTime(videoStartTime, 4.5), Value: "-0.1", Curve: "linear"}, // Almost level
		{Time: calculateAbsoluteTime(videoStartTime, 4.7), Value: "0.2", Curve: "linear"},  // Final adjust
		{Time: calculateAbsoluteTime(videoStartTime, 5.0), Value: "0", Curve: "linear"},    // Phase transition

		// PHASE 3: GLASS DISTORTION (5-7.5s) - Refraction angles
		{Time: calculateAbsoluteTime(videoStartTime, 5.2), Value: "-1.5", Curve: "linear"}, // Glass crack angle
		{Time: calculateAbsoluteTime(videoStartTime, 5.4), Value: "1.8", Curve: "linear"},  // Refraction tilt
		{Time: calculateAbsoluteTime(videoStartTime, 5.6), Value: "-2.2", Curve: "linear"}, // Fragment view
		{Time: calculateAbsoluteTime(videoStartTime, 5.8), Value: "2.5", Curve: "linear"},  // Glass distortion
		{Time: calculateAbsoluteTime(videoStartTime, 6.0), Value: "-1.9", Curve: "linear"}, // Crack line view
		{Time: calculateAbsoluteTime(videoStartTime, 6.2), Value: "1.6", Curve: "linear"},  // Fragment align
		{Time: calculateAbsoluteTime(videoStartTime, 6.4), Value: "-1.2", Curve: "linear"}, // Glass settle
		{Time: calculateAbsoluteTime(videoStartTime, 6.6), Value: "0.9", Curve: "linear"},  // Distortion ease
		{Time: calculateAbsoluteTime(videoStartTime, 6.8), Value: "-0.6", Curve: "linear"}, // Final refraction
		{Time: calculateAbsoluteTime(videoStartTime, 7.0), Value: "0.3", Curve: "linear"},  // Glass clear
		{Time: calculateAbsoluteTime(videoStartTime, 7.2), Value: "-0.2", Curve: "linear"}, // Last distortion
		{Time: calculateAbsoluteTime(videoStartTime, 7.5), Value: "0", Curve: "linear"},    // Return level

		// PHASE 4: ANALOG DECAY (7.5-10s) - Final tilt as memory falls
		{Time: calculateAbsoluteTime(videoStartTime, 7.7), Value: "-0.8", Curve: "linear"},      // Film edge curl
		{Time: calculateAbsoluteTime(videoStartTime, 7.9), Value: "-1.5", Curve: "linear"},      // Burn tilt
		{Time: calculateAbsoluteTime(videoStartTime, 8.1), Value: "-2.3", Curve: "linear"},      // Film melting
		{Time: calculateAbsoluteTime(videoStartTime, 8.3), Value: "-3.1", Curve: "linear"},      // Analog decay
		{Time: calculateAbsoluteTime(videoStartTime, 8.5), Value: "-3.8", Curve: "linear"},      // Memory fragment
		{Time: calculateAbsoluteTime(videoStartTime, 8.7), Value: "-4.5", Curve: "linear"},      // Film dissolution
		{Time: calculateAbsoluteTime(videoStartTime, 8.9), Value: "-5.2", Curve: "linear"},      // Final burn
		{Time: calculateAbsoluteTime(videoStartTime, 9.1), Value: "-5.8", Curve: "linear"},      // Near fall
		{Time: calculateAbsoluteTime(videoStartTime, 9.3), Value: "-6.3", Curve: "linear"},      // Memory echo
		{Time: calculateAbsoluteTime(videoStartTime, 9.5), Value: "-6.7", Curve: "linear"},      // Last tilt
		{Time: calculateAbsoluteTime(videoStartTime, 9.7), Value: "-7.0", Curve: "linear"},      // Final moments
		{Time: calculateAbsoluteTime(videoStartTime, 9.9), Value: "-7.2", Curve: "linear"},      // Almost fallen
		{Time: calculateAbsoluteTime(videoStartTime, duration), Value: "-7.5", Curve: "linear"}, // Fallen away
	}
}
