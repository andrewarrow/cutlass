#!/usr/bin/env python3
"""
Remove Squares Command

Rewritten to use the blading + enabled="0" approach from 9.fcpxml.
Takes background.fcpxmld and creates progressive square removal by:
1. Blading the timeline at regular intervals 
2. Using enabled="0" to disable squares progressively
3. Keeping exact structure but with cumulative square disabling
"""

import sys
import random
from pathlib import Path
import xml.etree.ElementTree as ET

from fcpxml_lib import create_empty_project, save_fcpxml
from fcpxml_lib.models.elements import Asset, Format, MediaRep
from fcpxml_lib.utils.ids import generate_uid
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration


def parse_fcpxml_structure(fcpxml_path):
    """Parse background.fcpxmld to extract the complete structure"""
    try:
        tree = ET.parse(fcpxml_path)
        root = tree.getroot()
        
        # Extract assets information
        assets = {}
        formats = {}
        
        # Parse resources
        resources = root.find('.//resources')
        if resources is None:
            raise ValueError("No resources section found in FCPXML")
        
        # Parse formats
        for format_elem in resources.findall('format'):
            format_id = format_elem.get('id')
            formats[format_id] = {
                'id': format_id,
                'name': format_elem.get('name', ''),
                'width': format_elem.get('width', ''),
                'height': format_elem.get('height', ''),
                'color_space': format_elem.get('colorSpace', ''),
                'frame_duration': format_elem.get('frameDuration', '')
            }
        
        # Parse assets
        for asset_elem in resources.findall('asset'):
            asset_id = asset_elem.get('id')
            media_rep = asset_elem.find('media-rep')
            if media_rep is not None:
                src = media_rep.get('src', '')
                assets[asset_id] = {
                    'id': asset_id,
                    'name': asset_elem.get('name', ''),
                    'src': src,
                    'format': asset_elem.get('format', ''),
                    'duration': asset_elem.get('duration', '0s'),
                    'has_video': asset_elem.get('hasVideo', '1'),
                    'video_sources': asset_elem.get('videoSources', '1'),
                    'uid': asset_elem.get('uid', '')
                }
        
        # Parse spine structure to get the exact layout
        spine = root.find('.//spine')
        if spine is None:
            raise ValueError("No spine section found in FCPXML")
        
        # Find the main video element with nested elements
        main_video = spine.find('video')
        if main_video is None:
            raise ValueError("No main video element found in spine")
        
        # Extract the main video properties and all nested elements
        main_video_data = {
            'ref': main_video.get('ref'),
            'name': main_video.get('name', ''),
            'start': main_video.get('start', ''),
            'duration': main_video.get('duration', ''),
            'position': None,
            'scale': None
        }
        
        # Get main video transform if present
        main_transform = main_video.find('adjust-transform')
        if main_transform is not None:
            main_video_data['position'] = main_transform.get('position')
            main_video_data['scale'] = main_transform.get('scale')
        
        # Parse all nested elements (background and tiles)
        nested_elements = []
        for nested in main_video.findall('video'):
            lane = nested.get('lane')
            ref = nested.get('ref')
            name = nested.get('name', '')
            
            element_data = {
                'ref': ref,
                'lane': int(lane) if lane else 0,
                'name': name,
                'offset': nested.get('offset', ''),
                'start': nested.get('start', ''),
                'duration': nested.get('duration', ''),
                'position': None,
                'scale': None
            }
            
            # Get transform data
            transform = nested.find('adjust-transform')
            if transform is not None:
                element_data['position'] = transform.get('position')
                element_data['scale'] = transform.get('scale')
            
            nested_elements.append(element_data)
        
        # Separate background from tiles
        background_element = None
        tile_elements = []
        
        # The main video itself is also a tile (col0_row0)
        main_tile = {
            'ref': main_video_data['ref'],
            'lane': 0,  # Lane 0 for the main video tile
            'name': main_video_data['name'],
            'offset': '',
            'start': main_video_data['start'],
            'duration': main_video_data['duration'],
            'position': main_video_data['position'],
            'scale': main_video_data['scale']
        }
        tile_elements.append(main_tile)
        
        for element in nested_elements:
            if element['lane'] == -1:
                background_element = element
            elif element['lane'] > 0:
                tile_elements.append(element)
        
        # Sort tiles by lane number
        tile_elements.sort(key=lambda x: x['lane'])
        
        return {
            'assets': assets,
            'formats': formats,
            'main_video': main_video_data,
            'background_element': background_element,
            'tile_elements': tile_elements
        }
        
    except Exception as e:
        raise ValueError(f"Failed to parse FCPXML file: {e}")


def create_bladed_removal_timeline(fcpxml, structure, total_duration=9.0):
    """Create timeline using blading + enabled=0 approach like 9.fcpxml"""
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    # Get the parsed structure
    assets_data = structure['assets']
    formats_data = structure['formats']
    main_video_data = structure['main_video']
    background_element = structure['background_element']
    tile_elements = structure['tile_elements']
    
    num_tiles = len(tile_elements)
    print(f"Found {num_tiles} tile squares to disable progressively")
    
    # Import assets into the new FCPXML
    asset_id_mapping = {}
    format_id_mapping = {}
    resource_counter = 2  # Start after the default format r1
    
    # Import formats first
    for old_format_id, format_data in formats_data.items():
        new_format_id = f"r{resource_counter}"
        resource_counter += 1
        format_id_mapping[old_format_id] = new_format_id
        
        new_format = Format(
            id=new_format_id,
            name=format_data['name'],
            width=format_data['width'],
            height=format_data['height'],
            color_space=format_data['color_space']
        )
        if format_data['frame_duration']:
            new_format.frame_duration = format_data['frame_duration']
        
        fcpxml.resources.formats.append(new_format)
    
    # Import assets
    for old_asset_id, asset_data in assets_data.items():
        new_asset_id = f"r{resource_counter}"
        resource_counter += 1
        asset_id_mapping[old_asset_id] = new_asset_id
        
        # Map the format ID
        old_format_id = asset_data['format']
        new_format_id = format_id_mapping.get(old_format_id, old_format_id)
        
        new_asset = Asset(
            id=new_asset_id,
            name=asset_data['name'],
            uid=asset_data['uid'] or generate_uid(asset_data['name']),
            start="0s",
            duration=asset_data['duration'],
            has_video=asset_data['has_video'],
            format=new_format_id,
            video_sources=asset_data['video_sources']
        )
        
        # Create media rep
        new_asset.media_rep = MediaRep(
            kind="original-media",
            sig=generate_uid(),
            src=asset_data['src']
        )
        
        fcpxml.resources.assets.append(new_asset)
    
    # Calculate blade intervals: fit all squares in 9 seconds exactly
    blade_interval = total_duration / num_tiles  # Remove one square per interval
    print(f"Blade interval: {blade_interval:.3f} seconds ({9}/{num_tiles} = {9/num_tiles:.3f}s per square)")
    
    # Create removal schedule: random order
    tile_indices = list(range(num_tiles))
    random.shuffle(tile_indices)
    print(f"Random removal order: {[tile_elements[i]['name'] for i in tile_indices]}")
    
    # Generate bladed segments following 9.fcpxml pattern exactly
    disabled_tile_names = set()
    
    for segment_idx in range(num_tiles):
        segment_start = segment_idx * blade_interval
        
        # Disable one new tile at each segment
        tile_to_disable_idx = tile_indices[segment_idx]
        tile_to_disable_name = tile_elements[tile_to_disable_idx]['name']
        disabled_tile_names.add(tile_to_disable_name)
        print(f"  Segment {segment_idx + 1} ({segment_start:.2f}s): disabling {tile_to_disable_name}")
        
        # Calculate timing values exactly like 9.fcpxml
        segment_start_fcp = convert_seconds_to_fcp_duration(segment_start)
        segment_duration_fcp = convert_seconds_to_fcp_duration(blade_interval)
        
        # Calculate synchronized start time (all elements start together)
        # Using the 9.fcpxml pattern: base time + segment offset
        base_time = 3600.0  # Base like 9.fcpxml uses 86399313/24000s ≈ 3600s
        sync_start_time = base_time + segment_start
        sync_start_fcp = convert_seconds_to_fcp_duration(sync_start_time)
        
        # Create main video element for this segment
        main_asset_id = asset_id_mapping[main_video_data['ref']]
        
        main_video = {
            "type": "video",
            "ref": main_asset_id,
            "offset": segment_start_fcp,
            "name": main_video_data['name'],
            "start": sync_start_fcp,
            "duration": segment_duration_fcp,
            "nested_elements": []
        }
        
        # Check if main video should be disabled by name
        if main_video_data['name'] in disabled_tile_names:
            main_video["enabled"] = "0"
        
        # Add position/scale transform if present
        if main_video_data['position'] or main_video_data['scale']:
            main_video["adjust_transform"] = {}
            if main_video_data['position']:
                main_video["adjust_transform"]["position"] = main_video_data['position']
            if main_video_data['scale']:
                main_video["adjust_transform"]["scale"] = main_video_data['scale']
        
        # Add ALL elements (background + tiles) with synchronized timing
        # Background element
        if background_element:
            bg_asset_id = asset_id_mapping[background_element['ref']]
            
            bg_nested = {
                "type": "video",
                "ref": bg_asset_id,
                "lane": background_element['lane'],
                "offset": sync_start_fcp,  # Synchronized timing
                "name": background_element['name'],
                "start": sync_start_fcp,   # All start at same time
                "duration": segment_duration_fcp
            }
            main_video["nested_elements"].append(bg_nested)
        
        # Add all tile elements with cumulative disabling
        for tile_idx, tile in enumerate(tile_elements):
            # Skip lane 0 (main video) as it's already handled above
            if tile['lane'] == 0:
                continue
                
            tile_asset_id = asset_id_mapping[tile['ref']]
            
            tile_nested = {
                "type": "video",
                "ref": tile_asset_id,
                "lane": tile['lane'],
                "offset": sync_start_fcp,  # Synchronized timing
                "name": tile['name'],
                "start": sync_start_fcp,   # All start at same time
                "duration": segment_duration_fcp
            }
            
            # Disable if this tile's name should be disabled
            if tile['name'] in disabled_tile_names:
                tile_nested["enabled"] = "0"
            
            # Add transform if present
            if tile['position'] or tile['scale']:
                tile_nested["adjust_transform"] = {}
                if tile['position']:
                    tile_nested["adjust_transform"]["position"] = tile['position']
                if tile['scale']:
                    tile_nested["adjust_transform"]["scale"] = tile['scale']
            
            main_video["nested_elements"].append(tile_nested)
        
        # Add this segment to the timeline
        sequence.spine.videos.append(main_video)
        sequence.spine.ordered_elements.append(main_video)
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(total_duration)


def remove_sq_cmd(args):
    """Create progressive square removal using blading + enabled=0 approach"""
    
    # Validate input FCPXML file
    input_path = Path(args.input_fcpxml)
    if not input_path.exists():
        print(f"❌ Input FCPXML file not found: {input_path}")
        sys.exit(1)
    
    print(f"🎨 Creating progressive square removal using blade + disable approach...")
    print(f"   Input FCPXML: {input_path}")
    print(f"   Duration: 9 seconds")
    
    # Parse the input FCPXML structure
    try:
        structure = parse_fcpxml_structure(input_path)
        print(f"   Found main video: {structure['main_video']['name']}")
        print(f"   Found background: {structure['background_element']['name'] if structure['background_element'] else 'None'}")
        print(f"   Found {len(structure['tile_elements'])} tile squares")
    except Exception as e:
        print(f"❌ Error parsing input FCPXML: {e}")
        sys.exit(1)
    
    # Create empty project (vertical format to match input)
    fcpxml = create_empty_project(
        project_name="Progressive Square Removal",
        event_name="Blade + Disable Effect",
        use_horizontal=False
    )
    
    # Generate the bladed removal timeline
    try:
        create_bladed_removal_timeline(fcpxml, structure, total_duration=9.0)
        print(f"✅ Timeline created with bladed progressive removal")
        
    except Exception as e:
        print(f"❌ Error creating timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path("remove_sq_bladed.fcpxml")
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)