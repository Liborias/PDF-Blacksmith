"""Object-preserving colour conversion primitives for PDF Blacksmith.

No raster fallback is performed here. Unsupported PDF constructs fail explicitly
instead of silently producing a false 'CMYK K only' claim.
"""
import datetime
import math
import struct
from pathlib import Path

import pikepdf
from PIL import Image


class ColorConversionError(ValueError):
    pass


def _fixed(value):
    return struct.pack(">i", round(value * 65536))


def gray_gamma_icc(gamma):
    """Generate a genuine ICC v2 one-channel Gray Gamma profile (D50 PCS)."""
    if gamma not in (1.8, 2.2):
        raise ColorConversionError("Nepodporovaná gamma šedého profilu.")
    d50 = b"".join(_fixed(v) for v in (0.9642, 1, 0.8249))
    name = ("PDF Blacksmith Gray Gamma " + str(gamma)).encode("ascii")
    entries = [
        (b"wtpt", b"XYZ " + b"\0"*4 + d50),
        (b"kTRC", b"curv" + b"\0"*4 + struct.pack(">IH", 1, round(gamma*256))),
        (b"desc", b"desc" + b"\0"*4 + struct.pack(">I", len(name)+1) + name + b"\0" + bytes(8)),
    ]
    header = bytearray(128)
    header[4:8] = b"PBsm"
    header[8:12] = b"\x02\x10\0\0"
    header[12:16], header[16:20], header[20:24] = b"mntr", b"GRAY", b"XYZ "
    now = datetime.datetime.now(datetime.timezone.utc)
    header[24:36] = struct.pack(">6H", now.year, now.month, now.day, now.hour, now.minute, now.second)
    header[36:40], header[40:44] = b"acsp", b"MSFT"
    header[64:68], header[68:80], header[80:84] = struct.pack(">I", 1), d50, b"PBsm"
    table, binary = struct.pack(">I", len(entries)), b""
    offset = 128 + 4 + 12*len(entries)
    for signature, body in entries:
        table += signature + struct.pack(">II", offset+len(binary), len(body))
        binary += body + b"\0"*((-len(body)) % 4)
    result = bytes(header) + table + binary
    return struct.pack(">I", len(result)) + result[4:]


def resolve_gray_icc(key, custom=""):
    """Return (bytes|None, label). 'automatic' uses Ghostscript defaults."""
    if key in ("", "automatic"):
        return None, "Automaticky"
    if key in ("gamma18", "gamma22"):
        gamma = 1.8 if key == "gamma18" else 2.2
        return gray_gamma_icc(gamma), "Gray Gamma " + str(gamma)
    if key == "custom":
        path = Path(custom)
        if not path.is_file() or path.stat().st_size > 32_000_000:
            raise ColorConversionError("Vyberte platný šedý profil ICC (.icc/.icm).")
        data = path.read_bytes()
        if not (128 <= len(data) <= 32_000_000 and int.from_bytes(data[:4], "big") == len(data)
                and data[16:20] == b"GRAY" and data[36:40] == b"acsp"):
            raise ColorConversionError("Zvolený ICC profil není jednosložkový Gray.")
        return data, path.stem
    raise ColorConversionError("Neznámý šedý ICC profil.")


def _walk_streams(pdf):
    """Yield page and Form XObject content streams, each only once."""
    seen = set()

    def walk(obj):
        reference = obj.objgen if obj.is_indirect else id(obj)
        if reference in seen:
            return
        seen.add(reference)
        yield obj
        resources = obj.get("/Resources", {})
        xobjects = resources.get("/XObject", {})
        for sub in xobjects.values():
            if sub.get("/Subtype") == pikepdf.Name.Form:
                yield from walk(sub)

    for page in pdf.pages:
        yield from walk(page.obj)


def _replace_instructions(pdf, obj, conversion):
    instructions = pikepdf.parse_content_stream(obj)
    updated, replacements = [], 0
    for args, op in instructions:
        new = conversion(args, str(op))
        if new is not None:
            updated.append(new)
            replacements += 1
        else:
            updated.append((args, op))
    if replacements:
        data = pikepdf.unparse_content_stream(updated)
        if obj.get("/Type") == pikepdf.Name.Page:
            obj.Contents = pdf.make_stream(data)
        else:
            obj.write(data)
    return replacements


def convert_rich_black(pdf, threshold=0.999):
    """Remove CMY from DeviceCMYK rich black vector/text paints only.

    Photos, spot colours, ICCBased values and otherwise-coloured objects are
    deliberately left unchanged (rather than being misclassified as black).
    """
    def replacement(args, operator):
        if operator not in ("k", "K") or len(args) != 4:
            return None
        c, m, y, black = map(float, args)
        if black >= threshold and max(c, m, y) > 0:
            return ([0, 0, 0, 1], pikepdf.Operator(operator))
        return None
    return sum(_replace_instructions(pdf, obj, replacement) for obj in _walk_streams(pdf))


def gray_to_explicit_k(pdf):
    """Convert grayscale vector operators and 8-bit grayscale images to DeviceCMYK.

    Abort for unsupported painting operators or shadings. This prevents
    advertising an explicit-CMYK output when some elements remain DeviceGray.
    """
    vector_count = image_count = 0
    for obj in _walk_streams(pdf):
        def replacement(args, operator):
            if operator in ("g", "G") and len(args) == 1:
                gray = min(1, max(0, float(args[0])))
                return ([0, 0, 0, round(1-gray, 8)], pikepdf.Operator("k" if operator == "g" else "K"))
            return None

        vector_count += _replace_instructions(pdf, obj, replacement)
        for args, op in pikepdf.parse_content_stream(obj):
            operator = str(op)
            if operator in ("g", "G", "rg", "RG", "sc", "SC", "scn", "SCN", "sh"):
                raise ColorConversionError("PDF obsahuje barvu nebo stínování, které nelze bezpečně převést do explicitního CMYK. Zvolte DeviceGray, nebo upravte problematický objekt.")
            if operator in ("k", "K") and (len(args) != 4 or any(abs(float(v)) > 0.000001 for v in args[:3])):
                raise ColorConversionError("V převedeném PDF zůstala nenulová složka CMY.")
            if operator in ("cs", "CS") and len(args) == 1 and args[0] != pikepdf.Name.DeviceCMYK:
                raise ColorConversionError("V PDF zůstal výběr jiného barevného prostoru než CMYK.")

    seen_images = set()
    for obj in _walk_streams(pdf):
        resources = obj.get("/Resources", {})
        for img in resources.get("/XObject", {}).values():
            if img.get("/Subtype") != pikepdf.Name.Image:
                continue
            ident = img.objgen if img.is_indirect else id(img)
            if ident in seen_images:
                continue
            seen_images.add(ident)
            if img.get("/ImageMask", False):
                continue
            if img.get("/ColorSpace") != pikepdf.Name.DeviceGray:
                raise ColorConversionError("Výstup obsahuje obrázek mimo prostor DeviceGray; přímý převod do K nebyl proveden.")
            if int(img.get("/BitsPerComponent", 8)) > 8 or "/Mask" in img:
                raise ColorConversionError("Obrázek používá masku nebo více než 8 bitů na vzorek; bezpečný převod do K není podporován.")
            gray = pikepdf.PdfImage(img).as_pil_image().convert("L")
            values = gray.tobytes()
            cmyk = bytearray(len(values)*4)
            cmyk[3::4] = bytes(255-value for value in values)
            img.write(bytes(cmyk), filter=pikepdf.Name.FlateDecode)
            img.ColorSpace = pikepdf.Name.DeviceCMYK
            img.BitsPerComponent = 8
            if "/Decode" in img:
                del img["/Decode"]
            if "/DecodeParms" in img:
                del img["/DecodeParms"]
            image_count += 1
    return vector_count, image_count


def attach_gray_output_intent(pdf, data, label):
    stream = pdf.make_stream(data, N=1, Alternate=pikepdf.Name.DeviceGray)
    pdf.Root.OutputIntents = pikepdf.Array([pdf.make_indirect(pikepdf.Dictionary(
        Type=pikepdf.Name.OutputIntent, S=pikepdf.Name.GTS_PDFX,
        OutputConditionIdentifier="CUSTOM_GRAY", OutputCondition=label,
        Info=label, DestOutputProfile=stream,
    ))])
