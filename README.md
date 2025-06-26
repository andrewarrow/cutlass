
## 📚 API Reference: FCPXML Go Structs

This comprehensive reference covers every Go struct used in FCPXML generation and their corresponding XML elements. Understanding these structs is essential for programmatic FCPXML manipulation.

### Core Document Structure

#### FCPXML
The root document container that wraps all FCPXML content.
```go
type FCPXML struct {
    XMLName   xml.Name  `xml:"fcpxml"`
    Version   string    `xml:"version,attr"`      // "1.13" - FCPXML format version
    Resources Resources `xml:"resources"`         // All assets, formats, effects, media
    Library   Library   `xml:"library"`          // Events, projects, sequences
}
```
**XML Output:**
```xml
<fcpxml version="1.13">
    <resources>...</resources>
    <library>...</library>
</fcpxml>
```

### Resource Management System

#### Resources
Container for all reusable assets, formats, effects, and media definitions. Every referenced element in the timeline must have a corresponding resource.
```go
type Resources struct {
    Assets     []Asset     `xml:"asset,omitempty"`      // Video/audio/image files
    Formats    []Format    `xml:"format"`               // Resolution, frame rate specs
    Effects    []Effect    `xml:"effect,omitempty"`     // Motion templates, titles
    Media      []Media     `xml:"media,omitempty"`      // Compound clips, sequences
}
```

#### Asset
Represents a media file (video, audio, image) with all its technical properties.
```go
type Asset struct {
    ID            string    `xml:"id,attr"`              // "r1", "r2" - unique identifier
    Name          string    `xml:"name,attr"`            // "video.mp4" - display name
    UID           string    `xml:"uid,attr"`             // Unique media identifier
    Start         string    `xml:"start,attr"`           // "0s" - media start timecode
    HasVideo      string    `xml:"hasVideo,attr,omitempty"`    // "1" if contains video
    Format        string    `xml:"format,attr,omitempty"`      // Reference to Format ID
    VideoSources  string    `xml:"videoSources,attr,omitempty"` // "1" - number of video tracks
    HasAudio      string    `xml:"hasAudio,attr,omitempty"`     // "1" if contains audio
    AudioSources  string    `xml:"audioSources,attr,omitempty"` // "1" - number of audio tracks
    AudioChannels string    `xml:"audioChannels,attr,omitempty"` // "2" - stereo, "1" - mono
    AudioRate     string    `xml:"audioRate,attr,omitempty"`     // "48000" - sample rate in Hz
    Duration      string    `xml:"duration,attr"`        // "240240/24000s" - FCP duration format
    MediaRep      MediaRep  `xml:"media-rep"`           // File path and signature
    Metadata      *Metadata `xml:"metadata,omitempty"`   // Optional metadata
}
```
**Critical Usage:** Images use `duration="0s"` (timeless), videos have calculated duration.

#### Format
Defines video/audio format specifications like resolution, frame rate, and color space.
```go
type Format struct {
    ID            string `xml:"id,attr"`                    // "r1" - unique identifier
    Name          string `xml:"name,attr,omitempty"`        // "FFVideoFormat1080p30"
    FrameDuration string `xml:"frameDuration,attr,omitempty"` // "1001/30000s" for 29.97fps
    Width         string `xml:"width,attr,omitempty"`       // "1920" - pixels
    Height        string `xml:"height,attr,omitempty"`      // "1080" - pixels
    ColorSpace    string `xml:"colorSpace,attr,omitempty"`  // "1-1-1 (Rec. 709)"
}
```
**Critical Usage:** Images omit `FrameDuration`, videos require it for proper playback.

#### Effect
References Motion templates, built-in effects, or title templates.
```go
type Effect struct {
    ID   string `xml:"id,attr"`        // "r1" - unique identifier
    Name string `xml:"name,attr"`      // "Text", "Blur" - display name
    UID  string `xml:"uid,attr,omitempty"` // ".../Text.moti" - template path
}
```

#### MediaRep
File system reference with security bookmarks for media assets.
```go
type MediaRep struct {
    Kind     string `xml:"kind,attr"`      // "original-media"
    Sig      string `xml:"sig,attr"`       // File signature hash
    Src      string `xml:"src,attr"`       // "file:///path/to/video.mp4" - MUST be absolute
    Bookmark string `xml:"bookmark,omitempty"` // macOS security bookmark
}
```

#### Metadata
Extensible metadata system for storing custom information.
```go
type Metadata struct {
    MDs []MetadataItem `xml:"md"`     // Array of key-value pairs
}

type MetadataItem struct {
    Key   string      `xml:"key,attr"`         // "author", "description"
    Value string      `xml:"value,attr,omitempty"` // Simple values
    Array *StringArray `xml:"array,omitempty"`      // Array values
}

type StringArray struct {
    Strings []string `xml:"string"`   // ["item1", "item2"]
}
```

### Project Structure

#### Library
Top-level container for all projects and events.
```go
type Library struct {
    Location          string            `xml:"location,attr,omitempty"` // Library file path
    Events            []Event           `xml:"event"`                   // Project containers
    SmartCollections  []SmartCollection `xml:"smart-collection,omitempty"` // Dynamic collections
}
```

#### Event
Container for related projects, similar to folders.
```go
type Event struct {
    Name     string    `xml:"name,attr"`           // "My Event" - display name
    UID      string    `xml:"uid,attr,omitempty"`  // Unique identifier
    Projects []Project `xml:"project"`             // Contained projects
}
```

#### Project
Individual project containing sequences (timelines).
```go
type Project struct {
    Name      string     `xml:"name,attr"`              // "My Project"
    UID       string     `xml:"uid,attr,omitempty"`     // Unique identifier
    ModDate   string     `xml:"modDate,attr,omitempty"` // "2023-12-01 10:30:00"
    Sequences []Sequence `xml:"sequence"`               // Timelines
}
```

#### Sequence
The main timeline containing all video/audio elements.
```go
type Sequence struct {
    Format      string `xml:"format,attr"`       // Reference to Format ID
    Duration    string `xml:"duration,attr"`     // "600600/24000s" - total length
    TCStart     string `xml:"tcStart,attr"`      // "0s" - timecode start
    TCFormat    string `xml:"tcFormat,attr"`     // "NDF" - Non-Drop Frame
    AudioLayout string `xml:"audioLayout,attr"`  // "stereo" - audio configuration
    AudioRate   string `xml:"audioRate,attr"`    // "48000" - sample rate
    Spine       Spine  `xml:"spine"`            // Main timeline container
}
```

### Timeline Elements

#### Spine
The main timeline container that holds all video/audio elements in chronological order.
```go
type Spine struct {
    XMLName    xml.Name    `xml:"spine"`
    AssetClips []AssetClip `xml:"asset-clip,omitempty"` // Video/audio files
    Gaps       []Gap       `xml:"gap,omitempty"`        // Empty timeline spaces
    Titles     []Title     `xml:"title,omitempty"`      // Text elements
    Videos     []Video     `xml:"video,omitempty"`      // Generators, shapes, images
}
```
**Critical Feature:** Custom `MarshalXML` automatically sorts elements by timeline offset.

#### AssetClip
References video/audio assets in the timeline. Used for MOV, MP4, WAV files.
```go
type AssetClip struct {
    XMLName         xml.Name         `xml:"asset-clip"`
    Ref             string           `xml:"ref,attr"`          // Asset ID reference
    Lane            string           `xml:"lane,attr,omitempty"` // "1", "2" - layer number
    Offset          string           `xml:"offset,attr"`       // "0s" - timeline position
    Name            string           `xml:"name,attr"`         // Display name
    Start           string           `xml:"start,attr,omitempty"` // "30030/24000s" - in-point
    Duration        string           `xml:"duration,attr"`     // "240240/24000s" - clip length
    Format          string           `xml:"format,attr,omitempty"` // Override format
    TCFormat        string           `xml:"tcFormat,attr,omitempty"` // Timecode format
    AudioRole       string           `xml:"audioRole,attr,omitempty"` // Audio role assignment
    ConformRate     *ConformRate     `xml:"conform-rate,omitempty"`   // Frame rate conversion
    AdjustCrop      *AdjustCrop      `xml:"adjust-crop,omitempty"`    // Cropping parameters
    AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"` // Position/scale/rotation
    NestedAssetClips []AssetClip     `xml:"asset-clip,omitempty"`     // Connected clips
    Titles          []Title          `xml:"title,omitempty"`          // Overlaid titles
    Videos          []Video          `xml:"video,omitempty"`          // Overlaid generators
    FilterVideos    []FilterVideo    `xml:"filter-video,omitempty"`   // Applied effects
}
```

#### Video
Used for generators, shapes, images, and synthetic content. Images use Video (not AssetClip).
```go
type Video struct {
    XMLName xml.Name `xml:"video"`
    Ref           string         `xml:"ref,attr"`              // Asset/Effect ID reference
    Lane          string         `xml:"lane,attr,omitempty"`   // Layer number
    Offset        string         `xml:"offset,attr"`           // Timeline position
    Name          string         `xml:"name,attr"`             // Display name
    Duration      string         `xml:"duration,attr"`         // Element duration
    Start         string         `xml:"start,attr,omitempty"`  // Media in-point
    Params        []Param        `xml:"param,omitempty"`       // Effect parameters
    AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"` // Spatial transforms
    FilterVideos     []FilterVideo   `xml:"filter-video,omitempty"`     // Applied effects
    NestedVideos     []Video     `xml:"video,omitempty"`      // Connected videos
    NestedAssetClips []AssetClip `xml:"asset-clip,omitempty"` // Connected clips
    NestedTitles     []Title     `xml:"title,omitempty"`      // Connected titles
}
```

#### Title
Text elements referencing title effects/templates.
```go
type Title struct {
    XMLName xml.Name `xml:"title"`
    Ref          string         `xml:"ref,attr"`              // Effect ID reference
    Lane         string         `xml:"lane,attr,omitempty"`   // Layer number
    Offset       string         `xml:"offset,attr"`           // Timeline position
    Name         string         `xml:"name,attr"`             // Display name
    Duration     string         `xml:"duration,attr"`         // Text duration
    Start        string         `xml:"start,attr,omitempty"`  // Start time
    Params       []Param        `xml:"param,omitempty"`       // Title parameters
    Text         *TitleText     `xml:"text,omitempty"`        // Text content
    TextStyleDefs []TextStyleDef `xml:"text-style-def,omitempty"` // Font/color definitions
}
```

#### Gap
Represents empty space in the timeline, used for timing control.
```go
type Gap struct {
    XMLName        xml.Name        `xml:"gap"`
    Name           string          `xml:"name,attr"`           // "Gap"
    Offset         string          `xml:"offset,attr"`         // Timeline position
    Duration       string          `xml:"duration,attr"`       // Gap length
    Titles         []Title         `xml:"title,omitempty"`     // Overlaid titles
    GeneratorClips []GeneratorClip `xml:"generator-clip,omitempty"` // Overlaid generators
}
```

### Transform and Effect Systems

#### AdjustTransform
Spatial transformations: position, scale, rotation, and keyframe animations.
```go
type AdjustTransform struct {
    Position string  `xml:"position,attr,omitempty"` // "100 50" - X Y coordinates
    Scale    string  `xml:"scale,attr,omitempty"`    // "1.5 1.5" - X Y scale factors
    Params   []Param `xml:"param,omitempty"`         // Keyframe animations
}
```

#### AdjustCrop
Cropping and trimming controls.
```go
type AdjustCrop struct {
    Mode     string    `xml:"mode,attr"`         // "trim" - crop mode
    TrimRect *TrimRect `xml:"trim-rect,omitempty"` // Crop boundaries
}

type TrimRect struct {
    Left   string `xml:"left,attr,omitempty"`   // "0.1" - left crop (0-1)
    Right  string `xml:"right,attr,omitempty"`  // "0.9" - right crop (0-1)
    Top    string `xml:"top,attr,omitempty"`    // "0.1" - top crop (0-1)
    Bottom string `xml:"bottom,attr,omitempty"` // "0.9" - bottom crop (0-1)
}
```

#### FilterVideo
Video effects and filters applied to clips.
```go
type FilterVideo struct {
    Ref    string  `xml:"ref,attr"`        // Effect ID reference
    Name   string  `xml:"name,attr"`       // "Blur", "Color Correction"
    Params []Param `xml:"param,omitempty"` // Effect parameters
}
```

#### ConformRate
Frame rate conversion and scaling settings.
```go
type ConformRate struct {
    ScaleEnabled string `xml:"scaleEnabled,attr,omitempty"` // "1" - enable scaling
    SrcFrameRate string `xml:"srcFrameRate,attr,omitempty"` // "29.97" - source fps
}
```

### Animation and Parameter System

#### Param
Flexible parameter system supporting static values and keyframe animations.
```go
type Param struct {
    Name               string              `xml:"name,attr"`           // "position", "scale", "opacity"
    Key                string              `xml:"key,attr,omitempty"`  // Alternative to Name
    Value              string              `xml:"value,attr,omitempty"` // "1.0" - static value
    KeyframeAnimation  *KeyframeAnimation  `xml:"keyframeAnimation,omitempty"` // Animated values
    NestedParams       []Param             `xml:"param,omitempty"`     // Hierarchical parameters
}
```

#### KeyframeAnimation
Container for keyframe-based animations.
```go
type KeyframeAnimation struct {
    Keyframes []Keyframe `xml:"keyframe"`    // Animation keyframes
}
```

#### Keyframe
Individual animation keyframe with timing and interpolation.
```go
type Keyframe struct {
    Time   string `xml:"time,attr"`              // "120120/24000s" - keyframe time
    Value  string `xml:"value,attr"`             // "100 50" - keyframe value
    Interp string `xml:"interp,attr,omitempty"`  // "linear", "easeIn", "easeOut"
    Curve  string `xml:"curve,attr,omitempty"`   // "linear", "smooth"
}
```
**Critical Rules:**
- Position keyframes: NO `interp` or `curve` attributes
- Scale/Rotation keyframes: Only `curve` attribute
- Opacity/Volume keyframes: Both `interp` and `curve` supported

### Text and Typography System

#### TitleText
Text content with style references, supports multi-style text.
```go
type TitleText struct {
    TextStyles []TextStyleRef `xml:"text-style"`  // Text segments with styles
}
```

#### TextStyleRef
Individual text segment with style reference.
```go
type TextStyleRef struct {
    Ref  string `xml:"ref,attr"`    // "ts1" - TextStyleDef ID reference
    Text string `xml:",chardata"`   // "Hello World" - actual text content
}
```

#### TextStyleDef
Font and typography style definition.
```go
type TextStyleDef struct {
    ID        string    `xml:"id,attr"`     // "ts1" - unique style identifier
    TextStyle TextStyle `xml:"text-style"`  // Style properties
}
```

#### TextStyle
Complete typography specification with font, colors, and effects.
```go
type TextStyle struct {
    Font            string  `xml:"font,attr"`                    // "Helvetica", "Arial-Bold"
    FontSize        string  `xml:"fontSize,attr"`                // "48" - points
    FontFace        string  `xml:"fontFace,attr,omitempty"`      // "Bold", "Italic"
    FontColor       string  `xml:"fontColor,attr"`               // "1 1 1 1" - RGBA (0-1)
    Bold            string  `xml:"bold,attr,omitempty"`          // "1" - bold flag
    Italic          string  `xml:"italic,attr,omitempty"`        // "1" - italic flag
    StrokeColor     string  `xml:"strokeColor,attr,omitempty"`   // "0 0 0 1" - outline color
    StrokeWidth     string  `xml:"strokeWidth,attr,omitempty"`   // "2" - outline width
    ShadowColor     string  `xml:"shadowColor,attr,omitempty"`   // "0 0 0 0.5" - shadow color
    ShadowOffset    string  `xml:"shadowOffset,attr,omitempty"`  // "2 2" - X Y shadow offset
    ShadowBlurRadius string `xml:"shadowBlurRadius,attr,omitempty"` // "3" - shadow blur
    Kerning         string  `xml:"kerning,attr,omitempty"`       // "0" - character spacing
    Alignment       string  `xml:"alignment,attr,omitempty"`     // "left", "center", "right"
    LineSpacing     string  `xml:"lineSpacing,attr,omitempty"`   // "1.2" - line height
    Params          []Param `xml:"param,omitempty"`              // Additional parameters
}
```

### Additional Elements

#### GeneratorClip
References generator effects (shapes, solids, gradients).
```go
type GeneratorClip struct {
    Ref      string  `xml:"ref,attr"`              // Effect ID reference
    Lane     string  `xml:"lane,attr,omitempty"`   // Layer number
    Offset   string  `xml:"offset,attr"`           // Timeline position
    Name     string  `xml:"name,attr"`             // Display name
    Duration string  `xml:"duration,attr"`         // Generator duration
    Start    string  `xml:"start,attr,omitempty"`  // Start time
    Params   []Param `xml:"param,omitempty"`       // Generator parameters
}
```

#### SmartCollection
Dynamic collections based on search criteria.
```go
type SmartCollection struct {
    Name     string      `xml:"name,attr"`                   // "All Videos"
    Match    string      `xml:"match,attr"`                  // "all", "any"
    Matches  []Match     `xml:"match-clip,omitempty"`        // Clip matching rules
    MediaMatches []MediaMatch `xml:"match-media,omitempty"`   // Media matching rules
    RatingMatches []RatingMatch `xml:"match-ratings,omitempty"` // Rating matching rules
}

type Match struct {
    Rule string `xml:"rule,attr"`     // "is", "contains", "starts with"
    Type string `xml:"type,attr"`     // "name", "format", "duration"
}

type MediaMatch struct {
    Rule string `xml:"rule,attr"`     // Matching rule
    Type string `xml:"type,attr"`     // Media property to match
}

type RatingMatch struct {
    Value string `xml:"value,attr"`   // Rating value to match
}
```

#### Media
Compound clips and nested sequences.
```go
type Media struct {
    ID       string   `xml:"id,attr"`              // Unique identifier
    Name     string   `xml:"name,attr"`            // Display name
    UID      string   `xml:"uid,attr"`             // Unique media identifier
    ModDate  string   `xml:"modDate,attr,omitempty"` // Modification date
    Sequence Sequence `xml:"sequence"`             // Nested sequence content
}
```

#### RefClip
References to other clips or sequences.
```go
type RefClip struct {
    XMLName         xml.Name         `xml:"ref-clip"`
    Ref             string           `xml:"ref,attr"`            // Referenced element ID
    Offset          string           `xml:"offset,attr"`         // Timeline position
    Name            string           `xml:"name,attr"`           // Display name
    Duration        string           `xml:"duration,attr"`       // Clip duration
    AdjustTransform *AdjustTransform `xml:"adjust-transform,omitempty"` // Transforms
    Titles          []Title          `xml:"title,omitempty"`     // Overlaid titles
}
```

### Critical Implementation Rules

#### 1. Images vs Videos Architecture
- **Images (PNG/JPG)**: Use `Video` elements with `duration="0s"` assets, no `frameDuration` in format
- **Videos (MOV/MP4)**: Use `AssetClip` elements with proper duration, include `frameDuration` in format
- **Wrong element type = immediate FCP crash**

#### 2. Resource ID Management
- All IDs must be unique: `r1`, `r2`, `r3`...
- Use `ResourceRegistry` pattern for collision prevention
- Every `ref` attribute must point to existing resource ID

#### 3. Duration and Timing
- Use frame-aligned rational format: `"240240/24000s"`
- FCP timebase: 24000 denominator, 1001 numerator per frame
- Never use decimal seconds: causes sync drift

#### 4. Effect UID Validation
- Only use verified effect UIDs from samples
- Built-in effects: `FFGaussianBlur`, `FFColorCorrection`
- Motion templates: `".../Text.moti"` with proper paths

#### 5. Lane System
- Lane 0 (or no lane): main timeline
- Lane 1, 2, 3...: layers above main
- All lane elements must fit within parent duration

#### 6. Keyframe Interpolation
- Position keyframes: NO attributes
- Scale/Rotation: Only `curve` attribute
- Opacity/Volume: Both `interp` and `curve`

### Validation and Safety

The package includes comprehensive validation:
- `ValidateFrameAlignment()` - Ensures timing precision
- `ValidateReferences()` - Checks all ref attributes
- `ValidateEffectUID()` - Verifies effect availability
- Custom `MarshalXML()` - Maintains chronological order

## 📚 Learn More: FCPXML Resources

Want to dive deeper into the mysterious world of FCPXML? Here are the resources that built Cutlass's expertise:

### Essential Reading
- [FCP.cafe Developer Resources](https://fcp.cafe/developers/fcpxml/) - Community knowledge base
- [Apple FCPXML Reference](https://developer.apple.com/documentation/professional-video-applications/fcpxml-reference) - Official documentation  
- [CommandPost DTD Files](https://github.com/CommandPost/CommandPost/tree/develop/src/extensions/cp/apple/fcpxml/dtd) - XML validation schemas
- [DAWFileKit](https://github.com/orchetect/DAWFileKit) - Swift FCPXML parser

### Installation
```bash
# Clone the future of video generation
git clone https://github.com/your-username/cutlass
cd cutlass
go build

# Start creating magic
./cutlass --help
```

### Quick Start Example

**YouTube Video → Auto-Segmented Clips**  
```bash
./cutlass youtube dQw4w9WgXcQ
./cutlass vtt-clips dQw4w9WgXcQ.en.vtt 00:30_10,01:45_15,02:30_20
# Generates: segmented_video.fcpxml  
# Result: Perfectly timed clips with transition animations
```
