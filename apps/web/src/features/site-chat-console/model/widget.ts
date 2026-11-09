import type { SiteChatWidget } from "../../../lib/api";
import type { WidgetForm } from "./types";

export const DEFAULT_WIDGET_ACCENT_COLOR = "var(--primary)";

function resolvedDefaultAccentColor() {
  if (typeof window === "undefined") return DEFAULT_WIDGET_ACCENT_COLOR;
  const computed = window.getComputedStyle(document.documentElement).getPropertyValue("--primary").trim();
  return computed || DEFAULT_WIDGET_ACCENT_COLOR;
}

export const initialWidget: WidgetForm = {
  name: "Live Assist",
  title: "Live Assist",
  subtitle: "",
  accent_color: DEFAULT_WIDGET_ACCENT_COLOR,
  position: "right",
  launcher_primary_text: "Консультация онлайн",
  launcher_call_text: "Звонок",
};

function widgetSettingString(settings: Record<string, unknown> | undefined, path: string[], fallback: string) {
  let current: unknown = settings;
  for (const key of path) {
    if (!current || typeof current !== "object") return fallback;
    current = (current as Record<string, unknown>)[key];
  }
  const value = typeof current === "string" ? current.trim() : "";
  if (!value) return fallback;
  if (path.join(".") === "launcher.call_text" && value === "Позвонить") return "Звонок";
  if (path.join(".") === "launcher.primary_text" && value === "Написать") return "Консультация онлайн";
  return value;
}

export function widgetFormFromWidget(widget: SiteChatWidget): WidgetForm {
  return {
    name: widget.name || initialWidget.name,
    title: widget.title || initialWidget.title,
    subtitle: widget.subtitle || "",
    accent_color: widget.accent_color || initialWidget.accent_color,
    position: widget.position === "left" ? "left" : "right",
    launcher_primary_text: widgetSettingString(widget.settings_json, ["launcher", "primary_text"], initialWidget.launcher_primary_text),
    launcher_call_text: widgetSettingString(widget.settings_json, ["launcher", "call_text"], initialWidget.launcher_call_text),
  };
}

export function widgetSettingsPayload(form: WidgetForm, current?: Record<string, unknown>) {
  const launcher = current?.launcher && typeof current.launcher === "object" ? current.launcher as Record<string, unknown> : {};
  const theme = current?.theme && typeof current.theme === "object" ? current.theme as Record<string, unknown> : {};
  return {
    ...(current ?? {}),
    enabled_modes: { chat: true, call: true, monitoring: true },
    launcher: {
      ...launcher,
      primary_text: form.launcher_primary_text.trim() || initialWidget.launcher_primary_text,
      call_text: form.launcher_call_text.trim() || initialWidget.launcher_call_text,
    },
    recording_enabled: false,
    recording_consent_required: true,
    theme: {
      ...theme,
      accent_color: form.accent_color === DEFAULT_WIDGET_ACCENT_COLOR ? resolvedDefaultAccentColor() : form.accent_color,
      position: form.position,
    },
  };
}
