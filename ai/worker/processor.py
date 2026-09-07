from dataclasses import dataclass

from extraction.docx import DocxExtractor
from extraction.pdf import PDFExtractor
from extraction.text import TextExtractor


@dataclass
class ProcessingJob:
    file_id: str
    file_path: str
    content_type: str


class FileProcessor:

    def __init__(self) -> None:
        self.pdf_extractor = PDFExtractor()
        self.docx_extractor = DocxExtractor()
        self.text_extractor = TextExtractor()

    def process(self, job: ProcessingJob) -> str:
        extractor = self._get_extractor(job.content_type)

        text = extractor.extract(job.file_path)

        if not text.strip():
            raise ValueError(
                f"No text could be extracted from file: {job.file_id}"
            )

        return text

    def _get_extractor(self, content_type: str):
        extractors = {
            "application/pdf": self.pdf_extractor,
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
                self.docx_extractor,
            "text/plain": self.text_extractor,
        }

        extractor = extractors.get(content_type)

        if extractor is None:
            raise ValueError(
                f"Unsupported content type: {content_type}"
            )

        return extractor