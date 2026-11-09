# Серверные фреймворки TypeScript

Установите [TypeScript SDK](../typescript/README.md). Экспорт `/frameworks`
не тянет фреймворки в зависимости. `PULSE_URL` содержит origin вашей установки.

## Next.js App Router

Файл `app/api/pulse/tasks/route.ts`:

```ts
import { handleTaskList } from "@riders-agency/pulse-sdk/frameworks";
export const runtime = "nodejs";
export function GET(request: Request) {
  return handleTaskList(process.env.PULSE_URL!, request);
}
```

Маршрут принимает `project_id` и пользовательский заголовок `Authorization`.
Без сессии возвращает 401. Чтение не кешируется; отмена запроса передаётся
транспорту. Тело ошибки провайдера не выдаётся посетителю.

## Nuxt / Nitro и SvelteKit

Nuxt/Nitro с H3 2, `server/api/pulse/tasks.get.ts`:

```ts
import { handleTaskList } from "@riders-agency/pulse-sdk/frameworks";
export default defineEventHandler(event =>
  handleTaskList(useRuntimeConfig(event).pulseUrl, event.req));
```

SvelteKit, `src/routes/api/pulse/tasks/+server.ts`:

```ts
import { PULSE_URL } from "$env/static/private";
import { handleTaskList } from "@riders-agency/pulse-sdk/frameworks";
export const GET = ({ request }: { request: Request }) => handleTaskList(PULSE_URL, request);
```

## NestJS

Provider привязан к запросу, чтобы токены пользователей не попадали в singleton:

```ts
import { Module, Scope } from "@nestjs/common";
import { REQUEST } from "@nestjs/core";
import { PulseClient } from "@riders-agency/pulse-sdk";
import { pulseFromAuthorization } from "@riders-agency/pulse-sdk/frameworks";

@Module({
  providers: [{
    provide: PulseClient, scope: Scope.REQUEST, inject: [REQUEST],
    useFactory: (request: { headers: { authorization?: string } }) =>
      pulseFromAuthorization(process.env.PULSE_URL!, request.headers.authorization),
  }],
  exports: [PulseClient],
})
export class PulseModule {}
```

Импортируйте модуль в feature-модуль и внедрите `PulseClient` в сервис.
Преобразуйте отсутствие Bearer-сессии в 401 вашим auth guard или exception filter.
[Механизм providers](https://docs.nestjs.com/fundamentals/custom-providers).

## Express и Fastify

Создавайте клиент внутри обработчика, после вашего auth middleware:

```ts
import { pulseFromAuthorization } from "@riders-agency/pulse-sdk/frameworks";
const pulse = pulseFromAuthorization(process.env.PULSE_URL!, request.headers.authorization);
const page = await pulse.listTasks(projectId, {limit: 20});
// Express: response.json(page). Fastify: return page.
```

Передайте ошибки в штатный error handler; `PulseError.status` содержит HTTP-код.
Не печатайте заголовок авторизации. Для cookie-сессии своего сайта получите
пользовательский токен на сервере и передайте его фабрике; браузерный cookie
вашего приложения не заменяет авторизацию Pulse автоматически.
