"""
Core FCPXML document handling.
"""

import sys
from pathlib import Path

from ..models.elements import Resources, Library, Format, Sequence, Project, Event, FCPXML
from ..utils.schema_loader import get_schema
from ..serialization.xml_serializer import serialize_to_xml
from ..validation.xml_validator import run_xml_validation


def create_empty_project(project_name: str = "New Project", event_name: str = "New Event") -> FCPXML:
    """
    Create an empty FCPXML project following schema.yaml template.
    
    This generates the minimal valid FCPXML structure that will import into Final Cut Pro
    without errors. Based on the empty_project template in schema.yaml.
    """
    
    # Get template from schema
    schema = get_schema()
    template = schema['fcpxml_rules']['templates']['empty_project']
    
    # Create format resource (HD 1080p 23.976 fps - FCP standard)
    format_def = Format(
        id="r1",
        name="FFVideoFormat1080p2398",
        frame_duration="1001/24000s",
        width="1920", 
        height="1080",
        color_space="1-1-1 (Rec. 709)"
    )
    
    # Create sequence with the format
    sequence = Sequence(
        format="r1",
        duration="0s",
        tc_start="0s",
        tc_format="NDF",
        audio_layout="stereo",
        audio_rate="48k"
    )
    
    # Create project containing the sequence
    project = Project(
        name=project_name,
        sequences=[sequence]
    )
    
    # Create event containing the project
    event = Event(
        name=event_name,
        projects=[project]
    )
    
    # Create library containing the event
    library = Library(
        events=[event]
    )
    
    # Create resources with the format
    resources = Resources(
        formats=[format_def]
    )
    
    # Create root FCPXML document
    fcpxml = FCPXML(
        version=template['version'],
        resources=resources,
        library=library
    )
    
    return fcpxml


def save_fcpxml(fcpxml: FCPXML, output_path: str) -> bool:
    """
    Save FCPXML document to file and validate it.
    
    Returns True if successful and well-formed, False otherwise.
    🚨 CRITICAL: XML validation is mandatory (from schema.yaml)
    """
    xml_content = serialize_to_xml(fcpxml)
    
    # Add XML declaration (no DTD for now as it requires Apple's server)
    fcpxml_with_header = f'''<?xml version="1.0" encoding="UTF-8"?>
{xml_content}'''
    
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(fcpxml_with_header)
    
    print(f"📄 FCPXML saved to: {output_path}")
    
    # Run basic XML validation
    print("🔍 Running XML well-formedness validation...")
    is_valid, error_msg = run_xml_validation(output_path)
    
    if is_valid:
        print("✅ XML VALIDATION PASSED")
        print("⚠️  Note: For full DTD validation, test import in Final Cut Pro")
        return True
    else:
        print("\n" + "="*60)
        print("🚨 VALIDATION FAILED - XML ERRORS DETECTED")
        print("="*60)
        print(f"❌ XML Error: {error_msg}")
        print("\n⚠️  FCPXML will likely fail to import into Final Cut Pro!")
        print("   Fix the validation errors before using this file.")
        print("="*60 + "\n")
        return False