from ai.extraction.pdf import PDFExtractor
from ai.extraction.docx import DocxExtractor
from ai.extraction.text import TextExtractor


class ExtractionFactory:
    """Factory to get appropriate extractor based on content type."""

    _extractors = {
        "application/pdf": PDFExtractor(),
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": DocxExtractor(),
        "text/plain": TextExtractor(),
    }

    def get_extractor(self, content_type: str):
        """Get extractor by content type."""
        if content_type not in self._extractors:
            raise ValueError(f"Unsupported content type: {content_type}")
        return self._extractors[content_type]
