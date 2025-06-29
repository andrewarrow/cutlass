#!/usr/bin/env python3
"""
Many Video FX Command

Creates a tiling animation effect where videos start in the center and animate to grid positions
across the screen. Based on the pattern from Info.fcpxml with keyframe animations.
"""

import sys
import math
from pathlib import Path

from fcpxml_lib import create_empty_project, save_fcpxml
from fcpxml_lib.core.fcpxml import create_media_asset
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration
from fcpxml_lib.utils.media import discover_video_files
from fcpxml_lib.models.elements import KeyframeAnimation


def calculate_grid_positions(num_videos, screen_width=1080, screen_height=1920):
    """Calculate grid positions to tile videos across screen"""
    # Calculate optimal grid dimensions
    aspect_ratio = screen_width / screen_height
    cols = math.ceil(math.sqrt(num_videos * aspect_ratio))
    rows = math.ceil(num_videos / cols)
    
    # Calculate spacing between videos
    video_width = screen_width / cols
    video_height = screen_height / rows
    
    positions = []
    for i in range(num_videos):
        row = i // cols
        col = i % cols
        
        # Convert to Final Cut Pro coordinate system (center-based)
        # FCP coordinates: center is (0,0), positive X is right, positive Y is up
        x = (col - (cols - 1) / 2) * (video_width * 0.8)  # 80% spacing to avoid overlap
        y = ((rows - 1) / 2 - row) * (video_height * 0.8)  # Flip Y and scale
        
        positions.append((x, y))
    
    return positions, video_width / screen_width, video_height / screen_height


def many_video_fx_cmd(args):
    """Create tiling video animation effect"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Find all video files
    video_files = discover_video_files(input_dir)
    
    if not video_files:
        print(f"❌ No video files found in {input_dir}")
        print(f"   Supported formats: MOV, MP4, AVI, MKV, M4V")
        sys.exit(1)
    
    num_videos = len(video_files)
    print(f"🎬 Creating many-video-fx animation...")
    print(f"   Input directory: {input_dir}")
    print(f"   Video files found: {num_videos}")
    print(f"   Animation duration: {args.duration}s")
    
    # Calculate grid positions and scaling
    positions, scale_x, scale_y = calculate_grid_positions(num_videos)
    
    print(f"   Grid layout: {math.ceil(math.sqrt(num_videos * (1080/1920)))} cols × {math.ceil(num_videos / math.ceil(math.sqrt(num_videos * (1080/1920))))} rows")
    print(f"   Video scale: {scale_x:.3f} × {scale_y:.3f}")
    
    # Create empty project (vertical format for 1080x1920)
    fcpxml = create_empty_project(
        project_name="Many Video FX",
        event_name="Video Tiling Animation",
        use_horizontal=False  # 1080x1920 vertical format
    )
    
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    resource_counter = len(fcpxml.resources.assets) + len(fcpxml.resources.formats) + 1
    
    # Animation timing constants (based on Info.fcpxml)
    animation_duration = 6.0  # seconds for each video to reach final position
    stagger_delay = 1.5  # seconds between each video starting
    
    total_duration = max(args.duration, (num_videos - 1) * stagger_delay + animation_duration)
    
    print(f"   Animation parameters:")
    print(f"     Move duration: {animation_duration}s per video")
    print(f"     Stagger delay: {stagger_delay}s between videos")
    print(f"     Total timeline: {total_duration}s")
    
    # Create video elements with keyframe animations
    for i, video_file in enumerate(video_files):
        print(f"   Processing video {i+1}/{num_videos}: {video_file.name}")
        
        # Create video asset
        asset_id = f"r{resource_counter}"
        format_id = f"r{resource_counter + 1}"
        resource_counter += 2
        
        asset, format_obj = create_media_asset(str(video_file), asset_id, format_id, total_duration)
        fcpxml.resources.assets.append(asset)
        fcpxml.resources.formats.append(format_obj)
        
        # Calculate timing for this video
        start_time = i * stagger_delay
        animation_end_time = start_time + animation_duration
        
        # Get final position for this video
        final_x, final_y = positions[i]
        
        # Create keyframe animations for position and scale
        position_animation = KeyframeAnimation()
        position_animation.add_keyframe("0s", "0 0")  # Start at center
        position_animation.add_keyframe(convert_seconds_to_fcp_duration(animation_duration), 
                                       f"{final_x:.6f} {final_y:.6f}")  # Move to final position
        
        scale_animation = KeyframeAnimation()
        scale_animation.add_keyframe(convert_seconds_to_fcp_duration(animation_duration), 
                                    f"{scale_x:.6f} {scale_y:.6f}", "linear")  # Scale to final size
        
        # Create asset-clip element with keyframe animation
        clip_element = {
            "type": "asset-clip",
            "ref": asset_id,
            "duration": convert_seconds_to_fcp_duration(total_duration - start_time),
            "offset": convert_seconds_to_fcp_duration(start_time),
            "name": f"{video_file.stem}_tile_{i+1}",
            "adjust_transform": {
                "animations": {
                    "position_animation": position_animation,
                    "scale_animation": scale_animation
                }
            }
        }
        
        # Add to spine
        sequence.spine.asset_clips.append(clip_element)
        sequence.spine.ordered_elements.append(clip_element)
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(total_duration)
    
    print(f"✅ Created timeline with {num_videos} animated video tiles")
    print(f"   Each video starts at center and moves to its grid position")
    print(f"   Final layout tiles the entire 1080×1920 screen")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "many_video_fx.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)