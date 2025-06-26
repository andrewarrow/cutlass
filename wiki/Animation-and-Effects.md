# Animation and Effects

This document covers keyframe animations, effects, and transformations in FCPXML, including the critical rules for parameter interpolation and effect UID verification.

## 🚨 CRITICAL: Keyframe Interpolation Rules

**Different parameters support different keyframe attributes - adding unsupported attributes causes "param element was ignored" warnings.**

## Keyframe Attribute Support Matrix

| Parameter Type | `interp` Attribute | `curve` Attribute | Example |
|---------------|-------------------|-------------------|---------|
| **position** | ❌ NOT supported | ❌ NOT supported | `<keyframe time="0s" value="0 0"/>` |
| **scale** | ❌ NOT supported | ✅ Supported | `<keyframe time="0s" value="1 1" curve="linear"/>` |
| **rotation** | ❌ NOT supported | ✅ Supported | `<keyframe time="0s" value="0" curve="smooth"/>` |
| **anchor** | ❌ NOT supported | ✅ Supported | `<keyframe time="0s" value="0.5 0.5" curve="linear"/>` |
| **opacity** | ✅ Supported | ✅ Supported | `<keyframe time="0s" value="1" interp="easeIn" curve="smooth"/>` |
| **volume** | ✅ Supported | ✅ Supported | `<keyframe time="0s" value="0dB" interp="linear" curve="smooth"/>` |

## Position Keyframes - Special Handling

**Position keyframes have NO attributes - spatial interpolation is handled by FCP's internal engine:**

```xml
<param name="position">
    <keyframe time="86399313/24000s" value="0 0"/>  <!-- NO interp/curve -->
    <keyframe time="172798626/24000s" value="100 50"/>  <!-- NO interp/curve -->
</param>
```

### Why Position Keyframes Are Different:
- Position interpolation is handled by Final Cut Pro's internal spatial engine
- Adding interp/curve attributes causes "param element was ignored" warnings
- The spatial engine uses Bezier curves automatically
- Linear interpolation is the default and cannot be overridden

### Position Implementation:
```go
func createPositionAnimation(startPos, endPos string, startTime, endTime string) fcp.Param {
    return fcp.Param{
        Name: "position",
        KeyframeAnimation: &fcp.KeyframeAnimation{
            Keyframes: []fcp.Keyframe{
                {Time: startTime, Value: startPos},     // No attributes
                {Time: endTime, Value: endPos},         // No attributes
            },
        },
    }
}
```

## Scale/Rotation/Anchor Keyframes - Curve Only

**These parameters support ONLY the curve attribute:**

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

### Curve Attribute Values:
- `linear`: Straight line interpolation between keyframes
- `smooth`: Bezier curve interpolation with automatic tangent handling
- `hold`: No interpolation, jump to next value
- Default: `smooth` if not specified

### Scale/Rotation Implementation:
```go
func createScaleAnimation(startScale, endScale string, startTime, endTime string) fcp.Param {
    return fcp.Param{
        Name: "scale",
        KeyframeAnimation: &fcp.KeyframeAnimation{
            Keyframes: []fcp.Keyframe{
                {Time: startTime, Value: startScale, Curve: "linear"},
                {Time: endTime, Value: endScale, Curve: "smooth"},
            },
        },
    }
}

func createRotationAnimation(startAngle, endAngle float64, startTime, endTime string) fcp.Param {
    return fcp.Param{
        Name: "rotation",
        KeyframeAnimation: &fcp.KeyframeAnimation{
            Keyframes: []fcp.Keyframe{
                {Time: startTime, Value: fmt.Sprintf("%f", startAngle), Curve: "linear"},
                {Time: endTime, Value: fmt.Sprintf("%f", endAngle), Curve: "smooth"},
            },
        },
    }
}
```

## Opacity/Volume Keyframes - Full Support

**These parameters support both interp and curve attributes:**

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

### Interp Attribute Values:
- `linear`: Constant rate of change
- `easeIn`: Slow start, fast finish
- `easeOut`: Fast start, slow finish
- `easeInOut`: Slow start and finish, fast middle
- Default: `linear` if not specified

### Opacity/Volume Implementation:
```go
func createOpacityAnimation(startOpacity, endOpacity float64, startTime, endTime string) fcp.Param {
    return fcp.Param{
        Name: "opacity",
        KeyframeAnimation: &fcp.KeyframeAnimation{
            Keyframes: []fcp.Keyframe{
                {
                    Time:   startTime,
                    Value:  fmt.Sprintf("%f", startOpacity),
                    Interp: "linear",
                    Curve:  "smooth",
                },
                {
                    Time:   endTime,
                    Value:  fmt.Sprintf("%f", endOpacity),
                    Interp: "easeOut",
                    Curve:  "linear",
                },
            },
        },
    }
}
```

## Color Keyframes - Special Handling

**Color parameters require special attention for interpolation:**

```xml
<param name="color">
    <keyframe time="0s" value="1 0 0 1" interp="linear" curve="smooth"/>
    <keyframe time="86399313/24000s" value="0 1 0 1" interp="linear" curve="smooth"/>
</param>
```

### Color Interpolation Rules:
- Colors interpolate in RGB space, not HSV
- Alpha channel interpolates separately
- Color values must be 0.0 to 1.0 range
- Invalid color values cause rendering failures

```go
func createColorAnimation(startColor, endColor string, startTime, endTime string) fcp.Param {
    return fcp.Param{
        Name: "color",
        KeyframeAnimation: &fcp.KeyframeAnimation{
            Keyframes: []fcp.Keyframe{
                {
                    Time:   startTime,
                    Value:  startColor, // "1 0 0 1" (RGBA)
                    Interp: "linear",
                    Curve:  "smooth",
                },
                {
                    Time:   endTime,
                    Value:  endColor,   // "0 1 0 1" (RGBA)
                    Interp: "linear",
                    Curve:  "smooth",
                },
            },
        },
    }
}
```

## Keyframe Validation System

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

## 🚨 CRITICAL: Effect UID Reality Check

**ONLY use verified effect UIDs from samples/ directory:**

## Verified Working UIDs - Complete Catalog

### Built-in Video Effects:
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

### Built-in Audio Effects:
- **Gain**: `FFAudioGain`
- **EQ**: `FFAudioEQ`
- **Compressor**: `FFAudioCompressor`
- **Limiter**: `FFAudioLimiter`
- **Gate**: `FFAudioGate`
- **DeEsser**: `FFAudioDeEsser`

### Motion Templates (Verified Paths):
- **Vivid Generator**: `.../Generators.localized/Solids.localized/Vivid.localized/Vivid.motn`
- **Text Title**: `.../Titles.localized/Basic Text.localized/Text.localized/Text.moti`
- **Lower Third**: `.../Titles.localized/Basic Text.localized/Lower Third.localized/Lower Third.moti`
- **Animated Title**: `.../Titles.localized/Animated.localized/Animated.localized/Animated.moti`

### Shape/Mask Effects:
- **Shape Mask**: `FFSuperEllipseMask`
- **Rectangle Mask**: `FFRectangleMask`
- **Circle Mask**: `FFCircleMask`
- **Polygon Mask**: `FFPolygonMask`

## ❌ Never Create Fictional UIDs

**Creating fictional UIDs causes "invalid effect ID" crashes:**

```go
❌ BAD: uid := "com.example.customeffect"
❌ BAD: uid := ".../Effects/MyCustomEffect.motn"
❌ BAD: uid := "user.defined.blur"
❌ BAD: uid := "/Library/Effects/CustomBlur.plugin"
❌ BAD: uid := "CustomEffect_" + generateUID()
```

## ✅ Prefer Built-in Elements

**Built-in elements are always safe and crash-free:**

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

## Effect UID Verification System

```go
var verifiedEffectUIDs = map[string]bool{
    "FFGaussianBlur":    true,
    "FFMotionBlur":      true,
    "FFColorCorrection": true,
    "FFSaturation":      true,
    "FFLevels":          true,
    "FFCurves":          true,
    "FFHueSaturation":   true,
    "FFBrightness":      true,
    "FFContrast":        true,
    "FFGamma":           true,
    "FFSharpen":         true,
    "FFUnsharpMask":     true,
    "FFAudioGain":       true,
    "FFAudioEQ":         true,
    "FFAudioCompressor": true,
    "FFAudioLimiter":    true,
    "FFAudioGate":       true,
    "FFAudioDeEsser":    true,
    "FFSuperEllipseMask": true,
    "FFRectangleMask":   true,
    "FFCircleMask":      true,
    "FFPolygonMask":     true,
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

## Transform Animations

### Basic Transform Animation

```go
func createBasicTransformAnimation(duration string) *fcp.AdjustTransform {
    startTime := "0s"
    endTime := duration
    
    return &fcp.AdjustTransform{
        Params: []fcp.Param{
            // Position animation (no attributes)
            {
                Name: "position",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: startTime, Value: "0 0"},
                        {Time: endTime, Value: "200 100"},
                    },
                },
            },
            // Scale animation (curve only)
            {
                Name: "scale",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: startTime, Value: "1 1", Curve: "linear"},
                        {Time: endTime, Value: "1.5 1.5", Curve: "smooth"},
                    },
                },
            },
            // Rotation animation (curve only)
            {
                Name: "rotation",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: startTime, Value: "0", Curve: "linear"},
                        {Time: endTime, Value: "45", Curve: "smooth"},
                    },
                },
            },
        },
    }
}
```

### Complex Multi-Parameter Animation

```go
func createComplexAnimation(duration string) []fcp.Param {
    midTime := fcp.ConvertSecondsToFCPDuration(
        fcp.ParseFCPDuration(duration) / 2.0,
    )
    
    return []fcp.Param{
        // Three-point position animation
        {
            Name: "position",
            KeyframeAnimation: &fcp.KeyframeAnimation{
                Keyframes: []fcp.Keyframe{
                    {Time: "0s", Value: "-100 -50"},
                    {Time: midTime, Value: "0 0"},
                    {Time: duration, Value: "100 50"},
                },
            },
        },
        // Scale with easing
        {
            Name: "scale",
            KeyframeAnimation: &fcp.KeyframeAnimation{
                Keyframes: []fcp.Keyframe{
                    {Time: "0s", Value: "0.5 0.5", Curve: "smooth"},
                    {Time: midTime, Value: "1.2 1.2", Curve: "smooth"},
                    {Time: duration, Value: "1 1", Curve: "linear"},
                },
            },
        },
        // Opacity fade
        {
            Name: "opacity",
            KeyframeAnimation: &fcp.KeyframeAnimation{
                Keyframes: []fcp.Keyframe{
                    {Time: "0s", Value: "0", Interp: "easeIn", Curve: "smooth"},
                    {Time: midTime, Value: "1", Interp: "linear", Curve: "smooth"},
                    {Time: duration, Value: "0.8", Interp: "easeOut", Curve: "linear"},
                },
            },
        },
    }
}
```

## Filter Effects

### Video Filter Implementation

```go
func createVideoFilter(effectUID, name string, params map[string]string) fcp.FilterVideo {
    filter := fcp.FilterVideo{
        Ref:  effectUID,
        Name: name,
    }
    
    for paramName, value := range params {
        filter.Params = append(filter.Params, fcp.Param{
            Name:  paramName,
            Value: value,
        })
    }
    
    return filter
}

// Example usage:
blurFilter := createVideoFilter("FFGaussianBlur", "Blur", map[string]string{
    "amount": "5.0",
})

colorFilter := createVideoFilter("FFColorCorrection", "Color", map[string]string{
    "saturation": "1.2",
    "exposure":   "0.5",
    "shadows":    "0.1",
    "highlights": "-0.1",
})
```

### Audio Filter Implementation

```go
func createAudioFilter(effectUID, name string, params map[string]string) fcp.FilterAudio {
    filter := fcp.FilterAudio{
        Ref:  effectUID,
        Name: name,
    }
    
    for paramName, value := range params {
        filter.Params = append(filter.Params, fcp.Param{
            Name:  paramName,
            Value: value,
        })
    }
    
    return filter
}

// Example usage:
eqFilter := createAudioFilter("FFAudioEQ", "EQ", map[string]string{
    "low-freq":  "100",
    "low-gain":  "2",
    "mid-freq":  "1000",
    "mid-gain":  "0",
    "high-freq": "10000",
    "high-gain": "-1",
})
```

## Animation Timing Helpers

```go
func calculateKeyframeTimes(totalDuration string, keyframeCount int) []string {
    total := fcp.ParseFCPDuration(totalDuration)
    times := make([]string, keyframeCount)
    
    for i := 0; i < keyframeCount; i++ {
        progress := float64(i) / float64(keyframeCount-1)
        time := total * progress
        times[i] = fcp.ConvertSecondsToFCPDuration(time)
    }
    
    return times
}

func createEaseInOutCurve(startValue, endValue float64, times []string) []fcp.Keyframe {
    keyframes := make([]fcp.Keyframe, len(times))
    
    for i, time := range times {
        progress := float64(i) / float64(len(times)-1)
        
        // Ease-in-out curve calculation
        easedProgress := progress
        if progress < 0.5 {
            easedProgress = 2 * progress * progress
        } else {
            easedProgress = 1 - 2*(1-progress)*(1-progress)
        }
        
        value := startValue + (endValue-startValue)*easedProgress
        
        keyframes[i] = fcp.Keyframe{
            Time:   time,
            Value:  fmt.Sprintf("%f", value),
            Interp: "smooth",
            Curve:  "smooth",
        }
    }
    
    return keyframes
}
```

## Media Type Specific Limitations

### Images - Limited Animation Support

Images support only simple transforms and basic effects:

```go
func createImageAnimation(imageVideo *fcp.Video, duration string) {
    // ✅ SAFE: Simple position and scale
    imageVideo.AdjustTransform = &fcp.AdjustTransform{
        Params: []fcp.Param{
            {
                Name: "position",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: "0s", Value: "0 0"},
                        {Time: duration, Value: "100 50"},
                    },
                },
            },
            {
                Name: "scale",
                KeyframeAnimation: &fcp.KeyframeAnimation{
                    Keyframes: []fcp.Keyframe{
                        {Time: "0s", Value: "1 1", Curve: "linear"},
                        {Time: duration, Value: "1.2 1.2", Curve: "smooth"},
                    },
                },
            },
        },
    }
    
    // ❌ AVOID: Complex effects that require temporal processing
    // imageVideo.FilterVideos = []fcp.FilterVideo{{Ref: "FFMotionBlur"}}
}
```

### Videos - Full Animation Support

Videos support complete animation and effect capabilities:

```go
func createVideoAnimation(assetClip *fcp.AssetClip, duration string) {
    // ✅ FULL SUPPORT: All animations work
    assetClip.AdjustTransform = createComplexAnimation(duration)
    
    // ✅ FULL SUPPORT: All effects work
    assetClip.FilterVideos = []fcp.FilterVideo{
        createVideoFilter("FFGaussianBlur", "Blur", map[string]string{"amount": "3"}),
        createVideoFilter("FFColorCorrection", "Color", map[string]string{"saturation": "1.2"}),
    }
}
```

This comprehensive guide ensures proper animation and effect implementation while avoiding the common pitfalls that cause FCP import failures.