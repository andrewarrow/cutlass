"""
Timeline Generation Algorithms

Complex timeline creation algorithms extracted from main.py for better code organization.
Contains algorithms for generating sophisticated FCPXML timelines with proper validation.
"""

import random
from pathlib import Path

from fcpxml_lib.core.fcpxml import create_media_asset, needs_vertical_scaling
from fcpxml_lib.models.elements import Format, Asset, MediaRep
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration
from fcpxml_lib.utils.ids import generate_uid
from fcpxml_lib.constants import (
    IMAGE_FORMAT_NAME, DEFAULT_IMAGE_WIDTH, DEFAULT_IMAGE_HEIGHT, IMAGE_COLOR_SPACE,
    IMAGE_DURATION, VERTICAL_SCALE_FACTOR
)


def create_edge_tiled_timeline(fcpxml, image_files, background_video, duration, num_lanes, tiles_per_lane):
    """
    Create timeline with images (PNG/JPG) tiled across the visible screen area using proper lane structure.
    
    🚨 CRITICAL: Uses Pattern A (nested elements) for multi-lane visibility with background video.
    Uses Pattern B (separate elements) as fallback when no background video is provided.
    
    Pattern A creates multiple visible lanes in Final Cut Pro by nesting image Video elements
    inside a background AssetClip, following the Go implementation approach.
    """
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    resource_counter = len(fcpxml.resources.assets) + len(fcpxml.resources.formats) + 1
    
    # Create shared format definitions to avoid redundancy
    image_format_id = f"r{resource_counter}"
    resource_counter += 1
    
    # Create shared image format (reused for all PNG tiles)
    shared_image_format = Format(
        id=image_format_id,
        name=IMAGE_FORMAT_NAME,
        width=DEFAULT_IMAGE_WIDTH,
        height=DEFAULT_IMAGE_HEIGHT,
        color_space=IMAGE_COLOR_SPACE
    )
    fcpxml.resources.formats.append(shared_image_format)
    
    # Create the main background element (if background video provided)
    if background_video:
        bg_path = Path(background_video)
        if bg_path.exists():
            print(f"   Adding background video: {bg_path.name}")
            
            # Create background asset
            asset_id = f"r{resource_counter}"
            format_id = f"r{resource_counter + 1}"
            resource_counter += 2
            
            asset, format_obj = create_media_asset(str(bg_path), asset_id, format_id, duration)
            fcpxml.resources.assets.append(asset)
            fcpxml.resources.formats.append(format_obj)
            
            # Determine if background video needs scaling
            is_video = bg_path.suffix.lower() in {'.mov', '.mp4', '.avi', '.mkv', '.m4v'}
            needs_scaling = needs_vertical_scaling(str(bg_path), is_image=not is_video)
            
            # Create background element (use appropriate type based on media)
            bg_duration = convert_seconds_to_fcp_duration(duration)
            
            if is_video:
                # Background is a video - use asset-clip
                bg_element = {
                    "type": "asset-clip",
                    "ref": asset_id,
                    "duration": bg_duration,
                    "offset": "0s",
                    "name": bg_path.stem
                }
            else:
                # Background is an image - use video element
                bg_element = {
                    "type": "video",
                    "ref": asset_id,
                    "duration": bg_duration,
                    "offset": "0s",
                    "start": "0s",  # Required for image video elements
                    "name": bg_path.stem
                }
            
            # Add scaling if needed
            if needs_scaling:
                bg_element["adjust_transform"] = {"scale": VERTICAL_SCALE_FACTOR}
                print(f"   Background video scaled for vertical format")
            
            # Add background to spine
            sequence.spine.ordered_elements.append(bg_element)
            if is_video:
                sequence.spine.asset_clips.append(bg_element)
            else:
                sequence.spine.videos.append(bg_element)
    
    # Generate image tiles as nested elements inside background video (Pattern A - like Go implementation)
    if background_video:
        # Find the background element to add nested images inside it
        bg_element = sequence.spine.ordered_elements[-1]  # Last added element (background)
        
        # Add nested_elements list if not exists  
        if "nested_elements" not in bg_element:
            bg_element["nested_elements"] = []
        
        current_lane = 1
        total_tiles = num_lanes * tiles_per_lane
        
        for tile_index in range(total_tiles):
            # Select random image file
            image_file = random.choice(image_files)
            
            # Create asset using shared image format
            asset_id = f"r{resource_counter}"
            resource_counter += 1
            
            # Create asset manually to use shared format
            abs_path = Path(image_file).resolve()
            uid = generate_uid(f"MEDIA_{abs_path.name}")
            media_rep = MediaRep(src=str(abs_path))
            
            asset = Asset(
                id=asset_id,
                name=abs_path.stem,
                uid=uid,
                duration=IMAGE_DURATION,
                has_video="1",
                format=image_format_id,  # Use shared format
                video_sources="1",
                media_rep=media_rep
            )
            fcpxml.resources.assets.append(asset)
            
            # Generate random position within visible screen area (much tighter bounds)
            x_pos = random.uniform(-30.0, 30.0)  # Narrower X range for better visibility
            y_pos = random.uniform(-50.0, 50.0)  # Narrower Y range for better visibility
            
            # Generate random scale (smaller tiles)
            scale = random.uniform(0.1, 0.5)  # 10% to 50% original size
            
            # Create PNG tile element as nested video inside background (like Go's addSlidingPngImageToAssetClip)
            tile_duration = convert_seconds_to_fcp_duration(duration)
            
            # Use proper timing like Info.fcpxml (start="3600s" pattern)
            tile_offset = "0s"  # When PNG appears relative to background start
            tile_start = "3600s"   # PNG start time like Info.fcpxml
            
            tile_element = {
                "type": "video",
                "ref": asset_id,
                "lane": current_lane,  # Each PNG gets its own lane (like Go implementation)
                "duration": tile_duration,
                "offset": tile_offset,
                "start": tile_start,  # Match Info.fcpxml timing
                "name": f"{image_file.stem}_lane_{current_lane}",
                "adjust_transform": {
                    "position": f"{x_pos:.3f} {y_pos:.3f}",
                    "scale": f"{scale:.3f} {scale:.3f}"
                }
            }
            
            # Add as nested element inside background (Pattern A - like Go/Info.fcpxml)
            bg_element["nested_elements"].append(tile_element)
            
            # Increment lane for next PNG
            current_lane += 1
    else:
        # If no background video, fall back to separate spine elements
        current_lane = 1
        total_tiles = num_lanes * tiles_per_lane
        
        for tile_index in range(total_tiles):
            # Select random image file
            image_file = random.choice(image_files)
            
            # Create asset using shared image format
            asset_id = f"r{resource_counter}"
            resource_counter += 1
            
            # Create asset manually to use shared format
            abs_path = Path(image_file).resolve()
            uid = generate_uid(f"MEDIA_{abs_path.name}")
            media_rep = MediaRep(src=str(abs_path))
            
            asset = Asset(
                id=asset_id,
                name=abs_path.stem,
                uid=uid,
                duration=IMAGE_DURATION,
                has_video="1",
                format=image_format_id,  # Use shared format
                video_sources="1",
                media_rep=media_rep
            )
            fcpxml.resources.assets.append(asset)
            
            # Generate random position within visible screen area (much tighter bounds)
            x_pos = random.uniform(-30.0, 30.0)  # Narrower X range for better visibility
            y_pos = random.uniform(-50.0, 50.0)  # Narrower Y range for better visibility
            
            # Generate random scale (smaller tiles)
            scale = random.uniform(0.1, 0.5)  # 10% to 50% original size
            
            # Create PNG tile element as separate spine element with lane
            tile_duration = convert_seconds_to_fcp_duration(duration)
            
            tile_element = {
                "type": "video",
                "ref": asset_id,
                "lane": current_lane,  # Each PNG gets its own lane
                "duration": tile_duration,
                "offset": "0s",
                "start": "0s",  # Required for image video elements
                "name": f"{image_file.stem}_lane_{current_lane}",
                "adjust_transform": {
                    "position": f"{x_pos:.3f} {y_pos:.3f}",
                    "scale": f"{scale:.3f} {scale:.3f}"
                }
            }
            
            # Add as separate spine element with lane number
            sequence.spine.videos.append(tile_element)
            sequence.spine.ordered_elements.append(tile_element)
            
            # Increment lane for next PNG
            current_lane += 1
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(duration)
    
    total_tiles = num_lanes * tiles_per_lane
    print(f"   Generated {total_tiles} random image tiles, each on its own lane (lanes 1-{total_tiles})")
    if background_video:
        print(f"   Structure: Background AssetClip + {total_tiles} nested Image Video elements (Pattern A - like Go/Info.fcpxml)")
        print(f"   Timing: Background at offset=0s, Images at start=3600s (like Info.fcpxml)")
    else:
        print(f"   Structure: {total_tiles} separate Image spine elements (Pattern B)")
    print(f"   Screen bounds: X(-30.0 to 30.0), Y(-50.0 to 50.0)")
    print(f"   Original request: {num_lanes} lanes × {tiles_per_lane} tiles = {total_tiles} total image lanes")


def create_stress_test_timeline(fcpxml, image_files, video_files):
    """
    Create an extremely complex timeline to stress test all library features.
    
    Stress Test Strategy:
    1. Multiple overlapping background videos (5 segments)
    2. 50+ lanes of content with complex transforms
    3. Mix Pattern A (nested) and Pattern B (separate) elements
    4. Random timing, positions, scales, rotations
    5. All available media files used multiple times
    6. Maximum resource utilization
    7. Edge case validation testing
    """
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    total_duration = 540.0  # 9 minutes
    segment_duration = 108.0  # Each background segment is ~1.8 minutes
    resource_counter = len(fcpxml.resources.assets) + len(fcpxml.resources.formats) + 1
    
    print(f"   Creating timeline with {total_duration}s total duration")
    print(f"   Background segments: 5 segments of {segment_duration}s each")
    
    # Create shared formats for efficiency
    image_format_id = f"r{resource_counter}"
    resource_counter += 1
    
    shared_image_format = Format(
        id=image_format_id,
        name=IMAGE_FORMAT_NAME,
        width=DEFAULT_IMAGE_WIDTH,
        height=DEFAULT_IMAGE_HEIGHT,
        color_space=IMAGE_COLOR_SPACE
    )
    fcpxml.resources.formats.append(shared_image_format)
    
    # Phase 1: Create multiple overlapping background video segments
    print("   Phase 1: Creating 5 overlapping background video segments...")
    
    if video_files:
        for segment_idx in range(5):
            video_file = video_files[segment_idx % len(video_files)]  # Cycle through videos
            segment_offset = segment_idx * (segment_duration * 0.8)  # 20% overlap between segments
            
            print(f"     Segment {segment_idx + 1}: {video_file.name} at offset {segment_offset:.1f}s")
            
            # Create background asset
            asset_id = f"r{resource_counter}"
            format_id = f"r{resource_counter + 1}"
            resource_counter += 2
            
            asset, format_obj = create_media_asset(str(video_file), asset_id, format_id, segment_duration)
            fcpxml.resources.assets.append(asset)
            fcpxml.resources.formats.append(format_obj)
            
            # Determine scaling needs
            needs_scaling = needs_vertical_scaling(str(video_file), is_image=False)
            
            # Create background element
            bg_duration = convert_seconds_to_fcp_duration(segment_duration)
            bg_offset = convert_seconds_to_fcp_duration(segment_offset)
            
            bg_element = {
                "type": "asset-clip",
                "ref": asset_id,
                "duration": bg_duration,
                "offset": bg_offset,
                "name": f"bg_segment_{segment_idx + 1}_{video_file.stem}"
            }
            
            # Add scaling and complex transforms
            if needs_scaling:
                # Add complex layered transforms for stress testing
                bg_element["adjust_transform"] = {
                    "scale": VERTICAL_SCALE_FACTOR,
                    "rotation": str(random.uniform(-5.0, 5.0)),  # Slight rotation
                    "position": f"{random.uniform(-5.0, 5.0)} {random.uniform(-5.0, 5.0)}"  # Slight position offset
                }
            
            # Add to spine
            sequence.spine.ordered_elements.append(bg_element)
            sequence.spine.asset_clips.append(bg_element)
            
            # Phase 2: Add nested content to this background (Pattern A)
            nested_count = random.randint(8, 15)  # 8-15 nested elements per background
            bg_element["nested_elements"] = []
            
            print(f"       Adding {nested_count} nested elements to segment {segment_idx + 1}")
            
            for nested_idx in range(nested_count):
                # Alternate between images and videos for nested content
                if nested_idx % 2 == 0 and image_files:
                    # Use image
                    media_file = random.choice(image_files)
                    is_image = True
                elif video_files:
                    # Use video
                    media_file = random.choice(video_files)
                    is_image = False
                else:
                    # Fallback to image if no videos
                    media_file = random.choice(image_files)
                    is_image = True
                
                # Create nested asset
                nested_asset_id = f"r{resource_counter}"
                resource_counter += 1
                
                if is_image:
                    # Use shared format for images
                    abs_path = Path(media_file).resolve()
                    uid = generate_uid(f"NESTED_IMG_{abs_path.name}_{segment_idx}_{nested_idx}")
                    media_rep = MediaRep(src=str(abs_path))
                    
                    asset = Asset(
                        id=nested_asset_id,
                        name=f"{abs_path.stem}_nested_s{segment_idx}_n{nested_idx}",
                        uid=uid,
                        duration=IMAGE_DURATION,
                        has_video="1",
                        format=image_format_id,
                        video_sources="1",
                        media_rep=media_rep
                    )
                    fcpxml.resources.assets.append(asset)
                    
                    element_type = "video"
                    element_start = "3600s"  # Standard image timing
                else:
                    # Create video asset
                    nested_format_id = f"r{resource_counter}"
                    resource_counter += 1
                    
                    asset, format_obj = create_media_asset(str(media_file), nested_asset_id, nested_format_id, segment_duration * 0.5)
                    fcpxml.resources.assets.append(asset)
                    fcpxml.resources.formats.append(format_obj)
                    
                    element_type = "asset-clip"
                    element_start = None  # Videos don't have start attribute
                
                # Create complex nested element with extreme transforms
                lane_number = (segment_idx * 15) + nested_idx + 1  # Unique lane numbers across segments
                
                nested_element = {
                    "type": element_type,
                    "ref": nested_asset_id,
                    "lane": lane_number,
                    "duration": convert_seconds_to_fcp_duration(segment_duration * random.uniform(0.3, 0.9)),
                    "offset": convert_seconds_to_fcp_duration(random.uniform(0, segment_duration * 0.5)),
                    "name": f"{media_file.stem}_L{lane_number}_S{segment_idx}"
                }
                
                if element_start:
                    nested_element["start"] = element_start
                
                # Add extreme transforms for stress testing
                nested_element["adjust_transform"] = {
                    "position": f"{random.uniform(-100.0, 100.0)} {random.uniform(-150.0, 150.0)}",
                    "scale": f"{random.uniform(0.05, 2.0)} {random.uniform(0.05, 2.0)}",
                    "rotation": str(random.uniform(-180.0, 180.0))
                }
                
                bg_element["nested_elements"].append(nested_element)
    
    # Phase 3: Add separate spine elements (Pattern B) for additional stress
    print("   Phase 3: Adding separate spine elements for additional complexity...")
    
    separate_elements_count = 25  # 25 additional separate elements
    for sep_idx in range(separate_elements_count):
        # Choose random media
        if sep_idx % 3 == 0 and image_files:
            media_file = random.choice(image_files)
            is_image = True
        elif video_files:
            media_file = random.choice(video_files)
            is_image = False
        else:
            media_file = random.choice(image_files) if image_files else random.choice(video_files)
            is_image = media_file.suffix.lower() in {'.png', '.jpg', '.jpeg'}
        
        # Create asset
        sep_asset_id = f"r{resource_counter}"
        resource_counter += 1
        
        if is_image:
            # Use shared format for images
            abs_path = Path(media_file).resolve()
            uid = generate_uid(f"SEP_IMG_{abs_path.name}_{sep_idx}")
            media_rep = MediaRep(src=str(abs_path))
            
            asset = Asset(
                id=sep_asset_id,
                name=f"{abs_path.stem}_separate_{sep_idx}",
                uid=uid,
                duration=IMAGE_DURATION,
                has_video="1",
                format=image_format_id,
                video_sources="1",
                media_rep=media_rep
            )
            fcpxml.resources.assets.append(asset)
            
            element_type = "video"
            element_start = "0s"  # Separate elements use different timing
        else:
            # Create video asset
            sep_format_id = f"r{resource_counter}"
            resource_counter += 1
            
            duration = random.uniform(30.0, 120.0)  # Variable durations
            asset, format_obj = create_media_asset(str(media_file), sep_asset_id, sep_format_id, duration)
            fcpxml.resources.assets.append(asset)
            fcpxml.resources.formats.append(format_obj)
            
            element_type = "asset-clip"
            element_start = None
        
        # Create separate element with complex timing
        lane_number = 100 + sep_idx  # Lane numbers 100+
        element_offset = random.uniform(0, total_duration * 0.8)
        element_duration = random.uniform(20.0, 180.0)
        
        separate_element = {
            "type": element_type,
            "ref": sep_asset_id,
            "lane": lane_number,
            "duration": convert_seconds_to_fcp_duration(element_duration),
            "offset": convert_seconds_to_fcp_duration(element_offset),
            "name": f"{media_file.stem}_separate_L{lane_number}"
        }
        
        if element_start is not None:
            separate_element["start"] = element_start
        
        # Add extreme transforms and effects
        separate_element["adjust_transform"] = {
            "position": f"{random.uniform(-200.0, 200.0)} {random.uniform(-300.0, 300.0)}",
            "scale": f"{random.uniform(0.01, 5.0)} {random.uniform(0.01, 5.0)}",
            "rotation": str(random.uniform(-360.0, 360.0))
        }
        
        # Add to spine
        if element_type == "video":
            sequence.spine.videos.append(separate_element)
        else:
            sequence.spine.asset_clips.append(separate_element)
        sequence.spine.ordered_elements.append(separate_element)
    
    # Update sequence duration to total
    sequence.duration = convert_seconds_to_fcp_duration(total_duration)
    
    total_elements = len(sequence.spine.ordered_elements)
    nested_elements = sum(len(elem.get("nested_elements", [])) for elem in sequence.spine.ordered_elements)
    
    print(f"   STRESS TEST COMPLETE:")
    print(f"     Total spine elements: {total_elements}")
    print(f"     Total nested elements: {nested_elements}")
    print(f"     Total resources created: {resource_counter}")
    print(f"     Maximum lane number: {100 + separate_elements_count}")
    print(f"     Timeline duration: {total_duration}s ({total_duration/60:.1f} minutes)")
    print(f"     Complexity level: EXTREME - Testing validation system limits")


def create_staircase_removal_timeline(fcpxml, background_path, tiles_dir, chunk_duration, total_duration, num_squares):
    """
    Create staircase removal effect where squares are removed in uniform chunks.
    
    Based on the analysis of remove.fcpxml, this creates a timeline where:
    1. Background image is on lane -1 (underneath)
    2. Square tiles are on positive lanes (1, 2, 3, etc.)
    3. Each "chunk" removes more squares progressively 
    4. Uniform chunk duration (fixes the uneven timing in original)
    5. Continues until all squares are removed
    """
    from pathlib import Path
    import os
    
    project = fcpxml.library.events[0].projects[0]
    sequence = project.sequences[0]
    
    resource_counter = len(fcpxml.resources.assets) + len(fcpxml.resources.formats) + 1
    
    # Create background asset
    bg_asset_id = f"r{resource_counter}"
    resource_counter += 1
    bg_format_id = f"r{resource_counter}"  
    resource_counter += 1
    
    background_asset, background_format = create_media_asset(
        background_path, bg_asset_id, bg_format_id, include_audio=False
    )
    fcpxml.resources.assets.append(background_asset)
    fcpxml.resources.formats.append(background_format)
    
    # Discover tile files
    tiles_path = Path(tiles_dir)
    tile_files = sorted([f for f in tiles_path.glob("*.png") if f.is_file()])
    
    if not tile_files:
        raise ValueError(f"No PNG files found in {tiles_dir}")
    
    # Limit to requested number of squares
    tile_files = tile_files[:num_squares]
    
    # Create assets for all tiles
    tile_assets = {}
    shared_tile_format_id = f"r{resource_counter}"
    resource_counter += 1
    
    # Create shared format for all tiles (assuming they're same size)
    shared_tile_format = Format(
        id=shared_tile_format_id,
        name=IMAGE_FORMAT_NAME,
        width=DEFAULT_IMAGE_WIDTH,
        height=DEFAULT_IMAGE_HEIGHT,
        color_space=IMAGE_COLOR_SPACE
    )
    fcpxml.resources.formats.append(shared_tile_format)
    
    # Create assets for each tile
    for tile_file in tile_files:
        tile_asset_id = f"r{resource_counter}"
        resource_counter += 1
        
        tile_asset = Asset(
            id=tile_asset_id,
            name=tile_file.stem,
            uid=generate_uid(),
            start="0s",
            duration=IMAGE_DURATION,
            has_video="1",
            format=shared_tile_format_id,
            video_sources="1"
        )
        tile_asset.media_rep = MediaRep(
            kind="original-media",
            sig=generate_uid(),
            src=f"file://{tile_file.absolute()}"
        )
        fcpxml.resources.assets.append(tile_asset)
        tile_assets[tile_file.stem] = tile_asset_id
    
    # Calculate chunk timing
    num_chunks = int(total_duration / chunk_duration)
    squares_per_chunk = max(1, len(tile_files) // (num_chunks - 1))  # Leave room for final reveal
    
    # Create timeline chunks
    current_offset = 0.0
    remaining_tiles = list(tile_files)
    
    for chunk_idx in range(num_chunks):
        chunk_start_time = convert_seconds_to_fcp_duration(current_offset)
        chunk_duration_fcp = convert_seconds_to_fcp_duration(chunk_duration)
        
        # Determine which tiles are visible in this chunk
        tiles_to_remove = min(squares_per_chunk, len(remaining_tiles))
        if chunk_idx == num_chunks - 1:  # Last chunk removes all remaining
            tiles_to_remove = len(remaining_tiles)
        
        visible_tiles = remaining_tiles[tiles_to_remove:]  # Keep the ones not being removed
        
        # Create main video element with background
        main_video = {
            "type": "video",
            "ref": bg_asset_id,
            "offset": chunk_start_time,
            "name": background_asset.name,
            "start": chunk_start_time,
            "duration": chunk_duration_fcp,
            "nested_elements": []
        }
        
        # Add background on lane -1
        bg_nested = {
            "type": "video", 
            "ref": bg_asset_id,
            "lane": -1,
            "offset": chunk_start_time,
            "name": background_asset.name,
            "start": "3600s",  # Using timing pattern from Info.fcpxml
            "duration": chunk_duration_fcp
        }
        main_video["nested_elements"].append(bg_nested)
        
        # Add visible tiles on positive lanes
        for lane_idx, tile_file in enumerate(visible_tiles):
            tile_asset_id = tile_assets[tile_file.stem]
            
            tile_nested = {
                "type": "video",
                "ref": tile_asset_id,
                "lane": lane_idx + 1,
                "offset": chunk_start_time,
                "name": tile_file.stem,
                "start": chunk_start_time,  
                "duration": chunk_duration_fcp
            }
            main_video["nested_elements"].append(tile_nested)
        
        # Add this chunk to the timeline
        sequence.spine.videos.append(main_video)
        sequence.spine.ordered_elements.append(main_video)
        
        # Remove tiles for next chunk
        remaining_tiles = remaining_tiles[tiles_to_remove:]
        current_offset += chunk_duration
        
        # Stop if no tiles left
        if not remaining_tiles:
            break
    
    # Update sequence duration
    sequence.duration = convert_seconds_to_fcp_duration(total_duration)
    
    print(f"   Created {len(sequence.spine.ordered_elements)} timeline chunks")
    print(f"   Uniform chunk duration: {chunk_duration}s")
    print(f"   Total tiles: {len(tile_files)}")
    print(f"   Squares per chunk: {squares_per_chunk}")