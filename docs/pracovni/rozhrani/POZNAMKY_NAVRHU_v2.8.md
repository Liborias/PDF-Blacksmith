# PDF Blacksmith – HTML v2.8: jednopozicové pomocné panely, historie a ikonové ovládání

**Datum:** 22. 11. 2026. **Stav:** interaktivní vizuální prototyp; aplikace Go/Win32 nebyla změněna.

## Pravá skupina panelů

- **Vlastnosti** jsou hlavní panel.
- **Výsledky vyhledávání, Záložky, Poznámky, Historie** sdílejí **jediný fyzický prostor** pro pomocný panel, bez ohledu na šířku okna. Otevřen může být vždy **nejvýše jeden**.
- Přepnutí pomocného ouška uzavře předchozí pomocný panel a zobrazí nový na stejné pozici; kliknutí na aktivní ouško panel zavře. Není povoleno řazení více pomocných panelů vedle sebe ani na ultraširokých monitorech.
- **Aktivní ouško:** šedé pozadí `#1e1f26` stejné jako obsah panelu, zvýrazněný je pouze jeho text do střídmé vínové; ikonka zůstává šedá. **Neaktivní ouška:** jiný, poněkud světlejší šedý podklad `#37353c`.
- Skupinové horní ouško s `« / »` stále sbalí/obnoví předchozí kombinaci panelů včetně vybraného pomocného panelu; jednotlivá ouška přesahují přímo do pracovní plochy (bez vyhrazené boční kolejnice).

## Kompaktní ikony v nastavení nástrojů

- Akční skupiny (Výběr objektů, Transformace, Uspořádání, řezy textu, zarovnání a další skupiny vytvořené pomocí `boxes()`) používají **ikonová tlačítka bez opakovaných popisků**.
- Význam je dostupný na `title` a `aria-label` (myš i klávesnice). **Pod skupinou** se při najetí/fokusu přepíše dočasný vysvětlující řádek s názvem a bližším účelem.
- Přidáno schválené skutečné SVG **Duplikovat** ze sady v4.1 (dříve symbol v HTML chyběl).
- Hlavní prováděcí tlačítka (např. Vložit strany) mají nadále jasný slovní popisek; přístupnost se nepodřizuje pouze ikoně.

## Zobrazení → Přizpůsobit šířce

- Režim vyplní **téměř celou skutečně volnou šířku** editoru mezi otevřenými panely; šířka se průběžně přepočítává při sbalení/rozbalení obou skupin, při přepínání pomocných panelů a při změně velikosti okna.
- Odstraněno staré omezení zoomu 160 %, které bránilo vyplnění širokého prostoru; maximální limit prototypu je nyní 900 %.
- Uživatelské ruční nastavení zoomu a 1:1 ukončí automatický režim Přizpůsobit šířce.
- V testu viewportu **1 600 × 900**: otevřené panely 134 %, zavřené obě strany 273 %, po obnovení znovu 134 %, každá z hodnot vyplnila asi 97–98 % aktuální šířky editoru. Při změně na 1 920 px se zoom automaticky přepočítal na 191 %.

## Úpravy → Historie úprav

- Z menu **Úpravy** lze nově otevřít pomocný panel **Historie** na sdílené pravé pozici, s vlastním ouškem a ikonou `history`.
- HTML obsahuje názornou časovou osu čtyř **ilustračních** kroků a ovládání předchozí/následující krok. Výběr pouze mění zvýraznění v časové ose a ukáže informativní hlášení, **NEVRACÍ skutečné změny PDF**.
- Skutečná implementace musí použít historii příkazů editoru pro aktuální dokument, navázat na stávající Ctrl+Z/Ctrl+Y, neporušit integritu uložených PDF a jasně odlišit nevratné změny.

## Testy HTML

V Chromium (skutečný browser):
- postupné přepínání všech čtyř pravých pomocných panelů a ověření, že je viditelný právě jeden;
- porovnání barvy aktivního ouška s panelem;
- nabídka Úpravy → Historie, výběr kroku a návrat původního rozložení po skupinovém sbalení;
- nástroj Objekty, ikona Duplikovat a změna nápovědného řádku po najetí;
- adaptivní šířka při otevření/sbalení/obnovení panelů, změně rozměrů okna a vypnutí režimu při 1:1.
**Výsledek: bez chyb JavaScriptu.** Ostatní nástroje včetně Soubor → Vložit strany zůstaly v prototypu zachovány.

**Pro Codex:** HTML je pouze orientační model chování a vzhledu. Bez dalšího schválení nezačínat implementovat Historii úprav ani měnit produkční panely; nejprve analyzovat existující Win32 UI a command history.
