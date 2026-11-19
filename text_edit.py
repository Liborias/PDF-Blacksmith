"""Read text with PDFium and replace positioned runs in PDF content streams."""

from collections import Counter
from contextlib import closing
from ctypes import c_int, c_ubyte, c_ushort, c_void_p, cast
from io import BytesIO
from math import isfinite
from pathlib import Path

import pikepdf as pdf
import pypdfium2 as pdfium

from proof import BOXES, page_box


raw = pdfium.raw
IDENTITY = (1, 0, 0, 1, 0, 0)
PAINT = {"Tj", "TJ", "'", '"'}
POSITION = {"Tm", "Td", "TD", "T*", "'", '"', "ET"}
FAMILIES = ("LiberationSans", "LiberationSerif", "LiberationMono")
UNDERLINE = pdf.Name.PBUnderline


class TextError(ValueError):
    pass


def instruction(name, *operands):
    return list(operands), pdf.Operator(name)


def compose(outer, inner):
    a, b, c, d, e, f = outer
    g, h, i, j, k, l = inner
    return a*g+c*h, b*g+d*h, a*i+c*j, b*i+d*j, a*k+c*l+e, b*k+d*l+f


def invert(matrix):
    a, b, c, d, e, f = matrix
    det = a*d-b*c
    if abs(det) < 1e-12:
        raise TextError("Tento text má neplatnou transformaci.")
    return d/det, -b/det, -c/det, a/det, (c*f-d*e)/det, (b*e-a*f)/det


def positions(operators, resources):
    ctm, line, leading, mode, saved = IDENTITY, IDENTITY, 0, 0, []
    independent, in_text = True, False
    usable = set()
    for index in range(len(operators)-1, -1, -1):
        name = str(operators[index].operator)
        if name in PAINT:
            if independent and name in ("Tj", "TJ"):
                usable.add(index)
            independent = name in ("'", '"')
        elif name in POSITION:
            independent = True
    states, text_tokens, marked = {}, [], []
    clipping = False
    properties = resources.get("/Properties", pdf.Dictionary())
    for index, entry in enumerate(operators):
        name = str(entry.operator)
        operands = entry.operands
        if name == "q":
            saved.append((ctm, leading, mode))
        elif name == "Q" and saved:
            ctm, leading, mode = saved.pop()
        elif name == "Tr":
            mode = int(operands[0])
        elif name in ("BMC", "BDC"):
            value = operands[1] if name == "BDC" else None
            if isinstance(value, pdf.Name):
                value = properties.get(value)
            marked.append(isinstance(value, pdf.Dictionary) and "/ActualText" in value)
        elif name == "EMC" and marked:
            marked.pop()
        elif name == "cm":
            ctm = compose(ctm, tuple(map(float, operands)))
        elif name == "BT":
            in_text, line, clipping = True, IDENTITY, False
            text_tokens.clear()
        elif name == "ET":
            in_text = False
            if clipping:
                for token in text_tokens:
                    states.pop(token, None)
        elif name == "TL":
            leading = float(operands[0])
        elif name == "Tm":
            line = tuple(map(float, operands))
        elif name in ("Td", "TD"):
            x, y = map(float, operands)
            if name == "TD":
                leading = -y
            line = compose(line, (1, 0, 0, 1, x, y))
        elif name in ("T*", "'", '"'):
            line = compose(line, (1, 0, 0, 1, 0, -leading))
        if in_text and name in PAINT:
            clipping |= mode >= 4
            if mode == 0 and index in usable and not any(marked):
                token = index + 1
                states[token] = ctm, line
                text_tokens.append(token)
    return states


def object_key(obj):
    return cast(obj.raw, c_void_p).value


def union(first, second):
    if first is None:
        return second
    return min(first[0], second[0]), min(first[1], second[1]), max(first[2], second[2]), max(first[3], second[3])


def has_underline(operators, token):
    following = operators[token:token+5]
    return ([str(entry.operator) for entry in following] == ["ET", "BMC", "re", "f", "EMC"]
            and tuple(following[1].operands) == (UNDERLINE,))


def source_objects(page, operators, states):
    """Tag a private one-page probe so text objects map to exact operators."""
    marked = []
    for index, entry in enumerate(operators, 1):
        if index in states:
            marked.extend((instruction("BDC", pdf.Name.Span, pdf.Dictionary(PBTextToken=index)), entry, instruction("EMC")))
        else:
            marked.append(entry)
    buffer = BytesIO()
    with pdf.Pdf.new() as probe:
        probe.pages.append(page)
        sample = probe.pages[0]
        sample.obj.Contents = probe.make_stream(pdf.unparse_content_stream(marked))
        # Probe every text object on the sheet, even outside the displayed crop.
        sample.obj.CropBox = sample.mediabox
        probe.save(buffer)
    objects = []
    with pdfium.PdfDocument(buffer.getvalue()) as probe, closing(probe[0]) as native, closing(native.get_textpage()) as layer:
        nested_bounds = {}
        nested = set()
        native_objects = list(native.get_objects(filter=[raw.FPDF_PAGEOBJ_TEXT], textpage=layer))
        for text in native_objects:
            if text.level:
                nested.add(object_key(text))
        if nested:
            for index in range(layer.count_chars()):
                handle = raw.FPDFText_GetTextObject(layer, index)
                key = cast(handle, c_void_p).value
                if key in nested:
                    nested_bounds[key] = union(nested_bounds.get(key), layer.get_charbox(index))
        for text in native_objects:
            mode = raw.FPDFTextObj_GetTextRenderMode(text)
            bounds = nested_bounds.get(object_key(text)) if text.level else text.get_bounds()
            if not bounds or not all(map(isfinite, bounds)) or bounds[0] >= bounds[2] or bounds[1] >= bounds[3]:
                continue
            try:
                value = text.extract().strip()
            except (UnicodeError, pdfium.PdfiumError):
                continue
            if not value:
                continue
            token = 0
            if text.level == 0 and mode == 0:
                marks = []
                for index in range(raw.FPDFPageObj_CountMarks(text)):
                    number = c_int()
                    if raw.FPDFPageObjMark_GetParamIntValue(raw.FPDFPageObj_GetMark(text, index), b"PBTextToken", number):
                        marks.append(number.value)
                if len(marks) == 1 and marks[0] in states:
                    token = marks[0]
            font = text.get_font()
            name = font.get_family_name() or font.get_base_name()
            generic = name.casefold()
            family = 2 if any(word in generic for word in ("courier", "mono", "consolas")) else 1 if any(
                word in generic for word in ("times", "serif", "georgia")) and "sans" not in generic else 0
            size = text.get_font_size()
            matrix = text.get_matrix().get()
            objects.append(dict(token=token, text=value, bounds=bounds, matrix=matrix, font=name, family=family,
                                size=size, bold=font.get_weight() >= 600 or "bold" in generic,
                                italic=any(word in font.get_base_name().casefold() for word in ("italic", "oblique")),
                                underline=bool(token and has_underline(operators, token))))
    counts = Counter(item["token"] for item in objects)
    for item in objects:
        if counts[item["token"]] != 1 or item["size"] <= 0 or not all(map(isfinite, item["matrix"])):
            item["token"] = 0
    return objects


def text_page(page):
    operators = pikepdf_operators(page)
    states = positions(operators, page.resources)
    return operators, states, source_objects(page, operators, states)


def pikepdf_operators(page):
    return pdf.parse_content_stream(page) if "/Contents" in page.obj else []


def describe_text(page):
    _, _, objects = text_page(page)
    return dict(objects=objects, rotation=int(page.obj.get("/Rotate", 0)) % 360,
                boxes={key: page_box(page, attribute) for key, (attribute, _) in BOXES.items()})


def encoded_runs(values, style, folder):
    family, size = style.get("family"), style.get("size")
    if type(family) is not int or not 0 <= family < len(FAMILIES) or type(size) not in (int, float) or not 1 <= size <= 600:
        raise TextError("Zvolte písmo a velikost od 1 do 600 bodů.")
    face = ("Bold" if style.get("bold") else "") + ("Italic" if style.get("italic") else "") or "Regular"
    font_path = Path(folder) / f"{FAMILIES[family]}-{face}.ttf"
    if not font_path.is_file():
        raise TextError("V instalaci chybí písma nástroje Text.")
    data = font_path.read_bytes()
    font_buffer = (c_ubyte*len(data)).from_buffer_copy(data)
    encoded = BytesIO()
    widths = []
    with pdfium.PdfDocument.new() as document:
        font = raw.FPDFText_LoadFont(document, font_buffer, len(data), raw.FPDF_FONT_TRUETYPE, True)
        if not font:
            raise TextError("Zvolené písmo se nepodařilo načíst.")
        try:
            for value in values:
                with closing(document.new_page(20000, 1000)) as page:
                    handle = raw.FPDFPageObj_CreateTextObj(document, font, size)
                    if not handle:
                        raise TextError("Textový objekt se nepodařilo vytvořit.")
                    units = (c_ushort*(len(value.encode("utf-16-le"))//2+1)).from_buffer_copy((value+"\0").encode("utf-16-le"))
                    if not raw.FPDFText_SetText(handle, units) or not raw.FPDFPage_InsertObject(page, handle):
                        raw.FPDFPageObj_Destroy(handle)
                        raise TextError("Text nelze zapsat zvoleným písmem.")
                    width = pdfium.PdfObject(handle, page=page, pdf=document).get_bounds()[2]
                    if not isfinite(width) or width > 19000:
                        raise TextError("Zadaný text je pro jeden textový objekt příliš dlouhý.")
                    widths.append(width)
                    page.gen_content()
            document.save(encoded)
        finally:
            raw.FPDFFont_Close(font)
    with pdfium.PdfDocument(encoded.getvalue()) as document:
        for index, value in enumerate(values):
            with closing(document[index]) as page, closing(page.get_textpage()) as layer:
                objects = list(page.get_objects(filter=[raw.FPDF_PAGEOBJ_TEXT], textpage=layer))
                if len(objects) != 1 or objects[0].extract().strip() != value:
                    raise TextError("Zadaný text nelze zapsat zvoleným písmem.")
    return encoded.getvalue(), widths


def replace_runs(document, page, edits, font_dir):
    operators, states, objects = text_page(page)
    indexed = {item["token"]: item for item in objects if item["token"]}
    changes = edits.get("runs")
    if not isinstance(changes, list) or not 1 <= len(changes) <= 64:
        raise TextError("Vyberte jeden až 64 upravitelných textových objektů.")
    replacements = {}
    for change in changes:
        token, content = change.get("token"), change.get("text")
        if type(token) is not int or token not in indexed or token in replacements or not isinstance(content, str):
            raise TextError("Výběr textu se mezitím změnil. Vyberte jej znovu.")
        if len(content) > 1400 or any(ord(char) < 32 or ord(char) > 0xffff for char in content):
            raise TextError("Každý textový objekt může mít nejvýše 1 400 znaků v jednom řádku.")
        replacements[token] = content.strip()
    values = [value for value in replacements.values() if value]
    drawing = {}
    if values:
        encoded, widths = encoded_runs(values, edits, font_dir)
        resources = pdf.Dictionary(page.resources)
        resources.Font = pdf.Dictionary(resources.get("/Font", pdf.Dictionary()))
        page.obj.Resources = resources
        size = edits["size"]
        with pdf.Pdf.open(BytesIO(encoded)) as templates:
            pages = iter(zip(templates.pages, widths))
            for token, content in replacements.items():
                if not content:
                    continue
                template, width = next(pages)
                show = next((item for item in pikepdf_operators(template) if str(item.operator) in ("Tj", "TJ")), None)
                if show is None:
                    raise TextError("Zadaný text se nepodařilo zakódovat.")
                source_font = next(iter(template.resources.Font.items()))[1]
                font = document.copy_foreign(templates.make_indirect(source_font))
                key = page.add_resource(font, pdf.Name.Font, prefix="Text")
                commands = [instruction("BT"), instruction("Tc", 0), instruction("Tw", 0), instruction("Tz", 100),
                            instruction("Ts", 0), instruction("Tr", 0), instruction("Tf", key, size), show, instruction("ET")]
                if edits.get("underline"):
                    commands.extend((instruction("BMC", UNDERLINE), instruction("re", 0, -size*0.13, width, size*0.055),
                                     instruction("f"), instruction("EMC")))
                drawing[token] = commands
    updated = []
    discard = set()
    for token, entry in enumerate(operators, 1):
        if token in discard:
            continue
        if token not in replacements:
            updated.append(entry)
            continue
        if has_underline(operators, token):
            discard.update(range(token+2, token+6))
        if token in drawing:
            ctm, line = states[token]
            local = compose(invert(ctm), indexed[token]["matrix"])
            updated.extend((instruction("ET"), instruction("q"), instruction("cm", *local), *drawing[token],
                            instruction("Q"), instruction("BT"), instruction("Tm", *line)))
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(updated))
