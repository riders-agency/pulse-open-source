export type ProviderCriterionSpec = {
  type: string;
  label: string;
  description?: string;
  expected?: string;
  source?: string;
};

export type ProviderCatalogEntry = {
  action_type_key: string;
  provider_key: string;
  display_name: string;
  criteria?: ProviderCriterionSpec[] | null;
  failure_categories?: Array<{ code: string; label: string; retryable: boolean; next_action?: string }> | null;
  retry?: { summary?: string; retryable_reason_ids?: string[] };
  raw_response_schema?: Record<string, unknown>;
  expose?: Array<{ key: string; label: string; description?: string }>;
  template_scopes?: string[];
  ui?: Record<string, unknown>;
};

export type IntegrationDocAvailableIntegration = {
  action_dependency_id: string;
  name: string;
  status: string;
  dependency_type_key: string;
  provider_key?: string;
};

export type IntegrationDocSearchResult = {
  doc_key: string;
  public_slug: string;
  title: string;
  summary: string;
  provider_key: string;
  action_type_keys: string[];
  dependency_type_keys: string[];
  tags: string[];
  intents: string[];
  score: number;
  reason: string;
  snippet?: string;
  public_url: string;
  markdown_url: string;
  available_integrations?: IntegrationDocAvailableIntegration[];
};
