#!/usr/bin/env python3
"""
Create Empty Project Command

Generates an empty FCPXML project with proper validation and crash prevention.
Following comprehensive crash prevention rules and NO_XML_TEMPLATES principle.
"""

import sys
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml, ValidationError,
    Sequence
)


def test_validation_failure():
    """Demonstrate validation failure detection"""
    print("\n🧪 Testing validation failure detection...")
    
    # Try to create a sequence with invalid audio rate
    try:
        bad_sequence = Sequence(
            format="r1",
            audio_rate="48000"  # Invalid - should be "48k"
        )
        print("❌ SHOULD HAVE FAILED - invalid audio rate was not caught!")
    except ValidationError as e:
        print(f"✅ Validation correctly caught error: {e}")


def create_empty_project_cmd(args):
    """Create an empty FCPXML project"""
    print("🎬 Creating empty FCPXML project...")
    print("Following crash prevention rules for safe FCPXML generation")
    print()
    
    # Test validation system first
    test_validation_failure()
    
    # Create empty project with format choice
    format_desc = "1280x720 horizontal" if args.horizontal else "1080x1920 vertical"
    print(f"   Format: {format_desc}")
    
    fcpxml = create_empty_project(
        project_name=args.project_name or "My First Project",
        event_name=args.event_name or "My First Event",
        use_horizontal=args.horizontal
    )
    
    # Validate the project
    print("✅ FCPXML structure created and validated")
    print(f"   Version: {fcpxml.version}")
    print(f"   Resources: {len(fcpxml.resources.formats)} formats")
    print(f"   Events: {len(fcpxml.library.events)}")
    print(f"   Projects: {len(fcpxml.library.events[0].projects)}")
    print()
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent.parent.parent / "empty_project.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
        print("🎯 Next steps:")
        print("1. Import into Final Cut Pro to test")
        print("2. Extend this library to add media assets") 
        print("3. Implement more spine elements (asset-clips, titles, etc.)")
        print("4. Add keyframe animation support")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)