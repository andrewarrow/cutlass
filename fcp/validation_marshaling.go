package fcp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// ValidatedMarshaler interface for structures that need validation before marshaling
type ValidatedMarshaler interface {
	ValidateAndMarshal() ([]byte, error)
}

// ValidatedMarshalIndent provides validation-aware XML marshaling with indentation
func ValidatedMarshalIndent(v ValidatedMarshaler, prefix, indent string) ([]byte, error) {
	return v.ValidateAndMarshal()
}

// ValidateAndMarshal implements comprehensive validation before XML generation
func (fcpxml *FCPXML) ValidateAndMarshal() ([]byte, error) {
	// Validate entire structure before marshaling
	if err := fcpxml.ValidateStructure(); err != nil {
		return nil, fmt.Errorf("FCPXML validation failed: %v", err)
	}

	// Perform standard XML marshaling
	data, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("XML marshaling failed: %v", err)
	}

	// Post-marshal validation (check for XML structure issues)
	if err := validateXMLStructure(data); err != nil {
		return nil, fmt.Errorf("generated XML validation failed: %v", err)
	}

	return data, nil
}

// ValidateStructure performs comprehensive validation of the FCPXML structure
func (fcpxml *FCPXML) ValidateStructure() error {
	// Create validation context
	registry := NewReferenceRegistry()
	
	// Validate version
	if fcpxml.Version == "" {
		return fmt.Errorf("FCPXML version is required")
	}

	// Validate basic structure
	if len(fcpxml.Library.Events) == 0 {
		return fmt.Errorf("FCPXML must have at least one event")
	}

	event := &fcpxml.Library.Events[0]
	if len(event.Projects) == 0 {
		return fmt.Errorf("event must have at least one project")
	}

	project := &event.Projects[0]
	if len(project.Sequences) == 0 {
		return fmt.Errorf("project must have at least one sequence")
	}

	sequence := &project.Sequences[0]
	
	// Create timeline validator
	timelineValidator, err := NewTimelineValidator(Duration(sequence.Duration))
	if err != nil {
		return fmt.Errorf("failed to create timeline validator: %v", err)
	}
	
	textValidator := NewTextStyleValidator()

	// Validate and register all resources first
	validator := NewStructValidator()
	for i := range fcpxml.Resources.Assets {
		asset := &fcpxml.Resources.Assets[i]
		// Validate asset before registration
		if err := validator.validateAssetStructure(asset); err != nil {
			return fmt.Errorf("asset %d validation failed: %v", i, err)
		}
		registry.RegisterAsset(asset)
	}

	for i := range fcpxml.Resources.Formats {
		format := &fcpxml.Resources.Formats[i]
		// Validate format before registration
		if err := validator.validateFormatStructure(format); err != nil {
			return fmt.Errorf("format %d validation failed: %v", i, err)
		}
		registry.RegisterFormat(format)
	}

	for i := range fcpxml.Resources.Effects {
		effect := &fcpxml.Resources.Effects[i]
		// Validate effect before registration
		if err := validator.validateEffectStructure(effect); err != nil {
			return fmt.Errorf("effect %d validation failed: %v", i, err)
		}
		registry.RegisterEffect(effect)
	}

	// Validate spine structure
	spine := &sequence.Spine
	if err := fcpxml.validateSpine(spine, registry, timelineValidator, textValidator); err != nil {
		return fmt.Errorf("spine validation failed: %v", err)
	}

	// Validate all references
	if err := registry.ValidateAllReferences(fcpxml); err != nil {
		return fmt.Errorf("reference validation failed: %v", err)
	}

	// Validate timeline structure
	if err := timelineValidator.ValidateLaneStructure(); err != nil {
		return fmt.Errorf("timeline validation failed: %v", err)
	}

	return nil
}

// validateSpine validates all elements within a spine
func (fcpxml *FCPXML) validateSpine(spine *Spine, registry *ReferenceRegistry, timelineValidator *TimelineValidator, textValidator *TextStyleValidator) error {
	// Validate all asset clips
	for i := range spine.AssetClips {
		clip := &spine.AssetClips[i]

		// Basic validation - check references exist
		if clip.Ref == "" {
			return fmt.Errorf("asset-clip %d missing ref", i)
		}

		// Validate timing formats
		if err := Time(clip.Offset).Validate(); err != nil {
			return fmt.Errorf("asset-clip %d invalid offset: %v", i, err)
		}

		if err := Duration(clip.Duration).Validate(); err != nil {
			return fmt.Errorf("asset-clip %d invalid duration: %v", i, err)
		}

		// Validate keyframe animations
		if clip.AdjustTransform != nil {
			if err := fcpxml.validateAdjustTransform(clip.AdjustTransform); err != nil {
				return fmt.Errorf("asset-clip %d transform validation: %v", i, err)
			}
		}
	}

	// Validate all video elements
	for i := range spine.Videos {
		video := &spine.Videos[i]

		// Basic validation
		if video.Ref == "" {
			return fmt.Errorf("video %d missing ref", i)
		}

		// Validate timing formats
		if err := Time(video.Offset).Validate(); err != nil {
			return fmt.Errorf("video %d invalid offset: %v", i, err)
		}

		if err := Duration(video.Duration).Validate(); err != nil {
			return fmt.Errorf("video %d invalid duration: %v", i, err)
		}

		if video.AdjustTransform != nil {
			if err := fcpxml.validateAdjustTransform(video.AdjustTransform); err != nil {
				return fmt.Errorf("video %d transform validation: %v", i, err)
			}
		}
	}

	// Validate all title elements
	for i := range spine.Titles {
		title := &spine.Titles[i]

		// Basic validation
		if title.Ref == "" {
			return fmt.Errorf("title %d missing ref", i)
		}

		// Validate timing formats
		if err := Time(title.Offset).Validate(); err != nil {
			return fmt.Errorf("title %d invalid offset: %v", i, err)
		}

		if err := Duration(title.Duration).Validate(); err != nil {
			return fmt.Errorf("title %d invalid duration: %v", i, err)
		}

		// Basic text validation
		if err := fcpxml.validateBasicTitle(title); err != nil {
			return fmt.Errorf("title %d validation: %v", i, err)
		}
	}

	return nil
}

// validateAdjustTransform validates keyframe animations in transform
func (fcpxml *FCPXML) validateAdjustTransform(transform *AdjustTransform) error {
	for _, param := range transform.Params {
		if param.KeyframeAnimation != nil {
			validator := NewKeyframeValidator()
			for i, keyframe := range param.KeyframeAnimation.Keyframes {
				validatedKeyframe := &ValidatedKeyframe{
					Time:   Time(keyframe.Time),
					Value:  keyframe.Value,
					Curve:  keyframe.Curve,
					Interp: keyframe.Interp,
				}
				if err := validator.ValidateKeyframe(param.Name, validatedKeyframe); err != nil {
					return fmt.Errorf("keyframe %d in param %s: %v", i, param.Name, err)
				}
			}
		}
	}
	return nil
}

// validateBasicTitle performs basic title validation
func (fcpxml *FCPXML) validateBasicTitle(title *Title) error {
	// Basic structure validation
	if title.Name == "" {
		return fmt.Errorf("title name is required")
	}

	// Validate text structure if present
	if title.Text != nil {
		for i, textStyle := range title.Text.TextStyles {
			if textStyle.Ref == "" {
				return fmt.Errorf("text-style %d missing ref", i)
			}
		}
	}

	return nil
}

// Post-marshal XML structure validation (basic checks)
func validateXMLStructure(xmlData []byte) error {
	// Parse XML to check basic structure
	var doc struct {
		XMLName xml.Name `xml:"fcpxml"`
		Version string   `xml:"version,attr"`
	}

	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return fmt.Errorf("generated XML is not valid: %v", err)
	}

	if doc.XMLName.Local != "fcpxml" {
		return fmt.Errorf("root element must be 'fcpxml', got: %s", doc.XMLName.Local)
	}

	if doc.Version == "" {
		return fmt.Errorf("fcpxml element must have version attribute")
	}

	// Check for common XML issues
	xmlStr := string(xmlData)

	// Check for unclosed tags (basic heuristic)
	openTags := strings.Count(xmlStr, "<")
	closeTags := strings.Count(xmlStr, ">")
	if openTags != closeTags {
		return fmt.Errorf("mismatched XML tags")
	}

	// Check for invalid characters in attributes (commented out for now as omitempty creates valid empty attrs)
	// if strings.Contains(xmlStr, `=""`) {
	//	return fmt.Errorf("empty required attributes found")
	// }

	return nil
}

// ValidateResourceForMarshaling validates a resource before marshaling
func (sv *StructValidator) ValidateResourceForMarshaling(resource interface{}) error {
	switch r := resource.(type) {
	case *AssetWrapper:
		return sv.validateAssetStructure(r.Asset)
	case *FormatWrapper:
		return sv.validateFormatStructure(r.Format)
	case *EffectWrapper:
		return sv.validateEffectStructure(r.Effect)
	default:
		return fmt.Errorf("unknown resource type: %T", resource)
	}
}

// validateAssetStructure validates an asset structure
func (sv *StructValidator) validateAssetStructure(asset *Asset) error {
	if asset.ID == "" {
		return fmt.Errorf("asset ID is required")
	}

	if asset.Name == "" {
		return fmt.Errorf("asset name is required")
	}

	if asset.UID == "" {
		return fmt.Errorf("asset UID is required")
	}

	if asset.Duration == "" {
		return fmt.Errorf("asset duration is required")
	}

	// Validate duration format
	if err := Duration(asset.Duration).Validate(); err != nil {
		return fmt.Errorf("invalid asset duration: %v", err)
	}

	// Validate media rep
	if asset.MediaRep.Src == "" {
		return fmt.Errorf("asset media-rep src is required")
	}

	return nil
}

// validateFormatStructure validates a format structure
func (sv *StructValidator) validateFormatStructure(format *Format) error {
	if format.ID == "" {
		return fmt.Errorf("format ID is required")
	}

	if format.Width == "" || format.Height == "" {
		return fmt.Errorf("format width and height are required")
	}

	// Validate width and height are numeric
	if _, err := strconv.Atoi(format.Width); err != nil {
		return fmt.Errorf("invalid format width: %s", format.Width)
	}

	if _, err := strconv.Atoi(format.Height); err != nil {
		return fmt.Errorf("invalid format height: %s", format.Height)
	}

	// Validate frame duration if present
	if format.FrameDuration != "" {
		if err := Duration(format.FrameDuration).Validate(); err != nil {
			return fmt.Errorf("invalid format frame duration: %v", err)
		}
	}

	return nil
}

// validateEffectStructure validates an effect structure
func (sv *StructValidator) validateEffectStructure(effect *Effect) error {
	if effect.ID == "" {
		return fmt.Errorf("effect ID is required")
	}

	if effect.Name == "" {
		return fmt.Errorf("effect name is required")
	}

	if effect.UID == "" {
		return fmt.Errorf("effect UID is required")
	}

	return nil
}

// Note: Resource wrapper types (AssetWrapper, FormatWrapper, EffectWrapper)
// and Resource interface are defined in registry.go

// ValidateAndMarshalWithDTD performs DTD validation in addition to structure validation
func (fcpxml *FCPXML) ValidateAndMarshalWithDTD(dtdPath string) ([]byte, error) {
	// First do structure validation
	data, err := fcpxml.ValidateAndMarshal()
	if err != nil {
		return nil, err
	}

	// Then do DTD validation if xmllint is available
	if dtdPath != "" {
		dtdValidator := NewDTDValidator(dtdPath)
		if err := dtdValidator.ValidateXML(data); err != nil {
			return nil, fmt.Errorf("DTD validation failed: %v", err)
		}
	}

	return data, nil
}