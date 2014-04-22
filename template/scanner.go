package template

import (
	"errors"
	"strings"
)

type scanner struct {
	// initialised with the source to compile, this gets reduces as the scanner processes it.
	source string

	// tells the scanner if we are in a template tag or not, which effects how we scan
	inTemplateTag bool

	// a stack of tokens that have been scanned but put back. The 0th item is the bottom of the stack.
	// When scanToken is called, if this is not empty the token at the top of the stack (end of list) is
	// returned.
	unprocessedStack []*token
}

func newScanner(s string) *scanner {
	return &scanner{source: s, unprocessedStack: make([]*token, 0)}
}

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
)

// Given a string s, scan 1 token and return it, and the reduced string.
func (sc *scanner) scanToken() (*token, error) {
	if ls := len(sc.unprocessedStack); ls > 0 {
		// if there are tokens that were put back. Return the last added
		result := sc.unprocessedStack[ls-1]
		sc.unprocessedStack = sc.unprocessedStack[0 : ls-1]
		return result, nil
	}

	for {
		// end of input
		if len(sc.source) == 0 {
			return newToken(TOKEN_END_SOURCE, ""), nil
		}

		var ls = len(sc.source)

		if sc.inTemplateTag {
			switch {
			case ls >= 2 && sc.source[0:2] == "%>":
				sc.inTemplateTag = false
				sc.source = sc.source[2:]
				return newToken(TOKEN_CLOSE, ""), nil
			case sc.source[0] == ' ':
				// space; ignore while in template
				sc.source = sc.source[1:]
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
				return newToken(TOKEN_IDENT, ident), nil

			case strings.Contains(digits, sc.source[0:1]):
				return sc.scanNumericLiteral()
			case sc.source[0] == '"':
				return sc.scanStringLiteral()

			default:
				if ls >= 2 {
					// look for 2-char operators
					t := sc.source[0:2]
					if t == "==" || t == "!=" || t == ">=" || t == "<=" || t == "&&" || t == "||" {
						sc.source = sc.source[2:]
						return newToken(TOKEN_SYMBOL, t), nil
					}
				}
				t := sc.source[0:1]
				sc.source = sc.source[1:]
				return newToken(TOKEN_SYMBOL, t), nil
			}
		} else {
			switch {
			case ls >= 2 && sc.source[0:2] == "<%":
				sc.inTemplateTag = true
				sc.source = sc.source[2:]
				return newToken(TOKEN_OPEN, ""), nil
			case ls >= 2 && sc.source[0] == '$' && sc.source[0:2] != "$$":
				// start of identifier in token
				sc.inTemplateTag = true
				sc.source = sc.source[1:]
				return newToken(TOKEN_SYMBOL, "$"), nil
			case ls >= 2 && sc.source[0:2] == "{$":
				// special case {$something}. We return "{" as a symbol here, and the next token will be the symbol $. The parse will have to work it out
				sc.source = sc.source[1:]
				return newToken(TOKEN_SYMBOL, "{"), nil
			default:
				// eat chars until we see the potential start
				lit := ""
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
							lit += string(sc.source[0])
							sc.source = sc.source[1:]
						} else {
							break
						}
					}

					// just copy the char over
					lit += string(sc.source[0])
					sc.source = sc.source[1:]
				}
				return newToken(TOKEN_LITERAL, lit), nil
			}
		}
	}
}

func (sc *scanner) putBack(t *token) {
	sc.unprocessedStack = append(sc.unprocessedStack, t)
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

func (sc *scanner) scanNumericLiteral() (*token, error) {
	num := ""
	for {
		lsr := len(sc.source)
		if lsr == 0 || !strings.Contains(digits, sc.source[0:1]) {
			break
		}
		num += sc.source[0:1]
		sc.source = sc.source[1:]
	}
	return newToken(TOKEN_NUMBER, num), nil
}

// @todo handle backquote within the literal, including \"
func (sc *scanner) scanStringLiteral() (*token, error) {
	// string literal
	str := ""
	sc.source = sc.source[1:]
	for {
		lsr := len(sc.source)
		if lsr == 0 {
			// an unterminated string
			return nil, errors.New("Unterminated string")
		}

		if sc.source[0] == '"' {
			sc.source = sc.source[1:] // drop the trailing quotes
			break
		}

		str += sc.source[0:1]
		sc.source = sc.source[1:]
	}
	return newToken(TOKEN_STRING, str), nil
}
