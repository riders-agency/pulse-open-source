import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { useBankInvoice } from "../model/useBankInvoice";
import { dateInputProps } from "../../../lib/dates";

export function BankInvoiceCard({ id, token, paymentStatus }: { id: string; token: string; paymentStatus: string }) {
  const state = useBankInvoice(id, token);
  const { workspace, input } = state;
  const closed = !["pending", "sent", "snoozed"].includes(paymentStatus) || Boolean(workspace && !["pending", "sent", "snoozed"].includes(workspace.payment_status));
  const disabled = state.busy || closed || state.conflict || !workspace?.can_edit;
  return <Card>
    <CardHeader className="bank-invoice-header">
      <CardTitle>Счёт через банк <Badge tone="muted">Черновик</Badge></CardTitle>
      <CardDescription>Подготовьте данные сейчас. Выпуск счёта станет доступен после подключения банка.</CardDescription>
    </CardHeader>
    <CardContent>
      {state.error ? <p role="alert">{state.error}</p> : null}
      {!workspace ? state.error ? <Button variant="secondary" onClick={state.refresh}>Повторить загрузку</Button> : <p role="status">Загружаем черновик…</p> : <form className="stack" onSubmit={(event) => { event.preventDefault(); void state.save(); }}>
        <p><strong>{(Number(workspace.amount_minor) / 100).toLocaleString("ru-RU", { style: "currency", currency: workspace.currency })}</strong> · сумма из карточки оплаты</p>
        {closed ? <p role="status">Оплата закрыта. Черновик доступен для просмотра.</p> : null}
        {!workspace.can_edit ? <p>У вас доступ только для просмотра.</p> : null}
        <fieldset disabled={disabled} className="bank-invoice-fields">
          <Field label="Что подготовить"><select aria-label="Что подготовить" className="ui-input" value={input.kind} onChange={(event) => state.change("kind", event.target.value as typeof input.kind)}>
            <option value="bank_invoice">Счёт по реквизитам</option><option value="payment_link">Платёжная ссылка</option>
          </select></Field>
          <Field label="Плательщик"><Input required maxLength={240} value={input.payer_name} onChange={(event) => state.change("payer_name", event.target.value)} /></Field>
          <Field label="ИНН / налоговый номер" hint="Необязательно на этапе подготовки"><Input maxLength={32} value={input.tax_id} onChange={(event) => state.change("tax_id", event.target.value)} /></Field>
          <Field label="Назначение платежа"><Textarea aria-label="Назначение платежа" required maxLength={1000} rows={3} value={input.purpose} onChange={(event) => state.change("purpose", event.target.value)} /></Field>
          <Field label="Оплатить до" hint="Необязательно"><Input aria-label="Оплатить до" {...dateInputProps} value={input.due_date} onChange={(event) => state.change("due_date", event.target.value)} /></Field>
        </fieldset>
        <div className="bank-invoice-actions">
          <Button type="submit" disabled={disabled}>{state.busy ? "Сохраняем…" : "Сохранить черновик"}</Button>
          <Button type="button" variant="secondary" disabled>Выпустить в банке</Button>
        </div>
        {state.saved ? <p role="status">Черновик сохранён. Клиенту ничего не отправлено.</p> : null}
        {state.conflict ? <Button type="button" variant="secondary" onClick={state.refresh}>Загрузить актуальную версию</Button> : null}
        <p className="muted">Банк ещё не подключён. Реквизиты продавца и требования банка проверяются перед выпуском.</p>
      </form>}
    </CardContent>
  </Card>;
}
