import { Button } from "../../components/ui/button";
import { Field, Input } from "../../components/ui/input";
import { Select } from "../../components/ui/select";
import type { WorkBoardTemplate } from "../../lib/api";

export function WorkBoardCreateForm({
  boardName,
  createBoard,
  presets,
  selectedPreset,
  setBoardName,
  setError,
  setTemplateKey,
  templateKey,
}: {
  boardName: string;
  createBoard: () => Promise<void>;
  presets: WorkBoardTemplate[];
  selectedPreset?: WorkBoardTemplate;
  setBoardName: (value: string) => void;
  setError: (value: string) => void;
  setTemplateKey: (value: string) => void;
  templateKey: string;
}) {
  return (
    <form
      className="work-create-panel"
      onSubmit={(event) => {
        event.preventDefault();
        void createBoard().catch((err) => setError(err.message));
      }}
    >
      <Field label="Шаблон доски">
        <Select value={templateKey} onChange={(event) => setTemplateKey(event.target.value)}>
          {presets.map((preset) => <option key={preset.key} value={preset.key}>{preset.name}</option>)}
        </Select>
      </Field>
      <Field label="Название">
        <Input value={boardName} onChange={(event) => setBoardName(event.target.value)} placeholder={selectedPreset?.name || "Новая доска"} />
      </Field>
      <p>{selectedPreset?.description}</p>
      <Button type="submit">Создать доску</Button>
    </form>
  );
}
