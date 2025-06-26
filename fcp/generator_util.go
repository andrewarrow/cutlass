package fcp

import (
	"fmt"
	"io"

	"math/rand"
	"os"

	"path/filepath"

	"strings"
	"time"
)

// Helper functions for random content generation
func generateRandomText() string {
	texts := []string{
		"BAFFLE TEST", "Random Text", "Complex Timeline", "Stress Test",
		"FCPXML Generation", "Multi-Lane Test", "Animation Check",
		"Effect Validation", "Resource Test", "Lane Assignment",
		"Keyframe Test", "Timeline Stress", "Generation Check",
	}
	return texts[rand.Intn(len(texts))]
}

func randomFont() string {
	fonts := []string{"Helvetica", "Arial", "Times", "Courier", "Georgia", "Verdana"}
	return fonts[rand.Intn(len(fonts))]
}

func randomColor() string {
	return fmt.Sprintf("%.2f %.2f %.2f 1", rand.Float64(), rand.Float64(), rand.Float64())
}

func randomAlignment() string {
	alignments := []string{"left", "center", "right"}
	return alignments[rand.Intn(len(alignments))]
}

// updateSequenceDuration updates the sequence duration to match content
func updateSequenceDuration(fcpxml *FCPXML, totalDuration float64) {
	sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
	sequence.Duration = ConvertSecondsToFCPDuration(totalDuration)
}

// createUniqueMediaCopy creates a temporary copy of a media file with a unique name
// This prevents FCP UID cache conflicts by ensuring each BAFFLE run uses truly unique files
func createUniqueMediaCopy(originalPath, prefix string) (string, error) {

	timestamp := time.Now().UnixNano()
	randomNum := rand.Int63()
	ext := filepath.Ext(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), ext)

	uniqueName := fmt.Sprintf("%s_%s_%d_%d%s", prefix, baseName, timestamp, randomNum, ext)

	tempDir := filepath.Join(os.TempDir(), "cutlass_baffle")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return originalPath, fmt.Errorf("failed to create temp directory: %v", err)
	}

	uniquePath := filepath.Join(tempDir, uniqueName)

	sourceFile, err := os.Open(originalPath)
	if err != nil {
		return originalPath, fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(uniquePath)
	if err != nil {
		return originalPath, fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return originalPath, fmt.Errorf("failed to copy file contents: %v", err)
	}

	return uniquePath, nil
}
