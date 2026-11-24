# PDF Blacksmith – UI v3.17: překlady, výběr a skupiny vodítek, cíle přichytávání

**Datum:** 24. listopadu 2026 · **Stav:** schválený návrh HTML a implementační zadání pro Codex, **nikoliv zásah do produkční Go/Win32/PDFium aplikace**.  
**Aktuální HTML:** [cely_vizual_v3.17.html](cely_vizual_v3.17.html)  
**Navazuje na:** [UI v3.16](UI_JAZYK_VODITKA_v3.16.md), [vodítka v3.15](UI_VODITKA_OPRAVY_v3.15.md), [specifikaci v5.0](../specifikace_Codex_v5.0.md), [mapu kláves v1.2](../mapa_klavesovych_zkratek_v1.2.md), [mapu ikon v5.0](../mapa_ikon_v5.0.md).

## 1. Dokončení anglického rozhraní

- Rozšířit překladový katalog přes všechny hlavní nástroje, nástrojové panely, rozšířená nastavení, hover nápovědy, menu, OCR a rekonstrukční dialog, nastavení jazyka, stavové zprávy a upozornění.
- **Nikdy nepřekládat** text otevřeného PDF, názvy existujících souborů, výběry uživatele, dokumentové záložky ani ukázkové hodnoty formulářů. Čeština v ukázkovém PDF a názvech souborů zůstává záměrně.
- Základní audit v Chromiu pokryl všechny hlavní panely a Nastavení, včetně dialogu OCR; ve zkoušených zobrazeních zůstaly české pouze údaje patřící k dokumentu. Nejedná se o úplný důkaz pokrytí všech možných stavů aplikace.
- V produkční implementaci používat **stabilní lokalizační klíče**, pluralizaci, fallback a automatické kontroly chybějících překladů. HTML dynamické nahrazování textů je pouze dočasný prototyp; překlady nesmějí měnit \`id\`, \`data-*\`, hodnoty operací ani OCR jazyk. Angličtina zůstává oddělená od rozpoznávaného jazyka PDF.

## 2. Výběr a vizuální vzhled jednotlivých vodítek

- Pevná vodítka mají **jemnější viditelnou linku cca 1 CSS px** (nepřisuzovat automaticky stejnou šířku fyzickému PDF); průhledná širší zásahová oblast pro uchopení a přístupnost zůstává.
- **Kliknutí na řádek seznamu** vybere právě jedno vodítko a odznačí předchozí řádek i jeho linku. **Kliknutí na jiný řádek** přesune jediný výběr. **Kliknutí na vodítko na stránce** se odrazí ve stejném seznamu.
- **Checkbox u řádku** přidává/ubírá vodítka do vícenásobného výběru, nemaže ostatní checkboxy. Skupinový checkbox umožňuje označit všechna vodítka ze skupiny. Jediný sdílený selection model, nikoli paralelní výběr seznamu a plátna.
- Kliknutí do volného místa plátna nebo na nepříslušný prvek výběr vodítka zruší; je-li panel Pomůcky rozvržení pouze dočasně otevřen, vrátí se nastavení předchozího nástroje, v souladu s v3.16. U rozpracovaného tažení, kontextové nabídky a fokusu v číselném vstupu nesmí docházet k náhodnému rušení.
- **Oko u každého vodítka** přepíná viditelnost. Skryté vodítko se nekreslí na plátně, neslouží jako cíl přichytávání a není interaktivně vybratelné ze stránky; v seznamu zůstává přístupné pro znovuzobrazení.
- Individuální barva, zamknutí, přesná souřadnice a odstranění stále platí. Zobrazení/skrytí nemění PDF, ani jeho dirty/undo stav.

## 3. Skupiny vodítek

- **Seskupit označená** vytvoří pojmenovatelnou skupinu v rámci jedné stránky, pouze ze skutečně označených vodítek. Výsledek není nový PDF objekt. **Oddělit ze skupiny** vrací členy mezi samostatná vodítka.
- Skupina má identifikátor, seznam členů a sdílená ovládání: **oko (skrýt/zobrazit), barva, zámek a koš**. Výchozí změna vlastnosti skupiny se promítne do všech jejích aktuálních členů; individuální změna mimo skupinovou operaci ovlivní pouze zvolené vodítko, pokud UI výslovně nenabízí skupinovou akci.
- **Mazání celé skupiny v aktuálním HTML znamená smazání skupiny i jejích vodítek**, nikoliv jen rozpuštění skupiny. Rozpuštění skupiny je odlišný, nedestruktivní příkaz. Uzamčené jednotlivé vodítko nebo globální zámek zabrání destruktivnímu smazání skupiny.
- Všechny změny se týkají pouze správného dokumentu/stránky, podle existujícího \`guideGroupsByPage\`. Při přepnutí stránky se výběr očistí; skupiny na jiných stránkách zůstanou. V budoucí produkční vrstvě budou skupiny ukládány do pracovních metadat mimo výrobní PDF, pokud uživatel výslovně nezvolí jiný postup.
- Při zámku/skrytí skupiny musí být stav i možnost manipulace členů konzistentní. Codex má před skutečnou implementací rozhodnout dědění stavu, priority individuální/skupinové vlastnosti a případné smíšené hodnoty; nevytvářet dva nezávislé výběrové engine.

## 4. Jednotný SnapEngine pro všechny editovatelné prvky

- Univerzální vyhodnocování přichytávání musí sloužit **vodítkům, textovým kurzorům a rámcům, vektorům, rastrům, skupinám a formulářovým polím**. Pomocná mřížka a vodítka nejsou samostatný duplicovaný snap engine.
- Schválené kandidáty rozšiřují: **střed stránky**, hranice a středy **formulářových polí/widgetů**, **předěly a dělicí čáry**, **textové rámce**. Nadále platí účaří, hrany a středy objektů a stran, uzly, vodítka, mřížka, okraje stránky a rovnoměrné rozestupy.
- Kandidáty musí pocházet ze skutečné geometrie a transformace PDF: PageBox/MediaBox/CropBox/UserUnit, rotace, zoom, DPI, popř. transformace a bounding-boxy formulářových widgetů, anotací, textů a cest. Hidden/locked/nesprávně zvolený kandidát se musí filtrovat podle aktuálních pravidel.
- HTML v3.17 demonstruje dostupné cíle nad omezenou **ukázkovou HTML stránkou**, nikoliv rozpoznávání všech vektorových prvků skutečného PDF. Nezaměňovat jejich názorné vizuální spojnice za funkční SnapEngine produkční aplikace.
- Průběžné vínově červené dočasné indikátory končí u konkrétního bodu (střed, uzel) nebo mohou pokračovat ke stupnici pravítka (zarovnání hran). Pevná vodítka jsou jemná, plná a mají uživatelskou barvu. Tolerance vůči kurzoru v pixelech musí být nezávislá na PDF zoomu. Ověřit přepočet a hysterézi v další etapé.

## 5. Ochrana před kolizemi a akceptační testy

- **Nepřepisovat** klávesy \`T/V/M/L\`, \`Ctrl+F\`, \`Ctrl+L/E/R/J\`, \`Ctrl+Z\`, \`Delete\`, \`Esc\` a jejich kontextové priority, mapu v1.2, původní univerzální výběr, Text/OCR/Objekty ani dvoustavové horní nástroje. \`Delete\` maže vodítka jen v jejich výběrovém kontextu, jinak zachovává původní význam.
- Otestovat klik na řádek A, řádek B, checkboxy A+B, klik do prázdné plochy, skrytí/obnovení oka, vytvoření skupiny, skupinovou barvu, skupinové zamknutí/skrytí/smazání, rozdělení a přechod na jinou stránku.
- Angličtinu kontrolovat u všech panelů, formulářů, tooltipů, dialogů i OCR; text vlastního PDF a dokumentové názvy **musí zůstat beze změn**.
- Testováno v Chromiu při šířkách 980, 1100, 1250 a 1600 px. HTML interakce a audit hlavních viditelných textů prošly bez nezachycených JS chyb. **Žádný skutečný PDF export, PDFium geometry snapping ani test produkční Go aplikace se zatím neprováděl**.

**Další krok:** definovat datové typy \`SnapCandidate\`, \`Anchor\`, \`SnapResult\`, \`GuideGroup\`, priority a hysterezi, následně samostatně schválit implementaci v Go/PDFium. Zachovat možnost vrátit se ke všem historickým schváleným HTML verzím.
