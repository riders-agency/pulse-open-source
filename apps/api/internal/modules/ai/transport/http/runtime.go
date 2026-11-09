package http

import (
	"os"
	"strings"
)

func aiEmployeeRuntime() string {
	runtime := strings.TrimSpace(os.Getenv("AI_EMPLOYEE_RUNTIME"))
	if runtime == "" {
		return "langgraph"
	}
	return runtime
}
