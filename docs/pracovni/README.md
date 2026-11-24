# PDF Blacksmith – pracovní podklady pro Codex

**Aktualizováno:** 24. listopadu 2026 · **Stav:** návrh rozhraní F1, implementace F1 schválena po etapách; ostatní položky jen plán.

Tato složka je centrální rozcestník průběžné dokumentace. **Nenahrazuje stávající implementaci** v Go/Win32 a Pythonu; přítomnost návrhu neznamená, že byl implementován.

## Nejnovější stav rozhraní (24. listopadu 2026)

**Aktuální návrh:** **v3.11** ([HTML](rozhrani/cely_vizual_v3.11.html) + [specifikace pro Codex](rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md) + [mapa ikon v4.7](mapa_ikon_v4.7.md) + [mapa klávesových zkratek v1.0](mapa_klavesovych_zkratek_v1.0.md)). V3.11 upřesňuje rozložení Separací, dočasný panel Výsledků, pevná vodítka, chytré čárkované linky, jemnost mřížky a dvoubarevný podkovový magnet. **Konsolidovaná v3.9** a **v3.10** zůstávají závaznými staršími podklady tam, kde je v3.11 výslovně nemění. Produkční Go/PDFium kód nezměněn.

**Dvoufunkční tlačítka:** společný šipka/textový kurzor a po stránkách/souvisle zobrazují ikonu příští akce; další kliknutí ji přepne. V produkční Go/Win32 aplikaci ještě není implementováno.

## Začít zde

1. [Postup spolupráce a předání Codexu](POSTUP_SPOLUPRACE_S_CODEXEM.md)
2. [Hlavní specifikace PDF Blacksmith v5.0](specifikace_Codex_v5.0.md)
3. [Plán rozšíření v2.0](plan_rozsireni_v2.0.md)
4. [Formuláře F1 – rozhodnutí F-Q1 až F-Q9](formulare/F1_rozhodnuti.md)
5. [Formuláře F1 – detailní specifikace v1.9](formulare/F1_specifikace_v1.9.md)
6. [Formuláře F1 – implementační postup](formulare/F1_plan_implementace.md)
7. **[Aktuální HTML prototyp v3.11](rozhrani/cely_vizual_v3.11.html)** — Separace, dočasné Výsledky vyhledávání, pevná a inteligentní vodítka, nastavení mřížky, nový magnet.
   - **[Pokyny pro Codex v3.11](rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md)** — poslední změny; přednost jen v rozsahu těchto změn.
   - **[Konsolidované pokyny pro Codex v3.9](rozhrani/AKTUALNI_UI_CODEX_v3.9.md)** — jednotná základna UI z v3.4–v3.9.
   - [Předchozí specifikace přichytávání v3.10](rozhrani/UI_PRICHYTAVANI_v3.10.md) a [HTML v3.10](rozhrani/cely_vizual_v3.10.html) pro historii.
   - [Předchozí HTML v3.9](rozhrani/cely_vizual_v3.9.html) a [Text v3.3](rozhrani/TEXT_VYBER_A_SMISENA_TYPOGRAFIE_v3.3.md) pro porovnání; dřívější soubory zůstávají v rozhrani/.

8. [Původní návrh Formulářů F1](formulare/F1_mockup_v1.0.html)
9. [Vyhledávání – budoucí upřesnění](vyhledavani_navrh_v0.1.md)
10. **[Aktuální mapa ikon v4.6](mapa_ikon_v4.6.md)** – [SVG atlas v3.10](rozhrani/ikony/pdf_blacksmith_nove_v3.10.svg).

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
