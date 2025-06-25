# 🗡️ Cutlass: The FCPXML Surgeon

> *"In a world full of video editing tools that make you click buttons, Cutlass wields the ancient art of **procedural FCPXML generation**. Meet the timeline's new best friend."*

![Wikipedia to Video Magic](https://i.imgur.com/mcAUx49.png)

**What you see above:** A boring Wikipedia table about Andre Agassi's tennis career.  
**What Cutlass sees:** Raw material for a procedurally generated video masterpiece.

Run this:
```bash
./cutlass wikipedia "Andre_Agassi"
```

Get this in Final Cut Pro:

![Generated Video](https://i.imgur.com/8CQmlQ4.png)

Those red animated lines? Each one precisely calculated and timed using `<adjust-transform>` keyframes, because Cutlass doesn't just *make* videos—it **architects** them at the XML level.

---

## 🧠 The FCPXML Whisperer

Cutlass isn't just another video tool. It's a **deep-dive expedition** into the mysterious world of Final Cut Pro's XML format—a realm where `<spine>` elements rule kingdoms, `Assets` and `Formats` form diplomatic alliances, and timing is measured in the sacred fractions of `24000/1001s`.

### 🎭 Meet the FCPXML Cast of Characters

In the Final Cut Pro universe, every video is a **story** with these quirky protagonists:

**🏛️ The Spine** - *The Master Timeline*  
Think of the Spine as the movie theater. It's where all the action happens, but it's very particular about seating arrangements. Videos and images each have their own social protocols:

```xml
<spine>
    <!-- Images are divas - they need Video elements -->
    <video ref="r2" duration="240240/24000s">🖼️</video>
    
    <!-- Videos are professionals - they use AssetClip -->
    <asset-clip ref="r4" duration="373400/3000s">🎬</asset-clip>
</spine>
```

**📦 Assets & Formats** - *The Media Managers*  
Assets are like talent agents—they know everything about their media files. Formats are like casting directors—they know exactly what technical specs each star brings to the table:

```go
// Images are timeless zen masters
asset := Asset{
    Duration: "0s",        // "Time is an illusion" - Image Asset
    HasVideo: "1",         // "I contain multitudes" 
    Format: "r3"           // "But no frameDuration, please"
}

// Videos are punctual professionals  
asset := Asset{
    Duration: "14122857/100000s",  // "I know exactly how long I am"
    HasVideo: "1", HasAudio: "1",  // "I multitask"
    FrameDuration: "1001/30000s"   // "And I have a very specific rhythm"
}
```

**⏰ The Timekeeper** - *Master of 24000/1001 Math*  
Final Cut Pro doesn't speak "seconds"—it speaks in **rational time**. Every moment is a fraction, every duration a very `real` number.

```go
// When a human says "5 seconds"
humanTime := 5.0

// The Timekeeper translates to FCP's language  
fcpTime := ConvertSecondsToFCPDuration(5.0)
// Result: "120120/24000s" (because 5 × 24000/1001 = exactly that)
```

**🎨 The Lane Society** - *Vertical Stack Politicians*  
Lanes are like floors in a skyscraper. Lane 0 is the ground floor (main action), Lanes 1, 2, 3 are the upper floors (overlays), and everyone must respect the hierarchy:

```xml
<spine>
    <video ref="r2" name="Background">         <!-- Ground floor -->
        <video ref="r3" lane="1" name="Overlay1"/> <!-- 1st floor -->
        <video ref="r4" lane="2" name="Overlay2"/> <!-- 2nd floor -->
        <title ref="r5" lane="3" name="Title"/>   <!-- Penthouse -->
    </video>
</spine>
```

---

## 🚀 Why Cutlass Exists: The Scale Problem

**Traditional video editing workflow:**
1. Open Final Cut Pro 
2. Import media files
3. Drag clips to timeline  
4. Add titles one by one
5. Adjust each animation keyframe manually
6. Repeat × 100 for scale projects

**Time investment:** Hours to days  
**Scalability:** Limited by human patience  
**Consistency:** Varies by human mood

**Cutlass workflow:**
1. `./cutlass wikipedia "List_of_earthquakes_in_the_United_States"`
2. Open generated `.fcpxml` in Final Cut Pro
3. Professional-grade timeline instantly loaded

**Time investment:** Seconds  
**Scalability:** Infinite (limited by CPU, not sanity)  
**Consistency:** Mathematical precision

---

## 🔬 Technical Deep Dive: The FCPXML Expertise

Cutlass doesn't just generate XML—it's a **full-stack FCPXML architect** with deep knowledge that took awhile to accumulate:

### 🚨 Critical Architecture Decisions

**The #1 Cause of FCP Crashes: Wrong Element Types**

```go
// 💥 CRASH GUARANTEED
// Using AssetClip for images
assetClip := AssetClip{Ref: "image.png"} // ← FCP will explode

// ✅ CUTLASS KNOWS BETTER  
// Images need Video elements
video := Video{
    Ref: imageAssetID,
    Duration: fcp.ConvertSecondsToFCPDuration(5.0), // Cutlass calculates
    AdjustTransform: createImageAnimation(),         // Cutlass animates
}
```

**Keyframe Interpolation Mysteries Solved:**

Different parameters in Final Cut Pro have different "social rules" for keyframes:

```xml
<!-- Position keyframes are antisocial - no attributes allowed -->
<param name="position">
    <keyframe time="0s" value="0 0"/>                    <!-- ✅ Clean -->
    <keyframe time="120120/24000s" value="100 50"/>       <!-- ✅ Perfect -->
</param>

<!-- Scale keyframes are selective - only 'curve' friends allowed -->
<param name="scale">  
    <keyframe time="0s" value="1 1" curve="linear"/>      <!-- ✅ Accepted -->
    <keyframe time="120120/24000s" value="1.5 1.5" curve="smooth"/> <!-- ✅ Welcome -->
</param>
```

Cutlass **knows these rules** and generates pixel-perfect keyframes every time*. *getting there!

### 🧮 The Sacred Math of Frame-Perfect Timing

Final Cut Pro's timing system is like a Swiss watchmaker's dream:

```go
// Humans think in seconds: "I want this to last 3.5 seconds"
humanSeconds := 3.5

// FCP thinks in 24000/1001 fractions: "Ah yes, 84084/24000s precisely"
fcpDuration := fcp.ConvertSecondsToFCPDuration(humanSeconds)
// Result: "84084/24000s" - frame-perfect, mathematically elegant

// Why? Because 3.5 × (24000/1001) = 84084/24000 exactly
// No floating point drift, no audio sync issues, pure rational math
```

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

## 📚 Learn More: FCPXML Resources

Want to dive deeper into the mysterious world of FCPXML? Here are the resources that built Cutlass's expertise:

### Essential Reading
- [FCP.cafe Developer Resources](https://fcp.cafe/developers/fcpxml/) - Community knowledge base
- [Apple FCPXML Reference](https://developer.apple.com/documentation/professional-video-applications/fcpxml-reference) - Official documentation  
- [CommandPost DTD Files](https://github.com/CommandPost/CommandPost/tree/develop/src/extensions/cp/apple/fcpxml/dtd) - XML validation schemas
- [DAWFileKit](https://github.com/orchetect/DAWFileKit) - Swift FCPXML parser

