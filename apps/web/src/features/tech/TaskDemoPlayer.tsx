"use client";

import { useRef, useState } from "react";
import { Play } from "lucide-react";

export function TaskDemoPlayer() {
  const video = useRef<HTMLVideoElement>(null);
  const [started, setStarted] = useState(false);
  const [error, setError] = useState("");
  const play = async () => {
    try { await video.current?.play(); setError(""); }
    catch { setError("Нажмите кнопку воспроизведения внизу записи."); }
  };
  return <div className="tech-task-player">
    <video ref={video} controls={started || !!error} playsInline preload="metadata" onPlay={() => setStarted(true)} poster="/tech/recordings/ai-task.webp" aria-label="Запись: создание задачи в чате">
      <source src="/tech/recordings/ai-task.mp4" type="video/mp4" />
      <track default kind="captions" src="/tech/recordings/ai-task.vtt" srcLang="ru" label="Русский" />
      <a href="/tech/recordings/ai-task.mp4">Скачать запись</a>
    </video>
    {!started && <button className="tech-task-play" type="button" onClick={() => void play()}><Play size={18} fill="currentColor" /> Смотреть запись</button>}
    {error && <p role="status">{error}</p>}
  </div>;
}
