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


def block_fixture(path):
    with pdf.Pdf.new() as document:
        page = document.add_blank_page(page_size=(420, 340))
        font = pdf.Dictionary(Type=pdf.Name.Font, Subtype=pdf.Name.Type1, BaseFont=pdf.Name.Helvetica)
        page.obj.Resources = pdf.Dictionary(Font=pdf.Dictionary(F=font))
        page.obj.Contents = document.make_stream(
            b"BT /F 16 Tf 30 303 Td (Title) Tj ET\n"
            b"BT /F 13 Tf 30 265 Td (First) Tj 31 0 Td (line) Tj 25 0 Td (continues) Tj ET\n"
            b"BT /F 13 Tf 30 247 Td (on) Tj 17 0 Td (this) Tj 24 0 Td (line.) Tj ET\n"
            b"BT /F 13 Tf 275 265 Td (Other column) Tj 0 -18 Td (on both rows.) Tj ET\n"
            b"BT /F 13 Tf 30 213 Td (Separate paragraph.) Tj ET\n"
            b"BT /F 13 Tf 30 163 Td 0.30 Tc 1.1 Tw 81 Tz [(Left editable) -15 ( words)] TJ ( Right) Tj ET\n"
            b"q 0 1 1 0 k 250 60 80 43 re f Q\n")
        document.save(path)


def check_block(folder, fonts):
    original = folder / "source-block.pdf"
    block_fixture(original)
    probe = dict(action="text_probe", input=str(original))
    before = dispatch(probe)["text_page"]
    run = find_text(before, "Left editable words")
    outside = find_text(before, "Right")
    phrase = "".join(item["text"] for item in run["glyphs"])
    start = phrase.index("editable")
    end = start + len("editable words")
    assert run["token"] == 0 and run["editable"] and run["block_token"] > 0
    original_bytes = original.read_bytes()
    selection = [{"token": run["block_token"], "glyphs": list(range(start, end))}]
    output = folder / "laid-out.pdf"
    content = "Příliš žluťoučký kůň přeběhl silnici a vrátil se do stáje.\n\nDruhý odstavec."
    description = {"text": content, "align": "justify", "width": 146, "height": 150,
                   "leading": 1.5, "tracking": .35, "rotation": 0, "origin": [27, 150],
                   "selection": selection}
    edit = dict(font="Liberation Serif", size=12, bold=False, italic=False, underline=True, block=description)
    request = dict(action="text_edit", input=str(original), output=str(output), text_fonts=str(fonts), text_edit=edit)
    dispatch(request)
    assert original.read_bytes() == original_bytes
    after = dispatch({**probe, "input": str(output)})["text_page"]
    with pdfium.PdfDocument(original) as a, pdfium.PdfDocument(output) as b:
        source_text, source_image = read_page(a, 0)
        result_text, result_image = read_page(b, 0)
        assert "editable words" in source_text and "editable words" not in result_text and "Druhý odstavec." in result_text
        assert "Příliš žluťoučký kůň" in result_text and "Left" in result_text and "Right" in result_text
        assert ImageChops.difference(source_image, result_image).crop((235, 0, 420, 340)).getbbox() is None
    assert all(abs(a-b)<.001 for a,b in zip(find_text(after, "Right")["bounds"], outside["bounds"]))
    inserted = [run for run in after["objects"] if run["underline"]]
    assert len(inserted) >= 4 and all(run["editable"] for run in inserted)
    assert len({run["block"]["id"] for run in inserted}) == 1
    assert inserted[-1]["block"]["paragraph"] == 2
    assert inserted[0]["block"]["width"] == 146 and inserted[0]["block"]["alignment"] == 3
    assert inserted[0]["matrix"][5] - inserted[1]["matrix"][5] == 18
    with pdf.Pdf.open(output) as result:
        page = result.pages[0]
        assert not result.check_pdf_syntax()
        assert any(str(op) == "Tc" and abs(float(args[0])-.35)<.0001 for args,op in pdf.parse_content_stream(page))
        assert any(str(op) == "TJ" and len(args[0]) >= 5 for args,op in pdf.parse_content_stream(page))

    rotated = folder / "centered.pdf"
    spec = {**description, "text": "Zarovnání", "align": "center", "width": 150,
            "rotation": 90, "origin": [230,90]}
    dispatch({**request, "output": str(rotated), "text_edit": {**edit, "block": spec}})
    written = find_text(dispatch({**probe, "input": str(rotated)})["text_page"], "Zarovnání")
    assert tuple(written["matrix"][:4]) == (0., 1., -1., 0.)
    assert written["matrix"][5] > 130

    removed = folder / "removed.pdf"
    entire = [{"token": run["block_token"], "glyphs": list(range(len(run["glyphs"])))}]
    dispatch({**request, "output": str(removed), "text_edit": {**edit, "block": {
        **description, "text": "", "selection": entire}}})
    remaining = dispatch({**probe, "input": str(removed)})["text_page"]
    assert all(abs(a-b)<.001 for a,b in zip(find_text(remaining, "Right")["bounds"], outside["bounds"]))


def check_font_fallback(folder, fonts):
    from fontTools.ttLib import TTFont

    fallback = folder / "fonts"
    fallback.mkdir()
    for file in fonts.glob("*.ttf"):
        (fallback / file.name).symlink_to(file)
    target = fallback / "LiberationMono-BoldItalic.ttf"
    target.unlink()
    # Simulate a family without a bold or italic variant and without one Czech glyph.
    with TTFont(fonts / "LiberationMono-Regular.ttf") as font:
        for table in font["cmap"].tables:
            table.cmap.pop(ord("ž"), None)
        font.save(target)
    original=folder / "mixed-source.pdf"
    output=folder / "mixed-fonts.pdf"
    block_fixture(original)
    probe=dict(action="text_probe",input=str(original))
    run=find_text(dispatch(probe)["text_page"],"Title")
    edit=dict(font="Liberation Mono",size=13,bold=True,italic=True,block=dict(
        text="AžB",align="left",width=100,height=80,rotation=0,leading=1.2,tracking=0,origin=[30,320],
        selection=[dict(token=run["block_token"],glyphs=list(range(len(run["glyphs"]))))]))
    dispatch(dict(action="text_edit",input=str(original),output=str(output),text_fonts=str(fallback),text_edit=edit))
    content=dispatch({**probe,"input":str(output)})["text_page"]["objects"]
    written=[run for run in content if run["block"]]
    assert "".join(run["text"] for run in written)=="AžB"
    assert len({run["font"] for run in written})==2
    assert all(run["editable"] and run["bold"] and run["italic"] for run in written)
    assert written[0]["matrix"][2]>.2
    with pdf.Pdf.open(output) as document:
        assert not document.check_pdf_syntax()
        assert any(str(op)=="Tr" and int(args[0])==2 for args,op in pdf.parse_content_stream(document.pages[0]))


def catalog_fixture(path, fonts):
    from block_fonts import block_face, embed_face
    from text_edit import instruction

    cells = ("002610699", "Armatura pod umyvadlo", "bez zdroje, chrom", "17030")
    with pdf.Pdf.new() as document:
        page = document.add_blank_page(page_size=(420, 340))
        face = block_face("Liberation Sans", {}, fonts)
        chars = sorted(set("".join(cells)))
        name, cids = embed_face(document, page, face, chars)
        # PDF glyph codes and advances are independent of Unicode and the
        # font program's nominal advances. Keep the next cells dependent.
        advances = {char: face.widths[char] + 110 for char in chars}
        page.Resources.Font[name].DescendantFonts[0].W = pdf.Array([
            1, pdf.Array(advances[char] for char in chars),
        ])
        x = (20, 80, 233, 360)
        positioned = []
        for i, cell in enumerate(cells):
            positioned.append(pdf.String(b"".join(cids[char] for char in cell)))
            if i+1 < len(cells):
                positioned.append(sum(advances[char] for char in cell) - (x[i+1]-x[i])*1000/10)
        page.obj.Contents = document.make_stream(pdf.unparse_content_stream([
            instruction("g", .9), instruction("re", 0, 181, 420, 17), instruction("f"),
            instruction("g", .77), instruction("re", 265, 181, 4, 17), instruction("f"),
            instruction("g", 0), instruction("BT"), instruction("Tf", name, 10),
            instruction("Td", x[0], 186), instruction("TJ", pdf.Array(positioned)),
            instruction("TL", 20), instruction("'", pdf.String(b"".join(cids[char] for char in cells[0]))),
            instruction("ET"),
        ]))
        document.save(path)


def check_catalog(folder, fonts):
    source, target = folder / "catalog.pdf", folder / "catalog-edited.pdf"
    catalog_fixture(source, fonts)
    probe = {"action": "text_probe", "input": str(source)}
    objects = dispatch(probe)["text_page"]["objects"]
    assert len(objects) == 2 and all(run["editable"] and not run["whole_run_only"] for run in objects), objects
    row = objects[0]
    phrase = "bez zdroje, chrom"
    text = "".join(glyph["text"] for glyph in row["glyphs"])
    start = text.index(phrase)
    selection = [{"token": row["block_token"], "glyphs": list(range(start, start+len(phrase)))}]
    patch = dispatch({**probe, "action": "text_background", "text_selection": selection,
                      "view": {"region": {"zoom": 2, "x": 462, "y": 286, "width": 215, "height": 33}}})
    pixels = bytes(patch["raster"])
    # The gray stripe beneath the edited cell is part of the PDF artwork.
    assert pixels[(2*215+70)*4] < pixels[(2*215+20)*4]
    saved = dict(font="Liberation Sans", size=10, bold=True, block=dict(
        text="Baterie s termostatem", align="left", width=125, height=40,
        rotation=0, leading=1.2, tracking=0, origin=[233, 195], selection=selection))
    dispatch({**probe, "action": "text_edit", "output": str(target), "text_edit": saved, "text_fonts": str(fonts)})
    updated = dispatch({**probe, "input": str(target)})["text_page"]["objects"]
    before_x = next(glyph["origin"] for i, glyph in enumerate(row["glyphs"]) if text[i:] == "17030")
    after_x = next(glyph["origin"] for run in updated for i, glyph in enumerate(run["glyphs"])
                   if "".join(g["text"] for g in run["glyphs"][i:]) == "17030")
    assert all(abs(old-new) < .001 for old, new in zip(before_x, after_x))
    assert any(run["text"] == "Baterie s termostatem" and run["bold"] for run in updated)



def check_rich_text_in_shared_form(folder, fonts):
    source, target = folder / "shared-text.pdf", folder / "shared-edited.pdf"
    with pdf.Pdf.new() as document:
        page = document.add_blank_page(page_size=(500, 400))
        font = document.make_indirect(pdf.Dictionary(Type=pdf.Name.Font, Subtype=pdf.Name.Type1,
                                                    BaseFont=pdf.Name.Helvetica))
        inner = document.make_stream(
            b"BT /F 12 Tf 1 0 0 1 10 38 Tm (Original heading) Tj 1 0 0 1 20 18 Tm (Second line) Tj ET",
            Type=pdf.Name.XObject, Subtype=pdf.Name.Form, BBox=[0, 0, 160, 62],
            Resources=pdf.Dictionary(Font=pdf.Dictionary(F=font)))
        outer = document.make_stream(b"/Inner Do", Type=pdf.Name.XObject, Subtype=pdf.Name.Form,
                                     BBox=[0, 0, 160, 62], Resources=pdf.Dictionary(XObject=pdf.Dictionary(Inner=inner)))
        page.obj.Resources = pdf.Dictionary(XObject=pdf.Dictionary(Outer=outer))
        page.Contents = document.make_stream(b"q .8 g 40 200 160 62 re f Q "
                                            b"q 1 0 0 1 40 200 cm /Outer Do Q "
                                            b"q 1 0 0 1 220 200 cm /Outer Do Q")
        document.save(source)

    probe = dict(action="text_probe", input=str(source))
    initial = dispatch(probe)["text_page"]["objects"]
    assert len({item["block_token"] for item in initial if item["editable"]}) == 4, initial
    group = [item for item in initial if item["bounds"][0] < 190]
    assert [item["text"] for item in group] == ["Original heading", "Second line"]
    selection = [dict(token=item["block_token"], glyphs=list(range(len(item["glyphs"])))) for item in group]
    content = "New heading\nSecond line"
    regular = dict(font="Liberation Sans", size=12, tracking=0, color=0, bold=False, italic=False, underline=False)
    bold = {**regular, "bold": True, "size": 14, "underline": True, "tracking": .25}
    runs = [dict(regular, start=0, end=4), dict(bold, start=4, end=11),
            dict(regular, start=11, end=len(content))]
    block = dict(text=content, runs=runs, selection=selection, width=160, height=245, leading=1.2, tracking=0,
                 align="left", rotation=0, origin=[50, 247], baseline_offset=9, original_lines=[[0, 9], [10, 29]])
    dispatch({**probe, "action": "text_edit", "output": str(target), "text_fonts": str(fonts),
              "text_edit": dict(font="Liberation Sans", size=12, block=block)})

    changed = dispatch({**probe, "input": str(target)})["text_page"]
    heading, first = (find_text(changed, phrase) for phrase in ("heading", "New"))
    assert heading["bold"] and heading["underline"] and heading["size"] == 14
    assert heading["block"]["tracking"] == .25
    assert not first["bold"] and tuple(first["glyphs"][0]["origin"]) == (50, 238)
    second = next(run for run in changed["objects"] if run["text"] == "Second line" and run["bounds"][0] < 190)
    assert tuple(second["glyphs"][0]["origin"]) == (60, 218)
    assert find_text(changed, "Original heading")["bounds"][0] > 210
    assert all(run["block"] and run["editable"] for run in (first, second, heading))



def check_embedded_subset(folder, fonts):
    from io import BytesIO
    from fontTools.ttLib import TTFont

    source, target = folder / "subset-catalog.pdf", folder / "subset-edited.pdf"
    catalog_fixture(source, fonts)
    with pdf.Pdf.open(source, allow_overwriting_input=True) as document:
        page = document.pages[0]
        font = next(iter(page.Resources.Font.items()))[1]
        file = font.DescendantFonts[0].FontDescriptor.FontFile2
        with TTFont(BytesIO(file.read_bytes())) as face:
            for table in ("OS/2", "name", "post"):
                del face[table]
            data = BytesIO()
            face.save(data)
        file.write(data.getvalue())
        document.save(source)

    probe = dict(action="text_probe", input=str(source))
    description = dispatch(probe)["text_page"]
    row = next(run for run in description["objects"] if "002610699" in run["text"] and len(run["glyphs"]) > 30)
    assert row["face"] in description["fonts"], description["fonts"]
    before = "".join(glyph["text"] for glyph in row["glyphs"])
    cell = before.index("bez zdroje, chrom")
    value = "bez zdroje"
    run = dict(font=row["face"], color=0, size=10, tracking=0, start=0, end=len(value))
    block = dict(text=value, runs=[run], width=125, height=80, align="left", rotation=0, origin=[233, 195],
                 baseline_offset=9, leading=1.2, tracking=0,
                 selection=[dict(token=row["block_token"], glyphs=list(range(cell, cell+len("bez zdroje, chrom"))))])
    dispatch({**probe, "action": "text_edit", "output": str(target), "text_fonts": str(fonts),
              "text_edit": dict(size=10, block=block)})
    changed = dispatch({**probe, "input": str(target)})["text_page"]
    saved = find_text(changed, value)
    assert saved["font"] == "LiberationSans" and saved["face"] in changed["fonts"]
    after = next(glyph["origin"] for run in changed["objects"] for index, glyph in enumerate(run["glyphs"])
                 if "".join(g["text"] for g in run["glyphs"][index:]) == "17030")
    assert all(abs(x-y) < .001 for x, y in zip(after, row["glyphs"][before.index("17030")]["origin"]))


if __name__ == "__main__":
    with TemporaryDirectory(prefix="blacksmith-check-text-") as tmp:
        fonts = Path(sys.argv[1]).resolve()
        check_text(Path(tmp), fonts)
        check_block(Path(tmp), fonts)
        check_font_fallback(Path(tmp), fonts)
        check_catalog(Path(tmp), fonts)
        check_rich_text_in_shared_form(Path(tmp), fonts)
        check_embedded_subset(Path(tmp), fonts)
    print("Text blocks preserve nearby glyphs and graphics, and reflow with spacing, alignment and rotation.")
