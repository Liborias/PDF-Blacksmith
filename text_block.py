"""Reflow a selected range and write searchable text into its page."""

from io import BytesIO
from math import hypot, isfinite
from pathlib import Path
import re
import sys

from fontTools.ttLib import TTFont, TTLibError
import pikepdf as pdf

from text_edit import (FAMILIES, TextError, UNDERLINE, code_size, encoded_runs, has_underline,
                       instruction, pikepdf_operators, show_items, source_styles, text_page)


def paragraph_font(style, folder):
    family = style.get("font", "")
    if not isinstance(family, str) or len(family) > 63 or "\0" in family:
        raise TextError("Vyberte dostupné písmo.")
    bundled = ("Liberation Sans", "Liberation Serif", "Liberation Mono")
    try:
        if not family or family in bundled:
            index = bundled.index(family) if family else style.get("family", 0)
            if type(index) is not int or not 0 <= index < len(bundled):
                raise TextError("Vyberte dostupné písmo.")
            face = ("Bold" if style.get("bold") else "") + ("Italic" if style.get("italic") else "") or "Regular"
            data = (Path(folder) / f"{FAMILIES[index]}-{face}.ttf").read_bytes()
        elif sys.platform == "win32":
            from system_fonts import load_system_font
            data = load_system_font(family, bool(style.get("bold")), bool(style.get("italic")))
        else:
            raise TextError("Toto systémové písmo je dostupné ve Windows.")
        with TTFont(BytesIO(data)) as font:
            if "glyf" not in font:
                raise TextError("Pro úpravu zvolte písmo s křivkami TrueType.")
            permissions = font["OS/2"].fsType if "OS/2" in font else 0
            if permissions & 0x202 or (permissions & 4 and not permissions & 8):
                raise TextError("Zvolené písmo neumožňuje vložení pro úpravy. Vyberte jiné písmo.")
            if "OS/2" in font and family not in (*bundled, ""):
                flags = font["OS/2"].fsSelection
                if bool(flags & 0x20) != bool(style.get("bold")) or bool(flags & 0x201) != bool(style.get("italic")):
                    raise TextError("Zvolené písmo nemá požadovaný řez.")
            units = font["head"].unitsPerEm
            mapping = font.getBestCmap() or {}
            metrics = font["hmtx"].metrics
            advances = {chr(code): metrics[name][0] / units for code, name in mapping.items()}
            return data, advances, max(.7, min(1.2, font["hhea"].ascent / units))
    except (OSError, TTLibError, KeyError, ValueError) as exc:
        # TextError is itself a ValueError and carries the actionable explanation.
        if isinstance(exc, (TextError, ValueError)) and not isinstance(exc, TTLibError):
            raise TextError(str(exc)) from exc
        raise TextError("Zvolené písmo se nepodařilo načíst.") from exc


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


def encoded_line(show, value, extra, size):
    if not extra:
        return show
    operands = show_items(show)
    if len(operands) != 1 or not isinstance(operands[0], pdf.String):
        raise TextError("Zvolené písmo nepodporuje zarovnání do bloku.")
    codes = bytes(operands[0])
    if len(codes) != 2*len(value):
        raise TextError("Zvolené písmo nepodporuje zarovnání do bloku.")
    items, start = [], 0
    for index, char in enumerate(value, 1):
        if char == " ":
            items.extend((pdf.String(codes[start:2*index]), -1000 * extra / size))
            start = 2*index
    if start < len(codes):
        items.append(pdf.String(codes[start:]))
    return instruction("TJ", pdf.Array(items))


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

    data, advances, ascent = paragraph_font(style, folder)
    lines = list(wrap_block(content, advances, size, width, tracking))
    if ascent*size + (len(lines)-1)*line_spacing*size + size*.22 > height or len(lines) > 256:
        raise TextError("Text se nevejde na stránku. Zvětšete šířku bloku nebo zmenšete řádkování či písmo.")
    values = [value for value, _, _, _ in lines if value]
    encoded, _ = encoded_runs(values, {**style, "family": 0}, folder, font_data=data)
    matrix = {0: (1, 0, 0, 1), 90: (0, 1, -1, 0), 180: (-1, 0, 0, -1),
              270: (0, -1, 1, 0)}[angle]
    marked = pdf.Dictionary(PBTextBlock=group, PBWidth=round(width*1000), PBFirstX=round(origin[0]*1000),
                            PBFirstY=round(origin[1]*1000), PBAngle=angle,
                            PBAlign=("left", "center", "right", "justify").index(alignment),
                            PBLeading=round(line_spacing*1000), PBTracking=round(tracking*1000))
    result = [instruction("q"), *color, instruction("cm", *matrix, *origin),
              instruction("BDC", pdf.Name.Div, marked)]
    resources = pdf.Dictionary(page.resources)
    resources.Font = pdf.Dictionary(resources.get("/Font", pdf.Dictionary()))
    page.obj.Resources = resources
    with pdf.Pdf.open(BytesIO(encoded)) as templates:
        pages = iter(templates.pages)
        key = None
        for row, (value, advance, justification, paragraph) in enumerate(lines):
            if not value:
                continue
            template = next(pages)
            show = next((item for item in pikepdf_operators(template) if str(item.operator) in ("Tj", "TJ")), None)
            if show is None:
                raise TextError("Sazbu textového bloku se nepodařilo zakódovat.")
            if key is None:
                font = next(iter(template.resources.Font.values()))
                key = page.add_resource(document.copy_foreign(templates.make_indirect(font)), pdf.Name.Font, prefix="Text")
            left = 0.
            if alignment == "right":
                left = max(0., width-advance)
            elif alignment == "center":
                left = max(0., (width-advance)/2)
            justify = alignment == "justify" and justification and value.count(" ") > 0
            extra = max(0., width-advance) / value.count(" ") if justify else 0.
            baseline = -(ascent+row*line_spacing)*size
            row_mark = pdf.Dictionary(PBRow=row, PBParagraph=paragraph, PBWordWrap=int(justification))
            result.extend((instruction("BDC", pdf.Name.Span, row_mark),
                           instruction("BT"), instruction("Tr", 0), instruction("Tc", tracking),
                           instruction("Tw", 0), instruction("Tz", 100), instruction("Ts", 0),
                           instruction("Tf", key, size), instruction("Tm", 1, 0, 0, 1, left, baseline),
                           encoded_line(show, value, extra, size), instruction("ET")))
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


def replace_block(document, page, edits, font_dir):
    from text_edit import positions

    operators, _, objects = text_page(page)
    block = edits["block"]
    selected = block.get("selection")
    if not isinstance(selected, list) or not 1 <= len(selected) <= 4000:
        raise TextError("Vyberte platnou oblast textu.")
    styles = source_styles(operators, page.resources, positions(operators, page.resources, include_dependent=True))
    indexed = {run["block_token"]: run for run in objects if run["block_token"] and run["advances"]}
    replacements, discard = {}, set()
    glyph_count = 0
    for item in selected:
        token, glyphs = item.get("token"), item.get("glyphs")
        if type(token) is not int or token in replacements or token not in indexed or not isinstance(glyphs, list):
            raise TextError("Tento úsek nelze upravit bez změny okolní sazby. Vyberte jiný úsek.")
        run = indexed[token]
        advances = run["advances"]
        if not glyphs or any(type(index) is not int or not 0 <= index < len(advances) for index in glyphs):
            raise TextError("Výběr textu se mezitím změnil. Vyberte jej znovu.")
        chosen = set(glyphs)
        glyph_count += len(chosen)
        if glyph_count > 8000:
            raise TextError("Pro jednu úpravu vyberte nejvýš 8 000 znaků.")
        replacements[token] = remove_glyphs(operators[token-1], code_size(styles[token][0]), advances, chosen)
        if len(chosen) == len(advances) and has_underline(operators, token):
            discard.update(range(token+2, token+6))

    first = next(iter(replacements))
    group = max((run["block"]["id"] for run in objects if run["block"]), default=0) + 1
    if group > 2**31 - 1:
        raise TextError("Na stránce nelze vytvořit další editovatelný blok.")
    commands = block_commands(document, page, block, edits, font_dir, styles[first][4], group)
    original = [replacements.get(index, entry) for index, entry in enumerate(operators, 1) if index not in discard]
    combined = [instruction("q"), *original, instruction("Q"), *commands] if commands else original
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(combined))
