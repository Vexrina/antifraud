from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator
from clickhouse_driver import Client
from cassandra.cluster import Cluster
import uuid

# ClickHouse
CLICKHOUSE_HOST = "clickhouse-server"
CLICKHOUSE_USER = "myuser"
CLICKHOUSE_PASSWORD = "mypass"
CLICKHOUSE_DB = "default"

# Cassandra
CASSANDRA_HOST = "cassandra"
CASSANDRA_KEYSPACE = "antifraud"

# Пользователь, для которого считаем расходы
TARGET_USER_ID = "1f78b288-66b7-417c-bc25-5ffa3989f6ef"

def aggregate_3h_spent(**kwargs):
    # ClickHouse client
    ch_client = Client(
        host=CLICKHOUSE_HOST,
        user=CLICKHOUSE_USER,
        password=CLICKHOUSE_PASSWORD,
        database=CLICKHOUSE_DB
    )

    # Берем окно последних 3 часов по data_interval_end
    execution_date = kwargs['data_interval_end']
    # округляем до начала часа
    window_end = execution_date.replace(minute=0, second=0, microsecond=0)
    window_start = window_end - timedelta(hours=3)

    query = f"""
        SELECT sum(amount) AS total_spent
        FROM transactions
        WHERE sender_id = '{TARGET_USER_ID}'
          AND transaction_type != 'CashIn'
          AND created_at >= '{window_start:%Y-%m-%d %H:%M:%S}'
          AND created_at < '{window_end:%Y-%m-%d %H:%M:%S}'
    """

    try:
        result = ch_client.execute(query)
        total_spent = result[0][0] if result and result[0][0] is not None else 0
    except Exception as e:
        print(f"[ERROR] ClickHouse query failed: {e}")
        raise

    # Cassandra client
    cluster = Cluster([CASSANDRA_HOST])
    session = cluster.connect(CASSANDRA_KEYSPACE)

    insert_stmt = session.prepare("""
        INSERT INTO spent_3h (user_id, window_start, total_spent)
        VALUES (?, ?, ?)
        USING TTL 172800
    """)

    try:
        session.execute(insert_stmt, (uuid.UUID(TARGET_USER_ID), window_start, total_spent))
    except Exception as e:
        print(f"[ERROR] Cassandra insert failed: {e}")
        raise

    print(f"Inserted total_spent={total_spent} for user {TARGET_USER_ID} from {window_start} to {window_end}")


default_args = {
    "owner": "airflow",
    "depends_on_past": False,
    "retries": 1,
    "retry_delay": timedelta(minutes=5)
}

with DAG(
    dag_id="user_spent_last_3h",
    default_args=default_args,
    description="Агрегируем расходы пользователя за последние 3 часа (все кроме CashIn)",
    schedule_interval="0 * * * *",  # каждый час
    start_date=datetime(2026, 1, 7),
    catchup=True,
    tags=["antifraud"],
) as dag:

    task_aggregate_3h_spent = PythonOperator(
        task_id="aggregate_3h_spent",
        python_callable=aggregate_3h_spent,
        provide_context=True
    )
