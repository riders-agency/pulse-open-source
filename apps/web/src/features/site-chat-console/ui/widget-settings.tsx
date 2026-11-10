import { Copy, MessageCircle, Settings, X } from "lucide-react";

import { Button } from "../../../components/ui/button";
import { Field, Input, Textarea } from "../../../components/ui/input";
import { Select } from "../../../components/ui/select";
import type { SiteChatWidget } from "../../../lib/api";
import type { WidgetForm } from "../model/types";

export function WidgetSettings({ widgets, form, setForm, selectedWidget, onCreate, onSave, onToggle, onCopy, onClose }: {
  widgets: SiteChatWidget[];
  form: WidgetForm;
  setForm: (form: WidgetForm) => void;
  selectedWidget: SiteChatWidget | null;
  onCreate: () => void;
  onSave: (widget: SiteChatWidget) => void;
  onToggle: (widget: SiteChatWidget) => void;
  onCopy: (value: string, label: string) => void;
  onClose: () => void;
}) {
  return (
    <aside className="live-settings-drawer" aria-label="Настройки виджета">
      <header>
        <strong>Виджет и ссылки</strong>
        <Button size="icon" variant="ghost" onClick={onClose} aria-label="Закрыть настройки"><X size={16} /></Button>
      </header>
      <div className="stack">
        <Field label="Название"><Input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} /></Field>
        <Field label="Заголовок"><Input value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} /></Field>
        <Field label="Подзаголовок"><Input value={form.subtitle} onChange={(event) => setForm({ ...form, subtitle: event.target.value })} /></Field>
        <Field label="Кнопка открытия"><Input value={form.launcher_primary_text} onChange={(event) => setForm({ ...form, launcher_primary_text: event.target.value })} /></Field>
        <Field label="Кнопка звонка"><Input value={form.launcher_call_text} onChange={(event) => setForm({ ...form, launcher_call_text: event.target.value })} /></Field>
        <div className="grid-2">
          <Field label="Акцент"><Input value={form.accent_color} onChange={(event) => setForm({ ...form, accent_color: event.target.value })} /></Field>
          <Field label="Позиция">
            <Select value={form.position} onChange={(event) => setForm({ ...form, position: event.target.value as "left" | "right" })}>
              <option value="right">Справа</option>
              <option value="left">Слева</option>
            </Select>
          </Field>
        </div>
        <div className="row">
          {selectedWidget ? <Button onClick={() => onSave(selectedWidget)}><Settings size={16} /> Сохранить</Button> : null}
          <Button variant={selectedWidget ? "secondary" : "default"} onClick={onCreate}><MessageCircle size={16} /> Создать виджет</Button>
        </div>
        {selectedWidget ? (
          <>
            <Textarea value={selectedWidget.embed_script ?? ""} readOnly rows={4} />
            <div className="row">
              <Button variant="secondary" onClick={() => onCopy(selectedWidget.embed_script ?? "", "Код установки скопирован")}><Copy size={15} /> Код установки</Button>
              <Button variant="secondary" onClick={() => onToggle(selectedWidget)}>{selectedWidget.status === "active" ? "Пауза" : "Включить"}</Button>
            </div>
          </>
        ) : <div className="empty-state">Создайте виджет для проекта.</div>}
        <small>Виджетов: {widgets.length}</small>
      </div>
    </aside>
  );
}
