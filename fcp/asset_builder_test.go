package fcp

import (
	"strings"
	"testing"
)

func TestAssetBuilder(t *testing.T) {
	registry := NewReferenceRegistry()

	tests := []struct {
		name        string
		mediaType   string
		id          ID
		filePath    string
		assetName   string
		duration    Duration
		expectError bool
		errorType   string
	}{
		{
			name:        "Valid image asset should be created successfully",
			mediaType:   "image",
			id:          ID("r1"),
			filePath:    "/test/image.png",
			assetName:   "test-image",
			duration:    Duration("0s"),
			expectError: false,
		},
		{
			name:        "Valid video asset should be created successfully",
			mediaType:   "video",
			id:          ID("r2"),
			filePath:    "/test/video.mp4",
			assetName:   "test-video",
			duration:    Duration("240240/24000s"),
			expectError: false,
		},
		{
			name:        "Valid audio asset should be created successfully",
			mediaType:   "audio",
			id:          ID("r3"),
			filePath:    "/test/audio.wav",
			assetName:   "test-audio",
			duration:    Duration("120120/24000s"),
			expectError: false,
		},
		{
			name:        "Image with non-zero duration should fail",
			mediaType:   "image",
			id:          ID("r4"),
			filePath:    "/test/image.png",
			assetName:   "test-image",
			duration:    Duration("240240/24000s"),
			expectError: true,
			errorType:   "image assets must have duration='0s'",
		},
		{
			name:        "Video with zero duration should fail",
			mediaType:   "video",
			id:          ID("r5"),
			filePath:    "/test/video.mp4",
			assetName:   "test-video",
			duration:    Duration("0s"),
			expectError: true,
			errorType:   "video assets cannot have duration='0s'",
		},
		{
			name:        "Audio with zero duration should fail",
			mediaType:   "audio",
			id:          ID("r6"),
			filePath:    "/test/audio.wav",
			assetName:   "test-audio",
			duration:    Duration("0s"),
			expectError: true,
			errorType:   "audio assets cannot have duration='0s'",
		},
		{
			name:        "Invalid ID should fail",
			mediaType:   "video",
			id:          ID("invalid"),
			filePath:    "/test/video.mp4",
			assetName:   "test-video",
			duration:    Duration("240240/24000s"),
			expectError: true,
			errorType:   "invalid asset ID",
		},
		{
			name:        "Invalid duration should fail",
			mediaType:   "video",
			id:          ID("r7"),
			filePath:    "/test/video.mp4",
			assetName:   "test-video",
			duration:    Duration("invalid"),
			expectError: true,
			errorType:   "invalid duration",
		},
		{
			name:        "Unsupported media type should fail",
			mediaType:   "unknown",
			id:          ID("r8"),
			filePath:    "/test/file.unknown",
			assetName:   "test-file",
			duration:    Duration("240240/24000s"),
			expectError: true,
			errorType:   "unsupported media type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewAssetBuilder(registry, tt.mediaType)
			asset, format, err := builder.CreateAsset(tt.id, tt.filePath, tt.assetName, tt.duration)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorType) {
					t.Errorf("Expected error type '%s' but got: %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if asset == nil {
					t.Error("Expected asset but got nil")
					return
				}

				// Validate asset properties based on media type
				switch tt.mediaType {
				case "image":
					if asset.Duration != "0s" {
						t.Errorf("Image asset duration should be '0s', got: %s", asset.Duration)
					}
					if asset.HasVideo != "1" {
						t.Errorf("Image asset should have hasVideo='1', got: %s", asset.HasVideo)
					}
					if asset.HasAudio != "" {
						t.Errorf("Image asset should not have hasAudio, got: %s", asset.HasAudio)
					}
					if format == nil {
						t.Error("Image asset should have a format")
					} else if format.FrameDuration != "" {
						t.Errorf("Image format should not have frameDuration, got: %s", format.FrameDuration)
					}

				case "video":
					if asset.Duration == "0s" {
						t.Error("Video asset should not have duration='0s'")
					}
					if asset.HasVideo != "1" {
						t.Errorf("Video asset should have hasVideo='1', got: %s", asset.HasVideo)
					}
					if format == nil {
						t.Error("Video asset should have a format")
					} else if format.FrameDuration == "" {
						t.Error("Video format should have frameDuration")
					}

				case "audio":
					if asset.Duration == "0s" {
						t.Error("Audio asset should not have duration='0s'")
					}
					if asset.HasAudio != "1" {
						t.Errorf("Audio asset should have hasAudio='1', got: %s", asset.HasAudio)
					}
					if asset.HasVideo != "" {
						t.Errorf("Audio asset should not have hasVideo, got: %s", asset.HasVideo)
					}
					if format != nil {
						t.Error("Audio asset should not have a format")
					}
				}

				// Common validations
				if asset.ID != string(tt.id) {
					t.Errorf("Asset ID should be %s, got: %s", tt.id, asset.ID)
				}
				if asset.Name != tt.assetName {
					t.Errorf("Asset name should be %s, got: %s", tt.assetName, asset.Name)
				}
				if asset.UID == "" {
					t.Error("Asset UID should not be empty")
				}
				if !strings.HasPrefix(asset.MediaRep.Src, "file://") {
					t.Errorf("Asset MediaRep src should start with 'file://', got: %s", asset.MediaRep.Src)
				}
			}
		})
	}
}

func TestAssetBuilderSpecificMethods(t *testing.T) {
	registry := NewReferenceRegistry()

	t.Run("CreateImageAsset", func(t *testing.T) {
		builder := NewAssetBuilder(registry, "image")
		asset, format, err := builder.CreateImageAsset(ID("r1"), "/test/image.png", "test-image", Duration("240240/24000s"))

		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
			return
		}

		if asset.Duration != "0s" {
			t.Errorf("Image asset duration should be '0s', got: %s", asset.Duration)
		}

		if format == nil {
			t.Error("Image should have a format")
		}
	})

	t.Run("CreateVideoAsset", func(t *testing.T) {
		builder := NewAssetBuilder(registry, "video")
		asset, format, err := builder.CreateVideoAsset(ID("r2"), "/test/video.mp4", "test-video", Duration("240240/24000s"))

		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
			return
		}

		if asset.Duration == "0s" {
			t.Error("Video asset should not have duration='0s'")
		}

		if format == nil {
			t.Error("Video should have a format")
		} else if format.FrameDuration == "" {
			t.Error("Video format should have frameDuration")
		}
	})

	t.Run("CreateAudioAsset", func(t *testing.T) {
		builder := NewAssetBuilder(registry, "audio")
		asset, err := builder.CreateAudioAsset(ID("r3"), "/test/audio.wav", "test-audio", Duration("120120/24000s"))

		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
			return
		}

		if asset.Duration == "0s" {
			t.Error("Audio asset should not have duration='0s'")
		}

		if asset.HasAudio != "1" {
			t.Errorf("Audio asset should have hasAudio='1', got: %s", asset.HasAudio)
		}
	})

	t.Run("CreateVideoAssetWithDetection", func(t *testing.T) {
		builder := NewAssetBuilder(registry, "video")
		asset, format, err := builder.CreateVideoAssetWithDetection(ID("r4"), "/test/video.mp4", "test-video", Duration("240240/24000s"))

		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
			return
		}

		// Should detect audio for .mp4 files
		if asset.HasAudio != "1" {
			t.Errorf("Video asset should have detected audio, hasAudio='%s'", asset.HasAudio)
		}

		if format == nil {
			t.Error("Video should have a format")
		}
	})
}

func TestFormatBuilder(t *testing.T) {
	tests := []struct {
		name        string
		id          ID
		formatName  string
		width       int
		height      int
		colorSpace  string
		mediaType   string
		expectError bool
		errorType   string
	}{
		{
			name:        "Valid image format should be created",
			id:          ID("r1"),
			formatName:  "Test Image Format",
			width:       1920,
			height:      1080,
			colorSpace:  "1-1-1",
			mediaType:   "image",
			expectError: false,
		},
		{
			name:        "Valid video format should be created",
			id:          ID("r2"),
			formatName:  "Test Video Format",
			width:       1920,
			height:      1080,
			colorSpace:  "1-1-1",
			mediaType:   "video",
			expectError: false,
		},
		{
			name:        "Unsupported media type should fail",
			id:          ID("r3"),
			formatName:  "Test Format",
			width:       1920,
			height:      1080,
			colorSpace:  "1-1-1",
			mediaType:   "unknown",
			expectError: true,
			errorType:   "unsupported format media type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewFormatBuilder()
			format, err := builder.CreateFormat(tt.id, tt.formatName, tt.width, tt.height, tt.colorSpace, tt.mediaType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorType) {
					t.Errorf("Expected error type '%s' but got: %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if format == nil {
					t.Error("Expected format but got nil")
					return
				}

				// Validate format properties
				if format.ID != string(tt.id) {
					t.Errorf("Format ID should be %s, got: %s", tt.id, format.ID)
				}
				if format.Width != "1920" {
					t.Errorf("Format width should be '1920', got: %s", format.Width)
				}
				if format.Height != "1080" {
					t.Errorf("Format height should be '1080', got: %s", format.Height)
				}

				// Check media-type specific properties
				switch tt.mediaType {
				case "image":
					if format.FrameDuration != "" {
						t.Errorf("Image format should not have frameDuration, got: %s", format.FrameDuration)
					}
				case "video":
					if format.FrameDuration == "" {
						t.Error("Video format should have frameDuration")
					}
				}
			}
		})
	}
}

func TestEffectBuilder(t *testing.T) {
	tests := []struct {
		name        string
		id          ID
		effectName  string
		uid         string
		expectError bool
		errorType   string
	}{
		{
			name:        "Valid text effect should be created",
			id:          ID("r1"),
			effectName:  "Text",
			uid:         ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
			expectError: false,
		},
		{
			name:        "Valid blur effect should be created",
			id:          ID("r2"),
			effectName:  "Gaussian Blur",
			uid:         "FFGaussianBlur",
			expectError: false,
		},
		{
			name:        "Unknown effect UID should fail",
			id:          ID("r3"),
			effectName:  "Unknown Effect",
			uid:         "com.unknown.effect",
			expectError: true,
			errorType:   "unknown effect UID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewEffectBuilder()
			effect, err := builder.CreateEffect(tt.id, tt.effectName, tt.uid)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorType) {
					t.Errorf("Expected error type '%s' but got: %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if effect == nil {
					t.Error("Expected effect but got nil")
					return
				}

				// Validate effect properties
				if effect.ID != string(tt.id) {
					t.Errorf("Effect ID should be %s, got: %s", tt.id, effect.ID)
				}
				if effect.Name != tt.effectName {
					t.Errorf("Effect name should be %s, got: %s", tt.effectName, effect.Name)
				}
				if effect.UID != tt.uid {
					t.Errorf("Effect UID should be %s, got: %s", tt.uid, effect.UID)
				}
			}
		})
	}
}

func TestAssetBuilderMediaTypeValidation(t *testing.T) {
	registry := NewReferenceRegistry()
	builder := NewAssetBuilder(registry, "image")

	// Test that asset builder enforces media type constraints
	t.Run("Image asset validation", func(t *testing.T) {
		asset := &Asset{
			ID:           "r1",
			Duration:     "0s",
			HasVideo:     "1",
			VideoSources: "1",
			// Correctly has no audio properties
		}

		err := builder.validateAssetMediaType(asset, "image")
		if err != nil {
			t.Errorf("Valid image asset should pass validation, got: %v", err)
		}

		// Test invalid image asset (with audio properties)
		invalidAsset := &Asset{
			ID:           "r1",
			Duration:     "0s",
			HasVideo:     "1",
			HasAudio:     "1", // Invalid for images
			VideoSources: "1",
		}

		err = builder.validateAssetMediaType(invalidAsset, "image")
		if err == nil {
			t.Error("Image asset with audio properties should fail validation")
		}
	})

	t.Run("Video asset validation", func(t *testing.T) {
		asset := &Asset{
			ID:           "r2",
			Duration:     "240240/24000s",
			HasVideo:     "1",
			VideoSources: "1",
			HasAudio:     "1",
			AudioSources: "1",
		}

		err := builder.validateAssetMediaType(asset, "video")
		if err != nil {
			t.Errorf("Valid video asset should pass validation, got: %v", err)
		}

		// Test invalid video asset (zero duration)
		invalidAsset := &Asset{
			ID:           "r2",
			Duration:     "0s", // Invalid for videos
			HasVideo:     "1",
			VideoSources: "1",
		}

		err = builder.validateAssetMediaType(invalidAsset, "video")
		if err == nil {
			t.Error("Video asset with zero duration should fail validation")
		}
	})

	t.Run("Format media type validation", func(t *testing.T) {
		// Test image format validation
		imageFormat := &Format{
			ID:     "r1",
			Width:  "1920",
			Height: "1080",
			// Correctly has no frameDuration
		}

		err := builder.validateFormatMediaType(imageFormat, "image")
		if err != nil {
			t.Errorf("Valid image format should pass validation, got: %v", err)
		}

		// Test invalid image format (with frameDuration)
		invalidFormat := &Format{
			ID:            "r1",
			Width:         "1920",
			Height:        "1080",
			FrameDuration: "1001/24000s", // Invalid for images
		}

		err = builder.validateFormatMediaType(invalidFormat, "image")
		if err == nil {
			t.Error("Image format with frameDuration should fail validation")
		}
	})
}

// Benchmark asset creation performance
func BenchmarkAssetCreation(b *testing.B) {
	registry := NewReferenceRegistry()
	builder := NewAssetBuilder(registry, "video")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := ID("r" + string(rune(i+1)))
		_, _, err := builder.CreateAsset(id, "/test/video.mp4", "test-video", Duration("240240/24000s"))
		if err != nil {
			b.Fatalf("Asset creation failed: %v", err)
		}
	}
}