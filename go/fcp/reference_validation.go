// Package reference_validation implements Step 4 of the FCPXMLKit-inspired refactoring plan:
// Reference validation system to ensure all FCPXML references point to valid resources.
//
// This prevents the most common FCPXML errors: dangling references that cause import failures.
package fcp

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// ReferenceRegistry tracks all resources and validates references
type ReferenceRegistry struct {
	mu             sync.RWMutex
	assets         map[ID]*Asset
	formats        map[ID]*Format
	effects        map[ID]*Effect
	media          map[ID]*Media
	danglingRefs   map[string][]string // Track what references what for debugging
	nextResourceID int                 // Next ID number to generate
}

// NewReferenceRegistry creates a new reference registry
func NewReferenceRegistry() *ReferenceRegistry {
	return &ReferenceRegistry{
		assets:         make(map[ID]*Asset),
		formats:        make(map[ID]*Format),
		effects:        make(map[ID]*Effect),
		media:          make(map[ID]*Media),
		danglingRefs:   make(map[string][]string),
		nextResourceID: 1, // Start with r1
	}
}

// RegisterAsset registers an asset in the registry
func (r *ReferenceRegistry) RegisterAsset(asset *Asset) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	id := ID(asset.ID)
	if err := id.Validate(); err != nil {
		return fmt.Errorf("invalid asset ID: %v", err)
	}
	
	if _, exists := r.assets[id]; exists {
		return fmt.Errorf("duplicate asset ID: %s", id)
	}
	
	r.assets[id] = asset
	return nil
}

// RegisterFormat registers a format in the registry
func (r *ReferenceRegistry) RegisterFormat(format *Format) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	id := ID(format.ID)
	if err := id.Validate(); err != nil {
		return fmt.Errorf("invalid format ID: %v", err)
	}
	
	if _, exists := r.formats[id]; exists {
		return fmt.Errorf("duplicate format ID: %s", id)
	}
	
	r.formats[id] = format
	return nil
}

// RegisterEffect registers an effect in the registry
func (r *ReferenceRegistry) RegisterEffect(effect *Effect) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	id := ID(effect.ID)
	if err := id.Validate(); err != nil {
		return fmt.Errorf("invalid effect ID: %v", err)
	}
	
	if _, exists := r.effects[id]; exists {
		return fmt.Errorf("duplicate effect ID: %s", id)
	}
	
	r.effects[id] = effect
	return nil
}

// RegisterMedia registers a media in the registry
func (r *ReferenceRegistry) RegisterMedia(media *Media) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	id := ID(media.ID)
	if err := id.Validate(); err != nil {
		return fmt.Errorf("invalid media ID: %v", err)
	}
	
	if _, exists := r.media[id]; exists {
		return fmt.Errorf("duplicate media ID: %s", id)
	}
	
	r.media[id] = media
	return nil
}

// ValidateReference validates that a reference points to an existing resource
func (r *ReferenceRegistry) ValidateReference(ref ID, refType string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	switch refType {
	case "asset":
		if _, exists := r.assets[ref]; !exists {
			return fmt.Errorf("dangling asset reference: %s", ref)
		}
	case "format":
		if _, exists := r.formats[ref]; !exists {
			return fmt.Errorf("dangling format reference: %s", ref)
		}
	case "effect":
		if _, exists := r.effects[ref]; !exists {
			return fmt.Errorf("dangling effect reference: %s", ref)
		}
	case "media":
		if _, exists := r.media[ref]; !exists {
			return fmt.Errorf("dangling media reference: %s", ref)
		}
	default:
		return fmt.Errorf("unknown reference type: %s", refType)
	}
	
	return nil
}

// GetAsset retrieves an asset by ID
func (r *ReferenceRegistry) GetAsset(id ID) (*Asset, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	asset, exists := r.assets[id]
	return asset, exists
}

// GetFormat retrieves a format by ID
func (r *ReferenceRegistry) GetFormat(id ID) (*Format, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	format, exists := r.formats[id]
	return format, exists
}

// GetEffect retrieves an effect by ID
func (r *ReferenceRegistry) GetEffect(id ID) (*Effect, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	effect, exists := r.effects[id]
	return effect, exists
}

// GetMedia retrieves a media by ID
func (r *ReferenceRegistry) GetMedia(id ID) (*Media, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	media, exists := r.media[id]
	return media, exists
}

// ValidateAllReferences validates all references in an FCPXML document
func (r *ReferenceRegistry) ValidateAllReferences(fcpxml *FCPXML) error {
	errors := []string{}
	
	// Validate spine references
	for _, event := range fcpxml.Library.Events {
		for _, project := range event.Projects {
			for _, sequence := range project.Sequences {
				if err := r.validateSpineReferences(sequence.Spine); err != nil {
					errors = append(errors, err.Error())
				}
			}
		}
	}
	
	// Validate asset format references
	for _, asset := range fcpxml.Resources.Assets {
		if asset.Format != "" {
			if err := r.ValidateReference(ID(asset.Format), "format"); err != nil {
				errors = append(errors, fmt.Sprintf("asset %s: %v", asset.ID, err))
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("reference validation failed: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// validateSpineReferences validates all references within a spine
func (r *ReferenceRegistry) validateSpineReferences(spine Spine) error {
	errors := []string{}
	
	// Validate asset-clip references
	for i, clip := range spine.AssetClips {
		if err := r.ValidateReference(ID(clip.Ref), "asset"); err != nil {
			errors = append(errors, fmt.Sprintf("asset-clip %d: %v", i, err))
		}
		
		// Validate nested elements
		if err := r.validateNestedReferencesInAssetClip(clip); err != nil {
			errors = append(errors, fmt.Sprintf("asset-clip %d nested: %v", i, err))
		}
	}
	
	// Validate video references
	for i, video := range spine.Videos {
		if err := r.ValidateReference(ID(video.Ref), "asset"); err != nil {
			errors = append(errors, fmt.Sprintf("video %d: %v", i, err))
		}
		
		// Validate nested elements
		if err := r.validateNestedReferencesInVideo(video); err != nil {
			errors = append(errors, fmt.Sprintf("video %d nested: %v", i, err))
		}
	}
	
	// Validate title references
	for i, title := range spine.Titles {
		if err := r.ValidateReference(ID(title.Ref), "effect"); err != nil {
			errors = append(errors, fmt.Sprintf("title %d: %v", i, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("spine validation failed: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// validateNestedReferencesInAssetClip validates references within an asset clip
func (r *ReferenceRegistry) validateNestedReferencesInAssetClip(clip AssetClip) error {
	errors := []string{}
	
	// Validate nested asset clips
	for i, nested := range clip.NestedAssetClips {
		if err := r.ValidateReference(ID(nested.Ref), "asset"); err != nil {
			errors = append(errors, fmt.Sprintf("nested asset-clip %d: %v", i, err))
		}
	}
	
	// Validate nested videos
	for i, nested := range clip.Videos {
		if err := r.ValidateReference(ID(nested.Ref), "asset"); err != nil {
			errors = append(errors, fmt.Sprintf("nested video %d: %v", i, err))
		}
	}
	
	// Validate nested titles
	for i, nested := range clip.Titles {
		if err := r.ValidateReference(ID(nested.Ref), "effect"); err != nil {
			errors = append(errors, fmt.Sprintf("nested title %d: %v", i, err))
		}
	}
	
	// Validate filter-video references
	for i, filter := range clip.FilterVideos {
		if err := r.ValidateReference(ID(filter.Ref), "effect"); err != nil {
			errors = append(errors, fmt.Sprintf("filter-video %d: %v", i, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("asset-clip nested validation failed: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// validateNestedReferencesInVideo validates references within a video element
func (r *ReferenceRegistry) validateNestedReferencesInVideo(video Video) error {
	errors := []string{}
	
	// Validate nested asset clips
	for i, nested := range video.NestedAssetClips {
		if err := r.ValidateReference(ID(nested.Ref), "asset"); err != nil {
			errors = append(errors, fmt.Sprintf("nested asset-clip %d: %v", i, err))
		}
	}
	
	// Validate nested videos
	for i, nested := range video.NestedVideos {
		if err := r.ValidateReference(ID(nested.Ref), "asset"); err != nil {
			errors = append(errors, fmt.Sprintf("nested video %d: %v", i, err))
		}
	}
	
	// Validate nested titles
	for i, nested := range video.NestedTitles {
		if err := r.ValidateReference(ID(nested.Ref), "effect"); err != nil {
			errors = append(errors, fmt.Sprintf("nested title %d: %v", i, err))
		}
	}
	
	// Validate filter-video references
	for i, filter := range video.FilterVideos {
		if err := r.ValidateReference(ID(filter.Ref), "effect"); err != nil {
			errors = append(errors, fmt.Sprintf("filter-video %d: %v", i, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("video nested validation failed: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// PopulateFromFCPXML populates the registry from an existing FCPXML document
func (r *ReferenceRegistry) PopulateFromFCPXML(fcpxml *FCPXML) error {
	// Register all assets
	for i := range fcpxml.Resources.Assets {
		if err := r.RegisterAsset(&fcpxml.Resources.Assets[i]); err != nil {
			return fmt.Errorf("failed to register asset %d: %v", i, err)
		}
	}
	
	// Register all formats
	for i := range fcpxml.Resources.Formats {
		if err := r.RegisterFormat(&fcpxml.Resources.Formats[i]); err != nil {
			return fmt.Errorf("failed to register format %d: %v", i, err)
		}
	}
	
	// Register all effects
	for i := range fcpxml.Resources.Effects {
		if err := r.RegisterEffect(&fcpxml.Resources.Effects[i]); err != nil {
			return fmt.Errorf("failed to register effect %d: %v", i, err)
		}
	}
	
	// Register all media
	for i := range fcpxml.Resources.Media {
		if err := r.RegisterMedia(&fcpxml.Resources.Media[i]); err != nil {
			return fmt.Errorf("failed to register media %d: %v", i, err)
		}
	}
	
	return nil
}

// GetResourceCounts returns the count of each resource type
func (r *ReferenceRegistry) GetResourceCounts() (assets, formats, effects, media int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.assets), len(r.formats), len(r.effects), len(r.media)
}

// GetNextAvailableID returns the next available ID for resource creation
func (r *ReferenceRegistry) GetNextAvailableID() ID {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Find the highest used ID number
	maxID := 0
	
	// Check assets
	for id := range r.assets {
		if num := extractIDNumber(id); num > maxID {
			maxID = num
		}
	}
	
	// Check formats
	for id := range r.formats {
		if num := extractIDNumber(id); num > maxID {
			maxID = num
		}
	}
	
	// Check effects
	for id := range r.effects {
		if num := extractIDNumber(id); num > maxID {
			maxID = num
		}
	}
	
	// Check media
	for id := range r.media {
		if num := extractIDNumber(id); num > maxID {
			maxID = num
		}
	}
	
	// Return next available ID
	nextID, _ := NewID(maxID + 1)
	return nextID
}

// extractIDNumber extracts the numeric part from an ID (e.g., "r5" -> 5)
func extractIDNumber(id ID) int {
	idStr := string(id)
	if !strings.HasPrefix(idStr, "r") {
		return 0
	}
	
	numStr := idStr[1:]
	if num, err := strconv.Atoi(numStr); err == nil {
		return num
	}
	
	return 0
}

// CheckIDConflict checks if an ID would conflict with existing resources
func (r *ReferenceRegistry) CheckIDConflict(id ID) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Check all resource types for conflicts
	if _, exists := r.assets[id]; exists {
		return fmt.Errorf("ID conflict: %s already exists as asset", id)
	}
	if _, exists := r.formats[id]; exists {
		return fmt.Errorf("ID conflict: %s already exists as format", id)
	}
	if _, exists := r.effects[id]; exists {
		return fmt.Errorf("ID conflict: %s already exists as effect", id)
	}
	if _, exists := r.media[id]; exists {
		return fmt.Errorf("ID conflict: %s already exists as media", id)
	}
	
	return nil
}

// ReserveID reserves an ID to prevent conflicts (used by transaction system)
func (r *ReferenceRegistry) ReserveID(id ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Create placeholder entries to reserve the IDs
	// These will be replaced by actual resources during transaction commit
	key := fmt.Sprintf("reserved:%s", id)
	r.danglingRefs[key] = []string{"reserved"}
	
	return nil
}

// ReleaseReservedID releases a reserved ID (used during transaction rollback)
func (r *ReferenceRegistry) ReleaseReservedID(id ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := fmt.Sprintf("reserved:%s", id)
	delete(r.danglingRefs, key)
}

// GetDanglingReferences returns a report of all dangling references
func (r *ReferenceRegistry) GetDanglingReferences(fcpxml *FCPXML) map[string][]string {
	danglingRefs := make(map[string][]string)
	
	// This would scan the FCPXML and identify all references that don't
	// point to valid resources, useful for debugging
	// Implementation would be similar to ValidateAllReferences but collect
	// references instead of returning errors
	
	return danglingRefs
}

// ReleaseID releases a reserved ID (used when transactions are rolled back)
func (r *ReferenceRegistry) ReleaseID(id string) {
	r.ReleaseReservedID(ID(id))
}

// HasFormat checks if a format with the given ID exists
func (r *ReferenceRegistry) HasFormat(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.formats[ID(id)]
	return exists
}

// HasAsset checks if an asset with the given ID exists
func (r *ReferenceRegistry) HasAsset(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.assets[ID(id)]
	return exists
}

// HasEffect checks if an effect with the given ID exists
func (r *ReferenceRegistry) HasEffect(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.effects[ID(id)]
	return exists
}

// ReserveIDs reserves multiple IDs for atomic operations
func (r *ReferenceRegistry) ReserveIDs(count int) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate next available ID
		id := r.generateNextID()
		ids[i] = id
		
		// Reserve the ID by adding to danglingRefs
		key := fmt.Sprintf("reserved:%s", id)
		r.danglingRefs[key] = []string{"reserved"}
	}
	
	return ids
}

// generateNextID generates the next available resource ID
func (r *ReferenceRegistry) generateNextID() string {
	for {
		id := fmt.Sprintf("r%d", r.nextResourceID)
		r.nextResourceID++
		
		// Check if ID is already in use
		if !r.isIDInUse(ID(id)) {
			return id
		}
	}
}

// isIDInUse checks if an ID is already in use by any resource type
func (r *ReferenceRegistry) isIDInUse(id ID) bool {
	// Check all resource types
	if _, exists := r.assets[id]; exists {
		return true
	}
	if _, exists := r.formats[id]; exists {
		return true
	}
	if _, exists := r.effects[id]; exists {
		return true
	}
	if _, exists := r.media[id]; exists {
		return true
	}
	
	// Check reserved IDs
	key := fmt.Sprintf("reserved:%s", id)
	if _, exists := r.danglingRefs[key]; exists {
		return true
	}
	
	return false
}

// Note: nextResourceID is now a field in ReferenceRegistry struct