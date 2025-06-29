#!/usr/bin/env python3
"""
Many Video FX Command - Rebuilt from animation.py template

Creates step-by-step video animation effects using exact animation.py logic.
When --steps is specified, uses animation.py structure for perfect XML compatibility.
When --steps is omitted, falls back to original tiling behavior.
"""

import sys
import math
from pathlib import Path

from fcpxml_lib import create_empty_project, save_fcpxml
from fcpxml_lib.core.fcpxml import create_media_asset, detect_video_properties
from fcpxml_lib.models.elements import (
    Clip, Video, AdjustTransform, KeyframeAnimation, Keyframe, Param
)
from fcpxml_lib.utils.ids import generate_resource_id, set_resource_id_counter
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration


def calculate_step_positions(num_steps):
    """
    Calculate step-by-step positions based on Info.fcpxml pattern.
    
    For steps=2: Exact Info.fcpxml positions
    For steps=4: Extend to 4 corners
    For other steps: Grid pattern in upper area
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
            (-17.2101, 43.0307),  # Upper-left
            (2.38541, 43.2326),   # Upper-right
            (-17.2101, -43.0307), # Lower-left
            (2.38541, -43.2326)   # Lower-right
        ]
    else:
        # Grid pattern for other step counts
        positions = []
        cols = math.ceil(math.sqrt(num_steps))
        rows = math.ceil(num_steps / cols)
        
        x_spacing = 40.0 / cols
        y_spacing = 20.0 / rows
        
        start_x = -20.0 + (x_spacing / 2)
        start_y = 30.0 + (y_spacing / 2)
        
        for i in range(num_steps):
            row = i // cols
            col = i % cols
            
            x = start_x + (col * x_spacing)
            y = start_y + (row * y_spacing)
            
            positions.append((x, y))
        
        return positions


def create_step_animation_from_template(fcpxml, video_files, total_duration, steps):
    """
    Create step-by-step animation using EXACT animation.py template logic.
    
    This function follows animation.py structure exactly to ensure XML compatibility.
    """
    if not video_files or steps < 1:
        print("❌ No video files or invalid steps")
        return
    
    # Limit to available videos
    videos_to_use = video_files[:steps]
    step_positions = calculate_step_positions(steps)
    
    print(f"   Step-by-step animation: {len(videos_to_use)} videos")
    print(f"   Using exact animation.py template structure")
    
    # Set ID counter to start from r2 since r1 is already used by project format
    set_resource_id_counter(1)
    
    # Create media assets using create_media_asset (exactly like animation.py)
    video_assets = []
    video_formats = []
    video_props_list = []
    
    for i, video_file in enumerate(videos_to_use):
        print(f"   Processing video {i+1}/{len(videos_to_use)}: {video_file.name}")
        
        asset_id = generate_resource_id()
        format_id = generate_resource_id()
        
        try:
            asset, format_obj = create_media_asset(
                str(video_file), asset_id, format_id
            )
            video_props = detect_video_properties(str(video_file))
            
            video_assets.append(asset)
            video_formats.append(format_obj)
            video_props_list.append((video_props, video_file))
            
        except Exception as e:
            print(f"❌ Failed to process {video_file.name}: {e}")
            continue
    
    if not video_assets:
        print("❌ No valid video assets created")
        return
    
    # Add assets and formats to resources (exactly like animation.py)
    fcpxml.resources.assets.extend(video_assets)
    fcpxml.resources.formats.extend(video_formats)
    
    # Create timeline sequence (exactly like animation.py)
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format
    
    # Animation durations (exactly like animation.py)
    first_anim_duration = "144144/24000s"  # ~6 seconds
    second_anim_duration = "108108/24000s" # ~4.5 seconds  
    second_clip_offset = "36036/24000s"    # ~1.5 seconds delay
    
    # Get first video data
    first_video_props, first_video_file = video_props_list[0]
    first_position = step_positions[0]
    
    # Clip durations MUST match or exceed video durations (exactly like animation.py)
    first_clip_duration = convert_seconds_to_fcp_duration(first_video_props['duration_seconds'])
    
    # Create first video clip with keyframe animation (exactly like animation.py)
    first_clip = Clip(
        offset="0s",
        name=first_video_file.stem,
        duration=first_clip_duration,
        format=video_formats[0].id,
        tc_format="NDF",
        nested_elements=[]
    )
    
    # First video keyframe animation: center → position (exactly like animation.py)
    position_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time="0s", value="0 0"),  # Start at center
        Keyframe(time=first_anim_duration, value=f"{first_position[0]} {first_position[1]}")  # End position
    ])
    
    scale_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=first_anim_duration, value="-0.356424 0.356424", curve="linear")
    ])
    
    anchor_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=first_anim_duration, value="0 0", curve="linear")
    ])
    
    rotation_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=first_anim_duration, value="0", curve="linear")
    ])
    
    first_transform = AdjustTransform(
        params=[
            Param(name="anchor", keyframe_animation=anchor_keyframes),
            Param(name="position", keyframe_animation=position_keyframes),
            Param(name="rotation", keyframe_animation=rotation_keyframes),
            Param(name="scale", keyframe_animation=scale_keyframes)
        ]
    )
    
    # Add video element (removes audio) - exactly like animation.py
    first_video = Video(
        ref=video_assets[0].id,
        offset="0s",
        duration=convert_seconds_to_fcp_duration(first_video_props['duration_seconds'])
    )
    
    # Convert transforms to dictionary format (exactly like animation.py)
    first_transform_dict = first_transform.to_dict()
    
    # Build nested clips for remaining videos (following animation.py pattern)
    nested_clips = []
    
    for i in range(1, len(video_assets)):
        video_props, video_file = video_props_list[i]
        position = step_positions[i]
        
        # Calculate offset and duration
        clip_offset = f"{36036 * i}/24000s"  # Stagger timing like animation.py
        clip_duration = convert_seconds_to_fcp_duration(video_props['duration_seconds'])
        
        # Create nested clip (exactly like animation.py)
        nested_clip = Clip(
            lane=str(i),
            offset=clip_offset,
            name=video_file.stem,
            duration=clip_duration,
            tc_format="NDF",
            nested_elements=[]
        )
        
        # Animation keyframes for this video (exactly like animation.py)
        nested_position_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time="0s", value="0 0"),  # Start at center
            Keyframe(time=second_anim_duration, value=f"{position[0]} {position[1]}")  # End position
        ])
        
        nested_scale_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=second_anim_duration, value="0.313976 0.313976", curve="linear")
        ])
        
        nested_anchor_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=second_anim_duration, value="0 0", curve="linear")
        ])
        
        nested_rotation_keyframes = KeyframeAnimation(keyframes=[
            Keyframe(time=second_anim_duration, value="0", curve="linear")
        ])
        
        nested_transform = AdjustTransform(
            params=[
                Param(name="anchor", keyframe_animation=nested_anchor_keyframes),
                Param(name="position", keyframe_animation=nested_position_keyframes),
                Param(name="rotation", keyframe_animation=nested_rotation_keyframes),
                Param(name="scale", keyframe_animation=nested_scale_keyframes)
            ]
        )
        
        # Add nested video element (removes audio) - exactly like animation.py
        nested_video = Video(
            ref=video_assets[i].id,
            offset="0s",
            duration=convert_seconds_to_fcp_duration(video_props['duration_seconds'])
        )
        
        # Convert nested transform to dictionary format (exactly like animation.py)
        nested_transform_dict = nested_transform.to_dict()
        
        # Create nested clip dictionary (exactly like animation.py structure)
        nested_clip_dict = {
            "type": "clip",
            "lane": str(i),
            "offset": clip_offset,
            "name": video_file.stem,
            "duration": clip_duration,
            "format": video_formats[i].id,
            "tcFormat": "NDF",
            "nested_elements": [
                # Transform for nested clip
                {"type": "adjust_transform", **nested_transform_dict},
                # Nested video element (no audio)
                {
                    "type": "video",
                    "ref": video_assets[i].id,
                    "offset": "0s",
                    "duration": convert_seconds_to_fcp_duration(video_props['duration_seconds'])
                }
            ]
        }
        
        nested_clips.append(nested_clip_dict)
    
    # Create main clip element (using dictionary format for serializer) - exactly like animation.py
    main_clip_dict = {
        "type": "clip",
        "offset": "0s",
        "name": f"Step Animation - {len(video_assets)} videos",
        "duration": first_clip_duration,
        "format": video_formats[0].id,
        "tcFormat": "NDF",
        "nested_elements": [
            # Transform for main clip
            {"type": "adjust_transform", **first_transform_dict},
            # Video element (no audio)
            {
                "type": "video",
                "ref": video_assets[0].id,
                "offset": "0s",
                "duration": convert_seconds_to_fcp_duration(first_video_props['duration_seconds'])
            }
        ] + nested_clips  # Add all nested clips (exactly like animation.py)
    }
    
    # Add to spine (exactly like animation.py)
    sequence.spine.ordered_elements = [main_clip_dict]
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(total_duration)
    
    print(f"   Created {len(video_assets)} step-by-step animated videos")
    print(f"   Using exact animation.py structure for perfect XML compatibility")


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
    
    steps = getattr(args, 'steps', None)
    if steps is not None:
        print(f"   Step mode: {steps} videos (animation.py template)")
    else:
        print(f"   Error: No --steps parameter provided")
        print(f"   This rebuilt version requires --steps parameter")
        sys.exit(1)
    
    # Create empty project (always vertical for tiling)
    fcpxml = create_empty_project(
        project_name="Many Video FX",
        event_name="Step Video Animation",
        use_horizontal=False  # Always use vertical format
    )
    
    # Generate the step animation timeline using animation.py template
    try:
        create_step_animation_from_template(
            fcpxml,
            video_files,
            args.duration,
            steps
        )
        
        actual_steps = min(steps, len(video_files))
        print(f"✅ Timeline created with {actual_steps} step-by-step animated videos")
        
    except Exception as e:
        print(f"❌ Error creating step animation timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "many_video_fx.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)