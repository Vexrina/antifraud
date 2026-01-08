# Antifraud kekw

## Введение

### Описание проекта

Это подобие антифрода.

- Так как весь антифрод -- внутрянка любого банка -> можно использовать любой канал связи внутри инфраструктуры "банка"
- Так как на разные авторизации у нас разный но малый таймаут (что-то вроде 500 мс) то никакой кафки быть не может.
- Рест можно (?) "околохакать" и ручками подергать, с grpc так не разгуляешься -- идем к нему.

Реализовано два сервиса -- ядро процессинга и сам по себе антифрод

Для ядра процессинга были реализованы 4 ручки:
- положить деньги на счет
- снять деньги со счета
- перевести по сбп (Сервис Быстрых Платежей)
- перевести по внутреннему

*Небольшие заметки на полях:*
- это ядро процессинга, т.е. физически перекладываем деньги из А в Б
- банк небольшой и ЦБ еще не заставил ввести лимитную историю
- мы **не подключены к кафке выпусков** на MVP -> если пользак дошел до нашей ручки "положить деньги на счет" -> Деньги положатся на счет
- в данной реализации интеграция с СБП и банкоматами опущена, считаем что за 20 мс справимся
- в нашем банке карточный процессинг просто маппит транзакцию и отдает ее в нас. Ну все.

### Стек технологий

<p>
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/go/go-original-wordmark.svg" title="Go" alt="Go" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/docker/docker-original.svg" title="Docker" alt="Docker" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/postgresql/postgresql-original.svg" title="PostgreSQL" alt="PostgreSQL" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/clickhouse/clickhouse-original.svg" title="ClickHouse" alt="ClickHouse" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/apachekafka/apachekafka-original.svg" title="Apache Kafka" alt="Apache Kafka" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/prometheus/prometheus-original.svg" title="Prometheus" alt="Prometheus" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/grafana/grafana-original.svg" title="Grafana" alt="Grafana" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/apacheairflow/apacheairflow-original.svg" title="Apache Airflow" alt="Apache Airflow" width="40" height="40"/>&nbsp;
    <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/cassandra/cassandra-original.svg" title="Apache Cassandra" alt="Apache Cassandra" width="40" height="40"/>&nbsp;
    <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/grpc/grpc-original.svg" title="grpc" alt="grpc"  width="40" height="40"/>&nbsp;
    <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/python/python-original-wordmark.svg" title="python" alt="python width="40" height="40"/>&nbsp;
          
          
</p>

## Запуск

В теории можно просто прописать 
```bash
$ docker-compose up -d
```
после чего зайти в клик и в cassandra и запустить ddl в конце readme

желательно перезапустить все контейнеры, чтобы быть увереным в работе .

для запуска DAGов в airflow -- Зайти в его UI, включить их триггеры.

чтобы начать отправлять транзакции, надо повозится с питоном
``` bash
$ python3 -m venv .venv/
$ source .venv/bin/activate # ну или че там в винде
$ pip install -r ./python_scrpits/req.txt
$ .venv/bin/python ./python_scripts/create_transactions.py 
```

## Основная часть

### Анализ предметной области

- Обоснуйте выбора архитектуры приложения

Почему grpc, почему не принимаем транзакции кафкой описано в пункте "введение".
Почему не принимаем транзакции .csv -- реализован онлайн антифрод, а не оффлайн.

Почему льем в клик, переливаем через airflow, кладем в кассандру:
1. клик хорошо подходит для сырых данных. Сюда бы зашел и ytsaurus/hdfs, но не хочется в них разбираться
2. потому что был выбор между ним и debesium. Бросилась монетка
3. cassandra Хорошо себя показывает в онлайн антифроде у меня в компании, поэтому скопипасчено оттуда


- Обзор существующих решений

Кроме антифрода в собственной компании нигде не видел как он реализован, но раскрывать его работу не могу [nda]

- Описание стека технологий для каждой части проекта

Click, Cassandra airflow описаны выше.

Кафка хорошо показывает себя для аутбокса, мы должны уметь хранить в кафке N дней, легко перелить данные повторно, поддерживать много консьюмер групп -> не RabbitMQ

Golang а не какой-то еще ЯП -- go основной язык

Postgre как основная таблица процессинга -- шардируется, когда у нас будет 2кк пользаков, да и в целом стандарт.

Grafana, Prometheus -- база

Python для тестов -- база

### Проектирование

#### Архитектура приложения

- Предоставьте **расчеты** нагрузки и требуемых **ресурсов** (память/процессор) для вашей системы.

покажу на защите, но в рамках 300 rps в docker desktop "естстя" 150% из 800% cpu и 7гигов из 16

- Приведите UML-диаграммы с **разьяснениями**:
  - [Юзкейсы](https://plantuml.com/ru-dark/use-case-diagram)
  - [Последовательности выполнения](https://plantuml.com/ru-dark/sequence-diagram)
  - [Блок-схемы](https://plantuml.com/ru-dark/activity-diagram-beta)
  - И другие по необходимости

#### Схемы баз данных

[DBML-схемы](https://dbml.dbdiagram.io/home/) **всех** таблиц во **всех** используемых базах данных

Схемы [тут](schemas/)

#### Описание API

Опишите взаимодействие с приложением по REST API.

GRPC. Представлено в архитектуре в сиквенсах

### Тестирование

- Опишите как именно проводили тестирование

Написал юниты и интеграционные юниты. Написал [скрипт](python_scripts/create_transactions.py) для "дудоса" сервисов

- Предоставьте **отчет о покрытии** модульными и интеграционными тестами (можно использовать готовые инструменты для вашего ЯП)

miss 

- Предоставьте файлы с тестовыми данными (.csv, .json и др.)

их нет, все делается через [скрипт](python_scripts/create_transactions.py)


- Предоставьте коллекцию [Postman](https://www.postman.com/) или [Insomnia](https://insomnia.rest/) с примерами запросов и ответов для демонстрации

[пупупу](https://github.com/postmanlabs/postman-app-support/issues/11252)

## Заключение

### Краткие выводы

прямо сейчас скрипт дудоса работает, в среднем на все ручки приходит по 40-45 рпс ~= 140 рпса всего. Не дотягивает до прода, конечно, но быстрее дудосить питон не умеет.

По юзажу: вся система пожирает 148%/800% cpu && 7.4GB RAM

На 95ом квантиле процессинг продолжает отвечать в среднем за 25мс. У меня, конечно, правил маловато, но при расширении антифрод аналитикам придется пересмотреть ценность тех или иных правил.

### Результаты

запустите скрипт дудоса, подождите полчаса на первую генерацию и зайдите в графану, все четко

### Перспективы развития

Больше дагов сделать, написать отдельный сервис под фичастор, упаковать все под одну единую либу, чтобы не надо было писать один и тот же код десять раз.

## приложение

#### clickhouse ddls:
```sql
create table default.kafka_transactions_raw
(
  id              String,
  transactionId   String,
  createdAt       String,
  amount          String,
  currency        String,
  merchant        String,
  country         String,
  senderId        String,
  receiverId      Nullable(String),
  receiverBic     Nullable(String),
  atmId           Nullable(String),
  transactionType String,
  revision        Nullable(String)
)
  engine = Kafka SETTINGS kafka_broker_list = 'redpanda:9092', kafka_topic_list = 'best_bank_transactions', kafka_group_name = 'ch_group', kafka_format = 'JSONEachRow';


create table default.transactions
(
  id               UInt64,
  transaction_id   String,
  created_at       DateTime64(3),
  amount           Int64,
  currency         String,
  merchant         String,
  country          String,
  sender_id        Nullable(String),
  receiver_id      Nullable(String),
  receiver_bic     Nullable(String),
  atm_id           Nullable(String),
  transaction_type Nullable(String),
  revision         Nullable(Int64) default 0
)
  engine = MergeTree ORDER BY id
    SETTINGS index_granularity = 8192;

CREATE MATERIALIZED VIEW default.mv_kafka_transactions
  TO default.transactions
  (
    `id` UInt64,
    `transaction_id` String,
    `created_at` DateTime,
    `amount` Int64,
    `currency` String,
    `merchant` String,
    `country` String,
    `sender_id` Nullable(String),
    `receiver_id` Nullable(String),
    `receiver_bic` Nullable(String),
    `atm_id` Nullable(String),
    `transaction_type` Nullable(String),
    `revision` Int64
  )
AS
SELECT 
  toUInt64OrZero(id)                              AS id,
  transactionId                                   AS transaction_id,
  parseDateTimeBestEffort(createdAt)              AS created_at,
  toInt64OrZero(amount)                           AS amount,
  currency,
  merchant,
  country,
  if(senderId = '', NULL, senderId)               AS sender_id,
  if(receiverId = '', NULL, receiverId)           AS receiver_id,
  if(receiverBic = '', NULL, receiverBic)         AS receiver_bic,
  if(atmId = '', NULL, atmId)                     AS atm_id,
  if(transactionType = '', NULL, transactionType) AS transaction_type,
  toInt64OrZero(coalesce(revision, '0'))          AS revision
FROM default.kafka_transactions_raw;

```


### cassandra ddls:
```sql
CREATE KEYSPACE antifraud
WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};

create table antifraud.cash_out_30m
(
  user_id       uuid,
  window_start  timestamp,
  total_cashout decimal,
  primary key (user_id, window_start)
);



CREATE TABLE antifraud.sbp_partners_30m (
user_id uuid,
window_start timestamp,
partner_id uuid,
PRIMARY KEY (user_id, window_start, partner_id)
) WITH CLUSTERING ORDER BY (window_start DESC, partner_id ASC)
    AND default_time_to_live = 172800;

CREATE TABLE antifraud.internal_partners_30m (
user_id uuid,
window_start timestamp,
partner_id uuid,
PRIMARY KEY (user_id, window_start, partner_id)
) WITH CLUSTERING ORDER BY (window_start DESC, partner_id ASC)
    AND default_time_to_live = 172800;

CREATE TABLE antifraud.spent_3h (
  user_id uuid,               -- целевой пользователь X
  window_start timestamp,     -- начало 3-часового окна
  total_spent decimal,        -- сумма всех расходов (кроме CashIn)
  PRIMARY KEY ((user_id), window_start)
) WITH CLUSTERING ORDER BY (window_start ASC)
  AND default_time_to_live = 172800; -- TTL 2 дня


```