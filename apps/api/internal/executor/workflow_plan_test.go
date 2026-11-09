package executor

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestActionPlanPreservesDocumentAttachmentConfig(t *testing.T) {
	raw := []byte(`{
		"action_key":"notify",
		"action_type_key":"max.send_message",
		"config":{
			"document_url":"{{ webhook.fileUrl }}",
			"document_urls_json":"{{ webhook.fileUrlsJson }}",
			"document_allowed_hosts":["shop.example.test","stage.shop.example.test"],
			"document_required":true,
			"continue_on_failure":true,
			"failure_required":true
		}
	}`)
	var plan actionPlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatalf("unmarshal action plan: %v", err)
	}
	if plan.Config.DocumentURL != "{{ webhook.fileUrl }}" {
		t.Fatalf("document URL was not preserved: %q", plan.Config.DocumentURL)
	}
	if plan.Config.DocumentURLsJSON != "{{ webhook.fileUrlsJson }}" {
		t.Fatalf("document URL list was not preserved: %q", plan.Config.DocumentURLsJSON)
	}
	wantHosts := []string{"shop.example.test", "stage.shop.example.test"}
	if !reflect.DeepEqual(plan.Config.DocumentAllowedHosts, wantHosts) {
		t.Fatalf("document allowlist was not preserved: %#v", plan.Config.DocumentAllowedHosts)
	}
	if !plan.Config.DocumentRequired {
		t.Fatal("document required flag was not preserved")
	}
	if !plan.Config.ContinueOnFailure || !plan.Config.FailureRequired {
		t.Fatal("independent required failure semantics were not preserved")
	}
	kernelPlan := kernelNode(plan)
	if !kernelPlan.ContinueOnFailure || !kernelPlan.FailureRequired {
		t.Fatal("kernel node lost independent required failure semantics")
	}

	encoded, err := json.Marshal(plan.Config)
	if err != nil {
		t.Fatalf("marshal action config: %v", err)
	}
	var materialized map[string]any
	if err := json.Unmarshal(encoded, &materialized); err != nil {
		t.Fatalf("decode materialized action config: %v", err)
	}
	if materialized["document_urls_json"] != "{{ webhook.fileUrlsJson }}" {
		t.Fatalf("materialized document URL list is missing: %#v", materialized)
	}
	if materialized["document_required"] != true {
		t.Fatalf("materialized document required flag is missing: %#v", materialized)
	}
	if materialized["failure_required"] != true {
		t.Fatalf("materialized failure required flag is missing: %#v", materialized)
	}
}

func TestActionPlanPreservesYClientsCertificateConfig(t *testing.T) {
	raw := []byte(`{
		"action_key":"yclients_issue",
		"action_type_key":"yclients.certificate.issue",
		"config":{
			"action_dependency_id":"7c6f9c6b-2fc8-4ce2-a7a3-19d338b7b101",
			"operation_key":"{{ webhook.operationKey }}",
			"order_id":"{{ webhook.orderId }}",
			"order_item_id":"{{ webhook.orderItemId }}",
			"instance_id":"{{ webhook.instanceId }}",
			"certificate_number":"{{ webhook.certificateNumber }}",
			"product_kind":"{{ webhook.productKind }}",
			"product_title":"{{ webhook.productTitle }}",
			"amount_rub":"{{ webhook.amountRub }}",
			"service_ids_json":"{{ webhook.serviceIdsJson }}"
		}
	}`)
	var plan actionPlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatalf("unmarshal YClients action plan: %v", err)
	}
	encoded, err := json.Marshal(plan.Config)
	if err != nil {
		t.Fatalf("marshal YClients action config: %v", err)
	}
	var materialized map[string]any
	if err := json.Unmarshal(encoded, &materialized); err != nil {
		t.Fatalf("decode YClients action config: %v", err)
	}
	want := map[string]string{
		"operation_key": "{{ webhook.operationKey }}", "order_id": "{{ webhook.orderId }}",
		"order_item_id": "{{ webhook.orderItemId }}", "instance_id": "{{ webhook.instanceId }}",
		"certificate_number": "{{ webhook.certificateNumber }}", "product_kind": "{{ webhook.productKind }}",
		"product_title": "{{ webhook.productTitle }}", "amount_rub": "{{ webhook.amountRub }}",
		"service_ids_json": "{{ webhook.serviceIdsJson }}",
	}
	for key, value := range want {
		if materialized[key] != value {
			t.Fatalf("YClients config %s was not preserved: %#v", key, materialized[key])
		}
	}
}
