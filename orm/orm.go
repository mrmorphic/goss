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

type DataList struct {
	Items []DataObject
}

func NewDataList() *DataList {
	// return &DataList{make([]*DataObject, 10)}
	return &DataList{make([]DataObject, 0, 10)}
}

func (set *DataList) First() DataObject {
	if len(set.Items) < 1 {
		return nil
	}
	return set.Items[0]
}

type DataQuery struct {
	where     []string
	columns   []string
	orderBy   string
	start     int
	limit     int
	baseClass string
}

func (q *DataQuery) Where(clause string) *DataQuery {
	q.where = append(q.where, clause)
	return q
}

func (q *DataQuery) Columns(columns []string) *DataQuery {
	q.columns = columns
	return q
}

func (q *DataQuery) Count(column string) *DataQuery {
	return q
}

func (q *DataQuery) OrderBy(clause string) *DataQuery {
	q.orderBy = clause
	return q
}

func (q *DataQuery) Limit(start, number int) *DataQuery {
	q.start = start
	q.limit = number
	return q
}

func (q *DataQuery) Filter(field string, filterValue interface{}) *DataQuery {
	return q
}

// Generate the SQL for this DataQuery
func (q *DataQuery) sql() (s string, e error) {
	if q.baseClass == "" {
		return "", errors.New("No base class")
	}

	// columns
	sql := "select "
	if len(q.columns) == 0 {
		sql += "* "
	} else {
		sql += "\"" + strings.Join(q.columns, "\",\"") + "\" "
	}

	// Tables. This is basically a join of all tables from base DataObject thru to the table for the class, and all
	// tables for subclasses. This will have been precalculated, so it's trivial here.
	baseClass := dbMetadata.GetClass(q.baseClass)
	sql += "from " + baseClass.defaultFrom

	// where clause
	sql += " where " + baseClass.defaultWhere
	if len(q.where) > 0 {
		sql += " and " + strings.Join(q.where, " and ")
	}

	if q.orderBy != "" {
		sql += " order by " + q.orderBy
	}

	if q.start >= 0 {
		sql += " limit " + strconv.Itoa(q.start) + ", " + strconv.Itoa(q.limit)
	}
	fmt.Printf("query is %s\n", sql)
	return sql, nil
}

func (q *DataQuery) Exec() (set *DataList, e error) {
	sql, e := q.sql()
	if e != nil {
		return nil, e
	}

	res, e := Query(sql)
	if e != nil {
		fmt.Printf("ERROR EXECUTING SQL: %s\n", e)
		return nil, e
	}

	set = NewDataList()

	for res.Next() {
		obj, e := DataObjectFromRow(res)
		if e != nil {
			return nil, e
		}
		set.Items = append(set.Items, obj)
	}

	return set, nil
}

func NewQuery(className string) *DataQuery {
	q := new(DataQuery)
	q.start = -1
	q.baseClass = className
	return q
}

func IsHierarchical(className string) bool {
	return dbMetadata.IsHierarchical(className)
}
