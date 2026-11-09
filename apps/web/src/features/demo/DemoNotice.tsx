"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import "./demo.css";

export function DemoNotice() {
  const [visible, setVisible] = useState(false);
  useEffect(() => { setVisible(Boolean(sessionStorage.getItem("pulse-demo-context"))); }, []);
  if (!visible) return null;
  return <div className="demo-notice" role="note"><strong>Демо-пространство</strong><span>Вымышленные данные · ИИ через APInet · внешние отправки отключены</span><Link href="/tech">Как устроен Pulse</Link></div>;
}
