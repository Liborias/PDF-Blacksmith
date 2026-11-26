"""Check object editing against transformed paths, images and shared forms."""

from base64 import b64decode
from io import BytesIO
from pathlib import Path
from tempfile import TemporaryDirectory

from pikepdf import Array, Dictionary, Name, OutlineItem, Pdf, parse_content_stream
from PIL import Image, ImageCms
from pypdfium2 import PdfDocument

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


def check_image_crop(work):
    source = work / "crop-source.pdf"
    jpeg = BytesIO()
    Image.new("RGB", (20, 16), (30, 80, 220)).save(jpeg, format="JPEG")
    profile = ImageCms.ImageCmsProfile(ImageCms.createProfile("sRGB")).tobytes()
    with Pdf.new() as document:
        page = document.add_blank_page(page_size=(420, 340))
        icc = document.make_stream(profile, N=3)
        mask = document.make_stream(bytes([192])*320, Type=Name.XObject, Subtype=Name.Image,
                                    Width=20, Height=16, BitsPerComponent=8, ColorSpace=Name.DeviceGray)
        image = document.make_stream(
            jpeg.getvalue(), Type=Name.XObject, Subtype=Name.Image, Filter=Name.DCTDecode,
            Width=20, Height=16, BitsPerComponent=8, ColorSpace=Array([Name.ICCBased, icc]), SMask=mask)
        page.obj.Resources = Dictionary(XObject=Dictionary(Picture=image))
        page.Contents = document.make_stream(b"q 80 20 -15 54 100 135 cm /Picture Do Q\n"
                                             b"q 60 0 0 48 245 138 cm /Picture Do Q\n")
        document.save(source)
    before = source.read_bytes()
    first, second = probe(source)

    def edit(input, output, operation, token, **fields):
        dispatch(dict(action="object_edit", input=str(input), output=str(output), view={"page": 1},
                      object_edit=dict(operation=operation, token=token, **fields)))

    def all_bounds(path):
        return [tuple(round(value, 3) for value in obj["bounds"]) for obj in probe(path)]

    def render(path):
        with PdfDocument(path) as document:
            return document[0].render().to_pil().copy()

    cropped = work / "cropped.pdf"
    edit(source, cropped, "crop", first["token"], crop=[115, 150, 145, 177])
    assert [obj["kind"] for obj in probe(cropped)] == ["image", "image"]
    assert all_bounds(cropped) == [(115, 150, 145, 177), second["bounds"]]
    with Pdf.open(cropped) as document:
        assert not document.check_pdf_syntax()
        image = document.pages[0].resources.XObject.Picture
        assert (image.Width, image.Height, image.Filter) == (20, 16, Name.DCTDecode)
        assert image.read_raw_bytes() == jpeg.getvalue()
        assert image.ColorSpace[1].read_bytes() == profile
        assert image.SMask.read_bytes() == bytes([192])*320
    original, final = render(source), render(cropped)
    inside, outside, shared = (132, 173), (155, 185), (268, 171)
    assert final.getpixel(inside) == original.getpixel(inside)
    assert final.getpixel(outside) != original.getpixel(outside)
    assert final.getpixel(shared) == original.getpixel(shared)

    narrower = work / "cropped-again.pdf"
    edit(cropped, narrower, "crop", probe(cropped)[0]["token"], crop=[108, 160, 139, 184])
    assert all_bounds(narrower) == [(115, 160, 139, 177), second["bounds"]]
    moved = work / "moved-crop.pdf"
    edit(narrower, moved, "move", probe(narrower)[0]["token"], delta=[20, -5])
    assert all_bounds(moved) == [(135, 155, 159, 172), second["bounds"]]
    copied = work / "copied-crop.pdf"
    edit(moved, copied, "duplicate", probe(moved)[0]["token"], delta=[-40, -30])
    assert all_bounds(copied) == [(135, 155, 159, 172), (95, 125, 119, 142), second["bounds"]]
    snapshot = dispatch(dict(action="object_snapshot", input=str(cropped), view={"page": 1},
                             object_edit={"token": probe(cropped)[0]["token"]}))
    with Pdf.open(BytesIO(b64decode(snapshot["object_clipboard"]))) as clipboard:
        assert any(str(op) == "W" for _, op in parse_content_stream(clipboard.pages[0]))
        assert clipboard.pages[0].resources.XObject.Picture.read_raw_bytes() == jpeg.getvalue()

    for invalid in ([130, 150, 110, 170], [0, 0, 100, 100], [85, 135, 180, 209]):
        output = work / "invalid-crop.pdf"
        try:
            edit(source, output, "crop", first["token"], crop=invalid)
        except UserError:
            assert not output.exists()
        else:
            raise AssertionError(f"Invalid crop modified the PDF: {invalid}")
    assert source.read_bytes() == before


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
    edit(source, moved_group, "move", tokens=tokens, delta=[20, -10], grouping="separate")
    assert [tuple(round(value, 3) for value in obj["bounds"]) for obj in probe(moved_group)] == [
        (60, 190, 130, 250), green["bounds"], (170, 200, 235, 252), group["bounds"],
    ]

    grouped = work / "grouped-move.pdf"
    edit(source, grouped, "move", tokens=tokens, delta=[20, -10], grouping="group")
    moved = probe(grouped)
    assert [obj["kind"] for obj in moved] == ["path", "form", "form"]
    assert [obj["bounds"] for obj in moved[:2]] == [green["bounds"], group["bounds"]]
    assert all(abs(a-b) < .03 for a, b in zip(moved[-1]["bounds"], (60, 190, 235, 252))), moved[-1]
    assert probe(grouped, 2) == probe(source, 2)
    with Pdf.open(grouped) as document:
        page = document.pages[0]
        form = next(value for key, value in page.resources.XObject.items() if str(key).startswith("/PBObject"))
        contents = list(parse_content_stream(form))
        assert [str(op) for _, op in contents if str(op) in ("f", "Do")] == ["f", "Do"]
        assert any(str(op) == "k" and list(map(float, args)) == [0, .78, 1, 0] for args, op in contents)
        assert b"(Spare text) Tj" in page.Contents.read_bytes()
        assert not document.check_pdf_syntax()

    duplicated_group = work / "grouped-duplicate.pdf"
    edit(source, duplicated_group, "duplicate", tokens=tokens, delta=[12, -12], grouping="group")
    duplicated = probe(duplicated_group)
    assert [(obj["kind"], obj["bounds"]) for obj in duplicated[:-1]] == [
        (obj["kind"], obj["bounds"]) for obj in scene
    ]
    assert duplicated[-1]["kind"] == "form"

    pasted_group = work / "grouped-paste.pdf"
    pasted_separately = work / "separate-paste.pdf"
    clipboard = group_snapshot["object_clipboard"]
    edit(source, pasted_group, "paste", page=2, grouping="group", rotation=90, at=[.5, .4], clipboard=clipboard)
    edit(source, pasted_separately, "paste", page=2, grouping="separate", rotation=90, at=[.5, .4], clipboard=clipboard)
    one = probe(pasted_group, 2)
    separate = probe(pasted_separately, 2)
    assert len(one) == 2 and len(separate) == 3
    first, second = [item["bounds"] for item in separate[-2:]]
    combined = (min(first[0], second[0]), min(first[1], second[1]),
                max(first[2], second[2]), max(first[3], second[3]))
    assert all(abs(a-b) < .02 for a, b in zip(combined, one[-1]["bounds"]))
    with Pdf.open(pasted_separately) as document:
        assert not document.check_pdf_syntax()
        parts = [value for key, value in document.pages[1].resources.XObject.items() if str(key).startswith("/PBObject")]
        paints = {tuple(str(op) for _, op in parse_content_stream(part) if str(op) in ("f", "Do")) for part in parts}
        assert paints == {("f",), ("Do",)}, paints
    moved_part = work / "single-pasted-part-moved.pdf"
    edit(pasted_separately, moved_part, "move", page=2, token=separate[-1]["token"], delta=[0, 8])
    assert [obj["bounds"] for obj in probe(moved_part, 2)[:-1]] == [obj["bounds"] for obj in separate[:-1]]
    assert source.read_bytes() == original

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
        check_image_crop(Path(directory))
    print("Object edits, image clipping, graphics state, rotated paste and bookmarks passed.")
