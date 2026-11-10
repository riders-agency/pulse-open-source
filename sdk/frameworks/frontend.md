# React, Vue, Angular и Svelte

UI обращается к вашему серверному маршруту, например `GET /api/pulse/tasks`.
Готовый обработчик описан в [серверных примерах](typescript.md). В заголовке
передаётся пользовательская сессия Pulse, полученная при авторизации; общий
серверный ключ сюда не подходит. При собственной cookie-авторизации настройте
обмен сессией на сервере и используйте `credentials: "same-origin"`.

Общая функция чтения, которую можно разместить в приложении:

```ts
export async function loadTasks(projectId: string, session: string, signal: AbortSignal) {
  const query = new URLSearchParams({project_id: projectId});
  const response = await fetch(`/api/pulse/tasks?${query}`, {
    headers: {Authorization: `Bearer ${session}`}, signal, cache: "no-store",
  });
  if (!response.ok) throw new Error("Не удалось загрузить задачи");
  return response.json();
}
```

## React

```tsx
useEffect(() => {
  const controller = new AbortController();
  setLoading(true);
  loadTasks(projectId, session, controller.signal)
    .then(result => { if (!controller.signal.aborted) setTasks(result.tasks.items); })
    .catch(() => { if (!controller.signal.aborted) setError("Не удалось загрузить задачи"); })
    .finally(() => { if (!controller.signal.aborted) setLoading(false); });
  return () => controller.abort();
}, [projectId, session]);
```

## Vue / Nuxt

```ts
watch([projectId, session], ([project, token], _previous, onCleanup) => {
  const controller = new AbortController();
  onCleanup(() => controller.abort());
  loading.value = true;
  loadTasks(project, token, controller.signal)
    .then(result => { if (!controller.signal.aborted) tasks.value = result.tasks.items; })
    .catch(() => { if (!controller.signal.aborted) error.value = "Не удалось загрузить задачи"; })
    .finally(() => { if (!controller.signal.aborted) loading.value = false; });
}, {immediate: true});
```

## Angular

В сервисе используйте `HttpClient`; `switchMap` отменяет предыдущую подписку:

```ts
taskState$ = combineLatest([projectId$, session$]).pipe(
  switchMap(([project, token]) => http.get<TaskPage>("/api/pulse/tasks", {
    params: {project_id: project}, headers: {Authorization: `Bearer ${token}`},
  }).pipe(
    map(page => ({page, loading: false, error: null})),
    startWith({page: null, loading: true, error: null}),
    catchError(() => of({page: null, loading: false, error: "Не удалось загрузить задачи"})),
  )),
);
```

Покажите `loading` и `error` отдельно от пустого списка. Тип `TaskPage` экспортируется из SDK.

## Svelte 5

```ts
$effect(() => {
  const controller = new AbortController();
  loading = true;
  loadTasks(projectId, session, controller.signal)
    .then(result => { if (!controller.signal.aborted) tasks = result.tasks.items; })
    .catch(() => { if (!controller.signal.aborted) error = "Не удалось загрузить задачи"; })
    .finally(() => { if (!controller.signal.aborted) loading = false; });
  return () => controller.abort();
});
```

Это фрагменты для включения в приложение: состояние и авторизация принадлежат
ему. Список задач, сообщение об ошибке и индикатор загрузки должны быть видимы
пользователю. Подтверждение ИИ оформляйте отдельной кнопкой после просмотра
предложения; не вызывайте его в effect, watcher или при монтировании компонента.
