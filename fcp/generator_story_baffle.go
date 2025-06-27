package fcp

import (
	"fmt"
	"math/rand"
	"strings"
)

// StoryBaffleConfig defines parameters for the AI video creation story
type StoryBaffleConfig struct {
	Duration          float64 // Total story duration in seconds
	OutputDir         string  // Directory for downloaded images  
	PixabayAPIKey     string  // API key for image downloads
	MaxComplexity     float64 // How chaotic it gets (0.0-1.0)
	ImageCount        int     // Total images to download/use
	Format            string  // "horizontal" or "vertical"
}

// DefaultStoryBaffleConfig returns default configuration
func DefaultStoryBaffleConfig() *StoryBaffleConfig {
	return &StoryBaffleConfig{
		Duration:      300.0, // 5 minutes of chaos
		OutputDir:     "./story_baffle_assets",
		MaxComplexity: 0.95,  // Very chaotic
		ImageCount:    50,    // Lots of images
		Format:        "horizontal",
	}
}

// StoryPhase represents a phase in the AI's video creation journey
type StoryPhase struct {
	Name        string
	StartTime   float64
	Duration    float64
	Text        string
	Complexity  float64  // 0.0-1.0 how complex this phase is
	ImageTheme  string   // Theme for image search
	LaneCount   int      // How many lanes to use
	AnimationStyle string // Style of animation
}

// aiVideoStoryPhases defines the narrative arc of the AI trying to make a video
var aiVideoStoryPhases = []StoryPhase{
	{
		Name: "Introduction", StartTime: 0, Duration: 20,
		Text: "Let me make a simple video...", Complexity: 0.1, 
		ImageTheme: "computer", LaneCount: 1, AnimationStyle: "simple",
	},
	{
		Name: "FirstImage", StartTime: 20, Duration: 15,
		Text: "Just one image to start", Complexity: 0.2,
		ImageTheme: "nature", LaneCount: 1, AnimationStyle: "ken_burns",
	},
	{
		Name: "GettingIdeas", StartTime: 35, Duration: 20,
		Text: "Actually, let me add more images!", Complexity: 0.3,
		ImageTheme: "creativity", LaneCount: 2, AnimationStyle: "slide_in",
	},
	{
		Name: "Excitement", StartTime: 55, Duration: 25,
		Text: "This is fun! More effects!", Complexity: 0.5,
		ImageTheme: "celebration", LaneCount: 4, AnimationStyle: "fly_in",
	},
	{
		Name: "GoingOverboard", StartTime: 80, Duration: 30,
		Text: "MOAR LAYERS! MOAR EFFECTS!", Complexity: 0.7,
		ImageTheme: "explosion", LaneCount: 8, AnimationStyle: "chaos",
	},
	{
		Name: "PeakChaos", StartTime: 110, Duration: 40,
		Text: "I AM THE MASTER OF VIDEO!", Complexity: 0.95,
		ImageTheme: "fireworks", LaneCount: 15, AnimationStyle: "insanity",
	},
	{
		Name: "TheCollapse", StartTime: 150, Duration: 35,
		Text: "Oh no! Everything is falling!", Complexity: 0.8,
		ImageTheme: "falling", LaneCount: 12, AnimationStyle: "collapse",
	},
	{
		Name: "Disaster", StartTime: 185, Duration: 30,
		Text: "HELP! MY IMAGES ARE EVERYWHERE!", Complexity: 0.6,
		ImageTheme: "chaos", LaneCount: 8, AnimationStyle: "scatter",
	},
	{
		Name: "Realization", StartTime: 215, Duration: 25,
		Text: "Maybe I went a bit overboard...", Complexity: 0.4,
		ImageTheme: "thinking", LaneCount: 4, AnimationStyle: "settle",
	},
	{
		Name: "Recovery", StartTime: 240, Duration: 30,
		Text: "Let me clean this up...", Complexity: 0.2,
		ImageTheme: "cleaning", LaneCount: 2, AnimationStyle: "organize",
	},
	{
		Name: "Conclusion", StartTime: 270, Duration: 30,
		Text: "Well, that was educational!", Complexity: 0.1,
		ImageTheme: "learning", LaneCount: 1, AnimationStyle: "simple",
	},
}

// GenerateStoryBaffle creates an engaging AI video creation blooper reel
func GenerateStoryBaffle(config *StoryBaffleConfig, verbose bool) (*FCPXML, error) {
	if config == nil {
		config = DefaultStoryBaffleConfig()
	}

	if verbose {
		fmt.Printf("🤖 GENERATING AI VIDEO CREATION STORY-BAFFLE 🤖\n")
		fmt.Printf("Duration: %.1f minutes, Max Complexity: %.1f\n", config.Duration/60, config.MaxComplexity)
	}

	// Create base FCPXML
	fcpxml, err := GenerateEmptyWithFormat("", config.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to create base FCPXML: %v", err)
	}

	// Set up resource management
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Download images for all themes
	allImages, err := downloadThemeImages(config, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to download images: %v", err)
	}

	// Create text effect for story narration
	textEffectID := tx.ReserveIDs(1)[0]
	_, err = tx.CreateEffect(textEffectID, "StoryText", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
	if err != nil {
		return nil, fmt.Errorf("failed to create text effect: %v", err)
	}

	// Commit resources
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit resources: %v", err)
	}

	if verbose {
		fmt.Printf("Building story timeline with %d phases...\n", len(aiVideoStoryPhases))
	}

	// Build the story timeline
	if err := buildStoryBaffleTimeline(fcpxml, allImages, textEffectID, config, verbose); err != nil {
		return nil, fmt.Errorf("failed to build timeline: %v", err)
	}

	// Update sequence duration
	updateSequenceDuration(fcpxml, config.Duration)

	if verbose {
		fmt.Printf("✅ Story-Baffle generation completed!\n")
	}

	return fcpxml, nil
}

// downloadThemeImages downloads images for all the story themes
func downloadThemeImages(config *StoryBaffleConfig, verbose bool) (map[string][]ImageAttribution, error) {
	allImages := make(map[string][]ImageAttribution)
	
	// Get unique themes from story phases
	themes := make(map[string]bool)
	for _, phase := range aiVideoStoryPhases {
		themes[phase.ImageTheme] = true
	}

	imagesPerTheme := config.ImageCount / len(themes)
	if imagesPerTheme < 3 {
		imagesPerTheme = 3
	}

	for theme := range themes {
		if verbose {
			fmt.Printf("Downloading %d images for theme: %s\n", imagesPerTheme, theme)
		}
		
		images, err := DownloadImagesFromPixabay(theme, imagesPerTheme, config.OutputDir, config.PixabayAPIKey)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Failed to download images for theme %s: %v\n", theme, err)
			}
			// Continue with other themes
			continue
		}
		
		allImages[theme] = images
	}

	return allImages, nil
}

// buildStoryBaffleTimeline creates MICHAEL BAY INTENSITY timeline with rapid cuts and multiple lanes
func buildStoryBaffleTimeline(fcpxml *FCPXML, allImages map[string][]ImageAttribution, textEffectID string, config *StoryBaffleConfig, verbose bool) error {
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	
	if verbose {
		fmt.Printf("🎬 BUILDING MICHAEL BAY TIMELINE 🎬\n")
		fmt.Printf("Target: 8+ lanes, rapid 0.5-2s cuts, overlapping chaos\n")
	}

	// Step 1: Add background images that span entire timeline to prevent black screens
	err := addBackgroundImageLayer(fcpxml, allImages, config.Duration, verbose)
	if err != nil {
		return fmt.Errorf("failed to add background layer: %v", err)
	}

	// Step 2: Create primary spine elements with nested connected clips (like ultimate_baffle)
	numPrimaryElements := 5 + int(config.Duration/30) // More primary elements for longer videos
	primaryDuration := config.Duration / float64(numPrimaryElements) * 1.8 // Overlap them
	
	allImagesList := flattenImageMap(allImages)
	if len(allImagesList) == 0 {
		return fmt.Errorf("no images available for timeline")
	}
	
	imageIndex := 0
	
	for primaryIndex := 0; primaryIndex < numPrimaryElements; primaryIndex++ {
		primaryStartTime := float64(primaryIndex) * config.Duration / float64(numPrimaryElements)
		
		// Find the current story phase for this time
		currentPhase := getCurrentPhase(primaryStartTime)
		
		if verbose {
			fmt.Printf("Primary element %d: %.1fs-%.1fs, Phase: %s\n", 
				primaryIndex, primaryStartTime, primaryStartTime+primaryDuration, currentPhase.Name)
		}

		// Create primary spine element  
		primaryImage := allImagesList[imageIndex%len(allImagesList)]
		primaryClip, err := createPrimaryStoryElement(fcpxml, primaryImage.FilePath, 
			primaryStartTime, primaryDuration, primaryIndex, currentPhase, config.Format)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Failed to create primary element %d: %v\n", primaryIndex, err)
			}
			continue
		}
		imageIndex++

		// Add 8-15 nested connected clips with RAPID CUTS and multiple lanes
		numConnected := 8 + int(currentPhase.Complexity*7) // 8-15 connected clips
		
		for connectedIndex := 0; connectedIndex < numConnected && imageIndex < len(allImagesList); connectedIndex++ {
			// RAPID CUT TIMING - 0.5 to 2 seconds max!
			cutDuration := 0.5 + rand.Float64()*1.5 // 0.5-2.0 seconds
			cutStartTime := primaryStartTime + rand.Float64()*(primaryDuration-cutDuration)
			
			// Multiple lanes: -8 to +8 
			laneNumber := (connectedIndex % 17) - 8 // Lanes -8 to +8
			
			connectedImage := allImagesList[imageIndex%len(allImagesList)]
			
			err := addConnectedClipWithLane(fcpxml, primaryClip, connectedImage.FilePath, 
				cutStartTime-primaryStartTime, cutDuration, laneNumber, 
				currentPhase, connectedIndex, config.Format)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Failed to add connected clip %d: %v\n", connectedIndex, err)
				}
				continue
			}
			imageIndex++
		}

		// Add story text for this phase (if we're in a new phase)
		if primaryIndex < len(aiVideoStoryPhases) {
			phase := aiVideoStoryPhases[primaryIndex]
			err := addStoryBaffleText(fcpxml, textEffectID, phase.Text, 
				primaryStartTime, primaryDuration, primaryIndex)
			if err != nil && verbose {
				fmt.Printf("Warning: Failed to add text for phase %s: %v\n", phase.Name, err)
			}
		}

		spine.AssetClips = append(spine.AssetClips, *primaryClip)
	}

	// Step 3: Add rapid-fire standalone spine elements for maximum chaos
	err = addRapidFireSpineElements(fcpxml, allImagesList, config.Duration, imageIndex, verbose)
	if err != nil && verbose {
		fmt.Printf("Warning: Failed to add rapid-fire elements: %v\n", err)
	}

	if verbose {
		fmt.Printf("🎆 MICHAEL BAY TIMELINE COMPLETE 🎆\n")
		fmt.Printf("Primary elements: %d, Expected connected clips: %d+\n", 
			numPrimaryElements, numPrimaryElements*10)
	}

	return nil
}

// addStoryBaffleText adds narrative text with crazy fonts and animations
func addStoryBaffleText(fcpxml *FCPXML, effectID, text string, startTime, duration float64, phaseIndex int) error {
	// Use existing resource registry  
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Generate unique style ID
	styleID := GenerateTextStyleID(text, fmt.Sprintf("story_baffle_%d", phaseIndex))
	
	// Get wild fonts and colors
	fonts := GetRandomFonts()
	colors := GetRandomHighContrastColors()
	
	// Seed randomness with phase index for consistency
	rand.Seed(int64(phaseIndex * 1337))
	selectedFont := fonts[rand.Intn(len(fonts))]
	selectedColor := colors[rand.Intn(len(colors))]
	
	// Make font size vary with phase (bigger = more chaotic)
	baseFontSize := 150 + int(float64(phaseIndex)*20) // Gets bigger each phase
	if baseFontSize > 400 {
		baseFontSize = 400
	}

	if phaseIndex >= 4 { // From "GoingOverboard" onwards
		fmt.Printf("CHAOTIC TEXT: \"%s\" -> Font: %s (Size: %d)\n", strings.ToUpper(text), selectedFont, baseFontSize)
	} else {
		fmt.Printf("Text: \"%s\" -> Font: %s\n", text, selectedFont)
	}

	// Create title with animation (no lanes for spine elements)
	title := Title{
		Ref:      effectID,
		Lane:     "", // No lanes for spine elements
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("StoryText_%d", phaseIndex),
		Start:    "86486400/24000s",
		Text: &TitleText{
			TextStyles: []TextStyleRef{
				{
					Ref:  styleID,
					Text: text,
				},
			},
		},
		TextStyleDefs: []TextStyleDef{
			{
				ID: styleID,
				TextStyle: TextStyle{
					Font:        selectedFont,
					FontSize:    fmt.Sprintf("%d", baseFontSize),
					FontFace:    "Regular",
					FontColor:   selectedColor.FaceColor,
					StrokeColor: selectedColor.OutlineColor,
					StrokeWidth: "-12",
					Alignment:   "center",
					LineSpacing: "-15",
					Bold:        fmt.Sprintf("%d", phaseIndex%2),    // Some bold
					Italic:      fmt.Sprintf("%d", (phaseIndex+1)%2), // Some italic
				},
			},
		},
	}

	// Add crazy animations for later phases
	if phaseIndex >= 3 { // From "Excitement" onwards
		title.Params = createTextChaosAnimation(startTime, duration, phaseIndex)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit text transaction: %v", err)
	}

	// Add to spine
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Titles = append(spine.Titles, title)

	return nil
}

// addStoryBaffleImage adds images with animations that match the story phase
func addStoryBaffleImage(fcpxml *FCPXML, imagePath string, startTime, duration float64, 
	animationStyle string, complexity float64, imageIndex, phaseIndex int, format string) error {
	
	// Use transaction for asset creation
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Reserve IDs
	ids := tx.ReserveIDs(2)
	assetID := ids[0]
	formatID := ids[1]

	// Create format for image
	width := "1280"
	height := "720"
	if format == "vertical" {
		width = "1080"
		height = "1920"
	}

	_, err := tx.CreateFormat(formatID, "StoryBaffleImage", width, height, "1-13-1")
	if err != nil {
		return fmt.Errorf("failed to create format: %v", err)
	}

	// Create asset
	_, err = tx.CreateAsset(assetID, imagePath, fmt.Sprintf("StoryImage_%d_%d", phaseIndex, imageIndex), "0s", formatID)
	if err != nil {
		return fmt.Errorf("failed to create asset: %v", err)
	}

	// Commit resources
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Create video element (for image)
	video := Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("StoryImage_%s_%d_%d", animationStyle, phaseIndex, imageIndex),
	}

	// For story-baffle, we always want spine elements (no lanes for now)
	// Complex nesting will be added in a future iteration
	video.Lane = ""

	// Create animation based on style
	video.AdjustTransform = createStoryBaffleAnimation(animationStyle, startTime, duration, complexity, imageIndex, phaseIndex)

	// Add to spine
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	spine.Videos = append(spine.Videos, video)

	return nil
}

// createTextChaosAnimation creates wild text animations for chaotic phases
func createTextChaosAnimation(startTime, duration float64, phaseIndex int) []Param {
	var params []Param

	// Position animation - text flies around more wildly as phases progress
	positionKeyframes := make([]Keyframe, 5+phaseIndex) // More keyframes = more chaos
	for i := 0; i < len(positionKeyframes); i++ {
		keyTime := startTime + (float64(i)/float64(len(positionKeyframes)-1))*duration
		
		// Wilder movement in later phases
		maxMove := 200.0 + float64(phaseIndex)*100.0
		x := (rand.Float64()-0.5) * maxMove
		y := (rand.Float64()-0.5) * maxMove
		
		positionKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.1f %.1f", x, y),
		}
	}

	params = append(params, Param{
		Name: "Position",
		Key:  "9999/10003/13260/3296672360/1/100/101",
		KeyframeAnimation: &KeyframeAnimation{
			Keyframes: positionKeyframes,
		},
	})

	// Scale animation - text grows/shrinks more dramatically
	scaleKeyframes := []Keyframe{
		{
			Time:  ConvertSecondsToFCPDuration(startTime),
			Value: "0.5 0.5",
			Curve: "linear",
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration/2),
			Value: fmt.Sprintf("%.1f %.1f", 1.5+float64(phaseIndex)*0.3, 1.5+float64(phaseIndex)*0.3),
			Curve: "linear",
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration),
			Value: "1.0 1.0",
			Curve: "linear",
		},
	}

	params = append(params, Param{
		Name: "Scale",
		Key:  "9999/10003/13260/3296672360/1/100/102",
		KeyframeAnimation: &KeyframeAnimation{
			Keyframes: scaleKeyframes,
		},
	})

	// Rotation animation - spins more in chaotic phases
	rotationAmount := float64(phaseIndex) * 90.0 // Up to 540 degrees in final phases
	if rotationAmount > 720 {
		rotationAmount = 720
	}

	rotationKeyframes := []Keyframe{
		{
			Time:  ConvertSecondsToFCPDuration(startTime),
			Value: "0",
			Curve: "linear",
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration),
			Value: fmt.Sprintf("%.1f", rotationAmount),
			Curve: "linear",
		},
	}

	params = append(params, Param{
		Name: "Rotation",
		Key:  "9999/10003/13260/3296672360/1/100/103",
		KeyframeAnimation: &KeyframeAnimation{
			Keyframes: rotationKeyframes,
		},
	})

	return params
}

// createStoryBaffleAnimation creates animations that match the story narrative
func createStoryBaffleAnimation(style string, startTime, duration, complexity float64, imageIndex, phaseIndex int) *AdjustTransform {
	switch style {
	case "simple":
		return createSimpleAnimation(startTime, duration)
	case "ken_burns": 
		return createStoryKenBurnsAnimation(startTime, duration, imageIndex)
	case "slide_in":
		return createSlideInAnimation(startTime, duration, imageIndex)
	case "fly_in":
		return createFlyInAnimation(startTime, duration, float64(imageIndex), int(complexity*10))
	case "chaos", "insanity":
		return createChaosAnimation(startTime, duration, float64(imageIndex), int(complexity*10))
	case "collapse":
		return createCollapseAnimation(startTime, duration, imageIndex)
	case "scatter":
		return createScatterAnimation(startTime, duration, imageIndex)
	case "settle":
		return createSettleAnimation(startTime, duration, imageIndex)
	case "organize":
		return createOrganizeAnimation(startTime, duration, imageIndex)
	default:
		return createSimpleAnimation(startTime, duration)
	}
}

// Simple fade in animation
func createSimpleAnimation(startTime, duration float64) *AdjustTransform {
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "0.8 0.8",
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "1.2 1.2",
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// Ken Burns effect (slow zoom and pan) for story-baffle
func createStoryKenBurnsAnimation(startTime, duration float64, imageIndex int) *AdjustTransform {
	// Alternate zoom direction
	startScale := "1.0 1.0"
	endScale := "1.3 1.3"
	if imageIndex%2 == 1 {
		startScale = "1.3 1.3"
		endScale = "1.0 1.0"
	}
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: startScale,
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: endScale,
							Curve: "linear",
						},
					},
				},
			},
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.1f %.1f", float64((imageIndex%3-1)*50), float64((imageIndex%3-1)*30)),
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f %.1f", float64(((imageIndex+1)%3-1)*50), float64(((imageIndex+1)%3-1)*30)),
						},
					},
				},
			},
		},
	}
}

// Slide in from edges
func createSlideInAnimation(startTime, duration float64, imageIndex int) *AdjustTransform {
	// Different slide directions
	directions := []string{"-800 0", "800 0", "0 -600", "0 600"} // left, right, top, bottom
	startPos := directions[imageIndex%len(directions)]
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: startPos,
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration*0.3),
							Value: "0 0",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "0 0",
						},
					},
				},
			},
		},
	}
}

// Wild flying in from random directions
func createFlyInAnimation(startTime, duration, imageIndexFloat float64, complexityInt int) *AdjustTransform {
	// More extreme starting positions based on complexity
	complexity := float64(complexityInt) / 10.0
	maxDistance := 1000 + complexity*1000
	startX := (rand.Float64()-0.5) * maxDistance
	startY := (rand.Float64()-0.5) * maxDistance
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.1f %.1f", startX, startY),
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration*0.2),
							Value: "0 0",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f %.1f", (rand.Float64()-0.5)*200, (rand.Float64()-0.5)*200),
						},
					},
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.1f", (rand.Float64()-0.5)*720*complexity),
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "0",
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// Complete chaos - multiple keyframes, wild movement
func createChaosAnimation(startTime, duration, imageIndexFloat float64, complexityInt int) *AdjustTransform {
	complexity := float64(complexityInt) / 10.0
	keyframeCount := 8 + int(complexity*12) // 8-20 keyframes
	
	// Position chaos
	positionKeyframes := make([]Keyframe, keyframeCount)
	for i := 0; i < keyframeCount; i++ {
		keyTime := startTime + (float64(i)/float64(keyframeCount-1))*duration
		maxMove := 300 + complexity*700
		x := (rand.Float64()-0.5) * maxMove
		y := (rand.Float64()-0.5) * maxMove
		
		positionKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.1f %.1f", x, y),
		}
	}
	
	// Scale chaos
	scaleKeyframes := make([]Keyframe, keyframeCount/2)
	for i := 0; i < len(scaleKeyframes); i++ {
		keyTime := startTime + (float64(i)/float64(len(scaleKeyframes)-1))*duration
		scale := 0.5 + rand.Float64()*(1.0+complexity)
		
		scaleKeyframes[i] = Keyframe{
			Time:  ConvertSecondsToFCPDuration(keyTime),
			Value: fmt.Sprintf("%.2f %.2f", scale, scale),
			Curve: "linear",
		}
	}
	
	// Rotation chaos
	rotationKeyframes := []Keyframe{
		{
			Time:  ConvertSecondsToFCPDuration(startTime),
			Value: "0",
			Curve: "linear",
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration),
			Value: fmt.Sprintf("%.1f", (rand.Float64()-0.5)*1440*complexity), // Up to 4 full rotations
			Curve: "linear",
		},
	}
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: positionKeyframes,
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: scaleKeyframes,
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: rotationKeyframes,
				},
			},
		},
	}
}

// Images fall down like they're collapsing
func createCollapseAnimation(startTime, duration float64, imageIndex int) *AdjustTransform {
	fallDistance := 800 + float64(imageIndex)*100 // Each image falls further
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.1f %.1f", float64((imageIndex%7-3)*100), 0.0),
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f %.1f", float64((imageIndex%7-3)*150), fallDistance),
						},
					},
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "0",
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f", float64(imageIndex%4)*90), // Different rotation amounts
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// Images scatter in all directions
func createScatterAnimation(startTime, duration float64, imageIndex int) *AdjustTransform {
	// Scatter to random directions
	endX := (rand.Float64()-0.5) * 1500
	endY := (rand.Float64()-0.5) * 1000
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "0 0",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f %.1f", endX, endY),
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "1.0 1.0",
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "0.3 0.3",
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// Images settle back to normal positions
func createSettleAnimation(startTime, duration float64, imageIndex int) *AdjustTransform {
	// Start from scattered position, settle to grid
	startX := (rand.Float64()-0.5) * 600
	startY := (rand.Float64()-0.5) * 400
	endX := float64((imageIndex%5-2) * 100) // Grid positions
	endY := float64((imageIndex/5%3-1) * 100)
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.1f %.1f", startX, startY),
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f %.1f", endX, endY),
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "0.5 0.5",
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "1.0 1.0",
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// Images organize into neat arrangement
func createOrganizeAnimation(startTime, duration float64, imageIndex int) *AdjustTransform {
	// Move to organized grid positions
	gridX := float64((imageIndex%4-2) * 150) // 4 columns
	gridY := float64((imageIndex/4%3-1) * 120) // 3 rows
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: fmt.Sprintf("%.1f %.1f", (rand.Float64()-0.5)*400, (rand.Float64()-0.5)*300),
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration*0.8),
							Value: fmt.Sprintf("%.1f %.1f", gridX, gridY),
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: fmt.Sprintf("%.1f %.1f", gridX, gridY),
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "0.7 0.7",
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "0.8 0.8",
							Curve: "linear",
						},
					},
				},
			},
		},
	}
}

// addBackgroundImageLayer adds continuous background images to prevent black screens
func addBackgroundImageLayer(fcpxml *FCPXML, allImages map[string][]ImageAttribution, totalDuration float64, verbose bool) error {
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	
	// Get any available images for background
	var backgroundImages []ImageAttribution
	for _, images := range allImages {
		backgroundImages = append(backgroundImages, images...)
		if len(backgroundImages) >= 5 {
			break
		}
	}
	
	if len(backgroundImages) == 0 {
		return fmt.Errorf("no images available for background")
	}
	
	// Create 2-3 overlapping background videos that span entire timeline
	numBackgrounds := 2 + int(totalDuration/60) // More backgrounds for longer videos
	
	for i := 0; i < numBackgrounds && i < len(backgroundImages); i++ {
		imageAttr := backgroundImages[i]
		
		startTime := float64(i) * 5.0 // Stagger starts slightly
		duration := totalDuration + 10.0 // Extend past end
		
		video, err := createBackgroundVideo(fcpxml, imageAttr.FilePath, startTime, duration, i)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Failed to create background video %d: %v\n", i, err)
			}
			continue
		}
		
		spine.Videos = append(spine.Videos, *video)
	}
	
	if verbose {
		fmt.Printf("Added %d background layers spanning %.1fs\n", numBackgrounds, totalDuration)
	}
	
	return nil
}

// flattenImageMap converts the theme-based image map to a flat list
func flattenImageMap(allImages map[string][]ImageAttribution) []ImageAttribution {
	var result []ImageAttribution
	for _, images := range allImages {
		result = append(result, images...)
	}
	return result
}

// getCurrentPhase finds which story phase is active at a given time
func getCurrentPhase(time float64) StoryPhase {
	for _, phase := range aiVideoStoryPhases {
		if time >= phase.StartTime && time < phase.StartTime+phase.Duration {
			return phase
		}
	}
	// Default to last phase if beyond timeline
	return aiVideoStoryPhases[len(aiVideoStoryPhases)-1]
}

// createBackgroundVideo creates a simple background video element
func createBackgroundVideo(fcpxml *FCPXML, imagePath string, startTime, duration float64, index int) (*Video, error) {
	// Use transaction for proper asset creation
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Reserve IDs
	ids := tx.ReserveIDs(2)
	assetID := ids[0]
	formatID := ids[1]

	// Create format for image
	_, err := tx.CreateFormat(formatID, "BackgroundImage", "1280", "720", "1-13-1")
	if err != nil {
		return nil, fmt.Errorf("failed to create background format: %v", err)
	}

	// Create asset
	_, err = tx.CreateAsset(assetID, imagePath, fmt.Sprintf("Background_%d", index), "0s", formatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create background asset: %v", err)
	}

	// Commit resources
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit background transaction: %v", err)
	}

	// Create video element with very subtle animation
	video := &Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Background_%d", index),
		Start:    "0s",
		Lane:     "", // Spine element
	}

	// Very subtle scale animation to keep it interesting
	video.AdjustTransform = &AdjustTransform{
		Params: []Param{
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  ConvertSecondsToFCPDuration(startTime),
							Value: "1.0 1.0",
							Curve: "linear",
						},
						{
							Time:  ConvertSecondsToFCPDuration(startTime + duration),
							Value: "1.1 1.1",
							Curve: "linear",
						},
					},
				},
			},
		},
	}

	return video, nil
}

// createPrimaryStoryElement creates a primary spine asset-clip that will contain nested clips
func createPrimaryStoryElement(fcpxml *FCPXML, imagePath string, startTime, duration float64, index int, phase StoryPhase, format string) (*AssetClip, error) {
	// Use transaction for proper asset creation
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Reserve IDs
	ids := tx.ReserveIDs(2)
	assetID := ids[0]
	formatID := ids[1]

	// Create format for image
	width := "1280"
	height := "720"
	if format == "vertical" {
		width = "1080"
		height = "1920"
	}

	_, err := tx.CreateFormat(formatID, "PrimaryStoryImage", width, height, "1-13-1")
	if err != nil {
		return nil, fmt.Errorf("failed to create primary format: %v", err)
	}

	// Create asset
	_, err = tx.CreateAsset(assetID, imagePath, fmt.Sprintf("Primary_%s_%d", phase.Name, index), "0s", formatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary asset: %v", err)
	}

	// Commit resources
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit primary transaction: %v", err)
	}

	// Create asset-clip (NOT video - we need asset-clip for nested connected clips)
	clip := &AssetClip{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Primary_%s_%d", phase.Name, index),
		Start:    "0s",
		Lane:     "", // Primary spine element has no lane
	}

	// Add primary animation based on phase
	clip.AdjustTransform = createStoryBaffleAnimation(phase.AnimationStyle, startTime, duration, phase.Complexity, index, index)

	return clip, nil
}

// addConnectedClipWithLane adds a connected clip with a specific lane to a primary element
func addConnectedClipWithLane(fcpxml *FCPXML, primaryClip *AssetClip, imagePath string, offsetFromPrimary, duration float64, laneNumber int, phase StoryPhase, index int, format string) error {
	// Create asset for connected clip
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Reserve IDs
	ids := tx.ReserveIDs(2)
	assetID := ids[0]
	formatID := ids[1]

	// Create format for image
	width := "1280"
	height := "720"
	if format == "vertical" {
		width = "1080"
		height = "1920"
	}

	_, err := tx.CreateFormat(formatID, "ConnectedImage", width, height, "1-13-1")
	if err != nil {
		return fmt.Errorf("failed to create connected format: %v", err)
	}

	// Create asset
	_, err = tx.CreateAsset(assetID, imagePath, fmt.Sprintf("Connected_%s_L%d_%d", phase.Name, laneNumber, index), "0s", formatID)
	if err != nil {
		return fmt.Errorf("failed to create connected asset: %v", err)
	}

	// Commit resources
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit connected transaction: %v", err)
	}

	// Create connected video element (nested inside the primary asset-clip)
	connectedVideo := Video{
		Ref:      assetID, // Use real asset ID
		Offset:   ConvertSecondsToFCPDuration(offsetFromPrimary),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("Connected_%s_L%d_%d", phase.Name, laneNumber, index),
		Start:    "0s",
		Lane:     fmt.Sprintf("%d", laneNumber), // This is where the lane magic happens!
	}

	// Create explosive animation for connected clip
	connectedVideo.AdjustTransform = createExplosiveAnimation(offsetFromPrimary, duration, phase.Complexity, index, laneNumber)

	// Add to primary clip's nested videos
	primaryClip.Videos = append(primaryClip.Videos, connectedVideo)

	return nil
}

// addRapidFireSpineElements adds additional standalone spine elements for maximum chaos
func addRapidFireSpineElements(fcpxml *FCPXML, allImages []ImageAttribution, totalDuration float64, startIndex int, verbose bool) error {
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine
	
	// Add 10-20 rapid-fire standalone videos throughout the timeline
	numRapidFire := 10 + int(totalDuration/10) // More for longer videos
	
	for i := 0; i < numRapidFire && startIndex+i < len(allImages); i++ {
		imageAttr := allImages[startIndex+i]
		
		// Random timing throughout the video
		startTime := rand.Float64() * totalDuration
		duration := 0.8 + rand.Float64()*1.7 // 0.8-2.5 seconds
		
		video, err := createRapidFireVideo(fcpxml, imageAttr.FilePath, startTime, duration, i)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Failed to create rapid-fire video %d: %v\n", i, err)
			}
			continue
		}
		
		spine.Videos = append(spine.Videos, *video)
	}
	
	if verbose {
		fmt.Printf("Added %d rapid-fire spine elements\n", numRapidFire)
	}
	
	return nil
}

// createRapidFireVideo creates a standalone spine video with explosive animation
func createRapidFireVideo(fcpxml *FCPXML, imagePath string, startTime, duration float64, index int) (*Video, error) {
	// Use transaction for proper asset creation
	registry := NewResourceRegistry(fcpxml)
	tx := NewTransaction(registry)
	defer tx.Rollback()

	// Reserve IDs
	ids := tx.ReserveIDs(2)
	assetID := ids[0]
	formatID := ids[1]

	// Create format for image
	_, err := tx.CreateFormat(formatID, "RapidFireImage", "1280", "720", "1-13-1")
	if err != nil {
		return nil, fmt.Errorf("failed to create rapid-fire format: %v", err)
	}

	// Create asset
	_, err = tx.CreateAsset(assetID, imagePath, fmt.Sprintf("RapidFire_%d", index), "0s", formatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create rapid-fire asset: %v", err)
	}

	// Commit resources
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit rapid-fire transaction: %v", err)
	}

	// Create video element
	video := &Video{
		Ref:      assetID,
		Offset:   ConvertSecondsToFCPDuration(startTime),
		Duration: ConvertSecondsToFCPDuration(duration),
		Name:     fmt.Sprintf("RapidFire_%d", index),
		Start:    "0s",
		Lane:     "", // Spine element
	}

	// Create rapid-fire explosive animation
	video.AdjustTransform = createExplosiveAnimation(startTime, duration, 0.9, index, 0)

	return video, nil
}

// createExplosiveAnimation creates intense Michael Bay-style animations
func createExplosiveAnimation(startTime, duration, complexity float64, imageIndex, laneNumber int) *AdjustTransform {
	// Seed randomness based on index for consistent but varied animations
	randSeed := int64(imageIndex*1337 + laneNumber*456)
	localRand := rand.New(rand.NewSource(randSeed))
	
	// EXPLOSIVE STARTING POSITION - fly in from way off screen
	startDistance := 1500 + complexity*1000 // 1500-2500 pixels off screen
	startX := startDistance * localRand.Float64() * 2 - startDistance // -startDistance to +startDistance
	startY := startDistance * localRand.Float64() * 2 - startDistance
	
	// Target position (slightly off center for chaos)
	endX := (localRand.Float64() - 0.5) * 400 // -200 to +200
	endY := (localRand.Float64() - 0.5) * 300 // -150 to +150
	
	// RAPID POSITION ANIMATION
	positionKeyframes := []Keyframe{
		{
			Time:  ConvertSecondsToFCPDuration(startTime),
			Value: fmt.Sprintf("%.1f %.1f", startX, startY),
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration*0.2), // Hit target fast
			Value: fmt.Sprintf("%.1f %.1f", endX, endY),
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration),
			Value: fmt.Sprintf("%.1f %.1f", endX+(localRand.Float64()-0.5)*100, endY+(localRand.Float64()-0.5)*100),
		},
	}
	
	// EXPLOSIVE SCALE ANIMATION
	scaleKeyframes := []Keyframe{
		{
			Time:  ConvertSecondsToFCPDuration(startTime),
			Value: "0.1 0.1", // Start tiny
			Curve: "linear",
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration*0.3),
			Value: fmt.Sprintf("%.1f %.1f", 1.5+complexity*0.5, 1.5+complexity*0.5), // EXPLODE bigger
			Curve: "linear",
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration),
			Value: fmt.Sprintf("%.1f %.1f", 0.8+localRand.Float64()*0.4, 0.8+localRand.Float64()*0.4),
			Curve: "linear",
		},
	}
	
	// CRAZY ROTATION
	rotationAmount := (localRand.Float64() - 0.5) * 720 * (1 + complexity) // Up to ±720 degrees * complexity
	rotationKeyframes := []Keyframe{
		{
			Time:  ConvertSecondsToFCPDuration(startTime),
			Value: fmt.Sprintf("%.1f", (localRand.Float64()-0.5)*180), // Random start rotation
			Curve: "linear",
		},
		{
			Time:  ConvertSecondsToFCPDuration(startTime + duration),
			Value: fmt.Sprintf("%.1f", rotationAmount),
			Curve: "linear",
		},
	}
	
	return &AdjustTransform{
		Params: []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: positionKeyframes,
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: scaleKeyframes,
				},
			},
			{
				Name: "rotation",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: rotationKeyframes,
				},
			},
		},
	}
}
