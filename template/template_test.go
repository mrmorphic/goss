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

func TestExec(t *testing.T) {
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

	source := `this is some markup for $name, with gratuitous nested var: {$parent.child}. $function1(name)  $function0
	<%if v1==v2 %>rhubarb equal bananas<% else %>of course a banana is not rhubarb.<% end_if %>  we like food
	$salutation(title, name)`
	compiled, e := newParser().parseSource(source)

	if e != nil {
		t.Error(e.Error())
		return
	}

	context := make(map[string]interface{})
	context["foo"] = "bar"
	context["name"] = "mickey mouse"
	sub := make(map[string]interface{})
	sub["child"] = "foobar!"
	context["parent"] = sub
	context["function0"] = func() string {
		return "hey, function with zero parameters"
	}
	context["function1"] = func(s string) string {
		return "hello '" + s + "' from function 1!"
	}
	context["salutation"] = func(t string, n string) string {
		return t + " " + n
	}
	context["v1"] = "rhubarb"
	context["v1"] = "bananas"
	context["title"] = "dear"

	// evaluate it
	exec := newExecuter(context)
	bytes, e := exec.renderChunk(compiled.chunk)

	if e != nil {
		t.Error(e.Error())
	}

	fmt.Printf("bytes are: %s\n", bytes)
}
