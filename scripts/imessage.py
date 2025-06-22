#!/usr/bin/env python3
"""
Realistic iMessage Screenshot Generator
Creates authentic-looking iPhone Messages conversations from scratch
"""

from PIL import Image, ImageDraw, ImageFont
import sys
import os
import random
from datetime import datetime, timedelta
import colorsys

class iMessageGenerator:
    def __init__(self, background_path=None):
        # Use high resolution like original screenshot
        self.width, self.height = 1179, 2556  # High res iPhone dimensions
        self.phone_bg = Image.new('RGB', (self.width, self.height), (0, 0, 0))
        
        # iOS Colors (exact values)
        self.ios_blue = (0, 122, 255)
        self.ios_gray_dark = (60, 60, 67)
        self.ios_gray_light = (229, 229, 234)
        self.ios_background = (0, 0, 0)  # Dark mode
        self.ios_text_primary = (255, 255, 255)
        self.ios_text_secondary = (174, 174, 178)
        
        # Layout constants (scaled for high resolution)
        self.status_bar_height = 150
        self.header_height = 360
        self.message_area_start = self.status_bar_height + self.header_height
        self.message_padding = 60
        self.avatar_size = 96  # 3x original size
        self.bubble_max_width = 750  # 3x original size
        self.bubble_min_width = 180
    
    def create_realistic_avatar(self, person_name):
        """Generate a realistic-looking avatar"""
        size = self.avatar_size
        avatar = Image.new('RGBA', (size, size), (0, 0, 0, 0))
        draw = ImageDraw.Draw(avatar)
        
        # Generate a consistent color based on name
        hash_val = hash(person_name) % 360
        hue = hash_val / 360.0
        rgb = colorsys.hsv_to_rgb(hue, 0.7, 0.9)
        color = tuple(int(c * 255) for c in rgb)
        
        # Draw circle background
        draw.ellipse([0, 0, size, size], fill=color)
        
        # Add initials with proper high-res font
        try:
            font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Bold.otf", 42)
        except:
            try:
                font = ImageFont.truetype("Arial.ttf", 42)
            except:
                font = ImageFont.load_default()
        
        initials = "".join([word[0].upper() for word in person_name.split()[:2]])
        bbox = draw.textbbox((0, 0), initials, font=font)
        text_width = bbox[2] - bbox[0]
        text_height = bbox[3] - bbox[1]
        
        text_x = (size - text_width) // 2
        text_y = (size - text_height) // 2
        
        draw.text((text_x, text_y), initials, fill=(255, 255, 255), font=font)
        
        return avatar
    
    def draw_status_bar(self, draw):
        """Draw iOS status bar like real screenshot"""
        try:
            font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Semibold.otf", 51)
        except:
            try:
                font = ImageFont.truetype("Arial.ttf", 51)
            except:
                font = ImageFont.load_default()
        
        # Time (left side) - match real screenshot
        time_str = "9:34"
        draw.text((60, 45), time_str, fill=self.ios_text_primary, font=font)
        
        # Right side indicators (scaled for high res)
        right_x = self.width - 270
        
        # Signal bars (using actual bars instead of dots)
        bar_width = 9
        bar_spacing = 12
        bar_heights = [12, 18, 24, 30]  # Different heights for signal strength
        
        for i, height in enumerate(bar_heights):
            x = right_x + i * (bar_width + bar_spacing)
            y = 75 - height
            draw.rectangle([x, y, x + bar_width, 75], fill=self.ios_text_primary)
        
        # WiFi symbol (text)
        wifi_x = right_x + 60
        try:
            wifi_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Medium.otf", 45)
        except:
            wifi_font = font
        draw.text((wifi_x, 36), "📶", fill=self.ios_text_primary, font=wifi_font)
        
        # Battery (more realistic, scaled)
        battery_x = right_x + 120
        battery_width = 72
        battery_height = 36
        battery_y = 48
        
        # Battery outline
        draw.rectangle([battery_x, battery_y, battery_x + battery_width, battery_y + battery_height], 
                      outline=self.ios_text_primary, width=3)
        
        # Battery fill (72%)
        fill_width = int(battery_width * 0.72) - 6
        draw.rectangle([battery_x + 3, battery_y + 3, battery_x + 3 + fill_width, battery_y + battery_height - 3], 
                      fill=self.ios_text_primary)
        
        # Battery tip
        draw.rectangle([battery_x + battery_width, battery_y + 12, battery_x + battery_width + 6, battery_y + 24], 
                      fill=self.ios_text_primary)
        
        # 72 text
        try:
            small_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Medium.otf", 42)
        except:
            small_font = font
        draw.text((battery_x + battery_width + 15, 45), "72", fill=self.ios_text_primary, font=small_font)
    
    def draw_messages_header(self, draw, contact_name="Family group chat"):
        """Draw Messages app header like real screenshot"""
        header_y = self.status_bar_height
        
        try:
            title_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Semibold.otf", 51)
            detail_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Regular.otf", 45)
        except:
            try:
                title_font = ImageFont.truetype("Arial.ttf", 51)
                detail_font = ImageFont.truetype("Arial.ttf", 45)
            except:
                title_font = ImageFont.load_default()
                detail_font = ImageFont.load_default()
        
        # Back button (blue chevron)
        draw.text((45, header_y + 45), "<", fill=self.ios_blue, font=title_font)
        
        # Group avatars (like in real screenshot, scaled)
        avatar_center_x = self.width // 2
        avatar_y = header_y + 30
        
        # Draw group avatar cluster
        avatar_colors = [(255, 182, 0), (255, 59, 48), (52, 199, 89), (0, 122, 255)]
        avatar_positions = [
            (-60, -30), (30, -30), (-60, 45), (30, 45)  # 2x2 grid, scaled
        ]
        
        for i, ((dx, dy), color) in enumerate(zip(avatar_positions, avatar_colors)):
            if i < 4:  # Only draw 4 avatars
                x = avatar_center_x + dx
                y = avatar_y + dy
                draw.ellipse([x, y, x + 60, y + 60], fill=color)
                
                # Add emoji/initials with proper font
                try:
                    emoji_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Medium.otf", 30)
                except:
                    emoji_font = ImageFont.load_default()
                
                if i == 0:
                    draw.text((x + 18, y + 12), "🦅", fill=(255, 255, 255), font=emoji_font)
                elif i == 1:
                    draw.text((x + 12, y + 12), "👨‍💼", fill=(255, 255, 255), font=emoji_font)
        
        # Contact name/group name (centered below avatars)
        bbox = draw.textbbox((0, 0), contact_name, font=title_font)
        text_width = bbox[2] - bbox[0]
        title_x = (self.width - text_width) // 2
        draw.text((title_x, header_y + 150), contact_name, fill=self.ios_text_primary, font=title_font)
        
        # Video call icon (blue camera outline, scaled)
        video_icon_x = self.width - 105
        video_icon_y = header_y + 45
        
        # Draw camera outline
        draw.rounded_rectangle([video_icon_x, video_icon_y, video_icon_x + 60, video_icon_y + 45], 
                              radius=9, outline=self.ios_blue, width=6)
        draw.polygon([(video_icon_x + 60, video_icon_y + 9), 
                     (video_icon_x + 75, video_icon_y + 3), 
                     (video_icon_x + 75, video_icon_y + 33),
                     (video_icon_x + 60, video_icon_y + 27)], fill=self.ios_blue)
    
    def calculate_bubble_size(self, text, font):
        """Calculate appropriate bubble size for text"""
        # Create temporary draw to measure text
        temp_img = Image.new('RGB', (1, 1))
        temp_draw = ImageDraw.Draw(temp_img)
        
        # Handle multi-line text (scaled for high res)
        lines = text.split('\n')
        max_line_width = 0
        total_height = 0
        line_height = 60  # 3x original
        
        for line in lines:
            bbox = temp_draw.textbbox((0, 0), line, font=font)
            line_width = bbox[2] - bbox[0]
            max_line_width = max(max_line_width, line_width)
            total_height += line_height
        
        # Add padding (scaled)
        padding_x = 48  # 3x original
        padding_y = 36  # 3x original
        
        bubble_width = min(self.bubble_max_width, max(self.bubble_min_width, max_line_width + padding_x * 2))
        bubble_height = max(108, total_height + padding_y * 2)  # 3x original
        
        return bubble_width, bubble_height
    
    def draw_message_bubble(self, draw, x, y, text, is_sent=False, show_tail=True):
        """Draw a realistic iOS message bubble with high quality fonts"""
        try:
            font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Regular.otf", 48)
        except:
            try:
                font = ImageFont.truetype("Arial.ttf", 48)
            except:
                font = ImageFont.load_default()
        
        bubble_width, bubble_height = self.calculate_bubble_size(text, font)
        
        # Choose colors
        if is_sent:
            bubble_color = self.ios_blue
            text_color = (255, 255, 255)
        else:
            bubble_color = self.ios_gray_dark
            text_color = (255, 255, 255)
        
        # Adjust position for sent messages (right-aligned)
        if is_sent:
            x = self.width - bubble_width - self.message_padding
        
        # Draw main bubble body (scaled)
        corner_radius = 54  # 3x original
        draw.rounded_rectangle(
            [x, y, x + bubble_width, y + bubble_height],
            radius=corner_radius,
            fill=bubble_color
        )
        
        # Draw bubble tail (the little point, scaled)
        if show_tail:
            tail_size = 18  # 3x original
            if is_sent:
                # Right tail for sent messages
                tail_points = [
                    (x + bubble_width, y + bubble_height - 45),
                    (x + bubble_width + tail_size, y + bubble_height - 24),
                    (x + bubble_width, y + bubble_height - 24)
                ]
            else:
                # Left tail for received messages
                tail_points = [
                    (x, y + bubble_height - 45),
                    (x - tail_size, y + bubble_height - 24),
                    (x, y + bubble_height - 24)
                ]
            draw.polygon(tail_points, fill=bubble_color)
        
        # Add text (handle multi-line, scaled)
        lines = text.split('\n')
        line_height = 60  # 3x original
        text_start_y = y + (bubble_height - len(lines) * line_height) // 2
        
        for i, line in enumerate(lines):
            bbox = draw.textbbox((0, 0), line, font=font)
            line_width = bbox[2] - bbox[0]
            text_x = x + (bubble_width - line_width) // 2
            text_y = text_start_y + i * line_height
            
            draw.text((text_x, text_y), line, fill=text_color, font=font)
        
        return bubble_height
    
    def draw_timestamp(self, draw, x, y, timestamp_text):
        """Draw message timestamp with high quality fonts"""
        try:
            font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Regular.otf", 36)
        except:
            try:
                font = ImageFont.truetype("Arial.ttf", 36)
            except:
                font = ImageFont.load_default()
        
        bbox = draw.textbbox((0, 0), timestamp_text, font=font)
        text_width = bbox[2] - bbox[0]
        timestamp_x = (self.width - text_width) // 2
        
        draw.text((timestamp_x, y), timestamp_text, fill=self.ios_text_secondary, font=font)
        
        return 45  # Height of timestamp (scaled)
    
    def generate_conversation(self, messages, contact_name="Family Group"):
        """Generate a complete conversation"""
        # Start with phone background
        result = self.phone_bg.copy()
        draw = ImageDraw.Draw(result)
        
        # Fill entire screen with black background
        draw.rectangle([0, 0, self.width, self.height], fill=self.ios_background)
        
        # Draw status bar and header
        self.draw_status_bar(draw)
        self.draw_messages_header(draw, contact_name)
        
        # Draw messages
        current_y = self.message_area_start + 20
        last_sender = None
        last_time = None
        
        # Create avatars for each unique sender
        senders = set(msg.get('sender') for msg in messages if not msg.get('is_sent', False))
        avatars = {}
        for sender in senders:
            if sender:
                avatars[sender] = self.create_realistic_avatar(sender)
        
        for i, message in enumerate(messages):
            text = message['text']
            is_sent = message.get('is_sent', False)
            sender = message.get('sender', 'You')
            timestamp = message.get('timestamp')
            
            # Add timestamp if significant time gap or first message
            if timestamp and (last_time is None or 
                             (timestamp - last_time).total_seconds() > 3600):  # 1 hour gap
                time_str = timestamp.strftime("%a, %b %d at %I:%M %p").replace(" 0", " ")
                timestamp_height = self.draw_timestamp(draw, 0, current_y, time_str)
                current_y += timestamp_height + 20
                last_time = timestamp
            
            # Add sender name for received messages (if sender changed)
            if not is_sent and sender != last_sender:
                try:
                    name_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Regular.otf", 39)
                except:
                    try:
                        name_font = ImageFont.truetype("Arial.ttf", 39)
                    except:
                        name_font = ImageFont.load_default()
                
                # Draw sender name (scaled position)
                draw.text((210, current_y), sender.lower(), fill=self.ios_text_secondary, font=name_font)
                current_y += 60
            
            # Add avatar for received messages (if sender changed)
            avatar_height = 0
            if not is_sent and sender != last_sender and sender in avatars:
                avatar_x = self.message_padding
                avatar_y = current_y
                result.paste(avatars[sender], (avatar_x, avatar_y), avatars[sender])
            
            # Determine bubble position (scaled)
            if is_sent:
                bubble_x = 0  # Will be adjusted in draw_message_bubble
            else:
                bubble_x = 210  # Leave space for avatar (scaled)
            
            # Draw message bubble
            show_tail = (sender != last_sender)
            bubble_height = self.draw_message_bubble(
                draw, bubble_x, current_y, text, is_sent, show_tail
            )
            
            # Add extra spacing between different senders (scaled)
            spacing = 45 if sender != last_sender else 24
            current_y += bubble_height + spacing
            last_sender = sender
        
        # Draw message input area at bottom (scaled)
        input_y = self.height - 240
        
        # Plus button (left, scaled)
        plus_size = 90
        plus_x = 45
        plus_y = input_y + 15
        draw.ellipse([plus_x, plus_y, plus_x + plus_size, plus_y + plus_size], 
                    fill=(44, 44, 46))
        
        try:
            plus_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Medium.otf", 45)
        except:
            plus_font = ImageFont.load_default()
        draw.text((plus_x + 24, plus_y + 15), "+", fill=self.ios_text_primary, font=plus_font)
        
        # Message input field (scaled)
        input_field_x = plus_x + plus_size + 30
        input_field_width = self.width - input_field_x - 180
        draw.rounded_rectangle(
            [input_field_x, input_y + 15, input_field_x + input_field_width, input_y + 105],
            radius=54,
            fill=(44, 44, 46)
        )
        
        # iMessage placeholder text
        try:
            font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Regular.otf", 48)
        except:
            try:
                font = ImageFont.truetype("Arial.ttf", 48)
            except:
                font = ImageFont.load_default()
        
        draw.text((input_field_x + 45, input_y + 36), "iMessage", fill=(99, 99, 102), font=font)
        
        # Microphone button (right, scaled)
        mic_x = input_field_x + input_field_width + 30
        draw.ellipse([mic_x, input_y + 15, mic_x + plus_size, input_y + 105], 
                    fill=(44, 44, 46))
        
        try:
            mic_font = ImageFont.truetype("/System/Library/Fonts/SF-Pro-Text-Medium.otf", 36)
        except:
            mic_font = font
        draw.text((mic_x + 24, input_y + 24), "🎤", fill=self.ios_text_primary, font=mic_font)
        
        return result

def create_sample_conversation():
    """Create a sample conversation matching the style of real screenshot"""
    base_time = datetime.now() - timedelta(hours=2)
    
    messages = [
        {
            'text': 'these probably will change a bit but\nthese are my grades atm', 
            'sender': 'jordan arrow',
            'timestamp': base_time
        },
        {
            'text': 'Wow!!!', 
            'is_sent': True,
            'timestamp': base_time + timedelta(minutes=2)
        },
        {
            'text': 'Great job.', 
            'is_sent': True,
            'timestamp': base_time + timedelta(minutes=2, seconds=30)
        },
        {
            'text': 'ok so monday is the deadline for\nteachers to put in grades', 
            'sender': 'jordan arrow',
            'timestamp': base_time + timedelta(hours=4)
        },
        {
            'text': 'Looks great babe 🥰', 
            'sender': 'Jen Arrow',
            'timestamp': base_time + timedelta(hours=4, minutes=20)
        },
        {
            'text': 'was the 1996 romeo and juliet the\none one of you wanted me to watch', 
            'sender': 'jordan arrow',
            'timestamp': base_time + timedelta(hours=4, minutes=25)
        },
        {
            'text': 'Yes', 
            'sender': 'Jen Arrow',
            'timestamp': base_time + timedelta(hours=4, minutes=26)
        },
        {
            'text': "i'm watching it in english rn", 
            'sender': 'jordan arrow',
            'timestamp': base_time + timedelta(hours=8)
        }
    ]
    
    return messages

def main():
    # Save output to current directory
    output_path = sys.argv[1] if len(sys.argv) > 1 else "./generated_imessage.png"
    
    try:
        generator = iMessageGenerator()
        
        # Create sample conversation
        messages = create_sample_conversation()
        
        # Generate the image
        result = generator.generate_conversation(messages, "Family Group")
        
        # Save result to current directory
        result.save(output_path)
        print(f"Generated realistic iMessage conversation: {output_path}")
        
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    main()