import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import type { DashboardModel } from "../model/useDashboardModel";
import { DemoLoginLink } from "../../demo/DemoLoginLink";

type DashboardLoginCardProps = {
  model: Pick<DashboardModel, "email" | "login" | "password" | "setEmail" | "setPassword">;
};

export function DashboardLoginCard({ model }: DashboardLoginCardProps) {
  return (
    <Card className="page-narrow">
      <CardHeader>
        <CardTitle>Вход</CardTitle>
        <CardDescription>Войдите в своё рабочее пространство или попробуйте демо.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <Field label="Электронная почта">
          <Input value={model.email} onChange={(event) => model.setEmail(event.target.value)} />
        </Field>
        <Field label="Пароль">
          <Input value={model.password} type="password" onChange={(event) => model.setPassword(event.target.value)} />
        </Field>
        <div className="form-actions">
          <Button onClick={model.login}>Войти</Button>
        </div>
        <DemoLoginLink />
      </CardContent>
    </Card>
  );
}
