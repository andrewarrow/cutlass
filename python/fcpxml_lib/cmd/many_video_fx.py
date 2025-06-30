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


def calculate_grid_layout(num_videos):
    """
    Calculate optimal grid layout for tiling videos.
    Returns (cols, rows) tuple.
    """
    if num_videos <= 4:
        # For 4 or fewer videos, use 2x2 grid like animation command
        return 2, 2
    else:
        # For more videos, calculate square-ish grid
        cols = math.ceil(math.sqrt(num_videos))
        rows = math.ceil(num_videos / cols)
        return cols, rows


def calculate_tile_positions(num_videos):
    """
    Calculate grid positions for videos within proper screen edge bounds.
    
    Based on test_video_at_edge.py screen bounds:
    - Center is at (0, 0)
    - X range: -30.0 to +30.0 (60 units total)
    - Y range: -50.0 to +50.0 (100 units total)
    
    Returns list of (x, y) positions for each video.
    """
    cols, rows = calculate_grid_layout(num_videos)
    
    # Screen edge bounds from test_video_at_edge.py
    screen_x_min, screen_x_max = -30.0, 30.0
    screen_y_min, screen_y_max = -50.0, 50.0
    
    # Calculate grid spacing to fit within screen bounds
    x_range = screen_x_max - screen_x_min  # 60 units
    y_range = screen_y_max - screen_y_min  # 100 units
    
    # Leave some margin from edges
    margin_factor = 0.9  # Use 90% of available space
    x_spacing = (x_range * margin_factor) / max(cols - 1, 1)
    y_spacing = (y_range * margin_factor) / max(rows - 1, 1)
    
    # Calculate starting position (center the grid)
    start_x = screen_x_min + (x_range * (1 - margin_factor) / 2)
    start_y = screen_y_min + (y_range * (1 - margin_factor) / 2)
    
    positions = []
    for i in range(num_videos):
        row = i // cols
        col = i % cols
        
        if cols == 1:
            # Single column - center horizontally
            x = 0.0
        else:
            x = start_x + (col * x_spacing)
        
        if rows == 1:
            # Single row - center vertically
            y = 0.0
        else:
            y = start_y + (row * y_spacing)
        
        # Ensure positions stay within bounds
        x = max(screen_x_min, min(screen_x_max, x))
        y = max(screen_y_min, min(screen_y_max, y))
        
        positions.append((x, y))
    
    return positions


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
    selected_videos = sorted(mov_files)
    num_videos = len(selected_videos)
    
    print(f"📁 Processing {num_videos} videos from {input_dir.name}")
    
    # Calculate grid layout
    cols, rows = calculate_grid_layout(num_videos)
    print(f"🎯 Grid layout: {cols} columns × {rows} rows")
    
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
    
    # Calculate tile positions
    tile_positions = calculate_tile_positions(num_videos)
    
    # Scale values for tiling (make videos smaller to fit in grid)
    if num_videos <= 4:
        # Use animation.py scale values for 4 or fewer videos
        scale_values = ["-0.356424 0.356424", "0.313976 0.313976", "0.362066 0.362066", "0.265712 0.265712"]
    else:
        # For more videos, use smaller scale to fit more on screen
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