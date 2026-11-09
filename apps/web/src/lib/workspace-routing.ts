export function companySwitchPath(pathname: string, currentCompanyId: string, nextCompanyId: string) {
  if (!nextCompanyId) return "/companies";

  if (pathname === "/clients" || pathname === "/client" || pathname.startsWith("/clients/")) return "/clients";
  if (pathname === "/company-legal-entities") return "/company-legal-entities";
  if (pathname === "/projects" || pathname === "/project" || pathname.startsWith("/projects/")) return "/projects";
  if (pathname === "/company-integrations" || pathname.startsWith("/company-integrations/")) return "/clients";

  if (currentCompanyId && pathname.startsWith(`/companies/${currentCompanyId}/clients`)) return "/clients";
  if (currentCompanyId && pathname.startsWith(`/companies/${currentCompanyId}/legal-entities`)) return "/company-legal-entities";
  if (currentCompanyId && pathname.startsWith(`/companies/${currentCompanyId}/projects`)) return "/projects";
  if (currentCompanyId && pathname.startsWith(`/companies/${currentCompanyId}/integrations`)) return "/clients";
  if (currentCompanyId && pathname === `/companies/${currentCompanyId}`) return "/clients";

  if (pathname.startsWith("/payment-actions/") || pathname === "/payment-action") return "/projects";
  if (pathname.startsWith("/workflows/")) return "/projects";
  if (pathname.startsWith("/runs/")) return "/projects";
  if (pathname === "/companies") return "/clients";

  return pathname || "/dashboard";
}

export function projectSwitchPath(pathname: string, currentProjectId: string, nextProjectId: string) {
  if (!nextProjectId) return pathname || "/dashboard";

  if (pathname === "/project-payments") return "/project-payments";
  if (pathname === "/clients" || pathname === "/client" || pathname.startsWith("/clients/")) return "/clients";
  if (pathname === "/project-workflows" || pathname.startsWith("/project-workflows/")) return "/project-workflows";
  if (pathname === "/project-integrations" || pathname.startsWith("/project-integrations/")) return "/project-integrations";
  if (pathname === "/project") return "/project";

  if (currentProjectId && pathname.startsWith(`/projects/${currentProjectId}/payments`)) return "/project-payments";
  if (currentProjectId && pathname.startsWith(`/projects/${currentProjectId}/workflows`)) return "/project-workflows";
  if (currentProjectId && pathname.startsWith(`/projects/${currentProjectId}/integrations`)) return "/project-integrations";
  if (currentProjectId && pathname.startsWith(`/projects/${currentProjectId}/templates`)) return "/project";
  if (currentProjectId && pathname === `/projects/${currentProjectId}`) return "/project";

  if (pathname.startsWith("/projects/")) return "/project";
  return "/project-workflows";
}

export function projectIdFromPath(pathname: string) {
  const match = pathname.match(/^\/projects\/([^/]+)/);
  return match?.[1] ?? "";
}

export function runIdFromPath(pathname: string) {
  const match = pathname.match(/^\/runs\/([^/]+)/);
  return match?.[1] ?? "";
}
