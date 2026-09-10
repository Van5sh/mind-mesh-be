from pydantic import BaseModel, Field


class ProcessFileRequest(BaseModel):
    """Request to process a file."""
    
    job_id: str = Field(..., description="Unique job identifier")
    file_id: str = Field(..., description="File identifier")
    project_id: str = Field(..., description="Project identifier")
    storage_key: str = Field(..., description="S3 object key (e.g., projects/proj-123/files/file-456.pdf)")
    content_type: str = Field(..., description="MIME type (application/pdf, application/vnd.openxmlformats-officedocument.wordprocessingml.document, text/plain)")


class ProcessFileResponse(BaseModel):
    """Response after queuing file for processing."""
    
    job_id: str
    file_id: str
    project_id: str
    status: str = "QUEUED"
    message: str


class JobStatusResponse(BaseModel):
    """Response with job processing status."""
    
    job_id: str
    file_id: str
    project_id: str
    status: str  # QUEUED, PROCESSING, COMPLETED, FAILED
    summary: str = None
    error_message: str = None
    processed_at: str = None


class HealthResponse(BaseModel):
    """Health check response."""
    
    status: str = "healthy"
    worker_status: str = "running"
