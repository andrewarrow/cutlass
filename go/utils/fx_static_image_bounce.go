package utils

import (
	"cutlass/fcp"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// createWordBounceEffect creates animated text words with blade-cut bouncing animation like three_words.fcpxml
//
// 🎬 WORD BOUNCE EFFECT: Creates N text elements with repeated blade cuts for smooth bouncing
// Based on three_words.fcpxml pattern:
// - Each word gets multiple blade cuts at different positions and times
// - Small blade durations (0.1-0.2 seconds) for smooth movement illusion
// - Random X,Y positioning for each blade cut to create bouncing effect
// - Uses Avenir Next Condensed Heavy Italic font with magenta color and white stroke
// - 9 seconds total duration with multiple position changes
// - Uses verified Text effect UID from samples
func createWordBounceEffect(fcpxml *fcp.FCPXML, durationSeconds float64, videoStartTime string, fontColor string, outlineColor string) error {
	// Get words from environment variable or use default set
	wordsParam := os.Getenv("WORDS")
	if wordsParam == "" {
		wordsParam = "anger,tattle,entertainment,compilation"
	}

	words := strings.Split(wordsParam, ",")
	// Support N words (remove the 4-word limit)

	// Add Text effect to resources if not already present
	textEffectID := "r4" // Use consistent ID like samples
	hasTextEffect := false
	for _, effect := range fcpxml.Resources.Effects {
		if effect.UID == ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti" {
			hasTextEffect = true
			textEffectID = effect.ID
			break
		}
	}

	if !hasTextEffect {
		fcpxml.Resources.Effects = append(fcpxml.Resources.Effects, fcp.Effect{
			ID:   textEffectID,
			Name: "Text",
			UID:  ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
		})
	}

	// Get the background video to add titles to
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	if len(sequence.Spine.Videos) == 0 {
		return fmt.Errorf("no video elements found in spine")
	}

	backgroundVideo := &sequence.Spine.Videos[len(sequence.Spine.Videos)-1]

	// Initialize random seed for positioning
	rand.Seed(time.Now().UnixNano())

	// Create blade-cut animated text elements for each word (following Info.fcpxml pattern)
	// Scale blade count with duration to maintain consistent visual density
	bladesPerSecond := 240.0 / 9.0 // ~26.67 blades per second (from original 9s design)
	totalBlades := int(durationSeconds * bladesPerSecond)
	bladeDuration := durationSeconds / float64(totalBlades)

	textStyleCounter := 1 // Global counter for unique text style IDs

	// Position tracking for each word (incremental movement)
	wordPositions := make(map[string]struct {
		x, y                   int
		directionX, directionY int
	})

	// Track occupied areas to prevent word overlap
	type wordArea struct {
		x, y, width, height int
	}
	occupiedAreas := make([]wordArea, 0, len(words))

	// Initialize starting positions and directions for each word with collision avoidance
	for i, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		// Add word index to seed for more variation between words
		rand.Seed(time.Now().UnixNano() + int64(i*1000))

		// Estimate word size (approximate based on character count and font size)
		// Using fontSize 400 from Info.fcpxml as reference
		wordWidth := len(word) * 300 // Rough estimate: 300px per character
		wordHeight := 500            // Rough estimate: 500px height for large text

		var newX, newY int
		maxAttempts := 100

		// Try to find a non-overlapping position
		for attempt := 0; attempt < maxAttempts; attempt++ {
			newX = rand.Intn(2000) - 1000 // Full X range: -1000 to +1000
			newY = -rand.Intn(4000)       // Full Y range: 0 to -4000

			// Check if this position overlaps with any existing word
			overlaps := false
			for _, area := range occupiedAreas {
				if newX < area.x+area.width && newX+wordWidth > area.x &&
					newY < area.y+area.height && newY+wordHeight > area.y {
					overlaps = true
					break
				}
			}

			if !overlaps {
				break // Found a good position
			}
		}

		// Record this word's occupied area
		occupiedAreas = append(occupiedAreas, wordArea{
			x: newX, y: newY, width: wordWidth, height: wordHeight,
		})

		wordPositions[word] = struct {
			x, y                   int
			directionX, directionY int
		}{
			x:          newX,
			y:          newY,
			directionX: []int{-1, 1}[rand.Intn(2)], // Random initial direction: -1 or +1
			directionY: []int{-1, 1}[rand.Intn(2)], // Random initial direction: -1 or +1
		}
	}

	for i, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		// Calculate lane number (distribute across lanes, one lane per word)
		laneNum := i + 1 // Lanes 1, 2, 3, ..., N for N words

		// Create multiple blade cuts for this word throughout the timeline
		for bladeIndex := 0; bladeIndex < totalBlades; bladeIndex++ {

			// Calculate timing for this blade cut
			bladeStartTime := float64(bladeIndex) * (durationSeconds / float64(totalBlades))

			// Use proper frame-aligned offset calculation
			bladeOffset := videoStartTime
			if bladeStartTime > 0 {
				// Parse the video start time and add frame-aligned offset
				var startNumerator, timeBase int
				if _, err := fmt.Sscanf(videoStartTime, "%d/%ds", &startNumerator, &timeBase); err != nil {
					startNumerator = 86399313
					timeBase = 24000
				}

				// Convert blade start time to frame-aligned offset using ConvertSecondsToFCPDuration
				bladeOffsetDuration := fcp.ConvertSecondsToFCPDuration(bladeStartTime)
				var offsetNumerator int
				if _, err := fmt.Sscanf(bladeOffsetDuration, "%d/%ds", &offsetNumerator, &timeBase); err == nil {
					bladeOffset = fmt.Sprintf("%d/%ds", startNumerator+offsetNumerator, timeBase)
				}
			}

			bladeDurationFCP := fcp.ConvertSecondsToFCPDuration(bladeDuration)

			// Update position incrementally for smooth movement (no jumps)
			pos := wordPositions[word]

			// Move at constant speed like a screensaver
			moveSpeed := 8 // Constant speed for smooth screensaver-like movement
			pos.x += pos.directionX * moveSpeed
			pos.y += pos.directionY * moveSpeed

			// Bounce off boundaries like a screensaver (flip direction when hitting wall)
			if pos.x > 1000 {
				pos.x = 1000
				pos.directionX = -1 // Bounce off right wall
			} else if pos.x < -1000 {
				pos.x = -1000
				pos.directionX = 1 // Bounce off left wall
			}

			if pos.y > 100 {
				pos.y = 100
				pos.directionY = -1 // Bounce off top wall (small positive allowed)
			} else if pos.y < -4000 {
				pos.y = -4000
				pos.directionY = 1 // Bounce off bottom wall
			}

			// Update the position in the map
			wordPositions[word] = pos

			// Use the updated position
			currentX := pos.x
			currentY := pos.y

			// Create unique text style ID for each blade cut
			textStyleID := fmt.Sprintf("ts%d", textStyleCounter)
			textStyleCounter++

			// Create title element based on three_words.fcpxml pattern
			titleElement := fcp.Title{
				Ref:      textEffectID,
				Lane:     fmt.Sprintf("%d", laneNum),
				Offset:   bladeOffset,
				Name:     fmt.Sprintf("%s - Text", word),
				Duration: bladeDurationFCP,
				Start:    "0s", // Relative to video start
				Params: []fcp.Param{
					// Build In/Out settings from sample
					{
						Name:  "Build In",
						Key:   "9999/10000/2/101",
						Value: "0",
					},
					{
						Name:  "Build Out",
						Key:   "9999/10000/2/102",
						Value: "0",
					},
					// Incremental position for smooth bounce effect (key from sample)
					{
						Name:  "Position",
						Key:   "9999/10003/13260/3296672360/1/100/101",
						Value: fmt.Sprintf("%d %d", currentX, currentY),
					},
					// Layout settings from three_words.fcpxml
					{
						Name:  "Layout Method",
						Key:   "9999/10003/13260/3296672360/2/314",
						Value: "1 (Paragraph)",
					},
					{
						Name:  "Left Margin",
						Key:   "9999/10003/13260/3296672360/2/323",
						Value: "-1210", // From sample
					},
					{
						Name:  "Right Margin",
						Key:   "9999/10003/13260/3296672360/2/324",
						Value: "1210", // From sample
					},
					{
						Name:  "Top Margin",
						Key:   "9999/10003/13260/3296672360/2/325",
						Value: "2160", // From sample
					},
					{
						Name:  "Bottom Margin",
						Key:   "9999/10003/13260/3296672360/2/326",
						Value: "-2160", // From sample
					},
					{
						Name:  "Alignment",
						Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
						Value: "1 (Center)", // From sample
					},
					{
						Name:  "Line Spacing",
						Key:   "9999/10003/13260/3296672360/2/354/3296667315/404",
						Value: "-19", // From sample
					},
					{
						Name:  "Auto-Shrink",
						Key:   "9999/10003/13260/3296672360/2/370",
						Value: "3 (To All Margins)", // From sample
					},
					{
						Name:  "Alignment",
						Key:   "9999/10003/13260/3296672360/2/373",
						Value: "0 (Left) 0 (Top)", // From sample
					},
					// Initial opacity (invisible)
					{
						Name:  "Opacity",
						Key:   "9999/10003/13260/3296672360/4/3296673134/1000/1044",
						Value: "0", // From sample
					},
					// Custom speed animation for dramatic entrance (from sample)
					{
						Name:  "Speed",
						Key:   "9999/10003/13260/3296672360/4/3296673134/201/208",
						Value: "6 (Custom)", // From sample
					},
					{
						Name: "Custom Speed",
						Key:  "9999/10003/13260/3296672360/4/3296673134/201/209",
						KeyframeAnimation: &fcp.KeyframeAnimation{
							Keyframes: []fcp.Keyframe{
								{
									Time:  "-469658744/1000000000s", // From sample - Start invisible
									Value: "0",
								},
								{
									Time:  "12328542033/1000000000s", // From sample - Fade in
									Value: "1",
								},
							},
						},
					},
					{
						Name:  "Apply Speed",
						Key:   "9999/10003/13260/3296672360/4/3296673134/201/211",
						Value: "2 (Per Object)", // From sample
					},
				},
				Text: &fcp.TitleText{
					TextStyles: []fcp.TextStyleRef{
						{
							Ref:  textStyleID,
							Text: word,
						},
					},
				},
				TextStyleDefs: []fcp.TextStyleDef{
					{
						ID: textStyleID,
						TextStyle: fcp.TextStyle{
							Font:        "Avenir Next Condensed", // From sample
							FontSize:    "400",                   // From sample
							FontFace:    "Heavy Italic",          // From sample
							FontColor:   fontColor,               // Custom font color from CLI parameter
							Bold:        "1",                     // From sample
							Italic:      "1",                     // From sample
							StrokeColor: outlineColor,            // Custom outline color from CLI parameter
							StrokeWidth: "-15",                   // From sample
							Alignment:   "center",                // From sample
							LineSpacing: "-19",                   // From sample
						},
					},
				},
			}

			// Add the title to the background video
			backgroundVideo.NestedTitles = append(backgroundVideo.NestedTitles, titleElement)

			fmt.Printf("🎯 Added word '%s' blade %d/%d in lane %d at position (%d, %d) at time %.2fs\n", 
				word, bladeIndex+1, totalBlades, laneNum, currentX, currentY, bladeStartTime)
		}
	}

	return nil
}
