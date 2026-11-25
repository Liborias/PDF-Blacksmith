# PDF Blacksmith – doplněk rozhraní z 25. listopadu 2026

**Navazuje na** [HTML v3.25](cely_vizual_v3.25.html) a [doplněk v3.25](UI_FINALNI_KONTRAST_AKTIVNI_IKONY_v3.25.md).
**Rozsah:** přesně cílené změny HTML prototypu (bez produkční Go/Win32/PDFium implementace) a návrh nového adaptivního režimu až jako poslední úkol Codexu. Schválená historie a předchozí funkce zůstávají zachovány.

## Přímé změny HTML prototypu
1. **Stránky PDF v hlavní ploše** jsou bez dekorativního vínového obrysu ve všech režimech (po stránkách, souvislé, na šířku); skutečná grafika PDF, uživatelská anotace, interaktivní selekce a formulářové rámečky se nemění. Stránku lze odlišit přirozeným kontrastem a stínem.
2. **Kolečko v režimu po stránkách:** obyčejné rolování posouvá uvnitř vysoké stránky; na dolním/horním okraji přejde na následující/předchozí stránku. Přechod má ochranu proti vícenásobnému přeskočení jediným gestem. Nedotýkat se scrollování vstupních polí ani současného souvislého režimu.
3. **Vysvětlující text OCR** o sdíleném rozpoznávacím jádru je přesunut do nápovědy u příslušné sekce (kruhový otazník bez obdélníkového rámu). Totéž pravidlo platí pro další vysvětlivky mimo přímé ovládání; žádné vstupní volby se nesmějí skrývat. Horní tlačítko Nápověda na druhé liště obsahuje jen znak ? ve výšce opticky sjednocené s disketou a nastavením, bez kruhu uvnitř tlačítka. Klávesový fokus a přístupná označení zachovat.
4. **Kompaktnější panely:** odstranit prázdné pevně vysoké kontejnery nápovědy; jednotlivé sekce v pravém panelu mají oddělovací čáru a malé sbalovací ouško. Sbalení nesmí vymazat nastavení, změnit výběr nástroje nebo zničit původní event handlery. Výchozí stav skupin je rozbalený; stav sbalení je zachováván po opětovném vykreslení v HTML.
5. Palety, jejich sémantické barvy, všechny schválené interakce SEL/TXT/OBJ, vodítka, zkratky, Undo/Redo, F1 a funkce PDF se nesmějí měnit.

## Kontrolní scénáře pro současnou v3.25
- Režim po stránkách: uvnitř stránky rolovat; na konci přejít právě o jednu stránku; při zpětném rolování navázat od jejího spodního okraje; Ctrl+wheel v této etapě nepoužívat jako navigaci stránkami.
- Souvislé zobrazení: kolečko plynule roluje jako předtím, žádný vínový obrys.
- Přizpůsobit šířce: zkontrolovat nepřítomnost dekorativního outline i po opakovaném přepnutí režimu.
- OCR, Objekty, Text: vysvětlivky pod otazníkem, prázdné sloty nápovědy neblokují prostor, skupiny lze sbalit a opět rozbalit bez ztráty hodnot, přístupná klávesnicí.
- Pomoc, všechny světlé/tmavé profily: stejné funkce, kontrast a kruhový otazník v sekci; na horní liště otazník bez dalšího kruhu.
- Kontroly HTML jsou návrhové, nikoli potvrzená funkčnost Win32. Kompletní prohlížečové testy ještě provést.

## Poslední plánovaná etapa – adaptivní vícestránkový režim (NEIMPLEMENTOVAT ZATÍM)
**Záměr:** nový režim navazující na „po stránkách“ s pružným uspořádáním stránek jako flex/grid. Při oddálení/změně dostupného prostoru se do viewportu začnou postupně přidávat další stránky; podle šířky a měřítka mohou být vedle sebe a v řádcích pod sebou. Stránky jsou nadále skutečnými samostatnými stránkami PDF, ne ztrátovou bitmapovou koláží.

- Klik vybere jednu stránku; `Enter` nastaví zobrazení vybrané stránky na výšku pracovní plochy a zaměří ji.
- `Ctrl + kolečko` nad hlavní plochou nebo nad miniaturami mění měřítko **tohoto kontejneru**, nikoli zoom celého HTML/desktopového okna. Prohlížeč může mít nativní Ctrl+wheel zoom: ověř `wheel` listener `{passive:false}`, `preventDefault()`, omezení na konkrétní oblast, konflikty s přístupností a platformní výjimky. Neslibuj nezávislé chování bez testu. Pro produkční Win32 navrhni ekvivalentní zprávy myši.
- `Ctrl + levý klik` přidává/odebírá stránky v hromadném výběru. Dál se vybraná sada chová jako panel miniatur: výběr, přesun a hromadné operace podle oprávnění a aktivního nástroje. Vycházej ze **stejného datového modelu a pravidel** jako panel miniatur, ne z duplikovaných handlerů.
- Vyhodnoť kolize se současnými gesty textu, objektů, vodítek, formulářů a přetahování. Operace musí být bezpečné z hlediska Undo/Redo, ukládání a zachování PDF.
- Tento režim řešit **až na úplném konci** iniciačního zadání Codexu. Nejprve pouze analýza proveditelnosti a návrh architektury. Implementace až po samostatném výslovném schválení.
