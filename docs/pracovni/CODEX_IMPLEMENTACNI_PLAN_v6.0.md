# PDF Blacksmith — JEDINÝ IMPLEMENTAČNÍ PLÁN v6.0

**Autorita:** [CODEX_MASTER_v6.0.md](CODEX_MASTER_v6.0.md). Novější výslovné přání autora: **nejprve stabilizovat a sjednotit rozhraní aplikace**, nikoliv rozpracovávat F1. Tento plán nahrazuje staré implementační pořadí v „specifikace_Codex_v5.0.md“, „plan_rozsireni_v2.0.md“, „F1_plan_implementace.md“ a iniciačních dodatcích. Je to plán, nikoli automatická autorizace změn.

## Řízení etap
- Každý krok má **vstupní audit → jeden malý diff → testy → výstupní kontrola → schválení dalšího kroku**. Bez výslovného souhlasu žádný commit ani push.
- Zachovat současné produkční Go/Win32 UI a PDFium, nepřepisovat funkční části podle HTML od nuly.
- Všechny nalezené změny Codexu v rozpracovaném F1 nejprve vypsat; bez resetu a automatického mazání **pozastavit další práci na formulářích**.
- Přehled hotového vs naplánovaného dokládat konkrétním zdrojovým souborem a testem. HTML v3.26 není důkaz Go implementace.

## Fáze 0 – převzetí a inventura (nyní, bez změny kódu)
| Krok | Výstup | Akceptace |
|---|---|---|
| 0.1 | Stav větve, pracovní strom, lokální změny a poslední commity | Žádná změna cizího souboru, bezpečný diff |
| 0.2 | Matice existujících UI modulů/handlerů a procesů PDF | U každého A/B/D/P, odkaz na soubor, riziko |
| 0.3 | Srovnání produkčního rozhraní s HTML v3.26 | Seznam konkrétních neshod a požadovaných ovládání |
| 0.4 | Důsledky rozpracovaných F1 kroků | Co bylo změněno, co funguje a co je nutno ponechat rozpracované |
| 0.5 | Doporučení nejmenší bezpečné UI úpravy | Jeden konkrétní další úkol, přesné soubory, testy, rollback |

**Výstup teď:** jediná zpráva uživateli bez práce v kódu. Codex se nesmí odvolávat na staré „F1 je první priorita“ pro změnu tohoto pořadí.

## Fáze 1 – design systém a nejzákladnější rozhraní
**1A. Tokeny a palety:** 6×light/dark, žádné pevné vínové relikty, tmavé inputy ve světlých profilech, sémantika swatch vs UI. Ověřit v native Win32 způsob barvení; uvést screenshoty, pokud běží na Windows. **Nevylepšovat automaticky render PDF**.

První díl 1A načítá finální mapu JSON přímo do aplikace. Záhlaví, panely, jejich ouška, základní ovládací prvky a okolí PDF používají tokeny výchozího tmavého šedo-vínového motivu. Testy kontrolují všech 12 kombinací a kontrast zapojených rolí. Přepínání motivu v aplikaci patří do kroku 2C.

**1B. Ikonový systém:** Součástí je ověření chybějících SVG symbolů `i-panel_left`, `i-arrow_line`, `i-transform_object`, `i-plate_all` a `i-plate_single` (v aktuálním HTML jsou některé volány, ale nejsou definované); prověřit původní atlas, nekreslit náhrady bez souhlasu.  normativní mapování symbolů a statusů; první horní lišta a druhá menu/ikony, logo a magnet z původních autorských kreseb, bez změny command ID. Zkontrolovat 150/200 % DPI a přístupnost.

Druhá horní lišta, levá navigace a ovladač celé obrazovky používají 23 schválených symbolů z HTML v3.26. Atlas lze obnovit příkazem `python3 tools/generate_ui_icons.py`. Lišta volá stávající příkazy aplikace. Nabídka Formuláře je neaktivní do schválení práce na F1. Pět chybějících symbolů čeká na autorské podklady.

**1C. Základní docky:** Nástroje vlevo, Vlastnosti vpravo, miniatury a jeden pomocný panel; korektně se skládat, změna šířky a vracení stavu, práce v malém okně. Neodpojovat stávající OCR/separace.

Krátká horní ouška ovládají celé skupiny a rozbalení obnoví stav jednotlivých panelů. Výpočet šířek počítá s místem na ouška i dokument a podle zbývajícího prostoru rozšiřuje panely podle obsahu. Testy geometrie pokrývají čtyři šířky a tři měřítka. Kontrola ve Windows proběhne při společném sestavení po dalších pracovních blocích.

**Vstup:** schválení 1A nebo 1B nebo 1C. **Done:** 12 motivů, 3 DPI, 4 šířky, nepoškozené PDF a všechna původní menu.

## Fáze 2 – konzistentní chování aplikace
**2A. Menu, textové vs ikonové:** horní Soubor/Zobrazení/Nástroje/Úpravy/Formuláře/Pomoc bez rámečků; levá navigace a dropdown bez obdélníkových rámů s levým akcentním pruhem; aktivní ikona druhé lišty s výrazným stavem. Žádné změny funkčnosti příkazů.

Levá navigace a otevřené nabídky sdílejí paletové stavy najetí a výběru s levým akcentem. Horní textové nabídky mají jemný podklad při najetí a otevření. Ikony aktivních režimů mají výraznější podklad a spodní linku. Testy kontrolují kontrast těchto stavů ve všech dvanácti motivech.

**2B. Vlastnosti aktivního nástroje:** související podskupiny s oddělovačem a sbalením/rozbalením, bez dlouhých prázdných hint slotů. Otazník vedle sekce ukazuje metodický popis; **varování/validace se nikdy neschovávají**. Vlastnosti zachovat při sbalení.

Vlastnosti mají sbalitelné sekce a kruhové otazníky s metodikou. Po opětovném otevření zůstávají hodnoty polí i volby zachované. Výsledky a varování zůstávají v rozložení i při sbalení a prázdné výpisy se zobrazí až po naplnění.

**2C. Nápověda + lokalizace:** nativní tooltipy odpovídají CZ/EN, pole popisů podle fokusu, čtení pro klávesnici; volba jazyka v Nastavení/Jazyk, téma v Nastavení/Rozhraní i přes horní půlený přepínač. Lokalizační katalog 809 CZ/EN klíčů, nezasahovat do PDF.

První díl 2C napojuje katalog na hlavní lišty, hlavní a souborovou nabídku a navigaci. Nastavení jazyka umožňuje okamžitý náhled, potvrzení a návrat původní volby. Popisky tlačítek mají nativní tooltip a zobrazují se i při fokusu ve stavové liště.

Druhý díl překládá sekce, nápovědu a ovládání Vlastností, předvolby exportu a vybraná stavová a validační hlášení Win32. Při přepnutí jazyka se přepíšou popisky a již zobrazená hlášení s lokalizačním klíčem. Formáty PDF, ICC názvy, hodnoty polí, dokument a text technických výpisů z backendu zůstávají ve své původní podobě.

Třetí díl přidává Nastavení → Rozhraní se šesti paletami a světlým a tmavým režimem. Vzhled má okamžitý náhled, uložení volby a návrat při zrušení. Půlený přepínač v horní liště uloží změnu režimu stejné palety. Oba ovladače sdílejí nastavení a používají původní ikony z v3.26.

**2D. Kontextové pomocné panely:** automatické Výsledky při hledání a korektní návrat, dočasné Pomůcky rozvržení při interakci s vodítkem. Uživatelova manuální volba má přednost. Nereplikovat další pravý dock.

První díl 2D přidává Výsledky do společného prostoru se Záložkami. Při aktivaci hledání se Výsledky otevřou. Změna nástroje nebo Esc obnoví předchozí stav panelů a zohlední ruční volby provedené během hledání. Přepnutí karty vyprázdní výsledky předchozího PDF.

Druhý díl 2D přidává do Vlastností Pomůcky rozvržení. Přepnutí mřížky v nabídce Zobrazení je otevře dočasně. Kliknutí mimo panel nebo Esc obnoví Vlastnosti dosavadního nástroje a předchozí sbalení panelů. Přímá volba Pomůcek rozvržení v nabídce Zobrazení je otevře trvale pro zvolený nástroj. Mřížka s hlavními čarami po 5 mm a dělením po 1 mm patří náhledu. Napojení na výběr vodítka navazuje v 4C.

**Done:** zachování všech stávajících ikon i kláves; návrat z vyhledávání; žádné zbytkové tmavé plochy; přístupný focus.

## Fáze 3 – zobrazení stránek a interaktivní plátno
**3A. Žádné dekorativní rámy stránky:** zcela odstranit externí vínový outline v hlavní ploše ve stránkovém, souvislém i šířkovém režimu; skutečný obsah PDF nerušit.

Společné vykreslení hlavního náhledu ve Win32 má holou hranu stránky. List od okolí odděluje barevný kontrast aktuálního motivu. Čáry v PDF i zvýraznění při práci s objekty a hledáním se nadále vykreslují.

**3B. Normální navigace:** po stránkách scroll uvnitř listu → při dosažení kraje další/předchozí, ochrana před několikanásobným přeskokem, souvislý režim bez přerušení.

Ve stránkovém režimu běžné kolečko nejprve posouvá obsah aktuálního listu. Na hraně jedno souvislé gesto přejde nejvýš o jednu stránku. Jemné události kolečka se na hraně sčítají do jednoho kroku. Pauza nejméně 300 ms nebo změna směru zahájí nové gesto. V souvislém režimu posuv pokračuje přes hranice listů. Během editace textového pole běžné kolečko náhled neposouvá.

**3C. Zoom:** přizpůsobit stránce, šířce, 100 %; Ctrl+kolečko v PDF mění dokument, v miniaturách pouze jejich velikost, mimo obě zóny bez zachycení. V HTML Chromium/Edge ověřit wheel capture/passive false/preventDefault; produkční Win32 mít lokální obsluhu bez globálních zásahů.

Volba 100 % je v nabídce Zobrazení a v dolním přepínači měřítka. Přepočet velikosti PDF vychází ze systémového DPI. Ruční zoom má rozsah 30–900 % a automatické přizpůsobení se dál řídí rozměry stránky a plochy. Ctrl+kolečko v náhledu zachovává bod pod kurzorem a v miniaturách mění jejich velikost v rozsahu 10–60 % šířky panelu. Mimo tyto dvě oblasti přenechává aplikace událost běžnému zpracování Win32.

**3D. Adaptivní Přehled stran:** až po dokončení 3A–3C a schválení. Samostatný třetí režim (dvoustavové tlačítko pages/continuous zůstává jen dvoustavové), flex-wrap, výběr klik/Ctrl/Shift, Enter fit vybrané na výšku, sdílený výběr s miniaturami. **Skutečné přetahování a hromadné transformace** řešit zvlášť po zajištění Undo/Redo/page modelu. Současné HTML v3.26 mění jen ukázkové pořadí.

Přehled stran má samostatnou ikonu a volbu v nabídce Zobrazení. Při změně šířky a zoomu se karty přeskupují. Kliknutí, Ctrl, Shift a Vybrat vše používají jeden výběr s miniaturami a navazujícími volbami rozsahu. Enter v plátně nebo tlačítko Na výšku otevře vybranou stránku v režimu Po stránkách. Náhledy viditelných karet se načítají postupně ze společné mezipaměti.

**Done:** testy 3 stránek + velký soubor, zoom 30–900 % podle návrhu, DPI, focus, scroll anchoring, otáčení stran a dual view.

## Fáze 4 – výběr a editační pomůcky
**4A. Dispatcher:** sjednotit šipku/I-kurzor, univerzální obdélník/laso, Text/Objekty/OCR bez paralelních handlerů; nejdříve provést audit všech aktuálních kláves T/V/M/L.

Audit Win32 potvrdil T pro bodový výběr textu, V pro odstavec, M pro obdélník a L pro laso. V nástroji Objekty mají V, M a L přednost pro bodový, obdélníkový a polygonální výběr objektů. Rozpracovaný obdélník (včetně tažení kolečkem) a laso mají společnou obsluhu pro každé plátno. Tažení kolečkem v OCR předá oblast nástroji Text, kde je možné vybrat nativní text nebo potvrdit OCR oblasti. Obdélník, laso i tažení kolečkem umějí označit znaky a grafické objekty v jedné sadě. Obdélník i laso mohou začít na pracovní ploše kolem stránky. Při zahájení mimo stránku se vybere jedna z viditelných stránek zasažených vyznačenou oblastí. Shift přidává, Ctrl přepíná označení a Shift+Ctrl odebírá. Kliknutí s modifikátorem může doplnit text k objektu nebo objekt k textu na téže stránce. Pro práci s jedním typem obsahu jsou k dispozici dosavadní operace nástrojů Text a Objekty. V horní liště je přepínač šipky a I-kurzoru s ikonou aktuálního režimu a nápovědou pro následující akci. Vedle stojí samostatný obdélník a laso napojené na stávající režimy. Označený nativní text lze přesunout šipkou v rámci stránky. Smíšená sada se přesouvá jako jedna úprava pracovní karty s možností vrácení. Společné kopírování, mazání a přenos mezi dokumenty vyžadují samostatný krok. Bodové vkládání textu a kreslení nového odstavcového rámce vyžadují samostatné editační kroky.

**4B. Ovládací prvky:** Text Výběr čtyři ikony na jediné řadě + nápověda ?, Typography mixed, Objekty raster editor, OCR s vlastní nápovědou, zachování převodu text ↔ křivky jen jako operace nad výběrem.

V panelu Text jsou ve společné řadě kurzor (T), obdélník (M), laso (L) a odstavcový textový rámec. Klávesou V lze dál zapnout výběr existujícího odstavce. Otazník rozbalí v panelu českou nebo anglickou nápovědu k gestům a modifikátorům. OCR výběru zůstává dostupné pod Typografií i při jejím sbalení.

Typografie ukazuje smíšené písmo, velikost, proklad a řezy aktivního výběru. Když je v rámci pouze kurzor, panel zobrazuje hodnoty celého aktivního rámce. Nastavení mění jen zvolenou vlastnost a A+/A− posouvá každý úsek vzhledem k jeho velikosti. Šířka, řádkování a zarovnání se nastavují pro aktivní rámec.

Tažení nástrojem odstavcového rámce na jedné stránce určuje jeho umístění a šířku zalamování. Nový rámec roste do výšky podle textu nejvýš po okraj stránky. Psaní se potvrdí kliknutím mimo rámec nebo Ctrl+Enter a zapíše do pracovní karty jako jedna vratná úprava. Esc ukončí rozpracované tažení. Po otevření editoru rámce Esc ponechá rozepsaný text k další editaci. Pouhý klik a prázdný rámec změnu nevytváří. Při opětovné editaci má vložený odstavec vlastní šířku a zachovává ruční odřádkování. Textový kurzor (T) umožňuje na volném místě stránky kliknutím otevřít bodový text. Řádek se při psaní rozšiřuje od tohoto místa a standardně se zalamuje jen ručním odřádkováním. Kliknutí na existující text otevře jeho editaci a kliknutí na grafický objekt jej vybere. Bodový text má stejné potvrzení a historii jako odstavcový rámec.

V nástroji Objekty umožňuje ikona Oříznout obrázek vyznačit obdélník na jednom vybraném rastrovém obrázku. Použít nebo Enter zapíše ořez do pracovní karty. Zrušit nebo Esc zavře náhled. Ořez je jedna vratná úprava a ukládá se jako maska při vykreslení. Původní obrazový stream, rozlišení, barevný prostor a případná průhlednost se používají dál. Maska se přesouvá a duplikuje spolu s obrázkem.

Nabídka Poměr stran začíná u každého ořezu volným poměrem a nabízí čtverec 1 : 1 i ostatní pevné poměry podle specifikace. Poměr odpovídá zobrazené orientaci stránky. Při tažení k hraně obrázku se výřez přizpůsobí dostupné ploše při zachování poměru. Změna poměru přepočítá náhled podle posledního tažení.

V náhledu ořezu jsou čtyři tvary. Čtverec a kruh uzamknou poměr 1 : 1. Obdélník i ovál používají zvolený poměr včetně volného. Oválná a kruhová maska se do PDF ukládá jako Bézierovy křivky. Dříve uložené průniky masek se při otevření zobrazují v původní podobě.

**4B.1. Opravy výběru a ořezu před pravítky**

Při změně režimu nebo zahájení nového gesta se zruší čekající výběr a odpověď staršího požadavku už do dokumentu nezasáhne. Obdélník i laso zpracují text a grafiku v nástrojích Text a Objekty. Podržením Alt při tažení výběr přesunete. Zkoušky pokryjí rychlé přepnutí, výběr z okraje pracovní plochy, prázdný záběr pro OCR a přesun smíšeného výběru. Vybraný obrázek ihned zpřístupní nastavení ořezu. Horní ikony šipky a zobrazení vyjadřují aktuální režim, nápověda popisuje akci po kliknutí. Celý blok ikon je vpravo.

Z obdélníku a lasa se horním přepínačem přejde na šipku se zachováním výběru. Šipka přesune vybraný obsah běžným tažením. V režimu obdélníku nebo lasa běžné gesto vybere nový obsah i uvnitř dosavadního výběru. Alt při tažení vyhrazuje pohyb existujícího výběru.

Uzavřené laso vybírá jen prvky celé uvnitř vyznačené oblasti. Opakovaná volba lasa přepíná na **Křížový výběr** s čárkovanou otevřenou čárou. Enter nebo dvojklik výběr dokončí. Levé tlačítko kreslí aktivní režim. Kliknutí kolečkem kreslí druhý režim. Tažení kolečkem stále vyznačuje výběrový obdélník. Shift přidává k výběru a Ctrl zachovává dosavadní chování. Zpět a Znovu v otevřené kartě obnovují i obsahový výběr, bez změny dat PDF. Po úpravě PDF mají přednost kroky historie dokumentu.

**4B.2. Editovatelný ořez obrázků a oddělené transformace**

Názvosloví rozhraní a nápovědy používá **Ořez** nebo **Nástroj ořezu**. **Ořezová maska** určuje viditelnou oblast. **Maska objektu** vymezuje rozpoznaný objekt a **Maska pozadí** jeho pozadí. **Obsah masky** nebo **Obrázek v masce** označuje vrstvu uvnitř výřezu. **Poměr stran** je volný, přednastavený nebo vlastní a **Zámek poměru stran** jej drží při změně velikosti. **Zjemnění okraje** určuje přechod na hraně. **Inverze masky** převrací viditelnou a skrytou část. Akce **Obnovit ořez** vrátí plný rozsah obrázku. **Potvrdit ořez** a **Zrušit ořez** dokončí nebo ukončí rozpracovanou úpravu.

1. **Datový model a zpětná kompatibilita.** Zachovat obrazový stream, rozlišení, ICC a alfa kanál. Evidovat geometrii masky (poloha, rozměry, úhel a zaoblení) a transformaci obsahu (posun, měřítko a úhel) odděleně. Načítat dosavadní PDF s ořezy a umožnit znovu otevřít, upravit a obnovit výřez. Stav uložit spolu se standardní PDF maskou. Každé potvrzení je jedna vratná úprava.

   V nástroji **Ořez** lze u jednoho vybraného obrázku vyznačit nový výřez přes celý rozsah zdroje. Při otevření ořezu je vidět obrys uloženého výřezu a rám plného obrázku. Potvrzení nahradí ořezovou masku a může odkrýt dříve skrytá místa. **Obnovit ořez** odstraní masku vytvořenou aplikací a zachová aktuální polohu obrázku i další masky z původního PDF. Úprava a obnova mají samostatný krok zpět. V PDF je geometrie vlastní masky před samostatným umístěním obsahu. Lze přečíst i dříve uložené masky, které sdílely s obrázkem transformaci z okolního obsahu.
2. **Manipulace na plátně.** Vnější rám s rohovými a hranovými úchyty mění masku a dodržuje zámek poměru. Tažení za tělo přesouvá masku s obrázkem, samostatný režim hýbe jen maskou. Dvojklik nebo Posun obsahu v masce aktivuje vnitřní režim. Tažení a změna měřítka kolečkem či posuvníkem působí uvnitř pevné masky. Prázdné okraje lze povolit výslovnou volbou. Oříznuté části zobrazit při práci přibližně s 30% krytím. Při tažení ukazovat rozměry v pixelech, poměr stran a zvolenou kompoziční mřížku (třetiny, středové osy nebo zlatý řez).
   Osm úchytů upravuje rozměry nové i přístupné uložené ořezové masky při zachování zvoleného poměru. Při aktivním ořezu je skrytá část zdroje vidět přibližně s 30% krytím. Původní rastrová data vstupují do náhledu s vlastní průhledností. Potvrzení používá stávající nedestruktivní zápis masky.

3. **Vstupy a společná obsluha.** Vlevo nástroj Ořez a masky. Plovoucí lišta u výběru poskytne ořez, masku objektu, obnovu výřezu, převrácení, poměr stran, potvrzení a zrušení. Pravý panel poskytne číselné rozměry, polohu, úhel, zaoblení, vlastní poměr a nastavení masky. Ovládací prvky používají zdokumentované stálé identifikátory a sémantické akce i stavy `active`, `selected` a `masked`. Plátno a ostatní vstupy sdílejí stejné akce. Symboly se berou z finální mapy ikon a atlasu Win32.
4. **Druhy masek.** Nejprve čtyři základní tvary a zaoblený obdélník, potom polygon a křivka s editovatelnými uzly. Poté náhled segmentace objektu nebo pozadí s doladěním hran, zjemněním, inverzí a správou vrstev. Odsouhlasit model segmentace, licenci a způsob zpracování dokumentů před jeho integrací.

**4B.3. Transformace výběru**

Přidat viditelný přesun a úchyty objektů. Samostatně ověřit změnu velikosti, rotaci a deformaci objektů i volného textu. Volný text se přesouvá po stránce, odstavcový text se upravuje uvnitř svého rámce. Pro vybraný text doplnit přesun šipkami mimo režim psaní a nabídnout jeho ovládání na plátně. Měřítko, rotace ani deformace nesmějí likvidovat editovatelný text. Ověřit práci se smíšeným výběrem, historie změn a zachování původního PDF.

**4C. Pravítka a vodítka:** osa X/Y správně, pouze drag z pravítka, Ctrl+klik vs Ctrl+drag, per-page stav, skupiny a zámky, barvy „Použít“, kontext pravým klikem bez „Vybrat“, dočasná aktivace Pomůcek rozvržení. Nahradit pouze dekorativní paralelní prvky, ne PDF data.

**4D. SnapEngine:** zvlášť po prokázané přesné geometrii a výkonu. Nikdy nepřepisovat aktuální výběr, neukládat vodítka do tiskového PDF.

**Done:** pro každé gesto alespoň pozitivní/negativní test, overlap, PDF s rotací, rozložení CZ/EN, historií Undo/Redo a dokumentem bez rizika.

## Fáze 5 – F1 vyplňování formulářů (až na další souhlas)
**5A. F1-0:** inventura skutečné PDFium/API, anotací, podpisů a UI.
**5B. F1-1:** detekce AcroForm/Fields/Kids/Widget, Tab pořadí, XFA, návrh přechodu do Formulářů bez automatického přepnutí.
**5C. F1-2/F1-3:** vyplňování standardních polí a jejich pravý panel; zvýraznění polí pouze UI; Text/Objekty/OCR nepoškozeny.
**5D. F1-4/F1-5:** kontrola a bezpečné ukládání /V /AS /AP, Uložit přesto, diakritika, AutoRecover, podpisy a nezávislé PDF čtečky.

**Poznámka:** Dřívější dokument F1 plánoval UI uvnitř F1 až později. Nově má **základní shell přednost**, ale vytváření samotné funkcionality F1 není tímto autorizováno.

## Fáze 6 – návrhy F2, data a F3
**F2:** tvorba polí a vlastností ve Win32. HTML v3.26 je pouze ukázka, nikoli libovolný povolený write do AcroForm.
**Importy:** CSV UTF-8 vč. BOM/quoted, XLSX list/range, Google Sheets read-only OAuth, bezpečná vazba sloupců na /T a správné /AS a /Opt, 1 row → 1 result PDF, náhled a kontrola před exportem. Až po schválení parseru, autorizace a batch exportu.
**F3:** explicitní flatten do nové kopie; neslučovat s „bez poznámek“.

## Fáze 7 – volitelné další moduly
OCR rekonstrukce fontu, náročnější SnapEngine, Preflight, vizuální diff, měření, batch, převody/importy cizích souborů, šipkové nálepky, rozšířená měření pokrytí, externí drag/drop mezi aplikacemi. Schvalovat individuálně a neplést rezervní ikonu s dostupnou funkcí.

## Řízení kvality a forma reportu
Za každý dokončený krok Codex předá:
1. Přesný rozsah změny a důvod.
2. Seznam změněných souborů a regresních rizik.
3. Čím testoval (kompilace, jednotkové, integrační, UI Windows, screenshot, PDF round-trip), co **neotestoval** a proč.
4. Ověření funkce v přímém porovnání s v3.26 a s finálními mapami.
5. Potřebný další dílčí krok a otázky, které vyžadují rozhodnutí.
**Neoznačovat jako implementované** něco, co bylo pouze nakresleno v HTML nebo přidáno do specifikace.
