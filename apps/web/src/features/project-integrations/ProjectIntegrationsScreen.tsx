"use client";

import Link from "next/link";
import { ArrowRight, Boxes, Plus } from "lucide-react";

import { Badge } from "../../components/ui/badge";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "../../components/ui/dialog";
import {
  integrationCategoryLabels,
  integrationProviders,
  providerForIntegration,
} from "../../lib/integration-providers";
import { providerLabel } from "./model/projectIntegrationsModel";
import { useProjectIntegrations } from "./model/useProjectIntegrations";
import { ProviderCatalog, ProviderLogo, StatusBadge } from "./ui/ProjectIntegrationCatalog";

export default function IntegrationsPage() {
  const integrations = useProjectIntegrations();

  return (
    <main className="page">
      <div className="page-title">
        <div>
          <h1>Интеграции проекта</h1>
          <p className="muted">Подключения проекта для сценариев: MAX, Telegram, почта и будущие провайдеры.</p>
        </div>
        <Dialog open={integrations.catalogOpen} onOpenChange={integrations.setCatalogOpen}>
          <DialogTrigger className="ui-button ui-button--default ui-button--md" disabled={!integrations.projectId}>
            <Plus size={18} />
            <span>Добавить интеграцию</span>
          </DialogTrigger>
          <DialogContent className="integration-catalog-dialog">
            <DialogHeader>
              <DialogTitle>Каталог провайдеров проекта</DialogTitle>
              <DialogDescription>
                Провайдер описывает тип подключения. Ключи и настройки хранятся только в конкретной интеграции проекта.
              </DialogDescription>
            </DialogHeader>
            <ProviderCatalog
              query={integrations.query}
              category={integrations.category}
              visibleProviders={integrations.visibleProviders}
              filteredCount={integrations.filteredProviderCount}
              visibleCount={integrations.visibleCount}
              countsByProvider={integrations.countsByProvider}
              providerStatus={integrations.providerStatus}
              onQueryChange={integrations.setQuery}
              onCategoryChange={integrations.setCategory}
              onLoadMore={integrations.loadMoreProviders}
              onOpenProvider={integrations.openProvider}
            />
          </DialogContent>
        </Dialog>
      </div>

      {!integrations.projectId ? <div className="empty-state integration-empty">Выберите проект слева, чтобы управлять интеграциями.</div> : null}
      {integrations.error ? <Badge tone="danger">{integrations.error}</Badge> : null}
      {integrations.notice ? <Badge tone="success">{integrations.notice}</Badge> : null}

      <section className="grid-3">
        <div className="metric"><span>Активные</span><strong>{integrations.activeRows.length}</strong></div>
        <div className="metric"><span>Провайдеры</span><strong>{integrations.countsByProvider.size}</strong></div>
        <div className="metric"><span>В каталоге</span><strong>{integrationProviders.length}</strong></div>
      </section>

      <section className="section-panel section-panel-spaced">
        <header>
          <div>
            <h2>Подключённые</h2>
            <p>Эти интеграции уже доступны в редакторе сценариев выбранного проекта.</p>
          </div>
          <Badge>{integrations.loading ? "Загрузка" : `${integrations.activeRows.length} активных`}</Badge>
        </header>
        <div className="section-panel-body">
          {integrations.loading ? (
            <div className="integration-grid">
              <div className="provider-card provider-card--loading" />
              <div className="provider-card provider-card--loading" />
            </div>
          ) : integrations.activeRows.length === 0 ? (
            <div className="empty-state integration-empty">
              <Boxes size={28} />
              <span>Подключений проекта пока нет. Откройте каталог и выберите провайдера.</span>
            </div>
          ) : (
            <div className="integration-grid">
              {integrations.activeRows.map((integration) => {
                const provider = providerForIntegration(integration.provider_key, integration.dependency_type_key);
                return (
                  <Link
                    className="provider-card provider-card--connected"
                    href="/project-integration"
                    key={integration.action_dependency_id}
                    onClick={() => integrations.openIntegration(integration.action_dependency_id)}
                  >
                    <ProviderLogo provider={provider} />
                    <div className="provider-card-main">
                      <div className="between">
                        <strong>{integration.name}</strong>
                        <StatusBadge status="connected" />
                      </div>
                      <p>{provider?.description ?? integration.dependency_type_display_name ?? integration.dependency_type_key}</p>
                      <div className="provider-card-meta">
                        <Badge>{provider ? integrationCategoryLabels[provider.category] : providerLabel(integration.provider_key)}</Badge>
                        <Badge tone="muted">{integration.dependency_type_display_name || integration.dependency_type_key}</Badge>
                      </div>
                    </div>
                    <ArrowRight size={18} />
                  </Link>
                );
              })}
            </div>
          )}
        </div>
      </section>

      <section className="section-panel section-panel-spaced">
        <header>
          <div>
            <h2>Каталог</h2>
            <p>Поиск, категории и плитки провайдеров. Живые провайдеры подключаются сразу, остальные помечены как будущие.</p>
          </div>
        </header>
        <div className="section-panel-body">
          <ProviderCatalog
            query={integrations.query}
            category={integrations.category}
            visibleProviders={integrations.visibleProviders}
            filteredCount={integrations.filteredProviderCount}
            visibleCount={integrations.visibleCount}
            countsByProvider={integrations.countsByProvider}
            providerStatus={integrations.providerStatus}
            onQueryChange={integrations.setQuery}
            onCategoryChange={integrations.setCategory}
            onLoadMore={integrations.loadMoreProviders}
            onOpenProvider={integrations.openProvider}
          />
        </div>
      </section>

      {integrations.archivedRows.length > 0 ? (
        <section className="section-panel section-panel-spaced">
          <header>
            <div>
              <h2>Архив</h2>
              <p>Отключённые интеграции сохраняются для истории запусков.</p>
            </div>
          </header>
          <div className="section-panel-body compact-list">
            {integrations.archivedRows.map((row) => (
              <div className="compact-item" key={row.action_dependency_id}>
                <div className="between">
                  <strong>{row.name}</strong>
                  <button type="button" className="ui-button ui-button--secondary ui-button--sm" onClick={() => void integrations.restoreIntegration(row.action_dependency_id)}>
                    Восстановить
                  </button>
                </div>
                <span className="muted">{row.dependency_type_display_name || row.dependency_type_key}</span>
              </div>
            ))}
          </div>
        </section>
      ) : null}
    </main>
  );
}
