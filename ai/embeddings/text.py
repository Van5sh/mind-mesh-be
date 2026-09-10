import logging

from langchain.embeddings import OllamaEmbeddings

from ai.config.settings import settings
from ai.embeddings.base import BaseEmbedder

logger = logging.getLogger(__name__)


class TextEmbedder(BaseEmbedder):
    """Generate embeddings using Ollama."""

    def __init__(self):
        self.embedder = OllamaEmbeddings(
            model=settings.OLLAMA_EMBEDDING_MODEL,
            base_url=settings.OLLAMA_BASE_URL,
        )

    def embed(self, text: str) -> list[float]:
        """Generate embedding for text."""
        try:
            embedding = self.embedder.embed_query(text)
            logger.debug(f"Generated embedding of dimension {len(embedding)}")
            return embedding
        except Exception as e:
            logger.error(f"Error generating embedding: {e}")
            raise