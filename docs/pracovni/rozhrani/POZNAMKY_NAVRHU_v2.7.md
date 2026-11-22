# PDF Blacksmith – pracovní HTML v2.7: Vložit strany a boční panely

**Datum:** 22. 11. 2026. **Stav:** ověřený interaktivní prototyp UI, nikoli implementace v aplikaci Go/Win32.

## Soubor → Vložit strany — uživatelem schválené rozvržení

V pravém panelu Vlastnosti jsou 3 samostatné sekce:

1. **Zdroj vkládaných stránek:** jiné PDF, nebo jiný dokument, který se před vložením převede stejným modulem jako samostatný příkaz Importovat.
2. **Výběr stránek:** všechny / vybrané v náhledech / vlastní rozsah `1, 3–5, 9` / pouze liché / pouze sudé.
3. **Umístění v aktivním PDF:** **Za aktuální stránku (výchozí schválená volba)**, Před aktuální, Na začátek, Na konec, Na zadanou pozici. Zobrazit náhled budoucího pořadí a počet nových/výsledných stran.

Obě větve operují nad **aktuálně otevřeným PDF**, nikdy nenahrazují originály. V HTML výběr souboru čte pouze název a počty zdrojových stran jsou testovací (12 pro PDF, 8 pro cizí dokument); **tlačítko Vložit strany neprovádí žádnou PDF operaci**. Konkrétní rozsah je validován včetně chybných a překračujících hodnot, mřížka čísel umožňuje výběr testovacích stran.

## Význam dalších funkcí Soubor sjednocen s dohodou

- **Otevřít:** pouze existující PDF.
- **Importovat:** převést externí dokument na **samostatné nové PDF otevřené v nové záložce**, nikoli vložit do aktuálního.
- **Sloučit a otevřít:** sestava **dvou a více** PDF i ne-PDF dokumentů, které se převedou a spojí v určeném pořadí do **nového PDF**.
- **Vložit objekt:** je součástí nástroje **Objekty**; vkládá grafiku na aktuální stránku, nikoli celé stránky.
- Požadovaná širší podpora DOCX/XLSX/PPTX, PSD/AI, OpenDocument, GIMP/Krita/Scribus a obrázků vyžaduje **technické ověření převodních enginů** před implementací.

## Ouška a posuvníky – dříve odložené úpravy, nyní v HTML

- **Odstraněny rezervované postranní lišty pro ouška pomocných panelů.** Ouška nyní vystupují přímo do prostoru dokumentu (absolutně umístěná na hraně skupiny), bez trvale zabraných 23 px boční kolejničky.
- Nadále platí **horní skupinové ouško na hraně hlavního panelu** vlevo i vpravo, se šipkami `«/»`; sbalení celé skupiny zachovává individuální stavy všech pomocných panelů.
- **Posuvníky v Náhledech, Nástrojích, Vlastnostech a pomocných panelech** mají tenké zaoblené běžce se středovou jemnou linkou, střídmý tmavý styl. Chromium použije přesný CSS návrh; na jiných systémech se může lišit kresba nativních posuvníků.
- Původní SVG kovadlina, dvě horní lišty, paleta a již fungující demo Formulářů zůstávají.

## Ověření

V Chromium testováno přepnutí Soubor → Vložit strany, schválená výchozí volba, volba PDF / jiný dokument, vlastní rozsah a validace, zařazení na konec, výběr konkrétních stránek, viditelný náhled a neaktivní tlačítko při chybě. Dále sbalení a obnovení obou skupin, zachování stavu Náhledů a funkční samostatná ouška. Testovány i nabídky Otevřít, Importovat, Sloučit a Objekty. **Bez chyb JavaScriptu.**

**Neimplementovat produkční převod jen podle HTML.** Nejprve prozkoumat existující Go/Win32, zvolit převodní knihovny, ověřit výstupní kvalitu, bezpečnost a práci s anotacemi, odkazy, AcroForm widgety a podpisy. Schválené rozvržení Vložit strany je podkladem pro samostatné implementační zadání.
