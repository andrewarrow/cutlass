package fcp

import (
	"testing"
)

func TestTimelineValidation_NewTimeRange(t *testing.T) {
	tests := []struct {
		name        string
		start       Time
		duration    Duration
		expectError bool
	}{
		{
			name:     "valid range",
			start:    Time("0s"),
			duration: Duration("240240/24000s"),
		},
		{
			name:     "valid range with offset",
			start:    Time("120120/24000s"),
			duration: Duration("240240/24000s"),
		},
		{
			name:        "invalid start time",
			start:       Time("invalid"),
			duration:    Duration("240240/24000s"),
			expectError: true,
		},
		{
			name:        "invalid duration",
			start:       Time("0s"),
			duration:    Duration("invalid"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTimeRange(tt.start, tt.duration)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTimelineValidator_AddElement(t *testing.T) {
	validator, err := NewTimelineValidator(Duration("600600/24000s")) // 25 seconds
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	tests := []struct {
		name        string
		id          string
		offset      Time
		duration    Duration
		lane        Lane
		elementType string
		expectError bool
	}{
		{
			name:        "valid element",
			id:          "test1",
			offset:      Time("0s"),
			duration:    Duration("240240/24000s"), // 10 seconds
			lane:        Lane(1),
			elementType: "video",
		},
		{
			name:        "element extends beyond timeline",
			id:          "test2",
			offset:      Time("480480/24000s"), // 20 seconds
			duration:    Duration("240240/24000s"), // 10 seconds (would end at 30s)
			lane:        Lane(2),
			elementType: "video",
			expectError: true,
		},
		{
			name:        "invalid lane",
			id:          "test3",
			offset:      Time("0s"),
			duration:    Duration("120120/24000s"),
			lane:        Lane(15), // Exceeds max lanes
			elementType: "video",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.AddElement(tt.id, tt.offset, tt.duration, tt.lane, tt.elementType)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTimelineValidator_ValidateLaneStructure(t *testing.T) {
	tests := []struct {
		name        string
		elements    []struct {
			id   string
			lane Lane
		}
		allowGaps   bool
		expectError bool
	}{
		{
			name: "continuous lanes",
			elements: []struct {
				id   string
				lane Lane
			}{
				{"test1", Lane(1)},
				{"test2", Lane(2)},
				{"test3", Lane(3)},
			},
			allowGaps: false,
		},
		{
			name: "lane gap",
			elements: []struct {
				id   string
				lane Lane
			}{
				{"test1", Lane(1)},
				{"test3", Lane(3)}, // Missing lane 2
			},
			allowGaps:   false,
			expectError: true,
		},
		{
			name: "lane gap allowed",
			elements: []struct {
				id   string
				lane Lane
			}{
				{"test1", Lane(1)},
				{"test3", Lane(3)}, // Missing lane 2
			},
			allowGaps: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewTimelineValidator(Duration("600600/24000s"))
			if err != nil {
				t.Fatalf("failed to create validator: %v", err)
			}

			validator.SetAllowGaps(tt.allowGaps)

			// Add elements
			for _, elem := range tt.elements {
				err := validator.AddElement(elem.id, Time("0s"), Duration("240240/24000s"), elem.lane, "video")
				if err != nil {
					t.Fatalf("failed to add element: %v", err)
				}
			}

			err = validator.ValidateLaneStructure()
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTimelineValidator_ValidateOverlaps(t *testing.T) {
	validator, err := NewTimelineValidator(Duration("600600/24000s"))
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Add non-overlapping elements
	err = validator.AddElement("test1", Time("0s"), Duration("120120/24000s"), Lane(1), "video")
	if err != nil {
		t.Fatalf("failed to add element 1: %v", err)
	}

	err = validator.AddElement("test2", Time("120120/24000s"), Duration("120120/24000s"), Lane(1), "video")
	if err != nil {
		t.Fatalf("failed to add element 2: %v", err)
	}

	// Should pass - no overlaps
	err = validator.ValidateOverlaps()
	if err != nil {
		t.Errorf("unexpected overlap error: %v", err)
	}

	// Add overlapping element
	err = validator.AddElement("test3", Time("60060/24000s"), Duration("120120/24000s"), Lane(1), "video")
	if err != nil {
		t.Fatalf("failed to add element 3: %v", err)
	}

	// Should fail - overlap detected
	err = validator.ValidateOverlaps()
	if err == nil {
		t.Errorf("expected overlap error but got none")
	}
}

func TestTimelineValidator_GetElementsInTimeRange(t *testing.T) {
	validator, err := NewTimelineValidator(Duration("600600/24000s"))
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Add elements at different times
	elements := []struct {
		id     string
		offset Time
		lane   Lane
	}{
		{"test1", Time("0s"), Lane(1)},
		{"test2", Time("120120/24000s"), Lane(1)},
		{"test3", Time("240240/24000s"), Lane(2)},
		{"test4", Time("360360/24000s"), Lane(2)},
	}

	for _, elem := range elements {
		err := validator.AddElement(elem.id, elem.offset, Duration("60060/24000s"), elem.lane, "video")
		if err != nil {
			t.Fatalf("failed to add element %s: %v", elem.id, err)
		}
	}

	// Query for elements in first 10 seconds
	foundElements, err := validator.GetElementsInTimeRange(Time("0s"), Duration("240240/24000s"))
	if err != nil {
		t.Fatalf("failed to get elements in range: %v", err)
	}

	// Should find test1 and test2
	if len(foundElements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(foundElements))
	}

	// Verify we got the right elements
	foundIDs := make(map[string]bool)
	for _, elem := range foundElements {
		foundIDs[elem.ID] = true
	}

	if !foundIDs["test1"] || !foundIDs["test2"] {
		t.Errorf("wrong elements found: %v", foundIDs)
	}
}

func TestTimelineValidator_GetTimelineStatistics(t *testing.T) {
	validator, err := NewTimelineValidator(Duration("600600/24000s"))
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Add elements
	err = validator.AddElement("video1", Time("0s"), Duration("240240/24000s"), Lane(1), "video")
	if err != nil {
		t.Fatalf("failed to add video1: %v", err)
	}

	err = validator.AddElement("video2", Time("120120/24000s"), Duration("120120/24000s"), Lane(2), "video")
	if err != nil {
		t.Fatalf("failed to add video2: %v", err)
	}

	err = validator.AddElement("title1", Time("60060/24000s"), Duration("60060/24000s"), Lane(3), "title")
	if err != nil {
		t.Fatalf("failed to add title1: %v", err)
	}

	stats := validator.GetTimelineStatistics()

	// Check total elements
	if stats.TotalElements != 3 {
		t.Errorf("expected 3 total elements, got %d", stats.TotalElements)
	}

	// Check used lanes
	expectedLanes := []int{1, 2, 3}
	if len(stats.UsedLanes) != len(expectedLanes) {
		t.Errorf("expected %d lanes, got %d", len(expectedLanes), len(stats.UsedLanes))
	}

	// Check elements by type
	if stats.ElementsByType["video"] != 2 {
		t.Errorf("expected 2 video elements, got %d", stats.ElementsByType["video"])
	}

	if stats.ElementsByType["title"] != 1 {
		t.Errorf("expected 1 title element, got %d", stats.ElementsByType["title"])
	}
}

func TestTimelineValidator_ValidateComplete(t *testing.T) {
	validator, err := NewTimelineValidator(Duration("600600/24000s"))
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// Add valid, non-overlapping elements
	err = validator.AddElement("test1", Time("0s"), Duration("120120/24000s"), Lane(1), "video")
	if err != nil {
		t.Fatalf("failed to add element 1: %v", err)
	}

	err = validator.AddElement("test2", Time("0s"), Duration("120120/24000s"), Lane(2), "video")
	if err != nil {
		t.Fatalf("failed to add element 2: %v", err)
	}

	// Should pass complete validation
	err = validator.ValidateComplete()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}