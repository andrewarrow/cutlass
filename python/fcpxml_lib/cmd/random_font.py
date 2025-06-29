#!/usr/bin/env python3
"""
Random Font Command

Creates a 9-minute 1080x1920 video with title elements using random fonts
from ../reference/fonts.txt with random positive messages from ../assets
"""

import sys
import random
import colorsys
from pathlib import Path

from fcpxml_lib import create_empty_project, save_fcpxml
from fcpxml_lib.models.elements import Title
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration
from fcpxml_lib.utils.ids import generate_resource_id, set_resource_id_counter


def get_contrasting_colors():
    """Generate random face and outline colors with good contrast"""
    # Generate a random hue
    hue = random.random()
    
    # For face color: moderate saturation and lightness
    face_saturation = random.uniform(0.6, 1.0)
    face_lightness = random.uniform(0.4, 0.8)
    
    # For outline: complementary approach
    # If face is light, make outline dark; if face is dark, make outline light
    if face_lightness > 0.6:
        outline_lightness = random.uniform(0.1, 0.3)
    else:
        outline_lightness = random.uniform(0.7, 0.9)
    
    outline_saturation = random.uniform(0.3, 0.8)
    
    # Convert HSL to RGB
    face_rgb = colorsys.hls_to_rgb(hue, face_lightness, face_saturation)
    outline_rgb = colorsys.hls_to_rgb(hue, outline_lightness, outline_saturation)
    
    # Convert to Final Cut Pro color format (0-1 range)
    face_color = f"{face_rgb[0]:.3f} {face_rgb[1]:.3f} {face_rgb[2]:.3f} 1"
    outline_color = f"{outline_rgb[0]:.3f} {outline_rgb[1]:.3f} {outline_rgb[2]:.3f} 1"
    
    return face_color, outline_color


def get_positive_messages():
    """Generate list of positive vibe messages"""
    messages = [
        "You are amazing!",
        "Believe in yourself",
        "Today is your day",
        "You've got this!",
        "Stay positive",
        "Dream big",
        "Be unstoppable",
        "You are enough",
        "Shine bright",
        "Make it happen",
        "Be fearless",
        "Stay strong",
        "You matter",
        "Be bold",
        "Create magic",
        "Live fully",
        "Be grateful",
        "Spread joy",
        "Stay curious",
        "Be kind",
        "Embrace change",
        "Find your spark",
        "Be present",
        "Choose happiness",
        "Stay hopeful",
        "Be authentic",
        "Love yourself",
        "Take the leap",
        "Be brilliant",
        "Trust the process",
        "You inspire others",
        "Be the change",
        "Stay focused",
        "Believe in magic",
        "Be limitless"
    ]
    return messages


def load_fonts_from_file(fonts_file_path):
    """Load fonts from the reference file"""
    try:
        with open(fonts_file_path, 'r') as f:
            fonts = []
            for line in f:
                line = line.strip()
                if line:  # Each line is a font name
                    fonts.append(line)
            return fonts
    except FileNotFoundError:
        print(f"❌ Fonts file not found: {fonts_file_path}")
        return []
    except Exception as e:
        print(f"❌ Error reading fonts file: {e}")
        return []


def create_title_effect(effect_id, font_name, face_color, outline_color):
    """Create a title effect with specified font and colors"""
    return {
        "id": effect_id,
        "name": "Custom Title",
        "uid": "9F3AC8E5-EF80-4F68-B2AE-6D5B0C1F5B8D",
        "src": "file:///System/Library/PrivateFrameworks/ProVideo.framework/Versions/A/Resources/Plugins/FxPlug/Text/Custom.fxplug",
        "filter-video": {
            "param": [
                {
                    "name": "FontName",
                    "value": font_name
                },
                {
                    "name": "FontSize",
                    "value": "290"
                },
                {
                    "name": "FaceColor",
                    "value": face_color
                },
                {
                    "name": "OutlineColor", 
                    "value": outline_color
                },
                {
                    "name": "OutlineWidth",
                    "value": "15"
                }
            ]
        }
    }


def create_title_element(effect_id, text_content, font_name, face_color, outline_color, offset_fcp, duration_fcp):
    """Create a title element with text content and styling"""
    return {
        "type": "title",
        "ref": effect_id,
        "offset": offset_fcp,
        "duration": duration_fcp,
        "start": "0s",
        "name": f"{text_content} - Text",
        "lane": "1",
        "text_content": text_content,
        "font_name": font_name,
        "face_color": face_color,
        "outline_color": outline_color
    }


def random_font_cmd(args):
    """Create a 9-minute video with random font titles"""
    
    # Load fonts from reference file
    fonts_file = Path(__file__).parent.parent.parent.parent / "reference" / "fonts.txt"
    fonts = load_fonts_from_file(fonts_file)
    
    if not fonts:
        print("❌ No fonts found in ../reference/fonts.txt")
        sys.exit(1)
    
    print(f"✅ Loaded {len(fonts)} fonts from reference file")
    
    # Get positive messages
    messages = get_positive_messages()
    
    # Calculate number of titles needed for 9 minutes (540 seconds)
    # Each title will be 5 seconds long, so we need 108 titles
    title_duration = 5.0
    total_duration = 9 * 60  # 9 minutes in seconds
    num_titles = int(total_duration / title_duration)
    
    print(f"🎬 Creating 9-minute video with {num_titles} title elements...")
    print(f"   Each title duration: {title_duration}s")
    print(f"   Format: 1080x1920 vertical")
    
    # Create empty project (vertical format)
    fcpxml = create_empty_project(
        project_name=args.project_name or "Random Font Video",
        event_name=args.event_name or "Font Videos",
        use_horizontal=False  # 1080x1920 vertical
    )
    
    # Reset resource ID counter to start after existing resources
    existing_count = len(fcpxml.resources.formats) + len(fcpxml.resources.assets) + len(fcpxml.resources.effects)
    set_resource_id_counter(existing_count)
    
    # Get the main sequence
    sequence = fcpxml.library.events[0].projects[0].sequences[0]
    
    # Pick a random background asset from ../assets/
    assets_dir = Path(__file__).parent.parent.parent.parent / "assets"
    from fcpxml_lib.utils.media import discover_all_media_files
    
    # Find all media files in assets directory
    image_files, video_files = discover_all_media_files(assets_dir)
    all_media_files = image_files + video_files
    
    if not all_media_files:
        print("❌ No media files found in ../assets/")
        sys.exit(1)
    
    # Pick a random background file
    background_file = random.choice(all_media_files)
    print(f"   Using background: {background_file.name}")
    
    # Create background asset and add to resources
    from fcpxml_lib.core.fcpxml import create_media_asset
    background_asset_id = generate_resource_id()
    background_format_id = generate_resource_id()
    
    background_asset, background_format = create_media_asset(
        str(background_file), background_asset_id, background_format_id
    )
    
    fcpxml.resources.assets.append(background_asset)
    fcpxml.resources.formats.append(background_format)
    
    # Create a single title effect in resources that all titles will reference
    title_effect_id = generate_resource_id()
    title_effect = {
        "id": title_effect_id,
        "name": "Text",
        "uid": ".../Titles.localized/Basic Text.localized/Text.localized/Text.moti"
    }
    fcpxml.resources.title_effects.append(title_effect)
    
    # Create the main video element that will contain all titles
    video_duration_fcp = convert_seconds_to_fcp_duration(total_duration)
    
    video_element = {
        "type": "video",
        "ref": background_asset_id,  # Reference the background asset
        "offset": "0s",
        "duration": video_duration_fcp,
        "start": "0s",
        "name": f"Background: {background_file.name}",
        "nested_elements": []  # Will contain all title elements
    }
    
    # Generate titles as nested elements
    for i in range(num_titles):
        # Pick random font and message
        font_name = random.choice(fonts)
        message = random.choice(messages)
        
        # Generate contrasting colors
        face_color, outline_color = get_contrasting_colors()
        
        # Calculate timing
        offset_seconds = i * title_duration
        offset_fcp = convert_seconds_to_fcp_duration(offset_seconds)
        duration_fcp = convert_seconds_to_fcp_duration(title_duration)
        
        # Create title element
        title_element = create_title_element(
            title_effect_id, message, font_name, 
            face_color, outline_color, offset_fcp, duration_fcp
        )
        
        # Add to the video element as nested element
        video_element["nested_elements"].append(title_element)
        
        if i % 20 == 0:
            print(f"   Created {i+1}/{num_titles} titles...")
    
    # Add the video element to the spine
    sequence.spine.videos.append(video_element)
    sequence.spine.ordered_elements.append(video_element)
    
    # Update sequence duration
    total_duration_fcp = convert_seconds_to_fcp_duration(total_duration)
    sequence.duration = total_duration_fcp
    
    print(f"✅ Created {num_titles} titles with random fonts and colors")
    print(f"   Total duration: {total_duration/60:.1f} minutes")
    
    # Save to file with validation
    output_path = Path(args.output) if args.output else Path("random_font_video.fcpxml")
    validation_passed = save_fcpxml(fcpxml, str(output_path))
    
    if validation_passed:
        print(f"✅ Saved to: {output_path}")
        print(f"   Import this file into Final Cut Pro to see your random font video!")
    else:
        print("❌ Cannot proceed - fix validation errors first")
        sys.exit(1)