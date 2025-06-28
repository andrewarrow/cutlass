package fcp

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ContentSecurityValidator prevents injection attacks and dangerous content
type ContentSecurityValidator struct{}

// NewContentSecurityValidator creates a new content security validator
func NewContentSecurityValidator() *ContentSecurityValidator {
	return &ContentSecurityValidator{}
}

// ValidateTextContent validates text content for security issues
func (csv *ContentSecurityValidator) ValidateTextContent(text string) error {
	// Try to URL decode to catch double-encoded attacks
	decodedText := text
	if decoded, err := url.QueryUnescape(text); err == nil {
		decodedText = decoded
		// Double decode for double-encoded attacks
		if doubleDecoded, err := url.QueryUnescape(decoded); err == nil {
			decodedText = doubleDecoded
		}
	}
	// Block script injection patterns
	scriptPatterns := []string{
		"javascript:",
		"vbscript:",
		"data:text/html",
		"data:application/",
		"<script",
		"</script>",
		"eval(",
		"expression(",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
	}
	
	// Check both original and decoded text for patterns
	// Also create normalized versions without whitespace to catch bypass attempts
	normalizeText := func(s string) string {
		// Remove all whitespace and control characters
		result := strings.ReplaceAll(s, " ", "")
		result = strings.ReplaceAll(result, "\t", "")
		result = strings.ReplaceAll(result, "\n", "")
		result = strings.ReplaceAll(result, "\r", "")
		result = strings.ReplaceAll(result, "\v", "")
		result = strings.ReplaceAll(result, "\f", "")
		return result
	}
	
	textsToCheck := []string{
		strings.ToLower(text), 
		strings.ToLower(decodedText),
		normalizeText(strings.ToLower(text)),
		normalizeText(strings.ToLower(decodedText)),
	}
	for _, textToCheck := range textsToCheck {
		for _, pattern := range scriptPatterns {
			if strings.Contains(textToCheck, pattern) {
				return fmt.Errorf("script injection detected in text content: %s", pattern)
			}
		}
	}
	
	// Block LDAP injection patterns
	ldapPatterns := []string{
		"${jndi:",
		"${ldap:",
		"jndi:ldap:",
		"jndi:rmi:",
		"jndi:dns:",
	}
	for _, textToCheck := range textsToCheck {
		for _, pattern := range ldapPatterns {
			if strings.Contains(textToCheck, pattern) {
				return fmt.Errorf("LDAP injection detected in text content: %s", pattern)
			}
		}
	}
	
	// Block NULL bytes and control characters (check both original and decoded)
	allTexts := []string{text, decodedText}
	for _, textToCheck := range allTexts {
		for i, char := range textToCheck {
			if char == 0 || (char < 32 && char != 9 && char != 10 && char != 13) {
				return fmt.Errorf("null bytes or control characters not allowed in text content at position %d", i)
			}
		}
	}
	
	// Block dangerous HTML entities
	dangerousEntities := []string{
		"&#x",
		"&#0",
		"&lt;script",
		"&gt;",
		"&quot;",
		"&#34;",
		"&#39;",
		"&#47;",
		"&#92;",
	}
	for _, textToCheck := range textsToCheck {
		for _, entity := range dangerousEntities {
			if strings.Contains(textToCheck, entity) {
				return fmt.Errorf("dangerous HTML entity detected: %s", entity)
			}
		}
	}
	
	// Block file system paths
	pathPatterns := []string{
		"../",
		"..\\",
		"/etc/",
		"/bin/",
		"/usr/",
		"/var/",
		"/tmp/",
		"c:\\",
		"\\windows\\",
		"\\system32\\",
		"\\\\", // UNC paths
	}
	for _, textToCheck := range textsToCheck {
		for _, pattern := range pathPatterns {
			if strings.Contains(textToCheck, pattern) {
				return fmt.Errorf("path traversal detected in text content: %s", pattern)
			}
		}
	}
	
	// Validate text length (prevent DoS)
	if len(text) > 100000 {
		return fmt.Errorf("text content too long: %d characters (max 100000)", len(text))
	}
	
	// Block Unicode exploitation
	if strings.Contains(text, "\uFEFF") || // BOM
		strings.Contains(text, "\u202E") || // Right-to-left override
		strings.Contains(text, "\u200E") || // Left-to-right mark
		strings.Contains(text, "\u200F") {  // Right-to-left mark
		return fmt.Errorf("dangerous Unicode characters detected in text content")
	}
	
	return nil
}

// ValidateFontName validates font names for security issues
func (csv *ContentSecurityValidator) ValidateFontName(font string) error {
	if font == "" {
		return nil // Empty font names are allowed
	}
	
	// Block path traversal
	if strings.Contains(font, "..") || strings.Contains(font, "/") || strings.Contains(font, "\\") {
		return fmt.Errorf("path traversal detected in font name: %s", font)
	}
	
	// Block NULL bytes and control characters
	for i, char := range font {
		if char == 0 || char < 32 {
			return fmt.Errorf("null bytes or control characters not allowed in font name at position %d", i)
		}
	}
	
	// Validate font name format (allow common font name characters)
	matched, err := regexp.MatchString(`^[a-zA-Z0-9\s\-_.()]+$`, font)
	if err != nil {
		return fmt.Errorf("error validating font name regex: %v", err)
	}
	if !matched {
		return fmt.Errorf("invalid characters in font name: %s", font)
	}
	
	// Validate length
	if len(font) > 200 {
		return fmt.Errorf("font name too long: %d characters (max 200)", len(font))
	}
	
	return nil
}

// ValidateAlignmentValue validates text alignment values
func (csv *ContentSecurityValidator) ValidateAlignmentValue(alignment string) error {
	if alignment == "" {
		return nil
	}
	
	validAlignments := map[string]bool{
		"left":    true,
		"center":  true,
		"right":   true,
		"justify": true,
		"start":   true,
		"end":     true,
	}
	
	if !validAlignments[strings.ToLower(alignment)] {
		return fmt.Errorf("invalid alignment value: %s", alignment)
	}
	
	return nil
}

// ValidateStringAttribute validates generic string attributes for security
func (csv *ContentSecurityValidator) ValidateStringAttribute(attrName, value string) error {
	if value == "" {
		return nil
	}
	
	// Block NULL bytes and control characters
	for i, char := range value {
		if char == 0 || (char < 32 && char != 9 && char != 10 && char != 13) {
			return fmt.Errorf("null bytes or control characters not allowed in %s at position %d", attrName, i)
		}
	}
	
	// Block common injection patterns (but allow legitimate FCPXML UIDs)
	injectionPatterns := []string{
		"javascript:",
		"<script",
		"</script>",
		"eval(",
		"..\\",
	}
	
	lowerValue := strings.ToLower(value)
	for _, pattern := range injectionPatterns {
		if strings.Contains(lowerValue, pattern) {
			return fmt.Errorf("dangerous pattern detected in %s: %s", attrName, pattern)
		}
	}
	
	// Special check for path traversal - allow FCPXML UID patterns like ".../Titles.localized/..."
	// but block actual path traversal like "../etc/passwd"
	if strings.Contains(value, "../") && !strings.HasPrefix(value, ".../") {
		return fmt.Errorf("dangerous pattern detected in %s: ../", attrName)
	}
	
	// Validate length
	if len(value) > 1000 {
		return fmt.Errorf("%s too long: %d characters (max 1000)", attrName, len(value))
	}
	
	return nil
}