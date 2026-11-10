export function replaceSelectedActionInURL(actionId: string) {
  if (typeof window === "undefined") return;
  const url = new URL(window.location.href);
  if (actionId) {
    url.searchParams.set("action", actionId);
  } else {
    url.searchParams.delete("action");
  }
  window.history.replaceState(window.history.state, "", `${url.pathname}${url.search}${url.hash}`);
}
