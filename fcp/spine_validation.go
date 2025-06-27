// Package spine_validation implements Step 9 of the FCPXMLKit-inspired refactoring plan:
// Spine element ordering validation with automatic chronological sorting.
//
// This replaces manual bubble sort implementations with validation-aware sorting
// that ensures proper spine element ordering and prevents timeline inconsistencies.
package fcp

import (
	"fmt"
	"sort"
	"strconv"
)

// SpineElementType represents different types of spine elements
type SpineElementType int

const (
	SpineElementUnknown SpineElementType = iota
	SpineElementAssetClip
	SpineElementVideo
	SpineElementTitle
	SpineElementGeneratorClip
	SpineElementGap
)

// String returns the string representation of the spine element type
func (set SpineElementType) String() string {
	switch set {
	case SpineElementAssetClip:
		return "asset-clip"
	case SpineElementVideo:
		return "video"
	case SpineElementTitle:
		return "title"
	case SpineElementGeneratorClip:
		return "generator-clip"
	case SpineElementGap:
		return "gap"
	default:
		return "unknown"
	}
}

// ValidatedSpineElement represents a spine element with validation
type ValidatedSpineElement struct {
	ID       string
	Type     SpineElementType
	Offset   Time
	Duration Duration
	Lane     Lane
	Name     string
	
	// Type-specific data
	AssetClip     *AssetClip
	Video         *Video
	Title         *Title
	GeneratorClip *GeneratorClip
	Gap           *Gap
}

// NewValidatedSpineElement creates a validated spine element
func NewValidatedSpineElement(elementType SpineElementType, offset Time, duration Duration, lane Lane, name string) (*ValidatedSpineElement, error) {
	if err := offset.Validate(); err != nil {
		return nil, fmt.Errorf("invalid offset: %v", err)
	}
	
	if err := duration.Validate(); err != nil {
		return nil, fmt.Errorf("invalid duration: %v", err)
	}
	
	if err := lane.Validate(); err != nil {
		return nil, fmt.Errorf("invalid lane: %v", err)
	}
	
	if name == "" {
		return nil, fmt.Errorf("element name cannot be empty")
	}
	
	element := &ValidatedSpineElement{
		Type:     elementType,
		Offset:   offset,
		Duration: duration,
		Lane:     lane,
		Name:     name,
	}
	
	// Generate ID based on name and type
	element.ID = fmt.Sprintf("%s_%s", elementType.String(), name)
	
	return element, nil
}

// GetEndTime calculates the end time of this element
func (vse *ValidatedSpineElement) GetEndTime() (Time, error) {
	offsetSeconds, err := vse.Offset.ToSeconds()
	if err != nil {
		return "", fmt.Errorf("failed to convert offset: %v", err)
	}
	
	durationSeconds, err := vse.Duration.ToSeconds()
	if err != nil {
		return "", fmt.Errorf("failed to convert duration: %v", err)
	}
	
	return NewTimeFromSeconds(offsetSeconds + durationSeconds), nil
}

// SetAssetClip sets the asset clip data for this element
func (vse *ValidatedSpineElement) SetAssetClip(clip *AssetClip) error {
	if vse.Type != SpineElementAssetClip {
		return fmt.Errorf("element type is %s, cannot set asset clip", vse.Type.String())
	}
	
	vse.AssetClip = clip
	return nil
}

// SetVideo sets the video data for this element
func (vse *ValidatedSpineElement) SetVideo(video *Video) error {
	if vse.Type != SpineElementVideo {
		return fmt.Errorf("element type is %s, cannot set video", vse.Type.String())
	}
	
	vse.Video = video
	return nil
}

// SetTitle sets the title data for this element
func (vse *ValidatedSpineElement) SetTitle(title *Title) error {
	if vse.Type != SpineElementTitle {
		return fmt.Errorf("element type is %s, cannot set title", vse.Type.String())
	}
	
	vse.Title = title
	return nil
}

// SetGeneratorClip sets the generator clip data for this element
func (vse *ValidatedSpineElement) SetGeneratorClip(generatorClip *GeneratorClip) error {
	if vse.Type != SpineElementGeneratorClip {
		return fmt.Errorf("element type is %s, cannot set generator clip", vse.Type.String())
	}
	
	vse.GeneratorClip = generatorClip
	return nil
}

// SetGap sets the gap data for this element
func (vse *ValidatedSpineElement) SetGap(gap *Gap) error {
	if vse.Type != SpineElementGap {
		return fmt.Errorf("element type is %s, cannot set gap", vse.Type.String())
	}
	
	vse.Gap = gap
	return nil
}

// SpineElementSorter provides validation-aware sorting of spine elements
type SpineElementSorter struct {
	elements      []*ValidatedSpineElement
	validator     *TimeArithmetic
	allowOverlaps bool
}

// NewSpineElementSorter creates a new spine element sorter
func NewSpineElementSorter() *SpineElementSorter {
	return &SpineElementSorter{
		elements:      make([]*ValidatedSpineElement, 0),
		validator:     NewTimeArithmetic(false), // Use lenient time validation
		allowOverlaps: false,                   // Default: disallow overlaps in same lane
	}
}

// SetAllowOverlaps controls whether overlaps in the same lane are allowed
func (ses *SpineElementSorter) SetAllowOverlaps(allow bool) {
	ses.allowOverlaps = allow
}

// AddElement adds an element to the sorter with validation
func (ses *SpineElementSorter) AddElement(element *ValidatedSpineElement) error {
	if element == nil {
		return fmt.Errorf("element cannot be nil")
	}
	
	// Validate element timing
	if err := ses.validateElementTiming(element); err != nil {
		return fmt.Errorf("element timing validation failed: %v", err)
	}
	
	// Check for overlaps if not allowed
	if !ses.allowOverlaps {
		if err := ses.validateNoOverlaps(element); err != nil {
			return fmt.Errorf("overlap validation failed: %v", err)
		}
	}
	
	ses.elements = append(ses.elements, element)
	return nil
}

// validateElementTiming validates the timing of an element
func (ses *SpineElementSorter) validateElementTiming(element *ValidatedSpineElement) error {
	// Validate offset and duration are frame-accurate
	if err := element.Offset.Validate(); err != nil {
		return fmt.Errorf("invalid offset: %v", err)
	}
	
	if err := element.Duration.Validate(); err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}
	
	// Validate that duration is positive (except for gaps which can be zero)
	if element.Type != SpineElementGap {
		durationSeconds, err := element.Duration.ToSeconds()
		if err != nil {
			return fmt.Errorf("failed to convert duration: %v", err)
		}
		
		if durationSeconds <= 0 {
			return fmt.Errorf("duration must be positive for %s elements", element.Type.String())
		}
	}
	
	return nil
}

// validateNoOverlaps checks if the new element overlaps with existing elements in the same lane
func (ses *SpineElementSorter) validateNoOverlaps(newElement *ValidatedSpineElement) error {
	for _, existing := range ses.elements {
		// Only check elements in the same lane
		if existing.Lane != newElement.Lane {
			continue
		}
		
		// Check for overlap
		if overlaps, err := ses.elementsOverlap(existing, newElement); err != nil {
			return fmt.Errorf("failed to check overlap: %v", err)
		} else if overlaps {
			return fmt.Errorf("element '%s' (lane %d, %s-%s) overlaps with existing element '%s' (%s-%s)",
				newElement.Name, newElement.Lane.Int(),
				newElement.Offset.String(), formatEndTime(newElement),
				existing.Name,
				existing.Offset.String(), formatEndTime(existing))
		}
	}
	
	return nil
}

// elementsOverlap checks if two elements overlap in time
func (ses *SpineElementSorter) elementsOverlap(elem1, elem2 *ValidatedSpineElement) (bool, error) {
	// Get timing information
	start1Seconds, err := elem1.Offset.ToSeconds()
	if err != nil {
		return false, err
	}
	
	duration1Seconds, err := elem1.Duration.ToSeconds()
	if err != nil {
		return false, err
	}
	
	start2Seconds, err := elem2.Offset.ToSeconds()
	if err != nil {
		return false, err
	}
	
	duration2Seconds, err := elem2.Duration.ToSeconds()
	if err != nil {
		return false, err
	}
	
	end1Seconds := start1Seconds + duration1Seconds
	end2Seconds := start2Seconds + duration2Seconds
	
	// No overlap if one element ends before the other starts
	return !(end1Seconds <= start2Seconds || end2Seconds <= start1Seconds), nil
}

// formatEndTime formats the end time of an element for error messages
func formatEndTime(element *ValidatedSpineElement) string {
	endTime, err := element.GetEndTime()
	if err != nil {
		return "unknown"
	}
	return endTime.String()
}

// SortAndValidate sorts all elements chronologically and validates the result
func (ses *SpineElementSorter) SortAndValidate() ([]*ValidatedSpineElement, error) {
	if len(ses.elements) == 0 {
		return []*ValidatedSpineElement{}, nil
	}
	
	// Create a copy for sorting
	sortedElements := make([]*ValidatedSpineElement, len(ses.elements))
	copy(sortedElements, ses.elements)
	
	// Sort by offset time, using frame-accurate comparison
	sort.Slice(sortedElements, func(i, j int) bool {
		timeI, errI := sortedElements[i].Offset.ToSeconds()
		timeJ, errJ := sortedElements[j].Offset.ToSeconds()
		
		if errI != nil || errJ != nil {
			// Fallback to string comparison if conversion fails
			return sortedElements[i].Offset.String() < sortedElements[j].Offset.String()
		}
		
		if timeI == timeJ {
			// If offsets are equal, sort by lane (lower lanes first)
			return sortedElements[i].Lane.Int() < sortedElements[j].Lane.Int()
		}
		
		return timeI < timeJ
	})
	
	// Validate sorted order
	if err := ses.validateSortedOrder(sortedElements); err != nil {
		return nil, fmt.Errorf("sorted order validation failed: %v", err)
	}
	
	return sortedElements, nil
}

// validateSortedOrder validates that elements are properly sorted
func (ses *SpineElementSorter) validateSortedOrder(elements []*ValidatedSpineElement) error {
	for i := 1; i < len(elements); i++ {
		prev := elements[i-1]
		curr := elements[i]
		
		prevTime, err := prev.Offset.ToSeconds()
		if err != nil {
			return fmt.Errorf("failed to convert previous element offset: %v", err)
		}
		
		currTime, err := curr.Offset.ToSeconds()
		if err != nil {
			return fmt.Errorf("failed to convert current element offset: %v", err)
		}
		
		if currTime < prevTime {
			return fmt.Errorf("elements not properly sorted: element %d (%s at %s) comes before element %d (%s at %s)",
				i, curr.Name, curr.Offset.String(),
				i-1, prev.Name, prev.Offset.String())
		}
	}
	
	return nil
}

// GetElementsByLane returns all elements in a specific lane
func (ses *SpineElementSorter) GetElementsByLane(lane Lane) []*ValidatedSpineElement {
	var laneElements []*ValidatedSpineElement
	
	for _, element := range ses.elements {
		if element.Lane == lane {
			laneElements = append(laneElements, element)
		}
	}
	
	return laneElements
}

// GetElementsByType returns all elements of a specific type
func (ses *SpineElementSorter) GetElementsByType(elementType SpineElementType) []*ValidatedSpineElement {
	var typeElements []*ValidatedSpineElement
	
	for _, element := range ses.elements {
		if element.Type == elementType {
			typeElements = append(typeElements, element)
		}
	}
	
	return typeElements
}

// GetElementsInTimeRange returns elements that intersect with the given time range
func (ses *SpineElementSorter) GetElementsInTimeRange(startTime, endTime Time) ([]*ValidatedSpineElement, error) {
	startSeconds, err := startTime.ToSeconds()
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %v", err)
	}
	
	endSeconds, err := endTime.ToSeconds()
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %v", err)
	}
	
	var intersectingElements []*ValidatedSpineElement
	
	for _, element := range ses.elements {
		elemStartSeconds, err := element.Offset.ToSeconds()
		if err != nil {
			continue // Skip elements with invalid timing
		}
		
		elemDurationSeconds, err := element.Duration.ToSeconds()
		if err != nil {
			continue // Skip elements with invalid timing
		}
		
		elemEndSeconds := elemStartSeconds + elemDurationSeconds
		
		// Check for intersection
		if !(elemEndSeconds <= startSeconds || elemStartSeconds >= endSeconds) {
			intersectingElements = append(intersectingElements, element)
		}
	}
	
	return intersectingElements, nil
}

// SpineBuilder provides a high-level interface for building validated spines
type SpineBuilder struct {
	sorter           *SpineElementSorter
	timelineValidator *TimelineValidator
	totalDuration    Duration
	registry         *ReferenceRegistry
}

// NewSpineBuilder creates a new spine builder
func NewSpineBuilder(totalDuration Duration, registry *ReferenceRegistry) (*SpineBuilder, error) {
	timelineValidator, err := NewTimelineValidator(totalDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create timeline validator: %v", err)
	}
	
	return &SpineBuilder{
		sorter:           NewSpineElementSorter(),
		timelineValidator: timelineValidator,
		totalDuration:    totalDuration,
		registry:         registry,
	}, nil
}

// SetAllowOverlaps controls whether overlaps are allowed
func (sb *SpineBuilder) SetAllowOverlaps(allow bool) {
	sb.sorter.SetAllowOverlaps(allow)
}

// AddAssetClip adds an asset clip to the spine
func (sb *SpineBuilder) AddAssetClip(ref, name string, offset Time, duration Duration, lane Lane) error {
	// Create validated spine element
	element, err := NewValidatedSpineElement(SpineElementAssetClip, offset, duration, lane, name)
	if err != nil {
		return fmt.Errorf("failed to create asset clip element: %v", err)
	}
	
	// Create asset clip data
	assetClip := &AssetClip{
		Ref:      ref,
		Offset:   offset.String(),
		Duration: duration.String(),
		Name:     name,
	}
	
	if lane != Lane(0) {
		assetClip.Lane = strconv.Itoa(lane.Int())
	}
	
	if err := element.SetAssetClip(assetClip); err != nil {
		return fmt.Errorf("failed to set asset clip data: %v", err)
	}
	
	// Add to timeline validator
	if err := sb.timelineValidator.AddElement(element.ID, offset, duration, lane, "asset-clip"); err != nil {
		return fmt.Errorf("timeline validation failed: %v", err)
	}
	
	// Add to sorter
	return sb.sorter.AddElement(element)
}

// AddVideo adds a video element to the spine
func (sb *SpineBuilder) AddVideo(ref, name string, offset Time, duration Duration, lane Lane) error {
	// Create validated spine element
	element, err := NewValidatedSpineElement(SpineElementVideo, offset, duration, lane, name)
	if err != nil {
		return fmt.Errorf("failed to create video element: %v", err)
	}
	
	// Create video data
	video := &Video{
		Ref:      ref,
		Offset:   offset.String(),
		Duration: duration.String(),
		Name:     name,
	}
	
	if lane != Lane(0) {
		video.Lane = strconv.Itoa(lane.Int())
	}
	
	if err := element.SetVideo(video); err != nil {
		return fmt.Errorf("failed to set video data: %v", err)
	}
	
	// Add to timeline validator
	if err := sb.timelineValidator.AddElement(element.ID, offset, duration, lane, "video"); err != nil {
		return fmt.Errorf("timeline validation failed: %v", err)
	}
	
	// Add to sorter
	return sb.sorter.AddElement(element)
}

// AddTitle adds a title element to the spine
func (sb *SpineBuilder) AddTitle(ref, name string, offset Time, duration Duration, lane Lane) error {
	// Create validated spine element
	element, err := NewValidatedSpineElement(SpineElementTitle, offset, duration, lane, name)
	if err != nil {
		return fmt.Errorf("failed to create title element: %v", err)
	}
	
	// Create title data
	title := &Title{
		Ref:      ref,
		Offset:   offset.String(),
		Duration: duration.String(),
		Name:     name,
	}
	
	if lane != Lane(0) {
		title.Lane = strconv.Itoa(lane.Int())
	}
	
	if err := element.SetTitle(title); err != nil {
		return fmt.Errorf("failed to set title data: %v", err)
	}
	
	// Add to timeline validator
	if err := sb.timelineValidator.AddElement(element.ID, offset, duration, lane, "title"); err != nil {
		return fmt.Errorf("timeline validation failed: %v", err)
	}
	
	// Add to sorter
	return sb.sorter.AddElement(element)
}

// AddGeneratorClip adds a generator clip to the spine
func (sb *SpineBuilder) AddGeneratorClip(ref, name string, offset Time, duration Duration, lane Lane) error {
	// Create validated spine element
	element, err := NewValidatedSpineElement(SpineElementGeneratorClip, offset, duration, lane, name)
	if err != nil {
		return fmt.Errorf("failed to create generator clip element: %v", err)
	}
	
	// Create generator clip data
	generatorClip := &GeneratorClip{
		Ref:      ref,
		Offset:   offset.String(),
		Duration: duration.String(),
		Name:     name,
	}
	
	if lane != Lane(0) {
		generatorClip.Lane = strconv.Itoa(lane.Int())
	}
	
	if err := element.SetGeneratorClip(generatorClip); err != nil {
		return fmt.Errorf("failed to set generator clip data: %v", err)
	}
	
	// Add to timeline validator
	if err := sb.timelineValidator.AddElement(element.ID, offset, duration, lane, "generator-clip"); err != nil {
		return fmt.Errorf("timeline validation failed: %v", err)
	}
	
	// Add to sorter
	return sb.sorter.AddElement(element)
}

// AddGap adds a gap element to the spine
func (sb *SpineBuilder) AddGap(name string, offset Time, duration Duration) error {
	// Create validated spine element (gaps always use lane 0)
	element, err := NewValidatedSpineElement(SpineElementGap, offset, duration, Lane(0), name)
	if err != nil {
		return fmt.Errorf("failed to create gap element: %v", err)
	}
	
	// Create gap data
	gap := &Gap{
		Name:     name,
		Offset:   offset.String(),
		Duration: duration.String(),
	}
	
	if err := element.SetGap(gap); err != nil {
		return fmt.Errorf("failed to set gap data: %v", err)
	}
	
	// Add to timeline validator (gaps use lane 0)
	if err := sb.timelineValidator.AddElement(element.ID, offset, duration, Lane(0), "gap"); err != nil {
		return fmt.Errorf("timeline validation failed: %v", err)
	}
	
	// Add to sorter
	return sb.sorter.AddElement(element)
}

// Build creates a validated spine with proper element ordering
func (sb *SpineBuilder) Build() (*Spine, error) {
	// Validate complete timeline
	if err := sb.timelineValidator.ValidateComplete(); err != nil {
		return nil, fmt.Errorf("timeline validation failed: %v", err)
	}
	
	// Sort elements
	sortedElements, err := sb.sorter.SortAndValidate()
	if err != nil {
		return nil, fmt.Errorf("element sorting failed: %v", err)
	}
	
	// Create spine structure
	spine := &Spine{
		AssetClips: make([]AssetClip, 0),
		Videos:     make([]Video, 0),
		Titles:     make([]Title, 0),
		Gaps:       make([]Gap, 0),
	}
	
	// Add elements to spine in sorted order
	for _, element := range sortedElements {
		switch element.Type {
		case SpineElementAssetClip:
			if element.AssetClip != nil {
				spine.AssetClips = append(spine.AssetClips, *element.AssetClip)
			}
		case SpineElementVideo:
			if element.Video != nil {
				spine.Videos = append(spine.Videos, *element.Video)
			}
		case SpineElementTitle:
			if element.Title != nil {
				spine.Titles = append(spine.Titles, *element.Title)
			}
		case SpineElementGeneratorClip:
			// NOTE: In FCPXML, generator clips typically appear within gaps
			// For direct spine placement, we'll skip this for now
			// TODO: Consider adding GeneratorClips field to Spine struct if needed
		case SpineElementGap:
			if element.Gap != nil {
				spine.Gaps = append(spine.Gaps, *element.Gap)
			}
		}
	}
	
	return spine, nil
}

// GetStatistics returns statistics about the spine being built
func (sb *SpineBuilder) GetStatistics() SpineStatistics {
	stats := SpineStatistics{
		TotalElements:  len(sb.sorter.elements),
		ElementsByType: make(map[string]int),
		ElementsByLane: make(map[int]int),
		UsedLanes:      make([]int, 0),
	}
	
	// Count elements by type and lane
	laneSet := make(map[int]bool)
	
	for _, element := range sb.sorter.elements {
		stats.ElementsByType[element.Type.String()]++
		stats.ElementsByLane[element.Lane.Int()]++
		laneSet[element.Lane.Int()] = true
	}
	
	// Extract used lanes
	for lane := range laneSet {
		stats.UsedLanes = append(stats.UsedLanes, lane)
	}
	
	sort.Ints(stats.UsedLanes)
	
	return stats
}

// SpineStatistics provides information about spine composition
type SpineStatistics struct {
	TotalElements  int
	ElementsByType map[string]int
	ElementsByLane map[int]int
	UsedLanes      []int
}