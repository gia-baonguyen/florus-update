"""
Florus Recommendation Pipeline DAG
Runs the ALS-based recommendation model training daily at 2:00 AM.
Reads from Delta Lake, trains the model, and exports recommendations to Redis.
"""

import os
from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator, BranchPythonOperator
from airflow.operators.bash import BashOperator
from airflow.operators.empty import EmptyOperator
from airflow.utils.trigger_rule import TriggerRule
import redis
import json

# Environment variables
SPARK_MASTER_HOST = os.getenv("SPARK_MASTER_HOST", "florus-spark-master")
REDIS_HOST = os.getenv("REDIS_HOST", "florus-redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "http://florus-minio:9000")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "florus-data")

default_args = {
    "owner": "florus",
    "depends_on_past": False,
    "email_on_failure": False,
    "email_on_retry": False,
    "retries": 2,
    "retry_delay": timedelta(minutes=5),
}


def check_data_freshness(**context):
    """Check if there's enough fresh data to train the model."""
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

        # Check when data was last updated in Delta Lake
        last_streaming_check = r.get("airflow:streaming:last_check")
        streaming_status = r.get("airflow:streaming:status")

        print(f"Last streaming check: {last_streaming_check}")
        print(f"Streaming status: {streaming_status}")

        # For now, always proceed with training
        # In production, you'd check Delta Lake for recent data
        context["ti"].xcom_push(key="data_check_passed", value=True)
        return "train_model"

    except Exception as e:
        print(f"Error checking data freshness: {e}")
        context["ti"].xcom_push(key="data_check_passed", value=True)
        return "train_model"


def validate_recommendations(**context):
    """Validate that recommendations were generated correctly."""
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

        # Check metadata
        last_updated = r.get("spark:recommendations:last_updated")
        user_count = r.get("spark:recommendations:user_count")
        product_count = r.get("spark:recommendations:product_count")

        print(f"Recommendations last updated: {last_updated}")
        print(f"User count: {user_count}")
        print(f"Product count: {product_count}")

        # Basic validation
        if user_count and int(user_count) > 0:
            # Check a sample recommendation
            sample_key = "spark:recommendations:user:1"
            sample_data = r.get(sample_key)
            if sample_data:
                recs = json.loads(sample_data)
                print(f"Sample recommendations for user 1: {len(recs)} products")
                context["ti"].xcom_push(key="validation_passed", value=True)
                return

        print("Warning: No recommendations found")
        context["ti"].xcom_push(key="validation_passed", value=False)

    except Exception as e:
        print(f"Error validating recommendations: {e}")
        context["ti"].xcom_push(key="validation_passed", value=False)


def update_pipeline_metrics(**context):
    """Update pipeline metrics and status in Redis."""
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

        validation_passed = context["ti"].xcom_pull(key="validation_passed", task_ids="validate_recommendations")

        metrics = {
            "last_run": datetime.now().isoformat(),
            "status": "success" if validation_passed else "warning",
            "validation_passed": str(validation_passed),
        }

        for key, value in metrics.items():
            r.set(f"airflow:recommendation_pipeline:{key}", value)

        print(f"Updated pipeline metrics: {metrics}")

    except Exception as e:
        print(f"Warning: Could not update Redis metrics: {e}")


# Spark submit command
SPARK_SUBMIT_CMD = """
docker exec {spark_host} /opt/spark/bin/spark-submit \
    --master spark://spark-master:7077 \
    --packages io.delta:delta-spark_2.12:3.0.0,org.apache.hadoop:hadoop-aws:3.3.4,com.amazonaws:aws-java-sdk-bundle:1.12.262 \
    --jars /jars/ojdbc11.jar \
    --conf spark.sql.extensions=io.delta.sql.DeltaSparkSessionExtension \
    --conf spark.sql.catalog.spark_catalog=org.apache.spark.sql.delta.catalog.DeltaCatalog \
    /scripts/recommendation_pipeline.py
""".format(spark_host=SPARK_MASTER_HOST)


with DAG(
    dag_id="florus_recommendation_pipeline",
    default_args=default_args,
    description="Daily ALS recommendation model training pipeline",
    schedule_interval="0 2 * * *",  # Daily at 2:00 AM
    start_date=datetime(2024, 1, 1),
    catchup=False,
    tags=["florus", "ml", "recommendations", "spark"],
) as dag:

    # Check if we have fresh data
    check_data = BranchPythonOperator(
        task_id="check_data_freshness",
        python_callable=check_data_freshness,
        provide_context=True,
    )

    # Skip training if no fresh data
    skip_training = EmptyOperator(
        task_id="skip_training",
    )

    # Train the ALS model
    train_model = BashOperator(
        task_id="train_model",
        bash_command=SPARK_SUBMIT_CMD,
        execution_timeout=timedelta(hours=2),
    )

    # Validate recommendations
    validate = PythonOperator(
        task_id="validate_recommendations",
        python_callable=validate_recommendations,
        provide_context=True,
        trigger_rule=TriggerRule.ONE_SUCCESS,
    )

    # Update metrics
    update_metrics = PythonOperator(
        task_id="update_metrics",
        python_callable=update_pipeline_metrics,
        provide_context=True,
        trigger_rule=TriggerRule.ALL_DONE,
    )

    # Define task dependencies
    check_data >> [skip_training, train_model]
    [skip_training, train_model] >> validate >> update_metrics
