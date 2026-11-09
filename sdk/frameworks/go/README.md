# Pulse SDK · Go

Go 1.23+. API v1: задачи, ИИ-сессии и подтверждение предложений.

```bash
go get github.com/riders-agency/pulse-sdk-go@v0.1.0
```

## HTTP-приложение / Gin / Echo

```go
import pulse "github.com/riders-agency/pulse-sdk-go"

client, ctx, err := pulse.ForRequest(origin, request)
if err != nil { /* вернуть 401; не печатать заголовки */ return }
page, response, err := client.TasksAPI.ListTasks(ctx, projectID).Limit(20).Execute()
_ = page
_ = response
_ = err
```

В Gin передайте `c.Request`, в Echo — `c.Request()`. Фабрика создаёт клиент для
пользователя запроса, переносит отмену и Bearer-сессию, требует HTTPS вне loopback,
задаёт таймаут 60 секунд и запрещает редиректы. Отсутствующая сессия отклоняется.

Для фоновых заданий создавайте `Configuration` и используйте ограниченный токен
через `ContextAccessToken`. Не держите пользовательский токен в глобальной
изменяемой переменной. ИИ-предложение подтверждается отдельным вызовом после
решения пользователя; SDK не делает это автоматически.

[Методы задач](docs/TasksAPI.md) · [Методы ИИ](docs/AIAPI.md).
Сборка, vet, race-тест request scope и живой сценарий задач/ИИ проверены.

API и модели сгенерированы OpenAPI Generator 7.20.0; `request.go` — отдельный
проверяемый адаптер. `go test -race ./...` запускает его тесты.
