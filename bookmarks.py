"""Read PDF navigation and apply bookmark edits to the original outline objects."""

import pikepdf


class BookmarkError(ValueError):
    pass


def outline_items(pdf):
    root = pdf.Root.get("/Outlines")
    if not isinstance(root, pikepdf.Dictionary):
        return
    pending = [(root.get("/First"), 0)]
    seen = set()
    while pending:
        item, depth = pending.pop()
        if not isinstance(item, pikepdf.Dictionary):
            continue
        if item.is_indirect:
            if item.objgen in seen:
                continue
            seen.add(item.objgen)
        yield item, depth
        pending.append((item.get("/Next"), depth))
        pending.append((item.get("/First"), depth + 1))


def bookmark_destination(pdf, item):
    target = item.get("/Dest")
    if target is None:
        action = item.get("/A")
        if isinstance(action, pikepdf.Dictionary) and action.get("/S") == pikepdf.Name.GoTo:
            target = action.get("/D")
    if isinstance(target, pikepdf.Name):
        legacy = pdf.Root.get("/Dests")
        target = legacy.get(str(target)) if isinstance(legacy, pikepdf.Dictionary) else None
    elif isinstance(target, pikepdf.String):
        names = pdf.Root.get("/Names")
        tree = names.get("/Dests") if isinstance(names, pikepdf.Dictionary) else None
        try:
            target = pikepdf.NameTree(tree).get(str(target)) if isinstance(tree, pikepdf.Dictionary) else None
        except (pikepdf.PdfError, ValueError, RuntimeError):
            return None
    if isinstance(target, pikepdf.Dictionary):
        target = target.get("/D")
    if isinstance(target, pikepdf.Array) and len(target):
        return target[0]
    return None


def describe_bookmarks(pdf):
    items = list(outline_items(pdf))
    if not items:
        return []
    pages = {page.obj.objgen: number for number, page in enumerate(pdf.pages, 1)}
    result = []
    for item, depth in items:
        target = bookmark_destination(pdf, item)
        if isinstance(target, pikepdf.Dictionary) and target.is_indirect:
            page = pages.get(target.objgen, 0)
        elif type(target) is int and 0 <= target < len(pages):
            page = target + 1
        else:
            page = 0
        title = item.get("/Title")
        result.append(dict(title=str(title) if isinstance(title, pikepdf.String) else "", page=page, depth=depth))
    return result


def append_bookmarks(pdf, top_level, additions):
    root = pdf.Root.get("/Outlines")
    if root is None:
        root = pdf.make_indirect(pikepdf.Dictionary(Type=pikepdf.Name.Outlines))
        pdf.Root.Outlines = root
    elif not isinstance(root, pikepdf.Dictionary):
        raise BookmarkError("Strom záložek má neplatnou strukturu.")
    elif not root.is_indirect:
        root = pdf.make_indirect(root)
        pdf.Root.Outlines = root
    last = top_level[-1] if top_level else None
    if (last is None and root.get("/First") is not None) or (last is not None and last.get("/Next") is not None):
        raise BookmarkError("Strom záložek má neplatnou strukturu.")
    visible = len(top_level)
    for item in top_level:
        count = item.get("/Count", 0)
        if type(count) is int:
            visible += max(0, count)
    for title, page in additions:
        node = pdf.make_indirect(pikepdf.Dictionary(Title=title, Parent=root,
            Dest=pikepdf.Array([pdf.pages[page-1].obj, pikepdf.Name.Fit])))
        if last is None:
            root.First = node
        else:
            last.Next = node
            node.Prev = last
        last = node
    root.Last, root.Count = last, visible + len(additions)


def apply_bookmarks(pdf, changes):
    if not isinstance(changes, list):
        raise BookmarkError("Neplatný seznam úprav záložek.")
    if not changes:
        return
    items = list(outline_items(pdf))
    updates, additions = {}, []
    for change in changes:
        if not isinstance(change, dict):
            raise BookmarkError("Neplatná úprava záložky.")
        index, title, page = change.get("index"), change.get("title"), change.get("page", 0)
        if (type(index) is not int or not 0 <= index <= len(items) or index in updates
                or not isinstance(title, str) or not title.strip() or "\x00" in title
                or type(page) is not int or not 0 <= page <= len(pdf.pages)
                or (index == 0 and page == 0)):
            raise BookmarkError("Neplatný název nebo cílová strana záložky.")
        if index:
            updates[index] = (title, page)
        else:
            additions.append((title, page))
    for index, (title, page) in updates.items():
        item, _ = items[index-1]
        item.Title = title
        if page:
            item.Dest = pikepdf.Array([pdf.pages[page-1].obj, pikepdf.Name.Fit])
            if "/A" in item:
                del item["/A"]
    if additions:
        append_bookmarks(pdf, [item for item, depth in items if depth == 0], additions)
