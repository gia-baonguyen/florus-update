"""
ALS-based Recommendation Pipeline
Trains collaborative filtering model using Spark MLlib ALS and exports recommendations to Redis.
Reads data from Delta Lake on MinIO and/or Oracle database.

Usage:
    spark-submit --packages io.delta:delta-spark_2.12:3.0.0,org.apache.hadoop:hadoop-aws:3.3.4,com.amazonaws:aws-java-sdk-bundle:1.12.262 \
        --jars /jars/ojdbc11.jar \
        /scripts/recommendation_pipeline.py

Environment Variables:
    MINIO_ENDPOINT: MinIO endpoint (default: http://minio:9000)
    MINIO_ACCESS_KEY: MinIO access key (default: minioadmin)
    MINIO_SECRET_KEY: MinIO secret key (default: minioadmin)
    MINIO_BUCKET: MinIO bucket name (default: florus-data)
    REDIS_HOST: Redis host (default: redis)
    REDIS_PORT: Redis port (default: 6379)
    ORACLE_JDBC_URL: Oracle JDBC URL (default: jdbc:oracle:thin:@oracle-db:1521/XEPDB1)
"""

import os
import json
import redis
from datetime import datetime, timedelta
from pyspark.sql import SparkSession
from pyspark.sql.functions import (
    col, sum as spark_sum, count, when, lit,
    collect_list, struct, explode, row_number
)
from pyspark.sql.window import Window
from pyspark.ml.recommendation import ALS
from pyspark.ml.evaluation import RegressionEvaluator

# MinIO Configuration
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "http://florus-minio:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin123")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "florus-data")

# Delta Lake paths on MinIO
DELTA_BRONZE_PATH = f"s3a://{MINIO_BUCKET}/delta/bronze/events"
DELTA_GOLD_USER_RECS = f"s3a://{MINIO_BUCKET}/delta/gold/user_recommendations"
DELTA_GOLD_PRODUCT_SIMS = f"s3a://{MINIO_BUCKET}/delta/gold/product_similarities"

# Redis Configuration
REDIS_HOST = os.getenv("REDIS_HOST", "florus-redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))

# Oracle Configuration (fallback)
ORACLE_JDBC_URL = os.getenv("ORACLE_JDBC_URL", "jdbc:oracle:thin:@florus-oracle:1521/XEPDB1")
ORACLE_USER = os.getenv("ORACLE_USER", "florus")
ORACLE_PASSWORD = os.getenv("ORACLE_PASSWORD", "florus123")

# Event weights for implicit ratings
EVENT_WEIGHTS = {
    "purchase": 5.0,
    "add_to_cart": 3.0,
    "add_to_wishlist": 2.5,
    "product_click": 1.5,
    "product_view": 1.0,
}

# Model parameters
ALS_PARAMS = {
    "maxIter": 15,
    "regParam": 0.1,
    "rank": 50,
    "userCol": "user_id_idx",
    "itemCol": "product_id_idx",
    "ratingCol": "implicit_rating",
    "implicitPrefs": True,
    "coldStartStrategy": "drop",
    "nonnegative": True,
}


def create_spark_session():
    """Create Spark session with Delta Lake and MinIO (S3A) support."""
    return SparkSession.builder \
        .appName("Florus-ALS-Recommendations") \
        .config("spark.sql.extensions", "io.delta.sql.DeltaSparkSessionExtension") \
        .config("spark.sql.catalog.spark_catalog", "org.apache.spark.sql.delta.catalog.DeltaCatalog") \
        .config("spark.driver.extraJavaOptions", "-Djava.security.manager=allow") \
        .config("spark.hadoop.fs.s3a.endpoint", MINIO_ENDPOINT) \
        .config("spark.hadoop.fs.s3a.access.key", MINIO_ACCESS_KEY) \
        .config("spark.hadoop.fs.s3a.secret.key", MINIO_SECRET_KEY) \
        .config("spark.hadoop.fs.s3a.path.style.access", "true") \
        .config("spark.hadoop.fs.s3a.impl", "org.apache.hadoop.fs.s3a.S3AFileSystem") \
        .config("spark.hadoop.fs.s3a.connection.ssl.enabled", "false") \
        .config("spark.hadoop.fs.s3a.aws.credentials.provider", "org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider") \
        .getOrCreate()


def get_redis_client():
    """Get Redis client for exporting recommendations."""
    return redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)


def load_events_from_delta(spark):
    """Load user events from Delta Bronze layer on MinIO."""
    print(f"Loading events from Delta Lake: {DELTA_BRONZE_PATH}")

    try:
        events_df = spark.read.format("delta").load(DELTA_BRONZE_PATH)

        # Filter to relevant event types
        events_df = events_df.filter(
            (col("event_type").isin(list(EVENT_WEIGHTS.keys()))) &
            (col("user_id").isNotNull()) &
            (col("product_id").isNotNull()) &
            (col("user_id") > 0) &
            (col("product_id") > 0)
        )

        count = events_df.count()
        print(f"Loaded {count} events from Delta Lake on MinIO")
        return events_df if count > 0 else None
    except Exception as e:
        print(f"Warning: Could not load from Delta Lake: {e}")
        return None


def load_events_from_oracle(spark):
    """Fallback: Load events directly from Oracle database."""
    print(f"Loading events from Oracle: {ORACLE_JDBC_URL}")

    try:
        events_df = spark.read \
            .format("jdbc") \
            .option("url", ORACLE_JDBC_URL) \
            .option("dbtable", "(SELECT * FROM USER_EVENTS WHERE CREATED_AT > SYSDATE - 90)") \
            .option("user", ORACLE_USER) \
            .option("password", ORACLE_PASSWORD) \
            .option("driver", "oracle.jdbc.driver.OracleDriver") \
            .load()

        # Normalize column names to lowercase
        for column in events_df.columns:
            events_df = events_df.withColumnRenamed(column, column.lower())

        count = events_df.count()
        print(f"Loaded {count} events from Oracle")
        return events_df if count > 0 else None
    except Exception as e:
        print(f"Error loading from Oracle: {e}")
        return None


def calculate_implicit_ratings(events_df):
    """Calculate implicit ratings based on event weights."""
    print("Calculating implicit ratings...")

    # Add event weight
    weight_expr = when(col("event_type") == "purchase", lit(EVENT_WEIGHTS["purchase"])) \
        .when(col("event_type") == "add_to_cart", lit(EVENT_WEIGHTS["add_to_cart"])) \
        .when(col("event_type") == "add_to_wishlist", lit(EVENT_WEIGHTS["add_to_wishlist"])) \
        .when(col("event_type") == "product_click", lit(EVENT_WEIGHTS["product_click"])) \
        .otherwise(lit(EVENT_WEIGHTS["product_view"]))

    events_df = events_df.withColumn("weight", weight_expr)

    # Aggregate ratings per user-product pair
    ratings_df = events_df.groupBy("user_id", "product_id") \
        .agg(
            spark_sum("weight").alias("implicit_rating"),
            count("*").alias("interaction_count")
        )

    # Create index mappings for ALS (requires integer indices)
    user_indexer = ratings_df.select("user_id").distinct() \
        .withColumn("user_id_idx", row_number().over(Window.orderBy("user_id")) - 1)

    product_indexer = ratings_df.select("product_id").distinct() \
        .withColumn("product_id_idx", row_number().over(Window.orderBy("product_id")) - 1)

    # Join indices
    ratings_df = ratings_df \
        .join(user_indexer, "user_id") \
        .join(product_indexer, "product_id")

    print(f"Created ratings matrix: {ratings_df.count()} user-product pairs")
    print(f"Unique users: {user_indexer.count()}, Unique products: {product_indexer.count()}")

    return ratings_df, user_indexer, product_indexer


def train_als_model(ratings_df):
    """Train ALS model for collaborative filtering."""
    print("Training ALS model...")
    print(f"Parameters: {ALS_PARAMS}")

    # Split data for evaluation
    train_df, test_df = ratings_df.randomSplit([0.8, 0.2], seed=42)

    # Create and train model
    als = ALS(**ALS_PARAMS)
    model = als.fit(train_df)

    # Evaluate model
    predictions = model.transform(test_df)
    evaluator = RegressionEvaluator(
        metricName="rmse",
        labelCol="implicit_rating",
        predictionCol="prediction"
    )
    rmse = evaluator.evaluate(predictions)
    print(f"Model RMSE: {rmse:.4f}")

    return model, {"rmse": rmse}


def generate_user_recommendations(model, user_indexer, product_indexer, n_recommendations=20):
    """Generate top-N recommendations for each user."""
    print(f"Generating top {n_recommendations} recommendations per user...")

    # Get recommendations from model
    user_recs = model.recommendForAllUsers(n_recommendations)

    # Explode recommendations array
    user_recs_exploded = user_recs.select(
        col("user_id_idx"),
        explode(col("recommendations")).alias("rec")
    ).select(
        col("user_id_idx"),
        col("rec.product_id_idx").alias("product_id_idx"),
        col("rec.rating").alias("score")
    )

    # Map indices back to original IDs
    user_recs_mapped = user_recs_exploded \
        .join(user_indexer, "user_id_idx") \
        .join(product_indexer, "product_id_idx") \
        .select("user_id", "product_id", "score") \
        .orderBy("user_id", col("score").desc())

    return user_recs_mapped


def generate_product_similarities(model, product_indexer, n_similar=10):
    """Generate similar products for each product (item-item CF)."""
    print(f"Generating top {n_similar} similar products...")

    # Item-item similarities using ALS item vectors
    item_recs = model.recommendForAllItems(n_similar)

    # Explode and map
    item_recs_exploded = item_recs.select(
        col("product_id_idx"),
        explode(col("recommendations")).alias("rec")
    ).select(
        col("product_id_idx"),
        col("rec.product_id_idx").alias("similar_product_id_idx"),
        col("rec.rating").alias("similarity_score")
    )

    # Map indices back to original IDs
    item_recs_mapped = item_recs_exploded \
        .join(product_indexer, "product_id_idx") \
        .join(
            product_indexer.withColumnRenamed("product_id", "similar_product_id")
                          .withColumnRenamed("product_id_idx", "similar_product_id_idx"),
            "similar_product_id_idx"
        ) \
        .select("product_id", "similar_product_id", "similarity_score") \
        .filter(col("product_id") != col("similar_product_id")) \
        .orderBy("product_id", col("similarity_score").desc())

    return item_recs_mapped


def export_to_redis(user_recs_df, product_sims_df, redis_client):
    """Export recommendations and similarities to Redis."""
    print("Exporting to Redis...")

    # Export user recommendations
    user_recs_grouped = user_recs_df.groupBy("user_id") \
        .agg(collect_list(struct("product_id", "score")).alias("recommendations"))

    user_recs_list = user_recs_grouped.collect()
    pipeline = redis_client.pipeline()

    for row in user_recs_list:
        user_id = row["user_id"]
        recs = [{"product_id": int(r["product_id"]), "score": float(r["score"])}
                for r in row["recommendations"]]
        key = f"spark:recommendations:user:{user_id}"
        pipeline.set(key, json.dumps(recs))
        pipeline.expire(key, 86400 * 7)  # 7 days TTL

    pipeline.execute()
    print(f"Exported recommendations for {len(user_recs_list)} users")

    # Export product similarities
    product_sims_grouped = product_sims_df.groupBy("product_id") \
        .agg(collect_list(struct("similar_product_id", "similarity_score")).alias("similar_products"))

    product_sims_list = product_sims_grouped.collect()
    pipeline = redis_client.pipeline()

    for row in product_sims_list:
        product_id = row["product_id"]
        sims = [{"product_id": int(s["similar_product_id"]), "score": float(s["similarity_score"])}
                for s in row["similar_products"]]
        key = f"spark:similar:product:{product_id}"
        pipeline.set(key, json.dumps(sims))
        pipeline.expire(key, 86400 * 7)  # 7 days TTL

    pipeline.execute()
    print(f"Exported similarities for {len(product_sims_list)} products")

    # Store metadata (RFC3339 format for Go parsing)
    redis_client.set("spark:recommendations:last_updated", datetime.now().astimezone().isoformat())
    redis_client.set("spark:recommendations:user_count", len(user_recs_list))
    redis_client.set("spark:recommendations:product_count", len(product_sims_list))


def save_to_delta(user_recs_df, product_sims_df):
    """Save recommendations to Delta Lake Gold layer on MinIO."""
    print("Saving to Delta Lake Gold layer...")

    user_recs_df.write.format("delta").mode("overwrite").save(DELTA_GOLD_USER_RECS)
    print(f"Saved user recommendations to {DELTA_GOLD_USER_RECS}")

    product_sims_df.write.format("delta").mode("overwrite").save(DELTA_GOLD_PRODUCT_SIMS)
    print(f"Saved product similarities to {DELTA_GOLD_PRODUCT_SIMS}")


def main():
    print("=" * 60)
    print("Florus ALS Recommendation Pipeline")
    print(f"Started at: {datetime.now().isoformat()}")
    print("=" * 60)
    print(f"MinIO endpoint: {MINIO_ENDPOINT}")
    print(f"Delta Bronze path: {DELTA_BRONZE_PATH}")
    print(f"Redis: {REDIS_HOST}:{REDIS_PORT}")
    print("=" * 60)

    spark = create_spark_session()
    spark.sparkContext.setLogLevel("WARN")

    # Load events (try Delta first, then Oracle)
    events_df = load_events_from_delta(spark)
    if events_df is None:
        print("Falling back to Oracle database...")
        events_df = load_events_from_oracle(spark)

    if events_df is None:
        print("ERROR: No events found. Cannot train model.")
        spark.stop()
        return

    # Calculate ratings
    ratings_df, user_indexer, product_indexer = calculate_implicit_ratings(events_df)

    # Train model
    model, metrics = train_als_model(ratings_df)

    # Generate recommendations
    user_recs = generate_user_recommendations(model, user_indexer, product_indexer)
    product_sims = generate_product_similarities(model, product_indexer)

    # Export to Redis
    try:
        redis_client = get_redis_client()
        redis_client.ping()
        export_to_redis(user_recs, product_sims, redis_client)
    except Exception as e:
        print(f"Warning: Could not export to Redis: {e}")

    # Also save to Delta Lake Gold layer on MinIO
    try:
        save_to_delta(user_recs, product_sims)
    except Exception as e:
        print(f"Warning: Could not save to Delta Lake: {e}")

    print("=" * 60)
    print(f"Pipeline completed at: {datetime.now().isoformat()}")
    print(f"Metrics: {metrics}")
    print("=" * 60)

    spark.stop()


if __name__ == "__main__":
    main()
