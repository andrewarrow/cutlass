"""
Squares FX Command Implementation

Creates a 7x4 grid layout of house tile PNGs with proper scaling and spacing.
"""

import sys
import os
from pathlib import Path
from fcpxml_lib.core.fcpxml import create_empty_project, save_fcpxml
from fcpxml_lib.models.elements import Video, AdjustTransform
from fcpxml_lib.utils.ids import generate_resource_id, set_resource_id_counter
from fcpxml_lib.utils.timing import convert_seconds_to_fcp_duration
from fcpxml_lib.core.fcpxml import create_media_asset

def squares_fx_cmd(args):
    """CLI implementation for squares-fx command"""
    
    # Configuration
    tiles_dir = Path.home() / "Documents" / "house" / "tiles"
    output_file = args.output if hasattr(args, 'output') and args.output else "squares_fx.fcpxml"
    duration_seconds = 10.0
    
    # Validate tiles directory
    if not tiles_dir.exists():
        print(f"Error: Tiles directory not found: {tiles_dir}")
        return False
    
    # Get all PNG files in the expected pattern
    png_files = []
    for col in range(4):  # col0, col1, col2, col3
        for row in range(7):  # row0-row6
            filename = f"col{col}_row{row}.png"
            filepath = tiles_dir / filename
            if filepath.exists():
                png_files.append(filepath)
            else:
                print(f"Warning: Missing tile: {filename}")
    
    if not png_files:
        print(f"Error: No tile PNG files found in {tiles_dir}")
        return False
    
    print(f"Found {len(png_files)} tile files")
    
    # Grid layout calculations
    canvas_width = 1080
    canvas_height = 1920
    tile_size = 270
    cols = 4  # 4 columns (col0, col1, col2, col3)
    rows = 7  # 7 rows (row0, row1, row2, row3, row4, row5, row6)
    
    # Use slightly bigger scale to fill screen better
    scale = 0.25
    
    print(f"Using scale: {scale:.6f}")
    
    # Use coordinates based on s2.fcpxml pattern, expanded for full grid
    # s2.fcpxml shows: col0 at X≈-22, col1 at X≈-7 (15 unit spacing)
    # Y goes from +43 to -40 (83 unit range for 7 rows ≈ 12 units spacing)
    
    # Perfect spacing with uniform gaps between rows and columns
    # Screen bounds from test_video_at_edge.py: X=(-30,+30), Y=(-50,+50)
    x_positions = [-22.5, -7.5, 7.5, 22.5]  # 4 columns, 15-unit spacing (same as rows)
    y_positions = [45.0, 30.0, 15.0, 0.0, -15.0, -30.0, -45.0]  # 7 rows, 15-unit spacing
    
    # Create FCPXML project
    try:
        fcpxml = create_empty_project()
        
        # Reset resource ID counter to avoid conflicts (r1 is used by create_empty_project)
        set_resource_id_counter(1)
        
        # Get project components
        project = fcpxml.library.events[0].projects[0]
        sequence = project.sequences[0]
        
        # Convert duration
        duration = convert_seconds_to_fcp_duration(duration_seconds)
        
        # Special timing pattern from working squares.fcpxml
        special_timing = "86399313/24000s"
        
        # Create assets and video elements for all tiles
        background_element = None
        nested_elements = []
        
        for i, filepath in enumerate(png_files):
            # Parse filename to get grid position
            filename = filepath.stem  # e.g., "col0_row1"
            try:
                parts = filename.split('_')
                col_num = int(parts[0][3:])  # "col0" -> 0
                row_num = int(parts[1][3:])  # "row1" -> 1
            except (IndexError, ValueError):
                print(f"Warning: Unexpected filename format: {filename}")
                continue
            
            # Skip if outside our 4x7 grid
            if col_num >= cols or row_num >= rows:
                continue
            
            # Generate IDs
            asset_id = generate_resource_id()
            format_id = generate_resource_id()
            
            # Create media asset
            asset, format_obj = create_media_asset(
                str(filepath.absolute()),
                asset_id,
                format_id,
                include_audio=False
            )
            
            # Add to resources
            fcpxml.resources.assets.append(asset)
            fcpxml.resources.formats.append(format_obj)
            
            # Calculate position for this tile (fix the mapping!)
            # colX_rowY: X=column position, Y=row position
            x_pos = x_positions[col_num]  # col_num maps to X (column position) 
            y_pos = y_positions[row_num]  # row_num maps to Y (row position)
            
            if i == 0:
                # First tile becomes background element (like s2.fcpxml)
                background_element = {
                    "type": "video",
                    "ref": asset_id,
                    "duration": special_timing,
                    "start": special_timing,
                    "offset": "0s",
                    "name": filename,
                    "adjust_transform": {
                        "position": f"{x_pos:.1f} {y_pos:.1f}",
                        "scale": f"{scale:.6f} {scale:.6f}"
                    },
                    "nested_elements": []
                }
            else:
                # All other tiles are nested inside background (s2.fcpxml pattern)
                nested_element = {
                    "type": "video",
                    "ref": asset_id,
                    "duration": special_timing,
                    "start": special_timing,
                    "offset": special_timing,
                    "lane": str(i),  # Sequential lanes: 1, 2, 3...
                    "name": filename,
                    "adjust_transform": {
                        "position": f"{x_pos:.1f} {y_pos:.1f}",
                        "scale": f"{scale:.6f} {scale:.6f}"
                    }
                }
                nested_elements.append(nested_element)
        
        # Nest all tiles inside single background element (s2.fcpxml pattern)
        if background_element and nested_elements:
            background_element["nested_elements"] = nested_elements
            # Add only the background element to spine (contains all nested tiles)
            sequence.spine.ordered_elements = [background_element]
        else:
            print("Error: Failed to create background element structure")
            return False
        
        # Update sequence duration to match working file
        sequence.duration = special_timing
        
        # Save FCPXML
        success = save_fcpxml(fcpxml, output_file)
        
        if success:
            print(f"Generated: {output_file}")
            print(f"Grid: {len(png_files)} tiles in {cols}x{rows} layout")
            return True
        else:
            print("Error: Failed to save FCPXML")
            return False
            
    except Exception as e:
        print(f"Error generating squares FX: {e}")
        import traceback
        traceback.print_exc()
        return False