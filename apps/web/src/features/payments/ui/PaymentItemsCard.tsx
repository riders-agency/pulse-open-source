import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Table } from "../../../components/ui/table";
import type { PaymentAction } from "../../../lib/api";
import { formatDate } from "../../../lib/dates";
import { formatMoney } from "../model/paymentsModel";

export function PaymentItemsCard({ action }: { action: PaymentAction | null }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Что входит в счет</CardTitle>
        <CardDescription>Услуги, по которым клиент должен оплатить.</CardDescription>
      </CardHeader>
      <CardContent>
        {!action?.items?.length ? (
          <div className="empty-state">Строки счета не загружены</div>
        ) : (
          <Table>
            <thead>
              <tr><th>Услуга</th><th>Дата</th><th>Сумма</th></tr>
            </thead>
            <tbody>
              {action.items.map((item) => (
                <tr key={item.payment_action_item_id}>
                  <td>{item.title}</td>
                  <td>{formatDate(item.due_date)}</td>
                  <td>{formatMoney(item.amount)}</td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
