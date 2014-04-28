package control

import (
	"fmt"
	"github.com/mrmorphic/goss/data"
	"github.com/mrmorphic/goss/orm"
	"testing"
)

type testController struct {
	BaseController
	Fallback orm.DataObject

	Test string
}

func (c *testController) Init(obj orm.DataObject) {
	c.Object = obj
}

func TestRequireJS(t *testing.T) {
	cc := &testController{}
	cc.Test = "test result"
	cc.TestBase = "test base result"

	obj := orm.NewDataObjectMap()
	obj["TestDOField"] = "DO field"
	obj["AnotherField"] = "foo"

	cc.Fallback = obj

	// test that properties on ContentController can be fetched
	v := data.Eval(cc, "Test")
	if v.(string) != "test result" {
		t.Error(fmt.Errorf("Expected 'test result', got %s", v).Error())
	}

	// test that properties within embedded BaseController can be fetched
	v = data.Eval(cc, "TestBase")
	if v.(string) != "test base result" {
		t.Error(fmt.Errorf("Expected 'test base result', got %s", v).Error())
	}

	// test that properties on the DataObject can be fetched
	v = data.Eval(cc, "TestDOField")
	if v.(string) != "DO field" {
		t.Error(fmt.Errorf("Expected 'DO field', got %s", v).Error())
	}

	// source := `<html><head><title>x</title></head><body><% require javascript("themes/simple/javascript/test.js") %><div>test</div></body></html>`
	// context := map[string]interface{}{}

	// b, e := compileAndExecute(source, context)
	// if e != nil {
	// 	t.Error(e.Error())
	// 	return
	// }
}
