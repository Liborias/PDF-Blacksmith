"""Build 4× alpha masks for Win32 from the original v3.26 SVG symbols."""

from pathlib import Path
from tempfile import TemporaryDirectory
import json
import re
import subprocess

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs" / "pracovni"
ATLAS = ROOT / "assets" / "ui_icons"
ACTUAL_SIZE_ICON = ROOT / "assets" / "ui_zoom_1_1.png"
TILE_SIZE = 96
SYMBOLS = (
    "cursor", "fit_page", "fit_width", "language", "lasso", "nav_file", "nav_merge_pdf", "nav_objects",
    "nav_ocr", "nav_page_order", "nav_search_pdf", "nav_separations",
    "nav_settings", "nav_signature", "nav_split_pdf", "nav_text", "nav_thumbnails", "nav_view", "panel_right", "paragraph",
    "rect_select", "redo", "save", "search", "settings_interface", "side_by_side", "text_caret",
    "undo", "view_continuous", "view_pages",
    "zoom_in", "zoom_out", "text_box", "image_crop",
)


def main():
    reference = (DOCS / "rozhrani" / "cely_vizual_v3.26.html").read_text(encoding="utf-8")
    original_tags = {
        re.search(r'\bid="(i-[^"]+)"', item)[1]: item
        for item in re.findall(r'<symbol\b[^>]*\bid="i-[^"]+"[^>]*>.*?</symbol>', reference, re.S)
    }
    ids = ["i-" + name for name in SYMBOLS]
    registry = json.loads((DOCS / "mapa_ikon_FINAL_v6.0.json").read_text(encoding="utf-8"))
    approved = {item["symbol"] for item in registry["icons"]}
    # The reference draws this icon inline in the toolbar instead of the SVG registry.
    quick_mode = re.search(
        r'<button\b[^>]*\bid="themeQuickToggle"[^>]*>\s*<svg\b[^>]*\bviewBox="([^"]+)"[^>]*>(.*?)</svg>',
        reference, re.S)
    if not quick_mode:
        raise ValueError("Missing the original light/dark icon")
    mode_id = "i-theme_quick_inline"
    original_tags[mode_id] = f'<symbol id="{mode_id}" viewBox="{quick_mode[1]}">{quick_mode[2]}</symbol>'
    ids.append(mode_id)
    ids.extend("i-" + name for name in (
        "crop_mask", "background_mask", "mask_move_content", "mask_aspect_lock", "mask_invert", "crossing_selection",
    ))
    actual_id = "i-zoom_1_1"
    for name in [*ids, actual_id]:
        if name == mode_id:
            continue
        if name not in original_tags or name not in approved:
            raise ValueError(f"Missing or unapproved icon: {name}")
    width = 24 * len(ids)
    definitions = "".join(original_tags[name] for name in ids)
    tiles = "".join(f'<use x="{24 * index}" width="24" height="24" xlink:href="#{name}"/>'
                    for index, name in enumerate(ids))
    svg = (f'<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" '
           f'width="{width}" height="24" viewBox="0 0 {width} 24">'
           f'<defs>{definitions}</defs>{tiles}</svg>')
    with TemporaryDirectory() as folder:
        source = Path(folder) / "atlas.svg"
        source.write_text(svg, encoding="utf-8")
        subprocess.run(["inkscape", str(source), "--export-type=png",
                        f"--export-filename={ATLAS.with_suffix('.png')}",
                        f"--export-width={TILE_SIZE * len(ids)}"], check=True)
        # Use the same SVG/renderer as the existing atlas without changing its slots.
        actual_svg = (
            '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" '
            'width="24" height="24" viewBox="0 0 24 24"><defs>'
            + original_tags[actual_id] + '</defs><use x="0" width="24" height="24" '
            f'xlink:href="#{actual_id}"/></svg>'
        )
        source.write_text(actual_svg, encoding="utf-8")
        subprocess.run(["inkscape", str(source), "--export-type=png",
                        f"--export-filename={ACTUAL_SIZE_ICON}",
                        f"--export-width={TILE_SIZE}"], check=True)
    ATLAS.with_suffix(".json").write_text(
        json.dumps({"tile_size": TILE_SIZE, "symbols": ids}, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
