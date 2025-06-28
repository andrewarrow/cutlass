"""
Tests for video-at-edge functionality with nested PNG lane structure.

This test suite validates the Pattern A vs Pattern B FCPXML generation approaches
discovered during the multi-lane visibility regression investigation.
"""

import pytest
import tempfile
import os
from pathlib import Path
from unittest.mock import patch, MagicMock

from fcpxml_lib import create_empty_project, save_fcpxml
from fcpxml_lib.serialization.xml_serializer import serialize_to_xml
from main import create_edge_tiled_timeline


class TestVideoAtEdge:
    """Test suite for video-at-edge functionality with Pattern A nested structure."""
    
    @pytest.fixture
    def temp_png_files(self):
        """Create temporary PNG files for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create mock PNG files
            png_files = []
            for i in range(5):
                png_file = temp_path / f"test_{i}.png"
                png_file.write_bytes(b"fake_png_data")  # Minimal fake PNG
                png_files.append(png_file)
            
            yield png_files

    @pytest.fixture
    def temp_video_file(self):
        """Create a temporary video file for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            video_file = temp_path / "background.mp4"
            video_file.write_bytes(b"fake_video_data")  # Minimal fake video
            yield video_file

    @pytest.fixture
    def base_fcpxml(self):
        """Create a base FCPXML project for testing."""
        return create_empty_project(
            project_name="Test Video at Edge",
            event_name="Test Edge Tiled Videos",
            use_horizontal=False  # Always vertical for edge detection
        )

    def test_pattern_a_nested_structure_with_background(self, base_fcpxml, temp_png_files, temp_video_file):
        """
        Test Pattern A: Nested PNG elements inside background AssetClip.
        
        This is the correct pattern that creates multiple visible lanes in FCP.
        """
        with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
             patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
            
            # Mock the background video asset creation
            mock_asset = MagicMock()
            mock_asset.id = "r5"  # Adjusted for existing assets in base_fcpxml
            mock_format = MagicMock()
            mock_format.id = "r6"
            mock_create_asset.return_value = (mock_asset, mock_format)
            
            # Mock video properties detection to avoid ffprobe dependency
            mock_detect.return_value = {
                'width': 1920,
                'height': 1080,
                'duration_seconds': 10.0,
                'has_audio': True,
                'frame_rate': '29.97'
            }
            
            with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=True):
                # Test with background video (Pattern A)
                create_edge_tiled_timeline(
                    base_fcpxml, 
                    temp_png_files[:3],  # Use 3 PNGs
                    str(temp_video_file), 
                    duration=10.0,
                    num_lanes=3,
                    tiles_per_lane=1
                )
                
                sequence = base_fcpxml.library.events[0].projects[0].sequences[0]
                
                # Verify background element exists
                assert len(sequence.spine.ordered_elements) == 1
                bg_element = sequence.spine.ordered_elements[0]
                
                # Verify background is asset-clip type
                assert bg_element["type"] == "asset-clip"
                # The asset ID is determined by resource counter logic
                assert bg_element["ref"].startswith("r")  # Should be a valid resource ID
                
                # Verify nested PNGs inside background (Pattern A)
                assert "nested_elements" in bg_element
                nested_pngs = bg_element["nested_elements"]
                assert len(nested_pngs) == 3  # 3 lanes × 1 tile = 3 PNGs
                
                # Verify each PNG has correct lane assignment
                for i, png in enumerate(nested_pngs):
                    assert png["type"] == "video"
                    assert png["lane"] == i + 1  # Sequential lanes 1, 2, 3
                    assert png["start"] == "3600s"  # Proper timing like Info.fcpxml
                    assert "adjust_transform" in png
                    assert "position" in png["adjust_transform"]
                    assert "scale" in png["adjust_transform"]

    def test_pattern_b_fallback_without_background(self, base_fcpxml, temp_png_files):
        """
        Test Pattern B: Separate spine elements (fallback when no background).
        
        This pattern works but creates separate timeline elements instead of visible lanes.
        """
        # Test without background video (Pattern B fallback)
        create_edge_tiled_timeline(
            base_fcpxml, 
            temp_png_files[:2],  # Use 2 PNGs
            background_video=None,  # No background
            duration=5.0,
            num_lanes=2,
            tiles_per_lane=1
        )
        
        sequence = base_fcpxml.library.events[0].projects[0].sequences[0]
        
        # Verify separate spine elements (Pattern B)
        assert len(sequence.spine.videos) == 2  # 2 separate video elements
        assert len(sequence.spine.ordered_elements) == 2
        
        # Verify each PNG is a separate spine element
        for i, video in enumerate(sequence.spine.videos):
            assert video["type"] == "video"
            assert video["lane"] == i + 1  # Sequential lanes 1, 2
            assert video["start"] == "0s"  # Different timing for separate elements
            assert "adjust_transform" in video

    def test_lane_numbering_sequential(self, base_fcpxml, temp_png_files, temp_video_file):
        """Test that lane numbers are assigned sequentially starting from 1."""
        with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
             patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
            
            mock_asset = MagicMock()
            mock_asset.id = "r5"
            mock_format = MagicMock()
            mock_format.id = "r6"
            mock_create_asset.return_value = (mock_asset, mock_format)
            
            mock_detect.return_value = {
                'width': 1920, 'height': 1080, 'duration_seconds': 8.0,
                'has_audio': True, 'frame_rate': '29.97'
            }
            
            with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=False):
                # Test with 5 lanes × 2 tiles = 10 total PNGs
                create_edge_tiled_timeline(
                    base_fcpxml, 
                    temp_png_files,  # Use all 5 PNGs
                    str(temp_video_file), 
                    duration=8.0,
                    num_lanes=5,
                    tiles_per_lane=2
                )
                
                sequence = base_fcpxml.library.events[0].projects[0].sequences[0]
                bg_element = sequence.spine.ordered_elements[0]
                nested_pngs = bg_element["nested_elements"]
                
                # Verify 10 total PNGs with sequential lane numbers
                assert len(nested_pngs) == 10
                
                # Check lane numbering is sequential 1-10
                lane_numbers = [png["lane"] for png in nested_pngs]
                assert lane_numbers == list(range(1, 11))

    def test_timing_matches_info_fcpxml_pattern(self, base_fcpxml, temp_png_files, temp_video_file):
        """Test that timing follows Info.fcpxml pattern with start='3600s'."""
        with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
             patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
            
            mock_asset = MagicMock()
            mock_asset.id = "r5"
            mock_format = MagicMock()
            mock_format.id = "r6"
            mock_create_asset.return_value = (mock_asset, mock_format)
            
            mock_detect.return_value = {
                'width': 1920, 'height': 1080, 'duration_seconds': 12.0,
                'has_audio': True, 'frame_rate': '29.97'
            }
            
            with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=False):
                create_edge_tiled_timeline(
                    base_fcpxml, 
                    temp_png_files[:1], 
                    str(temp_video_file), 
                    duration=12.0,
                    num_lanes=1,
                    tiles_per_lane=1
                )
                
                sequence = base_fcpxml.library.events[0].projects[0].sequences[0]
                bg_element = sequence.spine.ordered_elements[0]
                
                # Verify background timing
                assert bg_element["offset"] == "0s"
                
                # Verify PNG timing matches Info.fcpxml pattern
                nested_png = bg_element["nested_elements"][0]
                assert nested_png["offset"] == "0s"  # Relative to background start
                assert nested_png["start"] == "3600s"  # Like Info.fcpxml

    def test_xml_serialization_validity(self, base_fcpxml, temp_png_files, temp_video_file):
        """Test that the generated structure serializes to valid XML."""
        with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
             patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
            
            mock_asset = MagicMock()
            mock_asset.id = "r5"
            mock_format = MagicMock()
            mock_format.id = "r6"
            mock_create_asset.return_value = (mock_asset, mock_format)
            
            mock_detect.return_value = {
                'width': 1920, 'height': 1080, 'duration_seconds': 6.0,
                'has_audio': True, 'frame_rate': '29.97'
            }
            
            with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=False):
                create_edge_tiled_timeline(
                    base_fcpxml, 
                    temp_png_files[:2], 
                    str(temp_video_file), 
                    duration=6.0,
                    num_lanes=2,
                    tiles_per_lane=1
                )
                
                # Test XML serialization doesn't raise exceptions
                xml_content = serialize_to_xml(base_fcpxml)
                assert xml_content is not None
                assert isinstance(xml_content, str)
                
                # Verify nested structure appears in XML
                assert "asset-clip" in xml_content
                assert "lane=" in xml_content
                assert "start=\"3600s\"" in xml_content

    def test_random_positioning_within_bounds(self, base_fcpxml, temp_png_files, temp_video_file):
        """Test that PNG positions are within expected screen bounds."""
        with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
             patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
            
            mock_asset = MagicMock()
            mock_asset.id = "r5"
            mock_format = MagicMock()
            mock_format.id = "r6"
            mock_create_asset.return_value = (mock_asset, mock_format)
            
            mock_detect.return_value = {
                'width': 1920, 'height': 1080, 'duration_seconds': 4.0,
                'has_audio': True, 'frame_rate': '29.97'
            }
            
            with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=False):
                create_edge_tiled_timeline(
                    base_fcpxml, 
                    temp_png_files[:3], 
                    str(temp_video_file), 
                    duration=4.0,
                    num_lanes=3,
                    tiles_per_lane=1
                )
                
                sequence = base_fcpxml.library.events[0].projects[0].sequences[0]
                bg_element = sequence.spine.ordered_elements[0]
                nested_pngs = bg_element["nested_elements"]
                
                # Verify all positions are within expected bounds
                for png in nested_pngs:
                    position = png["adjust_transform"]["position"]
                    x, y = map(float, position.split())
                    
                    # Check bounds match create_edge_tiled_timeline implementation
                    assert -30.0 <= x <= 30.0  # X range
                    assert -50.0 <= y <= 50.0  # Y range

    def test_scale_randomization(self, base_fcpxml, temp_png_files, temp_video_file):
        """Test that PNG scales are randomized within expected range."""
        with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
             patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
            
            mock_asset = MagicMock()
            mock_asset.id = "r5"
            mock_format = MagicMock()
            mock_format.id = "r6"
            mock_create_asset.return_value = (mock_asset, mock_format)
            
            mock_detect.return_value = {
                'width': 1920, 'height': 1080, 'duration_seconds': 3.0,
                'has_audio': True, 'frame_rate': '29.97'
            }
            
            with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=False):
                create_edge_tiled_timeline(
                    base_fcpxml, 
                    temp_png_files[:2], 
                    str(temp_video_file), 
                    duration=3.0,
                    num_lanes=2,
                    tiles_per_lane=1
                )
                
                sequence = base_fcpxml.library.events[0].projects[0].sequences[0]
                bg_element = sequence.spine.ordered_elements[0]
                nested_pngs = bg_element["nested_elements"]
                
                # Verify all scales are within expected range
                for png in nested_pngs:
                    scale = png["adjust_transform"]["scale"]
                    scale_x, scale_y = map(float, scale.split())
                    
                    # Check scale range matches implementation (0.1 to 0.5)
                    assert 0.1 <= scale_x <= 0.5
                    assert 0.1 <= scale_y <= 0.5
                    assert scale_x == scale_y  # Uniform scaling


    def test_jpg_file_support(self, base_fcpxml, temp_video_file):
        """Test that JPG files are supported in addition to PNG files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Create mock JPG files
            jpg_files = []
            for i in range(3):
                jpg_file = temp_path / f"test_{i}.jpg"
                jpg_file.write_bytes(b"fake_jpg_data")
                jpg_files.append(jpg_file)
            
            with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
                 patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
                
                mock_asset = MagicMock()
                mock_asset.id = "r5"
                mock_format = MagicMock()
                mock_format.id = "r6"
                mock_create_asset.return_value = (mock_asset, mock_format)
                
                mock_detect.return_value = {
                    'width': 1920, 'height': 1080, 'duration_seconds': 8.0,
                    'has_audio': True, 'frame_rate': '29.97'
                }
                
                with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=False):
                    # Test with JPG files
                    create_edge_tiled_timeline(
                        base_fcpxml, 
                        jpg_files, 
                        str(temp_video_file), 
                        duration=6.0,
                        num_lanes=3,
                        tiles_per_lane=1
                    )
                    
                    sequence = base_fcpxml.library.events[0].projects[0].sequences[0]
                    bg_element = sequence.spine.ordered_elements[0]
                    
                    # Verify JPG files created nested structure
                    assert "nested_elements" in bg_element
                    nested_images = bg_element["nested_elements"]
                    assert len(nested_images) == 3
                    
                    # Verify each image has correct properties
                    for i, image in enumerate(nested_images):
                        assert image["type"] == "video"
                        assert image["lane"] == i + 1
                        assert "test_" in image["name"]  # Should contain the JPG filename


class TestPatternComparison:
    """Tests comparing Pattern A vs Pattern B approaches."""
    
    def test_pattern_a_creates_nested_structure(self):
        """Verify Pattern A creates nested elements inside asset-clip."""
        fcpxml = create_empty_project()
        
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            temp_png = temp_path / "test.png"
            temp_png.write_bytes(b"fake_png_data")
            temp_video = temp_path / "test.mp4"
            temp_video.write_bytes(b"fake_video_data")
            
            with patch('fcpxml_lib.core.fcpxml.create_media_asset') as mock_create_asset, \
                 patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
                
                mock_asset = MagicMock()
                mock_asset.id = "r5"  # Adjusted for existing assets
                mock_format = MagicMock()
                mock_format.id = "r6"
                mock_create_asset.return_value = (mock_asset, mock_format)
                
                # Mock video properties detection
                mock_detect.return_value = {
                    'width': 1920,
                    'height': 1080,
                    'duration_seconds': 5.0,
                    'has_audio': True,
                    'frame_rate': '29.97'
                }
                
                with patch('fcpxml_lib.core.fcpxml.needs_vertical_scaling', return_value=False):
                    # Pattern A: With background video
                    create_edge_tiled_timeline(
                        fcpxml, [temp_png], str(temp_video), 5.0, 1, 1
                    )
                    
                    sequence = fcpxml.library.events[0].projects[0].sequences[0]
                    
                    # Pattern A characteristics
                    assert len(sequence.spine.ordered_elements) == 1  # Single background
                    bg_element = sequence.spine.ordered_elements[0]
                    assert bg_element["type"] == "asset-clip"
                    assert "nested_elements" in bg_element
                    assert len(bg_element["nested_elements"]) == 1

    def test_pattern_b_creates_separate_elements(self):
        """Verify Pattern B creates separate spine elements."""
        fcpxml = create_empty_project()
        temp_png = Path("/tmp/test.png")
        
        # Pattern B: No background video
        create_edge_tiled_timeline(
            fcpxml, [temp_png], background_video=None, duration=5.0, num_lanes=1, tiles_per_lane=1
        )
        
        sequence = fcpxml.library.events[0].projects[0].sequences[0]
        
        # Pattern B characteristics
        assert len(sequence.spine.videos) == 1  # Separate video element
        assert len(sequence.spine.ordered_elements) == 1
        video_element = sequence.spine.videos[0]
        assert video_element["type"] == "video"
        assert video_element["start"] == "0s"  # Different timing