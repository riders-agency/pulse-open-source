import type { LucideIcon } from "lucide-react";

export type IntegrationCategory =
  | "ai"
  | "crm"
  | "email"
  | "data_enrichment"
  | "analytics"
  | "payments"
  | "telephony"
  | "support"
  | "marketing"
  | "storage"
  | "maps"
  | "delivery"
  | "other";

export type IntegrationProviderStatus = "available" | "coming_soon" | "disabled";
export type IntegrationConnectionStatus = "connected" | "not_connected" | "coming_soon" | "disabled" | "error";
export type CompanyIntegrationStatus = IntegrationConnectionStatus;

export type IntegrationProvider = {
  providerKey: string;
  providerKeyAlias?: string;
  dependencyTypeKey?: string;
  name: string;
  description: string;
  category: IntegrationCategory;
  status: IntegrationProviderStatus;
  tags: string[];
  logoSrc?: string;
  iconSrc?: string;
  logoKey?: string;
  icon: LucideIcon;
  connectPath?: () => string;
  notes?: string;
};

export type ProjectIntegrationConnection = {
  integrationId: string;
  providerKey: string;
  dependencyTypeKey?: string;
  status: IntegrationConnectionStatus;
  name: string;
  source?: string;
  updatedAt?: string;
  error?: string;
};

export type CompanyIntegration = ProjectIntegrationConnection;

export const integrationCategoryLabels: Record<IntegrationCategory, string> = {
  ai: "AI",
  crm: "CRM",
  email: "Почта",
  data_enrichment: "Данные",
  analytics: "Аналитика",
  payments: "Оплаты",
  telephony: "Телефония",
  support: "Поддержка",
  marketing: "Маркетинг",
  storage: "Файлы",
  maps: "Карты",
  delivery: "Доставка",
  other: "Другое",
};

export const integrationCategoryOptions: Array<{ key: IntegrationCategory | "all"; label: string }> = [
  { key: "all", label: "Все" },
  { key: "ai", label: integrationCategoryLabels.ai },
  { key: "support", label: integrationCategoryLabels.support },
  { key: "email", label: integrationCategoryLabels.email },
  { key: "crm", label: integrationCategoryLabels.crm },
  { key: "payments", label: integrationCategoryLabels.payments },
  { key: "telephony", label: integrationCategoryLabels.telephony },
  { key: "analytics", label: integrationCategoryLabels.analytics },
  { key: "marketing", label: integrationCategoryLabels.marketing },
  { key: "storage", label: integrationCategoryLabels.storage },
  { key: "maps", label: integrationCategoryLabels.maps },
  { key: "delivery", label: integrationCategoryLabels.delivery },
  { key: "data_enrichment", label: integrationCategoryLabels.data_enrichment },
  { key: "other", label: integrationCategoryLabels.other },
];
