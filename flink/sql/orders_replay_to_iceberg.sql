CREATE CATALOG lakehouse WITH (
  'type' = 'iceberg',
  'catalog-type' = 'rest',
  'uri' = 'http://iceberg-rest:8181',
  'warehouse' = 's3://warehouse/',
  'io-impl' = 'org.apache.iceberg.aws.s3.S3FileIO',
  's3.endpoint' = 'http://minio:9000',
  's3.path-style-access' = 'true'
);

CREATE DATABASE IF NOT EXISTS lakehouse.default;

CREATE TABLE IF NOT EXISTS orders_kafka (
  order_id STRING,
  customer_id STRING,
  amount DECIMAL(12, 2),
  currency STRING,
  status STRING,
  event_time TIMESTAMP(3),
  WATERMARK FOR event_time AS event_time - INTERVAL '5' SECOND
) WITH (
  'connector' = 'kafka',
  'topic' = 'orders',
  'properties.bootstrap.servers' = 'redpanda:9092',
  'properties.group.id' = 'streamforge-orders',
  'scan.startup.mode' = 'earliest-offset',
  'format' = 'json',
  'json.timestamp-format.standard' = 'ISO-8601',
  'json.ignore-parse-errors' = 'false'
);

CREATE TABLE IF NOT EXISTS lakehouse.default.orders (
  order_id STRING,
  customer_id STRING,
  amount DECIMAL(12, 2),
  currency STRING,
  status STRING,
  event_time TIMESTAMP(3),
  processed_at TIMESTAMP(3),
  PRIMARY KEY (order_id) NOT ENFORCED
) WITH (
  'format-version' = '2',
  'write.upsert.enabled' = 'true'
);

INSERT INTO lakehouse.default.orders
SELECT
  order_id,
  customer_id,
  amount,
  currency,
  status,
  event_time,
  CURRENT_TIMESTAMP
FROM orders_kafka;
