package fcp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// BoundaryValidator validates spatial and temporal boundaries
type BoundaryValidator struct{}

// NewBoundaryValidator creates a new boundary validator
func NewBoundaryValidator() *BoundaryValidator {
	return &BoundaryValidator{}
}

// ValidatePosition validates position coordinates (reasonable screen bounds)
func (bv *BoundaryValidator) ValidatePosition(position string) error {
	if position == "" {
		return nil
	}
	
	components := strings.Fields(position)
	if len(components) != 2 {
		return fmt.Errorf("position must have exactly 2 components (x y): %s", position)
	}
	
	for i, component := range components {
		value, err := strconv.ParseFloat(component, 64)
		if err != nil {
			return fmt.Errorf("invalid position component %d: %s", i, component)
		}
		
		// Check for special float values
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("position component %d cannot be NaN or infinity: %s", i, component)
		}
		
		// Reasonable screen bounds (allow for 8K displays and some offscreen positioning)
		if value < -50000 || value > 50000 {
			return fmt.Errorf("position component %d out of bounds: %.1f (must be -50000 to +50000)", i, value)
		}
	}
	
	return nil
}

// ValidateAnchorPoint validates anchor point coordinates (typically 0.0 to 1.0)
func (bv *BoundaryValidator) ValidateAnchorPoint(anchor string) error {
	if anchor == "" {
		return nil
	}
	
	components := strings.Fields(anchor)
	if len(components) != 2 {
		return fmt.Errorf("anchor point must have exactly 2 components (x y): %s", anchor)
	}
	
	for i, component := range components {
		value, err := strconv.ParseFloat(component, 64)
		if err != nil {
			return fmt.Errorf("invalid anchor component %d: %s", i, component)
		}
		
		// Check for special float values
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("anchor component %d cannot be NaN or infinity: %s", i, component)
		}
		
		// Allow some flexibility beyond 0-1 for advanced positioning
		if value < -5.0 || value > 5.0 {
			return fmt.Errorf("anchor component %d out of bounds: %.2f (must be -5.0 to +5.0)", i, value)
		}
	}
	
	return nil
}

// ValidateLaneNumber validates timeline lane numbers
func (bv *BoundaryValidator) ValidateLaneNumber(lane string) error {
	if lane == "" {
		return nil // Lane is optional
	}
	
	laneNum, err := strconv.Atoi(lane)
	if err != nil {
		return fmt.Errorf("invalid lane number: %s", lane)
	}
	
	// Reasonable lane limits for FCP timeline
	if laneNum < -100 || laneNum > 100 {
		return fmt.Errorf("lane number out of range: %d (must be -100 to +100)", laneNum)
	}
	
	return nil
}

// ValidateTimeOffset validates time offset values (must be non-negative)
func (bv *BoundaryValidator) ValidateTimeOffset(offset string) error {
	if offset == "" {
		return nil
	}
	
	// Parse as Duration to validate format first
	duration := Duration(offset)
	if err := duration.Validate(); err != nil {
		return fmt.Errorf("invalid time offset format: %v", err)
	}
	
	// Extract numeric value for range check
	offsetNoS := strings.TrimSuffix(offset, "s")
	if !strings.Contains(offsetNoS, "/") {
		// Handle simple numeric values
		value, err := strconv.ParseFloat(offsetNoS, 64)
		if err != nil {
			return fmt.Errorf("invalid time offset value: %s", offset)
		}
		
		if value < 0 {
			return fmt.Errorf("negative time offsets not allowed: %s", offset)
		}
		
		return nil
	}
	
	// Handle fractional format (numerator/denominator)
	parts := strings.Split(offsetNoS, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid fractional time offset format: %s", offset)
	}
	
	numerator, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid time offset numerator: %s", parts[0])
	}
	
	denominator, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid time offset denominator: %s", parts[1])
	}
	
	// Don't allow negative timing (would break timeline)
	if numerator < 0 {
		return fmt.Errorf("negative time offsets not allowed: %s", offset)
	}
	
	// Don't allow zero or negative denominators
	if denominator <= 0 {
		return fmt.Errorf("time offset denominator must be positive: %s", offset)
	}
	
	// Check for reasonable time bounds (max ~24 hours)
	maxSeconds := float64(numerator) / float64(denominator)
	if maxSeconds > 86400 {
		return fmt.Errorf("time offset too large: %.2f seconds (max 86400)", maxSeconds)
	}
	
	return nil
}

// ValidateDuration validates duration values
func (bv *BoundaryValidator) ValidateDuration(durationStr string) error {
	if durationStr == "" {
		return nil
	}
	
	// Special case for image durations
	if durationStr == "0s" {
		return nil
	}
	
	// Parse as Duration to validate format
	duration := Duration(durationStr)
	if err := duration.Validate(); err != nil {
		return fmt.Errorf("invalid duration format: %v", err)
	}
	
	// Extract numeric value for range check
	durationNoS := strings.TrimSuffix(durationStr, "s")
	if !strings.Contains(durationNoS, "/") {
		// Handle simple numeric values
		value, err := strconv.ParseFloat(durationNoS, 64)
		if err != nil {
			return fmt.Errorf("invalid duration value: %s", durationStr)
		}
		
		if value < 0 {
			return fmt.Errorf("negative durations not allowed: %s", durationStr)
		}
		
		return nil
	}
	
	// Handle fractional format (numerator/denominator)
	parts := strings.Split(durationNoS, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid fractional duration format: %s", durationStr)
	}
	
	numerator, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid duration numerator: %s", parts[0])
	}
	
	denominator, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid duration denominator: %s", parts[1])
	}
	
	// Don't allow negative duration
	if numerator < 0 {
		return fmt.Errorf("negative durations not allowed: %s", durationStr)
	}
	
	// Don't allow zero or negative denominators
	if denominator <= 0 {
		return fmt.Errorf("duration denominator must be positive: %s", durationStr)
	}
	
	// Check for reasonable duration bounds (max ~24 hours)
	maxSeconds := float64(numerator) / float64(denominator)
	if maxSeconds > 86400 {
		return fmt.Errorf("duration too large: %.2f seconds (max 86400)", maxSeconds)
	}
	
	return nil
}

// ValidateCropValues validates crop rectangle values
func (bv *BoundaryValidator) ValidateCropValues(left, top, right, bottom string) error {
	cropValues := map[string]string{
		"left":   left,
		"top":    top,
		"right":  right,
		"bottom": bottom,
	}
	
	for name, value := range cropValues {
		if value == "" {
			continue
		}
		
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid crop %s value: %s", name, value)
		}
		
		// Check for special float values
		if math.IsNaN(floatValue) || math.IsInf(floatValue, 0) {
			return fmt.Errorf("crop %s cannot be NaN or infinity: %s", name, value)
		}
		
		// Crop values should typically be between -1.0 and 2.0
		if floatValue < -10.0 || floatValue > 10.0 {
			return fmt.Errorf("crop %s out of reasonable bounds: %.2f (should be -10.0 to +10.0)", name, floatValue)
		}
	}
	
	return nil
}

// ValidateZIndex validates z-index values for layering
func (bv *BoundaryValidator) ValidateZIndex(zIndex string) error {
	if zIndex == "" {
		return nil
	}
	
	value, err := strconv.Atoi(zIndex)
	if err != nil {
		return fmt.Errorf("invalid z-index value: %s", zIndex)
	}
	
	// Reasonable z-index bounds
	if value < -1000 || value > 1000 {
		return fmt.Errorf("z-index out of range: %d (must be -1000 to +1000)", value)
	}
	
	return nil
}