"""
FCPXML Command Line Interface Package

Each command is implemented in a separate module for scalability.
This package contains CLI command implementations that handle argument processing,
user interaction, and output formatting while delegating business logic to other modules.

Command Structure:
- Each command has its own file (e.g., create_empty_project.py)
- Each command file exports a main function that takes parsed args
- Main function handles CLI concerns (output, error handling, file paths)
- Business logic is delegated to appropriate fcpxml_lib modules

Command Naming Convention:
- File names use underscores: create_empty_project.py
- Function names match: create_empty_project_cmd(args)
- Command names use hyphens: create-empty-project

Modules:
- create_empty_project: Create empty FCPXML project
- create_random_video: Generate random video from media files  
- video_at_edge: Create multi-lane edge-tiled video
- stress_test: Generate complex stress test timeline
- random_font: Create video with random font title elements
- assemble: Assemble media files matching a pattern into timeline
- many_video_fx: Create tiling animation effect with videos moving from center to grid positions
"""

from .create_empty_project import create_empty_project_cmd
from .create_random_video import create_random_video_cmd
from .video_at_edge import video_at_edge_cmd
from .stress_test import stress_test_cmd
from .random_font import random_font_cmd
from .assemble import assemble_cmd
from .many_video_fx import many_video_fx_cmd

__all__ = [
    'create_empty_project_cmd',
    'create_random_video_cmd', 
    'video_at_edge_cmd',
    'stress_test_cmd',
    'random_font_cmd',
    'assemble_cmd',
    'many_video_fx_cmd'
]