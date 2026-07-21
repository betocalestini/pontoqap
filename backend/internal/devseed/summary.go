package devseed

import "fmt"

type Result struct {
	Products      int
	Customers     int
	Staff         int
	Orders        int
	ClosedPeriods int
	OpenPeriods   int
}

func (r Result) Print(cfg Config) {
	fmt.Println()
	fmt.Println("=== Seed de demonstração concluído ===")
	fmt.Printf("  Produtos seed:     %d\n", r.Products)
	fmt.Printf("  Clientes:          %d\n", r.Customers)
	fmt.Printf("  Funcionários:      %d\n", r.Staff)
	fmt.Printf("  Pedidos:           %d\n", r.Orders)
	fmt.Printf("  Competências fechadas: %d\n", r.ClosedPeriods)
	fmt.Printf("  Competências abertas:  %d\n", r.OpenPeriods)
	fmt.Println()
	fmt.Println("Senha padrão (demo):", cfg.Password)
	fmt.Println()
	fmt.Println("Funcionários (painel admin + loja):")
	for _, s := range staffSpecs(cfg.Domain) {
		fmt.Printf("  %-12s %s\n", s.RoleCode+":", s.Email)
	}
	fmt.Println()
	fmt.Println("Clientes: demo-cliente-001@" + cfg.Domain + " …")
	fmt.Println("Bootstrap (inalterado): admin@loja.local / ChangeMe123!")
	fmt.Println()
}
