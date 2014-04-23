package template

import (
	"errors"
	"fmt"
	//	"reflect"
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
	fmt.Printf("rendering block: %s\n", ch.printable(0))
	result := []byte{}
	for _, nested := range ch.m["chunks"].([]*chunk) {
		b, e := exec.renderChunk(nested)
		if e != nil {
			return nil, e
		}
		result = append(result, b...)
	}
	fmt.Printf("... result is %s\n", result)
	return result, nil
}

func (exec *executer) renderChunkVarFunc(ch *chunk) ([]byte, error) {
	v, e := exec.eval(ch)
	if e != nil {
		return nil, e
	}
	return []byte(fmt.Sprintf("%s", v)), nil
}

func (exec *executer) renderChunkIf(ch *chunk) ([]byte, error) {
	condition := ch.m["condition"]
	truePart := ch.m["truePart"]
	falsePart := ch.m["falsePart"]

	cond, e := exec.eval(condition.(*chunk))
	if e != nil {
		return nil, e
	}

	b, ok := cond.(bool)
	if !ok {
		return nil, fmt.Errorf("If condition must be boolean")
	}

	render := truePart
	if !b {
		render = falsePart
	}

	if render.(*chunk) == nil {
		// there is nothng to render
		return []byte{}, nil
	}
	return exec.renderChunk(render.(*chunk))
}

// evalBlock evaluates a list of expressions in a block, which themselves are *chunk values. These
// are evaluated in the current context. They are returned as a slice of values. It will return an error
// if anything goes wrong.
func (exec *executer) evalBlock(block *chunk) ([]interface{}, error) {
	if block == nil {
		return nil, nil
	}
	chunks := block.m["chunks"].([]*chunk)
	result := make([]interface{}, 0)

	for _, c := range chunks {
		v, e := exec.eval(c)
		if e != nil {
			return nil, e
		}
		result = append(result, v)
	}

	return result, nil
}

// evaluate an expr chunk. Will return an error if its not an expr chunk.
func (exec *executer) eval(expr *chunk) (interface{}, error) {
	switch expr.kind {
	case CHUNK_BLOCK:
	// case CHUNK_LOOP:
	// case CHUNK_WITH:
	// case CHUNK_IF:
	case CHUNK_EXPR_VARFUNC:
		return exec.evalVarFunc(expr)
	case CHUNK_EXPR_NUMBER, CHUNK_EXPR_STRING:
		return expr.m["value"], nil
	case CHUNK_EXPR_NOT:
	case CHUNK_EXPR_OR:
		return exec.evalBoolFuncN(expr, func(p1 bool, p2 bool) (bool, bool) {
			return p1 || p2, !(p1 || p2)
		})
	case CHUNK_EXPR_AND:
		return exec.evalBoolFuncN(expr, func(p1 bool, p2 bool) (bool, bool) {
			return p1 && p2, !(p1 && p2)
		})
	case CHUNK_EXPR_EQUAL,
		CHUNK_EXPR_NOT_EQUAL,
		CHUNK_EXPR_LESS,
		CHUNK_EXPR_LESS_EQUAL,
		CHUNK_EXPR_GTR,
		CHUNK_EXPR_GTR_EQUAL:
		return exec.evalCompare(expr)
	}
	return nil, fmt.Errorf("Cannot evaluate a non-expression chunk: %s", expr.kind)
}

func (exec *executer) evalVarFunc(expr *chunk) (interface{}, error) {
	name := expr.m["name"].(string)
	fmt.Printf("... evaluating var/function %s\n", name)
	chained := expr.m["chained"].(*chunk)
	params := expr.m["params"].(*chunk) // block of further chunks
	if params != nil {
		fmt.Printf("exec sees params %s\n", params.printable(0))
	}
	paramList, e := exec.evalBlock(params)
	if e != nil {
		return nil, e
	}

	value, e := exec.locator.Locate(exec.context(), name, paramList)
	if e != nil {
		return nil, e
	}
	fmt.Printf("... locator said: %s\n", value)
	if chained == nil {
		return value, e
	} else {
		exec.push(value)
		v, e := exec.eval(chained)
		exec.pop()
		if e != nil {
			return nil, e
		}
		return v, nil
	}
}

// evaluate a boolean function expr with 2 or more arguments.
// process args in order until we exhaust the list or are told to stop. We don't evaluate args until we need
// to, so that operators || and && can be short-circuited.
func (exec *executer) evalBoolFuncN(expr *chunk, fn func(bool, bool) (result bool, stop bool)) (interface{}, error) {
	args := expr.m["args"].([]*chunk)

	arg1 := args[0]
	args = args[1:]

	// convert arg1 to a boolean
	v, e := exec.eval(arg1)
	if e != nil {
		return nil, e
	}

	result := v.(bool)
	var stop bool

	// compare arg1 to the first item in args, reducing args each time
	for {
		next := args[0]

		v, e := exec.eval(next)
		if e != nil {
			return nil, e
		}

		b := v.(bool)

		result, stop = fn(result, b)

		args = args[1:]

		if stop || len(args) == 0 {
			break
		}
	}

	return result, nil
}

// evaluate a comparison on two values.
func (exec *executer) evalCompare(expr *chunk) (interface{}, error) {
	args := expr.m["args"].([]*chunk)

	// evaluate the args into arg1 and arg2, both interface{}
	arg1, e := exec.eval(args[0])
	if e != nil {
		return nil, e
	}

	arg2, e := exec.eval(args[1])
	if e != nil {
		return nil, e
	}

	// if both are int, compare as int
	if exec.isInt(arg1) && exec.isInt(arg2) {
		return exec.evalCompareInt(expr, arg1, arg2)
	}

	// if both are float, compare as float
	if exec.isFloat(arg1) && exec.isFloat(arg2) {
		return exec.evalCompareFloat(expr, arg1, arg2)
	}

	// anything is converted to a string first, and then we compare
	s1 := fmt.Sprintf("%s", arg1)
	s2 := fmt.Sprintf("%s", arg2)

	switch expr.kind {
	case CHUNK_EXPR_EQUAL:
		return s1 == s2, nil
	case CHUNK_EXPR_NOT_EQUAL:
		return s1 != s2, nil
	case CHUNK_EXPR_LESS:
		return s1 < s2, nil
	case CHUNK_EXPR_LESS_EQUAL:
		return s1 <= s2, nil
	case CHUNK_EXPR_GTR:
		return s1 > s2, nil
	case CHUNK_EXPR_GTR_EQUAL:
		return s1 >= s2, nil
	}

	return nil, fmt.Errorf("Invalid comparison operator: %s\n", expr.kind)
}

func (exec *executer) isInt(v interface{}) bool {
	_, ok := v.(int)
	return ok
}

func (exec *executer) isFloat(v interface{}) bool {
	_, ok := v.(float32)
	return ok
}

func (exec *executer) evalCompareInt(expr *chunk, p1 interface{}, p2 interface{}) (interface{}, error) {
	v1 := p1.(int)
	v2 := p2.(int)

	switch expr.kind {
	case CHUNK_EXPR_EQUAL:
		return v1 == v2, nil
	case CHUNK_EXPR_NOT_EQUAL:
		return v1 != v2, nil
	case CHUNK_EXPR_LESS:
		return v1 < v2, nil
	case CHUNK_EXPR_LESS_EQUAL:
		return v1 <= v2, nil
	case CHUNK_EXPR_GTR:
		return v1 > v2, nil
	case CHUNK_EXPR_GTR_EQUAL:
		return v1 >= v2, nil
	}
	return false, nil
}

func (exec *executer) evalCompareFloat(expr *chunk, p1 interface{}, p2 interface{}) (interface{}, error) {
	v1 := p1.(float32)
	v2 := p2.(float32)

	switch expr.kind {
	case CHUNK_EXPR_EQUAL:
		return v1 == v2, nil
	case CHUNK_EXPR_NOT_EQUAL:
		return v1 != v2, nil
	case CHUNK_EXPR_LESS:
		return v1 < v2, nil
	case CHUNK_EXPR_LESS_EQUAL:
		return v1 <= v2, nil
	case CHUNK_EXPR_GTR:
		return v1 > v2, nil
	case CHUNK_EXPR_GTR_EQUAL:
		return v1 >= v2, nil
	}
	return false, nil
}
