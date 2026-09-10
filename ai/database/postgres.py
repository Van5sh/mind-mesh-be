import logging
from datetime import datetime

import psycopg2
from psycopg2.extras import RealDictCursor

from ai.config.settings import settings

logger = logging.getLogger(__name__)


class PostgresClient:
    """Client for updating file processing status in PostgreSQL."""

    def __init__(self):
        self.conn_string = (
            f"postgresql://{settings.DB_USER}:{settings.DB_PASSWORD}"
            f"@{settings.DB_HOST}:{settings.DB_PORT}/{settings.DB_NAME}"
        )

    def _get_connection(self):
        """Get database connection."""
        return psycopg2.connect(self.conn_string)

    def update_file_status(
        self,
        file_id: str,
        project_id: str,
        status: str,
        summary: str = None,
        error_message: str = None,
        processed_at: bool = False,
    ) -> None:
        """
        Update file processing status.

        Args:
            file_id: File ID
            project_id: Project ID
            status: PROCESSING, COMPLETED, or FAILED
            summary: Document summary (optional)
            error_message: Error message if failed (optional)
            processed_at: Set current timestamp if True
        """
        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            # Build update query
            updates = ["processing_status = %s"]
            params = [status]

            if summary:
                updates.append("summary = %s")
                params.append(summary)

            if error_message:
                updates.append("error_message = %s")
                params.append(error_message)

            if processed_at:
                updates.append("processed_at = %s")
                params.append(datetime.utcnow())

            params.extend([file_id, project_id])

            query = f"""
                UPDATE files
                SET {", ".join(updates)}
                WHERE id = %s AND project_id = %s
            """

            cursor.execute(query, params)
            conn.commit()

            logger.info(f"Updated file {file_id} status to {status}")

            cursor.close()
            conn.close()

        except Exception as e:
            logger.error(f"Error updating file status: {e}")
            raise

    def get_file_status(self, file_id: str, project_id: str) -> dict:
        """
        Get file processing status.

        Args:
            file_id: File ID
            project_id: Project ID

        Returns:
            Dictionary with file status and metadata
        """
        try:
            conn = self._get_connection()
            cursor = conn.cursor(cursor_factory=RealDictCursor)

            query = """
                SELECT id, file_id, project_id, processing_status, 
                       summary, error_message, processed_at
                FROM files
                WHERE id = %s AND project_id = %s
            """

            cursor.execute(query, (file_id, project_id))
            result = cursor.fetchone()

            cursor.close()
            conn.close()

            return dict(result) if result else None

        except Exception as e:
            logger.error(f"Error fetching file status: {e}")
            raise

    def check_connectivity(self) -> bool:
        """Check if PostgreSQL is reachable."""
        try:
            conn = self._get_connection()
            cursor = conn.cursor()
            cursor.execute("SELECT 1")
            cursor.close()
            conn.close()
            return True
        except Exception as e:
            logger.error(f"PostgreSQL connectivity check failed: {e}")
            raise
