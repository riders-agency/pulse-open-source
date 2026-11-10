export const PROJECT_INTEGRATIONS_PAGE_SIZE = 12;

export {
  editIntegrationMetaHint,
  editProjectIntegrationTypes,
  metadataPreset,
  newIntegrationMetaHint,
  newProjectIntegrationTypes,
  type EditProjectIntegrationType,
  type NewProjectIntegrationType,
} from "../registry/integrationTypes";

function parseIntegrationMetadata(metadata: string, showMetadata: boolean) {
  if (!showMetadata) return {};
  return JSON.parse(metadata.trim() === "" ? "{}" : metadata);
}

export function projectIntegrationCreatePayload(input: {
  name: string;
  typeKey: string;
  secret: string;
  metadata: string;
  showMetadata: boolean;
}) {
  return {
    name: input.name.trim(),
    dependency_type_key: input.typeKey,
    plaintext_secret: input.secret.trim(),
    metadata_json: parseIntegrationMetadata(input.metadata, input.showMetadata),
  };
}

export function projectIntegrationUpdatePayload(input: { name: string; metadata: string; showMetadata: boolean }) {
  return {
    name: input.name.trim(),
    metadata_json: parseIntegrationMetadata(input.metadata, input.showMetadata),
  };
}

export function projectIntegrationSecretPayload(secret: string) {
  return {
    plaintext_secret: secret.trim(),
  };
}

export function providerLabel(key?: string) {
  const labels: Record<string, string> = {
    max: "MAX",
    telegram: "Telegram",
    email: "Почта",
  };
  return labels[key ?? ""] ?? "Другое";
}
