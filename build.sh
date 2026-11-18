#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "$0")"
build_dir="$(mktemp -d)"
trap 'rm -rf -- "$build_dir"' EXIT
output_path="$(realpath -m -- "${1:-PDFBlaksmith.exe}")"
python_zip="$build_dir/python-win.zip"
gs_installer="$build_dir/gs-win.exe"
runtime_dir="$build_dir/runtime"
mkdir -p -- "$runtime_dir/packages" "$(dirname -- "$output_path")"

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
python3 package_runtime.py "$runtime_dir"
go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -manifest app.manifest -ico assets/blaksmith.ico \
  -o rsrc_windows_amd64.syso
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -trimpath \
  -ldflags '-H=windowsgui -s -w' -o "$output_path" .
