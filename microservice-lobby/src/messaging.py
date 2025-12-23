import os
import logging
import json
from typing import Any, Dict
from aiokafka import AIOKafkaProducer
from opentelemetry import trace
from opentelemetry.propagate import inject

tracer = trace.get_tracer(__name__)
logger = logging.getLogger(__name__)


class MessageProducer:
    _instance = None

    def __init__(self):
        self.bootstrap_servers = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
        self.kafka_user = os.getenv("KAFKA_USER")
        self.kafka_password = os.getenv("KAFKA_PASSWORD")
        self.check_in_topic = os.getenv("KAFKA_CHECK_IN_TOPIC", "check-in")
        self.check_out_topic = os.getenv("KAFKA_CHECK_OUT_TOPIC", "check-out")
        self.producer = None

    @classmethod
    def get_instance(cls):
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    async def start(self):
        """Initialize the AIOKafkaProducer."""
        if self.producer is None:
            logger.info(
                "Initializing Kafka producer",
                extra={
                    "bootstrap_servers": self.bootstrap_servers,
                    "component": "kafka_producer",
                },
            )

            # Configure SASL authentication if credentials are provided
            producer_config = {
                "bootstrap_servers": self.bootstrap_servers,
                "value_serializer": lambda v: json.dumps(v).encode("utf-8"),
            }

            if self.kafka_user and self.kafka_password:
                logger.info(
                    "Configuring Kafka SASL authentication",
                    extra={
                        "kafka_user": self.kafka_user,
                        "component": "kafka_producer",
                    },
                )
                producer_config.update(
                    {
                        "security_protocol": "SASL_PLAINTEXT",
                        "sasl_mechanism": "PLAIN",
                        "sasl_plain_username": self.kafka_user,
                        "sasl_plain_password": self.kafka_password,
                    }
                )

            self.producer = AIOKafkaProducer(**producer_config)
            await self.producer.start()

            logger.info(
                "Kafka producer started successfully",
                extra={
                    "bootstrap_servers": self.bootstrap_servers,
                    "component": "kafka_producer",
                },
            )

    async def stop(self):
        """Stop the AIOKafkaProducer."""
        if self.producer:
            logger.info(
                "Stopping Kafka producer", extra={"component": "kafka_producer"}
            )
            await self.producer.stop()
            logger.info(
                "Kafka producer stopped successfully",
                extra={"component": "kafka_producer"},
            )
            self.producer = None

    async def send_message(self, topic: str, value: Dict[str, Any]):
        """Send a message to a Kafka topic."""
        if not self.producer:
            logger.warning(
                "Kafka producer not initialized, attempting to start",
                extra={"topic": topic, "component": "kafka_producer"},
            )
            await self.start()
            if not self.producer:
                raise RuntimeError("Failed to initialize Kafka producer")

        with tracer.start_as_current_span(f"kafka.send_{topic}"):
            try:
                logger.info(
                    "Sending message to Kafka topic",
                    extra={
                        "topic": topic,
                        "kafka_payload": value,
                        "component": "kafka_producer",
                    },
                )

                # Inject trace context into Kafka headers
                headers = {}
                inject(headers)
                
                # Convert headers to the format expected by aiokafka (list of tuples)
                kafka_headers = [(k, v.encode('utf-8') if isinstance(v, str) else v) for k, v in headers.items()]
                
                logger.info(
                    "Injected trace context headers",
                    extra={
                        "headers": headers,
                        "component": "kafka_producer",
                    },
                )

                await self.producer.send_and_wait(topic, value, headers=kafka_headers)

                logger.info(
                    "Message sent to Kafka successfully",
                    extra={"topic": topic, "component": "kafka_producer"},
                )
            except Exception as e:
                logger.error(
                    "Failed to send message to Kafka",
                    extra={
                        "topic": topic,
                        "error": str(e),
                        "error_type": type(e).__name__,
                        "component": "kafka_producer",
                    },
                    exc_info=True,
                )
                # Depending on requirements, we might want to raise this or just log it.
                # For critical business flow, raising might be better if Kafka is mandatory.
                raise e
