import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { PaymentAction } from "../../../lib/api";
import { paymentMethodLabel, paymentSourceLabel } from "../../../lib/client-labels";
import { dateInputProps } from "../../../lib/dates";

export function PaymentConfirmCard({
  action,
  amount,
  busy,
  nextDueDate,
  onConfirm,
  paidAt,
  paymentMethod,
  setAmount,
  setNextDueDate,
  setPaidAt,
  setPaymentMethod,
}: {
  action: PaymentAction | null;
  amount: string;
  busy: boolean;
  nextDueDate: string;
  onConfirm: () => Promise<void>;
  paidAt: string;
  paymentMethod: string;
  setAmount: (value: string) => void;
  setNextDueDate: (value: string) => void;
  setPaidAt: (value: string) => void;
  setPaymentMethod: (value: string) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Счет оплачен</CardTitle>
        <CardDescription>Отметьте поступление денег по этому счету. Следующая дата по услугам сдвинется автоматически.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <div className="payment-source-line">
          <span>{paymentSourceLabel(action?.source)}</span>
          <span>{paymentMethodLabel(action?.payment_method)}</span>
        </div>
        <Field label="Дата оплаты">
          <Input {...dateInputProps} value={paidAt} onChange={(event) => setPaidAt(event.target.value)} />
        </Field>
        <Field label="Сумма">
          <Input inputMode="decimal" value={amount} onChange={(event) => setAmount(event.target.value)} />
        </Field>
        <Field label="Способ оплаты">
          <Select value={paymentMethod} onChange={(event) => setPaymentMethod(event.target.value)}>
            <option value="legal_invoice">По счету с юрлица</option>
            <option value="cash">Наличные</option>
            <option value="bank_transfer">Банковский перевод</option>
            <option value="card">Карта</option>
            <option value="other">Другое</option>
          </Select>
        </Field>
        <Field label="Следующая дата платежа">
          <Input {...dateInputProps} value={nextDueDate} onChange={(event) => setNextDueDate(event.target.value)} />
        </Field>
        <Button onClick={onConfirm} disabled={busy || action?.status === "paid"}>
          {busy ? "Сохраняем..." : "Счет оплачен"}
        </Button>
      </CardContent>
    </Card>
  );
}
