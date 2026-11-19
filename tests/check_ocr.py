"""Run with an OCR-enabled Ghostscript executable and a tessdata directory."""

from contextlib import closing
from pathlib import Path
from tempfile import TemporaryDirectory
import sys

import pikepdf
import pypdfium2 as pdfium
from PIL import Image, ImageChops, ImageDraw, ImageFont

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from engine import UserError, dispatch, run_ghostscript


def text_page(pdf, text, y):
    page = pdf.add_blank_page(page_size=(420, 300))
    font = pdf.make_indirect(pikepdf.Dictionary(Type=pikepdf.Name.Font, Subtype=pikepdf.Name.Type1,
                                                BaseFont=pikepdf.Name.Helvetica))
    page.obj.Resources = pikepdf.Dictionary(Font=pikepdf.Dictionary(F1=font))
    page.obj.Contents = pdf.make_stream(f"BT /F1 29 Tf 35 {y} Td ({text}) Tj ET".encode("ascii"))
    return page


def fixture(folder, exe):
    with pikepdf.Pdf.new() as text:
        text_page(text, "VECTOR SAMPLE", 202)
        text.save(folder / "text.pdf")
    run_ghostscript(exe, folder, ["-q", "-sDEVICE=pdfwrite", "-dNoOutputFonts",
                                 "-sOutputFile=curves.pdf", "-f", "text.pdf"])
    with pikepdf.Pdf.open(folder / "curves.pdf") as curves, pikepdf.Pdf.new() as pdf:
        curve_data = pikepdf.unparse_content_stream(pikepdf.parse_content_stream(curves.pages[0]))
        pdf.pages.append(curves.pages[0])
        pdf.pages[0].obj.Contents = pdf.make_stream(curve_data)
        with pikepdf.Pdf.new() as native:
            overlay = text_page(native, "NATIVE HEADING", 96)
            pdf.pages[0].add_overlay(overlay)
        scan = pdf.add_blank_page(page_size=(420, 300))
        image = Image.new("RGB", (840, 600), "white")
        ImageDraw.Draw(image).text((70, 170), "SCANNED SAMPLE", font=ImageFont.load_default(size=57), fill="black")
        stream = pdf.make_stream(image.tobytes(), Type=pikepdf.Name.XObject, Subtype=pikepdf.Name.Image,
                                 Width=image.width, Height=image.height, ColorSpace=pikepdf.Name.DeviceRGB,
                                 BitsPerComponent=8)
        scan.obj.Resources = pikepdf.Dictionary(XObject=pikepdf.Dictionary(Scan=stream))
        scan.obj.Contents = pdf.make_stream(b"q 420 0 0 300 0 0 cm /Scan Do Q")
        pdf.pages.append(curves.pages[0])
        pdf.pages[-1].obj.Contents = pdf.make_stream(curve_data)
        pdf.pages.append(curves.pages[0])
        rotated = pdf.pages[-1]
        rotated.obj.Contents = pdf.make_stream(curve_data)
        rotated.obj.MediaBox = pikepdf.Array([0, 0, 300, 420])
        rotated.obj.CropBox = pikepdf.Array([10, 15, 287, 400])
        rotated.contents_add(b"q 0 1 -1 0 300 0 cm\n", prepend=True)
        rotated.contents_add(b"\nQ")
        rotated.rotate(90)
        source = folder / "input.pdf"
        pdf.save(source)
    return source


def read_page(document, index):
    with closing(document[index]) as page, closing(page.get_textpage()) as text:
        with closing(page.render(scale=1)) as bitmap:
            return " ".join(text.get_text_range().split()), bitmap.to_pil().convert("RGB")


def check_ocr(folder, exe, models):
    source = fixture(folder, exe)
    output = folder / "recognized.pdf"
    request = dict(action="ocr", input=str(source), output=str(output), ghostscript=exe, ocr_models=models,
                   ocr_language="eng", view={"pages": [1, 2, 4]})
    original_bytes = source.read_bytes()
    description = dispatch(request)
    assert len(description["sizes"]) == 4 and source.read_bytes() == original_bytes
    with pikepdf.Pdf.open(output) as pdf:
        assert not pdf.check_pdf_syntax()
        assert len(pdf.pages) == 4
        assert int(pdf.pages[3].obj.Rotate) == 90
        xobjects = pdf.pages[0].resources.XObject
        hidden = [obj for name, obj in xobjects.items() if str(name).startswith("/OCR")]
        assert len(hidden) == 1 and "/XObject" not in hidden[0].Resources
    with pdfium.PdfDocument(source) as before, pdfium.PdfDocument(output) as after:
        contents = []
        for index in range(len(before)):
            _, original = read_page(before, index)
            text, recognized = read_page(after, index)
            assert ImageChops.difference(original, recognized).getbbox() is None, index
            contents.append(text)
    assert contents[0].count("NATIVE HEADING") == 1 and "VECTOR SAMPLE" in contents[0], contents
    assert "SCANNED SAMPLE" in contents[1] and contents[2] == "" and "VECTOR SAMPLE" in contents[3], contents
    region = folder / "region.pdf"
    dispatch({**request, "output": str(region), "view": {"pages": [1]},
              "ocr_region": [[20, 185], [150, 185], [150, 245], [20, 245]]})
    with pdfium.PdfDocument(region) as selected, pdfium.PdfDocument(source) as original:
        text, bitmap = read_page(selected, 0)
        assert "VECTOR" in text and "SAMPLE" not in text and "NATIVE HEADING" in text, text
        assert ImageChops.difference(bitmap, read_page(original, 0)[1]).getbbox() is None
    retry = folder / "repeated.pdf"
    try:
        dispatch({**request, "input": str(output), "output": str(retry)})
    except UserError as exc:
        assert "žádný nový text" in str(exc), str(exc)
    else:
        raise AssertionError("OCR copied text from an existing layer")
    assert not retry.exists()

    state = source.stat()
    save = dict(action="save_document", input=str(output), output=str(source), overwrite=True,
                replace_source=True, replace_document=str(source),
                source_size=state.st_size, source_mtime_ns=state.st_mtime_ns)
    try:
        dispatch({**save, "source_mtime_ns": state.st_mtime_ns - 1})
    except UserError as exc:
        assert "se mezitím změnilo" in str(exc)
    else:
        raise AssertionError("Saving an OCR draft overwrote a changed original")
    assert source.read_bytes() == original_bytes
    dispatch(save)
    assert source.read_bytes() == output.read_bytes()


if __name__ == "__main__":
    with TemporaryDirectory(prefix="blacksmith-check-ocr-") as folder:
        check_ocr(Path(folder), str(Path(sys.argv[1]).resolve()), str(Path(sys.argv[2]).resolve()))
    print("OCR preserves appearance and existing text on scan, vector and rotated pages.")
