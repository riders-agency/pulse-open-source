"use client";

import { useRef, useState } from "react";
import { ArrowUpRight, Check, Play } from "lucide-react";
import { Button } from "../../components/ui/button";
import { scenarios, sourceRoot } from "./scenarios";

export function ScenarioPlayer({ initialScenario = "ai-task" }: { initialScenario?: (typeof scenarios)[number]["id"] }) {
  const [active, setActive] = useState(() => Math.max(0, scenarios.findIndex((item) => item.id === initialScenario)));
  const tabs = useRef<HTMLDivElement | null>(null);
  const scenario = scenarios[active];
  return (
    <div className="tech-scenarios">
      <div className="tech-tabs" role="tablist" aria-label="Типовые сценарии" ref={tabs}>
        {scenarios.map((item, index) => (
          <Button key={item.id} role="tab" id={`tab-${item.id}`} aria-controls="scenario-panel" aria-selected={index === active} tabIndex={index === active ? 0 : -1}
            variant="ghost"
            onClick={() => setActive(index)} onKeyDown={(event) => {
              const next = event.key === "ArrowRight" ? (index + 1) % scenarios.length : event.key === "ArrowLeft" ? (index + scenarios.length - 1) % scenarios.length : event.key === "Home" ? 0 : event.key === "End" ? scenarios.length - 1 : -1;
              if (next < 0) return;
              event.preventDefault(); setActive(next); tabs.current?.querySelectorAll<HTMLButtonElement>('[role="tab"]')[next]?.focus();
            }}><span>{item.number}</span>{item.label}</Button>
        ))}
      </div>
      <div id="scenario-panel" role="tabpanel" aria-labelledby={`tab-${scenario.id}`} tabIndex={0} className="tech-scenario-panel">
        <div className="tech-recording">
          <div className="tech-window-bar"><span><Play size={13} /> Запись интерфейса</span><span>Тестовое пространство</span></div>
          <video key={scenario.id} controls playsInline preload="none" poster={`/tech/recordings/${scenario.id}.webp`} aria-label={`Запись: ${scenario.title}`}>
            <source src={`/tech/recordings/${scenario.id}.mp4`} type="video/mp4" />
            <track default kind="captions" src={`/tech/recordings/${scenario.id}.vtt`} srcLang="ru" label="Русский" />
            Ваш браузер не поддерживает видео. <a href={`/tech/recordings/${scenario.id}.mp4`}>Скачать запись</a>
          </video>
          <p className="tech-evidence">{scenario.evidence}</p>
        </div>
        <div className="tech-scenario-detail">
          <span className="tech-eyebrow">Сценарий {scenario.number}</span>
          <h3>{scenario.title}</h3><p>{scenario.description}</p>
          <blockquote>{scenario.prompt}</blockquote>
          <ol className="tech-steps">{scenario.steps.map((step) => <li key={step}><Check size={16} /><span>{step}</span></li>)}</ol>
          <details className="tech-code"><summary>{scenario.contract}</summary><pre><code>{scenario.code}</code></pre></details>
          <a className="tech-text-link" href={`${sourceRoot}/${scenario.path}`} target="_blank" rel="noreferrer">Посмотреть реализацию <ArrowUpRight size={16} /></a>
        </div>
      </div>
    </div>
  );
}
