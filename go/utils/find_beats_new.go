package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"cutlass/fcp"
)

// createColorPNGs creates solid color PNG files for backgrounds
func createColorPNGs() error {
	colors := map[string]color.RGBA{
		"blue.png":  {0, 76, 255, 255},  // Blue
		"green.png": {0, 255, 76, 255},  // Green
	}
	
	// Create 1280x720 images
	width, height := 1280, 720
	
	for filename, clr := range colors {
		// Create image
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		
		// Fill with solid color
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, clr)
			}
		}
		
		// Save PNG file
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		
		err = png.Encode(file, img)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// GenerateBeatsVisualizationPNG creates an FCPXML with alternating background colors using PNG images
func GenerateBeatsVisualizationPNG(wavFile string, beats []BeatDetection, outputFile string) error {
	// Get absolute path for WAV file
	absWavPath, err := filepath.Abs(wavFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for audio: %v", err)
	}

	// Check if audio file exists
	if _, err := os.Stat(absWavPath); os.IsNotExist(err) {
		return fmt.Errorf("audio file does not exist: %s", absWavPath)
	}

	// Create solid color PNG files for backgrounds
	err = createColorPNGs()
	if err != nil {
		return fmt.Errorf("failed to create color PNGs: %v", err)
	}

	// Get absolute paths for PNG files
	absBlueImage, err := filepath.Abs("blue.png")
	if err != nil {
		return fmt.Errorf("failed to get blue image path: %v", err)
	}
	
	absGreenImage, err := filepath.Abs("green.png")
	if err != nil {
		return fmt.Errorf("failed to get green image path: %v", err)
	}

	// Calculate total duration from audio file or beats
	totalDuration := 10.0 // Default minimum
	if len(beats) > 0 {
		lastBeat := beats[len(beats)-1]
		totalDuration = lastBeat.Timestamp + 5.0 // Add 5 seconds after last beat
	}

	// Create base FCPXML
	fcpxml, err := fcp.GenerateEmpty("")
	if err != nil {
		return fmt.Errorf("failed to create base FCPXML: %v", err)
	}

	// Create color segments based on beats using sequential AddImage calls
	currentTime := 0.0
	isBlue := true // Start with blue

	for _, beat := range beats {
		// Create color segment from currentTime to beat.Timestamp
		segmentDuration := beat.Timestamp - currentTime
		
		if segmentDuration > 0 {
			var imagePath string
			
			if isBlue {
				imagePath = absBlueImage
			} else {
				imagePath = absGreenImage
			}

			// Add image segment using proven AddImage function
			err = fcp.AddImage(fcpxml, imagePath, segmentDuration)
			if err != nil {
				return fmt.Errorf("failed to add color segment at %.2fs: %v", currentTime, err)
			}
		}

		currentTime = beat.Timestamp
		isBlue = !isBlue // Alternate color
	}

	// Add final color segment from last beat to end
	if currentTime < totalDuration {
		finalDuration := totalDuration - currentTime
		
		var imagePath string
		if isBlue {
			imagePath = absBlueImage
		} else {
			imagePath = absGreenImage
		}

		err = fcp.AddImage(fcpxml, imagePath, finalDuration)
		if err != nil {
			return fmt.Errorf("failed to add final color segment: %v", err)
		}
	}

	// Add audio last (on separate lane)
	err = fcp.AddAudio(fcpxml, absWavPath)
	if err != nil {
		return fmt.Errorf("failed to add audio: %v", err)
	}

	// Write FCPXML file
	err = fcp.WriteToFile(fcpxml, outputFile)
	if err != nil {
		return fmt.Errorf("failed to write FCPXML: %v", err)
	}

	return nil
}