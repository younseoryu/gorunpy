"""Image processing using Pillow.

Common image operations exposed to Go.
"""

import gorunpy
import base64
import io
from typing import Optional


@gorunpy.export
def get_info(file_path: str) -> dict:
    """Get image metadata (format, size, mode)."""
    from PIL import Image
    
    with Image.open(file_path) as img:
        return {
            "format": img.format,
            "width": img.width,
            "height": img.height,
            "mode": img.mode,
        }


@gorunpy.export
def resize(
    file_path: str,
    width: int,
    height: int,
    output_path: str,
) -> str:
    """Resize image to specified dimensions."""
    from PIL import Image
    
    with Image.open(file_path) as img:
        resized = img.resize((width, height))
        resized.save(output_path)
    return output_path


@gorunpy.export
def thumbnail(
    file_path: str,
    max_size: int,
    output_path: str,
) -> str:
    """Create thumbnail maintaining aspect ratio."""
    from PIL import Image
    
    with Image.open(file_path) as img:
        img.thumbnail((max_size, max_size))
        img.save(output_path)
    return output_path


@gorunpy.export
def convert_format(
    file_path: str,
    output_path: str,
    format: str,
) -> str:
    """Convert image to different format (PNG, JPEG, WEBP, etc.)."""
    from PIL import Image
    
    with Image.open(file_path) as img:
        # Convert RGBA to RGB for JPEG
        if format.upper() == "JPEG" and img.mode == "RGBA":
            img = img.convert("RGB")
        img.save(output_path, format=format.upper())
    return output_path


@gorunpy.export
def to_base64(file_path: str) -> str:
    """Read image and return as base64 string."""
    from PIL import Image
    
    with Image.open(file_path) as img:
        buffer = io.BytesIO()
        img.save(buffer, format=img.format or "PNG")
        return base64.b64encode(buffer.getvalue()).decode("utf-8")


@gorunpy.export
def apply_filter(
    file_path: str,
    filter_name: str,
    output_path: str,
) -> str:
    """Apply filter: blur, sharpen, contour, emboss, grayscale."""
    from PIL import Image, ImageFilter
    
    filters = {
        "blur": ImageFilter.BLUR,
        "sharpen": ImageFilter.SHARPEN,
        "contour": ImageFilter.CONTOUR,
        "emboss": ImageFilter.EMBOSS,
        "edge": ImageFilter.FIND_EDGES,
    }
    
    with Image.open(file_path) as img:
        if filter_name == "grayscale":
            result = img.convert("L")
        elif filter_name in filters:
            result = img.filter(filters[filter_name])
        else:
            raise gorunpy.ValidationError(
                f"Unknown filter: {filter_name}. Available: {', '.join(list(filters.keys()) + ['grayscale'])}"
            )
        result.save(output_path)
    return output_path

