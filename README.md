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






before any codegen:
``` bash
$ git clone https://github.com/googleapis/googleapis.git third_party/googleapis
```

check ur output:
```bash
$ tree -L 1
.
├── README.md
├── antifraud
├── chosen_theme.md
├── processing_core
├── quest.md
└── third_party

# or
$ ls
README.md  antifraud  chosen_theme.md  processing_core  quest.md  third_party
```

vscode settings
```json
// settings.json
{
  "go.testEnvVars": {
    "TEST_DB_CONNSTR": "postgres://proc_core_user:proc_core_pwd@localhost:5433/proc_core_db?sslmode=disable"
  }
}

// launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Go: Test Package",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}"
    }
  ]
}
```
