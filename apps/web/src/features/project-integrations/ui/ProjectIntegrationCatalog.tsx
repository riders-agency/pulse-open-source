import { ArrowRight, Boxes, Search } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import {
  type IntegrationCategory,
  type IntegrationConnectionStatus,
  type IntegrationProvider,
  integrationCategoryLabels,
  integrationCategoryOptions,
} from "../../../lib/integration-providers";
import { PROJECT_INTEGRATIONS_PAGE_SIZE } from "../model/projectIntegrationsModel";

export function ProviderCatalog({
  query,
  category,
  visibleProviders,
  filteredCount,
  visibleCount,
  countsByProvider,
  providerStatus,
  onQueryChange,
  onCategoryChange,
  onLoadMore,
  onOpenProvider,
}: {
  query: string;
  category: IntegrationCategory | "all";
  visibleProviders: IntegrationProvider[];
  filteredCount: number;
  visibleCount: number;
  countsByProvider: Map<string, number>;
  providerStatus: (provider: IntegrationProvider) => IntegrationConnectionStatus;
  onQueryChange: (value: string) => void;
  onCategoryChange: (value: IntegrationCategory | "all") => void;
  onLoadMore: () => void;
  onOpenProvider: (provider: IntegrationProvider) => void;
}) {
  return (
    <div className="stack">
      <div className="integration-filters">
        <label className="ui-field integration-search">
          <span>Поиск провайдера</span>
          <span className="row">
            <Search size={17} />
            <Input value={query} onChange={(event) => onQueryChange(event.target.value)} placeholder="Telegram, SMTP, MAX, CRM..." />
          </span>
        </label>
        <div className="integration-category-list" aria-label="Категории интеграций">
          {integrationCategoryOptions.map((item) => (
            <button
              type="button"
              className={`integration-category-chip ${category === item.key ? "is-active" : ""}`}
              onClick={() => onCategoryChange(item.key)}
              key={item.key}
            >
              {item.label}
            </button>
          ))}
        </div>
      </div>

      {visibleProviders.length === 0 ? (
        <div className="empty-state integration-empty">Провайдеры не найдены. Измените поиск или категорию.</div>
      ) : (
        <div className="provider-grid">
          {visibleProviders.map((provider) => (
            <ProviderCard
              provider={provider}
              status={providerStatus(provider)}
              count={countsByProvider.get(provider.providerKey) ?? 0}
              onOpen={() => onOpenProvider(provider)}
              key={provider.providerKey}
            />
          ))}
        </div>
      )}

      {visibleCount < filteredCount ? (
        <div className="form-actions">
          <Button variant="outline" onClick={onLoadMore}>
            Показать ещё {Math.min(PROJECT_INTEGRATIONS_PAGE_SIZE, filteredCount - visibleCount)}
          </Button>
        </div>
      ) : null}
    </div>
  );
}

function ProviderCard({
  provider,
  status,
  count,
  onOpen,
}: {
  provider: IntegrationProvider;
  status: IntegrationConnectionStatus;
  count: number;
  onOpen: () => void;
}) {
  const canOpen = Boolean(provider.connectPath) && provider.status === "available";
  return (
    <button type="button" className="provider-card" onClick={onOpen} disabled={!canOpen}>
      <ProviderLogo provider={provider} />
      <div className="provider-card-main">
        <div className="between">
          <strong>{provider.name}</strong>
          <StatusBadge status={status} />
        </div>
        <p>{provider.description}</p>
        <div className="provider-card-meta">
          <Badge>{integrationCategoryLabels[provider.category]}</Badge>
          {count > 0 ? <Badge tone="success">{count} подключений</Badge> : null}
          {provider.status !== "available" ? <Badge tone="muted">Заглушка</Badge> : null}
        </div>
      </div>
      <ArrowRight size={18} />
    </button>
  );
}

export function ProviderLogo({ provider }: { provider?: IntegrationProvider }) {
  const Icon = provider?.icon ?? Boxes;
  if (provider?.logoSrc || provider?.iconSrc) {
    return <img className="provider-logo" src={provider.logoSrc || provider.iconSrc} alt="" loading="lazy" />;
  }
  return (
    <span className="provider-logo provider-logo--fallback" aria-hidden="true">
      <Icon size={22} />
    </span>
  );
}

export function StatusBadge({ status }: { status: IntegrationConnectionStatus }) {
  if (status === "connected") return <Badge tone="success">Подключено</Badge>;
  if (status === "coming_soon") return <Badge tone="warning">Скоро</Badge>;
  if (status === "disabled") return <Badge tone="muted">Выключено</Badge>;
  if (status === "error") return <Badge tone="danger">Ошибка</Badge>;
  return <Badge>Не подключено</Badge>;
}
