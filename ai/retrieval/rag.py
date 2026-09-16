import logging

from ai.llm.provider import LLMProvider
from ai.retrieval.search import DocumentSearch

logger = logging.getLogger(__name__)


class RAGChatService:
    """Answers a question about a project's uploaded documents:
    embed the question -> retrieve relevant chunks -> ask the LLM,
    grounded in those chunks."""

    def __init__(self):
        self.search = DocumentSearch()
        self.llm = LLMProvider()

    def answer_question(self, question: str, project_id: str) -> dict:
        """
        Returns:
            {"answer": str, "sources": [{"file_id", "chunk_index"}]}
        """
        chunks = self.search.search(question, project_id)

        answer = self.llm.generate_answer(
            question=question,
            context_chunks=[c["text"] for c in chunks],
        )

        return {
            "answer": answer,
            "sources": [
                {"file_id": c["file_id"], "chunk_index": c["chunk_index"]}
                for c in chunks
            ],
        }
