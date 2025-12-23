import os
from datetime import datetime
from typing import Dict, Any 
import httpx
from motor.motor_asyncio import AsyncIOMotorClient
from fastapi import HTTPException
from messaging import MessageProducer

# --- CONFIGURACIÓN DE TELEMETRÍA ---
from opentelemetry import trace

# --- CONFIGURACIÓN DE LOGGING ---
from logging_config import get_logger

tracer = trace.get_tracer(__name__)
logger = get_logger(__name__)


class CheckOutService:
    def __init__(self):
        # Configuración de URLs y Seguridad (Sin defaults)
        self.cleaning_service_url = os.getenv("CLEANING_SERVICE_URL")
        self.cleaning_service_user = os.getenv("CLEANING_SERVICE_USER")
        self.cleaning_service_password = os.getenv("CLEANING_SERVICE_PASSWORD")
        self.billing_service_url = os.getenv("BILLING_SERVICE_URL")

        # MongoDB connection desde secret del pipeline
        connection_string = os.getenv("MONGODB_CONNECTION_STRING")

        # Inicializar variables para fallback
        db_user = db_pass = db_host = db_port = None

        if connection_string:
            mongodb_url = connection_string
            # Extraer nombre de BD para compatibilidad
            db_name = (
                connection_string.split("/")[-1]
                if "/" in connection_string
                else "chotel"
            )
        else:
            # Fallback a variables individuales para desarrollo local
            db_user = os.getenv("MONGODB_USER", "chotel_user")
            db_pass = os.getenv("MONGODB_PASSWORD", "")
            db_host = os.getenv(
                "MONGODB_HOST",
                "chotel-default-mongo.shared-infrastructure.svc.cluster.local",
            )
            db_port = os.getenv("MONGODB_PORT", "27017")
            db_name = os.getenv("MONGODB_DATABASE", "chotel")
            mongodb_url = f"mongodb://{db_user}:{db_pass}@{db_host}:{db_port}/{db_name}"

        # Validación Fail-fast: Si falta algo, el servicio no arranca
        if not self.cleaning_service_url:
            logger.critical("CONFIGURACIÓN FALTANTE: CLEANING_SERVICE_URL")
            raise ValueError("Falta variable crítica: CLEANING_SERVICE_URL")

        if not connection_string and not all([db_user, db_pass, db_host]):
            logger.critical("CONFIGURACIÓN FALTANTE: Variables de MongoDB")
            raise ValueError("Faltan variables críticas de MongoDB")
        
        self.client = AsyncIOMotorClient(mongodb_url)
        self.db = self.client[db_name]
        self.checkouts = self.db.checkouts  # Colección para registros de salida
        self.checkins = self.db.checkins    # Para obtener guest_id del check-in

    async def get_guest_from_room(self, room_number: str) -> Dict[str, Any]:
        """
        Obtiene el guest_id y folder_id de la habitación desde el último check-in confirmado.
        - room_number en Mongo está guardado como Number (int). Si llega como string, se normaliza.
        - folder_id se toma de payment_status.folder_id si existe; si no, fallback a guest_id.
        """
        with tracer.start_as_current_span("database.get_guest_from_room") as span:
            logger.info(
                "Getting guest and folder from room check-in record",
                extra={
                    "room_number": room_number,
                    "operation": "get_guest_from_room",
                },
            )

            # (Opcional) Anotar en el span
            try:
                span.set_attribute("room_number", room_number)
            except Exception:
                pass

            try:
                checkin_record = await self.checkins.find_one(
                    {"room_number": room_number, "status": "confirmed"},
                    sort=[("checkin_time", -1)],
                )

                if not checkin_record:
                    logger.warning(
                        "No active check-in found for room",
                        extra={
                            "room_number": room_number,
                            "operation_result": "no_checkin_found",
                        },
                    )
                    raise HTTPException(
                        status_code=404,
                        detail=f"No active check-in found for room {room_number}",
                    )

                guest_id = checkin_record.get("guest_id")
                if not guest_id:
                    logger.error(
                        "Confirmed check-in missing guest_id",
                        extra={
                            "room_number": room_number,
                            "checkin_id": checkin_record.get("checkin_id"),
                            "operation_result": "invalid_checkin_data",
                        },
                    )
                    raise HTTPException(
                        status_code=500,
                        detail=f"Invalid check-in data for room {room_number} (missing guest_id)",
                    )

                folder_id = (
                    checkin_record.get("folder_id")
                    or guest_id
                )

                logger.info(
                    "Guest found for room",
                    extra={
                        "room_number": room_number,
                        "guest_id": guest_id,
                        "folder_id": folder_id,
                        "checkin_id": checkin_record.get("checkin_id"),
                        "operation_result": "success",
                    },
                )

                return {"guest_id": guest_id, "folder_id": folder_id}

            except HTTPException:
                # No loguear como "database_error" los errores esperados HTTP (404/400/500)
                raise

            except Exception as e:
                logger.error(
                    "Failed to get guest from room",
                    extra={
                        "room_number": room_number,
                        "error_message": str(e),
                        "error_type": type(e).__name__,
                        "operation_result": "database_error",
                    },
                    exc_info=True,
                )
                raise

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
                "Checking guest payment status for check-out",
                extra={
                    "folder_id": folder_id,
                    "service": "billing",
                    "operation": "check_payment_status",
                },
            )

            if not self.billing_service_url:
                logger.warning(
                    "Billing service URL not configured, allowing check-out",
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
                    # Buscar todos los tickets del guest usando la API de billing (igual que check-in)
                    params = {
                        "folderId": folder_id,
                    }

                    logger.debug(
                        "Making request to billing service for check-out",
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
                            "Retrieved tickets from billing service for check-out",
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
                            "Payment status check completed for check-out",
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
                            "No tickets found for guest during check-out",
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
                            "Billing service returned error during check-out",
                            extra={
                                "folder_id": folder_id,
                                "http_status": response.status_code,
                                "service": "billing",
                                "error_type": "http_error",
                                "operation_result": "service_error"
                            },
                        )
                        # En caso de error del servicio, permitimos el check-out pero registramos el problema
                        return {
                            "has_unpaid_tickets": False,
                            "unpaid_tickets": [],
                            "total_unpaid_amount": 0.0,
                            "payment_check_skipped": True,
                            "error": f"Billing service error: {response.status_code}",
                        }

            except httpx.RequestError as exc:
                logger.error(
                    "Failed to connect to billing service during check-out",
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
                # En caso de error de conexión, permitimos el check-out pero registramos el problema
                return {
                    "has_unpaid_tickets": False,
                    "unpaid_tickets": [],
                    "total_unpaid_amount": 0.0,
                    "payment_check_skipped": True,
                    "error": f"Billing service unavailable: {str(exc)}",
                }

    async def save_checkout_record(self, room_number: str, guest_id: str) -> str:
        """
        Persiste el registro de salida en MongoDB con tracing.
        """
        with tracer.start_as_current_span("database.mongodb_save_checkout"):
            logger.info(
                "Saving check-out record to database",
                extra={
                    "room_number": room_number,
                    "guest_id": guest_id,
                    "operation": "save_checkout"
                },
            )

            checkout_id = f"checkout_{room_number}_{int(datetime.now().timestamp())}"

            checkout_record = {
                "checkout_id": checkout_id,
                "room_number": room_number,
                "guest_id": guest_id,
                "checkout_time": datetime.now().astimezone().isoformat(),
                "status": "confirmed",
            }

            # Agregar información de estado de pagos si está disponible
            try:
                await self.checkouts.insert_one(checkout_record)

                logger.info(
                    "Check-out record saved successfully",
                    extra={
                        "checkout_id": checkout_id,
                        "room_number": room_number,
                        "guest_id": guest_id,
                        "operation_result": "success"
                    },
                )

                return checkout_id
            except Exception as e:
                logger.error(
                    "Failed to save check-out record to database",
                    extra={
                        "room_number": room_number,
                        "guest_id": guest_id,
                        "error_message": str(e),
                        "error_type": type(e).__name__,
                        "operation_result": "database_error"
                    },
                    exc_info=True,
                )
                raise HTTPException(
                    status_code=500, detail="Error al guardar en base de datos"
                )


async def process_check_out(room_number: str, ignore_unpaid_tickets: bool = False) -> dict:
    """
    Procesa el check-out incluyendo verificación de pagos pendientes.
    
    Args:
        room_number: Numero de la habitación
        ignore_unpaid_tickets: Si es True, permite check-out aunque haya tickets sin pagar
    """
    # Span principal del flujo de salida
    with tracer.start_as_current_span("process_checkout_flow"):
        logger.info(
            "Starting check-out process",
            extra={
                "room_number": room_number,
                "ignore_unpaid_tickets": ignore_unpaid_tickets,
                "flow": "check_out"
            },
        )

        service = CheckOutService()


        # 1. Obtener guest_id y folder_id de la habitación
        guest_info = await service.get_guest_from_room(room_number)
        guest_id = guest_info["guest_id"]
        folder_id = guest_info["folder_id"]

        # 2. Verificar estado de pagos del guest
        payment_status = await service.check_guest_payment_status(folder_id)

        # 3. Decidir si proceder con el check-out
        if payment_status["has_unpaid_tickets"] and not ignore_unpaid_tickets:
            logger.warning(
                "Check-out blocked due to unpaid tickets",
                extra={
                    "room_number": room_number,
                    "guest_id": guest_id,
                    "unpaid_count": len(payment_status["unpaid_tickets"]),
                    "total_unpaid": payment_status["total_unpaid_amount"],
                    "flow": "check_out",
                    "operation_result": "blocked"
                },
            )
            raise HTTPException(
                status_code=402,  # Payment Required
                detail={
                    "message": "Check-out blocked: Guest has unpaid tickets",
                    "unpaid_tickets": payment_status["unpaid_tickets"],
                    "total_unpaid_amount": payment_status["total_unpaid_amount"],
                }
            )

        checkout_id = await service.save_checkout_record(room_number, guest_id)

        # 5. Notificar a Kafka (Limpieza Final)
        kafka_event = {
            "roomNumber": room_number,
            "guestId": guest_id,
            "timestamp": datetime.now().astimezone().isoformat()
        }

        try:
            logger.info(
                "Sending check-out event to Kafka",
                extra={
                    "checkout_id": checkout_id,
                    "room_id": room_number,
                    "guest_id": guest_id,
                    "topic": "check-out",
                    "operation": "kafka_publish"
                },
            )

            await MessageProducer.get_instance().send_message(
                topic=MessageProducer.get_instance().check_out_topic,
                value=kafka_event,
            )

            logger.info(
                "Check-out event sent to Kafka successfully",
                extra={
                    "checkout_id": checkout_id,
                    "room_id": room_number,
                    "guest_id": guest_id,
                    "topic": MessageProducer.get_instance().check_out_topic,
                    "operation_result": "success"
                },
            )
        except Exception as e:
            logger.error(
                "Failed to send check-out event to Kafka",
                extra={
                    "checkout_id": checkout_id,
                    "room_id": room_number,
                    "guest_id": guest_id,
                    "error_message": str(e),
                    "error_type": type(e).__name__,
                    "topic": "check-out",
                    "operation_result": "kafka_error"
                },
                exc_info=True,
            )

        logger.info(
            "Check-out process completed",
            extra={
                "checkout_id": checkout_id,
                "room_id": room_number,
                "guest_id": guest_id,
                "payment_status_checked": True,
                "had_unpaid_tickets": payment_status["has_unpaid_tickets"],
                "total_unpaid_amount": payment_status.get("total_unpaid_amount", 0.0),
                "payment_check_skipped": payment_status.get("payment_check_skipped", False),
                "flow": "check_out",
                "operation_result": "success"
            },
        )

        return {
            "status": "success",
            "message": f"Check-out successful for room {room_number}",
            "checkout_id": checkout_id,
            "room_number": room_number,
            "guest_id": guest_id,
            "timestamp": datetime.now().astimezone().isoformat(),
            "payment_status": payment_status,
        }
