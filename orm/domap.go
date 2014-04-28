package orm

import (
	"fmt"
	"github.com/mrmorphic/goss/convert"
)

// This is a basic implementation of DataObject, to be used when the ORM returns an object
// from the database where the ClassName is not registered. The object is represented as a map.
type DataObjectMap map[string]interface{}

func (obj DataObjectMap) Get(fieldName string, args ...interface{}) interface{} {
	return obj[fieldName]
}

// Return string representation of the field
func (obj DataObjectMap) GetStr(fieldName string, args ...interface{}) string {
	return convert.AsString(obj.Get(fieldName))
}

func (obj DataObjectMap) GetInt(fieldName string, args ...interface{}) (int, error) {
	return convert.AsInt(obj.Get(fieldName))
}

func (obj DataObjectMap) Set(fieldName string, value interface{}) {
	obj[fieldName] = value
}

func (obj DataObjectMap) Debug() string {
	s := "DataObject:\n"
	for f, v := range obj {
		s += fmt.Sprintf("  %s: %s\n", f, v)
	}
	return s
}

func NewDataObjectMap() DataObjectMap {
	return map[string]interface{}{}
}
