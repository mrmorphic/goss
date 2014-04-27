package orm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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

	res, e := Query(sql)
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
	fmt.Printf("query is %s\n", sql)
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
