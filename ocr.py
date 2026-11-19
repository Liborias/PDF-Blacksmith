"""Recognize page images and keep only Ghostscript's invisible text."""

from contextlib import closing
from pathlib import Path

import pikepdf
import pypdfium2 as pdfium

from proof import page_box, render_dpi


LANGUAGES = ("ces+eng", "ces", "eng")
BATCH_SIZE = 8


class OCRError(ValueError):
    pass


def checked_models(directory, language):
    path = Path(directory).resolve()
    if language not in LANGUAGES:
        raise OCRError("Vyberte podporovaný jazyk OCR.")
    if any(not (path / f"{code}.traineddata").is_file() for code in language.split("+")):
        raise OCRError("V aplikaci chybí jazyková data OCR. Použijte úplné sestavení PDF Blacksmith.")
    return path


def placement(original, layer):
    """Map the upright OCR output back to PDF user coordinates."""
    left, bottom, right, top = page_box(original, "cropbox")
    x, y, r, t = page_box(layer, "mediabox")
    width, height = r-x, t-y
    if min(width, height, right-left, top-bottom) <= 0:
        raise OCRError("Stránka OCR má neplatné rozměry.")
    rotation = int(original.obj.get("/Rotate", 0)) % 360
    if rotation == 90:
        a, b, c, d, e, f = 0, (top-bottom)/width, (left-right)/height, 0, right, bottom
    elif rotation == 180:
        a, b, c, d, e, f = (left-right)/width, 0, 0, (bottom-top)/height, right, top
    elif rotation == 270:
        a, b, c, d, e, f = 0, (bottom-top)/width, (right-left)/height, 0, left, top
    else:
        a, b, c, d, e, f = (right-left)/width, 0, 0, (top-bottom)/height, left, bottom
    return a, b, c, d, e-a*x-c*y, f-b*x-d*y


def map_bounds(bounds, matrix):
    left, bottom, right, top = bounds
    a, b, c, d, e, f = matrix
    x = [a*u+c*v+e for u in (left, right) for v in (bottom, top)]
    y = [b*u+d*v+f for u in (left, right) for v in (bottom, top)]
    return min(x), min(y), max(x), max(y)


def existing_text(page):
    with closing(page.get_textpage()) as text:
        return [text.get_rect(index) for index in range(text.count_rects())]


def already_searchable(bounds, existing, along_y):
    left, bottom, right, top = bounds
    if along_y:
        left, bottom, right, top = bottom, left, top, right
    covered = []
    for a, b, c, d in existing:
        if along_y:
            a, b, c, d = b, a, d, c
        start, end = max(a, left), min(c, right)
        if start < end and min(top, d)-max(bottom, b) >= 0.5*min(top-bottom, d-b):
            covered.append((start, end))
    overlap, edge = 0, left
    for start, end in sorted(covered):
        if end > edge:
            overlap += end - max(start, edge)
            edge = end
    return overlap >= 0.55*(right-left)


def invisible_words(layer, matrix, existing):
    """Copy only the pinned pdfocr8 text format, and force render mode 3."""
    instructions = []
    font, size, stretch, origin = None, 0, 100, None
    words = 0
    for args, operator in pikepdf.parse_content_stream(layer):
        name = str(operator)
        if name == "Tf":
            font, size = args[0], float(args[1])
        elif name == "Tz":
            stretch = float(args[0])
        elif name == "Tm":
            if tuple(args[:4]) != (1, 0, 0, 1):
                raise OCRError("Rozpoznaný text má nepodporovanou orientaci.")
            origin = float(args[4]), float(args[5])
        elif name == "TJ":
            # Ghostscript emits one UTF-16BE string per word, in GlyphLessFont
            # with a default glyph width of 500 units.
            if (font is None or origin is None or size <= 0 or stretch <= 0 or
                    len(args) != 1 or len(args[0]) != 1 or not isinstance(args[0][0], pikepdf.String)):
                raise OCRError("Ghostscript vrátil nepodporovanou textovou vrstvu OCR.")
            source_font = layer.resources.Font[font]
            if source_font.BaseFont != pikepdf.Name.GlyphLessFont or source_font.DescendantFonts[0].DW != 500:
                raise OCRError("Ghostscript vrátil nepodporované písmo OCR.")
            glyphs = len(bytes(args[0][0])) / 2
            if not glyphs:
                continue
            x, y = origin
            word = (x, y, x + glyphs*size*stretch/200, y + size)
            if not already_searchable(map_bounds(word, matrix), existing, matrix[0] == 0):
                instructions.extend((([font, size], pikepdf.Operator("Tf")),
                                     ([stretch], pikepdf.Operator("Tz")),
                                     ([1, 0, 0, 1, x, y], pikepdf.Operator("Tm")),
                                     (args, operator)))
                words += 1
        elif name == "Tj":
            raise OCRError("Ghostscript vrátil nepodporovanou textovou vrstvu OCR.")
    return instructions, words


def append_text(pdf, page, layer_pdf, layer, existing):
    matrix = placement(page, layer)
    instructions, words = invisible_words(layer, matrix, existing)
    if not words:
        return 0
    data = b"BT\n3 Tr\n" + pikepdf.unparse_content_stream(instructions) + b"\nET\n"
    fonts = pdf.copy_foreign(layer_pdf.make_indirect(layer.resources.Font))
    form = pdf.make_stream(data, Type=pikepdf.Name.XObject, Subtype=pikepdf.Name.Form,
                           BBox=pikepdf.Array(layer.mediabox), Resources=pikepdf.Dictionary(Font=fonts))
    resources = pikepdf.Dictionary(page.resources)
    resources.XObject = pikepdf.Dictionary(resources.get("/XObject", pikepdf.Dictionary()))
    page.obj.Resources = resources
    name = page.add_resource(form, pikepdf.Name.XObject, prefix="OCR")
    transform = " ".join(f"{value:.10f}" for value in matrix)
    page.contents_add(b"q\n", prepend=True)
    page.contents_add(f"\nQ\nq\n{transform} cm\n{name} Do\nQ\n".encode("ascii"))
    return words


def recognize(pdf, source, pages, folder, exe, gs, stage, language, model_dir):
    models = checked_models(model_dir, language)
    staged = folder / "source.pdf"
    stage(pdf, source, staged)
    output = folder / "layer.pdf"
    updated, words = 0, 0
    with staged.open("rb") as stream, pdfium.PdfDocument(stream) as document:
        for offset in range(0, len(pages), BATCH_SIZE):
            batch = pages[offset:offset+BATCH_SIZE]
            dpi = render_dpi(pdf, batch, 300, 6000, "cropbox")
            gs(exe, folder, ["-q", "-sDEVICE=pdfocr8", f"-r{dpi}", "-dUseCropBox", "-dPrinted=false",
                             "-dShowAnnots=false", f"-sOCRLanguage={language}",
                             "-sPageList=" + ",".join(map(str, batch)),
                             f"-sOutputFile={output.name}", "-f", staged.name],
               extra_env={"TESSDATA_PREFIX": str(models), "OMP_THREAD_LIMIT": "2"})
            with pikepdf.Pdf.open(output) as layer_pdf:
                if len(layer_pdf.pages) != len(batch):
                    raise OCRError("Ghostscript nevrátil textovou vrstvu pro všechny vybrané strany.")
                for number, layer in zip(batch, layer_pdf.pages):
                    with closing(document[number-1]) as page:
                        added = append_text(pdf, pdf.pages[number-1], layer_pdf, layer, existing_text(page))
                    updated += bool(added)
                    words += added
            output.unlink()
    return updated, words
