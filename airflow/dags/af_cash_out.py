from datetime import datetime, timedelta
from airflow import DAG
import uuid
import pendulum
from airflow.operators.python import PythonOperator
from clickhouse_driver import Client
from cassandra.cluster import Cluster

# ClickHouse
CLICKHOUSE_HOST = "clickhouse-server"
CLICKHOUSE_USER = "myuser"
CLICKHOUSE_PASSWORD = "mypass"
CLICKHOUSE_DB = "default"

# Cassandra
CASSANDRA_HOST = "cassandra"
CASSANDRA_KEYSPACE = "antifraud"

def aggregate_and_store(**kwargs):
    # ClickHouse client
    ch_client = Client(
        host=CLICKHOUSE_HOST,
        user=CLICKHOUSE_USER,
        password=CLICKHOUSE_PASSWORD,
        database=CLICKHOUSE_DB
    )

    # Берем интервал, который DAG планирует обработать
    window_start = kwargs['data_interval_start']  # уже datetime
    window_end = kwargs['data_interval_end']

    query = f"""
        SELECT sender_id, sum(amount) AS total_amount
        FROM transactions
        WHERE transaction_type = 'CashOut'
          AND created_at >= '{window_start:%Y-%m-%d %H:%M:%S}'
          AND created_at < '{window_end:%Y-%m-%d %H:%M:%S}'
        GROUP BY sender_id
    """

    try:
        result = ch_client.execute(query)
    except Exception as e:
        print(f"[ERROR] ClickHouse query failed: {e}")
        raise

    if not result:
        print("Нет cashout транзакций за интервал.")
        return

    # Cassandra client
    cluster = Cluster([CASSANDRA_HOST])
    session = cluster.connect(CASSANDRA_KEYSPACE)

    insert_stmt = session.prepare("""
        INSERT INTO cash_out_30m (user_id, window_start, total_cashout)
        VALUES (?, ?, ?)
        USING TTL 172800
    """)

    for sender_id, total_amount in result:
        try:
            sender_uuid = uuid.UUID(sender_id)  # конвертация строки в UUID
            session.execute(insert_stmt, (sender_uuid, window_start, total_amount))
        except Exception as e:
            print(f"[ERROR] Cassandra insert failed for sender_id={sender_id}: {e}")
            raise

    print(f"Сохранили {len(result)} записей в Cassandra за {window_start} - {window_end}.")

default_args = {
    "owner": "airflow",
    "depends_on_past": False,
    "retries": 1,
    "retry_delay": timedelta(minutes=5)
}

with DAG(
    dag_id="cashout_to_cassandra",
    default_args=default_args,
    description="Агрегируем cashout за последние 30 минут и пишем в Cassandra",
    schedule_interval="*/30 * * * *",  # каждые 30 минут
    start_date=pendulum.datetime(2026, 1, 6, 0, 0, 0, tz="UTC"),
    catchup=True,
    tags=["antifraud"],
) as dag:

    task_aggregate_and_store = PythonOperator(
        task_id="aggregate_and_store",
        python_callable=aggregate_and_store,
        provide_context=True
    )