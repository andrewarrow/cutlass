# API Reference

This document provides a comprehensive reference for all functions, types, and constants in the Cutlass FCPXML generation framework.

## Core Functions

### FCPXML Generation

#### `GenerateEmpty(filename string) (*FCPXML, error)`
Creates an empty FCPXML file structure with default settings.

```go
fcpxml, err := fcp.GenerateEmpty("")
if err != nil {
    return fmt.Errorf("failed to create base FCPXML: %v", err)
}
```

**Returns:** Pointer to FCPXML struct with initialized Library, Events, Projects, and Sequences.

#### `WriteToFile(fcpxml *FCPXML, filename string) error`
Writes an FCPXML struct to a file with proper XML formatting.

```go
err := fcp.WriteToFile(fcpxml, "output.fcpxml")
if err != nil {
    return fmt.Errorf("failed to write FCPXML: %v", err)
}
```

### Duration and Timing

#### `ConvertSecondsToFCPDuration(seconds float64) string`
Converts seconds to frame-aligned FCP duration format.

**🚨 CRITICAL:** Always use this function for duration conversion to ensure frame alignment.

```go
duration := fcp.ConvertSecondsToFCPDuration(5.5)  // Returns "132132/24000s"
```

**Parameters:**
- `seconds` - Duration in seconds (float64)

**Returns:** Frame-aligned duration string in format "numerator/24000s"

**Frame Alignment Rules:**
- Uses 24000/1001 ≈ 23.976 fps timebase
- Ensures all durations are multiples of 1001/24000s frame duration
- Prevents "not on edit frame boundary" errors

#### `ValidateFrameAlignment(duration string) error`
Validates that a duration string is properly frame-aligned.

```go
err := fcp.ValidateFrameAlignment("132132/24000s")
if err != nil {
    return fmt.Errorf("invalid duration: %v", err)
}
```

## Resource Management

### ResourceRegistry

#### `NewResourceRegistry(fcpxml *FCPXML) *ResourceRegistry`
Creates a new resource registry for managing IDs and resources.

```go
registry := fcp.NewResourceRegistry(fcpxml)
```

**Features:**
- Thread-safe resource management
- ID collision prevention
- Automatic ID generation
- UID consistency tracking

#### `ReserveIDs(count int) []string`
Reserves multiple sequential resource IDs.

```go
ids := registry.ReserveIDs(3)
assetID := ids[0]    // "r2"
formatID := ids[1]   // "r3"
effectID := ids[2]   // "r4"
```

### Transaction System

#### `NewTransaction(registry *ResourceRegistry) *ResourceTransaction`
Creates a new transaction for atomic resource operations.

```go
tx := fcp.NewTransaction(registry)
defer tx.Rollback()  // Cleanup on failure

// Perform operations...

if err := tx.Commit(); err != nil {
    return err
}
```

#### `CreateVideoAssetWithDetection(id, filePath, baseName, duration, formatID string) error`
Creates a video asset with automatic property detection.

```go
err := tx.CreateVideoAssetWithDetection(assetID, videoPath, "video.mp4", duration, formatID)
if err != nil {
    return fmt.Errorf("failed to create video asset: %v", err)
}
```

**Features:**
- Automatic width/height detection
- Audio track detection
- Frame rate validation
- Absolute path resolution
- Security bookmark generation

#### `CreateAsset(id, filePath, baseName, duration, formatID string) (*Asset, error)`
Creates a basic asset without property detection.

```go
asset, err := tx.CreateAsset(assetID, imagePath, "image.png", "0s", formatID)
if err != nil {
    return fmt.Errorf("failed to create asset: %v", err)
}
```

#### `CreateFormatWithFrameDuration(id, frameDuration, width, height, colorSpace string) (*Format, error)`
Creates a format with frame duration (for videos).

```go
format, err := tx.CreateFormatWithFrameDuration(formatID, "1001/24000s", "1920", "1080", "1-1-1 (Rec. 709)")
if err != nil {
    return fmt.Errorf("failed to create format: %v", err)
}
```

#### `CreateImageFormat(id, width, height, colorSpace string) (*Format, error)`
Creates a format without frame duration (for images).

```go
format, err := tx.CreateImageFormat(formatID, "1280", "720", "1-13-1")
if err != nil {
    return fmt.Errorf("failed to create image format: %v", err)
}
```

#### `CreateEffect(id, name, uid string) (*Effect, error)`
Creates an effect resource with verified UID.

```go
effect, err := tx.CreateEffect(effectID, "Gaussian Blur", "FFGaussianBlur")
if err != nil {
    return fmt.Errorf("failed to create effect: %v", err)
}
```

## ID Generation

### `GenerateUID(filePath string) string`
Generates a consistent UID based on filename.

```go
uid := fcp.GenerateUID("video.mp4")  // Always returns same UID for same filename
```

**Features:**
- Filename-based (not path-based) for consistency
- MD5 hash ensures uniqueness
- FCP-compatible format

### `GenerateTextStyleID(text, baseName string) string`
Generates unique text style IDs.

```go
styleID := fcp.GenerateTextStyleID("Hello World", "title")  // Returns "ts" + hash
```

### `GenerateResourceID(index int) string`
Generates resource IDs in FCP format.

```go
resourceID := fcp.GenerateResourceID(5)  // Returns "r5"
```

## Core Types

### FCPXML Structure

```go
type FCPXML struct {
    XMLName   xml.Name  `xml:"fcpxml"`
    Version   string    `xml:"version,attr"`
    Resources Resources `xml:"resources"`
    Library   Library   `xml:"library"`
}
```

### Resources

```go
type Resources struct {
    Assets  []Asset  `xml:"asset,omitempty"`
    Formats []Format `xml:"format"`
    Effects []Effect `xml:"effect,omitempty"`
    Media   []Media  `xml:"media,omitempty"`
}
```

### Asset

```go
type Asset struct {
    ID            string   `xml:"id,attr"`
    Name          string   `xml:"name,attr"`
    UID           string   `xml:"uid,attr"`
    Start         string   `xml:"start,attr"`
    Duration      string   `xml:"duration,attr"`
    HasVideo      string   `xml:"hasVideo,attr,omitempty"`
    HasAudio      string   `xml:"hasAudio,attr,omitempty"`
    Format        string   `xml:"format,attr,omitempty"`
    VideoSources  string   `xml:"videoSources,attr,omitempty"`
    AudioSources  string   `xml:"audioSources,attr,omitempty"`
    AudioChannels string   `xml:"audioChannels,attr,omitempty"`
    AudioRate     string   `xml:"audioRate,attr,omitempty"`
    MediaRep      MediaRep `xml:"media-rep"`
    Metadata      *Metadata `xml:"metadata,omitempty"`
}
```

**Asset Properties by Media Type:**

| Property | Images | Videos | Audio |
|----------|---------|--------|-------|
| `Duration` | `"0s"` | Actual duration | Actual duration |
| `HasVideo` | `"1"` | `"1"` | Not set |
| `HasAudio` | Not set | `"1"` if present | `"1"` |
| `Format` | Required | Required | Required |

### Format

```go
type Format struct {
    ID            string `xml:"id,attr"`
    Name          string `xml:"name,attr,omitempty"`
    FrameDuration string `xml:"frameDuration,attr,omitempty"`
    Width         string `xml:"width,attr,omitempty"`
    Height        string `xml:"height,attr,omitempty"`
    ColorSpace    string `xml:"colorSpace,attr,omitempty"`
}
```

**Format Properties by Media Type:**

| Property | Images | Videos |
|----------|---------|--------|
| `FrameDuration` | **OMITTED** | **REQUIRED** |
| `Width/Height` | Detected | Detected |
| `ColorSpace` | Standard values | Standard values |

### Timeline Elements

#### AssetClip (for Videos/Audio)

```go
type AssetClip struct {
    XMLName         xml.Name         `xml:"asset-clip"`
    Ref             string           `xml:"ref,attr"`
    Lane            string           `xml:"lane,attr,omitempty"`
    Offset          string           `xml:"offset,attr"`
    Name            string           `xml:"name,attr"`
    Start           string           `xml:"start,attr,omitempty"`
    Duration        string           `xml:"duration,attr"`
    AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"`
    FilterVideos    []FilterVideo    `xml:"filter-video,omitempty"`
}
```

#### Video (for Images/Generated Content)

```go
type Video struct {
    XMLName         xml.Name         `xml:"video"`
    Ref             string           `xml:"ref,attr"`
    Lane            string           `xml:"lane,attr,omitempty"`
    Offset          string           `xml:"offset,attr"`
    Name            string           `xml:"name,attr"`
    Duration        string           `xml:"duration,attr"`
    AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"`
    FilterVideos    []FilterVideo    `xml:"filter-video,omitempty"`
}
```

#### Title (for Text)

```go
type Title struct {
    XMLName       xml.Name      `xml:"title"`
    Ref           string        `xml:"ref,attr"`
    Lane          string        `xml:"lane,attr,omitempty"`
    Offset        string        `xml:"offset,attr"`
    Name          string        `xml:"name,attr"`
    Duration      string        `xml:"duration,attr"`
    Text          *TitleText    `xml:"text,omitempty"`
    TextStyleDefs []TextStyleDef `xml:"text-style-def,omitempty"`
}
```

### Animation and Effects

#### AdjustTransform

```go
type AdjustTransform struct {
    Position string  `xml:"position,attr,omitempty"`
    Scale    string  `xml:"scale,attr,omitempty"`
    Params   []Param `xml:"param,omitempty"`
}
```

#### Param (for Keyframe Animation)

```go
type Param struct {
    Name              string             `xml:"name,attr"`
    Value             string             `xml:"value,attr,omitempty"`
    KeyframeAnimation *KeyframeAnimation `xml:"keyframeAnimation,omitempty"`
}
```

#### Keyframe

```go
type Keyframe struct {
    Time   string `xml:"time,attr"`
    Value  string `xml:"value,attr"`
    Interp string `xml:"interp,attr,omitempty"` // Only for opacity, volume
    Curve  string `xml:"curve,attr,omitempty"`  // Only for scale, rotation, anchor
}
```

**Keyframe Attribute Support:**

| Parameter | `interp` | `curve` | Example |
|-----------|----------|---------|---------|
| `position` | ❌ | ❌ | `<keyframe time="0s" value="0 0"/>` |
| `scale` | ❌ | ✅ | `<keyframe time="0s" value="1 1" curve="linear"/>` |
| `rotation` | ❌ | ✅ | `<keyframe time="0s" value="0" curve="smooth"/>` |
| `opacity` | ✅ | ✅ | `<keyframe time="0s" value="1" interp="easeIn" curve="smooth"/>` |

#### FilterVideo

```go
type FilterVideo struct {
    Ref    string  `xml:"ref,attr"`
    Name   string  `xml:"name,attr"`
    Params []Param `xml:"param,omitempty"`
}
```

### Text System

#### TitleText

```go
type TitleText struct {
    TextStyles []TextStyleRef `xml:"text-style"`
}
```

#### TextStyleDef

```go
type TextStyleDef struct {
    ID        string    `xml:"id,attr"`
    TextStyle TextStyle `xml:"text-style"`
}
```

#### TextStyle

```go
type TextStyle struct {
    Font             string `xml:"font,attr"`
    FontSize         string `xml:"fontSize,attr"`
    FontColor        string `xml:"fontColor,attr"`
    ShadowColor      string `xml:"shadowColor,attr,omitempty"`
    ShadowOffset     string `xml:"shadowOffset,attr,omitempty"`
    ShadowBlurRadius string `xml:"shadowBlurRadius,attr,omitempty"`
    Alignment        string `xml:"alignment,attr,omitempty"`
}
```

## Constants

### Frame Rate Constants

```go
const (
    FCPTimebase      = 24000
    FCPFrameDuration = 1001
    FCPFrameRate     = 23.976023976023976
)
```

### Verified Effect UIDs

```go
var VerifiedEffectUIDs = map[string]bool{
    "FFGaussianBlur":      true,
    "FFMotionBlur":        true,
    "FFColorCorrection":   true,
    "FFSaturation":        true,
    "FFLevels":            true,
    "FFSuperEllipseMask":  true,
    "FFRectangleMask":     true,
    "FFCircleMask":        true,
    "FFPolygonMask":       true,
    "FFAudioGain":         true,
    "FFAudioEQ":           true,
    "FFAudioCompressor":   true,
}
```

### Motion Template Paths

```go
var VerifiedMotionTemplates = []string{
    ".../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn",
    ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
    ".../Titles.localized/Basic Text.localized/Lower Third.localized/Lower Third.moti",
}
```

## Validation Functions

### `ValidateClaudeCompliance(fcpxml *FCPXML) error`
Comprehensive validation of FCPXML structure and compliance with best practices.

```go
err := fcp.ValidateClaudeCompliance(fcpxml)
if err != nil {
    return fmt.Errorf("FCPXML validation failed: %v", err)
}
```

**Validates:**
- Resource reference integrity
- Media type consistency
- Frame alignment
- ID uniqueness
- Effect UID verification
- Keyframe attribute compliance

### `ValidateResourceReferences(fcpxml *FCPXML) error`
Validates that all `ref` attributes point to existing resources.

### `ValidateMediaTypes(fcpxml *FCPXML) error`
Validates that images use Video elements and videos use AssetClip elements.

### `ValidateEffectUIDs(fcpxml *FCPXML) error`
Validates that all effect UIDs are verified and safe to use.

## Error Types

### Common Error Patterns

```go
// Resource reference errors
fmt.Errorf("asset-clip ref='%s' not found", clip.Ref)

// Frame alignment errors
fmt.Errorf("duration not frame-aligned: %s", duration)

// Media type errors
fmt.Errorf("video element with ref='%s' should be asset-clip", video.Ref)

// Effect UID errors
fmt.Errorf("unverified effect UID: %s", uid)

// Keyframe attribute errors
fmt.Errorf("position keyframes don't support interp/curve attributes")
```

## Usage Examples

### Basic FCPXML Generation

```go
func GenerateBasicFCPXML(videoPath, outputPath string) error {
    // 1. Create base FCPXML
    fcpxml, err := fcp.GenerateEmpty("")
    if err != nil {
        return err
    }
    
    // 2. Set up resource management
    registry := fcp.NewResourceRegistry(fcpxml)
    tx := fcp.NewTransaction(registry)
    defer tx.Rollback()
    
    // 3. Reserve IDs
    ids := tx.ReserveIDs(2)
    assetID := ids[0]
    formatID := ids[1]
    
    // 4. Create video asset
    duration := fcp.ConvertSecondsToFCPDuration(10.0)
    err = tx.CreateVideoAssetWithDetection(assetID, videoPath, filepath.Base(videoPath), duration, formatID)
    if err != nil {
        return err
    }
    
    // 5. Add to timeline
    sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
    sequence.Spine.AssetClips = append(sequence.Spine.AssetClips, fcp.AssetClip{
        Ref:      assetID,
        Offset:   "0s",
        Duration: duration,
        Name:     filepath.Base(videoPath),
    })
    
    // 6. Commit and validate
    if err := tx.Commit(); err != nil {
        return err
    }
    
    if err := fcp.ValidateClaudeCompliance(fcpxml); err != nil {
        return err
    }
    
    // 7. Write output
    return fcp.WriteToFile(fcpxml, outputPath)
}
```

### Animation Example

```go
func addScaleAnimation(video *fcp.Video, duration string) {
    video.AdjustTransform = &fcp.AdjustTransform{
        Params: []fcp.Param{
            {
                Name: "scale",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: "0s", Value: "1 1", Curve: "linear"},
                        {Time: duration, Value: "1.5 1.5", Curve: "smooth"},
                    },
                },
            },
        },
    }
}
```

### Text Creation Example

```go
func addTextTitle(fcpxml *fcp.FCPXML, text string, offset, duration string) error {
    registry := fcp.NewResourceRegistry(fcpxml)
    tx := fcp.NewTransaction(registry)
    defer tx.Rollback()
    
    effectID := tx.ReserveIDs(1)[0]
    err := tx.CreateEffect(effectID, "Text", ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti")
    if err != nil {
        return err
    }
    
    styleID := fcp.GenerateTextStyleID(text, "title")
    
    title := fcp.Title{
        Ref:      effectID,
        Offset:   offset,
        Duration: duration,
        Name:     "Text Title",
        Text: &fcp.TitleText{
            TextStyles: []fcp.TextStyleRef{
                {Ref: styleID, Text: text},
            },
        },
        TextStyleDefs: []fcp.TextStyleDef{
            {
                ID: styleID,
                TextStyle: fcp.TextStyle{
                    Font:      "Helvetica",
                    FontSize:  "48",
                    FontColor: "1 1 1 1",
                    Alignment: "center",
                },
            },
        },
    }
    
    sequence := &fcpxml.Library.Events[0].Projects[0].Sequences[0]
    sequence.Spine.Titles = append(sequence.Spine.Titles, title)
    
    return tx.Commit()
}
```

This API reference provides the complete interface for the Cutlass FCPXML generation framework, ensuring proper usage of all functions and types while following the critical rules outlined in the best practices documentation.