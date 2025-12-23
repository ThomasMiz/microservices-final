# CHotel E2E Tests

End-to-end integration tests for the CHotel microservices architecture.

## Overview

This repository contains E2E tests that validate the complete integration of all CHotel microservices:

- **billing** - Payment processing and ticket management (Rust)
- **cleaning** - Room cleaning coordination (Java)
- **kitchen** - Kitchen order processing (Java/Quarkus)
- **lobby** - Guest lobby services (Python/FastAPI)
- **reservations** - Room reservation management (Java)
- **roomservice** - In-room dining orders (Go)

## Technology Stack

- **Runtime**: [Bun](https://bun.sh) - Fast all-in-one JavaScript runtime
- **Testing**: `bun:test` - Built-in Jest-compatible test runner
- **Language**: TypeScript
- **Kafka**: kafkajs
- **PostgreSQL**: pg library

## Project Structure

```
chotel-e2e-tests/
├── src/
│   ├── tests/
│   │   ├── scenarios/           # Business flow tests
│   │   │   ├── order-flow.test.ts
│   │   │   ├── reservation-flow.test.ts
│   │   │   └── cleaning-flow.test.ts
│   │   └── utils/               # Test utilities
│   │       ├── service-clients.ts
│   │       ├── kafka-utils.ts
│   │       ├── db-utils.ts
│   │       └── k8s-utils.ts
│   ├── config/
│   │   └── test-config.ts
│   └── types/
│       └── test-types.ts
├── scripts/
│   ├── health-check.ts
│   ├── setup-test-env.ts
│   └── cleanup-test-env.ts
└── helm/
    └── e2e-environment/
```

## Getting Started

### Prerequisites

- [Bun](https://bun.sh) >= 1.2
- Access to staging Kubernetes cluster
- PostgreSQL and Kafka connectivity

### Installation

```bash
bun install
```

### Running Tests

```bash
# Run all tests
bun test

# Run with coverage
bun test --coverage

# Run specific scenario
bun test src/tests/scenarios/order-flow.test.ts

# Watch mode
bun test --watch
```

### Environment Setup

```bash
# Check service health
bun run health-check

# Setup test environment
bun run setup

# Cleanup test data
bun run cleanup
```

## Configuration

Environment variables can be set to configure the test environment:

| Variable | Description | Default |
|----------|-------------|---------|
| `STAGING_BASE_URL` | Base URL for staging services (external) | `https://staging.chotel.cuini.me` |
| `STAGING_NAMESPACE` | Kubernetes namespace for service discovery (internal) | (none - uses external URLs) |
| `KAFKA_BROKERS` | Kafka broker addresses | `kafka.shared-infrastructure.svc.cluster.local:9092` |
| `POSTGRES_HOST` | PostgreSQL host | `chotel-default-db.shared-infrastructure.svc.cluster.local` |
| `POSTGRES_PASSWORD` | PostgreSQL password | (required) |

## Pipeline Integration

This repository is triggered by downstream pipelines from microservice repositories using GitLab's native pipeline trigger mechanism.

### Deployment Architecture

**Namespace Isolation Strategy:**
- **Staging**: Each microservice deploys to a separate staging namespace (e.g., `microservice-roomservice-staging`)
- **Production**: Each microservice deploys to a separate production namespace (e.g., `microservice-roomservice-prod`)

This ensures complete isolation between staging and production environments, preventing cross-environment conflicts.

### Pipeline Flow

1. **Microservice Change** - Developer pushes to main branch
2. **Build & Test** - Microservice runs unit tests and builds Docker image
3. **Deploy to Staging** - Microservice deploys to its staging namespace
4. **Trigger E2E Tests** - Microservice triggers this E2E pipeline with `strategy: depend`
5. **E2E Tests Run** - Tests run against the staging environment
6. **Test Results** - Pipeline waits for E2E test completion
7. **Automatic Production Deploy** - If tests pass, automatic production deployment occurs
8. **Deploy to Production** - Microservice automatically deploys to its production namespace

### GitLab Environments

Each microservice uses GitLab environments for deployment tracking:

**Staging Environment:**
```yaml
environment:
  name: staging
  url: https://staging.chotel.cuini.me/api/{service}
  deployment_tier: staging
```

**Production Environment:**
```yaml
environment:
  name: production
  url: https://chotel.cuini.me/api/{service}
  deployment_tier: production
  # Automatic deployment after E2E tests pass
```

### Trigger Variables

When triggered, the following variables are passed:

| Variable | Description | Example |
|----------|-------------|---------|
| `SERVICE_CHANGED` | Name of the service that triggered the tests | `roomservice` |
| `SERVICE_VERSION` | Git commit SHA of the service | `abc123f` |
| `STAGING_NAMESPACE` | Kubernetes namespace for staging deployment | `microservice-roomservice-staging` |
| `UPSTREAM_PIPELINE_ID` | ID of the triggering pipeline | `12345` |
| `UPSTREAM_PROJECT_PATH` | GitLab project path of the triggering service | `chotel/microservice-roomservice` |

### Service Discovery

The E2E tests automatically discover services based on the `STAGING_NAMESPACE` variable:

**With Namespace (GitLab CI):**
- Uses internal Kubernetes DNS
- Format: `http://{service}.{namespace}.svc.cluster.local`
- Example: `http://roomservice.microservice-roomservice-staging.svc.cluster.local`

**Without Namespace (Local Testing):**
- Falls back to external ingress URLs
- Format: `https://staging.chotel.cuini.me/api/{service}`

### Production Deployment Gates

Production deployments require manual approval to ensure:
- All E2E tests have passed
- Changes have been reviewed
- Timing is appropriate for production release

To approve production deployment:
1. Go to GitLab Pipeline view
2. Navigate to the `deploy-production` stage
3. Click the manual play button to approve

## Test Scenarios

### Order Flow (`order-flow.test.ts`)
Tests the complete room service order processing:
1. Create reservation
2. Check-in guest
3. Create room service order
4. Verify Kafka message to kitchen
5. Verify billing ticket creation
6. Complete order from kitchen
7. Check-out guest

### Reservation Flow (`reservation-flow.test.ts`)
Tests reservation lifecycle:
1. Create reservation
2. Confirm reservation
3. Check-in (verify Kafka event)
4. Check-out (verify Kafka event)
5. Cancellation handling
6. Double-booking prevention

### Cleaning Flow (`cleaning-flow.test.ts`)
Tests cleaning service integration:
1. Check-out triggers cleaning task
2. Cleaning task completion
3. Multiple task tracking

## Development

### Adding New Tests

1. Create a new test file in `src/tests/scenarios/`
2. Use existing utilities from `src/tests/utils/`
3. Follow the existing test patterns
4. Update this README with the new scenario

### Test Data Management

- Test data is prefixed with `e2e-test-` for easy identification
- Cleanup scripts remove test data after runs
- Use high room numbers (9xx) to avoid conflicts

## Troubleshooting

### Tests Timeout
- Check service health with `bun run health-check`
- Verify Kafka connectivity
- Check database connections
- Ensure correct namespace is being used

### Kafka Connection Issues
- Verify broker addresses
- Check network policies in Kubernetes
- Ensure test consumer group is unique

### Database Errors
- Verify PostgreSQL credentials
- Check database connectivity
- Ensure proper permissions

### Namespace-Related Issues

**Wrong Service URLs:**
- Check that `STAGING_NAMESPACE` is correctly passed from upstream pipeline
- Verify the service name mapping in `test-config.ts`
- Confirm Kubernetes service names match expectations

**Service Not Found:**
- Verify the staging deployment completed successfully
- Check that the namespace exists: `kubectl get ns | grep staging`
- Verify service is running: `kubectl get pods -n {namespace}`

**DNS Resolution Failures:**
- Ensure you're running within the Kubernetes cluster (not locally with K8s DNS)
- For local testing, omit `STAGING_NAMESPACE` to use external URLs
- Check CoreDNS is functioning: `kubectl get pods -n kube-system`

### Checking Staging Deployments

```bash
# List all staging namespaces
kubectl get ns | grep staging

# Check deployments in a specific staging namespace
kubectl get all -n microservice-roomservice-staging

# View logs from a staging service
kubectl logs -n microservice-roomservice-staging deployment/roomservice

# Port-forward to test a staging service locally
kubectl port-forward -n microservice-roomservice-staging svc/roomservice 8080:80
```
