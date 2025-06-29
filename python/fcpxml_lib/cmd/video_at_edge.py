#!/usr/bin/env python3
"""
Video at Edge Command

Creates video with random images (PNG/JPG) tiled across visible area on multiple lanes.
Following comprehensive crash prevention rules and NO_XML_TEMPLATES principle.
"""

import sys
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml
)
from fcpxml_lib.generators.timeline_generators import create_edge_tiled_timeline
from fcpxml_lib.utils.media import discover_image_files


def video_at_edge_cmd(args):
    """Create video with random images (PNG/JPG) tiled across visible area on multiple lanes"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Find all image files using utility function
    image_files = discover_image_files(input_dir)
    
    if not image_files:
        print(f"❌ No image files found in {input_dir}")
        print(f"   Supported formats: PNG, JPG, JPEG")
        sys.exit(1)
    
    print(f"🎨 Creating video with edge-tiled images...")
    print(f"   Input directory: {input_dir}")
    print(f"   Image files found: {len(image_files)}")
    print(f"   Duration: {args.duration}s")
    print(f"   Lanes: {args.num_lanes}")
    print(f"   Tiles per lane: {args.tiles_per_lane}")
    
    # Create empty project (always vertical for edge detection)
    fcpxml = create_empty_project(
        project_name="Video at Edge",
        event_name="Edge Tiled Videos",
        use_horizontal=False  # Always use vertical format
    )
    
    # Generate the edge-tiled timeline
    try:
        create_edge_tiled_timeline(
            fcpxml, 
            image_files, 
            args.background_video, 
            args.duration,
            args.num_lanes,
            args.tiles_per_lane
        )
        
        total_tiles = args.num_lanes * args.tiles_per_lane
        print(f"✅ Timeline created with {total_tiles} image tiles across {args.num_lanes} lanes")
        
    except Exception as e:
        print(f"❌ Error creating edge-tiled timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "video_at_edge.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)