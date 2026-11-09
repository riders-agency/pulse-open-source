"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, RotateCw } from "lucide-react";
import "./demo-embed.css";

/** Keep the public URL while isolating guest storage and cookies from the working account. */
export function DemoEmbed({ src }: { src: string }) {
  const [attempt, setAttempt] = useState(0);
  const [loaded, setLoaded] = useState(false);
  const [slow, setSlow] = useState(false);
  useEffect(() => {
    if (loaded) return;
    const timer = window.setTimeout(() => setSlow(true), 20_000);
    return () => window.clearTimeout(timer);
  }, [loaded, attempt]);
  const retry = () => { setLoaded(false); setSlow(false); setAttempt((value) => value + 1); };
  return <main className="demo-embed">
    <header><Link href="/tech"><ArrowLeft size={15} /> О продукте</Link><span>Демо Pulse</span><button type="button" onClick={retry} aria-label="Перезагрузить демо"><RotateCw size={16} /></button></header>
    {!loaded && <div className="demo-embed-loading" role="status"><p>{slow ? "Демо пока не открылось. Попробуйте ещё раз." : "Открываем демо…"}</p>{slow && <button type="button" onClick={retry}>Повторить</button>}</div>}
    <iframe key={attempt} src={src} title="Демонстрационное пространство Pulse" onLoad={() => setLoaded(true)} referrerPolicy="strict-origin" allow="clipboard-write" sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-downloads allow-modals" />
  </main>;
}
