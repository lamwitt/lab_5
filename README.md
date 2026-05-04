# Books API

RESTful API для управления каталогом книг. Реализовано на Go с использованием Gin, GORM и PostgreSQL.

## Стек

- **Go 1.22** + **Gin** — HTTP-фреймворк
- **GORM** — ORM
- **PostgreSQL 16** — база данных
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

| Переменная    | По умолчанию              | Описание               |
|---------------|---------------------------|------------------------|
| `DB_USER`     | `student`                 | Пользователь PostgreSQL |
| `DB_PASSWORD` | `student_secure_password` | Пароль PostgreSQL       |
| `DB_NAME`     | `wp_labs`                 | Имя базы данных         |
| `DB_HOST`     | `localhost`               | Хост базы данных        |
| `DB_PORT`     | `5432`                    | Порт базы данных        |
| `PORT`        | `4200`                    | Порт приложения         |

## API

Базовый URL: `http://localhost:4200`

| Метод    | URI           | Описание                              | Статус успеха    |
|----------|---------------|---------------------------------------|------------------|
| `GET`    | `/books`      | Список книг с пагинацией              | `200 OK`         |
| `GET`    | `/books/:id`  | Получить книгу по ID                  | `200 OK`         |
| `POST`   | `/books`      | Создать книгу                         | `201 Created`    |
| `PUT`    | `/books/:id`  | Полностью обновить книгу              | `200 OK`         |
| `PATCH`  | `/books/:id`  | Частично обновить книгу               | `200 OK`         |
| `DELETE` | `/books/:id`  | Мягкое удаление книги                 | `204 No Content` |

### Пагинация (GET /books)

Query-параметры:

| Параметр | Тип | По умолчанию | Описание                |
|----------|-----|--------------|-------------------------|
| `page`   | int | `1`          | Номер страницы (≥ 1)    |
| `limit`  | int | `10`         | Элементов на странице (1–100) |

Пример ответа:

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Мастер и Маргарита",
      "author": "Михаил Булгаков",
      "description": "Роман о добре и зле",
      "year": 1967,
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-01-01T10:00:00Z"
    }
  ],
  "meta": {
    "total": 10,
    "page": 1,
    "limit": 10,
    "totalPages": 1
  }
}
```

## Примеры запросов (cURL)

**Создать книгу:**
```bash
curl -X POST http://localhost:4200/books \
  -H "Content-Type: application/json" \
  -d '{"title": "Мастер и Маргарита", "author": "Михаил Булгаков", "description": "Роман", "year": 1967}'
```

**Список с пагинацией:**
```bash
curl "http://localhost:4200/books?page=1&limit=5"
```

**Получить по ID:**
```bash
curl http://localhost:4200/books/<uuid>
```

**Полное обновление (PUT):**
```bash
curl -X PUT http://localhost:4200/books/<uuid> \
  -H "Content-Type: application/json" \
  -d '{"title": "Новое название", "author": "Автор", "description": "Описание", "year": 2000}'
```

**Частичное обновление (PATCH):**
```bash
curl -X PATCH http://localhost:4200/books/<uuid> \
  -H "Content-Type: application/json" \
  -d '{"year": 2001}'
```

**Мягкое удаление:**
```bash
curl -X DELETE http://localhost:4200/books/<uuid>
```

## Миграции

Миграции встроены в бинарник (через Go embed) и применяются автоматически при старте. Файлы миграций находятся в директории [`migrations/`](migrations/).

## Мягкое удаление

При `DELETE /books/:id` запись не удаляется физически — устанавливается поле `deleted_at`. Такие записи не возвращаются в `GET /books` и `GET /books/:id` (возвращается `404`).
