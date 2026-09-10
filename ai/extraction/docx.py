import logging
from pathlib import Path
from docx import Document

from ai.extraction.base import BaseExtractor

logger = logging.getLogger(__name__)

class DocxExtractor(BaseExtractor):
    """Extract text from DOCX files."""

    def extract(self, file_path: str) -> str:
        """Extract text from DOCX."""
        path = Path(file_path)
        if not path.exists():
            raise FileNotFoundError(f"File not found: {file_path}")
        if not path.is_file():
            raise ValueError(f"Path is not a file: {file_path}")
        try:
            doc = Document(path)
            text_parts = []

            for paragraph in doc.paragraphs:
                if paragraph.text.strip():
                    text_parts.append(paragraph.text)

            full_text = "\n".join(text_parts)
            logger.info(f"Extracted {len(full_text)} chars from DOCX")
            return full_text

        except Exception as e:
            logger.error(f"Error extracting DOCX: {e}")
            raise