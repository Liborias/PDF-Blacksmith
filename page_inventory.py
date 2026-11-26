"""Read paint occurrences and optional content from the active PDF page."""

from itertools import islice

import pikepdf as pdf

from object_edit import PAINT, draws, has_xfa, movable_widgets
from text_edit import pikepdf_operators


MAX_ITEMS = 4000
MAX_OPERATORS = 180_000
MAX_DEPTH = 8
MAX_LAYERS = 256
MAX_MEMBERSHIPS = 16


def dictionary(value):
    return value if isinstance(value, pdf.Dictionary) else {}


def array(value):
    return value if isinstance(value, pdf.Array) else ()


def key(obj):
    if isinstance(obj, (pdf.Dictionary, pdf.Stream)) and obj.is_indirect:
        return "%d:%d" % obj.objgen
    return ""


def label(value):
    if not isinstance(value, (pdf.String, pdf.Name)):
        return ""
    return " ".join(str(value).split())[:160]


def declared_layers(document):
    properties = dictionary(document.Root.get("/OCProperties"))
    config = dictionary(properties.get("/D"))
    base = {"/ON": "on", "/OFF": "off"}.get(str(config.get("/BaseState", "/ON")), "")
    layers = {}
    limited = False
    for group in array(properties.get("/OCGs")):
        identity = key(group)
        if not isinstance(group, pdf.Dictionary) or not identity or identity in layers or group.get("/Type") != pdf.Name.OCG:
            continue
        if len(layers) >= MAX_LAYERS:
            limited = True
            break
        layers[identity] = dict(id=identity, name=label(group.get("/Name")), state=base, locked=False)
    for field, name, value in (("/ON", "state", "on"), ("/OFF", "state", "off"), ("/Locked", "locked", True)):
        for group in array(config.get(field)):
            if (identity := key(group)) in layers:
                layers[identity][name] = value
    return layers, limited


def optional_layers(entry, layers):
    if not layers:
        return ()
    identity = key(entry)
    if identity in layers:
        return (identity,)
    entry = dictionary(entry)
    if entry.get("/Type") != pdf.Name.OCMD:
        return ()
    groups = entry.get("/OCGs")
    candidates = [groups] if isinstance(groups, pdf.Dictionary) else list(islice(array(groups), MAX_LAYERS))
    # A visibility expression can mention groups outside /OCGs. Report
    # membership separately from the activation rule or effective visibility.
    expressions = [(entry.get("/VE"), 0)]
    while expressions:
        expression, depth = expressions.pop()
        if isinstance(expression, pdf.Array) and depth < MAX_DEPTH:
            expressions.extend((child, depth+1) for child in islice(expression, MAX_LAYERS))
        elif isinstance(expression, pdf.Dictionary):
            candidates.append(expression)
        if len(candidates) + len(expressions) > MAX_LAYERS * 2:
            break
    return tuple(dict.fromkeys(group for candidate in candidates if (group := key(candidate)) in layers))


def inspect_page(document, page):
    layers, layers_limited = declared_layers(document)
    remaining_items, remaining_ops = MAX_ITEMS, MAX_OPERATORS
    limited = layers_limited
    cached_forms = {}

    def add_layers(membership, entry):
        nonlocal limited
        merged = tuple(dict.fromkeys((*membership, *optional_layers(entry, layers))))
        if len(merged) > MAX_MEMBERSHIPS:
            limited = True
        return merged[:MAX_MEMBERSHIPS]

    def node(target, parent, index, kind, *, name="", membership=(), object_token=0, text_token=0):
        nonlocal remaining_items
        identity = f"{parent}/{index}" if parent else str(index)
        item = dict(id=identity, kind=kind)
        if name:
            item["name"] = name
        if membership:
            item["layers"] = membership
        if object_token:
            item["object"] = object_token
        if text_token:
            item["text"] = text_token
        remaining_items -= 1
        target.append(item)
        return item

    def visit(operators, resources, target, parent, ancestors, inherited, render_mode=0):
        nonlocal limited, remaining_ops
        xobjects = dictionary(resources.get("/XObject"))
        properties = dictionary(resources.get("/Properties"))
        # Root tokens use the same normalized operator offsets as existing
        # move/copy. Children of a Form are listed by occurrence in this pass.
        root = not ancestors
        candidates = {}
        crops = {}
        if root and len(operators) <= remaining_ops:
            try:
                for draw in draws(page, operators):
                    candidates[draw.token] = draw
                    if draw.mask:
                        crops[draw.end-2] = draw
            except (ArithmeticError, TypeError, ValueError, AttributeError, IndexError):
                # Malformed graphics can still be inventoried without edit tokens.
                candidates.clear()
                crops.clear()

        membership = inherited
        marked, modes = [], []
        for index, operation in enumerate(operators, 1):
            if remaining_items <= 0 or remaining_ops <= 0:
                limited = True
                break
            remaining_ops -= 1
            operator, args = str(operation.operator), operation.operands
            if operator in ("BMC", "BDC"):
                marked.append(membership)
                if operator == "BDC" and len(args) >= 2 and args[0] == pdf.Name.OC:
                    group = properties.get(str(args[1])) if isinstance(args[1], pdf.Name) else args[1]
                    membership = add_layers(membership, group)
            elif operator == "EMC" and marked:
                membership = marked.pop()
            elif operator == "q":
                modes.append(render_mode)
            elif operator == "Q" and modes:
                render_mode = modes.pop()
            elif operator == "Tr" and len(args) == 1 and isinstance(args[0], int):
                render_mode = args[0]
            elif operator in ("Tj", "TJ") and len(args) == 1:
                hidden = render_mode in (3, 7)
                node(target, parent, index, "hidden_text" if hidden else "text", membership=membership,
                     text_token=index if root and not hidden else 0)
            elif operator in PAINT:
                node(target, parent, index, "path", membership=membership,
                     object_token=index if index in candidates else 0)
            elif operator == "sh":
                node(target, parent, index, "shading", membership=membership)
            elif operator == "INLINE IMAGE":
                node(target, parent, index, "image", membership=membership)
            elif operator == "Do" and len(args) == 1:
                xobject = xobjects.get(str(args[0]))
                if not isinstance(xobject, pdf.Stream):
                    continue
                subtype = xobject.get("/Subtype")
                if subtype not in (pdf.Name.Image, pdf.Name.Form):
                    continue
                groups = add_layers(membership, xobject.get("/OC"))
                kind = "image" if subtype == pdf.Name.Image else "form"
                if kind == "form" and "/Group" in xobject:
                    kind = "group"
                token = crops[index].token if index in crops else index if index in candidates else 0
                item = node(target, parent, index, kind, membership=groups, object_token=token)
                if subtype == pdf.Name.Form:
                    identity = key(xobject)
                    if (identity in ancestors or len(ancestors) >= MAX_DEPTH or "/Ref" in xobject
                            or remaining_ops <= 0 or remaining_items <= 0):
                        limited = True
                        continue
                    children = []
                    try:
                        child_resources = xobject.get("/Resources")
                        if not isinstance(child_resources, pdf.Dictionary):
                            child_resources = resources
                        child_ops = cached_forms.get(identity)
                        if child_ops is None:
                            child_ops = pikepdf_operators(xobject)
                            if identity and len(cached_forms) < 32 and len(child_ops) < 2000:
                                cached_forms[identity] = child_ops
                        visit(child_ops, child_resources,
                              children, item["id"], (*ancestors, identity), groups, render_mode)
                    except (pdf.PdfError, TypeError, ValueError):
                        limited = True
                    if children:
                        item["children"] = children

    items = []
    visit(pikepdf_operators(page), page.resources, items, "", (), ())

    # Widgets are drawn separately by the PDF viewer, above page content.
    movable = set() if has_xfa(document) else {token for token, _, _ in movable_widgets(page)}
    for index, annot in enumerate(array(page.obj.get("/Annots")), 1):
        if not isinstance(annot, pdf.Dictionary) or annot.get("/Subtype") != pdf.Name.Widget:
            continue
        if remaining_items <= 0:
            limited = True
            break
        node(items, "", f"widget/{index}", "widget", name=label(annot.get("/T")),
             membership=add_layers((), annot.get("/OC")), object_token=-index if -index in movable else 0)
    return dict(items=items, layers=list(layers.values()), limited=limited)
