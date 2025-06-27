// Package validation_types implements validation-first type wrappers for FCPXML generation.
//
// This file implements Step 1 of the FCPXMLKit-inspired refactoring plan:
// Create validation-first type system with typed wrappers that enforce FCPXML rules.
//
// 🚨 CRITICAL: All FCPXML generation MUST use these validated types
// - Replace string-based attributes with typed wrappers
// - Validation happens at construction time, not runtime
// - Frame alignment is enforced automatically
// - Media type constraints are built into the type system
package fcp

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// FCP timing constants for frame-accurate calculations
const (
	FCPTimebase      = 24000
	FCPFrameDuration = 1001
	FCPFrameRate     = 23.976023976023976
)

// ID represents a validated FCPXML resource ID (e.g., "r1", "r2", etc.)
type ID string

// Validate ensures the ID follows FCPXML conventions
func (id ID) Validate() error {
	idStr := string(id)
	if idStr == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	
	if !strings.HasPrefix(idStr, "r") {
		return fmt.Errorf("ID must start with 'r': %s", idStr)
	}
	
	if len(idStr) < 2 {
		return fmt.Errorf("ID must be at least 2 characters: %s", idStr)
	}
	
	numPart := idStr[1:]
	if _, err := strconv.Atoi(numPart); err != nil {
		return fmt.Errorf("ID must be 'r' followed by number: %s", idStr)
	}
	
	return nil
}

// String returns the string representation
func (id ID) String() string {
	return string(id)
}

// Duration represents a validated FCP-compatible duration
type Duration string

// Validate ensures the duration is frame-aligned and follows FCP format
func (d Duration) Validate() error {
	durationStr := string(d)
	
	if durationStr == "" {
		return fmt.Errorf("duration cannot be empty")
	}
	
	if durationStr == "0s" {
		return nil // Valid for images
	}
	
	if !strings.HasSuffix(durationStr, "s") {
		return fmt.Errorf("duration must end with 's': %s", durationStr)
	}
	
	if !strings.Contains(durationStr, "/") {
		return fmt.Errorf("duration must be in rational format: %s", durationStr)
	}
	
	return validateFrameAlignment(durationStr)
}

// String returns the string representation
func (d Duration) String() string {
	return string(d)
}

// ToSeconds converts the duration to seconds for calculations
func (d Duration) ToSeconds() (float64, error) {
	if err := d.Validate(); err != nil {
		return 0, err
	}
	
	durationStr := string(d)
	if durationStr == "0s" {
		return 0.0, nil
	}
	
	// Parse rational format
	timeNoS := strings.TrimSuffix(durationStr, "s")
	parts := strings.Split(timeNoS, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rational format: %s", durationStr)
	}
	
	numerator, err1 := strconv.ParseFloat(parts[0], 64)
	denominator, err2 := strconv.ParseFloat(parts[1], 64)
	
	if err1 != nil || err2 != nil || denominator == 0 {
		return 0, fmt.Errorf("invalid rational parts: %s", durationStr)
	}
	
	return numerator / denominator, nil
}

// Time represents a validated FCP-compatible time offset
type Time string

// Validate ensures the time is frame-aligned and follows FCP format
func (t Time) Validate() error {
	timeStr := string(t)
	
	if timeStr == "" {
		return fmt.Errorf("time cannot be empty")
	}
	
	if timeStr == "0s" {
		return nil // Valid start time
	}
	
	if !strings.HasSuffix(timeStr, "s") {
		return fmt.Errorf("time must end with 's': %s", timeStr)
	}
	
	if !strings.Contains(timeStr, "/") {
		return fmt.Errorf("time must be in rational format: %s", timeStr)
	}
	
	return validateFrameAlignment(timeStr)
}

// String returns the string representation
func (t Time) String() string {
	return string(t)
}

// ToSeconds converts the time to seconds for calculations
func (t Time) ToSeconds() (float64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	
	timeStr := string(t)
	if timeStr == "0s" {
		return 0.0, nil
	}
	
	// Parse rational format
	timeNoS := strings.TrimSuffix(timeStr, "s")
	parts := strings.Split(timeNoS, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rational format: %s", timeStr)
	}
	
	numerator, err1 := strconv.ParseFloat(parts[0], 64)
	denominator, err2 := strconv.ParseFloat(parts[1], 64)
	
	if err1 != nil || err2 != nil || denominator == 0 {
		return 0, fmt.Errorf("invalid rational parts: %s", timeStr)
	}
	
	return numerator / denominator, nil
}

// Lane represents a validated timeline lane number
type Lane int

// Validate ensures the lane is within acceptable bounds
func (l Lane) Validate() error {
	if l < -10 || l > 10 {
		return fmt.Errorf("lane out of range [-10, 10]: %d", l)
	}
	return nil
}

// String returns the string representation for XML attributes
func (l Lane) String() string {
	if l == 0 {
		return "" // Lane 0 is omitted in XML
	}
	return strconv.Itoa(int(l))
}

// Int returns the integer value
func (l Lane) Int() int {
	return int(l)
}

// validateFrameAlignment ensures a time/duration string is frame-aligned with FCP's timebase
func validateFrameAlignment(timeStr string) error {
	if !strings.HasSuffix(timeStr, "s") {
		return fmt.Errorf("time must end with 's': %s", timeStr)
	}
	
	timeNoS := strings.TrimSuffix(timeStr, "s")
	
	if !strings.Contains(timeNoS, "/") {
		return fmt.Errorf("time must be in rational format: %s", timeStr)
	}
	
	parts := strings.Split(timeNoS, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid rational format: %s", timeStr)
	}
	
	numerator, err1 := strconv.Atoi(parts[0])
	denominator, err2 := strconv.Atoi(parts[1])
	
	if err1 != nil || err2 != nil {
		return fmt.Errorf("non-integer rational parts: %s", timeStr)
	}
	
	if denominator != FCPTimebase {
		return fmt.Errorf("wrong timebase, expected %d, got %d", FCPTimebase, denominator)
	}
	
	if numerator%FCPFrameDuration != 0 {
		return fmt.Errorf("time not frame-aligned: %s (numerator must be multiple of %d)", timeStr, FCPFrameDuration)
	}
	
	return nil
}

// NewDurationFromSeconds creates a validated Duration from seconds
func NewDurationFromSeconds(seconds float64) Duration {
	if seconds == 0 {
		return Duration("0s")
	}
	
	// Calculate exact frame count
	frames := int(seconds*FCPFrameRate + 0.5) // Round to nearest frame
	
	// Convert to FCP's rational format
	numerator := frames * FCPFrameDuration
	
	return Duration(fmt.Sprintf("%d/%ds", numerator, FCPTimebase))
}

// NewTimeFromSeconds creates a validated Time from seconds
func NewTimeFromSeconds(seconds float64) Time {
	if seconds == 0 {
		return Time("0s")
	}
	
	// Calculate exact frame count
	frames := int(seconds*FCPFrameRate + 0.5) // Round to nearest frame
	
	// Convert to FCP's rational format
	numerator := frames * FCPFrameDuration
	
	return Time(fmt.Sprintf("%d/%ds", numerator, FCPTimebase))
}

// NewID creates a validated ID from an integer
func NewID(number int) (ID, error) {
	if number < 1 {
		return "", fmt.Errorf("ID number must be positive: %d", number)
	}
	
	id := ID(fmt.Sprintf("r%d", number))
	if err := id.Validate(); err != nil {
		return "", err
	}
	
	return id, nil
}

// NewLane creates a validated Lane
func NewLane(number int) (Lane, error) {
	lane := Lane(number)
	if err := lane.Validate(); err != nil {
		return 0, err
	}
	return lane, nil
}

// AddTimes adds two Time values and returns a new Time
func AddTimes(t1, t2 Time) (Time, error) {
	seconds1, err := t1.ToSeconds()
	if err != nil {
		return "", fmt.Errorf("invalid time 1: %v", err)
	}
	
	seconds2, err := t2.ToSeconds()
	if err != nil {
		return "", fmt.Errorf("invalid time 2: %v", err)
	}
	
	return NewTimeFromSeconds(seconds1 + seconds2), nil
}

// AddDurations adds two Duration values and returns a new Duration
func AddDurations(d1, d2 Duration) (Duration, error) {
	seconds1, err := d1.ToSeconds()
	if err != nil {
		return "", fmt.Errorf("invalid duration 1: %v", err)
	}
	
	seconds2, err := d2.ToSeconds()
	if err != nil {
		return "", fmt.Errorf("invalid duration 2: %v", err)
	}
	
	return NewDurationFromSeconds(seconds1 + seconds2), nil
}

// CompareTimes compares two Time values (-1 if t1 < t2, 0 if equal, 1 if t1 > t2)
func CompareTimes(t1, t2 Time) (int, error) {
	seconds1, err := t1.ToSeconds()
	if err != nil {
		return 0, err
	}
	
	seconds2, err := t2.ToSeconds()
	if err != nil {
		return 0, err
	}
	
	if seconds1 < seconds2 {
		return -1, nil
	} else if seconds1 > seconds2 {
		return 1, nil
	}
	return 0, nil
}

// UID represents a validated media UID for asset identification
type UID string

// Validate ensures the UID follows proper format
func (uid UID) Validate() error {
	uidStr := string(uid)
	if uidStr == "" {
		return fmt.Errorf("UID cannot be empty")
	}
	
	// UID should be uppercase alphanumeric, typically MD5 hash format
	matched, err := regexp.MatchString("^[A-F0-9-]+$", uidStr)
	if err != nil {
		return fmt.Errorf("invalid UID regex: %v", err)
	}
	
	if !matched {
		return fmt.Errorf("UID must be uppercase alphanumeric with hyphens: %s", uidStr)
	}
	
	return nil
}

// String returns the string representation
func (uid UID) String() string {
	return string(uid)
}

// ColorSpace represents a validated color space specification
type ColorSpace string

// Validate ensures the color space follows FCP format
func (cs ColorSpace) Validate() error {
	csStr := string(cs)
	if csStr == "" {
		return fmt.Errorf("color space cannot be empty")
	}
	
	// Common FCP color spaces
	validColorSpaces := []string{
		"1-1-1 (Rec. 709)",
		"1-13-1",
		"1-1-1",
		"9-1-1 (Rec. 2020)",
		"1-14-18 (Rec. 2020 HLG)",
		"1-16-18 (Rec. 2020 PQ)",
	}
	
	for _, valid := range validColorSpaces {
		if csStr == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid color space: %s", csStr)
}

// String returns the string representation
func (cs ColorSpace) String() string {
	return string(cs)
}

// MediaType represents the type of media (image, video, audio)
type MediaType int

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeImage
	MediaTypeVideo
	MediaTypeAudio
)

// String returns the string representation
func (mt MediaType) String() string {
	switch mt {
	case MediaTypeImage:
		return "image"
	case MediaTypeVideo:
		return "video"
	case MediaTypeAudio:
		return "audio"
	default:
		return "unknown"
	}
}

// Validate ensures the media type is valid
func (mt MediaType) Validate() error {
	if mt == MediaTypeUnknown {
		return fmt.Errorf("media type cannot be unknown")
	}
	return nil
}

// DetectMediaTypeFromPath determines media type from file extension
func DetectMediaTypeFromPath(filePath string) (MediaType, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	
	switch ext {
	case "png", "jpg", "jpeg", "gif", "bmp", "tiff", "tif", "webp":
		return MediaTypeImage, nil
	case "mp4", "mov", "avi", "mkv", "m4v", "webm", "mpg", "mpeg":
		return MediaTypeVideo, nil
	case "mp3", "wav", "aac", "m4a", "flac", "ogg", "wma":
		return MediaTypeAudio, nil
	default:
		return MediaTypeUnknown, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// AudioRate represents a validated audio sample rate
type AudioRate string

// Validate ensures the audio rate is supported by FCP
func (ar AudioRate) Validate() error {
	rateStr := string(ar)
	validRates := []string{"44100", "48000", "96000", "192000"}
	
	for _, valid := range validRates {
		if rateStr == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid audio rate: %s (must be 44100, 48000, 96000, or 192000)", rateStr)
}

// String returns the string representation
func (ar AudioRate) String() string {
	return string(ar)
}