# PDF Blacksmith – UI v3.19: plnohodnotné motivy, kontrast a dynamické překlady

**Datum:** 24. 11. 2026 · **Stav:** závazná aktualizace HTML návrhu a dokumentace pro Codex; **nejde o povolení automatické implementace Go/Win32/PDFium**.

**HTML:** [cely_vizual_v3.19.html](cely_vizual_v3.19.html)  
**Mapa barev:** [mapa_barev_v1.0.md](../mapa_barev_v1.0.md) · [JSON hodnoty a testy](../mapa_barev_v1.0.json)  
**Lokalizace:** [i18n/ui_catalog_v1.1.json](../i18n/ui_catalog_v1.1.json) · [CSV](../i18n/ui_catalog_v1.1.csv)  
**Předchozí stav:** [v3.18](UI_TEMATA_PREKLADY_VODITKA_OREZ_v3.18.md), [hlavní specifikace v5.0](../specifikace_Codex_v5.0.md).

## 1. Kompletní barevná schémata

V `Nastavení → Rozhraní` (společně se světle/tmavým režimem) je **šest plných palet**:
1. **Šedo-vínová** – výchozí, grafit s vínovým akcentem.
2. **Celošedá** – neutrální achromatická, bez vínové.
3. **Krémová** – teplá písková/karamelová, bez vínové.
4. **Mořská modrá** – chladná šedomodrá, bez vínové.
5. **Olivově zelená** – přírodní tlumená zelená, bez vínové.
6. **Petrolejová** – modrozelená/břidlicová, bez vínové.

Každá má vlastní úplné **světlé a tmavé** tokeny: základní plocha, horní lišty, panely, vnořené povrchy, tlačítka, hover, selected, ouška, vstupy, text, sekundární text, obrysy, akcentní prvky i fokus. Vlastní barvy zůstávají dodatečnou možností. Nezdvojovat dřívější `Nastavení → Obecné → Vzhled aplikace`.

**Světlý režim** má jednoznačně světlá ouška i všechny běžné formulářové a editační prvky, včetně rozbalovacích nabídek, tlačítek, statistik a stavových štítků. Neponechávat pevné tmavé barvy `#26262f` nebo podobné z původního CSS. Pozadí/foreground se musí řídit příslušnými tokeny; celé schéma není pouhý accent swap.

Výchozí autorovo SVG magnetu a ostatní ikony zůstávají zachovány; použít `currentColor`. Ozubené kolečko Nastavení má **stejnou výšku 20 px jako disketa vedle něj**. Vlastní zvolená barva se použije až po potvrzení tlačítkem **Použít**; Zrušit zruší náhled.

## 2. Kontrast

V [JSON mapě](../mapa_barev_v1.0.json) je **96 kontrolovaných kombinací** šesti palet, dvou režimů a osmi dvojic foreground/background. Všechny mají kontrast alespoň **4,5 : 1** (minimum 4,58 : 1). Pro velké texty a rozlišovací grafické prvky je minimem **3 : 1**.

Codex má **automaticky testovat vypočtené styly skutečných komponent**, nejen tokeny: výchozí/hover/vybrané/disabled tlačítko, příslušné ucho pravého panelu, input/select, popisek, formulářovou statistiku, nápovědu, upozornění i klávesový fokus. Hlášení chyby/varování/úspěchu používají sémantickou barvu nezávislou na dekorativním akcentu. U vlastních barev nabídnout bezpečnou korekci kontrastu nebo výslovné upozornění.

Tematizace nesmí měnit obsah otevřeného PDF, originální barevné prostory, ICC/separace, tisk, uživatelské barvy vodítek, podpisy, dirty flag ani historii Undo. Výchozí barvu dočasných inteligentních vodítek lze aktualizovat podle palety, ale explicitní uživatelskou volbu zachovat.

## 3. Dynamické překlady a data dokumentu

V anglickém režimu musí být kompletně přeložen i dynamický pravý panel Formuláře:

- `Přehled polí` → `Field overview`;
- `Stránka 1 (8 polí)` → `Page 1 (8 fields)`;
- `Zkontrolovat` → `Validate`;
- `Neplatný formát e-mailu` → `Invalid e-mail format`;
- `Povinné pole není vyplněno` → `Required field is empty`;
- další dynamické validace, počty a akční texty sestavovat ze šablon se zástupnými parametry.

**Názvy polí** `Jméno a příjmení`, `Datum podání` apod. v ukázkovém PDF jsou **metadata a obsah dokumentu**, proto je nepřekládat při změně jazyka aplikace. Stejně se zachovávají hodnoty polí, název souboru a náhled miniatury. Tyto záměrné české texty nejsou chybou lokalizace UI.

Nový katalog `i18n/ui_catalog_v1.1` má 809 překladových položek a samostatné šablony `dynamic_messages`. Produkční lokalizace musí používat stabilní sémantické klíče (např. `form.validation.issueCount`) namísto dohledávání přes doslovné české řetězce. Další jazyky se přidávají do téhož katalogu, nikdy jako oddělená kopie rozhraní.

## 4. Zpětná kompatibilita a testy

- Zachovat všechna schválená pravidla v3.9–v3.18: výběr, text, OCR, Separace, ořez, skupiny vodítek, zámky, `Ctrl + klik`, `Ctrl + tažení`, vlastní zkratky, bez kolizí v SEL/TXT/OBJ/F1.
- Ověřit 6 × 2 kombinací vzhledu při zobrazení formulářů, nastavení a Pomůcek rozvržení. Ve světlém režimu nezůstane žádné tmavé ouško, pole nebo box s tmavým nečitelným textem.
- V angličtině projít všechny hlavní nástroje a dynamický formulářový panel; záměrně nepřekládat PDF data.
- HTML bylo testováno v Chromiu, bez zachycených JS výjimek, a základní kontrola barevných párů prošla. **Go aplikace a PDFium se v této etapě neměnily.**

**Priorita:** v3.19 mění témata, kontrast, motivová nastavení a dynamické překlady; nemění předchozí schválenou logiku ostatních nástrojů. Před produkční implementací je nutný samostatný schválený plán a regresní sada.
