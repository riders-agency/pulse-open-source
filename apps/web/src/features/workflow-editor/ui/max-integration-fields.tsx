import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { ActionConfig } from "../model/types";

export function MaxIntegrationFields({
  config,
  update,
}: {
  config: ActionConfig;
  update: (patch: Partial<ActionConfig>) => void;
}) {
  return (
    <>
      <Field label="user_id" hint="Для личного сообщения вместо chat_id. Значение должно быть integer int64 из MAX.">
        <Input value={config.user_id ?? ""} onChange={(event) => update({ user_id: event.target.value })} />
      </Field>
      <div className="grid-2">
        <Field label="Формат">
          <Select value={config.format ?? ""} onChange={(event) => update({ format: event.target.value })}>
            <option value="">Обычный текст</option><option value="markdown">Markdown</option><option value="html">HTML</option>
          </Select>
        </Field>
        <Field label="Уведомление">
          <Select value={config.notify === false ? "false" : "true"} onChange={(event) => update({ notify: event.target.value !== "false" })}>
            <option value="true">Да</option><option value="false">Нет</option>
          </Select>
        </Field>
      </div>
      <label className="row"><input type="checkbox" checked={config.disable_link_preview ?? false} onChange={(event) => update({ disable_link_preview: event.target.checked })} />disable_link_preview</label>
      <Field label="document_url" hint="Один PDF. Для набора используйте document_urls_json.">
        <Input value={config.document_url ?? ""} onChange={(event) => update({ document_url: event.target.value })} />
      </Field>
      <Field label="document_urls_json" hint='JSON-массив URL, например {{ webhook.fileUrlsJson }}.'>
        <Textarea value={config.document_urls_json ?? ""} onChange={(event) => update({ document_urls_json: event.target.value })} />
      </Field>
      <label className="row"><input type="checkbox" checked={config.document_required ?? false} onChange={(event) => update({ document_required: event.target.checked })} />document_required</label>
      <label className="row"><input type="checkbox" checked={config.failure_required ?? false} onChange={(event) => update({ failure_required: event.target.checked })} />failure_required</label>
    </>
  );
}
