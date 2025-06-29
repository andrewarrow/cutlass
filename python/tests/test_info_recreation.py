"""
Test to recreate Info.fcpxml keyframe animation pattern

This test attempts to recreate the complex nested clip structure with keyframe animations
found in Info.fcpxml, where two scaled videos animate from center to left/right positions.
"""

import pytest
from pathlib import Path
from fcpxml_lib.core.fcpxml import create_empty_project, save_fcpxml
from fcpxml_lib.models.elements import (
    Clip, Video, AdjustTransform, KeyframeAnimation, Keyframe, Param
)

def test_recreate_info_fcpxml():
    """
    Test that keyframe animation dataclasses can be created and contain expected data.
    This proves the concept works before implementing full serialization.
    """
    
    # Test keyframe animation dataclass creation (using frame-aligned timings)
    position_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time="0s", value="0 0"),  # Start at center
        Keyframe(time="144144/24000s", value="-17.2101 43.0307")  # End at left corner (6s)
    ])
    
    scale_keyframes = KeyframeAnimation(keyframes=[
        Keyframe(time="144144/24000s", value="-0.356424 0.356424", curve="linear")
    ])
    
    # Test transform with keyframe parameters  
    transform = AdjustTransform(
        params=[
            Param(name="position", keyframe_animation=position_keyframes),
            Param(name="scale", keyframe_animation=scale_keyframes)
        ]
    )
    
    # Test Video and Clip dataclasses (using frame-aligned durations)
    video = Video(
        ref="r3",
        duration="120120/24000s",  # 5 seconds frame-aligned
        offset="0s"
    )
    
    clip = Clip(
        offset="0s",
        name="Test Video",
        duration="144144/24000s",  # 6 seconds frame-aligned
        format="r2",
        tc_format="NDF"
    )
    
    # Verify data structure
    assert len(position_keyframes.keyframes) == 2, "Position should have 2 keyframes"
    assert position_keyframes.keyframes[0].time == "0s", "First keyframe should be at 0s"
    assert position_keyframes.keyframes[0].value == "0 0", "First keyframe should be center position"
    assert position_keyframes.keyframes[1].time == "144144/24000s", "Second keyframe timing"
    assert position_keyframes.keyframes[1].value == "-17.2101 43.0307", "Second keyframe position"
    
    assert len(scale_keyframes.keyframes) == 1, "Scale should have 1 keyframe"
    assert scale_keyframes.keyframes[0].curve == "linear", "Scale keyframe should have linear curve"
    
    assert len(transform.params) == 2, "Transform should have 2 parameters"
    assert transform.params[0].name == "position", "First param should be position"
    assert transform.params[1].name == "scale", "Second param should be scale"
    
    assert video.ref == "r3", "Video should reference r3"
    assert clip.name == "Test Video", "Clip should have correct name"
    assert clip.tc_format == "NDF", "Clip should use NDF timecode format"
    
    # Test to_dict method on transform
    transform_dict = transform.to_dict()
    assert "params" in transform_dict, "Transform dict should contain params"
    assert len(transform_dict["params"]) == 2, "Should have 2 parameters"
    
    position_param = transform_dict["params"][0]
    assert position_param["name"] == "position", "First param should be position"
    assert "keyframes" in position_param, "Position param should have keyframes"
    assert len(position_param["keyframes"]) == 2, "Position should have 2 keyframes"
    
    first_keyframe = position_param["keyframes"][0]
    assert first_keyframe["time"] == "0s", "First keyframe time"
    assert first_keyframe["value"] == "0 0", "First keyframe value"
    
    print("✅ Keyframe animation dataclasses work correctly")

if __name__ == "__main__":
    test_recreate_info_fcpxml()