# PDF Blacksmith — ROZCESTNÍK FINÁLNÍ DOKUMENTACE v6.0

**Aktualizováno:** 25. listopadu 2026 · **Jediné závazné zadání pro Codex** je vydání **v6.0**. Historické dokumenty zachované v tomto adresáři **nejsou další aktuální implementační pokyny**.

## Začít zde (nově jediná správná cesta)

1. **[CODEX_START.md](CODEX_START.md)** — nejkratší vstupní pokyn.
2. **[CODEX_MASTER_v6.0.md](CODEX_MASTER_v6.0.md)** — jedna kompletní konsolidovaná specifikace.
3. **[CODEX_IMPLEMENTACNI_PLAN_v6.0.md](CODEX_IMPLEMENTACNI_PLAN_v6.0.md)** — nový fázový plán s prioritou **rozhraní, nikoli formuláře**.
4. **[CODEX_KONSOLIDACE_ROZPORU_v6.0.md](CODEX_KONSOLIDACE_ROZPORU_v6.0.md)** — doložené rozpory a konkrétní náhrady, jen pro audit.

## Jediné finální mapy

| Oblast | Závazná mapa | Strojově |
|---|---|---|
| Ikony | [mapa_ikon_FINAL_v6.0.md](mapa_ikon_FINAL_v6.0.md) | [JSON](mapa_ikon_FINAL_v6.0.json) |
| Barvy / motivy | [mapa_barev_FINAL_v6.0.md](mapa_barev_FINAL_v6.0.md) | [JSON](mapa_barev_FINAL_v6.0.json) |
| Klávesy / gesta | [mapa_zkratek_FINAL_v6.0.md](mapa_zkratek_FINAL_v6.0.md) | [JSON](mapa_zkratek_FINAL_v6.0.json) |
| Lokalizace | [i18n/ui_catalog_FINAL_v6.0.json](i18n/ui_catalog_FINAL_v6.0.json) | [CSV](i18n/ui_catalog_FINAL_v6.0.csv) |

## Poslední vizuální návrh

**[HTML v3.26 – kompletní interaktivní prototyp](rozhrani/cely_vizual_v3.26.html).**
Zahrnuje barevné režimy, nástroje a panely, zoom Ctrl+kolečko, pružný přehled stran, výběry/pravítka/vodítka a demonstrační návrh Formuláře → Vyplnit / Návrh polí F2 / Z dat. Funkce F2, batch XLSX/Google a skutečné přepisování PDF **nejsou hotovou produkční implementací**.

## Nová priorita implementace

1. **Audit** aktuálního stavu Go/Win32, dosavadních rozpracovaných změn (včetně F1).
2. **Uživatelské rozhraní:** design systém, navigace, panely, ikony, motivy, jazyk, přístupnost a zkratky; potom zobrazení stránek/zoom a sjednocený výběr.
3. **Formuláře F1**: dříve schválený funkční rozsah, **termín až po UI a novém explicitním schválení**.
4. **F2, import CSV/XLSX/Google, F3, další moduly**: pouze plány.

**Závazná bezpečnost:** existující Go/Win32/PDFium nástroje, OCR, separace, podpisy, anotace, obsah a uživatelské barvy PDF nesmějí být ztraceny nebo nahrazeny. Žádné změny produkčního kódu bez schválení etapy.

## Historický archiv — jen při zjišťování původu rozhodnutí
Soubory jako \`specifikace_Codex_v5.0.md\`, \`plan_rozsireni_v2.0.md\`, \`POSTUP_SPOLUPRACE_S_CODEXEM.md\`, \`formulare/F1_plan_implementace.md\`, jednotlivé \`rozhrani/UI_...v3.x.md\`, staré HTML a mapy ikon v4.7/v5.1, barev v1.4 a zkratek v1.4 jsou **předchůdci této konsolidace**. Nejsou zrušená uživatelská rozhodnutí: podstatná a stále platná pravidla byla přenesena do v6.0. Codex **nemá jejich staré „aktuální“ poznámky interpretovat jako další prioritu**. Chybějící technický detail lze dohledat pouze v rámci explicitního auditu nebo po novém rozhodnutí autora.
