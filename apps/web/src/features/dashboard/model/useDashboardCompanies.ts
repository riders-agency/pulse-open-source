"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import { asArray, type Company } from "../../../lib/api";
import { usePulseStore } from "../../../store/pulse-store";
import { listDashboardCompanies } from "../api/dashboardApi";
import { messageFromError } from "./errors";

export function useDashboardCompanies(token?: string) {
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const companiesRaw = usePulseStore((s) => s.companies);
  const setCompanies = usePulseStore((s) => s.setCompanies);
  const [error, setError] = useState("");
  const companies = useMemo(() => asArray<Company>(companiesRaw), [companiesRaw]);
  const activeCompany = useMemo(
    () => companies.find((company) => company.company_id === activeCompanyId),
    [activeCompanyId, companies],
  );

  const loadCompanies = useCallback(async () => {
    if (!token) return;
    const safe = await listDashboardCompanies(token);
    setCompanies(safe);
    if (!activeCompanyId && safe[0]?.company_id) setActiveCompany(safe[0].company_id);
  }, [activeCompanyId, setActiveCompany, setCompanies, token]);

  useEffect(() => {
    void loadCompanies().catch((err) => setError(messageFromError(err, "Не удалось загрузить компании")));
  }, [loadCompanies]);

  return {
    activeCompany,
    activeCompanyId,
    error,
  };
}
