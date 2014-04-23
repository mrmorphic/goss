package template

import (
	"fmt"
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
type defaultLocator struct {
}

func newDefaultLocator() DataLocator {
	return &defaultLocator{}
}

func (d *defaultLocator) Locate(context interface{}, name string, args []interface{}) (interface{}, error) {
	val := reflect.ValueOf(context)
	switch {
	// context implements DataLocator
	case val.Kind() == reflect.Map:
		// @todo this is too restrictive. What we really want is to ensure that context is a map with a string key, and any type.
		m, ok := context.(map[string]interface{})
		if !ok {
			panic("locater: map must be map[string]interface{}")
		}
		return m[name], nil
		// struct
		// other value
	}
	return nil, fmt.Errorf("defaultLocator couldn't make sense of context")
}
