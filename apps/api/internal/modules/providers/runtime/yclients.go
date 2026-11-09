package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	providerdiagnostics "riders.agency/pulse/api/internal/modules/providers/diagnostics"
	yclientsprovider "riders.agency/pulse/api/internal/modules/providers/yclients"
)

func (r *Runner) executeYClientsCertificate(ctx context.Context, node actionPlan, runInput []byte, steps map[string]any, runID string) actionExecution {
	input := actionInput(node, runInput)
	dep, err := r.loadDependency(ctx, node.Config.ActionDependencyID)
	if err != nil {
		return failedAction(input, "{}", yclientsprovider.FailureResponse(err), false, "dependency", err)
	}
	if dep.TypeKey != "yclients.account" {
		err = errors.New("yclients.certificate.issue requires yclients.account dependency")
		return failedAction(input, "{}", yclientsprovider.FailureResponse(err), false, "dependency_type", err)
	}
	if dep.Status != "active" {
		err = fmt.Errorf("YClients dependency is %s", dep.Status)
		return failedAction(input, "{}", yclientsprovider.FailureResponse(err), false, "dependency_status", err)
	}
	credentials, err := parseYClientsCredentials(dep.Plaintext)
	if err != nil {
		return failedAction(input, "{}", yclientsprovider.FailureResponse(err), false, "dependency", err)
	}
	command, requestJSON, err := buildYClientsCommand(node.Config, runInput, steps, runID)
	if err != nil {
		return failedAction(input, "{}", yclientsprovider.FailureResponse(err), false, "validation", err)
	}
	operation, _, err := r.reserveYClientsOperation(ctx, node.Config.ActionDependencyID, command)
	if err != nil {
		return failedAction(string(requestJSON), "{}", yclientsprovider.FailureResponse(err), false, "idempotency", err)
	}
	if operation.InProgress {
		err = errors.New("YClients operation is already in progress; no second write was sent")
		return failedAction(string(requestJSON), "{}", yclientsprovider.FailureResponse(err), true, "operation_in_progress", err)
	}
	if operation.Replay {
		return r.replayYClientsOperation(ctx, node, dep, credentials, command, operation, requestJSON, runID)
	}
	timeout := durationOrDefault(node.Config.RetryPolicy.Timeout, 15*time.Second)
	clientConfig := yclientsClientConfig(dep, credentials, timeout)
	clientConfig.HTTPClient = r.httpClient()
	client, err := yclientsprovider.NewCertificateClient(clientConfig)
	if err != nil {
		_ = r.failYClientsOperation(ctx, operation.ID, err)
		return r.finishYClientsFailure(ctx, node, dep, credentials, command, operation.ID, requestJSON, runID, "validation", err)
	}
	result, issueErr := client.IssueCertificate(ctx, yclientsprovider.CertificateIssueInput{
		OperationKey: command.OperationKey, ProductKind: command.ProductKind,
		ProductTitle: command.ProductTitle, AmountRub: command.AmountRub, ServiceIDs: command.ServiceIDs,
	})
	if issueErr != nil {
		if persistErr := r.failYClientsOperation(ctx, operation.ID, issueErr); persistErr != nil {
			issueErr = errors.Join(issueErr, persistErr)
		}
		if yclientsprovider.CertificateRetryable(issueErr) && !yclientsprovider.CertificatePossibleEffect(issueErr) {
			return failedAction(string(requestJSON), "{}", yclientsprovider.FailureResponse(issueErr), true, "provider_http", issueErr)
		}
		return r.finishYClientsFailure(ctx, node, dep, credentials, command, operation.ID, requestJSON, runID, "provider_http", issueErr)
	}
	if err := r.finishYClientsOperation(ctx, operation.ID, result); err != nil {
		return failedAction(string(requestJSON), "{}", yclientsprovider.FailureResponse(err), false, "effect_ledger", err)
	}
	return r.finishYClientsSuccess(ctx, node, dep, credentials, command, operation.ID, requestJSON, runID, client.AuthSource(), result)
}

func buildYClientsCommand(config actionConfig, runInput []byte, steps map[string]any, runID string) (yclientsCommand, []byte, error) {
	tpl := templateSystemForRun(runID)
	render := func(value string) string { return strings.TrimSpace(renderTemplate(value, runInput, steps, tpl)) }
	amount, err := strconv.Atoi(render(config.AmountRub))
	if err != nil || amount <= 0 {
		return yclientsCommand{}, nil, errors.New("amount_rub must be a positive integer")
	}
	serviceIDs, err := parseServiceIDs(render(config.ServiceIDsJSON))
	if err != nil {
		return yclientsCommand{}, nil, err
	}
	command := yclientsCommand{
		OperationKey: render(config.OperationKey), OrderID: render(config.OrderID),
		OrderItemID: render(config.OrderItemID), InstanceID: render(config.InstanceID),
		CertificateNumber: render(config.CertificateNumber), ProductKind: strings.ToLower(render(config.ProductKind)),
		ProductTitle: render(config.ProductTitle), AmountRub: amount, ServiceIDs: serviceIDs,
	}
	if command.OperationKey == "" || command.OrderID == "" || command.OrderItemID == "" || command.InstanceID == "" || command.CertificateNumber == "" {
		return command, nil, errors.New("YClients command identity is incomplete")
	}
	if command.ProductKind != "nominal" && command.ProductKind != "service" {
		return command, nil, errors.New("product_kind must be nominal or service")
	}
	if command.ProductKind == "service" && len(command.ServiceIDs) == 0 {
		return command, nil, errors.New("service certificate requires service IDs")
	}
	raw, err := json.Marshal(command)
	return command, raw, err
}

func (r *Runner) replayYClientsOperation(ctx context.Context, node actionPlan, dep *dependencyRecord, credentials yclientsCredentials, command yclientsCommand, operation yclientsOperation, requestJSON []byte, runID string) actionExecution {
	if operation.Status == "succeeded" {
		result, err := decodeYClientsResult(operation.ResultJSON)
		if err != nil {
			return failedAction(string(requestJSON), "{}", yclientsprovider.FailureResponse(err), false, "effect_ledger", err)
		}
		return r.finishYClientsSuccess(ctx, node, dep, credentials, command, operation.ID, requestJSON, runID, "persisted", result)
	}
	return r.finishYClientsFailure(ctx, node, dep, credentials, command, operation.ID, requestJSON, runID, "provider_http", yclientsReplayError(operation))
}

func (r *Runner) finishYClientsSuccess(ctx context.Context, node actionPlan, dep *dependencyRecord, credentials yclientsCredentials, command yclientsCommand, operationID string, requestJSON []byte, runID, authSource string, result yclientsprovider.CertificateIssueResult) actionExecution {
	receipt := yclientsReceipt{
		WorkflowRunID: runID, OperationID: operationID, OperationKey: command.OperationKey,
		OrderID: command.OrderID, OrderItemID: command.OrderItemID, InstanceID: command.InstanceID,
		CertificateNumber: command.CertificateNumber, Outcome: "synced", CertificateCode: result.Code,
		CertificateTypeID: result.CertificateTypeID, ProductID: result.ProductID, LocationID: result.LocationID,
		DocumentID: result.DocumentID, TransactionID: result.TransactionID,
	}
	provider := map[string]any{
		"provider": "yclients", "ok": true, "certificate_issued": true,
		"receipt_delivered": false, "certificate_type_id": result.CertificateTypeID,
		"product_id": result.ProductID, "location_id": result.LocationID,
		"document_id": result.DocumentID, "transaction_id": result.TransactionID,
		"auth_source":                authSource,
		"provider_response_redacted": true,
	}
	if err := r.sendYClientsReceipt(ctx, dep, credentials, receipt); err != nil {
		return failedAction(string(requestJSON), "{}", provider, true, "receipt_network", err)
	}
	provider["receipt_delivered"] = true
	return finalizeMessagingOutput(node.ActionTypeKey, node.ActionKey, node.Config.SuccessCriteria, string(requestJSON), provider, map[string]any{
		"delivered": true, "provider": "yclients", "operation_key": command.OperationKey,
		"certificate_code": result.Code, "transaction_id": result.TransactionID,
	}, map[string]any{"operation_key": command.OperationKey, "certificate_code": result.Code, "transaction_id": result.TransactionID})
}

func (r *Runner) finishYClientsFailure(ctx context.Context, node actionPlan, dep *dependencyRecord, credentials yclientsCredentials, command yclientsCommand, operationID string, requestJSON []byte, runID, errorType string, providerErr error) actionExecution {
	provider := yclientsprovider.FailureResponse(providerErr)
	outcome := "failed"
	if yclientsprovider.CertificatePossibleEffect(providerErr) {
		outcome = "unknown"
	}
	receipt := yclientsReceipt{
		WorkflowRunID: runID, OperationID: operationID, OperationKey: command.OperationKey,
		OrderID: command.OrderID, OrderItemID: command.OrderItemID, InstanceID: command.InstanceID,
		CertificateNumber: command.CertificateNumber, Outcome: outcome,
		ErrorCode:      yclientsprovider.CertificateErrorCode(providerErr),
		PossibleEffect: yclientsprovider.CertificatePossibleEffect(providerErr), Retryable: false,
	}
	if err := r.sendYClientsReceipt(ctx, dep, credentials, receipt); err != nil {
		provider["receipt_delivered"] = false
		return failedAction(string(requestJSON), "{}", provider, true, "receipt_network", errors.Join(providerErr, err))
	}
	provider["receipt_delivered"] = true
	report := diagnosticsForFailure(node.ActionTypeKey, provider, false, errorType, providerErr)
	return actionExecution{
		InputJSON: string(requestJSON), OutputJSON: providerdiagnostics.OutputWithReport("{}", report),
		ProviderResponseJSON: mustYClientsJSON(provider), Retryable: false, ErrorType: errorType, Err: providerErr,
	}
}

func mustYClientsJSON(value any) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
