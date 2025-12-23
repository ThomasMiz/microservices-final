import os
import logging
from typing import Optional
from fastapi import FastAPI, HTTPException, status, Request
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import time

# --- IMPORTACIONES DE OPENTELEMETRY (Tracing Only) ---
from opentelemetry import trace
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from opentelemetry.instrumentation.aiokafka import AIOKafkaInstrumentor

# --- IMPORTACIONES DE TU LÓGICA ---
from check_in import process_checkin
from check_out import process_check_out
from logging_config import setup_json_logging, get_logger

# 1. Configuración de Identidad y Endpoint
OTEL_ENDPOINT = os.getenv(
    "OTEL_EXPORTER_OTLP_ENDPOINT",
    "http://observability-opentelemetry-collector.shared-infrastructure.svc.cluster.local:4317",
)
resource = Resource.create(
    {
        "service.name": os.getenv("OTEL_SERVICE_NAME", "microservice-lobby"),
        "deployment.environment": os.getenv("ENVIRONMENT", "production"),
    }
)

# 2. Setup JSON Logging FIRST (before any other logging)
log_level = (
    logging.INFO if os.getenv("LOG_LEVEL", "INFO").upper() == "INFO" else logging.DEBUG
)
setup_json_logging(log_level)
logger = get_logger(__name__)

logger.info(
    "Starting microservice-lobby",
    extra={
        "otel_endpoint": OTEL_ENDPOINT,
        "environment": os.getenv("ENVIRONMENT", "production"),
    },
)

# 3. Configuración de TRACING (Rastreo)
tracer_provider = TracerProvider(resource=resource)
trace_exporter = OTLPSpanExporter(endpoint=OTEL_ENDPOINT, insecure=True)
tracer_provider.add_span_processor(BatchSpanProcessor(trace_exporter))
trace.set_tracer_provider(tracer_provider)

# 4. Inicialización de FastAPI
from contextlib import asynccontextmanager
from messaging import MessageProducer


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup: Conectar a Kafka
    logger.info("Application startup: Initializing Kafka producer")
    try:
        await MessageProducer.get_instance().start()
        logger.info("Kafka producer started successfully")
    except Exception as e:
        logger.error(
            "Error starting Kafka producer",
            extra={"error": str(e), "error_type": type(e).__name__},
            exc_info=True,
        )

    yield

    # Shutdown: Desconectar Kafka
    logger.info("Application shutdown: Stopping Kafka producer")
    await MessageProducer.get_instance().stop()
    logger.info("Application shutdown complete")


# Get base path from environment variable
base_path = os.getenv("BASE_PATH", "").rstrip("/")

app = FastAPI(
    title="Lobby Microservice",
    lifespan=lifespan,
    openapi_url=f"{base_path}/openapi.json" if base_path else "/openapi.json",
    docs_url=f"{base_path}/docs" if base_path else "/docs",
    redoc_url=f"{base_path}/redoc" if base_path else "/redoc",
)

# Configuración de CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 4. Instrumentación Automática
FastAPIInstrumentor.instrument_app(app, excluded_urls=".*health.*")
HTTPXClientInstrumentor().instrument()
AIOKafkaInstrumentor().instrument()


# Add middleware to log all requests with trace context
@app.middleware("http")
async def log_requests(request: Request, call_next):
    start_time = time.time()

    # Log request
    logger.info(
        "Incoming request",
        extra={
            "method": request.method,
            "path": request.url.path,
            "query_params": str(request.query_params),
            "client_host": request.client.host if request.client else None,
        },
    )

    # Process request
    response = await call_next(request)

    # Log response
    duration = time.time() - start_time
    logger.info(
        "Request completed",
        extra={
            "method": request.method,
            "path": request.url.path,
            "status_code": response.status_code,
            "duration_seconds": round(duration, 4),
        },
    )

    return response


# --- MODELOS ---
class CheckInRequest(BaseModel):
    guest_id: Optional[str] = None
    ignore_unpaid_tickets: bool = False

class CheckOutRequest(BaseModel):
    ignore_unpaid_tickets: bool = False


# --- ENDPOINTS ---
@app.get(f"{base_path}/" if base_path else "/")
def read_root():
    logger.info("Root endpoint accessed")
    return {"Hello": "World", "service": "Lobby"}


@app.get(f"{base_path}/health" if base_path else "/health")
def health_check():
    return {"status": "ok"}


@app.post(
    f"{base_path}/check_in/{{room_number}}" if base_path else "/check_in/{room_number}"
)
async def check_in(room_number: str, request: Optional[CheckInRequest] = None):
    guest_id = request.guest_id if request else None
    ignore_unpaid_tickets = request.ignore_unpaid_tickets if request else False

    logger.info(
        "Check-in request received",
        extra={
            "room_number": room_number,
            "guest_id": guest_id,
            "ignore_unpaid_tickets": ignore_unpaid_tickets,
            "operation": "check_in",
        },
    )

    try:
        # Esta llamada aparecerá dentro de la traza de la petición
        result = await process_checkin(room_number, guest_id, ignore_unpaid_tickets)

        logger.info(
            "Check-in completed successfully",
            extra={
                "room_number": room_number,
                "guest_id": guest_id,
                "checkin_id": result.get("checkin_id"),
                "operation": "check_in",
            },
        )
        return result

    except HTTPException as e:
        logger.warning(
            "Check-in failed with controlled error",
            extra={
                "room_number": room_number,
                "guest_id": guest_id,
                "status_code": e.status_code,
                "detail": e.detail,
                "operation": "check_in",
            },
        )
        raise
    except Exception as e:
        logger.error(
            "Check-in failed with critical error",
            extra={
                "room_number": room_number,
                "guest_id": guest_id,
                "error": str(e),
                "error_type": type(e).__name__,
                "operation": "check_in",
            },
            exc_info=True,
        )
        raise HTTPException(status_code=500, detail="Internal server error")


@app.post(f"{base_path}/check_out/{{room_number}}" if base_path else "/check_out/{room_number}")
async def check_out(room_number: str, request: Optional[CheckOutRequest] = None):
    ignore_unpaid_tickets = request.ignore_unpaid_tickets if request else False

    logger.info(
        "Check-out request received",
        extra={
            "room_number": room_number, 
            "ignore_unpaid_tickets": ignore_unpaid_tickets,
            "operation": "check_out"
        },
    )

    try:
        result = await process_check_out(room_number, ignore_unpaid_tickets)

        logger.info(
            "Check-out completed successfully",
            extra={
                "room_number": room_number,
                "checkout_id": result.get("checkout_id"),
                "operation": "check_out",
            },
        )
        return result
    except HTTPException as e:
        logger.warning(
            "Check-out failed with controlled error",
            extra={
                "room_number": room_number,
                "status_code": e.status_code,
                "detail": e.detail,
                "operation": "check_out",
            },
        )
        raise
    except Exception as e:
        logger.error(
            "Check-out failed with critical error",
            extra={
                "room_number": room_number,
                "error": str(e),
                "error_type": type(e).__name__,
                "operation": "check_out",
            },
            exc_info=True,
        )
        raise HTTPException(status_code=500, detail="Internal error during check-out")
