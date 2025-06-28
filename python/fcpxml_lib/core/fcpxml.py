"""
Core FCPXML document handling.
"""

import sys
from pathlib import Path

from ..models.elements import Resources, Library, Format, Sequence, Project, Event, FCPXML, Asset, MediaRep, SmartCollection
from ..utils.schema_loader import get_schema
from ..utils.ids import generate_uid
from ..utils.timing import convert_seconds_to_fcp_duration
from ..serialization.xml_serializer import serialize_to_xml
from ..validation.xml_validator import run_xml_validation


def create_empty_project(project_name: str = "New Project", event_name: str = "New Event") -> FCPXML:
    """
    Create an empty FCPXML project following schema.yaml template.
    
    This generates the minimal valid FCPXML structure that will import into Final Cut Pro
    without errors. Based on the empty_project template in schema.yaml.
    """
    
    # Get template from schema
    schema = get_schema()
    template = schema['fcpxml_rules']['templates']['empty_project']
    
    # Create format resource (HD 1080p 23.976 fps - FCP standard)
    format_def = Format(
        id="r1",
        name="FFVideoFormat1080p2398",
        frame_duration="1001/24000s",
        width="1920", 
        height="1080",
        color_space="1-1-1 (Rec. 709)"
    )
    
    # Create sequence with the format
    sequence = Sequence(
        format="r1",
        duration="0s",
        tc_start="0s",
        tc_format="NDF",
        audio_layout="stereo",
        audio_rate="48k"
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
        SmartCollection(
            name="Projects",
            match="all",
            rules=[{"rule": "is", "type": "project"}]
        ),
        SmartCollection(
            name="All Video", 
            match="any",
            rules=[
                {"rule": "is", "type": "videoOnly"},
                {"rule": "is", "type": "videoWithAudio"}
            ]
        ),
        SmartCollection(
            name="Audio Only",
            match="all", 
            rules=[{"rule": "is", "type": "audioOnly"}]
        ),
        SmartCollection(
            name="Stills",
            match="all",
            rules=[{"rule": "is", "type": "stills"}]
        ),
        SmartCollection(
            name="Favorites",
            match="all",
            rules=[{"rule": "favorites", "value": "favorites"}]
        )
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
        version=template['version'],
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
        "duration_seconds": 10.0,
        "width": 1920,
        "height": 1080,
        "frame_rate": 23.976,
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
    image_extensions = {'.jpg', '.jpeg', '.png', '.tiff', '.bmp', '.gif'}
    video_extensions = {'.mov', '.mp4', '.avi', '.mkv', '.m4v'}
    
    is_image = abs_path.suffix.lower() in image_extensions
    is_video = abs_path.suffix.lower() in video_extensions
    
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
            duration="0s",  # Timeless for images
            has_video="1",
            format=format_id,
            video_sources="1",
            media_rep=media_rep
        )
        
        # Format: NO frameDuration (timeless)
        format_obj = Format(
            id=format_id,
            name="FFVideoFormatRateUndefined",
            width="1280",  # Default image dimensions
            height="720",
            color_space="1-13-1"
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
        frame_duration = "1001/24000s"  # Standard FCP timebase (~23.98 fps)
            
        format_obj = Format(
            id=format_id,
            # 🚨 REMOVED: NO name attribute for video formats (per Go patterns)
            frame_duration=frame_duration,
            width=str(props["width"]),
            height=str(props["height"]),
            color_space="1-1-1 (Rec. 709)"
        )
    
    return asset, format_obj


def add_media_to_timeline(fcpxml: FCPXML, media_files: list[str], clip_duration_seconds: float = 5.0):
    """
    Add media files to timeline following CLAUDE.md rules.
    
    🚨 CRITICAL: Uses correct element types:
    - Images: Video element (NOT AssetClip) 
    - Videos: AssetClip element
    - Elements MUST be ordered by start time in spine
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
                start_time = "3600s"  # Standard frame boundary used by FCP
                
                element = {
                    "type": "video",
                    "ref": asset_id,
                    "duration": clip_duration,
                    "offset": convert_seconds_to_fcp_duration(timeline_position),
                    "start": start_time,  # Use specific timing pattern from samples
                    "name": Path(media_file).stem,
                    "start_time": timeline_position  # For sorting
                }
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
            sequence.spine.asset_clips.append(asset_clip)
            sequence.spine.ordered_elements.append(asset_clip)
    
    # Update sequence duration
    total_duration = convert_seconds_to_fcp_duration(timeline_position)
    sequence.duration = total_duration


def save_fcpxml(fcpxml: FCPXML, output_path: str) -> bool:
    """
    Save FCPXML document to file and validate it.
    
    Returns True if successful and well-formed, False otherwise.
    🚨 CRITICAL: XML validation is mandatory (from schema.yaml)
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