package utils

import (
	"cutlass/fcp"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// createParticleEmitterEffect creates a fairy wand sparkle effect with multiple particles
// Each sparkle starts from the center and flies outward in different directions
// Uses multiple Video elements to simulate individual sparkles without needing Motion
func createParticleEmitterEffect(fcpxml *fcp.FCPXML, durationSeconds float64, videoStartTime string) error {
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]

	// Get the original image asset from the last added video
	if len(sequence.Spine.Videos) == 0 {
		return fmt.Errorf("no video elements found for particle emitter")
	}

	originalVideo := sequence.Spine.Videos[len(sequence.Spine.Videos)-1]

	// Create 30 sparkle particles (reasonable number for performance)
	numParticles := 30

	// Initialize random seed for consistent but varied particle behavior
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numParticles; i++ {
		// Create a new Video element for each sparkle
		sparkle := fcp.Video{
			Ref:             originalVideo.Ref, // Use same asset
			Name:            fmt.Sprintf("Sparkle_%d", i+1),
			Duration:        originalVideo.Duration,
			Start:           originalVideo.Start,
			AdjustTransform: createSparkleAnimation(i, durationSeconds, videoStartTime),
		}

		// Add sparkle to the spine
		sequence.Spine.Videos = append(sequence.Spine.Videos, sparkle)
	}

	// Make the original image invisible/very small so only sparkles show
	originalVideo.AdjustTransform = &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: []fcp.Keyframe{
						{Time: calculateAbsoluteTime(videoStartTime, 0), Value: "0.01 0.01"},
					},
				},
			},
		},
	}

	return nil
}

// createSparkleAnimation generates animation for a single sparkle particle
// Each sparkle has unique trajectory, timing, and scale animation
func createSparkleAnimation(particleIndex int, durationSeconds float64, videoStartTime string) *fcp.AdjustTransform {
	// Generate random direction and distance for this sparkle
	angle := float64(particleIndex)*(360.0/30.0) + rand.Float64()*30.0 - 15.0 // Spread around circle with some randomness
	distance := 400.0 + rand.Float64()*300.0                                  // Random distance 400-700 pixels

	// Calculate end position
	endX := distance * math.Cos(angle*math.Pi/180.0)
	endY := distance * math.Sin(angle*math.Pi/180.0)

	// Random timing offsets to make sparkles appear at different times
	startDelay := rand.Float64() * 0.5          // Delay up to 0.5 seconds
	sparkleLifetime := 2.0 + rand.Float64()*2.0 // Live for 2-4 seconds

	// Ensure sparkle doesn't go beyond total duration
	if startDelay+sparkleLifetime > durationSeconds {
		sparkleLifetime = durationSeconds - startDelay
	}

	return &fcp.AdjustTransform{
		Params: []fcp.Param{
			{
				Name: "position",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: []fcp.Keyframe{
						// Start at center
						{Time: calculateAbsoluteTime(videoStartTime, startDelay), Value: "0 0"},
						// Fly outward with easing
						{Time: calculateAbsoluteTime(videoStartTime, startDelay+sparkleLifetime),
							Value:  fmt.Sprintf("%.1f %.1f", endX, endY),
							Interp: "easeOut", Curve: "smooth"},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: []fcp.Keyframe{
						// Start invisible
						{Time: calculateAbsoluteTime(videoStartTime, startDelay), Value: "0 0"},
						// Burst to full size quickly
						{Time: calculateAbsoluteTime(videoStartTime, startDelay+0.1),
							Value: "0.15 0.15", Interp: "easeOut", Curve: "smooth"},
						// Hold size briefly
						{Time: calculateAbsoluteTime(videoStartTime, startDelay+sparkleLifetime*0.3),
							Value: "0.15 0.15"},
						// Fade out as it flies away
						{Time: calculateAbsoluteTime(videoStartTime, startDelay+sparkleLifetime),
							Value: "0.05 0.05", Interp: "easeIn", Curve: "smooth"},
					},
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: []fcp.Keyframe{
						// Rotate during flight for sparkle effect
						{Time: calculateAbsoluteTime(videoStartTime, startDelay), Value: "0"},
						{Time: calculateAbsoluteTime(videoStartTime, startDelay+sparkleLifetime),
							Value:  fmt.Sprintf("%.1f", 360.0+rand.Float64()*360.0),
							Interp: "linear", Curve: "linear"},
					},
				},
			},
		},
	}
}
