package template

import (
	"fmt"
)

type chunkKind string

const (
	CHUNK_LITERAL         chunkKind = "literal"  // literal text to emit
	CHUNK_BASE_TAG        chunkKind = "base_tag" // <% base_tag %> substitution
	CHUNK_INCLUDE         chunkKind = "include"  // process an include file
	CHUNK_BLOCK           chunkKind = "block"    // a sequence of sub-chunks
	CHUNK_LOOP            chunkKind = "loop"     // a loop block
	CHUNK_WITH            chunkKind = "with"     // a with block
	CHUNK_IF              chunkKind = "if"
	CHUNK_EXPR_VAR        chunkKind = "expr_var"
	CHUNK_EXPR_FUNC       chunkKind = "expr_func"
	CHUNK_EXPR_NUMBER     chunkKind = "expr_number"
	CHUNK_EXPR_STRING     chunkKind = "expr_string"
	CHUNK_EXPR_NOT        chunkKind = "expr_not"
	CHUNK_EXPR_OR         chunkKind = "expr_or"
	CHUNK_EXPR_AND        chunkKind = "expr_and"
	CHUNK_EXPR_EQUAL      chunkKind = "expr_equal"
	CHUNK_EXPR_NOT_EQUAL  chunkKind = "expr_not_equal"
	CHUNK_EXPR_LESS       chunkKind = "expr_less"
	CHUNK_EXPR_LESS_EQUAL chunkKind = "expr_less_equal"
	CHUNK_EXPR_GTR        chunkKind = "expr_gtr"
	CHUNK_EXPR_GTR_EQUAL  chunkKind = "expr_gtr_equal"
)

type chunk struct {
	kind chunkKind
	m    map[string]interface{}
}

func newChunk(kind chunkKind) *chunk {
	return &chunk{kind: kind, m: make(map[string]interface{})}
}

func newChunkLiteral(literal string) *chunk {
	r := newChunk(CHUNK_LITERAL)
	r.m["content"] = literal
	return r
}

func newChunkBaseTag() *chunk {
	return newChunk(CHUNK_BASE_TAG)
}

func newChunkInclude(c *compiledTemplate) *chunk {
	r := newChunk(CHUNK_INCLUDE)
	r.m["compiled"] = c
	return r
}

func newChunkBlock(chunks []*chunk) *chunk {
	r := newChunk(CHUNK_BLOCK)
	r.m["chunks"] = chunks
	return r
}

func newChunkIf(condition *chunk, truePart *chunk, falsePart *chunk) *chunk {
	r := newChunk(CHUNK_IF)
	r.m["condition"] = condition
	r.m["truePart"] = truePart
	r.m["falsePart"] = falsePart
	return r
}
func newChunkLoop(context *chunk, body *chunk) *chunk {
	r := newChunk(CHUNK_LOOP)
	r.m["context"] = context
	r.m["body"] = body
	return r
}

func newChunkWith(context *chunk, body *chunk) *chunk {
	r := newChunk(CHUNK_WITH)
	r.m["context"] = context
	r.m["body"] = body
	return r
}

func newChunkExprVar(name string, chained *chunk) *chunk {
	r := newChunk(CHUNK_EXPR_VAR)
	r.m["name"] = name
	r.m["chained"] = chained
	return r
}

// Create a new chunk representing a function call. 'name' is the name of the function. 'params' is a CHUNK_BLOCK
// containing a list of expression chunks. 'chained' is a single chunk to evaluate once we've evaluated this one.
func newChunkExprFunc(name string, params *chunk, chained *chunk) *chunk {
	r := newChunk(CHUNK_EXPR_FUNC)
	r.m["name"] = name
	r.m["params"] = params
	r.m["chained"] = chained
	return r
}

// Can be used for a variety of chunk kinds where there is only a single value.
func newChunkExprValue(kind chunkKind, value interface{}) *chunk {
	r := newChunk(kind)
	r.m["value"] = value
	return r
}

func newChunkExprNary(kind chunkKind, args []*chunk) *chunk {
	r := newChunk(kind)
	r.m["args"] = args
	return r
}

func (c *chunk) printable(nestLevel int) string {
	result := ""
	s := "                                "[0:nestLevel]
	result += fmt.Sprintf("%s%s:\n", s, c.kind)
	for key, value := range c.m {
		result += s + " " + key + ":"

		nested, ok := value.(*chunk)
		if ok {
			// the value is a *chunk, so we can nest on it as long as it's not null (which is valid for some cases)
			if nested == nil {
				result += "nil"
			} else {
				result += "\n " + nested.printable(nestLevel+1)
			}
		} else {
			list, ok := value.([]*chunk)
			if ok && list != nil {
				// it's a list of chunks
				result += "\n" + s + " ["
				for _, item := range list {
					result += "\n " + item.printable(nestLevel+1)
				}
				result += "\n" + s + " ]\n"
			} else {
				// could be anything else, just use Sprintf.
				result += fmt.Sprintf(" %s%s", s, value)
			}
		}
	}
	return result
}
