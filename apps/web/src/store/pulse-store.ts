"use client";

import { BROWSER_SESSION_TOKEN } from "../lib/api/http-errors";
import { create } from "zustand";
import { persist } from "zustand/middleware";

export type CompanyMini = { company_id: string; name: string; slug?: string; role?: string };
export type ProjectMini = { project_id: string; name: string; slug?: string; hook_key?: string };
type Theme = "light" | "dark";

type PulseStore = {
  token: string;
  hydrated: boolean;
  activeCompanyId: string;
  activeProjectId: string;
  activeClientId: string;
  activeClientServiceId: string;
  activePaymentActionId: string;
  activeProjectIntegrationId: string;
  theme: Theme;
  sidebarCollapsed: boolean;
  companies: CompanyMini[];
  projects: ProjectMini[];
  setToken: (t: string) => void;
  setHydrated: (hydrated: boolean) => void;
  setActiveCompany: (id: string) => void;
  setActiveProject: (id: string) => void;
  setActiveClient: (id: string) => void;
  setActiveClientService: (id: string) => void;
  setActivePaymentAction: (id: string) => void;
  setActiveProjectIntegration: (id: string) => void;
  setTheme: (theme: Theme) => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  setCompanies: (rows: CompanyMini[]) => void;
  setProjects: (rows: ProjectMini[]) => void;
  clearWorkspace: () => void;
};

function safeRows<T>(value: unknown): T[] {
  return Array.isArray(value) ? value : [];
}

/** Set after store initializer runs — avoids TDZ if persist fires during module init. */
let markHydratedFromPersist: (() => void) | undefined;

export const usePulseStore = create<PulseStore>()(
  persist(
    (set) => {
      markHydratedFromPersist = () => set({ hydrated: true });
      return {
        token: "",
        hydrated: false,
        activeCompanyId: "",
        activeProjectId: "",
        activeClientId: "",
        activeClientServiceId: "",
        activePaymentActionId: "",
        activeProjectIntegrationId: "",
        theme: "light",
        sidebarCollapsed: false,
        companies: [],
        projects: [],
        setToken: (token) => set({ token }),
        setHydrated: (hydrated) => set({ hydrated }),
        setActiveCompany: (activeCompanyId) =>
          set((state) => {
            if (state.activeCompanyId === activeCompanyId) return { activeCompanyId };
            return {
              activeCompanyId,
              activeProjectId: "",
              activeClientId: "",
              activeClientServiceId: "",
              activePaymentActionId: "",
              activeProjectIntegrationId: "",
              projects: [],
            };
          }),
        setActiveProject: (activeProjectId) =>
          set({ activeProjectId, activeClientServiceId: "", activePaymentActionId: "", activeProjectIntegrationId: "" }),
        setActiveClient: (activeClientId) => set({ activeClientId, activeClientServiceId: "" }),
        setActiveClientService: (activeClientServiceId) => set({ activeClientServiceId }),
        setActivePaymentAction: (activePaymentActionId) => set({ activePaymentActionId }),
        setActiveProjectIntegration: (activeProjectIntegrationId) => set({ activeProjectIntegrationId }),
        setTheme: (theme) => set({ theme }),
        setSidebarCollapsed: (sidebarCollapsed) => set({ sidebarCollapsed }),
        setCompanies: (companies) => set({ companies: safeRows<CompanyMini>(companies) }),
        setProjects: (projects) =>
          set((state) => {
            const safe = safeRows<ProjectMini>(projects);
            const activeProjectStillAvailable = safe.some((project) => project.project_id === state.activeProjectId);
            return {
              projects: safe,
              activeProjectId: activeProjectStillAvailable ? state.activeProjectId : "",
            };
          }),
        clearWorkspace: () =>
          set({
            activeCompanyId: "",
            activeProjectId: "",
            activeClientId: "",
            activeClientServiceId: "",
            activePaymentActionId: "",
            activeProjectIntegrationId: "",
            companies: [],
            projects: [],
          }),
      };
    },
    {
      name: "riders-pulse-ui",
      partialize: (state) => ({
        token: state.token === BROWSER_SESSION_TOKEN ? BROWSER_SESSION_TOKEN : "",
        activeCompanyId: state.activeCompanyId, activeProjectId: state.activeProjectId,
        activeClientId: state.activeClientId, activeClientServiceId: state.activeClientServiceId,
        activePaymentActionId: state.activePaymentActionId, activeProjectIntegrationId: state.activeProjectIntegrationId,
        theme: state.theme, sidebarCollapsed: state.sidebarCollapsed,
      }),
      merge: (persisted, current) => {
        const raw = persisted && typeof persisted === "object" ? (persisted as { state?: unknown }) : {};
        const incoming =
          raw.state && typeof raw.state === "object"
            ? (raw.state as Partial<PulseStore>)
            : (persisted as Partial<PulseStore>);
        const state = { ...current, ...incoming };
        return {
          ...state,
          token: state.token === BROWSER_SESSION_TOKEN ? BROWSER_SESSION_TOKEN : "",
          activeCompanyId: typeof state.activeCompanyId === "string" ? state.activeCompanyId : "",
          activeProjectId: typeof state.activeProjectId === "string" ? state.activeProjectId : "",
          activeClientId: typeof state.activeClientId === "string" ? state.activeClientId : "",
          activeClientServiceId: typeof state.activeClientServiceId === "string" ? state.activeClientServiceId : "",
          activePaymentActionId: typeof state.activePaymentActionId === "string" ? state.activePaymentActionId : "",
          activeProjectIntegrationId: typeof state.activeProjectIntegrationId === "string" ? state.activeProjectIntegrationId : "",
          theme: state.theme === "dark" ? "dark" : "light",
          sidebarCollapsed: Boolean(state.sidebarCollapsed),
          companies: safeRows<CompanyMini>(state.companies),
          projects: safeRows<ProjectMini>(state.projects),
        };
      },
      onRehydrateStorage: () => {
        return (_partialState, error) => {
          if (error) {
            console.warn("[pulse-store] persist rehydrate failed", error);
          }
          queueMicrotask(() => markHydratedFromPersist?.());
        };
      },
    },
  ),
);
