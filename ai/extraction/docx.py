from pathlib import Path
from docx import Document
from .base import BaseExtractor, ExtractedDocument


class DocxExtractor(BaseExtractor):
    extensions = (".docx",)

    def extract(self, path: Path) -> ExtractedDocument:
        doc = Document(str(path))
        parts = []

        for paragraph in doc.paragraphs:
            if paragraph.text.strip():
                parts.append(paragraph.text.strip())

        for table in doc.tables:
            for row in table.rows:
                parts.append(" | ".join(cell.text.strip() for cell in row.cells))

        return ExtractedDocument(
            text="\n".join(parts),
            filename=path.name,
            content_type=(
                "application/vnd.openxmlformats-officedocument."
                "wordprocessingml.document"
            ),
        )