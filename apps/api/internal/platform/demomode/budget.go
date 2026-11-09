package demomode

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrBudget = errors.New("demo_model_budget_exhausted")
var modelBudgetMu sync.Mutex

type budgetState struct {
	Day   string         `json:"day"`
	Total int            `json:"total"`
	Users map[string]int `json:"users"`
}

// ReserveModelCall persists a reservation before calling the provider, including failed calls.
// The public demo has one API process; this is deliberately not a distributed quota.
func ReserveModelCall(userID string) error {
	if !Enabled() {
		return nil
	}
	path := os.Getenv("PULSE_DEMO_BUDGET_PATH")
	if path == "" || userID == "" {
		return ErrBudget
	}
	modelBudgetMu.Lock()
	defer modelBudgetMu.Unlock()
	day := time.Now().UTC().Format("2006-01-02")
	state := budgetState{Day: day, Users: map[string]int{}}
	raw, err := os.ReadFile(path)
	if err == nil {
		if json.Unmarshal(raw, &state) != nil {
			return ErrBudget
		}
	} else if !os.IsNotExist(err) {
		return ErrBudget
	}
	if state.Day != day {
		state = budgetState{Day: day, Users: map[string]int{}}
	}
	if state.Users == nil {
		state.Users = map[string]int{}
	}
	if state.Total >= 300 || state.Users[userID] >= 30 {
		return ErrBudget
	}
	state.Total++
	state.Users[userID]++
	raw, err = json.Marshal(state)
	if err != nil {
		return ErrBudget
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return ErrBudget
	}
	if err = os.WriteFile(path+".next", raw, 0600); err != nil {
		return ErrBudget
	}
	if err = os.Rename(path+".next", path); err != nil {
		return ErrBudget
	}
	return nil
}
