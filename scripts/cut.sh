#!/bin/bash

INPUT="house100.png"
RESIZED_INPUT="house100_resized.png"
OUTDIR="tiles"
mkdir -p "$OUTDIR"

# Resize image to 1080x1920 filling entire area (crop to fit)
magick "$INPUT" -resize 1080x1920^ -gravity center -extent 1080x1920 "$RESIZED_INPUT"

# Image dimensions after resize
IMG_WIDTH=1080
IMG_HEIGHT=1920

# Choose number of columns (fit square tiles evenly across width)
NUM_COLS=4
TILE_SIZE=$((IMG_WIDTH / NUM_COLS))  # size of square tile

# Compute number of rows that fit completely
NUM_ROWS=$((IMG_HEIGHT / TILE_SIZE))

# Crop tiles (create all possible rows, even partial ones)
for ((row = 0; row <= 6; row++)); do
  Y=$((row * TILE_SIZE))
  
  for ((col = 0; col < NUM_COLS; col++)); do
    X=$((col * TILE_SIZE))
    OUTFILE="${OUTDIR}/col${col}_row${row}.png"
    
    # Calculate remaining height for last row
    if ((row == 6)); then
      REMAINING_HEIGHT=$((IMG_HEIGHT - Y))
      magick "$RESIZED_INPUT" -crop "${TILE_SIZE}x${REMAINING_HEIGHT}+${X}+${Y}" +repage "$OUTFILE"
    else
      magick "$RESIZED_INPUT" -crop "${TILE_SIZE}x${TILE_SIZE}+${X}+${Y}" +repage "$OUTFILE"
    fi
  done
done

