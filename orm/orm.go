package orm

import (
	"github.com/mrmorphic/goss"
)

// metadataSource is a path to the file containing metadata used by the ORM.
var metadataSource string

var dbMetadata *DBMetadata

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
