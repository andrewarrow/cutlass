"""
Test conform-rate element validation and FCP import compatibility.

This test ensures that conform-rate elements include the required srcFrameRate
attribute to prevent Final Cut Pro import warnings.
"""

import pytest
from pathlib import Path
from fcpxml_lib.core.fcpxml import create_empty_project, save_fcpxml
from fcpxml_lib.models.elements import (
    Asset, Format, MediaRep, Clip, Video, AdjustTransform, 
    KeyframeAnimation, Keyframe, Param
)

def test_conform_rate_includes_src_frame_rate():
    """
    Test that generated FCPXML includes srcFrameRate attribute on conform-rate elements.
    
    This prevents FCP import warnings:
    "Encountered an unexpected value. (conform-rate: /fcpxml[1]/...)"
    """
    
    # Create base project structure
    fcpxml = create_empty_project()
    
    # Create a simple clip structure that would generate conform-rate elements
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    
    # Create test clip with video element
    test_clip_dict = {
        "type": "clip",
        "offset": "0s",
        "name": "TestClip",
        "duration": "240240/24000s",
        "format": "r1",
        "tcFormat": "NDF",
        "nested_elements": [{
            "type": "video",
            "ref": "r1", 
            "offset": "0s",
            "duration": "240240/24000s"
        }]
    }
    
    # Add clip to spine
    sequence.spine.ordered_elements.append(test_clip_dict)
    
    # Generate XML
    output_file = "/tmp/test_conform_rate.fcpxml"
    success = save_fcpxml(fcpxml, output_file)
    
    # Read generated XML content
    with open(output_file, 'r') as f:
        content = f.read()
    
    # Verify conform-rate elements include srcFrameRate attribute
    assert 'conform-rate' in content, "Should contain conform-rate elements"
    assert 'srcFrameRate=' in content, "conform-rate elements must include srcFrameRate attribute"
    assert 'srcFrameRate="24"' in content, "Should include standard 24fps frame rate"
    
    # Verify the full conform-rate structure
    assert '<conform-rate scaleEnabled="0" srcFrameRate="24"/>' in content, \
        "conform-rate should have both scaleEnabled and srcFrameRate attributes"
    
    print("✅ conform-rate elements include required srcFrameRate attribute")
    print("   This prevents FCP import warnings about unexpected conform-rate values")

def test_nested_clips_conform_rate_attributes():
    """
    Test that nested clips also get proper conform-rate elements with srcFrameRate.
    
    Nested clips were specifically mentioned in the FCP error messages.
    """
    
    # Create base project structure
    fcpxml = create_empty_project()
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    
    # Create main clip with nested clips (multi-lane structure)
    main_clip_dict = {
        "type": "clip",
        "offset": "0s",
        "name": "MainClip",
        "duration": "240240/24000s",
        "format": "r1",
        "tcFormat": "NDF",
        "nested_elements": [
            # Main video element
            {
                "type": "video",
                "ref": "r1",
                "offset": "0s", 
                "duration": "240240/24000s"
            },
            # Nested clip on lane 1
            {
                "type": "clip",
                "lane": 1,
                "offset": "60060/24000s",
                "name": "NestedClip1", 
                "duration": "120120/24000s",
                "tcFormat": "NDF",
                "nested_elements": [{
                    "type": "video",
                    "ref": "r1",
                    "offset": "0s",
                    "duration": "120120/24000s"
                }]
            },
            # Nested clip on lane 2
            {
                "type": "clip", 
                "lane": 2,
                "offset": "120120/24000s",
                "name": "NestedClip2",
                "duration": "120120/24000s", 
                "tcFormat": "NDF",
                "nested_elements": [{
                    "type": "video",
                    "ref": "r1",
                    "offset": "0s",
                    "duration": "120120/24000s"
                }]
            }
        ]
    }
    
    # Add to spine
    sequence.spine.ordered_elements.append(main_clip_dict)
    
    # Generate XML
    output_file = "/tmp/test_nested_conform_rate.fcpxml"
    save_fcpxml(fcpxml, output_file)
    
    # Read and verify content
    with open(output_file, 'r') as f:
        content = f.read()
    
    # Count conform-rate elements - should have one for main clip + two for nested clips
    conform_rate_count = content.count('<conform-rate')
    assert conform_rate_count >= 3, f"Should have at least 3 conform-rate elements, found {conform_rate_count}"
    
    # Verify all conform-rate elements have srcFrameRate
    srcFrameRate_count = content.count('srcFrameRate="24"')
    assert srcFrameRate_count >= 3, f"All conform-rate elements should have srcFrameRate, found {srcFrameRate_count}"
    
    # Verify lane structure exists
    assert 'lane="1"' in content, "Should contain nested clip on lane 1"
    assert 'lane="2"' in content, "Should contain nested clip on lane 2"
    
    print("✅ Nested clips include proper conform-rate elements with srcFrameRate")
    print(f"   Found {conform_rate_count} conform-rate elements with srcFrameRate attributes")

def test_info_recreation_conform_rate_structure():
    """
    Test that the Info.fcpxml recreation includes proper conform-rate elements.
    
    This verifies the fix for the specific FCP import warnings seen in the user screenshot.
    """
    
    # Run the Info recreation test
    from tests.test_info_recreation import test_recreate_info_fcpxml
    test_recreate_info_fcpxml()
    
    # Read the generated file
    with open("test_info_recreation.fcpxml", 'r') as f:
        content = f.read()
    
    # Verify conform-rate structure matches expectations
    expected_patterns = [
        # Main clip conform-rate
        '<conform-rate scaleEnabled="0" srcFrameRate="24"/>',
        # Multiple conform-rate elements for nested clips
        'lane="1"',
        'lane="2"', 
        'lane="3"'
    ]
    
    for pattern in expected_patterns:
        assert pattern in content, f"Should contain: {pattern}"
    
    # Count total conform-rate elements (main + 3 nested clips = 4 total)
    conform_rate_count = content.count('<conform-rate')
    assert conform_rate_count == 4, f"Should have exactly 4 conform-rate elements, found {conform_rate_count}"
    
    # Verify all have srcFrameRate
    srcFrameRate_count = content.count('srcFrameRate="24"') 
    assert srcFrameRate_count == 4, f"All 4 conform-rate elements should have srcFrameRate, found {srcFrameRate_count}"
    
    print("✅ Info.fcpxml recreation includes proper conform-rate elements")
    print("   This should resolve FCP import warnings shown in user screenshot")

def test_conform_rate_different_frame_rates():
    """
    Test conform-rate elements with different frame rates (future enhancement).
    
    Currently hardcoded to 24fps, but this documents the expected behavior
    for when frame rate detection is implemented.
    """
    
    # This is a placeholder test for future frame rate detection
    # Currently our serializer hardcodes srcFrameRate="24"
    
    fcpxml = create_empty_project()
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    
    test_clip = {
        "type": "clip",
        "offset": "0s", 
        "name": "TestClip",
        "duration": "240240/24000s",
        "format": "r1",
        "tcFormat": "NDF",
        "nested_elements": [{
            "type": "video",
            "ref": "r1",
            "offset": "0s",
            "duration": "240240/24000s"
        }]
    }
    
    sequence.spine.ordered_elements.append(test_clip)
    
    output_file = "/tmp/test_frame_rates.fcpxml"
    save_fcpxml(fcpxml, output_file)
    
    with open(output_file, 'r') as f:
        content = f.read()
    
    # Currently expects 24fps (hardcoded)
    assert 'srcFrameRate="24"' in content, "Currently hardcoded to 24fps"
    
    print("✅ Frame rate handling documented for future enhancement") 
    print("   TODO: Implement dynamic frame rate detection from media properties")

if __name__ == "__main__":
    test_conform_rate_includes_src_frame_rate()
    test_nested_clips_conform_rate_attributes()
    test_info_recreation_conform_rate_structure()
    test_conform_rate_different_frame_rates()
    print("\n🎉 All conform-rate validation tests passed!")