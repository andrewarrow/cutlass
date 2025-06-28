package fcp

import (
	"fmt"
	"sync"
)

// SafeResourceTransaction provides comprehensive validation and atomic resource creation
// This implements Step 13 of the FCPXMLKit-inspired refactoring plan:
// Enhanced transaction system with comprehensive validation
type SafeResourceTransaction struct {
	registry    *ReferenceRegistry
	reserved    []ID
	created     []ResourceWrapper
	rolled      bool
	committed   bool
	validator   *StructValidator
	mu          sync.Mutex
}

// ResourceWrapper provides a common interface for all resource types
type ResourceWrapper interface {
	GetID() string
	GetResourceType() string
	Validate() error
}

// SafeAssetWrapper wraps an asset with validation capabilities
type SafeAssetWrapper struct {
	Asset     *Asset
	Format    *Format
	validator *StructValidator
}

func (saw *SafeAssetWrapper) GetID() string {
	return saw.Asset.ID
}

func (saw *SafeAssetWrapper) GetResourceType() string {
	return "asset"
}

func (saw *SafeAssetWrapper) Validate() error {
	if err := saw.validator.validateAssetStructure(saw.Asset); err != nil {
		return fmt.Errorf("asset validation failed: %v", err)
	}
	if saw.Format != nil {
		if err := saw.validator.validateFormatStructure(saw.Format); err != nil {
			return fmt.Errorf("format validation failed: %v", err)
		}
	}
	return nil
}

// SafeEffectWrapper wraps an effect with validation capabilities
type SafeEffectWrapper struct {
	Effect    *Effect
	validator *StructValidator
}

func (sew *SafeEffectWrapper) GetID() string {
	return sew.Effect.ID
}

func (sew *SafeEffectWrapper) GetResourceType() string {
	return "effect"
}

func (sew *SafeEffectWrapper) Validate() error {
	return sew.validator.validateEffectStructure(sew.Effect)
}

// NewSafeTransaction creates a new safe resource transaction
func NewSafeTransaction(registry *ReferenceRegistry) *SafeResourceTransaction {
	return &SafeResourceTransaction{
		registry:  registry,
		reserved:  make([]ID, 0),
		created:   make([]ResourceWrapper, 0),
		validator: NewStructValidator(),
	}
}

// ReserveIDs safely reserves multiple IDs for this transaction
func (tx *SafeResourceTransaction) ReserveIDs(count int) []ID {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.rolled || tx.committed {
		return nil
	}

	// Reserve IDs from registry
	rawIDs := tx.registry.ReserveIDs(count)
	ids := make([]ID, len(rawIDs))
	
	for i, rawID := range rawIDs {
		id := ID(rawID)
		if err := id.Validate(); err != nil {
			// This should never happen if registry is working correctly
			panic(fmt.Sprintf("Registry generated invalid ID: %v", err))
		}
		ids[i] = id
	}

	tx.reserved = append(tx.reserved, ids...)
	return ids
}

// CreateValidatedAsset creates a validated asset using the AssetBuilder pattern
func (tx *SafeResourceTransaction) CreateValidatedAsset(filePath, name string, mediaType string) (*Asset, *Format, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.rolled {
		return nil, nil, fmt.Errorf("transaction has been rolled back")
	}
	if tx.committed {
		return nil, nil, fmt.Errorf("transaction has already been committed")
	}

	// Reserve IDs safely
	ids := tx.reserveIDsInternal(2) // Asset + Format
	assetID := ids[0]
	formatID := ids[1]

	// Create asset using AssetBuilder
	builder := NewAssetBuilder(tx.registry, mediaType)
	asset, format, err := builder.CreateAsset(assetID, filePath, name, Duration(detectMediaDuration(filePath, mediaType)))
	if err != nil {
		return nil, nil, fmt.Errorf("asset creation failed: %v", err)
	}

	// Set format reference if format was created
	if format != nil {
		format.ID = string(formatID) // Ensure format uses reserved ID
		asset.Format = format.ID
	}

	// Create wrapped resource for transaction tracking
	wrapper := &SafeAssetWrapper{
		Asset:     asset,
		Format:    format,
		validator: tx.validator,
	}

	// Validate the wrapper
	if err := wrapper.Validate(); err != nil {
		return nil, nil, fmt.Errorf("resource validation failed: %v", err)
	}

	// Add to transaction
	tx.created = append(tx.created, wrapper)

	return asset, format, nil
}

// CreateValidatedEffect creates a validated effect using the EffectBuilder pattern
func (tx *SafeResourceTransaction) CreateValidatedEffect(name, uid string) (*Effect, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}
	if tx.committed {
		return nil, fmt.Errorf("transaction has already been committed")
	}

	// Reserve ID safely
	ids := tx.reserveIDsInternal(1)
	effectID := ids[0]

	// Create effect using EffectBuilder
	builder := NewEffectBuilder()
	effect, err := builder.CreateEffect(effectID, name, uid)
	if err != nil {
		return nil, fmt.Errorf("effect creation failed: %v", err)
	}

	// Create wrapped resource for transaction tracking
	wrapper := &SafeEffectWrapper{
		Effect:    effect,
		validator: tx.validator,
	}

	// Validate the wrapper
	if err := wrapper.Validate(); err != nil {
		return nil, fmt.Errorf("effect validation failed: %v", err)
	}

	// Add to transaction
	tx.created = append(tx.created, wrapper)

	return effect, nil
}

// CreateImageAsset creates a validated image asset
func (tx *SafeResourceTransaction) CreateImageAsset(filePath, name string, duration Duration) (*Asset, *Format, error) {
	return tx.CreateValidatedAsset(filePath, name, "image")
}

// CreateVideoAsset creates a validated video asset  
func (tx *SafeResourceTransaction) CreateVideoAsset(filePath, name string, duration Duration) (*Asset, *Format, error) {
	return tx.CreateValidatedAsset(filePath, name, "video")
}

// CreateAudioAsset creates a validated audio asset
func (tx *SafeResourceTransaction) CreateAudioAsset(filePath, name string, duration Duration) (*Asset, *Format, error) {
	return tx.CreateValidatedAsset(filePath, name, "audio")
}

// CreateEffect creates a validated effect with given ID and properties
func (tx *SafeResourceTransaction) CreateEffect(id, name, uid string) (*Effect, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}
	if tx.committed {
		return nil, fmt.Errorf("transaction has already been committed")
	}

	// Create effect using EffectBuilder with specific ID
	builder := NewEffectBuilder()
	effect, err := builder.CreateEffect(ID(id), name, uid)
	if err != nil {
		return nil, fmt.Errorf("effect creation failed: %v", err)
	}

	// Create wrapped resource for transaction tracking
	wrapper := &SafeEffectWrapper{
		Effect:    effect,
		validator: tx.validator,
	}

	// Validate the wrapper
	if err := wrapper.Validate(); err != nil {
		return nil, fmt.Errorf("effect validation failed: %v", err)
	}

	// Add to transaction
	tx.created = append(tx.created, wrapper)

	return effect, nil
}

// CreateFormat creates a validated format
func (tx *SafeResourceTransaction) CreateFormat(id, name, width, height, colorSpace string) (*Format, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.rolled {
		return nil, fmt.Errorf("transaction has been rolled back")
	}
	if tx.committed {
		return nil, fmt.Errorf("transaction has already been committed")
	}

	// Create format
	format := &Format{
		ID:         id,
		Name:       name,
		Width:      width,
		Height:     height,
		ColorSpace: colorSpace,
	}

	// Validate format structure
	if err := tx.validator.validateFormatStructure(format); err != nil {
		return nil, fmt.Errorf("format validation failed: %v", err)
	}

	return format, nil
}

// Commit atomically commits all resources in the transaction
func (tx *SafeResourceTransaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.rolled {
		return fmt.Errorf("transaction has been rolled back")
	}
	if tx.committed {
		return fmt.Errorf("transaction has already been committed")
	}

	// Validate all created resources as a group
	for i, resource := range tx.created {
		if err := resource.Validate(); err != nil {
			return fmt.Errorf("resource %d validation failed: %v", i, err)
		}
	}

	// Check for ID conflicts
	for _, resource := range tx.created {
		if err := tx.registry.CheckIDConflict(ID(resource.GetID())); err != nil {
			return fmt.Errorf("ID conflict for %s: %v", resource.GetID(), err)
		}
	}

	// Validate resource relationships (e.g., asset-format relationships)
	if err := tx.validateResourceRelationships(); err != nil {
		return fmt.Errorf("resource relationship validation failed: %v", err)
	}

	// Register all resources atomically
	for _, resource := range tx.created {
		switch r := resource.(type) {
		case *SafeAssetWrapper:
			if err := tx.registry.RegisterAsset(r.Asset); err != nil {
				return fmt.Errorf("failed to register asset: %v", err)
			}
			if r.Format != nil {
				if err := tx.registry.RegisterFormat(r.Format); err != nil {
					return fmt.Errorf("failed to register format: %v", err)
				}
			}
		case *SafeEffectWrapper:
			if err := tx.registry.RegisterEffect(r.Effect); err != nil {
				return fmt.Errorf("failed to register effect: %v", err)
			}
		default:
			return fmt.Errorf("unknown resource type: %T", resource)
		}
	}

	tx.committed = true
	return nil
}

// Rollback releases all reserved IDs and discards created resources
func (tx *SafeResourceTransaction) Rollback() {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.committed {
		return // Cannot rollback after commit
	}

	// Release all reserved IDs
	for _, id := range tx.reserved {
		tx.registry.ReleaseID(string(id))
	}

	// Clear all created resources
	tx.reserved = nil
	tx.created = nil
	tx.rolled = true
}

// GetCreatedResources returns all resources created in this transaction
func (tx *SafeResourceTransaction) GetCreatedResources() []ResourceWrapper {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	// Return a copy to prevent external modification
	result := make([]ResourceWrapper, len(tx.created))
	copy(result, tx.created)
	return result
}

// GetResourceByID finds a resource by ID within this transaction
func (tx *SafeResourceTransaction) GetResourceByID(id string) ResourceWrapper {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	for _, resource := range tx.created {
		if resource.GetID() == id {
			return resource
		}
	}
	return nil
}

// ValidateAll validates all resources in the transaction without committing
func (tx *SafeResourceTransaction) ValidateAll() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.rolled {
		return fmt.Errorf("transaction has been rolled back")
	}

	// Validate all created resources
	for i, resource := range tx.created {
		if err := resource.Validate(); err != nil {
			return fmt.Errorf("resource %d validation failed: %v", i, err)
		}
	}

	// Validate resource relationships
	if err := tx.validateResourceRelationships(); err != nil {
		return fmt.Errorf("resource relationship validation failed: %v", err)
	}

	return nil
}

// Internal helper methods

// reserveIDsInternal reserves IDs without locking (assumes caller has lock)
func (tx *SafeResourceTransaction) reserveIDsInternal(count int) []ID {
	rawIDs := tx.registry.ReserveIDs(count)
	ids := make([]ID, len(rawIDs))
	
	for i, rawID := range rawIDs {
		id := ID(rawID)
		ids[i] = id
	}

	tx.reserved = append(tx.reserved, ids...)
	return ids
}

// validateResourceRelationships validates relationships between resources
func (tx *SafeResourceTransaction) validateResourceRelationships() error {
	// Track format IDs and asset format references
	formatIDs := make(map[string]bool)
	assetFormatRefs := make(map[string]string) // assetID -> formatID

	// Collect format IDs and asset format references
	for _, resource := range tx.created {
		switch r := resource.(type) {
		case *SafeAssetWrapper:
			if r.Format != nil {
				formatIDs[r.Format.ID] = true
				assetFormatRefs[r.Asset.ID] = r.Asset.Format
			}
		}
	}

	// Validate that all asset format references point to existing formats
	for assetID, formatRef := range assetFormatRefs {
		if formatRef != "" && !formatIDs[formatRef] {
			// Check if format exists in registry
			if !tx.registry.HasFormat(formatRef) {
				return fmt.Errorf("asset %s references non-existent format %s", assetID, formatRef)
			}
		}
	}

	return nil
}

// detectMediaDuration detects the duration of a media file based on type
func detectMediaDuration(filePath, mediaType string) string {
	switch mediaType {
	case "image":
		return "0s" // Images are timeless
	case "video", "audio":
		// For now, return a default duration
		// In a real implementation, this would use ffprobe or similar
		return "240240/24000s" // 10 seconds at 23.976fps
	default:
		return "0s"
	}
}

// TransactionBuilder provides a fluent interface for building transactions
type TransactionBuilder struct {
	registry *ReferenceRegistry
	tx       *SafeResourceTransaction
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder(registry *ReferenceRegistry) *TransactionBuilder {
	return &TransactionBuilder{
		registry: registry,
		tx:       NewSafeTransaction(registry),
	}
}

// AddAsset adds an asset to the transaction
func (tb *TransactionBuilder) AddAsset(filePath, name, mediaType string) *TransactionBuilder {
	_, _, err := tb.tx.CreateValidatedAsset(filePath, name, mediaType)
	if err != nil {
		// In a real implementation, might want to collect errors rather than panic
		panic(fmt.Sprintf("Failed to add asset: %v", err))
	}
	return tb
}

// AddEffect adds an effect to the transaction
func (tb *TransactionBuilder) AddEffect(name, uid string) *TransactionBuilder {
	_, err := tb.tx.CreateValidatedEffect(name, uid)
	if err != nil {
		panic(fmt.Sprintf("Failed to add effect: %v", err))
	}
	return tb
}

// Build commits the transaction and returns the created resources
func (tb *TransactionBuilder) Build() ([]ResourceWrapper, error) {
	if err := tb.tx.Commit(); err != nil {
		tb.tx.Rollback()
		return nil, fmt.Errorf("transaction commit failed: %v", err)
	}
	return tb.tx.GetCreatedResources(), nil
}

// BuildWithRollback builds the transaction with automatic rollback on error
func (tb *TransactionBuilder) BuildWithRollback() ([]ResourceWrapper, error) {
	defer func() {
		if !tb.tx.committed {
			tb.tx.Rollback()
		}
	}()

	if err := tb.tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %v", err)
	}
	return tb.tx.GetCreatedResources(), nil
}