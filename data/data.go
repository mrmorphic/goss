package data

import (
	"fmt"
	"github.com/mrmorphic/goss"
	"github.com/mrmorphic/goss/convert"
	"reflect"
)

func Eval(context interface{}, name string, args ...interface{}) interface{} {
	return NewDefaultLocater(context).Get(name, args...)
}

// defaultLocator is an implementation of DataLocator, with specific behaviours that make it useful in
// the context of SilverStripe. Specifically:
//  * given a name and optional args, will locate a function or property that will meet that.
//  * if context is a map with string key, it will return properties
//  * if a matching symbol can't be found in the context, it will use a property or function called
//    _fallback if it exists, and try to interpret that instead. This allows for the SS-type behaviour
//    of passing a controller
//  * if the context value implements DataLocator, it will delegate to that.
type DefaultLocater struct {
	context interface{}
}

func NewDefaultLocater(context interface{}) goss.Evaluater {
	return &DefaultLocater{context}
}

func (d *DefaultLocater) Get(name string, args ...interface{}) interface{} {
	fmt.Printf("Locate %s (%s) in %s\n", name, args, d.context)

	// Get the Value of context, and dereference if the type is a pointer
	ctx := reflect.ValueOf(d.context)
	if ctx.Kind() == reflect.Ptr {
		ctx = ctx.Elem()
	}

	var value reflect.Value

	// Get the Value associated with the name, which depends on what kind of item the
	// context is..
	switch {
	case ctx.Kind() == reflect.Map:
		fmt.Printf("...is a map\n")
		m, ok := d.context.(map[string]interface{})
		if !ok {
			panic("locater: map must be map[string]interface{}")
		}
		value = reflect.ValueOf(m[name])
	case ctx.Kind() == reflect.Struct:
		// test first for a function of that name
		value = ctx.MethodByName(name)
		if IsZeroOfUnderlyingType(value) {
			// if no function, test for struct field of that name. @todo lowercase hidden?
			value = ctx.FieldByName(name)
		}
	}

	// Now we have the value, work out what to do with it. There are two special cases; value couldn't
	// be determined so try _fallback; the value's kind is a function, so call it with args
	switch {
	case IsZeroOfUnderlyingType(value):
		// see if there is a _fallback
		if name != "_fallback" {
			fallback := d.Get("_fallback")
			if fallback != nil {
				return Eval(fallback, name, args...)
			}
		}
		// we couldn't work it out, just return nil with no error.
		return nil
	case value.Kind() == reflect.Func:
		// reflection funkiness; create a slice of args asserted as reflect.Value.
		a := make([]reflect.Value, 0)
		for _, x := range args {
			a = append(a, reflect.ValueOf(x))
		}
		result := value.Call(a)

		// we ignore any other values returned.
		return result[0].Interface()
	}

	return value.Interface()
}

// Return string representation of the field
func (d *DefaultLocater) GetStr(fieldName string, args ...interface{}) string {
	return convert.AsString(d.Get(fieldName))
}

func (d *DefaultLocater) GetInt(fieldName string, args ...interface{}) (int, error) {
	return convert.AsInt(d.Get(fieldName))
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return x != nil && x == reflect.Zero(reflect.TypeOf(x)).Interface()
}
