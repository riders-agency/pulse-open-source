package executor

import (
	"context"
	"time"

	runtimekernel "riders.agency/pulse/api/internal/runtime/kernel"
)

func (e *Engine) SetNotifier(notify func()) {
	if e == nil {
		return
	}
	e.notify = notify
}

func (e *Engine) BuildRunPlan(_ context.Context, def []byte, runInput []byte, _ string) (runtimekernel.RunPlan, error) {
	nodes := flattenNodes(def)
	edges := flattenEdges(def)
	runPlan, err := planRunNodes(nodes, runInput)
	if err != nil {
		return runtimekernel.RunPlan{}, err
	}
	return runtimekernel.RunPlan{
		Execute:       kernelNodes(runPlan.Execute),
		Skipped:       kernelNodes(runPlan.Skipped),
		InitialSteps:  runPlan.InitialSteps,
		UseGraphEdges: len(edges) > 0 && len(runPlan.Skipped) == 0 && !runPlan.IgnoreGraphEdges,
		Graph:         edges,
	}, nil
}

func (e *Engine) DefinitionNeedsProjectID(def []byte) bool {
	return DefinitionHasPaymentDueTrigger(def)
}

func (e *Engine) EnrichRunInputProjectID(runInput []byte, projectID string) []byte {
	return EnrichRunInputProjectID(runInput, projectID)
}

func (e *Engine) ActionInput(node runtimekernel.Node, runInput []byte) string {
	return actionInput(actionPlanFromKernelNode(node), runInput)
}

func (e *Engine) SkippedOutput(node runtimekernel.Node) string {
	return skippedOutput(actionPlanFromKernelNode(node))
}

func (e *Engine) SkippedOutputWithReason(node runtimekernel.Node, reason string) string {
	return skippedOutputWithReason(actionPlanFromKernelNode(node), reason)
}

func (e *Engine) ShouldRunGraphNode(node runtimekernel.Node, graph any, completed map[string]bool, branchOutputs map[string]string) bool {
	edges, _ := graph.([]edgePlan)
	return shouldRunGraphNode(actionPlanFromKernelNode(node), edges, completed, branchOutputs)
}

func (e *Engine) BranchFromOutput(outputJSON string) string {
	return branchFromOutput(outputJSON)
}

func (e *Engine) MergeActionExpose(steps map[string]any, actionKey string, outputJSON string) {
	mergeActionExpose(steps, actionKey, outputJSON)
}

func (e *Engine) ExecuteAction(ctx context.Context, node runtimekernel.Node, runInput []byte, steps map[string]any, runID string, projectID string, attempt int) runtimekernel.ActionExecution {
	final := e.executeAction(ctx, actionPlanFromKernelNode(node), runInput, steps, runID, projectID, attempt)
	return runtimekernel.ActionExecution{
		InputJSON:            final.InputJSON,
		OutputJSON:           final.OutputJSON,
		ProviderResponseJSON: final.ProviderResponseJSON,
		Waiting:              final.Waiting,
		WaitingReason:        final.WaitingReason,
		WaitingDeadline:      final.WaitingDeadline,
		Retryable:            final.Retryable,
		ErrorType:            final.ErrorType,
		Err:                  final.Err,
	}
}

func (e *Engine) IncidentFromActionExecution(final runtimekernel.ActionExecution) (runtimekernel.Incident, bool) {
	incident, ok := incidentFromExecution(actionExecution{
		InputJSON:            final.InputJSON,
		OutputJSON:           final.OutputJSON,
		ProviderResponseJSON: final.ProviderResponseJSON,
		Waiting:              final.Waiting,
		Retryable:            final.Retryable,
		ErrorType:            final.ErrorType,
		Err:                  final.Err,
	})
	if !ok {
		return runtimekernel.Incident{}, false
	}
	return runtimekernel.Incident{
		Message:     incident.Message,
		PayloadJSON: incident.PayloadJSON,
	}, true
}

func kernelNodes(nodes []actionPlan) []runtimekernel.Node {
	out := make([]runtimekernel.Node, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, kernelNode(node))
	}
	return out
}

func kernelNode(node actionPlan) runtimekernel.Node {
	dependencyID := nullableUUID(node.Config.ActionDependencyID)
	if node.ActionTypeKey == "email.send" {
		dependencyID = nullableUUID(node.Config.SMTPActionDependencyID)
	}
	return runtimekernel.Node{
		Disabled:           node.Config.Disabled,
		ActionKey:          node.ActionKey,
		ActionTypeKey:      node.ActionTypeKey,
		ActionDependencyID: dependencyID,
		MaxAttempts:        clampAttempts(node.Config.RetryPolicy.MaxAttempts),
		RetryInitialDelay:  parsePositiveDuration(node.Config.RetryPolicy.InitialInterval),
		RetryMaxDelay:      parsePositiveDuration(node.Config.RetryPolicy.MaxInterval),
		RetryMaxElapsed:    parsePositiveDuration(node.Config.RetryPolicy.MaxElapsed),
		ContinueOnFailure:  node.Config.ContinueOnFailure,
		FailureRequired:    node.Config.FailureRequired,
		Payload:            node,
	}
}

func parsePositiveDuration(raw string) time.Duration {
	duration, err := time.ParseDuration(raw)
	if err != nil || duration <= 0 {
		return 0
	}
	return duration
}

func actionPlanFromKernelNode(node runtimekernel.Node) actionPlan {
	if plan, ok := node.Payload.(actionPlan); ok {
		return plan
	}
	return actionPlan{
		ActionKey:     node.ActionKey,
		ActionTypeKey: node.ActionTypeKey,
	}
}
