"""Local PDF operations for PDF Blacksmith. The Windows front end uses JSON on stdin/stdout."""

from contextlib import contextmanager
import json
import logging
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

import pikepdf
from print_profiles import ProfileError, attach_intent, get_profile, source_cmyk
from export_settings import export_settings, pdfwrite_settings
from proof import ProofError, describe, page_proof, scan_plates, screen_page


class UserError(Exception):
    pass


class SignedPDF(UserError):
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


@contextmanager
def output_file(src, request):
    dst = Path(request["output"])
    if dst.suffix.lower() != ".pdf":
        raise UserError("Výstupní soubor musí mít příponu .pdf.")
    original = None
    same_path = dst.resolve() == src.resolve()
    if same_path or (dst.exists() and dst.samefile(src)):
        if not (same_path and request["action"] == "save_document" and request.get("replace_source")
                and request.get("overwrite")):
            raise UserError("Pro výstup vyberte jiný soubor než vstupní PDF.")
        dst = src.resolve(strict=True)
        original = dst.stat()
        if ((original.st_size, original.st_mtime_ns)
                != (request.get("source_size"), request.get("source_mtime_ns"))):
            raise UserError("Původní PDF se mezitím změnilo. Otevřete je znovu.")
    with tempfile.TemporaryDirectory(prefix="pdftisk-", dir=dst.parent) as folder:
        pending = Path(folder) / "vystup.pdf"
        yield pending
        if original is not None:
            current = dst.stat()
            if (not os.path.samestat(original, current)
                    or (original.st_size, original.st_mtime_ns) != (current.st_size, current.st_mtime_ns)):
                raise UserError("Původní PDF se mezitím změnilo. Otevřete je znovu.")
        if request.get("overwrite"):
            os.replace(pending, dst)
        elif os.name == "nt":
            os.rename(pending, dst)  # Fails if the destination has appeared.
        else:
            os.link(pending, dst)


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
    gray = request.get("gray", False)
    outline = request.get("outline", False)
    profile_key = request.get("profile", "")
    if gray and profile_key:
        raise UserError("Vyberte buď desku K, nebo výstupní profil CMYK.")
    convert = gray or outline or bool(profile_key)
    if not (convert or request.get("decrypt")):
        raise UserError("Vyberte alespoň jednu úpravu PDF.")
    options = export_settings(request.get("export"))
    proof_dpi, repair_dpi = options["proof_dpi"], options["repair_dpi"]
    exe = ghostscript_path(request) if convert else None

    with open_pdf(src, request.get("pdf_password", "")) as pdf:
        modified = prepare_edit(pdf, request)
        if not convert:
            if not pdf.is_encrypted:
                raise UserError("Vstupní PDF už je bez šifrování a omezení heslem.")
            restricted_only = pdf.user_password_matched and not request.get("pdf_password")
            with output_file(src, request) as pending:
                pdf.save(pending, encryption=False, fix_metadata_version=False)
            if restricted_only:
                return "Uložena kopie bez omezení tisku a úprav. Heslo vlastníka nebylo potřeba."
            return "Uložena kopie PDF bez hesla."

        pages = len(pdf.pages)
        if not pages:
            raise UserError("Vstupní PDF nemá žádné strany.")
        repaired = []
        profile = get_profile(profile_key, request.get("custom_icc", "")) if profile_key else None
        with tempfile.TemporaryDirectory(prefix="pdftisk-") as tmp:
            folder = Path(tmp)
            staged = folder / "vstup.pdf"
            if pdf.is_encrypted or modified:
                pdf.save(staged, encryption=False, fix_metadata_version=False)
            else:
                shutil.copyfile(src, staged)

            args = [
                "-sDEVICE=pdfwrite", "-dCompatibilityLevel=1.7", "-dAutoRotatePages=/None",
                "-dPreserveAnnots=false", "-dPrinted=true", "-sOutputFile=hotovo.pdf", *pdfwrite_settings(options),
            ]
            if gray:
                args.extend([
                    "-sColorConversionStrategy=Gray", "-dProcessColorModel=/DeviceGray",
                    "-sBlendConversionStrategy=Simple", "-dPreserveSeparation=false", "-dPreserveDeviceN=false",
                ])
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
            if outline:
                args.append("-dNoOutputFonts=true")
            log = run_ghostscript(exe, folder, [*args, "-f", "vstup.pdf"])

            with pikepdf.Pdf.open(folder / "hotovo.pdf") as finished:
                if len(finished.pages) != pages:
                    raise UserError("Počet stran v převedeném PDF se liší od vstupu.")
                if gray:
                    revised = convert_spot_vectors(finished)
                    if "/OutputIntents" in finished.Root:
                        # A source CMYK output intent is inapplicable to a DeviceGray file.
                        del finished.Root["/OutputIntents"]
                        revised = True
                    if revised:
                        finished.save(folder / "vystup-upraveny.pdf", fix_metadata_version=False)
                elif profile:
                    attach_intent(finished, *profile)
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
    if gray:
        lines.append(f"Separace při {proof_dpi} dpi ověřeny na všech stranách. C, M, Y i případné přímé desky jsou prázdné.")
    if repaired:
        lines.append(f"Rastrová oprava {repair_dpi} dpi na stranách " + ", ".join(map(str, repaired)) + ".")
    if profile:
        lines.append(f"Výstupní profil CMYK: {profile[1]}. Profil ICC je vložen v PDF.")
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


def document_order(view, pages):
    order = view.get("order", list(range(1, pages + 1)))
    if (not isinstance(order, list) or len(order) != pages
            or any(type(page) is not int for page in order)
            or set(order) != set(range(1, pages + 1))):
        raise UserError("Pořadí stran neodpovídá otevřenému PDF.")
    return order


def inspect_document(src, request):
    view = request.get("view", {})
    with open_pdf(src, request.get("pdf_password", "")) as pdf:
        if not len(pdf.pages):
            raise UserError("Vstupní PDF nemá žádné strany.")
        if request["action"] == "describe":
            return describe(pdf)
        exe = ghostscript_path(request)
        if request["action"] in ("screen", "print_page"):
            return screen_page(pdf, src, view, exe, run_ghostscript, stage_for_render,
                               printed=request["action"] == "print_page")
        if request["action"] == "scan":
            order = document_order(view, len(pdf.pages))
            settings = export_settings(request.get("export"))
            return scan_plates(pdf, src, {**view, "order": order, "proof_dpi": settings["proof_dpi"]},
                               exe, run_ghostscript, stage_for_render)
        return page_proof(pdf, src, view, exe, run_ghostscript, stage_for_render)


def save_page_edits(src, request):
    view = request.get("view", {})
    with output_file(src, request) as pending, open_pdf(src, request.get("pdf_password", "")) as pdf:
        pages = list(pdf.pages)
        order = document_order(view, len(pages))
        rotations = view.get("rotations", [])
        if (not isinstance(rotations, list) or len(rotations) != len(pages)
                or any(type(angle) is not int or angle not in (0, 90, 180, 270) for angle in rotations)):
            raise UserError("Otočení stran má neplatný formát.")
        prepare_edit(pdf, request)
        for page, angle in zip(pages, rotations):
            if angle:
                page.rotate(angle, relative=True)
        if order != list(range(1, len(pages) + 1)):
            labels = [page.label for page in pages] if "/PageLabels" in pdf.Root else None
            # Remove before adding so QPDF reuses the page objects that bookmarks target.
            del pdf.pages[:]
            pdf.pages.extend(pages[number - 1] for number in order)
            if labels is not None:
                numbers = pikepdf.Array()
                for index, number in enumerate(order):
                    numbers.extend([index, pikepdf.Dictionary(P=labels[number - 1])])
                pdf.Root.PageLabels = pikepdf.Dictionary(Nums=numbers)
        pdf.save(pending, encryption=pdf.is_encrypted, fix_metadata_version=False)
    return "Uložena upravená kopie PDF."


def save_document(src, request):
    if "rotations" in request.get("view", {}):
        save_page_edits(src, request)
    else:
        with output_file(src, request) as pending:
            shutil.copyfile(src, pending)
    return "PDF bylo uloženo."


def dispatch(request):
    src = Path(request["input"])
    if not src.is_file():
        raise UserError("Vyberte existující soubor PDF.")
    operation = {"transform": transform, "sign": sign, "verify": verify,
                 "preview": inspect_document, "scan": inspect_document, "describe": inspect_document,
                 "screen": inspect_document, "print_page": inspect_document,
                 "edit_pages": save_page_edits, "save_document": save_document}.get(request.get("action"))
    if operation is None:
        raise UserError("Neznámá akce.")
    view = request.get("view", {})
    if request["action"] in ("transform", "sign") and "rotations" in view:
        # Keep draft edits in the same transaction as export or signing.
        with tempfile.TemporaryDirectory(prefix="pdf-draft-") as folder, output_file(src, request) as pending:
            draft = Path(folder) / "draft.pdf"
            save_page_edits(src, {**request, "output": str(draft), "overwrite": False})
            return operation(draft, {**request, "output": str(pending), "overwrite": False})
    return operation(src, request)


def respond(payload):
    try:
        result = dispatch(json.loads(payload))
        response = {"status": "ok", "message": result} if isinstance(result, str) else {"status": "ok", "view": result}
    except SignedPDF as exc:
        response = {"status": "signed_pdf", "message": str(exc)}
    except (UserError, ProfileError, ProofError) as exc:
        response = {"status": "error", "message": str(exc)}
    except FileExistsError:
        response = {"status": "error", "message": "Cílový soubor už existuje. Vyberte nový název."}
    except PermissionError:
        response = {"status": "error", "message": "Soubor nebo složku nelze otevřít. "
                    "Zkontrolujte přístup a zavřete soubor v jiné aplikaci."}
    except Exception as exc:
        response = {"status": "error", "message": f"PDF se nepodařilo zpracovat ({type(exc).__name__}).\n{str(exc)[:1600]}"}
    return json.dumps(response, ensure_ascii=True)


def main():
    logging.disable(logging.CRITICAL)
    if sys.argv[1:] == ["--stream"]:
        for line in sys.stdin.buffer:
            print(respond(line), flush=True)
    else:
        print(respond(sys.stdin.buffer.read()))


if __name__ == "__main__":
    main()
