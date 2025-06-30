#!/bin/bash

TILEDIR="tiles"
OUTPUT="assembled.png"

# Target canvas size
CANVAS_WIDTH=1080
CANVAS_HEIGHT=1920

# Tile info  
TILE_SIZE=270
NUM_COLS=4
NUM_ROWS=7
GAP=6

# Create blank canvas
magick -size ${CANVAS_WIDTH}x${CANVAS_HEIGHT} xc:white "$OUTPUT"

# Place each tile with gaps
for ((row = 0; row < NUM_ROWS; row++)); do
  for ((col = 0; col < NUM_COLS; col++)); do
    TILEFILE="${TILEDIR}/col${col}_row${row}.png"
    
    # Skip if tile doesn't exist
    if [[ ! -f "$TILEFILE" ]]; then
      continue
    fi
    
    # Calculate position with gaps
    X=$((col * (TILE_SIZE + GAP)))
    Y=$((row * (TILE_SIZE + GAP)))
    
    # Skip if position would exceed canvas bounds (but allow last row)
    if ((Y >= CANVAS_HEIGHT)); then
      continue
    fi
    
    # Composite tile onto canvas
    magick "$OUTPUT" "$TILEFILE" -geometry +${X}+${Y} -composite "$OUTPUT"
  done
done

echo "Assembled image saved as $OUTPUT (${CANVAS_WIDTH}x${CANVAS_HEIGHT})"