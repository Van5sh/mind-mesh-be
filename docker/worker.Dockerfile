# =========================
# Python AI Worker
# =========================
FROM python:3.12-slim

WORKDIR /app/ai

# System dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install uv
RUN pip install --no-cache-dir uv

# Copy dependency files first
COPY ai/pyproject.toml ai/uv.lock ./

# Install locked dependencies
RUN uv sync --frozen

# Copy AI application
COPY ai/ .

# Start worker
CMD ["uv", "run", "python", "-m", "worker.main"]