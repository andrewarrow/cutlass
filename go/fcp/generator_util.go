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
	// 🚨 EXTREME: Generate absolutely wild text content to break validation
	extremeTexts := []string{
		"", // Empty text
		"<>&\"'", // XML special characters
		"🚨💥🔥💀☠️🎭🎪🎨🎬🎮", // Extreme emojis
		strings.Repeat("A", 10000), // Massive text
		"NULL\x00BYTES", // Null bytes
		"\"Quotes'Inside<XML>&Tags", // Nested quotes and XML
		"Multi\nLine\nText\nWith\nBreaks", // Newlines
		"\t\t\tTabs\t\t\t", // Tabs
		"NEGATIVE-LANE-999999", // Reference to validation issues
		strings.Repeat("💩", 1000), // Emoji spam
		"Line1\r\nLine2\r\nLine3", // Windows line endings
		"🚨 KEYFRAME VALIDATION BREACH 🚨 TIMEBASE CORRUPTION 🚨",
		"&lt;&gt;&amp;&quot;&apos;", // HTML entities
		"javascript:alert('xss')", // XSS attempt
		"../../../etc/passwd", // Path traversal
		"DROP TABLE users;", // SQL injection
	}
	
	normalTexts := []string{
		"BAFFLE TEST", "Random Text", "Complex Timeline", "Stress Test",
		"FCPXML Generation", "Multi-Lane Test", "Animation Check",
		"Effect Validation", "Resource Test", "Lane Assignment",
		"Keyframe Test", "Timeline Stress", "Generation Check",
	}
	
	// 50% chance of extreme text, 50% normal
	if rand.Float32() < 0.5 {
		return extremeTexts[rand.Intn(len(extremeTexts))]
	}
	return normalTexts[rand.Intn(len(normalTexts))]
}

func randomFont() string {
	// Generate valid font names - can include unusual but valid fonts
	validFonts := []string{
		"Helvetica", "Arial", "Times", "Courier", "Georgia", "Verdana",
		"Times New Roman", "Helvetica Neue", "Comic Sans MS", "Impact",
		"Trebuchet MS", "Arial Black", "Palatino", "Garamond", "Bookman",
		"Lucida Sans Unicode", "Tahoma", "Monaco", "Andale Mono",
		"SF Pro Text", "SF Pro Display", "Avenir", "Avenir Next",
		"Futura", "Gill Sans", "Optima", "Baskerville",
	}
	
	// Always return valid font
	return validFonts[rand.Intn(len(validFonts))]
}

func randomColor() string {
	// Generate valid RGB color values (0.0 to 1.0 range with alpha 1.0)
	// Include edge cases that are valid but unusual
	colorOptions := [][]float64{
		{0.0, 0.0, 0.0, 1.0}, // Black
		{1.0, 1.0, 1.0, 1.0}, // White  
		{1.0, 0.0, 0.0, 1.0}, // Pure red
		{0.0, 1.0, 0.0, 1.0}, // Pure green
		{0.0, 0.0, 1.0, 1.0}, // Pure blue
		{1.0, 1.0, 0.0, 1.0}, // Yellow
		{1.0, 0.0, 1.0, 1.0}, // Magenta
		{0.0, 1.0, 1.0, 1.0}, // Cyan
	}
	
	// 20% chance of predefined edge case colors, 80% random valid colors
	if rand.Float32() < 0.2 {
		color := colorOptions[rand.Intn(len(colorOptions))]
		return fmt.Sprintf("%.2f %.2f %.2f %.2f", color[0], color[1], color[2], color[3])
	}
	
	return fmt.Sprintf("%.2f %.2f %.2f 1", rand.Float64(), rand.Float64(), rand.Float64())
}

func randomAlignment() string {
	// Generate valid but complex alignment combinations
	validAlignments := []string{
		"left", 
		"center", 
		"right", 
		"justify",  // Valid FCPXML alignment
		"start",    // Valid CSS-style alignment
		"end",      // Valid CSS-style alignment
	}
	
	// Always return valid alignment
	return validAlignments[rand.Intn(len(validAlignments))]
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
