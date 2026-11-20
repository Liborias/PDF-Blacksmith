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


def path_outline(operators, draw, bounds):
    """Sample the painted contour in page coordinates for crossing selections."""
    entries = operators[draw.first:draw.end]
    if len(entries) > 128:
        return None
    a, b, c, d, x, y = draw.matrix

    def transformed(point):
        u, v = point
        return a*u+c*v+x, b*u+d*v+y

    contours, current, pen = [], [], (0., 0.)
    paint = str(entries[-1].operator)
    close = paint not in {"S", "s"}

    def end_path():
        if len(current) > 1:
            if close:
                current.append(current[0])
            contours.append([transformed(point) for point in current])

    for entry in entries[:-1]:
        name = str(entry.operator)
        values = tuple(map(float, entry.operands))
        if name == "m":
            end_path()
            pen = values
            current = [pen]
        elif name == "re":
            end_path()
            u, v, w, h = values
            pen = (u, v)
            current = [pen, (u+w, v), (u+w, v+h), (u, v+h), pen]
        elif name == "l":
            pen = values
            current.append(pen)
        elif name == "h" and current:
            pen = current[0]
            current.append(pen)
        elif name in {"c", "v", "y"}:
            start, goal = pen, values[-2:]
            first = start if name == "v" else values[:2]
            second = goal if name == "y" else values[-4:-2]
            for step in range(1, 13):
                t = step / 12
                r = 1-t
                current.append(tuple(r**3*start[i] + 3*r*r*t*first[i] + 3*r*t*t*second[i] + t**3*goal[i]
                                     for i in range(2)))
            pen = goal
        if len(current) > 320:
            return None
    if paint in {"s", "b", "b*"} and current:
        current.append(current[0])
    end_path()
    points = [point for line in contours for point in line]
    if not points or len(points) > 320 or not all(isfinite(value) for point in points for value in point):
        return None
    xs, ys = zip(*points)
    margin = (max(0., min(xs)-bounds[0], bounds[2]-max(xs)),
              max(0., min(ys)-bounds[1], bounds[3]-max(ys)))
    axis_aligned_rectangle = (len(entries) == 2 and str(entries[0].operator) == "re"
                              and close and (a == d == 0 or b == c == 0))
    return {"contours": contours, "margin": margin, "rectangle": axis_aligned_rectangle}


def describe_objects(page, operators, candidates, *, outlines=True):
    objects = []
    identified = {candidate.token: candidate for candidate in candidates}
    outline_budget = 32000 if outlines else 0
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
                draw = identified[token]
                obj = {"token": token, "kind": draw.kind, "bounds": bounds}
                if draw.kind == "path" and outline_budget > 0:
                    try:
                        outline = path_outline(operators, draw, bounds)
                    except (ArithmeticError, TypeError, ValueError):
                        outline = None
                    if outline:
                        obj["outline"] = outline
                        outline_budget -= sum(map(len, outline["contours"]))
                objects.append(obj)
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
    if action == "delete":
        page.obj.Contents = document.make_stream(pdf.unparse_content_stream(omit_draws(operators, items)))
        return
    delta = clip_point(edit.get("delta"), "posun")
    if len(items) > 1 and edit.get("grouping") == "group":
        bounds, _ = selected_bounds(page, operators, candidates, items)
        form = object_form(document, pdf.Dictionary(page.resources), bounds, isolated_paint(operators, items))
        original = omit_draws(operators, items) if action == "move" else operators
        append_forms(document, page, original, [(form, (1, 0, 0, 1, *delta))])
        return
    combined, cursor = [], 0
    dx, dy = delta
    for item in items:
        combined.extend(operators[cursor:item.first])
        cursor = item.end
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


def omit_draws(operators, items):
    result, cursor = [], 0
    for item in items:
        result.extend(operators[cursor:item.first])
        cursor = item.end
    result.extend(operators[cursor:])
    return result


def isolated_paint(operators, items, *, parts=None):
    """Replay the selected paints in their original order and graphics state."""
    projected, stack = [], []
    marks, text_mode, current = 0, 0, 0
    for index, op in enumerate(operators[:items[-1].end]):
        item = items[current]
        if item.first <= index < item.end:
            if parts is not None and index == item.first:
                part_start = len(projected)
            projected.append(op)
            if index+1 == item.end:
                if parts is not None:
                    parts.append((part_start, len(projected)))
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
            raise ObjectError("Výběr je ovlivněný textovým ořezem. Na stejné straně lze objekty přesunout nebo duplikovat samostatně.")
        elif name not in TEXT | {"Do", "sh", "INLINE IMAGE"}:
            raise ObjectError("Obsah před objektem nelze při kopírování bezpečně přenést.")
    return [*projected, *[instruction("EMC")] * marks, *[instruction("Q")] * len(stack)]


def selected_bounds(page, operators, candidates, items):
    scene = describe_objects(page, operators, candidates, outlines=False)
    indexed = {obj["token"]: obj["bounds"] for obj in scene["objects"]}
    if any(item.token not in indexed for item in items):
        raise ObjectError("Hranice objektů se nepodařilo určit.")
    boxes = [(box[0]-.01, box[1]-.01, box[2]+.01, box[3]+.01)
             for box in (indexed[item.token] for item in items)]
    bounds = [min(box[0] for box in boxes), min(box[1] for box in boxes),
              max(box[2] for box in boxes), max(box[3] for box in boxes)]
    return bounds, boxes


def snapshot_object(document, number, page, operators, candidates, items, rotation):
    bounds, boxes = selected_bounds(page, operators, candidates, items)
    parts = []
    projection = isolated_paint(operators, items, parts=parts)
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
        isolated.obj.Rotate = (int(page.obj.get("/Rotate", 0)) + rotation) % 360
        if len(parts) > 1:
            isolated.obj.PBObjectParts = pdf.Array(pdf.Array([*span, *box]) for span, box in zip(parts, boxes))
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


def object_form(document, resources, bounds, operators):
    return document.make_stream(pdf.unparse_content_stream(operators), Type=pdf.Name.XObject,
                                Subtype=pdf.Name.Form, FormType=1, BBox=pdf.Array(bounds), Resources=resources)


def append_forms(document, page, operators, forms):
    depth = 0
    for op in operators:
        name = str(op.operator)
        depth += int(name == "q") - int(name == "Q")
        if depth < 0:
            raise ObjectError("Cílová strana má neplatný stav grafiky.")
    if depth:
        raise ObjectError("Cílová strana má neuzavřený stav grafiky.")
    resources = pdf.Dictionary(page.resources)
    resources.XObject = pdf.Dictionary(resources.get("/XObject", {}))
    page.obj.Resources = resources
    result = [instruction("q"), *operators, instruction("Q")]
    for form, matrix in forms:
        name = page.add_resource(form, pdf.Name.XObject, prefix="PBObject")
        result.extend((instruction("q"), instruction("cm", *matrix), instruction("Do", name), instruction("Q")))
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(result))


def clipboard_forms(document, clip, source, grouping):
    parts = source.obj.get("/PBObjectParts")
    if grouping != "separate" or parts is None:
        yield document.copy_foreign(clip.make_indirect(source.as_form_xobject(handle_transformations=False)))
        return
    if not isinstance(parts, pdf.Array):
        raise ObjectError("Schránka neobsahuje platný seznam objektů.")

    operators = pikepdf_operators(source)
    resources = document.copy_foreign(clip.make_indirect(source.resources))
    cursor = 0
    items = []
    for entry in parts:
        if not isinstance(entry, pdf.Array) or len(entry) != 6:
            raise ObjectError("Schránka neobsahuje platný seznam objektů.")
        first, last, *bounds = entry
        if type(first) is not int or type(last) is not int or not cursor <= first < last <= len(operators):
            raise ObjectError("Ve schránce chybí obsah samostatných objektů.")
        x0, y0, x1, y1 = map(float, bounds)
        if not all(map(isfinite, (x0, y0, x1, y1))) or x0 >= x1 or y0 >= y1:
            raise ObjectError("Objekty ve schránce mají neplatné hranice.")
        items.append((ObjectDraw(first, last, "", IDENTITY), (x0, y0, x1, y1)))
        cursor = last
    if not items or sum(item.end for item, _ in items) > 2_000_000:
        raise ObjectError("Tento výběr je pro samostatné vložení příliš složitý. Zkopírujte jej jako skupinu.")
    for item, bounds in items:
        yield object_form(document, resources, bounds, isolated_paint(operators, [item]))


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
        matrix = (a, b, c, d, x-a*cx-c*cy, y-b*cx-d*cy)
        forms = clipboard_forms(document, clip, source, edit.get("grouping"))
        append_forms(document, page, operators, ((form, matrix) for form in forms))
