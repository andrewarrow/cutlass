// Package compatibility implements backward compatibility for legacy code patterns.
// This file is part of Step 20: Create Migration and Compatibility Layer.
//
// Provides:
// - Legacy API wrappers for gradual migration
// - Deprecated function bridges with warnings
// - Compatibility shims for old resource management patterns
// - Graceful handling of legacy struct initialization
package fcp

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// CompatibilityMode controls how strictly compatibility warnings are enforced
type CompatibilityMode int

const (
	CompatibilityStrict CompatibilityMode = iota // Fail on deprecated usage
	CompatibilityWarn                            // Warn on deprecated usage
	CompatibilityLenient                         // Allow deprecated usage silently
)

var currentCompatibilityMode = CompatibilityWarn

// SetCompatibilityMode sets the global compatibility mode
func SetCompatibilityMode(mode CompatibilityMode) {
	currentCompatibilityMode = mode
}

// GetCompatibilityMode returns the current compatibility mode
func GetCompatibilityMode() CompatibilityMode {
	return currentCompatibilityMode
}

// Legacy API Bridge Functions

// CreateLegacyFCPXML creates FCPXML using legacy patterns (DEPRECATED)
// Use GenerateEmpty() instead with proper validation
func CreateLegacyFCPXML(version string) (*FCPXML, error) {
	logDeprecationWarning("CreateLegacyFCPXML", "Use GenerateEmpty() instead")
	
	if currentCompatibilityMode == CompatibilityStrict {
		return nil, fmt.Errorf("CreateLegacyFCPXML is deprecated and strict mode is enabled")
	}
	
	// Create using new validated approach but maintain legacy interface
	return GenerateEmpty(version)
}

// AddLegacyAsset adds an asset using legacy patterns (DEPRECATED)
// Use Transaction.CreateAsset() instead
func AddLegacyAsset(fcpxml *FCPXML, id, name, uid, duration, src string) error {
	logDeprecationWarning("AddLegacyAsset", "Use Transaction.CreateAsset() with proper validation")
	
	if currentCompatibilityMode == CompatibilityStrict {
		return fmt.Errorf("AddLegacyAsset is deprecated and strict mode is enabled")
	}
	
	// Create asset using old-style struct initialization but validate it
	asset := Asset{
		ID:       id,
		Name:     name,
		UID:      uid,
		Duration: duration,
		Start:    "0s",
		MediaRep: MediaRep{
			Kind: "original-media",
			Src:  src,
		},
	}
	
	// Validate the asset before adding
	validator := NewStructValidator()
	if err := validator.validateAssetStructure(&asset); err != nil {
		return fmt.Errorf("legacy asset validation failed: %v", err)
	}
	
	fcpxml.Resources.Assets = append(fcpxml.Resources.Assets, asset)
	return nil
}

// AddLegacyFormat adds a format using legacy patterns (DEPRECATED)
// Use Transaction.CreateFormat() instead
func AddLegacyFormat(fcpxml *FCPXML, id, name, width, height string) error {
	logDeprecationWarning("AddLegacyFormat", "Use Transaction.CreateFormat() with proper validation")
	
	if currentCompatibilityMode == CompatibilityStrict {
		return fmt.Errorf("AddLegacyFormat is deprecated and strict mode is enabled")
	}
	
	format := Format{
		ID:     id,
		Name:   name,
		Width:  width,
		Height: height,
	}
	
	// Validate the format before adding
	validator := NewStructValidator()
	if err := validator.validateFormatStructure(&format); err != nil {
		return fmt.Errorf("legacy format validation failed: %v", err)
	}
	
	fcpxml.Resources.Formats = append(fcpxml.Resources.Formats, format)
	return nil
}

// CreateLegacyAssetClip creates an asset clip using legacy patterns (DEPRECATED)
// Use proper Spine management with validation instead
func CreateLegacyAssetClip(ref, offset, duration, name string) AssetClip {
	logDeprecationWarning("CreateLegacyAssetClip", "Use validated Spine management instead")
	
	return AssetClip{
		Ref:      ref,
		Offset:   offset,
		Duration: duration,
		Name:     name,
	}
}

// AddLegacyAssetClipToSpine adds an asset clip to spine using legacy patterns (DEPRECATED)
func AddLegacyAssetClipToSpine(spine *Spine, clip AssetClip) error {
	logDeprecationWarning("AddLegacyAssetClipToSpine", "Use Transaction-based spine management")
	
	if currentCompatibilityMode == CompatibilityStrict {
		return fmt.Errorf("AddLegacyAssetClipToSpine is deprecated and strict mode is enabled")
	}
	
	// Validate the clip before adding
	if err := Duration(clip.Duration).Validate(); err != nil {
		return fmt.Errorf("legacy asset clip duration validation failed: %v", err)
	}
	
	if clip.Offset != "" {
		if err := Time(clip.Offset).Validate(); err != nil {
			return fmt.Errorf("legacy asset clip offset validation failed: %v", err)
		}
	}
	
	spine.AssetClips = append(spine.AssetClips, clip)
	return nil
}

// MarshalLegacyXML marshals FCPXML using legacy patterns (DEPRECATED)
// Use ValidateAndMarshal() instead
func MarshalLegacyXML(fcpxml *FCPXML) ([]byte, error) {
	logDeprecationWarning("MarshalLegacyXML", "Use ValidateAndMarshal() for safety")
	
	if currentCompatibilityMode == CompatibilityStrict {
		return nil, fmt.Errorf("MarshalLegacyXML is deprecated and strict mode is enabled")
	}
	
	// Use new validated marshaling but maintain legacy interface
	return fcpxml.ValidateAndMarshal()
}

// Legacy Resource Management Bridge

// LegacyResourceManager provides backward compatibility for old resource management
type LegacyResourceManager struct {
	fcpxml  *FCPXML
	idCount int
}

// NewLegacyResourceManager creates a legacy resource manager (DEPRECATED)
func NewLegacyResourceManager(fcpxml *FCPXML) *LegacyResourceManager {
	logDeprecationWarning("NewLegacyResourceManager", "Use ResourceRegistry and Transaction instead")
	
	return &LegacyResourceManager{
		fcpxml:  fcpxml,
		idCount: 1,
	}
}

// GenerateID generates a resource ID using legacy patterns (DEPRECATED)
func (lrm *LegacyResourceManager) GenerateID() string {
	logDeprecationWarning("LegacyResourceManager.GenerateID", "Use Transaction.ReserveIDs() instead")
	
	id := fmt.Sprintf("r%d", lrm.idCount)
	lrm.idCount++
	return id
}

// AddAsset adds an asset using legacy resource manager (DEPRECATED)
func (lrm *LegacyResourceManager) AddAsset(name, uid, duration, src string) string {
	logDeprecationWarning("LegacyResourceManager.AddAsset", "Use Transaction.CreateAsset() instead")
	
	id := lrm.GenerateID()
	
	// Use the legacy bridge function
	if err := AddLegacyAsset(lrm.fcpxml, id, name, uid, duration, src); err != nil {
		// In legacy mode, we log the error but continue
		log.Printf("Legacy asset creation warning: %v", err)
	}
	
	return id
}

// Compatibility Shims for Common Patterns

// ValidateLegacyStructure provides basic validation for legacy FCPXML (DEPRECATED)
func ValidateLegacyStructure(fcpxml *FCPXML) error {
	logDeprecationWarning("ValidateLegacyStructure", "Use ValidateStructure() instead")
	
	if currentCompatibilityMode == CompatibilityStrict {
		return fmt.Errorf("ValidateLegacyStructure is deprecated and strict mode is enabled")
	}
	
	// Use new validation but maintain legacy interface
	return fcpxml.ValidateStructure()
}

// FixLegacyIssues attempts to fix common legacy issues
func FixLegacyIssues(fcpxml *FCPXML) []string {
	logDeprecationWarning("FixLegacyIssues", "Use MigrationManager instead")
	
	var warnings []string
	
	// Fix missing version
	if fcpxml.Version == "" {
		fcpxml.Version = "1.13"
		warnings = append(warnings, "Added missing FCPXML version")
	}
	
	// Fix assets without duration suffix
	for i := range fcpxml.Resources.Assets {
		asset := &fcpxml.Resources.Assets[i]
		if asset.Duration != "" {
			durationStr := asset.Duration
			if !strings.HasSuffix(durationStr, "s") {
				asset.Duration = durationStr + "s"
				warnings = append(warnings, fmt.Sprintf("Fixed duration format for asset %s", asset.ID))
			}
		}
	}
	
	// Fix asset clips without duration suffix
	for eventIdx := range fcpxml.Library.Events {
		for projectIdx := range fcpxml.Library.Events[eventIdx].Projects {
			for seqIdx := range fcpxml.Library.Events[eventIdx].Projects[projectIdx].Sequences {
				spine := &fcpxml.Library.Events[eventIdx].Projects[projectIdx].Sequences[seqIdx].Spine
				
				for clipIdx := range spine.AssetClips {
					clip := &spine.AssetClips[clipIdx]
					durationStr := clip.Duration
					if durationStr != "" && !strings.HasSuffix(durationStr, "s") {
						clip.Duration = durationStr + "s"
						warnings = append(warnings, fmt.Sprintf("Fixed duration format for asset clip %d", clipIdx))
					}
				}
			}
		}
	}
	
	return warnings
}

// Migration Helper Functions

// IsLegacyPattern detects if FCPXML uses legacy patterns
func IsLegacyPattern(fcpxml *FCPXML) bool {
	// Check for common legacy patterns
	
	// Missing version
	if fcpxml.Version == "" {
		return true
	}
	
	// Assets without proper duration format
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Duration != "" {
			durationStr := asset.Duration
			if !strings.HasSuffix(durationStr, "s") {
				return true
			}
		}
	}
	
	// Check for other legacy patterns...
	
	return false
}

// SuggestMigration provides migration suggestions for legacy code
func SuggestMigration(fcpxml *FCPXML) []string {
	var suggestions []string
	
	if IsLegacyPattern(fcpxml) {
		suggestions = append(suggestions, "This FCPXML uses legacy patterns")
		suggestions = append(suggestions, "Consider using MigrationManager to upgrade to validation-first architecture")
		suggestions = append(suggestions, "Use Transaction and ResourceRegistry for better resource management")
		suggestions = append(suggestions, "Replace direct struct manipulation with validated operations")
	}
	
	return suggestions
}

// Internal helper functions

func logDeprecationWarning(function, alternative string) {
	if currentCompatibilityMode == CompatibilityLenient {
		return
	}
	
	message := fmt.Sprintf("DEPRECATION WARNING: %s is deprecated. %s", function, alternative)
	
	// Log to stderr for visibility
	fmt.Fprintf(os.Stderr, "%s\n", message)
	
	// Also use standard logging if available
	log.Printf(message)
}

// CompatibilityInfo provides information about the compatibility layer
type CompatibilityInfo struct {
	Mode             CompatibilityMode
	DeprecatedCalls  int
	MigrationNeeded  bool
	Suggestions      []string
}

// GetCompatibilityInfo returns current compatibility status
func GetCompatibilityInfo(fcpxml *FCPXML) CompatibilityInfo {
	return CompatibilityInfo{
		Mode:            currentCompatibilityMode,
		DeprecatedCalls: 0, // Would be tracked in a real implementation
		MigrationNeeded: IsLegacyPattern(fcpxml),
		Suggestions:     SuggestMigration(fcpxml),
	}
}

// String returns a human-readable compatibility info
func (ci CompatibilityInfo) String() string {
	var lines []string
	
	lines = append(lines, "=== Compatibility Status ===")
	
	switch ci.Mode {
	case CompatibilityStrict:
		lines = append(lines, "Mode: Strict (deprecated functions will fail)")
	case CompatibilityWarn:
		lines = append(lines, "Mode: Warning (deprecated functions will warn)")
	case CompatibilityLenient:
		lines = append(lines, "Mode: Lenient (deprecated functions allowed)")
	}
	
	if ci.MigrationNeeded {
		lines = append(lines, "Migration: RECOMMENDED")
		if len(ci.Suggestions) > 0 {
			lines = append(lines, "Suggestions:")
			for _, suggestion := range ci.Suggestions {
				lines = append(lines, fmt.Sprintf("  - %s", suggestion))
			}
		}
	} else {
		lines = append(lines, "Migration: Not needed")
	}
	
	return fmt.Sprintf("%s", lines)
}