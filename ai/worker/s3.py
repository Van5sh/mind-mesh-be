class S3Downloader:
    def __init__(self, s3_client):
        self.s3_client = s3_client

    def download_file(self, bucket_name: str, object_key: str, download_path: str) -> None:
        """
        Download a file from S3 to a local path.

        :param bucket_name: Name of the S3 bucket.
        :param object_key: Key of the object in S3.
        :param download_path: Local path where the file will be downloaded.
        """
        self.s3_client.download_file(bucket_name, object_key, download_path)