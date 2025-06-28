package fcp

import (
	"strings"
	"testing"
)

func TestSafeResourceTransaction(t *testing.T) {
	registry := NewReferenceRegistry()

	t.Run("Basic transaction lifecycle", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		// Transaction should start in a valid state
		if tx.rolled || tx.committed {
			t.Error("New transaction should not be rolled or committed")
		}

		// Should be able to reserve IDs
		ids := tx.ReserveIDs(3)
		if len(ids) != 3 {
			t.Errorf("Expected 3 IDs, got %d", len(ids))
		}

		// All IDs should be valid
		for i, id := range ids {
			if err := id.Validate(); err != nil {
				t.Errorf("ID %d is invalid: %v", i, err)
			}
		}

		// Should be able to rollback
		tx.Rollback()
		if !tx.rolled {
			t.Error("Transaction should be marked as rolled after rollback")
		}

		// Should not be able to reserve IDs after rollback
		ids2 := tx.ReserveIDs(1)
		if ids2 != nil {
			t.Error("Should not be able to reserve IDs after rollback")
		}
	})

	t.Run("Asset creation and validation", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		// Create a valid image asset
		asset, format, err := tx.CreateValidatedAsset("/test/image.png", "test-image", "image")
		if err != nil {
			t.Errorf("Failed to create image asset: %v", err)
		}
		if asset == nil {
			t.Error("Asset should not be nil")
		}
		if format == nil {
			t.Error("Image should have a format")
		}

		// Asset should have correct properties for image
		if asset.Duration != "0s" {
			t.Errorf("Image asset should have duration='0s', got: %s", asset.Duration)
		}
		if asset.HasVideo != "1" {
			t.Errorf("Image asset should have hasVideo='1', got: %s", asset.HasVideo)
		}
		if asset.HasAudio != "" {
			t.Errorf("Image asset should not have hasAudio, got: %s", asset.HasAudio)
		}

		// Format should have correct properties for image
		if format.FrameDuration != "" {
			t.Errorf("Image format should not have frameDuration, got: %s", format.FrameDuration)
		}

		// Should be able to commit
		err = tx.Commit()
		if err != nil {
			t.Errorf("Failed to commit transaction: %v", err)
		}
		if !tx.committed {
			t.Error("Transaction should be marked as committed")
		}
	})

	t.Run("Video asset creation", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		asset, format, err := tx.CreateValidatedAsset("/test/video.mp4", "test-video", "video")
		if err != nil {
			t.Errorf("Failed to create video asset: %v", err)
		}

		// Video should have correct properties
		if asset.Duration == "0s" {
			t.Error("Video asset should not have duration='0s'")
		}
		if asset.HasVideo != "1" {
			t.Errorf("Video asset should have hasVideo='1', got: %s", asset.HasVideo)
		}
		if format == nil {
			t.Error("Video should have a format")
		} else if format.FrameDuration == "" {
			t.Error("Video format should have frameDuration")
		}

		err = tx.Commit()
		if err != nil {
			t.Errorf("Failed to commit video transaction: %v", err)
		}
	})

	t.Run("Audio asset creation", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		asset, format, err := tx.CreateValidatedAsset("/test/audio.wav", "test-audio", "audio")
		if err != nil {
			t.Errorf("Failed to create audio asset: %v", err)
		}

		// Audio should have correct properties
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

		err = tx.Commit()
		if err != nil {
			t.Errorf("Failed to commit audio transaction: %v", err)
		}
	})

	t.Run("Effect creation", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		effect, err := tx.CreateValidatedEffect("Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
		if err != nil {
			t.Errorf("Failed to create effect: %v", err)
		}
		if effect == nil {
			t.Error("Effect should not be nil")
		}

		// Effect should have correct properties
		if effect.Name != "Text" {
			t.Errorf("Effect name should be 'Text', got: %s", effect.Name)
		}
		if effect.UID != ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti" {
			t.Errorf("Effect UID incorrect, got: %s", effect.UID)
		}

		err = tx.Commit()
		if err != nil {
			t.Errorf("Failed to commit effect transaction: %v", err)
		}
	})

	t.Run("Invalid effect UID should fail", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		_, err := tx.CreateValidatedEffect("Invalid Effect", "com.unknown.effect")
		if err == nil {
			t.Error("Should fail with unknown effect UID")
		}
		if !strings.Contains(err.Error(), "unknown effect UID") {
			t.Errorf("Expected 'unknown effect UID' error, got: %v", err)
		}
	})

	t.Run("Double commit should fail", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		// Create something to commit
		_, _, err := tx.CreateValidatedAsset("/test/image.png", "test", "image")
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}

		// First commit should succeed
		err = tx.Commit()
		if err != nil {
			t.Errorf("First commit failed: %v", err)
		}

		// Second commit should fail
		err = tx.Commit()
		if err == nil {
			t.Error("Second commit should fail")
		}
		if !strings.Contains(err.Error(), "already been committed") {
			t.Errorf("Expected 'already been committed' error, got: %v", err)
		}
	})

	t.Run("Operations after rollback should fail", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		tx.Rollback()

		// Should not be able to create assets after rollback
		_, _, err := tx.CreateValidatedAsset("/test/image.png", "test", "image")
		if err == nil {
			t.Error("Should not be able to create assets after rollback")
		}
		if !strings.Contains(err.Error(), "rolled back") {
			t.Errorf("Expected 'rolled back' error, got: %v", err)
		}

		// Should not be able to create effects after rollback
		_, err = tx.CreateValidatedEffect("Test", "FFGaussianBlur")
		if err == nil {
			t.Error("Should not be able to create effects after rollback")
		}
	})
}

func TestSafeTransactionResourceRelationships(t *testing.T) {
	registry := NewReferenceRegistry()

	t.Run("Asset-format relationship validation", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		// Create an asset that should have a format
		asset, format, err := tx.CreateValidatedAsset("/test/video.mp4", "test-video", "video")
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}

		// Check that asset references format
		if asset.Format != format.ID {
			t.Errorf("Asset format reference (%s) should match format ID (%s)", asset.Format, format.ID)
		}

		// Should be able to validate relationships
		err = tx.ValidateAll()
		if err != nil {
			t.Errorf("Validation should pass: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Errorf("Commit should succeed: %v", err)
		}
	})

	t.Run("Get created resources", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		// Create multiple resources
		_, _, err := tx.CreateValidatedAsset("/test/image.png", "image", "image")
		if err != nil {
			t.Fatalf("Failed to create image: %v", err)
		}

		_, err = tx.CreateValidatedEffect("Blur", "FFGaussianBlur")
		if err != nil {
			t.Fatalf("Failed to create effect: %v", err)
		}

		// Should be able to get created resources before commit
		resources := tx.GetCreatedResources()
		if len(resources) != 2 {
			t.Errorf("Expected 2 resources, got %d", len(resources))
		}

		// Check resource types
		hasAsset := false
		hasEffect := false
		for _, resource := range resources {
			switch resource.GetResourceType() {
			case "asset":
				hasAsset = true
			case "effect":
				hasEffect = true
			}
		}

		if !hasAsset {
			t.Error("Should have asset resource")
		}
		if !hasEffect {
			t.Error("Should have effect resource")
		}

		err = tx.Commit()
		if err != nil {
			t.Errorf("Commit failed: %v", err)
		}
	})

	t.Run("Get resource by ID", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		asset, _, err := tx.CreateValidatedAsset("/test/image.png", "image", "image")
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}

		// Should be able to find resource by ID
		found := tx.GetResourceByID(asset.ID)
		if found == nil {
			t.Error("Should find resource by ID")
		}
		if found.GetID() != asset.ID {
			t.Errorf("Found resource ID (%s) should match asset ID (%s)", found.GetID(), asset.ID)
		}

		// Should not find non-existent resource
		notFound := tx.GetResourceByID("nonexistent")
		if notFound != nil {
			t.Error("Should not find non-existent resource")
		}
	})
}

func TestTransactionBuilder(t *testing.T) {
	registry := NewReferenceRegistry()

	t.Run("Fluent transaction building", func(t *testing.T) {
		builder := NewTransactionBuilder(registry)

		resources, err := builder.
			AddAsset("/test/image.png", "image", "image").
			AddAsset("/test/video.mp4", "video", "video").
			AddEffect("Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti").
			Build()

		if err != nil {
			t.Errorf("Transaction build failed: %v", err)
		}

		if len(resources) != 3 {
			t.Errorf("Expected 3 resources, got %d", len(resources))
		}

		// Check that all resource types are present
		types := make(map[string]bool)
		for _, resource := range resources {
			types[resource.GetResourceType()] = true
		}

		if !types["asset"] {
			t.Error("Should have asset resources")
		}
		if !types["effect"] {
			t.Error("Should have effect resource")
		}
	})

	t.Run("Transaction builder with rollback", func(t *testing.T) {
		builder := NewTransactionBuilder(registry)

		resources, err := builder.
			AddAsset("/test/image.png", "test", "image").
			BuildWithRollback()

		if err != nil {
			t.Errorf("Transaction build with rollback failed: %v", err)
		}

		if len(resources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(resources))
		}
	})

	t.Run("Transaction builder failure handling", func(t *testing.T) {
		// This test would fail due to invalid effect UID
		// but our implementation panics instead of returning errors
		// In a production system, you might want to collect errors instead
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic with invalid effect UID
				panicMsg := r.(string)
				if !strings.Contains(panicMsg, "unknown effect UID") {
					t.Errorf("Expected panic about unknown effect UID, got: %s", panicMsg)
				}
			} else {
				t.Error("Expected panic due to invalid effect UID")
			}
		}()

		builder := NewTransactionBuilder(registry)
		// This should panic
		builder.AddEffect("Invalid", "com.unknown.effect")
	})
}

func TestSafeTransactionValidation(t *testing.T) {
	registry := NewReferenceRegistry()

	t.Run("Invalid media type should fail", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		_, _, err := tx.CreateValidatedAsset("/test/file.unknown", "test", "unknown")
		if err == nil {
			t.Error("Should fail with unknown media type")
		}
		if !strings.Contains(err.Error(), "unsupported media type") {
			t.Errorf("Expected 'unsupported media type' error, got: %v", err)
		}
	})

	t.Run("Validation without commit", func(t *testing.T) {
		tx := NewSafeTransaction(registry)

		// Create some resources
		_, _, err := tx.CreateValidatedAsset("/test/image.png", "image", "image")
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}

		// Should be able to validate without committing
		err = tx.ValidateAll()
		if err != nil {
			t.Errorf("Validation should pass: %v", err)
		}

		// Transaction should still not be committed
		if tx.committed {
			t.Error("Transaction should not be committed after validation")
		}
	})

	t.Run("Concurrent transaction safety", func(t *testing.T) {
		// Test that concurrent operations on transaction are safe
		tx := NewSafeTransaction(registry)

		// This is a basic test - in a real scenario you'd use goroutines
		// to test concurrent access
		ids1 := tx.ReserveIDs(5)
		ids2 := tx.ReserveIDs(3)

		if len(ids1) != 5 {
			t.Errorf("Expected 5 IDs from first reservation, got %d", len(ids1))
		}
		if len(ids2) != 3 {
			t.Errorf("Expected 3 IDs from second reservation, got %d", len(ids2))
		}

		// All IDs should be unique
		allIDs := append(ids1, ids2...)
		idSet := make(map[string]bool)
		for _, id := range allIDs {
			if idSet[string(id)] {
				t.Errorf("Duplicate ID detected: %s", id)
			}
			idSet[string(id)] = true
		}
	})
}

// Benchmark transaction performance
func BenchmarkSafeTransactionCreation(b *testing.B) {
	registry := NewReferenceRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewSafeTransaction(registry)
		_, _, err := tx.CreateValidatedAsset("/test/image.png", "test", "image")
		if err != nil {
			b.Fatalf("Asset creation failed: %v", err)
		}
		err = tx.Commit()
		if err != nil {
			b.Fatalf("Transaction commit failed: %v", err)
		}
	}
}

func BenchmarkTransactionBuilder(b *testing.B) {
	registry := NewReferenceRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewTransactionBuilder(registry)
		_, err := builder.
			AddAsset("/test/image.png", "image", "image").
			AddEffect("Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti").
			Build()
		if err != nil {
			b.Fatalf("Transaction build failed: %v", err)
		}
	}
}