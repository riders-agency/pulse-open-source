import type { ActionConfig } from "./types";

export type DefinitionNodeJSON = {
  action_key?: string;
  label?: string;
  action_type_key?: string;
  config?: ActionConfig;
};

export type RawGraphNodeJSON = {
  id?: string;
  action_key?: string;
  type?: string;
  position?: { x?: number; y?: number };
  label?: string;
  action_type_key?: string;
  config?: ActionConfig;
  data?: { label?: string; action_type_key?: string; config?: ActionConfig };
  integration?: { label?: string; action_type_key?: string; config?: ActionConfig };
};
