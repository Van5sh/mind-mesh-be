# =========================
# Python AI Worker / API
#
# One image, two roles: docker-compose runs this as the SQS-consuming
# worker (default CMD) and separately as the FastAPI chat service
# (command override: uvicorn ai.main:app --host 0.0.0.0 --port 8000).
# =========================
FROM python:3.12-slim

# WORKDIR is the package root, not the ai/ package itself: every module
# in this codebase imports absolute paths like "ai.config.settings" and
# "ai.worker.processor", so ai/ must remain a subdirectory of the CWD
# (previously WORKDIR was /app/ai with ai/'s contents copied directly
# into it, which broke every one of those imports).
WORKDIR /app

# System dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install uv
RUN pip install --no-cache-dir uv

# Copy dependency files first (better layer caching)
COPY ai/pyproject.toml ai/uv.lock ./
RUN uv sync --frozen

# Copy AI application, preserving the ai/ package directory
COPY ai/ ./ai/

ENV PYTHONUNBUFFERED=1

EXPOSE 8000

# Default: run the SQS worker loop.
CMD ["uv", "run", "python", "-m", "ai.worker.main"]
