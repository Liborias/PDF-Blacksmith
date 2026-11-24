# PDF Blacksmith — konsolidované pracovní zadání pro Codex, UI v3.9

**Datum:** 24. listopadu 2026  
**Repozitář:** `Liborias/PDF-Blacksmith`, větev `main`, pracovní složka `docs/pracovni/`.  
**Stav:** Schválená specifikace návrhu a HTML demonstrace, **nikoli pokyn k nevyžádanému přepsání produkčního Go/Win32/PDFium kódu**. Nejprve inventura skutečné implementace, kolizí handlerů, návrh dílčí implementace a testy, až potom explicitní povolení Codexu.

## 0. Rozcestník a priorita pravidel

1. Kořenový zdroj `../specifikace_Codex_v5.0.md` zůstává závazný pro schválené stávající funkce (SEL/TXT/OBJ, formuláře, soubor, podpis, ochrana PDF), nesmí se vymazat ani obejít.
2. Toto je **nejnovější, sjednocující UI zadání**, navazuje na `TEXT_VYBER_A_SMISENA_TYPOGRAFIE_v3.3.md` a průběžné odsouhlasené návrhy v3.4 až v3.8. Starší pravidla, která byla mezitím výslovně nahrazena (viz níže), nesmí Codex aplikovat současně s novými.
3. Referenční HTML: [cely_vizual_v3.9.html](cely_vizual_v3.9.html). Nová SVG sada: [ikony/pdf_blacksmith_nove_v3.9.svg](ikony/pdf_blacksmith_nove_v3.9.svg). Mapa ikon: [../mapa_ikon_v4.5.md](../mapa_ikon_v4.5.md).
4. Původní rozhodnutí o F1/F2/F3 a implementačních pravomocích z `../README.md` stále platí. Každá odsouhlasená změna je zatím návrhem, nikoliv hotovou funkcí PDF.

### Výslovně zrušené / nahrazené starší varianty

- Dvě konverzní tlačítka **nepatří do Typografie**: od v3.7 jsou pouze v **Text → Výběr**; v3.5/v3.6 byly přechodným stavem.
- Horní lišta nepoužívá oddělenou šipku a textový kurzor ani split-tlačítko s rozbalovací nabídkou z v3.8. Používá se **jedno dvoustavové cyklické tlačítko**. Volba univerzálního obdélníku a lasa zůstává samostatná.
- Textový obdélník a textové laso jsou v pravém panelu kontextovými zrcadly stejného výběru jako univerzální horní nástroje; **nevytvářet paralelní selection engine**. Odstavcový blok je stále zvláštní nástroj pro vytváření odstavcového textu, není univerzální výběr.
- Zobrazení „Po stránkách / Souvislé zobrazení“ musí být přístupné z druhé horní lišty; samostatný výběr v nabídce Zobrazení je synchronizovaným alternativním vstupem.

## 1. Definitivní změny v3.9 — dvě dvoustavová tlačítka na druhé horní liště

**Obecná konvence:** ikona na dvoustavovém tlačítku zobrazuje **akci příštího kliknutí**, nikoli označení právě aktivní funkce. Po kliknutí se spustí zobrazená funkce, ikona se přepne na alternativu. `title`/`aria-label` musí jasně označit nynější stav i následující akci, současný stav je případně uveden i na stavovém řádku. Klikání na dva stavy nesmí během editace samo měnit PDF ani provádět Undo/Redo. Po změně stavu z jiné nabídky se tlačítko znovu synchronizuje.

1. **Kurzor** — při výchozí aktivní šipce nabízí ikona *textový kurzor*. Klik volí I-kurzor (aktivuje Text), další klik volí šipku (automaticky hit-testuje text/objekty). Textový kurzor na prázdné ploše zakládá bodový text pouze po skutečném potvrzeném kliknutí, ne při samotném přepnutí.
2. **Zobrazení dokumentu** — při výchozím stránkovém režimu ikona nabízí *Souvislé zobrazení*. Klik zobrazí všechny stránky pod sebou s posuvníkem, další klik vrátí jedinou aktuální stránku. Změna režimu nezmění obsah PDF, měřítko stanovené uživatelem ani vybraný obsah; stránková navigace/miniatury i stavový řádek odpovídají aktuálnímu režimu.
3. Dvě ikony `i-view_pages`, `i-view_continuous` jsou v novém atlasu. Cyklický kurzor používá `i-cursor` a `i-text_caret`. Oba prvky jsou ovladatelné klávesnicí, mají kontextový tooltip a nesmí zmizet při automatické redukci sekundárních ikon.
4. Režimy výběru **Obdélník** a **Laso** zůstávají nezávislými třetími/případně čtvrtými nástroji. Dvoustavová konvence se používá pouze tam, kde jde přesně o dvě vzájemně výlučné funkce a není potřeba třetí.

## 2. Stav sjednocení UI — schválené požadavky 23.–24. listopadu

| Oblast | Aktuální závazné chování |
|---|---|
| Kontextový výběr | Šipka na objektech vybere objekty a aktivuje Objekty; v textu určí text a aktivuje Text; smíšené oblasti zachovají všechny typy. Výběr neznamená úpravu dokumentu. |
| Text → Výběr | Textový kurzor, obdélník, laso, odstavcový textový blok v jediné řadě; `?` vpravo. Bodový text vzniká klikem kurzoru mimo existující text; odstavcový rámec vymezuje tažení. |
| Text → Typografie | Smíšené hodnoty neměnit bez explicitního zásahu. Tlačítka mají jednotný, **bez diagonály** vzhled. A+/A− relativně; přímá hodnota sjednocuje jednu vlastnost. |
| Výběr v OCR | OCR používá sdílenou šipku/obdélník/laso a sdílenou službu rozpoznávání; zachovat stávající kontext OCR, není druhý dispatcher. |
| Převody textu | **Převést na křivky** jen vybraný editovatelný text; **Rekonstruovat text** u vektorů/rastrů/kombinací, dialog porovnání, volba fontu, ruční korekce a geometrie. Celý dokument až pozdější etapa. |
| Dvojí OCR zdroj | Vektorová analýza obrysů + OCR bitmap (a kombinace), bez automatického odstranění originálu. Vrstvy, maskování a skutečná výměna textu jsou odlišné operace. |
| Práce s obrázky | Objekty: pouze vybraný raster, rozměry px, PPI odlišné od resamplingu, algoritmus převzorkování, barva RGB/grayscale/1bit/CMYK s ICC, ořez, undo, per-instance XObject. |
| Přichytávání | Magnet, mřížka a pravítka v druhé horní liště, podrobnosti v Zobrazení → Přichytit, samostatně hrany, středy, strany, uzly, baseline, vodítka, mřížka, okraje stránky, rozestupy. Zatím jen vizuální UI; SnapEngine později. |
| Nápověda | Vysvětlení ikon se dynamicky mění na hover/focus v **rezervovaném prostoru** pod ikonami, bez barevného rámečku; nativní plovoucí tooltip/popup musí být NAD ikonami. Podrobnosti pod `?`. |
| Druhá lišta | Šest nabídek při dostatečné šířce, adaptivní ukrývání vedlejších ikon podle skutečného prostoru; vždy zachovat důležité nástroje, vyhledávání, zoom a uložit. |
| Dokument a panely | Pravá i levá skupina panelů musí zmenšovat volnou pracovní plochu; stránka nesmí být překryta; při automatickém zobrazení přepočítat měřítko. Min. návrhové okno 980×540 CSS px, min. zoom 30 %. |

## 3. Bezpečné vlastnictví gest a typová detekce

- Jediný `PointerGestureDispatcher` a struktura `SelectionScope` (záměr, druh nástroje, detekovaný obsah, page-space polygon/rect, běhy textu a objekty) nad existujícími PDF koordináty; nejprve inventura handlerů, klávesových modifikátorů a výkonu.
- Jedno fyzické gesto smí změnit výběr právě jednou. Univerzální obdélník/laso rozpoznají text, objekty i kombinace; po dokončení se kontextový panel přepne, nikoli smaže část výběru. OCR oblast může zůstat v OCR panelu až do potvrzení operace.
- Miniatury stránek, prostřední tlačítko pro výběr, caret gestikulace a dřívější SEL-01 až SEL-03/TXT/OBJ musí zachovat svou sémantiku; při konfliktu má platit výslovně poslední schválené pravidlo a nejasnost se před implementací projedná.
- `Esc` ruší rozpracované gesto, Enter/dvojklik ukončí laso. Při nejasném překryvu vyžadovat explicitní výběr vrstvy/objektu a možnost návratu.
- Převod textu na křivky a rekonstrukce jsou operace **nad výběrem**, nikoli pátý či šestý režim výběru; každý zásah je jeden vratný krok historie. Nezasahovat do podpisů, odkazů, anotací, formulářů bez jasného záměru.

## 4. Do budoucna, nikoliv automaticky implementovat

- Skutečné OCR/rozpoznávání fontů a rekonstrukce textu mají oddělené fáze: extrakce geometrie → znaky → identifikace kandidátů fontu → rekonstrukce sazby → náhled a potvrzení. Výzkumné projekty nutno nejprve licenčně prověřit. Bez zaručení shodného fontu či sazby z cizích křivek.
- Přichytávání jako centrální `SnapEngine`: prostorový index, preference cílů, tolerance a hystereze v obrazových pixelech, vizuální vodítka, uživatelská vodítka z pravítek, přesná mřížka, účaří, rovnoměrné rozestupy, dokumentové uložení, netiskové overlaye.
- Pozdější dávkové OCR / převod textu na křivky / rekonstrukce celé stránky či celého dokumentu; podpůrné kontroly při hromadném zásahu.
- Produkční uvedení jen po etapách a výslovném schválení; nevytvářet novou implementaci před inventurou existujících handlerů Go/Win32 a PDFium.

## 5. Akceptační testy pro v3.9

1. Dvoustavový kurzor: první klik textový kurzor a ikona se změní na šipku; druhý klik šipka a ikona se vrátí na I-kurzor. OCR, Text, Objekty reagují na výběr bez zdvojení handleru.
2. Stránkové/plynulé: první klik zobrazí 3 vzorové stránky místo jedné, ikonou nabídne návrat po stránkách; druhý klik vrátí jednu stránku. Menu Zobrazení i stavový řádek se synchronizují.
3. Testy šířek 980, 1100, 1250, 1600 px, v produkčním Win32 navíc 100/150/200 % DPI. Přizpůsobení stránky nesmí přikrýt pravé panely.
4. Text: smíšené hodnoty, pouze jedna měněná vlastnost, vzhled bez diagonály, délky popisků neskáčou, tooltip není před místním vysvětlením.
5. Výběr obdélníkem/lasem nad textem, bitmapou, vektorem a smíšenou oblastí; OCR zachová kontext.
6. Obrázek: 1200×800 px → 600×400 při zachování poměru, změna režimu je pouze vizuální simulace; žádný skutečný PDF export v HTML.
7. Žádný krok pro čtení UI neoznačit jako obsahové Undo; funkce snímání/rekonstrukce nepoškodí historický obsah.

---
## 6. Archiv schválených podrobných instrukcí v3.4–v3.8 (používat s prioritami z části 0)

Následují kompletní pracovní dodatky, aby se neztratily detaily, testy, technická omezení a dříve potvrzené konvence. Tam, kde text archivu popisuje přechodnou polohu tlačítka nebo starší split-button, **má vždy přednost část 0 až 5 tohoto dokumentu**.


### Archiv 1: `PDF_Blacksmith_Codex_UI_v3.4_kompaktni_panely_a_responzivita.md`

# PDF Blacksmith — UI v3.4: čisté panely, jednotná nápověda a adaptivní horní lišta

**Datum:** 23. listopadu 2026  
**Stav:** schválený návrh chování v HTML prototypu. **Není to pokyn k okamžité změně produkčního Go/Win32/PDFium kódu.** Implementace teprve po inventuře stávajících handlerů a výslovném souhlasu.  
**Navazuje na:** `PDF_Blacksmith_Codex_Text_v3.3_vyber_smisene_formatovani.md` (zachovat beze změny všechny čtyři režimy výběru a pravidla smíšeného formátování).  
**HTML referenční prototyp:** `PDF_Blacksmith_HTML_navrh_v3.4_ciste_panely_adaptivni_lista.html`.

## 1. Základní princip: čisté ovládání bez trvale otevřených instrukcí

- Pravý panel **Vlastnosti aktivního nástroje** zobrazuje **příkazy, ovládací pole, dostupné režimy, hodnoty a nezbytné stavové/validační informace**. Není to trvalá učebnice obsluhy.
- **Jednotné chování ikon**: kurzor nad ikonou dynamicky zobrazí její krátký konkrétní popis. Přejetí na další ikonu jej **ihned nahradí**, opuštění ikony jej odstraní. Klávesnicový fokus poskytne stejné vysvětlení. Nezavádět trvalý placeholder „Najeďte na ikonu…“.
- Dynamický popis je **obyčejný text přímo na pozadí panelu** (barva panelu, bez vínového rámečku/podbarvení); patří co nejblíže související skupině ikon. Vínové zvýraznění je vyhrazené aktivním režimům/tlačítkům, nikoli vysvětlivkám.
- **Delší vysvětlující texty**, provozní poznámky a instrukce uložit pod **`?` vedle nadpisu odpovídající sekce** nebo do existujícího místního popoveru. Po kliknutí se zobrazí, znovu kliknutím zavře. Tím se nemá skrývat zásadní výstraha, chyba, zobrazená hodnota nebo informace o neuložené změně.
- V záložce **Text → Výběr** odstranit trvale zobrazený vínový popis zvoleného režimu a nadbytečné instrukce. **Čtyři ikony ve stejném pořadí + `?` odsazený doprava** zůstávají; podrobné chování výběru je už v jejich místní nápovědě. Přejetí myši nabídne jednu krátkou vysvětlivku právě dotčené ikony.
- V záložce **Text → Typografie** nenechávat trvale otevřené instrukce o smíšených hodnotách, změnách stupňů písma, řádkování a barevných efektech. Zachovat skutečné ovládací prvky, demo přepínač výběru a náhled sazby; informace jsou přístupné v nápovědě a titulcích.
- Zachovat všechna existující pravidla formulářů, objektů, OCR, separací, Souboru, Zobrazení a dalších panelů; redesign nesmí měnit jejich operace ani handler výběru.

## 2. Druhá lišta odshora — přizpůsobení podle **skutečného místa**

- **Nikdy neskrývat část nabídky Soubor, Zobrazení, Nástroje, Úpravy, Formuláře a Pomoc jen na základě pevného breakpointu.** Šest nabídek se zachová i v nemaximalizovaném okně.
- Zjistit skutečně dostupnou šířku po odečtu menu, oddělovačů, vnitřních odsazení a ovládacích prvků. Po změně velikosti okna znovu vypočítat prostor.
- Standardně zobrazit všechny nástrojové ikony. **Teprve když se reálně nevejdou**, skrývat **volitelné** v pořadí: globální nápověda v pravé části, globální Nastavení, dočasné sbalení panelů, Přizpůsobit šířce, Přizpůsobit stránce. V menu Pomoc a Zobrazení zůstávají dané funkce dostupné.
- Prioritně zachovat běžný výběr, zoom − / procenta / zoom + / 100 %, rychlé hledání (jeho šířka se může zmenšit), předchozí/následující stranu, stránkový indikátor a Uložit.
- Žádné překrývání ikon, skryté části v oříznutém flexu ani skákání menu jen z důvodu změny stavu maximalizace. Rozhoduje **vypočtená šířka**, nikoliv název režimu okna.

## 3. Minimální rozměry okna a zobrazení dokumentu

- Pro **HTML ukázku** nastavena **minimální šířka 980 CSS px** a **minimální výška 540 CSS px**. Je to návrhový limit podle prostoru druhé lišty, který bude nutné znovu ověřit v nativním DPI prostředí a podle finální velikosti ikon. V prohlížeči uživatel stále může zmenšit samotné okno; prototyp má pak vodorovné posouvání místo ztráty ovládacích prvků.
- **Produkční Win32**: minimální rozměr okna řešit systémovým mechanismem (např. `WM_GETMINMAXINFO`) s přepočtem podle efektivního DPI a reálných rozměrů ovládacích prvků. Neomezovat okno jen na základě šířky PDF stránky.
- **Minimální zoom dokumentu: 30 %**, maximální dosavadních 900 %. V rozbalovacím poli nabídnout i 30 % a 50 %. Funkce Přizpůsobit šířce/stránce musí respektovat dolní mez; při úzké pracovní ploše lze dokument posouvat.
- **Minimální velikost okna a minimální zoom jsou nezávislé hranice**: změna šířky okna automaticky nepřepisuje uživatelem nastavené měřítko, s výjimkou aktivního režimu automatického přizpůsobení.

## 4. Technická pravidla pro Codex

1. Udržet současné `data-ui-action`, `data-text-command`, `data-file-*`, `data-view-*` a další funkční handlery; u nápovědy se nesmí spouštět žádná editace PDF.
2. Centrálně doplnit popisy ikon nebo jejich jednoznačné mapování. Hover, focus a opuštění používat napříč panely konzistentně; nepsat pro každý nový nástroj úplně jinou implementaci.
3. Žádná nápověda se nesmí započítat jako změna dokumentu, krok Undo/Redo ani AutoRecover událost.
4. Zachovat ovládání pomocí kláves: ikony s `aria-label`/`title`, `?` přístupné tabulátorem a aktivací Enter/Space; u existujícího textového popoveru zachovat Escape.
5. Neodstraňovat chybové hlášky, informaci o neplatné hodnotě, důležité stavební štítky (např. aktuální měřítko, stav podpisu, aktivní tisková deska) ani stavy tlačítek. Jde o úklid výhradně **vysvětlujících** textů.
6. Před převzetím výběrových událostí ze stávajícího Go/PDFium provést inventuru kolizí, jak vyžaduje v3.3.

## 5. Ověřovací scénáře

- Okna 1600, 1350, 1250, 1190, 1100, 1050, 980 CSS px: šest menu stále viditelných, volitelné ikony se mění jen při skutečném nedostatku prostoru, hledání a uložení zůstávají přístupné. Kontrola na 100 %, 150 % a 200 % DPI v produkčním GUI.
- Text → Výběr: čtyři ikony na jednom řádku; hover mezi ikonami okamžitě vymění nápovědu; žádný trvale zvýrazněný popis; `?` ukáže základní a rozšířený výběr bez změny režimu.
- Text → Typografie: smíšené formátování, A+/A−, změna jediné vlastnosti a výběrové režimy fungují stejně jako v3.3; delší vysvětlivky jsou schované.
- Objekty, OCR, Separace, Soubor, Zobrazení: hover/fokus jednotlivých ikon a otevření otazníku v sekci; žádné nechtěné operace nebo ztráta aktivního nastavení.
- Zoom 10 % se omezí na 30 %, 30 % zůstane platná hodnota, fit-výpočty nikdy nepadnou pod 30 %.
- Ověřit Javascript console bez chyb a minimálně následující nástroje: Formuláře, Objekty, Text, Soubor, Zobrazení.

## 6. Doporučené další pořadí ladění

1. **Jednotné ovládání napříč typy objektů:** režimy kurzoru, měřicí nástroje, klávesové zkratky, kontextové menu a jejich kolize; navázat na pravidla výběru.
2. **Vlastnosti textu a objektů:** editace více vybraných objektů, smíšené vlastnosti, stavy neaktivních polí a společný krok historie.
3. **Práce s panely a pracovní plochou:** oddokování, automatické uložení rozvržení, maximální využití plochy a přesné chování zoomu při přepínání nástrojů.
4. **Správa barev a export:** ICC, RGB→CMYK/K, grayscale profily, zachování vektorů a náhled separací; oddělit vlastnosti dokumentu od výstupních voleb.

---
**Výslovně:** Tento HTML soubor je **funkční prototyp ovládacích prvků**, nikoli editor PDF. Neprovádí skutečnou editaci existujícího PDF ani změny v Go/PDFium. Při převodu do produkce převzít pouze odsouhlasená pravidla, ne slepě přepsat existující funkční kód.



### Archiv 2: `PDF_Blacksmith_Codex_UI_Text_v3.5_vyber_krivky.md`

# PDF Blacksmith — specifikace v3.5: sjednocený výběr, stabilní nápověda, text ↔ křivky

**Datum:** 23. listopadu 2026  
**Stav:** odsouhlasený návrh UX pro HTML a výzkumný plán pro rekonstrukci křivkového textu; **NEPROVÁDĚT automaticky zásahy do produkční aplikace Go/Win32/PDFium**. K tomu je nutná inventura existujících handlerů a následné schválení konkrétní implementace.  
**Výchozí dokumenty:** `specifikace_Codex_v5.0.md`, `PDF_Blacksmith_Codex_Text_v3.3_vyber_smisene_formatovani.md`, `PDF_Blacksmith_Codex_UI_v3.4_kompaktni_panely_a_responzivita.md`.  
**Demonstrace:** `PDF_Blacksmith_HTML_navrh_v3.5_vyber_krivky_rekonstrukce.html`.

## 1. Priorita a pravidla změn

1. Nezrušit žádné již schválené chování výběru, Typografie, Formulářů, OCR ani Separací. Nenahrazovat práci s PDF jen nápisem nebo příslibem.
2. **Jeden fyzický vstup = jeden jediný výběrový handler.** Žádné paralelní zpracování téhož `pointerdown`, dragu, kliknutí a zkratek dvěma nástroji.
3. HTML pouze ukazuje chování UI a stavové přepínání; **neprovádí skutečný převod textu, OCR, font matching ani úpravy PDF**.
4. V produkci před jakýmkoliv zápisem ověřit dokument, typ objektů, zámky, oprávnění, existující digitální podpisy, AutoRecover a Undo/Redo. Zásahy do podepsaného PDF mohou zneplatnit podpis — explicitně varovat.
5. Vektorové objekty neměnit na bitmapu. Při neúspěšné rekonstrukci zachovat původní geometrii beze změny.

## 2. Stabilní oblast nápovědy pod ikonami

- Pod **každou skupinou ikon** se od začátku vykreslí slot s **pevnou výškou**, nikoli prvek vytvořený teprve prvním hoverem.
- Při najetí myší nebo fokusování klávesnicí se mění pouze obsah, nikdy rozměr či odsazení slotu. Po odjezdu/focusout se obsah vymaže, slot zůstává.
- Obsah přímo na pozadí panelu, bez vínového boxu. Dlouhé detaily jsou v klikací lokální `?`.
- Výchozí experimentální rozměry pro prototyp: ikonové skupiny `64 px`, skupiny s `?` pro výběr objektů/křivky `69 px`, výběr textu `85 px`, velikost písma `49 px`, tlačítkové příkazové řádky `59 px`; rozměry podléhají testům s 100/150/200 % DPI a může se použít pevná výška dle dostupných fontových metrik. Delší popis uvnitř slotu může mít vlastní scroll, **nikoliv zvětšovat výšku celého panelu**.
- V celém panelu nesmí při přejetí z jedné ikony na druhou poskočit jediný následující ovládací prvek (`Δy = 0`).
- Nepřekládat skryté dlouhé texty nápovědy zpět do hlavního UI.
- Zachovat adaptivní druhou horní lištu z v3.4: neredukovat její ikony při dostatečném volném prostoru. Respektovat nastavené minimum okna 980 × 540 a dokumentový zoom 30 % jako **limity návrhu**, nikoli omezení uloženého PDF.

## 3. Jednotný model výběru (výběrové handler registry)

### 3.1 Kontexty a stav

V produkci zavést **jediný koordinátor vstupu** (`SelectionCoordinator`, název pracovní), který spravuje oddělené kontexty:

| Kontext | Lokální režimy | Vybíraná data | Poznámka |
|---|---|---|---|
| **Text** | caret, obdélník, textové laso, odstavcový rámec | znaky/úseky a hranice textových bloků | `caret` klik mimo text vytváří **bodový text**; `text_box` jen tažení vytváří **odstavcový text** |
| **Objekty** | kurzor, obdélníkový výběr, objektové laso | PDF stránkové objekty/jejich skupiny | žádný textový caret, ani omylem |
| **Miniatury** | jednotlivý, Ctrl přidá/odebere, Shift rozsah od kotvy | indexy stránek | zachovat oddělenou množinu vybraných stran |
| **Formuláře** | fokus/widget editing | formulářová pole | dodržet vlastní editor/widget handler |
| **OCR, kreslení a další kontexty** | aktivní režimy lokálního nástroje | oblast nebo objekt svého nástroje | žádné skryté dědění Textu |

Udržovat explicitní `activeContext`, `activeMode`, `pointerOwner`, `gestureState`, `selectionAnchor` a vybrané entity pro každý kontext. Vypnutí/změna nástroje bezpečně ukončí rozpracované gesto a uvolní `pointer capture`; nepřeklíží výběr jiného kontextu.

### 3.2 Horní výběrové ikony

- Dvě horní ikony **nejsou další nezávislé módové přepínače**.
- Když je aktivní **Text**, horní kurzor odpovídá `caret`, horní obdélník `text-rect`. Při textovém lasu či odstavcovém rámci není žádná z těchto dvou ikon aktivní.
- Když jsou aktivní **Objekty**, horní kurzor odpovídá `obj-select`, horní obdélník `obj-rect`. U objektového lasa není aktivní žádná z těch dvou.
- Z ostatních kontextů horní výběrový příkaz **výslovně přepne do Objektů** a aktivuje příslušný režim, místo aby současně prováděl OCR nebo editaci formuláře. Zobrazit tento kontext v tooltipu.
- Vybraný režim Textu a Objektů se během přechodů mezi nástroji pamatuje, ale vždy aktivně zpracovává vstup pouze aktuální kontext.
- Lokální ikony, horní ikony, `aria-pressed`, hover stav a klávesové zkratky čtou **stejný stavový model**. Nikdy nedržet dvě kopie `active` v rozdílných handlerech.

### 3.3 Vstup a modifikátory

- **Levý klik/tažení:** výhradně aktuální lokální režim. V textu `caret`: single-click caret do textu / nový bodový text mimo text. Drag v textu vybírá znaky, nikoli objekty. `text-rect` tažením vybírá oblast textu; `text-box` tažením tvoří rámec.
- **Laso:** samostatný handler dle kontextu, vrcholy levými kliky, dokončení Enter/dvojklik, zrušení Esc. Křivka lasa nesmí být sama vložena jako PDF grafický objekt.
- **Prostřední tlačítko + drag:** zachovat existující schválené globální plošné gesto podle aktivního kontextu; nesmí paralelně vytvářet textové rámečky nebo objekty. Prověřit konflikt s panorámováním.
- **Ctrl+klik/Shift+klik, dvojkliky a další kombinace:** respektovat přesné platné zkratky `Text v3.3` a `specifikace_Codex_v5.0`. Text a Objekty nesmějí interpretovat stejnou kombinaci současně. Miniatury `Ctrl` přidat/odebrat, `Shift` označit rozsah od kotvy.
- **Esc** ukončuje pouze nejbližší rozpracovanou operaci (modal/popup/laso/drag) a nezpůsobuje úpravu PDF. `Ctrl+Z/Y` spravuje centrální skutečné změny; hover ani změna výběru nespouští obsahový zápis.
- Implementace musí používat mapu precedence událostí a testy kolizí (např. `pointerdown` ve formuláři / v textovém rámci / na objektu / v miniatuře / na horní liště).

## 4. Nové nástroje Text → Typografie → Křivky a rekonstrukce textu

### 4.1 Text → Převést na křivky (realizovatelná produkční funkce)

- Aktivovat při výběru skutečných textových objektů nebo textových úseků. Respektovat původní *glyph shaping*, ligatury, velikost, otočení/skew/transformace, kerning, barevný prostor, průhlednost a pořadí kreslení.
- Převod musí vytvořit **skutečné vektorové cesty**. Zákaz screenshot/rasterizačního obchvatu. U smíšeného výběru převádět jen explicitně zvolené úseky, ostatní části nezměnit.
- Má-li být převod vratný, původní textové informace (znaky, font, styl, transformace) uložit především do **pracovní historie/bezpečného zdroje**, nikoli bez upozornění tajně přidávat viditelně neprojevovaný text do výstupního PDF. Alternativa pro následnou extrakci je přístupný text s `ActualText` / neviditelnou textovou vrstvou; je to však **samostatná exportní volba** s dopady na strukturu, vyhledávání a důvěrnost, nikoli zaručená ztrátově vratná editace.
- Undo/Redo: jediná transakce, která obnoví původní textové objekty a jejich atributy. Po běžném novém otevření čistě obrysového PDF nelze spoléhat na Undo předchozí relace.
- Při použití fontu v podmnožině testovat dostupnost glyfů, clipping, průhlednost, barevné separace, text v XObjects a pozor na podpisy.

### 4.2 Křivky → Rekonstruovat text (experimentální výzkumná funkce)

Nejde o matematicky jednoznačný inverzní převod. Vnější PDF může obsahovat libovolné cesty, symboly, grafiku, fragmenty písmen, ligatury nebo text složený z mnoha křivek. Z obrysu samotného nelze v každém případě zjistit přesné Unicode, font, původní velikost ani řádkování.

**Pořadí rekonstrukce (od nejspolehlivější po nejobtížnější):**

1. **Dostupná původní typografická informace:** je-li text stále textovým objektem, použít původní metadata, font, `ToUnicode` nebo interní nepoškozený pracovní záznam. OCR nepouštět zbytečně.
2. **Křivky s textovou vrstvou/metadaty:** číst existující relevantní Unicode (`ActualText`, případně text z neviditelné vrstvy), ověřit mapování na vizuální obrysy; identita původního fontu z toho sama neplyne.
3. **Čistě vektorové křivky:** nejprve segmentovat kandidátní obrysy na glyfy, řádky a odstavce z geometrie a řádkové osnovy. Rozlišovat tečky, háčky, vnitřní otvory, ligatury, překrývání, otočený a deformovaný text; grafiku a loga nebrat za text bez kontroly. Zkusit rozpoznávání **přímo z obrysových dat** a porovnání s tvary fontů.
4. **Fallback na OCR:** lokálně vyrenderovat pouze vybranou oblast ve vysokém rozlišení a zavolat OCR. Vrátit text, bbox, confidence; obrysová předloha zůstane nedotčena. OCR výsledek není důkaz původní fontové rodiny.
5. **Identifikace písma:** nejprve zkontrolovat fonty dostupné v PDF/na disku uživatele (s příslušným právem použití), pak lokálně dostupné open-fonty. U každého kandidáta vyrenderovat rozpoznaný text; porovnat kontury/glyph width, rozteče, sklon, řez, osu, charakteristické znaky a další metriky. Vybrat **TOP N kandidátů** se skóre podobnosti, **nikoli jedno údajně jisté jméno**.
6. **Rekonstrukce sazby:** určit posuny v PDF souřadnicích, účaří, relativní velikost písma, vodorovné měřítko, proklad, mezery mezi slovy, výšku řádků, zarovnání, řádkový zlom a případný víceřádkový textový rámec. Pro známý font a vektorové cesty použít fyzické metriky fontu, ne odhad pouze podle výšky bbox.
7. **Interaktivní kontrola:** v UI vedle sebe ukázat *původní křivky* a *editovatelnou rekonstrukci*, překryv/difference map, volbu kandidátů fontů, opravu rozpoznaného textu a jednotlivých metrík, confidence/nejistoty. **Původní křivky zachovat až do výslovného potvrzení**; nabídnout vedle „Nahradit“ i „Ponechat obě vrstvy“ a „Zrušit“.
8. **Potvrzené změny:** jedna centrální transakce a AutoRecover. Chybný OCR výsledek ani rozpoznání podle podobného fontu nesmí přesunout/zdeformovat nevybrané PDF objekty.

**UX v HTML v3.5:** jeden rozbalovací oddíl `Křivky a rekonstrukce textu`, dvě ikony, `?` se souvislým vysvětlením, dvě volby zachování zdroje, srovnávací panel s výslovně demonstračním stavem. **V HTML nejsou spuštěny žádné OCR ani font matcher procesy** a nejsou simulovány skutečně nalezené fonty či výsledky rozpoznávání.

## 5. Relevantní open-source základy — stav průzkumu k 23. 11. 2026

| Projekt | Co umí a jak využít | Licence / omezení | Vztah k Blacksmith |
|---|---|---|---|
| [PDFium](https://pdfium.googlesource.com/pdfium/+/main/public/fpdf_edit.h) | Stávající PDF backend, textové objekty a stránkové PATH objekty, segmenty cest a maticové transformace. | BSD-style; kontrola použitých buildů/third-party komponent. | **Základ implementace**, nepřidávat další PDF parser bez důvodu. |
| [fontTools](https://github.com/fonttools/fonttools) | Rozbalování TTF/OTF, glyph outlines, fontové metriky a křivky; Python referenční matcher. | MIT | **Vysoká priorita** pro prototyp rozpoznání a porovnání fontů. |
| [Tesseract OCR](https://github.com/tesseract-ocr/tesseract) | OCR, bbox + confidence; model pro češtinu `ces`; není PDF renderer, nejprve vykreslit oblast. | Apache-2.0 | **Vysoká priorita** jako robustní offline fallback. |
| [TrueTypeTransformer](https://github.com/uchidalab/TrueTypeTransformer) | Výzkumný model klasifikace znaků a stylu podle křivkového TrueType obrysu. | MIT; výzkumné řešení, nutno otestovat na cestách extrahovaných z reálného PDF | **Nejslibnější cesta** pro přímé vektorové rozpoznávání. |
| [pdf-outline-fonts](https://github.com/audreyt/pdf-outline-fonts) | Převod některých PDF fontů na paths a volitelně neviditelná vrstva textu. | CC0; aktuálně omezené na CFF/Type-0 subsets, ne na všechny fonty | **Studie převodu a zachování textové sémantiky**, ne hotový univerzální konvertor. |
| [pdf-edit-engine](https://github.com/AryanBV/pdf-edit-engine) | Úprava existujících textových PDF operátorů při zachování fontu, subsetu, šířek/pozic. | MIT; neřeší obnovu čistých křivek | Inspirace pro bezpečnou editaci **nativního** textu a varianty fontových subsetů. |
| [PDF Lks Text Editor](https://github.com/Leksa667/PDF_Lks_Text_Editor) | OCR skenů a vlastní font matching: rasterizace kandidátních fontů a geometrické korelace. | Nestandardně formulovaná autorská licence („free ... provided credit“); právně prověřit před převzetím. | **Přímá inspirace algoritmem**, nikoliv přebírání kódu bez licenční analýzy. |
| [Mixfont Lens](https://github.com/mixfont/lens) | Model odhadu podobného open-source fontu z obrázku. | **Výslovně jen personal/academic/non-commercial** | **NEINTEGROVAT** do případného komerčního Blacksmith bez individuálního licenčního souhlasu. |
| [OCRmyPDF](https://github.com/ocrmypdf/OCRmyPDF) | Přidání OCR vrstvy k dokumentu, převážně vyhledávatelnost a kopírování. | MPL-2.0; nástroje v pipeline mají vlastní licence | Vzor OCR vrstvy, **nikoli** řešení identické grafické rekonstrukce a editovatelnosti. |
| [pikepdf](https://github.com/pikepdf/pikepdf) | Čtení a zápis PDF se strukturovanými streamy nad QPDF. | MPL-2.0 | Případně nástroj pro samostatné experimenty, ne automatické rozhodnutí o migraci backendu. |

Doplňkové diskuse: [Inkscape / GitLab — PDF import text vs paths](https://gitlab.com/inkscape/inbox/-/issues/7865), [Inkscape — převod PDF textu do SVG](https://gitlab.com/inkscape/inkscape/-/issues/254), [PDF Association — ActualText a přístupnost](https://pdfa.org/understanding-pdf-accessibility-techniques/). Přímý převod cesty → identický původní text a font jako bezchybnou obecnou funkci tyto zdroje neprokazují.

## 6. Implementační rozdělení pro Codex

**Etapa A — bezpečný UX základ (v3.5 prototyp je hotov):**
- sjednotit stav výběru v objektových a textových panelech / horní liště; implementovat vlastnictví gesture a zkratky; samostatné miniatury;
- pevné sloty nápovědy a `?` bez přeskoků;
- přidat příkazy a jejich kontextové povolení/zakázání podle reálného typu výběru.

**Etapa B — skutečný text → paths (po schválení):**
- vyčíst PDF font/glyfy a transformace, vektorově generovat cesty, bez rasterizace;
- transaction/undo, kontrola render diff, preflight, kompatibilita s exportem a PDF podpisem.

**Etapa C — R&D pilot cesty → text (samostatné schválení):**
- vytvořit testovací sadu PDF (Arial, Liberation Sans, Times, česká diakritika, smíšená typografie, ligatury, otočení, embedded font vs chybějící font, záměrné netextové křivky, víceřádkové odstavce);
- extrahovat skupiny path objektů; zkusit `fontTools` + porovnání vektoru, `TrueTypeTransformer`, a fallback `Tesseract ces`;
- měřit **CER/WER** rozpoznaného textu, shodu font-family/weight, rozdíl bbox/účar a vizuální render diff; výsledky rozdělit do skupin „přesné díky metadatům“, „dobře aproximované“, „nejisté“, „nerozpoznatelné“;
- až po výsledcích rozhodnout o implementaci v Go přímo / izolovaném Python workeru / on-device OCR. **Bez síťového odesílání dokumentů ve výchozím nastavení.**

## 7. Akceptační testy

- [ ] Hover přes každou skupinu ikon nemění geometrii navazujících sekcí (`Δy=0`) při širokém i úzkém panelu, klávesovém fokusování a DPI 100/150/200 %.
- [ ] Horní „kurzor/obdélník“ synchronizuje *tentýž* stav jako Text nebo Objekty; neaktivuje cizí handler. Přepnutí jiného nástroje na horní výběr explicitně přejde na Objekty.
- [ ] Textové laso a rámec mohou být aktivní, aniž by horní kurzor/obdélník podávaly falešné aktivní zvýraznění.
- [ ] Jediný drag neprovede dvojí výběr ani druhou změnu historie. Esc ukončí rozpracované gesto.
- [ ] `Ctrl+klik` miniatur neovlivní vybraný text; `Shift+klik` vybere jejich rozsah.
- [ ] Pro převod Text→Křivky zachovat rozměry, barvu, kreslicí pořadí a vektorovou kvalitu při 100–900% zoomu; render-diff na malých i velkých velikostech.
- [ ] Převod Text→Křivky jde jedním Undo vrátit, a nativní PDF text zůstane editovatelný po vrácení.
- [ ] Křivky→Text nepřisuzuje neprokázané fonty jako fakt; nabízí kandidáty, kontrolu a možnost žádnou změnu neprovést.
- [ ] Výsledek nejistého OCR nikdy tiše nenahradí původní křivky. Výstupní PDF ověřit, po potvrzení jediná Undo transakce.
- [x] HTML v3.5: testováno v Chromiu při 980, 1100, 1250 a 1600 px; bez runtime JavaScript chyby; hover nezměnil polohu následující sekce, horní lišta a panely synchronizovaly režimy, nové ovládací prvky typografie se otevřely, miniatury byly nezávislé. **Toto nenahrazuje produkční test skutečného PDF.**

## 8. Co dále ladit po v3.5

1. **Stavové interakce a kontextové povolení:** když je vybraný text vs skutečná path vs PDF kombinace, správně znepřístupnit nevhodné příkazy.
2. **Porovnávací dialog rekonstrukce:** režim vedle sebe/overlay, přepínač kandidátních fontů, ruční úpravy velikosti a prokladu, explicitní Nahradit/Ponechat obojí/Zrušit.
3. **Kvalita generovaného PDF:** zachování vektorových obrysů, tiskových barev a nepoškození podpisů, XObjects a neoznačených objektů.

MD je **specifikace a návrh výzkumu**, nikoli tvrzení, že byly nalezeny a úspěšně integrovány funkční moduly pro automatickou rekonstrukci.



### Archiv 3: `PDF_Blacksmith_Codex_UI_Text_v3.6_dialog_rekonstrukce.md`

# PDF Blacksmith — v3.6: dialog rekonstrukce, jednotné OCR a připravený přesun do Výběru

**Datum:** 23. 11. 2026  
**Navazuje na:** `PDF_Blacksmith_Codex_UI_Text_v3.5_vyber_krivky.md` a schválené chování UI předchozích verzí.  
**Stav:** HTML demonstrační návrh dialogu je hotový a odzkoušený v Chromiu. Není propojen s PDFium, OCR ani skutečným rozpoznáváním fontů. **Tento dokument není příkaz bez dalšího schválení přepsat produkční Go/Win32 kód.**

## 1. Pravidla schválená pro návrh

1. Obvyklý český název první funkce je **Převést na křivky** (*Convert to Outlines / Create Outlines*).
2. Druhá funkce je **Rekonstruovat text**; v kontextu obrázků se použije společný OCR backend.
3. **V této verzi obě ikony zůstávají v Text → Typografie → Křivky a rekonstrukce textu.** Jejich přesun do **Text → Výběr** je **výslovně odložen do dalšího kroku**. Nepřidávat dnes druhé paralelní ikony.
4. Po přesunu budou příkazy pracovat nad **aktuálním výběrem**, nikoli automaticky nad celou stránkou. Výběr může vzniknout kurzorem, úsekem v řádku, celým odstavcem, obdélníkem či lasem; vybraný obsah může být nativní text PDF, vektorové cesty, rastrový obrázek nebo jejich kombinace.
5. Později bude možné nabídnout **celou stránku a celý dokument** (hromadné zpracování s potvrzováním bloků), nikoli v tomto kroku.
6. **OCR je jedna sdílená služba** používaná jak v samostatném nástroji **OCR**, tak budoucím **Text → Výběr → Rekonstruovat text**. Nevytvářet dva různé algoritmy ani samostatné neuspořádané OCR historie.
7. Výsledek nikdy nesmí automaticky nahradit originální grafiku bez náhledu, potvrzení a možnosti Undo. U podepsaných dokumentů upozornit na invalidaci podpisu před změnou.

## 2. Co je nově v HTML v3.6

Kliknutí na **Rekonstruovat text** otevře samostatný dialog se dvěma sloupci: levý pro volbu zdroje, jazyka a metrik; pravý pro náhledy a úpravu rozpoznaného textu.

**Přepínače a interakce:**

- Výběr ukázkového zdroje **vektorové křivky / rastr / kombinace**. NENÍ automaticky detekováno skutečné PDF, jde o simulaci pro posouzení UI.
- Jazyk **čeština / angličtina / CZ+EN**. Volba je zatím vizuální, OCR neběží.
- Porovnání **vedle sebe** a **překrytí**, u překrytí posuvník průhlednosti upraveného textu. Náhled není algoritmicky počítaná *difference map*.
- **Přepsání rozpoznaného textu** v editoru, živá aktualizace obou pravých náhledů.
- Volba kandidátního písma: Arial, Liberation Sans, Georgia, Times New Roman. **Nejde o skutečně rozpoznané fonty ani vypočtené skóre**; výslovně to označit v UI.
- Parametry: **velikost, proklad, řádkování, horizontální šířka, posun účaří, otočení**. Změna se promítá do simulované rekonstrukce; po implementaci používat přesné souřadnice a fontové metriky PDF, nikoli CSS px jako PDF pt.
- Volba výsledku: **editovatelná kopie s ponecháním předlohy / původní vzhled + textová vrstva / nahradit vybraný vektorový obsah**. Poslední možnost je při rastru nebo smíšeném obsahu bezpečně zakázána.
- **Zrušit**, **Esc** a tlačítko **Potvrdit návrh (ukázka)**. Poslední akce pouze zavře okno a informuje, že nic v PDF nebylo provedeno. Dialog udržuje fokus a při zavření jej vrací původnímu prvku.
- Stávající pravý panel, historie, výběrové režimy a rezervované vysvětlovací sloty zůstávají zachovány.

### Bezpečnost rasterového zdroje

Při OCR rastrového podkladu text zůstává uvnitř obrázku. **Pouhé přidání editovatelného textového objektu tento původní obrazový text nesmaže.** Proto samostatně rozlišovat:

- **Vyhledatelná textová vrstva**: pozice a znaky, obvykle bez vizuální změny; slouží k hledání/kopírování. To není plnohodnotné grafické obnovení.
- **Editovatelný překryv**: nový viditelný textový objekt nad obrázkem; bez oddělené rekonstrukce pozadí může působit jako zdvojený tisk. Vždy nabídnout náhled a ruční maskování podkladu nebo výběrové odstranění původního obrazového textu.
- **Grafická náhrada**: potlačení původního textu a skutečná oprava pozadí obrázku. Je to náročnější samostatná funkce, nikoli výchozí OCR.

## 3. Další krok (v3.7) — přesun do Text → Výběr

Převést stávající dvojici ikon do kompaktní části **Úpravy vybraného textu** (nebo názvu vyhovujícího šířce panelu) v záložce Výběr, až po odsouhlasení v3.6:

- **Převést na křivky**: povolit pro skutečné vybrané textové objekty/textové fragmenty. Neprohlašovat obrázek za text; není-li žádný nativní text, nabídnout nejprve OCR/rekonstrukci.
- **Rekonstruovat text**: povolit, pokud výběr obsahuje path objekty, obrázky, obsah v ořezu nebo nerozpoznanou kombinaci. U už existujícího editovatelného PDF textu nejprve ověřit metadata a nabídnout běžnou editaci bez OCR.
- Ponechat stávající čtyři výběrové režimy **Textový kurzor / Obdélník / Laso / Odstavcový blok** v jedné řadě. Nové konverzní ikony nejsou pátým a šestým *výběrovým režimem*: jsou to **příkazy nad již vytvořeným výběrem**. Kliknutí na ně nesmí změnit samotný režim ani výběr.
- Vybraný řádek/úsek: zpracovat jen vybrané glyfy; odstavec: rozsah znakového nebo odstavcového výběru. Obdélník/laso: převést **jen skutečné objekty/části** spadající do výběru podle předem zdokumentovaných pravidel. Nesmí se náhodně oříznout půlka glyfu při převodu textu na křivky.
- `Esc` zruší rozpracovanou oblast nebo dialog, nikoli upravený dokument. Jeden pointer drag musí patřit jen jednomu handleru (`selectionUX`), žádné současné objektové/textové laso.

### Doporučený společný datový kontrakt (návrh, ne hotové API)

```text
SelectionScope {
  page_ids, document_id,
  mode: caret | span | paragraph | rectangle | polygon | whole_page | whole_document,
  text_runs: [font, transforms, glyph_range, Unicode?, bbox],
  paths: [path_id, clip, transforms, color],
  image_regions: [image_id, source_rect, page_transform, polygon_mask?],
  masks, readonly, signed_document_state
}
RecognitionResult {
  regions: [text, words, chars?, bbox, baseline, orientation, script,
            character_confidence?, source_kind],
  font_candidates: [font_id, verified_source, likeness_score?, availability],
  suggested_typography: [font_size, tracking, leading, width_scale,
                         transform, alignment],
  warnings, provenance
}
```

**Pozor:** confidence OCR znaků a podobnost fontů jsou různé skórovací veličiny; neslučovat je do jednoho domnělého „% přesnosti“.

## 4. Jedna pipeline pro OCR a Výběr

**Vstup:** nativní PDF text, PDF vektorové křivky, PDF image XObject nebo samostatné rastrové soubory (v navazující fázi import PNG/JPEG/TIFF/WebP podle knihoven dekodéru).

**Doporučené pořadí:**

1. Prohlédnout vybranou oblast v objektovém stromu / render streamu PDFium. Pokud lze přečíst zachovaný text, Unicode mapování a fontové údaje, **OCR přeskočit**.
2. Pro čisté **vektorové křivky** segmentovat obrysy a prostorové vztahy, zkusit rozpoznat znak z cest. Při nejistotě vyrenderovat pouze potřebný výřez do bitmapy a spustit OCR.
3. Pro **rastr** zvolit obrazové OCR. Respektovat rozlišení, orientaci, masku polygonu, průhlednost a souřadnicovou transformaci do PDF. Neprovádět zbytečně OCR celé stránky, pokud byl vybrán jen malý obdélník.
4. Pro **smíšené oblasti** provést rozpoznání po regionech, slučovat výsledky podle pozic a řádků; hlídat duplicity mezi neviditelnou textovou vrstvou a pixely nebo vektory téhož slova.
5. K dispozici-li odpovídající fontové metriky, použít je; jinak sestavit seřazené kandidáty podle tvarů glyphů (např. přes `fontTools`). Neztotožňovat OCR s identifikací fontu.
6. Převést rozpoznané řádky na skutečné editovatelné PDF textové objekty s zachovanými rotacemi, kerningem, baseline, bbox a odstavcovými pravidly. Pokud PDF neumí zvolený font legálně vložit, přiznat alternativu/substituci.
7. Ukázat **diff/review**, explicitně potvrdit; zápis až poté jako jedna transakce Undo/Redo. Zpracování lokálně; žádné neohlášené odesílání textu dokumentu do cloudu.

**Základní kandidáti komponent:** stávající **PDFium** pro práci s PDF objekty, **Tesseract OCR** (Apache 2.0) pro OCR z bitmap, volitelně **PaddleOCR** (Apache 2.0) pro složitější detekci layoutu a `fontTools` (MIT) pro analýzu dostupných fontů. PaddleOCR je větší závislost; pro Windows/Go zvolit oddělenou výkonnou službu až po benchmarku a právním ověření distribuce modelů. Žádný z těchto projektů **sám** negarantuje přesnou obnovu sazby z cizích křivek. Viz [Tesseract](https://github.com/tesseract-ocr/tesseract), [PaddleOCR](https://github.com/PaddlePaddle/PaddleOCR), [fontTools](https://fonttools.readthedocs.io/en/latest/pens/index.html).

## 5. Co považovat za úspěšnou rekonstrukci

**Tři oddělené výsledky:** (A) Unicode a pořadí znaků, (B) shoda písma a typografických vlastností, (C) vizuální/geometrická podobnost po zpětném vykreslení. Každou položku hodnotit zvlášť. Ruční oprava OCR musí mít přednost před výstupem automatického matcheru. Počet vybraných objektů a stránky mimo rozsah se nesmí změnit.

Návrh UI pro skutečnou verzi: v porovnání nabídnout **červený překryv/difference** založený na renderu, nikoli jen CSS opacity, a výstrahu, pokud nejsou metriky nebo některé znaky potvrzené. Pro větší výběry seznam problémových řádků; nejisté úseky kliknutím zvýraznit v náhledu.

## 6. Akceptační testy (HTML v3.6)

- [x] Otevření dialogu z **Text → Typografie → Křivky a rekonstrukce textu**; nic se prozatím nepřesunulo do záložky Výběr.
- [x] Přepnutí vedle sebe / překrytí, průhlednost překryvu a živá aktualizace textového náhledu.
- [x] Fontové kandidáty, velikost, jazyk a změnu vzorového zdroje lze ovládat bez JavaScriptové chyby.
- [x] Pro rastr/smíšený zdroj je nebezpečná volba „nahradit vybraný vektorový obsah“ zakázána a při změně typu zdroje se bezpečně přepne na textovou vrstvu.
- [x] `Esc` dialog zavře bez změny a `Potvrdit návrh (ukázka)` nespouští žádnou skutečnou editaci PDF.
- [x] Kontrola v Chromiu při 980 / 1100 / 1250 / 1600 px bez `pageerror`.

**Zatím neimplementováno, a tedy netestováno:** skutečný PDF výběr, rozpoznávání, font matching, ukládání textových objektů, přímá práce se samostatnými PNG/JPG/TIFF/WebP, export PDF, AutoRecover a Undo/Redo na skutečném dokumentu.



### Archiv 4: `PDF_Blacksmith_Codex_UI_Text_v3.7_vyber_sdilene_OCR_plan_voditek.md`

# PDF Blacksmith — UI/Text v3.7: převody nad výběrem, společné OCR a plán inteligentního přichytávání

**Datum:** 23. listopadu 2026  
**Vstupy:** HTML v3.6 a specifikace `PDF_Blacksmith_Codex_UI_Text_v3.6_dialog_rekonstrukce.md`, další schválené specifikace v3.3–v3.5.  
**Výstup návrhu:** `PDF_Blacksmith_HTML_navrh_v3.7_prevod_nad_vyberem_ocr.html`.  
**Status:** SCHVÁLENÝ NÁVRH UX V HTML, nikoli automaticky povolený zásah do produkční aplikace Go/Win32/PDFium. Produkční implementaci zahájit až po inventuře stávajícího kódu a schválení další etapy.

## 1. Co se mění právě teď (v3.7)

### 1.1 Umístění příkazů

- **Převést na křivky** (Convert to Outlines) a **Rekonstruovat text** jsou nově **pouze** v `Text → Výběr → Převody vybraného obsahu`. Odstranit jejich původní accordion `Text → Typografie → Křivky a rekonstrukce textu`. **Nezavádět duplicitní příkazy.**
- Jsou to **akce nad aktuálním výběrem**, ne nové režimy výběru. Stávající čtyři režimy (Textový kurzor / Obdélník / Laso / Odstavcový textový blok), jejich ikony a místní otazník se nemění.
- Oba příkazy nemění režim výběru, výběr samotný ani aktivní objekt před vykonáním. Pouhý hover, přepínání záložek, předvolby a otevření dialogu nevytvářejí obsahový krok Undo.
- Kompaktní, čistý panel: ikony, stálá rezervace výšky pro jejich vysvětlení (bez „poskakování“ následujících sekcí), dlouhá nápověda pod `?`.

### 1.2 Jeden model výběru pro několik druhů obsahu

| Vstupní výběr | Převést na křivky | Rekonstruovat text |
|---|---|---|
| Žádný obsah / samotný caret bez vybraných znaků | Zakázat | Zakázat |
| Nativní editovatelný úsek v řádku | Ano, jen označené glyfy | Obvykle zakázat, využít existující textové údaje |
| Nativní editovatelný odstavec | Ano, označené běhy | Obvykle zakázat |
| Obdélník obsahující nativní text | Ano, podle pravidel zahrnutí glyfů | Zakázat, pokud nejsou křivky či obraz |
| Laso obsahující nativní text | Ano, podle pravidel zahrnutí glyfů | Zakázat, pokud nejsou křivky či obraz |
| Vektorové obrysy připomínající text | Ne (již jde o křivky) | Ano, segmentace cest / vektorová analýza + případný OCR fallback |
| Rastrová oblast (PDF image XObject, vybraný obrázek) | Ne | Ano, obrazové OCR |
| Smíšená oblast obsahující nativní text, křivky i rastr | Ano, **jen nativní text** | Ano, **jen části vyžadující rekonstrukci** |

- V **obdélníku a lasu** musí algoritmus nejprve identifikovat objekty/glyfy v souřadnicích stránky, transformace, ořez a případné masky. **Nevykreslit polovinu glyfu jako samostatnou křivku jen proto, že na něj dopadla hranice výběru.** V první produkční verzi pracovat nad celými zasaženými glyfy dle jedné jasné geometrické definice a uživateli ji v `?` vysvětlit; pro výběr s částečnými zásahy upozornit a umožnit upravit oblast.
- Výběr textu se nesmí nenápadně přepnout na výběr objektů. Objektové Laso, textové Laso a výběr miniatur mají různé souřadnice i handlery. Jeden pointer gesture → právě jeden vlastník. Esc ruší rozpracované gesto nebo dialog, ne obsah jiné oblasti.
- V HTML v3.7 je pouze **viditelný simulátor** osmi scénářů výběru. Přepnutí simulátoru není skutečným kliknutím do PDF. Do produkčního UI tento testovací `<select>` **nepřenášet**; nahradit jej živým `SelectionScope`.

### 1.3 Rozpoznávání ve Výběru i v OCR = jedna služba

**Společný kontrakt** (návrh, přesné typy až po inventuře engine):

```text
SelectionScope {
 document_id, page_ids, source_revision, mode,
 selected_text_glyphs[], selected_paths[], selected_image_regions[],
 selection_polygon?, clipping_paths[], effective_transforms[],
 editability, security_flags, signed_document_state
}
RecognitionRequest {
 origin: text_selection | ocr_tool,
 scope: SelectionScope | page_range,
 language: ces | eng | ces+eng,
 output: text_layer | editable_copy | replace_vector_selection,
 local_only: true
}
RecognitionRegion {
 recognized_text, glyph_boxes[], word_boxes[], line_baselines[], paragraph_regions[],
 source: native_pdf_text | path_analysis | raster_ocr,
 text_confidence?, font_candidates[], typography_estimate?, geometry_transform,
 warnings[], provenance
}
```

**Processing:**

1. Nejprve zjistit, zda má oblast nativní text s validním Unicode/CMap. Takové údaje použít přímo, nerozpoznávat podruhé.
2. Obrysy křivek analyzovat jako vektory; nejisté úseky lze převést lokálně do obrazového výřezu a zpracovat OCR.
3. Obrázky (z PDF i samostatně naimportované soubory, v další produkční etapě např. PNG/JPEG/TIFF/WebP) zpracovat OCR, vždy jen v cílovém regionu. Podporovat orientaci, DPI, průhlednost, masky a transformaci z bitmapových souřadnic do stránky.
4. Smíšené oblasti rozdělit na regiony, potom spojit řádky a odstavce s kontrolou duplicit existující textové vrstvy.
5. Samostatně odhadnout font, velikost, proklad, řádkování, účaří, otočení a zalamování. Důvěra v znaky OCR ≠ podobnost fontu.
6. Vrátit návrh do **jednoho revizního dialogu** z v3.6: zdroj, typ regionu, porovnání vedle sebe/překryvem, kandidátní fonty, korektury a bezpečná volba výsledku. **Nic do PDF nezapisovat do potvrzení.**
7. Při potvrzení zápis jako jedna transakce Undo/Redo se zachováním podkladu pro Zpět, AutoRecover a ochrany podepsaných/uzamčených dokumentů. Pokud editace invaliduje digitální podpis, zobrazit výslovné varování před změnou.

Samostatný nástroj **OCR** zůstává dostupný pro oblast nebo rozsahy stránek. `Text → Výběr → Rekonstruovat text` zpracuje přesně vybraný obsah. Sdílejí stejné jádro, jazykové nastavení a návratový datový model; rozdíl je **pouze v definici vstupu a navrženém výsledku**. V HTML je obojí demonstrováno jedním dialogem a společnou předvolbou jazyka; skutečná OCR služba ještě připojena není.

**Technologické možnosti k posouzení při implementaci:** PDFium pro PDF objekty a vykreslení; Tesseract (Apache 2.0) jako lokální OCR baseline; případně PaddleOCR (Apache 2.0) pro layout/rotace; fontTools (MIT) pro metriky a obrysy fontů. Před zařazením binárek/modelů prověřit licenční podmínky, distribuci Windows/Go, velikost, aktualizace a benchmark na češtině. Žádný z nástrojů sám nezaručuje přesnou rekonstrukci neznámého původního fontu.

### 1.4 Povinná bezpečnost rekonstruovaného rastru

- **Vyhledatelná textová vrstva** je textová informace k obrázku bez změny původního vzhledu. Není to vizuální odstranění textu ze snímku.
- **Editovatelná kopie** nad rastrem vytvoří další viditelný text; dokud není původní rastrový text výslovně zakryt/odstraněn bezpečným postupem, může být zdvojený. V dialogu srozumitelně upozornit.
- **Nahradit vybraný vektorový obsah** je zakázáno pro rastr a smíšené zdroje; před povolením pro vektor je nutný jednoznačný, explicitně potvrzený výběr.
- Originální font může být nenalezitelný, text chybně dekódován, chybět glyph v náhradním písmu nebo být složitý layout; nikdy neoznačovat odhad za prokázanou totožnost.
- Bez souhlasu neodesílat obsah dokumentů na cloudové API; lokální rozpoznávání má být výchozí.

## 2. Co je v HTML v3.7 skutečně interaktivní

- Přesun obou ikon do Text → Výběr; Typografie již konverze neobsahuje.
- Osm **výslovně simulovaných** variant výběru (bez obsahu, řádek, odstavec, text v obdélníku, text v lasu, křivky, rastr, smíšený obsah) a dynamická dostupnost obou akcí.
- Příkaz „Převést na křivky“ zobrazí vysvětlující workflow (bez změny PDF).
- „Rekonstruovat text“ otevře existující dialog, výchozí typ zdroje a rozsah odpovídá vybranému **simulovanému** obsahu. V dialogu lze nadalé zdroj přepnout jen pro testování UI.
- Samostatný nástroj OCR otevírá **tentýž návrhový dialog** pro ukázkový rozsah. Volba jazyka z OCR i z dialogu aktualizuje sdílenou předvolbu.
- Zůstává porovnání vedle sebe/překryvem, oprava znaků, typografické parametry, kandidátní fonty a bezpečná volba výstupu.
- Zůstává rozdělený výběr text/objekty/miniatury, žádná druhá sada výběrových handlerů. Zůstávají rezervované sloty nápovědy, čisté ikony a responzivní horní lišta.

**V HTML není implementováno:** skutečná extrakce/selection PDF, převod glyfů na vektorové cesty, OCR, přímá úprava PDF, převod celého dokumentu, test skutečného AutoRecover ani souřadnicová analýza obrysů. Změny a výběry jsou demonstrační.

## 3. Následující samostatná etapa — inteligentní přichytávání, vodítka a mřížka (PLÁN pro Codex)

Toto je **nově schválený rozsah do PLÁNU**, nikoli implementace ve v3.7 HTML. Implementovat až po zavedení jednotné editační geometrie a výběru.

### 3.1 Rozsah

- Společná služba `SnapEngine` pro Text, Objekty, transformace i (kde dává smysl) úpravu textového rámce. Žádné paralelní přichytávací handlery.
- Cíle: **rohy, okraje, hrany, středy objektů, středy jednotlivých stran, transformační body a uzly křivek**; u textu **účaří/baseline, začátky/konce řádků, hranice odstavcového rámce**; také okraje, střed a spadávka stránky (pokud je definována).
- **Inteligentní vodítka** při přesunu, zarovnání a změně velikosti: dočasné osy, shodné rozestupy a pomocné vzdálenosti, vizuálně přímo nad stránkou, bez zapsání do PDF.
- **Uživatelská vodítka** vodorovná/svislá (z pravítek i přesným zadáním), uzamčení, skrytí, přesun, vymazání, uložení rozvržení na úrovni dokumentu podle rozhodnutí o metadatech.
- **Mřížka** s vlastní roztečí, počátkem, jednotkami a stylem; nezávislé přepínače „zobrazit mřížku“ / „přichytávat k mřížce“. Případná podmřížka.
- Zapínat/vypínat každou skupinu snap-cílů samostatně. Centrální ovládání v `Zobrazení`, jemnější nastavení v `Nastavení`; v panelech Text/Objekty jen relevantní přepínač.
- Tolerance a hystereze v **obrazových pixelech** (DPI-aware), přepočítaná do souřadnic stránky podle zoomu a transformací. Měnit polohu objektu v dokumentových souřadnicích, nikoli podle zaokrouhlení na pixely.
- Při shodě více cílů deterministická priorita a vizuální indikace zvoleného cíle; zabránit „cukání“ mezi blízkými body. Dočasná klávesová deaktivace až po inventuře stávajících zkratek; nehádat volnou zkratku.
- **Zarovnání účaří mezi různými fonty a velikostmi** má fungovat geometricky podle skutečných baseline, nikoli pouze podle bounding-boxu.
- Vodítka a mřížka se **standardně netisknou ani neexportují**; nesmí měnit obsah PDF nebo digitální podpis. Obsahová transformace potvrzená snapem je jeden krok centrální historie, pohyb kursorem a zobrazení vodítka není obsahová editace.

### 3.2 Doporučené pořadí realizace

1. Inventura stávajících transformací a systémů souřadnic: PDF user space (bottom-left), Win32 client space (top-left), zoom, rotace, crop/media/trim/bleed box, transformační matice textu a objektů.
2. Společný `SelectionScope`, jednotný pointer gesture dispatcher a protokol `GeometryTarget`/`SnapCandidate` s identitou stránky i objektu.
3. Přichytávání na okraje a středy objektů + vizuální diagnostika; testy při rotaci, přiblížení a HiDPI.
4. Přidat baseline textu, vodítka, pravítka, zamykání a uložení přehledných metadat.
5. Přidat mřížku, nastavení tolerance/priorit a rovnoměrné rozestupy; výkonnostní benchmark na hustých technických PDF.
6. Ověřit kolize zkratek, Undo/Redo, interakce s výběrem a podpisy. Až poté nabídnout další pokročilé snap cíle.

**Akceptace této budoucí etapy:** přesunutí textového bloku zarovná baseline se sousedním řádkem, různé velikosti písma fungují; rohové/středové snap body jsou správné při libovolném zoomu a otočení; příslušná vodítka ani mřížka se neexportují do PDF; po vypnutí jednotlivých kategorií se nespouštějí; výběr textu/lasa se neztratí.

## 4. Další, zatím neimplementované rozšíření: celý dokument

Hromadný převod **celé stránky nebo celého dokumentu** je samostatná etapa: rozsah, odhad počtu objektů a času, přerušení/zrušení operace, přehled bloků nejistého OCR, explicitní potvrzení náhrad, postupné zpracování bez blokování UI a jedna promyšlená transakční historie. **Nedělat zcela automatickou globální náhradu při prvním použití.**

## 5. Akceptační testy HTML v3.7

- [x] V Text → Výběr existuje právě jedna dvojice příkazů, v Typografii ani jinde není duplikát.
- [x] Všechny čtyři původní režimy výběru zůstávají na jediném řádku a jsou nezávislé na akcích převodu.
- [x] Osm testovacích scénářů správně zapíná/vypíná jednotlivé akce.
- [x] Laso rastru otevírá dialog s rastrovým zdrojem; obdélník křivek s vektorovým; smíšený výběr s kombinací.
- [x] Přímá náhrada křivek je pro rastr a kombinaci nedostupná.
- [x] OCR a Text používají jediný návrhový dialog a společnou předvolbu jazyka.
- [x] Neexistuje obsahová změna PDF ani zdvojená instance výběrového handleru.
- [x] Interakční testy prošly v Chromiu na šířce 980 a 1250 px bez `pageerror`; statická responzivita horní lišty zachovává všech šest nabídek.
- [ ] Reálné OCR, font matching, PDF zápis, Undo/Redo nad dokumentem a přichytávání: **plánované, neotestované**.

**Pro Codex důležité:** neodstraňovat existující bezpečnostní nebo výběrová pravidla při implementaci nového UI. Nezaměňovat HTML ukázku se skutečným PDF editorem. Každý produkční zásah musí začít inventurou souvisejících modulů a regresními testy.



### Archiv 5: `PDF_Blacksmith_Codex_UI_v3.8_univerzalni_vyber_obrazky_snap.md`

# PDF Blacksmith – návrh UI a implementační plán v3.8

**Datum:** 24. listopadu 2026  
**Výchozí verze:** `PDF_Blacksmith_HTML_navrh_v3.7_prevod_nad_vyberem_ocr.html` + `PDF_Blacksmith_Codex_UI_Text_v3.7_vyber_sdilene_OCR_plan_voditek.md`  
**Nový návrh:** `PDF_Blacksmith_HTML_navrh_v3.8_univerzalni_vyber_obrazky_prichytavani.html`  
**Status:** **UX prototyp a implementační specifikace. NENÍ to příkaz bez inventury a zvláštního schválení přepisovat produkční Go/Win32/PDFium.** HTML nepřevádí skutečné PDF, neprovádí OCR ani neupravuje rastrové soubory na disku.

## 1. Schválené vizuální opravy

### 1.1 Typografické tlačítkové ikony

- Odstranit diagonální dvoubarevnou výplň pro smíšené/neurčité `aria-pressed="mixed"` (`R`, **B**, *I* a další přepínače). Běžný, zapnutý a neurčitý stav používá **stejný jazyk tlačítek** jako zbytek aplikace: tmavá plocha, tenká kontura, aktivní vínové zvýraznění. Neurčitost lze ukázat pouze nenápadnou tečkou v rohu + popisem dostupným v nápovědě a stavem přístupnosti. Nesmí být zaměňována s vypnutou vlastností.
- Smíšené formátování zachovat na datové úrovni: klik sjednotí **jen konkrétní vlastnost**, ostatní atributy zůstanou odlišné, historie a zpětná obnova původních běhů podle v3.3.

### 1.2 Dva odlišné typy nápovědy

- **Plovoucí tooltip** je krátký a objevuje se **nad ikonou**, nikoliv přes stálý popis pod ikonami. Implementace v HTML používá jeden sdílený `#iconTooltip`, pozicovaný do viewportu, bez šipkového „skákání“ panelu. Tooltip má být dostupný i při `focus` z klávesnice; nesmí měnit aktivní nástroj ani historii.
- **Popis na podkladu panelu** je v předem rezervovaném slotu pod ikonami a mění se při postupném přejíždění myší. Při dlouhém textu se posouvá **uvnitř slotu** místo přesouvání následujících sekcí. Dlouhá nápověda zůstává pod `?`.
- V produkci sjednotit přes jednu komponentu `ToolIconGroup + HintSlot` a jedinou funkci pro tooltip. Nevytvářet vlastní tooltip pro každý nástroj zvlášť. `title` nepoužívat paralelně s vlastním tooltipem (dvojité překrytí).

### 1.3 Pravá strana a zoom při změně okna

- Levý i pravý panel **odebírají prostor** z plochy editoru (jsou flex sourozenci). Ouška mohou mírně přesahovat do editoru, ale editor počítá jejich skutečný průnik jako ochranný okraj.
- Při běžném automatickém režimu se stránka zmenšuje dle reálně dostupné šířky tak, aby žádná její část neležela skrytá pod pravým panelem. HTML omezuje automatické zvětšování na 100 %, minimální zoom zůstává 30 %.
- Ručně zadaný zoom není bez souhlasu přepisován; pokud je stránka větší než pracovní plocha, zůstává úplně dosažitelná **horizontálním posuvníkem**. Režimy „Přizpůsobit stránce“ a „Přizpůsobit šířce“ se přepočítají i při změně velikosti okna a otevření panelu.
- Testovat standardní, levé/pravé pomocné panely a sbalené stavy při CSS šířkách 980 / 1100 / 1250 / 1600 px i při 100 / 150 / 200 % DPI. **Není přípustné opravit to jen pevným offsetem.**

## 2. Jedna výběrová architektura pro horní lištu, Text, Objekty i OCR

### 2.1 Nové ovládání druhé horní lišty

**Rozdělené tlačítko kurzoru** má hlavní spouštěcí část a vedlejší šipku pro volbu:

1. **Šipka / automatický výběr** (`i-cursor`) – předvolené. Klik do textového objektu aktivuje `Text`; klik do grafického nebo rastrového objektu aktivuje `Objekty`. Klik na prázdné místo neupravuje obsah. Vícenásobný výběr je dovolen.
2. **Textový kurzor** (`i-text_caret`) – vědomé upřednostnění editace textu; klik v textu umístí caret a klik mimo existující text v odpovídajícím režimu zahajuje bodový řádkový text podle v3.3. Přepnutí v nabídce nastaví hlavní ikonu tlačítka; naposledy použitá volba se v rámci relace zachová.

Vedle jsou **obdélník** (`i-rect_select`) a **laso** (`i-lasso`) – **jednotné plošné režimy**. Při dokončení gesta se na základě skutečné geometrie určí, co obsahují; režim výběru **nepředurčuje obsah**. Čtyři textové příkazy v `Text → Výběr` zůstávají, včetně `Odstavcový textový blok`, který **není univerzálním výběrem**. Horní lišta a vlastnosti sdílejí stav; kliknutí nesmí založit druhý souběžný handler.

| Výsledek hit-testu | Výsledek UI | Zachování výběru |
| --- | --- | --- |
| Nativní editovatelný text | Aktivovat panel **Text → Výběr** | Vybrané znaky/běhy zůstávají označené |
| Obraz / netextový vektor | Aktivovat **Objekty** | Zachovat vybrané objekty, obrázek zpřístupní Úpravy obrázku |
| Nativní text + obrázky/křivky | Aktivovat **Objekty** jako inspektor smíšeného obsahu; Text/OCR zůstávají dostupné | Nesmí se odhodit žádná část výběru; akce se filtrují podle typu obsahu |
| Prázdný výběr | Neupravovat dokument; zrušit jen rozpracované gesto | Žádné prázdné obsahové objekty |
| OCR s aktivní oblastí | Zůstat v **OCR** do zahájení/dokončení rozpoznání | Zachovat definici oblasti a neměnit vlastníka gesta |

**Pravidlo vlastnictví gest:** `PointerDown` patří právě jednomu dispatcheru; obdélník/laso nevytváří zároveň výběr textu i objektů. Pravidla miniatur stránek a dosavadní prostřední tlačítko (plošný výběr) se nesmějí přepsat. `Esc` ruší gesto/laso; `Enter` nebo dvojklik dokončuje polygon; klávesové modifikátory respektují dřívější SEL/TXT pravidla a vyžadují inventuru kolizí.

### 2.2 Detekce skutečného PDF

Produkční `SelectionScope` musí používat PDF user-space, transformační matice, `CropBox`, rotaci stránky, textové matice, masky obrázků, `clip`, renderovací vrstvy a překryvy; nelze jej odvozovat z DOM demonstrační stránky. Hit-test nativního textu nejdřív použije existující Unicode/glyph informace, u dalších cest se rozlišuje text převedený na křivky, běžná grafika a bitmapa. Nelze 100% automaticky určit, že každá křivka připomínající písmeno je text. Pro nejednoznačný obsah použít bezpečný smíšený výběr + nabídku ruční korekce. Překryté objekty rozlišovat pomocí pořadí/z-indexu a explicitního procházení, ne náhodně.

### 2.3 OCR – výběr v samostatném panelu

- Přidat identické tři ovládací prvky do `OCR`: kurzor, obdélník, laso (jako **ovládací zrcadlo**, nikoliv kopii implementace). Sjednotit jazyk OCR s `Text → Výběr → Rekonstruovat text` podle v3.7.
- Výběr může obsahovat původní PDF text, křivky, rastrový obrázek i kombinaci. Výstup **jediné** rozpoznávací služby; neprovádět duplicitní OCR validní textové vrstvy. OCR panel může být zachován během kreslení oblasti, i když uvnitř rozpozná text.
- OCR nad celými stránkami zůstává samostatný přepínač. Obdélník/lasa pouze definují oblast; nenahrazují volbu „celé stránky“.
- V HTML jsou aktivní pouze ilustrativní DOM hit-testy nad textovým vzorkem a obrazem. Není připojen PDF parser ani OCR model.

## 3. Úpravy vybraného rastrového obrázku

V `Objekty → Úpravy vybraného obrázku` se editor nabídne **jen při označení rastrového objektu**. Bez něj je vidět krátká výzva a demonstrační příkaz pro výběr vzorku. Nezpřístupňovat převzorkování pro libovolné vektorové objekty.

| Ovládání | Produkční význam / požadavek |
| --- | --- |
| Šířka a výška v pixelech | Skutečná změna rozměrů bitmapy, volitelně zachovat proporce; validace maxima a paměti |
| Rozlišení (ppi) | Samostatná metadatová/umísťovací hodnota; **pouhá změna PPI není převzorkování**; rozlišit pixely a fyzickou velikost objektu v PDF |
| Převzorkování | Lanczos, bikubické, bilineární, nejbližší soused. Náhled a upozornění při upscale/downscale a možné ztrátě detailu |
| Barevný režim | RGB, stupně šedi, 1bit černobílý, CMYK; při profilu ICC provést řízený převod a explicitně uvést zdrojový/cílový profil |
| Ořez | Výřez myší přímo na stránce + numerické souřadnice, možnost upravit úchyty. Vybraný obraz musí zůstat v původních geometrických souřadnicích stránky nebo se změna umístění musí předem potvrdit |
| Potvrzení, Zpět a obnovení | Nedestruktivní pracovní kopie a jedna transakce Undo/Redo. Počítat s AutoRecover, zámky a digitálními podpisy |

**Zásadní technická omezení:** rastrový obraz může být PDF Image XObject s alpha/soft maskou, indexed color space, ICCBased/CalRGB, JBIG2, JPEG2000, DCT, CCITT či být sdílen více stránkami. Editovat **vybraný výskyt**, nikoli automaticky všechny odkazy na tentýž zdroj. U soft-mask, přímých barev nebo profilů s nejednoznačným převodem nabídnout bezpečný dialog a případně funkci zakázat. Šedá ≠ automaticky pouze tiskové K; pro tiskové separace zůstává samostatná konverzní logika. Vždy prověřit obsah s podpisem či ochranou práv.

**HTML demonstrace:** parametry lze měnit, pro převod do šedi/mono/CMYK je pouze vizuální aproximace pomocí CSS, „Oříznout“ simuluje změnu výřezu SVG náhledu; **nevzniká bitmapa ani PDF soubor**. Reálné funkce až v backendu. U náročných rastrů se musí náhled generovat asynchronně ve skutečné aplikaci; žádné blokování UI.

## 4. Přichytávání – nové ikony a zobrazení (implementace magnetické geometrie zůstává v plánu)

### 4.1 Mapa ikon

| Interní symbol | Umístění | Akce v HTML |
| --- | --- | --- |
| `i-snap_magnet` | Druhá horní lišta + Zobrazení | Přepnout `snap` a `aria-pressed`; **zatím neprovádí geometrii** |
| `i-toggle_grid` | Druhá horní lišta + Zobrazení | Viditelná netisková mřížka na stránce |
| `i-toggle_rulers` | Druhá horní lišta + Zobrazení | Viditelná netisková pravítka |
| `i-image_edit` | Objekty | Panel parametrů rastrového obrázku |
| `i-image_crop` | Objekty | Numerický ořez / náhled |
| `i-cursor`, `i-text_caret`, `i-rect_select`, `i-lasso` | Horní lišta + kontextové panely | Sdílené výběrové ikony, původní symboly beze změny |

**SVG:** symboly jsou jako vektorové křivky přímo v mapě ikon (inline SVG `<defs>` v HTML), navíc v souboru `PDF_Blacksmith_ikony_v3.8.svg`. Neměnit stávající vazby ani sémantiku stávajících symbolů. Nová tlačítka musí mít tooltip, `aria-label`, `aria-pressed` a klávesnicové ovládání.

### 4.2 Nabídka Zobrazení → Přichytit

Nabídka obsahuje `Zobrazit mřížku`, `Zobrazit pravítka`, `Zobrazit vodítka` a hlavní přepínač `Přichytávání zapnuto`. Pod **Přichytit – podrobnosti** jsou samostatně zaškrtávací skupiny:

- Hrany objektů; středy objektů; středy stran; uzly křivek.
- Účaří textu (baseline); vodicí čáry; mřížka.
- Okraje/střed stránky; shodné rozestupy.

**Nedělit** „zobrazení mřížky“ a „přichytit k mřížce“ jako totožný přepínač. Globální magnet nesmí při vypnutí smazat jednotlivá nastavení cílů. Nynější HTML jen zobrazuje overlay mřížky a pravítek a udržuje UI přepínače v paměti stránky; vlastní editace geometrie, vytahování vodítek z pravítek, definice rozteče mřížky a per-document persistence jsou **další implementační etapa**. Do exportovaného PDF se pomůcky nedostávají.

### 4.3 Pořadí do Codex roadmapy

1. Inventura stávajících výběrových handlerů, PDFium objektového modelu a režimů OCR, včetně mouse/keyboard gest; teprve pak společný `SelectionScope` a `PointerGestureDispatcher`.
2. Zavedení dvojí volby hlavního kurzoru a jednotného obdélníku/lasa; nejprve text/objekt, potom smíšený výběr bez ztráty dalších obsahových typů.
3. Opravy rozvržení a nápověd; testy nekolizního chování s formuláři, panely a náhledy stránek.
4. Minimální obrazový editor: detekce XObject, výběr výskytu, náhled ořezu, pixelové převzorkování, grayscale a profilované RGB/CMYK (ICC), s vratnou úpravou.
5. Společná geometrická služba `SnapEngine` až nad spolehlivým výběrem: target index, tolerance/hystereze v pixelech, hierarchie cílů, inteligentní vodítka, překryvné indikace. Následně uživatelská pravítka, vodítka, mřížka, baseline a rovnoměrné rozestupy. Nezávislé přepínače, dokumentové ukládání nastavení a zabezpečení.
6. Integrace s OCR/reconstruction a příkazy „Převést na křivky“ / „Rekonstruovat text“ z v3.7: respektovat aktuální výběr a vrstvy, nepřidávat duplicitní funkce. Teprve později dávkové zpracování celého dokumentu.

## 5. Kontrolní scénáře

- [x] Smíšené typografické tlačítko má solidní pozadí, bez diagonálního gradientu.
- [x] Tooltip ikony je nad tlačítkem a neskryje pevně rezervovaný popis.
- [x] Šířka 980, 1100, 1250, 1600 px: bez běhové chyby a bez předčasného překrytí dokumentu pravými panely při výchozím automatickém zoomu.
- [x] Při rozbalení pravého pomocného panelu na 980 px se dokument automaticky zmenší tak, aby byl celý vidět.
- [x] Volba textového kurzoru ve split-tlačítku aktivuje Text; volba šipky aktivuje Objekty; obdélník/laso používají jednotný stav.
- [x] HTML obdélník přes obrázek + text: smíšený výběr; laso kolem textového vzorku: textová oblast. OCR zachová výběr oblasti.
- [x] OCR zobrazuje společné výběrové nástroje, nikoliv další nezávislý výběrový engine.
- [x] Vybraný demo obrázek odemkne editační panel: změna barevného režimu aktualizuje vizuální náhled, při zachování proporcí 1200 × 800 → 600 × 400 px a ořez mění výřez náhledu.
- [x] Magnet, mřížka, pravítka, `Zobrazení → Přichytit` a příslušné zaškrtávací volby mají koordinovaný stav.
- [ ] Produkční PDF výběr, rastrové transformace, color management, snapping, reálné OCR, tisk/export, podpisy a Undo/Redo je nutné samostatně implementovat a otestovat; **nejsou hotové jen vytvořením HTML**.

**Pravidlo pro Codex:** neodstraňovat existující funkce jen proto, že HTML je simulace. Nejprve mapovat ovládací prvek na současný produkční handler, analyzovat konflikty s `specifikace_Codex_v5.0.md` a v3.3–v3.7 a před migrací handleru připravit cílené regresní testy.

