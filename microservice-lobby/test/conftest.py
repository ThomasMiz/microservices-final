import pytest
import asyncio
import sys
import os
import logging
from unittest.mock import AsyncMock, patch

# Agregar src al path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))


@pytest.fixture(scope="session")
def event_loop():
    """Fixture para el loop de eventos asyncio"""
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()


@pytest.fixture(autouse=True)
def setup_env_vars(monkeypatch):
    """Configurar variables de entorno para tests"""
    monkeypatch.setenv(
        "MONGODB_CONNECTION_STRING", "mongodb://test:test@localhost:27017/test_db"
    )
    monkeypatch.setenv("CLEANING_SERVICE_URL", "http://test-cleaning-service:8080")
    monkeypatch.setenv("CLEANING_SERVICE_USER", "test_user")
    monkeypatch.setenv("CLEANING_SERVICE_PASSWORD", "test_pass")
    monkeypatch.setenv(
        "RESERVATIONS_SERVICE_URL", "http://test-reservations-service:8080"
    )
    monkeypatch.setenv("BILLING_SERVICE_URL", "http://test-billing-service:8080")
    
    # Deshabilitar OpenTelemetry para tests
    monkeypatch.setenv("OTEL_SDK_DISABLED", "true")
    monkeypatch.setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")


@pytest.fixture(autouse=True)
def disable_logging():
    """Deshabilitar logging verboso durante tests"""
    # Deshabilitar OpenTelemetry logging
    logging.getLogger("opentelemetry").setLevel(logging.CRITICAL)
    logging.getLogger("grpc").setLevel(logging.CRITICAL)
    logging.getLogger("httpx").setLevel(logging.WARNING)
    yield


@pytest.fixture
def mock_mongo_client():
    """Mock para MongoDB client"""
    mock_client = AsyncMock()
    mock_db = AsyncMock()
    mock_checkins = AsyncMock()
    mock_checkouts = AsyncMock()

    mock_client.lobby_db = mock_db
    mock_client.chotel = mock_db  # Nombre alternativo de DB
    mock_db.checkins = mock_checkins
    mock_db.checkouts = mock_checkouts

    return mock_client


@pytest.fixture
def mock_cleaning_service():
    """Mock para el servicio de limpieza"""
    return AsyncMock()


@pytest.fixture
def checkin_service():
    """Mock del servicio de check-in configurado"""
    from check_in import CheckInService
    
    # Crear instancia sin llamar __init__
    service = CheckInService.__new__(CheckInService)
    
    # Configurar propiedades manualmente
    service.cleaning_service_url = "http://test-cleaning-service:8080"
    service.cleaning_service_user = "test_user"
    service.cleaning_service_password = "test_pass"
    service.reservations_service_url = "http://test-reservations-service:8080"
    service.billing_service_url = "http://test-billing-service:8080"
    
    # Mock de colecciones MongoDB
    service.checkins = AsyncMock()
    service.db = AsyncMock()
    service.client = AsyncMock()
    
    return service


@pytest.fixture
def checkout_service():
    """Mock del servicio de check-out configurado"""
    from check_out import CheckOutService
    
    # Crear instancia sin llamar __init__
    service = CheckOutService.__new__(CheckOutService)
    
    # Configurar propiedades manualmente
    service.cleaning_service_url = "http://test-cleaning-service:8080"
    service.cleaning_service_user = "test_user"
    service.cleaning_service_password = "test_pass"
    service.billing_service_url = "http://test-billing-service:8080"
    
    # Mock de colecciones MongoDB
    service.checkins = AsyncMock()
    service.checkouts = AsyncMock()
    service.db = AsyncMock()
    service.client = AsyncMock()
    
    return service


@pytest.fixture
def mock_payment_status_clean():
    """Mock de estado de pagos sin deudas"""
    return {
        "has_unpaid_tickets": False,
        "unpaid_tickets": [],
        "total_unpaid_amount": 0.0,
        "payment_check_skipped": False,
    }


@pytest.fixture
def mock_payment_status_unpaid():
    """Mock de estado de pagos con deudas pendientes"""
    return {
        "has_unpaid_tickets": True,
        "unpaid_tickets": [
            {
                "ticket_id": "019ab175-091e-70a3-b7e2-f8342e182474",
                "total": 123.45,
                "description": "Room service",
                "item_id": "123456",
                "folder_id": "019ab171-f38a-7900-8fca-7cb2ec35a96b"
            }
        ],
        "total_unpaid_amount": 123.45,
        "payment_check_skipped": False,
    }


@pytest.fixture
def mock_payment_status_service_error():
    """Mock de estado de pagos cuando el servicio falla"""
    return {
        "has_unpaid_tickets": False,
        "unpaid_tickets": [],
        "total_unpaid_amount": 0.0,
        "payment_check_skipped": True,
        "error": "Billing service unavailable: Connection failed",
    }


@pytest.fixture
def mock_reservation_data():
    """Mock de datos de reserva típicos"""
    return {
        "reservation": {
            "billingFolderId": "019b32ec-42d9-7261-823d-bba0c42995e6",
            "endDate": "2025-12-22T15:00:00Z",
            "guestId": "Marcelo",
            "id": 5,
            "rentedHourlyPrice": 599.99,
            "reservationBillingTicket": "019b32ec-4339-73e0-b15c-7c53496c22f3",
            "startDate": "2025-12-21T15:00:00Z",
            "totalPrice": 14399.76,
        },
        "room": {
            "active": True,
            "category": "SUITE",
            "description": "The best room in floor 1, with a gorgeous view of the garbage alley down below xd",
            "hourlyPrice": 599.99,
            "id": 2,
            "maxCapacity": 3,
            "name": "Room 101",
            "number": "101",
        },
    }


# Configuraciones globales para pytest
def pytest_configure(config):
    """Configuración global de pytest"""
    # Agregar marcadores personalizados
    config.addinivalue_line("markers", "unit: marca tests unitarios")
    config.addinivalue_line("markers", "integration: marca tests de integración")
    config.addinivalue_line("markers", "slow: marca tests lentos")


# Hook para ejecutar antes de cada test
def pytest_runtest_setup(item):
    """Setup que se ejecuta antes de cada test"""
    # Aquí puedes agregar setup específico si necesitas
    pass


# Hook para limpiar después de cada test
def pytest_runtest_teardown(item, nextitem):
    """Teardown que se ejecuta después de cada test"""
    # Limpiar cualquier estado global si es necesario
    pass