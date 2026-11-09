package tools

import (
	"encoding/json"
	"fmt"
)

// EmployeePolicy narrows the code-owned catalog; it cannot grant new capabilities.
type EmployeePolicy struct {
	AllowedTools             *[]string `json:"allowed_tools"`
	DeniedTools              []string  `json:"denied_tools"`
	WriteRequiresApproval    *bool     `json:"write_requires_approval"`
	CriticalRequiresApproval *bool     `json:"critical_requires_elevated_approval"`
}

func ParseEmployeePolicy(raw []byte) (EmployeePolicy, error) {
	var policy EmployeePolicy
	if len(raw) == 0 {
		return policy, nil
	}
	if err := json.Unmarshal(raw, &policy); err != nil {
		return policy, fmt.Errorf("invalid employee tool policy")
	}
	if policy.WriteRequiresApproval != nil && !*policy.WriteRequiresApproval {
		return policy, fmt.Errorf("write approval cannot be disabled")
	}
	if policy.CriticalRequiresApproval != nil && !*policy.CriticalRequiresApproval {
		return policy, fmt.Errorf("critical approval cannot be disabled")
	}
	return policy, nil
}

func (p EmployeePolicy) Allows(tool string) bool {
	for _, denied := range p.DeniedTools {
		if denied == tool {
			return false
		}
	}
	if p.AllowedTools == nil {
		return true
	}
	for _, allowed := range *p.AllowedTools {
		if allowed == tool {
			return true
		}
	}
	return false
}
