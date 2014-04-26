package template

import (
	"fmt"
	"github.com/mrmorphic/goss/orm"
	"testing"
)

// Test that == and = are treated the same
func TestFallback(t *testing.T) {
	context := map[string]interface{}{
		"prop1": "value1",
		"_fallback": map[string]interface{}{
			"prop2": "value2",
		},
	}

	loc := NewDefaultLocator()
	p1, e := loc.Locate(context, "prop1", []interface{}{})
	if e != nil {
		t.Error(e.Error())
	}

	p2, e := loc.Locate(context, "prop2", []interface{}{})
	if e != nil {
		t.Error(e.Error())
	}

	if p1 != "value1" {
		t.Errorf("p1 is expected to be 'value1', got %s", p1)
	}
	if p2 != "value2" {
		t.Errorf("p1 is expected to be 'value1', got %s", p1)
	}
}

func TestDataObject(t *testing.T) {
	fmt.Printf("running TestDataObject\n")
	obj := orm.NewDataObject()
	fmt.Printf("got data object\n")
	obj.Set("foo", "bar")
	fmt.Printf("set foo\n")
	loc := NewDefaultLocator()
	fmt.Printf("got locater\n")
	x := obj.AsString("foo")
	fmt.Printf("x is %s\n", x)
	v, e := loc.Locate(obj, "foo", nil)
	if e != nil {
		t.Error(e.Error())
	}
	fmt.Printf("foo is %s\n", v)
}
