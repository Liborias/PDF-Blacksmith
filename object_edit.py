"""Read and edit individually painted PDF paths and XObjects."""

from base64 import b64decode, b64encode
from contextlib import closing
from dataclasses import dataclass
from io import BytesIO
from math import isfinite

import pikepdf as pdf
import pypdfium2 as pdfium
from PIL import Image

from proof import BOXES, page_box
from reader_render import render_screen
from text_edit import IDENTITY, compose, instruction, integer_marks, invert, pikepdf_operators, TextError


class ObjectError(ValueError):
    pass


PATH = {"m", "l", "c", "v", "y", "h", "re"}
PAINT = {"S", "s", "f", "F", "f*", "B", "B*", "b", "b*"}
STATE = {"q", "Q", "cm", "w", "J", "j", "M", "d", "ri", "i", "gs",
         "CS", "cs", "SC", "sc", "SCN", "scn", "G", "g", "RG", "rg", "K", "k"}
MARKED = {"BMC", "BDC", "EMC", "MP", "DP", "BX", "EX"}
TEXT = {"BT", "ET", "Tc", "Tw", "Tz", "TL", "Tf", "Tr", "Ts", "Td", "TD", "Tm", "T*", "Tj", "TJ"}
CLIPBOARD_LIMIT = 16 << 20


@dataclass(frozen=True)
class ObjectDraw:
    first: int
    end: int
    kind: str
    matrix: tuple

    @property
    def token(self):
        return self.end


def draws(page, operators):
    matrix, stack = IDENTITY, []
    in_text, first, simple_path = False, None, False
    xobjects = page.resources.get("/XObject", {})
    for index, op in enumerate(operators):
        name, args = str(op.operator), op.operands
        if name == "q":
            stack.append(matrix)
        elif name == "Q" and stack:
            matrix = stack.pop()
        elif name == "cm" and len(args) == 6:
            matrix = compose(matrix, tuple(map(float, args)))
        elif name == "BT":
            in_text = True
        elif name == "ET":
            in_text = False

        if in_text:
            continue
        if name in PATH:
            if first is None:
                first, simple_path = index, name in {"m", "re"}
        elif name in PAINT or name == "n":
            if simple_path and name in PAINT:
                yield ObjectDraw(first, index+1, "path", matrix)
            first, simple_path = None, False
        elif first is not None:
            # Paths defining a clip or interleaved with graphics state cannot
            # be moved as one independent paint instruction.
            simple_path = False
        elif name == "Do":
            xobject = xobjects.get(str(args[0]))
            if isinstance(xobject, pdf.Stream):
                kind = {pdf.Name.Image: "image", pdf.Name.Form: "form"}.get(xobject.get("/Subtype"))
                if kind:
                    yield ObjectDraw(index, index+1, kind, matrix)


def marked_draws(page, operators, candidates):
    marked = []
    cursor = 0
    for item in candidates:
        marked.extend(operators[cursor:item.first])
        marked.append(instruction("BDC", pdf.Name.Span, pdf.Dictionary(PBObjectDraw=item.token)))
        marked.extend(operators[item.first:item.end])
        marked.append(instruction("EMC"))
        cursor = item.end
    marked.extend(operators[cursor:])
    buffer = BytesIO()
    with pdf.Pdf.new() as probe:
        probe.pages.append(page)
        sample = probe.pages[0]
        sample.obj.Contents = probe.make_stream(pdf.unparse_content_stream(marked))
        sample.obj.CropBox = sample.mediabox
        probe.save(buffer)
    return buffer.getvalue()


def describe_objects(page, operators, candidates):
    objects = []
    identified = {candidate.token: candidate.kind for candidate in candidates}
    with pdfium.PdfDocument(marked_draws(page, operators, candidates)) as probe, closing(probe[0]) as native:
        for item in native.get_objects(max_depth=1):
            token = integer_marks(item, ("PBObjectDraw",)).get("PBObjectDraw")
            if token not in identified:
                continue
            bounds = tuple(map(float, item.get_bounds()))
            if all(map(isfinite, bounds)) and bounds[0] <= bounds[2] and bounds[1] <= bounds[3]:
                x0, y0, x1, y1 = bounds
                if x0 == x1 or y0 == y1:
                    bounds = (x0-.25, y0-.25, x1+.25, y1+.25)
                objects.append({"token": token, "kind": identified[token], "bounds": bounds})
    return {"objects": objects, "rotation": int(page.obj.get("/Rotate", 0)) % 360,
            "boxes": {key: page_box(page, box) for key, (box, _) in BOXES.items()},
            "unit": float(page.obj.get("/UserUnit", 1))}


def clip_point(point, name):
    if not (isinstance(point, list) and len(point) == 2 and all(
            type(value) in (float, int) and isfinite(value) and abs(value) < 2e6 for value in point)):
        raise ObjectError(f"Neplatná poloha objektu ({name}).")
    return point


def selected_draws(candidates, edit):
    tokens = edit.get("tokens", [edit.get("token")])
    if not isinstance(tokens, list) or not 1 <= len(tokens) <= len(candidates) or any(type(token) is not int for token in tokens):
        raise ObjectError("Vyberte platné objekty.")
    chosen = set(tokens)
    items = [item for item in candidates if item.token in chosen]
    if len(items) != len(chosen) or len(chosen) != len(tokens):
        raise ObjectError("Výběr objektů se mezitím změnil.")
    return items


def modify_page(document, page, operators, candidates, edit):
    items = selected_draws(candidates, edit)
    action = edit.get("operation")
    if action not in {"move", "duplicate", "delete"}:
        raise ObjectError("Neznámá úprava objektu.")
    delta = clip_point(edit.get("delta"), "posun") if action != "delete" else None
    combined, cursor = [], 0
    for item in items:
        combined.extend(operators[cursor:item.first])
        cursor = item.end
        if action == "delete":
            continue
        dx, dy = delta
        try:
            a, b, c, d, _, _ = invert(item.matrix)
        except TextError as exc:
            raise ObjectError(str(exc)) from exc
        original = operators[item.first:item.end]
        if action == "duplicate":
            combined.extend(original)
        combined.extend((instruction("q"), instruction("cm", 1, 0, 0, 1, a*dx+c*dy, b*dx+d*dy),
                         *original, instruction("Q")))
    combined.extend(operators[cursor:])
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(combined))


def isolated_paint(operators, items):
    """Replay the selected paints in their original order and graphics state."""
    projected, stack = [], []
    marks, text_mode, current = 0, 0, 0
    for index, op in enumerate(operators[:items[-1].end]):
        item = items[current]
        if item.first <= index < item.end:
            projected.append(op)
            if index+1 == item.end:
                current += 1
            continue
        name = str(op.operator)
        if name == "q":
            stack.append(text_mode)
        elif name == "Q":
            if not stack:
                raise ObjectError("Stav grafiky před objektem nelze bezpečně zkopírovat.")
            text_mode = stack.pop()
        elif name in {"BMC", "BDC"}:
            marks += 1
        elif name == "EMC":
            marks -= 1
            if marks < 0:
                raise ObjectError("Značky obsahu před objektem nelze bezpečně zkopírovat.")
        if name in PAINT:
            projected.append(instruction("n"))
        elif name in STATE | PATH | MARKED | {"W", "W*", "n"}:
            projected.append(op)
        elif name == "Tr":
            text_mode = int(op.operands[0])
        elif name in {"Tj", "TJ"} and text_mode >= 4:
            raise ObjectError("Objekt je ovlivněný textovým ořezem. Zkopírovat jej lze duplikací na stejné straně.")
        elif name not in TEXT | {"Do", "sh", "INLINE IMAGE"}:
            raise ObjectError("Obsah před objektem nelze při kopírování bezpečně přenést.")
    return [*projected, *[instruction("EMC")] * marks, *[instruction("Q")] * len(stack)]


def snapshot_object(document, number, page, operators, candidates, items, rotation):
    scene = describe_objects(page, operators, candidates)
    included = {item.token for item in items}
    boxes = [obj["bounds"] for obj in scene["objects"] if obj["token"] in included]
    if len(boxes) != len(items):
        raise ObjectError("Hranice objektů se nepodařilo určit.")
    bounds = [min(box[0] for box in boxes)-.01, min(box[1] for box in boxes)-.01,
              max(box[2] for box in boxes)+.01, max(box[3] for box in boxes)+.01]
    projection = isolated_paint(operators, items)
    buffer = BytesIO()
    with pdf.Pdf.new() as sample:
        sample.add_pages_from(document, [number-1], forms="strip", outlines="ignore")
        isolated = sample.pages[0]
        isolated.obj.Contents = sample.make_stream(pdf.unparse_content_stream(projection))
        for key in ("/Annots", "/TrimBox", "/BleedBox", "/ArtBox", "/StructParents"):
            if key in isolated.obj:
                del isolated.obj[key]
        isolated.obj.MediaBox = pdf.Array(bounds)
        isolated.obj.CropBox = pdf.Array(bounds)
        isolated.obj.Rotate = (scene["rotation"] + rotation) % 360
        if intents := document.Root.get("/OutputIntents"):
            sample.Root.OutputIntents = sample.copy_foreign(document.make_indirect(intents))
        sample.remove_unreferenced_resources()
        sample.save(buffer, fix_metadata_version=False)
        data = buffer.getvalue()
        rendered = render_screen(sample, data, {"page": 1, "box": "crop", "pixels": 400}, "")
    result = {"object_clipboard": b64encode(data).decode("ascii") if len(data) <= CLIPBOARD_LIMIT else ""}
    if rendered:
        size = rendered["raster_width"], rendered["raster_height"]
        with Image.frombytes("RGB", size, bytes(rendered["raster"]), "raw", "BGRX") as preview:
            png = BytesIO()
            preview.save(png, format="PNG")
            result["image"] = b64encode(png.getvalue()).decode("ascii")
    return result


def paste_object(document, page, operators, edit, rotation, box):
    u, v = clip_point(edit.get("at", [.5, .5]), "vložení")
    if not 0 <= u <= 1 or not 0 <= v <= 1:
        raise ObjectError("Umístěte objekt dovnitř stránky.")
    target_angle = (int(page.obj.get("/Rotate", 0)) + rotation) % 360
    if target_angle not in (0, 90, 180, 270):
        raise ObjectError("Strana má neplatné otočení.")
    u, v = {0: (u, 1-v), 90: (v, u), 180: (1-u, v), 270: (1-v, 1-u)}[target_angle]
    left, bottom, right, top = page_box(page, box)
    x, y = left + u*(right-left), bottom + v*(top-bottom)
    encoded = edit.get("clipboard")
    if not isinstance(encoded, str) or not 0 < len(encoded) <= 4*((CLIPBOARD_LIMIT+2)//3):
        raise ObjectError("Ve schránce není dostupný objekt.")
    depth = 0
    for op in operators:
        name = str(op.operator)
        depth += int(name == "q") - int(name == "Q")
        if depth < 0:
            raise ObjectError("Cílová strana má neplatný stav grafiky.")
    if depth:
        raise ObjectError("Cílová strana má neuzavřený stav grafiky.")
    with pdf.Pdf.open(BytesIO(b64decode(encoded, validate=True))) as clip:
        if len(clip.pages) != 1:
            raise ObjectError("Ve schránce není platný objekt.")
        source = clip.pages[0]
        units = float(source.obj.get("/UserUnit", 1)) / float(page.obj.get("/UserUnit", 1))
        angle = (target_angle - int(source.obj.get("/Rotate", 0))) % 360
        if units <= 0 or not isfinite(units) or angle not in (0, 90, 180, 270):
            raise ObjectError("Strana má neplatné měřítko nebo otočení.")
        a, b, c, d = (units*i for i in {
            0: (1, 0, 0, 1), 90: (0, 1, -1, 0), 180: (-1, 0, 0, -1), 270: (0, -1, 1, 0),
        }[angle])
        left, bottom, right, top = map(float, source.mediabox)
        cx, cy = (left+right)/2, (bottom+top)/2
        form = document.copy_foreign(clip.make_indirect(source.as_form_xobject(handle_transformations=False)))
        resources = pdf.Dictionary(page.resources)
        resources.XObject = pdf.Dictionary(resources.get("/XObject", {}))
        page.obj.Resources = resources
        name = page.add_resource(form, pdf.Name.XObject, prefix="PBObject")
        page.obj.Contents = document.make_stream(pdf.unparse_content_stream([
            instruction("q"), *operators, instruction("Q"),
            instruction("q"), instruction("cm", a, b, c, d, x-a*cx-c*cy, y-b*cx-d*cy),
            instruction("Do", name), instruction("Q"),
        ]))
