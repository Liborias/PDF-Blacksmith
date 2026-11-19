"""Local PDF operations for PDF Blacksmith. The Windows front end uses JSON on stdin/stdout."""

from contextlib import contextmanager
from hashlib import sha256
import json
import logging
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

import pikepdf
from print_profiles import ProfileError, attach_intent, attach_srgb, clear_pdfx_claim, get_profile, source_cmyk, source_icc
from export_settings import export_preset, export_settings, pdfwrite_settings
from color_conversion import (ColorConversionError, resolve_gray_icc, convert_rich_black,
                              gray_to_explicit_k, attach_gray_output_intent, preserve_solid_black_for_gray)
from pdfx import PDFXError, check_profile, finish_pdfx, pdfx_prefix
from proof import ProofError, describe, page_box, page_proof, scan_plates, screen_page
from reader_render import render_screen


class UserError(Exception):
    pass


class SignedPDF(UserError):
    pass


class MergeEncryption(UserError):
    pass


class PlateError(UserError):
    def __init__(self, hits):
        self.pages = sorted(hits)
        details = "; ".join(f"{page}: {', '.join(sorted(hits[page]))}" for page in self.pages[:12])
        more = f"; dalších stran {len(self.pages) - 12}" if len(self.pages) > 12 else ""
        super().__init__("V kontrolních separacích zůstala barva mimo desku K. "
                         f"Strany a nalezené desky: {details}{more}.")


def open_pdf(path, password):
    try:
        return pikepdf.Pdf.open(path, password=password)
    except pikepdf.PasswordError as exc:
        if not password:
            raise UserError("Toto PDF vyžaduje heslo k otevření. Bez něj nelze šifrovaný obsah přečíst. "
                            "Omezení tisku a úprav lze odstranit bez hesla vlastníka, když se PDF otevře bez hesla.") from exc
        raise UserError("Zadejte platné stávající heslo k PDF.") from exc


def same_destination(dst, src):
    return dst.resolve() == src.resolve() or (dst.exists() and dst.samefile(src))


@contextmanager
def output_file(src, request, *, protected=()):
    dst = Path(request["output"])
    if dst.suffix.lower() != ".pdf":
        raise UserError("Výstupní soubor musí mít příponu .pdf.")
    original = None
    same_path = dst.resolve() == src.resolve()
    if same_destination(dst, src):
        if not (same_path and request["action"] == "save_document" and request.get("replace_source")
                and request.get("overwrite")):
            raise UserError("Pro výstup vyberte jiný soubor než vstupní PDF.")
        dst = src.resolve(strict=True)
        original = dst.stat()
        if ((original.st_size, original.st_mtime_ns)
                != (request.get("source_size"), request.get("source_mtime_ns"))):
            raise UserError("Původní PDF se mezitím změnilo. Otevřete je znovu.")
    if any(same_destination(dst, path) for path in protected):
        raise UserError("Jako výstup vyberte jiný soubor než vstupní PDF.")
    with tempfile.TemporaryDirectory(prefix="pdftisk-", dir=dst.parent) as folder:
        pending = Path(folder) / "vystup.pdf"
        yield pending
        if any(same_destination(dst, path) for path in protected):
            raise UserError("Jako výstup vyberte jiný soubor než vstupní PDF.")
        if original is not None:
            if not same_file_state(original, dst.stat()):
                raise UserError("Původní PDF se mezitím změnilo. Otevřete je znovu.")
        if request.get("overwrite"):
            os.replace(pending, dst)
        else:
            publish_new_file(pending, dst)


def publish_new_file(pending, destination):
    if os.name == "nt":
        os.rename(pending, destination)  # Windows rename fails if the destination exists.
    else:
        os.link(pending, destination)


def signature_fields(pdf):
    acroform = pdf.Root.get("/AcroForm")
    visited = set()

    def walk(fields, parent_type=None):
        for field in fields:
            if not isinstance(field, pikepdf.Dictionary):
                continue
            if field.is_indirect:
                if field.objgen in visited:
                    continue
                visited.add(field.objgen)
            field_type = field.get("/FT", parent_type)
            if field_type == pikepdf.Name.Sig and isinstance(field.get("/V"), pikepdf.Dictionary):
                yield field
            yield from walk(field.get("/Kids", []), field_type)

    if isinstance(acroform, pikepdf.Dictionary):
        yield from walk(acroform.get("/Fields", []))


def clear_signatures(pdf, signed_fields):
    for field in signed_fields:
        if "/V" in field:
            del field["/V"]
        for widget in (field, *field.get("/Kids", [])):
            for key in ("/AP", "/AS"):
                if key in widget:
                    del widget[key]
    perms = pdf.Root.get("/Perms")
    if isinstance(perms, pikepdf.Dictionary):
        for key in ("/DocMDP", "/UR", "/UR3"):
            if key in perms:
                del perms[key]


def prepare_edit(pdf, request):
    signed = list(signature_fields(pdf))
    perms = pdf.Root.get("/Perms")
    has_permissions_signature = isinstance(perms, pikepdf.Dictionary) and any(
        key in perms for key in ("/DocMDP", "/UR", "/UR3")
    )
    if not (signed or has_permissions_signature):
        return False
    if not request.get("clear_signatures"):
        raise SignedPDF("PDF obsahuje digitální podpis. Úprava vytvoří kopii bez stávajících podpisů.")
    clear_signatures(pdf, signed)
    return True


def ghostscript_path(request):
    bundled = Path(__file__).resolve().parent / "ghostscript" / "bin" / "gswin64c.exe"
    exe = request.get("ghostscript") or str(bundled)
    if not (Path(exe).is_absolute() and Path(exe).is_file()):
        raise UserError("V aplikaci chybí součást Ghostscript. Použijte úplné sestavení PDF Blacksmith.")
    return exe


def stage_for_render(pdf, src, destination):
    if pdf.is_encrypted:
        pdf.save(destination, encryption=False, fix_metadata_version=False)
    else:
        shutil.copyfile(src, destination)


def stage_flatten_source(pdf, destination):
    # Render the entire sheet, including marks outside the display crop.
    crops = [(page, page.obj.get("/CropBox")) for page in pdf.pages]
    try:
        for page, _ in crops:
            page.obj.CropBox = page.mediabox
        pdf.save(destination, encryption=False, fix_metadata_version=False)
    finally:
        for page, crop in crops:
            if crop is None:
                if "/CropBox" in page.obj:
                    del page.obj["/CropBox"]
            else:
                page.obj.CropBox = crop


def run_ghostscript(exe, folder, args):
    env = os.environ.copy()
    # Ghostscript otherwise loads additional options and resources from these variables.
    for name in list(env):
        if name.upper() in {"GS_OPTIONS", "GS_LIB", "GS_FONTPATH", "GS_DEVICE"}:
            env.pop(name)
    try:
        result = subprocess.run(
            [exe, "-P-", "-dSAFER", "-dBATCH", "-dNOPAUSE", "-dPDFSTOPONERROR", *args],
            cwd=folder,
            env=env,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            timeout=900,
            creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0),
            check=False,
        )
    except subprocess.TimeoutExpired as exc:
        raise UserError("Zpracování v Ghostscriptu překročilo časový limit 15 minut.") from exc
    if result.returncode:
        detail = result.stdout.decode("utf-8", errors="replace").strip()[-2200:]
        raise UserError(f"Ghostscript skončil s chybou {result.returncode}.\n{detail}")
    return result.stdout.decode("utf-8", errors="replace")


def srgb_conversion(pdf, folder):
    from PIL import ImageCms

    srgb = ImageCms.ImageCmsProfile(ImageCms.createProfile("sRGB")).tobytes()
    (folder / "srgb.icc").write_bytes(srgb)
    permitted = ["srgb.icc"]
    rgb_input = "srgb.icc"
    args = ["-dUsePDFX3Profile=false", "-sOutputICCProfile=srgb.icc"]
    if data := source_icc(pdf, b"RGB "):
        rgb_input = "zdroj-rgb.icc"
        (folder / rgb_input).write_bytes(data)
        permitted.append(rgb_input)
    if data := source_cmyk(pdf):
        (folder / "zdroj-cmyk.icc").write_bytes(data)
        permitted.append("zdroj-cmyk.icc")
        args.append("-sDefaultCMYKProfile=zdroj-cmyk.icc")
    return srgb, [*args, f"-sDefaultRGBProfile={rgb_input}", "--permit-file-read=" + os.pathsep.join(permitted)]


def restore_raster_geometry(source, rendered, dpi):
    for number, (page, raster) in enumerate(zip(source.pages, rendered.pages), 1):
        left, bottom, right, top = page_box(page, "mediabox")
        width, height = right-left, top-bottom
        _, _, raster_width, raster_height = map(float, raster.mediabox)
        unit = float(page.obj.get("/UserUnit", 1))
        angle = int(page.obj.get("/Rotate", 0)) % 360
        wanted = (height*unit, width*unit) if angle % 180 else (width*unit, height*unit)
        if (angle % 90 or min(width, height, raster_width, raster_height, unit) <= 0
                or any(abs(a-b) > 72/dpi + 0.01 for a, b in zip(wanted, (raster_width, raster_height)))):
            raise UserError(f"Zploštění by změnilo velikost nebo orientaci strany {number}.")
        # Ghostscript bakes page rotation into the image. Map its coordinates
        # back to the original sheet before restoring the page dictionaries.
        a, b, c, d, x, y = (
            (width, 0, 0, height, left, bottom),
            (0, height, -width, 0, right, bottom),
            (-width, 0, 0, -height, right, top),
            (0, -height, width, 0, left, top),
        )[angle // 90]
        matrix = (a/raster_width, b/raster_width, c/raster_height, d/raster_height, x, y)
        transform = " ".join(f"{value:.10f}" for value in matrix)
        raster.contents_add(f"q {transform} cm\n".encode("ascii"), prepend=True)
        raster.contents_add(b"\nQ\n")
        for key in ("/MediaBox", "/CropBox", "/TrimBox", "/BleedBox", "/ArtBox"):
            if key in page.obj:
                raster.obj[key] = pikepdf.Array(page.obj[key])
        for key in ("/Rotate", "/UserUnit"):
            if key in page.obj:
                raster.obj[key] = page.obj[key]


def verify_plates(exe, folder, pages, dpi=150):
    from PIL import Image

    plate_names = {"Cyan", "Magenta", "Yellow", "Black"}
    hits = {}
    # Render in bounded groups so large jobs do not accumulate plate TIFFs.
    batch = 4 if dpi > 150 else 10
    for first in range(1, pages + 1, batch):
        last = min(first + batch - 1, pages)
        run_ghostscript(exe, folder, [
            "-q", "-sDEVICE=tiffsep", f"-r{dpi}", "-sCompression=lzw", "-dDeviceGrayToK=true",
            f"-dFirstPage={first}", f"-dLastPage={last}",
            "-sOutputFile=separace-%d.tif", "-f", "hotovo.pdf",
        ])
        plates = {}
        painted = {}
        for path in folder.glob("separace-*.tif"):
            number, sep, rest = path.stem.removeprefix("separace-").partition("(")
            if sep:
                plate = rest.removesuffix(")")
                number = int(number)
                plates.setdefault(number, set()).add(plate)
                if plate != "Black":
                    with Image.open(path) as image:
                        if image.mode != "L":
                            raise UserError(f"Ghostscript vytvořil nečekaný formát desky {plate}.")
                        if image.getextrema()[0] < 255:
                            painted.setdefault(number, set()).add(plate)
            path.unlink()
        if len(plates) != last - first + 1 or any(
            not plate_names.issubset(names) for names in plates.values()
        ):
            raise UserError("Ghostscript nevytvořil všechny očekávané tiskové separace.")
        for page, number in enumerate(sorted(plates), first):
            if number in painted:
                hits[page] = painted[number]
    if hits:
        raise PlateError(hits)


def repair_k_pages(exe, folder, pages, dpi=600):
    page_list = ",".join(map(str, pages))
    run_ghostscript(exe, folder, [
        "-sDEVICE=pdfimage8", f"-r{dpi}", "-sCompression=Flate", f"-sPageList={page_list}",
        "-dOverprint=/simulate", "-sOutputFile=oprava.pdf", "-f", "hotovo.pdf",
    ])
    with pikepdf.Pdf.open(folder / "hotovo.pdf") as pdf, pikepdf.Pdf.open(folder / "oprava.pdf") as raster:
        if len(raster.pages) != len(pages):
            raise UserError("Rastrová oprava vrátila jiný počet stran.")
        for number, replacement in zip(pages, raster.pages):
            page = pdf.pages[number - 1]
            if (any(abs(float(a) - float(b)) > 72 / dpi + 0.01 for a, b in zip(page.mediabox, replacement.mediabox))
                    or page.obj.get("/Rotate", 0) != replacement.obj.get("/Rotate", 0)):
                raise UserError(f"Rastrová oprava by změnila velikost nebo orientaci strany {number}.")
            page.obj.Contents = pdf.copy_foreign(raster.make_indirect(replacement.obj.Contents))
            page.obj.Resources = pdf.copy_foreign(raster.make_indirect(replacement.obj.Resources))
            if "/Group" in page.obj:
                del page.obj["/Group"]
        pdf.save(folder / "opravene.pdf", fix_metadata_version=False)
    os.replace(folder / "opravene.pdf", folder / "hotovo.pdf")


def convert_spot_vectors(pdf):
    """Use the gray tint transform if pdfwrite retained a vector spot color."""
    visited = set()

    def gray_tint(spot, value):
        domain = spot.get("/Domain", [0, 1])
        tint = min(float(domain[1]), max(float(domain[0]), float(value)))
        c0, c1 = float(spot["/C0"][0]), float(spot["/C1"][0])
        gray = c0 + tint ** float(spot["/N"]) * (c1 - c0)
        if "/Range" in spot:
            low, high = map(float, spot["/Range"])
            gray = min(high, max(low, gray))
        return [round(min(1.0, max(0.0, gray)), 8)]

    def walk(obj, resources):
        if obj.objgen in visited:
            return False
        visited.add(obj.objgen)
        spots = {}
        for name, space in resources.get("/ColorSpace", {}).items():
            if (isinstance(space, pikepdf.Array) and len(space) == 4
                    and space[0] == pikepdf.Name.Separation
                    and space[1] != pikepdf.Name("/None")
                    and space[2] == pikepdf.Name.DeviceGray
                    and isinstance(space[3], pikepdf.Dictionary) and space[3].get("/FunctionType") == 2):
                spots[name] = space[3]

        changed = False
        if spots:
            stack = []
            stroke = fill = None
            rewritten = []
            local_change = False
            for instr in pikepdf.parse_content_stream(obj):
                args, op = instr
                command = str(op)
                if command == "q":
                    stack.append((stroke, fill))
                elif command == "Q" and stack:
                    stroke, fill = stack.pop()
                elif command == "CS":
                    stroke = spots.get(str(args[0]))
                    if stroke is not None:
                        instr = (gray_tint(stroke, 1), pikepdf.Operator("G"))
                        local_change = True
                elif command == "cs":
                    fill = spots.get(str(args[0]))
                    if fill is not None:
                        instr = (gray_tint(fill, 1), pikepdf.Operator("g"))
                        local_change = True
                elif command in ("SC", "SCN") and stroke is not None:
                    instr = (gray_tint(stroke, args[0]), pikepdf.Operator("G"))
                    local_change = True
                elif command in ("sc", "scn") and fill is not None:
                    instr = (gray_tint(fill, args[0]), pikepdf.Operator("g"))
                    local_change = True
                elif command in ("G", "RG", "K"):
                    stroke = None
                elif command in ("g", "rg", "k"):
                    fill = None
                rewritten.append(instr)
            if local_change:
                content = pikepdf.unparse_content_stream(rewritten)
                if obj.get("/Type") == pikepdf.Name.Page:
                    obj.Contents = pdf.make_stream(content)
                else:
                    obj.write(content)
                changed = True

        children = list(resources.get("/XObject", {}).values())
        children += list(resources.get("/Pattern", {}).values())
        for graphic in resources.get("/ExtGState", {}).values():
            mask = graphic.get("/SMask")
            if isinstance(mask, pikepdf.Dictionary) and "/G" in mask:
                children.append(mask["/G"])
        for child in children:
            if isinstance(child, pikepdf.Stream) and "/Resources" in child:
                changed = walk(child, child.Resources) or changed
        return changed

    changed = False
    for page in pdf.pages:
        changed = walk(page.obj, page.obj.get("/Resources", {})) or changed
    return changed


def transform(src, request):
    color_mode = request.get("color_mode") or ("gray" if request.get("gray") else "unchanged")
    if color_mode not in ("unchanged", "gray", "rich_black", "k_cmyk"):
        raise UserError("Neznámý režim barevného převodu.")
    gray = color_mode in ("gray", "k_cmyk")
    explicit_k = color_mode == "k_cmyk"
    rich_black = color_mode == "rich_black"
    outline = request.get("outline", False)
    profile_key = request.get("profile", "")
    if (gray or rich_black) and profile_key:
        raise UserError("Nelze současně zvolit barevný režim K a běžný převod CMYK.")
    if rich_black and outline:
        raise UserError("Selektivní opravu složené černé nelze spojit s převodem textu na křivky.")
    decrypt = request.get("decrypt", False)
    preset = export_preset(request.get("preset", ""))
    if preset is None:
        raise UserError("Vyberte platnou předvolbu PDF.")
    mode = preset.get("mode")
    variant = preset.get("pdfx")
    if mode and (gray or rich_black or outline or decrypt or (profile_key and mode != "cmyk")):
        raise UserError("Pro tyto úpravy vyberte vlastní nastavení exportu.")
    if mode == "cmyk":
        profile_key = profile_key or preset["profile"]
    flatten = mode == "flatten"
    convert = gray or rich_black or outline or bool(profile_key) or mode in ("rewrite", "rgb", "flatten")
    view = request.get("view", {})
    if not (mode or convert or decrypt or "pages" in view):
        raise UserError("Vyberte alespoň jednu úpravu PDF.")
    options = export_settings(preset.get("options") if mode else request.get("export"))
    proof_dpi, repair_dpi = options["proof_dpi"], options["repair_dpi"]
    exe = ghostscript_path(request) if convert and not rich_black else None

    with open_pdf(src, request.get("pdf_password", "")) as pdf:
        source_pages = len(pdf.pages)
        selection = selected_pages(view, source_pages)
        if mode == "copy" and (selection is None or len(selection) == source_pages):
            with output_file(src, request) as pending:
                shutil.copyfile(src, pending)
            return f"Uložena kopie PDF. Počet stran {source_pages}."
        modified = prepare_edit(pdf, request)
        if selection is not None:
            modified = apply_page_order(pdf, list(pdf.pages), selection) or modified
        if gray:
            modified = bool(preserve_solid_black_for_gray(pdf)) or modified
        pages = len(pdf.pages)
        if rich_black:
            changed = convert_rich_black(pdf)
            with output_file(src, request) as pending:
                pdf.save(pending, encryption=False, fix_metadata_version=False)
            return ("Uloženo PDF. Upravené operace složené černé: "
                    f"{changed}. Zachovány ostatní barvy a obrázky. Počet stran {pages}.")
        if not convert:
            if decrypt and not pdf.is_encrypted:
                raise UserError("Vstupní PDF už je bez šifrování a omezení heslem.")
            restricted_only = decrypt and pdf.user_password_matched and not request.get("pdf_password")
            with output_file(src, request) as pending:
                pdf.save(pending, encryption=pdf.is_encrypted and not decrypt, fix_metadata_version=False)
            if not decrypt:
                return f"Uloženo PDF. Počet stran {pages}."
            if restricted_only:
                return "Uložena kopie bez omezení tisku a úprav. Heslo vlastníka nebylo potřeba."
            return "Uložena kopie PDF bez hesla."

        if not pages:
            raise UserError("Vstupní PDF nemá žádné strany.")
        repaired = []
        if explicit_k:
            profile_key = request.get("k_profile") or "fogra51"
        profile = get_profile(profile_key, request.get("custom_icc", "")) if profile_key else None
        gray_icc, gray_label = resolve_gray_icc(
            request.get("gray_profile", "automatic"), request.get("gray_custom_icc", "")
        ) if gray else (None, "")

        if variant:
            check_profile(profile, variant)
        with tempfile.TemporaryDirectory(prefix="pdftisk-") as tmp:
            folder = Path(tmp)
            staged = folder / "vstup.pdf"
            if flatten:
                stage_flatten_source(pdf, staged)
            elif pdf.is_encrypted or modified:
                pdf.save(staged, encryption=False, fix_metadata_version=False)
            else:
                shutil.copyfile(src, staged)

            if flatten:
                args = ["-sDEVICE=pdfimage24", f"-r{preset['dpi']}", "-sCompression=Flate",
                        "-dPrinted=false", "-dTextAlphaBits=4", "-dGraphicsAlphaBits=4", "-dOverprint=/simulate"]
            else:
                printed = "false" if mode == "rgb" else "true"
                args = ["-sDEVICE=pdfwrite", "-dCompatibilityLevel=1.7", "-dAutoRotatePages=/None",
                        "-dPreserveAnnots=false", f"-dPrinted={printed}", *pdfwrite_settings(options)]
                if mode:
                    args.extend(["-dEmbedAllFonts=true", "-dSubsetFonts=true"])
            args.append("-sOutputFile=hotovo.pdf")
            srgb = None
            if gray:
                args.extend([
                    "-sColorConversionStrategy=Gray", "-dProcessColorModel=/DeviceGray",
                    "-sBlendConversionStrategy=Simple", "-dPreserveSeparation=false", "-dPreserveDeviceN=false",
                    "-dUsePDFX3Profile=false",
                ])
                permitted = []
                if gray_icc:
                    (folder / "gray-output.icc").write_bytes(gray_icc)
                    args.append("-sOutputICCProfile=gray-output.icc")
                    permitted.append("gray-output.icc")
                source = source_cmyk(pdf)
                if source:
                    (folder / "source-cmyk.icc").write_bytes(source)
                    args.append("-sDefaultCMYKProfile=source-cmyk.icc")
                    permitted.append("source-cmyk.icc")
                if permitted:
                    args.append("--permit-file-read=" + os.pathsep.join(permitted))
            elif profile:
                (folder / "cil.icc").write_bytes(profile[0])
                source = source_cmyk(pdf)
                source_path = "cil.icc"
                if source and source != profile[0]:
                    (folder / "zdroj.icc").write_bytes(source)
                    source_path = "zdroj.icc"
                args.extend([
                    "-sColorConversionStrategy=CMYK", "-dProcessColorModel=/DeviceCMYK",
                    "-sBlendConversionStrategy=Simple", "-dUsePDFX3Profile=false",
                    f"--permit-file-read=cil.icc{os.pathsep}{source_path}",
                    "-sOutputICCProfile=cil.icc", f"-sDefaultCMYKProfile={source_path}",
                ])
            elif mode in ("rgb", "flatten"):
                srgb, color_args = srgb_conversion(pdf, folder)
                args.extend(color_args)
                if not flatten:
                    args.extend(["-sColorConversionStrategy=RGB", "-dProcessColorModel=/DeviceRGB",
                                 "-dPreserveDeviceN=false", "-dPreserveSeparation=false"])
            elif mode == "rewrite":
                args.append("-sColorConversionStrategy=LeaveColorUnchanged")
            if outline:
                args.append("-dNoOutputFonts=true")
            inputs = ["vstup.pdf"]
            if variant:
                pdfx_prefix(folder, profile, variant)
                args.extend([f"-dPDFX={variant}", "-dPDFACompatibilityPolicy=2", "-r300",
                             "-dMaxInlineImageSize=0", "-dPreserveSeparation=false", "-dPreserveDeviceN=false"])
                inputs.insert(0, "pdfx.ps")
            if mode and not flatten:
                args.extend(["-c", "<< /NeverEmbed [] >> setdistillerparams"])
            log = run_ghostscript(exe, folder, [*args, "-f", *inputs])

            with pikepdf.Pdf.open(folder / "hotovo.pdf") as finished:
                if len(finished.pages) != pages:
                    raise UserError("Počet stran v převedeném PDF se liší od vstupu.")
                if gray:
                    revised = convert_spot_vectors(finished)
                    if "/OutputIntents" in finished.Root:
                        del finished.Root["/OutputIntents"]
                        revised = True
                    clear_pdfx_claim(finished)
                    if explicit_k:
                        converted_vectors, converted_images = gray_to_explicit_k(finished)
                        attach_intent(finished, *profile)
                        revised = True
                    elif gray_icc:
                        attach_gray_output_intent(finished, gray_icc, gray_label)
                        revised = True
                    if revised:
                        finished.save(folder / "vystup-upraveny.pdf", fix_metadata_version=False)
                elif variant:
                    finish_pdfx(pdf, finished, folder / "vystup-upraveny.pdf", variant, profile,
                                Path(request["input"]).stem)
                elif profile:
                    attach_intent(finished, *profile)
                    finished.save(folder / "vystup-upraveny.pdf", fix_metadata_version=False)
                elif srgb is not None:
                    if flatten:
                        restore_raster_geometry(pdf, finished, preset["dpi"])
                    attach_srgb(finished, srgb)
                    finished.save(folder / "vystup-upraveny.pdf", min_version="1.7", fix_metadata_version=False)
                elif mode == "rewrite":
                    if "/OutputIntents" in pdf.Root:
                        finished.Root.OutputIntents = finished.copy_foreign(pdf.make_indirect(pdf.Root.OutputIntents))
                    clear_pdfx_claim(finished)
                    finished.save(folder / "vystup-upraveny.pdf", fix_metadata_version=False)
            revised_pdf = folder / "vystup-upraveny.pdf"
            if revised_pdf.exists():
                os.replace(revised_pdf, folder / "hotovo.pdf")
            if gray:
                try:
                    verify_plates(exe, folder, pages, proof_dpi)
                except PlateError as exc:
                    if outline:
                        raise UserError(f"{exc}\nVýstup se neuložil. Rastrovou opravu K lze použít při převodu bez křivek.") from exc
                    if not request.get("allow_raster"):
                        raise UserError(f"{exc}\nVýstup se neuložil. Pro tyto strany lze povolit rastrovou opravu K ({repair_dpi} dpi).") from exc
                    repaired = exc.pages
                    repair_k_pages(exe, folder, repaired, repair_dpi)
                    verify_plates(exe, folder, pages, proof_dpi)
            with output_file(src, request) as pending:
                shutil.copyfile(folder / "hotovo.pdf", pending)
    lines = [f"Uloženo PDF. Počet stran {pages}."]
    if mode:
        lines.append(f"Předvolba {preset['name']}.")
    if gray:
        lines.append(f"Separace při {proof_dpi} dpi ověřeny na všech stranách. C, M, Y i případné přímé desky jsou prázdné.")
        if explicit_k:
            lines.append(f"Explicitní DeviceCMYK C0 M0 Y0 Kx. Operace: {converted_vectors}, obrázky: {converted_images}.")
        lines.append(f"Gray ICC pro přepočet: {gray_label}.")
    if repaired:
        lines.append(f"Rastrová oprava {repair_dpi} dpi na stranách " + ", ".join(map(str, repaired)) + ".")
    if profile:
        lines.append(f"Výstupní profil CMYK: {profile[1]}. Profil ICC je vložen v PDF.")
    if variant:
        lines.append("Ověřeny rámečky, vložená písma, ICC profil, barvy a pravidla průhlednosti.")
    if outline:
        lines.append("Text zpracován v režimu převodu na křivky.")
    if "substitut" in log.lower() or "couldn't find a font" in log.lower():
        lines.append("Ghostscript hlásí náhradu písma. Zkontrolujte vzhled výstupu.")
    return "\n".join(lines)


def unlock_reader(reader, password):
    from pyhanko.pdf_utils.crypt import AuthStatus

    if reader.encrypted and reader.decrypt(password).status == AuthStatus.FAILED:
        raise UserError("Zadejte platné stávající heslo k PDF.")


def signature_box(reader):
    from math import ceil, floor

    page, _ = reader.find_page_for_modification(0)
    node = page.get_object()
    crop = media = None
    for _ in range(100):
        if crop is None and "/CropBox" in node:
            crop = node["/CropBox"]
        if media is None and "/MediaBox" in node:
            media = node["/MediaBox"]
        if crop is not None and media is not None:
            break
        if "/Parent" not in node:
            break
        node = node["/Parent"]
    box = crop if crop is not None else media
    if box is None:
        raise UserError("Nelze určit rozměry první strany pro viditelný podpis.")
    left, bottom, right, top = (float(x) for x in box)
    x1, x2 = max(ceil(left) + 10, floor(right) - 185), floor(right) - 10
    y1, y2 = ceil(bottom) + 10, min(floor(top) - 10, ceil(bottom) + 58)
    if x2 - x1 < 35 or y2 - y1 < 18:
        raise UserError("První strana je pro viditelný podpis příliš malá.")
    return x1, y1, x2, y2


def sign(src, request):
    from uuid import uuid4

    from pyhanko.pdf_utils.incremental_writer import IncrementalPdfFileWriter
    from pyhanko.pdf_utils.reader import PdfFileReader
    from pyhanko.sign import fields, signers
    from pyhanko.stamp import TextStampStyle

    cert = Path(request["certificate"])
    if not cert.is_file():
        raise UserError("Vyberte soubor s osobním certifikátem .p12 nebo .pfx.")
    signer = signers.SimpleSigner.load_pkcs12(
        str(cert), passphrase=request.get("cert_password", "").encode("utf-8")
    )
    if signer is None:
        raise UserError("Certifikát se nepodařilo otevřít. Zkontrolujte soubor, soukromý klíč a heslo.")

    with src.open("rb") as stream:
        reader = PdfFileReader(stream)
        if reader.encrypted:
            raise UserError("Před podepsáním uložte kopii PDF bez hesla v režimu Upravit.")
        previous = reader.embedded_regular_signatures
        if any(sig.docmdp_level == fields.MDPPerm.NO_CHANGES for sig in previous):
            raise UserError("Platná pravidla certifikace v PDF zakazují další změny.")
        empty = list(fields.enumerate_sig_fields(reader, filled_status=False))
        existing = bool(empty)
        if len(empty) > 1:
            raise UserError("PDF má několik prázdných podpisových polí. V této verzi lze "
                            "automaticky vyplnit jedno připravené pole.")
        if existing:
            name, _, _ = empty[0]
            new_field = None
        else:
            if any(sig.docmdp_level is not None for sig in previous):
                raise UserError("Certifikované PDF musí mít pro další podpis připravené podpisové pole.")
            name = f"Podpis-{uuid4().hex[:12]}"
            box = signature_box(reader) if request.get("visible") else None
            new_field = fields.SigFieldSpec(name, box=box)

        writer = IncrementalPdfFileWriter(stream, prev=reader)
        pdf_signer = signers.PdfSigner(
            signers.PdfSignatureMetadata(
                field_name=name, md_algorithm="sha256", subfilter=fields.SigSeedSubFilter.PADES,
            ),
            signer,
            stamp_style=TextStampStyle(stamp_text="Digitální podpis\n%(ts)s"),
            new_field_spec=new_field,
        )
        with output_file(src, request) as pending:
            with pending.open("wb") as file:
                pdf_signer.sign_pdf(writer, output=file, existing_fields_only=existing)
    return f"Uloženo podepsané PDF.\nPodepisující {signer.signing_cert.subject.human_friendly}.\nPole {name}."


def yes_no(value):
    if value is None:
        return "nelze určit"
    return "ano" if value else "ne"


def verify(src, request):
    import ssl

    from asn1crypto import x509
    from pyhanko.keys import load_certs_from_pemder
    from pyhanko.pdf_utils.reader import PdfFileReader
    from pyhanko.sign.fields import enumerate_sig_fields
    from pyhanko.sign.validation import EmbeddedPdfSignature, validate_pdf_signature, validate_pdf_timestamp
    from pyhanko.sign.validation.settings import KeyUsageConstraints
    from pyhanko_certvalidator import ValidationContext

    roots = []
    if sys.platform == "win32":
        for der, encoding, _purpose in ssl.enum_certificates("ROOT"):
            if encoding == "x509_asn":
                try:
                    roots.append(x509.Certificate.load(der))
                except ValueError:
                    continue
    custom = request.get("trust", "")
    extra = list(load_certs_from_pemder([custom])) if custom else []
    if custom and not extra:
        raise UserError("Soubor s důvěryhodným kořenem neobsahuje certifikát PEM ani DER.")

    with src.open("rb") as file:
        reader = PdfFileReader(file)
        unlock_reader(reader, request.get("pdf_password", ""))
        signatures = list(enumerate_sig_fields(reader, filled_status=True))
        if not signatures:
            return "V PDF nebyl nalezen žádný vložený digitální podpis."
        result = [f"Podpisy a dokumentová časová razítka ({len(signatures)})",
                  "Offline kontrola. Aktuální stav odvolání u vydavatele se neověřuje."]
        coverage = {
            "ENTIRE_FILE": "celý aktuální soubor",
            "ENTIRE_REVISION": "dřívější úplná revize",
            "CONTIGUOUS_BLOCK_FROM_START": "neúplný rozsah",
            "UNCLEAR": "nejasný rozsah",
        }
        for name, _, field in signatures:
            result.extend(("", str(name)))
            try:
                sig = EmbeddedPdfSignature(reader, field, name)
                context = ValidationContext(
                    trust_roots=roots if sys.platform == "win32" else None,
                    extra_trust_roots=extra,
                    allow_fetching=False,
                    revocation_mode="soft-fail",
                )
                timestamp = sig.sig_object.get("/Type") == "/DocTimeStamp"
                if timestamp:
                    status = validate_pdf_timestamp(sig, context)
                    is_valid = status.intact and status.valid and status.trusted and status.docmdp_ok
                else:
                    # Let either signing key usage qualify, including certificates with both.
                    status = validate_pdf_signature(
                        sig, context,
                        key_usage_settings=KeyUsageConstraints(
                            key_usage={"digital_signature", "non_repudiation"}, match_all_key_usages=False
                        ),
                    )
                    is_valid = status.bottom_line
                cert = sig.signer_cert
                result.extend([
                    f"  {'Časové razítko' if timestamp else 'Podepisující'}  {cert.subject.human_friendly}",
                    f"  Vydavatel  {cert.issuer.human_friendly}",
                    f"  Integrita podpisu  {yes_no(status.intact)}",
                    f"  Kryptograficky platný  {yes_no(status.valid)}",
                    f"  Důvěryhodný řetězec certifikátů  {yes_no(status.trusted)}",
                    f"  Rozsah podpisu  {coverage.get(getattr(status.coverage, 'name', ''), 'nelze určit')}",
                    f"  Změny v souladu s pravidly podpisu  {yes_no(status.docmdp_ok)}",
                    f"  Výsledek offline kontroly  {'ověřeno' if is_valid else 'neověřeno'}",
                ])
                time = getattr(status, "timestamp", None) if timestamp else sig.self_reported_timestamp
                if time:
                    title = "Čas razítka" if timestamp else "Čas uvedený podepisujícím"
                    result.append(f"  {title}  {time.isoformat(sep=' ', timespec='seconds')}")
                time_status = getattr(status, "timestamp_validity", None)
                if time_status is not None:
                    ts_valid = time_status.intact and time_status.valid and time_status.trusted
                    result.append(f"  Ověřené časové razítko  {yes_no(ts_valid)}")
            except Exception as exc:
                # A damaged or unsupported signature must not hide other fields.
                result.append(f"  Podpis se nepodařilo ověřit ({type(exc).__name__}).")
    return "\n".join(result)


def selected_pages(view, count):
    if "pages" not in view:
        return None
    pages = view["pages"]
    if (not isinstance(pages, list) or not pages
            or any(type(number) is not int or not 1 <= number <= count for number in pages)):
        raise UserError("Výběr stran obsahuje neplatné číslo stránky.")
    return sorted(set(pages))


def document_order(view, pages):
    order = view.get("order", list(range(1, pages + 1)))
    if (not isinstance(order, list) or not order or len(order) > pages
            or any(type(page) is not int or not 1 <= page <= pages for page in order)
            or len(set(order)) != len(order)):
        raise UserError("Pořadí stran neodpovídá otevřenému PDF.")
    return order


def page_rotations(rotations, count):
    if (not isinstance(rotations, list) or len(rotations) != count
            or any(type(angle) is not int or angle not in (0, 90, 180, 270) for angle in rotations)):
        raise UserError("Otočení stran má neplatný formát.")
    return rotations


def apply_page_order(pdf, pages, order):
    if order == list(range(1, len(pages) + 1)):
        return False
    selected = [pages[number - 1] for number in order]
    labels = [page.label for page in selected] if "/PageLabels" in pdf.Root else None
    # Reuse the page objects so links can still target retained pages.
    del pdf.pages[:]
    pdf.pages.extend(selected)
    if labels is not None:
        numbers = []
        for index, label in enumerate(labels):
            numbers.extend([index, pikepdf.Dictionary(P=label)])
        pdf.Root.PageLabels = pikepdf.Dictionary(Nums=pikepdf.Array(numbers))
    return True


def inspect_document(src, request, *, raster=False):
    view = request.get("view", {})
    password = request.get("pdf_password", "")
    action = request["action"]
    with open_pdf(src, password) as pdf:
        if not len(pdf.pages):
            raise UserError("Vstupní PDF nemá žádné strany.")
        if action == "describe":
            return describe(pdf)
        if action in ("screen", "thumbnail") and raster:
            rendered = render_screen(pdf, src, view, password, thumbnail=action == "thumbnail")
            if rendered is not None:
                return rendered
        exe = ghostscript_path(request)
        if action in ("screen", "thumbnail", "print_page"):
            return screen_page(pdf, src, view, exe, run_ghostscript, stage_for_render,
                               printed=action == "print_page", thumbnail=action == "thumbnail")
        if action == "scan":
            order = document_order(view, len(pdf.pages))
            settings = export_settings(request.get("export"))
            return scan_plates(pdf, src, {**view, "order": order, "proof_dpi": settings["proof_dpi"]},
                               exe, run_ghostscript, stage_for_render)
        return page_proof(pdf, src, view, exe, run_ghostscript, stage_for_render)


def save_page_edits(src, request):
    view = request.get("view", {})
    password = request.get("pdf_password", "")
    with output_file(src, request) as pending, open_pdf(src, password) as pdf:
        pages = list(pdf.pages)
        order = document_order(view, len(pages))
        rotations = page_rotations(view.get("rotations", []), len(pages))
        prepare_edit(pdf, request)
        for page, angle in zip(pages, rotations):
            if angle:
                page.rotate(angle, relative=True)
        if len(order) == len(pages):
            apply_page_order(pdf, pages, order)
            pdf.save(pending, encryption=pdf.is_encrypted, fix_metadata_version=False)
        else:
            with (open_pdf(src, password) if pdf.is_encrypted else pikepdf.Pdf.new()) as kept:
                copy_page_subset(pdf, kept, [number - 1 for number in order],
                                 subset_attributes(pdf), pending, keep_document_options=True)
    return "Uložena upravená kopie PDF."


def save_document(src, request):
    if "rotations" in request.get("view", {}):
        save_page_edits(src, request)
    else:
        with output_file(src, request) as pending:
            shutil.copyfile(src, pending)
    return "PDF bylo uloženo."


def subset_attributes(pdf):
    return {key: pdf.make_indirect(pdf.Root[key]) for key in ("/OutputIntents", "/Metadata")
            if isinstance(pdf.Root.get(key), (pikepdf.Array, pikepdf.Stream))}


def copy_page_subset(source, target, indices, attributes, destination, *, rotation=0, keep_document_options=False):
    if source.is_encrypted:
        # Retain qpdf's original encryption parameters while rebuilding the catalog.
        del target.pages[:]
        for key in set(target.Root.keys()) - {"/Type", "/Pages"}:
            del target.Root[key]
    target.add_pages_from(source, indices)
    if rotation:
        target.pages[0].rotate(rotation, relative=True)
    if "/Lang" in source.Root:
        target.Root.Lang = str(source.Root.Lang)
    if "/PageLabels" in source.Root:
        labels = []
        for position, index in enumerate(indices):
            labels.extend([position, pikepdf.Dictionary(P=source.pages[index].label)])
        target.Root.PageLabels = pikepdf.Dictionary(Nums=pikepdf.Array(labels))
    for key, attribute in attributes.items():
        target.Root[key] = target.copy_foreign(attribute)
    if keep_document_options:
        for key in ("/PageLayout", "/PageMode"):
            if isinstance(source.Root.get(key), pikepdf.Name):
                target.Root[key] = source.Root[key]
        preferences = source.Root.get("/ViewerPreferences")
        if isinstance(preferences, pikepdf.Dictionary):
            target.Root.ViewerPreferences = target.copy_foreign(source.make_indirect(preferences))
        names = source.Root.get("/Names", {})
        files = names.get("/EmbeddedFiles") if isinstance(names, pikepdf.Dictionary) else None
        if isinstance(files, pikepdf.Dictionary):
            if "/Names" not in target.Root:
                target.Root.Names = pikepdf.Dictionary()
            target.Root.Names.EmbeddedFiles = target.copy_foreign(source.make_indirect(files))
        associated_files = source.Root.get("/AF")
        if isinstance(associated_files, pikepdf.Array):
            target.Root.AF = target.copy_foreign(source.make_indirect(associated_files))
    target.docinfo = target.copy_foreign(source.docinfo)
    # Document-level conformance does not carry over automatically to an extracted page.
    clear_pdfx_claim(target)
    if "/Metadata" in target.Root:
        with target.open_metadata(set_pikepdf_as_editor=False, update_docinfo=False) as metadata:
            for namespace in ("pdfaid", "pdfuaid"):
                for name in ("part", "conformance", "rev", "amd", "corr"):
                    key = f"{namespace}:{name}"
                    if key in metadata:
                        del metadata[key]
    target.save(destination, encryption=source.is_encrypted,
                min_version=(source.pdf_version, source.extension_level), fix_metadata_version=False)


def same_file_state(original, current):
    return (os.path.samestat(original, current)
            and original.st_size == current.st_size and original.st_mtime_ns == current.st_mtime_ns)


def publish_split_pages(staged):
    published = []
    try:
        for draft, target in staged:
            original = draft.stat()
            publish_new_file(draft, target)
            published.append((target, original))
    except OSError as failure:
        remaining = []
        for output, original in published:
            try:
                if same_file_state(original, output.lstat()):
                    output.unlink()
            except FileNotFoundError:
                pass
            except OSError:
                remaining.append(output.name)
        if remaining:
            raise UserError("Rozdělení se nepodařilo dokončit. Ve složce zůstaly soubory: "
                            + ", ".join(remaining[:5])) from failure
        if isinstance(failure, FileExistsError):
            raise UserError(f"Soubor {target.name} už existuje. Vyberte jinou složku.") from failure
        raise


def split_document(src, request):
    destination = Path(request.get("output_dir") or "")
    if not request.get("output_dir") or not destination.is_dir():
        raise UserError("Vyberte existující výstupní složku.")
    view = request.get("view", {})
    password = request.get("pdf_password", "")
    original = src.stat()
    with open_pdf(src, password) as pdf:
        count = len(pdf.pages)
        if not count:
            raise UserError("Vstupní PDF nemá žádné strany.")
        order = document_order(view, count)
        rotations = page_rotations(view.get("rotations", [0] * count), count)
        pages = selected_pages(view, len(order)) or range(1, len(order) + 1)
        width = max(3, len(str(len(order))))
        outputs = [(order[number - 1] - 1, destination / f"{src.stem}-strana-{number:0{width}d}.pdf")
                   for number in pages]
        for _, target in outputs:
            if os.path.lexists(target):
                raise UserError(f"Soubor {target.name} už existuje. Vyberte jinou složku.")

        prepare_edit(pdf, request)
        attributes = subset_attributes(pdf)
        with tempfile.TemporaryDirectory(prefix="pdftisk-", dir=destination) as folder:
            staged = []
            for index, target in outputs:
                draft = Path(folder) / target.name
                with (open_pdf(src, password) if pdf.is_encrypted else pikepdf.Pdf.new()) as single:
                    copy_page_subset(pdf, single, [index], attributes, draft, rotation=rotations[index])
                staged.append((draft, target))
            if not same_file_state(original, src.stat()):
                raise UserError("Původní PDF se mezitím změnilo. Otevřete je znovu.")
            publish_split_pages(staged)
    return f"Počet uložených souborů PDF: {len(outputs)}."


def merge_sources(request, minimum):
    entries = request.get("merge_inputs")
    if (not isinstance(entries, list) or len(entries) < minimum
            or any(not isinstance(entry, dict) or not isinstance(entry.get("path"), str)
                   or not entry["path"] for entry in entries)):
        raise UserError("Vyberte PDF soubory ke sloučení.")
    return entries


def merge_source_state(path, entry):
    if not path.is_file():
        raise UserError(f"Soubor {path.name} nelze otevřít.")
    current = path.stat()
    if "source_size" in entry and (current.st_size, current.st_mtime_ns) != (
            entry["source_size"], entry.get("source_mtime_ns")):
        raise UserError(f"Soubor {path.name} se změnil. Přidejte jej znovu.")
    return current


def inspect_merge(request):
    result = []
    for entry in merge_sources(request, 1):
        path = Path(entry["path"])
        try:
            merge_source_state(path, entry)
            with pikepdf.Pdf.open(path, password=entry.get("pdf_password", "")) as pdf:
                pages = len(pdf.pages)
                result.append({"pages": pages, "message": "" if pages else "PDF neobsahuje žádnou stranu."})
        except pikepdf.PasswordError:
            result.append({"message": "Zadejte platné heslo a načtěte počet stran."})
        except UserError as exc:
            result.append({"message": str(exc)})
        except (OSError, pikepdf.PdfError):
            result.append({"message": f"Soubor {path.name} se nepodařilo načíst jako PDF."})
    return {"merge_info": result}


def merge_intents(pdf):
    intents = pdf.Root.get("/OutputIntents")
    if not intents:
        return (), None
    if not isinstance(intents, pikepdf.Array):
        return None, None
    fingerprints = []
    for intent in intents:
        if not isinstance(intent, pikepdf.Dictionary) or not isinstance(intent.get("/DestOutputProfile"), pikepdf.Stream):
            return None, None
        profile = intent.DestOutputProfile.read_bytes()
        fingerprints.append(tuple(str(intent.get(key, "")) for key in (
            "/S", "/OutputConditionIdentifier", "/RegistryName")) + (sha256(profile).digest(),))
    return tuple(fingerprints), intents


def append_merge_source(target, pdf, entry):
    count = len(pdf.pages)
    view = entry.get("view", {})
    order = document_order(view, count)
    rotations = page_rotations(view.get("rotations", [0] * count), count)
    positions = selected_pages(view, len(order)) or range(1, len(order) + 1)
    indices = [order[position-1]-1 for position in positions]
    first = len(target.pages)
    target.add_pages_from(pdf, indices)
    for position, index in enumerate(indices, first):
        if rotations[index]:
            target.pages[position].rotate(rotations[index], relative=True)


def merge_documents(request):
    entries = merge_sources(request, 2)
    paths = [Path(entry["path"]) for entry in entries]
    states = [merge_source_state(path, entry) for path, entry in zip(paths, entries)]
    with output_file(paths[0], request, protected=paths) as pending, pikepdf.Pdf.new() as target:
        version = ("1.3", 0)
        shared_fingerprint, shared_intents, mixed = (), None, False
        for index, (path, entry) in enumerate(zip(paths, entries)):
            try:
                with open_pdf(path, entry.get("pdf_password", "")) as pdf:
                    if pdf.is_encrypted and not request.get("allow_unencrypted_merge"):
                        raise MergeEncryption("Vstupní PDF je šifrované.")
                    prepare_edit(pdf, request)
                    append_merge_source(target, pdf, entry)
                    version = max(version, (pdf.pdf_version, pdf.extension_level))
                    fingerprint, intents = merge_intents(pdf)
                    if index == 0:
                        shared_fingerprint = fingerprint
                        if intents is not None:
                            shared_intents = target.copy_foreign(pdf.make_indirect(intents))
                    mixed |= fingerprint is None or fingerprint != shared_fingerprint
            except UserError as exc:
                raise type(exc)(f"{path.name}: {exc}") from exc
            except (OSError, pikepdf.PdfError) as exc:
                raise UserError(f"Soubor {path.name} se nepodařilo přidat do sloučeného PDF.") from exc
        if shared_intents is not None and not mixed:
            target.Root.OutputIntents = shared_intents
        count = len(target.pages)
        target.save(pending, min_version=version, fix_metadata_version=False)
        for path, state in zip(paths, states):
            if not same_file_state(state, path.stat()):
                raise UserError(f"Soubor {path.name} se během sloučení změnil.")
    message = f"PDF bylo sloučeno. Počet stran {count}."
    if mixed:
        message += "\nVstupy nemají shodný vložený výstupní profil ICC."
    return message


def dispatch(request, *, raster=False):
    if request.get("action") == "merge_probe":
        return inspect_merge(request)
    if request.get("action") == "merge":
        return merge_documents(request)
    src = Path(request["input"])
    if not src.is_file():
        raise UserError("Vyberte existující soubor PDF.")
    operation = {"transform": transform, "sign": sign, "verify": verify,
                 "preview": inspect_document, "scan": inspect_document, "describe": inspect_document,
                 "screen": inspect_document, "thumbnail": inspect_document, "print_page": inspect_document,
                 "edit_pages": save_page_edits, "save_document": save_document,
                 "split": split_document}.get(request.get("action"))
    if operation is None:
        raise UserError("Neznámá akce.")
    if operation is inspect_document:
        return inspect_document(src, request, raster=raster)
    view = request.get("view", {})
    if request["action"] in ("transform", "sign") and "rotations" in view:
        # Keep draft edits in the same transaction as export or signing.
        with tempfile.TemporaryDirectory(prefix="pdf-draft-") as folder, output_file(src, request) as pending:
            draft = Path(folder) / "draft.pdf"
            save_page_edits(src, {**request, "output": str(draft), "overwrite": False})
            return operation(draft, {**request, "output": str(pending), "overwrite": False})
    return operation(src, request)


def respond(payload, *, raster=False):
    pixels = None
    try:
        result = dispatch(json.loads(payload), raster=raster)
        if raster and isinstance(result, dict):
            pixels = result.pop("raster", None)
        response = {"status": "ok", "message": result} if isinstance(result, str) else {"status": "ok", "view": result}
    except MergeEncryption as exc:
        response = {"status": "merge_encrypted", "message": str(exc)}
    except SignedPDF as exc:
        response = {"status": "signed_pdf", "message": str(exc)}
    except (UserError, ProfileError, ProofError, PDFXError, ColorConversionError) as exc:
        response = {"status": "error", "message": str(exc)}
    except FileExistsError:
        response = {"status": "error", "message": "Cílový soubor už existuje. Vyberte nový název."}
    except PermissionError:
        response = {"status": "error", "message": "Soubor nebo složku nelze otevřít. "
                    "Zkontrolujte přístup a zavřete soubor v jiné aplikaci."}
    except Exception as exc:
        response = {"status": "error", "message": f"PDF se nepodařilo zpracovat ({type(exc).__name__}).\n{str(exc)[:1600]}"}
    return json.dumps(response, ensure_ascii=True), pixels


def main():
    logging.disable(logging.CRITICAL)
    args = sys.argv[1:]
    raster = args == ["--reader-stream"]
    source, output = sys.stdin.buffer, sys.stdout.buffer
    for line in source if raster or args == ["--stream"] else (source.read(),):
        header, pixels = respond(line, raster=raster)
        output.write(header.encode("ascii") + b"\n")
        if pixels is not None:
            output.write(pixels)
        output.flush()
        del header, pixels


if __name__ == "__main__":
    main()
