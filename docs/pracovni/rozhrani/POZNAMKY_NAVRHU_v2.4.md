# PDF Blacksmith – doladění Soubor, Zobrazení a horního vyhledávání (HTML v2.4)

**Datum:** 22. 11. 2026.  
**Stav:** vizuální a interaktivní HTML prototyp, ne aktuální funkčnost programu v Go/Win32.

## Úpravy vůči verzi 2.3

1. **Soubor i Zobrazení** mají samostatnou rozbalovací nabídku v druhé horní liště, otevíranou najetím/kliknutím. Stejné položky se zobrazují jako odsazená podmenu v levém panelu Nástroje, generovaná z jediného pole `FILE_ACTIONS` / `VIEW_ACTIONS`.
2. **Kliknutí na hlavní položku Soubor nebo Zobrazení** vlevo přepíná rozbalení jejího podmenu. **Přechod na jiný nástroj** podmenu automaticky zavře. Zvolená podakce se uchovává pro pozdější návrat.
3. Pravý panel Soubor/Zobrazení zobrazuje jediný nadpis **Vlastnosti**; nepřidává druhý velký titul „Soubor – Vlastnosti“. Obsah se mění podle vybrané podakce.
4. Ze druhé horní lišty odstraněno **Posun dokumentu**: používalo stejný symbol kurzoru jako „Kurzor a výběr“ a v HTML nedělalo samostatnou akci. **Produkční posouvání kolečkem/mezerníkem je třeba zachovat.**
5. Do druhé lišty vloženo kompaktní **pole Najít v PDF s lupou**. Jeho aktivace přepne na nástroj **Najít v PDF** a otevře pravý panel, i když byl sbalený. `Ctrl+F` vybere horní hledací pole. Zadaný text se synchronizuje oběma směry s hledacím polem v pravém panelu.
6. Pro Zobrazení přibyla levá nabídka: Po stránkách, Souvislé zobrazení, Přizpůsobit stránce, Přizpůsobit šířce, 1:1 (100 %), Dvě PDF vedle sebe, Celá obrazovka, Panely a náhledy, Kvalita zobrazení. Vpravo má každá vlastní nastavení; **nové ovládací prvky jsou pracovní návrh** a nesmí nahradit stávající funkce PDFium.

## Ověřeno v HTML prototypu

- Dvojklik na Soubor / Zobrazení zavře jejich odsazené podmenu, výběr jiného nástroje je také zavře.
- Obě horní nabídky otevřené na hover, jejich příkazy volí shodně zvýrazněnou položku nalevo a kontextové nastavení vpravo.
- Soubor → Vlastnosti má jen jeden hlavní nadpis Vlastnosti a žádný duplicitní titul.
- Kliknutí do horního vyhledávání i `Ctrl+F` zpřístupní Najít v PDF; text se synchronizuje oběma směry.
- Zobrazení 1:1 nastavuje 100 % v HTML ilustraci.
- Testováno v Chromium na širokém i menším viewportu bez chyb JavaScriptu.

**Záměrné omezení:** prohledávají se pouze ukázkové texty v HTML, nikoli reálné PDF; implementace stávajícího fulltextu a budoucích vyhledávacích šipek podléhá zvlášť schválenému postupu. Nastavení většiny dalších režimů jsou ilustrační. Zdrojový Go/Win32 kód zůstává nezměněn.
