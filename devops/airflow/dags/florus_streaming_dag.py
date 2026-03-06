"""
Florus Streaming DAG
Monitors and manages the Kafka to Delta Lake streaming job.
Checks every 10 minutes if the streaming job is running and restarts if needed.
"""

import os
from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator, BranchPythonOperator
from airflow.operators.bash import BashOperator
from airflow.operators.empty import EmptyOperator
from airflow.utils.trigger_rule import TriggerRule
import requests
import redis

# Environment variables
SPARK_MASTER_HOST = os.getenv("SPARK_MASTER_HOST", "florus-spark-master")
REDIS_HOST = os.getenv("REDIS_HOST", "florus-redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))

default_args = {
    "owner": "florus",
    "depends_on_past": False,
    "email_on_failure": False,
    "email_on_retry": False,
    "retries": 1,
    "retry_delay": timedelta(minutes=2),
}


def check_streaming_job_status(**context):
    """Check if the Kafka to Delta streaming job is running."""
    try:
        # Check Spark applications via REST API
        response = requests.get(
            f"http://{SPARK_MASTER_HOST}:8080/json/",
            timeout=10
        )
        if response.status_code == 200:
            data = response.json()
            active_apps = data.get("activeapps", [])

            # Check if Kafka-to-Delta app is running
            for app in active_apps:
                if "Kafka-to-Delta" in app.get("name", "") or "Florus-Kafka" in app.get("name", ""):
                    context["ti"].xcom_push(key="streaming_status", value="running")
                    print(f"Streaming job is running: {app['name']}")
                    return "streaming_running"

            print("Streaming job not found in active applications")
            context["ti"].xcom_push(key="streaming_status", value="not_running")
            return "restart_streaming"

    except Exception as e:
        print(f"Error checking streaming status: {e}")
        context["ti"].xcom_push(key="streaming_status", value="error")
        return "restart_streaming"


def restart_streaming_job(**context):
    """Restart the Kafka to Delta streaming job."""
    import subprocess

    # Submit the streaming job via docker exec
    cmd = [
        "docker", "exec", SPARK_MASTER_HOST,
        "/opt/spark/bin/spark-submit",
        "--master", "spark://spark-master:7077",
        "--packages", "org.apache.spark:spark-sql-kafka-0-10_2.12:3.5.0,io.delta:delta-spark_2.12:3.0.0,org.apache.hadoop:hadoop-aws:3.3.4,com.amazonaws:aws-java-sdk-bundle:1.12.262",
        "--conf", "spark.driver.extraJavaOptions=-Djava.security.manager=allow",
        "/scripts/kafka_to_delta.py"
    ]

    print(f"Submitting streaming job: {' '.join(cmd)}")

    try:
        # Run in background (streaming job)
        process = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            start_new_session=True
        )
        print(f"Streaming job submitted with PID: {process.pid}")
        context["ti"].xcom_push(key="streaming_pid", value=process.pid)
        return True
    except Exception as e:
        print(f"Failed to submit streaming job: {e}")
        raise


def update_monitoring_metrics(**context):
    """Update monitoring metrics in Redis."""
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
        status = context["ti"].xcom_pull(key="streaming_status", task_ids="check_streaming_status")

        r.set("airflow:streaming:last_check", datetime.now().isoformat())
        r.set("airflow:streaming:status", status or "unknown")

        print(f"Updated monitoring metrics: status={status}")
    except Exception as e:
        print(f"Warning: Could not update Redis metrics: {e}")


with DAG(
    dag_id="florus_streaming_monitor",
    default_args=default_args,
    description="Monitor and manage Kafka to Delta Lake streaming job",
    schedule_interval="*/10 * * * *",  # Every 10 minutes
    start_date=datetime(2024, 1, 1),
    catchup=False,
    tags=["florus", "streaming", "spark", "kafka"],
) as dag:

    # Check if streaming job is running
    check_streaming = BranchPythonOperator(
        task_id="check_streaming_status",
        python_callable=check_streaming_job_status,
        provide_context=True,
    )

    # If running, just log and continue
    streaming_running = EmptyOperator(
        task_id="streaming_running",
    )

    # If not running, restart it
    restart_streaming = PythonOperator(
        task_id="restart_streaming",
        python_callable=restart_streaming_job,
        provide_context=True,
    )

    # Update metrics in Redis
    update_metrics = PythonOperator(
        task_id="update_metrics",
        python_callable=update_monitoring_metrics,
        provide_context=True,
        trigger_rule=TriggerRule.ONE_SUCCESS,
    )

    # Define task dependencies
    check_streaming >> [streaming_running, restart_streaming]
    [streaming_running, restart_streaming] >> update_metrics
