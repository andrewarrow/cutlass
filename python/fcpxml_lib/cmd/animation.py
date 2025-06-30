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
    
    # Set proper sequence duration (like test_info_recreation.py)
    sequence.duration = convert_seconds_to_fcp_duration(20.0)  # Match clip duration
    
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
                    Keyframe(time="3003/24000s", value="0 0"),  # Match Info.fcpxml timing
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
                    Keyframe(time="3003/24000s", value="0 0"),  # Match test pattern
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

    # Create main clip using dataclasses like test_info_recreation.py
    main_clip = Clip(
        offset="0s",
        name=selected_videos[0].stem,
        duration=clip_duration,
        format=format_ids[0],  # Only main clip has format
        tc_format="NDF"
    )
    
    # Create main clip's video element
    main_video = Video(
        ref=asset_ids[0],
        offset="0s", 
        duration=video_durations[0]
    )
    
    # Create nested clips using exact pattern from test_info_recreation.py
    nested_clips = [
        {
            "lane": 1,
            "offset": second_offset,
            "name": selected_videos[1].stem, 
            "duration": nested_durations[1],
            "ref": asset_ids[1],
            "video_duration": video_durations[1],
            "transform": second_transform
        },
        {
            "lane": 2,
            "offset": third_offset,
            "name": selected_videos[2].stem,
            "duration": nested_durations[2],
            "ref": asset_ids[2],
            "video_duration": video_durations[2],
            "transform": third_transform
        },
        {
            "lane": 3,
            "offset": fourth_offset,
            "name": selected_videos[3].stem,
            "duration": nested_durations[3],
            "ref": asset_ids[3],
            "video_duration": video_durations[3],
            "transform": fourth_transform
        }
    ]
    
    # Create nested clip objects
    nested_clip_objs = []
    for i, clip_info in enumerate(nested_clips):
        nested_clip = Clip(
            lane=clip_info["lane"],
            offset=clip_info["offset"],
            name=clip_info["name"],
            duration=clip_info["duration"],
            format=format_ids[i+1],  # Add format for validation (r7, r8, r9)
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
    main_clip.adjust_transform = first_transform
    main_clip.videos = [main_video]
    main_clip.clips = nested_clip_objs
    
    # Convert to dictionary format exactly like test_info_recreation.py
    main_clip_dict = {
        "type": "clip",
        "offset": main_clip.offset,
        "name": main_clip.name,
        "duration": main_clip.duration,
        "format": main_clip.format,
        "tcFormat": main_clip.tc_format,
        "nested_elements": []
    }
    
    # Add adjust-transform as nested element (this is what serializer expects)
    transform_dict = main_clip.adjust_transform.to_dict()
    transform_dict["type"] = "adjust_transform"
    main_clip_dict["nested_elements"].append(transform_dict)
    
    # Add video element to main clip nested_elements (required for "respective media")
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
            "format": nested_clip.format,  # Include format for validation
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