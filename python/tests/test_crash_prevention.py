"""
Tests for FCP crash prevention patterns.

This module tests all the critical patterns identified from the Go implementation
that prevent Final Cut Pro from crashing during FCPXML import.
"""

import pytest
import tempfile
import os
from pathlib import Path
from xml.etree.ElementTree import fromstring

from fcpxml_lib.core.fcpxml import create_empty_fcpxml, add_media_to_timeline, create_media_asset
from fcpxml_lib.serialization.xml_serializer import serialize_to_xml
from fcpxml_lib.models.elements import FCPXML, SmartCollection


class TestCrashPrevention:
    """Test critical crash prevention patterns."""

    def test_smart_collections_required(self):
        """Test that smart collections are always included to prevent FCP crashes."""
        fcpxml = create_empty_fcpxml()
        
        # Verify smart collections exist
        assert fcpxml.library is not None
        assert len(fcpxml.library.smart_collections) == 5
        
        # Verify required collection names
        collection_names = [sc.name for sc in fcpxml.library.smart_collections]
        required_names = ["Projects", "All Video", "Audio Only", "Stills", "Favorites"]
        
        for name in required_names:
            assert name in collection_names, f"Missing required smart collection: {name}"

    def test_smart_collections_xml_structure(self):
        """Test that smart collections serialize to correct XML structure."""
        fcpxml = create_empty_fcpxml()
        xml_content = serialize_to_xml(fcpxml)
        
        # Parse XML to verify structure
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Find all smart-collection elements
        smart_collections = root.findall('.//smart-collection')
        assert len(smart_collections) == 5, "Should have exactly 5 smart collections"
        
        # Verify Projects collection structure
        projects_collection = None
        for sc in smart_collections:
            if sc.get('name') == 'Projects':
                projects_collection = sc
                break
        
        assert projects_collection is not None
        assert projects_collection.get('match') == 'all'
        
        # Verify it has match-clip rule
        match_clip = projects_collection.find('match-clip')
        assert match_clip is not None
        assert match_clip.get('rule') == 'is'
        assert match_clip.get('type') == 'project'

    def test_video_assets_no_audio_properties(self):
        """Test that video assets never have hasAudio/audioSources properties."""
        # Create temporary video file
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as tmp_video:
            tmp_video.write(b'fake video content')
            tmp_video_path = tmp_video.name
        
        try:
            # Create asset for video
            asset, format_obj = create_media_asset(tmp_video_path, "r2", "r3")
            
            # Verify video asset has NO audio properties
            assert asset.has_audio is None, "Video assets should not have hasAudio property"
            assert asset.audio_sources is None, "Video assets should not have audioSources property"
            assert asset.audio_channels is None, "Video assets should not have audioChannels property"
            assert asset.audio_rate is None, "Video assets should not have audioRate property"
            
            # Verify it has video properties
            assert asset.has_video == "1"
            assert asset.video_sources == "1"
            
        finally:
            os.unlink(tmp_video_path)

    def test_image_assets_correct_properties(self):
        """Test that image assets have correct properties to prevent crashes."""
        # Create temporary image file
        with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp_image:
            tmp_image.write(b'fake image content')
            tmp_image_path = tmp_image.name
        
        try:
            # Create asset for image
            asset, format_obj = create_media_asset(tmp_image_path, "r2", "r3")
            
            # Verify image asset properties
            assert asset.duration == "0s", "Image assets must have duration='0s'"
            assert asset.has_video == "1"
            assert asset.video_sources == "1"
            
            # Verify NO audio properties
            assert asset.has_audio is None
            assert asset.audio_sources is None
            
            # Verify format properties
            assert format_obj.frame_duration is None, "Image formats must not have frameDuration"
            assert format_obj.name == "FFVideoFormatRateUndefined"
            assert format_obj.color_space == "1-13-1"
            
        finally:
            os.unlink(tmp_image_path)

    def test_video_format_no_name_attribute(self):
        """Test that video formats don't have name attribute (per Go patterns)."""
        # Create temporary video file
        with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as tmp_video:
            tmp_video.write(b'fake video content')
            tmp_video_path = tmp_video.name
        
        try:
            # Create asset for video
            asset, format_obj = create_media_asset(tmp_video_path, "r2", "r3")
            
            # Verify video format has NO name attribute
            assert format_obj.name is None, "Video formats should not have name attribute"
            assert format_obj.frame_duration == "1001/24000s"
            assert format_obj.color_space == "1-1-1 (Rec. 709)"
            
        finally:
            os.unlink(tmp_video_path)

    def test_library_location_required(self):
        """Test that library location is always set to prevent crashes."""
        fcpxml = create_empty_fcpxml()
        
        assert fcpxml.library is not None
        assert fcpxml.library.location != ""
        assert fcpxml.library.location.startswith("file://")

    def test_fcpxml_version_113(self):
        """Test that FCPXML version is 1.13 (matching Go implementation)."""
        fcpxml = create_empty_fcpxml()
        
        assert fcpxml.version == "1.13"

    def test_timeline_element_separation(self):
        """Test that images use <video> and videos use <asset-clip> elements."""
        # This is tested through integration but we can verify the logic exists
        # by checking that different file extensions are handled differently
        
        # Test image detection
        image_extensions = ['.png', '.jpg', '.jpeg']
        for ext in image_extensions:
            with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
                tmp.write(b'fake content')
                tmp_path = tmp.name
            
            try:
                asset, format_obj = create_media_asset(tmp_path, "r2", "r3")
                # Images should have duration="0s" 
                assert asset.duration == "0s"
                # Image formats should not have frameDuration
                assert format_obj.frame_duration is None
            finally:
                os.unlink(tmp_path)

    def test_absolute_file_paths(self):
        """Test that media-rep src uses absolute file paths."""
        with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp:
            tmp.write(b'fake content')
            tmp_path = tmp.name
        
        try:
            asset, _ = create_media_asset(tmp_path, "r2", "r3")
            
            # Verify absolute path
            assert asset.media_rep.src.startswith("file://")
            assert os.path.isabs(asset.media_rep.src.replace("file://", ""))
            
        finally:
            os.unlink(tmp_path)