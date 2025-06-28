package fcp

import (
	"testing"
)

func TestParseKeyframeParameterType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected KeyframeParameterType
	}{
		{
			name:     "position parameter",
			input:    "position",
			expected: KeyframeParameterPosition,
		},
		{
			name:     "scale parameter",
			input:    "scale",
			expected: KeyframeParameterScale,
		},
		{
			name:     "rotation parameter",
			input:    "rotation",
			expected: KeyframeParameterRotation,
		},
		{
			name:     "anchor parameter",
			input:    "anchor",
			expected: KeyframeParameterAnchor,
		},
		{
			name:     "opacity parameter",
			input:    "opacity",
			expected: KeyframeParameterOpacity,
		},
		{
			name:     "volume parameter",
			input:    "volume",
			expected: KeyframeParameterVolume,
		},
		{
			name:     "unknown parameter",
			input:    "unknown",
			expected: KeyframeParameterUnknown,
		},
		{
			name:     "case insensitive",
			input:    "POSITION",
			expected: KeyframeParameterPosition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseKeyframeParameterType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestKeyframeValidator_ValidateKeyframe_Position(t *testing.T) {
	validator := NewKeyframeValidator()

	tests := []struct {
		name        string
		keyframe    *ValidatedKeyframe
		expectError bool
	}{
		{
			name: "valid position keyframe",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "100 50",
				Interp: "",
				Curve:  "",
			},
		},
		{
			name: "position with interp (invalid)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "100 50",
				Interp: "linear",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "position with curve (invalid)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "100 50",
				Interp: "",
				Curve:  "linear",
			},
			expectError: true,
		},
		{
			name: "invalid position value format",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "100", // Missing Y component
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "invalid position value (non-numeric)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "abc def",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateKeyframe("position", tt.keyframe)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKeyframeValidator_ValidateKeyframe_Scale(t *testing.T) {
	validator := NewKeyframeValidator()

	tests := []struct {
		name        string
		keyframe    *ValidatedKeyframe
		expectError bool
	}{
		{
			name: "valid scale keyframe",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.5 1.5",
				Interp: "",
				Curve:  "linear",
			},
		},
		{
			name: "scale without curve",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.5 1.5",
				Interp: "",
				Curve:  "",
			},
		},
		{
			name: "scale with interp (invalid)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.5 1.5",
				Interp: "linear",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "invalid curve value",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.5 1.5",
				Interp: "",
				Curve:  "invalid",
			},
			expectError: true,
		},
		{
			name: "negative scale value",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "-1.0 1.0",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateKeyframe("scale", tt.keyframe)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKeyframeValidator_ValidateKeyframe_Opacity(t *testing.T) {
	validator := NewKeyframeValidator()

	tests := []struct {
		name        string
		keyframe    *ValidatedKeyframe
		expectError bool
	}{
		{
			name: "valid opacity keyframe with both attributes",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "0.5",
				Interp: "linear",
				Curve:  "smooth",
			},
		},
		{
			name: "opacity with only interp",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "0.8",
				Interp: "easeIn",
				Curve:  "",
			},
		},
		{
			name: "opacity with only curve",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "0.2",
				Interp: "",
				Curve:  "linear",
			},
		},
		{
			name: "invalid interp value",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "0.5",
				Interp: "invalid",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "opacity value out of range (too high)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.5",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "opacity value out of range (negative)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "-0.1",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateKeyframe("opacity", tt.keyframe)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKeyframeValidator_ValidateKeyframe_Volume(t *testing.T) {
	validator := NewKeyframeValidator()

	tests := []struct {
		name        string
		keyframe    *ValidatedKeyframe
		expectError bool
	}{
		{
			name: "valid dB volume",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "0dB",
				Interp: "",
				Curve:  "",
			},
		},
		{
			name: "valid negative dB volume",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "-6dB",
				Interp: "",
				Curve:  "",
			},
		},
		{
			name: "valid linear multiplier",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "0.5",
				Interp: "",
				Curve:  "",
			},
		},
		{
			name: "dB value out of range",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "-100dB",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "negative linear multiplier",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "-0.5",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateKeyframe("volume", tt.keyframe)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKeyframeValidator_ValidateKeyframe_Color(t *testing.T) {
	validator := NewKeyframeValidator()

	tests := []struct {
		name        string
		keyframe    *ValidatedKeyframe
		expectError bool
	}{
		{
			name: "valid RGB color",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.0 0.5 0.0",
				Interp: "",
				Curve:  "",
			},
		},
		{
			name: "valid RGBA color",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.0 0.5 0.0 0.8",
				Interp: "",
				Curve:  "",
			},
		},
		{
			name: "color component out of range",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.5 0.5 0.0",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "invalid color format (too few components)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.0 0.5",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
		{
			name: "invalid color format (too many components)",
			keyframe: &ValidatedKeyframe{
				Time:  Time("0s"),
				Value: "1.0 0.5 0.0 0.8 0.2",
				Interp: "",
				Curve:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateKeyframe("color", tt.keyframe)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKeyframeValidator_ValidateKeyframeSequence(t *testing.T) {
	validator := NewKeyframeValidator()

	// Create valid keyframes in chronological order
	keyframe1, _ := NewValidatedKeyframe(Time("0s"), "0 0", "", "")
	keyframe2, _ := NewValidatedKeyframe(Time("240240/24000s"), "100 50", "", "")
	keyframe3, _ := NewValidatedKeyframe(Time("480480/24000s"), "200 100", "", "")

	// Create keyframes in wrong order
	keyframe4, _ := NewValidatedKeyframe(Time("240240/24000s"), "100 50", "", "")
	keyframe5, _ := NewValidatedKeyframe(Time("0s"), "0 0", "", "")

	tests := []struct {
		name        string
		paramName   string
		keyframes   []*ValidatedKeyframe
		expectError bool
	}{
		{
			name:      "valid chronological sequence",
			paramName: "position",
			keyframes: []*ValidatedKeyframe{keyframe1, keyframe2, keyframe3},
		},
		{
			name:        "non-chronological sequence",
			paramName:   "position",
			keyframes:   []*ValidatedKeyframe{keyframe4, keyframe5},
			expectError: true,
		},
		{
			name:        "empty sequence",
			paramName:   "position",
			keyframes:   []*ValidatedKeyframe{},
			expectError: true,
		},
		{
			name:      "single keyframe",
			paramName: "position",
			keyframes: []*ValidatedKeyframe{keyframe1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateKeyframeSequence(tt.paramName, tt.keyframes)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestKeyframeBuilder(t *testing.T) {
	builder := NewKeyframeBuilder("position")

	// Add valid position keyframes (no interp/curve allowed)
	err := builder.AddKeyframe(Time("0s"), "0 0")
	if err != nil {
		t.Errorf("unexpected error adding first keyframe: %v", err)
	}

	err = builder.AddKeyframe(Time("240240/24000s"), "100 50")
	if err != nil {
		t.Errorf("unexpected error adding second keyframe: %v", err)
	}

	// Try to add invalid keyframe with curve (should fail for position)
	err = builder.AddKeyframe(Time("480480/24000s"), "200 100", WithCurve("linear"))
	if err == nil {
		t.Errorf("expected error adding position keyframe with curve")
	}

	// Build the sequence
	keyframes, err := builder.Build()
	if err != nil {
		t.Errorf("unexpected error building keyframes: %v", err)
	}

	if len(keyframes) != 2 {
		t.Errorf("expected 2 keyframes, got %d", len(keyframes))
	}

	// Verify keyframes are in order
	if keyframes[0].Time != Time("0s") {
		t.Errorf("first keyframe time incorrect: %s", keyframes[0].Time)
	}

	if keyframes[1].Time != Time("240240/24000s") {
		t.Errorf("second keyframe time incorrect: %s", keyframes[1].Time)
	}
}

func TestKeyframeBuilder_ScaleParameters(t *testing.T) {
	builder := NewKeyframeBuilder("scale")

	// Add valid scale keyframes (curve allowed, interp not allowed)
	err := builder.AddKeyframe(Time("0s"), "1.0 1.0", WithCurve("linear"))
	if err != nil {
		t.Errorf("unexpected error adding scale keyframe with curve: %v", err)
	}

	err = builder.AddKeyframe(Time("240240/24000s"), "1.5 1.5", WithCurve("smooth"))
	if err != nil {
		t.Errorf("unexpected error adding second scale keyframe: %v", err)
	}

	// Try to add invalid keyframe with interp (should fail for scale)
	err = builder.AddKeyframe(Time("480480/24000s"), "2.0 2.0", WithInterp("linear"))
	if err == nil {
		t.Errorf("expected error adding scale keyframe with interp")
	}

	// Build the sequence
	keyframes, err := builder.Build()
	if err != nil {
		t.Errorf("unexpected error building scale keyframes: %v", err)
	}

	if len(keyframes) != 2 {
		t.Errorf("expected 2 keyframes, got %d", len(keyframes))
	}

	// Verify curve attributes are preserved
	if keyframes[0].Curve != "linear" {
		t.Errorf("first keyframe curve incorrect: %s", keyframes[0].Curve)
	}

	if keyframes[1].Curve != "smooth" {
		t.Errorf("second keyframe curve incorrect: %s", keyframes[1].Curve)
	}
}

func TestKeyframeBuilder_OpacityParameters(t *testing.T) {
	builder := NewKeyframeBuilder("opacity")

	// Add valid opacity keyframes (both curve and interp allowed)
	err := builder.AddKeyframe(Time("0s"), "1.0", WithInterp("linear"), WithCurve("smooth"))
	if err != nil {
		t.Errorf("unexpected error adding opacity keyframe: %v", err)
	}

	err = builder.AddKeyframe(Time("240240/24000s"), "0.5", WithInterp("easeOut"))
	if err != nil {
		t.Errorf("unexpected error adding second opacity keyframe: %v", err)
	}

	// Build the sequence
	keyframes, err := builder.Build()
	if err != nil {
		t.Errorf("unexpected error building opacity keyframes: %v", err)
	}

	if len(keyframes) != 2 {
		t.Errorf("expected 2 keyframes, got %d", len(keyframes))
	}

	// Verify both interp and curve attributes are preserved
	if keyframes[0].Interp != "linear" {
		t.Errorf("first keyframe interp incorrect: %s", keyframes[0].Interp)
	}

	if keyframes[0].Curve != "smooth" {
		t.Errorf("first keyframe curve incorrect: %s", keyframes[0].Curve)
	}

	if keyframes[1].Interp != "easeOut" {
		t.Errorf("second keyframe interp incorrect: %s", keyframes[1].Interp)
	}
}

func TestKeyframeSequenceValidator(t *testing.T) {
	validator := NewKeyframeSequenceValidator()

	// Create test animations
	positionKeyframes := []*ValidatedKeyframe{
		{Time: Time("0s"), Value: "0 0"},
		{Time: Time("240240/24000s"), Value: "100 50"},
	}

	scaleKeyframes := []*ValidatedKeyframe{
		{Time: Time("0s"), Value: "1.0 1.0", Curve: "linear"},
		{Time: Time("240240/24000s"), Value: "1.5 1.5", Curve: "smooth"},
	}

	animations := map[string][]*ValidatedKeyframe{
		"position": positionKeyframes,
		"scale":    scaleKeyframes,
	}

	err := validator.ValidateMultipleParameters(animations)
	if err != nil {
		t.Errorf("unexpected error validating multiple parameters: %v", err)
	}
}

func TestNewValidatedKeyframe(t *testing.T) {
	tests := []struct {
		name        string
		time        Time
		value       string
		interp      string
		curve       string
		expectError bool
	}{
		{
			name:  "valid keyframe",
			time:  Time("0s"),
			value: "100 50",
			interp: "",
			curve: "",
		},
		{
			name:        "invalid time",
			time:        Time("invalid"),
			value:       "100 50",
			expectError: true,
		},
		{
			name:        "empty value",
			time:        Time("0s"),
			value:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewValidatedKeyframe(tt.time, tt.value, tt.interp, tt.curve)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}