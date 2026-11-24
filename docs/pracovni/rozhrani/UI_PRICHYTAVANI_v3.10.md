# PDF Blacksmith – UI v3.10: magnet, vodítka a cíle přichytávání

**Datum:** 24. listopadu 2026  
**Stav:** schválené změny **pouze pracovního HTML návrhu a dokumentace**, bez úpravy produkční Go/Win32/PDFium aplikace.  
**Navazuje na:** [konsolidaci UI v3.9](AKTUALNI_UI_CODEX_v3.9.md) a [aktuální HTML v3.10](cely_vizual_v3.10.html).  
**Mapa ikon:** [v4.6](../mapa_ikon_v4.6.md) a [SVG atlas v3.10](ikony/pdf_blacksmith_nove_v3.10.svg).

## Uživatelem schválené změny

1. **Magnet \`i-snap_magnet\`:** upravit kresbu na běžný **podkovový magnet** – silná a zřetelná U silueta, opticky oddělené horní póly. Zachovat význam i ID ve druhé horní liště, \`#magnetToggle\`, checkbox v Zobrazení a \`viewAids.snap\`.
2. **Vodítka \`i-toggle_guides\`:** zřídit samostatný symbol odlišný od \`i-toggle_rulers\`: dvě vodorovné úsečky v horní části a dvě svislé vedle sebe v levé části. \`Zobrazení → Zobrazit vodítka\` používá výhradně nový symbol. Pravítka a mřížka nadále vlastní odlišné ikony.
3. **Rozbalovací nabídka Zobrazení:** nejprve položky Zobrazit mřížku, Zobrazit pravítka, Zobrazit vodítka, **Přichytávání zapnuto**; bezprostředně poté **další vodorovný oddělovač** a rozbalovací skupina **Přichytávat k:**.

## Cíle v podmenu „Přichytávat k:“

| Pořadí | Popisek | Klíč \`viewAids.targets\` |
| --- | --- | --- |
| 1 | Účaří textu | \`baselines\` |
| 2 | Okraje objektů | \`edges\` |
| 3 | Vodicí linky | \`guides\` |
| 4 | Středy objektů | \`centers\` |
| 5 | Středy stran objektů | \`sideMidpoints\` |
| 6 | Uzly křivek | \`nodes\` |
| 7 | Mřížka | \`grid\` |
| 8 | Okraje a střed stránky | \`page\` |
| 9 | Rovnoměrné rozestupy | \`spacing\` |

**Stavová pravidla:** hlavní magnet je samostatné zapnuto/vypnuto. Vypnutí neruší preference cílů. Viditelnost mřížky, pravítek a vodítek má oddělené ovládání; nezaměňovat s cíli přichytávání. V HTML se změny checkboxů odrážejí bez změny obsahu PDF. Rozbalení podmenu a zobrazení menu lze ovládat myší i klávesnicí, výška nabídky je omezena dostupným oknem a menu se může posouvat.

## Výslovně zachovat z v3.9

- Dvoufunkční tlačítko Šipka / Textový kurzor a Po stránkách / Souvislé zobrazení; ikona ukazuje akci *následujícího* kliknutí.
- Univerzální obdélník a laso, společný výběrový dispatcher, automatická volba Text/Objekty/OCR podle obsahu, pravidla historie a modifikátorů.
- Pevná výška kontextové nápovědy pod ikonami, popisky nad ikonou, čistý panel Typografie.
- Text → Výběr a společná OCR cesta pro vektorový/rastrový obsah, dialog rekonstrukce, jednoduché úpravy obrázků.
- Pravidla z \`docs/pracovni/specifikace_Codex_v5.0.md\` a starší schválené SEL/TXT/OBJ/F1; žádná hromadná implementace bez zvláštního odsouhlasení.

## Akceptační testy návrhu HTML

1. \`i-snap_magnet\`, \`i-toggle_rulers\`, \`i-toggle_grid\`, \`i-toggle_guides\` existují v SVG a vodítka nepoužívají ikonu pravítek.
2. Menu Zobrazení obsahuje přesně jeden master checkbox pro snap, viditelné oddělení a pod ním nadpis **Přichytávat k:**; devět přepínačů se správnými stavovými klíči a zachovanými hodnotami.
3. Zavření a opětovné otevření nabídky neztrácí nastavené cíle; vypnutí magnetu nemění výběr kategorií.
4. V úzkém okně (980×540 CSS px) lze menu a všechny volby posouváním dosáhnout, bez nechtěných překryvů toolbaru nebo změny PDF.
5. Prověřit regresně výběrové nástroje a obě dvoustavová tlačítka z v3.9; konzolové JavaScript chyby musí být nulové.

**Pozor:** SnapEngine zatím nebyl implementován. Tyto ikonky a checkboxy jsou návrhem UI a preference budoucí geometrie; nesmí předstírat vykonaný přesun či přichycení objektu v PDF.
