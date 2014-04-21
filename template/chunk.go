package template

import (
	"fmt"
)

type chunkKind string

const (
	CHUNK_LITERAL     chunkKind = "chunk_literal"  // literal text to emit
	CHUNK_BASE_TAG              = "chunk_base_tag" // <% base_tag %> substitution
	CHUNK_INCLUDE               = "chunk_include"  // process an include file
	CHUNK_BLOCK                 = "chunk_block"    // a sequence of sub-chunks
	CHUNK_LOOP                  = "chunk_loop"     // a loop block
	CHUNK_WITH                  = "chunk_with"     // a with block
	CHUNK_EXPR_VAR              = "chunk_expr_var"
	CHUNK_EXPR_FUNC             = "chunk_expr_func"
	CHUNK_EXPR_NUMBER           = "chunk_expr_number"
	CHUNK_EXPR_STRING           = "chunk_expr_string"
	CHUNK_EXPR_NOT              = "chunk_expr_not"
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

func (c *chunk) printable(nestLevel int) string {
	result := ""
	s := "                                "[0:nestLevel]
	result += fmt.Sprintf("%s%s:\n", s, c.kind)
	for key, value := range c.m {
		result += s + " " + key + ":"
		nested, ok := value.(*chunk)
		if ok {
			result += nested.printable(nestLevel + 1)
		} else {
			result += fmt.Sprintf(" %s%s", s, value)
		}
	}
	return result
}
