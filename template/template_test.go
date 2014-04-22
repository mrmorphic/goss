package template

import (
	"fmt"
	"github.com/mrmorphic/goss"
	"github.com/mrmorphic/goss/config"
	"testing"
)

// test source that contains no <% or $. We expect just a single literal.
func TestLiteralOnly(t *testing.T) {
	_, e := newParser().parseSource("<html><body>simple</body></html>")
	if e != nil {
		t.Error(e.Error())
		return
	}
}

func TestVariants(t *testing.T) {
	conf, e := config.ReadFromFile("template_test_config.json")
	if e != nil {
		t.Error(e.Error())
		return
	}

	// Give goss the configuration object.
	e = goss.SetConfig(conf)
	if e != nil {
		t.Error(e.Error())
		return
	}

	fmt.Printf("Configuration is %s\n", configuration)

	sources := []string{
		"abc$foosdasd",
		"xyz<% base_tag %>abc",
		"abc <% include Footer %> xyz",
		"<% if $MyDinner==\"quiche\" && $YourDinner==\"kipper\" %>some text<% end_if %>",
		"<% if not $DinnerInOven %>I'm going out for dinner tonight.<% end_if %>",
		"<% if $Number>=\"5\" && $Number<=\"10\" %>foo<% end_if %>",
		"abc{$foo}xyz",
	}

	for _, s := range sources {
		fmt.Printf("scanning source: %s\n", s)
		scanner := newScanner(s)
		for {
			tk, _ := scanner.scanToken()
			fmt.Printf("...%s\n", tk.printable())
			if tk.kind == TOKEN_END_SOURCE {
				break
			}
		}
		fmt.Printf("parsing source: %s\n", s)
		_, e := newParser().parseSource(s)
		if e != nil {
			t.Error(e.Error())
			return
		}
	}
}
