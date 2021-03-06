package template

import (
	"bytes"
	"fmt"
	"github.com/mrmorphic/goss"
	"github.com/mrmorphic/goss/config"
	"github.com/mrmorphic/goss/requirements"
	"net/http"
	"testing"
)

// responseCapture is a simple ResponseWriter that captures the bytes written to the response.
type responseCapture struct {
	response []byte
}

func (r *responseCapture) Header() http.Header {
	// empty headers
	return make(map[string][]string)
}

func (r *responseCapture) Write(bytes []byte) (int, error) {
	r.response = append(r.response, bytes...)
	return len(bytes), nil
}

func (r *responseCapture) WriteHeader(status int) {
}

// utility function to compile and execute template source using the given context
func compileAndExecute(source string, context interface{}) ([]byte, error) {
	compiled, e := newParser().parseSource(source, true, "")
	if e != nil {
		return nil, e
	}
	exec := newExecuter([]*compiledTemplate{compiled}, context, requirements.NewRequirements())
	return exec.render()
}

// helper function to compile and execute a list of sources and comparing the output to
// expected results, using a supplied context. 'sources' is a map where the key is the
// source to be parsed, and the value is the expected result.
func testSourceList(sources map[string]string, context interface{}, t *testing.T) {
	for source, expected := range sources {
		b, e := compileAndExecute(source, context)
		if e != nil {
			t.Error(e.Error())
			return
		}
		if string(b) != expected {
			t.Errorf("Expected '%s', got '%s'", expected, b)
		}
	}
}

// test source that contains no <% or $. We expect just a single literal.
func TestLiteralOnly(t *testing.T) {
	_, e := newParser().parseSource("<html><body>simple</body></html>", true, "")
	if e != nil {
		t.Error(e.Error())
		return
	}
}

// handle loading the test configuration
func configure() error {
	conf, e := config.ReadFromFile("test/config.json")
	if e != nil {
		return e
	}

	// Give goss the configuration object.
	e = goss.SetConfig(conf)
	if e != nil {
		return e
	}

	return nil
}

// @todo refactor into a number of case-specific tests
func TestVariants(t *testing.T) {
	e := configure()
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
		scanner := newScanner(s, "")
		for {
			tk, _ := scanner.scanToken()
			fmt.Printf("...%s\n", tk.printable())
			if tk.kind == TOKEN_END_SOURCE {
				break
			}
		}
		fmt.Printf("parsing source: %s\n", s)
		_, e := newParser().parseSource(s, true, "")
		if e != nil {
			t.Error(e.Error())
			return
		}
	}
}

// @todo refactor into a number of case-specific tests
func TestExec(t *testing.T) {
	e := configure()
	if e != nil {
		t.Error(e.Error())
		return
	}

	source := `this is some markup for $name, with gratuitous nested var: {$parent.child}. $function1(name)  $function0
	<%if v1==v2 %>rhubarb equal bananas<% else %>of course a banana is not rhubarb.<% end_if %>  we like food
	$salutation(title, name)`
	compiled, e := newParser().parseSource(source, true, "")

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
	exec := newExecuter([]*compiledTemplate{compiled}, context, requirements.NewRequirements())
	bytes, e := exec.renderChunk(compiled.chunk)

	if e != nil {
		t.Error(e.Error())
	}

	fmt.Printf("bytes are: %s\n", bytes)
}

// test Layout inclusion.
func TestLayout(t *testing.T) {
	e := configure()
	if e != nil {
		t.Error(e.Error())
		return
	}

	capture := &responseCapture{}
	context := make(map[string]interface{})

	e = RenderWith(capture, []string{"TestA", "TestALayout"}, context, nil)

	if string(capture.response) != "startTestALayoutend" {
		t.Errorf("main/layout response was not expected: %s", capture.response)
	}
}

func TestComment(t *testing.T) {
	sources := map[string]string{
		`abc<%-- comment --%>def`: `abcdef`,
		`abc<%-- comment --%>`:    `abc`,
		`<%-- comment --%>def`:    `def`,
		`abc<%-- multiline
	comment --%>def`: `abcdef`,
	}

	context := make(map[string]interface{})

	testSourceList(sources, context, t)

	// now test for failure parsing an unclosed comment
	_, e := newParser().parseSource("abc<%-- comment", true, "")
	if e == nil {
		t.Error("Expected error about unterminated comment, didn't get an error")
	}
}

func TestBaseTag(t *testing.T) {
	sources := map[string]string{
		`<% base_tag %>`: `<base href="http://localhost/"><!--[if lte IE 6]></base><![endif]-->`,
	}
	context := make(map[string]interface{})

	testSourceList(sources, context, t)
}

func TestWith(t *testing.T) {
	sources := map[string]string{
		`<% with parent %>$foo<% end_with %>`: `bar`,
	}
	context := make(map[string]interface{})
	child := make(map[string]interface{})
	child["foo"] = "bar"
	context["parent"] = child

	testSourceList(sources, context, t)
}

func makeItem(title string) interface{} {
	return map[string]interface{}{
		"Title": title,
	}
}

func TestLoop(t *testing.T) {
	sources := map[string]string{
		`[<% loop Items %>$Title<% end_loop %>]`: `[abc]`,
	}
	context := make(map[string]interface{})
	items := make([]interface{}, 0)
	items = append(items, makeItem("a"))
	items = append(items, makeItem("b"))
	items = append(items, makeItem("c"))
	context["Items"] = items

	testSourceList(sources, context, t)
}

func TestRequireJS(t *testing.T) {
	source := `<html><head><title>x</title></head><body><% require javascript("themes/simple/javascript/test.js") %><div>test</div></body></html>`
	context := map[string]interface{}{}

	b, e := compileAndExecute(source, context)
	if e != nil {
		t.Error(e.Error())
		return
	}
	if bytes.Index(b, []byte(`<script type="text/javascript" src="themes/simple/javascript/test.js"></script></body>`)) < 0 {
		t.Errorf("Expecting to see test.js embedded before body")
	}
}

func TestRequireCSS(t *testing.T) {
	source := `<html><head><title>x</title></head><body><% require css("themes/simple/css/test.css") %><div>test</div></body></html>`
	context := map[string]interface{}{}

	b, e := compileAndExecute(source, context)
	if e != nil {
		t.Error(e.Error())
		return
	}
	if bytes.Index(b, []byte(`<link rel="stylesheet" type="text/css" href="themes/simple/css/test.css" /></head>`)) < 0 {
		t.Errorf("Expecting to see test.js embedded before body")
	}
}

func TestRequireThemedCSS(t *testing.T) {
	source := `<html><head><title>x</title></head><body><% require themedCSS("test") %><div>test</div></body></html>`
	context := map[string]interface{}{}

	b, e := compileAndExecute(source, context)
	if e != nil {
		t.Error(e.Error())
		return
	}
	if bytes.Index(b, []byte(`<link rel="stylesheet" type="text/css" href="themes/simple/css/test.css" /></head>`)) < 0 {
		t.Errorf("Expecting to see test.js embedded before body")
	}
}

func TestInclude(t *testing.T) {
	sources := map[string]string{
		`[<% include Footer %>]`: `[test footer]`,
	}
	context := map[string]interface{}{}

	testSourceList(sources, context, t)
}

func TestString(t *testing.T) {
	sources := map[string]string{
		`<div class="$foo"></div>`: `<div class="bar"></div>`,
	}
	context := map[string]interface{}{}
	context["foo"] = "bar"

	testSourceList(sources, context, t)
}

// Test that == and = are treated the same
func TestEquals(t *testing.T) {
	sources := map[string]string{
		`<% if $foo="bar" %>yes<% end_if %>`:  "yes",
		`<% if $foo=="bar" %>yes<% end_if %>`: "yes",
	}
	context := map[string]interface{}{}
	context["foo"] = "bar"

	testSourceList(sources, context, t)
}

// Test for a variable substitution following by an if-not. Case seen in debugging:
//	$ClassName<% if not $Menu(2) %>...
func TestVarIfNot(t *testing.T) {
	sources := map[string]string{
		`$var<% if not $foo(2) %>x<% end_if %>`: "v",  // case where cond is false
		`$var<% if not $foo(1) %>x<% end_if %>`: "vx", // case where cond is true
	}
	context := map[string]interface{}{}
	context["var"] = "v"
	context["foo"] = func(i string) bool {
		return i == "2"
	}

	testSourceList(sources, context, t)
}
