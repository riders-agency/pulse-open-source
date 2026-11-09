# Riders Pulse — сервер и исполнитель

Go API обслуживает авторизацию, компании и проекты, клиентов и платежи, задачи,
визуальные сценарии, чат и ИИ-сотрудника. Отдельный `cmd/worker` исполняет сценарии
и очередь событий. Бизнес-данные находятся в PostgreSQL.

## Границы ответственности

- `internal/modules` — доменные модули с собственными маршрутами, сервисами,
  репозиториями и обработчиками действий.
- `internal/platform/app` — сборка приложения и шлюзы gRPC.
- `internal/runtime/kernel` — граф исполнения, аренда задания, fencing и повторы.
- `internal/runtime/effects` — запись намерения и результата внешнего действия.
- `internal/modules/ai` — контекст, инструменты, права, подтверждения и вызовы модели.
- `internal/gen/ai_agent/v1` — сгенерированные Go-контракты gRPC.

Python вызывает типизированные шлюзы Go. Права и применение изменений остаются
на сервере; доступ к бизнес-таблицам не передаётся оркестратору.

## Локальный запуск

Запускайте из корня репозитория после подготовки PostgreSQL по корневому README:

```bash
export DATABASE_URL='postgres://pulse:pulse@127.0.0.1:55432/pulse?sslmode=disable'
export HTTP_ADDR='127.0.0.1:9820'
export PULSE_RUNTIME_TOPOLOGY='split-worker'
go run ./apps/api/cmd/api
```

Миграции применяются при запуске. Worker запускается отдельно:
`go run ./apps/api/cmd/worker` с той же конфигурацией БД и секретов.
`.env.example` содержит только локальные примеры; рабочие значения не коммитятся.

## Проверка

Из корня: `scripts/check-go-quality.sh`, `scripts/check-openapi-contracts.sh`
и `scripts/check-proto-contracts.sh`. Для тестов с БД задайте
`PULSE_TEST_DATABASE_URL` на отдельную тестовую базу. Общая проверка
`scripts/check.sh` проходит без архитектурных исключений. Изменения чужих данных
идут через команды модуля-владельца, чтение — через зарегистрированные контракты.
Реестр `docs/architecture/module-ownership-exceptions.tsv` пуст; устройство границ
описано в [архитектурном решении](../../docs/architecture/controlled-module-boundaries.md).
