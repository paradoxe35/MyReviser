#!/usr/bin/env python3
"""Generate Encre's icon set: an ink drop on a flat rounded tile."""

import math
import os

from PIL import Image, ImageDraw

SUPERSAMPLE = 8
CORNER_RATIO = 0.2237  # Apple's continuous-corner approximation

TILE = (79, 70, 229)
INK = (255, 255, 255)

# A teardrop: apex on top, circular bottom, sides running along the tangents
# from the apex to that circle. Straight tangents keep the silhouette crisp at
# 16 px where a fussier curve turns to mush.
DROP_CENTER_Y = 0.605
DROP_RADIUS = 0.255
DROP_APEX_Y = 0.145

APP_SIZES = [16, 20, 22, 24, 32, 36, 40, 48, 64, 72, 96, 128, 192, 256, 512, 1024]
ICO_SIZES = [16, 24, 32, 48, 64, 128, 256]
TRAY_SIZES = [16, 32, 48]
TRAY_MARGIN = 0.06


def drop_outline():
    center_y, radius, apex_y = DROP_CENTER_Y, DROP_RADIUS, DROP_APEX_Y
    height = center_y - apex_y

    # Where the tangent from the apex touches the circle.
    cos_beta = radius / height
    cos_beta = max(-1.0, min(1.0, cos_beta))
    beta = math.acos(cos_beta)

    points = [(0.5, apex_y)]

    steps = 96
    start = -math.pi / 2 + beta          # right tangent point
    end = start + (2 * math.pi - 2 * beta)  # sweep the long way, around the bottom
    for i in range(steps + 1):
        angle = start + (end - start) * i / steps
        points.append((0.5 + radius * math.cos(angle), center_y + radius * math.sin(angle)))

    return points


def scaled(points, size):
    return [(x * size, y * size) for x, y in points]


def drop_mask(size):
    mask = Image.new("L", (size, size), 0)
    draw = ImageDraw.Draw(mask)

    draw.polygon(scaled(drop_outline(), size), fill=255)
    return mask


def tile_mask(size):
    mask = Image.new("L", (size, size), 0)
    ImageDraw.Draw(mask).rounded_rectangle(
        (0, 0, size - 1, size - 1), radius=size * CORNER_RATIO, fill=255
    )
    return mask


def filled(size, color, mask):
    layer = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    layer.paste(Image.new("RGBA", (size, size), color + (255,)), (0, 0), mask)
    return layer


def app_icon(size):
    work = size * SUPERSAMPLE
    icon = filled(work, TILE, tile_mask(work))
    icon.paste(Image.new("RGBA", (work, work), INK + (255,)), (0, 0), drop_mask(work))
    return icon.resize((size, size), Image.LANCZOS)


def tray_icon(size):
    work = size * SUPERSAMPLE
    glyph = filled(work, INK, drop_mask(work))
    glyph = glyph.crop(glyph.getbbox())

    inner = round(size * (1 - TRAY_MARGIN * 2))
    scale = min(inner / glyph.width, inner / glyph.height)
    glyph = glyph.resize(
        (max(round(glyph.width * scale), 1), max(round(glyph.height * scale), 1)),
        Image.LANCZOS,
    )

    icon = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    icon.paste(glyph, ((size - glyph.width) // 2, (size - glyph.height) // 2))
    return icon


def main():
    root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    assets = os.path.join(root, "assets")
    build = os.path.join(root, "build")
    os.makedirs(assets, exist_ok=True)
    os.makedirs(build, exist_ok=True)

    icons = {size: app_icon(size) for size in APP_SIZES}
    for size, icon in icons.items():
        icon.save(os.path.join(assets, f"icon_{size}.png"))

    icons[256].save(os.path.join(assets, "icon.png"))
    icons[1024].save(os.path.join(build, "appicon.png"))
    icons[256].save(
        os.path.join(assets, "icon.ico"),
        format="ICO",
        sizes=[(s, s) for s in ICO_SIZES],
    )

    for size in TRAY_SIZES:
        tray_icon(size).save(os.path.join(assets, f"icon-{size}x{size}.png"))

    print(f"app icons   {APP_SIZES}")
    print(f"windows ico {ICO_SIZES}")
    print(f"tray icons  {TRAY_SIZES} (monochrome)")


if __name__ == "__main__":
    main()
