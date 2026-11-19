"""Exercise the merge request used by the Windows tool."""

import json
from pathlib import Path
import subprocess
import sys
from tempfile import TemporaryDirectory

from PIL import ImageCms
from pikepdf import Array, Dictionary, Encryption, Name, OutlineItem, Pdf


ROOT = Path(__file__).resolve().parents[1]


def call(action, inputs, status="ok", **options):
    request = {"action": action, "merge_inputs": inputs, **options}
    result = subprocess.run([sys.executable, str(ROOT / "engine.py")],
                            input=json.dumps(request), text=True, capture_output=True, check=True)
    response = json.loads(result.stdout)
    assert response["status"] == status, response
    return response


def document(path, width, profile=None, encrypted=False):
    with Pdf.new() as pdf:
        for offset in range(3):
            page = pdf.add_blank_page(page_size=(width + offset, 300))
            page.Contents = pdf.make_stream(b"0 0 0 1 k 10 10 35 40 re f")
        with pdf.open_outline() as outline:
            outline.root.extend([OutlineItem("První", 0), OutlineItem("Třetí", 2)])
        pdf.pages[0].Annots = Array([pdf.make_indirect(Dictionary(
            Type=Name.Annot, Subtype=Name.Link, Rect=Array([5, 5, 20, 25]),
            Dest=Array([pdf.pages[2].obj, Name.Fit]),
        ))])
        if profile:
            pdf.Root.OutputIntents = Array([Dictionary(
                Type=Name.OutputIntent, S=Name.GTS_PDFA1,
                OutputConditionIdentifier="sRGB", DestOutputProfile=pdf.make_stream(profile, N=3),
            )])
        pdf.save(path, encryption=Encryption(user="čtení", owner="vlastník") if encrypted else None)


def check_merge(folder):
    first, second, locked, mixed, output = (folder / name for name in (
        "obálka.pdf", "příloha.pdf", "chráněný.pdf", "bez záměru.pdf", "celek.pdf"))
    profile = ImageCms.ImageCmsProfile(ImageCms.createProfile("sRGB")).tobytes()
    document(first, 210, profile)
    document(second, 220, profile)
    document(locked, 230, encrypted=True)
    document(mixed, 240)
    state = second.stat()
    selected_first = {"path": str(first), "view": {"pages": [1, 3]}}
    selected_second = {"path": str(second), "view": {"order": [3, 1], "pages": [1], "rotations": [0, 0, 90]},
                       "source_size": state.st_size, "source_mtime_ns": state.st_mtime_ns}
    inputs = [selected_second, selected_first]

    probe = call("merge_probe", [selected_first, {"path": str(locked)}])["view"]["merge_info"]
    assert probe[0]["pages"] == 3 and "heslo" in probe[1]["message"]
    call("merge", inputs, output=str(output))
    with Pdf.open(output) as pdf:
        assert [page.mediabox[2] for page in pdf.pages] == [222, 210, 212]
        assert pdf.pages[0].Rotate == 90 and not pdf.check_pdf_syntax()
        assert pdf.pages[1].Annots[0].Dest[0].objgen == pdf.pages[2].objgen
        assert pdf.Root.OutputIntents[0].DestOutputProfile.read_bytes() == profile
        assert len(list(pdf.open_outline().root)) == 3
        assert "/GTS_PDFXVersion" not in pdf.docinfo

    before = output.read_bytes()
    change = {**selected_second, "source_mtime_ns": state.st_mtime_ns - 1}
    call("merge", [change, selected_first], status="error", output=str(output), overwrite=True)
    assert output.read_bytes() == before
    source = first.read_bytes()
    call("merge", inputs, status="error", output=str(first), overwrite=True)
    assert first.read_bytes() == source

    incompatible = [selected_first, {"path": str(mixed)}]
    result = call("merge", incompatible, output=str(output), overwrite=True)
    with Pdf.open(output) as pdf:
        assert "/OutputIntents" not in pdf.Root and "ICC" in result["message"]

    protected = [selected_first, {"path": str(locked), "pdf_password": "čtení"}]
    output.unlink()
    call("merge", protected, status="merge_encrypted", output=str(output))
    assert not output.exists()
    call("merge", protected, output=str(output), allow_unencrypted_merge=True)
    with Pdf.open(output) as pdf:
        assert not pdf.is_encrypted and len(pdf.pages) == 5


if __name__ == "__main__":
    with TemporaryDirectory(prefix="pdf-merge-test-") as folder:
        check_merge(Path(folder))
    print("Merge checks passed.")
