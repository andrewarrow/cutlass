"""
FCPXML Constants and Crash Prevention Rules.

These constants encode the exact patterns needed to prevent Final Cut Pro crashes,
identified through analysis of the working Go implementation.
"""

# FCPXML Version (critical for compatibility)
FCPXML_VERSION = "1.13"

# Frame timing constants (required for proper FCP compatibility)
STANDARD_FRAME_DURATION = "1001/24000s"  # ~23.976 fps (FCP standard)
STANDARD_TIMEBASE = 24000
STANDARD_FRAME_RATE = 23.976023976023976

# Image timing constants (critical for crash prevention)
IMAGE_DURATION = "0s"  # Images MUST have duration="0s"
IMAGE_START_TIME = "3600s"  # Standard start time for image elements

# Color spaces
VIDEO_COLOR_SPACE = "1-1-1 (Rec. 709)"  # For videos
IMAGE_COLOR_SPACE = "1-13-1"  # For images

# Format names
IMAGE_FORMAT_NAME = "FFVideoFormatRateUndefined"  # For images (no frameDuration)
# Video formats have empty name (per Go patterns)

# Default project settings
DEFAULT_SEQUENCE_SETTINGS = {
    "tc_start": "0s",
    "tc_format": "NDF",
    "audio_layout": "stereo",
    "audio_rate": "48k"
}

# Default image dimensions (fallback when detection fails)
DEFAULT_IMAGE_WIDTH = "1280"
DEFAULT_IMAGE_HEIGHT = "720"

# Default video properties (fallback when detection fails)
DEFAULT_VIDEO_WIDTH = 1920
DEFAULT_VIDEO_HEIGHT = 1080
DEFAULT_VIDEO_DURATION = 10.0  # seconds

# Resource ID pattern validation
RESOURCE_ID_PATTERN = r"^r\d+$"  # r1, r2, r3, etc.

# Valid audio rates (FCP enumerated values)
VALID_AUDIO_RATES = ["32k", "44.1k", "48k", "88.2k", "96k", "176.4k", "192k"]

# 🚨 CRITICAL: Required Smart Collections (prevents FCP crashes)
REQUIRED_SMART_COLLECTIONS = [
    {
        "name": "Projects",
        "match": "all",
        "rules": [{"rule": "is", "type": "project"}]
    },
    {
        "name": "All Video",
        "match": "any", 
        "rules": [
            {"rule": "is", "type": "videoOnly"},
            {"rule": "is", "type": "videoWithAudio"}
        ]
    },
    {
        "name": "Audio Only",
        "match": "all",
        "rules": [{"rule": "is", "type": "audioOnly"}]
    },
    {
        "name": "Stills",
        "match": "all",
        "rules": [{"rule": "is", "type": "stills"}]
    },
    {
        "name": "Favorites", 
        "match": "all",
        "rules": [{"rule": "favorites", "value": "favorites"}]
    }
]

# File extension mappings
IMAGE_EXTENSIONS = {".png", ".jpg", ".jpeg"}
VIDEO_EXTENSIONS = {".mp4", ".mov"}
AUDIO_EXTENSIONS = {".wav", ".mp3", ".m4a", ".aac", ".flac", ".caf"}

# 🚨 CRITICAL CRASH PREVENTION RULES:
"""
1. Video assets NEVER have hasAudio/audioSources properties
2. Image assets MUST have duration="0s" 
3. Image formats NEVER have frameDuration
4. Video formats have empty name attribute
5. Smart collections are MANDATORY (all 5 required)
6. Use absolute file paths only
7. Images use <video> elements, videos use <asset-clip> elements
8. Image <video> elements need start attribute, <asset-clip> elements don't
"""