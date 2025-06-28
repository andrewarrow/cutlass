package fcp

import (
	"strings"
	"testing"
)

func TestDTDValidationIntegration(t *testing.T) {
	tests := []struct {
		name           string
		setupFCPXML    func() *FCPXML
		expectValid    bool
		expectDTDTest  bool
	}{
		{
			name: "Valid FCPXML Structure",
			setupFCPXML: func() *FCPXML {
				return createValidTestFCPXML(t)
			},
			expectValid:   true,
			expectDTDTest: true,
		},
		{
			name: "Invalid FCPXML Structure",
			setupFCPXML: func() *FCPXML {
				fcpxml := createValidTestFCPXML(t)
				// Make it invalid by removing required version
				fcpxml.Version = ""
				return fcpxml
			},
			expectValid:   false,
			expectDTDTest: false,
		},
		{
			name: "FCPXML with Missing Resources",
			setupFCPXML: func() *FCPXML {
				fcpxml := createValidTestFCPXML(t)
				// Create dangling reference
				fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.AssetClips = []AssetClip{
					{
						Ref:      "nonexistent-asset",
						Offset:   "0s",
						Duration: "240240/24000s",
						Name:     "Broken",
					},
				}
				return fcpxml
			},
			expectValid:   false,
			expectDTDTest: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fcpxml := test.setupFCPXML()

			// Test structure validation first
			data, err := fcpxml.ValidateAndMarshal()
			
			if test.expectValid {
				if err != nil {
					t.Errorf("Expected valid FCPXML, got error: %v", err)
					return
				}
				
				// Test basic XML structure
				if !strings.Contains(string(data), "<fcpxml") {
					t.Error("Generated XML missing fcpxml root element")
				}
				
				if !strings.Contains(string(data), "version=") {
					t.Error("Generated XML missing version attribute")
				}
			} else {
				if err == nil {
					t.Error("Expected validation error, got none")
					return
				}
			}

			// Test DTD validation if xmllint is available and structure is valid
			if test.expectDTDTest && test.expectValid && IsXMLLintAvailable() {
				dtdValidator := NewDTDValidator("")
				
				// Test built-in validation rules
				if err := dtdValidator.ValidateWithBuiltinRules(data); err != nil {
					t.Errorf("Built-in validation failed: %v", err)
				}
				
				// Test validation summary
				summary := dtdValidator.GetValidationSummary()
				if !summary.XMLLintAvailable {
					t.Skip("xmllint not available, skipping DTD validation tests")
				}
			}
		})
	}
}

func TestValidateWithSummary(t *testing.T) {
	fcpxml := createValidTestFCPXML(t)
	
	// Test validation with summary
	data, report, err := fcpxml.ValidateWithSummary("")
	
	if err != nil {
		t.Errorf("ValidateWithSummary failed: %v", err)
		return
	}
	
	if data == nil {
		t.Error("ValidateWithSummary returned nil data")
	}
	
	// Check report structure
	if !report.StructureValidation {
		t.Error("Expected structure validation to pass")
	}
	
	if !report.XMLValidation {
		t.Error("Expected XML validation to pass")
	}
	
	if !report.IsValid() {
		t.Errorf("Expected report to be valid, got: %s", report.String())
	}
	
	// Test report string output
	reportStr := report.String()
	if !strings.Contains(reportStr, "FCPXML Validation Report") {
		t.Error("Report string missing title")
	}
	
	if !strings.Contains(reportStr, "PASSED") {
		t.Error("Report string should contain PASSED status")
	}
}

func TestValidateWithSummaryInvalidXML(t *testing.T) {
	fcpxml := createValidTestFCPXML(t)
	// Make it invalid
	fcpxml.Version = ""
	
	// Test validation with summary for invalid XML
	_, report, err := fcpxml.ValidateWithSummary("")
	
	if err == nil {
		t.Error("Expected validation error for invalid FCPXML")
		return
	}
	
	// Check report structure
	if report.StructureValidation {
		t.Error("Expected structure validation to fail")
	}
	
	if report.IsValid() {
		t.Error("Expected report to be invalid")
	}
	
	if len(report.Errors) == 0 {
		t.Error("Expected validation errors in report")
	}
	
	// Test report string output
	reportStr := report.String()
	if !strings.Contains(reportStr, "FAILED") {
		t.Error("Report string should contain FAILED status")
	}
	
	if !strings.Contains(reportStr, "Errors:") {
		t.Error("Report string should contain errors section")
	}
}

func TestComprehensiveValidation(t *testing.T) {
	if !IsXMLLintAvailable() {
		t.Skip("xmllint not available, skipping comprehensive validation tests")
	}
	
	fcpxml := createValidTestFCPXML(t)
	
	// Test comprehensive validation with empty paths (should use built-in rules)
	data, err := fcpxml.ValidateAndMarshalWithComprehensiveValidation("", "", "")
	
	if err != nil {
		t.Errorf("Comprehensive validation failed: %v", err)
		return
	}
	
	if data == nil {
		t.Error("Comprehensive validation returned nil data")
	}
	
	// Verify the generated XML
	if !strings.Contains(string(data), "<fcpxml") {
		t.Error("Generated XML missing fcpxml root element")
	}
}

func TestDTDValidatorBuiltinRules(t *testing.T) {
	validator := NewDTDValidator("")
	
	tests := []struct {
		name      string
		xmlData   string
		expectErr bool
	}{
		{
			name: "Valid FCPXML Structure",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.13">
	<resources>
		<asset id="r1" name="test" uid="test-uid" start="0s" duration="240240/24000s"/>
	</resources>
	<library>
		<event name="Test Event">
			<project name="Test Project">
				<sequence format="r1" duration="240240/24000s">
					<spine>
						<asset-clip ref="r1" offset="0s" duration="240240/24000s"/>
					</spine>
				</sequence>
			</project>
		</event>
	</library>
</fcpxml>`,
			expectErr: false,
		},
		{
			name: "Missing FCPXML Root",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<invalid>
	<test/>
</invalid>`,
			expectErr: true,
		},
		{
			name: "Missing Required Elements",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.13">
	<invalid/>
</fcpxml>`,
			expectErr: true,
		},
		{
			name: "Unbalanced Tags",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.13">
	<resources>
		<asset id="r1"
	</resources>
</fcpxml>`,
			expectErr: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validator.ValidateWithBuiltinRules([]byte(test.xmlData))
			
			if test.expectErr && err == nil {
				t.Error("Expected validation error, got none")
			} else if !test.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestValidationSummary(t *testing.T) {
	// Test with empty DTD path
	validator := NewDTDValidator("")
	summary := validator.GetValidationSummary()
	
	if summary.DTDPath != "" {
		t.Error("Expected empty DTD path")
	}
	
	if summary.DTDAvailable {
		t.Error("Expected DTD not available with empty path")
	}
	
	// Test summary string
	summaryStr := summary.String()
	if !strings.Contains(summaryStr, "Built-in Rules") {
		t.Error("Summary should mention built-in rules")
	}
	
	// Test with non-existent DTD path
	validator2 := NewDTDValidator("/nonexistent/path.dtd")
	summary2 := validator2.GetValidationSummary()
	
	if summary2.DTDAvailable {
		t.Error("Expected DTD not available with non-existent path")
	}
}

func TestXMLLintAvailability(t *testing.T) {
	available := IsXMLLintAvailable()
	
	// This test just verifies the function doesn't panic
	// The actual availability depends on the system
	t.Logf("xmllint available: %v", available)
	
	if available {
		t.Log("xmllint is available - DTD validation tests will run")
	} else {
		t.Log("xmllint is not available - DTD validation tests will be skipped")
	}
}

// Helper function to create a valid test FCPXML
func createValidTestFCPXML(t *testing.T) *FCPXML {
	t.Helper()
	
	fcpxml := &FCPXML{
		Version: "1.13",
		Resources: Resources{
			Assets: []Asset{
				{
					ID:       "r1",
					Name:     "test-asset",
					UID:      "test-uid-123",
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
					ID:     "r2",
					Name:   "FFVideoFormat1080p30",
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
									Format:   "r2",
									Duration: "240240/24000s",
									TCStart:  "0s",
									Spine: Spine{
										AssetClips: []AssetClip{
											{
												Ref:      "r1",
												Offset:   "0s",
												Duration: "240240/24000s",
												Name:     "Test Clip",
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
	
	return fcpxml
}