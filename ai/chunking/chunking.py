from qdrant_client import QdrantClient as QdrantSDKClient
from config.settings import get_settings

class QdrantClient:
    def __init__(self):
        settings = get_settings()
        self.client = QdrantSDKClient(
            url=settings.qdrant_url,
            api_key=settings.qdrant_api_key,
        )
    def health_check(self) -> bool:
        try:
            self.client.get_collections()
            return True
        except Exception:
            return False