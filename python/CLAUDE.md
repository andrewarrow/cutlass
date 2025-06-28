# Project Context for AI Assistance - Python FCPXML Library

## 🚨 CRITICAL: ALL CODE MUST USE VALIDATION 🚨
**There are extensive validation checks in `fcpxml_lib/validation/`. Never bypass these. Better to let an error stop generation because validation failed than to ever produce invalid FCPXML.**

## 🚨 CRITICAL: CHANGE CODE NOT XML 🚨
**NEVER EVER only change problem XML in an XML file, always change the code that generates it too**

## 🚨 CRITICAL: NO XML STRING TEMPLATES 🚨
**NEVER EVER generate XML from hardcoded string templates with f-strings or % formatting, use dataclasses**

❌ BAD: `xml = f"<video ref=\"{video_ref}\">{content}</video>"`
❌ BAD: `"<asset-clip ref=\"%s\" name=\"%s\"/>" % (ref, name)`
✅ GOOD: `Video(ref=video_ref, duration=duration)` → `serialize_to_xml()`

**All FCPXML generation MUST use the dataclasses in `fcpxml_lib/models/elements.py`.**

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

## 🚨 CRITICAL: Python Dataclass Usage 🚨

**Use dataclasses from `fcpxml_lib/models/elements.py`:**

```python
from fcpxml_lib.models.elements import Video, AssetClip, Asset, Format

# ✅ Images use Video elements
video_element = Video(
    ref="r2",
    duration="240240/24000s",
    start="0s"  # Required for images
)

# ✅ Videos use AssetClip elements  
asset_clip = AssetClip(
    ref="r2", 
    duration="373400/3000s"
    # NO start attribute for videos
)
```

## MANDATORY: Testing and Validation

### Required Tests (ALWAYS run):
1. **Python Tests**: `python -m pytest tests/` - MUST pass
2. **XML Validation**: `xmllint output.fcpxml --noout` - MUST pass  
3. **FCP Import Test**: Import into actual Final Cut Pro

### Study tests/* Package Patterns
**Before writing FCPXML code, review the logic in `tests/` files:**
- `tests/test_crash_prevention.py` - Shows critical crash prevention patterns
- `tests/test_timeline_elements.py` - Shows correct element type usage
- `tests/test_media_detection.py` - Shows proper media property detection
- These tests contain proven patterns that prevent crashes

### Common Error Patterns to Check:
1. **ID collisions** - Use proper ID generation from `fcpxml_lib/utils/ids.py`
2. **Missing resources** - Every `ref=` needs matching `id=`
3. **Wrong element types** - Images use Video, videos use AssetClip
4. **Missing smart collections** - All 5 required collections must be present

## 🏗️ Required Architecture Pattern

**ALWAYS follow this pattern (from `fcpxml_lib/core/fcpxml.py`):**

```python
from fcpxml_lib.core.fcpxml import create_empty_project, add_media_to_timeline, save_fcpxml

def generate_my_feature(input_files: list, output_file: str) -> bool:
    # 1. Create base FCPXML structure
    fcpxml = create_empty_project()
    
    # 2. Add media using validated functions
    success = add_media_to_timeline(
        fcpxml, 
        input_files, 
        clip_duration_seconds=3.0
    )
    if not success:
        raise RuntimeError("Failed to add media to timeline")
    
    # 3. Apply modifications to timeline elements
    # Access via: fcpxml.library.events[0].projects[0].sequences[0].spine
    
    # 4. Save with validation
    return save_fcpxml(fcpxml, output_file)
```

## Frame Boundary Alignment

**All durations MUST use `fcpxml_lib/utils/timing.py`:**
```python
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration

# ✅ Proper frame alignment
duration = convert_seconds_to_fcp_duration(3.5)  # "84084/24000s"

# ❌ Manual calculations cause "not on edit frame boundary" errors
duration = "3.5s"  # Wrong!
```

## Resource ID Management

**Use proper ID generation from `fcpxml_lib/utils/ids.py`:**
```python
from fcpxml_lib.utils.ids import generate_resource_id

# ✅ Thread-safe ID generation
asset_id = generate_resource_id()  # "r1", "r2", etc.

# ❌ Manual ID generation causes collisions
asset_id = f"r{count + 1}"  # Race conditions!
```

## 🚨 CRITICAL: Media Property Detection 🚨

**ALWAYS use `fcpxml_lib/core/fcpxml.py` detection functions:**

```python
from fcpxml_lib.core.fcpxml import detect_video_properties

# ✅ Automatic property detection
props = detect_video_properties("video.mp4")
# Returns: width, height, duration_seconds, has_audio, frame_rate

# ❌ Hardcoded properties cause import failures  
width = 1920  # Video might be 1080×1920 portrait!
has_audio = True  # Video might not have audio!
```

## 🚨 CRITICAL: Media Asset Creation 🚨

**Use `fcpxml_lib/core/fcpxml.py` functions for proper media handling:**

```python
from fcpxml_lib.core.fcpxml import create_media_asset

# ✅ Automatic media type detection and proper asset creation
asset, format_obj = create_media_asset(
    file_path="video.mp4",
    asset_id="r2", 
    format_id="r3"
)
# Handles: absolute paths, media type detection, audio properties

# ❌ Manual asset creation bypasses validation
asset = Asset(id="r2", src="./video.mp4")  # Relative path fails!
```

## 🚨 CRITICAL: Smart Collections Requirement 🚨

**All FCPXML must include 5 required smart collections (see `fcpxml_lib/constants.py`):**

```python
# ✅ Automatically included by create_empty_project()
fcpxml = create_empty_project()
# Includes: Projects, All Video, Audio Only, Stills, Favorites

# ❌ Missing smart collections cause FCP crashes
# Never create FCPXML without all 5 collections
```

## 🚨 CRITICAL: Timeline Element Rules 🚨

**Images vs Videos have different timeline element requirements:**

```python
from fcpxml_lib.models.elements import Video, AssetClip

# ✅ Images: Use Video elements with start attribute
if file_path.lower().endswith(('.jpg', '.png')):
    element = Video(
        ref=asset_id,
        duration=duration,
        start="0s"  # Required for images
    )

# ✅ Videos: Use AssetClip elements without start attribute  
elif file_path.lower().endswith(('.mp4', '.mov')):
    element = AssetClip(
        ref=asset_id,
        duration=duration
        # NO start attribute for videos
    )
```

## Validation Integration

**All dataclasses have built-in validation (see `fcpxml_lib/validation/`):**

```python
from fcpxml_lib.validation.validators import validate_frame_alignment

# ✅ Validation is automatic in dataclass __post_init__
asset = Asset(id="r2", duration="invalid")  # Raises ValueError

# ✅ Manual validation for complex cases
is_valid = validate_frame_alignment("84084/24000s")  # True
```

## XML Generation

**Use structured serialization from `fcpxml_lib/serialization/`:**

```python
from fcpxml_lib.serialization.xml_serializer import serialize_to_xml

# ✅ Structured XML generation
xml_content = serialize_to_xml(fcpxml)

# ❌ Never manipulate XML strings directly
xml_content = xml_content.replace("<video", "<asset-clip")  # Wrong!
```

## Testing Integration

**Run tests to verify crash prevention patterns:**

```python
# Required test commands:
python -m pytest tests/test_crash_prevention.py -v
python -m pytest tests/test_timeline_elements.py -v  
python -m pytest tests/test_xml_structure.py -v

# XML validation:
xmllint output.fcpxml --noout
```

## 🚨 CRITICAL: File Path Requirements 🚨

**Final Cut Pro requires absolute file paths:**

```python
import os

# ✅ Always convert to absolute paths
abs_path = os.path.abspath(file_path)
media_rep = MediaRep(src=f"file://{abs_path}")

# ❌ Relative paths cause "missing media" errors
media_rep = MediaRep(src="file://./video.mp4")  # Fails!
```

## Effect and Animation Safety

**Prefer built-in transforms over effects:**

```python
from fcpxml_lib.models.elements import AdjustTransform

# ✅ Built-in transforms are crash-safe
transform = AdjustTransform(
    scale="2.0 2.0",
    rotation="15.0"
)

# ❌ Avoid fictional effect UIDs - causes crashes
# Only use verified UIDs from samples/
```

---

**Key Principle: Use `fcpxml_lib/` functions and validation. If FCPXML generation requires more than 1 iteration to work, you're using the wrong approach.**

**Critical File References:**
- **Core Functions**: `fcpxml_lib/core/fcpxml.py`
- **Data Models**: `fcpxml_lib/models/elements.py` 
- **Validation**: `fcpxml_lib/validation/validators.py`
- **Testing Patterns**: `tests/test_crash_prevention.py`
- **Utilities**: `fcpxml_lib/utils/timing.py`, `fcpxml_lib/utils/ids.py`