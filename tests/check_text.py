"""Exercise native text editing without disturbing unrelated PDF content."""

from pathlib import Path
from tempfile import TemporaryDirectory
import sys

import pikepdf as pdf
import pypdfium2 as pdfium
from PIL import ImageChops

from check_ocr import read_page
from engine import UserError, dispatch


def fixture(path):
    with pdf.Pdf.new() as document:
        font = pdf.Dictionary(F1=document.make_indirect(pdf.Dictionary(
            Type=pdf.Name.Font, Subtype=pdf.Name.Type1, BaseFont=pdf.Name.Helvetica)))
        form = document.make_stream(b"BT /F1 12 Tf 10 25 Td (Nested) Tj ET", Type=pdf.Name.XObject,
                                    Subtype=pdf.Name.Form, BBox=pdf.Array([0, 0, 120, 100]),
                                    Resources=pdf.Dictionary(Font=font))
        for number in range(4):
            page = document.add_blank_page(page_size=(420, 340))
            page.obj.Resources = pdf.Dictionary(Font=font, XObject=pdf.Dictionary(Nested=form))
            page.obj.Contents = document.make_stream(
                f"q 1 0 0 1 30 300 cm 0 0.75 0.25 0.2 k BT /F1 18 Tf 0 0 Td (Header {number+1}) Tj ET Q\n"
                "BT /F1 14 Tf 35 238 Td (A line of editable text) Tj 0 -23 Td (Another line) Tj ET\n"
                "BT /F1 12 Tf 35 174 Td (Runs ) Tj (on one line) Tj ET\n"
                "q 1 0 0 1 40 91 cm /Nested Do Q\n"
                "q 0 1 1 0 k 220 50 77 65 re f Q\n"
                "BT /F1 16 Tf 315 140 Td 7 Tr (Clip) Tj 0 Tr 0 -30 Td (Clip dependency) Tj ET\n"
                "/Span << /ActualText (Tagged description) >> BDC BT /F1 10 Tf 235 80 Td (Tagged) Tj ET EMC\n"
                .encode("ascii")
            )
        with document.open_outline() as outline:
            outline.root.append(pdf.OutlineItem("Beginning", 0))
        document.save(path)


def find_text(description, value):
    return next(run for run in description["objects"] if run["text"] == value)


def check_text(folder, fonts):
    original = folder / "input.pdf"
    fixture(original)
    before = original.read_bytes()
    probe = dict(action="text_probe", input=str(original), view={"page": 1})
    description = dispatch(probe)["text_page"]
    assert find_text(description, "Nested")["token"] == 0
    assert find_text(description, "Runs")["token"] == 0
    assert find_text(description, "Clip dependency")["token"] == 0
    header = find_text(description, "Header 1")
    line = find_text(description, "A line of editable text")
    assert header["token"] > 0 and line["token"] > 0

    output = folder / "changed.pdf"
    edit = dict(family=1, size=20, bold=True, italic=True, underline=True,
                runs=[dict(token=header["token"], text="Příliš žluťoučký kůň"),
                      dict(token=line["token"], text="Změna řádku")])
    request = dict(action="text_edit", input=str(original), output=str(output),
                   text_fonts=str(fonts), text_edit=edit, view={"page": 1})
    dispatch(request)
    assert original.read_bytes() == before
    with pdfium.PdfDocument(original) as old, pdfium.PdfDocument(output) as changed:
        extracted, after_image = read_page(changed, 0)
        assert "Příliš žluťoučký kůň" in extracted and "Změna řádku" in extracted and "Header 1" not in extracted
        assert "Another line" in extracted and "Nested" in extracted and "Runs on one line" in extracted
        difference = ImageChops.difference(read_page(old, 0)[1], after_image).getbbox()
        assert difference and difference[3] < 125, difference
        for index in range(1, len(old)):
            assert ImageChops.difference(read_page(old, index)[1], read_page(changed, index)[1]).getbbox() is None

    revised = dispatch({**probe, "input": str(output)})["text_page"]
    updated = find_text(revised, "Příliš žluťoučký kůň")
    assert updated["token"] and updated["size"] == 20 and updated["underline"] and updated["bold"] and updated["italic"]
    assert find_text(revised, "Another line")["bounds"] == find_text(description, "Another line")["bounds"]
    stripped = folder / "revised.pdf"
    repeat = {**request, "input": str(output), "output": str(stripped),
              "view": {"page": 1, "order": [2, 1, 4], "rotations": [90, 0, 0, 0]}}
    repeat["text_edit"] = {**edit, "underline": False, "runs": [dict(token=updated["token"], text="Příliš žluťoučký kůň")]}
    result = dispatch(repeat)
    assert len(result["sizes"]) == 3 and result["bookmarks"][0]["page"] == 2
    revised = dispatch({**probe, "input": str(stripped), "view": {"page": 2}})["text_page"]
    assert revised["rotation"] == 90 and not find_text(revised, "Příliš žluťoučký kůň")["underline"]
    with pdf.Pdf.open(stripped) as saved:
        assert not saved.check_pdf_syntax()
        assert sum(1 for args, op in pdf.parse_content_stream(saved.pages[1])
                   if str(op) == "BMC" and args[0] == pdf.Name.PBUnderline) == 1

    rejected = folder / "rejected.pdf"
    try:
        dispatch({**request, "output": str(rejected), "text_edit": {**edit, "runs": [{"token": 0, "text": "Lost"}]}})
    except UserError:
        assert not rejected.exists()
    else:
        raise AssertionError("Read-only text was replaced")


if __name__ == "__main__":
    with TemporaryDirectory(prefix="blacksmith-check-text-") as tmp:
        check_text(Path(tmp), Path(sys.argv[1]).resolve())
    print("Text preserves other objects, page edits and outlines, including a second formatting pass.")
