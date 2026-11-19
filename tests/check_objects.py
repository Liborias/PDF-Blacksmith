"""Check object editing against transformed paths, images and shared forms."""

from base64 import b64decode
from io import BytesIO
from pathlib import Path
from tempfile import TemporaryDirectory

from pikepdf import Array, Dictionary, Name, OutlineItem, Pdf, parse_content_stream

from engine import UserError, dispatch


def fixture(path):
    with Pdf.new() as document:
        page = document.add_blank_page(page_size=(420, 340))
        second = document.add_blank_page(page_size=(420, 340))
        form = document.make_stream(b".72 .18 .20 rg 0 0 44 40 re f", Type=Name.XObject,
                                    Subtype=Name.Form, BBox=Array([0, 0, 44, 40]), Resources=Dictionary())
        image = document.make_stream(bytes([20, 60, 210]*4), Type=Name.XObject, Subtype=Name.Image, Width=2,
                                     Height=2, BitsPerComponent=8, ColorSpace=Name.DeviceRGB)
        font = document.make_indirect(Dictionary(Subtype=Name.Type1, Type=Name.Font, BaseFont=Name.Helvetica))
        page.obj.Resources = Dictionary(XObject=Dictionary(Group=form, Picture=image), Font=Dictionary(F=font))
        page.Contents = document.make_stream(
            b"0 .78 1 0 k 40 200 70 60 re f\n"
            b"q 2 0 0 1 220 80 cm 0 1 0 rg 0 0 33 38 re f Q\n"
            b"q 65 0 0 52 150 210 cm /Picture Do Q\n"
            b"q 1 0 0 1 300 210 cm /Group Do Q\n"
            b"BT /F 12 Tf 40 305 Td (Spare text) Tj ET\n"
        )
        second.obj.Resources = Dictionary(XObject=Dictionary(Group=form))
        second.Contents = document.make_stream(b"q 1 0 0 1 332 70 cm /Group Do Q\n")
        with document.open_outline() as outline:
            outline.root.append(OutlineItem("Second sheet", 1))
        document.save(path)


def probe(path, page=1):
    return dispatch({"action": "object_probe", "input": str(path), "view": {"page": page}})["object_page"]["objects"]


def check_objects(work):
    source = work / "original.pdf"
    fixture(source)
    original = source.read_bytes()
    scene = probe(source)
    assert [obj["kind"] for obj in scene] == ["path", "path", "image", "form"], scene
    red, green, image, group = scene

    def edit(source, output, operation, *, page=1, rotation=0, **fields):
        return dispatch({"action": "object_edit", "input": str(source), "output": str(output),
                         "view": {"page": page, "rotation": rotation},
                         "object_edit": {"operation": operation, **fields}})

    moved = work / "moved.pdf"
    edit(source, moved, "move", token=green["token"], delta=[24, -20])
    updated = probe(moved)
    assert updated[1]["bounds"] == (244, 60, 310, 98)
    for i in (0, 2, 3):
        assert updated[i]["bounds"] == scene[i]["bounds"]
    with Pdf.open(moved) as document:
        assert b"(Spare text) Tj" in document.pages[0].Contents.read_bytes()
        assert len(document.open_outline().root) == 1

    copied = work / "duplicated.pdf"
    edit(source, copied, "duplicate", token=group["token"], delta=[-22, 17])
    assert probe(copied)[-1]["bounds"] == (278, 227, 322, 267)
    assert probe(copied, 2) == probe(source, 2)

    removed = work / "removed.pdf"
    edit(source, removed, "delete", token=image["token"])
    assert [item["kind"] for item in probe(removed)] == ["path", "path", "form"]
    assert probe(removed, 2) == probe(source, 2)

    snapshot = dispatch({"action": "object_snapshot", "input": str(source), "view": {"page": 1},
                         "object_edit": {"token": red["token"]}})
    assert b64decode(snapshot["image"]).startswith(b"\x89PNG")
    with Pdf.open(BytesIO(b64decode(snapshot["object_clipboard"]))) as document:
        ops = parse_content_stream(document.pages[0])
        assert any(str(op) == "k" and list(map(float, args)) == [0, .78, 1, 0] for args, op in ops)
        assert sum(str(op) == "f" for _, op in ops) == 1

    pasted = work / "pasted.pdf"
    edit(source, pasted, "paste", page=2, rotation=90, clipboard=snapshot["object_clipboard"], at=[.5, .4])
    new = probe(pasted, 2)[-1]
    assert new["kind"] == "form" and new["bounds"] == (138, 135, 198, 205), new
    with Pdf.open(pasted) as document:
        assert not document.check_pdf_syntax()
        resources = document.pages[1].resources.XObject
        form = next(value for key, value in resources.items() if str(key).startswith("/PBObject"))
        assert form.Subtype == Name.Form
        assert any(str(op) == "k" and list(map(float, args)) == [0, .78, 1, 0]
                   for args, op in parse_content_stream(form))
    assert source.read_bytes() == original

    tokens = [red["token"], image["token"]]
    group_snapshot = dispatch({"action": "object_snapshot", "input": str(source),
                               "view": {"page": 1}, "object_edit": {"tokens": tokens}})
    with Pdf.open(BytesIO(b64decode(group_snapshot["object_clipboard"]))) as document:
        selected = list(parse_content_stream(document.pages[0]))
        assert sum(str(op) == "f" for _, op in selected) == 1
        assert sum(str(op) == "Do" for _, op in selected) == 1
    moved_group = work / "moved-group.pdf"
    edit(source, moved_group, "move", tokens=tokens, delta=[20, -10])
    assert [tuple(round(value, 3) for value in obj["bounds"]) for obj in probe(moved_group)] == [
        (60, 190, 130, 250), green["bounds"], (170, 200, 235, 252), group["bounds"],
    ]

    invalid = work / "invalid.pdf"
    try:
        edit(source, invalid, "delete", token=999)
    except UserError:
        assert not invalid.exists()
    else:
        raise AssertionError("An invalid object token modified the PDF.")


if __name__ == "__main__":
    with TemporaryDirectory(prefix="blacksmith-objects-") as directory:
        check_objects(Path(directory))
    print("Vector and image object edits, graphics state, rotated paste and bookmarks passed.")
