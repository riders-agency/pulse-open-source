<?php
declare(strict_types=1);

namespace Riders\Pulse\Symfony;

use Riders\Pulse\Integration\ClientFactory;
use Symfony\Component\HttpFoundation\RequestStack;
use Symfony\Component\HttpKernel\Exception\UnauthorizedHttpException;

final class RequestFactory
{
    public function __construct(private RequestStack $requests, private string $baseUrl) {}

    public function create(): ClientFactory
    {
        $header = $this->requests->getCurrentRequest()?->headers->get('Authorization', '') ?? '';
        if (!preg_match('/^Bearer ([^\s]+)$/iD', $header, $match)) {
            throw new UnauthorizedHttpException('Bearer', 'Pulse user session required');
        }
        return new ClientFactory($this->baseUrl, $match[1]);
    }
}
