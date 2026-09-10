import logging
import tempfile
from pathlib import Path

import boto3
from botocore.exceptions import ClientError

from ai.config.settings import settings

logger = logging.getLogger(__name__)


class S3Downloader:
    def __init__(self):
        self.s3 = boto3.client("s3", region_name=settings.AWS_REGION)
        self.bucket = settings.S3_BUCKET

    def download(self, storage_key: str) -> str:
        """
        Download file from S3 to temporary local path.

        Args:
            storage_key: S3 object key (e.g., "projects/proj-123/files/file-456.pdf")

        Returns:
            Path to temporary local file
        """
        try:
            # Create temp file with appropriate suffix
            suffix = Path(storage_key).suffix or ".tmp"
            temp_file = tempfile.NamedTemporaryFile(
                delete=False, suffix=suffix, prefix="meshMind_"
            )
            temp_path = temp_file.name
            temp_file.close()

            logger.info(f"Downloading s3://{self.bucket}/{storage_key} → {temp_path}")

            self.s3.download_file(self.bucket, storage_key, temp_path)

            logger.info(f"Successfully downloaded {storage_key}")
            return temp_path

        except ClientError as e:
            logger.error(f"Failed to download {storage_key} from S3: {e}")
            raise