FROM eclipse-temurin:25
WORKDIR /work

RUN chown 1001 /work \
    && chmod "g+rwX" /work \
    && chown 1001:root /work

COPY --chown=1001:root --chmod=755 target/microservice-reservations-*.jar app.jar

EXPOSE 8080
USER 1001

ENTRYPOINT ["sh", "-c"]
CMD ["java -jar app.jar"]
