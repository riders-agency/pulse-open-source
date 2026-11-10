"use client";

import { useParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Field, Input, Textarea } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import { Button } from "../../components/ui/button";

export default function TemplatesPage() {
  const { projectId } = useParams<{ projectId: string }>();
  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Шаблоны</h1>
          <p className="muted">Шаблоны сообщений и писем используются действиями MAX, Telegram и электронной почты.</p>
        </div>
      </div>
      <section className="grid-2">
        <Card>
          <CardHeader>
            <CardTitle>Шаблон письма</CardTitle>
            <CardDescription>Форма подготовлена для API проекта {projectId}.</CardDescription>
          </CardHeader>
          <CardContent className="stack">
            <Field label="Название"><Input defaultValue="Order Paid" /></Field>
            <Field label="Тема"><Input defaultValue="Событие {{ orders_paid.id }} обработано" /></Field>
            <Field label="Формат"><Select defaultValue="html"><option value="html">HTML</option><option value="text">Обычный текст</option><option value="markdown">Markdown</option></Select></Field>
            <Field label="Тело"><Textarea defaultValue="Данные: {{ orders_paid.name }}\nЗапуск: {{ system.link }}" /></Field>
            <div className="form-actions"><Button disabled>Сохранить</Button></div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Проверка шаблонов</CardTitle>
            <CardDescription>Поддерживаемые выражения.</CardDescription>
          </CardHeader>
          <CardContent>
            <pre className="code-block">{`{{ orders_paid.id }}
{{ orders_paid.name }}
{{ steps.get_data.json.value }}
{{ system.link }}
{{ system.id }}`}</pre>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}
