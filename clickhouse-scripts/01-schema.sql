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
