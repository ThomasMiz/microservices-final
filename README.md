# Microservices final

This repository contains a copy of all 9 repositories involved during the final group project for the university course on Microservice Architectures. The project consisted of building a distributed system using a microservices architecture, designing the user flows with domain driven design and implementing a set of CI/CD pipelines on GitLab to automatically deploy all the infrastructure and services to a Kubernetes cluster on AWS EC2.

The project consists of a management system for a hotel with different domains handled by different microservices.

This includes six microservices:
- Billing: written in Rust with Axum and Postgres
- Cleaning: written in Java with Micronaut and Postgres
- Reservations: written in Java with Micronaut and Postgres
- Kitchen: written in Java with Quarkus and Redis
- Lobby: written in Python with FastAPI and MongoDB
- Room service: written in Go with Gin and Redis

And three additional repositories:
- `chotel-front`: static front end (extremely simple, as this was not part of the project's requirements).
- `end-2-end-tests`: contains a set of end-to-end tests to be ran on a staging environment.
- `shared-infrastructure`: contains the charts for the PostgreSQL and MongodB database, Kafka, Redis, OpenTelemetry and OpenObserve, Traefik ingress controller, as well as the terraform code to deploy the infrastructure and the controllers for the AWS Elastic Block Store storage provider and Application Load Balancer.

The original repositories are private in GitLab, so a copy is provided here.

## AWS Infrastructure

The infrastructure in AWS is very simple, as all the complexity is inside the Kubernetes cluster, which is deployed in EC2.

![](readme/aws-infra.png)

## Observability

Distributed tracing provides the ability to follow a request's actions through the different microservices

![](readme/reservations-trace.png)
