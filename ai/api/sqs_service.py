import json
import logging

import boto3
from botocore.exceptions import ClientError

from ai.config.settings import settings
from ai.api.models import ProcessFileRequest

logger = logging.getLogger(__name__)


class SQSService:
    """Service to send jobs to SQS queue."""

    def __init__(self):
        self.sqs = boto3.client("sqs", region_name=settings.AWS_REGION)
        self.queue_url = settings.SQS_QUEUE_URL

    def send_job(self, request: ProcessFileRequest) -> str:
        """
        Send file processing job to SQS.
        
        Args:
            request: ProcessFileRequest containing job metadata
        
        Returns:
            SQS MessageId
        
        Raises:
            ClientError: If SQS operation fails
        """
        try:
            message_body = {
                "job_id": request.job_id,
                "file_id": request.file_id,
                "project_id": request.project_id,
                "storage_key": request.storage_key,
                "content_type": request.content_type,
            }

            response = self.sqs.send_message(
                QueueUrl=self.queue_url,
                MessageBody=json.dumps(message_body),
                MessageAttributes={
                    "file_id": {"StringValue": request.file_id, "DataType": "String"},
                    "project_id": {
                        "StringValue": request.project_id,
                        "DataType": "String",
                    },
                },
            )

            logger.info(
                f"Message sent to SQS: MessageId={response['MessageId']}, "
                f"file_id={request.file_id}"
            )
            return response["MessageId"]

        except ClientError as e:
            logger.error(f"Error sending message to SQS: {e}")
            raise

    def check_connectivity(self) -> bool:
        """Check if SQS is reachable."""
        try:
            self.sqs.get_queue_attributes(
                QueueUrl=self.queue_url,
                AttributeNames=["ApproximateNumberOfMessages"],
            )
            return True
        except ClientError as e:
            logger.error(f"SQS connectivity check failed: {e}")
            raise
