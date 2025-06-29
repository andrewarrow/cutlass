"""
Many Video FX 2 Command Implementation - Copy of working animation.py

Creates FCPXML with keyframe animations exactly like Info.fcpxml pattern:
- Multiple videos with scaled-down animated movement from center to positions
- Audio tracks removed (video-only elements)
- Nested clip structure with keyframe transforms
- Uses --steps parameter to control how many videos animate
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


def many_video_fx2_cmd(args):
    """CLI implementation for many-video-fx2 command - copy of animation.py logic"""
    
    # Get video files from directory
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}", file=sys.stderr)
        sys.exit(1)
    
    # Find all .mov files
    video_files = list(input_dir.glob("*.mov"))
    
    if not video_files:
        print(f"❌ No .mov files found in {input_dir}", file=sys.stderr)
        sys.exit(1)
    
    # Get steps parameter
    steps = getattr(args, 'steps', 2)  # Default to 2 like animation.py
    
    # Limit to available videos
    videos_to_use = video_files[:steps]
    
    if len(videos_to_use) < 2:
        print(f"❌ Need at least 2 video files for animation, found {len(videos_to_use)}", file=sys.stderr)
        sys.exit(1)
    
    print(f"🎬 Many Video FX 2 - using {len(videos_to_use)} videos from {len(video_files)} found")
    print(f"   Directory: {input_dir}")
    print(f"   Steps: {steps}")
    
    # Get step positions
    step_positions = calculate_step_positions(len(videos_to_use))
    
    # Validate video files (just like animation.py)
    for video_path in videos_to_use:
        if not video_path.exists():
            print(f"❌ Video file not found: {video_path}", file=sys.stderr)
            sys.exit(1)
        
        if video_path.suffix.lower() not in {'.mp4', '.mov'}:
            print(f"❌ Unsupported video format: {video_path}", file=sys.stderr)
            print("   Supported formats: .mp4, .mov")
            sys.exit(1)
    
    # Create base project (already creates r1 vertical format) - EXACTLY like animation.py
    fcpxml = create_empty_project(use_horizontal=False)
    
    # Set ID counter to start from r2 since r1 is already used by project format - EXACTLY like animation.py
    set_resource_id_counter(1)
    
    # Generate resource IDs for media assets - EXACTLY like animation.py
    video_assets = []
    video_formats = []
    video_props_list = []
    
    for i, video_path in enumerate(videos_to_use):
        asset_id = generate_resource_id()   # r2, r4, r6, ...
        format_id = generate_resource_id()  # r3, r5, r7, ...
        
        # Create media assets - EXACTLY like animation.py
        try:
            asset, format_obj = create_media_asset(
                str(video_path), asset_id, format_id
            )
            
            # Detect properties for duration calculation - EXACTLY like animation.py
            video_props = detect_video_properties(str(video_path))
            
            video_assets.append(asset)
            video_formats.append(format_obj)
            video_props_list.append((video_props, video_path))
            
        except Exception as e:
            print(f"❌ Failed to process video files: {e}", file=sys.stderr)
            sys.exit(1)
    
    # Add assets and formats to resources - EXACTLY like animation.py
    fcpxml.resources.assets.extend(video_assets)
    fcpxml.resources.formats.extend(video_formats)
    
    # Create timeline sequence - EXACTLY like animation.py
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format from create_empty_project
    
    # Animation durations (matching Info.fcpxml pattern) - EXACTLY like animation.py
    first_anim_duration = "144144/24000s"  # ~6 seconds
    second_anim_duration = "108108/24000s" # ~4.5 seconds  
    second_clip_offset = "36036/24000s"    # ~1.5 seconds delay
    
    # Get first video data
    first_video_props, first_video_path = video_props_list[0]
    first_position = step_positions[0]
    
    # Clip durations MUST match or exceed video durations to prevent "Invalid edit" errors - EXACTLY like animation.py
    first_clip_duration = convert_seconds_to_fcp_duration(first_video_props['duration_seconds'])
    
    # Create first video clip with keyframe animation - EXACTLY like animation.py
    first_clip = Clip(
        offset="0s",
        name=first_video_path.stem,
        duration=first_clip_duration,
        format=video_formats[0].id,
        tc_format="NDF",
        nested_elements=[]
    )
    
    # First video keyframe animation: center → position - EXACTLY like animation.py structure
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
    
    # Add video element (removes audio) - EXACTLY like animation.py
    first_video = Video(
        ref=video_assets[0].id,
        offset="0s",
        duration=convert_seconds_to_fcp_duration(first_video_props['duration_seconds'])
    )
    
    # Convert transforms to dictionary format - EXACTLY like animation.py
    first_transform_dict = first_transform.to_dict()
    
    # Build nested clips for remaining videos - following animation.py pattern EXACTLY
    nested_clips = []
    
    for i in range(1, len(video_assets)):
        video_props, video_path = video_props_list[i]
        position = step_positions[i]
        
        # Calculate clip offset and duration
        clip_offset = f"{36036 * i}/24000s"  # Stagger timing
        clip_duration = convert_seconds_to_fcp_duration(video_props['duration_seconds'])
        
        # Create nested clip that starts later - EXACTLY like animation.py
        nested_clip = Clip(
            lane=str(i),
            offset=clip_offset,
            name=video_path.stem,
            duration=clip_duration,
            tc_format="NDF",
            nested_elements=[]
        )
        
        # Animation keyframes for this video - EXACTLY like animation.py
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
        
        # Add nested video element (removes audio) - EXACTLY like animation.py
        nested_video = Video(
            ref=video_assets[i].id,
            offset="0s",
            duration=convert_seconds_to_fcp_duration(video_props['duration_seconds'])
        )
        
        # Convert nested transform to dictionary format - EXACTLY like animation.py
        nested_transform_dict = nested_transform.to_dict()
        
        # Create nested clip dictionary - EXACTLY like animation.py structure
        nested_clip_dict = {
            "type": "clip",
            "lane": str(i),
            "offset": clip_offset,
            "name": video_path.stem,
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
    
    # Create main clip element (using dictionary format for serializer) - EXACTLY like animation.py
    main_clip_dict = {
        "type": "clip",
        "offset": "0s",
        "name": first_video_path.stem,  # Use first video name like animation.py
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
        ] + nested_clips  # Add all nested clips - EXACTLY like animation.py
    }
    
    # Add to spine - EXACTLY like animation.py
    sequence.spine.ordered_elements = [main_clip_dict]
    
    # Save FCPXML - EXACTLY like animation.py
    output_path = args.output
    try:
        success = save_fcpxml(fcpxml, output_path)
        if not success:
            print(f"❌ Failed to save FCPXML to {output_path}", file=sys.stderr)
            sys.exit(1)
            
        print(f"✅ Many Video FX 2 FCPXML created: {output_path}")
        for i, (_, video_path) in enumerate(video_props_list):
            if i == 0:
                print(f"   🎬 Video {i+1}: {video_path.name} (animates to {step_positions[i]})")
            else:
                print(f"   🎬 Video {i+1}: {video_path.name} (animates to {step_positions[i]})")
        print(f"   ⏱️  Using animation.py structure")
        print(f"   🔇 Audio tracks removed (video-only)")
        
    except Exception as e:
        print(f"❌ Error saving FCPXML: {e}", file=sys.stderr)
        sys.exit(1)