// Package validation_engine implements Step 2 of the FCPXMLKit-inspired refactoring plan:
// Struct tag-based validation system using reflection for automatic validation.
//
// This provides a validation engine similar to FCPXMLKit's property wrappers
// but implemented using Go's struct tags and reflection.
package fcp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Value   interface{}
	Rule    string
	Message string
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %s: %s (rule: %s, value: %v)", 
		ve.Field, ve.Message, ve.Rule, ve.Value)
}

// StructValidator provides comprehensive struct validation using reflection
type StructValidator struct {
	referenceRegistry map[string]bool // For validating references
}

// NewStructValidator creates a new struct validator
func NewStructValidator() *StructValidator {
	return &StructValidator{
		referenceRegistry: make(map[string]bool),
	}
}

// RegisterReference registers a valid reference ID for reference validation
func (sv *StructValidator) RegisterReference(refType, id string) {
	key := fmt.Sprintf("%s:%s", refType, id)
	sv.referenceRegistry[key] = true
}

// ValidateStruct validates a struct using reflection and struct tags
func (sv *StructValidator) ValidateStruct(v interface{}) error {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	
	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("cannot validate nil pointer")
		}
		val = val.Elem()
		typ = typ.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("can only validate struct types, got %T", v)
	}
	
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}
		
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue // No validation rules for this field
		}
		
		if err := sv.validateField(field, fieldType, tag); err != nil {
			return ValidationError{
				Field:   fieldType.Name,
				Value:   field.Interface(),
				Rule:    tag,
				Message: err.Error(),
			}
		}
	}
	
	return nil
}

// validateField validates a single field based on its validation tag
func (sv *StructValidator) validateField(field reflect.Value, fieldType reflect.StructField, tag string) error {
	rules := parseValidationTag(tag)
	
	for _, rule := range rules {
		if err := sv.applyValidationRule(field, fieldType, rule); err != nil {
			return err
		}
	}
	
	return nil
}

// ValidationRule represents a single validation rule parsed from a tag
type ValidationRule struct {
	Name   string
	Params []string
}

// parseValidationTag parses a validation tag into individual rules
func parseValidationTag(tag string) []ValidationRule {
	var rules []ValidationRule
	
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Parse rule name and parameters
		if colonIndex := strings.Index(part, ":"); colonIndex != -1 {
			ruleName := part[:colonIndex]
			paramStr := part[colonIndex+1:]
			params := strings.Split(paramStr, " ")
			rules = append(rules, ValidationRule{Name: ruleName, Params: params})
		} else {
			rules = append(rules, ValidationRule{Name: part, Params: nil})
		}
	}
	
	return rules
}

// applyValidationRule applies a single validation rule to a field
func (sv *StructValidator) applyValidationRule(field reflect.Value, fieldType reflect.StructField, rule ValidationRule) error {
	switch rule.Name {
	case "required":
		return sv.validateRequired(field, fieldType)
	case "optional":
		return nil // Optional fields don't need validation if empty
	case "reference":
		return sv.validateReference(field, fieldType, rule.Params)
	case "min":
		return sv.validateMin(field, fieldType, rule.Params)
	case "max":
		return sv.validateMax(field, fieldType, rule.Params)
	case "oneof":
		return sv.validateOneOf(field, fieldType, rule.Params)
	case "fcpxml_id":
		return sv.validateFCPXMLID(field, fieldType)
	case "fcpxml_duration":
		return sv.validateFCPXMLDuration(field, fieldType)
	case "fcpxml_time":
		return sv.validateFCPXMLTime(field, fieldType)
	case "fcpxml_lane":
		return sv.validateFCPXMLLane(field, fieldType)
	case "media_type":
		return sv.validateMediaType(field, fieldType, rule.Params)
	case "color_space":
		return sv.validateColorSpace(field, fieldType)
	case "audio_rate":
		return sv.validateAudioRate(field, fieldType)
	default:
		return fmt.Errorf("unknown validation rule: %s", rule.Name)
	}
}

// validateRequired ensures the field is not empty/zero value
func (sv *StructValidator) validateRequired(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return fmt.Errorf("required field cannot be nil")
	}
	
	if field.Kind() == reflect.String && field.String() == "" {
		return fmt.Errorf("required field cannot be empty")
	}
	
	if field.Kind() == reflect.Slice && field.Len() == 0 {
		return fmt.Errorf("required slice cannot be empty")
	}
	
	return nil
}

// validateReference validates that a reference field points to a valid resource
func (sv *StructValidator) validateReference(field reflect.Value, fieldType reflect.StructField, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("reference rule requires reference type parameter")
	}
	
	refType := params[0]
	
	if field.Kind() != reflect.String {
		return fmt.Errorf("reference field must be string, got %s", field.Kind())
	}
	
	refValue := field.String()
	if refValue == "" {
		return nil // Empty reference is allowed (optional reference)
	}
	
	key := fmt.Sprintf("%s:%s", refType, refValue)
	if !sv.referenceRegistry[key] {
		return fmt.Errorf("dangling reference: %s (type: %s)", refValue, refType)
	}
	
	return nil
}

// validateMin validates minimum value constraints
func (sv *StructValidator) validateMin(field reflect.Value, fieldType reflect.StructField, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("min rule requires minimum value parameter")
	}
	
	minValue, err := strconv.Atoi(params[0])
	if err != nil {
		return fmt.Errorf("min rule parameter must be integer: %s", params[0])
	}
	
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() < int64(minValue) {
			return fmt.Errorf("value %d is less than minimum %d", field.Int(), minValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() < uint64(minValue) {
			return fmt.Errorf("value %d is less than minimum %d", field.Uint(), minValue)
		}
	case reflect.Slice:
		if field.Len() < minValue {
			return fmt.Errorf("slice length %d is less than minimum %d", field.Len(), minValue)
		}
	default:
		return fmt.Errorf("min rule not applicable to type %s", field.Kind())
	}
	
	return nil
}

// validateMax validates maximum value constraints
func (sv *StructValidator) validateMax(field reflect.Value, fieldType reflect.StructField, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("max rule requires maximum value parameter")
	}
	
	maxValue, err := strconv.Atoi(params[0])
	if err != nil {
		return fmt.Errorf("max rule parameter must be integer: %s", params[0])
	}
	
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() > int64(maxValue) {
			return fmt.Errorf("value %d is greater than maximum %d", field.Int(), maxValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() > uint64(maxValue) {
			return fmt.Errorf("value %d is greater than maximum %d", field.Uint(), maxValue)
		}
	case reflect.Slice:
		if field.Len() > maxValue {
			return fmt.Errorf("slice length %d is greater than maximum %d", field.Len(), maxValue)
		}
	default:
		return fmt.Errorf("max rule not applicable to type %s", field.Kind())
	}
	
	return nil
}

// validateOneOf validates that the field value is one of the allowed values
func (sv *StructValidator) validateOneOf(field reflect.Value, fieldType reflect.StructField, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("oneof rule requires at least one allowed value")
	}
	
	if field.Kind() != reflect.String {
		return fmt.Errorf("oneof rule only applicable to string fields")
	}
	
	value := field.String()
	for _, allowedValue := range params {
		if value == allowedValue {
			return nil
		}
	}
	
	return fmt.Errorf("value %s is not one of allowed values: %v", value, params)
}

// validateFCPXMLID validates an FCPXML ID field
func (sv *StructValidator) validateFCPXMLID(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("fcpxml_id rule only applicable to string fields")
	}
	
	id := ID(field.String())
	return id.Validate()
}

// validateFCPXMLDuration validates an FCPXML duration field
func (sv *StructValidator) validateFCPXMLDuration(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("fcpxml_duration rule only applicable to string fields")
	}
	
	duration := Duration(field.String())
	return duration.Validate()
}

// validateFCPXMLTime validates an FCPXML time field
func (sv *StructValidator) validateFCPXMLTime(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("fcpxml_time rule only applicable to string fields")
	}
	
	time := Time(field.String())
	return time.Validate()
}

// validateFCPXMLLane validates an FCPXML lane field
func (sv *StructValidator) validateFCPXMLLane(field reflect.Value, fieldType reflect.StructField) error {
	var lane Lane
	
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		lane = Lane(field.Int())
	case reflect.String:
		if field.String() == "" {
			return nil // Empty string is valid (lane 0)
		}
		laneNum, err := strconv.Atoi(field.String())
		if err != nil {
			return fmt.Errorf("invalid lane string: %s", field.String())
		}
		lane = Lane(laneNum)
	default:
		return fmt.Errorf("fcpxml_lane rule only applicable to int or string fields")
	}
	
	return lane.Validate()
}

// validateMediaType validates a media type field with constraints
func (sv *StructValidator) validateMediaType(field reflect.Value, fieldType reflect.StructField, params []string) error {
	if len(params) == 0 {
		return fmt.Errorf("media_type rule requires media type parameter")
	}
	
	expectedType := params[0]
	
	// This would integrate with media type detection logic
	// For now, we'll validate that the field contains expected attributes
	// based on the media type
	
	switch expectedType {
	case "image":
		// Images should not have audio properties
		return sv.validateImageConstraints(field, fieldType)
	case "video":
		// Videos can have both video and audio properties
		return sv.validateVideoConstraints(field, fieldType)
	case "audio":
		// Audio should not have video properties
		return sv.validateAudioConstraints(field, fieldType)
	default:
		return fmt.Errorf("unknown media type: %s", expectedType)
	}
}

// validateImageConstraints validates constraints specific to image assets
func (sv *StructValidator) validateImageConstraints(field reflect.Value, fieldType reflect.StructField) error {
	// Implementation would check that duration is "0s", no audio properties, etc.
	// This is a placeholder for media type specific validation
	return nil
}

// validateVideoConstraints validates constraints specific to video assets
func (sv *StructValidator) validateVideoConstraints(field reflect.Value, fieldType reflect.StructField) error {
	// Implementation would check video-specific constraints
	// This is a placeholder for media type specific validation
	return nil
}

// validateAudioConstraints validates constraints specific to audio assets
func (sv *StructValidator) validateAudioConstraints(field reflect.Value, fieldType reflect.StructField) error {
	// Implementation would check that no video properties exist, etc.
	// This is a placeholder for media type specific validation
	return nil
}

// validateColorSpace validates an FCPXML color space field
func (sv *StructValidator) validateColorSpace(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("color_space rule only applicable to string fields")
	}
	
	colorSpace := ColorSpace(field.String())
	return colorSpace.Validate()
}

// validateAudioRate validates an FCPXML audio rate field
func (sv *StructValidator) validateAudioRate(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("audio_rate rule only applicable to string fields")
	}
	
	audioRate := AudioRate(field.String())
	return audioRate.Validate()
}

// ValidateResource validates a resource using its type-specific validation
type ResourceValidator interface {
	ValidateAsResource() error
}

// ValidateResource validates a resource that implements ResourceValidator
func (sv *StructValidator) ValidateResource(resource ResourceValidator) error {
	// First validate using struct tags
	if err := sv.ValidateStruct(resource); err != nil {
		return err
	}
	
	// Then validate using resource-specific validation
	return resource.ValidateAsResource()
}