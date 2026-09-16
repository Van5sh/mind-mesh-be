import logging
import uuid

from qdrant_client import QdrantClient as QdrantBaseClient
from qdrant_client.models import (
    Distance,
    VectorParams,
    PointStruct,
    Filter,
    FieldCondition,
    MatchValue,
)

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

    def _ensure_collection(self, vector_size: int) -> None:
        """
        Create the collection if it doesn't exist yet.

        The vector size is derived from the actual embedding output
        rather than a fixed setting, since it must match whatever
        OLLAMA_EMBEDDING_MODEL actually produces (e.g. nomic-embed-text
        is 768-dim, not the 384 the old default assumed) - a mismatch
        here fails every upsert with a dimension error.
        """
        if self.client.collection_exists(self.collection_name):
            return

        logger.info(
            f"Creating Qdrant collection '{self.collection_name}' "
            f"(dim={vector_size})"
        )
        self.client.create_collection(
            collection_name=self.collection_name,
            vectors_config=VectorParams(
                size=vector_size,
                distance=Distance.COSINE,
            ),
        )

    def delete_document(self, file_id: str) -> None:
        """
        Remove every vector belonging to a file.

        Called before re-storing a file's chunks (reprocessing) and
        should also be called by whatever handles file deletion, so a
        file's embeddings never outlive the file itself or pile up
        as duplicates across repeated processing attempts.
        """
        try:
            if not self.client.collection_exists(self.collection_name):
                return

            self.client.delete(
                collection_name=self.collection_name,
                points_selector=Filter(
                    must=[
                        FieldCondition(
                            key="file_id",
                            match=MatchValue(value=file_id),
                        )
                    ]
                ),
            )
            logger.info(f"Deleted existing vectors for file {file_id}")
        except Exception as e:
            logger.error(f"Error deleting vectors for file {file_id}: {e}")
            raise

    def store_document(
        self,
        file_id: str,
        project_id: str,
        summary: str,
        chunks: list[dict],
    ) -> None:
        """
        Store document chunks as vectors in Qdrant.

        One collection (QDRANT_COLLECTION) holds the embeddings for
        every file, distinguished by the file_id/project_id payload
        fields - it is not one collection per file. Reprocessing a file
        (retry, re-upload) must replace its chunks, not add to them, so
        any existing vectors for this file_id are deleted first.

        Args:
            file_id: File ID
            project_id: Project ID
            summary: Document summary
            chunks: List of {"chunk_index", "text", "embedding"}
        """
        if not chunks:
            logger.warning(f"No chunks to store for file {file_id}")
            return

        try:
            self._ensure_collection(vector_size=len(chunks[0]["embedding"]))
            self.delete_document(file_id)

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
