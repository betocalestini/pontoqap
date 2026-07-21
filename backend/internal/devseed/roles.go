package devseed

import "github.com/google/uuid"

var (
	roleCustomer          = uuid.MustParse("a0000000-0000-4000-8000-000000000003")
	roleManager           = uuid.MustParse("a0000000-0000-4000-8000-000000000002")
	roleInventoryOperator = uuid.MustParse("a0000000-0000-4000-8000-000000000004")
	roleFinanceOperator   = uuid.MustParse("a0000000-0000-4000-8000-000000000005")
	collaboratorCategory  = uuid.MustParse("d0000000-0000-4000-8000-000000000001")
)

type staffSpec struct {
	Email    string
	Name     string
	RoleID   uuid.UUID
	RoleCode string
}

func staffSpecs(domain string) []staffSpec {
	return []staffSpec{
		{Email: "demo-gerente@" + domain, Name: "Gerente Demo", RoleID: roleManager, RoleCode: "manager"},
		{Email: "demo-gerente2@" + domain, Name: "Gerente Demo 2", RoleID: roleManager, RoleCode: "manager"},
		{Email: "demo-estoque@" + domain, Name: "Operador Estoque Demo", RoleID: roleInventoryOperator, RoleCode: "inventory_operator"},
		{Email: "demo-financeiro@" + domain, Name: "Financeiro Demo", RoleID: roleFinanceOperator, RoleCode: "finance_operator"},
	}
}
