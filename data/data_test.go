package data

import (
	"fmt"
	"testing"
)

// Test that == and = are treated the same
func TestFallback(t *testing.T) {
	context := map[string]interface{}{
		"prop1": "value1",
		"Fallback": map[string]interface{}{
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

// This type will represent a contained or embedded type within testContainer.
// This is analogous to BaseController
type first struct {
	TestFirstField string
}

func (c *first) TestFirstFunc() string {
	return "first func"
}

// This type is the container. It is analagous to ContentController.
type second struct {
	first
	Fallback interface{}

	TestSecondField string
}

func (c *second) TestSecondFunc() string {
	return "second func"
}

type third struct {
	second

	TestThirdField string
}

func (c *third) TestThirdFunc() string {
	return "third func"
}

// The above types replicate the structure of BaseControllerStruct, ContentControllerStruct and
// a custom Controller embedding them. This test checks for various cases of function and
// property access
func TestControllerStructure(t *testing.T) {
	cc := &third{}
	cc.TestFirstField = "first field"
	cc.TestSecondField = "second field"
	cc.TestThirdField = "third field"

	obj := map[string]interface{}{}
	obj["TestDOField"] = "DO field"
	obj["AnotherField"] = "foo"

	cc.Fallback = obj

	tests := map[string]interface{}{
		"TestFirstField":  "first field",
		"TestFirstFunc":   "first func",
		"TestSecondField": "second field",
		"TestSecondFunc":  "second func",
		"TestThirdField":  "third field",
		"TestThirdFunc":   "third func",
		"AnotherField":    "foo",
	}

	for expr, expected := range tests {
		v := Eval(cc, expr)
		fmt.Printf("v=%s\n", v)
		if v.(string) != expected {
			t.Error(fmt.Errorf("Expected '%s', got %s", expected, v).Error())
		}
	}
}
