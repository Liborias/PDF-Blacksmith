"""Search the PDF text layer and report matches in PDF page coordinates."""

import math
import re
from contextlib import closing

import pypdfium2 as pdfium

from proof import BOXES, page_box


RESULT_LIMIT = 2000


def search_pdf(pdf, source, password, order, query):
    total = 0
    pattern = re.compile(r"\s+".join(map(re.escape, query.split())), re.IGNORECASE)
    with source.open("rb") as stream, pdfium.PdfDocument(stream, password) as doc:
        for position, number in enumerate(order, 1):
            original = pdf.pages[number - 1]
            hits = []
            with closing(doc[number - 1]) as page:
                # Search the full sheet, including text outside the current crop.
                page.set_cropbox(*page_box(original, "mediabox"))
                with closing(page.get_textpage()) as text:
                    chars = text.count_chars()
                    if chars > 0:
                        total = search_page(text, chars, query, pattern, total, hits)
            if hits:
                boxes = {name: page_box(original, attr) for name, (attr, _) in BOXES.items()}
                rotation = int(original.obj.get("/Rotate", 0)) % 360
                for offset in range(0, len(hits), 64):
                    result = dict(number=number, position=position, hits=hits[offset:offset + 64],
                                  rotation=rotation, boxes=boxes)
                    yield dict(status="progress", scanned=position, total=total, page=result)
            elif position % 16 == 0:
                yield dict(status="progress", scanned=position, total=total)
    yield dict(status="ok", scanned=len(order), total=total)


def search_page(text, chars, query, pattern, total, hits):
    # PDFium's native case-insensitive search depends on the C runtime locale.
    # Read its character list so Python can match Unicode consistently and
    # each match still has the exact character indices expected by PDFium.
    get_unicode = pdfium.raw.FPDFText_GetUnicode
    source = "".join(chr(value) if 0 < value <= 0x10ffff else "\ufffd"
                     for value in (get_unicode(text, index) for index in range(chars)))
    rects_ready = False
    for match in pattern.finditer(source):
        total += 1
        if total > RESULT_LIMIT:
            continue
        index, last = match.span()
        if not rects_ready:
            text.count_rects()
            rects_ready = True
        rects = []
        for rect in range(text.count_rects(index, last-index)):
            bounds = text.get_rect(rect)
            if all(map(math.isfinite, bounds)):
                rects.append(bounds)
        start, end = max(0, index - 20), min(chars, last + 40)
        snippet = " ".join(source[start:end].split())
        if start:
            snippet = "… " + snippet
        if end < chars:
            snippet += " …"
        hits.append(dict(text=snippet or query, rects=rects))
    return total
