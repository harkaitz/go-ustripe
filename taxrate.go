package ustripe

import (
	"github.com/stripe/stripe-go/v73"
	"github.com/stripe/stripe-go/v73/taxrate"
	"fmt"
)

// TaxList returns all defined taxes.
func TaxList() (i *taxrate.Iter) {
	p := &stripe.TaxRateListParams{}
	p.Filters.AddFilter("limit", "", "100")
	return taxrate.List(p)
}

// TaxPrint prints the tax to terminal.
func TaxPrint(t *stripe.TaxRate) {
	var active string
	if t.Active {
		active = "true"
	} else {
		active = "false"
	}
	fmt.Printf(
		"%s %s %s %f active=%s\n",
		t.ID,
		t.Jurisdiction,
		t.DisplayName,
		t.Percentage,
		active)
}
