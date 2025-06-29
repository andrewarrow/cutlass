#!/usr/bin/env python3
"""
Assemble Command

Assembles media files matching a directory pattern into a timeline in filename order.
Takes files like house001.png, house002.jpg, house003.mov and adds them sequentially.
Following comprehensive crash prevention rules and NO_XML_TEMPLATES principle.
"""

import sys
import glob
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml, add_media_to_timeline
)
from fcpxml_lib.utils.media import get_media_type_info


def assemble_cmd(args):
    """Assemble media files matching a pattern into timeline in filename order"""
    
    # Parse directory and pattern
    if '/' in args.pattern:
        # Full path with pattern like ~/dir1/house*.png
        pattern_path = Path(args.pattern).expanduser()
        input_dir = pattern_path.parent
        pattern = pattern_path.name
    else:
        # Just pattern, use provided directory
        input_dir = Path(args.directory).expanduser()
        pattern = args.pattern
    
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Find all files matching the pattern using glob
    search_pattern = str(input_dir / pattern)
    matching_files = glob.glob(search_pattern)
    
    if not matching_files:
        print(f"❌ No files found matching pattern: {search_pattern}")
        sys.exit(1)
    
    # Convert to Path objects and sort by filename
    media_files = [Path(f) for f in matching_files]
    media_files.sort(key=lambda p: p.name.lower())
    
    # Filter to only supported media types
    supported_files = []
    for file_path in media_files:
        media_info = get_media_type_info(file_path)
        if media_info["is_video"] or media_info["is_image"]:
            supported_files.append(file_path)
    
    if not supported_files:
        print(f"❌ No supported media files found")
        print(f"   Supported formats: PNG, JPG, JPEG, MOV, MP4, AVI, MKV, M4V")
        sys.exit(1)
    
    # Default to 1080x1920 (vertical) format as requested
    use_horizontal = args.horizontal if hasattr(args, 'horizontal') else False
    format_desc = "1280x720 horizontal" if use_horizontal else "1080x1920 vertical"
    
    print(f"🎬 Assembling {len(supported_files)} media files in filename order...")
    print(f"   Pattern: {search_pattern}")
    print(f"   Format: {format_desc}")
    print(f"   Files found: {[f.name for f in supported_files[:5]]}{'...' if len(supported_files) > 5 else ''}")
    
    # Create empty project with default vertical format (1080x1920)
    project_name = f"Assembled - {pattern}"
    fcpxml = create_empty_project(
        project_name=project_name,
        event_name="Assembled Media",
        use_horizontal=use_horizontal
    )
    
    # Add media files to timeline in sorted order
    media_file_paths = [str(f) for f in supported_files]
    clip_duration = getattr(args, 'clip_duration', 3.0)  # Default 3 seconds per clip
    
    print(f"✅ Adding {len(supported_files)} media files to timeline...")
    print(f"   Each clip duration: {clip_duration}s")
    
    # Count media types
    image_count = sum(1 for f in supported_files if get_media_type_info(f)["is_image"])
    video_count = sum(1 for f in supported_files if get_media_type_info(f)["is_video"])
    print(f"   Found {video_count} videos")
    print(f"   Found {image_count} images")
    
    try:
        add_media_to_timeline(fcpxml, media_file_paths, clip_duration, use_horizontal)
        print(f"✅ Timeline created with {len(supported_files)} clips")
        
        # Calculate total duration
        total_duration = len(supported_files) * clip_duration
        print(f"   Total timeline duration: {total_duration:.1f}s")
        
    except Exception as e:
        print(f"❌ Error adding media to timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    if hasattr(args, 'output') and args.output:
        output_path = Path(args.output)
    else:
        # Default output filename based on pattern
        safe_pattern = pattern.replace("*", "X").replace("?", "Q")
        output_path = Path(f"assembled_{safe_pattern}.fcpxml")
    
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)