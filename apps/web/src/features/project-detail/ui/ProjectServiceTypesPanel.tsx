"use client";

import { Plus, Save, Trash2 } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";

import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import type { ClientServiceType } from "../../../lib/api";

type NewServiceType = { display_name: string; service_type_key: string; description: string };

export function ProjectServiceTypesPanel({
  archiveServiceType,
  busyTypeId,
  createServiceType,
  newServiceType,
  patchServiceType,
  serviceTypes,
  setNewServiceType,
  setServiceTypes,
}: {
  archiveServiceType: (type: ClientServiceType) => Promise<void>;
  busyTypeId: string;
  createServiceType: () => Promise<void>;
  newServiceType: NewServiceType;
  patchServiceType: (type: ClientServiceType) => Promise<void>;
  serviceTypes: ClientServiceType[];
  setNewServiceType: Dispatch<SetStateAction<NewServiceType>>;
  setServiceTypes: Dispatch<SetStateAction<ClientServiceType[]>>;
}) {
  return (
    <section className="section-panel section-panel-spaced" id="service-types">
      <header>
        <div>
          <h2>Типы услуг</h2>
          <p>Эти типы используются в клиентах, услугах и оплатах проекта.</p>
        </div>
      </header>
      <div className="section-panel-body stack">
        <div className="stack">
          {serviceTypes.map((type) => (
            <div className="grid-4 grid-align-center" key={type.client_service_type_id}>
              <Input value={type.display_name} onChange={(event) => patchLocalType(type.client_service_type_id, { display_name: event.target.value }, setServiceTypes)} placeholder="Название" />
              <Input value={type.service_type_key} onChange={(event) => patchLocalType(type.client_service_type_id, { service_type_key: event.target.value }, setServiceTypes)} placeholder="Ключ услуги" />
              <Input value={type.description ?? ""} onChange={(event) => patchLocalType(type.client_service_type_id, { description: event.target.value }, setServiceTypes)} placeholder="Описание" />
              <div className="row">
                <Button size="sm" onClick={() => patchServiceType(type)} disabled={busyTypeId === type.client_service_type_id || !type.display_name.trim()} aria-label={`Сохранить тип услуги ${type.display_name || type.service_type_key}`} title={`Сохранить тип услуги ${type.display_name || type.service_type_key}`}>
                  <Save size={16} />
                </Button>
                <Button size="sm" variant="destructive" onClick={() => archiveServiceType(type)} disabled={busyTypeId === type.client_service_type_id} aria-label={`Архивировать тип услуги ${type.display_name || type.service_type_key}`} title={`Архивировать тип услуги ${type.display_name || type.service_type_key}`}>
                  <Trash2 size={16} />
                </Button>
              </div>
            </div>
          ))}
        </div>
        <div className="grid-4 grid-align-center">
          <Input value={newServiceType.display_name} onChange={(event) => setNewServiceType({ ...newServiceType, display_name: event.target.value })} placeholder="Новый тип" />
          <Input value={newServiceType.service_type_key} onChange={(event) => setNewServiceType({ ...newServiceType, service_type_key: event.target.value })} placeholder="key, если нужен" />
          <Input value={newServiceType.description} onChange={(event) => setNewServiceType({ ...newServiceType, description: event.target.value })} placeholder="Описание" />
          <Button onClick={createServiceType} disabled={busyTypeId === "new" || !newServiceType.display_name.trim()}>
            <Plus size={18} />
            <span>Добавить</span>
          </Button>
        </div>
      </div>
    </section>
  );
}

function patchLocalType(id: string, patch: Partial<ClientServiceType>, setServiceTypes: Dispatch<SetStateAction<ClientServiceType[]>>) {
  setServiceTypes((items) => items.map((item) => item.client_service_type_id === id ? { ...item, ...patch } : item));
}
