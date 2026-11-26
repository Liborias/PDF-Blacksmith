"""Exercise the layer tree without changing the PDF or the screen reader."""

from pathlib import Path
from tempfile import TemporaryDirectory

from pikepdf import Array, Dictionary, Name, Pdf, String

from engine import dispatch
from page_inventory import inspect_page, MAX_ITEMS
from reader_render import ScreenReader


def fixture(path):
    with Pdf.new() as doc:
        page = doc.add_blank_page(page_size=(540, 400))
        doc.add_blank_page(page_size=(420, 340))
        printed = doc.make_indirect(Dictionary(Type=Name.OCG, Name=String("Tisková grafika")))
        internal = doc.make_indirect(Dictionary(Type=Name.OCG, Name=String("Interní podklad")))
        unknown = doc.make_indirect(Dictionary(Type=Name.OCG, Name=String("Beze změny")))
        doc.Root.OCProperties = Dictionary(
            OCGs=Array([printed, internal, unknown]),
            D=Dictionary(BaseState=Name.Unchanged, ON=Array([printed]), OFF=Array([internal]), Locked=Array([internal])),
        )
        condition = doc.make_indirect(Dictionary(Type=Name.OCMD, OCGs=Array([printed]),
                                               VE=Array([Name.And, printed, Array([Name.Not, internal])])))
        font = doc.make_indirect(Dictionary(Type=Name.Font, Subtype=Name.Type1, BaseFont=Name.Helvetica))
        inner = doc.make_stream(b"0 .1 .9 rg 0 0 64 29 re f",
                                Type=Name.XObject, Subtype=Name.Form, BBox=Array([0, 0, 64, 29]),
                                Resources=Dictionary())
        form = doc.make_stream(
            b"q .5 .1 .6 rg 0 0 78 50 re f 1 0 0 1 5 8 cm /Inner Do Q",
            Type=Name.XObject, Subtype=Name.Form, BBox=Array([0, 0, 80, 53]),
            Group=Dictionary(S=Name.Transparency, I=True), OC=printed,
            Resources=Dictionary(XObject=Dictionary(Inner=inner)),
        )
        rgba_background = doc.make_indirect(Dictionary(Type=Name.ExtGState, ca=.55))
        page.Resources = Dictionary(Font=Dictionary(F=font), XObject=Dictionary(Group=form),
                                   ExtGState=Dictionary(Alpha=rgba_background),
                                   Properties=Dictionary(Internal=internal, Condition=condition))
        page.Contents = doc.make_stream(
            b".9 .8 .5 rg 0 0 540 400 re f "
            b"/OC /Internal BDC q .9 .2 .1 rg 40 310 78 35 re f Q EMC "
            b"BT /F 18 Tf 18 TL 0 g 142 316 Td (Inventory) ' ET "
            b"/OC /Condition BDC q /Alpha gs 1 0 0 1 106 161 cm /Group Do Q EMC "
            b"q 1 0 0 1 280 161 cm /Group Do Q "
            b"q BT /F 13 Tf 3 Tr 12 55 Td (invisible) Tj ET Q "
            b"0 .5 .75 rg 38 102 75 23 re f"
        )
        appearance = doc.make_stream(b".5 0 .5 rg 0 0 70 24 re f", Type=Name.XObject, Subtype=Name.Form,
                                     BBox=Array([0, 0, 70, 24]))
        field = doc.make_indirect(Dictionary(Type=Name.Annot, Subtype=Name.Widget, FT=Name.Tx, T=String("Pole"),
                                             V=String("hodnota"), Rect=Array([372, 279, 442, 303]), F=4,
                                             AP=Dictionary(N=appearance), P=page.obj))
        fixed = doc.make_indirect(Dictionary(Type=Name.Annot, Subtype=Name.Widget, FT=Name.Tx, T=String("Zámek"),
                                             Rect=Array([372, 218, 442, 242]), F=132,
                                             AP=Dictionary(N=appearance), P=page.obj))
        page.Annots = Array([field, fixed])
        doc.Root.AcroForm = Dictionary(Fields=Array([field, fixed]), DR=Dictionary(Font=Dictionary(F=font)),
                                       DA=String("/F 11 Tf 0 g"))
        doc.save(path)


def check_tree(path):
    original = path.read_bytes()
    reader = ScreenReader()
    source = dict(input=str(path), view=dict(page=1, pixels=540))
    try:
        before = dispatch(dict(source, action="screen"), raster=True, reader=reader)
        tree = dispatch(dict(source, action="page_inventory"), raster=True, reader=reader)["inventory"]
        after = dispatch(dict(source, action="screen"), raster=True, reader=reader)
        assert before == after, "metadata read changed pixels or color settings"
        empty = dispatch(dict(source, action="page_inventory", view=dict(page=2)), reader=reader)["inventory"]
    finally:
        reader.close()
    assert path.read_bytes() == original
    printed, internal, _ = tree["layers"]
    assert [(layer["name"], layer["state"], layer["locked"]) for layer in tree["layers"]] == [
        ("Tisková grafika", "on", False), ("Interní podklad", "off", True), ("Beze změny", "", False),
    ]
    assert empty["items"] == [] and empty["layers"] == tree["layers"] and not tree["limited"]
    art = tree["items"]
    assert [item["kind"] for item in art] == [
        "path", "path", "text", "group", "group", "hidden_text", "path", "widget", "widget",
    ]
    assert art[1]["layers"] == (internal["id"],)
    assert set(art[3]["layers"]) == {printed["id"], internal["id"]}
    assert art[4]["layers"] == (printed["id"],)
    one, two = art[3]["children"], art[4]["children"]
    assert [item["kind"] for item in one] == ["path", "form"]
    assert one[1]["children"][0]["kind"] == "path"
    assert one[1]["children"][0]["layers"] == art[3]["layers"]
    assert one[1]["children"][0]["id"] != two[1]["children"][0]["id"]
    assert "object" not in one[0] and "text" not in art[5]
    vectors = dispatch(dict(source, action="object_probe"))["object_page"]["objects"]
    assert [item["object"] for item in art if "object" in item] == [item["token"] for item in vectors]
    assert art[-2]["object"] == -1 and "object" not in art[-1]
    text = dispatch(dict(source, action="text_probe"))["text_page"]["objects"]
    assert any(run["block_token"] == art[2]["text"] and run["text"] == "Inventory" for run in text)


def check_limit():
    with Pdf.new() as doc:
        page = doc.add_blank_page()
        page.Contents = doc.make_stream(b"1 2 3 4 re f\n" * (MAX_ITEMS + 3))
        tree = inspect_page(doc, page)
        assert tree["limited"] and len(tree["items"]) == MAX_ITEMS


if __name__ == "__main__":
    with TemporaryDirectory() as folder:
        path = Path(folder) / "layers.pdf"
        fixture(path)
        check_tree(path)
        check_limit()
    print("Page inventory preserves paint order, OCGs, Form instances, widget permissions and screen pixels.")
