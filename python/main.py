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
from fcpxml_lib.generators.timeline_generators import create_edge_tiled_timeline, create_stress_test_timeline
from fcpxml_lib.utils.media import discover_all_media_files, discover_image_files


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
    output_path = Path(args.output) if args.output else Path(__file__).parent / "random_video.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


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
    output_path = Path(args.output) if args.output else Path(__file__).parent / "video_at_edge.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)



def stress_test_cmd(args):
    """Create an extremely complex 9-minute stress test video to validate library robustness"""
    print("🔥 Creating FCPXML STRESS TEST - 9 minutes of maximum complexity...")
    print("   Testing all library features, validation, and edge cases")
    print("   Format: 1080x1920 vertical")
    print("   Duration: 9 minutes (540 seconds)")
    print("   Goal: Generate invalid FCPXML or prove validation system integrity")
    print()
    
    # Get all available assets
    assets_dir = Path(__file__).parent.parent / "assets"
    if not assets_dir.exists():
        print(f"❌ Assets directory not found: {assets_dir}")
        sys.exit(1)
    
    # Find all media files using utility function
    image_files, video_files = discover_all_media_files(assets_dir)
    
    print(f"   Found {len(image_files)} images: {[f.name for f in image_files]}")
    print(f"   Found {len(video_files)} videos: {[f.name for f in video_files]}")
    
    if not image_files and not video_files:
        print(f"❌ No media files found in {assets_dir}")
        sys.exit(1)
    
    # Create base project
    fcpxml = create_empty_project(
        project_name="FCPXML Stress Test - Maximum Complexity",
        event_name="Stress Test Validation",
        use_horizontal=False  # Always vertical 1080x1920 as requested
    )
    
    try:
        create_stress_test_timeline(fcpxml, image_files, video_files)
        print("✅ Stress test timeline created successfully")
    except Exception as e:
        print(f"❌ Error creating stress test timeline: {e}")
        print("   This indicates a potential library issue or validation gap")
        raise
    
    # Save with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent / "stress_test.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Stress test saved to: {output_path}")
        print("🎯 VALIDATION SUCCESS: Library generated valid FCPXML under extreme conditions")
        print("   Next steps:")
        print("   1. Validate XML with: xmllint --noout stress_test.fcpxml")
        print("   2. Import into Final Cut Pro to test import stability")
        print("   3. Play timeline to test performance with complex content")
    else:
        print("❌ VALIDATION FAILED: Library rejected its own output")
        print("   This indicates validation system is working correctly")
        sys.exit(1)




def main():
    """CLI entry point with command options"""
    parser = argparse.ArgumentParser(
        description="FCPXML Python Generator - Create Final Cut Pro projects",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s create-empty-project --output my_project.fcpxml
  %(prog)s create-random-video /path/to/media/folder --output random.fcpxml
  %(prog)s video-at-edge /path/to/image/folder --output edge_video.fcpxml --background-video bg.mp4
  %(prog)s stress-test --output stress_test.fcpxml
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
        help='Create video with random images (PNG/JPG) tiled across visible area on multiple lanes'
    )
    edge_parser.add_argument('input_dir', help='Directory containing image files (PNG, JPG, JPEG)')
    edge_parser.add_argument('--output', help='Output FCPXML file path')
    edge_parser.add_argument('--background-video', help='Background video file (optional)')
    edge_parser.add_argument('--duration', type=float, default=10.0, help='Duration in seconds (default: 10.0)')
    edge_parser.add_argument('--tiles-per-lane', type=int, default=8, help='Number of image tiles per lane (default: 8)')
    edge_parser.add_argument('--num-lanes', type=int, default=10, help='Number of lanes with image tiles (default: 10)')
    
    # Stress test command
    stress_parser = subparsers.add_parser(
        'stress-test',
        help='Create an extremely complex 9-minute stress test video to validate library robustness'
    )
    stress_parser.add_argument('--output', help='Output FCPXML file path (default: stress_test.fcpxml)')
    
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
    elif args.command == 'stress-test':
        stress_test_cmd(args)


if __name__ == "__main__":
    main()