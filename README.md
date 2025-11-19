# Menu Parser Microservice

Микросервис для парсинга меню ресторанов из Google Таблиц с сохранением в MongoDB и обработкой через RabbitMQ.

## Технологии

- **Go 1.21+**
- **MongoDB 6+**
- **RabbitMQ**
- **Docker & Docker Compose**
- **Google Sheets API v4**

## Архитектура

Проект реализован с использованием **Clean Architecture** и лучших практик проектирования.

Подробное описание архитектуры см. в [ARCHITECTURE.md](./ARCHITECTURE.md)

## Структура проекта

```
.
├── cmd/                    # Точки входа приложения
│   ├── api/               # HTTP API сервер
│   └── worker/            # Queue worker
├── internal/               # Внутренние пакеты
│   ├── domain/            # Доменный слой (entities, интерфейсы)
│   │   ├── entity/        # Бизнес-сущности
│   │   ├── repository/    # Интерфейсы репозиториев
│   │   └── service/       # Интерфейсы внешних сервисов
│   ├── usecase/           # Бизнес-логика (Use Cases)
│   ├── repository/        # Реализации репозиториев
│   └── transport/         # Слой доставки (HTTP, Queue)
│       ├── http/          # HTTP handlers и роутинг
│       │   ├── dto/       # Data Transfer Objects
│       │   └── handler/   # HTTP handlers
│       └── queue/         # Queue consumers
├── pkg/                    # Переиспользуемые пакеты
│   ├── config/            # Конфигурация приложения
│   ├── database/          # MongoDB подключение
│   ├── parser/            # Парсер Google Sheets
│   ├── queue/             # RabbitMQ адаптеры
│   └── health/            # Health check сервис
├── deployment/             # Docker конфигурация
│   ├── docker-compose.yml
│   ├── Dockerfile.api
│   └── Dockerfile.worker
├── credentials/            # Google Sheets credentials (не в git)
├── Makefile               # Команды для сборки и запуска
├── go.mod
└── go.sum
```

## Быстрый старт

### 1. Подготовка Google Sheets API

1. Создайте проект в [Google Cloud Console](https://console.cloud.google.com/)
2. Включите Google Sheets API
3. Создайте Service Account и скачайте JSON ключ
4. Сохраните ключ в `credentials/credentials.json`

### 2. Запуск через Docker Compose

```bash
# Создайте директорию для credentials
mkdir -p credentials

# Поместите ваш Google Sheets credentials в credentials/credentials.json

# Запустите все сервисы (из корня проекта)
cd deployment
docker-compose up -d

# Или используйте Makefile команды из корня проекта
make docker-up

# Проверьте логи
docker-compose -f deployment/docker-compose.yml logs -f

# Или через Makefile
make docker-logs
```

### 3. Проверка работы

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Парсинг меню
curl -X POST http://localhost:8080/api/v1/parse \
  -H "Content-Type: application/json" \
  -d '{
    "spreadsheet_id": "YOUR_SPREADSHEET_ID",
    "restaurant_name": "Burger King"
  }'

# Проверка статуса задачи
curl http://localhost:8080/api/v1/parse/{task_id}

# Получение меню
curl http://localhost:8080/api/v1/menu/{menu_id}
```

## API Endpoints

### POST `/api/v1/parse`
Создает задачу на парсинг меню.

**Request:**
```json
{
  "spreadsheet_id": "1ABC...",
  "restaurant_name": "Burger King"
}
```

**Response:**
```json
{
  "task_id": "uuid",
  "status": "queued"
}
```

### GET `/api/v1/parse/{task_id}`
Получает статус задачи парсинга.

**Response:**
```json
{
  "task_id": "uuid",
  "status": "completed|processing|failed|queued",
  "menu_id": "ObjectId",
  "error": "текст ошибки",
  "created_at": "2025-11-14T10:00:00Z",
  "updated_at": "2025-11-14T10:05:00Z"
}
```

### GET `/api/v1/menu/{menu_id}`
Получает меню по ID.

### PATCH `/api/v1/products/{product_id}/status`
Обновляет статус продукта.

**Request:**
```json
{
  "status": "available|not_available|deleted",
  "reason": "out_of_stock"
}
```

### GET `/api/v1/health`
Проверка здоровья сервиса.

## Очереди сообщений

### menu-parsing
Очередь для задач парсинга меню. Сообщения обрабатываются worker'ом с retry механизмом (3 попытки с экспоненциальной задержкой).

### product-status
Очередь для событий изменения статусов продуктов.

### dlq (Dead Letter Queue)
Очередь для сообщений, которые не удалось обработать после всех попыток.

## Makefile команды

Проект включает Makefile для упрощения работы:

```bash
# Сборка
make build              # Собрать API и Worker
make build-api          # Собрать только API
make build-worker       # Собрать только Worker

# Запуск локально
make run-api            # Запустить API
make run-worker         # Запустить Worker

# Docker команды (требуют запуска из директории deployment/)
make docker-build       # Собрать Docker образы
make docker-up          # Запустить все сервисы
make docker-down        # Остановить все сервисы
make docker-logs        # Показать логи
make docker-restart     # Перезапустить сервисы

# Или напрямую через docker-compose из директории deployment/
cd deployment
docker-compose up -d
docker-compose logs -f

# Утилиты
make setup              # Создать директорию credentials
make deps               # Установить зависимости
make test               # Запустить тесты
make clean              # Очистить артефакты сборки
```

## Переменные окружения

Создайте файл `.env` в корне проекта или используйте переменные в `deployment/docker-compose.yml`:

```env
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=menu_parser
RABBITMQ_URI=amqp://guest:guest@rabbitmq:5672/
RABBITMQ_MENU_PARSING_QUEUE=menu-parsing
RABBITMQ_PRODUCT_STATUS_QUEUE=product-status
RABBITMQ_DLQ_QUEUE=dlq
GOOGLE_SHEETS_CREDENTIALS_PATH=/app/credentials/credentials.json
API_PORT=8080
API_HOST=0.0.0.0
```

## Разработка

### Локальный запуск (без Docker)

```bash
# Установите зависимости
go mod download
# Или используйте Makefile
make deps

# Запустите MongoDB и RabbitMQ через Docker
cd deployment
docker-compose up -d mongodb rabbitmq
cd ..

# Запустите API
go run cmd/api/main.go
# Или через Makefile
make run-api

# В другом терминале запустите Worker
go run cmd/worker/main.go
# Или через Makefile
make run-worker
```

## Структура данных MongoDB

### Коллекция `menus`
```javascript
{
  _id: ObjectId,
  name: String,
  restaurant_id: String,
  products: Array,
  attributes_groups: Array,
  attributes: Array,
  created_at: ISODate,
  updated_at: ISODate
}
```

### Коллекция `parsing_tasks`
```javascript
{
  _id: UUID,
  status: String, // queued, processing, completed, failed
  spreadsheet_id: String,
  restaurant_name: String,
  menu_id: ObjectId,
  error_message: String,
  retry_count: Number,
  created_at: ISODate,
  updated_at: ISODate
}
```

### Коллекция `product_status_audit`
```javascript
{
  _id: ObjectId,
  product_id: String,
  event_type: String,
  old_status: String,
  new_status: String,
  reason: String,
  user_id: String,
  timestamp: ISODate
}
```

## Особенности реализации

- ✅ Graceful shutdown для всех сервисов
- ✅ Retry механизм с экспоненциальной задержкой
- ✅ Dead Letter Queue для проблемных сообщений
- ✅ Health checks для всех сервисов
- ✅ Connection pooling для MongoDB
- ✅ Таймауты для всех операций
- ✅ Multi-stage Dockerfile для минимизации размера
- ✅ Индексы в MongoDB для оптимизации запросов

## Мониторинг

- RabbitMQ Management UI: http://localhost:15672 (guest/guest)
- MongoDB: mongodb://localhost:27017
- API Health Check: http://localhost:8080/api/v1/health

## Дополнительная информация

- Docker конфигурация находится в директории `deployment/`
- Для запуска через docker-compose используйте: `cd deployment && docker-compose up -d`
- Или используйте Makefile команды из корня проекта

