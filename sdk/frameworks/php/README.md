# Pulse SDK · PHP

PHP 8.1+, Composer, Guzzle. API v1: задачи, ИИ-диалоги и отдельное подтверждение
предложений. Пакет включает Laravel provider и Symfony request factory.

## Установка

```bash
composer require riders-agency/pulse-sdk:^0.1
```

[Пакет в Packagist](https://packagist.org/packages/riders-agency/pulse-sdk).
Composer получает стабильную версию из общего каталога; добавлять VCS-репозиторий
в приложение не нужно.

Для установки непосредственно из GitHub также доступен Composer VCS:

```bash
composer config repositories.pulse vcs https://github.com/riders-agency/pulse-sdk-php
composer require riders-agency/pulse-sdk:^0.1
```

## Использование

```php
require 'vendor/autoload.php';
use Riders\Pulse\Integration\ClientFactory;

$pulse = new ClientFactory(getenv('PULSE_URL'), getenv('PULSE_ACCESS_TOKEN'));
$page = $pulse->tasks()->listTasks(getenv('PULSE_PROJECT_ID'), 20);
// Передайте $page в свой интерфейс. Не печатайте токен или ответы авторизации.
```

Фабрика создаёт отдельную конфигурацию, требует HTTPS вне loopback, задаёт таймаут
и запрещает редиректы. SDK не подтверждает предложения и не повторяет записи
автоматически. После потери ответа сначала прочитайте задачу/ИИ-сессию.

## Фреймворки

Laravel автоматически обнаруживает `Riders\Pulse\Laravel\PulseServiceProvider`.
Задайте `PULSE_URL`, выполните `php artisan vendor:publish --tag=pulse-config`.
Provider использует Bearer-сессию текущего запроса и scoped bindings для
`TasksApi`, `AIApi`, `AuthApi`. Общий ключ администратора не подставляется.

Symfony: импортируйте `symfony-services.yaml`. Фабрика получает сессию из
RequestStack, клиент создаётся с `shared: false`.

[Примеры контроллеров и проверка](https://github.com/riders-agency/pulse-open-source/blob/main/sdk/frameworks/php.md).
[Методы задач](docs/Api/TasksApi.md) · [Методы ИИ](docs/Api/AIApi.md).

Сгенерированные API/модели получены из OpenAPI Generator 7.20.0. Адаптеры
`lib/Integration`, `lib/Laravel`, `lib/Symfony` поддерживаются отдельно.
Live-сценарий задач/ИИ и Laravel 13 / Symfony 7.4 проверены на синтетических данных.
