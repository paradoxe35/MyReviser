#!/usr/bin/env python3
"""Generate Scribe's icon set: a pen nib on a flat rounded tile."""

import os

from PIL import Image, ImageDraw

SUPERSAMPLE = 8
CORNER_RATIO = 0.2237  # Apple's continuous-corner approximation

TILE = (79, 70, 229)
NIB = (255, 255, 255)

NIB_HALF_WIDTH = 0.190
NIB_TOP = 0.215
NIB_TIP = 0.840
NIB_CROWN = 0.022
VENT_CENTER = 0.430
VENT_RADIUS = 0.046
SLIT_TOP_HALF = 0.028
SLIT_BOTTOM_HALF = 0.009
SLIT_END = 0.800

APP_SIZES = [16, 20, 22, 24, 32, 36, 40, 48, 64, 72, 96, 128, 192, 256, 512, 1024]
ICO_SIZES = [16, 24, 32, 48, 64, 128, 256]
TRAY_SIZES = [16, 32, 48]
TRAY_MARGIN = 0.06


def quadratic(p0, p1, p2, steps):
    for i in range(steps + 1):
        t = i / steps
        u = 1 - t
        yield (
            u * u * p0[0] + 2 * u * t * p1[0] + t * t * p2[0],
            u * u * p0[1] + 2 * u * t * p1[1] + t * t * p2[1],
        )


def nib_outline():
    left = (0.5 - NIB_HALF_WIDTH, NIB_TOP)
    right = (0.5 + NIB_HALF_WIDTH, NIB_TOP)
    tip = (0.5, NIB_TIP)
    shoulder = NIB_TOP + 0.34

    return (
        list(quadratic(left, (0.5, NIB_TOP - NIB_CROWN * 2), right, 48))
        + list(quadratic(right, (0.5 + NIB_HALF_WIDTH * 0.92, shoulder), tip, 64))
        + list(quadratic(tip, (0.5 - NIB_HALF_WIDTH * 0.92, shoulder), left, 64))
    )


def scaled(points, size):
    return [(x * size, y * size) for x, y in points]


def nib_mask(size):
    mask = Image.new("L", (size, size), 0)
    draw = ImageDraw.Draw(mask)

    draw.polygon(scaled(nib_outline(), size), fill=255)

    vent = VENT_RADIUS * size
    cx, cy = 0.5 * size, VENT_CENTER * size
    draw.ellipse((cx - vent, cy - vent, cx + vent, cy + vent), fill=0)

    draw.polygon(
        scaled(
            [
                (0.5 - SLIT_TOP_HALF, VENT_CENTER),
                (0.5 + SLIT_TOP_HALF, VENT_CENTER),
                (0.5 + SLIT_BOTTOM_HALF, SLIT_END),
                (0.5 - SLIT_BOTTOM_HALF, SLIT_END),
            ],
            size,
        ),
        fill=0,
    )
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
    icon.paste(Image.new("RGBA", (work, work), NIB + (255,)), (0, 0), nib_mask(work))
    return icon.resize((size, size), Image.LANCZOS)


def tray_icon(size):
    work = size * SUPERSAMPLE
    glyph = filled(work, NIB, nib_mask(work))
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
