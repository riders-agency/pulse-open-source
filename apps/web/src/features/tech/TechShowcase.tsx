import Link from "next/link";
import { demoEntryHref } from "../../lib/config/demo";
import { Activity, ArrowUpRight, ChevronDown } from "lucide-react";
import { ScenarioPlayer } from "./ScenarioPlayer";
import { TechArchitecture } from "./TechArchitecture";
import { TechFeatures } from "./TechFeatures";
import { TechCodeTour } from "./TechCodeTour";
import { TechAgentRuntime } from "./TechAgentRuntime";
import { TaskDemoPlayer } from "./TaskDemoPlayer";
import { OfficeScenario } from "./OfficeScenario";
import "./tech.css";
import "./scenarios.css";
import "./engineering.css";
import "./task-demo.css";

export function TechShowcase() {
  return (
    <main className="tech-page">
      <a className="tech-skip" href="#task-demo">Перейти к записи</a>
      <header className="tech-header"><Link href="/tech" className="tech-logo"><Activity size={26} /><span>pulse<span className="tech-logo-dot">.</span></span></Link><nav aria-label="Навигация по витрине"><a href="https://github.com/riders-agency/pulse-open-source" target="_blank" rel="noreferrer">GitHub <ArrowUpRight size={14} /></a><Link href={demoEntryHref()} className="tech-cta">Попробовать <ArrowUpRight size={15} /></Link></nav></header>
      <section id="task-demo" className="tech-task-demo">
        <p className="tech-platform-label">Pulse — платформа для ИИ-агентов</p>
        <div className="tech-task-heading"><h1>Попросили в чате.<br /><span>Задача готова.</span></h1><p>Напишите, что нужно сделать. Подтвердите — и продолжайте работу.</p></div>
        <figure className="tech-task-recording">
          <TaskDemoPlayer />
          <figcaption><span>Запись работающего продукта · демонстрационные данные</span><a href="#agent-runtime">Как это выполнено <ChevronDown size={14} /></a></figcaption>
        </figure>
      </section>
      <TechAgentRuntime />
      <OfficeScenario />
      <TechFeatures />
      <section className="tech-section" id="more-scenarios"><div className="tech-section-heading"><span className="tech-eyebrow">Другие сценарии</span><h2>Посмотрите остальные возможности в работе</h2><p>Отключение шага, переписка с посетителем и подготовка счёта — в коротких записях интерфейса.</p></div><ScenarioPlayer initialScenario="workflow" /></section>
      <details className="tech-more" id="technical-details"><summary><span>Архитектура и исходный код<small>React · Go · Python · LangGraph · gRPC · PostgreSQL</small></span><ChevronDown size={20} /></summary><TechArchitecture /><TechCodeTour /></details>
      <footer className="tech-footer"><div><h2>Обсудим ваш проект?</h2></div><a className="tech-cta" href="mailto:hello@riders.agency">hello@riders.agency <ArrowUpRight size={18} /></a><div className="tech-footer-bottom"><a href="https://riders.agency">Riders Agency</a><Link href="/dashboard">Войти в Pulse</Link></div></footer>
    </main>
  );
}
