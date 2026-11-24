# PDF Blacksmith – mapa barev v1.4: doplnění aktivních ikon

**Aktuální HTML:** `cely_vizual_v3.25.html`. Rozšiřuje [v1.3](mapa_barev_v1.3.md), všechny předchozí role, schémata i historické kontroly zůstávají zachované.

Nové tokeny `toolbarActive` a `toolbarActiveHover` jsou odvozeny z `selected` a `accentStrong` každé palety (6 × 2). Používají se **výhradně pro aktivní ikonová tlačítka ve druhé liště**, nikoli pro textové hlavní menu. Související stavové pruhy, seznamy poznámek, označení „Připnuté“ a popisky jsou řešeny existujícími rolemi `subtle`, `text`, `muted`, `accentStrong` a `onAccent`.

Náhledy PDF, barvy dokumentu, ICC, skutečné hodnoty swatchů, individuální barvy vodítek a poznámek zůstávají nedotčené. Vizuální kontrolu všech stavů v prohlížeči je nutné provést zvlášť.
