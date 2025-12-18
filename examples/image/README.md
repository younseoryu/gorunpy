# Image Processing Example

Image manipulation using [Pillow](https://pillow.readthedocs.io/).

## Setup

```bash
cd examples/image

python -m venv .venv  # or python3 -m venv .venv
source .venv/bin/activate
pip install "gorunpy[build]" Pillow

gorunpy
```

## Run

```bash
go run . photo.jpg
```

## Exported Functions

| Function | Description |
|----------|-------------|
| `GetInfo(path)` | Get image metadata (format, size, mode) |
| `Resize(path, w, h, out)` | Resize to exact dimensions |
| `Thumbnail(path, maxSize, out)` | Create thumbnail preserving aspect ratio |
| `ConvertFormat(path, out, fmt)` | Convert to PNG, JPEG, WEBP, etc. |
| `ToBase64(path)` | Return image as base64 string |
| `ApplyFilter(path, filter, out)` | Apply blur, sharpen, grayscale, etc. |

## Available Filters

- `blur` - Gaussian blur
- `sharpen` - Sharpen edges
- `contour` - Find contours
- `emboss` - Emboss effect
- `edge` - Edge detection
- `grayscale` - Convert to grayscale

