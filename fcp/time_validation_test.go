package fcp

import (
	"testing"
)

func TestFrameAccurateTime_NewFromSeconds(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected int // expected frame count
	}{
		{
			name:     "zero seconds",
			seconds:  0.0,
			expected: 0,
		},
		{
			name:     "one second",
			seconds:  1.0,
			expected: 24, // approximately 24 frames at 23.976 fps
		},
		{
			name:     "ten seconds",
			seconds:  10.0,
			expected: 240, // approximately 240 frames
		},
		{
			name:     "fractional second",
			seconds:  0.5,
			expected: 12, // approximately 12 frames
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fat := NewFrameAccurateTimeFromSeconds(tt.seconds)
			if fat.GetFrames() != tt.expected {
				t.Errorf("expected %d frames, got %d", tt.expected, fat.GetFrames())
			}
		})
	}
}

func TestFrameAccurateTime_NewFromFCPString(t *testing.T) {
	tests := []struct {
		name        string
		fcpString   string
		expectError bool
		expected    int // expected frame count
	}{
		{
			name:      "zero time",
			fcpString: "0s",
			expected:  0,
		},
		{
			name:      "frame-aligned time",
			fcpString: "240240/24000s", // 10 seconds at proper frame alignment
			expected:  240,
		},
		{
			name:        "non-frame-aligned time",
			fcpString:   "100000/24000s", // Not multiple of 1001
			expectError: true,
		},
		{
			name:        "wrong timebase",
			fcpString:   "1001/30000s",
			expectError: true,
		},
		{
			name:        "invalid format",
			fcpString:   "invalid",
			expectError: true,
		},
		{
			name:        "decimal format",
			fcpString:   "10.0s",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fat, err := NewFrameAccurateTimeFromFCPString(tt.fcpString)
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
			if fat.GetFrames() != tt.expected {
				t.Errorf("expected %d frames, got %d", tt.expected, fat.GetFrames())
			}
		})
	}
}

func TestFrameAccurateTime_ToFCPString(t *testing.T) {
	tests := []struct {
		name     string
		frames   int
		expected string
	}{
		{
			name:     "zero frames",
			frames:   0,
			expected: "0s",
		},
		{
			name:     "240 frames (10 seconds)",
			frames:   240,
			expected: "240240/24000s",
		},
		{
			name:     "24 frames (1 second)",
			frames:   24,
			expected: "24024/24000s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fat := &FrameAccurateTime{frames: tt.frames}
			result := fat.ToFCPString()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFrameAccurateTime_Arithmetic(t *testing.T) {
	// Create test times
	time1 := &FrameAccurateTime{frames: 100}
	time2 := &FrameAccurateTime{frames: 50}

	// Test addition
	sum := time1.Add(time2)
	if sum.GetFrames() != 150 {
		t.Errorf("addition: expected 150 frames, got %d", sum.GetFrames())
	}

	// Test subtraction
	diff := time1.Subtract(time2)
	if diff.GetFrames() != 50 {
		t.Errorf("subtraction: expected 50 frames, got %d", diff.GetFrames())
	}

	// Test subtraction with clamping
	diff2 := time2.Subtract(time1)
	if diff2.GetFrames() != 0 {
		t.Errorf("subtraction with clamping: expected 0 frames, got %d", diff2.GetFrames())
	}

	// Test comparison
	if time1.Compare(time2) != 1 {
		t.Errorf("comparison: expected 1, got %d", time1.Compare(time2))
	}

	if time2.Compare(time1) != -1 {
		t.Errorf("comparison: expected -1, got %d", time2.Compare(time1))
	}

	if time1.Compare(time1) != 0 {
		t.Errorf("comparison: expected 0, got %d", time1.Compare(time1))
	}
}

func TestFrameAccurateTimeValidator_ValidateTimeString(t *testing.T) {
	strictValidator := NewFrameAccurateTimeValidator(true)
	lenientValidator := NewFrameAccurateTimeValidator(false)

	tests := []struct {
		name          string
		timeString    string
		strictError   bool
		lenientError  bool
	}{
		{
			name:       "valid frame-aligned time",
			timeString: "240240/24000s",
		},
		{
			name:        "non-frame-aligned time",
			timeString:  "100000/24000s",
			strictError: true,
		},
		{
			name:       "decimal time",
			timeString: "10.0s",
			strictError: true,
		},
		{
			name:         "invalid format",
			timeString:   "invalid",
			strictError:  true,
			lenientError: true,
		},
		{
			name:       "zero time",
			timeString: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_strict", func(t *testing.T) {
			err := strictValidator.ValidateTimeString(tt.timeString)
			if tt.strictError && err == nil {
				t.Errorf("expected error in strict mode but got none")
			}
			if !tt.strictError && err != nil {
				t.Errorf("unexpected error in strict mode: %v", err)
			}
		})

		t.Run(tt.name+"_lenient", func(t *testing.T) {
			err := lenientValidator.ValidateTimeString(tt.timeString)
			if tt.lenientError && err == nil {
				t.Errorf("expected error in lenient mode but got none")
			}
			if !tt.lenientError && err != nil {
				t.Errorf("unexpected error in lenient mode: %v", err)
			}
		})
	}
}

func TestFrameAccurateTimeValidator_ConvertToFrameAccurate(t *testing.T) {
	validator := NewFrameAccurateTimeValidator(false)

	tests := []struct {
		name        string
		input       string
		expectError bool
		checkOutput bool
		expected    string
	}{
		{
			name:        "already frame-accurate",
			input:       "240240/24000s",
			checkOutput: true,
			expected:    "240240/24000s",
		},
		{
			name:     "decimal format",
			input:    "10.0s",
			checkOutput: false, // Don't check exact output due to rounding
		},
		{
			name:        "zero time",
			input:       "0s",
			checkOutput: true,
			expected:    "0s",
		},
		{
			name:        "invalid format",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ConvertToFrameAccurate(tt.input)
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
			
			if tt.checkOutput && result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
			
			// Verify result is frame-accurate
			_, parseErr := NewFrameAccurateTimeFromFCPString(result)
			if parseErr != nil {
				t.Errorf("converted result is not frame-accurate: %v", parseErr)
			}
		})
	}
}

func TestFrameAccurateTimeValidator_ValidateSequenceOfTimes(t *testing.T) {
	validator := NewFrameAccurateTimeValidator(false)

	tests := []struct {
		name        string
		times       []string
		expectError bool
	}{
		{
			name:  "ascending times",
			times: []string{"0s", "240240/24000s", "480480/24000s"},
		},
		{
			name:        "non-ascending times",
			times:       []string{"240240/24000s", "0s", "480480/24000s"},
			expectError: true,
		},
		{
			name:  "single time",
			times: []string{"240240/24000s"},
		},
		{
			name:  "empty sequence",
			times: []string{},
		},
		{
			name:        "duplicate times",
			times:       []string{"240240/24000s", "240240/24000s"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSequenceOfTimes(tt.times)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTimeArithmetic_AddTimes(t *testing.T) {
	arithmetic := NewTimeArithmetic(false)

	tests := []struct {
		name        string
		time1       string
		time2       string
		expectError bool
		checkResult bool
		expected    string
	}{
		{
			name:        "add frame-accurate times",
			time1:       "240240/24000s",
			time2:       "240240/24000s",
			checkResult: true,
			expected:    "480480/24000s",
		},
		{
			name:  "add zero to time",
			time1: "240240/24000s",
			time2: "0s",
			checkResult: true,
			expected: "240240/24000s",
		},
		{
			name:        "invalid time1",
			time1:       "invalid",
			time2:       "240240/24000s",
			expectError: true,
		},
		{
			name:        "invalid time2",
			time1:       "240240/24000s",
			time2:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := arithmetic.AddTimes(tt.time1, tt.time2)
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

			if tt.checkResult && result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}

			// Verify result is frame-accurate
			_, parseErr := NewFrameAccurateTimeFromFCPString(result)
			if parseErr != nil {
				t.Errorf("result is not frame-accurate: %v", parseErr)
			}
		})
	}
}

func TestTimeArithmetic_SubtractTimes(t *testing.T) {
	arithmetic := NewTimeArithmetic(false)

	tests := []struct {
		name        string
		time1       string
		time2       string
		expectError bool
		checkResult bool
		expected    string
	}{
		{
			name:        "subtract equal times",
			time1:       "240240/24000s",
			time2:       "240240/24000s",
			checkResult: true,
			expected:    "0s",
		},
		{
			name:        "subtract smaller from larger",
			time1:       "480480/24000s",
			time2:       "240240/24000s",
			checkResult: true,
			expected:    "240240/24000s",
		},
		{
			name:        "subtract larger from smaller (clamp to zero)",
			time1:       "240240/24000s",
			time2:       "480480/24000s",
			checkResult: true,
			expected:    "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := arithmetic.SubtractTimes(tt.time1, tt.time2)
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

			if tt.checkResult && result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}

			// Verify result is frame-accurate
			_, parseErr := NewFrameAccurateTimeFromFCPString(result)
			if parseErr != nil {
				t.Errorf("result is not frame-accurate: %v", parseErr)
			}
		})
	}
}

func TestTimeArithmetic_CompareTimes(t *testing.T) {
	arithmetic := NewTimeArithmetic(false)

	tests := []struct {
		name     string
		time1    string
		time2    string
		expected int
	}{
		{
			name:     "equal times",
			time1:    "240240/24000s",
			time2:    "240240/24000s",
			expected: 0,
		},
		{
			name:     "time1 greater",
			time1:    "480480/24000s",
			time2:    "240240/24000s",
			expected: 1,
		},
		{
			name:     "time1 smaller",
			time1:    "240240/24000s",
			time2:    "480480/24000s",
			expected: -1,
		},
		{
			name:     "zero and positive",
			time1:    "0s",
			time2:    "240240/24000s",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := arithmetic.CompareTimes(tt.time1, tt.time2)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestTimeArithmetic_ValidateTimeRange(t *testing.T) {
	arithmetic := NewTimeArithmetic(false)

	tests := []struct {
		name            string
		elementStart    string
		elementDuration string
		rangeStart      string
		rangeDuration   string
		expectError     bool
	}{
		{
			name:            "element fits in range",
			elementStart:    "240240/24000s",
			elementDuration: "240240/24000s",
			rangeStart:      "0s",
			rangeDuration:   "720720/24000s",
		},
		{
			name:            "element starts before range",
			elementStart:    "0s",
			elementDuration: "240240/24000s",
			rangeStart:      "120120/24000s",
			rangeDuration:   "480480/24000s",
			expectError:     true,
		},
		{
			name:            "element extends beyond range",
			elementStart:    "480480/24000s",
			elementDuration: "480480/24000s",
			rangeStart:      "0s",
			rangeDuration:   "720720/24000s",
			expectError:     true,
		},
		{
			name:            "element exactly fits range",
			elementStart:    "0s",
			elementDuration: "720720/24000s",
			rangeStart:      "0s",
			rangeDuration:   "720720/24000s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := arithmetic.ValidateTimeRange(tt.elementStart, tt.elementDuration, tt.rangeStart, tt.rangeDuration)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTimeArithmetic_GetFrameCount(t *testing.T) {
	arithmetic := NewTimeArithmetic(false)

	tests := []struct {
		name     string
		duration string
		expected int
	}{
		{
			name:     "zero duration",
			duration: "0s",
			expected: 0,
		},
		{
			name:     "240 frames",
			duration: "240240/24000s",
			expected: 240,
		},
		{
			name:     "24 frames",
			duration: "24024/24000s",
			expected: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := arithmetic.GetFrameCount(tt.duration)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %d frames, got %d", tt.expected, result)
			}
		})
	}
}