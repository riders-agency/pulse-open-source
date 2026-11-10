"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Field, Input } from "../../components/ui/input";
import { createCompany, listCompanies } from "./api/companiesApi";
import { slugify } from "../../lib/utils";
import { usePulseStore } from "../../store/pulse-store";

export default function NewCompanyPage() {
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setCompanies = usePulseStore((s) => s.setCompanies);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugTouched, setSlugTouched] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const create = async () => {
    setError("");
    setBusy(true);
    try {
      const payload = { name: name.trim(), slug: slugify(slug || name) };
      const out = await createCompany(token, payload);
      setActiveCompany(out.company_id);
      const rows = await listCompanies(token).catch(() => []);
      setCompanies(rows);
      router.push("/projects");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать компанию");
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>Создать компанию</h1>
          <p className="muted">Один экран — одна задача. Здесь только создание компании.</p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Новая компания</CardTitle>
          <CardDescription>После создания вы сможете добавить проекты.</CardDescription>
        </CardHeader>
        <CardContent className="stack">
          <Field label="Название">
            <Input
              value={name}
              onChange={(event) => {
                const nextName = event.target.value;
                setName(nextName);
                if (!slugTouched) setSlug(slugify(nextName));
              }}
              placeholder="Например: Riders Agency"
            />
          </Field>
          <Field label="Короткий адрес">
            <Input
              value={slug}
              onChange={(event) => {
                setSlugTouched(true);
                setSlug(slugify(event.target.value));
              }}
              placeholder="riders-agency"
            />
          </Field>
          {error ? <p className="status-failed">{error}</p> : null}
          <div className="form-actions">
            <Button variant="outline" onClick={() => router.back()}>Отмена</Button>
            <Button onClick={create} disabled={busy || !token || !name.trim() || !slugify(slug || name)}>
              {busy ? "Создаём..." : "Создать компанию"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
