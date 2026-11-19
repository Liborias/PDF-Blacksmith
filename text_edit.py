"""Read text with PDFium and replace positioned runs in PDF content streams."""

from collections import Counter
from contextlib import closing
from ctypes import byref, c_double, c_float, c_int, c_ubyte, c_uint, c_ushort, c_void_p, cast
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


def positions(operators, resources, *, include_dependent=False):
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
            if (mode in (0, 2) and (index in usable or include_dependent and name in ('Tj', 'TJ'))
                    and not any(marked)):
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


def source_styles(operators, resources, marked, initial_color=()):
    fonts = resources.get("/Font", pdf.Dictionary())
    font, size, tracking, spacing, color = None, 0, 0, 0, initial_color
    stack, output = [], {}
    for token, entry in enumerate(operators, 1):
        name, args = str(entry.operator), entry.operands
        if name == "q":
            stack.append((font, size, tracking, spacing, color))
        elif name == "Q" and stack:
            font, size, tracking, spacing, color = stack.pop()
        elif name == "Tf":
            font, size = fonts.get(args[0]), float(args[1])
        elif name == "Tc":
            tracking = float(args[0])
        elif name == "Tw":
            spacing = float(args[0])
        elif name == '"':
            spacing, tracking = map(float, args[:2])
        elif name in ("g", "rg", "k", "cs"):
            color = (entry,)
        elif name in ("sc", "scn"):
            color = (*[item for item in color if str(item.operator) == "cs"], entry)
        if token in marked:
            output[token] = font, size, tracking, spacing, color
    return output


def show_items(entry):
    return [entry.operands[0]] if str(entry.operator) == "Tj" else entry.operands[0]


def code_size(font):
    if font is None:
        return 0
    if font.Subtype == pdf.Name.Type0:
        return 2 if font.get("/Encoding") == pdf.Name("/Identity-H") else 0
    return 1 if font.Subtype in (pdf.Name.Type1, pdf.Name.TrueType, pdf.Name.MMType1) else 0


def source_widths(font, step, codes):
    """Get advances from PDF character codes, even when they differ from Unicode."""
    if step == 1:
        widths = font.get("/Widths", [])
        first = int(font.get("/FirstChar", 0))
        return {code: float(widths[code-first]) for code in codes if 0 <= code-first < len(widths)}
    descendant = font.DescendantFonts[0]
    widths = {code: float(descendant.get("/DW", 1000)) for code in codes}
    rules = iter(descendant.get("/W", []))
    for first, values in zip(rules, rules):
        first = int(first)
        if isinstance(values, pdf.Array):
            for code in codes:
                if 0 <= code-first < len(values):
                    widths[code] = float(values[code-first])
        else:
            last, width = int(values), float(next(rules))
            for code in codes:
                if first <= code <= last:
                    widths[code] = width
    return widths


def glyph_advances(entry, style, font, glyphs, matrix):
    source, size, tracking, spacing, _ = style
    step = code_size(source)
    if not step or not glyphs or size <= 0:
        return None
    strings = [bytes(item) for item in show_items(entry) if isinstance(item, pdf.String)]
    if any(len(value) % step for value in strings):
        return None
    try:
        inverse = invert(matrix)
        codes = {int.from_bytes(value[offset:offset+step], "big") for value in strings
                 for offset in range(0, len(value), step)}
        widths = source_widths(source, step, codes)
    except (TextError, AttributeError, IndexError, TypeError, ValueError, StopIteration):
        return None
    advances, count, position, start = [], 0, 0.0, None
    for value in show_items(entry):
        if not isinstance(value, pdf.String):
            position -= float(value)
            continue
        data = bytes(value)
        for offset in range(0, len(data), step):
            if count >= len(glyphs):
                return None
            glyph = glyphs[count]
            unicode_code = ord(glyph["text"])
            if unicode_code > 0xffff or glyph["text"] == "\ufffd":
                return None
            code = int.from_bytes(data[offset:offset+step], "big")
            width = widths.get(code)
            if width is None:
                measured = c_float()
                if not raw.FPDFFont_GetGlyphWidth(font, unicode_code, 1000, byref(measured)):
                    return None
                width = measured.value
            x, y = glyph["origin"]
            origin = inverse[0]*x + inverse[2]*y + inverse[4]
            start = origin if start is None else start
            if abs(origin-start-position*size/1000) > max(.08, size*.012):
                return None
            advance = width + 1000 * (tracking + (spacing if step == 1 and code == 32 else 0)) / size
            if not isfinite(advance):
                return None
            advances.append(advance)
            position += advance
            count += 1
    return advances if count == len(glyphs) else None

def integer_marks(text, names):
    result = {}
    for mark_index in range(raw.FPDFPageObj_CountMarks(text)):
        handle = raw.FPDFPageObj_GetMark(text, mark_index)
        for name in names:
            value = c_int()
            if raw.FPDFPageObjMark_GetParamIntValue(handle, name.encode("ascii"), value):
                if name in result:
                    return {}
                result[name] = value.value
    return result


def stored_block(text):
    group = integer_marks(text, ("PBTextBlock",)).get("PBTextBlock", 0)
    if group <= 0:
        return None
    block = integer_marks(text, ("PBWidth", "PBFirstX", "PBFirstY", "PBAngle", "PBAlign",
                                 "PBLeading", "PBTracking", "PBRow", "PBParagraph", "PBWordWrap", "PBBold", "PBItalic"))
    required = ("PBWidth", "PBFirstX", "PBFirstY", "PBAngle", "PBAlign", "PBLeading", "PBTracking",
                "PBRow", "PBParagraph", "PBWordWrap")
    if not all(name in block for name in required):
        return None
    return dict(id=group, width=block["PBWidth"]/1000,
                origin=(block["PBFirstX"]/1000, block["PBFirstY"]/1000), rotation=block["PBAngle"],
                alignment=block["PBAlign"], leading=block["PBLeading"]/1000, tracking=block["PBTracking"]/1000,
                row=block["PBRow"], paragraph=block["PBParagraph"], word_wrap=bool(block["PBWordWrap"]),
                bold=block.get("PBBold"), italic=block.get("PBItalic"))

def text_tree(page, document=None, *, probe=False, replacements=None, discard=()):
    """Assign each show an ID in page paint order, including instances of shared Forms."""
    root_operators = pikepdf_operators(page)
    sources, serial, forms = {}, len(root_operators), 0

    def visit(operators, resources, parents, outer, initial_color):
        nonlocal serial, forms
        states = positions(operators, resources)
        block_states = positions(operators, resources, include_dependent=True)
        form_calls = {i for i, op in enumerate(operators, 1) if str(op.operator) == "Do"}
        styles = source_styles(operators, resources, block_states.keys() | form_calls, initial_color)
        mapping = {}
        for index in block_states:
            if parents:
                serial += 1
                token = serial
            else:
                token = index
            mapping[index] = token
            sources[token] = (operators, index, styles[index], index in states, outer, not parents)
        skip = {offset for index, token in mapping.items() if token in discard
                for offset in range(index+2, index+6)}
        output, new_resources, changed = [], None, False
        xobjects = resources.get("/XObject", {})
        ctm, matrices = outer, []
        for index, op in enumerate(operators, 1):
            name, operands = str(op.operator), op.operands
            if name == "q":
                matrices.append(ctm)
            elif name == "Q" and matrices:
                ctm = matrices.pop()
            elif name == "cm":
                ctm = compose(ctm, tuple(map(float, operands)))
            if (name == "Do" and len(parents) < 8 and forms < 256 and hasattr(xobjects, "get")):
                form = xobjects.get(operands[0])
                if (isinstance(form, pdf.Stream) and form.get("/Subtype") == pdf.Name.Form
                        and form.objgen not in parents and "/Ref" not in form
                        and "/Group" not in form):
                    matrix = form.get("/Matrix", IDENTITY)
                    try:
                        child_ops = pikepdf_operators(form)
                        child_matrix = compose(ctm, tuple(map(float, matrix)))
                    except (pdf.PdfError, ValueError, TypeError):
                        child_ops = []
                    if child_ops:
                        forms += 1
                        child_resources = form.get("/Resources")
                        if not isinstance(child_resources, pdf.Dictionary):
                            child_resources = resources
                        child, child_resources, update = visit(child_ops, child_resources, (*parents, form.objgen),
                                                               child_matrix, styles[index][4])
                        if update:
                            clone = document.make_stream(pdf.unparse_content_stream(child))
                            for key, value in form.stream_dict.items():
                                if key not in ("/Length", "/Filter", "/DecodeParms", "/Resources"):
                                    clone[key] = value
                            clone.Resources = child_resources
                            if new_resources is None:
                                new_resources = pdf.Dictionary(resources)
                                new_resources.XObject = pdf.Dictionary(xobjects)
                            key_index = index
                            key = pdf.Name(f"/PBTextForm{key_index}")
                            while key in new_resources.XObject:
                                key_index += 1
                                key = pdf.Name(f"/PBTextForm{key_index}")
                            new_resources.XObject[key] = clone
                            op = instruction("Do", key)
                            changed = True
            token = mapping.get(index)
            if probe and token is not None:
                output.extend((instruction("BDC", pdf.Name.Span, pdf.Dictionary(PBTextToken=token)),
                               op, instruction("EMC")))
                changed = True
            elif replacements is not None and token in replacements:
                output.append(replacements[token])
                changed = True
            elif index in skip:
                changed = True
            else:
                output.append(op)
        return output, new_resources if new_resources is not None else resources, changed

    output, resources, changed = visit(root_operators, page.resources, (), IDENTITY, ())
    if changed:
        page.obj.Resources = resources
    return output, sources


def source_objects(page, sources):
    """Tag a private one-page probe so text objects map to exact operators."""
    buffer = BytesIO()
    with pdf.Pdf.new() as probe:
        probe.pages.append(page)
        sample = probe.pages[0]
        marked, _ = text_tree(sample, probe, probe=True)
        sample.obj.Contents = probe.make_stream(pdf.unparse_content_stream(marked))
        # Probe every text object on the sheet, even outside the displayed crop.
        sample.obj.CropBox = sample.mediabox
        probe.save(buffer)
    objects = []
    with pdfium.PdfDocument(buffer.getvalue()) as probe, closing(probe[0]) as native, closing(native.get_textpage()) as layer:
        native_objects = list(native.get_objects(filter=[raw.FPDF_PAGEOBJ_TEXT], textpage=layer))
        glyphs = {}
        for index in range(layer.count_chars()):
            handle = raw.FPDFText_GetTextObject(layer, index)
            if not handle or raw.FPDFText_IsGenerated(layer, index) != 0:
                continue
            code = raw.FPDFText_GetUnicode(layer, index)
            if not 0 < code <= 0x10ffff or 0xd800 <= code <= 0xdfff:
                continue
            x, y = c_double(), c_double()
            if not raw.FPDFText_GetCharOrigin(layer, index, byref(x), byref(y)):
                continue
            bounds = layer.get_charbox(index)
            if bounds[0] >= bounds[2] or bounds[1] >= bounds[3]:
                bounds = layer.get_charbox(index, loose=True)
            if all(map(isfinite, (*bounds, x.value, y.value))):
                glyphs.setdefault(cast(handle, c_void_p).value, []).append(
                    dict(text=chr(code), bounds=bounds, origin=(x.value, y.value)))
        for text in native_objects:
            mode = raw.FPDFTextObj_GetTextRenderMode(text)
            characters = glyphs.get(object_key(text), [])
            bounds = text.get_bounds() if not text.level else None
            if text.level:
                for character in characters:
                    bounds = union(bounds, character["bounds"])
            if not bounds or not all(map(isfinite, bounds)) or bounds[0] >= bounds[2] or bounds[1] >= bounds[3]:
                continue
            try:
                value = text.extract().strip()
            except (UnicodeError, pdfium.PdfiumError):
                continue
            if not value:
                continue
            token = 0
            if mode in (0, 2):
                number = integer_marks(text, ("PBTextToken",)).get("PBTextToken", 0)
                if number in sources:
                    token = number
            font = text.get_font()
            name = font.get_family_name() or font.get_base_name()
            generic = name.casefold()
            family = 2 if any(word in generic for word in ("courier", "mono", "consolas")) else 1 if any(
                word in generic for word in ("times", "serif", "georgia")) and "sans" not in generic else 0
            size = text.get_font_size()
            matrix = text.get_matrix().get()
            advance, whole, root = [], False, False
            operator_list, index = (), 0
            if token:
                operator_list, index, style, whole, outer, root = sources[token]
                if not root:
                    matrix = compose(outer, matrix)
                advance = glyph_advances(operator_list[index-1], style, font, characters, matrix) or []
            block = stored_block(text)
            bold = font.get_weight() >= 600 or "bold" in generic
            italic = any(word in font.get_base_name().casefold() for word in ("italic", "oblique"))
            if block:
                bold = bold if block["bold"] is None else bool(block["bold"])
                italic = italic if block["italic"] is None else bool(block["italic"])
            red, green, blue, alpha = c_uint(), c_uint(), c_uint(), c_uint()
            raw.FPDFPageObj_GetFillColor(text, byref(red), byref(green), byref(blue), byref(alpha))
            objects.append(dict(token=token if whole and root else 0, block_token=token, advances=advance, whole=whole,
                                glyphs=characters, block=block, text=value, bounds=bounds, matrix=matrix, font=name, family=family,
                                fill=[red.value, green.value, blue.value],
                                size=size, bold=bold, italic=italic,
                                underline=bool(token and has_underline(operator_list, index))))
    counts = Counter(item["block_token"] for item in objects)
    for item in objects:
        if counts[item["block_token"]] != 1 or item["size"] <= 0 or not all(map(isfinite, item["matrix"])):
            item["token"] = item["block_token"] = 0
            item["whole"] = False
            item["advances"] = []
    return objects


def text_page(page):
    operators, sources = text_tree(page)
    states = positions(operators, page.resources)
    return operators, states, source_objects(page, sources)


def pikepdf_operators(page):
    if isinstance(page, pdf.Page) and "/Contents" not in page.obj:
        return []
    operators = []
    for entry in pdf.parse_content_stream(page):
        name = str(entry.operator)
        if name not in ("'", '"'):
            operators.append(entry)
            continue
        if name == '"':
            operators.extend((instruction("Tw", entry.operands[0]), instruction("Tc", entry.operands[1])))
        operators.extend((instruction("T*"), instruction("Tj", entry.operands[-1])))
    # Keep one representation for entries parsed from the PDF and expanded shorthand.
    return [pdf.ContentStreamInstruction(*entry) if isinstance(entry, tuple) else entry for entry in operators]


def describe_text(page):
    _, _, objects = text_page(page)
    for item in objects:
        partial = bool(item.pop("advances"))
        item["editable"] = item.pop("whole") or partial
        item["whole_run_only"] = not partial
    return dict(objects=objects, rotation=int(page.obj.get("/Rotate", 0)) % 360,
                boxes={key: page_box(page, attribute) for key, (attribute, _) in BOXES.items()})


def encoded_runs(values, style, folder, font_data=None):
    family, size = style.get("family"), style.get("size")
    if type(family) is not int or not 0 <= family < len(FAMILIES) or type(size) not in (int, float) or not 1 <= size <= 600:
        raise TextError("Zvolte písmo a velikost od 1 do 600 bodů.")
    face = ("Bold" if style.get("bold") else "") + ("Italic" if style.get("italic") else "") or "Regular"
    font_path = Path(folder) / f"{FAMILIES[family]}-{face}.ttf"
    if font_data is None and not font_path.is_file():
        raise TextError("V instalaci chybí písma nástroje Text.")
    data = font_path.read_bytes() if font_data is None else font_data
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
