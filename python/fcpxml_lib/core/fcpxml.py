"""
Core FCPXML document handling.
"""

import sys
from pathlib import Path

from ..models.elements import Resources, Library, Format, Sequence, Project, Event, FCPXML, Asset, MediaRep
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
    
    # Create library containing the event
    library = Library(
        events=[event]
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


def create_media_asset(file_path: str, asset_id: str, format_id: str, duration_seconds: float = 5.0) -> tuple[Asset, Format]:
    """
    Create media asset and format following CLAUDE.md validation rules.
    
    🚨 CRITICAL: Follows Images vs Videos Architecture from CLAUDE.md
    - Images: duration="0s", no frameDuration, use Video element
    - Videos: has duration and frameDuration, use AssetClip element
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
        # Videos: has duration and audio properties
        duration_fcp = convert_seconds_to_fcp_duration(duration_seconds)
        
        asset = Asset(
            id=asset_id,
            name=abs_path.stem,
            uid=uid,
            duration=duration_fcp,
            has_video="1",
            has_audio="1",  # Assume videos have audio
            format=format_id,
            video_sources="1",
            audio_sources="1",
            media_rep=media_rep
        )
        
        # Format: has frameDuration
        format_obj = Format(
            id=format_id,
            name="FFVideoFormat1080p2398",
            frame_duration="1001/24000s",  # 23.976 fps
            width="1920",
            height="1080",
            color_space="1-1-1 (Rec. 709)"
        )
    
    return asset, format_obj


def add_media_to_timeline(fcpxml: FCPXML, media_files: list[str], clip_duration_seconds: float = 5.0):
    """
    Add media files to timeline following CLAUDE.md rules.
    
    🚨 CRITICAL: Uses correct element types:
    - Images: Video element (NOT AssetClip)
    - Videos: AssetClip element
    """
    if not fcpxml.library or not fcpxml.library.events:
        raise ValueError("FCPXML must have library and events")
    
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    timeline_position = 0.0
    resource_counter = len(fcpxml.resources.assets) + len(fcpxml.resources.formats) + 1
    
    for media_file in media_files:
        try:
            # Generate unique IDs
            asset_id = f"r{resource_counter}"
            format_id = f"r{resource_counter + 1}"
            resource_counter += 2
            
            # Create asset and format
            asset, format_obj = create_media_asset(media_file, asset_id, format_id, clip_duration_seconds)
            
            # Add to resources
            fcpxml.resources.assets.append(asset)
            fcpxml.resources.formats.append(format_obj)
            
            # Detect media type for spine element
            image_extensions = {'.jpg', '.jpeg', '.png', '.tiff', '.bmp', '.gif'}
            is_image = Path(media_file).suffix.lower() in image_extensions
            
            if is_image:
                # Images: Use Video element with specific duration
                clip_duration = convert_seconds_to_fcp_duration(clip_duration_seconds)
                video_clip = {
                    "ref": asset_id,
                    "duration": clip_duration,
                    "start": convert_seconds_to_fcp_duration(timeline_position)
                }
                sequence.spine.videos.append(video_clip)
            else:
                # Videos: Use AssetClip element
                clip_duration = convert_seconds_to_fcp_duration(clip_duration_seconds)
                asset_clip = {
                    "ref": asset_id,
                    "duration": clip_duration,
                    "start": convert_seconds_to_fcp_duration(timeline_position)
                }
                sequence.spine.asset_clips.append(asset_clip)
            
            timeline_position += clip_duration_seconds
            
        except Exception as e:
            print(f"⚠️  Skipping {media_file}: {e}")
            continue
    
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