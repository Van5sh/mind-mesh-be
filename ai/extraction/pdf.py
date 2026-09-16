import logging

from pypdf import PdfReader

from ai.extraction.base import BaseExtractor

logger = logging.getLogger(__name__)


class PDFExtractor(BaseExtractor):
    """Extract text from PDF files."""

    def extract(self, file_path: str) -> str:
        """Extract text from PDF."""
        try:
            reader = PdfReader(file_path)
            text_parts = []

            for page_num, page in enumerate(reader.pages):
                text = page.extract_text()
                if text:
                    text_parts.append(text)

            full_text = "\n".join(text_parts)
            logger.info(f"Extracted {len(full_text)} chars from PDF")
            return full_text

        except Exception as e:
            logger.error(f"Error extracting PDF: {e}")
            raise