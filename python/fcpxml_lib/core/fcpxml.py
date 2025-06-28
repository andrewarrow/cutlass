"""
Core FCPXML document handling.
"""

import sys
from pathlib import Path

from ..models.elements import Resources, Library, Format, Sequence, Project, Event, FCPXML, Asset, MediaRep, SmartCollection, AdjustTransform
from ..constants import (
    FCPXML_VERSION, DEFAULT_SEQUENCE_SETTINGS, STANDARD_FRAME_DURATION,
    VIDEO_COLOR_SPACE, REQUIRED_SMART_COLLECTIONS, IMAGE_DURATION,
    IMAGE_FORMAT_NAME, IMAGE_COLOR_SPACE, DEFAULT_IMAGE_WIDTH, DEFAULT_IMAGE_HEIGHT,
    DEFAULT_VIDEO_WIDTH, DEFAULT_VIDEO_HEIGHT, DEFAULT_VIDEO_DURATION, IMAGE_START_TIME,
    STANDARD_FRAME_RATE, IMAGE_EXTENSIONS, VIDEO_EXTENSIONS,
    VERTICAL_FORMAT_WIDTH, VERTICAL_FORMAT_HEIGHT, HORIZONTAL_FORMAT_WIDTH, HORIZONTAL_FORMAT_HEIGHT,
    VERTICAL_SCALE_FACTOR
)
from ..utils.ids import generate_uid
from ..utils.timing import convert_seconds_to_fcp_duration
from ..serialization.xml_serializer import serialize_to_xml
from ..validation.xml_validator import run_xml_validation


def create_empty_project(project_name: str = "New Project", event_name: str = "New Event", 
                        use_horizontal: bool = False) -> FCPXML:
    """
    Create an empty FCPXML project following crash prevention rules.
    
    This generates the minimal valid FCPXML structure that will import into Final Cut Pro
    without errors. Uses established crash prevention patterns.
    
    Args:
        project_name: Name of the project
        event_name: Name of the event
        use_horizontal: If True, use 1280x720 horizontal format. If False, use 1080x1920 vertical (default)
    """
    
    # Choose dimensions based on format preference
    if use_horizontal:
        width = HORIZONTAL_FORMAT_WIDTH
        height = HORIZONTAL_FORMAT_HEIGHT
        format_name = "FFVideoFormat720p2398"
    else:
        width = VERTICAL_FORMAT_WIDTH
        height = VERTICAL_FORMAT_HEIGHT
        format_name = "FFVideoFormat1080p2398Vertical"
    
    # Create format resource with chosen dimensions
    format_def = Format(
        id="r1",
        name=format_name,
        frame_duration=STANDARD_FRAME_DURATION,
        width=width, 
        height=height,
        color_space=VIDEO_COLOR_SPACE
    )
    
    # Create sequence with the format
    sequence = Sequence(
        format="r1",
        duration="0s",
        tc_start=DEFAULT_SEQUENCE_SETTINGS["tc_start"],
        tc_format=DEFAULT_SEQUENCE_SETTINGS["tc_format"],
        audio_layout=DEFAULT_SEQUENCE_SETTINGS["audio_layout"],
        audio_rate=DEFAULT_SEQUENCE_SETTINGS["audio_rate"]
    )
    
    # Create project containing the sequence
    project = Project(
        name=project_name,
        sequences=[sequence]
    )
    
    # Create event containing the project
    event = Event(
        name=event_name,
        projects=[project]
    )
    
    # Create standard smart collections (required by FCP)
    smart_collections = [
        SmartCollection(name=sc["name"], match=sc["match"], rules=sc["rules"])
        for sc in REQUIRED_SMART_COLLECTIONS
    ]
    
    # Create library containing the event
    library = Library(
        location="file:///Users/aa/dev/cutlass/python/",
        events=[event],
        smart_collections=smart_collections
    )
    
    # Create resources with the format
    resources = Resources(
        formats=[format_def]
    )
    
    # Create root FCPXML document
    fcpxml = FCPXML(
        version=FCPXML_VERSION,
        resources=resources,
        library=library
    )
    
    return fcpxml


def detect_video_properties(file_path: str) -> dict:
    """
    Detect actual video properties to prevent FCP crashes.
    
    🚨 CRITICAL: This follows CLAUDE.md validation rules:
    - Detect actual video properties instead of hardcoding
    - Return safe defaults if detection fails
    - NEVER assume audio exists (causes crashes)
    """
    import subprocess
    
    try:
        # Get video properties using ffprobe
        cmd = [
            "ffprobe", "-v", "error", "-select_streams", "v:0",
            "-show_entries", "stream=width,height,duration,r_frame_rate,codec_name",
            "-of", "csv=p=0", str(file_path)
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        video_info = result.stdout.strip().split(',')
        
        if len(video_info) >= 5:
            codec, width, height, frame_rate_str, duration_str = video_info[:5]
            
            # Parse frame rate (e.g., "29821/994" -> 30.0)
            if '/' in frame_rate_str:
                num, den = map(int, frame_rate_str.split('/'))
                frame_rate = num / den
            else:
                frame_rate = float(frame_rate_str)
            
            # Check for audio streams
            audio_cmd = [
                "ffprobe", "-v", "error", "-select_streams", "a",
                "-show_entries", "stream=codec_name", "-of", "csv=p=0", str(file_path)
            ]
            audio_result = subprocess.run(audio_cmd, capture_output=True, text=True)
            has_audio = bool(audio_result.stdout.strip())
            
            return {
                "duration_seconds": float(duration_str),
                "width": int(width),
                "height": int(height),
                "frame_rate": frame_rate,
                "has_audio": has_audio
            }
    
    except Exception as e:
        print(f"⚠️  Failed to detect properties for {file_path}: {e}")
    
    # Return safe defaults if detection fails
    return {
        "duration_seconds": DEFAULT_VIDEO_DURATION,
        "width": DEFAULT_VIDEO_WIDTH,
        "height": DEFAULT_VIDEO_HEIGHT,
        "frame_rate": STANDARD_FRAME_RATE,
        "has_audio": False  # Safe default: no audio
    }


def create_media_asset(file_path: str, asset_id: str, format_id: str, clip_duration_seconds: float = 5.0) -> tuple[Asset, Format]:
    """
    Create media asset and format following CLAUDE.md validation rules.
    
    🚨 CRITICAL: Follows Images vs Videos Architecture from CLAUDE.md
    - Images: duration="0s", no frameDuration, use Video element
    - Videos: has duration and frameDuration, use AssetClip element
    - ALWAYS use actual video properties, never hardcode
    """
    abs_path = Path(file_path).resolve()
    if not abs_path.exists():
        raise FileNotFoundError(f"Media file not found: {abs_path}")
    
    # Generate UID based on file path for consistency
    uid = generate_uid(f"MEDIA_{abs_path.name}")
    
    # Detect media type
    is_image = abs_path.suffix.lower() in IMAGE_EXTENSIONS
    is_video = abs_path.suffix.lower() in VIDEO_EXTENSIONS
    
    if not (is_image or is_video):
        raise ValueError(f"Unsupported media type: {abs_path.suffix}")
    
    # Create MediaRep with absolute file:// URL
    media_rep = MediaRep(src=str(abs_path))
    
    if is_image:
        # Images: duration="0s" (timeless), no frameDuration
        asset = Asset(
            id=asset_id,
            name=abs_path.stem,
            uid=uid,
            duration=IMAGE_DURATION,  # Timeless for images
            has_video="1",
            format=format_id,
            video_sources="1",
            media_rep=media_rep
        )
        
        # Format: NO frameDuration (timeless)
        format_obj = Format(
            id=format_id,
            name=IMAGE_FORMAT_NAME,
            width=DEFAULT_IMAGE_WIDTH,
            height=DEFAULT_IMAGE_HEIGHT,
            color_space=IMAGE_COLOR_SPACE
        )
        
    else:  # is_video
        # 🚨 CRITICAL: Detect actual video properties to prevent crashes
        props = detect_video_properties(file_path)
        actual_duration = convert_seconds_to_fcp_duration(props["duration_seconds"])
        
        # Videos: Use ACTUAL properties but NO audio properties (per Go patterns)
        # 🚨 CRITICAL: Videos NEVER have hasAudio/audioSources in Go implementation
        asset = Asset(
            id=asset_id,
            name=abs_path.stem,
            uid=uid,
            duration=actual_duration,  # Use actual video duration
            has_video="1",
            # 🚨 REMOVED: NO audio properties on video assets (prevents crashes)
            format=format_id,
            video_sources="1",
            media_rep=media_rep
        )
        
        # Format: Use ACTUAL properties but NO name attribute (per Go patterns)
        # 🚨 CRITICAL: Video formats in Go have NO name attribute
        format_obj = Format(
            id=format_id,
            # 🚨 REMOVED: NO name attribute for video formats (per Go patterns)
            frame_duration=STANDARD_FRAME_DURATION,
            width=str(props["width"]),
            height=str(props["height"]),
            color_space=VIDEO_COLOR_SPACE
        )
    
    return asset, format_obj


def add_media_to_timeline(fcpxml: FCPXML, media_files: list[str], clip_duration_seconds: float = 5.0, use_horizontal: bool = False):
    """
    Add media files to timeline following CLAUDE.md rules.
    
    🚨 CRITICAL: Uses correct element types:
    - Images: Video element (NOT AssetClip) 
    - Videos: AssetClip element
    - Elements MUST be ordered by start time in spine
    - In vertical format, applies 3.27x scale to fill screen
    
    Args:
        fcpxml: The FCPXML document
        media_files: List of media file paths
        clip_duration_seconds: Duration for each clip
        use_horizontal: If False (default), apply vertical scaling for 1080x1920
    """
    if not fcpxml.library or not fcpxml.library.events:
        raise ValueError("FCPXML must have library and events")
    
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    timeline_position = 0.0
    resource_counter = len(fcpxml.resources.assets) + len(fcpxml.resources.formats) + 1
    
    # Collect all timeline elements to sort by start time
    all_timeline_elements = []
    
    for media_file in media_files:
        try:
            # Generate unique IDs
            asset_id = f"r{resource_counter}"
            format_id = f"r{resource_counter + 1}"
            resource_counter += 2
            
            # Create asset and format
            asset, format_obj = create_media_asset(media_file, asset_id, format_id, clip_duration_seconds)
            
            # 🚨 CRITICAL VALIDATION: Prevent AssetClip crash patterns
            image_extensions = {'.jpg', '.jpeg', '.png', '.tiff', '.bmp', '.gif'}
            is_image = Path(media_file).suffix.lower() in image_extensions
            
            # Validate against crash patterns from CLAUDE.md
            if is_image and asset.duration != "0s":
                print(f"⚠️  WARNING: Image asset {asset_id} has non-zero duration, fixing...")
                asset.duration = "0s"
            
            if is_image and format_obj.frame_duration:
                print(f"⚠️  WARNING: Image format {format_id} has frameDuration, fixing...")
                format_obj.frame_duration = None
            
            # Add to resources
            fcpxml.resources.assets.append(asset)
            fcpxml.resources.formats.append(format_obj)
            
            if is_image:
                # Images: Use Video element with offset and start attributes
                clip_duration = convert_seconds_to_fcp_duration(clip_duration_seconds)
                # 🚨 CRITICAL: Use frame boundary value from working samples
                # All working samples use "3600s" for Video elements
                start_time = IMAGE_START_TIME  # Standard frame boundary used by FCP
                
                element = {
                    "type": "video",
                    "ref": asset_id,
                    "duration": clip_duration,
                    "offset": convert_seconds_to_fcp_duration(timeline_position),
                    "start": start_time,  # Use specific timing pattern from samples
                    "name": Path(media_file).stem,
                    "start_time": timeline_position  # For sorting
                }
                
                # Add scaling for vertical format to fill screen (images need scaling too)
                if not use_horizontal:
                    element["adjust_transform"] = {"scale": VERTICAL_SCALE_FACTOR}
            else:
                # Videos: Use AssetClip element with NO start attribute
                clip_duration = convert_seconds_to_fcp_duration(clip_duration_seconds)
                element = {
                    "type": "asset-clip", 
                    "ref": asset_id,
                    "duration": clip_duration,  # Use clip duration
                    "offset": convert_seconds_to_fcp_duration(timeline_position),
                    # 🚨 REMOVED: AssetClips don't need start attribute per samples/simple_video1.fcpxml
                    "name": Path(media_file).stem,
                    "start_time": timeline_position  # For sorting
                }
                
                # Add scaling for vertical format to fill screen
                if not use_horizontal:
                    element["adjust_transform"] = {"scale": VERTICAL_SCALE_FACTOR}
            
            all_timeline_elements.append(element)
            timeline_position += clip_duration_seconds
            
        except Exception as e:
            print(f"⚠️  Skipping {media_file}: {e}")
            continue
    
    # 🚨 CRITICAL: Sort elements by start time (required by FCP)
    all_timeline_elements.sort(key=lambda x: x["start_time"])
    
    # Store elements in a single list for proper spine ordering
    sequence.spine.ordered_elements = []
    
    # Add sorted elements to spine (preserving order for serializer)
    for element in all_timeline_elements:
        if element["type"] == "video":
            video_clip = {
                "type": "video",
                "ref": element["ref"],
                "duration": element["duration"],
                "offset": element["offset"],
                "start": element["start"],  # Include start attribute for Video elements
                "name": element["name"]
            }
            
            # Add transform if present
            if "adjust_transform" in element:
                video_clip["adjust_transform"] = element["adjust_transform"]
            
            sequence.spine.videos.append(video_clip)
            sequence.spine.ordered_elements.append(video_clip)
        else:  # asset-clip
            asset_clip = {
                "type": "asset-clip",
                "ref": element["ref"],
                "duration": element["duration"],
                "offset": element["offset"],
                # NO start attribute for AssetClip elements
                "name": element["name"]
            }
            
            # Add transform if present
            if "adjust_transform" in element:
                asset_clip["adjust_transform"] = element["adjust_transform"]
            
            sequence.spine.asset_clips.append(asset_clip)
            sequence.spine.ordered_elements.append(asset_clip)
    
    # Update sequence duration
    total_duration = convert_seconds_to_fcp_duration(timeline_position)
    sequence.duration = total_duration


def save_fcpxml(fcpxml: FCPXML, output_path: str) -> bool:
    """
    Save FCPXML document to file and validate it.
    
    Returns True if successful and well-formed, False otherwise.
    🚨 CRITICAL: XML validation is mandatory for crash prevention
    """
    xml_content = serialize_to_xml(fcpxml)
    
    # Add XML declaration (no DTD for now as it requires Apple's server)
    fcpxml_with_header = f'''<?xml version="1.0" encoding="UTF-8"?>
{xml_content}'''
    
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(fcpxml_with_header)
    
    print(f"📄 FCPXML saved to: {output_path}")
    
    # Run basic XML validation
    print("🔍 Running XML well-formedness validation...")
    is_valid, error_msg = run_xml_validation(output_path)
    
    if is_valid:
        print("✅ XML VALIDATION PASSED")
        print("⚠️  Note: For full DTD validation, test import in Final Cut Pro")
        return True
    else:
        print("\n" + "="*60)
        print("🚨 VALIDATION FAILED - XML ERRORS DETECTED")
        print("="*60)
        print(f"❌ XML Error: {error_msg}")
        print("\n⚠️  FCPXML will likely fail to import into Final Cut Pro!")
        print("   Fix the validation errors before using this file.")
        print("="*60 + "\n")
        return False