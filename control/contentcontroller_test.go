package control

import (
	"fmt"
	"github.com/mrmorphic/goss/data"
	"github.com/mrmorphic/goss/orm"
	"testing"
)

// This type will represent a contained or embedded type within testContainer.
// This is analogous to BaseController
type testContained struct {
	TestBase string
}

func (c *testContainer) SiteConfig() orm.DataObject {
	r := orm.NewDataObjectMap()
	r["Name"] = "SiteName"
	return r
}

// This type is the container. It is analagous to ContentController.
type testContainer struct {
	testContained
	Fallback orm.DataObject

	Test string
}

func (c *testContainer) Init(obj orm.DataObject) {
	c.Fallback = obj
}

func TestRequireJS(t *testing.T) {
	cc := &testContainer{}
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

	v = data.Eval(cc, "SiteConfig")
	fmt.Printf("SiteConfig is %s\n", v)
	// test that a function on BaseController can be fetched

	// source := `<html><head><title>x</title></head><body><% require javascript("themes/simple/javascript/test.js") %><div>test</div></body></html>`
	// context := map[string]interface{}{}

	// b, e := compileAndExecute(source, context)
	// if e != nil {
	// 	t.Error(e.Error())
	// 	return
	// }
}
