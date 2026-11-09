"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { Table } from "../../components/ui/table";
import { asArray, type Company } from "../../lib/api";
import { listCompanies } from "./api/companiesApi";
import { usePulseStore } from "../../store/pulse-store";

export default function CompaniesPage() {
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const companies = usePulseStore((s) => s.companies);
  const setCompanies = usePulseStore((s) => s.setCompanies);
  const [error, setError] = useState("");
  const companyRows = useMemo(() => asArray<Company>(companies), [companies]);

  const selectCompany = (companyId: string) => {
    setActiveCompany(companyId);
    router.push("/clients");
  };

  useEffect(() => {
    if (!token) return;
    listCompanies(token)
      .then((rows) => setCompanies(rows))
      .catch((err) => setError(err.message));
  }, [setCompanies, token]);

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Компании</h1>
          <p className="muted">Компания — верхний уровень доступа. Проекты создаются внутри выбранной компании.</p>
        </div>
        <Link href="/companies/new">
          <Button>Создать компанию</Button>
        </Link>
      </div>

      {error ? <Badge tone="danger">{error}</Badge> : null}

      <Card>
        <CardHeader>
          <CardTitle>Список компаний</CardTitle>
          <CardDescription>Выберите активную компанию для работы с проектами и сценариями.</CardDescription>
        </CardHeader>
        <CardContent>
          {companyRows.length === 0 ? (
            <div className="empty-state">Компаний пока нет</div>
          ) : (
            <Table>
              <thead>
                <tr>
                  <th>Название</th>
                  <th>Короткий адрес</th>
                  <th>Роль</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {companyRows.map((company) => (
                  <tr className={company.company_id === activeCompanyId ? "is-active" : ""} key={company.company_id}>
                    <td>{company.name}</td>
                    <td className="mono">{company.slug || "—"}</td>
                    <td><Badge>{roleLabel(company.role)}</Badge></td>
                    <td className="row">
                      <Button
                        variant={company.company_id === activeCompanyId ? "secondary" : "outline"}
                        size="sm"
                        onClick={() => selectCompany(company.company_id)}
                      >
                        {company.company_id === activeCompanyId ? "Активна" : "Выбрать"}
                      </Button>
                      <Link href="/clients" onClick={() => setActiveCompany(company.company_id)}>
                        <Button variant="ghost" size="sm">Работать</Button>
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          )}
        </CardContent>
      </Card>
    </main>
  );
}

function roleLabel(role?: string) {
  if (role === "owner") return "Владелец";
  if (role === "admin") return "Администратор";
  return "Участник";
}
