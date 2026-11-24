# PDF Blacksmith – pracovní UI v3.13: pravítka, pomůcky rozvržení, vlastní zkratky

**Datum:** 24. 11. 2026  
**Status:** schválené UX chování a HTML demonstrace. **NEPŘIKAZUJE** automaticky implementovat Go/Win32/PDFium.  
**Aktuální soubor:** [cely_vizual_v3.13.html](cely_vizual_v3.13.html)  
**Základní pravidla:** [UI v3.11](UI_SEPARACE_PANELY_VODITKA_v3.11.md), [specifikace v5.0](../specifikace_Codex_v5.0.md), [mapa zkratek v1.1](../mapa_klavesovych_zkratek_v1.1.md)

## 1. Pravítka patří vně stránky

- **Horizontální pravítko je nad horním okrajem listu, vertikální je vlevo od levého okraje**, nikoli nad tisknutelným PDF obsahem.
- Má mezi sebou a papírem vlastní oddělovací linii a při zoomu/scrollu zůstává navázáno na souřadnice stránky. Při kontinuálním zobrazení má každá stránka správně navázané pomocné měřítko, anebo se používá jediná pracovní stupnice se správným přepočtem; Codex musí rozhodnout podle existujícího vieweru. Při otočení, ořezu stránky, PageBox/UserUnit se souřadnice přepočítají bez systematického posunu.
- Pevná vodítka jsou plná, přesahují list směrem k pravítkům; dočasné chytré přichytávací čáry jsou čárkované/tečkované a vínové podle schválení v3.11.
- Pravítka, mřížka, vodítka a snap preview jsou výhradně **editorové překryvy**, neobsahují PDF objekty ani anotace, netisknou se a nejsou Undo/dirty událostí.
- HTML v3.13 přesunulo pravítka pomocí CSS do vnějšího okraje. Nejde o finální fyzikální měřítko na monitoru ani o hotové souřadnicové vykreslování PDF.

## 2. Nástroje → Zobrazení → Pomůcky rozvržení

- Levý panel Zobrazení má samostatnou položku \`layout_aids\` (stávající ikona \`i-toggle_guides\`), bez vytváření nové hlavní záložky.
- Klik otevře příslušný **pravý panel Vlastnosti** a zachová již zvolený nástroj; nabídka Zobrazení v horní liště i druhá lišta sdílejí stejný \`viewAids\` stav.
- Vlastnosti obsahují přepínače viditelnosti mřížky, pravítek, vodítek, globální magnet; pro **pevná vodítka** jejich barvu, krytí, zamknutí a souřadnice; pro **mřížku** hlavní rozteč 1–50 mm, počet dílků 1/2/4/5/10, počátek X/Y, barvu a krytí; pro **inteligentní snap** barvu dočasných čar, čárkovanou/tečkovanou kresbu, toleranci v obrazových pixelech a jednotlivé cíle.
- Vypnutí globálního magnetu nesmí odstranit nastavené cíle. Zobrazení mřížky, pravítek a vodítek je nezávislé na přichytávání. Změny nastavení v pravém panelu se promítnou v nabídce Zobrazení a naopak.
- Barevné vstupy se vizuálně synchronizují okamžitě v HTML; finální implementace potřebuje správný převod souřadnic z mm/PDF pt do zařízení při zoomu, DPI a rotaci.
- Zbytečné vysvětlující texty zůstanou schované do kontextové nápovědy; aktuální panel v3.13 používá krátké věcné popisky a poznámku o neimplementované geometrii.

## 3. Globální Nastavení → Klávesové zkratky → Vlastní zkratka

- Jednotlivé příkazy mají přehlednou **tabulku příkaz / aktivní kombinace / tlačítko Vlastní zkratka / reset**.
- Kliknutí na Vlastní zkratka spustí režim zachycení další kombinace. **Esc** zruší zachycení beze změny, **Backspace/Delete** obnoví výchozí přiřazení. Po úspěšném zadání se zkratka zobrazí u příkazu.
- Při shodné kombinaci v překrývajícím se kontextu fokusu nebo při kolizi s rezervovanou zkratkou se zobrazí čitelná výstraha, **přiřazení se neprovede**. Kontrolovat také systémové kombinace, českou klávesnici a AltGr (nevolit pro výchozí zkratky Ctrl+Alt).
- **OK / Použít** potvrdí změny pouze v paměti HTML prototypu; **Zrušit** zahodí provedené změny. Není to funkční přemapování Win32 kláves. Produkční aplikace potřebuje centrální registr kláves, persistenci preferencí, zobrazení zkratky v menu a testy konfliktů.
- **Nesmí se bez schválené migrace změnit dosavadní význam \`T/V/M/L\`, \`Ctrl+L/E/R/J\`, \`Ctrl+F\`, \`Ctrl+Z\`, \`Ctrl+Tab\` atd.** Nové zkratky pro magnet, mřížku, pravítka a vodítka zůstávají pouze návrhem uvedeným v mapě klávesových zkratek.

## 4. Ověření v3.13

- Chrome/Chromium: 980×540, 1100×720, 1250×820, 1600×900 CSS px; nový View podnástroj, panel Vlastností, změna rozteče mřížky, změna barvy vodítka; horizontální pravítko nad listem, vertikální vlevo od něj.
- Nastavení: přiřazení vlastní zkratky, odmítnutí kolize \`Ctrl+F\` pro mřížku, opětovné otevření po OK a zachování nastavení jen v HTML; žádné JavaScriptové runtime chyby ve zkoušených scénářích.
- Testy Go/Win32, výsledného PDF, OCR a skutečného geometrického SnapEngine **neproběhly**. HTML nesmí hlásit, že skutečnou PDF změnu dokončilo.

## 5. Priorita dokumentace a další implementační etapy

1. Schválené skutečné pravidla projektu \`specifikace_Codex_v5.0.md\` se zachovají.
2. Toto **v3.13 upřesňuje pouze rozložení pravítek, umístění Pomůcek rozvržení a vzhled/editaci zkratek**. Nezruší v3.9–v3.12 výběry, text, OCR, separace, nápovědu ani ikonovou mapu.
3. Codex před každým produkčním commitem nejdřív porovná nové chování s existujícími handlery, otestuje změny v samostatné etapě a vyžádá si výslovný pokyn.
