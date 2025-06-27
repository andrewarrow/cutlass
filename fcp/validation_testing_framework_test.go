// Package validation_testing_framework_test tests the Step 18 implementation:
// Comprehensive Validation Testing Framework
package fcp

import (
	"testing"
)

// TestValidationTestingFramework runs the comprehensive validation testing framework
func TestValidationTestingFramework(t *testing.T) {
	t.Log("=== Step 18: Comprehensive Validation Testing Framework ===")
	t.Log("Testing all validation mechanisms from Steps 1-17...")
	
	// Run the complete validation framework
	ValidateAllSafetyMechanisms(t)
}

// TestNewValidationTester tests the creation of validation tester
func TestNewValidationTester(t *testing.T) {
	vt := NewValidationTester()
	
	if vt == nil {
		t.Fatal("Expected validation tester to be created")
	}
	
	if vt.registry == nil {
		t.Error("Expected validation tester to have registry")
	}
	
	if vt.testCases == nil {
		t.Error("Expected validation tester to have test cases slice")
	}
}

// TestCriticalValidationTests tests the critical validation test loading
func TestCriticalValidationTests(t *testing.T) {
	vt := NewValidationTester()
	
	// Initially no tests
	if len(vt.testCases) != 0 {
		t.Errorf("Expected 0 initial test cases, got %d", len(vt.testCases))
	}
	
	// Add critical tests
	vt.AddCriticalValidationTests()
	
	if len(vt.testCases) == 0 {
		t.Error("Expected critical validation tests to be added")
	}
	
	// Check we have high priority crash prevention tests
	highPriorityTests := 0
	crashPreventionTests := 0
	
	for _, test := range vt.testCases {
		if test.Priority == PriorityHigh {
			highPriorityTests++
		}
		if test.FCPXMLCrash {
			crashPreventionTests++
		}
	}
	
	if highPriorityTests == 0 {
		t.Error("Expected high priority tests")
	}
	
	if crashPreventionTests == 0 {
		t.Error("Expected crash prevention tests")
	}
	
	t.Logf("Added %d critical validation tests", len(vt.testCases))
	t.Logf("  - %d high priority tests", highPriorityTests)
	t.Logf("  - %d crash prevention tests", crashPreventionTests)
}

// TestPerformanceValidationTests tests performance-related validation
func TestPerformanceValidationTests(t *testing.T) {
	vt := NewValidationTester()
	vt.AddPerformanceValidationTests()
	
	if len(vt.testCases) == 0 {
		t.Error("Expected performance validation tests to be added")
	}
	
	// Check for performance category tests
	performanceTests := 0
	for _, test := range vt.testCases {
		if test.Category == "Performance" {
			performanceTests++
		}
	}
	
	if performanceTests == 0 {
		t.Error("Expected performance category tests")
	}
	
	t.Logf("Added %d performance validation tests", performanceTests)
}

// TestEdgeCaseValidationTests tests edge case validation
func TestEdgeCaseValidationTests(t *testing.T) {
	vt := NewValidationTester()
	vt.AddEdgeCaseValidationTests()
	
	if len(vt.testCases) == 0 {
		t.Error("Expected edge case validation tests to be added")
	}
	
	// Check for edge case category tests
	edgeCaseTests := 0
	for _, test := range vt.testCases {
		if test.Category == "Edge_Cases" {
			edgeCaseTests++
		}
	}
	
	if edgeCaseTests == 0 {
		t.Error("Expected edge case category tests")
	}
	
	t.Logf("Added %d edge case validation tests", edgeCaseTests)
}

// TestValidationTestCategories ensures all major validation categories are covered
func TestValidationTestCategories(t *testing.T) {
	vt := NewValidationTester()
	vt.AddCriticalValidationTests()
	vt.AddPerformanceValidationTests()
	vt.AddEdgeCaseValidationTests()
	
	expectedCategories := map[string]bool{
		"Type_System":         false,
		"Media_Constraints":   false,
		"Time_Validation":     false,
		"Keyframe_Validation": false,
		"Reference_Validation": false,
		"Timeline_Validation": false,
		"Text_Validation":     false,
		"Document_Validation": false,
		"Performance":         false,
		"Edge_Cases":          false,
	}
	
	// Mark categories that exist
	for _, test := range vt.testCases {
		if _, exists := expectedCategories[test.Category]; exists {
			expectedCategories[test.Category] = true
		}
	}
	
	// Check all categories are present
	missingCategories := []string{}
	for category, found := range expectedCategories {
		if !found {
			missingCategories = append(missingCategories, category)
		}
	}
	
	if len(missingCategories) > 0 {
		t.Errorf("Missing validation test categories: %v", missingCategories)
	}
	
	t.Logf("All %d validation categories are covered", len(expectedCategories))
}

// TestSpecificValidationMechanisms tests specific validation mechanisms individually
func TestSpecificValidationMechanisms(t *testing.T) {
	// Test Duration validation (Step 1)
	t.Run("Duration_Validation", func(t *testing.T) {
		// Valid duration
		duration := Duration("240240/24000s")
		if err := duration.Validate(); err != nil {
			t.Errorf("Valid duration should pass: %v", err)
		}
		
		// Invalid duration - no 's'
		invalidDuration := Duration("240240/24000")
		if err := invalidDuration.Validate(); err == nil {
			t.Error("Duration without 's' should fail")
		}
		
		// Invalid duration - decimal format
		decimalDuration := Duration("10.5s")
		if err := decimalDuration.Validate(); err == nil {
			t.Error("Decimal duration should fail")
		}
	})
	
	// Test ID validation (Step 1)
	t.Run("ID_Validation", func(t *testing.T) {
		// Valid ID
		id := ID("r123")
		if err := id.Validate(); err != nil {
			t.Errorf("Valid ID should pass: %v", err)
		}
		
		// Invalid ID - no 'r' prefix
		invalidID := ID("asset1")
		if err := invalidID.Validate(); err == nil {
			t.Error("ID without 'r' prefix should fail")
		}
	})
	
	// Test Lane validation (Step 1)
	t.Run("Lane_Validation", func(t *testing.T) {
		// Valid lanes
		validLanes := []Lane{0, 1, 5, -1, -5, 10, -10}
		for _, lane := range validLanes {
			if err := lane.Validate(); err != nil {
				t.Errorf("Valid lane %d should pass: %v", lane, err)
			}
		}
		
		// Invalid lanes
		invalidLanes := []Lane{11, -11, 15, -15}
		for _, lane := range invalidLanes {
			if err := lane.Validate(); err == nil {
				t.Errorf("Invalid lane %d should fail", lane)
			}
		}
	})
	
	// Test Media Type validation (Step 3)
	t.Run("Media_Type_Validation", func(t *testing.T) {
		// Valid image asset
		imageAsset := &Asset{
			ID:       "r1",
			Duration: "0s", // Must be 0s for images
			HasVideo: "1",
			MediaRep: MediaRep{Src: "file:///test.png"},
		}
		validator := NewMediaConstraintValidator()
		if err := validator.ValidateAsset(imageAsset, MediaTypeImage); err != nil {
			t.Errorf("Valid image asset should pass: %v", err)
		}
		
		// Invalid image asset - non-zero duration
		invalidImageAsset := &Asset{
			ID:       "r1",
			Duration: "240240/24000s", // Invalid for images
			HasVideo: "1",
			MediaRep: MediaRep{Src: "file:///test.png"},
		}
		if err := validator.ValidateAsset(invalidImageAsset, MediaTypeImage); err == nil {
			t.Error("Image asset with non-zero duration should fail")
		}
	})
	
	// Test Keyframe validation (Step 8)
	t.Run("Keyframe_Validation", func(t *testing.T) {
		// Valid position keyframe (no curve/interp)
		posValidator := NewKeyframeValidator()
		validPositionKeyframe, _ := NewValidatedKeyframe(Time("0s"), "0 0", "", "")
		if err := posValidator.ValidateKeyframe("position", validPositionKeyframe); err != nil {
			t.Errorf("Valid position keyframe should pass: %v", err)
		}
		
		// Invalid position keyframe (has curve)
		invalidPositionKeyframe, _ := NewValidatedKeyframe(Time("0s"), "0 0", "", "linear")
		if err := posValidator.ValidateKeyframe("position", invalidPositionKeyframe); err == nil {
			t.Error("Position keyframe with curve should fail")
		}
	})
}

// TestValidationFrameworkIntegration tests the full framework integration
func TestValidationFrameworkIntegration(t *testing.T) {
	vt := NewValidationTester()
	
	// Add all test types
	vt.AddCriticalValidationTests()
	vt.AddPerformanceValidationTests()
	vt.AddEdgeCaseValidationTests()
	
	totalTests := len(vt.testCases)
	if totalTests == 0 {
		t.Fatal("No validation tests were added")
	}
	
	t.Logf("Running %d total validation tests...", totalTests)
	
	// Run a subset of tests to verify framework works
	// (We don't run all tests here as that's done in the main test)
	testCount := 0
	passedCount := 0
	
	for i, testCase := range vt.testCases {
		if i >= 10 { // Only test first 10 to avoid long test times
			break
		}
		
		testCount++
		result := vt.runSingleTest(t, testCase)
		
		if result.Status == TestPassed {
			passedCount++
		}
		
		t.Logf("Test %d: %s - %s", i+1, testCase.Name, result.Message)
	}
	
	if testCount == 0 {
		t.Error("No tests were actually run")
	}
	
	passRate := float64(passedCount) / float64(testCount) * 100
	t.Logf("Sample test pass rate: %.1f%% (%d/%d)", passRate, passedCount, testCount)
	
	if passRate < 80.0 {
		t.Errorf("Sample test pass rate too low: %.1f%% (expected >= 80%%)", passRate)
	}
}

// BenchmarkValidationTestingFramework benchmarks the validation framework
func BenchmarkValidationTestingFramework(b *testing.B) {
	vt := NewValidationTester()
	vt.AddCriticalValidationTests()
	
	// Benchmark just the test setup execution
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, testCase := range vt.testCases {
			_ = testCase.Setup(vt)
		}
	}
}

// TestValidationFrameworkDocumentation verifies the framework is properly documented
func TestValidationFrameworkDocumentation(t *testing.T) {
	vt := NewValidationTester()
	vt.AddCriticalValidationTests()
	
	undocumentedTests := 0
	for _, test := range vt.testCases {
		if test.Description == "" {
			undocumentedTests++
			t.Logf("Test %s has no description", test.Name)
		}
		
		if test.Category == "" {
			t.Errorf("Test %s has no category", test.Name)
		}
	}
	
	if undocumentedTests > 0 {
		t.Logf("Warning: %d tests have no description", undocumentedTests)
	}
	
	t.Logf("Documentation check complete for %d tests", len(vt.testCases))
}