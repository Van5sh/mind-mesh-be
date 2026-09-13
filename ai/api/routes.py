import logging

from fastapi import APIRouter, HTTPException, status

from ai.api.models import (
    ChatQuestionRequest,
    ChatQuestionResponse,
    HealthResponse,
)
from ai.database.postgres import PostgresClient
from ai.retrieval.rag import RAGChatService

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/v1", tags=["chat"])

postgres_client = PostgresClient()
rag_service = RAGChatService()

# File processing itself is not routed through this API: the Go backend
# enqueues jobs to SQS directly (see internal/services/aws/sqs.service.go)
# and the worker (ai/worker/) writes status straight to Postgres, which Go
# reads directly too. This service only exists for requests that need a
# real-time answer - currently chat.


@router.post(
    "/chat",
    response_model=ChatQuestionResponse,
    summary="Answer a question about a project's documents",
)
async def ask_question(request: ChatQuestionRequest):
    """
    Answer a question grounded in the project's uploaded documents:
    embed the question, retrieve the most relevant chunks from Qdrant
    (scoped to this project), and ask the LLM to answer using only that
    context.

    Called synchronously by the Go backend when a user sends a message
    in an AI_ASSISTANT chat; the returned answer is persisted there as
    the AI's reply.
    """
    try:
        result = rag_service.answer_question(
            question=request.question,
            project_id=request.project_id,
        )
        return ChatQuestionResponse(**result)
    except Exception as e:
        logger.error(f"Error answering chat question: {e}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e),
        )


@router.get(
    "/health",
    response_model=HealthResponse,
    summary="Health check endpoint",
)
async def health_check():
    """Health check for the AI API service (Postgres connectivity)."""
    try:
        postgres_client.check_connectivity()
        return HealthResponse(status="healthy")
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Service unhealthy",
        )
