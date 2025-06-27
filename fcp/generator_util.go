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
	// 🚨 EXTREME: Include invalid, massive, and special character fonts
	extremeFonts := []string{
		"", // Empty font
		"NonExistentFont", // Font that doesn't exist
		"🚨💥🔥", // Emoji font name
		strings.Repeat("A", 1000), // Massive font name
		"<XML>Font", // XML in font name
		"Font\"With'Quotes", // Quotes in font name
		"../../../Windows/Fonts/arial.ttf", // Path traversal
		"NULL\x00Font", // Null bytes
		"\n\t\rWeird Font", // Control characters
	}
	
	normalFonts := []string{"Helvetica", "Arial", "Times", "Courier", "Georgia", "Verdana"}
	
	// 30% chance of extreme fonts
	if rand.Float32() < 0.3 {
		return extremeFonts[rand.Intn(len(extremeFonts))]
	}
	return normalFonts[rand.Intn(len(normalFonts))]
}

func randomColor() string {
	// 🚨 EXTREME: Generate invalid, negative, and massive color values
	extremeOptions := []string{
		"", // Empty color
		"red", // Invalid format (not RGBA floats)
		"<color>", // XML in color
		fmt.Sprintf("%.2f %.2f %.2f %.2f", -5+rand.Float64()*10, -5+rand.Float64()*10, -5+rand.Float64()*10, -5+rand.Float64()*10), // Negative/huge values
		"NaN NaN NaN NaN", // Invalid numbers
		"∞ ∞ ∞ ∞", // Infinity symbols
		"1 2 3", // Wrong number of components
		"1 2 3 4 5 6", // Too many components
	}
	
	// 30% chance of extreme colors
	if rand.Float32() < 0.3 {
		return extremeOptions[rand.Intn(len(extremeOptions))]
	}
	
	return fmt.Sprintf("%.2f %.2f %.2f 1", rand.Float64(), rand.Float64(), rand.Float64())
}

func randomAlignment() string {
	// 🚨 EXTREME: Include invalid alignments
	extremeAlignments := []string{
		"", // Empty
		"invalid", // Invalid alignment
		"<XML>", // XML in alignment
		"NULL\x00", // Null bytes
		"999999999", // Numeric
		"超级对齐", // Unicode
		"justify-super-extreme", // Made up alignment
	}
	
	normalAlignments := []string{"left", "center", "right"}
	
	// 40% chance of extreme alignments  
	if rand.Float32() < 0.4 {
		return extremeAlignments[rand.Intn(len(extremeAlignments))]
	}
	return normalAlignments[rand.Intn(len(normalAlignments))]
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
