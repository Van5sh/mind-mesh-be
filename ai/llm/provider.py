import logging

from langchain_ollama import OllamaLLM

from ai.config.settings import settings

logger = logging.getLogger(__name__)

_SYSTEM_PROMPT = (
    "You are MeshMind's project assistant. Answer the user's question using "
    "only the context excerpts below, which come from documents uploaded to "
    "this project. If the context does not contain the answer, say you "
    "don't have enough information from the project's documents - do not "
    "make anything up."
)


class LLMProvider:
    """Generates grounded chat answers using the configured Ollama model."""

    def __init__(self):
        self.llm = OllamaLLM(
            model=settings.OLLAMA_MODEL,
            base_url=settings.OLLAMA_BASE_URL,
        )

    def generate_answer(self, question: str, context_chunks: list[str]) -> str:
        """
        Generate a grounded answer to a question given retrieved context.

        Args:
            question: The user's question.
            context_chunks: Text excerpts retrieved from Qdrant, most
                relevant first. May be empty if nothing relevant was found.

        Returns:
            The generated answer text.
        """
        if context_chunks:
            context = "\n\n---\n\n".join(context_chunks)
        else:
            context = "(No relevant excerpts were found in this project's documents.)"

        prompt = (
            f"{_SYSTEM_PROMPT}\n\n"
            f"Context:\n{context}\n\n"
            f"Question: {question}\n\n"
            "Answer:"
        )

        try:
            answer = self.llm.invoke(prompt)
            return answer.strip()
        except Exception as e:
            logger.error(f"Error generating LLM answer: {e}")
            raise
