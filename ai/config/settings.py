from functools import lru_cache
from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict

# ai/.env, anchored to this file's location rather than left as a bare
# ".env" - a relative path resolves against the process's current working
# directory, which silently picks up the *Go backend's* root .env instead
# of this one if uvicorn/python is ever launched from the repo root
# instead of from inside ai/ (e.g. `uv run --project ai uvicorn ai.main:app`
# run from mind-mesh-be/, which is required for the "ai.main:app" import
# string to resolve at all - see BACKEND_HANDOFF.md).
ENV_FILE = Path(__file__).resolve().parent.parent / ".env"


class Settings(BaseSettings):
    # AWS
    AWS_REGION: str = "us-east-1"
    S3_BUCKET: str
    SQS_QUEUE_URL: str

    # Database
    DB_HOST: str
    DB_PORT: int = 5432
    DB_USER: str
    DB_PASSWORD: str
    DB_NAME: str

    # Qdrant
    QDRANT_URL: str = "http://localhost:6333"
    QDRANT_API_KEY: str = ""
    QDRANT_COLLECTION: str = "documents"

    # Ollama - still used for embeddings (ai/embeddings/text.py) and the
    # worker's file summarization step (ai/summarization/summarizer.py).
    OLLAMA_BASE_URL: str = "http://localhost:11434"
    OLLAMA_MODEL: str = "mistral"
    OLLAMA_EMBEDDING_MODEL: str = "nomic-embed-text"

    # Groq - chat LLM for RAG question-answering (ai/llm/provider.py) only.
    # Optional (not required at startup) so the worker, which never touches
    # Groq, doesn't need this set just because it shares this Settings
    # class with ai-api.
    GROQ_API_KEY: str = ""
    GROQ_MODEL: str = "openai/gpt-oss-120b"

    # API
    API_HOST: str = "0.0.0.0"
    API_PORT: int = 8000

    model_config = SettingsConfigDict(
        env_file=ENV_FILE,
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )


settings = Settings()


@lru_cache
def get_settings() -> Settings:
    return Settings()