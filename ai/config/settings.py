from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    # Application
    app_env: str = "development"

    # AWS
    aws_region: str

    # SQS
    sqs_queue_url: str

    # S3
    s3_bucket: str

    # PostgreSQL
    database_url: str

    # Qdrant
    qdrant_url: str
    qdrant_api_key: str | None = None
    qdrant_collection: str = "mindmesh_chunks"

    # Embeddings
    embedding_provider: str = "ollama"
    embedding_model: str = "nomic-embed-text"
    embedding_dimension: int = 768
    ollama_url: str = "http://localhost:11434"

    # LLM
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