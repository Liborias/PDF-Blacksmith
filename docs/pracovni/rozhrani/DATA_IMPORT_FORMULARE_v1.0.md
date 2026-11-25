# PDF Blacksmith – pravidla importu tabulkových dat do PDF formulářů v1.0

**Datum:** 25. 11. 2026 · **Rozsah:** návrh pro Codex, cílově hromadné vyplňování AcroForm. Pouze HTML demonstrační CSV, produkční import/export zatím neschválen.
**Napojení UI:** `cely_vizual_v3.26.html` → Formuláře → Z dat.

## Účel
Uživatel zvolí jednu PDF šablonu, připojí datovou tabulku, zmapuje sloupce na pole formuláře, prohlédne vzorek a validační výsledek, a poté vytvoří **jedno vyplněné PDF pro každý datový řádek** (nebo jedno konkrétní vybrané ID). V budoucnu možný alternativní výstup FDF/XFDF. Vždy chránit zdrojovou šablonu.

## Vstupní zdroje
1. **CSV (`.csv`)**: UTF-8 (akceptovat BOM), volitelné rozpoznání oddělovače čárka/středník/tabulátor, únik uvozovek a víceřádkové buňky podle RFC 4180. První neprázdný řádek je hlavička; zachovat nuly na začátku číslic, texty, diakritiku a skutečnou hodnotu buněk bez destruktivního převodu. CSV není spustitelný skript.
2. **Excel (`.xlsx`, pravděpodobně uživatelem myšlené „xmlx“)**: výběr sešitu, listu, pojmenované tabulky nebo oblasti, hlavička ve zvoleném řádku; práce s vypočtenými hodnotami buněk, neprovádět makra ani externí vzorce. Umožnit volbu zachování původního formátu datumu/času a výpis upozornění na nevyhodnocené vzorce, sloučené hlavičky či skryté řádky. `.xlsm` nepovolovat jako spustitelný dokument.
3. **Google Tabulky (Sheets)**: OAuth 2.0 s minimálním read-only oprávněním, URL/ID listu + výběr konkrétní záložky a A1 oblasti. Jednorázové načtení snapshotu a explicitní obnovení, aby změna zdrojové tabulky během dávky neprodukovala různá data. Uživatel musí vidět právě načtený počet řádků, vybraný list a čas synchronizace. Nepoužívat veřejný přístup nebo sdílené CSV URL jako náhradu za standardní autorizaci u neveřejných tabulek.
4. **Volitelná další etapa XML**: oddělit od XLSX, vyžadovat dokumentované schéma/XSD, zakázat externí entity a DTD; standardně nepovažovat XML za tabulku bez explicitního mapování. FDF/XFDF jsou samostatné PDF-specifické datové formáty, nikoli automatická náhrada libovolného XML.

## Jednotné schéma dat
- Jeden **logický řádek = jeden formulářový záznam**. Sloupce mají unikátní, neprázdné hlavičky; názvy nesmějí být záměnné jen mezerou/diakritikou bez výslovného potvrzení.
- Doporučený první sloupec `record_id`: unikátní stabilní identifikátor pro náhled, obnovitelnost a tvorbu názvu výstupu; pokud chybí, interně použít pořadí řádku, ale upozornit na chybějící identitu.
- Formulářová pole se mapují podle plně kvalifikovaného AcroForm `/T`; k dispozici explicitní ruční výběr a automatické doporučení přes přesnou shodu názvů/aliasů. Nikdy nepřiřadit více sloupců do jednoho pole mlčky; konflikty řešit v mapovacím dialogu.
- Zvláštní pole se stejným `/T` a více widgety sdílejí jednu hodnotu, pokud to model PDF vyžaduje. Radio skupina musí respektovat exportní /AS a /V hodnoty; zaškrtávací políčka nebrat jako libovolné české „ANO“, ale převést na skutečný exportní token (např. `Yes` / `Off`) na základě polí formuláře. Rozbalovací seznamy kontrolovat proti `/Opt`, ne jen proti textu na obrazovce.
- Před prvním importem uživatel potvrdí mapování, chování prázdných buněk, lokální formát dat a šablonu názvu výstupu. Mapovací profil lze uložit jen po souhlasu, s identifikací/fingerprintem verze PDF šablony.

### Doporučené datové typy
| Typ pole | Povolené vstupy | Výchozí pravidlo |
|---|---|---|
| Jednořádkový / víceřádkový text | UTF-8, délka podle omezení PDF | Neodstraňovat diakritiku, kontrolovat maximální délku |
| Datum | `YYYY-MM-DD` (ISO 8601), volitelný explicitní lokální převod | Před vyplněním převést do požadovaného formátu PDF; neodhadovat nejednoznačné datum |
| Číselné pole | Formát `1234.56`, případně explicitní `cs-CZ` | Neukládat telefon, PSČ či identifikátor jako číslo; zachovat nuly |
| Checkbox | `true/false`, `1/0`, `ANO/NE` po potvrzeném pravidle | Převést na exportní hodnoty daného pole, vč. `Off` |
| Radio / select | Stabilní kód nebo přesný exportní token | Nepřiřazené hodnoty hlásit, nežádat automatické „nejbližší“ |
| Podpis | Žádný obyčejný řetězec | Datový import nesmí simulovat digitální podpis |

### Příklad CSV (UTF-8; český oddělovač ;)
```csv
record_id;jmeno;email;datum;typ_spoluprace;souhlas
A001;Jan Novák;jan.novak@example.cz;2026-11-25;externí;ANO
A002;Eva Dvořáková;eva.dvorakova@example.cz;2026-11-26;interní;NE
```
První řádek jsou hlavičky, každý další je samostatný záznam. V buňce obsahující `;` použít řádné dvojité uvozovky a zdvojené vnitřní uvozovky. Názvy sloupců jsou případně odlišné od skutečných názvů AcroForm; mapují se ručně nebo potvrzeným pravidlem.

## Uživatelský průvodce
1. **Zdroj**: vybrat CSV/XLSX/Google list, kódování, oddělovač/list/oblast, případně rozsah řádků.
2. **Mapování**: na levé straně datové sloupce, na pravé formulářová pole (název, typ, stránka), jasné indikátory shody a konfliktů, volitelná složená hodnota z více sloupců bezpečnými operacemi (spojit, trim, datum, default) bez spouštění uživatelského JS.
3. **Náhled**: vybraný datový řádek vedle vyplněné kopie formuláře; odkazy na konkrétní chybná pole a možnost skočit k dalšímu řádku.
4. **Kontrola**: povinná pole, délky, datové typy, neznámé volby, hodnoty /AS /V /AP, duplicity ID, chybějící sloupce a konflikty mapování. Nabídnout režim „přeskočit chybný řádek“ vs „zastavit dávku“; nikdy mlčky nevyplnit nekonzistentní data.
5. **Výstup**: simulace a předběžný souhrn; po schválení cílová složka, jedno PDF na řádek, bezpečné názvy (`{record_id}_{jmeno}.pdf`, očištění zakázaných znaků, prevence kolizí a traversal), volitelný ZIP a protokol výsledků. Výchozí nastavení zachovává editovatelnost (neprovádí flatten), vedlejší volba explicitně a bezpečně vytvoří uzamčené/flatten výstupy jen pokud je daná funkce schválena a nepoškodí podpisy.
6. **Závěr**: počet úspěšných/chybných/přeskočených, odkazy na soubory, volitelný anonymizovaný log; neukládat obsah zdrojových řádků do telemetry.

## Bezpečnost a integrita
- Všechny importy jsou **read-only** vůči zdroji, bez skryté synchronizace a bez změny Google listu. Vyžadovat uživatelský souhlas ke čtení Google účtu. Omezit typy, velikosti a počty řádků; dávky zpracovávat inkrementálně a s možností přerušení.
- Neprovádět makra, vzorce, skripty ani CSV formula injection na exportních CSV; při tvorbě CSV výstupu escapovat buňky začínající `=`, `+`, `-` a `@` dle nastavené bezpečnostní politiky. Tabulka je **data**, nikoli kód.
- Zachovat `/AcroForm`, `/Fields`, `/Widget`, `/V`, `/AP`, strukturu pojmenovaných polí, podpisy, odkazy a kvalitu písem včetně české diakritiky; generovat nebo aktualizovat vzhled formuláře v souladu s PDF rendererem a testovat restart a otevření v Acrobat/Foxit/PDF-XChange.
- Před přepsáním existujícího výstupu explicitní souhlas. Originální PDF a zdroj tabulky nikdy nepřepisovat. Transakční temp soubor + atomické dokončení jednotlivého výstupu, počítání hashů a souhrn chyb.
- Při importu z Google API používat minimální autorizaci, nesdílet OAuth tokeny v HTML ani v GitHubu, respektovat omezení a expiraci tokenů.

## Regresní testy
CSV s UTF-8 BOM, středníkem, čárkou v uvozovkách, vnořenými uvozovkami, CRLF, multiline buňkou, vedoucími nulami a diakritikou; XLSX datum/formule/skrytý řádek; Google různá oprávnění a zrušení OAuth; chybějící/zaměnitelné hlavičky; checkbox exportní „Yes/Off“, opakovaná pole, fonty, chybná data, náhled, rollback, restart, export bez ztráty anotací a podepsaných částí.

### Odkazy pro implementaci
- CSV standard: https://datatracker.ietf.org/doc/rfc4180/
- Sheets values GET: https://developers.google.com/workspace/sheets/api/reference/rest/v4/spreadsheets.values/get
- Google OAuth scopes: https://developers.google.com/workspace/sheets/api/scopes
- Foxit data import/export: https://www.foxit.com/resource-hub/user-guide/foxit-pdf-editor-for-windows/importexport-form-data/
- PDF-XChange Populate Forms: https://help.pdf-xchange.com/pdfxt10/populate-forms_ed.html
