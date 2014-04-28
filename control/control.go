package control

import (
	"errors"
	"github.com/mrmorphic/goss/orm"
	"net/http"
	//	"reflect"
)

// Controller is a handler that has an initialisation method for passing through a DataObject.
type Controller interface {
	http.Handler
}

var controllers map[string]Controller

func init() {
	controllers = make(map[string]Controller)
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

	return c, nil
	//	// Get the type that the interface points to
	//	t := reflect.TypeOf(c)
	//
	//	new := reflect.New(t)
	//	return new, nil
}
