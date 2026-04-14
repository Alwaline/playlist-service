# Playlist Service

Сервис управления плейлистами — позволяет создавать персональные плейлисты, добавлять и удалять треки, получать содержимое плейлиста с метаданными. Часть backend-системы музыкального стримингового сервиса (аналог Spotify).

## API

| Method | Path | Description | Status codes |
|--------|------|-------------|--------------|
| `POST` | `/api/v1/playlists` | Создать плейлист | 201, 400, 401 |
| `POST` | `/api/v1/playlists/{id}/tracks` | Добавить трек в плейлист | 200, 400, 401, 403, 404, 409 |
| `GET` | `/api/v1/playlists/{id}/tracks` | Получить содержимое плейлиста | 200, 401, 403, 404 |
| `DELETE` | `/api/v1/playlists/{id}/tracks/{track_id}` | Удалить трек из плейлиста | 204, 401, 403, 404 |

Все запросы требуют заголовок `X-User-ID: <user_id>`.

## Быстрый старт

```bash
# Поднять инфраструктуру (PostgreSQL, Redis, Kafka)
make docker-up-infra

# Применить миграции
make migrate-up

# Запустить сервис
make run

# Тесты
make test

# Весь стек (сервис + инфра + observability)
make docker-up-all

# Swagger UI: http://localhost:8080/swagger/index.html
```

## Технологии

| Компонент | Технология |
|-----------|-----------|
| Язык | Go 1.26 |
| HTTP-сервер | fasthttp |
| DI | uber/fx |
| БД | PostgreSQL 17 (pgxpool) |
| Кэш | Redis 7 (go-redis) |
| Очередь | Apache Kafka 3.7 (segmentio/kafka-go) |
| Конфигурация | Viper |
| Трейсинг | OpenTelemetry + Jaeger |
| Метрики | Prometheus + Grafana |
| Логи | slog (JSON) + Loki |
| CI | GitHub Actions |

## Архитектура

```
HTTP Request
    │
    ▼
Handler (X-User-ID auth, JSON validation)
    │
    ▼
Use Case (бизнес-логика: проверка владельца, идемпотентность)
    ├── Cache.Get()     ──► Redis     (fast path, TTL 5 min)
    │       │ miss
    │       ▼
    └── Repository.Get() ──► PostgreSQL ──► Cache.Set()
    │
    ├── Repository.Write() ──► PostgreSQL
    └── Producer.Publish() ──► Kafka (playlist.track_added)
                                │
                    Kafka Consumer (track.deleted)
                                │
                    Repository.RemoveFromAll() ──► PostgreSQL
                    Cache.Invalidate()         ──► Redis
```

**Денормализация:** метаданные треков (название, исполнитель, длительность) хранятся локально в таблице `track_meta` — это позволяет отдавать содержимое плейлиста без синхронных вызовов в Track Service.

**Degraded mode:** при недоступном Redis все операции прозрачно продолжают работу через PostgreSQL.

## Kafka-топики

| Топик | Продюсер | Консьюмер |
|-------|----------|-----------|
| `playlists` | Playlist Service | Queue Service (инвалидация очереди) |
| `tracks` | Track Service (внешний) | Playlist Service (удаление трека из плейлистов) |

## Модель данных

```sql
playlists        — id, owner_id, name, created_at, updated_at
track_meta       — track_id, title, artist, duration_sec
playlist_tracks  — playlist_id, track_id, position, added_at
```

## Справочник конфигурации

| Ключ YAML | Переменная окружения | По умолчанию | Описание |
|-----------|---------------------|-------------|----------|
| `service_name` | `SERVICE_NAME` | `playlist-service` | Имя сервиса |
| `http_port` | `HTTP_PORT` | `8080` | Порт HTTP-сервера |
| `log_level` | `LOG_LEVEL` | `info` | Уровень: debug, info, warn, error |
| `shutdown_timeout` | `SHUTDOWN_TIMEOUT` | `15s` | Таймаут graceful shutdown |
| `tracing_enabled` | `TRACING_ENABLED` | `true` | Включить трейсинг |
| `otlp_endpoint` | `OTLP_ENDPOINT` | `localhost:4317` | Адрес OTLP-коллектора (Jaeger) |
| `postgres.dsn` | `POSTGRES_DSN` | `postgres://postgres:postgres@localhost:5432/service?sslmode=disable` | DSN PostgreSQL |
| `postgres.max_conns` | `POSTGRES_MAX_CONNS` | `10` | Макс. соединений в пуле |
| `postgres.min_conns` | `POSTGRES_MIN_CONNS` | `2` | Мин. соединений в пуле |
| `redis.addr` | `REDIS_ADDR` | `localhost:6379` | Адрес Redis |
| `redis.pool_size` | `REDIS_POOL_SIZE` | `10` | Размер пула Redis |
| `kafka.brokers` | `KAFKA_BROKERS` | `localhost:9092` | Адреса Kafka-брокеров (producer) |
| `kafka_consumer.brokers` | `KAFKA_CONSUMER_BROKERS` | `localhost:9092` | Адреса Kafka-брокеров (consumer) |
| `kafka_consumer.group_id` | `KAFKA_CONSUMER_GROUP_ID` | `playlist-service` | Consumer group ID |
| `kafka_consumer.topic` | `KAFKA_CONSUMER_TOPIC` | `tracks` | Топик для событий удаления треков |

## Makefile-цели

| Цель | Описание |
|------|----------|
| `make build` | Собрать бинарник в `./bin/` |
| `make run` | Запустить сервис локально |
| `make test` | Запустить тесты |
| `make lint` | Запустить golangci-lint |
| `make swagger` | Сгенерировать/обновить OpenAPI-спецификацию |
| `make docker-build` | Собрать Docker-образ |
| `make docker-up-infra` | Поднять PostgreSQL + Redis + Kafka |
| `make docker-up-all` | Весь стек (сервис + инфра + Jaeger + Prometheus + Grafana + Loki) |
| `make docker-down` | Остановить все контейнеры |
| `make migrate-up` | Применить миграции |
| `make migrate-down` | Откатить последнюю миграцию |
| `make migrate-create` | Создать новую миграцию |