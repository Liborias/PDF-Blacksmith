"""Render screen previews in memory with the bundled PDFium library."""

import math

from proof import box_options, page_box, page_options, page_size, render_dpi, screen_region, screen_target


def render_screen(pdf, src, req, password):
    number, rotation = page_options(pdf, req)
    box, _ = box_options(req)
    target = screen_target(req, printed=False)
    source_page = pdf.pages[number-1]
    if "region" in req:
        _, (x, y, width, height), full_size = screen_region(source_page, box, rotation, req["region"])
    else:
        dpi = render_dpi(pdf, [number], 4608, target, box)
        full_size = tuple(max(1, math.floor(points * dpi / 72 + 0.5)) for points in page_size(source_page, box))
        x, y = 0, 0
        width, height = full_size[::-1] if rotation % 180 else full_size
    full_width, full_height = full_size[::-1] if rotation % 180 else full_size
    if max(full_width, full_height) >= (1 << 23):
        # PDFium uses floats for the page-to-device transform.
        return None

    try:
        import pypdfium2 as pdfium
    except (ImportError, OSError):
        return None
    raw = pdfium.raw
    bounds = page_box(source_page, box)
    try:
        # Release the document and its native caches before responding. The
        # open viewer must still allow the original PDF to be replaced.
        with src.open("rb") as stream, pdfium.PdfDocument(stream, password) as doc:
            if doc.get_formtype() in (raw.FORMTYPE_XFA_FULL, raw.FORMTYPE_XFA_FOREGROUND):
                return None
            doc.init_forms()
            page = doc[number-1]
            if box != "cropbox":
                page.set_cropbox(*bounds)
            if (page.get_rotation() != int(source_page.obj.get("/Rotate", 0)) % 360
                    or not all(math.isclose(expected, actual, abs_tol=0.01, rel_tol=1e-7)
                               for expected, actual in zip(bounds, page.get_bbox()))):
                return None
            bitmap = pdfium.PdfBitmap.new_native(width, height, raw.FPDFBitmap_BGRx)
            try:
                bitmap.fill_rect((255, 255, 255, 255), 0, 0, width, height)
                flags = raw.FPDF_ANNOT | raw.FPDF_RENDER_LIMITEDIMAGECACHE
                render_args = (bitmap, page, -x, -y, full_width, full_height, rotation//90, flags)
                raw.FPDF_RenderPageBitmap(*render_args)
                if doc.formenv:
                    raw.FPDF_FFLDraw(doc.formenv, *render_args)
                # A native bitmap's ctypes buffer outlives its PDFium handle.
                return dict(raster_width=width, raster_height=height, raster=memoryview(bitmap.buffer))
            finally:
                bitmap.close()
    except (pdfium.PdfiumError, IndexError):
        return None
