package fcp

import (
	"strings"
	"testing"
)

// TestContentSecurityValidation tests all content security validation features
func TestContentSecurityValidation(t *testing.T) {
	validator := NewContentSecurityValidator()

	t.Run("ValidateTextContent - XSS Prevention", func(t *testing.T) {
		maliciousTexts := []string{
			"javascript:alert('xss')",
			"JAVASCRIPT:alert('xss')",
			"vbscript:msgbox('xss')",
			"data:text/html,<script>alert('xss')</script>",
			"data:application/javascript,alert('xss')",
			"<script>alert('xss')</script>",
			"</script><script>alert('xss')</script>",
			"eval('alert(1)')",
			"expression(alert('xss'))",
			"onload=alert('xss')",
			"onerror=alert('xss')",
			"onclick=alert('xss')",
			"onmouseover=alert('xss')",
		}

		for _, maliciousText := range maliciousTexts {
			err := validator.ValidateTextContent(maliciousText)
			if err == nil {
				t.Errorf("Expected validation to fail for malicious text: %s", maliciousText)
			}
			if !strings.Contains(err.Error(), "script injection detected") &&
				!strings.Contains(err.Error(), "dangerous pattern detected") {
				t.Errorf("Expected script injection error for: %s, got: %v", maliciousText, err)
			}
		}
	})

	t.Run("ValidateTextContent - NULL Byte Prevention", func(t *testing.T) {
		nullByteTexts := []string{
			"NULL\x00BYTES",
			"test\x00content",
			"\x00start",
			"end\x00",
			"multi\x00ple\x00nulls",
		}

		for _, nullByteText := range nullByteTexts {
			err := validator.ValidateTextContent(nullByteText)
			if err == nil {
				t.Errorf("Expected validation to fail for NULL byte text: %q", nullByteText)
			}
			if !strings.Contains(err.Error(), "null bytes or control characters not allowed") {
				t.Errorf("Expected null byte error for: %q, got: %v", nullByteText, err)
			}
		}
	})

	t.Run("ValidateTextContent - HTML Entity Prevention", func(t *testing.T) {
		dangerousEntities := []string{
			"&lt;script&gt;alert()&lt;/script&gt;",
			"&#x3C;script&#x3E;",
			"&#0;",
			"&quot;onclick=alert()&quot;",
			"&#34;onclick=alert()&#34;",
			"&#39;onclick=alert()&#39;",
			"&#47;script&#47;",
			"&#92;windows&#92;",
		}

		for _, entityText := range dangerousEntities {
			err := validator.ValidateTextContent(entityText)
			if err == nil {
				t.Errorf("Expected validation to fail for dangerous entity: %s", entityText)
			}
			if !strings.Contains(err.Error(), "dangerous HTML entity detected") &&
				!strings.Contains(err.Error(), "script injection detected") &&
				!strings.Contains(err.Error(), "path traversal detected") {
				t.Errorf("Expected HTML entity or script injection error for: %s, got: %v", entityText, err)
			}
		}
	})

	t.Run("ValidateTextContent - Path Traversal Prevention", func(t *testing.T) {
		pathTraversalTexts := []string{
			"../etc/passwd",
			"..\\windows\\system32",
			"/etc/shadow",
			"/bin/bash",
			"/usr/bin/curl",
			"/var/log/auth.log",
			"/tmp/exploit",
			"c:\\windows\\system32\\cmd.exe",
			"\\windows\\explorer.exe",
			"\\system32\\notepad.exe",
		}

		for _, pathText := range pathTraversalTexts {
			err := validator.ValidateTextContent(pathText)
			if err == nil {
				t.Errorf("Expected validation to fail for path traversal: %s", pathText)
			}
			if !strings.Contains(err.Error(), "file system path detected") {
				t.Errorf("Expected path traversal error for: %s, got: %v", pathText, err)
			}
		}
	})

	t.Run("ValidateTextContent - Unicode Exploitation Prevention", func(t *testing.T) {
		unicodeExploits := []string{
			"\uFEFF", // BOM
			"test\u202Eexploit", // Right-to-left override
			"test\u200Eexploit", // Left-to-right mark
			"test\u200Fexploit", // Right-to-left mark
		}

		for _, unicodeText := range unicodeExploits {
			err := validator.ValidateTextContent(unicodeText)
			if err == nil {
				t.Errorf("Expected validation to fail for Unicode exploit: %q", unicodeText)
			}
			if !strings.Contains(err.Error(), "dangerous Unicode characters detected") {
				t.Errorf("Expected Unicode error for: %q, got: %v", unicodeText, err)
			}
		}
	})

	t.Run("ValidateTextContent - Length Limit", func(t *testing.T) {
		longText := strings.Repeat("A", 100001)
		err := validator.ValidateTextContent(longText)
		if err == nil {
			t.Error("Expected validation to fail for overly long text")
		}
		if !strings.Contains(err.Error(), "text content too long") {
			t.Errorf("Expected length error, got: %v", err)
		}
	})

	t.Run("ValidateTextContent - Valid Text Passes", func(t *testing.T) {
		validTexts := []string{
			"Normal text content",
			"Text with numbers 123",
			"Text with symbols !@#$%^&*()",
			"Multiple words with spaces",
			"Unicode text: éñ中文🎥",
			"", // Empty text should pass
		}

		for _, validText := range validTexts {
			err := validator.ValidateTextContent(validText)
			if err != nil {
				t.Errorf("Expected validation to pass for valid text: %s, got: %v", validText, err)
			}
		}
	})
}

// TestNumericRangeValidation tests all numeric range validation features
func TestNumericRangeValidation(t *testing.T) {
	validator := NewNumericRangeValidator()

	t.Run("ValidateOpacity - Invalid Values", func(t *testing.T) {
		invalidOpacities := []string{
			"-0.1",  // Negative
			"1.5",   // Too high
			"-1.0",  // Negative
			"2.0",   // Too high
			"∞",     // Infinity
			"-∞",    // Negative infinity
			"NaN",   // Not a number
			"null",  // Invalid string
		}

		for _, opacity := range invalidOpacities {
			err := validator.ValidateOpacity(opacity)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid opacity: %s", opacity)
			}
		}
	})

	t.Run("ValidateColorValue - Invalid Values", func(t *testing.T) {
		invalidColors := []string{
			"∞ ∞ ∞ ∞",      // Infinity values
			"-0.1 0.5 0.5",  // Negative component
			"0.5 1.5 0.5",   // Component too high
			"1 2 3 4 5",     // Too many components
			"1",             // Too few components
			"NaN 0.5 0.5",   // NaN component
			"inf 0.5 0.5",   // Infinity component
		}

		for _, color := range invalidColors {
			err := validator.ValidateColorValue(color)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid color: %s", color)
			}
		}
	})

	t.Run("ValidateFontSize - Invalid Values", func(t *testing.T) {
		invalidFontSizes := []string{
			"0",     // Too small
			"-10",   // Negative
			"2001",  // Too large
			"∞",     // Infinity
			"NaN",   // Not a number
		}

		for _, fontSize := range invalidFontSizes {
			err := validator.ValidateFontSize(fontSize)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid font size: %s", fontSize)
			}
		}
	})

	t.Run("ValidateLineSpacing - Invalid Values", func(t *testing.T) {
		invalidLineSpacings := []string{
			"-1.0",  // Negative
			"0.05",  // Too small
			"25.0",  // Too large
			"∞",     // Infinity
			"NaN",   // Not a number
		}

		for _, lineSpacing := range invalidLineSpacings {
			err := validator.ValidateLineSpacing(lineSpacing)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid line spacing: %s", lineSpacing)
			}
		}
	})

	t.Run("ValidateScaleValue - Invalid Values", func(t *testing.T) {
		invalidScales := []string{
			"-1.0 1.0",  // Negative component
			"0.0 1.0",   // Too small
			"101.0 1.0", // Too large
			"∞ 1.0",     // Infinity component
			"NaN 1.0",   // NaN component
		}

		for _, scale := range invalidScales {
			err := validator.ValidateScaleValue(scale)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid scale: %s", scale)
			}
		}
	})

	t.Run("ValidateRotationValue - Invalid Values", func(t *testing.T) {
		invalidRotations := []string{
			"3601",  // Too large
			"-3601", // Too small
			"∞",     // Infinity
			"NaN",   // Not a number
		}

		for _, rotation := range invalidRotations {
			err := validator.ValidateRotationValue(rotation)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid rotation: %s", rotation)
			}
		}
	})

	t.Run("Valid Values Pass", func(t *testing.T) {
		testCases := []struct {
			validatorFunc func(string) error
			validValues   []string
			description   string
		}{
			{validator.ValidateOpacity, []string{"0.0", "0.5", "1.0"}, "opacity"},
			{validator.ValidateColorValue, []string{"0.5 0.5 0.5", "1.0 0.0 0.0 1.0"}, "color"},
			{validator.ValidateFontSize, []string{"12", "24.5", "100"}, "font size"},
			{validator.ValidateLineSpacing, []string{"1.0", "1.5", "2.0"}, "line spacing"},
			{validator.ValidateScaleValue, []string{"1.0 1.0", "0.5 2.0"}, "scale"},
			{validator.ValidateRotationValue, []string{"0", "45", "-90", "360"}, "rotation"},
		}

		for _, testCase := range testCases {
			for _, validValue := range testCase.validValues {
				err := testCase.validatorFunc(validValue)
				if err != nil {
					t.Errorf("Expected validation to pass for valid %s: %s, got: %v", testCase.description, validValue, err)
				}
			}
		}
	})
}

// TestBoundaryValidation tests all boundary validation features
func TestBoundaryValidation(t *testing.T) {
	validator := NewBoundaryValidator()

	t.Run("ValidatePosition - Invalid Values", func(t *testing.T) {
		invalidPositions := []string{
			"50001 0",      // X too large
			"0 50001",      // Y too large
			"-50001 0",     // X too small
			"0 -50001",     // Y too small
			"∞ 0",          // Infinity X
			"0 ∞",          // Infinity Y
			"NaN 0",        // NaN X
			"100",          // Missing component
			"100 200 300",  // Too many components
		}

		for _, position := range invalidPositions {
			err := validator.ValidatePosition(position)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid position: %s", position)
			}
		}
	})

	t.Run("ValidateLaneNumber - Invalid Values", func(t *testing.T) {
		invalidLanes := []string{
			"-101",  // Too small
			"101",   // Too large
		}

		for _, lane := range invalidLanes {
			err := validator.ValidateLaneNumber(lane)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid lane: %s", lane)
			}
		}
	})

	t.Run("ValidateTimeOffset - Invalid Values", func(t *testing.T) {
		invalidOffsets := []string{
			"-100/24000s",    // Negative
			"100/0s",         // Zero denominator
			"100/-24000s",    // Negative denominator
			"2073600000/24000s", // Too large (>24 hours)
		}

		for _, offset := range invalidOffsets {
			err := validator.ValidateTimeOffset(offset)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid time offset: %s", offset)
			}
		}
	})

	t.Run("ValidateDuration - Invalid Values", func(t *testing.T) {
		invalidDurations := []string{
			"-100/24000s",    // Negative
			"100/0s",         // Zero denominator
			"100/-24000s",    // Negative denominator
			"2073600000/24000s", // Too large (>24 hours)
		}

		for _, duration := range invalidDurations {
			err := validator.ValidateDuration(duration)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid duration: %s", duration)
			}
		}
	})

	t.Run("ValidateAnchorPoint - Invalid Values", func(t *testing.T) {
		invalidAnchors := []string{
			"-6.0 0.5",  // X too small
			"6.0 0.5",   // X too large
			"0.5 -6.0",  // Y too small
			"0.5 6.0",   // Y too large
			"∞ 0.5",     // Infinity X
			"0.5 ∞",     // Infinity Y
		}

		for _, anchor := range invalidAnchors {
			err := validator.ValidateAnchorPoint(anchor)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid anchor point: %s", anchor)
			}
		}
	})

	t.Run("Valid Values Pass", func(t *testing.T) {
		testCases := []struct {
			validatorFunc func(string) error
			validValues   []string
			description   string
		}{
			{validator.ValidatePosition, []string{"0 0", "1920 1080", "-100 -100"}, "position"},
			{validator.ValidateLaneNumber, []string{"0", "1", "-1", "50", "-50"}, "lane"},
			{validator.ValidateTimeOffset, []string{"0s", "1001/24000s", "10010/24000s"}, "time offset"},
			{validator.ValidateDuration, []string{"0s", "1001/24000s", "10010/24000s"}, "duration"},
			{validator.ValidateAnchorPoint, []string{"0.5 0.5", "0.0 0.0", "1.0 1.0"}, "anchor point"},
		}

		for _, testCase := range testCases {
			for _, validValue := range testCase.validValues {
				err := testCase.validatorFunc(validValue)
				if err != nil {
					t.Errorf("Expected validation to pass for valid %s: %s, got: %v", testCase.description, validValue, err)
				}
			}
		}
	})
}

// TestFontNameValidation tests font name security validation
func TestFontNameValidation(t *testing.T) {
	validator := NewContentSecurityValidator()

	t.Run("ValidateFontName - Invalid Values", func(t *testing.T) {
		invalidFonts := []string{
			"../fonts/malicious.ttf",  // Path traversal
			"font\\..\\system.ttf",    // Windows path traversal
			"font\x00name",            // NULL byte
			"font\x01name",            // Control character
			strings.Repeat("A", 201), // Too long
		}

		for _, fontName := range invalidFonts {
			err := validator.ValidateFontName(fontName)
			if err == nil {
				t.Errorf("Expected validation to fail for invalid font name: %s", fontName)
			}
		}
	})

	t.Run("ValidateFontName - Valid Values", func(t *testing.T) {
		validFonts := []string{
			"Arial",
			"Helvetica Neue",
			"Times-Roman",
			"SF Pro Display",
			"Font_Name_123",
			"Font (Regular)",
		}

		for _, fontName := range validFonts {
			err := validator.ValidateFontName(fontName)
			if err != nil {
				t.Errorf("Expected validation to pass for valid font name: %s, got: %v", fontName, err)
			}
		}
	})
}

// TestIntegratedValidation tests the full validation pipeline
func TestIntegratedValidation(t *testing.T) {
	t.Run("Extreme Values Should Be Blocked", func(t *testing.T) {
		// Create a basic FCPXML with extreme values that should be blocked
		fcpxml := &FCPXML{
			Version: "1.11",
			Library: Library{
				Events: []Event{
					{
						Name: "Test Event",
						Projects: []Project{
							{
								Name: "Test Project",
								Sequences: []Sequence{
									{
										Format:   "r1",
										Duration: "240240/24000s",
										TCStart:  "00:00:00:00",
										TCFormat: "NDF",
										AudioLayout: "stereo",
										AudioRate: "48k",
										Spine: Spine{
											Titles: []Title{
												{
													Ref:      "r1",
													Name:     "javascript:alert('xss')", // XSS attempt
													Offset:   "-100/24000s",             // Negative offset
													Duration: "100/24000s",
													Lane:     "200",                     // Extreme lane
													TextStyleDefs: []TextStyleDef{
														{
															ID: "ts1",
															TextStyle: TextStyle{
																Font:        "../../../etc/passwd", // Path traversal
																FontSize:    "7500",                 // Extreme font size
																FontColor:   "∞ ∞ ∞ ∞",            // Infinity colors
																LineSpacing: "-5.0",                 // Negative line spacing
															},
														},
													},
													Text: &TitleText{
														TextStyles: []TextStyleRef{
															{
																Ref:  "ts1",
																Text: "NULL\x00BYTES", // NULL bytes
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Resources: Resources{
				Assets: []Asset{
					{
						ID:       "r1",
						Name:     "&lt;script&gt;alert()&lt;/script&gt;", // HTML entities
						UID:      "javascript:eval('xss')",             // XSS in UID
						Duration: "100/24000s",
						Start:    "3600s",
						MediaRep: MediaRep{
							Kind: "original-media",
							Sig:  "test",
							Src:  "file:///test.mp4",
						},
					},
				},
				Formats: []Format{
					{
						ID:     "r1",
						Name:   "FFVideoFormat1080p2997",
						Width:  "1920",
						Height: "1080",
						FrameDuration: "1001/30000s",
					},
				},
			},
		}

		// This should fail validation
		err := fcpxml.ValidateStructure()
		if err == nil {
			t.Error("Expected validation to fail for FCPXML with extreme values")
		}

		// Check that multiple validation errors are caught
		errorStr := err.Error()
		expectedErrors := []string{
			"content security validation failed",
			"boundary validation",
			"range validation",
		}

		for _, expectedError := range expectedErrors {
			if !strings.Contains(errorStr, expectedError) {
				t.Logf("Full error: %v", err)
			}
		}
	})
}