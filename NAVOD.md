# PDF Blacksmith

Spusťte `PDFBlaksmith.exe` na 64bitových Windows 10 nebo 11. PDFium, Ghostscript a Python jsou součástí aplikace. Při prvním zpracování se rozbalí do `%LOCALAPPDATA%\PDF Blaksmith`.

## Otevření a čtení PDF

V levém panelu rozbalte **Soubor** a vyberte **Otevřít soubor**. Stejné tlačítko najdete vpravo v nastavení nástroje **Soubor**. Dialog můžete otevřít i zkratkou **Ctrl+O** nebo PDF přetáhnout do okna.

Panel nástrojů a nastavení jsou při spuštění otevřené. Nástroj vyberete vlevo, jeho nastavení se zobrazí vpravo. **Sloučit PDF** je dostupné hned po spuštění. Po načtení PDF se vlevo objeví také **Najít v PDF**, **Separace**, **Digitální podpis**, **Rozdělit PDF** a **Pořadí stránek**. Šířka panelů se přizpůsobuje obsahu a dostupnému místu. Na každé straně je jedna úzká lišta. Horní ouško **Levá strana** nebo **Pravá strana** sbalí otevřené panely na dané straně a dalším kliknutím je obnoví. Ostatní ouška ovládají jednotlivé panely.

Náhledy stránek jsou ve vlastním panelu vlevo mezi nástroji a dokumentem. Panel je společný pro všechny nástroje a ovládá se ouškem **Náhledy**. Kliknutím vyberete stránku, **Ctrl** přidává jednotlivé stránky a **Shift** vybírá souvislý rozsah. Kolečko seznam posouvá a **Ctrl + kolečko** mění velikost náhledů v rozsahu 10–60 % šířky panelu. Výběr se zachová při posouvání, změně velikosti, sbalení panelu i přepnutí nástroje.

- PDF se otevře v režimu **Po stránkách**. Kolečko posouvá zvětšenou stránku a na jejím okraji přejde na sousední stránku. V **Zobrazení → Náhled** lze přepnout na **Souvisle** a posouvat dokument plynule.
- Běžné náhledy stránek se načítají na pozadí podle volné paměti. Počet načtených stran je v **Zobrazení → Náhled**. Dokument zavřete křížkem vedle jeho názvu v horní liště nebo zkratkou **Ctrl+W**. Pokud máte neuložené změny, aplikace se před zavřením zeptá, zda je zahodit.
- **Ctrl + kolečko** zvětšuje nebo zmenšuje zobrazení kolem bodu pod kurzorem. Běžný náhled při větším přiblížení vykreslí viditelné výřezy z původních křivek a písem PDF v rozlišení obrazovky. Při listování a zoomu se předchozí obraz vymění najednou, jakmile je nový náhled hotový. Přiblížený výřez posunete tažením. Dolní lišta nabízí přechod na stránku, přizpůsobení celé stránky a přizpůsobení šířce.
- Okno se při spuštění maximalizuje. Tlačítkem vpravo nahoře nebo dvojklikem na volnou část záhlaví obnovíte jeho běžnou velikost. **Zobrazení → Náhled → Zobrazit na celou obrazovku** nebo **F11** otevře samotné PDF přes celou obrazovku. **Esc** obnoví předchozí rozložení panelů.
- Pokud PDF vyžaduje heslo k otevření, zadejte je v **Zobrazení → Náhled** a použijte **Načíst znovu s heslem**.

## Vyhledávání v PDF

Otevřete nástroj **Najít v PDF** nebo stiskněte **Ctrl+F**. Zadejte slovo či frázi a potvrďte **Enterem**. Seznam řadí shody podle aktuálního pořadí stran a u každé ukazuje krátký úryvek. Kliknutím na výsledek přejdete k nalezenému textu. **F3** přejde na další shodu, **Shift+F3** na předchozí. Vybraná shoda je vínová a ostatní jsou žluté.

Hledání pracuje s textovou vrstvou a rozlišuje diakritiku. Pro obrázky a skeny potřebuje dokument textovou vrstvu vytvořenou pomocí OCR. Aplikace spočítá všechny nalezené shody a v seznamu zobrazí prvních 2 000. Tlačítko **Vymazat** hledání zastaví a vyčistí výsledky.

## OCR skenů a textu v křivkách

Otevřete nástroj **OCR** a zvolte **Všechny strany**, **Aktuální strana**, **Vybrané náhledy** nebo **Vlastní výběr**. Pro vlastní výběr zadejte například `1, 3–5, 8`. Čísla odpovídají aktuálnímu pořadí v aplikaci. Před rozpoznáním otočte případné strany na bok do čitelné polohy.

Vyberte jazyk a klikněte na **Rozpoznat a uložit kopii…**. Čeština i angličtina jsou dostupné bez další instalace. Kopie obsahuje celý dokument. Na vybraných stranách přibude neviditelná textová vrstva a původní sken, křivky i barevné objekty zůstanou v PDF. Výsledek se otevře v aplikaci a text můžete zkontrolovat pomocí **Ctrl+F**. Pokud OCR nenajde žádný nový text, kopie se nevytvoří.

## Záložky v PDF

Na pravé liště otevřete ouško **Záložky**. Strom záložek má vlastní posuvník a zobrazí se mezi dokumentem a nastavením nástroje. Kliknutím na záložku přejdete na dostupnou cílovou stránku. Z klávesnice ji otevřete Enterem. Šipka u názvu sbalí či rozbalí větev.

Tlačítkem **Nová záložka…** vytvoříte záložku pro aktuální stranu. Název a cílovou stranu můžete změnit i u existující záložky. Změnu potvrďte tlačítkem **Použít úpravu**. Číslo cíle odpovídá aktuálnímu pořadí stran. **Vrátit změnu** obnoví stav před poslední úpravou.

Do souboru zapíšete změny přes **Soubor → Uložit** nebo **Uložit jako**. Záložky jsou součástí standardního stromu PDF `/Outlines` podle ISO 32000. Přejmenováním záložky se zachová její původní nastavení odkazu. Při změně cíle se záložka nastaví na celou vybranou stranu.

## Uložení a tisk

Stejné volby najdete v rozbalovací nabídce **Soubor** vlevo i nahoře v panelu nastavení vpravo. **Nedávné**, **Export** a **Podepsaná kopie** zobrazí své volby pod tlačítky vpravo.

- **Soubor → Nedávné** zobrazí až 10 naposledy otevřených PDF i po restartu. Vyberte dokument a stiskněte **Otevřít vybrané PDF**, Enter nebo na něj poklikejte. Pod seznamem je úplná cesta k vybranému souboru. Historie ukládá pouze cesty.
- **Soubor → Uložit** nebo **Ctrl+S** uloží změny do původního PDF po potvrzení přepsání. U podepsaného PDF aplikace vyžádá i potvrzení odstranění stávajících podpisů.
- **Soubor → Uložit jako** nebo **Ctrl+Shift+S** otevře dialog pro uložení kopie. Uložená kopie se otevře jako aktuální dokument.
- **Soubor → Tisk** nebo **Ctrl+P** otevře systémový dialog. V něm vyberte tiskárnu, rozsah stran a počet kopií. Volba **Výběr** vytiskne označené náhledy. Aplikace přizpůsobí stránky tisknutelné ploše a vykreslí je pro tisk v RGB s rozlišením až 300 dpi.
- **Soubor → Podepsaná kopie** otevře nastavení digitálního podpisu pod volbami souboru.

## Příprava pro tisk

1. Otevřete PDF a zvolte **Soubor → Export**.
2. V nabídce **Předvolba PDF** vyberte účel výstupu. **Vlastní nastavení** zpřístupní ruční výběr barev, převod na desku K i převod textu na křivky.
3. V nabídce **Strany k exportu** vyberte celý dokument, aktuální stranu, **Vybrané náhledy** nebo **Vlastní výběr**. Do pole můžete napsat jednotlivá čísla i rozsahy, například `1, 3–5, 8`. Čísla odpovídají aktuálnímu pořadí stran v aplikaci.
4. U předvolby pro tiskárnu vyberte profil ICC. Ve **Vlastním nastavení** zvolte požadovaný barevný výstup a případně převod textu na křivky. Text v naskenované bitmapě zůstává součástí bitmapy.
5. Zvolte **Uložit exportované PDF…** a vyberte výstupní soubor. Uložený výstup zobrazíte přes **Soubor → Otevřít výstup**.

### Předvolby PDF

| Předvolba | Výstup |
| --- | --- |
| PDF pro tisk | Ponechá barvy a průhlednost, obrázky převzorkuje nejvýše na 300 dpi a použije bezeztrátovou kompresi. |
| PDF připravené pro tiskárnu | Převede procesní barvy do vybraného profilu CMYK a vloží ho jako výstupní záměr. Výchozí profil je FOGRA39 (ISO Coated v2 300 %). Obrázky převzorkuje nejvýše na 300 dpi. |
| Digitální PDF · malá velikost | Použije sRGB pro obrazovku, obrázky převzorkuje nejvýše na 150 dpi a použije kompresi JPEG. |
| Digitální PDF · vysoká kvalita | Použije sRGB, ponechá rozlišení obrázků a komprimuje je bezeztrátově. |
| PDF pro další zpracování | Uloží původní obsah vybraných stran včetně textu, křivek a průhlednosti. Celý dokument bez úprav se kopíruje beze změny. |
| Zploštěné PDF | Složí viditelný obsah každé strany do jednoho obrazu sRGB při 300 dpi. Zachová rozměry a tiskové rámečky stran. Text se stane součástí obrazu. |
| PDF/X-1a:2003 a PDF/X-3:2003 | Procesní CMYK s vloženým tiskovým ICC profilem. Průhlednost se zploští při 300 dpi. |
| PDF/X-4 | Procesní CMYK s vloženým tiskovým ICC profilem. Zachovává živou průhlednost. |

### Export PDF/X

V **Soubor → Export → Předvolba PDF** vyberte požadovanou normu, rozsah stran a tiskový profil. Výchozí je FOGRA39. Tyto předvolby převedou i přímé barvy do procesního CMYK. Pokud zdroj nemá čistý formát ani rámeček kresby, použije se ořez, případně celý arch.

Export používá režim PDF/X Ghostscriptu a před uložením kontroluje tiskové rámečky, vložená písma, výstupní ICC profil, procesní barvy, průhlednost a strukturu PDF. Při chybě kontroly se výstup neuloží. Úplné nezávislé ověření podle ISO a kontrolu požadavků konkrétní tiskárny provádí její preflight.

### Nastavení exportu

Ve volbě **Vlastní nastavení** lze nastavit cílový barevný profil, záměr reprodukce ICC, kompenzaci černého bodu, převzorkování a kompresi obrázků i rozlišení kontroly a případné rastrové opravy desek K. Ruční parametry se použijí i při příštím spuštění.

### Pouze deska K

Aplikace ověřuje tiskové separace po jednotlivých stranách. Výchozí rozlišení je 150 dpi. Hláška s deskou **Cyan** znamená, že simulace tisku našla barvu na azurové desce. Může k tomu dojít například u objektu s vícebarevným prostorem DeviceN. Výpis uvádí konkrétní strany a desky, které vyšly aktivní.

Volba **Povolit rastrovou opravu K** zpracuje problematické strany jako bezeztrátové obrázky ve stupních šedi a poté znovu ověří všechny separace. Ostatní strany zůstanou vektorové tam, kde je zachoval převod v Ghostscriptu. Rastrová oprava je dostupná při převodu bez volby křivek. Výsledek v jiném tiskovém RIPu závisí také na jeho nastavení a výstupním profilu.

### Výstupní profily

| Profil | Použití |
| --- | --- |
| FOGRA39 · ISO Coated v2 (ECI) a varianta 300 % | Starší ofsetová podmínka pro natíraný papír |
| FOGRA47 · PSO Uncoated ISO12647 (ECI) | Starší ofsetová podmínka pro nenatíraný papír |
| FOGRA51 · PSO Coated v3 | Novější podmínka pro natíraný papír |
| FOGRA52 · PSO Uncoated v3 | Novější podmínka pro nenatíraný papír |
| FOGRA54 · PSO SC-B Paper v3 | Ofset na papír SC-B |

Při prvním použití stáhne aplikace vybraný profil přímo od ECI, ověří kontrolní součet a uloží jej pro další převody offline. Volbou **Vlastní výstupní profil ICC** lze načíst lokální tiskový profil CMYK ve formátu `.icc` nebo `.icm`. [European Color Initiative](https://www.eci.org/doku.php?id=en:downloads)

Export provede převod procesních barev CMYK a připojí zvolený profil ICC jako výstupní záměr PDF. Jako zdroj pro netagované hodnoty CMYK použije profil z výstupního záměru vstupního PDF, pokud je dostupný, jinak cílový profil. Přímé barvy mohou zůstat samostatnými separacemi. Pro ověření konkrétní normy PDF/X je potřeba preflight.

## Náhled PDF a tiskové separace

Nástroj **Separace** přepne náhled na simulaci tisku a vpravo zobrazí její nastavení.

- Tiskovou oblast lze přepnout na celý arch, ořez pro zobrazení, čistý formát nebo spadávku.
- Tiskový náhled simuluje přetisky. Panel uvádí výstupní záměr PDF a použitý profil CMYK. Simulaci lze přepnout na některý z profilů ECI.
- V seznamu desek jsou Cyan, Magenta, Yellow, Black a rozpoznané přímé barvy. Jakmile se seznam zobrazí, kliknutím na řádek zapnete nebo vypnete příslušnou desku. Náhled se skládá z vybraných desek. Údaje ukazují podíl bodů s nenulovým krytím a nejvyšší tón v každé desce při zobrazeném rozlišení.
- **Zkontrolovat celý dokument** nebo **Zkontrolovat vybrané náhledy** vypíše aktivní desky a čísla jejich stran. Kontrola bere v úvahu vybranou tiskovou oblast a profil simulace. U velkých formátů může snížit vzorkovací rozlišení a v přehledu uvede použitou hodnotu.

## Pořadí a náhledy stránek

Otevřete vlevo **Pořadí stránek** nebo **Zobrazení → Stránky**. Přesun, otočení i smazání lze použít na více označených náhledů najednou.

- **Přesunout** vloží vybrané stránky na zadanou pozici. Zadané číslo bude výsledná pozice první přesunuté strany. Stejného výsledku dosáhnete přetažením vybraných náhledů k vínové čáře mezi stránkami.
- **Smazat vybrané** odstraní výběr z rozpracovaného dokumentu. Použít lze také klávesu **Delete** při zaměření seznamu nebo přetažení náhledů na tlačítko. Dokument musí obsahovat alespoň jednu stranu.
- **Vrátit změnu** obnoví stav před posledním přesunem, smazáním nebo otočením. **Uložit upravenou kopii…** uloží úpravy do nového PDF. Rozpracované pořadí platí i pro export a rozdělení. U podepsaného souboru aplikace zobrazí potvrzení uložení bez původních podpisů.

Po smazání stránek se při uložení přenese výstupní záměr ICC a označení původní shody PDF/X se odstraní. Pro vytvoření nového PDF/X použijte příslušnou předvolbu v **Soubor → Export**.

## Rozdělení PDF po stránkách

Vlevo zvolte **Rozdělit PDF**. Vpravo vyberte všechny strany, aktuální stranu, **Vybrané náhledy** nebo vlastní rozsah, například `1, 3–5, 8`. Tlačítko **Vybrat složku a rozdělit PDF…** uloží každou vybranou stranu jako samostatné PDF, například `katalog-strana-003.pdf`.

Čísla souborů odpovídají aktuálnímu pořadí stran. Přenesou se i rozpracovaná otočení a existující ochrana heslem. Původní dokument zůstává beze změny. Pokud ve složce názvy už existují, aplikace požádá o jinou složku. U podepsaného dokumentu potvrdíte vytvoření souborů bez původních podpisů.

## Sloučení dokumentů

Vlevo otevřete **Sloučit PDF** a pomocí **Přidat PDF soubory…** vyberte vstupy. V tomto režimu je můžete také přetáhnout do okna. Tlačítky **Nahoru** a **Dolů** nastavíte pořadí. Pro každý soubor můžete zadat čísla a rozsahy stran, například `1, 3–5`. Prázdné pole zahrne celý vstup.

**Přidat otevřené PDF** převezme aktuální pořadí a otočení stran v rozpracovaném dokumentu. **Přidat výběr stran** předvyplní čísla označených náhledů. U souboru chráněného heslem zadejte heslo a klikněte na **Načíst počet stran**.

**Uložit sloučené PDF…** vytvoří jeden soubor z vybraných stran. Najdete jej také přes **Soubor → Otevřít výstup**. U podepsaných vstupů aplikace vyžádá potvrzení vytvoření kopie bez původních podpisů. U šifrovaných vstupů vyžádá potvrzení uložení výsledku bez šifrování.

Sloučení přenáší stránky ve zdrojových barevných prostorech. Pokud mají vstupy shodný vložený výstupní záměr ICC, přenese se do výsledku. Při rozdílu aplikace zobrazí zprávu o ICC profilu.

## Odstranění omezení heslem

V panelu **Export** zaškrtněte **Pouze odstranit omezení heslem** a nechte pole hesla prázdné. Pokud se vstupní PDF otevře bez hesla, aplikace uloží odemčenou kopii i bez hesla vlastníka. Pokud PDF vyžaduje heslo už při otevření, k přečtení šifrovaného obsahu je nutné platné heslo.

Při úpravě podepsaného PDF aplikace nabídne uložení kopie bez původních podpisů. Tiskový převod může zploštit tisknutelné poznámky a formuláře.

## Digitální podpisy

- **Podepsat** vytvoří podpis PAdES se SHA‑256 v nešifrované kopii PDF. Použijte certifikát se soukromým klíčem ve formátu `.p12` nebo `.pfx`. Viditelný podpis v novém poli se vloží na první stranu.
- **Ověřit podpisy** zobrazí podepisujícího, vydavatele, integritu dokumentu a rozsah podpisu. Offline kontrola využívá kořeny Windows a volitelný certifikát vlastní autority. Aktuální stav odvolání u vydavatele vyžaduje online ověření.
