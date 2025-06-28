package fcp

import (
	"strings"
	"testing"
)

// TestSpineStructuralValidation tests that the validateSpine method catches lane violations
// This ensures baffle and other direct manipulation is caught by the validation system
func TestSpineStructuralValidation(t *testing.T) {
	// Create an FCPXML with spine elements that have lane attributes (violates structure)
	fcpxml := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r2",
					Name:     "test.mp4",
					UID:      "test-uid",
					Start:    "0s",
					Duration: "240240/24000s",
					HasVideo: "1",
					MediaRep: MediaRep{
						Kind: "original-media",
						Src:  "file:///tmp/test/path.mp4",
					},
				},
			},
			Formats: []Format{
				{
					ID:     "r3",
					Name:   "FFVideoFormat1080p",
					Width:  "1920",
					Height: "1080",
				},
			},
		},
		Library: Library{
			Events: []Event{
				{
					Name: "Test Event",
					Projects: []Project{
						{
							Name: "Test Project",
							Sequences: []Sequence{
								{
									Duration: "240240/24000s",
									Spine: Spine{
										// These elements with lanes should be REJECTED
										AssetClips: []AssetClip{
											{
												Ref:      "r2",
												Lane:     "1", // ❌ INVALID: spine elements can't have lanes
												Offset:   "0s",
												Duration: "120120/24000s",
												Name:     "TestClip",
											},
										},
										Videos: []Video{
											{
												Ref:      "r2",
												Lane:     "5", // ❌ INVALID: spine elements can't have lanes
												Offset:   "120120/24000s", 
												Duration: "120120/24000s",
												Name:     "TestVideo",
											},
										},
										Titles: []Title{
											{
												Ref:      "r2",
												Lane:     "11", // ❌ INVALID: spine elements can't have lanes (baffle-style)
												Offset:   "60060/24000s",
												Duration: "60060/24000s",
												Name:     "TestTitle",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Test 1: ValidateStructure should reject spine elements with lanes
	err := fcpxml.ValidateStructure()
	if err == nil {
		t.Errorf("expected validation to fail for spine elements with lanes")
	}
	if err != nil && !strings.Contains(err.Error(), "spine") && !strings.Contains(err.Error(), "lane") {
		t.Errorf("expected spine lane validation error, got: %v", err)
	}

	// Test 2: ValidateAndMarshal should also reject
	_, err = fcpxml.ValidateAndMarshal()
	if err == nil {
		t.Errorf("expected marshal validation to fail for spine elements with lanes")
	}
	if err != nil && !strings.Contains(err.Error(), "spine") && !strings.Contains(err.Error(), "lane") {
		t.Errorf("expected spine lane validation error in marshal, got: %v", err)
	}
}

// TestSpineStructuralValidationValid tests that valid spine elements pass validation
func TestSpineStructuralValidationValid(t *testing.T) {
	// Create an FCPXML with valid spine elements (no lane attributes)
	fcpxml := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r2",
					Name:     "test.mp4",
					UID:      "test-uid",
					Start:    "0s",
					Duration: "240240/24000s",
					HasVideo: "1",
					MediaRep: MediaRep{
						Kind: "original-media",
						Src:  "file:///tmp/test/path.mp4",
					},
				},
			},
			Formats: []Format{
				{
					ID:     "r3",
					Name:   "FFVideoFormat1080p",
					Width:  "1920",
					Height: "1080",
				},
			},
		},
		Library: Library{
			Events: []Event{
				{
					Name: "Test Event",
					Projects: []Project{
						{
							Name: "Test Project",
							Sequences: []Sequence{
								{
									Duration: "240240/24000s",
									Spine: Spine{
										// These elements without lanes should be ACCEPTED
										AssetClips: []AssetClip{
											{
												Ref:      "r2",
												// Lane is omitted - valid for spine
												Offset:   "0s",
												Duration: "120120/24000s",
												Name:     "TestClip",
											},
										},
										Videos: []Video{
											{
												Ref:      "r2",
												// Lane is omitted - valid for spine
												Offset:   "120120/24000s", 
												Duration: "120120/24000s",
												Name:     "TestVideo",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Test: ValidateStructure should pass for valid spine elements
	err := fcpxml.ValidateStructure()
	if err != nil {
		t.Errorf("expected validation to pass for valid spine elements, got: %v", err)
	}

	// Test: ValidateAndMarshal should also pass
	_, err = fcpxml.ValidateAndMarshal()
	if err != nil {
		t.Errorf("expected marshal validation to pass for valid spine elements, got: %v", err)
	}
}

func TestValidationMarshaling(t *testing.T) {
	tests := []struct {
		name        string
		fcpxml      *FCPXML
		expectError bool
		errorType   string
	}{
		{
			name: "Valid FCPXML should marshal successfully",
			fcpxml: &FCPXML{
				Version: "1.13",
				Resources: Resources{
					Assets: []Asset{
						{
							ID:       "r2",
							Name:     "test.mp4",
							UID:      "test-uid",
							Start:    "0s",
							Duration: "240240/24000s",
							HasVideo: "1",
							MediaRep: MediaRep{
								Kind: "original-media",
								Src:  "file:///tmp/test/path.mp4",
							},
						},
					},
					Formats: []Format{
						{
							ID:            "r3",
							Name:          "FFVideoFormat1080p30",
							FrameDuration: "1001/24000s",
							Width:         "1920",
							Height:        "1080",
							ColorSpace:    "1-1-1",
						},
					},
				},
				Library: Library{
					Events: []Event{
						{
							Name: "Event",
							UID:  "event-uid",
							Projects: []Project{
								{
									Name: "Project",
									UID:  "project-uid",
									Sequences: []Sequence{
										{
											Format:      "r3",
											Duration:    "240240/24000s",
											TCStart:     "0s",
											TCFormat:    "NDF",
											AudioLayout: "stereo",
											AudioRate:   "48k",
											Spine: Spine{
												AssetClips: []AssetClip{
													{
														Ref:      "r2",
														Offset:   "0s",
														Duration: "240240/24000s",
														Name:     "test",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "FCPXML with missing version should fail",
			fcpxml: &FCPXML{
				// Missing Version
				Library: Library{
					Events: []Event{
						{
							Name: "Event",
							Projects: []Project{
								{
									Name: "Project",
									Sequences: []Sequence{
										{
											Duration: "240240/24000s",
											Spine:    Spine{},
										},
									},
								},
							},
						},
					},
				},
			},
			expectError: true,
			errorType:   "version is required",
		},
		{
			name: "FCPXML with no events should fail",
			fcpxml: &FCPXML{
				Version: "1.13",
				Library: Library{
					Events: []Event{}, // Empty events
				},
			},
			expectError: true,
			errorType:   "must have at least one event",
		},
		{
			name: "FCPXML with invalid duration format should fail",
			fcpxml: &FCPXML{
				Version: "1.13",
				Resources: Resources{
					Assets: []Asset{
						{
							ID:       "r2",
							Name:     "test.mp4",
							UID:      "test-uid",
							Duration: "invalid-duration", // Invalid format
							MediaRep: MediaRep{
								Src: "file:///tmp/test/path.mp4",
							},
						},
					},
				},
				Library: Library{
					Events: []Event{
						{
							Name: "Event",
							Projects: []Project{
								{
									Name: "Project",
									Sequences: []Sequence{
										{
											Duration: "240240/24000s",
											Spine:    Spine{},
										},
									},
								},
							},
						},
					},
				},
			},
			expectError: true,
			errorType:   "invalid asset duration",
		},
		{
			name: "FCPXML with dangling asset reference should fail",
			fcpxml: &FCPXML{
				Version: "1.13",
				Library: Library{
					Events: []Event{
						{
							Name: "Event",
							Projects: []Project{
								{
									Name: "Project",
									Sequences: []Sequence{
										{
											Duration: "240240/24000s",
											Spine: Spine{
												AssetClips: []AssetClip{
													{
														Ref:      "nonexistent", // Dangling reference
														Offset:   "0s",
														Duration: "240240/24000s",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectError: true,
			errorType:   "dangling asset reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.fcpxml.ValidateAndMarshal()

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
				if len(data) == 0 {
					t.Error("Expected XML data but got empty result")
				}
				// Verify it's valid XML by basic check
				if !strings.Contains(string(data), "<fcpxml") {
					t.Error("Generated data doesn't appear to be valid FCPXML")
				}
			}
		})
	}
}

func TestStructValidator(t *testing.T) {
	validator := NewStructValidator()

	tests := []struct {
		name        string
		resource    interface{}
		expectError bool
		errorType   string
	}{
		{
			name: "Valid asset should pass validation",
			resource: &AssetWrapper{
				Asset: &Asset{
					ID:       "r1",
					Name:     "test.mp4",
					UID:      "test-uid",
					Duration: "240240/24000s",
					MediaRep: MediaRep{
						Src: "file:///tmp/test/path.mp4",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Asset with missing ID should fail",
			resource: &AssetWrapper{
				Asset: &Asset{
					// Missing ID
					Name:     "test.mp4",
					UID:      "test-uid",
					Duration: "240240/24000s",
					MediaRep: MediaRep{
						Src: "file:///tmp/test/path.mp4",
					},
				},
			},
			expectError: true,
			errorType:   "ID is required",
		},
		{
			name: "Asset with invalid duration should fail",
			resource: &AssetWrapper{
				Asset: &Asset{
					ID:       "r1",
					Name:     "test.mp4",
					UID:      "test-uid",
					Duration: "invalid", // Invalid duration
					MediaRep: MediaRep{
						Src: "file:///tmp/test/path.mp4",
					},
				},
			},
			expectError: true,
			errorType:   "invalid asset duration",
		},
		{
			name: "Valid format should pass validation",
			resource: &FormatWrapper{
				Format: &Format{
					ID:            "r1",
					Name:          "Test Format",
					FrameDuration: "1001/24000s",
					Width:         "1920",
					Height:        "1080",
				},
			},
			expectError: false,
		},
		{
			name: "Format with missing width should fail",
			resource: &FormatWrapper{
				Format: &Format{
					ID:   "r1",
					Name: "Test Format",
					// Missing Width
					Height: "1080",
				},
			},
			expectError: true,
			errorType:   "width and height are required",
		},
		{
			name: "Valid effect should pass validation",
			resource: &EffectWrapper{
				Effect: &Effect{
					ID:   "r1",
					Name: "Test Effect",
					UID:  "com.test.effect",
				},
			},
			expectError: false,
		},
		{
			name: "Effect with missing name should fail",
			resource: &EffectWrapper{
				Effect: &Effect{
					ID: "r1",
					// Missing Name
					UID: "com.test.effect",
				},
			},
			expectError: true,
			errorType:   "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateResourceForMarshaling(tt.resource)

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
				}
			}
		})
	}
}

func TestValidateXMLStructure(t *testing.T) {
	tests := []struct {
		name        string
		xmlData     []byte
		expectError bool
		errorType   string
	}{
		{
			name: "Valid XML should pass",
			xmlData: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.13">
    <resources>
        <asset id="r1" name="test"/>
    </resources>
    <library>
        <event name="Event"/>
    </library>
</fcpxml>`),
			expectError: false,
		},
		{
			name: "XML with wrong root element should fail",
			xmlData: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<wrongroot version="1.13">
    <resources/>
</wrongroot>`),
			expectError: true,
			errorType:   "fcpxml",
		},
		{
			name: "XML without version should fail",
			xmlData: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<fcpxml>
    <resources/>
</fcpxml>`),
			expectError: true,
			errorType:   "must have version attribute",
		},
		{
			name: "Malformed XML should fail",
			xmlData: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.13">
    <unclosed-tag>
</fcpxml>`),
			expectError: true,
			errorType:   "not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateXMLStructure(tt.xmlData)

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
				}
			}
		})
	}
}

// Benchmark validation marshaling performance
func BenchmarkValidateAndMarshal(b *testing.B) {
	fcpxml := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r2",
					Name:     "test.mp4",
					UID:      "test-uid",
					Duration: "240240/24000s",
					HasVideo: "1",
					MediaRep: MediaRep{
						Src: "file:///tmp/test/path.mp4",
					},
				},
			},
		},
		Library: Library{
			Events: []Event{
				{
					Name: "Event",
					Projects: []Project{
						{
							Name: "Project",
							Sequences: []Sequence{
								{
									Duration: "240240/24000s",
									Spine: Spine{
										AssetClips: []AssetClip{
											{
												Ref:      "r2",
												Offset:   "0s",
												Duration: "240240/24000s",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fcpxml.ValidateAndMarshal()
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}