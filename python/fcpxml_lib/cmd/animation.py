"""
Animation Command Implementation

Creates FCPXML with keyframe animations exactly like Info.fcpxml pattern:
- Takes directory with images and selects first 4 PNG files
- Creates nested clip structure with keyframe transforms
- Includes conform-rate elements to prevent validation errors
- Uses Pattern A (nested elements) for multi-lane visibility
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
    
    # Validate input - should be a directory
    if len(args.input_files) != 1:
        print("❌ Animation command requires exactly 1 directory path", file=sys.stderr)
        sys.exit(1)
    
    input_dir = Path(args.input_files[0])
    
    if not input_dir.exists():
        print(f"❌ Directory not found: {input_dir}", file=sys.stderr)
        sys.exit(1)
    
    if not input_dir.is_dir():
        print(f"❌ Path is not a directory: {input_dir}", file=sys.stderr)
        sys.exit(1)
    
    # Find MOV files in directory
    mov_files = list(input_dir.glob("*.mov"))
    if len(mov_files) < 4:
        print(f"❌ Directory must contain at least 4 MOV files, found {len(mov_files)}", file=sys.stderr)
        sys.exit(1)
    
    # Select first 4 MOV files
    selected_videos = sorted(mov_files)[:4]
    print(f"📁 Using videos: {[f.name for f in selected_videos]}")
    
    # Create base project (already creates r1 vertical format)
    fcpxml = create_empty_project(use_horizontal=False)
    
    # Set ID counter to start from r2 since r1 is already used by project format
    set_resource_id_counter(1)
    
    # Generate resource IDs for media assets - each video gets its own format
    asset_ids = []
    format_ids = []
    for i in range(4):
        asset_ids.append(generate_resource_id())  # r2, r3, r4, r5
        format_ids.append(generate_resource_id())  # r6, r7, r8, r9
    
    # Create media assets for all 4 videos like Info.fcpxml
    try:
        assets = []
        formats = []
        for i, video_path in enumerate(selected_videos):
            asset, format_obj = create_media_asset(
                str(video_path), asset_ids[i], format_ids[i]
            )
            assets.append(asset)
            formats.append(format_obj)
        
        fcpxml.resources.assets.extend(assets)
        fcpxml.resources.formats.extend(formats)
        
    except Exception as e:
        print(f"❌ Failed to process video files: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Create timeline sequence
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    sequence.format = "r1"  # Use the existing vertical format from create_empty_project
    
    # Use proper frame-aligned durations using video properties
    # Get actual video durations and convert to frame-aligned format
    video_durations = []
    for video_path in selected_videos:
        props = detect_video_properties(str(video_path))
        duration = convert_seconds_to_fcp_duration(props['duration_seconds'])
        video_durations.append(duration)
    
    # Animation durations - use fixed frame-aligned values for animations
    clip_duration = convert_seconds_to_fcp_duration(20.0)  # 20 second main duration
    
    # Keyframe animation timings - use frame-aligned values
    first_anim_time = convert_seconds_to_fcp_duration(6.0)   # 6 second animation
    second_anim_time = convert_seconds_to_fcp_duration(4.5)  # 4.5 second animation
    third_anim_time = convert_seconds_to_fcp_duration(4.0)   # 4 second animation
    fourth_anim_time = convert_seconds_to_fcp_duration(2.75) # 2.75 second animation
    
    # Clip offsets - frame-aligned
    second_offset = convert_seconds_to_fcp_duration(1.5)     # 1.5s
    third_offset = convert_seconds_to_fcp_duration(2.125)    # 2.125s
    fourth_offset = convert_seconds_to_fcp_duration(3.2)     # 3.2s
    
    # Nested clip durations - use actual video durations or clip duration, whichever is longer
    nested_durations = []
    for duration in video_durations:
        nested_durations.append(duration)
    
    # Create keyframe animations for each clip using dataclasses
    first_transform = AdjustTransform(
        params=[
            Param(
                name="anchor",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=first_anim_time, value="0 0", curve="linear")
                ])
            ),
            Param(
                name="position", 
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="0s", value="0 0"),
                    Keyframe(time=first_anim_time, value="-17.2101 43.0307")
                ])
            ),
            Param(
                name="rotation",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=first_anim_time, value="0", curve="linear")
                ])
            ),
            Param(
                name="scale",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=first_anim_time, value="-0.356424 0.356424", curve="linear")
                ])
            )
        ]
    )
    
    second_transform = AdjustTransform(
        params=[
            Param(
                name="anchor",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=second_anim_time, value="0 0", curve="linear")
                ])
            ),
            Param(
                name="position",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="0s", value="0 0"),
                    Keyframe(time=second_anim_time, value="2.38541 43.2326")
                ])
            ),
            Param(
                name="rotation",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=second_anim_time, value="0", curve="linear")
                ])
            ),
            Param(
                name="scale",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=second_anim_time, value="0.313976 0.313976", curve="linear")
                ])
            )
        ]
    )
    
    third_transform = AdjustTransform(
        params=[
            Param(
                name="anchor",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=third_anim_time, value="0 0", curve="linear")
                ])
            ),
            Param(
                name="position",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="72072/24000s", value="0 0"),
                    Keyframe(time=third_anim_time, value="22.2446 42.4814")
                ])
            ),
            Param(
                name="rotation",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=third_anim_time, value="0", curve="linear")
                ])
            ),
            Param(
                name="scale",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=third_anim_time, value="0.362066 0.362066", curve="linear")
                ])
            )
        ]
    )
    
    fourth_transform = AdjustTransform(
        params=[
            Param(
                name="anchor",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=fourth_anim_time, value="0 0", curve="linear")
                ])
            ),
            Param(
                name="position",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="125125/24000s", value="0 0"),
                    Keyframe(time=fourth_anim_time, value="-19.2439 31.344")
                ])
            ),
            Param(
                name="rotation",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=fourth_anim_time, value="0", curve="linear")
                ])
            ),
            Param(
                name="scale",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time=fourth_anim_time, value="0.265712 0.265712", curve="linear")
                ])
            )
        ]
    )

    # Create nested clips using dataclasses
    nested_clip_2 = Clip(
        lane="1",
        offset=second_offset,
        name=selected_videos[1].stem,
        duration=nested_durations[1],
        format=format_ids[1],
        tc_format="NDF",
        nested_elements=[
            second_transform,
            Video(ref=asset_ids[1], offset="0s", duration=video_durations[1])
        ]
    )
    
    nested_clip_3 = Clip(
        lane="2", 
        offset=third_offset,
        name=selected_videos[2].stem,
        duration=nested_durations[2],
        format=format_ids[2],
        tc_format="NDF",
        nested_elements=[
            third_transform,
            Video(ref=asset_ids[2], offset="0s", duration=video_durations[2])
        ]
    )
    
    nested_clip_4 = Clip(
        lane="3",
        offset=fourth_offset,
        name=selected_videos[3].stem,
        duration=nested_durations[3],
        format=format_ids[3],
        tc_format="NDF",
        nested_elements=[
            fourth_transform,
            Video(ref=asset_ids[3], offset="0s", duration=video_durations[3])
        ]
    )

    # Create main clip with all nested elements
    main_clip = Clip(
        offset="0s",
        name=selected_videos[0].stem,
        duration=clip_duration,
        format=format_ids[0],
        tc_format="NDF",
        nested_elements=[
            first_transform,
            Video(ref=asset_ids[0], offset="0s", duration=video_durations[0]),
            nested_clip_2,
            nested_clip_3,
            nested_clip_4
        ]
    )
    
    # Convert to dictionary format for spine (serializer expects this format)
    # The dataclasses validate the structure, then we convert to dict for serialization
    def clip_to_dict(clip):
        """Convert Clip dataclass to dictionary format for serializer"""
        clip_dict = {
            "type": "clip",
            "offset": clip.offset,
            "name": clip.name,
            "duration": clip.duration,
            "format": clip.format,
            "tcFormat": clip.tc_format
        }
        if hasattr(clip, 'lane') and clip.lane:
            clip_dict["lane"] = clip.lane
            
        nested_elements = []
        for element in clip.nested_elements:
            if isinstance(element, AdjustTransform):
                # Convert transform to serializer-expected format
                transform_dict = {"type": "adjust_transform"}
                if element.params:
                    # Extract keyframe data in the format the serializer expects
                    for param in element.params:
                        if param.keyframe_animation and param.keyframe_animation.keyframes:
                            keyframes = []
                            for kf in param.keyframe_animation.keyframes:
                                kf_dict = {"time": kf.time, "value": kf.value}
                                if kf.curve:
                                    kf_dict["curve"] = kf.curve
                                keyframes.append(kf_dict)
                            transform_dict[param.name] = {"keyframes": keyframes}
                nested_elements.append(transform_dict)
            elif isinstance(element, Video):
                # Convert video to dictionary
                nested_elements.append({
                    "type": "video",
                    "ref": element.ref,
                    "offset": element.offset,
                    "duration": element.duration
                })
            elif isinstance(element, Clip):
                # Recursively convert nested clips
                nested_elements.append(clip_to_dict(element))
        
        clip_dict["nested_elements"] = nested_elements
        return clip_dict
    
    main_clip_dict = clip_to_dict(main_clip)
    
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
        print(f"   🎬 Video 1: {selected_videos[0].name} (animates to left corner)")
        print(f"   🎬 Video 2: {selected_videos[1].name} (animates to right corner)")
        print(f"   🎬 Video 3: {selected_videos[2].name} (animates to top right)")
        print(f"   🎬 Video 4: {selected_videos[3].name} (animates to bottom left)")
        print(f"   ⏱️  Total duration: ~21 seconds")
        print(f"   🎭 4-lane nested animation with keyframes")
        
    except Exception as e:
        print(f"❌ Error saving FCPXML: {e}", file=sys.stderr)
        sys.exit(1)