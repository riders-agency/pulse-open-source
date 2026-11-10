"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Field, Input, Textarea } from "../../components/ui/input";
import {
  getProjectIntegration,
  rotateProjectIntegrationSecret,
  updateProjectIntegration,
} from "./api/projectIntegrationsApi";
import {
  editIntegrationMetaHint,
  editProjectIntegrationTypes,
  projectIntegrationSecretPayload,
  projectIntegrationUpdatePayload,
} from "./model/projectIntegrationsModel";
import { usePulseStore } from "../../store/pulse-store";

const types = editProjectIntegrationTypes;

export default function EditIntegrationPage() {
  const params = useParams<{ projectId?: string; integrationId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const activeProjectIntegrationId = usePulseStore((s) => s.activeProjectIntegrationId);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setActiveProjectIntegration = usePulseStore((s) => s.setActiveProjectIntegration);
  const [name, setName] = useState("");
  const [typeKey, setTypeKey] = useState("");
  const [secret, setSecret] = useState("");
  const [metadata, setMetadata] = useState("{}");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const projectId = params.projectId || activeProjectId;
  const integrationId = params.integrationId || activeProjectIntegrationId;

  const selected = useMemo(
    () => types.find((t) => t.key === typeKey) ?? { key: typeKey, label: typeKey, secret: "", secretHint: "", showMetadata: true },
    [typeKey],
  );

  const load = useCallback(() => {
    if (!token) return;
    if (!projectId || !integrationId) {
      setLoading(false);
      return;
    }
    if (params.projectId) setActiveProject(projectId);
    if (params.integrationId) setActiveProjectIntegration(integrationId);
    setLoading(true);
    setError("");
    getProjectIntegration(projectId, integrationId, token)
      .then(({ row, metadataText }) => {
        setName(row.name);
        setTypeKey(row.dependency_type_key);
        setMetadata(metadataText);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  }, [integrationId, params.integrationId, params.projectId, projectId, setActiveProject, setActiveProjectIntegration, token]);

  useEffect(() => {
    load();
  }, [load]);

  const save = async () => {
    if (!token || !projectId || !integrationId) return;
    const nm = name.trim();
    if (!nm) {
      setError("Укажите название");
      return;
    }
    let updatePayload: Record<string, unknown>;
    try {
      updatePayload = projectIntegrationUpdatePayload({
        name: nm,
        metadata,
        showMetadata: selected.showMetadata,
      });
    } catch {
      setError("Некорректный JSON в настройках");
      return;
    }

    try {
      setError("");
      await updateProjectIntegration(projectId, integrationId, token, updatePayload);

      const newSecret = secret.trim();
      if (newSecret) {
        await rotateProjectIntegrationSecret(projectId, integrationId, token, projectIntegrationSecretPayload(newSecret));
      }

      router.push("/project-integrations");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить");
    }
  };

  if (loading) {
    return (
      <main className="page page-narrow">
        <div className="muted">Загрузка интеграции…</div>
      </main>
    );
  }

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>Изменить интеграцию</h1>
          <p className="muted">
            Обновление метаданных и необязательная смена секрета. Текущий секрет недоступен для просмотра — можно только записать новый.
          </p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>{name || "Интеграция"}</CardTitle>
          <CardDescription>Тип: {selected.label}</CardDescription>
        </CardHeader>
        <CardContent className="stack">
          <Field label="Название">
            <Input value={name} onChange={(event) => setName(event.target.value)} />
          </Field>
          {selected.secret ? (
            <Field
              label={typeKey.startsWith("telegram.") ? "Новый токен бота" : typeKey === "yclients.account" ? "Новый профиль доступа JSON" : typeKey.startsWith("ai.") ? "Новый API-ключ" : selected.secret}
              hint={
                typeKey === "telegram.bot"
                  ? "Пусто — токен не меняем. Для Telegram больше ничего на этой странице не нужно."
                  : typeKey === "telegram.relay"
                    ? selected.secretHint
                  : typeKey.startsWith("ai.") || typeKey === "network.vless_proxy"
                    ? selected.secretHint
                  : selected.secretHint || "Оставьте пустым, если токен/пароль менять не нужно."
              }
            >
              <Input value={secret} type="password" onChange={(event) => setSecret(event.target.value)} />
            </Field>
          ) : null}
          {selected.showMetadata ? (
            <Field label="Настройки JSON" hint={editIntegrationMetaHint(typeKey)}>
              <Textarea value={metadata} onChange={(event) => setMetadata(event.target.value)} />
            </Field>
          ) : null}
          {error ? <p className="status-failed">{error}</p> : null}
          <div className="form-actions">
            <Button variant="outline" onClick={() => router.back()}>Отмена</Button>
            <Button onClick={save} disabled={!token || !integrationId}>Сохранить</Button>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
