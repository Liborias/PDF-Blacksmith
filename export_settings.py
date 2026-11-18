"""Validated export options shared by the PDF export and its plate check."""

DEFAULTS = dict(intent=-1, black_point=-1, image_dpi=0, compression="auto",
                proof_dpi=150, repair_dpi=600)
VALUES = dict(intent=(-1, 0, 1, 2, 3), black_point=(-1, 0, 1),
              image_dpi=(0, 150, 300, 600), compression=("auto", "flate"),
              proof_dpi=(75, 150, 300), repair_dpi=(300, 600))


def export_settings(raw):
    raw = raw or {}
    if not isinstance(raw, dict):
        raise ValueError("Nastavení exportu má neplatný formát.")
    options = DEFAULTS.copy()
    for key, allowed in VALUES.items():
        if key in raw:
            value = raw[key]
            if type(value) is not type(allowed[0]) or value not in allowed:
                raise ValueError(f"Neplatný parametr exportu: {key}.")
            options[key] = value
    return options


def pdfwrite_settings(options):
    args = []
    if options["intent"] != -1:
        args.append(f"-dRenderIntent={options['intent']}")
    if options["black_point"] != -1:
        args.append(f"-dBlackPtComp={options['black_point']}")
    for channel in ("Color", "Gray", "Mono"):
        if options["image_dpi"]:
            args += [f"-dDownsample{channel}Images=true",
                     f"-d{channel}ImageDownsampleThreshold=1.0",
                     f"-d{channel}ImageResolution={options['image_dpi']}",
                     f"-d{channel}ImageDownsampleType=/{'Subsample' if channel == 'Mono' else 'Bicubic'}"]
        if options["compression"] == "flate":
            if channel != "Mono":
                args.append(f"-dAutoFilter{channel}Images=false")
            args.append(f"-d{channel}ImageFilter=/FlateEncode")
    if options["compression"] == "flate" or options["image_dpi"]:
        args += ["-dPassThroughJPEGImages=false", "-dPassThroughJPXImages=false"]
    return args
