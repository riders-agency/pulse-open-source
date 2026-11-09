import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";

export function AuthVerifyingScreen({
  loginAccepted,
  onRetry,
  onReturnToLogin,
  verifyError,
}: {
  loginAccepted: boolean;
  onRetry: () => void;
  onReturnToLogin: () => void;
  verifyError: string;
}) {
  return (
    <main className="auth-only auth-screen" data-testid="auth-screen">
      <Card className="auth-card" data-testid="auth-card">
        <CardHeader>
          <div className="auth-brand">
            <span className="auth-brand-mark">RP</span>
            <div>
              <CardTitle>Riders Pulse</CardTitle>
              <CardDescription>
                {verifyError ? "Не удалось проверить сессию." : loginAccepted ? "Логин и пароль верные. Открываем Riders Pulse." : "Проверяем сохраненный вход."}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="stack">
          {verifyError ? (
            <div className="form-error">{verifyError}</div>
          ) : (
            <p className="muted">
              {loginAccepted ? "Данные входа подтверждены, проверяем сессию браузера." : "Если сохраненная сессия истекла, форма входа откроется автоматически."}
            </p>
          )}
          <div className="form-actions">
            <Button type="button" onClick={onRetry}>
              Проверить еще раз
            </Button>
            <Button type="button" variant="outline" onClick={onReturnToLogin}>
              Вернуться ко входу
            </Button>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
