"""Build the Windows runtime archive embedded by Go."""

from pathlib import Path
import shutil
import sys
from zipfile import ZIP_DEFLATED, ZipFile

source = Path(__file__).resolve().parent
runtime = Path(sys.argv[1]).resolve()
for name in ("engine.py", "bookmarks.py", "color_conversion.py", "export_settings.py", "export_presets.json", "pdfx.py", "proof.py", "reader_render.py", "reader_drag.py", "reader_search.py", "ocr.py", "text_edit.py", "text_move.py", "embedded_fonts.py", "text_block.py", "block_fonts.py", "system_fonts.py",
             "print_profiles.py", "page_transfer.py", "object_edit.py", "profiles.json", "OSS.txt"):
    shutil.copyfile(source / name, runtime / name)
(runtime / "python314._pth").write_text("python314.zip\n.\npackages\n", encoding="ascii")
with ZipFile(source / "runtime.zip", "w", ZIP_DEFLATED, compresslevel=6) as archive:
    for file in sorted(runtime.rglob("*")):
        relative = file.relative_to(runtime)
        if (file.is_file() and "__pycache__" not in relative.parts and
                relative.parts[0] != "bin" and relative.parts[:2] != ("packages", "bin") and
                not (len(relative.parts) == 1 and relative.name == "pythonw.exe")):
            archive.write(file, relative.as_posix())
