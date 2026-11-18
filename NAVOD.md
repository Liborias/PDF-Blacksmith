# PDF Blacksmith

Spusťte `PDFBlaksmith.exe` na 64bitových Windows 10 nebo 11. PDFium, Ghostscript a Python jsou součástí aplikace. Při prvním zpracování se rozbalí do `%LOCALAPPDATA%\PDF Blaksmith`.

## Otevření a čtení PDF

PDF otevřete přes rozbalovací nabídku **Soubor** a volbu **Otevřít soubor…**, zkratkou **Ctrl+O** nebo přetažením do okna. Dokument se zobrazí uprostřed. Postranní panel se otevře po výběru nástroje nebo volby s nastavením. **Separace** a **Digitální podpis** se objeví po načtení PDF.

- PDF se otevře v režimu **Po stránkách**. Kolečko posouvá zvětšenou stránku a na jejím okraji přejde na sousední stránku. V panelu **Zobrazení** lze přepnout na **Souvisle** a posouvat dokument plynule.
- Běžné náhledy stránek se načítají na pozadí podle volné paměti. Panel **Zobrazení** ukazuje počet stránek v paměti. Dokument zavřete volbou **Soubor → Zavřít PDF** nebo zkratkou **Ctrl+W**.
- **Ctrl + kolečko** zvětšuje nebo zmenšuje zobrazení kolem bodu pod kurzorem. Běžný náhled při větším přiblížení vykreslí viditelné výřezy z původních křivek a písem PDF v rozlišení obrazovky. Při listování a zoomu se předchozí obraz vymění najednou, jakmile je nový náhled hotový. Přiblížený výřez posunete tažením. Dolní lišta nabízí přechod na stránku, přizpůsobení celé stránky a přizpůsobení šířce.
- Okno se při spuštění maximalizuje. Tlačítkem vpravo nahoře nebo dvojklikem na volnou část záhlaví obnovíte jeho běžnou velikost. **Skrýt panel** přenechá dokumentu celou šířku okna. **F11** otevře samotné PDF přes celou obrazovku a **Esc** vrátí předchozí rozložení.
- Pokud PDF vyžaduje heslo k otevření, zadejte je v panelu **Zobrazení** a použijte **Načíst znovu s heslem**.

## Uložení a tisk

- **Soubor → Uložit** nebo **Ctrl+S** uloží změny do původního PDF po potvrzení přepsání. U podepsaného PDF aplikace vyžádá i potvrzení odstranění stávajících podpisů.
- **Soubor → Uložit jako…** nebo **Ctrl+Shift+S** uloží kopii a otevře ji jako aktuální dokument.
- **Soubor → Tisk…** nebo **Ctrl+P** otevře systémový dialog. V něm vyberte tiskárnu, rozsah stran a počet kopií. Aplikace přizpůsobí stránky tisknutelné ploše a vykreslí je pro tisk v RGB s rozlišením až 300 dpi.
- **Soubor → Uložit podepsanou kopii…** otevře nastavení digitálního podpisu.

## Příprava pro tisk

1. Otevřete PDF a zvolte **Soubor → Export…**.
2. V nabídce **Strany k exportu** vyberte celý dokument, aktuální stranu nebo **Vlastní výběr**. Do pole můžete napsat jednotlivá čísla i rozsahy, například `1, 3–5, 8`. Čísla odpovídají aktuálnímu pořadí stran v aplikaci.
3. V poli **Barevný výstup** zvolte stupně šedi pro desku K, cílový profil CMYK nebo zachování barev. Pro uložení vybraných stran s původním obsahem zvolte **Zachovat barvy** a ponechte volby převodu vypnuté.
4. Potřebujete-li text jako vektorové obrysy, zaškrtněte **Převést text na křivky**. Text v naskenované bitmapě zůstává součástí bitmapy.
5. Zvolte **Uložit exportované PDF…** a vyberte výstupní soubor. Uložený výstup zobrazíte volbou **Soubor → Otevřít výstup**.

### Nastavení exportu

V postranním panelu **Export** se volí cílový barevný profil, záměr reprodukce ICC, kompenzace černého bodu, převzorkování a komprese obrázků i rozlišení kontroly a případné rastrové opravy desek K. Nastavené parametry se použijí i při příštím spuštění.

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

Nástroje záložky **Stránky** lze používat při prohlížení dokumentu. Záložka **Separace** přepne zobrazení na tiskovou simulaci.

- Tiskovou oblast lze přepnout na celý arch, ořez pro zobrazení, čistý formát nebo spadávku.
- Tiskový náhled simuluje přetisky. Panel uvádí výstupní záměr PDF a použitý profil CMYK. Simulaci lze přepnout na některý z profilů ECI.
- V seznamu desek jsou Cyan, Magenta, Yellow, Black a rozpoznané přímé barvy. Jakmile se seznam zobrazí, kliknutím na řádek zapnete nebo vypnete příslušnou desku. Náhled se skládá z vybraných desek. Údaje ukazují podíl bodů s nenulovým krytím a nejvyšší tón v každé desce při zobrazeném rozlišení.
- **Zkontrolovat celý dokument** vypíše desky, na kterých jsou při vykreslení skutečně tiskové body, a čísla jejich stran. Kontrola bere v úvahu vybranou tiskovou oblast a profil simulace. U velkých formátů může snížit vzorkovací rozlišení a v přehledu uvede použitou hodnotu.
- V editoru lze stránku otočit o 90°, přesunout ji o jednu pozici nebo poslední úpravu vrátit. **Uložit upravenou kopii…** uloží změny v PDF se zachováním obsahu stránek a odkazů na ně. Rozpracované otočení a pořadí se promítnou i do exportu. U podepsaného souboru si aplikace vyžádá potvrzení uložení kopie bez původních podpisů.

## Odstranění omezení heslem

V panelu **Export** zaškrtněte **Pouze odstranit omezení heslem** a nechte pole hesla prázdné. Pokud se vstupní PDF otevře bez hesla, aplikace uloží odemčenou kopii i bez hesla vlastníka. Pokud PDF vyžaduje heslo už při otevření, k přečtení šifrovaného obsahu je nutné platné heslo.

Při úpravě podepsaného PDF aplikace nabídne uložení kopie bez původních podpisů. Tiskový převod může zploštit tisknutelné poznámky a formuláře.

## Digitální podpisy

- **Podepsat** vytvoří podpis PAdES se SHA‑256 v nešifrované kopii PDF. Použijte certifikát se soukromým klíčem ve formátu `.p12` nebo `.pfx`. Viditelný podpis v novém poli se vloží na první stranu.
- **Ověřit podpisy** zobrazí podepisujícího, vydavatele, integritu dokumentu a rozsah podpisu. Offline kontrola využívá kořeny Windows a volitelný certifikát vlastní autority. Aktuální stav odvolání u vydavatele vyžaduje online ověření.
