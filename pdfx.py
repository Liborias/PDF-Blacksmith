"""Prepare and check PDF/X files written by Ghostscript in process CMYK."""

from datetime import datetime, timezone
from io import BytesIO
from math import isfinite
from uuid import uuid4

import pikepdf as p
from PIL import ImageCms

from print_profiles import attach_intent, cmyk_icc


STANDARDS = {1: ("PDF/X-1a:2003", "1.4"), 3: ("PDF/X-3:2003", "1.4"), 4: ("PDF/X-4", "1.6")}
BOX_KEYS = ("/CropBox", "/TrimBox", "/ArtBox", "/BleedBox")
XMP_VERSION = "pdfxid:GTS_PDFXVersion"
XMP_IDENTIFIERS = ("xmpMM:DocumentID", "xmpMM:VersionID", "xmpMM:RenditionClass")
XMP_DATES = ("xmp:CreateDate", "xmp:ModifyDate", "xmp:MetadataDate")
INFO_XMP_FIELDS = ("/Title", "/Author", "/Subject", "/Keywords", "/Creator",
                   "/Producer", "/CreationDate", "/ModDate")


class PDFXError(ValueError):
    pass


def check_profile(profile, variant):
    data = profile[0]
    if (not cmyk_icc(data) or data[12:16] != b"prtr"
            or data[8] not in ((2, 4) if variant == 4 else (2,))):
        raise PDFXError("Zvolený výstupní ICC profil CMYK neodpovídá požadavkům předvolby PDF/X.")
    try:
        ImageCms.getOpenProfile(BytesIO(data))
    except ImageCms.PyCMSError as exc:
        raise PDFXError("Výstupní ICC profil CMYK nelze načíst.") from exc


def pdfx_prefix(folder, profile, variant):
    def text(value):
        return "<" + ("\ufeff" + value).encode("utf-16-be").hex() + ">"

    _, label, condition = profile
    prefix = (
        "%!PS\n"
        f"[ /GTS_PDFXVersion {text(STANDARDS[variant][0])} /Trapped /False /DOCINFO pdfmark\n"
        "[ /_objdef {BlacksmithICC} /type /stream /OBJ pdfmark\n"
        "[ {BlacksmithICC} << /N 4 >> /PUT pdfmark\n"
        "[ {BlacksmithICC} (cil.icc) (r) file /PUT pdfmark\n"
        "[ {Catalog} << /OutputIntents [ << /Type /OutputIntent /S /GTS_PDFX "
        f"/OutputConditionIdentifier {text(condition)} /OutputCondition {text(label)} "
        "/DestOutputProfile {BlacksmithICC} >> ] >> /PUT pdfmark\n"
    )
    (folder / "pdfx.ps").write_text(prefix, encoding="ascii")


def rectangle(value):
    try:
        a, b, c, d = map(float, value)
    except (TypeError, ValueError, OverflowError) as exc:
        raise PDFXError("Tiskový rámeček má neplatné souřadnice.") from exc
    if not all(isfinite(x) for x in (a, b, c, d)) or a == c or b == d:
        raise PDFXError("Tiskový rámeček má neplatné souřadnice.")
    return min(a, c), min(b, d), max(a, c), max(b, d)


def rotate_box(box, width, height, angle):
    left, bottom, right, top = box
    if angle == 90:
        return bottom, width-right, top, width-left
    if angle == 180:
        return width-right, height-top, width-left, height-bottom
    if angle == 270:
        return height-top, left, height-bottom, right
    return box


def page_geometry(page):
    media = rectangle(page.mediabox)
    unit = float(page.obj.get("/UserUnit", 1))
    angle = int(page.obj.get("/Rotate", 0)) % 360
    if not isfinite(unit) or unit <= 0 or angle % 90:
        raise PDFXError("Neplatné měřítko nebo otočení strany.")
    width = (media[2]-media[0])*unit
    height = (media[3]-media[1])*unit
    return media, unit, angle, width, height


def check_boxes(page, number):
    def inside(inner, outer):
        a, b, c, d = inner
        e, f, g, h = outer
        return e <= a and f <= b and c <= g and d <= h

    media = rectangle(page.mediabox)
    boxes = {key: rectangle(page.obj[key]) for key in BOX_KEYS if key in page.obj}
    trim = boxes.get("/TrimBox", boxes.get("/ArtBox"))
    crop = boxes.get("/CropBox", media)
    bleed = boxes.get("/BleedBox", crop)
    if (trim is None or ("/TrimBox" in boxes and "/ArtBox" in boxes)
            or not inside(trim, bleed) or not inside(bleed, crop) or not inside(crop, media)):
        raise PDFXError(f"Strana {number} má neplatné vnoření rámečků. Zkontrolujte ořez, čistý formát a spadávku.")


def restore_boxes(source, converted):
    for number, (old, new) in enumerate(zip(source.pages, converted.pages), 1):
        media, unit, angle, width, height = page_geometry(old)
        dst, dst_unit, dst_angle, dst_width, dst_height = page_geometry(new)
        visual = (height, width) if angle % 180 else (width, height)
        dst_visual = (dst_height, dst_width) if dst_angle % 180 else (dst_width, dst_height)
        if any(abs(a-b) > 0.05 for a, b in zip(visual, dst_visual)):
            raise PDFXError(f"Převod PDF/X změnil fyzický rozměr strany {number}.")
        boxes = {key: old.obj[key] for key in BOX_KEYS if key in old.obj}
        if "/TrimBox" in boxes:
            boxes.pop("/ArtBox", None)
        elif "/ArtBox" not in boxes:
            boxes["/TrimBox"] = old.cropbox
        for key in BOX_KEYS:
            if key in new.obj:
                del new.obj[key]
        for key, values in boxes.items():
            bounds = rectangle(values)
            local = tuple((v-media[i%2])*unit for i, v in enumerate(bounds))
            aligned = rotate_box(local, width, height, angle)
            scaled = tuple(v*dst_visual[i%2]/visual[i%2] for i, v in enumerate(aligned))
            dest = rotate_box(scaled, *dst_visual, (-dst_angle) % 360)
            new.obj[key] = p.Array([
                round(dst[i%2] + v/dst_unit, 8) for i, v in enumerate(dest)
            ])
        check_boxes(new, number)


def check_color(value):
    if isinstance(value, p.Array):
        kind = str(value[0])
        if kind == "/Indexed":
            check_color(value[1])
        elif kind == "/Pattern" and len(value) == 2:
            check_color(value[1])
        else:
            raise PDFXError(f"Výstup obsahuje neočekávaný barevný prostor {kind}.")
    elif str(value) not in ("/DeviceGray", "/DeviceCMYK", "/Pattern"):
        raise PDFXError(f"Výstup obsahuje neočekávaný barevný prostor {value}.")


def check_content(stream):
    for instruction in p.parse_content_stream(stream):
        operator = str(instruction.operator)
        if operator in {"rg", "RG", "INLINE IMAGE"}:
            raise PDFXError("Výstup obsahuje přímo vložený obraz nebo RGB operátor mimo kontrolovaný barevný výstup.")


def dictionaries(root):
    visited = set()
    pending = [root]
    while pending:
        obj = pending.pop()
        if not isinstance(obj, (p.Dictionary, p.Stream, p.Array)):
            continue
        if obj.is_indirect:
            if obj.objgen in visited:
                continue
            visited.add(obj.objgen)
        if isinstance(obj, p.Array):
            pending.extend(obj)
        else:
            yield obj
            pending.extend(value for _, value in obj.items())


def check_candidate(pdf, variant, profile):
    if pdf.is_encrypted or any(key in pdf.Root for key in
                               ("/AcroForm", "/OpenAction", "/AA", "/OCProperties", "/Collection")):
        raise PDFXError("Výstup obsahuje šifrování, interaktivní funkce nebo externí zdroje.")
    intents = pdf.Root.get("/OutputIntents", [])
    if (len(intents) != 1 or intents[0].get("/S") != p.Name.GTS_PDFX
            or str(intents[0].get("/OutputConditionIdentifier")) != profile[2]
            or intents[0].DestOutputProfile.N != 4
            or intents[0].DestOutputProfile.read_bytes() != profile[0]):
        raise PDFXError("V PDF/X chybí očekávaný výstupní záměr ICC.")
    for number, page in enumerate(pdf.pages, 1):
        check_boxes(page, number)
        if page.obj.get("/Annots"):
            raise PDFXError(f"Strana {number} obsahuje nepodporované živé anotace.")
        if "/Contents" in page.obj:
            check_content(page)
    for obj in dictionaries(pdf.Root):
        if any(key in obj for key in ("/OPI", "/Ref", "/Alternates", "/FFilter", "/FDecodeParms",
                                      "/EmbeddedFiles", "/EF", "/JavaScript", "/AA")) or (
                isinstance(obj, p.Stream) and "/F" in obj):
            raise PDFXError("Výstup obsahuje externí reference nebo alternativní data.")
        if obj.get("/Type") == p.Name.Font:
            subtype = obj.get("/Subtype")
            if subtype == p.Name.Type3:
                for glyph in obj.CharProcs.values():
                    check_content(glyph)
            elif subtype == p.Name.Type0:
                if not obj.get("/DescendantFonts"):
                    raise PDFXError("Kompozitní písmo nemá vložený popis.")
            else:
                descriptor = obj.get("/FontDescriptor", {})
                if not any(isinstance(descriptor.get(key), p.Stream)
                           for key in ("/FontFile", "/FontFile2", "/FontFile3")):
                    raise PDFXError(f"Písmo {obj.get('/BaseFont', '')} není vložené.")
        space = obj.get("/ColorSpace")
        if isinstance(space, p.Dictionary):
            for value in space.values():
                check_color(value)
        elif space is not None:
            check_color(space)
        if obj.get("/S") == p.Name.Transparency:
            if variant != 4:
                raise PDFXError("Starší předvolba PDF/X obsahuje živou průhlednost.")
            if "/CS" in obj:
                check_color(obj.CS)
        if variant != 4 and (obj.get("/CA", 1) != 1 or obj.get("/ca", 1) != 1
                             or str(obj.get("/BM", "/Normal")) not in ("/Normal", "/Compatible")
                             or str(obj.get("/SMask", "/None")) != "/None"):
            raise PDFXError("Starší předvolba PDF/X obsahuje živou průhlednost.")
        if any(key in obj for key in ("/TR", "/HTP")) or str(obj.get("/TR2", "/Default")) != "/Default":
            raise PDFXError("Výstup obsahuje nepodporovanou převodní funkci nebo nastavení rastru.")
        if isinstance(obj, p.Stream):
            filters = obj.get("/Filter")
            if variant != 4 and ("/JBIG2Decode" in str(filters) or "/JPXDecode" in str(filters)):
                raise PDFXError("Starší předvolba PDF/X obsahuje kompresi JBIG2 nebo JPEG 2000.")
            if obj.get("/Subtype") == p.Name.Form or obj.get("/PatternType") == 1:
                check_content(obj)


def finish_pdfx(source, converted, destination, variant, profile, title):
    if converted.pdf_version not in (("1.6",) if variant == 4 else ("1.3", "1.4")):
        raise PDFXError("Ghostscript nevytvořil očekávanou verzi PDF pro zvolenou předvolbu PDF/X.")
    restore_boxes(source, converted)
    attach_intent(converted, *profile)
    check_candidate(converted, variant, profile)

    label, version = STANDARDS[variant]
    info = converted.docinfo
    info.Title = str(source.docinfo.get("/Title", "")).strip() or title
    info.Trapped = p.Name("/True") if str(source.docinfo.get("/Trapped")) == "/True" else p.Name("/False")
    info.GTS_PDFXVersion = label
    if "/Metadata" in converted.Root:
        del converted.Root["/Metadata"]
    if variant == 4:
        now = datetime.now(timezone.utc).replace(microsecond=0)
        info.CreationDate = info.ModDate = now.strftime("D:%Y%m%d%H%M%SZ")
        with converted.open_metadata(set_pikepdf_as_editor=False, update_docinfo=False) as metadata:
            metadata.load_from_docinfo({key: info[key] for key in INFO_XMP_FIELDS if key in info})
            metadata["pdf:Trapped"] = str(info.Trapped)[1:]
            metadata["xmp:MetadataDate"] = now.isoformat()
            metadata[XMP_VERSION] = label
            for key, value in zip(XMP_IDENTIFIERS, (f"uuid:{uuid4()}", "1", "default")):
                metadata[key] = value
    converted.save(destination, force_version=version, fix_metadata_version=False,
                   object_stream_mode=p.ObjectStreamMode.disable)
    with p.Pdf.open(destination, attempt_recovery=False) as saved:
        if (saved.pdf_version != version or saved.check_pdf_syntax() or "/ID" not in saved.trailer
                or str(saved.docinfo.get("/GTS_PDFXVersion")) != label):
            raise PDFXError("Uložený soubor neprošel kontrolou struktury nebo označení PDF/X.")
        if variant == 4:
            with saved.open_metadata(strict=True, set_pikepdf_as_editor=False, update_docinfo=False) as metadata:
                if metadata.get(XMP_VERSION) != label or not all(metadata.get(key) for key in
                        (*XMP_DATES, *XMP_IDENTIFIERS, "dc:title")):
                    raise PDFXError("Uložený soubor nemá požadovaná metadata PDF/X-4.")
