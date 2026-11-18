"""Integration checks for print proofs, export settings and page editing."""

from base64 import b64decode
from io import BytesIO
import json
from pathlib import Path
import shutil
import subprocess
import sys
from tempfile import TemporaryDirectory

from PIL import Image, ImageChops
from pikepdf import Array, Dictionary, Encryption, Name, Pdf, Stream, parse_content_stream


ROOT = Path(__file__).resolve().parents[1]


def call(exe, pdf, action, expected="ok", **options):
    payload = dict(input=str(pdf), action=action, ghostscript=exe, **options)
    result = subprocess.run([sys.executable, str(ROOT / "engine.py")],
                            input=json.dumps(payload), text=True, capture_output=True, check=True)
    response = json.loads(result.stdout)
    assert response["status"] == expected, response
    return response


def image(proof):
    return Image.open(BytesIO(b64decode(proof["image"])))


def check_stream(exe, fixture, proof, folder):
    encrypted = folder / "šifrovaný náhled.pdf"
    with Pdf.open(fixture) as pdf:
        pdf.save(encrypted, encryption=Encryption(user="cteni", owner="vlastnik"))
    request = dict(input=str(encrypted), action="preview", ghostscript=exe, view={"page": 1})
    payload = "\n".join(map(json.dumps, [request, {**request, "pdf_password": "cteni"},
                                        {**request, "input": str(fixture), "action": "describe"}]))
    result = subprocess.run([sys.executable, str(ROOT / "engine.py"), "--stream"],
                            input=payload, text=True, capture_output=True, check=True, timeout=30)
    refused, opened, described = map(json.loads, result.stdout.splitlines())
    assert refused["status"] == "error" and "heslo" in refused["message"]
    assert opened["status"] == "ok", opened
    assert opened["view"]["plates"] == proof["plates"]
    assert image(opened["view"]).tobytes() == image(proof).tobytes()
    assert described["status"] == "ok" and described["view"]["pages"] == 11


def check_screen_detail(exe, folder):
    fixture = folder / "detail křivek.pdf"
    with Pdf.new() as pdf:
        page = pdf.add_blank_page(page_size=(260, 330))
        page.obj.CropBox = Array([28, 30, 172, 246])
        page.obj.Rotate = 90
        page.obj.UserUnit = 2
        font = Dictionary(Type=Name.Font, Subtype=Name.Type1, BaseFont=Name.Helvetica)
        page.obj.Resources = Dictionary(Font=Dictionary(F1=font))
        page.obj.Contents = pdf.make_stream(
            b"0.4 w 32 60 m 150 120 35 190 170 230 c S "
            b"BT /F1 9 Tf 52 145 Td (Vector PDF text) Tj ET "
            b"0.5 0.2 0.6 rg 50 80 28 90 re f")
        pdf.save(fixture)

    dimensions = call(exe, fixture, "describe")["view"]["sizes"][0]
    assert dimensions["crop"] == dimensions["trim"] == [432, 288]
    screen_views = []
    for zoom, rotation, box in ((2, 0, "crop"), (3, 90, "trim"), (7, 270, "crop")):
        # The full-page reference at the same scale models both the PDF's own
        # rotation and a pending rotation made in the reader.
        view = dict(page=1, box=box, rotation=rotation, pixels=432*zoom)
        full = image(call(exe, fixture, "screen", view=view)["view"])
        x, y = full.width//4, full.height//4
        w, h = full.width//2, full.height//2
        region = dict(zoom=zoom, x=x, y=y, width=w, height=h)
        detail = image(call(exe, fixture, "screen", view={**view, "region": region})["view"])
        expected = full.crop((x, y, x+w, y+h))
        assert detail.size == expected.size
        assert detail.tobytes() == expected.tobytes()
        assert detail.getextrema() != ((255, 255),) * 3
        screen_views.extend((view, {**view, "region": region}))

    # The desktop stream mixes packed screen frames and ordinary JSON replies.
    # Use an invalid Ghostscript path for screen requests to require PDFium.
    request = dict(input=str(fixture), action="screen", ghostscript=str(folder / "missing"))
    requests = [{**request, "view": view} for view in screen_views]
    requests.extend(({**request, "view": {"region": {"zoom": 100}}},
                     {**request, "action": "preview", "ghostscript": exe},
                     {**request, "action": "describe"}))
    lines = b"".join(json.dumps(req).encode("ascii") + b"\n" for req in requests)
    response = subprocess.run([sys.executable, str(ROOT / "engine.py"), "--reader-stream"],
                              input=lines, capture_output=True, check=True, timeout=30)
    replies = BytesIO(response.stdout)
    for index in range(0, len(screen_views), 2):
        rendered = []
        for _ in range(2):
            frame = json.loads(replies.readline())
            assert frame["status"] == "ok", frame
            view = frame["view"]
            size = (view["raster_width"], view["raster_height"])
            rendered.append(Image.frombytes("RGB", size, replies.read(size[0]*size[1]*4), "raw", "BGRX", 0, 1))
        full, detail = rendered
        assert max(full.size) == screen_views[index]["pixels"]
        r = screen_views[index+1]["region"]
        x, y, w, h = r["x"], r["y"], r["width"], r["height"]
        assert detail.size == (w, h)
        assert detail.getextrema() != ((255, 255),) * 3
        expected = full.crop((x, y, x+w, y+h))
        # Clipping antialiased paths can round individual channels by one level.
        assert max(hi for _, hi in ImageChops.difference(detail, expected).getextrema()) <= 1
    assert json.loads(replies.readline())["status"] == "error"
    proof = json.loads(replies.readline())
    assert proof["status"] == "ok" and proof["view"]["plates"] and proof["view"]["image"]
    assert json.loads(replies.readline())["view"]["sizes"][0] == dimensions
    assert replies.read() == b""


def check_plate_switching(exe, folder):
    fixture = folder / "hladké separace.pdf"
    with Pdf.new() as pdf:
        page = pdf.add_blank_page(page_size=(271, 331))
        page.obj.Contents = pdf.make_stream(b"0.63 0.49 0.38 0.26 k 0 0 271 331 re f")
        pdf.save(fixture)
    toggles = [[], ["Magenta"], ["Cyan", "Magenta", "Black"],
               ["Cyan", "Magenta", "Yellow", "Black"]]
    request = dict(input=str(fixture), action="preview", ghostscript=exe)
    lines = "\n".join(json.dumps({**request, "view": {"hide_plates": hidden}}) for hidden in toggles)
    result = subprocess.run([sys.executable, str(ROOT / "engine.py"), "--stream"], input=lines,
                            text=True, capture_output=True, check=True, timeout=30)
    responses = [json.loads(line) for line in result.stdout.splitlines()]
    assert len(responses) == len(toggles)
    colors = []
    for response in responses:
        assert response["status"] == "ok", response
        assert response["view"]["plates"] == responses[0]["view"]["plates"]
        bitmap = image(response["view"])
        # The entire interior is one overprinted color. Selecting plates must
        # neither introduce stripes nor alter the measured coverage of a plate.
        flat = bitmap.crop((8, 8, bitmap.width-8, bitmap.height-8)).getcolors(maxcolors=2)
        assert flat is not None and len(flat) == 1
        colors.append(flat[0][1])
    assert len(set(colors)) == len(toggles) and colors[-1] == (255, 255, 255)


def check_proofs(exe, folder):
    fixture = ROOT / "tests/fixtures/barvy.pdf"
    described = call(exe, fixture, "describe")["view"]
    assert len(described["sizes"]) == 11
    assert described["sizes"][0]["media"] == [220, 245]
    assert described["sizes"][10]["media"] == [245, 220]
    screen = call(exe, fixture, "screen", view={"page": 1, "box": "crop", "pixels": 780})["view"]
    assert image(screen).height > image(screen).width
    screen_rotated = call(exe, fixture, "screen", view={"page": 1, "box": "crop", "pixels": 780, "rotation": 90})["view"]
    assert image(screen_rotated).size == image(screen).size[::-1]
    printed = call(exe, fixture, "print_page", view={"page": 1, "box": "crop", "pixels": 3508, "rotation": 90})["view"]
    assert image(printed).width > image(printed).height
    proof = call(exe, fixture, "preview", view={"page": 1})["view"]
    assert proof["pages"] == 11
    assert {plate["name"] for plate in proof["plates"] if plate["active"]} == {
        "Cyan", "Magenta", "Yellow", "Black", "TestSpot",
    }
    check_stream(exe, fixture, proof, folder)

    single = call(exe, fixture, "preview", view={"page": 1, "plates": ["TestSpot"]})["view"]
    assert image(single).tobytes() != image(proof).tobytes()
    assert image(single).getextrema() != ((255, 255),) * 3
    blank = call(exe, fixture, "preview", view={"page": 1, "plates": []})["view"]
    assert image(blank).getextrema() == ((255, 255),) * 3
    rotated = call(exe, fixture, "preview", view={"page": 1, "rotation": 90})["view"]
    assert image(rotated).size == image(proof).size[::-1]
    trim = call(exe, fixture, "preview", view={"page": 1, "box": "trim"})["view"]
    assert image(trim).width < image(proof).width

    sep = folder / "barevné strany.pdf"
    with Pdf.new() as pdf:
        for value in (b"0 0 0 1", b"1 0 0 0", b"0 1 0 0"):
            page = pdf.add_blank_page(page_size=(180, 220))
            page.obj.Contents = pdf.make_stream(value + b" k 10 20 40 30 re f")
        target = pdf.pages[2].obj
        pdf.Root.Names = Dictionary(Dests=Dictionary(Names=Array(["control", Array([target, Name.Fit])])))
        pdf.Root.PageLabels = Dictionary(Nums=Array([0, Dictionary(P="S-", S=Name.D)]))
        pdf.save(sep)

    exported = call(exe, sep, "scan", export={"proof_dpi": 75})["view"]
    assert exported["dpi"] == 75
    assert {plate["name"]: plate["pages"] for plate in exported["used_plates"]} == {
        "Black": [1], "Cyan": [2], "Magenta": [3],
    }
    plain_copy = folder / "uloženo jako.pdf"
    call(exe, sep, "save_document", output=str(plain_copy))
    assert plain_copy.read_bytes() == sep.read_bytes()
    src_stat = plain_copy.stat()
    overwriting = dict(output=str(plain_copy), overwrite=True,
                       view={"order": [3, 1, 2], "rotations": [0, 0, 90]})
    call(exe, plain_copy, "save_document", expected="error", **overwriting)
    call(exe, plain_copy, "save_document", expected="error", replace_source=True,
         source_size=src_stat.st_size, source_mtime_ns=src_stat.st_mtime_ns-1, **overwriting)
    assert plain_copy.read_bytes() == sep.read_bytes()
    call(exe, plain_copy, "save_document", replace_source=True,
         source_size=src_stat.st_size, source_mtime_ns=src_stat.st_mtime_ns, **overwriting)
    with Pdf.open(plain_copy) as pdf:
        assert pdf.pages[0].obj.Rotate == 90

    edited = folder / "otočené a přesunuté.pdf"
    call(exe, sep, "edit_pages", output=str(edited),
         view={"order": [3, 1, 2], "rotations": [0, 0, 90]})
    with Pdf.open(edited) as pdf:
        assert pdf.pages[0].obj.Rotate == 90
        dest = pdf.Root.Names.Dests.Names[1][0]
        assert dest.objgen == pdf.pages[0].objgen
        assert pdf.pages[0].label == "S-3"
    draft_export = folder / "export se změnami.pdf"
    call(exe, sep, "transform", gray=True, output=str(draft_export),
         view={"order": [3, 1, 2], "rotations": [0, 0, 90]})
    with Pdf.open(draft_export) as pdf:
        assert [int(v) for v in pdf.pages[0].mediabox[2:]] == [220, 180]

    subset = folder / "vybrané strany.pdf"
    call(exe, sep, "transform", expected="error", output=str(subset), view={"pages": [1, 4]})
    assert not subset.exists()
    call(exe, sep, "transform", output=str(subset), view={"pages": [3, 1, 3]})
    with Pdf.open(subset) as pdf:
        assert [page.label for page in pdf.pages] == ["S-1", "S-3"]
        assert pdf.Root.Names.Dests.Names[1][0].objgen == pdf.pages[1].objgen
        assert pdf.pages[1].obj.Contents.read_bytes().startswith(b"0 1 0 0 k")

    subset_draft = folder / "výběr po změnách stran.pdf"
    call(exe, sep, "transform", gray=True, output=str(subset_draft),
         view={"pages": [2, 3], "order": [2, 3, 1], "rotations": [0, 0, 90]})
    with Pdf.open(subset_draft) as pdf:
        assert len(pdf.pages) == 2
        assert [list(page.mediabox[2:]) for page in pdf.pages] == [[220, 180], [180, 220]]

    encrypted = folder / "šifrovaný výběr.pdf"
    encrypted_subset = folder / "vybraná šifrovaná strana.pdf"
    with Pdf.open(sep) as pdf:
        pdf.save(encrypted, encryption=Encryption(user="cteni", owner="vlastnik"))
    call(exe, encrypted, "transform", pdf_password="cteni", output=str(encrypted_subset), view={"pages": [2]})
    with Pdf.open(encrypted_subset, password="cteni") as pdf:
        assert pdf.is_encrypted and len(pdf.pages) == 1 and pdf.pages[0].label == "S-2"

    src = folder / "obrázek.pdf"
    output = folder / "nastavení.pdf"
    with Pdf.new() as pdf:
        page = pdf.add_blank_page(page_size=(180, 230))
        picture = pdf.make_stream(Image.new("RGB", (600, 600), "navy").tobytes(),
                                  Type=Name.XObject, Subtype=Name.Image, Width=600, Height=600,
                                  ColorSpace=Name.DeviceRGB, BitsPerComponent=8)
        page.obj.Resources = Dictionary(XObject=Dictionary(Im=picture))
        page.obj.Contents = pdf.make_stream(b"q 100 0 0 100 25 25 cm /Im Do Q")
        pdf.save(src)
    result = call(exe, src, "transform", gray=True, output=str(output),
                  export={"image_dpi": 150, "compression": "flate", "proof_dpi": 75})
    assert "při 75 dpi" in result["message"]
    with Pdf.open(output) as pdf:
        images = [obj for obj in pdf.objects if isinstance(obj, Stream) and obj.get("/Subtype") == Name.Image]
        assert images and all(190 <= int(obj.Width) <= 220 and obj.Filter == Name.FlateDecode for obj in images)


def check_export_presets(exe, folder):
    source = folder / "předvolby.pdf"
    with Pdf.new() as pdf:
        noise = Image.effect_noise((600, 600), 50)
        bitmap = Image.merge("RGB", (noise, Image.new("L", noise.size, 150), noise))
        picture = pdf.make_stream(bitmap.tobytes(), Type=Name.XObject, Subtype=Name.Image,
                                  Width=600, Height=600, ColorSpace=Name.DeviceRGB, BitsPerComponent=8)
        font = Dictionary(Type=Name.Font, Subtype=Name.Type1, BaseFont=Name.Helvetica)
        for index in range(2):
            page = pdf.add_blank_page(page_size=(180, 270))
            page.obj.MediaBox = Array([-15, 20, 165, 290])
            page.obj.CropBox = Array([-10, 27, 135, 270])
            page.obj.BleedBox = Array([-5, 30, 130, 250])
            page.obj.TrimBox = Array([0, 35, 120, 240])
            page.obj.Resources = Dictionary(XObject=Dictionary(Im=picture), Font=Dictionary(F1=font),
                                            ExtGState=Dictionary(Alpha=Dictionary(ca=0.5)))
            page.obj.Contents = pdf.make_stream(
                b"0 1 0 0 k 4 43 100 34 re f "
                b"q 100 0 0 100 5 65 cm /Im Do Q "
                b"q /Alpha gs 0 0 1 rg 42 72 72 125 re f Q "
                b"0 g BT /F1 12 Tf 23 210 Td (Vector text) Tj ET")
            if index:
                page.obj.Rotate = 90
                page.obj.UserUnit = 1.5
        pdf.save(source, min_version="1.7")

    expected = {"print": (300, Name.FlateDecode),
                "digital-small": (150, Name.DCTDecode),
                "digital-high": (432, Name.FlateDecode)}
    sizes = {}
    for preset, (dpi, image_filter) in expected.items():
        dst = folder / (preset + ".pdf")
        call(exe, source, "transform", preset=preset, output=str(dst), view={"pages": [1]})
        with Pdf.open(dst) as pdf:
            assert len(pdf.pages) == 1
            page = pdf.pages[0]
            assert "/Font" in page.Resources and "/ExtGState" in page.Resources
            img = next(iter(page.get_images().values()))
            assert img.Filter == image_filter
            assert abs(float(img.Width) * 72 / 100 - dpi) <= 1
            if preset.startswith("digital"):
                rgb = page.Resources.ColorSpace.DefaultRGB
                assert rgb[0] == Name.ICCBased and rgb[1].N == 3
                assert rgb[1].read_bytes()[16:20] == b"RGB "
            else:
                assert any(str(item.operator) == "k" for item in parse_content_stream(page))
        sizes[preset] = dst.stat().st_size
    assert sizes["digital-small"] < sizes["digital-high"]

    flat = folder / "zploštěné.pdf"
    call(exe, source, "transform", preset="flatten", output=str(flat), view={"pages": [2]})
    with Pdf.open(source) as old, Pdf.open(flat) as new:
        assert len(new.pages) == 1
        a, b = old.pages[1], new.pages[0]
        for key in ("/MediaBox", "/CropBox", "/BleedBox", "/TrimBox", "/Rotate", "/UserUnit"):
            assert a.obj[key] == b.obj[key]
        images = b.get_images()
        assert len(images) == 1
        assert [str(item.operator) for item in parse_content_stream(b)] == ["q", "cm", "cm", "Do", "Q"]
        image = next(iter(images.values()))
        assert abs(image.Width - 270*1.5*300/72) <= 1
        assert "/Font" not in b.Resources

    copy = folder / "předloha-export.pdf"
    call(exe, source, "transform", preset="export", output=str(copy))
    assert copy.read_bytes() == source.read_bytes()


if __name__ == "__main__":
    gs = shutil.which(sys.argv[1] if len(sys.argv) > 1 else "gs")
    if not gs:
        sys.exit("Test vyžaduje Ghostscript.")
    with TemporaryDirectory(prefix="pdf-proof-test-") as folder:
        check_proofs(str(Path(gs).resolve()), Path(folder))
        check_screen_detail(str(Path(gs).resolve()), Path(folder))
        check_plate_switching(str(Path(gs).resolve()), Path(folder))
        check_export_presets(str(Path(gs).resolve()), Path(folder))
    print("Print proof, editor and export settings tests passed.")
