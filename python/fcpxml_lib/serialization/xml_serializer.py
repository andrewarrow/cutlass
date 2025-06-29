"""
XML serialization for FCPXML documents.
"""

from xml.etree.ElementTree import Element, SubElement, tostring
from xml.dom.minidom import parseString

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from ..models.elements import FCPXML


def serialize_to_xml(fcpxml) -> str:
    """
    Serialize FCPXML to XML string using structured approach.
    
    🚨 CRITICAL: Uses structured XML building with no string templates.
    
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
        
        # Add title effects (if any)
        for title_effect in fcpxml.resources.title_effects:
            effect_elem = SubElement(resources_elem, "effect")
            effect_elem.set("id", title_effect["id"])
            effect_elem.set("name", title_effect["name"])
            effect_elem.set("uid", title_effect["uid"])
            if "src" in title_effect:
                effect_elem.set("src", title_effect["src"])

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
                    
                    # Add spine with media content
                    spine_elem = SubElement(seq_elem, "spine")
                    
                    # 🚨 CRITICAL: Use ordered_elements if available for proper FCP timeline order
                    if hasattr(sequence.spine, 'ordered_elements') and sequence.spine.ordered_elements:
                        # Use pre-sorted elements to maintain timeline order
                        for element in sequence.spine.ordered_elements:
                            if element["type"] == "asset-clip":
                                clip_elem = SubElement(spine_elem, "asset-clip")
                                clip_elem.set("ref", element["ref"])
                                if "duration" in element:
                                    clip_elem.set("duration", element["duration"])
                                if "offset" in element:
                                    clip_elem.set("offset", element["offset"])
                                # NO start attribute for asset-clip elements per samples/simple_video1.fcpxml
                                if "name" in element:
                                    clip_elem.set("name", element["name"])
                                
                                # Add adjust-transform if present
                                if "adjust_transform" in element:
                                    transform_elem = SubElement(clip_elem, "adjust-transform")
                                    if "scale" in element["adjust_transform"]:
                                        transform_elem.set("scale", element["adjust_transform"]["scale"])
                                
                                # Add nested elements (for multi-lane structure) - CRITICAL for asset-clip
                                if "nested_elements" in element:
                                    for nested in element["nested_elements"]:
                                        nested_video_elem = SubElement(clip_elem, "video")
                                        nested_video_elem.set("ref", nested["ref"])
                                        if "lane" in nested:
                                            nested_video_elem.set("lane", str(nested["lane"]))
                                        if "duration" in nested:
                                            nested_video_elem.set("duration", nested["duration"])
                                        if "offset" in nested:
                                            nested_video_elem.set("offset", nested["offset"])
                                        if "start" in nested:
                                            nested_video_elem.set("start", nested["start"])
                                        if "name" in nested:
                                            nested_video_elem.set("name", nested["name"])
                                        
                                        # Add nested transforms
                                        if "adjust_transform" in nested:
                                            nested_transform_elem = SubElement(nested_video_elem, "adjust-transform")
                                            if "scale" in nested["adjust_transform"]:
                                                nested_transform_elem.set("scale", nested["adjust_transform"]["scale"])
                                            if "position" in nested["adjust_transform"]:
                                                nested_transform_elem.set("position", nested["adjust_transform"]["position"])
                            elif element["type"] == "video":
                                video_elem = SubElement(spine_elem, "video")
                                video_elem.set("ref", element["ref"])
                                if "lane" in element:
                                    video_elem.set("lane", str(element["lane"]))
                                if "duration" in element:
                                    video_elem.set("duration", element["duration"])
                                if "offset" in element:
                                    video_elem.set("offset", element["offset"])
                                if "start" in element:
                                    video_elem.set("start", element["start"])  # Video elements need start attribute
                                if "name" in element:
                                    video_elem.set("name", element["name"])
                                
                                # Add adjust-transform if present
                                if "adjust_transform" in element:
                                    transform_elem = SubElement(video_elem, "adjust-transform")
                                    if "scale" in element["adjust_transform"]:
                                        transform_elem.set("scale", element["adjust_transform"]["scale"])
                                    if "position" in element["adjust_transform"]:
                                        transform_elem.set("position", element["adjust_transform"]["position"])
                                
                                # Add nested elements (for multi-lane structure)
                                if "nested_elements" in element:
                                    for nested in element["nested_elements"]:
                                        if nested.get("type") == "title":
                                            # Handle nested title elements within video
                                            nested_title_elem = SubElement(video_elem, "title")
                                            nested_title_elem.set("ref", nested["ref"])
                                            if "lane" in nested:
                                                nested_title_elem.set("lane", str(nested["lane"]))
                                            if "duration" in nested:
                                                nested_title_elem.set("duration", nested["duration"])
                                            if "offset" in nested:
                                                nested_title_elem.set("offset", nested["offset"])
                                            if "start" in nested:
                                                nested_title_elem.set("start", nested["start"])
                                            if "name" in nested:
                                                nested_title_elem.set("name", nested["name"])
                                            
                                            # Add param elements if available
                                            if "params" in nested:
                                                for param in nested["params"]:
                                                    param_elem = SubElement(nested_title_elem, "param")
                                                    param_elem.set("name", param["name"])
                                                    param_elem.set("key", param["key"])
                                                    param_elem.set("value", param["value"])
                                            
                                            # Add text content if available
                                            if "text_content" in nested:
                                                # Generate unique text style ID
                                                from fcpxml_lib.utils.ids import generate_text_style_id
                                                text_style_id = generate_text_style_id()
                                                
                                                text_elem = SubElement(nested_title_elem, "text")
                                                text_style_elem = SubElement(text_elem, "text-style")
                                                text_style_elem.set("ref", text_style_id)
                                                text_style_elem.text = nested["text_content"]
                                                
                                                # Add text style definition
                                                text_style_def = SubElement(nested_title_elem, "text-style-def")
                                                text_style_def.set("id", text_style_id)
                                                text_style = SubElement(text_style_def, "text-style")
                                                if "font_name" in nested:
                                                    text_style.set("font", nested["font_name"])
                                                # 🚨 CRITICAL: Always add fontFace="Regular" to ensure face checkbox is checked
                                                # This matches the working Go implementation pattern
                                                text_style.set("fontFace", nested.get("font_face", "Regular"))
                                                # Use font size from nested data or default to 290
                                                font_size = nested.get("font_size", "290")
                                                text_style.set("fontSize", str(font_size))
                                                # Apply face color only if face is enabled (default True)
                                                if nested.get("face_enabled", True) and "face_color" in nested:
                                                    text_style.set("fontColor", nested["face_color"])
                                                # 🚨 CRITICAL: NO bold attribute - Go version doesn't include it
                                                # Bold attribute actually breaks the face checkbox
                                                if nested.get("italic", False):
                                                    text_style.set("italic", "1")
                                                # 🚨 CRITICAL: Add alignment and lineSpacing to match Go exactly
                                                text_style.set("alignment", "center")
                                                text_style.set("lineSpacing", "-19")
                                                # Apply outline only if outline is enabled (default True)
                                                if nested.get("outline_enabled", True) and "outline_color" in nested:
                                                    text_style.set("strokeColor", nested["outline_color"])
                                                    # 🚨 CRITICAL: Use negative stroke width like Go version
                                                    stroke_width = nested.get("stroke_width", "15")
                                                    text_style.set("strokeWidth", f"-{stroke_width}")
                                        else:
                                            # Handle nested video elements
                                            nested_video_elem = SubElement(video_elem, "video")
                                            nested_video_elem.set("ref", nested["ref"])
                                            if "lane" in nested:
                                                nested_video_elem.set("lane", str(nested["lane"]))
                                            if "duration" in nested:
                                                nested_video_elem.set("duration", nested["duration"])
                                            if "offset" in nested:
                                                nested_video_elem.set("offset", nested["offset"])
                                            if "start" in nested:
                                                nested_video_elem.set("start", nested["start"])
                                            if "name" in nested:
                                                nested_video_elem.set("name", nested["name"])
                                            
                                            # Add nested transforms
                                            if "adjust_transform" in nested:
                                                nested_transform_elem = SubElement(nested_video_elem, "adjust-transform")
                                                if "scale" in nested["adjust_transform"]:
                                                    nested_transform_elem.set("scale", nested["adjust_transform"]["scale"])
                                                if "position" in nested["adjust_transform"]:
                                                    nested_transform_elem.set("position", nested["adjust_transform"]["position"])
                            
                            elif element["type"] == "title":
                                title_elem = SubElement(spine_elem, "title")
                                title_elem.set("ref", element["ref"])
                                if "duration" in element:
                                    title_elem.set("duration", element["duration"])
                                if "start" in element:
                                    title_elem.set("start", element["start"])
                                if "offset" in element:
                                    title_elem.set("offset", element["offset"])
                                if "name" in element:
                                    title_elem.set("name", element["name"])
                                if "lane" in element:
                                    title_elem.set("lane", str(element["lane"]))
                    else:
                        # Fallback to old method if ordered_elements not available
                        # Add asset-clips (for videos)
                        for asset_clip in sequence.spine.asset_clips:
                            clip_elem = SubElement(spine_elem, "asset-clip")
                            clip_elem.set("ref", asset_clip["ref"])
                            if "duration" in asset_clip:
                                clip_elem.set("duration", asset_clip["duration"])
                            if "offset" in asset_clip:
                                clip_elem.set("offset", asset_clip["offset"])
                            # NO start attribute for asset-clip elements per samples/simple_video1.fcpxml
                            if "name" in asset_clip:
                                clip_elem.set("name", asset_clip["name"])
                            
                            # Add adjust-transform if present
                            if "adjust_transform" in asset_clip:
                                transform_elem = SubElement(clip_elem, "adjust-transform")
                                if "scale" in asset_clip["adjust_transform"]:
                                    transform_elem.set("scale", asset_clip["adjust_transform"]["scale"])
                        
                        # Add videos (for images)
                        for video in sequence.spine.videos:
                            video_elem = SubElement(spine_elem, "video")
                            video_elem.set("ref", video["ref"])
                            if "lane" in video:
                                video_elem.set("lane", str(video["lane"]))
                            if "duration" in video:
                                video_elem.set("duration", video["duration"])
                            if "offset" in video:
                                video_elem.set("offset", video["offset"])
                            if "start" in video:
                                video_elem.set("start", video["start"])  # Video elements need start attribute
                            if "name" in video:
                                video_elem.set("name", video["name"])
                            
                            # Add adjust-transform if present
                            if "adjust_transform" in video:
                                transform_elem = SubElement(video_elem, "adjust-transform")
                                if "scale" in video["adjust_transform"]:
                                    transform_elem.set("scale", video["adjust_transform"]["scale"])
                                if "position" in video["adjust_transform"]:
                                    transform_elem.set("position", video["adjust_transform"]["position"])
                            
                            # Add nested elements (for multi-lane structure)
                            if "nested_elements" in video:
                                for nested in video["nested_elements"]:
                                    nested_video_elem = SubElement(video_elem, "video")
                                    nested_video_elem.set("ref", nested["ref"])
                                    if "lane" in nested:
                                        nested_video_elem.set("lane", str(nested["lane"]))
                                    if "duration" in nested:
                                        nested_video_elem.set("duration", nested["duration"])
                                    if "offset" in nested:
                                        nested_video_elem.set("offset", nested["offset"])
                                    if "start" in nested:
                                        nested_video_elem.set("start", nested["start"])
                                    if "name" in nested:
                                        nested_video_elem.set("name", nested["name"])
                                    
                                    # Add nested transforms
                                    if "adjust_transform" in nested:
                                        nested_transform_elem = SubElement(nested_video_elem, "adjust-transform")
                                        if "scale" in nested["adjust_transform"]:
                                            nested_transform_elem.set("scale", nested["adjust_transform"]["scale"])
                                        if "position" in nested["adjust_transform"]:
                                            nested_transform_elem.set("position", nested["adjust_transform"]["position"])
                    
                    # Add gaps (if any)
                    for gap in sequence.spine.gaps:
                        gap_elem = SubElement(spine_elem, "gap")
                        if "duration" in gap:
                            gap_elem.set("duration", gap["duration"])
                        if "start" in gap:
                            gap_elem.set("start", gap["start"])
                    
                    # Add titles (if any)
                    for title in sequence.spine.titles:
                        title_elem = SubElement(spine_elem, "title")
                        if "ref" in title:
                            title_elem.set("ref", title["ref"])
                        if "duration" in title:
                            title_elem.set("duration", title["duration"])
                        if "start" in title:
                            title_elem.set("start", title["start"])
                        if "offset" in title:
                            title_elem.set("offset", title["offset"])
                        if "name" in title:
                            title_elem.set("name", title["name"])
                        if "lane" in title:
                            title_elem.set("lane", str(title["lane"]))
        
        # Add smart collections (required by FCP)
        for smart_collection in fcpxml.library.smart_collections:
            collection_elem = SubElement(library_elem, "smart-collection")
            collection_elem.set("name", smart_collection.name)
            collection_elem.set("match", smart_collection.match)
            
            for rule in smart_collection.rules:
                if "rule" in rule and "type" in rule:
                    match_elem = SubElement(collection_elem, "match-clip" if rule.get("type") == "project" else "match-media")
                    match_elem.set("rule", rule["rule"])
                    match_elem.set("type", rule["type"])
                elif "rule" in rule and "value" in rule:
                    match_elem = SubElement(collection_elem, "match-ratings")
                    match_elem.set("value", rule["value"])

    # Convert to string without XML declaration
    rough_string = tostring(root, encoding='unicode')
    reparsed = parseString(rough_string)
    pretty_xml = reparsed.toprettyxml(indent="  ", encoding=None)
    
    # Remove the first XML declaration line that parseString adds
    lines = pretty_xml.split('\n')
    if lines[0].startswith('<?xml'):
        lines = lines[1:]
    
    return '\n'.join(lines).strip()