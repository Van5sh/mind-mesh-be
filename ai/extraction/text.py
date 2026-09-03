from pathlib import Path
from .base import BaseExtractor, ExtractedDocument


class TextExtractor(BaseExtractor):
    extensions = (".txt", ".md", ".csv", ".json", ".py", ".log")

    def extract(self, path: Path) -> ExtractedDocument:
        text = path.read_text(encoding="utf-8", errors="replace")
        return ExtractedDocument(
            text=text,
            filename=path.name,
            content_type="text/plain",
        )