// Package text_style_validation implements Step 10 of the FCPXMLKit-inspired refactoring plan:
// Text style validation for fonts, formatting, and text-related elements.
//
// This provides comprehensive text style validation that ensures proper font usage,
// formatting rules, and text styling to prevent text-related import errors in FCP.
package fcp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TextStyleParameterType represents different types of text style parameters
type TextStyleParameterType int

const (
	TextStyleParameterUnknown TextStyleParameterType = iota
	TextStyleParameterFont
	TextStyleParameterFontSize
	TextStyleParameterFontFace
	TextStyleParameterFontColor
	TextStyleParameterAlignment
	TextStyleParameterLineSpacing
	TextStyleParameterKerning
	TextStyleParameterStroke
	TextStyleParameterShadow
)

// String returns the string representation of the text style parameter type
func (tspt TextStyleParameterType) String() string {
	switch tspt {
	case TextStyleParameterFont:
		return "font"
	case TextStyleParameterFontSize:
		return "fontSize"
	case TextStyleParameterFontFace:
		return "fontFace"
	case TextStyleParameterFontColor:
		return "fontColor"
	case TextStyleParameterAlignment:
		return "alignment"
	case TextStyleParameterLineSpacing:
		return "lineSpacing"
	case TextStyleParameterKerning:
		return "kerning"
	case TextStyleParameterStroke:
		return "stroke"
	case TextStyleParameterShadow:
		return "shadow"
	default:
		return "unknown"
	}
}

// ParseTextStyleParameterType parses a parameter name to determine its type
func ParseTextStyleParameterType(paramName string) TextStyleParameterType {
	switch strings.ToLower(paramName) {
	case "font":
		return TextStyleParameterFont
	case "fontsize":
		return TextStyleParameterFontSize
	case "fontface":
		return TextStyleParameterFontFace
	case "fontcolor":
		return TextStyleParameterFontColor
	case "alignment":
		return TextStyleParameterAlignment
	case "linespacing":
		return TextStyleParameterLineSpacing
	case "kerning":
		return TextStyleParameterKerning
	case "stroke", "strokecolor", "strokewidth":
		return TextStyleParameterStroke
	case "shadow", "shadowcolor", "shadowoffset", "shadowblurradius":
		return TextStyleParameterShadow
	default:
		return TextStyleParameterUnknown
	}
}

// TextStyleValidationRules defines validation rules for text style parameters
type TextStyleValidationRules struct {
	RequiredFormat   string
	AllowedValues    []string
	ValueValidator   func(string) error
	RangeMin         *float64
	RangeMax         *float64
	AllowEmpty       bool
}

// ValidatedTextStyle represents a text style with validation
type ValidatedTextStyle struct {
	ID              string
	Font            string
	FontSize        string
	FontFace        string
	FontColor       string
	Bold            string
	Italic          string
	StrokeColor     string
	StrokeWidth     string
	ShadowColor     string
	ShadowOffset    string
	ShadowBlurRadius string
	Kerning         string
	Alignment       string
	LineSpacing     string
	
	// Validation metadata
	IsValid         bool
	ValidationErrors []string
}

// NewValidatedTextStyle creates a validated text style
func NewValidatedTextStyle(id string) (*ValidatedTextStyle, error) {
	if id == "" {
		return nil, fmt.Errorf("text style ID cannot be empty")
	}
	
	// Validate ID format (should be alphanumeric + underscore)
	if !isValidTextStyleID(id) {
		return nil, fmt.Errorf("invalid text style ID format: %s", id)
	}
	
	return &ValidatedTextStyle{
		ID:               id,
		IsValid:          false,
		ValidationErrors: make([]string, 0),
	}, nil
}

// TextStyleValidator validates text styles based on FCP requirements
type TextStyleValidator struct {
	rules           map[TextStyleParameterType]TextStyleValidationRules
	knownFonts      []string
	systemFonts     []string
	rangeValidator  *NumericRangeValidator
	securityValidator *ContentSecurityValidator
}

// NewTextStyleValidator creates a new text style validator with default rules
func NewTextStyleValidator() *TextStyleValidator {
	validator := &TextStyleValidator{
		rules:             make(map[TextStyleParameterType]TextStyleValidationRules),
		knownFonts:        make([]string, 0),
		systemFonts:       getSystemFonts(),
		rangeValidator:    NewNumericRangeValidator(),
		securityValidator: NewContentSecurityValidator(),
	}
	
	validator.initializeDefaultRules()
	return validator
}

// initializeDefaultRules sets up the validation rules based on FCPXML text specifications
func (tsv *TextStyleValidator) initializeDefaultRules() {
	// Font validation - should be system font names
	tsv.rules[TextStyleParameterFont] = TextStyleValidationRules{
		RequiredFormat: "Font family name",
		AllowedValues:  tsv.systemFonts,
		ValueValidator: tsv.validateFontName,
		AllowEmpty:     false,
	}
	
	// Font size validation - should be numeric with optional units
	tsv.rules[TextStyleParameterFontSize] = TextStyleValidationRules{
		RequiredFormat: "Numeric value with optional unit (e.g., '24', '24px')",
		ValueValidator: tsv.validateFontSize,
		RangeMin:       floatPtr(1.0),
		RangeMax:       floatPtr(500.0),
		AllowEmpty:     false,
	}
	
	// Font face validation - Bold, Regular, Italic, etc.
	tsv.rules[TextStyleParameterFontFace] = TextStyleValidationRules{
		RequiredFormat: "Font face style",
		AllowedValues:  []string{"Regular", "Bold", "Italic", "Bold Italic", "Light", "Medium", "SemiBold", "Heavy", "Black"},
		ValueValidator: tsv.validateFontFace,
		AllowEmpty:     true,
	}
	
	// Font color validation - RGB values
	tsv.rules[TextStyleParameterFontColor] = TextStyleValidationRules{
		RequiredFormat: "RGB color values (e.g., '1 0 0' for red)",
		ValueValidator: tsv.validateRGBColor,
		AllowEmpty:     false,
	}
	
	// Text alignment validation
	tsv.rules[TextStyleParameterAlignment] = TextStyleValidationRules{
		RequiredFormat: "Text alignment",
		AllowedValues:  []string{"left", "center", "right", "justify"},
		ValueValidator: tsv.validateAlignment,
		AllowEmpty:     true,
	}
	
	// Line spacing validation
	tsv.rules[TextStyleParameterLineSpacing] = TextStyleValidationRules{
		RequiredFormat: "Numeric line spacing multiplier",
		ValueValidator: tsv.validateLineSpacing,
		RangeMin:       floatPtr(0.5),
		RangeMax:       floatPtr(5.0),
		AllowEmpty:     true,
	}
	
	// Kerning validation
	tsv.rules[TextStyleParameterKerning] = TextStyleValidationRules{
		RequiredFormat: "Numeric kerning value",
		ValueValidator: tsv.validateKerning,
		RangeMin:       floatPtr(-1000.0),
		RangeMax:       floatPtr(1000.0),
		AllowEmpty:     true,
	}
	
	// Stroke validation
	tsv.rules[TextStyleParameterStroke] = TextStyleValidationRules{
		RequiredFormat: "RGB color or stroke width",
		ValueValidator: tsv.validateStrokeParameter,
		AllowEmpty:     true,
	}
	
	// Shadow validation
	tsv.rules[TextStyleParameterShadow] = TextStyleValidationRules{
		RequiredFormat: "Shadow color, offset, or blur radius",
		ValueValidator: tsv.validateShadowParameter,
		AllowEmpty:     true,
	}
}

// ValidateTextStyle validates a complete text style
func (tsv *TextStyleValidator) ValidateTextStyle(textStyle *ValidatedTextStyle) error {
	if textStyle == nil {
		return fmt.Errorf("text style cannot be nil")
	}
	
	// Reset validation state
	textStyle.IsValid = false
	textStyle.ValidationErrors = make([]string, 0)
	
	// Validate required fields
	if err := tsv.validateParameter(TextStyleParameterFont, textStyle.Font); err != nil {
		textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("Font: %v", err))
	}
	
	if err := tsv.validateParameter(TextStyleParameterFontSize, textStyle.FontSize); err != nil {
		textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("FontSize: %v", err))
	}
	
	if err := tsv.validateParameter(TextStyleParameterFontColor, textStyle.FontColor); err != nil {
		textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("FontColor: %v", err))
	}
	
	// Validate optional fields if present
	if textStyle.FontFace != "" {
		if err := tsv.validateParameter(TextStyleParameterFontFace, textStyle.FontFace); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("FontFace: %v", err))
		}
	}
	
	if textStyle.Alignment != "" {
		if err := tsv.validateParameter(TextStyleParameterAlignment, textStyle.Alignment); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("Alignment: %v", err))
		}
	}
	
	if textStyle.LineSpacing != "" {
		if err := tsv.validateParameter(TextStyleParameterLineSpacing, textStyle.LineSpacing); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("LineSpacing: %v", err))
		}
	}
	
	if textStyle.Kerning != "" {
		if err := tsv.validateParameter(TextStyleParameterKerning, textStyle.Kerning); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("Kerning: %v", err))
		}
	}
	
	// Validate stroke parameters if present
	if textStyle.StrokeColor != "" {
		if err := tsv.validateParameter(TextStyleParameterStroke, textStyle.StrokeColor); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("StrokeColor: %v", err))
		}
	}
	
	if textStyle.StrokeWidth != "" {
		if err := tsv.validateStrokeWidth(textStyle.StrokeWidth); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("StrokeWidth: %v", err))
		}
	}
	
	// Validate shadow parameters if present
	if textStyle.ShadowColor != "" {
		if err := tsv.validateParameter(TextStyleParameterShadow, textStyle.ShadowColor); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("ShadowColor: %v", err))
		}
	}
	
	if textStyle.ShadowOffset != "" {
		if err := tsv.validateShadowOffset(textStyle.ShadowOffset); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("ShadowOffset: %v", err))
		}
	}
	
	if textStyle.ShadowBlurRadius != "" {
		if err := tsv.validateShadowBlurRadius(textStyle.ShadowBlurRadius); err != nil {
			textStyle.ValidationErrors = append(textStyle.ValidationErrors, fmt.Sprintf("ShadowBlurRadius: %v", err))
		}
	}
	
	// Validate boolean fields
	if textStyle.Bold != "" && textStyle.Bold != "1" && textStyle.Bold != "0" {
		textStyle.ValidationErrors = append(textStyle.ValidationErrors, "Bold: must be '0' or '1'")
	}
	
	if textStyle.Italic != "" && textStyle.Italic != "1" && textStyle.Italic != "0" {
		textStyle.ValidationErrors = append(textStyle.ValidationErrors, "Italic: must be '0' or '1'")
	}
	
	// Check if validation passed
	textStyle.IsValid = len(textStyle.ValidationErrors) == 0
	
	if !textStyle.IsValid {
		return fmt.Errorf("text style validation failed: %v", textStyle.ValidationErrors)
	}
	
	return nil
}

// validateParameter validates a specific text style parameter
func (tsv *TextStyleValidator) validateParameter(paramType TextStyleParameterType, value string) error {
	rules, exists := tsv.rules[paramType]
	if !exists {
		return fmt.Errorf("no validation rules for parameter type: %s", paramType.String())
	}
	
	// Check if empty value is allowed
	if value == "" {
		if !rules.AllowEmpty {
			return fmt.Errorf("value cannot be empty")
		}
		return nil
	}
	
	// Check against allowed values if specified
	if len(rules.AllowedValues) > 0 {
		found := false
		for _, allowed := range rules.AllowedValues {
			if value == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value '%s' not in allowed values: %v", value, rules.AllowedValues)
		}
	}
	
	// Run custom validator if specified
	if rules.ValueValidator != nil {
		if err := rules.ValueValidator(value); err != nil {
			return err
		}
	}
	
	return nil
}

// Value validators for different text style parameters

// validateFontName validates font family names
func (tsv *TextStyleValidator) validateFontName(fontName string) error {
	if fontName == "" {
		return fmt.Errorf("font name cannot be empty")
	}
	
	// Use security validator for font name validation
	if err := tsv.securityValidator.ValidateFontName(fontName); err != nil {
		return fmt.Errorf("font name security validation failed: %v", err)
	}
	
	// Allow common system fonts and any font name for flexibility
	// In a production system, you might want to check against system fonts
	return nil
}

// validateFontSize validates font size values
func (tsv *TextStyleValidator) validateFontSize(fontSize string) error {
	// Remove unit suffix if present
	sizeStr := fontSize
	if strings.HasSuffix(fontSize, "px") || strings.HasSuffix(fontSize, "pt") {
		sizeStr = fontSize[:len(fontSize)-2]
	}
	
	// Use range validator for font size validation
	if err := tsv.rangeValidator.ValidateFontSize(sizeStr); err != nil {
		return fmt.Errorf("font size range validation failed: %v", err)
	}
	
	return nil
}

// validateFontFace validates font face styles
func (tsv *TextStyleValidator) validateFontFace(fontFace string) error {
	// Allow common font face values
	allowedFaces := []string{"Regular", "Bold", "Italic", "Bold Italic", "Light", "Medium", "SemiBold", "Heavy", "Black"}
	
	for _, allowed := range allowedFaces {
		if strings.EqualFold(fontFace, allowed) {
			return nil
		}
	}
	
	// Allow any other font face for flexibility, but warn about common ones
	return nil
}

// validateRGBColor validates RGB color values
func (tsv *TextStyleValidator) validateRGBColor(color string) error {
	// Use range validator for color validation
	if err := tsv.rangeValidator.ValidateColorValue(color); err != nil {
		return fmt.Errorf("color range validation failed: %v", err)
	}
	
	return nil
}

// validateAlignment validates text alignment values
func (tsv *TextStyleValidator) validateAlignment(alignment string) error {
	allowedAlignments := []string{"left", "center", "right", "justify"}
	
	for _, allowed := range allowedAlignments {
		if strings.EqualFold(alignment, allowed) {
			return nil
		}
	}
	
	return fmt.Errorf("invalid alignment: %s (allowed: %v)", alignment, allowedAlignments)
}

// validateLineSpacing validates line spacing values
func (tsv *TextStyleValidator) validateLineSpacing(lineSpacing string) error {
	// Use range validator for line spacing validation
	if err := tsv.rangeValidator.ValidateLineSpacing(lineSpacing); err != nil {
		return fmt.Errorf("line spacing range validation failed: %v", err)
	}
	
	return nil
}

// validateKerning validates kerning values
func (tsv *TextStyleValidator) validateKerning(kerning string) error {
	kern, err := strconv.ParseFloat(kerning, 64)
	if err != nil {
		return fmt.Errorf("invalid kerning format: %s", kerning)
	}
	
	if kern < -1000.0 || kern > 1000.0 {
		return fmt.Errorf("kerning out of reasonable range [-1000, 1000]: %f", kern)
	}
	
	return nil
}

// validateStrokeParameter validates stroke-related parameters
func (tsv *TextStyleValidator) validateStrokeParameter(stroke string) error {
	// Could be a color (RGB) or other stroke parameter
	if err := tsv.validateRGBColor(stroke); err == nil {
		return nil
	}
	
	// Could be other stroke parameter format
	return nil
}

// validateStrokeWidth validates stroke width values
func (tsv *TextStyleValidator) validateStrokeWidth(strokeWidth string) error {
	width, err := strconv.ParseFloat(strokeWidth, 64)
	if err != nil {
		return fmt.Errorf("invalid stroke width format: %s", strokeWidth)
	}
	
	if width < 0.0 {
		return fmt.Errorf("stroke width cannot be negative: %f", width)
	}
	
	if width > 50.0 {
		return fmt.Errorf("stroke width too large: %f", width)
	}
	
	return nil
}

// validateShadowParameter validates shadow-related parameters
func (tsv *TextStyleValidator) validateShadowParameter(shadow string) error {
	// Could be a color (RGB) or other shadow parameter
	if err := tsv.validateRGBColor(shadow); err == nil {
		return nil
	}
	
	// Could be other shadow parameter format
	return nil
}

// validateShadowOffset validates shadow offset values
func (tsv *TextStyleValidator) validateShadowOffset(shadowOffset string) error {
	parts := strings.Fields(shadowOffset)
	if len(parts) != 2 {
		return fmt.Errorf("shadow offset must have 2 components (x y): %s", shadowOffset)
	}
	
	for i, part := range parts {
		_, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return fmt.Errorf("invalid shadow offset component %d: %s", i, part)
		}
	}
	
	return nil
}

// validateShadowBlurRadius validates shadow blur radius values
func (tsv *TextStyleValidator) validateShadowBlurRadius(shadowBlurRadius string) error {
	radius, err := strconv.ParseFloat(shadowBlurRadius, 64)
	if err != nil {
		return fmt.Errorf("invalid shadow blur radius format: %s", shadowBlurRadius)
	}
	
	if radius < 0.0 {
		return fmt.Errorf("shadow blur radius cannot be negative: %f", radius)
	}
	
	if radius > 100.0 {
		return fmt.Errorf("shadow blur radius too large: %f", radius)
	}
	
	return nil
}

// TextStyleBuilder provides a safe way to build validated text styles
type TextStyleBuilder struct {
	textStyle *ValidatedTextStyle
	validator *TextStyleValidator
}

// NewTextStyleBuilder creates a new text style builder
func NewTextStyleBuilder(id string) (*TextStyleBuilder, error) {
	textStyle, err := NewValidatedTextStyle(id)
	if err != nil {
		return nil, fmt.Errorf("failed to create text style: %v", err)
	}
	
	return &TextStyleBuilder{
		textStyle: textStyle,
		validator: NewTextStyleValidator(),
	}, nil
}

// SetFont sets the font family
func (tsb *TextStyleBuilder) SetFont(font string) *TextStyleBuilder {
	tsb.textStyle.Font = font
	return tsb
}

// SetFontSize sets the font size
func (tsb *TextStyleBuilder) SetFontSize(fontSize string) *TextStyleBuilder {
	tsb.textStyle.FontSize = fontSize
	return tsb
}

// SetFontFace sets the font face
func (tsb *TextStyleBuilder) SetFontFace(fontFace string) *TextStyleBuilder {
	tsb.textStyle.FontFace = fontFace
	return tsb
}

// SetFontColor sets the font color
func (tsb *TextStyleBuilder) SetFontColor(fontColor string) *TextStyleBuilder {
	tsb.textStyle.FontColor = fontColor
	return tsb
}

// SetBold sets the bold flag
func (tsb *TextStyleBuilder) SetBold(bold bool) *TextStyleBuilder {
	if bold {
		tsb.textStyle.Bold = "1"
	} else {
		tsb.textStyle.Bold = "0"
	}
	return tsb
}

// SetItalic sets the italic flag
func (tsb *TextStyleBuilder) SetItalic(italic bool) *TextStyleBuilder {
	if italic {
		tsb.textStyle.Italic = "1"
	} else {
		tsb.textStyle.Italic = "0"
	}
	return tsb
}

// SetAlignment sets the text alignment
func (tsb *TextStyleBuilder) SetAlignment(alignment string) *TextStyleBuilder {
	tsb.textStyle.Alignment = alignment
	return tsb
}

// SetLineSpacing sets the line spacing
func (tsb *TextStyleBuilder) SetLineSpacing(lineSpacing string) *TextStyleBuilder {
	tsb.textStyle.LineSpacing = lineSpacing
	return tsb
}

// SetKerning sets the kerning
func (tsb *TextStyleBuilder) SetKerning(kerning string) *TextStyleBuilder {
	tsb.textStyle.Kerning = kerning
	return tsb
}

// SetStroke sets stroke color and width
func (tsb *TextStyleBuilder) SetStroke(color, width string) *TextStyleBuilder {
	tsb.textStyle.StrokeColor = color
	tsb.textStyle.StrokeWidth = width
	return tsb
}

// SetShadow sets shadow properties
func (tsb *TextStyleBuilder) SetShadow(color, offset, blurRadius string) *TextStyleBuilder {
	tsb.textStyle.ShadowColor = color
	tsb.textStyle.ShadowOffset = offset
	tsb.textStyle.ShadowBlurRadius = blurRadius
	return tsb
}

// Build creates the final validated text style
func (tsb *TextStyleBuilder) Build() (*ValidatedTextStyle, error) {
	if err := tsb.validator.ValidateTextStyle(tsb.textStyle); err != nil {
		return nil, fmt.Errorf("text style validation failed: %v", err)
	}
	
	return tsb.textStyle, nil
}

// TextStyleRegistry manages text style uniqueness and validation
type TextStyleRegistry struct {
	textStyles map[string]*ValidatedTextStyle
	validator  *TextStyleValidator
}

// NewTextStyleRegistry creates a new text style registry
func NewTextStyleRegistry() *TextStyleRegistry {
	return &TextStyleRegistry{
		textStyles: make(map[string]*ValidatedTextStyle),
		validator:  NewTextStyleValidator(),
	}
}

// RegisterTextStyle registers a text style with validation
func (tsr *TextStyleRegistry) RegisterTextStyle(textStyle *ValidatedTextStyle) error {
	if textStyle == nil {
		return fmt.Errorf("text style cannot be nil")
	}
	
	// Validate the text style
	if err := tsr.validator.ValidateTextStyle(textStyle); err != nil {
		return fmt.Errorf("text style validation failed: %v", err)
	}
	
	// Check for ID conflicts
	if existing, exists := tsr.textStyles[textStyle.ID]; exists {
		return fmt.Errorf("text style ID '%s' already exists: %+v", textStyle.ID, existing)
	}
	
	tsr.textStyles[textStyle.ID] = textStyle
	return nil
}

// GetTextStyle retrieves a text style by ID
func (tsr *TextStyleRegistry) GetTextStyle(id string) (*ValidatedTextStyle, error) {
	textStyle, exists := tsr.textStyles[id]
	if !exists {
		return nil, fmt.Errorf("text style with ID '%s' not found", id)
	}
	
	return textStyle, nil
}

// ValidateTextStyleReference validates that a text style reference exists
func (tsr *TextStyleRegistry) ValidateTextStyleReference(id string) error {
	if id == "" {
		return fmt.Errorf("text style reference cannot be empty")
	}
	
	_, exists := tsr.textStyles[id]
	if !exists {
		return fmt.Errorf("text style with ID '%s' not found", id)
	}
	
	return nil
}

// GetAllTextStyles returns all registered text styles
func (tsr *TextStyleRegistry) GetAllTextStyles() []*ValidatedTextStyle {
	textStyles := make([]*ValidatedTextStyle, 0, len(tsr.textStyles))
	for _, textStyle := range tsr.textStyles {
		textStyles = append(textStyles, textStyle)
	}
	return textStyles
}

// Helper functions

// isValidTextStyleID validates text style ID format
func isValidTextStyleID(id string) bool {
	// Text style IDs should be alphanumeric with underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, id)
	return matched
}

// getSystemFonts returns a list of common system fonts
func getSystemFonts() []string {
	return []string{
		"Arial", "Helvetica", "Times New Roman", "Times", "Courier New", "Courier",
		"Verdana", "Georgia", "Palatino", "Garamond", "Bookman", "Comic Sans MS",
		"Trebuchet MS", "Arial Black", "Impact", "Lucida Sans Unicode", "Tahoma",
		"Lucida Console", "Monaco", "Andale Mono", "Courier New", "System Font",
		"Helvetica Neue", "San Francisco", "SF Pro Text", "SF Pro Display",
		"Avenir", "Avenir Next", "Futura", "Gill Sans", "Optima", "Baskerville",
	}
}

// ValidateTextConfiguration validates a text configuration struct
func (tsv *TextStyleValidator) ValidateTextConfiguration(config *TextConfiguration) error {
	if config == nil {
		return fmt.Errorf("text configuration cannot be nil")
	}
	
	// Validate font
	if config.Font != "" {
		if err := tsv.validateFontName(config.Font); err != nil {
			return fmt.Errorf("invalid font: %v", err)
		}
	}
	
	// Validate font size
	if config.FontSize != "" {
		if err := tsv.validateFontSize(config.FontSize); err != nil {
			return fmt.Errorf("invalid font size: %v", err)
		}
	}
	
	// Validate font color
	if config.FontColor != "" {
		if err := tsv.validateRGBColor(config.FontColor); err != nil {
			return fmt.Errorf("invalid font color: %v", err)
		}
	}
	
	// Validate alignment
	if config.Alignment != "" {
		if err := tsv.validateAlignment(config.Alignment); err != nil {
			return fmt.Errorf("invalid alignment: %v", err)
		}
	}
	
	return nil
}

// floatPtr returns a pointer to a float64 value
func floatPtr(f float64) *float64 {
	return &f
}