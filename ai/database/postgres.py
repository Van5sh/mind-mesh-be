import logging

import psycopg2
from psycopg2.extras import RealDictCursor

from ai.config.settings import settings

logger = logging.getLogger(__name__)


class PostgresClient:
    """Client for updating file processing status in PostgreSQL.

    Status lives on file_ai_metadata (one row per file, keyed by file_id) -
    not on files itself. The Go backend creates that row (status PENDING)
    when the file is created, so every call here is an UPDATE against an
    already-existing row, never an INSERT.
    """

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
    ) -> None:
        """
        Update file processing status.

        Args:
            file_id: File ID
            project_id: Project ID (logging only - not a file_ai_metadata column)
            status: PENDING, PROCESSING, COMPLETED, or FAILED
            summary: Document summary (optional, set on COMPLETED)
            error_message: Error message (optional, set on FAILED)
        """
        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            updates = ["processing_status = %s", "updated_at = NOW()"]
            params = [status]

            if summary is not None:
                updates.append("summary = %s")
                params.append(summary)

            if error_message is not None:
                updates.append("error_message = %s")
                params.append(error_message)

            if status == "COMPLETED":
                updates.append("indexed_at = NOW()")
                updates.append("embedding_synced = TRUE")

            params.append(file_id)

            query = f"""
                UPDATE file_ai_metadata
                SET {", ".join(updates)}
                WHERE file_id = %s
            """

            cursor.execute(query, params)
            if cursor.rowcount == 0:
                logger.warning(
                    f"No file_ai_metadata row for file {file_id} "
                    f"(project {project_id}); status update had no effect"
                )
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
            project_id: Project ID (used only to confirm the file belongs
                to the caller's project; not a file_ai_metadata column)

        Returns:
            Dictionary with file status and metadata, or None if not found
        """
        try:
            conn = self._get_connection()
            cursor = conn.cursor(cursor_factory=RealDictCursor)

            query = """
                SELECT
                    fam.file_id,
                    f.project_id,
                    fam.processing_status,
                    fam.summary,
                    fam.error_message,
                    fam.indexed_at
                FROM file_ai_metadata fam
                JOIN files f ON f.id = fam.file_id
                WHERE fam.file_id = %s
                  AND f.project_id = %s
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
