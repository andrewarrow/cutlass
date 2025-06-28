"""
Integration tests for end-to-end FCPXML generation.

Tests the complete workflow from media files to final FCPXML output,
ensuring compatibility with Final Cut Pro.
"""

import pytest
import tempfile
import os
from pathlib import Path
from xml.etree.ElementTree import fromstring

from fcpxml_lib.core.fcpxml import create_empty_project, add_media_to_timeline, save_fcpxml
from fcpxml_lib.validation.xml_validator import run_xml_validation


class TestIntegration:
    """Integration tests for complete FCPXML generation workflow."""

    @pytest.fixture
    def mixed_media_files(self):
        """Create a mix of image and video files for testing."""
        files = []
        
        # Create multiple image files
        for i, ext in enumerate(['.png', '.jpg']):
            with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
                tmp.write(f'fake image content {i}'.encode())
                files.append(tmp.name)
        
        # Create multiple video files  
        for i, ext in enumerate(['.mp4', '.mov']):
            with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
                tmp.write(f'fake video content {i}'.encode())
                files.append(tmp.name)
        
        yield files
        
        # Cleanup
        for file_path in files:
            if os.path.exists(file_path):
                os.unlink(file_path)

    def test_end_to_end_fcpxml_generation(self, mixed_media_files):
        """Test complete FCPXML generation from media files."""
        # Create FCPXML with mixed media
        fcpxml = create_empty_project()
        add_media_to_timeline(fcpxml, mixed_media_files, clip_duration_seconds=3.0)
        
        # Write to temporary file
        with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as tmp:
            output_path = tmp.name
        
        try:
            success = save_fcpxml(fcpxml, output_path)
            assert success, "FCPXML generation should succeed"
            
            # Verify file was created
            assert os.path.exists(output_path)
            assert os.path.getsize(output_path) > 0
            
            # Verify XML is well-formed
            is_valid, error_msg = run_xml_validation(output_path)
            assert is_valid, f"Generated FCPXML should be valid XML: {error_msg}"
            
        finally:
            if os.path.exists(output_path):
                os.unlink(output_path)

    def test_fcpxml_contains_all_required_elements(self, mixed_media_files):
        """Test that generated FCPXML contains all elements needed to prevent crashes."""
        fcpxml = create_empty_project()
        add_media_to_timeline(fcpxml, mixed_media_files, clip_duration_seconds=2.0)
        
        # Write and read back the FCPXML
        with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as tmp:
            output_path = tmp.name
        
        try:
            save_fcpxml(fcpxml, output_path)
            
            # Parse the generated XML
            with open(output_path, 'r') as f:
                xml_content = f.read()
            
            root = fromstring(xml_content)
            
            # Verify root structure
            assert root.tag == "fcpxml"
            assert root.get("version") == "1.13"
            
            # Verify resources section
            resources = root.find('resources')
            assert resources is not None
            
            assets = resources.findall('asset')
            formats = resources.findall('format')
            assert len(assets) == len(mixed_media_files)
            assert len(formats) >= len(mixed_media_files)  # At least one format per asset
            
            # Verify library structure
            library = root.find('library')
            assert library is not None
            assert library.get('location') is not None
            
            # Verify smart collections (critical for crash prevention)
            smart_collections = library.findall('smart-collection')
            assert len(smart_collections) == 5
            
            collection_names = [sc.get('name') for sc in smart_collections]
            required_names = ["Projects", "All Video", "Audio Only", "Stills", "Favorites"]
            for name in required_names:
                assert name in collection_names
            
            # Verify timeline structure
            sequence = root.find('.//sequence')
            assert sequence is not None
            
            spine = sequence.find('spine')
            assert spine is not None
            
            # Count timeline elements
            timeline_elements = list(spine)
            assert len(timeline_elements) == len(mixed_media_files)
            
        finally:
            if os.path.exists(output_path):
                os.unlink(output_path)

    def test_proper_element_separation(self, mixed_media_files):
        """Test that images and videos create correct timeline elements."""
        fcpxml = create_empty_project()
        add_media_to_timeline(fcpxml, mixed_media_files, clip_duration_seconds=2.0)
        
        with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as tmp:
            output_path = tmp.name
        
        try:
            save_fcpxml(fcpxml, output_path)
            
            with open(output_path, 'r') as f:
                xml_content = f.read()
            
            root = fromstring(xml_content)
            
            # Count images and videos in input
            image_count = sum(1 for f in mixed_media_files if Path(f).suffix.lower() in ['.png', '.jpg', '.jpeg'])
            video_count = sum(1 for f in mixed_media_files if Path(f).suffix.lower() in ['.mp4', '.mov'])
            
            # Count elements in timeline
            video_elements = root.findall('.//spine/video')
            asset_clip_elements = root.findall('.//spine/asset-clip')
            
            assert len(video_elements) == image_count, "Should have video elements for images"
            assert len(asset_clip_elements) == video_count, "Should have asset-clip elements for videos"
            
        finally:
            if os.path.exists(output_path):
                os.unlink(output_path)

    def test_large_media_collection(self):
        """Test handling of larger media collections."""
        # Create a larger set of media files
        media_files = []
        
        try:
            # Create 10 image files and 5 video files
            for i in range(10):
                with tempfile.NamedTemporaryFile(suffix='.png', delete=False) as tmp:
                    tmp.write(f'image {i}'.encode())
                    media_files.append(tmp.name)
            
            for i in range(5):
                with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as tmp:
                    tmp.write(f'video {i}'.encode())
                    media_files.append(tmp.name)
            
            # Generate FCPXML
            fcpxml = create_empty_project()
            add_media_to_timeline(fcpxml, media_files, clip_duration_seconds=1.0)
            
            with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as tmp:
                output_path = tmp.name
            
            try:
                success = save_fcpxml(fcpxml, output_path)
                assert success
                
                # Verify the file is reasonable size (not empty, not too large)
                file_size = os.path.getsize(output_path)
                assert file_size > 1000, "FCPXML should be substantial for 15 media files"
                assert file_size < 1000000, "FCPXML should not be excessively large"
                
                # Verify XML validation
                is_valid, _ = run_xml_validation(output_path)
                assert is_valid
                
            finally:
                if os.path.exists(output_path):
                    os.unlink(output_path)
                    
        finally:
            # Cleanup media files
            for file_path in media_files:
                if os.path.exists(file_path):
                    os.unlink(file_path)

    def test_empty_media_list(self):
        """Test handling of empty media list."""
        fcpxml = create_empty_project()
        add_media_to_timeline(fcpxml, [], clip_duration_seconds=5.0)
        
        with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as tmp:
            output_path = tmp.name
        
        try:
            success = save_fcpxml(fcpxml, output_path)
            assert success
            
            # Should still be valid XML with basic structure
            is_valid, _ = run_xml_validation(output_path)
            assert is_valid
            
            # Parse and verify basic structure exists
            with open(output_path, 'r') as f:
                xml_content = f.read()
            
            root = fromstring(xml_content)
            
            # Should still have smart collections and basic structure
            smart_collections = root.findall('.//smart-collection')
            assert len(smart_collections) == 5
            
            # Timeline should be empty but valid
            spine = root.find('.//spine')
            assert spine is not None
            assert len(list(spine)) == 0  # No timeline elements
            
        finally:
            if os.path.exists(output_path):
                os.unlink(output_path)

    def test_timeline_duration_calculation(self, mixed_media_files):
        """Test that timeline duration is correctly calculated."""
        clip_duration = 2.5
        fcpxml = create_empty_project()
        add_media_to_timeline(fcpxml, mixed_media_files, clip_duration_seconds=clip_duration)
        
        with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as tmp:
            output_path = tmp.name
        
        try:
            save_fcpxml(fcpxml, output_path)
            
            with open(output_path, 'r') as f:
                xml_content = f.read()
            
            root = fromstring(xml_content)
            
            # Get sequence duration
            sequence = root.find('.//sequence')
            sequence_duration = sequence.get('duration')
            
            # Parse FCP duration format
            duration_str = sequence_duration.rstrip('s')
            num, den = duration_str.split('/')
            total_seconds = int(num) / int(den)
            
            expected_total = clip_duration * len(mixed_media_files)
            assert abs(total_seconds - expected_total) < 0.1
            
        finally:
            if os.path.exists(output_path):
                os.unlink(output_path)