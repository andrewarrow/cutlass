// Package time_validation implements Step 7 of the FCPXMLKit-inspired refactoring plan:
// Frame-accurate time validation with strict frame alignment checking.
//
// This provides comprehensive time validation that ensures all time values are
// frame-aligned with FCP's 24000/1001 timebase and provides safe time arithmetic.
package fcp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FrameAccurateTime provides frame-accurate time operations
type FrameAccurateTime struct {
	frames int // Internal representation in frames
}

// NewFrameAccurateTimeFromSeconds creates a frame-accurate time from seconds
func NewFrameAccurateTimeFromSeconds(seconds float64) *FrameAccurateTime {
	frames := int(math.Round(seconds * FCPFrameRate))
	return &FrameAccurateTime{frames: frames}
}

// NewFrameAccurateTimeFromFCPString creates a frame-accurate time from FCP string format
func NewFrameAccurateTimeFromFCPString(fcpTime string) (*FrameAccurateTime, error) {
	if fcpTime == "0s" {
		return &FrameAccurateTime{frames: 0}, nil
	}
	
	if !strings.HasSuffix(fcpTime, "s") {
		return nil, fmt.Errorf("time must end with 's': %s", fcpTime)
	}
	
	timeNoS := strings.TrimSuffix(fcpTime, "s")
	
	if !strings.Contains(timeNoS, "/") {
		return nil, fmt.Errorf("time must be in rational format: %s", fcpTime)
	}
	
	parts := strings.Split(timeNoS, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid rational format: %s", fcpTime)
	}
	
	numerator, err1 := strconv.Atoi(parts[0])
	denominator, err2 := strconv.Atoi(parts[1])
	
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("non-integer rational parts: %s", fcpTime)
	}
	
	if denominator != FCPTimebase {
		return nil, fmt.Errorf("wrong timebase, expected %d, got %d", FCPTimebase, denominator)
	}
	
	if numerator%FCPFrameDuration != 0 {
		return nil, fmt.Errorf("time not frame-aligned: %s (numerator must be multiple of %d)", fcpTime, FCPFrameDuration)
	}
	
	frames := numerator / FCPFrameDuration
	return &FrameAccurateTime{frames: frames}, nil
}

// ToSeconds converts to seconds
func (fat *FrameAccurateTime) ToSeconds() float64 {
	return float64(fat.frames) / FCPFrameRate
}

// ToFCPString converts to FCP string format
func (fat *FrameAccurateTime) ToFCPString() string {
	if fat.frames == 0 {
		return "0s"
	}
	
	numerator := fat.frames * FCPFrameDuration
	return fmt.Sprintf("%d/%ds", numerator, FCPTimebase)
}

// GetFrames returns the frame count
func (fat *FrameAccurateTime) GetFrames() int {
	return fat.frames
}

// Add adds another frame-accurate time
func (fat *FrameAccurateTime) Add(other *FrameAccurateTime) *FrameAccurateTime {
	return &FrameAccurateTime{frames: fat.frames + other.frames}
}

// Subtract subtracts another frame-accurate time
func (fat *FrameAccurateTime) Subtract(other *FrameAccurateTime) *FrameAccurateTime {
	result := fat.frames - other.frames
	if result < 0 {
		result = 0 // Clamp to zero
	}
	return &FrameAccurateTime{frames: result}
}

// Compare compares with another frame-accurate time (-1, 0, 1)
func (fat *FrameAccurateTime) Compare(other *FrameAccurateTime) int {
	if fat.frames < other.frames {
		return -1
	} else if fat.frames > other.frames {
		return 1
	}
	return 0
}

// IsZero checks if the time is zero
func (fat *FrameAccurateTime) IsZero() bool {
	return fat.frames == 0
}

// FrameAccurateTimeValidator provides comprehensive time validation
type FrameAccurateTimeValidator struct {
	strictMode bool // Whether to enforce strict frame alignment
}

// NewFrameAccurateTimeValidator creates a new time validator
func NewFrameAccurateTimeValidator(strictMode bool) *FrameAccurateTimeValidator {
	return &FrameAccurateTimeValidator{strictMode: strictMode}
}

// ValidateTimeString validates a time string for frame accuracy
func (fatv *FrameAccurateTimeValidator) ValidateTimeString(timeStr string) error {
	if timeStr == "" {
		return fmt.Errorf("time string cannot be empty")
	}
	
	// Try to parse as frame-accurate time
	_, err := NewFrameAccurateTimeFromFCPString(timeStr)
	if err != nil {
		if fatv.strictMode {
			return fmt.Errorf("strict mode: %v", err)
		}
		// In non-strict mode, try to validate basic format
		return fatv.validateBasicTimeFormat(timeStr)
	}
	
	return nil
}

// validateBasicTimeFormat validates basic time format without strict frame alignment
func (fatv *FrameAccurateTimeValidator) validateBasicTimeFormat(timeStr string) error {
	if timeStr == "0s" {
		return nil
	}
	
	if !strings.HasSuffix(timeStr, "s") {
		return fmt.Errorf("time must end with 's': %s", timeStr)
	}
	
	timeNoS := strings.TrimSuffix(timeStr, "s")
	
	// Allow both decimal and rational formats in non-strict mode
	if strings.Contains(timeNoS, "/") {
		// Rational format
		parts := strings.Split(timeNoS, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid rational format: %s", timeStr)
		}
		
		_, err1 := strconv.ParseFloat(parts[0], 64)
		_, err2 := strconv.ParseFloat(parts[1], 64)
		
		if err1 != nil || err2 != nil {
			return fmt.Errorf("invalid rational parts: %s", timeStr)
		}
	} else {
		// Decimal format
		_, err := strconv.ParseFloat(timeNoS, 64)
		if err != nil {
			return fmt.Errorf("invalid decimal format: %s", timeStr)
		}
	}
	
	return nil
}

// ConvertToFrameAccurate converts any time string to frame-accurate format
func (fatv *FrameAccurateTimeValidator) ConvertToFrameAccurate(timeStr string) (string, error) {
	if timeStr == "0s" {
		return "0s", nil
	}
	
	if !strings.HasSuffix(timeStr, "s") {
		return "", fmt.Errorf("time must end with 's': %s", timeStr)
	}
	
	timeNoS := strings.TrimSuffix(timeStr, "s")
	
	var seconds float64
	var err error
	
	if strings.Contains(timeNoS, "/") {
		// Rational format
		parts := strings.Split(timeNoS, "/")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid rational format: %s", timeStr)
		}
		
		numerator, err1 := strconv.ParseFloat(parts[0], 64)
		denominator, err2 := strconv.ParseFloat(parts[1], 64)
		
		if err1 != nil || err2 != nil || denominator == 0 {
			return "", fmt.Errorf("invalid rational parts: %s", timeStr)
		}
		
		seconds = numerator / denominator
	} else {
		// Decimal format
		seconds, err = strconv.ParseFloat(timeNoS, 64)
		if err != nil {
			return "", fmt.Errorf("invalid decimal format: %s", timeStr)
		}
	}
	
	// Convert to frame-accurate
	frameAccurateTime := NewFrameAccurateTimeFromSeconds(seconds)
	return frameAccurateTime.ToFCPString(), nil
}

// ValidateSequenceOfTimes validates a sequence of times for correct ordering
func (fatv *FrameAccurateTimeValidator) ValidateSequenceOfTimes(times []string) error {
	if len(times) <= 1 {
		return nil // Single time or empty is valid
	}
	
	var previousTime *FrameAccurateTime
	
	for i, timeStr := range times {
		// Validate individual time
		if err := fatv.ValidateTimeString(timeStr); err != nil {
			return fmt.Errorf("time %d validation failed: %v", i, err)
		}
		
		// Parse time
		currentTime, err := NewFrameAccurateTimeFromFCPString(timeStr)
		if err != nil {
			// Try conversion if not frame-accurate
			convertedTimeStr, convErr := fatv.ConvertToFrameAccurate(timeStr)
			if convErr != nil {
				return fmt.Errorf("time %d conversion failed: %v", i, convErr)
			}
			currentTime, err = NewFrameAccurateTimeFromFCPString(convertedTimeStr)
			if err != nil {
				return fmt.Errorf("time %d parsing failed: %v", i, err)
			}
		}
		
		// Check ordering
		if previousTime != nil {
			if currentTime.Compare(previousTime) <= 0 {
				return fmt.Errorf("times must be in ascending order: time %d (%s) is not greater than time %d", 
					i, timeStr, i-1)
			}
		}
		
		previousTime = currentTime
	}
	
	return nil
}

// TimeArithmetic provides safe time arithmetic operations
type TimeArithmetic struct {
	validator *FrameAccurateTimeValidator
}

// NewTimeArithmetic creates a new time arithmetic helper
func NewTimeArithmetic(strictMode bool) *TimeArithmetic {
	return &TimeArithmetic{
		validator: NewFrameAccurateTimeValidator(strictMode),
	}
}

// AddTimes adds two time strings and returns frame-accurate result
func (ta *TimeArithmetic) AddTimes(time1, time2 string) (string, error) {
	// Validate inputs
	if err := ta.validator.ValidateTimeString(time1); err != nil {
		return "", fmt.Errorf("invalid time1: %v", err)
	}
	if err := ta.validator.ValidateTimeString(time2); err != nil {
		return "", fmt.Errorf("invalid time2: %v", err)
	}
	
	// Convert to frame-accurate if needed
	frameTime1Str, err := ta.validator.ConvertToFrameAccurate(time1)
	if err != nil {
		return "", fmt.Errorf("failed to convert time1: %v", err)
	}
	
	frameTime2Str, err := ta.validator.ConvertToFrameAccurate(time2)
	if err != nil {
		return "", fmt.Errorf("failed to convert time2: %v", err)
	}
	
	// Parse as frame-accurate times
	frameTime1, err := NewFrameAccurateTimeFromFCPString(frameTime1Str)
	if err != nil {
		return "", fmt.Errorf("failed to parse time1: %v", err)
	}
	
	frameTime2, err := NewFrameAccurateTimeFromFCPString(frameTime2Str)
	if err != nil {
		return "", fmt.Errorf("failed to parse time2: %v", err)
	}
	
	// Add times
	result := frameTime1.Add(frameTime2)
	return result.ToFCPString(), nil
}

// SubtractTimes subtracts time2 from time1 and returns frame-accurate result
func (ta *TimeArithmetic) SubtractTimes(time1, time2 string) (string, error) {
	// Validate inputs
	if err := ta.validator.ValidateTimeString(time1); err != nil {
		return "", fmt.Errorf("invalid time1: %v", err)
	}
	if err := ta.validator.ValidateTimeString(time2); err != nil {
		return "", fmt.Errorf("invalid time2: %v", err)
	}
	
	// Convert to frame-accurate if needed
	frameTime1Str, err := ta.validator.ConvertToFrameAccurate(time1)
	if err != nil {
		return "", fmt.Errorf("failed to convert time1: %v", err)
	}
	
	frameTime2Str, err := ta.validator.ConvertToFrameAccurate(time2)
	if err != nil {
		return "", fmt.Errorf("failed to convert time2: %v", err)
	}
	
	// Parse as frame-accurate times
	frameTime1, err := NewFrameAccurateTimeFromFCPString(frameTime1Str)
	if err != nil {
		return "", fmt.Errorf("failed to parse time1: %v", err)
	}
	
	frameTime2, err := NewFrameAccurateTimeFromFCPString(frameTime2Str)
	if err != nil {
		return "", fmt.Errorf("failed to parse time2: %v", err)
	}
	
	// Subtract times
	result := frameTime1.Subtract(frameTime2)
	return result.ToFCPString(), nil
}

// CompareTimes compares two time strings (-1 if time1 < time2, 0 if equal, 1 if time1 > time2)
func (ta *TimeArithmetic) CompareTimes(time1, time2 string) (int, error) {
	// Validate inputs
	if err := ta.validator.ValidateTimeString(time1); err != nil {
		return 0, fmt.Errorf("invalid time1: %v", err)
	}
	if err := ta.validator.ValidateTimeString(time2); err != nil {
		return 0, fmt.Errorf("invalid time2: %v", err)
	}
	
	// Convert to frame-accurate if needed
	frameTime1Str, err := ta.validator.ConvertToFrameAccurate(time1)
	if err != nil {
		return 0, fmt.Errorf("failed to convert time1: %v", err)
	}
	
	frameTime2Str, err := ta.validator.ConvertToFrameAccurate(time2)
	if err != nil {
		return 0, fmt.Errorf("failed to convert time2: %v", err)
	}
	
	// Parse as frame-accurate times
	frameTime1, err := NewFrameAccurateTimeFromFCPString(frameTime1Str)
	if err != nil {
		return 0, fmt.Errorf("failed to parse time1: %v", err)
	}
	
	frameTime2, err := NewFrameAccurateTimeFromFCPString(frameTime2Str)
	if err != nil {
		return 0, fmt.Errorf("failed to parse time2: %v", err)
	}
	
	// Compare times
	return frameTime1.Compare(frameTime2), nil
}

// CalculateEndTime calculates the end time given start time and duration
func (ta *TimeArithmetic) CalculateEndTime(startTime, duration string) (string, error) {
	return ta.AddTimes(startTime, duration)
}

// ValidateTimeRange validates that an element fits within a time range
func (ta *TimeArithmetic) ValidateTimeRange(elementStart, elementDuration, rangeStart, rangeDuration string) error {
	// Calculate element end time
	elementEnd, err := ta.CalculateEndTime(elementStart, elementDuration)
	if err != nil {
		return fmt.Errorf("failed to calculate element end time: %v", err)
	}
	
	// Calculate range end time
	rangeEnd, err := ta.CalculateEndTime(rangeStart, rangeDuration)
	if err != nil {
		return fmt.Errorf("failed to calculate range end time: %v", err)
	}
	
	// Check if element start is after range start
	if startComparison, err := ta.CompareTimes(elementStart, rangeStart); err != nil {
		return fmt.Errorf("failed to compare start times: %v", err)
	} else if startComparison < 0 {
		return fmt.Errorf("element starts before range: %s < %s", elementStart, rangeStart)
	}
	
	// Check if element end is before range end
	if endComparison, err := ta.CompareTimes(elementEnd, rangeEnd); err != nil {
		return fmt.Errorf("failed to compare end times: %v", err)
	} else if endComparison > 0 {
		return fmt.Errorf("element ends after range: %s > %s", elementEnd, rangeEnd)
	}
	
	return nil
}

// GetFrameCount returns the number of frames in a duration
func (ta *TimeArithmetic) GetFrameCount(duration string) (int, error) {
	if err := ta.validator.ValidateTimeString(duration); err != nil {
		return 0, fmt.Errorf("invalid duration: %v", err)
	}
	
	frameAccurateDuration, err := ta.validator.ConvertToFrameAccurate(duration)
	if err != nil {
		return 0, fmt.Errorf("failed to convert duration: %v", err)
	}
	
	frameTime, err := NewFrameAccurateTimeFromFCPString(frameAccurateDuration)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}
	
	return frameTime.GetFrames(), nil
}