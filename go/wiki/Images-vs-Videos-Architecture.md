# Images vs Videos Architecture

**The #1 cause of crashes: Using wrong element types for images vs videos**

This document explains the fundamental architectural differences between handling images and videos in FCPXML, which is the most critical aspect of successful FCPXML generation.

## 🚨 CRITICAL ARCHITECTURE DISTINCTION

Final Cut Pro treats images and videos as fundamentally different media types with incompatible XML structures. Using the wrong structure causes immediate crashes.

## Media Type Comparison

| Aspect | Images (PNG/JPG/GIF) | Videos (MP4/MOV/AVI) |
|--------|---------------------|---------------------|
| **Asset Duration** | `"0s"` (timeless) | Actual file duration |
| **Format frameDuration** | **OMITTED** | **REQUIRED** |
| **Timeline Element** | `<video>` wrapper | `<asset-clip>` wrapper |
| **Effects Support** | Simple transforms only | Full effect support |
| **Animation Support** | Limited keyframes | Complete keyframe support |
| **Audio Properties** | Never present | Often present |

## Images (PNG/JPG/GIF) - Complete Specification

### ✅ CORRECT Image Structure:

```xml
<resources>
    <!-- Image Asset: duration="0s" (timeless) - CRITICAL -->
    <asset id="r2" name="image.png" uid="GENERATED_UID" start="0s" duration="0s" 
           hasVideo="1" format="r3" videoSources="1">
        <media-rep kind="original-media" sig="GENERATED_SIG" src="file:///absolute/path/to/image.png"/>
    </asset>

    <!-- Image Format: NO frameDuration (timeless) - CRITICAL -->
    <format id="r3" name="FFVideoFormatRateUndefined" width="1280" height="720" colorSpace="1-13-1"/>
</resources>

<spine>
    <!-- Image Timeline: Video element (NOT AssetClip) - CRITICAL -->
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
</spine>
```

### Image Properties Implementation:

```go
func createImageAsset(imagePath, assetID, formatID string) (*fcp.Asset, *fcp.Format, error) {
    // Images are timeless - always duration="0s"
    asset := &fcp.Asset{
        ID:           assetID,
        Name:         filepath.Base(imagePath),
        UID:          generateUID(filepath.Base(imagePath)),
        Start:        "0s",
        Duration:     "0s",              // CRITICAL: Images have no intrinsic duration
        HasVideo:     "1",               // Images are visual
        VideoSources: "1",
        Format:       formatID,
        MediaRep: fcp.MediaRep{
            Kind: "original-media",
            Src:  "file://" + absPath,
            Sig:  generateSig(imagePath),
        },
    }
    
    // Detect image dimensions
    width, height, err := getImageDimensions(imagePath)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to get image dimensions: %v", err)
    }
    
    format := &fcp.Format{
        ID:         formatID,
        Name:       "FFVideoFormatRateUndefined",
        Width:      strconv.Itoa(width),
        Height:     strconv.Itoa(height),
        ColorSpace: "1-13-1",
        // CRITICAL: NO frameDuration for images
    }
    
    return asset, format, nil
}
```

## Videos (MP4/MOV/AVI) - Complete Specification

### ✅ CORRECT Video Structure:

```xml
<resources>
    <!-- Video Asset: has duration, audio properties - CRITICAL -->
    <asset id="r4" name="video.mp4" uid="GENERATED_UID" start="0s" duration="14122857/100000s" 
           hasVideo="1" hasAudio="1" audioSources="1" audioChannels="2" audioRate="48000">
        <media-rep kind="original-media" sig="GENERATED_SIG" src="file:///absolute/path/to/video.mp4"/>
    </asset>

    <!-- Video Format: has frameDuration - CRITICAL -->
    <format id="r5" name="FFVideoFormat1080p30" frameDuration="1001/30000s" width="1920" height="1080" 
            colorSpace="1-1-1 (Rec. 709)"/>
</resources>

<spine>
    <!-- Video Timeline: AssetClip element - CRITICAL -->
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
</spine>
```

### Video Properties Implementation:

```go
func createVideoAsset(videoPath, assetID, formatID string) (*fcp.Asset, *fcp.Format, error) {
    // Detect video properties
    info, err := getVideoInfo(videoPath)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to get video info: %v", err)
    }
    
    asset := &fcp.Asset{
        ID:       assetID,
        Name:     filepath.Base(videoPath),
        UID:      generateUID(filepath.Base(videoPath)),
        Start:    "0s",
        Duration: fcp.ConvertSecondsToFCPDuration(info.Duration), // Actual duration
        HasVideo: "1",
        VideoSources: "1",
        Format:   formatID,
        MediaRep: fcp.MediaRep{
            Kind: "original-media",
            Src:  "file://" + absPath,
            Sig:  generateSig(videoPath),
        },
    }
    
    // Add audio properties if audio tracks present
    if info.HasAudio {
        asset.HasAudio = "1"
        asset.AudioSources = "1"
        asset.AudioChannels = strconv.Itoa(info.AudioChannels)
        asset.AudioRate = strconv.Itoa(info.AudioRate)
    }
    
    format := &fcp.Format{
        ID:            formatID,
        Name:          generateFormatName(info),
        FrameDuration: fcp.ConvertFrameRateToFCPDuration(info.FrameRate), // CRITICAL: Required for videos
        Width:         strconv.Itoa(info.Width),
        Height:        strconv.Itoa(info.Height),
        ColorSpace:    info.ColorSpace,
    }
    
    return asset, format, nil
}
```

## ❌ CRASH PATTERNS - Detailed Analysis

### 1. AssetClip for Images → `addAssetClip:toObject:parentFormatID` crash

**Root Cause**: FCP expects Video elements for timeless media  
**Symptoms**: Immediate crash on import  
**Fix**: Use Video element with duration parameter

```go
❌ BAD: spine.AssetClips = append(spine.AssetClips, fcp.AssetClip{Ref: imageAssetID})
✅ GOOD: spine.Videos = append(spine.Videos, fcp.Video{Ref: imageAssetID, Duration: timelineDuration})
```

### 2. frameDuration on Image Formats → `performAudioPreflightCheckForObject` crash

**Root Cause**: Images don't have frame rates  
**Symptoms**: Crash during audio analysis  
**Fix**: Omit frameDuration from image formats

```go
❌ BAD: format := fcp.Format{FrameDuration: "1001/24000s", Width: "1280", Height: "720"}
✅ GOOD: format := fcp.Format{Width: "1280", Height: "720"} // No frameDuration
```

### 3. Complex Effects on Images → Various import crashes

**Root Cause**: Images can't handle time-based effects  
**Symptoms**: Effect rendering failures  
**Fix**: Use only spatial effects on images

```go
❌ BAD: imageVideo.FilterVideos = []fcp.FilterVideo{{Ref: "motion-blur", Name: "Motion Blur"}}
✅ GOOD: imageVideo.AdjustTransform = &fcp.AdjustTransform{Scale: "1.5 1.5"}
```

### 4. Video Element for Videos → Timeline issues and playback failures

**Root Cause**: Videos need proper timeline integration  
**Symptoms**: Broken playback, audio sync issues  
**Fix**: Use AssetClip elements for video media

```go
❌ BAD: spine.Videos = append(spine.Videos, fcp.Video{Ref: videoAssetID})
✅ GOOD: spine.AssetClips = append(spine.AssetClips, fcp.AssetClip{Ref: videoAssetID})
```

### 5. Missing hasVideo/hasAudio Attributes → Import failures

**Root Cause**: FCP needs to know media capabilities  
**Symptoms**: Media not recognized  
**Fix**: Always specify media capabilities

```go
❌ BAD: asset := fcp.Asset{ID: id, Duration: duration}
✅ GOOD: asset := fcp.Asset{ID: id, Duration: duration, HasVideo: "1", HasAudio: "1"}
```

### 6. Wrong ColorSpace Values → Color profile crashes

**Root Cause**: Invalid color space specifications  
**Symptoms**: Color rendering failures  
**Fix**: Use standard color space values

```go
❌ BAD: format.ColorSpace = "RGB"
✅ GOOD: format.ColorSpace = "1-1-1 (Rec. 709)"
```

## Media Type Detection Logic

```go
type MediaType int

const (
    MediaTypeImage MediaType = iota
    MediaTypeVideo
    MediaTypeAudio
    MediaTypeUnknown
)

func detectMediaType(filePath string) (MediaType, error) {
    ext := strings.ToLower(filepath.Ext(filePath))
    
    switch ext {
    case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff", ".webp":
        return MediaTypeImage, nil
    case ".mp4", ".mov", ".avi", ".mkv", ".m4v", ".webm":
        return MediaTypeVideo, nil
    case ".mp3", ".wav", ".aac", ".m4a", ".flac", ".ogg":
        return MediaTypeAudio, nil
    default:
        return MediaTypeUnknown, fmt.Errorf("unsupported media type: %s", ext)
    }
}

func createAssetForMediaType(mediaType MediaType, filePath string, assetID, formatID string) (*fcp.Asset, *fcp.Format, error) {
    switch mediaType {
    case MediaTypeImage:
        return createImageAsset(filePath, assetID, formatID)
    case MediaTypeVideo:
        return createVideoAsset(filePath, assetID, formatID)
    case MediaTypeAudio:
        return createAudioAsset(filePath, assetID, formatID)
    default:
        return nil, nil, fmt.Errorf("unsupported media type")
    }
}
```

## Timeline Element Selection

```go
func addMediaToSpine(spine *fcp.Spine, asset *fcp.Asset, offset, duration string) error {
    mediaType, err := detectMediaType(asset.MediaRep.Src)
    if err != nil {
        return err
    }
    
    switch mediaType {
    case MediaTypeImage:
        // Images use Video elements
        video := fcp.Video{
            Ref:      asset.ID,
            Offset:   offset,
            Duration: duration,
            Name:     asset.Name,
        }
        spine.Videos = append(spine.Videos, video)
        
    case MediaTypeVideo, MediaTypeAudio:
        // Videos and audio use AssetClip elements
        clip := fcp.AssetClip{
            Ref:      asset.ID,
            Offset:   offset,
            Duration: duration,
            Name:     asset.Name,
        }
        spine.AssetClips = append(spine.AssetClips, clip)
        
    default:
        return fmt.Errorf("unsupported media type for timeline")
    }
    
    return nil
}
```

## Advanced Media Handling

### Portrait vs Landscape Detection

```go
func detectVideoOrientation(width, height int) string {
    if height > width {
        return "portrait"
    } else if width > height {
        return "landscape"
    }
    return "square"
}

func createFormatForVideo(info VideoInfo, formatID string) *fcp.Format {
    format := &fcp.Format{
        ID:            formatID,
        FrameDuration: fcp.ConvertFrameRateToFCPDuration(info.FrameRate),
        Width:         strconv.Itoa(info.Width),
        Height:        strconv.Itoa(info.Height),
    }
    
    // Set appropriate format name based on resolution and orientation
    orientation := detectVideoOrientation(info.Width, info.Height)
    if orientation == "portrait" {
        format.Name = fmt.Sprintf("FFVideoFormat%dx%d", info.Height, info.Width)
    } else {
        format.Name = fmt.Sprintf("FFVideoFormat%dx%d", info.Width, info.Height)
    }
    
    return format
}
```

### Frame Rate Validation

```go
var standardFrameRates = map[float64]string{
    23.976: "1001/24000s",
    24.0:   "1/24s",
    25.0:   "1/25s",
    29.97:  "1001/30000s",
    30.0:   "1/30s",
    50.0:   "1/50s",
    59.94:  "1001/60000s",
    60.0:   "1/60s",
}

func validateAndMapFrameRate(frameRate float64) (string, error) {
    // Reject obviously bogus frame rates
    if frameRate > 120 || frameRate < 1 {
        return "", fmt.Errorf("invalid frame rate: %f", frameRate)
    }
    
    // Find closest standard frame rate
    var closestRate float64
    var closestDiff float64 = math.MaxFloat64
    
    for standardRate := range standardFrameRates {
        diff := math.Abs(frameRate - standardRate)
        if diff < closestDiff {
            closestDiff = diff
            closestRate = standardRate
        }
    }
    
    // Allow small tolerance for frame rate matching
    if closestDiff < 0.1 {
        return standardFrameRates[closestRate], nil
    }
    
    return "", fmt.Errorf("unsupported frame rate: %f", frameRate)
}
```

### Audio Track Detection

```go
func detectAudioProperties(videoPath string) (AudioInfo, error) {
    // Use ffprobe or similar to detect audio properties
    cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", videoPath)
    output, err := cmd.Output()
    if err != nil {
        return AudioInfo{}, err
    }
    
    var probe FFProbeOutput
    if err := json.Unmarshal(output, &probe); err != nil {
        return AudioInfo{}, err
    }
    
    audioInfo := AudioInfo{HasAudio: false}
    
    for _, stream := range probe.Streams {
        if stream.CodecType == "audio" {
            audioInfo.HasAudio = true
            audioInfo.Channels = stream.Channels
            audioInfo.SampleRate = stream.SampleRate
            break
        }
    }
    
    return audioInfo, nil
}
```

## Best Practices Summary

1. **Always detect media type first** before creating assets or timeline elements
2. **Use correct timeline element** - Video for images, AssetClip for videos/audio
3. **Omit frameDuration for images** - Images are timeless
4. **Include frameDuration for videos** - Videos need frame rate specification
5. **Detect actual properties** - Don't hardcode dimensions, frame rates, or audio presence
6. **Use absolute file paths** - Relative paths cause missing media errors
7. **Validate frame rates** - Map to standard FCP frame rates
8. **Handle portrait videos correctly** - Detect orientation and set dimensions properly

**Remember: Getting the media type architecture wrong is the #1 cause of FCP import crashes. Always use the correct element type for each media type.**