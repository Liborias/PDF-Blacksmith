# PDF Blacksmith – Text v3.2: společná horní řada výběru a skrytá nápověda

**Datum:** 22. listopadu 2026  
**Stav:** schválené vizuální upřesnění pracovní HTML demonstrace. **Není implementována editace PDF v Go/Win32.**  
**Navazuje na:** `TEXT_VYBER_A_ZALOZKY_v3.1.md`, `specifikace_Codex_v5.0.md` (SEL-01 až SEL-03, TXT-01 až TXT-04, OBJ-01 až OBJ-03, Q9).

## 1. Záložka Text → Výběr: jeden řádek

V jediném horním řádku skupiny **Výběr a úpravy** jsou zleva doprava:

1. **Textový kurzor** (`text_caret`) – přímá editace, kliknutí umístí caret do existujícího textu.
2. **Nový textový blok** (`text_box`) – nový bodový/odstavcový text.
3. **Obdélníkový výběr** (`rect_select`) – ikona přesunuta ze spodní rozbalovací sekce přímo k hlavním volbám.
4. **Laso** (`lasso`) – rovněž přesunuto ze spodní sekce.
5. **Nápověda (?)** (`nav_help`) – **samostatná ikona zarovnaná k pravému okraji panelu**, ve **stejném řádku** jako čtyři režimy.

Nezavádět další „spodní“ skupinu jen pro obdélník a laso. Všechny ikony musí mít vlastní přesný tooltip a `aria-label` a nesmí se při změně šířky panelu překrývat. Použít dosavadní schválenou sadu SVG, ne Unicode/emodži místo ikon.

## 2. Nápověda – skrytá za otazníkem

- **Ve výchozím stavu** jsou rozsáhlé tabulky výběrových gest, podrobný text k režimu Text a pokročilé zkratky **skryté**.
- Kliknutí na **?** tyto informace rozbalí **pod horní řadou** v rámci pravého panelu Text; opětovné kliknutí je sbalí.
- Volitelná položka **Pokročilé zkratky výběru** zůstává sbalitelná **uvnitř nápovědy**; není to další samostatný řádek nástrojů.
- Otazník je skutečné tlačítko s `aria-expanded` a `aria-controls`, dostupné klávesnicí a s jasnou indikací otevřeného stavu.
- **Krátká kontextová nápověda pod ikonami** (dosavadní pravidlo Q9) zůstává zachována: při hover/fokusu se dočasně změní podle ikony; po opuštění se vrací ke skutečně aktivnímu nástroji. Je to něco jiného než rozsáhlá nápověda pod `?`.
- Výběr jiného nástroje nebo záložky nezmění obsah PDF a nepřevede původní výběr na jiný typ.

## 3. Bezkolizní význam – ZÁVAZNÉ

**Umístění ikon neznamená schválení nových gest.** Výběr textu a objektů musí nadále respektovat `SEL-01..03`, `TXT-01..04`, `OBJ-01..03`:

- **Prostřední tlačítko + tažení**: vždy globální obdélníkový výběr podle aktivního nástroje. Ani obdélníková ikona v panelu Text nesmí toto gesto vypnout či předefinovat. Samostatný prostřední klik nemá výběrový účinek.
- **Levý klik v nástroji Text**: aktivace editoru a umístění caret; **levý dvojklik** vybere slovo, další dvojkliky `slovo → segment → věta → odstavec`.
- **Ctrl + levý klik v Textu**: vybrat větu; **Ctrl + levý klik v Objektech**: cyklovat překryté objekty. Tyto funkce se nesmějí zaměnit.
- **Objektové laso** již používá levé klikání pro zadávání vrcholů a levý dvojklik/Enter pro uzavření. Ikona textového lasa **nesmí** bez samostatně schválené specifikace tento mechanismus převzít nebo globálně změnit.
- **Současný HTML prototyp:** obdélníkový režim Text a textové laso jsou **viditelné, ale záměrně neaktivní návrhy**, jak bylo stanoveno už ve v3.1. Jsou nyní pouze přemístěny mezi ostatní horní ikony. Nejde o indikaci, že jsou funkční v PDF editoru.
- Textový kurzor a nový textový blok jsou odlišné režimy a jejich dosavadní stav v prototypu zůstává zachován. Nápověda sama nikdy nezmění aktivní režim ani historii.

## 4. Akceptační kontrola UI

1. V panelu `Text → Výběr` jsou v jediné horní řadě čtyři ikony a vpravo `?`; bez spodního řádku s lasem/obdélníkem.
2. Otazník je na pravém okraji téže řady v běžném i užším okně.
3. Podrobná nápověda je **zpočátku skrytá**, kliknutím na `?` se otevře a znovu zavře; uvnitř lze rozbalit zkratky.
4. Obě plánované ikony jsou v HTML označeny nedostupností; žádné dosavadní gesto se kvůli nim nepřemapuje.
5. Záložka **Typografie** a její hodnoty, ovládání pomocných panelů, adaptivní šířka stránky, Nastavení i související příkazy zůstávají nedotčeny.
6. Před jakoukoli budoucí změnou produkčního výběru spustit regresní testy SEL/TXT/OBJ a ověřit všechny modifikátory myši včetně objektového lasa, miniatur a Ctrl+Z/Y.

**Soubory:** `cely_vizual_text_vyber_napoveda_v3.2.html` (HTML demonstrace), tento dokument. Předchozí verzi v3.1 zachovat pro porovnání. **Žádná změna Go/Win32 tímto nevzniká.**