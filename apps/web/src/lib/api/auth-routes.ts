export const AUTH_API_ROUTES = {
  register: "/api/v1/auth/register",
  login: "/api/v1/auth/login",
  refresh: "/api/v1/auth/refresh",
  refreshCSRF: "/api/v1/auth/refresh/csrf",
  csrf: "/api/v1/auth/csrf",
  logout: "/api/v1/auth/logout",
  me: "/api/v1/auth/me",
} as const;

const AUTH_BOOTSTRAP_PATHS = [
  AUTH_API_ROUTES.login,
  AUTH_API_ROUTES.register,
  AUTH_API_ROUTES.refresh,
  AUTH_API_ROUTES.refreshCSRF,
] as const;

export function isAuthBootstrapPathname(pathname: string) {
  return (AUTH_BOOTSTRAP_PATHS as readonly string[]).includes(pathname);
}
