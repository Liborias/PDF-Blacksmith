# Správa barev – Gray a pouze K

## Způsoby převodu

Ve **Soubor > Export > Vlastní nastavení > Barevný výstup** i v panelu **Separace > Převod barev aktuálního dokumentu** jsou k dispozici stejné režimy:

1. **Stupně šedi (DeviceGray).** Ghostscript `pdfwrite` přepočítá procesní barvy pomocí ICC do odstínů šedi; vykreslení na CMYK deskách se kontroluje (C/M/Y a spot desky musejí být prázdné). PDF nemusí obsahovat explicitní hodnoty CMYK. Primárně určené pro černobílý výstup.
2. **Složená černá → K100.** Změní jen barvy přímo zapsané jako DeviceCMYK, kde K >= 99,9 % a C/M/Y obsahují nenulové složky, na `C0 M0 Y0 K100`. *Ostatní barvy, fotografie, ICCBased/DeviceN/Separation a gradienty nemění.* Jde o záměrně úzkou a předvídatelnou operaci.
3. **Pouze K – explicitní CMYK.** Nejprve provede ICC převod do šedi, poté mění DeviceGray vektorové operace a běžné 8bitové bitmapy Gray na explicitní `DeviceCMYK [0, 0, 0, K]`. Připojí vybraný tiskový CMYK ICC jako výstupní záměr. U nepodporovaných konstrukcí skončí chybou místo nezaručeného výsledku. Po výstupu ověřuje tiskové separace.

Panel Separace provádí převod do **nové kopie**, kterou následně otevře. Nemění bez potvrzení původní soubor. Při změnách barev v PDF budou stávající elektronické podpisy neplatné.

## ICC profily a předtiskové podmínky

U režimů 1 a 3 jsou dostupné:

- **Automaticky podle PDF** – použijí se dostupné zdrojové profily/výstupní záměr nebo standardní předvolby Ghostscriptu.
- **Gray Gamma 2.2** – jeden kanál Gray, dynamicky vytvořený profil ICC v2 s tonální charakteristikou gamma 2,2. Vhodný pro obecnou digitální výměnu; není to norma konkrétního tisku.
- **Gray Gamma 1.8** – jeden kanál Gray, dynamicky vytvořený ICC v2 s gamma 1,8; především pro starší pracovní postupy.
- **Vlastní ICC Gray (.icc/.icm)** – doporučeno, pokud tiskárna poskytuje skutečný profil pro černou desku, ofsetový tisk či konkrétní papír. Zde lze použít i licencované/vlastní **Dot Gain 10/15/20/25/30 %** profily – *jejich přesná křivka nesmí být nahrazena pouhým přejmenováním profilu Gray Gamma*.

U režimu 3 se samostatně určuje **cílový CMYK výstupní záměr** (např. ECI PSO Coated v3 / FOGRA51 pro natíraný papír, PSO Uncoated v3 / FOGRA52 pro nenatíraný, další profily z katalogu). Tyto profily jsou ICC charakteristiky konkrétní tiskové podmínky, nikoli samotný Gray profil. Pro výběr je rozhodující specifikace tiskárny. Použití výstupního záměru samo o sobě nedokazuje shodu PDF/X.

Obvyklá volba reprodukčního záměru pro fotografii je relativně kolorimetrická s kompenzací černého bodu, případně perceptuální podle charakteru profilu a požadavků tiskárny. Text a technické čáry mohou vyžadovat zachování 100% černé, proto režimy Gray/K explicitně před převodem chrání vektory s K100.

**Normativní rámec:** ISO 12647-2 a FOGRA charakteristiky pro ofset; PDF/X-1a, PDF/X-3, PDF/X-4 pro výměnu tiskových PDF. PDF/X není Gray profil a nevyřeší automaticky separaci jen do K.

## Limity a kontroly

- `DeviceGray` není stejné jako *explicitní* `DeviceCMYK (0,0,0,K)`, ačkoli na CMYK zařízení často tiskne jen do K.
- Tisková separace se ověřuje přes Ghostscript `tiffsep` po převodu. Pořád je nutná závěrečná vizuální předtisková kontrola, hlavně přetisky, průhlednosti, stíny, spot barvy, rámečky a embedded fonty.
- U režimu 3 mohou neobvyklé masky, barevné vzorky, inline obrázky, 16bitové obrazy, ICCBased nebo DeviceN/shading operace vyžadovat pokročilejší transformaci. Současná implementace při zjištění některých nepodporovaných prvků *odmítne export*. Nejde o univerzální replacement za Adobe Preflight/Convert Colors.
- V režimu 2 se **žádná** fotografie nepřevádí pixelově; jedná se o opravu bohaté černé u DeviceCMYK kresby/textu.
- Rychlá kontrola lokálně: `python -m unittest discover -s tests -p 'test_color_conversion.py'`; pro release dále ruční integrační test s Ghostscriptem a realistickými výstupními PDF a ověřením desek.
- Některé profily ICC mohou podléhat samostatné licenci; aplikace vlastní ICC nepřibaluje bez prověření podmínek.

## Reference

- Ghostscript: https://ghostscript.readthedocs.io/en/latest/GhostscriptColorManagement.html
- Ghostscript PDF writer: https://ghostscript.readthedocs.io/en/latest/VectorDevices.html
- European Color Initiative: https://eci.org/en/downloads
- FOGRA: https://fogra.org/en/downloads/work-tools/characterisation-data
- Adobe Photoshop color working spaces: https://helpx.adobe.com/photoshop/desktop/adjust-color/color-effects-techniques/color-working-spaces.html
