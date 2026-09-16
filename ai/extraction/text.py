import logging

from ai.extraction.base import BaseExtractor

logger = logging.getLogger(__name__)


class TextExtractor(BaseExtractor):
    """Extract text from plain text files."""

    def extract(self, file_path: str) -> str:
        """Extract text from TXT file."""
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                text = f.read()
            logger.info(f"Extracted {len(text)} chars from TXT")
            return text
        except Exception as e:
            logger.error(f"Error extracting TXT: {e}")
            raise