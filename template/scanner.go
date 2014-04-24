package template

import (
	"errors"
	"strings"
)

type tokenKind string

const (
	TOKEN_END_SOURCE tokenKind = "{end source}" // end of stream
	TOKEN_LITERAL              = "literal"      // literal value that can be emitted as-is
	TOKEN_SYMBOL               = "symbol"       // a symbol of some kind. Like a literal, except it is something that has meaning to the template
	TOKEN_IDENT                = "ident"        // identifier, sequence of letters, digits and _, starting with letter or _
	TOKEN_NUMBER               = "number"       // sequence of digits
	TOKEN_STRING               = "string"       // string literal, value excludes the double quotes, and \ chars are processed
)

type token struct {
	kind  tokenKind
	value string

	// this contains the characters scanned for the token. putBack will put this back on the start
	// of the input stream. This is required because when we parse $x in a template, we have to scan
	// ahead to detect "." or "(", but these are interpreted with inTemplateTag true. Anything else that
	// might follow the name should be interpreted as a literal instead, so we need to re-interpret.
	restoreSource string

	// restore state at the point at which we started scanning this token
	restoreInTemplateTag bool
}

func newToken(kind tokenKind, s string, lit string, r bool) *token {
	return &token{kind: kind, value: s, restoreSource: lit, restoreInTemplateTag: r}
}

// isSym returns true if the token is a symbol that is the same as s
func (t *token) isSym(s string) bool {
	return t.kind == TOKEN_SYMBOL && t.value == s
}

func (t *token) isIdent(s string) bool {
	return t.kind == TOKEN_IDENT && t.value == s
}

func (t *token) printable() string {
	result := string(t.kind)
	if t.value != "" {
		result += " (" + t.value + ")"
	}
	result += " [" + t.restoreSource + "]"
	return result
}

type scanner struct {
	// initialised with the source to compile, this gets reduces as the scanner processes it.
	source string

	// tells the scanner if we are in a template tag or not, which effects how we scan
	inTemplateTag bool
}

func newScanner(s string) *scanner {
	return &scanner{source: s}
}

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
)

// Given a string s, scan 1 token and return it, and the reduced string.
func (sc *scanner) scanToken() (*token, error) {
	lit := ""
	r := sc.inTemplateTag
	for {
		// end of input
		if len(sc.source) == 0 {
			return newToken(TOKEN_END_SOURCE, "", "", r), nil
		}

		var ls = len(sc.source)

		if sc.inTemplateTag {
			switch {
			case ls >= 2 && sc.source[0:2] == "%>":
				sc.inTemplateTag = false
				sc.source = sc.source[2:]
				lit += "%>"
				return newToken(TOKEN_SYMBOL, "%>", lit, r), nil
			case sc.source[0] == ' ':
				// space; ignore while in template
				sc.source = sc.source[1:]
				lit += " "
			case strings.Contains(letters+"_", sc.source[0:1]):
				// identifier
				ident := ""
				for {
					lsr := len(sc.source)
					if lsr == 0 {
						break
					}
					if !strings.Contains(letters+"_", sc.source[0:1]) && !strings.Contains(digits, sc.source[0:1]) {
						break
					}
					ident += sc.source[0:1]
					sc.source = sc.source[1:]
				}
				lit += ident
				return newToken(TOKEN_IDENT, ident, lit, r), nil

			case strings.Contains(digits, sc.source[0:1]):
				return sc.scanNumericLiteral(lit, r)
			case sc.source[0] == '"' || sc.source[0] == '\'':
				return sc.scanStringLiteral(lit, r)
			default:
				if ls >= 2 {
					// look for 2-char operators
					t := sc.source[0:2]
					if t == "==" || t == "!=" || t == ">=" || t == "<=" || t == "&&" || t == "||" {
						sc.source = sc.source[2:]
						lit += t
						return newToken(TOKEN_SYMBOL, t, lit, r), nil
					}
				}
				t := sc.source[0:1]
				sc.source = sc.source[1:]
				lit += t
				return newToken(TOKEN_SYMBOL, t, lit, r), nil
			}
		} else {
			switch {
			case ls >= 4 && sc.source[0:4] == "<%--":
				// comment. scan until we find --%>
				lit += "<%--"
				sc.source = sc.source[4:]
				for {
					lsr := len(sc.source)
					if lsr < 4 {
						return nil, errors.New("Unterminated <%-- comment")
					}
					if sc.source[0:4] == "--%>" {
						lit += "--%>"
						sc.source = sc.source[4:]
						break
					}

					lit += sc.source[0:1]
					sc.source = sc.source[1:]
				}
			case ls >= 2 && sc.source[0:2] == "<%":
				sc.inTemplateTag = true
				sc.source = sc.source[2:]
				lit += "<%"
				return newToken(TOKEN_SYMBOL, "<%", lit, r), nil
			case ls >= 2 && sc.source[0] == '$' && sc.source[0:2] != "$$":
				// start of identifier in token
				sc.inTemplateTag = true
				sc.source = sc.source[1:]
				lit += "$"
				return newToken(TOKEN_SYMBOL, "$", lit, r), nil
			case ls >= 2 && sc.source[0:2] == "{$":
				// special case {$something}. We return "{" as a symbol here, and the next token will be the symbol $. The parse will have to work it out
				sc.source = sc.source[1:]
				lit += "{"
				return newToken(TOKEN_SYMBOL, "{", lit, r), nil
			default:
				// eat chars until we see the potential start
				literal := ""
				for {
					lsr := len(sc.source)
					if lsr == 0 {
						break
					}
					if lsr >= 2 && (sc.source[0:2] == "<%" || sc.source[0:2] == "{$") {
						break
					}

					if lsr >= 1 && sc.source[0] == '$' {
						// if it's $$, copy the first $ over, and when we break we'll copy the second $.
						if lsr >= 2 && sc.source[0:2] == "$$" {
							literal += string(sc.source[0])
							lit += "$$"
							sc.source = sc.source[1:]
						} else {
							break
						}
					}

					// just copy the char over
					literal += string(sc.source[0])
					lit += string(sc.source[0])
					sc.source = sc.source[1:]
				}
				return newToken(TOKEN_LITERAL, literal, lit, r), nil
			}
		}
	}
}

func (sc *scanner) putBack(t *token) {
	sc.source = t.restoreSource + sc.source
	sc.inTemplateTag = t.restoreInTemplateTag
}

// Get the next token from the stream, put it back and return it. This will leave the input string scanned, but the scanned
// token in the queue.
func (sc *scanner) peek() (*token, error) {
	t, e := sc.scanToken()
	if e != nil {
		return nil, e
	}
	sc.putBack(t)
	return t, nil
}

func (sc *scanner) scanNumericLiteral(lit string, r bool) (*token, error) {
	num := ""
	for {
		lsr := len(sc.source)
		if lsr == 0 || !strings.Contains(digits, sc.source[0:1]) {
			break
		}
		num += sc.source[0:1]
		sc.source = sc.source[1:]
	}
	lit += num
	return newToken(TOKEN_NUMBER, num, lit, r), nil
}

// @todo handle backquote within the literal, including \"
func (sc *scanner) scanStringLiteral(lit string, r bool) (*token, error) {
	// string literal
	str := ""
	delimiter := sc.source[0:1]
	sc.source = sc.source[1:]
	lit += delimiter
	for {
		lsr := len(sc.source)
		if lsr == 0 {
			// an unterminated string
			return nil, errors.New("Unterminated string")
		}

		if sc.source[0:1] == delimiter {
			sc.source = sc.source[1:] // drop the trailing quotes
			lit += delimiter
			break
		}

		str += sc.source[0:1]
		lit += sc.source[0:1]
		sc.source = sc.source[1:]
	}
	return newToken(TOKEN_STRING, str, lit, r), nil
}
