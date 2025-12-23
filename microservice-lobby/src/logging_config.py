"""
Structured JSON logging configuration with OpenTelemetry trace correlation.
This module sets up JSON-formatted logging that includes trace_id and span_id
for correlation with distributed traces.
"""

import logging
import sys
from pythonjsonlogger import jsonlogger
from opentelemetry import trace


class CustomJsonFormatter(jsonlogger.JsonFormatter):
    """
    Custom JSON formatter that adds OpenTelemetry trace context to every log record.
    This enables correlation between logs and traces in observability backends.
    """

    def add_fields(self, log_record, record, message_dict):
        super(CustomJsonFormatter, self).add_fields(log_record, record, message_dict)

        # Add standard fields
        log_record["timestamp"] = self.formatTime(record, self.datefmt)
        log_record["level"] = record.levelname
        log_record["logger"] = record.name
        log_record["module"] = record.module
        log_record["function"] = record.funcName
        log_record["line"] = record.lineno

        # Add OpenTelemetry trace context for correlation
        span = trace.get_current_span()
        if span:
            span_context = span.get_span_context()
            if span_context and span_context.is_valid:
                # Format trace_id and span_id as hex strings (32 and 16 chars respectively)
                log_record["trace_id"] = format(span_context.trace_id, "032x")
                log_record["span_id"] = format(span_context.span_id, "016x")
                log_record["trace_flags"] = span_context.trace_flags

        # Add exception info if present
        if record.exc_info:
            log_record["exception"] = self.formatException(record.exc_info)


def setup_json_logging(log_level=logging.INFO):
    """
    Configure JSON logging for the entire application with trace correlation.

    Args:
        log_level: The logging level (default: INFO)
    """
    # Create JSON formatter
    formatter = CustomJsonFormatter(
        fmt="%(timestamp)s %(level)s %(name)s %(message)s", datefmt="%Y-%m-%dT%H:%M:%S"
    )

    # Configure root logger
    root_logger = logging.getLogger()
    root_logger.setLevel(log_level)

    # Remove existing handlers to avoid duplicates
    for handler in root_logger.handlers[:]:
        root_logger.removeHandler(handler)

    # Add console handler with JSON formatting
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(log_level)
    console_handler.setFormatter(formatter)
    root_logger.addHandler(console_handler)

    # Configure uvicorn loggers to use JSON format
    for logger_name in ["uvicorn", "uvicorn.access", "uvicorn.error"]:
        logger = logging.getLogger(logger_name)
        logger.handlers.clear()
        logger.addHandler(console_handler)
        logger.setLevel(log_level)
        logger.propagate = False

    return root_logger


def get_logger(name: str) -> logging.Logger:
    """
    Get a logger instance with the specified name.

    Args:
        name: The logger name (typically __name__)

    Returns:
        A configured logger instance
    """
    return logging.getLogger(name)
