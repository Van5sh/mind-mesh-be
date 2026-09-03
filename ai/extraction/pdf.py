from extraction.base import BaseExtractor


class PdfExtractor(BaseExtractor):
    def extract(self, file_path: str) -> str:
        from io import BytesIO

        from pypdf import PdfReader

        with open(file_path, "rb") as file:
            reader = PdfReader(BytesIO(file.read()))
            text = ""
            for page in reader.pages:
                page_text = page.extract_text() or ""
                text += page_text + "\n"
            return text.strip()

    def extract_from_bytes(self, file_bytes: bytes) -> str:
        from io import BytesIO

        from pypdf import PdfReader

        reader = PdfReader(BytesIO(file_bytes))
        text = ""
        for page in reader.pages:
            page_text = page.extract_text() or ""
            text += page_text + "\n"
        return text.strip()
