package control

import (
	"fmt"
	"github.com/mrmorphic/goss/data"
	"github.com/mrmorphic/goss/orm"
	"testing"
)

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
	Fallback orm.DataObject

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

func TestRequireJS(t *testing.T) {
	cc := &third{}
	cc.TestFirstField = "first field"
	cc.TestSecondField = "second field"
	cc.TestThirdField = "third field"

	obj := orm.NewDataObjectMap()
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
		v := data.Eval(cc, expr)
		fmt.Printf("v=%s\n", v)
		if v.(string) != expected {
			t.Error(fmt.Errorf("Expected '%s', got %s", expected, v).Error())
		}
	}
}
