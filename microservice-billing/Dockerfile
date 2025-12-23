FROM gcr.io/distroless/cc-debian12

COPY --chown=1001:root --chmod=755 target/release/microservice-billing microservice-billing

EXPOSE 8080
USER 1001

ENTRYPOINT ["./microservice-billing"]
