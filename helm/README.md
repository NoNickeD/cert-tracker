# Cert-Tracker Helm Chart Documentation

## Overview

The Cert-Tracker Helm Chart facilitates the deployment of a Kubernetes-native microservice that continuously monitors SSL/TLS certificates, reporting on their expiration status. This chart not only deploys the `cert-tracker` service but also includes Redis as a caching layer to optimize certificate data retrieval. It is equipped with features such as scheduled certificate checks via Kubernetes CronJobs, Redis cache management, autoscaling using Kubernetes Horizontal Pod Autoscaler (HPA), and Prometheus-compatible metrics for monitoring.

## Configuration

The Helm chart provides a robust set of configurable parameters, allowing users to tailor the deployment to specific requirements. Below is a detailed list of these parameters:

### Global Parameters

| Parameter                             | Description                                                      | Default                |
|---------------------------------------|------------------------------------------------------------------|------------------------|
| `nameOverride`                        | Override the name of the chart                                   | `""`                   |
| `fullnameOverride`                    | Override the full name of the chart                              | `""`                   |
| `namespaceOverride`                   | Override the deployment namespace                                | Chart's namespace      |

### Image Parameters

| Parameter                             | Description                                                      | Default               |
|---------------------------------------|------------------------------------------------------------------|-----------------------|
| `image.repository`                    | Container image repository for Cert-Tracker                      | `ttl.sh/7417837a-c117-4371-9b99-8da5a3f33580` |
| `image.tag`                           | Image tag for Cert-Tracker                                       | `2h`                  |
| `image.pullPolicy`                    | Image pull policy                                                | `IfNotPresent`        |

### Service Parameters

| Parameter                             | Description                                                      | Default                |
|---------------------------------------|------------------------------------------------------------------|------------------------|
| `service.port`                        | Port exposed by the Cert-Tracker service                         | `8080`                 |
| `service.type`                        | Kubernetes service type                                          | `ClusterIP`            |

### Redis Parameters

| Parameter                             | Description                                                      | Default                |
|---------------------------------------|------------------------------------------------------------------|------------------------|
| `redis.image.repository`              | Redis container image repository                                 | `redis`                |
| `redis.image.tag`                     | Redis container image tag                                        | `latest`               |
| `redis.image.port`                    | Redis container port                                             | `6379`                 |
| `redis.resources.requests.memory`     | Memory resource requests for Redis                               | `64Mi`                 |
| `redis.resources.limits.memory`       | Memory resource limits for Redis                                 | `256Mi`                |
| `redis.schedule`                      | Cron schedule for clearing Redis cache                           | `"0 0 * * *"`          |

### CronJob Parameters

| Parameter                             | Description                                                      | Default                |
|---------------------------------------|------------------------------------------------------------------|------------------------|
| `checkjob.schedule`                   | Cron schedule for periodic certificate checks                    | `"0 0 * * *"`          |

### Resource Parameters

| Parameter                             | Description                                                      | Default                |
|---------------------------------------|------------------------------------------------------------------|------------------------|
| `resources.requests.memory`           | Memory resource requests for Cert-Tracker                        | `256Mi`                |
| `resources.limits.memory`             | Memory resource limits for Cert-Tracker                          | `512Mi`                |
| `resources.requests.cpu`              | CPU resource requests for Cert-Tracker                           | `100m`                 |
| `resources.limits.cpu`                | CPU resource limits for Cert-Tracker                             | `500m`                 |

### Horizontal Pod Autoscaler (HPA) Parameters

| Parameter                             | Description                                                      | Default                |
|---------------------------------------|------------------------------------------------------------------|------------------------|
| `hpa.minReplicas`                     | Minimum number of replicas                                       | `1`                    |
| `hpa.maxReplicas`                     | Maximum number of replicas                                       | `10`                   |
| `hpa.cpuUtilization`                  | Target CPU utilization for scaling                               | `80`                   |
| `hpa.memoryUtilization`               | Target memory utilization for scaling                            | `80`                   |

### Monitoring and Metrics Parameters

| Parameter                             | Description                                                      | Default                |
|---------------------------------------|------------------------------------------------------------------|------------------------|
| `monitoring.namespace`                | Namespace for Prometheus ServiceMonitor                          | `monitoring`           |

## Monitoring and Logs

To monitor the Cert-Tracker service and Redis, use the following commands:

- Cert-Tracker Logs:
```bash
kubectl logs -f deployment/cert-tracker --namespace cert-tracker
```

- Redis Logs:
```bash
kubectl logs -f deployment/redis --namespace cert-tracker
```

