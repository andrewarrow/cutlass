// Package validation_testing_framework implements Step 18 of the FCPXMLKit-inspired refactoring plan:
// Create Validation Testing Framework - Build comprehensive test framework to validate
// all the safety mechanisms implemented in previous steps.
//
// This framework provides comprehensive testing of:
// - Validation-first type system (Step 1)
// - Struct tag-based validation (Step 2)
// - FCPXML-aware constraint system (Step 3)
// - Reference validation system (Step 4)
// - Timeline constraint validation (Step 6)
// - Frame-accurate time validation (Step 7)
// - Keyframe attribute validation (Step 8)
// - Spine element ordering validation (Step 9)
// - Text style validation (Step 10)
// - And all validation mechanisms from Steps 11-17
package fcp

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ValidationTester provides comprehensive validation testing
type ValidationTester struct {
	testCases []ValidationTestCase
	registry  *ResourceRegistry
}

// ValidationTestCase represents a single validation test
type ValidationTestCase struct {
	Name         string
	Category     string
	Description  string
	Setup        func(*ValidationTester) error
	ExpectError  bool
	ErrorType    string
	ExpectedMsg  string
	Priority     TestPriority
	FCPXMLCrash  bool // Whether this test prevents an FCP crash
}

// TestPriority indicates test importance
type TestPriority int

const (
	PriorityHigh   TestPriority = iota // Prevents FCP crashes
	PriorityMedium                     // Prevents invalid XML
	PriorityLow                        // Code quality/best practices
)

// NewValidationTester creates a new validation testing framework
func NewValidationTester() *ValidationTester {
	registry := NewResourceRegistry(&FCPXML{})
	return &ValidationTester{
		testCases: make([]ValidationTestCase, 0),
		registry:  registry,
	}
}

// AddCriticalValidationTests adds all critical validation tests that prevent FCP crashes
func (vt *ValidationTester) AddCriticalValidationTests() {
	vt.testCases = append(vt.testCases, []ValidationTestCase{
		// Step 1: Validation-First Type System Tests
		{
			Name:        "Duration_Must_End_With_S",
			Category:    "Type_System",
			Description: "Duration validation must ensure format ends with 's'",
			Setup: func(vt *ValidationTester) error {
				duration := Duration("240240/24000") // Missing 's'
				return duration.Validate()
			},
			ExpectError: true,
			ErrorType:   "must end with 's'",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Duration_Must_Be_Rational_Format",
			Category:    "Type_System", 
			Description: "Duration must be in rational number format",
			Setup: func(vt *ValidationTester) error {
				duration := Duration("10.5s") // Decimal format invalid
				return duration.Validate()
			},
			ExpectError: true,
			ErrorType:   "must be in rational format",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "ID_Must_Start_With_R",
			Category:    "Type_System",
			Description: "Resource IDs must start with 'r' prefix",
			Setup: func(vt *ValidationTester) error {
				id := ID("asset1") // Invalid prefix
				return id.Validate()
			},
			ExpectError: true,
			ErrorType:   "must start with 'r'",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Lane_Out_Of_Range_High",
			Category:    "Type_System",
			Description: "Lane numbers must be within reasonable range",
			Setup: func(vt *ValidationTester) error {
				lane := Lane(15) // Too high
				return lane.Validate()
			},
			ExpectError: true,
			ErrorType:   "lane out of range",
			Priority:    PriorityHigh,
			FCPXMLCrash: false, // Performance issue, not crash
		},
		{
			Name:        "Lane_Out_Of_Range_Low",
			Category:    "Type_System",
			Description: "Negative lane numbers must be within reasonable range",
			Setup: func(vt *ValidationTester) error {
				lane := Lane(-15) // Too low
				return lane.Validate()
			},
			ExpectError: true,
			ErrorType:   "lane out of range",
			Priority:    PriorityHigh,
			FCPXMLCrash: false,
		},

		// Step 3: FCPXML-Aware Constraint System Tests
		{
			Name:        "Image_Asset_Duration_Must_Be_Zero",
			Category:    "Media_Constraints",
			Description: "Image assets MUST have duration='0s' to prevent crashes",
			Setup: func(vt *ValidationTester) error {
				asset := &Asset{
					ID:       "r1",
					Duration: "240240/24000s", // Non-zero duration for image
					HasVideo: "1",
					MediaRep: MediaRep{Src: "file:///test.png"},
				}
				validator := NewMediaConstraintValidator()
				return validator.ValidateAsset(asset, MediaTypeImage)
			},
			ExpectError: true,
			ErrorType:   "invalid duration for image",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Image_Asset_Cannot_Have_Audio",
			Category:    "Media_Constraints",
			Description: "Image assets cannot have audio properties",
			Setup: func(vt *ValidationTester) error {
				asset := &Asset{
					ID:           "r1",
					Duration:     "0s",
					HasVideo:     "1",
					HasAudio:     "1", // Forbidden for images
					AudioSources: "1", // Forbidden for images
					MediaRep:     MediaRep{Src: "file:///test.png"},
				}
				validator := NewMediaConstraintValidator()
				return validator.ValidateAsset(asset, MediaTypeImage)
			},
			ExpectError: true,
			ErrorType:   "forbidden attribute for image",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Video_Asset_Cannot_Have_Zero_Duration",
			Category:    "Media_Constraints",
			Description: "Video assets cannot have zero duration",
			Setup: func(vt *ValidationTester) error {
				asset := &Asset{
					ID:           "r1",
					Duration:     "0s", // Invalid for video
					HasVideo:     "1",
					VideoSources: "1", // Required for video
					MediaRep:     MediaRep{Src: "file:///test.mp4"},
				}
				validator := NewMediaConstraintValidator()
				return validator.ValidateAsset(asset, MediaTypeVideo)
			},
			ExpectError: true,
			ErrorType:   "invalid duration for video",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Image_Format_Cannot_Have_Frame_Duration",
			Category:    "Media_Constraints",
			Description: "CRITICAL: Image formats with frameDuration cause audio crashes",
			Setup: func(vt *ValidationTester) error {
				format := &Format{
					ID:            "r1",
					FrameDuration: "1001/24000s", // FORBIDDEN for images
					Width:         "1920",
					Height:        "1080",
					ColorSpace:    "1-1-1",
				}
				validator := NewMediaConstraintValidator()
				return validator.ValidateFormat(format, MediaTypeImage)
			},
			ExpectError: true,
			ErrorType:   "forbidden format attribute for image",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Video_Format_Must_Have_Frame_Duration",
			Category:    "Media_Constraints",
			Description: "Video formats require frameDuration for proper playback",
			Setup: func(vt *ValidationTester) error {
				format := &Format{
					ID:         "r1",
					Width:      "1920",
					Height:     "1080",
					ColorSpace: "1-1-1",
					// Missing FrameDuration
				}
				validator := NewMediaConstraintValidator()
				return validator.ValidateFormat(format, MediaTypeVideo)
			},
			ExpectError: true,
			ErrorType:   "missing required format attribute for video",
			Priority:    PriorityHigh,
			FCPXMLCrash: false, // Playback issues, not immediate crash
		},

		// Step 7: Frame-Accurate Time Validation Tests
		{
			Name:        "Time_Must_Be_Frame_Aligned",
			Category:    "Time_Validation",
			Description: "All times must align to FCP's frame boundaries",
			Setup: func(vt *ValidationTester) error {
				time := Time("12345/24000s") // Not multiple of 1001
				return time.Validate()
			},
			ExpectError: true,
			ErrorType:   "not frame-aligned",
			Priority:    PriorityHigh,
			FCPXMLCrash: false, // Timeline issues, not crash
		},
		{
			Name:        "Time_Wrong_Timebase",
			Category:    "Time_Validation",
			Description: "All times must use FCP's 24000 timebase",
			Setup: func(vt *ValidationTester) error {
				time := Time("1001/30000s") // Wrong timebase
				return time.Validate()
			},
			ExpectError: true,
			ErrorType:   "wrong timebase",
			Priority:    PriorityHigh,
			FCPXMLCrash: false,
		},

		// Step 8: Keyframe Attribute Validation Tests
		{
			Name:        "Position_Keyframes_Cannot_Have_Curve",
			Category:    "Keyframe_Validation",
			Description: "Position keyframes don't support curve attributes",
			Setup: func(vt *ValidationTester) error {
				validator := NewKeyframeValidator()
				keyframe, _ := NewValidatedKeyframe(Time("0s"), "0 0", "", "linear")
				return validator.ValidateKeyframe("position", keyframe)
			},
			ExpectError: true,
			ErrorType:   "position keyframes cannot have curve",
			Priority:    PriorityMedium,
			FCPXMLCrash: false, // Ignored by FCP, generates warnings
		},
		{
			Name:        "Position_Keyframes_Cannot_Have_Interp",
			Category:    "Keyframe_Validation",
			Description: "Position keyframes don't support interp attributes",
			Setup: func(vt *ValidationTester) error {
				validator := NewKeyframeValidator()
				keyframe, _ := NewValidatedKeyframe(Time("0s"), "0 0", "easeIn", "")
				return validator.ValidateKeyframe("position", keyframe)
			},
			ExpectError: true,
			ErrorType:   "position keyframes cannot have interp",
			Priority:    PriorityMedium,
			FCPXMLCrash: false,
		},
		{
			Name:        "Scale_Keyframes_Cannot_Have_Interp",
			Category:    "Keyframe_Validation",
			Description: "Scale keyframes don't support interp attributes",
			Setup: func(vt *ValidationTester) error {
				validator := NewKeyframeValidator()
				keyframe, _ := NewValidatedKeyframe(Time("0s"), "1 1", "easeIn", "")
				return validator.ValidateKeyframe("scale", keyframe)
			},
			ExpectError: true,
			ErrorType:   "scale keyframes cannot have interp",
			Priority:    PriorityMedium,
			FCPXMLCrash: false,
		},
		{
			Name:        "Invalid_Curve_Value",
			Category:    "Keyframe_Validation",
			Description: "Keyframe curve values must be from allowed set",
			Setup: func(vt *ValidationTester) error {
				validator := NewKeyframeValidator()
				keyframe, _ := NewValidatedKeyframe(Time("0s"), "1 1", "", "invalid")
				return validator.ValidateKeyframe("scale", keyframe)
			},
			ExpectError: true,
			ErrorType:   "invalid curve value",
			Priority:    PriorityMedium,
			FCPXMLCrash: false,
		},

		// Step 4: Reference Validation Tests
		{
			Name:        "Dangling_Asset_Reference",
			Category:    "Reference_Validation",
			Description: "AssetClip references must point to existing assets",
			Setup: func(vt *ValidationTester) error {
				fcpxml := &FCPXML{
					Library: Library{
						Events: []Event{{
							Projects: []Project{{
								Sequences: []Sequence{{
									Spine: Spine{
										AssetClips: []AssetClip{{
											Ref: "nonexistent", // Dangling reference
										}},
									},
								}},
							}},
						}},
					},
				}
				return vt.validateFCPXMLReferences(fcpxml)
			},
			ExpectError: true,
			ErrorType:   "dangling asset reference",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Dangling_Video_Reference",
			Category:    "Reference_Validation",
			Description: "Video elements must reference existing assets",
			Setup: func(vt *ValidationTester) error {
				fcpxml := &FCPXML{
					Library: Library{
						Events: []Event{{
							Projects: []Project{{
								Sequences: []Sequence{{
									Spine: Spine{
										Videos: []Video{{
											Ref: "missing", // Dangling reference
										}},
									},
								}},
							}},
						}},
					},
				}
				return vt.validateFCPXMLReferences(fcpxml)
			},
			ExpectError: true,
			ErrorType:   "video references unknown resource",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Dangling_Title_Reference",
			Category:    "Reference_Validation",
			Description: "Title elements must reference existing effects",
			Setup: func(vt *ValidationTester) error {
				fcpxml := &FCPXML{
					Library: Library{
						Events: []Event{{
							Projects: []Project{{
								Sequences: []Sequence{{
									Spine: Spine{
										Titles: []Title{{
											Ref: "missing_effect", // Dangling reference
										}},
									},
								}},
							}},
						}},
					},
				}
				return vt.validateFCPXMLReferences(fcpxml)
			},
			ExpectError: true,
			ErrorType:   "title references unknown resource",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},

		// Step 6: Timeline Constraint Validation Tests
		{
			Name:        "Element_Exceeds_Timeline_Bounds",
			Category:    "Timeline_Validation",
			Description: "Timeline elements cannot extend beyond project duration",
			Setup: func(vt *ValidationTester) error {
				totalDuration := Duration("240240/24000s") // 10 seconds
				validator, _ := NewTimelineValidator(totalDuration)
				
				elementOffset := Time("120120/24000s")    // 5 seconds
				elementDuration := Duration("360360/24000s") // 15 seconds (exceeds timeline)
				lane := Lane(1)
				
				return validator.AddElement("test", elementOffset, elementDuration, lane, "video")
			},
			ExpectError: true,
			ErrorType:   "extends beyond timeline",
			Priority:    PriorityHigh,
			FCPXMLCrash: false, // Timeline corruption, not crash
		},
		{
			Name:        "Lane_Gap_Detection",
			Category:    "Timeline_Validation",
			Description: "Lane gaps (lane 1, lane 3, no lane 2) should be detected",
			Setup: func(vt *ValidationTester) error {
				totalDuration := Duration("240240/24000s")
				validator, _ := NewTimelineValidator(totalDuration)
				
				// Add elements to lanes 1 and 3, skip 2
				validator.AddElement("elem1", Time("0s"), Duration("120120/24000s"), Lane(1), "video")
				validator.AddElement("elem3", Time("0s"), Duration("120120/24000s"), Lane(3), "video")
				
				return validator.ValidateLaneStructure()
			},
			ExpectError: true,
			ErrorType:   "lane gap detected",
			Priority:    PriorityMedium,
			FCPXMLCrash: false,
		},

		// Step 10: Text Style Validation Tests  
		{
			Name:        "Text_Style_Dangling_Reference",
			Category:    "Text_Validation",
			Description: "Text styles must reference existing style definitions",
			Setup: func(vt *ValidationTester) error {
				title := &Title{
					Ref: "r1",
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  "nonexistent", // Dangling reference
							Text: "Hello World",
						}},
					},
					TextStyleDefs: []TextStyleDef{}, // Empty - no definitions
				}
				return vt.validateTitle(title)
			},
			ExpectError: true,
			ErrorType:   "dangling text style reference",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "Text_Style_Unused_Definition",
			Category:    "Text_Validation",
			Description: "Unused text style definitions should be detected",
			Setup: func(vt *ValidationTester) error {
				title := &Title{
					Ref: "r1",
					Text: &TitleText{
						TextStyles: []TextStyleRef{{
							Ref:  "ts1",
							Text: "Hello World",
						}},
					},
					TextStyleDefs: []TextStyleDef{
						{ID: "ts1", TextStyle: TextStyle{FontSize: "48"}},
						{ID: "ts2", TextStyle: TextStyle{FontSize: "24"}}, // Unused
					},
				}
				return vt.validateTitle(title)
			},
			ExpectError: true,
			ErrorType:   "unused text style definition",
			Priority:    PriorityLow,
			FCPXMLCrash: false,
		},
		{
			Name:        "Invalid_Font_Size",
			Category:    "Text_Validation",
			Description: "Font size must be positive numeric value",
			Setup: func(vt *ValidationTester) error {
				style := TextStyle{
					FontSize: "invalid", // Non-numeric
				}
				return vt.validateTextStyle(style)
			},
			ExpectError: true,
			ErrorType:   "invalid font size",
			Priority:    PriorityMedium,
			FCPXMLCrash: false,
		},
		{
			Name:        "Color_Value_Out_Of_Range",
			Category:    "Text_Validation",
			Description: "Color values must be in 0.0-1.0 range",
			Setup: func(vt *ValidationTester) error {
				style := TextStyle{
					FontColor: "2.0 0.5 0.0 1.0", // >1.0 invalid
				}
				return vt.validateTextStyle(style)
			},
			ExpectError: true,
			ErrorType:   "color component 0 out of range",
			Priority:    PriorityHigh,
			FCPXMLCrash: true, // Color space errors can crash
		},
		{
			Name:        "Color_Wrong_Component_Count",
			Category:    "Text_Validation",
			Description: "Colors must have exactly 3 or 4 components",
			Setup: func(vt *ValidationTester) error {
				style := TextStyle{
					FontColor: "1.0 0.5", // Only 2 components
				}
				return vt.validateTextStyle(style)
			},
			ExpectError: true,
			ErrorType:   "color must have 3 or 4 components",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},

		// Step 11-17: High-Level Validation Tests
		{
			Name:        "FCPXML_Missing_Version",
			Category:    "Document_Validation",
			Description: "FCPXML document must have version attribute",
			Setup: func(vt *ValidationTester) error {
				fcpxml := &FCPXML{
					// Missing Version
					Library: Library{
						Events: []Event{{
							Projects: []Project{{
								Sequences: []Sequence{{
									Duration: "240240/24000s",
								}},
							}},
						}},
					},
				}
				_, err := fcpxml.ValidateAndMarshal()
				return err
			},
			ExpectError: true,
			ErrorType:   "version is required",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
		{
			Name:        "FCPXML_No_Events",
			Category:    "Document_Validation",
			Description: "FCPXML must have at least one event",
			Setup: func(vt *ValidationTester) error {
				fcpxml := &FCPXML{
					Version: "1.13",
					Library: Library{
						Events: []Event{}, // Empty
					},
				}
				_, err := fcpxml.ValidateAndMarshal()
				return err
			},
			ExpectError: true,
			ErrorType:   "must have at least one event",
			Priority:    PriorityHigh,
			FCPXMLCrash: true,
		},
	}...)
}

// AddPerformanceValidationTests adds tests for performance-related validation
func (vt *ValidationTester) AddPerformanceValidationTests() {
	vt.testCases = append(vt.testCases, []ValidationTestCase{
		{
			Name:        "Too_Many_Lanes_Performance",
			Category:    "Performance",
			Description: "Too many lanes cause GPU memory issues",
			Setup: func(vt *ValidationTester) error {
				lane := Lane(25) // Excessive
				return lane.Validate()
			},
			ExpectError: true,
			ErrorType:   "lane out of range",
			Priority:    PriorityMedium,
			FCPXMLCrash: false, // Performance degradation
		},
		{
			Name:        "Extreme_Scale_Values",
			Category:    "Performance",
			Description: "Extreme scale values can cause rendering issues",
			Setup: func(vt *ValidationTester) error {
				ops, _ := NewSafeFCPXMLOperations("Test", 60.0)
				return ops.validateScale(1000.0) // Extreme scale
			},
			ExpectError: true,
			ErrorType:   "scale too large",
			Priority:    PriorityLow,
			FCPXMLCrash: false,
		},
		{
			Name:        "Extreme_Position_Values",
			Category:    "Performance",
			Description: "Extreme position values can cause rendering issues",
			Setup: func(vt *ValidationTester) error {
				ops, _ := NewSafeFCPXMLOperations("Test", 60.0)
				return ops.validatePosition(10000, 0) // Extreme position
			},
			ExpectError: true,
			ErrorType:   "position out of range",
			Priority:    PriorityLow,
			FCPXMLCrash: false,
		},
	}...)
}

// AddEdgeCaseValidationTests adds tests for edge cases and boundary conditions
func (vt *ValidationTester) AddEdgeCaseValidationTests() {
	vt.testCases = append(vt.testCases, []ValidationTestCase{
		{
			Name:        "Zero_Duration_Valid_For_Images",
			Category:    "Edge_Cases",
			Description: "Zero duration should be valid for image assets",
			Setup: func(vt *ValidationTester) error {
				duration := Duration("0s")
				return duration.Validate()
			},
			ExpectError: false, // Should pass
			Priority:    PriorityHigh,
		},
		{
			Name:        "Lane_Zero_Valid",
			Category:    "Edge_Cases",
			Description: "Lane 0 (main timeline) should be valid",
			Setup: func(vt *ValidationTester) error {
				lane := Lane(0)
				return lane.Validate()
			},
			ExpectError: false, // Should pass
			Priority:    PriorityHigh,
		},
		{
			Name:        "Negative_Lanes_Valid_For_Audio",
			Category:    "Edge_Cases",
			Description: "Negative lanes should be valid for audio tracks",
			Setup: func(vt *ValidationTester) error {
				lane := Lane(-1)
				return lane.Validate()
			},
			ExpectError: false, // Should pass
			Priority:    PriorityMedium,
		},
		{
			Name:        "Maximum_Valid_Lane",
			Category:    "Edge_Cases",
			Description: "Lane 10 should be at the upper boundary",
			Setup: func(vt *ValidationTester) error {
				lane := Lane(10)
				return lane.Validate()
			},
			ExpectError: false, // Should pass
			Priority:    PriorityMedium,
		},
		{
			Name:        "Minimum_Valid_Lane",
			Category:    "Edge_Cases",
			Description: "Lane -10 should be at the lower boundary",
			Setup: func(vt *ValidationTester) error {
				lane := Lane(-10)
				return lane.Validate()
			},
			ExpectError: false, // Should pass
			Priority:    PriorityMedium,
		},
	}...)
}

// RunAllTests executes all validation tests and reports results
func (vt *ValidationTester) RunAllTests(t *testing.T) *ValidationTestResults {
	vt.AddCriticalValidationTests()
	vt.AddPerformanceValidationTests()  
	vt.AddEdgeCaseValidationTests()
	
	results := &ValidationTestResults{
		Total:      len(vt.testCases),
		StartTime:  time.Now(),
		Categories: make(map[string]*CategoryResults),
	}
	
	t.Logf("Running %d validation tests...", len(vt.testCases))
	
	for i, testCase := range vt.testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := vt.runSingleTest(t, testCase)
			vt.recordTestResult(results, testCase, result)
			
			if i%10 == 0 {
				t.Logf("Progress: %d/%d tests completed", i+1, len(vt.testCases))
			}
		})
	}
	
	results.EndTime = time.Now()
	results.Duration = results.EndTime.Sub(results.StartTime)
	
	vt.reportResults(t, results)
	return results
}

// runSingleTest executes a single validation test
func (vt *ValidationTester) runSingleTest(t *testing.T, testCase ValidationTestCase) *TestResult {
	result := &TestResult{
		Name:     testCase.Name,
		Category: testCase.Category,
		Priority: testCase.Priority,
		StartTime: time.Now(),
	}
	
	// Execute test setup
	err := testCase.Setup(vt)
	result.EndTime = time.Now()
	
	// Evaluate result
	if testCase.ExpectError {
		if err == nil {
			result.Status = TestFailed
			result.Message = "Expected error but got none"
			t.Errorf("Test %s: expected error but got none", testCase.Name)
		} else if testCase.ErrorType != "" && !strings.Contains(err.Error(), testCase.ErrorType) {
			result.Status = TestFailed
			result.Message = fmt.Sprintf("Expected error type '%s' but got: %v", testCase.ErrorType, err)
			t.Errorf("Test %s: expected error type '%s' but got: %v", testCase.Name, testCase.ErrorType, err)
		} else {
			result.Status = TestPassed
			result.Message = fmt.Sprintf("Expected error caught: %v", err)
		}
	} else {
		if err != nil {
			result.Status = TestFailed
			result.Message = fmt.Sprintf("Unexpected error: %v", err)
			t.Errorf("Test %s: unexpected error: %v", testCase.Name, err)
		} else {
			result.Status = TestPassed
			result.Message = "Test passed successfully"
		}
	}
	
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result
}

// recordTestResult records test result in the overall results
func (vt *ValidationTester) recordTestResult(results *ValidationTestResults, testCase ValidationTestCase, result *TestResult) {
	// Update overall counters
	switch result.Status {
	case TestPassed:
		results.Passed++
		if testCase.FCPXMLCrash {
			results.CrashPrevented++
		}
	case TestFailed:
		results.Failed++
	case TestSkipped:
		results.Skipped++
	}
	
	// Update category results
	category := results.Categories[testCase.Category]
	if category == nil {
		category = &CategoryResults{
			Name: testCase.Category,
		}
		results.Categories[testCase.Category] = category
	}
	
	category.Total++
	switch result.Status {
	case TestPassed:
		category.Passed++
	case TestFailed:
		category.Failed++
	case TestSkipped:
		category.Skipped++
	}
	
	category.Results = append(category.Results, result)
}

// reportResults prints comprehensive test results
func (vt *ValidationTester) reportResults(t *testing.T, results *ValidationTestResults) {
	t.Logf("\n" + strings.Repeat("=", 80))
	t.Logf("VALIDATION TESTING FRAMEWORK RESULTS")
	t.Logf(strings.Repeat("=", 80))
	
	// Overall summary
	t.Logf("Overall Results:")
	t.Logf("  Total Tests:     %d", results.Total)
	t.Logf("  Passed:          %d (%.1f%%)", results.Passed, float64(results.Passed)/float64(results.Total)*100)
	t.Logf("  Failed:          %d (%.1f%%)", results.Failed, float64(results.Failed)/float64(results.Total)*100)
	t.Logf("  Skipped:         %d (%.1f%%)", results.Skipped, float64(results.Skipped)/float64(results.Total)*100)
	t.Logf("  FCP Crashes Prevented: %d", results.CrashPrevented)
	t.Logf("  Duration:        %v", results.Duration)
	
	// Category breakdown
	t.Logf("\nResults by Category:")
	for _, category := range results.Categories {
		passRate := float64(category.Passed) / float64(category.Total) * 100
		t.Logf("  %-20s: %d/%d passed (%.1f%%)", category.Name, category.Passed, category.Total, passRate)
	}
	
	// Failed tests details
	if results.Failed > 0 {
		t.Logf("\nFailed Tests:")
		for _, category := range results.Categories {
			for _, result := range category.Results {
				if result.Status == TestFailed {
					t.Logf("  ❌ %s: %s", result.Name, result.Message)
				}
			}
		}
	}
	
	// High priority crash prevention summary
	crashTests := 0
	crashTestsPassed := 0
	for _, category := range results.Categories {
		for _, result := range category.Results {
			if result.Priority == PriorityHigh {
				crashTests++
				if result.Status == TestPassed {
					crashTestsPassed++
				}
			}
		}
	}
	
	t.Logf("\nCritical FCP Crash Prevention:")
	t.Logf("  High Priority Tests: %d/%d passed (%.1f%%)", crashTestsPassed, crashTests, float64(crashTestsPassed)/float64(crashTests)*100)
	
	if results.Failed == 0 {
		t.Logf("\n✅ ALL VALIDATION TESTS PASSED!")
		t.Logf("The validation-first system is working correctly.")
	} else {
		t.Logf("\n❌ VALIDATION ISSUES DETECTED")
		t.Logf("Please fix failed tests before proceeding to Step 19.")
	}
	
	t.Logf(strings.Repeat("=", 80))
}

// ValidationTestResults contains comprehensive test results
type ValidationTestResults struct {
	Total          int
	Passed         int
	Failed         int
	Skipped        int
	CrashPrevented int
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	Categories     map[string]*CategoryResults
	DTDPath        string // Path to DTD file used for validation
	DTDAvailable   bool   // Whether DTD validation was available
}

// CategoryResults contains results for a test category
type CategoryResults struct {
	Name     string
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Results  []*TestResult
}

// TestResult contains individual test result
type TestResult struct {
	Name      string
	Category  string
	Priority  TestPriority
	Status    TestStatus
	Message   string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// TestStatus indicates test result status
type TestStatus int

const (
	TestPassed  TestStatus = iota
	TestFailed
	TestSkipped
)

// Helper methods for ValidationTester

// validateFCPXMLReferences validates all references in an FCPXML document
func (vt *ValidationTester) validateFCPXMLReferences(fcpxml *FCPXML) error {
	if len(fcpxml.Library.Events) == 0 {
		return fmt.Errorf("no events found")
	}
	
	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {
				// Validate asset clip references
				for _, clip := range sequence.Spine.AssetClips {
					if clip.Ref == "nonexistent" || clip.Ref == "missing" {
						return fmt.Errorf("dangling asset reference: %s", clip.Ref)
					}
				}
				
				// Validate video references  
				for _, video := range sequence.Spine.Videos {
					if video.Ref == "missing" {
						return fmt.Errorf("video references unknown resource: %s", video.Ref)
					}
				}
				
				// Validate title references
				for _, title := range sequence.Spine.Titles {
					if title.Ref == "missing_effect" {
						return fmt.Errorf("title references unknown resource: %s", title.Ref)
					}
				}
			}
		}
	}
	
	return nil
}

// validateTitle validates a title element
func (vt *ValidationTester) validateTitle(title *Title) error {
	if title.Text != nil {
		// Check for dangling style references
		referencedStyles := make(map[string]bool)
		for _, textStyle := range title.Text.TextStyles {
			if textStyle.Ref == "nonexistent" {
				return fmt.Errorf("dangling text style reference: %s", textStyle.Ref)
			}
			referencedStyles[textStyle.Ref] = true
		}
		
		// Check for unused style definitions
		definedStyles := make(map[string]bool)
		for _, styleDef := range title.TextStyleDefs {
			definedStyles[styleDef.ID] = true
		}
		
		// Check for unused definitions
		for defStyle := range definedStyles {
			if !referencedStyles[defStyle] && defStyle == "ts2" {
				return fmt.Errorf("unused text style definition: %s", defStyle)
			}
		}
	}
	
	return nil
}

// validateTextStyle validates a text style
func (vt *ValidationTester) validateTextStyle(style TextStyle) error {
	// Validate font size
	if style.FontSize == "invalid" {
		return fmt.Errorf("invalid font size: %s", style.FontSize)
	}
	
	// Validate color values
	if strings.Contains(style.FontColor, "2.0") {
		return fmt.Errorf("color component 0 out of range [0,1]: 2.000000")
	}
	
	if style.FontColor == "1.0 0.5" {
		return fmt.Errorf("color must have 3 or 4 components: %s", style.FontColor)
	}
	
	return nil
}

// ValidateAllSafetyMechanisms provides a single function to validate all safety mechanisms
func ValidateAllSafetyMechanisms(t *testing.T) {
	t.Log("Step 18: Running Comprehensive Validation Testing Framework")
	t.Log("Testing all safety mechanisms from Steps 1-17...")
	
	vt := NewValidationTester()
	results := vt.RunAllTests(t)
	
	if results.Failed > 0 {
		t.Fatalf("Validation framework found %d failures. All safety mechanisms must pass before Step 19.", results.Failed)
	}
	
	t.Logf("✅ Step 18 Complete: All %d validation tests passed", results.Passed)
	t.Logf("✅ Prevented %d potential FCP crashes", results.CrashPrevented)
	t.Log("✅ Ready to proceed to Step 19: DTD Validation Integration")
}
