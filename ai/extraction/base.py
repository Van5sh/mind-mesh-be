from abc import ABC, abstractmethod


class BaseExtractor(ABC):
    @abstractmethod
    def extract(self, file_path: str) -> str:
        raise NotImplementedError("Subclasses must implement the extract method.")

    @abstractmethod
    def extract_from_bytes(self, file_bytes: bytes) -> str:
        raise NotImplementedError("Subclasses must implement the extract_from_bytes method.")