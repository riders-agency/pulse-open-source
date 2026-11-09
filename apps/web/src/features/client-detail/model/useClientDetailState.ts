"use client";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import type { Client, ClientContact, ClientProject, ClientResponsible, ClientService, PaymentAction } from "../../../lib/api";
import { usePulseStore } from "../../../store/pulse-store";
import {
  deleteClient as deleteClientRecord,
  deleteClientService,
  getClientDetail,
  updateClient,
} from "../api/clientDetailApi";
import {
  emptyContactForm,
  emptyPaymentForm,
  emptyResponsibleForm,
  formFromClient,
  type EditForm,
} from "./clientDetailModel";
import { defaultPaymentForm, runClientAction } from "./clientDetailStateHelpers";
import { useClientDetailPayments } from "./useClientDetailPayments";
import { useClientDetailRelations } from "./useClientDetailRelations";
import { useCompanyClientContext } from "./useCompanyClientContext";
import { useClientDetailDerived } from "./useClientDetailDerived";
export function useClientDetailState() {
  const params = useParams<{ clientId?: string }>();
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = usePulseStore((s) => s.token);
  const projects = usePulseStore((s) => s.projects);
  const activeCompanyId = usePulseStore((s) => s.activeCompanyId);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const activeClientId = usePulseStore((s) => s.activeClientId);
  const setActiveCompany = usePulseStore((s) => s.setActiveCompany);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setActiveClient = usePulseStore((s) => s.setActiveClient);
  const [client, setClient] = useState<Client | null>(null);
  const [edit, setEdit] = useState<EditForm | null>(null);
  const [contacts, setContacts] = useState<ClientContact[]>([]);
  const [responsibles, setResponsibles] = useState<ClientResponsible[]>([]);
  const [links, setLinks] = useState<ClientProject[]>([]);
  const [services, setServices] = useState<ClientService[]>([]);
  const [payments, setPayments] = useState<PaymentAction[]>([]);
  const [paymentFilter, setPaymentFilter] = useState("all");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [contact, setContact] = useState(emptyContactForm);
  const [paymentForm, setPaymentForm] = useState(emptyPaymentForm);
  const [projectId, setProjectId] = useState("");
  const [responsible, setResponsible] = useState(emptyResponsibleForm);
  const clientId = params.clientId || activeClientId;

  const {
    activeProjectName,
    activeServices,
    filteredPayments,
    latestPayment,
    nextService,
    overdueServices,
    primaryResponsible,
    recurringTotal,
  } = useClientDetailDerived({ activeProjectId, paymentFilter, payments, projects, responsibles, services });
  const companyContext = useCompanyClientContext({ activeCompanyId, projects, token });
  const load = useCallback(() => {
    if (!token || !clientId) return;
    getClientDetail(clientId, activeProjectId, token)
      .then(({ client: row, contacts: contactRows, responsibles: responsibleRows, links: projectRows, services: serviceRows, payments: paymentRows }) => {
        setClient(row);
        setEdit(formFromClient(row));
        setContacts(contactRows);
        setResponsibles(responsibleRows);
        setLinks(projectRows);
        setServices(serviceRows);
        setPayments(paymentRows);
        setPaymentForm((current) => defaultPaymentForm(current, serviceRows));
        setActiveCompany(row.company_id);
        if (params.clientId) setActiveClient(row.client_id);
        setError("");
      })
      .catch((err) => setError(err.message));
  }, [activeProjectId, clientId, params.clientId, setActiveClient, setActiveCompany, token]);

  useEffect(load, [load]);
  const { addContact, addResponsible, deleteResponsible, linkProject } = useClientDetailRelations({
    clientId,
    contact,
    contactsCount: contacts.length,
    load,
    projectId,
    responsible,
    responsiblesCount: responsibles.length,
    setContact,
    setError,
    setProjectId,
    setResponsible,
    token,
  });

  const saveClient = async () => {
    if (!edit) return;
    setBusy(true);
    setError("");
    try {
      if (!clientId || !token) return;
      await updateClient(clientId, token, edit);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить клиента");
    } finally {
      setBusy(false);
    }
  };

  const deleteClient = async () => {
    if (!client || !window.confirm("Удалить клиента из активного списка?")) return;
    setBusy(true);
    setError("");
    try {
      if (!clientId || !token) return;
      await deleteClientRecord(clientId, token);
      setActiveClient("");
      router.push("/clients");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось удалить клиента");
      setBusy(false);
    }
  };

  const deleteService = async (id: string) => {
    if (!window.confirm("Удалить услугу?")) return;
    await runClientAction(async () => {
      if (!token) return;
      await deleteClientService(id, token);
      load();
    }, setError, "Не удалось удалить услугу");
  };

  const { paymentBusy, recordPayment, selectPaymentService, updatePaymentPeriodStart } = useClientDetailPayments({
    activeProjectId,
    clientId,
    load,
    paymentForm,
    services,
    setError,
    setPaymentForm,
    token,
  });

  return {
    activeProjectId, activeProjectName, activeServices, allProjects: companyContext.allProjects, busy, client, clientId, contact, contacts,
    deleteClient, deleteResponsible, deleteService, edit, error, filteredPayments, latestPayment, legalEntities: companyContext.legalEntities,
    linkProject, links, nextService, overdueServices, paymentBusy, paymentFilter, paymentForm, primaryResponsible,
    projectId, recurringTotal, recordPayment, responsible, responsibles, saveClient, selectPaymentService, selectedClientServiceId: searchParams.get("client_service_id") || "",
    selectedPaymentActionId: searchParams.get("payment_action_id") || "", selectedPaymentId: searchParams.get("payment_id") || "",
    services, setActiveProject, setContact, setEdit, setPaymentFilter, setPaymentForm, setProjectId, setResponsible,
    updatePaymentPeriodStart, addContact, addResponsible,
  };
}

export type ClientDetailState = ReturnType<typeof useClientDetailState>;
