# OrderService

Микросервис для управления заказами, обеспечивающий получение, хранение, кэширование и отображение данных о заказах. Сервис потребляет сообщения из Kafka, сохраняет их в PostgreSQL и предоставляет REST API для доступа к данным.

## Архитектура и Технологии

### Технологический стек
- **Язык**: Go 1.24.1
- **Фреймворк**: Chi (роутинг)
- **База данных**: PostgreSQL 15
- **Кэш**: In-memory LRU кэш
- **Брокер сообщений**: Apache Kafka
- **Логирование**: Zap
- **Конфигурация**: Viper
- **Миграции БД**: golang-migrate
- **Контейнеризация**: Docker, Docker Compose
- **Тестирование**: Testcontainers, testify

### Архитектура приложения

Приложение построено по слоистой (Clean/Onion) архитектуре:

1.  **Domain Layer** (`internal/domain/`):
    - `entities/`: Бизнес-сущности (Order, Delivery, Payment, Item) и их валидация.
    - `errors.go`: Доменные ошибки.

2.  **Application Layer** (`internal/application/`):
    - `order_service.go`: Ядро бизнес-логики. Orchestrates работу между кэшем и репозиторием.

3.  **Infrastructure Layer** (`internal/infrastructure/`):
    - **Cache** (`cache/`): Реализация in-memory LRU кэша.
    - **Database** (`database/`): Репозитории для работы с PostgreSQL (основной и с retry-механизмом).
    - **Kafka** (`kafka/`): Consumer для обработки сообщений и отправки в DLQ.
    - **Config** (`config/`): Загрузка конфигурации из YAML и переменных окружения.
    - **Logger** (`logger/`): Обертка для Zap логгера.
    - **Migrations** (`migrations/`): Скрипты и runner для миграций БД.

4.  **Interface Layer** (`internal/interface/`):
    - **HTTP** (`http/`): Обработчики (handlers) и роутер (router) для REST API.

5.  **Bootstrap** (`internal/bootstrap/`):
    - Инициализация и зависимостей (логиger, кэш, БД, репозитории, сервисы, Kafka consumer, HTTP server).
    - Жизненный цикл приложения (запуск, graceful shutdown).

### Поток данных

1.  **Kafka Consumer** получает сообщение с заказом в JSON.
2.  Сообщение декодируется и валидируется.
3.  **OrderService** пытается сохранить заказ в БД.
4.  В случае успеха, заказ помещается в **LRU кэш**.
5.  REST API запросы на получение заказа сначала проверяют кэш, а затем БД.
6.  При старте приложения кэш автоматически восстанавливается из БД.

## Структура проекта
```
orderservice
├── cmd
│   └── server
│       └── main.go
├── config.yml
├── docker-compose.yml
├── Dockerfile
├── Dockerfile.script
├── go.mod
├── go.sum
├── internal
│   ├── application
│   │   └── order_service.go
│   ├── bootstrap
│   │   ├── app.go
│   │   ├── factory
│   │   │   ├── cache.go
│   │   │   ├── database.go
│   │   │   ├── kafka.go
│   │   │   ├── logger.go
│   │   │   └── server.go
│   │   ├── lifecycle.go
│   │   └── migrator.go
│   ├── domain
│   │   ├── entities
│   │   │   ├── delivery.go
│   │   │   ├── errors.go
│   │   │   ├── item.go
│   │   │   ├── order.go
│   │   │   └── payment.go
│   │   ├── errors.go
│   │   └── repository
│   │       ├── cache.go
│   │       ├── errors.go
│   │       ├── event_consumer.go
│   │       ├── logger.go
│   │       └── order_repository.go
│   ├── infrastructure
│   │   ├── cache
│   │   │   ├── order_cache_lru_test.go
│   │   │   ├── order_cache_lru.go
│   │   │   └── restorer.go
│   │   ├── config
│   │   │   └── config.go
│   │   ├── database
│   │   │   ├── migrations
│   │   │   │   ├── 0001_create_orders_table.down.sql
│   │   │   │   ├── 0001_create_orders_table.up.sql
│   │   │   │   ├── 0002_add_indexes_to_orders_table.down.sql
│   │   │   │   └── 0002_add_indexes_to_orders_table.up.sql
│   │   │   ├── postgres_repository_integration_test.go
│   │   │   ├── postgres_repository.go
│   │   │   ├── retrying_repository_test.go
│   │   │   └── retrying_repository.go
│   │   ├── kafka
│   │   │   ├── consumer_interface.go
│   │   │   ├── consumer.go
│   │   │   └── errors.go
│   │   ├── logger
│   │   │   └── logger.go
│   │   └── migrations
│   │       └── migrations.go
│   └── interface
│       └── http
│           ├── errors.go
│           ├── handler
│           │   └── order_handler.go
│           └── router
│               └── router.go
├── Makefile
├── README.md
├── scripts
│   └── main.go
└── web
    └── index.html
```

## Запуск проекта

### 1. Требования

- Docker и Docker Compose
- Go 1.24+ (только для локальной разработки без Docker)

### 2. Клонирование и настройка

```bash
git clone <your-repo-url>
cd orderservice
```

*Создайте файл `.env` в корне проекта (пример содержимого уже есть в проекте):*

```bash
# Logger
LOG_MODE=development

# Database
POSTGRES_DSN=postgres://orders:orders@postgres:5432/orders?sslmode=disable

# Kafka
KAFKA_BROKERS=kafka:9092

# Scripts
NUMBER_OF_MESSAGES=10
DELAY_MS=500
```

### 3. Запуск через Docker Compose (рекомендуемый способ)

*Эта команда запустит весь стек: Zookeeper, Kafka, PostgreSQL и само приложение.*
```bash
# стандартный запуск
make docker-up

# или полная пересборка и запуск с очисткой
docker compose down --remove-orphans && docker compose up -d --build
```

*Чтобы также запустить скрипт для генерации тестовых данных в Kafka:*
```bash
make script-up
```

### 4. Доступ к сервисам

- **OrderService API**: [http://localhost:8081](http://localhost:8081)
- **Простой UI**: [http://localhost:8081](http://localhost:8081) (откроется `web/index.html`)
- **PostgreSQL**: `localhost:5432` (user: `orders`, db: `orders`)
- **Kafka Broker**: `localhost:9092`

### 5. Остановка
```bash
# стандартная остановка
make docker-down

# или полная остановка с очисткой
docker compose down --remove-orphans
```

## API Endpoints
- `POST /orders` – Создать/обновить заказ
- `GET /orders/{id}` – Получить заказ по ID
- `GET /orders` – Получить все заказы (из кэша)
- `DELETE /orders/{id}` – Удалить заказ по ID
- `DELETE /orders` – Очистить все заказы

*Пример запроса `GET /orders/{id}:`*
```bash
curl http://localhost:8081/orders/b563feb7b2b84b6test
```

## Разработка и Тестирование

### Makefile команды

*Проект использует Makefile для автоматизации частых задач:*
```bash
make build           # собрать бинарный файл
make run             # запустить приложение локально (требует запущенных БД и Kafka)
make lint            # запустить линтер (golangci-lint)
make test-unit       # запустить unit-тесты
make test-integration # запустить integration-тесты (требует Docker)
make test-all        # запустить все тесты
make test-coverage   # запустить тесты с генерацией отчета о покрытии
make docker-up       # запустить весь стек через Docker Compose
make docker-logs     # показать логи Docker Compose
make docker-down     # остановить стек Docker Compose
make docker-restart
```

### Локальная разработка без Docker

1. Убедитесь, что запущены PostgreSQL и Kafka (например, через `docker-compose up -d postgres kafka`).

2. Скопируйте `config.yml` и настройте DSN и брокеры на `localhost`.

3. Запустите приложение:
```bash
make run
```

### Добавление миграций
Миграции находятся в `internal/infrastructure/database/migrations/.`
Для создания новой миграции используйте инструмент `migrate create`. Файлы миграций применяются автоматически при старте приложения.

## Конфигурация

*Основная конфигурация находится в config.yml*

```yaml
server:
  port: "8081" # порт HTTP сервера

database:
  dsn: "postgres://orders:orders@postgres:5432/orders?sslmode=disable" # DSN для подключения к PostgreSQL
  max_open_conns: 25
  max_idle_conns: 25
  conn_max_lifetime: "1h"

kafka:
  brokers: ["kafka:9092"]
  topic: "orders"
  group_id: "orderservice"
  dlq_topic: "orders-dlq" # топик для мертвых писем
  max_retries: 5 # макс. количество попыток обработки сообщения

cache:
  capacity: 10000 # максимальная вместимость LRU кэша
  get_all_limit: 1000 # лимит для получения всех заказов
```

Параметры из `config.yml` могут быть переопределены переменными окружения (например, `POSTGRES_DSN`, `KAFKA_BROKERS`).

## Мониторинг и Логи

Логирование реализовано с использованием Zap. Уровень логирования и режим (development/production) задаются переменной окружения `LOG_MODE`.

В development режиме логи выводятся в консоль в удобочитаемом формате. В production режиме логи структурированы в JSON.