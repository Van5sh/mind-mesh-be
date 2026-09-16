import logging

from qdrant_client.models import Filter, FieldCondition, MatchValue

from ai.embeddings.text import TextEmbedder
from ai.qdrant.client import QdrantClient

logger = logging.getLogger(__name__)


class DocumentSearch:
    """Retrieves the document chunks most relevant to a query, scoped to
    a single project so one project's chat can never surface another
    project's content even though they share one Qdrant collection."""

    def __init__(self):
        self.embedder = TextEmbedder()
        self.qdrant = QdrantClient()

    def search(self, query: str, project_id: str, limit: int = 5) -> list[dict]:
        """
        Args:
            query: The user's question.
            project_id: Restrict results to this project's chunks.
            limit: Max number of chunks to return.

        Returns:
            List of {"file_id", "chunk_index", "text", "score"}, most
            relevant first. Empty list if the collection doesn't exist
            yet (nothing has been embedded) or nothing matches.
        """
        if not self.qdrant.client.collection_exists(self.qdrant.collection_name):
            logger.info("Qdrant collection does not exist yet; no results")
            return []

        try:
            query_vector = self.embedder.embed(query)

            results = self.qdrant.client.query_points(
                collection_name=self.qdrant.collection_name,
                query=query_vector,
                query_filter=Filter(
                    must=[
                        FieldCondition(
                            key="project_id",
                            match=MatchValue(value=project_id),
                        )
                    ]
                ),
                limit=limit,
                with_payload=True,
            ).points

            return [
                {
                    "file_id": point.payload.get("file_id"),
                    "chunk_index": point.payload.get("chunk_index"),
                    "text": point.payload.get("text"),
                    "score": point.score,
                }
                for point in results
            ]
        except Exception as e:
            logger.error(f"Error searching Qdrant: {e}")
            raise
