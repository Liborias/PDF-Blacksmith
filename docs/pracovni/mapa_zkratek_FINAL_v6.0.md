# PDF Blacksmith — JEDINÁ MAPA ZKRATEK A GEST v6.0

**Normativní zdroj:** [CODEX_MASTER_v6.0.md](CODEX_MASTER_v6.0.md) · [strojová data](mapa_zkratek_FINAL_v6.0.json). Tato mapa nahrazuje v1.0–v1.4. Kódy řídící kontexty platí pro **stávající Go/Win32 + schválený návrh HTML v3.26**; jejich stav rozlišit.

**Legenda:** A = nalezeno statickou historickou inventurou Go a znovu nutno ověřit v aktuálním běhu Windows; U/D = závazné nové chování UI nebo HTML demo, **ne automaticky hotová Go implementace**; P = samostatný návrh.

## A – staticky doložené původní zkratky, produkčně znovu ověřit

| Zkratka / gesto | Význam | Kontext |
|---|---|---|
| Ctrl+O | Otevřít PDF | global/workspace |
| Ctrl+F | Najít v PDF | global/workspace |
| Ctrl+W | Zavřít dokument | global/workspace |
| Ctrl+S | Uložit | global/workspace |
| Ctrl+Shift+S | Uložit jako | global/workspace |
| Ctrl+P | Tisk | global/workspace |
| Ctrl+Tab | Další dokumentová karta | workspace |
| Ctrl+Shift+Tab | Předchozí dokumentová karta | workspace |
| F3 | Další výsledek vyhledávání | search |
| Shift+F3 | Předchozí výsledek vyhledávání | search |
| F11 | Celá obrazovka | workspace |
| Esc | Návrat z celého zobrazení / zrušení aktivního gesta | fullscreen/active gesture |
| Ctrl+Z | Zpět | document |
| Ctrl+Y | Znovu | document |
| Ctrl+Shift+Z | Znovu (alternativa) | document |
| Ctrl+X | Vyjmout | inline text |
| Ctrl+C | Kopírovat | inline text/canvas |
| Ctrl+V | Vložit | inline text |
| Ctrl+A | Vybrat vše | inline text/canvas |
| Ctrl+Enter | Použít úpravu textu | inline text |
| Alt + tažení | Přesunout vybraný text nebo objekt | vybraný obsah na plátně |
| Ctrl+B | Tučné | inline text |
| Ctrl+I | Kurzíva | inline text |
| Ctrl+U | Podtržené | inline text |
| Ctrl+L | Zarovnat vlevo | inline text |
| Ctrl+E | Zarovnat na střed | inline text |
| Ctrl+R | Zarovnat vpravo | inline text |
| Ctrl+J | Zarovnat do bloku | inline text |
| Ctrl++ | Přiblížit | canvas |
| Ctrl+- | Oddálit | canvas |
| Ctrl+0 | Přizpůsobit stránce | canvas |
| T / V / M / L | Původní režimy textového výběru — NEMIGROVAT AUTOMATICKY | canvas |
| Home / End | První / poslední stránka | canvas |
| Page Up / Page Down | Posun o stránku | canvas |
| Šipka vlevo / vpravo | Předchozí / následující stránka | canvas |
| Mezerník | Posun plátna dle fokusu | canvas |
| Prostřední tlačítko + tažení | Původní obdélníkový výběr, ověřit proti Q1 | canvas |

**Pozor:** `Ctrl+wheel` v plátně je historicky uváděno i mezi Go handlery, ale jeho nový **lokální oddělený režim pro miniatury** je až návrh U. Zkratky T/V/M/L ve staré Go aplikaci neznamenají automaticky nové univerzální výběry.

## U/D – platné cílové chování a interaktivní HTML

| Gesto | Výsledek | Kontext |
|---|---|---|
| Ctrl+kolečko | Lokální zoom PDF (HTML), přibližně pod kurzorem | editor |
| Ctrl+kolečko | Lokální změna velikosti miniatur, bez změny PDF zoomu | thumbnails |
| Ctrl+kolečko | Mimo editor a miniatury NEBLOKOVAT browser zoom | outside |
| Levé tlačítko + kolečko | Normální svislé listování, na okraji další/předchozí stránka | pages |
| Klik | Vybrat jednu stránku | page overview |
| Ctrl+levý klik | Přidat/odebrat stránku ze sdíleného stránkového výběru | page overview/thumbnails |
| Shift+klik | Souvislý rozsah stránek | page overview/thumbnails |
| Enter | Vybranou stránku otevřít na výšku pracovní plochy | page overview; canvas focus only |
| Tažení stran | Přesouvat vybraný seznam; HTML pouze simulace | page overview |
| Klik vodítko | Jediné vodítko + dočasné Pomůcky rozvržení | guide hit-test |
| Ctrl+klik vodítko | Přidat/odebrat označené vodítko | guide hit-test |
| Ctrl+tažení vodítka >7px | Duplikovat vodítko, nové ID; žádná duplicita při pouhém kliknutí | guide |
| Tažení vodítka >7px | Přesunout odemčené vodítko | guide |
| Pravý klik vodítko | Barva/zámek/koš, NE hover menu | guide |
| Delete | Smazat vybrané nezamčené vodítko jen v jeho kontextu | guide selection |
| Esc | Zrušit právě rozpracované tažení/kreslení, jinak původní význam | active gesture |
| Enter / dvojklik | Dokončit laso (ne v editačním poli) | lasso active |
| Klik volné plochy I-kurzorem | Nový bodový text | text caret |
| Tažení textovým rámcem | Nový odstavcový text | text box |
| Klik/tažení nového pole | Vložit/posunout demo F2 widget, nikoliv skutečný PDF záznam | form design HTML |

## P – pouze návrhy, nezavádět bez rozhodnutí

| Kombinace | Zamýšlený význam |
|---|---|
| Ctrl+Shift+M | Magnet zapnout/vypnout |
| F9 | Zobrazit/skryt mřížku |
| Ctrl+Shift+H | Zobrazit/skryt vodítka |
| Ctrl+Shift+R | Zobrazit/skryt pravítka |
| Shift+F9 | Volitelně přichytit k mřížce |
| Ctrl+1 | 100 % zobrazení |
| A / T / M / L | Navržená migrace univerzálních výběrových režimů; původní T/V/M/L zachovat do schválení |

## Jednotná priorita a řešení konfliktů
1. **Modální dialog / klávesnice systému / aktivní textový input / IME / formulářové pole** má přednost před zkratkami plátna. Zadávání znaků, Enter v poli, ESC v dialogu a AltGr musí fungovat přirozeně.
2. Poté **aktivní gesto** (tažení vodítka, objektu, textový caret/box, laso), pak konkrétní hit-test, potom výběr celé stránky jen v přehledu/miniaturách. Ctrl+klik na vodítku není Ctrl+klik na stránce; dokončit click vs drag při `pointerup`, práh cca 7 CSS px.
3. Každý příkaz má **jediný centrální registr**: výchozí binding, případná uživatelská výjimka, kontexty, štítek z i18n. Nekolidující vlastní zkratku lze změnit jen po potvrzení Použít/OK. Esc zruší záznam, Backspace resetuje, kolize nahlásit bez uložení.
4. Mění-li se pouze viditelnost mřížky/pravítek/vodítek/magnetu, nejde o editaci PDF a nesmí se vytvářet obsahové Undo/dirty. Přesunutí skutečné stránky nebo objektu musí být vratné, pokud se jednou implementuje.
5. Browser Ctrl+wheel mimo zóny `#editor`/`#thumbPane` neodchytávat. Native Chrome/Edge může mít výjimky a přístupnost: neprohlašovat absolutní podporu bez praktického testu. V Go/Win32 volit lokální vstupní handler.
6. Regrese: `T/V/M/L`, `Ctrl+F`, `Ctrl+S`, `Ctrl+Tab`, `Ctrl+L/E/R/J`, `Ctrl+B/I/U`, `Delete`, `Esc`, `Ctrl+wheel`, F3, F11 a štěpení výběru. Testovat CZ/EN, AltGr, klávesnici mimo okno, dialog/formulář/inline editor i plátno.

**Žádná klávesa v mapě sama o sobě nezakládá nový příkaz nebo neautorizuje produkční implementaci.**
