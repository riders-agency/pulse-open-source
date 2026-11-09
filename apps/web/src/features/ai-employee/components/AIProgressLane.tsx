import { Loader2 } from "lucide-react";

import { AI_EMPLOYEE_TEST_IDS } from "../aiEmployeeConstants";

export function AIProgressLane({
  sending,
  statuses,
  streamingText,
}: {
  sending: boolean;
  statuses: string[];
  streamingText: string;
}) {
  const visibleStatuses = statuses.map((item) => item.trim()).filter(isVisibleProgressStatus);
  const visibleStreamingText = isVisibleStreamingText(streamingText) ? streamingText : "";
  if (visibleStatuses.length === 0 && !visibleStreamingText) return null;
  return (
    <div className="ai-progress-lane" data-testid={AI_EMPLOYEE_TEST_IDS.progressLane} aria-live="polite">
      <div className="ai-progress-head">
        {sending ? <Loader2 size={14} className="spin" /> : null}
        <span>{visibleStatuses.at(-1) || (sending ? "Выполняю" : "Готово")}</span>
      </div>
      {visibleStatuses.length ? (
        <ol className="ai-progress-steps">
          {visibleStatuses.map((item, index) => (
            <li data-testid={AI_EMPLOYEE_TEST_IDS.progressStep} key={`${item}-${index}`}>
              {item}
            </li>
          ))}
        </ol>
      ) : null}
      {visibleStreamingText ? <p className="ai-stream-text">{visibleStreamingText}<span className="ai-caret" /></p> : null}
    </div>
  );
}

function isVisibleProgressStatus(status: string) {
  if (!status) return false;
  return !/^план действия[:\s]*$/i.test(status);
}

function isVisibleStreamingText(text: string) {
  const normalized = text.trim();
  if (!normalized) return false;
  return !/^план действия[:\s]*$/i.test(normalized);
}
