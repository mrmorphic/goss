package template

import (
	"errors"
	"fmt"
	"github.com/mrmorphic/goss"
	"reflect"
)

// executer will render a template using a context, data locater and requirements provider. It returns a rendered
// template.
type executer struct {
	// context is handled by a stack. The bottom of the stack is index 0. Expressions are evaluated in the context
	// of the top of the stack.
	contextStack []interface{}

	// All the templates. If there is 1, it is a main template. If there are 2, the first is the main template, and the second
	// is the layout template, used to substitute $Layout.
	templates []*compiledTemplate

	// The locator for symbol lookups
	locator DataLocator

	// interface for handling requirements processing
	require goss.RequirementsProvider
}

func newExecuter(templates []*compiledTemplate, context interface{}, locator DataLocator, require goss.RequirementsProvider) *executer {
	exec := &executer{contextStack: make([]interface{}, 0), templates: templates, locator: locator, require: require}
	exec.push(context)
	return exec
}

// push a context onto the context stack
func (exec *executer) push(context interface{}) {
	exec.contextStack = append(exec.contextStack, context)
}

// pop a context off the context stack. It will return an error if the context stack is empty.
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

// Invokes renderChunk to render the template, and then invokes the requirements provider
// to inject the correct bits into the output.
func (exec *executer) render() ([]byte, error) {
	// Render the main template
	bytes, e := exec.renderChunk(exec.templates[0].chunk)
	if e != nil {
		return nil, e
	}

	fmt.Printf("exec.render: output is %s\n", bytes)
	// insert the header tags
	bytes, e = exec.require.InsertHeadTags(bytes)
	if e != nil {
		return nil, e
	}

	// insert the body tags
	return exec.require.InsertBodyTags(bytes)
}

// given a chunk, render it using the current context
func (exec *executer) renderChunk(chunk *chunk) ([]byte, error) {
	switch chunk.kind {
	case CHUNK_LITERAL:
		return []byte(chunk.m["content"].(string)), nil
	case CHUNK_BASE_TAG:
		return exec.renderBaseTag(chunk)
	case CHUNK_INCLUDE:
		return exec.renderInclude(chunk)
	case CHUNK_BLOCK:
		return exec.renderChunkBlock(chunk)
	case CHUNK_LOOP:
		return exec.renderChunkLoop(chunk)
	case CHUNK_WITH:
		return exec.renderChunkWith(chunk)
	case CHUNK_IF:
		return exec.renderChunkIf(chunk)
	case CHUNK_LAYOUT:
		return exec.renderChunkLayout(chunk)
	case CHUNK_REQUIRE:
		return exec.renderRequire(chunk)
	case CHUNK_EXPR_VARFUNC:
		return exec.renderChunkVarFunc(chunk)
	}
	return nil, newTemplateError(fmt.Sprintf("could not render chunk of unknown kind '%s'", chunk.kind), chunk)
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
	condition := ch.m["condition"].(*chunk)
	truePart := ch.m["truePart"]
	falsePart := ch.m["falsePart"]

	cond, e := exec.eval(condition)
	if e != nil {
		return nil, e
	}
	fmt.Printf("cond: %s\n", cond)
	b, e := exec.boolOf(cond)
	if e != nil {
		return nil, newTemplateError(fmt.Sprintf("If condition must be boolean '%s'", condition.printable(0)), ch)
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

// Given a value, try and interpret the value as a boolean. If we can't interpret it, return an error.
func (exec *executer) boolOf(value interface{}) (bool, error) {
	// nil value is false. e.g. <% if $name %>, where name is not defined in the context
	if value == nil {
		return false, nil
	}
	fmt.Printf("boolOf %s\n", value)

	switch value := value.(type) {
	case bool, *bool:
		return value.(bool), nil
	case int, *int:
		return value != 0, nil
	case string, *string:
		return value != "", nil
	}

	// if value is a function, test that it returns boolean

	return false, fmt.Errorf("Cannot treat %s as bool", value)
}

// renderChunkLayout handles injection of $Layout in a main template.
func (exec *executer) renderChunkLayout(ch *chunk) ([]byte, error) {
	if len(exec.templates) < 2 {
		// there is no layout template. Treat it like a non-existent variable and render nothing.
		return []byte{}, nil
	}

	// Just render the layout template
	return exec.renderChunk(exec.templates[1].chunk)
}

func (exec *executer) renderBaseTag(ch *chunk) ([]byte, error) {
	url := configuration.siteUrl
	if url == "" {
		url = "localhost"
	}
	if url[0:7] != "http://" && url[0:8] != "https://" {
		url = configuration.defaultProtocol + "://" + url
	}
	tag := `<base href="` + url + ` /><!--[if lte IE 6]></base><![endif]-->`
	return []byte(tag), nil
}

func (exec *executer) renderChunkWith(ch *chunk) ([]byte, error) {
	ctxChunk := ch.m["context"].(*chunk)
	bodyChunk := ch.m["body"].(*chunk)

	ctx, e := exec.eval(ctxChunk)
	if e != nil {
		return nil, e
	}

	// push this new context
	exec.push(ctx)

	bytes, e := exec.renderChunk(bodyChunk)
	if e != nil {
		return nil, e
	}

	_, e = exec.pop()
	if e != nil {
		return nil, e
	}

	return bytes, nil
}

func (exec *executer) renderChunkLoop(ch *chunk) ([]byte, error) {
	ctxChunk := ch.m["context"].(*chunk)
	bodyChunk := ch.m["body"].(*chunk)

	ctxIntf, e := exec.eval(ctxChunk)
	if e != nil {
		return nil, e
	}

	result := []byte{}

	// we expect ctx to be a slice
	ctxV := reflect.ValueOf(ctxIntf)
	// @todo need to handle arrays as well?
	if ctxV.Kind() != reflect.Slice && ctxV.Kind() != reflect.Array {
		return nil, newTemplateError("loop context must be a slice", ch)
	}

	for i := 0; i < ctxV.Len(); i++ {
		// get the i-th element
		el := ctxV.Index(i)

		// make this element the context for the loop iteration
		exec.push(el.Interface())

		// render the chunk with the new context
		bytes, e := exec.renderChunk(bodyChunk)
		if e != nil {
			return nil, e
		}
		result = append(result, bytes...)

		_, e = exec.pop()
		if e != nil {
			return nil, e
		}
	}

	return result, nil
}

// renderRequire is something of a special case. It does not render inline, so returns an empty slice.
// But it tells the requirements interface to include a new file.
func (exec *executer) renderRequire(ch *chunk) ([]byte, error) {
	rtype := ch.m["type"].(string)
	path := ch.m["path"].(string)

	switch rtype {
	case "themedCSS":
		// because we only have one theme, we can calculate the path. Also, when using themedCSS the .css
		// is not put in.
		path = configuration.cssURL + path + ".css"
		exec.require.CSS(path)
	case "css":
		exec.require.CSS(path)
	case "javascript":
		exec.require.Javascript(path)
	}
	return []byte{}, nil
}

func (exec *executer) renderInclude(ch *chunk) ([]byte, error) {
	compiled := ch.m["compiled"].(*compiledTemplate)

	return exec.renderChunk(compiled.chunk)
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
	return nil, newTemplateError(fmt.Sprintf("Cannot evaluate a non-expression chunk: %s", expr.kind), expr)
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

	return nil, newTemplateError(fmt.Sprintf("Invalid comparison operator: %s\n", expr.kind), expr)
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
