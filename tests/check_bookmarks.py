"""Verify bookmark navigation data and saved edits through the app's PDF engine."""

from pathlib import Path
from tempfile import TemporaryDirectory

from pikepdf import Array, Dictionary, Name, NameTree, OutlineItem, Pdf, String
from pypdfium2 import PdfDocument

from check_merge import call as merge_request
from preview import call


def make_document(path):
    with Pdf.new() as pdf:
        for width in (200, 220, 240):
            pdf.add_blank_page(page_size=(width, 300))
        tree = NameTree.new(pdf)
        tree["podrobnosti"] = Array([pdf.pages[1].obj, Name.FitH, 240])
        pdf.Root.Names = Dictionary(Dests=tree.obj)
        pdf.Root.Dests = Dictionary(Legacy=Dictionary(D=Array([pdf.pages[2].obj, Name.Fit])))
        with pdf.open_outline() as outline:
            chapter = OutlineItem("Kapitola", 0)
            chapter.is_closed = True
            chapter.children.append(OutlineItem("Podrobnosti", String("podrobnosti"), color=(0.3, 0.2, 0.1)))
            outline.root.extend([
                chapter,
                OutlineItem("Dále", action=Dictionary(S=Name.GoTo, D=Name.Legacy)),
                OutlineItem("Web", action=Dictionary(S=Name.URI, URI="https://example.org/")),
            ])
        pdf.save(path)


def check_bookmarks(folder):
    source, edited, extracted, merged = (folder / name for name in ("vstup.pdf", "upravené.pdf", "výběr.pdf", "spojené.pdf"))
    make_document(source)
    original = source.read_bytes()
    bookmarks = call("", source, "describe")["view"]["bookmarks"]
    assert [(b["title"], b["page"], b["depth"]) for b in bookmarks] == [
        ("Kapitola", 1, 0), ("Podrobnosti", 2, 1), ("Dále", 3, 0), ("Web", 0, 0),
    ]

    changes = [{"index": 2, "title": "Přejmenované podrobnosti"},
               {"index": 3, "title": "Změněný cíl", "page": 1},
               {"index": 0, "title": "Nová záložka", "page": 3}]
    draft = {"order": [3, 1, 2], "bookmarks": changes}
    call("", source, "save_document", view=draft, output=str(edited))
    assert source.read_bytes() == original
    assert [(b["title"], b["page"], b["depth"]) for b in call("", edited, "describe")["view"]["bookmarks"]] == [
        ("Kapitola", 2, 0), ("Přejmenované podrobnosti", 3, 1), ("Změněný cíl", 2, 0), ("Web", 0, 0), ("Nová záložka", 1, 0),
    ]
    with Pdf.open(edited) as pdf:
        chapter, moved, external, added = pdf.open_outline().root
        assert chapter.is_closed and chapter.children[0].destination == String("podrobnosti")
        assert tuple(round(float(value), 1) for value in chapter.children[0].color) == (0.3, 0.2, 0.1)
        assert moved.action is None and moved.destination[0].objgen == pdf.pages[1].objgen
        assert external.action.URI == String("https://example.org/")
        assert added.destination[0].objgen == pdf.pages[0].objgen and not pdf.check_pdf_syntax()
    with PdfDocument(edited) as reader:
        items = {item.get_title(): item for item in reader.get_toc()}
        assert items["Přejmenované podrobnosti"].get_dest().get_index() == 2
        assert items["Změněný cíl"].get_dest().get_index() == 1
        assert items["Nová záložka"].get_dest().get_index() == 0

    bad = {"bookmarks": [{"index": 0, "title": "Chybná", "page": 4}]}
    call("", source, "save_document", expected="error", view=bad, output=str(extracted))
    assert not extracted.exists()
    call("", source, "transform", view=draft, preset="export", output=str(extracted))
    assert call("", extracted, "describe")["view"]["bookmarks"] == call("", edited, "describe")["view"]["bookmarks"]
    merge_request("merge", [{"path": str(source), "view": draft}, {"path": str(source)}], output=str(merged))
    assert any(b["title"] == "Nová záložka" and b["page"] == 1
               for b in call("", merged, "describe")["view"]["bookmarks"])


if __name__ == "__main__":
    with TemporaryDirectory(prefix="pdf-bookmark-test-") as folder:
        check_bookmarks(Path(folder))
    print("Bookmark checks passed.")
