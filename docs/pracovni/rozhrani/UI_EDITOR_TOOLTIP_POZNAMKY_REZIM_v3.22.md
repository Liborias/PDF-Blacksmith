# PDF Blacksmith – UI v3.22: editor, tooltipy, poznámky a rychlý motiv
**Datum:** 24. 11. 2026. **Status:** interaktivní HTML prototyp; produkční Go/Win32/PDFium beze změny.
**HTML:** [v3.22](cely_vizual_v3.22.html). **Předchozí:** [v3.21](cely_vizual_v3.21.html). **Mapa:** [v1.3](../mapa_barev_v1.3.md), [JSON](../mapa_barev_v1.3.json).

1. Editorová plocha mimo PDF se odvozuje ze základní palety: nikdy pevná černá napříč tmavými motivy. Obsah bílých stránek PDF se nemění.
2. Plovoucí nápověda `.floating-icon-tooltip` má barevné tokeny pro plochu, text, obrys a neutrální stín ve světlém i tmavém režimu.
3. Stavové proužky a informativní texty se řídí schématem.
4. Šipkové štítky poznámek standardně přebírají akcent motivu. V náhledu výsledků je výběr Automaticky / Vlastní barva, potvrzení Použít. U demonstračních štítků lze kliknutím zadat individuální hex barvu; prázdná vrátí automatické nastavení. Text se přepíná na kontrastní černou/bílou. Neměnit skutečné poznámkové anotace v PDF.
5. V horní liště vedle ozubeného kolečka je půlený kruhový přepínač režimů. Ovládá stejný stav `settingsApplied.theme` / při otevřeném dialogu `settingsDraft.theme`; původní volba v Nastavení → Rozhraní zůstává.
6. Horní logo (kovadlina) o ~1 mm nižší, ozubené kolečko opticky menší než původních 20 px (18 px), textový kurzor zmenšený. Nemenit geometrii originálních SVG symbolů ani vlastní velikosti ostatních ikon.
7. Nadále platí všechna schválená gesta, navigace, výběry, lokalizace, F1, ICC, separace a Undo.

**Ověření:** kontrola existence tokenů a napojení interakcí v HTML nestačí jako provedený vizuální test všech dvanácti kombinací. Před implementací v Go ověřit prohlížečem půlený přepínač, zrušení nastavení, změny motivů, kontrast štítků a všechny plovoucí popisky.
