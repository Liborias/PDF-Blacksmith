"""Embed TrueType fonts with an explicit mapping from PDF CIDs to Unicode."""

from dataclasses import dataclass
from io import BytesIO
from pathlib import Path
import re
import sys

from fontTools.ttLib import TTFont, TTLibError
import pikepdf as pdf

from text_edit import FAMILIES, TextError, instruction


@dataclass
class BlockFace:
    data: bytes
    font: TTFont
    cmap: dict
    widths: dict
    ascent: float
    fake_bold: bool
    fake_italic: bool


def block_face(family, style, folder, font_data=None):
    bundled = ("Liberation Sans", "Liberation Serif", "Liberation Mono")
    bold, italic = bool(style.get("bold")), bool(style.get("italic"))
    try:
        if font_data is not None:
            data = font_data
        elif family in bundled:
            variant = ("Bold" if bold else "") + ("Italic" if italic else "") or "Regular"
            data = (Path(folder) / f"{FAMILIES[bundled.index(family)]}-{variant}.ttf").read_bytes()
        elif sys.platform == "win32":
            from system_fonts import load_system_font
            data = load_system_font(family, bold, italic)
        else:
            raise TextError("Toto systémové písmo je dostupné ve Windows.")
        font = TTFont(BytesIO(data))
        if "glyf" not in font:
            raise TextError("Pro úpravu zvolte písmo s křivkami TrueType.")
        os2 = font.get("OS/2")
        permissions = os2.fsType if os2 is not None else 0
        if permissions & 0x202 or permissions & 4 and not permissions & 8:
            raise TextError("Toto písmo má omezené vložení do PDF. Vyberte jiné písmo.")
        units = font["head"].unitsPerEm
        cmap = {chr(code): glyph for code, glyph in (font.getBestCmap() or {}).items() if glyph != ".notdef"}
        metrics = font["hmtx"].metrics
        widths = {char: round(metrics[name][0]*1000/units) for char, name in cmap.items()}
        ascent = max(.7,min(1.2,font["hhea"].ascent/units))
        flags = os2.fsSelection if os2 is not None else (font["head"].macStyle & 1)*32 | bool(font["head"].macStyle & 2)
        return BlockFace(data,font,cmap,widths,ascent,bold and not flags&32,italic and not flags&0x201)
    except (OSError, TTLibError, KeyError, ValueError) as exc:
        if isinstance(exc, ValueError) and not isinstance(exc, TTLibError):
            raise TextError(str(exc)) from exc
        raise TextError("Zvolené písmo se nepodařilo načíst.") from exc


def fonts_for_block(style, content, folder, font_data=None):
    bundled = ("Liberation Sans", "Liberation Serif", "Liberation Mono")
    family = style.get("font", "")
    if not isinstance(family, str) or len(family)>63 or "\0" in family:
        raise TextError("Vyberte dostupné písmo.")
    if not family:
        family = bundled[style.get("family", 0)]
    faces = [block_face(family,style,folder,font_data)]
    needed={char for char in content if not char.isspace()}
    needed.add(" ")
    assigned={char:0 for char in needed if char in faces[0].cmap}
    options = (*bundled, "Segoe UI Symbol", "Arial") if sys.platform == "win32" else bundled
    for fallback in options:
        if len(assigned)==len(needed):
            break
        if family==fallback:
            continue
        try:
            candidate=block_face(fallback,style,folder)
        except TextError:
            if fallback in bundled:
                raise
            continue
        found=(needed-assigned.keys()) & candidate.cmap.keys()
        if found:
            assigned.update((char,len(faces)) for char in found)
            faces.append(candidate)
    if missing:=needed-assigned.keys():
        first=min(missing)
        raise TextError(f"Písmo ani dostupná náhradní písma neobsahují znak {first} (U+{ord(first):04X}).")
    return faces,assigned,{char:faces[index].widths[char]/1000 for char,index in assigned.items()}


def unicode_cmap(characters):
    lines = [
        "/CIDInit /ProcSet findresource begin", "12 dict begin", "begincmap",
        "/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def",
        "/CMapName /BlacksmithUnicode def", "/CMapType 2 def",
        "1 begincodespacerange", "<0001> <FFFF>", "endcodespacerange",
    ]
    for start in range(0,len(characters),100):
        group=characters[start:start+100]
        lines.append(f"{len(group)} beginbfchar")
        lines.extend(f"<{cid:04X}> <{char.encode('utf-16-be').hex().upper()}>"
                     for cid,char in enumerate(group,start+1))
        lines.append("endbfchar")
    lines.extend(("endcmap","CMapName currentdict /CMap defineresource pop","end","end"))
    return ("\n".join(lines)+"\n").encode("ascii")


def embed_face(document,page,face,characters):
    font=face.font
    head=font["head"]
    scale=1000/head.unitsPerEm
    name=re.sub(r"[^A-Za-z0-9_.+-]","",font["name"].getDebugName(6) or "") or "BlacksmithTrueType"
    base=pdf.Name("/"+name)
    os2=font.get("OS/2")
    hhea=font["hhea"]
    embed=document.make_stream(face.data,Length1=len(face.data))
    italic=float(font["post"].italicAngle)
    descriptor=document.make_indirect(pdf.Dictionary(
        Type=pdf.Name.FontDescriptor,FontName=base,Flags=32 | bool(italic)*64,
        FontBBox=pdf.Array(round(i*scale) for i in (head.xMin,head.yMin,head.xMax,head.yMax)),
        Ascent=round(hhea.ascent*scale),Descent=round(hhea.descent*scale),
        CapHeight=round(getattr(os2,"sCapHeight",hhea.ascent)*scale),
        ItalicAngle=italic,StemV=85,FontFile2=embed))
    glyph_ids=[font.getGlyphID(face.cmap[char]) for char in characters]
    mapping=b"\0\0"+b"".join(glyph.to_bytes(2,"big") for glyph in glyph_ids)
    cid=document.make_indirect(pdf.Dictionary(
        Type=pdf.Name.Font,Subtype=pdf.Name.CIDFontType2,BaseFont=base,
        CIDSystemInfo=pdf.Dictionary(Registry="Adobe",Ordering="Identity",Supplement=0),
        FontDescriptor=descriptor,CIDToGIDMap=document.make_stream(mapping),
        W=pdf.Array([1,pdf.Array(face.widths[char] for char in characters)])))
    composite=document.make_indirect(pdf.Dictionary(
        Type=pdf.Name.Font,Subtype=pdf.Name.Type0,BaseFont=base,Encoding=pdf.Name("/Identity-H"),
        DescendantFonts=pdf.Array([cid]),ToUnicode=document.make_stream(unicode_cmap(characters))))
    key=page.add_resource(composite,pdf.Name.Font,prefix="Text")
    return key,{char:i.to_bytes(2,"big") for i,char in enumerate(characters,1)}


def encoded_fragment(value,cids,extra,size):
    if extra and " " in value:
        data,items=bytearray(),[]
        for char in value:
            data.extend(cids[char])
            if char==" ":
                items.extend((pdf.String(bytes(data)),-1000*extra/size))
                data.clear()
        if data:
            items.append(pdf.String(bytes(data)))
        return instruction("TJ",pdf.Array(items))
    return instruction("Tj",pdf.String(b"".join(cids[char] for char in value)))
