"""
Timing and duration utilities for FCPXML.
"""

from .schema_loader import get_schema


def convert_seconds_to_fcp_duration(seconds: float) -> str:
    """
    Convert seconds to frame-aligned FCP duration format.
    
    🚨 CRITICAL: Frame alignment is mandatory (from schema.yaml timing rules)
    All durations MUST use 24000/1001 timebase for proper FCP compatibility.
    """
    if seconds == 0:
        return "0s"
    
    # Get timing constants from schema
    schema = get_schema()
    timing = schema['fcpxml_rules']['timing']
    frame_rate = timing['frame_rate']
    timebase = timing['timebase'] 
    frame_duration = timing['frame_duration']
    
    # Calculate exact frame count (round to nearest frame)
    frames = int(seconds * frame_rate + 0.5)
    
    # Convert to FCP's rational format: (frames × 1001)/24000s
    numerator = frames * frame_duration
    
    return f"{numerator}/{timebase}s"