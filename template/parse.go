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
	result := newCompiledTemplate()

	chunk, e := p.parseContent()
	if e != nil {
		return nil, e
	}

	result.chunk = chunk

	// ensure there is no left-over
	fmt.Printf("Parsed result: \n%s\n", chunk.printable(0))

	return result, nil
}

// parseContent parses content, which is broadly a sequence of literals, $xxx and <% %> blocks. It is also
// value to see TOKEN_END_SOURCE. This is used to parse the top-level, but will stop when it sees something it
// doesn't know, because it is also used to parse nested content.
func (p *parser) parseContent() (*chunk, error) {
	fmt.Printf("parseContent\n")

	var chunks []*chunk

loop:
	for {
		tk, e := p.scanner.scanToken()
		fmt.Printf("...tk: %s\n", tk)
		if e != nil {
			return nil, e
		}

		switch {
		case tk.kind == TOKEN_LITERAL:
			chunks = append(chunks, newChunkLiteral(tk.value))
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

			chunks = append(chunks, ch)
		case tk.kind == TOKEN_END_SOURCE:
			p.scanner.putBack(tk)
			break loop
		case tk.isSym("{") || tk.isSym("$"):
			// variable/function injection
		}
	}

	return newChunkBlock(chunks), nil
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

// Expect 3 tokens in a sequence: TOKEN_OPEN, TOKEN_SYMBOL (matching tag), TOKEN_CLOSE
func (p *parser) expectTag(tag string) error {
	e := p.expectKind(TOKEN_OPEN)
	if e != nil {
		return e
	}
	e = p.expectSym(tag)
	if e != nil {
		return e
	}
	return p.expectKind(TOKEN_CLOSE)
}

// parse the what is in between <% and %>, but not including those tokens. This is largely a dispatcher for what is in the tag.
func (p *parser) parseTag() (*chunk, error) {
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
func (p *parser) parseInclude() (*chunk, error) {
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

func (p *parser) parseIf() (*chunk, error) {
	// parse condition
	// parse %>
	// parse truePart
	// test for <% else_if %>
	// test if <% else %> is present
	// parse <% end_if %>
	return nil, nil
}

// parse a <% loop %> ... <% end_loop %> structure. "<% loop" has already been parsed.
func (p *parser) parseLoop() (*chunk, error) {
	loopContext, e := p.parseExpr(true)
	if e != nil {
		return nil, e
	}

	e = p.expectKind(TOKEN_CLOSE)
	if e != nil {
		return nil, e
	}

	loopBody, e := p.parseContent()
	if e != nil {
		return nil, e
	}

	e = p.expectTag("end_loop")
	if e != nil {
		return nil, e
	}

	return newChunkLoop(loopContext, loopBody), nil
}

// parse a <% with %> ... <% end_with %> structure. "<% with" has already been parsed.
func (p *parser) parseWith() (*chunk, error) {
	context, e := p.parseExpr(true)
	if e != nil {
		return nil, e
	}

	e = p.expectKind(TOKEN_CLOSE)
	if e != nil {
		return nil, e
	}

	body, e := p.parseContent()
	if e != nil {
		return nil, e
	}

	e = p.expectTag("end_with")
	if e != nil {
		return nil, e
	}

	return newChunkWith(context, body), nil
}

func (p *parser) parseRequire() (*chunk, error) {
	return nil, nil
}

func (p *parser) parseTranslation() (*chunk, error) {
	return nil, nil
}

// Parse cached block. At this point, template caching is not supported. So we ignore everything in the tag,
// then parseContent to get everything inside the tag, then expect the closing tag. The chunk we return is just
// from parseContent.
func (p *parser) parseCached() (*chunk, error) {
	// parse an expression list
	// parse %>
	// parse content
	// parse <% end_cached %>
	// return chunk for content
	return nil, nil
}

// Parse an expression that provides a value. This handles many forms:
// - string literal
// - numeric literal
// - variable reference (possibly nested)
// - function reference
// This returns a chunk of type chunkExpr, which itself is a tree of such objects.
// It will attempt to parse as many tokens as possible to make a valid expression.
// 'topLevel' should be true on non-nested calls
func (p *parser) parseExpr(topLevel bool) (*chunk, error) {
	tk, e := p.scanner.scanToken()
	if e != nil {
		return nil, e
	}

	switch {
	case tk.kind == TOKEN_NUMBER:
		return newChunkExprValue(CHUNK_EXPR_NUMBER, tk.value), nil

	case tk.kind == TOKEN_STRING:
		return newChunkExprValue(CHUNK_EXPR_STRING, tk.value), nil

	case tk.isSym("("):
		// nested expression
		v, e := p.parseExpr(false)
		if e != nil {
			return nil, e
		}
		p.expectSym(")")
		return v, nil

	case tk.isSym("$"):
		// property or function to follow, but only valid if topLevel==true
		return p.parseVariableOrFn()

	case tk.kind == TOKEN_IDENT:
		// identifier, different cases
		switch {
		case tk.value == "not":
			// not <expr>
			sub, e := p.parseExpr(false)
			if e != nil {
				return nil, e
			}
			return newChunkExprValue(CHUNK_EXPR_NOT, sub), nil
		default:
			// put the identifier back, and ask to parse a variable
			p.scanner.putBack(tk)
			return p.parseVariableOrFn()
		}
	}

	return nil, nil
}

// parse a variable or function. This can include chained references.
func (p *parser) parseVariableOrFn() (*chunk, error) {
	tk, e := p.scanner.scanToken()
	if e != nil {
		return nil, e
	}
	if tk.kind != TOKEN_IDENT {
		return nil, fmt.Errorf("Expected identifier for a variable or function, got '%s'", tk.printable())
	}

	// check for open parentheses, this indicates a function call
	tk2, e := p.scanner.scanToken()
	if e != nil {
		return nil, e
	}

	var params *chunk

	if tk2.isSym("(") {
		params, e = p.parseExpressionList(true)
		if e != nil {
			return nil, e
		}
		e = p.expectSym(")")
		if e != nil {
			return nil, e
		}
	} else {
		p.scanner.putBack(tk2)
	}

	// at this point, tk.value is the name; params is nil for a variable, and a chunkBlock for a function, representing parameters

	// check if there is a "."
	tk3, e := p.scanner.scanToken()
	if e != nil {
		return nil, e
	}

	var chained *chunk

	if tk3.isSym(".") {
		chained, e = p.parseVariableOrFn()
		if e != nil {
			return nil, e
		}
	} else {
		p.scanner.putBack(tk3)
	}

	if params == nil {
		// this is a variable definition
		return newChunkExprVar(tk.value, chained), nil
	}
	return newChunkExprFunc(tk.value, params, chained), nil
}

// parse a comma-delimited list of expressions, returning the values as a CHUNK_BLOCK
func (p *parser) parseExpressionList(allowEmpty bool) (*chunk, error) {
	var chunks []*chunk

	// if an empty list is allowed, check for the fact it's empty. Peek at the next token, and
	// if it is any valid symbol that may appear after an expression list, just return an empty CHUNK_BLOCK.
	if allowEmpty {
		tk, e := p.scanner.scanToken()
		if e != nil {
			return nil, e
		}
		if tk.isSym(")") {
			return newChunkBlock(chunks), nil
		} else {
			p.scanner.putBack(tk)
		}
	}

	expr, e := p.parseExpr(false)
	if e != nil {
		return nil, e
	}

	chunks = append(chunks, expr)

	for {
		// if the next token is a comma, then we need to parse another expression
		tk, e := p.scanner.scanToken()
		if e != nil {
			return nil, e
		}
		if !tk.isSym(",") {
			// not a comma, so put it back, and break
			p.scanner.putBack(tk)
			break
		}

		expr, e := p.parseExpr(false)
		if e != nil {
			return nil, e
		}

		chunks = append(chunks, expr)
	}

	return newChunkBlock(chunks), nil
}
