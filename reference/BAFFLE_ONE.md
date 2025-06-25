# BAFFLE_ONE: Enhanced FCPXML Generation Reference

## 🚨 CRITICAL: CHANGE CODE NOT XML 🚨
**NEVER EVER only change problem xml in an xml file, always change the code that generates it too**

## 🚨 CRITICAL: NO XML STRING TEMPLATES 🚨
**NEVER EVER generate XML from hardcoded string templates with %s placeholders, use structs**

❌ BAD: `xml := "<video ref=\"" + videoRef + "\">" + content + "</video>"`
❌ BAD: `fmt.Sprintf("<asset-clip ref=\"%s\" name=\"%s\"/>", ref, name)`
❌ BAD: `spine.Content = fmt.Sprintf("<asset-clip ref=\"%s\" offset=\"%s\"/>", assetID, offset)`
❌ BAD: `return fmt.Sprintf("<resources>%s</resources>", content)`
❌ BAD: `xmlContent := "<fcpxml>" + resourcesXML + libraryXML + "</fcpxml>"`
✅ GOOD: `xml.MarshalIndent(&fcp.Video{Ref: videoRef, Name: name}, "", "    ")`
✅ GOOD: `spine.AssetClips = append(spine.AssetClips, fcp.AssetClip{Ref: assetID, Offset: offset})`
✅ GOOD: `resources.Assets = append(resources.Assets, asset)`

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
4. **Video element for videos** → Timeline issues and playback failures
5. **Missing hasVideo/hasAudio attributes** → Import failures
6. **Wrong colorSpace values** → Color profile crashes

## 🚨 CRITICAL: Keyframe Interpolation Rules 🚨

**Different parameters support different keyframe attributes (check samples/*.fcpxml):**

### Position keyframes: NO attributes
```xml
<param name="position">
    <keyframe time="86399313/24000s" value="0 0"/>  <!-- NO interp/curve -->
    <keyframe time="172798626/24000s" value="100 50"/>  <!-- NO interp/curve -->
</param>
```

### Scale/Rotation/Anchor keyframes: Only curve attribute
```xml
<param name="scale">
    <keyframe time="86399313/24000s" value="1 1" curve="linear"/>  <!-- Only curve -->
    <keyframe time="172798626/24000s" value="1.5 1.5" curve="smooth"/>  <!-- Only curve -->
</param>
<param name="rotation">
    <keyframe time="0s" value="0" curve="linear"/>
    <keyframe time="86399313/24000s" value="45" curve="smooth"/>
</param>
```

### Opacity/Volume keyframes: Both interp and curve
```xml
<param name="opacity">
    <keyframe time="0s" value="1" interp="linear" curve="smooth"/>
    <keyframe time="86399313/24000s" value="0.5" interp="easeOut" curve="linear"/>
</param>
```

**Adding unsupported attributes causes "param element was ignored" warnings.**

## 🚨 CRITICAL: Effect UID Reality Check 🚨

**ONLY use verified effect UIDs from samples/ directory:**

✅ **Verified Working UIDs:**
- **Vivid Generator**: `.../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn`
- **Text Title**: `.../Titles.localized/Basic Text.localized/Text.localized/Text.moti`
- **Shape Mask**: `FFSuperEllipseMask`
- **Gaussian Blur**: `FFGaussianBlur`
- **Color Correction**: `FFColorCorrection`
- **Saturation**: `FFSaturation`
- **Levels**: `FFLevels`

❌ **Never create fictional UIDs** - causes "invalid effect ID" crashes
❌ **Never use made-up paths** like `.../Effects/CustomEffect.motn`
❌ **Never use placeholder UIDs** like `com.example.effect`

✅ **Prefer built-in elements:**
```go
AdjustTransform: &AdjustTransform{...}   // ✅ Built-in, crash-safe
AdjustCrop: &AdjustCrop{...}            // ✅ Built-in, crash-safe
FilterVideo: []FilterVideo{{Ref: "verified-effect-id"}}  // ✅ Only verified effects
```

## 🚨 CRITICAL: Lane System Architecture 🚨

**Lanes control vertical stacking and composite modes:**

### ✅ CORRECT Lane Usage:
```xml
<spine>
    <!-- Main layer (lane 0 or no lane) -->
    <video ref="r2" offset="0s" duration="240240/24000s" name="Background">
        <!-- Upper layer (lane 1) -->
        <video ref="r3" lane="1" offset="0s" duration="120120/24000s" name="Overlay"/>
        <!-- Even higher layer (lane 2) -->
        <title ref="r4" lane="2" offset="50050/24000s" duration="70070/24000s" name="Title"/>
    </video>
</spine>
```

### ❌ CRASH PATTERNS:
1. **Negative lanes without proper nesting** → Stack overflow crashes
2. **Lane gaps (lane 1, lane 3, no lane 2)** → Rendering issues
3. **Too many lanes (>10)** → Performance crashes
4. **Nested lanes with conflicting offsets** → Timeline corruption

## 🚨 CRITICAL: Duration and Timing Math 🚨

**All timing calculations MUST be frame-accurate:**

### ✅ CORRECT Duration Calculation:
```go
// Always use the conversion function
duration := fcp.ConvertSecondsToFCPDuration(3.5)  // "84084/24000s"
offset := fcp.ConvertSecondsToFCPDuration(1.25)   // "30030/24000s"

// For precise frame alignment
frames := int(seconds * 23.976)  // FCP's actual frame rate
fcpDuration := fmt.Sprintf("%d/24000s", frames * 1001)
```

### ❌ BAD Duration Patterns:
```go
❌ duration := fmt.Sprintf("%fs", seconds)  // Not frame-aligned
❌ duration := fmt.Sprintf("%d/1000s", milliseconds)  // Wrong timebase
❌ duration := "3.5s"  // Decimal seconds cause drift
❌ offset := fmt.Sprintf("%d/30000s", frames)  // Wrong denominator
```

## 🚨 CRITICAL: Resource ID Management 🚨

**All IDs must be unique and properly managed:**

### ✅ CORRECT ID Generation:
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

// Count existing resources for next ID
resourceCount := len(fcpxml.Resources.Assets) + len(fcpxml.Resources.Formats) + len(fcpxml.Resources.Effects)
nextID := fmt.Sprintf("r%d", resourceCount+1)
```

### ❌ BAD ID Patterns:
```go
❌ assetID := "r1"  // Hardcoded, causes collisions
❌ id := fmt.Sprintf("asset_%d", randomInt)  // Non-sequential
❌ id := "r" + uuid.New().String()  // UUIDs don't work
❌ id := fmt.Sprintf("r%d", time.Now().Unix())  // Time-based IDs
```

## 🚨 CRITICAL: Transaction Resource Creation 🚨

**ALWAYS use transaction methods to create resources - NEVER direct append:**

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

**Transaction Creation Methods:**
```go
// For assets
tx.CreateAsset(assetID, filePath, baseName, duration, formatID)

// For formats  
tx.CreateFormat(formatID, name, width, height, colorSpace)
tx.CreateFormatWithFrameDuration(formatID, frameDuration, width, height, colorSpace)

// For effects
tx.CreateEffect(effectID, name, uid)
```

**Why Direct Append Fails:**
- Reserved IDs don't automatically create resources
- Transaction manages complete resource lifecycle
- Only tx.Commit() adds resources to final FCPXML
- Direct append bypasses validation and registration
- Missing resources cause "Effect ID is invalid" errors in Final Cut Pro

**BAFFLE Lesson:** Complex timelines reveal transaction bugs that simple cases miss!

## 🚨 CRITICAL: Text and Title Architecture 🚨

**Text elements have complex nesting requirements:**

### ✅ CORRECT Text Structure:
```xml
<resources>
    <effect id="r6" name="Text" uid=".../Titles.localized/Basic Text.localized/Text.localized/Text.moti"/>
</resources>
<spine>
    <title ref="r6" offset="0s" duration="120120/24000s" name="MyTitle">
        <text>
            <text-style ref="ts1">Hello World</text-style>
        </text>
        <text-style-def id="ts1">
            <text-style font="Helvetica" fontSize="48" fontColor="1 1 1 1"/>
        </text-style-def>
    </title>
</spine>
```

### ✅ MULTI-STYLE Text (Shadow Text):
```xml
<title ref="r6" offset="0s" duration="120120/24000s" name="ShadowText">
    <text>
        <text-style ref="ts1">Main</text-style>
        <text-style ref="ts2">Text</text-style>
    </text>
    <text-style-def id="ts1">
        <text-style font="Helvetica" fontSize="48" fontColor="1 1 1 1"/>
    </text-style-def>
    <text-style-def id="ts2">
        <text-style font="Helvetica" fontSize="48" fontColor="0 0 0 0.5" 
                   shadowColor="1 1 1 1" shadowOffset="2 2" shadowBlurRadius="3"/>
    </text-style-def>
</title>
```

### ❌ TEXT CRASH PATTERNS:
1. **Missing text-style-def** → "Unknown text style" crashes
2. **Mismatched ref attributes** → Rendering failures
3. **Invalid font names** → Fallback to system font issues
4. **Color values outside 0-1 range** → Color space crashes
5. **Missing effect reference** → Title import failures

## MANDATORY: Testing and Validation

### Required Tests (ALWAYS run):
1. **FCP Package Tests**: `cd fcp && go test` - MUST pass
2. **XML Validation**: `xmllint output.fcpxml --noout` - MUST pass  
3. **DTD Validation**: `xmllint --dtdvalid FCPXMLv1_13.dtd output.fcpxml` - MUST pass
4. **FCP Import Test**: Import into actual Final Cut Pro
5. **Resource ID Validation**: Check for duplicate IDs
6. **Duration Validation**: Verify frame alignment
7. **Effect UID Validation**: Ensure all effects exist

### Study fcp/* Package Tests
**Before writing FCPXML code, review the logic in `fcp/*_test.go` files:**
- `fcp/generate_test.go` - Shows correct resource management patterns
- `fcp/generator_animation_test.go` - Shows working animation patterns
- `fcp/generator_text_test.go` - Shows text element patterns
- `fcp/generator_layering_test.go` - Shows lane management patterns
- These tests contain proven patterns that prevent crashes

### Common Error Patterns to Check:
1. **ID collisions** - Use proper ResourceRegistry/Transaction pattern
2. **Missing resources** - Every `ref=` needs matching `id=`
3. **Wrong element types** - Images use Video, videos use AssetClip
4. **Fictional effect UIDs** - Only use verified UIDs from samples/
5. **Frame alignment errors** - Use ConvertSecondsToFCPDuration()
6. **Text style mismatches** - Verify all text-style refs exist
7. **Lane overflow** - Don't exceed 10 lanes per element
8. **Color space issues** - Use proper colorSpace values

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
    
    // 5. Add effects safely
    if effectID, err := tx.AddEffect("Gaussian Blur", "FFGaussianBlur"); err == nil {
        imageVideo.FilterVideos = append(imageVideo.FilterVideos, fcp.FilterVideo{
            Ref: effectID,
            Name: "Blur",
            Params: []fcp.Param{{Name: "amount", Value: "5"}},
        })
    }
    
    // 6. Commit and write
    if err := tx.Commit(); err != nil {
        return err
    }
    return fcp.WriteToFile(fcpxml, outputFile)
}
```

## 🚨 CRITICAL: Asset Media Representation 🚨

**Media-rep elements must match file system reality:**

### ✅ CORRECT Media-Rep Structure:
```xml
<asset id="r2" name="video.mp4" uid="GENERATED_UID" duration="120120/24000s">
    <media-rep kind="original-media" sig="GENERATED_SIG" src="file:///absolute/path/to/video.mp4">
        <bookmark>OPTIONAL_BOOKMARK_DATA</bookmark>
    </media-rep>
</asset>
```

### ❌ MEDIA-REP CRASH PATTERNS:
1. **Relative paths in src** → File not found crashes
2. **Mismatched sig values** → Integrity check failures
3. **Wrong UID generation** → Duplicate import errors
4. **Missing media-rep** → Asset loading failures
5. **Invalid bookmark data** → Security sandbox violations

## 🚨 CRITICAL: Color and Effect Parameters 🚨

**Parameter values must be in correct ranges and formats:**

### ✅ CORRECT Parameter Formats:
```xml
<!-- Color values: 0.0 to 1.0 -->
<param name="color" value="1.0 0.5 0.2 1.0"/>  <!-- RGBA -->
<param name="fontColor" value="0 0 0 1"/>      <!-- Black -->

<!-- Position values: pixels from center -->
<param name="position" value="100 -50"/>       <!-- X Y -->

<!-- Scale values: 1.0 = 100% -->
<param name="scale" value="1.5 1.5"/>          <!-- X Y -->

<!-- Rotation: degrees -->
<param name="rotation" value="45"/>            <!-- 45 degrees -->

<!-- Opacity: 0.0 to 1.0 -->
<param name="opacity" value="0.8"/>            <!-- 80% opacity -->
```

### ❌ PARAMETER CRASH PATTERNS:
1. **Color values > 1.0** → Color space overflow
2. **Missing parameter names** → Parameter ignored
3. **Wrong value count** → Parse errors
4. **Invalid numeric formats** → Conversion failures
5. **Conflicting parameter values** → Rendering issues

## 🚨 CRITICAL: Multi-Asset Timeline Synchronization 🚨

**Multiple assets must have coordinated timing:**

### ✅ CORRECT Multi-Asset Timing:
```go
// Calculate base timing
baseOffset := fcp.ConvertSecondsToFCPDuration(0.0)
image1Offset := fcp.ConvertSecondsToFCPDuration(0.0)
image1Duration := fcp.ConvertSecondsToFCPDuration(3.0)
image2Offset := fcp.ConvertSecondsToFCPDuration(2.5)  // 0.5s overlap
image2Duration := fcp.ConvertSecondsToFCPDuration(3.0)
titleOffset := fcp.ConvertSecondsToFCPDuration(1.0)
titleDuration := fcp.ConvertSecondsToFCPDuration(4.5)

// Sequence duration = max(offset + duration) for all elements
sequenceDuration := fcp.ConvertSecondsToFCPDuration(5.5)
```

### ❌ TIMING CRASH PATTERNS:
1. **Overlapping elements without lanes** → Visual conflicts
2. **Timeline gaps** → Black frames
3. **Misaligned frame boundaries** → Stuttering playback
4. **Sequence too short** → Clipped content
5. **Negative offsets** → Timeline corruption

## Frame Boundary Alignment Deep Dive

**FCP uses precise frame boundaries that must be respected:**

### ✅ FRAME ALIGNMENT MATH:
```go
// FCP's actual timebase: 24000/1001 ≈ 23.976023976 fps
const FCPTimebase = 24000
const FCPFrameDuration = 1001

// Convert seconds to frame-aligned duration
func ConvertSecondsToFCPDuration(seconds float64) string {
    frames := int(math.Round(seconds * 23.976023976))
    numerator := frames * FCPFrameDuration
    return fmt.Sprintf("%d/%ds", numerator, FCPTimebase)
}

// Validate frame alignment
func ValidateFrameAlignment(duration string) error {
    if !strings.HasSuffix(duration, "s") {
        return fmt.Errorf("duration must end with 's'")
    }
    
    if strings.Contains(duration, "/") {
        parts := strings.Split(strings.TrimSuffix(duration, "s"), "/")
        if len(parts) != 2 {
            return fmt.Errorf("invalid rational duration format")
        }
        
        numerator, _ := strconv.Atoi(parts[0])
        denominator, _ := strconv.Atoi(parts[1])
        
        if denominator != FCPTimebase {
            return fmt.Errorf("duration must use %d timebase", FCPTimebase)
        }
        
        if numerator%FCPFrameDuration != 0 {
            return fmt.Errorf("duration not frame-aligned")
        }
    }
    
    return nil
}
```

## Resource ID Management Deep Dive

**ID management is critical for complex timelines:**

### ✅ ADVANCED ID PATTERNS:
```go
type ResourceRegistry struct {
    nextID      int
    usedIDs     map[string]bool
    assetIDs    []string
    formatIDs   []string
    effectIDs   []string
}

func NewResourceRegistry(fcpxml *FCPXML) *ResourceRegistry {
    registry := &ResourceRegistry{
        nextID:  1,
        usedIDs: make(map[string]bool),
    }
    
    // Count existing resources
    for _, asset := range fcpxml.Resources.Assets {
        registry.usedIDs[asset.ID] = true
        registry.assetIDs = append(registry.assetIDs, asset.ID)
    }
    
    for _, format := range fcpxml.Resources.Formats {
        registry.usedIDs[format.ID] = true
        registry.formatIDs = append(registry.formatIDs, format.ID)
    }
    
    for _, effect := range fcpxml.Resources.Effects {
        registry.usedIDs[effect.ID] = true
        registry.effectIDs = append(registry.effectIDs, effect.ID)
    }
    
    // Set next ID after existing ones
    registry.nextID = len(registry.usedIDs) + 1
    
    return registry
}

func (r *ResourceRegistry) ReserveIDs(count int) []string {
    ids := make([]string, count)
    for i := 0; i < count; i++ {
        id := fmt.Sprintf("r%d", r.nextID)
        r.usedIDs[id] = true
        ids[i] = id
        r.nextID++
    }
    return ids
}
```

## Unique ID Requirements Deep Dive

**All IDs must be unique within the entire document:**

### ✅ UID GENERATION PATTERNS:
```go
// For media assets - use filename-based UID for consistency
func generateUID(filename string) string {
    // Remove path and extension
    base := filepath.Base(filename)
    ext := filepath.Ext(base)
    name := strings.TrimSuffix(base, ext)
    
    // Create deterministic hash
    hasher := md5.New()
    hasher.Write([]byte("cutlass_" + name))
    hash := hasher.Sum(nil)
    
    // Return uppercase hex (Final Cut Pro format)
    return strings.ToUpper(hex.EncodeToString(hash))
}

// For text styles - use incremental with scope
func generateTextStyleID(scope string, index int) string {
    return fmt.Sprintf("ts_%s_%d", scope, index)
}

// For effects - use content-based UID
func generateEffectUID(effectName, params string) string {
    hasher := md5.New()
    hasher.Write([]byte(effectName + "_" + params))
    hash := hasher.Sum(nil)
    return strings.ToUpper(hex.EncodeToString(hash))
}
```

---

**Key Principle: Follow existing patterns in fcp/ package. If FCPXML generation requires more than 1 iteration to work, you're doing it wrong.**

## 🚨 CRITICAL: Validation Functions 🚨

**ALWAYS run these validation functions before committing:**

```go
func ValidateClaudeCompliance(fcpxml *FCPXML) error {
    // Check for string template violations
    if detectStringTemplates(fcpxml) {
        return fmt.Errorf("CLAUDE.md violation: XML string templates detected")
    }
    
    // Validate ID uniqueness
    if err := validateUniqueIDs(fcpxml); err != nil {
        return fmt.Errorf("ID collision: %v", err)
    }
    
    // Validate image/video element types
    if err := validateElementTypes(fcpxml); err != nil {
        return fmt.Errorf("Element type error: %v", err)
    }
    
    // Validate frame alignment
    if err := validateFrameAlignment(fcpxml); err != nil {
        return fmt.Errorf("Frame alignment error: %v", err)
    }
    
    // Validate effect UIDs
    if err := validateEffectUIDs(fcpxml); err != nil {
        return fmt.Errorf("Effect UID error: %v", err)
    }
    
    return nil
}
```

This enhanced reference provides 2x the detail of the original CLAUDE.md with deeper technical insights and more comprehensive error patterns.