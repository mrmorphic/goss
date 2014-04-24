package template

import (
	"fmt"
	"github.com/mrmorphic/goss"
	"github.com/mrmorphic/goss/config"
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
	compiled, e := newParser().parseSource(source, true)
	if e != nil {
		return nil, e
	}
	exec := newExecuter([]*compiledTemplate{compiled}, context, NewDefaultLocator())
	return exec.renderChunk(compiled.chunk)
}

// test source that contains no <% or $. We expect just a single literal.
func TestLiteralOnly(t *testing.T) {
	_, e := newParser().parseSource("<html><body>simple</body></html>", true)
	if e != nil {
		t.Error(e.Error())
		return
	}
}

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
		scanner := newScanner(s)
		for {
			tk, _ := scanner.scanToken()
			fmt.Printf("...%s\n", tk.printable())
			if tk.kind == TOKEN_END_SOURCE {
				break
			}
		}
		fmt.Printf("parsing source: %s\n", s)
		_, e := newParser().parseSource(s, true)
		if e != nil {
			t.Error(e.Error())
			return
		}
	}
}

func TestExec(t *testing.T) {
	e := configure()
	if e != nil {
		t.Error(e.Error())
		return
	}

	source := `this is some markup for $name, with gratuitous nested var: {$parent.child}. $function1(name)  $function0
	<%if v1==v2 %>rhubarb equal bananas<% else %>of course a banana is not rhubarb.<% end_if %>  we like food
	$salutation(title, name)`
	compiled, e := newParser().parseSource(source, true)

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
	exec := newExecuter([]*compiledTemplate{compiled}, context, NewDefaultLocator())
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
	validSources := make(map[string]string)
	validSources["abc<%-- comment --%>def"] = "abcdef"
	validSources["abc<%-- comment --%>"] = "abc"
	validSources["<%-- comment --%>def"] = "def"
	validSources[`abc<%-- multiline
	comment --%>def`] = "abcdef"

	context := make(map[string]interface{})

	for source, result := range validSources {
		bytes, e := compileAndExecute(source, context)

		if e != nil {
			t.Errorf("Unexpected exec error: %s [in source: %s]", e, source)
			return
		}

		if string(bytes) != result {
			t.Errorf("Unexpected result: %s [in source: %s]", bytes, source)
		}
	}

	// now test for failure parsing an unclosed comment
	_, e := newParser().parseSource("abc<%-- comment", true)
	if e == nil {
		t.Error("Expected error about unterminated comment, didn't get an error")
	}
}

func TestBaseTag(t *testing.T) {
	source := "<% base_tag %>"
	context := make(map[string]interface{})
	bytes, e := compileAndExecute(source, context)

	if e != nil {
		t.Error(e.Error())
		return
	}

	if string(bytes) != `<base href="http://localhost /><!--[if lte IE 6]></base><![endif]-->` {
		t.Errorf("Incorrect base tag calculation: '%s': check test/config.json", bytes)
	}
}

func TestWith(t *testing.T) {
	source := `<% with parent %>$foo<% end_with %>`
	context := make(map[string]interface{})
	child := make(map[string]interface{})
	child["foo"] = "bar"
	context["parent"] = child
	bytes, e := compileAndExecute(source, context)
	if e != nil {
		t.Error(e.Error())
		return
	}
	if string(bytes) != "bar" {
		t.Errorf("Expected 'foo' but got '%s'", bytes)
	}
}

func makeItem(title string) interface{} {
	return map[string]interface{}{
		"Title": title,
	}
}

func TestLoop(t *testing.T) {
	source := `[<% loop Items %>$Title<% end_loop %>]`
	context := make(map[string]interface{})
	items := make([]interface{}, 0)
	items = append(items, makeItem("a"))
	items = append(items, makeItem("b"))
	items = append(items, makeItem("c"))
	context["Items"] = items

	bytes, e := compileAndExecute(source, context)
	if e != nil {
		t.Error(e.Error())
		return
	}
	if string(bytes) != "[abc]" {
		t.Errorf("Unexpected loop output: %s", bytes)
	}
}
