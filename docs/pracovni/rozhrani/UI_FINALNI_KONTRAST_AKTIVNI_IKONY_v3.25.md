# PDF Blacksmith – UI v3.25: závěrečné sjednocení povrchů a aktivních ikon
**Datum:** 24. 11. 2026. **Stav:** revize HTML prototypu dle pěti referenčních snímků; produkční Go/Win32/PDFium beze změny.

**HTML:** [v3.25](cely_vizual_v3.25.html). **Předchozí:** [v3.24](cely_vizual_v3.24.html). **Palety:** [mapa v1.4](../mapa_barev_v1.4.md), [JSON v1.4](../mapa_barev_v1.4.json).

## Opravy podle referenčních snímků
1. **Soubor/Otevřít:** štítek „Připnuté“ nyní používá příslušný akcent a čitelný text, nikoliv pevnou vínovou. Pomocné a stavové texty odpovídají paletě.
2. **Objekty:** tmavý informační proužek „Bez výběru…“ nahrazen sémantickým světlým/tmavým povrchem. Text přebírá kontrastní barvu.
3. **Text:** souhrn „Řádek · pouze označené znaky“, jeho hlavní a doplňkové popisky a ukázkový blok nejsou tmavé v režimu light.
4. **OCR:** stejný výběrový proužek a popisky v nástrojovém panelu používají paletu.
5. **Poznámky:** popis uvnitř poznámky má čitelný sekundární text, kontrastní podklad a stav hover; žádný pevný vínový či tmavý fallback.
6. **Ikony ve druhé horní liště:** výraznější aktivní stav pomocí tónované plochy z příslušného schématu, kontrastního obrysu a spodní linky. `toolbtn.active` a přístupné stavy `aria-pressed=true` (mimo tlačítko přepínače motivu) mají jednotné chování. Hover zůstává patrný. Hlavní textová menu Soubor/Zobrazení/… zůstávají bez rámů.

## Zachovat
Uživatelská data, obsah PDF, miniatury jako obsah, ICC, uživatelské barvy štítků/vodítek, interakce, pokyny v3.1–v3.24, funkce OCR, textu a výběru, F1, Undo/Redo a původní HTML verze. Nedoplňovat produkční funkce nad rámec schváleného F1.

## Testy
Všech šest schémat × light/dark, rozhraní pro pět vyobrazených panelů, textové hlavní menu vs ikonová lišta, přepínání ikon a hover; porovnat vypočtené barvy a WCAG kontrasty. Statický zápis CSS není náhradou za end-to-end browser test.
