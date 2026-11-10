# FastAPI, Django и Flask

Установите [Python SDK](../python/README.md). Ядро работает без зависимости
от фреймворка; `from_request` принимает объект с case-insensitive `headers`.

## FastAPI

```python
import os
from typing import Annotated
from fastapi import Depends, FastAPI, HTTPException, Request
from pulse_sdk import PulseClient
from pulse_sdk.frameworks import from_request

app = FastAPI()

def pulse_client(request: Request):
    try:
        return from_request(os.environ["PULSE_URL"], request)
    except ValueError:
        raise HTTPException(status_code=401, detail="User session required") from None

@app.get("/pulse/tasks")
def tasks(project_id: str, pulse: Annotated[PulseClient, Depends(pulse_client)]):
    return pulse.list_tasks(project_id, limit=20)
```

SDK синхронный: используйте `def`-обработчик, чтобы FastAPI выполнял его в пуле
потоков. В `async def` перенесите вызов в thread pool.
[Зависимости FastAPI](https://fastapi.tiangolo.com/tutorial/dependencies/).

## Django

```python
from django.conf import settings
from django.http import JsonResponse
from pulse_sdk.frameworks import from_request

def tasks(request):
    if request.method != "GET":
        return JsonResponse({"error": "method_not_allowed"}, status=405)
    try:
        pulse = from_request(settings.PULSE_URL, request)
    except ValueError:
        return JsonResponse({"error": "user_session_required"}, status=401)
    project = request.GET.get("project_id")
    if not project:
        return JsonResponse({"error": "project_id_required"}, status=400)
    response = JsonResponse(pulse.list_tasks(project, limit=20))
    response["Cache-Control"] = "private, no-store"
    return response
```

Подключите view через `path("api/pulse/tasks", tasks)`. Не помещайте клиент
с токеном пользователя в глобальные settings или singleton middleware.

## Flask

```python
import os
from flask import Flask, abort, jsonify, request
from pulse_sdk.frameworks import from_request

app = Flask(__name__)

@app.get("/pulse/tasks")
def tasks():
    try:
        pulse = from_request(os.environ["PULSE_URL"], request)
    except ValueError:
        abort(401)
    project = request.args.get("project_id")
    if not project:
        abort(400)
    response = jsonify(pulse.list_tasks(project, limit=20))
    response.headers["Cache-Control"] = "private, no-store"
    return response
```

Во всех приложениях зарегистрируйте обработчик `PulseError`, возвращающий
статус ошибки без сырых тел/заголовков. Примеры не заменяют настройки авторизации,
CSRF для ваших cookie-маршрутов и журналирования собственного приложения.
