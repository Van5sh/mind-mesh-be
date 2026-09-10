import json
import logging
from typing import Optional

import boto3
from botocore.exceptions import ClientError
from pydantic import BaseModel, ValidationError

from ai.config.settings import settings
from ai.worker.processor import DocumentProcessor

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class ProcessingJob(BaseModel):
    job_id: str
    file_id: str
    project_id: str
    storage_key: str
    content_type: str


class AIWorker:
    def __init__(self):
        self.sqs = boto3.client("sqs", region_name=settings.AWS_REGION)
        self.queue_url = settings.SQS_QUEUE_URL
        self.processor = DocumentProcessor()

    def receive_messages(self) -> list:
        """Poll SQS with long polling."""
        try:
            response = self.sqs.receive_message(
                QueueUrl=self.queue_url,
                MaxNumberOfMessages=1,
                WaitTimeSeconds=20,  # Long polling
                MessageAttributeNames=["All"],
            )
            return response.get("Messages", [])
        except ClientError as e:
            logger.error(f"Error receiving messages from SQS: {e}")
            return []

    def parse_message(self, message: dict) -> Optional[ProcessingJob]:
        """Validate and parse SQS message into ProcessingJob."""
        try:
            body = json.loads(message["Body"])
            job = ProcessingJob(**body)
            return job
        except (json.JSONDecodeError, ValidationError) as e:
            logger.error(f"Invalid message format: {e}")
            return None

    def delete_message(self, message: dict) -> bool:
        """Delete message from SQS after successful processing."""
        try:
            self.sqs.delete_message(
                QueueUrl=self.queue_url,
                ReceiptHandle=message["ReceiptHandle"],
            )
            logger.info(f"Message deleted: {message['MessageId']}")
            return True
        except ClientError as e:
            logger.error(f"Error deleting message: {e}")
            return False

    def process_message(self, message: dict) -> bool:
        """Process a single SQS message."""
        job = self.parse_message(message)
        if not job:
            # Invalid message, delete it to avoid reprocessing
            self.delete_message(message)
            return False

        logger.info(f"Processing job {job.job_id} for file {job.file_id}")

        try:
            self.processor.process_document(job)
            self.delete_message(message)
            logger.info(f"Job {job.job_id} completed successfully")
            return True
        except Exception as e:
            logger.error(f"Job {job.job_id} failed: {e}", exc_info=True)
            # Message will be retried by SQS after visibility timeout
            return False

    def run(self):
        """Main worker loop."""
        logger.info("Starting AI Worker...")
        try:
            while True:
                messages = self.receive_messages()
                if messages:
                    for message in messages:
                        self.process_message(message)
                else:
                    logger.debug("No messages received, continuing poll...")
        except KeyboardInterrupt:
            logger.info("Worker interrupted, shutting down...")
        except Exception as e:
            logger.error(f"Fatal error in worker loop: {e}", exc_info=True)
            raise


def main():
    worker = AIWorker()
    worker.run()


if __name__ == "__main__":
    main()