# PDF Blacksmith — AUDIT ROZPORŮ A JEJICH ROZHODNUTÍ (25. 11. 2026)

**Výsledek:** všechny níže uvedené konflikty mají **jednu normativní odpověď** v [CODEX_MASTER_v6.0.md](CODEX_MASTER_v6.0.md). Tento audit je vysvětlení migrace, ne samostatný zdroj alternativních pokynů. Zdrojový kontext: uživatelská rozhodnutí z předchozích rozhovorů „Upřesnění pokynů pro Codex“, „Pokračovat v ladění PDF“ a tohoto navazujícího chatu, plus pracovní soubory v GitHubu (zejména v5.0, v3.9–v3.26, F1 rozhodnutí a všechny aktuální mapy). Historické spisy zůstávají, ale **jejich hlavičky „aktuální“ už nejsou závazné**.

| # | Rozpor | Finální rozhodnutí / proč |
|---|---|---|
| 1 | „F1 je další priorita“ ve starším plánu vs uživatel 25. 11. žádá nejdřív celé rozhraní | **Změnit pořadí: UI shell → chování UI → navigace → výběr/pomůcky → F1**; schválený rozsah F1 nerušit, jen odložit. Rozpracovanou F1 práci nezahazovat. |
| 2 | „Nejnovější HTML“ v různých částech README/specifikace uvádí v3.9, v3.19, v3.25 | **Jedno HTML v3.26** je poslední výchozí referenční prototyp; ostatní historické. |
| 3 | Mapa ikon v README uvedena v4.7, ale historický souběžný vývoj pokračoval přes v4.8–v5.1 | **Sloučit všechny validní dodatky z v5.1 i novější konkrétní i-settings_interface z později upravené v4.7**; zdrojem pravdy je nová mapa FINAL v6.0 odvozená z HTML v3.26 a zdrojů. |
| 4 | Horní split tlačítko vs oddělený pointer/text vs dvoustavový cyklus | **Jeden dvoustavový pointer ↔ caret** ukazuje příští akci; obdélník/laso vlastní ikony, odstavcový textový rámec v Text → Výběr. |
| 5 | Obdélník/laso v Text vs univerzální výběr ve všech nástrojích | **Jediný Selection Dispatcher** s kontextovým hit-testem text/raster/vektor/OCR, žádné kopie handlerů. |
| 6 | Konverze „text na křivky“ v Typografii v3.5/3.6 vs Výběr v3.7+ | **Pouze Text → Výběr**, jako akce nad aktuálním výběrem; nepřesouvat zpět. |
| 7 | Pevně vyhrazený 59/64px slot nápovědy v některých dřívějších UI verzích vs uživatel 25. 11. žádá odstranit mezery | **Bez prázdného slotu**; krátký popis při hover/focus má minimální dynamickou výšku, rozsáhlý text schovat pod ? u sekce. |
| 8 | Hover menu vodítka ve v3.15 vs v3.16+ pravé kliknutí | **Pravé kliknutí** otevírá jen barva/zámek/koš, hover žádné tlačítko menu; levý klik vybírá. |
| 9 | Ve v3.14 opačně určena orientace vodítka z pravítka | Platí **horní pravítko → vodorovná Y; levé → svislá X** podle oprav v3.15. |
| 10 | Ctrl+klik na vodítku vybere vs Ctrl+drag jej duplicitně vytvoří | Vyhodnotit **krátký klik až při pointerup**, nadprahový drag >~7px duplikuje, bez Ctrl posouvá. Jediný výběrový model, žádné dvojí provedení. |
| 11 | Skupinu vodítek smazat vs jen oddělit | **Smazat = skupina i členové**, **Oddělit = pouze rozvázat členství**; respektovat zámky. Sbalení nemění viditelnost linky. |
| 12 | Přesun barevných hodnot do UI akcentu a šest profilů vs staré jen čtyři akcenty/vínová všude | **Šest nezávislých plných palet × light/dark**, wine jen ve výchozí wine, custom potvrzení oddělené. PDF obsah a skutečné swatche se nikdy nebarví motivem. |
| 13 | Ozubené kolečko mělo být 21px (v3.18) vs následně opticky zmenšeno na 18px (v3.22) | **18px vizuální velikost** v horní liště; původní SVG souřadnice/ID zachovat. |
| 14 | Horní Nápověda měla kulaté tlačítko vs 25. 11. žádost o prostý otazník | **V horní liště tlačítko s prostým ? bez vnitřního kruhu**, v panelových sekcích kruhový otazník bez obdélníkového rámu. |
| 15 | Hlavní menu a navigační položky dostaly v light mode obdélníkové rámování | **Horní textové menu bez rámečku**, levá/vnořená navigace bez rámečku s levým akcentním pruhem; pouze aktivní **ikonová** tlačítka druhé lišty mají výrazný zvýrazněný obrys. |
| 16 | Editor měl vínový vnější obrys aktuálního PDF a některé zobrazení jej samo ztratilo | **Bez dekorativního vnějšího obrysu ve všech režimech**, nezaměnit s obrysy skutečných formulářových polí a objektů. |
| 17 | Zobrazení „po stránkách“ nereagovalo kolečkem vs dosavadní scroll | **Rolovat list a na kraji přejít dál/zpět**; souvislé posouvání neměnit. |
| 18 | Původně jen dvě zobrazení vs adaptivní uspořádání stran | **Třetí samostatný režim „Přehled stran“**, zachovat dvoustavové tlačítko Po stránkách/Souvisle; při oddálení flex-wrap. Produkční přesun/hromadné změny pouze po schválení a přes společné page-ops. |
| 19 | Prohlížeč při Ctrl+wheel zoomuje celé HTML vs požadované lokální plochy | **Ve dvou zónách lokální zoom**, editor ≠ miniatury; mimo zóny standardní zoom prohlížeče. Ověřit Chrome/Edge, neprezentovat jako garantované pro každý OS. |
| 20 | Automatické otevření všech pomocných panelů (odložený návrh 22. 11.) vs Search dočasně otevírá Výsledky a obnovuje stav (v3.11) | Implementována/normována **specifická dočasná logika Search**, vodítka dočasně ve Vlastnostech; ostatní obecná automatika se **nepovoluje plošně**. |
| 21 | „Otevřít“, „Importovat“, „Sloučit a otevřít“, „Vložit strany“, „Vložit objekt“ užívány jako synonyma | **Pět oddělených operací** s přesně daným zdrojem a cílem, uvedeno v master §6. |
| 22 | Formuláře „Návrh polí F2“ bylo vypnuté, v3.26 je klikací | **F2 pouze interaktivní HTML demonstrace**, nikoliv schválená Go funkčnost. F1 schválené po etapách, ale až po UI. |
| 23 | Zmínka „xmlx“ při importech vs tabulkové formáty | Pravděpodobně **XLSX**; samostatné **XML** řešit až se schématem a bezpečností. Google Tabulky přes read-only API, nikoliv anonymní obcházení oprávnění. |
| 24 | Mapa zkratek označuje existující T/V/M/L a zároveň je nabízí pro nové významy | **Stávající Win32 dispatch zachovat**, nové návrhové přemapování pouze po auditu konfliktů, až po schválení. |
| 25 | Přídavné mapy/slovníky stále tvrdí „v3.19 aktuální“ | Jednotné finální mapy a katalog, historické titulky nejsou aktuální instrukce. |
| 26 | „Nástroje“ vs „Formuláře“ řeší současně výběr/tažení stránek | **Gesto spravuje aktivní kontext/fokus**, stránky pouze v přehledu/miniaturách, objektový hit-test před stránkou, formulářový input před Enter. |
| 27 | Nápovědné poznámky byly některými CSS přesunuty do černých boxů ve světlém motivu | **Paletové info bannery a text**, žádná tvrdá černá/burgundy v nesouvisejících paletách. |
| 28 | Knihovní SVG atlas a mapa v5.1 vs nová ikonka Nastavení → Rozhraní | Nová **i-settings_interface** v inline HTML a samostatný settings_interface.svg; chybějící **i-panel_left se nepoužívá jako ikona Rozhraní**. |

| 29 | V HTML v3.26 jsou některé UI ikonové reference bez definice SVG symbolu | **Zachovat existující ikonovou sadu, ověřit chybějící `i-panel_left`, `i-arrow_line`, `i-transform_object`, `i-plate_all`, `i-plate_single` při prvním UI auditu; neprohlašovat prototyp za kompletně ověřený.** |

## Stav ověření a provenance
- Prošel jsem **dostupné odkazy předchozích rozhovorů a pracovní instrukce skutečně uložené v repozitáři**. Starší soubory často uvádějí status „schválené UI“ i u operací, které byly v HTML jen simulované; v master je to výslovně rozlišeno.
- Základní relevantní originály: hlavní specifikace v5.0, konsolidace UI v3.9, jednotlivé UI změny v3.11–v3.26, pravidla souborových operací z 22. 11., formulářová F-Q1–F-Q9, mapa ikon v5.1 a pozdější v4.7 (interface icon), mapa barev v1.4, mapa kláves v1.4, katalog i18n v1.1.
- Toto je **dokumentační konsolidace**, nikoli kompletní runtime audit produkční aplikace, ani důkaz, že v3.26 nemá chybné handlery či vizuální nedostatky. Fyzický Windows build/test a Chrome/Edge integrační test zůstává v Etapě 0/1.

**Pokud se Codex setká s historickým dokumentem označeným „aktuální“, není to další závazný plán. Má použít master v6.0, finální mapy a implementační plán v6.0.**