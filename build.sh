#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "$0")"
build_dir="$(mktemp -d)"
trap 'rm -rf -- "$build_dir"' EXIT
output_path="$(realpath -m -- "${1:-PDFBlaksmith.exe}")"
python_zip="$build_dir/python-win.zip"
gs_installer="$build_dir/gs-win.exe"
runtime_dir="$build_dir/runtime"
mkdir -p -- "$runtime_dir/packages" "$runtime_dir/tessdata" "$runtime_dir/text-fonts" "$(dirname -- "$output_path")"

curl --fail --silent --show-error --location \
  'https://www.python.org/ftp/python/3.14.10/python-3.14.10-embed-amd64.zip' -o "$python_zip"
printf '%s  %s\n' '90905e64a840557d8c1bc4f5b74db3d7ee85e3ffb217611cac1c7813acea4159' "$python_zip" | sha256sum -c -
unzip -q "$python_zip" -d "$runtime_dir"
python3 -m pip install --no-deps --only-binary=:all: --platform win_amd64 --python-version 3.14 \
  --implementation cp --target "$runtime_dir/packages" -r requirements-win.txt
curl --fail --silent --show-error --location \
  'https://github.com/ArtifexSoftware/ghostpdl-downloads/releases/download/gs10080/gs10080w64.exe' -o "$gs_installer"
printf '%s  %s\n' '52a91b8bf09298788d7a57b9206127026c23eacd75405f0a131e26dc381dce50' "$gs_installer" | sha256sum -c -
"$(command -v 7z)" x -y "-o$runtime_dir/ghostscript" "$gs_installer" \
  'bin/gswin64c.exe' 'bin/gsdll64.dll' 'Resource/*' 'lib/*' 'iccprofiles/*' 'doc/COPYING' >/dev/null
cp "$runtime_dir"/packages/pikepdf.libs/msvcp140-*.dll "$runtime_dir/ghostscript/bin/msvcp140.dll"
cp "$runtime_dir"/vcruntime140*.dll "$runtime_dir/ghostscript/bin/"
cp GO-LICENSE.txt "$runtime_dir/"
ocr_models='https://raw.githubusercontent.com/tesseract-ocr/tessdata_fast/87416418657359cb625c412a48b6e1d6d41c29bd'
for file in ces.traineddata eng.traineddata LICENSE; do
  curl --fail --silent --show-error --location "$ocr_models/$file" -o "$runtime_dir/tessdata/$file"
done
(
  cd "$runtime_dir/tessdata"
  sha256sum -c <<'EOF'
934bcaf97ef3348413263331131c9fa7f55f30db333c711929c124fb635f7e1b  ces.traineddata
7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2  eng.traineddata
cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30  LICENSE
EOF
)
font_bundle="$build_dir/fonts.deb"
curl --fail --silent --show-error --location \
  'https://deb.debian.org/debian/pool/main/f/fonts-liberation/fonts-liberation_2.1.5-3_all.deb' -o "$font_bundle"
printf '%s  %s\n' '3d5c47480067b9615f6bf9b0eda6eefcab9d837af609827015445e8e759b5b8c' "$font_bundle" | sha256sum -c -
ar p "$font_bundle" data.tar.xz | tar -xJ -C "$build_dir"
cp "$build_dir"/usr/share/fonts/truetype/liberation/*.ttf "$runtime_dir/text-fonts/"
cp "$build_dir/usr/share/doc/fonts-liberation/copyright" "$runtime_dir/text-fonts/LICENSE.txt"
python3 package_runtime.py "$runtime_dir"
python3 tools/generate_brand_icon.py "$build_dir/blacksmith.ico"
go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -manifest app.manifest -ico "$build_dir/blacksmith.ico" \
  -o rsrc_windows_amd64.syso
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -trimpath \
  -ldflags '-H=windowsgui -s -w' -o "$output_path" .
