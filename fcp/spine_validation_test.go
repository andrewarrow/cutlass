package fcp

import (
	"testing"
)

func TestSpineElementType_String(t *testing.T) {
	tests := []struct {
		name     string
		elemType SpineElementType
		expected string
	}{
		{
			name:     "asset-clip type",
			elemType: SpineElementAssetClip,
			expected: "asset-clip",
		},
		{
			name:     "video type",
			elemType: SpineElementVideo,
			expected: "video",
		},
		{
			name:     "title type",
			elemType: SpineElementTitle,
			expected: "title",
		},
		{
			name:     "unknown type",
			elemType: SpineElementUnknown,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.elemType.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNewValidatedSpineElement(t *testing.T) {
	tests := []struct {
		name        string
		elementType SpineElementType
		offset      Time
		duration    Duration
		lane        Lane
		elemName    string
		expectError bool
	}{
		{
			name:        "valid asset clip element",
			elementType: SpineElementAssetClip,
			offset:      Time("0s"),
			duration:    Duration("240240/24000s"),
			lane:        Lane(1),
			elemName:    "TestClip",
		},
		{
			name:        "valid video element",
			elementType: SpineElementVideo,
			offset:      Time("120120/24000s"),
			duration:    Duration("240240/24000s"),
			lane:        Lane(2),
			elemName:    "TestVideo",
		},
		{
			name:        "invalid offset",
			elementType: SpineElementAssetClip,
			offset:      Time("invalid"),
			duration:    Duration("240240/24000s"),
			lane:        Lane(1),
			elemName:    "TestClip",
			expectError: true,
		},
		{
			name:        "invalid duration",
			elementType: SpineElementAssetClip,
			offset:      Time("0s"),
			duration:    Duration("invalid"),
			lane:        Lane(1),
			elemName:    "TestClip",
			expectError: true,
		},
		{
			name:        "invalid lane",
			elementType: SpineElementAssetClip,
			offset:      Time("0s"),
			duration:    Duration("240240/24000s"),
			lane:        Lane(15), // Out of range
			elemName:    "TestClip",
			expectError: true,
		},
		{
			name:        "empty name",
			elementType: SpineElementAssetClip,
			offset:      Time("0s"),
			duration:    Duration("240240/24000s"),
			lane:        Lane(1),
			elemName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element, err := NewValidatedSpineElement(tt.elementType, tt.offset, tt.duration, tt.lane, tt.elemName)
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

			// Verify element properties
			if element.Type != tt.elementType {
				t.Errorf("expected type %d, got %d", tt.elementType, element.Type)
			}
			if element.Offset != tt.offset {
				t.Errorf("expected offset %s, got %s", tt.offset, element.Offset)
			}
			if element.Duration != tt.duration {
				t.Errorf("expected duration %s, got %s", tt.duration, element.Duration)
			}
			if element.Lane != tt.lane {
				t.Errorf("expected lane %d, got %d", tt.lane, element.Lane)
			}
			if element.Name != tt.elemName {
				t.Errorf("expected name %s, got %s", tt.elemName, element.Name)
			}
		})
	}
}

func TestValidatedSpineElement_GetEndTime(t *testing.T) {
	element, err := NewValidatedSpineElement(SpineElementAssetClip, Time("240240/24000s"), Duration("240240/24000s"), Lane(1), "TestClip")
	if err != nil {
		t.Fatalf("failed to create element: %v", err)
	}

	endTime, err := element.GetEndTime()
	if err != nil {
		t.Errorf("unexpected error getting end time: %v", err)
		return
	}

	// End time should be offset + duration = 10s + 10s = 20s = 480480/24000s
	expected := Time("480480/24000s")
	if endTime != expected {
		t.Errorf("expected end time %s, got %s", expected, endTime)
	}
}

func TestValidatedSpineElement_SetTypeSpecificData(t *testing.T) {
	// Test asset clip
	assetElement, _ := NewValidatedSpineElement(SpineElementAssetClip, Time("0s"), Duration("240240/24000s"), Lane(1), "TestClip")
	assetClip := &AssetClip{Ref: "r1", Name: "TestClip"}

	err := assetElement.SetAssetClip(assetClip)
	if err != nil {
		t.Errorf("unexpected error setting asset clip: %v", err)
	}

	// Try to set wrong type
	err = assetElement.SetVideo(&Video{})
	if err == nil {
		t.Errorf("expected error setting video on asset clip element")
	}

	// Test video
	videoElement, _ := NewValidatedSpineElement(SpineElementVideo, Time("0s"), Duration("240240/24000s"), Lane(1), "TestVideo")
	video := &Video{Ref: "r2", Name: "TestVideo"}

	err = videoElement.SetVideo(video)
	if err != nil {
		t.Errorf("unexpected error setting video: %v", err)
	}

	// Test title
	titleElement, _ := NewValidatedSpineElement(SpineElementTitle, Time("0s"), Duration("240240/24000s"), Lane(1), "TestTitle")
	title := &Title{Ref: "r3", Name: "TestTitle"}

	err = titleElement.SetTitle(title)
	if err != nil {
		t.Errorf("unexpected error setting title: %v", err)
	}
}

func TestSpineElementSorter_AddElement(t *testing.T) {
	sorter := NewSpineElementSorter()

	// Add valid elements
	element1, _ := NewValidatedSpineElement(SpineElementAssetClip, Time("0s"), Duration("240240/24000s"), Lane(1), "Clip1")
	err := sorter.AddElement(element1)
	if err != nil {
		t.Errorf("unexpected error adding element1: %v", err)
	}

	element2, _ := NewValidatedSpineElement(SpineElementVideo, Time("240240/24000s"), Duration("240240/24000s"), Lane(1), "Video1")
	err = sorter.AddElement(element2)
	if err != nil {
		t.Errorf("unexpected error adding element2: %v", err)
	}

	// Try to add overlapping element (should fail)
	overlappingElement, _ := NewValidatedSpineElement(SpineElementTitle, Time("120120/24000s"), Duration("240240/24000s"), Lane(1), "Title1")
	err = sorter.AddElement(overlappingElement)
	if err == nil {
		t.Errorf("expected error adding overlapping element")
	}

	// Allow overlaps and try again
	sorter.SetAllowOverlaps(true)
	err = sorter.AddElement(overlappingElement)
	if err != nil {
		t.Errorf("unexpected error adding overlapping element with overlaps allowed: %v", err)
	}
}

func TestSpineElementSorter_SortAndValidate(t *testing.T) {
	sorter := NewSpineElementSorter()

	// Add elements out of chronological order
	element3, _ := NewValidatedSpineElement(SpineElementTitle, Time("480480/24000s"), Duration("120120/24000s"), Lane(1), "Title1")
	element1, _ := NewValidatedSpineElement(SpineElementAssetClip, Time("0s"), Duration("240240/24000s"), Lane(1), "Clip1")
	element2, _ := NewValidatedSpineElement(SpineElementVideo, Time("240240/24000s"), Duration("240240/24000s"), Lane(2), "Video1")

	// Add non-overlapping elements to different lanes
	sorter.AddElement(element3)
	sorter.AddElement(element1)
	sorter.AddElement(element2)

	// Sort and validate
	sortedElements, err := sorter.SortAndValidate()
	if err != nil {
		t.Errorf("unexpected error sorting elements: %v", err)
		return
	}

	// Verify order
	if len(sortedElements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(sortedElements))
	}

	// Should be sorted by offset time
	expectedOrder := []*ValidatedSpineElement{element1, element2, element3}
	for i, expected := range expectedOrder {
		if sortedElements[i] != expected {
			t.Errorf("element %d: expected %s, got %s", i, expected.Name, sortedElements[i].Name)
		}
	}
}

func TestSpineElementSorter_GetElementsByLane(t *testing.T) {
	sorter := NewSpineElementSorter()

	// Add elements to different lanes
	element1, _ := NewValidatedSpineElement(SpineElementAssetClip, Time("0s"), Duration("240240/24000s"), Lane(1), "Clip1")
	element2, _ := NewValidatedSpineElement(SpineElementVideo, Time("240240/24000s"), Duration("240240/24000s"), Lane(2), "Video1")
	element3, _ := NewValidatedSpineElement(SpineElementTitle, Time("480480/24000s"), Duration("120120/24000s"), Lane(1), "Title1")

	sorter.AddElement(element1)
	sorter.AddElement(element2)
	sorter.AddElement(element3)

	// Get elements in lane 1
	lane1Elements := sorter.GetElementsByLane(Lane(1))
	if len(lane1Elements) != 2 {
		t.Errorf("expected 2 elements in lane 1, got %d", len(lane1Elements))
	}

	// Get elements in lane 2
	lane2Elements := sorter.GetElementsByLane(Lane(2))
	if len(lane2Elements) != 1 {
		t.Errorf("expected 1 element in lane 2, got %d", len(lane2Elements))
	}

	// Get elements in unused lane
	lane3Elements := sorter.GetElementsByLane(Lane(3))
	if len(lane3Elements) != 0 {
		t.Errorf("expected 0 elements in lane 3, got %d", len(lane3Elements))
	}
}

func TestSpineElementSorter_GetElementsByType(t *testing.T) {
	sorter := NewSpineElementSorter()

	// Add elements of different types
	element1, _ := NewValidatedSpineElement(SpineElementAssetClip, Time("0s"), Duration("240240/24000s"), Lane(1), "Clip1")
	element2, _ := NewValidatedSpineElement(SpineElementVideo, Time("240240/24000s"), Duration("240240/24000s"), Lane(2), "Video1")
	element3, _ := NewValidatedSpineElement(SpineElementAssetClip, Time("480480/24000s"), Duration("120120/24000s"), Lane(1), "Clip2")

	sorter.AddElement(element1)
	sorter.AddElement(element2)
	sorter.AddElement(element3)

	// Get asset clips
	assetClips := sorter.GetElementsByType(SpineElementAssetClip)
	if len(assetClips) != 2 {
		t.Errorf("expected 2 asset clips, got %d", len(assetClips))
	}

	// Get videos
	videos := sorter.GetElementsByType(SpineElementVideo)
	if len(videos) != 1 {
		t.Errorf("expected 1 video, got %d", len(videos))
	}

	// Get titles (none added)
	titles := sorter.GetElementsByType(SpineElementTitle)
	if len(titles) != 0 {
		t.Errorf("expected 0 titles, got %d", len(titles))
	}
}

func TestSpineElementSorter_GetElementsInTimeRange(t *testing.T) {
	sorter := NewSpineElementSorter()

	// Add elements at different times
	element1, _ := NewValidatedSpineElement(SpineElementAssetClip, Time("0s"), Duration("240240/24000s"), Lane(1), "Clip1")        // 0-10s
	element2, _ := NewValidatedSpineElement(SpineElementVideo, Time("240240/24000s"), Duration("240240/24000s"), Lane(2), "Video1") // 10-20s
	element3, _ := NewValidatedSpineElement(SpineElementTitle, Time("480480/24000s"), Duration("240240/24000s"), Lane(1), "Title1") // 20-30s

	sorter.AddElement(element1)
	sorter.AddElement(element2)
	sorter.AddElement(element3)

	// Query for elements in first 15 seconds (0-15s)
	elementsInRange, err := sorter.GetElementsInTimeRange(Time("0s"), Time("360360/24000s")) // 15 seconds
	if err != nil {
		t.Errorf("unexpected error getting elements in range: %v", err)
		return
	}

	// Should find element1 (0-10s) and element2 (10-20s)
	if len(elementsInRange) != 2 {
		t.Errorf("expected 2 elements in range, got %d", len(elementsInRange))
	}

	// Verify we got the right elements
	foundNames := make(map[string]bool)
	for _, elem := range elementsInRange {
		foundNames[elem.Name] = true
	}

	if !foundNames["Clip1"] || !foundNames["Video1"] {
		t.Errorf("wrong elements found in range: %v", foundNames)
	}
}

func TestSpineBuilder(t *testing.T) {
	registry := NewReferenceRegistry()
	builder, err := NewSpineBuilder(Duration("720720/24000s"), registry) // 30 seconds
	if err != nil {
		t.Fatalf("failed to create spine builder: %v", err)
	}

	// Add elements
	err = builder.AddAssetClip("r1", "Clip1", Time("0s"), Duration("240240/24000s"), Lane(1))
	if err != nil {
		t.Errorf("unexpected error adding asset clip: %v", err)
	}

	err = builder.AddVideo("r2", "Video1", Time("240240/24000s"), Duration("240240/24000s"), Lane(2))
	if err != nil {
		t.Errorf("unexpected error adding video: %v", err)
	}

	err = builder.AddTitle("r3", "Title1", Time("480480/24000s"), Duration("240240/24000s"), Lane(3))
	if err != nil {
		t.Errorf("unexpected error adding title: %v", err)
	}

	// Build spine
	spine, err := builder.Build()
	if err != nil {
		t.Errorf("unexpected error building spine: %v", err)
		return
	}

	// Verify spine structure
	if len(spine.AssetClips) != 1 {
		t.Errorf("expected 1 asset clip, got %d", len(spine.AssetClips))
	}

	if len(spine.Videos) != 1 {
		t.Errorf("expected 1 video, got %d", len(spine.Videos))
	}

	if len(spine.Titles) != 1 {
		t.Errorf("expected 1 title, got %d", len(spine.Titles))
	}

	// Verify chronological order (elements should be sorted by offset in XML)
	if spine.AssetClips[0].Name != "Clip1" {
		t.Errorf("expected first asset clip to be Clip1, got %s", spine.AssetClips[0].Name)
	}

	if spine.Videos[0].Name != "Video1" {
		t.Errorf("expected first video to be Video1, got %s", spine.Videos[0].Name)
	}

	if spine.Titles[0].Name != "Title1" {
		t.Errorf("expected first title to be Title1, got %s", spine.Titles[0].Name)
	}
}

func TestSpineBuilder_OverlapValidation(t *testing.T) {
	registry := NewReferenceRegistry()
	builder, err := NewSpineBuilder(Duration("720720/24000s"), registry) // 30 seconds
	if err != nil {
		t.Fatalf("failed to create spine builder: %v", err)
	}

	// Add first element
	err = builder.AddAssetClip("r1", "Clip1", Time("0s"), Duration("240240/24000s"), Lane(1))
	if err != nil {
		t.Errorf("unexpected error adding first clip: %v", err)
	}

	// Try to add overlapping element in same lane (should fail)
	err = builder.AddAssetClip("r2", "Clip2", Time("120120/24000s"), Duration("240240/24000s"), Lane(1))
	if err == nil {
		t.Errorf("expected error adding overlapping clip")
	}

	// Add overlapping element in different lane (should succeed)
	err = builder.AddAssetClip("r2", "Clip2", Time("120120/24000s"), Duration("240240/24000s"), Lane(2))
	if err != nil {
		t.Errorf("unexpected error adding clip in different lane: %v", err)
	}

	// Allow overlaps and try again in same lane
	builder.SetAllowOverlaps(true)
	err = builder.AddAssetClip("r3", "Clip3", Time("60060/24000s"), Duration("240240/24000s"), Lane(1))
	if err != nil {
		t.Errorf("unexpected error adding overlapping clip with overlaps allowed: %v", err)
	}
}

func TestSpineBuilder_GetStatistics(t *testing.T) {
	registry := NewReferenceRegistry()
	builder, err := NewSpineBuilder(Duration("720720/24000s"), registry) // 30 seconds
	if err != nil {
		t.Fatalf("failed to create spine builder: %v", err)
	}

	// Add elements
	builder.AddAssetClip("r1", "Clip1", Time("0s"), Duration("240240/24000s"), Lane(1))
	builder.AddAssetClip("r2", "Clip2", Time("240240/24000s"), Duration("240240/24000s"), Lane(2))
	builder.AddVideo("r3", "Video1", Time("480480/24000s"), Duration("240240/24000s"), Lane(1))
	builder.AddTitle("r4", "Title1", Time("0s"), Duration("720720/24000s"), Lane(3))

	stats := builder.GetStatistics()

	// Check total elements
	if stats.TotalElements != 4 {
		t.Errorf("expected 4 total elements, got %d", stats.TotalElements)
	}

	// Check elements by type
	if stats.ElementsByType["asset-clip"] != 2 {
		t.Errorf("expected 2 asset clips, got %d", stats.ElementsByType["asset-clip"])
	}

	if stats.ElementsByType["video"] != 1 {
		t.Errorf("expected 1 video, got %d", stats.ElementsByType["video"])
	}

	if stats.ElementsByType["title"] != 1 {
		t.Errorf("expected 1 title, got %d", stats.ElementsByType["title"])
	}

	// Check used lanes
	expectedLanes := []int{1, 2, 3}
	if len(stats.UsedLanes) != len(expectedLanes) {
		t.Errorf("expected %d used lanes, got %d", len(expectedLanes), len(stats.UsedLanes))
	}

	for i, expected := range expectedLanes {
		if i >= len(stats.UsedLanes) || stats.UsedLanes[i] != expected {
			t.Errorf("expected lane %d at position %d, got %v", expected, i, stats.UsedLanes)
		}
	}
}