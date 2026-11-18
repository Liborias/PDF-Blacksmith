"""Integration checks for print proofs, export settings and page editing."""

from base64 import b64decode
from io import BytesIO
import json
from pathlib import Path
import shutil
import subprocess
import sys
from tempfile import TemporaryDirectory

from PIL import Image
from pikepdf import Array, Dictionary, Name, Pdf, Stream


ROOT = Path(__file__).resolve().parents[1]


def call(exe, pdf, action, **options):
    payload = dict(input=str(pdf), action=action, ghostscript=exe, **options)
    result = subprocess.run([sys.executable, str(ROOT / "engine.py")],
                            input=json.dumps(payload), text=True, capture_output=True, check=True)
    response = json.loads(result.stdout)
    assert response["status"] == "ok", response
    return response


def image(proof):
    return Image.open(BytesIO(b64decode(proof["image"])))


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
    proof = call(exe, fixture, "preview", view={"page": 1})["view"]
    assert proof["pages"] == 11
    assert {plate["name"] for plate in proof["plates"] if plate["active"]} == {
        "Cyan", "Magenta", "Yellow", "Black", "TestSpot",
    }

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


if __name__ == "__main__":
    gs = shutil.which(sys.argv[1] if len(sys.argv) > 1 else "gs")
    if not gs:
        sys.exit("Test vyžaduje Ghostscript.")
    with TemporaryDirectory(prefix="pdf-proof-test-") as folder:
        check_proofs(str(Path(gs).resolve()), Path(folder))
    print("Print proof, editor and export settings tests passed.")
