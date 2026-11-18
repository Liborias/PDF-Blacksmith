# Sestavení PDF Tisk

Sestavení na Linuxu vyžaduje Go 1.26 nebo novější, Python 3 s pipem, curl,
unzip a sha256sum. Spusťte `bash build.sh`. Sestavovací skript stáhne a zabalí
verze knihoven uvedené v `requirements-win.txt` do samostatného programu
`PDFTisk.exe` pro Windows x64. Výstupní cestu lze předat jako první argument.

`main_windows.go` obsahuje uživatelské rozhraní Win32, `runtime_windows.go`
rozbalení součástí a komunikaci s backendem. `engine.py` provádí jednotlivé
operace nad PDF. Aplikace předává vstupy backendu ve formátu JSON přes
standardní vstup. Přehled licencí je v `OSS.txt`.
