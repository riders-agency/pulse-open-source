"use client";

import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import { Table } from "../../../components/ui/table";
import { formatDate } from "../../../lib/dates";
import type { useAccountSettings } from "../model/useAccountSettings";

type AccountSettings = ReturnType<typeof useAccountSettings>;

export function AccountApiKeysCard({ settings }: { settings: AccountSettings }) {
  const {
    apiKeys,
    busy,
    createAPIKey,
    keyName,
    newAPIKey,
    permission,
    scopeID,
    scopeKind,
    setKeyName,
    setScopeID,
    updatePermission,
    updateScopeKind,
  } = settings;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Ключи API</CardTitle>
        <CardDescription>Ключи для внешних интеграций. Доступ задаётся grants на конкретные сущности.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        <div className="grid-2">
          <Field label="Название">
            <Input value={keyName} onChange={(event) => setKeyName(event.target.value)} />
          </Field>
          <Field label="Область доступа">
            <Select value={scopeKind} onChange={(event) => updateScopeKind(event.target.value)}>
              <option value="company">Компания</option>
              <option value="project">Проект</option>
              <option value="workflow">Сценарий</option>
              <option value="system">Система</option>
            </Select>
          </Field>
          {scopeKind !== "system" ? (
            <Field label="Идентификатор области">
              <Input value={scopeID} onChange={(event) => setScopeID(event.target.value)} />
            </Field>
          ) : null}
          <Field label="Права">
            <Select value={permission} onChange={(event) => updatePermission(event.target.value)}>
              <option value="read">Чтение</option>
              <option value="write">Изменение</option>
              <option value="admin">Управление</option>
            </Select>
          </Field>
        </div>
        <div className="row">
          <Button type="button" onClick={createAPIKey} disabled={busy || (scopeKind !== "system" && !scopeID)}>
            Выпустить API key
          </Button>
        </div>
        {newAPIKey ? (
          <div className="form-error">
            <strong>Скопируйте ключ сейчас:</strong>
            <br />
            <code>{newAPIKey}</code>
          </div>
        ) : null}
        {apiKeys.length === 0 ? <p className="muted">Активных API keys нет.</p> : <AccountApiKeysTable settings={settings} />}
      </CardContent>
    </Card>
  );
}

function AccountApiKeysTable({ settings }: { settings: AccountSettings }) {
  const { apiKeys, revokeAPIKey } = settings;

  return (
    <Table>
      <thead>
        <tr>
          <th>Ключ</th>
          <th>Права</th>
          <th>Создан</th>
          <th>Последнее использование</th>
          <th />
        </tr>
      </thead>
      <tbody>
        {apiKeys.map((key) => (
          <tr key={key.api_key_id}>
            <td>
              <strong>{key.name}</strong>
              <br />
              <small className="muted">rp_live_{key.key_prefix}_...</small>
            </td>
            <td>
              <small>{key.grants.map((grant) => `${grant.scope_kind}:${grant.scope_id || "*"} [${grant.permissions.join(",")}]`).join("; ")}</small>
            </td>
            <td><small>{formatDate(key.created_at)}</small></td>
            <td><small>{formatDate(key.last_used_at)}</small></td>
            <td>
              <Button type="button" variant="destructive" onClick={() => void revokeAPIKey(key.api_key_id)}>
                Отозвать
              </Button>
            </td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
