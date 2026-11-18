import json
from pathlib import Path
import shutil
import subprocess
import sys
from tempfile import TemporaryDirectory

from pikepdf import Array, Dictionary, Name, Pdf, Permissions, Encryption
from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives.serialization import pkcs12
from cryptography.x509.oid import NameOID
from datetime import datetime, timezone, timedelta

def run_smoke(base, exe):
    project = Path(__file__).resolve().parents[1]
    src = base / 'barevný vzor.pdf'
    out = base / 'šedý výstup.pdf'
    encr = base / 'chráněný.pdf'
    plain = base / 'bez hesla.pdf'
    signed = base / 'podepsaný.pdf'
    signed2 = base / 'podepsaný dvakrát.pdf'
    tampered = base / 'změněný.pdf'
    cert_file = base / 'klíč.p12'
    trust = base / 'cert.pem'

    with Pdf.new() as pdf:
        font = pdf.make_indirect(Dictionary(Type=Name.Font, Subtype=Name.Type1, BaseFont=Name.Helvetica))
        image = pdf.make_stream(
            bytes([255, 0, 0, 0, 255, 0, 0, 0, 255, 0, 0, 0]),
            Type=Name.XObject, Subtype=Name.Image, Width=2, Height=2,
            BitsPerComponent=8, ColorSpace=Name.DeviceRGB,
        )
        for page_number in range(11):
            page = pdf.add_blank_page(page_size=(220, 245))
            spot = Array([
                Name.Separation, Name.TestSpot, Name.DeviceCMYK,
                Dictionary(FunctionType=2, Domain=Array([0, 1]), C0=Array([0, 0, 0, 0]),
                           C1=Array([1, 0, 0, 0]), N=1)])
            page.obj.Resources = Dictionary(
                Font=Dictionary(F1=font), ColorSpace=Dictionary(Spot=spot),
                ExtGState=Dictionary(Alpha=Dictionary(ca=0.5)),
                XObject=Dictionary(Im=image))
            page.obj.Contents = pdf.make_stream(b"""
              1 0 0 rg 20 40 80 60 re f
              0 1 1 0 k 40 70 80 60 re f
              /Spot cs 0.6 scn 120 70 80 60 re f
              q /Alpha gs 0 0 1 rg 40 100 80 60 re f Q
              q 55 0 0 55 140 135 cm /Im Do Q
              0 g BT /F1 19 Tf 30 160 Td (Text ke krivkam) Tj ET
            """)
            page.obj.TrimBox = Array([10, 10, 210, 235])
            page.obj.BleedBox = Array([5, 5, 215, 240])
            if page_number == 10:
                page.obj.Rotate = 90
        pdf.save(src, compress_streams=False)
        pdf.save(encr, encryption=Encryption(owner='majitel123', user='heslo123',
                 allow=Permissions(print_highres=False)))

    key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    name = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, 'Zkušební podpis – Jiří')])
    now = datetime.now(timezone.utc)
    cert = (x509.CertificateBuilder().subject_name(name).issuer_name(name).public_key(key.public_key())
            .serial_number(x509.random_serial_number()).not_valid_before(now - timedelta(days=1))
            .not_valid_after(now + timedelta(days=1))
            .add_extension(x509.BasicConstraints(ca=True, path_length=None), critical=True)
            .add_extension(x509.KeyUsage(digital_signature=True, content_commitment=True,
                key_encipherment=False, data_encipherment=False, key_agreement=False,
                key_cert_sign=True, crl_sign=True, encipher_only=False, decipher_only=False),
                critical=True).sign(key, hashes.SHA256()))
    cert_file.write_bytes(pkcs12.serialize_key_and_certificates(
        b'Test pdf', key, cert, None, serialization.BestAvailableEncryption(b'cert-heslo')))
    trust.write_bytes(cert.public_bytes(serialization.Encoding.PEM))

    def run(action, file, status='ok', **options):
        request = {'action': action, 'input': str(file), 'overwrite': True, **options}
        result = subprocess.run([sys.executable, str(project / "engine.py")],
                                input=json.dumps(request), check=True, text=True, capture_output=True)
        result = json.loads(result.stdout)
        print(action, file.name, result, flush=True)
        assert result['status'] == status, result
        return result['message']

    run('transform', encr, 'error', output=str(plain), decrypt=True, pdf_password='bad')
    run('transform', encr, output=str(plain), decrypt=True, pdf_password='heslo123')
    with Pdf.open(plain) as pdf:
        assert not pdf.is_encrypted

    run('transform', src, output=str(out), ghostscript=exe, gray=True, outline=True)
    with Pdf.open(out) as result, Pdf.open(src) as original:
        assert len(result.pages) == len(original.pages)
        for a, b in zip(list(result.pages)[:-1], list(original.pages)[:-1]):
            assert (a.mediabox == b.mediabox and a.trimbox == b.trimbox
                    and a.bleedbox == b.bleedbox), (a.obj, b.obj)
        assert list(result.pages[-1].mediabox) == [0,0,245,220]
    assert subprocess.check_output(['pdftotext', str(out), '-']).strip() == b''
    assert b'Text ke krivkam' in subprocess.check_output(['pdftotext', str(src), '-'])

    only_outline = base/'jen křivky.pdf'
    run('transform', src, output=str(only_outline), ghostscript=exe, outline=True)
    assert subprocess.check_output(['pdftotext', str(only_outline), '-']).strip() == b''
    assert ' CMYK' in subprocess.check_output([exe, '-q', '-dBATCH', '-dNOPAUSE', '-sDEVICE=inkcov',
                                                '-dLastPage=1', '-o', '-', str(only_outline)], text=True)

    run('sign', src, output=str(signed), certificate=str(cert_file), cert_password='cert-heslo', visible=True)
    assert signed.read_bytes().startswith(src.read_bytes())
    report = run('verify', signed, trust=str(trust))
    assert 'Kryptograficky platný  ano' in report
    assert 'Důvěryhodný řetězec certifikátů  ano' in report
    assert 'Integrita podpisu  ano' in report
    assert 'Výsledek offline kontroly  ověřeno' in report
    assert 'Důvěryhodný řetězec certifikátů  ne' in run('verify', signed)
    tampered.write_bytes(signed.read_bytes().replace(b'Text ke krivkam', b'text ke krivkam', 1))
    assert tampered.read_bytes() != signed.read_bytes()
    assert 'Integrita podpisu  ne' in run('verify', tampered, trust=str(trust))
    run('sign', signed, output=str(signed2), certificate=str(cert_file), cert_password='cert-heslo')
    report = run('verify', signed2, trust=str(trust))
    assert report.count('Integrita podpisu  ano') == 2
    assert 'dřívější úplná revize' in report

    run('transform', signed, 'signed_pdf', output=str(out), ghostscript=exe, gray=True)
    run('transform', signed, output=str(out), ghostscript=exe, gray=True, clear_signatures=True)
    assert 'nebyl nalezen' in run('verify', out)
    print('All smoke tests passed.', flush=True)


if __name__ == "__main__":
    ghostscript = shutil.which(sys.argv[1] if len(sys.argv) > 1 else "gs")
    if not ghostscript or not shutil.which("pdftotext"):
        sys.exit("Test vyžaduje Ghostscript a pdftotext z balíku Poppler.")
    with TemporaryDirectory(prefix="pdf-blacksmith-test-") as folder:
        run_smoke(Path(folder), str(Path(ghostscript).resolve()))
