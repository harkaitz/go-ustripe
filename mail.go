package ustripe

import (
	"github.com/stripe/stripe-go/v73"
	"fmt"
)

var ValidationMail func (c *stripe.Customer, to, url string) (mail string) = DefaultValidationMail
var ValidationURL  func (ecode, email string)                (url  string) = DefaultValidationURL

func DefaultValidationMail(c *stripe.Customer, to, url string) (s string) {
	subject     := "Confirm your mail with Lotorius"
	contentType := "text/html; charset=UTF-8"
	return fmt.Sprintf(""    +
		"To: %s"               + "\n" +
		"Subject: %s"          + "\n" +
		"Content-Type: %s"     + "\n" +
		""                     + "\n" +
		"<html>"               + "\n" +
		"  <body>"             + "\n" +
		"    <p>"              + "\n" +
		"      To complete the sign-up, we need you to confirm your" + "\n" +
		"      mail address by clicking the next button."            + "\n" +
		"    </p>"                                                   + "\n" +
		"    <p>"                                                    + "\n" +
		"      <a href=\"%s\">Confirm email</a>"                     + "\n" +
		"    </p>"                                                   + "\n" +
		"    <p>"                                                    + "\n" +
		"      Once your email has been validated you will receive"  + "\n" +
		"      a testing account."                                   + "\n" +
		"    </p>"                                                   + "\n" +
		"  </body>"                                                  + "\n" +
		"</html>"                                                    + "\n",
		to, subject, contentType, url)
}

func DefaultValidationURL(ecode, email string) (url string) {
	return "https://efferox.com/wellcome?mcode=" + ecode + "&email=" + email
}
