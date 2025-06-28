// Package animation_builder implements Step 15 of the FCPXMLKit-inspired refactoring plan:
// Create Animation Builder with Keyframe Validation.
//
// This provides safe animation building with automatic keyframe validation that ensures
// proper attribute usage for different parameter types and prevents common FCPXML
// animation generation errors.
package fcp

import (
	"fmt"
	"sort"
)

// AnimationBuilder provides safe animation building with automatic keyframe validation
type AnimationBuilder struct {
	paramName string
	validator *KeyframeValidator
	keyframes []*ValidatedKeyframe
}

// NewAnimationBuilder creates a new animation builder for a specific parameter
func NewAnimationBuilder(paramName string) *AnimationBuilder {
	return &AnimationBuilder{
		paramName: paramName,
		validator: NewKeyframeValidator(),
		keyframes: make([]*ValidatedKeyframe, 0),
	}
}

// AddKeyframe adds a keyframe with validation
func (ab *AnimationBuilder) AddKeyframe(time Time, value string, options ...KeyframeOption) error {
	keyframe, err := NewValidatedKeyframe(time, value, "", "")
	if err != nil {
		return fmt.Errorf("failed to create keyframe: %v", err)
	}
	
	// Apply options
	for _, option := range options {
		option(keyframe)
	}
	
	// Validate keyframe for this parameter type
	if err := ab.validator.ValidateKeyframe(ab.paramName, keyframe); err != nil {
		return fmt.Errorf("keyframe validation failed: %v", err)
	}
	
	ab.keyframes = append(ab.keyframes, keyframe)
	return nil
}

// AddLinearKeyframes adds a sequence of keyframes with linear interpolation
func (ab *AnimationBuilder) AddLinearKeyframes(keyframes []KeyframeData) error {
	for i, kf := range keyframes {
		var options []KeyframeOption
		
		// Add appropriate options based on parameter type
		paramType := ParseKeyframeParameterType(ab.paramName)
		switch paramType {
		case KeyframeParameterScale, KeyframeParameterRotation, KeyframeParameterAnchor:
			options = append(options, WithCurve("linear"))
		case KeyframeParameterOpacity, KeyframeParameterVolume:
			options = append(options, WithInterp("linear"), WithCurve("linear"))
		// Position keyframes don't get any options (no interp/curve allowed)
		}
		
		if err := ab.AddKeyframe(kf.Time, kf.Value, options...); err != nil {
			return fmt.Errorf("failed to add keyframe %d: %v", i, err)
		}
	}
	
	return nil
}

// AddSmoothKeyframes adds a sequence of keyframes with smooth interpolation
func (ab *AnimationBuilder) AddSmoothKeyframes(keyframes []KeyframeData) error {
	for i, kf := range keyframes {
		var options []KeyframeOption
		
		// Add appropriate options based on parameter type
		paramType := ParseKeyframeParameterType(ab.paramName)
		switch paramType {
		case KeyframeParameterScale, KeyframeParameterRotation, KeyframeParameterAnchor:
			options = append(options, WithCurve("smooth"))
		case KeyframeParameterOpacity, KeyframeParameterVolume:
			options = append(options, WithInterp("easeInOut"), WithCurve("smooth"))
		// Position keyframes don't get any options (no interp/curve allowed)
		}
		
		if err := ab.AddKeyframe(kf.Time, kf.Value, options...); err != nil {
			return fmt.Errorf("failed to add keyframe %d: %v", i, err)
		}
	}
	
	return nil
}

// Build creates the final parameter with validated keyframe sequence
func (ab *AnimationBuilder) Build() (*Param, error) {
	if len(ab.keyframes) == 0 {
		return nil, fmt.Errorf("animation must have at least one keyframe")
	}
	
	// Sort keyframes by time first
	sort.Slice(ab.keyframes, func(i, j int) bool {
		timeI, _ := ab.keyframes[i].Time.ToSeconds()
		timeJ, _ := ab.keyframes[j].Time.ToSeconds()
		return timeI < timeJ
	})
	
	// Then validate the complete sequence
	if err := ab.validator.ValidateKeyframeSequence(ab.paramName, ab.keyframes); err != nil {
		return nil, fmt.Errorf("keyframe sequence validation failed: %v", err)
	}
	
	// Convert to FCPXML keyframes
	fcpKeyframes := make([]Keyframe, len(ab.keyframes))
	for i, vk := range ab.keyframes {
		fcpKeyframes[i] = Keyframe{
			Time:   vk.Time.String(),
			Value:  vk.Value,
			Interp: vk.Interp,
			Curve:  vk.Curve,
		}
	}
	
	return &Param{
		Name: ab.paramName,
		KeyframeAnimation: &KeyframeAnimation{
			Keyframes: fcpKeyframes,
		},
	}, nil
}

// GetKeyframeCount returns the number of keyframes in the animation
func (ab *AnimationBuilder) GetKeyframeCount() int {
	return len(ab.keyframes)
}

// KeyframeData represents keyframe timing and value data
type KeyframeData struct {
	Time  Time
	Value string
}

// TransformBuilder provides safe transform building with automatic parameter validation
type TransformBuilder struct {
	positionAnimation *AnimationBuilder
	scaleAnimation    *AnimationBuilder
	rotationAnimation *AnimationBuilder
	anchorAnimation   *AnimationBuilder
	staticPosition    string
	staticScale       string
	staticRotation    string
	staticAnchor      string
}

// NewTransformBuilder creates a new transform builder
func NewTransformBuilder() *TransformBuilder {
	return &TransformBuilder{}
}

// AddPositionAnimation adds position keyframe animation
func (tb *TransformBuilder) AddPositionAnimation(keyframes []KeyframeData) error {
	tb.positionAnimation = NewAnimationBuilder("position")
	return tb.positionAnimation.AddLinearKeyframes(keyframes)
}

// AddPositionAnimationSmooth adds position keyframe animation with smooth interpolation
func (tb *TransformBuilder) AddPositionAnimationSmooth(keyframes []KeyframeData) error {
	tb.positionAnimation = NewAnimationBuilder("position")
	// Note: Position keyframes don't support curve/interp, so this behaves the same as linear
	return tb.positionAnimation.AddLinearKeyframes(keyframes)
}

// AddScaleAnimation adds scale keyframe animation
func (tb *TransformBuilder) AddScaleAnimation(keyframes []KeyframeData, interpolation string) error {
	tb.scaleAnimation = NewAnimationBuilder("scale")
	
	switch interpolation {
	case "linear":
		return tb.scaleAnimation.AddLinearKeyframes(keyframes)
	case "smooth":
		return tb.scaleAnimation.AddSmoothKeyframes(keyframes)
	default:
		return fmt.Errorf("invalid interpolation type for scale: %s (use 'linear' or 'smooth')", interpolation)
	}
}

// AddRotationAnimation adds rotation keyframe animation
func (tb *TransformBuilder) AddRotationAnimation(keyframes []KeyframeData, interpolation string) error {
	tb.rotationAnimation = NewAnimationBuilder("rotation")
	
	switch interpolation {
	case "linear":
		return tb.rotationAnimation.AddLinearKeyframes(keyframes)
	case "smooth":
		return tb.rotationAnimation.AddSmoothKeyframes(keyframes)
	default:
		return fmt.Errorf("invalid interpolation type for rotation: %s (use 'linear' or 'smooth')", interpolation)
	}
}

// AddAnchorAnimation adds anchor point keyframe animation
func (tb *TransformBuilder) AddAnchorAnimation(keyframes []KeyframeData, interpolation string) error {
	tb.anchorAnimation = NewAnimationBuilder("anchor")
	
	switch interpolation {
	case "linear":
		return tb.anchorAnimation.AddLinearKeyframes(keyframes)
	case "smooth":
		return tb.anchorAnimation.AddSmoothKeyframes(keyframes)
	default:
		return fmt.Errorf("invalid interpolation type for anchor: %s (use 'linear' or 'smooth')", interpolation)
	}
}

// SetStaticPosition sets a static position value (no animation)
func (tb *TransformBuilder) SetStaticPosition(position string) error {
	// Validate position format
	validator := NewKeyframeValidator()
	dummyKeyframe := &ValidatedKeyframe{
		Time:  Time("0s"),
		Value: position,
	}
	
	if err := validator.ValidateKeyframe("position", dummyKeyframe); err != nil {
		return fmt.Errorf("invalid static position: %v", err)
	}
	
	tb.staticPosition = position
	return nil
}

// SetStaticScale sets a static scale value (no animation)
func (tb *TransformBuilder) SetStaticScale(scale string) error {
	// Validate scale format
	validator := NewKeyframeValidator()
	dummyKeyframe := &ValidatedKeyframe{
		Time:  Time("0s"),
		Value: scale,
	}
	
	if err := validator.ValidateKeyframe("scale", dummyKeyframe); err != nil {
		return fmt.Errorf("invalid static scale: %v", err)
	}
	
	tb.staticScale = scale
	return nil
}

// SetStaticRotation sets a static rotation value (no animation)
func (tb *TransformBuilder) SetStaticRotation(rotation string) error {
	// Validate rotation format
	validator := NewKeyframeValidator()
	dummyKeyframe := &ValidatedKeyframe{
		Time:  Time("0s"),
		Value: rotation,
	}
	
	if err := validator.ValidateKeyframe("rotation", dummyKeyframe); err != nil {
		return fmt.Errorf("invalid static rotation: %v", err)
	}
	
	tb.staticRotation = rotation
	return nil
}

// SetStaticAnchor sets a static anchor point value (no animation)
func (tb *TransformBuilder) SetStaticAnchor(anchor string) error {
	// Validate anchor format
	validator := NewKeyframeValidator()
	dummyKeyframe := &ValidatedKeyframe{
		Time:  Time("0s"),
		Value: anchor,
	}
	
	if err := validator.ValidateKeyframe("anchor", dummyKeyframe); err != nil {
		return fmt.Errorf("invalid static anchor: %v", err)
	}
	
	tb.staticAnchor = anchor
	return nil
}

// Build creates the final AdjustTransform with validated parameters
func (tb *TransformBuilder) Build() (*AdjustTransform, error) {
	transform := &AdjustTransform{
		Params: make([]Param, 0),
	}
	
	// Add position (animation takes precedence over static)
	if tb.positionAnimation != nil {
		positionParam, err := tb.positionAnimation.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build position animation: %v", err)
		}
		transform.Params = append(transform.Params, *positionParam)
	} else if tb.staticPosition != "" {
		transform.Position = tb.staticPosition
	}
	
	// Add scale (animation takes precedence over static)
	if tb.scaleAnimation != nil {
		scaleParam, err := tb.scaleAnimation.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build scale animation: %v", err)
		}
		transform.Params = append(transform.Params, *scaleParam)
	} else if tb.staticScale != "" {
		transform.Scale = tb.staticScale
	}
	
	// Add rotation animation
	if tb.rotationAnimation != nil {
		rotationParam, err := tb.rotationAnimation.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build rotation animation: %v", err)
		}
		transform.Params = append(transform.Params, *rotationParam)
	}
	
	// Add anchor animation
	if tb.anchorAnimation != nil {
		anchorParam, err := tb.anchorAnimation.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build anchor animation: %v", err)
		}
		transform.Params = append(transform.Params, *anchorParam)
	}
	
	return transform, nil
}

// GetAnimationCount returns the number of animated parameters
func (tb *TransformBuilder) GetAnimationCount() int {
	count := 0
	if tb.positionAnimation != nil {
		count++
	}
	if tb.scaleAnimation != nil {
		count++
	}
	if tb.rotationAnimation != nil {
		count++
	}
	if tb.anchorAnimation != nil {
		count++
	}
	return count
}

// OpacityAnimationBuilder provides safe opacity animation building
type OpacityAnimationBuilder struct {
	animation *AnimationBuilder
}

// NewOpacityAnimationBuilder creates a new opacity animation builder
func NewOpacityAnimationBuilder() *OpacityAnimationBuilder {
	return &OpacityAnimationBuilder{
		animation: NewAnimationBuilder("opacity"),
	}
}

// AddKeyframes adds opacity keyframes with specified interpolation
func (oab *OpacityAnimationBuilder) AddKeyframes(keyframes []KeyframeData, interpolation string) error {
	switch interpolation {
	case "linear":
		return oab.animation.AddLinearKeyframes(keyframes)
	case "smooth", "easeInOut":
		return oab.animation.AddSmoothKeyframes(keyframes)
	case "easeIn":
		for i, kf := range keyframes {
			if err := oab.animation.AddKeyframe(kf.Time, kf.Value, WithInterp("easeIn"), WithCurve("smooth")); err != nil {
				return fmt.Errorf("failed to add keyframe %d: %v", i, err)
			}
		}
		return nil
	case "easeOut":
		for i, kf := range keyframes {
			if err := oab.animation.AddKeyframe(kf.Time, kf.Value, WithInterp("easeOut"), WithCurve("smooth")); err != nil {
				return fmt.Errorf("failed to add keyframe %d: %v", i, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("invalid interpolation type for opacity: %s", interpolation)
	}
}

// Build creates the final opacity parameter
func (oab *OpacityAnimationBuilder) Build() (*Param, error) {
	return oab.animation.Build()
}

// KenBurnsAnimationBuilder provides specialized Ken Burns effect building
type KenBurnsAnimationBuilder struct {
	startTime     Time
	duration      Duration
	startPosition string
	endPosition   string
	startScale    string
	endScale      string
}

// NewKenBurnsAnimationBuilder creates a Ken Burns animation builder
func NewKenBurnsAnimationBuilder(startTime Time, duration Duration) *KenBurnsAnimationBuilder {
	return &KenBurnsAnimationBuilder{
		startTime: startTime,
		duration:  duration,
		// Default Ken Burns settings
		startPosition: "0 0",
		endPosition:   "-20 -15",
		startScale:    "1.2 1.2",
		endScale:      "1.35 1.35",
	}
}

// SetPositionRange sets the position movement range
func (kbab *KenBurnsAnimationBuilder) SetPositionRange(startPos, endPos string) error {
	// Validate position values
	validator := NewKeyframeValidator()
	
	startKeyframe := &ValidatedKeyframe{Time: kbab.startTime, Value: startPos}
	if err := validator.ValidateKeyframe("position", startKeyframe); err != nil {
		return fmt.Errorf("invalid start position: %v", err)
	}
	
	endKeyframe := &ValidatedKeyframe{Time: kbab.startTime, Value: endPos}
	if err := validator.ValidateKeyframe("position", endKeyframe); err != nil {
		return fmt.Errorf("invalid end position: %v", err)
	}
	
	kbab.startPosition = startPos
	kbab.endPosition = endPos
	return nil
}

// SetScaleRange sets the scale zoom range
func (kbab *KenBurnsAnimationBuilder) SetScaleRange(startScale, endScale string) error {
	// Validate scale values
	validator := NewKeyframeValidator()
	
	startKeyframe := &ValidatedKeyframe{Time: kbab.startTime, Value: startScale}
	if err := validator.ValidateKeyframe("scale", startKeyframe); err != nil {
		return fmt.Errorf("invalid start scale: %v", err)
	}
	
	endKeyframe := &ValidatedKeyframe{Time: kbab.startTime, Value: endScale}
	if err := validator.ValidateKeyframe("scale", endKeyframe); err != nil {
		return fmt.Errorf("invalid end scale: %v", err)
	}
	
	kbab.startScale = startScale
	kbab.endScale = endScale
	return nil
}

// Build creates the complete Ken Burns transform
func (kbab *KenBurnsAnimationBuilder) Build() (*AdjustTransform, error) {
	// Calculate end time
	endTime, err := AddTimes(kbab.startTime, Time(kbab.duration))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate end time: %v", err)
	}
	
	// Create transform builder
	builder := NewTransformBuilder()
	
	// Add position animation
	positionKeyframes := []KeyframeData{
		{Time: kbab.startTime, Value: kbab.startPosition},
		{Time: endTime, Value: kbab.endPosition},
	}
	if err := builder.AddPositionAnimation(positionKeyframes); err != nil {
		return nil, fmt.Errorf("failed to add position animation: %v", err)
	}
	
	// Add scale animation
	scaleKeyframes := []KeyframeData{
		{Time: kbab.startTime, Value: kbab.startScale},
		{Time: endTime, Value: kbab.endScale},
	}
	if err := builder.AddScaleAnimation(scaleKeyframes, "linear"); err != nil {
		return nil, fmt.Errorf("failed to add scale animation: %v", err)
	}
	
	return builder.Build()
}

// AnimationPreset represents a preset animation configuration
type AnimationPreset struct {
	Name        string
	Description string
	Builder     func(startTime Time, duration Duration) (*AdjustTransform, error)
}

// GetKenBurnsPresets returns common Ken Burns effect presets
func GetKenBurnsPresets() map[string]AnimationPreset {
	return map[string]AnimationPreset{
		"subtle_zoom_in": {
			Name:        "Subtle Zoom In",
			Description: "Gentle zoom in with slight pan",
			Builder: func(startTime Time, duration Duration) (*AdjustTransform, error) {
				kb := NewKenBurnsAnimationBuilder(startTime, duration)
				kb.SetScaleRange("1.0 1.0", "1.1 1.1")
				kb.SetPositionRange("0 0", "-5 -3")
				return kb.Build()
			},
		},
		"dramatic_zoom_in": {
			Name:        "Dramatic Zoom In",
			Description: "Strong zoom in with noticeable pan",
			Builder: func(startTime Time, duration Duration) (*AdjustTransform, error) {
				kb := NewKenBurnsAnimationBuilder(startTime, duration)
				kb.SetScaleRange("1.0 1.0", "1.5 1.5")
				kb.SetPositionRange("0 0", "-30 -20")
				return kb.Build()
			},
		},
		"zoom_out": {
			Name:        "Zoom Out",
			Description: "Zoom out effect with reverse pan",
			Builder: func(startTime Time, duration Duration) (*AdjustTransform, error) {
				kb := NewKenBurnsAnimationBuilder(startTime, duration)
				kb.SetScaleRange("1.3 1.3", "1.0 1.0")
				kb.SetPositionRange("-20 -15", "0 0")
				return kb.Build()
			},
		},
		"left_to_right": {
			Name:        "Left to Right Pan",
			Description: "Horizontal pan with slight zoom",
			Builder: func(startTime Time, duration Duration) (*AdjustTransform, error) {
				kb := NewKenBurnsAnimationBuilder(startTime, duration)
				kb.SetScaleRange("1.1 1.1", "1.2 1.2")
				kb.SetPositionRange("-30 0", "30 0")
				return kb.Build()
			},
		},
	}
}