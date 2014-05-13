package template

import (
	"github.com/mrmorphic/goss"
	"github.com/mrmorphic/goss/requirements"
	"io/ioutil"
	"net/http"
)

// syntax to consider:
// $var
// {$var}
// \$var
// $property(params)
// $property.subproperty
// <% include Name %>
// <% if value %>x<% end_if %>
// <% with value %><% end_with %>
// <% include MemberDetails PageTitle=$Top.Title, PageID=$Top.ID %>
// <% require themedCSS("LeftNavMenu") %>
// <% if $MyDinner=="kipper" %>
// $MyDinner==$YourDinner
// <% if not $DinnerInOven %>
// <% if $MyDinner=="kipper" || $MyDinner=="salmon" %>
// <% if $MyDinner=="quiche" && $YourDinner=="kipper" %>
// <% if $Number>="5" && $Number<="10" %>
// <% loop $Children %> <% end_loop %>
// <% loop $Children.Sort(Title) %> DataList
// <% loop $Children.Limit(10) %> DataList
// <% loop $Children.Reverse %> DataList
// <% loop $Children.Reverse.Limit(3) %>
// $Modulus(value, offset) // returns an int
// $MultipleOf(factor, offset) // returns a boolean.
// $Up
// $Top
// <%t Member.WELCOME 'Welcome {name} to {site}' name=$Member.Name site="Foobar.com" %>
// <% base_tag %>
// In order to prevent this situation, the SSViewer template renderer will automatically rewrite any fragment link that doesn't specify a URL before the fragment, prefixing the URL of the current page
// <% cached values %> <% end_cached %>

// kinds of things:
// - literal text
// - conditions
//	  - "not" condition
//	  - value operator value
//	  - "||" and "&&" operators
// - value
//    - may be atomic
//	  - may be contexts (implement DataLocator)
//
// other notes:
// - includes can be processed at compile time
//   - also binds
// - context stack. with, loops, include explicitly push things on the stack. renames with include.
// - cached blocks introduce no new semantics, and can be ignored.

// compiledTemplate is just a list of chunks to process in order. Some chunks may contain nested compiledTemplate.
type compiledTemplate struct {
	// a compiled template only contains one chunk, which is always a chunkBlock
	chunk *chunk
}

func newCompiledTemplate() *compiledTemplate {
	return &compiledTemplate{}
}

// compiledTemplates maps template paths (relative to theme folder) to compiled templates.
var compiledTemplates map[string]*compiledTemplate

// RenderWith renders the template(s) using the locator to fill in variable references and writes to the
// writer. 'templates' is an array of SilverStripe templates minus the ".ss" extension. If there is one template,
// it is assumed to be in the base templates folder. If two are present, the first is the base template, the
// second is the $Layout template.
func RenderWith(w http.ResponseWriter, templates []string, context interface{}, require goss.RequirementsProvider, request *http.Request) error {
	if require == nil {
		require = requirements.NewRequirements()
	}

	// make templates relative to templates folder
	if len(templates) > 1 {
		templates[1] = "Layout/" + templates[1]
	}

	// corresponding compiled templates will go in here for execution.
	var compiled []*compiledTemplate
	first := true
	for _, t := range templates {
		template, e := compileTemplate(t, first)
		first = false
		if e != nil {
			return e
		}
		compiled = append(compiled, template)
	}

	// execute the template using the data locator
	r, e := executeTemplate(compiled, context, require, request)
	if e != nil {
		return e
	}

	// finally write the result
	_, e = w.Write([]byte(r))
	return e
}

func executeTemplate(templates []*compiledTemplate, context interface{}, require goss.RequirementsProvider, request *http.Request) ([]byte, error) {
	exec := newExecuter(templates, context, require, request)
	return exec.render()
}

// compileTemplate takes a template by path (relative to templates folder) and compiles it into a compiledTemplate.
// If there is a parse error, that is returned. If the template is already in compiledTemplates, the pre-compiled version
// is returned. Otherwise it is added to compiledTemplates as well
func compileTemplate(path string, mainTemplate bool) (*compiledTemplate, error) {
	// Get it from cache if it's there
	result := compiledTemplates[path]
	if result != nil {
		return result, nil
	}

	// read from path
	filename := configuration.templatesPath + path + ".ss"
	s, e := ioutil.ReadFile(filename)
	if e != nil {
		return nil, e
	}

	// convert the returned []byte to a string, so that the parsing will handle UTF8 characters.
	source := string(s)

	result, e = newParser().parseSource(source, mainTemplate, filename)
	if e != nil {
		return nil, e
	}

	compiledTemplates[path] = result
	return result, nil
}

func init() {
	compiledTemplates = make(map[string]*compiledTemplate)
}
