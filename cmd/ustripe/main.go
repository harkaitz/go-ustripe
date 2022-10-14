package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	"github.com/harkaitz/ustripe"
)

const help string =
`Usage: ustripe [OPTIONS...]

Simple mechanism to handle subscriptions with Stripe.

Environment variables:

    RELEASE_MODE, STRIPE[_TEST]_SECRET_KEY, SENDMAIL_COMMAND

Subcommands:

    hash1  p=PASSWORD     : Calculate hash of the password.
    login  e=EMAIL p=PASS : Check password.
    chpass e=EMAIL p=PASS : Change password.
    www    ...            : Open the stripe dashboard and resources.

    tax-list  : List defined taxes.

    prod-list              : List defined products.
    prod-price PRODUCTS... : Convert from product to price.

    user-list                            : List users.
    user-get-json  e=EMAIL               : Print user JSON.
    user-get-subs  e=EMAIL [PROD1[,...]] : User's subscriptions.
    user-info      e=EMAIL               : User information.
    
    user-add      e=EMAIL p=PASS l=LANG  : Add new user.
    user-del      EMAIL...               : Delete user.
    user-edit     e=EMAIL PARAMS...      : Edit user.
    user-mail-v   e=EMAIL                : Send validation mail.
    user-validate e=EMAIL ecode=ECODE    : Validate.

    subscribe ... : Create session (us,uc,c|e required)

      url_success | us   = SUCCESS-URL
      url_cancel  | uc   = CANCEL-URL
      customer    | c    = CUSTOMER or
      email       | e    = EMAIL
      reference   | r    = REFERENCE
      browse             = y|n
      @PROD=NUM[,TAX]`
const copyrightLine string =
`Bug reports, feature requests to gemini|https://harkadev.com/oss
Copyright (c) 2022 Harkaitz Agirre, harkaitz.aguirre@gmail.com`

func main() {

	if len(os.Args)==1 || os.Args[1]=="-h" || os.Args[1]=="--help" || os.Args[1]=="help" {
		fmt.Println(help + "\n\n" + copyrightLine)
		return
	}

	cmd  := os.Args[1]
	argv := os.Args[2:]
	
	switch cmd {
	case "hash1":
		kvs, _ :=mainParams(argv, "password")
		hash, err := ustripe.PasswordHash(kvs["password"])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", hash)
	case "login":
		kvs, _ :=mainParams(argv, "email", "password")
		user, err := ustripe.UserLogin(kvs["email"], kvs["password"])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", user.ID)
	case "chpass":
		kvs, _ :=mainParams(argv, "email", "password")
		_, err := ustripe.UserChangePass(kvs["email"], kvs["password"])
		if err != nil {
			log.Fatal(err)
		}
	case "www":
		err := mainBrowse(argv...)
		if err != nil {
			log.Fatal(err)
		}
	case "tax-list":
		for i := ustripe.TaxList(); i.Next(); {
			ustripe.TaxPrint(i.TaxRate())
		}
	case "prod-list":
		for i := ustripe.ProductList(); i.Next(); {
			p := i.Product()
			ustripe.ProductPrint(p, true, true)
		}
	case "prod-price":
		for _, prodID := range argv {
			priceID, err := ustripe.Product2Price(prodID)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", priceID)
		}
	case "user-list":
		for i := ustripe.UserIter(); i.Next(); {
			c := i.Customer()
			ustripe.UserPrint(c)
		}
	case "user-get-json":
		kvs, _ := mainParams(argv, "email")
		user, found := ustripe.UserSearch(kvs["email"])
		if found {
			customerJSON, err := ustripe.UserJSON(user)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", customerJSON)
		}
	case "user-get-subs":
		kvs, args := mainParams(argv, "email")
		user, found := ustripe.UserSearch(kvs["email"])
		if found {
			prods := ustripe.Subscription2Product(ustripe.UserPaidSubs(user.ID))
			if len(args)>0 {
				for _, p := range args {
					if _, f := prods[p]; f {
						fmt.Printf("%s\n", p)
						break
					}
				}
			} else {
				for p := range prods {
					fmt.Printf("%s\n", p)
				}
			}
		}
	case "user-info":
		kvs, _ := mainParams(argv, "email")
		user, found := ustripe.UserSearch(kvs["email"])
		if found {
			ustripe.UserPrintREC(user)
		}
	case "user-add":
		kvs, _ := mainParams(argv, "email", "password")
		cus, err := ustripe.UserAdd(kvs["email"], kvs["password"], kvs)
		if err != nil {
			log.Fatal(err)
		}
		ustripe.UserPrint(cus)
	case "user-del":
		for _, email := range argv {
			_, err := ustripe.UserDel(email)
			if err != nil {
				log.Print(err)
			}
		}
	case "user-edit":
		kvs, _ := mainParams(argv, "email")
		cus, err := ustripe.UserEdit(kvs["email"], kvs)
		if err != nil {
			log.Fatal(err)
		}
		ustripe.UserPrintREC(cus)
	case "user-mail-v":
		kvs, _ :=mainParams(argv, "email")
		id, found := ustripe.UserID(kvs["email"])
		if !found {
			log.Fatal("Customer not found.")
		}
		err := ustripe.UserSendValidationMail(id)
		if err != nil {
			log.Fatal(err)
		}
	case "user-validate":
		kvs, _ :=mainParams(argv, "email", "ecode")
		_, err := ustripe.UserValidate(kvs["email"], kvs["ecode"])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", kvs["email"])
	case "subscribe":
		kvs, _ :=mainParams(argv)
		ses, err := ustripe.SubscriptionNew(kvs)
		if err != nil {
			log.Fatal(err)
		}
		switch {
		case kvs["browse"] == "y":
			err = ustripe.OpenBrowser(ses.URL)
			if err != nil {
				log.Fatal(err)
			}
		default:
			ustripe.SubscriptionPrintREC(ses)
		}
	default:
		log.Fatal("Invalid argument: " + cmd)
	}
}

func mainParams(args []string, reqs ...string) (m map[string]string, r []string) {
	m = map[string]string{}
	r = []string{}
	for _, arg := range args {
		key, val, splet := strings.Cut(arg, "=")
		if splet {
			switch key {
			case "e"       : key = "email";
			case "p"       : key = "password";
			case "pass"    : key = "password";
			case "l"       : key = "language";
			case "lang"    : key = "language";
			case "language": key = "language";
			case "us"      : key = "url_success";
			case "uc"      : key = "url_cancel";
			case "c"       : key = "customer";
			case "t"       : key = "tax_rate";
			case "r"       : key = "reference";
			case "v"       : key = "verified";
			}
			m[key] = val
		} else {
			r = append(r, arg)
		}
	}
	missing := ""
	for _, req := range reqs {
		if _, f := m[req]; f == false {
			missing += " " + req
		}
	}
	if len(missing)>0 {
		log.Fatalf("Missing parameters:%s", missing)
	}
	return
}

func mainBrowse(args ...string) (err error) {
	help := []string{
		`doc-api       : API documentation.`,
		`doc-testing   : Fake testing user account doc.`,
		`dashboard     : Show graphs.`,
		`customers     : Show customers.`,
		`api-keys      : Open API key configuration place.`,
		`api-keys-test : Open API key configuration place (Testing).`,
		`webhooks      : Open STRIPE webhooks page.`,
		`cfg-invoice   : Set the company NIF.`,
		`cfg-branding  : Set the company colors etc.`,
		`cfg-profile   : Set the language, ...`,
		`cfg-team      : Team members, etc.`,
	}
	if len(args)==0 {
		fmt.Println(strings.Join(help, "\n"))
		return nil
	}
	cmd := args[0]
	rel := ustripe.ReleaseMode
	switch {
	case cmd == "doc-api"          : ustripe.OpenBrowser("https://stripe.com/docs/api/balance/balance_retrieve?lang=go")
	case cmd == "doc-testing"      : ustripe.OpenBrowser("https://stripe.com/docs/testing")
	case cmd == "dashboard"        : ustripe.OpenBrowser("https://dashboard.stripe.com/login")
	case cmd == "customers" &&  rel: ustripe.OpenBrowser("https://dashboard.stripe.com/customers")
	case cmd == "customers" && !rel: ustripe.OpenBrowser("https://dashboard.stripe.com/test/customers")
	case cmd == "api-keys"  &&  rel: ustripe.OpenBrowser("https://dashboard.stripe.com/apikeys")
	case cmd == "api-keys"  && !rel: ustripe.OpenBrowser("https://dashboard.stripe.com/test/apikeys")
	case cmd == "webhooks"         : ustripe.OpenBrowser("https://dashboard.stripe.com/webhooks")
	case cmd == "cfg-invoice"      : ustripe.OpenBrowser("https://dashboard.stripe.com/settings/billing/invoice")
	case cmd == "cfg-branding"     : ustripe.OpenBrowser("https://dashboard.stripe.com/settings/branding")
	case cmd == "cfg-profile"      : ustripe.OpenBrowser("https://dashboard.stripe.com/settings/user")
	case cmd == "cfg-team"         : ustripe.OpenBrowser("https://dashboard.stripe.com/settings/team")
	default: return fmt.Errorf("Invalid argument: "+ cmd)
	}
	return nil
}
