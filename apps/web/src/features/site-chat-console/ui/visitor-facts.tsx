import { MonitorSmartphone } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import type { SiteVisitorSession, VisitorStateSnapshot } from "../../../lib/api";
import { formatDate, formatDateTime } from "../../../lib/utils";
import { formatDuration, formatDurationMinutes, shortURL } from "../model/format";

export function VisitorFacts({ snapshot, session }: { snapshot: VisitorStateSnapshot; session: SiteVisitorSession | null }) {
  const metrika = snapshot.analytics?.metrika;
  const primaryBlock = snapshot.viewport_context?.primary_block;
  return (
    <Card>
      <CardHeader>
        <CardTitle><MonitorSmartphone size={16} /> Сессия</CardTitle>
        <CardDescription>{session ? `Старт ${formatDateTime(session.started_at)} · активность ${formatDateTime(session.last_activity_at)}` : "Нет выбранного посетителя"}</CardDescription>
      </CardHeader>
      <CardContent className="live-facts">
        <div><span>Посетитель</span><strong>{session?.visitor_name || "Не указано"}</strong></div>
        <div><span>Контакт</span><strong>{session?.visitor_contact || "Не указано"}</strong></div>
        <div><span>Заголовок</span><strong>{session?.current_title || snapshot.current_title || "Нет данных"}</strong></div>
        <div><span>Адрес</span><strong>{shortURL(session?.current_url)}</strong></div>
        <div><span>Источник</span><strong>{snapshot.analytics?.referrer ? shortURL(snapshot.analytics.referrer) : "Прямой переход"}</strong></div>
        <div><span>Область просмотра</span><strong>{snapshot.viewport?.width ?? 0}×{snapshot.viewport?.height ?? 0}</strong></div>
        <div><span>Экран</span><strong>{snapshot.analytics?.screen?.width ?? 0}×{snapshot.analytics?.screen?.height ?? 0}</strong></div>
        <div><span>Устройство</span><strong>{snapshot.analytics?.device_type ?? "Не указано"}</strong></div>
        <div><span>Браузер</span><strong>{snapshot.analytics?.browser ?? "Не определён"}</strong></div>
        <div><span>Язык и часовой пояс</span><strong>{snapshot.analytics?.language ?? "Не указано"} · {snapshot.analytics?.timezone ?? "—"}</strong></div>
        <div><span>Метрика</span><strong>{metrika?.client_id || metrika?.user_id || metrika?.session_id || "Не определено"}</strong></div>
        <div><span>Просматривает</span><strong>{primaryBlock?.heading || primaryBlock?.text || snapshot.viewport_context?.visible_text || "Не указано"}</strong></div>
        <div><span>Блок страницы</span><strong>{primaryBlock?.selector || primaryBlock?.tag || "Нет данных"} · {Math.round(Number(primaryBlock?.visible_ratio ?? 0) * 100)}%</strong></div>
        <div><span>Соединение</span><strong>{snapshot.activity?.connection_status ?? "Не указано"}</strong></div>
        <div><span>Время на сайте</span><strong>{formatDuration(session?.duration_seconds)}</strong></div>
        <div><span>Минут на сайте</span><strong>{formatDurationMinutes(session?.duration_seconds)}</strong></div>
        <div><span>Начало сессии</span><strong>{session ? formatDateTime(session.started_at) : "Не указано"}</strong></div>
        <div><span>Сообщения посетителя</span><strong>{session?.visitor_message_count ?? 0}</strong></div>
        <div><span>Последняя активность</span><strong>{snapshot.activity?.last_seen_at ? formatDate(snapshot.activity.last_seen_at) : "Не указано"}</strong></div>
      </CardContent>
    </Card>
  );
}
