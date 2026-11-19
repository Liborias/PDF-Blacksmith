"""Inspect page colors and convert only the artwork being inserted into a PDF."""

from dataclasses import dataclass, field
import os
from pathlib import Path

import pikepdf as pdf

from color_conversion import ColorConversionError, convert_spot_vectors
from pdfx import page_geometry
from print_profiles import icc_matches


DEVICE = {"/DeviceGray": "Gray", "/DeviceRGB": "RGB", "/DeviceCMYK": "CMYK",
          "/G": "Gray", "/RGB": "RGB", "/CMYK": "CMYK"}
ICC = {"Gray": b"GRAY", "RGB": b"RGB ", "CMYK": b"CMYK"}
CHANNELS = {"Gray": 1, "RGB": 3, "CMYK": 4}
DEVICE_NAME = {space: pdf.Name("/Device" + space) for space in ICC}
LABELS = {"Gray": "Stupně šedi", "RGB": "RGB", "CMYK": "CMYK"}


def icc_stream(stream):
    if isinstance(stream, pdf.Stream):
        data = stream.read_bytes()
        for space, signature in ICC.items():
            if icc_matches(data, signature) and stream.get("/N", CHANNELS[space]) == CHANNELS[space]:
                return space, data
    return "unknown", None


def color_space(value, resources, names=()):
    if isinstance(value, pdf.Array) and value:
        kind = str(value[0])
        if kind == "/ICCBased" and len(value) == 2:
            return icc_stream(value[1])
        if kind == "/Indexed" and len(value) == 4:
            return color_space(value[1], resources, names)
        if kind == "/CalGray":
            return "CalGray", None
        if kind in ("/Separation", "/DeviceN"):
            return "spot", None
        if kind in ("/CalRGB", "/Lab"):
            return "calibrated", None
        if kind == "/Pattern":
            return "pattern", None
    if isinstance(value, pdf.Name):
        name = str(value)
        device = DEVICE.get(name)
        default = "/Default" + device if device else name
        space = resources.get("/ColorSpace", {}).get(default)
        if space is not None:
            if default in names or len(names) >= 16:
                return "unknown", None
            return color_space(space, resources, (*names, default))
        if device:
            return device, None
        if name == "/Pattern":
            return "pattern", None
    return "unknown", None


@dataclass
class PageColors:
    profiles: dict = field(default_factory=dict)
    risks: set = field(default_factory=set)
    colored: bool = False

    def add(self, model, profile=None, neutral=False):
        if model == "pattern":
            return
        self.profiles.setdefault(model, set()).add(profile)
        self.colored |= not neutral and model not in ("Gray", "CalGray")
        if model == "spot":
            self.risks.add("spot")


def page_colors(page, *, check_black_images=True):
    result = PageColors()
    if any(str(item.get("/Subtype")) != "/Link" or "/AP" in item for item in page.obj.get("/Annots", [])):
        result.risks.add("annotations")

    def image_color(image, metadata, resources, fill):
        if metadata.get("/ImageMask", False):
            result.add(*fill)
            return
        model, profile = color_space(metadata.get("/ColorSpace"), resources)
        neutral = False
        if check_black_images and model == "CMYK" and "/Decode" not in metadata:
            try:
                with image.as_pil_image() as pixels:
                    neutral = pixels.mode == "CMYK" and pixels.getextrema()[:3] == ((0, 0),) * 3
            except (OSError, ValueError, NotImplementedError, pdf.PdfError):
                pass
        result.add(model, profile, neutral)

    def walk(obj, resources, fill, stroke, drawing=()):
        ident = obj.objgen
        if ident in drawing or len(drawing) >= 32:
            result.add("unknown")
            return
        drawing = (*drawing, ident)
        resources = obj.get("/Resources", resources)
        group = obj.get("/Group")
        if group is not None and str(group.get("/S")) == "/Transparency":
            result.risks.add("transparency")
        if "/OC" in obj:
            result.risks.add("layers")
        if str(obj.get("/Type")) == "/Page" and "/Contents" not in obj:
            return
        state, mode = [], 0
        for op in pdf.parse_content_stream(obj):
            operator, args = str(op.operator), op.operands
            if operator == "q":
                state.append((fill, stroke, mode))
            elif operator == "Q" and state:
                fill, stroke, mode = state.pop()
            elif operator in ("g", "G", "rg", "RG", "k", "K", "cs", "CS") and args:
                name = {"g": "/DeviceGray", "rg": "/DeviceRGB", "k": "/DeviceCMYK"}.get(operator.lower())
                model, profile = color_space(pdf.Name(name) if name else args[0], resources)
                paint = (model, profile, operator.lower() == "k" and len(args) == 4 and not any(args[:3]))
                if operator.islower():
                    fill = paint
                else:
                    stroke = paint
            elif operator in ("sc", "SC", "scn", "SCN"):
                paint = fill if operator.islower() else stroke
                if paint[0] == "pattern":
                    pattern = resources.get("/Pattern", {}).get(str(args[-1])) if args else None
                    if isinstance(pattern, pdf.Stream):
                        # Uncoloured tiling patterns borrow a base colour from the caller.
                        if pattern.get("/PaintType") != 1:
                            result.add("unknown")
                        else:
                            walk(pattern, resources, fill, stroke, drawing)
                    elif isinstance(pattern, pdf.Dictionary) and "/Shading" in pattern:
                        result.add(*color_space(pattern.Shading.get("/ColorSpace"), resources))
                    else:
                        result.add("unknown")
                elif paint[0] == "CMYK" and not paint[2] and len(args) == 4 and not any(args[:3]):
                    paint = paint[0], paint[1], True
                    if operator.islower():
                        fill = paint
                    else:
                        stroke = paint
                elif paint[2]:
                    paint = paint[0], paint[1], False
                    if operator.islower():
                        fill = paint
                    else:
                        stroke = paint
            elif operator == "Tr":
                mode = int(args[0])
            elif operator in ("f", "F", "f*", "B", "B*", "b", "b*", "S", "s"):
                if operator not in ("S", "s"):
                    result.add(*fill)
                if operator in ("B", "B*", "b", "b*", "S", "s"):
                    result.add(*stroke)
            elif operator in ("Tj", "TJ", "'", '"'):
                if mode in (0, 2, 4, 6):
                    result.add(*fill)
                if mode in (1, 2, 5, 6):
                    result.add(*stroke)
            elif operator == "Do":
                xobject = resources.get("/XObject", {}).get(str(args[0]))
                if not isinstance(xobject, pdf.Stream):
                    result.add("unknown")
                elif str(xobject.get("/Subtype")) == "/Form":
                    walk(xobject, resources, fill, stroke, drawing)
                elif str(xobject.get("/Subtype")) == "/Image":
                    if "/OC" in xobject:
                        result.risks.add("layers")
                    if any(key in xobject for key in ("/SMask", "/Mask")):
                        result.risks.add("transparency")
                    image_color(pdf.PdfImage(xobject), xobject, resources, fill)
                else:
                    result.add("unknown")
            elif operator == "sh":
                shading = resources.get("/Shading", {}).get(str(args[0]))
                result.add(*color_space(shading.get("/ColorSpace") if shading is not None else None, resources))
            elif operator == "INLINE IMAGE":
                image_color(args[0], args[0].obj, resources, fill)
            elif operator == "gs":
                graphics = resources.get("/ExtGState", {}).get(str(args[0]), {})
                if graphics.get("/CA", 1) != 1 or graphics.get("/ca", 1) != 1 or str(graphics.get("/BM", "/Normal")) != "/Normal" or str(graphics.get("/SMask", "/None")) != "/None":
                    result.risks.add("transparency")
                if graphics.get("/OP") or graphics.get("/op"):
                    result.risks.add("overprint")
            elif operator == "BDC" and args[0] == pdf.Name.OC:
                result.risks.add("layers")
            elif operator == "Tf":
                font = resources.get("/Font", {}).get(str(args[0]), {})
                if str(font.get("/Subtype")) == "/Type3":
                    result.add("unknown")

    resources = page.resources
    initial = (*color_space(pdf.Name.DeviceGray, resources), False)
    try:
        walk(page.obj, resources, initial, initial)
    except (pdf.PdfError, AttributeError, IndexError, ValueError, TypeError):
        result.add("unknown")
    return result


@dataclass
class DestinationColor:
    space: str
    profile: bytes | None
    label: str
    page_profile: bool = False


def destination_color(target, order):
    intents = target.Root.get("/OutputIntents")
    if intents:
        if not isinstance(intents, pdf.Array):
            return None
        resolved = {}
        for intent in intents:
            if not isinstance(intent, pdf.Dictionary):
                return None
            space, profile = icc_stream(intent.get("/DestOutputProfile"))
            if profile is None:
                return None
            label = str(intent.get("/OutputConditionIdentifier") or intent.get("/Info") or LABELS[space])[:160]
            resolved[space, profile] = DestinationColor(space, profile, f"{label} · {space}")
        return next(iter(resolved.values())) if len(resolved) == 1 else None

    profiles = {}
    for number in order:
        for space, values in page_colors(target.pages[number-1], check_black_images=False).profiles.items():
            profiles.setdefault(space, set()).update(values)
        if set(profiles) - {"Gray", "CalGray", "RGB", "CMYK"} or {"RGB", "CMYK"} <= profiles.keys():
            return None
    space = next((model for model in ("RGB", "CMYK", "Gray") if model in profiles), None)
    if space is None or len(profiles[space]) != 1 or space == "Gray" and "CalGray" in profiles:
        return None
    profile, = profiles[space]
    return DestinationColor(space, profile, f"{LABELS[space]} · ICC na stránkách" if profile
                            else f"{LABELS[space]} · výchozí profil Ghostscriptu", bool(profile))


def source_profiles(donor):
    profiles = {}
    intents = donor.Root.get("/OutputIntents")
    if isinstance(intents, pdf.Array):
        for intent in intents:
            if isinstance(intent, pdf.Dictionary):
                model, data = icc_stream(intent.get("/DestOutputProfile"))
                if data is not None:
                    profiles.setdefault(model, set()).add(data)
    return {model: next(iter(values)) for model, values in profiles.items() if len(values) == 1}


RISKS = {
    "spot": "Přímé barvy a jejich samostatné tiskové desky se nahradí procesními barvami.",
    "transparency": "Vzhled průhlednosti a režimů prolnutí se může změnit.",
    "overprint": "Tiskové přetisky se mohou po převodu chovat jinak.",
    "layers": "Převod použije výchozí viditelnost vrstev a může je sloučit.",
    "annotations": "Poznámky, formulářová pole a jejich barvy se ponechají v původní podobě.",
}


def transfer_color_plan(target, target_order, donor, indices):
    destination = destination_color(target, target_order)
    summaries = [page_colors(donor.pages[index]) for index in indices]
    conversion = [i for i, info in enumerate(summaries) if info.colored]
    reason = ""
    if destination is None:
        reason = "Cílové PDF nemá jednoznačný barevný prostor nebo vložený výstupní profil."
    elif any("unknown" in info.profiles for info in summaries):
        reason = "U některých vybraných prvků nelze spolehlivě určit barevný prostor."
    elif not conversion:
        reason = "Vybrané stránky už používají stupně šedi nebo pouze černou desku K."
    warnings = []
    if not reason:
        risks = set.union(set(), *(summaries[i].risks for i in conversion))
        warnings = ["Odstíny mimo rozsah cílového prostoru se mohou nevratně změnit."]
        warnings.extend(message for key, message in RISKS.items() if key in risks)
        assigned = source_profiles(donor)
        if any(space not in assigned and any(None in summaries[i].profiles.get(space, ()) for i in conversion)
               for space in ICC):
            warnings.append("Pro neoznačené barvy se použijí výchozí zdrojové profily Ghostscriptu.")
    details = {"target": destination.label if destination else "", "reason": reason, "warnings": warnings,
               "convert": len(conversion) if not reason else 0, "gray": len(indices)-len(conversion)}
    return destination, conversion, details


def profile_space(document, space, data):
    return pdf.Array([pdf.Name.ICCBased,
                      document.make_stream(data, N=CHANNELS[space], Alternate=DEVICE_NAME[space])])


def assign_page_defaults(page, defaults, *, override=False):
    resources = pdf.Dictionary(page.resources)
    spaces = pdf.Dictionary(resources.get("/ColorSpace", {}))
    for name, value in defaults.items():
        if override or name not in spaces:
            spaces[name] = value
    resources.ColorSpace = spaces
    page.obj.Resources = resources


def convert_inserted_artwork(target, pages, donor, destination, folder, exe, run):
    """Render new page content in the destination device space and keep page identities."""
    staged, output = folder / "barvy-vstup.pdf", folder / "barvy-vystup.pdf"
    inputs = source_profiles(donor)
    bundled = Path(exe).parent.parent / "iccprofiles"
    for model, signature in ICC.items():
        if model not in inputs:
            path = bundled / f"default_{model.lower()}.icc"
            if path.is_file() and icc_matches(data := path.read_bytes(), signature):
                inputs[model] = data
    with pdf.Pdf.new() as artwork:
        artwork.add_pages_from(target, pages, forms="strip", outlines="ignore")
        # A Ghostscript default alone can leave DeviceCMYK values untouched
        # during CMYK-to-CMYK pdfwrite. Explicit PDF defaults activate ICC mapping.
        defaults = {"/Default" + model: profile_space(artwork, model, data) for model, data in inputs.items()}
        for page in artwork.pages:
            if defaults:
                assign_page_defaults(page, defaults)
            page.obj.Rotate = 0
            for key in ("/Annots", "/CropBox", "/TrimBox", "/BleedBox", "/ArtBox"):
                if key in page.obj:
                    del page.obj[key]
        artwork.save(staged, fix_metadata_version=False)

    space = destination.space
    arguments = ["-sDEVICE=pdfwrite", "-dCompatibilityLevel=1.7", "-dAutoRotatePages=/None",
                 "-dPreserveAnnots=false", "-dUsePDFX3Profile=false", "-dPrinted=false",
                 "-dPreserveDeviceN=false", "-dPreserveSeparation=false", "-sBlendConversionStrategy=Simple",
                 "-dEmbedAllFonts=true", "-dSubsetFonts=true", "-dDownsampleColorImages=false",
                 "-dDownsampleGrayImages=false", "-dDownsampleMonoImages=false",
                 "-dAutoFilterColorImages=false", "-dAutoFilterGrayImages=false",
                 "-sColorImageFilter=FlateEncode", "-sGrayImageFilter=FlateEncode",
                 f"-sColorConversionStrategy={space}", f"-dProcessColorModel=/Device{space}",
                 "-sOutputFile=" + output.name]
    permitted = []
    for model, profile in inputs.items():
        name = "zdroj-" + model + ".icc"
        (folder / name).write_bytes(profile)
        permitted.append(name)
        arguments.append(f"-sDefault{model}Profile={name}")
    if destination.profile:
        name = "cil.icc"
        (folder / name).write_bytes(destination.profile)
        permitted.append(name)
        arguments.append("-sOutputICCProfile=" + name)
    if permitted:
        arguments.append("--permit-file-read=" + os.pathsep.join(permitted))
    run(exe, folder, [*arguments, "-c", "<< /NeverEmbed [] >> setdistillerparams", "-f", staged.name])

    with pdf.Pdf.open(output) as converted:
        if len(converted.pages) != len(pages):
            raise ColorConversionError("Převod změnil počet vkládaných stran.")
        destination_defaults = ({"/Default" + space: profile_space(target, space, destination.profile)}
                                if destination.page_profile else {})
        convert_spot_vectors(converted, space)
        for index, new in zip(pages, converted.pages):
            summary = page_colors(new, check_black_images=False)
            if set(summary.profiles) - {space, "Gray", "CalGray"}:
                raise ColorConversionError("Výstup obsahuje barvy mimo zvolený prostor. Strany nebyly vloženy.")
            old = target.pages[index]
            media, unit, _, width, height = page_geometry(old)
            dst, dst_unit, angle, dst_width, dst_height = page_geometry(new)
            if angle or abs(width-dst_width) > .05 or abs(height-dst_height) > .05:
                raise ColorConversionError("Převod změnil rozměry strany. Strany nebyly vloženy.")
            factor = dst_unit / unit
            x, y = (media[i] - dst[i]*factor for i in (0, 1))
            prefix = target.make_stream(f"q {factor:.12f} 0 0 {factor:.12f} {x:.12f} {y:.12f} cm\n".encode("ascii"))
            streams = new.obj.get("/Contents", [])
            if isinstance(streams, pdf.Stream):
                streams = [streams]
            old.obj.Contents = pdf.Array([prefix, *(target.copy_foreign(stream) for stream in streams),
                                          target.make_stream(b"\nQ\n")])
            old.obj.Resources = target.copy_foreign(converted.make_indirect(new.resources))
            if destination_defaults:
                assign_page_defaults(old, destination_defaults, override=True)
            if "/Group" in new.obj:
                old.obj.Group = target.copy_foreign(converted.make_indirect(new.obj.Group))
            elif "/Group" in old.obj:
                del old.obj["/Group"]
            if "/StructParents" in old.obj:
                del old.obj["/StructParents"]
