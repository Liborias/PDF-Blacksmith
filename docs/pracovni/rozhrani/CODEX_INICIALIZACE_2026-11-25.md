# PDF Blacksmith – iniciační pokyn pro Codex, 25. listopadu 2026

Repozitář: https://github.com/Liborias/PDF-Blacksmith

Pokračujeme na existující desktopové aplikaci PDF Blacksmith (Go/Win32, PDFium a navazující Python/PDF operace). **První úkol je analýza, nikoliv implementace.** Není dovoleno nahrazovat aplikaci zjednodušenou kopií nebo předpokládat, že HTML návrh je produkční funkční editor.

## Před zahájením
Ověř aktuální větev, poslední commity, změny ostatních autorů a čistotu pracovního stromu. Přečti v uvedeném pořadí:
1. Kořenové `PLAN.md`, `NAVOD.md` a odpovídající produkční zdroje (zejména `workspace_windows.go`, `viewer_windows.go`, ukládání a export).
2. `docs/pracovni/README.md` a `docs/pracovni/POSTUP_SPOLUPRACE_S_CODEXEM.md`.
3. `docs/pracovni/specifikace_Codex_v5.0.md` a `docs/pracovni/plan_rozsireni_v2.0.md`.
4. `docs/pracovni/formulare/F1_rozhodnuti.md`, `formulare/F1_specifikace_v1.9.md`, `formulare/F1_plan_implementace.md`.
5. **Aktuální HTML** `docs/pracovni/rozhrani/cely_vizual_v3.25.html` a `rozhrani/UI_FINALNI_KONTRAST_AKTIVNI_IKONY_v3.25.md`; pro vývojové upřesnění zkontroluj také starší dodatky v3.1–v3.24 podle kontextu.
6. `docs/pracovni/mapa_barev_v1.4.md`, `mapa_barev_v1.4.json`, **`mapa_ikon_v4.7.md`**, `mapa_klavesovych_zkratek_v1.3.md` a `i18n/ui_catalog_v1.1.json`.

## Poslední schválený stav k 24. listopadu
- **v3.25 je nejnovější HTML a byla opravena přímo v tomto souboru.** Nezačínej ze starší kopie. Git historie uchovává předchozí stav.
- Hlavní textová menu druhé lišty (Soubor, Zobrazení, Nástroje, Úpravy, Formuláře, Pomoc) nejsou orámovaná tlačítka. Aktivní ikonová tlačítka v téže liště mají zřetelně zvýrazněné pozadí.
- Levý panel Nástroje a rozbalovací navigační nabídky mají **položky bez obdélníkových rámečků**: při hover nebo aktivním stavu pouze barevné podbarvení a tenkou akcentní linku vlevo, vždy podle zvoleného motivu.
- V Nastavení → Rozhraní nyní používáme samostatnou SVG ikonu `i-settings_interface`, zdroj `rozhrani/ikony/settings_interface.svg`; vazba je dokumentovaná v `mapa_ikon_v4.7.md`. Původní chybná vazba odkazovala na neexistující `i-panel_left`. Nevytvářej duplicitní symbol.
- Šest palet ve světlém i tmavém režimu se řídí mapou v1.4. Barevné vzorky, PDF obsah, uživatelské barvy vodítek a štítků se nesmějí přebarvovat pomocí UI motivu.
- Půlený přepínač světlý/tmavý sdílí stav s Nastavení → Rozhraní. Obsah HTML je schválený **prototyp**, nikoliv důkaz implementace Win32.
- Následující pravidla a předchozí schválená gesta se zachovávají: režimy SEL/TXT/OBJ, Ctrl+klik/Ctrl+tažení, výběr textu/objektů, vodítka a přichytávání, klávesové zkratky, panely, Undo/Redo, související lokalizace.

## Ochrana projektu a GitHubu
Před úpravami vždy prozkoumej existující implementaci, odliš návrh od funkčního kódu a vyhodnoť závislosti. Nic nemaž ani nezjednodušuj; neodstraňuj funkce Text, Objekty, OCR, separace, podpisy nebo formuláře. Nezasahuj do dokumentových barev, ICC/CMYK, podpisů, odkazů ani struktury PDF. F1 se implementuje pouze po schválených dílčích krocích. F2/F3, dávky, nové anotační mechanismy a další plánované funkce nejsou automaticky povoleny.

Při ukládání zachovej `/AcroForm`, `/Fields`, `/Widget`, `/V`, `/AP`, podpisy a odkazy. Nespouštěj neověřený PDF JavaScript; XFA identifikuj bezpečně. Neprováděj destruktivní reset, `push --force` ani změny cizí práce. Do veřejného repozitáře nepatří zákaznické soubory, osobní údaje, certifikáty ani tajné klíče. Po schválení pracuj po malých testovatelných etapách s diffem, regresními testy, aktualizací dokumentace a přehlednými commity.

## První výstup, který chci
1. Shrň skutečně implementované, částečné a pouze navržené funkce na základě kódu.
2. Porovnej produkční UI s posledním HTML v3.25 a napiš konkrétní rozdíly i závislosti. Zkontroluj také opravenou navigaci bez rámečků a ikonu Rozhraní.
3. Prověř stav formulářů F1 a navrhni **první malou bezpečnou etapu**, s přesnými soubory, riziky a regresními testy: diakritika, /AP, /V, opakovaná pole, restart, nevalidní vstupy a export bez poznámek.
4. Identifikuj rozpory a zastaralé odkazy v dokumentaci; starší titulky někde stále uvádějí v3.19 jako aktuální, ale rozhoduje v3.25. Opravy dokumentace zatím pouze navrhni.
5. Předlož seznam nejbližších kroků a vyčkej na schválení.

**Zatím nic neimplementuj, nepřepisuj soubory, nevytvářej commit ani push.**
