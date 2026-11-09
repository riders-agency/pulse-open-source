"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import type { PaymentAction, Project, ProjectPaymentScheduleRow } from "../../../lib/api";
import { usePulseStore } from "../../../store/pulse-store";
import { createProjectPaymentAction, getProjectPayments, uploadPaymentActionDocument } from "../api/paymentsApi";
import { compareDueDate, missingDocumentLabel, paymentActionPath } from "./paymentsModel";

export function useProjectPayments() {
  const params = useParams<{ projectId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setActiveClient = usePulseStore((s) => s.setActiveClient);
  const setActivePaymentAction = usePulseStore((s) => s.setActivePaymentAction);
  const [project, setProject] = useState<Project | null>(null);
  const [actions, setActions] = useState<PaymentAction[]>([]);
  const [schedule, setSchedule] = useState<ProjectPaymentScheduleRow[]>([]);
  const [error, setError] = useState("");
  const [busyActionKey, setBusyActionKey] = useState("");
  const [uploadingDocumentKey, setUploadingDocumentKey] = useState("");
  const projectId = params.projectId || activeProjectId;

  const load = useCallback(() => {
    if (!token || !projectId) return;
    if (params.projectId) setActiveProject(projectId);
    getProjectPayments(projectId, token)
      .then(({ project: projectRow, actions: actionRows, schedule: scheduleRows }) => {
        setProject(projectRow);
        setActions(actionRows);
        setSchedule(scheduleRows);
      })
      .catch((err) => setError(err.message));
  }, [projectId, params.projectId, setActiveProject, token]);

  useEffect(() => {
    load();
  }, [load]);

  const waitingMoney = useMemo(
    () => actions.filter((item) => ["sent", "snoozed"].includes(item.status)).sort(compareDueDate),
    [actions],
  );
  const actionById = useMemo(() => new Map(actions.map((item) => [item.payment_action_id, item])), [actions]);
  const waitingActionIds = useMemo(() => new Set(waitingMoney.map((item) => item.payment_action_id)), [waitingMoney]);
  const needInvoice = useMemo(
    () => schedule
      .filter((item) => !item.payment_action_id || !waitingActionIds.has(item.payment_action_id))
      .sort((a, b) => String(a.due_date || "").localeCompare(String(b.due_date || ""))),
    [schedule, waitingActionIds],
  );
  const needDocuments = useMemo(() => waitingMoney.filter((item) => missingDocumentLabel(item)), [waitingMoney]);
  const totals = useMemo(
    () => ({
      waitingCount: waitingMoney.length,
      waitingAmount: waitingMoney.reduce((sum, item) => sum + (item.total_amount || 0), 0),
      invoiceCount: needInvoice.length,
      invoiceAmount: needInvoice.reduce((sum, item) => sum + (item.total_amount || 0), 0),
      documentCount: needDocuments.length,
    }),
    [needDocuments.length, needInvoice, waitingMoney],
  );

  const createPaymentAction = async (row: ProjectPaymentScheduleRow, invoicePrepared = false) => {
    if (!token || !projectId) return "";
    const out = await createProjectPaymentAction(projectId, token, {
      client_id: row.client_id,
      due_date: row.due_date,
      invoice_prepared: invoicePrepared,
    });
    setActivePaymentAction(out.payment_action_id);
    return out.payment_action_id;
  };

  const markInvoicePrepared = async (row: ProjectPaymentScheduleRow) => {
    if (!token || !projectId) return;
    const key = `${row.client_id}:${row.due_date}:prepared`;
    setBusyActionKey(key);
    setError("");
    try {
      await createPaymentAction(row, true);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось отметить счет подготовленным");
    } finally {
      setBusyActionKey("");
    }
  };

  const openPaymentAction = async (row: ProjectPaymentScheduleRow) => {
    if (!token || !projectId) return;
    if (row.payment_action_id) {
      setActivePaymentAction(row.payment_action_id);
      router.push(paymentActionPath(row.payment_action_id));
      return;
    }
    const key = `${row.client_id}:${row.due_date}`;
    setBusyActionKey(key);
    setError("");
    try {
      const paymentActionId = await createPaymentAction(row);
      if (paymentActionId) router.push(paymentActionPath(paymentActionId));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось открыть счет");
    } finally {
      setBusyActionKey("");
    }
  };

  const uploadActionDocument = async (paymentActionId: string, kind: string, file?: File, busyKey = `${paymentActionId}:${kind}`) => {
    if (!token || !file) return;
    setUploadingDocumentKey(busyKey);
    setError("");
    try {
      const form = new FormData();
      form.append("file", file);
      form.append("kind", kind);
      await uploadPaymentActionDocument(paymentActionId, token, form);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить документ");
    } finally {
      setUploadingDocumentKey("");
    }
  };

  const uploadScheduleInvoice = async (row: ProjectPaymentScheduleRow, file?: File) => {
    if (!token || !file) return;
    const busyKey = row.payment_action_id ? `${row.payment_action_id}:invoice` : `schedule:${row.client_id}:${row.due_date}:invoice`;
    setUploadingDocumentKey(busyKey);
    setError("");
    try {
      const paymentActionId = row.payment_action_id || await createPaymentAction(row);
      if (!paymentActionId) return;
      const form = new FormData();
      form.append("file", file);
      form.append("kind", "invoice");
      await uploadPaymentActionDocument(paymentActionId, token, form);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить счет");
    } finally {
      setUploadingDocumentKey("");
    }
  };

  return {
    actionById,
    busyActionKey,
    error,
    markInvoicePrepared,
    needInvoice,
    openPaymentAction,
    project,
    selectClient: setActiveClient,
    selectPaymentAction: setActivePaymentAction,
    totals,
    uploadActionDocument,
    uploadingDocumentKey,
    uploadScheduleInvoice,
    waitingMoney,
  };
}
