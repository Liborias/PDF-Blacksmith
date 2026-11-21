# PDF Blacksmith – Vyhledávání: funkční rozšíření k upřesnění po Formulářích (v0.1)

**Stav:** uživatelsky požadované směry funkce; detaily UI, ukládání, výkonu a interoperability **zatím čekají na samostatné projití a odsouhlasení**.  
**Datum:** 21. 11. 2026.  
**Priorita v roadmapě:** po rozpracování Formulářů, před zahájením zcela nových modulů, podle kapacity.

## 1. Co již existuje – zachovat

Současný nástroj „Najít v PDF“ prohledává textovou vrstvu, odlišuje aktivní výskyt (vínově) a ostatní nálezy (žlutě), nabízí seznam výskytů, přechod mezi výsledky přes F3/Shift+F3 a Ctrl+F a rozlišuje diakritiku. Podle dosavadního návodu se nyní dotaz potvrzuje Enterem. **Nevytvářet druhý vyhledávač a nezrušit stávající navigaci.** U PDF bez textové vrstvy zůstává pro rozpoznání potřeba OCR. Současný limit viditelného seznamu 2 000 výsledků vyžaduje uvážené řešení, nemá se tiše zaměňovat za limit všech nalezených výskytů.

## 2. Uživatelské požadavky – založit po Formulářích

1. **Okamžité vyhledávání při psaní:** při změně dotazu spustit vyhledávání bez nutnosti stisku Enter. Doporučený technický návrh: krátký debounce, zrušení neaktuálních výsledků a postupné zpracování rozsáhlých PDF bez blokování UI; přesnou prodlevu a zobrazení průběhu dohodnout později.
2. **„Označit všechny výskyty šipkou“:** samostatná akce vykreslí u každého nalezeného výskytu **šipku směřující na konkrétní výskyt**. Nesmí změnit původní text ani znemožnit stávající podbarvení nalezených výrazů. U vícesloupcové sazby, rotovaných textů, zdvojených řádků a vícenásobných výskytů na stejné straně mají šipky jednoznačně ukazovat na správný nález.
3. **„Odstranit šipky“:** hromadné zrušení označení vytvořených danou akcí. Nevztahovat bez výslovného příkazu na jiné existující komentáře ani originální grafiku dokumentu.
4. **Křížek `×` na každé šipce:** každý jednotlivý značkovač má vlastní malý křížek pro jeho zavření/odstranění. Odebrání značky neodstraňuje nalezený text ani nevymazává ostatní výsledky hledání. Zvažovat zobrazení křížku jen při hoveru/výběru, aby nezahltil stránku.
5. **Uložit výskyty i se šipkami:** uživatel může uložit označený dokument se šipkami tak, aby označení zůstala i po zavření a znovuotevření PDF. **Doporučená technologie k posouzení:** standardní PDF anotace (čára/šipka + komentář), nikoli přimalování přímo do původní vektorové kresby. V průběhu kontroly se může používat dočasná UI vrstva; před uložením nabídnout explicitní vytvoření trvalých anotací.
6. **Textová poznámka ke každé šipce:** po vybrání šipky může uživatel otevřít/zapsat její komentář a později jej upravit. Zachovávat vztah poznámky k danému výskytu; po uložení použít interoperabilní PDF struktury, pokud to zvolený model anotací dovolí.

## 3. Pravidla napojení na již schválené funkce

- Zachovat existující vyhledávání, navigaci výsledky, běžný výběr textu, objekty a OCR (Q31).
- **Poznámky:** trvalé šipky s komentářem chápat jako pracovní PDF anotace v rámci navrženého modulu Poznámky. Ukládání pracovního PDF je uchovává, zatímco **výchozí tisk a export bez poznámek zůstávají čisté**; pro konzultační PDF uživatel explicitně zapne zahrnutí anotací. Nevytvářet druhý nekompatibilní formát komentářů.
- **Historie a AutoRecover:** přidání/odebrání uložitelných šipek a úprava jejich poznámek jsou vratné obsahové operace (`Ctrl+Z/Y`), jež mohou znamenat neuloženou změnu. Samotný hledaný výraz a dočasné zvýraznění nemají měnit PDF ani příznak změny.
- **Soukromí a interoperabilita:** bez automatického ukládání hledaných citlivých výrazů mimo dokument; pokud jsou vepsány do komentářů, jsou pak součástí PDF, což musí být jasné. Při opravách a převodech zachovat digitální podpisy a upozornit na jejich možnou invalidaci; při práci s chráněnými soubory respektovat povolené operace.
- Šipky musí reagovat na zoom, DPI a rotaci a zůstat přiřazeny ke konkrétnímu nálezu; není vhodné ukládat pouze obrazovkové pixelové souřadnice.
- Po změně dotazu nevytvářet další duplicitní uložené šipky automaticky.

## 4. Otevřené otázky – vrátit se k nim po Formulářích

**S-Q1:** mají být šipky pouze dočasné až do zvláštního příkazu „Převést na poznámky“, anebo okamžitě vytvářet trvalé (zatím neuložené) PDF anotace?  
**S-Q2:** jak se chovat při stovkách/tisících výskytů – přehledné dávkování, označit jen vybrané strany nebo opravdu všechny?  
**S-Q3:** poloha, velikost, čitelnost a barva šipek a režim křížku; možnost skrýt celou vrstvu označení.  
**S-Q4:** zda ukládat s šipkami také hledanou frázi do textu anotace, nebo jen volitelný vlastní komentář.  
**S-Q5:** chování již uložených šipek při změně původního textu, OCR a přesunu stránek, aby nezůstaly přiřazeny k nesprávnému místu.

**Stav:** uživatelské požadavky na vyhledávání jsou zaznamenány, konkrétní rozhodnutí S-Q1–S-Q5 zůstávají otevřená. Nepřestavovat vyhledávání automaticky bez jejich postupného odsouhlasení.