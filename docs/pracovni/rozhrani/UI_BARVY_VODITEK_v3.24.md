# PDF Blacksmith – UI v3.24: světlé podklady barevných polí vodítek

**Datum:** 24. 11. 2026 · **Stav:** drobná cílená oprava HTML prototypu, nikoli implementace produkční Go/PDFium aplikace.

**HTML:** [v3.24](cely_vizual_v3.24.html) · **Předchozí verze:** [v3.23](cely_vizual_v3.23.html) · **Mapa barev:** [v1.3](../mapa_barev_v1.3.md).

## Problém a příčina
Ve světlém motivu zůstávala tmavá políčka s barvami jednotlivých vodítek X/Y. Starší obecné pravidlo `.layout-guide-row input` používalo pevný tmavý podklad a v předchozích opravách chyběl přesný selektor `.layout-guide-row input[type=color]`.

## Oprava
- Zacílit skutečné barevné inputy v řádcích vodítek, skupinách, panelu označených vodítek a místní nabídce vodítka.
- Pro podklad, rámeček a hover použít sémantické proměnné aktivní palety; barevný vzorek uvnitř zůstává skutečnou barvou vodítka.
- U nativního color inputu nastavit také WebKit/Firefox swatch wrapper, aby nevznikal tmavý vnitřní obrys.
- Tlačítko potvrzení barvy musí využívat stejné světlé/tmavé paletové tokeny.
- Neměnit vybrané barvy vodítek, logiku potvrzení Použít, skupiny, souřadnice, výběr, Ctrl+klik, Ctrl+tažení, Undo, PDF či ICC.

## Akceptace
Zkontrolovat pole barvy u dvou jednotlivých vodítek, výběru více vodítek, skupinových barev a místní nabídky, ve všech šesti paletách ve světlém i tmavém režimu. Náhled původní barvy musí zůstat přesný. Toto je kontrolní požadavek; samotný zápis CSS není důkazem provedeného browser testu.
