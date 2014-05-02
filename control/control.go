package control

import (
	"errors"
	"net/http"
	"reflect"
)

// Controller is a handler that has an initialisation method for passing through a DataObject.
type Controller interface {
	http.Handler

	// this sets the request before ServeHTTP is called so that each implementation of ServeHTTP doesn't
	// have to do it. It allows BaseController to implement a bunch of common methods that want the request.
	// @todo is there a more elegant way to do this, so that a controller is just a handler, but the request
	// @todo is still available to utility functions in BaseController.
	Init(r *http.Request)
}

// a map from controller class names to "prototype" controller instances, used by getControllerInstance
// to generate new controllers.
var controllers map[string]Controller

func init() {
	controllers = map[string]Controller{}
}

// AddController registers a controller for a data object type
func AddController(className string, c Controller) {
	controllers[className] = c
}

// getControllerInstance generates a new instance of the controller for a data object. If it used the same controller
// instance as provided by AddController, it would not work concurrently.
func getControllerInstance(className string) (Controller, error) {
	c := controllers[className]
	if c == nil {
		return nil, errors.New("Could not locate a controller for DataObject of class '" + className + "'")
	}

	// Get the type that the interface points to
	t := reflect.TypeOf(c)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	new := reflect.New(t)
	return new.Interface().(Controller), nil
}
