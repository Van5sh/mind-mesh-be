from pydantic import BaseModel, Field


class ChatQuestionRequest(BaseModel):
    """A question asked in a project's AI-assistant chat."""

    project_id: str = Field(..., description="Project to search within")
    question: str = Field(..., description="The user's question")


class ChatSource(BaseModel):
    """One retrieved chunk that contributed to an answer."""

    file_id: str
    chunk_index: int


class ChatQuestionResponse(BaseModel):
    """A grounded answer plus the chunks it was drawn from."""

    answer: str
    sources: list[ChatSource]


class HealthResponse(BaseModel):
    """Health check response."""

    status: str = "healthy"
