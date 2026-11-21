# PDF Blacksmith – Formuláře: podrobný návrh pro Codex (v1.9)

**Verze dokumentu:** v1.9 · 21. listopadu 2026  
**Stav:** etapa **F1 – schválena k postupné implementaci**; detaily UX/technologie v rámci F1 k dopracování a otestování. **F2 a F3 – dosud plán, nikoli schválení implementace.**  
**Vazba na celkovou specifikaci:** `PDF_Blacksmith_specifikace_Codex_v5.0.md` a plán `PDF_Blacksmith_plan_rozsireni_v1.9.md`.  
**Hlavní omezení:** dříve schválený požadavek Q31 – zachovat všechny stávající funkce a provádět regresní testy.

## 1. Cíl a rozsah

Do hlavní levé navigace přidat nový nástroj **Formuláře**. Zůstane zachována koncepce PDF Blacksmith: **levý seznam nástrojů → obsah PDF uprostřed → vlastnosti nástroje v pravém panelu**, tmavé téma, vínové zvýraznění, jednotná SVG sada s `currentColor`, změna obsahu jen na výslovnou akci uživatele.

Primární formát: **interaktivní formuláře AcroForm**. Systém musí rozlišovat **formulářové pole (`/AcroForm`, `/Fields`)**, **widget anotaci (`/Subtype /Widget`)**, **pracovní poznámku/anotaci**, **odkaz** a **digitální podpis**. Nepoužívat heuristiku „odstranit všechny `/Annots`“ – způsobilo by to ztrátu formulářů, odkazů a podpisů.

**Záměrně mimo základní rozsah:** provozování PDF JavaScriptu, komplexní XFA, databázové konektory, plnohodnotná implementace Adobe Forms API a nemožný slib kompatibility se všemi PDF generátory.

## 2. Uživatelské prostředí – návrh

### 2.1 Levá navigace

Přidat `Formuláře` mezi ostatní nástroje pracující s otevřeným PDF. Pro hlavní ikonu použít již připravený ekvivalent z rozšířené sady (mapa ikon v4.2). Když není dokument otevřený, nástroj není aktivní; nevytvářet bez schválení druhé samostatné PDF prostředí.

### 2.2 Pravý panel – návrh režimů

V horní části zůstane název **Formuláře**. Pod ním jeden nenápadný přepínač režimu; **nepřestavovat celou aplikaci ani jiné nástroje**:

1. **Vyplnit** *(výchozí při existujících polích)*: přímá práce s interaktivními poli ve stránce; stav polí, jednoduché upozornění na povinné údaje; v panelu přehled aktuálního formuláře. Nevyžadovat před každým kliknutím přepnutí na textový nástroj.
2. **Návrh polí**: malá ikonová paleta pro textové pole, víceřádkový text, zaškrtávací pole, radio skupinu, seznam, rozbalovací nabídku, běžné tlačítko a podpisové pole; pod ikonami obvyklý trvalý popis aktivního nástroje a krátká nápověda při hoveru podle již schváleného UI modelu.
3. **Vlastnosti**: zvolený název a typ pole, hodnota, výchozí hodnota, povinné / pouze pro čtení, tooltip, souřadnice, rozměry, písmo, zarovnání, barva, obrys, zobrazit/tisknout, volby seznamu a pořadí Tab. Bez vybraného pole ukázat vysvětlující stav namísto chybového prázdného panelu.
4. **Kontrola**: seznam povinných chybějících polí, problémů s výskytem vzhledových streamů a nekompatibilních operací; po kliknutí přejít na správnou stránku a pole. Rozlišovat **chybu / varování / informaci**. Nedělat falešná upozornění u existujících opakovaných polí se záměrně sdílenou hodnotou.

Vhodné jako **režimy jednoho panelu**, nikoli jako čtyři nové nástroje v levé navigaci. Grafickou podobu odsouhlasit podle současných screenshotů před implementací.

### 2.3 Práce se stránkou a myší

- V režimu **Vyplnit** klik do interaktivního pole zaměří dané pole; stisk **Tab / Shift+Tab** přejde na další/předchozí pole; **mezerník** aktivuje checkbox a radio; šipky ovládají seznam. Stránku lze dál přibližovat a posouvat zaběhnutými gesty.
- Režim **Návrh**: levým tažením se založí hranice nového pole; **Shift** podle typu udrží poměr stran, **Ctrl** kreslí od středu pouze tam, kde to dává smysl. U existujících polí výběr a úprava pozice/rozměrů; červená/vínová zvýrazní pouze aktuální pole.
- Globální **stisk kolečka + tažení pro obdélníkový výběr** a běžný zoom/scroll zůstanou zachovány dle Q1 a souvisejících pravidel. Přednost má gesto kolečka, nikoliv textový vstup.
- Zaměřené pole musí být vizuálně rozlišitelné od vybraného vektorového objektu či textového rámce. V případě překryvu pole má při aktivním režimu Vyplnit přednost interaktivní formulářové pole; v nástroji Objekty/ Text platí jejich vlastní pravidla výběru.
- Formulářové pole lze vybrat i z přehledu polí; nabídnout přejít na příslušnou stranu. Při více otevřených PDF se příkazy vztahují jen k aktivní kartě/panelu.

## 3. Etapa F1 – existující formuláře (priorita)

### F1-01 Detekce a identifikace

- Otevření PDF provést s detekcí `/AcroForm`, `/Fields`, hierarchie `/Kids`, stránky widgetu, `/Rect`, `/FT`, `/Ff`, `/V`, `/DV`, `/AS`, `/AP`, `/DA`, `/DR`; u tlačítek rozlišit checkbox/radio/pushbutton pomocí příznaků.
- Správně mapovat úplné názvy polí i dědičné vlastnosti v hierarchii; neměnit `objgen` identitu zbytečným přepisem. Zachovat volby a záměrně opakované widgety, pokud odkazují na jedno pole.
- V editoru zobrazit, kolik polí PDF má, a na kterých stránkách se nacházejí.
- Pokud dokument obsahuje XFA, zobrazit upozornění **„Formulář XFA není v této verzi plně podporován. Dokument nebude upraven ani uložen způsobem, který by mohl odstranit data formuláře.“** Bez tichého převodu a ztráty dat.

### F1-02 Typy polí

- `/Tx`: krátký text a víceřádkové pole, správné řádkování, limit znaků pokud je definován, výchozí hodnota, vyznačení pouze pro čtení.
- `/Btn`: checkbox a radio s pojmenovanými exportními stavy (nepředpokládat vždy `/Yes`); výběr ve skupině s platnými hodnotami. Pushbutton existujícího formuláře se uchová, ale neznámé skripty / akce se bez schválení nevykonávají.
- `/Ch`: výběrový seznam a rozbalovací nabídka, hodnoty i zobrazené popisky, příznaky pro případný volný text.
- `/Sig`: rozpoznat podpisové pole; jeho skutečné vyplnění delegovat existujícímu nástroji Digitální podpis. Nepřepisovat certifikáty ani nevytvářet neplatnou podobu podpisu.

### F1-03 Ukládání, vzhled a editace

- Editace hodnoty znamená neuloženou změnu a vstup do schválené historie Ctrl+Z / Ctrl+Y a AutoRecover; pouhý fokus, procházení Tab nebo klik bez změny **neznamenají editaci PDF**.
- Uložit obsažené hodnoty do odpovídajících polí; správně udržovat/revitalizovat `/AP` (appearance stream), `/AS` a `/V`. Pouhý `/NeedAppearances` není dostatečná záruka, že se data vytisknou shodně v jiných čtečkách.
- **Čeština a diakritika:** vybrat font s glyphy, udržet editovatelnost, zajistit funkční vzhled pro české znaky v poli; není přípustné tiché nahrazení čtverečky ani ztráta diakritiky. Pro případ nepodporovaného fontu zobrazit konkrétní hlášení a nabídnout bezpečný náhradní font.
- Vykreslování interaktivních prvků ve Windows navázat na form-fill API PDFium (`FPDFDOC_InitFormFillEnvironment`, interakční události, `FPDF_FFLDraw`), nebo ekvivalent s prokazatelně stejným výsledkem; nastudovat aktuální verzi API a ověřit chování v aplikaci.
- **Uložit upravenou kopii…** musí zachovat dříve schválené přepnutí aktivní karty na novou kopii a zavření starého pracovního dokumentu až po úspěchu.

### F1-04 Jednoduchá validace (upřesněno schválením F-Q8)

- Varovat u **povinných prázdných** polí a u jednoduchých ručně nastavených typů formátu (např. e-mail, datum) jen pokud jsou v naší aplikaci definovány; neinterpretovat libovolný stávající PDF JavaScript.
- **Uložit** nedokončený rozpracovaný formulář smí být možné; neúplné povinné údaje nesmí automaticky blokovat ukládání, protože uživatel často uloží rozpracovaný formulář a vrátí se k němu.
- Před tiskem/exportem může aplikace nabídnout upozornění, nikoliv bez výslovného souhlasu smazání polí.

## 4. Etapa F2 – tvorba a úpravy polí

Nové ikony do pravého panelu pocházejí ze schválené v4.2 rozšířené mapy. Zpřístupňovat je **až po implementaci skutečné funkce**.

| Typ | Základní vlastnosti | Důležitá poznámka |
|---|---|---|
| Textové pole | název, tooltip, hodnota, font, velikost, barva, maximum znaků, zarovnání, hranice | podpora češtiny a potřebných glyphů |
| Víceřádkové pole | vše výše + automatické zalamování | žádné ořezané řádky |
| Checkbox | exportní stav, výchozí zaškrtnutí, popisek | nepředpokládat `/Yes` |
| Radio skupina | název skupiny, hodnoty odpovědí, výchozí výběr | v rámci skupiny jedna zvolená hodnota |
| Combo | seznam zobrazovaných a exportních hodnot, případný volný zápis | správně uchovat `/Opt` |
| Listbox | položky, volba, případný vícenásobný výběr | dle podpory standardu |
| Tlačítko | pojmenování, vzhled, bezpečná vestavěná akce | nevykonávat neprověřený skript |
| Podpisové pole | poloha, účel, tooltip | podepisování navázat na současný podpisový modul |

**Vlastnosti a navigace:** souřadnice v PDF bodech + zobrazení mm; přesné pozicování; přichytávání až po schválení měřicího modulu; kopírování a vložení s upozorněním při kolizi názvů; nepoškozovat existující opakované widgety stejného pole; pořadí Tab (včetně klávesnice a asistivních technologií). Vyskakovací tooltip dle přístupnosti PDF.

## 5. Etapa F3 – tisk, export a zploštění

- **Uložit / Ctrl+S a Uložit upravenou kopii…** – formulářová pole zůstávají **interaktivní** včetně hodnot. Původní PDF nepřepsat bez výslovné volby.
- **Export bez poznámek** (dříve schválený předtiskový default Q24) – odstranit jen pracovní komentářové/grafické anotace. **Neodstraňovat formulářové widgety, jejich vyplněné hodnoty, odkazy, ani podpisová pole.**
- **Tisk** – pole se zpracují podle platných PDF příznaků tisknutelnosti, nikoliv jako pracovní poznámky. Náhled i tisk musí obsahovat vyplněné hodnoty, pokud je pole určeno pro tisk.
- **Zploštit formulář pro tisk…** – explicitní operace na nové kopii, v níž se vzhled vyplněných polí přenese do výsledného obsahu stránky a neinteraktivní pole se odstraní až po úspěšné kontrole; nesmí se přitom ztratit text, šířky čar, barvy, spot a geometrie ostatního PDF. Pokud zploštění zachování vzhledu bezpečně negarantuje, operaci odmítnout s vysvětlením. Ovládání musí přiznat, že hodnoty pak již nepůjdou editovat jako formulář.
- **PDF/X:** interaktivní formulář není synonymem tiskově připraveného statického PDF. Při exportu do PDF/X vyhodnotit profil a účel; nabídnout samostatné zploštění / tiskovou reprezentaci s předběžnou vizuální kontrolou a znovu spuštěným Preflight, jinak srozumitelnou chybu. Nedefinovat falešnou shodu s PDF/X.
- **Podepsaná PDF:** změna pole může porušit platnost podpisu; upozornit a podpisy zachovat nebo práci odmítnout podle povoleného rozsahu. „Zploštění“ podpisu nevytváří platný elektronický podpis.

## 6. Co nedělat

1. Nepoužít obecnou operaci typu „smazat všechny anotace“; zničila by AcroForm widgety a odkazy.
2. Neredukovat text, objekty, OCR, práci se stránkami, záložkami, poznámkami, exportem a nástrojem podpisu kvůli novému panelu.
3. Nezakládat duplicitní druhý mechanismus historie/AutoRecover.
4. Nevytvářet vzhled polí pouze jako overlay obrazovky, pokud se vzhled nezapíše i do PDF.
5. Neinterpretovat nepodporovaný XFA formulář jako prázdný AcroForm.
6. Nevydávat POC vytvořený pomocí ReportLabu za hotovou implementaci Go/Windows.

## 7. Integrace do aktuálního kódu – výchozí investigace pro Codex

Před implementací ověřit v repozitáři `Liborias/PDF-Blacksmith` aktuální obsah souborů a Git SHA. Na podkladě původního snapshotu:

- `workspace_windows.go`: `workspaceTabs` dosud nezahrnuje samostatné **Formuláře**; zavést `idFormsTab` a panel bez kolize s existujícími identifikátory.
- `reader_engine_windows.go`, `reader_render.py`, `viewer_windows.go`: existující PDFium pipeline a pracovní vykreslování; ověřit, jak zprostředkovat focus, vstup a appearance widgetů při interakci.
- `engine.py`, `object_edit.py`: backend s pikepdf/Ghostscript; doplnit model formularových hodnot, nikdy nepromíchat se standardními anotacemi poznámek.
- `history_windows.go`, `file_menu_windows.go` a související logika dokumentů: použít existující historie/uložení, nové formulářové změny auditovat stejně jako úpravy textu či stránky.
- `print_windows.go`, `export_settings.py`, `proof.py`: zajistit tiskovou/exp. reprezentaci polí, oddělenou viditelnost poznámek a správný výstup v barevné kontrole.
- `document_commands_windows.go`: klávesové zkratky v souladu s F1 nápovědou, Ctrl+Z/Y, Ctrl+S, Tab a běžnými textovými editory. Žádná neověřená kolize kláves.

**Před prvním commitem:** inventura stávajících funkcí v dotčených souborech, testovací fixture a rollback scénář. Práce na samostatné větvi, menší dávky a testy před spojením.

## 8. Testovací PDF a akceptační scénáře

Přiložený soubor `PDF_Blacksmith_Formulare_testovaci_AcroForm_v1.pdf` má **2 strany a 17 interaktivních polí** různých typů (včetně radio skupiny, readonly pole a víceřádkových vstupů). Slouží jako jeden základní průchozí test, **nepokrývá digitální podpis ani složitá XFA PDF**.

| Test | Akce | Očekávaný výsledek |
|---|---|---|
| T-01 | Otevřít zkušební formulář | 2 strany; správně detekovaných 17 polí; správná místa a typy |
| T-02 | Kliknout, Tab, Shift+Tab | navigace mezi skutečnými poli, viditelný focus |
| T-03 | Vyplnit text s diakritikou (např. „Jiří Černý – Znojmo“) | diakritika v `/V`, správné `/AP`, bez ztráty znaků |
| T-04 | Upravit multiline, checkbox, radio, combo, listbox | uložení a nové otevření zachová každou hodnotu |
| T-05 | Pokus upravit readonly pole | pole zůstane neměnné, bez nechtěného dirty stavu |
| T-06 | Nechat povinné pole prázdné | kontrola ukáže varování, ale lze uložit rozpracovanou kopii |
| T-07 | Ctrl+Z/Ctrl+Y | návrat a obnovení hodnot každého pole |
| T-08 | Uložit upravenou kopii | nová kopie nahradí starý dokument v původní kartě, originál beze změny |
| T-09 | Simulovaný pád aplikace, opětovné spuštění | AutoRecover nabídne platnou obnovu hodnot |
| T-10 | Export bez poznámek | formulářová pole, odkazy a hodnoty zachovány |
| T-11 | Tisk | vyplněné hodnoty viditelné na papíře a v náhledu |
| T-12 | Explicitní zploštění nové kopie | výsledek vizuálně shodný, pole už nejsou editovatelná, originál beze změny |
| T-13 | Otevřít XFA PDF | varování; nelze nechtěně odstranit ani nahradit data |
| T-14 | Podepsaný dokument | nevykonat změnu bez posouzení dopadu na podpisy |
| T-15 | Duální zobrazení a přepnutí nástroje | nepřeskakuje fokus na jiné PDF, ovládání ostatních nástrojů beze změny |
| T-16 | Otevřít AcroForm v Acrobat/Edge | vzhled a hodnoty se významově shodují, případné rozdíly řešit před uvedením |
| T-17 | Otevřít PDF s odkazy a pracovními poznámkami | každá anotace zachová správnou roli a nezaměňuje se s widgetem |

Testy 01 a vizuální výchozí stav jsme mohli prověřit na vytvořené fixture pomocí pypdf a vykreslení; ostatní testy jsou **akceptační scénáře pro budoucí implementaci**, nikoliv dnes prokázaná funkčnost PDF Blacksmith.

## 9. Otevřené otázky k odsouhlasení

**F-Q1 – SCHVÁLENO (21. 11. 2026), varianta A:** první implementační etapou je **F1 – čtení, vyplňování a bezpečné ukládání existujících interaktivních formulářů AcroForm**. Implementace návrhu nových polí patří do **F2** a explicitní zploštění a specializované výstupy do **F3**, nikoli do F1. Schválení etapizace samo nepotvrzuje další nerozhodnutá UI nebo technologická pravidla.

Následně v samostatných rozhodnutích: přesná vizuální podoba pravého panelu, zda explicitní prázdný nástroj umožní vytvořit formulář na běžném PDF, výchozí chování při tisku nevalidních formulářů, autorizované podpisové pole, a pořadí Tab při sloučení několika dokumentů.

## 10. Ověřené technické reference

- PDFium Form Fill SDK: https://pdfium.googlesource.com/pdfium/+/refs/heads/chromium/7417/public/fpdf_formfill.h
- pikepdf – interactive forms: https://pikepdf.readthedocs.io/en/latest/topics/interactive_forms.html
- pikepdf – form API: https://pikepdf.readthedocs.io/en/stable/api/form.html
- ReportLab – PDF interactive forms (použito jen pro návrh testovacího souboru): https://docs.reportlab.com/reportlab/userguide/ch4_pdffeatures/



---

## 11. Implementační plán F1 po schválení F-Q1 – závazný rozsah, doporučené pořadí

**Schválené rozhodnutí:** uživatel výslovně vybral **variantu A**: F1 odděleně a jako první. Kód v repozitáři zatím tímto dokumentem měněn nebyl. Technické fáze níže jsou doporučeným způsobem dodání F1, nikoli nově schváleným redesignem jiných částí UI.

1. **F1-A – Inventura a model polí:** před úpravami načíst skutečné `workspaceTabs`, stávající ovládání a ukládání, historii a PDFium pipeline; identifikovat `/AcroForm`, `/Fields`, `/Widget`, dědičné atributy, názvy a vazby stránky–widget. Zachovat i neznámé části PDF.
2. **F1-B – Zobrazení a fokus:** pasivně zobrazovat existující hodnoty a vzhledové streamy, zavést bezpečný fokus a Tab/Shift+Tab pro textové a výběrové widgety, respektovat už schválená gesta myši a jiné nástroje; žádné nevyžádané vykonávání PDF JavaScriptu.
3. **F1-C – Editace:** krátký a víceřádkový text, checkbox, radio, combo a listbox, ReadOnly/Required/MaxLen a různé exportní stavy tlačítek; potvrzení změny hodnoty musí být právě jeden srozumitelný krok historie.
4. **F1-D – Bezpečné uložení:** správně zapisovat `/V`, `/AS`, `/AP` a podpůrné odkazy s kompatibilním vzhledem v jiných čtečkách, s důrazem na českou diakritiku. `Ctrl+S`, „Uložit upravenou kopii…“, přepnutí na kopii a původní nezměněný soubor. Výstupy kontrolovat nezávislým parserem a vykreslovačem.
5. **F1-E – Odolnost:** využít jedno společné Ctrl+Z/Y a AutoRecover, provést simulaci selhání zápisu, vědomé „Neuložit“, pád aplikace a novou relaci. Načtení XFA nebo podepsaného souboru nesmí poškodit obsah ani podpis bez výslovné autorizované operace.
6. **F1-F – Přijímací testy:** před zahájením F2 doložit průchod testy **T-01 až T-11, T-13 až T-17** podle relevantnosti pro F1; **T-12 (zploštění)** je plán pro F3 a není blokující podmínkou F1. Ověřit hlavně češtinu, vzhled při tisku, nezměněné odkazy/poznámky a rozdíl mezi uložením a čistým exportem.

**Co není součástí F1:** přidávání a přemisťování polí, editor jejich typů či vlastností, pořadí Tab v editoru návrhu, vlastní výpočtové skripty, XFA konverze, zploštění pro tisk. Existující podpisové pole se v F1 rozpozná, ale podepisování zůstává úkolem stávajícího nástroje **Digitální podpis**.

**F-Q2 – SCHVÁLENO (21. 11. 2026), varianta A:** při otevření PDF s rozpoznanými interaktivními formulářovými poli se **nepřepne automaticky žádný nástroj**. Aplikace zobrazí nenápadnou, neblokující nabídku **„Dokument obsahuje formulář – Otevřít Formuláře“**. Teprve uživatelovým kliknutím se aktivuje **Formuláře → Vyplnit** pro daný dokument. Detekce ani zobrazení nabídky samy o sobě nemění dokument a nesmějí zakládat položku historie úprav nebo nastavovat příznak neuložených změn. Ostatní činnosti v právě aktivním nástroji mohou bez přerušení pokračovat. V duálním zobrazení je nabídka přiřazena ke konkrétnímu PDF, nikoli k nesouvisející kartě. Nepoužívat modální dialog ani automaticky převzít fokus textového editoru. Pokud je soubor XFA nebo nelze jeho formulářová data bezpečně načíst, nenabízet zavádějící příslib editace; zobrazit pravdivé upozornění na omezení.

**Doporučený detail UX (bez dalšího samostatného schválení):** v pasivní nabídce tlačítko „Otevřít Formuláře“ a nenápadné zavření `×`; jednou zavřenou nabídku neukazovat po každém přepnutí stránky. Nástroj Formuláře je nadále ručně dostupný v levé navigaci. Rozhodnutí F-Q3 následně určuje, že **pole lze vyplňovat pouze v aktivním nástroji Formuláře → Vyplnit**; v jiných nástrojích zůstává jejich běžné chování.



---

## 12. F-Q3 – SCHVÁLENO (21. 11. 2026), varianta A s typografickou podmínkou

**Závazné uživatelské rozhodnutí:** existující interaktivní formulářová pole PDF lze aktivně **vyplňovat pouze tehdy, když je vybraný nástroj `Formuláře → Vyplnit`**. Kliknutí na stejnou oblast při použití `Text`, `Objekty`, `OCR`, `Separace`, `Poznámky` nebo běžného prohlížení nemá měnit hodnoty formulářového pole; nástroje si ponechají dosavadní význam kliknutí, výběru a klávesových zkratek. V duálním zobrazení se respektuje právě aktivní dokument. Žádné automatické přepínání do formuláře po kliknutí.

**Podmínka schválení:** při budoucím **vytváření a úpravách formulářových polí v etapě F2** musí být k dispozici **srovnatelný rozsah typografických a editačních nastavení jako v existujícím nástroji Text**, ovšem přizpůsobený logice PDF formulářů. To znamená zejména:
- typ a řez písma (Regular/Tučné/Kurzíva, případně podtržení pouze když ho daný typ pole dovoluje), velikost a barva textu;
- zarovnání, podle typu pole i víceřádkový text, řádkování, vnitřní odsazení a zalamování; srovnatelná ergonomie ovládacích prvků a ikony ze společné SVG mapy;
- výchozí vzhled pole, chování při delším obsahu, maximální délka a volitelné označení pole jako povinného;
- název/identifikátor pole, výchozí hodnota, textový popisek a nápověda, pozice a rozměry pole;
- zachování české diakritiky, kompatibilních PDF fontů, appearance streams a tiskového vzhledu.

**Zásadní rozlišení:** ve **F1/Vyplnit** uživatel píše hodnoty **ve vzhledu definovaném autorem formuláře**. Nezobrazovat při vyplňování běžně ovládání pro změnu fontu, velikosti, barvy a geometrie pole. Možnosti měnit vlastnosti pole a jeho typografii se týkají až **F2/Návrh/Vlastnosti**. Při atypicky chybějících či nekompatibilních fontech může F1 bezpečně ošetřit vykreslení a nabídnout kompatibilní fallback, ne však svévolně přepsat vzhled pole.

**Stav etap:** F1 je schválena k postupné implementaci; F2/F3 nadále zůstávají plánem. F-Q3 potvrzuje požadovaný budoucí UX F2, **nepovoluje automaticky zahájit implementaci F2**. Nástroj Text ani objektové výběry se kvůli tomu nepřestavují.

### Akceptační scénáře navíc

- **T-18:** aktivní Text/Objekty → kliknutí na widget PDF nezmění jeho `/V` ani nevloží text do pole; poté přepnutí na Formuláře/Vyplnit → editace stejného pole je dostupná.
- **T-19:** F1 používá písmo/vzhled nastavený formulářem; české znaky se správně vykreslí v náhledu, po uložení i v jiné čtečce.
- **T-20 (později F2):** tvůrce formuláře může nastavit typografické vlastnosti pole podobně jako v nástroji Text; při následném vyplňování se tento vzhled respektuje a není svévolně měněn.


---

## 13. F-Q4 – SCHVÁLENO (21. 11. 2026), varianta A: zvýraznění vyplnitelných polí

**Závazné chování:** po ruční aktivaci **Formuláře → Vyplnit** se všechna **dostupná, viditelná a editovatelná** formulářová pole aktivního PDF jemně barevně zvýrazní. V pravém panelu bude přepínač **„Zvýraznit vyplnitelná pole“**, ve výchozím stavu **zapnutý**. Vypnutí okamžitě odstraní zvýraznění, ale zachová editovatelnost polí. Přepnutí nepoškodí ani nezmění dokument.

- Zvýraznění se kreslí **výhradně jako dočasná UI vrstva** nad obrazovkovým náhledem a není součástí původní grafiky, PDF anotací ani appearance streams; neukládá se do PDF a netiskne se ani neexportuje. Nezakládá položku Undo/Redo, AutoRecover ani příznak neuložených změn.
- Jemné tónování odlišovat od vyznačení aktivního pole: **fokus/editace mají viditelnější obrys** a zůstávají přístupné i při vypnutém plošném zvýraznění. Konkrétní odstín a průhlednost přizpůsobit stávajícímu tmavému tématu a vínovému akcentu, aby nesplýval text a čitelnost polí.
- **Nezvýrazňovat** skrytá či nezobrazovaná pole, pole `ReadOnly`, vypnuté/nepovolené widgety a objekty, které vůbec nejsou formulářovým polem (například komentáře či odkazy). Podpisové pole se rozpoznává, ale nepovažuje za běžně vyplnitelné textové pole; může mít vlastní zřetelné označení dostupné přes nástroj Digitální podpis.
- Zvýraznění je viditelné **pouze v aktivním nástroji Formuláře → Vyplnit**. Po přepnutí na Text, Objekty, OCR, Separace, Poznámky aj. zmizí, aby neměnilo zavedenou logiku výběru a práce s PDF. Po návratu do Formulářů respektuje stav přepínače. V duálním náhledu se vizuální stav nevztahuje na nesouvisející dokument a upravuje se pouze aktivní PDF.
- Po posunutí, změně velikosti okna, přiblížení, otočení stránky nebo přepnutí stran je zvýraznění stále zarovnáno s příslušnými formulářovými widgety; nesmí být odvozené napevno z pixelových souřadnic.
- Budoucí nastavitelnost mezi jednotlivými dokumenty lze navrhnout podle již schváleného modelu viditelnosti poznámek, ale nevyžaduje nyní nový globální přepínač ani změnu již uloženého PDF.

**Akceptační testy:**

- **T-21:** otevřít PDF s několika viditelnými editovatelnými poli a jedním ReadOnly; aktivace Formuláře → Vyplnit zvýrazní pouze editovatelná pole, nevytvoří neuloženou změnu.
- **T-22:** vypnutí přepínače odstraní pomocnou barvu, přesto lze pole vyplnit; opětovné zapnutí ji obnoví.
- **T-23:** přepnutí do Text/Objekty odstraní zvýraznění; návrat do Formulářů je obnoví podle stavu nastavení.
- **T-24:** uložené PDF, export bez poznámek i tiskový náhled a tisk neobsahují žádné dodatečné zvýraznění.
- **T-25:** změna zoomu, DPI, otočení a panelů neodsune zvýraznění od widgetů; při duálním zobrazení se nemění nesouvisející PDF.

**Stav:** F-Q4 je schváleno. Jde o rozšíření F1, nikoli o schválení realizace editoru polí F2 či zploštění F3.


---

## 14. F-Q5 – SCHVÁLENO (21. 11. 2026), varianta A: sbalitelný přehled formulářových polí

**Závazné rozhodnutí:** do pravého panelu nástroje **Formuláře → Vyplnit** přidat **sbalitelný přehled formulářových polí právě aktivního dokumentu**. Jde o doplněk k přímému vyplňování na stránce, nikoliv jeho náhradu. Zachovat dosavadní vzhled aplikace a strukturu pravého panelu; nevytvářet další hlavní okno ani další samostatný nástroj.

### Obsah a interakce seznamu

- U každé položky zobrazit **srozumitelný popisek** (název/tooltip pole, pokud existuje), informaci o **stránce** a podle skutečných atributů pole zřetelně odlišit **vyplněné**, **nevyplněné**, **povinné**, **pouze ke čtení** a případně **nedostupné** pole. Nepřikládat status „povinné“ automaticky každému prázdnému poli; stav povinnosti číst z vlastností PDF.
- **Kliknutí na řádek** otevře odpovídající stránku a posune zobrazení k poli, případně jej zaměří, pokud je editovatelné a viditelné. V případě ReadOnly se přejde na místo, ale nepovolí se zápis. U nedostupných/skrytých widgetů nepředstírat, že je lze běžně editovat.
- Výběr pole na stránce synchronizovat se seznamem (zvýraznění odpovídajícího řádku). Navigace **Tab/Shift+Tab** zůstane zachována pro skutečná vyplnitelná pole na stránce; přehled ji nesmí narušit. Fokus musí zůstat přístupný klávesnicí, se standardním označením vybraného řádku.
- **Sbalení přehledu** pouze uvolní místo v panelu. Nezmění hodnotu žádného pole, jeho vzhled, stav validace, historii, výběr v dokumentu ani stav neuložených úprav. Po rozbalení se obnoví seznam a zvýrazní aktuálně vybrané pole.
- U velkých formulářů zajistit plynulý posuvný seznam a neblokovat UI při inventuře stovek položek. V duálním zobrazení přehled vždy patří **aktivnímu PDF** a po přepnutí dokumentu nesmí zobrazovat ani měnit pole druhé karty.
- Pokud dokument nemá interaktivní pole, ukázat stručný prázdný stav „Tento dokument neobsahuje vyplnitelná formulářová pole“, nikoliv prázdný technický seznam. Neznámé typy polí a XFA jasně označit podle skutečně dostupné podpory.
- Jestliže **jedno logické pole má více widgetů**, neduplikovat jeho hodnotu v modelu. Přehled může zobrazit odpovídající výskyty/stránky, ale změna jediného pole se musí projevit ve všech relevantních reprezentacích. U radio skupiny uvádět smysluplný společný status a stav výběru.
- **Přehled F1 je navigační a informační.** Nemá v sobě editor písma, pozice ani definic polí; tyto možnosti budou součástí až F2 a budou navazovat na schválenou typografickou podobnost s nástrojem Text.

### Řazení – F-Q6 SCHVÁLENO (21. 11. 2026), varianta A

**Přehled v pravém panelu se řadí podle aktuálního pořadí stránek a vizuální polohy polí na stránce**, obvykle shora dolů, na stejné úrovni zleva doprava. **Pořadí procházení klávesou Tab je nezávislé a musí zůstat zachované podle původních pravidel a nastavení PDF.** Žádné automatické přeuspořádání widgetů, změna atributů formuláře, aktualizace `/Annots`, přepis pořadí ve stromu `/Fields` ani automatická modifikace dokumentu jen kvůli seznamu.

**Doplňující interpretační pravidla pro implementaci** (technikálie k ověření, nikoli rozšíření schváleného UX):
- Řadit podle **aktuálního pořadí stránek v otevřeném dokumentu**. Při přesunu či otočení stránky přepočítat polohu/vizuální pořadí položek tak, jak je vidí uživatel; neukládat obrazovkové pixely jako trvalá data formuláře.
- V rámci stránky přednostně porovnávat **vizuální umístění widgetů**; blízké svislé souřadnice chápat jako tentýž řádek a rozhodnout zleva doprava, při rozdílných řádcích shora dolů. Zohlednit natočení stránky, CropBox a aktuálně vykreslované souřadnice. Řazení je stabilní i při pouhém zoomu, změně DPI nebo sbalení panelu.
- U více widgetů **téhož logického pole** nezdvojovat datovou hodnotu. V seznamu uvést jedno logické pole s informací o umístění/instancích (například podřízené odkazy na další výskyty); navigaci přiřadit správné viditelné instanci. U radio skupiny zachovat společnou hodnotu skupiny.
- Formulářová pole **bez viditelného widgetu** (například skrytá) nevkládat mezi vizuální pole zavádějícím náhodným pořadím; v informačním přehledu je lze oddělit do sekce „Bez viditelného umístění“ a jasně označit, že na ně nelze běžně přejít.
- Pro **shodnou pozici** dvou různých widgetů používat deterministické pomocné řazení (např. původní index v dokumentu), aby se seznam samovolně nepřehazoval.
- Výslovně **nezaměňovat pořadí seznamu a navigaci Tab/Shift+Tab**. Klávesa Tab při vyplňování používá logiku formuláře/PDFium a nesmí automaticky používat vizuální pořadí přehledu.
- Seznam je pouze prezentace a navigace. Řazení nezakládá záznam Undo/Redo ani neuloženou změnu, nemění podpisy, nezasahuje do OCR/textové vrstvy ani do pracovních poznámek.

### Akceptační testy

- **T-26:** Formuláře/Vyplnit ukazuje sbalitelný seznam se správným počtem logických polí, rozlišuje povinné/vyplněné/ReadOnly, nepočítá pouze widgety jako vždy odlišná pole.
- **T-27:** kliknutí na položku přejde na správnou stránku a viditelnou oblast; při ReadOnly nenabídne změnu hodnoty. Výběr na stránce aktualizuje příslušný řádek.
- **T-28:** sbalení/rozbalení seznamu nevyvolá neuloženou změnu a nezmění hodnoty, historii Ctrl+Z/Y ani AutoRecover.
- **T-29:** přepnutí karet a duální režim zobrazují stav správného aktivního PDF bez přenosu hodnot do jiného dokumentu.
- **T-30:** stovky polí, opakované widgety a delší české popisky se zobrazí bez blokování aplikace nebo ztráty synchronizace.

**Stav:** F-Q5 schváleno, nástroj F1 stále čeká na implementaci. F2/F3 neschváleny k implementaci. F-Q6 schváleno: vizuální pořadí podle stránek; pořadí Tab zůstává nezávislé.


---

## 15. F-Q6 – souhrn rozhodnutí a akceptační testy (21. 11. 2026)

**Schváleno A:** seznam v pravém panelu řadit podle **aktuálního pořadí stránek a umístění na stránce**. **Neměnit a nepřepisovat Tab order definované PDF.** Tento bod doplňuje F-Q5, neovlivňuje chování při otevření F-Q2, zobrazení polí F-Q4 ani rozsah typografie F2 stanovený v F-Q3.

- **T-31:** dvě stránky s poli zadanými ve strukturálním pořadí odlišném od jejich vizuální pozice → pravý seznam ukáže správné stránky a pořadí shora dolů/zleva doprava, ale Tab prochází původně definovanou sekvenci.
- **T-32:** otevření a rozbalení seznamu, změna zoomu a DPI → seřazení se samovolně nemění, PDF zůstává bez příznaku změny.
- **T-33:** po přetažení stránky na jiné místo (existujícím nástrojem Pořadí stránek) se seznam aktualizuje podle aktuálního pořadí, aniž se změní hodnoty formulářových polí.
- **T-34:** formulář s radio skupinou a s jedním polem na více stránkách → jeden datový model hodnot, korektní odkazy na jednotlivé viditelné výskyty, stabilní řazení.
- **T-35:** otočená stránka, skrytý widget a dvě různé položky ve stejné pozici → logické vizuální pořadí, skryté pole zřetelně oddělené, deterministické pomocné řazení.

**Navazující rozhodnutí F-Q7:** uložení neúplných formulářů bylo později **schváleno variantou A**; úplná pravidla jsou v oddílu 16.


---

## 16. F-Q7 – SCHVÁLENO (21. 11. 2026): upozornit na neúplný formulář, ale umožnit uložení (varianta A)

**Závazné rozhodnutí:** pokud uživatel ukládá **rozpracovaný AcroForm** a některá skutečně **povinná** pole nejsou vyplněna, PDF Blacksmith **upozorní na neúplnost**, umožní přejít k chybějícím polím, ale **vždy nabídne možnost pokračovat v uložení**. Uložení rozpracované práce **není podmíněno finálním dokončením formuláře**. Rozhodnutí platí pro `Uložit`, `Ctrl+S` a `Uložit upravenou kopii…`, pokud se na dokument vztahují běžná oprávnění k zápisu a nejsou jiné technické překážky.

### Funkční pravidla F1

1. Kontrola sleduje pole podle jejich skutečného příznaku **Required** (`/Ff`, včetně dědičnosti atributů), nikoli podle textu popisku či barvy zvýraznění. Prázdnou hodnotu vyhodnocovat podle typu (text, checkbox, radio, combo/listbox); nepoužívat jen počet widgetů. Dvě vizuální instance stejného logického pole nesmějí vyprodukovat dvě samostatná upozornění.
2. **Při explicitním příkazu k uložení** a zjištění nevyplněných povinných polí zobrazit srozumitelné upozornění, např. **„Formulář není dokončen: chybí 3 povinná pole.“**, a možnost **„Přejít na chybějící pole“**, **„Uložit přesto“**, **„Zrušit“**. Počet je skutečný, není pevně daný. Kliknutí na chybějící pole otevře jeho stránku a odpovídající položku v již schváleném přehledu F-Q5/F-Q6.
3. **„Uložit přesto“** dokončí běžný bezpečný postup uložení se všemi dosud vyplněnými hodnotami a dosavadním stavem povinných polí. Nesmí automaticky doplňovat výchozí odpovědi, vyplnit údaje domněnkou, měnit nastavení Required, odstraňovat prázdná pole, převádět formulář na statický text ani zahazovat jiné anotace.
4. **„Přejít na chybějící pole“** odloží potvrzení uložení a umožní uživateli formulář doplnit. Zůstávají zachovány všechny jeho rozpracované hodnoty a pracovní karta. Zrušení upozornění nebo samotného uložení nemění dokument.
5. **Prázdná nepovinná pole nezakládají upozornění.** Zvlášť ošetřit povinná pole `ReadOnly`, skrytá, nepřístupná či nekompatibilní s aktuální implementací: transparentně uvést, že je nelze v aplikaci běžně doplnit, ale nesnažit se je svévolně odblokovat. I pak má uživatel možnost uložit rozpracovaný dokument, pokud je zápis jinak bezpečný.
6. **AutoRecover nesmí blokovat modálním upozorněním** kvůli neúplným polím. Uloží obnovitelný pracovní stav s hodnotami; varování patří do explicitní interakce uživatele. Pouhé prohlédnutí seznamu nevyplněných položek nepředstavuje obsahovou změnu ani položku historie.
7. **Jiné bezpečnostní podmínky nejsou F-Q7 zrušeny.** Potenciální zneplatnění digitálních podpisů, zákaz editace, XFA, selhání zápisu či nepodporovaná struktura PDF mohou vyžadovat zvláštní postup, varování nebo odmítnutí operace. Schválení možnosti uložit neúplný formulář neznamená možnost překročit oprávnění PDF ani slíbit uložení po technickém selhání.
8. **Odlišovat pracovní uložení od odeslání/finální validace.** Případné budoucí akce „Odeslat“, „Potvrdit dokončení“ nebo pokročilé kontroly ve F2/F3 mohou mít jiná pravidla, ale nejsou součástí tohoto schválení. Nevykonávat neprověřené PDF JavaScript validátory.
9. Po úspěšném uložení a opětovném otevření se zachovají vyplněné i nevyplněné hodnoty a jejich appearance streams; dokument nesmí nepravdivě hlásit, že je celý formulář dokončený. **Uložit upravenou kopii…** se řídí dříve schválenou výměnou souboru v **téže kartě až po ověřeném úspěšném zápisu**.

### Akceptační scénáře navíc

- **T-36:** formulář se třemi prázdnými povinnými poli → `Ctrl+S` zobrazí upozornění na počet chybějících polí; volba **Uložit přesto** bezpečně uloží hodnoty bez jejich automatického doplnění.
- **T-37:** volba **Přejít na chybějící pole** přesune náhled i seznam F-Q5 k příslušnému poli, nezahodí rozpracované změny a uložení ještě neprovede.
- **T-38:** formulář s neprázdnými povinnými poli a s prázdnými nepovinnými poli → `Uložit` nepřidává varování o povinných polích; následné opětovné otevření zachovává stav.
- **T-39:** povinné pole se dvěma widgety se započítá jen jednou; nenaplněná radiogroup se vyhodnotí jako **jedna** chybějící logická odpověď; pole ReadOnly/skrytá mají srozumitelný stav.
- **T-40:** AutoRecover v neúplném formuláři proběhne bez blokujícího dialogu; návrat po pádu nabídne obnovené vyplněné hodnoty i prázdná povinná pole.
- **T-41:** `Uložit upravenou kopii…` po potvrzení „Uložit přesto“ přepne aktivní kartu na kopii **jen po úspěchu**; při zrušení, chybě zápisu nebo ohrožení podpisu zachová pracovní data a respektuje dosavadní ochranná pravidla.

**Stav:** F-Q7 schváleno jako pravidlo F1. F-Q1–F-Q6 zůstávají beze změny; F2/F3 tímto schválením nejsou uvedeny do implementace.


---

## 17. F-Q8 – SCHVÁLENO (21. 11. 2026): validace až po opuštění pole, souhrn při ukládání (varianta A)

**Závazné rozhodnutí:** v aktivním nástroji **Formuláře → Vyplnit** se stav každého pole (prázdné/vyplněné, povinné/volitelné, dostupné/pouze pro čtení) **průběžně aktualizuje** v přehledu F-Q5. **Chybový text k nesprávnému formátu se poprvé zobrazí až po opuštění pole** (ztráta fokusu: klik do jiného pole, Tab, případně potvrzení editace), nikoli po každém napsaném znaku. **Při ručním ukládání** se znovu vyhodnotí ověřitelná pravidla a zobrazí **souhrn chyb a nevyplněných povinných polí** s možností přejít na konkrétní pole. Bezpečné pracovní uložení zůstává možné podle F-Q7.

### Chování a omezení F1

1. **Živá indikace bez rušivých hlášení.** Při změně hodnoty průběžně aktualizovat stav vyplnění příslušného logického pole a souvisejících widgetů. Když uživatel teprve píše nedokončenou hodnotu (např. datum nebo e-mail), neposílat při každém znaku výstražné dialogy, neoznačovat pole předčasně za chybný formát a nepřerušovat zadávání. Stav přehledu může rozlišovat „právě upravováno“ / „dosud neověřeno“.
2. **Kontrola při opuštění pole.** Po ztrátě fokusu provést jen ty kontroly, které lze **spolehlivě odvodit** z podporovaného typu pole, platných deklarativních omezení PDF nebo bezpečně známého formátovacího pravidla. Zobrazení chyby má být **neblokující a srozumitelné**, s určením pole a doporučenou opravou; nevnucovat opravu ani nenahrazovat vstup. Při návratu do pole nechávat původní hodnotu editovatelnou.
3. **Nepředstírat validaci neznámých pravidel.** Nespouštět žádný libovolný PDF JavaScript, akce dokumentu nebo externí URL kvůli validaci. Pravidla uložená v nespuštěném skriptu **nepovažovat automaticky za úspěšně zkontrolovaná**; zobrazení „Ověření nepodporovaného pravidla není k dispozici“ je vhodnější než falešný výsledek. E-mail či datum kontrolovat pouze, pokud je k nim bezpečně zjistitelná konkrétní formátová podmínka (nikoli podle názvu nebo obsahu pole); ostatní volný text není chybně označován jen odhadem.
4. **Souhrn při ručním ukládání.** Pro `Ctrl+S`, `Uložit` a `Uložit upravenou kopii…` spustit aktuální kontrolu všech relevantních polí, seskupit **chybné formáty**, **prázdná povinná pole** a **neověřitelná pravidla** do rozlišitelných kategorií. Logické pole se dvěma widgety počítat jen jednou. V seznamu F-Q5 ukázat status a umět přejít na stránku/pole. **Neověřitelné pravidlo není samo o sobě chyba hodnoty**.
5. **Tlačítka a soulad s F-Q7.** V případě zjištěných problémů nabídnout **„Přejít na pole“**, **„Uložit přesto“**, **„Zrušit“**. „Uložit přesto“ provede stejný bezpečný postup jako F-Q7, včetně validace samotného zápisu a ochran podepsaných/XFA/omezených PDF. „Přejít na pole“ odloží uložení a vrátí pozornost do daného pole, „Zrušit“ odmítne pouze aktuální uložení. Varování o formátu nesmí zabránit vytvoření **rozpracované kopie**.
6. **AutoRecover.** Během automatické obnovovací zálohy zobrazovat **žádné modální validační dialogy**. Uložit poslední bezpečně obnovitelnou pracovní hodnotu, včetně případně nevalidního rozpracovaného vstupu, bez automatické opravy. Uvnitř téhož dokumentu zachovat stav historie Ctrl+Z/Y a význam příznaku neuložených změn: změna hodnoty je úprava, samotný výsledek kontroly není úprava.
7. **Výjimečné ukončení.** Pokud uživatel přepne kartu nebo nástroj uprostřed práce s polem, bezpečně dokončit/převzít rozpracovanou hodnotu podle schváleného životního cyklu editace a při vhodném okamžiku ji vyhodnotit; bez ztráty textu. Při duálním zobrazení ani během upozornění nezasahovat do hodnot druhého PDF.
8. **Vztah k F2.** Uživatelské definování vlastních validačních schémat či změna typografie existujících polí není součástí F1. Základní interoperabilita vzhledových streamů, písem, seznamových voleb a originálního formátování zůstává zachována. F2/F3 zůstávají jen plánem.
9. **Bezpečnostní výjimky.** Volba „Uložit přesto“ neobchází skutečná omezení PDF (práva, podpisy, formát, selhání diskového zápisu). Pokud je třeba upozornit na zneplatnění podpisu, toto bezpečnostní upozornění může vyžadovat vlastní potvrzení; nelze ho potichu přeskočit.

### Akceptační testy T-42 až T-49

- **T-42:** při postupném zadávání neúplného e-mailu/datu s explicitně známým omezením se aktualizuje stav vyplnění, **neobjevují se okamžitá chybová upozornění po každém znaku**.
- **T-43:** po `Tab`/kliknutí mimo pole s nesprávnou hodnotou se zobrazí **neblokující a správně označené** upozornění; uživatel může hodnotu opravit bez ztráty znaků.
- **T-44:** oprava hodnoty a opuštění pole odstraní zastaralou chybu; nezůstanou duplicitní upozornění. Samotné přepnutí řádku v seznamu neoznačuje PDF za upravené.
- **T-45:** `Ctrl+S` při současném chybějícím Required a chybné hodnotě nabídne **jeden přehledný souhrn** s možnostmi Přejít na pole, Uložit přesto, Zrušit; klik na položku přejde ke správnému widgetu.
- **T-46:** **Uložit přesto** zachová i rozpracované hodnoty (včetně formátově neplatné) a prázdná povinná pole; při otevření kopie se hodnoty načtou beze změny a varování lze znovu zkontrolovat.
- **T-47:** dokument s pravidlem jen v nepodporovaném JavaScriptu se automaticky nevyhodnotí jako bezchybný ani se v něm JavaScript nespustí; stav je „nelze ověřit“ a obyčejné pracovní uložení není automaticky zablokováno.
- **T-48:** AutoRecover během rozepsaného pole proběhne bez modálního dialogu; obnova zachová rozepsanou hodnotu a dokumentovou historii v rozsahu již schválené podpory.
- **T-49:** dvě vizuální instance téhož logického pole / radio skupina nevytvoří duplicitní hlášení; přepnutí karty v duálním pohledu zobrazí validaci jen aktuálního PDF, bez zásahu do druhého.

**Stav:** F-Q8 schváleno jako F1, testy T-42 až T-49 požadovány. Všechny body F-Q1 až F-Q7 zůstávají platné; F2/F3 neimplementovat na základě tohoto rozhodnutí. Zdrojový kód aplikace nebyl v rámci této aktualizace změněn.

---

## 18. F-Q9 – SCHVÁLENO (21. 11. 2026), varianta A: zvýraznění ověřených chyb

**Rozhodnutí uživatele:** V aktivním režimu **Formuláře → Vyplnit** se formulářové pole s **prokazatelně nesprávným formátem** po **opuštění pole** jemně **červeně orámuje na stránce** a současně získá srozumitelný chybový stav v **sbalitelném přehledu polí v pravém panelu**. **Jakmile uživatel chybnou hodnotu opraví a dokončí kontrolu, upozornění v obou místech zmizí.** Jde o dočasnou vrstvu UI, nikoli o změnu samotného PDF.

### Přesná pravidla propojení F-Q4, F-Q7 a F-Q8

1. **Načasování:** při průběžném psaní se aktualizuje informace o vyplnění, ale nevzniká nové červené orámování kvůli teprve rozepsané hodnotě. Validaci spustit **po opuštění pole**; u jednotlivých polí se zachová stav vyhodnocení a případná dříve nalezená chyba. Nová kontrola při ručním ukládání přepočítá problémy a zvýrazní prokázané chyby i v pravém seznamu, nezpůsobí však další duplicitní modální dialog.
2. **Dvě odlišné vizuální pomůcky:** běžné vínové zvýraznění vyplnitelných polí podle F-Q4 je odlišné od červeného chybového orámování. Přepínač **„Zvýraznit vyplnitelná pole“** vypíná pouze běžné podbarvení; chybové orámování ověřeného problému zůstává při aktivním nástroji Formuláře viditelné, aby nešlo chybu omylem skrýt. Při opuštění nástroje Formuláře se **obě** pomůcky na stránce skryjí. Stav v přehledu polí se uchová pro návrat do Formulářů.
3. **Význam stavů:** formátová chyba = pouze pravidlo, které aplikace umí **bezpečně a jednoznačně ověřit** (F-Q8). Chybějící povinná hodnota má samostatné vysvětlení; vyznačit ji podle schválené kontroly, typicky po opuštění pole či při pokusu o uložení, **nikoli jako chybu hned po otevření prázdného formuláře**. Neověřitelný PDF JavaScript či neznámé pravidlo = „Nelze ověřit“, **nikoli červený obrys a nikoli falešné „Bez chyby“**.
4. **Zobrazení:** jemný, ale čitelný obrys okolo geometrie widgetu, nikoli přemalování jeho hodnoty nebo překrytí výběrových úchytů. Zobrazení se musí přizpůsobit zoomu, rotaci, DPI, ořezu a zobrazení dvou PDF. Systém musí rozlišovat barevné a aktivní/fokusové obrysy; **chybový stav nikdy nekomunikovat pouze červenou barvou** – také symbol, dostupný text a popis ve správném řádku seznamu.
5. **Oprava:** po změně hodnoty a potvrzení/opustění pole znovu vyhodnotit ověřitelná pravidla; při úspěchu odstranit červený obrys i příslušný chybový stav. V několika widgetech jednoho logického pole udržovat jediný zdroj hodnoty a konzistentní stav.
6. **Úložiště a výstup:** obrys existuje **pouze ve vrstvě obrazovkového UI**; nezapisovat ho do `/Annots`, do obsahu stránek, do appearance streamů ani do exportu či tisku. Pouhé výpočty validace nevytvářejí historii Ctrl+Z/Y ani příznak neuložených úprav. Zůstává zachováno **„Uložit přesto“** (F-Q7/F-Q8) a neblokovaný AutoRecover; skutečná bezpečnostní omezení dokumentu nadále platí.
7. **Izolace nástrojů:** mimo Formuláře → Vyplnit se vizuální validace na stránce nenačítá ani nezachytává kliknutí; neovlivňuje Text, Objekty, Poznámky, Separace ani stávající hledání.


### Akceptační testy F-Q9 – T-50 až T-57

- **T-50:** V režimu Formuláře → Vyplnit se při chybném formátu **během psaní** neobjeví chybový červený obrys; po opuštění pole (Tab, klik mimo, potvrzení) se obrys a položka s chybou v pravém přehledu zobrazí synchronně.
- **T-51:** Oprava hodnoty a následné opuštění pole odstraní červený obrys i indikaci chyby; staré výsledky validace se neponechají ani nezdvojí.
- **T-52:** Přepínač „Zvýraznit vyplnitelná pole“ vypne vínovou výplň běžných polí, **nikoli** nezávislou výstrahu ověřené chyby; po přepnutí na jiný nástroj nejsou zobrazovány žádné výstražné obrysy formulářů.
- **T-53:** Uložení rozpracovaného PDF, Uložit upravenou kopii, tisk ani export neobsahují pomocný červený obrys, pokud jej zvlášť nevytvořil uživatel jako vlastní obsah PDF; pouze validační UI nesmí vzniknout jako anotace ani modifikace dokumentu.
- **T-54:** Červený obrys a značka chyby se drží správného widgetu při zoomu, otočení, DPI, posunu a duálním zobrazení; nezasáhnou do druhého dokumentu.
- **T-55:** Ruční uložení s formátovou chybou a chybějícím povinným polem zobrazí **jediný souhrn** podle F-Q7/F-Q8 a ponechá tlačítko „Uložit přesto“ funkční; nevalidovaná pravidla jsou rozlišena od prokázaných chyb.
- **T-56:** Červené orámování není jedinou informací o chybě: stav je textově/symbolicky odlišen v přehledu, ovladatelný klávesnicí; ReadOnly či nepřístupné widgety se nesmějí jevit jako editovatelné.
- **T-57:** Dvě zobrazení jednoho logického pole ukazují stav konzistentně; zobrazení chyby nemění hodnotu `/V`, PDF streamy, historii Ctrl+Z/Y ani příznak neuloženého souboru; AutoRecover probíhá bez modálního dialogu.

**Stav:** F-Q9 je schváleným upřesněním F1. Formuláře F2/F3 zůstávají plánované, nikoli autorizované k implementaci. Žádné změny v repozitáři PDF Blacksmith nebyly provedeny tímto dokumentem.