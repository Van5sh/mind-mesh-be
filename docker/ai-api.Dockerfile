# =========================
# Python AI API (chat Q&A over HTTP)
#
# Same image as docker/worker.Dockerfile, except the command is baked in
# here instead of relied on as a compose/CLI override - Render's "New Web
# Service" (Docker) creation flow has no start-command override field, so
# the Dockerfile itself has to be the thing that decides which of the two
# roles (worker vs API) this container plays. Use this Dockerfile path for
# the ai-api service; use docker/worker.Dockerfile for the worker.
# =========================
FROM python:3.12-slim

# WORKDIR is the package root, not the ai/ package itself: every module
# in this codebase imports absolute paths like "ai.config.settings" and
# "ai.worker.processor", so ai/ must remain a subdirectory of the CWD.
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

# Shell form (not exec-form JSON array) so $PORT is substituted at
# container start - Render assigns this dynamically and health-checks
# whatever's actually bound, so a hardcoded port here would break that.
CMD uv run uvicorn ai.main:app --host 0.0.0.0 --port ${PORT:-8000}
