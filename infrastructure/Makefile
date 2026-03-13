SHELL := /bin/bash

COMPOSE ?= docker compose
ROOT_COMPOSE := docker-compose.yml
METRICS_COMPOSE := metrics/docker-compose.yml
ANALYTICS_COMPOSE := kafka/docker-compose.analytics.yml
NETWORK := infra_shared

.PHONY: network validate validate-root validate-metrics validate-analytics \
        up down restart ps logs \
        metrics-up metrics-down metrics-restart metrics-logs \
        analytics-up analytics-down analytics-restart analytics-logs \
        alertmanager-logs prometheus-logs grafana-logs kafka-logs

network:
	@docker network inspect $(NETWORK) >/dev/null 2>&1 || docker network create $(NETWORK)

validate: validate-root validate-metrics validate-analytics

validate-root: network
	$(COMPOSE) -f $(ROOT_COMPOSE) config >/dev/null

validate-metrics: network
	$(COMPOSE) -f $(METRICS_COMPOSE) config >/dev/null

validate-analytics: network
	$(COMPOSE) -f $(ANALYTICS_COMPOSE) config >/dev/null

up: network
	$(COMPOSE) -f $(ROOT_COMPOSE) up -d

down:
	$(COMPOSE) -f $(ROOT_COMPOSE) down

restart: down up

ps:
	$(COMPOSE) -f $(ROOT_COMPOSE) ps

logs:
	$(COMPOSE) -f $(ROOT_COMPOSE) logs -f --tail=200

metrics-up: network
	$(COMPOSE) -f $(METRICS_COMPOSE) up -d

metrics-down:
	$(COMPOSE) -f $(METRICS_COMPOSE) down

metrics-restart: metrics-down metrics-up

metrics-logs:
	$(COMPOSE) -f $(METRICS_COMPOSE) logs -f --tail=200

analytics-up: network
	$(COMPOSE) -f $(ANALYTICS_COMPOSE) up -d

analytics-down:
	$(COMPOSE) -f $(ANALYTICS_COMPOSE) down

analytics-restart: analytics-down analytics-up

analytics-logs:
	$(COMPOSE) -f $(ANALYTICS_COMPOSE) logs -f --tail=200

alertmanager-logs:
	$(COMPOSE) -f $(METRICS_COMPOSE) logs -f --tail=200 alertmanager

prometheus-logs:
	$(COMPOSE) -f $(METRICS_COMPOSE) logs -f --tail=200 prometheus

grafana-logs:
	$(COMPOSE) -f $(METRICS_COMPOSE) logs -f --tail=200 grafana

kafka-logs:
	$(COMPOSE) -f $(ANALYTICS_COMPOSE) logs -f --tail=200 kafka zookeeper kafka-connect
