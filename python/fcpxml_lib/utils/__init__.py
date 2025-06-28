"""
Utility functions for FCPXML generation.
"""

from .timing import convert_seconds_to_fcp_duration
from .ids import generate_uid

__all__ = ["convert_seconds_to_fcp_duration", "generate_uid"]