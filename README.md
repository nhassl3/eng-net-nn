# IpBuild Backend

> REST API рекрутинговой платформы: вакансии, отклики кандидатов с резюме и заявки на индивидуальные планы развития.

Построен на **Go + Gin** с чистой слоистой архитектурой (`transport → service → repository → db`). Хранение данных — **PostgreSQL** (через [sqlc](https://sqlc.dev)), сессии и кэш — **Redis**, файлы резюме — **MinIO** (S3-совместимое объектное хранилище), аутентификация — **PASETO**-токены с хешированием паролей **Argon2id**.

---

## Содержание

- [Возможности](#возможности)
- [Технологии](#технологии)
- [Архитектура](#архитектура)
- [Быстрый старт](#быстрый-старт)
- [Конфигурация](#конфигурация)
- [API](#api)
- [Команды Makefile](#команды-makefile)
- [Структура проекта](#структура-проекта)

---

## Возможности

- 🔐 **Аутентификация и авторизация** — регистрация, вход, refresh- и access-токены на PASETO, выход с blacklist-токенов в Redis, роли (`user` / `admin`).
- 🔑 **Безопасное хранение паролей** — Argon2id.
- 💼 **Вакансии** — публичный листинг и просмотр, CRUD для администраторов.
- 📨 **Отклики кандидатов** — приём анкеты вместе с файлом резюме, загрузка в MinIO через presigned-URL с поддержкой докачки (resume upload).
- 📋 **Заявки на план развития** — приём заявок, привязка к пользователям, просмотр администратором.
- 📧 **Уведомления по email** — асинхронные оповещения через SMTP (с graceful-фолбэком на no-op при отсутствии настроек).
- 🗄️ **Кэширование профилей** в Redis с TTL.
- 📝 **Структурированное логирование** (slog): человекочитаемый вывод локально, JSON в проде.
- 🛡️ **Graceful shutdown**, CORS, ограничение размера multipart-загрузок.

## Технологии

| Категория        | Стек                                                        |
|------------------|-------------------------------------------------------------|
| Язык             | Go 1.26                                                     |
| HTTP-фреймворк   | [Gin](https://github.com/gin-gonic/gin)                     |
| База данных      | PostgreSQL + [pgx/v5](https://github.com/jackc/pgx)         |
| Генерация SQL    | [sqlc](https://sqlc.dev)                                    |
| Миграции         | [golang-migrate](https://github.com/golang-migrate/migrate)|
| Кэш / сессии     | Redis ([go-redis/v9](https://github.com/redis/go-redis))   |
| Объектное хранилище | MinIO ([minio-go/v7](https://github.com/minio/minio-go)) |
| Аутентификация   | PASETO ([go-paseto](https://aidanwoods.dev/go-paseto))     |
| Хеширование      | Argon2id (`golang.org/x/crypto`)                            |
| Валидация        | [validator/v10](https://github.com/go-playground/validator)|
| Конфигурация     | [Viper](https://github.com/spf13/viper) (YAML + `.env`)    |
| Email            | [go-mail](https://github.com/wneessen/go-mail)             |

## Архитектура

Запрос проходит сквозь слои сверху вниз; зависимости направлены внутрь:

```
HTTP (Gin)
  └─ internal/transport/gin-http   — handlers, middleware, валидация, маппинг ошибок
       └─ internal/service          — бизнес-логика
            └─ internal/repository  — Postgres / Redis репозитории
                 └─ internal/db     — сгенерированный sqlc-код
```

Вспомогательные пакеты вынесены в `pkg/`: `auth` (PASETO + blacklist), `hash` (Argon2id),
`postgres`, `redis`, `minio`, `mailer`, `logger`.

## Быстрый старт

### Требования

- Go **1.26+**
- Docker (для PostgreSQL, Redis, MinIO)
- [`migrate`](https://github.com/golang-migrate/migrate) и [`sqlc`](https://sqlc.dev) (для миграций и генерации кода)

### 1. Клонирование и зависимости

```bash
git clone https://github.com/nhassl3/IpBuild-backend.git
cd IpBuild-backend
go mod download
```

### 2. Настройте `.env`

Создайте файл `.env` в корне проекта (см. раздел [Конфигурация](#конфигурация)).

### 3. Поднимите инфраструктуру

```bash
make postgres   # PostgreSQL 18
make redis      # Redis 7 с ACL
make minio      # MinIO (API :9000, консоль :9001)
make createdb   # создать базу данных
make migrate-up # применить миграции
```

### 4. Сборка и запуск

```bash
make build      # собрать бинарник в ./bin
make run        # или запустить напрямую через go run
```

Сервер поднимется на `localhost:8080` (по умолчанию для `local`-конфига).

## Конфигурация

Конфигурация разделена на две части:

- **Публичные настройки** — YAML-файлы в `config/` (`local.yaml`, `prod.yaml`).
- **Секреты** — файл `.env` в корне проекта.

Выбор конфига управляется переменными окружения:

| Переменная     | Назначение                                    | По умолчанию |
|----------------|-----------------------------------------------|--------------|
| `CONFIG_FILE`  | Прямой путь к YAML-конфигу                     | —            |
| `ENVIRONMENT`  | `local` / `prod` → выбирает YAML из `config/` | `local`      |
| `ENV_FILE`     | Путь к файлу секретов                          | `.env`       |

### Пример `.env`

```dotenv
# PostgreSQL
DB_NAME=unet
DB_USER=ipbuild
DB_PASSWORD=your-db-password

# Redis
REDIS_PASSWORD=your-redis-password

# PASETO (hex-ключ, 32 байта)
PASETO_KEY=c22782a3587fa8cf474eb336550d7b83e80941de8e696ddeee4038ad924399e9

# SMTP (опционально — без него уведомления отключаются)
SMTP_USERNAME=your-smtp-user
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=noreply@example.com
WORK_EMAIL=hr@example.com

# MinIO
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
```

> ⚠️ `.env` содержит секреты и не должен попадать в репозиторий. Сгенерируйте свой `PASETO_KEY`, например: `openssl rand -hex 32`.

## API

Базовый адрес: `http://localhost:8080`

### Публичные эндпоинты

| Метод  | Путь                     | Описание                                   |
|--------|--------------------------|--------------------------------------------|
| `POST` | `/auth/signup`           | Регистрация пользователя                   |
| `POST` | `/auth/login`            | Вход, выдача access/refresh-токенов        |
| `POST` | `/auth/refresh`          | Обновление токенов                         |
| `GET`  | `/api/vacancies/`        | Список вакансий                            |
| `GET`  | `/api/vacancies/:id`     | Вакансия по ID                             |
| `POST` | `/api/vacancies/respond` | Отклик на вакансию (анкета + резюме-файл)  |
| `POST` | `/api/plan/`             | Заявка на план развития                    |

### Требуют авторизации (`Authorization: Bearer <access_token>`)

| Метод  | Путь              | Описание                       |
|--------|-------------------|--------------------------------|
| `POST` | `/api/logout`     | Выход (токен в blacklist)      |
| `GET`  | `/api/plan/:id`   | Получить ответ по своей заявке |

### Только для администраторов (`role: admin`)

| Метод    | Путь                       | Описание                    |
|----------|----------------------------|-----------------------------|
| `POST`   | `/api/admin/vacancies/`    | Создать вакансию            |
| `PUT`    | `/api/admin/vacancies/:id` | Обновить вакансию           |
| `DELETE` | `/api/admin/vacancies/:id` | Удалить вакансию            |
| `GET`    | `/api/admin/vacancies/`    | Список откликов на вакансии |
| `GET`    | `/api/admin/vacancies/:id` | Отклик по ID                |
| `GET`    | `/api/admin/plans/`        | Список заявок на планы      |
| `GET`    | `/api/admin/plans/:id`     | Заявка по ID                |

<details>
<summary>Пример: регистрация</summary>

```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "full_name": "John Doe",
    "email": "john@example.com",
    "password": "supersecret"
  }'
```
</details>

## Команды Makefile

| Команда             | Назначение                          |
|---------------------|-------------------------------------|
| `make build`        | Собрать бинарник в `./bin`          |
| `make run`          | Запустить через `go run`            |
| `make runb`         | Запустить собранный бинарник        |
| `make test`         | Тесты с race-детектором и покрытием |
| `make lint`         | golangci-lint                       |
| `make sqlc`         | Сгенерировать Go-код из SQL         |
| `make migrate-up`   | Применить миграции                  |
| `make migrate-down` | Откатить последнюю миграцию         |
| `make postgres`     | Поднять PostgreSQL в Docker         |
| `make redis`        | Поднять Redis в Docker              |
| `make minio`        | Поднять MinIO в Docker              |
| `make tidy` / `vet` | `go mod tidy` / `go vet`            |

## Структура проекта

```
.
├── cmd/api/             # точка входа (main.go)
├── config/              # YAML-конфиги (local, prod)
├── db/query/            # SQL-запросы для sqlc
├── migrations/          # SQL-миграции (golang-migrate)
├── internal/
│   ├── app/             # сборка и запуск сервера
│   ├── config/          # загрузка конфигурации
│   ├── domain/          # доменные модели и ошибки
│   ├── db/              # сгенерированный sqlc-код
│   ├── repository/      # Postgres / Redis репозитории
│   ├── service/         # бизнес-логика
│   └── transport/       # HTTP-слой (Gin): handlers, middleware, валидация
├── pkg/                 # переиспользуемые пакеты (auth, hash, minio, redis, mailer, logger)
├── redis-config/        # конфиг и ACL для Redis
├── Makefile
└── sqlc.yaml
```
