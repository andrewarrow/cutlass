# Cutlass FCPXML Generation Framework

Welcome to the comprehensive documentation for the Cutlass FCPXML generation framework. This wiki provides detailed guidance for working with Final Cut Pro XML files programmatically.

## 🚨 Critical First Reads

Before diving into development, **you must** read these essential documents:

- **[FCPXML Generation Best Practices](FCPXML-Generation-Best-Practices.md)** - Core rules and patterns
- **[Images vs Videos Architecture](Images-vs-Videos-Architecture.md)** - Media type handling (prevents #1 crash cause)
- **[Testing and Debugging](Testing-and-Debugging.md)** - Validation and troubleshooting

## 📚 Wiki Navigation

### Core Documentation
- **[Architecture Overview](Architecture-Overview.md)** - System design and patterns
- **[FCPXML Generation Best Practices](FCPXML-Generation-Best-Practices.md)** - Critical rules and patterns
- **[Images vs Videos Architecture](Images-vs-Videos-Architecture.md)** - Media type handling
- **[Animation and Effects](Animation-and-Effects.md)** - Keyframes, animations, and effects
- **[Testing and Debugging](Testing-and-Debugging.md)** - Validation and troubleshooting
- **[API Reference](API-Reference.md)** - Complete function and type reference

### Quick Reference
- [Resource ID Management](#resource-id-management)
- [Duration and Timing](#duration-and-timing)
- [Common Crash Patterns](#common-crash-patterns)
- [Effect UID Verification](#effect-uid-verification)

## 🛡️ Critical Architecture Rules

### 1. NO XML STRING TEMPLATES
**NEVER EVER generate XML from hardcoded string templates:**

```go
❌ BAD: xml := "<video ref=\"" + videoRef + "\">" + content + "</video>"
❌ BAD: fmt.Sprintf("<asset-clip ref=\"%s\" name=\"%s\"/>", ref, name)
✅ GOOD: xml.MarshalIndent(&fcp.Video{Ref: videoRef, Name: name}, "", "    ")
```

**All FCPXML generation MUST use the fcp.* structs in the fcp package.**

### 2. CHANGE CODE NOT XML
**NEVER EVER only change problem xml in an xml file, always change the code that generates it too**

### 3. Images vs Videos Fundamentals

| Media Type | Asset Duration | Format frameDuration | Spine Element | Effects Support |
|------------|----------------|---------------------|---------------|----------------|
| **Images** | `"0s"` (timeless) | **NONE** | `<video>` | Simple only |
| **Videos** | Actual duration | Required | `<asset-clip>` | Full support |

## 🚨 Top Crash Patterns to Avoid

1. **AssetClip for images** → `addAssetClip:toObject:parentFormatID` crash
2. **frameDuration on image formats** → `performAudioPreflightCheckForObject` crash  
3. **Complex effects on images** → Various import crashes
4. **Fictional effect UIDs** → "invalid effect ID" crashes
5. **Non-frame-aligned durations** → "not on edit frame boundary" errors

## Resource ID Management

All IDs must be unique within the document:

```go
// ✅ GOOD: Use ResourceRegistry pattern
registry := fcp.NewResourceRegistry(fcpxml)
tx := fcp.NewTransaction(registry)
defer tx.Rollback()

ids := tx.ReserveIDs(3)
assetID := ids[0]    // "r2"
formatID := ids[1]   // "r3"
effectID := ids[2]   // "r4"

// ❌ BAD: Hardcoded IDs cause collisions
assetID := "r1"  // Will conflict with other generators
```

## Duration and Timing

**All durations MUST use `fcp.ConvertSecondsToFCPDuration()`:**

```go
// ✅ GOOD: Frame-aligned duration
duration := fcp.ConvertSecondsToFCPDuration(5.5)  // "132132/24000s"

// ❌ BAD: Decimal seconds cause drift
duration := "5.5s"  // Not frame-aligned
```

FCP uses 24000/1001 ≈ 23.976 fps timebase for frame alignment.

## Effect UID Verification

**ONLY use verified effect UIDs:**

### ✅ Verified Working UIDs:
- **Gaussian Blur**: `FFGaussianBlur`
- **Color Correction**: `FFColorCorrection` 
- **Text Title**: `.../Titles.localized/Basic Text.localized/Text.localized/Text.moti`
- **Shape Mask**: `FFSuperEllipseMask`

### ✅ Prefer built-in elements:
```go
// Spatial transformations - always safe
video.AdjustTransform = &fcp.AdjustTransform{
    Position: "100 50",
    Scale:    "1.5 1.5",
}

// Cropping - always safe  
assetClip.AdjustCrop = &fcp.AdjustCrop{
    Mode: "trim",
    TrimRect: &fcp.TrimRect{Left: "0.1", Right: "0.9"},
}
```

## Required Architecture Pattern

**ALWAYS follow this pattern:**

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

## Package Structure

```
fcp/
├── types.go           # All FCPXML struct definitions
├── generator.go       # Core generation functions  
├── registry.go        # Resource ID management
├── transaction.go     # Transaction-based operations
├── ids.go            # ID generation utilities
└── *_test.go         # Comprehensive test patterns
```

## Getting Started

1. **Read** [FCPXML Generation Best Practices](FCPXML-Generation-Best-Practices.md)
2. **Study** the test files in `fcp/*_test.go` for proven patterns
3. **Validate** your code with `fcp.ValidateClaudeCompliance()`
4. **Test** imports in actual Final Cut Pro

---

**Key Principle: Follow existing patterns in fcp/ package. If FCPXML generation requires more than 1 iteration to work, you're doing it wrong.**