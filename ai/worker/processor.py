from dataclasses import dataclass

from extraction.docx import DocxExtractor
from extraction.pdf import PdfExtractor
from extraction.text import TxtExtractor


@dataclass
class ProcessingJob:
    job_id: str
    file_id: str
    project_id: str
    storage_key: str
    mime_type: str


class DocumentProcessor:
    def __init__(self) -> None:
        self.extractors = {
            "application/pdf": PdfExtractor(),
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document": DocxExtractor(),
            "text/plain": TxtExtractor(),
        }

    def process(self, job: ProcessingJob, file_path: str) -> str:
        extractor = self._get_extractor(job.mime_type)

        text = extractor.extract(file_path)

        if not text.strip():
            raise ValueError(f"No text extracted from file: {job.file_id}")

        return text

    def _get_extractor(self, mime_type: str):
        try:
            return self.extractors[mime_type]
        except KeyError:
            raise ValueError(f"Unsupported file type: {mime_type}")