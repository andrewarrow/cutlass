package fcp

import (
	"testing"
)

func TestParseTextStyleParameterType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TextStyleParameterType
	}{
		{
			name:     "font parameter",
			input:    "font",
			expected: TextStyleParameterFont,
		},
		{
			name:     "fontSize parameter",
			input:    "fontSize",
			expected: TextStyleParameterFontSize,
		},
		{
			name:     "fontFace parameter",
			input:    "fontFace",
			expected: TextStyleParameterFontFace,
		},
		{
			name:     "fontColor parameter",
			input:    "fontColor",
			expected: TextStyleParameterFontColor,
		},
		{
			name:     "alignment parameter",
			input:    "alignment",
			expected: TextStyleParameterAlignment,
		},
		{
			name:     "lineSpacing parameter",
			input:    "lineSpacing",
			expected: TextStyleParameterLineSpacing,
		},
		{
			name:     "kerning parameter",
			input:    "kerning",
			expected: TextStyleParameterKerning,
		},
		{
			name:     "stroke parameter",
			input:    "stroke",
			expected: TextStyleParameterStroke,
		},
		{
			name:     "shadow parameter",
			input:    "shadow",
			expected: TextStyleParameterShadow,
		},
		{
			name:     "unknown parameter",
			input:    "unknown",
			expected: TextStyleParameterUnknown,
		},
		{
			name:     "case insensitive",
			input:    "FONT",
			expected: TextStyleParameterFont,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTextStyleParameterType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestNewValidatedTextStyle(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name: "valid text style ID",
			id:   "ts1",
		},
		{
			name: "valid text style ID with underscore",
			id:   "text_style_1",
		},
		{
			name: "valid text style ID alphanumeric",
			id:   "TextStyle123",
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
		},
		{
			name:        "invalid ID starting with number",
			id:          "1textStyle",
			expectError: true,
		},
		{
			name:        "invalid ID with special characters",
			id:          "text-style",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			textStyle, err := NewValidatedTextStyle(tt.id)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if textStyle.ID != tt.id {
				t.Errorf("expected ID %s, got %s", tt.id, textStyle.ID)
			}
			if textStyle.IsValid {
				t.Errorf("new text style should not be valid initially")
			}
		})
	}
}

func TestTextStyleValidator_ValidateFontName(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		fontName    string
		expectError bool
	}{
		{
			name:     "valid font name",
			fontName: "Arial",
		},
		{
			name:     "valid font name with spaces",
			fontName: "Times New Roman",
		},
		{
			name:        "empty font name",
			fontName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateFontName(tt.fontName)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateFontSize(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		fontSize    string
		expectError bool
	}{
		{
			name:     "valid numeric font size",
			fontSize: "24",
		},
		{
			name:     "valid font size with px unit",
			fontSize: "18px",
		},
		{
			name:     "valid font size with pt unit",
			fontSize: "12pt",
		},
		{
			name:     "valid decimal font size",
			fontSize: "14.5",
		},
		{
			name:        "invalid font size format",
			fontSize:    "abc",
			expectError: true,
		},
		{
			name:        "negative font size",
			fontSize:    "-12",
			expectError: true,
		},
		{
			name:        "zero font size",
			fontSize:    "0",
			expectError: true,
		},
		{
			name:        "font size too large",
			fontSize:    "1000",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateFontSize(tt.fontSize)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateRGBColor(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		color       string
		expectError bool
	}{
		{
			name:  "valid RGB color",
			color: "1.0 0.0 0.0",
		},
		{
			name:  "valid RGBA color",
			color: "0.5 0.5 0.5 0.8",
		},
		{
			name:  "valid RGB with zeros",
			color: "0 0 0",
		},
		{
			name:        "invalid RGB format (too few components)",
			color:       "1.0 0.0",
			expectError: true,
		},
		{
			name:        "invalid RGB format (too many components)",
			color:       "1.0 0.0 0.0 0.5 0.2",
			expectError: true,
		},
		{
			name:        "invalid RGB format (non-numeric)",
			color:       "red green blue",
			expectError: true,
		},
		{
			name:        "RGB component out of range (too high)",
			color:       "1.5 0.0 0.0",
			expectError: true,
		},
		{
			name:        "RGB component out of range (negative)",
			color:       "-0.1 0.0 0.0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRGBColor(tt.color)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateAlignment(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		alignment   string
		expectError bool
	}{
		{
			name:      "valid left alignment",
			alignment: "left",
		},
		{
			name:      "valid center alignment",
			alignment: "center",
		},
		{
			name:      "valid right alignment",
			alignment: "right",
		},
		{
			name:      "valid justify alignment",
			alignment: "justify",
		},
		{
			name:      "case insensitive alignment",
			alignment: "LEFT",
		},
		{
			name:        "invalid alignment",
			alignment:   "middle",
			expectError: true,
		},
		{
			name:        "invalid alignment",
			alignment:   "top",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateAlignment(tt.alignment)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateLineSpacing(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		lineSpacing string
		expectError bool
	}{
		{
			name:        "valid line spacing",
			lineSpacing: "1.0",
		},
		{
			name:        "valid tight line spacing",
			lineSpacing: "0.8",
		},
		{
			name:        "valid loose line spacing",
			lineSpacing: "2.0",
		},
		{
			name:        "line spacing too small",
			lineSpacing: "0.2",
			expectError: true,
		},
		{
			name:        "line spacing too large",
			lineSpacing: "10.0",
			expectError: true,
		},
		{
			name:        "invalid line spacing format",
			lineSpacing: "abc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateLineSpacing(tt.lineSpacing)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateKerning(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		kerning     string
		expectError bool
	}{
		{
			name:    "valid positive kerning",
			kerning: "5.0",
		},
		{
			name:    "valid negative kerning",
			kerning: "-3.0",
		},
		{
			name:    "valid zero kerning",
			kerning: "0",
		},
		{
			name:        "kerning too large",
			kerning:     "2000",
			expectError: true,
		},
		{
			name:        "kerning too small",
			kerning:     "-2000",
			expectError: true,
		},
		{
			name:        "invalid kerning format",
			kerning:     "abc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateKerning(tt.kerning)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateStrokeWidth(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		strokeWidth string
		expectError bool
	}{
		{
			name:        "valid stroke width",
			strokeWidth: "2.0",
		},
		{
			name:        "valid zero stroke width",
			strokeWidth: "0",
		},
		{
			name:        "valid decimal stroke width",
			strokeWidth: "1.5",
		},
		{
			name:        "negative stroke width",
			strokeWidth: "-1.0",
			expectError: true,
		},
		{
			name:        "stroke width too large",
			strokeWidth: "100",
			expectError: true,
		},
		{
			name:        "invalid stroke width format",
			strokeWidth: "abc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateStrokeWidth(tt.strokeWidth)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateShadowOffset(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name         string
		shadowOffset string
		expectError  bool
	}{
		{
			name:         "valid shadow offset",
			shadowOffset: "2.0 3.0",
		},
		{
			name:         "valid negative shadow offset",
			shadowOffset: "-1.0 -2.0",
		},
		{
			name:         "valid zero shadow offset",
			shadowOffset: "0 0",
		},
		{
			name:         "invalid shadow offset (too few components)",
			shadowOffset: "2.0",
			expectError:  true,
		},
		{
			name:         "invalid shadow offset (too many components)",
			shadowOffset: "2.0 3.0 4.0",
			expectError:  true,
		},
		{
			name:         "invalid shadow offset (non-numeric)",
			shadowOffset: "abc def",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateShadowOffset(tt.shadowOffset)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateShadowBlurRadius(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name              string
		shadowBlurRadius  string
		expectError       bool
	}{
		{
			name:             "valid shadow blur radius",
			shadowBlurRadius: "5.0",
		},
		{
			name:             "valid zero shadow blur radius",
			shadowBlurRadius: "0",
		},
		{
			name:             "valid decimal shadow blur radius",
			shadowBlurRadius: "2.5",
		},
		{
			name:             "negative shadow blur radius",
			shadowBlurRadius: "-1.0",
			expectError:      true,
		},
		{
			name:             "shadow blur radius too large",
			shadowBlurRadius: "200",
			expectError:      true,
		},
		{
			name:             "invalid shadow blur radius format",
			shadowBlurRadius: "abc",
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateShadowBlurRadius(tt.shadowBlurRadius)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTextStyleValidator_ValidateTextStyle(t *testing.T) {
	validator := NewTextStyleValidator()

	tests := []struct {
		name        string
		textStyle   *ValidatedTextStyle
		expectError bool
	}{
		{
			name: "valid complete text style",
			textStyle: &ValidatedTextStyle{
				ID:        "ts1",
				Font:      "Arial",
				FontSize:  "24",
				FontColor: "0 0 0",
			},
		},
		{
			name: "valid text style with optional fields",
			textStyle: &ValidatedTextStyle{
				ID:          "ts2",
				Font:        "Helvetica",
				FontSize:    "18px",
				FontColor:   "1 0 0",
				FontFace:    "Bold",
				Alignment:   "center",
				LineSpacing: "1.2",
				Kerning:     "2.0",
				Bold:        "1",
				Italic:      "0",
			},
		},
		{
			name:        "nil text style",
			textStyle:   nil,
			expectError: true,
		},
		{
			name: "missing required font",
			textStyle: &ValidatedTextStyle{
				ID:        "ts3",
				FontSize:  "24",
				FontColor: "0 0 0",
			},
			expectError: true,
		},
		{
			name: "missing required font size",
			textStyle: &ValidatedTextStyle{
				ID:        "ts4",
				Font:      "Arial",
				FontColor: "0 0 0",
			},
			expectError: true,
		},
		{
			name: "missing required font color",
			textStyle: &ValidatedTextStyle{
				ID:       "ts5",
				Font:     "Arial",
				FontSize: "24",
			},
			expectError: true,
		},
		{
			name: "invalid font size",
			textStyle: &ValidatedTextStyle{
				ID:        "ts6",
				Font:      "Arial",
				FontSize:  "abc",
				FontColor: "0 0 0",
			},
			expectError: true,
		},
		{
			name: "invalid font color",
			textStyle: &ValidatedTextStyle{
				ID:        "ts7",
				Font:      "Arial",
				FontSize:  "24",
				FontColor: "invalid color",
			},
			expectError: true,
		},
		{
			name: "invalid alignment",
			textStyle: &ValidatedTextStyle{
				ID:        "ts8",
				Font:      "Arial",
				FontSize:  "24",
				FontColor: "0 0 0",
				Alignment: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid bold flag",
			textStyle: &ValidatedTextStyle{
				ID:        "ts9",
				Font:      "Arial",
				FontSize:  "24",
				FontColor: "0 0 0",
				Bold:      "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid italic flag",
			textStyle: &ValidatedTextStyle{
				ID:        "ts10",
				Font:      "Arial",
				FontSize:  "24",
				FontColor: "0 0 0",
				Italic:    "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTextStyle(tt.textStyle)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.textStyle != nil && tt.textStyle.IsValid {
					t.Errorf("text style should not be valid after failed validation")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !tt.textStyle.IsValid {
					t.Errorf("text style should be valid after successful validation")
				}
			}
		})
	}
}

func TestTextStyleBuilder(t *testing.T) {
	builder, err := NewTextStyleBuilder("test_style")
	if err != nil {
		t.Fatalf("failed to create text style builder: %v", err)
	}

	textStyle, err := builder.
		SetFont("Arial").
		SetFontSize("24").
		SetFontColor("0 0 0").
		SetFontFace("Bold").
		SetAlignment("center").
		SetBold(true).
		SetItalic(false).
		SetLineSpacing("1.2").
		SetKerning("1.0").
		SetStroke("1 0 0", "2.0").
		SetShadow("0.5 0.5 0.5", "2.0 2.0", "3.0").
		Build()

	if err != nil {
		t.Errorf("unexpected error building text style: %v", err)
		return
	}

	// Verify properties
	if textStyle.ID != "test_style" {
		t.Errorf("expected ID 'test_style', got %s", textStyle.ID)
	}
	if textStyle.Font != "Arial" {
		t.Errorf("expected font 'Arial', got %s", textStyle.Font)
	}
	if textStyle.FontSize != "24" {
		t.Errorf("expected font size '24', got %s", textStyle.FontSize)
	}
	if textStyle.FontColor != "0 0 0" {
		t.Errorf("expected font color '0 0 0', got %s", textStyle.FontColor)
	}
	if textStyle.FontFace != "Bold" {
		t.Errorf("expected font face 'Bold', got %s", textStyle.FontFace)
	}
	if textStyle.Alignment != "center" {
		t.Errorf("expected alignment 'center', got %s", textStyle.Alignment)
	}
	if textStyle.Bold != "1" {
		t.Errorf("expected bold '1', got %s", textStyle.Bold)
	}
	if textStyle.Italic != "0" {
		t.Errorf("expected italic '0', got %s", textStyle.Italic)
	}
	if !textStyle.IsValid {
		t.Errorf("text style should be valid after successful build")
	}
}

func TestTextStyleBuilder_InvalidTextStyle(t *testing.T) {
	builder, err := NewTextStyleBuilder("invalid_style")
	if err != nil {
		t.Fatalf("failed to create text style builder: %v", err)
	}

	// Build an invalid text style (missing required fields)
	_, err = builder.
		SetFont("Arial").
		// Missing font size and color
		Build()

	if err == nil {
		t.Errorf("expected error building invalid text style")
	}
}

func TestTextStyleRegistry(t *testing.T) {
	registry := NewTextStyleRegistry()

	// Create valid text styles
	builder1, err := NewTextStyleBuilder("style1")
	if err != nil {
		t.Fatalf("failed to create builder 1: %v", err)
	}
	textStyle1, err := builder1.
		SetFont("Arial").
		SetFontSize("24").
		SetFontColor("0 0 0").
		Build()
	if err != nil {
		t.Fatalf("failed to build text style 1: %v", err)
	}

	builder2, err := NewTextStyleBuilder("style2")
	if err != nil {
		t.Fatalf("failed to create builder 2: %v", err)
	}
	textStyle2, err := builder2.
		SetFont("Helvetica").
		SetFontSize("18").
		SetFontColor("1 0 0").
		Build()
	if err != nil {
		t.Fatalf("failed to build text style 2: %v", err)
	}

	// Register text styles
	err = registry.RegisterTextStyle(textStyle1)
	if err != nil {
		t.Errorf("unexpected error registering text style 1: %v", err)
	}

	err = registry.RegisterTextStyle(textStyle2)
	if err != nil {
		t.Errorf("unexpected error registering text style 2: %v", err)
	}

	// Try to register duplicate ID (should fail)
	builder3, err := NewTextStyleBuilder("style1")
	if err != nil {
		t.Fatalf("failed to create builder 3: %v", err)
	}
	textStyle3, err := builder3.
		SetFont("Times").
		SetFontSize("20").
		SetFontColor("0 1 0").
		Build()
	if err != nil {
		t.Fatalf("failed to build text style 3: %v", err)
	}

	err = registry.RegisterTextStyle(textStyle3)
	if err == nil {
		t.Errorf("expected error registering duplicate text style ID")
	}

	// Retrieve text styles
	retrieved1, err := registry.GetTextStyle("style1")
	if err != nil {
		t.Errorf("unexpected error retrieving text style 1: %v", err)
	}
	if retrieved1.ID != "style1" {
		t.Errorf("expected retrieved text style ID 'style1', got %s", retrieved1.ID)
	}

	// Try to retrieve non-existent text style
	_, err = registry.GetTextStyle("nonexistent")
	if err == nil {
		t.Errorf("expected error retrieving non-existent text style")
	}

	// Validate text style reference
	err = registry.ValidateTextStyleReference("style1")
	if err != nil {
		t.Errorf("unexpected error validating existing text style reference: %v", err)
	}

	err = registry.ValidateTextStyleReference("nonexistent")
	if err == nil {
		t.Errorf("expected error validating non-existent text style reference")
	}

	// Get all text styles
	allStyles := registry.GetAllTextStyles()
	if len(allStyles) != 2 {
		t.Errorf("expected 2 text styles, got %d", len(allStyles))
	}
}

func TestTextStyleRegistry_InvalidTextStyle(t *testing.T) {
	registry := NewTextStyleRegistry()

	// Try to register nil text style
	err := registry.RegisterTextStyle(nil)
	if err == nil {
		t.Errorf("expected error registering nil text style")
	}

	// Try to register invalid text style
	invalidStyle := &ValidatedTextStyle{
		ID:   "invalid",
		Font: "Arial",
		// Missing required fields
	}

	err = registry.RegisterTextStyle(invalidStyle)
	if err == nil {
		t.Errorf("expected error registering invalid text style")
	}
}

func TestIsValidTextStyleID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "valid simple ID",
			id:       "ts1",
			expected: true,
		},
		{
			name:     "valid ID with underscore",
			id:       "text_style_1",
			expected: true,
		},
		{
			name:     "valid alphanumeric ID",
			id:       "TextStyle123",
			expected: true,
		},
		{
			name:     "invalid ID starting with number",
			id:       "1textStyle",
			expected: false,
		},
		{
			name:     "invalid ID with dash",
			id:       "text-style",
			expected: false,
		},
		{
			name:     "invalid ID with space",
			id:       "text style",
			expected: false,
		},
		{
			name:     "empty ID",
			id:       "",
			expected: false,
		},
		{
			name:     "invalid ID with special characters",
			id:       "text@style",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTextStyleID(tt.id)
			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}