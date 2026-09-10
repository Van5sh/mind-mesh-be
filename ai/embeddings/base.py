from abc import ABC, abstractmethod


class BaseEmbedder(ABC):
    """Base class for embedding models."""

    @abstractmethod
    def embed(self, text: str) -> list[float]:
        """
        Generate embedding for text.

        Returns:
            List of floats representing the embedding vector
        """
        pass