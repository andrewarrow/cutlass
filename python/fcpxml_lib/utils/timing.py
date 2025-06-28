"""
Timing and duration utilities for FCPXML.
"""

from ..constants import STANDARD_FRAME_RATE, STANDARD_TIMEBASE


def convert_seconds_to_fcp_duration(seconds: float) -> str:
    """
    Convert seconds to frame-aligned FCP duration format.
    
    🚨 CRITICAL: Frame alignment is mandatory for proper FCP compatibility.
    All durations MUST use 24000/1001 timebase for proper FCP compatibility.
    """
    if seconds == 0:
        return "0s"
    
    # Calculate exact frame count (round to nearest frame)
    frames = int(seconds * STANDARD_FRAME_RATE + 0.5)
    
    # Convert to FCP's rational format: (frames × 1001)/24000s
    # frame_duration = 1001 (derived from 1001/24000s)
    numerator = frames * 1001
    
    return f"{numerator}/{STANDARD_TIMEBASE}s"