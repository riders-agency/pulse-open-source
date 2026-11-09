# Laravel и Symfony

PHP-пакет: `riders-agency/pulse-sdk`, исходники в отдельном репозитории
[pulse-sdk-php](https://github.com/riders-agency/pulse-sdk-php).
Пока каталог Packagist не подключён, установите через Composer VCS:

```bash
composer config repositories.pulse vcs https://github.com/riders-agency/pulse-sdk-php
composer require riders-agency/pulse-sdk:^0.1
```

## Laravel

Provider обнаруживается через `extra.laravel.providers`. Опубликуйте конфигурацию:

```bash
php artisan vendor:publish --tag=pulse-config
```

Укажите `PULSE_URL` в серверном окружении. Provider берёт Bearer-сессию текущего
пользователя из запроса и регистрирует `TasksApi`, `AIApi`, `AuthApi` через
`scoped`, включая корректную смену request scope в долгоживущем worker.

```php
use Riders\Pulse\Api\TasksApi;

// В контроллере после middleware вашей авторизации:
public function index(\Illuminate\Http\Request $request, TasksApi $pulse) {
    $input = $request->validate(['project_id' => 'required|uuid']);
    return response()->json($pulse->listTasks($input['project_id'], 20))
        ->header('Cache-Control', 'private, no-store');
}
```

[Механизм discovery](https://laravel.com/framework/docs/13.x/packages).
Без пользовательского Bearer-токена provider возвращает ошибку авторизации;
он не подставляет общий ключ администратора.

## Symfony

Скопируйте [объявления сервисов](php/symfony-services.yaml) в конфигурацию
контейнера либо импортируйте файл из установленного пакета. `PULSE_URL`
передаётся через env. `RequestFactory` читает текущий запрос из `RequestStack`;
создаваемые клиенты имеют `shared: false`.

```php
use Riders\Pulse\Api\TasksApi;
use Symfony\Component\HttpFoundation\{JsonResponse, Request};

public function tasks(Request $request, TasksApi $pulse): JsonResponse {
    $project = (string) $request->query->get('project_id', '');
    if ($project === '') { return new JsonResponse(['error' => 'project_id_required'], 400); }
    return new JsonResponse($pulse->listTasks($project, 20), 200,
        ['Cache-Control' => 'private, no-store']);
}
```

[Контейнер Symfony](https://symfony.com/doc/current/service_container.html).
Включите `auth()` в конфигурацию при необходимости. Обрабатывайте `ApiException`
на уровне приложения; не показывайте сырое тело исключения посетителю.

## Обычный PHP / серверное задание

`Integration\ClientFactory` не зависит от фреймворков. Она использует отдельную
Configuration, HTTPS, таймаут и запрет редиректов. Передайте ограниченный токен
из окружения только доверенному серверному заданию. Для запросов пользователей
создавайте фабрику с токеном соответствующего пользователя.
