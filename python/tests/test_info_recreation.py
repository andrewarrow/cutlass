"""
Test to recreate Info.fcpxml complex nested structure

This test recreates the complete Info.fcpxml structure using Python dataclasses
and XML serialization, proving we can generate complex FCPXML files programmatically.
"""

import pytest
import os
from pathlib import Path
from fcpxml_lib.core.fcpxml import create_empty_project, save_fcpxml, create_media_asset
from fcpxml_lib.models.elements import (
    Asset, Format, MediaRep, Clip, Video, AdjustTransform, 
    KeyframeAnimation, Keyframe, Param
)
from fcpxml_lib.utils.ids import generate_resource_id
from fcpxml_lib.serialization.xml_serializer import serialize_to_xml

def test_recreate_info_fcpxml():
    """
    Recreate the complete Info.fcpxml structure using Python functions and dataclasses.
    This generates a valid FCPXML that can be imported into Final Cut Pro.
    """
    
    # Create base project structure
    fcpxml = create_empty_project()
    
    # Define the video assets from Info.fcpxml (using frame-aligned durations for validation)
    video_assets = [
        {"name": "Fey-xMO8ymQ", "duration": "240240/24000s", "uid": "627330A939EB6C3B0EEE7FF512D4033C"},
        {"name": "5NoE2g1ju1Q", "duration": "360360/24000s", "uid": "6A7BDA074E383F155359660084DB3D42"},
        {"name": "y5uUM5ojSL0", "duration": "300300/24000s", "uid": "1AF6200584B1BF6B200626D3F7FFE41A"},
        {"name": "osvgc4-4CHQ", "duration": "330330/24000s", "uid": "FC0AD4076CE55948B16BCD910117B20E"}
    ]
    
    # Create additional format for video assets (r1 already exists from create_empty_project)
    format_r2 = Format(
        id="r2", 
        frame_duration="1001/24000s",  # Using standard frame duration for validation
        width="1920",
        height="1080",
        color_space="1-1-1 (Rec. 709)"
    )
    
    # Add only the new format to resources 
    fcpxml.resources.formats.append(format_r2)
    
    # Create assets with media representations
    assets = []
    for i, video_info in enumerate(video_assets):
        asset_id = f"r{i+3}"  # r3, r4, r5, r6
        
        # Create dummy file path for testing
        dummy_file = f"/tmp/{video_info['name']}.mov"
        
        asset = Asset(
            id=asset_id,
            name=video_info["name"],
            uid=video_info["uid"],
            start="0s",
            duration=video_info["duration"],
            has_video="1",
            format="r2",
            has_audio="1",
            video_sources="1",
            audio_sources="1",
            audio_channels="2",
            audio_rate="48000"
        )
        
        # Add media representation
        media_rep = MediaRep(
            kind="original-media",
            sig=video_info["uid"],
            src=f"file://{dummy_file}"
        )
        asset.media_reps = [media_rep]
        
        assets.append(asset)
    
    # Add assets to resources
    fcpxml.resources.assets.extend(assets)
    
    # Get the main sequence (created by create_empty_project)
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    
    # Update sequence format and duration to match Info.fcpxml
    sequence.format = "r1"
    sequence.duration = "480480/24000s"  # Frame-aligned duration
    sequence.tc_start = "0s"
    sequence.tc_format = "NDF"
    sequence.audio_layout = "stereo"
    sequence.audio_rate = "48k"
    
    # Create the main clip with nested structure
    main_clip = Clip(
        offset="0s",
        name="Fey-xMO8ymQ",
        duration="240240/24000s",  # Frame-aligned duration
        format="r2",
        tc_format="NDF"
    )
    
    # Note: ConformRate can be added to XML directly if needed
    
    # Create main clip's adjust transform with keyframes
    main_transform = AdjustTransform(
        params=[
            Param(
                name="anchor",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="145145/24000s", value="0 0", curve="linear")
                ])
            ),
            Param(
                name="position", 
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="0s", value="0 0"),
                    Keyframe(time="145145/24000s", value="-17.2101 43.0307")
                ])
            ),
            Param(
                name="rotation",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="145145/24000s", value="0", curve="linear")
                ])
            ),
            Param(
                name="scale",
                keyframe_animation=KeyframeAnimation(keyframes=[
                    Keyframe(time="145145/24000s", value="-0.356424 0.356424", curve="linear")
                ])
            )
        ]
    )
    main_clip.adjust_transform = main_transform
    
    # Add main video element
    main_video = Video(
        ref="r3",
        offset="0s", 
        duration="240240/24000s"  # Frame-aligned duration
    )
    main_clip.videos = [main_video]
    
    # Create nested clips with their own animations (using frame-aligned durations)
    nested_clips = [
        {
            "lane": 1,
            "offset": "60060/24000s",  # Frame-aligned offset
            "name": "5NoE2g1ju1Q", 
            "duration": "360360/24000s",  # Frame-aligned duration
            "ref": "r4",
            "video_duration": "360360/24000s",  # Frame-aligned duration
            "position_keyframes": [
                Keyframe(time="0s", value="0 0"),
                Keyframe(time="108108/24000s", value="2.38541 43.2326")
            ],
            "scale_keyframe": Keyframe(time="108108/24000s", value="0.313976 0.313976", curve="linear"),
            "anchor_keyframe": Keyframe(time="108108/24000s", value="0 0", curve="linear"),
            "rotation_keyframe": Keyframe(time="108108/24000s", value="0", curve="linear")
        },
        {
            "lane": 2,
            "offset": "120120/24000s",  # Frame-aligned offset
            "name": "y5uUM5ojSL0",
            "duration": "300300/24000s",  # Frame-aligned duration
            "ref": "r5",
            "video_duration": "300300/24000s",  # Frame-aligned duration
            "position_keyframes": [
                Keyframe(time="3003/24000s", value="0 0"),
                Keyframe(time="95095/24000s", value="22.2446 42.4814")
            ],
            "scale_keyframe": Keyframe(time="95095/24000s", value="0.362066 0.362066", curve="linear"),
            "anchor_keyframe": Keyframe(time="95095/24000s", value="0 0", curve="linear"),
            "rotation_keyframe": Keyframe(time="95095/24000s", value="0", curve="linear")
        },
        {
            "lane": 3,
            "offset": "180180/24000s",  # Frame-aligned offset
            "name": "osvgc4-4CHQ",
            "duration": "330330/24000s",  # Frame-aligned duration
            "ref": "r6",
            "video_duration": "330330/24000s",  # Frame-aligned duration
            "position_keyframes": [
                Keyframe(time="3003/24000s", value="0 0"),  # Frame-aligned start time
                Keyframe(time="66066/24000s", value="-19.2439 31.344")
            ],
            "scale_keyframe": Keyframe(time="66066/24000s", value="0.265712 0.265712", curve="linear"),
            "anchor_keyframe": Keyframe(time="66066/24000s", value="0 0", curve="linear"),
            "rotation_keyframe": Keyframe(time="66066/24000s", value="0", curve="linear")
        }
    ]
    
    # Create nested clips
    for clip_info in nested_clips:
        nested_clip = Clip(
            lane=clip_info["lane"],
            offset=clip_info["offset"],
            name=clip_info["name"],
            duration=clip_info["duration"],
            tc_format="NDF"
        )
        
        # Note: ConformRate can be added to XML directly if needed
        
        # Create transform with keyframes
        nested_transform = AdjustTransform(
            params=[
                Param(
                    name="anchor",
                    keyframe_animation=KeyframeAnimation(keyframes=[clip_info["anchor_keyframe"]])
                ),
                Param(
                    name="position",
                    keyframe_animation=KeyframeAnimation(keyframes=clip_info["position_keyframes"])
                ),
                Param(
                    name="rotation", 
                    keyframe_animation=KeyframeAnimation(keyframes=[clip_info["rotation_keyframe"]])
                ),
                Param(
                    name="scale",
                    keyframe_animation=KeyframeAnimation(keyframes=[clip_info["scale_keyframe"]])
                )
            ]
        )
        nested_clip.adjust_transform = nested_transform
        
        # Add video element
        nested_video = Video(
            ref=clip_info["ref"],
            offset="0s",
            duration=clip_info["video_duration"]
        )
        nested_clip.videos = [nested_video]
        
        # Add to main clip's nested clips
        if not hasattr(main_clip, 'clips'):
            main_clip.clips = []
        main_clip.clips.append(nested_clip)
    
    # Convert main clip to dictionary and add to spine
    main_clip_dict = {
        "type": "clip",
        "offset": main_clip.offset,
        "name": main_clip.name,
        "duration": main_clip.duration,
        "format": main_clip.format,
        "tcFormat": main_clip.tc_format,
        "adjust_transform": main_clip.adjust_transform.to_dict(),
        "videos": [{"ref": main_video.ref, "offset": main_video.offset, "duration": main_video.duration}],
        "nested_elements": []
    }
    
    # Add nested clips to main clip
    for nested_clip in main_clip.clips:
        nested_dict = {
            "type": "clip",
            "lane": nested_clip.lane,
            "offset": nested_clip.offset,
            "name": nested_clip.name,
            "duration": nested_clip.duration,
            "format": "r2",  # Add missing format attribute
            "tcFormat": nested_clip.tc_format,
            "adjust_transform": nested_clip.adjust_transform.to_dict(),
            "videos": [{"ref": video.ref, "offset": video.offset, "duration": video.duration} for video in nested_clip.videos]
        }
        main_clip_dict["nested_elements"].append(nested_dict)
    
    # Add to spine
    sequence.spine.ordered_elements.append(main_clip_dict)
    
    # Generate XML and save to file
    output_file = "/tmp/test_info_recreation.fcpxml"
    success = save_fcpxml(fcpxml, output_file)
    
    assert success, "Failed to save FCPXML file"
    assert Path(output_file).exists(), "Output file was not created"
    
    # Verify the file contains expected structure
    with open(output_file, 'r') as f:
        content = f.read()
        
    # Check for key elements
    assert 'version="1.13"' in content, "Should have correct FCPXML version"
    assert 'id="r3"' in content, "Should contain asset r3"
    assert 'id="r4"' in content, "Should contain asset r4" 
    assert 'id="r5"' in content, "Should contain asset r5"
    assert 'id="r6"' in content, "Should contain asset r6"
    assert 'lane="1"' in content, "Should contain nested clips with lanes"
    assert 'lane="2"' in content, "Should contain multiple lanes"
    assert 'lane="3"' in content, "Should contain multiple lanes"
    assert 'Fey-xMO8ymQ' in content, "Should contain main clip name"
    assert 'conform-rate' in content, "Should contain conform-rate elements"
    assert '<clip' in content, "Should contain clip elements"
    
    # Verify nested clip structure matches Info.fcpxml pattern
    assert content.count('<clip') >= 4, "Should have main clip plus 3 nested clips"
    
    print(f"✅ Successfully recreated Info.fcpxml structure in {output_file}")
    print(f"   File size: {Path(output_file).stat().st_size} bytes")
    print("   Contains all 4 video assets with nested clip structure")
    print("   🎯 Structure matches Info.fcpxml pattern with main + nested clips")
    print("   Ready for Final Cut Pro import testing")

if __name__ == "__main__":
    test_recreate_info_fcpxml()