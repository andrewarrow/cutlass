package fcp

import (
	"strings"
	"testing"
)

func TestMigrationManager_MigrateFromLegacy(t *testing.T) {
	tests := []struct {
		name          string
		setupLegacy   func() *FCPXML
		sourceVersion string
		targetVersion string
		expectError   bool
		expectWarnings bool
	}{
		{
			name: "Valid Legacy FCPXML Migration",
			setupLegacy: func() *FCPXML {
				return &FCPXML{
					Version: "1.11",
					Resources: Resources{
						Assets: []Asset{
							{
								ID:       "r1",
								Name:     "test-asset",
								UID:      "test-uid",
								Duration: "240240/24000s",
								Start:    "0s",
								MediaRep: MediaRep{
									Kind: "original-media",
									Src:  "file:///test.mp4",
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
			},
			sourceVersion:  "1.11",
			targetVersion:  "1.13",
			expectError:    false,
			expectWarnings: true,
		},
		{
			name: "Invalid Legacy FCPXML - Missing Asset ID",
			setupLegacy: func() *FCPXML {
				return &FCPXML{
					Version: "1.11",
					Resources: Resources{
						Assets: []Asset{
							{
								// Missing ID
								Name:     "test-asset",
								UID:      "test-uid",
								Duration: "240240/24000s",
								Start:    "0s",
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
												TCStart:  "0s",
												Spine:    Spine{},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			sourceVersion:  "1.11",
			targetVersion:  "1.13",
			expectError:    true,
			expectWarnings: false,
		},
		{
			name: "Legacy FCPXML with Missing Duration Suffix",
			setupLegacy: func() *FCPXML {
				return &FCPXML{
					Version: "1.12",
					Resources: Resources{
						Assets: []Asset{
							{
								ID:       "r1",
								Name:     "test-asset",
								UID:      "test-uid",
								Duration: "240240/24000", // Missing 's' suffix
								Start:    "0s",
								MediaRep: MediaRep{
									Kind: "original-media",
									Src:  "file:///test.mp4",
								},
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
												TCStart:  "0s",
												Spine: Spine{
													AssetClips: []AssetClip{
														{
															Ref:      "r1",
															Offset:   "0s",
															Duration: "240240/24000", // Missing 's' suffix
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
			},
			sourceVersion:  "1.12",
			targetVersion:  "1.13",
			expectError:    false,
			expectWarnings: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			legacyFCPXML := test.setupLegacy()
			mm := NewMigrationManager(test.sourceVersion, test.targetVersion)

			migratedFCPXML, err := mm.MigrateFromLegacy(legacyFCPXML)

			if test.expectError && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if !test.expectError {
				// Validate migrated structure
				if migratedFCPXML == nil {
					t.Error("Expected migrated FCPXML but got nil")
					return
				}

				if migratedFCPXML.Version != test.targetVersion {
					t.Errorf("Expected target version %s, got %s", test.targetVersion, migratedFCPXML.Version)
				}

				// Check migration report
				report := mm.GetMigrationReport()
				if test.expectWarnings && len(report.Warnings) == 0 {
					t.Error("Expected warnings but got none")
				}

				if !test.expectWarnings && len(report.Warnings) > 0 {
					t.Errorf("Expected no warnings but got: %v", report.Warnings)
				}
			}
		})
	}
}

func TestMigrateFromXML(t *testing.T) {
	legacyXML := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.11">
	<resources>
		<asset id="r1" name="test-asset" uid="test-uid" start="0s" duration="240240/24000s" hasVideo="1">
			<media-rep kind="original-media" src="file:///test.mp4"/>
		</asset>
		<format id="r2" name="FFVideoFormat1080p30" width="1920" height="1080"/>
	</resources>
	<library>
		<event name="Test Event">
			<project name="Test Project">
				<sequence format="r2" duration="240240/24000s" tcStart="0s">
					<spine>
						<asset-clip ref="r1" offset="0s" duration="240240/24000s" name="Test Clip"/>
					</spine>
				</sequence>
			</project>
		</event>
	</library>
</fcpxml>`

	migratedFCPXML, report, err := MigrateFromXML([]byte(legacyXML), "1.13")

	if err != nil {
		t.Errorf("Migration failed: %v", err)
		return
	}

	if migratedFCPXML == nil {
		t.Error("Expected migrated FCPXML but got nil")
		return
	}

	if migratedFCPXML.Version != "1.13" {
		t.Errorf("Expected version 1.13, got %s", migratedFCPXML.Version)
	}

	if !report.Success {
		t.Errorf("Expected successful migration, got errors: %v", report.Errors)
	}

	// Validate the migrated FCPXML can be marshaled
	data, err := migratedFCPXML.ValidateAndMarshal()
	if err != nil {
		t.Errorf("Failed to marshal migrated FCPXML: %v", err)
	}

	if !strings.Contains(string(data), `version="1.13"`) {
		t.Error("Migrated XML should contain version 1.13")
	}
}

func TestBatchMigration(t *testing.T) {
	legacyXML1 := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.11">
	<resources>
		<asset id="r1" name="asset1" uid="uid1" start="0s" duration="240240/24000s">
			<media-rep kind="original-media" src="file:///test1.mp4"/>
		</asset>
	</resources>
	<library>
		<event name="Event1">
			<project name="Project1">
				<sequence duration="240240/24000s" tcStart="0s">
					<spine>
						<asset-clip ref="r1" offset="0s" duration="240240/24000s"/>
					</spine>
				</sequence>
			</project>
		</event>
	</library>
</fcpxml>`

	legacyXML2 := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.12">
	<resources>
		<asset id="r2" name="asset2" uid="uid2" start="0s" duration="480480/24000s">
			<media-rep kind="original-media" src="file:///test2.mp4"/>
		</asset>
	</resources>
	<library>
		<event name="Event2">
			<project name="Project2">
				<sequence duration="480480/24000s" tcStart="0s">
					<spine>
						<asset-clip ref="r2" offset="0s" duration="480480/24000s"/>
					</spine>
				</sequence>
			</project>
		</event>
	</library>
</fcpxml>`

	// Invalid XML for testing error handling
	invalidXML := `<?xml version="1.0" encoding="UTF-8"?>
<fcpxml version="1.11">
	<resources>
		<asset name="missing-id" uid="uid3" start="0s" duration="240240/24000s">
			<media-rep kind="original-media" src="file:///test3.mp4"/>
		</asset>
	</resources>
</fcpxml>`

	bm := NewBatchMigration()
	bm.AddFile([]byte(legacyXML1), "1.13")
	bm.AddFile([]byte(legacyXML2), "1.13")
	bm.AddFile([]byte(invalidXML), "1.13")

	batchReport := bm.GetBatchReport()

	if batchReport.TotalFiles != 3 {
		t.Errorf("Expected 3 total files, got %d", batchReport.TotalFiles)
	}

	if batchReport.SuccessfulFiles != 2 {
		t.Errorf("Expected 2 successful files, got %d", batchReport.SuccessfulFiles)
	}

	if batchReport.FailedFiles != 1 {
		t.Errorf("Expected 1 failed file, got %d", batchReport.FailedFiles)
	}

	// Test batch report string
	reportStr := batchReport.String()
	if !strings.Contains(reportStr, "Total Files: 3") {
		t.Error("Batch report should contain total files count")
	}

	if !strings.Contains(reportStr, "Successful: 2") {
		t.Error("Batch report should contain successful files count")
	}

	if !strings.Contains(reportStr, "Failed: 1") {
		t.Error("Batch report should contain failed files count")
	}
}

func TestMigrationManager_FixMethods(t *testing.T) {
	mm := NewMigrationManager("1.11", "1.13")

	// Test fixDurationFormat
	t.Run("Fix Duration Format", func(t *testing.T) {
		// Test adding missing 's' suffix
		duration := "240240/24000"
		err := mm.fixDurationFormat(&duration)
		if err != nil {
			t.Errorf("Failed to fix duration format: %v", err)
		}
		if duration != "240240/24000s" {
			t.Errorf("Expected '240240/24000s', got '%s'", duration)
		}

		// Test empty duration
		emptyDuration := ""
		err = mm.fixDurationFormat(&emptyDuration)
		if err != nil {
			t.Errorf("Failed to fix empty duration: %v", err)
		}
		if emptyDuration != "240240/24000s" {
			t.Errorf("Expected default duration, got '%s'", emptyDuration)
		}

		// Test already correct duration
		correctDuration := "480480/24000s"
		err = mm.fixDurationFormat(&correctDuration)
		if err != nil {
			t.Errorf("Failed to handle correct duration: %v", err)
		}
		if correctDuration != "480480/24000s" {
			t.Errorf("Correct duration should not change, got '%s'", correctDuration)
		}
	})

	// Test fixTimeFormat
	t.Run("Fix Time Format", func(t *testing.T) {
		// Test adding missing 's' suffix
		time := "120120/24000"
		err := mm.fixTimeFormat(&time)
		if err != nil {
			t.Errorf("Failed to fix time format: %v", err)
		}
		if time != "120120/24000s" {
			t.Errorf("Expected '120120/24000s', got '%s'", time)
		}

		// Test empty time
		emptyTime := ""
		err = mm.fixTimeFormat(&emptyTime)
		if err != nil {
			t.Errorf("Failed to fix empty time: %v", err)
		}
		if emptyTime != "0s" {
			t.Errorf("Expected '0s', got '%s'", emptyTime)
		}
	})

	// Test fixAssetIssues
	t.Run("Fix Asset Issues", func(t *testing.T) {
		// Test asset with missing fields
		asset := Asset{
			ID: "r1",
			// Missing Name, UID, Duration
		}

		err := mm.fixAssetIssues(&asset)
		if err != nil {
			t.Errorf("Failed to fix asset issues: %v", err)
		}

		if asset.Name == "" {
			t.Error("Asset name should be set to default")
		}

		if asset.UID == "" {
			t.Error("Asset UID should be generated")
		}

		if asset.Duration == "" {
			t.Error("Asset duration should be set to default")
		}

		// Test asset without ID (should fail)
		assetNoID := Asset{}
		err = mm.fixAssetIssues(&assetNoID)
		if err == nil {
			t.Error("Expected error for asset without ID")
		}
	})
}

func TestMigrationReport_String(t *testing.T) {
	report := MigrationReport{
		SourceVersion: "1.11",
		TargetVersion: "1.13",
		Warnings:      []string{"Warning 1", "Warning 2"},
		Errors:        []string{"Error 1"},
		Success:       false,
	}

	reportStr := report.String()

	// Check required sections
	if !strings.Contains(reportStr, "Migration Report") {
		t.Error("Report should contain title")
	}

	if !strings.Contains(reportStr, "Source Version: 1.11") {
		t.Error("Report should contain source version")
	}

	if !strings.Contains(reportStr, "Target Version: 1.13") {
		t.Error("Report should contain target version")
	}

	if !strings.Contains(reportStr, "Status: FAILED") {
		t.Error("Report should show failed status")
	}

	if !strings.Contains(reportStr, "Warnings:") {
		t.Error("Report should contain warnings section")
	}

	if !strings.Contains(reportStr, "Warning 1") {
		t.Error("Report should contain specific warnings")
	}

	if !strings.Contains(reportStr, "Errors:") {
		t.Error("Report should contain errors section")
	}

	if !strings.Contains(reportStr, "Error 1") {
		t.Error("Report should contain specific errors")
	}
}

func TestMigrationManager_VersionHandling(t *testing.T) {
	tests := []struct {
		name           string
		sourceVersion  string
		targetVersion  string
		expectWarnings int
	}{
		{
			name:           "1.11 to 1.13 upgrade",
			sourceVersion:  "1.11",
			targetVersion:  "1.13",
			expectWarnings: 1, // Should warn about major upgrade
		},
		{
			name:           "1.12 to 1.13 upgrade",
			sourceVersion:  "1.12",
			targetVersion:  "1.13",
			expectWarnings: 1, // Should warn about upgrade
		},
		{
			name:           "1.13 to 1.12 downgrade",
			sourceVersion:  "1.13",
			targetVersion:  "1.12",
			expectWarnings: 1, // Should warn about downgrade
		},
		{
			name:           "Same version",
			sourceVersion:  "1.13",
			targetVersion:  "1.13",
			expectWarnings: 0, // No warnings for same version
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mm := NewMigrationManager(test.sourceVersion, test.targetVersion)
			
			// Create minimal valid FCPXML
			legacyFCPXML := &FCPXML{
				Version: test.sourceVersion,
				Resources: Resources{},
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
											TCStart:  "0s",
											Spine:    Spine{},
										},
									},
								},
							},
						},
					},
				},
			}

			_, err := mm.MigrateFromLegacy(legacyFCPXML)
			if err != nil {
				t.Errorf("Migration failed: %v", err)
				return
			}

			report := mm.GetMigrationReport()
			if len(report.Warnings) != test.expectWarnings {
				t.Errorf("Expected %d warnings, got %d: %v", test.expectWarnings, len(report.Warnings), report.Warnings)
			}
		})
	}
}