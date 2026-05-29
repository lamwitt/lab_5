# Books API

RESTful API для управления каталогом книг с JWT-аутентификацией, OAuth 2.0 (Yandex) и кешированием через Redis.

## Стек

- **Go 1.22** + **Gin** — HTTP-фреймворк
- **GORM** — ORM
- **PostgreSQL 16** — база данных
- **Redis 7** — кеширование данных и хранение JTI токенов
- **golang-migrate** — миграции схемы БД
- **Docker / Docker Compose** — контейнеризация

## Запуск

### 1. Создайте файл `.env`

```bash
cp .env.example .env
```

### 2. Запустите через Docker Compose

```bash
docker-compose up --build
```

Миграции применяются автоматически при старте приложения.

API будет доступно на `http://localhost:4200`.

## Переменные окружения

| Переменная         | По умолчанию                | Описание                        |
|--------------------|-----------------------------|---------------------------------|
| `DB_USER`          | `student`                   | Пользователь PostgreSQL         |
| `DB_PASSWORD`      | `student_secure_password`   | Пароль PostgreSQL               |
| `DB_NAME`          | `wp_labs`                   | Имя базы данных                 |
| `DB_HOST`          | `localhost`                  | Хост базы данных                |
| `DB_PORT`          | `5432`                      | Порт базы данных                |
| `PORT`             | `4200`                      | Порт приложения                 |
| `JWT_ACCESS_SECRET`  | `change_me_access_secret` | Секрет access-токена            |
| `JWT_REFRESH_SECRET` | `change_me_refresh_secret`| Секрет refresh-токена           |
| `JWT_ACCESS_EXPIRATION`  | `15m`                 | Время жизни access-токена       |
| `JWT_REFRESH_EXPIRATION` | `7d`                  | Время жизни refresh-токена      |
| `CLIENT_ID`        | —                           | Yandex OAuth Client ID          |
| `CLIENT_SECRET`    | —                           | Yandex OAuth Client Secret      |
| `CALLBACK_URL`     | `http://localhost:4200/auth/oauth/yandex/callback` | OAuth callback URL |
| `REDIS_HOST`       | `localhost`                 | Хост Redis                      |
| `REDIS_PORT`       | `6379`                      | Порт Redis                      |
| `REDIS_PASSWORD`   | `redis_secure_password`     | Пароль Redis                    |
| `CACHE_TTL_DEFAULT`| `300`                       | TTL кеша в секундах (5 минут)   |

## Redis и кеширование

### Архитектура

Логика работы с Redis инкапсулирована в модуле `internal/cache/cache.go`. Сервисный слой использует стратегию **Cache-Aside (Lazy Loading)**:

1. При GET-запросе сначала проверяется кеш.
2. При промахе (Cache Miss) данные загружаются из БД, сохраняются в кеш с TTL и возвращаются клиенту.
3. При любой операции записи (POST/PUT/PATCH/DELETE) соответствующие ключи инвалидируются.

### Структура ключей

| Ключ | Описание |
|------|----------|
| `wp:books:list:user:{userId}:page:{page}:limit:{limit}` | Список книг пользователя (с пагинацией) |
| `wp:books:item:{bookId}` | Отдельная книга |
| `wp:auth:user:{userId}:profile` | Профиль пользователя |
| `wp:auth:user:{userId}:jti:{jti}` | JTI активного access-токена |

Все ключи имеют TTL. Пароли и полные токены в кеше не хранятся.

### Поведение кеша по эндпоинтам

| Метод | URI | Действие с кешем |
|-------|-----|-----------------|
| `GET` | `/books` | Проверка кеша. При промахе — запись результата в кеш. |
| `POST` | `/books` | Инвалидация ключей списка книг пользователя. |
| `PUT/PATCH` | `/books/:id` | Инвалидация ключа списка и ключа конкретной книги. |
| `DELETE` | `/books/:id` | Инвалидация ключа списка и ключа книги. |
| `GET` | `/auth/whoami` | Проверка кеша профиля. При промахе — запись в кеш. |
| `POST` | `/auth/logout` | Инвалидация JTI access-токена и кеша профиля. |

### Проверка кеша через Redis CLI

Подключиться к контейнеру Redis:

```bash
docker exec -it wp_labs_redis_5 redis-cli -a redis_secure_password
```

Полезные команды:

```bash
# Просмотр всех ключей приложения
KEYS wp:*

# Получить значение ключа
GET wp:books:list:user:{userId}:page:1:limit:10

# Проверить TTL ключа
TTL wp:books:list:user:{userId}:page:1:limit:10

# Удалить ключ вручную
DEL wp:books:list:user:{userId}:page:1:limit:10

# Массовое удаление по паттерну
UNLINK wp:books:*

# Очистить всю базу (только для тестов)
FLUSHDB
```

### Access токены и Logout

При входе JTI access-токена сохраняется в Redis с TTL = времени жизни токена (15 минут). При каждом запросе middleware проверяет наличие JTI в Redis. При `/auth/logout` JTI удаляется из Redis — доступ запрещается мгновенно, не дожидаясь истечения JWT.

## API

Базовый URL: `http://localhost:4200`

### Авторизация

| Метод | URI | Описание | Статус успеха |
|-------|-----|----------|---------------|
| `POST` | `/auth/register` | Регистрация | `201 Created` |
| `POST` | `/auth/login` | Вход (устанавливает HttpOnly cookies) | `200 OK` |
| `POST` | `/auth/refresh` | Обновление токенов | `200 OK` |
| `GET` | `/auth/whoami` | Профиль текущего пользователя | `200 OK` |
| `POST` | `/auth/logout` | Выход из текущей сессии | `200 OK` |
| `POST` | `/auth/logout-all` | Выход со всех устройств | `200 OK` |
| `POST` | `/auth/forgot-password` | Запрос сброса пароля | `200 OK` |
| `POST` | `/auth/reset-password` | Сброс пароля | `200 OK` |
| `GET` | `/auth/oauth/yandex` | Инициация OAuth через Яндекс | `302` |
| `GET` | `/auth/oauth/yandex/callback` | Callback OAuth | `302` |

### Книги (требуют авторизации)

| Метод | URI | Описание | Статус успеха |
|-------|-----|----------|---------------|
| `GET` | `/books` | Список книг с пагинацией | `200 OK` |
| `GET` | `/books/:id` | Получить книгу по ID | `200 OK` |
| `POST` | `/books` | Создать книгу | `201 Created` |
| `PUT` | `/books/:id` | Полностью обновить книгу | `200 OK` |
| `PATCH` | `/books/:id` | Частично обновить книгу | `200 OK` |
| `DELETE` | `/books/:id` | Мягкое удаление книги | `204 No Content` |

### Пагинация (GET /books)

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|--------------|----------|
| `page` | int | `1` | Номер страницы (≥ 1) |
| `limit` | int | `10` | Элементов на странице (1–100) |

## Миграции

Миграции встроены в бинарник (через Go embed) и применяются автоматически при старте. Файлы миграций находятся в директории [`migrations/`](migrations/).

## Мягкое удаление

При `DELETE /books/:id` запись не удаляется физически — устанавливается поле `deleted_at`. Такие записи не возвращаются в `GET /books` и `GET /books/:id` (возвращается `404`).
