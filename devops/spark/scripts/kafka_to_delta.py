"""
Kafka to Delta Bronze Streaming Job
Reads events from Kafka topic florus-user-events and writes to Delta Lake Bronze layer on MinIO.

Usage:
    spark-submit --packages org.apache.spark:spark-sql-kafka-0-10_2.12:3.5.0,io.delta:delta-spark_2.12:3.0.0,org.apache.hadoop:hadoop-aws:3.3.4,com.amazonaws:aws-java-sdk-bundle:1.12.262 \
        /scripts/kafka_to_delta.py
"""

import os
from pyspark.sql import SparkSession
from pyspark.sql.functions import (
    col, from_json, current_timestamp,
    year, month, dayofmonth
)
from pyspark.sql.types import (
    StructType, StructField, StringType,
    LongType
)

# Kafka Configuration
KAFKA_BOOTSTRAP_SERVERS = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:29092")
KAFKA_TOPIC = os.getenv("KAFKA_TOPIC", "florus-user-events")

# MinIO Configuration
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "http://florus-minio:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin123")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "florus-data")

# Delta Lake paths (S3A protocol for MinIO)
DELTA_BRONZE_PATH = f"s3a://{MINIO_BUCKET}/delta/bronze/events"
CHECKPOINT_PATH = f"s3a://{MINIO_BUCKET}/checkpoints/events"

# Event schema matching the Go backend event structure
EVENT_SCHEMA = StructType([
    StructField("event_id", LongType(), True),
    StructField("user_id", LongType(), True),
    StructField("session_id", StringType(), True),
    StructField("event_type", StringType(), True),
    StructField("product_id", LongType(), True),
    StructField("category_id", LongType(), True),
    StructField("search_query", StringType(), True),
    StructField("referrer", StringType(), True),
    StructField("metadata", StringType(), True),
    StructField("created_at", StringType(), True),
])


def create_spark_session():
    """Create Spark session with Delta Lake, Kafka, and MinIO (S3A) support."""
    return SparkSession.builder \
        .appName("Florus-Kafka-to-Delta") \
        .config("spark.sql.extensions", "io.delta.sql.DeltaSparkSessionExtension") \
        .config("spark.sql.catalog.spark_catalog", "org.apache.spark.sql.delta.catalog.DeltaCatalog") \
        .config("spark.streaming.stopGracefullyOnShutdown", "true") \
        .config("spark.hadoop.fs.s3a.endpoint", MINIO_ENDPOINT) \
        .config("spark.hadoop.fs.s3a.access.key", MINIO_ACCESS_KEY) \
        .config("spark.hadoop.fs.s3a.secret.key", MINIO_SECRET_KEY) \
        .config("spark.hadoop.fs.s3a.path.style.access", "true") \
        .config("spark.hadoop.fs.s3a.impl", "org.apache.hadoop.fs.s3a.S3AFileSystem") \
        .config("spark.hadoop.fs.s3a.connection.ssl.enabled", "false") \
        .config("spark.hadoop.fs.s3a.aws.credentials.provider", "org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider") \
        .getOrCreate()


def main():
    print("=" * 60)
    print("Florus Kafka to Delta Streaming Job")
    print("=" * 60)
    print(f"Kafka servers: {KAFKA_BOOTSTRAP_SERVERS}")
    print(f"Kafka topic: {KAFKA_TOPIC}")
    print(f"MinIO endpoint: {MINIO_ENDPOINT}")
    print(f"Delta Bronze path: {DELTA_BRONZE_PATH}")
    print(f"Checkpoint path: {CHECKPOINT_PATH}")
    print("=" * 60)

    spark = create_spark_session()
    spark.sparkContext.setLogLevel("WARN")

    # Read from Kafka
    kafka_df = spark.readStream \
        .format("kafka") \
        .option("kafka.bootstrap.servers", KAFKA_BOOTSTRAP_SERVERS) \
        .option("subscribe", KAFKA_TOPIC) \
        .option("startingOffsets", "earliest") \
        .option("failOnDataLoss", "false") \
        .load()

    # Parse JSON value from Kafka
    events_df = kafka_df \
        .selectExpr("CAST(value AS STRING) as json_value", "timestamp as kafka_timestamp") \
        .select(
            from_json(col("json_value"), EVENT_SCHEMA).alias("event"),
            col("kafka_timestamp")
        ) \
        .select(
            col("event.*"),
            col("kafka_timestamp"),
            current_timestamp().alias("processed_at")
        ) \
        .withColumn("year", year(col("kafka_timestamp"))) \
        .withColumn("month", month(col("kafka_timestamp"))) \
        .withColumn("day", dayofmonth(col("kafka_timestamp")))

    # Write to Delta Lake on MinIO with partitioning
    query = events_df.writeStream \
        .format("delta") \
        .outputMode("append") \
        .option("checkpointLocation", CHECKPOINT_PATH) \
        .partitionBy("year", "month", "day") \
        .start(DELTA_BRONZE_PATH)

    print(f"Streaming query started. Writing to {DELTA_BRONZE_PATH}")
    print("Press Ctrl+C to stop...")

    query.awaitTermination()


if __name__ == "__main__":
    main()
