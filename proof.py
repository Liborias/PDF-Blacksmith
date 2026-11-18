"""Plate proofs from Ghostscript's actual rendered separations."""

from base64 import b64encode
from collections import OrderedDict
from dataclasses import dataclass
from io import BytesIO
import math
from pathlib import Path
import re
from tempfile import TemporaryDirectory

from PIL import Image, ImageChops, ImageCms, ImageMath

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
    max_inches = max(max(page_size(pdf.pages[n-1], box)) for n in pages) / 72
    if max_inches <= 0 or max_inches > pixel_limit:
        raise ProofError("Formát stránky je pro bezpečný náhled příliš velký nebo neplatný.")
    return min(requested, int(pixel_limit / max_inches))


def render(exe, folder, gs, source, pages, dpi, profile_args, prefix, spot_colors=None):
    args = ["-q", "-sDEVICE=tiffsep", "-dPrinted=true",
            "-dOverprint=/enable", "-dDeviceGrayToK=true", "-sCompression=lzw",
            f"-r{dpi}", "-sPageList=" + ",".join(map(str, pages)),
            f"-sOutputFile={prefix}-%d.tif", *profile_args]
    if spot_colors is not None:
        args.append("-dPrintSpotCMYK=true")
    log = gs(exe, folder, [*args, "-f", source])
    if re.search(r"too many (?:spot colors|colorants)|maximum number.*(?:spots|colorants)", log, re.I):
        raise ProofError("Ghostscript překročil počet podporovaných tiskových desek.")
    if spot_colors is not None:
        for name, c, m, y, k in re.findall(
                r'^%%SeparationColor: "(.*)" 100% ink = (\d+) (\d+) (\d+) (\d+) CMYK\r?$', log, re.M):
            weights = tuple(map(int, (c, m, y, k)))
            if max(weights) <= FULL_INK:
                spot_colors[name] = weights
    groups = {}
    pattern = re.compile(rf"^{re.escape(prefix)}-(\d+)(?:\((.*)\))?\.tif$", re.S)
    for path in folder.glob(f"{prefix}-*.tif"):
        match = pattern.fullmatch(path.name)
        if match:
            groups.setdefault(int(match[1]), {})[match[2]] = path
    rendered = [group for _, group in sorted(groups.items())]
    if len(rendered) != len(pages) or any(None not in group for group in rendered):
        raise ProofError("Ghostscript nevytvořil tiskový náhled všech vybraných stran.")
    if any(not set(PROCESS).issubset(group) for group in rendered):
        raise ProofError("Ghostscript nevytvořil úplné procesní separace.")
    return rendered


def box_options(req):
    box = BOXES.get(req.get("box", "media"))
    if box is None:
        raise ProofError("Typ rámečku stránky je neplatný.")
    attribute, flag = box
    if attribute in ("trimbox", "bleedbox"):
        # A missing TrimBox or BleedBox takes the CropBox as its default.
        return attribute, [flag, "-dUseCropBox"]
    return attribute, [flag] if flag else []


def page_box(page, box):
    left, bottom, right, top = map(float, getattr(page, box))
    left, right = sorted((left, right))
    bottom, top = sorted((bottom, top))
    if box != "mediabox":
        ml, mb, mr, mt = map(float, page.mediabox)
        left, right = max(left, min(ml, mr)), min(right, max(ml, mr))
        bottom, top = max(bottom, min(mb, mt)), min(top, max(mb, mt))
    return left, bottom, right, top


def page_size(page, box):
    left, bottom, right, top = page_box(page, box)
    unit = float(page.obj.get("/UserUnit", 1))
    width, height = max(1, (right-left)*unit), max(1, (top-bottom)*unit)
    if int(page.obj.get("/Rotate", 0)) % 180 == 90:
        return height, width
    return width, height


def describe(pdf):
    sizes = [{name: page_size(page, attr) for name, (attr, _) in BOXES.items()} for page in pdf.pages]
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


def screen_target(req, printed):
    target = req.get("pixels", 1850)
    if type(target) is not int or not 400 <= target <= (4096 if printed else 3200):
        raise ProofError("Rozlišení náhledu je neplatné.")
    return target


def screen_region(page, box, rotation, region):
    if not isinstance(region, dict):
        raise ProofError("Výřez náhledu je neplatný.")
    zoom = region.get("zoom")
    coordinates = [region.get(key) for key in ("x", "y", "width", "height")]
    if (type(zoom) not in (int, float) or not 0 < zoom <= 64
            or any(type(value) is not int for value in coordinates)):
        raise ProofError("Výřez náhledu je neplatný.")
    x, y, w, h = coordinates
    if min(x, y) < 0 or not (1 <= w <= 4096 and 1 <= h <= 4096) or w*h > (4 << 20):
        raise ProofError("Rozměry výřezu náhledu jsou neplatné.")
    width, height = (max(1, math.floor(points*zoom + 0.5)) for points in page_size(page, box))
    rotated_width, rotated_height = (height, width) if rotation % 180 else (width, height)
    if x + w > rotated_width or y + h > rotated_height:
        raise ProofError("Výřez přesahuje stránku.")
    return zoom, (x, y, w, h), (width, height)


def screen_region_args(page, box, rotation, region):
    zoom, (x, y, w, h), (width, height) = screen_region(page, box, rotation, region)
    # Undo the reader's draft rotation before placing the region on the device.
    if rotation == 90:
        x, y, w, h = y, height-x-w, h, w
    elif rotation == 180:
        x, y = width-x-w, height-y-h
    elif rotation == 270:
        x, y, w, h = width-y-h, x, h, w
    # PageOffset is in device directions. Its Y origin shifts when the device
    # height is reduced from the full page to the region height.
    dx, dy = -x/zoom, (height-y-h)/zoom
    return [f"-r{zoom*72:.12f}", f"-g{w}x{h}", "-dFIXEDMEDIA",
            "-c", f"<< /PageOffset [{dx:.12f} {dy:.12f}] >> setpagedevice"]


def screen_page(pdf, src, req, exe, gs, stage, *, printed=False):
    page, rotation = page_options(pdf, req)
    box, args = box_options(req)
    target = screen_target(req, printed)
    if not printed and "region" in req:
        resolution = screen_region_args(pdf.pages[page-1], box, rotation, req["region"])
    else:
        dpi = render_dpi(pdf, [page], 300 if printed else 4608, target, box)
        resolution = [f"-r{dpi}"]
    with TemporaryDirectory(prefix="pdf-view-") as name:
        folder = Path(name)
        source = render_source(pdf, src, folder, stage)
        gs(exe, folder, ["-q", "-sDEVICE=png16m", f"-dPrinted={'true' if printed else 'false'}", "-dGraphicsAlphaBits=4",
                         "-dTextAlphaBits=4", f"-sPageList={page}",
                         "-sOutputFile=obraz.png", *args, *resolution, "-f", source])
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


def proof_image(raw, profile):
    if profile:
        try:
            return ImageCms.profileToProfile(raw, BytesIO(profile), ImageCms.createProfile("sRGB"),
                                            renderingIntent=1, outputMode="RGB")
        except ImageCms.PyCMSError as exc:
            raise ProofError("Výstupní ICC nelze použít pro zobrazení na monitoru.") from exc
    return raw.convert("RGB")


@dataclass
class ProofSnapshot:
    profile: bytes | None
    spot_colors: dict
    rasters: dict
    overview: dict

    def open_raster(self, name):
        source = self.rasters[name]
        return Image.open(BytesIO(source) if isinstance(source, bytes) else source)


# Keep the current page and a visible neighboring page within a single budget.
# The reader process closes with the document, and nothing is retained on disk.
_proof_cache = OrderedDict()
PROOF_CACHE_LIMIT = 24 << 20
FULL_INK = 32760  # Ghostscript writes spot equivalents in its fixed-point frac units.


def composite_plates(snapshot, selected, size):
    empty = Image.new("L", size, 0)
    channels = [empty] * 4
    spot_ink = [None] * 4
    for name in selected:
        with snapshot.open_raster(name) as gray:
            if gray.size != size or gray.mode != "L":
                raise ProofError("Tiskové desky nemají shodný rozměr nebo formát.")
            coverage = ImageChops.invert(gray)
        if name in PROCESS:
            channels[PROCESS.index(name)] = coverage
            continue
        weights = snapshot.spot_colors.get(name)
        if weights is None:
            raise ProofError(f"Chybí náhledová barva přímé desky {name}.")
        intensity = coverage.convert("I")
        for index, weight in enumerate(weights):
            if not weight:
                continue
            total = spot_ink[index]
            spot_ink[index] = ImageMath.lambda_eval(
                lambda operands: operands["previous"] + operands["coverage"] * weight,
                previous=0 if total is None else total, coverage=intensity)
    for index, total in enumerate(spot_ink):
        if total is not None:
            extra = ImageMath.lambda_eval(lambda operands: operands["ink"] / FULL_INK, ink=total)
            channels[index] = ImageChops.add(channels[index], extra.convert("L"))
    return Image.merge("CMYK", channels)


def render_proof_selection(snapshot, selection, hidden, rotation):
    names = [plate["name"] for plate in snapshot.overview["plates"]]
    enabled = ([name for name in names if name not in hidden] if selection is None
               else [name for name in names if name in selection])
    with snapshot.open_raster(None) as full:
        if not enabled:
            image = Image.new("RGB", full.size, "white")
        elif len(enabled) == len(names):
            image = proof_image(full, snapshot.profile)
        else:
            image = proof_image(composite_plates(snapshot, enabled, full.size), snapshot.profile)
    if rotation:
        image = image.rotate(-rotation, expand=True)
    encoded = BytesIO()
    image.save(encoded, "PNG", compress_level=1)
    return {**snapshot.overview, "image": b64encode(encoded.getvalue()).decode("ascii")}


def page_proof(pdf, src, req, exe, gs, stage):
    page, rotation = page_options(pdf, req)
    selected = req.get("plates")
    hidden = req.get("hide_plates", [])
    if not isinstance(hidden, list) or any(not isinstance(name, str) for name in hidden):
        raise ProofError("Výběr tiskových desek je neplatný.")
    if selected is not None and (not isinstance(selected, list) or any(not isinstance(name, str) for name in selected)):
        raise ProofError("Výběr tiskových desek je neplatný.")
    box, box_args = box_options(req)
    profile_key = req.get("proof_profile", "")
    file_status = src.stat()
    key = (src.resolve(), file_status.st_dev, file_status.st_ino, file_status.st_size,
           file_status.st_mtime_ns, exe, page, box, profile_key)
    if key in _proof_cache:
        _proof_cache.move_to_end(key)
        return render_proof_selection(_proof_cache[key][0], selected, hidden, rotation)
    with TemporaryDirectory(prefix="pdf-proof-") as name:
        folder = Path(name)
        source = render_source(pdf, src, folder, stage)
        profile, label, args = print_profile(exe, folder, pdf, profile_key)
        dpi = render_dpi(pdf, [page], 192, 1850, box)
        spot_colors = {}
        group = render(exe, folder, gs, source, [page], dpi, args+box_args, "plate", spot_colors)[0]
        names = sorted((name for name in group if name is not None), key=plate_order)
        overview = dict(pages=len(pdf.pages), dpi=dpi, plates=[stat(group[name], name) for name in names],
                        output_intent=output_intent(pdf), simulation=label)
        cost = sum(path.stat().st_size for path in group.values()) + len(profile or b"")
        if cost <= PROOF_CACHE_LIMIT:
            while len(_proof_cache) >= 2 or sum(entry[1] for entry in _proof_cache.values()) + cost > PROOF_CACHE_LIMIT:
                _proof_cache.popitem(last=False)
            group = {name: path.read_bytes() for name, path in group.items()}
        snapshot = ProofSnapshot(profile, spot_colors, group, overview)
        if cost <= PROOF_CACHE_LIMIT:
            _proof_cache[key] = snapshot, cost
        return render_proof_selection(snapshot, selected, hidden, rotation)


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
