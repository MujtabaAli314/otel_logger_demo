-- Run this ONCE, manually, after the otel-collector has started successfully
-- at least one time and created otel.otel_logs. It intentionally does NOT
-- live in clickhouse-init/ (docker-entrypoint-initdb.d), because otel_logs
-- doesn't exist yet at ClickHouse's own first-boot time — it's only created
-- later, when the collector connects and runs its own create_schema step.
--
-- Usage:
--   docker exec -i clickhouse clickhouse-client --multiquery < create-log-mv.sql

CREATE MATERIALIZED VIEW IF NOT EXISTS otel.mv_logs_to_custom
TO default.logs
AS SELECT
    Timestamp                      AS timestamp,
    LogAttributes['msgId']         AS msgId,
    TraceId                        AS traceId,
    SpanId                         AS spanId,
    LogAttributes['parentSpanId']  AS parentSpanId,
    ServiceName                    AS service,
    SeverityText                   AS code,
    Body                           AS msg,
    SeverityText                   AS msgType
FROM otel.otel_logs;

-- Optional one-time backfill: a materialized view only processes rows
-- inserted AFTER it's created. If you want the logs already sitting in
-- otel_logs to also show up in `logs`, run this once as well:
--
-- INSERT INTO default.logs
-- SELECT
--     Timestamp                      AS timestamp,
--     LogAttributes['msgId']         AS msgId,
--     TraceId                        AS traceId,
--     SpanId                         AS spanId,
--     LogAttributes['parentSpanId']  AS parentSpanId,
--     ServiceName                    AS service,
--     SeverityText                   AS code,
--     Body                           AS msg,
--     SeverityText                   AS msgType
-- FROM otel.otel_logs;
