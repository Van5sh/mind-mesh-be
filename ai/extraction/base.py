from abc import ABC, abstractmethod


class BaseExtractor(ABC):
    """Base class for document text extractors."""

    @abstractmethod
    def extract(self, file_path: str) -> str:
        """
        Extract text from file.

        Args:
            file_path: Path to local file

        Returns:
            Extracted text
        """
        pass