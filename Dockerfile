FROM python:3.11-slim

# Environment variables to avoid interactive prompts during package installs
ENV DEBIAN_FRONTEND=noninteractive
ENV PYTHONUNBUFFERED=1

WORKDIR /app

# 1. Install Node.js and Git required by Pretzel (if the project needs them)
RUN apt-get update && apt-get install -y \
    git \
    curl \
    build-essential \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# 2. Install PretzelAI from PyPI (preferred for stable releases).
# If the package is not available on PyPI, switch to installing from source.
RUN pip install --no-cache-dir pretzelai

# 3. Expose the default application port
EXPOSE 8888

# 4. Startup command
# Disable token for demo purposes so the UI is immediately accessible.
CMD ["pretzel", "lab", "--ip=0.0.0.0", "--port=8888", "--no-browser", "--allow-root", "--ServerApp.token=''", "--ServerApp.allow_remote_access=True"]
