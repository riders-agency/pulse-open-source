"use client";

import { addClientContact, addClientResponsible, deleteClientResponsible, linkClientProject } from "../api/clientDetailApi";
import { emptyContactForm, emptyResponsibleForm, type ContactForm, type ResponsibleForm } from "./clientDetailModel";
import { runClientAction } from "./clientDetailStateHelpers";

export function useClientDetailRelations({
  clientId,
  contact,
  contactsCount,
  load,
  projectId,
  responsible,
  responsiblesCount,
  setContact,
  setError,
  setProjectId,
  setResponsible,
  token,
}: {
  clientId: string;
  contact: ContactForm;
  contactsCount: number;
  load: () => void;
  projectId: string;
  responsible: ResponsibleForm;
  responsiblesCount: number;
  setContact: (value: ContactForm) => void;
  setError: (value: string) => void;
  setProjectId: (value: string) => void;
  setResponsible: (value: ResponsibleForm) => void;
  token: string | null;
}) {
  const addResponsible = async () => runClientAction(async () => {
    if (!clientId || !token) return;
    await addClientResponsible(clientId, token, responsible, responsiblesCount === 0);
    setResponsible(emptyResponsibleForm());
    load();
  }, setError, "Не удалось добавить ответственного");

  const deleteResponsible = async (id: string) => runClientAction(async () => {
    if (!clientId || !token) return;
    await deleteClientResponsible(clientId, id, token);
    load();
  }, setError, "Не удалось удалить ответственного");

  const addContact = async () => runClientAction(async () => {
    if (!clientId || !token) return;
    await addClientContact(clientId, token, contact, contactsCount === 0);
    setContact(emptyContactForm());
    load();
  }, setError, "Не удалось добавить контакт");

  const linkProject = async () => runClientAction(async () => {
    if (!clientId || !token || !projectId) return;
    await linkClientProject(clientId, projectId, token);
    setProjectId("");
    load();
  }, setError, "Не удалось связать проект");

  return { addContact, addResponsible, deleteResponsible, linkProject };
}
