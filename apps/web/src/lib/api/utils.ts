export function jsonText(value: unknown) {
  return JSON.stringify(value ?? {}, null, 2);
}

export function asArray<T>(value: unknown): T[] {
  return Array.isArray(value) ? value as T[] : [];
}
