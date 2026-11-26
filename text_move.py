"""Translate native PDF text using the existing font and text state."""

from itertools import groupby
from math import isfinite

import pikepdf as pdf

from object_edit import ObjectError, clip_point, selected_draws, translated_draw
from text_edit import TextError, code_size, instruction, invert, show_items, source_objects, text_tree, underline_marks


def moved_show(entry, selected_codes, step, x_adjustment, rise, vertical):
    """Apply and compensate the displacement without changing the text advance."""
    def pieces():
        index = 0
        for value in show_items(entry):
            if selected_codes is None:
                yield True, value
            elif isinstance(value, pdf.String):
                data = bytes(value)
                size = len(data) // step
                for selected, indices in groupby(range(size), lambda i: index+i in selected_codes):
                    indices = list(indices)
                    yield selected, pdf.String(data[step*indices[0]:step*(indices[-1]+1)])
                index += size
            else:
                yield False, value

    commands = []
    for selected, group in groupby(pieces(), key=lambda item: item[0]):
        values = [value for _, value in group]
        if selected:
            if vertical:
                commands.append(instruction("Ts", rise+vertical))
            values = [-x_adjustment, *values, x_adjustment]
        commands.append(instruction("TJ", pdf.Array(values)))
        if selected and vertical:
            commands.append(instruction("Ts", rise))
    return commands


def move_selection(document, page, operators, candidates, edit, selected):
    delta = clip_point(edit.get("delta"), "posun")
    items = selected_draws(candidates, edit) if edit.get("tokens") else []
    graphics = {}
    for item in items:
        for index in range(item.first+1, item.end+1):
            graphics[index] = []
        graphics[item.first+1] = translated_draw(operators, item, delta)

    context = {}
    _, sources = text_tree(page, move_context=context)
    runs = source_objects(page, sources)
    indexed = {run["block_token"]: run for run in runs if run["block_token"]}
    if not isinstance(selected, list) or not 1 <= len(selected) <= 4000:
        raise ObjectError("Vyberte platný text k přesunu.")
    translations, moved_glyphs, blocks, underlines = {}, {}, {}, {}
    for item in selected:
        if not isinstance(item, dict):
            raise ObjectError("Neplatný výběr textu.")
        token, glyphs = item.get("token"), item.get("glyphs")
        if type(token) is not int or token not in indexed or token in moved_glyphs or not isinstance(glyphs, list):
            raise ObjectError("Tento text nelze bezpečně přesunout. Výběr zůstal na původním místě.")
        run = indexed[token]
        count = len(run["glyphs"])
        if not glyphs or any(type(index) is not int or not 0 <= index < count for index in glyphs):
            raise ObjectError("Výběr textu se mezitím změnil. Vyberte jej znovu.")
        chosen = set(glyphs)
        matrix, (scale, rise), root, graphic_matrix = context[token]
        moved_glyphs[token] = len(chosen)
        if root in graphics:
            continue  # The containing Form is already translated as a whole.
        source, index, style, _, _, _ = sources[token]
        font = style[0]
        if font is None:
            raise ObjectError("Původní písmo textu se nepodařilo najít.")
        if font.get("/Subtype") == pdf.Name.Type0:
            encoding = font.get("/Encoding")
            if str(encoding).endswith("-V") or isinstance(encoding, pdf.Stream) and encoding.get("/WMode", 0) != 0:
                raise ObjectError("Tento svisle sázený text nelze bezpečně přesunout.")
        step = code_size(font)
        raw_codes = None
        if len(chosen) != count:
            if not step or not run["glyph_indices"]:
                raise ObjectError("U tohoto kódování písma vyberte celý textový objekt.")
            raw_codes = {run["glyph_indices"][i] for i in chosen}
        try:
            underline = underline_marks(source, index)
            if underline and (root or underline.start+1 not in graphics):
                if len(chosen) != count:
                    raise ObjectError("Pro přesun vyberte celý podtržený úsek textu.")
                a, b, c, d, _, _ = invert(graphic_matrix)
                underlines[token] = a*delta[0]+c*delta[1], b*delta[0]+d*delta[1]
            a, b, c, d, _, _ = invert(matrix)
        except TextError as exc:
            raise ObjectError(str(exc)) from exc
        horizontal, vertical = a*delta[0]+c*delta[1], b*delta[0]+d*delta[1]
        size = style[1]*scale/100
        if not size or not all(map(isfinite, (size, rise, horizontal, vertical))):
            raise ObjectError("Text má neplatnou transformaci.")
        translations[token] = moved_show(source[index-1], raw_codes, step, horizontal*1000/size, rise, vertical)

    for run in runs:
        token, block = run["block_token"], run["block"]
        if not block or not token:
            continue
        root = context[token][2]
        key = root, block["id"]
        if root in graphics:
            continue
        if token in translations:
            blocks.setdefault(key, delta)
    for run in runs:
        token, block = run["block_token"], run["block"]
        if block and token and (key := (context[token][2], block["id"])) in blocks:
            if moved_glyphs.get(token, 0) != len(run["glyphs"]):
                blocks[key] = None  # A split block will be inferred from its new layout.

    content, _ = text_tree(page, document, translations=translations, graphics=graphics, blocks=blocks, underlines=underlines)
    page.obj.Contents = document.make_stream(pdf.unparse_content_stream(content))
