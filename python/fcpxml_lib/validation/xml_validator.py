"""
XML validation utilities.
"""

import subprocess
import xml.etree.ElementTree as ET


def run_xml_validation(xml_file_path: str) -> tuple[bool, str]:
    """
    Run comprehensive XML validation for FCPXML files.
    
    🚨 CRITICAL: XML must be well-formed AND semantically valid for FCPXML crash prevention
    
    Performs:
    1. XML well-formedness validation using xmllint
    2. Semantic validation (ref integrity, required elements)
    """
    # Step 1: XML well-formedness validation
    try:
        result = subprocess.run(
            ['xmllint', '--noout', xml_file_path],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode != 0:
            error_msg = result.stderr.strip()
            return False, f"XML well-formedness error: {error_msg}"
            
    except subprocess.TimeoutExpired:
        return False, "XML validation timed out"
    except subprocess.CalledProcessError as e:
        return False, f"xmllint error: {e.stderr}"
    except FileNotFoundError:
        return False, "xmllint not found - install libxml2-utils"
    
    # Step 2: Semantic validation
    semantic_valid, semantic_error = validate_fcpxml_semantics(xml_file_path)
    if not semantic_valid:
        return False, semantic_error
    
    return True, ""


def validate_fcpxml_semantics(xml_file_path: str) -> tuple[bool, str]:
    """
    Validate FCPXML semantic correctness.
    
    🚨 CRITICAL: Prevents "Invalid edit with no respective media" errors
    
    Checks:
    - All ref attributes have corresponding asset/format/effect IDs
    - Required smart collections are present
    - No duplicate IDs
    """
    try:
        tree = ET.parse(xml_file_path)
        root = tree.getroot()
        
        # Collect all available IDs from resources
        available_ids = set()
        resources = root.find('resources')
        if resources is not None:
            for child in resources:
                if 'id' in child.attrib:
                    available_ids.add(child.attrib['id'])
        
        # Collect all ref attributes used in timeline and format references in assets
        used_refs = set()
        
        def collect_refs(element):
            if 'ref' in element.attrib:
                used_refs.add(element.attrib['ref'])
            if 'format' in element.attrib:
                used_refs.add(element.attrib['format'])
            # Check nested elements recursively
            for child in element:
                collect_refs(child)
        
        # Start from root and collect all refs (including from resources section)
        collect_refs(root)
        
        # Find missing references
        missing_refs = used_refs - available_ids
        if missing_refs:
            missing_list = ', '.join(sorted(missing_refs))
            return False, f"Invalid edit with no respective media. Missing resource IDs: {missing_list}"
        
        # Check for duplicate IDs
        all_ids = []
        if resources is not None:
            for child in resources:
                if 'id' in child.attrib:
                    all_ids.append(child.attrib['id'])
        
        duplicate_ids = set()
        seen_ids = set()
        for id_val in all_ids:
            if id_val in seen_ids:
                duplicate_ids.add(id_val)
            seen_ids.add(id_val)
        
        if duplicate_ids:
            duplicate_list = ', '.join(sorted(duplicate_ids))
            return False, f"Duplicate resource IDs found: {duplicate_list}"
        
        # Check for required smart collections
        library = root.find('library')
        smart_collections = library.findall('smart-collection') if library is not None else []
        required_collections = {'Projects', 'All Video', 'Audio Only', 'Stills', 'Favorites'}
        found_collections = {sc.get('name', '') for sc in smart_collections}
        missing_collections = required_collections - found_collections
        
        if missing_collections:
            missing_list = ', '.join(sorted(missing_collections))
            return False, f"Missing required smart collections: {missing_list}"
        
        return True, ""
        
    except ET.ParseError as e:
        return False, f"XML parsing error: {e}"
    except Exception as e:
        return False, f"Semantic validation error: {e}"