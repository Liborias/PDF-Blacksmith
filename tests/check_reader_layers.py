"""Check transparent drag passes, native PDF moves and durable page-local edits."""

from pathlib import Path
from tempfile import TemporaryDirectory

from PIL import Image
from pikepdf import Array, Dictionary, Name, Pdf, String, parse_content_stream

from engine import UserError, dispatch, open_pdf
from reader_render import ScreenReader


def fixture(path):
    with Pdf.new() as doc:
        page = doc.add_blank_page(page_size=(600, 440))
        second = doc.add_blank_page(page_size=(420, 360))
        normal = doc.make_stream(b".7 .1 .9 rg 0 0 60 26 re f", Type=Name.XObject, Subtype=Name.Form,
                                 BBox=Array([0, 0, 60, 26]))
        font = doc.make_indirect(Dictionary(Type=Name.Font, Subtype=Name.Type1, BaseFont=Name.Helvetica))
        field = doc.make_indirect(Dictionary(Type=Name.Annot, Subtype=Name.Widget, FT=Name.Tx,
            T=String("Native field"), V=String("Keep this value"), Rect=Array([230, 286, 290, 312]), F=4,
            AP=Dictionary(N=normal), P=page.obj))
        locked = doc.make_indirect(Dictionary(Type=Name.Annot, Subtype=Name.Widget, FT=Name.Tx,
            T=String("Fixed field"), Rect=Array([330, 286, 390, 312]), F=132, AP=Dictionary(N=normal), P=page.obj))
        page.Annots = Array([field, locked])
        doc.Root.AcroForm = Dictionary(Fields=Array([field, locked]), DA=String("/F 10 Tf 0 g"),
                                       DR=Dictionary(Font=Dictionary(F=font)))
        blend = doc.make_indirect(Dictionary(Type=Name.ExtGState, ca=.5))
        page.Resources = Dictionary(Font=Dictionary(F=font), ExtGState=Dictionary(Translucent=blend))
        page.Contents = doc.make_stream(
            b".9 .8 .5 rg 0 0 600 440 re f "
            b"q /Translucent gs 1 0 0 rg 36 300 64 28 re f Q "
            b"q 110 300 55 28 re W n 0 .7 0 rg 95 292 88 42 re f Q "
            b"0 0 1 rg 176 300 25 25 re f "
            b"BT /F 36 Tf .1 .15 .2 rg 40 205 Td (ABC) Tj ET "
        )
        second.Contents = doc.make_stream(b"0 1 1 rg 0 0 141 250 re f")
        doc.save(path)


def pixels(frame, key="raster"):
    return Image.frombytes("RGBA", (frame["raster_width"], frame["raster_height"]),
                           bytes(frame[key]), "raw", "BGRA")


def screen(source, page=1, **view):
    return dispatch({"input": str(source), "action": "screen", "view": {"box": "crop", "page": page, **view}},
                    raster=True)


def check_layers(folder):
    src = folder / "original.pdf"
    fixture(src)
    original = src.read_bytes()
    api = {"input": str(src)}
    objects = dispatch({**api, "action": "object_probe"})["object_page"]["objects"]
    assert [item["kind"] for item in objects] == ["path", "path", "path", "path", "widget"]
    backdrop, red, green, blue, field = objects
    text = dispatch({**api, "action": "text_probe"})["text_page"]["objects"]
    chosen = [{"token": text[0]["block_token"], "glyphs": [1]}]
    req = {**api, "action": "object_drag_preview",
           "view": {"page": 1, "box": "crop", "drag_bounds": [0, 0, .74, .72], "drag_zoom": 1},
           "object_edit": {"tokens": [red["token"], green["token"], field["token"]]}, "text_selection": chosen}
    preview = dispatch(req, raster=True)
    fg, bg = pixels(preview, "drag_raster"), pixels(preview)
    assert fg.getpixel((17, 40))[3] == 0
    assert fg.getpixel((185, 130))[3] == 0, "Unselected blue vector appeared in drag layer"
    assert fg.getpixel((50, 125))[0] > 245 and 110 < fg.getpixel((50, 125))[3] < 145
    assert fg.getpixel((125, 126))[:3] == (0, 179, 0)
    assert fg.getpixel((245, 145))[3] > 230, "Form appearance missing"
    assert bg.getpixel((185, 130))[:3] == (0, 0, 255)
    for point in [(50, 125), (125, 126), (245, 145)]:
        assert bg.getpixel(point)[:3] == bg.getpixel((17, 40))[:3], "Selected paint remains at the origin"
    glyphs = text[0]["glyphs"]
    box = lambda g: tuple(int(x) for x in (g[0], 440-g[3], g[2]+1, 440-g[1]+1))
    for i, glyph in enumerate(glyphs):
        visible = fg.crop(box(glyph["bounds"])).getchannel("A").getbbox()
        assert bool(visible) == (i==1), (i, glyph, visible)

    out = folder / "moved.pdf"
    result = dispatch({**api, "action": "object_edit", "output": str(out), "page_local": True,
        "view": {"page": 1}, "object_edit": {**req["object_edit"], "operation": "move_selection", "delta": [38, -45]},
        "text_selection": chosen})
    assert result["pages"] == 2 and result["changed_pages"] == [1] and "sizes" not in result
    assert bytes(screen(src, 2)["raster"]) == bytes(screen(out, 2)["raster"])
    new = pixels(screen(out, pixels=600))
    assert new.getpixel((163, 171))[:3] == (0, 179, 0), "Moved clip should travel with its paint"
    assert new.getpixel((125, 126))[:3] == bg.getpixel((17, 40))[:3]
    assert bytes(src.read_bytes()) == original
    with Pdf.open(out) as edited:
        assert not edited.check_pdf_syntax()
        moved, fixed = edited.pages[0].Annots
        assert tuple(moved.Rect) == (268, 241, 328, 267)
        assert tuple(fixed.Rect) == (330, 286, 390, 312)
        assert str(moved.V) == "Keep this value"
        assert moved.AP.N.read_bytes().endswith(b"re f")
        assert edited.Root.AcroForm.Fields[0].objgen == moved.objgen
        ops = {str(op.operator) for op in parse_content_stream(edited.pages[0])}
        assert {"re", "W", "TJ"} <= ops, ops
        assert not edited.pages[0].resources.get("/XObject"), "Movement rasterized PDF artwork"

    invalid = folder / "locked-move.pdf"
    try:
        dispatch({**api, "action": "object_edit", "output": str(invalid),
                  "object_edit": {"operation": "move", "tokens": [-2], "delta": [10, 10]}})
    except UserError:
        assert not invalid.exists()
    else:
        raise AssertionError("The locked form field moved")

    reader = ScreenReader()
    try:
        first = reader.document(src, "", open_pdf)
        assert reader.document(src, "", open_pdf) is first
        requests = [(1, "crop", 90), (1, "media", 0), (2, "crop", 270), (1, "crop", 0)]
        for number, box, rotation in requests:
            req = {**api, "action": "screen", "view": {"box": box, "page": number, "rotation": rotation}}
            cached = dispatch(req, raster=True, reader=reader)
            fresh = dispatch(req, raster=True)
            assert bytes(cached["raster"]) == bytes(fresh["raster"])
        assert reader.document(out, "", open_pdf) is not first, "Source switch reused a stale PDF"
    finally:
        reader.close()


if __name__ == "__main__":
    with TemporaryDirectory(prefix="pdfbs-layers-") as directory:
        check_layers(Path(directory))
    print("Transparent layers, editable mixed movement, form fields and reader reuse passed.")
