package ustripe

import (
	"github.com/stripe/stripe-go/v73"
	"os"
	"fmt"
	"strings"
)

var ReleaseMode         bool
var TaxRate             string = ""
var SendmailCommand     string = "msmtp -t"
var MasterPasswordHash  string = ""

func init() {
	var envKey, envTax string
	ReleaseMode = len(os.Getenv("RELEASE_MODE"))>0
	if (ReleaseMode) {
		envKey = "STRIPE_SECRET_KEY"
		envTax = "STRIPE_DEFAULT_TAXID"
	} else {
		envKey = "STRIPE_TEST_SECRET_KEY"
		envTax = "STRIPE_TEST_DEFAULT_TAXID"
	}

	stripe.Key = os.Getenv(envKey)
	if len(stripe.Key)==0 {
		panic("Please set " + envKey)
	}
	TaxRate = os.Getenv(envTax)
	
	s := os.Getenv("SENDMAIL_COMMAND")
	if len(s)>0 {
		SendmailCommand = s
	}

	MasterPasswordHash = os.Getenv("STRIPE_MASTER_PASSWORD_HASH1")
}

func Language(f string) (t string) {
	language, country_coding, _ := strings.Cut(f, "_")
	country, _, _               := strings.Cut(country_coding, ".")
	if language == "en" && country == "GB" { return "en-GB" }
	if language == "fr" && country == "CA" { return "fr-CA" }
	if language == "pt" && country == "BR" { return "pt-BR" }
	if language == "zh" && country == "HK" { return "zh-HK" }
	if language == "zh" && country == "TW" { return "zh-TW" }
	for _, l := range []string {
		"bg",     "cs",    "da",    "de",    "el",    "en",
		"es",     "et",    "fi",    "fil",   "fr",    "hr",
		"hu",     "id",    "it",    "ja",    "ko",    "lt",
		"lv",     "ms",    "mt",    "nb",    "nl",    "pl",
		"pt",     "ro",    "ru",    "sk",    "sl",    "sv",
		"th",     "tr",    "vi",    "zh",    "or",
	} {
		if language == l {
			return l
		}
	}
	return "auto"
}

func DefaultTaxRate() (t string, err error) {
	if (len(TaxRate)==0) {
		return "", fmt.Errorf("Tax rate not specified.")
	}
	return t, nil
}

