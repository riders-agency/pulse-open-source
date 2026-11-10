package main

import "testing"

func TestReadContractRejectsWritesUnknownConsumerAndSecrets(t *testing.T) {
	c := readContract{Name: "pulse_read.projects_v1", Owner: "projects", Table: "projects", Columns: []string{"project_id"}, Consumers: []string{"ai"}}
	ownership := map[string]ownershipRow{"projects": {PrimaryTables: []string{"projects"}}, "ai": {}}
	if err := validateReadContract(c, ownership); err != nil {
		t.Fatal(err)
	}
	contracts := map[string]readContract{c.Name: c}
	for _, op := range []string{"UPDATE", "INTO", "DELETE FROM"} {
		if checkReadReference(contracts, "ai", op, c.Name) == nil {
			t.Fatalf("allowed %s", op)
		}
	}
	if checkReadReference(contracts, "clients", "FROM", c.Name) == nil {
		t.Fatal("allowed unknown consumer")
	}
	if checkReadReference(contracts, "ai", "JOIN", c.Name) != nil {
		t.Fatal("denied registered reader")
	}
	c.Columns = append(c.Columns, "hook_secret_hash")
	if validateReadContract(c, ownership) == nil {
		t.Fatal("allowed credential column")
	}
}
