#!/usr/bin/env python3
"""
iMessage Screenshot Editor
Removes text from iPhone message bubbles and replaces with test messages
"""

from PIL import Image, ImageDraw, ImageFont
import sys
import os

class iMessageEditor:
    def __init__(self, image_path):
        self.image = Image.open(image_path)
        self.width, self.height = self.image.size
        self.draw = ImageDraw.Draw(self.image)
        
        # iPhone message bubble colors (RGB)
        self.blue_bubble = (0, 122, 255)  # iOS blue (sent messages)
        self.gray_bubble_light = (229, 229, 234)  # iOS gray (light mode)
        self.gray_bubble_dark = (60, 60, 67)  # iOS gray (dark mode)
        self.white_text = (255, 255, 255)
        self.black_text = (0, 0, 0)
        
        # Detect if this is dark mode based on background
        self.is_dark_mode = self._detect_dark_mode()
        
    def find_message_bubbles(self):
        """
        Detect both blue and gray message bubbles in the image
        Returns list of tuples: (bubble_region, bubble_type)
        bubble_type: 'blue' for sent messages, 'gray' for received messages
        """
        bubbles = []
        pixels = self.image.load()
        
        # Scan for message bubble regions
        visited = set()
        
        for y in range(self.height):
            for x in range(self.width):
                if (x, y) in visited:
                    continue
                    
                pixel = pixels[x, y]
                bubble_type = None
                
                # Check if pixel matches bubble colors
                if self._is_blue_bubble_color(pixel):
                    bubble_type = 'blue'
                elif self._is_gray_bubble_color(pixel):
                    bubble_type = 'gray'
                
                if bubble_type:
                    bubble_region = self._trace_bubble_region(x, y, pixels, visited, bubble_type)
                    if bubble_region and self._is_valid_bubble_size(bubble_region):
                        bubbles.append((bubble_region, bubble_type))
        
        return bubbles
    
    def _detect_dark_mode(self):
        """Detect if screenshot is in dark mode by sampling background pixels"""
        # Sample pixels from top area (status bar background)
        sample_pixels = []
        for y in range(50, 100):  # Sample from status bar area
            for x in range(50, self.width - 50, 20):
                if x < self.width and y < self.height:
                    pixel = self.image.getpixel((x, y))
                    if len(pixel) >= 3:
                        brightness = sum(pixel[:3]) / 3
                        sample_pixels.append(brightness)
        
        if sample_pixels:
            avg_brightness = sum(sample_pixels) / len(sample_pixels)
            return avg_brightness < 100  # Dark if average brightness is low
        return False
    
    def _is_blue_bubble_color(self, pixel):
        """Check if pixel color matches iOS blue bubble (with tolerance)"""
        if len(pixel) < 3:
            return False
        r, g, b = pixel[:3]
        
        # iOS blue with tolerance
        target_r, target_g, target_b = self.blue_bubble
        tolerance = 40
        
        return (abs(r - target_r) <= tolerance and 
                abs(g - target_g) <= tolerance and 
                abs(b - target_b) <= tolerance)
    
    def _is_gray_bubble_color(self, pixel):
        """Check if pixel color matches iOS gray bubble (with tolerance)"""
        if len(pixel) < 3:
            return False
        r, g, b = pixel[:3]
        
        # Choose target gray based on mode
        if self.is_dark_mode:
            target_r, target_g, target_b = self.gray_bubble_dark
        else:
            target_r, target_g, target_b = self.gray_bubble_light
        
        tolerance = 40
        
        return (abs(r - target_r) <= tolerance and 
                abs(g - target_g) <= tolerance and 
                abs(b - target_b) <= tolerance)
    
    def _trace_bubble_region(self, start_x, start_y, pixels, visited, bubble_type):
        """Trace the boundaries of a message bubble"""
        min_x = max_x = start_x
        min_y = max_y = start_y
        
        stack = [(start_x, start_y)]
        bubble_pixels = set()
        
        while stack:
            x, y = stack.pop()
            
            if (x, y) in visited or x < 0 or x >= self.width or y < 0 or y >= self.height:
                continue
                
            pixel = pixels[x, y]
            
            # Check if pixel matches the bubble type we're tracing
            is_matching_pixel = False
            if bubble_type == 'blue':
                is_matching_pixel = self._is_blue_bubble_color(pixel)
            elif bubble_type == 'gray':
                is_matching_pixel = self._is_gray_bubble_color(pixel)
            
            if not is_matching_pixel:
                continue
                
            visited.add((x, y))
            bubble_pixels.add((x, y))
            
            min_x = min(min_x, x)
            max_x = max(max_x, x)
            min_y = min(min_y, y)
            max_y = max(max_y, y)
            
            # Add neighbors
            for dx, dy in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
                stack.append((x + dx, y + dy))
        
        if len(bubble_pixels) > 200:  # Minimum size for a bubble
            return (min_x, min_y, max_x - min_x, max_y - min_y)
        return None
    
    def _is_valid_bubble_size(self, region):
        """Check if region is a reasonable size for a message bubble"""
        x, y, width, height = region
        return width > 50 and height > 20 and width < self.width * 0.8
    
    def clear_bubble_text(self, bubble_region, bubble_type):
        """Clear text from inside a message bubble by filling with appropriate bubble color"""
        x, y, width, height = bubble_region
        
        # Create a mask for the rounded rectangle shape
        bubble_img = Image.new('RGBA', (width, height), (0, 0, 0, 0))
        bubble_draw = ImageDraw.Draw(bubble_img)
        
        # Choose color based on bubble type
        if bubble_type == 'blue':
            fill_color = self.blue_bubble + (255,)
        else:  # gray
            if self.is_dark_mode:
                fill_color = self.gray_bubble_dark + (255,)
            else:
                fill_color = self.gray_bubble_light + (255,)
        
        # Draw rounded rectangle with appropriate color
        corner_radius = min(width, height) // 4
        bubble_draw.rounded_rectangle(
            [0, 0, width, height], 
            radius=corner_radius, 
            fill=fill_color
        )
        
        # Paste the clean bubble back
        self.image.paste(bubble_img, (x, y), bubble_img)
    
    def add_text_to_bubble(self, bubble_region, text, bubble_type):
        """Add text to a message bubble with appropriate text color"""
        x, y, width, height = bubble_region
        
        # Try to load system font, fall back to PIL default
        try:
            font_size = max(12, min(height // 3, 18))
            font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Regular.otf", font_size)
        except:
            try:
                font = ImageFont.truetype("arial.ttf", font_size)
            except:
                font = ImageFont.load_default()
        
        # Calculate text position (centered)
        bbox = self.draw.textbbox((0, 0), text, font=font)
        text_width = bbox[2] - bbox[0]
        text_height = bbox[3] - bbox[1]
        
        text_x = x + (width - text_width) // 2
        text_y = y + (height - text_height) // 2
        
        # Choose text color based on bubble type and mode
        if bubble_type == 'blue':
            text_color = self.white_text  # White text on blue bubble
        else:  # gray bubble
            if self.is_dark_mode:
                text_color = self.white_text  # White text on dark gray bubble
            else:
                text_color = self.black_text  # Black text on light gray bubble
        
        self.draw.text((text_x, text_y), text, fill=text_color, font=font)
    
    def process_image(self, test_messages=None):
        """
        Main processing function:
        1. Find message bubbles (both blue and gray)
        2. Clear existing text
        3. Add new test messages
        """
        if test_messages is None:
            test_messages = ["Test1", "Test2", "Test3", "Test4", "Test5", "Test6", "Test7", "Test8"]
        
        print(f"Dark mode detected: {self.is_dark_mode}")
        
        # Find all message bubbles
        bubbles = self.find_message_bubbles()
        
        if not bubbles:
            print("No message bubbles found. Trying alternative method...")
            # If no bubbles found, look for regions manually
            bubbles = self._find_regions_manual()
        
        print(f"Found {len(bubbles)} message bubble(s)")
        
        # Sort bubbles by vertical position (top to bottom)
        bubbles.sort(key=lambda x: x[0][1])  # Sort by y coordinate
        
        # Process each bubble
        for i, (bubble_region, bubble_type) in enumerate(bubbles):
            print(f"Processing bubble {i+1}: {bubble_type} at {bubble_region}")
            
            # Clear existing text
            self.clear_bubble_text(bubble_region, bubble_type)
            
            # Add new test message
            if i < len(test_messages):
                self.add_text_to_bubble(bubble_region, test_messages[i], bubble_type)
            else:
                # If more bubbles than messages, cycle through messages
                self.add_text_to_bubble(bubble_region, test_messages[i % len(test_messages)], bubble_type)
    
    def _find_regions_manual(self):
        """Manual detection for message bubble regions when automatic detection fails"""
        bubbles = []
        
        # Search the entire image systematically
        step_size = 20
        
        for search_y in range(100, self.height - 100, step_size):
            for search_x in range(50, self.width - 50, step_size):
                pixel = self.image.getpixel((search_x, search_y))
                
                # Check for blue bubbles
                if self._is_blue_bubble_color(pixel):
                    bubble_width = min(300, self.width - search_x - 50)
                    bubble_height = 60
                    bubble_x = search_x - 20
                    bubble_y = search_y - 20
                    
                    # Ensure bubble is within bounds
                    bubble_x = max(0, min(bubble_x, self.width - bubble_width))
                    bubble_y = max(0, min(bubble_y, self.height - bubble_height))
                    
                    region = (bubble_x, bubble_y, bubble_width, bubble_height)
                    bubbles.append((region, 'blue'))
                
                # Check for gray bubbles
                elif self._is_gray_bubble_color(pixel):
                    bubble_width = min(300, self.width - search_x - 50)
                    bubble_height = 60
                    bubble_x = search_x - 20
                    bubble_y = search_y - 20
                    
                    # Ensure bubble is within bounds
                    bubble_x = max(0, min(bubble_x, self.width - bubble_width))
                    bubble_y = max(0, min(bubble_y, self.height - bubble_height))
                    
                    region = (bubble_x, bubble_y, bubble_width, bubble_height)
                    bubbles.append((region, 'gray'))
        
        # Remove duplicate/overlapping bubbles
        filtered_bubbles = []
        for bubble in bubbles:
            region, bubble_type = bubble
            x, y, w, h = region
            
            # Check if this bubble overlaps significantly with existing ones
            is_duplicate = False
            for existing_region, _ in filtered_bubbles:
                ex, ey, ew, eh = existing_region
                
                # Check for significant overlap
                if (abs(x - ex) < 100 and abs(y - ey) < 40):
                    is_duplicate = True
                    break
            
            if not is_duplicate:
                filtered_bubbles.append(bubble)
        
        return filtered_bubbles
    
    def save(self, output_path):
        """Save the processed image"""
        self.image.save(output_path)
        print(f"Saved processed image to: {output_path}")

def main():
    if len(sys.argv) < 2:
        print("Usage: python imessage.py <image_path> [output_path]")
        print("Example: python imessage.py screenshot.png output.png")
        sys.exit(1)
    
    input_path = sys.argv[1]
    output_path = sys.argv[2] if len(sys.argv) > 2 else "imessage_output.png"
    
    if not os.path.exists(input_path):
        print(f"Error: Image file '{input_path}' not found")
        sys.exit(1)
    
    try:
        # Create editor instance
        editor = iMessageEditor(input_path)
        
        # Define test messages
        test_messages = ["Test1", "Test2", "Test3", "Test4", "Test5"]
        
        # Process the image
        editor.process_image(test_messages)
        
        # Save result
        editor.save(output_path)
        
        print("Processing complete!")
        
    except Exception as e:
        print(f"Error processing image: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()