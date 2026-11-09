export function publicDemoEnabled() {
  return process.env.PULSE_PUBLIC_DEMO === "true";
}

export function externalDemoUrl(autoStart: boolean) {
  const value = process.env.PULSE_PUBLIC_DEMO_URL;
  if (!value) return null;
  const url = new URL(value);
  if (url.protocol !== "https:" || url.username || url.password) {
    throw new Error("PULSE_PUBLIC_DEMO_URL must be a public HTTPS URL without credentials");
  }
  if (autoStart) url.searchParams.set("start", "1");
  return url.toString();
}
