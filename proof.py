"""Plate proofs from Ghostscript's actual rendered separations."""

from base64 import b64encode
from io import BytesIO
from pathlib import Path
import re
from tempfile import TemporaryDirectory

from PIL import Image, ImageCms

from print_profiles import get_profile, source_cmyk


PROCESS = ("Cyan", "Magenta", "Yellow", "Black")
PLATE_ORDER = {name: n for n, name in enumerate(PROCESS)}
BOXES = {"media": ("mediabox", None), "crop": ("cropbox", "-dUseCropBox"),
         "trim": ("trimbox", "-dUseTrimBox"), "bleed": ("bleedbox", "-dUseBleedBox")}


class ProofError(ValueError):
    pass


def plate_order(name):
    return PLATE_ORDER.get(name, 4), name.casefold(), name


def output_intent(pdf):
    for intent in pdf.Root.get("/OutputIntents", []):
        condition = str(intent.get("/OutputConditionIdentifier", "")).strip()
        description = str(intent.get("/Info", "")).strip()
        if condition and condition.casefold() not in description.casefold():
            return f"{condition} · {description}" if description else condition
        if description:
            return description
    return ""


def print_profile(exe, folder, pdf, key):
    source = source_cmyk(pdf)
    label = "Výchozí CMYK Ghostscriptu"
    profile = None
    if key:
        profile, label, _ = get_profile(key, "")
    elif source:
        profile = source
        label = "CMYK z výstupního záměru PDF"
    if not profile:
        bundled = Path(exe).parent.parent / "iccprofiles" / "default_cmyk.icc"
        system = Path("/usr/share/color/icc/ghostscript/default_cmyk.icc")
        for path in (bundled, system):
            if path.is_file():
                profile = path.read_bytes()
                break
    if not profile:
        return None, label, []

    (folder / "simulace.icc").write_bytes(profile)
    if source and source != profile:
        (folder / "zdroj.icc").write_bytes(source)
        source_file = "zdroj.icc"
    else:
        source_file = "simulace.icc"
    args = ["-dUsePDFX3Profile=false", "--permit-file-read=simulace.icc",
            "-sOutputICCProfile=simulace.icc", f"-sDefaultCMYKProfile={source_file}"]
    if source_file != "simulace.icc":
        args.append("--permit-file-read=zdroj.icc")
    return profile, label, args


def render_dpi(pdf, pages, requested, pixel_limit, box):
    max_inches = 0
    for n in pages:
        page = pdf.pages[n - 1]
        left, bottom, right, top = (float(v) for v in getattr(page, box))
        scale = float(page.obj.get("/UserUnit", 1))
        max_inches = max(max_inches, max(right - left, top - bottom) * scale / 72)
    if max_inches <= 0 or max_inches > pixel_limit:
        raise ProofError("Formát stránky je pro bezpečný náhled příliš velký nebo neplatný.")
    return min(requested, int(pixel_limit / max_inches))


def render(exe, folder, gs, source, pages, dpi, profile_args, prefix, selected=None):
    args = ["-q", "-sDEVICE=tiffsep", "-dPrinted=true",
            "-dOverprint=/enable", "-dDeviceGrayToK=true", "-sCompression=lzw",
            f"-r{dpi}", "-sPageList=" + ",".join(map(str, pages)),
            f"-sOutputFile={prefix}-%d.tif", *profile_args]
    if selected is not None:
        # Names come from the PDF. Only hexadecimal data is placed in PostScript.
        order = " ".join(f"<{name.encode('utf-8').hex()}> cvn" for name in selected)
        args += ["-dNoSeparationFiles", "-c", f"<< /SeparationOrder [{order}] >> setpagedevice"]
    log = gs(exe, folder, [*args, "-f", source])
    if re.search(r"too many (?:spot colors|colorants)|maximum number.*(?:spots|colorants)", log, re.I):
        raise ProofError("Ghostscript překročil počet podporovaných tiskových desek.")
    groups = {}
    pattern = re.compile(rf"^{re.escape(prefix)}-(\d+)(?:\((.*)\))?\.tif$", re.S)
    for path in folder.glob(f"{prefix}-*.tif"):
        match = pattern.fullmatch(path.name)
        if match:
            groups.setdefault(int(match[1]), {})[match[2]] = path
    rendered = [group for _, group in sorted(groups.items())]
    if len(rendered) != len(pages) or any(None not in group for group in rendered):
        raise ProofError("Ghostscript nevytvořil tiskový náhled všech vybraných stran.")
    if selected is None and any(not set(PROCESS).issubset(group) for group in rendered):
        raise ProofError("Ghostscript nevytvořil úplné procesní separace.")
    return rendered


def box_options(req):
    box = BOXES.get(req.get("box", "media"))
    if box is None:
        raise ProofError("Typ rámečku stránky je neplatný.")
    attribute, flag = box
    return attribute, ([flag] if flag else [])


def describe(pdf):
    sizes = []
    for page in pdf.pages:
        unit = float(page.obj.get("/UserUnit", 1))
        rotated = int(page.obj.get("/Rotate", 0)) % 180 == 90
        boxes = {}
        for name, (attribute, _) in BOXES.items():
            left, bottom, right, top = map(float, getattr(page, attribute))
            width, height = max(1, (right-left)*unit), max(1, (top-bottom)*unit)
            if rotated:
                width, height = height, width
            boxes[name] = [width, height]
        sizes.append(boxes)
    return dict(pages=len(sizes), sizes=sizes, output_intent=output_intent(pdf))


def page_options(pdf, req):
    page = req.get("page", 1)
    rotation = req.get("rotation", 0)
    if type(page) is not int or not 1 <= page <= len(pdf.pages):
        raise ProofError("Vyberte existující stránku PDF.")
    if type(rotation) is not int or rotation not in (0, 90, 180, 270):
        raise ProofError("Otočení stránky musí být násobkem 90 stupňů.")
    return page, rotation


def render_source(pdf, src, folder, stage):
    if pdf.is_encrypted:
        stage(pdf, src, folder / "nahled.pdf")
        return "nahled.pdf"
    return str(src.resolve())


def screen_page(pdf, src, req, exe, gs, stage, *, printed=False):
    page, rotation = page_options(pdf, req)
    box, args = box_options(req)
    target = req.get("pixels", 1850)
    if type(target) is not int or not 400 <= target <= (4096 if printed else 3200):
        raise ProofError("Rozlišení náhledu je neplatné.")
    dpi = render_dpi(pdf, [page], 300, target, box)
    with TemporaryDirectory(prefix="pdf-view-") as name:
        folder = Path(name)
        source = render_source(pdf, src, folder, stage)
        gs(exe, folder, ["-q", "-sDEVICE=png16m", f"-dPrinted={'true' if printed else 'false'}", "-dGraphicsAlphaBits=4",
                         "-dTextAlphaBits=4", f"-r{dpi}", f"-sPageList={page}",
                         "-sOutputFile=obraz.png", *args, "-f", source])
        png = folder / "obraz.png"
        if rotation:
            with Image.open(png) as image:
                data = BytesIO()
                image.rotate(-rotation, expand=True).save(data, "PNG", compress_level=3)
                return dict(image=b64encode(data.getvalue()).decode("ascii"))
        return dict(image=b64encode(png.read_bytes()).decode("ascii"))


def stat(path, name):
    with Image.open(path) as image:
        if image.mode != "L":
            raise ProofError(f"Deska {name} má neočekávaný formát.")
        histogram = image.histogram()
        size = image.width * image.height
    pixels = size - histogram[255]
    peak = 255 - next(index for index, count in enumerate(histogram) if count)
    return dict(name=name, active=pixels > 0, coverage=100 * pixels / size, maximum=100 * peak / 255)


def proof_image(composite, profile):
    with Image.open(composite) as raw:
        if profile:
            try:
                return ImageCms.profileToProfile(raw, BytesIO(profile), ImageCms.createProfile("sRGB"),
                                                renderingIntent=1, outputMode="RGB")
            except ImageCms.PyCMSError as exc:
                raise ProofError("Výstupní ICC nelze použít pro zobrazení na monitoru.") from exc
        return raw.convert("RGB")


def page_proof(pdf, src, req, exe, gs, stage):
    page, rotation = page_options(pdf, req)
    with TemporaryDirectory(prefix="pdf-proof-") as name:
        folder = Path(name)
        source = render_source(pdf, src, folder, stage)
        profile, label, args = print_profile(exe, folder, pdf, req.get("proof_profile", ""))
        box, box_args = box_options(req)
        args += box_args
        dpi = render_dpi(pdf, [page], 192, 1850, box)
        group = render(exe, folder, gs, source, [page], dpi, args, "plate")[0]
        names = sorted((name for name in group if name is not None), key=plate_order)
        stats = [stat(group[name], name) for name in names]
        selected = req.get("plates")
        hidden = req.get("hide_plates", [])
        if not isinstance(hidden, list) or any(not isinstance(name, str) for name in hidden):
            raise ProofError("Výběr tiskových desek je neplatný.")
        if selected is not None:
            if not isinstance(selected, list) or any(not isinstance(name, str) for name in selected):
                raise ProofError("Výběr tiskových desek je neplatný.")
            selected = [name for name in names if name in selected]
        elif hidden:
            selected = [name for name in names if name not in hidden]
        if selected == []:
            with Image.open(group[None]) as raw:
                image = Image.new("RGB", raw.size, "white")
        else:
            composite = group[None]
            if selected is not None and len(selected) < len(names):
                composite = render(exe, folder, gs, source, [page], dpi, args, "selected", selected)[0][None]
            image = proof_image(composite, profile)
        if rotation:
            image = image.rotate(-rotation, expand=True)
        encoded = BytesIO()
        image.save(encoded, "PNG", compress_level=3)
        return dict(pages=len(pdf.pages), dpi=dpi, plates=stats,
                    output_intent=output_intent(pdf), simulation=label,
                    image=b64encode(encoded.getvalue()).decode("ascii"))


def scan_plates(pdf, src, req, exe, gs, stage):
    pages = sorted(req.get("order") or range(1, len(pdf.pages) + 1))
    found = {}
    with TemporaryDirectory(prefix="pdf-plates-") as name:
        folder = Path(name)
        stage(pdf, src, folder / "nahled.pdf")
        _, label, args = print_profile(exe, folder, pdf, req.get("proof_profile", ""))
        box, box_args = box_options(req)
        args += box_args
        requested = req.get("proof_dpi", 150)
        lowest = requested
        for offset in range(0, len(pages), 4):
            batch = pages[offset:offset + 4]
            dpi = render_dpi(pdf, batch, requested, 2200, box)
            lowest = min(lowest, dpi)
            frames = render(exe, folder, gs, "nahled.pdf", batch, dpi, args, "scan")
            for number, group in zip(batch, frames):
                for ink, path in group.items():
                    if ink is not None:
                        with Image.open(path) as image:
                            if image.mode != "L":
                                raise ProofError(f"Deska {ink} má neočekávaný formát.")
                            if image.getextrema()[0] < 255:
                                found.setdefault(ink, []).append(number)
                    path.unlink()
        return dict(pages=len(pages), dpi=lowest, simulation=label,
                    output_intent=output_intent(pdf),
                    used_plates=[dict(name=name, pages=found[name]) for name in sorted(found, key=plate_order)])
