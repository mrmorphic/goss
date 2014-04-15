package goss

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Execute a SQL query, returning the resulting rows. Creates a database connection on demand.
func (ctx *DBContext) Query(sql string) (q *sql.Rows, e error) {
	//	fmt.Println("sql: " + sql)
	if ctx.db == nil {
		// get databae connection on demand
		ctx.db, e = dbFactory()
		if e != nil {
			return
		}
	}

	st, e := ctx.db.Prepare(sql)
	if e != nil {
		return
	}

	q, e = st.Query()
	return
}

type DataObject struct {
	fields map[string]interface{}
	// @todo add ClassInfo to DataObject
	//	fields []interface{}
}

func (obj *DataObject) FieldByName(fieldName string) interface{} {
	return obj.fields[fieldName]
}

// Return string representation of the field
func (obj *DataObject) AsString(fieldName string) string {
	iv := obj.FieldByName(fieldName)

	t := reflect.TypeOf(iv)
	switch t.String() {
	case "*sql.NullString":
		s := iv.(*sql.NullString)
		return s.String
	default:
		return "don't know this type: " + t.String()
	}
	return "don't know"
}

func (obj *DataObject) AsInt(fieldName string) (int, error) {
	iv := obj.FieldByName(fieldName)

	t := reflect.TypeOf(iv)
	switch t.String() {
	case "*sql.NullInt64":
		s := iv.(*sql.NullInt64)
		return int(s.Int64), nil
	case "*sql.NullString":
		// observed but not expected
		v := iv.(*sql.NullString)
		return strconv.Atoi(v.String)
	}
	return 0, errors.New("AsInt doesn't understand type " + t.String())
}

func (obj *DataObject) Debug() string {
	s := "DataObject:\n"
	for f, v := range obj.fields {
		s += fmt.Sprintf("  %s: %s\n", f, v)
	}
	return s
}

func (ctx *DBContext) DataObjectFromRow(r *sql.Rows) (obj *DataObject, e error) {
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

	m := make(map[string]interface{}, colCount)

	for i, c := range cols {
		m[c] = field[i]
	}

	return &DataObject{m}, nil
}

type DataList struct {
	Items []*DataObject
}

func NewDataList() *DataList {
	// return &DataList{make([]*DataObject, 10)}
	return &DataList{make([]*DataObject, 0, 10)}
}

func (set *DataList) First() *DataObject {
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

func (q *DataQuery) Count(coumn string) *DataQuery {
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
func (q *DataQuery) sql(ctx *DBContext) (s string, e error) {
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
	baseClass := ctx.Metadata.GetClass(q.baseClass)
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

func (q *DataQuery) Exec(ctx *DBContext) (set *DataList, e error) {
	sql, e := q.sql(ctx)
	if e != nil {
		return nil, e
	}

	res, e := ctx.Query(sql)
	if e != nil {
		fmt.Printf("ERROR EXECUTING SQL: %s\n", e)
		return nil, e
	}

	set = NewDataList()

	for res.Next() {
		obj, e := ctx.DataObjectFromRow(res)
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
