# Нативные серверные и мобильные клиенты

## Go: Gin, Echo и net/http

```bash
go get github.com/riders-agency/pulse-sdk-go@v0.1.0
```

`pulse.ForRequest(origin, request)` создаёт отдельный клиент, переносит
пользовательский Bearer-токен и отмену, включает таймаут и запрещает редиректы.
В Gin передайте `c.Request`, в Echo — `c.Request()`, в net/http — `r`.

```go
client, ctx, err := pulse.ForRequest(os.Getenv("PULSE_URL"), r)
if err != nil { http.Error(w, "User session required", http.StatusUnauthorized); return }
page, _, err := client.TasksAPI.ListTasks(ctx, r.URL.Query().Get("project_id")).Limit(20).Execute()
if err != nil { http.Error(w, "Pulse request failed", http.StatusBadGateway); return }
w.Header().Set("Content-Type", "application/json")
w.Header().Set("Cache-Control", "private, no-store")
_ = json.NewEncoder(w).Encode(page)
```

## ASP.NET Core

Клиент находится в [generated/csharp](../generated/csharp). Используйте
регистрацию сервисов из README внутри `src/Riders.Pulse`, добавляя авторизацию
на уровне scoped token provider. Не храните изменяемый пользовательский токен
в singleton. Настройте `HttpClient` с HTTPS, таймаутом и запретом редиректов;
передавайте `HttpContext.RequestAborted` в асинхронные операции клиента.

## Rails

Пакет [generated/ruby](../generated/ruby) подключается как локальный gem.
Создавайте `PulseSDK::Configuration` и `PulseSDK::ApiClient` в контексте
запроса; `configuration.access_token` получает сессию текущего пользователя.
Настройки общей интеграции можно держать в Rails credentials, пользовательский
токен — только в его сессии. Не вызывайте глобальный `configure` на каждом запросе.

## Rust: Axum и Actix Web

Клиент [generated/rust](../generated/rust) использует `reqwest` и async API.
Общий HTTP connection pool можно хранить в состоянии приложения, но
`apis::configuration::Configuration` с `bearer_access_token` создавайте для
каждого запроса. В `reqwest::Client::builder()` задайте таймаут и
`redirect(reqwest::redirect::Policy::none())`. Передавайте конфигурацию
в `apis::tasks_api::list_tasks`; не записывайте ошибки с исходными телами в публичный лог.

## Flutter

Добавьте исходники [Dart-клиента](../generated/dart) в `pubspec.yaml`:

```yaml
dependencies:
  pulse_sdk:
    git:
      url: https://github.com/riders-agency/pulse-open-source
      path: sdk/generated/dart
      ref: v0.1.0
```

`HttpBearerAuth.accessToken` получает сессию пользователя, `ApiClient` — origin
вашей установки. Создавайте `TasksApi` с этим клиентом. В Riverpod/Provider
жизненный цикл клиента должен следовать входу и выходу пользователя. Служебный
API-ключ нельзя включать в APK/IPA или Flutter Web bundle.

## SwiftUI

Исходники [Swift SDK](../generated/swift) подключаются как локальный Swift Package.
Слой модели представления вызывает async-методы клиента и отображает loading,
ошибку и результат. Пользовательская сессия хранится через системное защищённое
хранилище; приложение не содержит общего ключа владельца. При выходе отмените
задачи и удалите сессию. Настройте URLSession для своего origin и политики редиректов.

Эти рецепты не означают проверку всех версий перечисленных фреймворков.
Точный статус сборки и живых вызовов каждого SDK указан в [отчёте](../../docs/VALIDATION.md).
