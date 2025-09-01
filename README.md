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
- **Конфигурация**: Viper + переменные окружения
- **Миграции БД**: golang-migrate
- **Контейнеризация**: Docker, Docker Compose
- **Тестирование**: Testcontainers, testify

### Архитектура приложения

Приложение построено по слоистой (Clean/Onion) архитектуре:

1.  **Domain Layer** (`internal/domain/`):
    - `entities/`: Бизнес-сущности (Order, Delivery, Payment, Item) и их валидация
    - `errors.go`: Доменные ошибки
    - `repository/`: Интерфейсы репозиториев

2.  **Application Layer** (`internal/application/`):
    - `order_service.go`: Ядро бизнес-логики
    - `ports.go`: Интерфейсы сервисов
    - `errors.go`: Ошибки приложения

3.  **Infrastructure Layer** (`internal/infrastructure/`):
    - **Cache** (`cache/`): Реализация in-memory LRU кэша
    - **Database** (`database/`): Репозитории PostgreSQL с retry-механизмом
    - **Kafka** (`kafka/`): Consumer с DLQ и retry
    - **Config** (`config/`): Загрузка конфигурации и переменных окружения
    - **Logger** (`logger/`): Обертка для Zap логгера
    - **Migrations** (`migrations/`): Скрипты и runner для миграций БД

4.  **Interface Layer** (`internal/interface/`):
    - Инициализация зависимостей и жизненный цикл приложения

5.  **Bootstrap** (`internal/bootstrap/`):
    - Инициализация и зависимостей (логиger, кэш, БД, репозитории, сервисы, Kafka consumer, HTTP server)
    - Жизненный цикл приложения (запуск, graceful shutdown)

### Поток данных

1.  **Kafka Consumer** получает сообщение с заказом в JSON
2.  Сообщение декодируется и валидируется
3.  **OrderService** сохраняет заказ в БД с retry-механизмом
4.  В случае успеха, заказ помещается в **LRU кэш**
5.  REST API запросы сначала проверяют кэш, затем БД
6.  При старте приложения кэш автоматически восстанавливается из БД

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
│   │   ├── errors.go
│   │   ├── order_service_test.go
│   │   ├── order_service.go
│   │   └── ports.go
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
│   │   │   ├── errors.go
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
│   ├── main.go
│   └── postman
│       └── file model.json
└── web
    └── index.html
```

## Запуск проекта

### 1. Требования

- Docker и Docker Compose
- Go 1.24+

### 2. Клонирование и настройка

```bash
git clone github.com/Dmitrii-Khramtsov/orderservice
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

*Стандартный запуск:*
```bash
make docker-up
```
*Полная пересборка и запуск с очисткой:*
```bash
make docker-restart
```

*Запуск скрипта для генерации тестовых данных в Kafka:*
```bash
make script-up
```

### 4. Доступ к сервисам

- **OrderService API**: [http://localhost:8081](http://localhost:8081)
- **Простой UI**: [http://localhost:8081](http://localhost:8081) (откроется `web/index.html`)
- **PostgreSQL**: `localhost:5432` (user: `orders`, db: `orders`)
- **Kafka Broker**: `localhost:9092`

### 5. Остановка
*Стандартная остановка:*
```bash
make docker-down
```
*Полная остановка с очисткой:*
```bash
docker compose down --remove-orphans --volumes
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

*Пример запроса `GET /orders:`*
```bash
curl http://localhost:8081/orders
```

## Разработка и Тестирование

### Makefile команды
```bash
make build                      # собрать бинарный файл
make run                        # запустить приложение локально
make lint                       # запустить линтер
make test-unit                  # запустить unit-тесты
make test-integration           # запустить integration-тесты
make test-all                   # запустить все тесты
make test-coverage              # запустить тесты с покрытием
make test-coverage-unit         # unit-тесты с покрытием
make test-coverage-integration  # integration-тесты с покрытием
make docker-up                  # запустить весь стек
make docker-logs                # показать логи
make docker-down                # остановить стек
make docker-restart             # перезапустить с пересборкой
make script-up                  # запустить генератор тестовых данных
```

### Локальная разработка без Docker

1. *Запустите необходимые сервисы:*
```bash
docker compose up -d postgres kafka
```

2. *Настройте подключение к БД и Kafka в `config.yml:`*
```yml
database:
  dsn: "postgres://orders:orders@localhost:5432/orders?sslmode=disable"

kafka:
  brokers: ["localhost:9092"]
```

3. *Запустите приложение:*
```bash
make run
```

### Добавление миграций
Миграции находятся в `internal/infrastructure/database/migrations/.`
Для создания новой миграции используйте инструмент `migrate create`. Миграции применяются автоматически при старте приложения.

## Конфигурация

*Основная конфигурация (config.yml):*

```yml
cache:
  capacity: 10000
  get_all_limit: 1000
  restoration:
    timeout: 5m
    batch_size: 1000

database:
  dsn: "postgres://orders:orders@postgres:5432/orders?sslmode=disable"
  max_open_conns: 25
  max_idle_conns: 25
  conn_max_lifetime: "1h"
  statement_timeout: 30s
  idle_in_tx_session_timeout: 10s

kafka:
  brokers: ["kafka:9092"]
  topic: "orders"
  group_id: "orderservice"
  dlq_topic: "orders-dlq"
  max_retries: 5
  processing_time: 60s
  min_bytes: 10000
  max_bytes: 10000000
  max_wait: 1s
  commit_interval: 1s
  batch_timeout: 100ms
  batch_size: 1
  retry:
    initial_interval: 1s
    multiplier: 2
    max_interval: 30s
    max_elapsed_time: 5m
    randomization_factor: 0.5

server:
  port: "8081"

migrations:
  migrations_path: "/app/internal/infrastructure/database/migrations"
```

Параметры из `config.yml` могут быть переопределены переменными окружения (например, `POSTGRES_DSN`, `KAFKA_BROKERS`).

## Мониторинг и Логи

Логирование реализовано с использованием Zap. Уровень логирования и режим (development/production) задаются переменной окружения `LOG_MODE`.

В development режиме логи выводятся в консоль в удобочитаемом формате. В production режиме логи структурированы в JSON.