# Testing and Debugging

This document covers comprehensive testing strategies, validation procedures, and debugging techniques for FCPXML generation.

## 🚨 MANDATORY: Required Tests

**ALWAYS run these tests before using generated FCPXML:**

### 1. FCP Package Tests
```bash
cd fcp && go test
```
**MUST pass** - These tests validate core FCPXML generation logic

### 2. XML Validation
```bash
xmllint output.fcpxml --noout
```
**MUST pass** - Validates XML structure and DTD compliance

### 3. Final Cut Pro Import Test
Import the generated FCPXML into actual Final Cut Pro
**MUST import without errors** - Ultimate validation

## Comprehensive Validation System

### Built-in Validation Function

```go
func ValidateClaudeCompliance(fcpxml *fcp.FCPXML) error {
    errors := []string{}
    
    // 1. Validate resource references
    if err := validateResourceReferences(fcpxml); err != nil {
        errors = append(errors, fmt.Sprintf("Resource references: %v", err))
    }
    
    // 2. Validate media type consistency
    if err := validateMediaTypes(fcpxml); err != nil {
        errors = append(errors, fmt.Sprintf("Media types: %v", err))
    }
    
    // 3. Validate frame alignment
    if err := validateFrameAlignment(fcpxml); err != nil {
        errors = append(errors, fmt.Sprintf("Frame alignment: %v", err))
    }
    
    // 4. Validate ID uniqueness
    if err := validateIDUniqueness(fcpxml); err != nil {
        errors = append(errors, fmt.Sprintf("ID uniqueness: %v", err))
    }
    
    // 5. Validate effect UIDs
    if err := validateEffectUIDs(fcpxml); err != nil {
        errors = append(errors, fmt.Sprintf("Effect UIDs: %v", err))
    }
    
    // 6. Validate keyframe attributes
    if err := validateKeyframeAttributes(fcpxml); err != nil {
        errors = append(errors, fmt.Sprintf("Keyframe attributes: %v", err))
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed:\n- %s", strings.Join(errors, "\n- "))
    }
    
    return nil
}
```

## Validation Rule Implementations

### Resource Reference Validation

**Every `ref` attribute must point to an existing resource:**

```go
func validateResourceReferences(fcpxml *fcp.FCPXML) error {
    // Build resource ID map
    resourceIDs := make(map[string]bool)
    
    for _, asset := range fcpxml.Resources.Assets {
        resourceIDs[asset.ID] = true
    }
    for _, format := range fcpxml.Resources.Formats {
        resourceIDs[format.ID] = true
    }
    for _, effect := range fcpxml.Resources.Effects {
        resourceIDs[effect.ID] = true
    }
    for _, media := range fcpxml.Resources.Media {
        resourceIDs[media.ID] = true
    }
    
    // Validate all references
    errors := []string{}
    
    for _, event := range fcpxml.Library.Events {
        for _, project := range event.Projects {
            for _, sequence := range project.Sequences {
                if err := validateSpineReferences(sequence.Spine, resourceIDs); err != nil {
                    errors = append(errors, err.Error())
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("missing references: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

func validateSpineReferences(spine fcp.Spine, resourceIDs map[string]bool) error {
    errors := []string{}
    
    // Check asset-clip references
    for _, clip := range spine.AssetClips {
        if !resourceIDs[clip.Ref] {
            errors = append(errors, fmt.Sprintf("asset-clip ref='%s' not found", clip.Ref))
        }
        if clip.Format != "" && !resourceIDs[clip.Format] {
            errors = append(errors, fmt.Sprintf("asset-clip format='%s' not found", clip.Format))
        }
    }
    
    // Check video references
    for _, video := range spine.Videos {
        if !resourceIDs[video.Ref] {
            errors = append(errors, fmt.Sprintf("video ref='%s' not found", video.Ref))
        }
    }
    
    // Check title references
    for _, title := range spine.Titles {
        if !resourceIDs[title.Ref] {
            errors = append(errors, fmt.Sprintf("title ref='%s' not found", title.Ref))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf(strings.Join(errors, "; "))
    }
    
    return nil
}
```

### Media Type Consistency Validation

**Images must use Video elements, videos must use AssetClip elements:**

```go
func validateMediaTypes(fcpxml *fcp.FCPXML) error {
    // Build asset type map
    assetTypes := make(map[string]string) // assetID -> mediaType
    
    for _, asset := range fcpxml.Resources.Assets {
        mediaType := detectAssetMediaType(asset)
        assetTypes[asset.ID] = mediaType
    }
    
    errors := []string{}
    
    for _, event := range fcpxml.Library.Events {
        for _, project := range event.Projects {
            for _, sequence := range project.Sequences {
                if err := validateSpineMediaTypes(sequence.Spine, assetTypes); err != nil {
                    errors = append(errors, err.Error())
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("media type violations: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

func detectAssetMediaType(asset fcp.Asset) string {
    // Images have duration="0s"
    if asset.Duration == "0s" {
        return "image"
    }
    
    // Videos have audio properties
    if asset.HasAudio == "1" {
        return "video"
    }
    
    // Audio-only files
    if asset.HasVideo != "1" && asset.HasAudio == "1" {
        return "audio"
    }
    
    return "video" // Default assumption
}

func validateSpineMediaTypes(spine fcp.Spine, assetTypes map[string]string) error {
    errors := []string{}
    
    // Check that images use Video elements
    for _, video := range spine.Videos {
        if mediaType, exists := assetTypes[video.Ref]; exists {
            if mediaType != "image" {
                errors = append(errors, fmt.Sprintf("video element with ref='%s' should be asset-clip (media type: %s)", video.Ref, mediaType))
            }
        }
    }
    
    // Check that videos/audio use AssetClip elements
    for _, clip := range spine.AssetClips {
        if mediaType, exists := assetTypes[clip.Ref]; exists {
            if mediaType == "image" {
                errors = append(errors, fmt.Sprintf("asset-clip with ref='%s' should be video element (media type: image)", clip.Ref))
            }
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf(strings.Join(errors, "; "))
    }
    
    return nil
}
```

### Frame Alignment Validation

**All durations must be frame-aligned:**

```go
func validateFrameAlignment(fcpxml *fcp.FCPXML) error {
    errors := []string{}
    
    for _, event := range fcpxml.Library.Events {
        for _, project := range event.Projects {
            for _, sequence := range project.Sequences {
                // Validate sequence duration
                if err := ValidateFrameAlignment(sequence.Duration); err != nil {
                    errors = append(errors, fmt.Sprintf("sequence duration: %v", err))
                }
                
                // Validate spine elements
                if err := validateSpineFrameAlignment(sequence.Spine); err != nil {
                    errors = append(errors, err.Error())
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("frame alignment errors: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

func validateSpineFrameAlignment(spine fcp.Spine) error {
    errors := []string{}
    
    // Validate asset-clips
    for _, clip := range spine.AssetClips {
        if err := ValidateFrameAlignment(clip.Offset); err != nil {
            errors = append(errors, fmt.Sprintf("asset-clip offset: %v", err))
        }
        if err := ValidateFrameAlignment(clip.Duration); err != nil {
            errors = append(errors, fmt.Sprintf("asset-clip duration: %v", err))
        }
    }
    
    // Validate videos
    for _, video := range spine.Videos {
        if err := ValidateFrameAlignment(video.Offset); err != nil {
            errors = append(errors, fmt.Sprintf("video offset: %v", err))
        }
        if err := ValidateFrameAlignment(video.Duration); err != nil {
            errors = append(errors, fmt.Sprintf("video duration: %v", err))
        }
    }
    
    // Validate titles
    for _, title := range spine.Titles {
        if err := ValidateFrameAlignment(title.Offset); err != nil {
            errors = append(errors, fmt.Sprintf("title offset: %v", err))
        }
        if err := ValidateFrameAlignment(title.Duration); err != nil {
            errors = append(errors, fmt.Sprintf("title duration: %v", err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf(strings.Join(errors, "; "))
    }
    
    return nil
}

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
    
    if denominator != 24000 {
        return fmt.Errorf("wrong timebase, expected 24000, got %d", denominator)
    }
    
    if numerator%1001 != 0 {
        return fmt.Errorf("duration not frame-aligned: %s", duration)
    }
    
    return nil
}
```

### ID Uniqueness Validation

**All resource IDs must be unique:**

```go
func validateIDUniqueness(fcpxml *fcp.FCPXML) error {
    usedIDs := make(map[string]string) // ID -> resource type
    errors := []string{}
    
    // Check assets
    for _, asset := range fcpxml.Resources.Assets {
        if existingType, exists := usedIDs[asset.ID]; exists {
            errors = append(errors, fmt.Sprintf("duplicate ID '%s' in asset (already used in %s)", asset.ID, existingType))
        } else {
            usedIDs[asset.ID] = "asset"
        }
    }
    
    // Check formats
    for _, format := range fcpxml.Resources.Formats {
        if existingType, exists := usedIDs[format.ID]; exists {
            errors = append(errors, fmt.Sprintf("duplicate ID '%s' in format (already used in %s)", format.ID, existingType))
        } else {
            usedIDs[format.ID] = "format"
        }
    }
    
    // Check effects
    for _, effect := range fcpxml.Resources.Effects {
        if existingType, exists := usedIDs[effect.ID]; exists {
            errors = append(errors, fmt.Sprintf("duplicate ID '%s' in effect (already used in %s)", effect.ID, existingType))
        } else {
            usedIDs[effect.ID] = "effect"
        }
    }
    
    // Check media
    for _, media := range fcpxml.Resources.Media {
        if existingType, exists := usedIDs[media.ID]; exists {
            errors = append(errors, fmt.Sprintf("duplicate ID '%s' in media (already used in %s)", media.ID, existingType))
        } else {
            usedIDs[media.ID] = "media"
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("ID uniqueness violations: %s", strings.Join(errors, "; "))
    }
    
    return nil
}
```

### Effect UID Validation

**Only verified effect UIDs are allowed:**

```go
var verifiedEffectUIDs = map[string]bool{
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

func validateEffectUIDs(fcpxml *fcp.FCPXML) error {
    errors := []string{}
    
    for _, effect := range fcpxml.Resources.Effects {
        if err := validateEffectUID(effect.UID); err != nil {
            errors = append(errors, fmt.Sprintf("effect '%s': %v", effect.ID, err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("effect UID violations: %s", strings.Join(errors, "; "))
    }
    
    return nil
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

### Keyframe Attribute Validation

**Validate keyframe attributes according to parameter type:**

```go
func validateKeyframeAttributes(fcpxml *fcp.FCPXML) error {
    errors := []string{}
    
    for _, event := range fcpxml.Library.Events {
        for _, project := range event.Projects {
            for _, sequence := range project.Sequences {
                if err := validateSpineKeyframes(sequence.Spine); err != nil {
                    errors = append(errors, err.Error())
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("keyframe attribute violations: %s", strings.Join(errors, "; "))
    }
    
    return nil
}

func validateSpineKeyframes(spine fcp.Spine) error {
    errors := []string{}
    
    // Validate asset-clip keyframes
    for _, clip := range spine.AssetClips {
        if clip.AdjustTransform != nil {
            for _, param := range clip.AdjustTransform.Params {
                if err := validateKeyframes(param); err != nil {
                    errors = append(errors, fmt.Sprintf("asset-clip '%s': %v", clip.Name, err))
                }
            }
        }
    }
    
    // Validate video keyframes
    for _, video := range spine.Videos {
        if video.AdjustTransform != nil {
            for _, param := range video.AdjustTransform.Params {
                if err := validateKeyframes(param); err != nil {
                    errors = append(errors, fmt.Sprintf("video '%s': %v", video.Name, err))
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf(strings.Join(errors, "; "))
    }
    
    return nil
}

func validateKeyframes(param fcp.Param) error {
    if param.KeyframeAnimation == nil {
        return nil
    }
    
    for i, keyframe := range param.KeyframeAnimation.Keyframes {
        switch param.Name {
        case "position":
            if keyframe.Interp != "" || keyframe.Curve != "" {
                return fmt.Errorf("position keyframe %d has unsupported attributes", i)
            }
            
        case "scale", "rotation", "anchor":
            if keyframe.Interp != "" {
                return fmt.Errorf("%s keyframe %d has unsupported interp attribute", param.Name, i)
            }
            if keyframe.Curve != "" && !isValidCurve(keyframe.Curve) {
                return fmt.Errorf("%s keyframe %d has invalid curve: %s", param.Name, i, keyframe.Curve)
            }
            
        case "opacity", "volume":
            if keyframe.Interp != "" && !isValidInterp(keyframe.Interp) {
                return fmt.Errorf("%s keyframe %d has invalid interp: %s", param.Name, i, keyframe.Interp)
            }
            if keyframe.Curve != "" && !isValidCurve(keyframe.Curve) {
                return fmt.Errorf("%s keyframe %d has invalid curve: %s", param.Name, i, keyframe.Curve)
            }
        }
    }
    
    return nil
}

func isValidCurve(curve string) bool {
    validCurves := []string{"linear", "smooth", "hold"}
    for _, valid := range validCurves {
        if curve == valid {
            return true
        }
    }
    return false
}

func isValidInterp(interp string) bool {
    validInterps := []string{"linear", "easeIn", "easeOut", "easeInOut"}
    for _, valid := range validInterps {
        if interp == valid {
            return true
        }
    }
    return false
}
```

## Common Debugging Patterns

### XML Output Inspection

```go
func debugPrintXML(fcpxml *fcp.FCPXML) {
    xmlData, err := xml.MarshalIndent(fcpxml, "", "    ")
    if err != nil {
        fmt.Printf("Failed to marshal XML: %v\n", err)
        return
    }
    
    fmt.Println("Generated FCPXML:")
    fmt.Println(string(xmlData))
}

func saveDebugXML(fcpxml *fcp.FCPXML, filename string) error {
    xmlData, err := xml.MarshalIndent(fcpxml, "", "    ")
    if err != nil {
        return fmt.Errorf("failed to marshal XML: %v", err)
    }
    
    return os.WriteFile(filename, xmlData, 0644)
}
```

### Resource Inspection

```go
func debugPrintResources(fcpxml *fcp.FCPXML) {
    fmt.Printf("Resources Summary:\n")
    fmt.Printf("- Assets: %d\n", len(fcpxml.Resources.Assets))
    fmt.Printf("- Formats: %d\n", len(fcpxml.Resources.Formats))
    fmt.Printf("- Effects: %d\n", len(fcpxml.Resources.Effects))
    fmt.Printf("- Media: %d\n", len(fcpxml.Resources.Media))
    
    fmt.Printf("\nAsset Details:\n")
    for _, asset := range fcpxml.Resources.Assets {
        fmt.Printf("- %s: %s (duration=%s, hasVideo=%s, hasAudio=%s)\n", 
            asset.ID, asset.Name, asset.Duration, asset.HasVideo, asset.HasAudio)
    }
    
    fmt.Printf("\nFormat Details:\n")
    for _, format := range fcpxml.Resources.Formats {
        fmt.Printf("- %s: %s (%sx%s, frameDuration=%s)\n", 
            format.ID, format.Name, format.Width, format.Height, format.FrameDuration)
    }
}
```

### Timeline Inspection

```go
func debugPrintTimeline(spine fcp.Spine) {
    fmt.Printf("Timeline Summary:\n")
    fmt.Printf("- AssetClips: %d\n", len(spine.AssetClips))
    fmt.Printf("- Videos: %d\n", len(spine.Videos))
    fmt.Printf("- Titles: %d\n", len(spine.Titles))
    fmt.Printf("- Gaps: %d\n", len(spine.Gaps))
    
    fmt.Printf("\nTimeline Elements (chronological):\n")
    
    // Collect all elements with offsets
    type TimelineElement struct {
        Type     string
        Name     string
        Ref      string
        Offset   string
        Duration string
        Lane     string
    }
    
    var elements []TimelineElement
    
    for _, clip := range spine.AssetClips {
        elements = append(elements, TimelineElement{
            Type: "AssetClip", Name: clip.Name, Ref: clip.Ref,
            Offset: clip.Offset, Duration: clip.Duration, Lane: clip.Lane,
        })
    }
    
    for _, video := range spine.Videos {
        elements = append(elements, TimelineElement{
            Type: "Video", Name: video.Name, Ref: video.Ref,
            Offset: video.Offset, Duration: video.Duration, Lane: video.Lane,
        })
    }
    
    for _, title := range spine.Titles {
        elements = append(elements, TimelineElement{
            Type: "Title", Name: title.Name, Ref: title.Ref,
            Offset: title.Offset, Duration: title.Duration, Lane: title.Lane,
        })
    }
    
    // Sort by offset (simplified)
    for _, elem := range elements {
        laneStr := elem.Lane
        if laneStr == "" {
            laneStr = "0"
        }
        fmt.Printf("- %s '%s' (ref=%s, offset=%s, duration=%s, lane=%s)\n",
            elem.Type, elem.Name, elem.Ref, elem.Offset, elem.Duration, laneStr)
    }
}
```

## Testing Utilities

### Test Data Generation

```go
func createTestFCPXML() *fcp.FCPXML {
    fcpxml := &fcp.FCPXML{
        Version: "1.11",
        Resources: fcp.Resources{
            Assets: []fcp.Asset{
                {
                    ID: "r1", Name: "test.mp4", UID: "TEST123", Duration: "240240/24000s",
                    HasVideo: "1", HasAudio: "1", Format: "r2",
                    MediaRep: fcp.MediaRep{Kind: "original-media", Src: "file:///test.mp4"},
                },
            },
            Formats: []fcp.Format{
                {ID: "r2", FrameDuration: "1001/24000s", Width: "1920", Height: "1080"},
            },
        },
        Library: fcp.Library{
            Events: []fcp.Event{
                {
                    Name: "Test Event",
                    Projects: []fcp.Project{
                        {
                            Name: "Test Project",
                            Sequences: []fcp.Sequence{
                                {
                                    Format: "r2", Duration: "240240/24000s",
                                    Spine: fcp.Spine{
                                        AssetClips: []fcp.AssetClip{
                                            {Ref: "r1", Offset: "0s", Duration: "240240/24000s", Name: "Test Clip"},
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    
    return fcpxml
}
```

### Validation Test Suite

```go
func TestValidationSuite(t *testing.T) {
    testCases := []struct {
        name    string
        fcpxml  *fcp.FCPXML
        wantErr bool
    }{
        {
            name:    "valid FCPXML",
            fcpxml:  createTestFCPXML(),
            wantErr: false,
        },
        {
            name:    "missing asset reference",
            fcpxml:  createInvalidReferenceFCPXML(),
            wantErr: true,
        },
        {
            name:    "wrong media type element",
            fcpxml:  createWrongMediaTypeFCPXML(),
            wantErr: true,
        },
        // Add more test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := ValidateClaudeCompliance(tc.fcpxml)
            if (err != nil) != tc.wantErr {
                t.Errorf("ValidateClaudeCompliance() error = %v, wantErr %v", err, tc.wantErr)
            }
        })
    }
}
```

## Error Classification

### Critical Errors (Crash FCP)
- Wrong media type element usage
- Missing resource references
- Fictional effect UIDs
- Non-frame-aligned durations

### Warning Errors (Import with issues)
- Unsupported keyframe attributes
- Missing media files
- Invalid color values
- Non-standard frame rates

### Info Errors (Suboptimal but functional)
- Inefficient resource usage
- Non-sequential IDs
- Redundant effects

## Debugging Workflow

1. **Generate FCPXML** with your code
2. **Run ValidateClaudeCompliance()** to catch architectural issues
3. **Run xmllint** to validate XML structure
4. **Test import in FCP** for final validation
5. **If import fails**, check console logs and compare with working samples
6. **Fix the generation code** (not the XML output)
7. **Repeat until successful**

## Performance Testing

### Large Timeline Testing

```go
func BenchmarkLargeTimeline(b *testing.B) {
    for i := 0; i < b.N; i++ {
        fcpxml := createLargeTimelineFCPXML(1000) // 1000 elements
        
        if err := ValidateClaudeCompliance(fcpxml); err != nil {
            b.Fatalf("Validation failed: %v", err)
        }
        
        _, err := xml.MarshalIndent(fcpxml, "", "    ")
        if err != nil {
            b.Fatalf("Marshal failed: %v", err)
        }
    }
}
```

**Remember: If FCPXML generation requires more than 1 iteration to work, you're doing it wrong. Proper validation and testing should catch issues before they reach Final Cut Pro.**