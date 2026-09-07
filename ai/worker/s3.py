from pathlib import Path
import boto3
from config.settings import get_settings

class S3Downloader:
    def __init__(self) -> None:
        settings = get_settings()
        self.client = boto3.client(
            "s3",
            region_name=settings.aws_region,
        )
        self.bucket = settings.s3_bucket

    def download(self, storage_key: str, destination: str) -> str:
        destination_path = Path(destination)
        destination_path.parent.mkdir(
            parents=True,
            exist_ok=True,
        )
        self.client.download_file(
            self.bucket,
            storage_key,
            str(destination_path),
        )
        return str(destination_path)