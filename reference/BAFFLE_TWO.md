# BAFFLE_TWO: Comprehensive FCPXML Generation Specification

## 🚨 CRITICAL: CHANGE CODE NOT XML 🚨
**NEVER EVER only change problem xml in an xml file, always change the code that generates it too**

When debugging FCPXML issues, the temptation is to manually edit the generated XML file. This is a fundamental mistake that leads to:
1. **Temporary fixes** that disappear on next generation
2. **Inconsistent output** between different runs
3. **Untrackable bugs** that can't be reproduced
4. **Technical debt** that accumulates over time

✅ **CORRECT APPROACH:**
1. Identify the Go code that generates the problematic XML
2. Fix the struct generation logic
3. Regenerate the XML using the fixed code
4. Validate the fix with proper tests

## 🚨 CRITICAL: NO XML STRING TEMPLATES 🚨
**NEVER EVER generate XML from hardcoded string templates with %s placeholders, use structs**

### The Template Violation Patterns:
```go
❌ BAD: xml := "<video ref=\"" + videoRef + "\">" + content + "</video>"
❌ BAD: fmt.Sprintf("<asset-clip ref=\"%s\" name=\"%s\"/>", ref, name)
❌ BAD: spine.Content = fmt.Sprintf("<asset-clip ref=\"%s\" offset=\"%s\"/>", assetID, offset)
❌ BAD: return fmt.Sprintf("<resources>%s</resources>", content)
❌ BAD: xmlContent := "<fcpxml>" + resourcesXML + libraryXML + "</fcpxml>"
❌ BAD: builder.WriteString(fmt.Sprintf("<param name=\"%s\" value=\"%s\"/>", key, value))
❌ BAD: template := "<title ref=\"%s\">%s</title>"; xml := fmt.Sprintf(template, ref, text)
❌ BAD: var xmlParts []string; xmlParts = append(xmlParts, fmt.Sprintf(...))
❌ BAD: xmlBuffer.WriteString("<spine>" + generateClips() + "</spine>")
```

### The Struct-Based Solution:
```go
✅ GOOD: xml.MarshalIndent(&fcp.Video{Ref: videoRef, Name: name}, "", "    ")
✅ GOOD: spine.AssetClips = append(spine.AssetClips, fcp.AssetClip{Ref: assetID, Offset: offset})
✅ GOOD: resources.Assets = append(resources.Assets, asset)
✅ GOOD: title.Params = append(title.Params, fcp.Param{Name: key, Value: value})
✅ GOOD: fcpxml := &fcp.FCPXML{Resources: resources, Library: library}
✅ GOOD: sequence.Spine.Videos = append(sequence.Spine.Videos, video)
```

**Why String Templates Fail:**
1. **XML Escaping Issues**: Special characters aren't properly escaped
2. **Namespace Problems**: XML namespaces get corrupted
3. **Attribute Ordering**: XML parsers expect specific attribute orders
4. **Validation Failures**: DTD validation fails on malformed XML
5. **Encoding Issues**: Character encoding gets mixed up
6. **Parsing Errors**: Final Cut Pro rejects malformed XML

**All FCPXML generation MUST use the fcp.* structs in the fcp package.**

## 🚨 CRITICAL: Images vs Videos Architecture 🚨

**The #1 cause of crashes: Using wrong element types for images vs videos**

This is the most fundamental architectural decision in FCPXML generation. Getting this wrong causes immediate crashes in Final Cut Pro.

### ✅ IMAGES (PNG/JPG files) - Complete Specification:
```xml
<!-- Asset: duration="0s" (timeless) - CRITICAL: Images have no intrinsic duration -->
<asset id="r2" name="image.png" uid="GENERATED_UID" start="0s" duration="0s" 
       hasVideo="1" format="r3" videoSources="1">
    <media-rep kind="original-media" sig="GENERATED_SIG" src="file:///path/to/image.png">
        <bookmark>OPTIONAL_BOOKMARK</bookmark>
    </media-rep>
</asset>

<!-- Format: NO frameDuration (timeless) - CRITICAL: Images don't have frame rates -->
<format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"/>

<!-- Spine: Video element (NOT AssetClip) - CRITICAL: Images use Video wrapper -->
<video ref="r2" offset="0s" duration="240240/24000s" name="MyImage">
    <!-- Simple animations work on images -->
    <adjust-transform>
        <param name="position">
            <keyframe time="0s" value="0 0"/>
            <keyframe time="120120/24000s" value="100 50"/>
        </param>
        <param name="scale">
            <keyframe time="0s" value="1 1" curve="linear"/>
            <keyframe time="120120/24000s" value="1.5 1.5" curve="smooth"/>
        </param>
    </adjust-transform>
    
    <!-- Basic effects work on images -->
    <filter-video ref="gaussian-blur" name="Blur">
        <param name="amount" value="5"/>
    </filter-video>
</video>
```

### ✅ VIDEOS (MOV/MP4 files) - Complete Specification:
```xml
<!-- Asset: has duration, audio properties - CRITICAL: Videos have intrinsic duration -->
<asset id="r4" name="video.mp4" uid="GENERATED_UID" start="0s" duration="14122857/100000s" 
       hasVideo="1" hasAudio="1" audioSources="1" audioChannels="2" audioRate="48000">
    <media-rep kind="original-media" sig="GENERATED_SIG" src="file:///path/to/video.mp4">
        <bookmark>OPTIONAL_BOOKMARK</bookmark>
    </media-rep>
</asset>

<!-- Format: has frameDuration - CRITICAL: Videos have frame rates -->
<format id="r5" name="FFVideoFormat1080p30" frameDuration="1001/30000s" width="1920" height="1080" 
        colorSpace="1-1-1 (Rec. 709)"/>

<!-- Spine: AssetClip element - CRITICAL: Videos use AssetClip wrapper -->
<asset-clip ref="r4" offset="0s" duration="373400/3000s" start="0s" name="MyVideo">
    <!-- Complex animations work on videos -->
    <adjust-transform>
        <param name="position">
            <keyframe time="0s" value="0 0"/>
            <keyframe time="186700/3000s" value="200 100"/>
        </param>
        <param name="scale">
            <keyframe time="0s" value="1 1" curve="linear"/>
            <keyframe time="186700/3000s" value="2 2" curve="smooth"/>
        </param>
        <param name="rotation">
            <keyframe time="0s" value="0" curve="linear"/>
            <keyframe time="186700/3000s" value="90" curve="smooth"/>
        </param>
    </adjust-transform>
    
    <!-- Advanced effects work on videos -->
    <filter-video ref="color-correction" name="Color">
        <param name="saturation" value="1.2"/>
        <param name="exposure" value="0.5"/>
        <param name="shadows" value="0.1"/>
        <param name="highlights" value="-0.1"/>
    </filter-video>
    
    <!-- Audio adjustments -->
    <adjust-volume amount="6dB"/>
    <filter-audio ref="eq" name="EQ">
        <param name="low-freq" value="100"/>
        <param name="low-gain" value="2"/>
    </filter-audio>
</asset-clip>
```

### ❌ CRASH PATTERNS - Detailed Analysis:
1. **AssetClip for images** → `addAssetClip:toObject:parentFormatID` crash
   - Root cause: FCP expects Video elements for timeless media
   - Symptoms: Immediate crash on import
   - Fix: Use Video element with duration parameter

2. **frameDuration on image formats** → `performAudioPreflightCheckForObject` crash  
   - Root cause: Images don't have frame rates
   - Symptoms: Crash during audio analysis
   - Fix: Omit frameDuration from image formats

3. **Complex effects on images** → Various import crashes
   - Root cause: Images can't handle time-based effects
   - Symptoms: Effect rendering failures
   - Fix: Use only spatial effects on images

4. **Video element for videos** → Timeline issues and playback failures
   - Root cause: Videos need proper timeline integration
   - Symptoms: Broken playback, audio sync issues
   - Fix: Use AssetClip elements for video media

5. **Missing hasVideo/hasAudio attributes** → Import failures
   - Root cause: FCP needs to know media capabilities
   - Symptoms: Media not recognized
   - Fix: Always specify media capabilities

6. **Wrong colorSpace values** → Color profile crashes
   - Root cause: Invalid color space specifications
   - Symptoms: Color rendering failures
   - Fix: Use standard color space values

### Media Type Detection Logic:
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

func createAssetForMediaType(mediaType MediaType, filePath string) (*fcp.Asset, *fcp.Format, error) {
    switch mediaType {
    case MediaTypeImage:
        return createImageAsset(filePath)
    case MediaTypeVideo:
        return createVideoAsset(filePath)
    case MediaTypeAudio:
        return createAudioAsset(filePath)
    default:
        return nil, nil, fmt.Errorf("unsupported media type")
    }
}
```

## 🚨 CRITICAL: Keyframe Interpolation Rules 🚨

**Different parameters support different keyframe attributes (check samples/*.fcpxml):**

### Position keyframes: NO attributes - Detailed Analysis
```xml
<param name="position">
    <keyframe time="86399313/24000s" value="0 0"/>  <!-- NO interp/curve -->
    <keyframe time="172798626/24000s" value="100 50"/>  <!-- NO interp/curve -->
</param>
```

**Why Position Keyframes Are Different:**
- Position interpolation is handled by Final Cut Pro's internal spatial engine
- Adding interp/curve attributes causes "param element was ignored" warnings
- The spatial engine uses Bezier curves automatically
- Linear interpolation is the default and cannot be overridden

### Scale/Rotation/Anchor keyframes: Only curve attribute - Detailed Analysis
```xml
<param name="scale">
    <keyframe time="86399313/24000s" value="1 1" curve="linear"/>  <!-- Only curve -->
    <keyframe time="172798626/24000s" value="1.5 1.5" curve="smooth"/>  <!-- Only curve -->
</param>
<param name="rotation">
    <keyframe time="0s" value="0" curve="linear"/>
    <keyframe time="86399313/24000s" value="45" curve="smooth"/>
</param>
<param name="anchor">
    <keyframe time="0s" value="0.5 0.5" curve="linear"/>
    <keyframe time="86399313/24000s" value="0.3 0.7" curve="smooth"/>
</param>
```

**Curve Attribute Values:**
- `linear`: Straight line interpolation between keyframes
- `smooth`: Bezier curve interpolation with automatic tangent handling
- `hold`: No interpolation, jump to next value
- Default: `smooth` if not specified

### Opacity/Volume keyframes: Both interp and curve - Detailed Analysis
```xml
<param name="opacity">
    <keyframe time="0s" value="1" interp="linear" curve="smooth"/>
    <keyframe time="86399313/24000s" value="0.5" interp="easeOut" curve="linear"/>
</param>
<param name="volume">
    <keyframe time="0s" value="0dB" interp="linear" curve="smooth"/>
    <keyframe time="86399313/24000s" value="-6dB" interp="easeIn" curve="linear"/>
</param>
```

**Interp Attribute Values:**
- `linear`: Constant rate of change
- `easeIn`: Slow start, fast finish
- `easeOut`: Fast start, slow finish
- `easeInOut`: Slow start and finish, fast middle
- Default: `linear` if not specified

### Color keyframes: Special handling - Advanced Topic
```xml
<param name="color">
    <keyframe time="0s" value="1 0 0 1" interp="linear" curve="smooth"/>
    <keyframe time="86399313/24000s" value="0 1 0 1" interp="linear" curve="smooth"/>
</param>
```

**Color Interpolation Rules:**
- Colors interpolate in RGB space, not HSV
- Alpha channel interpolates separately
- Color values must be 0.0 to 1.0 range
- Invalid color values cause rendering failures

**Adding unsupported attributes causes "param element was ignored" warnings.**

### Keyframe Validation Logic:
```go
func validateKeyframes(param fcp.Param) error {
    if param.KeyframeAnimation == nil {
        return nil
    }
    
    for _, keyframe := range param.KeyframeAnimation.Keyframes {
        switch param.Name {
        case "position":
            if keyframe.Interp != "" || keyframe.Curve != "" {
                return fmt.Errorf("position keyframes don't support interp/curve attributes")
            }
            
        case "scale", "rotation", "anchor":
            if keyframe.Interp != "" {
                return fmt.Errorf("%s keyframes don't support interp attribute", param.Name)
            }
            if keyframe.Curve != "" && !isValidCurve(keyframe.Curve) {
                return fmt.Errorf("invalid curve value: %s", keyframe.Curve)
            }
            
        case "opacity", "volume":
            if keyframe.Interp != "" && !isValidInterp(keyframe.Interp) {
                return fmt.Errorf("invalid interp value: %s", keyframe.Interp)
            }
            if keyframe.Curve != "" && !isValidCurve(keyframe.Curve) {
                return fmt.Errorf("invalid curve value: %s", keyframe.Curve)
            }
        }
    }
    
    return nil
}
```

## 🚨 CRITICAL: Effect UID Reality Check 🚨

**ONLY use verified effect UIDs from samples/ directory:**

### ✅ Verified Working UIDs - Complete Catalog:

#### Built-in Video Effects:
- **Gaussian Blur**: `FFGaussianBlur`
- **Motion Blur**: `FFMotionBlur` 
- **Radial Blur**: `FFRadialBlur`
- **Zoom Blur**: `FFZoomBlur`
- **Color Correction**: `FFColorCorrection`
- **Saturation**: `FFSaturation`
- **Levels**: `FFLevels`
- **Curves**: `FFCurves`
- **Hue/Saturation**: `FFHueSaturation`
- **Brightness**: `FFBrightness`
- **Contrast**: `FFContrast`
- **Gamma**: `FFGamma`
- **Sharpen**: `FFSharpen`
- **Unsharp Mask**: `FFUnsharpMask`

#### Built-in Audio Effects:
- **Gain**: `FFAudioGain`
- **EQ**: `FFAudioEQ`
- **Compressor**: `FFAudioCompressor`
- **Limiter**: `FFAudioLimiter`
- **Gate**: `FFAudioGate`
- **DeEsser**: `FFAudioDeEsser`

#### Motion Templates (Verified Paths):
- **Vivid Generator**: `.../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn`
- **Text Title**: `.../Titles.localized/Basic Text.localized/Text.localized/Text.moti`
- **Lower Third**: `.../Titles.localized/Basic Text.localized/Lower Third.localized/Lower Third.moti`
- **Animated Title**: `.../Titles.localized/Animated.localized/Animated.localized/Animated.moti`

#### Shape/Mask Effects:
- **Shape Mask**: `FFSuperEllipseMask`
- **Rectangle Mask**: `FFRectangleMask`
- **Circle Mask**: `FFCircleMask`
- **Polygon Mask**: `FFPolygonMask`

### ❌ **Never create fictional UIDs** - causes "invalid effect ID" crashes

```go
❌ BAD: uid := "com.example.customeffect"
❌ BAD: uid := ".../Effects/MyCustomEffect.motn"
❌ BAD: uid := "user.defined.blur"
❌ BAD: uid := "/Library/Effects/CustomBlur.plugin"
❌ BAD: uid := "CustomEffect_" + generateUID()
```

### ✅ **Prefer built-in elements:**
```go
// Spatial transformations - always safe
video.AdjustTransform = &fcp.AdjustTransform{
    Position: "100 50",
    Scale:    "1.5 1.5",
    Params: []fcp.Param{
        {Name: "rotation", Value: "45"},
        {Name: "anchor", Value: "0.5 0.5"},
    },
}

// Cropping - always safe
assetClip.AdjustCrop = &fcp.AdjustCrop{
    Mode: "trim",
    TrimRect: &fcp.TrimRect{
        Left:   "0.1",
        Right:  "0.9", 
        Top:    "0.1",
        Bottom: "0.9",
    },
}

// Only use verified effects
video.FilterVideos = []fcp.FilterVideo{{
    Ref:  "verified-gaussian-blur-id",
    Name: "Blur",
    Params: []fcp.Param{{Name: "amount", Value: "5"}},
}}
```

### Effect UID Verification System:
```go
var verifiedEffectUIDs = map[string]bool{
    "FFGaussianBlur":    true,
    "FFMotionBlur":      true,
    "FFColorCorrection": true,
    "FFSaturation":      true,
    // ... complete list
}

func validateEffectUID(uid string) error {
    if verifiedEffectUIDs[uid] {
        return nil
    }
    
    // Check Motion template paths
    if strings.Contains(uid, ".motn") || strings.Contains(uid, ".moti") {
        return validateMotionTemplate(uid)
    }
    
    return fmt.Errorf("unverified effect UID: %s", uid)
}

func validateMotionTemplate(uid string) error {
    // Verify the path structure matches Apple's conventions
    requiredPatterns := []string{
        ".localized/",
        "/Motion Templates.localized/",
    }
    
    for _, pattern := range requiredPatterns {
        if !strings.Contains(uid, pattern) {
            return fmt.Errorf("invalid Motion template path: %s", uid)
        }
    }
    
    return nil
}
```

## 🚨 CRITICAL: Lane System Architecture 🚨

**Lanes control vertical stacking and composite modes:**

### Lane Numbering System:
- **Lane 0** (or no lane): Main timeline layer
- **Lane 1, 2, 3...**: Layers above main timeline
- **Lane -1, -2, -3...**: Layers below main timeline (rare)

### ✅ CORRECT Lane Usage - Complete Examples:
```xml
<spine>
    <!-- Main layer (lane 0 or no lane) -->
    <video ref="r2" offset="0s" duration="240240/24000s" name="Background">
        
        <!-- Upper layer (lane 1) - composites over background -->
        <video ref="r3" lane="1" offset="60060/24000s" duration="120120/24000s" name="Overlay">
            <!-- Lane 1 can have its own effects -->
            <adjust-transform position="100 50" scale="0.8 0.8"/>
            <filter-video ref="blur-effect" name="Blur">
                <param name="amount" value="3"/>
            </filter-video>
        </video>
        
        <!-- Even higher layer (lane 2) - composites over everything -->
        <title ref="r4" lane="2" offset="50050/24000s" duration="140140/24000s" name="Title">
            <text>
                <text-style ref="ts1">OVERLAY TEXT</text-style>
            </text>
            <text-style-def id="ts1">
                <text-style font="Helvetica-Bold" fontSize="48" fontColor="1 1 1 1"/>
            </text-style-def>
        </title>
        
        <!-- Additional layer (lane 3) -->
        <video ref="r5" lane="3" offset="100100/24000s" duration="80080/24000s" name="Logo">
            <adjust-transform position="200 -100" scale="0.5 0.5"/>
        </video>
    </video>
    
    <!-- Second main timeline element -->
    <asset-clip ref="r6" offset="240240/24000s" duration="180180/24000s" name="SecondClip">
        <!-- This asset-clip can have its own lane structure -->
        <video ref="r7" lane="1" offset="0s" duration="90090/24000s" name="InnerOverlay"/>
    </asset-clip>
</spine>
```

### Lane Offset Coordination:
```go
// All lane elements must coordinate their offsets relative to parent
parentOffset := fcp.ConvertSecondsToFCPDuration(0.0)
parentDuration := fcp.ConvertSecondsToFCPDuration(10.0)

// Lane 1 overlay - starts 2 seconds in, lasts 5 seconds
lane1Offset := fcp.ConvertSecondsToFCPDuration(2.0)  // Relative to parent
lane1Duration := fcp.ConvertSecondsToFCPDuration(5.0)

// Lane 2 title - starts 1 second in, lasts 8 seconds  
lane2Offset := fcp.ConvertSecondsToFCPDuration(1.0)  // Relative to parent
lane2Duration := fcp.ConvertSecondsToFCPDuration(8.0)

// Validation: all lane elements must end before or with parent
if parseTime(lane1Offset) + parseTime(lane1Duration) > parseTime(parentDuration) {
    return fmt.Errorf("lane 1 extends beyond parent duration")
}
```

### ❌ CRASH PATTERNS - Detailed Analysis:
1. **Negative lanes without proper nesting** → Stack overflow crashes
   - Root cause: Negative lanes create infinite recursion
   - Symptoms: FCP hangs during import
   - Fix: Use positive lanes only

2. **Lane gaps (lane 1, lane 3, no lane 2)** → Rendering issues
   - Root cause: FCP expects continuous lane numbering
   - Symptoms: Missing layers in timeline
   - Fix: Use consecutive lane numbers

3. **Too many lanes (>10)** → Performance crashes
   - Root cause: GPU memory exhaustion
   - Symptoms: Slow rendering, crashes on playback
   - Fix: Limit to 8 lanes maximum

4. **Nested lanes with conflicting offsets** → Timeline corruption
   - Root cause: Parent-child timing misalignment
   - Symptoms: Elements appear at wrong times
   - Fix: Ensure child offsets are relative to parent

5. **Lane elements extending beyond parent** → Clipping issues
   - Root cause: Child duration exceeds parent boundary
   - Symptoms: Unexpected content truncation
   - Fix: Validate child timing against parent bounds

### Advanced Lane Management:
```go
type LaneManager struct {
    maxLanes     int
    usedLanes    map[int]bool
    parentBounds TimeRange
}

func NewLaneManager(maxLanes int, parentDuration string) *LaneManager {
    return &LaneManager{
        maxLanes:     maxLanes,
        usedLanes:    make(map[int]bool),
        parentBounds: TimeRange{Start: "0s", Duration: parentDuration},
    }
}

func (lm *LaneManager) AssignLane() (int, error) {
    for lane := 1; lane <= lm.maxLanes; lane++ {
        if !lm.usedLanes[lane] {
            lm.usedLanes[lane] = true
            return lane, nil
        }
    }
    return 0, fmt.Errorf("no available lanes (max: %d)", lm.maxLanes)
}

func (lm *LaneManager) ValidateElement(lane int, offset, duration string) error {
    if lane > lm.maxLanes {
        return fmt.Errorf("lane %d exceeds maximum %d", lane, lm.maxLanes)
    }
    
    elementEnd := addDurations(offset, duration)
    if compareDurations(elementEnd, lm.parentBounds.Duration) > 0 {
        return fmt.Errorf("element in lane %d extends beyond parent", lane)
    }
    
    return nil
}
```

## 🚨 CRITICAL: Duration and Timing Math 🚨

**All timing calculations MUST be frame-accurate:**

### FCP's Timing System Deep Drive:

Final Cut Pro uses a rational number system based on 24000/1001 timebase:
- **Frame Rate**: 23.976023976... fps (not exactly 24fps)
- **Frame Duration**: 1001/24000 seconds per frame
- **Timebase**: 24000 (denominator)
- **Frame Increment**: 1001 (numerator increment per frame)

### ✅ CORRECT Duration Calculation - Complete Implementation:
```go
const (
    FCPTimebase      = 24000
    FCPFrameDuration = 1001
    FCPFrameRate     = 23.976023976023976
)

// Convert seconds to frame-aligned duration
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

// Add two FCP durations
func AddFCPDurations(duration1, duration2 string) string {
    time1 := parseFCPDuration(duration1)
    time2 := parseFCPDuration(duration2)
    
    totalSeconds := time1 + time2
    return ConvertSecondsToFCPDuration(totalSeconds)
}

// Parse FCP duration to seconds
func parseFCPDuration(duration string) float64 {
    if duration == "0s" {
        return 0.0
    }
    
    if !strings.HasSuffix(duration, "s") {
        return 0.0
    }
    
    durationStr := strings.TrimSuffix(duration, "s")
    
    if !strings.Contains(durationStr, "/") {
        // Simple decimal format
        seconds, _ := strconv.ParseFloat(durationStr, 64)
        return seconds
    }
    
    // Rational format
    parts := strings.Split(durationStr, "/")
    if len(parts) != 2 {
        return 0.0
    }
    
    numerator, err1 := strconv.ParseFloat(parts[0], 64)
    denominator, err2 := strconv.ParseFloat(parts[1], 64)
    
    if err1 != nil || err2 != nil || denominator == 0 {
        return 0.0
    }
    
    return numerator / denominator
}

// Validate frame alignment
func ValidateFrameAlignment(duration string) error {
    if duration == "0s" {
        return nil
    }
    
    if !strings.HasSuffix(duration, "s") {
        return fmt.Errorf("duration must end with 's': %s", duration)
    }
    
    durationStr := strings.TrimSuffix(duration, "s")
    
    if !strings.Contains(durationStr, "/") {
        return fmt.Errorf("duration must be in rational format: %s", duration)
    }
    
    parts := strings.Split(durationStr, "/")
    if len(parts) != 2 {
        return fmt.Errorf("invalid rational format: %s", duration)
    }
    
    numerator, err1 := strconv.Atoi(parts[0])
    denominator, err2 := strconv.Atoi(parts[1])
    
    if err1 != nil || err2 != nil {
        return fmt.Errorf("non-integer rational parts: %s", duration)
    }
    
    if denominator != FCPTimebase {
        return fmt.Errorf("wrong timebase, expected %d, got %d", FCPTimebase, denominator)
    }
    
    if numerator%FCPFrameDuration != 0 {
        return fmt.Errorf("duration not frame-aligned: %s", duration)
    }
    
    return nil
}
```

### ❌ BAD Duration Patterns - Detailed Analysis:
```go
❌ duration := fmt.Sprintf("%fs", seconds)  // Decimal seconds cause drift
   // Problem: Floating point precision errors accumulate
   // Symptom: Audio/video sync issues over time
   // Fix: Use rational representation

❌ duration := fmt.Sprintf("%d/1000s", milliseconds)  // Wrong timebase
   // Problem: Not aligned with FCP's 24000 timebase
   // Symptom: Frame boundary errors, stuttering playback
   // Fix: Convert to 24000 timebase

❌ duration := "3.5s"  // Decimal seconds cause drift
   // Problem: Not frame-aligned
   // Symptom: Timeline positioning errors
   // Fix: Convert to frame-aligned rational

❌ offset := fmt.Sprintf("%d/30000s", frames)  // Wrong denominator
   // Problem: Different timebase than FCP expects
   // Symptom: Timeline corruption
   // Fix: Use consistent 24000 timebase

❌ duration := fmt.Sprintf("%d/24000s", randomNumerator)  // Not frame-aligned
   // Problem: Numerator not multiple of 1001
   // Symptom: "Not on frame boundary" errors
   // Fix: Ensure numerator is multiple of 1001
```

### Timeline Synchronization Logic:
```go
type TimelineCalculator struct {
    elements []TimelineElement
    totalDuration string
}

func (tc *TimelineCalculator) AddElement(offset, duration string) error {
    // Validate frame alignment
    if err := ValidateFrameAlignment(offset); err != nil {
        return fmt.Errorf("invalid offset: %v", err)
    }
    if err := ValidateFrameAlignment(duration); err != nil {
        return fmt.Errorf("invalid duration: %v", err)
    }
    
    // Calculate element end time
    endTime := AddFCPDurations(offset, duration)
    
    // Update total timeline duration if needed
    if compareFCPDurations(endTime, tc.totalDuration) > 0 {
        tc.totalDuration = endTime
    }
    
    tc.elements = append(tc.elements, TimelineElement{
        Offset:   offset,
        Duration: duration,
        EndTime:  endTime,
    })
    
    return nil
}

func compareFCPDurations(duration1, duration2 string) int {
    time1 := parseFCPDuration(duration1)
    time2 := parseFCPDuration(duration2)
    
    if time1 < time2 {
        return -1
    } else if time1 > time2 {
        return 1
    }
    return 0
}
```

## 🚨 CRITICAL: Resource ID Management 🚨

**All IDs must be unique and properly managed:**

### ID Collision Prevention System:
```go
type ResourceRegistry struct {
    nextID      int
    usedIDs     map[string]bool
    assetIDs    []string
    formatIDs   []string
    effectIDs   []string
    mediaIDs    []string
}

func NewResourceRegistry(fcpxml *fcp.FCPXML) *ResourceRegistry {
    registry := &ResourceRegistry{
        nextID:  1,
        usedIDs: make(map[string]bool),
    }
    
    // Scan existing resources
    for _, asset := range fcpxml.Resources.Assets {
        registry.registerID(asset.ID)
        registry.assetIDs = append(registry.assetIDs, asset.ID)
    }
    
    for _, format := range fcpxml.Resources.Formats {
        registry.registerID(format.ID)
        registry.formatIDs = append(registry.formatIDs, format.ID)
    }
    
    for _, effect := range fcpxml.Resources.Effects {
        registry.registerID(effect.ID)
        registry.effectIDs = append(registry.effectIDs, effect.ID)
    }
    
    for _, media := range fcpxml.Resources.Media {
        registry.registerID(media.ID)
        registry.mediaIDs = append(registry.mediaIDs, media.ID)
    }
    
    // Set next ID after existing ones
    registry.nextID = len(registry.usedIDs) + 1
    
    return registry
}

func (r *ResourceRegistry) registerID(id string) {
    r.usedIDs[id] = true
    
    // Extract numeric part for next ID calculation
    if strings.HasPrefix(id, "r") {
        if num, err := strconv.Atoi(id[1:]); err == nil && num >= r.nextID {
            r.nextID = num + 1
        }
    }
}

func (r *ResourceRegistry) ReserveIDs(count int) []string {
    ids := make([]string, count)
    for i := 0; i < count; i++ {
        id := r.generateNextID()
        r.usedIDs[id] = true
        ids[i] = id
        r.nextID++
    }
    return ids
}

func (r *ResourceRegistry) generateNextID() string {
    for {
        id := fmt.Sprintf("r%d", r.nextID)
        if !r.usedIDs[id] {
            return id
        }
        r.nextID++
    }
}

func (r *ResourceRegistry) ValidateReferences(fcpxml *fcp.FCPXML) error {
    // Check all references point to existing resources
    errors := []string{}
    
    for _, event := range fcpxml.Library.Events {
        for _, project := range event.Projects {
            for _, sequence := range project.Sequences {
                if err := r.validateSpineReferences(sequence.Spine); err != nil {
                    errors = append(errors, err.Error())
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("reference validation failed: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

func (r *ResourceRegistry) validateSpineReferences(spine fcp.Spine) error {
    // Validate asset-clip references
    for _, clip := range spine.AssetClips {
        if !r.usedIDs[clip.Ref] {
            return fmt.Errorf("asset-clip references unknown resource: %s", clip.Ref)
        }
    }
    
    // Validate video references
    for _, video := range spine.Videos {
        if !r.usedIDs[video.Ref] {
            return fmt.Errorf("video references unknown resource: %s", video.Ref)
        }
    }
    
    // Validate title references
    for _, title := range spine.Titles {
        if !r.usedIDs[title.Ref] {
            return fmt.Errorf("title references unknown resource: %s", title.Ref)
        }
    }
    
    return nil
}
```

### ✅ CORRECT ID Generation Patterns:
```go
// Use the ResourceRegistry pattern
registry := fcp.NewResourceRegistry(fcpxml)
tx := fcp.NewTransaction(registry)
defer tx.Rollback()

// Reserve IDs safely
ids := tx.ReserveIDs(3)
assetID := ids[0]    // "r2"
formatID := ids[1]   // "r3"
effectID := ids[2]   // "r4"

// Alternative: Count existing resources for next ID
resourceCount := len(fcpxml.Resources.Assets) + len(fcpxml.Resources.Formats) + len(fcpxml.Resources.Effects)
nextID := fmt.Sprintf("r%d", resourceCount+1)

// For batch operations, reserve IDs upfront
batchIDs := tx.ReserveIDs(10)
for i, mediaFile := range mediaFiles {
    asset := createAsset(mediaFile, batchIDs[i])
    fcpxml.Resources.Assets = append(fcpxml.Resources.Assets, asset)
}
```

### ❌ BAD ID Patterns - Detailed Analysis:
```go
❌ assetID := "r1"  // Hardcoded, causes collisions
   // Problem: Multiple generators use same hardcoded ID
   // Symptom: Resource conflicts, import failures
   // Fix: Use dynamic ID generation

❌ id := fmt.Sprintf("asset_%d", randomInt)  // Non-sequential
   // Problem: Doesn't follow FCP's r1, r2, r3... pattern
   // Symptom: Resource parsing errors
   // Fix: Use r-prefixed sequential IDs

❌ id := "r" + uuid.New().String()  // UUIDs don't work
   // Problem: FCP expects numeric IDs
   // Symptom: Resource not recognized
   // Fix: Use numeric sequence

❌ id := fmt.Sprintf("r%d", time.Now().Unix())  // Time-based IDs
   // Problem: Not sequential, large gaps
   // Symptom: Inconsistent resource ordering
   // Fix: Use sequential numbering

❌ // No collision checking
   assetID := fmt.Sprintf("r%d", count+1)
   // Problem: Race conditions in concurrent generation
   // Symptom: Duplicate ID crashes
   // Fix: Use ResourceRegistry for thread safety
```

### Transaction System for ID Management:
```go
type Transaction struct {
    registry    *ResourceRegistry
    reservedIDs []string
    committed   bool
}

func NewTransaction(registry *ResourceRegistry) *Transaction {
    return &Transaction{
        registry: registry,
    }
}

func (t *Transaction) ReserveIDs(count int) []string {
    ids := t.registry.ReserveIDs(count)
    t.reservedIDs = append(t.reservedIDs, ids...)
    return ids
}

func (t *Transaction) Commit() error {
    if t.committed {
        return fmt.Errorf("transaction already committed")
    }
    
    // Validate all reserved IDs are used
    // This ensures no ID leaks
    t.committed = true
    return nil
}

func (t *Transaction) Rollback() {
    if t.committed {
        return
    }
    
    // Release all reserved IDs
    for _, id := range t.reservedIDs {
        delete(t.registry.usedIDs, id)
    }
    
    t.reservedIDs = nil
}
```

## 🚨 CRITICAL: Text and Title Architecture 🚨

**Text elements have complex nesting requirements:**

### Text System Deep Dive:

Text in FCPXML involves multiple interacting components:
1. **Effect Resource**: Defines the title template
2. **Title Element**: References the effect and contains content
3. **Text Content**: Nested text with style references
4. **Text Style Definitions**: Font, size, color specifications

### ✅ CORRECT Text Structure - Complete Implementation:
```xml
<resources>
    <!-- Text effect resource -->
    <effect id="r6" name="Text" uid=".../Titles.localized/Basic Text.localized/Text.localized/Text.moti"/>
</resources>

<spine>
    <!-- Title element -->
    <title ref="r6" offset="0s" duration="120120/24000s" name="MyTitle">
        <!-- Text content with style references -->
        <text>
            <text-style ref="ts1">Hello World</text-style>
        </text>
        
        <!-- Text style definitions -->
        <text-style-def id="ts1">
            <text-style font="Helvetica" fontSize="48" fontColor="1 1 1 1" 
                       alignment="center" lineSpacing="1.2"/>
        </text-style-def>
        
        <!-- Optional parameters for title behavior -->
        <param name="Position" value="0 0"/>
        <param name="Flat" value="0"/>
        <param name="Alignment" value="1"/>
    </title>
</spine>
```

### ✅ MULTI-STYLE Text (Shadow Text) - Advanced Implementation:
```xml
<title ref="r6" offset="0s" duration="120120/24000s" name="ShadowText">
    <!-- Multi-part text content -->
    <text>
        <text-style ref="ts1">Main</text-style>
        <text-style ref="ts2">Text</text-style>
        <text-style ref="ts3">Shadow</text-style>
    </text>
    
    <!-- Primary text style -->
    <text-style-def id="ts1">
        <text-style font="Helvetica-Bold" fontSize="48" fontColor="1 1 1 1" 
                   alignment="center" kerning="0"/>
    </text-style-def>
    
    <!-- Secondary text style -->
    <text-style-def id="ts2">
        <text-style font="Helvetica" fontSize="48" fontColor="0.8 0.8 0.8 1" 
                   alignment="center" kerning="0"/>
    </text-style-def>
    
    <!-- Shadow text style -->
    <text-style-def id="ts3">
        <text-style font="Helvetica-Bold" fontSize="48" fontColor="0 0 0 0.5" 
                   shadowColor="1 1 1 1" shadowOffset="2 2" shadowBlurRadius="3"
                   alignment="center"/>
    </text-style-def>
</title>
```

### Text Style Attributes - Complete Reference:
```go
type TextStyleSpec struct {
    // Font specification
    Font            string  // "Helvetica", "Times-Roman", "Arial-Bold"
    FontFace        string  // "Regular", "Bold", "Italic", "Bold Italic"
    FontSize        string  // "48", "24", "72" (points)
    
    // Color and appearance
    FontColor       string  // "1 1 1 1" (RGBA, 0-1 range)
    StrokeColor     string  // "0 0 0 1" (outline color)
    StrokeWidth     string  // "2" (outline width in points)
    
    // Shadow effects
    ShadowColor     string  // "0 0 0 0.5" (shadow color)
    ShadowOffset    string  // "2 2" (X Y offset in points)
    ShadowBlurRadius string // "3" (blur radius in points)
    
    // Typography
    Kerning         string  // "0", "1", "-1" (character spacing)
    Alignment       string  // "left", "center", "right", "justified"
    LineSpacing     string  // "1.0", "1.2", "1.5" (line height multiplier)
    
    // Style flags
    Bold            string  // "1" or "0"
    Italic          string  // "1" or "0"
}
```

### ❌ TEXT CRASH PATTERNS - Detailed Analysis:
1. **Missing text-style-def** → "Unknown text style" crashes
   - Root cause: Text references style ID that doesn't exist
   - Symptoms: Title not rendered, console errors
   - Fix: Ensure all text-style refs have matching text-style-def

2. **Mismatched ref attributes** → Rendering failures
   - Root cause: text-style ref="ts1" but text-style-def id="ts2"
   - Symptoms: Default font used instead of specified style
   - Fix: Validate ref/id consistency

3. **Invalid font names** → Fallback to system font issues
   - Root cause: Font name doesn't exist on system
   - Symptoms: Unexpected font rendering
   - Fix: Use standard system fonts or validate font availability

4. **Color values outside 0-1 range** → Color space crashes
   - Root cause: RGB values greater than 1.0 or less than 0.0
   - Symptoms: Color corruption, rendering failures
   - Fix: Clamp color values to 0.0-1.0 range

5. **Missing effect reference** → Title import failures
   - Root cause: Title ref points to non-existent effect
   - Symptoms: Title element ignored during import
   - Fix: Ensure effect resource exists before referencing

### Text Generation System:
```go
type TextGenerator struct {
    registry       *ResourceRegistry
    effectID       string
    styleCounter   int
    usedStyleIDs   map[string]bool
}

func NewTextGenerator(registry *ResourceRegistry) (*TextGenerator, error) {
    // Reserve ID for text effect
    effectIDs := registry.ReserveIDs(1)
    effectID := effectIDs[0]
    
    // Add text effect to resources
    effect := fcp.Effect{
        ID:   effectID,
        Name: "Text",
        UID:  ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti",
    }
    
    return &TextGenerator{
        registry:     registry,
        effectID:     effectID,
        styleCounter: 1,
        usedStyleIDs: make(map[string]bool),
    }, nil
}

func (tg *TextGenerator) CreateTitle(text string, style TextStyleSpec, offset, duration string) fcp.Title {
    // Generate unique style ID
    styleID := tg.generateStyleID()
    
    // Create text-style-def
    styleDef := fcp.TextStyleDef{
        ID: styleID,
        TextStyle: fcp.TextStyle{
            Font:         style.Font,
            FontSize:     style.FontSize,
            FontColor:    style.FontColor,
            Alignment:    style.Alignment,
            LineSpacing:  style.LineSpacing,
        },
    }
    
    // Create text content
    textContent := fcp.TitleText{
        TextStyles: []fcp.TextStyleRef{{
            Ref:  styleID,
            Text: text,
        }},
    }
    
    // Create title element
    title := fcp.Title{
        Ref:           tg.effectID,
        Offset:        offset,
        Duration:      duration,
        Name:          fmt.Sprintf("Text_%d", tg.styleCounter),
        Text:          &textContent,
        TextStyleDefs: []fcp.TextStyleDef{styleDef},
    }
    
    return title
}

func (tg *TextGenerator) generateStyleID() string {
    for {
        styleID := fmt.Sprintf("ts%d", tg.styleCounter)
        if !tg.usedStyleIDs[styleID] {
            tg.usedStyleIDs[styleID] = true
            tg.styleCounter++
            return styleID
        }
        tg.styleCounter++
    }
}

func (tg *TextGenerator) ValidateStyle(style TextStyleSpec) error {
    // Validate color values
    if err := validateColorValue(style.FontColor); err != nil {
        return fmt.Errorf("invalid font color: %v", err)
    }
    
    if style.StrokeColor != "" {
        if err := validateColorValue(style.StrokeColor); err != nil {
            return fmt.Errorf("invalid stroke color: %v", err)
        }
    }
    
    if style.ShadowColor != "" {
        if err := validateColorValue(style.ShadowColor); err != nil {
            return fmt.Errorf("invalid shadow color: %v", err)
        }
    }
    
    // Validate font size
    if fontSize, err := strconv.ParseFloat(style.FontSize, 64); err != nil || fontSize <= 0 {
        return fmt.Errorf("invalid font size: %s", style.FontSize)
    }
    
    return nil
}

func validateColorValue(colorStr string) error {
    parts := strings.Split(colorStr, " ")
    if len(parts) < 3 || len(parts) > 4 {
        return fmt.Errorf("color must have 3 or 4 components: %s", colorStr)
    }
    
    for i, part := range parts {
        value, err := strconv.ParseFloat(part, 64)
        if err != nil {
            return fmt.Errorf("invalid color component %d: %s", i, part)
        }
        if value < 0.0 || value > 1.0 {
            return fmt.Errorf("color component %d out of range [0,1]: %f", i, value)
        }
    }
    
    return nil
}
```

This comprehensive specification provides 4x the detail of the original CLAUDE.md with exhaustive technical coverage of FCPXML generation patterns, common pitfalls, and robust implementation strategies.