import { useMemo } from "react";

import { Field, Input, Textarea } from "../../../components/ui/input";
import { emailDefaultsFromVariables } from "../registry/actions";
import type { ActionConfig } from "../model/types";

export function EmailIntegrationSettings({
  config,
  webhookSlug,
  templateVariables,
  update,
}: {
  config: ActionConfig;
  webhookSlug: string;
  templateVariables: string[];
  update: (patch: Partial<ActionConfig>) => void;
}) {
  const emailDefaults = useMemo(() => emailDefaultsFromVariables(templateVariables), [templateVariables]);

  return (
    <>
      <div className="template-help">
        <div>
          <strong>Переменные письма</strong>
          <p className="muted">Поля приходят из JSON входящего webhook. Префикс {webhookSlug} равен endpoint key триггера.</p>
        </div>
        <div className="row">
          {templateVariables.map((item) => (
            <code key={item}>{item}</code>
          ))}
        </div>
      </div>
      <Field label="to">
        <Input value={config.to ?? emailDefaults.to} onChange={(event) => update({ to: event.target.value })} />
      </Field>
      <Field label="Тема письма">
        <Input value={config.subject ?? emailDefaults.subject} onChange={(event) => update({ subject: event.target.value })} />
      </Field>
      <Field label="Текст письма">
        <Textarea value={config.body ?? emailDefaults.body} onChange={(event) => update({ body: event.target.value })} />
      </Field>
      <Field label="document_urls_json" hint='JSON-массив PDF URL, например {{ webhook.fileUrlsJson }}.'>
        <Textarea value={config.document_urls_json ?? ""} onChange={(event) => update({ document_urls_json: event.target.value })} />
      </Field>
      <label className="row">
        <input
          type="checkbox"
          checked={config.document_required ?? false}
          onChange={(event) => update({ document_required: event.target.checked })}
        />
        document_required
      </label>
      <label className="row">
        <input
          type="checkbox"
          checked={config.failure_required ?? false}
          onChange={(event) => update({ failure_required: event.target.checked })}
        />
        failure_required
      </label>
    </>
  );
}
