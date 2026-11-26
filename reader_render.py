"""Render screen previews with PDFium and keep the active source open for its reader."""

from contextlib import ExitStack, nullcontext
import math
import os

from proof import box_options, page_box, page_options, page_size, render_dpi, screen_region, screen_target


def reader_file(path):
    if os.name != "nt":
        return path.open("rb")

    # A reader may still hold a snapshot while the user saves an atomic
    # replacement or closes its tab. Let Windows rename/delete the open file.
    import ctypes
    from ctypes import wintypes
    import msvcrt

    kernel = ctypes.WinDLL("kernel32", use_last_error=True)
    create = kernel.CreateFileW
    create.argtypes = (wintypes.LPCWSTR, wintypes.DWORD, wintypes.DWORD, wintypes.LPVOID,
                       wintypes.DWORD, wintypes.DWORD, wintypes.HANDLE)
    create.restype = wintypes.HANDLE
    handle = create(os.fspath(path), 0x80000000, 7, None, 3, 0x80, None)
    if handle == ctypes.c_void_p(-1).value:
        raise ctypes.WinError(ctypes.get_last_error())
    try:
        fd = msvcrt.open_osfhandle(handle, os.O_RDONLY | os.O_BINARY)
    except OSError:
        kernel.CloseHandle.argtypes = (wintypes.HANDLE,)
        kernel.CloseHandle(handle)
        raise
    return os.fdopen(fd, "rb")


class ScreenReader:
    """One document and one parsed PDFium page, local to a single worker."""

    def __init__(self):
        self.resources = ExitStack()
        self.key = self.pdf = self.doc = self.page = None
        self.page_number = 0

    def close(self):
        self.resources.close()
        self.key = self.pdf = self.doc = self.page = None
        self.page_number = 0

    def document(self, path, password, open_pdf):
        stat = path.stat()
        key = path, stat.st_dev, stat.st_ino, stat.st_size, stat.st_mtime_ns, password
        if self.key != key:
            self.close()
            with ExitStack() as loading:
                stream = loading.enter_context(reader_file(path))
                self.pdf = loading.enter_context(open_pdf(stream, password))
                self.resources = loading.pop_all()
            self.key = key
        return self.pdf

    def native_page(self, pdfium, path, password, number):
        if self.doc is None:
            with ExitStack() as loading:
                stream = loading.enter_context(reader_file(path))
                doc = loading.enter_context(pdfium.PdfDocument(stream, password))
                init_screen_forms(doc, pdfium.raw)
                self.doc = doc
                self.resources.enter_context(loading.pop_all())
        if number != self.page_number:
            if self.page is not None:
                self.page.close()
                self.page = None
                self.page_number = 0
            self.page = self.doc[number-1]
            self.page_number = number
        return self.doc, self.page


def init_screen_forms(doc, raw):
    if doc.get_formtype() in (raw.FORMTYPE_XFA_FULL, raw.FORMTYPE_XFA_FOREGROUND):
        # Retain the Ghostscript fallback for documents outside this renderer.
        raise UnsupportedPreview()
    doc.init_forms()


class UnsupportedPreview(Exception):
    pass


def paint_screen(pdfium, doc, page, rect, full_size, rotation, *, transparent=False, widgets=False):
    raw = pdfium.raw
    x, y, width, height = rect
    mode = raw.FPDFBitmap_BGRA if transparent else raw.FPDFBitmap_BGRx
    bitmap = pdfium.PdfBitmap.new_native(width, height, mode)
    try:
        color = (0, 0, 0, 0) if transparent else (255, 255, 255, 255)
        bitmap.fill_rect(color, 0, 0, width, height)
        flags = raw.FPDF_RENDER_LIMITEDIMAGECACHE
        if not transparent or widgets:
            flags |= raw.FPDF_ANNOT
        render_args = (bitmap, page, -x, -y, *full_size, rotation//90, flags)
        raw.FPDF_RenderPageBitmap(*render_args)
        if doc.formenv and (not transparent or widgets):
            raw.FPDF_FFLDraw(doc.formenv, *render_args)
        # The ctypes buffer outlives the PDFium bitmap handle.
        return memoryview(bitmap.buffer)
    finally:
        bitmap.close()


def render_screen(pdf, src, req, password, *, thumbnail=False, reader=None, transparent=False, widgets=False):
    number, rotation = page_options(pdf, req)
    box, _ = box_options(req)
    target = screen_target(req, printed=False, thumbnail=thumbnail)
    source_page = pdf.pages[number-1]
    if "region" in req:
        scale, (x, y, width, height), full_size = screen_region(source_page, box, rotation, req["region"])
    else:
        dpi = render_dpi(pdf, [number], 4608, target, box, fractional=thumbnail)
        scale = dpi / 72
        full_size = tuple(max(1, math.floor(points * scale + 0.5)) for points in page_size(source_page, box))
        x, y = 0, 0
        width, height = full_size[::-1] if rotation % 180 else full_size
    device_size = full_size[::-1] if rotation % 180 else full_size
    if max(device_size) >= (1 << 23):
        # PDFium uses floats for the page-to-device transform.
        return None

    try:
        import pypdfium2 as pdfium
    except (ImportError, OSError):
        return None
    bounds = page_box(source_page, box)
    angle = int(source_page.obj.get("/Rotate", 0)) % 360
    try:
        with ExitStack() as temporary:
            if reader is not None:
                doc, page = reader.native_page(pdfium, src, password, number)
            else:
                source = nullcontext(src) if isinstance(src, bytes) else reader_file(src)
                stream = temporary.enter_context(source)
                doc = temporary.enter_context(pdfium.PdfDocument(stream, password))
                init_screen_forms(doc, pdfium.raw)
                page = doc[number-1]
            # PDFium's loaded page is reused. Set the requested box again
            # when switching from a wider/narrower box on the same page.
            if (box != "cropbox" or reader is not None) and tuple(page.get_cropbox()) != tuple(bounds):
                page.set_cropbox(*bounds)
            if (page.get_rotation() != angle
                    or not all(math.isclose(expected, actual, abs_tol=0.01, rel_tol=1e-7)
                               for expected, actual in zip(bounds, page.get_bbox()))):
                return None
            pixels = paint_screen(pdfium, doc, page, (x, y, width, height), device_size, rotation,
                                  transparent=transparent, widgets=widgets)
            return dict(raster_width=width, raster_height=height, raster_scale=scale, raster=pixels)
    except (pdfium.PdfiumError, UnsupportedPreview, IndexError):
        return None
