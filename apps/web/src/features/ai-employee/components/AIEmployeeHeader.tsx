import { Maximize2, Minimize2, X } from "lucide-react";

export function AIEmployeeHeader({
  fullscreen,
  onToggleFullscreen,
  onClose,
}: {
  fullscreen: boolean;
  onToggleFullscreen: () => void;
  onClose: () => void;
}) {
  return (
    <div className="ai-employee-head">
      <div className="ai-employee-title">
        <h2>ИИ-сотрудник</h2><span className="ai-header-description">Помощник в контексте вашего проекта</span>
      </div>
      <button
        className="icon-button ai-employee-head-button"
        type="button"
        onClick={onToggleFullscreen}
        aria-label={fullscreen ? "Свернуть из полноэкранного режима" : "Открыть на весь экран"}
      >
        {fullscreen ? <Minimize2 size={18} /> : <Maximize2 size={18} />}
      </button>
      <button className="icon-button ai-employee-close" type="button" onClick={onClose} aria-label="Закрыть">
        <X size={18} />
      </button>
    </div>
  );
}
