"""
Validation functions for FCPXML elements.
"""

import re
from ..utils.schema_loader import get_schema


def validate_frame_alignment(duration: str) -> bool:
    """Validate that a duration string is frame-aligned according to schema rules"""
    schema = get_schema()
    timing = schema['fcpxml_rules']['timing']
    
    if duration == "0s":
        return True
        
    if not duration.endswith("s"):
        return False
        
    if "/" not in duration:
        return False
        
    try:
        time_part = duration.rstrip("s")
        numerator, denominator = map(int, time_part.split("/"))
        
        return (denominator == timing['timebase'] and 
                numerator % timing['frame_duration'] == 0)
    except (ValueError, IndexError):
        return False


def validate_resource_id(resource_id: str) -> bool:
    """Validate resource ID follows FCP pattern (r1, r2, etc.)"""
    schema = get_schema()
    pattern = schema['fcpxml_rules']['resource_ids']['pattern']
    return bool(re.match(pattern, resource_id))


def validate_audio_rate(audio_rate: str) -> bool:
    """Validate audio rate is in DTD enumerated set"""
    schema = get_schema()
    valid_rates = schema['fcpxml_rules']['audio_rates']['valid_values']
    return audio_rate in valid_rates