from functools import lru_cache
from .base import EmbeddingProvider


class SentenceTransformerEmbeddings(EmbeddingProvider):
    def __init__(self, model_name: str, batch_size: int = 32):
        from sentence_transformers import SentenceTransformer

        self.model = SentenceTransformer(model_name)
        self.batch_size = batch_size
        self._dimension = self.model.get_sentence_embedding_dimension()

    @property
    def dimension(self) -> int:
        return int(self._dimension)

    def embed_documents(self, texts: list[str]) -> list[list[float]]:
        if not texts:
            return []
        vectors = self.model.encode(
            texts,
            batch_size=self.batch_size,
            normalize_embeddings=True,
            show_progress_bar=False,
        )
        return vectors.tolist()

    def embed_query(self, text: str) -> list[float]:
        return self.embed_documents([text])[0]


@lru_cache
def get_embedding_provider(model_name: str, batch_size: int) -> EmbeddingProvider:
    return SentenceTransformerEmbeddings(model_name, batch_size)