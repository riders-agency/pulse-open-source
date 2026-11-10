import Link from "next/link";
import type { FormEvent } from "react";

import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import type { AuthMode, AuthSubmitInput } from "../model/useAuthGateState";
import { DemoLoginLink } from "../../demo/DemoLoginLink";

export function AuthLoginScreen({
  ready,
  busy,
  displayName,
  email,
  error,
  mode,
  password,
  setDisplayName,
  setEmail,
  setMode,
  setPassword,
  submit,
}: {
  ready: boolean;
  busy: boolean;
  displayName: string;
  email: string;
  error: string;
  mode: AuthMode;
  password: string;
  setDisplayName: (value: string) => void;
  setEmail: (value: string) => void;
  setMode: (value: AuthMode) => void;
  setPassword: (value: string) => void;
  submit: (input?: AuthSubmitInput) => void;
}) {
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    submit({
      displayName: String(data.get("displayName") ?? ""),
      email: String(data.get("email") ?? ""),
      password: String(data.get("password") ?? ""),
    });
  };

  return (
    <main className="auth-only auth-screen" data-testid="auth-screen">
      <Card className="auth-card" data-testid="auth-card">
        <CardHeader>
          <div className="auth-brand">
            <span className="auth-brand-mark">RP</span>
            <div>
              <CardTitle>{mode === "login" ? "Вход в Riders Pulse" : "Создать аккаунт"}</CardTitle>
              <CardDescription>
                {mode === "login" ? "Войдите, чтобы работать с задачами, клиентами и командой." : "Создайте пользователя для рабочего пространства Riders Pulse."}
              </CardDescription>
            </div>
          </div>
          <div className="segmented" role="tablist" aria-label="Режим">
            <button type="button" disabled={!ready || busy} className={`segment ${mode === "login" ? "segment-active" : ""}`} role="tab" aria-selected={mode === "login"} onClick={() => setMode("login")}>
              Вход
            </button>
            <button type="button" disabled={!ready || busy} className={`segment ${mode === "register" ? "segment-active" : ""}`} role="tab" aria-selected={mode === "register"} onClick={() => setMode("register")}>
              Регистрация
            </button>
          </div>
        </CardHeader>
        <CardContent>
          <form className="stack" method="post" onSubmit={handleSubmit}>
            <Field label="Электронная почта">
              <Input disabled={!ready || busy} autoComplete="email" name="email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} />
            </Field>
            <Field label="Пароль">
              <Input
                disabled={!ready || busy}
                autoComplete={mode === "register" ? "new-password" : "current-password"}
                name="password"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </Field>
            {mode === "register" ? (
              <Field label="Имя (отображаемое)">
                <Input disabled={!ready || busy} autoComplete="name" name="displayName" value={displayName} onChange={(event) => setDisplayName(event.target.value)} placeholder="Необязательно" />
              </Field>
            ) : null}
            {error ? <div className="form-error">{error}</div> : null}
            <Button type="submit" disabled={!ready || busy}>
              {busy ? (mode === "login" ? "Проверяем логин и пароль..." : "Регистрация...") : mode === "login" ? "Войти" : "Создать аккаунт"}
            </Button>
            <p className="auth-footer-links">
              <Link href="/docs">Документация (без входа)</Link>
            </p>
          </form>
          <DemoLoginLink />
        </CardContent>
      </Card>
    </main>
  );
}
