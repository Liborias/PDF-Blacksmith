# PDF Blaksmith

Spusťte `PDFBlaksmith.exe` na 64bitových Windows 10 nebo 11. Ghostscript a Python jsou součástí aplikace. Při prvním zpracování se rozbalí do `%LOCALAPPDATA%\PDF Blaksmith`.

## Příprava pro tisk

1. Vyberte vstupní PDF a kartu **Upravit pro tisk**.
2. V poli **Barevný výstup** zvolte stupně šedi pro desku K, cílový profil CMYK nebo zachování barev.
3. Potřebujete-li text jako vektorové obrysy, zaškrtněte **Převést text na křivky**. Text v naskenované bitmapě zůstává součástí bitmapy.
4. Zvolte **Uložit upravené PDF…** a vyberte výstupní soubor.

### Pouze deska K

Aplikace ověřuje tiskové separace po jednotlivých stranách při 150 dpi. Hláška s deskou **Cyan** znamená, že simulace tisku našla barvu na azurové desce. Může k tomu dojít například u objektu s vícebarevným prostorem DeviceN. Výpis uvádí konkrétní strany a desky, které vyšly aktivní.

Volba **Povolit rastrovou opravu K (600 dpi)** zpracuje problematické strany jako bezeztrátové obrázky ve stupních šedi a poté znovu ověří všechny separace. Ostatní strany zůstanou vektorové tam, kde je zachoval převod v Ghostscriptu. Rastrová oprava je dostupná při převodu bez volby křivek. Výsledek v jiném tiskovém RIPu závisí také na jeho nastavení a výstupním profilu.

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

## Odstranění omezení heslem

V režimu **Upravit pro tisk** zaškrtněte **Pouze odstranit omezení heslem** a nechte pole hesla prázdné. Pokud se vstupní PDF otevře bez hesla, aplikace uloží odemčenou kopii i bez hesla vlastníka. Pokud PDF vyžaduje heslo už při otevření, k přečtení šifrovaného obsahu je nutné platné heslo.

Při úpravě podepsaného PDF aplikace nabídne uložení kopie bez původních podpisů. Tiskový převod může zploštit tisknutelné poznámky a formuláře.

## Digitální podpisy

- **Podepsat** vytvoří podpis PAdES se SHA‑256 v nešifrované kopii PDF. Použijte certifikát se soukromým klíčem ve formátu `.p12` nebo `.pfx`. Viditelný podpis v novém poli se vloží na první stranu.
- **Ověřit podpisy** zobrazí podepisujícího, vydavatele, integritu dokumentu a rozsah podpisu. Offline kontrola využívá kořeny Windows a volitelný certifikát vlastní autority. Aktuální stav odvolání u vydavatele vyžaduje online ověření.
