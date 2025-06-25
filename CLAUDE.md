# Project Context for AI Assistance

## 🚨 CRITICAL: CHANGE CODE NOT XML 🚨
**NEVER EVER only change problem xml in an xml file, always change the code that generates it too**

## 🚨 CRITICAL: NO XML STRING TEMPLATES 🚨
**NEVER EVER generate XML from hardcoded string templates with %s placeholders, use structs**

❌ BAD: `xml := "<video ref=\"" + videoRef + "\">" + content + "</video>"`
❌ BAD: `fmt.Sprintf("<asset-clip ref=\"%s\" name=\"%s\"/>", ref, name)`
✅ GOOD: `xml.MarshalIndent(&fcp.Video{Ref: videoRef, Name: name}, "", "    ")`

**All FCPXML generation MUST use the fcp.* structs in the fcp package.**

## 🚨 CRITICAL: Images vs Videos Architecture 🚨

**The #1 cause of crashes: Using wrong element types for images vs videos**

### ✅ IMAGES (PNG/JPG files):
```xml
<!-- Asset: duration="0s" (timeless) -->
<asset id="r2" duration="0s" hasVideo="1" format="r3" videoSources="1"/>
<!-- Format: NO frameDuration (timeless) -->
<format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"/>
<!-- Spine: Video element (NOT AssetClip) -->
<video ref="r2" duration="240240/24000s">
    <adjust-transform><!-- Simple animations work --></adjust-transform>
</video>
```

### ✅ VIDEOS (MOV/MP4 files):
```xml
<!-- Asset: has duration, audio properties -->
<asset id="r2" duration="14122857/100000s" hasVideo="1" hasAudio="1" audioSources="1"/>
<!-- Format: has frameDuration -->
<format id="r3" frameDuration="13335/400000s" width="1920" height="1080"/>
<!-- Spine: AssetClip element -->
<asset-clip ref="r2" duration="373400/3000s">
    <adjust-transform><!-- Complex animations work --></adjust-transform>
    <filter-video><!-- Effects work --></filter-video>
</asset-clip>
```

### ❌ CRASH PATTERNS:
1. **AssetClip for images** → `addAssetClip:toObject:parentFormatID` crash
2. **frameDuration on image formats** → `performAudioPreflightCheckForObject` crash  
3. **Complex effects on images** → Various import crashes

## 🚨 CRITICAL: Keyframe Interpolation Rules 🚨

**Different parameters support different keyframe attributes (check samples/*.fcpxml):**

### Position keyframes: NO attributes
```xml
<param name="position">
    <keyframe time="86399313/24000s" value="0 0"/>  <!-- NO interp/curve -->
</param>
```

### Scale/Rotation/Anchor keyframes: Only curve attribute
```xml
<param name="scale">
    <keyframe time="86399313/24000s" value="1 1" curve="linear"/>  <!-- Only curve -->
</param>
```

**Adding unsupported attributes causes "param element was ignored" warnings.**

## 🚨 CRITICAL: Effect UID Reality Check 🚨

**ONLY use verified effect UIDs from samples/ directory:**

✅ **Verified Working UIDs:**
- **Vivid Generator**: `.../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn`
- **Text Title**: `.../Titles.localized/Basic Text.localized/Text.localized/Text.moti`
- **Shape Mask**: `FFSuperEllipseMask`

❌ **Never create fictional UIDs** - causes "invalid effect ID" crashes

✅ **Prefer built-in elements:**
```go
AdjustTransform: &AdjustTransform{...}   // ✅ Built-in, crash-safe
AdjustCrop: &AdjustCrop{...}            // ✅ Built-in, crash-safe
```

## MANDATORY: Testing and Validation

### Required Tests (ALWAYS run):
1. **FCP Package Tests**: `cd fcp && go test` - MUST pass
2. **XML Validation**: `xmllint output.fcpxml --noout` - MUST pass  
3. **FCP Import Test**: Import into actual Final Cut Pro

### Study fcp/* Package Tests
**Before writing FCPXML code, review the logic in `fcp/*_test.go` files:**
- `fcp/generate_test.go` - Shows correct resource management patterns
- `fcp/generator_*_test.go` - Shows working animation/effect patterns
- These tests contain proven patterns that prevent crashes

### Common Error Patterns to Check:
1. **ID collisions** - Use proper ResourceRegistry/Transaction pattern
2. **Missing resources** - Every `ref=` needs matching `id=`
3. **Wrong element types** - Images use Video, videos use AssetClip
4. **Fictional effect UIDs** - Only use verified UIDs from samples/

## 🏗️ Required Architecture Pattern

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

## Frame Boundary Alignment

**All durations MUST use `fcp.ConvertSecondsToFCPDuration()`:**
- FCP uses 24000/1001 ≈ 23.976 fps timebase
- Manual calculations cause "not on edit frame boundary" errors
- Always use the conversion function

## Resource ID Management

**NEVER manually generate IDs:**
```go
❌ BAD: assetID := fmt.Sprintf("r%d", count+1)  // Race conditions!
✅ GOOD: ids := tx.ReserveIDs(1); assetID := ids[0]  // Thread-safe
```

## 🚨 CRITICAL: Transaction Resource Creation 🚨

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

## Unique ID Requirements

**All IDs must be unique within document:**
- Text style IDs: Use `generateUID()` not hardcoded "ts1"
- Asset/format/effect IDs: Use ResourceRegistry
- Media UIDs: Use filename-based generation for consistency

---

**Key Principle: Follow existing patterns in fcp/ package. If FCPXML generation requires more than 1 iteration to work, you're doing it wrong.**