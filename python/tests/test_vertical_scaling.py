"""
Tests for vertical/horizontal format scaling functionality.

Tests that scaling transforms are correctly applied in vertical mode (1080x1920)
and not applied in horizontal mode (1280x720) for both images and videos.
"""

import pytest
import tempfile
import os
from pathlib import Path

from fcpxml_lib import create_empty_project, add_media_to_timeline, save_fcpxml
from fcpxml_lib.constants import VERTICAL_SCALE_FACTOR, VERTICAL_FORMAT_WIDTH, VERTICAL_FORMAT_HEIGHT, HORIZONTAL_FORMAT_WIDTH, HORIZONTAL_FORMAT_HEIGHT
from fcpxml_lib.serialization.xml_serializer import serialize_to_xml


class TestVerticalScaling:
    """Test scaling transforms for vertical vs horizontal formats."""

    def test_create_empty_project_vertical_format(self):
        """Test that create_empty_project defaults to vertical format."""
        fcpxml = create_empty_project()
        
        # Should default to vertical format
        format_def = fcpxml.resources.formats[0]
        assert format_def.width == VERTICAL_FORMAT_WIDTH
        assert format_def.height == VERTICAL_FORMAT_HEIGHT
        assert format_def.name == "FFVideoFormat1080p2398Vertical"

    def test_create_empty_project_horizontal_format(self):
        """Test that create_empty_project with use_horizontal=True creates horizontal format."""
        fcpxml = create_empty_project(use_horizontal=True)
        
        # Should use horizontal format
        format_def = fcpxml.resources.formats[0]
        assert format_def.width == HORIZONTAL_FORMAT_WIDTH
        assert format_def.height == HORIZONTAL_FORMAT_HEIGHT
        assert format_def.name == "FFVideoFormat720p2398"

    def test_vertical_scaling_applied_to_videos(self):
        """Test that videos get scaling transforms in vertical mode."""
        fcpxml = create_empty_project(use_horizontal=False)
        
        # Create temporary video file
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as f:
            video_path = f.name
            f.write(b"dummy video content")
        
        try:
            # Add video to timeline in vertical mode
            add_media_to_timeline(fcpxml, [video_path], clip_duration_seconds=3.0, use_horizontal=False)
            
            # Check that asset-clip has scaling transform
            spine = fcpxml.library.events[0].projects[0].sequences[0].spine
            assert len(spine.ordered_elements) == 1
            
            element = spine.ordered_elements[0]
            assert element["type"] == "asset-clip"
            assert "adjust_transform" in element
            assert element["adjust_transform"]["scale"] == VERTICAL_SCALE_FACTOR
            
        finally:
            os.unlink(video_path)

    def test_vertical_scaling_applied_to_images(self):
        """Test that images get scaling transforms in vertical mode."""
        fcpxml = create_empty_project(use_horizontal=False)
        
        # Create temporary image file
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as f:
            image_path = f.name
            f.write(b"dummy image content")
        
        try:
            # Add image to timeline in vertical mode
            add_media_to_timeline(fcpxml, [image_path], clip_duration_seconds=3.0, use_horizontal=False)
            
            # Check that video element has scaling transform
            spine = fcpxml.library.events[0].projects[0].sequences[0].spine
            assert len(spine.ordered_elements) == 1
            
            element = spine.ordered_elements[0]
            assert element["type"] == "video"
            assert "adjust_transform" in element
            assert element["adjust_transform"]["scale"] == VERTICAL_SCALE_FACTOR
            
        finally:
            os.unlink(image_path)

    def test_horizontal_no_scaling_videos(self):
        """Test that videos do NOT get scaling transforms in horizontal mode."""
        fcpxml = create_empty_project(use_horizontal=True)
        
        # Create temporary video file
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as f:
            video_path = f.name
            f.write(b"dummy video content")
        
        try:
            # Add video to timeline in horizontal mode
            add_media_to_timeline(fcpxml, [video_path], clip_duration_seconds=3.0, use_horizontal=True)
            
            # Check that asset-clip has NO scaling transform
            spine = fcpxml.library.events[0].projects[0].sequences[0].spine
            assert len(spine.ordered_elements) == 1
            
            element = spine.ordered_elements[0]
            assert element["type"] == "asset-clip"
            assert "adjust_transform" not in element
            
        finally:
            os.unlink(video_path)

    def test_horizontal_no_scaling_images(self):
        """Test that images do NOT get scaling transforms in horizontal mode."""
        fcpxml = create_empty_project(use_horizontal=True)
        
        # Create temporary image file
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as f:
            image_path = f.name
            f.write(b"dummy image content")
        
        try:
            # Add image to timeline in horizontal mode
            add_media_to_timeline(fcpxml, [image_path], clip_duration_seconds=3.0, use_horizontal=True)
            
            # Check that video element has NO scaling transform
            spine = fcpxml.library.events[0].projects[0].sequences[0].spine
            assert len(spine.ordered_elements) == 1
            
            element = spine.ordered_elements[0]
            assert element["type"] == "video"
            assert "adjust_transform" not in element
            
        finally:
            os.unlink(image_path)

    def test_mixed_media_vertical_scaling(self):
        """Test that both images and videos get scaling in vertical mode."""
        fcpxml = create_empty_project(use_horizontal=False)
        
        # Create temporary files
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as video_file:
            video_path = video_file.name
            video_file.write(b"dummy video content")
        
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as image_file:
            image_path = image_file.name
            image_file.write(b"dummy image content")
        
        try:
            # Add both to timeline in vertical mode
            add_media_to_timeline(fcpxml, [video_path, image_path], clip_duration_seconds=3.0, use_horizontal=False)
            
            # Check that both elements have scaling transforms
            spine = fcpxml.library.events[0].projects[0].sequences[0].spine
            assert len(spine.ordered_elements) == 2
            
            # Both elements should have scaling
            for element in spine.ordered_elements:
                assert "adjust_transform" in element
                assert element["adjust_transform"]["scale"] == VERTICAL_SCALE_FACTOR
            
            # One should be asset-clip (video), one should be video (image)
            element_types = [el["type"] for el in spine.ordered_elements]
            assert "asset-clip" in element_types
            assert "video" in element_types
            
        finally:
            os.unlink(video_path)
            os.unlink(image_path)

    def test_mixed_media_horizontal_no_scaling(self):
        """Test that neither images nor videos get scaling in horizontal mode."""
        fcpxml = create_empty_project(use_horizontal=True)
        
        # Create temporary files
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as video_file:
            video_path = video_file.name
            video_file.write(b"dummy video content")
        
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as image_file:
            image_path = image_file.name
            image_file.write(b"dummy image content")
        
        try:
            # Add both to timeline in horizontal mode
            add_media_to_timeline(fcpxml, [video_path, image_path], clip_duration_seconds=3.0, use_horizontal=True)
            
            # Check that neither element has scaling transforms
            spine = fcpxml.library.events[0].projects[0].sequences[0].spine
            assert len(spine.ordered_elements) == 2
            
            # Neither element should have scaling
            for element in spine.ordered_elements:
                assert "adjust_transform" not in element
            
        finally:
            os.unlink(video_path)
            os.unlink(image_path)

    def test_xml_serialization_includes_transforms_vertical(self):
        """Test that XML serialization includes adjust-transform elements in vertical mode."""
        fcpxml = create_empty_project(use_horizontal=False)
        
        # Create temporary files
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as video_file:
            video_path = video_file.name
            video_file.write(b"dummy video content")
        
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as image_file:
            image_path = image_file.name
            image_file.write(b"dummy image content")
        
        try:
            # Add both to timeline in vertical mode
            add_media_to_timeline(fcpxml, [video_path, image_path], clip_duration_seconds=3.0, use_horizontal=False)
            
            # Serialize to XML
            xml_content = serialize_to_xml(fcpxml)
            
            # Check that XML contains adjust-transform elements with correct scale
            assert '<adjust-transform scale="3.27127 3.27127"/>' in xml_content
            
            # Should appear twice (once for video, once for image)
            assert xml_content.count('<adjust-transform scale="3.27127 3.27127"/>') == 2
            
            # Check that both asset-clip and video elements have transforms
            assert '<asset-clip' in xml_content
            assert '<video' in xml_content
            
        finally:
            os.unlink(video_path)
            os.unlink(image_path)

    def test_xml_serialization_no_transforms_horizontal(self):
        """Test that XML serialization does NOT include adjust-transform elements in horizontal mode."""
        fcpxml = create_empty_project(use_horizontal=True)
        
        # Create temporary files
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as video_file:
            video_path = video_file.name
            video_file.write(b"dummy video content")
        
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as image_file:
            image_path = image_file.name
            image_file.write(b"dummy image content")
        
        try:
            # Add both to timeline in horizontal mode
            add_media_to_timeline(fcpxml, [video_path, image_path], clip_duration_seconds=3.0, use_horizontal=True)
            
            # Serialize to XML
            xml_content = serialize_to_xml(fcpxml)
            
            # Check that XML does NOT contain adjust-transform elements
            assert '<adjust-transform' not in xml_content
            
            # Should still have asset-clip and video elements, just no transforms
            assert '<asset-clip' in xml_content
            assert '<video' in xml_content
            
        finally:
            os.unlink(video_path)
            os.unlink(image_path)

    def test_end_to_end_vertical_file_generation(self):
        """Test end-to-end generation of vertical FCPXML file with scaling."""
        fcpxml = create_empty_project(use_horizontal=False)
        
        # Create temporary files
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as video_file:
            video_path = video_file.name
            video_file.write(b"dummy video content")
        
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as image_file:
            image_path = image_file.name
            image_file.write(b"dummy image content")
        
        with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as fcpxml_file:
            output_path = fcpxml_file.name
        
        try:
            # Add media and save file
            add_media_to_timeline(fcpxml, [video_path, image_path], clip_duration_seconds=2.0, use_horizontal=False)
            success = save_fcpxml(fcpxml, output_path)
            
            assert success
            assert os.path.exists(output_path)
            
            # Read the generated file and check contents
            with open(output_path, 'r') as f:
                content = f.read()
            
            # Should have vertical format
            assert 'width="1080" height="1920"' in content
            
            # Should have scaling transforms
            assert '<adjust-transform scale="3.27127 3.27127"/>' in content
            assert content.count('<adjust-transform scale="3.27127 3.27127"/>') == 2
            
        finally:
            os.unlink(video_path)
            os.unlink(image_path)
            if os.path.exists(output_path):
                os.unlink(output_path)

    def test_end_to_end_horizontal_file_generation(self):
        """Test end-to-end generation of horizontal FCPXML file without scaling."""
        fcpxml = create_empty_project(use_horizontal=True)
        
        # Create temporary files
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as video_file:
            video_path = video_file.name
            video_file.write(b"dummy video content")
        
        with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as image_file:
            image_path = image_file.name
            image_file.write(b"dummy image content")
        
        with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as fcpxml_file:
            output_path = fcpxml_file.name
        
        try:
            # Add media and save file
            add_media_to_timeline(fcpxml, [video_path, image_path], clip_duration_seconds=2.0, use_horizontal=True)
            success = save_fcpxml(fcpxml, output_path)
            
            assert success
            assert os.path.exists(output_path)
            
            # Read the generated file and check contents
            with open(output_path, 'r') as f:
                content = f.read()
            
            # Should have horizontal format
            assert 'width="1280" height="720"' in content
            
            # Should NOT have scaling transforms
            assert '<adjust-transform' not in content
            
        finally:
            os.unlink(video_path)
            os.unlink(image_path)
            if os.path.exists(output_path):
                os.unlink(output_path)

    def test_scale_factor_constant(self):
        """Test that the vertical scale factor is the expected value."""
        assert VERTICAL_SCALE_FACTOR == "3.27127 3.27127"

    def test_format_dimensions_constants(self):
        """Test that format dimension constants are correct."""
        assert VERTICAL_FORMAT_WIDTH == "1080"
        assert VERTICAL_FORMAT_HEIGHT == "1920"
        assert HORIZONTAL_FORMAT_WIDTH == "1280"
        assert HORIZONTAL_FORMAT_HEIGHT == "720"