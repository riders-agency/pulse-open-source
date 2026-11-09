export function boardColumnsClass(count: number) {
  return `work-board-columns-${clamp(Math.round(count || 1), 1, 5)}`;
}

export function meterFillClass(value: number, prefix: "work-task-progress-fill" | "workload-meter-fill") {
  const bucket = clamp(Math.round(value / 10) * 10, 0, 100);
  return `${prefix}-${bucket}`;
}

export function timelineOffsetClass(index: number) {
  return `work-timeline-offset-${clamp(index % 8, 0, 7)}`;
}

export function timelineSpanClass(index: number) {
  return `work-timeline-span-${clamp(Math.max(2, (index % 6) + 2), 2, 7)}`;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}
