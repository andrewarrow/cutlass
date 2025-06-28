"""
Pytest configuration and shared fixtures.
"""

import pytest
import tempfile
import os
from pathlib import Path


@pytest.fixture
def temp_image_file():
    """Create a temporary image file for testing."""
    with tempfile.NamedTemporaryFile(suffix='.jpg', delete=False) as tmp:
        tmp.write(b'fake image content')
        tmp_path = tmp.name
    
    yield tmp_path
    
    if os.path.exists(tmp_path):
        os.unlink(tmp_path)


@pytest.fixture
def temp_video_file():
    """Create a temporary video file for testing."""
    with tempfile.NamedTemporaryFile(suffix='.mp4', delete=False) as tmp:
        tmp.write(b'fake video content')
        tmp_path = tmp.name
    
    yield tmp_path
    
    if os.path.exists(tmp_path):
        os.unlink(tmp_path)


@pytest.fixture
def temp_output_file():
    """Create a temporary output file path for testing."""
    with tempfile.NamedTemporaryFile(suffix='.fcpxml', delete=False) as tmp:
        tmp_path = tmp.name
    
    # Remove the file so tests can create it
    os.unlink(tmp_path)
    
    yield tmp_path
    
    # Cleanup if test created the file
    if os.path.exists(tmp_path):
        os.unlink(tmp_path)


@pytest.fixture
def sample_media_directory():
    """Create a temporary directory with sample media files."""
    import tempfile
    import shutil
    
    temp_dir = tempfile.mkdtemp()
    
    # Create sample files
    files = []
    
    # Create images
    for i, ext in enumerate(['.png', '.jpg']):
        file_path = os.path.join(temp_dir, f'image_{i}{ext}')
        with open(file_path, 'wb') as f:
            f.write(f'fake image {i}'.encode())
        files.append(file_path)
    
    # Create videos
    for i, ext in enumerate(['.mp4', '.mov']):
        file_path = os.path.join(temp_dir, f'video_{i}{ext}')
        with open(file_path, 'wb') as f:
            f.write(f'fake video {i}'.encode())
        files.append(file_path)
    
    yield temp_dir, files
    
    # Cleanup
    shutil.rmtree(temp_dir, ignore_errors=True)