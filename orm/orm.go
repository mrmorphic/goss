package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mrmorphic/goss"
	"github.com/mrmorphic/goss/convert"
	"strconv"
	"strings"
	"time"
)

// database is actually a connection pool. The pool is automatically managed, and works across go-routines.
var database *sql.DB

// metadataSource is a path to the file containing metadata used by the ORM.
var metadataSource string

var dbMetadata *DBMetadata

// Execute a SQL query, returning the resulting rows.
func Query(sql string) (q *sql.Rows, e error) {
	st, e := database.Prepare(sql)
	if e != nil {
		return
	}

	q, e = st.Query()
	return
}

type DataObject interface {
	goss.Evaluater
	// set a field
	Set(field string, value interface{})
}

// This is a basic implementation of DataObject, to be used when the ORM returns an object
// from the database where the ClassName is not registered.
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

func DataObjectFromRow(r *sql.Rows) (obj DataObject, e error) {
	// Get columns
	cols, e := r.Columns()
	colCount := len(cols)

	var field []interface{}
	for i := 0; i < colCount; i++ {
		switch {
		case cols[i][:2] == "b:":
			field = append(field, new(sql.NullBool))
		case cols[i][:2] == "f:":
			field = append(field, new(sql.NullFloat64))
		case cols[i][:2] == "i:":
			field = append(field, new(sql.NullInt64))
		case cols[i][:2] == "s:":
			field = append(field, new(sql.NullString))
		case cols[i][:2] == "t:":
			field = append(field, new(time.Time))
		default:
			field = append(field, new(sql.NullString))
		}
	}

	//fmt.Printf("cols are %s\n", cols)
	//fmt.Printf("there are %d columns\n", colCount)
	//fmt.Println("about to scan values")
	// get associated values
	e = r.Scan(field...)

	//fmt.Println("scanned fields")
	if e != nil {
		fmt.Printf("got an error though: %s\n", e)
		return nil, e
	}

	m := NewDataObjectMap()

	for i, c := range cols {
		m[c] = field[i]
	}

	return m, nil
}

func IsHierarchical(className string) bool {
	return dbMetadata.IsHierarchical(className)
}
