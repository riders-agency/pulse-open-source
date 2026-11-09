<?php
declare(strict_types=1);

namespace Riders\Pulse\Integration;

use GuzzleHttp\Client;
use Riders\Pulse\Api\{AIApi, AuthApi, TasksApi};
use Riders\Pulse\Configuration;

/** Per-user factory. It does not use the generated global default configuration. */
final class ClientFactory
{
    private Configuration $configuration;
    private Client $http;

    public function __construct(string $baseUrl, string $token, float $timeout = 60)
    {
        $url = parse_url($baseUrl);
        if (!$url || empty($url['host']) || isset($url['user'], $url['pass']) || isset($url['user']) || isset($url['query']) || isset($url['fragment']) || !in_array($url['path'] ?? '', ['', '/'], true)) {
            throw new \InvalidArgumentException('Pulse URL must be an origin');
        }
        if (($url['scheme'] ?? '') !== 'https' && !(($url['scheme'] ?? '') === 'http' && in_array($url['host'], ['localhost', '127.0.0.1', '[::1]'], true))) {
            throw new \InvalidArgumentException('Pulse requires HTTPS outside loopback');
        }
        if (!preg_match('/^\S+$/D', $token) || $timeout <= 0) {
            throw new \InvalidArgumentException('A user session and positive timeout are required');
        }
        $this->configuration = (new Configuration())->setHost(rtrim($baseUrl, '/'))->setAccessToken($token);
        $this->http = new Client(['timeout' => $timeout, 'allow_redirects' => false]);
    }

    public function tasks(): TasksApi { return new TasksApi($this->http, $this->configuration); }
    public function ai(): AIApi { return new AIApi($this->http, $this->configuration); }
    public function auth(): AuthApi { return new AuthApi($this->http, $this->configuration); }
}
