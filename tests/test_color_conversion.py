"""Regression checks for ICC Gray and K-only PDF transformations.

Run from repository root: python -m unittest discover -s tests -p 'test_color_conversion.py'
"""
import unittest
from io import BytesIO

import pikepdf
from PIL import ImageCms

from color_conversion import (
    ColorConversionError, convert_rich_black, gray_gamma_icc,
    gray_to_explicit_k, preserve_solid_black_for_gray, resolve_gray_icc,
)


def document(content):
    pdf = pikepdf.Pdf.new()
    page = pdf.add_blank_page(page_size=(120, 120))
    page.obj.Contents = pdf.make_stream(content)
    return pdf


def operations(pdf):
    return [(list(args), str(operator)) for args, operator
            in pikepdf.parse_content_stream(pdf.pages[0].obj)]


class ColorConversionTests(unittest.TestCase):
    def test_gamma_profiles_are_real_gray_icc(self):
        for gamma in (1.8, 2.2):
            data = gray_gamma_icc(gamma)
            self.assertEqual(int.from_bytes(data[:4], "big"), len(data))
            self.assertEqual(data[16:20], b"GRAY")
            profile = ImageCms.ImageCmsProfile(BytesIO(data))
            self.assertIn("Gray", ImageCms.getProfileName(profile))
        self.assertNotEqual(gray_gamma_icc(1.8), gray_gamma_icc(2.2))

    def test_custom_gray_profile_validation(self):
        self.assertRaises(ColorConversionError, resolve_gray_icc, "other")
        self.assertEqual(resolve_gray_icc("automatic"), (None, "Automaticky"))

    def test_rich_black_preserves_colored_objects(self):
        with document(b"0.4 0.3 0.2 1 k  0.4 0.3 0.2 0.5 K  0.2 0.7 0.3 0 k") as pdf:
            count = convert_rich_black(pdf)
            ops = operations(pdf)
            self.assertEqual(count, 1)
            self.assertEqual(ops[0], ([0, 0, 0, 1], "k"))
            self.assertEqual(tuple(float(x) for x in ops[1][0]), (0.4, 0.3, 0.2, 0.5))
            self.assertEqual(tuple(float(x) for x in ops[2][0]), (0.2, 0.7, 0.3, 0.0))

    def test_preserve_source_k100(self):
        with document(b"0.2 0.3 0.1 1 k  0 0 0 0.5 K") as pdf:
            self.assertEqual(preserve_solid_black_for_gray(pdf), 1)
            self.assertEqual(operations(pdf)[0], ([0], "g"))
            self.assertEqual(operations(pdf)[1][1], "K")

    def test_gray_vectors_to_explicit_k(self):
        with document(b"0 g 0.5 G") as pdf:
            vectors, images = gray_to_explicit_k(pdf)
            self.assertEqual((vectors, images), (2, 0))
            self.assertEqual(operations(pdf)[0], ([0, 0, 0, 1], "k"))
            self.assertEqual(operations(pdf)[1], ([0, 0, 0, 0.5], "K"))

    def test_non_gray_paint_is_refused(self):
        with document(b"0.5 0 0 rg") as pdf:
            self.assertRaises(ColorConversionError, gray_to_explicit_k, pdf)


if __name__ == "__main__":
    unittest.main()
