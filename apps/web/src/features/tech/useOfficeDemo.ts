"use client";

import { useEffect, useState } from "react";
import { officeDemoRequest, type OfficeState } from "../../lib/api/officeDemo";
import recording from "./officeRecording.generated.json";

export function useOfficeDemo(endpoint: string | null) {
  const saved = recording as OfficeState;
  const [state, setState] = useState<OfficeState>(endpoint ? { ...saved, status: "idle", events: [], results: {} } : saved);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [replay, setReplay] = useState(0);
  useEffect(() => {
    if (!endpoint) return;
    let active = true;
    let timer: ReturnType<typeof setTimeout>;
    const poll = async () => {
      try { const next = await officeDemoRequest(endpoint); if (active) { setState(next); setError(""); } }
      catch { if (active) setError("Связь с локальным запуском потеряна. Результаты не подменяются записью."); }
      if (active) timer = setTimeout(() => void poll(), 350);
    };
    void poll();
    return () => { active = false; clearTimeout(timer); };
  }, [endpoint]);
  useEffect(() => {
    if (!replay || endpoint) return;
    let index = 0;
    const timer = setInterval(() => {
      index++;
      const events = saved.events.slice(0, index);
      const done = index >= saved.events.length;
      setState({ ...saved, events, status: done ? saved.status : "replaying", results: done ? saved.results : {} });
      if (done) clearInterval(timer);
    }, 950);
    return () => clearInterval(timer);
  }, [replay, endpoint, saved]);
  const act = async (action: string, data: object = {}) => {
    if (!endpoint || busy) return;
    setBusy(true); setError("");
    try { await officeDemoRequest(endpoint, action, data); setState(await officeDemoRequest(endpoint)); }
    catch (err) { setError(err instanceof Error ? err.message : "Не удалось выполнить действие."); }
    finally { setBusy(false); }
  };
  const restart = () => { setState({ ...saved, status: "replaying", events: [], results: {} }); setReplay((value) => value + 1); };
  return { state, error, busy, act, restart };
}
