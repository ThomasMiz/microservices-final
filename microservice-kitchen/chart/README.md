# Microservice Kitchen Helm Chart

This Helm chart deploys the microservice-kitchen Quarkus application to Kubernetes.

## Installation

```bash
helm install my-release ./chart
```

## Configuration

The following table lists the configurable parameters of the microservice-kitchen chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `microservice-kitchen` |
| `image.tag` | Image tag | `1.0.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `250m` |
| `resources.requests.memory` | Memory request | `256Mi` |

## Health Checks

The application includes built-in health checks:
- Liveness: `/q/health/live`
- Readiness: `/q/health/ready`

These are automatically configured in the Kubernetes deployment.