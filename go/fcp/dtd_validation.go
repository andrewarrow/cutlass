package fcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DTDValidator provides DTD validation integration (optional but recommended)
type DTDValidator struct {
	dtdPath string
}

// NewDTDValidator creates a new DTD validator
func NewDTDValidator(dtdPath string) *DTDValidator {
	return &DTDValidator{dtdPath: dtdPath}
}

// ValidateXML validates XML data against the DTD using xmllint
func (dtd *DTDValidator) ValidateXML(xmlData []byte) error {
	// Write XML to temporary file
	tmpFile, err := os.CreateTemp("", "fcpxml_*.xml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(xmlData); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Run xmllint with DTD validation
	cmd := exec.Command("xmllint", "--dtdvalid", dtd.dtdPath, "--noout", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("DTD validation failed: %s", string(output))
	}

	return nil
}

// ValidateXMLWithSchema validates XML data against an XSD schema using xmllint
func (dtd *DTDValidator) ValidateXMLWithSchema(xmlData []byte, schemaPath string) error {
	// Write XML to temporary file
	tmpFile, err := os.CreateTemp("", "fcpxml_*.xml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(xmlData); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Run xmllint with schema validation
	cmd := exec.Command("xmllint", "--schema", schemaPath, "--noout", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("schema validation failed: %s", string(output))
	}

	return nil
}

// IsXMLLintAvailable checks if xmllint is available on the system
func IsXMLLintAvailable() bool {
	cmd := exec.Command("xmllint", "--version")
	err := cmd.Run()
	return err == nil
}

// ValidateXMLWithRelaxNG validates XML data against a RelaxNG schema
func (dtd *DTDValidator) ValidateXMLWithRelaxNG(xmlData []byte, rngPath string) error {
	// Write XML to temporary file
	tmpFile, err := os.CreateTemp("", "fcpxml_*.xml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(xmlData); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Run xmllint with RelaxNG validation
	cmd := exec.Command("xmllint", "--relaxng", rngPath, "--noout", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("RelaxNG validation failed: %s", string(output))
	}

	return nil
}

// ValidateWithBuiltinRules validates FCPXML against built-in validation rules
func (dtd *DTDValidator) ValidateWithBuiltinRules(xmlData []byte) error {
	// Check for common FCPXML structure issues
	xmlStr := string(xmlData)
	
	// Ensure root element is fcpxml
	if !strings.Contains(xmlStr, "<fcpxml") {
		return fmt.Errorf("missing fcpxml root element")
	}
	
	// Check for required structure
	requiredElements := []string{"resources", "library", "event", "project", "sequence", "spine"}
	for _, element := range requiredElements {
		if !strings.Contains(xmlStr, "<"+element) {
			return fmt.Errorf("missing required element: %s", element)
		}
	}
	
	// Check for basic XML well-formedness (simplified check)
	// Count opening tags, closing tags, and self-closing tags
	totalOpenTags := strings.Count(xmlStr, "<")
	totalCloseTags := strings.Count(xmlStr, "</")
	selfClosingTags := strings.Count(xmlStr, "/>")
	declarationTags := strings.Count(xmlStr, "<?") // XML declarations like <?xml
	
	// Basic balance check: opening tags should roughly match closing + self-closing
	// This is a simplified heuristic, not a full XML parser
	expectedOpenTags := totalCloseTags + selfClosingTags + declarationTags
	if totalOpenTags < expectedOpenTags {
		return fmt.Errorf("unbalanced XML tags detected: insufficient opening tags")
	}
	
	return nil
}

// ValidateCompatibility validates XML against multiple validation methods
func (dtd *DTDValidator) ValidateCompatibility(xmlData []byte, dtdPath, xsdPath, rngPath string) error {
	// First check built-in rules
	if err := dtd.ValidateWithBuiltinRules(xmlData); err != nil {
		return fmt.Errorf("built-in validation failed: %v", err)
	}
	
	// Try DTD validation if available
	if dtdPath != "" && dtd.isFileAccessible(dtdPath) {
		if err := dtd.ValidateXML(xmlData); err != nil {
			return fmt.Errorf("DTD validation failed: %v", err)
		}
	}
	
	// Try XSD validation if available
	if xsdPath != "" && dtd.isFileAccessible(xsdPath) {
		if err := dtd.ValidateXMLWithSchema(xmlData, xsdPath); err != nil {
			return fmt.Errorf("XSD validation failed: %v", err)
		}
	}
	
	// Try RelaxNG validation if available
	if rngPath != "" && dtd.isFileAccessible(rngPath) {
		if err := dtd.ValidateXMLWithRelaxNG(xmlData, rngPath); err != nil {
			return fmt.Errorf("RelaxNG validation failed: %v", err)
		}
	}
	
	return nil
}

// isFileAccessible checks if a file exists and is readable
func (dtd *DTDValidator) isFileAccessible(path string) bool {
	if path == "" {
		return false
	}
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	
	// Check if file is readable
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	file.Close()
	
	return true
}

// GetValidationSummary returns a summary of available validation methods
func (dtd *DTDValidator) GetValidationSummary() ValidationSummary {
	summary := ValidationSummary{
		XMLLintAvailable: IsXMLLintAvailable(),
		DTDPath:         dtd.dtdPath,
		DTDAvailable:    dtd.isFileAccessible(dtd.dtdPath),
		BuiltinRules:    true, // Always available
	}
	
	return summary
}

// ValidationSummary provides information about available validation methods
type ValidationSummary struct {
	XMLLintAvailable bool
	DTDPath         string
	DTDAvailable    bool
	BuiltinRules    bool // Always true
}

// String returns a human-readable summary
func (vs ValidationSummary) String() string {
	var methods []string
	
	if vs.BuiltinRules {
		methods = append(methods, "Built-in Rules")
	}
	
	if vs.XMLLintAvailable {
		methods = append(methods, "xmllint")
	}
	
	if vs.DTDAvailable {
		methods = append(methods, fmt.Sprintf("DTD (%s)", filepath.Base(vs.DTDPath)))
	}
	
	if len(methods) == 0 {
		return "No validation methods available"
	}
	
	return fmt.Sprintf("Available validation: %s", strings.Join(methods, ", "))
}