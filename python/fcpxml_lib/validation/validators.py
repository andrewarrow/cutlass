"""
Validation functions for FCPXML elements.
"""

import re
from ..constants import STANDARD_TIMEBASE, RESOURCE_ID_PATTERN, VALID_AUDIO_RATES


def validate_frame_alignment(duration: str) -> bool:
    """Validate that a duration string is frame-aligned according to FCP rules"""
    if duration == "0s":
        return True
        
    if not duration.endswith("s"):
        return False
        
    if "/" not in duration:
        return False
        
    try:
        time_part = duration.rstrip("s")
        numerator, denominator = map(int, time_part.split("/"))
        
        # Check timebase and frame alignment (1001 is the frame duration component)
        return (denominator == STANDARD_TIMEBASE and 
                numerator % 1001 == 0)
    except (ValueError, IndexError):
        return False


def validate_resource_id(resource_id: str) -> bool:
    """Validate resource ID follows FCP pattern (r1, r2, etc.)"""
    return bool(re.match(RESOURCE_ID_PATTERN, resource_id))


def validate_audio_rate(audio_rate: str) -> bool:
    """Validate audio rate is in DTD enumerated set"""
    return audio_rate in VALID_AUDIO_RATES