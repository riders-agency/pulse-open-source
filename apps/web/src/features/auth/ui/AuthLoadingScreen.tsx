import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";

export function AuthLoadingScreen() {
  return (
    <main className="auth-only auth-screen" aria-label="Загрузка авторизации" data-testid="auth-screen">
      <Card className="auth-card" data-testid="auth-card">
        <CardHeader>
          <div className="auth-brand">
            <span className="auth-brand-mark">RP</span>
            <div>
              <CardTitle>Riders Pulse</CardTitle>
              <CardDescription>Загружаем сессию входа.</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="stack">
          <p className="muted">Если экран не обновился, перезагрузите страницу или очистите сессию входа.</p>
        </CardContent>
      </Card>
    </main>
  );
}
