# PDF Blacksmith – pracovní UI v3.16: jazyk rozhraní a kontextové vodítko

**Datum:** 24. listopadu 2026  
**Stav:** HTML demonstrace a schválená pravidla pro budoucí Codex; **žádná úprava produkční Go/Win32/PDFium aplikace**.  
**Aktuální HTML:** [cely_vizual_v3.16.html](cely_vizual_v3.16.html)  
**Navazuje:** [UI v3.15](UI_VODITKA_OPRAVY_v3.15.md), [UI v3.14](UI_VODITKA_INTERAKCE_v3.14.md), [hlavní specifikace v5.0](../specifikace_Codex_v5.0.md), [mapa ikon v4.9](../mapa_ikon_v4.9.md), [mapa zkratek v1.2](../mapa_klavesovych_zkratek_v1.2.md).

## 1. Nastavení → Jazyk

- V globálním Nastavení pod ozubeným kolečkem existuje **samostatná kategorie Jazyk** (ikona zeměkoule), nikoli duplicitní nezávislé výběry v Obecné a Jazyku.
- Pro první etapu jsou dostupné **Čeština / English**. Vybráním se **okamžitě zobrazí náhled** jazyka textů rozhraní; **Zrušit / zavřít bez potvrzení** obnoví předchozí uložený jazyk, **Použít / OK** jej potvrdí. Výběr lze v HTML uložit do prohlížečového \`localStorage\` pouze tam, kde ho prostředí dovolí, jinak zůstává v paměti. V produkční aplikaci ukládat do standardní uživatelské konfigurace.
- **Jazyk aplikace není jazyk OCR ani jazyk dokumentu**. Přepnutí rozhraní nikdy nepřepisuje text otevřeného PDF, metadata, soubory ani výsledky OCR; samostatná OCR předvolba je v **Text a OCR**.
- V HTML zaveden zdrojový CZ/EN slovník \`UI_TRANSLATIONS_EN\`, pomocná funkce \`localizeUI()\`, převod dynamicky generovaných popisků, titulky tlačítek a změny jazyka bez restartu. Zachovat stálé identifikátory příkazů, ID ikon, datové klíče a skutečné hodnoty formulářových prvků – překládat pouze text UI. Volby \`<option>\` musí mít pevné \`value\`, takže překlad popisku nezmění chování přepínače.
- **Rozsah v3.16:** významná část stávajících ovládacích prvků, panelů a nastavení má překlady. **Některé rozšířené popisné texty ve starších demonstračních panelech mohou nadále zůstat v češtině.** Je potřeba pokračovat až k úplnému pokrytí UI. V produkční implementaci **nepoužívat obecné plošné přepisování DOM textů**, které je pouze prototypový přechod; zavést stabilní identifikátory \`i18n\` a katalog CZ/EN, na jednom místě kontrolovat chybějící klíče a plurály.
- V budoucí aplikaci uvést jazyk pro každou novou funkci, tooltip, pomocný panel, přístupné \`aria-label\`, dialog a chybové hlášení. Umožnit fallback češtiny jen ve vývoji a v testech hlásit chybějící překlady. Přizpůsobit šířky panelů delším anglickým názvům.

## 2. Přechod do Pomůcek rozvržení při výběru

- **Kliknutí na existující vodítko** automaticky označí konkrétní linku a **dočasně** přepne pravý panel Vlastností na **Zobrazení → Pomůcky rozvržení**, i když byl předtím otevřen jiný nástroj (Text, Objekty, Soubor, OCR apod.). V HTML se uchová \`temporaryLayoutContext={tool, viewAction, source}\` a původní výběr PDF se kvůli samotnému panelu nemění.
- **Kliknutí mimo vodítka do volné pracovní plochy** odznačí vodítko/vícenásobný výběr, zavře jeho kontextové menu a **obnoví nastavení předchozího nástroje**. Pokud mezitím uživatel vědomě změnil nástroj, tato změna má přednost. Při kliknutí na obsah PDF může následná dosavadní logika automatického výběru textu/objektů přepnout správný editační nástroj; výsledné chování nesmí blokovat obvyklou práci s PDF.
- **Volba mřížky, pravítek, vodítek, magnetu a snap cílů** na druhé horní liště či v nabídce Zobrazení aktivuje odpovídající dočasný panel Pomůcek. Kliknutí mimo vodítko/na volnou pracovní plochu jej může vrátit. Naproti tomu **výslovný výběr nástroje Zobrazení → Pomůcky rozvržení** je trvalý: pouhé kliknutí mimo vodítko jej nezavírá.
- Nesmí vzniknout další paralelní pravý panel. Nepřepisovat stav ručně otevřených samostatných Pomocných panelů (Záložky, Poznámky, Výsledky atd.); dočasně se mění pouze nastavení ve společné části **Vlastnosti**.
- Zachovat pravidla výběru a kláves z minulých verzí a nerozbít příkazy na stránce: aktivní přesun vodítka má přednost před selekcí objektů, ale klik do textu dále smí aktivovat Text.

## 3. Kontextová nabídka vodítka

- **Na hover žádná automatická lišta nevyskakuje.** Kurzor nad volným vodítkem ukazuje směr jeho pohybu: vodorovné \`ns-resize\`, svislé \`ew-resize\`. Při zamknutí se přesun zakáže.
- **Levé kliknutí do vodítka pouze vybere vodítko**, volitelně s rozšířeným výběrem \`Ctrl+klik\`. Uživatel může následně tažením přesunout neuzamčené vodítko podle v3.15.
- **Pravé kliknutí na vodítko** vyvolá malé menu v místě kurzoru, pouze **barva (color picker), zámek/odemknout (SVG ikona), odstranit (SVG ikona koše)**. **Neobsahuje položku Vybrat.** Popis zámečku a koše je pouze v \`title\` / \`aria-label\`, běžně při najetí.
- Respektovat globální i individuální zámky; mazání nesmí odstranit uzamčené vodítko. Barva z menu upravuje **pouze vybrané konkrétní vodítko**, nikoliv výchozí barvu nových nebo všechny ostatní.
- Zavření menu kliknutím jinam, po dokončení operace či přepnutí nástroje; nástroje zůstávají vykreslené na stránce, standardně neexportované.
- **Pokročilé symetrické zrcadlení** stále pouze v plánu na později; do menu je zatím nepřidávat.

## 4. Klávesy a regresní bezpečnost

- **Žádná nová globální zkratka** pro jazyk nebo pravé menu se nepřidává. \`Esc\` nadále ruší aktivní tažení; existující \`Delete\` maže vodítko jen v příslušném kontextu, jinak původní PDF obsah. \`Ctrl+F\`, \`T/V/M/L\`, \`Ctrl+L/E/R/J\` i ostatní [zkratky v1.2](../mapa_klavesovych_zkratek_v1.2.md) beze změny.
- Kontrola kolizí vlastních zkratek v Nastavení je povinná a nesmí být oslabena překladem UI.
- Produkční jazykový modul a trvalé vodítkové nástroje budou implementovány až po výslovném schválení, inventuře existujícího kódu a návrhu izolované etapy.

## 5. Ověření HTML v3.16

- Prohlížeč Chromium, viewporty **980×540, 1100×720, 1250×820, 1600×900**: po kliknutí na vodítko otevřít Pomůcky rozvržení, obnovit Text po kliknutí mimo; kontextové menu po pravém kliknutí má 2 SVG ikony + jediný color picker, po najetí nevyskakuje.
- V doplňkovém testu: ze záložky **Objekty** zapnout mřížku na horní liště, vidět Pomůcky rozvržení a po kliknutí do volné pracovní plochy vrátit Objekty; výslovně zvolený podnástroj Pomůcky zůstává otevřený.
- Pravý klik → koš: odstraní neuzamčené vodítko; změny vlastního textu PDF nejsou provedeny.
- Nastavení → Jazyk: přepnutí do angličtiny mění popisky (včetně nastavení), **Cancel vrací češtinu**, OK ponechá angličtinu, pozdější volba češtiny funguje; obsah PDF se nezmění a OCR jazyk zůstává nezávislý.
- V provedených testech bez neodchycených JavaScriptových chyb. **Neproběhl test produkčního Go/Win32, exportu PDF ani úplný audit všech překladových řetězců.**

## 6. Následující etapa

Pokračovat v úplném auditu překladových klíčů, nápověd a stavových hlášení; současně postupně připravit matematický SnapEngine nad skutečnými PDF souřadnicemi, odlišný od HTML demonstrace. Původní předpisy v3.9 až v3.15 se mění pouze tam, kde výslovně rozhoduje tento dodatek.
