<?php
declare(strict_types=1);

namespace Riders\Pulse\Laravel;

use Illuminate\Auth\AuthenticationException;
use Illuminate\Support\ServiceProvider;
use Riders\Pulse\Api\{AIApi, AuthApi, TasksApi};
use Riders\Pulse\Integration\ClientFactory;

final class PulseServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        $this->mergeConfigFrom(__DIR__ . '/../../config/pulse.php', 'pulse');
        $this->app->scoped(ClientFactory::class, function ($app) {
            $token = $app['request']->bearerToken();
            if (!$token) { throw new AuthenticationException('Pulse user session required'); }
            return new ClientFactory((string) $app['config']->get('pulse.url'), $token, (float) $app['config']->get('pulse.timeout', 60));
        });
        foreach ([TasksApi::class => 'tasks', AIApi::class => 'ai', AuthApi::class => 'auth'] as $class => $method) {
            $this->app->scoped($class, fn ($app) => $app->make(ClientFactory::class)->$method());
        }
    }

    public function boot(): void
    {
        $this->publishes([__DIR__ . '/../../config/pulse.php' => config_path('pulse.php')], 'pulse-config');
    }
}
