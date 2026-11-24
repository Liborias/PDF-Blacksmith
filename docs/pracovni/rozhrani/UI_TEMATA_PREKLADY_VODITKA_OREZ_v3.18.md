# PDF Blacksmith — UI v3.18: témata, tabulka překladů, skupiny vodítek a ořez

**Datum:** 24. 11. 2026 · **Typ:** schválená iterace HTML návrhu a podklady pro Codex. **Nejde o implementaci do Go/Win32/PDFium.**  
**HTML:** [cely_vizual_v3.18.html](cely_vizual_v3.18.html).  
**Překlady:** [i18n/ui_catalog_v1.json](../i18n/ui_catalog_v1.json) + [CSV](../i18n/ui_catalog_v1.csv).  
**Předchozí specifikace:** [UI v3.17](UI_PREKLADY_SKUPINY_VODITEK_v3.17.md), [hlavní Codex v5.0](../specifikace_Codex_v5.0.md), [mapa zkratek v1.3](../mapa_klavesovych_zkratek_v1.3.md), [mapa ikon v5.1](../mapa_ikon_v5.1.md).

## 1. Sbalitelné skupiny a kompaktní seznam vodítek

- Skupina má na začátku řádku rozbalovací/zavírací šipku. Sbalení **skryje pouze členy seznamu**, nemění viditelnost vodítek na stránce ani jejich snap stav.
- Jednotlivé členy zobrazit **mírně odsazené doprava** od názvu skupiny. Samostatná vodítka jsou na stejné úrovni jako skupiny, bez nadpisu „Samostatná vodítka“.
- Existující checkboxy, oko, barvu, zámek a koš zachovat. Rozbalená/sbalená skupina má trvalý stav alespoň po dobu práce v HTML; v produkční aplikaci podle dokumentového profilu.
- Odstranění vodítka **všude** znamená ikona koše; křížek vyhradit pro zavření okna/karty nebo zrušení rozepsané operace. Zámek a oko mají konzistentní ikony a popisky pouze na hover.
- Při výběru řádku či čáry bez modifikátoru vybrat právě jedno vodítko, **Ctrl+levý klik** přidává/ubírá další. Checkboxy používají **stejný model vícenásobného výběru**. Kliknutí mimo odznačí a zavře dočasný panel Pomůcek dle v3.16.

## 2. Bezkolizní gesta Ctrl + klik versus Ctrl + tažení

- **Ctrl+klik na vodítku** přepíná členství ve výběru, analogicky k objektům. Rozhoduje dokončený krátký klik; neprovádět během \`pointerdown\` nevratnou modifikaci výběru.
- **Ctrl + tažení levým tlačítkem** (> 7 CSS px a aktivní manipulace) **duplikuje** konkrétní vodítko. Originál nezměnit. Kopie přebírá osu, barvu a uživatelské parametry, dostává **nové ID**, není automaticky členem původní skupiny. Výsledek je vybraná kopie, umožňující pokračovat v úpravě.
- Tažení bez Ctrl posouvá originální neuzamčené vodítko. Samotný Ctrl+klik nesmí kopírovat. Pro odemčené/uzamčené vodítko má klávesová kombinace jednotnou selekční logiku; uzamčené se neduplikuje.
- Esc/pointercancel ruší rozpracované tažení bez kopie. Přichytávání duplikátu používá stejný SnapEngine jako přesouvané vodítko.
- Hit-test vodítka má přednost **pouze při zasažení vodítka**; veškerá ostatní Ctrl+výběrová gesta textu, objektů a stran zachovat. Nezavádět paralelní výběrové handlery.

## 3. Barvy a potvrzování

- Výběr barvy v color pickeru **pouze připraví hodnotu**. Skutečné použití na jednotlivé, označené, skupinu nebo výchozí barvu nových vodítek se provede až po kliknutí **Použít** (v úsporném řádku může být ✓ s tooltipem).
- Odlišovat: individuální \`guide.color\`, barvu více označených, skupinovou operaci a výchozí \`viewAids.guideColor\` pro **nová** vodítka. Samotné zvolení nové výchozí barvy nepřebarvuje již existující linky.
- Nastavení barev rozhraní má rovněž potvrzovací tlačítko u **vlastního** akcentu a plochy panelů. Přednastavené palety se v dialogu zobrazují jako okamžitý **náhled**; \`Zrušit\` obnoví uložené téma, \`OK/Použít\` uloží do nastavení (v HTML do paměti).
- Nepovažovat zobrazení náhledu ani změnu uživatelských pomůcek za PDF editaci nebo za operaci Undo.
- Neprodlužovat nechtěně horní lišty při potvrzovacích tlačítkách; delší popisy dát do tooltipů.

## 4. Uživatelská témata + světlý režim

- Pod **Nastavení → Rozhraní** nabídnout tmavý/světlý režim, čtyři předvolby akcentů: vínovou, **neutrální světle šedou**, modrou a zelenou, plus **Vlastní**. Volitelná druhá vlastní barva může upravit plochy panelů.
- Barva akcentu a světlost jsou **nezávislé dimenze**. Světlý režim má čitelné tmavé texty, světlé panely, kontrastní zvýraznění, upravené okraje a čitelné vstupy. Kontrolovat WCAG kontrast a nenechat bílé/pastelové popisky na bílém povrchu.
- Nepřebarvovat skutečný obsah PDF, barevnou separaci, formulářové hodnoty, náhledy obrázků nebo výstupní ICC podle motivu editoru.
- Ikona \`i-nav_settings\` má **větší a symetrický** obrys ozubeného kolečka. Zachovat identifikátor, změnit pouze \`<symbol>\`, ostatní SVG id neměnit.
- Produkovaný Go/Win32 kód musí oddělit základní design tokeny (base/top/panel/text/border/accent) od barev dokumentů; nevyužívat pouze CSS experimenty z HTML.

## 5. Ořez vybraného rastru

- V **Objekty → Úpravy obrázku** čtyři základní režimy: **Obdélník, Čtverec, Ovál, Kruh**. Pro jiné obdélníkové poměry jedno rozbalovací menu: **Volný, 1:1, 4:3, 3:4, 16:9, 9:16, 3:2, 2:3, 5:4, 4:5, 21:9**.
- Čtverec a kruh zamykají poměr 1:1; ovál zachovává eliptický tvar s možným jiným poměrem os. Volba je demonstrací ořezové masky v náhledu, nikoli skutečným nezvratným přepsáním původního rastru.
- Při reálné implementaci rozhodnout mezi vektorovým clipping path pro elipsu/kruh a destruktivním přerenderováním. Preferovat **nedestruktivní masku** a možnost obnovy původní plochy, včetně Undo/Redo a exportu výsledného PDF.
- Resampling, ICC, gray/CMYK a vlastní rozměry zůstávají oddělené funkce a nemají se samovolně provést při změně ořezu.

## 6. Samostatný zdroj překladů a další jazyky

- Zdroj se nachází v [\`../i18n/ui_catalog_v1.json\`](../i18n/ui_catalog_v1.json) a přehledné [CSV](../i18n/ui_catalog_v1.csv). Základní tabulka má **783 položek** \`id, cs-CZ, en-US, context, status\`.
- ID jsou pro prvotní migraci unikátní a konzistentní, odvozená z původní české formulace. **Po jejich převzetí v aplikaci je zmrazit**; při další editaci češtiny neregenerovat, jinak se ztratí vazba do všech jazyků. Následně přejít na sémantické ID podle funkce (\`toolbar.snap.enabled\` apod.).
- Pro další jazyk přidat nový sloupec do CSV a odpovídající locale hodnotu do JSON; implementace musí kontrolovat chybějící klíče, proměnné/placeholders, plurály, přístupné popisky a fallback.
- **Nikdy nepřekládat** uživatelská data, obsah PDF, názvy souborů, dokumentové záložky ani jazyk OCR. \`UI_TRANSLATIONS_EN\` zůstává uvnitř offline HTML jako demonstrační slovník, **produkční aplikace má čerpat z externího katalogu**, nikoliv DOM heuristik.
- Katalog zatím **není automaticky načítán** zvenku do offline HTML; to umožňuje nadále otevřít samostatný soubor bez serveru.

## 7. Testy a limity

- HTML v Chromiu při 980×540, 1100×720, 1250×820 a 1600×900 px: výběr, Ctrl+klik, skupiny, sbalení, dvě barevná potvrzení, light/dark a neutral accent, ořez kruh/4:3. Bez JS výjimek v testovaných scénářích.
- Extra reálná myší simulace **Ctrl+tažení** na vodítku: původní X 44 mm zůstalo, kopie s novým ID vznikla poblíž 56 mm, nezařadila se do skupiny. Samotný Ctrl+klik nezaložil kopii.
- Vizuální test odhalil a opravil nízký kontrast v Nastavení ve světlém režimu.
- **Neproběhl** test produkční Go/Win32 aplikace, ani skutečný ořez/editace PDF, ani skutečný SnapEngine pro obecné PDF objekty.

**Pro Codex:** aktualizace UI v3.18 **není** souhlas s implementací F2/F3 a neanuluje dříve schválené chování SEL/TXT/OBJ/F1. Před produkční změnou nejprve inventarizace handlerů, datových struktur, možných kolizí a akceptačních testů.
