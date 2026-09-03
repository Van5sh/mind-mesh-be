from config.settings import get_settings


def main() -> None:
    settings = get_settings()

    print(f"Starting MeshMind AI worker")
    print(f"Environment: {settings.app_env}")
    print(f"SQS queue: {settings.sqs_queue_url}")
    print(f"S3 bucket: {settings.s3_bucket}")


if __name__ == "__main__":
    main()