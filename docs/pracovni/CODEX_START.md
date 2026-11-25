# START PRO CODEX — PDF Blacksmith v6.0

**Toto je jediné aktuální vstupní zadání.** Staré iniciační dokumenty a historická specifikace v5.0 byly **nahrazeny** jednotnou vydanou sadou 25. listopadu 2026.

## Nejdříve a pouze tyto dokumenty
1. [CODEX_MASTER_v6.0.md](CODEX_MASTER_v6.0.md) — veškerá závazná pravidla projektu, chování UI, ochrany funkcí a PDF.
2. [CODEX_IMPLEMENTACNI_PLAN_v6.0.md](CODEX_IMPLEMENTACNI_PLAN_v6.0.md) — **nové pořadí: rozhraní před formuláři**.
3. [mapa_ikon_FINAL_v6.0.md](mapa_ikon_FINAL_v6.0.md) + [strojový JSON](mapa_ikon_FINAL_v6.0.json).
4. [mapa_barev_FINAL_v6.0.md](mapa_barev_FINAL_v6.0.md) + [strojový JSON](mapa_barev_FINAL_v6.0.json).
5. [mapa_zkratek_FINAL_v6.0.md](mapa_zkratek_FINAL_v6.0.md) + [strojový JSON](mapa_zkratek_FINAL_v6.0.json).
6. [i18n/ui_catalog_FINAL_v6.0.json](i18n/ui_catalog_FINAL_v6.0.json), případně [CSV](i18n/ui_catalog_FINAL_v6.0.csv) — **stabilní** CZ/EN klíče.
7. **Jediný referenční HTML prototyp**: [rozhrani/cely_vizual_v3.26.html](rozhrani/cely_vizual_v3.26.html).

Historické dokumenty v3.x/v4.x/v5.x včetně „specifikace_Codex_v5.0.md“, „F1_plan_implementace.md“, „plan_rozsireni_v2.0.md“ a starých map nejsou konkurenční instrukce. **Codex je nemá procházet, porovnávat ani slučovat** při běžném vývoji. Pouze pokud se řeší konkrétní provenience či chybějící fakt, lze se na ně obrátit; aktuální rozhodnutí už jsou v master v6.0. Audit rozporů [CODEX_KONSOLIDACE_ROZPORU_v6.0.md](CODEX_KONSOLIDACE_ROZPORU_v6.0.md) je pro kontrolu, ne další variantu zadání.

## První pracovní postup
1. Ověř skutečný stav větve main, změny a novější commity, rozpracované soubory, zejména již započaté F1/formuláře. Nic nemazat, nedělat reset, nepoškozovat práci.
2. Prostuduj **existující produkční Go/Win32/PDFium kód a základní kořenové PLAN.md / NAVOD.md** jen za účelem technické inventury. **HTML v3.26 není implementace**.
3. Připrav přehled aktuálního **UI**: již fungující/částečné/plánované, rozdíly proti schválenému prototypu, rizika a testy.
4. Navrhni **jediný malý první krok v rozhraní** (Etapa 1). F1 je obsahově schválené, ale **v této chvíli má nižší prioritu a nepokračuje se v jeho implementaci**. F2 a import dat jsou návrhy.
5. **Dokud uživatel konkrétní krok neschválí, nic neimplementuj, neupravuj, necommituj ani neposílej na GitHub.**

## Ochrana architektury
Zachovat Go/Win32, PDFium a všechny existující nástroje včetně separací, OCR, objektů, textu, ukládání, podpisů, historie, AutoRecover a formulářových widgetů. Nevytvářet náhradní „jednodušší PDF editor“. Nové barvy a ikony jsou UI, nikoliv zápis do PDF. Nevykonávat PDF JavaScript ani neexportovat citlivá data.

**První odpověď uživateli:** audit rozhraní a rozpracovaných změn, seznam rizik/kolizí a návrh prvního drobného kroku implementace UI — bez vlastních commitů.
