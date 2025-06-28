#!/usr/bin/env python3
"""
FCPXML Python Library - Demo Application
Generates FCPXML documents following comprehensive crash prevention rules

Based on the Go and Swift implementations, this library ensures:
- Frame-aligned timing calculations  
- Proper media type handling
- Validated resource ID management
- Crash pattern prevention

🚨 CRITICAL: This follows the "NO_XML_TEMPLATES" rule
All XML is generated from structured data objects, never string templates.
"""

import sys
import argparse
import random
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml, add_media_to_timeline, ValidationError,
    Sequence
)
from fcpxml_lib.core.fcpxml import create_media_asset
from fcpxml_lib.constants import (
    SCREEN_EDGE_LEFT, SCREEN_EDGE_RIGHT, SCREEN_EDGE_TOP, SCREEN_EDGE_BOTTOM,
    SCREEN_WIDTH, SCREEN_HEIGHT
)
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration


def test_validation_failure():
    """Demonstrate validation failure detection"""
    print("\n🧪 Testing validation failure detection...")
    
    # Try to create a sequence with invalid audio rate
    try:
        bad_sequence = Sequence(
            format="r1",
            audio_rate="48000"  # Invalid - should be "48k"
        )
        print("❌ SHOULD HAVE FAILED - invalid audio rate was not caught!")
    except ValidationError as e:
        print(f"✅ Validation correctly caught error: {e}")


def create_empty_project_cmd(args):
    """Create an empty FCPXML project"""
    print("🎬 Creating empty FCPXML project...")
    print("Following crash prevention rules for safe FCPXML generation")
    print()
    
    # Test validation system first
    test_validation_failure()
    
    # Create empty project with format choice
    format_desc = "1280x720 horizontal" if args.horizontal else "1080x1920 vertical"
    print(f"   Format: {format_desc}")
    
    fcpxml = create_empty_project(
        project_name=args.project_name or "My First Project",
        event_name=args.event_name or "My First Event",
        use_horizontal=args.horizontal
    )
    
    # Validate the project
    print("✅ FCPXML structure created and validated")
    print(f"   Version: {fcpxml.version}")
    print(f"   Resources: {len(fcpxml.resources.formats)} formats")
    print(f"   Events: {len(fcpxml.library.events)}")
    print(f"   Projects: {len(fcpxml.library.events[0].projects)}")
    print()
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent / "empty_project.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
        print("🎯 Next steps:")
        print("1. Import into Final Cut Pro to test")
        print("2. Extend this library to add media assets") 
        print("3. Implement more spine elements (asset-clips, titles, etc.)")
        print("4. Add keyframe animation support")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


def create_random_video_cmd(args):
    """Create a random video from media files in a directory"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Supported media extensions
    video_extensions = {'.mov', '.mp4', '.avi', '.mkv', '.m4v'}
    image_extensions = {'.jpg', '.jpeg', '.png', '.tiff', '.bmp', '.gif'}
    all_extensions = video_extensions | image_extensions
    
    # Find all media files
    media_files = []
    for ext in all_extensions:
        media_files.extend(input_dir.glob(f"*{ext}"))
        media_files.extend(input_dir.glob(f"*{ext.upper()}"))
    
    if not media_files:
        print(f"❌ No media files found in {input_dir}")
        print(f"   Supported formats: {', '.join(sorted(all_extensions))}")
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
    print(f"   Found {len([f for f in media_files if f.suffix.lower() in video_extensions])} videos")
    print(f"   Found {len([f for f in media_files if f.suffix.lower() in image_extensions])} images")
    
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
    output_path = Path(args.output) if args.output else Path(__file__).parent / "random_video.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


def video_at_edge_cmd(args):
    """Create video with random PNGs tiled across visible area on multiple lanes"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Find all PNG files
    png_files = []
    for pattern in ['*.png', '*.PNG']:
        png_files.extend(input_dir.glob(pattern))
    
    if not png_files:
        print(f"❌ No PNG files found in {input_dir}")
        sys.exit(1)
    
    print(f"🎨 Creating video with edge-tiled PNGs...")
    print(f"   Input directory: {input_dir}")
    print(f"   PNG files found: {len(png_files)}")
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
            png_files, 
            args.background_video, 
            args.duration,
            args.num_lanes,
            args.tiles_per_lane
        )
        
        total_tiles = args.num_lanes * args.tiles_per_lane
        print(f"✅ Timeline created with {total_tiles} PNG tiles across {args.num_lanes} lanes")
        
    except Exception as e:
        print(f"❌ Error creating edge-tiled timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent / "video_at_edge.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


def create_edge_tiled_timeline(fcpxml, png_files, background_video, duration, num_lanes, tiles_per_lane):
    """
    Create timeline with PNGs tiled across the visible screen area using proper lane structure.
    
    🚨 CRITICAL: Uses separate lane elements instead of nested structure to prevent FCP crashes.
    
    According to BAFFLE_TWO.md, deeply nested video elements cause "Invalid edit with no respective media" errors.
    This implementation uses the proper lane system where each element is on its own lane.
    """
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    resource_counter = len(fcpxml.resources.assets) + len(fcpxml.resources.formats) + 1
    
    # Create shared format definitions to avoid redundancy
    image_format_id = f"r{resource_counter}"
    resource_counter += 1
    
    # Create shared image format (reused for all PNG tiles)
    from fcpxml_lib.models.elements import Format
    from fcpxml_lib.constants import IMAGE_FORMAT_NAME, DEFAULT_IMAGE_WIDTH, DEFAULT_IMAGE_HEIGHT, IMAGE_COLOR_SPACE
    
    shared_image_format = Format(
        id=image_format_id,
        name=IMAGE_FORMAT_NAME,
        width=DEFAULT_IMAGE_WIDTH,
        height=DEFAULT_IMAGE_HEIGHT,
        color_space=IMAGE_COLOR_SPACE
    )
    fcpxml.resources.formats.append(shared_image_format)
    
    # Create the main background element (if background video provided)
    if background_video:
        bg_path = Path(background_video)
        if bg_path.exists():
            print(f"   Adding background video: {bg_path.name}")
            
            # Create background asset
            asset_id = f"r{resource_counter}"
            format_id = f"r{resource_counter + 1}"
            resource_counter += 2
            
            asset, format_obj = create_media_asset(str(bg_path), asset_id, format_id, duration)
            fcpxml.resources.assets.append(asset)
            fcpxml.resources.formats.append(format_obj)
            
            # Determine if background video needs scaling
            from fcpxml_lib.core.fcpxml import needs_vertical_scaling
            is_video = bg_path.suffix.lower() in {'.mov', '.mp4', '.avi', '.mkv', '.m4v'}
            needs_scaling = needs_vertical_scaling(str(bg_path), is_image=not is_video)
            
            # Create background element (use appropriate type based on media)
            bg_duration = convert_seconds_to_fcp_duration(duration)
            
            if is_video:
                # Background is a video - use asset-clip
                bg_element = {
                    "type": "asset-clip",
                    "ref": asset_id,
                    "duration": bg_duration,
                    "offset": "0s",
                    "name": bg_path.stem
                }
            else:
                # Background is an image - use video element
                bg_element = {
                    "type": "video",
                    "ref": asset_id,
                    "duration": bg_duration,
                    "offset": "0s",
                    "start": "0s",  # Required for image video elements
                    "name": bg_path.stem
                }
            
            # Add scaling if needed
            if needs_scaling:
                from fcpxml_lib.constants import VERTICAL_SCALE_FACTOR
                bg_element["adjust_transform"] = {"scale": VERTICAL_SCALE_FACTOR}
                print(f"   Background video scaled for vertical format")
            
            # Add background to spine
            sequence.spine.ordered_elements.append(bg_element)
            if is_video:
                sequence.spine.asset_clips.append(bg_element)
            else:
                sequence.spine.videos.append(bg_element)
    
    # Generate PNG tiles as nested lane elements (like safe.fcpxml structure)
    if background_video:
        # Find the background element to add nested lanes
        bg_element = sequence.spine.ordered_elements[-1]  # Last added element (background)
        
        # Add nested_elements list if not exists
        if "nested_elements" not in bg_element:
            bg_element["nested_elements"] = []
        
        # Generate tiles with each PNG on its own lane
        current_lane = 1
        total_tiles = num_lanes * tiles_per_lane
        
        for tile_index in range(total_tiles):
            # Select random PNG
            png_file = random.choice(png_files)
            
            # Create asset using shared image format
            asset_id = f"r{resource_counter}"
            resource_counter += 1
            
            # Create asset manually to use shared format
            from fcpxml_lib.models.elements import Asset, MediaRep
            from fcpxml_lib.utils.ids import generate_uid
            from fcpxml_lib.constants import IMAGE_DURATION
            
            abs_path = Path(png_file).resolve()
            uid = generate_uid(f"MEDIA_{abs_path.name}")
            media_rep = MediaRep(src=str(abs_path))
            
            asset = Asset(
                id=asset_id,
                name=abs_path.stem,
                uid=uid,
                duration=IMAGE_DURATION,
                has_video="1",
                format=image_format_id,  # Use shared format
                video_sources="1",
                media_rep=media_rep
            )
            fcpxml.resources.assets.append(asset)
            
            # Generate random position within screen bounds
            x_pos = random.uniform(SCREEN_EDGE_LEFT, SCREEN_EDGE_RIGHT)
            y_pos = random.uniform(SCREEN_EDGE_TOP, SCREEN_EDGE_BOTTOM)
            
            # Generate random scale (smaller tiles)
            scale = random.uniform(0.1, 0.5)  # 10% to 50% original size
            
            # Create PNG tile element as nested lane element (like safe.fcpxml)
            tile_duration = convert_seconds_to_fcp_duration(duration)
            
            # Use timing values that match Info.fcpxml pattern
            # Some elements use simple "3600s", others use frame-aligned versions
            if tile_index < 2:  # First two elements use simple timing like Info.fcpxml
                tile_offset = "0s"
                tile_start = "3600s" if tile_index == 0 else "86399313/24000s"
            else:  # Remaining elements use frame-aligned timing
                tile_offset = "86400314/24000s"  # Frame-aligned version of ~3600s
                tile_start = "86486400/24000s"   # Frame-aligned start
            
            tile_element = {
                "type": "video",
                "ref": asset_id,
                "lane": current_lane,  # Each PNG gets its own lane
                "duration": tile_duration,
                "offset": tile_offset,  # Frame-aligned offset
                "start": tile_start,  # Frame-aligned start
                "name": f"{png_file.stem}_lane_{current_lane}",
                "adjust_transform": {
                    "position": f"{x_pos:.3f} {y_pos:.3f}",
                    "scale": f"{scale:.3f} {scale:.3f}"
                }
            }
            
            # Add as nested element inside background (like safe.fcpxml structure)
            bg_element["nested_elements"].append(tile_element)
            
            # Increment lane for next PNG
            current_lane += 1
    else:
        # If no background, create PNG tiles as separate spine elements
        current_lane = 1
        total_tiles = num_lanes * tiles_per_lane
        
        for tile_index in range(total_tiles):
            # Select random PNG
            png_file = random.choice(png_files)
            
            # Create asset using shared image format
            asset_id = f"r{resource_counter}"
            resource_counter += 1
            
            # Create asset manually to use shared format
            from fcpxml_lib.models.elements import Asset, MediaRep
            from fcpxml_lib.utils.ids import generate_uid
            from fcpxml_lib.constants import IMAGE_DURATION
            
            abs_path = Path(png_file).resolve()
            uid = generate_uid(f"MEDIA_{abs_path.name}")
            media_rep = MediaRep(src=str(abs_path))
            
            asset = Asset(
                id=asset_id,
                name=abs_path.stem,
                uid=uid,
                duration=IMAGE_DURATION,
                has_video="1",
                format=image_format_id,  # Use shared format
                video_sources="1",
                media_rep=media_rep
            )
            fcpxml.resources.assets.append(asset)
            
            # Generate random position within screen bounds
            x_pos = random.uniform(SCREEN_EDGE_LEFT, SCREEN_EDGE_RIGHT)
            y_pos = random.uniform(SCREEN_EDGE_TOP, SCREEN_EDGE_BOTTOM)
            
            # Generate random scale (smaller tiles)
            scale = random.uniform(0.1, 0.5)  # 10% to 50% original size
            
            # Create PNG tile element as separate spine element
            tile_duration = convert_seconds_to_fcp_duration(duration)
            tile_element = {
                "type": "video",
                "ref": asset_id,
                "lane": current_lane,  # Each PNG gets its own lane
                "duration": tile_duration,
                "offset": "0s",  # All tiles start at the same time
                "start": "0s",  # Required for image video elements
                "name": f"{png_file.stem}_lane_{current_lane}",
                "adjust_transform": {
                    "position": f"{x_pos:.3f} {y_pos:.3f}",
                    "scale": f"{scale:.3f} {scale:.3f}"
                }
            }
            
            # Add as separate element to spine
            sequence.spine.videos.append(tile_element)
            sequence.spine.ordered_elements.append(tile_element)
            
            # Increment lane for next PNG
            current_lane += 1
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(duration)
    
    total_tiles = num_lanes * tiles_per_lane
    if background_video:
        print(f"   Generated {total_tiles} random PNG tiles, each on its own lane (lanes 1-{total_tiles})")
        print(f"   Structure: Background video + {total_tiles} nested lane elements")
    else:
        print(f"   Generated {total_tiles} random PNG tiles, each on its own lane (lanes 1-{total_tiles})")
        print(f"   Structure: {total_tiles} separate lane elements")
    print(f"   Screen bounds: X({SCREEN_EDGE_LEFT:.1f} to {SCREEN_EDGE_RIGHT:.1f}), Y({SCREEN_EDGE_TOP:.1f} to {SCREEN_EDGE_BOTTOM:.1f})")
    print(f"   Original request: {num_lanes} lanes × {tiles_per_lane} tiles = {total_tiles} total PNG lanes")


def main():
    """CLI entry point with command options"""
    parser = argparse.ArgumentParser(
        description="FCPXML Python Generator - Create Final Cut Pro projects",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s create-empty-project --output my_project.fcpxml
  %(prog)s create-random-video /path/to/media/folder --output random.fcpxml
  %(prog)s video-at-edge /path/to/png/folder --output edge_video.fcpxml --background-video bg.mp4
        """
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Create empty project command
    empty_parser = subparsers.add_parser(
        'create-empty-project',
        help='Create an empty FCPXML project'
    )
    empty_parser.add_argument('--project-name', help='Name of the project')
    empty_parser.add_argument('--event-name', help='Name of the event')
    empty_parser.add_argument('--output', help='Output FCPXML file path')
    empty_parser.add_argument('--horizontal', action='store_true', help='Use 1280x720 horizontal format instead of default 1080x1920 vertical')
    
    # Create random video command
    random_parser = subparsers.add_parser(
        'create-random-video',
        help='Create a random video from media files in a directory'
    )
    random_parser.add_argument('input_dir', help='Directory containing media files')
    random_parser.add_argument('--project-name', help='Name of the project')
    random_parser.add_argument('--event-name', help='Name of the event')
    random_parser.add_argument('--output', help='Output FCPXML file path')
    random_parser.add_argument('--clip-duration', type=float, default=5.0, help='Duration in seconds for each clip (default: 5.0)')
    random_parser.add_argument('--horizontal', action='store_true', help='Use 1280x720 horizontal format instead of default 1080x1920 vertical')
    
    # Video at edge command
    edge_parser = subparsers.add_parser(
        'video-at-edge',
        help='Create video with random PNGs tiled across visible area on multiple lanes'
    )
    edge_parser.add_argument('input_dir', help='Directory containing PNG files')
    edge_parser.add_argument('--output', help='Output FCPXML file path')
    edge_parser.add_argument('--background-video', help='Background video file (optional)')
    edge_parser.add_argument('--duration', type=float, default=10.0, help='Duration in seconds (default: 10.0)')
    edge_parser.add_argument('--tiles-per-lane', type=int, default=8, help='Number of PNG tiles per lane (default: 8)')
    edge_parser.add_argument('--num-lanes', type=int, default=10, help='Number of lanes with PNG tiles (default: 10)')
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    if args.command == 'create-empty-project':
        create_empty_project_cmd(args)
    elif args.command == 'create-random-video':
        create_random_video_cmd(args)
    elif args.command == 'video-at-edge':
        video_at_edge_cmd(args)


if __name__ == "__main__":
    main()