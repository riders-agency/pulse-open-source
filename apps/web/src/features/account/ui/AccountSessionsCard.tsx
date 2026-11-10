"use client";

import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Table } from "../../../components/ui/table";
import { formatDate } from "../../../lib/dates";
import type { useAccountSettings } from "../model/useAccountSettings";

type AccountSettings = ReturnType<typeof useAccountSettings>;

export function AccountSessionsCard({ settings }: { settings: AccountSettings }) {
  const { endSession, logoutEverywhereCurrent, sessions } = settings;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Сессии</CardTitle>
        <CardDescription>
          Это устройства с входом под вашей учётной записью. Текущее устройство — завершите через «Выйти» в шапке.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {sessions.length === 0 ? (
          <p className="muted">Нет активных сессий или список ещё загружается.</p>
        ) : (
          <Table>
            <thead>
              <tr>
                <th>Устройство / IP</th>
                <th>Создана</th>
                <th>Действует</th>
                <th>Статус</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {sessions.map((session) => (
                <tr key={session.session_id}>
                  <td>
                    <small>{session.user_agent || "—"}</small>
                    <br />
                    <small className="muted">{session.ip || "—"}</small>
                  </td>
                  <td><small>{formatDate(session.created_at)}</small></td>
                  <td><small>{formatDate(session.expires_at)}</small></td>
                  <td>{session.current ? "Текущая" : "Активна"}</td>
                  <td>
                    {!session.current ? (
                      <Button type="button" variant="destructive" onClick={() => void endSession(session.session_id, false)}>
                        Завершить
                      </Button>
                    ) : (
                      <Button type="button" variant="secondary" onClick={() => void logoutEverywhereCurrent()}>
                        Выйти здесь
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
