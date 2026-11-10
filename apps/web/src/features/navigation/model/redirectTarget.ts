export type RedirectTarget =
  | { href: string; kind: "external" }
  | { href: string; kind: "internal" }
  | { kind: "blocked" };

export function resolveRedirectTarget(to: string, currentHref: string): RedirectTarget {
  const current = new URL(currentHref);
  const next = new URL(to, current);

  if (next.protocol !== "http:" && next.protocol !== "https:") return { kind: "blocked" };
  if (next.origin !== current.origin) return { href: next.href, kind: "external" };

  return { href: `${next.pathname}${next.search}${next.hash}`, kind: "internal" };
}
