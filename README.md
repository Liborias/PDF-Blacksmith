# PDF Blacksmith

Česká utilita pro přípravu PDF k tisku a práci s digitálními podpisy na Windows x64. Převádí dokumenty na desku K, text na křivky a procesní barvy do profilů CMYK, včetně FOGRA39, 47, 51, 52 a 54.

PDF se otevírá přímo v hlavním okně. Podporuje souvislé a stránkové čtení, přiblížení ke kurzoru pomocí Ctrl a kolečka a celoobrazovkový režim. Horní záložky přepínají nástroje v postranním panelu.

Editor umožňuje otáčet a přesouvat stránky a uložit upravenou kopii. Tiskový náhled zobrazuje výstupní záměr, procesní i přímé desky a jejich pokrytí. Parametry převodu jsou v panelu **Export**.

## Použití

Stáhněte `PDFBlaksmith.exe` z vydání a spusťte jej. Součástí aplikace je PDFium pro prohlížení a Ghostscript pro tiskové operace. [Vydání](https://github.com/Liborias/PDF-Blacksmith/releases) · [Návod](NAVOD.md)

## Vývoj

Zdroj rozhraní a backendu je přímo v repozitáři. Návrh kovadliny a ikona jsou v `assets/`. Na Linuxu spusťte `./build.sh` pro sestavení aplikace pro Windows. [Požadavky na sestavení](SESTAVENI.md) · [Plán vývoje](PLAN.md)

Backend lze otestovat příkazy `python3 tests/smoke.py` a `python3 tests/preview.py`. Potřebuje Ghostscript, `pdftotext` a Python s balíčky z `requirements-win.txt`. Příkaz `go test ./...` na Linuxu ověří rozložení stránek, limit mezipaměti a zápis rozsahů. Složka `tests/fixtures` obsahuje syntetické vzory barev a registračních značek.
