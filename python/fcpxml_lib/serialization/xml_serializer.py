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
                            elif element["type"] == "video":
                                video_elem = SubElement(spine_elem, "video")
                                video_elem.set("ref", element["ref"])
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