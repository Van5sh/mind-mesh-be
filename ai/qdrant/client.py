from qdrant_client import QdrantClient
from qdrant_client.models import PointStruct

from config.settings import get_settings


class QdrantService:

    def __init__(self) -> None:
        settings = get_settings()

        self.client = QdrantClient(
            url=settings.qdrant_url,
            api_key=settings.qdrant_api_key,
        )

        self.collection_name = settings.qdrant_collection

    def upsert(
        self,
        points: list[PointStruct],
    ) -> None:
        self.client.upsert(
            collection_name=self.collection_name,
            points=points,
        )

    def delete_by_file_id(self, file_id: str) -> None:
        self.client.delete(
            collection_name=self.collection_name,
            points_selector={
                "filter": {
                    "must": [
                        {
                            "key": "file_id",
                            "match": {
                                "value": file_id,
                            },
                        }
                    ]
                }
            },
        )