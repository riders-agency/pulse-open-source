import type { AriaLabelConfig } from "@xyflow/react";

export const graphAriaLabels = {
  "controls.ariaLabel": "Управление графом",
  "controls.zoomIn.ariaLabel": "Увеличить масштаб",
  "controls.zoomOut.ariaLabel": "Уменьшить масштаб",
  "controls.fitView.ariaLabel": "Показать весь граф",
  "node.a11yDescription.default": "Нажмите Enter, чтобы выбрать узел и прочитать его инструкцию. Escape снимает выбор.",
  "edge.a11yDescription.default": "Переход между узлами графа.",
} satisfies Partial<AriaLabelConfig>;

export const graphLabels: Record<string, { title: string; instruction: string; kind: string }> = {
  __start__: { title: "Вход", instruction: "Go передаёт подтверждённый контекст пользователя и проекта через gRPC.", kind: "boundary" },
  ingest_turn: { title: "Принять запрос", instruction: "Принять сообщение и восстановить состояние диалога. Контекст страницы не даёт дополнительных прав.", kind: "context" },
  recover_active_intent: { title: "Восстановить контекст", instruction: "Учитывать активную задачу и незавершённое предложение, чтобы не потерять текущую работу.", kind: "context" },
  classify_intent: { title: "Определить намерение", instruction: "Выбрать намерение из заданной схемы. Вызов ModelGateway учитывается отдельно; результат модели ещё не является разрешением на действие.", kind: "model" },
  proposal_gate: { title: "Выбрать разрешённый путь", instruction: "Проверить активную карточку и направить запрос по зарегистрированной ветке графа.", kind: "policy" },
  proposal_editor: { title: "Изменить предложение", instruction: "Обновить черновик карточки. Изменённые поля снова предъявляются человеку до записи бизнес-данных.", kind: "proposal" },
  approval_decision: { title: "Применить решение", instruction: "Передать решение в Go ToolGateway. Сервер повторно проверяет доступ, состояние предложения и ключ идемпотентности.", kind: "policy" },
  conversation_reply: { title: "Ответить", instruction: "Подготовить ответ или совет. Ветка совета не получает инструменты изменения данных.", kind: "model" },
  calculator_reply: { title: "Вычислить", instruction: "Выполнить только разрешённые арифметические операции. Произвольный код не исполняется.", kind: "read" },
  conflict_choice: { title: "Уточнить выбор", instruction: "Попросить пользователя разрешить конфликт с уже открытым предложением.", kind: "policy" },
  read_context: { title: "Прочитать контекст", instruction: "Вызвать типизированные инструменты чтения. Go проверяет проект и права на каждое обращение.", kind: "read" },
  workflow_question: { title: "Разобрать сценарий", instruction: "Получить данные сценария через инструменты чтения и объяснить их без изменения рабочего процесса.", kind: "read" },
  workflow_patch_workflow: { title: "Предложить правку сценария", instruction: "Подготовить структурированный patch. Применение требует подтверждения и серверной проверки.", kind: "proposal" },
  task_create_workflow: { title: "Подготовить задачу", instruction: "Собрать поля task.create, при необходимости найти клиента через разрешённый инструмент. Передать карточку на подтверждение.", kind: "proposal" },
  task_move_workflow: { title: "Предложить изменение задачи", instruction: "Подготовить task.move с явными полями изменений и показать их человеку.", kind: "proposal" },
  client_create_workflow: { title: "Подготовить клиента", instruction: "Сформировать client.create. Создание записи остаётся за Go после подтверждения.", kind: "proposal" },
  compound_workflow: { title: "Составить последовательность", instruction: "Разложить составной запрос на ограниченный план, сохраняя контроль подтверждений для изменений.", kind: "proposal" },
  await_approval: { title: "Ждать человека", instruction: "LangGraph interrupt сохраняет контрольную точку и приостанавливает эту ветку. Продолжение — после решения пользователя.", kind: "approval" },
  finalizer: { title: "Зафиксировать результат", instruction: "Собрать ответ из наблюдений инструментов, сохранить состояние диалога и сведения о выполнении.", kind: "result" },
  __end__: { title: "Завершение", instruction: "Вернуть интерфейсу результат. Пользователь может открыть созданную сущность или продолжить диалог.", kind: "boundary" },
};

export const taskPath = ["__start__", "ingest_turn", "recover_active_intent", "classify_intent", "proposal_gate", "task_create_workflow", "await_approval", "approval_decision", "finalizer", "__end__"];
export const readPath = ["__start__", "ingest_turn", "recover_active_intent", "classify_intent", "proposal_gate", "read_context", "finalizer", "__end__"];
