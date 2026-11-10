"use client";

import { useRouter, useParams } from "next/navigation";
import { useState } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Field, Input } from "../../components/ui/input";
import { createCompanyProject } from "./api/companiesApi";
import { slugify } from "../../lib/utils";
import { usePulseStore } from "../../store/pulse-store";

export default function NewProjectPage() {
  const params = useParams<{ companyId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const [name, setName] = useState("Orders Pulse");
  const [slug, setSlug] = useState("orders-pulse");
  const [secret, setSecret] = useState("");
  const [error, setError] = useState("");
  const companyId = params.companyId || activeCompanyId;

  const create = async () => {
    setError("");
    try {
      const project = await createCompanyProject(companyId, token, { name, slug });
      setActiveCompany(companyId);
      setActiveProject(project.project_id);
      setSecret(project.hook_secret || "");
      router.push("/project");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать проект");
    }
  };

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>Создать проект</h1>
          <p className="muted">Проект содержит сценарии, интеграции, входящие запросы и запуски.</p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Новый проект</CardTitle>
          <CardDescription>После создания откроется обзор проекта.</CardDescription>
        </CardHeader>
        <CardContent className="stack">
          <Field label="Название">
            <Input
              value={name}
              onChange={(event) => {
                setName(event.target.value);
                setSlug(slugify(event.target.value));
              }}
            />
          </Field>
          <Field label="Короткий адрес">
            <Input value={slug} onChange={(event) => setSlug(event.target.value)} />
          </Field>
          {secret ? <p className="code-block">Секрет входа показан один раз: {secret}</p> : null}
          {error ? <p className="status-failed">{error}</p> : null}
          <div className="form-actions">
            <Button variant="outline" onClick={() => router.back()}>Отмена</Button>
            <Button onClick={create} disabled={!token || !companyId || !name || !slug}>Создать проект</Button>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
