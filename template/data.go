package template

import (
	"fmt"
	"github.com/mrmorphic/goss/orm"
	"reflect"
)

// defaultLocator is an implementation of DataLocator, with specific behaviours that make it useful in
// the context of SilverStripe. Specifically:
//  * given a name and optional args, will locate a function or property that will meet that.
//  * if context is a map with string key, it will return properties
//  * if a matching symbol can't be found in the context, it will use a property or function called
//    _fallback if it exists, and try to interpret that instead. This allows for the SS-type behaviour
//    of passing a controller
//  * if the context value implements DataLocator, it will delegate to that.
type DefaultLocator struct {
}

func NewDefaultLocator() DataLocator {
	return &DefaultLocator{}
}

func (d *DefaultLocator) Locate(context interface{}, name string, args []interface{}) (interface{}, error) {
	fmt.Printf("Locate %s (%s) in %s\n", name, args, context)
	ctx := reflect.ValueOf(context)
	ctxElem := ctx
	var value interface{}

	if ctx.Kind() == reflect.Ptr {
		// dereference before the switch
		ctxElem = ctx.Elem()
	}

	typ := ctx.Type().Name()
	fmt.Printf("type is %s\n", typ)
	// interpret the context based on what kind of object we're passed. The intent is to look up the name in the context,
	// and populate 'value'.
	switch {
	// context implements DataLocator
	case ctx.Kind() == reflect.Map:
		fmt.Printf("...is a map\n")
		// @todo this is too restrictive. What we really want is to ensure that context is a map with a string key, and any type.
		m, ok := context.(map[string]interface{})
		if !ok {
			panic("locater: map must be map[string]interface{}")
		}
		value = m[name]
	case ctxElem.Kind() == reflect.Struct && typ == "DataObject":
		fmt.Printf("Locate found DataObject\n")
		value = ctx.Interface().(*orm.DataObject).FieldByName(name)
	case ctx.Kind() == reflect.Struct:
		// struct. look up field or method of that name
		value = ""
		fmt.Printf("...is a struct\n")
	case ctx.Kind() == reflect.Func:
		// if the context itself is a function, call the function and use it's value recursively. This would let
		// the caller provide a closure that would produce the values.
		// other value
	}

	fmt.Printf("kind is %s\n", ctx.Kind())
	// Now we have the value at that place, see what we can do with it:
	// - if it's a function, execute it with the parameters.
	// - if it's a value, return it.
	// - if it's undefined, see if there is a _fallback property and recurse on that if there is.

	// Get the underlying value.
	v := reflect.ValueOf(value)

	switch {
	case value == nil:
		// see if there is a _fallback
		if name != "_fallback" {
			fallback, e := d.Locate(context, "_fallback", nil)
			if e != nil {
				return nil, e
			}
			if fallback != nil {
				return d.Locate(fallback, name, args)
			}
		}
		// we couldn't work it out, just return nil with no error.
		return nil, nil
	case v.Kind() == reflect.Func:
		// reflection funkiness; create a slice of args asserted as reflect.Value.
		a := make([]reflect.Value, 0)
		for _, x := range args {
			a = append(a, reflect.ValueOf(x))
		}
		result := v.Call(a)

		// we ignore any other values returned.
		return result[0].Interface(), nil
	}

	// default behaviour is to return the value uninterpreted.
	return value, nil
}
