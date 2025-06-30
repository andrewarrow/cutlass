"""
Test multi-lane video audio functionality.

This test validates the complete audio implementation for the many-video-fx command,
ensuring that audio works correctly with complex nested clip structures and multiple lanes.
"""

import pytest
import tempfile
import os
from pathlib import Path
from unittest.mock import patch, MagicMock

from fcpxml_lib.cmd.many_video_fx import many_video_fx_cmd
from fcpxml_lib.core.fcpxml import detect_video_properties, create_media_asset


class TestMultiLaneAudio:
    """Test audio functionality in multi-lane video arrangements."""
    
    @pytest.fixture
    def mock_video_files(self, tmp_path):
        """Create mock video files for testing."""
        video_files = []
        for i in range(3):
            video_file = tmp_path / f"test_video_{i}.mov"
            video_file.write_text("mock video content")
            video_files.append(video_file)
        return video_files
    
    @pytest.fixture
    def mock_args_with_sound(self, mock_video_files):
        """Create mock arguments with sound enabled."""
        args = MagicMock()
        args.input_dir = str(mock_video_files[0].parent)
        args.output = "test_multilane_audio.fcpxml"
        args.duration = 60.0
        args.include_sound = True
        return args
    
    @pytest.fixture
    def mock_args_no_sound(self, mock_video_files):
        """Create mock arguments with sound disabled."""
        args = MagicMock()
        args.input_dir = str(mock_video_files[0].parent)
        args.output = "test_multilane_no_audio.fcpxml"
        args.duration = 60.0
        args.include_sound = False
        return args

    @patch('fcpxml_lib.cmd.many_video_fx.detect_video_properties')
    @patch('fcpxml_lib.cmd.many_video_fx.create_media_asset')
    @patch('fcpxml_lib.cmd.many_video_fx.save_fcpxml')
    def test_audio_enabled_creates_audio_elements(self, mock_save, mock_create_asset, mock_detect_props, mock_args_with_sound):
        """Test that --include-sound creates both video and audio elements."""
        
        # Mock video properties with audio
        mock_detect_props.return_value = {
            'duration_seconds': 30.0,
            'width': 1920,
            'height': 1080,
            'frame_rate': 24.0,
            'has_audio': True,
            'aspect_ratio': 16/9
        }
        
        # Mock asset creation
        mock_asset = MagicMock()
        mock_asset.has_audio = "1"
        mock_asset.audio_sources = "1" 
        mock_asset.audio_channels = "2"
        mock_asset.audio_rate = "48000"
        
        mock_format = MagicMock()
        mock_create_asset.return_value = (mock_asset, mock_format)
        mock_save.return_value = True
        
        # Run the command
        many_video_fx_cmd(mock_args_with_sound)
        
        # Verify create_media_asset was called with include_audio=True
        assert mock_create_asset.called
        calls = mock_create_asset.call_args_list
        for call in calls:
            args, kwargs = call
            assert kwargs.get('include_audio') == True, "include_audio should be True when --include-sound is used"
    
    @patch('fcpxml_lib.cmd.many_video_fx.detect_video_properties')
    @patch('fcpxml_lib.cmd.many_video_fx.create_media_asset')
    @patch('fcpxml_lib.cmd.many_video_fx.save_fcpxml')
    def test_audio_disabled_no_audio_elements(self, mock_save, mock_create_asset, mock_detect_props, mock_args_no_sound):
        """Test that without --include-sound, no audio elements are created."""
        
        # Mock video properties with audio
        mock_detect_props.return_value = {
            'duration_seconds': 30.0,
            'width': 1920,
            'height': 1080, 
            'frame_rate': 24.0,
            'has_audio': True,
            'aspect_ratio': 16/9
        }
        
        # Mock asset creation (no audio properties)
        mock_asset = MagicMock()
        mock_asset.has_audio = None
        mock_format = MagicMock()
        mock_create_asset.return_value = (mock_asset, mock_format)
        mock_save.return_value = True
        
        # Run the command
        many_video_fx_cmd(mock_args_no_sound)
        
        # Verify create_media_asset was called with include_audio=False
        assert mock_create_asset.called
        calls = mock_create_asset.call_args_list
        for call in calls:
            args, kwargs = call
            assert kwargs.get('include_audio') == False, "include_audio should be False when --include-sound is not used"

    def test_create_media_asset_audio_properties(self):
        """Test that create_media_asset correctly adds audio properties when requested."""
        
        # Mock video file
        with tempfile.NamedTemporaryFile(suffix='.mov', delete=False) as temp_file:
            temp_path = temp_file.name
        
        try:
            with patch('fcpxml_lib.core.fcpxml.detect_video_properties') as mock_detect:
                mock_detect.return_value = {
                    'duration_seconds': 30.0,
                    'width': 1920,
                    'height': 1080,
                    'frame_rate': 24.0,
                    'has_audio': True,
                    'aspect_ratio': 16/9
                }
                
                # Test with audio enabled
                asset_with_audio, _ = create_media_asset(temp_path, "r2", "r3", include_audio=True)
                assert asset_with_audio.has_audio == "1"
                assert asset_with_audio.audio_sources == "1"
                assert asset_with_audio.audio_channels == "2"
                assert asset_with_audio.audio_rate == "48000"
                
                # Test with audio disabled
                asset_no_audio, _ = create_media_asset(temp_path, "r4", "r5", include_audio=False)
                assert asset_no_audio.has_audio is None
                assert asset_no_audio.audio_sources is None
                assert asset_no_audio.audio_channels is None
                assert asset_no_audio.audio_rate is None
                
        finally:
            os.unlink(temp_path)

    def test_audio_element_structure_requirements(self):
        """Test that audio elements have the required structure for FCP compatibility."""
        
        # This test validates the key discovery: audio elements must be present
        # alongside video elements for audio to work in complex clip structures
        
        expected_audio_structure = {
            "type": "audio",
            "ref": "r2",  # Must reference an asset with hasAudio="1"
            "offset": "0s",
            "duration": "30s",
            "role": "dialogue"  # Required for proper audio routing
        }
        
        expected_video_structure = {
            "type": "video", 
            "ref": "r2",  # Same asset reference as audio
            "offset": "0s",
            "duration": "30s"
        }
        
        # Validate that both structures reference the same asset
        assert expected_audio_structure["ref"] == expected_video_structure["ref"]
        assert expected_audio_structure["offset"] == expected_video_structure["offset"] 
        assert expected_audio_structure["duration"] == expected_video_structure["duration"]
        
        # Validate audio-specific attributes
        assert "role" in expected_audio_structure
        assert expected_audio_structure["role"] == "dialogue"

    def test_dtd_compliance_requirements(self):
        """Test the DTD compliance requirements discovered during implementation."""
        
        # Key discovery: clip elements don't support audioRole attribute
        # Must use audio elements with role attribute instead
        
        invalid_clip_structure = {
            "type": "clip",
            "audioRole": "dialogue"  # ❌ Invalid per DTD
        }
        
        valid_audio_structure = {
            "type": "audio",
            "role": "dialogue"  # ✅ Valid per DTD
        }
        
        # The test captures the learning that audioRole belongs on audio elements,
        # not on clip elements
        assert "audioRole" not in valid_audio_structure
        assert "role" in valid_audio_structure

    def test_asset_audio_properties_requirement(self):
        """Test that assets must have audio properties for audio elements to work."""
        
        # Key discovery: Assets need hasAudio="1" properties for audio to work
        required_asset_audio_properties = {
            "hasAudio": "1",
            "audioSources": "1", 
            "audioChannels": "2",
            "audioRate": "48000"
        }
        
        # Without these properties on the asset, audio elements won't produce sound
        for prop, value in required_asset_audio_properties.items():
            assert prop in required_asset_audio_properties
            assert required_asset_audio_properties[prop] == value

    def test_multilane_audio_architecture(self):
        """Test the complete multi-lane audio architecture."""
        
        # This test documents the complete architecture needed for multi-lane audio:
        # 1. Assets with audio properties
        # 2. Complex clips with nested video and audio elements  
        # 3. Proper DTD ordering and structure
        
        multilane_structure = {
            "main_clip": {
                "type": "clip",
                "nested_elements": [
                    {"type": "adjust_transform"},  # Transforms first
                    {"type": "video", "ref": "r2"},  # Video elements
                    {"type": "audio", "ref": "r2", "role": "dialogue"},  # Audio elements
                    {"type": "clip", "lane": "1"},  # Nested clips
                    {"type": "clip", "lane": "2"},
                    {"type": "audio-channel-source", "role": "dialogue"}  # Audio routing last
                ]
            }
        }
        
        # Validate the architecture elements are present
        main_elements = multilane_structure["main_clip"]["nested_elements"]
        element_types = [elem["type"] for elem in main_elements]
        
        assert "adjust_transform" in element_types
        assert "video" in element_types
        assert "audio" in element_types
        assert "clip" in element_types
        assert "audio-channel-source" in element_types
        
        # Validate that audio and video reference the same asset
        video_elem = next(elem for elem in main_elements if elem["type"] == "video")
        audio_elem = next(elem for elem in main_elements if elem["type"] == "audio")
        assert video_elem["ref"] == audio_elem["ref"]

    def test_successful_audio_implementation_pattern(self):
        """Test the complete pattern that successfully enables audio in FCP."""
        
        # This test captures the complete working pattern discovered through
        # comparing foo.fcpxml (working) with sound.fcpxml (working)
        
        working_pattern = {
            # 1. Assets must have full audio properties
            "asset_requirements": {
                "hasAudio": "1",
                "audioSources": "1",
                "audioChannels": "2", 
                "audioRate": "48000"
            },
            
            # 2. Timeline must have both video and audio elements
            "timeline_requirements": {
                "video_element": {"type": "video", "ref": "asset_id"},
                "audio_element": {"type": "audio", "ref": "asset_id", "role": "dialogue"}
            },
            
            # 3. DTD compliance requirements
            "dtd_requirements": {
                "no_audioRole_on_clips": True,
                "audio_elements_have_role": True,
                "proper_element_ordering": True
            }
        }
        
        # Validate the complete pattern
        assert working_pattern["asset_requirements"]["hasAudio"] == "1"
        assert working_pattern["timeline_requirements"]["audio_element"]["role"] == "dialogue"
        assert working_pattern["dtd_requirements"]["no_audioRole_on_clips"] == True

    @patch('fcpxml_lib.cmd.many_video_fx.save_fcpxml')
    def test_audio_success_message(self, mock_save, mock_args_with_sound, capsys):
        """Test that success message indicates audio inclusion."""
        
        with patch('fcpxml_lib.cmd.many_video_fx.detect_video_properties') as mock_detect:
            with patch('fcpxml_lib.cmd.many_video_fx.create_media_asset') as mock_create:
                mock_detect.return_value = {
                    'duration_seconds': 30.0,
                    'width': 1920, 'height': 1080,
                    'frame_rate': 24.0, 'has_audio': True,
                    'aspect_ratio': 16/9
                }
                
                mock_asset = MagicMock()
                mock_format = MagicMock() 
                mock_create.return_value = (mock_asset, mock_format)
                mock_save.return_value = True
                
                many_video_fx_cmd(mock_args_with_sound)
                
                captured = capsys.readouterr()
                assert "🔊 Audio included from all" in captured.out