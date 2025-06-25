// Package fcp defines the struct types for FCPXML generation.
//
// 🚨 CRITICAL: These structs are the ONLY way to generate XML (see CLAUDE.md)
// - NEVER use string templates → USE xml.MarshalIndent() function only
// - NEVER set .Content or .InnerXML → APPEND to struct slices (e.g., spine.AssetClips)  
// - VALIDATE output → RUN ValidateClaudeCompliance() + xmllint DTD validation
// - FOR frame alignment → USE ConvertSecondsToFCPDuration() function
package fcp

import (
	"encoding/xml"
	"strconv"
	"strings"
)

type FCPXML struct {
	XMLName   xml.Name  `xml:"fcpxml"`
	Version   string    `xml:"version,attr"`
	Resources Resources `xml:"resources"`
	Library   Library   `xml:"library"`
}

// Resources contains all assets, formats, effects, and media definitions.
//
// 🚨 CLAUDE.md Rule: Unique ID Requirements → USE this counting pattern:
// resourceCount := len(Assets)+len(Formats)+len(Effects)+len(Media)
// nextID := fmt.Sprintf("r%d", resourceCount+1)
// NEVER hardcode IDs like "r1", "r2" - ALWAYS count existing resources
type Resources struct {
	Assets     []Asset     `xml:"asset,omitempty"`
	Formats    []Format    `xml:"format"`
	Effects    []Effect    `xml:"effect,omitempty"`
	Media      []Media     `xml:"media,omitempty"`
}

// Effect represents a Motion or standard FCP title effect referenced by <title ref="…"> elements.
type Effect struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	UID  string `xml:"uid,attr,omitempty"`
}


type Format struct {
	ID            string `xml:"id,attr"`
	Name          string `xml:"name,attr,omitempty"` // CRITICAL: omitempty allows compatible formats without names
	FrameDuration string `xml:"frameDuration,attr,omitempty"`
	Width         string `xml:"width,attr,omitempty"`
	Height        string `xml:"height,attr,omitempty"`
	ColorSpace    string `xml:"colorSpace,attr,omitempty"`
}

// Asset represents a media asset (video, audio, image) in FCPXML.
//
// 🚨 CLAUDE.md Rule: UID Consistency Requirements → USE generateUID() function
// - UID = generateUID(filename) for deterministic UIDs based on filename  
// - NEVER base UID on file path (causes "cannot be imported again" errors)
// - FOR durations → USE ConvertSecondsToFCPDuration() function
type Asset struct {
	ID            string    `xml:"id,attr"`
	Name          string    `xml:"name,attr"`
	UID           string    `xml:"uid,attr"`
	Start         string    `xml:"start,attr"`
	HasVideo      string    `xml:"hasVideo,attr,omitempty"`
	Format        string    `xml:"format,attr,omitempty"`
	VideoSources  string    `xml:"videoSources,attr,omitempty"`
	HasAudio      string    `xml:"hasAudio,attr,omitempty"`
	AudioSources  string    `xml:"audioSources,attr,omitempty"`
	AudioChannels string    `xml:"audioChannels,attr,omitempty"`
	AudioRate     string    `xml:"audioRate,attr,omitempty"`
	Duration      string    `xml:"duration,attr"`
	MediaRep      MediaRep  `xml:"media-rep"`
	Metadata      *Metadata `xml:"metadata,omitempty"`
}

type MediaRep struct {
	Kind     string `xml:"kind,attr"`
	Sig      string `xml:"sig,attr"`
	Src      string `xml:"src,attr"`
	Bookmark string `xml:"bookmark,omitempty"`
}

type Metadata struct {
	MDs []MetadataItem `xml:"md"`
}

type MetadataItem struct {
	Key   string      `xml:"key,attr"`
	Value string      `xml:"value,attr,omitempty"`
	Array *StringArray `xml:"array,omitempty"`
}

type StringArray struct {
	Strings []string `xml:"string"`
}

type Media struct {
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	UID      string   `xml:"uid,attr"`
	ModDate  string   `xml:"modDate,attr,omitempty"`
	Sequence Sequence `xml:"sequence"`
}

type RefClip struct {
	XMLName         xml.Name         `xml:"ref-clip"`
	Ref             string           `xml:"ref,attr"`
	Offset          string           `xml:"offset,attr"`
	Name            string           `xml:"name,attr"`
	Duration        string           `xml:"duration,attr"`
	AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"`
	Titles          []Title          `xml:"title,omitempty"`
}

type Library struct {
	Location          string            `xml:"location,attr,omitempty"`
	Events            []Event           `xml:"event"`
	SmartCollections  []SmartCollection `xml:"smart-collection,omitempty"`
}

type Event struct {
	Name     string    `xml:"name,attr"`
	UID      string    `xml:"uid,attr,omitempty"`
	Projects []Project `xml:"project"`
}

type Project struct {
	Name      string     `xml:"name,attr"`
	UID       string     `xml:"uid,attr,omitempty"`
	ModDate   string     `xml:"modDate,attr,omitempty"`
	Sequences []Sequence `xml:"sequence"`
}

type Sequence struct {
	Format      string `xml:"format,attr"`
	Duration    string `xml:"duration,attr"`
	TCStart     string `xml:"tcStart,attr"`
	TCFormat    string `xml:"tcFormat,attr"`
	AudioLayout string `xml:"audioLayout,attr"`
	AudioRate   string `xml:"audioRate,attr"`
	Spine       Spine  `xml:"spine"`
}

// TimelineElement represents any element that can appear in a spine with an offset
type TimelineElement interface {
	GetOffset() string
	GetEndOffset() string
}

// Spine represents the main timeline container in FCPXML.
//
// 🚨 CLAUDE.md Rule: NO XML STRING TEMPLATES → USE struct slices:
// spine.AssetClips = append(spine.AssetClips, assetClip) ✅
// spine.Content = fmt.Sprintf("<asset-clip...") ❌ CRITICAL VIOLATION!
// FOR durations → USE ConvertSecondsToFCPDuration() function
type Spine struct {
	XMLName    xml.Name    `xml:"spine"`
	AssetClips []AssetClip `xml:"asset-clip,omitempty"`
	Gaps       []Gap       `xml:"gap,omitempty"`
	Titles     []Title     `xml:"title,omitempty"`
	Videos     []Video     `xml:"video,omitempty"`
}

// MarshalXML implements custom XML marshaling to maintain chronological order
func (s Spine) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Start the spine element
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// Collect all elements with their offsets
	type elementWithOffset struct {
		offset  int
		element interface{}
	}
	var elements []elementWithOffset

	// Add all element types
	for _, clip := range s.AssetClips {
		elements = append(elements, elementWithOffset{
			offset:  parseFCPDurationForSort(clip.Offset),
			element: clip,
		})
	}
	for _, video := range s.Videos {
		elements = append(elements, elementWithOffset{
			offset:  parseFCPDurationForSort(video.Offset),
			element: video,
		})
	}
	for _, title := range s.Titles {
		elements = append(elements, elementWithOffset{
			offset:  parseFCPDurationForSort(title.Offset),
			element: title,
		})
	}
	for _, gap := range s.Gaps {
		elements = append(elements, elementWithOffset{
			offset:  parseFCPDurationForSort(gap.Offset),
			element: gap,
		})
	}

	// Sort by offset
	for i := 0; i < len(elements)-1; i++ {
		for j := 0; j < len(elements)-i-1; j++ {
			if elements[j].offset > elements[j+1].offset {
				elements[j], elements[j+1] = elements[j+1], elements[j]
			}
		}
	}

	// Encode elements in chronological order
	for _, elem := range elements {
		if err := e.Encode(elem.element); err != nil {
			return err
		}
	}

	// End the spine element
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// parseFCPDurationForSort parses FCP duration for sorting with frame-aligned values
func parseFCPDurationForSort(duration string) int {
	if duration == "0s" {
		return 0
	}
	
	// Parse rational duration formats like "12345/24000s", "547547/60000s", etc.
	if strings.HasSuffix(duration, "s") && strings.Contains(duration, "/") {
		// Remove the "s" suffix
		durationNoS := strings.TrimSuffix(duration, "s")
		
		// Split by "/"
		parts := strings.Split(durationNoS, "/")
		if len(parts) == 2 {
			numerator, err1 := strconv.Atoi(parts[0])
			denominator, err2 := strconv.Atoi(parts[1])
			
			if err1 == nil && err2 == nil && denominator != 0 {
				// 🚨 CLAUDE.md CRITICAL: Frame Boundary Alignment
				// FCP uses 1001/24000s frame duration (≈ 23.976 fps)
				// All durations MUST be frame-aligned: (frames × 1001)/24000s
				
				// Convert to exact frame count using FCP's frame duration
				// frames = (numerator/denominator) / (1001/24000) = (numerator * 24000) / (denominator * 1001)
				framesFloat := float64(numerator * 24000) / float64(denominator * 1001)
				frames := int(framesFloat + 0.5) // Round to nearest frame
				
				// Return frame-aligned value: frames * 1001
				return frames * 1001
			}
		}
	}
	
	return 0
}

type AssetClip struct {
	XMLName         xml.Name         `xml:"asset-clip"`
	Ref             string           `xml:"ref,attr"`
	Lane            string           `xml:"lane,attr,omitempty"`
	Offset          string           `xml:"offset,attr"`
	Name            string           `xml:"name,attr"`
	Start           string           `xml:"start,attr,omitempty"`
	Duration        string           `xml:"duration,attr"`
	Format          string           `xml:"format,attr,omitempty"`
	TCFormat        string           `xml:"tcFormat,attr,omitempty"`
	AudioRole       string           `xml:"audioRole,attr,omitempty"`
	ConformRate     *ConformRate     `xml:"conform-rate,omitempty"`
	AdjustCrop      *AdjustCrop      `xml:"adjust-crop,omitempty"`
	AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"`
	NestedAssetClips []AssetClip     `xml:"asset-clip,omitempty"`
	Titles          []Title          `xml:"title,omitempty"`
	Videos          []Video          `xml:"video,omitempty"`
	FilterVideos    []FilterVideo    `xml:"filter-video,omitempty"`
}

// GetOffset implements TimelineElement interface
func (ac AssetClip) GetOffset() string {
	return ac.Offset
}

// GetEndOffset implements TimelineElement interface
func (ac AssetClip) GetEndOffset() string {
	// This would require parsing offset and duration to calculate end time
	// For now, return offset (implementation can be added later if needed)
	return ac.Offset
}

type Gap struct {
	XMLName        xml.Name        `xml:"gap"`
	Name           string          `xml:"name,attr"`
	Offset         string          `xml:"offset,attr"`
	Duration       string          `xml:"duration,attr"`
	Titles         []Title         `xml:"title,omitempty"`
	GeneratorClips []GeneratorClip `xml:"generator-clip,omitempty"`
}

type Title struct {
	XMLName xml.Name `xml:"title"`
	Ref          string         `xml:"ref,attr"`
	Lane         string         `xml:"lane,attr,omitempty"`
	Offset       string         `xml:"offset,attr"`
	Name         string         `xml:"name,attr"`
	Duration     string         `xml:"duration,attr"`
	Start        string         `xml:"start,attr,omitempty"`
	Params       []Param        `xml:"param,omitempty"`
	Text         *TitleText     `xml:"text,omitempty"`         // Pointer so it can be nil
	TextStyleDefs []TextStyleDef `xml:"text-style-def,omitempty"` // 🚨 BREAKING CHANGE: Was single TextStyleDef, now slice for shadow text
}

// Video represents a video element (shapes, colors, etc.)
type Video struct {
	XMLName xml.Name `xml:"video"`
	Ref           string         `xml:"ref,attr"`
	Lane          string         `xml:"lane,attr,omitempty"`
	Offset        string         `xml:"offset,attr"`
	Name          string         `xml:"name,attr"`
	Duration      string         `xml:"duration,attr"`
	Start         string         `xml:"start,attr,omitempty"`
	Params        []Param        `xml:"param,omitempty"`
	AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"`
	FilterVideos     []FilterVideo   `xml:"filter-video,omitempty"`   // Support filter-video effects
	NestedVideos     []Video     `xml:"video,omitempty"`      // Support nested video elements with lanes
	NestedAssetClips []AssetClip `xml:"asset-clip,omitempty"` // Support nested asset-clip elements with lanes
	NestedTitles     []Title     `xml:"title,omitempty"`      // Support nested title elements with lanes
}

// GetOffset implements TimelineElement interface
func (v Video) GetOffset() string {
	return v.Offset
}

// GetEndOffset implements TimelineElement interface
func (v Video) GetEndOffset() string {
	// This would require parsing offset and duration to calculate end time
	// For now, return offset (implementation can be added later if needed)
	return v.Offset
}

type ConformRate struct {
	ScaleEnabled string `xml:"scaleEnabled,attr,omitempty"`
	SrcFrameRate string `xml:"srcFrameRate,attr,omitempty"`
}

type AdjustCrop struct {
	Mode     string    `xml:"mode,attr"`
	TrimRect *TrimRect `xml:"trim-rect,omitempty"`
}

type TrimRect struct {
	Left   string `xml:"left,attr,omitempty"`
	Right  string `xml:"right,attr,omitempty"`
	Top    string `xml:"top,attr,omitempty"`
	Bottom string `xml:"bottom,attr,omitempty"`
}

type FilterVideo struct {
	Ref    string  `xml:"ref,attr"`
	Name   string  `xml:"name,attr"`
	Params []Param `xml:"param,omitempty"`
}

type AdjustTransform struct {
	Position string  `xml:"position,attr,omitempty"`
	Scale    string  `xml:"scale,attr,omitempty"`
	Params   []Param `xml:"param,omitempty"`
}


type GeneratorClip struct {
	Ref      string  `xml:"ref,attr"`
	Lane     string  `xml:"lane,attr,omitempty"`
	Offset   string  `xml:"offset,attr"`
	Name     string  `xml:"name,attr"`
	Duration string  `xml:"duration,attr"`
	Start    string  `xml:"start,attr,omitempty"`
	Params   []Param `xml:"param,omitempty"`
}

type Param struct {
	Name               string              `xml:"name,attr"`
	Key                string              `xml:"key,attr,omitempty"`
	Value              string              `xml:"value,attr,omitempty"`
	KeyframeAnimation  *KeyframeAnimation  `xml:"keyframeAnimation,omitempty"`
	NestedParams       []Param             `xml:"param,omitempty"`
}

type KeyframeAnimation struct {
	Keyframes []Keyframe `xml:"keyframe"`
}

type Keyframe struct {
	Time   string `xml:"time,attr"`
	Value  string `xml:"value,attr"`
	Interp string `xml:"interp,attr,omitempty"` // linear | ease | easeIn | easeOut (default: linear)
	Curve  string `xml:"curve,attr,omitempty"`  // linear | smooth (default: smooth)
}

// TitleText represents the text content within a title element
// 🚨 BREAKING CHANGE: Changed from single TextStyle to TextStyles slice
// This was needed to support shadow text with multiple text-style elements
// like: <text><text-style ref="ts1">Main</text-style><text-style ref="ts2">Split</text-style></text>
type TitleText struct {
	TextStyles []TextStyleRef `xml:"text-style"`
}

type TextStyleRef struct {
	Ref  string `xml:"ref,attr"`
	Text string `xml:",chardata"`
}

type TextStyleDef struct {
	ID        string    `xml:"id,attr"`
	TextStyle TextStyle `xml:"text-style"`
}

type TextStyle struct {
	Font            string  `xml:"font,attr"`
	FontSize        string  `xml:"fontSize,attr"`
	FontFace        string  `xml:"fontFace,attr,omitempty"`
	FontColor       string  `xml:"fontColor,attr"`
	Bold            string  `xml:"bold,attr,omitempty"`
	Italic          string  `xml:"italic,attr,omitempty"`
	StrokeColor     string  `xml:"strokeColor,attr,omitempty"`
	StrokeWidth     string  `xml:"strokeWidth,attr,omitempty"`
	ShadowColor     string  `xml:"shadowColor,attr,omitempty"`
	ShadowOffset    string  `xml:"shadowOffset,attr,omitempty"`
	ShadowBlurRadius string `xml:"shadowBlurRadius,attr,omitempty"`
	Kerning         string  `xml:"kerning,attr,omitempty"`
	Alignment       string  `xml:"alignment,attr,omitempty"`
	LineSpacing     string  `xml:"lineSpacing,attr,omitempty"`
	Params          []Param `xml:"param,omitempty"`
}

type SmartCollection struct {
	Name     string      `xml:"name,attr"`
	Match    string      `xml:"match,attr"`
	Matches  []Match     `xml:"match-clip,omitempty"`
	MediaMatches []MediaMatch `xml:"match-media,omitempty"`
	RatingMatches []RatingMatch `xml:"match-ratings,omitempty"`
}

type Match struct {
	Rule string `xml:"rule,attr"`
	Type string `xml:"type,attr"`
}

type MediaMatch struct {
	Rule string `xml:"rule,attr"`
	Type string `xml:"type,attr"`
}

type RatingMatch struct {
	Value string `xml:"value,attr"`
}

type ParseOptions struct {
	Tier          int
	ShowElements  bool
	ShowParams    bool
	ShowAnimation bool
	ShowResources bool
	ShowStructure bool
}