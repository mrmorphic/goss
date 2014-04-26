package convert

import (
	"database/sql"
	"errors"
	"reflect"
	"strconv"
)

func AsString(v interface{}) string {
	t := reflect.TypeOf(v)
	switch t.String() {
	case "*sql.NullString":
		s := v.(*sql.NullString)
		return s.String
	case "string":
		return v.(string)
	}
	return "don't know this type: " + t.String()
}

func AsInt(v interface{}) (int, error) {
	t := reflect.TypeOf(v)
	switch t.String() {
	case "*sql.NullInt64":
		s := v.(*sql.NullInt64)
		return int(s.Int64), nil
	case "*sql.NullString":
		// observed but not expected
		v := v.(*sql.NullString)
		return strconv.Atoi(v.String)
	}
	return 0, errors.New("AsInt doesn't understand type " + t.String())

}
