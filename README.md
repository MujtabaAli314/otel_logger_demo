# OTel Demo — Setup Guide

Local demo stack: Go services → OTel Collector → Jaeger (traces) + ClickHouse (logs),
backed by Postgres for application data.

Run the steps below in order — later steps depend on earlier ones being up and healthy.

## Prerequisites

- Docker + Docker Compose
- Go toolchain (for running the services in step 4)
- This directory containing: `docker-compose.yml`, `otel-collector-config.yaml`,
  `clickhouse-init/`, `clickhouse-scripts/`, `postgres-init/`

---

## 1. Start the infrastructure

```bash
docker compose up -d
docker compose ps
```

Confirm all four containers show as `Up` (Postgres and ClickHouse should say `healthy`):
`postgres`, `clickhouse`, `jaeger`, `otel-collector`.

If a container name doesn't match what's below, it's likely been renamed locally —
substitute your actual container name in the commands that follow.

```bash
docker compose logs otel-collector   # should show both pipelines starting cleanly
```

---

## 2. ClickHouse setup

`clickhouse-init/01-schema.sql` and the `otel_logs` table (created automatically by the
collector's `clickhouse` exporter) only run on a genuinely **fresh** ClickHouse volume.
If you've brought this stack up before, they likely already exist — check first:

```bash
docker exec -i clickhouse clickhouse-client --query "SHOW TABLES FROM otel"
docker exec -i clickhouse clickhouse-client --query "SHOW TABLES FROM default"
```

If `default.logs` is missing, create it manually (idempotent, safe to re-run):

```bash
docker exec -i clickhouse clickhouse-client --multiquery << 'EOF'
CREATE DATABASE IF NOT EXISTS otel;

CREATE TABLE IF NOT EXISTS default.logs
(
    timestamp     DateTime64(9),
    msgId         String,
    traceId       String,
    spanId        String,
    parentSpanId  String,
    service       String,
    code          String,
    msg           String,
    msgType       String
)
ENGINE = MergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (service, timestamp)
TTL toDateTime(timestamp) + INTERVAL 30 DAY;
EOF
```

If `otel.otel_logs` is missing, the `otel-collector` container hasn't started cleanly —
check `docker compose logs otel-collector` before continuing.

**Then create the materialized view** that reshapes `otel_logs` into the `logs` table
above (only run this once `otel_logs` exists — it never runs automatically):

```bash
docker exec -i clickhouse clickhouse-client --multiquery < clickhouse-scripts/create-log-mv.sql
```

Verify:

```sql
-- via the Play UI at http://localhost:8123/play, or clickhouse-client
SELECT count() FROM otel.otel_logs;
SELECT count() FROM default.logs;
```

---

## 3. Postgres setup

Same caveat as ClickHouse: `postgres-init/` scripts only run against a fresh volume.

```bash
docker exec -it postgres psql -U oteldemo -d oteldemo -c "\dt"
```

If `users`/`transactions` aren't listed, create them and seed data manually:

```bash
docker exec -i postgres psql -U oteldemo -d oteldemo -f - < postgres-init/01-schema.sql
docker exec -i postgres psql -U oteldemo -d oteldemo -f - < postgres-init/02-seed.sql
```

> Make sure `postgres-init/01-schema.sql` reflects the current schema (`users` +
> `transactions` only) before running it — if it still has the older `payment_methods`
> version, update it first.

Verify:

```bash
docker exec -it postgres psql -U oteldemo -d oteldemo -c "SELECT * FROM users;"
docker exec -it postgres psql -U oteldemo -d oteldemo -c "SELECT * FROM transactions LIMIT 5;"
```

---

## 4. Run the services

Just run the services my friend :)
NOTE: The clickhouse and postgres scripts are AI generated (I mean not only those...). Test them please O_o


## 5. Testing
In the scripts, you will find a python script that sends tons of requests async so you can test it. you can run it like (just be sure to install aiohttp):
```bash
python scripts/loadtest.py --host localhost:8080 --total 1000 --concurrency 50 --mix 50:50
```

That's it, go the Jaeger UI to see the traces, or the clickhouse to get the logs.
