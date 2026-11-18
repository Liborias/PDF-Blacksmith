# Sestavení PDF Blaksmith

Na Linuxu nainstalujte Go 1.26 nebo novější, Python 3 s pipem, curl, unzip, 7-Zip s podporou NSIS a sha256sum. Spuštěním `./build.sh` vznikne `PDFBlaksmith.exe` pro Windows x64. Výstupní cestu můžete předat jako první argument.

Skript ověřuje distribuci Ghostscriptu 10.08.0 a Python 3.14.10 kontrolním součtem. Do EXE balí konzolový Ghostscript s potřebnými prostředky, Python a závislosti z `requirements-win.txt`. Podmínky třetích stran a odkaz na odpovídající zdroj Ghostscriptu jsou v `OSS.txt`.

`main_windows.go` implementuje obsluhu okna a `theme_windows.go` vykreslení. `runtime_windows.go` rozbaluje balíček a komunikuje s backendem přes JSON. `engine.py` provádí operace nad PDF a `print_profiles.py` spravuje profily ICC. Přesné verze profilů se kontrolují podle `profiles.json`.

Test backendu spustíte příkazem `python3 tests/smoke.py`. Cestu ke Ghostscriptu lze předat jako první argument. Nastavením `PDFBLAKSMITH_TEST_PROFILES=1` lze otestovat také export do profilů FOGRA39 a FOGRA51.

`python3 tests/preview.py` ověří oddělené desky, tiskové rámečky, nastavení exportu a zachování odkazů po přesunu stránek. `proof.py` připravuje obraz tiskového náhledu a statistiky z separací Ghostscriptu. `viewer_windows.go` a `tools_windows.go` implementují náhled, editor a nastavení exportu.
