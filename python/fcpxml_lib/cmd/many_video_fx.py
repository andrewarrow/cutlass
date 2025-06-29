#!/usr/bin/env python3
"""
Many Video FX Command

Creates tiled video animation effect where videos start in center and animate to tile positions.
Each video is scaled down and follows keyframe animation from center to final position.
Pattern recreates Info.fcpxml timing where next video starts as previous one clears center.
"""

import sys
import math
from pathlib import Path

from fcpxml_lib import create_empty_project, save_fcpxml
from fcpxml_lib.core.fcpxml import create_media_asset, detect_video_properties
from fcpxml_lib.models.elements import (
    Clip, Video, AdjustTransform, KeyframeAnimation, Keyframe, Param,
    Asset, Format, MediaRep
)
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration
from fcpxml_lib.utils.ids import generate_resource_id, generate_uid
from fcpxml_lib.constants import (
    SCREEN_EDGE_LEFT, SCREEN_EDGE_RIGHT, SCREEN_EDGE_TOP, SCREEN_EDGE_BOTTOM,
    SCREEN_WIDTH, SCREEN_HEIGHT
)


def calculate_tile_positions(num_videos):
    """
    Calculate positions to tile videos across 1080x1920 screen.
    
    Based on screen constants:
    - Width: -46.875 to 46.875 (93.75 total)
    - Height: -93.75 to 93.75 (187.5 total)
    
    Returns list of (x, y) positions for each video.
    """
    positions = []
    
    # Calculate grid dimensions (try to make it roughly square)
    cols = math.ceil(math.sqrt(num_videos))
    rows = math.ceil(num_videos / cols)
    
    # Calculate spacing between tiles
    x_spacing = SCREEN_WIDTH / cols
    y_spacing = SCREEN_HEIGHT / rows
    
    # Calculate starting positions (top-left of grid)
    start_x = SCREEN_EDGE_LEFT + (x_spacing / 2)
    start_y = SCREEN_EDGE_TOP + (y_spacing / 2)
    
    # Generate positions
    for i in range(num_videos):
        row = i // cols
        col = i % cols
        
        x = start_x + (col * x_spacing)
        y = start_y + (row * y_spacing)
        
        positions.append((x, y))
    
    return positions


def calculate_step_positions(num_steps):
    """
    Calculate step-by-step positions based on Info.fcpxml pattern.
    
    For steps=2: Upper-left (-17.2101, 43.0307), Upper-right (2.38541, 43.2326)
    For steps=4: Add bottom positions
    
    Returns list of (x, y) positions for step-by-step animation.
    """
    if num_steps == 2:
        # Exact Info.fcpxml pattern
        return [
            (-17.2101, 43.0307),  # Upper-left (first video)
            (2.38541, 43.2326)    # Upper-right (second video)
        ]
    elif num_steps == 4:
        # Extend to 4 corners
        return [
            (-17.2101, 43.0307),  # Upper-left (first video)
            (2.38541, 43.2326),   # Upper-right (second video)
            (-17.2101, -43.0307), # Lower-left (third video)
            (2.38541, -43.2326)   # Lower-right (fourth video)
        ]
    else:
        # For other step counts, create a simple grid pattern in upper area
        positions = []
        cols = math.ceil(math.sqrt(num_steps))
        rows = math.ceil(num_steps / cols)
        
        # Focus on upper half of screen
        x_spacing = 40.0 / cols  # Spread across ~40 units width
        y_spacing = 20.0 / rows  # Use upper ~20 units height
        
        start_x = -20.0 + (x_spacing / 2)
        start_y = 30.0 + (y_spacing / 2)
        
        for i in range(num_steps):
            row = i // cols
            col = i % cols
            
            x = start_x + (col * x_spacing)
            y = start_y + (row * y_spacing)
            
            positions.append((x, y))
        
        return positions


def create_step_animation_timeline(fcpxml, video_files, total_duration, steps):
    """
    Create step-by-step animation timeline using exact animation.py logic.
    
    For steps=2: Uses identical structure to animation command
    For other steps: Creates similar nested structure  
    """
    if not video_files or steps < 1:
        print("❌ No video files or invalid steps")
        return
    
    # Limit to available videos
    videos_to_use = video_files[:steps]
    step_positions = calculate_step_positions(steps)
    
    print(f"   Step-by-step animation: {len(videos_to_use)} videos")
    print(f"   Using animation.py logic for proper dataclass handling")
    
    # Set ID counter to start from r2 since r1 is already used by project format
    from fcpxml_lib.utils.ids import set_resource_id_counter
    set_resource_id_counter(1)
    
    # Create media assets using create_media_asset (like animation.py)
    video_assets = []
    for i, video_file in enumerate(videos_to_use):
        print(f"   Processing video {i+1}/{len(videos_to_use)}: {video_file.name}")
        
        asset_id = generate_resource_id()
        format_id = generate_resource_id()
        
        try:
            asset, format_obj = create_media_asset(
                str(video_file), asset_id, format_id
            )
            video_props = detect_video_properties(str(video_file))
            video_assets.append((asset, format_obj, video_props, video_file))
        except Exception as e:
            print(f"❌ Failed to process {video_file.name}: {e}")
            continue
    
    if not video_assets:
        print("❌ No valid video assets created")
        return
    
    # Add assets and formats to resources
    for asset, format_obj, _, _ in video_assets:
        fcpxml.resources.assets.append(asset)
        fcpxml.resources.formats.append(format_obj)
    
    # Get sequence
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format
    
    # Animation timing (matching Info.fcpxml)
    first_anim_duration = "144144/24000s"  # ~6 seconds
    second_anim_duration = "108108/24000s" # ~4.5 seconds
    clip_offset = "36036/24000s"           # ~1.5 seconds delay between clips
    
    # Create first video (main container)
    first_asset, first_format, first_props, first_file = video_assets[0]
    first_position = step_positions[0]
    
    # First clip duration
    first_clip_duration = convert_seconds_to_fcp_duration(first_props['duration_seconds'])
    
    # First video keyframe animation
    first_position_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time="0s", value="0 0"),  # Start at center
        Keyframe(time=first_anim_duration, value=f"{first_position[0]} {first_position[1]}")  # End position
    ])
    
    first_scale_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=first_anim_duration, value="-0.356424 0.356424", curve="linear")
    ])
    
    first_anchor_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=first_anim_duration, value="0 0", curve="linear")
    ])
    
    first_rotation_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=first_anim_duration, value="0", curve="linear")
    ])
    
    first_transform = AdjustTransform(
        params=[
            Param(name="anchor", keyframe_animation=first_anchor_keyframes),
            Param(name="position", keyframe_animation=first_position_keyframes),
            Param(name="rotation", keyframe_animation=first_rotation_keyframes),
            Param(name="scale", keyframe_animation=first_scale_keyframes)
        ]
    )
    
    # Create first video element
    first_video = Video(
        ref=first_asset.id,
        offset="0s",
        duration=convert_seconds_to_fcp_duration(first_props['duration_seconds'])
    )
    
    # Build nested elements list (like animation.py)
    nested_elements = []
    
    # Add remaining videos as nested clips
    for i in range(1, len(video_assets)):
        asset, format_obj, video_props, video_file = video_assets[i]
        position = step_positions[i]
        
        # Calculate offset for this video
        video_offset = f"{36036 * i}/24000s"  # Stagger timing
        clip_duration = convert_seconds_to_fcp_duration(video_props['duration_seconds'])
        
        # Animation keyframes for this video
        position_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time="0s", value="0 0"),  # Start at center
            Keyframe(time=second_anim_duration, value=f"{position[0]} {position[1]}")  # End position
        ])
        
        scale_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=second_anim_duration, value="0.313976 0.313976", curve="linear")
        ])
        
        anchor_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=second_anim_duration, value="0 0", curve="linear")
        ])
        
        rotation_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=second_anim_duration, value="0", curve="linear")
        ])
        
        transform = AdjustTransform(
            params=[
                Param(name="anchor", keyframe_animation=anchor_keyframes),
                Param(name="position", keyframe_animation=position_keyframes),
                Param(name="rotation", keyframe_animation=rotation_keyframes),
                Param(name="scale", keyframe_animation=scale_keyframes)
            ]
        )
        
        # Create nested clip (using dictionary format for serializer)
        nested_clip_dict = {
            "type": "clip",
            "lane": str(i),
            "offset": video_offset,
            "name": video_file.stem,
            "duration": clip_duration,
            "format": format_obj.id,
            "tcFormat": "NDF",
            "nested_elements": [
                # Transform
                {"type": "adjust_transform", **transform.to_dict()},
                # Video element
                {
                    "type": "video",
                    "ref": asset.id,
                    "offset": "0s",
                    "duration": convert_seconds_to_fcp_duration(video_props['duration_seconds'])
                }
            ]
        }
        
        nested_elements.append(nested_clip_dict)
    
    # Create main clip using dictionary format (like animation.py)
    main_clip_dict = {
        "type": "clip",
        "offset": "0s",
        "name": f"Step Animation - {len(video_assets)} videos",
        "duration": first_clip_duration,
        "format": first_format.id,
        "tcFormat": "NDF",
        "nested_elements": [
            # Transform for main clip
            {"type": "adjust_transform", **first_transform.to_dict()},
            # First video element
            {
                "type": "video",
                "ref": first_asset.id,
                "offset": "0s",
                "duration": convert_seconds_to_fcp_duration(first_props['duration_seconds'])
            }
        ] + nested_elements  # Add all nested clips
    }
    
    # Add to spine
    sequence.spine.ordered_elements = [main_clip_dict]
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(total_duration)
    
    print(f"   Created {len(video_assets)} step-by-step animated videos")
    print(f"   Using exact animation.py structure for XML compatibility")


def create_tiled_video_timeline(fcpxml, video_files, total_duration, steps=None):
    """
    Create timeline with videos that animate from center to tile positions.
    
    Animation pattern (based on Info.fcpxml):
    - Each video starts at center (0, 0) and animates to final position
    - Videos are scaled down for tiling
    - Next video starts as soon as previous one clears center
    - Uses proven structure from animation.py command
    
    Args:
        steps: If provided, use step-by-step animation (2=Info.fcpxml pattern)
               If None, tile all videos simultaneously
    """
    if not video_files:
        print("❌ No video files provided")
        return
    
    num_videos = len(video_files)
    
    # Determine positioning strategy
    if steps is not None:
        # Step-by-step animation (Info.fcpxml pattern)
        if steps > num_videos:
            print(f"⚠️  Warning: --steps {steps} > {num_videos} videos, using {num_videos} steps")
            steps = num_videos
        
        tile_positions = calculate_step_positions(steps)
        print(f"   Step-by-step animation: {steps} videos")
        print(f"   Pattern: Info.fcpxml style (center to corners)")
    else:
        # Traditional tiling
        tile_positions = calculate_tile_positions(num_videos)
        print(f"   Tiling {num_videos} videos in {math.ceil(math.sqrt(num_videos))} columns")
        print(f"   Screen bounds: X({SCREEN_EDGE_LEFT} to {SCREEN_EDGE_RIGHT}), Y({SCREEN_EDGE_TOP} to {SCREEN_EDGE_BOTTOM})")
    
    # Animation timing (based on Info.fcpxml pattern)
    animation_duration_fcp = "144144/24000s"  # ~6 seconds (same as Info.fcpxml)
    overlap_time_fcp = "36036/24000s"         # ~1.5 seconds delay
    
    # Use exact scale values from Info.fcpxml pattern
    # First video uses negative X scale (flip), others use positive
    info_scales = ["-0.356424 0.356424", "0.313976 0.313976"]
    
    # For step-by-step animation, limit to steps count
    videos_to_animate = min(len(video_files), steps) if steps is not None else len(video_files)
    
    # Get sequence and set it up like animation.py
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format
    
    # Set ID counter to start from r2 since r1 is already used by project format
    from fcpxml_lib.utils.ids import set_resource_id_counter
    set_resource_id_counter(1)
    
    # Create assets for videos (limited by steps if specified)
    video_assets = []
    video_formats = []
    
    for i, video_file in enumerate(video_files[:videos_to_animate]):
        print(f"   Processing video {i+1}/{videos_to_animate}: {video_file.name}")
        
        # Generate resource IDs
        asset_id = generate_resource_id()
        format_id = generate_resource_id()
        
        # Create asset for this video (video-only, no audio to match Info.fcpxml pattern)
        try:
            from fcpxml_lib.core.fcpxml import detect_video_properties
            video_props = detect_video_properties(str(video_file))
            
            # Create video-only asset (strip audio properties like animation.py does)
            abs_path = video_file.resolve()
            uid = generate_uid(f"VIDEO_{abs_path.name}")
            media_rep = MediaRep(src=str(abs_path))
            
            # Create video-only asset (no hasAudio, audioSources, etc.)
            asset = Asset(
                id=asset_id,
                name=abs_path.stem,
                uid=uid,
                duration=convert_seconds_to_fcp_duration(video_props['duration_seconds']),
                has_video="1",
                format=format_id,
                video_sources="1",
                media_rep=media_rep
            )
            
            # Create format 
            format_obj = Format(
                id=format_id,
                frame_duration="1001/24000s",
                width=str(video_props['width']),
                height=str(video_props['height']),
                color_space="1-1-1 (Rec. 709)"
            )
            
            video_assets.append((asset, format_obj, video_props, video_file))
        except Exception as e:
            print(f"❌ Failed to process {video_file.name}: {e}")
            continue
    
    # Add all assets and formats to resources
    for asset, format_obj, _, _ in video_assets:
        fcpxml.resources.assets.append(asset)
        fcpxml.resources.formats.append(format_obj)
    
    # Create main clip structure (first video is the container, like Info.fcpxml)
    if not video_assets:
        print("❌ No valid video assets created")
        return
    
    first_asset, first_format, first_props, first_file = video_assets[0]
    final_x, final_y = tile_positions[0]
    
    # Create first video animation (container clip)
    first_position_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time="0s", value="0 0"),  # Start at center
        Keyframe(time=animation_duration_fcp, value=f"{final_x:.4f} {final_y:.4f}")  # End at tile position
    ])
    
    first_scale_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=animation_duration_fcp, value=info_scales[0], curve="linear")
    ])
    
    first_anchor_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=animation_duration_fcp, value="0 0", curve="linear")
    ])
    
    first_rotation_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=animation_duration_fcp, value="0", curve="linear")
    ])
    
    first_transform = AdjustTransform(
        params=[
            Param(name="anchor", keyframe_animation=first_anchor_keyframes),
            Param(name="position", keyframe_animation=first_position_keyframes),
            Param(name="rotation", keyframe_animation=first_rotation_keyframes),
            Param(name="scale", keyframe_animation=first_scale_keyframes)
        ]
    )
    
    # Build nested clips for remaining videos
    nested_clips = []
    
    for i in range(1, len(video_assets)):
        asset, format_obj, video_props, video_file = video_assets[i]
        final_x, final_y = tile_positions[i]
        
        # Calculate offset for this video (staggered timing)
        video_offset_multiplier = i
        video_offset = f"{36036 * video_offset_multiplier}/24000s"  # Multiply base offset
        
        # Create animation for this video
        position_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time="0s", value="0 0"),  # Start at center
            Keyframe(time=animation_duration_fcp, value=f"{final_x:.4f} {final_y:.4f}")  # End at tile position
        ])
        
        # Use Info.fcpxml scale pattern - alternate between the two scale values
        scale_value = info_scales[1] if i < len(info_scales) else info_scales[1]  # Use second scale for all others
        scale_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=animation_duration_fcp, value=scale_value, curve="linear")
        ])
        
        anchor_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=animation_duration_fcp, value="0 0", curve="linear")
        ])
        
        rotation_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=animation_duration_fcp, value="0", curve="linear")
        ])
        
        transform = AdjustTransform(
            params=[
                Param(name="anchor", keyframe_animation=anchor_keyframes),
                Param(name="position", keyframe_animation=position_keyframes),
                Param(name="rotation", keyframe_animation=rotation_keyframes),
                Param(name="scale", keyframe_animation=scale_keyframes)
            ]
        )
        
        # Create nested clip dict (following animation.py pattern)
        nested_clip_dict = {
            "type": "clip",
            "lane": str(i),  # Each video on its own lane
            "offset": video_offset,
            "name": video_file.stem,
            "duration": convert_seconds_to_fcp_duration(video_props['duration_seconds']),
            "format": format_obj.id,
            "tcFormat": "NDF",
            "nested_elements": [
                # Transform for nested clip
                {"type": "adjust_transform", **transform.to_dict()},
                # Video element (no audio)
                {
                    "type": "video",
                    "ref": asset.id,
                    "offset": "0s",
                    "duration": convert_seconds_to_fcp_duration(video_props['duration_seconds'])
                }
            ]
        }
        
        nested_clips.append(nested_clip_dict)
    
    # Create main clip dict (following animation.py pattern exactly)
    first_clip_duration = convert_seconds_to_fcp_duration(first_props['duration_seconds'])
    
    main_clip_dict = {
        "type": "clip",
        "offset": "0s",
        "name": f"Many Video FX - {len(video_assets)} videos",
        "duration": first_clip_duration,
        "format": first_format.id,
        "tcFormat": "NDF",
        "nested_elements": [
            # Transform for main clip
            {"type": "adjust_transform", **first_transform.to_dict()},
            # First video element (no audio)
            {
                "type": "video",
                "ref": first_asset.id,
                "offset": "0s",
                "duration": convert_seconds_to_fcp_duration(first_props['duration_seconds'])
            }
        ] + nested_clips  # Add all nested clips
    }
    
    # Add to spine (like animation.py)
    sequence.spine.ordered_elements = [main_clip_dict]
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(total_duration)
    
    print(f"   Created {len(video_assets)} animated video tiles")
    print(f"   Animation duration: ~6s per video")
    print(f"   Overlap timing: ~1.5s between starts")
    print(f"   Total timeline duration: {total_duration}s")


def many_video_fx_cmd(args):
    """Create tiled video animation effect from directory of .mov files"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Find all .mov files
    video_files = list(input_dir.glob("*.mov"))
    
    if not video_files:
        print(f"❌ No .mov files found in {input_dir}")
        sys.exit(1)
    
    print(f"🎬 Creating many-video-fx animation...")
    print(f"   Input directory: {input_dir}")
    print(f"   Video files found: {len(video_files)}")
    print(f"   Duration: {args.duration}s")
    if hasattr(args, 'steps') and args.steps:
        print(f"   Step mode: {args.steps} videos (Info.fcpxml pattern)")
    else:
        print(f"   Tile mode: All videos simultaneously")
    
    # Create empty project (always vertical for tiling)
    fcpxml = create_empty_project(
        project_name="Many Video FX",
        event_name="Tiled Video Animation",
        use_horizontal=False  # Always use vertical format
    )
    
    # Generate the tiled video timeline
    try:
        steps = getattr(args, 'steps', None)
        if steps is not None:
            # Use step-by-step animation (animation.py logic)
            create_step_animation_timeline(
                fcpxml,
                video_files,
                args.duration,
                steps
            )
        else:
            # Use original tiling logic
            create_tiled_video_timeline(
                fcpxml, 
                video_files, 
                args.duration,
                steps=None
            )
        
        steps_used = getattr(args, 'steps', None)
        if steps_used:
            actual_steps = min(steps_used, len(video_files))
            print(f"✅ Timeline created with {actual_steps} step-by-step animated videos")
        else:
            print(f"✅ Timeline created with {len(video_files)} animated video tiles")
        
    except Exception as e:
        print(f"❌ Error creating tiled video timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "many_video_fx.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)