package identity

// StaffRoleCodes are internal (non-customer) panel roles.
var StaffRoleCodes = []string{
	"system_admin",
	"manager",
	"inventory_operator",
	"finance_operator",
}

func IsStaffRole(code string) bool {
	for _, c := range StaffRoleCodes {
		if c == code {
			return true
		}
	}
	return false
}

func HasStaffRole(roles []string) bool {
	for _, r := range roles {
		if IsStaffRole(r) {
			return true
		}
	}
	return false
}

func HasSystemAdminRole(roles []string) bool {
	for _, r := range roles {
		if r == "system_admin" {
			return true
		}
	}
	return false
}
