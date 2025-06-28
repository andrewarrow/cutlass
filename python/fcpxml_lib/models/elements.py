"""
Data models for FCPXML elements.
"""

import os
import time
from dataclasses import dataclass, field
from typing import List, Optional, Dict

from ..exceptions import ValidationError
from ..validation.validators import validate_resource_id, validate_frame_alignment, validate_audio_rate
from ..utils.ids import generate_uid


@dataclass
class MediaRep:
    """Media representation with file path and metadata"""
    kind: str = "original-media"
    sig: str = ""
    src: str = ""
    
    def __post_init__(self):
        if not self.src.startswith("file://"):
            # Ensure absolute file:// URL format
            abs_path = os.path.abspath(self.src) if self.src else ""
            self.src = f"file://{abs_path}"


@dataclass 
class Format:
    """Video/audio format definition"""
    id: str
    name: str = ""
    frame_duration: Optional[str] = None
    width: Optional[str] = None
    height: Optional[str] = None
    color_space: Optional[str] = None
    
    def __post_init__(self):
        if not validate_resource_id(self.id):
            raise ValidationError(f"Invalid format ID: {self.id}")
        
        if self.frame_duration and not validate_frame_alignment(self.frame_duration):
            raise ValidationError(f"Frame duration not aligned: {self.frame_duration}")


@dataclass
class Asset:
    """Media asset (video, image, audio)"""
    id: str
    name: str
    uid: str
    start: str = "0s"
    duration: str = "0s"
    has_video: Optional[str] = None
    format: Optional[str] = None
    video_sources: Optional[str] = None
    has_audio: Optional[str] = None
    audio_sources: Optional[str] = None
    audio_channels: Optional[str] = None
    audio_rate: Optional[str] = None
    media_rep: Optional[MediaRep] = None
    
    def __post_init__(self):
        if not validate_resource_id(self.id):
            raise ValidationError(f"Invalid asset ID: {self.id}")
            
        if not validate_frame_alignment(self.duration):
            raise ValidationError(f"Asset duration not frame-aligned: {self.duration}")
            
        if not validate_frame_alignment(self.start):
            raise ValidationError(f"Asset start time not frame-aligned: {self.start}")


@dataclass
class Resources:
    """Container for all shared resources"""
    assets: List[Asset] = field(default_factory=list)
    formats: List[Format] = field(default_factory=list)
    effects: List[Dict] = field(default_factory=list)
    media: List[Dict] = field(default_factory=list)


@dataclass
class Spine:
    """Main timeline container - currently empty for minimal implementation"""
    asset_clips: List[Dict] = field(default_factory=list)
    videos: List[Dict] = field(default_factory=list) 
    titles: List[Dict] = field(default_factory=list)
    gaps: List[Dict] = field(default_factory=list)
    ordered_elements: List[Dict] = field(default_factory=list)


@dataclass
class Sequence:
    """Timeline sequence definition"""
    format: str
    duration: str = "0s"
    tc_start: str = "0s"
    tc_format: str = "NDF"
    audio_layout: str = "stereo"
    audio_rate: str = "48k"  # Use DTD-valid enumerated value
    spine: Spine = field(default_factory=Spine)
    
    def __post_init__(self):
        if not validate_frame_alignment(self.duration):
            raise ValidationError(f"Sequence duration not frame-aligned: {self.duration}")
        if not validate_frame_alignment(self.tc_start):
            raise ValidationError(f"Sequence tc_start not frame-aligned: {self.tc_start}")
        if not validate_audio_rate(self.audio_rate):
            from ..constants import VALID_AUDIO_RATES
            raise ValidationError(f"Invalid audio rate: {self.audio_rate}. Must be one of {VALID_AUDIO_RATES}")


@dataclass
class Project:
    """FCPXML Project definition"""
    name: str
    uid: str = ""
    mod_date: str = ""
    sequences: List[Sequence] = field(default_factory=list)
    
    def __post_init__(self):
        if not self.uid:
            self.uid = generate_uid("PROJECT")
        if not self.mod_date:
            self.mod_date = time.strftime("%Y-%m-%d %H:%M:%S %z")


@dataclass
class Event:
    """FCPXML Event definition"""
    name: str
    uid: str = ""
    projects: List[Project] = field(default_factory=list)
    
    def __post_init__(self):
        if not self.uid:
            self.uid = generate_uid("EVENT")


@dataclass
class SmartCollection:
    """FCPXML Smart Collection definition"""
    name: str
    match: str
    rules: List[dict] = field(default_factory=list)


@dataclass
class Library:
    """FCPXML Library definition"""
    location: str = ""
    events: List[Event] = field(default_factory=list)
    smart_collections: List[SmartCollection] = field(default_factory=list)


@dataclass
class FCPXML:
    """
    Root FCPXML document.
    
    🚨 CRITICAL: Follows schema.yaml principles:
    - Uses structured data objects (NO_XML_TEMPLATES)
    - Validates frame alignment
    - Ensures proper resource management
    """
    version: str = "1.13"
    resources: Resources = field(default_factory=Resources)
    library: Optional[Library] = None
    
    def __post_init__(self):
        # Validate version follows FCP pattern (major.minor format)
        import re
        version_pattern = r"^\d+\.\d+$"  # e.g., "1.13", "1.11", etc.
        if not re.match(version_pattern, self.version):
            raise ValidationError(f"Invalid FCPXML version: {self.version}. Must be in format 'major.minor'")