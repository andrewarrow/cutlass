# Architecture Overview

This document provides a comprehensive overview of the Cutlass FCPXML generation framework architecture, covering the core design patterns, resource management, and structural principles.

## System Design Principles

### 1. Struct-Based XML Generation
The framework uses Go structs with XML tags to generate FCPXML, ensuring proper XML formatting and preventing common errors:

```go
type Video struct {
    XMLName xml.Name `xml:"video"`
    Ref     string   `xml:"ref,attr"`
    Offset  string   `xml:"offset,attr"`
    Duration string  `xml:"duration,attr"`
    Name    string   `xml:"name,attr"`
    AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"`
}
```

### 2. Resource Registry Pattern
All resources (assets, formats, effects) are managed through a centralized registry that ensures ID uniqueness and proper resource lifecycle management:

```go
type ResourceRegistry struct {
    resources map[string]Resource
    assets    map[string]*Asset
    formats   map[string]*Format  
    effects   map[string]*Effect
    nextResourceID int
    usedIDs        map[string]bool
}
```

### 3. Transaction-Based Operations
Complex operations use transactions to ensure atomicity and proper resource cleanup:

```go
registry := fcp.NewResourceRegistry(fcpxml)
tx := fcp.NewTransaction(registry)
defer tx.Rollback()

// Perform operations...

if err := tx.Commit(); err != nil {
    return err
}
```

## Core Architecture Components

### FCPXML Structure Hierarchy

```
FCPXML
├── Resources
│   ├── Assets (media files)
│   ├── Formats (video/audio specifications)
│   ├── Effects (filters, generators, titles)
│   └── Media (compound clips, multicam)
└── Library
    └── Events
        └── Projects
            └── Sequences
                └── Spine (main timeline)
                    ├── AssetClips (video/audio clips)
                    ├── Videos (images, shapes, colors)
                    ├── Titles (text elements)
                    └── Gaps (spacers)
```

### Resource Types and Their Relationships

#### Assets
Media files with intrinsic properties:
```go
type Asset struct {
    ID            string   `xml:"id,attr"`
    Name          string   `xml:"name,attr"`
    UID           string   `xml:"uid,attr"`        // File-based unique identifier
    Duration      string   `xml:"duration,attr"`   // "0s" for images, actual for videos
    HasVideo      string   `xml:"hasVideo,attr"`   // "1" if contains video
    HasAudio      string   `xml:"hasAudio,attr"`   // "1" if contains audio  
    Format        string   `xml:"format,attr"`     // References Format.ID
    MediaRep      MediaRep `xml:"media-rep"`       // File path and metadata
}
```

#### Formats  
Technical specifications for media:
```go
type Format struct {
    ID            string `xml:"id,attr"`
    Name          string `xml:"name,attr"`
    FrameDuration string `xml:"frameDuration,attr"` // Only for videos, NOT images
    Width         string `xml:"width,attr"`
    Height        string `xml:"height,attr"`
    ColorSpace    string `xml:"colorSpace,attr"`
}
```

#### Timeline Elements
Elements that appear in the spine with temporal properties:

| Element Type | Use Case | Required Attributes | Media Type |
|-------------|----------|-------------------|------------|
| `AssetClip` | Video/audio files | `ref`, `offset`, `duration` | Videos, Audio |
| `Video` | Images, shapes, colors | `ref`, `offset`, `duration` | Images, Generated |
| `Title` | Text elements | `ref`, `offset`, `duration` | Text/Graphics |
| `Gap` | Timeline spacers | `offset`, `duration` | None |

## Media Type Architecture

### Critical Media Type Distinctions

The framework handles three fundamental media types with different architectural requirements:

#### Images (PNG/JPG/GIF)
- **Asset Duration**: `"0s"` (timeless)
- **Format**: NO `frameDuration` attribute
- **Timeline Element**: `<video>` wrapper
- **Effects**: Simple transforms only
- **Animation**: Limited keyframe support

```xml
<!-- Image Asset -->
<asset id="r2" duration="0s" hasVideo="1" format="r3"/>
<format id="r3" width="1280" height="720"/>
<video ref="r2" duration="240240/24000s"/>
```

#### Videos (MP4/MOV/AVI)
- **Asset Duration**: Actual file duration  
- **Format**: Required `frameDuration` attribute
- **Timeline Element**: `<asset-clip>` wrapper
- **Effects**: Full effect support
- **Animation**: Complete keyframe support

```xml
<!-- Video Asset -->
<asset id="r4" duration="14122857/100000s" hasVideo="1" hasAudio="1" format="r5"/>
<format id="r5" frameDuration="1001/30000s" width="1920" height="1080"/>
<asset-clip ref="r4" duration="373400/3000s"/>
```

#### Audio (MP3/WAV/AAC)
- **Asset Duration**: Actual file duration
- **Format**: Audio-specific attributes
- **Timeline Element**: `<asset-clip>` wrapper  
- **Effects**: Audio effects only
- **Animation**: Volume/pan keyframes

### Media Type Detection Logic

```go
func detectMediaType(filePath string) (MediaType, error) {
    ext := strings.ToLower(filepath.Ext(filePath))
    
    switch ext {
    case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff":
        return MediaTypeImage, nil
    case ".mp4", ".mov", ".avi", ".mkv", ".m4v":
        return MediaTypeVideo, nil  
    case ".mp3", ".wav", ".aac", ".m4a", ".flac":
        return MediaTypeAudio, nil
    default:
        return MediaTypeUnknown, fmt.Errorf("unsupported media type: %s", ext)
    }
}
```

## Resource Management

### ID Generation Strategy

FCP requires sequential resource IDs following the pattern `r1`, `r2`, `r3`, etc. The ResourceRegistry manages this:

```go
func (r *ResourceRegistry) ReserveIDs(count int) []string {
    ids := make([]string, count)
    for i := 0; i < count; i++ {
        id := fmt.Sprintf("r%d", r.nextResourceID)
        r.usedIDs[id] = true
        ids[i] = id
        r.nextResourceID++
    }
    return ids
}
```

### UID Management for FCP Compatibility

FCP caches media UIDs in its library database. The framework ensures consistent UIDs:

```go
func generateUID(filename string) string {
    // Use filename (not full path) for consistent UIDs
    basename := filepath.Base(filename)
    hash := sha256.Sum256([]byte(basename))
    return fmt.Sprintf("%X", hash[:8])
}
```

### Asset Reuse Pattern

When the same media file is used multiple times, reuse the same asset to prevent UID collisions:

```go
type AssetManager struct {
    createdAssets map[string]string // filepath -> assetID
    registry      *ResourceRegistry
}

func (am *AssetManager) GetOrCreateAsset(filepath string) (string, error) {
    if existingID, exists := am.createdAssets[filepath]; exists {
        return existingID, nil // Reuse existing asset
    }
    
    // Create new asset
    assetID := am.registry.ReserveIDs(1)[0]
    asset := createAssetFromFile(filepath, assetID)
    am.createdAssets[filepath] = assetID
    
    return assetID, nil
}
```

## Timeline Architecture

### Spine Structure and Ordering

The spine maintains chronological ordering of timeline elements:

```go
func (s Spine) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
    // Collect all elements with their offsets
    var elements []elementWithOffset
    
    // Add all element types
    for _, clip := range s.AssetClips {
        elements = append(elements, elementWithOffset{
            offset:  parseFCPDurationForSort(clip.Offset),
            element: clip,
        })
    }
    
    // Sort by offset and encode
    sort.Slice(elements, func(i, j int) bool {
        return elements[i].offset < elements[j].offset
    })
    
    for _, elem := range elements {
        if err := e.Encode(elem.element); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Lane System

FCP supports vertical layering through lanes:

- **Lane 0** (or no lane): Main timeline layer
- **Lane 1, 2, 3...**: Layers above main timeline
- **Lane -1, -2, -3...**: Layers below main timeline (rarely used)

```xml
<spine>
    <!-- Main layer -->
    <video ref="r2" offset="0s" duration="240240/24000s"/>
    
    <!-- Upper layer -->
    <video ref="r3" lane="1" offset="60060/24000s" duration="120120/24000s"/>
    
    <!-- Even higher layer -->
    <title ref="r4" lane="2" offset="50050/24000s" duration="140140/24000s"/>
</spine>
```

## Timing and Duration System

### FCP's Rational Time System

FCP uses a rational number system based on 24000/1001 timebase:

- **Frame Rate**: 23.976023976... fps (not exactly 24fps)
- **Frame Duration**: 1001/24000 seconds per frame
- **Timebase**: 24000 (denominator)
- **Frame Increment**: 1001 (numerator increment per frame)

### Frame-Aligned Duration Conversion

```go
const (
    FCPTimebase      = 24000
    FCPFrameDuration = 1001  
    FCPFrameRate     = 23.976023976023976
)

func ConvertSecondsToFCPDuration(seconds float64) string {
    if seconds == 0 {
        return "0s"
    }
    
    // Calculate exact frame count
    frames := int(math.Round(seconds * FCPFrameRate))
    
    // Convert to FCP's rational format
    numerator := frames * FCPFrameDuration
    
    return fmt.Sprintf("%d/%ds", numerator, FCPTimebase)
}
```

### Timeline Synchronization

All timeline elements must be frame-aligned and properly synchronized:

```go
func ValidateFrameAlignment(duration string) error {
    if !strings.Contains(duration, "/") {
        return fmt.Errorf("duration must be in rational format: %s", duration)
    }
    
    parts := strings.Split(strings.TrimSuffix(duration, "s"), "/")
    numerator, _ := strconv.Atoi(parts[0])
    denominator, _ := strconv.Atoi(parts[1])
    
    if denominator != FCPTimebase {
        return fmt.Errorf("wrong timebase, expected %d, got %d", FCPTimebase, denominator)
    }
    
    if numerator%FCPFrameDuration != 0 {
        return fmt.Errorf("duration not frame-aligned: %s", duration)
    }
    
    return nil
}
```

## Error Prevention Architecture

### Validation Layer

The framework includes comprehensive validation to prevent common errors:

```go
func ValidateClaudeCompliance(fcpxml *FCPXML) error {
    errors := []string{}
    
    // Validate resource references
    if err := validateResourceReferences(fcpxml); err != nil {
        errors = append(errors, err.Error())
    }
    
    // Validate media type consistency  
    if err := validateMediaTypes(fcpxml); err != nil {
        errors = append(errors, err.Error())
    }
    
    // Validate frame alignment
    if err := validateFrameAlignment(fcpxml); err != nil {
        errors = append(errors, err.Error())
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }
    
    return nil
}
```

### Common Anti-Patterns Detection

The framework detects and prevents common anti-patterns:

1. **String template usage** - Detected through static analysis
2. **Wrong media type handling** - Validated during asset creation
3. **Non-frame-aligned durations** - Validated during conversion
4. **Missing resource references** - Validated during XML generation
5. **UID collisions** - Prevented through asset reuse patterns

## Package Structure

```
fcp/
├── types.go              # Core FCPXML struct definitions
├── generator.go          # Main generation functions
├── registry.go           # Resource management
├── transaction.go        # Transaction system
├── ids.go               # ID generation utilities
├── generator_*_test.go  # Comprehensive test suites
└── test_*.fcpxml        # Reference XML files
```

### Dependency Flow

```
User Code
    ↓
generator.go (public API)
    ↓
registry.go (resource management)
    ↓ 
transaction.go (atomic operations)
    ↓
types.go (XML structures) 
    ↓
xml.MarshalIndent() (Go standard library)
```

## Extensibility Points

### Custom Media Types

Add support for new media types by extending the detection logic:

```go
case ".webm", ".ogg":
    return MediaTypeVideo, nil
case ".opus", ".weba":
    return MediaTypeAudio, nil
```

### Custom Effects

Add new effects by registering verified UIDs:

```go
var verifiedEffectUIDs = map[string]bool{
    "FFGaussianBlur":    true,
    "FFNewCustomEffect": true, // Add new verified effects
}
```

### Custom Validation Rules

Extend validation by adding new rule functions:

```go
func validateCustomRules(fcpxml *FCPXML) error {
    // Custom validation logic
    return nil
}
```

This architecture ensures robust, maintainable, and FCP-compatible FCPXML generation while preventing the most common causes of import failures and crashes.