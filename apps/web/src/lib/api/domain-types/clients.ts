export type {
  ManualPaymentRecordRequest,
  ManualPaymentRecordResponse,
  PaymentAction,
  PaymentActionItem,
  PaymentDocument,
  PaymentFact,
  ProjectPaymentScheduleRow,
} from "./client-payments";

export type Client = {
  client_id: string;
  company_id: string;
  client_type: "individual" | "individual_entrepreneur" | "legal_entity";
  status: string;
  name: string;
  short_name?: string;
  full_name?: string;
  inn?: string;
  kpp?: string;
  ogrn?: string;
  ogrnip?: string;
  legal_address?: string;
  actual_address?: string;
  okved?: string;
  comment?: string;
  projects_count?: number;
  services_count?: number;
};

export type DaDataIntegration = {
  configured: boolean;
  action_dependency_id?: string;
  name?: string;
  base_url?: string;
  source?: string;
  updated_at?: string;
};

export type CompanyLegalEntity = {
  company_legal_entity_id: string;
  company_id: string;
  entity_type: "individual_entrepreneur" | "legal_entity";
  status: string;
  name: string;
  short_name?: string;
  full_name?: string;
  inn?: string;
  kpp?: string;
  ogrn?: string;
  ogrnip?: string;
  legal_address?: string;
  actual_address?: string;
  okved?: string;
  comment?: string;
  dadata_raw_json?: unknown;
};

export type ClientResponsible = {
  client_responsible_id: string;
  company_id: string;
  client_id: string;
  user_id?: string;
  company_legal_entity_id?: string;
  legal_entity_name?: string;
  display_name: string;
  role: string;
  is_primary: boolean;
  status: string;
  comment?: string;
};

export type DaDataPartySuggestion = {
  client_type: "individual_entrepreneur" | "legal_entity";
  entity_type?: "individual_entrepreneur" | "legal_entity";
  name: string;
  short_name?: string;
  full_name?: string;
  inn?: string;
  kpp?: string;
  ogrn?: string;
  ogrnip?: string;
  legal_address?: string;
  actual_address?: string;
  okved?: string;
  dadata_status?: string;
  dadata_raw_json?: unknown;
};

export type ClientContact = {
  client_contact_id: string;
  client_id: string;
  first_name?: string;
  last_name?: string;
  middle_name?: string;
  position?: string;
  email?: string;
  phone?: string;
  telegram?: string;
  preferred_channel: string;
  comment?: string;
  is_primary: boolean;
  status: string;
};

export type ClientProject = {
  project_id: string;
  company_id: string;
  name: string;
  slug: string;
};

export type ClientServiceType = {
  client_service_type_id: string;
  company_id?: string;
  project_id?: string;
  scope_kind?: string;
  service_type_key: string;
  display_name: string;
  description?: string;
  status?: string;
  sort_order: number;
};

export type ClientService = {
  client_service_id: string;
  company_id: string;
  client_id: string;
  project_id: string;
  client_service_type_id: string;
  service_type_key: string;
  service_type_name: string;
  name: string;
  status: string;
  amount: number;
  raw_price_text?: string;
  currency: string;
  billing_period: string;
  next_due_date?: string;
  remind_days_before: number;
  remind_every_days: number;
  metadata_json?: Record<string, unknown>;
};
