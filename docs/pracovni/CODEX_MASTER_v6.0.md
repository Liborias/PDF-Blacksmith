# PDF Blacksmith — JEDNOTNÁ ZÁVAZNÁ SPECIFIKACE v6.0

**Datum konsolidace:** 25. listopadu 2026
**Repozitář:** Liborias/PDF-Blacksmith, hlavní větev main
**Stav:** jedna normativní verze pro další práci Codexu. Vychází z uživatelských rozhodnutí a doložených pracovních dokumentů po UI v3.26.
**Referenční HTML:** rozhrani/cely_vizual_v3.26.html
**Související jediné platné mapy:** mapa_ikon_FINAL_v6.0.md + JSON, mapa_barev_FINAL_v6.0.md + JSON, mapa_zkratek_FINAL_v6.0.md; lokalizace i18n/ui_catalog_v1.1.json (809 hesel).

> **PŘEDNOST A PRIORITA:** Nejprve implementovat a doladit **rozhraní stávající desktopové aplikace** a jeho základní interakce; **nezahajovat nyní formuláře F1**. Jejich dřívější schválení rozsahu zůstává v platnosti, ale **pořadí realizace** mění explicitní rozhodnutí uživatele z 25. 11. 2026. F2/automatický datový import jsou návrhové, nikoli povolená produkční implementace. Codex před každou etapou předloží analýzu a vyžádá si konkrétní souhlas.

## 0. Autorita tohoto dokumentu a hranice

Tento text **nahrazuje veškeré předchozí kumulativní pracovní instrukce pro účely řízení další implementace** (včetně hlavní specifikace v5.0, plánu v2.0, iniciačních promptů a dodatků UI v3.1–v3.26), pouze pokud jde o **výklad, pořadí, rozhraní a řešení kolizí**. Staré soubory jsou **historický důkaz rozhodnutí**, nikoli samostatné konkurenční zadání. Codex je kvůli běžné implementaci **nemá procházet**. Vznikne-li skutečná nová nejasnost, předloží ji s návrhem řešení; nevybere si svévolně starší variantu.

**Nepřekonatelná omezení:** zachovat existující Go/Win32 a Python/PDFium architekturu i funkce; nevymazat žádné původní akce, upravovat pouze schválenou malou etapu; nepoškodit původní PDF, podpisy, odkazy, anotace ani formulářové widgety. Produkt není webová aplikace: HTML je **vzor vzhledu a interakcí**, nikoli zdrojová implementace a nikoli důkaz funkčnosti PDF.

**Povinná klasifikace pro každou funkci:**
- **A — produkčně ověřeno:** skutečně nalezený handler, konkrétní soubor a test Windows.
- **B — přítomno v kódu, runtime neověřeno:** staticky doložený existující kód, jen částečně funkční.
- **U — schválené chování UI:** zobrazit v produkčním návrhu až v příslušné schválené etapě.
- **D — pouze demonstrace HTML:** klikací simulace bez zásahu do PDF (například F2, batch, SnapEngine).
- **P — plán bez oprávnění implementovat:** budoucí funkce.
Jedině konkrétní kontrola skutečného repozitáře může potvrdit A/B. Status v HTML nikdy nepovyšuje D na A.

## 1. Schválené pořadí další práce (nové, závazné)

**Etapa 0 — pouze inventura a bezpečné převzetí:**
1. Ověřit větev, pracovní strom, změny vytvořené Codexem (zejména rozpracované F1), commity, existující testy a build; bez resetu či smazání rozpracovaného kódu.
2. Prověřit hlavní Go/Win32 UI: workspace_windows.go, viewer_windows.go, document_commands_windows.go, menu, docking, text/objekty, historie, ukládání/export, DLL a PDFium. Přesný seznam doplnit podle skutečných souborů.
3. Předložit rozdílový audit: co v aplikaci běží, co je pouze navrženo v HTML v3.26, co je rozpracované, kde jsou zásadní rizika; navrhnout první **malou UI etapu** a její testy. Zatím žádné implementace ani kommit.

**Etapa 1 — stabilní kostra a styl desktopového rozhraní:**
- Identita okna, horní dva pásy, stávající příkazy a jejich ID, hlavní panel Nástroje, panel Vlastnosti, miniatury a pomocné panely, záložky dokumentů, stavový řádek, responzivní šířky, DPI, minimalizace/focus.
- Jednotné SVG ikony podle finální mapy, zachovat původní uživatelské logo kovadliny; karty, panelové ouška, scrollbary, nápovědy a aktivní stavy.
- Jediný registr barev, šest palet × světlý/tmavý režim; lokální jazyk CZ/EN; žádné šedé/tmavé „relikty“, vínové obrysy v nevínových paletách.
- V této fázi pouze shell/chrome, **neměnit PDF geometrii/ukládání ani výběrové handlery**.

**Etapa 2 — rozložení panelů a nabídky:**
- Kontextové pravé Vlastnosti, logické sekce s malým sbalovacím ouškem, bez prázdných vysokých nápovědných boxů. Samostatná pomocná pozice, řádné sbalení obou docků a návrat jejich stavu.
- Soubor/Zobrazení: struktura příkazů bez ztráty starých funkcí; menu v druhé liště pouze textové. Levý seznam nástrojů a vnořené menu bez obdélníkových rámů; aktivní ikonové tlačítko druhé lišty naopak kontrastně zvýraznit. Všechno otestovat na různých šířkách/DPI.
- Otevřít/Importovat/Vložit/Sloučit nepřejmenovat na synonymní operace; viz §6.

**Etapa 3 — navigace stran a měřítko:**
- Stránkový, souvislý a nový pružný přehled; sdílené miniatury; správné Ctrl+kolečko v pracovní ploše vs miniaturách; přizpůsobit stránce/šířce/100 %, zachovat stabilní UI při změně rozměrů.
- Zrušit **pouze dekorativní rám kolem celého PDF listu** (nikoli skutečné obrysy PDF/Widgetů). Na hraně stránky obyčejné kolečko posouvá na další/předchozí list.
- Pružný přehled má své vlastní zobrazení v menu, s flex-wrap, klik/Ctrl+klik/Shift+klik a Enter=vybraný list na výšku; pořadí a dávkové akce návrhově sdílejí datový model miniatur. Začít čistou bezpečnou navigací; skutečný přesun stran až po etapě s Undo/Redo.
- Původně označeno „na konec pokynů“: **v rámci této etapy oddělit jednoduchý zoom/navigaci od pokročilého pružného přehledu a hromadných operací**; implementovat jen po schválení dílčích kroků. F2 tím nepřeskakuje UI.

**Etapa 4 — jednotný výběr, pracovní pomůcky a regresní ochrana:**
- Až po funkčním UI shellu řešit přenesení univerzálního selection dispatcheru, kurzoru, obdélníku, lasa, miniatur a podpůrných gest bez konkurenčních handlerů. Samostatně testovat režimy Text/Objekty/OCR.
- Pravítka, mřížka, vodítka, skupiny, barvy, SnapEngine: nejprve overlay a uživatelské preference; **geometrický SnapEngine je samostatná riziková etapa** se zvoleným testovacím plánem.
- Zachovat výběr na protnutí/zásah, výběr textových úseků, smíšené formátování, obnovu panelů a Undo/Redo; nic nedělat destruktivně.

**Etapa 5 — Formuláře F1** (rozsah již dříve schválen, **termín nyní odložen za UI**):
- Inventura modelu a anotací, detekce AcroForm, aktivní vyplňování pouze ve Formuláře → Vyplnit, bezpečné ukládání a kontrola, kompatibilita v dalších prohlížečích. Jednotlivé kroky F1-0 až F1-5; každý zvlášť schválit.
- Pokud Codex mezitím vytvořil změny ve formulářích, **neodstraňovat je automaticky**, sepsat jejich stav a nechat bez dalších úprav do rozhodnutí uživatele.

**Etapa 6 — F2 a datové dávky (pouze návrh):**
- Produkční tvorba/editace polí, podpisových widgetů, vlastností, pořadí Tab a návrhu formuláře **zatím neschválena**. HTML v3.26 obsahuje interaktivní demonstraci.
- CSV/XLSX/Google Sheets, mapování, náhled a „1 řádek → 1 výsledné PDF“ zatím návrh. Produkční parser, OAuth, batch export, FDF/XFDF a F3 flatten pouze po samostatném schválení.

**Etapa 7 — ostatní samostatné moduly (návrhy):** hlubší textová rekonstrukce z křivek/OCR, náročný SnapEngine, Preflight, grafické porovnání, měření, dávky, asistovaná importní konverze a další podle samostatných priorit. Neodvozovat oprávnění z ikony nebo HTML.

## 2. Hlavní okno a vrchní lišty

- Dvě odlišné řady: **první horní pruh** s autorským logem, dokumentovými kartami, systémovými operacemi; **druhá lišta** začíná prostým textovým menu **Soubor · Zobrazení · Nástroje · Úpravy · Formuláře · Pomoc**. Trvalý obdélníkový rám položek hlavního menu je **zrušen**. Jemný hover/otevřený stav bez rámu.
- Ikonová část druhé lišty má kontrastní aktivní stav podkladem, akcentním obrysem a spodní linkou; odlišit idle/hover/active/focus/disabled. Při nedostatku šířky dynamicky skrývat skutečně vedlejší ikony, nikoli celou hlavní textovou nabídku a klíčové akce.
- Jedno **dvoustavové tlačítko šipka ↔ textový kurzor**, ikona ukazuje **příští akci**; jeho klik stav přepne, sám nezasahuje do PDF. Obdélník a laso stojí vedle samostatně. Textový blok pro odstavcový text je pouze v Text → Výběr, není univerzální výběr. Bodový text tvoří textový kurzor na volné ploše po kliknutí.
- Jedno dvoustavové tlačítko **po stránkách ↔ souvisle**, ikona ukazuje další volbu, stav se synchronizuje s menu a spodní lištou. **Pružný přehled stran je třetí samostatný režim/tlačítko**, nikoli třetí stav cyklické dvoustavové ikony.
- Horní souborové akce: uložit (disketa), nastavení (opticky ~18 px ozubené kolečko), nápověda (znak ? bez vnitřního kruhu, stejně vysoký jako sousední ikony), půlený kruh pro okamžitou změnu light/dark propojený s Nastavení → Rozhraní. Horní logo kovadliny o cca 1 mm menší než původní předchozí verze.
- Jazyková lokalizace CZ/EN překládá **jen text UI**, ne obsah PDF, uživatelská data, názvy souborů, OCR jazyk. Identifikátory, data-* a uložené hodnoty se nemění. Jazyk v samostatné záložce Nastavení → Jazyk, okamžitý náhled a Zrušit vrací uloženou hodnotu.

## 3. Panelové rozvržení, nápověda a barvy

- Hlavní levý panel **Nástroje** a pravý **Vlastnosti aktivního nástroje** mají stabilní šířku, obě skupiny reagují na resize a min. ~980×540 CSS px v HTML; vlastní Win32 DPI 100/150/200 %.
- Položky v levém panelu a rozbalovacích navigačních menu **bez obdélníkového rámečku, stínu a trvalé výplně**. Hover/aktivní: jemný paletový podklad + úzká levá akcentní linka. Oproti tomu běžná formulářová tlačítka mohou mít rám a aktivní ikonová tlačítka druhé lišty musí být výrazně označena.
- Pravé Vlastnosti: kompaktní sekce s oddělovací čárou a subtilním tlačítkem sbalit/rozbalit; neodstraňovat nebo neresetovat ovládání při sbalení. Žádné výplňové 59px/64px boxy bez textu.
- Dlouhé metodické vysvětlivky schovat pod **kruhový otazník** u nadpisu; bez obdélníkového rámečku. Hover/focus nad ikonou může vyvolat krátké dynamické vysvětlení, ale prázdný slot po opuštění **nezůstává přehnaně vysoký**. Skutečná varování a stavové chyby musejí zůstat viditelné.
- Levá skupina: Nástroje + Náhledy. Pravá skupina: Vlastnosti + jeden společný prostor pro pomocné **Výsledky/Záložky/Poznámky/Historie**. Skupinová ouška sbalí/obnoví skupinu do jejího předchozího stavu; hlavní panely mají ouško u vlastní hrany, pomocné ouška navazují vně pracovního listu, nedotýkají se PDF. Docky nesmí překrývat dokument; jejich rozměr se započítá do režimu fit. Obě horní a levá/pravá pomocná ouška mají jednotný, jemný hover pro světlé i tmavé.
- **Automatické pomocné panely nejsou všeobecným příkazem otvírat úplně všechno.** Vyhledávání automaticky a dočasně otevírá **Výsledky** a při opuštění obnoví přesně původní pomocný panel/sbalení, pokud mezitím uživatel ručně nezvolil jiné nastavení. Výběr vodítka dočasně otevře **Pomůcky rozvržení ve Vlastnostech**, klik mimo vrátí předchozí nástroj, pokud nebyl výslovně přepnut. Ostatní starší plošná automatika je odložena; uživatelská manuální volba má přednost.
- Scrollbary v každém z 12 motivů používají vlastní dráhu/thumb/hover, ve světlém motivu **žádná tmavá reliktní plocha**. Vhodné zaoblení a tenká středová kolejnička jako schválený vizuální směr; bez ztráty čitelnosti.

## 4. Barevný a ikonový systém

- Šest **skutečných samostatných palet**: šedo-vínová (výchozí), celošedá, krémová, mořská modrá, olivově zelená a petrolejová, každá light/dark. Vlastní akcent či plocha panelů je samostatná potvrzovaná volba, nikoli náhrada celé sady palet. Pouze šedo-vínová smí nést vínový akcent; **žádná hardcoded burgundy** v ostatních. Vypnutý/aktivní i hover stav včetně ikon a tooltips musí respektovat paletu.
- Tokeny UI NESMÍ přebarvovat obsah PDF: obrázky, CMYK/spot vzorky, skutečné pole PDF, uživatelem určené barvy vodítek/štítků, výsledné ICC, tiskové desky. Barvení vlastního UI smí zasáhnout **pouze ovládací prvky/editorové okolí**, ne stránkový obsah.
- Editorová plocha okolo stránky je **odvozena z palety**, nikoli pevně černá. Všechny PDF listy jsou v pracovní ploše **bez dekorativního vínového vnějšího rámečku** ve všech režimech; oddělit je stínem/kontrastem plochy. NEMAZAT rámečky přímo uvnitř PDF, widgetů, aktivního objektového výběru nebo anotací.
- Barevná políčka vodítek v light režimu mají světlý **rám/podklad**, ale skutečný barevný vzorek zůstává autentický, stejně jako vlastnosti nativního výběru barvy. Číselné/color inputy, OCR, textové boxy, poznámky, štítek Připnuté a vybrané miniatury plně respektují paletu.
- Minimální kontrast běžného textu 4,5:1, relevantních grafických hran/focus 3:1; testovat **computed style a screenshoty** pro 6×2 motivů, nejen předpočítané hodnoty. Mapa barvy FINAL obsahuje přesné tokeny.
- Ikony: jeden SVG atlas/registry s opakovaně použitelnými symboly 24×24 a currentColor, vyjma přesné dodané autorské geometrie (logo a podkova magnetu). Ikony nesmějí samy o sobě zakládat příkaz Win32. Rozhraní v Nastavení má **i-settings_interface** (samostatný existující vektor), jazyk ikonu zeměkoule, vodítka vlastní oko/zámek/koš/skupiny. Nepoužívat chybějící i-panel_left.
- Podrobné normativní ID a kontrolní soupis: **mapa_ikon_FINAL_v6.0.md**. Nečíst zastaralé „aktuální“ odkazy z map v4.7 nebo v5.1.

## 5. Dokumenty, stránky a měřítko

- **Po stránkách:** normální kolečko roluje na stránce, na jejím spodním/horním okraji přejde na další/předchozí stránku (jednou na gesto), bez reakce během editace inputu. Souvisle: celé stránky s plynulým posuvníkem. Fit page, fit width a 100 %: přepočítat skutečné PDF pixely z PDFium a DPI, nepřibližovat hotový bitmapový obrázek jako finální render.
- **Ctrl + kolečko v pracovní ploše** řídí přiblížení PDF v lokálním editoru, v panelu **miniatur** mění **jen jejich velikost**; mimo tyto oblasti běžný prohlížeč/OS nadále smí použít své zkratky. HTML musí používat scoped nepasivní wheel listener/preventDefault, ale prohlížeč/OS může mít výjimky: **žádná garance bez testu Chrome/Edge**. Ve Win32 zavést odpovídající lokální zpracování bez globálního hooku.
- Přiblížení podle kurzoru přibližně zachová bod pod ukazatelem. V adaptivním přehledu je povolené přeskládání stran, i tak nesmí zmizet aktuální fokus a výběr. Při změně docků fit přepočítá čistou šířku včetně uší.
- **Pružný Přehled stran:** samostatný třetí režim s flex-wrap, při menším zoomu se podle dostupného místa postupně zobrazí více stran. Klik vybere jednu, Ctrl+klik přidá/odebere, Shift+klik rozsah, Enter přepne vybranou stranu na výšku. Sdílený model výběru s miniaturami; drag/move a hromadné akce **nejprve jen HTML demo**, produkčně vždy přes jedno centrální page-ops API a Undo/Redo. Hromadný výběr je **stránkový**, nesmí se míchat s výběrem textu/objektů.
- Miniatury: reprezentují skutečné PDF strany, podporují výběr, přesun, pořadí a kontextové operace podle práv a aktivního nástroje. Nesmí se měnit PDF potichu při pouhém zoomování nebo výběru.

## 6. Význam příkazů Soubor/Objekty

**Nezaměnitelné operace:**
- **Otevřít:** pouze existující PDF; zachovat bezpečné ověřování/historii.
- **Importovat:** vytvoření **nového PDF** převodem jiných podporovaných zdrojových formátů; ne vložení do otevřeného PDF. Formáty a konverzní engine potvrdit samostatně.
- **Sloučit a otevřít:** vybrat více podkladů a vytvořit/otevřít **nové výsledné PDF**.
- **Vložit strany:** vložení vybraných stran či rozsahů do **aktuálního PDF**, s výběrem umístění a bezpečným Undo.
- **Vložit objekt:** pouze Objekty, konkrétní raster/vektor do aktuální stránky; nikoliv celá stránka.
- **Uložit / Uložit jako / Uložit upravenou kopii:** zachovat ochranu originálu, zvolenou aktivní kartu, history/dirty; AutoRecover uložených i rozpracovaných dokumentů. PDF/X, ICC a ostatní výstupní profily jsou samostatné.

## 7. Jediný model výběru Text / Objekty / OCR / Stránky

- Před změnou handlerů provést inventuru současných Go modulů. Jedno fyzické gesto zpracuje **jediný centrální dispatcher**; na základě hit-testu text/raster/vektor/vodítko/stránka rozhoduje, co označit. Univerzální obdélník a laso umějí smíšený obsah; nesmí vzniknout tři konkurenční enginy (Text, OCR, Objekty).
- Text → Výběr: jedna řada **I-kurzor, obdélník, laso, odstavcový textový rámec**, vpravo kruhové ?. I-kurzor kliknutím mimo text vkládá bodový text, odstavcový rámec tažením tvoří odstavcový blok. Kontextový popis základního/rozšířeného výběru pod ?. Převést na křivky a Rekonstruovat text **nepatří do Typografie**.
- Text → Typografie: smíšená hodnota není chyba ani pokyn sjednotit všechny úseky. Jedna změněná vlastnost se uplatní jen na tuto vlastnost; A+/A− relativní. Zachovat zdrojová písma, řádkování, proklad, velikost, zarovnání, horní/dolní index, malá kapitálky, styl, odsazení a případné textové obrysy. Konverze text na křivky je samostatná operace nad vybraným editovatelným textem.
- Rekonstrukce textu z vektorových obrysů/rastru/kombinace má **oddělené fáze** analýza → OCR/rozpoznání → návrh fontu → manuální náhled → potvrzení; neslibovat shodný font ani identický layout. Nebourat původní PDF bez explicitního potvrzení; skutečná rekonstrukce je projektová etapa.
- Objekty: přesný hit-test včetně alfa bitmap, vektorové uzly, procházení překrytými objekty, inteligentní laso „protnuté“ vs „zcela uvnitř“, kontextové operace bez konkurence s Text. Ořez rastru (obdélník, čtverec, ovál, kruh; poměry Volný/1:1/4:3/3:4/16:9/9:16/3:2/2:3/5:4/4:5/21:9) preferuje nedestruktivní clipping a Undo. PPI a resampling jsou nezávislé; RGB/Gray/1bit/CMYK a ICC chránit.
- OCR: vlastní panel ovládání rozpoznávání, ale sdílená geometrie výběru, společná rozpoznávací služba s Text → Rekonstruovat text. Neprocesovat opakovaně nativní editovatelný text bez potřeby.
- **Priorita událostí:** modal/focus input/IME > aktivní tažení konkrétní pomůcky/pole > hit-test konkrétního objektu/textu > stránkový výběr v přehledu nebo miniaturách > obecné gesto plátna. Ctrl+klik na vodítku nevybere omylem stránku; Ctrl+klik mimo vodítko zachovává odpovídající výběr objektů/textu. Esc ruší rozpracované gesto; Enter/dvojklik dokončuje laso, pokud je v režimu lasa. Nedestruktivní výběr sám nemění dirty stav PDF.

## 8. Pravítka, vodítka, skupiny a přichytávání

- **Pravítka leží mimo stránku PDF**: horní vodorovné nad listem; levé svislé nalevo. Tažení **z horního vodorovného pravítka tvoří VODOROVNÉ vodítko Y**, z levého svislého tvoří **SVISLÉ vodítko X**. Starší opačné mapování v3.14 je výslovně zrušeno. Samotný klik pravítka nevytváří linku, teprve drag přes práh ~7 px a ven z pravítka. Fokus pravítka a Enter/Space mohou být přístupná alternativa.
- Každé vodítko má per-document/page ID, osu, přesnou souřadnici mm/PDF, individuální barvu, zámek, viditelnost a volitelnou skupinu. Výchozí barva nového vodítka nepřebarvuje stará. Pevná linka ~1 CSS px, širší transparentní hit-area, plná čára; neplést s dočasným přerušovaným snap.
- Levý klik vybere jedno; **Ctrl+klik** přidá/odebere. Checkboxy seznamu sdílejí stejný výběr. **Ctrl+tažení** nadprahové duplikuje (nové ID, původní nezměněno, kopie nezdědí skupinu automaticky); tažení bez Ctrl posouvá, globální a individuální zámek obojí chrání. Esc/pointercancel vrací stav. Barva se pouze předvybere v pickeru a potvrzuje **Použít/✓**, pro jeden/více/skupinu/default odděleně.
- **Pravé kliknutí**, nikoli hover, otevře malé lokální menu obsahující pouze barevné políčko, zámek a ikonu **koše**. Není tam položka Vybrat. Koš = odstranit, křížek = zavřít/zrušit. Uzamčené vodítko nelze mazat/posouvat/duplikovat.
- Klik na linku nebo změna magnetu/mřížky/vodítek dočasně otevře Pomůcky rozvržení ve Vlastnostech; klik mimo nebo dokončení vrátí předchozí nástroj, pokud uživatel mezitím neprovedl ruční explicitní přepnutí.
- Skupiny: pojmenované per-page, členové vizuálně mírně odsazeni, skupina má rozevírací šipku a oko/barvu/zámek/koš. **Sbalení seznamu NEMĚNÍ viditelnost na stránce ani snap**. „Rozdělit“ ponechá členy, „Smazat skupinu“ odstraní i její členy (s kontrolou zámků). Přepnutí stránky nepoškozuje ostatní data.
- Magnet je nadřazený přepínač; jeho vypnutí uchová všechny vybrané cíle. Mřížka viditelná ≠ přichytávání k mřížce. Cíle: účaří, hrany objektů, vodítka, středy objektů, středy stran, uzly, mřížka, okraje/střed stránky, rozestupy, případně widgety/textové rámy/předěly. Tolerance v obrazových px, geometrie podle PDF bodů/CropBox/rotace/DPI, hysteréze a prostorový index. Zamčené/skryté cíle filtrovat. **Skutečný SnapEngine je zatím návrh**; UI overlay není důkaz geometického engine.
- Dočasné smart indikátory jsou přerušované/tečkované a tématické (akcent motivu), mizí při ukončení manipulace; trvalá vodítka mají uživatelskou barvu. Nejsou tiskovým obsahem a standardně se neukládají do PDF.

## 9. Separace, tisk, poznámky a PDF kompatibilita

- Panel Separace zachovat ve stávajícím členění; žádné dvě nově vymyšlené záložky. Převody „stupně šedi (DeviceGray)“, „složená K100“ a „pouze K“ jsou **tři odlišné operace**; správa ICC rozlišuje přiřazení vs převod. V jednom bloku Výstup a konverze 3 svislé řádky, každý s jednou ikonou a dropdownem. Jediná izolovaná separace má výchozí šedý (volitelně barevný) pohled. Značky CMYK/spot jsou autentické informační vzorky, ne CSS akcent. TAC limit a měření hodnoty pod kurzorem nepředstírat, pokud není skutečně vypočteno.
- Formulářové widgety/odkazy/podpisy nejsou „poznámky“. Pracovní poznámky zůstávají oddělené standardní PDF anotace; zobrazení nastavovat zvlášť pro každý dokument. Výchozí tisk/export **bez pracovních poznámek**, běžné pracovní uložení je zachová. Výslovně volitelný tisk/export s poznámkami.
- Bezpečné ukládání AutoRecover, atomic write, potvrzování úprav, oddělení zdrojového a výsledného PDF, zachování grafických/písmových dat a podpisů. Nikdy neodstranit náhodně /AcroForm, /Fields, /Kids, /Widget, /V, /AS, /AP, /Annots, odkazy a podpisy při exportu „bez poznámek“.
- PDF JavaScript nespouštět bez bezpečnostního modelu; XFA bezpečně detekovat a uvést omezení, ne tiše zničit. Náhled/preflight a rasterizace nesmějí poškodit editovatelné PDF objekty. Konečné produkty testovat v nezávislých PDF aplikacích.

## 10. Formuláře: schválená F1 pravidla, návrh F2 a datové importy

**F1 stále SCHVÁLENO co do funkčního rozsahu, ale implementace je až po dokončení potřebného UI:**
- F-Q1: F1 vyplnit/uložit jako první formulářová etapa, později F2 návrh a F3 zploštění.
- F-Q2: po detekci AcroForm jen nabídnout přechod, **automaticky nepřepínat**.
- F-Q3: formulář lze měnit pouze v aktivním Formuláře → Vyplnit; respektovat předepsaný vzhled; typografický editor ve F2 později.
- F-Q4: pouze UI jemný akcent vyplnitelných polí, uživatelsky vypínatelný.
- F-Q5: sbalitelný přehled v pravém panelu, klikací navigace.
- F-Q6: řadit seznam podle stránky/polohy, skutečné pořadí Tab podle PDF nezávisle.
- F-Q7: upozornění na chybějící údaje **neblokuje „Uložit přesto“** ani AutoRecover.
- F-Q8: stav průběžně, formátové chyby až po opuštění pole; souhrn při ukládání.
- F-Q9: chybné pole jemně červeně zvýraznit na stránce i v seznamu, po opravě odstranit; netisknout, nezapisovat do PDF.

**Návrhové režimy HTML v3.26:** Ve Formuláře jsou tři záložky **Vyplnit | Návrh polí · F2 | Z dat**. F2 editor: paleta text, více řádků, checkbox, radio, combo, seznam, tlačítko, datum, podpisové pole; klik/tažení na stránce, výběr, přesun, seznam a vlastnosti (název/T, tooltip/TU, výchozí DV, aktuální V, exportní volby, písmo, velikost, Tab pořadí, povinné, čtení, vzhled a viditelnost). Dělá pouze demo objekty v paměti; **nepředstírat skutečné AcroForm ani podepsání**.

**Importy — pouze návrh:** CSV UTF-8/BOM s validním escapováním a oddělovači; Excel XLSX se sheet/list/range; Google Tabulky OAuth read-only a snapshot konkrétní záložky; volitelné XML řešit odlišně od XLSX (vyžaduje XSD, žádné externí entity). Jeden záznam/řádek = jedno zamýšlené výsledné PDF. Mapování sloupců na plně kvalifikované AcroForm ID (se stabilním record_id), explicitní pravidla prázdných buněk, datum ISO 8601, číselné formáty, ANO/NE na správné exportní hodnoty /AS, rádio/combo podle /Opt; validace a preview, bezpečné názvy souborů, žádný přepis zdroje, rollback a log. V HTML funguje jen **náhled CSV**, XLSX/Sheets parser a skutečný batch export **neexistují**. Návrhová specifikace je rozpracovaná v historickém spisu, zde jsou závazné koncepční zásady; **bez nové explicitní autorizace F2 nebo import NEIMPLEMENTOVAT**.

## 11. Zkratky a gesta – centrální registr

Jednotný registr s kontextem focus (global / canvas / inline text / dialog / thumbnails / guides / flow). Stávající Go handler má přednost, dokud nebude prověřena a schválena migrace. Vlastní klávesové kombinace: zaznamenat, kontrolovat kolize, Esc ruší, Backspace vrací původní, OK/Použít ukládá, Zrušit vrací. AltGr a českou klávesnici chránit; nové výchozí Ctrl+Alt kombinace raději nevytvářet.

Doložený **statický** inventář existujících zkratek zahrnuje Ctrl+O/F/W/S/P, Ctrl+Tab, Ctrl+Z/Y, Ctrl+X/C/V/A, F3, F11, PageUp/PageDown, Ctrl+plus/minus/0, Ctrl+wheel v plátně, T/V/M/L v původních kontextech. **NEPŘEPISOVAT** původní T/V/M/L jen kvůli HTML: jejich plánovaná migrace na univerzální výběry zůstává požadavkem na samostatný audit a schválení. Nové F9 a Ctrl+Shift+M/H/R jsou **návrhy**, ne potvrzené produkční zkratky.

Gesta v plátně: priorita konkrétního vodítka/objektu vs stránka; Page overview Ctrl+klik výběr stran, Enter přizpůsobit; Guide Ctrl+klik multiselect a Ctrl+drag duplikace; mazání Delete jen v příslušném výběrovém kontextu; Esc ruší aktuální gesto. Přesné tabulky a testy jsou v **mapa_zkratek_FINAL_v6.0.md** (jediná platná mapa).

## 12. Co je pouze koncept a nesmí být samovolně spuštěno

Preflight nad existujícím PDF/X, vizuální diff dokumentů, pokročilá měření, vlastní nový OCR engine, detekce cizích fontů, hromadné editace stránky/objektů, konverze importovaných ne-PDF souborů, tvorba F2, zploštění F3, datové dávky, externí cloudový import, export barevných separací s novými algoritmy, dlouhé zpracování a nové šipkové anotační workflow. U každého návrhu nejprve zjistit, co již existuje, a zvolit malý implementační krok.

## 13. Testovací matice před označením za hotovo

**Vždy:** statická kontrola a build ve Windows; start aplikace, otevření konkrétního PDF, close/reopen, Ctrl+S/Uložit jako, AutoRecover, render při 100/150/200 % DPI, resize okna, zkratky/fokus/IME, EN/CZ, Undo/Redo, neztratit existující příkazy.

**UI:** 6 × 2 palet a computed styles (min. text 4,5:1, grafické prvky 3:1), 980/1100/1250/1600 px, horní textové menu bez rámu, aktivní ikonová lišta, řádky nástrojů bez rámů a s levým akcentem, dvě horní + levá/pravá pomocná ouška, scrollbary, nápověda ?, odpojení PDF dat od motivu. Náhled neznamená realizovanou Win32 funkci.

**Navigace:** kolečko/hranice stránek, přizpůsobení šířce bez vnějšího rámu, lokální Ctrl+kolečko hlavní PDF/miniatury, mimo zóny bez odchytu, Enter z pružného režimu na výšku, Ctrl/Shift klik, přetahování s Undo a zachováním správného pořadí. Chrome/Edge HTML zvlášť od Win32.

**Výběr a pomůcky:** text/objekty/rastr/mix, rámeček/lasso/overlap, guide X/Y z příslušného pravítka, klik vs drag >7 px, Ctrl+klik vs Ctrl+drag, zámek/skrytí/skupina, barvy Použít, dočasné Pomůcky rozvržení a návrat, Snap overlay bez zápisu do PDF.

**Formuláře (až etapa F1):** text/multiline/checkbox/radio/combo, /AcroForm /Fields /Kids /Widget /V /AS /AP, česká diakritika, více widgetů jednoho názvu, neplatný formát, souhrn, Uložit přesto, XFA, podpisy, vložené fonty, skutečná reprodukce po znovuotevření, export bez pracovních poznámek.

## 14. Způsob práce Codexu

1. Nejdříve přečíst **tento jediný dokument**, **CODEX_IMPLEMENTACNI_PLAN_v6.0.md**, finální mapy a root AGENTS.md; pak analyzovat skutečný kód. **Neprocházet staré v3.x/v4.x/v5.x dokumenty jako souběžná zadání**.
2. Zkontrolovat aktuální větev/commity/změny, zapsat, co už je rozpracované. Nic nemaž ani automaticky nemigruj.
3. Předložit stručně **A/B/D/P** inventuru, rozdíl UI v3.26 vs Go/Win32, **návrh jediného prvního UI kroku** včetně dotčených souborů, testů, rizik a bezpečného rollback plánu. Počkat na souhlas.
4. Potom pouze výslovně schválená etapa; malé diffs a ověřené testy. Nepřepisovat funkční moduly jinou architekturou, nevyměnit PDFium ani nenasazovat nový webový shell bez schválení. Před commitem diff+testy; push podle dohody. Nikdy force-push/destruktivní reset.
5. Dokumentaci aktualizovat jen v této **v6.0 sadě**; historické spisy neměnit. Zjištěné skutečné neshody kódu s mapami označit **odchylka k ověření** a předložit, nikoli uměle „vyřešit“ přejmenováním starého příkazu.
6. Veřejné repo: žádná zákaznická PDF, neveřejná data, certifikáty, hesla ani přístupové tokeny. Výstupní logy anonymizovat.

**První očekávaná odpověď Codexu:** „Audit existujícího UI, seznam toho, co je skutečně hotové, seznam rozpracovaných změn včetně formulářů, konfliktní místa, návrh Etapy 1 – nejmenší samostatné UI změny a testy. Žádné změny kódu ani commity před schválením.“
## 15. Známé nedokončené kontrolní body prototypu

HTML v3.26 prošlo dříve jen syntax checkem, **ne kompletními prohlížečovými a Windows testy**. Navíc se našly **použité, ale nedefinované SVG symboly** `i-panel_left`, `i-arrow_line`, `i-transform_object`, `i-plate_all`, `i-plate_single`. Tyto chybějící vazby patří do **úvodního UI auditu**, nikoli do produkční F1 nebo do samočinného generování nových ikon. Nejprve ověřit původní atlas, pak po schválení opravit bez změny command ID. Pokud se najdou další chyby HTML, uvést je jako B/D/U problémy s reprodukcí a testy; aktuální v3.26 nepovažovat za bezchybný hotový editor.
