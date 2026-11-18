# PDF Tisk

Česká utilita pro Windows 10 a 11 (64bitové x86). Spusťte `PDFTisk.exe`. Aplikace při prvním zpracování rozbalí své součásti do `%LOCALAPPDATA%\PDF Tisk`. Používá lokální soubory a nevkládá hesla do příkazové řádky.

## Příprava PDF pro tisk

1. Nainstalujte aktuální 64bitový **Ghostscript AGPL**. Tlačítko **Stáhnout…** otevře oficiální stránku. Aplikace hledá Ghostscript v běžných instalačních složkách. Ručně lze vybrat `gswin64c.exe` ve složce `bin`. [Ghostscript pro Windows](https://www.ghostscript.com/releases/gsdnld.html)
2. Vyberte vstupní PDF a režim **Upravit pro tisk**.
3. Zaškrtněte **Převést do šedi a ověřit desku K**, **Převést text na křivky** nebo obojí. U chráněného PDF zadejte platné stávající heslo.
4. Klikněte na **Uložit upravené PDF…** a zvolte jiný název souboru.

Převod používá šedý prostor DeviceGray a ověřuje tiskové separace všech stran při 150 dpi. Do cílové složky uloží výstup, pokud jsou desky C, M, Y a případné přímé desky při této kontrole prázdné. Pokud některá složitá přímá barva zůstane aktivní, aplikace vypíše její desku a výstup neuloží. Výsledek v jiném tiskovém RIPu závisí také na jeho nastavení a výstupním profilu.

Převod textu na křivky odstraňuje běžnou textovou vrstvu, takže text se po převodu běžně nedá vyhledat nebo kopírovat. Text ve skenu je už součástí bitmapy. Ghostscript vytváří tiskovou kopii PDF, v níž se mohou zploštit tisknutelné poznámky a formuláře. Křivky nebo převod do šedi udělejte před podepsáním finálního souboru.

## Kopie bez hesla

V režimu **Upravit pro tisk** zaškrtněte **Pouze uložit kopii bez hesla**, zadejte stávající heslo k PDF a uložte novou kopii. Pro tuto operaci Ghostscript nepotřebujete. Při úpravě podepsaného PDF si aplikace vyžádá potvrzení k uložení kopie bez původních podpisů.

## Digitální podpisy

- **Podepsat** vyžaduje nešifrované PDF a soubor `.p12` nebo `.pfx` s osobním certifikátem a soukromým klíčem. Zadejte heslo certifikátu a uložte podepsanou kopii. Podpis je typu PAdES se SHA‑256. Aplikace vyplní jedno připravené prázdné podpisové pole nebo založí nové. Viditelný podpis v novém poli se vloží k okraji první strany.
- **Ověřit podpisy** zobrazí podepisujícího, vydavatele certifikátu, integritu, důvěryhodnost a rozsah podpisu. Používá kořenové certifikáty Windows. Důvěryhodný kořen vlastní certifikační autority lze doplnit jako soubor PEM či DER.
- Ověření probíhá offline, takže se nedotazuje vydavatelů na aktuální odvolání certifikátu. Tato verze nabízí podepisování se souborem PKCS#12 a ověření vložených časových razítek. Čipové karty, systémové soukromé klíče, přidávání autoritativního časového razítka a dlouhodobé ověření LTV do této verze nepatří.

## Zdrojový kód

Archiv `PDFTisk-zdroj.zip` obsahuje zdrojový kód a postup sestavení. Převody zajišťuje Ghostscript, odšifrování pikepdf s QPDF a podpisy pyHanko.
