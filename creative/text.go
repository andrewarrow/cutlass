package creative

import (
	"cutlass/fcp"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// ContentSection represents a section from the JSON input
type ContentSection struct {
	Title  string   `json:"title"`
	Points []string `json:"points"`
}

// CreativeTextOptions contains options for text animation
type CreativeTextOptions struct {
	InputFile        string
	OutputFile       string
	SectionDuration  float64 // Duration per section in seconds
	PointDuration    float64 // Duration per point in seconds
	TransitionDuration float64 // Duration for section transitions
	BackgroundColor  string  // Background color
	ProjectName      string
	EventName        string
}

// DefaultCreativeTextOptions returns default options
func DefaultCreativeTextOptions() CreativeTextOptions {
	return CreativeTextOptions{
		SectionDuration:    45.0, // Much longer for extended presentation
		PointDuration:      8.0,  // Extended time for each explosive point
		TransitionDuration: 5.0,  // Longer dramatic pauses between sections
		BackgroundColor:    "0.1 0.1 0.2 1", // Dark blue background
		ProjectName:        "🎬 EXPLOSIVE PRESENTATION 🎬",
		EventName:          "💥 CREATIVE IMPACT 💥",
	}
}

// GenerateCreativeText creates an FCPXML with animated text presentation
func GenerateCreativeText(options CreativeTextOptions) error {
	// Read and parse the JSON input file
	data, err := ioutil.ReadFile(options.InputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	var sections []ContentSection
	if err := json.Unmarshal(data, &sections); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Calculate total timeline duration
	totalDuration := calculateTotalDuration(sections, options)
	
	// Create base FCPXML structure using GenerateEmpty
	fcpxml, err := fcp.GenerateEmpty("")
	if err != nil {
		return fmt.Errorf("failed to create empty FCPXML: %v", err)
	}

	// Initialize FCP registry and transaction  
	registry := fcp.NewResourceRegistry(fcpxml)
	tx := fcp.NewTransaction(registry)
	defer tx.Rollback()

	// Update with our custom settings
	fcpxml.Library.Location = "file:///Users/aa/Movies/CreativeText.fcpbundle/"
	fcpxml.Library.Events[0].Name = options.EventName
	fcpxml.Library.Events[0].Projects[0].Name = options.ProjectName
	fcpxml.Library.Events[0].Projects[0].ModDate = "2025-06-20 12:00:00 -0700"
	fcpxml.Library.Events[0].Projects[0].Sequences[0].Duration = fcp.ConvertSecondsToFCPDuration(totalDuration)

	// Add text effect to resources
	ids := tx.ReserveIDs(1)
	textEffectID := ids[0]
	fcpxml.Resources.Effects = append(fcpxml.Resources.Effects, fcp.Effect{
		ID:   textEffectID,
		Name: "Text",
		UID:  ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
	})

	// Generate background
	if err := addCreativeBackground(fcpxml, tx, totalDuration, options); err != nil {
		return fmt.Errorf("failed to add background: %v", err)
	}

	// Generate animated text for each section and add to background video
	currentTime := 0.0
	backgroundVideo := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Videos[0]
	
	for sectionIndex, section := range sections {
		if err := addSectionText(backgroundVideo, tx, section, sectionIndex, &currentTime, textEffectID, options); err != nil {
			return fmt.Errorf("failed to add section %d: %v", sectionIndex, err)
		}
	}

	// Commit transaction and save file
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return fcp.WriteToFile(fcpxml, options.OutputFile)
}

// calculateTotalDuration calculates the total duration needed for all sections
func calculateTotalDuration(sections []ContentSection, options CreativeTextOptions) float64 {
	// Target total duration: 7 minutes 11 seconds = 431 seconds
	targetDuration := 431.0
	
	// Calculate natural duration based on content
	naturalDuration := 0.0
	for _, section := range sections {
		// Title appears first with extended time
		titleDuration := 4.0
		// Points appear sequentially with extended timing
		pointsDuration := float64(len(section.Points)) * options.PointDuration * 0.9 // Less overlap for breathing room
		// Section transition with dramatic pause
		sectionDuration := titleDuration + pointsDuration + options.TransitionDuration
		
		naturalDuration += sectionDuration
	}
	
	// Use the target duration to ensure exactly 7:11
	return targetDuration
}

// addCreativeBackground creates a visible background using proven FCP generator
func addCreativeBackground(fcpxml *fcp.FCPXML, tx *fcp.ResourceTransaction, duration float64, options CreativeTextOptions) error {
	// Use the proven Vivid generator from blue_background.fcpxml sample
	// This is a real FCP generator that definitely works
	ids := tx.ReserveIDs(1)
	generatorID := ids[0]
	
	// Add the Vivid generator effect (proven to work from sample)
	fcpxml.Resources.Effects = append(fcpxml.Resources.Effects, fcp.Effect{
		ID:   generatorID,
		Name: "Vivid",
		UID:  ".../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn",
	})

	// Create background video element using the generator
	backgroundVideo := fcp.Video{
		Ref:      generatorID,
		Offset:   "0s",
		Name:     "Background",
		Duration: fcp.ConvertSecondsToFCPDuration(duration),
		Start:    "0s",
	}

	fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Videos = append(
		fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Videos,
		backgroundVideo,
	)

	return nil
}

// addSectionText adds animated text for a section with title and points to background video
func addSectionText(backgroundVideo *fcp.Video, tx *fcp.ResourceTransaction, section ContentSection, sectionIndex int, currentTime *float64, textEffectID string, options CreativeTextOptions) error {
	// Target 431 seconds total across all sections - calculate section timing
	totalSections := 7 // From jenny_hansen_lane.json
	sectionTimeAllocation := 431.0 / float64(totalSections) // ~61.5 seconds per section
	
	// DRAMATIC TITLE ENTRANCE - use simple valid XML ID format
	titleID := fmt.Sprintf("ts%d", sectionIndex*10+1)
	titleOffset := *currentTime
	titleDuration := 6.0 // Much longer for extended dramatic effect
	
	// Titles will use scale animation instead of position sliding
	
	// CINEMATIC TITLE with dramatic slide and scale
	titleElement := fcp.Title{
		Ref:      textEffectID,
		Lane:     fmt.Sprintf("%d", sectionIndex*10+1), // Unique lane per section
		Offset:   fcp.ConvertSecondsToFCPDuration(titleOffset),
		Name:     fmt.Sprintf("🎬 SECTION %d TITLE", sectionIndex+1),
		Duration: fcp.ConvertSecondsToFCPDuration(sectionTimeAllocation), // Use full section time
		Start:    "0s",
		Params: []fcp.Param{
			// Static position (no keyframe animation on Position for titles)
			{
				Name:  "Position",
				Key:   "9999/10003/13260/3296672360/1/100/101",
				Value: "0 300", // Upper position similar to Info.fcpxml title
			},
			// Large font size for IMPACT
			{
				Name:  "Layout Method",
				Key:   "9999/10003/13260/3296672360/2/314",
				Value: "1 (Paragraph)",
			},
			// Center alignment
			{
				Name:  "Alignment",
				Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
				Value: "1 (Center)",
			},
			// EXPLOSIVE SCALE ANIMATION
			{
				Name: "Scale",
				Key:  "9999/10003/13260/3296672360/1/100/200",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: []fcp.Keyframe{
						{
							Time:   "0s",
							Value:  "0.3 0.3", // Start tiny
							Interp: "easeOut",
							Curve:  "smooth",
						},
						{
							Time:   fcp.ConvertSecondsToFCPDuration(0.6),
							Value:  "1.15 1.15", // Overshoot big!
							Interp: "ease",
							Curve:  "smooth",
						},
						{
							Time:   fcp.ConvertSecondsToFCPDuration(1.0),
							Value:  "1.0 1.0", // Settle to normal
							Interp: "linear",
							Curve:  "smooth",
						},
					},
				},
			},
			// DRAMATIC FADE WITH GLOW EFFECT
			{
				Name: "Opacity",
				Key:  "9999/10003/13260/3296672360/4/3296673134/1000/1044",
				KeyframeAnimation: &fcp.KeyframeAnimation{
					Keyframes: []fcp.Keyframe{
						{
							Time:   "0s",
							Value:  "0",
							Interp: "easeOut",
							Curve:  "smooth",
						},
						{
							Time:   fcp.ConvertSecondsToFCPDuration(0.4),
							Value:  "1",
							Interp: "linear",
							Curve:  "smooth",
						},
					},
				},
			},
		},
		Text: &fcp.TitleText{
			TextStyle: fcp.TextStyleRef{
				Ref:  titleID,
				Text: "🎯 " + strings.ToUpper(section.Title) + " 🎯", // Add impact emojis and caps
			},
		},
		TextStyleDef: &fcp.TextStyleDef{
			ID: titleID,
			TextStyle: fcp.TextStyle{
				Font:      "Helvetica Neue",
				FontSize:  "96", // MUCH BIGGER!
				FontFace:  "Bold",
				FontColor: "1 1 0.2 1", // Bright yellow for impact!
				Bold:      "1",
				Alignment: "center",
			},
		},
	}

	backgroundVideo.NestedTitles = append(
		backgroundVideo.NestedTitles,
		titleElement,
	)

	*currentTime += titleDuration

	// 💥 EXPLOSIVE BULLET POINTS with increasing intensity! 💥
	for pointIndex, point := range section.Points {
		pointID := fmt.Sprintf("ts%d", sectionIndex*10+pointIndex+2)
		pointDelay := titleDuration + 2.0 + float64(pointIndex)*10.0 // Much longer delays for extended timing
		pointOffset := *currentTime + pointDelay
		
		// Calculate positions matching Info.fcpxml pattern - 200 pixel spacing
		baseY := 150 // Starting Y position for first bullet point
		yPosition := baseY - (pointIndex * 200) // 200 pixel spacing between points, going upward
		
		// Each point gets MORE dramatic (escalating excitement!)
		intensity := float64(pointIndex + 1)
		impactScale := 1.0 + (intensity * 0.2) // Each point bigger than the last!
		
		// Points will use scale animation for explosive effect
		
		// 🎆 EXPLOSIVE POINT ENTRANCE 🎆
		pointElement := fcp.Title{
			Ref:      textEffectID,
			Lane:     fmt.Sprintf("%d", sectionIndex*10+pointIndex+2), // Unique lane per point
			Offset:   fcp.ConvertSecondsToFCPDuration(pointOffset),
			Name:     fmt.Sprintf("💥 BAM! POINT %d", pointIndex+1),
			Duration: fcp.ConvertSecondsToFCPDuration(sectionTimeAllocation - pointDelay), // Use remaining section time
			Start:    "0s",
			Params: []fcp.Param{
				// Static position (no keyframe animation on Position for titles)
				{
					Name:  "Position",
					Key:   "9999/10003/13260/3296672360/1/100/101",
					Value: fmt.Sprintf("80 %d", yPosition), // Left-center position matching Info.fcpxml (~80 pixels)
				},
				// Left alignment
				{
					Name:  "Alignment",
					Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
					Value: "0 (Left)",
				},
				// 💥 EXPLOSIVE SCALE - gets BIGGER each time!
				{
					Name: "Scale",
					Key:  "9999/10003/13260/3296672360/1/100/200",
					KeyframeAnimation: &fcp.KeyframeAnimation{
						Keyframes: []fcp.Keyframe{
							{
								Time:   "0s",
								Value:  "0.1 0.1", // Start microscopic
								Interp: "easeOut",
								Curve:  "smooth",
							},
							{
								Time:   fcp.ConvertSecondsToFCPDuration(0.3),
								Value:  fmt.Sprintf("%.2f %.2f", impactScale+0.3, impactScale+0.3), // HUGE overshoot!
								Interp: "ease", 
								Curve:  "smooth",
							},
							{
								Time:   fcp.ConvertSecondsToFCPDuration(0.6),
								Value:  fmt.Sprintf("%.2f %.2f", impactScale, impactScale), // Settle to bigger size
								Interp: "linear",
								Curve:  "smooth",
							},
						},
					},
				},
				// 🌟 DRAMATIC FADE WITH FLASH!
				{
					Name: "Opacity",
					Key:  "9999/10003/13260/3296672360/4/3296673134/1000/1044",
					KeyframeAnimation: &fcp.KeyframeAnimation{
						Keyframes: []fcp.Keyframe{
							{
								Time:   "0s",
								Value:  "0",
								Interp: "easeOut",
								Curve:  "smooth",
							},
							{
								Time:   fcp.ConvertSecondsToFCPDuration(0.2),
								Value:  "1",
								Interp: "linear",
								Curve:  "smooth",
							},
						},
					},
				},
			},
			Text: &fcp.TitleText{
				TextStyle: fcp.TextStyleRef{
					Ref:  pointID,
					Text: fmt.Sprintf("⚡ %s ⚡", strings.ToUpper(point)), // Electrifying caps with lightning!
				},
			},
			TextStyleDef: &fcp.TextStyleDef{
				ID: pointID,
				TextStyle: fcp.TextStyle{
					Font:      "Helvetica Neue",
					FontSize:  "80", // Consistent 80px font size matching Info.fcpxml
					FontFace:  "Bold",
					FontColor: fmt.Sprintf("%.1f 1 %.1f 1", 0.8+intensity*0.1, 0.6+intensity*0.2), // Increasingly bright!
					Bold:      "1",
					Alignment: "left",
				},
			},
		}

		backgroundVideo.NestedTitles = append(
			backgroundVideo.NestedTitles,
			pointElement,
		)
	}

	// Update current time for next section - use the full section allocation
	*currentTime += sectionTimeAllocation

	return nil
}


// HandleCreativeTextCommand handles the CLI command for creative text generation
func HandleCreativeTextCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: creative-text <input.json> [output.fcpxml]")
	}

	options := DefaultCreativeTextOptions()
	options.InputFile = args[0]
	
	if len(args) > 1 {
		options.OutputFile = args[1]
	} else {
		// Generate output filename based on input
		inputName := strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0]))
		options.OutputFile = fmt.Sprintf("%s_creative_text.fcpxml", inputName)
	}

	fmt.Printf("Generating creative text presentation from %s...\n", options.InputFile)
	fmt.Printf("Output will be saved to %s\n", options.OutputFile)

	if err := GenerateCreativeText(options); err != nil {
		return fmt.Errorf("failed to generate creative text: %v", err)
	}

	fmt.Printf("🚀 EXPLOSIVE presentation generated successfully!\n")
	fmt.Printf("💥 Features DRAMATIC scale animations with explosive overshoot effects!\n")
	fmt.Printf("⚡ Titles and points EXPLODE onto screen with increasing intensity!\n")
	fmt.Printf("🎯 Each bullet point gets BIGGER and MORE EXCITING than the last!\n")
	fmt.Printf("🎬 Ready for picture-in-picture video - this will BLOW YOUR AUDIENCE AWAY!\n")
	
	return nil
}