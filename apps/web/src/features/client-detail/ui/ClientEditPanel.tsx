import { Save } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Client } from "../../../lib/api";
import { clientTypeLabel, statusLabel, statusTone } from "../../../lib/client-labels";
import type { ClientDetailState } from "../model/useClientDetailState";

export function ClientEditPanel({ state }: { state: ClientDetailState }) {
  const edit = state.edit;
  return (
    <section className="section-panel" id="client-responsibles">
      <header>
        <div>
          <h2>Реквизиты</h2>
          <p>{state.client ? clientTypeLabel(state.client.client_type) : "Загрузка"}</p>
        </div>
        <Badge tone={statusTone(state.client?.status)}>{statusLabel(state.client?.status)}</Badge>
      </header>
      {edit ? (
        <div className="section-panel-body stack">
          <div className="form-grid">
            <Field label="Тип">
              <Select value={edit.client_type} onChange={(event) => state.setEdit({ ...edit, client_type: event.target.value as Client["client_type"] })}>
                <option value="legal_entity">Юрлицо</option>
                <option value="individual_entrepreneur">ИП</option>
                <option value="individual">Физлицо</option>
              </Select>
            </Field>
            <Field label="Статус">
              <Select value={edit.status} onChange={(event) => state.setEdit({ ...edit, status: event.target.value })}>
                <option value="active">Активно</option>
                <option value="archived">В архиве</option>
              </Select>
            </Field>
            <Field label="Название" className="span-2">
              <Input value={edit.name} onChange={(event) => state.setEdit({ ...edit, name: event.target.value })} />
            </Field>
            <Field label="Короткое название">
              <Input value={edit.short_name} onChange={(event) => state.setEdit({ ...edit, short_name: event.target.value })} />
            </Field>
            <Field label="Полное название">
              <Input value={edit.full_name} onChange={(event) => state.setEdit({ ...edit, full_name: event.target.value })} />
            </Field>
            <Field label="ИНН">
              <Input value={edit.inn} onChange={(event) => state.setEdit({ ...edit, inn: event.target.value })} inputMode="numeric" />
            </Field>
            <Field label="КПП">
              <Input value={edit.kpp} onChange={(event) => state.setEdit({ ...edit, kpp: event.target.value })} inputMode="numeric" />
            </Field>
            <Field label="ОГРН">
              <Input value={edit.ogrn} onChange={(event) => state.setEdit({ ...edit, ogrn: event.target.value })} inputMode="numeric" />
            </Field>
            <Field label="ОГРНИП">
              <Input value={edit.ogrnip} onChange={(event) => state.setEdit({ ...edit, ogrnip: event.target.value })} inputMode="numeric" />
            </Field>
            <Field label="ОКВЭД">
              <Input value={edit.okved} onChange={(event) => state.setEdit({ ...edit, okved: event.target.value })} />
            </Field>
            <Field label="Юридический адрес" className="span-2">
              <Input value={edit.legal_address} onChange={(event) => state.setEdit({ ...edit, legal_address: event.target.value })} />
            </Field>
            <Field label="Фактический адрес" className="span-2">
              <Input value={edit.actual_address} onChange={(event) => state.setEdit({ ...edit, actual_address: event.target.value })} />
            </Field>
            <Field label="Комментарий" className="span-2">
              <Textarea value={edit.comment} onChange={(event) => state.setEdit({ ...edit, comment: event.target.value })} />
            </Field>
          </div>
          <div className="form-actions">
            <Button onClick={state.saveClient} disabled={state.busy || !edit.name.trim()}>
              <Save size={18} />
              <span>{state.busy ? "Сохраняем..." : "Сохранить"}</span>
            </Button>
          </div>
        </div>
      ) : null}
    </section>
  );
}
