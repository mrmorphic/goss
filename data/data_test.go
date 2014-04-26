package data

import (
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

	p1 := Eval(context, "prop1")
	p2 := Eval(context, "prop2")

	if p1.(string) != "value1" {
		t.Errorf("p1 is expected to be 'value1', got %s", p1)
	}
	if p2.(string) != "value2" {
		t.Errorf("p1 is expected to be 'value1', got %s", p1)
	}
}

func TestDataObject(t *testing.T) {
	obj := orm.NewDataObjectMap()
	obj.Set("foo", "bar")

	if x := obj.GetStr("foo"); x != "bar" {
		t.Error("Expected foo to be 'bar', got %s", x)
	}
}
