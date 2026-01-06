# Тема вашего проекта

## Введение

### Описание проекта

Приведите вводную информацию в проект и раскройте его тему.
Опишите объект и предмет исследования/разработки.

### Стек технологий

<p>
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/go/go-original-wordmark.svg" title="Go" alt="Go" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/docker/docker-original.svg" title="Docker" alt="Docker" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/kubernetes/kubernetes-original.svg" title="Kubernetes" alt="Kubernetes" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/postgresql/postgresql-original.svg" title="PostgreSQL" alt="PostgreSQL" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/clickhouse/clickhouse-original.svg" title="ClickHouse" alt="ClickHouse" width="40" height="40"/>&nbsp;
   <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/apachekafka/apachekafka-original.svg" title="Apache Kafka" alt="Apache Kafka" width="40" height="40"/>&nbsp;
    <img src="https://assets.streamlinehq.com/image/private/w_300,h_300,ar_1/f_auto/v1/icons/1/apache-superset-icon-cyc19fiufldpekdt6c7jg.png/apache-superset-icon-80ygkwbe76iyhvftejjahm.png?_a=DATAg1AAZAA0" title="Apache Superset" alt="Apache Superset" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/prometheus/prometheus-original.svg" title="Prometheus" alt="Prometheus" width="40" height="40"/>&nbsp;
    <img src="https://raw.githubusercontent.com/devicons/devicon/54cfe13ac10eaa1ef817a343ab0a9437eb3c2e08/icons/grafana/grafana-original.svg" title="Grafana" alt="Grafana" width="40" height="40"/>&nbsp;
</p>

## Запуск

Опишите как именно запускать проект, на каких хостах "поднимаются" компоненты для взаимодействия и другую информацию для работы с системой.

## Основная часть

### Анализ предметной области

- Обоснуйте выбора архитектуры приложения
- Обзор существующих решений
- Описание стека технологий для каждой части проекта

### Проектирование

#### Архитектура приложения

- Предоставьте **расчеты** нагрузки и требуемых **ресурсов** (память/процессор) для вашей системы.
- Приведите UML-диаграммы с **разьяснениями**:
  - [Юзкейсы](https://plantuml.com/ru-dark/use-case-diagram)
  - [Последовательности выполнения](https://plantuml.com/ru-dark/sequence-diagram)
  - [Блок-схемы](https://plantuml.com/ru-dark/activity-diagram-beta)
  - И другие по необходимости

#### Схемы баз данных

[DBML-схемы](https://dbml.dbdiagram.io/home/) **всех** таблиц во **всех** используемых базах данных

#### Описание API

Опишите взаимодействие с приложением по REST API.

### Тестирование

- Опишите как именно проводили тестирование
- Предоставьте **отчет о покрытии** модульными и интеграционными тестами (можно использовать готовые инструменты для вашего ЯП)
- Предоставьте файлы с тестовыми данными (.csv, .json и др.)
- Предоставьте коллекцию [Postman](https://www.postman.com/)
или [Insomnia](https://insomnia.rest/) с примерами запросов и ответов для демонстрации

## Заключение

### Краткие выводы

Подведите итоги выполненной работы. Оцените эффективность вашего проекта для разных уровней нагрузки

### Результаты

Приведите результаты, соответствующие задаче и цели, которые вы получили в ходе выполнения проекта.

### Перспективы развития

Подсветите моменты, которые, по вашему мнению, положительно повлияют на оптимизацию, производительность и безопасность.


clickhouse ddls:
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