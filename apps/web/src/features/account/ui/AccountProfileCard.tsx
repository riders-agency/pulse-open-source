"use client";

import Link from "next/link";

import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { Field, Input } from "../../../components/ui/input";
import type { useAccountSettings } from "../model/useAccountSettings";

type AccountSettings = ReturnType<typeof useAccountSettings>;

export function AccountProfileCard({ settings }: { settings: AccountSettings }) {
  const { busy, displayName, error, firstName, me, saveProfile, setDisplayName, setFirstName } = settings;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Профиль</CardTitle>
        <CardDescription>Адрес используется для входа в аккаунт.</CardDescription>
      </CardHeader>
      <CardContent className="stack">
        {me ? (
          <>
            <Field label="Электронная почта">
              <Input value={me.email} readOnly aria-readonly />
            </Field>
            <Field label="Отображаемое имя">
              <Input value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
            </Field>
            <Field label="Имя">
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </Field>
          </>
        ) : (
          <p className="muted">Загрузка…</p>
        )}
        {error ? <div className="form-error">{error}</div> : null}
        <div className="row">
          <Button type="button" onClick={saveProfile} disabled={busy}>
            Сохранить
          </Button>
          <Link className="ui-button ui-button--secondary ui-button--md" href="/dashboard">
            На обзор
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}
