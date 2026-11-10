package kernel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"riders.agency/pulse/api/internal/platform/demomode"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"riders.agency/pulse/api/internal/runtime/contracts"
	"riders.agency/pulse/api/internal/runtime/effects"
	"riders.agency/pulse/api/internal/runtime/registry"
	runtimeretry "riders.agency/pulse/api/internal/runtime/retry"
	"riders.agency/pulse/api/internal/runtime/runqueue"
)

const (
	defaultProcessInterval      = 400 * time.Millisecond
	defaultTriggerInterval      = 10 * time.Second
	defaultRunLeaseDuration     = 2 * time.Minute
	defaultRunHeartbeatInterval = 15 * time.Second
)

type LegacyRunner interface {
	Run(ctx context.Context)
	Notify()
}

type DirectDueTriggerHandler interface {
	contracts.TriggerHandler
	EnqueueDue(ctx context.Context, limit int) (int, error)
}

type IncidentDetector interface {
	IncidentFromActionExecution(final ActionExecution) (Incident, bool)
}

type Runtime interface {
	BuildRunPlan(ctx context.Context, definitionJSON []byte, runInputJSON []byte, projectID string) (RunPlan, error)
	DefinitionNeedsProjectID(definitionJSON []byte) bool
	EnrichRunInputProjectID(runInputJSON []byte, projectID string) []byte
	ActionInput(node Node, runInputJSON []byte) string
	SkippedOutput(node Node) string
	SkippedOutputWithReason(node Node, reason string) string
	ShouldRunGraphNode(node Node, graph any, completed map[string]bool, branchOutputs map[string]string) bool
	BranchFromOutput(outputJSON string) string
	MergeActionExpose(steps map[string]any, actionKey string, outputJSON string)
	ExecuteAction(ctx context.Context, node Node, runInputJSON []byte, steps map[string]any, runID string, projectID string, attempt int) ActionExecution
}

type Node struct {
	Disabled           bool
	ActionKey          string
	ActionTypeKey      string
	ActionDependencyID *string
	MaxAttempts        int
	RetryInitialDelay  time.Duration
	RetryMaxDelay      time.Duration
	RetryMaxElapsed    time.Duration
	ContinueOnFailure  bool
	FailureRequired    bool
	Payload            any
}

type RunPlan struct {
	Execute       []Node
	Skipped       []Node
	InitialSteps  map[string]any
	UseGraphEdges bool
	Graph         any
}

type ActionExecution struct {
	InputJSON            string
	OutputJSON           string
	ProviderResponseJSON string
	Waiting              bool
	WaitingReason        string
	WaitingDeadline      time.Time
	Retryable            bool
	ErrorType            string
	Err                  error
}

type Incident struct {
	Message     string
	PayloadJSON string
}

type Option func(*Kernel)

type Kernel struct {
	pool                 *pgxpool.Pool
	runtime              Runtime
	triggers             *registry.Registry
	legacy               LegacyRunner
	ch                   chan struct{}
	processInterval      time.Duration
	triggerInterval      time.Duration
	runClaimOwner        string
	runLeaseDuration     time.Duration
	runHeartbeatInterval time.Duration
	runClaimAllowlist    []string
	runClaimFilterActive bool
	effectLedger         *effects.Ledger
	workerCount          int
}

func New(pool *pgxpool.Pool, runtime Runtime, opts ...Option) *Kernel {
	k := &Kernel{
		pool:                 pool,
		runtime:              runtime,
		ch:                   make(chan struct{}, 1),
		processInterval:      defaultProcessInterval,
		triggerInterval:      defaultTriggerInterval,
		runClaimOwner:        defaultRunClaimOwner(),
		runLeaseDuration:     defaultRunLeaseDuration,
		runHeartbeatInterval: defaultRunHeartbeatInterval,
		workerCount:          defaultRuntimeWorkerCount,
	}
	if pool != nil {
		k.effectLedger = &effects.Ledger{P: pool}
	}
	for _, opt := range opts {
		if opt != nil {
			opt(k)
		}
	}
	return k
}

func NewLegacy(runner LegacyRunner) *Kernel {
	return &Kernel{legacy: runner}
}

func WithIntervals(processInterval time.Duration, triggerInterval time.Duration) Option {
	return func(k *Kernel) {
		if processInterval > 0 {
			k.processInterval = processInterval
		}
		if triggerInterval > 0 {
			k.triggerInterval = triggerInterval
		}
	}
}

func WithTriggerRegistry(reg *registry.Registry) Option {
	return func(k *Kernel) {
		k.triggers = reg
	}
}

func WithRunClaimLease(owner string, leaseDuration, heartbeatInterval time.Duration) Option {
	return func(k *Kernel) {
		owner = strings.TrimSpace(owner)
		if owner != "" {
			k.runClaimOwner = owner
		}
		if leaseDuration > 0 {
			k.runLeaseDuration = leaseDuration
		}
		if heartbeatInterval > 0 {
			k.runHeartbeatInterval = heartbeatInterval
		}
	}
}

func WithRunClaimAllowlist(raw string) Option {
	return func(k *Kernel) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		k.runClaimFilterActive = true
		seen := map[string]struct{}{}
		for _, candidate := range strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t'
		}) {
			parsed, err := uuid.Parse(strings.TrimSpace(candidate))
			if err != nil {
				continue
			}
			value := parsed.String()
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			k.runClaimAllowlist = append(k.runClaimAllowlist, value)
		}
	}
}

func (k *Kernel) Run(ctx context.Context) {
	if k == nil {
		return
	}
	if k.legacy != nil {
		k.legacy.Run(ctx)
		return
	}
	if k.pool == nil || k.runtime == nil {
		return
	}
	k.runWorkerPool(ctx)
}

func (k *Kernel) Notify() {
	if k == nil {
		return
	}
	if k.legacy != nil {
		k.legacy.Notify()
		return
	}
	select {
	case k.ch <- struct{}{}:
	default:
	}
}

func (k *Kernel) EnqueueDue(ctx context.Context) {
	if k == nil {
		return
	}
	if k.triggers != nil {
		k.EnqueueRegisteredTriggers(ctx)
	}
}

func (k *Kernel) EnqueueRegisteredTriggers(ctx context.Context) {
	for _, desc := range k.triggers.ListTriggers() {
		handler, ok := k.triggers.Trigger(desc.TypeKey)
		if !ok || handler == nil {
			continue
		}
		if err := handler.Sync(ctx); err != nil {
			slog.Warn("trigger sync failed", slog.String("trigger_type_key", desc.TypeKey), slog.Any("err", err))
			continue
		}
		if direct, ok := handler.(DirectDueTriggerHandler); ok {
			count, err := direct.EnqueueDue(ctx, 10)
			if err != nil {
				slog.Warn("trigger enqueue failed", slog.String("trigger_type_key", desc.TypeKey), slog.Any("err", err))
				continue
			}
			if count > 0 {
				slog.Info("trigger events enqueued",
					slog.String("trigger_type_key", desc.TypeKey),
					slog.Int("count", count),
				)
			}
			continue
		}
		events, err := handler.ClaimDue(ctx, 10)
		if err != nil {
			slog.Warn("trigger claim failed", slog.String("trigger_type_key", desc.TypeKey), slog.Any("err", err))
			continue
		}
		for _, ev := range events {
			if !k.EnqueueTriggerEvent(ctx, handler, ev) {
				break
			}
		}
	}
}

func (k *Kernel) EnqueueTriggerEvent(ctx context.Context, handler contracts.TriggerHandler, ev contracts.TriggerEvent) bool {
	if k == nil || k.pool == nil || handler == nil {
		return false
	}
	projectID := strings.TrimSpace(ev.ProjectID)
	workflowID := strings.TrimSpace(ev.WorkflowID)
	versionID := strings.TrimSpace(ev.WorkflowVersionID)
	if projectID == "" || workflowID == "" || versionID == "" {
		slog.Warn("trigger event missing workflow identity",
			slog.String("trigger_type_key", handler.TypeKey()),
			slog.String("project_id", projectID),
			slog.String("workflow_id", workflowID),
			slog.String("workflow_version_id", versionID),
		)
		return false
	}
	sourceType := strings.TrimSpace(ev.SourceType)
	if sourceType == "" {
		sourceType = strings.TrimSuffix(strings.ReplaceAll(handler.TypeKey(), ".", "_"), "_trigger")
	}
	input := ev.InputJSON
	if len(strings.TrimSpace(string(input))) == 0 {
		input = json.RawMessage(`{}`)
	}
	runID := uuid.New()
	tx, err := k.pool.Begin(ctx)
	if err != nil {
		return false
	}
	defer tx.Rollback(ctx)
	message := "trigger event queued"
	if ev.IdempotencyKey != "" {
		message = "trigger event queued: " + ev.IdempotencyKey
	}
	if _, _, err := runqueue.RequestRunTx(ctx, tx, runqueue.Request{
		WorkflowRunID:     runID,
		ProjectID:         projectID,
		WorkflowID:        workflowID,
		WorkflowVersionID: versionID,
		SourceType:        sourceType,
		InputJSON:         input,
		EventActionKey:    ev.ActionKey,
		EventMessage:      message,
	}); err != nil {
		slog.Warn("trigger workflow run insert failed",
			slog.String("trigger_type_key", handler.TypeKey()),
			slog.String("workflow_id", workflowID),
			slog.Any("err", err),
		)
		return false
	}
	if err := tx.Commit(ctx); err != nil {
		return false
	}
	if err := handler.MarkEnqueued(ctx, ev, runID.String()); err != nil {
		slog.Warn("trigger mark enqueued failed",
			slog.String("trigger_type_key", handler.TypeKey()),
			slog.String("workflow_run_id", runID.String()),
			slog.Any("err", err),
		)
		return false
	}
	slog.Info("trigger workflow run queued",
		slog.String("trigger_type_key", handler.TypeKey()),
		slog.String("workflow_run_id", runID.String()),
		slog.String("workflow_id", workflowID),
		slog.String("project_id", projectID),
	)
	k.Notify()
	return true
}

func (k *Kernel) Drain(ctx context.Context) {
	if k == nil {
		return
	}
	for k.ProcessOne(ctx) {
	}
}

func (k *Kernel) ProcessOne(ctx context.Context) bool {
	if k == nil || k.pool == nil || k.runtime == nil {
		return false
	}
	if k.timeoutWaitingRun(ctx) {
		return true
	}
	claim, ok := k.claimNextRun(ctx)
	if !ok {
		return false
	}
	if allowed, err := demomode.RunAllowed(ctx, k.pool, claim.RunID, claim.ProjectID); err != nil || !allowed {
		const message = "Запуск недоступен: демо выполняет только свой почтовый сценарий"
		if claim.markFailed(ctx, k.pool, message) {
			claim.insertRunEvent(ctx, k.pool, "", "workflow.failed", message, `{}`)
		}
		return true
	}
	rid := claim.RunID
	wvid := claim.WorkflowVersionID
	def := claim.DefinitionJSON
	runInput := claim.InputJSON
	runProjectID := claim.ProjectID
	if k.runtime.DefinitionNeedsProjectID(def) {
		runInput = k.runtime.EnrichRunInputProjectID(runInput, runProjectID)
	}
	runPlan, filtErr := k.runtime.BuildRunPlan(ctx, def, runInput, runProjectID)
	if filtErr != nil {
		failMsg := filtErr.Error()
		if claim.markFailed(ctx, k.pool, failMsg) {
			claim.insertRunEvent(ctx, k.pool, "", "workflow.failed", failMsg, `{}`)
		}
		slog.Error("workflow run failed after claim",
			slog.String("workflow_run_id", rid),
			slog.String("workflow_version_id", wvid),
			slog.String("project_id", runProjectID),
			slog.String("error", failMsg),
		)
		return true
	}
	if claim.ClaimCount > 1 {
		claim.insertRunEvent(ctx, k.pool, "", "workflow.reclaimed", "executor resumed existing run", `{}`)
	} else {
		claim.insertRunEvent(ctx, k.pool, "", "workflow.started", "executor started", `{}`)
	}
	slog.Info("workflow run claimed",
		slog.String("workflow_run_id", rid),
		slog.String("workflow_version_id", wvid),
		slog.String("project_id", runProjectID),
		slog.String("claim_owner", claim.Owner),
		slog.Int64("fencing_token", claim.FencingToken),
		slog.Int("claim_count", claim.ClaimCount),
		slog.Duration("lease", claim.LeaseDuration),
		slog.Int("actions", len(runPlan.Execute)),
		slog.Int("skipped_actions", len(runPlan.Skipped)),
	)

	db := k.pool
	var err error
	completedKeys := make([]string, 0, len(runPlan.Execute))
	completedMap := map[string]bool{}
	branchOutputs := map[string]string{}
	continuedFailures := make([]string, 0)
	requiredFailures := make([]string, 0)
	stepExpose := runPlan.InitialSteps
	for _, node := range runPlan.Execute {
		if !claim.heartbeat(ctx, db) {
			slog.Warn("workflow run fencing lost before action", slog.String("workflow_run_id", rid), slog.String("action_key", node.ActionKey))
			return false
		}
		stored, done, restoreErr := k.completedOutput(ctx, claim, node)
		if restoreErr != nil {
			slog.Error("restore completed action", slog.String("run", rid), slog.Any("err", restoreErr))
			return false // Keep the lease recoverable; never replay on an uncertain read.
		}
		if done {
			completedKeys = append(completedKeys, node.ActionKey)
			completedMap[node.ActionKey] = true
			if branch := k.runtime.BranchFromOutput(stored); branch != "" {
				branchOutputs[node.ActionKey] = branch
			}
			k.runtime.MergeActionExpose(stepExpose, node.ActionKey, stored)
			continue
		}
		inputJSON := k.runtime.ActionInput(node, runInput)
		maxAttempts := node.MaxAttempts
		if maxAttempts < 1 {
			maxAttempts = 1
		}
		if runPlan.UseGraphEdges && !k.runtime.ShouldRunGraphNode(node, runPlan.Graph, completedMap, branchOutputs) {
			output := k.runtime.SkippedOutputWithReason(node, "branch_not_taken")
			_, _ = db.Exec(ctx, `
INSERT INTO action_runs (
  workflow_run_id, action_key, action_type_key, action_dependency_id, status,
  attempt_count, max_attempts, input_json, output_json, provider_response_json, started_at, finished_at)
VALUES ($1,$2,$3,$4,'skipped',0,$5,$6::jsonb,$7::jsonb,'{}'::jsonb, now(), now())
ON CONFLICT (workflow_run_id, action_key) DO UPDATE SET
  status = 'skipped',
  attempt_count = 0,
  max_attempts = EXCLUDED.max_attempts,
  input_json = EXCLUDED.input_json,
  output_json = EXCLUDED.output_json,
	  provider_response_json = '{}'::jsonb,
	  started_at = now(),
	  finished_at = now(),
	  next_retry_at = NULL,
	  retry_class = NULL,
	  retry_policy_json = '{}'::jsonb,
	  retry_started_at = NULL,
	  last_retry_at = NULL,
	  retry_exhausted_at = NULL,
	  dead_lettered_at = NULL,
	  dead_letter_reason = NULL,
	  dead_letter_owner = NULL,
	  replayable = false,
	  error = NULL`,
				rid, node.ActionKey, node.ActionTypeKey, node.ActionDependencyID, maxAttempts, inputJSON, output)
			_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
VALUES ($1,$2,'action.skipped','branch not taken',$3::jsonb, clock_timestamp())`,
				rid, node.ActionKey, output)
			continue
		}
		if node.Disabled {
			if err := k.recordDisabledNode(ctx, &claim, node, inputJSON); err != nil {
				claim.markFailed(ctx, db, err.Error())
				return false
			}
			completedMap[node.ActionKey] = true
			continue
		}

		var actionRunID string
		err = db.QueryRow(ctx, `
INSERT INTO action_runs (
  workflow_run_id, action_key, action_type_key, action_dependency_id, status,
  attempt_count, max_attempts, input_json, started_at)
VALUES ($1,$2,$3,$4,'running',0,$5,$6::jsonb, now())
ON CONFLICT (workflow_run_id, action_key) DO UPDATE SET
  action_type_key = EXCLUDED.action_type_key,
  action_dependency_id = EXCLUDED.action_dependency_id,
  status = 'running',
  attempt_count = 0,
  max_attempts = EXCLUDED.max_attempts,
  input_json = EXCLUDED.input_json,
  output_json = NULL,
	  provider_response_json = NULL,
	  started_at = now(),
	  finished_at = NULL,
	  next_retry_at = NULL,
	  retry_class = NULL,
	  retry_policy_json = '{}'::jsonb,
	  retry_started_at = NULL,
	  last_retry_at = NULL,
	  retry_exhausted_at = NULL,
	  dead_lettered_at = NULL,
	  dead_letter_reason = NULL,
	  dead_letter_owner = NULL,
	  replayable = false,
	  error = NULL
	RETURNING action_run_id::text`,
			rid, node.ActionKey, node.ActionTypeKey, node.ActionDependencyID, maxAttempts, inputJSON).
			Scan(&actionRunID)
		if err != nil {
			slog.Error("action insert", slog.String("run", rid), slog.Any("err", err))
			claim.markFailed(ctx, db, err.Error())
			return false
		}
		_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, action_key, event_type, message, created_at)
VALUES ($1,$2,'action.started',$3, clock_timestamp())`, rid, node.ActionKey, node.ActionTypeKey)

		var final ActionExecution
		var succeeded bool
		attemptsMade := 0
		actionStartedAt := time.Now().UTC()
		var lastFailureEval runtimeretry.Evaluation
		haveFailureEval := false
		for attemptNo := 1; attemptNo <= maxAttempts; attemptNo++ {
			attemptsMade = attemptNo
			effect, ledgerFinal, shortCircuit := k.beginEffect(ctx, claim, actionRunID, node, attemptNo, inputJSON, runInput)
			if shortCircuit {
				final = ledgerFinal
			} else {
				actionCtx, stopHeartbeat := claim.startHeartbeat(ctx, db)
				final = k.runtime.ExecuteAction(actionCtx, node, runInput, stepExpose, rid, runProjectID, attemptNo)
				stopHeartbeat()
				final = k.finishEffect(ctx, effect, final)
			}
			if !claim.heartbeat(ctx, db) {
				slog.Warn("workflow run fencing lost after action",
					slog.String("workflow_run_id", rid),
					slog.String("action_key", node.ActionKey),
					slog.Int("attempt", attemptNo),
				)
				return false
			}
			if final.Waiting {
				waitingReason := strings.TrimSpace(final.WaitingReason)
				if waitingReason == "" {
					waitingReason = "manual_approval_required"
				}
				var waitingDeadline *time.Time
				if !final.WaitingDeadline.IsZero() {
					deadline := final.WaitingDeadline.UTC()
					waitingDeadline = &deadline
				}
				_, err = db.Exec(ctx, `
UPDATE action_runs
   SET status = 'waiting',
       attempt_count = 0,
       output_json = $2::jsonb,
	       provider_response_json = $3::jsonb,
	       finished_at = NULL,
	       waiting_reason = $4,
	       waiting_deadline_at = $5,
	       wake_reason = '',
	       last_state_transition_at = now(),
	       next_retry_at = NULL,
	       retry_class = NULL,
	       retry_policy_json = '{}'::jsonb,
	       retry_exhausted_at = NULL,
	       dead_lettered_at = NULL,
	       dead_letter_reason = NULL,
	       dead_letter_owner = NULL,
	       replayable = false,
	       error = NULL
	 WHERE action_run_id = $1
	   AND EXISTS (
	     SELECT 1 FROM workflow_runs wr
	      WHERE wr.workflow_run_id = action_runs.workflow_run_id
	        AND wr.run_claim_owner = $6
	        AND wr.run_fencing_token = $7
	        AND wr.status = 'running'
	   )`, actionRunID, final.OutputJSON, final.ProviderResponseJSON, waitingReason, waitingDeadline, claim.Owner, claim.FencingToken)
				if err != nil {
					slog.Error("action waiting update", slog.String("run", rid), slog.Any("err", err))
					claim.markFailed(ctx, db, err.Error())
					return false
				}
				if !claim.markWaiting(ctx, db, waitingReason, waitingDeadline) {
					slog.Warn("workflow run fencing lost before waiting", slog.String("workflow_run_id", rid))
					return false
				}
				_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
VALUES ($1,$2,'action.waiting',$3,$4::jsonb, clock_timestamp())`,
					rid, node.ActionKey, waitingReason, final.OutputJSON)
				_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, event_type, message, payload_json, created_at)
VALUES ($1,'workflow.waiting',$2,$3::jsonb, clock_timestamp())`,
					rid, waitingReason, final.OutputJSON)
				slog.Info("workflow run waiting",
					slog.String("workflow_run_id", rid),
					slog.String("action_key", node.ActionKey),
					slog.String("action_type_key", node.ActionTypeKey),
				)
				return true
			}
			attemptStatus := "succeeded"
			errorText := any(nil)
			errorType := any(nil)
			if final.Err != nil {
				attemptStatus = "failed"
				errorText = runtimeretry.Redact(final.Err.Error())
				errorType = final.ErrorType
			}
			_, err = db.Exec(ctx, upsertActionAttemptSQL,
				actionRunID, attemptNo, attemptStatus, final.InputJSON, final.OutputJSON, final.ProviderResponseJSON, errorText, errorType)
			if err != nil {
				slog.Error("attempt insert", slog.String("run", rid), slog.Any("err", err))
				claim.markFailed(ctx, db, err.Error())
				return false
			}
			if final.Err == nil {
				succeeded = true
				_, err = db.Exec(ctx, `
UPDATE action_runs
   SET status = 'succeeded',
       attempt_count = $2,
       output_json = $3::jsonb,
	       provider_response_json = $4::jsonb,
	       finished_at = now(),
	       next_retry_at = NULL,
	       retry_class = NULL,
	       retry_policy_json = '{}'::jsonb,
	       retry_exhausted_at = NULL,
	       dead_lettered_at = NULL,
	       dead_letter_reason = NULL,
	       dead_letter_owner = NULL,
	       replayable = false,
	       error = NULL
	 WHERE action_run_id = $1`, actionRunID, attemptNo, final.OutputJSON, final.ProviderResponseJSON)
				if err != nil {
					slog.Error("action update", slog.String("run", rid), slog.Any("err", err))
					claim.markFailed(ctx, db, err.Error())
					return false
				}
				_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
VALUES ($1,$2,'action.succeeded',$3,$4::jsonb, clock_timestamp())`,
					rid, node.ActionKey, node.ActionTypeKey, final.OutputJSON)
				break
			}

			failureEval := evaluateActionFailure(node, attemptNo, maxAttempts, actionStartedAt, claim, final)
			lastFailureEval = failureEval
			haveFailureEval = true
			if failureEval.ShouldRetry {
				nextRetryAt := time.Now().UTC().Add(failureEval.NextDelay)
				policyJSON := retryPolicyJSON(failureEval)
				_, err = db.Exec(ctx, `
	UPDATE action_runs
	   SET status = 'retrying',
	       attempt_count = $2,
	       provider_response_json = $3::jsonb,
	       error = $4,
	       next_retry_at = $5,
	       retry_class = $6,
	       retry_policy_json = $7::jsonb,
	       retry_started_at = COALESCE(retry_started_at, clock_timestamp()),
	       last_retry_at = clock_timestamp(),
	       retry_exhausted_at = NULL,
	       dead_lettered_at = NULL,
	       dead_letter_reason = NULL,
	       dead_letter_owner = NULL,
	       replayable = false
	 WHERE action_run_id = $1`, actionRunID, attemptNo, final.ProviderResponseJSON, failureEval.RedactedReason,
					nextRetryAt, string(failureEval.Class), policyJSON)
				if err != nil {
					slog.Error("action retry update", slog.String("run", rid), slog.Any("err", err))
					claim.markFailed(ctx, db, err.Error())
					return false
				}
				_, _ = db.Exec(ctx, `
	INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
	VALUES ($1,$2,'action.retry_scheduled',$3,$4::jsonb, clock_timestamp())`,
					rid, node.ActionKey, failureEval.RedactedReason, retryEventPayload(failureEval, final.ProviderResponseJSON))
				slog.Warn("workflow action retry scheduled",
					slog.String("workflow_run_id", rid),
					slog.String("action_key", node.ActionKey),
					slog.String("action_type_key", node.ActionTypeKey),
					slog.String("retry_class", string(failureEval.Class)),
					slog.Int("attempt", attemptNo),
					slog.Int("max_attempts", maxAttempts),
					slog.Duration("next_delay", failureEval.NextDelay),
				)
				if !claim.waitForRetry(ctx, db, failureEval.NextDelay) {
					slog.Warn("workflow run fencing lost before retry",
						slog.String("workflow_run_id", rid),
						slog.String("action_key", node.ActionKey),
						slog.Int("attempt", attemptNo),
					)
					return false
				}
				continue
			}
			break
		}

		if !succeeded {
			errText := "action failed"
			if final.Err != nil {
				errText = final.Err.Error()
			}
			if !haveFailureEval {
				lastFailureEval = evaluateActionFailure(node, attemptsMade, maxAttempts, actionStartedAt, claim, final)
				haveFailureEval = true
			}
			if lastFailureEval.RedactedReason != "" {
				errText = lastFailureEval.RedactedReason
			}
			policyJSON := retryPolicyJSON(lastFailureEval)
			retryExhausted := lastFailureEval.Exhausted && lastFailureEval.Retryable
			_, err = db.Exec(ctx, `
	UPDATE action_runs
	   SET status = 'dead_lettered',
	       attempt_count = $2,
	       output_json = $3::jsonb,
	       provider_response_json = $4::jsonb,
	       finished_at = now(),
	       error = $5,
	       next_retry_at = NULL,
	       retry_class = $6,
	       retry_policy_json = $7::jsonb,
	       retry_exhausted_at = CASE WHEN $8 THEN clock_timestamp() ELSE retry_exhausted_at END,
	       dead_lettered_at = clock_timestamp(),
	       dead_letter_reason = $9,
	       dead_letter_owner = $10,
	       replayable = true
	 WHERE action_run_id = $1`, actionRunID, attemptsMade, final.OutputJSON, final.ProviderResponseJSON, errText,
				string(lastFailureEval.Class), policyJSON, retryExhausted, lastFailureEval.RedactedReason, lastFailureEval.Owner)
			if err != nil {
				slog.Error("action dead-letter update", slog.String("run", rid), slog.Any("err", err))
				claim.markFailed(ctx, db, err.Error())
				return false
			}
			recordDeadLetter(ctx, db, claim, actionRunID, node, lastFailureEval)
			_, _ = db.Exec(ctx, `
	INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
	VALUES ($1,$2,'action.failed',$3,$4::jsonb, clock_timestamp())`,
				rid, node.ActionKey, errText, final.ProviderResponseJSON)
			_, _ = db.Exec(ctx, `
	INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
	VALUES ($1,$2,'action.dead_lettered',$3,$4::jsonb, clock_timestamp())`,
				rid, node.ActionKey, lastFailureEval.RedactedReason, retryEventPayload(lastFailureEval, final.ProviderResponseJSON))
			if retryExhausted {
				_, _ = db.Exec(ctx, `
	INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
	VALUES ($1,$2,'action.retry_exhausted',$3,$4::jsonb, clock_timestamp())`,
					rid, node.ActionKey, lastFailureEval.RedactedReason, retryEventPayload(lastFailureEval, final.ProviderResponseJSON))
			}
			if detector, ok := k.runtime.(IncidentDetector); ok {
				if incident, ok := detector.IncidentFromActionExecution(final); ok {
					_, _ = db.Exec(ctx, `
	INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
VALUES ($1,$2,'incident.detected',$3,$4::jsonb, clock_timestamp())`,
						rid, node.ActionKey, incident.Message, incident.PayloadJSON)
				}
			}
			if node.ContinueOnFailure {
				continuedFailures = append(continuedFailures, node.ActionKey)
				if node.FailureRequired {
					requiredFailures = append(requiredFailures, node.ActionKey)
				}
				completedMap[node.ActionKey] = true
				continue
			}
			if !claim.markFailed(ctx, db, errText) {
				slog.Warn("workflow run fencing lost before failure", slog.String("workflow_run_id", rid))
				return false
			}
			_, _ = db.Exec(ctx, `
	INSERT INTO run_events (workflow_run_id, event_type, message, created_at)
	VALUES ($1,'workflow.failed',$2, clock_timestamp())`, rid, errText)
			slog.Error("workflow run failed",
				slog.String("workflow_run_id", rid),
				slog.String("action_key", node.ActionKey),
				slog.String("action_type_key", node.ActionTypeKey),
				slog.String("retry_class", string(lastFailureEval.Class)),
				slog.String("dead_letter_owner", lastFailureEval.Owner),
				slog.String("error", errText),
			)
			return true
		}
		completedKeys = append(completedKeys, node.ActionKey)
		completedMap[node.ActionKey] = true
		if branch := k.runtime.BranchFromOutput(final.OutputJSON); branch != "" {
			branchOutputs[node.ActionKey] = branch
		}
		k.runtime.MergeActionExpose(stepExpose, node.ActionKey, final.OutputJSON)
	}
	for _, node := range runPlan.Skipped {
		if !claim.heartbeat(ctx, db) {
			slog.Warn("workflow run fencing lost before skipped action", slog.String("workflow_run_id", rid), slog.String("action_key", node.ActionKey))
			return false
		}
		maxAttempts := node.MaxAttempts
		if maxAttempts < 1 {
			maxAttempts = 1
		}
		output := k.runtime.SkippedOutput(node)
		_, _ = db.Exec(ctx, `
INSERT INTO action_runs (
  workflow_run_id, action_key, action_type_key, action_dependency_id, status,
  attempt_count, max_attempts, input_json, output_json, provider_response_json, started_at, finished_at)
VALUES ($1,$2,$3,$4,'skipped',0,$5,$6::jsonb,$7::jsonb,'{}'::jsonb, now(), now())
ON CONFLICT (workflow_run_id, action_key) DO UPDATE SET
  status = 'skipped',
  attempt_count = 0,
  max_attempts = EXCLUDED.max_attempts,
  input_json = EXCLUDED.input_json,
  output_json = EXCLUDED.output_json,
	  provider_response_json = '{}'::jsonb,
	  started_at = now(),
	  finished_at = now(),
	  next_retry_at = NULL,
	  retry_class = NULL,
	  retry_policy_json = '{}'::jsonb,
	  retry_started_at = NULL,
	  last_retry_at = NULL,
	  retry_exhausted_at = NULL,
	  dead_lettered_at = NULL,
	  dead_letter_reason = NULL,
	  dead_letter_owner = NULL,
	  replayable = false,
	  error = NULL`,
			rid, node.ActionKey, node.ActionTypeKey, node.ActionDependencyID, maxAttempts, k.runtime.ActionInput(node, runInput), output)
		_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, action_key, event_type, message, payload_json, created_at)
VALUES ($1,$2,'action.skipped','test run skipped downstream action',$3::jsonb, clock_timestamp())`,
			rid, node.ActionKey, output)
	}
	out := workflowCompletionOutput(completedKeys, continuedFailures, requiredFailures)
	ob, _ := json.Marshal(out)
	if len(requiredFailures) > 0 {
		errorText := fmt.Sprintf("%d required action(s) failed", len(requiredFailures))
		if !claim.markFailedWithOutput(ctx, db, errorText, string(ob), "required_action_failed") {
			slog.Warn("workflow run fencing lost before required failure", slog.String("workflow_run_id", rid))
			return false
		}
		_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, event_type, message, payload_json, created_at)
VALUES ($1,'workflow.failed',$2,$3::jsonb, clock_timestamp())`, rid, errorText, string(ob))
		slog.Error("workflow run failed after independent required actions",
			slog.String("workflow_run_id", rid),
			slog.Int("required_failures", len(requiredFailures)),
		)
		return true
	}
	if len(continuedFailures) > 0 {
		warnText := fmt.Sprintf("%d optional action(s) failed", len(continuedFailures))
		if !claim.markSucceeded(ctx, db, string(ob), true) {
			slog.Warn("workflow run fencing lost before partial success", slog.String("workflow_run_id", rid))
			return false
		}
		_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, event_type, message, payload_json, created_at)
VALUES ($1,'workflow.warning',$2,$3::jsonb, clock_timestamp())`, rid, warnText, string(ob))
		_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, event_type, message, payload_json, created_at)
VALUES ($1,'workflow.succeeded','executor completed critical actions with optional failures',$2::jsonb, clock_timestamp())`, rid, string(ob))
		slog.Info("workflow run finished with continued failures",
			slog.String("workflow_run_id", rid),
			slog.Int("continued_failures", len(continuedFailures)),
		)
		return true
	}
	if !claim.markSucceeded(ctx, db, string(ob), false) {
		slog.Warn("workflow run fencing lost before success", slog.String("workflow_run_id", rid))
		return false
	}
	_, _ = db.Exec(ctx, `
INSERT INTO run_events (workflow_run_id, event_type, message, payload_json, created_at)
VALUES ($1,'workflow.succeeded','executor completed actions',$2::jsonb, clock_timestamp())`, rid, string(ob))
	slog.Info("workflow run finished",
		slog.String("workflow_run_id", rid),
		slog.String("project_id", runProjectID),
		slog.Int("completed_actions", len(completedKeys)),
	)
	return true
}

func workflowCompletionOutput(completed, continuedFailures, requiredFailures []string) map[string]any {
	out := map[string]any{"completed_actions": completed}
	if len(continuedFailures) == 0 {
		return out
	}
	out["continued_failures"] = continuedFailures
	if len(requiredFailures) > 0 {
		out["required_failures"] = requiredFailures
		out["status_detail"] = "failed_required_actions"
		out["error"] = fmt.Sprintf("%d required action(s) failed", len(requiredFailures))
		return out
	}
	out["status_detail"] = "partial_success"
	out["warning"] = fmt.Sprintf("%d optional action(s) failed", len(continuedFailures))
	return out
}
