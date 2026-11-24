# PDF Blacksmith – specifikace oprav a rozšíření pro Codex

**Verze:** 5.0 – F-Q9 schváleno (21. 11. 2026); nejnovější pracovní upřesnění UI v3.9 (24. 11. 2026)
**Nejnovější UI doplnění v3.13 (24. 11. 2026):** [UI Pomůcky rozvržení a vlastní zkratky](rozhrani/UI_POMUCKY_ZKRATKY_v3.13.md), [HTML v3.13](rozhrani/cely_vizual_v3.13.html), [mapa zkratek v1.1](mapa_klavesovych_zkratek_v1.1.md). Pravítka jsou vně listu; v Zobrazení přibyl podnástroj s vlastním panelem nastavení; Nastavení umožňuje demo vlastní klávesové přiřazení s varováním při kolizi. Vše je jen návrh UI, ne povolení automatické implementace.\n**Poslední úprava ikon (24. 11. 2026):** [HTML v3.12](rozhrani/cely_vizual_v3.12.html) a [mapa ikon v4.8](mapa_ikon_v4.8.md) nyní používají přímo uživatelem nakreslený podkovový magnet ze SVG; vyplněné tělo a obrysové póly zachovávají geometrii předlohy, barva je výhradně `currentColor` podle ostatních ikon. Jiné funkce v3.11 se nemění.\n**Nejnovější doplněk UI (24. 11. 2026):** [UI v3.11 – Separace, pomocné panely, vodítka a snap](rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md), [HTML v3.11](rozhrani/cely_vizual_v3.11.html), [mapa ikon v4.7](mapa_ikon_v4.7.md), [mapa klávesových zkratek v1.0](mapa_klavesovych_zkratek_v1.0.md). Starší upřesnění zůstávají v platnosti, pokud nejsou novější verzí výslovně změněna. **Nejde o povolení hromadné implementace.**
**Aktuální UI a pokyny pro Codex (24. 11. 2026):** [aktuální HTML v3.10](rozhrani/cely_vizual_v3.10.html), [doplněk Přichytávání v3.10](rozhrani/UI_PRICHYTAVANI_v3.10.md), [mapa ikon v4.6](mapa_ikon_v4.6.md), [předchozí úplná konsolidace UI v3.9](rozhrani/AKTUALNI_UI_CODEX_v3.9.md). Doplněk v3.10 mění pouze magnet, vodítka a menu „Přichytávat k:“; ostatní schválené chování v3.9 a původní bezpečnostní pravidla zůstávají platná.\n**Zdroj:** https://github.com/Liborias/PDF-Blacksmith (`main`)  
**Stav:** NEZADÁVAT HROMADNOU IMPLEMENTACI bez schválení sporných rozhodnutí.  
**Zaměření:** Windows desktop; stávající rozhraní Go/Win32, PDFium, Python/Ghostscript.
**Doplněk pro Text a rozhraní (22. 11. 2026):** [Bezkolizní výběr a vnořené záložky Výběr/Typografie](rozhrani/TEXT_VYBER_A_ZALOZKY_v3.1.md). Jde o pracovní UI upřesnění; zejména **nepřepisovat schválené SEL-01 až SEL-03 a TXT-01 až TXT-04** zavedením kliknutí pro výběr celého bloku nebo dvojkliku k otevření editace. Rozdělení záložek se týká pouze panelu Text, nikoli Separací či globálního Nastavení.
**Aktuálnější vizuální doplnění Text v3.2:** [Jedna řada ikon, otazník napravo a skrytá nápověda](rozhrani/TEXT_VYBER_NAPOVEDA_v3.2.md). Mění **pouze rozmístění** ikon Text; neaktivuje nové režimy, nepřemapovává prostřední tažení, dvojklik ani objektové laso.\n
**Předchozí pravidla Textu v3.3 (23. 11. 2026; novější změny viz UI v3.9):** [TEXT_VYBER_A_SMISENA_TYPOGRAFIE_v3.3.md](rozhrani/TEXT_VYBER_A_SMISENA_TYPOGRAFIE_v3.3.md). Závazné uživatelské upřesnění: I-kurzor může na volné stránce zakládat bodový text, samostatný nástroj textového bloku vymezuje pouze odstavcový rámec; aktivně zvolené obdélníkové textové označení a textové laso mají svůj vlastní význam pro **levá** gesta, aniž by jakkoli měnila schválený globální obdélník stiskem kolečka. Rozšířený výběr je nápověda ke klávesám a klikání, nikoliv samostatná výběrová ikona. Smíšené typografické hodnoty mají neurčitý stav; A+/A− působí relativně na každý běh, zadaná konkrétní hodnota sjednotí pouze dotčenou vlastnost. Tato pravidla **nahrazují pouze rozporné návrhy v3.1/v3.2**, nikoliv ostatní schválené SEL/TXT/OBJ, Q31, formuláře nebo Separace. Žádná automatická implementace produkčního kódu.
**Schválené rozhodnutí Q1:** **Prostřední tlačítko (stisk kolečka) slouží pouze ke stisknutí a tažení obdélníkového výběrového rámečku na plátně PDF.** Samostatné, dvojité i modifikované kliknutí prostředním tlačítkem není výběrovým příkazem. **Levé tlačítko** zajišťuje běžné klikací výběry objektů, textu, sloupců a miniatur, přímou editaci, laso a manipulaci, podobně jako v běžných textových a grafických editorech. Nenarušit scrollování kolečkem bez stisku ani zoom `Ctrl + kolečko`.  
**Schválené rozhodnutí Q3:** **`Ctrl+Z` vrací také kroky výběru textu; `Ctrl+Y` je znovu provádí.** Výběrové operace a obsahové úpravy používají jednu chronologickou historii v rámci konkrétního dokumentu. Výběr sám nesmí změnit obsah PDF ani příznak neuložených změn.  
**Schválené rozhodnutí Q4:** **Kreslení tvarů levým tažením: bez modifikátoru = libovolný obdélník/elipsa; `Shift` = čtverec/kruh; `Ctrl` = kreslení od středu; `Ctrl+Shift` = čtverec/kruh od středu.** Tato konvence platí pouze při použití nástroje kreslení, nepřebírá se doslovně z Inkscape ani Illustratoru.
**Rozhodnutí Q5 – cílová interakce i fallback SCHVÁLENY, konkrétní kompatibilitu ověřit:** Přetahování textu/objektů/stran mimo aplikaci má **vizuálně působit jako obyčejný drag-and-drop s preferovanou nabídkou formátu až po uvolnění myši**. **Prioritní cíle: CorelDRAW, Adobe Photoshop, Affinity (Designer/Photo) a další editory běžně přijímající soubory z Průzkumníka.** Primárně se pokusit o skutečné vložení kompatibilního formátu; **pokud jej cílová aplikace odmítne nebo jej nelze spolehlivě provést, zobrazit zprávu a nabídnout „Uložit jako…“ do uživatelem vybrané složky**, případně také kopírování do schránky. Dočasná data mohou být v paměti či `%TEMP%`. Nepředstírat úspěšné vložení. Volba *až po uvolnění* u cizích aplikací zůstává cílem prototypu, nikoli již potvrzenou schopností OLE. Podrobnosti DND-00 až DND-03 a DOC-04.  
**Schválené rozhodnutí Q6:** **6 % šířky klientské oblasti okna aplikace** se vztahuje k **samotnému vykreslenému obrázku miniatury stránky**, nikoliv k celému panelu. Panel musí mít navíc prostor pro vnitřní okraje, číslování a posuvník. Při změně velikosti okna se miniatury automaticky přepočítají, dokument dostane zbývající prostor; detaily UI-08.
**Schválené rozhodnutí Q7:** Po sbalení hlavního **i** pomocného panelu zůstává **úzká svislá lišta** (varianta B), nikoli samotné volně plovoucí ouško. **Hlavní panely:** malé zaoblené čtvercové ouško pouze se směrovou šipkou, bez textu „Levá strana“/„Pravá strana“. **Pomocné panely:** ouško s viditelným textovým názvem panelu ve sbaleném i rozbaleném stavu. Ouško **otevřeného/aktivního pomocného panelu** má totožný podklad jako panel a opticky na něj navazuje; ouška **sbalených/neaktivních** panelů mají odlišný, tlumenější odstín. Podrobnosti UI-04.  
**Schválené rozhodnutí Q9:** **Podrobnější nápověda pod skupinou ikon primárně popisuje aktivní nástroj. Při najetí na jinou ikonu se dočasně přepne na tuto ikonu; po opuštění ikony se automaticky vrátí k aktivnímu nástroji.** Krátký tooltip zůstává samostatný a uvádí název a zkratku. Platí společně pro všechny skupiny nástrojů. Viz UI-07.
**Schválené rozhodnutí Q8:** **Oficiální název aplikace je vždy `PDF Blacksmith` (s písmenem „c“).** `PDF Blaksmith` je historický překlep a nesmí se používat v nových názvech, uživatelských textech, ikonách s názvem ani distribučních artefaktech. Podrobnosti BRAND-03.  
**Schválené rozhodnutí Q10 – Nastavení a Nápověda v horní liště:** Přidat **ikonu ozubeného kolečka (Nastavení)** a **ikonu otazníku (Nápověda)**. Globální Nastavení se otevře v **samostatném okně s kategoriemi v levém sloupci a jejich obsahem vpravo**, nikoli v panelu aktivního nástroje. Povinné jsou **verze aplikace, informace o vydání a uživatelské mapování klávesových zkratek**. Osm kategorií globálního Nastavení je **schváleno (Q14)**; konkrétní jednotlivé volby se implementují postupně a zobrazí se pouze funkční možnosti. Modalita je závazně **modální** (Q13). Viz UI-09 až UI-11.
**Schválené rozhodnutí Q12 – potvrzování Nastavení:** V samostatném okně globálního Nastavení jsou tlačítka **OK**, **Použít** a **Zrušit**. `OK` potvrdí čekající změny a zavře okno; `Použít` je potvrdí a ponechá Nastavení otevřené; `Zrušit` zahodí pouze nepotvrzené změny od posledního `Použít` (nevrací již potvrzené), obnoví případný pouze dočasný náhled a zavře okno. Podrobnosti UI-10. **Modalita okna je SCHVÁLENA jako modální (Q13); osm kategorií je SCHVÁLENO (Q14), jednotlivé konkrétní možnosti zůstávají postupným implementačním úkolem.**  
**Schválené rozhodnutí Q15 – inteligentní výběr objektů:** Při běžném **levém kliknutí** se kandidáti určují podle skutečné geometrie, viditelnosti a pořadí kreslení, **nikoliv pouze podle jejich obdélníkových hranic**. Přesně zasažená viditelná kontura nebo vyplněná plocha má přednost před vzdáleným velkým objektem v pozadí; při srovnatelném zásahu rozhoduje pořadí vykreslování. Vektorové tahy berou v úvahu skutečnou tloušťku, bitmapy ořez a alfa průhlednost. **Výchozí tolerance je přibližně 5 obrazových pixelů**, s konzistentním chováním při zoomu a HiDPI. `Ctrl + levý klik` v režimu Objekty cyklicky prochází geometricky zasažené kandidáty směrem do hloubky a po posledním se vrací k prvnímu. Tyto podmínky mají být regresně testovány; podrobnosti OBJ-02, OBJ-03.  
**Schválené rozhodnutí Q16 – režimy lasa:** Nástroj **Laso** má dva režimy s přepínačem v panelu nástroje: **A „Protnutí“ (výchozí)** označí každý objekt, jehož skutečně viditelná kresba se protíná s oblastí lasa; **B „Celý objekt“** označí pouze objekty, jejichž veškerá viditelná kresba leží uvnitř lasa. Rozhodovat podle geometrie, ořezu a skutečně viditelných tahů, výplní a bitmap, nikoli podle středu či pouhého bounding boxu. Výběr potvrzuje levý dvojklik, levý klik poblíž počátku nebo `Enter`; `Esc` rozpracované laso ruší. Stisk kolečka + tažení vyznačuje i v nástroji Laso obdélník. Viz OBJ-01.  
**Schválené rozhodnutí Q17 – potvrzení přesunu stránek:** Po dokončení levého tažení miniatur nebo po zadání cílové pozice zobrazit **vždy před provedením skutečné změny** dialog „Potvrdit přesunutí stránek“ s **počtem vybraných stránek, jejich čísly či přehledem, původním umístěním a přesnou cílovou pozicí v novém pořadí**. Volby **Přesunout / Zrušit**, bez „příště se neptat“. Při zrušení zůstává dokument, pořadí a historie beze změny; potvrzení vytvoří jediný vratný krok pro `Ctrl+Z` / `Ctrl+Y`. Pokud se pořadí nemění, neprovádět operaci ani nenabízet zbytečné potvrzení. Viz DOC-02.  
**Schválené rozhodnutí Q18 – Uložit upravenou kopii = varianta B (UPŘESNĚNO):** Příkaz **„Uložit upravenou kopii…“** vytvoří nový PDF soubor bez zásahu do původního souboru. **Teprve po úspěšném zápisu a ověření nové kopie se původní PDF zavře a jeho dosavadní karta se nahradí otevřenou kopií. Původní dokument nesmí zůstat v aplikaci otevřený jako druhá karta.** Další `Uložit` / `Ctrl+S` ukládá výhradně novou kopii. Při selhání nebo stornu zůstane původní pracovní karta včetně rozpracovaných změn otevřená. Ostatní nezúčastněné karty zůstávají beze změny. Viz DOC-03.  
**Schválené rozhodnutí Q20 – zavírání dokumentů:** Při duálním zobrazení nabídnout „Zavřít levý dokument / Zavřít pravý dokument / Zavřít všechny dokumenty / Zrušit“. V jediné otevřené kartě nabídnout „Zavřít aktuální dokument / Zrušit“; při více kartách lze navíc zavřít všechny. **Při zavírání více upravených dokumentů zobrazit jediný souhrnný modální dialog**, v němž má každý soubor vlastní volbu **Uložit / Uložit upravenou kopii… / Neuložit** a celý dialog **Pokračovat / Zrušit**. Neuložit není předvybráno jako destruktivní výchozí akce; všechny skutečně potřebné zápisy bezpečně dokončit před zavřením karet. Při chybě ukládání se zpracování přeruší a nedotčená rozpracovaná data i AutoRecover zůstanou zachována. V případě Uložit upravenou kopii… zachovat pravidlo Q18: nejprve úspěšně vytvořit kopii a otevřít ji ve stávající kartě, teprve poté uzavřít původní kontext. Viz DOC-05.  
**Schválené rozhodnutí Q28 – náhled tiskových separací:** V nástroji **Separace** po načtení PDF zobrazit **všechny skutečně dostupné tiskové desky** (CMYK a případné přímé barvy) jako výchozí **zapnuté**. U každé desky přepínač její viditelnosti a příkaz **„Jen tato“** (izolace jediné desky); společný příkaz **„Zobrazit všechny“**. Vypínání a zapínání desek mění **výhradně obrazovkový náhled**, nikoli PDF, separační data, tisk ani export. Výchozí plný barevný souběh; podporovat korektní simulaci přetisku, průhledností a barevného řízení, navíc adaptivní detail při zoomu dle Q21. Názvy přímých barev a existence kanálů zjišťovat z reálného dokumentu – **nevymýšlet ani natvrdo nezobrazovat PANTONE**; nepodporované přesné separace nesmí být prezentovány jako věrný náhled. Detailní pravidla a akceptace v RND-02.  
**Schválené rozhodnutí Q29 – podoba izolované tiskové desky:** Pokud je v náhledu Separace zapnutá **právě jedna** tisková deska (např. po „Jen tato“), zobrazit ji **ve výchozím stavu v odstínech šedi**, aby bylo snadné kontrolovat pokrytí barvou. Umožnit přepnutí na náhled **v tiskové barvě** vybrané desky, včetně správně simulované přímé barvy. **Nulové pokrytí = bílý/netisknutý podklad; vyšší pokrytí = tmavší šedá, 100 % = černá v šedotónovém pohledu**. Nejde o převod PDF na grayscale ani o skutečnou úpravu separací. Plný souběh více desek zůstává barevný a používá správu barev podle Q28. Podrobnosti RND-02.  

**Schválené rozhodnutí Q30 – měření tiskového pokrytí v nástroji Separace:** Zobrazovat pod kurzorem **procentuální pokrytí každé reálně dostupné tiskové desky** (C, M, Y, K a samostatně pojmenované přímé barvy) a **TAC jako součet procesních CMYK hodnot** na daném místě, nikoliv součet spot barev. Umožnit konfigurovatelný limit TAC (výchozí návrhová hodnota 300 %, podle tiskové podmínky uživatelem změnitelná) a volitelný vizuální překryv/zvýraznění přesahů, pouze v náhledu. Bez falešné přesnosti: výpočet musí odpovídat separovanému tiskovému výsledku se správou barev, overprintem a průhledností; pokud je pouze přibližný nebo nedostupný, jasně uvést omezení. **Jde o ROZŠÍŘENÍ existujícího nástroje, nikoli náhradu jeho současných funkcí.** Viz RND-02.  
**Schválené rozhodnutí Q32 – JEDINÁ grafická změna současného panelu Separace:** U každé tiskové desky v **existujícím seznamu** přidat **malou barevnou identifikační značku** (C = azurová, M = purpurová, Y = žlutá, K = černá; u přímých barev orientační vzorek podle dostupné simulace). Značku umístit vedle dosavadního checkboxu a textového názvu, zachovat všechna stávající čísla `Plocha / max. tón`, ovládání, výšku a strukturu panelu, nabídky ICC i přepínače. **NEPŘESTAVOVAT panel do dvou záložek, šesti sbalitelných sekcí ani jiného nového rozložení**; ostatní grafické návrhy z v3.6/v3.7 uživatel nevybral. Funkční rozšíření schválená samostatně (Q28, Q29, Q30) platí dál, musí se ale přidat do stávajícího rozhraní bez vynuceného redesignu. Viz RND-03.  
**Schválené celoprojektové pravidlo Q31 – žádné zrušení původních funkcí:** Každou stávající funkci, ovladač, chování, předvolbu a exportní/tiskovou cestu dotčeného nástroje **zachovat, pokud není výslovně schválena její změna nebo odstranění**. Před úpravou inventarizovat existující schopnosti v kódu a běžícím UI, zachovat přístupnost a návaznosti; nové funkce integrovat do současného panelu a mechanik. Po změně provést regresní testy původních i přidaných funkcí. Zvlášť chránit stávající **Separace → Převod barev aktuálního dokumentu**, režimy DeviceGray, složená černá → K100, pouze K explicitní CMYK, ICC/profily, tiskový náhled, analýzy i export dle `COLOR_MANAGEMENT.md`. Žádné mazání nebo nahrazení zjednodušeným mockupem.  **Schválené rozhodnutí Q19 – automatická obnova (AutoRecover):** Automatické obnovovací snímky v `%LOCALAPPDATA%\PDFBlacksmith\Recovery`, oddělené od `%TEMP%` a běžných pracovních návrhů. Průběžně ukládat obsahové změny včetně rozpracovaného textu bez přepisu originálu či aktuálního pojmenovaného PDF. **Výchozí hodnoty: po 60 s nečinnosti, při souvislé práci nejpozději po 2 min, uchovávat 2 validní poslední snímky.** Po pádu nabídnout obnovu; po vědomém „Neuložit“ obnovovací data daného dokumentu odstranit. Po úspěšném uložení kopie navázat obnovovací relaci na kopii a uzavřít původní dokument. Podrobnosti DOC-06.  
**Schválené rozhodnutí Q21 – obrazovkové vykreslování po vzoru Adobe Acrobat:** Zachovat PDFium a implementovat **adaptivní překreslování z původního PDF**: při zoomu lze dočasně škálovat existující náhled kvůli odezvě, poté asynchronně vykreslit detail v odpovídajícím rozlišení; při vysokém zoomu použít výřezy/dlaždice a paměťově omezenou mezipaměť. **Volby v Nastavení → Zobrazení PDF:** vyhlazování textu, vyhlazování vektorových čar, volitelné zvýraznění tenkých čar, rychlé zobrazení během pohybu a vysoce kvalitní překreslení po zastavení. Jde **výhradně o obrazovkové zobrazení**: původní geometrie, tloušťky čar, barvy, separace, vektory a výsledný tisk se nesmějí měnit. Podrobnosti RND-01.  
**Schválený požadavek Q22 – nástroj Poznámky:** Přidat samostatný nástroj **Poznámky** pro **textové komentáře i grafické poznámky a kresby**, vizuálně oddělené od původního obsahu PDF a uložené spolu s PDF. Preferovat standardní **PDF anotace** (`/Annots`, nikoli prosté vrstvy OCG ani vypálení do `/Contents`) kvůli interoperabilitě s Adobe Acrobat a jinými editory. **Základní paleta a samostatný seznam poznámek jsou SCHVÁLENY jako Q25 pro první funkční verzi**; pokročilé odpovědi na komentáře, razítka a uživatelské symboly patří do backlogu. **Tisk bez poznámek (Q23) i export bez poznámek (Q24) jsou výchozí a SCHVÁLENÉ**; při konzultačním exportu lze poznámky výslovně zahrnout. Podrobnosti ANNO-01 až ANNO-04.  
**Schválené rozhodnutí Q23 – tisk anotací, varianta A:** **Ve výchozím stavu tisknout PDF bez poznámek** (textových i grafických); v tiskovém dialogu jasný volitelný přepínač **„Tisknout včetně poznámek“** (výchozí vypnutý). Výsledný tiskový náhled a tisk musí vždy odpovídat volbě; volba nesmí mazat poznámky z PDF, měnit jejich viditelnost na obrazovce ani ovlivňovat trvalý obsah, separace a barvové profily podkladového PDF. Při zapnutí respektovat pravidla příznaků tisknutelnosti anotací dle PDF; více v ANNO-04.  
**Schválené rozhodnutí Q24 – export PDF, varianta B:** **Výchozí export PDF je bez komentářů a grafických poznámek** pro čistý předtiskový výstup. V exportním dialogu zobrazit nepřehlédnutelnou volbu **„Zahrnout poznámky (editovatelné)“**, výchozí **vypnuto**. Zapnutí je určeno pro konzultaci, kontrolu a revize a zachovává anotace ve standardní editovatelné podobě. **Běžné Uložit / Ctrl+S / Uložit upravenou kopii… není export**, a proto poznámky v pracovním PDF vždy zachová, nejsou-li výslovně smazány. Čistý export nesmí přepsat aktivní zdrojové PDF, odebrat pracovní poznámky ani zplošťovat podkladové vektory a musí respektovat barvy, ICC, separace a overprint. Viz ANNO-04.  
**Schválené rozhodnutí Q25 – základní paleta nástroje Poznámky v první funkční verzi:** Šest skupin nástrojů: **Komentář** (`/Text`), **Textové pole** (`/FreeText`), **Volná kresba** (`/Ink`), **Čára / šipka** (`/Line`), **Obdélník / elipsa** (`/Square`, `/Circle`) a **Zvýraznění** (`/Highlight`); v textových značkách také **Podtržení** (`/Underline`) a **Přeškrtnutí** (`/StrikeOut`). U relevantních typů měnit barvu, tloušťku čáry, průhlednost a další dostupné vlastnosti. Součástí první funkční verze je **panel Poznámky se seznamem podle stránek**, navigací na označenou poznámku a možností její úpravy či odstranění. Pokročilé odpovědi na komentáře, razítka, vlastní symboly a další rozšíření až do backlogu. Při práci zachovat editovatelnost, Undo/Redo, AutoRecover, standardní `/Annots` a výchozí čistý tisk a export (Q23–Q24). **Výchozí zobrazení poznámek na obrazovce je SCHVÁLENO jako Q26, varianta A: po otevření PDF viditelné, s rychlým přepínačem „Zobrazit poznámky“.** Viz ANNO-01 až ANNO-04. **Rozsah přepínače je SCHVÁLEN jako Q27 (varianta A): každý dokument má vlastní nezávislý stav; přepnutí v jedné kartě nemění ostatní PDF, včetně duálního zobrazení.**  
**Schválené rozhodnutí Q14 – struktura Nastavení:** Schváleno osm kategorií: **Obecné, Rozhraní, Zobrazení PDF, Klávesové zkratky, Soubory, Text a OCR, Barvy a tisk, O aplikaci**. V návrhu rozhraní se počítá se všemi; v reálné aplikaci se kategorie zobrazí až ve chvíli, kdy obsahuje aspoň jednu skutečně fungující volbu (resp. u „O aplikaci“ platné informace). Nevytvářet nefunkční ovladače, prázdné stránky ani zavádějící položky. Schválení kategorie neznamená automaticky schválení všech navrhovaných funkcí v ní. Viz UI-10.  
**Schválené rozhodnutí Q11 – způsob otevření Nápovědy:** **Kliknutí na ikonu otazníku v horní liště otevře úvodní stránku celé uživatelské příručky**. **Klávesa `F1` otevře konkrétní kapitolu odpovídající právě aktivnímu nástroji nebo zaměřenému ovládacímu prvku či nastavení**, přednost má přesný kontext; při neznámém kontextu se bezpečně přejde na příručku aktivního nástroje nebo na úvod. Otevření Nápovědy nesmí změnit stav PDF. Zbývající funkce Nápovědy (fulltext v příručce, struktura kapitol, diagnostika aj.) jsou dosud návrhy. Viz UI-11.  
**Značení:** `BUG` = oprava, `UX` = chování/rozhraní, `FEATURE` = nová funkce, `TECH` = architektura/kvalita, `BACKLOG` = později.  
**Priority:** P0 = blokující nebo ochrana dat, P1 = důležité, P2 = vylepšení, P3 = později.

## Instrukce pro Codex – společná pravidla

1. Před změnou každého bodu prozkoumat současnou implementaci, ověřit pozorované chování na Windows, určit kořenovou příčinu a oznámit související moduly. **Samotné přečtení zdrojového kódu nepotvrzuje reprodukci chyby.**
2. **PŘÍSNÁ ZPĚTNÁ KOMPATIBILITA – SCHVÁLENO Q31:** Nezačínat nový nástroj „od nuly“, nenahrazovat existující mechaniky ani UI zjednodušeným prototypem, pokud není odstranění či změna konkrétní stávající funkce samostatně výslovně schválena. **Před každou úpravou udělat inventuru** existujících akcí, nastavení, klávesových zkratek, panelů, výstupů, pracovních postupů a hraničních případů v reálném UI a kódu; označit, co je zachováno, co je opravováno, co se přidává. Respektovat `PLAN.md` a `COLOR_MANAGEMENT.md`. Po změně doložit regresní test původních funkcí. **Žádné zhoršení či tiché odebrání již existující schopnosti.**
3. **Neslučovat nezávislá zadání do jednoho rozsáhlého PR.** Jedna logická změna / jedna sada testů / jedna kontrola regrese. Před implementací problematické etapy předložit krátký návrh a rizika.
4. Nikdy nezploštit ani nerasterizovat původní obsah PDF pouze kvůli zobrazení, výběru nebo OCR. Bez výslovného požadavku nepřevádět vektor na bitmapu, neměnit barvové prostory, embedded fonty ani výstupní tiskový záměr.
5. Oddělit souřadnice PDF, lokální souřadnice objektu a souřadnice obrazovky; při zoomu, rotaci, ořezu stránky a HiDPI zachovat přesnost hit-testu a pozice kurzoru.
6. Každá akce měnící obsah či pořadí dokumentu musí být podle konkrétních pravidel vratná; historie patří ke správné kartě dokumentu. Náhledy a neuložené změny nesmějí nechtěně přepsat původní PDF.
7. U nevratných následků (ztráta formátování, poškozený font, neplatnost digitálního podpisu, změna barev) předem konkrétně upozornit a nabídnout práci s kopií. Žádné tiché náhrady fontů ani tichá degradace PDF.
8. Zachovat stávající grafické zdroje `assets/anvil.png` a `assets/blacksmith.png` v repozitáři. Historický soubor `assets/blaksmith.ico` může být při sjednocení názvosloví **bez ztráty grafického obsahu** migrován na správně pojmenovaný `assets/blacksmith.ico`; v takovém případě aktualizovat všechny odkazy a sestavovací skripty. Původní renderované obrázky nikdy neodstraňovat jen kvůli sjednocení názvu.
9. U každého úkolu dodat: **před/po inventurní tabulku stávajících funkcí dotčeného nástroje a výsledky regresních zkoušek**, seznam změněných souborů, shrnutí, ověřená akceptační kritéria, provedené testy a známé limity. Ověřit `go test ./...` a relevantní Python testy, pokud jsou dostupná sestavovací prostředí; manuální Windows testy neoznačovat za dokončené, dokud se neprovedou.
10. Terminologii a české texty UI sjednotit. Nejasné chování označené „ROZHODNOUT“ nejprve navrhnout ve 2–3 variantách, nevymýšlet si jednu jako schválenou.

## 1. Interakce s plátnem a výběrem

### SEL-01 – Oddělení obdélníkového výběru a běžného klikání (UX, P0; SCHVÁLENO)

**Závazné pravidlo:** Stisk kolečka slouží **pouze k tažení výběrového obdélníku**; nevyžadovat jednoduchý, dvojitý ani modifikovaný klik kolečkem. Funkční základ je v `canvas_selection_windows.go` (`middleSelection`). Gesto funguje v každém nástroji PDF, kde lze vybrat obsah dokumentu, bez potřeby přepnutí režimu.

- **Prostřední tlačítko + tažení:** od místa stisku se vykresluje obdélníkový rámeček; uvolnění provede výběr textu či objektů v oblasti, podle aktivního nástroje nebo podle inteligentní detekce obsahu. Začátek může být i nad existujícím objektem. Zachovat průběžný náhled, toleranci pohybu a přesné souřadnice při zoomu, rotaci a HiDPI.
- **Samostatné stisknutí a uvolnění kolečka bez tažení:** žádný výběrový příkaz; neotevře editor, nepřepne aktivní objekt a nezmění výběr.
- **Levé kliknutí:** podle kontextu standardní interakce – u objektů inteligentní výběr pod kurzorem, u textu vstup do editace a umístění textového kurzoru, u miniatur aktivace/výběr stránky. Běžné modifikátory výběru jsou v SEL-02.
- **Levý dvojklik a další výběrová kliknutí:** inteligentní textový výběr dle TXT-02; nepoužívat dvojitý prostřední klik.
- **Levé tažení:** v aktivním textovém editoru standardní rozsah znaků; nad vybraným objektem přesun; u kreslicího nástroje kreslení. Nezaměňovat toto gesto za obdélníkový plošný výběr. **Ten je vždy vyhrazen stisknutému kolečku.**
- **Laso:** zůstává samostatným režimem s vrcholy zadávanými **levým** klikáním a uzavřením levým dvojklikem, poblíž počátku nebo Enter. Ani v režimu Laso nemění prostřední tažení svůj význam – stále vyznačuje obdélník. Otestovat také Esc a zrušení rozpracovaného polygonu.
- **Miniatury stránek:** běžný výběr levým klikem, vícenásobný výběr Ctrl/Shift + levý klik, přesun levým tažením. Kolečko není pro miniatury vyžadováno.
- **Technické pravidlo:** nerozbít existující scrollování otočením kolečka, zoom `Ctrl + otočení kolečka`, kontextové menu pravým tlačítkem ani běžnou práci s klávesnicí. Otestovat rozdíl mezi prostředním stiskem a tažením v různém DPI.

**Akceptace:** Prostředním tažením lze vyznačit obdélník při každém vhodném nástroji na volné stránce i nad objektem. Samostatný prostřední klik nic nevybere. Levým klikáním lze vybírat objekty, editovat text a vybírat miniatury; levý dvojklik vybere slovo. Laso se uzavírá levým dvojklikem. Uložení PDF se při samotném výběru nemění.

### SEL-02 – Klávesové modifikátory běžného výběru (UX, P1; ZÁKLAD SCHVÁLEN / DETAILY TESTOVAT)

| Gesto | Chování |
|---|---|
| `Prostřední tlačítko + tažení` | Obdélníkový plošný výběr. Bez tažení kolečko nic nevybere. |
| `Levý klik` na objekt | Vybrat nejvhodnější viditelný objekt pod kurzorem (OBJ-02). |
| `Levý dvojklik` do textu | Vybrat slovo; další dvojkliky ve stávajícím výběru: segment → věta napříč řádky → odstavec (TXT-02). |
| `Shift + levý klik` | Rozšířit výběr; v textu souvislý rozsah od kotvy, u miniatur souvislý rozsah stran. |
| `Ctrl + levý klik` | **V textu:** přímo vybrat celou větu pod kurzorem (TXT-02). **U objektů:** procházet překrývajícími se objekty podle OBJ-03; **u miniatur:** přidat/odebrat stránku. Funkce se určují podle aktivního kontextu a nesmějí se zaměňovat. |
| `Ctrl + Shift + levý klik` | Odebrat prvek z vícečetného výběru v nástroji Objekty; v textu přednostně respektovat standardní rozšíření výběru tam, kde je podporuje textový editor. |
| `Ctrl / Alt / Shift + levý dvojklik` | Pokročilý výběr textu podle TXT-02; uvést zkratky v nápovědě. |

**Důležité:** Na kolečko **nepřesouvat** funkce pro výběr slova, objektu, odstavce, výběr v hloubce ani rozšíření výběru. Modifikátory prostředního *tažení* lze doplnit později, ale nejsou podmínkou této změny. V `canvasSelectionChange()` zachovat přiměřeně kompatibilitu se stávající logikou a zabránit kolizi s levými textovými zkratkami. Rozlišit výběr textového kurzoru od výběru objektů a neblokovat nativní chování editační komponenty.

### SEL-03 – Společná chronologická historie výběru textu a úprav (FEATURE, P1; SCHVÁLENO Q3)

**Závazné chování:** `Ctrl+Z` vrací **poslední provedenou akci v aktuálním dokumentu**, ať jde o výběr textu, jeho rozšíření, zúžení či zrušení, nebo skutečnou úpravu obsahu PDF. `Ctrl+Y` standardně opakuje vrácenou akci. Stejné příkazy musejí obsluhovat i příslušné ikony **Zpět / Znovu** v horní liště. **Nezavádět jinou povinnou zkratku pro historii výběru.**

**Jak zaznamenávat kroky:**

1. Každá **dokončená uživatelská operace výběru textu** vytváří jeden vratný krok: první dvojklik (slovo), další dvojklik (segment), další (věta), další (odstavec), výběr `Ctrl/Alt/Shift + klik/dvojklik`, výběr oblasti stiskem kolečka a tažením, standardní výběr znaků myší či klávesnicí a explicitní zrušení výběru. Zaznamenat původní i výslednou množinu textových úseků, aktivní kotvu a polohu kurzoru, aby obnovení bylo přesné i u nesouvislých výběrů. Výběr, který se fakticky nezměnil, nevkládat do historie.
2. **Jedno tažení = jeden krok** až po uvolnění tlačítka; mezisnímky pohybu myši, hover, překreslování, čekání na OCR, focus či automatické vizuální aktualizace historii neplní. Dlouhé podržení šipek nebo postupné označování při psaní mohou být účelně seskupeny do jednoho logického kroku, ale jednotlivé stupně inteligentního dvojkliku zůstávají **vždy samostatné**.
3. **Výběr a editace v jedné časové ose:** `Ctrl+Z` vrací akce **v opačném pořadí jejich dokončení**, bez přednostního vyvolání historie RichEdit před historií výběru. Při aktivním textovém editoru musí centrální správce historie korektně integrovat kroky z RichEdit (včetně psaní a formátování); nesmí nejprve automaticky smazat poslední napsaný znak, pokud poslední akcí bylo rozšíření výběru.
4. **Příklad:** dvojklik na slovo → dvojklik na segment → dvojklik na větu → `Ctrl+Z` obnoví segment → `Ctrl+Z` obnoví slovo → `Ctrl+Y` znovu vybere segment. Jestliže následuje úprava textu, další `Ctrl+Z` postupně vrací nejprve pozdější akce (např. novou změnu výběru) a poté samotnou úpravu.
5. **Izolace dokumentů:** historie patří konkrétní kartě PDF a aktivnímu dokumentu v duálním zobrazení; změna karty nesmí míchat kroky dvou souborů. Staré souřadnice a objekty výběru po změně stránky či přepočtu zobrazení obnovovat podle stabilních identifikátorů textového modelu, nikoli podle původních obrazových pixelů. Pokud obnovení už není možné, nesmí dojít k výběru nesprávného textu nebo poškození PDF; situaci bezpečně ošetřit.
6. **Storno/Redo:** po novém uživatelském výběru či nové úpravě po provedeném Undo zrušit neplatnou větev Redo daného dokumentu. `Ctrl+Y` opakuje vrácené stupně výběru i změny obsahu v původním pořadí. Zachovat správný stav tlačítek a zkratek.
7. **Bez změny dat:** samotný výběr, jeho Undo a Redo nemění PDF, nezneplatňují podpisy, nevytvářejí obsahovou pracovní kopii a **nesmějí zapnout indikátor neuložených změn**. Naopak editace, přesuny stran a jiné obsahové operace indikátor mění obvyklým způsobem.
8. Prozkoumat existující `history_windows.go`, `document_commands_windows.go`, `inline_text_windows.go`, `pages_windows.go` a způsob evidence výběrů v `text_windows.go` / `canvas_selection_windows.go`. Navrhnout centrální chronologický zápis akcí či kompatibilní nadstavbu nad stávajícími zásobníky; **nekombinovat dvě nekoordinační historie tak, že se poruší časové pořadí**. Zachovat již funkční Undo/Redo stránky, OCR a editace objektů.

**Akceptace:** Na textu se dá pomocí `Ctrl+Z` vrátit z odstavce na větu, ze věty na segment, ze segmentu na slovo a případně na původní stav. `Ctrl+Y` vše znovu provede. Pokud mezi výběry vznikne úprava PDF, Ctrl+Z vrací skutečně poslední krok bez přeskočení výběru ani změny obsahu. Historie dvou dokumentů je oddělená; pouhé vybírání neoznačí PDF jako změněné. Otestovat také editor s textovým kurzorem a nezávisle funkci Zpět/Znovu v horní liště.

## 2. Text – rozpoznání, výběr a editace

### TXT-01 – Oddělit výběr slova, segmentu, řádku, odstavce a sloupce (BUG/FEATURE, P0)

**Zrušit pravidlo:** Levý dvojklik nesmí automaticky označovat celý odstavec. Standardní **levý dvojklik vybere slovo**; případné další rozšíření výběru je popsáno v TXT-02. Dvojklik kolečkem se nepoužívá.

**Nový model textu:**
- **Znak/glyf:** nejmenší jednotka s polohou, textem a vlastním formátovacím během (*run*).
- **Slovo:** souvislý jazykový celek včetně správného zacházení s interpunkcí, českou diakritikou a Unicode.
- **Segment:** významově a prostorově souvislý úsek na jednom řádku; hranice tvoří tabulátory, výrazné geometrické mezery nebo více mezer, nikoli pouhá změna fontu či barvy.
- **Řádek:** jeden vizuální řádek, případně oddělené buňky, pokud jde o tabulkový řádek. Přímý výběr řádku zůstává `Ctrl + levý dvojklik`; není stupněm běžného postupného dvojkliku.
- **Věta:** jazyková jednotka ohraničená interpunkcí a kontextem; může pokračovat přes **několik vizuálních řádků** stejného odstavce/sloupce. Nesmí se rozdělit jen kvůli zlomu řádku, změně fontu ani rozdělení textu na PDF objekty.
- **Odstavec:** skupina sousedních řádků se společným zarovnáním, přibližným řádkováním a prostorovou návazností.
- **Sloupec:** navazující vertikální řada řádků či segmentů se srovnatelným horizontálním intervalem; nesmí přeskočit do sousedního sloupce.
- **Návazné odstavce:** více sousedících odstavců ve stejném sloupcovém pásmu; používá se jen při výslovném rozšíření výběru.

**Důležitý princip:** Hranice PDF textových objektů (*runs*, proudy, fonty) **nejsou** automaticky hranicemi vět, řádků či odstavců. Výběr se nesmí zastavit na neočekávané změně formátování. Textový model musí používat pozice glyphů, mezery, zarovnání a orientaci.

### TXT-02 – Klikací výběry textu levým tlačítkem (FEATURE, P0; SCHVÁLENO Q2)

**Závazné ovládání:** Všechny pokliky pro inteligentní výběr textu používají **levé tlačítko myši**. Stisknuté kolečko slouží pouze k tažení obdélníkového rámečku. Výběr musí procházet hranicemi PDF textových objektů a jednotlivých formátovacích běhů; rozdílný font, barva, řez či velikost písma samy o sobě nerozdělují textový segment, řádek ani odstavec.

| Gesto | Výsledek |
|---|---|
| `Levý klik` do editovatelného textu | Aktivovat editor, umístit textový kurzor na místo kliknutí. |
| `Levé tažení` v aktivním textovém rámci | Standardní znakový rozsah výběru jako v textovém editoru, nikoli obdélník. |
| **1. levý dvojklik** | Vybrat **slovo** pod kurzorem. |
| **2. levý dvojklik na již vybraném slově** | Rozšířit výběr na **textový segment** daného řádku (celek mezi tabulátory, vícenásobnými mezerami nebo výraznými optickými mezerami). |
| **3. levý dvojklik v právě vybraném segmentu** | Vybrat **celou větu obsahující původně zvolené slovo**, včetně pokračování na následujících/předchozích vizuálních řádcích, pokud patří do stejného odstavce a pořadí čtení. |
| **4. levý dvojklik v právě vybrané větě** | Vybrat **celý odstavec**, do kterého věta patří, bez přechodu do sousedního sloupce nebo nesouvisející tabulkové buňky. |
| `Ctrl + levý klik` do textu | **Přímo vybrat větu** pod kurzorem (návrh zkratky inspirovaný textovými editory). V režimu Objekty zůstává `Ctrl + klik` určen pro hloubkové procházení, v náhledech pro multivýběr. |
| `Ctrl + levý dvojklik` | **Přímo vybrat celý řádek**, bez nutnosti projít slovem a segmentem. |
| `Alt + levý dvojklik` | V příslušném textovém segmentu vybrat text se **stejným formátováním** jako slovo nebo znaky pod kurzorem. Výsledkem mohou být nesouvislé úseky. |
| `Shift + levý dvojklik` | Vybrat **odstavec** v aktuálním sloupci, podle vizuálního uspořádání a odstavcových hranic. |
| Další `Shift + levý dvojklik` v již vybraném odstavci | Rozšířit výběr na **celý odpovídající textový sloupec na dané stránce**, nikoliv do sousedních sloupců. |
| `Alt + Shift + levý dvojklik` | V **aktuálním sloupci** vybrat všechny úseky, které mají **shodné formátování** jako text pod kurzorem, i pokud nejsou souvislé. |

**Posloupnost bez modifikátorů je `slovo → segment → věta → odstavec`.** Věta může přesáhnout více vizuálních řádků; naopak **celý řádek** se vybírá samostatně `Ctrl + levý dvojklik`. Další dvojklik po dosažení odstavce ponechá odstavec vybraný; na sloupec se přechází výlučně přes příkaz `Shift + dvojklik`.

**Hraniční případ hierarchie:** Segment mezi výraznými mezerami může obsahovat dvě nebo více vět. Ve třetím kroku se tedy výběr změní na **větu původního kotevního slova** a může být užší než druhý stupeň; nesmí se svévolně přibrat sousední věta jen proto, aby výběr rostl. O této výjimce informovat v nápovědě. Hlavní je přesnost významového výběru.

**Rozlišení segmentu a řádku:** Segment má hranice na *výrazných* vodorovných mezerách, které se určují podle fyzického rozmístění glyphů, běžných mezer a tabulátorů. Obyčejná mezera mezi slovy sama o sobě netvoří hranici segmentu. Celý vizuální řádek zahrnuje i sousední segmenty, leží-li skutečně na stejné řádkové ose. U tabulek nesmí vyvolat výběr nesouvisejícího řádku nad nebo pod buňkou; je třeba respektovat hranice tabulkového obsahu.

**Požadavky na obsluhu opakovaných dvojkliků:**

1. Další dvojklik postoupí na vyšší stupeň, pokud je uvnitř aktuálně vybraného rozsahu, výběr je stále aktivní a jde o stejný kotvený textový celek na téže stránce. **Věta a odstavec mohou být vícerádkové** a nemusí zůstat uvnitř výchozího vizuálního řádku. Zaznamenat původní kotevní slovo/glyf pro všechny úrovně; nevyžadovat absolutně totožnou souřadnici.
2. Samostatné sekvence rozšíření podle modifikátorů jsou **nezávislé**: `Shift + dvojklik` přejde z odstavce na sloupec, ale nenavazuje na bezmodifikátorový stav `slovo → segment → věta → odstavec`. `Ctrl + klik` přímo vybere větu, `Ctrl + dvojklik` přímo vybere vizuální řádek; `Alt` a `Alt+Shift` mají vlastní formátovací výběry.
3. Po kliknutí mimo výběr, změně nástroje, textového bloku, stránky, po Esc nebo po explicitním zrušení výběru začíná nová sekvence. Nativní dvojklik a vícenásobné klikání Windows zpracovat tak, aby se nezaměnilo za nechtěný trojklik nebo čtyřklik.
4. **Nerozšiřovat výběr při samotné změně formátování:** střídání tučného a normálního textu je pro prostorový výběr transparentní; naopak režimy `Alt` pracují právě se shodou formátování. Shodu pro formátovací výběr definovat přes porovnávané typografické vlastnosti, ne pouze název fontu.
5. Aktivní editor při výběru zůstává otevřený, obsah se nepoškodí ani automaticky nepřepisuje a jednotlivé úseky jsou viditelně zvýrazněné. Kolečko nikdy nevyžaduje samostatný ani dvojitý klik.
6. **Rozpoznání hranic vět:** respektovat české `.` / `!` / `?` / `…` a související závěrečné uvozovky/závorky. Rozlišovat konce vět od zkratek (např., tj., atd., Ing., č.), desetinných čísel (např. 3,14 nebo 1.25), iniciál, číslování, URL a e-mailů; použít jazykový kontext, nikoli prosté dělení na každé tečce. U nejednoznačného textu neprezentovat odhad jako zaručený.
7. **Rekonstrukce pořadí čtení z PDF:** spojovat textové běhy přes zalomení řádku podle pozice glyfů, návaznosti baseline, odsazení a rozpoznaného sloupce; tvrdé hranice odstavců a buněk tabulky respektovat. Odstranění čistě vizuálního zalomení je pouze operace textového modelu **pro výběr**, nikoli zásah do obsahu originálního PDF. U dělených slov na konci řádku rozlišit spojovník od koncového dělení.
8. **PDF bez použitelné textové vrstvy:** automaticky netvrdit, že rozpoznávání vět funguje; nabídnout cílené OCR a výběr podle rozpoznaného textu, bez trvalé změny PDF bez souhlasu. Složitější rotované/svislé psaní a více sloupců mají vlastní testy.
9. Ošetřit rozdíl mezi `WM_LBUTTONDBLCLK` a posloupností jednotlivých kliků tak, aby `Ctrl + klik` nevyvolával náhodné cyklování objektů a `Ctrl + dvojklik` nevybral nejprve větu a pak chybně zrušil výběr řádku.

**Akceptační scénáře:** Na řádku `Číslo výrobku     Popis výrobku     12345` první dva dvojkliky na „Popis“ vyberou `Popis`, následně `Popis výrobku`; třetí se pokusí vybrat **skutečnou větu** daného kotevního slova, čtvrtý **odstavec**. `Ctrl + dvojklik` přímo vybere celý vizuální řádek, včetně okolních segmentů. Na dvou/třech řádcích odstavce „Tento dokument obsahuje technické údaje, / které se používají pro tisk a další / zpracování.“ musí třetí dvojklik nebo `Ctrl + klik` vybrat **všechna tři zalomená pokračování**, ale nesmí zasáhnout další odstavec ani vedlejší sloupec. U zkratek `Ing.` a `např.` výběr věty nesmí předčasně skončit. U segmentu se dvěma větami vybere třetí krok větu kotevního slova; čtvrtý celý odstavec. `Shift + dvojklik` vybere odstavec, další celý sloupec, nikoliv vedlejší. `Alt + dvojklik` vybírá stejný styl v segmentu, `Alt + Shift + dvojklik` v aktuálním sloupci. Testovat střídání stylů, rotaci stránky, malé přiblížení, různá DPI a přítomnost skutečné nebo OCR textové vrstvy.

### TXT-03 – Inteligentní tabulky a sloupce (FEATURE, P0)

1. Nejprve rozpoznat orientaci textu a vizuální řádky podle baseline, směrů písmových souřadnic a geometrie glyfů; zohlednit natočení stránky i jednotlivých textových objektů.
2. Identifikovat výrazné mezery relativně k běžným mezerám a velikosti písma v daném řádku; nepoužívat jednu absolutní konstantu pro všechny PDF.
3. Odhadnout pásma sloupců z opakovaného zarovnání segmentů a projekce jejich obdélníků do horizontální osy.
4. Pokud v některém řádku místo očekávané mezery zasahuje text přes hranici pásma, vyhodnotit ukončení či začátek předchozího bloku. Nepropojit přes takový řádek dva logicky oddělené odstavce.
5. Odstavce rozlišovat podle řádkování, odsazení, odstavcových mezer, shody hran a hranic sloupce. Změna fontu v jediném slově nesmí odstavec rozdělit.
6. Implementovat vizuální indikaci rozsahu před potvrzením složitějšího vícečetného výběru; například zvýraznit kandidátní části.
7. U skenů bez textové vrstvy nabídnout OCR vybrané oblasti; neoznačovat optickou detekci za jistou, pokud je výsledek nejednoznačný.

**Akceptace:** Odstavec obsahující střídání bold/regular i změnu velikosti písma zůstane jedním odstavcem; výběr jedné buňky tabulky nezasáhne sousední sloupec; český vícesloupcový text respektuje pořadí čtení.

### TXT-04 – Přepnutí do editace a caret (UX/FEATURE, P0; ROZHODNOUT)

Při **jednoduchém levém kliknutí** na editovatelný text v nástroji Text ihned zobrazit caret v nejbližší pozici mezi znaky; **levý dvojklik a jeho kombinace** rozšiřují textový výběr podle TXT-02, zatímco **prostřední tažení** vybírá oblast obdélníkem. Při výběru ponechat editovatelný text aktivní a umožnit aplikaci stylů; pro psaní se používá jednoznačný caret. Zachovat původní formátovací běhy. Pokud je označeno **více nesouvislých oblastí** (laso, několik podobně formátovaných úseků), neumisťovat jeden falešný caret do všech oblastí; výběr může měnit společné vlastnosti, ale přímé psaní je možné jen v jednoznačném aktivním textovém rámci.

U textu tvořeného pouze křivkami či bitmapou nabídnout varianty „Rozpoznat OCR“, „Přidat novou textovou vrstvu“ a „Zrušit“; bez souhlasu nepřevádět křivky na editovatelný text. Pokud konkrétní text není bezpečně editovatelný (subset fontu, chybějící glyfy, neobvyklá transformace), ukázat konkrétní důvod.

### TXT-05 – Zachování formátování při editaci (TECH/FEATURE, P0)

Při otevření editace a při potvrzení úpravy zachovat ve všech **nezměněných úsecích** původní PDF data, pokud to architektura dovolí. U změněných znaků respektovat: rodinu a řez fontu, dostupné embedded fonty a jejich omezení, velikost, barvu a prostor, horizontální škálování, vertikální posun, rotaci/sklon a obecnou textovou transformaci, tracking, word spacing, řádkování, odsazení, zarovnání, pořadí a hranice znakových stylů, případně režim vyplňování/obrysu. Nezaměňovat pouze podobně vypadající font za původní bez varování.

Uživatel chce pokročilé formátování jako v textovém procesoru: změna fontu/řezu/velikosti, barvy, zarovnání (vlevo, vpravo, střed, do bloku), řádkování, mezery před/za odstavcem, odsazení, tabulátory, proklad a šířka znaků, případně horní/dolní index, změna velikosti písmen. V první etapě pouze funkce prokazatelně realizovatelné bez nevratného poškození původního PDF; další ve druhé etapě.

**Výslovný limit:** Dokonalé zachování typografie není vždy možné, když PDF neobsahuje plnohodnotný font nebo potřebné znaky. Codex musí definovat pravidla výstrah a nabízených náhrad; **nikdy nesmí slíbit bezeztrátovou editaci každého PDF**.

### TXT-06 – Nový text (FEATURE, P1)

Dvě samostatné volby: **Bodový text** (kliknutím vzniká krátký textový objekt; rozšiřuje se přirozeně podle obsahu) a **Odstavcový text** (tažením se vymezí textový rámec s šířkou, zalamováním a vlastním formátováním). Při vložení do existujícího rámce volit zachování zdrojového formátování nebo převzetí cílového stylu.

### TXT-07 – Kontextové nabídky (UX, P1; ROZHODNOUT)

Zahrnout „Vybrat slovo“, „Vybrat segment“, „Vybrat větu“, „Vybrat řádek“, „Vybrat odstavec“, „Vybrat sloupec“, „Vybrat navazující odstavce“, „Vybrat shodné formátování v segmentu/sloupci“. Při dlouhém menu použít přehledné podmenu **Rozšířený výběr**, nikoliv hned nové `Ctrl + pravé tlačítko`. Rozšířený výběr nabídnout také ikonami v panelu Text, pokud se tím UI nezahltí. Tooltip s názvem a zkratkou; v panelu delší kontextová nápověda. Zajistit přístup pomocí klávesnice.

## 3. Objekty, laso a kreslení

### OBJ-01 – Opravit laso a zavést dva režimy (BUG/FEATURE, P0; SCHVÁLENO Q16)

**Cíl:** Funkční polygonální výběr, který podle uživatelsky zvoleného režimu vybere objekty protnuté hranicí/oblastí lasa, nebo pouze celé objekty uvnitř oblasti. Opravit současné vyhodnocování podle geometrického středu objektu v `object_selection_windows.go` (`objectsAt`), které přehlíží dlouhé čáry nebo velké objekty dotčené pouze okrajem.

**Schválené režimy – dvě malé ikony v panelu Laso (s tooltipem a společnou nápovědou dle UI-07):**

1. **A – Protnutí objektů (VÝCHOZÍ, aktivní při prvním spuštění).** Vybrat objekt, pokud má *jakoukoli skutečně viditelnou část svého tahu, výplně nebo bitmapy* uvnitř polygonu lasa, včetně objektu, který polygon protíná, a objektu, který viditelnou výplní polygon zcela pokrývá. Není nutné, aby střed objektu nebo jeho obdélníkové hranice ležely uvnitř. U nevyplněného objektu neposuzovat prázdný vnitřek jako viditelnou kresbu. Průhledná oblast obrázku bez viditelného obsahu nemá sama vytvořit zásah. Respektovat transformace, aktuální ořez a geometrii PDF.
2. **B – Pouze celé objekty uvnitř lasa (VOLITELNÝ).** Vybrat objekt jen tehdy, když **celá jeho skutečně viditelná malba** leží uvnitř polygonu. Nestačí samotná přítomnost středu ani přibližná podobnost bounding boxu. Nevybírat čáru pokračující ven z polygonu; vybrat plně obklopený malý objekt. U složitých PDF konstrukcí (skupiny/Form XObject, ořezové masky, průhlednosti) posuzovat vykreslenou viditelnou část; pokud přesnost nejde bezpečně zaručit, nepředstírat ji a označit omezení.

**Interakce:**

- V nástroji Laso **levý klik** přidá vrchol polygonu; při pohybu myši se vykresluje průběžná spojnice k kurzoru. **Levý dvojklik** poslední bod vloží nejvýše jednou a polygon uzavře. Stejně lze uzavřít **levým klikem v přiměřené obrazovkové toleranci od počátečního bodu** nebo klávesou **`Enter`** (je-li zadán alespoň trojúhelník). **`Esc`** zruší rozpracovaný obrys bez změny předchozího výběru.
- Uzavřením proběhne skutečné vyhodnocení geometrie, vybrané objekty se zřetelně zvýrazní. **Pouhé nakreslení obrysu není funkční výběr.** Prázdný či degenerovaný polygon nesmí změnit dokument, vyvolat falešný výběr ani havarovat.
- **Stisk kolečka + tažení** nadále provede globální obdélníkový výběr podle SEL-01, i když je aktivní Laso. Režimy A/B jsou preference **lasa**; nezasahují do geometrických pravidel obdélníkového výběru.
- Klávesové modifikátory rozšiřování/odebírání výběru používat v souladu s již existujícím společným pravidlem SEL-02. Samotný výběr objektů nesmí měnit obsah PDF ani aktivovat příznak neuloženého dokumentu.
- Volba A/B má zřetelný aktivní stav a dostupné názvy pro klávesnici a tooltip: **„Laso – protnutí“** a **„Laso – celé objekty“**. Přepnutí režimu nesmí zničit právě vybrané objekty. Při rozpracovaném polygonu režim aplikovat při uzavření a jednoznačně jej zobrazovat; přepnutí režimu během kreslení nesmí poškodit stav.

**Kvalita a architektura:** Prověřit `object_selection_windows.go` (`objectRegionPointer`, `finishObjectRegion`, `objectsAt`) a kontury/extrakci objektů v `object_edit.py`. Pro výběr využít skutečné křivky, tahy a výplně, případně bezpečný detailní test průniku viditelné kresby; **nepřevádět kvůli výběru finální PDF na bitmapu**. Případná pomocná bitmapová maska pro přesný hit-test smí existovat pouze dočasně v paměti. Zakázané je zjednodušit režim A na „střed uvnitř“ a režim B na „střed nebo bounding box uvnitř“ bez testování viditelné geometrie.

**Akceptační scénáře:**

| Scénář | Režim A: Protnutí | Režim B: Celý objekt |
|---|---|---|
| Dlouhá tenká čára protíná oblast, oba konce mimo | **Vybrat** | Nevybrat |
| Malá plně obklopená vyplněná elipsa | **Vybrat** | **Vybrat** |
| Částečně obklopený obdélník, viditelný okraj přesahuje | **Vybrat** | Nevybrat |
| Velký nevyplněný obrys leží mimo laso, pouze jeho prázdný vnitřek ho obklopuje | Nevybrat | Nevybrat |
| Vektorová výplň zcela pokryje prostor uvnitř lasa | **Vybrat** | Nevybrat |
| Transparentní část PNG v polygonu, viditelná malba mimo | Nevybrat | Nevybrat |
| Skutečně viditelná malba bitmapy zcela uvnitř lasa | **Vybrat** | **Vybrat** |

Dále otestovat polygon konkávní i konvexní, objekty s transformací/rotací, rozdílné zoomy a HiDPI, ořezové masky a překrytí, více vybraných objektů, přepnutí režimů, dvojklik/Enter/uzavření u počátečního bodu a Esc. Žádné změny původních objektů ani exportního PDF pouhým výběrem.


### OBJ-02 – Inteligentní zásah objektu kurzorem (BUG/FEATURE, P0; SCHVÁLENO Q15)

**Cíl:** běžný **levý klik** vybírá ten objekt, jehož skutečně viditelná část odpovídá ukazateli, nikoli automaticky největší obdélník či objekt na pozadí.

1. **Vektorová čára:** vyhodnotit vzdálenost k reálně vykreslené kontuře včetně křivek a tloušťky tahu; neomezovat se na hranice objektu. Neviditelná nebo ořezaná část se pro běžný klik nepovažuje za viditelný zásah.
2. **Vyplněná cesta:** přednost má skutečný zásah do viditelné výplně, pokud poblíž není přesněji zasažená kontura jiného viditelného objektu. Zohlednit pravidlo vyplňování, průhlednost, clipping a masky.
3. **Bitmapa:** respektovat transformaci, ořez a průhledné pixely/masky. Průhledný obdélník obrázku nesmí blokovat kliknutí na skutečnou kresbu pod ním.
4. **Zásahy a pořadí:** sestavit kandidáty ze skutečné geometrie (včetně PDF `Form XObject`), seřadit podle kvality viditelného zásahu; **při srovnatelném zásahu** dát přednost objektu vykreslenému výše. Obyčejný klik nemá vybírat objekt zcela skrytý neprůhledným objektem nad ním.
5. **Tolerance:** výchozí přibližně **5 obrazových pixelů kolem kurzoru**, definovaných v obrazovkových souřadnicích s korektním přepočtem při zoomu a HiDPI, ne jako pevná vzdálenost v PDF bodech. Hit-test nemá upřednostnit objekt vzdálený desítky pixelů jen proto, že kurzor leží v jeho bounding boxu.
6. **Odezva:** výběr jasně zvýraznit, ale hover sám nemění výběr; při nejednoznačnosti lze budoucím rozšířením zobrazit předvýběr. Výběr nesmí měnit data PDF, ani vyžadovat OCR nebo rasterizaci finálního obsahu.

**Akceptace:** na stránce s velkým průhledným Form XObject, tenkou čárou, malou elipsou, výplní a průhledným PNG klik na čáru do 5 px vybere čáru; klik do průhledného pixelu obrázku vybere objekt pod ním; klik mimo viditelný objekt nevybere jen samotný bounding box. Chování je stabilní při zoomu a rotaci.

### OBJ-03 – Procházení překryvů v hloubce (FEATURE, P1; SCHVÁLENO Q15)

Opakované `Ctrl + levý klik` na **stejném místě** cyklicky prochází kandidáty zasažené skutečnou geometrií v pořadí od nejvrchnějšího po nejspodnější; po posledním se vrací k prvnímu. Záměrně tak lze zvolit i objekt pod neprůhledným překryvem, jestliže jeho skutečná geometrie v daném bodě leží. Na rozdíl od běžného kliknutí se pro tuto cílenou funkci mohou kandidáti nacházet pod překrytím. Změna místa, aktivního nástroje, stránky nebo transformace plátna cyklus resetuje. Výběr se viditelně zvýrazní, ale nesmí přeskočit na vzdálený objekt, jehož obdélníkový rámec pouze obsahuje ukazatel.

**Akceptace:** při třech skutečně překrývajících se tvarech vznikne předvídatelný cyklus 1 → 2 → 3 → 1; žádný prázdný stisk nemění dokument. Běžný `Ctrl + levý klik` v nástroji Text či v miniaturách stále znamená jejich vlastní, dříve schválenou akci – cyklování platí jen pro objekty.

**Poznámka ke konvencím:** V Inkscape existuje obdobné procházení přes `Ctrl+Alt+klik` nebo `Alt+kolečko`. V PDF Blacksmith se držíme původně zamýšleného `Ctrl + levý klik`. Je však nutné odlišit cyklování překryvů od běžného Ctrl-přidání do vícečetného výběru a otestovat zvolenou prioritu. Stisk kolečka samostatně nepoužívat.

### OBJ-04 – Vkládání primitivních tvarů (FEATURE, P1; SCHVÁLENO Q4)

Základní tvary **Obdélník** a **Elipsa** tvořit **levým tlačítkem + tažením**; výsledek je editovatelný PDF objekt se samostatnou výplní, obrysem, barvou, tloušťkou čáry, průhledností a přesnými rozměry. Platí jednotná konvence PDF Blacksmith:

| Modifikátor při levém tažení | Obdélník | Elipsa |
|---|---|---|
| Bez modifikátoru | Libovolný obdélník z výchozího rohu | Libovolná elipsa z výchozího rohu |
| `Shift` | Čtverec (poměr stran 1 : 1) | Kruh (poměr stran 1 : 1) |
| `Ctrl` | Obdélník rozšiřovaný symetricky od výchozího středu | Elipsa rozšiřovaná symetricky od středu |
| `Ctrl+Shift` | Čtverec rozšiřovaný symetricky od středu | Kruh rozšiřovaný symetricky od středu |

**Technická interpretace:** V režimu kreslení je výchozím bodem pozice prvního stisku levého tlačítka. Bez `Ctrl` tento bod určuje roh ohraničujícího obdélníku; s `Ctrl` se stává středem. `Shift` vynutí stejnou šířku a výšku. Podle aktuální polohy kurzoru se rozměry mění i směrem doleva/nahoru; znaménko směru nevede k záporným výsledným rozměrům. Průběžně zobrazovat náhled a reagovat na přidání/odebrání modifikátorů ještě před puštěním tlačítka. `Esc` zruší rozpracovaný tvar bez vytvoření objektu. Krátký klik bez skutečného tažení nevytvoří náhodný objekt nulové velikosti. Kreslení je změnou dokumentu s podporou `Ctrl+Z` a `Ctrl+Y`.

**Konflikty ovládání:** Prostřední tlačítko + tažení zůstává v tomto režimu **obdélníkovým výběrem**, nikoliv kreslením. Samostatné `Ctrl`/`Shift` se uplatňují pouze během levého tažení s aktivním nástrojem vytváření tvaru; mimo tento nástroj nenarušovat stávající textové, výběrové a klávesové akce ani `Ctrl+kolečko` pro zoom. V tooltipu obou nástrojů uvést všechny čtyři varianty. Nepřejímat opačné zkratky z Inkscape či Illustratoru.

**Akceptační kritéria:** (1) Obdélník a elipsa při běžném tažení, (2) přesný čtverec a kruh se `Shift`, (3) symetrické kreslení od středu s `Ctrl`, (4) kombinace obou s `Ctrl+Shift`, (5) přesnost při zoomu, rotaci stránky a HiDPI, (6) náhled, `Esc`, Undo/Redo a možnost následné editace bez rasterizace.

### OBJ-05 – Rozložení ovládání objektů (BUG/UX, P1)

V panelu Objekty nahradit nadbytečné textové tlačítko ikonou, pokud je význam ikony jednoznačný. Textový štítek lze zobrazit v tooltipu se zkratkou; pod řadou ikon zobrazovat detailní popis právě najeté či aktivní funkce. Uspořádat ovládání tak, aby tlačítka nekolidovala s popisky ani při změně DPI a zúžení panelu. Nepoužívat ikonky bez textového ekvivalentu pro klávesnici a asistivní technologie.

## 4. Přetahování textu a objektů mezi aplikacemi

### DND-00 – Společný mechanismus: vizuální tažení a přenos dat odděleně (FEATURE/TECH, P1–P2; CÍL SCHVÁLEN, IMPLEMENTACI OVĚŘIT)

**Cíl uživatelské zkušenosti:** Uživatel vybere obsah v PDF Blacksmith, tažením jej přesune do jiného otevřeného PDF, cizí aplikace, složky Průzkumníka nebo na plochu. **Pokud je cíl technicky rozpoznatelný a obsluhovatelný, po uvolnění myši zobrazit malou nabídku typu přenosu a teprve po potvrzení vytvořit/vložit výstup.** Uživatel nemá být nucen vybírat formát před zahájením tažení. Toto je cílové chování, nikoliv tvrzení, že je možné do každé cizí aplikace vložit data libovolným způsobem.

**Návrh vnitřní architektury:**

1. Při zahájení tažení vytvořit **přenosový balíček**: identifikátor původního PDF a stránky, výběr, typ (`text`, `image`, `vector`, `page`), nativní data/strukturu, přístupné výstupní formáty, metadata, barevný prostor/ICC a volitelný náhled. Při pouhém výběru nepřepisovat PDF ani nerozkládat vektory do rastru. Originál zůstává nedotčen.
2. Vedle kurzoru zobrazovat **průhledný plovoucí náhled** (*ghost image*) prostřednictvím vhodné techniky Win32 (`WS_EX_LAYERED`, drag-image helper); to **není skutečný editor ani mechanismus vložení**. Překryv nesmí blokovat vstup cílového okna ani překrývat systémová potvrzení. Doporučit ověřit `IDragSourceHelper` s OLE jako první možnost.
3. Během tažení sledovat typ cíle a jeho podporované formáty. **Kde je to technicky bezpečné a možné**, po puštění nabídnout pouze smysluplné akce: **vložit se zachováním formátování / jako prostý text / jako původní obrázek / jako PDF nebo vektorový objekt / zrušit**. Přesný výběr závisí na obsahu a cíli. Je-li cíl rozpoznaný, ale nekompatibilní, uvést proč; nikdy nepředstírat úspěšný přenos.
4. **Dokončení přenosu po potvrzení se liší podle cíle:** (a) interní PDF Blacksmith – vlastní aplikační vložení, (b) běžná složka/plocha Windows – vytvoření zvoleného souboru do ověřené cesty, (c) cizí editor – primárně standardní OLE `IDataObject` / `DoDragDrop` s kompatibilními formáty. Požadavek na volbu *až po uvolnění* pro cizí aplikace vyžaduje **samostatný prototyp**: během běžného OLE dropu totiž obsah přijímá a vyžaduje cílová aplikace. Překryv sám neumí dodatečně vložit do cizího editovacího pole.
5. **Přednostně nabídnout cíli souborový přenos jako při tažení z Průzkumníka** (`CF_HDROP` pro skutečný dočasný soubor, případně `CFSTR_FILEDESCRIPTOR` + `CFSTR_FILECONTENTS` tam, kde jsou virtuální soubory podporovány). Průhledný náhled nesmí blokovat skutečnou OLE interakci. Pro kompatibilní aplikace lze volitelně vyzkoušet **řízené předání přes schránku a vložení do aktivního cíle** (uživatelem potvrzená akce) nebo jinou cílově specifickou integraci. Předtím ověřit zaměřený vstupní prvek, možnosti cílové aplikace a oprávnění. Nespoléhat na `SendInput` jako univerzální náhradu OLE: zabezpečení UIPI, elevované procesy, aplikace ve webu i rozdílné editory mohou vložení blokovat. Neposílat `Ctrl+V` naslepo, nepřepsat nechtěně obsah a nepoškodit cizí schránku; pokud vložení nelze bezpečně ověřit, nabídnout **Kopírovat do schránky** / **Uložit soubor** a jasně popsat výsledek.
6. Dočasná data lze držet v paměti; souborová mezivrstva se zakládá na **`%TEMP%/PDFBlacksmith/<session-id>/`**, nikoli automaticky v uživatelských Dokumentech. Vytvořit jedinečné názvy, nepovolovat kolize ani přístup k jiným souborům, předcházet úniku původního PDF. Dočasné soubory smazat **až po uvolnění všech příjemců/přenosů**, při ukončení nebo při příštím spuštění uklidit osiřelá data; počítat s pádem aplikace. Uložené cílové soubory se nikdy nesmějí smazat při ukončení aplikace.
7. **Výstupní formát se odvozuje od povahy vybraného obsahu, nikoliv pouze od jeho vzhledu:** prostý text → UTF-8 `.txt`; formátovaný text → RTF/HTML, případně fragment PDF; bitmapa → původní kódovaný formát, pokud jej lze bezpečně a správně získat, jinak PNG jako bezztrátová možnost; JPEG jen pro vhodný fotografický obsah a s jasným upozorněním na ztrátu kvality; vektor → nativní vektorový obsah/PDF, ne automaticky JPEG/PNG; stránka/stránky → PDF. Zachovat dostupnou průhlednost, fonty, ICC a transformace; nepřislíbit věrnost vlastností, které cílový formát neumí. Rozšířený export EPS/TIFF/SVG zůstává samostatným budoucím úkolem, nikoli povinnou součástí prvního přenosu.
8. **Jednotný fallback:** pokud se vybraný objekt nepodaří vložit do cizí aplikace nebo kompatibilitu nelze potvrdit, zobrazit např. „Vybraný obsah se do aplikace [název] nepodařilo vložit. Můžete jej uložit jako soubor a následně otevřít nebo přetáhnout ručně.“ Ovládací prvky: **Uložit jako…** (standardní dialog s volbou libovolné cílové složky, předvyplněným názvem a podporovaným typem), volitelně **Kopírovat do schránky**, **Zrušit**. Neodstraňovat zdrojový výběr a nezapisovat do cizího okna bez ověřené podpory. Úspěch označit pouze po potvrzeném dokončení operace; odmítnutí formátu nesmí znamenat ztrátu původního obsahu.
9. Napříč všemi cestami implementovat zrušení `Esc`, chování při odmítnutí cílem, správné čištění, indikátor skutečného dokončení a testy změny DPI, dvou monitorů, UAC a kompatibility. Při uložení pro pozdější vložení zachovat soubor **trvale**; dočasné soubory se nesmějí používat jako trvalý zdroj pro propojené (linked) objekty a nesmějí se smazat, dokud je přijímající editor může ještě potřebovat.

**Povinná ověřovací etapa před produkční implementací:** Prototyp především pro: druhý PDF v aplikaci, běžnou složku v Průzkumníku, **CorelDRAW, Adobe Photoshop, Affinity Designer/Photo**, dále Word/LibreOffice Writer a jednoduchý textový editor. Zkoušet na konkrétních dostupných verzích Windows a editorů, pro každý typ (`text`, `image`, `vector`, `page`) evidovat podporovaný formát a přímé vložení vs. náhradní uložení. Ověřit, **zda lze nabídku otevřít až po uvolnění myši a teprve pak provést přenos**. Výsledky explicitně zdokumentovat pro každou kategorii cíle; nerozšiřovat nepodložené předpoklady na všechny Windows aplikace. Pokud cílové chování nejde pro konkrétní cizí aplikaci zajistit, použít kompatibilní standardní OLE přetažení nebo nabídnout bezpečné manuální dokončení, nikoli iluzorní úspěch.

### DND-01 – Přenos mezi otevřenými PDF (FEATURE, P1; NABÍDKA PO UVOLNĚNÍ)

Při přetažení vybraného editovatelného textu do jiného dokumentu PDF Blacksmith po uvolnění myši nabídnout **Vložit s původním formátováním** / **Vložit jako prostý text** / **Zrušit**. Prostý text převezme styl cílového textového rámce, nebo přednastavený styl pro novou pozici. Zachovat správnou polohu a nepoškodit zdroj. Pro objekty nabídnout přenos se zachováním struktury a barev; případné volby převodu podle barevného prostoru cílového PDF řešit zvlášť. V této interní variantě aplikace ovládá zdroj i cíl, a proto je nabídka po uvolnění plně pod její kontrolou. Respektovat Undo/Redo.

### DND-02 – Přenos do cizí aplikace (FEATURE/TECH, P2; PRIORITNÍ GRAFICKÉ EDITORY, POVINNÝ FALLBACK)

**Požadované UX – SCHVÁLENO:** vybrat text, bitmapu, vektor nebo stranu v PDF Blacksmith, **vizuálně je přetáhnout do otevřeného profesionálního editoru** a zvolit požadovanou podobu obsahu. Cílem je, aby se chování blížilo přetahování běžných souborů z Průzkumníka Windows. **Výslovně prioritní testované cíle:** CorelDRAW, Adobe Photoshop, Affinity Designer a Affinity Photo (v konkrétních dostupných verzích); rozšířit na další aplikace obdobně přijímající PDF, texty a obrázky. Nepovažovat tyto editory automaticky za nepodporované jen proto, že jde o externí okna.

**Primární přenos – technicky preferovaný:** standardní Windows OLE drag-and-drop pomocí `IDataObject`, `IDropSource`, `DoDragDrop`, `CF_HDROP` (reálný dočasný soubor vytvořený před předáním) a podle možností `CFSTR_FILEDESCRIPTOR` + `CFSTR_FILECONTENTS` (virtuální soubor). Pro text navíc `CF_UNICODETEXT`, RTF a HTML, pro bitmapy například DIB/PNG a pro vektory samostatný PDF fragment; nabízet jen skutečně dostupné formáty. Cílový editor vyhodnotí příjem podporovaných dat sám. Použít `IDragSourceHelper` nebo podobný mechanismus pro plovoucí miniaturu, není nutné ji simulovat nad cílovou aplikací vlastní záchytnou vrstvou.

**Volba formátu po uvolnění:** zachovat jako požadované UX tam, kde se podaří bezpečně oddělit volbu a dokončení operace. Implementovat **ověřovací prototyp**, protože při nativním OLE přetažení cizí aplikace obvykle zpracuje `Drop` již při uvolnění tlačítka. Není přípustné tvrdit, že libovolný editor automaticky přijme odložené vložení, vyvolávané dalšími syntetickými událostmi. Pro konkrétní nekompatibilní editor zdokumentovat odchylku a nabídnout uživateli volbu formátu v nejbližším technicky proveditelném okamžiku nebo náhradní uložení, nikoliv tiché vložení v nesprávném formátu.

**Co předat:**

- **Editovatelný text:** prostý UTF-8 text / RTF nebo HTML s formátováním, případně PDF pro graficky věrnou podobu. Nezaručovat, že cílový program po importu zachová editovatelnost textu či všechny fonty; ověřit na testovacích souborech.
- **Rastrový obraz:** pokud lze korektně extrahovat původní obrazová data, nabídnout původní formát; jako bezpečnou bezztrátovou volbu PNG. JPEG jen se známým dopadem na kvalitu.
- **Vektor / text převedený na křivky:** primárně PDF fragment se zachováním kreslicích příkazů a případných dostupných barevných profilů; nesmí dojít k automatické rasterizaci na JPEG jen proto, že cílovou aplikací je Photoshop.
- **Jedna či více stran:** samostatné PDF podle DOC-04. Pokud aplikace importuje PDF jen jako obraz nebo Smart Object, je to chování cíle; PDF Blacksmith nesmí nepozorovaně poškodit vlastní zdrojový obsah.

**Povinná reakce při neúspěchu – SCHVÁLENO:** pokud cílový program předání odmítne, není rozpoznán správný cíl, nebo nelze po volbě bezpečně dokončit vložení, zobrazit například:

> **Obsah se nepodařilo vložit do vybrané aplikace.** Můžete jej uložit jako soubor a otevřít nebo přetáhnout ručně.

Nabídnout **Uložit jako…** (otevře systémový dialog; uživatel může vybrat jakékoli přístupné umístění a vhodný formát), **Kopírovat do schránky** (je-li relevantní) a **Zrušit**. Nabídku zobrazit i při odmítnutí OLE cílem, ale *nepředstírat, že cíl skutečně obsah přijal*. Při úspěšném uložení oznámit umístění. Zdrojový dokument ani výběr nezměnit. Nepřepisovat systémovou schránku nebo otevřený cizí dokument bez vyžádané akce. Nepokoušet se obejít bezpečnostní hranice Windows.

**Povinné testy:** u každého prioritního editoru rozlišit drop na plátno vs. titulní lištu/prázdné okno; import PNG/JPEG/PDF, zachování průhlednosti, PDF jako vložený dokument/Smart Object vs. rastrová vrstva, text TXT/RTF, selhání nepodporovaného souboru, jehož výsledkem je dialog „Uložit jako…“, a zrušení bez ztráty dat. Zjistit, zda cílový editor soubor **kopíruje nebo na něj dlouhodobě odkazuje**; v druhém případě nabídnout trvalé uložení místo dočasného souboru, který by byl při ukončení PDF Blacksmith smazán. Testovat stav běžného i zvýšeného oprávnění a zobrazit chybu, pokud neprojde předání.

### DND-03 – Přetažení do Průzkumníka / na plochu (FEATURE/TECH, P2; PO UVOLNĚNÍ PREFEROVÁNO)

Pro **skutečnou lokální složku s ověřenou zapisovatelnou cestou** a pro plochu: po uvolnění myši zobrazit kontextovou nabídku „**Uložit jako…**“, výběr odpovídajícího formátu a název souboru; až po potvrzení soubor uložit do rozpoznaného cílového umístění. Na výběr například `.txt` (UTF-8 bez formátování), `.rtf` (formátovaný text), odpovídající originální obrazový formát / `.png` a `.pdf` (vektory, stránky). Formát `.txt` nikdy neoznačovat za formátovaný. Veškeré generování začít až při potřebě nebo využít dočasnou souborovou mezivrstvu podle DND-00.

**Rozpoznání cíle je důležité:** z pouhé pozice kurzoru nad Průzkumníkem není zaručeno, že aplikace zná přesnou cílovou složku (např. položky v levém stromu, virtuální složky, ZIP, OneDrive nebo plocha). Počítat se standardním OLE přenosem virtuálního souboru (`CFSTR_FILEDESCRIPTOR` + `CFSTR_FILECONTENTS`) nebo `CF_HDROP` tam, kde přesné určení cílové cesty není možné. V takovém případě lze nabídnout formát před finalizací kompatibilního přenosu nebo nabídnout výběr cílového adresáře; **nikdy neukládat jen podle odhadu okna či ikonky**. Testovat Průzkumníka samostatně od cizích editorů.

**Akceptace:** Ve známé lokální složce se po uvolnění nabídne formát + předvyplněný editovatelný název; při potvrzení vznikne přesně jeden správný soubor a během zrušení nevznikne žádný. U neznámého cíle aplikace transparentně oznámí omezení a nabídne bezpečnou alternativu. Ověřit kolize názvů, přípony, dlouhé cesty, znaky mimo ASCII, více vybraných prvků a restart/úklid dočasných souborů.

## 5. Dokumenty, stránky a ukládání

### DOC-01 – Duální zobrazení a názvy souborů (UX, P1)

V dialogu „Zobrazit vedle sebe“ zobrazit záhlaví **Vyberte druhý soubor** a seznam přívětivých názvů souborů bez úplných cest. Při shodném názvu zobrazit krátké rozlišení, např. rodičovský adresář v tooltipu nebo za názvem; nikdy nepřimět uživatele hádat, kterou kartu vybral. Nezobrazovat současný soubor jako druhý.

### DOC-02 – Přesun náhledů (BUG/UX, P0)

Při viditelných náhledech povolit přesouvání stránek **ve všech relevantních režimech**, nejen při aktivaci panelu Pořadí stránek, pokud uživatel netáhne text/objekt v hlavním dokumentu. Stránky označovat **levým kliknutím**, více stránek pomocí `Ctrl + levý klik` nebo rozsah pomocí `Shift + levý klik` podle SEL-02. Jednotlivé i vícenásobně vybrané stránky přesouvat **levým tažením**, zachovat pořadí označených stránek. Levý klik zároveň aktivuje příslušnou stránku pro zobrazení; prostřední klik na miniaturu není požadován.

**Povinné potvrzení před každým skutečným přesunem – SCHVÁLENO (Q17):** Po uvolnění levého tlačítka nad platným cílovým místem v miniaturách nebo po zadání nové cílové pozice **nejprve vypočítat navrhované konečné pořadí, ale neprovést zápis**. Otevřít modální dialog s názvem **„Potvrdit přesunutí stránek“**, který musí obsahovat:

- **Počet přesouvaných stránek** (např. „Přesouváte 3 stránky“).
- **Jejich aktuální čísla** v dokumentu (např. „Stránky 3–5“); nesouvislý výběr vždy ukázat přesně, například „Stránky 2, 5 a 8“, nikoliv zavádějící rozsah.
- **Cílovou pozici v konečném pořadí**: jednoznačně formulovat „na pozici 12“ jako pozici první z přesouvaných stránek **po dokončení přesunu**, nikoliv index původního kurzoru u hranice miniatur. Pokud přehlednost vyžaduje, formulovat i „za stránku… / před stránku…“.
- **Tlačítka „Přesunout“ a „Zrušit“**; žádné „příště se neptat“. `Esc` a zavření dialogu křížkem = „Zrušit“. Tlačítko „Přesunout“ musí být výchozím potvrzením až ve správně připraveném dialogu.

**Příklad:** „Opravdu chcete přesunout **3 stránky (3–5)** na **pozici 12**?“ Dialog nepoužije nesprávnou větu „stránky 3–5“ pro nesouvislou množinu. **Zrušení** ponechá obsah, pořadí, výběr i historii beze změny; nevytváří příznak neuloženého dokumentu. **Potvrzení** provede přesun jako jedinou atomickou, vratnou operaci. Na poškozený či zastaralý cíl přetažení nereagovat nesprávným přesunem; ověřit, že zdroj a stav dokumentu jsou stále platné. Pokud vypočítané pořadí odpovídá původnímu, jde o **no-op**: nevytvářet historii, příznak změny ani zbytečný potvrzovací dialog.

**Undo/Redo:** každé potvrzené přeuspořádání jako jeden atomický krok historii příslušného dokumentu. Testovat opakované přesuny, smazání, rotaci, přenos mezi dokumenty a návrat `Ctrl+Z` / `Ctrl+Y`. Současný kód v `pages_windows.go` již historii pořadí zapisuje; nejprve opravit její propojení.

### DOC-03 – Ukládání a aktivace nové kopie (BUG/UX, P0; SCHVÁLENO Q18)

V panelu Pořadí stránek i na dostupných standardních místech nabídnout **Uložit** = uložit změny do **aktuálně aktivního souboru** (se zajištěním bezpečného zápisu) a **Uložit upravenou kopii…** = vytvořit nový soubor při zachování původního, **poté nově vytvořenou kopii aktivovat v právě používané kartě**. V panelu lze zobrazit ikony (disketa; disketa se šipkou), s čitelnými tooltipy, klávesnicovými zkratkami a konzistentními názvy napříč aplikací. Příkaz „Uložit upravenou kopii…“ je v tomto kontextu funkčně blízký „Uložit jako“, nikoliv anonymnímu exportu bez přepnutí aktivního dokumentu.

**Schválený scénář:** Otevřít `Katalog.pdf` → upravit → **Uložit upravenou kopii…** jako `Katalog-upraveno.pdf` → po prokazatelně úspěšném uložení a ověření nové kopie se **původní `Katalog.pdf` uzavře a jeho karta se přímo nahradí novou kopií** (nevznikne druhá karta se starým dokumentem) → další `Ctrl+S` ukládá pouze do `Katalog-upraveno.pdf`. Soubor `Katalog.pdf` zůstane nedotčený. Výchozí návrh názvu lze použít, ale uživatel musí mít možnost zvolit vlastní název a složku. Při dalším „Uložit upravenou kopii…“ se stejným způsobem přejde na novou variantu.

**Bezpečnost a návaznost:**

1. **Původní dokument nezavírat předčasně.** Zavřít jej a v jeho stávající kartě zobrazit kopii **až po potvrzeném dokončení zápisu a ověření otevření výstupního PDF** (např. validní struktura a čitelnost). Selže-li zápis nebo otevření, nepřepínat na nefunkční soubor a nezahazovat rozpracované změny; nabídnout opakování / výběr jiné cesty.
2. **Žádná duplicitní karta originálu.** Původní pracovní karta přestane odkazovat na originál a stane se kartou kopie; původní dokument se zavře včetně jeho původních čtecích prostředků a kontextu. Ostatní již otevřené dokumenty zůstávají nedotčené. V titulku okna, kartě, seznamu nedávných dokumentů a následných příkazech `Uložit` změnit aktivní cestu na novou kopii. Zachovat stránku, zoom, orientaci zobrazení a odpovídající výběr, pokud je bezpečně obnovitelný. Zachování Undo/Redo po přepnutí vyžaduje explicitní ověření konzistence před a po serializaci; nikdy nesmí vracet změnu do původního souboru.
3. Původní PDF ponechat beze změny. U podepsaného, zamčeného nebo šifrovaného dokumentu předem upozornit na případnou neplatnost podpisu, omezení oprávnění nebo rizika související se šifrováním. Nikdy neztratit ochranné vlastnosti bez souhlasu.
4. Zápis cílového PDF provádět bezpečně přes dočasný soubor **na stejném svazku** a následné ověřené nahrazení; nepřepisovat napůl. Při souběžné změně cíle jiným programem hlásit konflikt a nepřepisovat data nepozorovaně. Při selhání zápisu obnovovací data nemažte.
5. Kontrolovat existující implementaci v `file_menu_windows.go` (`saveDocument`, `reopenSaved`) a `workspace_windows.go` (zpracování výsledku `save_document`): **základ přepnutí na uložený výstup již existuje**, ale musí se ověřit, že správně funguje pro kopii, pro selhání zápisu, pro rozpracovanou editaci a pro uložení z panelu Pořadí stránek. Nepřepisovat zbytečně fungující část.

**Akceptační test:** Po uložení kopie ověřit, že **původní PDF už není otevřený dokument v dané kartě ani samostatná další karta**. Změnit ještě jednu stránku a stisknout `Ctrl+S`; novou úpravu musí obsahovat **jen kopie**, nikoli původní soubor. Při stornu dialogu, nedostatku místa, zamítnutí oprávnění, chybném výstupním PDF nebo chybě opětovného otevření musí být původní i rozpracovaný dokument bezpečné.

### DOC-06 – Automatické obnovovací kopie / AutoRecover (SCHVÁLENO, P0 pro ochranu dat; Q19)

**Cíl:** Uživatel může bezpečně pracovat na PDF bez průběžného ručního ukládání a současně nezměnit původní dokument. Po záměrném uzavření volbou **Neuložit** nezůstanou obnovitelné změny, ale po pádu aplikace, ukončení procesu či výpadku napájení má být dostupná poslední validní obnovovací verze. **AutoRecover není totéž co automatické přepsání originálu.**

**Oddělit tři druhy souborů:**

- **Uložené dokumenty:** skutečné původní PDF a explicitně vytvořené kopie, které vlastní uživatel; aplikace je nikdy neodstraní jako „dočasná data“.
- **Přenosové a běžné pracovní soubory:** již používané dočasné výstupy pro editace, OCR a OLE přenos (`os.MkdirTemp` / `%TEMP%`); nejsou samy o sobě spolehlivou obnovou po pádu.
- **Trvalejší snímky pro obnovu:** samostatně ve **`%LOCALAPPDATA%\PDFBlacksmith\Recovery\<session-id>\<document-id>\`**. Nepoužít Dokumenty; samotné `%TEMP%` může Windows či čistič vymazat. Výchozí kořen je schválen; respektovat soukromí, oprávnění a případnou volbu umístění v Nastavení.

**Návrh mechanismu:**

1. Po první **obsahové změně** vytvořit obnovovací záznam s identitou relace a dokumentu. Automatická kopie nesmí být aktivním cílem příkazu `Ctrl+S` a nesmí aktualizovat originál. Změny pouhého výběru textu nebo výběru objektů, které nezměnily obsah, nejsou důvodem vytvářet nové plné PDF; lze však zachovat stav výběru jako drobná metadata relace.
2. Zachovávat především **kompaktní žurnál operací a/nebo konzistentní snímky stavu editace**; při jejich vytváření se nesmí zploštit vektorový obsah, změnit formátování nebo barvy PDF. Kvůli velkým PDF negenerovat celý soubor na každé stisknutí klávesy. Snímky připravovat mimo UI vlákno a používat omezení počtu úloh.
3. **Schválené výchozí časování:** po poslední změně a přibližně **60 s nečinnosti**, během nepřetržité práce nejpozději po **2 minutách**; bezprostředně po větších dokončených operacích (např. přesun stránek, OCR) a před rizikovým restartem/ukončením, pokud to lze bezpečně provést. V Nastavení → Soubory umožnit zapnout/vypnout a případně nastavit interval. Pokud autosave nestihne doběhnout před náhlým výpadkem, obnoví se poslední dokončený snímek, nikoli nutně poslední úhoz.
4. **Neuložený rozepsaný text** musí být součástí obnovovacího stavu, včetně použitých stylů, pozice kurzoru a výběru, pokud jeho editace dosud nebyla potvrzena do dokumentového modelu. Snímek nesmí nechtěně potvrdit nebo uzavřít běžící editor.
5. Zápis snímku musí být **atomický** (nová dočasná verze → kontrola čitelnosti/checksum → přepnutí manifestu). Uchovat **poslední dvě validní verze ve výchozím nastavení**, aby přerušený zápis nebo poškozený novější snímek nezničil poslední použitelnou. U plné kopie hlídat dostatek místa a velikost mezipaměti.
6. Při běžném **Uložit** nebo **Uložit upravenou kopii…** nejprve ověřit dokončení zápisu výsledného PDF; teprve poté lze odstranit redundantní staré obnovovací kopie. **Při „Uložit upravenou kopii…“ nejprve zprovoznit novou kopii ve stávající kartě, poté uzavřít pracovní kontext původního PDF bez vytvoření druhé karty.** Staré obnovovací záznamy lze odstranit jen tehdy, když jsou všechna potřebná data pro kopii bezpečně uložena a dokončena migrace relace. Pokud po uložení uživatel provede další změny, začíná nový obnovovací cyklus výhradně pro novou kopii.
7. Při vědomém uzavření karty/ukončení aplikace s volbou **Neuložit** odstranit obnovovací kopie **jen daného dokumentu**. Volba **Zrušit** ponechá otevřený dokument i kopie. Při úspěšném **Uložit** kopie odstraní jen tehdy, když je vše skutečně uloženo a žádná nová úprava mezitím nevznikla. Uložené originály a kopie se neodstraňují nikdy.
8. **Pád / násilné ukončení / výpadek:** při příštím spuštění detekovat nedokončené relace a nabídnout přehled obnovitelných dokumentů s názvem, časem posledního snímku a původní cestou. Volby **Obnovit**, **Zahodit** a **Přeskočit** (ponechat pro pozdější rozhodnutí). Obnovu zprvu otevřít jako neuložený pracovní stav, **nikdy bez potvrzení nepřepsat původní PDF**. Pokud je zdrojový soubor nedostupný či změněný, upozornit a nabídnout bezpečné uložení obnovené verze jako nové PDF.
9. **Soukromí a oprávnění:** ukládat do uživatelského profilu s vhodným přístupem. Nepřepisovat ani veřejně neodkládat přístupová hesla; u šifrovaných nebo citlivých dokumentů zajistit přiměřenou ochranu obnovovacích dat (např. zašifrování přes Windows DPAPI nebo odpovídající mechanismus), jinak uživatele upozornit a nenabízet falešně bezpečné automatické ukládání. Po pádu automaticky nemažte snímky dřív, než se uživatel rozhodne.
10. **Chyby / úklid:** při nedostatku prostoru nebo chybě ukládání oznámit „Automatickou obnovu se nepodařilo aktualizovat. Poslední změny nemusejí být chráněny.“ Nezaměňovat tuto chybu za chybu běžného uložení PDF. Osiřelé snímky nabízet k obnovení; stará data mazat teprve po zjištění, že nejsou aktivně používaná, ne automaticky po samotném startu. Čištění `%TEMP%` pro přetahování/bezpečné drafty musí být **oddělené** od čištění Recovery.

**Vazba na současný zdrojový kód:** `ocr_windows.go` tvoří návrhy pomocí `documentDraftPath()` do `os.MkdirTemp`, `main_windows.go` na korektní ukončení volá `cleanupDrafts()`. Tyto dočasné soubory již umožňují rozpracované úpravy, ale samy nemají definovanou trvalou obnovu po neočekávaném ukončení. Nový AutoRecover má doplnit stávající architekturu, nikoli bezdůvodně kopírovat veškerý obsah a odstranit současný Undo/Redo.

**Povinné testy před schválením produkční implementace:** několik otevřených dokumentů; změny pořadí stran; vícenásobná editace bez uložení; rozepsaný text; vypnutí procesu během zápisu; restart Windows; uložení upravené kopie a následná práce na ní; zrušení ukládání; volba Neuložit; chyba disku; nečitelný poslední snímek (fallback na předchozí); změna zdrojového PDF mimo aplikaci; PDF s podpisem, s heslem a s ICC profily; zachování vektorů, textů a rotací. Při běžném ukončení bez uložení nesmí docházet k nabídce obnovení záznamu, který byl vědomě zahozen.

**Stav rozhodnutí:** **Q19 schválen.** Codex má AutoRecover implementovat podle pravidel v DOC-06. Technické detaily souběžného zápisu, zabezpečení a atomických snímků se mají před nasazením prověřit testy. **Q18 upřesněno:** originál se po bezpečném přepnutí na uloženou kopii zavře a nesmí zůstat otevřený jako další karta.

### DOC-04 – Přetažení stránek mimo PDF Blacksmith (FEATURE, P2)

Vybrané stránky převedené do **samostatného PDF** poskytovat při přetažení do jiné aplikace jako souborový obsah; zda cílová aplikace dokument přijme, určuje její podpora drag-and-drop. Název generovat z názvu původního dokumentu a z původních zobrazovaných čísel či rozsahu stran, např. `Dokument_str_003.pdf` / `Dokument_str_003-005.pdf`. Při nesouvislých stránkách používat jednoznačný seznam nebo více souborů, nikdy zavádějící rozsah.

**Přetažení do Průzkumníka nebo na plochu:** u rozpoznané cílové složky po uvolnění myši nabídnout **Uložit jako: `[přednastavený název].pdf`** s možností název upravit a tlačítky **Uložit / Zrušit**. Použít společný mechanismus DND-00/DND-03; u nerozpoznaných či virtuálních cílů použít kompatibilní OLE nebo bezpečnou alternativu bez hádání cesty. U cizích aplikací se soubor předává podle jejich podpory OLE, prioritně otestovat CorelDRAW, Photoshop a Affinity. Nabídka po uvolnění je požadovaný cíl, který je nutné ověřit prototypem. **Když vložení není možné, informovat uživatele a nabídnout „Uložit jako…“ s libovolnou cílovou složkou.** Zachovat vektory, text, barevné prostory a relevantní metadata podle možností PDF.

### DOC-05 – Zavírání dokumentů v jednoduchém a duálním zobrazení (UX/TECH, P1; SCHVÁLENO Q20)

**Záměr – SCHVÁLENO:** Nahradit nejasné „Ano/Ne“ explicitně pojmenovanými akcemi a bezpečně navázat ochranu neuložených změn na AutoRecover (DOC-06). **Schváleno Q20:** při zavírání více upravených dokumentů se používá **jeden souhrnný modální dialog**, nikoliv posloupnost samostatných potvrzení po dokumentech.

**Krok 1 – určit, co zavírat:**

- **Duální zobrazení:** **Zavřít levý dokument**, **Zavřít pravý dokument**, **Zavřít všechny dokumenty**, **Zrušit**. U levé/pravé volby zobrazit jednoznačně **jméno konkrétního souboru** vedle akce, například `Katalog.pdf` / `Cenik.pdf`. Pokud mají dokumenty stejné názvy, doplnit rozlišovací cestu do tooltipu nebo detailu, nikoliv samotný název zaměnit za celou cestu.
- **Jednoduché zobrazení:** **Zavřít aktuální dokument**, **Zrušit**. Pokud jsou další PDF otevřená v kartách, ale právě nejsou vidět vedle sebe, navíc **Zavřít všechny dokumenty**. Pokud je otevřený právě jeden dokument, „Zavřít všechny“ je duplicitní a nezobrazí se (doporučení k potvrzení).
- **Zavřít všechny dokumenty** znamená všechny otevřené PDF karty, nejen právě viditelnou dvojici. Při příkazu **Zavřít** ve smyslu zavření okna aplikace se po úspěšném dokončení všech kontrol také ukončí aplikace; při zavření konkrétního PDF zůstane aplikace otevřená.
- Po zavření jednoho ze dvou současně zobrazených PDF se zbývající dokument zobrazí na dostupné šířce, bez prázdného duálního slotu. Ostatní nesouvisející karty zůstanou zachovány.
- Zavírání přes křížek konkrétní karty má jednoznačný cíl – danou kartu – a nevyžaduje navíc dialog „který soubor“, ale při neuložených změnách musí projít krokem 2. Příkaz Escape v dialogu jej bezpečně zruší.

**Krok 2 – rozhodnout o neuložených změnách:**

- Pokud žádný z dotčených dokumentů nemá obsahové neuložené změny ani rozpracovanou editaci textu, zavřít bez dalšího dotazu.
- Pokud je špinavý jeden dokument, nabídnout **Uložit** (do aktuálního PDF), **Uložit upravenou kopii…** (bez přepsání aktuálního PDF), **Neuložit** (zahodit neuložené změny a jejich obnovovací data), **Zrušit** (neprovést zavření). U příkazů, které mohou změnit podpisy či ochranu PDF, uplatnit již existující upozornění a bezpečnostní pravidla.
- Pokud se zavírá více upravených dokumentů, zobrazit **jeden souhrnný modální dialog – SCHVÁLENO**: seznam přesně identifikovaných dokumentů a u každého zvolená akce **Uložit / Uložit upravenou kopii… / Neuložit**. Nahoře vysvětlit, že se budou zpracovávat pouze uvedené dokumenty; dole **Pokračovat** / **Zrušit**. Dovolí-li to UX, výchozí volba nesmí způsobit skryté mazání práce: preferovat bezpečnou volbu Uložit nebo vyžádat explicitní volbu.
- **„Neuložit“ není automatický default.** Záměrné Neuložit odstraní odpovídající AutoRecover snímky daného dokumentu až po bezpečném ukončení relace; tím nesmí být dotčeny snapshoty ostatních karet. Samotné výběry (bez změny obsahu) nevyvolávají dialog uložení.
- Pokud dojde k chybě, stornu, konfliktu výstupního souboru nebo nemožnosti zapsat některý dokument, **nezahazovat neuložené změny žádného dosud nezavřeného dokumentu**. Zastavit zbytek zavírání, ponechat uživatele v aplikaci, nabídnout opakování nebo Uložit upravenou kopii…; již bezpečně zapsané soubory mohou zůstat uložené, ale přesný výsledek musí být viditelný. U vícedokumentového zavírání preferovat nejprve dokončení všech uložení, teprve pak odstranit nebo zavřít příslušné karty, aby nedocházelo k částečné ztrátě rozpracované relace.
- Pokud uživatel při zavírání zvolí **Uložit upravenou kopii…**, uplatní se DOC-03: nejprve bezpečně vytvořit a ověřit kopii a převázat identitu karty, teprve potom dokončit její zavření; původní diskový soubor se nezmění.
- Dialogy nepoužívat se zavádějícími volbami „Ano / Ne“. Příkaz zavřít aplikaci ani jednotlivou kartu nesmí být ekvivalentem nevratného smazání obnovovacích souborů bez výslovné volby Neuložit.

**Stav aktuálního kódu (kontrola 20. 11. 2026):** `document_tabs_windows.go` má `closeWindow` s dotazem „Ano zavře všechny karty, Ne pouze aktivní“ a `confirmCloseDocuments` s dalšími potvrzeními Ano/Ne/Zrušit; `continueCloseDocuments` ukládá do originálu. První krok Codexu: refaktorovat tuto konkrétní cestu bez zrušení ověřené historie, stavu karet a asynchronního ukládání; napojit na DOC-03, DOC-06. **Z pozorovaného kódu nevyplývá, že zcela funguje uvedené cílové chování na Windows – nutná reprodukce a testy.**

**Akceptace:** zavřít pouze levý / pouze pravý PDF; stejné názvy v různých adresářích; zavřít všechny karty, včetně skrytých; jeden špinavý a jeden čistý; dva špinavé s rozdílnou volbou; Neuložit smaže jen recovery dotčeného dokumentu; Zrušit zachová všechny karty a nedokončenou editaci; chyba uložení nic nezahodí; úspěšné Uložit upravenou kopii nemění originál; po zavření jednoho dokumentu se duální plocha přepočítá; zavření okna programu přes křížek má bezpečné chování i při běžící úloze. Testy klávesnice, Esc, DPI a správného titulku dialogu.

## 6. Rozhraní a chování hlavního okna

### UI-01 – Rozbalovací nabídka Soubor (BUG/UX, P1)

Po opuštění kurzorem zavřít nabídku, **ale ne při přejezdu z tlačítka Soubor na její vlastní položky**. Použít společný hover region a krátké zpoždění proti náhodnému zmizení; okamžitě zavřít při kliknutí mimo, Escape či aktivaci jiného nástroje. Zachovat ovládání klávesnicí a schopnost nabídku otevřít i kliknutím.

### UI-02 – Změna velikosti okna (BUG, P0)

Opravit tažení pravého okraje i pravého horního/dolního rohu; otestovat všechny čtyři hrany a rohy, maximalizaci, obnovení a více monitorů s různým DPI. V `chrome_windows.go` metoda `hitTest` již vrací Windows hit-test kódy pro okraje; prověřit, zda zprávy nepohlcuje `pointer`, nevhodná interakce s neklientskou oblastí nebo podřízené ovládací prvky. V `main_windows.go` je použito okno se změnitelnou velikostí. Neopravovat naslepo samotnou konstantu stylu.

### UI-03 – Jednotný vzhled nabídek (UX, P1)

Rozbalovací nabídky, kontextová menu a seznamy výběru sjednotit podle decentního vizuálního stylu nabídky Soubor: pozadí, kontrast, hover, disabled stav, oddělovače, vnitřní odsazení, poloměr rohů, responzivní šířka. Rozlišovat otevřená menu a běžné combo boxy; důležitá je konzistence a dostupnost, nikoliv jen kosmetické sjednocení.

### UI-04 – Sbalování panelů, svislé lišty a ouška (UX, P1; SCHVÁLENO Q7)

**Schválený základ – varianta B:** Po sbalení hlavního nebo pomocného panelu se panelový obsah skryje, avšak **úzká svislá lišta zůstane po celé výšce příslušné oblasti**, aby bylo patrné, že se panel dá znovu otevřít. Není dovoleno zmenšit sbalený panel jen na osamocené ouško plovoucí na okraji dokumentu. Udržet stejné pravidlo vlevo i vpravo a pro pomocné panely (například náhledy stránek a záložky), pokud jsou v daném uspořádání dostupné.

**Dva vizuálně odlišné typy oušek:**

1. **Hlavní panely (např. Nástroje / Nastavení):** Ouško je **malý čtverec s mírně zaoblenými rohy** a pouze **šipkou** pro sbalení či rozbalení. Nepoužívat v něm nápisy „Levá strana“ ani „Pravá strana“; textový význam zůstane dostupný v tooltipu a přes přístupný název tlačítka. Ouško je umístěné nahoře u příslušné svislé lišty. Směr šipky ukazuje akci po kliknutí a odpovídá levé/pravé straně.
2. **Pomocné panely (např. Náhledy stránek / Záložky):** **Každé ouško obsahuje viditelný text identifikující daný panel** – jak při sbaleném, tak při rozbaleném stavu. Jde o štítek panelu, nikoli o samotnou šipku. Případná šipka může text doplňovat, ale nesmí ho nahrazovat. Protože zůstává jen úzká lišta, štítky lze orientovat svisle nebo jako úzké podélné záložky, **tak aby byly celé čitelné a nezasahovaly do textu jiných panelů**. Konkrétní orientaci vybrat až po kontrole skutečného UI v různých velikostech okna, nejde o další schválenou uživatelskou preferenci.

**Rozlišení stavu u pomocných panelů:**

- **Otevřený / aktivní panel:** ouško má **stejnou barvu podkladu jako příslušný rozbalený panel** a vizuálně tvoří jeho součást; jemná souvislá kontura obíhá panel i vnější obvod ouška bez zdvojeného předělu v místě napojení.
- **Sbalený / neaktivní panel:** jeho ouško má **jiný, tlumenější odstín** než aktivní ouško, aby bylo patrné, že je panel skrytý. I ve sbaleném stavu zůstává textový název čitelný a ouško lze kliknutím aktivovat.
- Je-li sbaleno více pomocných panelů, jejich štítky jsou stále jednotlivě přístupné; nemají se vzájemně překrýt. Pokud není žádný rozbalený, žádné skryté ouško se nesmí falešně tvářit jako aktivní.

**Chování a rozvržení:**

- **Kliknutí** na příslušné ouško okamžitě sbalí nebo rozbalí konkrétní panel; samotný přejezd kurzorem nemá panely nechtěně vysouvat. Ukázat správný tooltip (např. „Zobrazit záložky“ / „Skrýt záložky“).
- Šířku úzké lišty, velikost čtvercového hlavního ouška, čitelnost názvů pomocných panelů a klikací oblast definovat v **DPI-aware** logických rozměrech. Nepřekrývat dokument ani ovládací prvky sousedních panelů.
- Je-li hlavní panel otevřený, **jeho ouško může jemně vystupovat do prostoru pomocného panelu**, ale příslušný pomocný panel pro něj musí rezervovat místo. Při sbalení musí zůstat lišta a prvek pro rozbalení, bez zbytečně široké prázdné plochy.
- V duálním zobrazení a při změně velikosti okna, maximalizaci nebo změně DPI zachovat dostupnost všech ovladačů a použitelné rozměry plátna. Responzivní miniatury podle UI-08 mají nadále cílovou šířku **6 % celé klientské oblasti okna**; při kontrole, zda se vejdou, je nutné odečíst také prostor svislých lišt. Pravidlo 6 % platí pro samotný obrázek miniatury, pokud to zbývající prostor dovoluje.
- Zachovat ovládání myší i klávesnicí, viditelný stav focus a přístupné názvy; samotné barevné odlišení nesmí být jediným indikátorem otevření/skrytí (pomůže tvar/napojení a případně šipka).

**Kde začít v kódu:** `side_panels_windows.go` (`sidebarRailWidth`, `layoutDocumentPanels()`, `layoutRail()`, kreslení a obsluha postranních přepínačů), `workspace_windows.go` (`layout()`, viditelnost panelů) a případně `theme_windows.go` (odstíny aktivního/neaktivního stavu). Neprovádět plošný redesign všech barev mimo tento úkol.

**Akceptace:** Po sbalení kteréhokoli panelu zůstává příslušná úzká lišta. Hlavní panel má pouze zaoblené čtvercové ouško se šipkou; pomocné panely mají vždy textové názvy. Otevřený pomocný panel a jeho ouško mají totožný podklad a spojitou konturu; skrytá pomocná ouška mají odlišný odstín a lze je jednotlivě rozbalit. Žádná část nezasahuje do textu, ovládacích prvků nebo zobrazovaného PDF při typických velikostech oken a při 100/150/200 % DPI. Otestovat obě strany a kombinace otevřených/skrytých panelů.

### UI-05 – Fulltext v horní liště (UX, P1)

Přidat napravo kompaktní textové pole s ikonou lupy. Kliknutí do pole aktivuje panel vyhledávání s pokročilými volbami. Základní dotaz a výsledky mají sdílet stejný stav jako stávající fulltext; při přepnutí karty zobrazovat odpovídající dokument a jeho výsledky. Samostatné tlačítko vyhledávání v hlavním seznamu nástrojů odstranit jen pokud nezmizí žádná dostupná funkce. `Ctrl+F` zaměří horní hledání.

### UI-06 – Režimy čtení v horní liště (UX, P1)

Ikony pro **Celá stránka**, **Přizpůsobit šířce**, **Po jednotlivých stránkách**, **Souvislé zobrazení**, **Vedle sebe**. Rozlišit *přizpůsobení stránky do viewportu* od **skutečné celoobrazovkové aplikace** (`F11`), aby jedna ikona neměla dvojí význam. U přepínacích režimů zobrazit aktivní stav; zachovat aktuální stránku a pokud možno místo čtení. V duálním režimu lze nastavení měnit nezávisle pro oba dokumenty.

### UI-07 – Jednotné ikony, tooltipy a kontextová nápověda (UX, P1; SCHVÁLENO Q9)

**Schválený princip:** Tam, kde je význam zřejmý, používat tematické piktogramy namísto nadbytečných tlačítek s nápisem. **Pod sadou ikon je jedno společné trvalé pole podrobnější nápovědy**, které standardně vysvětluje **aktuálně aktivní nástroj**. Nápověda se při najetí myši na jinou ikonu **dočasně změní na popis právě najeté funkce** a **po odjetí z této ikony se vrátí k popisu aktivního nástroje**, aniž by se měnil samotný výběr nástroje. Toto chování uživatel výslovně schválil.

**Konkrétní pravidla:**

1. **Výchozí stav:** pole ukazuje název, význam a případné klávesové kombinace **aktivního nástroje**, dokud není jiná ikona najetá myší nebo zaměřená klávesnicí. Pokud ještě není vybrán žádný nástroj dané skupiny, zobrazit krátkou nápovědu ke skupině, nikoli prázdnou či chybnou informaci.
2. **Hover:** při najetí na jinou ikonu přepnout obsah pole na její vysvětlení **bez aktivace či změny nastavení**. Krátký tooltip nezávisle zobrazí název a případnou klávesovou zkratku. Při přejezdu přímo na další ikonu aktualizovat popis podle nové ikony; neproblikávat mezitím zbytečně základním popisem.
3. **Mouse leave:** po opuštění všech ikon příslušné skupiny obnovit text **skutečně aktivního nástroje**; přechod přes malou mezeru mezi ikonami nesmí vyvolat rušivé blikání. Nesmí zůstat viset vysvětlení už nenajeté funkce. Návrat nemění aktivní stav ani obsah dokumentu.
4. **Aktivace:** teprve levý klik na ikonu, příslušná zkratka či jiná platná akce změní aktivní nástroj. Pole se poté standardně váže na novou aktivní funkci; po ukončení hoveru se vrací k ní, nikoli k předchozímu nástroji.
5. **Klávesnice a přístupnost:** při pohybu fokusu klávesnicí na ikonu nabídnout stejné dočasné vysvětlení jako při hoveru; po odchodu fokusu obnovit nápovědu aktivní funkce, pokud žádná ikona není pod myší. Každá ikona musí mít přístupný textový název a viditelný focus. Esc a obsluha formulářových polí nesmí být narušeny.
6. **Disabled a destruktivní akce:** neaktivní ikona má tlumený stav a nesmí provést akci; nápověda může vysvětlit, proč není dostupná. U nejednoznačných, potenciálně destruktivních nebo potvrzovacích úkonů zachovat jasné textové označení v dialogu; ikonou nenahrazovat povinné potvrzení.
7. **Jednotný vizuální systém:** preferovat jednoduché škálovatelné jednobarevné piktogramy bez 3D efektů; orientačně ikona **20–22 logických px** v klikací ploše **36–40 logických px** (ověřit vůči stávajícímu DPI-aware rozhraní). Sjednotit aktivní, hover, focus, disabled a stisknutý stav; aktivní nástroj lze jemně zvýraznit vínově červenou, avšak také obrysem či jiným nebarevným signálem. Responzivní rozestupy a zalamování skupin nesmějí kolidovat s textem nápovědy.
8. **Žádné překrytí:** pod každou skupinou ikon vyhradit přiměřenou výšku pro delší popis. Pokud se text nevejde, použít zalomení a případně posouvání panelu; nenechat jej překrývat ikony nebo další ovládací prvky. Drobné přepnutí popisu nemá posouvat celé rozvržení panelu při každém najetí.
9. **Sdílená logika:** chování implementovat konzistentně pro panely **Text, Objekty, Pořadí stránek** a další skupiny nástrojů. Ideálně mít jednotnou reprezentaci `id nástroje → název, zkratka, krátký tooltip, podrobný popis`, nikoli různá neslučitelná řešení pro každý panel.

**Technická poznámka:** Pro Win32 oddělit `activeTool`, `hoveredTool` a `focusedTool`; zobrazovaný popis určit prioritou **právě hoverovaná ikona → klávesnicí zaměřená ikona → aktivní nástroj**. Nenastavovat `activeTool` při `WM_MOUSEMOVE` / `WM_MOUSELEAVE`. Omezit zbytečné překreslování a nepouštět opakované layouty při každém pixelu pohybu. Vycházet ze současných ikon a tooltipů (`text_ui_windows.go`, `text_windows.go`, `object_windows.go`, `pages_windows.go`, případně `workspace_windows.go`, `theme_windows.go`) a nejdříve identifikovat místa, kde už společný mechanismus existuje. Při použití systémového tooltipu dodržet jeho přirozené zpoždění; konkrétní dobu nevymýšlet bez testování.

**Akceptace:** Aktivní například nástroj **Obdélník** má pod ikonami svůj trvalý popis. Přejetí na **Elipsa** zobrazí popis elipsy, aniž by Elipsu aktivovalo. Po odjetí z ikon se okamžitě či plynule bez rušivého blikání vrátí popis Obdélníku. Kliknutí na Elipsu ji aktivuje a po odjetí zůstane její popis. Přejezd mezi ikonami, vypnutá ikona, tabulátor a Shift+Tab, DPI 100/150/200 %, zúžený panel a přepnutí jiného dokumentu nezpůsobí kolizi, nesprávné trvalé zvýraznění ani nechtěnou změnu PDF.

### UI-08 – Responzivní pás náhledů (UX, P1; SCHVÁLENO Q6)

**Schválený princip:** Šířka **samotného viditelného obrázku miniatury stránky** má v automatickém režimu odpovídat **6 % aktuální šířky klientské oblasti celého okna PDF Blacksmith**. Šířka *celého panelu* je o potřebný prostor pro okraje, případné číslo stránky, zvýraznění výběru a svislý posuvník větší. Nezaměňovat 6 % obrázku za 6 % panelu. Zachovat poměr stran miniatury; výška z něj vyplývá.

**Výpočet a chování:**

1. Při každé změně velikosti okna přepočítat cílovou šířku jako `thumbWidth = round(0.06 × clientWidth)` v souřadnicích klientské oblasti okna, s korektním DPI převodem; například při šířce 1 200 / 1 600 / 1 920 / 2 560 px odpovídá 72 / 96 / přibližně 115 / přibližně 154 px. Nepřidávat svévolný pevný strop 140 px: na širokých monitorech by porušoval požadované poměrné zvětšování.
2. Vypočítat šířku panelu `panelWidth = thumbWidth + horizontalPadding + prostor potřebný pro ovládání`, bez překrývání číslování stran, posuvníku či přepínačů. Čísla stránek lze umístit pod miniatury, pokud tím vznikne úspora místa.
3. Výsledek musí respektovat **minimální použitelnou šířku hlavního dokumentu**. V extrémně úzkém okně či duálním zobrazení smí být skutečná miniatura menší než cílových 6 %, případně se panel sbalí podle stávajícího způsobu ovládání. Nikdy ale nesmí zakrývat plátno či sousední panel. Toto je jediná nutná prostorová výjimka, nikoli další trvalá pevná maximální šířka.
4. Při zvětšení, zmenšení, maximalizaci, obnovení a změně DPI přepočítat zároveň panely nástrojů, souborový panel, náhledy, přepínače panelů a šířku zobrazované plochy dokumentu. Nedopustit překrývání prvků ani skok do nesprávného místa PDF; pokud možno zachovat zobrazenou stránku a pozici čtení.
5. **Duální zobrazení:** oba pásy náhledů pracují s vlastním stavem, ale jejich automatická cílová velikost vychází ze stejné celkové šířky okna. Při nedostatku prostoru nejdříve zachovat použitelnost obou dokumentů a přiměřeně miniatury zmenšit. Žádný panel nesmí vytlačit dokument z obrazovky.
6. Stávající `Ctrl + kolečko` nad miniaturami jako **ruční přiblížení miniatur** zachovat. Automatických 6 % je výchozí velikost; při explicitním ručním nastavení zachovat uživatelem zvolenou relativní změnu i po změně šířky okna, pokud to dovoluje prostor. Případný návrat na **Automatická velikost (6 %)** nabídnout přímo u náhledů. Při zkoušce jednoznačně rozlišit manuální a automatický režim.

**Kde začít v kódu:** `side_panels_windows.go` (`layoutDocumentPanels()`: nyní pevně volených 182/220 logických px), `thumbnails_windows.go` (`thumbnailPanelPixels`, `thumbnailScale`, `layoutThumbnails()`), dále `workspace_windows.go` (`layout()`). Nejprve odstranit kolizi pevné šířky panelu se šířkou obrázku; pak doplnit výpočet rozměru a regresní testy. Nenutit změnu samotné struktury PDF ani ukládání zobrazení do původního souboru.

**Akceptační kritéria:** Při šířce klientské oblasti 1 200/1 600/1 920/2 560 px naměřit v automatickém režimu šířku *obrázku* přibližně 72/96/115/154 px (pokud je dost prostoru), panel je širší o nutné ovládací prvky; při změně velikosti se aktualizuje bez překrytí dokumentu; obraz má správný poměr stran; změna DPI a dvojité zobrazení nerozbije ovládání; funguje ruční `Ctrl + kolečko` i návrat na automatický režim.

### UI-09 – Ikony Nastavení a Nápověda v horní liště (UX/FEATURE, P1; SCHVÁLENO Q10)

**Závazný rozsah:** Na pravou stranu horní lišty přidat dvě samostatné ikonové akce:

- **Nastavení:** jednoznačná ikona ozubeného kolečka (`settings`), přístupný popisek „Nastavení“, zkratka až po odsouhlasení výchozího mapování. Zpřístupňuje **globální** předvolby aplikace, nikoli vlastnosti aktivního nástroje z pravého pracovního panelu. Otevření nesmí rušit výběr, přepnout PDF kartu ani změnit režim nástroje.
- **Nápověda – SCHVÁLENO Q11:** ikona otazníku v kruhu (`help`) s přístupným popiskem „Nápověda“ otevírá **úvodní stránku celé příručky**. `F1` je výchozí kontextová zkratka pro **kapitolu k právě aktivnímu nástroji či zaměřenému ovládacímu prvku** (při neznámém kontextu bezpečný návrat na nástroj/úvod); ve vlastní konfiguraci zkratek ji lze přemapovat. Nápověda nesmí vyžadovat internetové připojení pro základní obsah.

**Rozvržení – navrženo:** Zleva napravo v pravé části lišty: rychlé fulltextové vyhledávání (UI-05), kompaktní skupina režimů zobrazení (UI-06), vizuální oddělovač, **Nápověda**, **Nastavení**; zachovat místo pro standardní ovládání okna a pro karty dokumentů. Při zúžení okna neukrývat důležitá tlačítka bez náhradního přístupu (např. menu „Více“). Ikony mají vizuálně odpovídat systému UI-07, s tooltipem, fokusem, aktivními a zakázanými stavy; žádná kolize s přetažením okna nebo zónami změny velikosti.

**Otevření globálních Nastavení – SCHVÁLENO:** Otevřít samostatné okno Nastavení s **levým svislým seznamem kategorií a pravým obsahem zvolené kategorie**. Nesmí nahrazovat pracovní panel parametrů aktivního nástroje ani měnit aktivní PDF, výběr, polohu kurzoru, zoom nebo nástroj. Okno musí mít vlastní zavírání a být dosažitelné také z klávesnice; při opakovaném kliknutí na ikonu se nemají vytvářet duplicity. Modalita je **SCHVÁLENA: modální okno**, které po dobu otevření blokuje vstupy do pracovního prostoru (viz Q13 a UI-10). Rozměry, minimální šířka a uložení velikosti/pozice zůstávají technickým návrhem. V rámci akceptace přezkoušet obnovení stavu po zavření i na více monitorech.

**Kde začít v kódu:** `chrome_windows.go` (layout hlavičky, titulková oblast, ikony), `workspace_windows.go` (vytváření a aktualizace ovladačů), `main_windows.go` (zprávy okna a hit-test), `document_commands_windows.go` (příkazy a zkratky), případně `theme_windows.go` (jednotný vzhled). Ověřit, zda stávající akce nebo dialogy Nastavení/Nápovědy již nejsou realizovány jinou cestou. Neimplementovat druhý paralelní systém zkratek.

**Akceptace:** Obě ikony jsou viditelné a ovladatelné myší i klávesnicí při běžném i zúženém okně a při 100/150/200 % DPI; tooltipy označují akce; kliknutí na Nastavení neukončí rozpracovanou editaci ani nezmění aktivní nástroj; Nápověda zůstává dostupná bez internetu; nedochází ke kolizi s Windows titulkovou lištou.

### UI-10 – Globální Nastavení, klávesové zkratky a informace o vydání (FEATURE/UX, P1–P2; ROZVRŽENÍ, MODALITA A OSM KATEGORIÍ SCHVÁLENO)

**Povinné, schválené jádro:**

1. **Verze a vydání:** v Nastavení zobrazit sekci **O aplikaci** s názvem „PDF Blacksmith“, číslem **skutečně spuštěné verze**, identifikátorem sestavení a datem vydání, pokud jsou k dispozici. Přidat **Poznámky k vydání / Co je nového** pro nainstalovanou verzi. Data získávat z jediného zdroje buildu/release metadat; nezaměňovat „nainstalovanou verzi“ za poslední vydání na GitHubu. Základní informace a poznámky mají být zabalené do aplikace pro offline použití; volitelné ruční vyhledání nové verze až po odsouhlasení. Nezobrazovat vymyšlené datum nebo číslo buildu, pokud neexistuje.
2. **Vlastní mapování kláves:** zobrazit přehled všech mapovatelných akcí s výchozí a aktuální zkratkou, oblastí platnosti (globální, text, objekty, stránky) a polem pro filtrování. Uživatel vybere akci, zadá kombinaci v režimu záznamu, uvidí **kolizi se stávajícím příkazem** a rozhodne, zda konflikt vyřešit přeřazením či operaci zrušit. Umožnit „Vrátit výchozí“ pro jednotlivou akci i všechny zkratky. Zajistit, že uživatel může změnu zrušit a že vždy existuje přístup k nastavení myší, i po nevhodném přemapování.
3. **Jednotná aktualizace nápovědy:** nová zkratka se musí promítnout do tooltipů, kontextových nabídek a seznamu zkratek. V předvoleném mapování zachovat schválená gesta a významy (`Ctrl+Z`/`Ctrl+Y`, `Ctrl+F`, `F11`, výběrové kombinace), pokud je uživatel výslovně nepřemapuje. **Modifikátory držené během kreslení (`Shift`, `Ctrl`, `Ctrl+Shift`) a přímá gesta myši jsou chování nástrojů, nikoli samostatné příkazy, proto se automaticky nestávají mapovatelnými klávesovými zkratkami.**
4. **Trvalost a bezpečnost:** nastavení ukládat pro aktuálního uživatele (navržené umístění `%APPDATA%\PDFBlacksmith\settings.json`, nikoli do složky dokumentů). Ověřit načtení při dalším spuštění, bezpečné a atomické ukládání, obnovu výchozích hodnot při poškození konfigurace, nepovolovat systémem vyhrazené kombinace a zachovat textové vstupy, kde by globální zkratka bránila psaní. Zkušebně otestovat české i anglické rozložení klávesnice.

**Schválená struktura okna a funkční pravidla:**

- **Levá navigační oblast:** svislý seznam kategorií, vždy jednoznačně zvýrazněná aktivní kategorie. Výběr myší a klávesnicí; v delším seznamu svislé posouvání. Jména kategorií čitelná při standardním i vyšším DPI.
- **Pravá obsahová oblast:** nadpis vybrané kategorie, logicky seskupené volby, stručné vysvětlení nastavení a odkazy na příslušnou Nápovědu. Větší skupiny mohou být rolovatelné, ale nesmějí překrývat ovládací prvky.
- **Společné akce – SCHVÁLENO:** dialog obsahuje tlačítka **OK / Použít / Zrušit** v jednotném spodním pruhu. **Nedokončené editace a rozepsaný dokument musejí zůstat zachovány.** Přímé změny globálních preferencí se nesmějí vydávat za obsahové změny PDF.
- **Zkratky:** v kategorii Klávesové zkratky zobrazit vyhledávací pole, tabulku příkazů, aktivní přiřazení, kontext platnosti, tlačítka Změnit / Vrátit výchozí a upozornění na konflikty. Nezaměňovat klávesové zkratky s gesty myši a s modifikátory během kreslení.
- **O aplikaci:** nabídnout skutečnou nainstalovanou verzi a přístup k informacím o vydání; duplikovat odkaz v Nápovědě je dovoleno, ale samotná data musejí mít jediný zdroj.

**Potvrzování Nastavení – SCHVÁLENO (Q12):**

1. Po otevření si dialog načte poslední **potvrzenou** konfiguraci. Editace ovládacích prvků mění nejprve pracovní kopii nastavení (stav „čeká na potvrzení“), nikoliv automaticky trvalé hodnoty na disku.
2. **OK:** validuje všechny čekající změny, atomicky je uloží, zavede je do aplikace a **zavře** okno. Pokud validace nebo zápis selže, dialog se nezavře; zobrazí srozumitelný důvod a umožní opravu.
3. **Použít:** provede tutéž validaci, uložení a aplikaci, ale **nechá okno otevřené**. Úspěšně uložený stav se stane novou výchozí hodnotou pro případné následné storno. Pokud žádná změna nečeká, tlačítko má být neaktivní.
4. **Zrušit:** zahodí **pouze nepotvrzené změny** vytvořené od posledního otevření nebo úspěšného stisku `Použít`; zruší dočasné náhledy těchto změn a zavře okno. **Nikdy nevrací nastavení, která už byla potvrzena tlačítkem `Použít`.** Klávesa `Esc` a zavření okna křížkem se mají řídit stejným pravidlem, pokud zrovna nejsou součástí jiného aktivního editoru či poddialogu.
5. **Živý náhled:** u bezpečně vratných voleb (například motiv, velikost ikon, velikost textu UI) smí uživatel vidět okamžitý náhled ještě před potvrzením. Náhled musí mít uloženou předchozí hodnotu a jít zcela vrátit pomocí `Zrušit`; u změn vyžadujících restart nebo u nevratných akcí náhled nesimulovat.
6. **Klávesové zkratky:** nově přiřazené kombinace se v dialogu připravují pouze jako návrh, s okamžitou kontrolou konfliktů. Do běžné aplikace a tooltipů se přenesou až tlačítkem `Použít` nebo `OK`. Neplatné či konfliktní nastavení bez vyřešení konfliktu nesmí být uloženo.
7. **Restart / výjimky:** pokud konkrétní volba nabude účinku až po restartu, přehledně to oznámit před potvrzením; nastavení se uloží až při `Použít` nebo `OK`, restart nesmí být proveden bez výslovného souhlasu. Jednorázové akce typu „Vymazat mezipaměť“ či „Zkontrolovat aktualizace“ nejsou nastavovací volby s náhledem – mají samostatné potvrzení a vysvětlení případné nevratnosti.
8. Přepnutí kategorie vlevo **nesmí zahodit čekající změny** jiné kategorie. Změna globální konfigurace nesmí vložit žádný krok do historie Undo/Redo PDF, změnit příznak upraveného dokumentu ani přepsat jeho obsah. Nastavení je uživatelské, nikoli automaticky vázané na právě otevřený dokument.

**Modalita globálního Nastavení – SCHVÁLENO (Q13):** Okno je **modální vůči hlavnímu oknu PDF Blacksmith**. Je-li otevřené, hlavní okno zůstává viditelné, ale nelze v něm klikat, psát, měnit výběr, aktivovat nástroje, provádět dokumentové zkratky, přetahovat objekty ani zahájit akce měnící PDF. Živý náhled bezpečně vratných nastavení (např. motiv a velikost ikon) se může ve viditelném hlavním okně vykreslit, aniž je zpřístupněna editace PDF. Modální okno patří hlavnímu oknu jako vlastník; nemá být nezávislým oknem, které lze ztratit za aplikací. Při `Alt+Tab` a při práci s více monitory zůstane logicky navázané na PDF Blacksmith; při opětovné aktivaci aplikace bude dosažitelné a před hlavním oknem. Pouze jedna instance dialogu Nastavení. Po uzavření obnovit zaměření původního ovládacího prvku, pokud ještě existuje, jinak bezpečný výchozí fokus.

- **OK / Použít / Zrušit:** zachovat schválenou transakční sémantiku Q12. Zavření `Esc`, křížkem v titulku či systémovým příkazem Zavřít je ekvivalent **Zrušit** (pokud probíhá vnořený dialog nebo editace zkratky, má nejprve přednost zrušení vnořené operace). `Použít` nepovoluje práci v hlavním okně: modalita trvá až do uzavření Nastavení.
- **Nedokončená editace PDF:** otevření Nastavení nesmí potichu potvrdit, ztratit ani přepsat rozepsaný text, výběr, původní stav kurzoru či neuložené změny. Pokud je v jiné části aplikace rozpracovaná modální operace, nové Nastavení neotevírat navrch bez řádného řešení jejího stavu.
- **Vykreslování a běžící úlohy:** již spuštěné asynchronní vykreslování či zpracování mohou běžet a aktualizovat stav, pokud je to bezpečné; nové uživatelské příkazy v hlavním okně blokovat. Nedopustit zamrznutí okna kvůli vnořené smyčce zpráv, čekání na render či při ukládání konfigurace.
- **Okno Nastavení:** přístupné klávesnicí (Tab, Shift+Tab, šipky v kategoriích), použitelná tlačítka OK/Použít/Zrušit a odpovídající fokus; textové pole pro mapování zkratek nesmí spouštět akce v dokumentu. `F1` z dialogu otevře příslušnou kontextovou nápovědu bez rozbití modality.

**Akceptace modality:** 1) při otevřeném dialogu nelze upravit dokument myší ani klávesnicí; 2) změna motivu se může projevit na hlavním okně a `Zrušit` ji vrátí; 3) po `Použít` je změna potvrzená, ale PDF stále nepřijímá vstupy; 4) `Esc`/křížek vrací nepotvrzené změny stejně jako `Zrušit`; 5) otevření/zavření zachová rozpracovaný text, výběr a aktivní kartu; 6) dialog se neztratí za hlavním oknem při `Alt+Tab` či více monitorech; 7) zaměření se po zavření korektně vrátí; 8) při zpracování na pozadí se aplikace nezasekne.


**Kategorie globálního Nastavení – SCHVÁLENO (Q14), jednotlivé volby zavádět postupně:**

| Kategorie | Zamýšlený rozsah (výčet funkcí je návrh, ne závazek vše ihned implementovat) | Počáteční priorita |
|---|---|---|
| **Obecné** | Jazyk rozhraní, motiv, chování po spuštění, obnova pracovní relace | P2 |
| **Rozhraní** | Velikost ikon, měřítko UI, chování panelů, kontextová nápověda | P2 |
| **Zobrazení PDF** | Výchozí zoom a režim čtení, cache; **schválené volby Q21:** vyhlazování textu a vektorů, zvýraznění tenkých čar (pouze optické), rychlé vykreslování během pohybu a kvalitní překreslení po zastavení | P1–P2 |
| **Klávesové zkratky** | Přehled a vlastní mapování příkazů, kontrola konfliktů, obnovení výchozího stavu | P1 |
| **Soubory** | Výchozí složky, dočasná data, bezpečné ukládání, obnova po pádu | P1–P2 |
| **Text a OCR** | Výchozí jazyk OCR, předvolby rozpoznávání, nastavení textového editoru | P2 |
| **Barvy a tisk** | ICC profily, výchozí barevný prostor, separace a předvolby exportu | P2 |
| **O aplikaci** | Skutečně spuštěná verze, identifikátor sestavení, datum a informace o vydání, licence a knihovny | P1 |

**Závazné pravidlo postupného zpřístupnění:** Architektura a návrh menu musí podporovat všech osm schválených kategorií. **V hotovém rozhraní ukazovat jen ty kategorie, v nichž je alespoň jedna skutečně funkční volba** (v kategorii O aplikaci alespoň reálné údaje); ostatní se zpřístupní až s dokončením implementace. Nezobrazovat prázdné stránky, klikatelné nefunkční přepínače ani předstírané funkce. Samotný souhlas s kategorií neznamená, že všechny položky z tabulky jsou schváleny k okamžité implementaci. Nepřidávat devátou hlavní kategorii bez dalšího schválení. Budoucí soukromí, diagnostiku a aktualizace lze přidat do příslušných kategorií nebo později znovu projednat.

**Akceptace Q14:** Při první implementaci musí být viditelné funkční kategorie „Klávesové zkratky“ a „O aplikaci“ s reálným obsahem; další se zpřístupňují po dokončení prvních funkčních možností. Při následném rozšíření konfigurace nesmějí být ztracena dříve uložená nastavení. Volba nové kategorie nesmí změnit aktivní PDF ani rušit čekající změny v jiných kategoriích.

**Nápověda – doporučený NÁVRH:** Kontextové vyvolání klávesou `F1`; lokální český návod rozdělený podle nástrojů; hledání v nápovědě; přehled skutečně nastavených zkratek; přechod na „Co je nového“ a „O aplikaci“, odkazy na dokumentaci a případně na nahlášení chyby. Uživatel může otevřít nápovědu bez změny aktuálního PDF. Diagnostické informace lze nabídnout ke zkopírování, ale nesmějí automaticky přenášet soubory nebo osobní data mimo počítač.

**Technické zásady:** Existující přímé mapování příkazů v `document_commands_windows.go` rozšířit na jednotný registr příkazů, jejich kontextů a konfigurovatelných vazeb; neměnit chování Win32 akcelerátorů pouze kosmeticky. Prověřit také `main_windows.go`, `workspace_windows.go`, `chrome_windows.go`, `text_windows.go`, `object_windows.go`, `pages_windows.go` a stávající nápovědné tooltipy. Budoucí import/export konfigurace (doporučeno JSON) umožnit až po odsouhlasení, protože musí správně řešit verze konfigurace a migrace. Pokud při buildu chybějí metadata vydání, doplnit je do procesu sestavení a při chybě nabídnout pravdivé „Údaj není dostupný“.

**Testy potvrzování – SCHVÁLENO:** Změnit motiv nebo ikony → Zrušit obnoví stav; změnit zkratku → OK ji aktivuje a po opětovném otevření je nastavena; změnit zkratku → Použít ji aktivuje, další změna → Zrušit vrátí stav **po Použít**; při změně kategorie čekající hodnoty zůstávají; při chybě zápisu nebo konfliktu se dialog nezavře a původní konfigurace zůstane použitelná; Esc/křížek mají odpovídající stornovací chování; samotné Nastavení nemění stav PDF.

**Testy:** Změnit `Ctrl+F` na volnou kombinaci → funguje nová, původní se neuplatňuje (kromě případně explicitně ponechané druhé vazby); restart aplikace zachová volbu; konflikt `Ctrl+Z` s jinou akcí je včas oznámen; `Ctrl+Z`/`Ctrl+Y` dál bezpečně spravují historii výběrů i úprav ve výchozím mapování; textové pole přijímá běžné znaky; české rozložení nezpůsobuje nechtěné příkazy. S aktivním editorem otevřít Nastavení/Nápovědu a po zavření pokračovat bez ztráty kurzoru nebo dokumentu. Informace o aktuálně nainstalované verzi musejí souhlasit s release metadaty buildu, i offline.

### UI-11 – Lokální a kontextová Nápověda (FEATURE/UX, P1–P2; ZPŮSOB OTEVŘENÍ SCHVÁLEN Q11, DALŠÍ FUNKCE NAVRŽENY)

**Cíl:** Uživatel se má rychle dostat buď k vysvětlení právě ovládaného prvku, nebo k plnému návodu. **Schváleno** je rozlišení ikona → úvod celé příručky a `F1` → kontextová kapitola. Ostatní možnosti níže jsou prozatím návrhy.

1. **Ikona Nápověda v horní liště – SCHVÁLENO:** vždy otevře **úvodní stránku celé příručky**, nikoli automaticky stránku aktivního nástroje. Samotná příručka nesmí měnit stav upravovaného PDF. Vyhledávání v ní je zatím doporučené rozšíření.
2. **Kontextová nápověda `F1` – SCHVÁLENO:** otevře kapitolu odpovídající zaměřenému ovládacímu prvku či konkrétní volbě v Nastavení; pokud takový kontext není dostupný, kapitolu aktivního nástroje. Pokud není rozpoznán ani nástroj, otevře úvod příručky. `F1` musí respektovat uživatelské přemapování zkratek; při standardním používání nevyžaduje předchozí výběr kapitoly. Drobné otazníky u specializovaných voleb jsou pouze návrh.
3. **Příručka použitelná offline:** vycházet z existujícího `NAVOD.md` a související dokumentace, avšak **nezobrazovat přímo interní nebo zastaralé vývojové pasáže**, pokud neodpovídají aktuální verzi UI. Obsah svázat s vydanou verzí; hyperlinky musí fungovat i v offline sestavení. Dlouhé kapitoly rozdělit podle nástrojů a běžných uživatelských úloh.
4. **Obsah a vyhledávání:** úvodní témata „Začínáme“, „Otevírání a ukládání“, „Výběr textu a objektů“, „Stránky“, „OCR“, „Barvy a separace“, „Export a tisk“, „Klávesové zkratky“ a „Řešení problémů“. Fulltext v Nápovědě má vyhledávat **v nápovědě**, nikoliv v otevřeném PDF (UI-05).
5. **Přehled skutečných zkratek:** vyčítat aktuální mapování ze stejného registru jako Nastavení; nepoužívat natvrdo vpisované zkratky, které po přemapování přestanou platit.
6. **Co je nového / O aplikaci:** odkazy na informace o nainstalované verzi v Nastavení. Dokumentace aktuální verze musí zůstat dostupná bez připojení. Volitelné online odkazy na repozitář a hlášení problémů jasně označit jako externí.
7. **Diagnostika:** doporučeno lokálně zkopírovat verzi, systémové parametry a popis použitých komponent; **bez automatického odesílání dat**, bez přidávání původního PDF a bez vypisování citlivých úplných cest bez vědomého souhlasu uživatele.

**Technická vazba:** `NAVOD.md`, `chrome_windows.go`, `workspace_windows.go`, registr příkazů ze `document_commands_windows.go`. Doplňovat průběžně společně se změnami funkcí a s vydáními; netvořit druhou nesynchronizovanou databázi zkratek.

**Závazná akceptace schválené části:** Kliknutí na ikonu `?` otevře **úvod celé příručky**, i když je aktivní například nástroj Text. Klávesa `F1` v nástroji Text otevře kapitolu pro Text; `F1` při zaměření volby ICC v Nastavení otevře odpovídající vysvětlení této volby; chybějící přesný kontext vede na kapitolu nástroje, popřípadě na úvod. Otevření/zavření nemění aktivní PDF, jeho výběr ani editaci. **Doporučené, dosud neschválené akceptace:** český návod offline, přehled přemapovaných zkratek, vyhledávání pouze v nápovědě, diagnostika bez automatického odesílání.

## 7. Separace, vykreslování a OCR

### RND-01 – Adaptivní ostré vykreslování vektorového PDF po vzoru Acrobatu (TECH/BUG, P0; SCHVÁLENO Q21)

**Schválená cílová koncepce:** Zachovat **PDFium** jako první volbu a vylepšit aktuální zobrazovací řetězec, místo automatické výměny rendereru. Cílovým vzorem je kvalitní, plynulé vizuální chování Adobe Acrobat, nikoli tvrzení, že známe či kopírujeme jeho neveřejnou implementaci.

1. **Při otevření / rychlém posunu či zoomu:** zobrazit rychle dostupnou, případně dočasně převzorkovanou cache, aby uživatel nemusel čekat na ostrý výstup. Následně za běhu požádat o přesné nové vykreslení **z původního PDF** v rozlišení odpovídajícím skutečnému zoomu a monitorovému DPI; zastaralé požadavky zrušit či ignorovat podle identifikátoru generace dokumentu.
2. **Vysoké přiblížení:** stránku nemusí celý renderer rastrově uchovávat na úrovni 800 %; preferovat **výřezy/dlaždice** viditelné oblasti s rozumným přesahem, paměťovým limitem a LRU či obdobným omezeným cache. Po posunu dočasně zobrazit existující data a nové dlaždice dopočítat; žádné trvalé zvětšování původní nízké bitmapy.
3. **Pan/zoom a kvalita:** při pohybu prioritizovat plynulost, po ustálení pohybu kvalitu. Jasně odlišit pracovní náhled od finálního vykresleného výsledku; žádné blikání, opakované přepočty stále stejného výřezu ani dlouhodobě rozmazané čáry.
4. **Schválené globální volby Nastavení → Zobrazení PDF:** (a) **Vyhlazovat text**, (b) **Vyhlazovat vektorové čáry**, (c) **Zvýraznit tenké čáry** (volitelný čistě optický efekt), (d) **Rychlé vykreslování při posouvání a zoomu**, (e) **Vysoce kvalitní překreslení po zastavení**. U voleb poskytovat vysvětlení, že ovlivňují jen obrazovku; při nepodporované volbě nezobrazovat falešně funkční přepínač.
5. **Přesnost:** zachovat správné mapování CSS/logických/zařízení pixelů, PDF souřadnic, DPI, otočení, CropBox/MediaBox, antialiasing a průhlednosti. Ostré mají zůstat vektorové kontury a fonty/obrysy písem; vložené bitmapy jsou i po zoomu stále bitmapami.
6. **Bez zásahu do PDF:** volby náhledu **nesmějí měnit** původní vektorové objekty, barvy, profily ICC, overprint, šířky čar, seřazení objektů ani obsah PDF při uložení, exportu a tisku. „Zvýraznit tenké čáry“ znamená pouze obrazovkové zvýraznění a musí jít vypnout pro kontrolu skutečné tloušťky. Náhled separací (RND-02) má vlastní barevně správný výpočet a nesmí se jím zkracovat cesta.
7. **Zjištěný stav – nejdříve diagnostika:** `reader_render.py` používá PDFium `FPDF_RenderPageBitmap` do bitmapy; `viewer_windows.go` již udržuje rastrové náhledy. To samo o sobě není chyba ani důkaz rasterizace uloženého PDF. Změřit přepočet po zoomu, cache, asynchronní invalidaci a limit RAM/VRAM. Přímé vykreslení přes `FPDF_RenderPage` na Windows HDC lze **porovnat** jako alternativu, ale neprovádět neověřený plošný přepis.

**Akceptační testy:** mikroskopické tenké čáry v technickém PDF, křivky písma, drobný text, velké stránky, přiblížení 100/200/400/800 % a více, prudké rychlé změny zoomu, otáčení a posouvání, 100/150/200 % DPI, duální režim a dokument s vysokým počtem objektů. Po zklidnění se musí objevit ostrá reprezentace z původního PDF; RAM nesmí nekontrolovatelně růst. Export, tisk a porovnání původních objektů mají prokázat nulovou změnu skutečné šířky čar a barvových dat způsobenou jen nastavením zobrazování.

### RND-02 – Náhled Separace a odečet pokrytí barev (BUG/UX/FEATURE, P1; Q28, Q29, Q30 SCHVÁLENO)

**Povinný zachovávací režim Q31:** Již existující nástroj **Separace** má funkce nad rámec zde nově projednávaného náhledu. **Nevytvářet nový zjednodušený panel, který by vytlačil nebo zrušil stávající nastavení a akce.** Před implementací dohledat a zdokumentovat všechny existující ovladače a backendové cesty, zejména převody barev aktuálního dokumentu (DeviceGray, K100 z kompozitní černé, explicitní CMYK pouze K), související nastavení ICC a výstupních záměrů, kontrolu tiskových desek, volby pro náhled a export. Řídit se `COLOR_MANAGEMENT.md`, zachovat současné bezpečné vytváření nové kopie při převodu a nezaměňovat **náhled** za **destruktivní převod**. U dosud neověřených současných funkcí nejprve provést inventuru a zkoušku, nikoli předpokládat, že chybí.


**SCHVÁLENO Q28 + Q29 – ovládání náhledu barevných separací a šedotónový výchozí pohled na jednu desku po vzoru profesionálního Output Preview (interakce, nikoli tvrzení o totožné technologii):**

1. **Seznam skutečných desek:** podle aktuálně zobrazovaného PDF a příslušné stránky/rozsahu zobrazit příslušné separace CMYK **Cyan, Magenta, Yellow, Black (K)** a pojmenované **přímé barvy (spot)**, pokud se opravdu v dokumentu vyskytují. Počet a názvy přímých barev načítat z PDF; žádný trvale fiktivní kanál „PANTONE“. Při změně stránky aktualizovat nabídku a transparentně zohlednit rozdílné sady separací; není-li spolehlivě dostupný přesný seznam, nelze předstírat jeho úplnost.
2. **Výchozí zobrazení:** všechny dostupné desky **zapnuté**; výsledný náhled ukazuje jejich společný tiskový souběh při zvoleném soft-proofingu. Každá deska má srozumitelný název, odpovídající identifikační barevný vzorek a vlastní přepínač zapnuto/vypnuto. Stav ovladače musí odpovídat tomu, co se právě vykresluje.
3. **Izolace a obnovení:** akce **„Jen tato“** u dané desky jedním kliknutím zapne jen její náhled a ostatní vypne. Akce **„Zobrazit všechny“** opět zapne veškeré aktuálně dostupné desky. Další individuální přepínání je možné v libovolné kombinaci; pokud uživatel vypne všechny desky, zobrazit prázdný tiskový podklad s jasným oznámením „Žádná separace není zapnuta“, nikoli chybu nebo předstíranou ztrátu dat.
4. **Přesná simulace:** změna viditelnosti desek musí vést k nové správné výpočetní kompozici přes CMYK i přímé barvy, respektující **přetisk (overprint), knockout, průhlednosti, skupiny, profily ICC, simulaci substrátu a výstupní záměr** v rozsahu podporovaném náhledovou cestou. Nestačí vypnout jeden RGB kanál z již hotové barevné bitmapy; takový postup by nevytvořil správnou simulaci tiskové desky. Při nepodporované funkci nebo pouze aproximovaném náhledu zobrazit varování a neprezentovat jej jako tiskově přesný.
5. **Kvalita při zoomu:** zlepšit detaily a text při náhledu separací. Pokud se zobrazuje výsledná tisková deska z rastrové simulace, při vyšším přiblížení **znovu přegenerovat viditelný výřez ve vyšším rozlišení**; nepřepočítávat prostým zvětšením staré bitmapy. Pokud lze bezpečně přímo vykreslovat PDF objekty bez ztráty simulace separací, zachovat ostré hrany. **Nepřeskakovat separační simulaci** a nepředstírat vektorový náhled tam, kde výpočet přetisku nebo průhlednosti potřebuje kompozici. Použít rozumné omezení paměti, asynchronní výpočty a zneplatnění zastaralých renderů dle RND-01.
6. **Nedesktruktivní pracovní náhled:** žádný přepínač desek, izolace ani obnova všech **nesmí editovat PDF nebo vytvářet krok neuložených obsahových změn**, přepisovat ICC/CMYK/K/spot data, měnit skutečné tiskové či exportní výstupy, anebo upravit volbu poznámek pro tisk/export. Náhled je samostatný od **nástroje pro skutečné úpravy separací**: převody, vypuštění kanálů a separační zásahy do PDF vyžadují explicitní jinou akci.
7. **Q29 – dvě podoby izolované separace (SCHVÁLENO, varianta A):** Je-li zapnutá **právě jedna deska**, nabídnout viditelný přepínač **„Odstíny šedi / Tisková barva“**; výchozí poloha je **Odstíny šedi**. Šedotónové vykreslení vyjadřuje hustotu/pokrytí **vybrané tiskové desky**, nikoli výsledný vizuální jas původní barevné stránky: **0 % barvy = bílý podklad, 100 % = černá, mezihodnoty = odpovídající šedá**. To platí pro Cyan, Magenta, Yellow, K i skutečné spot barvy; důležité zejména pro čitelnou kontrolu Yellow. **Tisková barva** vykreslí jednu desku v barvě odpovídající danému procesu či věrně simulované spot barvě podle správy barev/profilu, je-li k dispozici. Pokud přesnou spot simulaci nelze zaručit, přiznat aproximaci. Při zapnutí více desek zobrazovat jejich **barevný souběh** dle Q28; jednodeskový šedotónový režim nesmí být chybně použit na všechny desky dohromady. Přepínač, izolace i změny režimu jsou **výhradně zobrazovací**: nemění CMYK/spot data, ICC, output intent, overprint, uložený PDF, dirty stav, tisk/export ani skutečný nástroj převodu na stupně šedi. Při nulovém pokrytí zobrazovat barvu podkladu; pokud je podklad simulován jinak než bíle, musí být odlišen od ink coverage.


8. **Q30 – odečet pokrytí pod kurzorem (SCHVÁLENO):** Po najetí nad konkrétní bod PDF zobrazovat hodnoty jednotlivých **skutečně dostupných separací** v procentech (0–100 %) a **TAC = C + M + Y + K** (teoreticky 0–400 %). Přímé barvy uvádět **samostatně**, nikoliv je automaticky přičítat k TAC; případné zvláštní limity pro kombinaci přímých a procesních barev jsou jiný technologický problém. Odečet musí vycházet z výsledné separované reprezentace **po uplatnění relevantního ICC přepočtu, output intentu, přetisku, knockout a průhledností**, nikoliv z barev RGB monitoru nebo desaturovaného screenshotu. Při nedostupnosti věrného výpočtu zobrazit „Není dostupný přesný odečet“ / „Orientační výpočet“ s příčinou, nikoli zavádějící čísla. Zachovat kompatibilitu s doposud existujícími postupy měření či diagnostiky, pokud jsou v kódu.
9. **Q30 – nastavitelný limit TAC a vizuální kontrola:** Nabídnout pole pro **limit TAC**, jako počáteční návrhovou hodnotu **300 %**, ale bez vydávání 300 % za univerzálně správný limit pro všechny profily, papíry a tiskové procesy. Umožnit **zapnout/vypnout zvýraznění míst, kde TAC překročí nastavenou mez**, a při aktivaci zobrazit neutrální srozumitelnou legendu. Výsledné zvýraznění se generuje pouze do obrazovkové vrstvy náhledu a **nikdy nepřekrývá/neupravuje uložená separační data, tisk či export**. Po zoomování správně přepočítat přesnou polohu a vyhnout se optickým artefaktům; limit mimo 0–400 % odmítnout nebo srozumitelně validovat. Kontrola pod kurzorem musí fungovat i při vypnutí některé desky pro zobrazení: hodnoty mají výslovně reprezentovat **skutečné desky dokumentu**, nikoli zbylé zapnuté vizualizační filtry. Pokud uživatel chce analyzovat pouze jednu zapnutou desku, lze použít její samostatný odečet.
10. **Q31 – regrese původních funkcí:** Prokázat, že po doplnění Q28–Q30 stále fungují původní převodní režimy, ICC/profilové volby, stávající výstupní a tiskové operace, související varování/validace a interakce s ostatními kartami; nepřesměrovat tlačítko skutečného převodu na pouhé přepnutí náhledu. Funkční prototyp nové části není důvodem vypnout či nahradit dosavadní ovládací prvky.

**Akceptační testy Q28 + Q29:** PDF pouze v CMYK a PDF s CMYK + dvěma skutečnými přímými barvami; všechny desky po otevření zapnuté, každou lze vypnout a znovu zapnout; „Jen tato“ zobrazuje jedinou příslušnou desku a „Zobrazit všechny“ plný souběh; názvy spot odpovídají souboru. Opakovat přes více stran s odlišnými barvami, zoom 100–800 %, kombinace overprint/knockout/průhlednosti/ICC, všechny desky vypnuté, rychlé přepínání a duální zobrazení. Ověřit **žádné obsahové změny ani dirty stav**, žádné ovlivnění skutečného tisku/exportu. Barevné výstupy musí odpovídat ověřeným pravidlům z `COLOR_MANAGEMENT.md`. **Doplňkové testy Q29:** izolovat zvlášť C, M, Y, K a spot; na vzorníku s 0/25/50/75/100 % barvy ověřit šedotónové stupně včetně žluté, bez nahrazování náhledu RGB desaturací výsledné kompozice. Přepnout na „Tisková barva“ a zkontrolovat příslušný kanál a jeho ICC/spot simulaci. Vrátit všechny desky → barevný souběh; ověřit, že běžné uložení, tisk, export ani kontrolní re-extrakce skutečných separací se přepínáním nijak nezměnily.

**Akceptační testy Q30 + Q31:** Použít referenční CMYK vzorník s ověřenými hodnotami (např. C72 M45 Y30 K80 = TAC 227 %) a několik bodů s pokrytím pod i nad nastaveným limitem, včetně transparentního objektu, přetisku, ICC převodu, přesahu limitu a spot barvy. Ověřit odečet na 100/200/400/800 % a při rotaci; změnit limit 300 → 250 → 320 %, kontrolovat lokalizaci a zapnutí/vypnutí indikace. Vypnutí vizualizační desky nesmí potichu měnit skutečný odečet TAC z dokumentu. Ověřit, že export/uložení/vlastní separační úprava se z pouhého odečtu nespouští a nevzniká dirty stav. Pro **každou původní funkci** Separace vést seznam stávající stav → stav po změně a výsledky příslušného manuálního či automatického testu; včetně všech režimů v `COLOR_MANAGEMENT.md`, profilu, převodu do nové kopie, tisku a exportu. Pokud některý test nelze provést, uvést jej jako neověřený, ne jako splněný. 

### OCR-01 – OCR jako samostatná neztrátová vrstva (TECH, P0)

OCR nesmí vyžadovat převedení celého PDF do bitmapy. Textovou vrstvu lze extrahovat nativně z PDF, kde existuje. U vektorových křivek, bitmap či nečitelného obsahu OCR spouštět jen na dočasně vyrenderovaném výřezu do **mezipaměti**. Obrázek pro analýzu nemá být zapsán místo původních vektorů do výsledného PDF. Trvalou, volitelnou textovou vrstvu přidat pouze na výslovný příkaz OCR; kontrolovat orientaci, souřadnice, textovou extrakci a to, zda vrstva nezměnila vizuální vzhled. Při samotném výběru nebo prohlížení nic automaticky neserializovat do PDF.

### RND-03 – Audit existujícího nástroje Separace a minimální schválené grafické doplnění (Q32 SCHVÁLENO)

**Podklad:** statická kontrola výchozí větve repozitáře `Liborias/PDF-Blacksmith` dne 20. 11. 2026, zejména `viewer_windows.go`, `workspace_windows.go`, `proof.py`, `engine.py`, `NAVOD.md`, `COLOR_MANAGEMENT.md`, `tests/preview.py` a `tests/smoke.py`. Nebyla spuštěna sestavená aplikace ani integrační testy Ghostscriptu v systému Windows. **Přítomnost zdrojové implementace není důkaz, že funkce funguje bezchybně.** Před kódováním provést ověření aktuálního UI a porovnat s tímto auditem.

#### A. Co už v aplikaci existuje – POVINNĚ ZACHOVAT

| Stávající schopnost / ovládání | Důkaz v kódu | Pokyn pro Codex |
|---|---|---|
| Samostatný nástroj „Separace“, výběr náhledového CMYK/ICC profilu | `workspace_windows.go` (`idProofTab`), `viewer_windows.go` (`buildProofPanel`) | Zachovat profily i výběr „Podle PDF / výchozí CMYK“ a dosavadní zdroje profilů. |
| Tisková oblast: celý arch, crop, trim, bleed | `viewer_windows.go`, `proof.py` (`BOXES`, `box_options`) | Zachovat všechny čtyři volby; při redesignu možno přejmenovat jen se zachováním významu. |
| Informace o výstupním záměru PDF a použité simulaci | `viewer_windows.go` (`syncProofPanel`), `proof.py` (`output_intent`, `print_profile`) | Zobrazovat dál, přehledněji v horní sekci. |
| Rozpoznání procesních CMYK desek a přímých spot barev | `proof.py` (`render` / zařízení Ghostscript `tiffsep`, `page_proof`) | Převzít existující výstupy včetně reálných názvů spot barev. |
| Zapínání/vypínání desek jednotlivě, „Všechny desky“, „Vypnout všechny“ | `viewer_windows.go` (`idReaderPlateList`, `idReaderAll`, `idReaderNone`, `platesChanged`) | Zachovat i po doplnění akce „Jen tato“. Vypnutí desky mění jen náhled. |
| Skládání tiskového náhledu, včetně ICC a overprint | `proof.py` (`-dOverprint=/enable`, `composite_plates`, `proof_image`) | Nahradit pouze při prokazatelně věrnější alternativě; nezavádět prosté RGB maskování. |
| Zobrazení plošného podílu pixelů s nenulovou barvou a nejvyššího tónu desky | `proof.py` (`stat`), `viewer_windows.go` (`plateRow`) | Zachovat a správně **pojmenovat**; nelze vydávat za nový lokální odečet pod kurzorem. |
| Analýza aktivních desek v celém PDF nebo na vybraných stránkách | `viewer_windows.go` (`idReaderScan`, `idReaderScanSelection`), `proof.py` (`scan_plates`) | Zachovat obě akce a přehled výsledků s čísly stránek. |
| Převody barev aktuálního dokumentu: DeviceGray, složená černá → K100, explicitní CMYK pouze K; Gray Gamma 1.8/2.2, vlastní Gray ICC a cílový CMYK ICC | `viewer_windows.go` (`buildProofPanel`), `workspace_windows.go` (`idProofApply`), `COLOR_MANAGEMENT.md` | **Nedotýkat se rozsahu funkcí a bezpečné cesty převodu do nové kopie**. Nový design je musí zpřehlednit, nikoli skrýt bez dostupného otevření. |
| Základní integrační test přepínání desek a simulace spot | `tests/preview.py` (`check_plate_switching`, `check_proofs`) | Rozšířit testy, nikdy je při redesignu neodstraňovat. |

#### B. Technická zjištění a již schválená funkční rozšíření (grafický redesign mimo Q32 NENÍ schválen)

1. **Současné rozložení ponechat:** `buildProofPanel` používá pevné souřadnice ovladačů v jednom pravém panelu. Přestože rozložení není ideální, uživatel pro tuto etapu **výslovně vybral jen barevné značky desek**; přidání nadpisů, sbalitelných skupin, dynamických výšek, přesunů ovladačů nebo další vizuální změny jsou **neschválené**. Schválené funkce Q28–Q30 doplnit šetrně a bez přestavby zbytku panelu.
2. **„Jen tato“ chybí jako přímý ovladač u řádku desky:** backend v `proof.py` umí omezit výběr na vybrané desky (`plates`), ale současný viditelný panel v `viewer_windows.go` nabízí jednotlivé přepínače a hromadné „Všechny desky“/„Vypnout všechny“. Přidat přímou izolaci jedním kliknutím bez mazání dosavadních akcí (Q28).
3. **Šedotónové zobrazení jediné desky:** dosavadní `render_proof_selection` pro izolovanou desku standardně sestavuje CMYK a ICC RGB náhled, nikoli speciální šedotónovou kontrolu 0 % → bílá / 100 % → černá. Doplnit přepínač „Odstíny šedi / Tisková barva“, zobrazovaný relevantně při jediné aktivní desce (Q29).
4. **Současné procento není lokální pokrytí ani TAC:** `proof.py` v `stat()` počítá `pixels = size - histogram[255]`, tj. **podíl pixelů, kde se barva vůbec vyskytuje**, a samostatně `maximum` (nejvyšší tón). **Nejde o průměrnou spotřebu barvy, lokální procenta C/M/Y/K pod kurzorem, ani součet TAC**. Při přidání Q30 nesmí Codex přejmenovat existující údaj „plocha“ na „pokrytí pod kurzorem“ ani z něj dopočítávat TAC; musí číst skutečné hodnoty z rastrů separací příslušného místa při známém výstupním profilu, po zarovnání stránky, rotace a ořezu. Viditelnost desek v náhledu nesmí měnit analytické hodnoty původních desek.
5. **Přesnost detailu náhledu:** aktuální cesta `page_proof` používá pro kompletní stránku omezení `render_dpi(..., 192, 1850, ...)`; běžná cestovní mezipaměť `PROOF_CACHE_LIMIT = 24 << 20`. Při vysokém zoomu a malých detailech může být nutné znovu vzorkovat samotné separace ve viditelném výřezu. Vyhnout se prostému zvětšování náhledu a nevyřadit stávající cache (Q21/Q28).
6. **Bezpečné rozlišení náhledu a převodu:** náhled beze změny PDF a skutečný převod do nové kopie jsou dvě odlišné operace. Zachovat jejich dosavadní oddělení, texty i chování; další přesuny a grafické oddělování sekcí nejsou součástí Q32.

#### C. Schválená grafická změna Q32 – pouze identifikační značky separací

**Rozsah schválení je záměrně omezený.** Uživatel po prohlédnutí reálných screenshotů panelu Separace z návrhů grafického redesignu zvolil **jen doplnění barevných značek u tiskových desek**. Nic dalšího z vizuálních návrhů v3.6 a v3.7 nebylo schváleno. Nevykládat rozhodnutí jako souhlas s přeuspořádáním panelu.

Implementace:

1. Do **stávajícího řádku** desky mezi checkbox a textový název vložit drobný barevný čtverec nebo kruhový vzorek, orientačně **12 × 12 logických px**, přizpůsobený DPI. Řádek má dál ukazovat současné hodnoty **`Plocha / max. tón`**, název desky i stav zaškrtnutí. Ovládací prvky neroztahovat nepřiměřeně a neměnit jejich funkci.
2. **CMYK:** C = cyan, M = magenta, Y = yellow, K = black. Pro K použít viditelný světlý okraj, pro žlutou kontrastní obrys na tmavém panelu. Barvy slouží k orientaci v seznamu, **nejde o výstupní barevně řízený tiskový důkaz**.
3. **Přímé barvy (spot):** pokud je k dispozici jejich simulovaná barva, ukázat informativní vzorek, ale ponechat originální název separace. Není-li věrná simulace dostupná, použít neutrální značku a případně tooltip „Náhled barvy není dostupný“; **nevymýšlet RGB hodnotu ani název PANTONE**.
4. **Přístupnost:** zachovat checkbox, název a číselné údaje; význam ani stav desky nesmí být závislý pouze na barvě. Barevné značky nesmějí změnit data separací, ICC, overprint, tisk, export ani pořadí klávesového ovládání.
5. **Žádné jiné grafické zásahy:** nepřidávat vnitřní záložky `Náhled a kontrola` / `Převod barev`, nepřestavovat pravý panel na šest sekcí, nesbalovat původní bloky, nepřesouvat tlačítka a nenahrazovat aktuální styly, pokud to není nezbytné pouze k umístění značky.

**Důležité oddělení funkčnosti a vzhledu:** Q28 `Jen tato`, Q29 šedotónové zobrazení izolované desky a Q30 měření pod kurzorem + TAC zůstávají samostatně schválenými **funkčními požadavky**. Jejich realizace může nezbytně přidat ovladače, ale není oprávněním k celkovému redesignu současného panelu. Každou změnu implementovat s minimálním dopadem a regresními testy Q31.

**Akceptace Q32:** Zobrazit PDF s C/M/Y/K a alespoň jednou spot separací, ověřit správné malé vzorky vedle názvů, čitelnost žluté i černé na tmavém podkladu, zachování `Plocha / max. tón`, funkčnost zaškrtávání a hromadných tlačítek, všechny dostupné profily a převody, HiDPI a klávesnici. Při spot barvě bez dostupného vzorku ověřit neutrální fallback. Před a po porovnat skutečné PDF a náhledovou logiku: pouze vizuální značka, bez zásahu do dokumentu.


## 7A. Poznámky, komentáře a grafické anotace (NOVÝ POŽADAVEK Q22)

### ANNO-01 – Samostatný nástroj Poznámky a standardní PDF anotace (FEATURE, P1; ZÁKLADNÍ PALETA SCHVÁLENA Q25)

**Výslovný požadavek uživatele:** Přidat nástroj **Poznámky**, který dovolí přímo nad otevřeným PDF ukládat **textové poznámky po způsobu Adobe Acrobat** a **grafické poznámky, značky a volné kresby**. Poznámky musí být **logicky i technicky oddělené od původního obsahu stránky** a zůstat upravitelné/odstranitelné bez destruktivního zásahu do původní kresby.

**Doporučený interoperabilní formát:** Standardní **PDF anotace** uvedené v `/Annots` jednotlivých stránek; vzhled reprezentovat podle typu anotace pomocí standardních *appearance streams* `/AP`. To je něco jiného než volitelná grafická vrstva (*Optional Content Group*, OCG) i něco jiného než obsahový stream stránky `/Contents`. **Anotace se může v UI chovat jako oddělená vrstva „Poznámky“, ale technicky to nemusí a standardně nemá být OCG.** Po běžném uložení do PDF musí být otevřitelná a pokud možno editovatelná v Adobe Acrobat i dalších kompatibilních editorech; nezávislý proprietární sidecar nesmí být jediným úložištěm poznámek.

**Závazný rozsah první funkční verze – SCHVÁLENO Q25:**

- **Komentář:** připínací textová poznámka s ikonou a otevřením komentáře (PDF `/Text`).
- **Textové pole:** přímo na stránce editovatelné textové pole (PDF `/FreeText`) s podporou české diakritiky.
- **Volná kresba:** kreslení perem jako anotace `/Ink`, bez zásahu do původních vektorových objektů stránky.
- **Čára / šipka:** čárové a šipkové anotace `/Line` s vhodným zakončením.
- **Obdélník / elipsa:** geometrické anotace `/Square` a `/Circle`, případně s výplní podle možností formátu a zvolených vlastností.
- **Zvýrazňování a textové značky:** zvýrazňovač `/Highlight`, **podtržení** `/Underline` a **přeškrtnutí** `/StrikeOut`. Při práci s existujícím textem používat odpovídající čtyřúhelníky (quadpoints) i přes více řádků, nikoli pouhý bounding box.
- **Vlastnosti anotací:** tam, kde jsou pro daný typ smysluplné, nastavit barvu, tloušťku tahu, výplň, průhlednost, velikost/styl písma textového pole, rozměry a umístění. V UI zobrazovat pouze vlastnosti použitelné pro vybraný typ. Nenabízet nefunkční ovladače.
- **Panel Poznámky:** samostatný panel se seznamem anotací uspořádaných podle stránek, stručným náhledem obsahu a typem; kliknutí posune dokument na odpovídající stránku a zvýrazní příslušnou poznámku. U podporovaných typů umožnit vybrat, upravit a odstranit poznámku. Zobrazit autora a čas pouze pokud jsou údaje skutečně k dispozici. Sbalování panelu odpovídá UI-04.
- **Umístění nástrojů:** šest základních vizuálních skupin / ikon (Komentář, Textové pole, Volná kresba, Čára/šipka, Obdélník/elipsa, Zvýraznění); podtržení a přeškrtnutí jsou přístupné jako související volby textových značek. Dodržet sjednocený design ikon, tooltip a vysvětlující popis podle UI-07.

**Odložené doplňky (BACKLOG, nejsou požadavkem první funkční verze):** odpovědi ve vláknech, razítka, uživatelské symboly, textové bubliny s odkazovací čarou (*callout*), polygon, lomená čára, vyhledávání a pokročilé filtrování seznamu. Je možné je později rozšířit, nikoli nahrazovat schválenou základní paletu.

**Akceptace Q25:** Vytvořit, znovu otevřít, upravit a smazat každý z podporovaných typů; ověřit uložení ve standardní PDF podobě a čitelnost v Adobe Acrobat, přesné umístění při zoomu/rotaci, českou diakritiku ve `/FreeText`, změnu příslušných vlastností, navigaci ze seznamu podle stránek, historii `Ctrl+Z` / `Ctrl+Y` a obnovu přes AutoRecover. Čistý výchozí tisk i export nezahrnují pracovní anotace (Q23, Q24); pracovní Uložit je zachovává. Pokud konkrétní PDF knihovna typ nepodporuje, prototypovat interoperabilní implementaci a limit otevřeně zdokumentovat; neschovávat chybu rasterizací.

### ANNO-02 – Viditelnost a oddělení od editace dokumentu (UX/TECH, P1; SCHVÁLENO Q26 + Q27)

1. **Zobrazení – SCHVÁLENO Q26 (varianta A):** **Při otevření PDF jsou existující poznámky ve výchozím stavu viditelné.** V pracovním rozhraní je přímo dostupný jediný srozumitelný přepínač **„Zobrazit poznámky“** (zapnuto = viditelné, vypnuto = skryté), dostupný i bez otevření celého panelu Poznámky. Přepínač ovlivňuje **pouze obrazovkové zobrazení a hit-test anotací v pracovní ploše**, nikoli jejich přítomnost nebo obsah v PDF, nevyvolává ukládání ani neoznačuje dokument jako změněný. Skryté poznámky nesmějí překážet výběru podkladového textu a vektorových objektů. V panelu Poznámky lze i při skrytí zobrazit seznam a uživateli jasně sdělit, že poznámky jsou nyní na stránce skryté; z navigace na konkrétní poznámku lze bezpečně nabídnout jejich zviditelnění. Po otevření dokumentu bez anotací je přepínač rovněž jednoznačný, ale žádný grafický komentář se nevykreslí. **Rozsah přepínače – SCHVÁLENO Q27 (varianta A):** Stav „Zobrazit poznámky“ se udržuje nezávisle **pro každý otevřený PDF dokument**, nikoliv globálně pro celou aplikaci. Při přepnutí aktivní karty musí přepínač vždy zobrazovat stav daného dokumentu; skrytí poznámek v `A.pdf` nemění viditelnost poznámek v `B.pdf`. V duálním zobrazení mají levý a pravý dokument každý svůj stav a přepínač se vztahuje k právě aktivnímu dokumentu (zvolený panel/jeho karta); žádná neúmyslná synchronizace. Pokud stejný dokument figurálně existuje ve více pohledech, jde o tentýž dokumentový stav, ne o dvě nezávislé kopie. Stav patří do dočasné pracovní relace / modelu karty, nesmí měnit soubor PDF ani jeho příznak neuložených změn. Při novém otevření souboru bez obnoveného pracovního stavu platí schválené výchozí **viditelné (Q26)**. Při `Uložit upravenou kopii…` a nahrazení dokumentu ve stejné kartě zachovat aktuální nastavení viditelnosti pro plynulost práce; nejde o zápis do PDF.
2. **Výběr:** aktivní nástroj Poznámky upřednostní výběr/anotování poznámek levým tlačítkem; v režimu hlavních Objektů/Textu nesmí poznámky nečekaně blokovat zásahy do originální kresby. Možnost cíleně vybrat existující poznámku, upravit, přemístit, změnit vlastnosti a smazat. **Stisk kolečka + tažení** nadále slouží ke schválenému obdélníkovému výběru; v nástroji Poznámky se bez dalšího rozhodnutí nepřemapovává.
3. **Pořadí a geometrie:** anotace mají vlastní hit-test s podporou zoomu, DPI, otočení stránek, ořezu a průhlednosti. Zabránit falešnému zasažení poznámky jen kvůli velkému průhlednému bounding boxu. Prioritizace při překryvu má být konzistentní s nástrojem Objekty, ale oddělená podle režimu.
4. **Věrnost:** základní PDF obsah ponechat beze změny. Barvy, vrstvy tisku, separace a poznámky neslučovat automaticky; případné **Sloučit poznámky s obsahem (flatten)** může být až **samostatná výslovná destruktivní akce s varováním a doporučenou kopií**, nikoli vedlejší účinek uložení či exportu.

### ANNO-03 – Ukládání, kompatibilita a ochrana práce (TECH, P0–P1; DETAILY NÁVRH)

1. **Přenositelný PDF výstup:** anotace a jejich vizuální reprezentace ukládat tak, aby byly rozpoznatelné a co nejvěrnější při opětovném otevření v PDF Blacksmith a Adobe Acrobat; neztrácet již existující anotace, odkazy či formulářové widgety. Pokud konkrétní knihovna nepodporuje vytvoření či editaci subtype, nejdříve doložit omezení a předložit kompatibilní řešení, nikoli potichu převést anotaci na bitmapu.
2. **Uložit / Uložit upravenou kopii:** přidání/editace/smazání poznámky nastaví stav „neuložené změny“, zapíše se pouze při vědomém ukládání PDF a respektuje Q18 (po uložení kopie uzavřít originál a používat kopii ve stejné kartě).
3. **Historie a obnova:** každá logická akce s poznámkami je součástí chronologického Undo/Redo dané karty (`Ctrl+Z`/`Ctrl+Y`); při rozepsaném komentáři neztratit text. **AutoRecover Q19** musí zahrnovat poznámky i rozpracované komentáře, po vědomém Neuložit je zahodit a po havárii obnovit.
4. **Existující anotace:** umět načíst stávající anotace z PDF, bezpečně zobrazit a zachovat i typy, které v první implementační etapě není možné editovat. Neodstraňovat neznámé nebo nepodporované typy při uložení; ověřit interoperabilitu a zachování původních PDF vlastností.
5. **Ochrana dokumentů:** respektovat oprávnění PDF, zámky anotací, podpisy a PDF/A nebo další konformní režimy; změny mohou zneplatnit digitální podpis nebo narušit soulad s PDF/A. Upozornit předem, umožnit uložit kopii a nikdy neslibovat nezměněný podpis. Neodemykat chráněné dokumenty skrytě.
6. **Ověření knihovny:** prohlédnout existující Go/Win32 UI a Python PDF backend, schopnosti konkrétní verze PDFium API `fpdf_annot.h` a save API pro jednotlivé typy, včetně vytvoření správného `/AP` a práce s Unicode českým textem. Technický prototyp s uložením/opětovným otevřením v Acrobatu před masovou implementací.

### ANNO-04 – Tisk, export a separace (UX/FEATURE, P1; TISK SCHVÁLEN Q23 / EXPORT SCHVÁLEN Q24)

- **Na obrazovce – SCHVÁLENO Q26 (varianta A):** poznámky jsou po otevření PDF **standardně viditelné** a lze je ihned skrýt/zobrazit přepínačem **„Zobrazit poznámky“**. Přepínání je čistě prezentační stav, neovlivňuje pracovní PDF, uložené anotace, tiskový náhled ani čistý export; jejich režimy se řídí výhradně vlastními explicitními přepínači Q23 a Q24.
- **Při tisku – SCHVÁLENO, varianta A (Q23):** V tiskovém dialogu nabídnout přepínač **„Tisknout včetně poznámek“**, jehož **výchozí stav je vypnuto**. Bez zaškrtnutí se tiskne pouze základní obsah dokumentu, bez textových, zvýrazňovacích a grafických anotací. Po zaškrtnutí se zahrnou tisknutelné anotace podle odpovídajících pravidel PDF (mj. příznak `/Print`); netisknutelné a skryté typy správně respektovat, případná omezení uvést. **Tiskový náhled musí přesně odpovídat aktuálnímu nastavení přepínače.** Volba je výhradně vlastnost tiskové úlohy: nevymaže anotace z uloženého PDF, nemění přepínač Zobrazit/Skrýt poznámky na obrazovce a neprovádí flatten ani nevynucuje rasterizaci či změnu barev a separací původního obsahu. Při otevírání další tiskové úlohy opět výchozí vypnuto, ledaže bude později výslovně schválena uživatelská preference.
- **Při exportu PDF – SCHVÁLENO, varianta B (Q24):** běžný export slouží zejména předtiskové přípravě a **ve výchozím stavu nezahrne textové komentáře, textové rámečky, kresby, grafické značky ani textová zvýraznění vytvořená jako poznámky**. Exportní dialog musí obsahovat srozumitelný přepínač **„Zahrnout poznámky (editovatelné)“**, výchozí **vypnuto**; pokud jej uživatel zapne pro konzultační/revizní dokument, zachovat anotace jako standardní editovatelné PDF objekty, nikoli je automaticky vypálit do stránky. Výstupní PDF smí vzniknout jen jako **nový exportní soubor**; originál i aktivní pracovní kopie zůstávají otevřené a obsahují své poznámky.
- **Rozlišovat export od ukládání:** `Uložit` (`Ctrl+S`) a `Uložit upravenou kopii…` podle DOC-03 **vždy zachovávají poznámky**; samotná výchozí volba čistého *exportu* nesmí odstranit komentáře z pracovního PDF nebo historie ani z AutoRecover. Export nezmění příznak neuložených změn zdrojového dokumentu. Rozdíl musí být jasně popsaný přímo v UI a v Nápovědě.
- **Selektivní odebrání poznámek:** nezaměňovat odstranění komentářových a grafických anotací za bezhlavé vymazání celé struktury `/Annots`. Mohou zde být také **odkazy `/Link`, interaktivní formulářové widgety `/Widget` a prvky podpisů**, které nejsou uživatelskými poznámkami a mají se zachovat, pokud je jiná výslovná předtisková operace nemění. U případných PDF/X, PDF/A a podepsaných PDF před exportem ověřit kompatibilitu výstupu / soulad s cílovým standardem; nepředstírat zachování platnosti podpisu, pokud změna souboru podpis ovlivní.
- **Zploštění pouze výslovně:** případné „Sloučit poznámky s obsahem stránky“ (*flatten*) může být jen samostatná, potvrzená volba/operace pro výstupní kopii, **není součástí výchozího exportu ani zahrnutí editovatelných poznámek**. Při této akci výslovně varovat před ztrátou editovatelnosti a možnými dopady na barvy, separace, průhlednost a přetisky. Dokud nebude tato funkce implementována a otestována, nezobrazovat nefunkční ovladač.
- **Náhled Separace:** jasně rozlišit tiskové barvové desky samotného dokumentu a případně přidané anotace. **Čistý výchozí export Q24** i výchozí tisk Q23 počítají pouze separace skutečného tiskového obsahu; případné zahrnutí anotací vyžaduje explicitní volbu pro konkrétní operaci. Přepínání poznámek na obrazovce nesmí samo změnit výsledné separace. Respektovat COLOR_MANAGEMENT.md a nepřevádět originální CMYK/K objekty jen kvůli anotacím.
- **Akceptační test tisku (Q23):** PDF s podkladovou vektorovou kresbou, připínací poznámkou, textovým polem, zvýrazněním a grafickou kresbou → otevřít tiskový dialog → přepínač vypnutý, náhled i výsledek bez anotací; zapnout „Tisknout včetně poznámek“ → náhled i výsledek obsahují příslušné tisknutelné anotace; po ukončení tisku otevřené PDF i uložený soubor stále obsahují všechny poznámky beze změny. Totéž otestovat při obrazovkově skrytých poznámkách, u cizího PDF s netisknutelnými anotacemi a při zapnutých separacích.
- **Akceptační test exportu (Q24):** načíst PDF s původním vektorovým CMYK/K obsahem, anotacemi `/Text`, `/FreeText`, `/Ink`, `/Highlight`, odkazem `/Link` a formulářovým widgetem `/Widget` → běžný export bez zásahu uživatele: nový soubor bez poznámek, ale s nezměněnými potřebnými odkazy/widgety a bez zploštění/rasterizace původní stránky; zdrojové PDF nadále obsahuje všechny anotace. Zapnout „Zahrnout poznámky (editovatelné)“ → nově exportovaný soubor obsahuje standardní anotace editovatelné v Acrobatu. Pak `Ctrl+S` i `Uložit upravenou kopii…` → anotace se zachovávají. Otestovat také separace/overprint, českou diakritiku, podepsané a chráněné dokumenty, rozdílnou obrazovkovou viditelnost poznámek a chybu exportu (zdroj i AutoRecover zůstanou nedotčené).

**Návrh akceptačních testů:** původní PDF s vektory a českým textem → vložit textovou poznámku, viditelný textový rámeček a ink kresbu → uložit jako kopii → zkontrolovat původní soubor bitově nezměněný a nový soubor s editovatelnými anotacemi v Acrobatu; znovu otevřít v Blacksmith; zapnout/vypnout viditelnost; zkusit přesun, smazání a Undo/Redo; pád během rozpracovaného komentáře a obnova AutoRecover; OCR, otáčení, zoom, změny pořadí stran, tisk s/bez poznámek a zachování původních anotací cizího autora. Zvlášť otestovat podepsaný, chráněný a PDF/A dokument a zobrazení české diakritiky.

## 8. Identita aplikace a načítání

### BRAND-01 – Ikona aplikace (UX, P2; ROZHODNOUT)

Aktuální značku zjednodušit na **vínově červený plochý piktogram kovadliny s označením PDF**, bez slovního názvu Blacksmith (ani historického chybného „Blaksmith“). Zachovat proporce a identitu stávající kovadliny, zvětšit její optickou váhu oproti nynější ikoně; nepoužít realistický render, odlesky ani stíny. Připravit škálovatelný zdroj a z něj více rozlišení ikony pro Windows. Původní renderované obrázky v `assets/` neodstraňovat.

### BRAND-02 – Načítací obrazovka (FEATURE, P2)

Nad průběhem načítání PDF zobrazit stávající renderovanou kovadlinu z assetů; pod ní název aplikace. Grafický celek orientačně odpovídá asi **10 % výšky okna** (nutno ověřit vizuálním náhledem na menší obrazovce). Kolem kovadliny roste po obvodu kružnice vínově červený ukazatel průběhu. Stav kroužku musí odpovídat **reálnému progresu** (načtení souboru, popis stran, náhledy), a pokud procenta backend nezná, ukázat indeterminate animaci namísto fingovaných procent. Zajistit, aby dlouhé načítání nezablokovalo UI a bylo možné ho zrušit.

**SCHVÁLENO:** Oficiální text pod kovadlinou je **„PDF Blacksmith“**. Nepoužívat chybné „PDF Blaksmith“. Samotný piktogram v ikoně však obsahuje pouze kovadlinu a značku PDF, nikoliv celý slovní název.

### BRAND-03 – Oficiální název aplikace (UX/TECH, P1; SCHVÁLENO)

**Jediný oficiální název: `PDF Blacksmith`.** Varianta `PDF Blaksmith` je překlep, nikoliv alternativní značka.

- Sjednotit název viditelný uživateli v titulku okna, obrazovce načítání, systémových dialozích, informacích „O aplikaci“, tooltipu ikony aplikace, dokumentaci, nabídkách a metadatech sestavení.
- Výsledný distribuční soubor pojmenovat **`PDFBlacksmith.exe`**, nikoli historicky chybně `PDFBlaksmith.exe`. Opravit i případné odkazy v README, dokumentaci, sestavovacích skriptech, manifestech, instalačních či release souborech. Názvy měnit **pouze tam, kde opravdu obsahují překlep**; správně pojmenovaný repozitář `Liborias/PDF-Blacksmith` neměnit.
- Vnitřní technické identifikátory (například třída Windows okna `PDFBlaksmithWindow`, názvy assetů či hodnoty konfigurace) migrovat bezpečně a konzistentně; před přejmenováním ověřit všechny závislosti. Zachovat funkčnost starších uživatelských nastavení, je-li na těchto názvech závislá.
- Historický soubor `assets/blaksmith.ico` při příslušné změně přejmenovat či nahradit novým `assets/blacksmith.ico`, ale nikdy kvůli přejmenování nemazat renderované `assets/anvil.png` a `assets/blacksmith.png`. Piktogram z BRAND-01 může být vytvořen jako nový zdroj až v jeho etapě.
- **Akceptace:** uživatelské rozhraní ani nově vydané artefakty neobsahují chybné `Blaksmith`; funguje spuštění, načítání ikony a sestavení Windows aplikace. Výjimkou mohou být zdokumentované historické názvy v migračním kódu, které se už nezobrazují uživateli.

## 9. Testovací scénáře a akceptace napříč aplikací

- **PDF s vektory:** ostré kontury při vysokém zoomu, žádná bitmapa místo křivek v uloženém PDF; test rotovaného objektu a tlustých/tenkých tahů.
- **PDF se smíšeným formátováním:** odstavec s tučným slovem a rozdílnou barvou, rozdělený do několika PDF textových runů; všechny výběry probíhají napříč runy, při editaci se nezměněné styly zachovají.
- **Vícesloupcová sazba, tabulka:** čísla a popisy oddělené více mezerami/tabulátory; levý dvojklik nevybere sousední sloupec, nadpis mimo sloupec nevytvoří falešné propojení.
- **PDF se subset fontem a chybějícím znakem:** editor nabídne konkrétní výstrahu/náhradní postup, nevloží chybné znaky bez upozornění.
- **PDF s textem na křivkách:** zůstane vektorové; OCR jen dočasná analýza nebo explicitně přidaná textová vrstva.
- **Objekty v překryvu:** prázdný bounding box velké cesty, viditelná čára, průhledný PNG a vyplněný tvar; výběr se řídí skutečným zásahem a kandidáty lze cyklovat. Ověřit toleranci 5 px, klik v transparentním pixelu a nejméně tři překrývající se skutečné objekty; stejné chování při zvětšení a otočení PDF.
- **Globální gesta:** **prostřední tažení** označuje obdélníkem v každém vhodném nástroji; samostatný prostřední klik nevybírá; **levé klikání** vybírá objekty a miniatury, opakovaný levý dvojklik vybírá **slovo → segment → věta → odstavec**, `Ctrl + klik` v textu vybírá větu a `Ctrl + dvojklik` vizuální řádek; levé tažení provádí odpovídající přímou akci.
- **Kreslení primitiv:** levým tažením obdélník/elipsa; `Shift` přesný čtverec/kruh; `Ctrl` geometrie od středu; `Ctrl+Shift` přesný tvar od středu. Ověřit všechny čtyři varianty pro oba nástroje, přepínání modifikátorů v průběhu tahu, Esc a Undo/Redo.
- **Laso – dva schválené režimy (Q16):** výchozí **A = protnutí viditelné kresby** (i když střed zůstane mimo polygon), volitelný **B = celá viditelná kresba objektu uvnitř**. Testovat čáry, výplně, bitmapy s alfa průhledností, konkávní polygon a hraniční průniky; uzavření **levým** dvojklikem, kliknutím blízko začátku nebo Enter, zrušení Esc; prostřední tažení vždy obdélník.
- **Historie:** přesun 2–3 vybraných stran, potvrzení, Undo, Redo, uložení a znovuotevření; **ověřit obsah dialogu pro souvislé i nesouvislé výběry, zrušení přesunu a no-op bez zápisu historie**; krokové Undo/Redo textového výběru v jednotné chronologické historii (`Ctrl+Z` / `Ctrl+Y`), včetně prokládání obsahovými úpravami a kontroly příznaku změněného PDF.
- **Více dokumentů:** duální zobrazení, identické názvy v různých adresářích, přenos stran, nezávislý stav panelů a ukládání.
- **Zavírání dokumentů (DOC-05; schváleno Q20):** ověřit jednoznačnou volbu levý/pravý/vše, správné zachování ostatních karet, přepočet duálního zobrazení a zejména souhrnnou ochranu neuložených změn. Při Neuložit odstranit pouze recovery daného dokumentu; při chybě ukládání či Esc nic neztratit. Testy zahrnou zavření karty, ukončení aplikace a asynchronní ukládání.
- **Windows UX:** otevřená nabídka Soubor a přejezd do ní; pravý okraj/rohy okna, DPI 100/150/200 %, minimalizace, maximalizace, klávesnice.
- **Panely a ouška:** při sbalení hlavního/pomocného panelu trvalá úzká svislá lišta; hlavní ouško pouze zaoblený čtvereček se šipkou, pomocná ouška s čitelným názvem. Otevřené pomocné ouško splývá barvou a obrysem s panelem, zavřená mají jiný odstín. Otestovat několik současně sbalených panelů, pravou/levou stranu, konflikt s náhledy, klávesnici a DPI.
- **Ikony a nápověda:** aktivní Obdélník → najetí na Elipsa ukáže její delší popis, aktivní režim se nemění → odjetí obnoví Obdélník; najetí na další ikonu bez blikání, skutečná aktivace až kliknutím, návrat po odchodu klávesnicového fokusu, disabled stav, vysoké DPI, úzký panel.
- **Hybridní drag-and-drop:** otestovat vizuální náhled a menu **po uvolnění** pro interní PDF, známou složku a plochu; zejména pro **CorelDRAW, Adobe Photoshop, Affinity Designer/Photo** a dále Word/LibreOffice a textový editor zvlášť ověřit reálné vložení (ne pouze animaci) a podporu jednotlivých formátů. **Při neúspěchu povinně nabídnout „Uložit jako…“ na jiné uživatelem zvolené místo.** Pokud se vloží pouze PNG namísto zachování PDF vektoru nebo se ztratí možnost editace textu, explicitně uvést degradaci. Zkontrolovat `%TEMP%`, práci se schránkou, práva, kolize názvů, Unicode, více monitorů a storno. Zkontrolovat `%TEMP%`, práci se schránkou, práva, kolize názvů, Unicode, více monitorů a storno.

**Náhled separací a pokrytí (Q28 + Q29 + Q30, SCHVÁLENO; Q31 zachování funkcí):** Po otevření zobrazit všechny reálné tiskové desky CMYK/spot zapnuté, jednotlivé vypínat, izolovat přes „Jen tato“ a obnovit „Zobrazit všechny“. **Při právě jedné aktivní desce** výchozí odstíny šedi pro kontrolu pokrytí, přepínač na zobrazení v tiskové barvě; 0 % barvy znamená bílou, 100 % černou v šedotónovém režimu. Při více deskách zůstává barevný souběh. Simulace overprint a ICC nesmějí být nahrazeny pouhým RGB filtrem. Ověřit přesné názvy přímých barev, přiblížení i nezávislost na změnách PDF a na tisku/exportu; viz RND-02.

**Anotace (Q22 – nástroj; Q23 – tisk; Q24 – export; Q25 – základní paleta; Q26 – výchozí obrazovková viditelnost; Q27 – viditelnost jednotlivých dokumentů SCHVÁLENY):** textové komentáře a grafické kresby ukládat jako standardní PDF anotace mimo původní page content; ověřit interoperabilitu s Acrobatem, vyhlazování při zoomu, historii, AutoRecover, viditelnost a nechtěné rasterizace. **Schválená první paleta: komentář, textové pole, volná kresba, čára/šipka, obdélník/elipsa, zvýraznění včetně podtržení a přeškrtnutí; panel Poznámky podle stránek (Q25). Tisk výchozí bez poznámek, volitelně „Tisknout včetně poznámek“ (Q23). Export PDF rovněž výchozí bez poznámek, konzultační export se standardními editovatelnými anotacemi (Q24); ukládání pracovního PDF je zachovává.**

**Uložení pracovní kopie (Q18):** `Katalog.pdf` → úprava → `Uložit upravenou kopii…` = `Katalog-upraveno.pdf` → stejná karta ukazuje kopii → další změna + `Ctrl+S` upraví pouze kopii. Test chyb zápisu a otevření, žádný nechtěný přepis originálu. **AutoRecover (Q19, SCHVÁLENO):** samostatně otestovat pád, restart, správné čištění a obnovu snímků.

**Horní lišta / Nastavení:** dvě nové ikony s tooltipy; globální Nastavení je **samostatné okno s kategoriemi vlevo a obsahem vpravo**, nikoliv pracovní panel nástroje. **Schválené potvrzování: OK = uložit a zavřít, Použít = uložit a zůstat, Zrušit = zahodit jen změny od posledního potvrzení.** Otevření/zavření nemění aktivní dokument ani editaci; uložené vlastní zkratky přetrvají restart a kolize se ohlásí předem; informace o verzi odpovídají sestavení. **Po případném schválení UI-11** testovat `F1` s kontextem, offline návod, hledání v nápovědě a synchronizaci přemapovaných zkratek.

## 10. Navržené pořadí realizace

Práce provádět po menších ověřitelných etapách. Následující pořadí je návrh, který lze po diskusi změnit:

1. **P0 / stabilita a ochrana práce:** UI-02 (pravý okraj okna), OBJ-01 (laso), DOC-02 (přesun stran + potvrzení + Undo/Redo), DOC-03 (Uložit / Uložit kopii a zavření originálu), **DOC-06 (AutoRecover)**, UI-01 (Soubor se zavírá).
2. **P0 / jádro výběru:** SEL-01, SEL-02, OBJ-02, OBJ-03; nejprve sjednotit **MMB tažení obdélníku vs. LMB klikací výběry/editaci/manipulaci**, otestovat levý klik, levý dvojklik, `Ctrl/Shift/Alt` + levé klikání a střední tažení napříč nástroji, potom inteligentní zásah objektů. Nepřepisovat jednotlivé nástroje na neslučitelné mechanismy.
3. **P0 / text:** TXT-01 až TXT-05; začít rozpoznáním slov/segmentů/sloupců a testy, potom caret a věrnost editace, až následně rozšířené styly.
4. **P0 / obraz a OCR:** RND-01 (schválené adaptivní vykreslování Q21), OCR-01, RND-02; nejprve diagnostika cache/zoom a experimentální benchmark vykreslování, nikoliv neověřená výměna renderovací technologie.
5. **P1 / sjednocení UI a historie:** DOC-01, DOC-05, UI-03 až UI-11 (UI-09 a schválené jádro i rozvržení UI-10; další kategorie a UI-11 po rozhodnutí), OBJ-05, TXT-07, SEL-03. **SEL-03 musí být součástí implementace sjednoceného Undo/Redo**, nikoli pouze samostatnou vizuální funkcí. Zkratky a gesta mají být před dokončením této etapy stabilní.
6. **P1–P2 / nová tvorba a přenos:** TXT-06, OBJ-04, DND-00 až DND-03, DOC-04. **Nejprve samostatný technický prototyp** pro interní aplikaci, Průzkumníka a různé cizí editory; až po měření kompatibility přistoupit k produkčnímu provedení. Oddělit plovoucí vizuální náhled od skutečného předání dat a testovat selhání.
7. **P1 / samostatné anotace:** ANNO-01 až ANNO-04; standardní PDF `/Annots` pro textové i grafické poznámky, oddělení od obsahu, kompatibilita Acrobat; nejprve interoperabilní prototyp (nový požadavek Q22). **První funkční paleta včetně panelu podle stránek SCHVÁLENA (Q25); poznámky na obrazovce standardně viditelné s přepínačem (Q26). Tisk bez poznámek ve výchozím stavu (Q23) a export bez poznámek ve výchozím stavu (Q24), s volitelným zahrnutím editovatelných anotací pro konzultaci, jsou SCHVÁLENY.**
8. **P2 / značka:** BRAND-01 a BRAND-02; zachovat renderované assets a otestovat různé velikosti okna.
9. **BACKLOG / až nakonec:** Přidat jednoduchou samostatnou exportní funkci pro vybrané stránky či objekty s dalšími formáty **EPS, JPG, TIFF, SVG** a případně později dalšími; nejprve ověřit možnost zachování vektorové kvality, transparentnost, ICC, rozlišení a limity jednotlivých formátů. **Tento bod není součástí nynější implementační etapy.**

## 11. Otázky k odsouhlasení před implementací

**Q1 – globální výběr – SCHVÁLENO, AKTUALIZOVÁNO:** **Stisknuté kolečko slouží jen k tažení obdélníkového rámečku.** Samostatné stisky ani dvojkliky kolečkem nevybírají. Levé tlačítko slouží ke standardním klikacím výběrům objektů, textu i miniatur, k editaci, kreslení, lasu a přesunu. Tím se ruší rozhodnutí z verze 0.2 o veškerém výběru prostředním tlačítkem.

**Q2 – opakovaný levý dvojklik a výběr textu – SCHVÁLENO, AKTUALIZOVÁNO:** Bez modifikátorů **slovo → segment → věta (i přes více řádků) → odstavec**; `Ctrl + levý klik` v textu = **celá věta** (navržená přímá zkratka); `Ctrl + levý dvojklik` = **vizuální řádek**; `Alt + levý dvojklik` = stejné formátování v segmentu; `Shift + levý dvojklik` = odstavec, opakování = celý sloupec; `Alt + Shift + levý dvojklik` = stejné formátování v aktuálním sloupci. `Ctrl + levý klik` v režimu Objekty nadále prochází hloubku objektů; v miniaturách rozšiřuje výběr stran. Neprovádět výběr věty pouze podle jednoho vizuálního řádku. Určení větných hranic je heuristické, vyžaduje regresní testy.

**Q3 – Ctrl+Z / Ctrl+Y – SCHVÁLENO:** `Ctrl+Z` vrací i jednotlivé kroky výběru textu a také obsahové úpravy, vždy podle jejich **skutečného chronologického pořadí**. `Ctrl+Y` akce znovu provádí. Výběr sám není změnou PDF; nesmí aktivovat příznak neuložených změn. Podrobnosti a testy v SEL-03.

**Q4 – zkratky pro nové tvary – SCHVÁLENO:** Běžné **levé tažení** = libovolný obdélník/elipsa; **`Shift`** = čtverec/kruh; **`Ctrl`** = kreslení od středu; **`Ctrl+Shift`** = čtverec/kruh kreslený od středu. Přesná pravidla, chování náhledu a testy jsou v OBJ-04.

**Q5 – drag-and-drop mimo aplikaci – CÍL A NÁHRADNÍ POSTUP SCHVÁLENY, TECHNICKÉ OVĚŘENÍ VYŽADOVÁNO:** Uživatel požaduje **opticky běžné přetažení, pokud možno s nabídkou formátu až po uvolnění myši**, a možnost připravit obsah v paměti nebo jako dočasný soubor. **Prioritní cílové editory: CorelDRAW, Adobe Photoshop a Affinity**; tyto aplikace mají být nejprve reálně otestovány pro přetažení souborového obsahu jako z Průzkumníka. Pro interní cíl je nabídka po uvolnění plně proveditelná; pro ověřené místní složky lze po uvolnění vytvořit soubor; u cizích editorů jde o **cílové UX, které vyžaduje prototyp**, protože jejich OLE příjem nemáme pod kontrolou. **Nejde-li vložení dokončit, zobrazit chybovou informaci a nabídnout „Uložit jako…“ se svobodnou volbou cílové složky a formátu**, případně kopírování do schránky. Nekončit bez výsledku a nepředstírat úspěch. **Neposouvat volbu před přetažení bez doloženého důvodu.**

**Q6 – náhledy – SCHVÁLENO:** Cílových **6 % šířky klientské oblasti okna** náleží **samotnému obrázku miniatury**; panel se zvětší o ovládací prvky. Režim výchozí automatický, s prostorovou výjimkou pro malé okno a ruční úpravou přes Ctrl+kolečko. Viz UI-08.

**Q7 – sbalování a ouška panelů – SCHVÁLENO:** Po sbalení zůstane **úzká svislá lišta** (varianta B), nikoli osamocené ouško. **Hlavní panely**: malý čtverec se zaoblenými rohy a šipkou. **Pomocné panely**: ouško s viditelným identifikačním textem v otevřeném i sbaleném stavu. **Otevřené pomocné ouško** má barvu panelu a navazující konturu; **sbalená** ouška mají odlišný tlumený odstín. Konkrétní orientace textu může být přizpůsobena dostupné šířce, ale musí být čitelná. Viz UI-04.

**Q9 – ikony a kontextová nápověda – SCHVÁLENO:** Krátký tooltip při hoveru uvede název a zkratku. **Trvalé pole pod ikonami ukazuje aktivní nástroj; při hoveru nad jinou ikonou dočasně ukáže její podrobný popis a po odjetí myši obnoví popis aktivního nástroje.** Stejné dočasné vysvětlení je dostupné při zaměření klávesnicí. Hover nástroj neaktivuje. Viz UI-07.

**Q10 – Nastavení/Nápověda v horní liště, klávesové zkratky a rozvržení Nastavení – SCHVÁLENO:** Vpravo v horní liště jsou samostatné ikony ozubeného kolečka a nápovědy. Nastavení obsahuje skutečnou nainstalovanou verzi, informace o vydání a mapování klávesových zkratek. **Globální Nastavení se otevírá jako samostatné okno: seznam kategorií vlevo, obsah právě zvolené kategorie vpravo; pravý pracovní panel aktivního nástroje zůstává beze změny.** Viz UI-09 a UI-10. **Osm kategorií SCHVÁLENO (Q14); jednotlivé volby mají vlastní implementační priority.** **Modalita schválena – viz Q13.** Potvrzovací tlačítka viz schválené Q12.

**Q12 – potvrzování globálního Nastavení – SCHVÁLENO:** **OK** = validovat, uložit a zavřít; **Použít** = validovat, uložit a pokračovat v Nastavení; **Zrušit** = zahodit pouze změny od posledního úspěšného `Použít` či od otevření, vrátit dočasný náhled a zavřít. Změna kategorií nezahazuje pracovní hodnoty; konfliktní zkratky nelze potvrdit; žádná volba Nastavení není úpravou PDF. Modalita je schválena jako **modální** (Q13); rozsah osmi kategorií je schválen (Q14), jejich konkrétní dílčí volby se zavádějí postupně. Viz UI-10.

**Q14 – osm kategorií Nastavení – SCHVÁLENO:** **Obecné, Rozhraní, Zobrazení PDF, Klávesové zkratky, Soubory, Text a OCR, Barvy a tisk, O aplikaci.** V uživatelském okně zobrazovat kategorii pouze tehdy, když obsahuje alespoň jednu funkční volbu (u O aplikaci platné informace). Žádná nefunkční nastavení ani prázdné kategorie. Schválení osmi kategorií neznamená schválení všech navrhovaných voleb k okamžité implementaci. Viz UI-10.

**Q13 – modální okno Nastavení – SCHVÁLENO:** Globální Nastavení je **modální vůči hlavnímu oknu PDF Blacksmith**: hlavní okno zůstává viditelné pro případný živý náhled, ale nepřijímá úpravy dokumentů ani další interaktivní příkazy. `Použít` nechá dialog otevřený a modální; `OK` a `Zrušit` jej zavřou. `Esc` a křížek odpovídají `Zrušit`; po zavření se vrátí fokus a rozpracované PDF zůstane nedotčené. Podrobné testy viz UI-10.

**Q11 – přístup k Nápovědě – SCHVÁLENO:** Ikona `?` v horní liště otevře **úvod celé příručky**; `F1` otevře **kontextovou kapitolu aktivního nástroje či zaměřené volby**, s bezpečným návratem na nástroj/úvod, když přesná vazba chybí. **K DALŠÍMU ODSOUHLASENÍ:** vlastní česká příručka dostupná offline, vyhledávání **v nápovědě**, odkazy z problematických předvoleb, přehled aktuálních zkratek, „Co je nového“ / „O aplikaci“ a diagnostika. Základem může být `NAVOD.md`, ale obsah musí odpovídat vydané verzi. Viz UI-11.

**Q15 – inteligentní výběr objektů – SCHVÁLENO:** Geometrický hit-test viditelných tahů, výplní a bitmap včetně průhlednosti, tolerance přibližně 5 px a při shodě přednost vyššího objektu. `Ctrl + levý klik` cykluje skutečně zasažené objekty do hloubky a zpět na první. Nevybírat pouze podle bounding boxu; viz OBJ-02 a OBJ-03.

**Q16 – Laso – SCHVÁLENO:** Oba režimy výběru se ovládají dvěma ikonami v panelu Laso. **A „Protnutí“ je výchozí** a vybírá každý objekt s alespoň částí skutečně viditelné kresby protínající oblast polygonu; **B „Celý objekt“** vybírá jen objekty s veškerou skutečně viditelnou kresbou uvnitř. Nepoužívat samotný střed ani hraniční obdélník jako rozhodující kritérium. Uzavření polygonu levým dvojklikem, levým kliknutím u počátečního bodu nebo `Enter`; `Esc` ruší. Kolečko + tažení stále kreslí obdélník. Viz OBJ-01.

**Q18 – ukládání upravené kopie, varianta B – SCHVÁLENO A UPŘESNĚNO:** Po úspěšném ověření kopie se **původní PDF zavře; stávající karta nyní obsahuje výhradně novou kopii**. Originál nezůstane otevřený jako další karta. Následné `Uložit` a `Ctrl+S` ukládají jen do kopie; originál zůstává na disku nezměněný. Při chybě či stornu se nic nezavírá ani nezahazuje. Viz DOC-03.

**Q20 – zavírání dokumentů – SCHVÁLENO:** V duálním zobrazení přímo vybrat „Zavřít levý / Zavřít pravý / Zavřít všechny dokumenty / Zrušit“, s názvem souboru u levé a pravé možnosti. Při jediném otevřeném PDF pouze „Zavřít aktuální dokument / Zrušit“, při více kartách „Zavřít všechny“ navíc. Při neuložených změnách nabídnout jasně pojmenované „Uložit / Uložit upravenou kopii… / Neuložit / Zrušit“; pro vícedokumentové zavírání doporučen **společný souhrnný dialog s rozhodnutím po dokumentech**. Úspěšné uložení předchází uzavření, Neuložit odstraní jen odpovídající AutoRecover; neúspěch zachová práci. Viz DOC-05. **Schváleno:** jeden souhrnný modální dialog s volbou pro každý upravovaný dokument, nikoli postupné dotazování.

**Q19 – automatické obnovovací snímky – SCHVÁLENO:** Oddělený AutoRecover v `%LOCALAPPDATA%\PDFBlacksmith\Recovery`; výchozí časování 60 s nečinnosti / nejvýše 2 min souvislé práce / 2 validní snímky. Po havárii nabídnout obnovu, po vědomém Neuložit snímky odstranit; ukládané PDF nikdy automaticky nepřepisovat. Po uložení kopie přesunout kontext obnovy na kopii a uzavřít originál. Viz DOC-06.

**Q17 – potvrzení změny pořadí stránek – SCHVÁLENO:** Před každým skutečným přeuspořádáním musí modální dialog uvést **počet stránek, přesné zdrojové stránky a pozici v novém pořadí** a nabídnout **Přesunout / Zrušit**. Cancel/Esc/křížek neprovede nic; potvrzení vytvoří jediný krok historie a podporuje `Ctrl+Z` / `Ctrl+Y`. Platí pro tažení miniatur i přesun zadáním čísla. Nedochází-li ke změně pořadí, žádný dialog není nutný; viz DOC-02.

**Q8 – název – SCHVÁLENO:** Oficiální název značky i aplikace je **`PDF Blacksmith`**. `Blaksmith` je překlep a má být opraven ve viditelných textech, výsledném EXE i relevantních názvech zdrojů podle BRAND-03. Ikona zůstává bez slovního názvu (kovadlina + PDF).

**Q21 – adaptivní vykreslování – SCHVÁLENO:** PDFium zůstává primární vykreslovací knihovnou; při pan/zoom rychlé dočasné zobrazení, potom ostré asynchronní překreslení z původního PDF, při vysokém zoomu výřezy/dlaždice a omezená cache. **Schválené obrazovkové volby:** vyhlazování textu, vyhlazování vektorových čar, zvýraznění tenkých čar, rychlý náhled za pohybu a kvalitní překreslení po zastavení. Bez modifikace originálních PDF dat; viz RND-01.

**Q22 – nástroj Poznámky – SCHVÁLENO (Q25 určuje základní paletu):** uživatel požaduje textové poznámky ve stylu Adobe Acrobat i grafické kresby a značky **oddělené od původního obsahu PDF**. Doporučenou technickou reprezentací jsou standardní **PDF anotace** (`/Annots`, `/AP`), nikoli automatické zploštění ani proprietární vrstvy. **Základní paleta i panel poznámek jsou schváleny jako Q25**; **výchozí tisk bez poznámek (Q23) i čistý předtiskový export bez poznámek (Q24) s možností jejich výslovného zahrnutí jsou schválené**. Viz ANNO-01 až ANNO-04.


**Q23 – výchozí tisk anotací – SCHVÁLENO (varianta A):** v tiskovém dialogu má být standardně vypnuté **„Tisknout včetně poznámek“**. Náhled i tisk standardně vynechají veškeré anotace; při výslovném zapnutí se vytisknou tisknutelné anotace v souladu s PDF příznaky. Tato volba nijak nemění anotace v uloženém PDF ani jejich zobrazení při editaci. Podrobnosti ANNO-04.

**Q24 – výchozí export PDF – SCHVÁLENO (varianta B):** pro hlavní účel PDF Blacksmith, tedy předtiskovou přípravu, má běžný export PDF **bez poznámek**. Explicitní přepínač **„Zahrnout poznámky (editovatelné)“** vytvoří konzultační PDF s komentáři/kresbami zachovanými jako anotace. Čistý export nesmí pozměnit či smazat poznámky v otevřeném zdrojovém PDF; `Uložit` a `Uložit upravenou kopii…` anotace zachovávají. Neodstraňovat netextové `/Annots` bez rozlišení (odkazy, formulářové widgety, podpisy), zachovat skutečné vektory a tiskové vlastnosti. Viz ANNO-04.

**Q25 – paleta nástroje Poznámky pro první funkční verzi – SCHVÁLENO:** komentář (`/Text`), textové pole (`/FreeText`), volná kresba (`/Ink`), čára/šipka (`/Line`), obdélník/elipsa (`/Square`/`/Circle`), zvýrazňovač (`/Highlight`), podtržení (`/Underline`) a přeškrtnutí (`/StrikeOut`). Příslušné vlastnosti (barva, síla čáry, průhlednost atd.) a samostatný panel se seznamem podle stránek, navigací, editací a odstraněním podporovaných anotací. Odpovědi na komentáře, razítka a vlastní symboly patří do backlogu. **Výchozí viditelnost poznámek na obrazovce byla schválena jako Q26 (varianta A): standardně viditelné, ručně skrýt/zobrazit jedním přepínačem.** Neměnit již schválené chování tisku/exportu (Q23–Q24). Viz ANNO-02.

**Q26 – obrazovková viditelnost poznámek – SCHVÁLENO (varianta A):** Při otevření PDF se existující poznámky ve výchozím stavu **zobrazují**. Jediný rychle dostupný přepínač **„Zobrazit poznámky“** umí všechny pracovní anotace na stránce skrýt/zobrazit, aniž by cokoliv mazal, měnil dokument nebo vyvolal stav neuložených změn. Skryté poznámky neblokují výběr podkladových objektů. **Nezávislé na Q23 a Q24:** běžný tisk a export zůstávají bez anotací bez ohledu na jejich obrazovkovou viditelnost. Rozsah přepínače byl následně schválen jako **Q27: samostatně pro každý otevřený dokument**; viz ANNO-02.

**Q27 – rozsah přepínače „Zobrazit poznámky“ – SCHVÁLENO (varianta A):** Každý otevřený dokument má **svůj vlastní stav viditelnosti anotací**. Přepnutí jedné karty nezmění ostatní dokumenty; při přepnutí se ovladač aktualizuje podle aktivního PDF. V duálním zobrazení jsou levý a pravý dokument nezávislé; přepínání platí pro zvolený aktivní dokument. Skrývání anotací není editace PDF, neaktivuje dirty stav, nemění obnovovací kopie ani výchozí čistý tisk/export (Q23–Q24). **Akceptace:** otevřít `A.pdf` a `B.pdf`, v A skrýt poznámky, v B nechat viditelné; několikrát přepnout karty a duální zobrazení, ověřit nezávislý stav a správnou polohu přepínače; uložit/tisknout/exportovat → PDF bez změny způsobené pouhou viditelností; po zcela novém otevření bez obnovení relace jsou poznámky zase výchozí viditelné podle Q26. Viz ANNO-02.

**Q30 – pokrytí tiskových barev pod kurzorem a limit TAC – SCHVÁLENO:** V panelu Separace zobrazovat skutečné lokální procento jednotlivých tiskových desek včetně odděleného odečtu spot barev a **součet procesních složek C + M + Y + K (TAC)**. Nabídnout nastavitelnou mez (přednastavení 300 %, uživatelem změnitelné podle tiskové technologie) a volitelné **zvýraznění míst nad limitem**. Odečet musí pocházet z věrného separačního výpočtu, ne z RGB bitmapy; při neúplné podpoře viditelně přiznat limity. Náhled ani měření nesmějí změnit PDF či ovlivnit skutečné tiskové/exportní cesty. Viz RND-02.

**Q31 – zachování stávajících schopností aplikace při každé úpravě – SCHVÁLENO:** Nové požadavky jsou zásadně **rozšíření/opravou současných nástrojů**, nikoli povolením odstranit jejich již funkční prvky. Před prací pro každý dotčený nástroj inventura funkcí v kódu a pokud možno v reálném UI, potom změna a **regresní ověření všech původních i nových schopností**. U Separace chránit zejména existující převodní režimy, ICC/Gray/K, práci s originálem a novou kopií, tisk, export a diagnostiku. Bez výslovného souhlasu nesmazat přepínače, akce, předvolby, klávesové vazby ani obslužný backend. Závazné pro všechny sekce specifikace.

## 12. Ověřené referenční vzorce z profesionálních aplikací

- [Adobe Acrobat: editace a formátování textu](https://helpx.adobe.com/acrobat/desktop/edit-documents/edit-text-in-pdfs/add-text.html) – textové rámy, vlastnosti písma; omezení fontů.
- [Adobe Acrobat: omezení editace textu při chybějícím písmu](https://helpx.adobe.com/uk/acrobat/desktop/edit-documents/edit-text-in-pdfs/modify-text.html) – proč nelze garantovat zachování každého fontu.
- [Adobe Acrobat: přesouvání stran](https://helpx.adobe.com/nz/acrobat/desktop/edit-documents/organize-pages/move-pages.html) – vybrat a přesunout miniatury.
- [Inkscape: seskupování a výběr pod objektem](https://inkscape-manuals.readthedocs.io/en/latest/grouping.html) – procházení překrytých objektů.
- [Inkscape: tvary](https://inkscape-manuals.readthedocs.io/en/latest/shape-tools.html) – Ctrl omezení proporcí a Shift kreslení od středu.
- [Adobe Illustrator: klávesové zkratky](https://helpx.adobe.com/illustrator/desktop/get-started/preferences-and-settings/default-keyboard-shortcuts-for-illustrator.html) – jiná konvence Shift/Alt pro tvar; rozhodnutí nelze nechat náhodě.
- [PDFium API `FPDF_RenderPage` a `FPDF_RenderPageBitmap`](https://pdfium.googlesource.com/pdfium/+/refs/heads/main/public/fpdfview.h) – oddělení věrnosti zobrazení a struktury PDF.
- [Adobe PDF Reference – annotations, `/Annots`, appearance streams](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/pdfreference1.6.pdf) – anotace jsou uloženy mimo stream obsahu stránky; nejde o běžnou OCG vrstvu.
- [PDFium – `fpdf_annot.h`](https://pdfium.googlesource.com/pdfium/+/refs/heads/main/public/fpdf_annot.h) – API a podporované typy anotací, ověřit konkrétní verzi před implementací.
- [Adobe Acrobat – komentáře a grafické značky](https://www.adobe.com/acrobat/how-to/annotate-pdf-online.html) – referenční uživatelské chování nástroje Poznámky.
- [PDF Association – pravidla appearance streams v PDF 2.0](https://pdf-issues.pdfa.org/32000-2-2020/clause12.html) – interoperabilita a viditelnost anotací.
- [Microsoft: Windows Shell clipboard / drag-and-drop](https://learn.microsoft.com/en-us/windows/win32/shell/clipboard) – virtuální soubory a formáty dat.
- [Microsoft: OLE drag-and-drop](https://learn.microsoft.com/en-us/windows/win32/com/drag-and-drop) – zdroj, cíl a předání přes `IDataObject`.
- [Microsoft: Shell data transfer scenarios](https://learn.microsoft.com/en-us/windows/win32/shell/datascenarios) – `CFSTR_FILEDESCRIPTOR` / `CFSTR_FILECONTENTS`.
- [Microsoft: layered windows](https://learn.microsoft.com/en-us/windows/win32/winmsg/window-features) – transparentní okna a hit-test.
- [Microsoft: SendInput](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendinput) – omezení UIPI pro syntetické vstupy.

---

**Použití:** Tento dokument je podklad pro společné upřesňování. **Q21 (adaptivní zobrazení), Q22 (nástroj Poznámky), Q23 (tisk bez poznámek), Q24 (export bez poznámek), Q25 (základní paleta a panel Poznámky), Q26 (standardně viditelné poznámky), Q27 (samostatná viditelnost v každém dokumentu), Q28 (přepínání tiskových desek) a Q29 (šedotónový náhled jediné desky s volitelnou tiskovou barvou) jsou schváleny; pokročilé doplňky zůstávají k dalšímu rozhodnutí.** **Varianta B (Q18), včetně zavření originálu po úspěšném přepnutí na kopii, i AutoRecover (Q19) jsou schváleny.** **Q1–Q4, Q6–Q16 mají schválené jádro podle výše uvedeného; Q5 má schválený uživatelský cíl s povinným technickým prototypem. Osm kategorií globálního Nastavení je schváleno; konkrétní dílčí volby a doplňkové funkce Nápovědy se zavádějí po částech. Modální provedení Nastavení je schválené jako Q13.** Po dokončení rozhodnutí rozděl jednotlivé úkoly na malé implementační prompty nebo GitHub Issues. Při každém promptu doplň konkrétní zjištěnou příčinu, změněné moduly a ověřitelná akceptační kritéria; neopakuj zbytečně celý rozsáhlý dokument.

### Doplňující technické reference pro přetažení do editorů (v0.9)

- [Adobe Photoshop – Place a file](https://helpx.adobe.com/photoshop/desktop/create-open-import-images/import-files/place-a-file-in-photoshop.html) – potvrzuje podporu přetažení souboru z Průzkumníka přímo do dokumentu.
- [Adobe Photoshop – Place Illustrator and PDF files](https://helpx.adobe.com/photoshop/desktop/app-integrations/place-illustrator-files-or-pdf-in-photoshop.html) – příjem PDF a umístění jako Smart Object, s omezeními dle cílové aplikace.
- [CorelDRAW – Supported file formats](https://product.corel.com/help/CorelDRAW/540111176/Documentation-Windows/CorelDRAW-en/CorelDRAW-Supported-file-formats.html) – přehled podporovaných formátů; samotný import souboru automaticky nedokazuje chování při OLE dropu.
- [Affinity – Embedding documents](https://s3-eu-west-1.amazonaws.com/affinity-docs/help/photo/en-US.lproj/pages/GetStarted/embeddedDocuments.html) – podpora PDF, SVG, EPS a dalších dokumentů při vkládání.
- [Microsoft – DoDragDrop](https://learn.microsoft.com/en-us/windows/win32/api/ole2/nf-ole2-dodragdrop) a [Shell Data Transfer Scenarios](https://learn.microsoft.com/en-us/windows/win32/shell/datascenarios) – OLE kontrakt zdroj/cíl, skutečné a virtuální soubory.


## Změny verze 2.2

- **Schváleno Q17:** potvrzení přesunu jedné či více stránek před provedením změny, včetně počtu, identifikace stránek a konečné cílové pozice.
- Zpřesněna DOC-02: kontext souvislých/nesouvislých výběrů, Esc/Zrušit, nulová změna bez dialogu, atomické Undo/Redo a testovací scénáře.
- **Q18 SCHVÁLENO:** „Uložit upravenou kopii…“ aktivuje po ověřeném úspěšném uložení novou kopii ve stávající kartě; další úpravy a ukládání se týkají kopie, nikoli originálu.


## Změny verze 2.3

- **Schváleno Q18, varianta B:** „Uložit upravenou kopii…“ po úspěšném zápisu a otevření nové kopie aktivuje kopii v aktuální kartě; následné `Ctrl+S` zapisuje už do kopie.
- Upraveno DOC-03: přesné chování při úspěchu, stornu, chybě a změně cesty; bezpečnost původního dokumentu.
- **Q19 následně SCHVÁLENO ve verzi 2.4:** samostatný mechanismus AutoRecover, odlišný od prostých dočasných souborů; obnova po havárii, životní cyklus kopií, bezpečnost a testy v DOC-06.
- Nové testy pro opakované ukládání aktivní kopie a obnova po havárii.


## Změny verze 2.4

- **Q18 UPŘESNĚNO:** Po úspěšném příkazu **Uložit upravenou kopii…** se původní PDF **zavře**. Stávající karta se nahradí kopií; původní soubor nezůstává zároveň otevřený. Nezavírat před dokončením zápisu a ověření, při stornu/selhání zachovat původní pracovní stav.
- **Q19 SCHVÁLENO:** AutoRecover pro obsahové změny, oddělená složka `%LOCALAPPDATA%\PDFBlacksmith\Recovery`, výchozí interval 60 s po nečinnosti nebo nejpozději 2 min při souvislé práci, poslední 2 validní snímky, nabídka obnovy po neočekávaném ukončení. Po **Neuložit** obnovovací data daného dokumentu bezpečně zahodit.
- **Návaznost:** při uložení kopie neztratit rozpracované údaje ani recovery; po ověření bezpečně převázat relaci obnovy na novou kopii a odstranit nadbytečné snímky originálu. Originál a další nezúčastněné dokumenty na disku neměnit.


## Změny verze 2.5 – dřívější návrh

- Revidováno **DOC-05** podle požadavku nepoužívat při zavírání tlačítka Ano/Ne. Rozděleno na volbu rozsahu zavření a bezpečné rozhodnutí o neuložených změnách, včetně samostatných akcí pro levý, pravý a všechny dokumenty.
- Navržen **jeden souhrnný dialog** pro více rozpracovaných dokumentů s volbou Uložit / Uložit upravenou kopii… / Neuložit po dokumentech; zachování AutoRecover a bezpečnostní chování při neúspěšném ukládání.
- **Q20 byl následně schválen ve verzi 2.6.** Viz DOC-05 a souhrn rozhodnutí.


## Změny verze 2.6 – schváleno Q20

- **Q20 SCHVÁLENO:** jeden souhrnný modální dialog při zavírání více upravených PDF; nezobrazovat řadu otázek pro každý soubor.
- Každý dotčený dokument má samostatnou volbu **Uložit / Uložit upravenou kopii… / Neuložit**, celkově **Pokračovat / Zrušit**. Pokud se zavírá pouze jeden změněný dokument, zůstává jednoduchý jednotlivý dialog.
- Rozdělit výběr rozsahu zavření (levý / pravý / všechny / aktuální) od ochrany neuložených změn. Ostatní karty a nezúčastněné soubory se nesmějí změnit.
- **Bezpečnost a AutoRecover:** žádné vymazání neuložené práce při stornu či chybě; uložit a ověřit potřebné soubory před zavřením příslušných karet.
- Další projednávaný bod: **RND-01** – jemné vektorové hrany při přiblížení, souběh PDFium rendereru a OCR.


## Změny verze 2.7 – schválené zobrazení a nástroj Poznámky

- **Q21 SCHVÁLENO:** adaptivní přepočet náhledu z původních vektorových PDF dat, dlaždicové vykreslování při přiblížení, omezená mezipaměť a pět samostatných voleb obrazovkového zobrazení. Aktualizován RND-01 i kategorie Nastavení → Zobrazení PDF.
- **Q22 NOVÝ POŽADAVEK:** přidat nástroj **Poznámky** pro textové komentáře, textové rámečky a grafické anotace/kresby. Preferovat standardní PDF anotace (`/Annots` + odpovídající appearance streams) oddělené od původního page contentu. Zapsáno ANNO-01 až ANNO-04 včetně vazby na Undo/Redo, AutoRecover, bezpečné ukládání, oddělenou viditelnost a kompatibilitu s Acrobatem.
- **Pozor na schvalování:** existence nástroje Poznámky a princip oddělení obsahu jsou požadavkem uživatele; detailní druhy anotačních nástrojů, vzhled panelu a výchozí chování exportu jsou zatím **návrhy**. Výchozí zahrnutí do **tisku** se schvaluje až jako Q23 (varianta A) ve verzi 2.8.


## Změny verze 2.8 – schválen výchozí tisk poznámek (Q23)

- **Q23 SCHVÁLENO – varianta A:** běžný tisk je bez poznámek; v tiskovém dialogu volitelný přepínač **„Tisknout včetně poznámek“**, výchozí vypnutý.
- Zapnutí promítat do **tiskového náhledu i skutečné tiskové úlohy**; respektovat příznaky tisknutelnosti PDF anotací a prověřit kombinaci s náhledem Separace.
- Volba tisku neovlivňuje viditelnost poznámek v editoru, editovatelné anotace uložené v PDF ani vektorovou a barvovou věrnost původní stránky. Neprovádět automatické zploštění poznámek.
- **Export anotací byl následně schválen jako Q24 (varianta B):** výchozí čisté předtiskové PDF bez poznámek, konzultační export s volitelnými editovatelnými poznámkami; běžné ukládání pracovní PDF komentáře zachová. Nadále k rozhodnutí zůstává přesný rozsah anotačních nástrojů a panelu.


## Změny verze 2.9 – schválen výchozí export bez poznámek (Q24)

- **Q24 SCHVÁLENO – varianta B:** u výstupu určeného k předtiskové přípravě běžný export PDF bez textových komentářů, grafických kreseb, zvýraznění a dalších anotací typu poznámka.
- V exportním dialogu přepínač **„Zahrnout poznámky (editovatelné)“**, výchozí vypnuto. Po zapnutí export konzultačního PDF včetně standardních editovatelných anotací (bez automatického flatten).
- **Rozhraní a bezpečnost:** `Exportovat PDF` a `Uložit` / `Uložit upravenou kopii…` jsou odlišné operace. Ukládání zachovává anotace v pracovním dokumentu; čistý export vytváří nový soubor a zdroj nemění. Rozlišovat komentářové anotace od odkazů, formulářových a podpisových widgetů.
- Vytvořit regresní testy čistého výstupu, editovatelných anotací pro konzultace, zachování CMYK/K/ICC/overprint a PDF vektorů, bezpečného uložení i AutoRecover.
- **K této verzi:** detailní paleta a panel poznámek stále čekaly na schválení; toto rozhodnutí bylo uzavřeno až jako Q25 ve verzi 3.0.

## Změny verze 3.0 – základní paleta nástroje Poznámky SCHVÁLENA (Q25)

- **Q25 SCHVÁLENO:** do první funkční verze zahrnout šest vizuálních skupin nástrojů: Komentář, Textové pole, Volná kresba, Čára/šipka, Obdélník/elipsa a Zvýraznění, navíc podtržení a přeškrtnutí textu. Preferovat standardní PDF typy `/Text`, `/FreeText`, `/Ink`, `/Line`, `/Square`, `/Circle`, `/Highlight`, `/Underline` a `/StrikeOut`.
- **Vlastnosti:** relevantní nastavení barvy, šířky čáry, průhlednosti, výplně a vlastností textového pole; možnost úprav a mazání poznámek, příslušné Undo/Redo a AutoRecover.
- **Panel Poznámky:** samostatný seznam podle stránek s navigací k anotaci, možností úprav/odstranění podporovaných poznámek, respektování pravidel UI pomocných panelů.
- **Pozdější etapa:** odpovědi na komentáře, razítka, vlastní symboly, callouts, polygony a pokročilé filtrování. **Výchozí viditelnost anotací na obrazovce je schválena následně jako Q26 ve verzi 3.1.**
- **Nezměněné schválené chování:** pracovní Uložit zachovává anotace; **běžný tisk i export bez poznámek** a jejich volitelné explicitní zahrnutí pro konzultace (Q23–Q24).


## Změny verze 3.1 – výchozí viditelnost poznámek SCHVÁLENA (Q26)

- **Q26 SCHVÁLENO, varianta A:** všechny stávající poznámky na otevřené stránce se standardně zobrazují; uživatel je může skrýt či znovu zobrazit jediným přepínačem „Zobrazit poznámky“, bez nutnosti otevření panelu Poznámky.
- Skrývání je **pouze obrazovková vlastnost** a nemění ani PDF, ani příznak neuložených změn. Skryté komentáře neblokují hit-test podkladových objektů; v panelu Poznámky zůstávají přehledně vedené.
- **Tisk a export se nemění:** výchozí tisk a export jsou nadále bez poznámek (Q23–Q24). Zobrazení anotací v editoru nesmí samo přepínat jejich zahrnutí do tiskového či exportního výstupu.
- **Akceptační test Q26:** otevřít PDF s textovým komentářem, textovým rámečkem, volnou kresbou a zvýrazněním → všechny viditelné; vypnout „Zobrazit poznámky“ → všechny skryté, lze vybrat podklad, PDF není označeno jako změněné; znovu zapnout → anotace opět viditelné a editovatelné. V obou stavech provést běžný tisk/export → bez poznámek; s výslovným zahrnutím → podle příslušné tiskové či exportní volby.
- **Tato dříve otevřená otázka je uzavřena jako Q27 ve verzi 3.2:** samostatně pro každý dokument.


## Změny verze 3.2 – samostatná viditelnost poznámek pro každý dokument SCHVÁLENA (Q27)

- **Q27 SCHVÁLENO, varianta A:** „Zobrazit poznámky“ je vlastnost každého otevřeného PDF samostatně, ne globální přepínač. V jedné kartě mohou být poznámky skryté a v jiné viditelné.
- V duálním zobrazení je stav levého a pravého dokumentu oddělený; přepínač vždy ukazuje a mění stav aktuálně aktivního dokumentu. Při přepnutí karty či aktivní poloviny se musí synchronizovat stav ovladače, **ne** stav ostatních dokumentů.
- Skrytí pouze mění obrazovkové vykreslení a výběr anotací; neoznačí PDF jako změněné, nic nemaže a nemá vliv na schválený výchozí **tisk/export bez poznámek** (Q23–Q24).
- Nově otevřený dokument má výchozí viditelnost zapnutou podle Q26. Při přechodu na upravenou kopii v téže kartě zůstane pracovní stav plynule zachován.
- **Akceptační test:** A/B s různými anotacemi; skrýt pouze v A, přepínat karty a duální zobrazení, prověřit viditelnost i zásahy myší, dále uložit, tisknout a exportovat bez zásahu do dat dokumentů.
- **Tento bod je následně schválen jako Q28 ve verzi 3.3:** výchozí všechny reálné separace zapnuté, individuální přepínání a „Jen tato“/„Zobrazit všechny“.


## Změny verze 3.3 – ovládání náhledu separací SCHVÁLENO (Q28)

- **Q28 SCHVÁLENO:** panel Separace po otevření dokumentu zapne **všechny skutečně dostupné tiskové desky**; jednotlivé barvy CMYK a případné skutečné pojmenované spot barvy lze nezávisle vypnout/zapnout. Akce „Jen tato“ izoluje jednu desku, „Zobrazit všechny“ obnoví plný souběh.
- **Neztrátový charakter náhledu:** přepínače mění pouze obrazovkové vykreslení, nikoli obsah souboru PDF, reálné separace, dirty stav, výsledný tisk ani export. Simulace přetisku a spot barev musí respektovat barevný management; nezaměňovat skutečnou separaci za jednoduchý filtr nad RGB bitmapou.
- **Kvalita:** adaptivně překreslovat z odpovídajících dat ve vysokém rozlišení po zoomování a posouvání; ověřit zachování detailů, textu a věrnosti. Pokud renderer nepodporuje přesný náhled, přiznat omezení.
- **Testy:** CMYK, CMYK + spot, odlišné kanály na různých stranách, všechny vypnuté / izolace jedné / obnova všech; overprint, ICC, OCR/anotace a různé zoomy.
- **Dřívější otevřená otázka je uzavřena jako Q29 ve verzi 3.4:** výchozí šedotónový náhled jediné desky, s možností přepnout na simulovanou tiskovou barvu.


## Změny verze 3.4 – zobrazení jediné separace SCHVÁLENO (Q29)

- **Q29 SCHVÁLENO, varianta A:** V režimu izolace jediné tiskové desky je výchozím zobrazením **šedotónový náhled jejího pokrytí**; přepínač **„Odstíny šedi / Tisková barva“** umožní vizualizovat tentýž kanál jeho procesní nebo přímou barvou.
- **Interpretace šedi:** 0 % barvy = bílý podklad, 100 % barvy = černá, mezihodnoty = úměrné odstíny. Nepoužívat prostou desaturaci barevně složeného náhledu; jde o hodnoty skutečně vybrané separace.
- **Více desek:** Při souběžném zobrazení více separací se zachová profilovaný barevný souběh a simulace přetisků dle Q28; šedotónové nastavení jediné desky není převodem celé stránky ani úpravou výstupních barev.
- **Bezpečnost dat:** Přepnutí režimu nemění PDF, barvové prostory, profily, přetisky, procenta krytí, dirty stav ani tisk a export. Doplněny testy v RND-02.
- **Následně SCHVÁLENO v3.5 jako Q30:** odečet skutečných podkladových separací a TAC pod kurzorem, nastavitelný limit a volitelné zvýraznění; globální pravidlo zachování stávajících funkcí schváleno jako Q31.


## Změny verze 3.5 – měření pokrytí a zachování původních funkcí SCHVÁLENO (Q30, Q31)

- **Q30 SCHVÁLENO:** Doplnit do existujícího nástroje **Separace** procentuální odečet jednotlivých procesních i přímých barev pod kurzorem, výpočet **TAC jako C+M+Y+K**, nastavitelný limit (výchozí 300 % jako editovatelná pracovní předvolba) a volitelnou vizuální indikaci překročeného limitu. Nepoužívat pouhý RGB výsledek jako zdroj přesných hodnot, při nedostatečné přesnosti zobrazit omezení. Nezávislé na přepínání viditelnosti náhledových desek; hodnoty reprezentují skutečnou separaci dokumentu.
- **Q31 SCHVÁLENO – hlavní nový pokyn pro Codex:** **Neodstraňovat ani nenahrazovat stávající funkce při přidávání nových.** Před každým zásahem katalog původních schopností, potom rozšíření/integrace a regresní testy. Již existující režimy převodu barev a ICC z `COLOR_MANAGEMENT.md` a ostatní operace Separace jsou povinně zachované.
- Doplněna společná instrukce Codexu, RND-02, samostatná akceptační kritéria a uzavření Q30/Q31; všechna dřívější schválená rozhodnutí zůstávají platná.

## Změny verze 3.6 – audit existujícího nástroje Separace a návrh grafického vylepšení

- Reálný audit implementace v `viewer_windows.go`, `workspace_windows.go`, `proof.py`, `tests/preview.py` a technické dokumentace. Již přítomné náhledové, skenovací, ICC a převodní operace jsou katalogizovány jako **zachovat**. Nejde o požadavek vybudovat Separace znovu.
- Odhalena důležitá odlišnost: dosavadní metrika „plocha“ znamená podíl pixelů s nenulovou separací; nejde o lokální tón ani TAC. Q30 je proto nové měření, ne přejmenování stávající metriky.
- Navrženo šest vizuálně oddělených skupin pravého panelu: Náhled tisku, Tiskové desky, Zobrazení jedné desky, Měření pod kurzorem, Kontrola dokumentu, Převod barev aktuálního dokumentu. **Konkrétní UI návrh čeká na schválení.**
- Audit je založen na kódu. Windows UI nebylo v tomto kroku spuštěno a testy nebyly lokálně vykonány.


## Změny verze 3.7 – vizuální audit stávajícího panelu Separace (na základě snímků z 20. 11. 2026)

**Podklad:** Čtyři uživatelem zaslané obrázky současného běžícího Windows rozhraní: kompletní panel Separace, rozbalený seznam režimů barevného převodu, rozbalená nabídka Gray ICC a rozbalené cílové CMYK profily. Zjištění níže se opírají o viditelné UI a dříve provedenou kontrolu `viewer_windows.go` + `proof.py`. Bez provedeného funkčního testu spuštěné aplikace; neoznačovat chování za ověřené mimo obrazovku.

### Skutečně zobrazené a zachovávané existující prvky

1. **Náhledový profil tisku:** `Podle PDF / výchozí CMYK`, další volitelné ICC profily (předvolby jsou dynamické podle katalogu aplikace). Neměnit význam výchozí volby.
2. **Tisková oblast:** `Celý arch`, alternativně ořez zobrazení, čistý formát, spadávka; vazba na PDF rámečky se musí zachovat.
3. **Výstupní záměr a simulace:** samostatné informace; snímek ukazuje `Neuveden` a `Výchozí CMYK Ghostscriptu`. Případ bez OutputIntent popsat srozumitelně, bez předstírání skutečného profilu z dokumentu.
4. **Tiskové desky:** zaškrtávací seznam aktuálních C, M, Y, K (přímé barvy, pokud jsou v PDF); u každé dosud údaj `plocha / max tón`. Například na snímku `Cyan 96,55 % / max 72,2 %`. To je **podíl bitmapových bodů s nenulovým tónem** a **maximální nalezený tón v renderu**; není to tón pod kurzorem ani pokrytí TAC. Zásadně zachovat všechny tři informace a správně je vysvětlit.
5. **Hromadná tlačítka:** `Všechny desky`, `Vypnout všechny`; **neodstraňovat**, jen doplnit již schválenou akci `Jen tato` na každém řádku.
6. **Aktuální stránka / rozlišení:** dosavadní text `Str. 1 · plocha / maximální tón · 155 dpi` zachovat fakticky i místem; rozlišení je dynamická vlastnost konkrétního renderu, nikoli slib pevné hodnoty 150 dpi.
7. **Kontrola rozsahu:** `Zkontrolovat celý dokument`, `Zkontrolovat vybrané náhledy`, výsledková oblast; současné umístění zachovat (případné úpravy mezer nejsou součástí Q32).
8. **Převod barev aktuálního dokumentu:** nabídka zachovat přesně všechny tři režimy: `Stupně šedi · DeviceGray`, `Složená černá → K100`, `Pouze K · CMYK C0 M0 Y0`. Tytéž tři režimy mají odlišnou technickou sémantiku; neslučovat.
9. **ICC profil převodu do šedi:** zachovat `Automaticky podle dokumentu`, `Gray Gamma 2.2`, `Gray Gamma 1.8`, `Vlastní ICC Gray…` a odpovídající výběr souboru, pokud je relevantní pro režim.
10. **Cílový tiskový profil CMYK:** zachovat stávající dostupné volby včetně obou FOGRA39 variant, FOGRA47, FOGRA51, FOGRA52 a FOGRA54, s podporou případného rozšíření katalogu a aktuálního výběru. Nezplošťovat do jediné předvolby.
11. **Tlačítko skutečné operace:** `Použít převod a otevřít kopii…`, samostatné od pouhého obrazovkového náhledu; zachovat a navázat na schválenou práci s kopií.
12. **Globální styl:** tmavý panel, vínový/růžový akcent, názvosloví panelových oušek a stávající konvence ovládacích prvků. Redesign musí působit jako jedna aplikace; neimportovat cizorodý webový panel nebo překvapivé barevné téma.

### Rozhodnutí po zhlédnutí screenshotů – Q32 SCHVÁLENO (nahrazuje návrhy v3.6/v3.7)

Uživatel z návrhů grafických úprav vybral **výhradně barevné identifikační značky tiskových desek**. Ostatní připomínky ze screenshotů (velké svislé mezery, přeuspořádání číselných údajů, přesun sekcí, sbalitelné skupiny, dvě vnitřní záložky, přesunutí převodů do zvláštní karty) zůstávají **jen historickými návrhy, nikoli zadáním pro Codex**. Bez nového výslovného souhlasu se nesmějí implementovat.

**Q32 – přesné zadání:** Do původních řádků `Cyan`, `Magenta`, `Yellow`, `Black` a skutečných přímých barev doplnit malé čitelné barevné vzorky podle RND-03/C; nic dalšího nepřestavovat. Zachovat včetně rozbalovacích nabídek, popisků, současných statistik, kontrolních tlačítek, převodů a nastavení ICC. Samostatně schválená funkční vylepšení Q28–Q30 tím **nejsou zrušena**; jejich implementaci řešit jako lokální doplnění existujícího panelu.

**Ověření:** viz RND-03/C a regresní požadavky Q31.

## Změny verze 3.8 – minimalistické vizuální doplnění Separace

- **Q32 SCHVÁLENO:** pouze malé barevné značky CMYK/spot u stávajících řádků tiskových desek.
- **VÝSLOVNĚ NESCHVÁLENO:** dvě vnitřní záložky, šest sbalitelných bloků, automatické přeuspořádání, přesuny ICC/profilů nebo další celkové grafické změny navržené ve verzích 3.6 a 3.7.
- **Zachovat:** veškeré stávající funkce a vzhled kromě barevných značek; funkční požadavky Q28, Q29 a Q30 stále platí. Východiskem jsou přiložené screenshoty a aktuální Go/Win32 kód (Q31).

## Změny verze 3.9 – referenční snímky dalších nástrojů a celé aplikace (BEZ NOVÝCH ROZHODNUTÍ)

**Povaha této aktualizace:** Uživatel 20. 11. 2026 poskytl reálné snímky současných panelů `Objekty`, `Text`, `OCR`, `Digitální podpis`, `Rozdělit PDF`, `Sloučit PDF`, `Pořadí stránek` a celého okna aplikace. Slouží jako **referenční stav před rozšiřováním**. Pořízení snímků **není souhlas s redesignem**, změnou rozložení ani odebráním ovládacích prvků. Přednost má Q31 (zachovat stávající funkce) a poslední schválené rozhodnutí Q32 (u Separace pouze barevné značky).

### Inventura existujícího rozhraní ze snímků

| Panel / oblast | Na snímku doložené dosavadní prvky – povinně zachovat | Poznámka k implementaci |
|---|---|---|
| **Objekty** | Tři volby výběru: ukazatel, obdélníkový výběr a laso; kontextové vysvětlení nástroje; stav výběru; `Kopírovat`, `Vložit`, `Duplikovat`, `Smazat`, `Vrátit změnu` včetně zkratek. | Případné zkracování textového vysvětlení řešit bez ztráty informace; zobrazená neaktivní tlačítka odpovídají stavu bez výběru, nikoli prokazatelné závadě. Q1–Q3 a Q16–Q18 navazují na původní chování. |
| **Text** | Ikony textového režimu, odstavce, obdélníkového výběru a lasa; popis; typ písma; velikost v bodech; zvětšení/zmenšení; styl, zarovnání; šířka, řádkování, proklad; spodní akce. | Zachovat textovou editaci i všechny existující formátovací ovladače, byť jsou v neaktivním stavu kvůli chybějícímu výběru; Q2, Q3 a textové požadavky přidávat, nikoli nahrazovat. |
| **OCR** | `Strany pro OCR` s `Všechny strany`, `Aktuální strana`, `Vlastní výběr`, `Vybrané náhledy`; `Jazyk rozpoznávání` (snímek `Čeština + angličtina`); `Rozpoznat text` a `Vrátit OCR`. | Zachovat výběr rozsahu, jazyků, operaci a vrácení; případné OCR na oblasti stránky je rozšíření, nikoli zrušení dávkového OCR. |
| **Digitální podpis** | Volba certifikátu PKCS#12 `.p12`/`.pfx`, heslo k certifikátu, `Viditelný podpis v novém poli`, uložení podepsaného PDF, ověření s možností doplňkového kořenového certifikátu. | Zachovat **oba** pracovní postupy: vytváření a ověřování podpisů. Nezaměňovat s poznámkami. |
| **Rozdělit PDF** | Popis rozdělení po stránkách; výběr `Všechny strany`, `Aktuální strana`, `Vlastní výběr`, `Vybrané náhledy`; `Vybrat složku a rozdělit PDF…`. | Zachovat vazbu na aktuální pořadí stránek a výběr miniatur. |
| **Sloučit PDF** | `Přidat PDF soubory…`, `Přidat otevřené PDF`, `Přidat výběr stran`, `Nahoru`, `Dolů`, `Odebrat`, `Strany k přidání`, `Uložit sloučené PDF…`, heslo k otevření vybraného souboru, `Načíst počet stran`. | V prázdném stavu jsou některé ovladače neaktivní a seznam bez položek; neoznačovat tyto skutečnosti samy o sobě za chybu. Zachovat změnu pořadí a práci se zabezpečenými zdroji. |
| **Pořadí stránek** | Cílová pozice, `Přesunout`, `Posunout před`, `Posunout za`, otočení vlevo/vpravo, `Smazat vybrané`, `Vrátit změnu`, `Uložit upravenou kopii…`. | Nová potvrzení přesunu Q21–Q22 doplnit k existujícím operacím; neredukovat na samotný drag-and-drop. |
| **Celé okno** | Horní lišta s kartou otevřeného PDF, ikonami historie a značkou, levé menu nástrojů, panel miniatur se samostatným ouškem, centrální plátno, pravý panel s oušky, dolní lišta navigace a měřítka. | Zachovat obecnou vizuální identitu tmavého Win32 GUI; použít jako výchozí před/po pro práci s uspořádáním panelů. Snímek horní značky zobrazuje starší nápis `BLAKSMITH`, ale schváleným oficiálním názvem je **PDF Blacksmith** (BRAND-03). |

### Povinné zásady pro Codex při využití snímků

1. **Snímky reprezentují současnou implementaci**, nikoliv nový požadovaný návrh. Přítomné funkce zachovat podle Q31.
2. U každé úpravy uvést, které dosavadní ovladače a klávesové zkratky zůstaly; přidat regresní test nebo explicitní ruční kontrolu.
3. **Neaktivní (šedá) tlačítka** nepovažovat bez ověření kontextu za chybu; často jde o prázdný výběr či neúplné zadání.
4. Stávající délky a rozvržení panelů automaticky nepřestavovat; přidávání schválených funkcí nesmí poškodit dosavadní ovládání.
5. Vizuální reference v ZIP balíčku jsou umístěny relativně v adresáři `reference/`; v případě použití samostatného `.md` musí být obrázky předány Codexu zvlášť.

### Seznam referenčních obrázků (viz přiložený ZIP balíček)

- `01_separace_panel.png` – Separace – kompletní panel.
- `02_separace_rezimy_prevody.png` – Separace – volba barevného převodu.
- `03_separace_gray_icc.png` – Separace – ICC profily převodu do šedi.
- `04_separace_cmyk_icc.png` – Separace – cílové profily CMYK.
- `05_objekty.png` – Objekty – výběr, operace a historie.
- `06_text.png` – Text – výběr, typografie a akce.
- `07_ocr.png` – OCR – rozsah, jazyk a spuštění.
- `08_ocr_vyber_stran.png` – OCR – možnosti výběru stránek.
- `09_digitalni_podpis.png` – Digitální podpis – podepsání a ověření.
- `10_rozdelit_pdf.png` – Rozdělit PDF – hlavní panel.
- `11_rozdelit_pdf_vyber.png` – Rozdělit PDF – možnosti výběru stránek.
- `12_sloucit_pdf.png` – Sloučit PDF – hlavní panel.
- `13_poradi_stranek.png` – Pořadí stránek – úpravy a ukládání.
- `14_cela_aplikace.png` – PDF Blacksmith – pracovní plocha a panely.

**Stav schvalování:** Tato verze nezavádí nové návrhy ani neotevírá nové rozhodnutí Q33. Jde pouze o dokumentační doplnění stávajícího stavu, zachování všech dřívějších schválení a zpřístupnění obrazových referencí pro následnou implementaci.


---

# ROZŠÍŘENÍ ROADMAPY – 21. 11. 2026 (verze 4.0)

**Tato sekce je plán, nikoliv schválení implementace.** Všechny dříve potvrzené požadavky Q1–Q32 a BRAND/UI zůstávají nadále beze změny. U rozšíření se použije princip Q31: před zavedením musí být inventura stávajících funkcí a regresní testy. U nástroje Separace se stále nesmí bez schválení přestavět jeho uspořádání (Q32).

## RP-01: Nový plánovaný nástroj Formuláře – AcroForm

Formuláře doplnit do seznamu budoucích nástrojů jako samostatnou kartu a panel. Etapizovat: **F1 čtení/vyplnění a zachování formulářů → F2 vytváření/úprava polí → F3 bezpečné zploštění a specializovaný export**. Podrobné požadavky, omezení, akceptační testy a návaznosti jsou v přiloženém plánu `PDF_Blacksmith_plan_rozsireni_v1.0.md`.

- Standardní AcroForm (text, více řádků, checkbox, radio, combo/list, tlačítko, podpisové pole). Při práci s PDF respektovat strom `/Fields`, widget anotace, stavy polí a jejich appearance streams.
- Pole umět vybrat, umístit a editovat s přesnou geometrií; pořadí Tab, Required, ReadOnly, jedinečné názvy a vlastnosti zobrazování/tisku. Bez spuštění nevyžádaného PDF JavaScriptu.
- **Export bez poznámek (již schválené chování) nesmí odstranit formulářové widgety, vyplněné hodnoty, jejich podpisy nebo odkazy.** Jde o různé druhy anotací. Při požadovaném „flatten“ samostatné explicitní nastavení a výstupní kontrola, nikoli implicitní odstranění polí.
- XFA: detekovat a přiznat omezení; pikepdf nepodporuje XFA v běžném AcroForm API. Neměnit ani nepřepisovat nepodporovaný formulář, pokud by se ztratila data.
- Integrovat se stávajícím nástrojem Digitální podpis, AutoRecover, Ctrl+Z/Y, „Uložit upravenou kopii…“ a PDF/X exportem. Respektovat ochranu dokumentu a platnost podpisů.

## RP-02 až RP-05 – další plánované moduly (bez automatické implementace)

- **Preflight:** nad současnou dílčí PDF/X/separační kontrolou přidat kontroly písem, obrázků, RGB, přetisků, spadávky, hairlines, omezení TAC a prohlížení konkrétních chyb s výsledným protokolem.
- **Porovnávání PDF:** navázat na existující duální náhled, doplnit synchronizaci posunu/zoomu, overlay, před/po a grafické rozdíly.
- **Měření a geometrie:** délka/úhel/plocha/šířka čar, pravítka, vodítka; respektovat měřítko a rotaci PDF, rozlišit 100% zoom od fyzického rozměru.
- **Dávkové zpracování:** fronta souborů, přenos stávajících výstupních předvoleb, protokoly, bezpečné výstupní kopie.

## RP-06 – SVG a mapa nástrojů

- **Sada v4.2:** 218 SVG (z toho 60 nově připravených pro plánované moduly), `currentColor`, plátno 24 × 24, bez rastrových závislostí.
- **Mapa pro Codex:** `mapa_zapojeni_codex.md`, úplný katalog `ikonova_mapa.md`, strojová `ikonova_mapa.csv`, `manifest.json`.
- **Stavy ikon:** existující funkce ≠ SVG již zapojené v kódu; schválené rozšíření ≠ hotová funkce; plánovaná funkce ≠ schválená implementace; rezerva ≠ příkaz přidat nástroj.
- **Logo:** používat výhradně uživatelem dodané `kovadlina křivky.svg`, případně upravit pouze jeho vínovou barvu, bez změny kresby nebo nahrazování generovaným logem.
- **Zobrazení 1:1:** schválená ikona `zoom_1_1.svg` z v4.1 zůstává i v4.2; chování 100 % podle `dpi/72`, nikoli garantovaná fyzická velikost.

**Technický a časový plán v samostatném souboru je pracovní návrh. Do GitHub repozitáře ani stávajícího kódu se tímto krokem nic nezapisuje.**


---

## Doplnění v4.1: Formuláře – rozpracovaný detailní návrh

**Stav: etapa F1 nyní SCHVÁLENA, zatím neimplementována; F2/F3 zůstávají návrhy.**

Podrobná funkční specifikace, plán F1/F2/F3, vztahy k poznámkám, podpisům, výstupům, pravidla práce s `/AcroForm`, `/Fields`, `/Widget`, appearance streams a 17 akceptačních scénářů jsou v samostatném souboru **`PDF_Blacksmith_Formulare_specifikace_navrh_v1.1.md`**. Přiložený interaktivní testovací soubor `PDF_Blacksmith_Formulare_testovaci_AcroForm_v1.pdf` obsahuje 2 strany a 17 PDF formulářových polí; slouží jako skutečná testovací fixture.

**Pravidlo integrace:** Formuláře jsou novým levým nástrojem, ale nesmějí snížit funkčnost stávajících panelů. Interaktivní formulářová pole nejsou pracovními poznámkami: schválený defaultní export bez poznámek nesmí odstraňovat formuláře ani odkazy. Etapové pořadí F1 → F2 → F3 schváleno dne 21. 11. 2026; kód zatím nezměněn.

**Schválený F-Q1 – varianta A:** nejdříve F1: čtení/vyplňování existujících AcroForm a bezpečné ukládání. F2 (tvorba polí) a F3 (zploštění) zůstávají plánem.


---

## ROZHODNUTÍ FORMULÁŘE F-Q1 – 21. 11. 2026 – SCHVÁLENO (v4.2)

**Uživatel zvolil variantu A: etapový vývoj nástroje Formuláře.** První realizovat **F1 – detekci a vyplňování existujících interaktivních formulářů AcroForm, jejich bezpečné uložení a zachování původního PDF**. Priorita je interoperabilita (formulářová pole a widgety nejsou pracovní komentáře), správa změn s Ctrl+Z/Y, AutoRecover, podpora českých znaků a zachování vzhledových streamů po opětovném otevření. Režim „Uložit upravenou kopii…“ respektuje již schválené zavření původního kontextu a nahrazení aktivní karty novou kopií až po úspěšném zápisu.

**Tvorba a úpravy formulářových polí (F2) a speciální zploštění pro tisk (F3) jsou stále plánované, nikoli tímto schválené k implementaci.** Stávající nástroj Digitální podpis zůstává zdrojem operací skutečného podepisování; F1 pouze rozpoznává podpisová pole. U podepsaných/XFA PDF platí ochrana před ztrátou dat. F1 není prohlášena za již implementovanou.

**Etapový implementační plán / testy:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.1.md`. **Aktualizovaná roadmapa:** `PDF_Blacksmith_plan_rozsireni_v1.2.md`. Všechny původní požadavky a zákaz regresí Q31 zůstávají beze změny. **F-Q2 schváleno (varianta A):** při otevření formuláře zobrazit jen nenápadnou možnost přejít do nástroje Formuláře; žádné automatické přepnutí nástroje.


---

## Formuláře – rozhodnutí F-Q2 (21. 11. 2026, v4.3): pasivní rozpoznání formuláře

**SCHVÁLENO – varianta A:** pokud otevřené PDF obsahuje rozpoznaný vyplnitelný AcroForm, PDF Blacksmith zobrazí neblokující informaci **„Dokument obsahuje formulář – Otevřít Formuláře“**. Zůstane aktivní dosavadní nástroj a žádná jiná karta se nezmění. Kliknutím na nabídku uživatel přejde do **Formuláře → Vyplnit** pro právě dotčený dokument. Přímá detekce ani nabídka nesmějí změnit PDF, historii nebo stav rozpracovaných úprav. Zachovat pravidla přepínání karet, duálního pohledu, textových a objektových výběrů.

**XFA a nepodporované formáty:** nezobrazovat klamnou výzvu k plně podporovanému vyplňování; bezpečně informovat o omezení, nikdy potichu neztrácet data.

**F-Q3 SCHVÁLENO (varianta A s podmínkou):** vyplňování existujících formulářů jen v aktivním režimu **Formuláře → Vyplnit**. Budoucí F2 při tvorbě/editaci polí nabídne typografii a nastavení podobné stávajícímu nástroji Text. Ve F1/Vyplnit se naopak respektuje vzhled definovaný formulářem.

Vlastní **F1 je dosud schválený rozsah, nikoli hotová implementace**. F2 a F3 zůstávají plánem. Podrobnosti: `PDF_Blacksmith_Formulare_specifikace_navrh_v1.2.md`.


---

## FORMULÁŘE F-Q3 – SCHVÁLENO (21. 11. 2026, specifikace v4.4)

**Varianta A platí pod podmínkou plnohodnotného návrhu polí v F2.** Aktivace a vyplňování hodnot existujících formulářových polí je možné pouze při zvoleném **Formuláře → Vyplnit**. V nástrojích Text, Objekty a dalších zůstává jejich dosavadní ovládání, výběry a zkratky. Tím se zabrání nechtěným změnám formuláře při běžné editaci PDF.

**F1 / Vyplnit:** vzhled písma, velikost, barvu, zarovnání a geometrii určuje autor existujícího formuláře; uživatel mění hodnotu, ne design pole. **F2 / Návrh a vlastnosti polí:** umožnit odpovídající typografické a editační nastavení jako v aktuálním nástroji Text – výběr fontu a řezu, velikost, barvu, zarovnání, víceřádkové vlastnosti, pozici, rozměry, povinné/jen ke čtení, nápovědu a správné PDF appearance streams. F2 je stále naplánovaná, nikoli aktuálně autorizovaná k implementaci.

**Podrobnosti a nové akceptační testy T-18 až T-20:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.3.md`.

## PLÁNOVANÉ VYLEPŠENÍ VYHLEDÁVÁNÍ – po Formulářích, detaily zatím otevřené

Uživatel požaduje rozšíření stávajícího nástroje „Najít v PDF“ bez ztráty dnešních funkcí: **okamžitě vyhledávat při psaní**, **označit všechny výskyty šipkou**, **odstranit všechny šipky**, **individuálně odstranit šipku jejím křížkem ×**, **uložit označené výskyty včetně šipek** a **připojit k jednotlivým šipkám textové poznámky**. Trvalé označení vhodně řešit interoperabilními PDF anotacemi navázanými na schválený nástroj Poznámky; nesmí se stát trvalou změnou původní vektorové kresby. Respektovat výchozí tisk/export **bez pracovních poznámek**, možnost jejich explicitního zahrnutí, AutoRecover a Ctrl+Z/Y. Přesné chování při ukládání a mazání a UX bude projednáno **po dokončení návrhu Formulářů**.

**Stav:** tyto požadavky jsou výslovně zařazeny do plánu; **S-Q1 až S-Q5 čekají na budoucí schválení**, samotná funkce není implementována ani tímto dokumentem objednána k okamžitému kódování. Viz `PDF_Blacksmith_Vyhledavani_navrh_k_upresneni_v0.1.md` a `PDF_Blacksmith_plan_rozsireni_v1.4.md`.


---

## FORMULÁŘE F-Q4 – SCHVÁLENO (21. 11. 2026), VARIANTA A

V nástroji **Formuláře → Vyplnit** budou dostupná, viditelná a editovatelná formulářová pole **jemně barevně zvýrazněna**, ve výchozím stavu zapnuto. V pravém panelu přepínač **„Zvýraznit vyplnitelná pole“** pro okamžité skrytí/zobrazení této pomůcky, bez omezení možnosti vyplňovat pole. Mimo aktivní režim Formuláře → Vyplnit se pomocná barva **nezobrazuje**; pole ReadOnly/skrytá nejsou falešně označována za vyplnitelná.

Zvýraznění musí být **pouze obrazovková, dočasná UI vrstva**: nemění obsah PDF, anotace, tisk, export ani stav rozpracovaných úprav. Fokus se označuje samostatným zřetelnějším obrysem. Respektovat zoom, rotaci, DPI, změny stránky a duální zobrazení. Viz F1 specifikace `PDF_Blacksmith_Formulare_specifikace_navrh_v1.4.md` (F-Q4, T-21 až T-25). Předchozí schválení F-Q1 až F-Q3 zůstávají beze změny.


---

## FORMULÁŘE F-Q5 – SCHVÁLENO (21. 11. 2026)

**Varianta A:** Pravý panel **Formuláře → Vyplnit** obsahuje **sbalitelný přehled polí aktivního dokumentu**, jejich stav (vyplněno, povinné, pouze ke čtení) a navigaci kliknutím na konkrétní pole na stránce. Výběr na stránce a řádek v seznamu se synchronizují. Sbalení je pouze změna zobrazení, nesmí nastavit dokument jako změněný. V duálním zobrazení ani mezi kartami se nesmí přepnout nebo změnit jiný dokument. U logických polí s více widgety nenásobit hodnoty ani nesprávně nevyhodnocovat stav.

**F-Q6 – SCHVÁLENO, varianta A:** přehled řadit podle stránek a vizuální polohy; Tab order zachovat nezávislé. F1 pokračuje postupně; implementaci F2/F3 toto schválení nepovoluje. **Zachovat** všechny dříve schválené funkce, včetně F-Q1–F-Q4, funkce Text a hledání. Podrobnosti a testy T-26–T-30: `PDF_Blacksmith_Formulare_specifikace_navrh_v1.5.md`; roadmapa `PDF_Blacksmith_plan_rozsireni_v1.6.md`.


---

## FORMULÁŘE F-Q6 – SCHVÁLENO (21. 11. 2026), VARIANTA A

Přehled formulářových polí v pravém panelu **Formuláře → Vyplnit** se řadí podle **aktuálního pořadí stránek a vizuálního umístění polí v rámci stránky** (typicky shora dolů a ve stejném řádku zleva doprava). **Navigace Tab/Shift+Tab zůstává nezávislá a respektuje formulářové pořadí/standardní chování PDF**. Řazení seznamu nesmí přepisovat pořadí widgetů, logické hodnoty polí ani původní PDF, a nemá samo vytvářet položku historie či neuloženou změnu. Při přeskupení stránek se přehled obnoví, u opakovaných widgetů nerozmnoží logická data a u skrytých či nepřístupných prvků uvede odpovídající stav. Zachovat správnou orientaci a stabilitu pořadí při zoomu, DPI a otočení.

**Konkrétní pravidla a testy T-31 až T-35:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.6.md`. **Aktualizovaný plán:** `PDF_Blacksmith_plan_rozsireni_v1.7.md`.

**F-Q7 je již SCHVÁLENO (varianta A):** varovat před ručním uložením neúplného formuláře, umožnit navigaci na chybějící pole a volbu „Uložit přesto“. F2 a F3 se nezahajují bez dalšího schválení; F1 není v projektu hotová pouhým schválením tohoto rozhodnutí. Zachovat F-Q1–F-Q5 a dříve odsouhlasené funkce aplikace.


---

## FORMULÁŘE F-Q7 – SCHVÁLENO (21. 11. 2026), VARIANTA A

Při **explicitním ukládání** existujícího formuláře AcroForm v nástroji **Formuláře → Vyplnit** má aplikace **upozornit na nevyplněná povinná pole**, umožnit uživateli na ně přejít v již schváleném sbalitelném přehledu pravého panelu a také **uložit rozpracovaný formulář i bez jejich dokončení**. Uživatel musí mít srozumitelné možnosti „Přejít na chybějící pole“, „Uložit přesto“ a „Zrušit“; žádná možnost nesmí sama vyplnit či smazat hodnoty. Formulářový stav a hodnoty se zachovají i po opětovném otevření dokumentu.

Rozpoznávání skutečné **povinnosti** vychází z příznaků AcroForm `/Ff`; opakované widgety jednoho logického pole se nepočítají víckrát. Formuláře s kombinací prázdných/ReadOnly/skrytých polí mají mít pochopitelné chybové stavy. **AutoRecover je neblokující** a musí zachovávat částečně vyplněná data; nedotýkat se tím standardních klávesových zkratek ani stavu nástroje Text.

**Povolení uložení neúplného formuláře nepřebíjí bezpečnostní překážky:** ochranu podepsaných PDF, XFA, oprávnění dokumentu, požadavek na bezpečný zápis, ochranu původního souboru a již schválené chování „Uložit upravenou kopii…“ (kopie ve stejné kartě až po ověřeném úspěchu). Nezaměňovat s případným budoucím finálním odesláním formuláře.

**Podrobná pravidla a nové testy T-36 až T-41:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.7.md`. **Plán:** `PDF_Blacksmith_plan_rozsireni_v1.8.md`. Rozhodnutí F-Q1–F-Q6 i plány rozšíření vyhledávání zůstávají beze změny. Implementace nástroje F1 v repozitáři tímto dokumentem nebyla provedena; F2/F3 dál čekají na samostatné schválení.


---

## Formuláře F-Q8 – SCHVÁLENO (21. 11. 2026, varianta A)

Při **Formuláře → Vyplnit** průběžně aktualizovat stav vyplnění v přehledu polí (F-Q5); **chybový formát zobrazit až po opuštění pole**, nikoli rušivě během každého úhozu. Při ručním **Ctrl+S / Uložit / Uložit upravenou kopii…** znovu zkontrolovat podporovaná pravidla a ukázat **souhrn chybného formátu, nevyplněných povinných polí a neověřitelných pravidel**. Nabídnout **Přejít na pole, Uložit přesto, Zrušit**; volba Uložit přesto zachová rozpracovaný stav podle F-Q7, ale **nepřekračuje oprávnění PDF ani bezpečnost podpisů**. Žádné neprověřené PDF JavaScript validátory. Neodvozovat formát e-mailu/datu pouhým odhadem z názvu pole. Kontrola sama o sobě nesmí nastavovat neuloženou změnu; AutoRecover nesmí otevírat modální dialog. Zabránit duplicitám u opakovaných widgetů, zachovat Tab order dle F-Q6 a kompatibilitu s Text/Objekty/OCR/poznámkami. F2/F3 zůstávají v plánu.

**Podrobná specifikace a akceptační testy T-42 až T-49:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.8.md`. **Roadmapa:** `PDF_Blacksmith_plan_rozsireni_v1.9.md`.


---

## FORMULÁŘE F-Q9 – SCHVÁLENO (21. 11. 2026), VARIANTA A

V nástroji **Formuláře → Vyplnit** po opuštění pole s **bezpečně ověřenou chybou formátu** zobrazit **jemný červený obrys kolem pole na stránce** a zároveň **chybový stav v pravém sbalitelném přehledu polí** (F-Q5). Během rozepsání nedokončené hodnoty neobtěžovat (F-Q8). Po opravě hodnoty a opětovné validaci chybu z obou míst odstranit. Nevyplněné povinné pole je samostatný stav, neověřitelný PDF JavaScript není potvrzenou chybou.

Červený obrys je **pouze obrazovková vrstva**, neukládá se do PDF, netiskne ani neexportuje a nesmí měnit stav rozpracovaného dokumentu. Rozlišovat od vínového zvýraznění dostupných polí (F-Q4): jeho vypínací přepínač nezakrývá ověřené chyby. Při přepnutí na jiný nástroj se obrys skryje, ale interní chybový stav je k dispozici po návratu. Zajistit srozumitelný text/symbol chyby pro přístupnost, zoom, DPI, otočení a duální zobrazení. F-Q7/F-Q8 i možnost **„Uložit přesto“** zůstávají nedotčené, AutoRecover neblokovat.

**Podrobný postup a akceptační testy T-50 až T-57:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.9.md`. **Plán:** `PDF_Blacksmith_plan_rozsireni_v2.0.md`. F2/F3 jsou dál plánované, kód aplikace tímto textem změněn nebyl.

---

## Aktuální návrhový dodatek – UI v3.9 (24. 11. 2026)

**Aktuální pracovní HTML:** [rozhrani/cely_vizual_v3.9.html](rozhrani/cely_vizual_v3.9.html).  
**Úplné schválené dodatky UI v3.4–v3.9:** [rozhrani/AKTUALNI_UI_CODEX_v3.9.md](rozhrani/AKTUALNI_UI_CODEX_v3.9.md).  
**Mapa ikon:** [mapa_ikon_v4.5.md](mapa_ikon_v4.5.md), [SVG atlas](rozhrani/ikony/pdf_blacksmith_nove_v3.9.svg).

Doplněk je závazný **pro návrh rozhraní**, nevyžaduje okamžitou implementaci v Go/Win32/PDFium. Platí původní bezpečnostní a funkční omezení tohoto souboru (SEL/TXT/OBJ, formuláře F1, podpisy, uložit/export, undo/redo). Codex má před úpravami backendu vždy provést inventuru skutečného kódu, návrh etapy a požádat o výslovné schválení.

**Shrnutí nejnovějších změn:** konzistentní dynamická nápověda pod ikonami v pevném prostoru a plovoucí tooltip nad ikonami; žádný diagonální styl typografických tlačítek; adaptivní druhá lišta při min. návrhových 980×540 CSS px a 30% přiblížení PDF; nepřekrývat dokument pravými panely. Univerzální výběr šipkou, obdélníkem a lasem, automatická kontextová aktivace Textu/Objektů/OCR a ochrana smíšeného výběru; samostatný odstavcový textový blok; textové převody v Text → Výběr a sdílené OCR pro vektory, rastrové oblasti i kombinace. Dialog rekonstrukce textu s porovnáním a korekcemi, základní úpravy vybraných rastrových obrázků (pixely/PPI, resampling, grayscale/ICC/CMYK, ořez), inteligentní magnet, mřížka, pravítka a plánovaný SnapEngine se samostatnými cíli v Zobrazení → Přichytit.

**V3.9 nahrazuje starší split-tlačítko kurzoru a vrací druhý horní přepínač zobrazení:** dvoustavová ikona šipka/textový kurzor a dvoustavová ikona Po stránkách/Souvislé zobrazení zobrazují vždy *příští akci*, po kliknutí se stav i symbol vystřídají. Obě volby využívají již existující sdílené stavové a dokumentové režimy bez druhé kopie handlerů. Kontrolní scénáře a přesné výjimky jsou v aktuálním pracovnímu doplňku v3.9.


---

## Návrhový dodatek UI v3.11 (24. 11. 2026)

- Separace → Výstup a konverze: ikony svisle ve třech řadách; každý řádek má vedle ikony výběrové okno s předvolbami převodu/ICC. Nepřenášet UI volbu bez kontroly do exportu PDF.
- Dočasně automaticky otevřený panel Výsledky vyhledávání se při opuštění hledání obnoví na **původní ručně nastavený pomocný panel** (nebo žádný); ruční změny během hledání mají přednost.
- Pevná vodítka jsou **plná**, přesahují hranice stránky k pravítkům; inteligentní přichytávací linky jsou **vínově červené, čárkované/tečkované**, dočasné a mají končit v cíli nebo na pravítkách u hran.
- Mřížka má nastavitelnou rozteč a jemné dělení; zobrazení mřížky, přichytávání k mřížce a globální magnet jsou oddělené stavy. Magnet je vínově vyplněná podkova s **obrysovými koncovými póly**.
- Konkrétní stávající zkratky a jejich konflikty ověřeny ve zdrojích: `document_commands_windows.go`, `workspace_windows.go`, `viewer_windows.go`. Nové zkratky jsou **jen návrhy** v [mapě v1.0](mapa_klavesovych_zkratek_v1.0.md); bez schválené migrace neměnit současné `T/V/M/L`.
- Změna nastavení viditelnosti/rozvržení nemění PDF, nemá vytvářet Undo krok ani dirty; návrh SnapEngine implementovat až po zvláštním souhlasu.


---

## Doplňující pravidlo UI v3.13 (24. 11. 2026)

- Pravítka a jejich popisky se kreslí mimo stránku PDF. Vodítka přecházejí od objektů přes okraj stránky ke stupnicím. Pomůcky rozvržení mají vlastní položku pod Zobrazení, vlastní panel vlastností a společný stav s nabídkou Zobrazení a horní lištou.
- Nastavení → Klávesové zkratky obsahuje tlačítko Vlastní zkratka u jednotlivých akcí, režim zachycení, reset, upozornění a **odmítnutí** konfliktní kombinace v překrývajícím se kontextu. Neztratit stávající význam zkratek a nevázat nové klávesy přímo na produktové handlery bez explicitního souhlasu.
- [Plný návrh a testy](rozhrani/UI_POMUCKY_ZKRATKY_v3.13.md), [mapa zkratek](mapa_klavesovych_zkratek_v1.1.md). Před implementací do Go/Win32 nejprve analyzovat dispatcher, fokus, český AltGr a kontexty.
