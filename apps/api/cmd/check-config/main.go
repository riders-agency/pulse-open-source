package main

import (
	"flag"
	"fmt"
	"os"

	"riders.agency/pulse/api/internal/platform/config"
)

func main() {
	roleName := flag.String("role", "api", "startup role: api or worker")
	flag.Parse()

	role, ok := parseRole(*roleName)
	if !ok {
		fmt.Fprintln(os.Stderr, "role must be api or worker")
		os.Exit(2)
	}
	cfg := config.Load()
	if err := config.ValidateStartup(role, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "config validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("config validation ok schema=%s role=%s\n", config.SchemaVersion, role)
}

func parseRole(value string) (config.StartupRole, bool) {
	switch value {
	case string(config.RoleAPI):
		return config.RoleAPI, true
	case string(config.RoleWorker):
		return config.RoleWorker, true
	default:
		return "", false
	}
}
