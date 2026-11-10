"use client";

import { useMemo } from "react";

import type { ClientResponsible, ClientService, PaymentAction } from "../../../lib/api";
import { compareDate, isBeforeDate, todayDateInput } from "../../../lib/dates";
import { filterPayments } from "./clientDetailStateHelpers";

type ProjectName = { name?: string; project_id: string };

export function useClientDetailDerived({
  activeProjectId,
  paymentFilter,
  payments,
  projects,
  responsibles,
  services,
}: {
  activeProjectId: string;
  paymentFilter: string;
  payments: PaymentAction[];
  projects: ProjectName[];
  responsibles: ClientResponsible[];
  services: ClientService[];
}) {
  const activeProjectName = useMemo(() => projects.find((project) => project.project_id === activeProjectId)?.name ?? "", [activeProjectId, projects]);
  const primaryResponsible = useMemo(() => responsibles.find((item) => item.is_primary) ?? responsibles[0], [responsibles]);
  const activeServices = useMemo(() => services.filter((item) => item.status === "active"), [services]);
  const recurringTotal = useMemo(() => activeServices.filter((item) => item.billing_period !== "once").reduce((sum, item) => sum + (item.amount || 0), 0), [activeServices]);
  const latestPayment = useMemo(
    () => [...payments].filter((item) => item.status === "paid").sort((a, b) => String(b.confirmed_at || b.created_at || b.due_date).localeCompare(String(a.confirmed_at || a.created_at || a.due_date)))[0],
    [payments],
  );
  const nextService = useMemo(() => [...activeServices].filter((item) => item.next_due_date).sort((a, b) => compareDate(a.next_due_date, b.next_due_date))[0], [activeServices]);
  const overdueServices = useMemo(() => activeServices.filter((item) => isBeforeDate(item.next_due_date, todayDateInput())), [activeServices]);
  const filteredPayments = useMemo(() => filterPayments(payments, paymentFilter), [paymentFilter, payments]);

  return {
    activeProjectName,
    activeServices,
    filteredPayments,
    latestPayment,
    nextService,
    overdueServices,
    primaryResponsible,
    recurringTotal,
  };
}
