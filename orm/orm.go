package orm

import (
	"github.com/mrmorphic/goss"
	"reflect"
)

// metadataSource is a path to the file containing metadata used by the ORM.
var metadataSource string

var dbMetadata *DBMetadata

// A map of class names to DataObject instances which is used when we get objects
// from the database.
var models map[string]DataObject

// DataObject is an interface for items returned by the ORM.
type DataObject interface {
	goss.Evaluater
	// set a field
	Set(field string, value interface{})
}

// DataQuery is an interface for constructing queries. The interface is chainable, with many
// methods returning a new or modified DataQuery object.
type DataQuery interface {
	// Where adds an object selection condition to a query. There are two forms allowable,
	// one where
	Where(interface{}) DataQuery

	// Sort specifies a sort order for query results. Sort takes one or more string parameters.
	// Each parameter is a field name, optionally suffixed with a space and "asc" or "desc".
	Sort(string, ...string) DataQuery

	Limit(offset int, length int) DataQuery

	// Execute the query and return it's result. All error handling is returned via Run to
	// simplify the signatures of chainable methods.
	Run() (interface{}, error)
}

func IsHierarchical(className string) bool {
	return dbMetadata.IsHierarchical(className)
}

// Register one or more model instances. The map key is the ClassName value returned in
// a data object fetch, and the instance is an object that will be used as a prototype
// for generating new DataObject instances.
func RegisterModels(m map[string]DataObject) {
	// @todo make concurrency-safe.
	for k, v := range m {
		models[k] = v
	}
}

// Given a class name, return a DataObject instance. If className has been registered
// using RegisterModels, then a new, empty instance of the data object concrete type
// is returned. Otherwise, a DataObjectMap is returned.
func GetModelInstance(className string) DataObject {
	proto := models[className]
	if proto == nil {
		return NewDataObjectMap()
	}

	// Get the type that the interface points to
	t := reflect.TypeOf(proto)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	new := reflect.New(t)
	return new.Interface().(DataObject)
}
