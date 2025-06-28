"""
ID generation utilities for FCPXML.
"""

import time
import hashlib


def generate_uid(prefix: str = "") -> str:
    """Generate a unique identifier following FCPXML conventions"""
    timestamp = str(int(time.time() * 1000000))  # microsecond precision
    source = f"{prefix}-{timestamp}"
    return hashlib.md5(source.encode()).hexdigest().upper()