"""Exercise the operation behind thumbnail transfers between open documents."""

from pathlib import Path
import sys
from tempfile import TemporaryDirectory

from PIL import ImageCms
from pikepdf import Array, Dictionary, Name, Pdf, parse_content_stream

from check_merge import call, document
from color_conversion import attach_gray_output_intent, gray_gamma_icc
from page_transfer import page_colors
from print_profiles import attach_intent, attach_srgb


def check_insertion(folder):
    target, source, draft, locked = (folder / name for name in (
        "cílový.pdf", "zdrojový.pdf", "pracovní.pdf", "šifrovaný.pdf",
    ))
    profile = ImageCms.ImageCmsProfile(ImageCms.createProfile("sRGB")).tobytes()
    document(target, 210, profile)
    document(source, 220)
    document(locked, 230, encrypted=True)
    originals = {path: path.read_bytes() for path in (source, target, locked)}
    state = source.stat()
    selection = {"path": str(source), "source_size": state.st_size,
                 "source_mtime_ns": state.st_mtime_ns,
                 "view": {"order": [3, 2, 1], "pages": [3, 1], "rotations": [0, 0, 90]}}
    insertion = {"source": selection, "before": 1}
    options = {"input": str(target), "output": str(draft), "insert_pages": insertion,
               "view": {"order": [3, 1], "rotations": [180, 0, 0]}}
    view = call("insert_pages", [], **options)["view"]
    assert view["pages"] == 4
    with Pdf.open(draft) as pdf:
        assert [int(page.mediabox[2]) for page in pdf.pages] == [212, 222, 220, 210]
        assert [int(page.get("/Rotate", 0)) for page in pdf.pages] == [0, 90, 0, 180]
        assert pdf.pages[2].Contents.read_bytes() == b"0 0 0 1 k 10 10 35 40 re f"
        assert pdf.pages[2].Annots[0].Dest[0].objgen == pdf.pages[1].objgen
        assert pdf.pages[3].Annots[0].Dest[0].objgen == pdf.pages[0].objgen
        assert pdf.Root.OutputIntents[0].DestOutputProfile.read_bytes() == profile
        assert len(pdf.open_outline().root) == 4
        assert not pdf.check_pdf_syntax()

    draft.unlink()
    for invalid in (
        {**insertion, "before": 3},
        {**insertion, "source": {**selection, "source_mtime_ns": state.st_mtime_ns - 1}},
    ):
        call("insert_pages", [], status="error", **{**options, "insert_pages": invalid})
        assert not draft.exists()

    protected = {"path": str(locked), "pdf_password": "čtení", "view": {"pages": [2]}}
    options["insert_pages"] = {"source": protected, "before": 2}
    call("insert_pages", [], status="merge_encrypted", **options)
    assert not draft.exists()
    call("insert_pages", [], allow_unencrypted_merge=True, **options)
    with Pdf.open(draft) as pdf:
        assert [int(page.mediabox[2]) for page in pdf.pages] == [212, 210, 231]
    assert all(path.read_bytes() == content for path, content in originals.items())


def color_fixture(path, profile=None):
    with Pdf.new() as pdf:
        color = pdf.add_blank_page(page_size=(220, 300))
        color.obj.MediaBox = Array([10, -12, 230, 288])
        color.obj.CropBox = Array([13, -9, 227, 285])
        color.obj.TrimBox = Array([16, -6, 224, 282])
        color.obj.Rotate = 90
        spot = Array([Name.Separation, Name.BrandGreen, Name.DeviceRGB, Dictionary(
            FunctionType=2, Domain=Array([0, 1]), C0=Array([1, 1, 1]), C1=Array([0, 1, 0]), N=1)])
        color.obj.Resources = Dictionary(ColorSpace=Dictionary(Brand=spot),
                                         ExtGState=Dictionary(Alpha=Dictionary(ca=.65, op=True)))
        color.Contents = pdf.make_stream(b"/Alpha gs .15 .45 .8 rg 25 30 35 40 re f /Brand cs 0.9 scn 90 30 30 30 re f")
        gray = pdf.add_blank_page(page_size=(221, 300))
        gray.Contents = pdf.make_stream(b".42 g 15 30 40 60 re f")
        black = pdf.add_blank_page(page_size=(222, 300))
        image = pdf.make_stream(bytes([0, 0, 0, 200, 0, 0, 0, 70]), Type=Name.XObject, Subtype=Name.Image,
                                Width=2, Height=1, ColorSpace=Name.DeviceCMYK, BitsPerComponent=8)
        black.obj.Resources = Dictionary(XObject=Dictionary(Ink=image))
        black.Contents = pdf.make_stream(b"q 80 0 0 40 12 17 cm /Ink Do Q\n")
        color.obj.Annots = Array([pdf.make_indirect(Dictionary(
            Type=Name.Annot, Subtype=Name.Link, Rect=Array([15, 20, 40, 50]), Dest=Array([gray.obj, Name.Fit]),
        ))])
        if profile:
            attach_intent(pdf, profile, "CMYK", "Test")
        pdf.save(path)


def check_colors(folder, ghostscript, cmyk):
    source, destination, output = (folder / name for name in ("color.pdf", "target.pdf", "converted.pdf"))
    color_fixture(source)
    originals = source.read_bytes()
    selection = {"path": str(source), "view": {"pages": [1, 2, 3], "rotations": [180, 0, 0]}}
    insertion = {"source": selection, "before": 1}
    settings = dict(input=str(destination), output=str(output), insert_pages=insertion, ghostscript=ghostscript)
    srgb = ImageCms.ImageCmsProfile(ImageCms.createProfile("sRGB")).tobytes()
    for space, profile in (("CMYK", cmyk), ("Gray", gray_gamma_icc(2.2)), ("RGB", srgb)):
        with Pdf.new() as pdf:
            page = pdf.add_blank_page(page_size=(200, 300))
            page.Contents = pdf.make_stream(b"0 0 0 1 k 10 10 35 40 re f" if space == "CMYK" else
                                           b".5 g 10 10 35 40 re f" if space == "Gray" else
                                           b"0 .3 .8 rg 10 10 35 40 re f")
            if space == "CMYK":
                attach_intent(pdf, profile, "Tisková podmínka", "FOGRA51")
            elif space == "Gray":
                attach_gray_output_intent(pdf, profile, "Gray Gamma 2.2")
            else:
                attach_srgb(pdf, profile)
            pdf.save(destination)
        original_target = destination.read_bytes()
        prepared = call("insert_probe", [], **settings)["view"]["insert_color"]
        assert prepared["convert"] == 1 and prepared["gray"] == 2 and space in prepared["target"], prepared
        assert any("Přímé barvy" in item for item in prepared["warnings"])
        assert any("přetisky" in item for item in prepared["warnings"])
        assert any("průhlednosti" in item for item in prepared["warnings"])
        assert not output.exists()
        result = call("insert_pages", [], **{**settings, "insert_pages": {**insertion, "convert": True}})
        assert "s převodem 1" in result["message"], result
        with Pdf.open(destination) as old, Pdf.open(output) as new, Pdf.open(source) as donor:
            assert new.pages[0].Contents.read_bytes() == old.pages[0].Contents.read_bytes()
            assert space in page_colors(new.pages[1]).profiles
            assert set(page_colors(new.pages[1]).profiles) <= {space, "Gray"}
            assert new.pages[1].obj.Rotate == 270
            for box in ("/MediaBox", "/CropBox", "/TrimBox"):
                assert list(new.pages[1].obj[box]) == list(donor.pages[0].obj[box])
            assert new.pages[1].Annots[0].Dest[0].objgen == new.pages[2].objgen
            for i in (1, 2):
                assert new.pages[1+i].Contents.read_bytes() == donor.pages[i].Contents.read_bytes()
            assert not new.check_pdf_syntax()
        assert destination.read_bytes() == original_target and source.read_bytes() == originals
        output.unlink()
    with Pdf.new() as pdf:
        page = pdf.add_blank_page(page_size=(200, 300))
        page.Contents = pdf.make_stream(b"0 0 0 1 k 10 10 35 40 re f 0 .3 .8 rg 50 50 25 35 re f")
        pdf.save(destination)
    prepared = call("insert_probe", [], **settings)["view"]["insert_color"]
    assert not prepared["convert"] and "jednoznačný" in prepared["reason"]
    call("insert_pages", [], status="error", **{**settings, "insert_pages": {**insertion, "convert": True}})
    assert not output.exists()
    call("insert_pages", [], **settings)
    with Pdf.open(output) as result:
        assert "spot" in page_colors(result.pages[1]).profiles


def check_cmyk_profile_change(folder, ghostscript, source_profile, destination_profile):
    source, destination, output = (folder / name for name in ("source-cmyk.pdf", "target-cmyk.pdf", "new-cmyk.pdf"))
    for path, profile in ((source, source_profile), (destination, destination_profile)):
        with Pdf.new() as pdf:
            page = pdf.add_blank_page(page_size=(200, 300))
            page.Contents = pdf.make_stream(b".25 .38 .53 .28 k 10 10 35 40 re f")
            attach_intent(pdf, profile, "Testovací ICC", "Test")
            pdf.save(path)
    insertion = {"source": {"path": str(source), "view": {"pages": [1]}}, "before": 1, "convert": True}
    call("insert_pages", [], input=str(destination), output=str(output), insert_pages=insertion, ghostscript=ghostscript)
    with Pdf.open(output) as converted:
        paint = next(args for args, op in parse_content_stream(converted.pages[1]) if str(op) == "k")
        assert max(abs(float(a)-b) for a, b in zip(paint, (.25, .38, .53, .28))) > .01, paint
        assert converted.Root.OutputIntents[0].DestOutputProfile.read_bytes() == destination_profile


if __name__ == "__main__":
    with TemporaryDirectory(prefix="pdf-insert-test-") as folder:
        folder = Path(folder)
        check_insertion(folder)
        if len(sys.argv) in (3, 4):
            ghostscript, target_profile = str(Path(sys.argv[1]).resolve()), Path(sys.argv[2]).read_bytes()
            check_colors(folder, ghostscript, target_profile)
            if len(sys.argv) == 4:
                check_cmyk_profile_change(folder, ghostscript, Path(sys.argv[3]).read_bytes(), target_profile)
    print("Page insertion checks passed.")
