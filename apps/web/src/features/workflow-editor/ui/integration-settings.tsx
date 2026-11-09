import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Integration } from "../../../lib/api";
import type { ActionConfig, ActionKind } from "../model/types";
import { EmailIntegrationSettings } from "./email-integration-settings";
import { MaxIntegrationFields } from "./max-integration-fields";

export function IntegrationSettings({
  integrations,
  type,
  config,
  webhookSlug,
  templateVariables,
  update,
}: {
  integrations: Integration[];
  type: ActionKind;
  config: ActionConfig;
  webhookSlug: string;
  templateVariables: string[];
  update: (patch: Partial<ActionConfig>) => void;
}) {
  const isMax = type === "max.send_message";
  const filtered = integrations.filter((item) => {
    if (item.status !== "active") return false;
    if (isMax) return item.dependency_type_key === "max.bot";
    if (type === "telegram.send_message") return item.dependency_type_key === "telegram.bot" || item.dependency_type_key === "telegram.relay";
    if (type === "telegram.send_document") return item.dependency_type_key === "telegram.bot" || item.dependency_type_key === "telegram.relay";
    if (type === "email.send") return item.dependency_type_key === "smtp.server";
    return true;
  });
  return (
    <>
      <Field label="Интеграция">
        <Select
          value={type === "email.send" ? config.smtp_action_dependency_id ?? "" : config.action_dependency_id ?? ""}
          onChange={(event) =>
            type === "email.send"
              ? update({ smtp_action_dependency_id: event.target.value })
              : update({ action_dependency_id: event.target.value })
          }
        >
          <option value="">Выберите интеграцию</option>
          {filtered.map((item) => (
            <option value={item.action_dependency_id} key={item.action_dependency_id}>{item.name}</option>
          ))}
        </Select>
      </Field>
      {type !== "email.send" ? (
        <>
          <Field
            label="chat_id"
            hint={
              isMax
                ? `ID чата MAX. Укажите chat_id или user_id, не оба. Поля {{ ${webhookSlug}.… }} берутся из JSON вебхука.`
                : undefined
            }
          >
            <Input value={config.chat_id ?? ""} onChange={(event) => update({ chat_id: event.target.value })} />
          </Field>
          {isMax ? (
            <MaxIntegrationFields config={config} update={update} />
          ) : null}
          {type === "telegram.send_document" ? (
            <>
              <Field label="document_url" hint="URL PDF или другого файла, напр. {{ ag-gift-certificate.pdf_url }}">
                <Input value={config.document_url ?? ""} onChange={(event) => update({ document_url: event.target.value })} />
              </Field>
              <Field label="Подпись">
                <Textarea value={config.caption ?? ""} onChange={(event) => update({ caption: event.target.value })} />
              </Field>
            </>
          ) : (
            <>
              <Field label="Текст">
                <Textarea value={config.text ?? ""} onChange={(event) => update({ text: event.target.value })} />
              </Field>
              {type === "telegram.send_message" ? (
                <Field label="reply_markup_json" hint="JSON объект Telegram inline keyboard. Поддерживает шаблоны.">
                  <Textarea value={config.reply_markup_json ?? ""} onChange={(event) => update({ reply_markup_json: event.target.value })} />
                </Field>
              ) : null}
            </>
          )}
        </>
      ) : (
        <EmailIntegrationSettings
          config={config}
          templateVariables={templateVariables}
          webhookSlug={webhookSlug}
          update={update}
        />
      )}
    </>
  );
}
