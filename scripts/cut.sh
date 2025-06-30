#!/bin/bash

INPUT="house100.png"
OUTDIR="tiles"
mkdir -p "$OUTDIR"

# Image dimensions
IMG_WIDTH=1080
IMG_HEIGHT=1920

# Choose number of columns (fit square tiles evenly across width)
NUM_COLS=4
TILE_SIZE=$((IMG_WIDTH / NUM_COLS))  # size of square tile

# Compute number of rows based on square height (only complete tiles)
NUM_ROWS=$((IMG_HEIGHT / TILE_SIZE))

# Crop tiles (only create tiles that fit completely within image bounds)
for ((row = 0; row < NUM_ROWS; row++)); do
  for ((col = 0; col < NUM_COLS; col++)); do
    X=$((col * TILE_SIZE))
    Y=$((row * TILE_SIZE))
    
    # Skip if tile would extend beyond image boundaries
    if ((Y + TILE_SIZE > IMG_HEIGHT)); then
      continue
    fi
    
    OUTFILE="${OUTDIR}/col${col}_row${row}.png"
    magick "$INPUT" -crop "${TILE_SIZE}x${TILE_SIZE}+${X}+${Y}" +repage "$OUTFILE"
  done
done

