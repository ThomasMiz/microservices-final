import os
from datetime import datetime
from typing import Optional, Dict, Any 
import httpx
from motor.motor_asyncio import AsyncIOMotorClient
from fastapi import HTTPException

# --- CONFIGURACIÓN DE TELEMETRÍA ---
from opentelemetry import trace

# --- CONFIGURACIÓN DE LOGGING ---
from logging_config import get_logger

# Obtenemos el tracer y el logger para este módulo
tracer = trace.get_tracer(__name__)
logger = get_logger(__name__)


class CheckInService:
    def __init__(self):
        self.cleaning_service_url = os.getenv("CLEANING_SERVICE_URL")
        self.cleaning_service_user = os.getenv("CLEANING_SERVICE_USER")
        self.cleaning_service_password = os.getenv("CLEANING_SERVICE_PASSWORD")
        self.reservations_service_url = os.getenv("RESERVATIONS_SERVICE_URL")
        self.billing_service_url = os.getenv("BILLING_SERVICE_URL")

        # MongoDB connection desde secret del pipeline
        connection_string = os.getenv("MONGODB_CONNECTION_STRING")
        if not connection_string:
            raise ValueError(
                "MONGODB_CONNECTION_STRING environment variable is required"
            )

        # Extraer nombre de BD para compatibilidad
        db_name = (
            connection_string.split("/")[-1] if "/" in connection_string else "chotel"
        )

        self.client = AsyncIOMotorClient(connection_string)
        self.db = self.client[db_name]
        self.checkins = self.db.checkins

    async def check_guest_payment_status(self, folder_id: str) -> Dict[str, Any]:
        """
        Verifica si el guest tiene tickets pendientes de pago en el servicio de billing.
        
        Returns:
            Dict con información del estado de pagos del guest:
            - has_unpaid_tickets: bool - Si tiene tickets sin pagar
            - unpaid_tickets: List - Lista de tickets sin pagar
            - total_unpaid_amount: float - Monto total pendiente
        """
        with tracer.start_as_current_span("service.check_guest_payment_status"):
            logger.info(
                "Checking guest payment status",
                extra={
                    "folder_id": folder_id,
                    "service": "billing",
                    "operation": "check_payment_status",
                },
            )

            if not self.billing_service_url:
                logger.warning(
                    "Billing service URL not configured, skipping payment check",
                    extra={
                        "folder_id": folder_id, 
                        "service": "billing",
                        "configuration_issue": "missing_billing_url"
                    },
                )
                return {
                    "has_unpaid_tickets": False,
                    "unpaid_tickets": [],
                    "total_unpaid_amount": 0.0,
                    "payment_check_skipped": True,
                }

            try:
                async with httpx.AsyncClient() as client:
                    # Buscar todos los tickets del guest usando la API de billing
                    params = {
                        "folderId": folder_id,
                    }

                    logger.debug(
                        "Making request to billing service",
                        extra={
                            "folder_id": folder_id,
                            "service": "billing",
                            "endpoint": "/tickets",
                            "params": params,
                            "url": f"{self.billing_service_url}/tickets"
                        },
                    )

                    response = await client.get(
                        f"{self.billing_service_url}/tickets",
                        params=params,
                        timeout=10.0,
                    )

                    if response.status_code == 200:
                        tickets_data = response.json()
                        
                        # Extraer tickets de la respuesta (puede ser un array directo o tener estructura paginada)
                        tickets = tickets_data if isinstance(tickets_data, list) else tickets_data.get('tickets', [])
                        
                        logger.debug(
                            "Retrieved tickets from billing service",
                            extra={
                                "folder_id": folder_id,
                                "ticket_count": len(tickets),
                                "tickets_data": tickets_data,
                                "service": "billing",
                                "response_structure": "list" if isinstance(tickets_data, list) else "paginated"
                            },
                        )

                        unpaid_tickets = []
                        total_unpaid_amount = 0.0

                        # Verificar cada ticket para ver si está pagado
                        for ticket in tickets:
                            is_paid = ticket.get("state", "") == "Paid"
                            
                            if not is_paid:
                                ticket_amount = float(ticket.get("total", 0))
                                unpaid_tickets.append({
                                    "ticket_id": ticket.get("id"),
                                    "total": ticket_amount,
                                    "description": ticket.get("description", ""),
                                    "item_id": ticket.get("item_id", ""),
                                    "folder_id": ticket.get("folder_id", ""),
                                })
                                total_unpaid_amount += ticket_amount

                        has_unpaid = len(unpaid_tickets) > 0

                        logger.info(
                            "Payment status check completed",
                            extra={
                                "folder_id": folder_id,
                                "has_unpaid_tickets": has_unpaid,
                                "unpaid_count": len(unpaid_tickets),
                                "total_unpaid_amount": total_unpaid_amount,
                                "total_tickets_checked": len(tickets),
                                "service": "billing",
                                "operation_result": "success"
                            },
                        )

                        return {
                            "has_unpaid_tickets": has_unpaid,
                            "unpaid_tickets": unpaid_tickets,
                            "total_unpaid_amount": total_unpaid_amount,
                            "payment_check_skipped": False,
                        }

                    elif response.status_code == 404:
                        # No se encontraron tickets para este guest
                        logger.info(
                            "No tickets found for guest",
                            extra={
                                "folder_id": folder_id,
                                "service": "billing",
                                "http_status": response.status_code,
                                "operation_result": "no_tickets_found"
                            },
                        )
                        return {
                            "has_unpaid_tickets": False,
                            "unpaid_tickets": [],
                            "total_unpaid_amount": 0.0,
                            "payment_check_skipped": False,
                        }
                    else:
                        logger.error(
                            "Billing service returned error",
                            extra={
                                "guest_id": folder_id,
                                "http_status": response.status_code,
                                "service": "billing",
                                "error_type": "http_error",
                                "operation_result": "service_error"
                            },
                        )
                        # En caso de error del servicio, permitimos el check-in pero registramos el problema
                        return {
                            "has_unpaid_tickets": False,
                            "unpaid_tickets": [],
                            "total_unpaid_amount": 0.0,
                            "payment_check_skipped": True,
                            "error": f"Billing service error: {response.status_code}",
                        }

            except httpx.RequestError as exc:
                logger.error(
                    "Failed to connect to billing service",
                    extra={
                        "folder_id": folder_id,
                        "error_message": str(exc),
                        "error_type": type(exc).__name__,
                        "service": "billing",
                        "operation_result": "connection_error",
                        "billing_url": self.billing_service_url
                    },
                    exc_info=True,
                )
                # En caso de error de conexión, permitimos el check-in pero registramos el problema
                return {
                    "has_unpaid_tickets": False,
                    "unpaid_tickets": [],
                    "total_unpaid_amount": 0.0,
                    "payment_check_skipped": True,
                    "error": f"Billing service unavailable: {str(exc)}",
                }

    async def find_reservation(
        self, room_number: str, guest_id: Optional[str] = None
    ) -> Dict[Any, Any]:
        """
        Busca la reserva activa. Cada paso importante está envuelto en un 'span'.
        """
        with tracer.start_as_current_span("service.find_reservation"):
            logger.info(
                "Querying reservation service",
                extra={
                    "room_number": room_number,
                    "guest_id": guest_id,
                    "service": "reservations",
                    "operation": "find_reservation",
                },
            )

            try:
                async with httpx.AsyncClient() as client:
                    current_time = datetime.now().astimezone().isoformat()
                    params = {
                        "from": current_time,
                        "to": current_time,
                        "page": 0,
                        "pageSize": 10,
                    }
                    if guest_id:
                        params["guestId"] = guest_id

                    response = await client.get(
                        f"{self.reservations_service_url}/reservations",
                        params=params,
                        timeout=10.0,
                    )

                    if response.status_code == 200:
                        reservations = response.json()

                        logger.debug(
                            "Received reservations from service",
                            extra={
                                "room_number": room_number,
                                "reservation_count": len(reservations),
                                "service": "reservations",
                            },
                        )

                        for reservation in reservations:
                            if reservation["room"].get("number") == room_number:
                                logger.info(
                                    "Active reservation found",
                                    extra={
                                        "room_number": room_number,
                                        "reservation_id": reservation.get("id"),
                                        "guest_id": reservation.get("guestId"),
                                        "service": "reservations",
                                    },
                                )
                                return {
                                    "reservation": reservation,
                                    "room": reservation["room"],
                                }

                        logger.warning(
                            "No active reservation found for room",
                            extra={
                                "room_number": room_number,
                                "guest_id": guest_id,
                                "service": "reservations",
                            },
                        )
                        raise HTTPException(
                            status_code=404, detail="No active reservation found"
                        )
                    else:
                        logger.error(
                            "Reservations service returned error",
                            extra={
                                "room_number": room_number,
                                "status_code": response.status_code,
                                "service": "reservations",
                            },
                        )
                        raise HTTPException(
                            status_code=400, detail="Could not retrieve reservations"
                        )

            except httpx.RequestError as exc:
                logger.error(
                    "Failed to connect to reservations service",
                    extra={
                        "room_number": room_number,
                        "error": str(exc),
                        "error_type": type(exc).__name__,
                        "service": "reservations",
                    },
                    exc_info=True,
                )
                raise HTTPException(
                    status_code=503, detail="Reservations service unavailable"
                )

    async def save_checkin_record(self, reservation_data: Dict[Any, Any]) -> str:
        with tracer.start_as_current_span("database.mongodb_insert"):
            room_number = reservation_data["room"]["number"]

            logger.info(
                "Saving check-in record to database",
                extra={
                    "room_number": room_number,
                    "room_id": reservation_data["room"]["id"],
                    "reservation_id": reservation_data["reservation"]["id"],
                    "guest_id": reservation_data["reservation"]["guestId"],
                    "operation": "save_checkin",
                },
            )

            checkin_id = f"checkin_{room_number}_{int(datetime.now().timestamp())}"

            checkin_record = {
                "checkin_id": checkin_id,
                "room_id": reservation_data["room"]["id"],
                "room_number": room_number,
                "reservation_id": reservation_data["reservation"]["id"],
                "guest_id": reservation_data["reservation"]["guestId"],
                "folder_id": reservation_data["reservation"]["billingFolderId"],
                "checkin_time": datetime.now().astimezone().isoformat(),
                "status": "confirmed",
            }
            
            await self.checkins.insert_one(checkin_record)

            logger.info(
                "Check-in record saved successfully",
                extra={
                    "checkin_id": checkin_id,
                    "room_number": room_number,
                    "operation": "save_checkin",
                },
            )

            return checkin_id


async def process_checkin(room_number: str, guest_id: Optional[str] = None, ignore_unpaid_tickets: bool = False) -> dict:
    """
    Procesa el check-in incluyendo verificación de pagos pendientes.
    
    Args:
        room_number: Número de habitación
        guest_id: ID del huésped (opcional)
        ignore_unpaid_tickets: Si es True, permite check-in aunque haya tickets sin pagar
    """
    # Este span engloba todo el proceso de negocio
    with tracer.start_as_current_span("process_checkin_flow"):
        logger.info(
            "Starting check-in process",
            extra={
                "room_number": room_number,
                "guest_id": guest_id,
                "ignore_unpaid_tickets": ignore_unpaid_tickets,
                "flow": "check_in",
            },
        )

        service = CheckInService()

        # 1. Buscar reserva
        reservation_data = await service.find_reservation(room_number, guest_id)
        
        # Obtener guest_id de la reserva si no se proporcionó
        guest_id = guest_id or reservation_data["reservation"]["guestId"]
        folder_id = reservation_data["reservation"]["billingFolderId"]

        # 2. Verificar estado de pagos del guest
        payment_status = await service.check_guest_payment_status(folder_id)
        
        # 3. Decidir si proceder con el check-in
        if payment_status["has_unpaid_tickets"] and not ignore_unpaid_tickets:
            logger.warning(
                "Check-in blocked due to unpaid tickets",
                extra={
                    "room_number": room_number,
                    "guest_id": guest_id,
                    "unpaid_count": len(payment_status["unpaid_tickets"]),
                    "total_unpaid": payment_status["total_unpaid_amount"],
                    "flow": "check_in",
                    "operation_result": "blocked"
                },
            )
            raise HTTPException(
                status_code=402,  # Payment Required
                detail={
                    "message": "Check-in blocked: Guest has unpaid tickets",
                    "unpaid_tickets": payment_status["unpaid_tickets"],
                    "total_unpaid_amount": payment_status["total_unpaid_amount"],
                }
            )

        # 4. Guardar registro (incluyendo estado de pagos)
        checkin_id = await service.save_checkin_record(reservation_data)

        # 5. Notificar a Kafka (Limpieza/Estado)
        from messaging import MessageProducer

        kafka_event = {
            "roomNumber": room_number,
            "guestId": guest_id,
            "timestamp": datetime.now().astimezone().isoformat()
        }

        try:
            logger.info(
                "Sending check-in event to Kafka",
                extra={
                    "checkin_id": checkin_id,
                    "room_number": room_number,
                    "guest_id": guest_id,
                    "topic": "check-in",
                    "operation": "kafka_publish"
                },
            )

            await MessageProducer.get_instance().send_message(
                topic=MessageProducer.get_instance().check_in_topic,
                value=kafka_event,
            )

            logger.info(
                "Check-in event sent to Kafka successfully",
                extra={
                    "checkin_id": checkin_id,
                    "room_number": room_number,
                    "guest_id": guest_id,
                    "topic": MessageProducer.get_instance().check_in_topic,
                    "operation_result": "success"
                },
            )
        except Exception as e:
            # No bloqueamos el check-in si falla el aviso, pero logueamos error grave
            logger.error(
                "Failed to send check-in event to Kafka",
                extra={
                    "checkin_id": checkin_id,
                    "room_number": room_number,
                    "guest_id": guest_id,
                    "error_message": str(e),
                    "error_type": type(e).__name__,
                    "topic": "check-in",
                    "operation_result": "kafka_error"
                },
                exc_info=True,
            )

        logger.info(
            "Check-in process completed",
            extra={
                "checkin_id": checkin_id,
                "room_number": room_number,
                "guest_id": guest_id,
                "payment_status_checked": True,
                "had_unpaid_tickets": payment_status["has_unpaid_tickets"],
                "total_unpaid_amount": payment_status.get("total_unpaid_amount", 0.0),
                "payment_check_skipped": payment_status.get("payment_check_skipped", False),
                "flow": "check_in",
                "operation_result": "success"
            },
        )

        return {
            "status": "success",
            "checkin_id": checkin_id,
            "room_number": room_number,
            "guest_id": guest_id,
            "timestamp": datetime.now().astimezone().isoformat(),
            "payment_status": payment_status,
        }
