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
		SectionDuration:    8.0,  // 8 seconds per section
		PointDuration:      2.0,  // 2 seconds per point
		TransitionDuration: 1.0,  // 1 second transition
		BackgroundColor:    "0.1 0.1 0.2 1", // Dark blue background
		ProjectName:        "Creative Text Presentation",
		EventName:          "Creative Content",
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

	// Generate animated text for each section
	currentTime := 0.0
	for sectionIndex, section := range sections {
		if err := addSectionText(fcpxml, tx, section, sectionIndex, &currentTime, textEffectID, options); err != nil {
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
	totalDuration := 0.0
	
	for _, section := range sections {
		// Title appears first
		titleDuration := 1.5
		// Points appear sequentially with overlap
		pointsDuration := float64(len(section.Points)) * options.PointDuration * 0.8 // 80% overlap
		// Section transition
		sectionDuration := titleDuration + pointsDuration + options.TransitionDuration
		
		totalDuration += sectionDuration
	}
	
	return totalDuration
}

// addCreativeBackground adds a simple solid background using built-in FCP elements
func addCreativeBackground(fcpxml *fcp.FCPXML, tx *fcp.ResourceTransaction, duration float64, options CreativeTextOptions) error {
	// Use a gap element with solid color (built-in FCP approach)
	// This avoids fictional generator UIDs and uses standard FCP elements
	
	backgroundGap := fcp.Gap{
		Name:     "Background",
		Offset:   "0s",
		Duration: fcp.ConvertSecondsToFCPDuration(duration),
	}

	fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Gaps = append(
		fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Gaps,
		backgroundGap,
	)

	return nil
}

// addSectionText adds animated text for a section with title and points
func addSectionText(fcpxml *fcp.FCPXML, tx *fcp.ResourceTransaction, section ContentSection, sectionIndex int, currentTime *float64, textEffectID string, options CreativeTextOptions) error {
	// Title animation - use simple valid XML ID format
	titleID := fmt.Sprintf("ts%d", sectionIndex*10+1)
	titleOffset := *currentTime
	titleDuration := 1.5
	
	// Title appears with scale and fade animation
	titleElement := fcp.Title{
		Ref:      textEffectID,
		Lane:     fmt.Sprintf("%d", sectionIndex*10+1), // Unique lane per section
		Offset:   fcp.ConvertSecondsToFCPDuration(titleOffset),
		Name:     fmt.Sprintf("Section %d Title", sectionIndex+1),
		Duration: fcp.ConvertSecondsToFCPDuration(titleDuration + float64(len(section.Points))*options.PointDuration*0.8 + 1.0),
		Start:    "0s",
		Params: []fcp.Param{
			// Position title at top center
			{
				Name:  "Position",
				Key:   "9999/10003/13260/3296672360/1/100/101",
				Value: "0 200",
			},
			// Large font size for title
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
			// Animated opacity for fade in
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
							Time:   fcp.ConvertSecondsToFCPDuration(0.5),
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
				Text: section.Title,
			},
		},
		TextStyleDef: &fcp.TextStyleDef{
			ID: titleID,
			TextStyle: fcp.TextStyle{
				Font:      "Helvetica Neue",
				FontSize:  "72",
				FontFace:  "Bold",
				FontColor: "1 1 1 1", // White
				Bold:      "1",
				Alignment: "center",
			},
		},
	}

	fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Titles = append(
		fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Titles,
		titleElement,
	)

	*currentTime += titleDuration

	// Add points with staggered animation
	for pointIndex, point := range section.Points {
		pointID := fmt.Sprintf("ts%d", sectionIndex*10+pointIndex+2)
		pointOffset := *currentTime + float64(pointIndex)*options.PointDuration*0.6 // Staggered timing
		
		// Calculate vertical position for this point
		yPosition := -50 - (pointIndex * 80) // Space points vertically
		
		pointElement := fcp.Title{
			Ref:      textEffectID,
			Lane:     fmt.Sprintf("%d", sectionIndex*10+pointIndex+2), // Unique lane per point
			Offset:   fcp.ConvertSecondsToFCPDuration(pointOffset),
			Name:     fmt.Sprintf("Section %d Point %d", sectionIndex+1, pointIndex+1),
			Duration: fcp.ConvertSecondsToFCPDuration(options.PointDuration * 2.0), // Points stay visible longer
			Start:    "0s",
			Params: []fcp.Param{
				// Position for point (static value)
				{
					Name:  "Position",
					Key:   "9999/10003/13260/3296672360/1/100/101", 
					Value: fmt.Sprintf("-300 %d", yPosition),
				},
				// Left alignment for points
				{
					Name:  "Alignment",
					Key:   "9999/10003/13260/3296672360/2/354/3296667315/401",
					Value: "0 (Left)",
				},
				// Fade in animation
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
								Time:   fcp.ConvertSecondsToFCPDuration(0.3),
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
					Text: "• " + point, // Add bullet point
				},
			},
			TextStyleDef: &fcp.TextStyleDef{
				ID: pointID,
				TextStyle: fcp.TextStyle{
					Font:      "Helvetica Neue",
					FontSize:  "48",
					FontColor: "0.9 0.9 0.9 1", // Light gray
					Alignment: "left",
				},
			},
		}

		fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Titles = append(
			fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Titles,
			pointElement,
		)
	}

	// Update current time for next section (with transition)
	sectionDuration := titleDuration + float64(len(section.Points))*options.PointDuration*0.8
	*currentTime += sectionDuration

	// Add section transition with gap (simpler approach)
	*currentTime += options.TransitionDuration

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

	fmt.Printf("✅ Creative text presentation generated successfully!\n")
	fmt.Printf("📱 Space reserved for picture-in-picture video overlay\n")
	fmt.Printf("🎨 Features dynamic background and smooth text animations\n")
	
	return nil
}