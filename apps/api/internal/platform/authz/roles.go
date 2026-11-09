package authz

func companyRoleAllows(role, permission string) bool {
	switch permission {
	case PermRead:
		return role == "owner" || role == "admin" || role == "member"
	case PermWrite, PermAdmin:
		return role == "owner" || role == "admin"
	default:
		return false
	}
}

func projectRoleAllows(role, permission string) bool {
	switch permission {
	case PermAdmin:
		return role == "owner" || role == "admin"
	case PermWrite:
		switch role {
		case "owner", "admin", "operations_manager", "project_manager", "account_manager", "team_lead", "employee", "contractor", "member":
			return true
		}
	case PermRead:
		switch role {
		case "owner", "admin", "operations_manager", "project_manager", "account_manager", "team_lead", "employee", "contractor", "member", "client_guest", "finance", "hr", "read_only_auditor":
			return true
		}
	}
	return false
}
