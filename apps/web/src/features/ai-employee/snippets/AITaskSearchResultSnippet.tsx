import {
  taskSearchItems,
  textValue,
} from "../aiEmployeeUtils";

export function AITaskSearchResultSnippet({ payload }: { payload?: Record<string, unknown> }) {
  const searchItems = taskSearchItems(payload);
  if (!searchItems.length) return null;
  return (
    <div className="ai-result-card ai-search-result-card" data-testid="ai-task-search-result-card">
      <strong>Найденные задачи</strong>
      <div className="ai-result-list">
        {searchItems.map((task) => {
          const url = textValue(task.url) || `/tasks/${textValue(task.task_id)}`;
          const label = textValue(task.task_key) ? `${textValue(task.task_key)} · ${textValue(task.title)}` : textValue(task.title) || "Задача";
          return (
            <a className="ai-entity-link" href={url} key={url}>
              <span>{label}</span>
              {textValue(task.client_name) ? <em>{textValue(task.client_name)}</em> : null}
              {textValue(task.status_name) ? <small>{textValue(task.status_name)}</small> : null}
            </a>
          );
        })}
      </div>
    </div>
  );
}
