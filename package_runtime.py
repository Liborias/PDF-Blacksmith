"""Build the Windows runtime archive embedded by Go."""

from pathlib import Path
import shutil
import sys
from zipfile import ZIP_DEFLATED, ZipFile

source = Path(__file__).resolve().parent
runtime = Path(sys.argv[1]).resolve()
for name in ("engine.py", "export_settings.py", "proof.py", "reader_render.py",
             "print_profiles.py", "profiles.json", "OSS.txt"):
    shutil.copyfile(source / name, runtime / name)
(runtime / "python314._pth").write_text("python314.zip\n.\npackages\n", encoding="ascii")
with ZipFile(source / "runtime.zip", "w", ZIP_DEFLATED, compresslevel=6) as archive:
    for file in sorted(runtime.rglob("*")):
        relative = file.relative_to(runtime)
        if (file.is_file() and "__pycache__" not in relative.parts and
                relative.parts[0] != "bin" and relative.parts[:2] != ("packages", "bin") and
                not (len(relative.parts) == 1 and relative.name == "pythonw.exe")):
            archive.write(file, relative.as_posix())
