#!/usr/bin/env python3
"""
FCPXML Python Library - Demo Application
Generates FCPXML documents following the comprehensive rules from schema.yaml

Based on the Go and Swift implementations, this library ensures:
- Frame-aligned timing calculations  
- Proper media type handling
- Validated resource ID management
- Crash pattern prevention

🚨 CRITICAL: This follows the "NO_XML_TEMPLATES" rule from schema.yaml
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
    print("Following schema.yaml rules for safe FCPXML generation")
    print()
    
    # Test validation system first
    test_validation_failure()
    
    # Create empty project
    fcpxml = create_empty_project(
        project_name=args.project_name or "My First Project",
        event_name=args.event_name or "My First Event"
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
    
    print(f"🎬 Creating random video from {len(media_files)} media files...")
    print(f"   Input directory: {input_dir}")
    print(f"   Files found: {[f.name for f in media_files[:5]]}{'...' if len(media_files) > 5 else ''}")
    
    # Create empty project
    fcpxml = create_empty_project(
        project_name=args.project_name or f"Random Video - {input_dir.name}",
        event_name=args.event_name or "Random Videos"
    )
    
    # Add media files to timeline
    media_file_paths = [str(f) for f in media_files]
    clip_duration = args.clip_duration
    
    print(f"✅ Adding {len(media_files)} media files to timeline...")
    print(f"   Each clip duration: {clip_duration}s")
    print(f"   Found {len([f for f in media_files if f.suffix.lower() in video_extensions])} videos")
    print(f"   Found {len([f for f in media_files if f.suffix.lower() in image_extensions])} images")
    
    try:
        add_media_to_timeline(fcpxml, media_file_paths, clip_duration)
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


def main():
    """CLI entry point with command options"""
    parser = argparse.ArgumentParser(
        description="FCPXML Python Generator - Create Final Cut Pro projects",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s create-empty-project --output my_project.fcpxml
  %(prog)s create-random-video /path/to/media/folder --output random.fcpxml
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
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    if args.command == 'create-empty-project':
        create_empty_project_cmd(args)
    elif args.command == 'create-random-video':
        create_random_video_cmd(args)


if __name__ == "__main__":
    main()