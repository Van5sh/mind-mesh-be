from extraction.base import BaseExtractor


class TxtExtractor(BaseExtractor):
    def extract(self, file_path: str) -> str:
        with open(file_path, "r", encoding="utf-8", errors="replace") as file:
            return file.read()

    def extract_from_bytes(self, file_bytes: bytes) -> str:
        return file_bytes.decode("utf-8", errors="replace")