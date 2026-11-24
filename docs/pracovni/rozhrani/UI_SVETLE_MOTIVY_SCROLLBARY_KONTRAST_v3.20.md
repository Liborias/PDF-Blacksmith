# PDF Blacksmith – UI v3.20: kompletní světlé motivy, posuvníky a ouška

**Datum:** 24. 11. 2026. **Stav:** aktualizace HTML prototypu a podkladů Codex; bez automatické implementace Go/Win32/PDFium.

**Aktuální HTML:** [v3.20](cely_vizual_v3.20.html). **Předchozí:** [v3.19](cely_vizual_v3.19.html), [pokyny v3.19](UI_MOTIVY_KONTRAST_PREKLADY_v3.19.md). **Mapa barev:** [v1.1](../mapa_barev_v1.1.md) a [JSON](../mapa_barev_v1.1.json).

## Závazné úpravy
- Světlé motivy: žádné tmavé fallback plochy u záložek, textových nástrojů, tlačítek, boxů, ovládání a posuvníků. Text a sekundární popisky zůstávají kontrastní.
- Levá ouška pomocných panelů (Výsledky, Záložky, Poznámky, Historie), pravá ouška a **obě horní ouška hlavních panelů** používají stejný jemný hover na celé ploše. Světlý hover nesmí přejít do tmavého schématu. Po opuštění se vrací odpovídající idle/active podklad; aktivní ouško stále vizuálně navazuje na panel.
- Posuvníky: samostatná světlá dráha, thumb a hover thumb ve světlých motivech; zrcadlově tmavé v tmavých. Žádné pevné tmavé barvy.
- Textové záložky Výběr/Typografie, ikony volby, výstupní boxy a sekundární informace přebírají sémantické tokeny všech šesti palet.
- Mapa v1.1 definuje panelTabIdle/Hover/Active, tabIdle/Hover/Active, scrollTrack/Thumb/ThumbHover, interactiveText/helpText. Kontrolovat též skutečné vypočtené CSS.

## Zachování stávající práce
Žádné změny SEL/TXT/OBJ/F1, Ctrl+klik, Ctrl+tažení, zásad vodítek, SnapEngine, klávesových zkratek, dokumentových hodnot, ICC, separace, OCR, Undo/Redo a způsobu sbalování/rozbalování. Vzhled aplikace nesmí přebarvovat obsah PDF.

## Ověření
Ve všech 6 × 2 paletách změřit computed styles idle/hover/active/focus hlavních a pomocných oušek, scrollbars, textových tabů, panelů, tlačítek, vstupů a sekundárních textů. Běžný text má dosáhnout kontrastu alespoň 4,5:1. Pro relevantní grafické prvky a fokus alespoň 3:1. Kontroly uvedené v JSON jsou kontroly deklarovaných tokenů, nikoliv důkaz úplného end-to-end testování prohlížeče nebo Go aplikace.
