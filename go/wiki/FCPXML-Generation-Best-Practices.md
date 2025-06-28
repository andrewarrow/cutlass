# FCPXML Generation Best Practices

This document outlines the critical rules and patterns for generating valid FCPXML files that import successfully into Final Cut Pro.

## 🚨 CRITICAL RULES

### 1. NO XML STRING TEMPLATES

**NEVER EVER generate XML from hardcoded string templates with %s placeholders, use structs**

#### ❌ Template Violation Patterns:
```go
// CRITICAL VIOLATIONS - NEVER DO THESE:
xml := "<video ref=\"" + videoRef + "\">" + content + "</video>"
fmt.Sprintf("<asset-clip ref=\"%s\" name=\"%s\"/>", ref, name)
spine.Content = fmt.Sprintf("<asset-clip ref=\"%s\" offset=\"%s\"/>", assetID, offset)
return fmt.Sprintf("<resources>%s</resources>", content)
xmlContent := "<fcpxml>" + resourcesXML + libraryXML + "</fcpxml>"
builder.WriteString(fmt.Sprintf("<param name=\"%s\" value=\"%s\"/>", key, value))
template := "<title ref=\"%s\">%s</title>"; xml := fmt.Sprintf(template, ref, text)
var xmlParts []string; xmlParts = append(xmlParts, fmt.Sprintf(...))
xmlBuffer.WriteString("<spine>" + generateClips() + "</spine>")
```

#### ✅ Struct-Based Solution:
```go
// CORRECT APPROACH:
xml.MarshalIndent(&fcp.Video{Ref: videoRef, Name: name}, "", "    ")
spine.AssetClips = append(spine.AssetClips, fcp.AssetClip{Ref: assetID, Offset: offset})
resources.Assets = append(resources.Assets, asset)
title.Params = append(title.Params, fcp.Param{Name: key, Value: value})
fcpxml := &fcp.FCPXML{Resources: resources, Library: library}
sequence.Spine.Videos = append(sequence.Spine.Videos, video)
```

#### Why String Templates Fail:
1. **XML Escaping Issues**: Special characters aren't properly escaped
2. **Namespace Problems**: XML namespaces get corrupted
3. **Attribute Ordering**: XML parsers expect specific attribute orders
4. **Validation Failures**: DTD validation fails on malformed XML
5. **Encoding Issues**: Character encoding gets mixed up
6. **Parsing Errors**: Final Cut Pro rejects malformed XML

### 2. CHANGE CODE NOT XML

**NEVER EVER only change problem xml in an xml file, always change the code that generates it too**

#### ❌ Wrong Approach:
1. Generate FCPXML with code
2. Manually edit the XML file to fix issues
3. Use the manually edited XML

#### ✅ Correct Approach:
1. Generate FCPXML with code
2. Identify the Go code that generates the problematic XML
3. Fix the struct generation logic
4. Regenerate the XML using the fixed code
5. Validate the fix with proper tests

### 3. MANDATORY TESTING

**ALWAYS run these tests before using generated FCPXML:**

1. **FCP Package Tests**: `cd fcp && go test` - MUST pass
2. **XML Validation**: `xmllint output.fcpxml --noout` - MUST pass
3. **FCP Import Test**: Import into actual Final Cut Pro

## Required Architecture Pattern

**ALWAYS follow this pattern (from working tests):**

```go
func GenerateMyFeature(inputFile, outputFile string) error {
    // 1. Use existing infrastructure  
    fcpxml, err := fcp.GenerateEmpty("")
    if err != nil {
        return fmt.Errorf("failed to create base FCPXML: %v", err)
    }
    
    // 2. Use proper resource management
    registry := fcp.NewResourceRegistry(fcpxml)
    tx := fcp.NewTransaction(registry)
    defer tx.Rollback()
    
    // 3. Add content using existing functions
    if err := fcp.AddImage(fcpxml, imagePath, duration); err != nil {
        return err
    }
    
    // 4. Apply animations (simple transforms only for images)
    imageVideo := &fcpxml.Library.Events[0].Projects[0].Sequences[0].Spine.Videos[0]
    imageVideo.AdjustTransform = createAnimation(duration, startTime)
    
    // 5. Commit and write
    if err := tx.Commit(); err != nil {
        return err
    }
    return fcp.WriteToFile(fcpxml, outputFile)
}
```

## Resource Management Best Practices

### Unique ID Requirements

**NEVER manually generate IDs:**

```go
❌ BAD: assetID := "r1"  // Hardcoded, causes collisions
❌ BAD: id := fmt.Sprintf("asset_%d", randomInt)  // Non-sequential
❌ BAD: id := "r" + uuid.New().String()  // UUIDs don't work
❌ BAD: id := fmt.Sprintf("r%d", time.Now().Unix())  // Time-based IDs

✅ GOOD: Use ResourceRegistry pattern
registry := fcp.NewResourceRegistry(fcpxml)
tx := fcp.NewTransaction(registry)
ids := tx.ReserveIDs(3)
assetID := ids[0]    // "r2"
formatID := ids[1]   // "r3"
effectID := ids[2]   // "r4"
```

### Asset Reuse to Prevent UID Collisions

**Same media file used multiple times MUST reuse same asset:**

```go
❌ BAD: Create new asset for each use (causes UID collisions)
// Multiple assets with same UID = FCP import crash
asset1 := Asset{ID: "r2", UID: "ABC-123", Src: "file.mp4"} 
asset2 := Asset{ID: "r5", UID: "ABC-123", Src: "file.mp4"} // Same UID!

✅ GOOD: Reuse asset, create multiple timeline references
createdAssets := make(map[string]string) // filepath -> assetID
if existingID, exists := createdAssets[filepath]; exists {
    assetID = existingID  // Reuse existing asset
} else {
    assetID = tx.ReserveIDs(1)[0]
    tx.CreateAsset(assetID, filepath, ...)
    createdAssets[filepath] = assetID  // Remember for reuse
}
// Multiple timeline elements can reference same asset:
// <asset-clip ref="r2"... /> and <asset-clip ref="r2"... />
```

### UID Consistency

**FCP caches media file UIDs in its library database:**

```go
// ✅ GOOD: Consistent UIDs (same file = same UID always)
func generateUID(filename string) string {
    basename := filepath.Base(filename) // Use filename, not full path
    hash := sha256.Sum256([]byte(basename))
    return fmt.Sprintf("%X", hash[:8])
}

// ❌ BAD: Path-based UIDs cause "cannot be imported again" errors
func generateUID(fullPath string) string {
    hash := sha256.Sum256([]byte(fullPath)) // Path changes = different UIDs
    return fmt.Sprintf("%X", hash[:8])
}
```

## Duration and Timing Best Practices

### Frame Boundary Alignment

**All durations MUST use `fcp.ConvertSecondsToFCPDuration()`:**

```go
// ✅ GOOD: Frame-aligned duration
duration := fcp.ConvertSecondsToFCPDuration(5.5)  // "132132/24000s"

// ❌ BAD Duration Patterns:
duration := fmt.Sprintf("%fs", seconds)  // Decimal seconds cause drift
duration := fmt.Sprintf("%d/1000s", milliseconds)  // Wrong timebase
duration := "3.5s"  // Decimal seconds cause drift
offset := fmt.Sprintf("%d/30000s", frames)  // Wrong denominator
duration := fmt.Sprintf("%d/24000s", randomNumerator)  // Not frame-aligned
```

### FCP's Timing System

FCP uses a rational number system based on 24000/1001 timebase:

- **Frame Rate**: 23.976023976... fps (not exactly 24fps)
- **Frame Duration**: 1001/24000 seconds per frame
- **Timebase**: 24000 (denominator)
- **Frame Increment**: 1001 (numerator increment per frame)

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

## Media Type Handling

### Video Property Detection

**ALWAYS detect actual video properties instead of hardcoding:**

```go
❌ BAD: Hardcoded properties cause import failures
asset.HasAudio = "1"  // Video might not have audio!
format.Width = "1920" // Video might be 1080×1920 portrait!
format.FrameDuration = "1001/30000s" // Video might be different fps!

✅ GOOD: Use CreateVideoAssetWithDetection() for proper detection
tx.CreateVideoAssetWithDetection(assetID, videoPath, baseName, duration, formatID)
// Automatically detects: width, height, frame rate, audio presence
// Matches samples: portrait videos get correct 1080×1920 dimensions
// Audio-only if file actually has audio tracks
// Frame rate validation: Rejects bogus rates >120fps, maps to standard FCP rates
```

### Asset File Paths

**Final Cut Pro requires absolute file paths:**

```go
❌ BAD: Relative paths cause "missing media" errors
MediaRep{
    Src: "file://./assets/video.mp4",  // Relative path
}

✅ GOOD: Always use absolute paths
absPath, err := filepath.Abs(videoPath)
MediaRep{
    Src: "file://" + absPath,  // Absolute path
}
```

## Transaction Resource Creation

**ALWAYS use transaction methods to create resources:**

```go
❌ BAD: Direct manipulation bypasses transaction
effectID := tx.ReserveIDs(1)[0]
effect := Effect{ID: effectID, Name: "Blur", UID: "FFGaussianBlur"}
fcpxml.Resources.Effects = append(fcpxml.Resources.Effects, effect)
// Result: "Effect ID is invalid" - resource never committed!

✅ GOOD: Use transaction creation methods
effectID := tx.ReserveIDs(1)[0]
tx.CreateEffect(effectID, "Gaussian Blur", "FFGaussianBlur")
// Resource properly managed and committed with tx.Commit()
```

**Why Direct Append Fails:**
- Reserved IDs don't automatically create resources
- Transaction manages resource lifecycle
- Only tx.Commit() adds resources to final FCPXML
- Direct append bypasses validation and registration

## Effect and Animation Best Practices

### Effect UID Reality Check

**ONLY use verified effect UIDs from samples/ directory:**

#### ✅ Verified Working UIDs:
- **Gaussian Blur**: `FFGaussianBlur`
- **Motion Blur**: `FFMotionBlur`
- **Color Correction**: `FFColorCorrection`
- **Saturation**: `FFSaturation`
- **Text Title**: `.../Titles.localized/Basic Text.localized/Text.localized/Text.moti`
- **Shape Mask**: `FFSuperEllipseMask`

#### ❌ Never create fictional UIDs:
```go
❌ BAD: uid := "com.example.customeffect"
❌ BAD: uid := ".../Effects/MyCustomEffect.motn"
❌ BAD: uid := "user.defined.blur"
❌ BAD: uid := "/Library/Effects/CustomBlur.plugin"
❌ BAD: uid := "CustomEffect_" + generateUID()
```

#### ✅ Prefer built-in elements:
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
```

### Keyframe Interpolation Rules

**Different parameters support different keyframe attributes:**

#### Position keyframes: NO attributes
```xml
<param name="position">
    <keyframe time="86399313/24000s" value="0 0"/>  <!-- NO interp/curve -->
</param>
```

#### Scale/Rotation/Anchor keyframes: Only curve attribute
```xml
<param name="scale">
    <keyframe time="86399313/24000s" value="1 1" curve="linear"/>  <!-- Only curve -->
</param>
```

#### Opacity/Volume keyframes: Both interp and curve
```xml
<param name="opacity">
    <keyframe time="0s" value="1" interp="linear" curve="smooth"/>
</param>
```

**Adding unsupported attributes causes "param element was ignored" warnings.**

## Validation Best Practices

### Comprehensive Validation

**Always validate your FCPXML before using:**

```go
func validateFCPXML(fcpxml *fcp.FCPXML) error {
    // 1. Validate resource references
    if err := validateResourceReferences(fcpxml); err != nil {
        return fmt.Errorf("resource validation failed: %v", err)
    }
    
    // 2. Validate frame alignment
    if err := validateFrameAlignment(fcpxml); err != nil {
        return fmt.Errorf("timing validation failed: %v", err)
    }
    
    // 3. Validate media types
    if err := validateMediaTypes(fcpxml); err != nil {
        return fmt.Errorf("media type validation failed: %v", err)
    }
    
    return nil
}
```

### Common Validation Rules

1. **Resource Reference Validation**: Every `ref` attribute must point to an existing resource
2. **Frame Alignment Validation**: All durations must be frame-aligned
3. **Media Type Consistency**: Images use Video elements, videos use AssetClip elements
4. **UID Uniqueness**: No duplicate UIDs within the same FCPXML
5. **Lane Validation**: Lane numbers must be consecutive and within reasonable limits

## Error Prevention Patterns

### Study Existing Tests

**Before writing FCPXML code, review the logic in `fcp/*_test.go` files:**

- `fcp/generate_test.go` - Shows correct resource management patterns
- `fcp/generator_*_test.go` - Shows working animation/effect patterns
- These tests contain proven patterns that prevent crashes

### Common Error Patterns to Check

1. **ID collisions** - Use proper ResourceRegistry/Transaction pattern
2. **Missing resources** - Every `ref=` needs matching `id=`
3. **Wrong element types** - Images use Video, videos use AssetClip
4. **Fictional effect UIDs** - Only use verified UIDs from samples/
5. **Non-frame-aligned durations** - Use ConvertSecondsToFCPDuration()
6. **Path issues** - Use absolute paths for media files
7. **UID inconsistency** - Same file must have same UID always

## Performance Best Practices

### Efficient Resource Management

```go
// ✅ GOOD: Batch ID reservation
ids := tx.ReserveIDs(totalNeeded)
for i, mediaFile := range mediaFiles {
    asset := createAsset(mediaFile, ids[i])
    // Process asset...
}

// ❌ BAD: Individual ID requests
for _, mediaFile := range mediaFiles {
    id := tx.ReserveIDs(1)[0]  // Inefficient
    asset := createAsset(mediaFile, id)
}
```

### Memory Management

```go
// ✅ GOOD: Reuse slices
spine.AssetClips = make([]fcp.AssetClip, 0, expectedCount)
for _, clip := range clips {
    spine.AssetClips = append(spine.AssetClips, clip)
}

// ❌ BAD: Repeated allocations
for _, clip := range clips {
    spine.AssetClips = append(spine.AssetClips, clip) // Grows slice repeatedly
}
```

## Summary

**Key Principle: Follow existing patterns in fcp/ package. If FCPXML generation requires more than 1 iteration to work, you're doing it wrong.**

The most critical rules are:
1. **NO XML string templates** - Use structs only
2. **Change code, not XML** - Fix generation logic, not output
3. **Proper resource management** - Use ResourceRegistry and transactions
4. **Frame alignment** - Use ConvertSecondsToFCPDuration()
5. **Media type consistency** - Images use Video, videos use AssetClip
6. **Verified effects only** - Don't create fictional UIDs
7. **Comprehensive testing** - Validate before using