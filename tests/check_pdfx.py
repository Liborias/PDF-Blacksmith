"""Exercise all three PDF/X presets with the bundled Ghostscript version."""

from pathlib import Path
import shutil
import sys
from tempfile import TemporaryDirectory

import pikepdf as p

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from engine import dispatch
from pdfx import PDFXError, STANDARDS, XMP_VERSION, check_candidate, dictionaries
from proof import page_size


def check_presets(gs, profile, folder):
    source = folder / "vstup s průhledností.pdf"
    with p.Pdf.new() as pdf:
        pdf.docinfo.Title = "Zkouška tiskových norem"
        for index in range(3):
            page = pdf.add_blank_page(page_size=(180, 250))
            page.obj.MediaBox = p.Array([-10, 20, 170, 270])
            page.obj.CropBox = p.Array([-5, 25, 165, 265])
            page.obj.TrimBox = p.Array([0, 30, 160, 260])
            page.obj.BleedBox = p.Array([-2, 26, 162, 264])
            page.obj.Resources = p.Dictionary(Font=p.Dictionary(F1=p.Dictionary(
                Type=p.Name.Font, Subtype=p.Name.Type1, BaseFont=p.Name.Helvetica)))
            content = b"0.3 0.6 0 0 k 15 65 90 150 re f BT /F1 12 Tf 20 110 Td (Print) Tj ET"
            if index == 1:
                page.obj.Resources.ExtGState = p.Dictionary(Alpha=p.Dictionary(ca=0.5))
                page.obj.Rotate = 90
                page.obj.UserUnit = 1.5
                content += b" q /Alpha gs 1 0 0 rg 19 35 91 85 re f Q"
            page.obj.Contents = pdf.make_stream(content)
        pdf.save(source, min_version="1.7")

    with p.Pdf.open(source) as original:
        for variant, key in ((1, "pdfx-1a"), (3, "pdfx-3"), (4, "pdfx-4")):
            output = folder / (key + ".pdf")
            req = dict(input=str(source), output=str(output), action="transform", preset=key, ghostscript=gs,
                       profile="custom", custom_icc=str(profile), view={"pages": [2, 3]})
            assert "Ověřeny rámečky" in dispatch(req)
            label, version = STANDARDS[variant]
            with p.Pdf.open(output, attempt_recovery=False) as pdf:
                assert pdf.pdf_version == version and not pdf.check_pdf_syntax() and len(pdf.pages) == 2
                assert str(pdf.docinfo.GTS_PDFXVersion) == label
                assert str(pdf.docinfo.Title) == str(original.docinfo.Title)
                assert pdf.Root.OutputIntents[0].DestOutputProfile.read_bytes() == profile.read_bytes()
                for new, old in zip(pdf.pages, original.pages[1:]):
                    for box in ("mediabox", "cropbox", "trimbox", "bleedbox"):
                        assert page_size(new, box) == page_size(old, box)
                assert "/Font" in pdf.pages[1].Resources
                assert any(obj.get("/ca") == 0.5 for obj in dictionaries(pdf.Root)) == (variant == 4)
                if variant == 4:
                    with pdf.open_metadata(set_pikepdf_as_editor=False, update_docinfo=False) as metadata:
                        assert metadata[XMP_VERSION] == label
                        assert metadata["xmpMM:DocumentID"].startswith("uuid:")
                pdf.pages[0].obj.TrimBox = p.Array([-100, -100, 400, 400])
                try:
                    check_candidate(pdf, variant, (profile.read_bytes(), profile.stem, profile.stem))
                except PDFXError as error:
                    assert "rámečků" in str(error)
                else:
                    raise AssertionError("Invalid trim box accepted")
            print(f"{label} passed")


if __name__ == "__main__":
    if len(sys.argv) != 3 or not (gs := shutil.which(sys.argv[1])):
        sys.exit("Použití: python3 tests/check_pdfx.py CESTA_K_GS_10.08 CESTA_K_PROFILU_CMYK.icc")
    with TemporaryDirectory(prefix="pdfx-test-") as folder:
        check_presets(str(Path(gs).resolve()), Path(sys.argv[2]).resolve(), Path(folder))
