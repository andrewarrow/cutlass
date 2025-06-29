#!/usr/bin/env python3
"""
FCPXML Python Library - CLI Entry Point

Main entry point for the FCPXML Python CLI tool.
This file handles only argument parsing and command dispatching.
All command implementations are in the fcpxml_lib.cmd package.

🚨 CRITICAL: Keep this file minimal - CLI structure only
All business logic must be in fcpxml_lib modules.
Each command is implemented in a separate file in fcpxml_lib.cmd/
"""

import sys
import argparse

from fcpxml_lib.cmd import (
    create_empty_project_cmd,
    create_random_video_cmd,
    video_at_edge_cmd,
    stress_test_cmd,
    random_font_cmd,
    animation_cmd,
    many_video_fx_cmd,
    many_video_fx2_cmd
)


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
  %(prog)s random-font --output random_font.fcpxml
  %(prog)s animation video1.mp4 video2.mp4 --output animated.fcpxml
  %(prog)s many-video-fx /path/to/video/folder --output tiled_videos.fcpxml
  %(prog)s many-video-fx2 /path/to/video/folder --output fx2_videos.fcpxml --steps 2
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
    
    # Random font command
    font_parser = subparsers.add_parser(
        'random-font',
        help='Create 9-minute 1080x1920 video with random font title elements'
    )
    font_parser.add_argument('--project-name', help='Name of the project')
    font_parser.add_argument('--event-name', help='Name of the event')
    font_parser.add_argument('--output', help='Output FCPXML file path')
    
    # Animation command
    animation_parser = subparsers.add_parser(
        'animation',
        help='Create keyframe animated video (Info.fcpxml pattern - 2 videos animating from center to corners)'
    )
    animation_parser.add_argument('input_files', nargs=2, help='Two video files (.mp4 or .mov)')
    animation_parser.add_argument('--output', dest='output_path', required=True, help='Output FCPXML file path')
    
    # Many video FX command
    many_fx_parser = subparsers.add_parser(
        'many-video-fx',
        help='Create tiled video animation effect where videos start in center and animate to tile positions'
    )
    many_fx_parser.add_argument('input_dir', help='Directory containing .mov video files')
    many_fx_parser.add_argument('--output', help='Output FCPXML file path')
    many_fx_parser.add_argument('--duration', type=float, default=60.0, help='Total timeline duration in seconds (default: 60.0)')
    many_fx_parser.add_argument('--steps', type=int, help='Number of videos to animate step-by-step (2=Info.fcpxml pattern, 4=continue with more). If omitted, tiles all videos simultaneously.')
    
    # Many video FX 2 command (copy of animation.py)
    many_fx2_parser = subparsers.add_parser(
        'many-video-fx2',
        help='Create step-by-step video animation (copy of animation.py logic) with --steps parameter'
    )
    many_fx2_parser.add_argument('input_dir', help='Directory containing .mov video files')
    many_fx2_parser.add_argument('--output', required=True, help='Output FCPXML file path')
    many_fx2_parser.add_argument('--steps', type=int, default=2, help='Number of videos to animate step-by-step (default: 2)')
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    # Dispatch to appropriate command handler
    if args.command == 'create-empty-project':
        create_empty_project_cmd(args)
    elif args.command == 'create-random-video':
        create_random_video_cmd(args)
    elif args.command == 'video-at-edge':
        video_at_edge_cmd(args)
    elif args.command == 'stress-test':
        stress_test_cmd(args)
    elif args.command == 'random-font':
        random_font_cmd(args)
    elif args.command == 'animation':
        animation_cmd(args)
    elif args.command == 'many-video-fx':
        many_video_fx_cmd(args)
    elif args.command == 'many-video-fx2':
        many_video_fx2_cmd(args)
    else:
        print(f"❌ Unknown command: {args.command}", file=sys.stderr)
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()