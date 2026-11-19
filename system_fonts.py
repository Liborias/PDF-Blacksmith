"""Read the selected Windows TrueType face, including faces in collections."""

from ctypes import WinDLL, create_string_buffer, create_unicode_buffer, wintypes
from io import BytesIO

from fontTools.ttLib import TTCollection, TTFont


def load_system_font(family, bold, italic):
    gdi = WinDLL("gdi32")
    gdi.CreateCompatibleDC.argtypes = [wintypes.HDC]
    gdi.CreateCompatibleDC.restype = wintypes.HDC
    gdi.CreateFontW.argtypes = [wintypes.INT] * 5 + [wintypes.DWORD] * 8 + [wintypes.LPCWSTR]
    gdi.CreateFontW.restype = wintypes.HFONT
    gdi.SelectObject.argtypes = [wintypes.HDC, wintypes.HGDIOBJ]
    gdi.SelectObject.restype = wintypes.HGDIOBJ
    gdi.GetTextFaceW.argtypes = [wintypes.HDC, wintypes.INT, wintypes.LPWSTR]
    gdi.GetFontData.argtypes = [wintypes.HDC, wintypes.DWORD, wintypes.DWORD, wintypes.LPVOID, wintypes.DWORD]
    gdi.GetFontData.restype = wintypes.DWORD
    gdi.DeleteObject.argtypes = [wintypes.HGDIOBJ]
    gdi.DeleteDC.argtypes = [wintypes.HDC]

    dc = gdi.CreateCompatibleDC(None)
    if not dc:
        raise ValueError("Seznam písem systému Windows není dostupný.")
    face = previous = None
    try:
        face = gdi.CreateFontW(-16, 0, 0, 0, 700 if bold else 400, italic, 0, 0, 1, 4, 0, 0, 0, family)
        if not face:
            raise ValueError("Zvolené písmo se nepodařilo načíst.")
        previous = gdi.SelectObject(dc, face)
        actual_name = create_unicode_buffer(64)
        gdi.GetTextFaceW(dc, len(actual_name), actual_name)
        if not previous or not actual_name.value:
            raise ValueError("Windows nemají zvolené písmo k dispozici.")

        def read_table(tag):
            size = gdi.GetFontData(dc, tag, 0, None, 0)
            if not 0 < size <= 64 << 20:
                raise ValueError("Zvolené písmo nemá dostupná data TrueType.")
            result = create_string_buffer(size)
            if gdi.GetFontData(dc, tag, 0, result, size) != size:
                raise ValueError("Data zvoleného písma se nepodařilo přečíst.")
            return result.raw

        data = read_table(0)
        if data.startswith(b"ttcf"):
            names = read_table(int.from_bytes(b"name", "little"))
            with TTCollection(BytesIO(data)) as collection:
                for candidate in collection.fonts:
                    if candidate.reader["name"] == names and "glyf" in candidate:
                        buffer = BytesIO()
                        candidate.save(buffer)
                        return buffer.getvalue()
            raise ValueError("Zvolený řez se v kolekci písem nepodařilo najít.")
        with TTFont(BytesIO(data)) as font:
            if "glyf" not in font:
                raise ValueError("Pro úpravu zvolte písmo s křivkami TrueType.")
        return data
    finally:
        if previous:
            gdi.SelectObject(dc, previous)
        if face:
            gdi.DeleteObject(face)
        gdi.DeleteDC(dc)
