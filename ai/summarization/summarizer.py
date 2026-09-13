import logging

from langchain_core.prompts import PromptTemplate
from langchain_ollama import OllamaLLM

from ai.config.settings import settings

logger = logging.getLogger(__name__)


class DocumentSummarizer:
    """Generate document summary using LLM."""

    def __init__(self):
        self.llm = OllamaLLM(
            model=settings.OLLAMA_MODEL,
            base_url=settings.OLLAMA_BASE_URL,
        )
        self.prompt = PromptTemplate(
            input_variables=["text"],
            template="""Provide a concise summary of the following document.
Keep it to 2-3 sentences maximum.

Document:
{text}

Summary:""",
        )
        self.chain = self.prompt | self.llm

    def summarize(self, text: str, max_tokens: int = 500) -> str:
        """Generate summary of text."""
        try:
            # Limit text to avoid token limits
            text = text[:5000]
            summary = self.chain.invoke({"text": text})
            logger.info(f"Generated summary: {len(summary)} chars")
            return summary.strip()
        except Exception as e:
            logger.error(f"Error generating summary: {e}")
            raise