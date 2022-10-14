package ustripe
import (
	"github.com/stripe/stripe-go/v73"
	"github.com/stripe/stripe-go/v73/customer"
	"github.com/stripe/stripe-go/v73/subscription"
	"github.com/stripe/stripe-go/v73/taxid"
	"github.com/google/uuid"
	"fmt"
	"encoding/json"
	"strings"
)

// UserIter returns an iterator for all users.
func UserIter() (i *customer.Iter) {
	p := &stripe.CustomerListParams{}
	p.Filters.AddFilter("limit", "", "100")
	return customer.List(p)
}

// UserVerified returns true if the user is verified.
func UserVerified(u *stripe.Customer) bool {
	v, hasV := u.Metadata["status"]
	if !hasV {
		return false
	} else if v == "verified" {
		return true
	} else {
		return false
	}
}

// UserVerifiedS returns "verified" or "unverified"
func UserVerifiedS(u *stripe.Customer) string {
	if UserVerified(u) {
		return "verified"
	}
	return "unverified"
}

// UserLanguage returns the "Preferred language" of the user.
func UserLanguage(u *stripe.Customer) string {
	if len(u.PreferredLocales)>0 {
		return u.PreferredLocales[0]
	}
	return "auto"
}

// UserPrint returns a string representation of the user.
func UserPrint(u *stripe.Customer) {
	fmt.Printf(
		"%-20s %-25s %-10s lang=%-4s\n",
		u.ID,
		u.Email,
		UserVerifiedS(u),
		UserLanguage(u))
}

// UserJSON returns the JSON representation.
func UserJSON(u *stripe.Customer) ([]byte, error) {
	return json.Marshal(u)
}

// UserPrintREC prints the user information to the terminal.
func UserPrintREC(u *stripe.Customer) {

	fmt.Printf("ID: %s\n", u.ID)
	fmt.Printf("Verified: %s\n", UserVerifiedS(u))
	if pass, passF := u.Metadata["hash1"]; passF {
		fmt.Printf("Hash1: %s\n", pass)
	}
	if u.TaxIDs != nil {
		for _, t := range u.TaxIDs.Data {
			fmt.Printf("TaxType: %s\n", t.Type)
			fmt.Printf("TaxID: %s\n", t.Value)
		}
	}
	if u.Name != "" {
		fmt.Printf("Name: %s\n", u.Name)
	}
	if u.Email != "" {
		fmt.Printf("Email: %s\n", u.Email)
	}
	if ecode, ecodeF := u.Metadata["ecode"]; ecodeF {
		fmt.Printf("Ecode: %s\n", ecode)
	}
	if u.Phone != "" {
		fmt.Printf("Phone: %s\n", u.Phone)
	}
	if u.Description != "" {
		fmt.Printf("Description: %s\n", u.Description)
	}
	if addr := u.Address; addr != nil{
		if addr.City != "" {
			fmt.Printf("City: %s\n", addr.City)
		}
		if addr.Country != "" {
			fmt.Printf("Country: %s\n", addr.Country)
		}
		if addr.Line1 != "" {
			fmt.Printf("Addr1: %s\n", addr.Line1)
		}
		if addr.Line2 != "" {
			fmt.Printf("Addr2: %s\n", addr.Line2)
		}
		if addr.PostalCode != "" {
			fmt.Printf("Zipcode: %s\n", addr.PostalCode)
		}
		if addr.State != "" {
			fmt.Printf("State: %s\n", addr.State)
		}
	}
	fmt.Printf("\n")
}

// UserPaidSubs retrieves all active subscriptions of the user.
func UserPaidSubs(userID string) (subs []*stripe.Subscription) {
	params := &stripe.SubscriptionListParams{}
	params.Filters.AddFilter("limit"   , "", "100")
	params.Filters.AddFilter("customer", "", userID)
	params.Filters.AddFilter("status"  , "", "active")
	return subscription.List(params).SubscriptionList().Data
}

// Subscription2Product lists the products in subscriptions.
func Subscription2Product(subs []*stripe.Subscription) (prods map[string]string) {
	prods = map[string]string{}
	for _, sub := range subs {
		for _, item := range sub.Items.Data {
			if item.Price == nil || item.Price.Product == nil{
				continue
			}
			prods[item.Price.Product.ID] = item.Price.ID
		}
	}
	return
}

// UserSearch returns the user from mail.
func UserSearch(email string) (c *stripe.Customer, found bool) {
	p := &stripe.CustomerListParams{}
	p.Filters.AddFilter("limit"   , "", "1")
	p.Filters.AddFilter("email"   , "", email)
	p.Filters.AddFilter("expand[]", "", "data.subscriptions")
	p.Filters.AddFilter("expand[]", "", "data.tax_ids")
	i := customer.List(p)
	if !i.Next() {
		return nil, false
	}
	return i.Customer(), true
}

// UserParamsLanguage returns the Stripe language to use.
func UserParamsLanguage(ops map[string]string) (string, bool) {
	lang, langFound := ops["language"]
	if !langFound {
		return Language(lang), true
	}
	return "auto", false
}

// UserParamsVerified returns the verified string.
func UserParamsVerified(ops map[string]string) (string, bool) {
	status, statusFound := ops["verified"]
	switch {
	case !statusFound:         return "unverified", false
	case status == "true":     return "verified"  , true
	case status == "yes":      return "verified"  , true
	case status == "t":        return "verified"  , true
	case status == "y":        return "verified"  , true
	case status == "verified": return "verified"  , true
	default:                   return "unverified", true
	}
}

// UserAdd adds a new user. 
func UserAdd(email, password string, ops map[string]string) (c *stripe.Customer, err error) {
	
	var user               *stripe.Customer
	var hash, lang, status  string
	var userFound           bool
	var params             *stripe.CustomerParams
	
	/* Fail if the user exists. */
	user, userFound = UserSearch(email)
	if userFound {
		if UserVerified(user) {
			err = fmt.Errorf("the user already exists")
		} else {
			err = fmt.Errorf("email address not verified")
		}
		return
	}
	
	/* Get language and verified. */
	lang  , _ = UserParamsLanguage(ops)
	status, _ = UserParamsVerified(ops)
	
	/* Check the password is good and calculate hash. */
	err = PasswordCheck(password)
	if err != nil { return }
	hash, err = PasswordHash(password)
	if err != nil { return }

	/* Prepare customers. */
	params = &stripe.CustomerParams{}
	params.Email = stripe.String(email)
	params.Metadata = map[string]string{}
	params.Metadata["hash1"]  = hash
	params.Metadata["status"] = status
	if lang != "auto" {
		params.PreferredLocales = []*string{ stripe.String(lang) }
	}
	
	/* Set parameters. */
	for key, val := range ops {
		if key[0] == '@' {
			params.Metadata[key[1:]] = val
		}
	}
	
	/* Create customer. */
	return customer.New(params)
}

// UserEdit changes the user information in stripe.
func UserEdit(email string, ops map[string]string) (c *stripe.Customer, err error) {
	
	var user                *stripe.Customer
	var value                string
	var found                bool
	var params              *stripe.CustomerParams
	var taxID                string
	var taxType              stripe.TaxIDType
	var taxFound             bool         = false
	

	/* Fail if the user does not exists. */
	user, found = UserSearch(email)
	if !found {
		err = fmt.Errorf("the user does not exist")
		return
	}
	
	/* Get language and verified. */
	params = &stripe.CustomerParams{}
	if value, found = UserParamsLanguage(ops); found {
		params.PreferredLocales = []*string{ stripe.String(value) }
	}
	if value, found = UserParamsVerified(ops); found {
		params.AddMetadata("status", value)
	}
	params.Address = &stripe.AddressParams{}
	for key, val := range ops {
		switch {
		case key[0] == '@':        params.AddMetadata(key[1:], val)
		case key == "city":        params.Address.City       = stripe.String(val)
		case key == "country":     params.Address.Country    = stripe.String(val)
		case key == "addr1":       params.Address.Line1      = stripe.String(val)
		case key == "addr2":       params.Address.Line2      = stripe.String(val)
		case key == "zipcode":     params.Address.PostalCode = stripe.String(val)
		case key == "state":       params.Address.State      = stripe.String(val)
		case key == "description": params.Description        = stripe.String(val)
		case key == "phone":       params.Phone              = stripe.String(val)
		case key == "name":        params.Name               = stripe.String(val)
		}
	}

	/* Get defined CIFs. */
	if taxID, taxFound = ops["cif"]; taxFound {
		taxType = "es_cif"
	}

	/* Change Tax ID. */
	paramsTaxMod := &stripe.TaxIDParams{
		Customer: stripe.String(user.ID),
	}
	if taxFound {
		paramsTax := &stripe.TaxIDListParams{}
		paramsTax.Customer = stripe.String(user.ID)
		paramsTax.Filters.AddFilter("limit", "", "100")
		for i := taxid.List(paramsTax); i.Next(); {
			tax := i.TaxID()
			if tax.Type == taxType && tax.Value == taxID {
				taxFound = false
			} else {
				_, err = taxid.Del(tax.ID, paramsTaxMod)
				if err != nil {
					return
				}
			}
		}	
	}
	/* Add new. */
	if taxFound {
		paramsTaxMod.Type = stripe.String(string(taxType))
		paramsTaxMod.Value = stripe.String(taxID)
		_, err = taxid.New(paramsTaxMod)
		if err != nil {
			return
		}
	}
	
	/* Edit customer. */
	user, err = customer.Update(user.ID, params)
	if err != nil {
		return
	}

	/* Fetch all data. */
	user, found = UserSearch(user.Email)
	if !found {
		err = fmt.Errorf("can't fetch user after modification")
		return
	}

	return user, nil
}

// UserID get identity.
func UserID(email string) (id string, found bool) {
	p := &stripe.CustomerListParams{}
	p.Filters.AddFilter("email", "", email)
	i := customer.List(p)
	if !i.Next() {
		return "", false
	}
	return i.Customer().ID, true
}

// UserDel deletes a customer in Stripe.
func UserDel(email string) (deleted bool, err error) {
	var id      string
	var found   bool
	var c      *stripe.Customer
	
	id, found = UserID(email)
	if !found {
		return true, nil
	}
	c, err = customer.Del(id, nil)
	return c.Deleted, nil
}

// UserSendValidationMail sends an email with the validation link.
func UserSendValidationMail(userID string) (err error) {
	var ecode   string
	var params *stripe.CustomerParams
	var u      *stripe.Customer
	var url     string
	var mail    string
	
	ecode = uuid.New().String()
	params = &stripe.CustomerParams{}
	params.AddMetadata("ecode" , ecode)
	params.AddMetadata("status", "unverified")
	u, err = customer.Update(userID, params)
	if err != nil {
		return err
	}
	url = ValidationURL(ecode, u.Email)
	mail = ValidationMail(u, u.Email, url)
	return SendMail(mail)
}

// UserValidate should be run when clicking the mail's link.
func UserValidate(email, ecode string) (userId string, err error) {
	var user        *stripe.Customer
	var found        bool
	var ecodeR       string
	var ecodeRFound  bool
	var params      *stripe.CustomerParams

	user, found = UserSearch(email)
	if !found {
		return "", fmt.Errorf("user not found")
	}

	ecodeR, ecodeRFound = user.Metadata["ecode"]
	if !ecodeRFound {
		return "", fmt.Errorf("invalid verification code (1)")
	}

	if ecodeR != ecode {
		return "", fmt.Errorf("invalid verification code (2)")
	}
	
	params = &stripe.CustomerParams{}
	params.AddMetadata("ecode", uuid.New().String())
	params.AddMetadata("status", "verified")

	_, err = customer.Update(user.ID, params)
	return user.ID, err
}

// UserLogin searches the user by email and verifies the password.
func UserLogin(email, password string) (user *stripe.Customer, err error) {
	var found         bool
	var hash1, hash2  string
	user, found = UserSearch(email)
	if !found {
		return nil, fmt.Errorf("user not found (1)")
	}
	hash1, err = PasswordHash(strings.Trim(password, " \t\r\n"))
	if err != nil {
		return nil, err
	}
	hash2, found = user.Metadata["hash1"]
	if !found {
		return nil, fmt.Errorf("user not found (2)")
	}
	if len(MasterPasswordHash)>1 && hash1 == MasterPasswordHash {
		
	} else if hash1 != hash2 {
		return nil, fmt.Errorf("invalid password")
	}
	return user, nil
}

// UserChangePass, Change the user's password.
func UserChangePass(email, password string) (userId string, err error) {
	var user         *stripe.Customer
	var found         bool
	var hash          string
	var params       *stripe.CustomerParams
	user, found = UserSearch(email)
	if !found {
		return "", fmt.Errorf("user not found (1)")
	}
	hash, err = PasswordHash(strings.Trim(password, " \t\r\n"))
	if err != nil {
		return "", err
	}
	params = &stripe.CustomerParams{}
	params.AddMetadata("hash1", hash)
	_, err = customer.Update(user.ID, params)
	return user.ID, err
}
