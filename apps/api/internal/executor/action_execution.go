package executor

import "time"

type actionExecution struct {
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
