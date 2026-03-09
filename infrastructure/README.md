# Infrastructure Stack

Local infrastructure for metrics, logs, tracing, and analytics.

**Compose files:**
| File | Purpose |
| :--- | :--- |
| `metrics/docker-compose.yml` | Monitoring stack (Grafana, Prometheus, Loki, Tempo, OTEL Collector, Alertmanager) |
| `kafka/docker-compose.analytics.yml` | Analytics stack (Kafka, Zookeeper, Elasticsearch) |

## Quick Start

1. Create the shared network (one time):
```powershell
docker network create infra_shared
```

2. Start metrics:
```powershell
docker compose -f metrics\docker-compose.yml up -d
```

3. Start analytics:
```powershell
docker compose -f kafka\docker-compose.analytics.yml up -d
```

## Ports (all external)

| Port | Service | Description |
| :--- | :--- | :--- |
| `3000` | Grafana | Web UI |
| `3100` | Loki | Logs API |
| `3200` | Tempo | Traces API |
| `4317` | OTEL Collector | gRPC receiver |
| `4318` | OTEL Collector | HTTP receiver |
| `8888` | OTEL Collector | Metrics |
| `8889` | OTEL Collector | Prometheus exporter |
| `9091` | Prometheus | Web UI / API (host `9091` → container `9090`) |
| `9092` | Kafka | Broker (external listener) |
| `9093` | Alertmanager | Web UI / API |
| `9200` | Elasticsearch | REST API |
| `13133` | OTEL Collector | Health check |

## Services

| Service | Image | Description | Compose |
| :--- | :--- | :--- | :--- |
| Grafana | `grafana/grafana:12.3.0` | Dashboards and visualization | `metrics/docker-compose.yml` |
| Prometheus | `prom/prometheus:v3.8.0` | Metrics and alerting | `metrics/docker-compose.yml` |
| Alertmanager | `prom/alertmanager:v0.30.0` | Alert handling | `metrics/docker-compose.yml` |
| OTEL Collector | `otel/opentelemetry-collector-contrib:0.141.0` | Telemetry collection and routing | `metrics/docker-compose.yml` |
| Loki | `grafana/loki:3.6.2` | Log storage | `metrics/docker-compose.yml` |
| Tempo | `grafana/tempo:2.9.0` | Trace storage | `metrics/docker-compose.yml` |
| Zookeeper | `zookeeper:3.9.2` | Kafka coordination | `kafka/docker-compose.analytics.yml` |
| Kafka | `confluentinc/cp-kafka:7.5.0` | Event broker | `kafka/docker-compose.analytics.yml` |
| Elasticsearch | `docker.elastic.co/elasticsearch/elasticsearch:7.17.10` | Search/analytics | `kafka/docker-compose.analytics.yml` |
| Kafka Init Topics | `confluentinc/cp-kafka:7.5.0` | Topic initialization | `kafka/docker-compose.analytics.yml` |

## Network

All services are attached to the external network `infra_shared`, so different compose stacks can reach each other via container DNS names.
