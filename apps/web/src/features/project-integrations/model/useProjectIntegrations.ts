"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import type { Integration } from "../../../lib/api";
import {
  type IntegrationCategory,
  type IntegrationConnectionStatus,
  type IntegrationProvider,
  integrationProviders,
  providerForIntegration,
  providerSearchText,
} from "../../../lib/integration-providers";
import { usePulseStore } from "../../../store/pulse-store";
import { listProjectIntegrations, restoreProjectIntegration } from "../api/projectIntegrationsApi";
import { PROJECT_INTEGRATIONS_PAGE_SIZE } from "./projectIntegrationsModel";

export function useProjectIntegrations() {
  const params = useParams<{ projectId?: string }>();
  const router = useRouter();
  const token = usePulseStore((s) => s.token);
  const activeProjectId = usePulseStore((s) => s.activeProjectId);
  const setActiveProject = usePulseStore((s) => s.setActiveProject);
  const setActiveProjectIntegration = usePulseStore((s) => s.setActiveProjectIntegration);
  const [rows, setRows] = useState<Integration[]>([]);
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState<IntegrationCategory | "all">("all");
  const [visibleCount, setVisibleCount] = useState(PROJECT_INTEGRATIONS_PAGE_SIZE);
  const [catalogOpen, setCatalogOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const projectId = params.projectId || activeProjectId;

  const load = useCallback(() => {
    if (!token || !projectId) {
      setLoading(false);
      return;
    }
    if (params.projectId) setActiveProject(projectId);
    setLoading(true);
    setError("");
    listProjectIntegrations(projectId, token)
      .then((data) => setRows(data))
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [params.projectId, projectId, setActiveProject, token]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    setVisibleCount(PROJECT_INTEGRATIONS_PAGE_SIZE);
  }, [category, query]);

  const activeRows = useMemo(() => rows.filter((row) => row.status !== "deleted"), [rows]);
  const archivedRows = useMemo(() => rows.filter((row) => row.status === "deleted"), [rows]);
  const countsByProvider = useMemo(() => {
    const counts = new Map<string, number>();
    for (const row of activeRows) {
      const provider = providerForIntegration(row.provider_key, row.dependency_type_key);
      const key = provider?.providerKey ?? row.dependency_type_key;
      counts.set(key, (counts.get(key) ?? 0) + 1);
    }
    return counts;
  }, [activeRows]);
  const filteredProviders = useMemo(() => {
    const q = query.trim().toLowerCase();
    return integrationProviders.filter((provider) => {
      if (category !== "all" && provider.category !== category) return false;
      if (!q) return true;
      return providerSearchText(provider).includes(q);
    });
  }, [category, query]);
  const visibleProviders = filteredProviders.slice(0, visibleCount);

  const providerStatus = useCallback(
    (provider: IntegrationProvider): IntegrationConnectionStatus => {
      if ((countsByProvider.get(provider.providerKey) ?? 0) > 0) return "connected";
      if (provider.status === "coming_soon") return "coming_soon";
      if (provider.status === "disabled") return "disabled";
      return "not_connected";
    },
    [countsByProvider],
  );

  const openProvider = useCallback(
    (provider: IntegrationProvider) => {
      if (!provider.connectPath || provider.status !== "available") return;
      setCatalogOpen(false);
      router.push(provider.connectPath());
    },
    [router],
  );

  const restoreIntegration = async (id: string) => {
    if (!token || !projectId) return;
    try {
      setError("");
      await restoreProjectIntegration(projectId, id, token);
      setNotice("Интеграция восстановлена");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось восстановить");
    }
  };

  return {
    activeRows,
    archivedRows,
    catalogOpen,
    category,
    countsByProvider,
    error,
    filteredProviderCount: filteredProviders.length,
    loadMoreProviders: () => setVisibleCount((value) => value + PROJECT_INTEGRATIONS_PAGE_SIZE),
    loading,
    notice,
    openIntegration: setActiveProjectIntegration,
    openProvider,
    projectId,
    providerStatus,
    query,
    restoreIntegration,
    setCatalogOpen,
    setCategory,
    setQuery,
    visibleCount,
    visibleProviders,
  };
}
