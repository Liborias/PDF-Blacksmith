# PDF Blacksmith – pracovní podklady pro Codex

**Aktualizováno:** 24. listopadu 2026 · **Stav:** návrh rozhraní F1, implementace F1 schválena po etapách; ostatní položky jen plán.

Tato složka je centrální rozcestník průběžné dokumentace. **Nenahrazuje stávající implementaci** v Go/Win32 a Pythonu; přítomnost návrhu neznamená, že byl implementován.

## Nejnovější stav rozhraní (24. listopadu 2026)

**v3.24:** [HTML](rozhrani/cely_vizual_v3.24.html) · [pokyny pro Codex](rozhrani/UI_BARVY_VODITEK_v3.24.md). Podklady barevných polí vodítek, skupin a místní nabídky respektují světlý/tmavý motiv; vybraná barva vzorku zůstává nezměněna. v3.23 ponechána pro srovnání.

**v3.23:** [HTML](rozhrani/cely_vizual_v3.23.html) · [pokyny pro Codex](rozhrani/UI_HLAVNI_MENU_TEXTOVE_v3.23.md). Položky Soubor, Zobrazení, Nástroje, Úpravy, Formuláře a Pomoc ve druhé liště jsou opět pouze textové bez viditelných tlačítkových rámečků; hover a rozbalené nabídky zůstávají funkční. Ostatní v3.22 beze změny.

**v3.22:** [HTML](rozhrani/cely_vizual_v3.22.html), [upřesnění Codex](rozhrani/UI_EDITOR_TOOLTIP_POZNAMKY_REZIM_v3.22.md), [barevná mapa v1.3](mapa_barev_v1.3.md) a [JSON](mapa_barev_v1.3.json). Paletová plocha kolem PDF, světlé tooltipy, štítky poznámek dle motivu s vlastní barvou, půlený kruhový přepínač nahoře, jemně zmenšené logo a ikony. Starší verze zachovat.

**v3.21:** [HTML](rozhrani/cely_vizual_v3.21.html), [pokyny pro Codex](rozhrani/UI_AUDIT_BAREV_OCR_MINIATURY_v3.21.md), [mapa barev v1.2](mapa_barev_v1.2.md) a [JSON](mapa_barev_v1.2.json). Audit vínových rámečků miniatur, tmavých polí výběru barev, OCR a dalších starých CSS dekorací. Starší verze neměnit; Go/PDFium stále beze změny.

**v3.20:** [HTML](rozhrani/cely_vizual_v3.20.html), [pravidla pro Codex](rozhrani/UI_SVETLE_MOTIVY_SCROLLBARY_KONTRAST_v3.20.md), [barevná mapa 1.1](mapa_barev_v1.1.md) a [JSON](mapa_barev_v1.1.json). Světlé ovládací plochy, kontrast, světlé posuvníky a stejný hover oušek pomocných panelů vlevo/vpravo a obou horních oušek. Předchozí v3.19 zůstává zachována; produkční kód se nemění.

**Aktuální návrh:** **v3.24** ([HTML](rozhrani/cely_vizual_v3.24.html), [pokyny pro Codex v3.19](rozhrani/UI_MOTIVY_KONTRAST_PREKLADY_v3.19.md), [mapa barev a kontrastu](mapa_barev_v1.0.md) + [JSON](mapa_barev_v1.0.json), [překladový katalog v1.1](i18n/ui_catalog_v1.1.json)). V3.19 obsahuje šest kompletních světlých i tmavých palet (šedo-vínová, celošedá, krémová, modrá, olivová a petrolejová), světlá ouška, tlačítka a vstupy ve světlých motivech, opravu kontrastů a překlady dynamických formulářových upozornění. Ozubené kolečko je výškově sjednocené s disketou. **Zůstává interaktivním HTML prototypem, nikoli implementací Go/PDFium**. Starší schválená pravidla a historie se nemění.

**Dvoufunkční tlačítka:** společný šipka/textový kurzor a po stránkách/souvisle zobrazují ikonu příští akce; další kliknutí ji přepne. V produkční Go/Win32 aplikaci ještě není implementováno.

## Začít zde

1. [Postup spolupráce a předání Codexu](POSTUP_SPOLUPRACE_S_CODEXEM.md)
2. [Hlavní specifikace PDF Blacksmith v5.0](specifikace_Codex_v5.0.md)
3. [Plán rozšíření v2.0](plan_rozsireni_v2.0.md)
4. [Formuláře F1 – rozhodnutí F-Q1 až F-Q9](formulare/F1_rozhodnuti.md)
5. [Formuláře F1 – detailní specifikace v1.9](formulare/F1_specifikace_v1.9.md)
6. [Formuláře F1 – implementační postup](formulare/F1_plan_implementace.md)
7. **[Předchozí HTML v3.19 – plné barevné palety, kontrast a dynamické překlady](rozhrani/cely_vizual_v3.19.html)**
   - **[Codex v3.19: kompletní témata, světlé ovládací prvky, překlady](rozhrani/UI_MOTIVY_KONTRAST_PREKLADY_v3.19.md)**.
   - **[Mapa barev a pravidel kontrastu v1.0](mapa_barev_v1.0.md)** – [strojová mapa JSON](mapa_barev_v1.0.json).
   - [Aktualizované překlady JSON v1.1](i18n/ui_catalog_v1.1.json), [CSV v1.1](i18n/ui_catalog_v1.1.csv) – 809 položek.
   - [Předchozí HTML v3.18](rozhrani/cely_vizual_v3.18.html) – skupiny, barevné editace a ořez.
   - **[Nové pokyny pro Codex v3.18](rozhrani/UI_TEMATA_PREKLADY_VODITKA_OREZ_v3.18.md)** – poslední závazné upřesnění UI.
   - [Překlady: JSON pro Codex](i18n/ui_catalog_v1.1.json), [CSV tabulka](i18n/ui_catalog_v1.1.csv), [postup](i18n/README.md).
   - [Předchozí HTML v3.17](rozhrani/cely_vizual_v3.17.html) – skupiny a dokončení základních překladů.
   - **[Nejnovější zadání pro Codex v3.17](rozhrani/UI_PREKLADY_SKUPINY_VODITEK_v3.17.md)** – výběr, skupinové operace a rozšířené cíle přichytávání.
   - [Předchozí HTML v3.16](rozhrani/cely_vizual_v3.16.html) – jazyk a kontextové vodítko. – pracovní prototyp, nikoli hotové funkce PDF.
   - **[Nový dodatek Codex v3.16: jazyk, dočasné Vlastnosti a kontextové menu](rozhrani/UI_JAZYK_VODITKA_v3.16.md)**.
   - [Předchozí HTML v3.15](rozhrani/cely_vizual_v3.15.html) – interaktivní vodítka, barvy a přichytávání.
   - **[Pokyny pro Codex v3.15](rozhrani/UI_VODITKA_OPRAVY_v3.15.md)** – nové směry pravítek, odstranění konfliktů, další plán jazyka.
   - [Předchozí HTML v3.14](rozhrani/cely_vizual_v3.14.html) – původní interaktivní vodítka.
   - **[Nové pokyny pro Codex v3.14](rozhrani/UI_VODITKA_INTERAKCE_v3.14.md)** – jen interakce, geometrii PDF nepřepisovat.
   - [Předchozí HTML v3.13](rozhrani/cely_vizual_v3.13.html) – vzhled pravítek, Pomůcky rozvržení a vlastní zkratky.
   - **[Nové pokyny pro Codex v3.13](rozhrani/UI_POMUCKY_ZKRATKY_v3.13.md)** – pouze poslední schválené změny.
   - [Předchozí HTML v3.12](rozhrani/cely_vizual_v3.12.html) – autorský magnet s currentColor.
   - **[V3.11: Separace, pomocné panely, vodítka](rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md)**.
   - **[Předchozí HTML v3.12: autorský magnet + všechna pravidla v3.11](rozhrani/cely_vizual_v3.12.html)** – pracovní prototyp, nemění skutečné PDF.
   - [Předchozí HTML v3.11](rozhrani/cely_vizual_v3.11.html) pro porovnání.
   - **[Pokyny pro Codex – UI v3.11](rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md)** – nové závazné upřesnění.
   - **[Mapa klávesových zkratek v1.3](mapa_klavesovych_zkratek_v1.3.md)** – existující Go vs. návrhy a kolize.
   - [Předchozí HTML v3.10](rozhrani/cely_vizual_v3.10.html) pro historii.
   - **[Nové pokyny pro Codex v3.10: přesné změny přichytávání](rozhrani/UI_PRICHYTAVANI_v3.10.md)** – mají přednost před v3.9 jen v popsaných třech změnách.
   - [Předchozí HTML v3.9: univerzální výběr, OCR, obrázky a dvoustavové zobrazení](rozhrani/cely_vizual_v3.9.html) pro porovnání.
   - **[Aktuální konsolidované pokyny pro Codex v3.9 – všechny změny v3.4 až v3.9](rozhrani/AKTUALNI_UI_CODEX_v3.9.md)** – hlavní zdroj aktuálních pravidel UI a jejich priorit.
   - [Předchozí HTML v3.3 pro porovnání](rozhrani/cely_vizual_text_vyber_typografie_v3.3.html).
   - [Předchozí pokyny pro Codex v3.3: bezkolizní výběr a smíšené formátování](rozhrani/TEXT_VYBER_A_SMISENA_TYPOGRAFIE_v3.3.md) – dřívější upřesnění, zachovat původní pravidla SEL/TXT/OBJ.
   - [Předchozí HTML v3.2: horní řada a nápověda](rozhrani/cely_vizual_text_vyber_napoveda_v3.2.html) pro porovnání.
   - **[Pokyny pro Codex – rozmístění výběrových ikon a nápovědy v3.2](rozhrani/TEXT_VYBER_NAPOVEDA_v3.2.md)** – včetně ochrany schválených gest.
   - [Předchozí HTML v3.1: záložky Výběr/Typografie](rozhrani/cely_vizual_text_zalozky_v3.1.html) pro porovnání.
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
