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
from fcpxml_lib.core.fcpxml import create_media_asset
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


def create_tiled_video_timeline(fcpxml, video_files, total_duration):
    """
    Create timeline with videos that animate from center to tile positions.
    
    Animation pattern (based on Info.fcpxml):
    - Each video starts at center (0, 0) and animates to final position
    - Videos are scaled down for tiling
    - Next video starts as soon as previous one clears center
    - Uses proven structure from animation.py command
    """
    if not video_files:
        print("❌ No video files provided")
        return
    
    # Calculate tile positions
    num_videos = len(video_files)
    tile_positions = calculate_tile_positions(num_videos)
    
    print(f"   Tiling {num_videos} videos in {math.ceil(math.sqrt(num_videos))} columns")
    print(f"   Screen bounds: X({SCREEN_EDGE_LEFT} to {SCREEN_EDGE_RIGHT}), Y({SCREEN_EDGE_TOP} to {SCREEN_EDGE_BOTTOM})")
    
    # Animation timing (based on Info.fcpxml pattern)
    animation_duration_fcp = "144144/24000s"  # ~6 seconds (same as Info.fcpxml)
    overlap_time_fcp = "36036/24000s"         # ~1.5 seconds delay
    
    # Use exact scale values from Info.fcpxml pattern
    # First video uses negative X scale (flip), others use positive
    info_scales = ["-0.356424 0.356424", "0.313976 0.313976"]
    
    # Get sequence and set it up like animation.py
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format
    
    # Set ID counter to start from r2 since r1 is already used by project format
    from fcpxml_lib.utils.ids import set_resource_id_counter
    set_resource_id_counter(1)
    
    # Create assets for all videos
    video_assets = []
    video_formats = []
    
    for i, video_file in enumerate(video_files):
        print(f"   Processing video {i+1}/{num_videos}: {video_file.name}")
        
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
    
    # Create empty project (always vertical for tiling)
    fcpxml = create_empty_project(
        project_name="Many Video FX",
        event_name="Tiled Video Animation",
        use_horizontal=False  # Always use vertical format
    )
    
    # Generate the tiled video timeline
    try:
        create_tiled_video_timeline(
            fcpxml, 
            video_files, 
            args.duration
        )
        
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