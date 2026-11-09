import { ApiError, apiGet } from "./http";
import type { ClientServiceType } from "./types";
import { asArray } from "./utils";

export async function apiGetProjectClientServiceTypes(projectId: string, token: string) {
  try {
    return asArray<ClientServiceType>(
      await apiGet<ClientServiceType[]>(`/api/v1/projects/${projectId}/client-service-types`, token),
    );
  } catch (err) {
    if (!(err instanceof ApiError) || err.status !== 404) {
      throw err;
    }
    return asArray<ClientServiceType>(await apiGet<ClientServiceType[]>("/api/v1/client-service-types", token));
  }
}
