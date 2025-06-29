#!/usr/bin/env python3
"""
Stress Test Command

Creates an extremely complex 9-minute stress test video to validate library robustness.
Following comprehensive crash prevention rules and NO_XML_TEMPLATES principle.
"""

import sys
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml
)
from fcpxml_lib.generators.timeline_generators import create_stress_test_timeline
from fcpxml_lib.utils.media import discover_all_media_files


def stress_test_cmd(args):
    """Create an extremely complex 9-minute stress test video to validate library robustness"""
    print("🔥 Creating FCPXML STRESS TEST - 9 minutes of maximum complexity...")
    print("   Testing all library features, validation, and edge cases")
    print("   Format: 1080x1920 vertical")
    print("   Duration: 9 minutes (540 seconds)")
    print("   Goal: Generate invalid FCPXML or prove validation system integrity")
    print()
    
    # Get all available assets
    assets_dir = Path(__file__).parent.parent.parent.parent / "assets"
    if not assets_dir.exists():
        print(f"❌ Assets directory not found: {assets_dir}")
        sys.exit(1)
    
    # Find all media files using utility function
    image_files, video_files = discover_all_media_files(assets_dir)
    
    print(f"   Found {len(image_files)} images: {[f.name for f in image_files]}")
    print(f"   Found {len(video_files)} videos: {[f.name for f in video_files]}")
    
    if not image_files and not video_files:
        print(f"❌ No media files found in {assets_dir}")
        sys.exit(1)
    
    # Create base project
    fcpxml = create_empty_project(
        project_name="FCPXML Stress Test - Maximum Complexity",
        event_name="Stress Test Validation",
        use_horizontal=False  # Always vertical 1080x1920 as requested
    )
    
    try:
        create_stress_test_timeline(fcpxml, image_files, video_files)
        print("✅ Stress test timeline created successfully")
    except Exception as e:
        print(f"❌ Error creating stress test timeline: {e}")
        print("   This indicates a potential library issue or validation gap")
        raise
    
    # Save with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "stress_test.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Stress test saved to: {output_path}")
        print("🎯 VALIDATION SUCCESS: Library generated valid FCPXML under extreme conditions")
        print("   Next steps:")
        print("   1. Validate XML with: xmllint --noout stress_test.fcpxml")
        print("   2. Import into Final Cut Pro to test import stability")
        print("   3. Play timeline to test performance with complex content")
    else:
        print("❌ VALIDATION FAILED: Library rejected its own output")
        print("   This indicates validation system is working correctly")
        sys.exit(1)