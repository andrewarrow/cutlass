"""
Tests for media file detection and property extraction.

Tests the logic that detects video properties and handles different media types
to prevent crashes caused by incorrect property assumptions.
"""

import pytest
import tempfile
import os
from unittest.mock import patch, MagicMock

from fcpxml_lib.core.fcpxml import detect_video_properties, create_media_asset


class TestMediaDetection:
    """Test media file detection and property extraction."""

    def test_detect_video_properties_fallback(self):
        """Test that detect_video_properties returns safe defaults on failure."""
        # Test with non-existent file
        props = detect_video_properties("nonexistent_file.mp4")
        
        # Should return safe defaults
        assert props["duration_seconds"] == 10.0
        assert props["width"] == 1920
        assert props["height"] == 1080
        assert abs(props["frame_rate"] - 23.976) < 0.001
        assert props["has_audio"] == False

    @patch('subprocess.run')
    def test_detect_video_properties_success(self, mock_run):
        """Test successful video property detection."""
        # Mock ffprobe output for video properties
        mock_video_result = MagicMock()
        mock_video_result.stdout = "h264,1920,1080,30000/1001,120.5"
        mock_video_result.check = True
        
        # Mock ffprobe output for audio detection
        mock_audio_result = MagicMock()
        mock_audio_result.stdout = "aac"
        
        mock_run.side_effect = [mock_video_result, mock_audio_result]
        
        props = detect_video_properties("test_video.mp4")
        
        assert props["width"] == 1920
        assert props["height"] == 1080
        assert props["duration_seconds"] == 120.5
        assert props["frame_rate"] == 30000/1001
        assert props["has_audio"] == True

    @patch('subprocess.run')
    def test_detect_video_properties_no_audio(self, mock_run):
        """Test video property detection with no audio."""
        # Mock ffprobe output for video properties
        mock_video_result = MagicMock()
        mock_video_result.stdout = "h264,1280,720,24000/1001,60.0"
        mock_video_result.check = True
        
        # Mock ffprobe output for no audio
        mock_audio_result = MagicMock()
        mock_audio_result.stdout = ""
        
        mock_run.side_effect = [mock_video_result, mock_audio_result]
        
        props = detect_video_properties("test_video.mp4")
        
        assert props["width"] == 1280
        assert props["height"] == 720
        assert props["duration_seconds"] == 60.0
        assert props["has_audio"] == False

    @patch('subprocess.run')
    def test_detect_video_properties_ffprobe_error(self, mock_run):
        """Test handling of ffprobe errors."""
        # Mock ffprobe failure
        mock_run.side_effect = Exception("ffprobe not found")
        
        props = detect_video_properties("test_video.mp4")
        
        # Should return safe defaults
        assert props["duration_seconds"] == 10.0
        assert props["width"] == 1920
        assert props["height"] == 1080
        assert props["has_audio"] == False

    def test_image_file_detection(self):
        """Test that image files are correctly identified."""
        image_extensions = ['.png', '.jpg', '.jpeg', '.PNG', '.JPG', '.JPEG']
        
        for ext in image_extensions:
            with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
                tmp.write(b'fake image content')
                tmp_path = tmp.name
            
            try:
                asset, format_obj = create_media_asset(tmp_path, "r2", "r3")
                
                # Images should have specific properties
                assert asset.duration == "0s"
                assert format_obj.frame_duration is None
                assert format_obj.name == "FFVideoFormatRateUndefined"
                assert format_obj.color_space == "1-13-1"
                
            finally:
                os.unlink(tmp_path)

    def test_video_file_detection(self):
        """Test that video files are correctly identified."""
        video_extensions = ['.mp4', '.mov', '.MP4', '.MOV']
        
        for ext in video_extensions:
            with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
                tmp.write(b'fake video content')
                tmp_path = tmp.name
            
            try:
                asset, format_obj = create_media_asset(tmp_path, "r2", "r3")
                
                # Videos should have specific properties
                assert asset.duration != "0s"  # Should have actual duration
                assert format_obj.frame_duration is not None
                assert format_obj.name == ""  # Video formats have empty name
                assert format_obj.color_space == "1-1-1 (Rec. 709)"
                
            finally:
                os.unlink(tmp_path)

    def test_unsupported_file_type(self):
        """Test handling of unsupported file types."""
        with tempfile.NamedTemporaryFile(suffix='.txt', delete=False) as tmp:
            tmp.write(b'text content')
            tmp_path = tmp.name
        
        try:
            with pytest.raises(ValueError, match="Unsupported media type"):
                create_media_asset(tmp_path, "r2", "r3")
        finally:
            os.unlink(tmp_path)

    def test_absolute_path_conversion(self):
        """Test that relative paths are converted to absolute."""
        with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp:
            tmp.write(b'fake image content')
            tmp_path = tmp.name
        
        try:
            # Use just the filename (relative path)
            relative_path = os.path.basename(tmp_path)
            
            # Change to temp directory so relative path works
            original_cwd = os.getcwd()
            temp_dir = os.path.dirname(tmp_path)
            os.chdir(temp_dir)
            
            try:
                asset, _ = create_media_asset(relative_path, "r2", "r3")
                
                # Should result in absolute path
                assert asset.media_rep.src.startswith("file://")
                src_path = asset.media_rep.src.replace("file://", "")
                assert os.path.isabs(src_path)
                
            finally:
                os.chdir(original_cwd)
        finally:
            os.unlink(tmp_path)

    def test_uid_generation_format(self):
        """Test that UID generation follows the expected format."""
        with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp:
            tmp.write(b'fake image content')
            tmp_path = tmp.name
        
        try:
            # Create asset
            asset, _ = create_media_asset(tmp_path, "r2", "r3")
            
            # UID should be a valid hex string
            assert len(asset.uid) == 32  # MD5 hash length
            assert all(c in '0123456789ABCDEF' for c in asset.uid)
            
        finally:
            os.unlink(tmp_path)

    @patch('subprocess.run')
    def test_frame_rate_parsing(self, mock_run):
        """Test different frame rate formats are parsed correctly."""
        test_cases = [
            ("30000/1001", 30000/1001),  # 29.97 fps
            ("24000/1001", 24000/1001),  # 23.976 fps  
            ("30", 30.0),                # 30 fps
            ("25", 25.0),                # 25 fps
        ]
        
        for frame_rate_str, expected in test_cases:
            mock_video_result = MagicMock()
            mock_video_result.stdout = f"h264,1920,1080,{frame_rate_str},60.0"
            mock_video_result.check = True
            
            mock_audio_result = MagicMock()
            mock_audio_result.stdout = ""
            
            mock_run.side_effect = [mock_video_result, mock_audio_result]
            
            props = detect_video_properties("test.mp4")
            assert abs(props["frame_rate"] - expected) < 0.001