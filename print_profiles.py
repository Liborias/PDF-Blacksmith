"""Resolve output profiles and write a PDF output intent."""

from hashlib import sha256
from io import BytesIO
import json
import os
from pathlib import Path
from tempfile import NamedTemporaryFile
from urllib.error import URLError
from urllib.request import urlopen
from zipfile import ZipFile

import pikepdf


MAX_ICC_BYTES = 32_000_000
HERE = Path(__file__).resolve().parent


class ProfileError(ValueError):
    pass


def icc_matches(data, space):
    return (128 <= len(data) <= MAX_ICC_BYTES
            and int.from_bytes(data[:4], "big") == len(data)
            and data[16:20] == space and data[36:40] == b"acsp")


def cmyk_icc(data):
    return icc_matches(data, b"CMYK")


def profile_cache():
    if os.name == "nt":
        root = Path(os.environ["LOCALAPPDATA"])
    else:
        root = Path(os.environ.get("XDG_CACHE_HOME", Path.home() / ".cache"))
    return root / "PDF Blaksmith" / "profiles"


def sha_matches(data, expected):
    return sha256(data).hexdigest() == expected


def cache_file(path, data):
    with NamedTemporaryFile(dir=path.parent, prefix="icc-", delete=False) as file:
        pending = Path(file.name)
        file.write(data)
    try:
        os.replace(pending, path)
    finally:
        pending.unlink(missing_ok=True)


def download_profiles(selected, catalog, cache):
    try:
        with urlopen(selected["archive"], timeout=30) as response:
            data = response.read(MAX_ICC_BYTES + 1)
    except (URLError, TimeoutError) as exc:
        raise ProfileError("Profil ECI se nepodařilo stáhnout. Zkontrolujte připojení "
                           "nebo vyberte vlastní soubor ICC.") from exc
    if len(data) > MAX_ICC_BYTES or not sha_matches(data, selected["archive_sha256"]):
        raise ProfileError("Stažený archiv ECI se liší od ověřené verze. "
                           "Profil lze stáhnout z webu ECI a vybrat jako vlastní ICC.")
    cache.mkdir(parents=True, exist_ok=True)
    with ZipFile(BytesIO(data)) as archive:
        for item in catalog:
            if item["archive"] == selected["archive"]:
                profile = archive.read(item["member"])
                if not cmyk_icc(profile) or not sha_matches(profile, item["sha256"]):
                    raise ProfileError("Kontrolní součet profilu ECI nesouhlasí.")
                cache_file(cache / (item["key"] + ".icc"), profile)


def get_profile(key, custom):
    if key == "custom":
        path = Path(custom)
        if not path.is_file() or path.stat().st_size > MAX_ICC_BYTES:
            raise ProfileError("Vyberte soubor s výstupním profilem CMYK ICC nebo ICM.")
        data = path.read_bytes()
        if not (cmyk_icc(data) and data[12:16] == b"prtr"):
            raise ProfileError("Vlastní profil musí být výstupní profil tiskárny v CMYK (.icc nebo .icm).")
        return data, path.stem, path.stem

    catalog = json.loads((HERE / "profiles.json").read_text(encoding="utf-8"))
    selected = next((item for item in catalog if item["key"] == key), None)
    if selected is None:
        raise ProfileError("Vyberte výstupní profil.")
    cache = profile_cache()
    path = cache / (key + ".icc")
    try:
        data = path.read_bytes()
    except FileNotFoundError:
        data = b""
    if not sha_matches(data, selected["sha256"]):
        download_profiles(selected, catalog, cache)
        data = path.read_bytes()
    return data, selected["name"], selected["condition"]


def source_icc(pdf, space):
    for intent in pdf.Root.get("/OutputIntents", []):
        profile = intent.get("/DestOutputProfile")
        if isinstance(profile, pikepdf.Stream):
            data = profile.read_bytes()
            if icc_matches(data, space):
                return data
    return None


def source_cmyk(pdf):
    return source_icc(pdf, b"CMYK")


def attach_srgb(pdf, data):
    stream = pdf.make_stream(data, N=3, Alternate=pikepdf.Name.DeviceRGB)
    default_rgb = pikepdf.Array([pikepdf.Name.ICCBased, stream])
    for page in pdf.pages:
        if "/Resources" not in page.obj:
            page.obj.Resources = pikepdf.Dictionary()
        resources = page.obj.Resources
        if "/ColorSpace" not in resources:
            resources.ColorSpace = pikepdf.Dictionary()
        resources.ColorSpace.DefaultRGB = default_rgb
    if "/OutputIntents" in pdf.Root:
        del pdf.Root["/OutputIntents"]
    clear_pdfx_claim(pdf)


def attach_intent(pdf, data, label, condition):
    stream = pdf.make_stream(data, N=4, Alternate=pikepdf.Name.DeviceCMYK)
    intent = pikepdf.Dictionary(
        Type=pikepdf.Name.OutputIntent, S=pikepdf.Name.GTS_PDFX,
        OutputConditionIdentifier=condition, OutputCondition=label,
        Info=label, DestOutputProfile=stream,
    )
    if condition.startswith("FOGRA") and condition[5:].isdigit():
        intent.RegistryName = "http://www.color.org"
    pdf.Root.OutputIntents = pikepdf.Array([pdf.make_indirect(intent)])
    clear_pdfx_claim(pdf)


def clear_pdfx_claim(pdf):
    # An output intent alone does not certify the document against a PDF/X standard.
    for key in ("/GTS_PDFXVersion", "/GTS_PDFXConformance"):
        if key in pdf.docinfo:
            del pdf.docinfo[key]
    if "/Metadata" in pdf.Root:
        with pdf.open_metadata(set_pikepdf_as_editor=False, update_docinfo=False) as metadata:
            for key in ("pdfxid:GTS_PDFXVersion", "pdfxid:GTS_PDFXConformance",
                        "pdfx:GTS_PDFXVersion", "pdfx:GTS_PDFXConformance"):
                if key in metadata:
                    del metadata[key]
