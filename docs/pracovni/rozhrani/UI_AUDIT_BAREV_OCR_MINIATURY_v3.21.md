# PDF Blacksmith – UI v3.21: audit zbytkových barev všech motivů

**Datum:** 24. listopadu 2026. **Stav:** oprava HTML prototypu + dokumentace pro Codex, ne produkční implementace.

## Zjištěné chyby v3.20
Historické natvrdo vložené vínové rámečky miniatur a přidružené stavy výběru; tmavé pozadí barevných inputů a voličů v Rozvržení; tmavé oblasti v OCR a některých textech; hardcoded tmavé nápovědy, průběhová pole, nástroje, vybrané skupiny, okraje a popisky.

## Opravené součásti
- Miniatury: aktivní outline, vybraný rám, číslo stránky a hover používají paletu, nikoliv vínové RGB hodnoty. Obsah miniatur zůstává dokumentem.
- Rozvržení: výběr barev má světlý/tmavý podklad podle motivu, bez zásahu do samotného uživatelem vybraného barevného vzorku. Tlačítka Použít, vstupy a popisky používají tokeny.
- OCR: náhledová poznámka, tlačítka režimu a smíšené stavy nemají natvrdo tmavý podklad ve světlém motivu.
- Další: nápovědy, seznamy výsledků, historie, poznámky, stavové bloky, nabídky, textové pomocné prvky, ikony a vybraná tlačítka.
- Zachovány všechny interakce oušek ze v3.20, světlé posuvníky, uživatelské barvy vodítek, vlastní barvy UI a šest kompletních palet.

## Výjimky a regresní hranice
Netematizovat obsah skutečného PDF, ukázkové CMYK separační vzorky, pixelová data, skutečné uživatelem vybrané barvy, obrázkové náhledy ani uživatelské dokumentové barvy. Nepřepisovat datové vazby a gesta SEL/TXT/OBJ, OCR logiku, F1, výběry, undo a správu stran.

## Přijetí
Prohlédnout 6 × 2 varianty včetně miniaturních rámů a výběru stran, OCR, Pomůcek rozvržení a barev, převodů, textu, formulářů, nabídek a všech pomocných panelů. Pro každý stav idle/hover/active/disabled zkontrolovat computed styles v prohlížeči. Text WCAG AA 4,5:1; relevantní grafické rozlišení 3:1. **Statická kontrola a doplněná CSS pravidla nejsou úplným vizuálním ani funkčním testem v prohlížeči.**
