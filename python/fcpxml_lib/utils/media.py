"""
Media File Utilities

Shared utilities for media file discovery, processing, and asset creation.
Extracted from main.py to eliminate code duplication and improve maintainability.
"""

from pathlib import Path
from typing import List, Set, Tuple

from fcpxml_lib.models.elements import Format, Asset, MediaRep
from fcpxml_lib.utils.ids import generate_uid
from fcpxml_lib.constants import (
    IMAGE_FORMAT_NAME, DEFAULT_IMAGE_WIDTH, DEFAULT_IMAGE_HEIGHT, IMAGE_COLOR_SPACE,
    IMAGE_DURATION
)


def discover_media_files(input_dir: Path, extensions: Set[str]) -> List[Path]:
    """
    Discover media files in a directory with specified extensions.
    
    Args:
        input_dir: Directory to search in
        extensions: Set of file extensions to look for (e.g., {'.png', '.jpg'})
        
    Returns:
        List of Path objects for found media files
    """
    if not input_dir.exists() or not input_dir.is_dir():
        return []
    
    media_files = []
    for ext in extensions:
        # Search for both lowercase and uppercase versions
        media_files.extend(input_dir.glob(f"*{ext}"))
        media_files.extend(input_dir.glob(f"*{ext.upper()}"))
    
    return media_files


def discover_image_files(input_dir: Path) -> List[Path]:
    """
    Discover image files (PNG, JPG, JPEG) in a directory.
    
    Args:
        input_dir: Directory to search in
        
    Returns:
        List of Path objects for found image files
    """
    image_extensions = {'.png', '.jpg', '.jpeg'}
    return discover_media_files(input_dir, image_extensions)


def discover_video_files(input_dir: Path) -> List[Path]:
    """
    Discover video files (MOV, MP4, AVI, MKV, M4V) in a directory.
    
    Args:
        input_dir: Directory to search in
        
    Returns:
        List of Path objects for found video files
    """
    video_extensions = {'.mov', '.mp4', '.avi', '.mkv', '.m4v'}
    return discover_media_files(input_dir, video_extensions)


def discover_all_media_files(input_dir: Path) -> Tuple[List[Path], List[Path]]:
    """
    Discover both image and video files in a directory.
    
    Args:
        input_dir: Directory to search in
        
    Returns:
        Tuple of (image_files, video_files) lists
    """
    image_files = discover_image_files(input_dir)
    video_files = discover_video_files(input_dir)
    return image_files, video_files


def create_shared_image_format(format_id: str) -> Format:
    """
    Create a shared image format that can be reused across multiple image assets.
    
    Args:
        format_id: Unique format ID to use
        
    Returns:
        Format object configured for images
    """
    return Format(
        id=format_id,
        name=IMAGE_FORMAT_NAME,
        width=DEFAULT_IMAGE_WIDTH,
        height=DEFAULT_IMAGE_HEIGHT,
        color_space=IMAGE_COLOR_SPACE
    )


def create_image_asset_with_shared_format(
    image_path: Path, 
    asset_id: str, 
    format_id: str
) -> Asset:
    """
    Create an image asset that uses a shared format.
    
    Args:
        image_path: Path to the image file
        asset_id: Unique asset ID to use
        format_id: Format ID to reference (should be created separately)
        
    Returns:
        Asset object configured for the image
    """
    abs_path = image_path.resolve()
    uid = generate_uid(f"MEDIA_{abs_path.name}")
    media_rep = MediaRep(src=str(abs_path))
    
    return Asset(
        id=asset_id,
        name=abs_path.stem,
        uid=uid,
        duration=IMAGE_DURATION,
        has_video="1",
        format=format_id,  # Reference to shared format
        video_sources="1",
        media_rep=media_rep
    )


def get_media_type_info(file_path: Path) -> dict:
    """
    Get media type information for a file.
    
    Args:
        file_path: Path to the media file
        
    Returns:
        Dictionary with media type information:
        - is_video: bool
        - is_image: bool  
        - extension: str
        - type_name: str
    """
    ext = file_path.suffix.lower()
    
    video_extensions = {'.mov', '.mp4', '.avi', '.mkv', '.m4v'}
    image_extensions = {'.png', '.jpg', '.jpeg', '.tiff', '.bmp', '.gif'}
    
    is_video = ext in video_extensions
    is_image = ext in image_extensions
    
    if is_video:
        type_name = "video"
    elif is_image:
        type_name = "image"
    else:
        type_name = "unknown"
    
    return {
        "is_video": is_video,
        "is_image": is_image,
        "extension": ext,
        "type_name": type_name
    }