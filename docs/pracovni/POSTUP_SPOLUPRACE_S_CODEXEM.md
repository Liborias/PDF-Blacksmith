> **ARCHIV / NAHRAZENO 25. 11. 2026:** Tento dokument slouží pouze k dohledání historie rozhodnutí. **NENÍ samostatným aktuálním pokynem pro Codex.** Jediné závazné zadání: [CODEX_START.md](/Liborias/PDF-Blacksmith/blob/main/docs/pracovni/CODEX_START.md) a [CODEX_MASTER_v6.0.md](/Liborias/PDF-Blacksmith/blob/main/docs/pracovni/CODEX_MASTER_v6.0.md). **Nová priorita: nejprve rozhraní, F1 až později po schválení.** Platí finální mapy v6.0. Původní historický obsah níže zůstává zachován.

---

# Pracovní postup – ChatGPT, GitHub, Codex

1. V chatu probereme nové požadavky a uživatel schválí přesné varianty.
2. Schválení zapíšeme do `docs/pracovni/` a do relevantní specifikace.
3. Codex začne čtením [README](README.md), [implementačního postupu F1](formulare/F1_plan_implementace.md) a hlavní specifikace; nejprve ověří, co projekt skutečně umí.
4. Implementuje se jen konkrétně vyžádaná etapa a menší, samostatně testovatelná část. Existující nástroje se nesmějí nahradit jednodušší verzí.
5. Před každým PR/commitem vyhodnotí testy a dopady na ukládání, PDF kompatibilitu, render i UX.
6. V chatu vyhodnotíme výsledek a případně doplníme specifikaci, nikoli neověřené prohlášení „hotovo“.

## Kontrolní body

- Rozlišit skutečný stav repozitáře od návrhu; v kořeni číst `PLAN.md` a `NAVOD.md`.
- Zohlednit `workspace_windows.go`, `viewer_windows.go` a ukládání/export; PDFium, pikepdf/Ghostscript používat podle skutečných závislostí projektu.
- F1 nesmí při ukládání odstranit `/AcroForm`, `/Fields`, `/Widget`, podpisy nebo odkazy.
- Nespouštět libovolný PDF JavaScript, XFA bezpečně identifikovat a neslibovat plnou podporu.
- Testovat českou diakritiku, vzhled `/AP`, hodnoty `/V`, opakovaná pole, uzavření/obnovu, chybný formát a export bez poznámek.
- Skutečné UI se ověřuje ve Windows; HTML soubor slouží jen k vizuálnímu a interakčnímu návrhu.

**Doporučený prompt:** „Přečti `docs/pracovni/README.md`, `docs/pracovni/formulare/F1_specifikace_v1.9.md` a `docs/pracovni/formulare/F1_plan_implementace.md`. Nejdřív mi ukaž analýzu aktuálního kódu, navržené kroky F1 a regresní testy, teprve po mém pokynu implementuj konkrétní krok. F2/F3 a další plánované moduly zatím neimplementuj.“
