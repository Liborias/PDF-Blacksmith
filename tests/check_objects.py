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
    def clip_operators(path):
        with Pdf.open(path) as document:
            assert not document.check_pdf_syntax()
            page = document.pages[0]
            image = page.resources.XObject.Picture
            assert (image.Width, image.Height, image.Filter) == (20, 16, Name.DCTDecode)
            assert image.read_raw_bytes() == jpeg.getvalue()
            assert image.ColorSpace[1].read_bytes() == profile
            assert image.SMask.read_bytes() == bytes([192])*320
            return [str(op) for _, op in parse_content_stream(page)]

    assert "W" in clip_operators(cropped)
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

    oval = work / "oval.pdf"
    edit(cropped, oval, "crop", probe(cropped)[0]["token"], crop=[115, 150, 145, 177], crop_shape="ellipse")
    assert clip_operators(oval).count("c") == 4
    assert all_bounds(oval) == [(115, 150, 145, 177), second["bounds"]]
    oval_render = render(oval)
    corner = (117, 188)
    assert final.getpixel(corner) == original.getpixel(corner)
    assert oval_render.getpixel(corner) == (255, 255, 255)
    assert oval_render.getpixel(inside) == original.getpixel(inside)
    assert oval_render.getpixel(shared) == original.getpixel(shared)

    partial = work / "partial-oval.pdf"
    edit(oval, partial, "crop", probe(oval)[0]["token"], crop=[115, 151, 133, 176])
    ops = clip_operators(partial)
    assert ops.count("c") == 4 and ops.count("W") == 3
    result = render(partial)
    assert result.getpixel(corner) == (255, 255, 255)
    assert result.getpixel(inside) == original.getpixel(inside)
    assert result.getpixel((139, 173)) == (255, 255, 255)

    circle = work / "circle.pdf"
    edit(source, circle, "crop", first["token"], crop=[118, 152, 142, 176], crop_shape="ellipse")
    assert clip_operators(circle).count("W") == 1
    assert all_bounds(circle) == [(118, 152, 142, 176), second["bounds"]]
    moved_circle = work / "moved-circle.pdf"
    edit(circle, moved_circle, "move", probe(circle)[0]["token"], delta=[8, 15])
    assert clip_operators(moved_circle).count("c") == 4
    assert all_bounds(moved_circle) == [(126, 167, 150, 191), second["bounds"]]
    assert render(moved_circle).getpixel((127, 172)) == (255, 255, 255)
    assert render(moved_circle).getpixel((138, 161)) == original.getpixel(inside)

    legacy_mask = probe(partial)[0]
    assert tuple(legacy_mask["image_mask"]["source_bounds"]) == first["bounds"]
    assert len(legacy_mask["image_mask"]["outline"]) > 4
    replaced = work / "replaced-oval.pdf"
    edit(partial, replaced, "replace_crop", legacy_mask["token"], crop=[114, 147, 151, 180], crop_shape="ellipse")
    assert all_bounds(replaced) == [(114, 147, 151, 180), second["bounds"]]
    assert clip_operators(replaced).count("W") == 1
    recovered = (139, 173)
    assert result.getpixel(recovered) == (255, 255, 255)
    assert render(replaced).getpixel(recovered) == original.getpixel(recovered)
    restored = work / "restored.pdf"
    edit(replaced, restored, "reset_crop", probe(replaced)[0]["token"])
    assert all_bounds(restored) == [first["bounds"], second["bounds"]]
    assert "image_mask" not in probe(restored)[0] and "W" not in clip_operators(restored)
    assert render(restored).tobytes() == original.tobytes()

    # A separately positioned image retains its placement and any enclosing
    # graphics clip when replacing or removing its own crop mask.
    positioned = work / "positioned-mask.pdf"
    expected = work / "positioned-without-mask.pdf"
    header = b"q 60 0 0 54 100 135 cm 0 0 .6 1 re W n "
    placement = b"q .8 0 0 .8 .1 .1 cm /Picture Do Q "
    with Pdf.open(source) as document:
        page = document.pages[0]
        page.Contents = document.make_stream(
            header + b"/PBImageCrop BMC q .25 .25 m .55 .25 l .55 .55 l .25 .55 l h W n "
            + placement[2:] + b"EMC Q")
        document.save(positioned)
        page.Contents = document.make_stream(header + placement + b"Q")
        document.save(expected)
    positioned_mask = probe(positioned)[0]["image_mask"]
    assert all_bounds(positioned) == [(115, 148.5, 133, 164.7)]
    assert tuple(round(value, 3) for value in positioned_mask["source_bounds"]) == (106, 140.4, 154, 183.6)
    expanded = work / "positioned-expanded.pdf"
    edit(positioned, expanded, "replace_crop", probe(positioned)[0]["token"], crop=[110, 146, 145, 171])
    assert tuple(probe(expanded)[0]["image_mask"]["source_bounds"]) == tuple(positioned_mask["source_bounds"])
    assert clip_operators(expanded).count("W") == 2
    restored_positioned = work / "positioned-reset.pdf"
    edit(expanded, restored_positioned, "reset_crop", probe(expanded)[0]["token"])
    assert clip_operators(restored_positioned).count("W") == 1
    assert render(restored_positioned).tobytes() == render(expected).tobytes()

    unchanged = work / "no-restore.pdf"
    try:
        edit(source, unchanged, "reset_crop", first["token"])
    except UserError:
        assert not unchanged.exists()
    else:
        raise AssertionError("Resetting a bare image unexpectedly edited the PDF.")

    for input_path, rect, shape in [(oval, [115, 150, 117, 152], "ellipse"),
                                    (oval, [115, 150, 145, 177], "ellipse"),
                                    (source, [115, 150, 145, 177], "unknown")]:
        output = work / "invalid-oval.pdf"
        try:
            edit(input_path, output, "crop", probe(input_path)[0]["token"], crop=rect, crop_shape=shape)
        except UserError:
            assert not output.exists()
        else:
            raise AssertionError(f"Invalid or unchanged {shape} crop modified the PDF: {rect}")

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
