# Архитектура проекта

Проект реализован с использованием **Clean Architecture** и лучших практик проектирования.

## Структура проекта

```
menu-parser/
├── cmd/                    # Точки входа приложения
│   ├── api/               # HTTP API сервер
│   └── worker/             # Queue worker
├── internal/               # Внутренние пакеты (не экспортируются)
│   ├── domain/            # Доменный слой (бизнес-логика)
│   │   ├── entity/        # Сущности домена
│   │   ├── repository/    # Интерфейсы репозиториев
│   │   └── service/       # Интерфейсы внешних сервисов
│   ├── usecase/           # Слой бизнес-логики (Use Cases)
│   ├── repository/        # Реализации репозиториев
│   └── delivery/          # Слой доставки (внешние интерфейсы)
│       ├── http/          # HTTP handlers
│       └── queue/         # Queue consumers
└── pkg/                    # Переиспользуемые пакеты
    ├── database/          # MongoDB подключение
    ├── parser/            # Google Sheets парсер
    ├── queue/             # RabbitMQ адаптеры
    └── health/            # Health check сервис
```

## Принципы Clean Architecture

### 1. Domain Layer (Доменный слой)
**Расположение:** `internal/domain/`

Содержит:
- **Entities** - бизнес-сущности (Menu, Product, ParsingTask, etc.)
- **Repository Interfaces** - интерфейсы для работы с данными
- **Service Interfaces** - интерфейсы для внешних сервисов

**Принципы:**
- Не зависит от других слоев
- Содержит только бизнес-логику и правила
- Интерфейсы определяют контракты, а не реализации

### 2. Use Case Layer (Слой бизнес-логики)
**Расположение:** `internal/usecase/`

Содержит:
- `MenuUseCase` - бизнес-логика работы с меню
- `ProductUseCase` - бизнес-логика работы с продуктами
- `HealthUseCase` - проверка здоровья сервисов

**Принципы:**
- Реализует бизнес-правила и сценарии использования
- Зависит только от domain layer
- Использует интерфейсы из domain layer

### 3. Repository Layer (Слой данных)
**Расположение:** `internal/repository/`

Содержит реализации интерфейсов из domain layer:
- `MenuRepository` - работа с меню в MongoDB
- `TaskRepository` - работа с задачами парсинга
- `AuditRepository` - работа с аудит-логами

**Принципы:**
- Реализует интерфейсы из domain layer
- Инкапсулирует логику работы с БД
- Может быть легко заменен на другую реализацию

### 4. Delivery Layer (Слой доставки)
**Расположение:** `internal/delivery/`

Содержит:
- **HTTP Handlers** - обработка HTTP запросов
- **Queue Consumers** - обработка сообщений из очереди
- **DTOs** - Data Transfer Objects для API

**Принципы:**
- Зависит от usecase layer
- Преобразует внешние форматы в доменные сущности
- Не содержит бизнес-логики

### 5. Infrastructure Layer (Инфраструктурный слой)
**Расположение:** `pkg/`

Содержит:
- `database` - подключение к MongoDB
- `parser` - парсер Google Sheets
- `queue` - адаптеры для RabbitMQ
- `health` - health check сервисы

**Принципы:**
- Реализует технические детали
- Может использоваться любым слоем
- Изолирует внешние зависимости

## Dependency Injection

Все зависимости инжектируются через конструкторы в `cmd/api/main.go` и `cmd/worker/main.go`:

```go
// 1. Инициализация инфраструктуры
db := database.NewMongoDB(cfg)
rabbitmq := queue.NewRabbitMQ(cfg)

// 2. Инициализация репозиториев
menuRepo := repository.NewMenuRepository(db)
taskRepo := repository.NewTaskRepository(db)

// 3. Инициализация сервисов
parser := parser.NewSheetsParser(cfg)
queuePublisher := queue.NewQueuePublisher(rabbitmq)

// 4. Инициализация use cases
menuUseCase := usecase.NewMenuUseCase(menuRepo, taskRepo, parser, queuePublisher)

// 5. Инициализация handlers
router := http.SetupRouter(menuUseCase, productUseCase, healthUseCase)
```

## Преимущества архитектуры

1. **Разделение ответственности** - каждый слой имеет четко определенную роль
2. **Тестируемость** - легко мокировать зависимости через интерфейсы
3. **Гибкость** - можно заменить реализацию без изменения бизнес-логики
4. **Независимость** - domain layer не зависит от внешних библиотек
5. **Масштабируемость** - легко добавлять новые use cases и handlers

## Поток данных

### HTTP Request Flow:
```
HTTP Request 
  → HTTP Handler (delivery/http)
  → Use Case (usecase)
  → Repository (repository)
  → Database (pkg/database)
  → Response
```

### Queue Message Flow:
```
Queue Message
  → Queue Consumer (delivery/queue)
  → Use Case (usecase)
  → Repository (repository)
  → Database (pkg/database)
```

## Следование SOLID принципам

- **S**ingle Responsibility - каждый класс/структура имеет одну ответственность
- **O**pen/Closed - открыт для расширения, закрыт для модификации
- **L**iskov Substitution - интерфейсы могут быть заменены реализациями
- **I**nterface Segregation - маленькие, специфичные интерфейсы
- **D**ependency Inversion - зависимости направлены к абстракциям, а не к конкретным реализациям



