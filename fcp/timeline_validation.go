// Package timeline_validation implements Step 6 of the FCPXMLKit-inspired refactoring plan:
// Timeline constraint validation for ensuring elements fit within timeline bounds
// and lane assignments are valid.
//
// This provides timeline-aware validation that prevents common FCPXML generation errors
// like elements extending beyond sequence duration or invalid lane configurations.
package fcp

import (
	"fmt"
	"sort"
)

// TimeRange represents a time span with start, duration, and calculated end
type TimeRange struct {
	Start    Time
	Duration Duration
	End      Time
}

// NewTimeRange creates a validated TimeRange
func NewTimeRange(start Time, duration Duration) (*TimeRange, error) {
	if err := start.Validate(); err != nil {
		return nil, fmt.Errorf("invalid start time: %v", err)
	}
	
	if err := duration.Validate(); err != nil {
		return nil, fmt.Errorf("invalid duration: %v", err)
	}
	
	// Calculate end time
	startSeconds, err := start.ToSeconds()
	if err != nil {
		return nil, fmt.Errorf("failed to convert start time: %v", err)
	}
	
	durationSeconds, err := duration.ToSeconds()
	if err != nil {
		return nil, fmt.Errorf("failed to convert duration: %v", err)
	}
	
	endTime := NewTimeFromSeconds(startSeconds + durationSeconds)
	
	return &TimeRange{
		Start:    start,
		Duration: duration,
		End:      endTime,
	}, nil
}

// Contains checks if this range contains another range
func (tr *TimeRange) Contains(other *TimeRange) (bool, error) {
	startSeconds, err := tr.Start.ToSeconds()
	if err != nil {
		return false, err
	}
	
	endSeconds, err := tr.End.ToSeconds()
	if err != nil {
		return false, err
	}
	
	otherStartSeconds, err := other.Start.ToSeconds()
	if err != nil {
		return false, err
	}
	
	otherEndSeconds, err := other.End.ToSeconds()
	if err != nil {
		return false, err
	}
	
	return startSeconds <= otherStartSeconds && otherEndSeconds <= endSeconds, nil
}

// Overlaps checks if this range overlaps with another range
func (tr *TimeRange) Overlaps(other *TimeRange) (bool, error) {
	startSeconds, err := tr.Start.ToSeconds()
	if err != nil {
		return false, err
	}
	
	endSeconds, err := tr.End.ToSeconds()
	if err != nil {
		return false, err
	}
	
	otherStartSeconds, err := other.Start.ToSeconds()
	if err != nil {
		return false, err
	}
	
	otherEndSeconds, err := other.End.ToSeconds()
	if err != nil {
		return false, err
	}
	
	// No overlap if one range ends before the other starts
	return !(endSeconds <= otherStartSeconds || otherEndSeconds <= startSeconds), nil
}

// ValidatedTimelineElement represents an element on the timeline with timing and lane info
type ValidatedTimelineElement struct {
	ID       string
	Offset   Time
	Duration Duration
	Lane     Lane
	Range    *TimeRange
	Type     string // "asset-clip", "video", "title", etc.
}

// NewValidatedTimelineElement creates a validated timeline element
func NewValidatedTimelineElement(id string, offset Time, duration Duration, lane Lane, elementType string) (*ValidatedTimelineElement, error) {
	if err := offset.Validate(); err != nil {
		return nil, fmt.Errorf("invalid offset: %v", err)
	}
	
	if err := duration.Validate(); err != nil {
		return nil, fmt.Errorf("invalid duration: %v", err)
	}
	
	if err := lane.Validate(); err != nil {
		return nil, fmt.Errorf("invalid lane: %v", err)
	}
	
	timeRange, err := NewTimeRange(offset, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create time range: %v", err)
	}
	
	return &ValidatedTimelineElement{
		ID:       id,
		Offset:   offset,
		Duration: duration,
		Lane:     lane,
		Range:    timeRange,
		Type:     elementType,
	}, nil
}

// TimelineValidator validates timeline structure and constraints
type TimelineValidator struct {
	totalDuration     Duration
	totalRange        *TimeRange
	elements          []*ValidatedTimelineElement
	laneAssignments   map[Lane][]*ValidatedTimelineElement
	maxLanes          int
	allowGaps         bool
	allowOverlaps     bool
}

// NewTimelineValidator creates a new timeline validator
func NewTimelineValidator(totalDuration Duration) (*TimelineValidator, error) {
	if err := totalDuration.Validate(); err != nil {
		return nil, fmt.Errorf("invalid total duration: %v", err)
	}
	
	totalRange, err := NewTimeRange(Time("0s"), totalDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create total range: %v", err)
	}
	
	return &TimelineValidator{
		totalDuration:   totalDuration,
		totalRange:      totalRange,
		elements:        make([]*ValidatedTimelineElement, 0),
		laneAssignments: make(map[Lane][]*ValidatedTimelineElement),
		maxLanes:        10, // Default maximum lanes
		allowGaps:       false, // Default: require continuous lane numbering
		allowOverlaps:   false, // Default: disallow overlaps
	}, nil
}

// SetMaxLanes sets the maximum allowed lane number
func (tv *TimelineValidator) SetMaxLanes(maxLanes int) {
	tv.maxLanes = maxLanes
}

// SetAllowGaps controls whether lane gaps are allowed
func (tv *TimelineValidator) SetAllowGaps(allowGaps bool) {
	tv.allowGaps = allowGaps
}

// SetAllowOverlaps controls whether overlapping elements in the same lane are allowed
func (tv *TimelineValidator) SetAllowOverlaps(allowOverlaps bool) {
	tv.allowOverlaps = allowOverlaps
}

// AddElement adds an element to the timeline for validation
func (tv *TimelineValidator) AddElement(id string, offset Time, duration Duration, lane Lane, elementType string) error {
	element, err := NewValidatedTimelineElement(id, offset, duration, lane, elementType)
	if err != nil {
		return fmt.Errorf("failed to create timeline element: %v", err)
	}
	
	// Validate element against timeline constraints
	if err := tv.validateTimelineElement(element); err != nil {
		return fmt.Errorf("timeline element validation failed: %v", err)
	}
	
	// Add to tracking structures
	tv.elements = append(tv.elements, element)
	tv.laneAssignments[lane] = append(tv.laneAssignments[lane], element)
	
	return nil
}

// validateTimelineElement validates a single element against timeline constraints
func (tv *TimelineValidator) validateTimelineElement(element *ValidatedTimelineElement) error {
	// Check if element fits within timeline bounds
	contains, err := tv.totalRange.Contains(element.Range)
	if err != nil {
		return fmt.Errorf("failed to check timeline bounds: %v", err)
	}
	
	if !contains {
		endSeconds, _ := element.Range.End.ToSeconds()
		totalSeconds, _ := tv.totalDuration.ToSeconds()
		return fmt.Errorf("element '%s' extends beyond timeline: ends at %.3fs but timeline is %.3fs", 
			element.ID, endSeconds, totalSeconds)
	}
	
	// Validate lane assignment
	if err := element.Lane.Validate(); err != nil {
		return fmt.Errorf("invalid lane for element '%s': %v", element.ID, err)
	}
	
	// Check maximum lane constraint
	if element.Lane.Int() > tv.maxLanes {
		return fmt.Errorf("element '%s' lane %d exceeds maximum %d", 
			element.ID, element.Lane.Int(), tv.maxLanes)
	}
	
	return nil
}

// ValidateLaneStructure validates the overall lane structure
func (tv *TimelineValidator) ValidateLaneStructure() error {
	if len(tv.laneAssignments) == 0 {
		return nil // No elements to validate
	}
	
	// Collect used positive lanes (ignore lane 0 and negative lanes for gap checking)
	var usedLanes []int
	for lane := range tv.laneAssignments {
		if lane.Int() > 0 {
			usedLanes = append(usedLanes, lane.Int())
		}
	}
	
	if len(usedLanes) == 0 {
		return nil // Only lane 0 or negative lanes used
	}
	
	// Sort lanes for gap checking
	sort.Ints(usedLanes)
	
	// Check for lane gaps if not allowed
	if !tv.allowGaps {
		for i := 1; i <= usedLanes[len(usedLanes)-1]; i++ {
			found := false
			for _, used := range usedLanes {
				if used == i {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("lane gap detected: missing lane %d (used lanes: %v)", i, usedLanes)
			}
		}
	}
	
	return nil
}

// ValidateOverlaps checks for overlapping elements in the same lane
func (tv *TimelineValidator) ValidateOverlaps() error {
	for lane, elements := range tv.laneAssignments {
		if len(elements) <= 1 {
			continue // No overlaps possible
		}
		
		// Sort elements by start time
		sortedElements := make([]*ValidatedTimelineElement, len(elements))
		copy(sortedElements, elements)
		sort.Slice(sortedElements, func(i, j int) bool {
			startI, _ := sortedElements[i].Offset.ToSeconds()
			startJ, _ := sortedElements[j].Offset.ToSeconds()
			return startI < startJ
		})
		
		// Check for overlaps
		for i := 0; i < len(sortedElements)-1; i++ {
			current := sortedElements[i]
			next := sortedElements[i+1]
			
			overlaps, err := current.Range.Overlaps(next.Range)
			if err != nil {
				return fmt.Errorf("failed to check overlap between elements: %v", err)
			}
			
			if overlaps {
				currentEnd, _ := current.Range.End.ToSeconds()
				nextStart, _ := next.Range.Start.ToSeconds()
				return fmt.Errorf("overlap detected in lane %d: element '%s' ends at %.3fs but element '%s' starts at %.3fs", 
					lane.Int(), current.ID, currentEnd, next.ID, nextStart)
			}
		}
	}
	
	return nil
}

// GetTimelineStatistics returns statistics about the timeline
func (tv *TimelineValidator) GetTimelineStatistics() TimelineStatistics {
	stats := TimelineStatistics{
		TotalElements:   len(tv.elements),
		UsedLanes:       make([]int, 0),
		ElementsByLane:  make(map[int]int),
		ElementsByType:  make(map[string]int),
	}
	
	// Collect lane usage
	for lane, elements := range tv.laneAssignments {
		laneNum := lane.Int()
		stats.UsedLanes = append(stats.UsedLanes, laneNum)
		stats.ElementsByLane[laneNum] = len(elements)
	}
	
	// Sort lanes
	sort.Ints(stats.UsedLanes)
	
	// Collect type usage
	for _, element := range tv.elements {
		stats.ElementsByType[element.Type]++
	}
	
	// Calculate timeline utilization
	if len(tv.elements) > 0 {
		totalSeconds, _ := tv.totalDuration.ToSeconds()
		usedSeconds := 0.0
		
		for _, element := range tv.elements {
			elementSeconds, _ := element.Duration.ToSeconds()
			usedSeconds += elementSeconds
		}
		
		stats.TimelineUtilization = usedSeconds / totalSeconds
	}
	
	return stats
}

// TimelineStatistics provides information about timeline usage
type TimelineStatistics struct {
	TotalElements       int
	UsedLanes          []int
	ElementsByLane     map[int]int
	ElementsByType     map[string]int
	TimelineUtilization float64
}

// ValidateComplete performs comprehensive timeline validation
func (tv *TimelineValidator) ValidateComplete() error {
	// Validate lane structure
	if err := tv.ValidateLaneStructure(); err != nil {
		return fmt.Errorf("lane structure validation failed: %v", err)
	}
	
	// Validate overlaps (only if not allowed)
	if !tv.allowOverlaps {
		if err := tv.ValidateOverlaps(); err != nil {
			return fmt.Errorf("overlap validation failed: %v", err)
		}
	}
	
	return nil
}

// GetElementsInTimeRange returns all elements that intersect with the given time range
func (tv *TimelineValidator) GetElementsInTimeRange(start Time, duration Duration) ([]*ValidatedTimelineElement, error) {
	queryRange, err := NewTimeRange(start, duration)
	if err != nil {
		return nil, fmt.Errorf("invalid query range: %v", err)
	}
	
	var intersectingElements []*ValidatedTimelineElement
	
	for _, element := range tv.elements {
		overlaps, err := element.Range.Overlaps(queryRange)
		if err != nil {
			return nil, fmt.Errorf("failed to check overlap: %v", err)
		}
		
		if overlaps {
			intersectingElements = append(intersectingElements, element)
		}
	}
	
	return intersectingElements, nil
}

// GetElementsInLane returns all elements in a specific lane
func (tv *TimelineValidator) GetElementsInLane(lane Lane) []*ValidatedTimelineElement {
	return tv.laneAssignments[lane]
}

// Clear resets the timeline validator for reuse
func (tv *TimelineValidator) Clear() {
	tv.elements = make([]*ValidatedTimelineElement, 0)
	tv.laneAssignments = make(map[Lane][]*ValidatedTimelineElement)
}