# PDF Blacksmith — předání práce ChatGPT
Aktualizace: 28. 11. 2026. Toto je průběžný stav práce, nikoliv další konkurenční produktová specifikace.

## Kde pracovat
- Repo: Liborias/PDF-Blacksmith, vlastní větev `chatgpt/ui-review`.
- Výchozí snapshot větve `codex/ui-palette-v6`: `89a7c4c24d6e531368ed79936e8df22005362b65` (21 commitů před main).
- Před každou změnou ověř HEAD této větve, HEAD Codexu, případné PR a kolize. Lokální nepublikované změny Codexu nejsou přístupné. Bez výslovného souhlasu neměnit main ani větev Codexu, nemergovat, nepoužívat force-push/reset.
- Autoritativní produktové zadání: [AGENTS.md](../../../AGENTS.md), [CODEX_START.md](../CODEX_START.md), [MASTER v6.0](../CODEX_MASTER_v6.0.md), [implementační plán v6.0](../CODEX_IMPLEMENTACNI_PLAN_v6.0.md), finální mapy a [HTML v3.26](../rozhrani/cely_vizual_v3.26.html), vše ve skutečně používané větvi. Historické verze nejsou konkurenční zadání. Pozdější schválené změny zaznamenej sem a dle potřeby do normativního plánu.
- Go/Win32 + PDFium/Python; HTML je reference, ne náhrada aplikace. Zachovat existující příkazy a PDF data, barvy, ICC, separace, podpisy, formuláře, historii. F1 pozastaveno; žádný nový velký modul bez schválení.

## Pracovní režim (schváleno uživatelem 28. 11. 2026)
1. Vždy jeden malý, konkrétní krok, předem pojmenovaný v předchozí odpovědi. Uživatelovo „ok“ je souhlas s právě navrženým dalším krokem, nikoli generální souhlas se všemi dalšími kroky. Upřesnění se nejdříve zapracuje do rozsahu. Nevynucuj další potvrzení stejného kroku.
2. Proveď inventuru/diff, změnu na vlastní větvi, přiměřené lehké testy, commit/push a aktualizuj tento jediný předávací soubor. V reportu uveď soubory, commit, testy a co nebylo ověřeno; navrhni jediný následující krok.
3. EXE vytvoř po kroku 1 (výchozí sestavovací infrastruktura), poté po každých čtyřech dokončených dílčích krocích: 5, 9, 13… Dokumentační commity, opravy téhož kroku a průběžné testy nejsou další samostatný krok. Mezi sestaveními žádná automatická plná kompilace EXE; lehké testy dle potřeby. Oprava neúspěšného sestavení je součástí téhož milníku.
4. Výstup: Windows x64 EXE přes existující build.sh, GitHub Actions artifact, SHA-256 a zdrojový commit. Nepublikovat veřejný release automaticky. V každém předání eviduj skutečný stav sestavení (ne spuštěno / běží / neúspěšné / úspěšné), odkaz a korektury. Windows GUI a integritu PDF je nutné ověřit, samotná kompilace nestačí.
5. Před ukončením každého chatu musí být stav, poslední změny, čítač kroků, neschválený další krok a blokery v tomto souboru; nový agent načte i [START.md](START.md).

## Stav při založení větve
- Audit pouze pro čtení: Codex má nové palety 6×2, ikonový atlas, command bar, nastavení, lokalizaci, docky, přehled stran, výběr, typografii, ořez, vrstvy a základ mřížky. Přítomnost kódu není důkaz funkčnosti. Pravítka/vodítka/SnapEngine vyžadují inventuru. Rozsah F1 nelze považovat za hotový.
- Existující sestavení: `build.sh`, Go 1.26, Python s pip, curl, unzip, 7z, binutils, xz, stažení kontrolovaných runtime závislostí. `SESTAVENI.md`. Složka .github/workflows při auditu nedostupná (404), žádný funkční workflow nebyl potvrzen.
- Krok 1 schválen: vytvořit samostatnou sestavovací infrastrukturu a předání. Stav: rozpracováno. Pracovní EXE zatím není ověřeno.
- Dokončené kroky: 0. Po dokončení infrastruktury evidovat krok 1 a požádat o první sestavení, další plánovaný build po kroku 5.
- Následující návrh kroku 2: audit stávající horní lišty proti HTML v3.26 a jedna nejmenší bezpečná korekce napojení existujícího příkazu nebo jeho vizuálního stavu; přesný rozsah předložit uživateli před implementací. Bez jeho „ok“ neimplementovat.
