package fcp

import (
	"fmt"
	"os"
	"os/exec"
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