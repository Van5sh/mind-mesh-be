import logging

from langchain_text_splitters import RecursiveCharacterTextSplitter

logger = logging.getLogger(__name__)


class TextChunker:
    """Split text into chunks using RecursiveCharacterTextSplitter."""

    def __init__(self, chunk_size: int = 1000, chunk_overlap: int = 200):
        self.splitter = RecursiveCharacterTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap,
            separators=["\n\n", "\n", " ", ""],
        )

    def split(self, text: str) -> list[str]:
        """Split text into chunks."""
        try:
            chunks = self.splitter.split_text(text)
            logger.info(f"Split text into {len(chunks)} chunks")
            return chunks
        except Exception as e:
            logger.error(f"Error splitting text: {e}")
            raise
