"""Compose an editable selection from PDF paint, with a clean page under it."""

from contextlib import contextmanager
from io import BytesIO
import math

import pikepdf as pdf

from object_edit import ObjectError, isolated_paint, omit_draws, selected_page_items
from proof import box_options, page_size
from reader_render import render_screen
from text_edit import TextError, text_tree
from text_move import selected_text_paint


def preview_region(page, view):
    bounds, zoom = view.get("drag_bounds"), view.get("drag_zoom")
    if (not isinstance(bounds, list) or len(bounds) != 4
            or not all(type(x) in (int, float) and 0 <= x <= 1 for x in bounds)
            or bounds[0] >= bounds[2] or bounds[1] >= bounds[3]
            or type(zoom) not in (int, float) or not 0 < zoom <= 64):
        raise ObjectError("Náhled přesunu nemá platné rozměry.")
    box, _ = box_options(view)
    wide, high = page_size(page, box)
    if view.get("rotation", 0) % 180:
        wide, high = high, wide
    # Use the screen scale where possible. Limit the allocation for remote
    # or large items. Both layers always use the same pixel transform.
    width, height = (bounds[2]-bounds[0])*wide, (bounds[3]-bounds[1])*high
    zoom = min(zoom, 1520/max(width, height), math.sqrt(1_400_000/(width*height)))
    full_w, full_h = max(1, math.floor(wide*zoom+0.5)), max(1, math.floor(high*zoom+0.5))
    x, y = min(full_w-1, math.floor(bounds[0]*full_w)), min(full_h-1, math.floor(bounds[1]*full_h))
    r, b = max(x+1, math.ceil(bounds[2]*full_w)), max(y+1, math.ceil(bounds[3]*full_h))
    return dict(zoom=zoom, x=x, y=y, width=r-x, height=b-y), (x/full_w, y/full_h, r/full_w, b/full_h)


@contextmanager
def temporary_contents(page):
    contents = page.obj.get("/Contents")
    resources = page.resources
    try:
        yield
    finally:
        page.obj.Resources = resources
        if contents is not None:
            page.obj.Contents = contents
        elif "/Contents" in page.obj:
            del page.obj["/Contents"]


def paint_layer(document, page, number, operators, view, transparent, widgets):
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(operators))
    annots = page.obj.get("/Annots")
    # Copy the real widgets and their AcroForm field trees into the appropriate pass.
    try:
        if widgets:
            page.obj.Annots = pdf.Array(annot for index, annot in enumerate(annots, 1)
                                        if (-index in widgets) == transparent)
        with pdf.Pdf.new() as sample:
            return render_layer(document, number, sample, view, transparent, bool(widgets))
    finally:
        if widgets:
            page.obj.Annots = annots


def render_layer(document, number, sample, view, transparent, widgets):
    sample.add_pages_from(document, [number-1],
                          forms="strip" if transparent and not widgets else "preserve", outlines="ignore")
    if intents := document.Root.get("/OutputIntents"):
        sample.Root.OutputIntents = sample.copy_foreign(document.make_indirect(intents))
    sample.remove_unreferenced_resources()
    buffer = BytesIO()
    sample.save(buffer, fix_metadata_version=False)
    return render_screen(sample, buffer.getvalue(), {**view, "page": 1}, "", transparent=transparent, widgets=widgets)


def drag_preview(document, page, number, operators, candidates, edit, selected, view):
    items, widgets, _ = selected_page_items(document, page, candidates, edit) if edit.get("tokens") else ([], [], {})
    widgets = set(widgets)
    if not items and not selected and not widgets:
        raise ObjectError("Vyberte obsah k přesunu.")
    region, display = preview_region(page, view)
    view = {**view, "region": region}
    if selected:
        try:
            paint, remove, keep, underline = selected_text_paint(page, items, selected)
            graphics = {index: [] for index in paint[0]}
            with temporary_contents(page):
                clean, _ = text_tree(page, document, replacements=remove, discard=underline, graphics=graphics)
                background = paint_layer(document, page, number, clean, view, False, widgets)
            with temporary_contents(page):
                isolated, _ = text_tree(page, document, replacements=keep, paint_only=paint)
                foreground = paint_layer(document, page, number, isolated, view, True, widgets)
        except TextError as exc:
            raise ObjectError(str(exc)) from exc
    else:
        with temporary_contents(page):
            background = paint_layer(document, page, number, omit_draws(operators, items), view, False, widgets)
        with temporary_contents(page):
            foreground = paint_layer(document, page, number, isolated_paint(operators, items) if items else [], view, True, widgets)
    if background is None or foreground is None:
        raise ObjectError("Průhledný náhled výběru se nepodařilo vykreslit.")
    background["drag_bounds"] = display
    background["drag_raster"] = foreground["raster"]
    return background
