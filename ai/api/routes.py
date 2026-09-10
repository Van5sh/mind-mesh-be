import json
import logging

from fastapi import APIRouter, HTTPException, status
from botocore.exceptions import ClientError

from ai.api.models import (
    ProcessFileRequest,
    ProcessFileResponse,
    JobStatusResponse,
    HealthResponse,
)
from ai.api.sqs_service import SQSService
from ai.database.postgres import PostgresClient

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/v1", tags=["files"])

sqs_service = SQSService()
postgres_client = PostgresClient()


@router.post(
    "/process",
    response_model=ProcessFileResponse,
    status_code=status.HTTP_202_ACCEPTED,
    summary="Queue a file for processing",
)
async def process_file(request: ProcessFileRequest):
    """
    Queue a file for document processing.
    
    The Go backend should:
    1. Upload file to S3
    2. Call this endpoint with file metadata
    3. Receive job_id for status tracking
    
    The Python worker will then:
    - Download from S3
    - Extract text
    - Generate embeddings
    - Store in Qdrant
    - Update PostgreSQL status
    """
    try:
        logger.info(
            f"Received processing request: job={request.job_id}, file={request.file_id}"
        )

        # Validate content type
        valid_types = [
            "application/pdf",
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            "text/plain",
        ]
        if request.content_type not in valid_types:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Unsupported content type: {request.content_type}",
            )

        message_id = sqs_service.send_job(request)
        logger.info(f"Job queued: {request.job_id} (SQS MessageId: {message_id})")

        return ProcessFileResponse(
            job_id=request.job_id,
            file_id=request.file_id,
            project_id=request.project_id,
            status="QUEUED",
            message=f"File queued for processing. Job ID: {request.job_id}",
        )

    except ClientError as e:
        logger.error(f"AWS error: {e}")
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Failed to queue job. AWS service unavailable.",
        )
    except Exception as e:
        logger.error(f"Error processing file: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e),
        )


@router.get(
    "/status/{file_id}",
    response_model=JobStatusResponse,
    summary="Get processing status of a file",
)
async def get_file_status(file_id: str, project_id: str):
    """
    Get the processing status of a file.
    
    Returns the current status: QUEUED, PROCESSING, COMPLETED, or FAILED.
    """
    try:
        logger.info(f"Fetching status for file={file_id}, project={project_id}")

        status_data = postgres_client.get_file_status(file_id, project_id)

        if not status_data:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"File {file_id} not found",
            )

        return JobStatusResponse(
            job_id=status_data.get("job_id"),
            file_id=status_data.get("file_id"),
            project_id=status_data.get("project_id"),
            status=status_data.get("processing_status"),
            summary=status_data.get("summary"),
            error_message=status_data.get("error_message"),
            processed_at=status_data.get("processed_at"),
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error fetching status: {e}")
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
    """
    Health check for the AI API service.
    
    Returns status of the API and worker connectivity.
    """
    try:
        # Check SQS connectivity
        sqs_service.check_connectivity()
        
        # Check PostgreSQL connectivity
        postgres_client.check_connectivity()

        return HealthResponse(
            status="healthy",
            worker_status="running",
        )

    except Exception as e:
        logger.error(f"Health check failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Service unhealthy",
        )
