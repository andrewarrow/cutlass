"""
Many Video FX Command Implementation

Creates FCPXML with dynamic tiled video animation effect. Videos start in center and animate to grid positions.
Based on the successful animation.py pattern but extended to handle any number of videos.
Uses proper keyframe animations and nested clip structure for multi-lane visibility.
"""

import sys
import math
from pathlib import Path

from fcpxml_lib.core.fcpxml import create_empty_project, save_fcpxml, create_media_asset, detect_video_properties
from fcpxml_lib.models.elements import (
    Clip, Video, AdjustTransform, KeyframeAnimation, Keyframe, Param
)
from fcpxml_lib.utils.ids import generate_resource_id, set_resource_id_counter
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration


def calculate_screen_filling_grid(available_videos, video_scale=0.25):
    """
    Calculate how many videos can fit on screen with tight packing and no gaps.
    
    Screen bounds: X(-30 to +30), Y(-50 to +50)
    Video scale: determines the size of each video tile
    
    Returns: (total_videos_needed, cols, rows, positions, video_list)
    """
    # Screen dimensions
    screen_width = 60.0   # -30 to +30
    screen_height = 100.0  # -50 to +50
    
    # Calculate video tile dimensions based on scale
    # Assume original video is roughly 16:9 aspect ratio
    # In FCP coordinate space, a scale of 0.25 means the video takes up 
    # roughly 25% of the screen width/height
    tile_width = screen_width * video_scale   # ~15 units wide
    tile_height = screen_height * video_scale * 0.6  # ~15 units tall (adjusted for aspect ratio)
    
    # Calculate how many videos fit horizontally and vertically
    cols = max(1, int(screen_width / tile_width))
    rows = max(1, int(screen_height / tile_height))
    
    total_videos_needed = cols * rows
    
    print(f"   📐 Calculated grid: {cols} cols × {rows} rows = {total_videos_needed} total positions")
    print(f"   📏 Tile size: {tile_width:.1f} × {tile_height:.1f} units")
    
    # Create video list by repeating available videos to fill all positions
    video_list = []
    for i in range(total_videos_needed):
        video_index = i % len(available_videos)
        video_list.append(available_videos[video_index])
    
    # Calculate tight-packed positions with no gaps
    positions = []
    
    # Calculate starting position (top-left corner)
    start_x = -screen_width / 2 + (tile_width / 2)
    start_y = -screen_height / 2 + (tile_height / 2)
    
    for i in range(total_videos_needed):
        row = i // cols
        col = i % cols
        
        x = start_x + (col * tile_width)
        y = start_y + (row * tile_height)
        
        positions.append((x, y))
    
    return total_videos_needed, cols, rows, positions, video_list


def calculate_tile_positions(num_videos):
    """
    Legacy function for backward compatibility.
    Now delegates to calculate_screen_filling_grid.
    """
    # This will be replaced by the new screen-filling logic in the main function
    return []


def many_video_fx_cmd(args):
    """CLI implementation for many-video-fx command"""
    
    # Get input directory from args
    input_dir = Path(args.input_dir)
    
    if not input_dir.exists():
        print(f"❌ Directory not found: {input_dir}", file=sys.stderr)
        sys.exit(1)
    
    if not input_dir.is_dir():
        print(f"❌ Path is not a directory: {input_dir}", file=sys.stderr)
        sys.exit(1)
    
    # Find MOV files in directory
    mov_files = list(input_dir.glob("*.mov"))
    if len(mov_files) < 1:
        print(f"❌ Directory must contain at least 1 MOV file, found {len(mov_files)}", file=sys.stderr)
        sys.exit(1)
    
    # Sort files for consistent ordering
    available_videos = sorted(mov_files)
    
    print(f"📁 Found {len(available_videos)} videos in {input_dir.name}")
    
    # Calculate screen-filling grid (will repeat videos to fill entire screen)
    total_videos_needed, cols, rows, tile_positions, selected_videos = calculate_screen_filling_grid(available_videos)
    num_videos = total_videos_needed
    
    print(f"🎯 Screen-filling grid: {cols} columns × {rows} rows = {num_videos} total videos")
    print(f"🔄 Repeating {len(available_videos)} source videos to fill {num_videos} positions")
    
    # Create base project (vertical format like animation command)
    fcpxml = create_empty_project(use_horizontal=False)
    
    # Set ID counter to start from r2 since r1 is already used by project format
    set_resource_id_counter(1)
    
    # Generate resource IDs for media assets - each video gets its own format
    asset_ids = []
    format_ids = []
    for i in range(num_videos):
        asset_ids.append(generate_resource_id())
        format_ids.append(generate_resource_id())
    
    # Create media assets for all videos
    try:
        assets = []
        formats = []
        video_properties = []
        
        for i, video_path in enumerate(selected_videos):
            asset, format_obj = create_media_asset(
                str(video_path), asset_ids[i], format_ids[i]
            )
            props = detect_video_properties(str(video_path))
            
            assets.append(asset)
            formats.append(format_obj)
            video_properties.append(props)
        
        fcpxml.resources.assets.extend(assets)
        fcpxml.resources.formats.extend(formats)
        
    except Exception as e:
        print(f"❌ Failed to process video files: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Create timeline sequence
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format
    
    # Animation timing constants (based on animation.py successful pattern)
    base_animation_duration = 6.0   # 6 seconds animation time
    stagger_delay = 1.5            # 1.5 seconds between video starts
    
    # Calculate durations to ensure videos keep playing after reaching final positions
    max_video_duration = max(props['duration_seconds'] for props in video_properties)
    last_video_start_time = (num_videos - 1) * stagger_delay  # When last video starts animating
    animation_end_time = last_video_start_time + base_animation_duration  # When last animation ends
    
    # Timeline should continue for a while after all animations complete
    post_animation_duration = 10.0  # Keep playing 10 seconds after animations end
    total_timeline_duration = animation_end_time + post_animation_duration
    
    # Each video needs to play long enough to cover its animation + post duration
    min_video_duration_needed = base_animation_duration + post_animation_duration
    
    sequence.duration = convert_seconds_to_fcp_duration(total_timeline_duration)
    
    # Use consistent scale for tight packing (matches calculate_screen_filling_grid)
    base_scale = 0.25
    scale_values = [f"{base_scale} {base_scale}"] * num_videos
    # Make first video flipped like animation pattern
    scale_values[0] = f"-{base_scale} {base_scale}"
    
    # Create keyframe animations for each video
    transforms = []
    
    for i in range(num_videos):
        final_x, final_y = tile_positions[i]
        animation_duration_fcp = convert_seconds_to_fcp_duration(base_animation_duration)
        scale_value = scale_values[i] if i < len(scale_values) else scale_values[-1]
        
        transform = AdjustTransform(
            params=[
                Param(
                    name="anchor",
                    keyframe_animation=KeyframeAnimation(keyframes=[
                        Keyframe(time=animation_duration_fcp, value="0 0", curve="linear")
                    ])
                ),
                Param(
                    name="position", 
                    keyframe_animation=KeyframeAnimation(keyframes=[
                        Keyframe(time="0s", value="0 0"),
                        Keyframe(time=animation_duration_fcp, value=f"{final_x:.4f} {final_y:.4f}")
                    ])
                ),
                Param(
                    name="rotation",
                    keyframe_animation=KeyframeAnimation(keyframes=[
                        Keyframe(time=animation_duration_fcp, value="0", curve="linear")
                    ])
                ),
                Param(
                    name="scale",
                    keyframe_animation=KeyframeAnimation(keyframes=[
                        Keyframe(time=animation_duration_fcp, value=scale_value, curve="linear")
                    ])
                )
            ]
        )
        transforms.append(transform)
    
    # Create main clip using first video as container (like animation.py)
    # Main clip duration should cover the entire timeline
    main_clip_duration = convert_seconds_to_fcp_duration(total_timeline_duration)
    
    # Each video needs to play long enough to stay visible after animation
    # Use the longer of: original video duration or minimum needed duration
    def get_video_duration(video_props, needed_duration):
        original_duration = video_props['duration_seconds']
        return max(original_duration, needed_duration)
    
    main_video_duration = convert_seconds_to_fcp_duration(
        get_video_duration(video_properties[0], min_video_duration_needed)
    )
    
    main_clip = Clip(
        offset="0s",
        name=f"Many Video FX - {num_videos} videos",
        duration=main_clip_duration,
        format=format_ids[0],
        tc_format="NDF"
    )
    
    # Create main clip's video element
    main_video = Video(
        ref=asset_ids[0],
        offset="0s", 
        duration=main_video_duration
    )
    
    # Create nested clips for remaining videos (if any)
    nested_clips = []
    
    for i in range(1, num_videos):
        video_offset = convert_seconds_to_fcp_duration(i * stagger_delay)
        
        # Each nested video also needs to play long enough to stay visible
        video_duration = convert_seconds_to_fcp_duration(
            get_video_duration(video_properties[i], min_video_duration_needed)
        )
        
        nested_clip_info = {
            "lane": i,
            "offset": video_offset,
            "name": selected_videos[i].stem, 
            "duration": video_duration,
            "ref": asset_ids[i],
            "video_duration": video_duration,
            "transform": transforms[i]
        }
        nested_clips.append(nested_clip_info)
    
    # Create nested clip objects
    nested_clip_objs = []
    for i, clip_info in enumerate(nested_clips):
        nested_clip = Clip(
            lane=clip_info["lane"],
            offset=clip_info["offset"],
            name=clip_info["name"],
            duration=clip_info["duration"],
            format=format_ids[i+1],
            tc_format="NDF"
        )
        nested_clip.adjust_transform = clip_info["transform"]
        
        # Add video element
        nested_video = Video(
            ref=clip_info["ref"],
            offset="0s",
            duration=clip_info["video_duration"]
        )
        nested_clip.videos = [nested_video]
        nested_clip_objs.append(nested_clip)
    
    # Set main clip's transform and video
    main_clip.adjust_transform = transforms[0]
    main_clip.videos = [main_video]
    main_clip.clips = nested_clip_objs
    
    # Convert to dictionary format exactly like animation.py
    main_clip_dict = {
        "type": "clip",
        "offset": main_clip.offset,
        "name": main_clip.name,
        "duration": main_clip.duration,
        "format": main_clip.format,
        "tcFormat": main_clip.tc_format,
        "nested_elements": []
    }
    
    # Add adjust-transform as nested element
    transform_dict = main_clip.adjust_transform.to_dict()
    transform_dict["type"] = "adjust_transform"
    main_clip_dict["nested_elements"].append(transform_dict)
    
    # Add video element to main clip nested_elements
    main_clip_dict["nested_elements"].append({
        "type": "video",
        "ref": main_video.ref,
        "offset": main_video.offset,
        "duration": main_video.duration
    })
    
    # Add nested clips to main clip
    for nested_clip in main_clip.clips:
        nested_dict = {
            "type": "clip",
            "lane": nested_clip.lane,
            "offset": nested_clip.offset,
            "name": nested_clip.name,
            "duration": nested_clip.duration,
            "format": nested_clip.format,
            "tcFormat": nested_clip.tc_format,
            "nested_elements": []
        }
        
        # Add adjust-transform as nested element
        nested_transform_dict = nested_clip.adjust_transform.to_dict()
        nested_transform_dict["type"] = "adjust_transform"
        nested_dict["nested_elements"].append(nested_transform_dict)
        
        # Add video elements to nested clip's nested_elements
        for video in nested_clip.videos:
            nested_dict["nested_elements"].append({
                "type": "video",
                "ref": video.ref,
                "offset": video.offset,
                "duration": video.duration
            })
        main_clip_dict["nested_elements"].append(nested_dict)
    
    # Add to spine
    sequence.spine.ordered_elements = [main_clip_dict]
    
    # Save FCPXML
    output_path = args.output if args.output else "many_video_fx.fcpxml"
    try:
        success = save_fcpxml(fcpxml, output_path)
        if not success:
            print(f"❌ Failed to save FCPXML to {output_path}", file=sys.stderr)
            sys.exit(1)
            
        print(f"✅ Many Video FX FCPXML created: {output_path}")
        print(f"   🎬 {num_videos} videos in {cols}×{rows} grid")
        print(f"   🎭 Each video animates from center to tile position")
        print(f"   ⏱️  Animation: {base_animation_duration}s per video")
        print(f"   📏 Stagger delay: {stagger_delay}s between starts")
        print(f"   🎯 Screen bounds: X(-30 to +30), Y(-50 to +50)")
        print(f"   ⏱️  Total timeline: {total_timeline_duration:.1f}s")
        print(f"   🎞️  Videos play {post_animation_duration}s after animations end")
        
    except Exception as e:
        print(f"❌ Error saving FCPXML: {e}", file=sys.stderr)
        sys.exit(1)