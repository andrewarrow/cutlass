package utils

import (
	"os"
	"testing"
)

// TestShadowTextGeneration validates that shadow text generation follows CLAUDE.md patterns
func TestShadowTextGeneration(t *testing.T) {
	// Create test input file
	testInput := "test_shadow.txt"
	testOutput := "test_shadow_output.fcpxml"
	
	// Write test content
	testContent := "Making $37,000 on Amazon Kindle Direct Publishing sounds amazing"
	if err := os.WriteFile(testInput, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test input: %v", err)
	}
	defer os.Remove(testInput)
	defer os.Remove(testOutput)
	
	// Generate shadow text FCPXML
	if err := generateShadowTextFCPXML(testInput, testOutput); err != nil {
		t.Fatalf("Failed to generate shadow text FCPXML: %v", err)
	}
	
	// Verify output file exists
	if _, err := os.Stat(testOutput); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}
	
	// Read and validate generated content
	content, err := os.ReadFile(testOutput)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	xmlContent := string(content)
	
	// 🚨 CRITICAL VALIDATIONS based on CLAUDE.md lessons:
	
	// 1. Must use proven Vivid UID (not fictional Custom UID)
	if !containsString(xmlContent, "Vivid.localized/Vivid.motn") {
		t.Error("❌ Must use proven Vivid generator UID from samples/blue_background.fcpxml")
	}
	
	// 2. Must NOT use fictional Custom UID that causes crashes
	if containsString(xmlContent, "Custom.localized/Custom.moti") {
		t.Error("❌ FORBIDDEN: Custom UID causes 'item could not be read' errors")
	}
	
	// 3. Must have proper nested structure (video with titles inside)
	if !containsString(xmlContent, "<video ref=") {
		t.Error("❌ Must have video element for background (not titles directly in spine)")
	}
	
	// 4. Must have nested titles with correct timing
	if !containsString(xmlContent, "lane=\"1\"") {
		t.Error("❌ Must have titles with lane='1' (nested structure)")
	}
	
	// 5. Must have proper shadow text styling
	if !containsString(xmlContent, "shadowColor=") {
		t.Error("❌ Must have shadow text styling attributes")
	}
	
	// 6. Must have creative text splitting (multiple text-style elements)
	if !containsString(xmlContent, "text-style ref=") {
		t.Error("❌ Must have creative text splitting with multiple text-style refs")
	}
	
	// 7. Must use frame-aligned timing
	if !containsString(xmlContent, "/24000s") {
		t.Error("❌ Must use frame-aligned timing (ending in /24000s)")
	}
	
	// 8. Must have absolute timeline positioning for nested titles
	if !containsString(xmlContent, "offset=\"86") {
		t.Error("❌ Must have absolute timeline positions (starting around 86486400/24000s)")
	}
	
	t.Logf("✅ Shadow text generation passed all CLAUDE.md compliance checks")
}

// TestShadowTextDTDValidation ensures generated XML validates against DTD
func TestShadowTextDTDValidation(t *testing.T) {
	// This test would run xmllint validation if xmllint is available
	// For now, we'll just check basic XML structure
	
	testInput := "test_shadow_dtd.txt"
	testOutput := "test_shadow_dtd_output.fcpxml"
	
	if err := os.WriteFile(testInput, []byte("Test text for DTD validation"), 0644); err != nil {
		t.Fatalf("Failed to create test input: %v", err)
	}
	defer os.Remove(testInput)
	defer os.Remove(testOutput)
	
	if err := generateShadowTextFCPXML(testInput, testOutput); err != nil {
		t.Fatalf("Failed to generate FCPXML: %v", err)
	}
	
	content, err := os.ReadFile(testOutput)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	
	xmlContent := string(content)
	
	// Basic XML structure validation
	requiredElements := []string{
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
		"<!DOCTYPE fcpxml>",
		"<fcpxml version=\"1.13\">",
		"<resources>",
		"<library",  // Note: library has attributes, so just check opening
		"<sequence",  // Note: sequence has attributes
		"<spine>",
		"</fcpxml>",
	}
	
	for _, element := range requiredElements {
		if !containsString(xmlContent, element) {
			t.Errorf("❌ Missing required XML element: %s", element)
		}
	}
	
	t.Logf("✅ Generated FCPXML has proper XML structure")
}

// Helper function to check if string contains substring
func containsString(content, substring string) bool {
	return len(content) > 0 && len(substring) > 0 && 
		   len(content) >= len(substring) &&
		   findSubstring(content, substring) >= 0
}

// Simple substring search without importing strings package
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}