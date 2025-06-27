// Package migration implements Step 20 of the FCPXMLKit-inspired refactoring plan:
// Create Migration and Compatibility Layer - Provide smooth transition from legacy
// code patterns to the new validation-first architecture.
//
// This layer provides:
// - Version compatibility for different FCPXML versions (1.11, 1.12, 1.13)
// - Legacy API bridge for gradual migration
// - Migration tools for existing codebases
// - Backward compatibility for deprecated patterns
package fcp

import (
	"fmt"
	"strings"
	"encoding/xml"
)

// MigrationManager handles migration from legacy patterns to validation-first architecture
type MigrationManager struct {
	sourceVersion string
	targetVersion string
	warnings      []string
	errors        []string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(sourceVersion, targetVersion string) *MigrationManager {
	return &MigrationManager{
		sourceVersion: sourceVersion,
		targetVersion: targetVersion,
		warnings:      make([]string, 0),
		errors:        make([]string, 0),
	}
}

// MigrateFromLegacy migrates legacy FCPXML struct to validation-first structure
func (mm *MigrationManager) MigrateFromLegacy(legacyFCPXML *FCPXML) (*FCPXML, error) {
	// Create a new FCPXML with validation-first patterns
	migratedFCPXML := &FCPXML{
		Version: mm.targetVersion,
		Resources: Resources{
			Assets:  make([]Asset, 0),
			Formats: make([]Format, 0),
			Effects: make([]Effect, 0),
		},
		Library: Library{
			Events: make([]Event, 0),
		},
	}

	// Migrate version if necessary
	if err := mm.migrateVersion(legacyFCPXML, migratedFCPXML); err != nil {
		return nil, fmt.Errorf("version migration failed: %v", err)
	}

	// Migrate resources with validation
	if err := mm.migrateResources(legacyFCPXML, migratedFCPXML); err != nil {
		return nil, fmt.Errorf("resource migration failed: %v", err)
	}

	// Migrate library structure with validation
	if err := mm.migrateLibrary(legacyFCPXML, migratedFCPXML); err != nil {
		return nil, fmt.Errorf("library migration failed: %v", err)
	}

	// Perform final validation on migrated structure
	if err := migratedFCPXML.ValidateStructure(); err != nil {
		return nil, fmt.Errorf("migrated FCPXML validation failed: %v", err)
	}

	return migratedFCPXML, nil
}

// migrateVersion handles version-specific migration logic
func (mm *MigrationManager) migrateVersion(legacy, migrated *FCPXML) error {
	if legacy.Version == "" {
		mm.warnings = append(mm.warnings, "Source FCPXML missing version, assuming 1.11")
		legacy.Version = "1.11"
	}

	// Set target version
	migrated.Version = mm.targetVersion

	// Version-specific migrations
	switch legacy.Version {
	case "1.11":
		if mm.targetVersion == "1.13" {
			mm.warnings = append(mm.warnings, "Migrating from 1.11 to 1.13 - some features may need manual review")
		}
	case "1.12":
		if mm.targetVersion == "1.13" {
			mm.warnings = append(mm.warnings, "Migrating from 1.12 to 1.13")
		}
	case "1.13":
		// Already current version
		if mm.targetVersion != "1.13" {
			mm.warnings = append(mm.warnings, "Downgrading from 1.13 may lose some features")
		}
	default:
		return fmt.Errorf("unsupported source version: %s", legacy.Version)
	}

	return nil
}

// migrateResources migrates resources with proper validation
func (mm *MigrationManager) migrateResources(legacy, migrated *FCPXML) error {
	validator := NewStructValidator()

	// Migrate assets
	for i, asset := range legacy.Resources.Assets {
		// Create a copy to modify
		migratedAsset := asset
		
		// Try to fix common issues first
		if err := mm.fixAssetIssues(&migratedAsset); err != nil {
			return fmt.Errorf("failed to fix asset %d: %v", i, err)
		}
		
		// Validate the fixed asset
		if err := validator.validateAssetStructure(&migratedAsset); err != nil {
			mm.warnings = append(mm.warnings, fmt.Sprintf("Asset %d validation warning after fixes: %v", i, err))
			// If still invalid after fixes, this is an error
			if !strings.Contains(err.Error(), "missing") {
				return fmt.Errorf("asset %d could not be migrated: %v", i, err)
			}
		}
		
		migrated.Resources.Assets = append(migrated.Resources.Assets, migratedAsset)
	}

	// Migrate formats
	for i, format := range legacy.Resources.Formats {
		if err := validator.validateFormatStructure(&format); err != nil {
			mm.warnings = append(mm.warnings, fmt.Sprintf("Format %d validation warning: %v", i, err))
			if err := mm.fixFormatIssues(&format); err != nil {
				return fmt.Errorf("failed to fix format %d: %v", i, err)
			}
		}
		migrated.Resources.Formats = append(migrated.Resources.Formats, format)
	}

	// Migrate effects
	for i, effect := range legacy.Resources.Effects {
		if err := validator.validateEffectStructure(&effect); err != nil {
			mm.warnings = append(mm.warnings, fmt.Sprintf("Effect %d validation warning: %v", i, err))
			if err := mm.fixEffectIssues(&effect); err != nil {
				return fmt.Errorf("failed to fix effect %d: %v", i, err)
			}
		}
		migrated.Resources.Effects = append(migrated.Resources.Effects, effect)
	}

	return nil
}

// migrateLibrary migrates library structure with validation
func (mm *MigrationManager) migrateLibrary(legacy, migrated *FCPXML) error {
	if len(legacy.Library.Events) == 0 {
		mm.warnings = append(mm.warnings, "No events found in legacy FCPXML")
		return nil
	}

	// Migrate events
	for i, event := range legacy.Library.Events {
		migratedEvent := event
		
		// Migrate projects within the event
		for j, project := range event.Projects {
			migratedProject := project
			
			// Migrate sequences within the project
			for k, sequence := range project.Sequences {
				migratedSequence := sequence
				
				// Validate and migrate spine elements
				if err := mm.migrateSpine(&sequence.Spine, &migratedSequence.Spine); err != nil {
					return fmt.Errorf("failed to migrate spine in event %d, project %d, sequence %d: %v", i, j, k, err)
				}
				
				migratedProject.Sequences[k] = migratedSequence
			}
			
			migratedEvent.Projects[j] = migratedProject
		}
		
		migrated.Library.Events = append(migrated.Library.Events, migratedEvent)
	}

	return nil
}

// migrateSpine migrates spine elements with proper validation
func (mm *MigrationManager) migrateSpine(legacy, migrated *Spine) error {
	// Migrate asset clips
	for i, clip := range legacy.AssetClips {
		// Create a copy to modify
		migratedClip := clip
		
		// Fix timing formats first
		if err := mm.fixDurationFormat(&migratedClip.Duration); err != nil {
			return fmt.Errorf("failed to fix asset clip %d duration: %v", i, err)
		}
		
		if migratedClip.Offset != "" {
			if err := mm.fixTimeFormat(&migratedClip.Offset); err != nil {
				return fmt.Errorf("failed to fix asset clip %d offset: %v", i, err)
			}
		}
		
		// Validate after fixes
		if err := Duration(migratedClip.Duration).Validate(); err != nil {
			mm.warnings = append(mm.warnings, fmt.Sprintf("Asset clip %d duration still invalid after fixes: %v", i, err))
		}
		
		if migratedClip.Offset != "" {
			if err := Time(migratedClip.Offset).Validate(); err != nil {
				mm.warnings = append(mm.warnings, fmt.Sprintf("Asset clip %d offset still invalid after fixes: %v", i, err))
			}
		}
		
		migrated.AssetClips = append(migrated.AssetClips, migratedClip)
	}

	// Migrate video elements
	for i, video := range legacy.Videos {
		// Create a copy to modify
		migratedVideo := video
		
		// Fix timing formats first
		if err := mm.fixDurationFormat(&migratedVideo.Duration); err != nil {
			return fmt.Errorf("failed to fix video %d duration: %v", i, err)
		}
		
		if migratedVideo.Offset != "" {
			if err := mm.fixTimeFormat(&migratedVideo.Offset); err != nil {
				return fmt.Errorf("failed to fix video %d offset: %v", i, err)
			}
		}
		
		migrated.Videos = append(migrated.Videos, migratedVideo)
	}

	// Migrate titles
	for i, title := range legacy.Titles {
		// Create a copy to modify
		migratedTitle := title
		
		// Fix timing formats first
		if err := mm.fixDurationFormat(&migratedTitle.Duration); err != nil {
			return fmt.Errorf("failed to fix title %d duration: %v", i, err)
		}
		
		if migratedTitle.Offset != "" {
			if err := mm.fixTimeFormat(&migratedTitle.Offset); err != nil {
				return fmt.Errorf("failed to fix title %d offset: %v", i, err)
			}
		}
		
		migrated.Titles = append(migrated.Titles, migratedTitle)
	}

	return nil
}

// Fix methods for common migration issues

func (mm *MigrationManager) fixAssetIssues(asset *Asset) error {
	if asset.ID == "" {
		return fmt.Errorf("asset missing required ID")
	}
	
	if asset.Name == "" {
		asset.Name = "Migrated Asset"
		mm.warnings = append(mm.warnings, "Asset missing name, set to default")
	}
	
	if asset.UID == "" {
		asset.UID = generateUID("asset")
		mm.warnings = append(mm.warnings, "Asset missing UID, generated new one")
	}
	
	if asset.Duration == "" {
		asset.Duration = "240240/24000s"
		mm.warnings = append(mm.warnings, "Asset missing duration, set to default 10s")
	} else {
		// Fix duration format if missing 's' suffix
		if err := mm.fixDurationFormat(&asset.Duration); err != nil {
			return err
		}
	}
	
	return nil
}

func (mm *MigrationManager) fixFormatIssues(format *Format) error {
	if format.ID == "" {
		return fmt.Errorf("format missing required ID")
	}
	
	if format.Width == "" {
		format.Width = "1920"
		mm.warnings = append(mm.warnings, "Format missing width, set to 1920")
	}
	
	if format.Height == "" {
		format.Height = "1080"
		mm.warnings = append(mm.warnings, "Format missing height, set to 1080")
	}
	
	return nil
}

func (mm *MigrationManager) fixEffectIssues(effect *Effect) error {
	if effect.ID == "" {
		return fmt.Errorf("effect missing required ID")
	}
	
	if effect.Name == "" {
		effect.Name = "Migrated Effect"
		mm.warnings = append(mm.warnings, "Effect missing name, set to default")
	}
	
	if effect.UID == "" {
		effect.UID = "FFGenericEffect"
		mm.warnings = append(mm.warnings, "Effect missing UID, set to generic")
	}
	
	return nil
}

func (mm *MigrationManager) fixDurationFormat(duration *string) error {
	if *duration == "" {
		*duration = "240240/24000s"
		return nil
	}
	
	// Check if duration already ends with 's'
	if !strings.HasSuffix(*duration, "s") {
		*duration = *duration + "s"
		mm.warnings = append(mm.warnings, "Added missing 's' suffix to duration")
	}
	
	return nil
}

func (mm *MigrationManager) fixTimeFormat(time *string) error {
	if *time == "" {
		*time = "0s"
		return nil
	}
	
	// Check if time already ends with 's'
	if !strings.HasSuffix(*time, "s") {
		*time = *time + "s"
		mm.warnings = append(mm.warnings, "Added missing 's' suffix to time")
	}
	
	return nil
}

// GetMigrationReport returns a report of the migration process
func (mm *MigrationManager) GetMigrationReport() MigrationReport {
	return MigrationReport{
		SourceVersion: mm.sourceVersion,
		TargetVersion: mm.targetVersion,
		Warnings:      mm.warnings,
		Errors:        mm.errors,
		Success:       len(mm.errors) == 0,
	}
}

// MigrationReport provides details about the migration process
type MigrationReport struct {
	SourceVersion string
	TargetVersion string
	Warnings      []string
	Errors        []string
	Success       bool
}

// String returns a human-readable migration report
func (mr MigrationReport) String() string {
	var lines []string
	
	lines = append(lines, "=== FCPXML Migration Report ===")
	lines = append(lines, fmt.Sprintf("Source Version: %s", mr.SourceVersion))
	lines = append(lines, fmt.Sprintf("Target Version: %s", mr.TargetVersion))
	lines = append(lines, fmt.Sprintf("Status: %s", boolToStatus(mr.Success)))
	
	if len(mr.Warnings) > 0 {
		lines = append(lines, "\nWarnings:")
		for _, warning := range mr.Warnings {
			lines = append(lines, fmt.Sprintf("  - %s", warning))
		}
	}
	
	if len(mr.Errors) > 0 {
		lines = append(lines, "\nErrors:")
		for _, err := range mr.Errors {
			lines = append(lines, fmt.Sprintf("  - %s", err))
		}
	}
	
	return strings.Join(lines, "\n")
}

// MigrateFromXML migrates FCPXML from raw XML data
func MigrateFromXML(xmlData []byte, targetVersion string) (*FCPXML, MigrationReport, error) {
	// Parse the XML first
	var legacyFCPXML FCPXML
	if err := xml.Unmarshal(xmlData, &legacyFCPXML); err != nil {
		return nil, MigrationReport{}, fmt.Errorf("failed to parse legacy XML: %v", err)
	}
	
	// Create migration manager
	mm := NewMigrationManager(legacyFCPXML.Version, targetVersion)
	
	// Perform migration
	migratedFCPXML, err := mm.MigrateFromLegacy(&legacyFCPXML)
	if err != nil {
		return nil, mm.GetMigrationReport(), err
	}
	
	return migratedFCPXML, mm.GetMigrationReport(), nil
}

// BatchMigration handles migration of multiple FCPXML files
type BatchMigration struct {
	reports []MigrationReport
	errors  []error
}

// NewBatchMigration creates a new batch migration handler
func NewBatchMigration() *BatchMigration {
	return &BatchMigration{
		reports: make([]MigrationReport, 0),
		errors:  make([]error, 0),
	}
}

// AddFile adds a file to the batch migration
func (bm *BatchMigration) AddFile(xmlData []byte, targetVersion string) {
	fcpxml, report, err := MigrateFromXML(xmlData, targetVersion)
	bm.reports = append(bm.reports, report)
	if err != nil {
		bm.errors = append(bm.errors, err)
	} else {
		// Successful migration - fcpxml is ready for use
		_ = fcpxml
	}
}

// GetBatchReport returns a summary of all migrations
func (bm *BatchMigration) GetBatchReport() BatchMigrationReport {
	successCount := 0
	totalWarnings := 0
	
	for _, report := range bm.reports {
		if report.Success {
			successCount++
		}
		totalWarnings += len(report.Warnings)
	}
	
	return BatchMigrationReport{
		TotalFiles:      len(bm.reports),
		SuccessfulFiles: successCount,
		FailedFiles:     len(bm.errors),
		TotalWarnings:   totalWarnings,
		Reports:         bm.reports,
		Errors:          bm.errors,
	}
}

// BatchMigrationReport provides summary of batch migration
type BatchMigrationReport struct {
	TotalFiles      int
	SuccessfulFiles int
	FailedFiles     int
	TotalWarnings   int
	Reports         []MigrationReport
	Errors          []error
}

// String returns a human-readable batch migration report
func (bmr BatchMigrationReport) String() string {
	var lines []string
	
	lines = append(lines, "=== Batch Migration Report ===")
	lines = append(lines, fmt.Sprintf("Total Files: %d", bmr.TotalFiles))
	lines = append(lines, fmt.Sprintf("Successful: %d", bmr.SuccessfulFiles))
	lines = append(lines, fmt.Sprintf("Failed: %d", bmr.FailedFiles))
	lines = append(lines, fmt.Sprintf("Total Warnings: %d", bmr.TotalWarnings))
	
	if bmr.FailedFiles > 0 {
		lines = append(lines, "\nFailed Files:")
		for i, err := range bmr.Errors {
			lines = append(lines, fmt.Sprintf("  %d. %v", i+1, err))
		}
	}
	
	return strings.Join(lines, "\n")
}