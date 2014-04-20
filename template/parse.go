package template

import (
	"fmt"
)

type parser struct {
	scanner *scanner
}

func newParser() *parser {
	return &parser{}
}

// parseSource converts the source of a template into a compiled template object. The approach taken is to repetitively
// reduce the source string into chunks.
func (p *parser) parseSource(source string) (*compiledTemplate, error) {
	p.scanner = newScanner(source)

	result, e := p.parseContent()
	if e != nil {
		return nil, e
	}

	// ensure there is no left-over
	fmt.Printf("Parsed result: %s\n", result)
	return result, nil
}

// parseContent parses content, which is broadly a sequence of literals, $xxx and <% %> blocks. It is also
// value to see TOKEN_END_SOURCE. This is used to parse the top-level, but will stop when it sees something it
// doesn't know, because it is also used to parse nested content.
func (p *parser) parseContent() (*compiledTemplate, error) {
	fmt.Printf("parseContent\n")
	result := newCompiledTemplate()

loop:
	for {
		tk, e := p.scanner.scanToken()
		fmt.Printf("...tk: %s\n", tk)
		if e != nil {
			return nil, e
		}

		switch {
		case tk.kind == TOKEN_LITERAL:
			result.push(newChunkLiteral(tk.value))
		case tk.kind == TOKEN_OPEN:
			// parse the content of the tag
			ch, e := p.parseTag()
			if e != nil {
				return nil, e
			}

			// we expect a TOKEN_CLOSE
			e = p.expectKind(TOKEN_CLOSE)
			if e != nil {
				return nil, e
			}

			result.push(ch)
		case tk.kind == TOKEN_END_SOURCE:
			p.scanner.putBack(tk)
			break loop
		case tk.isSym("{") || tk.isSym("$"):
			// variable/function injection
		}
	}

	return result, nil
}

// read one token, and check that it is of the required kind. Return an error on scanning error, or if the kind doesn't match.
func (p *parser) expectKind(kind tokenKind) error {
	tk, e := p.scanner.scanToken()
	if e != nil {
		return e
	}
	if tk.kind != kind {
		return fmt.Errorf("Expected token of kind '%s', got '%s' instead.", kind, tk.printable())
	}
	return nil
}

// read one token, and check that it is a symbol with the required value. Return an error on scanning error, or if the kind doesn't match.
func (p *parser) expectSym(s string) error {
	tk, e := p.scanner.scanToken()
	if e != nil {
		return e
	}
	if tk.kind != TOKEN_SYMBOL {
		return fmt.Errorf("Expected token of kind '%s', got '%s' instead.", TOKEN_SYMBOL, tk.printable())
	}
	if tk.value != s {
		return fmt.Errorf("Expected symbol '%s', got '%s' instead.", s, tk.printable())
	}
	return nil
}

// parse the what is in between <% and %>, but not including those tokens. This is largely a dispatcher for what is in the tag.
func (p *parser) parseTag() (chunk, error) {
	tk, e := p.scanner.scanToken()
	if e != nil {
		return nil, e
	}

	// eat the first token and dispatch
	switch {
	case tk.isIdent("include"):
		return p.parseInclude()
	case tk.isIdent("if"):
		return p.parseIf()
	case tk.isIdent("loop"):
		return p.parseLoop()
	case tk.isIdent("with"):
		return p.parseWith()
	case tk.isIdent("require"):
		return p.parseRequire()
	case tk.isIdent("base_tag"):
		// nothing else to parse
		return newChunkBaseTag(), nil
	case tk.isIdent("t"):
		return p.parseTranslation()
	case tk.isIdent("cached"):
		return p.parseCached()
	}

	// shouldn't be anything else starting a tag
	return nil, fmt.Errorf("Invalid token for a tag '%s'", tk.printable())
}

// Parse an include tag. The 'include' keyword has already been scanned. In it's simplest form, this will just have another identifier which
// is the include file. It may also have a comma-separate list of name=value pairs which provide context into the included template.
func (p *parser) parseInclude() (chunk, error) {
	tk, e := p.scanner.scanToken()
	if e != nil {
		return nil, e
	}

	// we expect an identifier next, which is the name of the include template.
	if tk.kind != TOKEN_IDENT {
		return nil, fmt.Errorf("Expected identifier for the included template, got '%s'", tk.printable())
	}

	// parse the included file
	path := configuration.includesPath + tk.value
	fmt.Printf("requesting include file %s\n", path)
	compiled, e := compileTemplate(path)

	if e != nil {
		return nil, fmt.Errorf("In include file %s: %s", path, e)
	}

	// @todo parse the variable bindings afterwards and put these in the chunk as well.

	return newChunkInclude(compiled), nil
}

func (p *parser) parseIf() (chunk, error) {
	return nil, nil
}

// parse a <% loop %> ... <% end_loop %> structure. "<% loop" has already been parsed.
func (p *parser) parseLoop() (chunk, error) {

	return nil, nil
}

func (p *parser) parseWith() (chunk, error) {
	return nil, nil
}

func (p *parser) parseRequire() (chunk, error) {
	return nil, nil
}

func (p *parser) parseTranslation() (chunk, error) {
	return nil, nil
}

// Parse cached block. At this point, template caching is not supported. So we ignore everything in the tag,
// then parseContent to get everything inside the tag, then expect the closing tag. The chunk we return is just
// from parseContent.
func (p *parser) parseCached() (chunk, error) {
	return nil, nil
}
