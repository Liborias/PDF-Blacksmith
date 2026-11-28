# Sestavení PDF Blacksmith

Na Linuxu nainstalujte Go 1.26 nebo novější, Python 3 s pipem, curl, unzip, 7-Zip s podporou NSIS, sha256sum a Inkscape (pro generování ikony aplikace). Spuštěním `./build.sh` vznikne `PDFBlaksmith.exe` pro Windows x64. Výstupní cestu můžete předat jako první argument.

Ikona EXE i logo v první liště používají společný prostředek Windows. `tools/generate_brand_icon.py` při sestavení vytvoří průhledné ICO (16–256 px) přímo z originálního SVG v HTML v3.26 v odstínu `#AD5B80` ze schválené vínové palety. Staré `assets/blaksmith.ico` a `assets/blacksmith.png` již nejsou použity tímto sestavením; ponechány jsou kvůli historii. Samostatnou ikonu lze vytvořit příkazem `python3 tools/generate_brand_icon.py /tmp/blacksmith.ico`.

Skript ověřuje distribuci Ghostscriptu 10.08.0 a Python 3.14.10 kontrolním součtem. Do EXE balí konzolový Ghostscript s potřebnými prostředky, Python a závislosti z `requirements-win.txt`. Podmínky třetích stran a odkaz na odpovídající zdroj Ghostscriptu jsou v `OSS.txt`.

`main_windows.go` implementuje obsluhu okna a `theme_windows.go` vykreslení. `runtime_windows.go` rozbaluje balíček a komunikuje s backendem přes JSON. `engine.py` provádí operace nad PDF a `print_profiles.py` spravuje profily ICC. Přesné verze profilů se kontrolují podle `profiles.json`.

Test backendu spustíte příkazem `python3 tests/smoke.py`. Cestu ke Ghostscriptu lze předat jako první argument. Nastavením `PDFBLAKSMITH_TEST_PROFILES=1` lze otestovat také export do profilů FOGRA39 a FOGRA51.

`python3 tests/check_ocr.py CESTA_K_GHOSTSCRIPTU CESTA_K_TESSDATA` ověří OCR na skenu a křivkách, včetně vyznačené oblasti, otočené strany, zachování vzhledu a opakovaného rozpoznání. Použijte Ghostscript s výstupním zařízením `pdfocr8` a jazykovými daty `eng.traineddata`.

Všechny tři předvolby PDF/X otestujete příkazem `python3 tests/check_pdfx.py CESTA_K_GS_10.08 CESTA_K_PROFILU_CMYK.icc`. Použijte tiskový profil CMYK ve verzi ICC v2, například zabudovaný testovací profil Ghostscriptu `iccprofiles/default_cmyk.icc`.

`python3 tests/preview.py` ověří čtecí a tiskové vykreslení, tiskové rámečky, nastavení exportu a zachování odkazů po přesunu stránek. `proof.py` připravuje obraz náhledu a statistiky ze separací Ghostscriptu. `viewer_windows.go` načítá stránky pracovního prostoru a zobrazuje je podle aktivního panelu. `workspace_windows.go` spravuje rozložení a nástroje. Test rozložení a zoomu lze spustit i na Linuxu příkazem `go test reader_layout.go reader_layout_test.go`.


`python3 tests/check_text.py CESTA_K_PISMUM_LIBERATION` ověří změnu sazby a český text, zachování okolní grafiky, záložek a změn pořadí stran.
