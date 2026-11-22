# PDF Blacksmith – HTML koncept v2.2: panely a ikony všech nástrojů

**Stav:** pracovní UI/UX prototyp, nejde o implementaci PDF editoru v Go/Win32. **Datum:** 22. 11. 2026.

## Aktualizace proti v2.1
- Dvě horní lišty, původní SVG kovadlina v záhlaví, tmavé neutrální schéma, nebarevné hranaté panely a jemné vínové akcenty beze změny.
- V levé hlavní liště mají všechny existující nástroje odpovídající ikony ze schválené SVG sady. **Formuláře nyní používají skutečné `nav_forms.svg`** namísto chybné reference.
- Pravý panel **Vlastnosti aktivního nástroje** ukazuje různé součásti podle volby Soubor, Zobrazení, Najít v PDF, Objekty, Text, OCR, Formuláře, Separace, Digitální podpis, Rozdělit PDF, Sloučit PDF, Pořadí stránek.
- **Záložky a Poznámky jsou samostatné doplňkové panely** napravo od dokumentu, obdobně jako Náhledy vlevo. Není možné, aby po otevření Záložek/Poznámek zmizel či se nahradil obsah Vlastností.
- Doplňkové panely se otevírají a sbalují vlastními drobnými vertikálními oušky; při dostatečné šířce lze zobrazit oba. Na užších oknech se kvůli prostoru zobrazí nejvýše jeden z dvojice.
- Poznámky už nejsou zdvojené jako samostatná položka hlavní lišty nástrojů; přístupné jsou přes doplňkový panel.
- Vedle existujících nástrojů je rozbalitelná sekce **Budoucí nástroje**, obsahující čistě pracovní návrhy Preflight, Porovnávání PDF, Měření a Dávky. **Tyto nástroje nesmějí být interpretovány jako schválené k implementaci.**

## Funkční obsah ilustrace
- Klikání na nástroje v hlavní liště a v menu Nástroje mění odpovídající vlastnosti na pravé straně.
- Záložky umějí v HTML přejít na jednu ze tří ukázkových stran; u jejich vlastností se pouze simuluje změna.
- Poznámky mají vlastní seznam a umožňují zadat **ukázkový lokální text poznámky**; nejde o PDF anotaci.
- Formuláře zachovávají přepínání stran, interaktivní pole a souhrn validace.
- Vyhledávání při psaní v panelu se týká jen demo textů HTML; u skutečného PDF zůstává okamžité hledání požadavkem k upřesnění po Formulářích.
- Funkce práce s objekty, Separace a ostatní obsahují konkrétní ilustrační rozvržení ovládacích prvků podle existujících funkcí, nikoli skutečné operace s PDF.
- Soubory jsou samostatné HTML bez externích závislostí; ikony jsou inline SVG, původní geometrie kovadliny zůstala zachovaná.

## Ověření
Otevřeno v Chromium; zkontrolováno přepnutí 12 existujících nástrojů a 4 plánovaných panelů, přechody mezi stránkami, Záložky/Poznámky nezávislé na Vlastnostech, nový komentář, seznam šablon a přítomnost ikon. Zaznamenáno **0 JavaScript chyb** při testu.

## Pokyn pro Codex
Nový vizuální návrh používej jako orientační vzor. Neměň automaticky existující nástroje ani jejich funkce. Nezaváděj F2/F3, nové vyhledávání ani plánované moduly jen proto, že je zde jejich panel naznačen. Detaily jako správné panelové řazení, souběžné zobrazení a responzivní chování implementuj v Go/Win32 podle schválených pravidel a po porovnání se současným kódem.
