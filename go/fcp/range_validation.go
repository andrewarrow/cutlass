package fcp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NumericRangeValidator validates numeric values are within reasonable ranges
type NumericRangeValidator struct{}

// NewNumericRangeValidator creates a new numeric range validator
func NewNumericRangeValidator() *NumericRangeValidator {
	return &NumericRangeValidator{}
}

// ValidateOpacity validates opacity values (0.0 to 1.0)
func (nrv *NumericRangeValidator) ValidateOpacity(value string) error {
	if value == "" {
		return nil
	}
	
	opacity, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid opacity value: %s", value)
	}
	
	// Check for special float values
	if math.IsNaN(opacity) || math.IsInf(opacity, 0) {
		return fmt.Errorf("opacity cannot be NaN or infinity: %s", value)
	}
	
	if opacity < 0.0 || opacity > 1.0 {
		return fmt.Errorf("opacity out of range: %.2f (must be 0.0-1.0)", opacity)
	}
	
	return nil
}

// ValidateColorComponent validates a single color component (0.0 to 1.0)
func (nrv *NumericRangeValidator) ValidateColorComponent(component string) error {
	// Block special strings that represent infinity or NaN
	switch strings.ToLower(component) {
	case "∞", "-∞", "infinity", "-infinity", "inf", "-inf", "nan", "null":
		return fmt.Errorf("invalid color component: %s", component)
	}
	
	value, err := strconv.ParseFloat(component, 64)
	if err != nil {
		return fmt.Errorf("invalid color component: %s", component)
	}
	
	// Check for special float values
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("color component cannot be NaN or infinity: %s", component)
	}
	
	if value < 0.0 || value > 1.0 {
		return fmt.Errorf("color component out of range: %.2f (must be 0.0-1.0)", value)
	}
	
	return nil
}

// ValidateColorValue validates RGB or RGBA color values
func (nrv *NumericRangeValidator) ValidateColorValue(color string) error {
	if color == "" {
		return nil
	}
	
	components := strings.Fields(color)
	if len(components) != 3 && len(components) != 4 {
		return fmt.Errorf("invalid color format: %s (must be 'r g b' or 'r g b a')", color)
	}
	
	for i, component := range components {
		if err := nrv.ValidateColorComponent(component); err != nil {
			return fmt.Errorf("color component %d validation failed: %v", i, err)
		}
	}
	
	return nil
}

// ValidateFontSize validates font size values (1.0 to 2000.0 pixels)
func (nrv *NumericRangeValidator) ValidateFontSize(fontSize string) error {
	if fontSize == "" {
		return nil
	}
	
	size, err := strconv.ParseFloat(fontSize, 64)
	if err != nil {
		return fmt.Errorf("invalid font size: %s", fontSize)
	}
	
	// Check for special float values
	if math.IsNaN(size) || math.IsInf(size, 0) {
		return fmt.Errorf("font size cannot be NaN or infinity: %s", fontSize)
	}
	
	if size < 1.0 || size > 500.0 {
		return fmt.Errorf("font size out of range: %.1f (must be 1.0-500.0)", size)
	}
	
	return nil
}

// ValidateLineSpacing validates line spacing values (0.5 to 5.0)
func (nrv *NumericRangeValidator) ValidateLineSpacing(lineSpacing string) error {
	if lineSpacing == "" {
		return nil
	}
	
	spacing, err := strconv.ParseFloat(lineSpacing, 64)
	if err != nil {
		return fmt.Errorf("invalid line spacing: %s", lineSpacing)
	}
	
	// Check for special float values
	if math.IsNaN(spacing) || math.IsInf(spacing, 0) {
		return fmt.Errorf("line spacing cannot be NaN or infinity: %s", lineSpacing)
	}
	
	if spacing < 0.5 || spacing > 5.0 {
		return fmt.Errorf("line spacing out of range: %.2f (must be 0.5-5.0)", spacing)
	}
	
	return nil
}

// ValidateScaleValue validates scale values (0.001 to 100.0)
func (nrv *NumericRangeValidator) ValidateScaleValue(scale string) error {
	if scale == "" {
		return nil
	}
	
	// Handle multi-component scale values like "2.5 2.5"
	components := strings.Fields(scale)
	if len(components) == 0 {
		return fmt.Errorf("empty scale value")
	}
	
	for i, component := range components {
		value, err := strconv.ParseFloat(component, 64)
		if err != nil {
			return fmt.Errorf("invalid scale component %d: %s", i, component)
		}
		
		// Check for special float values
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("scale component %d cannot be NaN or infinity: %s", i, component)
		}
		
		if value < 0.001 || value > 100.0 {
			return fmt.Errorf("scale component %d out of range: %.3f (must be 0.001-100.0)", i, value)
		}
	}
	
	return nil
}

// ValidateRotationValue validates rotation values (-3600 to 3600 degrees)
func (nrv *NumericRangeValidator) ValidateRotationValue(rotation string) error {
	if rotation == "" {
		return nil
	}
	
	value, err := strconv.ParseFloat(rotation, 64)
	if err != nil {
		return fmt.Errorf("invalid rotation value: %s", rotation)
	}
	
	// Check for special float values
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("rotation cannot be NaN or infinity: %s", rotation)
	}
	
	if value < -3600.0 || value > 3600.0 {
		return fmt.Errorf("rotation out of range: %.1f (must be -3600.0 to 3600.0 degrees)", value)
	}
	
	return nil
}

// ValidatePercentageValue validates percentage values (0.0 to 100.0)
func (nrv *NumericRangeValidator) ValidatePercentageValue(percentage string) error {
	if percentage == "" {
		return nil
	}
	
	value, err := strconv.ParseFloat(percentage, 64)
	if err != nil {
		return fmt.Errorf("invalid percentage value: %s", percentage)
	}
	
	// Check for special float values
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("percentage cannot be NaN or infinity: %s", percentage)
	}
	
	if value < 0.0 || value > 100.0 {
		return fmt.Errorf("percentage out of range: %.2f (must be 0.0-100.0)", value)
	}
	
	return nil
}

// ValidateNumericValue validates any numeric value for NaN/infinity
func (nrv *NumericRangeValidator) ValidateNumericValue(value string, fieldName string) error {
	if value == "" {
		return nil
	}
	
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid %s value: %s", fieldName, value)
	}
	
	// Check for special float values
	if math.IsNaN(floatValue) || math.IsInf(floatValue, 0) {
		return fmt.Errorf("%s cannot be NaN or infinity: %s", fieldName, value)
	}
	
	return nil
}