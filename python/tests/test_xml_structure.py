"""
Tests for XML structure and serialization.

Tests that the generated XML follows the correct FCPXML schema and includes
all required elements to prevent FCP crashes.
"""

import pytest
from xml.etree.ElementTree import fromstring
import xml.etree.ElementTree as ET

from fcpxml_lib.core.fcpxml import create_empty_project
from fcpxml_lib.serialization.xml_serializer import serialize_to_xml
from fcpxml_lib.models.elements import SmartCollection


class TestXMLStructure:
    """Test XML structure and schema compliance."""

    def test_fcpxml_root_element(self):
        """Test that root element has correct attributes."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        assert root.tag == "fcpxml"
        assert root.get("version") == "1.13"

    def test_resources_section(self):
        """Test that resources section is properly structured."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        resources = root.find('resources')
        assert resources is not None
        
        # Should have at least one format
        formats = resources.findall('format')
        assert len(formats) >= 1

    def test_library_structure(self):
        """Test that library has required structure."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        library = root.find('library')
        assert library is not None
        assert library.get('location') is not None
        assert library.get('location').startswith('file://')
        
        # Should have at least one event
        events = library.findall('event')
        assert len(events) >= 1
        
        # Event should have project
        event = events[0]
        assert event.get('name') is not None
        assert event.get('uid') is not None
        
        projects = event.findall('project')
        assert len(projects) >= 1

    def test_smart_collections_xml(self):
        """Test smart collections XML structure."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        library = root.find('library')
        smart_collections = library.findall('smart-collection')
        
        assert len(smart_collections) == 5
        
        # Test Projects collection specifically
        projects_collection = None
        for sc in smart_collections:
            if sc.get('name') == 'Projects':
                projects_collection = sc
                break
        
        assert projects_collection is not None
        assert projects_collection.get('match') == 'all'
        
        match_clip = projects_collection.find('match-clip')
        assert match_clip is not None
        assert match_clip.get('rule') == 'is'
        assert match_clip.get('type') == 'project'

    def test_all_video_collection(self):
        """Test All Video smart collection structure."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        library = root.find('library')
        smart_collections = library.findall('smart-collection')
        
        all_video_collection = None
        for sc in smart_collections:
            if sc.get('name') == 'All Video':
                all_video_collection = sc
                break
        
        assert all_video_collection is not None
        assert all_video_collection.get('match') == 'any'
        
        match_media_elements = all_video_collection.findall('match-media')
        assert len(match_media_elements) == 2
        
        # Check for videoOnly and videoWithAudio rules
        rules = [elem.get('type') for elem in match_media_elements]
        assert 'videoOnly' in rules
        assert 'videoWithAudio' in rules

    def test_favorites_collection(self):
        """Test Favorites smart collection uses match-ratings."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        library = root.find('library')
        smart_collections = library.findall('smart-collection')
        
        favorites_collection = None
        for sc in smart_collections:
            if sc.get('name') == 'Favorites':
                favorites_collection = sc
                break
        
        assert favorites_collection is not None
        assert favorites_collection.get('match') == 'all'
        
        match_ratings = favorites_collection.find('match-ratings')
        assert match_ratings is not None
        assert match_ratings.get('value') == 'favorites'

    def test_sequence_structure(self):
        """Test sequence has proper attributes."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        sequence = root.find('.//sequence')
        assert sequence is not None
        
        # Check required attributes
        assert sequence.get('format') is not None
        assert sequence.get('duration') is not None
        assert sequence.get('tcStart') == '0s'
        assert sequence.get('tcFormat') == 'NDF'
        assert sequence.get('audioLayout') == 'stereo'
        assert sequence.get('audioRate') == '48k'
        
        # Should have spine
        spine = sequence.find('spine')
        assert spine is not None

    def test_xml_wellformed(self):
        """Test that generated XML is well-formed."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        # This should not raise an exception if XML is well-formed
        full_xml = f'<?xml version="1.0" encoding="UTF-8"?>{xml_content}'
        try:
            ET.fromstring(full_xml)
        except ET.ParseError as e:
            pytest.fail(f"Generated XML is not well-formed: {e}")

    def test_no_empty_elements(self):
        """Test that there are no unexpected empty elements."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Check that library has children
        library = root.find('library')
        assert len(list(library)) > 0, "Library should not be empty"
        
        # Check that events have children
        events = library.findall('event')
        for event in events:
            assert len(list(event)) > 0, "Events should not be empty"

    def test_uid_uniqueness(self):
        """Test that all UIDs in the document are unique."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Collect all uid attributes
        uids = []
        for elem in root.iter():
            uid = elem.get('uid')
            if uid:
                uids.append(uid)
        
        # Check uniqueness
        assert len(uids) == len(set(uids)), "All UIDs should be unique"

    def test_resource_id_format(self):
        """Test that resource IDs follow the r1, r2, r3... pattern."""
        fcpxml = create_empty_project()
        xml_content = serialize_to_xml(fcpxml)
        
        root = fromstring(f'<?xml version="1.0"?>{xml_content}')
        
        # Collect all id attributes from resources
        resource_ids = []
        resources = root.find('resources')
        if resources is not None:
            for elem in resources:
                id_attr = elem.get('id')
                if id_attr:
                    resource_ids.append(id_attr)
        
        # Check format (should be r1, r2, r3, etc.)
        for resource_id in resource_ids:
            assert resource_id.startswith('r'), f"Resource ID {resource_id} should start with 'r'"
            assert resource_id[1:].isdigit(), f"Resource ID {resource_id} should be r followed by digits"