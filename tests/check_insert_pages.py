"""Exercise the operation behind thumbnail transfers between open documents."""

from pathlib import Path
from tempfile import TemporaryDirectory

from PIL import ImageCms
from pikepdf import Pdf

from check_merge import call, document


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


if __name__ == "__main__":
    with TemporaryDirectory(prefix="pdf-insert-test-") as folder:
        check_insertion(Path(folder))
    print("Page insertion checks passed.")
