#!/usr/bin/env python3
"""
FCPXML Python Library - Simple Implementation
Generates FCPXML documents following the comprehensive rules from schema.yaml

Based on the Go and Swift implementations, this library ensures:
- Frame-aligned timing calculations  
- Proper media type handling
- Validated resource ID management
- Crash pattern prevention

🚨 CRITICAL: This follows the "NO_XML_TEMPLATES" rule from schema.yaml
All XML is generated from structured data objects, never string templates.
"""

import os
import time
import yaml
import subprocess
import sys
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any
from xml.etree.ElementTree import Element, SubElement, tostring
from xml.dom.minidom import parseString
from pathlib import Path


# Load the FCPXML rules schema
def load_schema() -> Dict[str, Any]:
    """Load the FCPXML rules schema from schema.yaml"""
    schema_path = Path(__file__).parent / "schema.yaml"
    with open(schema_path, 'r') as f:
        return yaml.safe_load(f)

SCHEMA = load_schema()


class FCPXMLError(Exception):
    """Base exception for FCPXML generation errors"""
    pass


class ValidationError(FCPXMLError):
    """Raised when validation rules are violated"""
    pass


def generate_uid(prefix: str = "") -> str:
    """Generate a unique identifier following FCPXML conventions"""
    import hashlib
    timestamp = str(int(time.time() * 1000000))  # microsecond precision
    source = f"{prefix}-{timestamp}"
    return hashlib.md5(source.encode()).hexdigest().upper()


def convert_seconds_to_fcp_duration(seconds: float) -> str:
    """
    Convert seconds to frame-aligned FCP duration format.
    
    🚨 CRITICAL: Frame alignment is mandatory (from schema.yaml timing rules)
    All durations MUST use 24000/1001 timebase for proper FCP compatibility.
    """
    if seconds == 0:
        return "0s"
    
    # Get timing constants from schema
    timing = SCHEMA['fcpxml_rules']['timing']
    frame_rate = timing['frame_rate']
    timebase = timing['timebase'] 
    frame_duration = timing['frame_duration']
    
    # Calculate exact frame count (round to nearest frame)
    frames = int(seconds * frame_rate + 0.5)
    
    # Convert to FCP's rational format: (frames × 1001)/24000s
    numerator = frames * frame_duration
    
    return f"{numerator}/{timebase}s"


def validate_frame_alignment(duration: str) -> bool:
    """Validate that a duration string is frame-aligned according to schema rules"""
    timing = SCHEMA['fcpxml_rules']['timing']
    
    if duration == "0s":
        return True
        
    if not duration.endswith("s"):
        return False
        
    if "/" not in duration:
        return False
        
    try:
        time_part = duration.rstrip("s")
        numerator, denominator = map(int, time_part.split("/"))
        
        return (denominator == timing['timebase'] and 
                numerator % timing['frame_duration'] == 0)
    except (ValueError, IndexError):
        return False


def validate_resource_id(resource_id: str) -> bool:
    """Validate resource ID follows FCP pattern (r1, r2, etc.)"""
    import re
    pattern = SCHEMA['fcpxml_rules']['resource_ids']['pattern']
    return bool(re.match(pattern, resource_id))


def validate_audio_rate(audio_rate: str) -> bool:
    """Validate audio rate is in DTD enumerated set"""
    valid_rates = SCHEMA['fcpxml_rules']['audio_rates']['valid_values']
    return audio_rate in valid_rates


def run_xml_validation(xml_file_path: str) -> tuple[bool, str]:
    """
    Run basic XML well-formedness validation using xmllint.
    
    🚨 CRITICAL: XML must be well-formed for FCPXML (from schema.yaml)
    Note: Full DTD validation requires Apple's DTD but basic validation catches most issues.
    """
    try:
        result = subprocess.run(
            ['xmllint', '--noout', xml_file_path],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            return True, ""
        else:
            # Extract meaningful error from xmllint output
            error_msg = result.stderr.strip()
            return False, error_msg
            
    except subprocess.TimeoutExpired:
        return False, "XML validation timed out"
    except subprocess.CalledProcessError as e:
        return False, f"xmllint error: {e.stderr}"
    except FileNotFoundError:
        return False, "xmllint not found - install libxml2-utils"


@dataclass
class MediaRep:
    """Media representation with file path and metadata"""
    kind: str = "original-media"
    sig: str = ""
    src: str = ""
    
    def __post_init__(self):
        if not self.src.startswith("file://"):
            # Ensure absolute file:// URL format
            abs_path = os.path.abspath(self.src) if self.src else ""
            self.src = f"file://{abs_path}"


@dataclass 
class Format:
    """Video/audio format definition"""
    id: str
    name: str = ""
    frame_duration: Optional[str] = None
    width: Optional[str] = None
    height: Optional[str] = None
    color_space: Optional[str] = None
    
    def __post_init__(self):
        if not validate_resource_id(self.id):
            raise ValidationError(f"Invalid format ID: {self.id}")
        
        if self.frame_duration and not validate_frame_alignment(self.frame_duration):
            raise ValidationError(f"Frame duration not aligned: {self.frame_duration}")


@dataclass
class Asset:
    """Media asset (video, image, audio)"""
    id: str
    name: str
    uid: str
    start: str = "0s"
    duration: str = "0s"
    has_video: Optional[str] = None
    format: Optional[str] = None
    video_sources: Optional[str] = None
    has_audio: Optional[str] = None
    audio_sources: Optional[str] = None
    audio_channels: Optional[str] = None
    audio_rate: Optional[str] = None
    media_rep: Optional[MediaRep] = None
    
    def __post_init__(self):
        if not validate_resource_id(self.id):
            raise ValidationError(f"Invalid asset ID: {self.id}")
            
        if not validate_frame_alignment(self.duration):
            raise ValidationError(f"Asset duration not frame-aligned: {self.duration}")
            
        if not validate_frame_alignment(self.start):
            raise ValidationError(f"Asset start time not frame-aligned: {self.start}")


@dataclass
class Resources:
    """Container for all shared resources"""
    assets: List[Asset] = field(default_factory=list)
    formats: List[Format] = field(default_factory=list)
    effects: List[Dict] = field(default_factory=list)
    media: List[Dict] = field(default_factory=list)


@dataclass
class Spine:
    """Main timeline container - currently empty for minimal implementation"""
    asset_clips: List[Dict] = field(default_factory=list)
    videos: List[Dict] = field(default_factory=list) 
    titles: List[Dict] = field(default_factory=list)
    gaps: List[Dict] = field(default_factory=list)


@dataclass
class Sequence:
    """Timeline sequence definition"""
    format: str
    duration: str = "0s"
    tc_start: str = "0s"
    tc_format: str = "NDF"
    audio_layout: str = "stereo"
    audio_rate: str = "48k"  # Use DTD-valid enumerated value
    spine: Spine = field(default_factory=Spine)
    
    def __post_init__(self):
        if not validate_frame_alignment(self.duration):
            raise ValidationError(f"Sequence duration not frame-aligned: {self.duration}")
        if not validate_frame_alignment(self.tc_start):
            raise ValidationError(f"Sequence tc_start not frame-aligned: {self.tc_start}")
        if not validate_audio_rate(self.audio_rate):
            valid_rates = SCHEMA['fcpxml_rules']['audio_rates']['valid_values']
            raise ValidationError(f"Invalid audio rate: {self.audio_rate}. Must be one of {valid_rates}")


@dataclass
class Project:
    """FCPXML Project definition"""
    name: str
    uid: str = ""
    mod_date: str = ""
    sequences: List[Sequence] = field(default_factory=list)
    
    def __post_init__(self):
        if not self.uid:
            self.uid = generate_uid("PROJECT")
        if not self.mod_date:
            self.mod_date = time.strftime("%Y-%m-%d %H:%M:%S %z")


@dataclass
class Event:
    """FCPXML Event definition"""
    name: str
    uid: str = ""
    projects: List[Project] = field(default_factory=list)
    
    def __post_init__(self):
        if not self.uid:
            self.uid = generate_uid("EVENT")


@dataclass
class Library:
    """FCPXML Library definition"""
    location: str = ""
    events: List[Event] = field(default_factory=list)


@dataclass
class FCPXML:
    """
    Root FCPXML document.
    
    🚨 CRITICAL: Follows schema.yaml principles:
    - Uses structured data objects (NO_XML_TEMPLATES)
    - Validates frame alignment
    - Ensures proper resource management
    """
    version: str = "1.11"
    resources: Resources = field(default_factory=Resources)
    library: Optional[Library] = None
    
    def __post_init__(self):
        # Validate version follows schema pattern
        version_pattern = SCHEMA['fcpxml_rules']['elements']['fcpxml']['version_pattern']
        import re
        if not re.match(version_pattern, self.version):
            raise ValidationError(f"Invalid FCPXML version: {self.version}")


def create_empty_project(project_name: str = "New Project", event_name: str = "New Event") -> FCPXML:
    """
    Create an empty FCPXML project following schema.yaml template.
    
    This generates the minimal valid FCPXML structure that will import into Final Cut Pro
    without errors. Based on the empty_project template in schema.yaml.
    """
    
    # Get template from schema
    template = SCHEMA['fcpxml_rules']['templates']['empty_project']
    
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


def serialize_to_xml(fcpxml: FCPXML) -> str:
    """
    Serialize FCPXML to XML string using structured approach.
    
    🚨 CRITICAL: This implements the "STRUCT_BASED_GENERATION" principle
    from schema.yaml - no string templates, only structured XML building.
    
    Returns only the XML content without declaration (added separately).
    """
    
    # Create root element
    root = Element("fcpxml")
    root.set("version", fcpxml.version)
    
    # Add resources
    if fcpxml.resources:
        resources_elem = SubElement(root, "resources")
        
        # Add formats
        for fmt in fcpxml.resources.formats:
            format_elem = SubElement(resources_elem, "format")
            format_elem.set("id", fmt.id)
            if fmt.name:
                format_elem.set("name", fmt.name)
            if fmt.frame_duration:
                format_elem.set("frameDuration", fmt.frame_duration)
            if fmt.width:
                format_elem.set("width", fmt.width)
            if fmt.height:
                format_elem.set("height", fmt.height)
            if fmt.color_space:
                format_elem.set("colorSpace", fmt.color_space)
        
        # Add assets (if any)
        for asset in fcpxml.resources.assets:
            asset_elem = SubElement(resources_elem, "asset")
            asset_elem.set("id", asset.id)
            asset_elem.set("name", asset.name)
            asset_elem.set("uid", asset.uid)
            asset_elem.set("start", asset.start)
            asset_elem.set("duration", asset.duration)
            
            if asset.has_video:
                asset_elem.set("hasVideo", asset.has_video)
            if asset.format:
                asset_elem.set("format", asset.format)
            if asset.video_sources:
                asset_elem.set("videoSources", asset.video_sources)
            if asset.has_audio:
                asset_elem.set("hasAudio", asset.has_audio)
            if asset.audio_sources:
                asset_elem.set("audioSources", asset.audio_sources)
            if asset.audio_channels:
                asset_elem.set("audioChannels", asset.audio_channels)
            if asset.audio_rate:
                asset_elem.set("audioRate", asset.audio_rate)
                
            if asset.media_rep:
                media_rep_elem = SubElement(asset_elem, "media-rep")
                media_rep_elem.set("kind", asset.media_rep.kind)
                if asset.media_rep.sig:
                    media_rep_elem.set("sig", asset.media_rep.sig)
                media_rep_elem.set("src", asset.media_rep.src)
    
    # Add library
    if fcpxml.library:
        library_elem = SubElement(root, "library")
        if fcpxml.library.location:
            library_elem.set("location", fcpxml.library.location)
            
        # Add events
        for event in fcpxml.library.events:
            event_elem = SubElement(library_elem, "event")
            event_elem.set("name", event.name)
            if event.uid:
                event_elem.set("uid", event.uid)
                
            # Add projects
            for project in event.projects:
                project_elem = SubElement(event_elem, "project")
                project_elem.set("name", project.name)
                if project.uid:
                    project_elem.set("uid", project.uid)
                if project.mod_date:
                    project_elem.set("modDate", project.mod_date)
                    
                # Add sequences
                for sequence in project.sequences:
                    seq_elem = SubElement(project_elem, "sequence")
                    seq_elem.set("format", sequence.format)
                    seq_elem.set("duration", sequence.duration)
                    seq_elem.set("tcStart", sequence.tc_start)
                    seq_elem.set("tcFormat", sequence.tc_format)
                    seq_elem.set("audioLayout", sequence.audio_layout)
                    seq_elem.set("audioRate", sequence.audio_rate)
                    
                    # Add spine (empty for now)
                    spine_elem = SubElement(seq_elem, "spine")
                    # TODO: Add spine content when implementing media elements
    
    # Convert to string without XML declaration
    rough_string = tostring(root, encoding='unicode')
    reparsed = parseString(rough_string)
    pretty_xml = reparsed.toprettyxml(indent="  ", encoding=None)
    
    # Remove the first XML declaration line that parseString adds
    lines = pretty_xml.split('\n')
    if lines[0].startswith('<?xml'):
        lines = lines[1:]
    
    return '\n'.join(lines).strip()


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