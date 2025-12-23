use init_tracing_opentelemetry::{Guard, LogTimer};
use tracing_subscriber::fmt::format::FmtSpan;

pub fn init_logging() -> Guard {
    let init_tracing_result = init_tracing_opentelemetry::TracingConfig::production()
        .with_span_events(FmtSpan::FULL)
        .with_log_directives("info")
        .with_timer(LogTimer::Time)
        .init_subscriber();

    match init_tracing_result {
        Ok(guard) => guard,
        Err(e) => panic!("Could not initialize tracing: {e}"),
    }
}
