#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

violations=0

required=(
  "AGENTS.md"
  ".vibe/profile.yaml"
  ".vibe/enabled.yaml"
  ".vibe/registry.lock"
  "standards/active/AGENTS.md"
  "docs/architecture/system.md"
  "docs/architecture/module-boundaries.md"
  "docs/architecture/runtime-boundaries.md"
  "docs/architecture/ai-boundary.md"
  "docs/architecture/migration-plan.md"
  "docs/architecture/module-ownership.tsv"
  "docs/architecture/module-ownership-exceptions.tsv"
  "apps/api/internal/platform/redact/redact.go"
  "scripts/check-go-quality.sh"
)

for item in "${required[@]}"; do
  if [ ! -e "$item" ]; then
    echo "VIOLATION missing_required_path path=$item"
    violations=$((violations + 1))
  fi
done

check_disabled_path() {
  local feature="$1"
  shift
  if awk -v feature="$feature" '
    $0 == "disabled:" { in_disabled = 1; next }
    in_disabled && $0 !~ /^ / { in_disabled = 0 }
    in_disabled && $0 == "  " feature ": true" { found = 1 }
    END { exit(found ? 0 : 1) }
  ' .vibe/profile.yaml 2>/dev/null; then
    for item in "$@"; do
      if [ -e "$item" ]; then
        echo "VIOLATION disabled_feature_path_exists feature=$feature path=$item"
        violations=$((violations + 1))
      fi
    done
  fi
}

check_disabled_path mobile apps/mobile mobile
check_disabled_path desktop apps/desktop desktop
check_disabled_path worker apps/worker worker
check_disabled_path admin apps/admin admin
check_disabled_path payments apps/payments packages/payments payments
check_disabled_path queue apps/queue queue
check_disabled_path ai apps/ai ai

if [ -d apps/ai ]; then
  echo "VIOLATION forbidden_ai_app_path path=apps/ai expected=apps/ai-agent"
  violations=$((violations + 1))
fi

if [ -d apps/worker ]; then
  echo "VIOLATION forbidden_worker_app_path path=apps/worker expected=apps/api/cmd/worker"
  violations=$((violations + 1))
fi

if [ -d registry ] || [ -d frameworks ]; then
  echo "VIOLATION generated_project_contains_registry_or_legacy_frameworks"
  violations=$((violations + 1))
fi

if [ -d contracts/openapi ]; then
  for item in contracts/openapi/openapi.root.yaml contracts/openapi/dist/openapi.yaml contracts/openapi/endpoints.inventory.tsv contracts/openapi/modules contracts/openapi/components contracts/openapi/redocly.yaml scripts/check-openapi-contracts.sh apps/api/internal/platform/httpserver/openapi_pulse_stub.yaml; do
    if [ ! -e "$item" ]; then
      echo "VIOLATION missing_openapi_path path=$item"
      violations=$((violations + 1))
    fi
  done
fi

if [ -d contracts/proto ]; then
  for item in contracts/proto/buf.yaml contracts/proto/descriptors/ai_agent_v1.binpb contracts/proto/fixtures/ai_agent_v1_serialization.json scripts/check-proto-contracts.sh scripts/gen-proto.sh; do
    if [ ! -e "$item" ]; then
      echo "VIOLATION missing_proto_contract_path path=$item"
      violations=$((violations + 1))
    fi
  done
  if [ -f scripts/gen-proto.sh ] && grep -En 'skipped (Go|Python|gRPC|proto)|not found; skipped' scripts/gen-proto.sh; then
    echo "VIOLATION proto_generation_must_fail_closed"
    violations=$((violations + 1))
  fi
fi

if [ -f contracts/openapi/endpoints.inventory.tsv ] && [ -f apps/api/internal/platform/httpserver/doc_inventory.tsv ]; then
  contracts_sorted="$(mktemp)"
  runtime_sorted="$(mktemp)"
  grep -v '^#' contracts/openapi/endpoints.inventory.tsv | sed '/^[[:space:]]*$/d' | sort > "$contracts_sorted"
  grep -v '^#' apps/api/internal/platform/httpserver/doc_inventory.tsv | sed '/^[[:space:]]*$/d' | sort > "$runtime_sorted"
  if ! diff -u "$contracts_sorted" "$runtime_sorted"; then
    echo "VIOLATION openapi_inventory_runtime_inventory_drift"
    violations=$((violations + 1))
  fi
  rm -f "$contracts_sorted" "$runtime_sorted"
fi

if [ -f apps/api/cmd/check-route-inventory/main.go ] && [ -f contracts/openapi/endpoints.inventory.tsv ]; then
  (cd apps/api && go run ./cmd/check-route-inventory --inventory ../../contracts/openapi/endpoints.inventory.tsv)
fi

if [ -f apps/api/cmd/check-openapi-inventory/main.go ] && [ -f contracts/openapi/dist/openapi.yaml ] && [ -f contracts/openapi/endpoints.inventory.tsv ]; then
  (cd apps/api && go run ./cmd/check-openapi-inventory --inventory ../../contracts/openapi/endpoints.inventory.tsv --openapi ../../contracts/openapi/dist/openapi.yaml)
fi

if [ -f apps/api/cmd/check-module-contracts/main.go ] && [ -f docs/architecture/module-ownership.tsv ]; then
  (cd apps/api && go run ./cmd/check-module-contracts)
fi

if [ -f apps/api/cmd/check-tenant-isolation/main.go ] && [ -f docs/architecture/tenant-isolation-catalog.tsv ]; then
  (cd apps/api && go run ./cmd/check-tenant-isolation --inventory internal/platform/httpserver/doc_inventory.tsv --catalog ../../docs/architecture/tenant-isolation-catalog.tsv)
fi

if [ -x deploy/check-migration-policy.sh ]; then
  deploy/check-migration-policy.sh
fi

if [ -f apps/api/cmd/check-frontend-api-paths/main.go ] && [ -d apps/web/src ] && [ -f contracts/openapi/endpoints.inventory.tsv ]; then
  (cd apps/api && go run ./cmd/check-frontend-api-paths --inventory ../../contracts/openapi/endpoints.inventory.tsv --web-root ../web/src)
fi

if [ -f apps/web/src/lib/api/routes.ts ]; then
  echo "VIOLATION frontend_global_api_routes_registry path=apps/web/src/lib/api/routes.ts expected=module-owned route contracts"
  violations=$((violations + 1))
fi

if command -v rg >/dev/null 2>&1; then
  if rg -n '\b[a-zA-Z0-9_]+module\.Mount[a-zA-Z0-9_]*HTTP\(' apps/api/internal/platform/app --glob '*.go'; then
    echo "VIOLATION app_http_composition_calls_module_mount expected=module RegisterHTTP ownership"
    violations=$((violations + 1))
  fi

  if rg -n '"riders\.agency/pulse/api/internal/modules/.*/transport/http"' apps/api/internal/platform/app/http_modules.go apps/api/internal/platform/httpserver --glob '*.go' --glob '!dependencies.go'; then
    echo "VIOLATION platform_imports_module_http_transport expected=module-owned route registration"
    violations=$((violations + 1))
  fi
fi

if command -v rg >/dev/null 2>&1; then
  secret_shape_pattern='BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY|AKIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{35}|xox[baprs]-|ghp_[0-9A-Za-z]{36}|github_pat_[0-9A-Za-z_]{80,}|sk-[A-Za-z0-9]{20,}'
  if rg -n --glob '!node_modules/**' --glob '!.git/**' --glob '!dist/**' --glob '!scripts/check.sh' "$secret_shape_pattern" .; then
    echo "VIOLATION generated_project_secret_shape"
    violations=$((violations + 1))
  fi
  if [ -d apps/ai-agent/src ]; then
    ai_business_sql_pattern='(FROM|JOIN|UPDATE|INSERT[[:space:]]+INTO|DELETE[[:space:]]+FROM)[[:space:]]+(tasks|clients|workflows|payments|action_dependencies|workflow_runs|action_runs|run_events)\b'
    if rg -n -i --glob '*.py' --glob '!**/proto/**' "$ai_business_sql_pattern" apps/ai-agent/src; then
      echo "VIOLATION ai_agent_direct_business_table_access"
      violations=$((violations + 1))
    fi
    if rg -n --glob '*.py' --glob '!**/proto/**' 'ExecuteTool|def execute_tool|def execute_read_tool' apps/ai-agent/src; then
      echo "VIOLATION ai_agent_generic_tool_execution_boundary"
      violations=$((violations + 1))
    fi
  fi
  if [ -f apps/ai-agent/src/pulse_agent/security.py ]; then
    for redaction_key in api_key authorization cookie password secret token vless_url x-pulse-service-token; do
      if ! rg -q --fixed-strings "\"$redaction_key\"" apps/ai-agent/src/pulse_agent/security.py; then
        echo "VIOLATION ai_agent_redaction_key_missing key=$redaction_key"
        violations=$((violations + 1))
      fi
    done
  fi
  if [ -f apps/api/internal/platform/redact/redact.go ]; then
    for redaction_key in api_key authorization cookie password secret signature token vless_url; do
      if ! rg -q --fixed-strings "\"$redaction_key\"" apps/api/internal/platform/redact/redact.go; then
        echo "VIOLATION api_redaction_key_missing key=$redaction_key"
        violations=$((violations + 1))
      fi
    done
  fi
  if rg -n --glob 'engine.go' 'ai\.agent\.run|ai_agent_run|AI agent run' apps/api/internal/executor; then
    echo "VIOLATION ai_action_added_to_legacy_executor"
    violations=$((violations + 1))
  fi
  if [ -d deploy ] && rg -n 'AI_AGENT_INSECURE_DEV=true' deploy --glob '*.service' --glob '*.sh'; then
    echo "VIOLATION production_ai_agent_insecure_dev"
    violations=$((violations + 1))
  fi
  runtime_source_dirs=(scripts docs apps)
  if [ -d deploy ]; then runtime_source_dirs+=(deploy); fi
  if rg -n 'EnvironmentFile=/etc/riders-pulse|(^|[[:space:]])source[[:space:]]+/etc/riders-pulse|godotenv\.Load\("/etc/riders-pulse|runtime_env=.*?/etc/riders-pulse' "${runtime_source_dirs[@]}" --glob '!scripts/check.sh'; then
    echo "VIOLATION external_riders_pulse_env_path"
    violations=$((violations + 1))
  fi
  if rg -n 'TriggerSyncer|EnqueueDueSchedules|EnqueueDueClientPayments' apps/api/internal/runtime apps/api/internal/executor; then
    echo "VIOLATION legacy_trigger_syncer_boundary"
    violations=$((violations + 1))
  fi
  if rg -n 'RegisterCompatibilityTriggers|CompatibilityTriggerSource|ScheduleTriggerSource|ClientPaymentDueTriggerSource' apps/api/internal/runtime apps/api/internal/executor; then
    echo "VIOLATION legacy_trigger_source_boundary"
    violations=$((violations + 1))
  fi
  if rg -n 'syncScheduleStates|enqueueOneDueSchedule|syncClientPaymentDueStates|enqueueOneDueClientPayment|enqueueExistingPaymentActionReminder|enqueueNewPaymentActionReminder|enqueueBatchPaymentDueRun' apps/api/internal/executor --glob '*.go'; then
    echo "VIOLATION executor_trigger_queue_implementation"
    violations=$((violations + 1))
  fi
  if sed -n '/func (e \*Engine) registerExecutorRuntimeActions()/,/^}/p' apps/api/internal/executor/registered_actions.go | rg -n 'telegram\.send_message|max\.send_message|email\.send|clients\.manual_payment\.trigger|clients\.payment_due\.trigger|clients\.payment_cycle\.collect'; then
    echo "VIOLATION executor_registers_module_owned_action"
    violations=$((violations + 1))
  fi
  if sed -n '/func (e \*Engine) executeCompatibility(/,/^}/p' apps/api/internal/executor/registered_actions.go | rg -n 'telegram\.send_message|max\.send_message|email\.send|executeTelegram|executeMax|executeEmail'; then
    echo "VIOLATION executor_compatibility_executes_provider_action"
    violations=$((violations + 1))
  fi
  if sed -n '/func (e \*Engine) executeCompatibility(/,/^}/p' apps/api/internal/executor/registered_actions.go | rg -n 'clients\.manual_payment\.trigger|clients\.payment_due\.trigger|clients\.payment_cycle\.collect|executeManualPaymentTrigger|executePaymentDueTrigger|executePaymentCycleCollect'; then
    echo "VIOLATION executor_compatibility_executes_client_payment_action"
    violations=$((violations + 1))
  fi
  if rg -n 'executeTelegram|executeMax|executeEmail' apps/api/internal/executor --glob '*.go'; then
    echo "VIOLATION executor_provider_action_implementation"
    violations=$((violations + 1))
  fi
  if rg -n 'executeActionLegacy|ExecuteCompatibilityAction|executeCompatibility|executorRuntimeAction|CompatibilityActionRunner' apps/api/internal/executor --glob '*.go'; then
    echo "VIOLATION executor_compatibility_fallback"
    violations=$((violations + 1))
  fi
  if sed -n '/func (e \*Engine) registerExecutorRuntimeActions()/,/^}/p' apps/api/internal/executor/registered_actions.go | rg -n 'webhook\.trigger|manual\.trigger|schedule\.trigger|telegram\.relay\.trigger|module\.event\.trigger'; then
    echo "VIOLATION executor_registers_trigger_pseudo_action"
    violations=$((violations + 1))
  fi
  if rg -n 'type .*TriggerHandler|type scheduleTriggerHandler|type clientPaymentDueTriggerHandler|type eventDrivenTriggerHandler' apps/api/internal/executor --glob '*.go'; then
    echo "VIOLATION executor_owns_trigger_handlers"
    violations=$((violations + 1))
  fi
  if sed -n '/func (e \*Engine) executeActionLegacy(/,/^}/p' apps/api/internal/executor/engine.go | rg -n 'executeHTTP|executeCondition|actionOutput\(node\)|providerResponse\(node\)|case "http\.request"|case "condition\.branch"'; then
    echo "VIOLATION executor_legacy_action_fallback"
    violations=$((violations + 1))
  fi
  if rg -n '"riders\.agency/pulse/api/internal/executor"' apps/api/internal/modules/providers apps/api/internal/modules/clients --glob '*.go'; then
    echo "VIOLATION module_runtime_imports_executor"
    violations=$((violations + 1))
  fi
  if rg -n 'Executor[[:space:]]+\*executor\.Engine|Executor[[:space:]]+executor\.Engine' apps/api/internal/platform/app/module.go; then
    echo "VIOLATION app_module_deps_exposes_executor_engine"
    violations=$((violations + 1))
  fi
  if rg -n 'CompatibilityActionRunner|RuntimeActionRunner|ActionRunner' apps/api/internal/platform/app apps/api/internal/modules/providers apps/api/internal/modules/clients --glob '*.go'; then
    echo "VIOLATION module_runtime_uses_compatibility_runner_type"
    violations=$((violations + 1))
  fi
  if rg -n 'INSERT INTO outbox_events' apps/api/internal apps/api/cmd --glob '*.go' --glob '!apps/api/internal/platform/outbox/repository.go'; then
    echo "VIOLATION domain_direct_outbox_insert"
    violations=$((violations + 1))
  fi
  if rg -n 'INSERT INTO workflow_runs' apps/api/internal --glob '*.go' --glob '!apps/api/internal/runtime/runqueue/*.go'; then
    echo "VIOLATION workflow_run_direct_insert"
    violations=$((violations + 1))
  fi
  if rg -n 'runqueue\.Enqueue(Tx)?\(' apps/api/internal --glob '*.go'; then
    echo "VIOLATION workflow_run_legacy_enqueue_api"
    violations=$((violations + 1))
  fi
  if rg -n 'next_retry_at[[:space:]]*=[[:space:]]*clock_timestamp\(\)' apps/api/internal/runtime apps/api/internal/modules --glob '*.go'; then
    echo "VIOLATION workflow_action_retry_without_backoff"
    violations=$((violations + 1))
  fi
  if [ -d apps/web/src/features/workflow-editor ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete)' apps/web/src/features/workflow-editor --glob '!**/api/**'; then
      echo "VIOLATION workflow_editor_api_call_outside_api_adapter"
      violations=$((violations + 1))
    fi
    if rg -n 'style=\{\{' apps/web/src/features/workflow-editor apps/web/src/app/\(pulse\)/workflows; then
      echo "VIOLATION workflow_editor_inline_style"
      violations=$((violations + 1))
    fi
  fi
  if [ -d apps/web/src/features/work ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(' apps/web/src/features/work --glob '!**/api/**'; then
      echo "VIOLATION work_feature_api_call_outside_api_adapter"
      violations=$((violations + 1))
    fi
    if rg -n 'style=\{\{' apps/web/src/features/work; then
      echo "VIOLATION work_feature_inline_style"
      violations=$((violations + 1))
    fi
  fi
  if [ -d apps/web/src/app ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{|/api/v1' apps/web/src/app --glob '*.ts' --glob '*.tsx'; then
      echo "VIOLATION app_layer_legacy_transport_or_inline_style"
      violations=$((violations + 1))
    fi
  fi
  if [ -d apps/ai-agent/src/pulse_agent ]; then
    if rg -n 'grpc\.aio\.insecure_channel' apps/ai-agent/src/pulse_agent --glob '*.py' --glob '!**/proto/**' --glob '!**/clients/grpc_security.py'; then
      echo "VIOLATION ai_agent_direct_insecure_grpc_channel"
      violations=$((violations + 1))
    fi
    if rg -n 'legacy = plan_turn|legacy\.__dict__' apps/ai-agent/src/pulse_agent/graph/builder.py; then
      echo "VIOLATION ai_agent_legacy_planner_name"
      violations=$((violations + 1))
    fi
  fi
  if rg -n 'ReadLangGraph[A-Za-z]+' apps/api/internal/modules/ai/transport/http --glob '*handlers.go'; then
    echo "VIOLATION ai_http_handler_bypasses_tool_gateway"
    violations=$((violations + 1))
  fi
  for feature_path in \
    apps/web/src/features/ai-employee \
    apps/web/src/features/client-service-detail \
    apps/web/src/features/company-clients \
    apps/web/src/features/project-integrations; do
    if [ -d "$feature_path" ] && rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(' "$feature_path" --glob '!**/api/**'; then
      echo "VIOLATION feature_api_call_outside_api_adapter path=$feature_path"
      violations=$((violations + 1))
    fi
    if [ -d "$feature_path" ] && rg -n 'style=\{\{' "$feature_path"; then
      echo "VIOLATION feature_inline_style path=$feature_path"
      violations=$((violations + 1))
    fi
  done
  if [ -f 'apps/web/src/app/(pulse)/site-chat/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete)|apiBase|fetch\(' 'apps/web/src/app/(pulse)/site-chat/page.tsx'; then
      echo "VIOLATION site_chat_route_direct_api_call"
      violations=$((violations + 1))
    fi
  fi
  if [ -d apps/web/src/features/site-chat-console ]; then
    if rg -n 'style=\{\{' apps/web/src/features/site-chat-console 'apps/web/src/app/(pulse)/site-chat/page.tsx'; then
      echo "VIOLATION site_chat_console_inline_style"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/runs/[workflowRunId]/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete)|style=\{\{' 'apps/web/src/app/(pulse)/runs/[workflowRunId]/page.tsx'; then
      echo "VIOLATION workflow_run_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/clients/[clientId]/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/clients/[clientId]/page.tsx'; then
      echo "VIOLATION client_detail_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/projects/[projectId]/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/projects/[projectId]/page.tsx'; then
      echo "VIOLATION project_detail_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/client-services/[clientServiceId]/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/client-services/[clientServiceId]/page.tsx'; then
      echo "VIOLATION client_service_detail_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/projects/[projectId]/integrations/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/projects/[projectId]/integrations/page.tsx'; then
      echo "VIOLATION project_integrations_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/companies/[companyId]/clients/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/companies/[companyId]/clients/page.tsx'; then
      echo "VIOLATION company_clients_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/projects/[projectId]/payments/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/projects/[projectId]/payments/page.tsx'; then
      echo "VIOLATION project_payments_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/payment-actions/[paymentActionId]/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/payment-actions/[paymentActionId]/page.tsx'; then
      echo "VIOLATION payment_action_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/webhook-events/[webhookEventId]/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/webhook-events/[webhookEventId]/page.tsx'; then
      echo "VIOLATION webhook_event_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/webhook-endpoints/[webhookEndpointId]/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/webhook-endpoints/[webhookEndpointId]/page.tsx'; then
      echo "VIOLATION webhook_endpoint_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  if [ -f 'apps/web/src/app/(pulse)/status/page.tsx' ]; then
    if rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' 'apps/web/src/app/(pulse)/status/page.tsx'; then
      echo "VIOLATION status_route_legacy_logic"
      violations=$((violations + 1))
    fi
  fi
  for route_path in \
    'apps/web/src/app/(pulse)/account/page.tsx' \
    'apps/web/src/app/(pulse)/dashboard/page.tsx' \
    'apps/web/src/app/(pulse)/companies/page.tsx' \
    'apps/web/src/app/(pulse)/companies/[companyId]/page.tsx' \
    'apps/web/src/app/(pulse)/companies/[companyId]/projects/page.tsx' \
    'apps/web/src/app/(pulse)/companies/[companyId]/projects/new/page.tsx' \
    'apps/web/src/app/(pulse)/companies/[companyId]/clients/new/page.tsx' \
    'apps/web/src/app/(pulse)/companies/[companyId]/legal-entities/page.tsx' \
    'apps/web/src/app/(pulse)/companies/new/page.tsx' \
    'apps/web/src/app/(pulse)/clients/[clientId]/services/new/page.tsx' \
    'apps/web/src/app/(pulse)/projects/[projectId]/integrations/new/page.tsx' \
    'apps/web/src/app/(pulse)/projects/[projectId]/integrations/[integrationId]/edit/page.tsx' \
    'apps/web/src/app/(pulse)/projects/[projectId]/workflows/page.tsx' \
    'apps/web/src/app/(pulse)/projects/[projectId]/workflows/new/page.tsx'; do
    if [ -f "$route_path" ] && rg -n 'api(Get|Post|Put|Patch|Delete|Upload)<?|apiBase|fetch\(|style=\{\{' "$route_path"; then
      echo "VIOLATION route_legacy_logic path=$route_path"
      violations=$((violations + 1))
    fi
  done
fi

check_file_max_lines() {
  local path="$1"
  local max_lines="$2"
  if [ ! -f "$path" ]; then
    return
  fi
  local lines
  lines="$(wc -l < "$path" | tr -d ' ')"
  if [ "$lines" -gt "$max_lines" ]; then
    echo "VIOLATION file_line_budget_exceeded path=$path lines=$lines max=$max_lines"
    violations=$((violations + 1))
  fi
}

if [ -f .vibe/size-budgets.tsv ]; then
  while IFS=$'\t' read -r budget_path budget_max _budget_note; do
    if [ -z "${budget_path:-}" ] || [[ "$budget_path" == \#* ]]; then
      continue
    fi
    check_file_max_lines "$budget_path" "$budget_max"
  done < .vibe/size-budgets.tsv
fi

if [ -d apps/api/internal/modules/providers/runtime ]; then
  if rg -n 'func \(r \*Runner\) Execute|switch strings\.TrimSpace\(req\.ActionTypeKey\)' apps/api/internal/modules/providers/runtime --glob '*.go'; then
    echo "VIOLATION provider_runtime_central_switch"
    violations=$((violations + 1))
  fi
fi

if [ -d apps/api/internal/modules/clients/runtime ]; then
  if rg -n 'func \(r \*Runner\) execute\(ctx context\.Context, typeKey string|switch strings\.TrimSpace\(typeKey\)' apps/api/internal/modules/clients/runtime --glob '*.go'; then
    echo "VIOLATION client_runtime_central_switch"
    violations=$((violations + 1))
  fi
fi

if [ -d db/migrations ]; then
  while IFS= read -r up_file; do
    down_file="${up_file%.up.sql}.down.sql"
    if [ ! -f "$down_file" ]; then
      echo "VIOLATION migration_missing_down up=$up_file"
      violations=$((violations + 1))
    fi
  done < <(find db/migrations -maxdepth 1 -name '*.up.sql' | sort)
  while IFS= read -r down_file; do
    up_file="${down_file%.down.sql}.up.sql"
    if [ ! -f "$up_file" ]; then
      echo "VIOLATION migration_missing_up down=$down_file"
      violations=$((violations + 1))
    fi
  done < <(find db/migrations -maxdepth 1 -name '*.down.sql' | sort)
fi

if [ "$violations" -gt 0 ]; then
  echo "Generated project check failed: violations=$violations"
  exit 1
fi

if [ -f apps/api/cmd/ai-scenario-runner/main.go ]; then
  (cd apps/api && go run ./cmd/ai-scenario-runner)
fi

echo "Generated project check passed"
