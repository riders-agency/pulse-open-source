package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := requiredEnv("DATABASE_URL")
	passwordHash := hashSecret(requiredEnv("PULSE_DEMO_PASSWORD"), "PULSE_DEMO_PASSWORD")
	hookSecretHash := hashSecret(requiredEnv("PULSE_DEMO_WEBHOOK_SECRET"), "PULSE_DEMO_WEBHOOK_SECRET")

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fail("db connect: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		fail("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE pulse_demo_seed_input (
  password_hash text NOT NULL,
  hook_secret_hash text NOT NULL
) ON COMMIT DROP`); err != nil {
		fail("seed input table: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO pulse_demo_seed_input (password_hash, hook_secret_hash) VALUES ($1, $2)`, passwordHash, hookSecretHash); err != nil {
		fail("seed input: %v", err)
	}
	if _, err := tx.Exec(ctx, demoSeedSQL); err != nil {
		fail("seed demo tenant: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		fail("commit: %v", err)
	}

	fmt.Println("seeded demo tenant demo@riders-pulse.local")
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		fail("%s is required", key)
	}
	return value
}

func hashSecret(value string, key string) string {
	if len(value) < 12 {
		fail("%s must be at least 12 characters", key)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		fail("hash %s: %v", key, err)
	}
	return string(hash)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

const demoSeedSQL = `
DO $pulse_demo$
DECLARE
  v_user    uuid := '11111111-1111-4111-8111-111111111101'::uuid;
  v_company uuid := '11111111-1111-4111-8111-111111111102'::uuid;
  v_project uuid := '11111111-1111-4111-8111-111111111103'::uuid;
  v_wf      uuid := '11111111-1111-4111-8111-111111111104'::uuid;
  v_wfv     uuid := '11111111-1111-4111-8111-111111111105'::uuid;
  v_hook    uuid := '11111111-1111-4111-8111-111111111106'::uuid;
  v_etpl    uuid := '11111111-1111-4111-8111-111111111107'::uuid;
  v_password_hash text;
  v_hook_secret_hash text;
  v_def     jsonb := '{
    "nodes": [
      {"id":"t1","action_key":"webhook","action_type_key":"webhook.trigger"},
      {"id":"c1","action_key":"branch","action_type_key":"condition.branch"},
      {"id":"m1","action_key":"max1","action_type_key":"max.send_message"},
      {"id":"tg1","action_key":"tg1","action_type_key":"telegram.send_message"},
      {"id":"em1","action_key":"em1","action_type_key":"email.send"},
      {"id":"s1","action_key":"stop","action_type_key":"stop"}
    ],
    "edges": [
      {"from":"t1","to":"c1"},
      {"from":"c1","to":"m1"},
      {"from":"m1","to":"tg1"},
      {"from":"tg1","to":"em1"},
      {"from":"em1","to":"s1"}
    ]
  }'::jsonb;
BEGIN
  SELECT password_hash, hook_secret_hash
    INTO v_password_hash, v_hook_secret_hash
    FROM pulse_demo_seed_input
   LIMIT 1;

  IF EXISTS (
    SELECT 1 FROM users
     WHERE email = 'demo@riders-pulse.local'::citext
       AND user_id <> v_user
       AND deleted_at IS NULL
  ) THEN
    RAISE EXCEPTION 'demo email is already owned by another active user';
  END IF;

  INSERT INTO users (
    user_id, email, password_hash, first_name, display_name, is_activated, deleted_at
  ) VALUES (
    v_user, 'demo@riders-pulse.local', v_password_hash, 'Денис', 'Денис', true, NULL
  )
  ON CONFLICT (user_id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    first_name = EXCLUDED.first_name,
    display_name = EXCLUDED.display_name,
    is_activated = true,
    deleted_at = NULL,
    updated_at = now();

  INSERT INTO companies (
    company_id, owner_user_id, created_by_user_id, slug, name, status
  ) VALUES (
    v_company, v_user, v_user, 'demo-pulse-co', 'Демо-компания Pulse', 'active'
  )
  ON CONFLICT (company_id) DO UPDATE SET
    owner_user_id = EXCLUDED.owner_user_id,
    created_by_user_id = EXCLUDED.created_by_user_id,
    slug = EXCLUDED.slug,
    name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now();

  UPDATE users SET current_company_id = v_company WHERE user_id = v_user;

  INSERT INTO company_memberships (company_id, user_id, role)
  VALUES (v_company, v_user, 'owner')
  ON CONFLICT (company_id, user_id) DO UPDATE SET role = EXCLUDED.role;

  INSERT INTO projects (
    project_id, company_id, name, slug, hook_key, hook_secret_hash, created_by_user_id
  ) VALUES (
    v_project, v_company, 'Демонстрационный проект', 'demo-logistics', 'demo-orders-paid-hook', v_hook_secret_hash, v_user
  )
  ON CONFLICT (project_id) DO UPDATE SET
    company_id = EXCLUDED.company_id,
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    hook_key = EXCLUDED.hook_key,
    hook_secret_hash = EXCLUDED.hook_secret_hash,
    created_by_user_id = EXCLUDED.created_by_user_id,
    updated_at = now();

  INSERT INTO project_members (project_id, user_id, role)
  VALUES (v_project, v_user, 'owner')
  ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role;

  INSERT INTO workflows (
    workflow_id, project_id, name, slug, status, draft_definition_json, created_by_user_id
  ) VALUES (
    v_wf, v_project, 'New Order Pulse', 'new-order-pulse', 'active', v_def, v_user
  )
  ON CONFLICT (workflow_id) DO UPDATE SET
    project_id = EXCLUDED.project_id,
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    status = EXCLUDED.status,
    draft_definition_json = EXCLUDED.draft_definition_json,
    created_by_user_id = EXCLUDED.created_by_user_id,
    updated_at = now();

  INSERT INTO workflow_versions (
    workflow_version_id, project_id, workflow_id, version_no, status, definition_json, created_by_user_id, published_at
  ) VALUES (
    v_wfv, v_project, v_wf, 1, 'published', v_def, v_user, now()
  )
  ON CONFLICT (workflow_version_id) DO UPDATE SET
    project_id = EXCLUDED.project_id,
    workflow_id = EXCLUDED.workflow_id,
    status = EXCLUDED.status,
    definition_json = EXCLUDED.definition_json,
    created_by_user_id = EXCLUDED.created_by_user_id,
    published_at = EXCLUDED.published_at;

  UPDATE workflows SET current_workflow_version_id = v_wfv WHERE workflow_id = v_wf;

  INSERT INTO webhook_endpoints (
    webhook_endpoint_id, project_id, workflow_id, workflow_version_id,
    name, endpoint_key, status, auth_type, created_by_user_id
  ) VALUES (
    v_hook, v_project, v_wf, v_wfv, 'Orders Paid', 'orders-paid', 'active', 'project_secret', v_user
  )
  ON CONFLICT (webhook_endpoint_id) DO UPDATE SET
    project_id = EXCLUDED.project_id,
    workflow_id = EXCLUDED.workflow_id,
    workflow_version_id = EXCLUDED.workflow_version_id,
    name = EXCLUDED.name,
    endpoint_key = EXCLUDED.endpoint_key,
    status = EXCLUDED.status,
    auth_type = EXCLUDED.auth_type,
    created_by_user_id = EXCLUDED.created_by_user_id,
    updated_at = now();

  INSERT INTO email_templates (
    email_template_id, project_id, name, slug, subject_template, body_template, body_format, created_by_user_id
  ) VALUES (
    v_etpl, v_project, 'Order Paid', 'order-paid',
    'Order {{order_id}} paid',
    '<p>Thank you. Order {{order_id}} total {{order_total}}.</p>',
    'html', v_user
  )
  ON CONFLICT (email_template_id) DO UPDATE SET
    project_id = EXCLUDED.project_id,
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    subject_template = EXCLUDED.subject_template,
    body_template = EXCLUDED.body_template,
    body_format = EXCLUDED.body_format,
    created_by_user_id = EXCLUDED.created_by_user_id,
    updated_at = now();
END;
$pulse_demo$;
`
