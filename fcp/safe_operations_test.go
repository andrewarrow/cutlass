// Package safe_operations_test tests the Step 17 implementation:
// Safe High-Level Operations with built-in validation.
package fcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSafeFCPXMLOperations(t *testing.T) {
	tests := []struct {
		name            string
		projectName     string
		durationSeconds float64
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid creation",
			projectName:     "Test Project",
			durationSeconds: 60.0,
			expectError:     false,
		},
		{
			name:            "Empty project name",
			projectName:     "",
			durationSeconds: 60.0,
			expectError:     true,
			errorContains:   "project name cannot be empty",
		},
		{
			name:            "Zero duration",
			projectName:     "Test Project",
			durationSeconds: 0.0,
			expectError:     true,
			errorContains:   "duration must be positive",
		},
		{
			name:            "Negative duration",
			projectName:     "Test Project",
			durationSeconds: -10.0,
			expectError:     true,
			errorContains:   "duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := NewSafeFCPXMLOperations(tt.projectName, tt.durationSeconds)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if ops == nil {
				t.Errorf("expected ops to be non-nil")
			}
		})
	}
}

func TestAddBackgroundVideo(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 120.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	tests := []struct {
		name            string
		videoPath       string
		durationSeconds float64
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid video",
			videoPath:       "/tmp/path/to/video.mp4",
			durationSeconds: 60.0,
			expectError:     false,
		},
		{
			name:            "Empty path",
			videoPath:       "",
			durationSeconds: 60.0,
			expectError:     true,
			errorContains:   "path cannot be empty",
		},
		{
			name:            "Unsupported format",
			videoPath:       "/path/to/file.txt",
			durationSeconds: 60.0,
			expectError:     true,
			errorContains:   "unsupported file format",
		},
		{
			name:            "Zero duration",
			videoPath:       "/tmp/path/to/video.mp4",
			durationSeconds: 0.0,
			expectError:     true,
			errorContains:   "duration must be positive",
		},
		{
			name:            "Negative duration",
			videoPath:       "/tmp/path/to/video.mp4",
			durationSeconds: -5.0,
			expectError:     true,
			errorContains:   "duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ops.AddBackgroundVideo(tt.videoPath, tt.durationSeconds)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddImageOverlay(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 120.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	tests := []struct {
		name            string
		imagePath       string
		startSeconds    float64
		durationSeconds float64
		lane            int
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid image overlay",
			imagePath:       "/path/to/image.png",
			startSeconds:    10.0,
			durationSeconds: 5.0,
			lane:            0,
			expectError:     false,
		},
		{
			name:            "Negative start time",
			imagePath:       "/path/to/image.png",
			startSeconds:    -5.0,
			durationSeconds: 5.0,
			lane:            1,
			expectError:     true,
			errorContains:   "start time cannot be negative",
		},
		{
			name:            "Zero duration",
			imagePath:       "/path/to/image.png",
			startSeconds:    10.0,
			durationSeconds: 0.0,
			lane:            1,
			expectError:     true,
			errorContains:   "duration must be positive",
		},
		{
			name:            "Invalid lane",
			imagePath:       "/path/to/image.png",
			startSeconds:    10.0,
			durationSeconds: 5.0,
			lane:            15,
			expectError:     true,
			errorContains:   "lane out of range",
		},
		{
			name:            "Unsupported image format",
			imagePath:       "/path/to/file.pdf",
			startSeconds:    10.0,
			durationSeconds: 5.0,
			lane:            1,
			expectError:     true,
			errorContains:   "unsupported file format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ops.AddImageOverlay(tt.imagePath, tt.startSeconds, tt.durationSeconds, tt.lane)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddTitleCard(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 120.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	tests := []struct {
		name            string
		text            string
		startSeconds    float64
		durationSeconds float64
		lane            int
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid title card",
			text:            "Hello World",
			startSeconds:    5.0,
			durationSeconds: 3.0,
			lane:            0,
			expectError:     false,
		},
		{
			name:            "Empty text",
			text:            "",
			startSeconds:    5.0,
			durationSeconds: 3.0,
			lane:            2,
			expectError:     true,
			errorContains:   "title text cannot be empty",
		},
		{
			name:            "Negative start time",
			text:            "Hello World",
			startSeconds:    -2.0,
			durationSeconds: 3.0,
			lane:            2,
			expectError:     true,
			errorContains:   "start time cannot be negative",
		},
		{
			name:            "Zero duration",
			text:            "Hello World",
			startSeconds:    5.0,
			durationSeconds: 0.0,
			lane:            2,
			expectError:     true,
			errorContains:   "duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ops.AddTitleCard(tt.text, tt.startSeconds, tt.durationSeconds, tt.lane)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddPanAnimation(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 120.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	tests := []struct {
		name            string
		startSeconds    float64
		durationSeconds float64
		fromX, fromY    float64
		toX, toY        float64
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid pan animation",
			startSeconds:    10.0,
			durationSeconds: 2.0,
			fromX:           0, fromY: 0,
			toX: 100, toY: 50,
			expectError: false,
		},
		{
			name:            "Negative start time",
			startSeconds:    -5.0,
			durationSeconds: 2.0,
			fromX:           0, fromY: 0,
			toX: 100, toY: 50,
			expectError:   true,
			errorContains: "start time cannot be negative",
		},
		{
			name:            "Zero duration",
			startSeconds:    10.0,
			durationSeconds: 0.0,
			fromX:           0, fromY: 0,
			toX: 100, toY: 50,
			expectError:   true,
			errorContains: "duration must be positive",
		},
		{
			name:            "Extreme position values",
			startSeconds:    10.0,
			durationSeconds: 2.0,
			fromX:           10000, fromY: 0,
			toX: 100, toY: 50,
			expectError:   true,
			errorContains: "position out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ops.AddPanAnimation(tt.startSeconds, tt.durationSeconds, tt.fromX, tt.fromY, tt.toX, tt.toY)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddZoomAnimation(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 120.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	tests := []struct {
		name            string
		startSeconds    float64
		durationSeconds float64
		fromScale       float64
		toScale         float64
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Valid zoom animation",
			startSeconds:    5.0,
			durationSeconds: 3.0,
			fromScale:       1.0,
			toScale:         1.5,
			expectError:     false,
		},
		{
			name:            "Zero scale",
			startSeconds:    5.0,
			durationSeconds: 3.0,
			fromScale:       0.0,
			toScale:         1.5,
			expectError:     true,
			errorContains:   "scale must be positive",
		},
		{
			name:            "Negative scale",
			startSeconds:    5.0,
			durationSeconds: 3.0,
			fromScale:       -1.0,
			toScale:         1.5,
			expectError:     true,
			errorContains:   "scale must be positive",
		},
		{
			name:            "Too large scale",
			startSeconds:    5.0,
			durationSeconds: 3.0,
			fromScale:       1.0,
			toScale:         200.0,
			expectError:     true,
			errorContains:   "scale too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ops.AddZoomAnimation(tt.startSeconds, tt.durationSeconds, tt.fromScale, tt.toScale)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddMultiLayerComposite(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 120.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	tests := []struct {
		name        string
		mediaFiles  []MediaLayerSpec
		expectError bool
		errorContains string
	}{
		{
			name: "Overlapping elements on same lane",
			mediaFiles: []MediaLayerSpec{
				{
					Path:            "/tmp/path/to/video.mp4",
					Type:            "video",
					StartSeconds:    0.0,
					DurationSeconds: 10.0,
					Lane:            0,
				},
				{
					Path:            "/path/to/overlay.png",
					Type:            "image",
					StartSeconds:    2.0,
					DurationSeconds: 5.0,
					Lane:            0,
				},
			},
			expectError: true,
			errorContains: "overlap",
		},
		{
			name:        "Empty media files",
			mediaFiles:  []MediaLayerSpec{},
			expectError: true,
			errorContains: "at least one media file is required",
		},
		{
			name: "Too many layers",
			mediaFiles: make([]MediaLayerSpec, 15), // Create 15 layers
			expectError: true,
			errorContains: "too many layers",
		},
		{
			name: "Invalid media file",
			mediaFiles: []MediaLayerSpec{
				{
					Path:            "", // Empty path
					Type:            "video",
					StartSeconds:    0.0,
					DurationSeconds: 10.0,
					Lane:            0,
				},
			},
			expectError: true,
			errorContains: "path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill in valid dummy data for "too many layers" test
			if len(tt.mediaFiles) > 10 {
				for i := range tt.mediaFiles {
					tt.mediaFiles[i] = MediaLayerSpec{
						Path:            "/tmp/path/to/video.mp4",
						Type:            "video",
						StartSeconds:    0.0,
						DurationSeconds: 10.0,
						Lane:            i,
					}
				}
			}

			err := ops.AddMultiLayerComposite(tt.mediaFiles)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("CreateSimpleVideo", func(t *testing.T) {
		ops, err := CreateSimpleVideo("Simple Project", "/tmp/path/to/video.mp4", 60.0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if ops == nil {
			t.Errorf("expected ops to be non-nil")
		}

		stats := ops.GetDocumentStatistics()
		if stats.ProjectName != "Simple Project" {
			t.Errorf("expected project name 'Simple Project', got: %s", stats.ProjectName)
		}
	})

	t.Run("CreateImageSlideshow", func(t *testing.T) {
		imagePaths := []string{
			"/path/to/image1.jpg",
			"/path/to/image2.png",
			"/path/to/image3.gif",
		}

		ops, err := CreateImageSlideshow("Slideshow Project", imagePaths, 3.0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if ops == nil {
			t.Errorf("expected ops to be non-nil")
		}

		stats := ops.GetDocumentStatistics()
		if stats.ProjectName != "Slideshow Project" {
			t.Errorf("expected project name 'Slideshow Project', got: %s", stats.ProjectName)
		}

		expectedDuration := "216216/24000s" // 3 images * 3 seconds each = 9 seconds
		if stats.TotalDuration != expectedDuration {
			t.Errorf("expected duration %s, got: %s", expectedDuration, stats.TotalDuration)
		}
	})

	t.Run("CreateTitleSequence", func(t *testing.T) {
		titles := []string{
			"Title One",
			"Title Two", 
			"Title Three",
		}

		ops, err := CreateTitleSequence("Title Project", titles, 2.0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if ops == nil {
			t.Errorf("expected ops to be non-nil")
		}

		stats := ops.GetDocumentStatistics()
		if stats.ProjectName != "Title Project" {
			t.Errorf("expected project name 'Title Project', got: %s", stats.ProjectName)
		}
	})
}

func TestValidation(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 120.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	t.Run("validateColorString", func(t *testing.T) {
		tests := []struct {
			name        string
			colorStr    string
			expectError bool
		}{
			{"Valid RGBA", "1.0 0.5 0.0 1.0", false},
			{"Valid RGBA with zeros", "0.0 0.0 0.0 0.0", false},
			{"Too few components", "1.0 0.5 0.0", true},
			{"Too many components", "1.0 0.5 0.0 1.0 0.5", true},
			{"Invalid component", "1.0 abc 0.0 1.0", true},
			{"Out of range high", "2.0 0.5 0.0 1.0", true},
			{"Out of range low", "-0.5 0.5 0.0 1.0", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ops.validateColorString(tt.colorStr)
				if tt.expectError && err == nil {
					t.Errorf("expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
		}
	})

	t.Run("validateLane", func(t *testing.T) {
		tests := []struct {
			name        string
			lane        int
			expectError bool
		}{
			{"Valid positive lane", 5, false},
			{"Valid negative lane", -3, false},
			{"Valid zero lane", 0, false},
			{"Too high lane", 15, true},
			{"Too low lane", -15, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ops.validateLane(tt.lane)
				if tt.expectError && err == nil {
					t.Errorf("expected error but got none")
				}
				if !tt.expectError && err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
		}
	})
}

func TestGenerateAndValidate(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 60.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	// Add some basic content
	if err := ops.AddBackgroundVideo("/tmp/path/to/video.mp4", 30.0); err != nil {
		t.Fatalf("failed to add background video: %v", err)
	}

	if err := ops.AddTitleCard("Test Title", 35.0, 3.0, 0); err != nil {
		t.Fatalf("failed to add title card: %v", err)
	}

	// Generate FCPXML
	data, err := ops.GenerateAndValidate()
	if err != nil {
		t.Errorf("failed to generate FCPXML: %v", err)
		return
	}

	if len(data) == 0 {
		t.Errorf("expected non-empty FCPXML data")
	}

	// Check that it contains expected elements
	xmlStr := string(data)
	if !strings.Contains(xmlStr, "fcpxml") {
		t.Errorf("expected FCPXML to contain 'fcpxml' element")
	}

	if !strings.Contains(xmlStr, "Test Project") {
		t.Errorf("expected FCPXML to contain project name")
	}
}

func TestSaveToFile(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 60.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	// Add some content
	if err := ops.AddBackgroundVideo("/tmp/path/to/video.mp4", 30.0); err != nil {
		t.Fatalf("failed to add background video: %v", err)
	}

	// Create a temporary file
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test")

	// Save to file
	if err := ops.SaveToFile(filename); err != nil {
		t.Errorf("failed to save file: %v", err)
		return
	}

	// Check that file was created with .fcpxml extension
	expectedPath := filename + ".fcpxml"
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file to be created at %s", expectedPath)
	}

	// Check file contents
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Errorf("failed to read saved file: %v", err)
		return
	}

	if len(data) == 0 {
		t.Errorf("expected non-empty file")
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, "fcpxml") {
		t.Errorf("expected saved file to contain 'fcpxml' element")
	}
}

func TestTitleCardOptions(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 60.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	// Test title card with options
	err = ops.AddTitleCard("Custom Title", 10.0, 5.0, 0,
		WithTitleFont("Arial"),
		WithTitleSize("72"),
		WithTitleColor("1.0 0.0 0.0 1.0"), // Red
		WithTitleAlignment("left"),
		WithTitleBold(true),
		WithTitleItalic(true),
	)

	if err != nil {
		t.Errorf("failed to add title card with options: %v", err)
	}
}

func TestDocumentSettings(t *testing.T) {
	ops, err := NewSafeFCPXMLOperations("Test Project", 60.0)
	if err != nil {
		t.Fatalf("failed to create operations: %v", err)
	}

	// Test valid settings
	settings := DocumentSettings{
		MaxLanes:      8,
		AllowOverlaps: true,
		AllowLaneGaps: true,
	}

	if err := ops.SetDocumentSettings(settings); err != nil {
		t.Errorf("failed to set valid document settings: %v", err)
	}

	// Test invalid settings
	invalidSettings := DocumentSettings{
		MaxLanes:      25, // Too high
		AllowOverlaps: false,
		AllowLaneGaps: false,
	}

	if err := ops.SetDocumentSettings(invalidSettings); err == nil {
		t.Errorf("expected error for invalid settings but got none")
	}
}