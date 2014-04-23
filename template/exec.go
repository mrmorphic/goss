package template

import (
	"errors"
	"fmt"
)

type executer struct {
	contextStack []interface{}
	locator      DataLocator
}

func newExecuter(context interface{}) *executer {
	exec := &executer{contextStack: make([]interface{}, 0), locator: newDefaultLocator()}
	exec.push(context)
	return exec
}

func (exec *executer) push(context interface{}) {
	exec.contextStack = append(exec.contextStack, context)
}

func (exec *executer) pop() (interface{}, error) {
	numItems := len(exec.contextStack)
	if numItems == 0 {
		return nil, errors.New("template executor context stack is empty, but a value has been requested")
	}
	r := exec.context()                                   // get what's on top of stack
	exec.contextStack = exec.contextStack[0 : numItems-1] // drop item off stack

	return r, nil
}

// return the top of the context stack. panics if there is nothing there.
func (exec *executer) context() interface{} {
	n := len(exec.contextStack)
	if n == 0 {
		panic("template stack is empty. cannot get current context")
	}
	return exec.contextStack[n-1]
}

// given a chunk, render it using the current context
func (exec *executer) renderChunk(chunk *chunk) ([]byte, error) {
	switch chunk.kind {
	case CHUNK_LITERAL:
		return []byte(chunk.m["content"].(string)), nil
	case CHUNK_BASE_TAG:
	case CHUNK_INCLUDE:
	case CHUNK_BLOCK:
		return exec.renderChunkBlock(chunk)
	case CHUNK_LOOP:
	case CHUNK_WITH:
	case CHUNK_IF:
		return exec.renderChunkIf(chunk)
	case CHUNK_EXPR_VARFUNC:
		return exec.renderChunkVarFunc(chunk)
	case CHUNK_EXPR_NUMBER:
	case CHUNK_EXPR_STRING:
	case CHUNK_EXPR_NOT:
	case CHUNK_EXPR_OR:
	case CHUNK_EXPR_AND:
	case CHUNK_EXPR_EQUAL:
	case CHUNK_EXPR_NOT_EQUAL:
	case CHUNK_EXPR_LESS:
	case CHUNK_EXPR_LESS_EQUAL:
	case CHUNK_EXPR_GTR:
	case CHUNK_EXPR_GTR_EQUAL:
	}
	return nil, fmt.Errorf("could not render chunk of unknown kind '%s'", chunk.kind)
}

func (exec *executer) renderChunkBlock(ch *chunk) ([]byte, error) {
	result := []byte{}
	for _, nested := range ch.m["chunks"].([]*chunk) {
		b, e := exec.renderChunk(nested)
		if e != nil {
			return nil, e
		}
		result = append(result, b...)
	}
	return result, nil
}

func (exec *executer) renderChunkVarFunc(ch *chunk) ([]byte, error) {
	v, e := exec.eval(ch)
	if e != nil {
		return nil, e
	}
	return []byte(fmt.Sprintf("%s", v)), nil
}

	value, e := exec.locator.Locate(exec.context(), name, nil)
	if e != nil {
		return nil, e
	}

	if chained == nil {
		return []byte(fmt.Sprintf("%s", value)), nil
	} else {
		exec.push(value)
		v, e := exec.renderChunk(chained)
		exec.pop()
		if e != nil {
			return nil, e
		}
		return v, nil
	}
}
