"use client";

import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Field, Input } from "../../components/ui/input";
import { createProjectWorkflow } from "./api/projectWorkflowsApi";
import { slugify } from "../../lib/utils";
import { usePulseStore } from "../../store/pulse-store";

export default function NewWorkflowPage() {
  const params = useParams<{ projectId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugTouched, setSlugTouched] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const projectId = params.projectId || activeProjectId;

  const create = async () => {
    setError("");
    setBusy(true);
    try {
      const workflow = await createProjectWorkflow(projectId, token, {
        name: name.trim(),
        slug: slugify(slug || name),
      });
      router.push(`/workflows/${workflow.workflow_id}/edit`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать сценарий");
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>Создать сценарий</h1>
          <p className="muted">Здесь только название. Сам сценарий собирается в визуальном редакторе.</p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Новый сценарий</CardTitle>
          <CardDescription>После создания откроется визуальный редактор.</CardDescription>
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
              placeholder="Например: Telegram бот поддержки"
            />
          </Field>
          <Field label="Короткий адрес">
            <Input
              value={slug}
              onChange={(event) => {
                setSlugTouched(true);
                setSlug(slugify(event.target.value));
              }}
              placeholder="telegram-support-bot"
            />
          </Field>
          {error ? <p className="status-failed">{error}</p> : null}
          <div className="form-actions">
            <Button variant="outline" onClick={() => router.back()}>Отмена</Button>
            <Button onClick={create} disabled={busy || !token || !projectId || !name.trim() || !slugify(slug || name)}>
              {busy ? "Создаем..." : "Создать сценарий"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
