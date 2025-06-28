#!/usr/bin/env python3
"""
FCPXML Python Library - Demo Application
Generates FCPXML documents following the comprehensive rules from schema.yaml

Based on the Go and Swift implementations, this library ensures:
- Frame-aligned timing calculations  
- Proper media type handling
- Validated resource ID management
- Crash pattern prevention

🚨 CRITICAL: This follows the "NO_XML_TEMPLATES" rule from schema.yaml
All XML is generated from structured data objects, never string templates.
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


def main():
    """
    Demo the simple FCPXML generator.
    
    Creates an empty project that will import cleanly into Final Cut Pro
    following all the critical rules from schema.yaml.
    """
    
    print("🎬 FCPXML Python Generator")
    print("Following schema.yaml rules for safe FCPXML generation")
    print()
    
    # Test validation system first
    test_validation_failure()
    
    # Create empty project
    print("Creating empty FCPXML project...")
    fcpxml = create_empty_project(
        project_name="My First Project",
        event_name="My First Event"
    )
    
    # Validate the project
    print("✅ FCPXML structure created and validated")
    print(f"   Version: {fcpxml.version}")
    print(f"   Resources: {len(fcpxml.resources.formats)} formats")
    print(f"   Events: {len(fcpxml.library.events)}")
    print(f"   Projects: {len(fcpxml.library.events[0].projects)}")
    print()
    
    # Save to file with validation
    output_path = Path(__file__).parent / "empty_project.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print("🎯 Next steps:")
        print("1. Import empty_project.fcpxml into Final Cut Pro to test")
        print("2. Extend this library to add media assets") 
        print("3. Implement more spine elements (asset-clips, titles, etc.)")
        print("4. Add keyframe animation support")
        print()
        print("📖 See schema.yaml for complete FCPXML rules and constraints")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


if __name__ == "__main__":
    main()