/** Only an explicitly configured local development page can control the executable example. */
export function officeDemoEndpoint() {
  if (process.env.NODE_ENV !== "development" || !process.env.PULSE_OFFICE_DEMO_URL) return null;
  const url = new URL(process.env.PULSE_OFFICE_DEMO_URL);
  if (url.origin !== "http://127.0.0.1:19862" || url.pathname !== "/" || url.search || url.hash) throw new Error("Office example must use the dedicated loopback endpoint");
  return url.origin;
}
