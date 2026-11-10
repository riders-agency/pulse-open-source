# Riders Pulse contributor rules

Read `standards/active/AGENTS.md` and the active stack standards before changes.
Use an isolated local database and synthetic data. Never test against a live
customer workspace. Keep business authorization in Go, and Python orchestration
behind the typed gRPC contracts. Do not commit secrets or tenant data.

Run applicable Go, web, Python and contract checks. The module exception register is empty. Keep owner commands and versioned read
contracts intact; do not add exceptions or suppress checks to obtain a passing result.
Public baseline migrations are for fresh installations, not existing private
production databases. Subsequent schema changes are append-only.
