"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import type { Company } from "../../lib/api";
import { getCompany } from "./api/companiesApi";
import { usePulseStore } from "../../store/pulse-store";

export default function CompanyPage() {
  const { companyId } = useParams<{ companyId: string }>();
  const token = usePulseStore((s) => s.token);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const [company, setCompany] = useState<Company | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token) return;
    getCompany(companyId, token)
      .then((row) => {
        setCompany(row);
        setActiveCompany(row.company_id);
      })
      .catch((err) => setError(err.message));
  }, [companyId, setActiveCompany, token]);

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>{company?.name ?? "Компания"}</h1>
          <p className="muted">Обзор компании без редактирования проектов на этом экране.</p>
        </div>
        <Link href="/projects/new">
          <Button>Создать проект</Button>
        </Link>
      </div>
      {error ? <Badge tone="danger">{error}</Badge> : null}
      <section className="grid-2">
        <Card>
          <CardHeader>
            <CardTitle>Данные компании</CardTitle>
            <CardDescription>Базовая информация и роль текущего пользователя.</CardDescription>
          </CardHeader>
          <CardContent className="stack">
            <div className="between"><span className="muted">ID</span><span className="mono">{companyId}</span></div>
            <div className="between"><span className="muted">Короткий адрес</span><span>{company?.slug || "—"}</span></div>
            <div className="between"><span className="muted">Роль</span><Badge>{roleLabel(company?.role)}</Badge></div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Следующий шаг</CardTitle>
            <CardDescription>Проекты компании находятся на отдельном экране.</CardDescription>
          </CardHeader>
          <CardContent className="row">
            <Link href="/clients">
              <Button variant="secondary">Открыть клиентов</Button>
            </Link>
            <Link href="/projects">
              <Button variant="secondary">Открыть проекты</Button>
            </Link>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}

function roleLabel(role?: string) {
  if (role === "owner") return "Владелец";
  if (role === "admin") return "Администратор";
  return "Участник";
}
