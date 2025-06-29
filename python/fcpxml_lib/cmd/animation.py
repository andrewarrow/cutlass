"""
Animation Command Implementation

Creates FCPXML with keyframe animations exactly like Info.fcpxml pattern:
- Two videos with scaled-down animated movement from center to corners
- Audio tracks removed (video-only elements)
- Nested clip structure with keyframe transforms
"""

import sys
from pathlib import Path

from fcpxml_lib.core.fcpxml import create_empty_project, save_fcpxml, create_media_asset, detect_video_properties
from fcpxml_lib.models.elements import (
    Clip, Video, AdjustTransform, KeyframeAnimation, Keyframe, Param
)
from fcpxml_lib.utils.ids import generate_resource_id, set_resource_id_counter
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration


def animation_cmd(args):
    """CLI implementation for animation command"""
    
    # Validate input files
    if len(args.input_files) != 2:
        print("❌ Animation command requires exactly 2 video files", file=sys.stderr)
        sys.exit(1)
    
    video1_path = Path(args.input_files[0])
    video2_path = Path(args.input_files[1])
    
    for video_path in [video1_path, video2_path]:
        if not video_path.exists():
            print(f"❌ Video file not found: {video_path}", file=sys.stderr)
            sys.exit(1)
        
        if video_path.suffix.lower() not in {'.mp4', '.mov'}:
            print(f"❌ Unsupported video format: {video_path}", file=sys.stderr)
            print("   Supported formats: .mp4, .mov")
            sys.exit(1)
    
    # Create base project (already creates r1 vertical format)
    fcpxml = create_empty_project(use_horizontal=False)
    
    # Set ID counter to start from r2 since r1 is already used by project format
    set_resource_id_counter(1)
    
    # Generate resource IDs for media assets
    video1_asset_id = generate_resource_id()     # r2
    video1_format_id = generate_resource_id()    # r3
    video2_asset_id = generate_resource_id()     # r4
    video2_format_id = generate_resource_id()    # r5
    
    # Create media assets
    try:
        video1_asset, video1_format = create_media_asset(
            str(video1_path), video1_asset_id, video1_format_id
        )
        video2_asset, video2_format = create_media_asset(
            str(video2_path), video2_asset_id, video2_format_id
        )
        
        # Detect properties for duration calculation
        video1_props = detect_video_properties(str(video1_path))
        video2_props = detect_video_properties(str(video2_path))
        
    except Exception as e:
        print(f"❌ Failed to process video files: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Add assets and formats to resources
    fcpxml.resources.assets.extend([video1_asset, video2_asset])
    fcpxml.resources.formats.extend([video1_format, video2_format])
    
    # Create timeline sequence
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format from create_empty_project
    
    # Animation durations (matching Info.fcpxml pattern)
    first_anim_duration = "144144/24000s"  # ~6 seconds
    second_anim_duration = "108108/24000s" # ~4.5 seconds  
    second_clip_offset = "36036/24000s"    # ~1.5 seconds delay
    
    # Total clip duration should accommodate both animations
    total_duration = convert_seconds_to_fcp_duration(8.0)  # 8 seconds total
    
    # Create first video clip with keyframe animation
    first_clip = Clip(
        offset="0s",
        name=video1_path.stem,
        duration=total_duration,
        format=video1_format_id,
        tc_format="NDF",
        nested_elements=[]
    )
    
    # First video keyframe animation: center → left corner
    position_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time="0s", value="0 0"),  # Start at center
        Keyframe(time=first_anim_duration, value="-17.2101 43.0307")  # End at left corner
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
    
    # Add video element (removes audio)
    first_video = Video(
        ref=video1_asset_id,
        offset="0s",
        duration=convert_seconds_to_fcp_duration(video1_props['duration_seconds'])
    )
    
    # Create nested second clip that starts later
    second_clip = Clip(
        lane="1",
        offset=second_clip_offset,  # 1.5s delay
        name=video2_path.stem,
        duration=convert_seconds_to_fcp_duration(6.0),  # Enough for animation
        tc_format="NDF",
        nested_elements=[]
    )
    
    # Second video keyframe animation: center → right corner  
    second_position_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time="0s", value="0 0"),  # Start at center
        Keyframe(time=second_anim_duration, value="2.38541 43.2326")  # End at right corner
    ])
    
    second_scale_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=second_anim_duration, value="0.313976 0.313976", curve="linear")
    ])
    
    second_anchor_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=second_anim_duration, value="0 0", curve="linear")
    ])
    
    second_rotation_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time=second_anim_duration, value="0", curve="linear")
    ])
    
    second_transform = AdjustTransform(
        params=[
            Param(name="anchor", keyframe_animation=second_anchor_keyframes),
            Param(name="position", keyframe_animation=second_position_keyframes),
            Param(name="rotation", keyframe_animation=second_rotation_keyframes),
            Param(name="scale", keyframe_animation=second_scale_keyframes)
        ]
    )
    
    # Add second video element (removes audio)
    second_video = Video(
        ref=video2_asset_id,
        offset="0s",
        duration=convert_seconds_to_fcp_duration(video2_props['duration_seconds'])
    )
    
    # Convert dataclasses to dictionary format for serializer
    # Build nested structure: first_clip contains second_clip (Info.fcpxml pattern)
    
    # Convert transforms to dictionary format
    first_transform_dict = first_transform.to_dict()
    second_transform_dict = second_transform.to_dict()
    
    # Create main clip element (using dictionary format for serializer)
    main_clip_dict = {
        "type": "clip",
        "offset": "0s",
        "name": video1_path.stem,
        "duration": total_duration,
        "format": video1_format_id,
        "tcFormat": "NDF",
        "nested_elements": [
            # Transform for main clip
            {"type": "adjust_transform", **first_transform_dict},
            # Video element (no audio)
            {
                "type": "video",
                "ref": video1_asset_id,
                "offset": "0s",
                "duration": convert_seconds_to_fcp_duration(video1_props['duration_seconds'])
            },
            # Nested second clip
            {
                "type": "clip",
                "lane": "1",
                "offset": second_clip_offset,
                "name": video2_path.stem,
                "duration": convert_seconds_to_fcp_duration(6.0),
                "format": video2_format_id,
                "tcFormat": "NDF",
                "nested_elements": [
                    # Transform for nested clip
                    {"type": "adjust_transform", **second_transform_dict},
                    # Second video element (no audio)
                    {
                        "type": "video",
                        "ref": video2_asset_id,
                        "offset": "0s",
                        "duration": convert_seconds_to_fcp_duration(video2_props['duration_seconds'])
                    }
                ]
            }
        ]
    }
    
    # Add to spine
    sequence.spine.ordered_elements = [main_clip_dict]
    
    # Save FCPXML
    output_path = args.output_path
    try:
        success = save_fcpxml(fcpxml, output_path)
        if not success:
            print(f"❌ Failed to save FCPXML to {output_path}", file=sys.stderr)
            sys.exit(1)
            
        print(f"✅ Animation FCPXML created: {output_path}")
        print(f"   🎬 Video 1: {video1_path.name} (animates to left corner)")
        print(f"   🎬 Video 2: {video2_path.name} (animates to right corner)")
        print(f"   ⏱️  Total duration: 8 seconds")
        print(f"   🔇 Audio tracks removed (video-only)")
        
    except Exception as e:
        print(f"❌ Error saving FCPXML: {e}", file=sys.stderr)
        sys.exit(1)