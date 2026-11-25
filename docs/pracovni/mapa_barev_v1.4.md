> **ARCHIV / NAHRAZENO 25. 11. 2026:** Tento dokument slouží pouze k dohledání historie rozhodnutí. **NENÍ samostatným aktuálním pokynem pro Codex.** Jediné závazné zadání: [CODEX_START.md](/Liborias/PDF-Blacksmith/blob/main/docs/pracovni/CODEX_START.md) a [CODEX_MASTER_v6.0.md](/Liborias/PDF-Blacksmith/blob/main/docs/pracovni/CODEX_MASTER_v6.0.md). **Nová priorita: nejprve rozhraní, F1 až později po schválení.** Platí finální mapy v6.0. Původní historický obsah níže zůstává zachován.

---

# PDF Blacksmith – mapa barev v1.4: doplnění aktivních ikon

**Aktuální HTML:** `cely_vizual_v3.25.html`. Rozšiřuje [v1.3](mapa_barev_v1.3.md), všechny předchozí role, schémata i historické kontroly zůstávají zachované.

Nové tokeny `toolbarActive` a `toolbarActiveHover` jsou odvozeny z `selected` a `accentStrong` každé palety (6 × 2). Používají se **výhradně pro aktivní ikonová tlačítka ve druhé liště**, nikoli pro textové hlavní menu. Související stavové pruhy, seznamy poznámek, označení „Připnuté“ a popisky jsou řešeny existujícími rolemi `subtle`, `text`, `muted`, `accentStrong` a `onAccent`.

Náhledy PDF, barvy dokumentu, ICC, skutečné hodnoty swatchů, individuální barvy vodítek a poznámek zůstávají nedotčené. Vizuální kontrolu všech stavů v prohlížeči je nutné provést zvlášť.
