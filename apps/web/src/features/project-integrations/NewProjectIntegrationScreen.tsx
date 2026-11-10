"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Field, Input, Textarea } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import { createProjectIntegration } from "./api/projectIntegrationsApi";
import {
  metadataPreset,
  newIntegrationMetaHint,
  newProjectIntegrationTypes,
  projectIntegrationCreatePayload,
} from "./model/projectIntegrationsModel";
import { usePulseStore } from "../../store/pulse-store";

const types = newProjectIntegrationTypes;

export default function NewIntegrationPage() {
  const params = useParams<{ projectId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const [name, setName] = useState("Основная интеграция");
  const [typeKey, setTypeKey] = useState("telegram.bot");
  const [secret, setSecret] = useState("");
  const [metadata, setMetadata] = useState(() => metadataPreset("telegram.bot"));
  const [error, setError] = useState("");
  const selected = useMemo(() => types.find((type) => type.key === typeKey) ?? types[0], [typeKey]);
  const projectId = params.projectId || activeProjectId;

  useEffect(() => {
    const provider = new URLSearchParams(window.location.search).get("provider");
    if (provider && types.some((type) => type.key === provider)) setTypeKey(provider);
  }, []);

  useEffect(() => {
    setMetadata(metadataPreset(typeKey));
  }, [typeKey]);

  const create = async () => {
    setError("");
    try {
      const trimmedName = name.trim();
      const trimmedSecret = secret.trim();
      if (!trimmedName) {
        setError("Укажите название");
        return;
      }
      if (selected.secret && !trimmedSecret) {
        setError(typeKey.startsWith("telegram.") ? "Введите токен бота Telegram" : typeKey.startsWith("ai.") ? "Введите API-ключ провайдера модели" : "Введите секрет интеграции");
        return;
      }
      await createProjectIntegration(
        projectId,
        token,
        projectIntegrationCreatePayload({
          name: trimmedName,
          typeKey,
          secret: trimmedSecret,
          metadata,
          showMetadata: selected.showMetadata,
        }),
      );
      router.push("/project-integrations");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось создать интеграцию");
    }
  };

  return (
    <main className="page page-narrow">
      <div className="page-title">
        <div>
          <h1>Добавить интеграцию</h1>
          <p className="muted">Один экран — одна интеграция. Секреты сохраняются зашифрованно и не возвращаются в UI.</p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Новая интеграция</CardTitle>
          <CardDescription>Выберите тип и заполните параметры провайдера.</CardDescription>
        </CardHeader>
        <CardContent className="stack">
          <Field label="Тип">
            <Select value={typeKey} onChange={(event) => setTypeKey(event.target.value)}>
              {types.map((type) => (
                <option value={type.key} key={type.key}>{type.label}</option>
              ))}
            </Select>
          </Field>
          <Field label="Название">
            <Input value={name} onChange={(event) => setName(event.target.value)} />
          </Field>
          {selected.secret ? (
            <Field
              label={typeKey.startsWith("telegram.") ? "Токен бота" : typeKey === "yclients.account" ? "Зашифрованный профиль доступа JSON" : typeKey === "dadata.party" || typeKey.startsWith("ai.") ? "API-ключ" : typeKey === "network.vless_proxy" ? "VLESS URL" : "Пароль"}
              hint={
                typeKey === "telegram.bot"
                  ? "Вставьте bot token из BotFather. Больше для обычной отправки Telegram здесь ничего не нужно."
                  : typeKey === "telegram.relay"
                    ? "Вставьте bot token из BotFather. Relay будет принимать updates через HTTPS webhook и проверять secret_token."
                    : typeKey === "dadata.party"
                      ? "Вставьте API-ключ DaData. Он будет использоваться в создании клиентов этого проекта."
                      : typeKey === "yclients.account"
                        ? "JSON: partner_token, user_token либо login/password fallback, receipt_hmac_secret; optional receipt_basic_username/receipt_basic_password только для закрытого stage callback. Секреты не попадут в metadata и не вернутся в UI."
                      : typeKey.startsWith("ai.")
                        ? "Вставьте API-ключ провайдера модели. Он хранится зашифрованно и не возвращается в интерфейс."
                        : typeKey === "network.vless_proxy"
                          ? "Вставьте vless:// ссылку. Она хранится зашифрованно и не используется как общий proxy."
                      : "Секрет не будет показан повторно. Для MAX это токен из business.max.ru → Чат-боты → Интеграция; для Яндекс-почты используйте пароль приложения (не основной пароль)."
              }
            >
              <Input value={secret} type="password" onChange={(event) => setSecret(event.target.value)} />
            </Field>
          ) : null}
          {selected.showMetadata ? (
            <Field label="Настройки JSON" hint={newIntegrationMetaHint(typeKey)}>
              <Textarea value={metadata} onChange={(event) => setMetadata(event.target.value)} />
            </Field>
          ) : null}
          {error ? <p className="status-failed">{error}</p> : null}
          <div className="form-actions">
            <Button variant="outline" onClick={() => router.back()}>Отмена</Button>
            <Button onClick={create} disabled={!token || !projectId || !name}>Сохранить интеграцию</Button>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
