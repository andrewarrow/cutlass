# BAFFLE_THREE: Exhaustive FCPXML Generation Master Specification

This is the most comprehensive FCPXML reference, containing 8x the detail of CLAUDE.md with exhaustive technical coverage, complete implementation patterns, and advanced error prevention strategies.

## 🚨 CRITICAL: CHANGE CODE NOT XML 🚨
**NEVER EVER only change problem xml in an xml file, always change the code that generates it too**

## 🚨 CRITICAL: NO XML STRING TEMPLATES 🚨
**All FCPXML generation MUST use the fcp.* structs in the fcp package.**

## 🚨 CRITICAL: Images vs Videos Architecture 🚨
**The #1 cause of crashes: Using wrong element types for images vs videos**

### ✅ IMAGES (PNG/JPG files):
- Use `<video>` elements in spine
- Asset duration = "0s" (timeless)
- Format has NO frameDuration
- Simple animations only

### ✅ VIDEOS (MOV/MP4 files):
- Use `<asset-clip>` elements in spine  
- Asset has real duration and audio properties
- Format has frameDuration
- Complex effects supported

### ❌ CRASH PATTERNS:
1. **AssetClip for images** → Immediate crash
2. **frameDuration on image formats** → Audio analysis crash
3. **Complex temporal effects on images** → Rendering failures

## 🚨 CRITICAL: Effect UID Reality Check 🚨
**ONLY use verified effect UIDs from samples/ directory**

✅ **Verified Working UIDs:**
- Gaussian Blur: `FFGaussianBlur`
- Text Title: `.../Titles.localized/Basic Text.localized/Text.localized/Text.moti`
- Color Correction: `FFColorCorrection`

❌ **Never create fictional UIDs** - causes crashes

## 🚨 CRITICAL: Resource ID Management 🚨
**Use ResourceRegistry pattern for unique IDs:**

```go
registry := fcp.NewResourceRegistry(fcpxml)
tx := fcp.NewTransaction(registry)
ids := tx.ReserveIDs(3)  // Get r2, r3, r4
```

## 🚨 CRITICAL: Duration Math 🚨
**All durations MUST use frame-aligned rational format:**

```go
duration := fcp.ConvertSecondsToFCPDuration(3.5)  // "84084/24000s"
```

## 🚨 CRITICAL: Lane System 🚨
**Lanes control vertical stacking:**
- Lane 0 (main): Base layer
- Lane 1, 2, 3...: Upper layers
- Max 8 lanes for stability
- Child offsets relative to parent

## 🚨 CRITICAL: Text Architecture 🚨
**Text requires effect + title + text-style-def:**

```xml
<effect id="r6" name="Text" uid=".../Text.moti"/>
<title ref="r6">
    <text><text-style ref="ts1">Hello</text-style></text>
    <text-style-def id="ts1">
        <text-style font="Helvetica" fontSize="48" fontColor="1 1 1 1"/>
    </text-style-def>
</title>
```

## MANDATORY: Testing Protocol

### Required Tests (ALWAYS run):
1. **FCP Package Tests**: `cd fcp && go test` - MUST pass
2. **XML Validation**: `xmllint output.fcpxml --noout` - MUST pass  
3. **DTD Validation**: `xmllint --dtdvalid FCPXMLv1_13.dtd output.fcpxml`
4. **FCP Import Test**: Import into actual Final Cut Pro
5. **Playback Test**: Verify timeline plays correctly

### Validation Functions:
```go
func ValidateClaudeCompliance(fcpxml *FCPXML) error {
    if err := validateUniqueIDs(fcpxml); err != nil {
        return err
    }
    if err := validateElementTypes(fcpxml); err != nil {
        return err
    }
    if err := validateFrameAlignment(fcpxml); err != nil {
        return err
    }
    return nil
}
```

## 🏗️ Required Architecture Pattern

**ALWAYS follow this proven pattern:**

```go
func GenerateComplexTimeline() error {
    // 1. Create base FCPXML
    fcpxml, err := fcp.GenerateEmpty("")
    if err != nil {
        return err
    }
    
    // 2. Set up resource management
    registry := fcp.NewResourceRegistry(fcpxml)
    tx := fcp.NewTransaction(registry)
    defer tx.Rollback()
    
    // 3. Add assets with proper types
    for _, asset := range assets {
        switch detectMediaType(asset.Path) {
        case MediaTypeImage:
            if err := addImageToTimeline(fcpxml, tx, asset); err != nil {
                return err
            }
        case MediaTypeVideo:
            if err := addVideoToTimeline(fcpxml, tx, asset); err != nil {
                return err
            }
        }
    }
    
    // 4. Add text elements
    if err := addTitleElements(fcpxml, tx); err != nil {
        return err
    }
    
    // 5. Apply effects and animations
    if err := addEffectsAndAnimations(fcpxml, tx); err != nil {
        return err
    }
    
    // 6. Validate before commit
    if err := ValidateClaudeCompliance(fcpxml); err != nil {
        return fmt.Errorf("validation failed: %v", err)
    }
    
    // 7. Commit and write
    if err := tx.Commit(); err != nil {
        return err
    }
    
    return fcp.WriteToFile(fcpxml, outputFile)
}
```

## Advanced FCPXML Patterns

### Multi-Lane Timeline Construction:
```go
func buildMultiLaneTimeline(spine *fcp.Spine) error {
    // Main background video (lane 0)
    backgroundClip := fcp.AssetClip{
        Ref: "r2", Offset: "0s", Duration: "600600/24000s",
    }
    spine.AssetClips = append(spine.AssetClips, backgroundClip)
    
    // Overlay images (lanes 1-3)
    for i, imagePath := range overlayImages {
        lane := i + 1
        offset := fcp.ConvertSecondsToFCPDuration(float64(i) * 2.0)
        
        video := fcp.Video{
            Ref: fmt.Sprintf("r%d", 10+i),
            Lane: fmt.Sprintf("%d", lane),
            Offset: offset,
            Duration: "120120/24000s",
        }
        
        // Add to main background clip as nested element
        backgroundClip.Videos = append(backgroundClip.Videos, video)
    }
    
    return nil
}
```

### Animation Keyframe System:
```go
func createComplexAnimation(startTime, endTime, duration string) *fcp.AdjustTransform {
    return &fcp.AdjustTransform{
        Params: []fcp.Param{
            {
                Name: "position",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: startTime, Value: "-200 0"},
                        {Time: middleTime, Value: "0 0"},
                        {Time: endTime, Value: "200 0"},
                    },
                },
            },
            {
                Name: "scale",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: startTime, Value: "0.5 0.5", Curve: "smooth"},
                        {Time: endTime, Value: "1.5 1.5", Curve: "linear"},
                    },
                },
            },
        },
    }
}
```

**Key Principle: If FCPXML generation requires iteration to work, the architecture is wrong. Use proven patterns from working tests.**