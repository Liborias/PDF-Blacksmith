# PDF Blacksmith — lokalizační zdroje pro Codex

**Platné pracovní zdroje:** [ui_catalog_v1.json](ui_catalog_v1.json) (strojově čitelný) a [ui_catalog_v1.csv](ui_catalog_v1.csv) (tabulka pro překladatele). Jde o první export ze samostatného offline HTML prototypu [v3.18](../rozhrani/cely_vizual_v3.18.html); **produkční aplikace zatím z těchto souborů překlady nenačítá**.

## 1. Formát a struktura

- **JSON**: \`schema\`, \`version\`, \`updated\`, \`supported_locales\`, \`fallback_locale\`, \`entries\`; u každého klíče \`cs\`, \`en\`, \`context\`, \`status\`.
- **CSV**: přesně stejné záznamy v pěti sloupcích \`id, cs-CZ, en-US, context, status\`. Pro nový jazyk přidat např. \`de-DE\`, \`es-ES\`, \`fr-FR\`, \`pl-PL\`, \`sk-SK\` do CSV a poté generovat odpovídající překlady ve strukturovaném JSON.
- V1 obsahuje **783 řádků** a slouží jako převodový most z již existujícího českého→anglického slovníku. Nepovažovat některé shodné anglické/české hodnoty za chybné, ale označit je k případné kontrole.

## 2. Zásadní pravidlo identifikátorů

Současné ID je deterministický otisk původního českého textu, například \`ui.soubor.5ee136c77a\`. **Po převzetí ID do aplikace už jeho hodnotu nepřepočítávat ze znění češtiny.** To by při změně formulace rozbilo návaznost dalších překladů. Postupně nahradit vazbu na textové originály sémantickými klíči, např. \`toolbar.file\`, \`guide.group.collapse\`, \`image.crop.circle\`, \`settings.theme.light\`; jednou stabilně namapovat stávající ID.

## 3. Implementační doporučení (Go/Win32)

- Jeden lokalizační modul s loaderem katalogu a \`Translate(key, params)\`/fallback \`cs-CZ\`; žádná přímá translatace DOM textů v produkčním UI.
- Dialogy, nabídky, popisky ikon, tooltipy, chybové hlášky, stavy, dynamické parametry a přístupné názvy (\`aria-label\` v HTML) musí mít vlastní stabilní překladové klíče.
- Podporovat plurály (1 vodítko / 2 vodítka / 5 vodítek), lokální čísla, datum, jednotky a kontextové varianty. Pozor na \`{count}\`, \`{name}\`, zkratky a zástupné hodnoty; zkontrolovat soulad placeholders.
- Oddělit **jazyk aplikace** od **jazyka OCR**, obsahu PDF, uživatelských poznámek, názvů souborů a dokumentových záložek. Tato data se automaticky nepřekládají.
- Uživatel může kdykoliv měnit jazyk v Nastavení → Jazyk. Aktualizace textu UI nesmí měnit příkazová ID, klávesové zkratky ani zvolený nástroj.
- V testech ověřovat úplnost katalogu, prázdné texty, duplicitní ID, neznámé překlady, nekompatibilní zástupné argumenty a přetékání ovládacích prvků při delších překladech.
- Větvení lokalizací pro jiné jazyky nesmí duplikovat aplikaci ani měnit grafický/editační obsah PDF.

## 4. Validace před nasazením

1. Přidat novou jazykovou mutaci do katalogu a validovat všechny klíče + placeholders.
2. Vyzkoušet primárně Nastavení, výběrové nástroje, Pomůcky rozvržení, Separace, OCR, Objekty/Ořez, rekonstrukci textu, tisk, chybové dialogy, tooltipy.
3. Změnit jazyk a následně otevřít PDF obsahující jiný jazyk; obsah dokumentu musí zůstat 1:1 beze změny.
4. Ověřit uložení jazyka mezi spuštěními, přístupnost, české a anglické rozložení kláves, barvy a světlý režim.
5. Veškeré změny přidat do specifikace Codex a testů; neprovádět automatickou implementaci Go/PDFium bez schválení.

**Poznámka k licenci:** Katalog je autorské pracovní zadání uživatelského rozhraní, nikoliv slovník převzatý z externího komerčního produktu. Pro širší komunitní překlady lze používat CSV jako vstup do webových překladových systémů, pokud jejich bezpečnostní a licenční podmínky odpovídají použití.
