"""
ID generation utilities for FCPXML.
"""

import time
import hashlib
import threading


def generate_uid(prefix: str = "") -> str:
    """Generate a unique identifier following FCPXML conventions"""
    timestamp = str(int(time.time() * 1000000))  # microsecond precision
    source = f"{prefix}-{timestamp}"
    return hashlib.md5(source.encode()).hexdigest().upper()


# Thread-safe resource ID counter
_resource_id_counter = 0
_resource_id_lock = threading.Lock()

# Thread-safe text style ID counter
_text_style_id_counter = 0
_text_style_id_lock = threading.Lock()


def generate_resource_id() -> str:
    """Generate sequential resource IDs (r1, r2, r3, ...)"""
    global _resource_id_counter
    with _resource_id_lock:
        _resource_id_counter += 1
        return f"r{_resource_id_counter}"


def set_resource_id_counter(start_value: int) -> None:
    """Set the resource ID counter to start from a specific value"""
    global _resource_id_counter
    with _resource_id_lock:
        _resource_id_counter = start_value


def generate_text_style_id() -> str:
    """Generate sequential text style IDs (ts1, ts2, ts3, ...)"""
    global _text_style_id_counter
    with _text_style_id_lock:
        _text_style_id_counter += 1
        return f"ts{_text_style_id_counter}"