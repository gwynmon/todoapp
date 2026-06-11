# TodoApp

Учебный REST-сервис для управления задачами и заметками.

## Архитектура
При создании опирался на вот этот шаблон: 
https://github.com/evrone/go-clean-template/blob/master/README_RU.md



Слои:
- `internal/controller/restapi` — HTTP хендлеры
- `internal/usecases` — бизнес-логика
- `internal/repository` — работа с БД
- `internal/entity` — модели и интерфейсы
- `pkg/cache` — Redis кеш
- `pkg/broker` — RabbitMQ продюсер

## Как запустить

```bash
cp .env.example .env
docker compose up --build
```

Сервис будет доступен на `http://localhost:8080`.

### Makefile команды

| Команда | Описание |
|---------|----------|
| `make init` | Создать `.env` из `.env.example` |
| `make up` | Поднять все сервисы в Docker |
| `make down` | Остановить все сервисы |
| `make logs` | Логи приложения |
| `make run` | Запустить локально без Docker |
| `make migrate-up` | Применить миграции локально |
| `make migrate-down` | Откатить миграцию локально |
| `make migrate-create name=<name>` | Создать новую миграцию |
| `make test` | Запустить тесты |

## Переменные окружения

| Переменная | Описание | Пример |
|------------|----------|--------|
| `SERVER_PORT` | Порт сервера | `:8080` |
| `POSTGRES_DSN` | Строка подключения к PostgreSQL | `postgres://user:pass@postgres:5432/tododb?sslmode=disable` |
| `MONGO_DSN` | Строка подключения к MongoDB | `mongodb://mongo:27017/tododb` |
| `REDIS_DSN` | Строка подключения к Redis | `redis://redis:6379/0` |
| `RABBITMQ_DSN` | Строка подключения к RabbitMQ | `amqp://guest:guest@rabbitmq:5672/` |
| `JWT_SECRET` | Секрет для подписи JWT | `secret` |
| `JWT_EXPIRE` | Время жизни токена | `24h` |

Для локального запуска без Docker замените хосты `postgres`, `mongo`, `redis`, `rabbitmq` на `localhost`.

## API

Все эндпоинты кроме `/api/register`, `/api/login`, `/healthz`, `/readyz` требуют заголовок:
Authorization: Bearer <token>

### Аутентификация

```bash
# Регистрация
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Ivan","email":"ivan@example.com","password":"secret"}'

# Логин
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ivan@example.com","password":"secret"}'
```

### Задачи

```bash
# Создать задачу
curl -X POST http://localhost:8080/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy milk","description":"2 litres","deadline":"2026-12-01T00:00:00Z"}'

# Список задач (с фильтром и пагинацией)
curl "http://localhost:8080/api/tasks?status=todo&limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# Получить задачу с заметками
curl http://localhost:8080/api/tasks/1 \
  -H "Authorization: Bearer $TOKEN"

# Обновить задачу
curl -X PUT http://localhost:8080/api/tasks/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"in_progress"}'

# Удалить задачу
curl -X DELETE http://localhost:8080/api/tasks/1 \
  -H "Authorization: Bearer $TOKEN"
```

### Заметки

```bash
# Добавить заметку
curl -X POST http://localhost:8080/api/tasks/1/notes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"Important note","meta":{"priority":"high"}}'

# Список заметок
curl http://localhost:8080/api/tasks/1/notes \
  -H "Authorization: Bearer $TOKEN"

# Удалить заметку
curl -X DELETE http://localhost:8080/api/notes/<noteId> \
  -H "Authorization: Bearer $TOKEN"
```

### Health checks

```bash
curl http://localhost:8080/healthz  # liveness
curl http://localhost:8080/readyz   # readiness
```

## Кеширование

Кешируется список задач пользователя (`GET /api/tasks`) без фильтров - TTL 5 минут.
Кеш инвалидируется при создании, обновлении и удалении задачи.

## События RabbitMQ

Я выбрал RabbitMQ из-за простоты настройки и достаточной функциональности для событийной модели приложения. 
Он позволяет публиковать события задач и легко расширяется консьюмерами на следующих этапах проекта.

Exchange: `task-events` (topic, durable).

| Routing key | Когда |
|-------------|-------|
| `task.created` | Создание задачи |
| `task.status_changed` | Изменение статуса |
| `task.deleted` | Удаление задачи |

Структура события:
```json
{
  "event_type": "task.created",
  "task_id": 1,
  "user_id": 6,
  "timestamp": "2026-06-08T10:00:00Z",
  "payload": {"title": "Buy milk", "status": "todo"}
}
```

Management UI: `http://localhost:15672` (guest / guest).

## Библиотеки

| Библиотека | Почему |
|------------|--------|
| `sqlx` | Удобный маппинг строк БД в структуры без ORM |
| `pgx` | Быстрый драйвер PostgreSQL |
| `mongo-driver` | Официальный драйвер MongoDB |
| `go-redis` | Идиоматичный Redis клиент для Go |
| `amqp091-go` | Официальная библиотека RabbitMQ для Go |
| `goose` | Простые SQL миграции с версионированием |
| `bcrypt` | Хеширование паролей |
| `jwt` | JWT токены для авторизации |
| `validator` | Валидация входных данных |
| `env` | Загрузка конфига из переменных окружения |