package fcp

import (
	"testing"
)

func TestAnimationBuilder_PositionAnimation(t *testing.T) {
	builder := NewAnimationBuilder("position")
	
	// Add position keyframes (no interp/curve allowed)
	keyframes := []KeyframeData{
		{Time: Time("0s"), Value: "0 0"},
		{Time: Time("120120/24000s"), Value: "100 50"}, // 5 seconds
	}
	
	err := builder.AddLinearKeyframes(keyframes)
	if err != nil {
		t.Fatalf("Failed to add position keyframes: %v", err)
	}
	
	param, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build position animation: %v", err)
	}
	
	if param.Name != "position" {
		t.Errorf("Expected param name 'position', got '%s'", param.Name)
	}
	
	if len(param.KeyframeAnimation.Keyframes) != 2 {
		t.Errorf("Expected 2 keyframes, got %d", len(param.KeyframeAnimation.Keyframes))
	}
	
	// Position keyframes should not have interp or curve attributes
	for i, kf := range param.KeyframeAnimation.Keyframes {
		if kf.Interp != "" {
			t.Errorf("Position keyframe %d should not have interp attribute, got '%s'", i, kf.Interp)
		}
		if kf.Curve != "" {
			t.Errorf("Position keyframe %d should not have curve attribute, got '%s'", i, kf.Curve)
		}
	}
}

func TestAnimationBuilder_ScaleAnimation(t *testing.T) {
	builder := NewAnimationBuilder("scale")
	
	keyframes := []KeyframeData{
		{Time: Time("0s"), Value: "1.0 1.0"},
		{Time: Time("120120/24000s"), Value: "1.5 1.5"},
	}
	
	err := builder.AddLinearKeyframes(keyframes)
	if err != nil {
		t.Fatalf("Failed to add scale keyframes: %v", err)
	}
	
	param, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build scale animation: %v", err)
	}
	
	if param.Name != "scale" {
		t.Errorf("Expected param name 'scale', got '%s'", param.Name)
	}
	
	// Scale keyframes should have curve but not interp
	for i, kf := range param.KeyframeAnimation.Keyframes {
		if kf.Interp != "" {
			t.Errorf("Scale keyframe %d should not have interp attribute, got '%s'", i, kf.Interp)
		}
		if kf.Curve != "linear" {
			t.Errorf("Scale keyframe %d should have curve='linear', got '%s'", i, kf.Curve)
		}
	}
}

func TestAnimationBuilder_OpacityAnimation(t *testing.T) {
	builder := NewAnimationBuilder("opacity")
	
	keyframes := []KeyframeData{
		{Time: Time("0s"), Value: "1.0"},
		{Time: Time("120120/24000s"), Value: "0.5"},
	}
	
	err := builder.AddLinearKeyframes(keyframes)
	if err != nil {
		t.Fatalf("Failed to add opacity keyframes: %v", err)
	}
	
	param, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build opacity animation: %v", err)
	}
	
	if param.Name != "opacity" {
		t.Errorf("Expected param name 'opacity', got '%s'", param.Name)
	}
	
	// Opacity keyframes should have both interp and curve
	for i, kf := range param.KeyframeAnimation.Keyframes {
		if kf.Interp != "linear" {
			t.Errorf("Opacity keyframe %d should have interp='linear', got '%s'", i, kf.Interp)
		}
		if kf.Curve != "linear" {
			t.Errorf("Opacity keyframe %d should have curve='linear', got '%s'", i, kf.Curve)
		}
	}
}

func TestAnimationBuilder_InvalidKeyframeAttributes(t *testing.T) {
	// Test that position keyframes reject interp/curve attributes
	builder := NewAnimationBuilder("position")
	
	err := builder.AddKeyframe(Time("0s"), "0 0", WithInterp("linear"))
	if err == nil {
		t.Error("Position keyframes should reject interp attribute")
	}
	
	err = builder.AddKeyframe(Time("0s"), "0 0", WithCurve("smooth"))
	if err == nil {
		t.Error("Position keyframes should reject curve attribute")
	}
}

func TestAnimationBuilder_KeyframeOrdering(t *testing.T) {
	builder := NewAnimationBuilder("opacity")
	
	// Add keyframes out of chronological order
	err := builder.AddKeyframe(Time("120120/24000s"), "0.5", WithInterp("linear"))
	if err != nil {
		t.Fatalf("Failed to add first keyframe: %v", err)
	}
	
	err = builder.AddKeyframe(Time("0s"), "1.0", WithInterp("linear"))
	if err != nil {
		t.Fatalf("Failed to add second keyframe: %v", err)
	}
	
	param, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build animation: %v", err)
	}
	
	// Check that keyframes are sorted by time
	if param.KeyframeAnimation.Keyframes[0].Time != "0s" {
		t.Errorf("First keyframe should be at 0s, got %s", param.KeyframeAnimation.Keyframes[0].Time)
	}
	
	if param.KeyframeAnimation.Keyframes[1].Time != "120120/24000s" {
		t.Errorf("Second keyframe should be at 120120/24000s, got %s", param.KeyframeAnimation.Keyframes[1].Time)
	}
}

func TestTransformBuilder_PositionAndScale(t *testing.T) {
	builder := NewTransformBuilder()
	
	// Add position animation
	positionKeyframes := []KeyframeData{
		{Time: Time("0s"), Value: "0 0"},
		{Time: Time("120120/24000s"), Value: "100 50"},
	}
	err := builder.AddPositionAnimation(positionKeyframes)
	if err != nil {
		t.Fatalf("Failed to add position animation: %v", err)
	}
	
	// Add scale animation
	scaleKeyframes := []KeyframeData{
		{Time: Time("0s"), Value: "1.0 1.0"},
		{Time: Time("120120/24000s"), Value: "1.5 1.5"},
	}
	err = builder.AddScaleAnimation(scaleKeyframes, "linear")
	if err != nil {
		t.Fatalf("Failed to add scale animation: %v", err)
	}
	
	transform, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build transform: %v", err)
	}
	
	if len(transform.Params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(transform.Params))
	}
	
	// Check parameter names
	paramNames := make(map[string]bool)
	for _, param := range transform.Params {
		paramNames[param.Name] = true
	}
	
	if !paramNames["position"] {
		t.Error("Missing position parameter")
	}
	if !paramNames["scale"] {
		t.Error("Missing scale parameter")
	}
}

func TestTransformBuilder_StaticValues(t *testing.T) {
	builder := NewTransformBuilder()
	
	// Set static values
	err := builder.SetStaticPosition("50 25")
	if err != nil {
		t.Fatalf("Failed to set static position: %v", err)
	}
	
	err = builder.SetStaticScale("1.2 1.2")
	if err != nil {
		t.Fatalf("Failed to set static scale: %v", err)
	}
	
	transform, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build transform: %v", err)
	}
	
	if transform.Position != "50 25" {
		t.Errorf("Expected position '50 25', got '%s'", transform.Position)
	}
	
	if transform.Scale != "1.2 1.2" {
		t.Errorf("Expected scale '1.2 1.2', got '%s'", transform.Scale)
	}
	
	// Should have no animated parameters
	if len(transform.Params) != 0 {
		t.Errorf("Expected 0 animated parameters, got %d", len(transform.Params))
	}
}

func TestTransformBuilder_InvalidStaticValues(t *testing.T) {
	builder := NewTransformBuilder()
	
	// Test invalid position format
	err := builder.SetStaticPosition("invalid")
	if err == nil {
		t.Error("Should reject invalid position format")
	}
	
	// Test invalid scale format
	err = builder.SetStaticScale("not a scale")
	if err == nil {
		t.Error("Should reject invalid scale format")
	}
	
	// Test negative scale
	err = builder.SetStaticScale("-1.0 1.0")
	if err == nil {
		t.Error("Should reject negative scale values")
	}
}

func TestKenBurnsAnimationBuilder(t *testing.T) {
	startTime := Time("60060/24000s") // 2.5 seconds
	duration := Duration("72072/24000s") // 3 seconds
	
	builder := NewKenBurnsAnimationBuilder(startTime, duration)
	
	transform, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build Ken Burns animation: %v", err)
	}
	
	if len(transform.Params) != 2 {
		t.Errorf("Expected 2 parameters (position, scale), got %d", len(transform.Params))
	}
	
	// Check parameter names
	paramNames := make(map[string]bool)
	for _, param := range transform.Params {
		paramNames[param.Name] = true
	}
	
	if !paramNames["position"] {
		t.Error("Missing position parameter")
	}
	if !paramNames["scale"] {
		t.Error("Missing scale parameter")
	}
	
	// Verify keyframe count for each parameter
	for _, param := range transform.Params {
		if param.KeyframeAnimation == nil {
			t.Errorf("Parameter %s should have keyframe animation", param.Name)
			continue
		}
		
		if len(param.KeyframeAnimation.Keyframes) != 2 {
			t.Errorf("Parameter %s should have 2 keyframes, got %d", 
				param.Name, len(param.KeyframeAnimation.Keyframes))
		}
	}
}

func TestKenBurnsAnimationBuilder_CustomRanges(t *testing.T) {
	startTime := Time("0s")
	duration := Duration("120120/24000s")
	
	builder := NewKenBurnsAnimationBuilder(startTime, duration)
	
	// Set custom ranges
	err := builder.SetPositionRange("-10 -5", "10 5")
	if err != nil {
		t.Fatalf("Failed to set position range: %v", err)
	}
	
	err = builder.SetScaleRange("0.9 0.9", "1.1 1.1")
	if err != nil {
		t.Fatalf("Failed to set scale range: %v", err)
	}
	
	transform, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build Ken Burns animation: %v", err)
	}
	
	// Find position parameter
	var positionParam *Param
	for i := range transform.Params {
		if transform.Params[i].Name == "position" {
			positionParam = &transform.Params[i]
			break
		}
	}
	
	if positionParam == nil {
		t.Fatal("Position parameter not found")
	}
	
	// Check position keyframe values
	keyframes := positionParam.KeyframeAnimation.Keyframes
	if keyframes[0].Value != "-10 -5" {
		t.Errorf("Expected start position '-10 -5', got '%s'", keyframes[0].Value)
	}
	if keyframes[1].Value != "10 5" {
		t.Errorf("Expected end position '10 5', got '%s'", keyframes[1].Value)
	}
}

func TestOpacityAnimationBuilder(t *testing.T) {
	builder := NewOpacityAnimationBuilder()
	
	keyframes := []KeyframeData{
		{Time: Time("0s"), Value: "1.0"},
		{Time: Time("60060/24000s"), Value: "0.0"},
		{Time: Time("120120/24000s"), Value: "1.0"},
	}
	
	err := builder.AddKeyframes(keyframes, "easeInOut")
	if err != nil {
		t.Fatalf("Failed to add opacity keyframes: %v", err)
	}
	
	param, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build opacity animation: %v", err)
	}
	
	if param.Name != "opacity" {
		t.Errorf("Expected param name 'opacity', got '%s'", param.Name)
	}
	
	if len(param.KeyframeAnimation.Keyframes) != 3 {
		t.Errorf("Expected 3 keyframes, got %d", len(param.KeyframeAnimation.Keyframes))
	}
	
	// Check interpolation settings
	for i, kf := range param.KeyframeAnimation.Keyframes {
		if kf.Interp != "easeInOut" {
			t.Errorf("Keyframe %d should have interp='easeInOut', got '%s'", i, kf.Interp)
		}
		if kf.Curve != "smooth" {
			t.Errorf("Keyframe %d should have curve='smooth', got '%s'", i, kf.Curve)
		}
	}
}

func TestAnimationPresets(t *testing.T) {
	presets := GetKenBurnsPresets()
	
	expectedPresets := []string{"subtle_zoom_in", "dramatic_zoom_in", "zoom_out", "left_to_right"}
	
	for _, expectedName := range expectedPresets {
		preset, exists := presets[expectedName]
		if !exists {
			t.Errorf("Missing expected preset: %s", expectedName)
			continue
		}
		
		if preset.Name == "" {
			t.Errorf("Preset %s has empty name", expectedName)
		}
		
		if preset.Description == "" {
			t.Errorf("Preset %s has empty description", expectedName)
		}
		
		if preset.Builder == nil {
			t.Errorf("Preset %s has nil builder", expectedName)
			continue
		}
		
		// Test that the preset builder works
		startTime := Time("0s")
		duration := Duration("120120/24000s")
		
		transform, err := preset.Builder(startTime, duration)
		if err != nil {
			t.Errorf("Preset %s builder failed: %v", expectedName, err)
			continue
		}
		
		if transform == nil {
			t.Errorf("Preset %s builder returned nil transform", expectedName)
			continue
		}
		
		if len(transform.Params) == 0 {
			t.Errorf("Preset %s produced no animated parameters", expectedName)
		}
	}
}

func TestAnimationBuilder_EmptyKeyframes(t *testing.T) {
	builder := NewAnimationBuilder("position")
	
	_, err := builder.Build()
	if err == nil {
		t.Error("Should reject empty keyframe sequence")
	}
}

func TestTransformBuilder_AnimationOverridesStatic(t *testing.T) {
	builder := NewTransformBuilder()
	
	// Set static position
	err := builder.SetStaticPosition("10 10")
	if err != nil {
		t.Fatalf("Failed to set static position: %v", err)
	}
	
	// Add position animation (should override static)
	positionKeyframes := []KeyframeData{
		{Time: Time("0s"), Value: "0 0"},
		{Time: Time("120120/24000s"), Value: "100 50"},
	}
	err = builder.AddPositionAnimation(positionKeyframes)
	if err != nil {
		t.Fatalf("Failed to add position animation: %v", err)
	}
	
	transform, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build transform: %v", err)
	}
	
	// Should have animated position, not static
	if transform.Position != "" {
		t.Errorf("Animation should override static position, but got static position: '%s'", transform.Position)
	}
	
	// Should have one parameter (position animation)
	if len(transform.Params) != 1 {
		t.Errorf("Expected 1 animated parameter, got %d", len(transform.Params))
	}
	
	if transform.Params[0].Name != "position" {
		t.Errorf("Expected position parameter, got '%s'", transform.Params[0].Name)
	}
}