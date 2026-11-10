import { Play } from "lucide-react";

import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { ChangeDescriptionControl } from "./panels";
import type { Project, Workflow } from "../../../lib/api";

type EditorToolbarProps = {
  workflow: Workflow | null;
  project: Project | null;
  notice: string;
  error: string;
  versionDescription: string;
  descriptionOpen: boolean;
  testBusy: "full" | "through_action" | "selected_only" | "";
  setVersionDescription: (value: string) => void;
  setDescriptionOpen: (value: boolean) => void;
  save: () => void;
  validate: () => void;
  publish: () => void;
  runTest: (mode: "full" | "through_action" | "selected_only") => void;
  openRuns: () => void;
};

export function EditorToolbar({
  workflow,
  project,
  notice,
  error,
  versionDescription,
  descriptionOpen,
  testBusy,
  setVersionDescription,
  setDescriptionOpen,
  save,
  validate,
  publish,
  runTest,
  openRuns,
}: EditorToolbarProps) {
  return (
    <div className="flow-toolbar">
      <div>
        <strong>{workflow?.name ?? "Сценарий"}</strong>
        <div className="flow-toolbar-subtitle muted">
          {project?.name ?? "Проект"} · визуальный редактор
        </div>
      </div>
      <div className="row">
        {notice ? <Badge tone="success">{notice}</Badge> : null}
        {error ? <Badge tone="danger">{error}</Badge> : null}
        <ChangeDescriptionControl
          value={versionDescription}
          open={descriptionOpen}
          setValue={setVersionDescription}
          setOpen={setDescriptionOpen}
        />
        <Button variant="outline" onClick={save}>Сохранить</Button>
        <Button variant="secondary" onClick={validate}>Проверить</Button>
        <Button variant="secondary" disabled={!!testBusy} onClick={() => runTest("full")}>
          <Play size={15} />
          {testBusy === "full" ? "Запуск…" : "Тест черновика"}
        </Button>
        <Button onClick={publish}>Опубликовать</Button>
        <Button variant="outline" onClick={openRuns}>Запуски</Button>
      </div>
    </div>
  );
}
