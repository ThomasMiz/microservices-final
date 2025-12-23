import pytest
from fastapi.testclient import TestClient
from fastapi import HTTPException
from unittest.mock import AsyncMock, patch, Mock

from main import app
from check_in import CheckInService, process_checkin
from check_out import CheckOutService, process_check_out

client = TestClient(app)

# MAIN ENDPOINTS

def test_read_root():
    response = client.get("/")
    assert response.status_code == 200
    assert response.json() == {"Hello": "World", "service": "Lobby"}

def test_health_check():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}

# CHECK IN ENDPOINTS

@patch("main.process_checkin")
def test_check_in_successful(mock_process_checkin, mock_payment_status_clean):
    mock_process_checkin.return_value = {
        "status": "success",
        "checkin_id": "checkin_101_1703347200",
        "room_number": "101",
        "guest_id": "Marcelo",
        "timestamp": "2023-12-23T10:00:00+00:00",
        "payment_status": mock_payment_status_clean,
    }

    response = client.post("/check_in/101")

    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "success"
    assert data["room_number"] == "101"
    assert "checkin_id" in data
    assert "payment_status" in data
    assert data["payment_status"]["has_unpaid_tickets"] is False
    mock_process_checkin.assert_called_once_with("101", None, False)

@patch("main.process_checkin")
def test_check_in_with_guest_id_and_ignore_unpaid(mock_process_checkin, mock_payment_status_unpaid):
    mock_process_checkin.return_value = {
        "status": "success",
        "checkin_id": "checkin_101_1703347200",
        "room_number": "101",
        "guest_id": "Marcelo",
        "timestamp": "2023-12-23T10:00:00+00:00",
        "payment_status": mock_payment_status_unpaid,
    }

    response = client.post("/check_in/101", json={
        "guest_id": "Marcelo", 
        "ignore_unpaid_tickets": True
    })

    assert response.status_code == 200
    data = response.json()
    assert data["guest_id"] == "Marcelo"
    assert data["payment_status"]["has_unpaid_tickets"] is True
    mock_process_checkin.assert_called_once_with("101", "Marcelo", True)

@patch("main.process_checkin")
def test_check_in_blocked_by_unpaid_tickets(mock_process_checkin):
    mock_process_checkin.side_effect = HTTPException(
        status_code=402,
        detail={
            "message": "Check-in blocked: Guest has unpaid tickets",
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
        }
    )

    response = client.post("/check_in/101")

    assert response.status_code == 402
    detail = response.json()["detail"]
    assert "unpaid tickets" in detail["message"]
    assert detail["total_unpaid_amount"] == 123.45
    assert len(detail["unpaid_tickets"]) == 1

@patch("main.process_checkin")
def test_check_in_room_not_clean(mock_process_checkin):
    mock_process_checkin.side_effect = HTTPException(
        status_code=300, detail="Room 101 is not clean and ready for check-in"
    )

    response = client.post("/check_in/101")

    assert response.status_code == 300
    assert "not clean" in response.json()["detail"]

@patch("main.process_checkin")
def test_check_in_no_reservation(mock_process_checkin):
    mock_process_checkin.side_effect = HTTPException(
        status_code=404, detail="No active reservation found for room 101"
    )

    response = client.post("/check_in/101")

    assert response.status_code == 404
    assert "No active reservation" in response.json()["detail"]

@patch("main.process_checkin")
def test_check_in_internal_error(mock_process_checkin):
    mock_process_checkin.side_effect = Exception("Database connection error")

    response = client.post("/check_in/101")

    assert response.status_code == 500
    assert "Internal server error" in response.json()["detail"]

# CHECK IN SERVICE TESTS

@pytest.mark.asyncio
async def test_check_guest_payment_status_no_tickets(checkin_service):
    # Mock httpx directamente en la función
    with patch("httpx.AsyncClient") as mock_client_class:
        mock_client = AsyncMock()
        mock_response = Mock()  # ← Mock normal, no AsyncMock
        mock_response.status_code = 404
        mock_response.json.return_value = {}  # ← json() es síncrono
        mock_client.get = AsyncMock(return_value=mock_response)
        mock_client_class.return_value.__aenter__ = AsyncMock(return_value=mock_client)
        mock_client_class.return_value.__aexit__ = AsyncMock(return_value=None)

        result = await checkin_service.check_guest_payment_status("guest123")

        assert result["has_unpaid_tickets"] is False
        assert result["unpaid_tickets"] == []
        assert result["total_unpaid_amount"] == 0.0
        assert result["payment_check_skipped"] is False

@pytest.mark.asyncio
async def test_check_guest_payment_status_with_unpaid(checkin_service):
    with patch("httpx.AsyncClient") as mock_client_class:
        mock_client = AsyncMock()
        mock_response = Mock()  # ← Mock normal, no AsyncMock
        mock_response.status_code = 200
        mock_response.json.return_value = [  # ← json() es síncrono
            {
                "id": "ticket_123",
                "total": "50.00",
                "description": "Room service",
                "item_id": "item_001",
                "folder_id": "folder_001",
                "state": "Pending",
                "payment_id": None,
            },
            {
                "id": "ticket_456",
                "total": "75.50",
                "description": "Minibar",
                "item_id": "item_002",
                "folder_id": "folder_001",
                "state": "Paid",
                "payment_id": "payment_123",
            }
        ]
        mock_client.get = AsyncMock(return_value=mock_response)
        mock_client_class.return_value.__aenter__ = AsyncMock(return_value=mock_client)
        mock_client_class.return_value.__aexit__ = AsyncMock(return_value=None)

        result = await checkin_service.check_guest_payment_status("guest123")

        assert result["has_unpaid_tickets"] is True
        assert len(result["unpaid_tickets"]) == 1
        assert result["total_unpaid_amount"] == 50.0
        assert result["unpaid_tickets"][0]["ticket_id"] == "ticket_123"

@pytest.mark.asyncio
async def test_check_guest_payment_status_billing_service_unavailable(checkin_service):
    with patch("httpx.AsyncClient") as mock_client_class:
        # Mock de error de conexión
        from httpx import RequestError
        mock_client = AsyncMock()
        mock_client.get = AsyncMock(side_effect=RequestError("Connection failed"))
        mock_client_class.return_value.__aenter__ = AsyncMock(return_value=mock_client)
        mock_client_class.return_value.__aexit__ = AsyncMock(return_value=None)

        result = await checkin_service.check_guest_payment_status("guest123")

        assert result["has_unpaid_tickets"] is False
        assert result["payment_check_skipped"] is True
        assert "error" in result
        assert "unavailable" in result["error"]

@pytest.mark.asyncio
async def test_find_reservation_success(checkin_service, mock_reservation_data):
    with patch("httpx.AsyncClient") as mock_client_class:
        mock_client = AsyncMock()
        mock_response = Mock()  # ← Mock normal, no AsyncMock
        mock_response.status_code = 200
        # Usar la estructura correcta de la reserva
        reservation_with_room = dict(mock_reservation_data["reservation"])
        reservation_with_room["room"] = mock_reservation_data["room"]
        mock_response.json.return_value = [reservation_with_room]  # ← json() es síncrono
        mock_client.get = AsyncMock(return_value=mock_response)
        mock_client_class.return_value.__aenter__ = AsyncMock(return_value=mock_client)
        mock_client_class.return_value.__aexit__ = AsyncMock(return_value=None)

        result = await checkin_service.find_reservation("101", "Marcelo")

        assert result["reservation"]["id"] == 5
        assert result["room"]["number"] == "101"

@pytest.mark.asyncio
async def test_save_checkin_record(checkin_service, mock_reservation_data):
    checkin_id = await checkin_service.save_checkin_record(mock_reservation_data)

    assert checkin_id.startswith("checkin_101_")
    checkin_service.checkins.insert_one.assert_called_once()
    
@pytest.mark.asyncio
@patch("check_in.CheckInService")
@patch("messaging.MessageProducer")
async def test_process_checkin_success_with_payment_check(mock_producer_class, mock_service_class, mock_payment_status_clean, mock_reservation_data):
    mock_service = AsyncMock()

    # Mock find_reservation
    mock_service.find_reservation.return_value = mock_reservation_data

    # Mock check_guest_payment_status
    mock_service.check_guest_payment_status.return_value = mock_payment_status_clean

    # Mock save_checkin_record
    mock_service.save_checkin_record.return_value = "checkin_101_1703347200"

    # Mock MessageProducer instance and methods
    mock_producer_instance = AsyncMock()
    mock_producer_instance.send_message = AsyncMock()
    mock_producer_instance.check_in_topic = "check-in"
    mock_producer_class.get_instance.return_value = mock_producer_instance

    mock_service_class.return_value = mock_service

    result = await process_checkin("101", "Marcelo", False)

    assert result["status"] == "success"
    assert result["room_number"] == "101"
    assert result["guest_id"] == "Marcelo"
    assert result["checkin_id"] == "checkin_101_1703347200"
    assert result["payment_status"] == mock_payment_status_clean
    
    # Verificar que se llamó la verificación de pagos
    mock_service.check_guest_payment_status.assert_called_once_with("019b32ec-42d9-7261-823d-bba0c42995e6")

# CHECK OUT ENDPOINTS

@patch("main.process_check_out")
def test_check_out_successful(mock_process_check_out, mock_payment_status_clean):
    mock_process_check_out.return_value = {
        "status": "success",
        "message": "Check-out successful for room 101",
        "checkout_id": "checkout_101_12345",
        "room_id": "101",
        "guest_id": "guest123",
        "timestamp": "2023-12-23T12:00:00+00:00",
        "payment_status": mock_payment_status_clean,
    }

    response = client.post("/check_out/101")

    assert response.status_code == 200
    data = response.json()
    assert data["checkout_id"] == "checkout_101_12345"
    assert "payment_status" in data
    mock_process_check_out.assert_called_once_with("101", False)

@patch("main.process_check_out")
def test_check_out_blocked_by_unpaid_tickets(mock_process_check_out):
    mock_process_check_out.side_effect = HTTPException(
        status_code=402,
        detail={
            "message": "Check-out blocked: Guest has unpaid tickets",
            "unpaid_tickets": [
                {
                    "ticket_id": "019ab175-091e-70a3-b7e2-f8342e182474",
                    "total": 99.99,
                    "description": "Minibar consumption",
                    "item_id": "123456",
                    "folder_id": "019ab171-f38a-7900-8fca-7cb2ec35a96b"
                }
            ],
            "total_unpaid_amount": 99.99,
        }
    )

    response = client.post("/check_out/101")

    assert response.status_code == 402
    detail = response.json()["detail"]
    assert "unpaid tickets" in detail["message"]
    assert detail["total_unpaid_amount"] == 99.99

@patch("main.process_check_out")
def test_check_out_with_ignore_unpaid(mock_process_check_out, mock_payment_status_unpaid):
    mock_process_check_out.return_value = {
        "status": "success",
        "message": "Check-out successful for room 101",
        "checkout_id": "checkout_101_12345",
        "room_id": "101",
        "guest_id": "guest123",
        "timestamp": "2023-12-23T12:00:00+00:00",
        "payment_status": mock_payment_status_unpaid,
    }

    response = client.post("/check_out/101", json={"ignore_unpaid_tickets": True})

    assert response.status_code == 200
    data = response.json()
    assert data["payment_status"]["has_unpaid_tickets"] is True
    mock_process_check_out.assert_called_once_with("101", True)

@patch("main.process_check_out")
def test_check_out_not_occupied(mock_process_check_out):
    mock_process_check_out.side_effect = HTTPException(
        status_code=400, detail="Room 101 is not occupied or ready for check-out"
    )

    response = client.post("/check_out/101")

    assert response.status_code == 400
    assert "not occupied" in response.json()["detail"]

@patch("main.process_check_out")
def test_check_out_internal_error(mock_process_check_out):
    mock_process_check_out.side_effect = Exception("Database connection error")

    response = client.post("/check_out/101")

    assert response.status_code == 500
    assert "Internal error during check-out" in response.json()["detail"]

# CHECK OUT SERVICE TESTS

@pytest.mark.asyncio
async def test_get_guest_from_room_success(checkout_service):
    checkout_service.checkins.find_one.return_value = {
        "checkin_id": "checkin_101_12345",
        "room_id": "101",
        "guest_id": "guest123",
        "status": "confirmed"
    }

    guest_info = await checkout_service.get_guest_from_room("101")
    
    assert guest_info == {
        "guest_id": "guest123",
        "folder_id": "guest123"  # fallback to guest_id as per implementation
    }
    checkout_service.checkins.find_one.assert_called_once()

@pytest.mark.asyncio
async def test_get_guest_from_room_not_found(checkout_service):
    checkout_service.checkins.find_one.return_value = None

    with pytest.raises(HTTPException) as exc_info:
        await checkout_service.get_guest_from_room("101")
    
    assert exc_info.value.status_code == 404
    assert "No active check-in found" in str(exc_info.value.detail)

@pytest.mark.asyncio
async def test_check_guest_payment_status_checkout_with_unpaid(checkout_service):
    with patch("httpx.AsyncClient") as mock_client_class:
        mock_client = AsyncMock()
        mock_response = Mock()  # ← Mock normal, no AsyncMock
        mock_response.status_code = 200
        mock_response.json.return_value = [  # ← json() es síncrono
            {
                "id": "ticket_789",
                "total": "150.00",
                "description": "Late checkout fee",
                "item_id": "item_003",
                "folder_id": "folder_002",
                "state": "Pending",
                "payment_id": None,
            }
        ]
        mock_client.get = AsyncMock(return_value=mock_response)
        mock_client_class.return_value.__aenter__ = AsyncMock(return_value=mock_client)
        mock_client_class.return_value.__aexit__ = AsyncMock(return_value=None)

        result = await checkout_service.check_guest_payment_status("guest123")

        assert result["has_unpaid_tickets"] is True
        assert len(result["unpaid_tickets"]) == 1
        assert result["total_unpaid_amount"] == 150.0
        assert result["unpaid_tickets"][0]["ticket_id"] == "ticket_789"


@pytest.mark.asyncio
async def test_save_checkout_record(checkout_service):
    checkout_id = await checkout_service.save_checkout_record("101", "guest123")

    assert checkout_id.startswith("checkout_101_")
    checkout_service.checkouts.insert_one.assert_called_once()
    
@pytest.mark.asyncio
@patch("check_out.CheckOutService")
@patch("messaging.MessageProducer")
async def test_process_check_out_success_with_payment_check(mock_producer_class, mock_service_class, mock_payment_status_clean):
    mock_service = AsyncMock()

    # Mock is_room_occupied
    mock_service.is_room_occupied.return_value = True
    
    # Mock get_guest_from_room
    mock_service.get_guest_from_room.return_value = {
        "guest_id": "guest123",
        "folder_id": "folder123"
    }
    
    # Mock check_guest_payment_status
    mock_service.check_guest_payment_status.return_value = mock_payment_status_clean

    # Mock save_checkout_record
    mock_service.save_checkout_record.return_value = "checkout_101_12345"

    # Mock MessageProducer instance and methods
    mock_producer_instance = AsyncMock()
    mock_producer_instance.send_message = AsyncMock()
    mock_producer_instance.check_out_topic = "check-out"
    mock_producer_class.get_instance.return_value = mock_producer_instance

    mock_service_class.return_value = mock_service

    result = await process_check_out("101", False)

    assert result["status"] == "success"
    assert result["checkout_id"] == "checkout_101_12345"
    assert result["room_number"] == "101"
    assert result["guest_id"] == "guest123"
    assert result["payment_status"] == mock_payment_status_clean
    
    # Verificar que se llamaron todos los métodos nuevos
    mock_service.get_guest_from_room.assert_called_once_with("101")
    mock_service.check_guest_payment_status.assert_called_once_with("folder123")
