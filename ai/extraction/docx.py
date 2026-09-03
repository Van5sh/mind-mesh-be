from extraction.base import BaseExtractor


class DocxExtractor(BaseExtractor):
    def extract(self, file_path: str) -> str:
        from docx import Document

        doc = Document(file_path)
        text = "\n".join([para.text for para in doc.paragraphs])
        return text

    def extract_from_bytes(self, file_bytes: bytes) -> str:
        from io import BytesIO

        from docx import Document

        with BytesIO(file_bytes) as file_stream:
            doc = Document(file_stream)
            text = "\n".join([para.text for para in doc.paragraphs])
            return text