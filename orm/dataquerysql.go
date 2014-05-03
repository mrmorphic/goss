package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// database is actually a connection pool. The pool is automatically managed, and works across go-routines.
var database *sql.DB

// Execute a SQL query, returning the resulting rows. Caller should ensure that rows.Close is called.
func Query(sql string) (q *sql.Rows, e error) {
	st, e := database.Prepare(sql)
	if e != nil {
		return
	}
	defer st.Close()

	q, e = st.Query()
	return
}

// DataQuerySQL is an implementer of DataQuery for SQL databases.
type DataQuerySQL struct {
	where     []string
	columns   []string
	orderBy   string
	start     int
	limit     int
	baseClass string
}

func (q *DataQuerySQL) Where(clause interface{}) DataQuery {
	s := clause.(string)
	q.where = append(q.where, s)
	return q
}

func (q *DataQuerySQL) Sort(clause string, rest ...string) DataQuery {
	q.orderBy = clause
	return q
}

func (q *DataQuerySQL) Run() (interface{}, error) {
	sql, e := q.sql()
	if e != nil {
		return nil, e
	}

	// ensure rows are closed. Implication is that this function must read all rows, can't
	// leave the query open for incremental reading.
	res, e := Query(sql)
	defer res.Close()

	if e != nil {
		fmt.Printf("ERROR EXECUTING SQL: %s\n", e)
		return nil, e
	}

	set := NewDataList(q)

	for res.Next() {
		obj, e := DataObjectFromRow(res)
		if e != nil {
			return nil, e
		}
		set.Append(obj)
	}

	return set, nil

}

func (q *DataQuerySQL) Columns(columns []string) DataQuery {
	q.columns = columns
	return q
}

func (q *DataQuerySQL) Count(column string) DataQuery {
	return q
}

func (q *DataQuerySQL) Limit(start, number int) DataQuery {
	q.start = start
	q.limit = number
	return q
}

func (q *DataQuerySQL) Filter(field string, filterValue interface{}) DataQuery {
	return q
}

// Generate the SQL for this DataQuery
func (q *DataQuerySQL) sql() (s string, e error) {
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
	//	fmt.Printf("query is %s\n", sql)
	return sql, nil
}

func NewQuery(className string) DataQuery {
	return NewQuerySQL(className)
}

func NewQuerySQL(className string) DataQuery {
	q := new(DataQuerySQL)
	q.start = -1
	q.baseClass = className
	return q
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
	// get associated value
	e = r.Scan(field...)

	//fmt.Println("scanned fields")
	if e != nil {
		fmt.Printf("got an error though: %s\n", e)
		return nil, e
	}

	m := NewDataObjectMap()

	for i, c := range cols {
		m[c] = flatten(field[i])
	}

	return m, nil
}

// This is a bit hassly. When we copy over field values, we need to ignore the Valid property
// of each field, since SilverStripe effectively uses the zero values when underlying
// SQL field is null. Without this conversion, all consumers of data object need to be aware
// of sql package field values.
func flatten(sqlField interface{}) interface{} {
	switch v := sqlField.(type) {
	case *sql.NullBool:
		return v.Bool
	case *sql.NullFloat64:
		return v.Float64
	case *sql.NullInt64:
		return v.Int64
	case *sql.NullString:
		return v.String
	}

	return sqlField
}
