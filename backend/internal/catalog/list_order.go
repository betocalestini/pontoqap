package catalog

import "strings"

func catalogOrderClause(f ListProductsFilter) string {
	if f.Admin {
		return "ORDER BY p.name ASC"
	}
	sort := strings.TrimSpace(f.Sort)
	switch sort {
	case "price_asc":
		return "ORDER BY sku_price.min_price ASC NULLS LAST, p.name ASC"
	case "price_desc":
		return "ORDER BY sku_price.min_price DESC NULLS LAST, p.name ASC"
	case "name":
		return "ORDER BY p.name ASC"
	case "purchases":
		return "ORDER BY cust_purch.purchase_qty DESC, p.name ASC"
	case "", "default":
		if f.Search == "" && f.Category == "" {
			return "ORDER BY (p.promo_active AND p.promo_quantity_remaining > 0) DESC, p.name ASC"
		}
		return "ORDER BY p.name ASC"
	default:
		return "ORDER BY p.name ASC"
	}
}
