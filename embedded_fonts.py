"""Recover Unicode and Windows-compatible metadata from embedded TrueType subsets."""

from hashlib import sha256
from io import BytesIO
import re

from fontTools.fontBuilder import FontBuilder
from fontTools.ttLib import TTFont, TTLibError
import pikepdf as pdf


def unicode_codes(font):
    stream = font.get("/ToUnicode")
    if not isinstance(stream, pdf.Stream):
        return {}
    data = stream.read_bytes()
    if len(data) > 1 << 20:
        return {}
    result = {}
    for kind, body in re.findall(rb"begin(bfchar|bfrange)\b(.*?)end\1\b", data, re.S):
        if kind == b"bfchar":
            sources = ((int(code, 16), value) for code, value in re.findall(rb"<([\da-f]+)>\s*<([\da-f]+)>", body, re.I))
        else:
            ranges = re.findall(rb"<([\da-f]+)>\s*<([\da-f]+)>\s*(<[^>]+>|\[[^]]*\])", body, re.I | re.S)

            def expand():
                for first, last, target in ranges:
                    start, stop = int(first, 16), int(last, 16)
                    if stop < start or stop-start > 65535:
                        continue
                    if target.startswith(b"["):
                        yield from enumerate(re.findall(rb"<([\da-f]+)>", target, re.I)[:stop-start+1], start)
                    else:
                        value = target[1:-1]
                        base = int(value, 16)
                        for index in range(stop-start+1):
                            yield start+index, f"{base+index:0{len(value)}X}".encode("ascii")

            sources = expand()
        for code, value in sources:
            if len(result) >= 65536:
                return result
            try:
                result[code] = bytes.fromhex(value.decode("ascii")).decode("utf-16-be")
            except (ValueError, UnicodeError):
                continue
    return result


def embedded_truetype(font, codes, stretch=1.):
    """Return an SFNT with a real Unicode map and a family private to this process."""
    try:
        descendant = font.DescendantFonts[0]
        stream = descendant.FontDescriptor.get("/FontFile2")
        if not isinstance(stream, pdf.Stream) or descendant.Subtype != pdf.Name.CIDFontType2:
            return None
        data = stream.read_bytes()
        if len(data) > 1 << 20 or not codes:
            return None
        tt = TTFont(BytesIO(data))
        if "glyf" not in tt or "OS/2" in tt and tt["OS/2"].fsType & 0x202:
            return None
        glyphs = tt.getGlyphOrder()
        mapping = descendant.get("/CIDToGIDMap")
        if isinstance(mapping, pdf.Stream):
            mapping = mapping.read_bytes()
        else:
            mapping = None
        cmap = {}
        for code, char in codes.items():
            if len(char) != 1 or not (0 < ord(char) < 0xd800 or 0xe000 <= ord(char) <= 0x10ffff):
                continue
            gid = code if mapping is None else int.from_bytes(mapping[2*code:2*code+2], "big")
            if 0 < gid < len(glyphs):
                cmap.setdefault(ord(char), glyphs[gid])
        if 32 not in cmap:
            # FontBuilder and the text writer require a space metric.
            return None
        pdf_name = str(font.BaseFont).lstrip("/")
        ps_name = re.sub(r"^[A-Z]{6}\+", "", pdf_name)
        real_family = tt["name"].getBestFamilyName() if "name" in tt else None
        if real_family and re.fullmatch(r"PB[0-9a-f]{16}", real_family):
            real_family = None
        family = real_family or re.sub(r"(?<=[a-z])(?=[A-Z])", " ", ps_name)
        digest = sha256(data)
        digest.update(f"{stretch:.5f}".encode("ascii"))
        for code, char in sorted(cmap.items()):
            digest.update(f"{code}:{char};".encode("utf-8"))
        alias = "PB" + digest.hexdigest()[:16]
        bold = bool(tt["head"].macStyle & 1) or "bold" in ps_name.casefold() or "-bd" in ps_name.casefold()
        italic = bool(tt["head"].macStyle & 2) or "italic" in ps_name.casefold()
        variant = ("Bold " if bold else "") + ("Italic" if italic else "") or "Regular"
        builder = FontBuilder(font=tt)
        builder.setupCharacterMap(cmap)
        builder.setupNameTable(dict(familyName=alias, styleName=variant, uniqueFontIdentifier=alias,
                                    fullName=alias, psName=ps_name, version="Version 1.0"))
        if "OS/2" not in tt:
            hhea = tt["hhea"]
            builder.setupOS2(sTypoAscender=hhea.ascent, sTypoDescender=hhea.descent,
                             usWinAscent=max(0, hhea.ascent), usWinDescent=max(0, -hhea.descent),
                             usWeightClass=700 if bold else 400, fsType=0,
                             fsSelection=(32 if bold else 0) | int(italic) or 64)
        if "post" not in tt:
            builder.setupPost(keepGlyphNames=False, italicAngle=-12 if italic else 0)
        if stretch != 1.:
            glyf = tt["glyf"]
            for name in glyphs:
                glyph = glyf[name]
                if glyph.isComposite():
                    for part in glyph.components:
                        if hasattr(part, "x"):
                            part.x = round(part.x*stretch)
                elif glyph.numberOfContours > 0:
                    glyph.coordinates.scale((stretch, 1.))
                    glyph.coordinates.toInt()
                width, bearing = tt["hmtx"][name]
                tt["hmtx"][name] = max(0, min(65535, round(width*stretch))), round(bearing*stretch)
        output = BytesIO()
        tt.save(output)
        return alias, family, output.getvalue()
    except (AttributeError, IndexError, KeyError, OSError, TypeError, ValueError, TTLibError, pdf.PdfError):
        return None
