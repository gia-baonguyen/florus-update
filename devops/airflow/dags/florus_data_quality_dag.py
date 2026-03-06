"""
Florus Data Quality DAG
Runs daily at 3:00 AM to check data quality across the pipeline.
Validates Delta Lake data, Redis cache, and overall system health.
"""

import os
from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.utils.trigger_rule import TriggerRule
import redis
import json
import requests

# Environment variables
SPARK_MASTER_HOST = os.getenv("SPARK_MASTER_HOST", "florus-spark-master")
REDIS_HOST = os.getenv("REDIS_HOST", "florus-redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "http://florus-minio:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin123")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "florus-data")

default_args = {
    "owner": "florus",
    "depends_on_past": False,
    "email_on_failure": False,
    "email_on_retry": False,
    "retries": 1,
    "retry_delay": timedelta(minutes=5),
}


def check_minio_health(**context):
    """Check MinIO service health and bucket accessibility."""
    try:
        import boto3
        from botocore.client import Config

        # Create S3 client for MinIO
        s3_client = boto3.client(
            "s3",
            endpoint_url=MINIO_ENDPOINT,
            aws_access_key_id=MINIO_ACCESS_KEY,
            aws_secret_access_key=MINIO_SECRET_KEY,
            config=Config(signature_version="s3v4"),
        )

        # Check if bucket exists
        response = s3_client.list_buckets()
        buckets = [b["Name"] for b in response["Buckets"]]

        if MINIO_BUCKET in buckets:
            # List objects in delta path
            objects = s3_client.list_objects_v2(
                Bucket=MINIO_BUCKET,
                Prefix="delta/",
                MaxKeys=10
            )

            object_count = objects.get("KeyCount", 0)
            print(f"MinIO health: OK - Found {object_count} objects in delta/")

            context["ti"].xcom_push(key="minio_health", value={
                "status": "healthy",
                "bucket_exists": True,
                "delta_objects": object_count
            })
        else:
            print(f"Warning: Bucket {MINIO_BUCKET} not found")
            context["ti"].xcom_push(key="minio_health", value={
                "status": "warning",
                "bucket_exists": False
            })

    except Exception as e:
        print(f"Error checking MinIO: {e}")
        context["ti"].xcom_push(key="minio_health", value={
            "status": "error",
            "error": str(e)
        })


def check_redis_health(**context):
    """Check Redis cache health and recommendation data."""
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

        # Ping Redis
        r.ping()

        # Check recommendation keys
        user_rec_keys = len(r.keys("spark:recommendations:user:*"))
        product_sim_keys = len(r.keys("spark:similar:product:*"))

        # Check metadata
        last_updated = r.get("spark:recommendations:last_updated")
        user_count = r.get("spark:recommendations:user_count")
        product_count = r.get("spark:recommendations:product_count")

        health_data = {
            "status": "healthy",
            "ping": "OK",
            "user_recommendation_keys": user_rec_keys,
            "product_similarity_keys": product_sim_keys,
            "last_updated": last_updated,
            "user_count": user_count,
            "product_count": product_count
        }

        print(f"Redis health: {health_data}")
        context["ti"].xcom_push(key="redis_health", value=health_data)

    except Exception as e:
        print(f"Error checking Redis: {e}")
        context["ti"].xcom_push(key="redis_health", value={
            "status": "error",
            "error": str(e)
        })


def check_spark_health(**context):
    """Check Spark cluster health."""
    try:
        response = requests.get(
            f"http://{SPARK_MASTER_HOST}:8080/json/",
            timeout=10
        )

        if response.status_code == 200:
            data = response.json()

            health_data = {
                "status": "healthy",
                "workers_alive": len(data.get("aliveworkers", [])),
                "workers_dead": len(data.get("deadworkers", [])),
                "active_apps": len(data.get("activeapps", [])),
                "completed_apps": len(data.get("completedapps", [])),
                "cores_used": data.get("coresused", 0),
                "memory_used": data.get("memoryused", 0)
            }

            print(f"Spark health: {health_data}")
            context["ti"].xcom_push(key="spark_health", value=health_data)
        else:
            print(f"Spark API returned status {response.status_code}")
            context["ti"].xcom_push(key="spark_health", value={
                "status": "error",
                "error": f"HTTP {response.status_code}"
            })

    except Exception as e:
        print(f"Error checking Spark: {e}")
        context["ti"].xcom_push(key="spark_health", value={
            "status": "error",
            "error": str(e)
        })


def check_kafka_connectivity(**context):
    """Check Kafka connectivity (basic check via Spark master container)."""
    try:
        import subprocess

        # Use docker exec to check Kafka from Spark container
        result = subprocess.run(
            ["docker", "exec", "florus-kafka", "kafka-broker-api-versions",
             "--bootstrap-server", "localhost:9092"],
            capture_output=True,
            text=True,
            timeout=30
        )

        if result.returncode == 0:
            health_data = {
                "status": "healthy",
                "message": "Kafka broker responding"
            }
        else:
            health_data = {
                "status": "warning",
                "message": result.stderr[:200] if result.stderr else "Unknown error"
            }

        print(f"Kafka health: {health_data}")
        context["ti"].xcom_push(key="kafka_health", value=health_data)

    except Exception as e:
        print(f"Error checking Kafka: {e}")
        context["ti"].xcom_push(key="kafka_health", value={
            "status": "error",
            "error": str(e)
        })


def generate_quality_report(**context):
    """Generate overall data quality report and store in Redis."""
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

        # Collect all health checks
        minio_health = context["ti"].xcom_pull(key="minio_health", task_ids="check_minio")
        redis_health = context["ti"].xcom_pull(key="redis_health", task_ids="check_redis")
        spark_health = context["ti"].xcom_pull(key="spark_health", task_ids="check_spark")
        kafka_health = context["ti"].xcom_pull(key="kafka_health", task_ids="check_kafka")

        # Determine overall status
        statuses = [
            minio_health.get("status", "unknown") if minio_health else "unknown",
            redis_health.get("status", "unknown") if redis_health else "unknown",
            spark_health.get("status", "unknown") if spark_health else "unknown",
            kafka_health.get("status", "unknown") if kafka_health else "unknown",
        ]

        if all(s == "healthy" for s in statuses):
            overall_status = "healthy"
        elif "error" in statuses:
            overall_status = "error"
        else:
            overall_status = "warning"

        # Build report
        report = {
            "timestamp": datetime.now().isoformat(),
            "overall_status": overall_status,
            "components": {
                "minio": minio_health or {"status": "unknown"},
                "redis": redis_health or {"status": "unknown"},
                "spark": spark_health or {"status": "unknown"},
                "kafka": kafka_health or {"status": "unknown"},
            }
        }

        # Store report in Redis
        r.set("airflow:data_quality:latest_report", json.dumps(report))
        r.set("airflow:data_quality:last_check", datetime.now().isoformat())
        r.set("airflow:data_quality:overall_status", overall_status)

        # Keep history (last 7 days)
        history_key = f"airflow:data_quality:history:{datetime.now().strftime('%Y-%m-%d')}"
        r.set(history_key, json.dumps(report))
        r.expire(history_key, 86400 * 7)

        print(f"Data quality report generated: {overall_status}")
        print(f"Report: {json.dumps(report, indent=2)}")

    except Exception as e:
        print(f"Error generating report: {e}")
        raise


with DAG(
    dag_id="florus_data_quality",
    default_args=default_args,
    description="Daily data quality checks for Florus data pipeline",
    schedule_interval="0 3 * * *",  # Daily at 3:00 AM
    start_date=datetime(2024, 1, 1),
    catchup=False,
    tags=["florus", "data-quality", "monitoring"],
) as dag:

    # Check MinIO health
    check_minio = PythonOperator(
        task_id="check_minio",
        python_callable=check_minio_health,
        provide_context=True,
    )

    # Check Redis health
    check_redis = PythonOperator(
        task_id="check_redis",
        python_callable=check_redis_health,
        provide_context=True,
    )

    # Check Spark health
    check_spark = PythonOperator(
        task_id="check_spark",
        python_callable=check_spark_health,
        provide_context=True,
    )

    # Check Kafka connectivity
    check_kafka = PythonOperator(
        task_id="check_kafka",
        python_callable=check_kafka_connectivity,
        provide_context=True,
    )

    # Generate overall report
    generate_report = PythonOperator(
        task_id="generate_report",
        python_callable=generate_quality_report,
        provide_context=True,
        trigger_rule=TriggerRule.ALL_DONE,
    )

    # All health checks run in parallel, then generate report
    [check_minio, check_redis, check_spark, check_kafka] >> generate_report
