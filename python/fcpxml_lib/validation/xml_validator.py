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
    - Nested video element structure (causes FCP crashes)
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
            # Skip text-style refs - they are locally scoped within title elements
            if 'ref' in element.attrib and element.tag != 'text-style':
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
        
        # Check for problematic nested video structures
        video_nesting_error = validate_video_nesting(root)
        if video_nesting_error:
            return False, video_nesting_error
        
        # 🚨 CRITICAL: Check frame boundary alignment
        frame_error = validate_frame_boundaries(root)
        if frame_error:
            return False, frame_error
        
        # 🚨 CRITICAL: Check for FCP-specific media validation issues
        media_error = validate_media_clip_association(root)
        if media_error:
            return False, media_error
        
        return True, ""
        
    except ET.ParseError as e:
        return False, f"XML parsing error: {e}"
    except Exception as e:
        return False, f"Semantic validation error: {e}"


def validate_video_nesting(root_element) -> str:
    """
    Validate video element nesting to prevent FCP crashes.
    
    🚨 CRITICAL: Deeply nested video elements cause "Invalid edit with no respective media" errors
    
    According to BAFFLE_TWO.md, nested video structures can cause crashes.
    This function detects problematic patterns.
    
    Returns:
        str: Error message if problem found, empty string if valid
    """
    def count_video_nesting(element, depth=0):
        max_depth = depth
        
        for child in element:
            if child.tag == 'video':
                child_depth = count_video_nesting(child, depth + 1)
                max_depth = max(max_depth, child_depth)
            else:
                child_depth = count_video_nesting(child, depth)
                max_depth = max(max_depth, child_depth)
        
        return max_depth
    
    def find_problematic_nesting(element, path=""):
        problems = []
        
        # Check if this video element has many nested videos
        if element.tag == 'video':
            nested_videos = element.findall('.//video')  # Find all nested video elements
            # ❌ REMOVED: Bogus 20-element limit - FCP can handle hundreds of lanes
            # Final Cut Pro has no practical limit on number of video lanes/nested elements
            
            # Check if nested videos have deeply nested structure
            depth = count_video_nesting(element)
            if depth > 2:  # More than 2 levels of nesting is problematic
                problems.append(f"Video element at {path} has nesting depth {depth} (limit: 2)")
        
        # Recursively check children
        for i, child in enumerate(element):
            child_path = f"{path}/{child.tag}[{i}]" if path else f"{child.tag}[{i}]"
            problems.extend(find_problematic_nesting(child, child_path))
        
        return problems
    
    problems = find_problematic_nesting(root_element)
    if problems:
        return "Problematic video nesting detected: " + "; ".join(problems)
    
    return ""


def validate_frame_boundaries(root_element) -> str:
    """
    Validate that all timing attributes are on frame boundaries.
    
    🚨 CRITICAL: Frame boundary violations cause "The item is not on an edit frame boundary" errors
    
    FCP requires all timing values (offset, duration, start) to be aligned to frame boundaries
    using the standard 23.976fps timebase (1001/24000s per frame).
    
    Returns:
        str: Error message if frame boundary violation found, empty string if valid
    """
    from ..constants import STANDARD_TIMEBASE
    
    def is_frame_aligned(time_str: str) -> bool:
        """Check if a time string like '49049/24000s' is frame-aligned."""
        if not time_str or time_str == "0s":
            return True
            
        # Parse rational format: "numerator/denominator s"
        if time_str.endswith('s'):
            time_str = time_str[:-1]  # Remove 's'
            
        if '/' in time_str:
            try:
                numerator, denominator = map(int, time_str.split('/'))
                # For frame alignment, numerator must be divisible by 1001
                # and denominator must be STANDARD_TIMEBASE (24000)
                if denominator != STANDARD_TIMEBASE:
                    return False
                # Frame-aligned if numerator is multiple of 1001
                return numerator % 1001 == 0
            except ValueError:
                return False
        else:
            # Simple format like "2s" - always frame aligned for integer seconds
            try:
                float(time_str)
                return True
            except ValueError:
                return False
    
    def find_element_path(element, root, current_path=""):
        """Build XPath-like path by traversing from root to find element."""
        def build_path(current_elem, target_elem, path_parts):
            if current_elem == target_elem:
                return "/fcpxml[1]" + "".join(path_parts)
            
            for i, child in enumerate(current_elem):
                # Count siblings of same tag for position
                siblings_of_same_tag = [c for c in current_elem if c.tag == child.tag]
                position = siblings_of_same_tag.index(child) + 1 if child in siblings_of_same_tag else 1
                
                child_path = f"/{child.tag}[{position}]"
                result = build_path(child, target_elem, path_parts + [child_path])
                if result:
                    return result
            return None
        
        return build_path(root, element, [])
    
    def check_timing_attributes(element, root, path=""):
        """Recursively check all timing attributes in the tree."""
        errors = []
        
        # Check timing attributes
        timing_attrs = ['offset', 'duration', 'start']
        for attr in timing_attrs:
            if attr in element.attrib:
                value = element.attrib[attr]
                if not is_frame_aligned(value):
                    element_path = find_element_path(element, root)
                    errors.append(f'The item is not on an edit frame boundary ({attr}="{value}": {element_path}/@{attr})')
        
        # Recursively check children
        for child in element:
            errors.extend(check_timing_attributes(child, root, path))
            
        return errors
    
    errors = check_timing_attributes(root_element, root_element)
    if errors:
        return "; ".join(errors)
    
    return ""


def validate_media_clip_association(root_element) -> str:
    """
    Validate clip-to-media associations to prevent "Invalid edit with no respective media" errors.
    
    🚨 CRITICAL: FCP requires specific patterns for clip elements to be valid:
    - Clip elements must have format attributes that reference valid format resources
    - Video elements within clips must reference valid asset resources  
    - Assets must have media-rep elements with valid file URLs
    - Clip elements should have conform-rate elements
    
    Returns:
        str: Error message if problem found, empty string if valid
    """
    
    def find_element_path(element, root, current_path=""):
        """Build XPath-like path for error reporting."""
        def build_path(current_elem, target_elem, path_parts):
            if current_elem == target_elem:
                return "/fcpxml[1]" + "".join(path_parts)
            
            for i, child in enumerate(current_elem):
                siblings_of_same_tag = [c for c in current_elem if c.tag == child.tag]
                position = siblings_of_same_tag.index(child) + 1 if child in siblings_of_same_tag else 1
                child_path = f"/{child.tag}[{position}]"
                result = build_path(child, target_elem, path_parts + [child_path])
                if result:
                    return result
            return None
        
        return build_path(root, element, [])
    
    # Collect all resources for lookup
    resources = root_element.find('resources')
    if not resources:
        return "Invalid edit with no respective media: No resources section found"
    
    assets = {}
    formats = {}
    
    for child in resources:
        if child.tag == 'asset' and 'id' in child.attrib:
            assets[child.attrib['id']] = child
        elif child.tag == 'format' and 'id' in child.attrib:
            formats[child.attrib['id']] = child
    
    errors = []
    
    def check_clip_validity(element, path=""):
        """Check if clip elements are properly structured for FCP."""
        
        if element.tag == 'clip':
            clip_path = find_element_path(element, root_element)
            
            # Check 1: Clip should have format attribute
            if 'format' not in element.attrib:
                errors.append(f"Invalid edit with no respective media: Clip missing format attribute ({clip_path})")
            else:
                format_id = element.attrib['format']
                if format_id not in formats:
                    errors.append(f"Invalid edit with no respective media: Clip references unknown format '{format_id}' ({clip_path})")
            
            # Check 2: Clip should have conform-rate element
            conform_rate = element.find('conform-rate')
            if conform_rate is None:
                errors.append(f"Invalid edit with no respective media: Clip missing conform-rate element ({clip_path})")
            
            # Check 3: Video elements within clip should reference valid assets
            for video in element.findall('.//video'):
                if 'ref' not in video.attrib:
                    video_path = find_element_path(video, root_element)
                    errors.append(f"Invalid edit with no respective media: Video element missing ref attribute ({video_path})")
                else:
                    asset_id = video.attrib['ref']
                    if asset_id not in assets:
                        video_path = find_element_path(video, root_element)
                        errors.append(f"Invalid edit with no respective media: Video references unknown asset '{asset_id}' ({video_path})")
                    else:
                        # Check 4: Asset should have media-rep with valid src
                        asset = assets[asset_id]
                        media_rep = asset.find('media-rep')
                        if media_rep is None:
                            errors.append(f"Invalid edit with no respective media: Asset '{asset_id}' missing media-rep element")
                        elif 'src' not in media_rep.attrib:
                            errors.append(f"Invalid edit with no respective media: Asset '{asset_id}' media-rep missing src attribute")
                        elif not media_rep.attrib['src'].startswith('file://'):
                            src = media_rep.attrib['src']
                            errors.append(f"Invalid edit with no respective media: Asset '{asset_id}' media-rep src should start with 'file://' (got: {src})")
        
        # Recursively check children
        for child in element:
            check_clip_validity(child, path)
    
    check_clip_validity(root_element)
    
    if errors:
        return "; ".join(errors)
    
    return ""