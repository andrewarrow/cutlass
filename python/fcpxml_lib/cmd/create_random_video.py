#!/usr/bin/env python3
"""
Create Random Video Command

Creates a random video from media files in a directory.
Following comprehensive crash prevention rules and NO_XML_TEMPLATES principle.
"""

import sys
import random
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml, add_media_to_timeline
)
from fcpxml_lib.utils.media import discover_all_media_files


def create_random_video_cmd(args):
    """Create a random video from media files in a directory"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Find all media files using utility function
    image_files, video_files = discover_all_media_files(input_dir)
    media_files = image_files + video_files
    
    if not media_files:
        print(f"❌ No media files found in {input_dir}")
        print(f"   Supported formats: PNG, JPG, JPEG, MOV, MP4, AVI, MKV, M4V")
        sys.exit(1)
    
    # Randomly shuffle the files
    random.shuffle(media_files)
    
    format_desc = "1280x720 horizontal" if args.horizontal else "1080x1920 vertical"
    print(f"🎬 Creating random video from {len(media_files)} media files...")
    print(f"   Input directory: {input_dir}")
    print(f"   Format: {format_desc}")
    print(f"   Files found: {[f.name for f in media_files[:5]]}{'...' if len(media_files) > 5 else ''}")
    
    # Create empty project with format choice
    fcpxml = create_empty_project(
        project_name=args.project_name or f"Random Video - {input_dir.name}",
        event_name=args.event_name or "Random Videos",
        use_horizontal=args.horizontal
    )
    
    # Add media files to timeline
    media_file_paths = [str(f) for f in media_files]
    clip_duration = args.clip_duration
    
    print(f"✅ Adding {len(media_files)} media files to timeline...")
    print(f"   Each clip duration: {clip_duration}s")
    print(f"   Found {len(video_files)} videos")
    print(f"   Found {len(image_files)} images")
    
    try:
        add_media_to_timeline(fcpxml, media_file_paths, clip_duration, args.horizontal)
        print(f"✅ Timeline created with {len(media_files)} clips")
        
        # Calculate total duration
        total_duration = len(media_files) * clip_duration
        print(f"   Total timeline duration: {total_duration:.1f}s")
        
    except Exception as e:
        print(f"❌ Error adding media to timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "random_video.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)