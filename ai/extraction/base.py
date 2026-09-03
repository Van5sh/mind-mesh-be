from abc import ABC, abstractmethod
from dataclasses import dataclass
from pathlib import Path


@dataclass
class ExtractedDocument:
    text: str
    filename: str
    content_type: str | None = None


class BaseExtractor(ABC):
    extensions: tuple[str, ...] = ()

    @abstractmethod
    def extract(self, path: Path) -> ExtractedDocument:
        raise NotImplementedError