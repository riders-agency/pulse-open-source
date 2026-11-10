CREATE TABLE workflow_office_workspaces (
  project_id uuid PRIMARY KEY REFERENCES projects(project_id) ON DELETE CASCADE,
  workflow_id uuid NOT NULL UNIQUE REFERENCES workflows(workflow_id) ON DELETE CASCADE,
  owner_user_id uuid NOT NULL REFERENCES users(user_id),
  enabled boolean NOT NULL DEFAULT true,
  next_poll_at timestamptz NOT NULL DEFAULT now() + interval '30 seconds',
  last_poll_at timestamptz,
  poll_count bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE workflow_office_messages (
  workflow_office_message_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES workflow_office_workspaces(project_id) ON DELETE CASCADE,
  delivery_key uuid NOT NULL,
  subject text NOT NULL,
  body text NOT NULL,
  sender text NOT NULL,
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  workflow_run_id uuid UNIQUE REFERENCES workflow_runs(workflow_run_id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (project_id, delivery_key)
);
CREATE INDEX idx_workflow_office_messages_pending ON workflow_office_messages(project_id, created_at) WHERE workflow_run_id IS NULL;
CREATE TABLE workflow_office_decisions (
  workflow_run_id uuid PRIMARY KEY REFERENCES workflow_runs(workflow_run_id) ON DELETE CASCADE,
  approved boolean NOT NULL,
  decided_by_user_id uuid NOT NULL REFERENCES users(user_id),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE workflow_office_artifacts (
  workflow_run_id uuid PRIMARY KEY REFERENCES workflow_runs(workflow_run_id) ON DELETE CASCADE,
  project_id uuid NOT NULL REFERENCES workflow_office_workspaces(project_id) ON DELETE CASCADE,
  page_json jsonb NOT NULL DEFAULT '{}',
  invoice_json jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_workflow_office_artifacts_project ON workflow_office_artifacts(project_id);
