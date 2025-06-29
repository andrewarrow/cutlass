#!/usr/bin/env python3
"""
FCPXML Python Library - Demo Application
Generates FCPXML documents following comprehensive crash prevention rules

Based on the Go and Swift implementations, this library ensures:
- Frame-aligned timing calculations  
- Proper media type handling
- Validated resource ID management
- Crash pattern prevention

🚨 CRITICAL: This follows the "NO_XML_TEMPLATES" rule
All XML is generated from structured data objects, never string templates.
"""

import sys
import argparse
import random
from pathlib import Path

from fcpxml_lib import (
    create_empty_project, save_fcpxml, add_media_to_timeline, ValidationError,
    Sequence
)
from fcpxml_lib.core.fcpxml import create_media_asset
from fcpxml_lib.constants import (
    SCREEN_EDGE_LEFT, SCREEN_EDGE_RIGHT, SCREEN_EDGE_TOP, SCREEN_EDGE_BOTTOM,
    SCREEN_WIDTH, SCREEN_HEIGHT
)
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration


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


def create_empty_project_cmd(args):
    """Create an empty FCPXML project"""
    print("🎬 Creating empty FCPXML project...")
    print("Following crash prevention rules for safe FCPXML generation")
    print()
    
    # Test validation system first
    test_validation_failure()
    
    # Create empty project with format choice
    format_desc = "1280x720 horizontal" if args.horizontal else "1080x1920 vertical"
    print(f"   Format: {format_desc}")
    
    fcpxml = create_empty_project(
        project_name=args.project_name or "My First Project",
        event_name=args.event_name or "My First Event",
        use_horizontal=args.horizontal
    )
    
    # Validate the project
    print("✅ FCPXML structure created and validated")
    print(f"   Version: {fcpxml.version}")
    print(f"   Resources: {len(fcpxml.resources.formats)} formats")
    print(f"   Events: {len(fcpxml.library.events)}")
    print(f"   Projects: {len(fcpxml.library.events[0].projects)}")
    print()
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent / "empty_project.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
        print("🎯 Next steps:")
        print("1. Import into Final Cut Pro to test")
        print("2. Extend this library to add media assets") 
        print("3. Implement more spine elements (asset-clips, titles, etc.)")
        print("4. Add keyframe animation support")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


def create_random_video_cmd(args):
    """Create a random video from media files in a directory"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Supported media extensions
    video_extensions = {'.mov', '.mp4', '.avi', '.mkv', '.m4v'}
    image_extensions = {'.jpg', '.jpeg', '.png', '.tiff', '.bmp', '.gif'}
    all_extensions = video_extensions | image_extensions
    
    # Find all media files
    media_files = []
    for ext in all_extensions:
        media_files.extend(input_dir.glob(f"*{ext}"))
        media_files.extend(input_dir.glob(f"*{ext.upper()}"))
    
    if not media_files:
        print(f"❌ No media files found in {input_dir}")
        print(f"   Supported formats: {', '.join(sorted(all_extensions))}")
        sys.exit(1)
    
    # Randomly shuffle the files
    random.shuffle(media_files)
    
    format_desc = "1280x720 horizontal" if args.horizontal else "1080x1920 vertical"
    print(f"🎬 Creating random video from {len(media_files)} media files...")
    print(f"   Input directory: {input_dir}")
    print(f"   Format: {format_desc}")
    print(f"   Files found: {[f.name for f in media_files[:5]]}{'...' if len(media_files) > 5 else ''}")
    
    # Create empty project with format choice
    fcpxml = create_empty_project(
        project_name=args.project_name or f"Random Video - {input_dir.name}",
        event_name=args.event_name or "Random Videos",
        use_horizontal=args.horizontal
    )
    
    # Add media files to timeline
    media_file_paths = [str(f) for f in media_files]
    clip_duration = args.clip_duration
    
    print(f"✅ Adding {len(media_files)} media files to timeline...")
    print(f"   Each clip duration: {clip_duration}s")
    print(f"   Found {len([f for f in media_files if f.suffix.lower() in video_extensions])} videos")
    print(f"   Found {len([f for f in media_files if f.suffix.lower() in image_extensions])} images")
    
    try:
        add_media_to_timeline(fcpxml, media_file_paths, clip_duration, args.horizontal)
        print(f"✅ Timeline created with {len(media_files)} clips")
        
        # Calculate total duration
        total_duration = len(media_files) * clip_duration
        print(f"   Total timeline duration: {total_duration:.1f}s")
        
    except Exception as e:
        print(f"❌ Error adding media to timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent / "random_video.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


def video_at_edge_cmd(args):
    """Create video with random images (PNG/JPG) tiled across visible area on multiple lanes"""
    input_dir = Path(args.input_dir)
    if not input_dir.exists() or not input_dir.is_dir():
        print(f"❌ Directory not found: {input_dir}")
        sys.exit(1)
    
    # Find all image files (PNG and JPG)
    image_files = []
    for pattern in ['*.png', '*.PNG', '*.jpg', '*.JPG', '*.jpeg', '*.JPEG']:
        image_files.extend(input_dir.glob(pattern))
    
    if not image_files:
        print(f"❌ No image files found in {input_dir}")
        print(f"   Supported formats: PNG, JPG, JPEG")
        sys.exit(1)
    
    print(f"🎨 Creating video with edge-tiled images...")
    print(f"   Input directory: {input_dir}")
    print(f"   Image files found: {len(image_files)}")
    print(f"   Duration: {args.duration}s")
    print(f"   Lanes: {args.num_lanes}")
    print(f"   Tiles per lane: {args.tiles_per_lane}")
    
    # Create empty project (always vertical for edge detection)
    fcpxml = create_empty_project(
        project_name="Video at Edge",
        event_name="Edge Tiled Videos",
        use_horizontal=False  # Always use vertical format
    )
    
    # Generate the edge-tiled timeline
    try:
        create_edge_tiled_timeline(
            fcpxml, 
            image_files, 
            args.background_video, 
            args.duration,
            args.num_lanes,
            args.tiles_per_lane
        )
        
        total_tiles = args.num_lanes * args.tiles_per_lane
        print(f"✅ Timeline created with {total_tiles} image tiles across {args.num_lanes} lanes")
        
    except Exception as e:
        print(f"❌ Error creating edge-tiled timeline: {e}")
        print("   Creating empty project instead")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent / "video_at_edge.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)


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
    from fcpxml_lib.models.elements import Format
    from fcpxml_lib.constants import IMAGE_FORMAT_NAME, DEFAULT_IMAGE_WIDTH, DEFAULT_IMAGE_HEIGHT, IMAGE_COLOR_SPACE
    
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
            from fcpxml_lib.core.fcpxml import needs_vertical_scaling
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
                from fcpxml_lib.constants import VERTICAL_SCALE_FACTOR
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
            from fcpxml_lib.models.elements import Asset, MediaRep
            from fcpxml_lib.utils.ids import generate_uid
            from fcpxml_lib.constants import IMAGE_DURATION
            
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
            from fcpxml_lib.models.elements import Asset, MediaRep
            from fcpxml_lib.utils.ids import generate_uid
            from fcpxml_lib.constants import IMAGE_DURATION
            
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


def stress_test_cmd(args):
    """Create an extremely complex 9-minute stress test video to validate library robustness"""
    print("🔥 Creating FCPXML STRESS TEST - 9 minutes of maximum complexity...")
    print("   Testing all library features, validation, and edge cases")
    print("   Format: 1080x1920 vertical")
    print("   Duration: 9 minutes (540 seconds)")
    print("   Goal: Generate invalid FCPXML or prove validation system integrity")
    print()
    
    # Get all available assets
    assets_dir = Path(__file__).parent.parent / "assets"
    if not assets_dir.exists():
        print(f"❌ Assets directory not found: {assets_dir}")
        sys.exit(1)
    
    # Find all media files
    image_files = []
    video_files = []
    
    for pattern in ['*.png', '*.PNG', '*.jpg', '*.JPG', '*.jpeg', '*.JPEG']:
        image_files.extend(assets_dir.glob(pattern))
    
    for pattern in ['*.mov', '*.MOV', '*.mp4', '*.MP4']:
        video_files.extend(assets_dir.glob(pattern))
    
    print(f"   Found {len(image_files)} images: {[f.name for f in image_files]}")
    print(f"   Found {len(video_files)} videos: {[f.name for f in video_files]}")
    
    if not image_files and not video_files:
        print(f"❌ No media files found in {assets_dir}")
        sys.exit(1)
    
    # Create base project
    fcpxml = create_empty_project(
        project_name="FCPXML Stress Test - Maximum Complexity",
        event_name="Stress Test Validation",
        use_horizontal=False  # Always vertical 1080x1920 as requested
    )
    
    try:
        create_stress_test_timeline(fcpxml, image_files, video_files)
        print("✅ Stress test timeline created successfully")
    except Exception as e:
        print(f"❌ Error creating stress test timeline: {e}")
        print("   This indicates a potential library issue or validation gap")
        raise
    
    # Save with validation
    output_path = Path(args.output) if args.output else Path(__file__).parent / "stress_test.fcpxml"
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Stress test saved to: {output_path}")
        print("🎯 VALIDATION SUCCESS: Library generated valid FCPXML under extreme conditions")
        print("   Next steps:")
        print("   1. Validate XML with: xmllint --noout stress_test.fcpxml")
        print("   2. Import into Final Cut Pro to test import stability")
        print("   3. Play timeline to test performance with complex content")
    else:
        print("❌ VALIDATION FAILED: Library rejected its own output")
        print("   This indicates validation system is working correctly")
        sys.exit(1)


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
    
    from fcpxml_lib.models.elements import Format
    from fcpxml_lib.constants import IMAGE_FORMAT_NAME, DEFAULT_IMAGE_WIDTH, DEFAULT_IMAGE_HEIGHT, IMAGE_COLOR_SPACE
    
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
            from fcpxml_lib.core.fcpxml import needs_vertical_scaling
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
                from fcpxml_lib.constants import VERTICAL_SCALE_FACTOR
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
                    from fcpxml_lib.models.elements import Asset, MediaRep
                    from fcpxml_lib.utils.ids import generate_uid
                    from fcpxml_lib.constants import IMAGE_DURATION
                    
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
            from fcpxml_lib.models.elements import Asset, MediaRep
            from fcpxml_lib.utils.ids import generate_uid
            from fcpxml_lib.constants import IMAGE_DURATION
            
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


def main():
    """CLI entry point with command options"""
    parser = argparse.ArgumentParser(
        description="FCPXML Python Generator - Create Final Cut Pro projects",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s create-empty-project --output my_project.fcpxml
  %(prog)s create-random-video /path/to/media/folder --output random.fcpxml
  %(prog)s video-at-edge /path/to/image/folder --output edge_video.fcpxml --background-video bg.mp4
  %(prog)s stress-test --output stress_test.fcpxml
        """
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Create empty project command
    empty_parser = subparsers.add_parser(
        'create-empty-project',
        help='Create an empty FCPXML project'
    )
    empty_parser.add_argument('--project-name', help='Name of the project')
    empty_parser.add_argument('--event-name', help='Name of the event')
    empty_parser.add_argument('--output', help='Output FCPXML file path')
    empty_parser.add_argument('--horizontal', action='store_true', help='Use 1280x720 horizontal format instead of default 1080x1920 vertical')
    
    # Create random video command
    random_parser = subparsers.add_parser(
        'create-random-video',
        help='Create a random video from media files in a directory'
    )
    random_parser.add_argument('input_dir', help='Directory containing media files')
    random_parser.add_argument('--project-name', help='Name of the project')
    random_parser.add_argument('--event-name', help='Name of the event')
    random_parser.add_argument('--output', help='Output FCPXML file path')
    random_parser.add_argument('--clip-duration', type=float, default=5.0, help='Duration in seconds for each clip (default: 5.0)')
    random_parser.add_argument('--horizontal', action='store_true', help='Use 1280x720 horizontal format instead of default 1080x1920 vertical')
    
    # Video at edge command
    edge_parser = subparsers.add_parser(
        'video-at-edge',
        help='Create video with random images (PNG/JPG) tiled across visible area on multiple lanes'
    )
    edge_parser.add_argument('input_dir', help='Directory containing image files (PNG, JPG, JPEG)')
    edge_parser.add_argument('--output', help='Output FCPXML file path')
    edge_parser.add_argument('--background-video', help='Background video file (optional)')
    edge_parser.add_argument('--duration', type=float, default=10.0, help='Duration in seconds (default: 10.0)')
    edge_parser.add_argument('--tiles-per-lane', type=int, default=8, help='Number of image tiles per lane (default: 8)')
    edge_parser.add_argument('--num-lanes', type=int, default=10, help='Number of lanes with image tiles (default: 10)')
    
    # Stress test command
    stress_parser = subparsers.add_parser(
        'stress-test',
        help='Create an extremely complex 9-minute stress test video to validate library robustness'
    )
    stress_parser.add_argument('--output', help='Output FCPXML file path (default: stress_test.fcpxml)')
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    if args.command == 'create-empty-project':
        create_empty_project_cmd(args)
    elif args.command == 'create-random-video':
        create_random_video_cmd(args)
    elif args.command == 'video-at-edge':
        video_at_edge_cmd(args)
    elif args.command == 'stress-test':
        stress_test_cmd(args)


if __name__ == "__main__":
    main()