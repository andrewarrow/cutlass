"""
FCPXML Python Library

A modular library for generating valid FCPXML documents following comprehensive validation rules.
"""

from .core.fcpxml import create_empty_project, save_fcpxml
from .models.elements import Asset, Format, MediaRep, Resources, Spine, Sequence, Project, Event, Library, FCPXML
from .validation.validators import validate_frame_alignment, validate_resource_id, validate_audio_rate
from .utils.timing import convert_seconds_to_fcp_duration
from .utils.ids import generate_uid
from .exceptions import FCPXMLError, ValidationError

__version__ = "1.0.0"
__all__ = [
    "FCPXML", "create_empty_project", "save_fcpxml",
    "Asset", "Format", "MediaRep", "Resources", "Spine", "Sequence", "Project", "Event", "Library",
    "validate_frame_alignment", "validate_resource_id", "validate_audio_rate",
    "convert_seconds_to_fcp_duration", "generate_uid",
    "FCPXMLError", "ValidationError"
]