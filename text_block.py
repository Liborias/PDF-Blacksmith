"""Reflow a selected range and write searchable text into its page."""

from itertools import groupby
from math import isfinite
import re

import pikepdf as pdf

from block_fonts import embed_face, encoded_fragment, fonts_for_block
from text_edit import TextError, UNDERLINE, code_size, has_underline, instruction, show_items


def numeric(value, low, high, message):
    if type(value) not in (float, int) or not low <= value <= high:
        raise TextError(message)
    return float(value)


def wrap_block(content, advances, size, width, tracking):
    missing = next((char for char in content if not char.isspace() and char not in advances), None)
    if missing:
        raise TextError(f"Zvolené písmo neobsahuje znak {missing}.")
    guessed = {char: advance*size + tracking for char, advance in advances.items()}
    if " " not in guessed:
        raise TextError("Zvolené písmo nemá znak mezery.")
    for paragraph_index, paragraph in enumerate(content.split("\n")):
        pieces = re.finditer(r"(\s*)(\S+)", paragraph.expandtabs(4))
        value, extent = "", 0.
        for piece in pieces:
            spaces, word = piece.groups()
            advance = sum(guessed[char] for char in word)
            gap = len(spaces) * guessed[" "]
            if value and extent + gap + advance - tracking <= width:
                value += " " * len(spaces) + word
                extent += gap + advance
                continue
            if value:
                yield value, extent-tracking, True, paragraph_index
                value, extent = "", 0.
            if advance-tracking <= width:
                value, extent = word, advance
                continue
            for char in word:
                cost = guessed[char]
                if cost-tracking > width:
                    raise TextError("Zvětšete šířku bloku nebo zmenšete písmo.")
                if value and extent + cost - tracking > width:
                    yield value, extent-tracking, False, paragraph_index
                    value, extent = "", 0.
                value += char
                extent += cost
        yield value, extent-tracking if value else 0, False, paragraph_index


def rich_lines(glyphs, width, wrap):
    for row, source in enumerate(glyphs):
        line = source
        while wrap and line:
            extent = 0.
            end, gap = len(line), 0
            for index, (char, _, advance) in enumerate(line):
                if index and extent + advance > width:
                    end = gap or index
                    break
                extent += advance
                if char == " ":
                    gap = index + 1
            if end == len(line):
                break
            yield line[:end], row, True
            line = line[end:]
        yield line, row, False


def rich_block_commands(document, page, block, content, folder, size, width, height,
                        leading, alignment, origin, angle, color, group, palette):
    ranges = block.get("runs")
    if not isinstance(ranges, list) or not 1 <= len(ranges) <= len(content):
        raise TextError("Textový blok obsahuje neplatné formátování.")
    style_keys, letters, cursor = {}, {}, 0
    for run in ranges:
        start, end = run.get("start"), run.get("end")
        if type(start) is not int or type(end) is not int or not start == cursor < end <= len(content):
            raise TextError("Textový blok obsahuje neplatný rozsah formátování.")
        family = run.get("font")
        if not isinstance(family, str):
            raise TextError("Vyberte dostupné písmo.")
        key = family, bool(run.get("bold")), bool(run.get("italic"))
        span_size = numeric(run.get("size"), 1, 600, "Zadejte platnou velikost písma.")
        tracking = numeric(run.get("tracking"), -72, 72, "Zadejte platný proklad písmen.")
        ink = int(numeric(run.get("color"), 0, 0xffffff, "Zadejte platnou barvu textu."))
        layout_style = key, span_size, tracking, bool(run.get("underline")), ink
        value = content[start:end].replace("\t", "    ")
        style_keys.setdefault(key, set()).update(" " if char.isspace() else char for char in value if char != "\n")
        for index in range(start, end):
            letters[index] = layout_style
        cursor = end
    if cursor != len(content) or len(style_keys) > 128:
        raise TextError("Textový blok obsahuje příliš mnoho různých písem.")

    resources = pdf.Dictionary(page.resources)
    resources.Font = pdf.Dictionary(resources.get("/Font", pdf.Dictionary()))
    page.obj.Resources = resources
    fonts = {}
    for key, chars in style_keys.items():
        family, bold, italic = key
        faces, assigned, advances = fonts_for_block(dict(font=family, bold=bold, italic=italic), "".join(chars), folder)
        references = {}
        for index, face in enumerate(faces):
            used = sorted(char for char in chars if assigned.get(char) == index)
            if used:
                references[index] = embed_face(document, page, face, used)
        fonts[key] = faces, assigned, advances, references

    rows = [[]]
    for index, char in enumerate(content):
        if char == "\n":
            rows.append([])
            continue
        style = letters[index]
        advances = fonts[style[0]][2]
        for letter in ("    " if char == "\t" else char):
            letter = " " if letter.isspace() else letter
            rows[-1].append((letter, style, advances[letter]*style[1]+style[2]))
    lines = list(rich_lines(rows, width, block.get("wrap", False)))
    if len(lines) > 256:
        raise TextError("Zmenšete počet řádků v bloku.")
    ascent = numeric(block.get("baseline_offset", size*.9), 0, 1200, "Vyberte platné umístění textového bloku.")
    original = block.get("original_lines", [])
    if len(original) != len(lines) or block.get("wrap"):
        original = []
    for offset in original:
        if not isinstance(offset, list) or len(offset) != 2:
            raise TextError("Textový blok obsahuje neplatné umístění řádku.")
        numeric(offset[0], -20000, 20000, "Textový blok obsahuje neplatné umístění řádku.")
        numeric(offset[1], 0, 20000, "Textový blok obsahuje neplatné umístění řádku.")

    matrix = {0: (1, 0, 0, 1), 90: (0, 1, -1, 0), 180: (-1, 0, 0, -1), 270: (0, -1, 1, 0)}[angle]
    marked = pdf.Dictionary(PBTextBlock=group, PBWidth=round(width*1000), PBFirstX=round(origin[0]*1000),
                            PBFirstY=round(origin[1]*1000), PBAngle=angle,
                            PBAlign=("left", "center", "right", "justify").index(alignment),
                            PBLeading=round(leading*1000))
    result = [instruction("q"), instruction("cm", *matrix, *origin), instruction("BDC", pdf.Name.Div, marked)]
    stroke = {"g": "G", "rg": "RG", "k": "K", "cs": "CS", "sc": "SC", "scn": "SCN"}
    first_ink = ranges[0]["color"]
    baseline = ascent
    for row, (line, paragraph, wrapped) in enumerate(lines):
        line_size = max((style[1] for _, style, _ in line), default=size)
        indent = 0.
        if original:
            indent, baseline = original[row]
        if baseline + line_size*.22 > height:
            raise TextError("Text se nevejde na stránku. Zmenšete řádkování nebo písmo.")
        if not line:
            baseline += line_size*leading
            continue
        extent = sum(advance for _, _, advance in line) - line[-1][1][2]
        left = indent
        if alignment == "center":
            left = max(0., (width-extent)/2)
        elif alignment == "right":
            left = max(0., width-extent)
        spaces = sum(char == " " for char, _, _ in line)
        extra = max(0., width-extent)/spaces if spaces and alignment == "justify" and wrapped else 0.
        x = left
        for style, spans in groupby(line, key=lambda glyph: glyph[1]):
            key, span_size, tracking, underlined, ink = style
            _, bold, italic = key
            faces, assigned, advances, references = fonts[key]
            span = list(spans)
            paint = palette.get(ink, color if ink == first_ink else (
                pdf.ContentStreamInstruction(*instruction("rg", *((ink >> shift & 255)/255 for shift in (0, 8, 16)))),
            ))
            for index, letters in groupby(span, key=lambda glyph: assigned[glyph[0]]):
                piece = "".join(glyph[0] for glyph in letters)
                face = faces[index]
                name, cids = references[index]
                row_mark = pdf.Dictionary(PBRow=row, PBParagraph=paragraph, PBWordWrap=int(wrapped),
                                          PBBold=int(bold), PBItalic=int(italic), PBTracking=round(tracking*1000))
                length = sum(advances[char]*span_size+tracking for char in piece) + extra*piece.count(" ")
                result.extend((instruction("BDC", pdf.Name.Span, row_mark), instruction("q"), *paint))
                if face.fake_bold:
                    result.extend(instruction(stroke[str(op.operator)], *op.operands)
                                  for op in paint if str(op.operator) in stroke)
                    result.append(instruction("w", span_size*.035))
                result.extend((instruction("BT"), instruction("Tr", 2 if face.fake_bold else 0),
                               instruction("Tc", tracking), instruction("Tf", name, span_size),
                               instruction("Tm", 1, 0, .21 if face.fake_italic else 0, 1, x, -baseline),
                               encoded_fragment(piece, cids, extra, span_size), instruction("ET")))
                if underlined:
                    result.extend((instruction("BMC", UNDERLINE),
                                   instruction("re", x, -baseline-span_size*.13,
                                               max(0., length-tracking), span_size*.055),
                                   instruction("f"), instruction("EMC")))
                result.extend((instruction("Q"), instruction("EMC")))
                x += length
        baseline += line_size*leading
    return [*result, instruction("EMC"), instruction("Q")]


def block_commands(document, page, block, style, folder, color, group, palette=None):
    content = block.get("text")
    if not isinstance(content, str) or len(content) > 8000 or any(
            ord(char) > 0xffff or (ord(char) < 32 and char not in "\t\n") for char in content):
        raise TextError("Blok může obsahovat nejvýš 8 000 běžných znaků.")
    if not content.strip():
        return []
    size = numeric(style.get("size"), 1, 600, "Zadejte velikost písma od 1 do 600 bodů.")
    width = numeric(block.get("width"), 1, 20000, "Zadejte platnou šířku textového bloku.")
    height = numeric(block.get("height"), 1, 20000, "Vyberte platné umístění textového bloku.")
    line_spacing = numeric(block.get("leading"), .7, 5, "Řádkování musí být v rozsahu 0,7 až 5.")
    tracking = numeric(block.get("tracking"), -72, 72, "Zadejte menší záporný rozestup nebo rozestup nejvýš 72 bodů.")
    alignment, origin, angle = block.get("align"), block.get("origin"), block.get("rotation")
    if (alignment not in ("left", "center", "right", "justify") or angle not in (0, 90, 180, 270)
            or not isinstance(origin, list) or len(origin) != 2
            or not all(type(value) in (float, int) and isfinite(value) and abs(value) <= 2e6 for value in origin)):
        raise TextError("Vyberte platné umístění a zarovnání textového bloku.")

    if block.get("runs") is not None:
        return rich_block_commands(document, page, block, content, folder, size, width, height,
                                   line_spacing, alignment, origin, angle, color, group, palette or {})

    faces, assigned, advances = fonts_for_block(style, content, folder)
    ascent = numeric(block.get("baseline_offset", faces[0].ascent*size), 0, 1200,
                     "Vyberte platné umístění textového bloku.")
    lines = list(wrap_block(content, advances, size, width, tracking))
    if ascent + (len(lines)-1)*line_spacing*size + size*.22 > height or len(lines) > 256:
        raise TextError("Text se nevejde na stránku. Zvětšete šířku bloku nebo zmenšete řádkování či písmo.")
    used = {char for value, _, _, _ in lines for char in value}
    matrix = {0: (1, 0, 0, 1), 90: (0, 1, -1, 0), 180: (-1, 0, 0, -1),
              270: (0, -1, 1, 0)}[angle]
    marked = pdf.Dictionary(PBTextBlock=group, PBWidth=round(width*1000), PBFirstX=round(origin[0]*1000),
                            PBFirstY=round(origin[1]*1000), PBAngle=angle,
                            PBAlign=("left", "center", "right", "justify").index(alignment),
                            PBLeading=round(line_spacing*1000), PBTracking=round(tracking*1000),
                            PBBold=int(bool(style.get("bold"))), PBItalic=int(bool(style.get("italic"))))
    result = [instruction("q"), *color, instruction("cm", *matrix, *origin),
              instruction("BDC", pdf.Name.Div, marked)]
    resources = pdf.Dictionary(page.resources)
    resources.Font = pdf.Dictionary(resources.get("/Font", pdf.Dictionary()))
    page.obj.Resources = resources
    references = {}
    for index, face in enumerate(faces):
        chars = sorted(char for char in used if assigned[char] == index)
        if chars:
            references[index] = embed_face(document, page, face, chars)
    stroke_ops = {"g":"G", "rg":"RG", "k":"K", "cs":"CS", "sc":"SC", "scn":"SCN"}
    result.extend(instruction(stroke_ops[str(item.operator)], *item.operands)
                  for item in color if str(item.operator) in stroke_ops)
    result.append(instruction("w", size*.035))
    for row, (value, advance, justification, paragraph) in enumerate(lines):
        if not value:
            continue
        left = 0.
        if alignment == "right":
            left = max(0., width-advance)
        elif alignment == "center":
            left = max(0., (width-advance)/2)
        justify = alignment == "justify" and justification and value.count(" ") > 0
        extra = max(0., width-advance) / value.count(" ") if justify else 0.
        baseline = -(ascent+row*line_spacing*size)
        row_mark = pdf.Dictionary(PBRow=row, PBParagraph=paragraph, PBWordWrap=int(justification))
        result.extend((instruction("BDC", pdf.Name.Span, row_mark), instruction("BT"),
                       instruction("Tc", tracking), instruction("Tw", 0), instruction("Tz", 100), instruction("Ts", 0)))
        x = left
        for index, group_chars in groupby(value, key=assigned.__getitem__):
            piece = "".join(group_chars)
            face = faces[index]
            key, cids = references[index]
            result.extend((instruction("Tr", 2 if face.fake_bold else 0), instruction("Tf", key, size),
                           instruction("Tm", 1, 0, .21 if face.fake_italic else 0, 1, x, baseline),
                           encoded_fragment(piece, cids, extra, size)))
            x += sum(advances[char]*size+tracking for char in piece) + extra*piece.count(" ")
        result.append(instruction("ET"))
        if style.get("underline"):
            result.extend((instruction("BMC", UNDERLINE),
                           instruction("re", left, baseline-size*.13, width if justify else advance, size*.055),
                           instruction("f"), instruction("EMC")))
        result.append(instruction("EMC"))
    return [*result, instruction("EMC"), instruction("Q")]


def remove_glyphs(entry, step, advances, chosen):
    output, count = [], 0
    remaining = bytearray()
    shift = 0.

    def flush():
        nonlocal shift
        if remaining:
            output.append(pdf.String(bytes(remaining)))
            remaining.clear()
        if shift:
            output.append(-shift)
            shift = 0.

    for value in show_items(entry):
        if not isinstance(value, pdf.String):
            flush()
            output.append(value)
            continue
        data = bytes(value)
        for start in range(0, len(data), step):
            if count in chosen:
                if remaining:
                    flush()
                shift += advances[count]
            else:
                if shift:
                    flush()
                remaining.extend(data[start:start+step])
            count += 1
    flush()
    return instruction("TJ", pdf.Array(output))


def strip_selection(document, page, selected, *, palette=None):
    from text_edit import source_objects, text_tree

    _, sources = text_tree(page)
    objects = source_objects(page, sources)
    if not isinstance(selected, list) or not 1 <= len(selected) <= 4000:
        raise TextError("Vyberte platnou oblast textu.")
    indexed = {run["block_token"]: run for run in objects if run["block_token"]}
    replacements, discard = {}, set()
    glyph_count = 0
    for item in selected:
        token, glyphs = item.get("token"), item.get("glyphs")
        if type(token) is not int or token in replacements or token not in indexed or not isinstance(glyphs, list):
            raise TextError("Tento úsek nelze upravit bez změny okolní sazby. Vyberte jiný úsek.")
        run = indexed[token]
        operators, index, style, independent, _, root = sources[token]
        paint = style[4]
        if not root and any(str(op.operator) == "cs" and op.operands[0] not in (
                pdf.Name.DeviceGray, pdf.Name.DeviceRGB, pdf.Name.DeviceCMYK) for op in paint):
            paint = (pdf.ContentStreamInstruction(*instruction("rg", *(value/255 for value in run["fill"]))),)
        if palette is not None:
            ink = sum(channel << (8 * i) for i, channel in enumerate(run["fill"]))
            palette.setdefault(ink, paint)
        if not replacements:
            first_color = paint
        advances = run["advances"]
        count = len(run["glyphs"])
        if not glyphs or any(type(index) is not int or not 0 <= index < count for index in glyphs):
            raise TextError("Výběr textu se mezitím změnil. Vyberte jej znovu.")
        chosen = set(glyphs)
        glyph_count += len(chosen)
        if glyph_count > 8000:
            raise TextError("Pro jednu úpravu vyberte nejvýš 8 000 znaků.")
        if advances:
            replacements[token] = remove_glyphs(operators[index-1], code_size(style[0]), advances, chosen)
        elif independent and len(chosen) == count:
            replacements[token] = instruction("TJ", pdf.Array())
        else:
            raise TextError("U tohoto kódování písma vyberte celý textový objekt.")
        if len(chosen) == count and has_underline(operators, index):
            discard.add(token)

    original, _ = text_tree(page, document, replacements=replacements, discard=discard)
    group = max((run["block"]["id"] for run in objects if run["block"]), default=0) + 1
    return original, first_color, group


def replace_block(document, page, edits, font_dir):
    block = edits["block"]
    palette = {}
    original, color, group = strip_selection(document, page, block.get("selection"), palette=palette)
    if group > 2**31 - 1:
        raise TextError("Na stránce nelze vytvořit další editovatelný blok.")
    commands = block_commands(document, page, block, edits, font_dir, color, group, palette)
    combined = [instruction("q"), *original, instruction("Q"), *commands] if commands else original
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(combined))
