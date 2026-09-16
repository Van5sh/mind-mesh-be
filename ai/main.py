import logging

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from ai.api.routes import router

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="MeshMind AI API",
    description="Document processing and AI orchestration API",
    version="1.0.0",
)

# CORS middleware for Go backend integration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routes
app.include_router(router)


@app.on_event("startup")
async def startup():
    logger.info("MeshMind AI API starting up...")


@app.on_event("shutdown")
async def shutdown():
    logger.info("MeshMind AI API shutting down...")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
