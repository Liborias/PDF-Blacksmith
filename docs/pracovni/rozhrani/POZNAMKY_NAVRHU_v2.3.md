# PDF Blacksmith – doladění nabídky Soubor (HTML návrh v2.3)

**Stav:** rozpracovaný HTML vizuální prototyp, ne implementace Go/Win32 ani schválení dalších funkcí.  
**Datum:** 22. listopadu 2026.

## Co bylo změněno

1. Původní **SVG kovadlina** beze změny geometrie zůstala v levém horním rohu; šipky historie **Zpět a Znovu** jsou od loga vzdálenější.
2. Druhá horní lišta má **Soubor** s rozbalením **po najetí myší**, podporuje i kliknutí a otevření z klávesnice.
3. Jediná společná definice `FILE_ACTIONS` ovládá horní rozbalovací nabídku a **odsazené podmenu Soubor** na levé straně:
   - Otevřít
   - Importovat
   - Uložit
   - Uložit jako
   - Sloučit a otevřít
   - Exportovat
   - Tisknout
   - Vlastnosti
4. Zvolený příkaz zůstává vlevo **jemně zvýrazněn**; na vrchu levého panelu se pod Soubor podmenu automaticky otevře, i pokud byl panel dříve sbalený.
5. Pravý panel **Vlastnosti** nyní nezdvojuje osm hlavních akcí, ale ukazuje kontextová **pokročilá nastavení vybraného příkazu**: zdroj, cílový dokument, rozsah stran, barvy a ICC, metadata, tiskové volby, apod.
6. „Sloučit PDF“ v této vizuální variantě není druhá samostatná položka hlavních nástrojů: **Sloučit a otevřít** v Soubor navazuje na původní koncepci slučování, kterou musí produkční aplikace zachovat. Zobrazování a ovládání stávajícího nástroje se smí ve skutečném Go kódu měnit teprve po schválení.

## Chování prototypu

- Horní nabídka reaguje na hover a kliknutí; po volbě se otevře odpovídající pravý panel a zůstane vybrané podmenu vlevo.
- Každá z osmi voleb má odlišná nastavení v pravém panelu.
- Otevřít/Importovat/Přidat PDF může spustit systémový výběr lokálního PDF v prohlížeči, ale **čte jen název souboru** – nedochází k odeslání obsahu ani k editaci PDF.
- Uložit/Uložit jako vedou jen na existující ilustrační dialog HTML.
- Exportovat, Tisknout a Sloučit jsou pouze grafické návrhy nastavení. Není prováděna skutečná PDF operace.
- Stávající panel Záložky, Poznámky, všechny ostatní nástroje a horní/dolní navigace jsou zachovány.

## Návrhy k pozdějšímu odsouhlasení

Do menu **zatím nepřidáno** kvůli přehlednosti: „Nedávné soubory“, „Zavřít dokument“, případně „Nový dokument“. Po schválení lze zavést jako oddělenou skupinu či kontextovou položku; některé další činnosti (např. přeuspořádání stránek) mají zůstat ve vlastních nástrojích.

## Kontrola

Ověřeno v Chromium: všech osm voleb aktualizuje levé zvýraznění a pravý panel, hover/click fungují, Esc menu zavře, klávesnice otevře a podmenu zachovává výběr i po přepnutí na Text a zpět; JavaScript bez chyb. Ostatní části HTML zůstávají jen ilustrací.

**Pro Codex:** používat tento soubor jako návrh vzhledu a interakcí; před jakoukoli změnou kódu porovnat se stávajícími příkazy Soubor, ukládáním a slučováním. Nepřepisovat starší hotovou funkcionalitu pouze podle zjednodušeného HTML.
