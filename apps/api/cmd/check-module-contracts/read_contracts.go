package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type readContract struct {
	Name, Owner, Table string
	Columns, Consumers []string
}

func loadReadContracts(root string, ownership map[string]ownershipRow) (map[string]readContract, error) {
	raw, err := os.ReadFile(filepath.Join(root, "docs/architecture/read-contracts.json"))
	if err != nil {
		return nil, err
	}
	var list []readContract
	if err = json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	migration, err := os.ReadFile(filepath.Join(root, "db/migrations/000115_module_read_contracts.up.sql"))
	if err != nil {
		return nil, err
	}
	out := map[string]readContract{}
	for _, c := range list {
		if err := validateReadContract(c, ownership); err != nil {
			return nil, err
		}
		if _, exists := out[c.Name]; exists {
			return nil, fmt.Errorf("duplicate read contract %s", c.Name)
		}
		expected := "CREATE VIEW " + c.Name + " WITH (security_barrier = true) AS SELECT " + strings.Join(c.Columns, ", ") + " FROM public." + c.Table + " OFFSET 0;"
		if !strings.Contains(string(migration), expected) {
			return nil, fmt.Errorf("read contract %s does not match its immutable migration", c.Name)
		}
		ownerRaw, err := os.ReadFile(filepath.Join(root, "apps/api/internal/modules", c.Owner, "contracts/read_models.json"))
		if err != nil {
			return nil, err
		}
		var ownerContracts []readContract
		if err = json.Unmarshal(ownerRaw, &ownerContracts); err != nil {
			return nil, err
		}
		expectedJSON, _ := json.Marshal(c)
		found := false
		for _, candidate := range ownerContracts {
			actual, _ := json.Marshal(candidate)
			if string(actual) == string(expectedJSON) {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("read contract %s is missing from owner manifest", c.Name)
		}
		out[c.Name] = c
	}
	return out, nil
}

func validateReadContract(c readContract, ownership map[string]ownershipRow) error {
	if c.Name != "pulse_read."+c.Table+"_v1" || !slices.Contains(ownership[c.Owner].PrimaryTables, c.Table) {
		return fmt.Errorf("invalid read contract owner/name %s", c.Name)
	}
	if len(c.Columns) == 0 || len(c.Consumers) == 0 {
		return fmt.Errorf("empty read contract %s", c.Name)
	}
	identifier := regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
	forbidden := regexp.MustCompile(`password|token|secret|encrypted|country_relays_json|pending_email`)
	for _, column := range c.Columns {
		if !identifier.MatchString(column) || forbidden.MatchString(column) {
			return fmt.Errorf("unsafe column %s in %s", column, c.Name)
		}
	}
	for _, consumer := range c.Consumers {
		if _, ok := ownership[consumer]; !ok {
			return fmt.Errorf("unknown consumer %s", consumer)
		}
	}
	return nil
}

func checkReadReference(contracts map[string]readContract, source, operation, name string) error {
	c, ok := contracts[name]
	if !ok {
		return fmt.Errorf("unknown read contract %s", name)
	}
	if operation != "FROM" && operation != "JOIN" {
		return fmt.Errorf("write through read contract %s", name)
	}
	if !slices.Contains(c.Consumers, source) {
		return fmt.Errorf("unregistered read consumer %s -> %s", source, name)
	}
	return nil
}
