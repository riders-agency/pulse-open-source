import { availableIntegrationProviders } from "./integration-provider-available";
import { roadmapIntegrationProviders } from "./integration-provider-roadmap";
import {
  integrationCategoryLabels,
  type IntegrationProvider,
} from "./integration-provider-types";

export type {
  CompanyIntegration,
  CompanyIntegrationStatus,
  IntegrationCategory,
  IntegrationConnectionStatus,
  IntegrationProvider,
  IntegrationProviderStatus,
  ProjectIntegrationConnection,
} from "./integration-provider-types";
export {
  integrationCategoryLabels,
  integrationCategoryOptions,
} from "./integration-provider-types";

export const integrationProviders: IntegrationProvider[] = [
  ...availableIntegrationProviders,
  ...roadmapIntegrationProviders,
];

export function providerByKey(providerKey: string) {
  return integrationProviders.find((provider) => provider.providerKey === providerKey);
}

export function providerForIntegration(providerKey?: string, dependencyTypeKey?: string) {
  return integrationProviders.find((provider) => {
    if (dependencyTypeKey && provider.dependencyTypeKey === dependencyTypeKey) return true;
    if (providerKey && provider.providerKeyAlias === providerKey) return true;
    if (providerKey && provider.providerKey === providerKey) return true;
    return false;
  });
}

export function providerSearchText(provider: IntegrationProvider) {
  return [
    provider.name,
    provider.description,
    integrationCategoryLabels[provider.category],
    provider.providerKey,
    provider.providerKeyAlias,
    provider.dependencyTypeKey,
    ...provider.tags,
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
}
