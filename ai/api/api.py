from fastapi import FastAPI

from config.settings import get_settings


settings = get_settings()

app = FastAPI(
    title="MeshMind AI API",
    version="0.1.0",
)


@app.get("/health")
def health() -> dict[str, str]:
    return {
        "status": "ok",
    }


@app.get("/ready")
def ready() -> dict[str, str]:
    return {
        "status": "ready",
    }