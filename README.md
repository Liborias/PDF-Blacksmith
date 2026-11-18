# PDF Blacksmith

PDF Tisk je česká utilita pro Windows 10 a 11 (x64). Pomocí Ghostscriptu převádí PDF do stupňů šedi a kontroluje, zda na tiskových separacích zůstává aktivní pouze deska K. Umí také převést text na křivky, uložit kopii PDF bez hesla a vytvářet a ověřovat digitální podpisy.

## Použití

Stáhněte `PDFTisk.exe` z vydání a spusťte jej. Pro převod do šedi a na křivky se Ghostscript instaluje samostatně. [Vydání](https://github.com/Liborias/PDF-Blacksmith/releases) · [Návod k použití](PDFTisk-navod.md)

## Sestavení a test

Na Linuxu spusťte `./build.sh`. Výsledkem je `PDFTisk.exe` pro Windows x64. [Požadavky a postup sestavení](SESTAVENI.md)

Test backendu vyžaduje Python 3 s balíčky z `requirements-win.txt`, Ghostscript a `pdftotext` z balíku Poppler. Spusťte `python3 tests/smoke.py`. Testovací PDF a podpisový certifikát se při běhu vytvoří v dočasné složce.

Složka `tests/fixtures` obsahuje syntetické tiskové vzory s RGB, CMYK, přímou barvou a registračními značkami i ukázky jejich převodu na desku K.
