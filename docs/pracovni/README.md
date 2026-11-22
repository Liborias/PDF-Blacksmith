# PDF Blacksmith – pracovní podklady pro Codex

**Aktualizováno:** 22. listopadu 2026 · **Stav:** návrh rozhraní F1, implementace F1 schválena po etapách; ostatní položky jen plán.

Tato složka je centrální rozcestník průběžné dokumentace. **Nenahrazuje stávající implementaci** v Go/Win32 a Pythonu; přítomnost návrhu neznamená, že byl implementován.

## Začít zde

1. [Postup spolupráce a předání Codexu](POSTUP_SPOLUPRACE_S_CODEXEM.md)
2. [Hlavní specifikace PDF Blacksmith v5.0](specifikace_Codex_v5.0.md)
3. [Plán rozšíření v2.0](plan_rozsireni_v2.0.md)
4. [Formuláře F1 – rozhodnutí F-Q1 až F-Q9](formulare/F1_rozhodnuti.md)
5. [Formuláře F1 – detailní specifikace v1.9](formulare/F1_specifikace_v1.9.md)
6. [Formuláře F1 – implementační postup](formulare/F1_plan_implementace.md)
7. **[Aktuální HTML návrh v3.1: Text se záložkami Výběr / Typografie a pokročilými sekcemi](rozhrani/cely_vizual_text_zalozky_v3.1.html)** – pracovní prototyp, bez editace skutečného PDF.
   - **[Pokyny pro Codex: bezkolizní výběr textu a nové záložky](rozhrani/TEXT_VYBER_A_ZALOZKY_v3.1.md)** – doplněk k SEL/TXT/OBJ pravidlům, zachovat schválená gesta.
   - [Předchozí v3.0: rozšířená typografie a nové SVG ikony](rozhrani/cely_vizual_typografie_v3.0.html) pro porovnání.
   - **[Nové pravidlo automatického otevření pomocných panelů](rozhrani/PRAVIDLO_AUTOMATICKE_POMOCNE_PANELY_2026-11-22.md)** — pouze zaznamenáno; HTML zatím neměnit.
   - [Předchozí v2.9: globální Nastavení a šířka respektující ouška](rozhrani/cely_vizual_nastaveni_bezpecna_sirka_v2.9.html) pro porovnání.
   - [Předchozí návrh v2.8: historie a sdílené pomocné panely](rozhrani/cely_vizual_panely_historie_v2.8.html) pro porovnání.
   - [Předchozí verze v2.7 – vložení stran a posuvníky](rozhrani/cely_vizual_vlozit_strany_panely_v2.7.html) pro porovnání.
   - [Předchozí návrh panelů v2.6](rozhrani/cely_vizual_panely_ouska_v2.6.html) pro porovnání.
   - **[Rozlišení Vložit / Importovat / Sloučit a otevřít (Otevřít pouze PDF; širší formáty)](rozhrani/ROZHODNUTI_Vlozit_Importovat_Sloucit_2026-11-22.md)** — schválený směr, HTML a kód zatím neupravovat.
   - **[Pokyny k ouškům a posuvníkům](rozhrani/ODLOZENE_UPRAVY_PANELU_PO_v2.6.md)** — zaneseny do HTML v2.7; zatím neimplementovány v Go.
   - [Starší nabídky Soubor/Zobrazení v2.4](rozhrani/cely_vizual_soubor_zobrazeni_v2.4.html) pro porovnání.
   - [Předchozí návrh v2.3](rozhrani/cely_vizual_soubor_v2.3.html) pro porovnání.
   - [Původní rozložení v2.2](rozhrani/cely_vizual_vsechny_nastroje_v2.2.html).
   - [Starší varianta v2.1](rozhrani/cely_vizual_dve_listy_v2.1.html) pro porovnání.
8. [Původní návrh Formulářů F1](formulare/F1_mockup_v1.0.html)
9. [Vyhledávání – budoucí upřesnění](vyhledavani_navrh_v0.1.md)
10. [Mapa ikon v4.4](mapa_ikon_v4.4.md)

## Závazné rozlišení

- **Schváleno k postupné implementaci:** Formuláře **F1** (čtení, vyplňování, bezpečné ukládání) a dříve schválená pravidla P0.
- **Plán, nikoli povolení k implementaci:** F2 návrh polí, F3 zploštění, Preflight, grafické porovnání, měření, dávky, nová vyhledávací šipková anotace.
- **Vizuální prototyp** je jen HTML demonstrace, nikoli nový webový frontend ani funkční PDF editor.
- Nástroje Text, Objekty, OCR, Separace, podpis, náhledy, tisk a export se nesmí při přidávání F1 zjednodušit nebo ztratit.
- Při ukládání/volbě „export bez poznámek“ nikdy nesmazat formulářové widgety, odkazy či podpisové anotace.

## Pravidla průběžného aktualizování

Změnu návrhu nejprve odsouhlasit v chatu, aktualizovat příslušný Markdown/HTML, přidat přehledný commit. Ověřenou implementaci označovat za „hotovou“ až po testech, ne po napsání specifikace. Zachovávat historii změn v Gitu.

**Veřejné úložiště:** neukládat do něj soukromé nebo zákaznické dokumenty, certifikáty, hesla, tokeny či jiné neveřejné podklady.

### Krátké zadání pro Codex

> Přečti `docs/pracovni/README.md`, plán F1 a detailní specifikaci. Prozkoumej současný Go/Win32 a Python/PDFium kód a nejdřív připrav seznam konkrétních kroků, rizik a testů. Bez dalšího schválení nezačínej F2/F3 ani jiné plánované rozšíření. Zachovej všechny existující funkce.
