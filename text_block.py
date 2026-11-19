"""Reflow a selected range and write searchable text into its page."""

from itertools import groupby
from math import isfinite
import re

import pikepdf as pdf

from block_fonts import embed_face, encoded_fragment, fonts_for_block
from text_edit import TextError, UNDERLINE, code_size, has_underline, instruction, show_items, source_styles, text_page


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


def block_commands(document, page, block, style, folder, color, group):
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
    tracking = numeric(block.get("tracking"), -size*.18, 72, "Zadejte menší záporný rozestup nebo rozestup nejvýš 72 bodů.")
    alignment, origin, angle = block.get("align"), block.get("origin"), block.get("rotation")
    if (alignment not in ("left", "center", "right", "justify") or angle not in (0, 90, 180, 270)
            or not isinstance(origin, list) or len(origin) != 2
            or not all(type(value) in (float, int) and isfinite(value) and abs(value) <= 2e6 for value in origin)):
        raise TextError("Vyberte platné umístění a zarovnání textového bloku.")

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


def strip_selection(page, selected):
    from text_edit import positions

    operators, states, objects = text_page(page)
    if not isinstance(selected, list) or not 1 <= len(selected) <= 4000:
        raise TextError("Vyberte platnou oblast textu.")
    styles = source_styles(operators, page.resources, positions(operators, page.resources, include_dependent=True))
    indexed = {run["block_token"]: run for run in objects if run["block_token"]}
    replacements, discard = {}, set()
    glyph_count = 0
    for item in selected:
        token, glyphs = item.get("token"), item.get("glyphs")
        if type(token) is not int or token in replacements or token not in indexed or not isinstance(glyphs, list):
            raise TextError("Tento úsek nelze upravit bez změny okolní sazby. Vyberte jiný úsek.")
        run = indexed[token]
        advances = run["advances"]
        count = len(run["glyphs"])
        if not glyphs or any(type(index) is not int or not 0 <= index < count for index in glyphs):
            raise TextError("Výběr textu se mezitím změnil. Vyberte jej znovu.")
        chosen = set(glyphs)
        glyph_count += len(chosen)
        if glyph_count > 8000:
            raise TextError("Pro jednu úpravu vyberte nejvýš 8 000 znaků.")
        if advances:
            replacements[token] = remove_glyphs(operators[token-1], code_size(styles[token][0]), advances, chosen)
        elif token in states and len(chosen) == count:
            replacements[token] = instruction("TJ", pdf.Array())
        else:
            raise TextError("U tohoto kódování písma vyberte celý textový objekt.")
        if len(chosen) == count and has_underline(operators, token):
            discard.update(range(token+2, token+6))

    original = [replacements.get(index, entry) for index, entry in enumerate(operators, 1) if index not in discard]
    group = max((run["block"]["id"] for run in objects if run["block"]), default=0) + 1
    return original, styles[next(iter(replacements))][4], group


def replace_block(document, page, edits, font_dir):
    block = edits["block"]
    original, color, group = strip_selection(page, block.get("selection"))
    if group > 2**31 - 1:
        raise TextError("Na stránce nelze vytvořit další editovatelný blok.")
    commands = block_commands(document, page, block, edits, font_dir, color, group)
    combined = [instruction("q"), *original, instruction("Q"), *commands] if commands else original
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(combined))
