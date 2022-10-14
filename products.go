package ustripe

import (
	"github.com/stripe/stripe-go/v73"
	"github.com/stripe/stripe-go/v73/product"
	"github.com/stripe/stripe-go/v73/checkout/session"
	"strconv"
	"strings"
	"fmt"
)

// ProductList returns all defined products.
func ProductList() (i *product.Iter) {
	p := &stripe.ProductListParams{}
	p.Filters.AddFilter("limit"   , "", "100")
	p.Filters.AddFilter("expand[]", "", "data.default_price")
	return product.List(p)
}

// ProductPrint .
func ProductPrint(p *stripe.Product, withPriceID, withPriceInfo bool) {
	price := p.DefaultPrice
	fmt.Printf("%-20s", p.ID)
	if price != nil && withPriceID {
		fmt.Printf(" p=%-30s", price.ID)
	}
	if price != nil && withPriceInfo {
		fmt.Printf(" i=%v%s",
			price.UnitAmount/100,
			price.Currency)
		if price.Recurring != nil {
			fmt.Printf(",%v,%s",
				price.Recurring.IntervalCount,
				price.Recurring.Interval)
		}
	}
	fmt.Printf(" n=%s\n", p.Name)
}

// ProductFetch .
func ProductFetch(prodID string) (p *stripe.Product, err error) {
	return product.Get(prodID, nil)
}

// Product2Price .
func Product2Price(prodID string) (priceID string, err error) {
	var prod *stripe.Product
	prod, err = ProductFetch(prodID)
	if err != nil {
		return
	}
	priceID = prod.DefaultPrice.ID
	return
}

// SubscriptionNew .
func SubscriptionNew(m map[string]string) (ses *stripe.CheckoutSession, err error) {

	var successURL, cancelURL    string
	var items                 []*stripe.CheckoutSessionLineItemParams
	var found                    bool
	var quantity                 int
	var quantityS, tax           string
	var priceID                  string
	var customerID               string = m["customer"]
	var email                    string = m["email"]
	var reference                string = m["reference"]

	items = []*stripe.CheckoutSessionLineItemParams {}

	successURL, found = m["success_url"]
	if !found {
		err = fmt.Errorf("missing success_url")
		return
	}
	
	cancelURL, found = m["cancel_url"]
	if !found {
		err = fmt.Errorf("missing cancel_url")
		return
	}

	switch {
	case len(customerID)>0:
	case len(email)>0:
		var customer *stripe.Customer
		customer, found = UserSearch(email)
		if !found {
			err = fmt.Errorf("user not found")
			return
		}
		customerID = customer.ID
	default:
		err = fmt.Errorf("missing email/customer")
		return
	}
	
	for key, val := range m {
		
		if len(key)==0 || key[0]!='@' {
			continue
		}

		priceID, err = Product2Price(key[1:])
		if err != nil {
			return
		}
		
		quantityS, tax, found = strings.Cut(val, ",")
		if !found {
			tax, err = DefaultTaxRate()
			if err != nil {
				return
			}
		}
		quantity, err = strconv.Atoi(quantityS)
		if err != nil {
			return
		}
		
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(priceID),
			Quantity: stripe.Int64(int64(quantity)),
			TaxRates: []*string{stripe.String(tax)},
		})
	}

	if len(items) == 0 {
		err = fmt.Errorf("missing product list")
		return
	}
	
	params := &stripe.CheckoutSessionParams{}
	params.SuccessURL = stripe.String(successURL)
	params.CancelURL  = stripe.String(cancelURL)
	params.LineItems  = items
	params.Mode       = stripe.String(string(stripe.CheckoutSessionModeSubscription))
	params.Customer   = stripe.String(customerID)
	if len(reference)>0 {
		params.ClientReferenceID = stripe.String(reference)
	}
	
	return session.New(params)
}

// SubscriptionPrint .
func SubscriptionPrintREC(ses *stripe.CheckoutSession) {
	fmt.Printf("ID: %s\n",  ses.ID)
	fmt.Printf("URL: %s\n", ses.URL)
}

