# PDF Blacksmith – iniciační pokyn pro Codex, 25. listopadu 2026

Repozitář: https://github.com/Liborias/PDF-Blacksmith

Pokračujeme na existující desktopové aplikaci PDF Blacksmith (Go/Win32, PDFium a navazující Python/PDF operace). **První úkol je analýza, nikoliv implementace.** Není dovoleno nahrazovat aplikaci zjednodušenou kopií nebo předpokládat, že HTML návrh je produkční funkční editor.

## Doplnění schváleného HTML z rána 25. 11. 2026

Před analýzou přečti také [UI_ZOBRAZENI_OCR_PANELY_2026-11-25.md](UI_ZOBRAZENI_OCR_PANELY_2026-11-25.md). Do aktuálního HTML `cely_vizual_v3.25.html` byly zapracovány následující návrhové úpravy:
- Dekorativní vínový obrys stránek v hlavní ploše odstraněn ve všech režimech, včetně přizpůsobení na šířku. Nevypínat rámečky formulářů ani reálné anotace PDF.
- Režim „po stránkách“ reaguje na kolečko: uvnitř vysoké stránky scroll, na horním/dolním konci přechod mezi stránkami; nepřepisovat souvislý režim ani Ctrl+wheel.
- Vysvětlivka OCR sdíleného rozpoznávání je skryta pod kruhovým otazníkem v příslušné sekci. Další nepodstatné vysvětlivky mají mít stejnou politiku. V horní druhé liště zůstává Help tlačítko, ale jeho znak ? je bez vlastní kruhové dekorace a opticky vysoký jako ostatní ikony.
- Panely se zbytečnými prázdnými svislými sloty jsou kompaktnější, logické skupiny mají oddělovač a skládací ouško; dostupné volby se při sbalení neztrácejí.
- Stejně jako u předchozích iterací jde o **HTML návrh**, nikoliv hotovou produkční Win32 implementaci. Ověř regresně výběrová gesta, dokumentové barvy, formuláře a nápovědu. Aktuální HTML v3.25 není nové číslo verze, ale aktualizace jejího obsahu.

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

## ÚPLNĚ POSLEDNÍ BOD – odložený experimentální adaptivní režim stránek

Toto je poslední bod zadání, nikoliv současná implementační priorita. **Pouze analyzuj proveditelnost**, návrh samostatně schválíme později.

- Nový režim pro více stran na pracovní ploše, podobný `flex-wrap`: při oddálení a větším dostupném prostoru se stránky skládají vedle sebe i pod sebou; při vhodném měřítku mohou být zobrazené všechny. Nepřekládat obyčejný scroll na bezpodmínečné oddálení; rozlišit scroll, přiblížení a výběr.
- Klik zvolí jednu stránku; `Enter` vybranou okamžitě přizpůsobí výšce pracovní plochy.
- `Ctrl + kolečko` v hlavní pracovní ploše a panelu miniatur přibližuje/oddaluje příslušnou oblast, nikoliv celou HTML stránku. Ověř, zda prohlížeč umožní zastavit nativní Ctrl+wheel pomocí aktivního `wheel` listeneru, `preventDefault` a ohraničení oblasti; na Win32 navrhni alternativu. Otestovat možné konflikty se systémovými zkratkami a přístupností.
- `Ctrl + levý klik` přidává a odebírá stránky z výběru. Hromadné přesouvání a operace se řídí stejnými pravidly, modelem a ochranou Undo/Redo jako v panelu miniatur; respektovat aktuální nástroj a nepřebíjet Ctrl+klik používaný pro objekty a vodítka.
- Zásah do vykreslování, přiřazení výběru, změn měřítka a ukládání řešit jako návrh architektury a samostatně testovatelnou etapu **teprve po souhlasu autora**. Nepřepisovat existující „po stránkách“ ani „souvislé“.

**Upozornění: první úkol se nemění. Zatím nic neupravuj, neimplementuj a necommituj; nejprve mi předlož analýzu projektu.**
