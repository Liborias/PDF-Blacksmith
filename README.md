# PDF Blaksmith

Česká utilita pro přípravu PDF k tisku a práci s digitálními podpisy na Windows x64. Převádí dokumenty na desku K, text na křivky a procesní barvy do profilů CMYK, včetně FOGRA39, 47, 51, 52 a 54.

## Použití

Stáhněte `PDFBlaksmith.exe` z vydání a spusťte jej. Ghostscript je vestavěný. [Vydání](https://github.com/Liborias/PDF-Blacksmith/releases) · [Návod](NAVOD.md)

## Vývoj

Zdroj rozhraní a backendu je přímo v repozitáři. Návrh kovadliny a ikona jsou v `assets/`. Na Linuxu spusťte `./build.sh` pro sestavení aplikace pro Windows. [Požadavky na sestavení](SESTAVENI.md)

Backend lze otestovat příkazem `python3 tests/smoke.py`. Potřebuje Ghostscript, `pdftotext` a Python s balíčky z `requirements-win.txt`. Složka `tests/fixtures` obsahuje syntetické vzory barev a registračních značek.
