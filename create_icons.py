from PIL import Image, ImageDraw
import os

def create_squid_icon(size):
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    
    # Hot pink to purple colors
    colors = {
        'hot_pink': (255, 20, 147),
        'crimson': (255, 0, 100), 
        'purple': (138, 43, 226),
        'dark_purple': (75, 0, 130)
    }
    
    # Simple squid shape
    center = size // 2
    
    # Body
    body_radius = size // 3
    draw.ellipse([center - body_radius//2, center - body_radius, 
                  center + body_radius//2, center + body_radius//2], 
                 fill=colors['crimson'])
    
    # Head
    head_radius = size // 4
    draw.ellipse([center - head_radius//2, center - body_radius - head_radius//2,
                  center + head_radius//2, center - body_radius + head_radius//2],
                 fill=colors['hot_pink'])
    
    # Eyes
    eye_size = max(1, size // 16)
    if size >= 32:  # Only draw eyes for larger icons
        draw.ellipse([center - head_radius//3 - eye_size, center - body_radius - eye_size,
                      center - head_radius//3 + eye_size, center - body_radius + eye_size],
                     fill=(255, 255, 255))
        draw.ellipse([center + head_radius//3 - eye_size, center - body_radius - eye_size,
                      center + head_radius//3 + eye_size, center - body_radius + eye_size],
                     fill=(255, 255, 255))
        
        # Eye pupils
        pupil_size = max(1, eye_size // 2)
        draw.ellipse([center - head_radius//3 - pupil_size, center - body_radius - pupil_size,
                      center - head_radius//3 + pupil_size, center - body_radius + pupil_size],
                     fill=(45, 27, 105))
        draw.ellipse([center + head_radius//3 - pupil_size, center - body_radius - pupil_size,
                      center + head_radius//3 + pupil_size, center - body_radius + pupil_size],
                     fill=(45, 27, 105))
    
    # Tentacles (simple lines)
    tentacle_width = max(1, size // 32)
    for i in range(6):
        x_offset = (i - 2.5) * size // 12
        start_y = center + body_radius//2
        end_y = center + body_radius + size//4
        draw.line([center + x_offset, start_y, center + x_offset, end_y], 
                 fill=colors['purple'], width=tentacle_width)
    
    # Add some ink splashes for larger icons
    if size >= 48:
        ink_spots = [
            (center + size//3, center + size//3, size//20),
            (center - size//4, center + size//4, size//25),
            (center + size//5, center - size//6, size//30)
        ]
        for x, y, radius in ink_spots:
            draw.ellipse([x - radius, y - radius, x + radius, y + radius],
                        fill=(*colors['purple'], 128))  # Semi-transparent
    
    return img

# Create all required sizes
sizes = [16, 32, 48, 128]
os.makedirs('icons', exist_ok=True)

for size in sizes:
    icon = create_squid_icon(size)
    icon.save(f'icons/icon{size}.png')
    print(f'Created icon{size}.png')

print("🦑 All icons created successfully!") 