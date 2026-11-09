export type PaymentActionItem = {
  payment_action_item_id: string;
  client_service_id: string;
  service_type_key?: string;
  title: string;
  due_date: string;
  amount: number;
  currency: string;
};

export type PaymentFact = {
  payment_fact_id: string;
  source: string;
  payment_method?: string;
  amount: number;
  currency: string;
  paid_at: string;
  created_at?: string;
};

export type PaymentDocument = {
  payment_action_document_id: string;
  payment_action_id: string;
  file_name: string;
  content_type: string;
  size_bytes: number;
  kind: "invoice" | "act" | "receipt" | "reconciliation" | "file" | string;
  download_url?: string;
  created_at?: string;
};

export type PaymentAction = {
  payment_action_id: string;
  company_id: string;
  client_id: string;
  client_name: string;
  project_id: string;
  project_name: string;
  workflow_run_id?: string;
  status: string;
  source: string;
  payment_method?: string;
  period_start: string;
  period_end: string;
  due_date: string;
  total_amount: number;
  currency: string;
  last_reminded_at?: string;
  snoozed_until?: string;
  confirmed_at?: string;
  comment?: string;
  items?: PaymentActionItem[];
  facts?: PaymentFact[];
  documents?: PaymentDocument[];
  created_at?: string;
};

export type ManualPaymentRecordRequest = {
  project_id?: string;
  client_service_id: string;
  amount: number;
  currency?: string;
  period_start: string;
  period_end?: string;
  paid_at: string;
  payment_method: string;
  next_due_date?: string;
  comment?: string;
  idempotency_key: string;
  metadata?: Record<string, unknown>;
};

export type ManualPaymentRecordResponse = {
  event_id: string;
  event_type: string;
  event_source: string;
  event_subject: string;
  idempotency_key: string;
  payment_action_id: string;
  payment_fact_id: string;
  status: string;
  duplicate: boolean;
};

export type ProjectPaymentScheduleRow = {
  client_id: string;
  client_name: string;
  project_id: string;
  due_date: string;
  total_amount: number;
  currency: string;
  service_count: number;
  payment_action_id?: string;
  payment_status?: string;
  invoice_scenario_key?: string;
};
