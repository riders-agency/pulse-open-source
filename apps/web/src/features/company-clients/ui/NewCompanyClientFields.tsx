"use client";

import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { Client } from "../../../lib/api";
import type { CompanyClientForm } from "../model/companyPartyForms";

export function NewCompanyClientFields({
  form,
  setForm,
}: {
  form: CompanyClientForm;
  setForm: (form: CompanyClientForm) => void;
}) {
  return (
    <div className="form-grid">
      <Field label="Тип">
        <Select value={form.client_type} onChange={(event) => setForm({ ...form, client_type: event.target.value as Client["client_type"] })}>
          <option value="legal_entity">Юрлицо</option>
          <option value="individual_entrepreneur">ИП</option>
          <option value="individual">Физлицо</option>
        </Select>
      </Field>
      <Field label="Название">
        <Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
      </Field>
      <Field label="Короткое название">
        <Input value={form.short_name} onChange={(event) => setForm({ ...form, short_name: event.target.value })} />
      </Field>
      <Field label="Полное название">
        <Input value={form.full_name} onChange={(event) => setForm({ ...form, full_name: event.target.value })} />
      </Field>
      <Field label="ИНН">
        <Input value={form.inn} onChange={(event) => setForm({ ...form, inn: event.target.value })} inputMode="numeric" />
      </Field>
      <Field label="КПП">
        <Input value={form.kpp} onChange={(event) => setForm({ ...form, kpp: event.target.value })} inputMode="numeric" />
      </Field>
      <Field label="ОГРН">
        <Input value={form.ogrn} onChange={(event) => setForm({ ...form, ogrn: event.target.value })} inputMode="numeric" />
      </Field>
      <Field label="ОГРНИП">
        <Input value={form.ogrnip} onChange={(event) => setForm({ ...form, ogrnip: event.target.value })} inputMode="numeric" />
      </Field>
      <Field label="ОКВЭД">
        <Input value={form.okved} onChange={(event) => setForm({ ...form, okved: event.target.value })} />
      </Field>
      <Field label="Юридический адрес" className="span-2">
        <Input value={form.legal_address} onChange={(event) => setForm({ ...form, legal_address: event.target.value })} />
      </Field>
      <Field label="Фактический адрес" className="span-2">
        <Input value={form.actual_address} onChange={(event) => setForm({ ...form, actual_address: event.target.value })} />
      </Field>
      <Field label="Комментарий" className="span-2">
        <Textarea value={form.comment} onChange={(event) => setForm({ ...form, comment: event.target.value })} />
      </Field>
    </div>
  );
}
