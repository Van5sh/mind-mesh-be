from pathlib import Path
from .base import ExtractedDocument
from .text import TextExtractor
from .pdf import PdfExtractor
from .docx import DocxExtractor

_EXTRACTORS = [PdfExtractor(), DocxExtractor(), TextExtractor()]


def extract_file(path: str | Path) -> ExtractedDocument:
    path = Path(path)
    suffix = path.suffix.lower()

    for extractor in _EXTRACTORS:
        if suffix in extractor.extensions:
            return extractor.extract(path)

    raise ValueError(
        f"Unsupported file type: {suffix or '<none>'}. "
        f"Supported: {sorted({e for x in _EXTRACTORS for e in x.extensions})}"
    )