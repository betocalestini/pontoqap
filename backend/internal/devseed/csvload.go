package devseed

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	productsFile  = "products.csv"
	customersFile = "customers.csv"
)

// ProductCSVRow is one row from products.csv.
type ProductCSVRow struct {
	Name           string
	Slug           string
	SKUCode        string
	Category       string
	Unit           string
	UnitCostCents  int64
	MarginPercent  *float64
	StockQty       int
	ImageSlug      string
}

// CustomerCSVRow is one row from customers.csv.
type CustomerCSVRow struct {
	Name              string
	Email             string
	CreditLimitCents  int64
	Collaborator      bool
}

// ResolveDataDir returns the directory containing seed CSV files.
func ResolveDataDir(cfg Config) string {
	if cfg.DataDir != "" {
		return cfg.DataDir
	}
	if v := os.Getenv("SEED_DATA_DIR"); v != "" {
		return v
	}
	return "devdata"
}

func loadProductsCSV(dir string, defaultStockQty int) ([]ProductCSVRow, error) {
	if defaultStockQty <= 0 {
		defaultStockQty = DefaultStockQty
	}
	path := filepath.Join(dir, productsFile)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("abrir %s: %w", path, err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ler %s: %w", path, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("%s: precisa de cabeçalho e ao menos uma linha", path)
	}
	header := records[0]
	col := indexColumns(header, "name", "slug", "sku_code", "category", "unit", "unit_cost_cents", "margin_percent", "stock_qty", "image_slug")
	var out []ProductCSVRow
	for i, rec := range records[1:] {
		if rowEmpty(rec) {
			continue
		}
		if len(rec) < len(header) {
			return nil, fmt.Errorf("%s linha %d: colunas insuficientes", path, i+2)
		}
		cost, err := strconv.ParseInt(strings.TrimSpace(rec[col["unit_cost_cents"]]), 10, 64)
		if err != nil || cost <= 0 {
			return nil, fmt.Errorf("%s linha %d: unit_cost_cents inválido", path, i+2)
		}
		stock, err := parseStockQty(strings.TrimSpace(rec[col["stock_qty"]]), defaultStockQty)
		if err != nil {
			return nil, fmt.Errorf("%s linha %d: stock_qty inválido", path, i+2)
		}
		row := ProductCSVRow{
			Name:          strings.TrimSpace(rec[col["name"]]),
			Slug:          strings.TrimSpace(rec[col["slug"]]),
			SKUCode:       strings.TrimSpace(rec[col["sku_code"]]),
			Category:      strings.TrimSpace(rec[col["category"]]),
			Unit:          strings.TrimSpace(rec[col["unit"]]),
			UnitCostCents: cost,
			StockQty:      stock,
			ImageSlug:     strings.TrimSpace(rec[col["image_slug"]]),
		}
		if row.Unit == "" {
			row.Unit = "UN"
		}
		mp := strings.TrimSpace(rec[col["margin_percent"]])
		if mp != "" {
			f, err := strconv.ParseFloat(mp, 64)
			if err != nil {
				return nil, fmt.Errorf("%s linha %d: margin_percent inválido", path, i+2)
			}
			row.MarginPercent = &f
		}
		out = append(out, row)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: nenhum produto", path)
	}
	return out, nil
}

func loadCustomersCSV(dir, domain string) ([]CustomerCSVRow, error) {
	path := filepath.Join(dir, customersFile)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("abrir %s: %w", path, err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ler %s: %w", path, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("%s: precisa de cabeçalho e ao menos uma linha", path)
	}
	header := records[0]
	col := indexColumns(header, "name", "email", "credit_limit_cents", "collaborator")
	var out []CustomerCSVRow
	for i, rec := range records[1:] {
		if rowEmpty(rec) {
			continue
		}
		limit, err := strconv.ParseInt(strings.TrimSpace(rec[col["credit_limit_cents"]]), 10, 64)
		if err != nil || limit < 0 {
			return nil, fmt.Errorf("%s linha %d: credit_limit_cents inválido", path, i+2)
		}
		collab, err := parseBool(strings.TrimSpace(rec[col["collaborator"]]))
		if err != nil {
			return nil, fmt.Errorf("%s linha %d: collaborator inválido", path, i+2)
		}
		out = append(out, CustomerCSVRow{
			Name:             strings.TrimSpace(rec[col["name"]]),
			Email:            expandDomain(strings.TrimSpace(rec[col["email"]]), domain),
			CreditLimitCents: limit,
			Collaborator:     collab,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: nenhum cliente", path)
	}
	return out, nil
}

func expandDomain(s, domain string) string {
	return strings.ReplaceAll(s, "{domain}", domain)
}

func parseStockQty(s string, defaultQty int) (int, error) {
	if s == "" || s == "0" {
		return defaultQty, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("negativo")
	}
	if n == 0 {
		return defaultQty, nil
	}
	return n, nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "sim":
		return true, nil
	case "false", "0", "no", "nao", "não":
		return false, nil
	default:
		return false, fmt.Errorf("valor booleano: %q", s)
	}
}

func indexColumns(header []string, names ...string) map[string]int {
	m := make(map[string]int, len(names))
	for i, h := range header {
		key := strings.TrimSpace(strings.ToLower(h))
		m[key] = i
	}
	for _, n := range names {
		if _, ok := m[n]; !ok {
			panic(fmt.Sprintf("coluna CSV ausente: %s", n))
		}
	}
	return m
}

func rowEmpty(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func categorySlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "categoria"
	}
	return out
}

func productImageStorageKey(slug string) string {
	if slug == "" {
		return ""
	}
	if strings.HasPrefix(slug, "/") {
		return slug
	}
	return "/product-images/" + slug
}

// Test hooks exported to devseed_test.
func ExpandDomainForTest(s, domain string) string  { return expandDomain(s, domain) }
func ParseBoolForTest(s string) (bool, error)       { return parseBool(s) }
func LoadProductsCSVForTest(dir string) ([]ProductCSVRow, error) { return loadProductsCSV(dir, DefaultStockQty) }
func ParseStockQtyForTest(s string, defaultQty int) (int, error) { return parseStockQty(s, defaultQty) }
