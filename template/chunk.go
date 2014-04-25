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
	CHUNK_REQUIRE         chunkKind = "require"
	CHUNK_LAYOUT          chunkKind = "layout"
	CHUNK_EXPR_VARFUNC    chunkKind = "expr_varfunc"
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

// chunk is the primary structure of compiled templates. Chunks are nested, so form a hierarchy. The parser
// generates this tree, and the executer will process it given a specific context, and render the results.
// There several kinds of chunks, as indicated by the chunkKind. 'm' is a map of values stored in the chunk;
// each kind of chunk has specific keys in 'm'. Chunks are created by the newChunk* methods, which determine
// the values set based on what that kind of chunk requires.
// Further, chunks are either for rendering, or for expression evaluation. CHUNK_EXPR_* constants are the
// expression chunks. The only exception is CHUNK_BLOCK which represents a sequence of rendering chunks, or a
// sequence of expression chunks (e.g. parameters to a function call.)
type chunk struct {
	kind chunkKind
	m    map[string]interface{}

	// populated by the parser during construction, these are used for error reporting when actually executing chunk
	file string
	line int
}

// Create new generic chunk of a kind, but no map values.
func newChunk(kind chunkKind, reporter errorLocationReporter) *chunk {
	file, line := reporter.errorLocation()
	return &chunk{kind: kind, m: make(map[string]interface{}), file: file, line: line}
}

// Create a chunk representing a literal piece of markup to be output without further processing.
func newChunkLiteral(literal string, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_LITERAL, reporter)
	r.m["content"] = literal
	return r
}

// Create a chunk for '<% base_tag %>' rendering.
func newChunkBaseTag(reporter errorLocationReporter) *chunk {
	return newChunk(CHUNK_BASE_TAG, reporter)
}

// Create a chunk for '<% include ... %>' rendering.
func newChunkInclude(c *compiledTemplate, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_INCLUDE, reporter)
	r.m["compiled"] = c
	return r
}

// Create a block chunk, which is a linear sequence of sub-chunks.
func newChunkBlock(chunks []*chunk, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_BLOCK, reporter)
	r.m["chunks"] = chunks
	return r
}

// Create a chunk for '<% require ... %>' rendering
func newChunkRequire(rType string, path string, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_REQUIRE, reporter)
	r.m["type"] = rType
	r.m["path"] = path
	return r
}

// Create a chunk for '<% if ... %>...<% end_if %>' rendering
func newChunkIf(condition *chunk, truePart *chunk, falsePart *chunk, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_IF, reporter)
	r.m["condition"] = condition
	r.m["truePart"] = truePart
	r.m["falsePart"] = falsePart
	return r
}

// Create a chunk for '<% loop ... %>...<% end_loop %>' rendering.
func newChunkLoop(context *chunk, body *chunk, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_LOOP, reporter)
	r.m["context"] = context
	r.m["body"] = body
	return r
}

// Create a chunk for '<% with ... %>...<% end_with %>' rendering.
func newChunkWith(context *chunk, body *chunk, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_WITH, reporter)
	r.m["context"] = context
	r.m["body"] = body
	return r
}

// Create a new chunk representing a function call. 'name' is the name of the function. 'params' is a CHUNK_BLOCK
// containing a list of expression chunks. 'chained' is a single chunk to evaluate once we've evaluated this one.
func newChunkExprVarFunc(name string, params *chunk, chained *chunk, reporter errorLocationReporter) *chunk {
	r := newChunk(CHUNK_EXPR_VARFUNC, reporter)
	r.m["name"] = name
	r.m["params"] = params
	r.m["chained"] = chained
	return r
}

// Can be used for a variety of chunk kinds where there is only a single value.
func newChunkExprValue(kind chunkKind, value interface{}, reporter errorLocationReporter) *chunk {
	r := newChunk(kind, reporter)
	r.m["value"] = value
	return r
}

// Create a chunk for an expression chunk that has a number of arguments.
func newChunkExprNary(kind chunkKind, args []*chunk, reporter errorLocationReporter) *chunk {
	r := newChunk(kind, reporter)
	r.m["args"] = args
	return r
}

// generate a printable version of a chunk for debugging.
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

// implement errorLocationReporter
func (c *chunk) errorLocation() (string, int) {
	return c.file, c.line
}
