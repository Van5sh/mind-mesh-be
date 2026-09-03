from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    app_env: str = "development"

    aws_region: str

    sqs_queue_url: str

    s3_bucket: str

    database_url: str

    qdrant_url: str
    qdrant_api_key: str | None = None
    qdrant_collection: str = "mindmesh_chunks"

    embedding_provider: str = "ollama"
    embedding_model: str = "nomic-embed-text"
    embedding_dimension: int = 768
    ollama_url: str = "http://localhost:11434"

    llm_provider: str = "ollama"
    llm_model: str

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )


@lru_cache
def get_settings() -> Settings:
    return Settings()