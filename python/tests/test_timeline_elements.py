"""
Tests for timeline element generation and structure.

Tests the correct separation of images (video elements) vs videos (asset-clip elements)
and proper timeline construction.
"""

import pytest
import tempfile
import os
from xml.etree.ElementTree import fromstring

from fcpxml_lib.core.fcpxml import create_empty_fcpxml, add_media_to_timeline
from fcpxml_lib.serialization.xml_serializer import serialize_to_xml


class TestTimelineElements:
    """Test timeline element creation and structure."""

    @pytest.fixture
    def sample_media_files(self):
        """Create temporary media files for testing."""
        files = {}
        
        # Create temporary image file
        with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp_image:
            tmp_image.write(b'fake image content')
            files['image'] = tmp_image.name
        
        # Create temporary video file
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as tmp_video:
            tmp_video.write(b'fake video content')
            files['video'] = tmp_video.name
        
        yield files
        
        # Cleanup
        for file_path in files.values():
            if os.path.exists(file_path):
                os.unlink(file_path)

    def test_images_create_video_elements(self, sample_media_files):
        """Test that image files create <video> elements in the timeline."""
        fcpxml = create_empty_fcpxml()
        add_media_to_timeline(fcpxml, [sample_media_files['image']], clip_duration_seconds=5.0)
        
        xml_content = serialize_to_xml(fcpxml)
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Find video elements in spine
        video_elements = root.findall('.//spine/video')
        assert len(video_elements) == 1, "Should have exactly one video element for image"
        
        video_elem = video_elements[0]
        assert video_elem.get('ref') is not None
        assert video_elem.get('duration') is not None
        assert video_elem.get('start') is not None  # Images need start attribute
        assert video_elem.get('offset') is not None

    def test_videos_create_asset_clip_elements(self, sample_media_files):
        """Test that video files create <asset-clip> elements in the timeline."""
        fcpxml = create_empty_fcpxml()
        add_media_to_timeline(fcpxml, [sample_media_files['video']], clip_duration_seconds=5.0)
        
        xml_content = serialize_to_xml(fcpxml)
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Find asset-clip elements in spine
        asset_clip_elements = root.findall('.//spine/asset-clip')
        assert len(asset_clip_elements) == 1, "Should have exactly one asset-clip element for video"
        
        asset_clip_elem = asset_clip_elements[0]
        assert asset_clip_elem.get('ref') is not None
        assert asset_clip_elem.get('duration') is not None
        assert asset_clip_elem.get('offset') is not None
        # Asset-clip elements should NOT have start attribute

    def test_mixed_media_timeline(self, sample_media_files):
        """Test timeline with both images and videos."""
        fcpxml = create_empty_fcpxml()
        media_files = [sample_media_files['image'], sample_media_files['video']]
        add_media_to_timeline(fcpxml, media_files, clip_duration_seconds=3.0)
        
        xml_content = serialize_to_xml(fcpxml)
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Should have both video and asset-clip elements
        video_elements = root.findall('.//spine/video')
        asset_clip_elements = root.findall('.//spine/asset-clip')
        
        assert len(video_elements) == 1, "Should have one video element for image"
        assert len(asset_clip_elements) == 1, "Should have one asset-clip element for video"
        
        # Verify sequence duration accounts for both clips
        sequence = root.find('.//sequence')
        assert sequence is not None
        sequence_duration = sequence.get('duration')
        assert sequence_duration is not None

    def test_timeline_ordering(self, sample_media_files):
        """Test that timeline elements are properly ordered by offset."""
        fcpxml = create_empty_fcpxml()
        media_files = [sample_media_files['image'], sample_media_files['video']]
        add_media_to_timeline(fcpxml, media_files, clip_duration_seconds=2.0)
        
        xml_content = serialize_to_xml(fcpxml)
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Get all timeline elements
        spine = root.find('.//spine')
        assert spine is not None
        
        timeline_elements = list(spine)
        assert len(timeline_elements) == 2
        
        # Check that offsets are properly ordered
        offsets = []
        for elem in timeline_elements:
            offset_str = elem.get('offset', '0s')
            # Convert to numeric for comparison (simplified)
            if '/' in offset_str:
                offset_str = offset_str.rstrip('s')
                num, den = offset_str.split('/')
                offset_value = int(num) / int(den)
            else:
                offset_value = float(offset_str.rstrip('s'))
            offsets.append(offset_value)
        
        # Offsets should be in ascending order
        assert offsets == sorted(offsets), "Timeline elements should be ordered by offset"

    def test_image_start_attribute(self, sample_media_files):
        """Test that image video elements have the required start attribute."""
        fcpxml = create_empty_fcpxml()
        add_media_to_timeline(fcpxml, [sample_media_files['image']], clip_duration_seconds=5.0)
        
        xml_content = serialize_to_xml(fcpxml)
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        video_elem = root.find('.//spine/video')
        assert video_elem is not None
        
        start_attr = video_elem.get('start')
        assert start_attr is not None, "Image video elements must have start attribute"
        assert start_attr == "3600s", "Standard image start time should be 3600s"

    def test_video_no_start_attribute(self, sample_media_files):
        """Test that video asset-clip elements do NOT have start attribute."""
        fcpxml = create_empty_fcpxml()
        add_media_to_timeline(fcpxml, [sample_media_files['video']], clip_duration_seconds=5.0)
        
        xml_content = serialize_to_xml(fcpxml)
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        asset_clip_elem = root.find('.//spine/asset-clip')
        assert asset_clip_elem is not None
        
        start_attr = asset_clip_elem.get('start')
        assert start_attr is None, "Video asset-clip elements should NOT have start attribute"

    def test_duration_calculation(self, sample_media_files):
        """Test that timeline durations are correctly calculated."""
        fcpxml = create_empty_fcpxml()
        clip_duration = 4.0
        media_files = [sample_media_files['image'], sample_media_files['video']]
        
        add_media_to_timeline(fcpxml, media_files, clip_duration_seconds=clip_duration)
        
        xml_content = serialize_to_xml(fcpxml)
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Check sequence duration
        sequence = root.find('.//sequence')
        sequence_duration = sequence.get('duration')
        
        # Should be total duration for both clips
        # Duration is in format like "240240/24000s" 
        assert sequence_duration is not None
        assert '/' in sequence_duration
        
        # Parse duration
        duration_str = sequence_duration.rstrip('s')
        num, den = duration_str.split('/')
        total_seconds = int(num) / int(den)
        
        expected_total = clip_duration * len(media_files)
        assert abs(total_seconds - expected_total) < 0.1, f"Expected {expected_total}s, got {total_seconds}s"