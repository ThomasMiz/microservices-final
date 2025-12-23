FROM python:3.13-slim AS base

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Agregamos tanto el binario de pip como el directorio de la app al PATH y PYTHONPATH
ENV PATH="/home/lobbyms/.local/bin:$PATH"
ENV PYTHONPATH="/app"

RUN apt-get -y update \
    && apt-get install -y --no-install-recommends curl \
    && groupadd --system --gid 1001 lobbyms \
    && useradd --system --uid 1001 -g lobbyms --create-home --home-dir /home/lobbyms --shell /bin/bash lobbyms \
    && mkdir -p /app \
    && chown -R lobbyms:lobbyms /app \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
USER lobbyms

################################################################################################################

FROM base AS builder

COPY --chown=lobbyms:lobbyms requirements.txt ./requirements.txt

RUN pip install --user --no-cache-dir -r requirements.txt

# Importante: Copiamos el contenido de src a la raíz del WORKDIR (/app)
COPY --chown=lobbyms:lobbyms src/ ./

################################################################################################################

FROM builder AS production

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Uvicorn es la mejor opción para producción
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]