// Package keyframe_validation implements Step 8 of the FCPXMLKit-inspired refactoring plan:
// Keyframe attribute validation based on parameter type.
//
// This provides comprehensive keyframe validation that ensures proper attribute usage
// for different parameter types to prevent "param element was ignored" warnings in FCP.
package fcp

import (
	"fmt"
	"strconv"
	"strings"
)

// KeyframeParameterType represents different types of keyframe parameters
type KeyframeParameterType int

const (
	KeyframeParameterUnknown KeyframeParameterType = iota
	KeyframeParameterPosition
	KeyframeParameterScale
	KeyframeParameterRotation
	KeyframeParameterAnchor
	KeyframeParameterOpacity
	KeyframeParameterVolume
	KeyframeParameterColor
	KeyframeParameterCrop
)

// String returns the string representation of the parameter type
func (kpt KeyframeParameterType) String() string {
	switch kpt {
	case KeyframeParameterPosition:
		return "position"
	case KeyframeParameterScale:
		return "scale"
	case KeyframeParameterRotation:
		return "rotation"
	case KeyframeParameterAnchor:
		return "anchor"
	case KeyframeParameterOpacity:
		return "opacity"
	case KeyframeParameterVolume:
		return "volume"
	case KeyframeParameterColor:
		return "color"
	case KeyframeParameterCrop:
		return "crop"
	default:
		return "unknown"
	}
}

// ParseKeyframeParameterType parses a parameter name to determine its type
func ParseKeyframeParameterType(paramName string) KeyframeParameterType {
	switch strings.ToLower(paramName) {
	case "position":
		return KeyframeParameterPosition
	case "scale":
		return KeyframeParameterScale
	case "rotation":
		return KeyframeParameterRotation
	case "anchor":
		return KeyframeParameterAnchor
	case "opacity":
		return KeyframeParameterOpacity
	case "volume":
		return KeyframeParameterVolume
	case "color":
		return KeyframeParameterColor
	case "crop":
		return KeyframeParameterCrop
	default:
		return KeyframeParameterUnknown
	}
}

// KeyframeAttributeRules defines what attributes are allowed for each parameter type
type KeyframeAttributeRules struct {
	AllowInterp      bool
	AllowCurve       bool
	ValidInterpValues []string
	ValidCurveValues  []string
	ValueValidator   func(string) error
}

// ValidatedKeyframe represents a keyframe with validation
type ValidatedKeyframe struct {
	Time   Time
	Value  string
	Interp string
	Curve  string
}

// NewValidatedKeyframe creates a validated keyframe
func NewValidatedKeyframe(time Time, value, interp, curve string) (*ValidatedKeyframe, error) {
	if err := time.Validate(); err != nil {
		return nil, fmt.Errorf("invalid keyframe time: %v", err)
	}
	
	if value == "" {
		return nil, fmt.Errorf("keyframe value cannot be empty")
	}
	
	return &ValidatedKeyframe{
		Time:   time,
		Value:  value,
		Interp: interp,
		Curve:  curve,
	}, nil
}

// KeyframeValidator validates keyframes based on parameter types
type KeyframeValidator struct {
	rules map[KeyframeParameterType]KeyframeAttributeRules
}

// NewKeyframeValidator creates a new keyframe validator with default rules
func NewKeyframeValidator() *KeyframeValidator {
	validator := &KeyframeValidator{
		rules: make(map[KeyframeParameterType]KeyframeAttributeRules),
	}
	
	validator.initializeDefaultRules()
	return validator
}

// initializeDefaultRules sets up the validation rules based on FCPXML specifications
func (kv *KeyframeValidator) initializeDefaultRules() {
	// Position keyframes: NO attributes allowed
	kv.rules[KeyframeParameterPosition] = KeyframeAttributeRules{
		AllowInterp:      false,
		AllowCurve:       false,
		ValidInterpValues: []string{},
		ValidCurveValues:  []string{},
		ValueValidator:   kv.validatePositionValue,
	}
	
	// Scale keyframes: Only curve attribute allowed
	kv.rules[KeyframeParameterScale] = KeyframeAttributeRules{
		AllowInterp:      false,
		AllowCurve:       true,
		ValidInterpValues: []string{},
		ValidCurveValues:  []string{"linear", "smooth", "hold"},
		ValueValidator:   kv.validateScale2DValue,
	}
	
	// Rotation keyframes: Only curve attribute allowed
	kv.rules[KeyframeParameterRotation] = KeyframeAttributeRules{
		AllowInterp:      false,
		AllowCurve:       true,
		ValidInterpValues: []string{},
		ValidCurveValues:  []string{"linear", "smooth", "hold"},
		ValueValidator:   kv.validateSingleFloatValue,
	}
	
	// Anchor keyframes: Only curve attribute allowed
	kv.rules[KeyframeParameterAnchor] = KeyframeAttributeRules{
		AllowInterp:      false,
		AllowCurve:       true,
		ValidInterpValues: []string{},
		ValidCurveValues:  []string{"linear", "smooth", "hold"},
		ValueValidator:   kv.validate2DValue,
	}
	
	// Opacity keyframes: Both interp and curve allowed
	kv.rules[KeyframeParameterOpacity] = KeyframeAttributeRules{
		AllowInterp:      true,
		AllowCurve:       true,
		ValidInterpValues: []string{"linear", "easeIn", "easeOut", "easeInOut"},
		ValidCurveValues:  []string{"linear", "smooth", "hold"},
		ValueValidator:   kv.validateOpacityValue,
	}
	
	// Volume keyframes: Both interp and curve allowed
	kv.rules[KeyframeParameterVolume] = KeyframeAttributeRules{
		AllowInterp:      true,
		AllowCurve:       true,
		ValidInterpValues: []string{"linear", "easeIn", "easeOut", "easeInOut"},
		ValidCurveValues:  []string{"linear", "smooth", "hold"},
		ValueValidator:   kv.validateVolumeValue,
	}
	
	// Color keyframes: Both interp and curve allowed
	kv.rules[KeyframeParameterColor] = KeyframeAttributeRules{
		AllowInterp:      true,
		AllowCurve:       true,
		ValidInterpValues: []string{"linear", "easeIn", "easeOut", "easeInOut"},
		ValidCurveValues:  []string{"linear", "smooth", "hold"},
		ValueValidator:   kv.validateColorValue,
	}
	
	// Crop keyframes: Custom rules for crop parameters
	kv.rules[KeyframeParameterCrop] = KeyframeAttributeRules{
		AllowInterp:      false,
		AllowCurve:       true,
		ValidInterpValues: []string{},
		ValidCurveValues:  []string{"linear", "smooth", "hold"},
		ValueValidator:   kv.validate2DValue,
	}
}

// ValidateKeyframe validates a keyframe for a specific parameter type
func (kv *KeyframeValidator) ValidateKeyframe(paramName string, keyframe *ValidatedKeyframe) error {
	paramType := ParseKeyframeParameterType(paramName)
	
	// Get rules for this parameter type
	rules, exists := kv.rules[paramType]
	if !exists {
		// Unknown parameter - be permissive but validate basic structure
		return kv.validateUnknownParameterKeyframe(paramName, keyframe)
	}
	
	// Validate interp attribute
	if keyframe.Interp != "" {
		if !rules.AllowInterp {
			return fmt.Errorf("%s keyframes cannot have interp attribute", paramName)
		}
		
		if !containsString(rules.ValidInterpValues, keyframe.Interp) {
			return fmt.Errorf("invalid interp value for %s: %s (valid: %v)", 
				paramName, keyframe.Interp, rules.ValidInterpValues)
		}
	}
	
	// Validate curve attribute
	if keyframe.Curve != "" {
		if !rules.AllowCurve {
			return fmt.Errorf("%s keyframes cannot have curve attribute", paramName)
		}
		
		if !containsString(rules.ValidCurveValues, keyframe.Curve) {
			return fmt.Errorf("invalid curve value for %s: %s (valid: %v)", 
				paramName, keyframe.Curve, rules.ValidCurveValues)
		}
	}
	
	// Validate value using parameter-specific validator
	if rules.ValueValidator != nil {
		if err := rules.ValueValidator(keyframe.Value); err != nil {
			return fmt.Errorf("invalid value for %s keyframe: %v", paramName, err)
		}
	}
	
	return nil
}

// ValidateKeyframeSequence validates a sequence of keyframes for chronological order
func (kv *KeyframeValidator) ValidateKeyframeSequence(paramName string, keyframes []*ValidatedKeyframe) error {
	if len(keyframes) == 0 {
		return fmt.Errorf("keyframe sequence cannot be empty")
	}
	
	// Validate individual keyframes
	for i, keyframe := range keyframes {
		if err := kv.ValidateKeyframe(paramName, keyframe); err != nil {
			return fmt.Errorf("keyframe %d validation failed: %v", i, err)
		}
	}
	
	// Validate chronological order
	for i := 1; i < len(keyframes); i++ {
		prevTime := keyframes[i-1].Time
		currTime := keyframes[i].Time
		
		comparison, err := CompareTimes(prevTime, currTime)
		if err != nil {
			return fmt.Errorf("failed to compare keyframe times: %v", err)
		}
		
		if comparison >= 0 {
			return fmt.Errorf("keyframes must be in chronological order: keyframe %d (%s) is not after keyframe %d (%s)", 
				i, currTime, i-1, prevTime)
		}
	}
	
	return nil
}

// validateUnknownParameterKeyframe validates basic structure for unknown parameters
func (kv *KeyframeValidator) validateUnknownParameterKeyframe(paramName string, keyframe *ValidatedKeyframe) error {
	// For unknown parameters, just validate that attributes are reasonable
	if keyframe.Interp != "" {
		validInterps := []string{"linear", "easeIn", "easeOut", "easeInOut"}
		if !containsString(validInterps, keyframe.Interp) {
			return fmt.Errorf("unknown interp value for %s: %s", paramName, keyframe.Interp)
		}
	}
	
	if keyframe.Curve != "" {
		validCurves := []string{"linear", "smooth", "hold"}
		if !containsString(validCurves, keyframe.Curve) {
			return fmt.Errorf("unknown curve value for %s: %s", paramName, keyframe.Curve)
		}
	}
	
	return nil
}

// Value validators for different parameter types

// validatePositionValue validates position values (should be "x y" format)
func (kv *KeyframeValidator) validatePositionValue(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return fmt.Errorf("position value must have 2 components (x y): %s", value)
	}
	
	for i, part := range parts {
		if _, err := strconv.ParseFloat(part, 64); err != nil {
			return fmt.Errorf("invalid position component %d: %s", i, part)
		}
	}
	
	return nil
}

// validateScale2DValue validates 2D scale values (should be "sx sy" format)
func (kv *KeyframeValidator) validateScale2DValue(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return fmt.Errorf("scale value must have 2 components (sx sy): %s", value)
	}
	
	for i, part := range parts {
		scale, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return fmt.Errorf("invalid scale component %d: %s", i, part)
		}
		
		// Validate reasonable scale values
		if scale < 0 {
			return fmt.Errorf("scale cannot be negative: %f", scale)
		}
		
		if scale > 100 {
			return fmt.Errorf("scale value seems too large: %f", scale)
		}
	}
	
	return nil
}

// validateSingleFloatValue validates single float values (rotation, etc.)
func (kv *KeyframeValidator) validateSingleFloatValue(value string) error {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float value: %s", value)
	}
	
	return nil
}

// validate2DValue validates 2D values (should be "x y" format)
func (kv *KeyframeValidator) validate2DValue(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 2 {
		return fmt.Errorf("2D value must have 2 components: %s", value)
	}
	
	for i, part := range parts {
		if _, err := strconv.ParseFloat(part, 64); err != nil {
			return fmt.Errorf("invalid 2D component %d: %s", i, part)
		}
	}
	
	return nil
}

// validateOpacityValue validates opacity values (0.0 to 1.0)
func (kv *KeyframeValidator) validateOpacityValue(value string) error {
	opacity, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid opacity value: %s", value)
	}
	
	if opacity < 0.0 || opacity > 1.0 {
		return fmt.Errorf("opacity must be between 0.0 and 1.0: %f", opacity)
	}
	
	return nil
}

// validateVolumeValue validates volume values (supports dB notation)
func (kv *KeyframeValidator) validateVolumeValue(value string) error {
	if strings.HasSuffix(value, "dB") {
		// Parse dB value
		dbStr := strings.TrimSuffix(value, "dB")
		db, err := strconv.ParseFloat(dbStr, 64)
		if err != nil {
			return fmt.Errorf("invalid dB value: %s", value)
		}
		
		// Reasonable dB range
		if db < -60.0 || db > 20.0 {
			return fmt.Errorf("dB value out of reasonable range [-60, 20]: %f", db)
		}
	} else {
		// Parse as linear multiplier
		multiplier, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid volume multiplier: %s", value)
		}
		
		if multiplier < 0.0 {
			return fmt.Errorf("volume multiplier cannot be negative: %f", multiplier)
		}
	}
	
	return nil
}

// validateColorValue validates color values (RGBA format)
func (kv *KeyframeValidator) validateColorValue(value string) error {
	parts := strings.Fields(value)
	if len(parts) != 3 && len(parts) != 4 {
		return fmt.Errorf("color value must have 3 or 4 components (RGB or RGBA): %s", value)
	}
	
	for i, part := range parts {
		component, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return fmt.Errorf("invalid color component %d: %s", i, part)
		}
		
		if component < 0.0 || component > 1.0 {
			return fmt.Errorf("color component %d out of range [0,1]: %f", i, component)
		}
	}
	
	return nil
}

// KeyframeBuilder provides a safe way to build keyframe animations
type KeyframeBuilder struct {
	paramName string
	validator *KeyframeValidator
	keyframes []*ValidatedKeyframe
}

// NewKeyframeBuilder creates a new keyframe builder for a parameter
func NewKeyframeBuilder(paramName string) *KeyframeBuilder {
	return &KeyframeBuilder{
		paramName: paramName,
		validator: NewKeyframeValidator(),
		keyframes: make([]*ValidatedKeyframe, 0),
	}
}

// AddKeyframe adds a keyframe with validation
func (kb *KeyframeBuilder) AddKeyframe(time Time, value string, options ...KeyframeOption) error {
	keyframe := &ValidatedKeyframe{
		Time:  time,
		Value: value,
	}
	
	// Apply options
	for _, option := range options {
		option(keyframe)
	}
	
	// Validate keyframe
	if err := kb.validator.ValidateKeyframe(kb.paramName, keyframe); err != nil {
		return fmt.Errorf("keyframe validation failed: %v", err)
	}
	
	kb.keyframes = append(kb.keyframes, keyframe)
	return nil
}

// Build creates the final keyframe sequence with validation
func (kb *KeyframeBuilder) Build() ([]*ValidatedKeyframe, error) {
	if len(kb.keyframes) == 0 {
		return nil, fmt.Errorf("keyframe sequence cannot be empty")
	}
	
	// Validate the complete sequence
	if err := kb.validator.ValidateKeyframeSequence(kb.paramName, kb.keyframes); err != nil {
		return nil, fmt.Errorf("keyframe sequence validation failed: %v", err)
	}
	
	return kb.keyframes, nil
}

// KeyframeOption configures keyframe attributes
type KeyframeOption func(*ValidatedKeyframe)

// WithInterp sets the interpolation type
func WithInterp(interp string) KeyframeOption {
	return func(k *ValidatedKeyframe) {
		k.Interp = interp
	}
}

// WithCurve sets the curve type
func WithCurve(curve string) KeyframeOption {
	return func(k *ValidatedKeyframe) {
		k.Curve = curve
	}
}

// KeyframeSequenceValidator validates complete parameter animations
type KeyframeSequenceValidator struct {
	validator *KeyframeValidator
}

// NewKeyframeSequenceValidator creates a new sequence validator
func NewKeyframeSequenceValidator() *KeyframeSequenceValidator {
	return &KeyframeSequenceValidator{
		validator: NewKeyframeValidator(),
	}
}

// ValidateParameterAnimation validates a complete parameter animation
func (ksv *KeyframeSequenceValidator) ValidateParameterAnimation(paramName string, keyframes []*ValidatedKeyframe) error {
	return ksv.validator.ValidateKeyframeSequence(paramName, keyframes)
}

// ValidateMultipleParameters validates multiple parameter animations together
func (ksv *KeyframeSequenceValidator) ValidateMultipleParameters(animations map[string][]*ValidatedKeyframe) error {
	for paramName, keyframes := range animations {
		if err := ksv.ValidateParameterAnimation(paramName, keyframes); err != nil {
			return fmt.Errorf("parameter %s validation failed: %v", paramName, err)
		}
	}
	
	return nil
}

// containsString checks if a slice contains a string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}