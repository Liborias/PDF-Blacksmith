"""Generate a transparent Windows ICO from the approved original HTML brand SVG."""

from pathlib import Path
import re
import struct
import subprocess
import sys
from tempfile import TemporaryDirectory
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
REFERENCE = ROOT / "docs/pracovni/rozhrani/cely_vizual_v3.26.html"
WINE = "#AD5B80"  # approved wine/dark accentStrong; fixed brand color
SIZES = (16, 24, 32, 48, 64, 96, 128, 256)
SVG = "http://www.w3.org/2000/svg"
ET.register_namespace("", SVG)


def brand_svg(reference):
    match = re.search(r'<button\b[^>]*\bclass="brand-mark"[^>]*>\s*(<svg\b.*?</svg>)',
                      reference, re.S)
    if not match:
        raise ValueError("Missing approved brand SVG")
    original = ET.fromstring(match[1])
    if original.tag != f"{{{SVG}}}svg" or original.get("viewBox") != "13 69 177 162":
        raise ValueError("Unexpected brand SVG geometry")
    paths = list(original)
    if len(paths) != 1 or paths[0].tag != f"{{{SVG}}}path" or paths[0].get("fill") != "currentColor":
        raise ValueError("Unexpected brand path; review changes to the reference")
    # Retain the original path and viewBox. A square transparent canvas gives
    # the icon equal breathing room without distorting the geometry.
    paths[0].set("fill", WINE)
    for key, value in {"x": "8", "y": "8", "width": "240", "height": "240"}.items():
        original.set(key, value)
    outer = ET.Element(f"{{{SVG}}}svg", {"width": "256", "height": "256", "viewBox": "0 0 256 256"})
    outer.append(original)
    return ET.tostring(outer, encoding="utf-8", xml_declaration=True)


def write_ico(svg, output):
    with TemporaryDirectory() as folder:
        source = Path(folder) / "brand.svg"
        source.write_bytes(svg)
        images = []
        for size in SIZES:
            png = Path(folder) / f"brand-{size}.png"
            subprocess.run(["inkscape", str(source), "--export-type=png",
                            f"--export-filename={png}", f"--export-width={size}",
                            "--export-background-opacity=0"], check=True)
            data = png.read_bytes()
            if (data[:8] != b"\x89PNG\r\n\x1a\n" or
                    struct.unpack_from(">II", data, 16) != (size, size) or data[25] != 6):
                raise ValueError(f"Expected transparent RGBA PNG: {size}px")
            images.append(data)
    directory = bytearray(struct.pack("<HHH", 0, 1, len(images)))
    offset = 6 + 16 * len(images)
    for size, data in zip(SIZES, images):
        dimension = size % 256  # zero denotes 256px in ICO
        directory.extend(struct.pack("<BBBBHHII", dimension, dimension, 0, 0, 1, 32, len(data), offset))
        offset += len(data)
    output.write_bytes(directory + b"".join(images))


if __name__ == "__main__":
    write_ico(brand_svg(REFERENCE.read_text(encoding="utf-8")), Path(sys.argv[1]))
