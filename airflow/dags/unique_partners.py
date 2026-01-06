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

# ID пользователя, для которого ищем партнеров
TARGET_USER_ID = "1f78b288-66b7-417c-bc25-5ffa3989f6ef"

def aggregate_hourly_partners(transaction_type, table_name, **kwargs):
    # ClickHouse client
    ch_client = Client(
        host=CLICKHOUSE_HOST,
        user=CLICKHOUSE_USER,
        password=CLICKHOUSE_PASSWORD,
        database=CLICKHOUSE_DB
    )

    # Берем окно последнего часа по data_interval
    execution_date = kwargs['data_interval_end']
    # округляем вниз до начала часа
    window_end = execution_date.replace(minute=0, second=0, microsecond=0)
    window_start = window_end - timedelta(hours=1)

    query = f"""
        SELECT DISTINCT receiver_id AS partner_id
        FROM transactions
        WHERE sender_id = '{TARGET_USER_ID}'
          AND transaction_type = '{transaction_type}'
          AND created_at >= '{window_start:%Y-%m-%d %H:%M:%S}'
          AND created_at < '{window_end:%Y-%m-%d %H:%M:%S}'
    """

    try:
        result = ch_client.execute(query)
    except Exception as e:
        print(f"[ERROR] ClickHouse query failed: {e}")
        raise

    if not result:
        print(f"No {transaction_type} transactions for user {TARGET_USER_ID} in the last hour.")
        return []

    # Cassandra client
    cluster = Cluster([CASSANDRA_HOST])
    session = cluster.connect(CASSANDRA_KEYSPACE)

    insert_stmt = session.prepare(f"""
        INSERT INTO {table_name} (user_id, window_start, partner_id)
        VALUES (?, ?, ?)
        USING TTL 172800
    """)

    try:
        for (partner_id_str,) in result:
            partner_id_uuid = uuid.UUID(partner_id_str)
            session.execute(insert_stmt, (uuid.UUID(TARGET_USER_ID), window_start, partner_id_uuid))
    except Exception as e:
        print(f"[ERROR] Cassandra insert failed: {e}")
        raise

    print(f"Inserted {len(result)} {transaction_type} partners into {table_name}.")


default_args = {
    "owner": "airflow",
    "depends_on_past": False,
    "retries": 1,
    "retry_delay": timedelta(minutes=5)
}

with DAG(
    dag_id="hourly_unique_partners_to_cassandra",
    default_args=default_args,
    description="Уникальные партнеры пользователя по СБП и внутренним переводам (часовые окна)",
    schedule_interval="0 * * * *",  # каждый час на начало часа
    start_date=datetime(2026, 1, 6),
    catchup=True,
    tags=["antifraud"],
) as dag:

    task_sbp = PythonOperator(
        task_id="aggregate_sbp_partners_hourly",
        python_callable=aggregate_hourly_partners,
        op_kwargs={"transaction_type": "SbpOutgoing", "table_name": "sbp_partners_30m"},
        provide_context=True
    )

    task_internal = PythonOperator(
        task_id="aggregate_internal_partners_hourly",
        python_callable=aggregate_hourly_partners,
        op_kwargs={"transaction_type": "Internal", "table_name": "internal_partners_30m"},
        provide_context=True
    )

    task_sbp >> task_internal
