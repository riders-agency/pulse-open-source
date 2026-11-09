import { asRecord, jsonPreview, recordList, textValue } from "../aiEmployeeUtils";

const operationLabels: Record<string, string> = {
  add_node: "Добавить действие", add_action: "Добавить действие", add_edge: "Добавить связь",
  update_node: "Изменить действие", update_action: "Изменить действие", remove_node: "Удалить действие",
  remove_edge: "Удалить связь", add_status: "Добавить статус", add_transition: "Добавить переход",
};

export function ToolProposalFields({ input }: { input: Record<string, unknown> }) {
  const patch = asRecord(input.patch_json);
  const operations = recordList(patch.operations);
  return (
    <div className="ai-tool-proposal-fields">
      {textValue(input.summary) ? <p>{textValue(input.summary)}</p> : null}
      {operations.length ? <ol>{operations.map((operation, index) => (
        <li key={index}><strong>{operationLabels[textValue(operation.op)] || textValue(operation.op) || "Изменение"}</strong><pre>{jsonPreview(operation)}</pre></li>
      ))}</ol> : <pre>{jsonPreview(input)}</pre>}
      {input.workflow_id ? <p>Изменения сохранятся в черновик. Для запуска потребуется публикация сценария.</p> : null}
    </div>
  );
}
