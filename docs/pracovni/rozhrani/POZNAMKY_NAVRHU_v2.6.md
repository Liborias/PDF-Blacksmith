# PDF Blacksmith – skupinová ouška panelů (HTML v2.6)

**Datum:** 22. 11. 2026. **Stav:** pracovní demonstrace HTML, nejde o implementaci ve Windows.

## Hlavní chování

- **Levá skupina:** hlavní panel `Nástroje`, pomocný `Náhledy stránek`.
- **Pravá skupina:** hlavní panel `Vlastnosti aktivního nástroje`, pomocné `Výsledky vyhledávání`, `Záložky`, `Poznámky`.
- U horního okraje hlavního panelu je malé **skupinové ouško** s dvojitými šipkami. Přesahuje do sousedního pomocného panelu. Obsah pomocných panelů má kvůli tomuto přesahu horní odstup 38 px a individuální ouška začínají níže.
- Směr šipek:
  - levá skupina otevřená `«` = sbalit doleva; zavřená `»` = rozbalit doprava;
  - pravá skupina otevřená `»` = sbalit doprava; zavřená `«` = rozbalit doleva.
- **Zachování stavu:** sbalení celé skupiny pouze skrývá její obsah; **nemění** individuální příznaky otevření/zavření náhledů, vlastností, záložek, poznámek a výsledků. Otevření tudíž vrací identický předchozí stav, včetně aktuálního nástroje a výběru.
- Jednotlivé pomocné panely mají pod skupinovým ouškem vlastní svislé ouško pro otevření/sbalení bez skrytí hlavního panelu.
- Při menší pracovní šířce (< 1900 px) se kvůli čitelnosti zobrazí současně nejvýše jeden pravý doplňkový panel; v extrémně širokém okně lze otevřít více. Oba hlavní panely tím nejsou dotčeny.
- Doplňkový panel **Výsledky** obsahuje seznam výskytů nezávisle na nastavení nástroje Najít v PDF. Aktivace hledání otevře jeho vlastní pomocný panel; tlačítko v nastaveních výsledky případně opět zobrazí.

## Testy HTML provedené v Chromium

1. Levá skupina: před složením skrytý náhled + otevřené nástroje → sbalit/rozbalit → zachovat stejnou kombinaci.
2. Pravá skupina: Poznámky + Vlastnosti → sbalit/rozbalit → obnovit oba panely.
3. Hledání → otevřít samostatný panel Výsledků; zadání dotazu aktualizuje jeho obsah.
4. Při sbalení obou stran se prohlížeč PDF rozšíří; po otevření se původní stav obnoví.
5. Pravá pomocná ouška přepínají Výsledky a Záložky bez ztráty Vlastností.

**Výsledek:** interakční testy prošly, v použitém Chromium bez chyb JavaScriptu. Jde o vizuální HTML simulaci; panelové chování je podkladem pro budoucí implementaci v Go/Win32, nikoli hotová implementace PDF editoru.

## Implementační upozornění pro Codex
Nedělat z HTML produkční frontend. Při přidávání zavíracích oušek zachovat původní funkce objektů, textu, separací, podpisů, ukládání, OCR, hledání i panelů; nepřepisovat historické schválené požadavky. Před úpravou zdrojů prověřit současnou správu panelů a uložených stavů, zvlášť duální zobrazení a obnovení rozložení.
