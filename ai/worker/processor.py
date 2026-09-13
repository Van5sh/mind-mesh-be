import logging
from pathlib import Path
from typing import TYPE_CHECKING

from ai.worker.s3 import S3Downloader
from ai.extraction.factory import ExtractionFactory
from ai.chunking.text_splitter import TextChunker
from ai.summarization.summarizer import DocumentSummarizer
from ai.embeddings.text import TextEmbedder
from ai.qdrant.client import QdrantClient
from ai.database.postgres import PostgresClient

if TYPE_CHECKING:
    from ai.worker.main import ProcessingJob

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class DocumentProcessor:
    def __init__(self):
        self.s3_downloader = S3Downloader()
        self.extraction_factory = ExtractionFactory()
        self.chunker = TextChunker()
        self.summarizer = DocumentSummarizer()
        self.embedder = TextEmbedder()
        self.qdrant = QdrantClient()
        self.postgres = PostgresClient()

    def process_document(self, job: "ProcessingJob") -> None:
        """
        Complete pipeline:
        SQS job -> S3 download -> extract -> chunk -> summarize -> embed -> Qdrant -> PostgreSQL
        """
        temp_file = None
        try:
            # 1. Update status: PROCESSING
            self.postgres.update_file_status(
                job.file_id, job.project_id, "PROCESSING"
            )

            # 2. Download file from S3
            logger.info(f"Downloading {job.storage_key} from S3...")
            temp_file = self.s3_downloader.download(job.storage_key)

            # 3. Extract text based on content type
            logger.info(f"Extracting text from {job.content_type}...")
            extractor = self.extraction_factory.get_extractor(job.content_type)
            raw_text = extractor.extract(temp_file)

            if not raw_text.strip():
                raise ValueError(f"No text could be extracted from file: {job.file_id}")

            # 4. Chunk text
            logger.info("Chunking text...")
            chunks = self.chunker.split(raw_text)

            # 5. Generate summary
            logger.info("Generating summary...")
            summary = self.summarizer.summarize(raw_text)

            # 6. Generate embeddings for each chunk
            logger.info(f"Generating embeddings for {len(chunks)} chunks...")
            chunk_embeddings = []
            for i, chunk in enumerate(chunks):
                embedding = self.embedder.embed(chunk)
                chunk_embeddings.append(
                    {
                        "chunk_index": i,
                        "text": chunk,
                        "embedding": embedding,
                    }
                )

            # 7. Store in Qdrant
            logger.info("Storing vectors in Qdrant...")
            self.qdrant.store_document(
                file_id=job.file_id,
                project_id=job.project_id,
                summary=summary,
                chunks=chunk_embeddings,
            )

            # 8. Update PostgreSQL: COMPLETED
            logger.info("Updating PostgreSQL status...")
            self.postgres.update_file_status(
                job.file_id,
                job.project_id,
                "COMPLETED",
                summary=summary,
            )

            logger.info(
                f"Job {job.job_id} completed: file {job.file_id} processed successfully"
            )

        except Exception as e:
            logger.error(f"Processing failed for job {job.job_id}: {e}", exc_info=True)
            # Update status to FAILED
            self.postgres.update_file_status(
                job.file_id,
                job.project_id,
                "FAILED",
                error_message=str(e),
            )
            raise

        finally:
            # Cleanup temp file
            if temp_file and Path(temp_file).exists():
                Path(temp_file).unlink()
                logger.info(f"Cleaned up temp file: {temp_file}")
