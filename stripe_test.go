package ustripe

import (
	"testing"
)

func TestLanguage(t *testing.T) {
	res := map[string]string{
		"es_ES.UTF-8": "es",
		"es_ES":       "es",
		"es":          "es",
	}

	for k, v := range res {
		if Language(k) != v {
			t.Fatalf(`Language("`+k+`") != "`+v+`"`)
		}
	}
}
