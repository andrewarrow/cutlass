"""
FCPXML Timeline Generation Modules

This package contains complex timeline generation algorithms moved from main.py
to maintain clean separation between CLI interface and business logic.

Modules:
- timeline_generators: Complex timeline creation algorithms
"""

from .timeline_generators import create_edge_tiled_timeline, create_stress_test_timeline

__all__ = ['create_edge_tiled_timeline', 'create_stress_test_timeline']