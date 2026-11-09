"use client";

import { Plus } from "lucide-react";

import { Button } from "../../../components/ui/button";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { CompanyLegalEntityForm } from "../model/companyPartyForms";

export function CompanyLegalEntityFields({
  busy,
  form,
  onCreate,
  setForm,
}: {
  busy: boolean;
  form: CompanyLegalEntityForm;
  onCreate: () => Promise<void>;
  setForm: (form: CompanyLegalEntityForm) => void;
}) {
  return (
    <>
      <Field label="Тип">
        <Select value={form.entity_type} onChange={(event) => setForm({ ...form, entity_type: event.target.value as CompanyLegalEntityForm["entity_type"] })}>
          <option value="legal_entity">Юрлицо</option>
          <option value="individual_entrepreneur">ИП</option>
        </Select>
      </Field>
      <Field label="Название">
        <Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
      </Field>
      <div className="grid-2">
        <Field label="ИНН">
          <Input value={form.inn} onChange={(event) => setForm({ ...form, inn: event.target.value })} inputMode="numeric" />
        </Field>
        <Field label="КПП">
          <Input value={form.kpp} onChange={(event) => setForm({ ...form, kpp: event.target.value })} inputMode="numeric" />
        </Field>
      </div>
      <Field label="Юридический адрес">
        <Input value={form.legal_address} onChange={(event) => setForm({ ...form, legal_address: event.target.value })} />
      </Field>
      <Field label="Комментарий">
        <Textarea value={form.comment} onChange={(event) => setForm({ ...form, comment: event.target.value })} />
      </Field>
      <Button onClick={onCreate} disabled={busy || !form.name.trim()}>
        <Plus size={18} />
        <span>Добавить</span>
      </Button>
    </>
  );
}
