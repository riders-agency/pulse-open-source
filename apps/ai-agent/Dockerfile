FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    LANGGRAPH_STRICT_MSGPACK=true

WORKDIR /app

COPY pyproject.toml README.md ./
RUN pip install --no-cache-dir uv && uv pip install --system .

COPY src ./src

EXPOSE 9097
CMD ["python", "-m", "pulse_agent.main"]
