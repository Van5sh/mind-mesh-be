import logging
import uuid

from qdrant_client import QdrantClient as QdrantBaseClient
from qdrant_client.models import Distance, VectorParams, PointStruct

from ai.config.settings import settings

logger = logging.getLogger(__name__)


class QdrantClient:
    """Client for storing and retrieving vectors from Qdrant."""

    def __init__(self):
        self.client = QdrantBaseClient(
            url=settings.QDRANT_URL,
            api_key=settings.QDRANT_API_KEY,
        )
        self.collection_name = settings.QDRANT_COLLECTION

        # Ensure collection exists
        self._ensure_collection()

    def _ensure_collection(self):
        """Create collection if it doesn't exist."""
        try:
            self.client.get_collection(self.collection_name)
        except Exception:
            logger.info(f"Creating Qdrant collection: {self.collection_name}")
            self.client.create_collection(
                collection_name=self.collection_name,
                vectors_config=VectorParams(
                    size=settings.EMBEDDING_DIMENSION,
                    distance=Distance.COSINE,
                ),
            )

    def store_document(
        self,
        file_id: str,
        project_id: str,
        summary: str,
        chunks: list[dict],
    ) -> None:
        """
        Store document chunks as vectors in Qdrant.

        Args:
            file_id: File ID
            project_id: Project ID
            summary: Document summary
            chunks: List of {"chunk_index", "text", "embedding"}
        """
        try:
            points = []
            for chunk in chunks:
                point_id = str(uuid.uuid4())
                points.append(
                    PointStruct(
                        id=point_id,
                        vector=chunk["embedding"],
                        payload={
                            "file_id": file_id,
                            "project_id": project_id,
                            "chunk_index": chunk["chunk_index"],
                            "text": chunk["text"],
                            "summary": summary,
                        },
                    )
                )

            self.client.upsert(
                collection_name=self.collection_name,
                points=points,
            )
            logger.info(f"Stored {len(points)} vectors for file {file_id}")

        except Exception as e:
            logger.error(f"Error storing vectors in Qdrant: {e}")
            raise