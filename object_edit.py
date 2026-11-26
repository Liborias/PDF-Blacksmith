"""Read and edit individually painted PDF paths and XObjects."""

from base64 import b64decode, b64encode
from contextlib import closing
from dataclasses import dataclass
from io import BytesIO
from math import isclose, isfinite

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
IMAGE_CROP = pdf.Name.PBImageCrop
IMAGE_QUAD = ((0, 0), (1, 0), (1, 1), (0, 1))
CLIP_END = ("h", "W", "n")


@dataclass(frozen=True)
class ImageMask:
    content_start: int
    shape: str
    content_matrix: tuple  # Image placement in page coordinates, independent of the clip.
    clip_outline: tuple    # Mask geometry in page coordinates.
    outline: tuple         # Visible intersection of the mask and the image in page coordinates.

    def source_outline(self):
        return transformed_points(self.content_matrix, IMAGE_QUAD)


@dataclass(frozen=True)
class ObjectDraw:
    first: int
    end: int
    kind: str
    matrix: tuple
    mask: ImageMask | None = None

    @property
    def token(self):
        return self.end


def draws(page, operators):
    matrix, stack = IDENTITY, []
    in_text, first, simple_path = False, None, False
    xobjects = page.resources.get("/XObject", {})
    skip = 0
    for index, op in enumerate(operators):
        if index < skip:
            continue
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
        if first is None and name == "BMC" and list(args) == [IMAGE_CROP]:
            if cropped := cropped_image_draw(operators, index, matrix, xobjects):
                skip = cropped.end
                yield cropped
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


def cropped_image_draw(operators, first, matrix, xobjects):
    # Recognize only our isolated clip and a single raster paint. The full span
    # becomes the editable object so its clipping path travels with the image.
    if [str(op.operator) for op in operators[first+1:first+3]] != ["q", "m"]:
        return None
    clips = []
    curved = False
    start = first+2
    while start < len(operators) and str(operators[start].operator) == "m":
        close = start+1
        while close < min(start+32, len(operators)) and str(operators[close].operator) in {"l", "c"}:
            close += 1
        if tuple(str(op.operator) for op in operators[close:close+3]) != CLIP_END:
            return None
        try:
            contour = image_clip_contour(operators[start:close], matrix)
        except (TypeError, ValueError, OverflowError):
            return None
        if not contour:
            return None
        curved = str(operators[start+1].operator) == 'c'
        clips.append(contour)
        start = close+3
    paint = start
    placement = matrix
    if paint < len(operators) and str(operators[paint].operator) == "cm":
        try:
            transform = tuple(map(float, operators[paint].operands))
        except (TypeError, ValueError, OverflowError):
            return None
        if len(transform) != 6 or not all(map(isfinite, transform)):
            return None
        placement = compose(matrix, transform)
        paint += 1
    tail = operators[paint:paint+3]
    if [str(op.operator) for op in tail] != ["Do", "Q", "EMC"]:
        return None
    args = tail[0].operands
    target = xobjects.get(str(args[0])) if len(args) == 1 else None
    if not isinstance(target, pdf.Stream) or target.get("/Subtype") != pdf.Name.Image:
        return None
    points = transformed_points(placement, IMAGE_QUAD)
    if not all(isfinite(value) for point in points for value in point):
        return None
    clip = clips[0]
    for contour in clips[1:]:
        clip = clip_convex(clip, contour)
    points = clip_convex(points, clip) if clip else ()
    shape = "ellipse" if len(clips) == 1 and curved else "rectangle"
    return ObjectDraw(first, paint+len(tail), "image", matrix,
                      ImageMask(start, shape, placement, tuple(clip), tuple(points)))


def image_clip_contour(path, matrix):
    # Use the vector path for painting. Sample curves only for bounds and
    # overlap checks, keeping all four extremal vertices of an ellipse.
    commands = tuple(str(op) for _, op in path)
    operands = [tuple(map(float, args)) for args, _ in path]
    if len(path) >= 3 and commands == ("m",) + ("l",)*(len(path)-1) and all(len(pair) == 2 for pair in operands):
        points = transformed_points(matrix, operands)
    elif (commands == ("m", "c", "c", "c", "c") and len(operands[0]) == 2
          and all(len(curve) == 6 for curve in operands[1:])):
        start, = transformed_points(matrix, [operands[0]])
        pen, points = start, [start]
        for curve in operands[1:]:
            first, second, end = transformed_points(matrix, (curve[:2], curve[2:4], curve[4:]))
            points.extend(cubic_samples(pen, first, second, end, steps=24))
            pen = end
        if pen != start:
            return ()
        points.pop()
    else:
        return ()
    return points if all(isfinite(value) for point in points for value in point) else ()


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
            current.extend(cubic_samples(start, first, second, goal))
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
            draw = identified[token]
            mask = draw.mask
            if mask and not polygon_area(mask.outline):
                continue
            bounds = polygon_bounds(mask.outline) if mask else tuple(map(float, item.get_bounds()))
            if all(map(isfinite, bounds)) and bounds[0] <= bounds[2] and bounds[1] <= bounds[3]:
                x0, y0, x1, y1 = bounds
                if x0 == x1 or y0 == y1:
                    bounds = (x0-.25, y0-.25, x1+.25, y1+.25)
                obj = {"token": token, "kind": draw.kind, "bounds": bounds}
                if mask and outlines:
                    obj["image_mask"] = {"source_bounds": polygon_bounds(mask.source_outline()),
                                         "outline": mask.outline, "shape": mask.shape,
                                         "source_outline": mask.source_outline(), "clip_outline": mask.clip_outline}
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
    if action in {"crop", "replace_crop", "reset_crop", "translate_crop"}:
        item = selected_image(items)
        if action == "reset_crop":
            if item.mask is None:
                raise ObjectError("Vyberte obrázek s ořezovou maskou vytvořenou v PDF Blacksmith.")
            # Keep the local image placement and the surrounding graphics state.
            clipped = [*operators[:item.first], operators[item.first+1],
                       *operators[item.mask.content_start:item.end-1], *operators[item.end:]]
        elif action == "translate_crop":
            clipped = translate_image_crop(page, operators, item, edit)
        else:
            clipped = crop_image(operators, item, edit.get("crop"), edit.get("crop_shape", "rectangle"),
                                 replace=action == "replace_crop")
        page.obj.Contents = document.make_stream(pdf.unparse_content_stream(clipped))
        return
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
    for item in items:
        combined.extend(operators[cursor:item.first])
        cursor = item.end
        if action == "duplicate":
            combined.extend(operators[item.first:item.end])
        combined.extend(translated_draw(operators, item, delta))
    combined.extend(operators[cursor:])
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(combined))


def transformed_points(matrix, polygon):
    a, b, c, d, e, f = matrix
    return [(a*x+c*y+e, b*x+d*y+f) for x, y in polygon]


def polygon_bounds(points):
    xs, ys = zip(*points)
    return min(xs), min(ys), max(xs), max(ys)


def signed_area(points):
    if len(points) < 3:
        return 0.
    x, y = points[0]
    return sum((a-x)*(d-y) - (b-y)*(c-x)
               for (a, b), (c, d) in zip(points[1:], points[2:])) / 2


def polygon_area(points):
    return abs(signed_area(points))


def clip_convex(points, clip):
    # Intersect the image contour with the inward half-plane of each clip edge.
    orientation = 1 if signed_area(clip) > 0 else -1
    for (x, y), (u, v) in zip(clip, (*clip[1:], clip[0])):
        if not points:
            break
        dx, dy = u-x, v-y
        clipped = []
        before = points[-1]
        previous = orientation*(dx*(before[1]-y) - dy*(before[0]-x))
        for point in points:
            distance = orientation*(dx*(point[1]-y) - dy*(point[0]-x))
            if distance < 0 < previous or previous < 0 < distance:
                fraction = previous/(previous-distance)
                clipped.append((before[0]+fraction*(point[0]-before[0]), before[1]+fraction*(point[1]-before[1])))
            if distance >= 0:
                clipped.append(point)
            before, previous = point, distance
        points = clipped
    return points


def cubic_samples(start, first, second, end, *, steps=12):
    for step in range(1, steps+1):
        t = step / steps
        r = 1-t
        yield tuple(r**3*start[i] + 3*r*r*t*first[i] + 3*r*t*t*second[i] + t**3*end[i] for i in range(2))


def ellipse_path(rect):
    left, bottom, right, top = rect
    x, y = (left+right)/2, (bottom+top)/2
    k = 0.5522847498307936  # Tangent control offset for a quarter ellipse.
    dx, dy = (right-x)*k, (top-y)*k
    return [
        instruction("m", right, y),
        instruction("c", right, y+dy, x+dx, top, x, top),
        instruction("c", x-dx, top, left, y+dy, left, y),
        instruction("c", left, y-dy, x-dx, bottom, x, bottom),
        instruction("c", x+dx, bottom, right, y-dy, right, y),
    ]


def selected_image(items):
    if len(items) != 1 or items[0].kind != "image":
        raise ObjectError("Vyberte jeden rastrový obrázek.")
    return items[0]


def crop_path(rect, shape):
    if (shape not in ("rectangle", "ellipse") or not isinstance(rect, list) or len(rect) != 4 or not all(
            type(value) in (float, int) and isfinite(value) and abs(value) < 2e6 for value in rect)
            or rect[2] <= rect[0] or rect[3] <= rect[1]):
        raise ObjectError("Vyznačte platný výřez obrázku.")
    left, bottom, right, top = rect
    if shape == "ellipse":
        # Build the vector path in page coordinates to keep its proportions
        # even when the underlying image matrix is skewed or nonuniform.
        path = ellipse_path(rect)
        contour = image_clip_contour(path, IDENTITY)
    else:
        contour = ((left, bottom), (right, bottom), (right, top), (left, top))
        path = [instruction("m", *contour[0]), *(instruction("l", *point) for point in contour[1:])]
    return contour, path


def crop_image(operators, item, rect, shape, *, replace):
    contour, path = crop_path(rect, shape)
    mask = item.mask
    source = mask.source_outline() if mask else transformed_points(item.matrix, IMAGE_QUAD)
    previous_clip = []
    if mask and not replace:
        source = mask.outline
        previous_clip = operators[item.first+2:mask.content_start]
    points = clip_convex(source, contour)
    area = polygon_area(points)
    if area <= 1e-8:
        raise ObjectError("Výřez musí zasahovat do viditelné části obrázku.")
    # PDF stream numbers were rounded when writing any previous mask.
    if isclose(area, polygon_area(source), rel_tol=1e-5, abs_tol=1e-8):
        raise ObjectError("Vyznačte menší výřez obrázku.")
    if mask and replace and isclose(area, polygon_area(mask.outline), rel_tol=1e-5, abs_tol=1e-8):
        overlap = polygon_area(clip_convex(points, mask.outline))
        if isclose(overlap, area, rel_tol=1e-5, abs_tol=1e-8):
            raise ObjectError("Vyznačte jiný výřez obrázku.")
    try:
        inverse = invert(item.matrix)
    except TextError as exc:
        raise ObjectError(str(exc)) from exc
    if not replace and shape == "rectangle" and not any(str(op.operator) == "c" for op in previous_clip):
        # Keep successive polygonal crops compact. Curved paths are retained
        # verbatim and intersected by PDF clipping instead of flattening them.
        previous_clip = []
        path = [instruction("m", *points[0]), *(instruction("l", *point) for point in points[1:])]
    if len(path) > 32:
        raise ObjectError("Tento obrázek nelze bezpečně oříznout.")
    local_path = []
    for operands, operator in path:
        local = transformed_points(inverse, zip(operands[::2], operands[1::2]))
        values = [value for point in local for value in point]
        if not all(map(isfinite, values)):
            raise ObjectError("Tento obrázek nelze bezpečně oříznout.")
        local_path.append(instruction(str(operator), *values))
    content = operators[mask.content_start:item.end-2] if mask else operators[item.first:item.end]
    return [*operators[:item.first], instruction("BMC", IMAGE_CROP), instruction("q"), *previous_clip,
            *local_path, *(instruction(name) for name in CLIP_END), *content,
            instruction("Q"), instruction("EMC"), *operators[item.end:]]


def translate_image_crop(page, operators, item, edit):
    mask = item.mask
    if not mask or not mask.clip_outline:
        raise ObjectError("Vyberte obrázek s ořezovou maskou vytvořenou v PDF Blacksmith.")
    mask_delta = clip_point(edit.get("mask_delta", [0, 0]), "posun masky")
    image_delta = clip_point(edit.get("image_delta", [0, 0]), "posun obrázku")
    scale = edit.get("image_scale", 1)
    if type(scale) not in (float, int) or not isfinite(scale) or not .125 <= scale <= 8:
        raise ObjectError("Obrázek má neplatné měřítko.")
    if scale == 1 and not any((*mask_delta, *image_delta)):
        raise ObjectError("Nejprve posuňte výřez nebo obrázek.")
    quad = mask.source_outline()
    orientation = 1 if signed_area(quad) > 0 else -1
    final_clip = [(x+mask_delta[0], y+mask_delta[1]) for x, y in mask.clip_outline]
    if scale != 1 and "crop" in edit:
        final_clip, _ = crop_path(edit["crop"], edit.get("crop_shape", "rectangle"))
    for (x, y), (u, v) in zip(quad, (*quad[1:], quad[0])):
        nx, ny = (y-v)*orientation, (u-x)*orientation
        before = min(nx*(px-x)+ny*(py-y) for px, py in mask.clip_outline)
        left, bottom = x*scale+image_delta[0], y*scale+image_delta[1]
        after = min(nx*(px-left)+ny*(py-bottom) for px, py in final_clip)
        if after < min(0., before) - .002 * (abs(nx)+abs(ny)):
            raise ObjectError("Úprava by odkryla prázdné místo uvnitř ořezu.")
    try:
        inverse = invert(item.matrix)
    except TextError as exc:
        raise ObjectError(str(exc)) from exc
    clips = operators[item.first+2:mask.content_start]
    if any(mask_delta):
        a,b,c,d,_,_ = inverse
        dx,dy = a*mask_delta[0]+c*mask_delta[1], b*mask_delta[0]+d*mask_delta[1]
        clips = [instruction(str(operator), *(float(value)+(dx if i%2==0 else dy)
                                               for i,value in enumerate(args)))
                 if args else (args, operator) for args, operator in clips]
    content = operators[mask.content_start:item.end-2]
    if scale != 1 or any(image_delta):
        *linear, x, y = mask.content_matrix
        matrix = compose(inverse, (*(scale*value for value in linear),
                                   scale*x+image_delta[0], scale*y+image_delta[1]))
        if not all(map(isfinite, matrix)):
            raise ObjectError("Obrázek má neplatnou transformaci.")
        content = [instruction("cm", *matrix), content[-1]]
    moved = [*operators[:item.first+2], *clips, *content, *operators[item.end-2:]]
    if "crop" in edit:
        end = item.first+2+len(clips)+len(content)
        for index in range(item.first+2, end):
            if isinstance(moved[index], tuple):
                moved[index] = pdf.ContentStreamInstruction(*moved[index])
        shifted = cropped_image_draw(moved, item.first, item.matrix, page.resources.get("/XObject", {}))
        if shifted is None:
            raise ObjectError("Uloženou masku se nepodařilo načíst.")
        return crop_image(moved, shifted, edit["crop"], edit.get("crop_shape", "rectangle"), replace=True)
    return moved


def crop_motion_preview(page, operators, items, edit, view):
    updated = translate_image_crop(page, operators, selected_image(items), edit)
    buffer = BytesIO()
    with pdf.Pdf.new() as sample:
        sample.pages.append(page)
        sample.pages[0].obj.Contents = sample.make_stream(pdf.unparse_content_stream(updated))
        sample.save(buffer, fix_metadata_version=False)
        preview = render_screen(sample, buffer.getvalue(), {**view, "page": 1, "pixels": 1200}, "")
    if preview is None:
        raise ObjectError("Náhled úpravy ořezu se nepodařilo vykreslit.")
    payload = BytesIO()
    with Image.frombytes("RGB", (preview["raster_width"], preview["raster_height"]),
                         bytes(preview["raster"]), "raw", "BGRX") as image:
        image.save(payload, format="PNG")
    return {"image": b64encode(payload.getvalue()).decode("ascii")}


def translated_draw(operators, item, delta):
    dx, dy = delta
    try:
        a, b, c, d, _, _ = invert(item.matrix)
    except TextError as exc:
        raise ObjectError(str(exc)) from exc
    return [instruction("q"), instruction("cm", 1, 0, 0, 1, a*dx+c*dy, b*dx+d*dy),
            *operators[item.first:item.end], instruction("Q")]


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


def crop_preview(page, operators, items, rotation):
    """Render the source image without its app-owned crop, on transparent paper."""
    item = selected_image(items)
    mask = item.mask
    points = mask.source_outline() if mask else transformed_points(item.matrix, IMAGE_QUAD)
    bounds = polygon_bounds(points)
    if not all(map(isfinite, bounds)) or min(bounds[2]-bounds[0], bounds[3]-bounds[1]) <= 0:
        raise ObjectError("Obrázek nemá platný rozsah.")
    parts = []
    projection = isolated_paint(operators, [item], parts=parts)
    if mask:
        first, last = parts[0]
        projection[first:last] = [operators[item.first+1], *operators[mask.content_start:item.end-1]]
    data = BytesIO()
    with pdf.Pdf.new() as sample:
        sample.pages.append(page)
        isolated = sample.pages[0]
        isolated.obj.Contents = sample.make_stream(pdf.unparse_content_stream(projection))
        isolated.obj.MediaBox = pdf.Array(bounds)
        isolated.obj.CropBox = pdf.Array(bounds)
        isolated.obj.Rotate = (int(page.obj.get("/Rotate", 0)) + rotation) % 360
        sample.remove_unreferenced_resources()
        sample.save(data, fix_metadata_version=False)
    with pdfium.PdfDocument(data.getvalue()) as document, closing(document[0]) as preview:
        with closing(preview.render(scale=650 / max(preview.get_size()),
                                    fill_color=(0, 0, 0, 0), draw_annots=False)) as bitmap:
            png = BytesIO()
            bitmap.to_pil().save(png, format="PNG")
            return {"image": b64encode(png.getvalue()).decode("ascii")}


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
