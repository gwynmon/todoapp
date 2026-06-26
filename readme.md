# TodoApp

Учебный REST-сервис для управления задачами и заметками. Реализован как монорепо с тремя микросервисами.

При разработке опирался на этот шаблон архитектуры:
https://github.com/evrone/go-clean-template/blob/master/README_RU.md

## Архитектура

```
Client
  ├── :8081  auth-service    — регистрация, логин, JWT
  │                 └── PostgreSQL (users)
  ├── :8082  tasks-service   — задачи, заметки, уведомления
  │                 ├── PostgreSQL (tasks)
  │                 ├── MongoDB    (notes, notifications)
  │                 ├── Redis      (кеш задач)
  │                 └── RabbitMQ   (события задач)
  │
  └──:8083  notifier-service
                    ├── читает события из RabbitMQ
                    └── проверяет дедлайны по расписанию
```

Слои каждого сервиса:
- `internal/controller/restapi` — HTTP хендлеры
- `internal/usecases` — бизнес-логика
- `internal/repository` — работа с БД
- `internal/entity` — модели и интерфейсы
- `pkg/cache` — Redis кеш
- `pkg/broker` — RabbitMQ продюсер/консьюмер

## Как запустить

```bash
cp .env.example .env
docker compose up --build
```

| Сервис | URL |
|--------|-----|
| auth-service | http://localhost:8081 |
| tasks-service | http://localhost:8082 |
| notifier-service (healthz) | http://localhost:8083 |
| RabbitMQ Management UI | http://localhost:15672 (guest / guest) |

### Makefile команды

| Команда | Описание |
|---------|----------|
| `make init` | Создать `.env` из `.env.example` |
| `make up` | Поднять все сервисы в Docker |
| `make down` | Остановить все сервисы |
| `make logs` | Логи приложения |
| `make test` | Запустить тесты |
| `make migrate-up` | Применить миграции локально |
| `make migrate-down` | Откатить миграцию локально |
| `make migrate-create name=<name>` | Создать новую миграцию |

## Переменные окружения

| Переменная | Описание | Пример |
|------------|----------|--------|
| `AUTH_SERVER_PORT` | Порт auth-service | `:8081` |
| `TASKS_SERVER_PORT` | Порт tasks-service | `:8082` |
| `NOTIFIER_SERVER_PORT` | Порт notifier-service | `:8083` |
| `POSTGRES_DSN` | PostgreSQL для tasks | `postgres://user:pass@postgres:5432/tododb?sslmode=disable` |
| `AUTH_POSTGRES_*` | PostgreSQL для auth | см. `.env.example` |
| `MONGO_DSN` | MongoDB | `mongodb://mongo:27017/tododb` |
| `REDIS_DSN` | Redis | `redis://redis:6379/0` |
| `RABBITMQ_DSN` | RabbitMQ | `amqp://guest:guest@rabbitmq:5672/` |
| `JWT_SECRET` | Секрет для подписи JWT | `dev_secret_change_me_in_production` |
| `JWT_EXPIRE` | Время жизни токена | `24h` |
| `INTERNAL_SECRET` | Секрет для internal API | `change_me_in_production` |
| `TASKS_SERVICE_URL` | URL tasks-service для notifier | `http://tasks-service:8082` |
| `DEADLINE_CHECK_INTERVAL` | Интервал проверки дедлайнов | `1h` |

Для локального запуска без Docker замените хосты сервисов на `localhost`.

## API

### Аутентификация (auth-service :8081)

Все эндпоинты tasks-service требуют заголовок:
```
Authorization: Bearer <token>
```

```bash
# Регистрация
curl -X POST http://localhost:8081/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Ivan","email":"ivan@example.com","password":"secret"}'

# Логин → возвращает JWT токен
curl -X POST http://localhost:8081/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ivan@example.com","password":"secret"}'

export TOKEN="<токен из ответа>"
```

### Задачи (tasks-service :8082)

```bash
# Создать задачу
curl -X POST http://localhost:8082/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy milk","description":"2 litres","deadline":"2026-12-01T00:00:00Z"}'

# Список задач (с фильтром и пагинацией)
curl "http://localhost:8082/api/tasks?status=todo&limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# Получить задачу с заметками
curl http://localhost:8082/api/tasks/1 \
  -H "Authorization: Bearer $TOKEN"

# Обновить задачу
curl -X PUT http://localhost:8082/api/tasks/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"in_progress"}'

# Удалить задачу
curl -X DELETE http://localhost:8082/api/tasks/1 \
  -H "Authorization: Bearer $TOKEN"
```

### Заметки (tasks-service :8082)

```bash
# Добавить заметку к задаче
curl -X POST http://localhost:8082/api/tasks/1/notes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"Important note","meta":{"priority":"high"}}'

# Список заметок задачи
curl http://localhost:8082/api/tasks/1/notes \
  -H "Authorization: Bearer $TOKEN"

# Удалить заметку
curl -X DELETE http://localhost:8082/api/notes/<noteId> \
  -H "Authorization: Bearer $TOKEN"
```

### Уведомления (tasks-service :8082)

```bash
# Список уведомлений текущего пользователя
curl http://localhost:8082/api/notifications \
  -H "Authorization: Bearer $TOKEN"
```

### Health checks

```bash
curl http://localhost:8081/healthz  # auth liveness
curl http://localhost:8081/readyz   # auth readiness (postgres)

curl http://localhost:8082/healthz  # tasks liveness
curl http://localhost:8082/readyz   # tasks readiness (postgres, mongo, redis, rabbitmq)

curl http://localhost:8083/healthz  # notifier liveness
curl http://localhost:8083/readyz   # notifier readiness (mongo, rabbitmq)
```

## Кеширование

Кешируется список задач пользователя (`GET /api/tasks`) без фильтров — TTL 5 минут.
Кеш инвалидируется при создании, обновлении и удалении задачи.

## События RabbitMQ

Выбрал RabbitMQ из-за простоты настройки и достаточной функциональности для событийной модели.
Exchange: `task-events` (topic, durable). Очередь: `notifier-queue`.

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

## Notifier service

Читает события из RabbitMQ и сохраняет уведомления в MongoDB.
Дополнительно по расписанию (`DEADLINE_CHECK_INTERVAL`) запрашивает у tasks-service
задачи с дедлайном в ближайшие 24 часа и создаёт уведомления типа `task.deadline_approaching`.
Дедупликация: повторное уведомление о дедлайне одной задачи не создаётся.

Межсервисное взаимодействие защищено заголовком `X-Internal-Secret`.

## Библиотеки

| Библиотека | Почему |
|------------|--------|
| `sqlx` | Удобный маппинг строк БД в структуры без ORM |
| `pgx` | Быстрый нативный драйвер PostgreSQL |
| `mongo-driver` | Официальный драйвер MongoDB |
| `go-redis` | Идиоматичный Redis клиент для Go |
| `amqp091-go` | Официальная библиотека RabbitMQ для Go |
| `goose` | SQL миграции с версионированием |
| `bcrypt` | Хеширование паролей |
| `jwt` | JWT токены для авторизации |
| `validator` | Валидация входных данных |
| `env` | Загрузка конфига из переменных окружения |
| `uuid` | Request ID для логирования |
