#!/usr/bin/env python3
"""
Remove Squares Command

Creates a staircase removal effect where squares are removed in uniform chunks,
revealing the background image underneath. Fixes the uneven timing in the first
few chunks and continues the pattern until all squares are removed.
"""

import sys
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml
)
from fcpxml_lib.generators.timeline_generators import create_staircase_removal_timeline


def remove_sq_cmd(args):
    """Create staircase removal effect with uniform chunk timing"""
    
    # Validate background image
    background_path = Path(args.background_image)
    if not background_path.exists():
        print(f"❌ Background image not found: {background_path}")
        sys.exit(1)
    
    # Validate tiles directory
    tiles_dir = Path(args.tiles_dir)
    if not tiles_dir.exists() or not tiles_dir.is_dir():
        print(f"❌ Tiles directory not found: {tiles_dir}")
        sys.exit(1)
    
    print(f"🎨 Creating staircase removal effect...")
    print(f"   Background image: {background_path}")
    print(f"   Tiles directory: {tiles_dir}")
    print(f"   Chunk duration: {args.chunk_duration}s")
    print(f"   Total duration: {args.total_duration}s")
    
    # Create empty project (vertical format like original)
    fcpxml = create_empty_project(
        project_name="Square Removal",
        event_name="Staircase Effect",
        use_horizontal=False  # Use vertical format
    )
    
    # Generate the staircase removal timeline
    try:
        create_staircase_removal_timeline(
            fcpxml,
            str(background_path),
            str(tiles_dir),
            args.chunk_duration,
            args.total_duration,
            args.num_squares
        )
        
        print(f"✅ Timeline created with {args.num_squares} squares in staircase removal pattern")
        
    except Exception as e:
        print(f"❌ Error creating staircase removal timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "remove_sq.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)